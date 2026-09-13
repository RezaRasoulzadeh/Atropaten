package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func productionFlowFixture(t *testing.T) (*Store, domain.Order, domain.ProductionJob) {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "production.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	ctx := context.Background()
	now := time.Now().UTC()
	m, err := domain.NewMaterial("MAT-paper", domain.MaterialDraft{Name: "Paper", PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale, PhysicalStock: 200 * domain.QuantityScale, AverageUnitCostRial: 100}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	o := domain.NewOrder("ORD-linked", "", now)
	o.CommercialStatus = domain.CommercialConfirmed
	// The actual legacy shape: a material selection carried by a CHOICE field.
	o.Items = []domain.OrderItem{{ID: "ITEM-linked", OrderID: o.ID, ServiceNameSnapshot: "Digital print", Quantity: 10 * domain.QuantityScale, QuantityUnit: "piece", ResolvedParametersJSON: `[{"key":"paper","type":"choice","value":"A4","materialId":"MAT-paper"}]`, CostBreakdownJSON: `[{"id":"legacy-material","type":"material","name":"Paper","enabled":true,"usageQuantity":"2"}]`, PricingSnapshotJSON: "{}", EstimatedCostRial: 2000, SuggestedPriceRial: 10000, SellingPriceRial: 10000}}
	if err = o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.CreateOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	j := domain.ProductionJob{ID: "JOB-linked", OrderID: o.ID, OrderItemID: o.Items[0].ID, Priority: "Normal", Status: "Pending", CreatedAt: now, UpdatedAt: now}
	if err = s.CreateProductionJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	j, err = s.GetProductionJob(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	o, err = s.GetOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	return s, o, j
}

func TestProductionOrderEditsOverridesConsumptionAndPartialOutsourcing(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	check := func(physical, reserved domain.Quantity) {
		t.Helper()
		state, err := s.InventoryState(ctx, "MAT-paper")
		if err != nil {
			t.Fatal(err)
		}
		if state.PhysicalStock != physical*domain.QuantityScale || state.ReservedStock != reserved*domain.QuantityScale {
			t.Fatalf("stock=%+v, expected physical=%d reserved=%d", state, physical, reserved)
		}
	}
	check(200, 20)
	plans, err := s.ProductionMaterials(ctx, j.ID)
	if err != nil || len(plans) != 1 {
		t.Fatalf("plans=%+v err=%v", plans, err)
	}
	// Edits preserve the item ID and synchronize the existing job immediately.
	o.Items[0].Quantity = 20 * domain.QuantityScale
	o.Items[0].EstimatedCostRial = 4000
	if err = o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	check(200, 40)
	current, err := s.GetProductionJob(ctx, j.ID)
	if err != nil || current.Quantity != 20*domain.QuantityScale || current.EstimatedCostRial != 4000 {
		t.Fatalf("job=%+v err=%v", current, err)
	}
	if err = s.UpdateReservation(ctx, plans[0].ReservationID, 44*domain.QuantityScale); err != nil {
		t.Fatal(err)
	}
	// Refresh must not overwrite an operator's extra allowance.
	if _, err = s.ProductionMaterials(ctx, j.ID); err != nil {
		t.Fatal(err)
	}
	check(200, 44)
	if err = s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	usage, err := s.RecordProductionConsumption(ctx, j.ID, "MAT-paper", "first", 24*domain.QuantityScale, 0, "run")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProductionMaterials(ctx, j.ID); err != nil {
		t.Fatal(err)
	}
	check(176, 20)
	if err = s.UpdateProductionConsumption(ctx, usage.ID, "correct", 20*domain.QuantityScale, 0, "corrected"); err != nil {
		t.Fatal(err)
	}
	check(180, 24)
	// Correction retries cannot duplicate movements.
	if err = s.UpdateProductionConsumption(ctx, usage.ID, "correct", 20*domain.QuantityScale, 0, "retry"); err != nil {
		t.Fatal(err)
	}
	check(180, 24)
	current, err = s.GetProductionJob(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	current.OutsourceQuantity = 10 * domain.QuantityScale
	current.OutsourceUnitCostRial = 300
	current.OutsourceFinancialAccountID = "missing"
	// A failed expense must roll back both material returns and outsourcing metadata.
	if err = s.UpdateProductionOutsourcing(ctx, current); err == nil {
		t.Fatal("invalid expense account accepted")
	}
	check(180, 24)
	current.OutsourceFinancialAccountID = "FIN-CASH"
	if err = s.UpdateProductionOutsourcing(ctx, current); err != nil {
		t.Fatal(err)
	}
	check(190, 12) // Half the usage returned; half the planned total remains in house.
	expense, err := s.GetExpense(ctx, "EXP-OUTSOURCE-"+j.ID)
	if err != nil || expense.AmountRial != 3000 {
		t.Fatalf("expense=%+v err=%v", expense, err)
	}
	cost, err := s.ProductionCostSummary(ctx, o.ID)
	if err != nil || cost != 4000 {
		t.Fatalf("cost=%d err=%v", cost, err)
	}
	view, err := application.NewOrdersService(s, s, nil).Get(ctx, o.ID)
	if err != nil || view.ActualCostRial != 4000 || view.ProjectedCostRial != 6000 || view.MarginRial != 4000 {
		t.Fatalf("order=%+v err=%v", view, err)
	}
	journals := countRows(t, s, "SELECT COUNT(*) FROM journal_entries")
	if err = s.UpdateProductionOutsourcing(ctx, current); err != nil {
		t.Fatal(err)
	}
	check(190, 12)
	if countRows(t, s, "SELECT COUNT(*) FROM journal_entries") != journals {
		t.Fatal("same outsourcing save duplicated accounting")
	}
	current.OutsourceUnitCostRial = 350
	if err = s.UpdateProductionOutsourcing(ctx, current); err != nil {
		t.Fatal(err)
	}
	if countRows(t, s, "SELECT COUNT(*) FROM expenses") != 1 {
		t.Fatal("cost edit duplicated expense")
	}
	current.OutsourceQuantity = 20 * domain.QuantityScale
	if err = s.UpdateProductionOutsourcing(ctx, current); err != nil {
		t.Fatal(err)
	}
	check(200, 0)
	cost, err = s.ProductionCostSummary(ctx, o.ID)
	if err != nil || cost != 7000 {
		t.Fatalf("full outsource cost=%d err=%v", cost, err)
	}
	// Reduce outsourcing: reserve material again without fabricating consumption.
	current.OutsourceQuantity = 10 * domain.QuantityScale
	if err = s.UpdateProductionOutsourcing(ctx, current); err != nil {
		t.Fatal(err)
	}
	check(200, 22)
	if err = s.DeleteProductionJob(ctx, j.ID); err != nil {
		t.Fatal(err)
	}
	check(200, 0)
	expense, err = s.GetExpense(ctx, "EXP-OUTSOURCE-"+j.ID)
	if err != nil || expense.Status != "Reversed" {
		t.Fatalf("deleted job expense=%+v err=%v", expense, err)
	}
	// Re-add uses the current order and does not resurrect the old operator override.
	j.ID = "JOB-readded"
	j.Quantity = 1
	j.Status = "Pending"
	if err = s.CreateProductionJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	check(200, 40)
	if err = s.ReleaseReservation(ctx, "RES-AUTO-"+j.ID+"-MAT-paper", "released"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProductionMaterials(ctx, j.ID); err != nil {
		t.Fatal(err)
	}
	check(200, 0)
	// Removing the item cancels its remaining work and retains its historical links.
	o.Items = nil
	if err = o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	if retained, err := s.GetProductionJob(ctx, j.ID); err != nil || retained.Status != domain.ProductionCancelled {
		t.Fatalf("retained job=%+v err=%v", retained, err)
	}
}

func TestProductionShortageAndOtherReservationsRemainProtected(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	// An unrelated allocation blocks stock even when no reservation is required for usage.
	if err := s.CreateReservation(ctx, domain.InventoryReservation{ID: "RES-other", MaterialID: "MAT-paper", Quantity: 150 * domain.QuantityScale, Status: "active"}); err != nil {
		t.Fatal(err)
	}
	if err := s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordProductionConsumption(ctx, j.ID, "MAT-paper", "blocked", 51*domain.QuantityScale, 0, ""); !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("other reservations not protected: %v", err)
	}
	if _, err := s.RecordProductionConsumption(ctx, j.ID, "MAT-paper", "allowed", 50*domain.QuantityScale, 0, ""); err != nil {
		t.Fatal(err)
	}
	stock, err := s.InventoryState(ctx, "MAT-paper")
	if err != nil || stock.ReservedStock != 150*domain.QuantityScale || stock.AvailableStock != 0 {
		t.Fatalf("stock=%+v err=%v", stock, err)
	}
	// Raising order requirements must show the shortage without over-reserving.
	o.Items[0].Quantity = 100 * domain.QuantityScale
	if err = o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	plans, err := s.ProductionMaterials(ctx, j.ID)
	if err != nil || len(plans) != 1 || plans[0].Shortage != "150" {
		t.Fatalf("plans=%+v err=%v", plans, err)
	}
}
