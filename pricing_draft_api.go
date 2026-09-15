package main

import "Atropaten/internal/application"

func (a *App) CalculateDraftServicePrice(input ServiceInput, request PricingRequest) (PricingDTO, error) {
	pricing, err := a.pricingService()
	if err != nil {
		return PricingDTO{}, err
	}
	draft, err := applicationServiceInput(input)
	if err != nil {
		return PricingDTO{}, err
	}
	result, err := pricing.CalculateDraft(a.materialContext(), draft, application.PricingRequest{Quantity: request.Quantity, Parameters: request.Parameters, ManualCosts: request.ManualCosts, SellingPriceOverrideRial: request.SellingPriceOverrideRial})
	if err != nil {
		return PricingDTO{}, err
	}
	return pricingDTO(result), nil
}
