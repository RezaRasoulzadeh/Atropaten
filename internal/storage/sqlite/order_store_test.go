package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestM1004DatabaseUpgradesToM2001WithoutCatalogLoss(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m1004.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	for version := 1; version <= 4; version++ {
		if _, err := legacy.Exec(migrations[version-1].sql); err != nil {
			t.Fatalf("migration %d: %v", version, err)
		}
		if _, err := legacy.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (?, ?)`, version, "2026-01-01T00:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := legacy.Exec(`INSERT INTO materials(id,name,purchase_unit,consumption_unit,conversion_factor_units,physical_stock_units,reorder_level_units,average_unit_cost_rial,created_at,updated_at) VALUES('MAT-v4','Paper','pack','sheet',500000000,1000000,100000,987654321,'2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`INSERT INTO services(id,name,created_at,updated_at) VALUES('SVC-v4','Legacy service','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	material, err := store.Get(context.Background(), "MAT-v4")
	if err != nil || material.AverageUnitCostRial != 987654321 {
		t.Fatalf("material lost in v4 upgrade: %+v, %v", material, err)
	}
	service, err := store.GetService(context.Background(), "SVC-v4")
	if err != nil || service.Name != "Legacy service" {
		t.Fatalf("service lost in v4 upgrade: %+v, %v", service, err)
	}
}

func TestCustomerAndOrderPersistenceKeepsSnapshotsAndOrderNumbers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "orders.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Date(2026, time.January, 12, 8, 0, 0, 0, time.UTC)
	customer, err := domain.NewCustomer("CUS-1", domain.CustomerDraft{Name: "Mehr Studio", Phone: "+98 21 1"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveCustomer(ctx, customer); err != nil {
		t.Fatal(err)
	}
	order := domain.NewOrder("ORDER-1", customer.ID, now)
	order.CustomerNameSnapshot = customer.Name
	order.CustomerPhoneSnapshot = customer.Phone
	order.Items = []domain.OrderItem{
		{ID: "ITEM-2", OrderID: order.ID, Position: 0, ServiceID: "SVC-design", ServiceNameSnapshot: "Graphic Design", Quantity: 250001, QuantityUnit: "hour", ResolvedParametersJSON: `[{"key":"estimated_hours","value":"0.250001"}]`, CostBreakdownJSON: `[{"name":"Labor","amountRial":100001}]`, PricingSnapshotJSON: `{"estimatedCostRial":100001,"effectiveSellingPriceRial":250001}`, EstimatedCostRial: 100001, SuggestedPriceRial: 225001, SellingPriceRial: 250001},
		{ID: "ITEM-1", OrderID: order.ID, Position: 1, ServiceID: "SVC-print", ServiceNameSnapshot: "Digital Print", Quantity: 500, QuantityUnit: "sheet", ResolvedParametersJSON: `[{"key":"quantity","value":"500"}]`, CostBreakdownJSON: `[{"name":"Paper","amountRial":125000}]`, PricingSnapshotJSON: `{"estimatedCostRial":125000,"effectiveSellingPriceRial":300000}`, EstimatedCostRial: 125000, SuggestedPriceRial: 300000, SellingPriceRial: 300000},
	}
	if err := order.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	first, err := store.GetOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.OrderNumber != "ORD-1001" || first.TotalRial != 550001 || first.EstimatedCostRial != 225001 || len(first.Items) != 2 {
		t.Fatalf("unexpected persisted order: %+v", first)
	}
	if first.Items[0].PricingSnapshotJSON != order.Items[0].PricingSnapshotJSON || first.Items[1].Position != 1 {
		t.Fatal("item snapshot/order was changed")
	}
	second := domain.NewOrder("ORDER-2", "", now.Add(time.Minute))
	if err := second.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateOrder(ctx, second); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	got, err := reopened.GetOrder(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.OrderNumber != first.OrderNumber || got.Items[0].ServiceNameSnapshot != "Graphic Design" || got.Items[1].SellingPriceRial != 300000 {
		t.Fatalf("reopen changed order: %+v", got)
	}
	customer.Active = false
	customer.UpdatedAt = now.Add(time.Hour)
	if err := reopened.SaveCustomer(ctx, customer); err != nil {
		t.Fatal(err)
	}
	archived, err := reopened.GetCustomer(ctx, customer.ID)
	if err != nil || archived.Active {
		t.Fatalf("customer archive = %+v, err=%v", archived, err)
	}
	if historical, err := reopened.GetOrder(ctx, first.ID); err != nil || historical.CustomerNameSnapshot != "Mehr Studio" {
		t.Fatalf("historical customer snapshot unreadable: %+v, %v", historical, err)
	}
}
