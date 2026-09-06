package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestMigrationUpgradeKeepsExistingMaterials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	if _, err := legacy.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatalf("create legacy migration table: %v", err)
	}
	if _, err := legacy.Exec(migrations[0].sql); err != nil {
		t.Fatalf("create legacy materials table: %v", err)
	}
	created := "2024-08-12T07:00:00Z"
	if _, err := legacy.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (1, ?)`, created); err != nil {
		t.Fatalf("record legacy migration: %v", err)
	}
	if _, err := legacy.Exec(`INSERT INTO materials
		(id, name, sku, category, purchase_unit, consumption_unit, conversion_factor_units,
		physical_stock_units, reorder_level_units, average_unit_cost_rial, preferred_supplier,
		notes, active, created_at, updated_at)
		VALUES ('MAT-legacy', 'Legacy paper', 'LEG-1', 'Paper', 'pack', 'sheet', 500000000, 1250000, 500000, 123456789, 'Supplier', 'Keep', 1, ?, ?)`, created, created); err != nil {
		t.Fatalf("insert legacy material: %v", err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	store, err := Open(path)
	if err != nil {
		t.Fatalf("upgrade database: %v", err)
	}
	defer store.Close()
	var version int
	if err := store.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if version != 3 {
		t.Fatalf("migration version = %d, want 3", version)
	}
	material, err := store.Get(context.Background(), "MAT-legacy")
	if err != nil {
		t.Fatalf("read legacy material: %v", err)
	}
	if material.AverageUnitCostRial != 123456789 || material.PhysicalStock != domain.Quantity(1250000) {
		t.Fatalf("legacy material changed during migration: %+v", material)
	}
}

func TestM1002DatabaseUpgradesToM1003WithoutLosingServices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m1002.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open M1-002 database: %v", err)
	}
	if _, err := legacy.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	for version := 1; version <= 2; version++ {
		if _, err := legacy.Exec(migrations[version-1].sql); err != nil {
			t.Fatalf("apply legacy migration %d: %v", version, err)
		}
		if _, err := legacy.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (?, ?)`, version, "2024-08-12T07:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := legacy.Exec(`INSERT INTO services (id, name, code, category, description, active, created_at, updated_at) VALUES ('SVC-existing', 'Existing service', '', '', '', 1, '2024-08-12T07:00:00Z', '2024-08-12T07:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`INSERT INTO service_parameters (id, service_id, parameter_key, label, parameter_type, required, display_order, default_value, unit_label, active, created_at, updated_at) VALUES ('PAR-existing', 'SVC-existing', 'quantity', 'Quantity', 'integer', 1, 0, '1', '', 1, '2024-08-12T07:00:00Z', '2024-08-12T07:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatalf("upgrade M1-002 database: %v", err)
	}
	defer store.Close()
	service, err := store.GetService(context.Background(), "SVC-existing")
	if err != nil {
		t.Fatalf("read existing service: %v", err)
	}
	if service.Name != "Existing service" || len(service.Parameters) != 1 || len(service.Components) != 0 {
		t.Fatalf("existing service changed during migration: %+v", service)
	}
	var machineTable, componentTable string
	if err := store.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='machines'`).Scan(&machineTable); err != nil {
		t.Fatalf("machines table missing: %v", err)
	}
	if err := store.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='service_cost_components'`).Scan(&componentTable); err != nil {
		t.Fatalf("components table missing: %v", err)
	}
}

func TestServicePersistenceOrderingAndTransactionalRollback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "services.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	ctx := context.Background()
	now := time.Date(2024, time.August, 12, 7, 0, 0, 0, time.UTC)
	min := domain.Quantity(125001)
	service, err := domain.NewService("SVC-print", domain.ServiceDraft{
		Name: "Digital Print", Code: "PRINT", Category: "Production", Description: "Generic print service",
		Parameters: []domain.ServiceParameterDraft{
			{ID: "PAR-quantity", Key: "quantity", Label: "Quantity", Type: domain.ParameterInteger, Required: true, DefaultValue: "1"},
			{ID: "PAR-hours", Key: "estimated_hours", Label: "Estimated hours", Type: domain.ParameterDecimal, Required: true, DefaultValue: "0.125001", MinValue: &min},
			{ID: "PAR-size", Key: "paper_size", Label: "Paper size", Type: domain.ParameterChoice, Options: []string{"A4", "A3"}, DefaultValue: "A4"},
		},
		Components: []domain.ServiceCostComponentDraft{
			{ID: "CMP-paper", Name: "Paper", Type: domain.CostMaterial, ReferenceID: "MAT-paper", UsageMode: domain.UsageParameter, ParameterKey: "quantity", Multiplier: 2 * domain.QuantityScale},
			{ID: "CMP-labor", Name: "Design labor", Type: domain.CostLabor, UsageMode: domain.UsageParameter, ParameterKey: "estimated_hours", Multiplier: domain.QuantityScale, RateRial: 987654321, RateBasis: domain.RatePerHour},
			{ID: "CMP-overhead", Name: "Overhead", Type: domain.CostOverhead, Multiplier: domain.QuantityScale, Percentage: 7_125_001},
		},
	}, now)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	if err := store.SaveServiceDefinition(ctx, service); err != nil {
		t.Fatalf("save service: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	store, err = Open(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	got, err := store.GetService(ctx, service.ID)
	if err != nil {
		t.Fatalf("read service: %v", err)
	}
	if len(got.Parameters) != 3 || got.Parameters[0].Key != "quantity" || got.Parameters[1].Key != "estimated_hours" || got.Parameters[2].Key != "paper_size" {
		t.Fatalf("parameter ordering was not preserved: %+v", got.Parameters)
	}
	if got.Parameters[1].DefaultValue != "0.125001" || got.Parameters[1].MinValue == nil || got.Parameters[1].MinValue.String() != "0.125001" {
		t.Fatalf("decimal metadata was not round-tripped: %+v", got.Parameters[1])
	}
	if len(got.Parameters[2].Options) != 2 || got.Parameters[2].Options[1] != "A3" {
		t.Fatalf("choice options were not round-tripped: %+v", got.Parameters[2])
	}
	if len(got.Components) != 3 || got.Components[0].Name != "Paper" || got.Components[1].ParameterKey != "estimated_hours" || got.Components[2].Percentage.String() != "7.125001" || got.Components[0].Multiplier.String() != "2" {
		t.Fatalf("cost components were not round-tripped in order: %+v", got.Components)
	}

	bad := got
	bad.Name = "Should roll back"
	bad.UpdatedAt = now.Add(time.Hour)
	bad.Parameters = append([]domain.ServiceParameter(nil), got.Parameters...)
	bad.Parameters[1].Key = bad.Parameters[0].Key
	if err := store.SaveServiceDefinition(ctx, bad); err == nil {
		t.Fatal("duplicate parameter key did not fail")
	}
	unchanged, err := store.GetService(ctx, service.ID)
	if err != nil {
		t.Fatalf("read service after rollback: %v", err)
	}
	if unchanged.Name != "Digital Print" || len(unchanged.Parameters) != 3 || unchanged.Parameters[1].Key != "estimated_hours" {
		t.Fatalf("service definition was partially written: %+v", unchanged)
	}
	unchanged.Active = false
	if err := store.SaveServiceDefinition(ctx, unchanged); err != nil {
		t.Fatalf("archive service: %v", err)
	}
	active, err := store.ListServices(ctx, false)
	if err != nil || len(active) != 0 {
		t.Fatalf("active services = %+v, err=%v", active, err)
	}
	all, err := store.ListServices(ctx, true)
	if err != nil || len(all) != 1 || all[0].Active {
		t.Fatalf("all services = %+v, err=%v", all, err)
	}
	_ = store.Close()
}
