package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mattn/go-sqlite3"
)

const CurrentSchemaVersion = 13

// SnapshotTo uses SQLite's online backup API so a live WAL database is copied
// from a consistent SQLite snapshot rather than by copying its files.
func (s *Store) SnapshotTo(ctx context.Context, destination string) error {
	if strings.TrimSpace(destination) == "" {
		return errors.New("snapshot destination is required")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	_ = os.Remove(destination)
	dest, err := sql.Open("sqlite3", destination+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open snapshot: %w", err)
	}
	dest.SetMaxOpenConns(1)
	defer func() { _ = dest.Close() }()
	srcConn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer srcConn.Close()
	destConn, err := dest.Conn(ctx)
	if err != nil {
		return err
	}
	defer destConn.Close()
	err = destConn.Raw(func(dst any) error {
		return srcConn.Raw(func(src any) error {
			dstSQLite, ok := dst.(*sqlite3.SQLiteConn)
			if !ok {
				return errors.New("snapshot destination is not sqlite")
			}
			srcSQLite, ok := src.(*sqlite3.SQLiteConn)
			if !ok {
				return errors.New("snapshot source is not sqlite")
			}
			backup, err := dstSQLite.Backup("main", srcSQLite, "main")
			if err != nil {
				return err
			}
			defer backup.Close()
			for {
				done, err := backup.Step(128)
				if err != nil {
					return err
				}
				if done {
					return nil
				}
			}
		})
	})
	if err != nil {
		_ = os.Remove(destination)
		return fmt.Errorf("copy sqlite snapshot: %w", err)
	}
	return nil
}

func (s *Store) SchemaVersion(ctx context.Context) (int, error) {
	var version int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version); err != nil {
		return 0, err
	}
	return version, nil
}

func (s *Store) ManagedFilePaths(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT path FROM attachments WHERE trim(path) <> '' UNION SELECT value FROM shop_settings WHERE key='logo_path' AND trim(value) <> ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := map[string]bool{}
	paths := make([]string, 0)
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		path = strings.TrimSpace(path)
		if path != "" && !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

// ValidateDatabaseFile is intentionally independent of Store.Open: it checks
// integrity and foreign keys before a backup can be restored or trusted.
func ValidateDatabaseFile(path string) error {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return err
	}
	var foreignKeys int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		return err
	}
	if foreignKeys != 1 {
		return errors.New("sqlite foreign keys are disabled")
	}
	var integrity string
	if err := db.QueryRow(`PRAGMA integrity_check`).Scan(&integrity); err != nil {
		return err
	}
	if strings.ToLower(integrity) != "ok" {
		return fmt.Errorf("sqlite integrity check failed: %s", integrity)
	}
	rows, err := db.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("sqlite foreign key check failed")
	}
	if err := rows.Err(); err != nil {
		return err
	}
	var version int
	if err := db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version < 1 || version > CurrentSchemaVersion {
		return fmt.Errorf("unsupported schema version %d", version)
	}
	if version < CurrentSchemaVersion {
		if err := db.Close(); err != nil {
			return fmt.Errorf("close pre-migration validation database: %w", err)
		}
		migrated, err := Open(path)
		if err != nil {
			return fmt.Errorf("validate migrations: %w", err)
		}
		if err := migrated.Close(); err != nil {
			return fmt.Errorf("close migrated validation database: %w", err)
		}
	}
	return nil
}
