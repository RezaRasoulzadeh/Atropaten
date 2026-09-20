package domain

import (
	"math"
	"testing"
	"time"
)

func TestSellingPriceCeilsToHundredTomans(t *testing.T) {
	for _, tt := range []struct {
		rate     int64
		quantity Quantity
		want     int64
		wantErr  bool
	}{
		{0, QuantityScale, 0, false},
		{1, QuantityScale, 1000, false},
		{1000, QuantityScale, 1000, false},
		{1001, QuantityScale, 2000, false},
		{1671942, QuantityScale, 1672000, false},
		{1000, 1000001, 2000, false}, // Do not round a fractional Rial down first.
		{3000, 500000, 2000, false},
		{math.MaxInt64, QuantityScale, 0, true},
		{-1, QuantityScale, 0, true},
	} {
		got, err := MulQuantitySellingPriceRial(tt.quantity, tt.rate)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Fatalf("rate=%d quantity=%d: got %d/%v want %d/error=%v", tt.rate, tt.quantity, got, err, tt.want, tt.wantErr)
		}
	}
}

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
	// 100 + 12.5 + 11.25 = 123.75 Rial. The aggregate cost rounds to
	// 1,000 Rial before the 20% markup, which rounds to a 2,000-Rial price.
	if result.EstimatedCostRial != 1000 || result.SuggestedSellingPriceRial != 2000 {
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
		"fixed":    {&ServicePricingRuleDraft{Type: PricingFixed, FixedPriceRial: 250}, 1000},
		"markup":   {&ServicePricingRuleDraft{Type: PricingMarkup, MarkupPercentage: Quantity(25 * QuantityScale)}, 2000},
		"margin":   {&ServicePricingRuleDraft{Type: PricingFixedMargin, FixedMarginRial: 40}, 2000},
		"per unit": {&ServicePricingRuleDraft{Type: PricingPerUnit, ParameterKey: "quantity", PerUnitRateRial: 20}, 1000},
		"tiers":    {&ServicePricingRuleDraft{Type: PricingTiers, ParameterKey: "quantity", Tiers: []ServicePricingTierDraft{{MinimumQuantity: 0, PriceRial: 100}, {MinimumQuantity: 10 * QuantityScale, PriceRial: 80}}}, 1000},
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

func TestEvaluatePricingUsesSelectedMaterialVariationPrice(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-variation", ServiceDraft{
		Name:             "Variation pricing",
		Parameters:       []ServiceParameterDraft{{ID: "P-finish", Key: "finish", Label: "Finish", Type: ParameterChoice, Required: true, DefaultValue: "matte", MaterialSource: &MaterialParameterSource{ExposedAttributeKey: "finish"}}},
		Components:       []ServiceCostComponentDraft{{ID: "C", Name: "Cost", Type: CostFixed, RateRial: 100, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true}},
		PricingRule:      &ServicePricingRuleDraft{Type: PricingVariation},
		MaterialVariants: []ServiceMaterialVariant{{ID: "VAR-matte", MaterialID: "MAT-matte", Values: map[string]string{"finish": "matte"}, SellingPriceRial: 4321, Position: 0, Active: true}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := EvaluatePricing(PricingInput{Service: service, Parameters: map[string]ResolvedParameter{"finish": {Key: "finish", Type: ParameterChoice, Value: "matte"}}})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.SuggestedSellingPriceRial != 5000 {
		t.Fatalf("variation selling price = %d, want 5000 after rounding", result.SuggestedSellingPriceRial)
	}
}

func TestEvaluatePricingUsesGenericMaterialMachineVariation(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-generic-variation", ServiceDraft{
		Name: "Generic variation pricing",
		Parameters: []ServiceParameterDraft{
			{ID: "P-finish", Key: "finish", Label: "Finish", Type: ParameterChoice, Required: true, MaterialSource: &MaterialParameterSource{ExposedAttributeKey: "finish"}},
			{ID: "P-machine", Key: "machine", Label: "Machine", Type: ParameterMachineReference, Required: true},
		},
		Components:  []ServiceCostComponentDraft{{ID: "C", Name: "Cost", Type: CostFixed, RateRial: 100, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true}},
		PricingRule: &ServicePricingRuleDraft{Type: PricingVariation, Variations: []ServicePricingVariationDraft{{ID: "VAR-gloss-press", Values: map[string]string{"finish": "gloss", "machine": "PRESS-2"}, PriceRial: 4321, Position: 0, Active: true}}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := EvaluatePricing(PricingInput{Service: service, Parameters: map[string]ResolvedParameter{
		"finish":  {Key: "finish", Type: ParameterChoice, Value: "gloss"},
		"machine": {Key: "machine", Type: ParameterMachineReference, Value: "PRESS-2", MachineID: "PRESS-2"},
	}})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.SuggestedSellingPriceRial != 5000 {
		t.Fatalf("variation selling price = %d, want 5000 after rounding", result.SuggestedSellingPriceRial)
	}
}

func TestEvaluatePricingUsesQuantityTierForSelectedVariation(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-tier-variation", ServiceDraft{
		Name:        "Tier variation pricing",
		Parameters:  []ServiceParameterDraft{{ID: "P-qty", Key: "quantity", Label: "Quantity", Type: ParameterInteger}, {ID: "P-machine", Key: "machine", Label: "Machine", Type: ParameterMachineReference}},
		Components:  []ServiceCostComponentDraft{{ID: "C", Name: "Cost", Type: CostFixed, RateRial: 100, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true}},
		PricingRule: &ServicePricingRuleDraft{Type: PricingTiers, ParameterKey: "quantity", Variations: []ServicePricingVariationDraft{{ID: "VAR-press", Values: map[string]string{"machine": "PRESS-1"}, Position: 0, Active: true, Tiers: []ServicePricingTierDraft{{MinimumQuantity: 0, PriceRial: 120}, {MinimumQuantity: 10 * QuantityScale, PriceRial: 90}}}}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := EvaluatePricing(PricingInput{Service: service, Parameters: map[string]ResolvedParameter{
		"quantity": {Key: "quantity", Type: ParameterInteger, Value: "12", Quantity: 12 * QuantityScale},
		"machine":  {Key: "machine", Type: ParameterMachineReference, Value: "PRESS-1", MachineID: "PRESS-1"},
	}})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.SuggestedSellingPriceRial != 1000 {
		t.Fatalf("tier variation selling price = %d, want 1000 after rounding", result.SuggestedSellingPriceRial)
	}
}

func TestEvaluatePricingRoundsAutomaticSellingPriceUp(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-ceil", ServiceDraft{
		Name:        "Ceiling price",
		Components:  []ServiceCostComponentDraft{{ID: "C", Name: "Cost", Type: CostFixed, RateRial: 100, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true}},
		PricingRule: &ServicePricingRuleDraft{Type: PricingMarkup, MarkupPercentage: Quantity(12_500_000)},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := EvaluatePricing(PricingInput{Service: service})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.SuggestedSellingPriceRial != 2000 {
		t.Fatalf("selling price = %d, want rounded 2000", result.SuggestedSellingPriceRial)
	}
}

func TestEvaluatePricingReportsBelowCostOverride(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-override", ServiceDraft{Name: "Override", Components: []ServiceCostComponentDraft{{ID: "C", Name: "Cost", Type: CostFixed, RateRial: 10000, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true}}, PricingRule: &ServicePricingRuleDraft{Type: PricingFixed, FixedPriceRial: 10000}}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	override := int64(5000)
	result, err := EvaluatePricing(PricingInput{Service: service, SellingPriceOverrideRial: &override})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if !result.BelowCost || result.ProfitRial != -5000 || result.MarginPercentage != Quantity(-100*QuantityScale) {
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
	if result.EstimatedCostRial != 1000 || result.Components[0].RateRial != 175 {
		t.Fatalf("material price = %+v, want rounded cost 1000 at rate 175", result)
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
	if result.EstimatedCostRial != 1000 || result.Components[0].RateRial != 225 {
		t.Fatalf("selected material price = %+v, want rounded cost 1000 at rate 225", result)
	}
}

func TestEvaluatePricingCountsOneMaterialForGroupedCombination(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-grouped-material-cost", ServiceDraft{
		Name: "Grouped material cost",
		Parameters: []ServiceParameterDraft{
			{ID: "P-size", Key: "size", Label: "Paper size", Type: ParameterChoice, MaterialSource: &MaterialParameterSource{ExposedAttributeKey: "width_mm"}},
			{ID: "P-type", Key: "type", Label: "Paper type", Type: ParameterChoice, MaterialSource: &MaterialParameterSource{ExposedAttributeKey: "finish"}},
		},
		Components: []ServiceCostComponentDraft{
			{ID: "C-size", Name: "Paper size", Type: CostMaterial, UsageMode: UsageParameter, ParameterKey: "size", UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true},
			// A previous service can retain a fixed material component from the old
			// parameter flow. Grouped material options must still resolve to one
			// inventory material and one material cost.
			{ID: "C-type", Name: "Paper type", Type: CostMaterial, ReferenceID: "MAT-legacy-paper", UsageMode: UsageFixed, UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true},
		},
		MaterialVariants: []ServiceMaterialVariant{{ID: "VAR-a4-coated", MaterialID: "MAT-a4-coated", Values: map[string]string{"size": "210", "type": "coated"}, Position: 0, Active: true}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	result, err := EvaluatePricing(PricingInput{
		Service: service,
		Parameters: map[string]ResolvedParameter{
			"size": {Key: "size", Type: ParameterChoice, Value: "210", MaterialID: "MAT-a4-coated"},
			"type": {Key: "type", Type: ParameterChoice, Value: "coated", MaterialID: "MAT-a4-coated"},
		},
		Materials: map[string]Material{"MAT-a4-coated": {ID: "MAT-a4-coated", HighestPurchaseUnitCostRial: 175}},
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.EstimatedCostRial != 1000 || len(result.Components) != 1 {
		t.Fatalf("grouped material cost = %+v, want one component and rounded cost", result)
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
	if result.EstimatedCostRial != 1000 || result.Components[0].RateRial != 100 {
		t.Fatalf("black and white rate = %+v, want 100", result)
	}
	input.Parameters["color"] = ResolvedParameter{Key: "color", Type: ParameterChoice, Value: "Full color"}
	result, err = EvaluatePricing(input)
	if err != nil {
		t.Fatalf("evaluate full color: %v", err)
	}
	if result.EstimatedCostRial != 1000 || result.Components[0].RateRial != 250 {
		t.Fatalf("full color rate = %+v, want 250", result)
	}
}

func TestEvaluatePricingUsesMachineMeterAndSquareMeterRatesForRollLayout(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-large-format-machine", ServiceDraft{
		Name: "Banner print",
		Parameters: []ServiceParameterDraft{
			{ID: "P-width", Key: "finished_width_mm", Label: "Finished width", Type: ParameterDecimal, Required: true},
			{ID: "P-height", Key: "finished_height_mm", Label: "Finished height", Type: ParameterDecimal, Required: true},
			{ID: "P-quantity", Key: "layout_quantity", Label: "Finished pieces", Type: ParameterInteger, Required: true},
		},
		FinishedSize: &ServiceFinishedSizeDefinition{AllowCustom: true, AllowRotation: true, WidthParameterKey: "finished_width_mm", HeightParameterKey: "finished_height_mm", QuantityParameterKey: "layout_quantity"},
		Components: []ServiceCostComponentDraft{
			{ID: "C-material", Name: "Banner roll", Type: CostMaterial, ReferenceID: "MAT-roll", UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true},
			{ID: "C-machine", Name: "Large format printer", Type: CostMachine, ReferenceID: "MAC-large", UsageQuantity: QuantityScale, Multiplier: QuantityScale, Enabled: true},
		},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	material := Material{ID: "MAT-roll", Name: "3200 mm banner", Kind: MaterialKindRollMedia, ConsumptionUnit: "meter", Attributes: []MaterialAttributeValue{{Key: "width_mm", ValueType: MaterialAttributeDecimal, DecimalValue: 3200 * QuantityScale}}, Active: true}
	parameters := map[string]ResolvedParameter{
		"finished_width_mm":  {Key: "finished_width_mm", Type: ParameterDecimal, Value: "1000", Quantity: 1000 * QuantityScale},
		"finished_height_mm": {Key: "finished_height_mm", Type: ParameterDecimal, Value: "3000", Quantity: 3000 * QuantityScale},
		"layout_quantity":    {Key: "layout_quantity", Type: ParameterInteger, Value: "1", Quantity: QuantityScale},
	}
	for _, test := range []struct {
		basis string
		rate  int64
		want  int64
	}{
		{RatePerMeter, 20, 20},
		{RatePerSquareMeter, 10, 32},
	} {
		machine := Machine{ID: "MAC-large", Name: "Large format printer", RateBasis: test.basis, RateRial: test.rate, Active: true, Rates: []MachineRate{{ID: "standard", Name: "Standard", RateBasis: test.basis, RateRial: test.rate, Active: true}}}
		result, evalErr := EvaluatePricing(PricingInput{BatchQuantity: QuantityScale, Service: service, Parameters: parameters, Materials: map[string]Material{material.ID: material}, Machines: map[string]Machine{machine.ID: machine}})
		if evalErr != nil {
			t.Fatalf("%s pricing: %v", test.basis, evalErr)
		}
		if result.Components[1].AmountRial != test.want {
			t.Fatalf("%s machine amount = %d, want %d", test.basis, result.Components[1].AmountRial, test.want)
		}
	}
}

func TestEvaluatePricingRequiresMatchingPredefinedMachineSelector(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-machine-color-catalog", ServiceDraft{
		Name:       "Catalog color print",
		Parameters: []ServiceParameterDraft{{ID: "P-color", Key: "color", Label: "Color", Type: ParameterChoice, PredefinedKey: PredefinedParameterColor, Required: true}},
		Components: []ServiceCostComponentDraft{{ID: "C-machine", Name: "Printer", Type: CostMachine, ReferenceID: "MAC-printer", UsageMode: UsageFixed, UsageQuantity: QuantityScale, Multiplier: QuantityScale, RateParameterKey: "color", Enabled: true}},
	}, now)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	machine := Machine{ID: "MAC-printer", Name: "Printer", RateBasis: RatePerUnit, RateRial: 100, Active: true, Rates: []MachineRate{{ID: "color", Name: "Full color", SelectorValue: "full-color", SelectorPredefinedKey: PredefinedParameterColor, RateBasis: RatePerUnit, RateRial: 250, Active: true}}}
	result, err := EvaluatePricing(PricingInput{Service: service, Parameters: map[string]ResolvedParameter{"color": {Key: "color", Type: ParameterChoice, Value: "full-color", PredefinedKey: PredefinedParameterColor}}, Machines: map[string]Machine{"MAC-printer": machine}})
	if err != nil || result.Components[0].RateRial != 250 {
		t.Fatalf("catalog color rate = %+v, err=%v", result, err)
	}
	_, err = EvaluatePricing(PricingInput{Service: service, Parameters: map[string]ResolvedParameter{"color": {Key: "color", Type: ParameterChoice, Value: "full-color"}}, Machines: map[string]Machine{"MAC-printer": machine}})
	if err == nil {
		t.Fatal("expected a predefined machine selector not to match an ordinary choice")
	}
}
