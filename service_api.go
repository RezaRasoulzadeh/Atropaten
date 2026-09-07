package main

import (
	"fmt"

	"Atropaten/internal/application"
)

type ServiceParameterInput struct {
	ID           string   `json:"id"`
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	DefaultValue string   `json:"defaultValue"`
	Options      []string `json:"options"`
	MinValue     *string  `json:"minValue"`
	MaxValue     *string  `json:"maxValue"`
	Unit         string   `json:"unit"`
}

type ServiceInput struct {
	Name        string                      `json:"name"`
	Code        string                      `json:"code"`
	Category    string                      `json:"category"`
	Description string                      `json:"description"`
	Parameters  []ServiceParameterInput     `json:"parameters"`
	Components  []ServiceCostComponentInput `json:"components"`
	PricingRule *PricingRuleInput           `json:"pricingRule"`
}

type ServiceCostComponentInput struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ReferenceID   string `json:"referenceId"`
	UsageMode     string `json:"usageMode"`
	ParameterKey  string `json:"parameterKey"`
	UsageQuantity string `json:"usageQuantity"`
	Multiplier    string `json:"multiplier"`
	RateRial      int64  `json:"rateRial"`
	Percentage    string `json:"percentage"`
	RateBasis     string `json:"rateBasis"`
	Enabled       bool   `json:"enabled"`
	Notes         string `json:"notes"`
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
	ID           string   `json:"id"`
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	Position     int      `json:"position"`
	DefaultValue string   `json:"defaultValue"`
	Options      []string `json:"options"`
	MinValue     *string  `json:"minValue"`
	MaxValue     *string  `json:"maxValue"`
	Unit         string   `json:"unit"`
	Active       bool     `json:"active"`
}

type ServiceDTO struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Code        string                    `json:"code"`
	Category    string                    `json:"category"`
	Description string                    `json:"description"`
	Active      bool                      `json:"active"`
	CreatedAt   string                    `json:"createdAt"`
	UpdatedAt   string                    `json:"updatedAt"`
	Parameters  []ServiceParameterDTO     `json:"parameters"`
	Components  []ServiceCostComponentDTO `json:"components"`
	PricingRule *PricingRuleDTO           `json:"pricingRule"`
}

type ServiceCostComponentDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ReferenceID   string `json:"referenceId"`
	UsageMode     string `json:"usageMode"`
	ParameterKey  string `json:"parameterKey"`
	UsageQuantity string `json:"usageQuantity"`
	Multiplier    string `json:"multiplier"`
	RateRial      int64  `json:"rateRial"`
	Percentage    string `json:"percentage"`
	RateBasis     string `json:"rateBasis"`
	Enabled       bool   `json:"enabled"`
	Position      int    `json:"position"`
	Notes         string `json:"notes"`
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

func (a *App) CreateService(input ServiceInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Create(a.materialContext(), applicationServiceInput(input))
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
	view, err := service.Update(a.materialContext(), id, applicationServiceInput(input))
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
	view, err := service.AddParameter(a.materialContext(), id, applicationParameterInput(input))
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
	view, err := service.UpdateParameter(a.materialContext(), id, parameterID, applicationParameterInput(input))
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

func applicationServiceInput(input ServiceInput) application.ServiceInput {
	parameters := make([]application.ParameterInput, 0, len(input.Parameters))
	for _, parameter := range input.Parameters {
		parameters = append(parameters, applicationParameterInput(parameter))
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
	return application.ServiceInput{Name: input.Name, Code: input.Code, Category: input.Category, Description: input.Description, Parameters: parameters, Components: components, PricingRule: pricingRule}
}

func applicationCostComponentInput(input ServiceCostComponentInput) application.CostComponentInput {
	return application.CostComponentInput{ID: input.ID, Name: input.Name, Type: input.Type, ReferenceID: input.ReferenceID, UsageMode: input.UsageMode, ParameterKey: input.ParameterKey, UsageQuantity: input.UsageQuantity, Multiplier: input.Multiplier, RateRial: input.RateRial, Percentage: input.Percentage, RateBasis: input.RateBasis, Enabled: input.Enabled, Notes: input.Notes}
}

func applicationParameterInput(input ServiceParameterInput) application.ParameterInput {
	return application.ParameterInput{ID: input.ID, Key: input.Key, Label: input.Label, Type: input.Type, Required: input.Required, DefaultValue: input.DefaultValue, Options: input.Options, MinValue: input.MinValue, MaxValue: input.MaxValue, Unit: input.Unit}
}

func serviceDTO(view application.ServiceView) ServiceDTO {
	parameters := make([]ServiceParameterDTO, 0, len(view.Parameters))
	for _, parameter := range view.Parameters {
		parameters = append(parameters, ServiceParameterDTO{ID: parameter.ID, Key: parameter.Key, Label: parameter.Label, Type: parameter.Type, Required: parameter.Required, Position: parameter.Position, DefaultValue: parameter.DefaultValue, Options: parameter.Options, MinValue: parameter.MinValue, MaxValue: parameter.MaxValue, Unit: parameter.Unit, Active: parameter.Active})
	}
	components := make([]ServiceCostComponentDTO, 0, len(view.Components))
	for _, component := range view.Components {
		components = append(components, ServiceCostComponentDTO{ID: component.ID, Name: component.Name, Type: component.Type, ReferenceID: component.ReferenceID, UsageMode: component.UsageMode, ParameterKey: component.ParameterKey, UsageQuantity: component.UsageQuantity, Multiplier: component.Multiplier, RateRial: component.RateRial, Percentage: component.Percentage, RateBasis: component.RateBasis, Enabled: component.Enabled, Position: component.Position, Notes: component.Notes})
	}
	var pricingRule *PricingRuleDTO
	if view.PricingRule != nil {
		pricingRule = &PricingRuleDTO{ID: view.PricingRule.ID, Type: view.PricingRule.Type, FixedPriceRial: view.PricingRule.FixedPriceRial, MarkupPercentage: view.PricingRule.MarkupPercentage, FixedMarginRial: view.PricingRule.FixedMarginRial, PerUnitRateRial: view.PricingRule.PerUnitRateRial, ParameterKey: view.PricingRule.ParameterKey}
		for _, tier := range view.PricingRule.Tiers {
			pricingRule.Tiers = append(pricingRule.Tiers, PricingTierDTO{Position: tier.Position, MinimumQuantity: tier.MinimumQuantity, PriceRial: tier.PriceRial})
		}
	}
	return ServiceDTO{ID: view.ID, Name: view.Name, Code: view.Code, Category: view.Category, Description: view.Description, Active: view.Active, CreatedAt: view.CreatedAt, UpdatedAt: view.UpdatedAt, Parameters: parameters, Components: components, PricingRule: pricingRule}
}
