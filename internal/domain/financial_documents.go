package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var (
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvoiceNotDraft      = errors.New("only draft invoices can be edited or deleted")
	ErrInvoiceAlreadyPosted = errors.New("invoice is already posted")
	ErrInvoiceCannotVoid    = errors.New("only posted invoices without active allocations can be voided")
	ErrInvoiceProtected     = errors.New("posted invoice history cannot be deleted")
	ErrInvoiceOrderExists   = errors.New("this order already has an invoice")
	ErrExpenseNotFound      = errors.New("expense not found")
	ErrExpenseProtected     = errors.New("posted expense history cannot be deleted")
	ErrTransferNotFound     = errors.New("transfer not found")
	ErrTransferProtected    = errors.New("posted transfer history cannot be deleted")
	ErrTransferAccount      = errors.New("transfer source and destination must be different active accounts")
)

const (
	InvoiceDraft         = "Draft"
	InvoicePosted        = "Posted"
	InvoicePartiallyPaid = "Partially Paid"
	InvoicePaid          = "Paid"
	InvoiceVoided        = "Voided"
)

type Invoice struct {
	ID, InvoiceNumber, CustomerID, CustomerNameSnapshot, CustomerPhoneSnapshot, OrderID string
	IssueDate                                                                           time.Time
	DueDate                                                                             *time.Time
	Status, Notes                                                                       string
	SubtotalRial, DiscountRial, TotalRial                                               int64
	PaidRial, RemainingRial                                                             int64
	AccountingJournalEntryID, COGSJournalEntryID                                        string
	CreatedAt, UpdatedAt                                                                time.Time
	Items                                                                               []InvoiceItem
}

type InvoiceItem struct {
	ID, InvoiceID, OrderItemID, DescriptionSnapshot, ServiceID, QuantityUnit, Notes string
	Position                                                                        int
	QuantityUnits                                                                   int64
	UnitPriceRial, LineTotalRial                                                    int64
}

func (i Invoice) Validate() error {
	if i.ID == "" || i.CustomerNameSnapshot == "" || i.IssueDate.IsZero() || i.CreatedAt.IsZero() || i.UpdatedAt.IsZero() {
		return fmt.Errorf("invoice identity, issue date, customer snapshot, and timestamps are required")
	}
	if i.Status != InvoiceDraft && i.InvoiceNumber == "" {
		return fmt.Errorf("posted invoice number is required")
	}
	if i.Status != InvoiceDraft && i.Status != InvoicePosted && i.Status != InvoicePartiallyPaid && i.Status != InvoicePaid && i.Status != InvoiceVoided {
		return fmt.Errorf("unsupported invoice status")
	}
	if i.SubtotalRial < 0 || i.DiscountRial < 0 || i.TotalRial < 0 || i.DiscountRial > i.SubtotalRial || i.TotalRial != i.SubtotalRial-i.DiscountRial {
		return fmt.Errorf("invalid invoice totals")
	}
	var total big.Int
	for n, item := range i.Items {
		if item.ID == "" || item.InvoiceID != i.ID || item.Position != n || item.DescriptionSnapshot == "" || item.QuantityUnits < 0 || item.UnitPriceRial < 0 || item.LineTotalRial < 0 {
			return fmt.Errorf("invalid invoice item %d", n)
		}
		total.Add(&total, big.NewInt(item.LineTotalRial))
	}
	if !total.IsInt64() || total.Int64() != i.SubtotalRial {
		return fmt.Errorf("invoice subtotal does not equal line totals")
	}
	return nil
}

type Expense struct {
	ID, ExpenseNumber                                                                                                                   string
	ExpenseDate                                                                                                                         time.Time
	CategoryAccountID, Payee, SupplierID, Description, PaymentMethod, FinancialAccountID, Notes, Status, JournalEntryID, IdempotencyKey string
	AmountRial                                                                                                                          int64
	CreatedAt, UpdatedAt                                                                                                                time.Time
}

func (e Expense) Validate() error {
	if e.ID == "" || e.CategoryAccountID == "" || e.FinancialAccountID == "" || e.Description == "" || e.AmountRial <= 0 || e.Status != "Posted" {
		return fmt.Errorf("expense requires account, financial account, description, positive amount, and Posted status")
	}
	return nil
}

type FinancialTransfer struct {
	ID, TransferNumber, SourceFinancialAccountID, DestinationFinancialAccountID, Reference, Notes, Status, JournalEntryID, IdempotencyKey string
	AmountRial                                                                                                                            int64
	TransferDate                                                                                                                          time.Time
	CreatedAt, UpdatedAt                                                                                                                  time.Time
}

func (t FinancialTransfer) Validate() error {
	if t.ID == "" || t.SourceFinancialAccountID == "" || t.DestinationFinancialAccountID == "" || t.SourceFinancialAccountID == t.DestinationFinancialAccountID || t.AmountRial <= 0 || t.TransferDate.IsZero() || t.Status != "Posted" {
		return ErrTransferAccount
	}
	return nil
}

func NormalizeFinancialText(v string) string { return strings.TrimSpace(v) }
