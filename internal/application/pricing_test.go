package application

import (
	"context"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestPricingServiceResolvesDynamicParametersAndRejectsInvalidValues(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	minimum := domain.Quantity(125001)
	service, err := domain.NewService("SVC-config", domain.ServiceDraft{
		Name: "Configurator",
		Parameters: []domain.ServiceParameterDraft{
			{ID: "P-hours", Key: "estimated_hours", Label: "Estimated hours", Type: domain.ParameterDecimal, Required: true, MinValue: &minimum},
			{ID: "P-size", Key: "size", Label: "Size", Type: domain.ParameterChoice, Options: []string{"A4", "A3"}, DefaultValue: "A4"},
			{ID: "P-paper", Key: "paper", Label: "Paper", Type: domain.ParameterMaterialReference, Required: true},
		},
		Components:  []domain.ServiceCostComponentDraft{{ID: "C", Name: "Labor", Type: domain.CostLabor, UsageMode: domain.UsageParameter, ParameterKey: "estimated_hours", Multiplier: domain.QuantityScale, RateRial: 1_000_000, RateBasis: domain.RatePerHour, Enabled: true}},
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingMarkup, MarkupPercentage: 20 * domain.QuantityScale},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	repository := &serviceRepositoryStub{service: service}
	pricing := NewPricingService(repository, materialLookupStub{material: domain.Material{ID: "MAT-paper", Name: "Paper", Active: true}}, machineLookupStub{})
	result, err := pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"estimated_hours": "0.125001", "paper": "MAT-paper"}})
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if result.Parameters[0].Quantity != "0.125001" || result.EstimatedCostRial != 125001 || result.SuggestedSellingPriceRial != 151000 {
		t.Fatalf("unexpected pricing result: %+v", result)
	}
	_, err = pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"estimated_hours": "0.1", "paper": "MAT-paper"}})
	if err == nil {
		t.Fatal("below-minimum decimal was accepted")
	}
	_, err = pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"estimated_hours": "0.2", "size": "Letter", "paper": "MAT-paper"}})
	if err == nil {
		t.Fatal("invalid choice was accepted")
	}
}

func TestPricingServiceUsesPersistedMonetaryRoundingStep(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-rounding-policy", domain.ServiceDraft{
		Name:        "Policy service",
		Components:  []domain.ServiceCostComponentDraft{{ID: "C", Name: "Cost", Type: domain.CostFixed, RateRial: 12001, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true}},
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingFixedMargin, FixedMarginRial: 1},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	repository := &roundingServiceRepository{serviceRepositoryStub: serviceRepositoryStub{service: service}, step: 100}
	result, err := NewPricingService(repository, materialLookupStub{}, machineLookupStub{}).Calculate(context.Background(), PricingRequest{ServiceID: service.ID})
	if err != nil {
		t.Fatal(err)
	}
	if result.EstimatedCostRial != 12100 || result.SuggestedSellingPriceRial != 12100 || result.RoundingStepRial != 100 {
		t.Fatalf("rounded pricing=%+v", result)
	}
	repository.step = 1
	result, err = NewPricingService(repository, materialLookupStub{}, machineLookupStub{}).Calculate(context.Background(), PricingRequest{ServiceID: service.ID})
	if err != nil || result.EstimatedCostRial != 12001 || result.SuggestedSellingPriceRial != 12002 {
		t.Fatalf("step change pricing=%+v err=%v", result, err)
	}
	override := int64(12001)
	repository.step = 100
	result, err = NewPricingService(repository, materialLookupStub{}, machineLookupStub{}).Calculate(context.Background(), PricingRequest{ServiceID: service.ID, SellingPriceOverrideRial: &override})
	if err != nil || result.EffectiveSellingPriceRial != override {
		t.Fatalf("explicit override was rounded: %+v err=%v", result, err)
	}
}

type roundingServiceRepository struct {
	serviceRepositoryStub
	step int64
}

func (r *roundingServiceRepository) GetShopSettings(context.Context) (domain.ShopSettings, error) {
	return domain.ShopSettings{MonetaryRoundingStepRial: r.step}, nil
}

func TestPricingServiceCalculatesMaterialSelectedByParameter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-paper-selection", domain.ServiceDraft{
		Name:       "Full color print",
		Parameters: []domain.ServiceParameterDraft{{ID: "P-paper", Key: "paper", Label: "Paper size", Type: domain.ParameterMaterialReference, Required: true}},
		Components: []domain.ServiceCostComponentDraft{
			{ID: "C-paper", Name: "Paper", Type: domain.CostMaterial, UsageMode: domain.UsageParameter, ParameterKey: "paper", UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true},
			{ID: "C-machine", Name: "Printer", Type: domain.CostMachine, ReferenceID: "MAC-printer", UsageMode: domain.UsageFixed, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true},
		},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	pricing := NewPricingService(
		&serviceRepositoryStub{service: service},
		materialLookupStub{material: domain.Material{ID: "MAT-a4", Name: "A4 paper", Active: true, HighestPurchaseUnitCostRial: 225}},
		machineLookupStub{machine: domain.Machine{ID: "MAC-printer", Name: "Printer", Active: true, RateRial: 500}},
	)
	result, err := pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"paper": "MAT-a4"}})
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if result.EstimatedCostRial != 725 || len(result.Components) != 2 {
		t.Fatalf("material and machine costs = %+v, want 725 across 2 components", result)
	}
}

func TestPricingServiceMapsChoiceMaterialOptionToMaterialCost(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-choice-paper", domain.ServiceDraft{
		Name:       "Choice paper",
		Parameters: []domain.ServiceParameterDraft{{ID: "P-paper", Key: "paper", Label: "Paper size", Type: domain.ParameterChoice, Required: true, Options: []string{"A4", "A5"}, DefaultValue: "A4"}},
		Components: []domain.ServiceCostComponentDraft{{ID: "C-paper", Name: "Paper", Type: domain.CostMaterial, UsageMode: domain.UsageParameter, ParameterKey: "paper", UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	pricing := NewPricingService(
		&serviceRepositoryStub{service: service},
		materialLookupStub{items: []domain.Material{
			{ID: "MAT-a4", Name: "A4 80gsm paper", SKU: "PAPER-A4", Active: true, HighestPurchaseUnitCostRial: 225},
			{ID: "MAT-a5", Name: "A5 120gsm paper", SKU: "PAPER-A5", Active: true, HighestPurchaseUnitCostRial: 450},
		}},
		machineLookupStub{},
	)
	result, err := pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"paper": "A4"}})
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if result.EstimatedCostRial != 225 {
		t.Fatalf("choice material cost = %d, want 225", result.EstimatedCostRial)
	}
	result, err = pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"paper": "A5"}})
	if err != nil {
		t.Fatalf("calculate alternate choice: %v", err)
	}
	if result.EstimatedCostRial != 450 {
		t.Fatalf("alternate choice material cost = %d, want 450", result.EstimatedCostRial)
	}
}

func TestPricingServiceCalculatesMachineSelectedByParameter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-machine-selection", domain.ServiceDraft{
		Name:       "Selectable machine service",
		Parameters: []domain.ServiceParameterDraft{{ID: "P-machine", Key: "print_method", Label: "Print method", Type: domain.ParameterMachineReference, Required: true}},
		Components: []domain.ServiceCostComponentDraft{{ID: "C-machine", Name: "Machine", Type: domain.CostMachine, UsageMode: domain.UsageParameter, ParameterKey: "print_method", UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	pricing := NewPricingService(
		&serviceRepositoryStub{service: service},
		materialLookupStub{},
		machineLookupStub{machine: domain.Machine{ID: "MAC-dtf", Name: "DTF printer", Active: true, RateRial: 750}},
	)
	result, err := pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"print_method": "MAC-dtf"}})
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if result.EstimatedCostRial != 750 {
		t.Fatalf("machine cost = %d, want 750", result.EstimatedCostRial)
	}
}

type serviceMapRepository struct {
	services map[string]domain.Service
}

func (r *serviceMapRepository) ListServices(_ context.Context, _ bool) ([]domain.Service, error) {
	services := make([]domain.Service, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}
	return services, nil
}

func (r *serviceMapRepository) GetService(_ context.Context, id string) (domain.Service, error) {
	service, ok := r.services[id]
	if !ok {
		return domain.Service{}, domain.ErrServiceNotFound
	}
	return service, nil
}

func (r *serviceMapRepository) SaveServiceDefinition(_ context.Context, service domain.Service) error {
	r.services[service.ID] = service
	return nil
}

func TestPricingServiceIncludesNestedServiceCostAndRejectsCycles(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	finishing, err := domain.NewService("SVC-finishing", domain.ServiceDraft{
		Name:       "Lamination",
		Components: []domain.ServiceCostComponentDraft{{ID: "C-lamination", Name: "Lamination labor", Type: domain.CostFixed, UsageMode: domain.UsageFixed, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, RateRial: 300, Enabled: true}},
	}, now)
	if err != nil {
		t.Fatalf("new finishing service: %v", err)
	}
	printing, err := domain.NewService("SVC-printing", domain.ServiceDraft{
		Name:       "Full color printing",
		Components: []domain.ServiceCostComponentDraft{{ID: "C-finishing", Name: "Lamination service", Type: domain.CostService, ReferenceID: finishing.ID, UsageMode: domain.UsageFixed, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true}},
	}, now)
	if err != nil {
		t.Fatalf("new printing service: %v", err)
	}
	repository := &serviceMapRepository{services: map[string]domain.Service{finishing.ID: finishing, printing.ID: printing}}
	pricing := NewPricingService(repository, materialLookupStub{}, machineLookupStub{})
	result, err := pricing.Calculate(context.Background(), PricingRequest{ServiceID: printing.ID})
	if err != nil {
		t.Fatalf("calculate nested service cost: %v", err)
	}
	if result.EstimatedCostRial != 300 || len(result.Components) != 1 || result.Components[0].RateRial != 300 {
		t.Fatalf("nested service cost = %+v, want 300 Rial", result)
	}

	finishing.Components = []domain.ServiceCostComponent{{ID: "C-printing", Name: "Printing service", Type: domain.CostService, ReferenceID: printing.ID, UsageMode: domain.UsageFixed, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true, CreatedAt: now, UpdatedAt: now}}
	repository.services[finishing.ID] = finishing
	if _, err := pricing.Calculate(context.Background(), PricingRequest{ServiceID: printing.ID}); err == nil {
		t.Fatal("circular service cost was accepted")
	}
}
