package main

import (
	"Atropaten/internal/application"
	"fmt"
	"time"
)

type OrderInput struct {
	CustomerID   string  `json:"customerId"`
	PromisedAt   *string `json:"promisedAt"`
	Priority     string  `json:"priority"`
	Notes        string  `json:"notes"`
	DiscountRial int64   `json:"discountRial"`
}
type OrderItemInput struct {
	ServiceID                string            `json:"serviceId"`
	Parameters               map[string]string `json:"parameters"`
	ManualCosts              map[string]int64  `json:"manualCosts"`
	SellingPriceOverrideRial *int64            `json:"sellingPriceOverrideRial"`
	Quantity                 string            `json:"quantity"`
	QuantityUnit             string            `json:"quantityUnit"`
	Notes                    string            `json:"notes"`
}
type OrderDTO struct {
	ID                       string         `json:"id"`
	OrderNumber              string         `json:"orderNumber"`
	CustomerID               string         `json:"customerId"`
	CustomerName             string         `json:"customerName"`
	CustomerPhone            string         `json:"customerPhone"`
	Notes                    string         `json:"notes"`
	CreatedAt                string         `json:"createdAt"`
	UpdatedAt                string         `json:"updatedAt"`
	PromisedAt               *string        `json:"promisedAt"`
	Priority                 string         `json:"priority"`
	CommercialStatus         string         `json:"commercialStatus"`
	FulfillmentStatus        string         `json:"fulfillmentStatus"`
	PaymentStatus            string         `json:"paymentStatus"`
	SubtotalRial             int64          `json:"subtotalRial"`
	DiscountRial             int64          `json:"discountRial"`
	TotalRial                int64          `json:"totalRial"`
	EstimatedCostRial        int64          `json:"estimatedCostRial"`
	PaidRial                 int64          `json:"paidRial"`
	RemainingRial            int64          `json:"remainingRial"`
	QuoteID                  string         `json:"quoteId"`
	ProductionJobCount       int            `json:"productionJobCount"`
	CompletedProductionJobs  int            `json:"completedProductionJobs"`
	InProgressProductionJobs int            `json:"inProgressProductionJobs"`
	Items                    []OrderItemDTO `json:"items"`
}
type OrderItemDTO struct {
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

func (a *App) orderService() (*application.OrdersService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.orders == nil {
		return nil, fmt.Errorf("orders service is not initialized")
	}
	return a.orders, nil
}
func (a *App) ListOrders() ([]OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return nil, e
	}
	rows, e := s.List(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]OrderDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, orderDTO(r))
	}
	return out, nil
}
func (a *App) GetOrder(id string) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.Get(a.materialContext(), id)
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) CreateOrder(i OrderInput) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	input, e := orderInput(i)
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.Create(a.materialContext(), input)
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) UpdateOrder(id string, i OrderInput) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	input, e := orderInput(i)
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.Update(a.materialContext(), id, input)
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) AddOrderItem(id string, i OrderItemInput) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.AddItem(a.materialContext(), id, itemInput(i))
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) ReplaceOrderItem(id, itemID string, i OrderItemInput) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.ReplaceItem(a.materialContext(), id, itemID, itemInput(i))
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) RemoveOrderItem(id, itemID string) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.RemoveItem(a.materialContext(), id, itemID)
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) ReorderOrderItems(id string, ids []string) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.ReorderItems(a.materialContext(), id, ids)
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) ApplyOrderDiscount(id string, discount int64) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.ApplyDiscount(a.materialContext(), id, discount)
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) UpdateOrderCommercialStatus(id, status string) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.SetCommercialStatus(a.materialContext(), id, status)
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func (a *App) UpdateOrderFulfillmentStatus(id, status string) (OrderDTO, error) {
	s, e := a.orderService()
	if e != nil {
		return OrderDTO{}, e
	}
	r, e := s.SetFulfillmentStatus(a.materialContext(), id, status)
	if e != nil {
		return OrderDTO{}, e
	}
	return orderDTO(r), nil
}
func orderInput(i OrderInput) (application.OrderInput, error) {
	var p *time.Time
	if i.PromisedAt != nil && *i.PromisedAt != "" {
		v, e := time.Parse(time.RFC3339, *i.PromisedAt)
		if e != nil {
			return application.OrderInput{}, fmt.Errorf("promisedAt: invalid timestamp: %w", e)
		}
		p = &v
	}
	return application.OrderInput{CustomerID: i.CustomerID, PromisedAt: p, Priority: i.Priority, Notes: i.Notes, DiscountRial: i.DiscountRial}, nil
}
func itemInput(i OrderItemInput) application.OrderItemInput {
	return application.OrderItemInput{ServiceID: i.ServiceID, Parameters: i.Parameters, ManualCosts: i.ManualCosts, SellingPriceOverrideRial: i.SellingPriceOverrideRial, Quantity: i.Quantity, QuantityUnit: i.QuantityUnit, Notes: i.Notes}
}
func orderDTO(v application.OrderView) OrderDTO {
	o := OrderDTO{ID: v.ID, OrderNumber: v.OrderNumber, CustomerID: v.CustomerID, CustomerName: v.CustomerName, CustomerPhone: v.CustomerPhone, Notes: v.Notes, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, PromisedAt: v.PromisedAt, Priority: v.Priority, CommercialStatus: v.CommercialStatus, FulfillmentStatus: v.FulfillmentStatus, PaymentStatus: v.PaymentStatus, SubtotalRial: v.SubtotalRial, DiscountRial: v.DiscountRial, TotalRial: v.TotalRial, EstimatedCostRial: v.EstimatedCostRial, PaidRial: v.PaidRial, RemainingRial: v.RemainingRial, QuoteID: v.QuoteID, ProductionJobCount: v.ProductionJobCount, CompletedProductionJobs: v.CompletedProductionJobs, InProgressProductionJobs: v.InProgressProductionJobs}
	for _, i := range v.Items {
		o.Items = append(o.Items, OrderItemDTO{ID: i.ID, Position: i.Position, ServiceID: i.ServiceID, ServiceName: i.ServiceName, ServiceCode: i.ServiceCode, Quantity: i.Quantity, QuantityUnit: i.QuantityUnit, ResolvedParametersJSON: i.ResolvedParametersJSON, CostBreakdownJSON: i.CostBreakdownJSON, PricingSnapshotJSON: i.PricingSnapshotJSON, EstimatedCostRial: i.EstimatedCostRial, SuggestedPriceRial: i.SuggestedPriceRial, SellingPriceRial: i.SellingPriceRial, Notes: i.Notes})
	}
	return o
}
