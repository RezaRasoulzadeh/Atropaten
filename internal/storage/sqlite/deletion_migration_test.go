package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestDeletionAuditMigrationUpgradesBothV26Schemas(t *testing.T) {
	for _, legacy := range []bool{true, false} {
		name := "with_invoice_column"
		if legacy {
			name = "missing_invoice_column"
		}
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "v26.db")
			raw, err := sql.Open("sqlite3", path)
			if err != nil {
				t.Fatal(err)
			}
			defer raw.Close()
			if _, err = raw.Exec(`CREATE TABLE schema_migrations(version INTEGER PRIMARY KEY,applied_at TEXT NOT NULL)`); err != nil {
				t.Fatal(err)
			}
			for _, m := range migrations {
				if m.version > 26 {
					break
				}
				statement := m.sql
				if legacy && m.version == 26 {
					statement = strings.Replace(statement, "invoice_id TEXT, deleted_at", "deleted_at", 1)
				}
				if _, err = raw.Exec(statement); err != nil {
					t.Fatalf("migration %d: %v", m.version, err)
				}
				if _, err = raw.Exec(`INSERT INTO schema_migrations VALUES(?,?)`, m.version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = raw.Exec(`INSERT INTO deleted_order_records(id,order_number,deleted_at) VALUES('OLD','ORD-OLD','2026-09-01T00:00:00Z')`); err != nil {
				t.Fatal(err)
			}
			if !legacy {
				if _, err = raw.Exec(`UPDATE deleted_order_records SET invoice_id='INV-OLD' WHERE id='OLD'`); err != nil {
					t.Fatal(err)
				}
			}
			if err = raw.Close(); err != nil {
				t.Fatal(err)
			}
			for n := 0; n < 2; n++ {
				s, err := Open(path)
				if err != nil {
					t.Fatal(err)
				}
				var number, deleted string
				var invoice sql.NullString
				err = s.db.QueryRow(`SELECT order_number,deleted_at,invoice_id FROM deleted_order_records WHERE id='OLD'`).Scan(&number, &deleted, &invoice)
				if err != nil || number != "ORD-OLD" || deleted != "2026-09-01T00:00:00Z" || (!legacy && invoice.String != "INV-OLD") {
					t.Fatalf("audit record changed: %q %q %+v err=%v", number, deleted, invoice, err)
				}
				if n == 0 {
					ctx := context.Background()
					order := domain.NewOrder("NEW", "", time.Now().UTC())
					if err = s.CreateOrder(ctx, order); err != nil {
						t.Fatal(err)
					}
					if err = s.DeleteOrder(ctx, order.ID); err != nil {
						t.Fatalf("delete after upgrade: %v", err)
					}
				}
				if err = s.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
