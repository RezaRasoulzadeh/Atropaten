package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestOrderViewsKeepPaymentsAcrossInvoiceLifecycle(t *testing.T) {
	s, order, _ := productionFlowFixture(t)
	ctx := context.Background()
	service := application.NewOrdersService(s, s, nil)
	var invoiceID, invoiceState string
	assertViews := func(paid int64, status domain.PaymentStatus) {
		t.Helper()
		get, err := service.Get(ctx, order.ID)
		if err != nil {
			t.Fatal(err)
		}
		list, err := service.List(ctx)
		if err != nil || len(list) != 1 {
			t.Fatalf("list=%+v error=%v", list, err)
		}
		for _, view := range []application.OrderView{get, list[0]} {
			if view.PaidRial != paid || view.RemainingRial != order.TotalRial-paid || view.PaymentStatus != string(status) {
				t.Fatalf("paid=%d remaining=%d status=%s, want %d/%d/%s", view.PaidRial, view.RemainingRial, view.PaymentStatus, paid, order.TotalRial-paid, status)
			}
		}
		if invoiceID != "" {
			invoice, err := s.GetInvoice(ctx, invoiceID)
			if err != nil {
				t.Fatal(err)
			}
			invoices, err := s.ListInvoices(ctx)
			if err != nil || len(invoices) != 1 {
				t.Fatalf("invoices=%+v err=%v", invoices, err)
			}
			wantPaid, wantStatus := paid, string(status)
			if wantStatus == string(domain.PaymentUnpaid) {
				wantStatus = domain.InvoicePosted
			}
			if invoiceState == domain.InvoiceDraft || invoiceState == domain.InvoiceVoided {
				wantStatus = invoiceState
			}
			if invoiceState == domain.InvoiceVoided {
				wantPaid = 0
			}
			for _, inv := range []domain.Invoice{invoice, invoices[0]} {
				if inv.PaidRial != wantPaid || inv.RemainingRial != order.TotalRial-wantPaid || inv.Status != wantStatus {
					t.Fatalf("invoice paid=%d remaining=%d status=%s; want %d/%d/%s", inv.PaidRial, inv.RemainingRial, inv.Status, wantPaid, order.TotalRial-wantPaid, wantStatus)
				}
			}
			_, summaryStatus, _, summaryPaid, summaryRemaining, err := s.OrderInvoiceSummary(ctx, order.ID)
			if err != nil || summaryPaid != wantPaid || summaryRemaining != order.TotalRial-wantPaid || summaryStatus != wantStatus {
				t.Fatalf("invoice summary mismatch: %d/%d/%s %v", summaryPaid, summaryRemaining, summaryStatus, err)
			}
		}
	}
	pay := func(id string, allocations ...domain.PaymentAllocation) {
		t.Helper()
		var amount int64
		for _, a := range allocations {
			amount += a.AmountRial
		}
		now := time.Now().UTC()
		_, err := s.CreatePayment(ctx, domain.Payment{ID: id, Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", AmountRial: amount, PostedAt: now, CreatedAt: now, Allocations: allocations})
		if err != nil {
			t.Fatal(err)
		}
	}
	assertViews(0, domain.PaymentUnpaid)
	pay("PAY-deposit", domain.PaymentAllocation{TargetType: "order", TargetID: order.ID, AmountRial: 2000})
	assertViews(2000, domain.PaymentPartiallyPaid)
	invoice, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	invoiceID, invoiceState = invoice.ID, domain.InvoiceDraft
	assertViews(2000, domain.PaymentPartiallyPaid) // Draft must not erase deposits.
	if err = s.PostInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	invoiceState = domain.InvoicePosted
	assertViews(2000, domain.PaymentPartiallyPaid) // Previously overwritten with zero.
	pay("PAY-mixed",
		domain.PaymentAllocation{TargetType: "order", TargetID: order.ID, AmountRial: 1000},
		domain.PaymentAllocation{TargetType: "invoice", TargetID: invoice.ID, AmountRial: 3000},
	)
	assertViews(6000, domain.PaymentPartiallyPaid) // Count each allocation, not payment twice.
	pay("PAY-balance", domain.PaymentAllocation{TargetType: "invoice", TargetID: invoice.ID, AmountRial: 4000})
	assertViews(10000, domain.PaymentPaid)
	if _, err = s.ReversePayment(ctx, "PAY-mixed", ""); err != nil {
		t.Fatal(err)
	}
	assertViews(6000, domain.PaymentPartiallyPaid)
	if _, err = s.ReversePayment(ctx, "PAY-balance", ""); err != nil {
		t.Fatal(err)
	}
	assertViews(2000, domain.PaymentPartiallyPaid)
	if err = s.VoidInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	invoiceState = domain.InvoiceVoided
	assertViews(2000, domain.PaymentPartiallyPaid) // Voiding invoice doesn't reverse the deposit.
	if _, err = s.ReversePayment(ctx, "PAY-deposit", ""); err != nil {
		t.Fatal(err)
	}
	assertViews(0, domain.PaymentUnpaid)
}

func TestLinkedOrderInvoiceCannotBePaidTwice(t *testing.T) {
	for _, firstTarget := range []string{"order", "invoice"} {
		t.Run(firstTarget, func(t *testing.T) {
			s, order, _ := productionFlowFixture(t)
			ctx := context.Background()
			invoice, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, order.ID)
			if err != nil {
				t.Fatal(err)
			}
			if err = s.PostInvoice(ctx, invoice.ID); err != nil {
				t.Fatal(err)
			}
			now := time.Now().UTC()
			targets := map[string]string{"order": order.ID, "invoice": invoice.ID}
			payment := domain.Payment{ID: "PAY-first", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", AmountRial: 6000, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{TargetType: firstTarget, TargetID: targets[firstTarget], AmountRial: 6000}}}
			if _, err = s.CreatePayment(ctx, payment); err != nil {
				t.Fatal(err)
			}
			other := "order"
			if firstTarget == "order" {
				other = "invoice"
			}
			payment.ID, payment.AmountRial = "PAY-overpay", 5000
			payment.Allocations = []domain.PaymentAllocation{{TargetType: other, TargetID: targets[other], AmountRial: 5000}}
			if _, err = s.CreatePayment(ctx, payment); !errors.Is(err, domain.ErrAllocationExceeded) {
				t.Fatalf("cross-target overpayment allowed: %v", err)
			}
			payment.ID = "PAY-mixed-overpay"
			payment.Allocations = []domain.PaymentAllocation{{TargetType: firstTarget, TargetID: targets[firstTarget], AmountRial: 2500}, {TargetType: other, TargetID: targets[other], AmountRial: 2500}}
			if _, err = s.CreatePayment(ctx, payment); !errors.Is(err, domain.ErrAllocationExceeded) {
				t.Fatalf("mixed overpayment allowed: %v", err)
			}
			paid, _, _, err := s.OrderPaymentSummary(ctx, order.ID)
			if err != nil || paid != 6000 {
				t.Fatalf("failed payment changed totals: %d %v", paid, err)
			}
		})
	}
}

func TestOrderPaymentSummaryExcludesOtherOrdersAndUnallocatedMoney(t *testing.T) {
	s, order, _ := productionFlowFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()
	other := domain.NewOrder("ORD-other-payment", "", now)
	other.Items = []domain.OrderItem{{ID: "ITEM-other-payment", OrderID: other.ID, ServiceNameSnapshot: "Other work", Quantity: domain.QuantityScale, SellingPriceRial: 5000}}
	if err := other.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateOrder(ctx, other); err != nil {
		t.Fatal(err)
	}
	_, err := s.CreatePayment(ctx, domain.Payment{ID: "PAY-multiple-orders", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", AmountRial: 6000, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{
		{TargetType: "order", TargetID: order.ID, AmountRial: 1500},
		{TargetType: "order", TargetID: other.ID, AmountRial: 2500},
	}})
	if err != nil {
		t.Fatal(err)
	}
	paid, remaining, _, err := s.OrderPaymentSummary(ctx, order.ID)
	if err != nil || paid != 1500 || remaining != 8500 {
		t.Fatalf("paid=%d remaining=%d err=%v", paid, remaining, err)
	}
}
