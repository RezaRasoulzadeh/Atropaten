package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type InvoiceRepository interface {
	ListInvoices(context.Context) ([]domain.Invoice, error)
	GetInvoice(context.Context, string) (domain.Invoice, error)
	GetInvoiceForOrder(context.Context, string) (domain.Invoice, error)
	SaveInvoice(context.Context, domain.Invoice) error
	DeleteDraftInvoice(context.Context, string) error
	PostInvoice(context.Context, string) error
	VoidInvoice(context.Context, string) error
}
type InvoiceOrderRepository interface {
	GetOrder(context.Context, string) (domain.Order, error)
}
type InvoiceView struct {
	ID, InvoiceNumber, CustomerID, CustomerName, CustomerPhone, OrderID, IssueDate, DueDate, Status, Notes, AccountingJournalEntryID, COGSJournalEntryID, CreatedAt, UpdatedAt string
	SubtotalRial, DiscountRial, TotalRial, PaidRial, RemainingRial                                                                                                             int64
	Items                                                                                                                                                                      []InvoiceItemView
}
type InvoiceItemView struct {
	ID, OrderItemID, Description, ServiceID, QuantityUnit, Notes string
	Position                                                     int
	Quantity                                                     string
	UnitPriceRial, LineTotalRial                                 int64
}
type InvoicesService struct {
	repository InvoiceRepository
	orders     InvoiceOrderRepository
	now        func() time.Time
}

func NewInvoicesService(r InvoiceRepository, o InvoiceOrderRepository) *InvoicesService {
	return &InvoicesService{repository: r, orders: o, now: time.Now}
}
func (s *InvoicesService) List(ctx context.Context) ([]InvoiceView, error) {
	v, e := s.repository.ListInvoices(ctx)
	if e != nil {
		return nil, e
	}
	out := make([]InvoiceView, 0, len(v))
	for _, i := range v {
		out = append(out, invoiceView(i))
	}
	return out, nil
}
func (s *InvoicesService) Get(ctx context.Context, id string) (InvoiceView, error) {
	v, e := s.repository.GetInvoice(ctx, id)
	if e != nil {
		return InvoiceView{}, e
	}
	return invoiceView(v), nil
}
func (s *InvoicesService) CreateFromOrder(ctx context.Context, orderID string) (InvoiceView, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return InvoiceView{}, fmt.Errorf("order is required")
	}
	order, e := s.orders.GetOrder(ctx, orderID)
	if e != nil {
		return InvoiceView{}, e
	}
	if _, e = s.repository.GetInvoiceForOrder(ctx, orderID); e == nil {
		return InvoiceView{}, domain.ErrInvoiceOrderExists
	} else if !errors.Is(e, domain.ErrInvoiceNotFound) {
		return InvoiceView{}, e
	}
	now := s.now().UTC()
	id := mustID("INV-")
	name := order.CustomerNameSnapshot
	if name == "" {
		name = "Walk-in customer"
	}
	v := domain.Invoice{ID: id, CustomerID: order.CustomerID, CustomerNameSnapshot: name, CustomerPhoneSnapshot: order.CustomerPhoneSnapshot, OrderID: order.ID, IssueDate: now, Status: domain.InvoiceDraft, Notes: order.Notes, SubtotalRial: order.SubtotalRial, DiscountRial: order.DiscountRial, TotalRial: order.TotalRial, CreatedAt: now, UpdatedAt: now}
	for position, item := range order.Items {
		v.Items = append(v.Items, domain.InvoiceItem{ID: mustID("INVL-"), InvoiceID: id, OrderItemID: item.ID, Position: position, DescriptionSnapshot: item.ServiceNameSnapshot, ServiceID: item.ServiceID, QuantityUnits: int64(item.Quantity), QuantityUnit: item.QuantityUnit, UnitPriceRial: item.SellingPriceRial, LineTotalRial: item.SellingPriceRial, Notes: item.Notes})
	}
	if e = v.Validate(); e != nil {
		return InvoiceView{}, e
	}
	if e = s.repository.SaveInvoice(ctx, v); e != nil {
		return InvoiceView{}, e
	}
	return s.Get(ctx, id)
}
func (s *InvoicesService) Post(ctx context.Context, id string) (InvoiceView, error) {
	if e := s.repository.PostInvoice(ctx, id); e != nil {
		return InvoiceView{}, e
	}
	return s.Get(ctx, id)
}
func (s *InvoicesService) Void(ctx context.Context, id string) (InvoiceView, error) {
	if e := s.repository.VoidInvoice(ctx, id); e != nil {
		return InvoiceView{}, e
	}
	return s.Get(ctx, id)
}
func (s *InvoicesService) DeleteDraft(ctx context.Context, id string) error {
	return s.repository.DeleteDraftInvoice(ctx, id)
}
func invoiceView(i domain.Invoice) InvoiceView {
	v := InvoiceView{ID: i.ID, InvoiceNumber: i.InvoiceNumber, CustomerID: i.CustomerID, CustomerName: i.CustomerNameSnapshot, CustomerPhone: i.CustomerPhoneSnapshot, OrderID: i.OrderID, IssueDate: i.IssueDate.UTC().Format(time.RFC3339Nano), Status: i.Status, Notes: i.Notes, AccountingJournalEntryID: i.AccountingJournalEntryID, COGSJournalEntryID: i.COGSJournalEntryID, CreatedAt: i.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: i.UpdatedAt.UTC().Format(time.RFC3339Nano), SubtotalRial: i.SubtotalRial, DiscountRial: i.DiscountRial, TotalRial: i.TotalRial, PaidRial: i.PaidRial, RemainingRial: i.RemainingRial}
	if i.DueDate != nil {
		v.DueDate = i.DueDate.UTC().Format(time.RFC3339Nano)
	}
	for _, x := range i.Items {
		v.Items = append(v.Items, InvoiceItemView{ID: x.ID, OrderItemID: x.OrderItemID, Description: x.DescriptionSnapshot, ServiceID: x.ServiceID, QuantityUnit: x.QuantityUnit, Notes: x.Notes, Position: x.Position, Quantity: fmt.Sprint(x.QuantityUnits), UnitPriceRial: x.UnitPriceRial, LineTotalRial: x.LineTotalRial})
	}
	return v
}
