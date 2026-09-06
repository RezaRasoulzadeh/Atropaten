package domain

import (
	"fmt"
	"math/big"
	"sort"
)

// ResolvedParameter is the canonical, validated value used by the pricing engine.
// Value remains textual so booleans, choices, and material IDs do not lose meaning.
type ResolvedParameter struct {
	Key        string
	Type       ParameterType
	Value      string
	Quantity   Quantity
	MaterialID string
}

type PricingInput struct {
	Service                  Service
	Parameters               map[string]ResolvedParameter
	Materials                map[string]Material
	Machines                 map[string]Machine
	ManualCosts              map[string]int64
	SellingPriceOverrideRial *int64
}

type PricingComponentResult struct {
	ID            string
	Name          string
	Type          CostComponentType
	Enabled       bool
	UsageQuantity Quantity
	RateRial      int64
	Percentage    Quantity
	AmountRial    int64
	Explanation   string
}

type PricingResult struct {
	ServiceID                 string
	ServiceName               string
	Parameters                []ResolvedParameter
	Components                []PricingComponentResult
	EstimatedCostRial         int64
	SuggestedSellingPriceRial int64
	EffectiveSellingPriceRial int64
	ProfitRial                int64
	MarginPercentage          Quantity
	Warnings                  []string
	BelowCost                 bool
}

// EvaluatePricing evaluates the persisted definition in display order. All
// fractional work uses integers and big.Int; monetary results use half-up
// rounding at each explicitly defined operation.
func EvaluatePricing(input PricingInput) (PricingResult, error) {
	if err := input.Service.Validate(); err != nil {
		return PricingResult{}, err
	}
	parameters := make([]ResolvedParameter, 0, len(input.Parameters))
	for _, parameter := range input.Parameters {
		parameters = append(parameters, parameter)
	}
	sort.SliceStable(parameters, func(left, right int) bool { return parameters[left].Key < parameters[right].Key })
	result := PricingResult{ServiceID: input.Service.ID, ServiceName: input.Service.Name, Parameters: parameters, Warnings: []string{}}
	manualComponents := make(map[string]bool)
	for _, component := range input.Service.Components {
		if component.Type == CostManual {
			manualComponents[component.ID] = true
		}
	}
	for id, amount := range input.ManualCosts {
		if !manualComponents[id] {
			return PricingResult{}, fmt.Errorf("manual cost %q does not reference a manual component", id)
		}
		if amount < 0 {
			return PricingResult{}, fmt.Errorf("manual cost %q cannot be negative", id)
		}
	}
	var total int64
	for _, component := range input.Service.Components {
		item := PricingComponentResult{ID: component.ID, Name: component.Name, Type: component.Type, Enabled: component.Enabled, UsageQuantity: component.UsageQuantity, RateRial: component.RateRial, Percentage: component.Percentage}
		if !component.Enabled {
			item.Explanation = "Disabled component"
			result.Components = append(result.Components, item)
			continue
		}
		var amount int64
		var err error
		switch component.Type {
		case CostOverhead, CostWaste:
			amount, err = percentageAmount(total, component.Percentage)
			item.Explanation = fmt.Sprintf("%s of accumulated cost before this component", component.Percentage.String())
		default:
			usage, usageErr := componentUsage(component, input.Parameters)
			if usageErr != nil {
				return PricingResult{}, usageErr
			}
			item.UsageQuantity = usage
			switch component.Type {
			case CostMaterial:
				material, exists := input.Materials[component.ReferenceID]
				if !exists {
					return PricingResult{}, fmt.Errorf("component %q: material %q not found", component.Name, component.ReferenceID)
				}
				amount, err = scaledMoney(usage, material.AverageUnitCostRial)
				item.Explanation = fmt.Sprintf("%s × %d Rial average unit cost", usage.String(), material.AverageUnitCostRial)
			case CostMachine:
				machine, exists := input.Machines[component.ReferenceID]
				if !exists {
					return PricingResult{}, fmt.Errorf("component %q: machine %q not found", component.Name, component.ReferenceID)
				}
				amount, err = scaledMoney(usage, machine.RateRial)
				item.RateRial = machine.RateRial
				item.Explanation = fmt.Sprintf("%s × %d Rial %s rate", usage.String(), machine.RateRial, machine.RateBasis)
			case CostLabor, CostOutsourced, CostFixed:
				amount, err = scaledMoney(usage, component.RateRial)
				item.Explanation = fmt.Sprintf("%s × %d Rial rate", usage.String(), component.RateRial)
			case CostManual:
				amount = component.RateRial
				if manual, exists := input.ManualCosts[component.ID]; exists {
					amount = manual
				}
				if amount == 0 {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Manual component %q has no entered amount", component.Name))
				}
				item.Explanation = "Manual amount"
			default:
				return PricingResult{}, fmt.Errorf("component %q: unsupported type %q", component.Name, component.Type)
			}
		}
		if err != nil {
			return PricingResult{}, fmt.Errorf("component %q: %w", component.Name, err)
		}
		item.AmountRial = amount
		total, err = addMoney(total, amount)
		if err != nil {
			return PricingResult{}, err
		}
		result.Components = append(result.Components, item)
	}
	result.EstimatedCostRial = total
	suggested, warnings, err := suggestedPrice(input.Service.PricingRule, input.Parameters, total)
	if err != nil {
		return PricingResult{}, err
	}
	result.SuggestedSellingPriceRial = suggested
	result.Warnings = append(result.Warnings, warnings...)
	result.EffectiveSellingPriceRial = suggested
	if input.SellingPriceOverrideRial != nil {
		if *input.SellingPriceOverrideRial < 0 {
			return PricingResult{}, fmt.Errorf("selling price override cannot be negative")
		}
		result.EffectiveSellingPriceRial = *input.SellingPriceOverrideRial
	}
	result.ProfitRial, err = subtractMoney(result.EffectiveSellingPriceRial, total)
	if err != nil {
		return PricingResult{}, err
	}
	if result.EffectiveSellingPriceRial > 0 {
		result.MarginPercentage, err = ratioQuantity(result.ProfitRial, result.EffectiveSellingPriceRial, 100)
		if err != nil {
			return PricingResult{}, err
		}
	} else {
		result.Warnings = append(result.Warnings, "Selling price is zero; margin cannot be calculated")
	}
	result.BelowCost = result.EffectiveSellingPriceRial < total
	if result.BelowCost {
		result.Warnings = append(result.Warnings, "Selling price is below estimated cost")
	}
	return result, nil
}

func componentUsage(component ServiceCostComponent, parameters map[string]ResolvedParameter) (Quantity, error) {
	base := component.UsageQuantity
	if base == 0 {
		base = Quantity(QuantityScale)
	}
	if component.UsageMode == UsageParameter {
		parameter, exists := parameters[component.ParameterKey]
		if !exists {
			return 0, fmt.Errorf("component %q: parameter %q is not resolved", component.Name, component.ParameterKey)
		}
		base = parameter.Quantity
	}
	return quantityProduct(base, component.Multiplier)
}

func quantityProduct(left, right Quantity) (Quantity, error) {
	value, err := roundBig(new(big.Int).Mul(big.NewInt(int64(left)), big.NewInt(int64(right))), big.NewInt(QuantityScale))
	if err != nil {
		return 0, err
	}
	return Quantity(value), nil
}

func scaledMoney(quantity Quantity, rate int64) (int64, error) {
	return roundBig(new(big.Int).Mul(big.NewInt(int64(quantity)), big.NewInt(rate)), big.NewInt(QuantityScale))
}

func percentageAmount(base int64, percentage Quantity) (int64, error) {
	denominator := big.NewInt(100 * QuantityScale)
	return roundBig(new(big.Int).Mul(big.NewInt(base), big.NewInt(int64(percentage))), denominator)
}

func suggestedPrice(rule *ServicePricingRule, parameters map[string]ResolvedParameter, cost int64) (int64, []string, error) {
	if rule == nil || rule.Type == PricingManual {
		return 0, []string{"No automatic selling-price rule is configured"}, nil
	}
	switch rule.Type {
	case PricingFixed:
		return rule.FixedPriceRial, nil, nil
	case PricingMarkup:
		markup, err := percentageAmount(cost, rule.MarkupPercentage)
		if err != nil {
			return 0, nil, err
		}
		total, err := addMoney(cost, markup)
		return total, nil, err
	case PricingFixedMargin:
		total, err := addMoney(cost, rule.FixedMarginRial)
		return total, nil, err
	case PricingPerUnit:
		parameter, exists := parameters[rule.ParameterKey]
		if !exists {
			return 0, nil, fmt.Errorf("pricing rule parameter %q is not resolved", rule.ParameterKey)
		}
		price, err := scaledMoney(parameter.Quantity, rule.PerUnitRateRial)
		return price, nil, err
	case PricingTiers:
		parameter, exists := parameters[rule.ParameterKey]
		if !exists {
			return 0, nil, fmt.Errorf("pricing tier parameter %q is not resolved", rule.ParameterKey)
		}
		var selected *ServicePricingTier
		for index := range rule.Tiers {
			if parameter.Quantity >= rule.Tiers[index].MinimumQuantity {
				selected = &rule.Tiers[index]
			}
		}
		if selected == nil {
			return 0, []string{"No quantity tier matches the resolved quantity"}, nil
		}
		return selected.PriceRial, nil, nil
	default:
		return 0, nil, fmt.Errorf("unsupported pricing rule %q", rule.Type)
	}
}

func ratioQuantity(numerator, denominator int64, multiplier int64) (Quantity, error) {
	n := new(big.Int).Mul(big.NewInt(numerator), big.NewInt(multiplier*QuantityScale))
	value, err := roundBig(n, big.NewInt(denominator))
	return Quantity(value), err
}

func roundBig(numerator, denominator *big.Int) (int64, error) {
	if denominator.Sign() <= 0 {
		return 0, fmt.Errorf("pricing denominator must be positive")
	}
	negative := numerator.Sign() < 0
	absolute := new(big.Int).Abs(numerator)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(absolute, denominator, remainder)
	if new(big.Int).Mul(remainder, big.NewInt(2)).Cmp(denominator) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if negative {
		quotient.Neg(quotient)
	}
	if !quotient.IsInt64() {
		return 0, fmt.Errorf("pricing result exceeds Rial range")
	}
	return quotient.Int64(), nil
}

func addMoney(left, right int64) (int64, error) {
	return roundBig(new(big.Int).Add(big.NewInt(left), big.NewInt(right)), big.NewInt(1))
}

func subtractMoney(left, right int64) (int64, error) {
	return roundBig(new(big.Int).Sub(big.NewInt(left), big.NewInt(right)), big.NewInt(1))
}
