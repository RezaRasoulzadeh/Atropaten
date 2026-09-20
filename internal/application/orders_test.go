package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

type orderRepositoryStub struct {
	rows        map[string]domain.Order
	createCalls int
}

func newOrderRepositoryStub() *orderRepositoryStub {
	return &orderRepositoryStub{rows: map[string]domain.Order{}}
}

func (r *orderRepositoryStub) ListOrders(context.Context) ([]domain.Order, error) {
	rows := make([]domain.Order, 0, len(r.rows))
	for _, order := range r.rows {
		rows = append(rows, order)
	}
	return rows, nil
}

func (r *orderRepositoryStub) GetOrder(_ context.Context, id string) (domain.Order, error) {
	order, ok := r.rows[id]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	order.Items = append([]domain.OrderItem(nil), order.Items...)
	return order, nil
}

func (r *orderRepositoryStub) CreateOrder(_ context.Context, order domain.Order) error {
	r.createCalls++
	order.OrderNumber = fmt.Sprintf("ORD-%04d", r.createCalls)
	if err := order.Validate(); err != nil {
		return err
	}
	r.rows[order.ID] = order
	return nil
}

func (r *orderRepositoryStub) SaveOrder(_ context.Context, order domain.Order) error {
	r.rows[order.ID] = order
	return nil
}

func (r *orderRepositoryStub) SaveOrderMetadata(ctx context.Context, order domain.Order) error {
	return r.SaveOrder(ctx, order)
}

func (r *orderRepositoryStub) DeleteOrder(_ context.Context, id string) error {
	delete(r.rows, id)
	return nil
}

func TestCreateOrderSavesConfiguredItemsTogether(t *testing.T) {
	orders, repository := testOrdersService(t)
	view, err := orders.Create(context.Background(), OrderInput{Items: []OrderItemInput{
		{ServiceID: "SVC-simple", Quantity: "2", QuantityUnit: "piece"},
		{ServiceID: "SVC-simple", Quantity: "1", QuantityUnit: "piece"},
	}})
	if err != nil {
		t.Fatalf("create order with items: %v", err)
	}
	if repository.createCalls != 1 || len(view.Items) != 2 || view.SubtotalRial != 6_000 {
		t.Fatalf("order save calls=%d view=%+v; want one create, two items, 6000 Rial subtotal", repository.createCalls, view)
	}
	if view.Items[0].Position != 0 || view.Items[1].Position != 1 || view.Items[0].EstimatedCostRial != 2_000 {
		t.Fatalf("item snapshots or positions are inconsistent: %+v", view.Items)
	}
}

func TestCreateOrderDoesNotPersistWhenAnyConfiguredItemIsInvalid(t *testing.T) {
	orders, repository := testOrdersService(t)
	_, err := orders.Create(context.Background(), OrderInput{Items: []OrderItemInput{
		{ServiceID: "SVC-simple", Quantity: "1"},
		{ServiceID: "SVC-simple", Quantity: "0"},
	}})
	if err == nil {
		t.Fatal("invalid configured item was accepted")
	}
	if repository.createCalls != 0 || len(repository.rows) != 0 {
		t.Fatalf("invalid order partially persisted: create calls=%d, orders=%d", repository.createCalls, len(repository.rows))
	}
}

func TestOutsourcedItemCostsAreOptionalLineCosts(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-outsourced", domain.ServiceDraft{
		Name: "External finishing", FulfillmentMode: domain.ServiceFulfillmentOutsourced,
		DefaultOutsourcedCostRial: 350, DefaultOutsourcedShippingRial: 50,
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingFixed, FixedPriceRial: 2_000},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	repository := newOrderRepositoryStub()
	pricing := NewPricingService(&serviceRepositoryStub{service: service}, materialLookupStub{}, machineLookupStub{})
	orders := NewOrdersService(repository, nil, pricing)
	view, err := orders.Create(context.Background(), OrderInput{Items: []OrderItemInput{{
		ServiceID: service.ID, Quantity: "1", QuantityUnit: "job", OutsourcedCostRial: 350, OutsourcedShippingRial: 50,
		OutsourcedSupplierID: "SUP-1", OutsourcedNotes: "Vendor quote",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	item := view.Items[0]
	if item.OutsourcedCostRial != 350 || item.OutsourcedShippingRial != 50 || item.OutsourcedSupplierID != "SUP-1" || item.OutsourcedNotes != "Vendor quote" {
		t.Fatalf("outsourced fields were not preserved: %+v", item)
	}
	if item.EstimatedCostRial != 1000 || item.SellingPriceRial != 2_000 {
		t.Fatalf("outsourced line cost changed the wrong totals: %+v", item)
	}
	defaultView, err := orders.Create(context.Background(), OrderInput{Items: []OrderItemInput{{
		ServiceID: service.ID, Quantity: "1", QuantityUnit: "job",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := defaultView.Items[0]; got.OutsourcedCostRial != 350 || got.OutsourcedShippingRial != 50 || got.EstimatedCostRial != 1000 {
		t.Fatalf("service outsourced defaults were not applied: %+v", got)
	}
	if _, err = orders.Create(context.Background(), OrderInput{Items: []OrderItemInput{{
		ServiceID: service.ID, Quantity: "1", OutsourcedCostRial: -1,
	}}}); err == nil {
		t.Fatal("negative outsourced cost was accepted")
	}
	inHouse, err := domain.NewService("SVC-in-house", domain.ServiceDraft{
		Name: "Internal finishing", PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingFixed, FixedPriceRial: 2_000},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	inHouseOrders := NewOrdersService(newOrderRepositoryStub(), nil, NewPricingService(&serviceRepositoryStub{service: inHouse}, materialLookupStub{}, machineLookupStub{}))
	inHouseView, err := inHouseOrders.Create(context.Background(), OrderInput{Items: []OrderItemInput{{
		ServiceID: inHouse.ID, Quantity: "1", OutsourcedCostRial: 500, OutsourcedShippingRial: 100, OutsourcedSupplierID: "SUP-1", OutsourcedNotes: "must be ignored",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := inHouseView.Items[0]; got.EstimatedCostRial != 0 || got.OutsourcedCostRial != 0 || got.OutsourcedShippingRial != 0 || got.OutsourcedSupplierID != "" || got.OutsourcedNotes != "" {
		t.Fatalf("in-house item retained outsourced fields: %+v", got)
	}
}

func testOrdersService(t *testing.T) (*OrdersService, *orderRepositoryStub) {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := domain.NewService("SVC-simple", domain.ServiceDraft{
		Name:        "Simple print",
		Components:  []domain.ServiceCostComponentDraft{{ID: "C-cost", Name: "Base cost", Type: domain.CostFixed, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, RateRial: 1_000, Enabled: true}},
		PricingRule: &domain.ServicePricingRuleDraft{Type: domain.PricingFixed, FixedPriceRial: 2_000},
	}, now)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	orderRepository := newOrderRepositoryStub()
	pricing := NewPricingService(&serviceRepositoryStub{service: service}, materialLookupStub{}, machineLookupStub{})
	return NewOrdersService(orderRepository, nil, pricing), orderRepository
}
