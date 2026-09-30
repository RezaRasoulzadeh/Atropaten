package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
	"Atropaten/internal/platform"
)

func newM5Store(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "m5.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestV10ToV11PreservesExistingCustomerAndAddsTreasuryTables(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy-v10.db")
	raw, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations[:10] {
		if _, err = raw.Exec(migration.sql); err != nil {
			raw.Close()
			t.Fatalf("migration %d: %v", migration.version, err)
		}
	}
	if _, err = raw.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		raw.Close()
		t.Fatal(err)
	}
	for _, migration := range migrations[:10] {
		if _, err = raw.Exec(`INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, migration.version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			raw.Close()
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err = raw.Exec(`INSERT INTO customers(id,name,phone,email,address,notes,active,created_at,updated_at) VALUES('CUS-OLD','Legacy customer','','','','',1,?,?)`, now, now); err != nil {
		raw.Close()
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
	var name string
	if err = s.db.QueryRow(`SELECT name FROM customers WHERE id='CUS-OLD'`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Legacy customer" {
		t.Fatalf("preserved customer=%q", name)
	}
	var version int
	if err = s.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil || version != platform.CurrentSchemaVersion {
		t.Fatalf("schema version=%d err=%v", version, err)
	}
	var tables int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('checks','check_events','loans','loan_installments','loan_payments')`).Scan(&tables); err != nil || tables != 5 {
		t.Fatalf("M5 tables=%d err=%v", tables, err)
	}
}

func m5Check(id, direction, status string) domain.Check {
	now := time.Now().UTC()
	return domain.Check{ID: id, CheckNumber: "USER-" + id, Direction: domain.CheckDirection(direction), Bank: "Test Bank", AmountRial: 1234567, IssueDate: now.Add(-24 * time.Hour), DueDate: now.Add(24 * time.Hour), PayerPayee: "Test party", Status: status, CreatedAt: now, UpdatedAt: now}
}

func journalCount(t *testing.T, s *Store) int {
	t.Helper()
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM journal_entries`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
func assertBalancedJournals(t *testing.T, s *Store) {
	t.Helper()
	rows, err := s.db.Query(`SELECT id FROM journal_entries`)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		var debit, credit int64
		if err := s.db.QueryRow(`SELECT COALESCE(SUM(debit_rial),0),COALESCE(SUM(credit_rial),0) FROM journal_lines WHERE journal_entry_id=?`, id).Scan(&debit, &credit); err != nil {
			t.Fatal(err)
		}
		if debit != credit {
			t.Fatalf("journal %s unbalanced: debit=%d credit=%d", id, debit, credit)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestIncomingCheckLifecycleIsControlledAndExcludesUnclearedBankCash(t *testing.T) {
	s := newM5Store(t)
	ctx := context.Background()
	check := m5Check("CHK-IN", string(domain.CheckIncoming), domain.CheckDraft)
	if _, err := s.CreateCheck(ctx, check); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ChangeCheckStatus(ctx, check.ID, domain.CheckCleared, "", "invalid"); !errors.Is(err, domain.ErrCheckTransition) {
		t.Fatalf("invalid transition error=%v", err)
	}
	if _, err := s.ChangeCheckStatus(ctx, check.ID, domain.CheckReceived, "received", "receive-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ChangeCheckStatus(ctx, check.ID, domain.CheckReceived, "received", "receive-1"); err != nil {
		t.Fatal(err)
	}
	bank, err := s.GetAccount(ctx, "ACC-BANK")
	if err != nil {
		t.Fatal(err)
	}
	if bank.BalanceRial != 0 {
		t.Fatalf("uncleared check changed bank balance: %d", bank.BalanceRial)
	}
	if _, err := s.ChangeCheckStatus(ctx, check.ID, domain.CheckDeposited, "deposited", "deposit-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ChangeCheckStatus(ctx, check.ID, domain.CheckCleared, "cleared", "clear-1"); err != nil {
		t.Fatal(err)
	}
	bank, _ = s.GetAccount(ctx, "ACC-BANK")
	if bank.BalanceRial != check.AmountRial {
		t.Fatalf("cleared bank balance=%d want %d", bank.BalanceRial, check.AmountRial)
	}
	count := journalCount(t, s)
	if _, err := s.ChangeCheckStatus(ctx, check.ID, domain.CheckCleared, "cleared", "clear-1"); err != nil {
		t.Fatal(err)
	}
	if journalCount(t, s) != count {
		t.Fatal("retry created duplicate check journal")
	}
	if _, err := s.ChangeCheckStatus(ctx, check.ID, domain.CheckReturned, "returned", "return-1"); err != nil {
		t.Fatal(err)
	}
	bank, _ = s.GetAccount(ctx, "ACC-BANK")
	if bank.BalanceRial != 0 {
		t.Fatalf("returned check left bank balance=%d", bank.BalanceRial)
	}
	events, err := s.ListCheckEvents(ctx, check.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 4 {
		t.Fatalf("events=%d want 4", len(events))
	}
	if _, err := s.ChangeCheckStatus(ctx, check.ID, domain.CheckCancelled, "cancelled", "cancel-1"); err != nil {
		t.Fatal(err)
	}
	if journalCount(t, s) != count+3 {
		t.Fatalf("cancel should not reverse history again; journal count=%d", journalCount(t, s))
	}
	assertBalancedJournals(t, s)
}

func TestOutgoingCheckClearingAndReturnRestoresPayable(t *testing.T) {
	s := newM5Store(t)
	ctx := context.Background()
	c := m5Check("CHK-OUT", string(domain.CheckOutgoing), domain.CheckDraft)
	if _, err := s.CreateCheck(ctx, c); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ChangeCheckStatus(ctx, c.ID, domain.CheckIssued, "", "i"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ChangeCheckStatus(ctx, c.ID, domain.CheckDelivered, "delivered", "d"); err != nil {
		t.Fatal(err)
	}
	ap, _ := s.GetAccount(ctx, "ACC-AP")
	if ap.BalanceRial != -c.AmountRial {
		t.Fatalf("AP after check delivery=%d", ap.BalanceRial)
	}
	if _, err := s.ChangeCheckStatus(ctx, c.ID, domain.CheckCleared, "cleared", "c"); err != nil {
		t.Fatal(err)
	}
	bank, _ := s.GetAccount(ctx, "ACC-BANK")
	if bank.BalanceRial != -c.AmountRial {
		t.Fatalf("bank after outgoing clear=%d", bank.BalanceRial)
	}
	if _, err := s.ChangeCheckStatus(ctx, c.ID, domain.CheckReturned, "returned", "r"); err != nil {
		t.Fatal(err)
	}
	ap, _ = s.GetAccount(ctx, "ACC-AP")
	bank, _ = s.GetAccount(ctx, "ACC-BANK")
	if ap.BalanceRial != 0 || bank.BalanceRial != 0 {
		t.Fatalf("return did not restore AP/bank: ap=%d bank=%d", ap.BalanceRial, bank.BalanceRial)
	}
	assertBalancedJournals(t, s)
}

func m5Loan(id, direction string) domain.Loan {
	now := time.Now().UTC()
	a := domain.LoanInstallment{ID: id + "-I1", LoanID: id, Position: 0, DueDate: now.Add(24 * time.Hour), PrincipalRial: 1000, InterestFeeRial: 100, TotalDueRial: 1100}
	return domain.Loan{ID: id, Direction: direction, CounterpartyName: "Counterparty", PrincipalRial: 1000, InterestFeeRial: 100, StartDate: now, Status: domain.LoanActive, FinancialAccountID: "FIN-CASH", IdempotencyKey: "key-" + id, CreatedAt: now, UpdatedAt: now, Installments: []domain.LoanInstallment{a}}
}

func TestLoansPostBalancedOpenPartialPaymentAndIdempotentReversal(t *testing.T) {
	s := newM5Store(t)
	ctx := context.Background()
	l := m5Loan("LOAN-PAY", domain.LoanPayable)
	if _, err := s.CreateLoan(ctx, l); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateLoan(ctx, l); err != nil {
		t.Fatal(err)
	}
	cash, _ := s.GetAccount(ctx, "ACC-CASH")
	liability, _ := s.GetAccount(ctx, "ACC-LOANS-PAYABLE")
	if cash.BalanceRial != 1000 || liability.BalanceRial != 1000 {
		t.Fatalf("opening balances cash=%d liability=%d", cash.BalanceRial, liability.BalanceRial)
	}
	p := domain.LoanPayment{ID: "LP-1", LoanID: l.ID, FinancialAccountID: "FIN-CASH", AmountRial: 550, PrincipalRial: 500, InterestRial: 50, PaidAt: time.Now().UTC(), Status: string(domain.PaymentPosted), IdempotencyKey: "pay-key", CreatedAt: time.Now().UTC(), Allocations: []domain.LoanPaymentAllocation{{ID: "LPA-1", PaymentID: "LP-1", InstallmentID: l.ID + "-I1", Position: 0, PrincipalRial: 500, InterestRial: 50}}}
	if _, err := s.CreateLoanPayment(ctx, p); err != nil {
		t.Fatal(err)
	}
	if payments, err := s.ListLoanPayments(ctx, l.ID); err != nil || len(payments) != 1 {
		t.Fatalf("list loan payments: n=%d err=%v", len(payments), err)
	}
	if _, err := s.CreateLoanPayment(ctx, p); err != nil {
		t.Fatal(err)
	}
	cash, _ = s.GetAccount(ctx, "ACC-CASH")
	liability, _ = s.GetAccount(ctx, "ACC-LOANS-PAYABLE")
	expense, _ := s.GetAccount(ctx, "ACC-FINANCE-EXPENSE")
	if cash.BalanceRial != 450 || liability.BalanceRial != 500 || expense.BalanceRial != 50 {
		t.Fatalf("payment balances cash=%d liability=%d expense=%d", cash.BalanceRial, liability.BalanceRial, expense.BalanceRial)
	}
	loan, _ := s.GetLoan(ctx, l.ID)
	if loan.PaidPrincipalRial != 500 || loan.PaidInterestRial != 50 || loan.RemainingPrincipalRial != 500 {
		t.Fatalf("loan derived balances %+v", loan)
	}
	if _, err := s.ReverseLoanPayment(ctx, p.ID, "reverse-pay"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReverseLoanPayment(ctx, p.ID, "reverse-pay"); err != nil {
		t.Fatal(err)
	}
	cash, _ = s.GetAccount(ctx, "ACC-CASH")
	liability, _ = s.GetAccount(ctx, "ACC-LOANS-PAYABLE")
	expense, _ = s.GetAccount(ctx, "ACC-FINANCE-EXPENSE")
	if cash.BalanceRial != 1000 || liability.BalanceRial != 1000 || expense.BalanceRial != 0 {
		t.Fatalf("reversal balances cash=%d liability=%d expense=%d", cash.BalanceRial, liability.BalanceRial, expense.BalanceRial)
	}
	assertBalancedJournals(t, s)
}

func TestReplaceLoanPaymentIsAtomicAndRebuildsBalances(t *testing.T) {
	s := newM5Store(t)
	ctx := context.Background()
	l := m5Loan("LOAN-EDIT", domain.LoanPayable)
	if _, err := s.CreateLoan(ctx, l); err != nil {
		t.Fatal(err)
	}
	original := domain.LoanPayment{ID: "LP-EDIT-1", LoanID: l.ID, FinancialAccountID: "FIN-CASH", AmountRial: 550, PrincipalRial: 500, InterestRial: 50, PaidAt: time.Now().UTC(), Status: string(domain.PaymentPosted), IdempotencyKey: "edit-original", CreatedAt: time.Now().UTC(), Allocations: []domain.LoanPaymentAllocation{{ID: "LPA-EDIT-1", PaymentID: "LP-EDIT-1", InstallmentID: l.ID + "-I1", PrincipalRial: 500, InterestRial: 50}}}
	if _, err := s.CreateLoanPayment(ctx, original); err != nil {
		t.Fatal(err)
	}

	invalid := domain.LoanPayment{ID: "LP-EDIT-BAD", LoanID: l.ID, FinancialAccountID: "FIN-CASH", AmountRial: 2200, PrincipalRial: 2000, InterestRial: 200, PaidAt: time.Now().UTC(), Status: string(domain.PaymentPosted), IdempotencyKey: "edit-invalid", CreatedAt: time.Now().UTC(), Allocations: []domain.LoanPaymentAllocation{{ID: "LPA-EDIT-BAD", PaymentID: "LP-EDIT-BAD", InstallmentID: l.ID + "-I1", PrincipalRial: 2000, InterestRial: 200}}}
	if _, err := s.ReplaceLoanPayment(ctx, original.ID, invalid); !errors.Is(err, domain.ErrLoanPaymentExceeded) {
		t.Fatalf("expected exceeded replacement error, got %v", err)
	}
	unchanged, err := s.GetLoanPayment(ctx, original.ID)
	if err != nil || unchanged.Status != string(domain.PaymentPosted) {
		t.Fatalf("failed replacement changed original payment: payment=%+v err=%v", unchanged, err)
	}

	replacement := domain.LoanPayment{ID: "LP-EDIT-2", LoanID: l.ID, FinancialAccountID: "FIN-CASH", AmountRial: 330, PrincipalRial: 300, InterestRial: 30, PaidAt: time.Now().UTC(), Notes: "corrected", Status: string(domain.PaymentPosted), IdempotencyKey: "edit-replacement", CreatedAt: time.Now().UTC(), Allocations: []domain.LoanPaymentAllocation{{ID: "LPA-EDIT-2", PaymentID: "LP-EDIT-2", InstallmentID: l.ID + "-I1", PrincipalRial: 300, InterestRial: 30}}}
	created, err := s.ReplaceLoanPayment(ctx, original.ID, replacement)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != replacement.ID || created.Status != string(domain.PaymentPosted) || created.Notes != "corrected" {
		t.Fatalf("unexpected replacement: %+v", created)
	}
	reversed, err := s.GetLoanPayment(ctx, original.ID)
	if err != nil || reversed.Status != string(domain.PaymentReversedState) {
		t.Fatalf("original was not reversed: payment=%+v err=%v", reversed, err)
	}
	loan, err := s.GetLoan(ctx, l.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loan.PaidPrincipalRial != 300 || loan.PaidInterestRial != 30 || loan.RemainingPrincipalRial != 700 {
		t.Fatalf("replacement did not rebuild loan totals: %+v", loan)
	}
	cash, _ := s.GetAccount(ctx, "ACC-CASH")
	liability, _ := s.GetAccount(ctx, "ACC-LOANS-PAYABLE")
	expense, _ := s.GetAccount(ctx, "ACC-FINANCE-EXPENSE")
	if cash.BalanceRial != 670 || liability.BalanceRial != 700 || expense.BalanceRial != 30 {
		t.Fatalf("replacement balances cash=%d liability=%d expense=%d", cash.BalanceRial, liability.BalanceRial, expense.BalanceRial)
	}
	assertBalancedJournals(t, s)
}

func TestUpdateAndRemoveLoanPreserveAccountingHistory(t *testing.T) {
	s := newM5Store(t)
	ctx := context.Background()
	l := m5Loan("LOAN-UPDATE", domain.LoanPayable)
	created, err := s.CreateLoan(ctx, l)
	if err != nil {
		t.Fatal(err)
	}
	updated := l
	updated.LoanNumber = created.LoanNumber
	updated.CounterpartyName = "Updated lender"
	updated.PrincipalRial = 1200
	updated.InterestFeeRial = 120
	updated.Installments = []domain.LoanInstallment{{ID: l.ID + "-EDIT-I1", LoanID: l.ID, Position: 0, DueDate: time.Now().UTC().Add(48 * time.Hour), PrincipalRial: 1200, InterestFeeRial: 120, TotalDueRial: 1320}}
	updated.UpdatedAt = time.Now().UTC()
	result, err := s.UpdateLoan(ctx, updated, "update-loan-once")
	if err != nil {
		t.Fatal(err)
	}
	if result.CounterpartyName != "Updated lender" || result.PrincipalRial != 1200 || result.InterestFeeRial != 120 || len(result.Installments) != 1 || result.Installments[0].ID != l.ID+"-EDIT-I1" {
		t.Fatalf("unexpected updated loan: %+v", result)
	}
	cash, _ := s.GetAccount(ctx, "ACC-CASH")
	liability, _ := s.GetAccount(ctx, "ACC-LOANS-PAYABLE")
	if cash.BalanceRial != 1200 || liability.BalanceRial != 1200 {
		t.Fatalf("loan update balances cash=%d liability=%d", cash.BalanceRial, liability.BalanceRial)
	}
	if _, err = s.UpdateLoan(ctx, updated, "update-loan-once"); err != nil {
		t.Fatalf("idempotent update failed: %v", err)
	}
	payment := domain.LoanPayment{ID: "LP-REMOVE", LoanID: l.ID, FinancialAccountID: "FIN-CASH", AmountRial: 330, PrincipalRial: 300, InterestRial: 30, PaidAt: time.Now().UTC(), Status: string(domain.PaymentPosted), IdempotencyKey: "remove-payment", CreatedAt: time.Now().UTC(), Allocations: []domain.LoanPaymentAllocation{{ID: "LPA-REMOVE", PaymentID: "LP-REMOVE", InstallmentID: l.ID + "-EDIT-I1", PrincipalRial: 300, InterestRial: 30}}}
	if _, err = s.CreateLoanPayment(ctx, payment); err != nil {
		t.Fatal(err)
	}
	metadata := updated
	metadata.CounterpartyName = "Metadata only"
	metadata.PrincipalRial = 9999
	metadata.Installments = []domain.LoanInstallment{{ID: "IGNORED", LoanID: l.ID, Position: 0, DueDate: time.Now().UTC(), PrincipalRial: 9999, InterestFeeRial: 120, TotalDueRial: 10119}}
	metadata.UpdatedAt = time.Now().UTC()
	result, err = s.UpdateLoan(ctx, metadata, "metadata-update")
	if err != nil {
		t.Fatal(err)
	}
	if result.CounterpartyName != "Metadata only" || result.PrincipalRial != 1200 || result.Installments[0].ID != l.ID+"-EDIT-I1" {
		t.Fatalf("payment history should lock financial terms: %+v", result)
	}
	if err = s.CancelLoan(ctx, l.ID, "remove-with-payment"); !errors.Is(err, domain.ErrLoanProtected) {
		t.Fatalf("expected protected removal, got %v", err)
	}
	if _, err = s.ReverseLoanPayment(ctx, payment.ID, "reverse-before-remove"); err != nil {
		t.Fatal(err)
	}
	if err = s.CancelLoan(ctx, l.ID, "remove-loan"); err != nil {
		t.Fatal(err)
	}
	cancelled, err := s.GetLoan(ctx, l.ID)
	if err != nil || cancelled.Status != domain.LoanCancelled {
		t.Fatalf("loan not cancelled: loan=%+v err=%v", cancelled, err)
	}
	visible, err := s.ListLoans(ctx, "", "")
	if err != nil || len(visible) != 0 {
		t.Fatalf("removed loan remained in active register: n=%d err=%v", len(visible), err)
	}
	cash, _ = s.GetAccount(ctx, "ACC-CASH")
	liability, _ = s.GetAccount(ctx, "ACC-LOANS-PAYABLE")
	if cash.BalanceRial != 0 || liability.BalanceRial != 0 {
		t.Fatalf("loan removal did not restore opening balances: cash=%d liability=%d", cash.BalanceRial, liability.BalanceRial)
	}
	assertBalancedJournals(t, s)
}

func TestReceivableLoanRepaymentUsesReceivableAndInterestIncome(t *testing.T) {
	s := newM5Store(t)
	ctx := context.Background()
	l := m5Loan("LOAN-REC", domain.LoanReceivable)
	if _, err := s.CreateLoan(ctx, l); err != nil {
		t.Fatal(err)
	}
	p := domain.LoanPayment{ID: "LP-REC", LoanID: l.ID, FinancialAccountID: "FIN-CASH", AmountRial: 1100, PrincipalRial: 1000, InterestRial: 100, PaidAt: time.Now().UTC(), Status: string(domain.PaymentPosted), IdempotencyKey: "pay-rec", CreatedAt: time.Now().UTC(), Allocations: []domain.LoanPaymentAllocation{{ID: "LPA-REC", PaymentID: "LP-REC", InstallmentID: l.ID + "-I1", Position: 0, PrincipalRial: 1000, InterestRial: 100}}}
	if _, err := s.CreateLoanPayment(ctx, p); err != nil {
		t.Fatal(err)
	}
	closed, err := s.ListLoans(ctx, domain.LoanReceivable, domain.LoanClosed)
	if err != nil || len(closed) != 1 {
		t.Fatalf("closed loan query: n=%d err=%v", len(closed), err)
	}
	cash, _ := s.GetAccount(ctx, "ACC-CASH")
	receivable, _ := s.GetAccount(ctx, "ACC-LOANS-RECEIVABLE")
	income, _ := s.GetAccount(ctx, "ACC-INTEREST-INCOME")
	if cash.BalanceRial != 100 || receivable.BalanceRial != 0 || income.BalanceRial != 100 {
		t.Fatalf("receivable repayment balances cash=%d receivable=%d income=%d", cash.BalanceRial, receivable.BalanceRial, income.BalanceRial)
	}
	assertBalancedJournals(t, s)
}
