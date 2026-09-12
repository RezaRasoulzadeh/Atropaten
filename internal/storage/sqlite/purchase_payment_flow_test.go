package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestPurchasePayLaterCashAndCheckReconcileBalances(t *testing.T) {
	s := newM5Store(t)
	ctx := context.Background()
	now := time.Now().UTC()
	supplier := domain.Supplier{ID: "SUP-later", Name: "Supplier", Active: true, CreatedAt: now, UpdatedAt: now}
	if err := s.SaveSupplier(ctx, supplier); err != nil {
		t.Fatal(err)
	}
	material, err := domain.NewMaterial("MAT-later", domain.MaterialDraft{Name: "Paper", PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, material); err != nil {
		t.Fatal(err)
	}
	service := application.NewPurchasesService(s, s, s)
	purchase, err := service.Create(ctx, application.PurchaseInput{SupplierID: supplier.ID, PurchaseDate: now.Format("2006-01-02")})
	if err != nil {
		t.Fatalf("pay-later purchase without account: %v", err)
	}
	if purchase.FinancialAccountID != "" {
		t.Fatal("pay-later account defaulted")
	}
	purchase, err = service.AddItem(ctx, purchase.ID, application.PurchaseItemInput{MaterialID: material.ID, PurchaseQuantity: "10", UnitAcquisitionCostRial: "100"})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostPurchase(ctx, purchase.ID); err != nil {
		t.Fatal(err)
	}
	check := func(paid, ap, checksPayable int64) {
		t.Helper()
		view, err := service.Get(ctx, purchase.ID)
		if err != nil || view.PaidRial != paid || view.RemainingRial != 1000-paid {
			t.Fatalf("purchase=%+v err=%v", view, err)
		}
		account, err := s.GetAccount(ctx, "ACC-AP")
		if err != nil || account.BalanceRial != ap {
			t.Fatalf("accounting AP=%+v err=%v want=%d", account, err, ap)
		}
		bankCheck, err := s.GetAccount(ctx, "ACC-CHECKS-PAYABLE")
		if err != nil || bankCheck.BalanceRial != checksPayable {
			t.Fatalf("checks payable=%+v err=%v", bankCheck, err)
		}
		dashboard, err := s.Dashboard(ctx, now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
		if err != nil || dashboard.PayableRial != ap {
			t.Fatalf("dashboard payable=%d want=%d err=%v", dashboard.PayableRial, ap, err)
		}
		balance, err := s.SupplierPayableBalance(ctx, supplier.ID)
		if err != nil || balance != ap {
			t.Fatalf("supplier payable=%d want=%d err=%v", balance, ap, err)
		}
	}
	check(0, 1000, 0)
	payment := domain.Payment{ID: "PAY-later", Direction: domain.PaymentOutgoing, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", AmountRial: 400, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{TargetType: "purchase", TargetID: purchase.ID, AmountRial: 400}}}
	posted, err := s.CreatePayment(ctx, payment)
	if err != nil {
		t.Fatal(err)
	}
	if posted.SupplierID != supplier.ID {
		t.Fatal("supplier not derived for journal")
	}
	if _, err = s.CreatePayment(ctx, payment); err != nil {
		t.Fatal(err)
	}
	check(400, 600, 0)
	c, err := s.CreateCheck(ctx, domain.Check{ID: "CHK-later", CheckNumber: "123456", Direction: domain.CheckOutgoing, Bank: "Bank", AmountRial: 600, IssueDate: now, DueDate: now.AddDate(0, 0, 10), PayerPayee: supplier.Name, SupplierID: supplier.ID, SourceType: "purchase", SourceID: purchase.ID, FinancialAccountID: "FIN-BANK", Status: domain.CheckDraft, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	check(400, 600, 0)
	if _, err = s.ChangeCheckStatus(ctx, c.ID, domain.CheckIssued, "", ""); err != nil {
		t.Fatal(err)
	}
	check(400, 600, 0)
	payment.ID = "PAY-double"
	payment.Allocations = []domain.PaymentAllocation{{TargetType: "purchase", TargetID: purchase.ID, AmountRial: 400}}
	if _, err = s.CreatePayment(ctx, payment); !errors.Is(err, domain.ErrAllocationExceeded) {
		t.Fatalf("pending check share paid again: %v", err)
	}
	if _, err = s.ChangeCheckStatus(ctx, c.ID, domain.CheckDelivered, "", ""); err != nil {
		t.Fatal(err)
	}
	check(400, 0, 600)
	if _, err = s.ChangeCheckStatus(ctx, c.ID, domain.CheckCleared, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ChangeCheckStatus(ctx, c.ID, domain.CheckCleared, "", ""); err != nil {
		t.Fatal(err)
	}
	check(1000, 0, 0)
	if _, err = s.ChangeCheckStatus(ctx, c.ID, domain.CheckReturned, "", ""); err != nil {
		t.Fatal(err)
	}
	check(400, 600, 0)
	if _, err = s.ReversePayment(ctx, "PAY-later", ""); err != nil {
		t.Fatal(err)
	}
	check(0, 1000, 0)
	payment.ID = "PAY-final"
	payment.Method = domain.PaymentBankTransfer
	payment.FinancialAccountID = "FIN-BANK"
	payment.AmountRial = 1000
	payment.Allocations = []domain.PaymentAllocation{{TargetType: "purchase", TargetID: purchase.ID, AmountRial: 1000}}
	if _, err = s.CreatePayment(ctx, payment); err != nil {
		t.Fatal(err)
	}
	check(1000, 0, 0)
}
