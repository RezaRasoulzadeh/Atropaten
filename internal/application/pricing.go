package application

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"Atropaten/internal/domain"
)

type PricingRequest struct {
	ServiceID                string
	Parameters               map[string]string
	ManualCosts              map[string]int64
	SellingPriceOverrideRial *int64
}

type ResolvedParameterView struct {
	Key        string
	Label      string
	Type       string
	Value      string
	Quantity   string
	MaterialID string
	Unit       string
}

type PricingComponentView struct {
	ID            string
	Name          string
	Type          string
	Enabled       bool
	UsageQuantity string
	RateRial      int64
	Percentage    string
	AmountRial    int64
	Explanation   string
}

type PricingView struct {
	ServiceID                 string
	ServiceName               string
	Parameters                []ResolvedParameterView
	Components                []PricingComponentView
	EstimatedCostRial         int64
	SuggestedSellingPriceRial int64
	EffectiveSellingPriceRial int64
	ProfitRial                int64
	MarginPercentage          string
	Warnings                  []string
	BelowCost                 bool
}

type PricingService struct {
	repository ServiceRepository
	material   MaterialLookup
	machine    MachineLookup
}

func NewPricingService(repository ServiceRepository, material MaterialLookup, machine MachineLookup) *PricingService {
	return &PricingService{repository: repository, material: material, machine: machine}
}

func (s *PricingService) Calculate(ctx context.Context, request PricingRequest) (PricingView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(request.ServiceID))
	if err != nil {
		return PricingView{}, err
	}
	resolved, err := s.resolveParameters(ctx, service.Parameters, request.Parameters)
	if err != nil {
		return PricingView{}, err
	}
	parameterMap := make(map[string]domain.ResolvedParameter, len(resolved))
	for _, parameter := range resolved {
		parameterMap[parameter.Key] = parameter
	}
	materials := make(map[string]domain.Material)
	machines := make(map[string]domain.Machine)
	for _, component := range service.Components {
		if !component.Enabled {
			continue
		}
		if component.Type == domain.CostMaterial {
			if s.material == nil {
				return PricingView{}, fmt.Errorf("material lookup is not available")
			}
			material, getErr := s.material.Get(ctx, component.ReferenceID)
			if getErr != nil {
				return PricingView{}, fmt.Errorf("component %q: %w", component.Name, getErr)
			}
			if !material.Active {
				return PricingView{}, fmt.Errorf("component %q: material is archived", component.Name)
			}
			materials[material.ID] = material
		}
		if component.Type == domain.CostMachine {
			if s.machine == nil {
				return PricingView{}, fmt.Errorf("machine lookup is not available")
			}
			machine, getErr := s.machine.GetMachine(ctx, component.ReferenceID)
			if getErr != nil {
				return PricingView{}, fmt.Errorf("component %q: %w", component.Name, getErr)
			}
			if !machine.Active {
				return PricingView{}, fmt.Errorf("component %q: machine is archived", component.Name)
			}
			machines[machine.ID] = machine
		}
	}
	result, err := domain.EvaluatePricing(domain.PricingInput{Service: service, Parameters: parameterMap, Materials: materials, Machines: machines, ManualCosts: request.ManualCosts, SellingPriceOverrideRial: request.SellingPriceOverrideRial})
	if err != nil {
		return PricingView{}, err
	}
	return pricingView(service, result), nil
}

func (s *PricingService) resolveParameters(ctx context.Context, definitions []domain.ServiceParameter, submitted map[string]string) ([]domain.ResolvedParameter, error) {
	resolved := make([]domain.ResolvedParameter, 0, len(definitions))
	for _, definition := range definitions {
		value, exists := submitted[definition.Key]
		if !exists || strings.TrimSpace(value) == "" {
			value = definition.DefaultValue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			if definition.Required {
				return nil, fmt.Errorf("parameter %q is required", definition.Label)
			}
			resolved = append(resolved, domain.ResolvedParameter{Key: definition.Key, Type: definition.Type, Value: value})
			continue
		}
		item := domain.ResolvedParameter{Key: definition.Key, Type: definition.Type, Value: value}
		switch definition.Type {
		case domain.ParameterInteger:
			parsed, parseErr := strconv.ParseInt(value, 10, 64)
			if parseErr != nil || parsed < 0 {
				return nil, fmt.Errorf("parameter %q must be a non-negative integer", definition.Label)
			}
			if parsed > math.MaxInt64/domain.QuantityScale {
				return nil, fmt.Errorf("parameter %q is too large", definition.Label)
			}
			item.Quantity = domain.Quantity(parsed * domain.QuantityScale)
		case domain.ParameterDecimal:
			parsed, parseErr := domain.ParseQuantity(value)
			if parseErr != nil {
				return nil, fmt.Errorf("parameter %q must be a fixed-scale decimal", definition.Label)
			}
			item.Quantity = parsed
		case domain.ParameterBoolean:
			if value != "true" && value != "false" {
				return nil, fmt.Errorf("parameter %q must be true or false", definition.Label)
			}
		case domain.ParameterChoice:
			valid := false
			for _, option := range definition.Options {
				if option == value {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("parameter %q must use one of its configured choices", definition.Label)
			}
		case domain.ParameterMaterialReference:
			if s.material == nil {
				return nil, fmt.Errorf("material lookup is not available")
			}
			material, getErr := s.material.Get(ctx, value)
			if getErr != nil {
				return nil, fmt.Errorf("parameter %q: %w", definition.Label, getErr)
			}
			if !material.Active {
				return nil, fmt.Errorf("parameter %q must reference an active material", definition.Label)
			}
			item.MaterialID = material.ID
		default:
			return nil, fmt.Errorf("parameter %q has unsupported type", definition.Label)
		}
		if definition.Type == domain.ParameterInteger || definition.Type == domain.ParameterDecimal {
			if definition.MinValue != nil && item.Quantity < *definition.MinValue {
				return nil, fmt.Errorf("parameter %q is below its minimum", definition.Label)
			}
			if definition.MaxValue != nil && item.Quantity > *definition.MaxValue {
				return nil, fmt.Errorf("parameter %q is above its maximum", definition.Label)
			}
		}
		resolved = append(resolved, item)
	}
	return resolved, nil
}

func pricingView(service domain.Service, result domain.PricingResult) PricingView {
	view := PricingView{ServiceID: result.ServiceID, ServiceName: result.ServiceName, EstimatedCostRial: result.EstimatedCostRial, SuggestedSellingPriceRial: result.SuggestedSellingPriceRial, EffectiveSellingPriceRial: result.EffectiveSellingPriceRial, ProfitRial: result.ProfitRial, MarginPercentage: result.MarginPercentage.String(), Warnings: result.Warnings, BelowCost: result.BelowCost}
	for _, parameter := range result.Parameters {
		label := parameter.Key
		for _, definition := range service.Parameters {
			if definition.Key == parameter.Key {
				label = definition.Label
				view.Parameters = append(view.Parameters, ResolvedParameterView{Key: parameter.Key, Label: label, Type: string(parameter.Type), Value: parameter.Value, Quantity: parameter.Quantity.String(), MaterialID: parameter.MaterialID, Unit: definition.Unit})
				break
			}
		}
	}
	for _, component := range result.Components {
		view.Components = append(view.Components, PricingComponentView{ID: component.ID, Name: component.Name, Type: string(component.Type), Enabled: component.Enabled, UsageQuantity: component.UsageQuantity.String(), RateRial: component.RateRial, Percentage: component.Percentage.String(), AmountRial: component.AmountRial, Explanation: component.Explanation})
	}
	return view
}
