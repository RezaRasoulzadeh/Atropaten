package main

import (
	"fmt"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

type ServiceParameterInput struct {
	ID             string                        `json:"id"`
	Key            string                        `json:"key"`
	Label          string                        `json:"label"`
	Type           string                        `json:"type"`
	Required       bool                          `json:"required"`
	DefaultValue   string                        `json:"defaultValue"`
	Options        []string                      `json:"options"`
	MinValue       *string                       `json:"minValue"`
	MaxValue       *string                       `json:"maxValue"`
	Unit           string                        `json:"unit"`
	PredefinedKey  string                        `json:"predefinedKey"`
	MaterialSource *MaterialParameterSourceInput `json:"materialSource"`
}

type MaterialParameterSourceInput struct {
	AllowedKinds         []string                       `json:"allowedKinds"`
	ExposedAttributeKey  string                         `json:"exposedAttributeKey"`
	ExposedAttributeKeys []string                       `json:"exposedAttributeKeys"`
	AllowedValues        []MaterialAttributeInput       `json:"allowedValues"`
	SelectMaterial       bool                           `json:"selectMaterial"`
	AdditionalFilters    []MaterialAttributeFilterInput `json:"additionalFilters"`
}

type MaterialAttributeFilterInput struct {
	Key   string                 `json:"key"`
	Value MaterialAttributeInput `json:"value"`
}

type ServiceInput struct {
	Name             string                        `json:"name"`
	Code             string                        `json:"code"`
	Category         string                        `json:"category"`
	Description      string                        `json:"description"`
	ImagePath        string                        `json:"imagePath"`
	DefaultUnit      string                        `json:"defaultUnit"`
	DefaultPriority  string                        `json:"defaultPriority"`
	Parameters       []ServiceParameterInput       `json:"parameters"`
	Components       []ServiceCostComponentInput   `json:"components"`
	PricingRule      *PricingRuleInput             `json:"pricingRule"`
	FinishedSize     *FinishedSizeInput            `json:"finishedSize"`
	MaterialVariants []ServiceMaterialVariantInput `json:"materialVariants"`
}

type ServiceMaterialVariantInput struct {
	ID         string            `json:"id"`
	MaterialID string            `json:"materialId"`
	Values     map[string]string `json:"values"`
	Position   int               `json:"position"`
	Active     bool              `json:"active"`
}

type FinishedSizeInput struct {
	ParameterKey         string                    `json:"parameterKey"`
	QuantityParameterKey string                    `json:"quantityParameterKey"`
	WidthParameterKey    string                    `json:"widthParameterKey"`
	HeightParameterKey   string                    `json:"heightParameterKey"`
	AllowCustom          bool                      `json:"allowCustom"`
	AllowRotation        bool                      `json:"allowRotation"`
	Options              []FinishedSizeOptionInput `json:"options"`
}

type FinishedSizeOptionInput struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Label    string `json:"label"`
	WidthMM  string `json:"widthMM"`
	HeightMM string `json:"heightMM"`
	Position int    `json:"position"`
	Active   bool   `json:"active"`
}

type ServiceCostComponentInput struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	ReferenceID      string `json:"referenceId"`
	UsageMode        string `json:"usageMode"`
	ParameterKey     string `json:"parameterKey"`
	RateID           string `json:"rateId"`
	RateParameterKey string `json:"rateParameterKey"`
	UsageQuantity    string `json:"usageQuantity"`
	Multiplier       string `json:"multiplier"`
	RateRial         int64  `json:"rateRial"`
	Percentage       string `json:"percentage"`
	RateBasis        string `json:"rateBasis"`
	Enabled          bool   `json:"enabled"`
	Notes            string `json:"notes"`
}

type PricingRuleInput struct {
	ID               string             `json:"id"`
	Type             string             `json:"type"`
	FixedPriceRial   int64              `json:"fixedPriceRial"`
	MarkupPercentage string             `json:"markupPercentage"`
	FixedMarginRial  int64              `json:"fixedMarginRial"`
	PerUnitRateRial  int64              `json:"perUnitRateRial"`
	ParameterKey     string             `json:"parameterKey"`
	Tiers            []PricingTierInput `json:"tiers"`
}

type PricingTierInput struct {
	Position        int    `json:"position"`
	MinimumQuantity string `json:"minimumQuantity"`
	PriceRial       int64  `json:"priceRial"`
}

type ServiceParameterDTO struct {
	ID                string                         `json:"id"`
	Key               string                         `json:"key"`
	Label             string                         `json:"label"`
	Type              string                         `json:"type"`
	Required          bool                           `json:"required"`
	Position          int                            `json:"position"`
	DefaultValue      string                         `json:"defaultValue"`
	Options           []string                       `json:"options"`
	MinValue          *string                        `json:"minValue"`
	MaxValue          *string                        `json:"maxValue"`
	Unit              string                         `json:"unit"`
	PredefinedKey     string                         `json:"predefinedKey"`
	PredefinedOptions []PredefinedParameterOptionDTO `json:"predefinedOptions"`
	MaterialSource    *MaterialParameterSourceDTO    `json:"materialSource"`
	Active            bool                           `json:"active"`
}

type PredefinedParameterOptionDTO struct {
	Code     string  `json:"code"`
	Label    string  `json:"label"`
	WidthMM  *string `json:"widthMM"`
	HeightMM *string `json:"heightMM"`
	Active   bool    `json:"active"`
	Position int     `json:"position"`
}

type PredefinedParameterDTO struct {
	Key       string                         `json:"key"`
	Label     string                         `json:"label"`
	ValueType string                         `json:"valueType"`
	Unit      string                         `json:"unit"`
	Active    bool                           `json:"active"`
	Position  int                            `json:"position"`
	Options   []PredefinedParameterOptionDTO `json:"options"`
}

type MaterialParameterSourceDTO struct {
	AllowedKinds         []string                     `json:"allowedKinds"`
	ExposedAttributeKey  string                       `json:"exposedAttributeKey"`
	ExposedAttributeKeys []string                     `json:"exposedAttributeKeys"`
	AllowedValues        []MaterialAttributeDTO       `json:"allowedValues"`
	SelectMaterial       bool                         `json:"selectMaterial"`
	AdditionalFilters    []MaterialAttributeFilterDTO `json:"additionalFilters"`
}

type MaterialAttributeFilterDTO struct {
	Key   string               `json:"key"`
	Value MaterialAttributeDTO `json:"value"`
}

type ServiceDTO struct {
	ID               string                      `json:"id"`
	Name             string                      `json:"name"`
	Code             string                      `json:"code"`
	Category         string                      `json:"category"`
	Description      string                      `json:"description"`
	ImagePath        string                      `json:"imagePath"`
	DefaultUnit      string                      `json:"defaultUnit"`
	DefaultPriority  string                      `json:"defaultPriority"`
	Active           bool                        `json:"active"`
	CreatedAt        string                      `json:"createdAt"`
	UpdatedAt        string                      `json:"updatedAt"`
	Parameters       []ServiceParameterDTO       `json:"parameters"`
	Components       []ServiceCostComponentDTO   `json:"components"`
	PricingRule      *PricingRuleDTO             `json:"pricingRule"`
	FinishedSize     *FinishedSizeDTO            `json:"finishedSize"`
	MaterialVariants []ServiceMaterialVariantDTO `json:"materialVariants"`
}

type ServiceMaterialVariantDTO struct {
	ID         string            `json:"id"`
	MaterialID string            `json:"materialId"`
	Values     map[string]string `json:"values"`
	Position   int               `json:"position"`
	Active     bool              `json:"active"`
}

type FinishedSizeDTO struct {
	ParameterKey         string                  `json:"parameterKey"`
	QuantityParameterKey string                  `json:"quantityParameterKey"`
	WidthParameterKey    string                  `json:"widthParameterKey"`
	HeightParameterKey   string                  `json:"heightParameterKey"`
	AllowCustom          bool                    `json:"allowCustom"`
	AllowRotation        bool                    `json:"allowRotation"`
	Options              []FinishedSizeOptionDTO `json:"options"`
}

type FinishedSizeOptionDTO struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Label    string `json:"label"`
	WidthMM  string `json:"widthMM"`
	HeightMM string `json:"heightMM"`
	Position int    `json:"position"`
	Active   bool   `json:"active"`
}

type ServiceMaterialOptionDTO struct {
	Value       string   `json:"value"`
	Label       string   `json:"label"`
	MaterialIDs []string `json:"materialIds"`
}

type ServiceMaterialOptionsDTO struct {
	ParameterKey string                     `json:"parameterKey"`
	Options      []ServiceMaterialOptionDTO `json:"options"`
	Message      string                     `json:"message"`
}

type ServiceCostComponentDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	ReferenceID      string `json:"referenceId"`
	UsageMode        string `json:"usageMode"`
	ParameterKey     string `json:"parameterKey"`
	RateID           string `json:"rateId"`
	RateParameterKey string `json:"rateParameterKey"`
	UsageQuantity    string `json:"usageQuantity"`
	Multiplier       string `json:"multiplier"`
	RateRial         int64  `json:"rateRial"`
	Percentage       string `json:"percentage"`
	RateBasis        string `json:"rateBasis"`
	Enabled          bool   `json:"enabled"`
	Position         int    `json:"position"`
	Notes            string `json:"notes"`
}

type PricingRuleDTO struct {
	ID               string           `json:"id"`
	Type             string           `json:"type"`
	FixedPriceRial   int64            `json:"fixedPriceRial"`
	MarkupPercentage string           `json:"markupPercentage"`
	FixedMarginRial  int64            `json:"fixedMarginRial"`
	PerUnitRateRial  int64            `json:"perUnitRateRial"`
	ParameterKey     string           `json:"parameterKey"`
	Tiers            []PricingTierDTO `json:"tiers"`
}

type PricingTierDTO struct {
	Position        int    `json:"position"`
	MinimumQuantity string `json:"minimumQuantity"`
	PriceRial       int64  `json:"priceRial"`
}

func (a *App) serviceService() (*application.ServicesService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.services == nil {
		return nil, fmt.Errorf("services service is not initialized")
	}
	return a.services, nil
}

func (a *App) ListServices(includeArchived bool) ([]ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return nil, err
	}
	views, err := service.List(a.materialContext(), includeArchived)
	if err != nil {
		return nil, err
	}
	result := make([]ServiceDTO, 0, len(views))
	for _, view := range views {
		result = append(result, serviceDTO(view))
	}
	return result, nil
}

func (a *App) GetService(id string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Get(a.materialContext(), id)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) GetServiceMaterialOptions(serviceID string, selected map[string]string) ([]ServiceMaterialOptionsDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return nil, err
	}
	options, err := service.MaterialOptions(a.materialContext(), serviceID, selected)
	if err != nil {
		return nil, err
	}
	result := make([]ServiceMaterialOptionsDTO, 0, len(options))
	for _, group := range options {
		dto := ServiceMaterialOptionsDTO{ParameterKey: group.ParameterKey, Message: group.Message}
		for _, option := range group.Options {
			dto.Options = append(dto.Options, ServiceMaterialOptionDTO{Value: option.Value, Label: option.Label, MaterialIDs: option.MaterialIDs})
		}
		result = append(result, dto)
	}
	return result, nil
}

func (a *App) ListPredefinedParameters() ([]PredefinedParameterDTO, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.database == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	definitions, err := a.database.ListPredefinedParameters(a.materialContext())
	if err != nil {
		return nil, err
	}
	result := make([]PredefinedParameterDTO, 0, len(definitions))
	for _, definition := range definitions {
		dto := PredefinedParameterDTO{Key: definition.Key, Label: definition.Label, ValueType: string(definition.ValueType), Unit: definition.Unit, Active: definition.Active, Position: definition.Position}
		for _, option := range definition.Options {
			item := PredefinedParameterOptionDTO{Code: option.Code, Label: option.Label, Active: option.Active, Position: option.Position}
			if option.WidthMM != nil {
				value := option.WidthMM.String()
				item.WidthMM = &value
			}
			if option.HeightMM != nil {
				value := option.HeightMM.String()
				item.HeightMM = &value
			}
			dto.Options = append(dto.Options, item)
		}
		result = append(result, dto)
	}
	return result, nil
}

func (a *App) CreateService(input ServiceInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	converted, err := applicationServiceInput(input)
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Create(a.materialContext(), converted)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) UpdateService(id string, input ServiceInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	converted, err := applicationServiceInput(input)
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Update(a.materialContext(), id, converted)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) ArchiveService(id string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Archive(a.materialContext(), id)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) ReactivateService(id string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Reactivate(a.materialContext(), id)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) DeleteService(id string) error {
	service, err := a.serviceService()
	if err != nil {
		return err
	}
	return service.Delete(a.materialContext(), id)
}

func (a *App) AddServiceParameter(id string, input ServiceParameterInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	converted, err := applicationParameterInput(input)
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.AddParameter(a.materialContext(), id, converted)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) UpdateServiceParameter(id, parameterID string, input ServiceParameterInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	converted, err := applicationParameterInput(input)
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.UpdateParameter(a.materialContext(), id, parameterID, converted)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) RemoveServiceParameter(id, parameterID string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.RemoveParameter(a.materialContext(), id, parameterID)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) ReorderServiceParameters(id string, parameterIDs []string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.ReorderParameters(a.materialContext(), id, parameterIDs)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) AddServiceCostComponent(id string, input ServiceCostComponentInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.AddCostComponent(a.materialContext(), id, applicationCostComponentInput(input))
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) UpdateServiceCostComponent(id, componentID string, input ServiceCostComponentInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.UpdateCostComponent(a.materialContext(), id, componentID, applicationCostComponentInput(input))
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) RemoveServiceCostComponent(id, componentID string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.RemoveCostComponent(a.materialContext(), id, componentID)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) ReorderServiceCostComponents(id string, componentIDs []string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.ReorderCostComponents(a.materialContext(), id, componentIDs)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func applicationServiceInput(input ServiceInput) (application.ServiceInput, error) {
	parameters := make([]application.ParameterInput, 0, len(input.Parameters))
	for _, parameter := range input.Parameters {
		converted, err := applicationParameterInput(parameter)
		if err != nil {
			return application.ServiceInput{}, err
		}
		parameters = append(parameters, converted)
	}
	components := make([]application.CostComponentInput, 0, len(input.Components))
	for _, component := range input.Components {
		components = append(components, applicationCostComponentInput(component))
	}
	var pricingRule *application.PricingRuleInput
	if input.PricingRule != nil {
		pricingRule = &application.PricingRuleInput{ID: input.PricingRule.ID, Type: input.PricingRule.Type, FixedPriceRial: input.PricingRule.FixedPriceRial, MarkupPercentage: input.PricingRule.MarkupPercentage, FixedMarginRial: input.PricingRule.FixedMarginRial, PerUnitRateRial: input.PricingRule.PerUnitRateRial, ParameterKey: input.PricingRule.ParameterKey}
		for _, tier := range input.PricingRule.Tiers {
			pricingRule.Tiers = append(pricingRule.Tiers, application.PricingTierInput{Position: tier.Position, MinimumQuantity: tier.MinimumQuantity, PriceRial: tier.PriceRial})
		}
	}
	var finishedSize *domain.ServiceFinishedSizeDefinition
	if input.FinishedSize != nil {
		converted := &domain.ServiceFinishedSizeDefinition{
			ParameterKey: input.FinishedSize.ParameterKey, QuantityParameterKey: input.FinishedSize.QuantityParameterKey, WidthParameterKey: input.FinishedSize.WidthParameterKey,
			HeightParameterKey: input.FinishedSize.HeightParameterKey, AllowCustom: input.FinishedSize.AllowCustom,
			AllowRotation: input.FinishedSize.AllowRotation,
		}
		for index, option := range input.FinishedSize.Options {
			width, err := domain.ParseQuantity(option.WidthMM)
			if err != nil {
				return application.ServiceInput{}, fmt.Errorf("finishedSize.options[%d].widthMM: %w", index, err)
			}
			height, err := domain.ParseQuantity(option.HeightMM)
			if err != nil {
				return application.ServiceInput{}, fmt.Errorf("finishedSize.options[%d].heightMM: %w", index, err)
			}
			converted.Options = append(converted.Options, domain.FinishedSizeOption{ID: option.ID, Code: option.Code, Label: option.Label, WidthMM: width, HeightMM: height, Position: index, Active: option.Active})
		}
		finishedSize = converted
	}
	variants := make([]domain.ServiceMaterialVariant, 0, len(input.MaterialVariants))
	for index, variant := range input.MaterialVariants {
		values := make(map[string]string, len(variant.Values))
		for key, value := range variant.Values {
			values[key] = value
		}
		variants = append(variants, domain.ServiceMaterialVariant{ID: variant.ID, MaterialID: variant.MaterialID, Values: values, Position: index, Active: variant.Active})
	}
	return application.ServiceInput{Name: input.Name, Code: input.Code, Category: input.Category, Description: input.Description, ImagePath: input.ImagePath, DefaultUnit: input.DefaultUnit, DefaultPriority: input.DefaultPriority, Parameters: parameters, Components: components, PricingRule: pricingRule, FinishedSize: finishedSize, MaterialVariants: variants}, nil
}

func applicationCostComponentInput(input ServiceCostComponentInput) application.CostComponentInput {
	return application.CostComponentInput{ID: input.ID, Name: input.Name, Type: input.Type, ReferenceID: input.ReferenceID, UsageMode: input.UsageMode, ParameterKey: input.ParameterKey, RateID: input.RateID, RateParameterKey: input.RateParameterKey, UsageQuantity: input.UsageQuantity, Multiplier: input.Multiplier, RateRial: input.RateRial, Percentage: input.Percentage, RateBasis: input.RateBasis, Enabled: input.Enabled, Notes: input.Notes}
}

func applicationParameterInput(input ServiceParameterInput) (application.ParameterInput, error) {
	var source *domain.MaterialParameterSource
	if input.MaterialSource != nil {
		source = &domain.MaterialParameterSource{ExposedAttributeKey: input.MaterialSource.ExposedAttributeKey, ExposedAttributeKeys: append([]string(nil), input.MaterialSource.ExposedAttributeKeys...), SelectMaterial: input.MaterialSource.SelectMaterial}
		for _, kind := range input.MaterialSource.AllowedKinds {
			source.AllowedKinds = append(source.AllowedKinds, domain.MaterialKind(kind))
		}
		for _, raw := range input.MaterialSource.AllowedValues {
			value, err := materialAttributeInputValue(raw)
			if err != nil {
				return application.ParameterInput{}, err
			}
			source.AllowedValues = append(source.AllowedValues, value)
		}
		for _, raw := range input.MaterialSource.AdditionalFilters {
			value, err := materialAttributeInputValue(raw.Value)
			if err != nil {
				return application.ParameterInput{}, err
			}
			value.Key = raw.Key
			source.AdditionalFilters = append(source.AdditionalFilters, domain.MaterialAttributeFilter{Key: raw.Key, Value: value})
		}
	}
	return application.ParameterInput{ID: input.ID, Key: input.Key, Label: input.Label, Type: input.Type, Required: input.Required, DefaultValue: input.DefaultValue, Options: input.Options, MinValue: input.MinValue, MaxValue: input.MaxValue, Unit: input.Unit, PredefinedKey: input.PredefinedKey, MaterialSource: source}, nil
}

func materialAttributeInputValue(raw MaterialAttributeInput) (domain.MaterialAttributeValue, error) {
	value := domain.MaterialAttributeValue{Key: raw.Key, ValueType: domain.MaterialAttributeValueType(raw.ValueType), IntegerValue: raw.IntegerValue, EnumCode: raw.EnumCode, TextValue: raw.TextValue, BooleanValue: raw.BooleanValue}
	if value.ValueType == domain.MaterialAttributeDecimal {
		parsed, err := domain.ParseQuantity(raw.DecimalValue)
		if err != nil {
			return value, fmt.Errorf("attribute %q decimalValue: %w", raw.Key, err)
		}
		value.DecimalValue = parsed
	}
	return value, nil
}

func serviceDTO(view application.ServiceView) ServiceDTO {
	parameters := make([]ServiceParameterDTO, 0, len(view.Parameters))
	for _, parameter := range view.Parameters {
		item := ServiceParameterDTO{ID: parameter.ID, Key: parameter.Key, Label: parameter.Label, Type: parameter.Type, Required: parameter.Required, Position: parameter.Position, DefaultValue: parameter.DefaultValue, Options: parameter.Options, MinValue: parameter.MinValue, MaxValue: parameter.MaxValue, Unit: parameter.Unit, PredefinedKey: parameter.PredefinedKey, MaterialSource: materialParameterSourceDTO(parameter.MaterialSource), Active: parameter.Active}
		for _, option := range parameter.PredefinedOptions {
			optionDTO := PredefinedParameterOptionDTO{Code: option.Code, Label: option.Label, Active: option.Active, Position: option.Position}
			if option.WidthMM != nil {
				value := option.WidthMM.String()
				optionDTO.WidthMM = &value
			}
			if option.HeightMM != nil {
				value := option.HeightMM.String()
				optionDTO.HeightMM = &value
			}
			item.PredefinedOptions = append(item.PredefinedOptions, optionDTO)
		}
		parameters = append(parameters, item)
	}
	components := make([]ServiceCostComponentDTO, 0, len(view.Components))
	for _, component := range view.Components {
		components = append(components, ServiceCostComponentDTO{ID: component.ID, Name: component.Name, Type: component.Type, ReferenceID: component.ReferenceID, UsageMode: component.UsageMode, ParameterKey: component.ParameterKey, RateID: component.RateID, RateParameterKey: component.RateParameterKey, UsageQuantity: component.UsageQuantity, Multiplier: component.Multiplier, RateRial: component.RateRial, Percentage: component.Percentage, RateBasis: component.RateBasis, Enabled: component.Enabled, Position: component.Position, Notes: component.Notes})
	}
	var pricingRule *PricingRuleDTO
	if view.PricingRule != nil {
		pricingRule = &PricingRuleDTO{ID: view.PricingRule.ID, Type: view.PricingRule.Type, FixedPriceRial: view.PricingRule.FixedPriceRial, MarkupPercentage: view.PricingRule.MarkupPercentage, FixedMarginRial: view.PricingRule.FixedMarginRial, PerUnitRateRial: view.PricingRule.PerUnitRateRial, ParameterKey: view.PricingRule.ParameterKey}
		for _, tier := range view.PricingRule.Tiers {
			pricingRule.Tiers = append(pricingRule.Tiers, PricingTierDTO{Position: tier.Position, MinimumQuantity: tier.MinimumQuantity, PriceRial: tier.PriceRial})
		}
	}
	var finishedSize *FinishedSizeDTO
	if view.FinishedSize != nil {
		finishedSize = &FinishedSizeDTO{ParameterKey: view.FinishedSize.ParameterKey, QuantityParameterKey: view.FinishedSize.QuantityParameterKey, WidthParameterKey: view.FinishedSize.WidthParameterKey, HeightParameterKey: view.FinishedSize.HeightParameterKey, AllowCustom: view.FinishedSize.AllowCustom, AllowRotation: view.FinishedSize.AllowRotation}
		for _, option := range view.FinishedSize.Options {
			finishedSize.Options = append(finishedSize.Options, FinishedSizeOptionDTO{ID: option.ID, Code: option.Code, Label: option.Label, WidthMM: option.WidthMM.String(), HeightMM: option.HeightMM.String(), Position: option.Position, Active: option.Active})
		}
	}
	variants := make([]ServiceMaterialVariantDTO, 0, len(view.MaterialVariants))
	for _, variant := range view.MaterialVariants {
		values := make(map[string]string, len(variant.Values))
		for key, value := range variant.Values {
			values[key] = value
		}
		variants = append(variants, ServiceMaterialVariantDTO{ID: variant.ID, MaterialID: variant.MaterialID, Values: values, Position: variant.Position, Active: variant.Active})
	}
	return ServiceDTO{ID: view.ID, Name: view.Name, Code: view.Code, Category: view.Category, Description: view.Description, ImagePath: view.ImagePath, DefaultUnit: view.DefaultUnit, DefaultPriority: view.DefaultPriority, Active: view.Active, CreatedAt: view.CreatedAt, UpdatedAt: view.UpdatedAt, Parameters: parameters, Components: components, PricingRule: pricingRule, FinishedSize: finishedSize, MaterialVariants: variants}
}

func materialParameterSourceDTO(source *domain.MaterialParameterSource) *MaterialParameterSourceDTO {
	if source == nil {
		return nil
	}
	dto := &MaterialParameterSourceDTO{ExposedAttributeKey: source.ExposedAttributeKey, ExposedAttributeKeys: append([]string(nil), source.ExposedAttributeKeys...), SelectMaterial: source.SelectMaterial}
	for _, kind := range source.AllowedKinds {
		dto.AllowedKinds = append(dto.AllowedKinds, string(kind))
	}
	for _, value := range source.AllowedValues {
		dto.AllowedValues = append(dto.AllowedValues, materialAttributeDTO(value))
	}
	for _, filter := range source.AdditionalFilters {
		dto.AdditionalFilters = append(dto.AdditionalFilters, MaterialAttributeFilterDTO{Key: filter.Key, Value: materialAttributeDTO(filter.Value)})
	}
	return dto
}
