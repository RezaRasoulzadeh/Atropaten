package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestInvoiceCopiesOrderSnapshotsPostsOnceAndVoidReverses(t *testing.T) {
	s := openFinanceTestStore(t)
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)
	c := domain.Customer{ID: "CUS-invoice", Name: "Snapshot Customer", Phone: "0912", Active: true, CreatedAt: now, UpdatedAt: now}
	if err := s.SaveCustomer(ctx, c); err != nil {
		t.Fatal(err)
	}
	o := domain.NewOrder("ORD-invoice", c.ID, now)
	o.OrderNumber = "ORD-1001"
	o.CustomerNameSnapshot, o.CustomerPhoneSnapshot = c.Name, c.Phone
	o.Items = []domain.OrderItem{
		{ID: "ITEM-invoice-1", OrderID: o.ID, Position: 0, ServiceNameSnapshot: "Cards snapshot", Quantity: domain.QuantityScale, QuantityUnit: "piece", PricingSnapshotJSON: `{"price":1200}`, SellingPriceRial: 1200},
		{ID: "ITEM-invoice-2", OrderID: o.ID, Position: 1, ServiceNameSnapshot: "Flyers snapshot", Quantity: domain.QuantityScale, QuantityUnit: "piece", PricingSnapshotJSON: `{"price":800}`, SellingPriceRial: 800},
	}
	if err := o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	service := application.NewInvoicesService(s, s)
	created, err := service.CreateFromOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if created.TotalRial != 2000 || len(created.Items) != 2 || created.Items[1].Position != 1 || created.Items[0].LineTotalRial != 1200 {
		t.Fatalf("invoice snapshot = %+v", created)
	}
	if _, err = service.CreateFromOrder(ctx, o.ID); err != domain.ErrInvoiceOrderExists {
		t.Fatalf("duplicate invoice error=%v", err)
	}
	if err = s.PostInvoice(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, created.ID); err != nil {
		t.Fatal("idempotent post: ", err)
	}
	entries, err := s.ListJournalEntries(ctx)
	if err != nil || len(entries) != 1 {
		t.Fatalf("invoice entries=%d err=%v", len(entries), err)
	}
	got, err := s.GetInvoice(ctx, created.ID)
	if err != nil || got.Status != domain.InvoicePosted || got.AccountingJournalEntryID == "" {
		t.Fatalf("posted invoice=%+v err=%v", got, err)
	}
	if err = s.VoidInvoice(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if err = s.VoidInvoice(ctx, created.ID); err != nil {
		t.Fatal("idempotent void: ", err)
	}
	entries, err = s.ListJournalEntries(ctx)
	if err != nil || len(entries) != 2 {
		t.Fatalf("reversal entries=%d err=%v", len(entries), err)
	}
	got, err = s.GetInvoice(ctx, created.ID)
	if err != nil || got.Status != domain.InvoiceVoided {
		t.Fatalf("voided invoice=%+v err=%v", got, err)
	}
}

func TestInvoicePaymentsAllocateAcrossInvoicesAndReverse(t *testing.T) {
	s := openFinanceTestStore(t)
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 2, 2, 8, 0, 0, 0, time.UTC)
	c := domain.Customer{ID: "CUS-multi-invoice", Name: "Multi Invoice Customer", Active: true, CreatedAt: now, UpdatedAt: now}
	if err := s.SaveCustomer(ctx, c); err != nil {
		t.Fatal(err)
	}
	ids := []string{"ORD-invoice-a", "ORD-invoice-b"}
	for n, id := range ids {
		o := domain.NewOrder(id, c.ID, now)
		o.OrderNumber = "ORD-20" + string(rune('0'+n))
		o.CustomerNameSnapshot = c.Name
		o.Items = []domain.OrderItem{{ID: id + "-ITEM", OrderID: id, Position: 0, ServiceNameSnapshot: "Service", Quantity: domain.QuantityScale, QuantityUnit: "unit", SellingPriceRial: 1000}}
		if err := o.RecalculateTotals(); err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrder(ctx, o); err != nil {
			t.Fatal(err)
		}
		if _, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, id); err != nil {
			t.Fatal(err)
		}
		inv, err := s.GetInvoiceForOrder(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if err = s.PostInvoice(ctx, inv.ID); err != nil {
			t.Fatal(err)
		}
	}
	a, _ := s.GetInvoiceForOrder(ctx, ids[0])
	b, _ := s.GetInvoiceForOrder(ctx, ids[1])
	p := domain.Payment{ID: "PAY-invoices", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", CustomerID: c.ID, AmountRial: 1500, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{ID: "PAY-invoices-A", TargetType: "invoice", TargetID: a.ID, AmountRial: 1000}, {ID: "PAY-invoices-B", TargetType: "invoice", TargetID: b.ID, AmountRial: 500}}}
	if _, err := s.CreatePayment(ctx, p); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreatePayment(ctx, p); err != nil {
		t.Fatal("idempotent payment: ", err)
	}
	a, _ = s.GetInvoice(ctx, a.ID)
	b, _ = s.GetInvoice(ctx, b.ID)
	if a.Status != domain.InvoicePaid || a.PaidRial != 1000 || b.Status != domain.InvoicePartiallyPaid || b.PaidRial != 500 || b.RemainingRial != 500 {
		t.Fatalf("invoice balances a=%+v b=%+v", a, b)
	}
	if _, err := s.ReversePayment(ctx, p.ID, ""); err != nil {
		t.Fatal(err)
	}
	a, _ = s.GetInvoice(ctx, a.ID)
	b, _ = s.GetInvoice(ctx, b.ID)
	if a.PaidRial != 0 || b.PaidRial != 0 || a.RemainingRial != 1000 || b.RemainingRial != 1000 {
		t.Fatalf("reversed allocation a=%+v b=%+v", a, b)
	}
}

func TestInvoicePostingRecognizesActualProductionMovementCostOnce(t *testing.T) {
	s := openFinanceTestStore(t)
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 2, 4, 8, 0, 0, 0, time.UTC)
	c := domain.Customer{ID: "CUS-cogs", Name: "COGS Customer", Active: true, CreatedAt: now, UpdatedAt: now}
	if err := s.SaveCustomer(ctx, c); err != nil {
		t.Fatal(err)
	}
	m, err := domain.NewMaterial("MAT-cogs", domain.MaterialDraft{Name: "Paper", PurchaseUnit: "pack", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale, PhysicalStock: 10 * domain.QuantityScale, AverageUnitCostRial: 100}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	o := domain.NewOrder("ORD-cogs", c.ID, now)
	o.OrderNumber, o.CustomerNameSnapshot, o.CommercialStatus = "ORD-COGS", c.Name, domain.CommercialConfirmed
	o.Items = []domain.OrderItem{{ID: "ITEM-cogs", OrderID: o.ID, Position: 0, ServiceNameSnapshot: "Printed item", Quantity: domain.QuantityScale, QuantityUnit: "unit", SellingPriceRial: 1000}}
	if err = o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.CreateOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	job := domain.ProductionJob{ID: "JOB-cogs", OrderID: o.ID, OrderItemID: o.Items[0].ID, Quantity: domain.QuantityScale, QuantityUnit: "unit", Status: domain.ProductionPending, Priority: string(domain.PriorityNormal), CreatedAt: now, UpdatedAt: now}
	if err = s.CreateProductionJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, job.ID, domain.ProductionInProgress); err != nil {
		t.Fatal(err)
	}
	if err = s.CreateReservation(ctx, domain.InventoryReservation{ID: "RES-cogs", MaterialID: m.ID, OrderID: o.ID, OrderItemID: o.Items[0].ID, ProductionJobID: job.ID, Quantity: 2 * domain.QuantityScale, Status: domain.ReservationActive}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.RecordProductionConsumption(ctx, job.ID, m.ID, "consume-once", domain.QuantityScale, domain.QuantityScale, "actual usage and waste"); err != nil {
		t.Fatal(err)
	}
	service := application.NewInvoicesService(s, s)
	inv, err := service.CreateFromOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, inv.ID); err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, inv.ID); err != nil {
		t.Fatal(err)
	}
	entries, err := s.ListJournalEntries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("journal count=%d", len(entries))
	}
	var cogs int64
	for _, entry := range entries {
		if entry.SourceType == "invoice_cogs" {
			for _, line := range entry.Lines {
				cogs += line.DebitRial
			}
		}
	}
	if cogs != 200 {
		t.Fatalf("actual COGS=%d, want 200", cogs)
	}
}

func TestExpenseAndTransferAreBalancedIdempotentAndReversible(t *testing.T) {
	s := openFinanceTestStore(t)
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 2, 3, 8, 0, 0, 0, time.UTC)
	exp := domain.Expense{ID: "EXP-test", ExpenseDate: now, CategoryAccountID: "ACC-EXP-OTHER", Description: "Office supplies", PaymentMethod: string(domain.PaymentCash), FinancialAccountID: "FIN-CASH", AmountRial: 123456789, Status: "Posted", CreatedAt: now, UpdatedAt: now}
	if _, err := s.CreateExpense(ctx, exp); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateExpense(ctx, exp); err != nil {
		t.Fatal("idempotent expense: ", err)
	}
	tr := domain.FinancialTransfer{ID: "TRF-test", SourceFinancialAccountID: "FIN-CASH", DestinationFinancialAccountID: "FIN-BANK", AmountRial: 987654321, TransferDate: now, Reference: "Deposit", Status: "Posted", CreatedAt: now, UpdatedAt: now}
	if _, err := s.CreateTransfer(ctx, tr); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTransfer(ctx, tr); err != nil {
		t.Fatal("idempotent transfer: ", err)
	}
	if _, err := s.ReverseExpense(ctx, exp.ID, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReverseTransfer(ctx, tr.ID, ""); err != nil {
		t.Fatal(err)
	}
	entries, err := s.ListJournalEntries(ctx)
	if err != nil || len(entries) != 4 {
		t.Fatalf("entries=%d err=%v", len(entries), err)
	}
	for _, entry := range entries {
		var debit, credit int64
		for _, line := range entry.Lines {
			debit += line.DebitRial
			credit += line.CreditRial
		}
		if debit != credit {
			t.Fatalf("unbalanced entry %s: debit=%d credit=%d", entry.ID, debit, credit)
		}
	}
}

func openFinanceTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "finance.db"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}
