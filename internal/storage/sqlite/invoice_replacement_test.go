package sqlite

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestVoidedInvoiceAllowsReplacementAndPreservesHistory(t *testing.T) {
	s, order, _ := productionFlowFixture(t)
	ctx := context.Background()
	invoices := application.NewInvoicesService(s, s)
	originalView, err := invoices.CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, originalView.ID); err != nil {
		t.Fatal(err)
	}
	original, err := s.GetInvoice(ctx, originalView.ID)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := s.GetJournalEntry(ctx, original.AccountingJournalEntryID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.VoidInvoice(ctx, original.ID); err != nil {
		t.Fatal(err)
	}
	voided, err := s.GetInvoice(ctx, original.ID)
	if err != nil || voided.Status != domain.InvoiceVoided {
		t.Fatalf("voided=%+v err=%v", voided, err)
	}
	if !reflect.DeepEqual(voided.Items, original.Items) || voided.TotalRial != original.TotalRial || voided.AccountingJournalEntryID != original.AccountingJournalEntryID {
		t.Fatal("void changed historical invoice")
	}
	journalAfter, err := s.GetJournalEntry(ctx, journal.ID)
	if err != nil || !reflect.DeepEqual(journalAfter, journal) {
		t.Fatal("original journal changed")
	}

	order.Items[0].Quantity = 15 * domain.QuantityScale
	if err = order.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	replacementView, err := invoices.CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if replacementView.ID == original.ID || replacementView.Status != domain.InvoiceDraft || replacementView.TotalRial != order.TotalRial || replacementView.Items[0].Quantity != fmt.Sprint(15*domain.QuantityScale) {
		t.Fatalf("replacement=%+v", replacementView)
	}
	replacement, err := s.GetInvoice(ctx, replacementView.ID)
	if err != nil {
		t.Fatal(err)
	}
	if replacement.PreviousInvoiceID != original.ID || replacement.ReplacementInvoiceID != "" {
		t.Fatalf("replacement linkage=%+v", replacement)
	}
	originalAfter, err := s.GetInvoice(ctx, original.ID)
	if err != nil || originalAfter.ReplacementInvoiceID != replacement.ID {
		t.Fatalf("original linkage=%+v err=%v", originalAfter, err)
	}
	if _, err = invoices.CreateFromOrder(ctx, order.ID); err != domain.ErrInvoiceOrderExists {
		t.Fatalf("duplicate replacement error=%v", err)
	}
	if err = s.PostInvoice(ctx, replacement.ID); err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, replacement.ID); err != nil {
		t.Fatal(err)
	}
	if countRows(t, s, `SELECT COUNT(*) FROM journal_entries WHERE source_type='invoice' AND reversal_of_id IS NULL`) != 2 {
		t.Fatal("replacement posting duplicated revenue journal")
	}
	if countRows(t, s, `SELECT COUNT(*) FROM journal_entries WHERE source_type='invoice_cogs'`) > 1 {
		t.Fatal("replacement posting duplicated COGS journal")
	}
}

func TestVoidingInvoiceReleasesAllocationToCustomerCredit(t *testing.T) {
	s, order, _ := productionFlowFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()
	customer := domain.Customer{ID: "CUS-void-credit", Name: "Void Credit", Active: true, CreatedAt: now, UpdatedAt: now}
	if err := s.SaveCustomer(ctx, customer); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE orders SET customer_id=?,customer_name_snapshot=? WHERE id=?`, customer.ID, customer.Name, order.ID); err != nil {
		t.Fatal(err)
	}
	order.CustomerID, order.CustomerNameSnapshot = customer.ID, customer.Name
	invoice, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	payment := domain.Payment{ID: "PAY-void-credit", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", CustomerID: customer.ID, AmountRial: order.TotalRial, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{TargetType: "invoice", TargetID: invoice.ID, AmountRial: order.TotalRial}}}
	if _, err = s.CreatePayment(ctx, payment); err != nil {
		t.Fatal(err)
	}
	if err = s.VoidInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	var reversed int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM payment_allocations WHERE payment_id=? AND reversed=1`, payment.ID).Scan(&reversed); err != nil || reversed != 1 {
		t.Fatalf("released allocations=%d err=%v", reversed, err)
	}
	_, credit, err := s.CustomerFinancialSummary(ctx, customer.ID)
	if err != nil || credit != order.TotalRial {
		t.Fatalf("credit=%d err=%v", credit, err)
	}
	if _, err = s.GetInvoiceForOrder(ctx, order.ID); err != domain.ErrInvoiceNotFound {
		t.Fatalf("voided invoice still blocks creation: %v", err)
	}
	if _, err = application.NewInvoicesService(s, s).CreateFromOrder(ctx, order.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReversePayment(ctx, payment.ID, "void-credit-reversal"); err != nil {
		t.Fatal(err)
	}
	_, credit, err = s.CustomerFinancialSummary(ctx, customer.ID)
	if err != nil || credit != 0 {
		t.Fatalf("credit after payment reversal=%d err=%v", credit, err)
	}
}

func TestDraftInvoiceVoidAllowsFreshDraftWithoutAccounting(t *testing.T) {
	s, order, _ := productionFlowFixture(t)
	ctx := context.Background()
	service := application.NewInvoicesService(s, s)
	draft, err := service.CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.VoidInvoice(ctx, draft.ID); err != nil {
		t.Fatal(err)
	}
	if countRows(t, s, `SELECT COUNT(*) FROM journal_entries`) != 0 {
		t.Fatal("draft void posted accounting")
	}
	fresh, err := service.CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.ID == draft.ID || fresh.Status != domain.InvoiceDraft {
		t.Fatalf("fresh=%+v", fresh)
	}
}

func TestSequentialInvoiceCorrectionsRemainTraceable(t *testing.T) {
	s, order, _ := productionFlowFixture(t)
	ctx := context.Background()
	service := application.NewInvoicesService(s, s)
	created, err := service.CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{created.ID}
	for n := 0; n < 3; n++ {
		if err = s.PostInvoice(ctx, ids[len(ids)-1]); err != nil {
			t.Fatal(err)
		}
		if err = s.VoidInvoice(ctx, ids[len(ids)-1]); err != nil {
			t.Fatal(err)
		}
		next, e := service.CreateFromOrder(ctx, order.ID)
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, next.ID)
	}
	for i := 1; i < len(ids); i++ {
		current, err := s.GetInvoice(ctx, ids[i])
		if err != nil || current.PreviousInvoiceID != ids[i-1] {
			t.Fatalf("invoice %d=%+v err=%v", i, current, err)
		}
		previous, err := s.GetInvoice(ctx, ids[i-1])
		if err != nil || previous.ReplacementInvoiceID != ids[i] {
			t.Fatalf("previous %d=%+v err=%v", i-1, previous, err)
		}
	}
}
