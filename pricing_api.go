package main

import (
	"fmt"

	"Atropaten/internal/application"
)

type PricingRequest struct {
	ServiceID                string            `json:"serviceId"`
	Parameters               map[string]string `json:"parameters"`
	ManualCosts              map[string]int64  `json:"manualCosts"`
	SellingPriceOverrideRial *int64            `json:"sellingPriceOverrideRial"`
}

type ResolvedParameterDTO struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Type       string `json:"type"`
	Value      string `json:"value"`
	Quantity   string `json:"quantity"`
	MaterialID string `json:"materialId"`
	Unit       string `json:"unit"`
}

type PricingComponentDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Enabled       bool   `json:"enabled"`
	UsageQuantity string `json:"usageQuantity"`
	RateRial      int64  `json:"rateRial"`
	Percentage    string `json:"percentage"`
	AmountRial    int64  `json:"amountRial"`
	Explanation   string `json:"explanation"`
}

type PricingDTO struct {
	ServiceID                 string                 `json:"serviceId"`
	ServiceName               string                 `json:"serviceName"`
	Parameters                []ResolvedParameterDTO `json:"parameters"`
	Components                []PricingComponentDTO  `json:"components"`
	EstimatedCostRial         int64                  `json:"estimatedCostRial"`
	SuggestedSellingPriceRial int64                  `json:"suggestedSellingPriceRial"`
	EffectiveSellingPriceRial int64                  `json:"effectiveSellingPriceRial"`
	ProfitRial                int64                  `json:"profitRial"`
	MarginPercentage          string                 `json:"marginPercentage"`
	Warnings                  []string               `json:"warnings"`
	BelowCost                 bool                   `json:"belowCost"`
}

func (a *App) pricingService() (*application.PricingService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.pricing == nil {
		return nil, fmt.Errorf("pricing service is not initialized")
	}
	return a.pricing, nil
}

func (a *App) CalculateServicePrice(request PricingRequest) (PricingDTO, error) {
	service, err := a.pricingService()
	if err != nil {
		return PricingDTO{}, err
	}
	view, err := service.Calculate(a.materialContext(), application.PricingRequest{ServiceID: request.ServiceID, Parameters: request.Parameters, ManualCosts: request.ManualCosts, SellingPriceOverrideRial: request.SellingPriceOverrideRial})
	if err != nil {
		return PricingDTO{}, err
	}
	return pricingDTO(view), nil
}

func pricingDTO(view application.PricingView) PricingDTO {
	dto := PricingDTO{ServiceID: view.ServiceID, ServiceName: view.ServiceName, EstimatedCostRial: view.EstimatedCostRial, SuggestedSellingPriceRial: view.SuggestedSellingPriceRial, EffectiveSellingPriceRial: view.EffectiveSellingPriceRial, ProfitRial: view.ProfitRial, MarginPercentage: view.MarginPercentage, Warnings: view.Warnings, BelowCost: view.BelowCost}
	for _, parameter := range view.Parameters {
		dto.Parameters = append(dto.Parameters, ResolvedParameterDTO{Key: parameter.Key, Label: parameter.Label, Type: parameter.Type, Value: parameter.Value, Quantity: parameter.Quantity, MaterialID: parameter.MaterialID, Unit: parameter.Unit})
	}
	for _, component := range view.Components {
		dto.Components = append(dto.Components, PricingComponentDTO{ID: component.ID, Name: component.Name, Type: component.Type, Enabled: component.Enabled, UsageQuantity: component.UsageQuantity, RateRial: component.RateRial, Percentage: component.Percentage, AmountRial: component.AmountRial, Explanation: component.Explanation})
	}
	return dto
}
