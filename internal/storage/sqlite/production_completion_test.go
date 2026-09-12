package sqlite

import (
	"context"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestCompletionRecordsRemainingPlanExactlyOnce(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	if err := s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordProductionConsumption(ctx, j.ID, "MAT-paper", "early", 4*domain.QuantityScale, domain.QuantityScale, ""); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 2; n++ {
		if err := s.TransitionProductionJob(ctx, j.ID, "Completed"); err != nil {
			t.Fatal(err)
		}
	}
	state, err := s.InventoryState(ctx, "MAT-paper")
	if err != nil || state.PhysicalStock != 180*domain.QuantityScale || state.ReservedStock != 0 {
		t.Fatalf("stock=%+v err=%v", state, err)
	}
	job, err := s.GetProductionJob(ctx, j.ID)
	if err != nil || job.ActualMaterialCostRial != 1900 || job.ActualWasteCostRial != 100 || job.ProjectedCostRial != 2000 {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	order, err := s.GetOrder(ctx, o.ID)
	if err != nil || order.FulfillmentStatus != domain.FulfillmentReady {
		t.Fatalf("order=%+v err=%v", order, err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, "Completed"); err != nil {
		t.Fatal(err)
	}
	if countRows(t, s, `SELECT COUNT(*) FROM production_consumptions`) != 2 {
		t.Fatal("reopening duplicated usage")
	}
}

func TestCompletionShortageRollsBackAllMaterialsAndStatus(t *testing.T) {
	s, _, j := productionFlowFixture(t)
	ctx := context.Background()
	if err := s.SetProductionMaterialTarget(ctx, j.ID, "MAT-paper", 201*domain.QuantityScale); err != nil {
		t.Fatal(err)
	}
	if err := s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err := s.TransitionProductionJob(ctx, j.ID, "Completed"); err == nil {
		t.Fatal("completed with shortage")
	}
	job, err := s.GetProductionJob(ctx, j.ID)
	if err != nil || job.Status != "In Progress" || job.ActualMaterialCostRial != 0 {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	if countRows(t, s, `SELECT COUNT(*) FROM production_consumptions`) != 0 {
		t.Fatal("partial consumption committed")
	}
}

func TestProjectedMarginIncludesUnfinishedMaterialsConversionAndDiscount(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	o.Items[0].CostBreakdownJSON = `[{"type":"material","enabled":true,"materialId":"MAT-paper","usageQuantity":"2","amountRial":200},{"type":"machine","enabled":true,"amountRial":50},{"type":"labor","enabled":true,"amountRial":30},{"type":"overhead","enabled":true,"amountRial":20},{"type":"waste","enabled":true,"amountRial":10}]`
	o.Items[0].EstimatedCostRial = 3100
	o.DiscountRial = 1000
	if err := o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	view, err := application.NewOrdersService(s, s, nil).Get(ctx, o.ID)
	if err != nil || view.ProjectedCostRial != 3000 || view.ActualCostRial != 0 || view.MarginRial != 6000 || view.MarginPercentage != "66.67" {
		t.Fatalf("view=%+v err=%v", view, err)
	}
	job, err := s.GetProductionJob(ctx, j.ID)
	if err != nil || job.EstimatedConversionCostRial != 1000 || job.RemainingMaterialCostRial != 2000 {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	job.OutsourceQuantity = 5 * domain.QuantityScale
	job.OutsourceUnitCostRial = 400
	job.OutsourceFinancialAccountID = "FIN-CASH"
	if err = s.UpdateProductionOutsourcing(ctx, job); err != nil {
		t.Fatal(err)
	}
	job, err = s.GetProductionJob(ctx, j.ID)
	if err != nil || job.ProjectedCostRial != 3500 || job.EstimatedConversionCostRial != 500 || job.RemainingMaterialCostRial != 1000 {
		t.Fatalf("partial outsource=%+v err=%v", job, err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, "Completed"); err != nil {
		t.Fatal(err)
	}
	job, err = s.GetProductionJob(ctx, j.ID)
	if err != nil || job.ProjectedCostRial != 3500 || job.RemainingMaterialCostRial != 0 || job.ActualMaterialCostRial != 1000 {
		t.Fatalf("completed=%+v err=%v", job, err)
	}
}

func TestLateProductionCostAdjustsPostedInvoiceAndVoidsCleanly(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	inv, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, inv.ID); err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, "Completed"); err != nil {
		t.Fatal(err)
	}
	check := func(want int64) {
		t.Helper()
		var got int64
		if err := s.db.QueryRow(`SELECT COALESCE(SUM(debit_rial-credit_rial),0) FROM journal_lines WHERE account_id='ACC-COGS'`).Scan(&got); err != nil || got != want {
			t.Fatalf("COGS=%d want=%d err=%v", got, want, err)
		}
	}
	check(2000)
	// Reversing stock after invoicing also corrects financial profit.
	usage, err := s.ListProductionConsumptions(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.ReverseProductionConsumption(ctx, usage[0].ID, "Return"); err != nil {
		t.Fatal(err)
	}
	check(0)
	if err = s.VoidInvoice(ctx, inv.ID); err != nil {
		t.Fatal(err)
	}
	check(0)
}

func TestDeletingInvoicedProductionReversesRecognizedCost(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	if err := s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err := s.TransitionProductionJob(ctx, j.ID, "Completed"); err != nil {
		t.Fatal(err)
	}
	inv, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, inv.ID); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteProductionJob(ctx, j.ID); err != nil {
		t.Fatal(err)
	}
	var cost int64
	if err = s.db.QueryRow(`SELECT SUM(debit_rial-credit_rial) FROM journal_lines WHERE account_id='ACC-COGS'`).Scan(&cost); err != nil || cost != 0 {
		t.Fatalf("COGS=%d err=%v", cost, err)
	}
}

func TestProductionUsesExactInventoryValueAndReturnsExactCost(t *testing.T) {
	s, _, j := productionFlowFixture(t)
	ctx := context.Background()
	// A three-unit pool valued at two Rial has a rounded unit cost of one Rial.
	m, err := domain.NewMaterial("MAT-rounding", domain.MaterialDraft{Name: "Rounding", PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale, AverageUnitCostRial: 1}, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES('MOV-rounding','MAT-rounding',?,'opening_balance',3000000,1,2,'opening_balance','MAT-rounding','Exact pool value',?)`, time.Now().UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	u, err := s.RecordProductionConsumption(ctx, j.ID, m.ID, "rounding", 2*domain.QuantityScale, domain.QuantityScale, "")
	if err != nil || u.MaterialCostRial+u.WasteCostRial != 2 {
		t.Fatalf("usage=%+v err=%v", u, err)
	}
	state, err := s.InventoryState(ctx, m.ID)
	if err != nil || state.InventoryValueRial != 0 || state.PhysicalStock != 0 {
		t.Fatalf("stock=%+v err=%v", state, err)
	}
	if err = s.ReverseProductionConsumption(ctx, u.ID, "Return exact value"); err != nil {
		t.Fatal(err)
	}
	state, err = s.InventoryState(ctx, m.ID)
	if err != nil || state.InventoryValueRial != 2 || state.PhysicalStock != 3*domain.QuantityScale {
		t.Fatalf("returned=%+v err=%v", state, err)
	}
}

func TestServiceReportIncludesOutsourcingAndInvoiceDiscount(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	o.DiscountRial = 1001
	if err := o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	j.OutsourceQuantity = j.Quantity
	j.OutsourceUnitCostRial = 300
	j.OutsourceFinancialAccountID = "FIN-CASH"
	if err := s.UpdateProductionOutsourcing(ctx, j); err != nil {
		t.Fatal(err)
	}
	inv, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, inv.ID); err != nil {
		t.Fatal(err)
	}
	r, err := s.reportSalesByService(ctx, domain.Report{}, time.Now().Add(-24*time.Hour).Format(time.RFC3339Nano), time.Now().Add(24*time.Hour).Format(time.RFC3339Nano))
	if err != nil || len(r.Rows) != 1 || r.Rows[0].AmountRial != 8999 || r.Rows[0].TertiaryAmountRial != 3000 {
		t.Fatalf("report=%+v err=%v", r, err)
	}
}
