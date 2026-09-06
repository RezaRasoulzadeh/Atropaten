package domain

import (
	"testing"
	"time"
)

func TestEvaluatePricingUsesDeterministicFixedScaleArithmetic(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-price", ServiceDraft{
		Name:       "Generic service",
		Parameters: []ServiceParameterDraft{{ID: "P-qty", Key: "quantity", Label: "Quantity", Type: ParameterInteger, Required: true}},
		Components: []ServiceCostComponentDraft{
			{ID: "C-fixed", Name: "Fixed", Type: CostFixed, RateRial: 100, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true},
			{ID: "C-waste", Name: "Waste", Type: CostWaste, Percentage: Quantity(12_500_000), UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true},
			{ID: "C-overhead", Name: "Overhead", Type: CostOverhead, Percentage: Quantity(10 * QuantityScale), UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true},
		},
		PricingRule: &ServicePricingRuleDraft{Type: PricingMarkup, MarkupPercentage: Quantity(20 * QuantityScale)},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := EvaluatePricing(PricingInput{Service: service, Parameters: map[string]ResolvedParameter{"quantity": {Key: "quantity", Type: ParameterInteger, Value: "3", Quantity: 3 * QuantityScale}}})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	// 100 + 12.5 + 11.25 = 123.75, then 20% markup = 148.5 Rial.
	if result.EstimatedCostRial != 124 || result.SuggestedSellingPriceRial != 149 {
		t.Fatalf("unexpected rounded totals: %+v", result)
	}
	if result.Components[1].Explanation != "12.5 of accumulated cost before this component" {
		t.Fatalf("percentage explanation lost: %+v", result.Components[1])
	}
}

func TestEvaluatePricingSupportsGenericRules(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	base := func(rule *ServicePricingRuleDraft) Service {
		service, err := NewService("SVC-rule", ServiceDraft{Name: "Rule", Parameters: []ServiceParameterDraft{{ID: "P-qty", Key: "quantity", Label: "Quantity", Type: ParameterInteger}}, Components: []ServiceCostComponentDraft{{ID: "C", Name: "Cost", Type: CostFixed, RateRial: 100, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true}}, PricingRule: rule}, now)
		if err != nil {
			t.Fatalf("new service: %v", err)
		}
		return service
	}
	parameter := map[string]ResolvedParameter{"quantity": {Key: "quantity", Type: ParameterInteger, Value: "12", Quantity: 12 * QuantityScale}}
	for name, test := range map[string]struct {
		rule     *ServicePricingRuleDraft
		expected int64
	}{
		"fixed":    {&ServicePricingRuleDraft{Type: PricingFixed, FixedPriceRial: 250}, 250},
		"markup":   {&ServicePricingRuleDraft{Type: PricingMarkup, MarkupPercentage: Quantity(25 * QuantityScale)}, 125},
		"margin":   {&ServicePricingRuleDraft{Type: PricingFixedMargin, FixedMarginRial: 40}, 140},
		"per unit": {&ServicePricingRuleDraft{Type: PricingPerUnit, ParameterKey: "quantity", PerUnitRateRial: 20}, 240},
		"tiers":    {&ServicePricingRuleDraft{Type: PricingTiers, ParameterKey: "quantity", Tiers: []ServicePricingTierDraft{{MinimumQuantity: 0, PriceRial: 100}, {MinimumQuantity: 10 * QuantityScale, PriceRial: 80}}}, 80},
		"manual":   {&ServicePricingRuleDraft{Type: PricingManual}, 0},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := EvaluatePricing(PricingInput{Service: base(test.rule), Parameters: parameter})
			if err != nil {
				t.Fatalf("evaluate: %v", err)
			}
			if result.SuggestedSellingPriceRial != test.expected {
				t.Fatalf("suggested price = %d, want %d", result.SuggestedSellingPriceRial, test.expected)
			}
		})
	}
}

func TestEvaluatePricingReportsBelowCostOverride(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-override", ServiceDraft{Name: "Override", Components: []ServiceCostComponentDraft{{ID: "C", Name: "Cost", Type: CostFixed, RateRial: 100, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true}}, PricingRule: &ServicePricingRuleDraft{Type: PricingFixed, FixedPriceRial: 100}}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	override := int64(50)
	result, err := EvaluatePricing(PricingInput{Service: service, SellingPriceOverrideRial: &override})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if !result.BelowCost || result.ProfitRial != -50 || result.MarginPercentage != Quantity(-100*QuantityScale) {
		t.Fatalf("below-cost result incorrect: %+v", result)
	}
}
