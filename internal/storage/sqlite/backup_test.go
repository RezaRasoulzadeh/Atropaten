package sqlite

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"Atropaten/internal/domain"
	"Atropaten/internal/platform"
	_ "github.com/mattn/go-sqlite3"
)

func TestBackupRoundTripPreservesDatabaseAndManagedFiles(t *testing.T) {
	root := t.TempDir()
	paths := platform.DataPaths{Root: root, Database: filepath.Join(root, "atropaten.db"), Attachments: filepath.Join(root, "attachments"), Backups: filepath.Join(root, "backups")}
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(paths.Database)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	when := time.Date(2026, 9, 7, 8, 0, 0, 0, time.UTC)
	if _, err = store.db.Exec(`INSERT INTO materials(id,name,purchase_unit,consumption_unit,conversion_factor_units,physical_stock_units,reorder_level_units,average_unit_cost_rial,created_at,updated_at) VALUES('MAT-backup','Paper','pack','sheet',100,250,10,40,?,?)`, when, when); err != nil {
		t.Fatal(err)
	}
	if _, err = store.db.Exec(`INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES('MOV-backup','MAT-backup',?,'purchase',250,40,10000,'purchase','PUR-backup','seed',?)`, when, when); err != nil {
		t.Fatal(err)
	}
	entry := domain.JournalEntry{ID: "JE-backup", Description: "backup seed", SourceType: "test", SourceID: "backup", IdempotencyKey: "backup-seed", PostedAt: when, CreatedAt: when, Lines: []domain.JournalLine{{ID: "JE-backup-L1", JournalEntryID: "JE-backup", Position: 0, AccountID: "ACC-CASH", DebitRial: 10000}, {ID: "JE-backup-L2", JournalEntryID: "JE-backup", Position: 1, AccountID: "ACC-REVENUE", CreditRial: 10000}}}
	if _, err = store.PostJournalEntry(ctx, entry); err != nil {
		t.Fatal(err)
	}
	settings, err := store.GetShopSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	settings.ShopName = "Roundtrip Shop"
	if err = store.SaveShopSettings(ctx, settings); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(paths.Attachments, "artwork", "sample.txt")
	if err = os.MkdirAll(filepath.Dir(managed), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(managed, []byte("immutable artwork"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.db.Exec(`INSERT INTO attachments(id,owner_type,owner_id,file_name,path,mime_type,size_bytes,checksum,category,notes,created_at) VALUES('ATT-backup','quote','QUO-backup','sample.txt',?,'text/plain',16,'','artwork','',?)`, managed, when); err != nil {
		t.Fatal(err)
	}
	service := platform.NewBackupService(paths, store, ValidateDatabaseFile)
	backup, err := service.Create(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.SchemaVersion != CurrentSchemaVersion || backup.ManagedFileCount != 1 {
		t.Fatalf("backup=%+v", backup)
	}
	if _, err = service.Verify(backup.Path); err != nil {
		t.Fatal(err)
	}
	if _, err = store.db.Exec(`UPDATE shop_settings SET value='Changed' WHERE key='shop_name'`); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(managed, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	var reopened *Store
	reopen := func() error { var e error; reopened, e = Open(paths.Database); return e }
	if _, err = service.Restore(ctx, backup.Path, func() error { return nil }, reopen); err != nil {
		t.Fatal(err)
	}
	service.SetRepository(reopened)
	defer reopened.Close()
	settings, err = reopened.GetShopSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if settings.ShopName != "Roundtrip Shop" {
		t.Fatalf("settings=%+v", settings)
	}
	var quantity, journalLines int
	if err = reopened.db.QueryRow(`SELECT quantity_delta_units FROM inventory_movements WHERE id='MOV-backup'`).Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	if quantity != 250 {
		t.Fatalf("quantity=%d", quantity)
	}
	if err = reopened.db.QueryRow(`SELECT COUNT(*) FROM journal_lines WHERE journal_entry_id='JE-backup'`).Scan(&journalLines); err != nil {
		t.Fatal(err)
	}
	if journalLines != 2 {
		t.Fatalf("journal lines=%d", journalLines)
	}
	content, err := os.ReadFile(managed)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "immutable artwork" {
		t.Fatalf("managed file=%q", content)
	}
}

func TestBackupValidationRejectsCorruptMissingAndFutureArchives(t *testing.T) {
	root := t.TempDir()
	paths := platform.DataPaths{Root: root, Database: filepath.Join(root, "db"), Attachments: filepath.Join(root, "attachments"), Backups: filepath.Join(root, "backups")}
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(paths.Database)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := platform.NewBackupService(paths, store, ValidateDatabaseFile)
	databaseBytes, err := os.ReadFile(paths.Database)
	if err != nil {
		t.Fatal(err)
	}
	corrupt := filepath.Join(root, "corrupt.zip")
	if err = os.WriteFile(corrupt, []byte("not zip"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Verify(corrupt); err == nil {
		t.Fatal("corrupt archive accepted")
	}
	missing := filepath.Join(root, "missing.zip")
	if err = writeTestArchive(missing, nil, platform.BackupManifest{}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Verify(missing); err == nil || !strings.Contains(err.Error(), "manifest is missing") {
		t.Fatalf("missing manifest error=%v", err)
	}
	future := filepath.Join(root, "future.zip")
	if err = writeTestArchive(future, []byte("invalid"), platform.BackupManifest{FormatVersion: platform.BackupFormatVersion, ApplicationVersion: platform.ApplicationVersion, SchemaVersion: 999, Database: "database.sqlite"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Verify(future); err == nil || !strings.Contains(err.Error(), "unsupported schema") {
		t.Fatalf("future error=%v", err)
	}
	invalid := filepath.Join(root, "invalid-db.zip")
	if err = writeTestArchive(invalid, []byte("not sqlite"), platform.BackupManifest{FormatVersion: platform.BackupFormatVersion, ApplicationVersion: platform.ApplicationVersion, SchemaVersion: platform.CurrentSchemaVersion, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Database: "database.sqlite"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Verify(invalid); err == nil || !strings.Contains(err.Error(), "validate database") {
		t.Fatalf("invalid database error=%v", err)
	}
	missingFile := filepath.Join(root, "missing-file.zip")
	if err = writeTestArchive(missingFile, databaseBytes, platform.BackupManifest{FormatVersion: platform.BackupFormatVersion, ApplicationVersion: platform.ApplicationVersion, SchemaVersion: platform.CurrentSchemaVersion, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Database: "database.sqlite", Files: []platform.BackupFile{{Path: "attachments/missing.bin", SHA256: "00", Size: 1}}}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Verify(missingFile); err == nil || !strings.Contains(err.Error(), "managed file is missing") {
		t.Fatalf("missing managed file error=%v", err)
	}
	traversal := filepath.Join(root, "traversal.zip")
	if err = writeNamedArchive(traversal, "../escape", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Verify(traversal); err == nil || !strings.Contains(err.Error(), "unsafe archive") {
		t.Fatalf("traversal error=%v", err)
	}
}

func writeTestArchive(path string, database []byte, manifest platform.BackupManifest) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	z := zip.NewWriter(f)
	if database != nil {
		w, e := z.Create("database.sqlite")
		if e != nil {
			return e
		}
		if _, e = w.Write(database); e != nil {
			return e
		}
	}
	if manifest.FormatVersion != 0 {
		raw, e := json.Marshal(manifest)
		if e != nil {
			return e
		}
		w, e := z.Create("manifest.json")
		if e != nil {
			return e
		}
		if _, e = w.Write(raw); e != nil {
			return e
		}
	}
	if err = z.Close(); err != nil {
		return err
	}
	return f.Close()
}
func writeNamedArchive(path, name string, content []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	z := zip.NewWriter(f)
	w, err := z.Create(name)
	if err == nil {
		_, err = w.Write(content)
	}
	if closeErr := z.Close(); err == nil {
		err = closeErr
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}

func TestHistoricalSchemasUpgradeAndFutureSchemaIsRejected(t *testing.T) {
	for _, version := range []int{8, 10, 11, 12, 13} {
		t.Run("v"+itoa(version), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "legacy.db")
			raw, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
			if err != nil {
				t.Fatal(err)
			}
			for _, migration := range migrations[:version] {
				if _, err = raw.Exec(migration.sql); err != nil {
					raw.Close()
					t.Fatal(err)
				}
			}
			if _, err = raw.Exec(`CREATE TABLE schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
				raw.Close()
				t.Fatal(err)
			}
			for _, migration := range migrations[:version] {
				if _, err = raw.Exec(`INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, migration.version, time.Now().UTC()); err != nil {
					raw.Close()
					t.Fatal(err)
				}
			}
			if err = raw.Close(); err != nil {
				t.Fatal(err)
			}
			upgraded, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer upgraded.Close()
			got, err := upgraded.SchemaVersion(context.Background())
			if err != nil || got != CurrentSchemaVersion {
				t.Fatalf("version=%d err=%v", got, err)
			}
		})
	}
	path := filepath.Join(t.TempDir(), "future.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec(`INSERT INTO schema_migrations(version,applied_at) VALUES(999,?)`, time.Now().UTC())
	if err != nil {
		raw.Close()
		t.Fatal(err)
	}
	raw.Close()
	if _, err = Open(path); err == nil || !strings.Contains(err.Error(), "unsupported future schema") {
		t.Fatalf("future open error=%v", err)
	}
}

func TestRestoreFailureLeavesLiveDataUsableAndRollsBackFailedReopen(t *testing.T) {
	root := t.TempDir()
	paths := platform.DataPaths{Root: root, Database: filepath.Join(root, "db"), Attachments: filepath.Join(root, "attachments"), Backups: filepath.Join(root, "backups")}
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(paths.Database)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	settings, err := store.GetShopSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	settings.ShopName = "Live"
	if err = store.SaveShopSettings(ctx, settings); err != nil {
		t.Fatal(err)
	}
	service := platform.NewBackupService(paths, store, ValidateDatabaseFile)
	backup, err := service.Create(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Restore(ctx, backup.Path, func() error { return errors.New("live store busy") }, func() error { return errors.New("must not reopen") }); err == nil {
		t.Fatal("close failure accepted")
	}
	if err = store.db.QueryRow(`SELECT value FROM shop_settings WHERE key='shop_name'`).Scan(&settings.ShopName); err != nil || settings.ShopName != "Live" {
		t.Fatalf("live data after close failure=%q err=%v", settings.ShopName, err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	var reopened *Store
	attempts := 0
	closeLive := func() error {
		if reopened != nil {
			err := reopened.Close()
			reopened = nil
			return err
		}
		return nil
	}
	reopen := func() error {
		attempts++
		if attempts == 1 {
			return errors.New("simulated reopen failure")
		}
		var e error
		reopened, e = Open(paths.Database)
		return e
	}
	service.SetRepository(store)
	if _, err = service.Restore(ctx, backup.Path, closeLive, reopen); err == nil {
		t.Fatal("reopen failure accepted")
	}
	if attempts != 2 || reopened == nil {
		t.Fatalf("rollback reopen attempts=%d store=%v", attempts, reopened != nil)
	}
	defer reopened.Close()
	if err = reopened.db.QueryRow(`SELECT value FROM shop_settings WHERE key='shop_name'`).Scan(&settings.ShopName); err != nil || settings.ShopName != "Live" {
		t.Fatalf("rolled back data=%q err=%v", settings.ShopName, err)
	}
}

func itoa(v int) string { return strconv.Itoa(v) }
