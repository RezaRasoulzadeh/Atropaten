package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestAccountingPostingAllocationAndReversalAreIdempotent(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "accounting.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 1, 2, 8, 0, 0, 0, time.UTC)
	c, _ := domain.NewCustomer("CUS-pay", domain.CustomerDraft{Name: "Paying customer"}, now)
	if err = s.SaveCustomer(ctx, c); err != nil {
		t.Fatal(err)
	}
	o := domain.NewOrder("ORD-pay", c.ID, now)
	o.Items = []domain.OrderItem{{ID: "ITEM-pay", OrderID: o.ID, Position: 0, ServiceNameSnapshot: "Work", Quantity: domain.QuantityScale, QuantityUnit: "unit", ResolvedParametersJSON: "{}", CostBreakdownJSON: "[]", PricingSnapshotJSON: "{}", SellingPriceRial: 1000}}
	if err = o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.CreateOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	p := domain.Payment{ID: "PAY-pay", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", CustomerID: c.ID, AmountRial: 400, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{ID: "AL-pay", TargetType: "order", TargetID: o.ID, AmountRial: 400}}}
	if _, err = s.CreatePayment(ctx, p); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreatePayment(ctx, p); err != nil {
		t.Fatal("idempotent retry:", err)
	}
	paid, remaining, status, err := s.OrderPaymentSummary(ctx, o.ID)
	if err != nil || paid != 400 || remaining != 600 || status != domain.PaymentPartiallyPaid {
		t.Fatalf("summary=%d,%d,%s,%v", paid, remaining, status, err)
	}
	entries, err := s.ListJournalEntries(ctx)
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries=%d err=%v", len(entries), err)
	}
	if _, err = s.ReversePayment(ctx, p.ID, ""); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReversePayment(ctx, p.ID, ""); err != nil {
		t.Fatalf("idempotent second reverse=%v", err)
	}
	paid, remaining, status, err = s.OrderPaymentSummary(ctx, o.ID)
	if err != nil || paid != 0 || remaining != 1000 || status != domain.PaymentUnpaid {
		t.Fatalf("reversed summary=%d,%d,%s,%v", paid, remaining, status, err)
	}
	entries, err = s.ListJournalEntries(ctx)
	if err != nil || len(entries) != 2 {
		t.Fatalf("reversal entries=%d err=%v", len(entries), err)
	}
	var balance int64
	for _, a := range mustAccounts(t, s, ctx) {
		if a.Code == "1000" {
			balance = a.BalanceRial
		}
	}
	if balance != 0 {
		t.Fatalf("cash derived balance=%d", balance)
	}
}

func TestSupplierPaymentReducesPayableWithoutMutatingPurchaseHistory(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "supplier-payment.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 1, 3, 8, 0, 0, 0, time.UTC)
	supplier := domain.Supplier{ID: "SUP-pay", Name: "Supplier", Active: true, CreatedAt: now, UpdatedAt: now}
	if err = s.SaveSupplier(ctx, supplier); err != nil {
		t.Fatal(err)
	}
	m, err := domain.NewMaterial("MAT-pay", domain.MaterialDraft{Name: "Material", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.QuantityScale}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	p := domain.Purchase{ID: "PUR-pay", SupplierID: supplier.ID, SupplierNameSnapshot: supplier.Name, PurchaseDate: now, Status: domain.PurchaseDraft, CreatedAt: now, UpdatedAt: now, Items: []domain.PurchaseItem{{ID: "PITM-pay", MaterialID: m.ID, MaterialNameSnapshot: m.Name, PurchaseUnitSnapshot: m.PurchaseUnit, ConsumptionUnitSnapshot: m.ConsumptionUnit, PurchaseQuantity: domain.QuantityScale, ConversionFactorSnapshot: domain.QuantityScale, ConsumptionQuantity: domain.QuantityScale, UnitAcquisitionCostRial: 100, LineTotalRial: 100}}, SubtotalRial: 100, TotalRial: 100}
	if err = s.SavePurchase(ctx, p); err != nil {
		t.Fatal(err)
	}
	if err = s.PostPurchase(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	payment := domain.Payment{ID: "PAY-supplier", Direction: domain.PaymentOutgoing, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", SupplierID: supplier.ID, AmountRial: 40, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{TargetType: "purchase", TargetID: p.ID, AmountRial: 40}}}
	if _, err = s.CreatePayment(ctx, payment); err != nil {
		t.Fatal(err)
	}
	paid, remaining, err := s.PurchasePaymentSummary(ctx, p.ID)
	if err != nil || paid != 40 || remaining != 60 {
		t.Fatalf("payable summary=%d,%d err=%v", paid, remaining, err)
	}
	ap, err := s.GetAccount(ctx, "ACC-AP")
	if err != nil || ap.BalanceRial != 60 {
		t.Fatalf("AP balance=%d err=%v", ap.BalanceRial, err)
	}
	got, err := s.GetPurchase(ctx, p.ID)
	if err != nil || got.Status != domain.PurchasePosted {
		t.Fatalf("purchase history changed: %+v err=%v", got, err)
	}
}

func TestIncomingPaymentDerivesCustomerPartyFromAllocation(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "payment-party.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 1, 4, 8, 0, 0, 0, time.UTC)
	c, _ := domain.NewCustomer("CUS-derived-party", domain.CustomerDraft{Name: "Derived party"}, now)
	if err = s.SaveCustomer(ctx, c); err != nil {
		t.Fatal(err)
	}
	o := domain.NewOrder("ORD-derived-party", c.ID, now)
	o.Items = []domain.OrderItem{{ID: "ITEM-derived-party", OrderID: o.ID, Position: 0, ServiceNameSnapshot: "Work", Quantity: domain.QuantityScale, QuantityUnit: "unit", SellingPriceRial: 1000}}
	if err = o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.CreateOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	created, err := s.CreatePayment(ctx, domain.Payment{ID: "PAY-derived-party", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", AmountRial: 1000, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{TargetType: "order", TargetID: o.ID, AmountRial: 1000}}})
	if err != nil {
		t.Fatal(err)
	}
	if created.CustomerID != c.ID {
		t.Fatalf("payment customer=%q, want %q", created.CustomerID, c.ID)
	}
	var party string
	if err = s.db.QueryRow(`SELECT party_id FROM journal_lines WHERE journal_entry_id=? AND account_id='ACC-AR'`, "JE-PAY-"+created.ID).Scan(&party); err != nil {
		t.Fatal(err)
	}
	if party != c.ID {
		t.Fatalf("AR party=%q, want %q", party, c.ID)
	}
}

func mustAccounts(t *testing.T, s *Store, ctx context.Context) []domain.Account {
	t.Helper()
	v, e := s.ListAccounts(ctx)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
