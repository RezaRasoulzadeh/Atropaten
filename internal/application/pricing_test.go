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
		Components: []domain.ServiceCostComponentDraft{{ID: "C", Name: "Labor", Type: domain.CostLabor, UsageMode: domain.UsageParameter, ParameterKey: "estimated_hours", Multiplier: domain.QuantityScale, RateRial: 1_000_000, RateBasis: domain.RatePerHour, Enabled: true}},
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingMarkup, MarkupPercentage: 20 * domain.QuantityScale},
	}, now)
	if err != nil { t.Fatalf("new service: %v", err) }
	repository := &serviceRepositoryStub{service: service}
	pricing := NewPricingService(repository, materialLookupStub{material: domain.Material{ID: "MAT-paper", Name: "Paper", Active: true}}, machineLookupStub{})
	result, err := pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"estimated_hours": "0.125001", "paper": "MAT-paper"}})
	if err != nil { t.Fatalf("calculate: %v", err) }
	if result.Parameters[0].Quantity != "0.125001" || result.EstimatedCostRial != 125001 || result.SuggestedSellingPriceRial != 150001 { t.Fatalf("unexpected pricing result: %+v", result) }
	_, err = pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"estimated_hours": "0.1", "paper": "MAT-paper"}})
	if err == nil { t.Fatal("below-minimum decimal was accepted") }
	_, err = pricing.Calculate(context.Background(), PricingRequest{ServiceID: service.ID, Parameters: map[string]string{"estimated_hours": "0.2", "size": "Letter", "paper": "MAT-paper"}})
	if err == nil { t.Fatal("invalid choice was accepted") }
}
