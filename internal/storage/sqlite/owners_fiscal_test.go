package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestV11ToV12PreservesDataAndSeedsOwnerAccounts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v11.db")
	raw, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range migrations[:11] {
		if _, err = raw.Exec(m.sql); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = raw.Exec(`CREATE TABLE schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	for _, m := range migrations[:11] {
		if _, err = raw.Exec(`INSERT INTO schema_migrations VALUES(?,?)`, m.version, "2026-01-01T00:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = raw.Exec(`INSERT INTO customers(id,name,created_at,updated_at) VALUES('CUS-v11','Legacy','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err = raw.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := s.GetCustomer(context.Background(), "CUS-v11")
	if err != nil || c.Name != "Legacy" {
		t.Fatalf("customer=%+v err=%v", c, err)
	}
	var count int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM accounts WHERE id LIKE 'ACC-OWNER-%'`).Scan(&count); err != nil || count != 5 {
		t.Fatalf("owner account count=%d err=%v", count, err)
	}
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM accounts WHERE id='ACC-RETAINED-EARNINGS'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("retained account count=%d err=%v", count, err)
	}
}

func TestOwnerTransactionsAreBalancedDerivedAndDrawingIsNotExpense(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "owners.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 1, 5, 8, 0, 0, 0, time.UTC)
	o := domain.Owner{ID: "OWN-A", Name: "A", OwnershipBPS: 7000, ProfitSharingBPS: 6000, Active: true, CreatedAt: now, UpdatedAt: now}
	if _, err = s.CreateOwner(ctx, o); err != nil {
		t.Fatal(err)
	}
	post := func(id, typ string, amount int64, category string) {
		_, e := s.CreateOwnerTransaction(ctx, domain.OwnerTransaction{ID: id, OwnerID: o.ID, Type: typ, AmountRial: amount, OccurredAt: now, FinancialAccountID: "FIN-CASH", CategoryAccountID: category, Status: domain.OwnerTxPosted, IdempotencyKey: id, CreatedAt: now, UpdatedAt: now})
		if e != nil {
			t.Fatal(id, e)
		}
	}
	post("OTX-cap", domain.OwnerTxCapitalContribution, 1000, "")
	post("OTX-draw", domain.OwnerTxDrawing, 100, "")
	if _, err = s.CreateOwnerTransaction(ctx, domain.OwnerTransaction{ID: "OTX-exp", OwnerID: o.ID, Type: domain.OwnerTxPersonalExpense, AmountRial: 50, OccurredAt: now, CategoryAccountID: "ACC-EXP-OTHER", Status: domain.OwnerTxPosted, IdempotencyKey: "OTX-exp", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	post("OTX-reimb", domain.OwnerTxReimbursement, 50, "")
	post("OTX-loan-in", domain.OwnerTxLoanToBusiness, 200, "")
	post("OTX-loan-out", domain.OwnerTxLoanFromBusiness, 80, "")
	post("OTX-repay-in", domain.OwnerTxLoanRepaymentToOwner, 50, "")
	post("OTX-repay-out", domain.OwnerTxLoanRepaymentFromOwner, 30, "")
	if _, err = s.CreateOwnerTransaction(ctx, domain.OwnerTransaction{ID: "OTX-cap", OwnerID: o.ID, Type: domain.OwnerTxCapitalContribution, AmountRial: 1000, OccurredAt: now, FinancialAccountID: "FIN-CASH", Status: domain.OwnerTxPosted, IdempotencyKey: "OTX-cap", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal("retry:", err)
	}
	b, err := s.OwnerBalances(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != 1000 || b[1] != 100 || b[2] != 0 || b[3] != 150 || b[4] != 50 {
		t.Fatalf("balances=%v", b)
	}
	var expense int64
	if err = s.db.QueryRow(`SELECT COALESCE(SUM(debit_rial-credit_rial),0) FROM journal_lines WHERE account_id='ACC-EXP-OTHER'`).Scan(&expense); err != nil || expense != 50 {
		t.Fatalf("expense=%d err=%v", expense, err)
	}
	var entries int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM journal_entries`).Scan(&entries); err != nil || entries != 8 {
		t.Fatalf("entries=%d err=%v", entries, err)
	}
	if _, err = s.ReverseOwnerTransaction(ctx, "OTX-draw", ""); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReverseOwnerTransaction(ctx, "OTX-draw", ""); err != nil {
		t.Fatal("reverse retry", err)
	}
	b, err = s.OwnerBalances(ctx, o.ID)
	if err != nil || b[1] != 0 {
		t.Fatalf("reversed balances=%v err=%v", b, err)
	}
}

func TestFiscalPeriodPnLAllocationIsDeterministicAndClosedIsProtected(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "periods.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 8, 0, 0, 0, time.UTC)
	owners := []domain.Owner{{ID: "OWN-A", Name: "A", OwnershipBPS: 8000, ProfitSharingBPS: 7000, Active: true, CreatedAt: now, UpdatedAt: now}, {ID: "OWN-B", Name: "B", OwnershipBPS: 2000, ProfitSharingBPS: 3000, Active: true, CreatedAt: now, UpdatedAt: now}}
	for _, o := range owners {
		if _, err = s.CreateOwner(ctx, o); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.CreateFiscalPeriod(ctx, domain.FiscalPeriod{ID: "FP-2026", Name: "1405", StartDate: now, EndDate: end, Status: domain.FiscalPeriodOpen, IdempotencyKey: "FP-2026", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	entry := domain.JournalEntry{ID: "JE-PNL", Description: "Period activity", SourceType: "test", SourceID: "PNL", IdempotencyKey: "PNL", PostedAt: now, CreatedAt: now, Lines: []domain.JournalLine{{ID: "JE-PNL-L1", JournalEntryID: "JE-PNL", Position: 0, AccountID: "ACC-REVENUE", CreditRial: 1000}, {ID: "JE-PNL-L2", JournalEntryID: "JE-PNL", Position: 1, AccountID: "ACC-COGS", DebitRial: 200}, {ID: "JE-PNL-L3", JournalEntryID: "JE-PNL", Position: 2, AccountID: "ACC-EXP-OTHER", DebitRial: 101}, {ID: "JE-PNL-L4", JournalEntryID: "JE-PNL", Position: 3, AccountID: "ACC-CASH", DebitRial: 699}}}
	if _, err = s.PostJournalEntry(ctx, entry); err != nil {
		t.Fatal(err)
	}
	p, err := s.GetFiscalPeriod(ctx, "FP-2026")
	if err != nil || p.ProfitLossRial != 699 || p.RevenueRial != 1000 || p.COGSRial != 200 || p.ExpensesRial != 101 {
		t.Fatalf("period=%+v err=%v", p, err)
	}
	closed, err := s.CloseFiscalPeriod(ctx, "FP-2026", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(closed.Allocations) != 2 || closed.Allocations[0].AmountRial != 489 || closed.Allocations[1].AmountRial != 210 {
		t.Fatalf("allocations=%+v", closed.Allocations)
	}
	if _, err = s.CloseFiscalPeriod(ctx, "FP-2026", ""); err != nil {
		t.Fatal("close retry", err)
	}
	var count int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM journal_entries WHERE source_type='profit_allocation'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("close journals=%d err=%v", count, err)
	}
	if _, err = s.UpdateOwnerShares(ctx, "OWN-B", 2000, 2000, "new agreement"); err != nil {
		t.Fatal(err)
	}
	stable, err := s.GetFiscalPeriod(ctx, "FP-2026")
	if err != nil || stable.Allocations[1].AmountRial != 210 {
		t.Fatalf("stable allocations=%+v err=%v", stable.Allocations, err)
	}
	entry.ID = "JE-AFTER-CLOSE"
	entry.IdempotencyKey = "after-close"
	entry.Lines[0].ID = "JE-AFTER-CLOSE-L1"
	entry.Lines[0].JournalEntryID = entry.ID
	entry.Lines[1].ID = "JE-AFTER-CLOSE-L2"
	entry.Lines[1].JournalEntryID = entry.ID
	entry.Lines[2].ID = "JE-AFTER-CLOSE-L3"
	entry.Lines[2].JournalEntryID = entry.ID
	entry.Lines[3].ID = "JE-AFTER-CLOSE-L4"
	entry.Lines[3].JournalEntryID = entry.ID
	if _, err = s.PostJournalEntry(ctx, entry); !errors.Is(err, domain.ErrPeriodClosed) {
		t.Fatalf("post closed err=%v", err)
	}
	if _, err = s.ReverseJournalEntry(ctx, "JE-PNL", "", "", now.AddDate(1, 0, 0)); !errors.Is(err, domain.ErrPeriodClosed) {
		t.Fatalf("reverse closed err=%v", err)
	}
	if _, err = s.CreateFiscalPeriod(ctx, domain.FiscalPeriod{ID: "FP-overlap", Name: "overlap", StartDate: now.AddDate(0, 1, 0), EndDate: end.AddDate(0, -1, 0), Status: domain.FiscalPeriodOpen, IdempotencyKey: "FP-overlap", CreatedAt: now, UpdatedAt: now}); !errors.Is(err, domain.ErrPeriodOverlap) {
		t.Fatalf("overlap err=%v", err)
	}
}
