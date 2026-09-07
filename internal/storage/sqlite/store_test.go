package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestFreshDatabaseMigratesAndReopens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "atropaten.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	var version int
	if err := reopened.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil {
		t.Fatalf("migration metadata: %v", err)
	}
	if version != 13 {
		t.Fatalf("migration version = %d, want 13", version)
	}
	var foreignKeys int
	if err := reopened.db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatalf("foreign keys pragma: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign keys = %d, want 1", foreignKeys)
	}
}

func TestV9ToV10PreservesPaymentAllocations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy-v9.db")
	raw, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations[:9] {
		if _, err = raw.Exec(migration.sql); err != nil {
			raw.Close()
			t.Fatalf("migration %d: %v", migration.version, err)
		}
	}
	if _, err = raw.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		raw.Close()
		t.Fatal(err)
	}
	for _, migration := range migrations[:9] {
		if _, err = raw.Exec(`INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, migration.version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			raw.Close()
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO accounts(id,code,name,type,active,system,created_at,updated_at) VALUES('ACC-CASH','1000','Cash','asset',1,1,?,?)`, []any{now, now}},
		{`INSERT INTO financial_accounts(id,name,type,ledger_account_id,active,created_at,updated_at) VALUES('FIN-CASH','Cash','cash','ACC-CASH',1,?,?)`, []any{now, now}},
		{`INSERT INTO journal_entries(id,entry_number,posted_at,description,source_type,source_id,idempotency_key,created_at) VALUES('JE-OLD','JE-1001',?,'Legacy payment','payment','PAY-OLD','legacy-payment',?)`, []any{now, now}},
		{`INSERT INTO payments(id,payment_number,direction,method,amount_rial,posted_at,financial_account_id,reference,notes,status,journal_entry_id,idempotency_key,created_at) VALUES('PAY-OLD','PAY-1001','incoming','cash',50,?,'FIN-CASH','','','posted','JE-OLD','legacy-pay-key',?)`, []any{now, now}},
		{`INSERT INTO payment_allocations(id,payment_id,position,target_type,target_id,amount_rial,reversed) VALUES('AL-OLD','PAY-OLD',0,'order','ORD-OLD',50,0)`, nil},
	}
	for _, statement := range statements {
		if _, err = raw.Exec(statement.query, statement.args...); err != nil {
			raw.Close()
			t.Fatal(err)
		}
	}
	if err = raw.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var target string
	if err = store.db.QueryRow(`SELECT target_type||':'||target_id FROM payment_allocations WHERE id='AL-OLD'`).Scan(&target); err != nil {
		t.Fatal(err)
	}
	if target != "order:ORD-OLD" {
		t.Fatalf("legacy allocation=%q", target)
	}
}

func TestMaterialPersistenceAndLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "atropaten.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	now := time.Date(2024, time.August, 12, 7, 0, 0, 0, time.UTC)
	material, err := domain.NewMaterial("MAT-test", domain.MaterialDraft{
		Name: "A4 80gsm Paper", SKU: "PAPER-A4", Category: "Paper", PurchaseUnit: "pack", ConsumptionUnit: "sheet",
		ConversionFactor: 500 * domain.QuantityScale, PhysicalStock: 1_250_001, ReorderLevel: 500 * domain.QuantityScale,
		AverageUnitCostRial: 123_456_789, PreferredSupplier: "Pars Paper", Notes: "Keep dry",
	}, now)
	if err != nil {
		t.Fatalf("NewMaterial: %v", err)
	}
	ctx := context.Background()
	if err := store.Create(ctx, material); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := store.Get(ctx, material.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.AverageUnitCostRial != 123_456_789 || got.PhysicalStock != 1_250_001 {
		t.Fatalf("lossy values after create: cost=%d stock=%d", got.AverageUnitCostRial, got.PhysicalStock)
	}
	got.Name = "A4 80gsm recycled"
	got.UpdatedAt = now.Add(time.Hour)
	if err := store.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	got, err = reopened.Get(ctx, material.ID)
	if err != nil || got.Name != "A4 80gsm recycled" || got.AverageUnitCostRial != 123_456_789 {
		t.Fatalf("reopened material = %+v, err=%v", got, err)
	}
	got.Active = false
	if err := reopened.Update(ctx, got); err != nil {
		t.Fatalf("archive: %v", err)
	}
	active, err := reopened.List(ctx, false)
	if err != nil || len(active) != 0 {
		t.Fatalf("active list = %v, err=%v", active, err)
	}
	all, err := reopened.List(ctx, true)
	if err != nil || len(all) != 1 || all[0].Active {
		t.Fatalf("all list = %v, err=%v", all, err)
	}
}
