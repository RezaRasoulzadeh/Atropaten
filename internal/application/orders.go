package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type OrderRepository interface {
	ListOrders(context.Context) ([]domain.Order, error)
	GetOrder(context.Context, string) (domain.Order, error)
	CreateOrder(context.Context, domain.Order) error
	SaveOrder(context.Context, domain.Order) error
}
type OrderCustomerLookup interface {
	GetCustomer(context.Context, string) (domain.Customer, error)
}

type OrderInput struct {
	CustomerID   string
	PromisedAt   *time.Time
	Priority     string
	Notes        string
	DiscountRial int64
}
type OrderItemInput struct {
	ServiceID                string
	Parameters               map[string]string
	ManualCosts              map[string]int64
	SellingPriceOverrideRial *int64
	Quantity                 string
	QuantityUnit             string
	Notes                    string
}
type OrderView struct {
	ID, OrderNumber, CustomerID, CustomerName, CustomerPhone, Notes string
	CreatedAt, UpdatedAt                                            string
	PromisedAt                                                      *string
	Priority, CommercialStatus, FulfillmentStatus, PaymentStatus    string
	SubtotalRial, DiscountRial, TotalRial, EstimatedCostRial        int64
	Items                                                           []OrderItemView
}
type OrderItemView struct {
	ID                                                             string
	Position                                                       int
	ServiceID, ServiceName, ServiceCode                            string
	Quantity, QuantityUnit                                         string
	ResolvedParametersJSON, CostBreakdownJSON, PricingSnapshotJSON string
	EstimatedCostRial, SuggestedPriceRial, SellingPriceRial        int64
	Notes                                                          string
}

type OrdersService struct {
	repository OrderRepository
	customers  OrderCustomerLookup
	pricing    *PricingService
	now        func() time.Time
}

func NewOrdersService(repository OrderRepository, customers OrderCustomerLookup, pricing *PricingService) *OrdersService {
	return &OrdersService{repository: repository, customers: customers, pricing: pricing, now: time.Now}
}
func (s *OrdersService) List(ctx context.Context) ([]OrderView, error) {
	rows, err := s.repository.ListOrders(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]OrderView, 0, len(rows))
	for _, row := range rows {
		out = append(out, orderView(row))
	}
	return out, nil
}
func (s *OrdersService) Get(ctx context.Context, id string) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, strings.TrimSpace(id))
	if err != nil {
		return OrderView{}, err
	}
	return orderView(row), nil
}
func (s *OrdersService) Create(ctx context.Context, input OrderInput) (OrderView, error) {
	row := domain.NewOrder("", strings.TrimSpace(input.CustomerID), s.now())
	if err := s.applyInput(ctx, &row, input); err != nil {
		return OrderView{}, err
	}
	id, err := randomID("ORD-")
	if err != nil {
		return OrderView{}, err
	}
	row.ID = id
	if err := row.RecalculateTotals(); err != nil {
		return OrderView{}, err
	}
	if err := s.repository.CreateOrder(ctx, row); err != nil {
		return OrderView{}, err
	}
	created, err := s.repository.GetOrder(ctx, row.ID)
	if err != nil {
		return OrderView{}, err
	}
	return orderView(created), nil
}
func (s *OrdersService) Update(ctx context.Context, id string, input OrderInput) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, strings.TrimSpace(id))
	if err != nil {
		return OrderView{}, err
	}
	if row.CommercialStatus != domain.CommercialDraft {
		return OrderView{}, fmt.Errorf("only draft orders can be edited")
	}
	if err := s.applyInput(ctx, &row, input); err != nil {
		return OrderView{}, err
	}
	row.UpdatedAt = s.now().UTC()
	if err := row.RecalculateTotals(); err != nil {
		return OrderView{}, err
	}
	if err := s.repository.SaveOrder(ctx, row); err != nil {
		return OrderView{}, err
	}
	return orderView(row), nil
}
func (s *OrdersService) applyInput(ctx context.Context, row *domain.Order, input OrderInput) error {
	row.CustomerID = strings.TrimSpace(input.CustomerID)
	row.CustomerNameSnapshot, row.CustomerPhoneSnapshot = "", ""
	row.PromisedAt = input.PromisedAt
	row.Notes = strings.TrimSpace(input.Notes)
	row.DiscountRial = input.DiscountRial
	if strings.TrimSpace(input.Priority) == "" {
		input.Priority = string(domain.PriorityNormal)
	}
	row.Priority = domain.Priority(input.Priority)
	if !domain.ValidPriority(row.Priority) {
		return fmt.Errorf("unsupported priority")
	}
	if row.CustomerID != "" {
		if s.customers == nil {
			return fmt.Errorf("customer lookup unavailable")
		}
		customer, err := s.customers.GetCustomer(ctx, row.CustomerID)
		if err != nil {
			return err
		}
		row.CustomerNameSnapshot = customer.Name
		row.CustomerPhoneSnapshot = customer.Phone
	}
	return nil
}
func (s *OrdersService) AddItem(ctx context.Context, id string, input OrderItemInput) (OrderView, error) {
	return s.saveConfiguredItem(ctx, id, -1, input)
}
func (s *OrdersService) ReplaceItem(ctx context.Context, id, itemID string, input OrderItemInput) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return OrderView{}, err
	}
	pos := -1
	for i, item := range row.Items {
		if item.ID == itemID {
			pos = i
			break
		}
	}
	if pos < 0 {
		return OrderView{}, fmt.Errorf("order item not found")
	}
	return s.saveConfiguredItem(ctx, id, pos, input)
}
func (s *OrdersService) saveConfiguredItem(ctx context.Context, id string, pos int, input OrderItemInput) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return OrderView{}, err
	}
	if row.CommercialStatus != domain.CommercialDraft {
		return OrderView{}, fmt.Errorf("only draft orders can be changed")
	}
	if s.pricing == nil {
		return OrderView{}, fmt.Errorf("pricing service unavailable")
	}
	price, err := s.pricing.Calculate(ctx, PricingRequest{ServiceID: input.ServiceID, Parameters: input.Parameters, ManualCosts: input.ManualCosts, SellingPriceOverrideRial: input.SellingPriceOverrideRial})
	if err != nil {
		return OrderView{}, err
	}
	qty, err := parseOrderQuantity(input.Quantity, price)
	if err != nil {
		return OrderView{}, err
	}
	parametersJSON, _ := json.Marshal(price.Parameters)
	componentsJSON, _ := json.Marshal(price.Components)
	snapshotJSON, _ := json.Marshal(price)
	item := domain.OrderItem{OrderID: row.ID, ServiceID: price.ServiceID, ServiceNameSnapshot: price.ServiceName, ServiceCodeSnapshot: price.ServiceCode, Quantity: qty, QuantityUnit: strings.TrimSpace(input.QuantityUnit), ResolvedParametersJSON: string(parametersJSON), CostBreakdownJSON: string(componentsJSON), PricingSnapshotJSON: string(snapshotJSON), EstimatedCostRial: price.EstimatedCostRial, SuggestedPriceRial: price.SuggestedSellingPriceRial, SellingPriceRial: price.EffectiveSellingPriceRial, Notes: strings.TrimSpace(input.Notes)}
	if item.QuantityUnit == "" {
		item.QuantityUnit = "unit"
	}
	if pos < 0 {
		item.ID, err = randomID("ITM-")
		if err != nil {
			return OrderView{}, err
		}
		row.Items = append(row.Items, item)
	} else {
		item.ID = row.Items[pos].ID
		row.Items[pos] = item
	}
	for i := range row.Items {
		row.Items[i].Position = i
		row.Items[i].OrderID = row.ID
	}
	row.UpdatedAt = s.now().UTC()
	if err := row.RecalculateTotals(); err != nil {
		return OrderView{}, err
	}
	if err := s.repository.SaveOrder(ctx, row); err != nil {
		return OrderView{}, err
	}
	return orderView(row), nil
}
func parseOrderQuantity(value string, price PricingView) (domain.Quantity, error) {
	if strings.TrimSpace(value) != "" {
		return domain.ParseQuantity(value)
	}
	for _, p := range price.Parameters {
		if p.Key == "quantity" && p.Quantity != "" {
			return domain.ParseQuantity(p.Quantity)
		}
	}
	return domain.Quantity(domain.QuantityScale), nil
}
func (s *OrdersService) RemoveItem(ctx context.Context, id, itemID string) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return OrderView{}, err
	}
	if row.CommercialStatus != domain.CommercialDraft {
		return OrderView{}, fmt.Errorf("only draft orders can be changed")
	}
	out := row.Items[:0]
	found := false
	for _, item := range row.Items {
		if item.ID == itemID {
			found = true
			continue
		}
		out = append(out, item)
	}
	if !found {
		return OrderView{}, fmt.Errorf("order item not found")
	}
	row.Items = out
	for i := range row.Items {
		row.Items[i].Position = i
	}
	if err := row.RecalculateTotals(); err != nil {
		return OrderView{}, err
	}
	row.UpdatedAt = s.now().UTC()
	if err := s.repository.SaveOrder(ctx, row); err != nil {
		return OrderView{}, err
	}
	return orderView(row), nil
}
func (s *OrdersService) ReorderItems(ctx context.Context, id string, itemIDs []string) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return OrderView{}, err
	}
	if len(itemIDs) != len(row.Items) {
		return OrderView{}, fmt.Errorf("all order items must be included")
	}
	byID := map[string]domain.OrderItem{}
	for _, item := range row.Items {
		byID[item.ID] = item
	}
	ordered := make([]domain.OrderItem, 0, len(itemIDs))
	for _, id := range itemIDs {
		item, ok := byID[id]
		if !ok {
			return OrderView{}, fmt.Errorf("unknown order item %q", id)
		}
		ordered = append(ordered, item)
	}
	for i := range ordered {
		ordered[i].Position = i
	}
	row.Items = ordered
	row.UpdatedAt = s.now().UTC()
	if err := s.repository.SaveOrder(ctx, row); err != nil {
		return OrderView{}, err
	}
	return orderView(row), nil
}
func (s *OrdersService) ApplyDiscount(ctx context.Context, id string, discount int64) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return OrderView{}, err
	}
	if row.CommercialStatus != domain.CommercialDraft {
		return OrderView{}, fmt.Errorf("only draft orders can be changed")
	}
	row.DiscountRial = discount
	if err := row.RecalculateTotals(); err != nil {
		return OrderView{}, err
	}
	row.UpdatedAt = s.now().UTC()
	if err := s.repository.SaveOrder(ctx, row); err != nil {
		return OrderView{}, err
	}
	return orderView(row), nil
}
func (s *OrdersService) SetCommercialStatus(ctx context.Context, id string, status string) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return OrderView{}, err
	}
	next := domain.CommercialStatus(status)
	if !domain.ValidCommercialTransition(row.CommercialStatus, next) {
		return OrderView{}, fmt.Errorf("invalid commercial transition")
	}
	row.CommercialStatus = next
	return s.saveStatus(ctx, row)
}
func (s *OrdersService) SetFulfillmentStatus(ctx context.Context, id string, status string) (OrderView, error) {
	row, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return OrderView{}, err
	}
	next := domain.FulfillmentStatus(status)
	if !domain.ValidFulfillmentTransition(row.FulfillmentStatus, next) {
		return OrderView{}, fmt.Errorf("invalid fulfillment transition")
	}
	row.FulfillmentStatus = next
	return s.saveStatus(ctx, row)
}
func (s *OrdersService) saveStatus(ctx context.Context, row domain.Order) (OrderView, error) {
	row.UpdatedAt = s.now().UTC()
	if err := s.repository.SaveOrder(ctx, row); err != nil {
		return OrderView{}, err
	}
	return orderView(row), nil
}
func orderView(o domain.Order) OrderView {
	v := OrderView{ID: o.ID, OrderNumber: o.OrderNumber, CustomerID: o.CustomerID, CustomerName: o.CustomerNameSnapshot, CustomerPhone: o.CustomerPhoneSnapshot, Notes: o.Notes, CreatedAt: o.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: o.UpdatedAt.UTC().Format(time.RFC3339Nano), Priority: string(o.Priority), CommercialStatus: string(o.CommercialStatus), FulfillmentStatus: string(o.FulfillmentStatus), PaymentStatus: string(o.PaymentStatus), SubtotalRial: o.SubtotalRial, DiscountRial: o.DiscountRial, TotalRial: o.TotalRial, EstimatedCostRial: o.EstimatedCostRial}
	if o.PromisedAt != nil {
		x := o.PromisedAt.UTC().Format(time.RFC3339Nano)
		v.PromisedAt = &x
	}
	for _, i := range o.Items {
		v.Items = append(v.Items, OrderItemView{ID: i.ID, Position: i.Position, ServiceID: i.ServiceID, ServiceName: i.ServiceNameSnapshot, ServiceCode: i.ServiceCodeSnapshot, Quantity: i.Quantity.String(), QuantityUnit: i.QuantityUnit, ResolvedParametersJSON: i.ResolvedParametersJSON, CostBreakdownJSON: i.CostBreakdownJSON, PricingSnapshotJSON: i.PricingSnapshotJSON, EstimatedCostRial: i.EstimatedCostRial, SuggestedPriceRial: i.SuggestedPriceRial, SellingPriceRial: i.SellingPriceRial, Notes: i.Notes})
	}
	return v
}
