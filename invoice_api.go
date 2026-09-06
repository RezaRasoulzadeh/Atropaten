package main

import (
	"Atropaten/internal/application"
	"fmt"
)

type InvoiceDTO struct {
	ID                       string           `json:"id"`
	InvoiceNumber            string           `json:"invoiceNumber"`
	CustomerID               string           `json:"customerId"`
	CustomerName             string           `json:"customerName"`
	CustomerPhone            string           `json:"customerPhone"`
	OrderID                  string           `json:"orderId"`
	IssueDate                string           `json:"issueDate"`
	DueDate                  string           `json:"dueDate"`
	Status                   string           `json:"status"`
	Notes                    string           `json:"notes"`
	SubtotalRial             int64            `json:"subtotalRial"`
	DiscountRial             int64            `json:"discountRial"`
	TotalRial                int64            `json:"totalRial"`
	PaidRial                 int64            `json:"paidRial"`
	RemainingRial            int64            `json:"remainingRial"`
	AccountingJournalEntryID string           `json:"accountingJournalEntryId"`
	COGSJournalEntryID       string           `json:"cogsJournalEntryId"`
	CreatedAt                string           `json:"createdAt"`
	UpdatedAt                string           `json:"updatedAt"`
	Items                    []InvoiceItemDTO `json:"items"`
}
type InvoiceItemDTO struct {
	ID            string `json:"id"`
	OrderItemID   string `json:"orderItemId"`
	Description   string `json:"description"`
	ServiceID     string `json:"serviceId"`
	QuantityUnit  string `json:"quantityUnit"`
	Notes         string `json:"notes"`
	Position      int    `json:"position"`
	Quantity      string `json:"quantity"`
	UnitPriceRial int64  `json:"unitPriceRial"`
	LineTotalRial int64  `json:"lineTotalRial"`
}

func (a *App) invoiceService() (*application.InvoicesService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.invoices == nil {
		return nil, fmt.Errorf("invoice service is not initialized")
	}
	return a.invoices, nil
}
func (a *App) ListInvoices() ([]InvoiceDTO, error) {
	s, e := a.invoiceService()
	if e != nil {
		return nil, e
	}
	v, e := s.List(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]InvoiceDTO, 0, len(v))
	for _, x := range v {
		out = append(out, invoiceDTO(x))
	}
	return out, nil
}
func (a *App) GetInvoice(id string) (InvoiceDTO, error) {
	s, e := a.invoiceService()
	if e != nil {
		return InvoiceDTO{}, e
	}
	v, e := s.Get(a.materialContext(), id)
	return invoiceDTO(v), e
}
func (a *App) CreateInvoiceFromOrder(orderID string) (InvoiceDTO, error) {
	s, e := a.invoiceService()
	if e != nil {
		return InvoiceDTO{}, e
	}
	v, e := s.CreateFromOrder(a.materialContext(), orderID)
	return invoiceDTO(v), e
}
func (a *App) PostInvoice(id string) (InvoiceDTO, error) {
	s, e := a.invoiceService()
	if e != nil {
		return InvoiceDTO{}, e
	}
	v, e := s.Post(a.materialContext(), id)
	return invoiceDTO(v), e
}
func (a *App) VoidInvoice(id string) (InvoiceDTO, error) {
	s, e := a.invoiceService()
	if e != nil {
		return InvoiceDTO{}, e
	}
	v, e := s.Void(a.materialContext(), id)
	return invoiceDTO(v), e
}
func (a *App) DeleteDraftInvoice(id string) error {
	s, e := a.invoiceService()
	if e != nil {
		return e
	}
	return s.DeleteDraft(a.materialContext(), id)
}
func invoiceDTO(v application.InvoiceView) InvoiceDTO {
	out := InvoiceDTO{ID: v.ID, InvoiceNumber: v.InvoiceNumber, CustomerID: v.CustomerID, CustomerName: v.CustomerName, CustomerPhone: v.CustomerPhone, OrderID: v.OrderID, IssueDate: v.IssueDate, DueDate: v.DueDate, Status: v.Status, Notes: v.Notes, SubtotalRial: v.SubtotalRial, DiscountRial: v.DiscountRial, TotalRial: v.TotalRial, PaidRial: v.PaidRial, RemainingRial: v.RemainingRial, AccountingJournalEntryID: v.AccountingJournalEntryID, COGSJournalEntryID: v.COGSJournalEntryID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
	for _, x := range v.Items {
		out.Items = append(out.Items, InvoiceItemDTO{ID: x.ID, OrderItemID: x.OrderItemID, Description: x.Description, ServiceID: x.ServiceID, QuantityUnit: x.QuantityUnit, Notes: x.Notes, Position: x.Position, Quantity: x.Quantity, UnitPriceRial: x.UnitPriceRial, LineTotalRial: x.LineTotalRial})
	}
	return out
}
