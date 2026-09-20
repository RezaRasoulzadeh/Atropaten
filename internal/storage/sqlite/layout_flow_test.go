package sqlite

import (
	"Atropaten/internal/application"
	"Atropaten/internal/domain"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func TestLegacyRollServicePricesAndAddsWithoutSavedLayout(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	store, err := Open(filepath.Join(t.TempDir(), "legacy-roll.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	material, err := domain.NewMaterial("legacy-roll-stock", domain.MaterialDraft{
		Name: "Banner 320", Kind: domain.MaterialKindRollMedia, PurchaseUnit: "meter", ConsumptionUnit: "meter",
		ConversionFactor: domain.QuantityScale, PhysicalStock: 100 * domain.QuantityScale, AverageUnitCostRial: 100,
		Attributes: []domain.MaterialAttributeValue{{Key: "width_mm", ValueType: domain.MaterialAttributeDecimal, DecimalValue: 3200 * domain.QuantityScale}},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Create(ctx, material); err != nil {
		t.Fatal(err)
	}
	service, err := domain.NewService("legacy-roll-service", domain.ServiceDraft{
		Name: "Banner", Category: "Large format",
		Components:  []domain.ServiceCostComponentDraft{{ID: "legacy-roll-component", Name: "Roll", Type: domain.CostMaterial, ReferenceID: material.ID, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true}},
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingFixed, FixedPriceRial: 1000},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate an existing database record from before layout configuration.
	service.FinishedSize = nil
	service.Parameters = nil
	if err = store.SaveServiceDefinition(ctx, service); err != nil {
		t.Fatal(err)
	}
	pricing := application.NewPricingService(store, store, store)
	parameters := map[string]string{"finished_width_mm": "3200", "finished_height_mm": "1500"}
	preview, err := pricing.Calculate(ctx, application.PricingRequest{ServiceID: service.ID, Quantity: "2", Parameters: parameters})
	if err != nil {
		t.Fatal(err)
	}
	if preview.BatchQuantity != "2" || preview.EstimatedCostRial != 1000 || len(preview.Layouts) != 1 || preview.Layouts[0].ConsumedQuantity != "3" {
		t.Fatalf("legacy roll pricing did not use custom length: %+v", preview)
	}
	order := domain.NewOrder("legacy-roll-order", "", now)
	if err = store.CreateOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	orders := application.NewOrdersService(store, store, pricing)
	saved, err := orders.AddItem(ctx, order.ID, application.OrderItemInput{ServiceID: service.ID, Quantity: "2", QuantityUnit: "piece", Parameters: parameters})
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Items) != 1 || saved.Items[0].EstimatedCostRial != 1000 {
		t.Fatalf("legacy roll order cost: %+v", saved.Items)
	}
}

func TestLayoutOrderAndProductionUseWholeBatch(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	s, err := Open(filepath.Join(t.TempDir(), "layout.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	material, err := domain.NewMaterial("MAT-layout", domain.MaterialDraft{Name: "Sheet", Kind: domain.MaterialKindSheetStock, PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale, PhysicalStock: 100 * domain.QuantityScale, AverageUnitCostRial: 100,
		Attributes: []domain.MaterialAttributeValue{{Key: "width_mm", ValueType: domain.MaterialAttributeDecimal, DecimalValue: 200 * domain.QuantityScale}, {Key: "height_mm", ValueType: domain.MaterialAttributeDecimal, DecimalValue: 200 * domain.QuantityScale}}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, material); err != nil {
		t.Fatal(err)
	}
	machine, err := domain.NewMachine("MAC-layout", domain.MachineDraft{Name: "Printer", RateBasis: "unit", RateRial: 50, SetupCostRial: 30}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.SaveMachine(ctx, machine); err != nil {
		t.Fatal(err)
	}
	service, err := domain.NewService("SVC-layout", domain.ServiceDraft{Name: "Cards", DefaultUnit: "piece",
		Parameters: []domain.ServiceParameterDraft{
			{ID: "P-w", Key: "w", Label: "Width", Type: domain.ParameterDecimal, DefaultValue: "100", Required: true},
			{ID: "P-h", Key: "h", Label: "Height", Type: domain.ParameterDecimal, DefaultValue: "100", Required: true},
			{ID: "P-q", Key: "q", Label: "Pieces", Type: domain.ParameterInteger, DefaultValue: "1", Required: true}},
		FinishedSize: &domain.ServiceFinishedSizeDefinition{AllowCustom: true, AllowRotation: true, WidthParameterKey: "w", HeightParameterKey: "h", QuantityParameterKey: "q"},
		Components: []domain.ServiceCostComponentDraft{
			{ID: "C-stock", Name: "Stock", Type: domain.CostMaterial, ReferenceID: material.ID, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true},
			{ID: "C-print", Name: "Printing", Type: domain.CostMachine, ReferenceID: machine.ID, RateBasis: "sheet", UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true}},
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingFixed, FixedPriceRial: 1000}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.SaveServiceDefinition(ctx, service); err != nil {
		t.Fatal(err)
	}
	pricing := application.NewPricingService(s, s, s)
	preview, err := pricing.Calculate(ctx, application.PricingRequest{ServiceID: service.ID, Quantity: "5"})
	if err != nil {
		t.Fatal(err)
	}
	if preview.BatchQuantity != "5" || preview.EstimatedCostRial != 1000 || len(preview.Layouts) != 1 || preview.Layouts[0].Sheets != "2" {
		t.Fatalf("batch pricing: %+v", preview)
	}
	draftPrice, err := pricing.CalculateDraft(ctx, application.ServiceInput{Name: "Cards", DefaultUnit: "piece", FinishedSize: service.FinishedSize,
		Parameters:  []application.ParameterInput{{ID: "P-w", Key: "w", Label: "Width", Type: "decimal", DefaultValue: "100", Required: true}, {ID: "P-h", Key: "h", Label: "Height", Type: "decimal", DefaultValue: "100", Required: true}, {ID: "P-q", Key: "q", Label: "Pieces", Type: "integer", DefaultValue: "1", Required: true}},
		Components:  []application.CostComponentInput{{ID: "C-stock", Name: "Stock", Type: "material", ReferenceID: material.ID, UsageQuantity: "1", Multiplier: "1", Enabled: true}, {ID: "C-print", Name: "Printing", Type: "machine", ReferenceID: machine.ID, RateBasis: "sheet", UsageQuantity: "1", Multiplier: "1", Enabled: true}},
		PricingRule: &application.PricingRuleInput{Type: "fixed", FixedPriceRial: 1000}}, application.PricingRequest{Quantity: "5"})
	if err != nil || draftPrice.EstimatedCostRial != preview.EstimatedCostRial || draftPrice.Layouts[0].Sheets != "2" {
		t.Fatalf("draft must match saved pricing: %+v %v", draftPrice, err)
	}
	order := domain.NewOrder("ORD-layout", "", now)
	order.CommercialStatus = domain.CommercialConfirmed
	if err = s.CreateOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	orders := application.NewOrdersService(s, s, pricing)
	saved, err := orders.AddItem(ctx, order.ID, application.OrderItemInput{ServiceID: service.ID, Quantity: "5", QuantityUnit: "piece"})
	if err != nil {
		t.Fatal(err)
	}
	item := saved.Items[0]
	if item.Quantity != "5" || item.EstimatedCostRial != 1000 || item.SellingPriceRial != 1000 {
		t.Fatalf("order double counted batch: %+v", item)
	}
	var snapshot application.PricingView
	if err = json.Unmarshal([]byte(item.PricingSnapshotJSON), &snapshot); err != nil || len(snapshot.Layouts) != 1 {
		t.Fatalf("layout snapshot lost: %+v %v", snapshot, err)
	}
	job := domain.ProductionJob{ID: "JOB-layout", OrderID: order.ID, OrderItemID: item.ID, Status: domain.ProductionPending, Priority: "Normal", CreatedAt: now, UpdatedAt: now}
	if err = s.CreateProductionJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	plans, err := s.ProductionMaterials(ctx, job.ID)
	if err != nil || len(plans) != 1 || plans[0].Required != "2" {
		t.Fatalf("production must require two sheets: %+v %v", plans, err)
	}
	cost, err := productionConversionCost(item.CostBreakdownJSON, 5*domain.QuantityScale, 0)
	if err != nil || cost != 130 {
		t.Fatalf("machine batch cost multiplied again: %d %v", cost, err)
	}
	cost, err = productionConversionCost(item.CostBreakdownJSON, 2*domain.QuantityScale, 0)
	if err != nil || cost != 52 {
		t.Fatalf("partial production cost: %d %v", cost, err)
	}
	updated, err := orders.ReplaceItem(ctx, order.ID, item.ID, application.OrderItemInput{ServiceID: service.ID, Quantity: "9", QuantityUnit: "piece"})
	if err != nil || updated.Items[0].EstimatedCostRial != 1000 {
		t.Fatalf("edited batch cost: %+v %v", updated, err)
	}
	plans, err = s.ProductionMaterials(ctx, job.ID)
	if err != nil || len(plans) != 1 || plans[0].Required != "3" {
		t.Fatalf("edited batch must require three sheets: %+v %v", plans, err)
	}
}
