package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

func TestV7UpgradePreservesInventoryAndPurchasingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v7.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = legacy.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	for version := 1; version <= 7; version++ {
		if _, err = legacy.Exec(migrations[version-1].sql); err != nil {
			t.Fatalf("migration %d: %v", version, err)
		}
		if _, err = legacy.Exec(`INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, version, "2026-01-01T00:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = legacy.Exec(`INSERT INTO materials(id,name,purchase_unit,consumption_unit,conversion_factor_units,physical_stock_units,reorder_level_units,average_unit_cost_rial,created_at,updated_at) VALUES('MAT-v7','Legacy paper','pack','sheet',1000000,3000000,100000,250,'2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err = legacy.Exec(`INSERT INTO suppliers(id,name,created_at,updated_at) VALUES('SUP-v7','Legacy supplier','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err = legacy.Exec(`INSERT INTO purchases(id,purchase_number,supplier_id,purchase_date,status,subtotal_rial,discount_rial,shipping_rial,tax_rial,additional_costs_rial,total_rial,created_at,updated_at) VALUES('PUR-v7','PUR-1001','SUP-v7','2026-01-01','Posted',250,0,0,0,0,250,'2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err = legacy.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	material, err := store.Get(context.Background(), "MAT-v7")
	if err != nil {
		t.Fatal(err)
	}
	if material.PhysicalStock != 3*domain.QuantityScale || material.AverageUnitCostRial != 250 {
		t.Fatalf("legacy material changed: %+v", material)
	}
	if _, err = store.GetSupplier(context.Background(), "SUP-v7"); err != nil {
		t.Fatalf("legacy supplier lost: %v", err)
	}
	var version int
	if err = store.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil || version != 11 {
		t.Fatalf("migration version=%d err=%v", version, err)
	}
	if got := countRows(t, store, `SELECT COUNT(*) FROM inventory_reservations`); got != 0 {
		t.Fatalf("unexpected reservations after migration: %d", got)
	}
	if got := countRows(t, store, `SELECT COUNT(*) FROM inventory_movements WHERE movement_type='opening_balance'`); got != 1 {
		t.Fatalf("legacy opening movement count=%d", got)
	}
}

func TestReservationsAndProductionLedgerAreAtomicAndIdempotent(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "production.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Date(2026, time.January, 12, 8, 0, 0, 0, time.UTC)

	material, err := domain.NewMaterial("MAT-prod", domain.MaterialDraft{
		Name: "Paper", PurchaseUnit: "pack", ConsumptionUnit: "sheet", ConversionFactor: 100 * domain.QuantityScale,
		PhysicalStock: 10 * domain.QuantityScale, AverageUnitCostRial: 100, ReorderLevel: domain.QuantityScale,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Create(ctx, material); err != nil {
		t.Fatal(err)
	}
	order := domain.NewOrder("ORD-prod", "", now)
	order.CommercialStatus = domain.CommercialConfirmed
	order.Items = []domain.OrderItem{{ID: "ITEM-prod", OrderID: order.ID, Position: 0, ServiceNameSnapshot: "Print", Quantity: 2 * domain.QuantityScale, QuantityUnit: "sheet", ResolvedParametersJSON: "{}", CostBreakdownJSON: "[]", PricingSnapshotJSON: `{"sellingPriceRial":500}`, EstimatedCostRial: 100, SuggestedPriceRial: 500, SellingPriceRial: 500}}
	if err = order.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = store.CreateOrder(ctx, order); err != nil {
		t.Fatal(err)
	}

	job := domain.ProductionJob{ID: "JOB-prod", OrderID: order.ID, OrderItemID: "ITEM-prod", Quantity: 2 * domain.QuantityScale, QuantityUnit: "sheet", Status: domain.ProductionPending, Priority: string(domain.PriorityNormal), CreatedAt: now, UpdatedAt: now}
	if err = store.CreateProductionJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	if got := countRows(t, store, `SELECT COUNT(*) FROM inventory_movements`); got != 1 {
		t.Fatalf("job creation changed movement ledger: %d", got)
	}

	reservation := domain.InventoryReservation{ID: "RES-prod", MaterialID: material.ID, OrderID: order.ID, OrderItemID: "ITEM-prod", ProductionJobID: job.ID, Quantity: 4 * domain.QuantityScale, Status: domain.ReservationActive, CreatedAt: now, UpdatedAt: now}
	if err = store.CreateReservation(ctx, reservation); err != nil {
		t.Fatal(err)
	}
	if err = store.UpdateReservation(ctx, reservation.ID, 3*domain.QuantityScale); err != nil {
		t.Fatal(err)
	}
	releasable := reservation
	releasable.ID = "RES-release"
	releasable.Quantity = domain.QuantityScale
	if err = store.CreateReservation(ctx, releasable); err != nil {
		t.Fatal(err)
	}
	if err = store.ReleaseReservation(ctx, releasable.ID, domain.ReservationReleased); err != nil {
		t.Fatal(err)
	}
	state, err := store.InventoryState(ctx, material.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.PhysicalStock != 10*domain.QuantityScale || state.ReservedStock != 3*domain.QuantityScale || state.AvailableStock != 7*domain.QuantityScale {
		t.Fatalf("unexpected reserved state: %+v", state)
	}
	if got := countRows(t, store, `SELECT COUNT(*) FROM inventory_movements`); got != 1 {
		t.Fatalf("reservation created a movement: %d", got)
	}
	tooMuch := reservation
	tooMuch.ID = "RES-too-much"
	tooMuch.Quantity = 8 * domain.QuantityScale
	if err = store.CreateReservation(ctx, tooMuch); !errors.Is(err, domain.ErrReservationExceeded) {
		t.Fatalf("over-reservation error=%v", err)
	}

	if err = store.TransitionProductionJob(ctx, job.ID, domain.ProductionInProgress); err != nil {
		t.Fatal(err)
	}
	editable, err := store.GetProductionJob(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	editable.ActualOutsourcedCostRial = 50
	editable.OutsourceDescription = "Finishing"
	if err = store.UpdateProductionJob(ctx, editable); err != nil {
		t.Fatal(err)
	}
	consumption, err := store.RecordProductionConsumption(ctx, job.ID, material.ID, "attempt-1", 2*domain.QuantityScale, domain.QuantityScale, "trim")
	if err != nil {
		t.Fatal(err)
	}
	if consumption.MaterialCostRial != 200 || consumption.WasteCostRial != 100 {
		t.Fatalf("unexpected fixed-scale costs: %+v", consumption)
	}
	if got := countRows(t, store, `SELECT COUNT(*) FROM inventory_movements`); got != 3 {
		t.Fatalf("consumption movement count=%d, want 3", got)
	}
	if _, err = store.RecordProductionConsumption(ctx, job.ID, material.ID, "attempt-1", 2*domain.QuantityScale, domain.QuantityScale, "retry"); err != nil {
		t.Fatal(err)
	}
	if got := countRows(t, store, `SELECT COUNT(*) FROM inventory_movements`); got != 3 {
		t.Fatalf("idempotent retry duplicated movements: %d", got)
	}
	state, err = store.InventoryState(ctx, material.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.PhysicalStock != 7*domain.QuantityScale || state.ReservedStock != 0 || state.AvailableStock != 7*domain.QuantityScale {
		t.Fatalf("unexpected consumed state: %+v", state)
	}

	if err = store.TransitionProductionJob(ctx, job.ID, domain.ProductionCompleted); err != nil {
		t.Fatal(err)
	}
	updated, err := store.GetProductionJob(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.ProductionCompleted || updated.ActualMaterialCostRial != 200 || updated.ActualWasteCostRial != 100 || updated.ActualOutsourcedCostRial != 50 {
		t.Fatalf("unexpected completed job: %+v", updated)
	}
	state, err = store.InventoryState(ctx, material.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.ReservedStock != 0 {
		t.Fatalf("completion did not release remaining reservation: %+v", state)
	}
	completedOrder, err := store.GetOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completedOrder.FulfillmentStatus != domain.FulfillmentReady {
		t.Fatalf("order fulfillment=%s, want Ready", completedOrder.FulfillmentStatus)
	}

	if err = store.ReverseProductionConsumption(ctx, consumption.ID, "count correction"); err != nil {
		t.Fatal(err)
	}
	if got := countRows(t, store, `SELECT COUNT(*) FROM inventory_movements`); got != 5 {
		t.Fatalf("correction did not append compensating movements: %d", got)
	}
	if err = store.ReverseProductionConsumption(ctx, consumption.ID, "retry correction"); err == nil {
		t.Fatal("duplicate correction unexpectedly succeeded")
	}
	if _, err = store.db.Exec(`UPDATE inventory_movements SET note='mutated' WHERE id=?`, "MOV-"+consumption.ID+"-CONSUMED"); err == nil {
		t.Fatal("immutable movement update unexpectedly succeeded")
	}
	job2 := job
	job2.ID = "JOB-cancel"
	if err = store.CreateProductionJob(ctx, job2); err != nil {
		t.Fatal(err)
	}
	reservation2 := reservation
	reservation2.ID = "RES-cancel"
	reservation2.ProductionJobID = job2.ID
	reservation2.Quantity = domain.QuantityScale
	if err = store.CreateReservation(ctx, reservation2); err != nil {
		t.Fatal(err)
	}
	if err = store.TransitionProductionJob(ctx, job2.ID, domain.ProductionCancelled); err != nil {
		t.Fatal(err)
	}
	state, err = store.InventoryState(ctx, material.ID)
	if err != nil || state.ReservedStock != 0 {
		t.Fatalf("cancel did not release reservation: %+v, %v", state, err)
	}
	job3 := job
	job3.ID = "JOB-draft-delete"
	job3.Quantity = 0
	job3.QuantityUnit = ""
	if err = store.CreateProductionJob(ctx, job3); err != nil {
		t.Fatal(err)
	}
	if err = store.DeleteProductionJob(ctx, job3.ID); err != nil {
		t.Fatalf("safe draft delete failed: %v", err)
	}
	if _, err = store.GetProductionJob(ctx, job3.ID); !errors.Is(err, domain.ErrProductionJobNotFound) {
		t.Fatalf("deleted draft still exists: %v", err)
	}
}

func countRows(t *testing.T, store *Store, query string) int {
	t.Helper()
	var count int
	if err := store.db.QueryRow(query).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
