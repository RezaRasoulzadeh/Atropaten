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

func TestEvaluatePricingUsesHighestPostedPurchaseCostForMaterials(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-material-price", ServiceDraft{
		Name: "Material pricing",
		Components: []ServiceCostComponentDraft{{
			ID: "C-material", Name: "Paper", Type: CostMaterial, ReferenceID: "MAT-paper",
			UsageMode: UsageFixed, UsageQuantity: 2 * QuantityScale, Multiplier: QuantityScale, Enabled: true,
		}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := EvaluatePricing(PricingInput{
		Service:   service,
		Materials: map[string]Material{"MAT-paper": {ID: "MAT-paper", AverageUnitCostRial: 100, HighestPurchaseUnitCostRial: 175}},
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.EstimatedCostRial != 350 || result.Components[0].RateRial != 175 {
		t.Fatalf("material price = %+v, want cost 350 at rate 175", result)
	}
}

func TestEvaluatePricingUsesSelectedMaterialParameter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-material-parameter", ServiceDraft{
		Name:       "Selectable paper",
		Parameters: []ServiceParameterDraft{{ID: "P-paper", Key: "paper", Label: "Paper", Type: ParameterMaterialReference, Required: true}},
		Components: []ServiceCostComponentDraft{{
			ID: "C-paper", Name: "Paper", Type: CostMaterial, UsageMode: UsageParameter, ParameterKey: "paper",
			UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true,
		}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := EvaluatePricing(PricingInput{
		Service:    service,
		Parameters: map[string]ResolvedParameter{"paper": {Key: "paper", Type: ParameterMaterialReference, Value: "MAT-a5", MaterialID: "MAT-a5"}},
		Materials:  map[string]Material{"MAT-a5": {ID: "MAT-a5", HighestPurchaseUnitCostRial: 225}},
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.EstimatedCostRial != 225 || result.Components[0].RateRial != 225 {
		t.Fatalf("selected material price = %+v, want cost 225 at rate 225", result)
	}
}

func TestEvaluatePricingUsesMachineRateSelectedByChoice(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-machine-rates", ServiceDraft{
		Name:       "Color print",
		Parameters: []ServiceParameterDraft{{ID: "P-color", Key: "color", Label: "Color", Type: ParameterChoice, Required: true, Options: []string{"Black & white", "Full color"}, DefaultValue: "Black & white"}},
		Components: []ServiceCostComponentDraft{{ID: "C-machine", Name: "Printer", Type: CostMachine, ReferenceID: "MAC-printer", UsageMode: UsageFixed, UsageQuantity: QuantityScale, Multiplier: QuantityScale, RateParameterKey: "color", Enabled: true}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	machine := Machine{ID: "MAC-printer", Name: "Printer", RateBasis: RatePerUnit, RateRial: 100, Active: true, Rates: []MachineRate{
		{ID: "bw", Name: "Black & white", SelectorValue: "Black & white", RateBasis: RatePerUnit, RateRial: 100, Active: true},
		{ID: "color", Name: "Full color", SelectorValue: "Full color", RateBasis: RatePerUnit, RateRial: 250, Active: true},
	}}
	input := PricingInput{Service: service, Parameters: map[string]ResolvedParameter{"color": {Key: "color", Type: ParameterChoice, Value: "Black & white"}}, Machines: map[string]Machine{"MAC-printer": machine}}
	result, err := EvaluatePricing(input)
	if err != nil {
		t.Fatalf("evaluate black and white: %v", err)
	}
	if result.EstimatedCostRial != 100 || result.Components[0].RateRial != 100 {
		t.Fatalf("black and white rate = %+v, want 100", result)
	}
	input.Parameters["color"] = ResolvedParameter{Key: "color", Type: ParameterChoice, Value: "Full color"}
	result, err = EvaluatePricing(input)
	if err != nil {
		t.Fatalf("evaluate full color: %v", err)
	}
	if result.EstimatedCostRial != 250 || result.Components[0].RateRial != 250 {
		t.Fatalf("full color rate = %+v, want 250", result)
	}
}
