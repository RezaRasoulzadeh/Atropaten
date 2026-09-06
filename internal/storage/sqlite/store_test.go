package sqlite

import (
	"context"
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
	if version != 9 {
		t.Fatalf("migration version = %d, want 9", version)
	}
	var foreignKeys int
	if err := reopened.db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatalf("foreign keys pragma: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign keys = %d, want 1", foreignKeys)
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
