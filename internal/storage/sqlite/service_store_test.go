package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
	"Atropaten/internal/platform"
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
	if version != platform.CurrentSchemaVersion {
		t.Fatalf("migration version = %d, want %d", version, platform.CurrentSchemaVersion)
	}
	material, err := store.Get(context.Background(), "MAT-legacy")
	if err != nil {
		t.Fatalf("read legacy material: %v", err)
	}
	if material.AverageUnitCostRial != 123456789 || material.PhysicalStock != domain.Quantity(1250000) || material.Kind != domain.MaterialKindGenericConsumable || len(material.Attributes) != 0 {
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

func TestM1003DatabaseUpgradesToM1004WithoutChangingComponents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m1003.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open M1-003 database: %v", err)
	}
	if _, err := legacy.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	for version := 1; version <= 3; version++ {
		if _, err := legacy.Exec(migrations[version-1].sql); err != nil {
			t.Fatalf("apply legacy migration %d: %v", version, err)
		}
		if _, err := legacy.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (?, ?)`, version, "2026-01-01T00:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := legacy.Exec(`INSERT INTO services (id, name, created_at, updated_at) VALUES ('SVC-legacy', 'Legacy service', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`INSERT INTO service_cost_components (id, service_id, component_name, component_type, usage_mode, multiplier_units, rate_rial, display_order, created_at, updated_at) VALUES ('C-legacy', 'SVC-legacy', 'Legacy fixed', 'fixed', 'fixed', 1000000, 500, 0, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatalf("upgrade M1-003 database: %v", err)
	}
	defer store.Close()
	service, err := store.GetService(context.Background(), "SVC-legacy")
	if err != nil {
		t.Fatalf("read legacy service: %v", err)
	}
	if len(service.Components) != 1 || service.Components[0].UsageQuantity != domain.Quantity(domain.QuantityScale) || service.PricingRule != nil {
		t.Fatalf("legacy component changed during M1-004 migration: %+v", service)
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
		Name: "Digital Print", Code: "PRINT", Category: "Production", Description: "Generic print service", ImagePath: "data:image/png;base64,preview", DefaultUnit: "sheet", DefaultPriority: domain.PriorityHigh,
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
	if got.ImagePath != "data:image/png;base64,preview" || got.DefaultUnit != "sheet" || got.DefaultPriority != domain.PriorityHigh {
		t.Fatalf("service basic metadata was not round-tripped: %+v", got)
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
	if err := store.SetServiceActive(ctx, unchanged.ID, false, now.Add(time.Minute)); err != nil {
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

func TestFinishedSizeDefinitionRoundTrip(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "finished-size.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-finished", domain.ServiceDraft{
		Name: "Finished size service",
		Parameters: []domain.ServiceParameterDraft{
			{ID: "P-quantity", Key: "quantity", Label: "Quantity", Type: domain.ParameterInteger},
			{ID: "P-size", Key: "size", Label: "Size", Type: domain.ParameterChoice, Options: []string{"a4"}},
		},
		FinishedSize: &domain.ServiceFinishedSizeDefinition{
			ParameterKey: "size", QuantityParameterKey: "quantity", AllowRotation: true,
			Options: []domain.FinishedSizeOption{{ID: "SIZE-a4", Code: "a4", Label: "A4", WidthMM: 210 * domain.QuantityScale, HeightMM: 297 * domain.QuantityScale, Position: 0, Active: true}},
		},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveServiceDefinition(context.Background(), service); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetService(context.Background(), service.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.FinishedSize == nil || got.FinishedSize.QuantityParameterKey != "quantity" || len(got.FinishedSize.Options) != 1 || got.FinishedSize.Options[0].WidthMM != 210*domain.QuantityScale {
		t.Fatalf("finished size did not round-trip: %+v", got.FinishedSize)
	}
}

func TestPredefinedPrintSizeCatalogAndServiceParameterRoundTrip(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "predefined-size.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	definitions, err := store.ListPredefinedParameters(context.Background())
	if err != nil || len(definitions) < 3 || definitions[0].Key != domain.PredefinedParameterPrintSize || definitions[1].Key != domain.PredefinedParameterPaperType || definitions[2].Key != domain.PredefinedParameterColor {
		t.Fatalf("predefined catalog = %+v, err=%v", definitions, err)
	}
	if len(definitions[1].Options) == 0 || definitions[1].Options[0].Code != "uncoated" || definitions[1].Options[0].WidthMM != nil {
		t.Fatalf("paper type catalog = %+v", definitions[1])
	}
	if len(definitions[2].Options) == 0 || definitions[2].Options[0].Code != "black-and-white" || definitions[2].Options[0].Label != "Black & white" {
		t.Fatalf("color catalog = %+v", definitions[2])
	}
	var a4 domain.PredefinedParameterOption
	for _, option := range definitions[0].Options {
		if option.Code == "a4" {
			a4 = option
		}
	}
	if a4.WidthMM == nil || a4.HeightMM == nil || *a4.WidthMM != 210*domain.QuantityScale {
		t.Fatalf("A4 catalog option = %+v", a4)
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-predefined", domain.ServiceDraft{
		Name: "Predefined print size",
		Parameters: []domain.ServiceParameterDraft{
			{ID: "P-quantity", Key: "quantity", Label: "Quantity", Type: domain.ParameterInteger},
			{ID: "P-size", Key: "print_size", Label: "Print size", Type: domain.ParameterChoice, PredefinedKey: domain.PredefinedParameterPrintSize},
		},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveServiceDefinition(context.Background(), service); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetService(context.Background(), service.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Parameters[1].PredefinedKey != domain.PredefinedParameterPrintSize || len(got.Parameters[1].PredefinedOptions) == 0 || got.FinishedSize == nil {
		t.Fatalf("predefined service did not hydrate: %+v finished=%+v", got.Parameters[1], got.FinishedSize)
	}
	width, height, err := got.ResolveFinishedDimensions(map[string]string{"print_size": "a4"})
	if err != nil || width != 210*domain.QuantityScale || height != 297*domain.QuantityScale {
		t.Fatalf("resolved A4 dimensions = %v x %v, err=%v", width, height, err)
	}
}

func TestMaterialBackedServiceParameterRoundTrip(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "material-backed-service.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-material-backed", domain.ServiceDraft{Name: "Paper service", Parameters: []domain.ServiceParameterDraft{{ID: "P-grammage", Key: "grammage", Label: "Grammage", Type: domain.ParameterChoice, Required: true, MaterialSource: &domain.MaterialParameterSource{AllowedKinds: []domain.MaterialKind{domain.MaterialKindSheetStock}, ExposedAttributeKey: "grammage_gsm", AllowedValues: []domain.MaterialAttributeValue{{Key: "grammage_gsm", ValueType: domain.MaterialAttributeInteger, IntegerValue: 170}}, AdditionalFilters: []domain.MaterialAttributeFilter{{Key: "finish", Value: domain.MaterialAttributeValue{Key: "finish", ValueType: domain.MaterialAttributeEnum, EnumCode: "matte"}}}}}}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveServiceDefinition(context.Background(), service); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetService(context.Background(), service.ID)
	if err != nil {
		t.Fatal(err)
	}
	source := got.Parameters[0].MaterialSource
	if source == nil || source.ExposedAttributeKey != "grammage_gsm" || len(source.AllowedKinds) != 1 || len(source.AllowedValues) != 1 || source.AllowedValues[0].IntegerValue != 170 || len(source.AdditionalFilters) != 1 || source.AdditionalFilters[0].Value.EnumCode != "matte" {
		t.Fatalf("material source=%+v", source)
	}
}

func TestServiceMaterialVariantRoundTrip(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "material-variant.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-variant", domain.ServiceDraft{Name: "Variant service", Parameters: []domain.ServiceParameterDraft{{ID: "P-finish", Key: "finish", Label: "Finish", Type: domain.ParameterChoice, MaterialSource: &domain.MaterialParameterSource{ExposedAttributeKey: "finish"}}}, MaterialVariants: []domain.ServiceMaterialVariant{{ID: "VAR-1", MaterialID: "MAT-1", Values: map[string]string{"finish": "matte"}, Position: 0, Active: true}}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveServiceDefinition(context.Background(), service); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetService(context.Background(), service.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.MaterialVariants) != 1 || got.MaterialVariants[0].MaterialID != "MAT-1" || got.MaterialVariants[0].Values["finish"] != "matte" {
		t.Fatalf("material variants did not round-trip: %+v", got.MaterialVariants)
	}
}

func TestPricingRuleAndUsageQuantityRoundTripAfterV4Migration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pricing.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-pricing", domain.ServiceDraft{
		Name:        "Digital Print",
		Parameters:  []domain.ServiceParameterDraft{{ID: "P-qty", Key: "quantity", Label: "Quantity", Type: domain.ParameterInteger}},
		Components:  []domain.ServiceCostComponentDraft{{ID: "C-paper", Name: "Paper", Type: domain.CostFixed, UsageQuantity: domain.Quantity(2_500_001), Multiplier: domain.Quantity(1_125_001), RateRial: 987654321, Enabled: true}},
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingTiers, ParameterKey: "quantity", Tiers: []domain.ServicePricingTierDraft{{MinimumQuantity: 0, PriceRial: 111}, {MinimumQuantity: 10 * domain.QuantityScale, PriceRial: 222}}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := store.SaveServiceDefinition(context.Background(), service); err != nil {
		t.Fatalf("save service: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	store, err = Open(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer store.Close()
	got, err := store.GetService(context.Background(), service.ID)
	if err != nil {
		t.Fatalf("get service: %v", err)
	}
	if got.PricingRule == nil || got.PricingRule.Type != domain.PricingTiers || len(got.PricingRule.Tiers) != 2 || got.PricingRule.Tiers[1].MinimumQuantity != 10*domain.QuantityScale {
		t.Fatalf("pricing rule not round-tripped: %+v", got.PricingRule)
	}
	if got.Components[0].UsageQuantity != domain.Quantity(2_500_001) || got.Components[0].Multiplier != domain.Quantity(1_125_001) {
		t.Fatalf("usage quantities lost: %+v", got.Components[0])
	}
}
