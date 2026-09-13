package sqlite

import (
	"context"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestOrderPricingRoundsChargesAndPreservesExactCost(t *testing.T) {
	s, _, _ := productionFlowFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()
	service, err := domain.NewService("SVC-round-charge", domain.ServiceDraft{
		Name:        "Rounded charge",
		Components:  []domain.ServiceCostComponentDraft{{ID: "COMP-round-charge", Name: "Cost", Type: domain.CostFixed, RateRial: 1671942, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true}},
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingMarkup, MarkupPercentage: 20 * domain.QuantityScale},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.SaveServiceDefinition(ctx, service); err != nil {
		t.Fatal(err)
	}
	order := domain.NewOrder("ORD-round-charge", "", now)
	if err = s.CreateOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	orders := application.NewOrdersService(s, s, application.NewPricingService(s, s, s))
	view, err := orders.AddItem(ctx, order.ID, application.OrderItemInput{ServiceID: service.ID, Quantity: "0.5", QuantityUnit: "job"})
	if err != nil || view.EstimatedCostRial != 835971 || view.TotalRial != 1004000 || view.MarginRial != 168000 || view.ProjectedCostRial != 836000 {
		t.Fatalf("rounded order=%+v err=%v", view, err)
	}
	invoice, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, order.ID)
	if err != nil || invoice.TotalRial != view.TotalRial {
		t.Fatalf("invoice does not match rounded order: %+v err=%v", invoice, err)
	}
}

func TestCompletedProductionOrderEditReconcilesAdditionalRequirements(t *testing.T) {
	s, order, job := productionFlowFixture(t)
	ctx := context.Background()
	if err := s.TransitionProductionJob(ctx, job.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err := s.TransitionProductionJob(ctx, job.ID, "Completed"); err != nil {
		t.Fatal(err)
	}
	order.Items[0].Quantity = 20 * domain.QuantityScale
	order.Items[0].EstimatedCostRial = 4000
	if err := order.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	current, err := s.GetProductionJob(ctx, job.ID)
	if err != nil || current.Quantity != 20*domain.QuantityScale || current.ActualMaterialCostRial != 2000 || current.RemainingMaterialCostRial != 2000 || current.ProjectedCostRial != 4000 {
		t.Fatalf("reopened forecast=%+v err=%v", current, err)
	}
}

func TestProductionForecastReconcilesQuantityAndCostWithoutMutating(t *testing.T) {
	s, order, job := productionFlowFixture(t)
	ctx := context.Background()
	order.Items[0].CostBreakdownJSON = `[{"type":"material","enabled":true,"materialId":"MAT-paper","usageQuantity":"2","amountRial":200},{"type":"machine","enabled":true,"amountRial":100}]`
	order.Items[0].EstimatedCostRial = 3000
	if err := order.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	// Simulate an older job whose persisted quantities have not been reconciled.
	if _, err := s.db.Exec(`UPDATE production_jobs SET quantity_units=?,estimated_cost_rial=1500 WHERE id=?`, 5*domain.QuantityScale, job.ID); err != nil {
		t.Fatal(err)
	}
	current, err := s.GetProductionJob(ctx, job.ID)
	if err != nil || current.Quantity != 10*domain.QuantityScale || current.EstimatedCostRial != 3000 || current.EstimatedConversionCostRial != 1000 || current.RemainingMaterialCostRial != 2000 || current.ProjectedCostRial != 3000 {
		t.Fatalf("reconciled forecast=%+v err=%v", current, err)
	}
	var persisted domain.Quantity
	if err = s.db.QueryRow(`SELECT quantity_units FROM production_jobs WHERE id=?`, job.ID).Scan(&persisted); err != nil || persisted != 5*domain.QuantityScale {
		t.Fatalf("forecast mutated storage: %d err=%v", persisted, err)
	}
}
