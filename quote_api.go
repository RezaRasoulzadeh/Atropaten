package main

import (
	"fmt"
	"time"

	"Atropaten/internal/application"
)

type QuoteInput struct {
	CustomerID   string  `json:"customerId"`
	ExpiryDate   *string `json:"expiryDate"`
	Notes        string  `json:"notes"`
	DiscountRial int64   `json:"discountRial"`
}
type QuoteDTO struct {
	ID                string         `json:"id"`
	QuoteNumber       string         `json:"quoteNumber"`
	CustomerID        string         `json:"customerId"`
	CustomerName      string         `json:"customerName"`
	CustomerPhone     string         `json:"customerPhone"`
	Notes             string         `json:"notes"`
	CreatedAt         string         `json:"createdAt"`
	UpdatedAt         string         `json:"updatedAt"`
	ExpiryDate        *string        `json:"expiryDate"`
	Status            string         `json:"status"`
	SubtotalRial      int64          `json:"subtotalRial"`
	DiscountRial      int64          `json:"discountRial"`
	TotalRial         int64          `json:"totalRial"`
	EstimatedCostRial int64          `json:"estimatedCostRial"`
	ConvertedOrderID  string         `json:"convertedOrderId"`
	Items             []QuoteItemDTO `json:"items"`
}
type QuoteItemDTO struct {
	ID                     string `json:"id"`
	Position               int    `json:"position"`
	ServiceID              string `json:"serviceId"`
	ServiceName            string `json:"serviceName"`
	ServiceCode            string `json:"serviceCode"`
	Quantity               string `json:"quantity"`
	QuantityUnit           string `json:"quantityUnit"`
	ResolvedParametersJSON string `json:"resolvedParametersJson"`
	CostBreakdownJSON      string `json:"costBreakdownJson"`
	PricingSnapshotJSON    string `json:"pricingSnapshotJson"`
	EstimatedCostRial      int64  `json:"estimatedCostRial"`
	SuggestedPriceRial     int64  `json:"suggestedPriceRial"`
	SellingPriceRial       int64  `json:"sellingPriceRial"`
	Notes                  string `json:"notes"`
}

func (a *App) quoteService() (*application.QuotesService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.quotes == nil {
		return nil, fmt.Errorf("quotes service is not initialized")
	}
	return a.quotes, nil
}
func (a *App) ListQuotes() ([]QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return nil, e
	}
	rows, e := s.List(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]QuoteDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, quoteDTO(r))
	}
	return out, nil
}
func (a *App) GetQuote(id string) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.Get(a.materialContext(), id)
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) CreateQuote(i QuoteInput) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	in, e := quoteInput(i)
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.Create(a.materialContext(), in)
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) UpdateQuote(id string, i QuoteInput) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	in, e := quoteInput(i)
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.Update(a.materialContext(), id, in)
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) AddQuoteItem(id string, i OrderItemInput) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.AddItem(a.materialContext(), id, itemInput(i))
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) ReplaceQuoteItem(id, itemID string, i OrderItemInput) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.ReplaceItem(a.materialContext(), id, itemID, itemInput(i))
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) RemoveQuoteItem(id, itemID string) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.RemoveItem(a.materialContext(), id, itemID)
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) ReorderQuoteItems(id string, ids []string) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.ReorderItems(a.materialContext(), id, ids)
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) ApplyQuoteDiscount(id string, discount int64) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.ApplyDiscount(a.materialContext(), id, discount)
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) UpdateQuoteStatus(id, status string) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.SetStatus(a.materialContext(), id, status)
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func (a *App) ConvertQuoteToOrder(id string) (QuoteDTO, error) {
	s, e := a.quoteService()
	if e != nil {
		return QuoteDTO{}, e
	}
	r, e := s.Convert(a.materialContext(), id)
	if e != nil {
		return QuoteDTO{}, e
	}
	return quoteDTO(r), nil
}
func quoteInput(i QuoteInput) (application.QuoteInput, error) {
	var d *time.Time
	if i.ExpiryDate != nil && *i.ExpiryDate != "" {
		v, e := time.Parse(time.RFC3339, *i.ExpiryDate)
		if e != nil {
			return application.QuoteInput{}, fmt.Errorf("expiryDate: invalid timestamp: %w", e)
		}
		d = &v
	}
	return application.QuoteInput{CustomerID: i.CustomerID, ExpiryDate: d, Notes: i.Notes, DiscountRial: i.DiscountRial}, nil
}
func quoteDTO(v application.QuoteView) QuoteDTO {
	o := QuoteDTO{ID: v.ID, QuoteNumber: v.QuoteNumber, CustomerID: v.CustomerID, CustomerName: v.CustomerName, CustomerPhone: v.CustomerPhone, Notes: v.Notes, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, ExpiryDate: v.ExpiryDate, Status: v.Status, SubtotalRial: v.SubtotalRial, DiscountRial: v.DiscountRial, TotalRial: v.TotalRial, EstimatedCostRial: v.EstimatedCostRial, ConvertedOrderID: v.ConvertedOrderID}
	for _, i := range v.Items {
		o.Items = append(o.Items, QuoteItemDTO{ID: i.ID, Position: i.Position, ServiceID: i.ServiceID, ServiceName: i.ServiceName, ServiceCode: i.ServiceCode, Quantity: i.Quantity, QuantityUnit: i.QuantityUnit, ResolvedParametersJSON: i.ResolvedParametersJSON, CostBreakdownJSON: i.CostBreakdownJSON, PricingSnapshotJSON: i.PricingSnapshotJSON, EstimatedCostRial: i.EstimatedCostRial, SuggestedPriceRial: i.SuggestedPriceRial, SellingPriceRial: i.SellingPriceRial, Notes: i.Notes})
	}
	return o
}
