package application

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type PricingRequest struct {
	Quantity                 string
	ServiceID                string
	Parameters               map[string]string
	ManualCosts              map[string]int64
	SellingPriceOverrideRial *int64
}

type ResolvedParameterView struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Type       string `json:"type"`
	Value      string `json:"value"`
	Quantity   string `json:"quantity"`
	MaterialID string `json:"materialId"`
	Unit       string `json:"unit"`
}

type PricingComponentView struct {
	BatchQuantity string `json:"batchQuantity,omitempty"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ReferenceID   string `json:"referenceId"`
	MaterialID    string `json:"materialId"`
	ParameterKey  string `json:"parameterKey"`
	Enabled       bool   `json:"enabled"`
	UsageQuantity string `json:"usageQuantity"`
	RateRial      int64  `json:"rateRial"`
	Percentage    string `json:"percentage"`
	AmountRial    int64  `json:"amountRial"`
	Explanation   string `json:"explanation"`
}

type PricingView struct {
	BatchQuantity                 string                  `json:"batchQuantity,omitempty"`
	Layouts                       []domain.PrintLayout    `json:"layouts,omitempty"`
	ServiceID                     string                  `json:"serviceId"`
	ServiceName                   string                  `json:"serviceName"`
	ServiceCode                   string                  `json:"serviceCode"`
	FulfillmentMode               string                  `json:"fulfillmentMode"`
	DefaultOutsourcedCostRial     int64                   `json:"defaultOutsourcedCostRial"`
	DefaultOutsourcedShippingRial int64                   `json:"defaultOutsourcedShippingRial"`
	OutsourcedCostOverridden      bool                    `json:"outsourcedCostOverridden,omitempty"`
	OutsourcedShippingOverridden  bool                    `json:"outsourcedShippingOverridden,omitempty"`
	SellingPriceOverridden        bool                    `json:"sellingPriceOverridden,omitempty"`
	Parameters                    []ResolvedParameterView `json:"parameters"`
	Components                    []PricingComponentView  `json:"components"`
	EstimatedCostRial             int64                   `json:"estimatedCostRial"`
	SuggestedSellingPriceRial     int64                   `json:"suggestedSellingPriceRial"`
	EffectiveSellingPriceRial     int64                   `json:"effectiveSellingPriceRial"`
	ProfitRial                    int64                   `json:"profitRial"`
	MarginPercentage              string                  `json:"marginPercentage"`
	Warnings                      []string                `json:"warnings"`
	BelowCost                     bool                    `json:"belowCost"`
	RoundingStepRial              int64                   `json:"roundingStepRial"`
	FinishedWidthMM               string                  `json:"finishedWidthMM"`
	FinishedHeightMM              string                  `json:"finishedHeightMM"`
}

type PricingService struct {
	repository ServiceRepository
	material   MaterialLookup
	machine    MachineLookup
}

type PricingSettingsLookup interface {
	GetShopSettings(context.Context) (domain.ShopSettings, error)
}

func NewPricingService(repository ServiceRepository, material MaterialLookup, machine MachineLookup) *PricingService {
	return &PricingService{repository: repository, material: material, machine: machine}
}

func (s *PricingService) Calculate(ctx context.Context, request PricingRequest) (PricingView, error) {
	return s.calculate(ctx, request, map[string]bool{})
}

func (s *PricingService) calculate(ctx context.Context, request PricingRequest, stack map[string]bool) (PricingView, error) {
	serviceID := strings.TrimSpace(request.ServiceID)
	if stack[serviceID] {
		return PricingView{}, fmt.Errorf("service cost cycle detected at %q", serviceID)
	}
	stack[serviceID] = true
	defer delete(stack, serviceID)

	service, err := s.repository.GetService(ctx, serviceID)
	if err != nil {
		return PricingView{}, err
	}
	return s.calculateDefinition(ctx, request, stack, service)
}

// CalculateDraft runs the production pricing engine without saving a service.
func (s *PricingService) CalculateDraft(ctx context.Context, input ServiceInput, request PricingRequest) (PricingView, error) {
	definitions := NewServicesService(s.repository, s.material, s.machine)
	draft, err := definitions.parseDraft(ctx, input, "SVC-preview", nil, nil)
	if err != nil {
		return PricingView{}, err
	}
	service, err := domain.NewService("SVC-preview", draft, time.Now())
	if err != nil {
		return PricingView{}, err
	}
	return s.calculateDefinition(ctx, request, map[string]bool{"SVC-preview": true}, service)
}

func (s *PricingService) calculateDefinition(ctx context.Context, request PricingRequest, stack map[string]bool, service domain.Service) (PricingView, error) {
	service = service.WithRollSizeDefaults()
	var err error
	if !service.Active {
		return PricingView{}, fmt.Errorf("service is archived")
	}
	step := domain.DefaultMonetaryRoundingStepRial
	if lookup, ok := s.repository.(PricingSettingsLookup); ok {
		settings, settingsErr := lookup.GetShopSettings(ctx)
		if settingsErr != nil {
			return PricingView{}, fmt.Errorf("read monetary rounding settings: %w", settingsErr)
		}
		if settings.MonetaryRoundingStepRial > 0 {
			step = settings.MonetaryRoundingStepRial
		}
	}
	batchQuantity := domain.Quantity(0)
	if service.FinishedSize != nil && service.FinishedSize.QuantityParameterKey != "" && request.Quantity != "" {
		batchQuantity, err = domain.ParseQuantity(request.Quantity)
		if err != nil || batchQuantity <= 0 || batchQuantity%domain.QuantityScale != 0 {
			return PricingView{}, fmt.Errorf("layout quantity must be a positive whole number")
		}
		selected := make(map[string]string, len(request.Parameters)+1)
		for key, value := range request.Parameters {
			selected[key] = value
		}
		selected[service.FinishedSize.QuantityParameterKey] = batchQuantity.String()
		request.Parameters = selected
	}
	resolved, err := s.resolveParameters(ctx, service, request.Parameters)
	if err != nil {
		return PricingView{}, err
	}
	parameterMap := make(map[string]domain.ResolvedParameter, len(resolved))
	for _, parameter := range resolved {
		parameterMap[parameter.Key] = parameter
	}
	materials := make(map[string]domain.Material)
	machines := make(map[string]domain.Machine)
	serviceCosts := make(map[string]int64)
	for _, component := range service.Components {
		if !component.Enabled {
			continue
		}
		if component.Type == domain.CostMaterial {
			if s.material == nil {
				return PricingView{}, fmt.Errorf("material lookup is not available")
			}
			materialID := component.ReferenceID
			if component.UsageMode == domain.UsageParameter && component.ReferenceID == "" {
				parameter, exists := parameterMap[component.ParameterKey]
				if !exists || parameter.MaterialID == "" {
					return PricingView{}, fmt.Errorf("component %q: material parameter %q has no selected material", component.Name, component.ParameterKey)
				}
				materialID = parameter.MaterialID
			}
			material, getErr := s.material.Get(ctx, materialID)
			if getErr != nil {
				return PricingView{}, fmt.Errorf("component %q: material reference: %w", component.Name, getErr)
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
			machineID := component.ReferenceID
			if component.UsageMode == domain.UsageParameter && component.ReferenceID == "" {
				parameter, exists := parameterMap[component.ParameterKey]
				if !exists || parameter.MachineID == "" {
					return PricingView{}, fmt.Errorf("component %q: machine parameter %q has no selected machine", component.Name, component.ParameterKey)
				}
				machineID = parameter.MachineID
			}
			machine, getErr := s.machine.GetMachine(ctx, machineID)
			if getErr != nil {
				return PricingView{}, fmt.Errorf("component %q: %w", component.Name, getErr)
			}
			if !machine.Active {
				return PricingView{}, fmt.Errorf("component %q: machine is archived", component.Name)
			}
			machines[machine.ID] = machine
		}
		if component.Type == domain.CostService {
			nested, nestedErr := s.calculate(ctx, PricingRequest{ServiceID: component.ReferenceID}, stack)
			if nestedErr != nil {
				return PricingView{}, fmt.Errorf("component %q: %w", component.Name, nestedErr)
			}
			serviceCosts[component.ReferenceID] = nested.EstimatedCostRial
		}
	}
	result, err := domain.EvaluatePricing(domain.PricingInput{BatchQuantity: batchQuantity, Service: service, Parameters: parameterMap, Materials: materials, Machines: machines, ServiceCosts: serviceCosts, ManualCosts: request.ManualCosts, SellingPriceOverrideRial: request.SellingPriceOverrideRial, MonetaryRoundingStepRial: step})
	if err != nil {
		return PricingView{}, err
	}
	view := pricingView(service, result)
	view.RoundingStepRial = step
	if batchQuantity > 0 {
		view.BatchQuantity = batchQuantity.String()
		for i := range view.Components {
			view.Components[i].BatchQuantity = view.BatchQuantity
		}
	}
	seenLayouts := map[string]bool{}
	materialCosts := map[string]int64{}
	for _, component := range result.Components {
		if component.Enabled && component.Type == domain.CostMaterial && component.MaterialID != "" {
			materialCosts[component.MaterialID] += component.AmountRial
		}
	}
	for _, component := range result.Components {
		if !component.Enabled || component.MaterialID == "" || seenLayouts[component.MaterialID] || service.FinishedSize == nil || service.FinishedSize.QuantityParameterKey == "" {
			continue
		}
		material := materials[component.MaterialID]
		if _, ok := domain.ConsumptionStrategyForMaterialKind(material.Kind); !ok {
			continue
		}
		layout, e := domain.CalculatePrintLayout(material, service, parameterMap)
		if e != nil {
			return PricingView{}, e
		}
		layout.WasteCostRial, e = domain.CalculateMaterialWasteCost(layout, materialCosts[component.MaterialID])
		if e != nil {
			return PricingView{}, fmt.Errorf("calculate material waste cost: %w", e)
		}
		view.Layouts = append(view.Layouts, layout)
		seenLayouts[component.MaterialID] = true
	}
	dimensionValues := make(map[string]string, len(resolved))
	for _, parameter := range resolved {
		dimensionValues[parameter.Key] = parameter.Value
	}
	if width, height, dimensionErr := service.ResolveFinishedDimensions(dimensionValues); dimensionErr == nil {
		view.FinishedWidthMM = width.String()
		view.FinishedHeightMM = height.String()
	}
	return view, nil
}

func (s *PricingService) resolveParameters(ctx context.Context, service domain.Service, submitted map[string]string) ([]domain.ResolvedParameter, error) {
	definitions := service.Parameters
	machineRateParameters := make(map[string]struct{})
	for _, component := range service.Components {
		if component.Type == domain.CostMachine && strings.TrimSpace(component.RateParameterKey) != "" {
			machineRateParameters[component.RateParameterKey] = struct{}{}
		}
	}
	resolved := make([]domain.ResolvedParameter, 0, len(definitions))
	for _, definition := range definitions {
		predefinedKey := definition.PredefinedKey
		if _, dynamic := machineRateParameters[definition.Key]; dynamic {
			// Machine-rate choices are generated from the selected machine's
			// active profiles, not from the static predefined catalog.
			predefinedKey = ""
		}
		value, exists := submitted[definition.Key]
		if !exists || strings.TrimSpace(value) == "" {
			value = definition.DefaultValue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			if definition.Required {
				return nil, fmt.Errorf("parameter %q is required", definition.Label)
			}
			_, dynamicMachineRate := machineRateParameters[definition.Key]
			resolved = append(resolved, domain.ResolvedParameter{Key: definition.Key, Type: definition.Type, Value: value, PredefinedKey: predefinedKey, DynamicMachineRate: dynamicMachineRate})
			continue
		}
		_, dynamicMachineRate := machineRateParameters[definition.Key]
		item := domain.ResolvedParameter{Key: definition.Key, Type: definition.Type, Value: value, PredefinedKey: predefinedKey, DynamicMachineRate: dynamicMachineRate}
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
			if _, dynamic := machineRateParameters[definition.Key]; dynamic {
				break
			}
			if definition.MaterialSource == nil {
				valid := false
				if definition.PredefinedKey != "" {
					for _, option := range definition.PredefinedOptions {
						if option.Active && option.Code == value {
							valid = true
							break
						}
					}
				} else {
					for _, option := range definition.Options {
						if option == value {
							valid = true
							break
						}
					}
				}
				if !valid {
					return nil, fmt.Errorf("parameter %q must use one of its configured choices", definition.Label)
				}
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
		case domain.ParameterMachineReference:
			if len(definition.Options) > 0 {
				allowed := false
				for _, option := range definition.Options {
					if strings.TrimSpace(option) == value {
						allowed = true
						break
					}
				}
				if !allowed {
					return nil, fmt.Errorf("parameter %q must use one of its configured machines", definition.Label)
				}
			}
			if s.machine == nil {
				return nil, fmt.Errorf("machine lookup is not available")
			}
			machine, getErr := s.machine.GetMachine(ctx, value)
			if getErr != nil {
				return nil, fmt.Errorf("parameter %q: %w", definition.Label, getErr)
			}
			if !machine.Active {
				return nil, fmt.Errorf("parameter %q must reference an active machine", definition.Label)
			}
			item.MachineID = machine.ID
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
	if err := s.resolveMaterialBackedParameters(ctx, service, submitted, resolved); err != nil {
		return nil, err
	}
	return resolved, nil
}

func (s *PricingService) resolveMaterialBackedParameters(ctx context.Context, service domain.Service, submitted map[string]string, resolved []domain.ResolvedParameter) error {
	definitions := service.Parameters
	hasMaterialRequirement := false
	for _, definition := range definitions {
		if definition.MaterialSource != nil || definition.Type == domain.ParameterMaterialReference {
			hasMaterialRequirement = true
			break
		}
	}
	if !hasMaterialRequirement {
		return nil
	}
	if s.material == nil {
		return fmt.Errorf("material lookup is not available")
	}
	materials, ok := s.material.(interface {
		List(context.Context, bool) ([]domain.Material, error)
	})
	if !ok {
		return fmt.Errorf("material list lookup is not available for material-backed parameters")
	}
	items, err := materials.List(ctx, false)
	if err != nil {
		return fmt.Errorf("list compatible materials: %w", err)
	}
	selected := make(map[string]string, len(resolved))
	for _, parameter := range resolved {
		selected[parameter.Key] = parameter.Value
	}
	if len(service.MaterialVariants) > 0 {
		variant, ok := service.ResolveMaterialVariant(selected)
		if !ok {
			return fmt.Errorf("selected material combination is unavailable; choose a supported material option")
		}
		var selectedMaterial *domain.Material
		for _, material := range items {
			if material.Active && material.ID == variant.MaterialID {
				materialCopy := material
				selectedMaterial = &materialCopy
				break
			}
		}
		if selectedMaterial == nil {
			return fmt.Errorf("configured material %q is unavailable", variant.MaterialID)
		}
		compatible, compatibilityErr := domain.CompatibleMaterials(service, []domain.Material{*selectedMaterial}, selected)
		if compatibilityErr != nil {
			return compatibilityErr
		}
		if len(compatible) == 0 {
			return fmt.Errorf("configured material %q is incompatible with the selected service options", variant.MaterialID)
		}
		for index, definition := range definitions {
			if definition.MaterialSource != nil || definition.Type == domain.ParameterMaterialReference {
				resolved[index].MaterialID = variant.MaterialID
			}
		}
		return nil
	}
	candidates, err := domain.CompatibleMaterials(service, items, selected)
	if err != nil {
		return err
	}
	for index, definition := range definitions {
		if definition.MaterialSource == nil {
			continue
		}
		value := strings.TrimSpace(selected[definition.Key])
		if value == "" {
			continue
		}
		matches, matchErr := domain.CompatibleMaterials(service, items, selected)
		if matchErr != nil {
			return matchErr
		}
		if len(matches) == 0 {
			return fmt.Errorf("parameter %q does not match any compatible active material", definition.Label)
		}
		if len(matches) == 1 {
			resolved[index].MaterialID = matches[0].ID
		}
	}
	if len(candidates) == 0 {
		return fmt.Errorf("material-backed service configuration has no compatible active material")
	}
	if len(candidates) > 1 {
		for _, definition := range definitions {
			if definition.MaterialSource != nil && definition.Required && strings.TrimSpace(selected[definition.Key]) == "" {
				return fmt.Errorf("parameter %q is required", definition.Label)
			}
		}
		// A material cost cannot safely price an unresolved interchangeable set.
		for index, definition := range definitions {
			if definition.MaterialSource != nil && resolved[index].MaterialID == "" {
				return fmt.Errorf("parameter %q leaves %d compatible materials; select a material explicitly", definition.Label, len(candidates))
			}
		}
	}
	return nil
}

func pricingView(service domain.Service, result domain.PricingResult) PricingView {
	view := PricingView{ServiceID: result.ServiceID, ServiceName: result.ServiceName, ServiceCode: service.Code, FulfillmentMode: service.FulfillmentMode, DefaultOutsourcedCostRial: service.DefaultOutsourcedCostRial, DefaultOutsourcedShippingRial: service.DefaultOutsourcedShippingRial, EstimatedCostRial: result.EstimatedCostRial, SuggestedSellingPriceRial: result.SuggestedSellingPriceRial, EffectiveSellingPriceRial: result.EffectiveSellingPriceRial, ProfitRial: result.ProfitRial, MarginPercentage: result.MarginPercentage.String(), Warnings: result.Warnings, BelowCost: result.BelowCost}
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
		view.Components = append(view.Components, PricingComponentView{ID: component.ID, Name: component.Name, Type: string(component.Type), ReferenceID: component.ReferenceID, MaterialID: component.MaterialID, ParameterKey: component.ParameterKey, Enabled: component.Enabled, UsageQuantity: component.UsageQuantity.String(), RateRial: component.RateRial, Percentage: component.Percentage.String(), AmountRial: component.AmountRial, Explanation: component.Explanation})
	}
	return view
}
