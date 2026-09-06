package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type QuoteRepository interface {
	ListQuotes(context.Context) ([]domain.Quote, error)
	GetQuote(context.Context, string) (domain.Quote, error)
	CreateQuote(context.Context, domain.Quote) error
	SaveQuote(context.Context, domain.Quote) error
	ConvertQuoteToOrder(context.Context, string, string) (string, error)
}
type QuoteView struct {
	ID, QuoteNumber, CustomerID, CustomerName, CustomerPhone, Notes string
	CreatedAt, UpdatedAt                                            string
	ExpiryDate                                                      *string
	Status                                                          string
	SubtotalRial, DiscountRial, TotalRial, EstimatedCostRial        int64
	ConvertedOrderID                                                string
	Items                                                           []QuoteItemView
}
type QuoteItemView struct {
	ID                                                             string
	Position                                                       int
	ServiceID, ServiceName, ServiceCode, Quantity, QuantityUnit    string
	ResolvedParametersJSON, CostBreakdownJSON, PricingSnapshotJSON string
	EstimatedCostRial, SuggestedPriceRial, SellingPriceRial        int64
	Notes                                                          string
}
type QuoteInput struct {
	CustomerID   string
	ExpiryDate   *time.Time
	Notes        string
	DiscountRial int64
}

type QuotesService struct {
	repository QuoteRepository
	customers  OrderCustomerLookup
	pricing    *PricingService
	now        func() time.Time
}

func NewQuotesService(repository QuoteRepository, customers OrderCustomerLookup, pricing *PricingService) *QuotesService {
	return &QuotesService{repository: repository, customers: customers, pricing: pricing, now: time.Now}
}
func (s *QuotesService) List(ctx context.Context) ([]QuoteView, error) {
	rows, e := s.repository.ListQuotes(ctx)
	if e != nil {
		return nil, e
	}
	out := make([]QuoteView, 0, len(rows))
	for _, q := range rows {
		out = append(out, quoteView(q))
	}
	return out, nil
}
func (s *QuotesService) Get(ctx context.Context, id string) (QuoteView, error) {
	q, e := s.repository.GetQuote(ctx, strings.TrimSpace(id))
	if e != nil {
		return QuoteView{}, e
	}
	return quoteView(q), nil
}
func (s *QuotesService) Create(ctx context.Context, input QuoteInput) (QuoteView, error) {
	q := domain.NewQuote("", input.CustomerID, s.now())
	if e := s.applyInput(ctx, &q, input); e != nil {
		return QuoteView{}, e
	}
	id, e := randomID("QUO-")
	if e != nil {
		return QuoteView{}, e
	}
	q.ID = id
	if e := q.RecalculateTotals(); e != nil {
		return QuoteView{}, e
	}
	if e := s.repository.CreateQuote(ctx, q); e != nil {
		return QuoteView{}, e
	}
	return s.Get(ctx, id)
}
func (s *QuotesService) Update(ctx context.Context, id string, input QuoteInput) (QuoteView, error) {
	q, e := s.repository.GetQuote(ctx, id)
	if e != nil {
		return QuoteView{}, e
	}
	if q.Status != domain.QuoteDraft {
		return QuoteView{}, fmt.Errorf("only draft quotes can be edited")
	}
	if e = s.applyInput(ctx, &q, input); e != nil {
		return QuoteView{}, e
	}
	q.UpdatedAt = s.now().UTC()
	if e = q.RecalculateTotals(); e != nil {
		return QuoteView{}, e
	}
	if e = s.repository.SaveQuote(ctx, q); e != nil {
		return QuoteView{}, e
	}
	return quoteView(q), nil
}
func (s *QuotesService) applyInput(ctx context.Context, q *domain.Quote, input QuoteInput) error {
	q.CustomerID = strings.TrimSpace(input.CustomerID)
	q.CustomerNameSnapshot, q.CustomerPhoneSnapshot = "", ""
	q.ExpiryDate = input.ExpiryDate
	q.Notes = strings.TrimSpace(input.Notes)
	q.DiscountRial = input.DiscountRial
	if q.CustomerID != "" {
		if s.customers == nil {
			return fmt.Errorf("customer lookup unavailable")
		}
		c, e := s.customers.GetCustomer(ctx, q.CustomerID)
		if e != nil {
			return e
		}
		q.CustomerNameSnapshot, q.CustomerPhoneSnapshot = c.Name, c.Phone
	}
	return nil
}
func (s *QuotesService) AddItem(ctx context.Context, id string, input OrderItemInput) (QuoteView, error) {
	return s.saveConfiguredItem(ctx, id, "", input)
}
func (s *QuotesService) ReplaceItem(ctx context.Context, id, itemID string, input OrderItemInput) (QuoteView, error) {
	q, e := s.repository.GetQuote(ctx, id)
	if e != nil {
		return QuoteView{}, e
	}
	pos := -1
	for i, item := range q.Items {
		if item.ID == itemID {
			pos = i
			break
		}
	}
	if pos < 0 {
		return QuoteView{}, fmt.Errorf("quote item not found")
	}
	return s.saveConfiguredItem(ctx, id, itemID, input)
}
func (s *QuotesService) saveConfiguredItem(ctx context.Context, id, itemID string, input OrderItemInput) (QuoteView, error) {
	q, e := s.repository.GetQuote(ctx, id)
	if e != nil {
		return QuoteView{}, e
	}
	if q.Status != domain.QuoteDraft {
		return QuoteView{}, fmt.Errorf("only draft quotes can be changed")
	}
	if s.pricing == nil {
		return QuoteView{}, fmt.Errorf("pricing service unavailable")
	}
	price, e := s.pricing.Calculate(ctx, PricingRequest{ServiceID: input.ServiceID, Parameters: input.Parameters, ManualCosts: input.ManualCosts, SellingPriceOverrideRial: input.SellingPriceOverrideRial})
	if e != nil {
		return QuoteView{}, e
	}
	qty, e := parseOrderQuantity(input.Quantity, price)
	if e != nil {
		return QuoteView{}, e
	}
	parametersJSON, _ := json.Marshal(price.Parameters)
	componentsJSON, _ := json.Marshal(price.Components)
	snapshotJSON, _ := json.Marshal(price)
	item := domain.QuoteItem{QuoteID: q.ID, ServiceID: price.ServiceID, ServiceNameSnapshot: price.ServiceName, ServiceCodeSnapshot: price.ServiceCode, Quantity: qty, QuantityUnit: strings.TrimSpace(input.QuantityUnit), ResolvedParametersJSON: string(parametersJSON), CostBreakdownJSON: string(componentsJSON), PricingSnapshotJSON: string(snapshotJSON), EstimatedCostRial: price.EstimatedCostRial, SuggestedPriceRial: price.SuggestedSellingPriceRial, SellingPriceRial: price.EffectiveSellingPriceRial, Notes: strings.TrimSpace(input.Notes)}
	if item.QuantityUnit == "" {
		item.QuantityUnit = "unit"
	}
	if itemID == "" {
		item.ID, e = randomID("QITM-")
		if e != nil {
			return QuoteView{}, e
		}
		q.Items = append(q.Items, item)
	} else {
		for i := range q.Items {
			if q.Items[i].ID == itemID {
				item.ID = itemID
				q.Items[i] = item
			}
		}
	}
	for i := range q.Items {
		q.Items[i].Position = i
		q.Items[i].QuoteID = q.ID
	}
	q.UpdatedAt = s.now().UTC()
	if e = q.RecalculateTotals(); e != nil {
		return QuoteView{}, e
	}
	if e = s.repository.SaveQuote(ctx, q); e != nil {
		return QuoteView{}, e
	}
	return quoteView(q), nil
}
func (s *QuotesService) RemoveItem(ctx context.Context, id, itemID string) (QuoteView, error) {
	q, e := s.repository.GetQuote(ctx, id)
	if e != nil {
		return QuoteView{}, e
	}
	if q.Status != domain.QuoteDraft {
		return QuoteView{}, fmt.Errorf("only draft quotes can be changed")
	}
	out := q.Items[:0]
	found := false
	for _, item := range q.Items {
		if item.ID == itemID {
			found = true
			continue
		}
		out = append(out, item)
	}
	if !found {
		return QuoteView{}, fmt.Errorf("quote item not found")
	}
	q.Items = out
	for i := range q.Items {
		q.Items[i].Position = i
	}
	q.UpdatedAt = s.now().UTC()
	if e = q.RecalculateTotals(); e != nil {
		return QuoteView{}, e
	}
	if e = s.repository.SaveQuote(ctx, q); e != nil {
		return QuoteView{}, e
	}
	return quoteView(q), nil
}
func (s *QuotesService) ReorderItems(ctx context.Context, id string, ids []string) (QuoteView, error) {
	q, e := s.repository.GetQuote(ctx, id)
	if e != nil {
		return QuoteView{}, e
	}
	if q.Status != domain.QuoteDraft {
		return QuoteView{}, fmt.Errorf("only draft quotes can be changed")
	}
	if len(ids) != len(q.Items) {
		return QuoteView{}, fmt.Errorf("all quote items must be included")
	}
	byID := map[string]domain.QuoteItem{}
	for _, item := range q.Items {
		byID[item.ID] = item
	}
	ordered := make([]domain.QuoteItem, 0, len(ids))
	for _, itemID := range ids {
		item, ok := byID[itemID]
		if !ok {
			return QuoteView{}, fmt.Errorf("unknown quote item %q", itemID)
		}
		ordered = append(ordered, item)
	}
	for i := range ordered {
		ordered[i].Position = i
	}
	q.Items = ordered
	q.UpdatedAt = s.now().UTC()
	if e = s.repository.SaveQuote(ctx, q); e != nil {
		return QuoteView{}, e
	}
	return quoteView(q), nil
}
func (s *QuotesService) ApplyDiscount(ctx context.Context, id string, discount int64) (QuoteView, error) {
	q, e := s.repository.GetQuote(ctx, id)
	if e != nil {
		return QuoteView{}, e
	}
	if q.Status != domain.QuoteDraft {
		return QuoteView{}, fmt.Errorf("only draft quotes can be changed")
	}
	q.DiscountRial = discount
	if e = q.RecalculateTotals(); e != nil {
		return QuoteView{}, e
	}
	q.UpdatedAt = s.now().UTC()
	if e = s.repository.SaveQuote(ctx, q); e != nil {
		return QuoteView{}, e
	}
	return quoteView(q), nil
}
func (s *QuotesService) SetStatus(ctx context.Context, id, status string) (QuoteView, error) {
	q, e := s.repository.GetQuote(ctx, id)
	if e != nil {
		return QuoteView{}, e
	}
	next := domain.QuoteStatus(status)
	if !domain.ValidQuoteTransition(q.Status, next) {
		return QuoteView{}, fmt.Errorf("invalid quote status transition")
	}
	q.Status = next
	q.UpdatedAt = s.now().UTC()
	if e = s.repository.SaveQuote(ctx, q); e != nil {
		return QuoteView{}, e
	}
	return quoteView(q), nil
}
func (s *QuotesService) Convert(ctx context.Context, id string) (QuoteView, error) {
	orderID, e := randomID("ORD-")
	if e != nil {
		return QuoteView{}, e
	}
	if _, e = s.repository.ConvertQuoteToOrder(ctx, id, orderID); e != nil {
		return QuoteView{}, e
	}
	return s.Get(ctx, id)
}
func quoteView(q domain.Quote) QuoteView {
	v := QuoteView{ID: q.ID, QuoteNumber: q.QuoteNumber, CustomerID: q.CustomerID, CustomerName: q.CustomerNameSnapshot, CustomerPhone: q.CustomerPhoneSnapshot, Notes: q.Notes, CreatedAt: q.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: q.UpdatedAt.UTC().Format(time.RFC3339Nano), Status: string(q.Status), SubtotalRial: q.SubtotalRial, DiscountRial: q.DiscountRial, TotalRial: q.TotalRial, EstimatedCostRial: q.EstimatedCostRial, ConvertedOrderID: q.ConvertedOrderID}
	if q.ExpiryDate != nil {
		x := q.ExpiryDate.UTC().Format(time.RFC3339Nano)
		v.ExpiryDate = &x
	}
	for _, i := range q.Items {
		v.Items = append(v.Items, QuoteItemView{ID: i.ID, Position: i.Position, ServiceID: i.ServiceID, ServiceName: i.ServiceNameSnapshot, ServiceCode: i.ServiceCodeSnapshot, Quantity: i.Quantity.String(), QuantityUnit: i.QuantityUnit, ResolvedParametersJSON: i.ResolvedParametersJSON, CostBreakdownJSON: i.CostBreakdownJSON, PricingSnapshotJSON: i.PricingSnapshotJSON, EstimatedCostRial: i.EstimatedCostRial, SuggestedPriceRial: i.SuggestedPriceRial, SellingPriceRial: i.SellingPriceRial, Notes: i.Notes})
	}
	return v
}
