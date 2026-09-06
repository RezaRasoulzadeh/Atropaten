package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

const checkSelect = `SELECT id,check_number,direction,bank,branch,account_descriptor,amount_rial,issue_date,due_date,payer_payee,COALESCE(customer_id,''),COALESCE(supplier_id,''),source_type,source_id,COALESCE(financial_account_id,''),notes,status,created_at,updated_at FROM checks`
const loanSelect = `SELECT id,loan_number,direction,counterparty_name,COALESCE(customer_id,''),COALESCE(supplier_id,''),principal_rial,interest_fee_rial,start_date,COALESCE(end_date,''),status,notes,financial_account_id,journal_entry_id,idempotency_key,created_at,updated_at FROM loans`

func (s *Store) ListChecks(ctx context.Context, direction, status string) ([]domain.Check, error) {
	q, args := checkSelect+` WHERE 1=1`, []any{}
	if direction != "" && direction != "All" {
		q += ` AND direction=?`
		args = append(args, direction)
	}
	if status != "" && status != "All" {
		q += ` AND status=?`
		args = append(args, status)
	}
	q += ` ORDER BY due_date,status,check_number`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Check
	for rows.Next() {
		v, e := scanCheck(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) GetCheck(ctx context.Context, id string) (domain.Check, error) {
	v, err := scanCheck(s.db.QueryRowContext(ctx, checkSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Check{}, domain.ErrCheckNotFound
	}
	return v, err
}
func (s *Store) ListCheckEvents(ctx context.Context, id string) ([]domain.CheckEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,check_id,from_status,to_status,note,journal_entry_id,idempotency_key,occurred_at FROM check_events WHERE check_id=? ORDER BY occurred_at,id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CheckEvent
	for rows.Next() {
		var v domain.CheckEvent
		var at string
		if err = rows.Scan(&v.ID, &v.CheckID, &v.FromStatus, &v.ToStatus, &v.Note, &v.JournalEntryID, &v.IdempotencyKey, &at); err != nil {
			return nil, err
		}
		v.OccurredAt, err = time.Parse(time.RFC3339Nano, at)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) CreateCheck(ctx context.Context, c domain.Check) (domain.Check, error) {
	if c.Status == "" {
		c.Status = domain.CheckDraft
	}
	if c.Status != domain.CheckDraft {
		return domain.Check{}, domain.ErrCheckTransition
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}
	if err := c.Validate(); err != nil {
		return domain.Check{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Check{}, err
	}
	fail := func(e error) (domain.Check, error) { tx.Rollback(); return domain.Check{}, e }
	var existing string
	if err = tx.QueryRowContext(ctx, `SELECT id FROM checks WHERE id=?`, c.ID).Scan(&existing); err == nil {
		tx.Rollback()
		return s.GetCheck(ctx, c.ID)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fail(err)
	}
	if c.CheckNumber == "" {
		var n int64
		if err = tx.QueryRowContext(ctx, `SELECT next_number FROM check_number_sequences WHERE id=1`).Scan(&n); err != nil {
			return fail(err)
		}
		c.CheckNumber = fmt.Sprintf("CHK-%04d", n)
		if _, err = tx.ExecContext(ctx, `UPDATE check_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
			return fail(err)
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO checks(id,check_number,direction,bank,branch,account_descriptor,amount_rial,issue_date,due_date,payer_payee,customer_id,supplier_id,source_type,source_id,financial_account_id,notes,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, c.ID, c.CheckNumber, c.Direction, c.Bank, c.Branch, c.AccountDescriptor, c.AmountRial, c.IssueDate.UTC().Format(time.RFC3339Nano), c.DueDate.UTC().Format(time.RFC3339Nano), c.PayerPayee, nullableString(c.CustomerID), nullableString(c.SupplierID), c.SourceType, c.SourceID, nullableString(c.FinancialAccountID), c.Notes, c.Status, c.CreatedAt.UTC().Format(time.RFC3339Nano), c.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.Check{}, err
	}
	return s.GetCheck(ctx, c.ID)
}
func (s *Store) DeleteDraftCheck(ctx context.Context, id string) error {
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM checks WHERE id=?`, id).Scan(&status); errors.Is(err, sql.ErrNoRows) {
		return domain.ErrCheckNotFound
	} else if err != nil {
		return err
	}
	if status != domain.CheckDraft {
		return domain.ErrCheckProtected
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM checks WHERE id=? AND status='Draft'`, id)
	return err
}

func (s *Store) ChangeCheckStatus(ctx context.Context, id, to, note, key string) (domain.Check, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Check{}, err
	}
	fail := func(e error) (domain.Check, error) { tx.Rollback(); return domain.Check{}, e }
	var c domain.Check
	var issue, due, created, updated string
	err = tx.QueryRowContext(ctx, checkSelect+` WHERE id=?`, id).Scan(&c.ID, &c.CheckNumber, &c.Direction, &c.Bank, &c.Branch, &c.AccountDescriptor, &c.AmountRial, &issue, &due, &c.PayerPayee, &c.CustomerID, &c.SupplierID, &c.SourceType, &c.SourceID, &c.FinancialAccountID, &c.Notes, &c.Status, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrCheckNotFound)
	}
	if err != nil {
		return fail(err)
	}
	if key == "" {
		key = "check:transition:" + id + ":" + to
	}
	var prior string
	if e := tx.QueryRowContext(ctx, `SELECT id FROM check_events WHERE idempotency_key=?`, key).Scan(&prior); e == nil {
		tx.Rollback()
		return s.GetCheck(ctx, id)
	} else if !errors.Is(e, sql.ErrNoRows) {
		return fail(e)
	}
	if !domain.ValidCheckTransition(c.Direction, c.Status, to) {
		return fail(domain.ErrCheckTransition)
	}
	if checkNeedsAccounting(c.Direction, c.Status, to) && c.SourceType != "" && c.SourceID != "" {
		var paid int
		if e := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type=? AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, c.SourceType, c.SourceID).Scan(&paid); e != nil {
			return fail(e)
		}
		if paid > 0 {
			return fail(domain.ErrCheckObligationPaid)
		}
	}
	var journals []string
	if checkNeedsAccounting(c.Direction, c.Status, to) {
		if c.Direction == domain.CheckIncoming && c.Status == domain.CheckDraft && to == domain.CheckReceived {
			j, e := s.postCheckJournal(ctx, tx, c, "receive", c.FinancialAccountID)
			if e != nil {
				return fail(e)
			}
			journals = append(journals, j.ID)
		}
		if c.Direction == domain.CheckIncoming && c.Status == domain.CheckReceived && to == domain.CheckDeposited {
			j, e := s.postCheckJournal(ctx, tx, c, "deposit", c.FinancialAccountID)
			if e != nil {
				return fail(e)
			}
			journals = append(journals, j.ID)
		}
		if c.Direction == domain.CheckIncoming && c.Status == domain.CheckDeposited && to == domain.CheckCleared {
			j, e := s.postCheckJournal(ctx, tx, c, "clear", c.FinancialAccountID)
			if e != nil {
				return fail(e)
			}
			journals = append(journals, j.ID)
		}
		if c.Direction == domain.CheckOutgoing && c.Status == domain.CheckIssued && to == domain.CheckDelivered {
			j, e := s.postCheckJournal(ctx, tx, c, "deliver", c.FinancialAccountID)
			if e != nil {
				return fail(e)
			}
			journals = append(journals, j.ID)
		}
		if c.Direction == domain.CheckOutgoing && c.Status == domain.CheckDelivered && to == domain.CheckCleared {
			j, e := s.postCheckJournal(ctx, tx, c, "clear", c.FinancialAccountID)
			if e != nil {
				return fail(e)
			}
			journals = append(journals, j.ID)
		}
	}
	if checkIsReturn(c.Direction, c.Status, to) {
		events, e := s.checkEventsTx(ctx, tx, id)
		if e != nil {
			return fail(e)
		}
		for n := len(events) - 1; n >= 0; n-- {
			for _, jID := range strings.Split(events[n].JournalEntryID, ",") {
				if jID == "" {
					continue
				}
				rev, e := s.reverseJournalTx(ctx, tx, jID, "check:return:"+id+":"+to+":"+fmt.Sprint(n), "Return check "+c.CheckNumber, time.Now().UTC())
				if e != nil {
					return fail(e)
				}
				journals = append(journals, rev.ID)
			}
		}
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE checks SET status=?,updated_at=? WHERE id=?`, to, now.Format(time.RFC3339Nano), id); err != nil {
		return fail(err)
	}
	eventID := "CHE-" + id + "-" + fmt.Sprint(now.UnixNano())
	if _, err = tx.ExecContext(ctx, `INSERT INTO check_events(id,check_id,from_status,to_status,note,journal_entry_id,idempotency_key,occurred_at) VALUES(?,?,?,?,?,?,?,?)`, eventID, id, c.Status, to, note, strings.Join(journals, ","), key, now.Format(time.RFC3339Nano)); err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.Check{}, err
	}
	return s.GetCheck(ctx, id)
}
func checkNeedsAccounting(direction domain.CheckDirection, from, to string) bool {
	return (direction == domain.CheckIncoming && (from == domain.CheckDraft && to == domain.CheckReceived || from == domain.CheckReceived && to == domain.CheckDeposited || from == domain.CheckDeposited && to == domain.CheckCleared)) || (direction == domain.CheckOutgoing && (from == domain.CheckIssued && to == domain.CheckDelivered || from == domain.CheckDelivered && to == domain.CheckCleared))
}
func checkIsReturn(direction domain.CheckDirection, from, to string) bool {
	if from == domain.CheckReturned || from == domain.CheckRejected || from == domain.CheckCancelled {
		return false
	}
	return (direction == domain.CheckIncoming && (to == domain.CheckReturned || to == domain.CheckCancelled)) || (direction == domain.CheckOutgoing && (to == domain.CheckReturned || to == domain.CheckRejected || to == domain.CheckCancelled))
}
func (s *Store) checkEventsTx(ctx context.Context, tx *sql.Tx, id string) ([]domain.CheckEvent, error) {
	rows, e := tx.QueryContext(ctx, `SELECT id,check_id,from_status,to_status,note,journal_entry_id,idempotency_key,occurred_at FROM check_events WHERE check_id=? ORDER BY occurred_at,id`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.CheckEvent
	for rows.Next() {
		var v domain.CheckEvent
		var at string
		if e = rows.Scan(&v.ID, &v.CheckID, &v.FromStatus, &v.ToStatus, &v.Note, &v.JournalEntryID, &v.IdempotencyKey, &at); e != nil {
			return nil, e
		}
		v.OccurredAt, _ = time.Parse(time.RFC3339Nano, at)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) postCheckJournal(ctx context.Context, tx *sql.Tx, c domain.Check, action, account string) (domain.JournalEntry, error) {
	bank := account
	if bank == "" {
		bank = "FIN-BANK"
	}
	var ledger, accountType string
	var active int
	if err := tx.QueryRowContext(ctx, `SELECT active,ledger_account_id,type FROM financial_accounts WHERE id=?`, bank).Scan(&active, &ledger, &accountType); errors.Is(err, sql.ErrNoRows) {
		return domain.JournalEntry{}, domain.ErrFinancialAccountNotFound
	} else if err != nil {
		return domain.JournalEntry{}, err
	}
	if active == 0 {
		return domain.JournalEntry{}, domain.ErrAccountInactive
	}
	if action == "clear" && accountType != "bank" {
		return domain.JournalEntry{}, fmt.Errorf("a cleared check must use an active bank account")
	}
	id := "JE-CHK-" + c.ID + "-" + action
	var lines []domain.JournalLine
	switch {
	case c.Direction == domain.CheckIncoming && action == "receive":
		lines = []domain.JournalLine{{ID: id + "-L1", JournalEntryID: id, Position: 0, AccountID: "ACC-CHECKS-RECEIVABLE", DebitRial: c.AmountRial, PartyType: "customer", PartyID: c.CustomerID, Memo: "Receive customer check"}, {ID: id + "-L2", JournalEntryID: id, Position: 1, AccountID: "ACC-AR", CreditRial: c.AmountRial, PartyType: "customer", PartyID: c.CustomerID, Memo: "Settle receivable by check"}}
	case c.Direction == domain.CheckIncoming && action == "deposit":
		lines = []domain.JournalLine{{ID: id + "-L1", JournalEntryID: id, Position: 0, AccountID: "ACC-CHECKS-IN-TRANSIT", DebitRial: c.AmountRial, Memo: "Deposit check in transit"}, {ID: id + "-L2", JournalEntryID: id, Position: 1, AccountID: "ACC-CHECKS-RECEIVABLE", CreditRial: c.AmountRial, Memo: "Deposit check"}}
	case c.Direction == domain.CheckIncoming && action == "clear":
		lines = []domain.JournalLine{{ID: id + "-L1", JournalEntryID: id, Position: 0, AccountID: ledger, DebitRial: c.AmountRial, Memo: "Cleared incoming check"}, {ID: id + "-L2", JournalEntryID: id, Position: 1, AccountID: "ACC-CHECKS-IN-TRANSIT", CreditRial: c.AmountRial, Memo: "Clear check in transit"}}
	case c.Direction == domain.CheckOutgoing && action == "deliver":
		lines = []domain.JournalLine{{ID: id + "-L1", JournalEntryID: id, Position: 0, AccountID: "ACC-AP", DebitRial: c.AmountRial, PartyType: "supplier", PartyID: c.SupplierID, Memo: "Deliver outgoing check"}, {ID: id + "-L2", JournalEntryID: id, Position: 1, AccountID: "ACC-CHECKS-PAYABLE", CreditRial: c.AmountRial, Memo: "Outstanding check"}}
	case c.Direction == domain.CheckOutgoing && action == "clear":
		lines = []domain.JournalLine{{ID: id + "-L1", JournalEntryID: id, Position: 0, AccountID: "ACC-CHECKS-PAYABLE", DebitRial: c.AmountRial, Memo: "Clear outgoing check"}, {ID: id + "-L2", JournalEntryID: id, Position: 1, AccountID: ledger, CreditRial: c.AmountRial, Memo: "Bank cleared outgoing check"}}
	}
	entry := domain.JournalEntry{ID: id, Description: action + " check " + c.CheckNumber, SourceType: "check", SourceID: c.ID, IdempotencyKey: "check:" + c.ID + ":" + action, PostedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(), Lines: lines}
	return s.postJournalTx(ctx, tx, entry)
}

func (s *Store) ListLoans(ctx context.Context, direction, status string) ([]domain.Loan, error) {
	q, args := loanSelect+` WHERE 1=1`, []any{}
	if direction != "" && direction != "All" {
		q += ` AND direction=?`
		args = append(args, direction)
	}
	if status != "" && status != "All" {
		q += ` AND status=?`
		args = append(args, status)
	}
	q += ` ORDER BY start_date DESC,loan_number DESC`
	rows, e := s.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.Loan
	for rows.Next() {
		v, e := scanLoan(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	if e = rows.Close(); e != nil {
		return nil, e
	}
	for i := range out {
		if e = s.loadLoanDetails(ctx, &out[i]); e != nil {
			return nil, e
		}
	}
	return out, nil
}
func (s *Store) GetLoan(ctx context.Context, id string) (domain.Loan, error) {
	v, e := scanLoan(s.db.QueryRowContext(ctx, loanSelect+` WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.Loan{}, domain.ErrLoanNotFound
	}
	if e != nil {
		return v, e
	}
	e = s.loadLoanDetails(ctx, &v)
	return v, e
}
func (s *Store) CreateLoan(ctx context.Context, l domain.Loan) (domain.Loan, error) {
	if l.Status == "" {
		l.Status = domain.LoanActive
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	if l.UpdatedAt.IsZero() {
		l.UpdatedAt = l.CreatedAt
	}
	if err := l.Validate(); err != nil {
		return domain.Loan{}, err
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return domain.Loan{}, e
	}
	fail := func(x error) (domain.Loan, error) { tx.Rollback(); return domain.Loan{}, x }
	var existing string
	if e = tx.QueryRowContext(ctx, `SELECT id FROM loans WHERE idempotency_key=?`, l.IdempotencyKey).Scan(&existing); e == nil {
		tx.Rollback()
		return s.GetLoan(ctx, existing)
	} else if !errors.Is(e, sql.ErrNoRows) {
		return fail(e)
	}
	var active int
	var ledgerID, accountType string
	if e = tx.QueryRowContext(ctx, `SELECT active,ledger_account_id,type FROM financial_accounts WHERE id=?`, l.FinancialAccountID).Scan(&active, &ledgerID, &accountType); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrFinancialAccountNotFound)
	} else if e != nil {
		return fail(e)
	}
	if active == 0 {
		return fail(domain.ErrAccountInactive)
	}
	if accountType != "cash" && accountType != "bank" {
		return fail(fmt.Errorf("loans require a cash or bank financial account"))
	}
	var n int64
	if e = tx.QueryRowContext(ctx, `SELECT next_number FROM loan_number_sequences WHERE id=1`).Scan(&n); e != nil {
		return fail(e)
	}
	l.LoanNumber = fmt.Sprintf("LOAN-%04d", n)
	if _, e = tx.ExecContext(ctx, `UPDATE loan_number_sequences SET next_number=next_number+1 WHERE id=1`); e != nil {
		return fail(e)
	}
	now := time.Now().UTC()
	je := "JE-LOAN-" + l.ID
	var lines []domain.JournalLine
	if l.Direction == domain.LoanPayable {
		lines = []domain.JournalLine{{ID: je + "-L1", JournalEntryID: je, Position: 0, AccountID: ledgerID, DebitRial: l.PrincipalRial, Memo: "Borrowed loan receipt"}, {ID: je + "-L2", JournalEntryID: je, Position: 1, AccountID: "ACC-LOANS-PAYABLE", CreditRial: l.PrincipalRial, PartyID: l.CounterpartyName, Memo: "Loan payable"}}
	} else {
		lines = []domain.JournalLine{{ID: je + "-L1", JournalEntryID: je, Position: 0, AccountID: "ACC-LOANS-RECEIVABLE", DebitRial: l.PrincipalRial, PartyID: l.CounterpartyName, Memo: "Loan disbursement"}, {ID: je + "-L2", JournalEntryID: je, Position: 1, AccountID: ledgerID, CreditRial: l.PrincipalRial, Memo: "Lent loan cash"}}
	}
	if _, e = s.postJournalTx(ctx, tx, domain.JournalEntry{ID: je, Description: "Open loan " + l.LoanNumber, SourceType: "loan", SourceID: l.ID, IdempotencyKey: "loan:" + l.IdempotencyKey, PostedAt: l.StartDate, CreatedAt: now, Lines: lines}); e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `INSERT INTO loans(id,loan_number,direction,counterparty_name,customer_id,supplier_id,principal_rial,interest_fee_rial,start_date,end_date,status,notes,financial_account_id,journal_entry_id,idempotency_key,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, l.ID, l.LoanNumber, l.Direction, l.CounterpartyName, nullableString(l.CustomerID), nullableString(l.SupplierID), l.PrincipalRial, l.InterestFeeRial, l.StartDate.UTC().Format(time.RFC3339Nano), nullableTime(l.EndDate), l.Status, l.Notes, l.FinancialAccountID, je, l.IdempotencyKey, l.CreatedAt.UTC().Format(time.RFC3339Nano), l.UpdatedAt.UTC().Format(time.RFC3339Nano)); e != nil {
		return fail(e)
	}
	for _, i := range l.Installments {
		if _, e = tx.ExecContext(ctx, `INSERT INTO loan_installments(id,loan_id,position,due_date,principal_rial,interest_fee_rial,total_due_rial) VALUES(?,?,?,?,?,?,?)`, i.ID, l.ID, i.Position, i.DueDate.UTC().Format(time.RFC3339Nano), i.PrincipalRial, i.InterestFeeRial, i.TotalDueRial); e != nil {
			return fail(e)
		}
	}
	if e = tx.Commit(); e != nil {
		return domain.Loan{}, e
	}
	return s.GetLoan(ctx, l.ID)
}

func (s *Store) ListLoanPayments(ctx context.Context, loanID string) ([]domain.LoanPayment, error) {
	q := `SELECT id,payment_number,loan_id,financial_account_id,amount_rial,principal_rial,interest_rial,paid_at,notes,status,journal_entry_id,idempotency_key,created_at FROM loan_payments`
	args := []any{}
	if loanID != "" {
		q += ` WHERE loan_id=?`
		args = append(args, loanID)
	}
	q += ` ORDER BY paid_at DESC,payment_number DESC`
	rows, e := s.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.LoanPayment
	for rows.Next() {
		v, e := scanLoanPayment(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	if e = rows.Close(); e != nil {
		return nil, e
	}
	for i := range out {
		out[i].Allocations, e = s.loadLoanPaymentAllocations(ctx, out[i].ID)
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
func (s *Store) CreateLoanPayment(ctx context.Context, p domain.LoanPayment) (domain.LoanPayment, error) {
	if p.Status == "" {
		p.Status = string(domain.PaymentPosted)
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = p.PaidAt
	}
	if err := p.Validate(); err != nil {
		return domain.LoanPayment{}, err
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return domain.LoanPayment{}, e
	}
	fail := func(x error) (domain.LoanPayment, error) { tx.Rollback(); return domain.LoanPayment{}, x }
	var existing string
	if e = tx.QueryRowContext(ctx, `SELECT id FROM loan_payments WHERE idempotency_key=?`, p.IdempotencyKey).Scan(&existing); e == nil {
		tx.Rollback()
		return s.GetLoanPayment(ctx, existing)
	} else if !errors.Is(e, sql.ErrNoRows) {
		return fail(e)
	}
	var dir string
	var principal int64
	if e = tx.QueryRowContext(ctx, `SELECT direction,principal_rial FROM loans WHERE id=? AND status='Active'`, p.LoanID).Scan(&dir, &principal); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrLoanNotFound)
	} else if e != nil {
		return fail(e)
	}
	var active int
	var ledger, accountType string
	if e = tx.QueryRowContext(ctx, `SELECT active,ledger_account_id,type FROM financial_accounts WHERE id=?`, p.FinancialAccountID).Scan(&active, &ledger, &accountType); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrFinancialAccountNotFound)
	} else if e != nil {
		return fail(e)
	}
	if active == 0 {
		return fail(domain.ErrAccountInactive)
	}
	if accountType != "cash" && accountType != "bank" {
		return fail(fmt.Errorf("loan payments require a cash or bank financial account"))
	}
	var paid int64
	if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(principal_rial),0) FROM loan_payments WHERE loan_id=? AND status='posted'`, p.LoanID).Scan(&paid); e != nil {
		return fail(e)
	}
	if paid+p.PrincipalRial > principal {
		return fail(domain.ErrLoanPaymentExceeded)
	}
	for _, a := range p.Allocations {
		var lp, li int64
		if e = tx.QueryRowContext(ctx, `SELECT principal_rial,interest_fee_rial FROM loan_installments WHERE id=? AND loan_id=?`, a.InstallmentID, p.LoanID).Scan(&lp, &li); e != nil {
			return fail(e)
		}
		var ap, ai int64
		if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.principal_rial),0),COALESCE(SUM(a.interest_rial),0) FROM loan_payment_allocations a JOIN loan_payments p ON p.id=a.payment_id WHERE a.installment_id=? AND p.status='posted'`, a.InstallmentID).Scan(&ap, &ai); e != nil {
			return fail(e)
		}
		if ap+a.PrincipalRial > lp || ai+a.InterestRial > li {
			return fail(domain.ErrLoanPaymentExceeded)
		}
	}
	var totalDue, totalPaid int64
	if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_due_rial),0) FROM loan_installments WHERE loan_id=?`, p.LoanID).Scan(&totalDue); e != nil {
		return fail(e)
	}
	if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.principal_rial+a.interest_rial),0) FROM loan_payment_allocations a JOIN loan_payments lp ON lp.id=a.payment_id WHERE lp.loan_id=? AND lp.status='posted'`, p.LoanID).Scan(&totalPaid); e != nil {
		return fail(e)
	}
	now := time.Now().UTC()
	je := "JE-LPAY-" + p.ID
	lines := []domain.JournalLine{}
	pos := 0
	if p.PrincipalRial > 0 {
		if dir == domain.LoanPayable {
			lines = append(lines, domain.JournalLine{ID: je + "-L" + fmt.Sprint(pos+1), JournalEntryID: je, Position: pos, AccountID: "ACC-LOANS-PAYABLE", DebitRial: p.PrincipalRial, Memo: "Loan principal repayment"})
		} else {
			lines = append(lines, domain.JournalLine{ID: je + "-L" + fmt.Sprint(pos+1), JournalEntryID: je, Position: pos, AccountID: "ACC-LOANS-RECEIVABLE", CreditRial: p.PrincipalRial, Memo: "Loan receivable principal"})
		}
		pos++
	}
	if p.InterestRial > 0 {
		acct := "ACC-INTEREST-INCOME"
		if dir == domain.LoanPayable {
			acct = "ACC-FINANCE-EXPENSE"
		}
		if dir == domain.LoanPayable {
			lines = append(lines, domain.JournalLine{ID: je + "-L" + fmt.Sprint(pos+1), JournalEntryID: je, Position: pos, AccountID: acct, DebitRial: p.InterestRial, Memo: "Loan interest/fee"})
		} else {
			lines = append(lines, domain.JournalLine{ID: je + "-L" + fmt.Sprint(pos+1), JournalEntryID: je, Position: pos, AccountID: acct, CreditRial: p.InterestRial, Memo: "Loan interest income"})
		}
		pos++
	}
	lines = append(lines, domain.JournalLine{ID: je + "-L" + fmt.Sprint(pos+1), JournalEntryID: je, Position: pos, AccountID: ledger, DebitRial: choose(dir == domain.LoanReceivable, p.AmountRial, 0), CreditRial: choose(dir == domain.LoanPayable, p.AmountRial, 0), Memo: "Loan payment cash/bank"})
	if _, e = s.postJournalTx(ctx, tx, domain.JournalEntry{ID: je, Description: "Loan payment", SourceType: "loan_payment", SourceID: p.ID, IdempotencyKey: "loan-payment:" + p.IdempotencyKey, PostedAt: p.PaidAt, CreatedAt: now, Lines: lines}); e != nil {
		return fail(e)
	}
	var n int64
	if e = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(CAST(SUBSTR(payment_number,6) AS INTEGER)),1000)+1 FROM loan_payments`).Scan(&n); e != nil {
		return fail(e)
	}
	p.PaymentNumber = fmt.Sprintf("LPAY-%04d", n)
	if _, e = tx.ExecContext(ctx, `INSERT INTO loan_payments(id,payment_number,loan_id,financial_account_id,amount_rial,principal_rial,interest_rial,paid_at,notes,status,journal_entry_id,idempotency_key,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, p.ID, p.PaymentNumber, p.LoanID, p.FinancialAccountID, p.AmountRial, p.PrincipalRial, p.InterestRial, p.PaidAt.UTC().Format(time.RFC3339Nano), p.Notes, p.Status, je, p.IdempotencyKey, p.CreatedAt.UTC().Format(time.RFC3339Nano)); e != nil {
		return fail(e)
	}
	for _, a := range p.Allocations {
		if _, e = tx.ExecContext(ctx, `INSERT INTO loan_payment_allocations(id,payment_id,installment_id,position,principal_rial,interest_rial) VALUES(?,?,?,?,?,?)`, a.ID, p.ID, a.InstallmentID, a.Position, a.PrincipalRial, a.InterestRial); e != nil {
			return fail(e)
		}
	}
	if totalPaid+p.AmountRial == totalDue && totalDue > 0 {
		if _, e = tx.ExecContext(ctx, `UPDATE loans SET status='Closed',updated_at=? WHERE id=? AND status='Active'`, now.Format(time.RFC3339Nano), p.LoanID); e != nil {
			return fail(e)
		}
	}
	if e = tx.Commit(); e != nil {
		return domain.LoanPayment{}, e
	}
	return s.GetLoanPayment(ctx, p.ID)
}
func (s *Store) GetLoanPayment(ctx context.Context, id string) (domain.LoanPayment, error) {
	v, e := scanLoanPayment(s.db.QueryRowContext(ctx, `SELECT id,payment_number,loan_id,financial_account_id,amount_rial,principal_rial,interest_rial,paid_at,notes,status,journal_entry_id,idempotency_key,created_at FROM loan_payments WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.LoanPayment{}, domain.ErrLoanPaymentNotFound
	}
	if e != nil {
		return v, e
	}
	v.Allocations, e = s.loadLoanPaymentAllocations(ctx, id)
	return v, e
}
func (s *Store) ReverseLoanPayment(ctx context.Context, id, key string) (domain.LoanPayment, error) {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return domain.LoanPayment{}, e
	}
	fail := func(x error) (domain.LoanPayment, error) { tx.Rollback(); return domain.LoanPayment{}, x }
	var status, je string
	if e = tx.QueryRowContext(ctx, `SELECT status,journal_entry_id FROM loan_payments WHERE id=?`, id).Scan(&status, &je); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrLoanPaymentNotFound)
	} else if e != nil {
		return fail(e)
	}
	if status == string(domain.PaymentReversedState) {
		tx.Rollback()
		return s.GetLoanPayment(ctx, id)
	}
	if key == "" {
		key = "loan-payment:reverse:" + id
	}
	if _, e = s.reverseJournalTx(ctx, tx, je, key, "Reverse loan payment", time.Now().UTC()); e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `UPDATE loan_payments SET status='reversed' WHERE id=?`, id); e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `UPDATE loans SET status='Active',updated_at=? WHERE id=(SELECT loan_id FROM loan_payments WHERE id=?) AND status='Closed'`, time.Now().UTC().Format(time.RFC3339Nano), id); e != nil {
		return fail(e)
	}
	if e = tx.Commit(); e != nil {
		return domain.LoanPayment{}, e
	}
	return s.GetLoanPayment(ctx, id)
}

func (s *Store) loadLoanDetails(ctx context.Context, l *domain.Loan) error {
	rows, e := s.db.QueryContext(ctx, `SELECT id,loan_id,position,due_date,principal_rial,interest_fee_rial,total_due_rial FROM loan_installments WHERE loan_id=? ORDER BY position`, l.ID)
	if e != nil {
		return e
	}
	defer rows.Close()
	var installments []domain.LoanInstallment
	now := time.Now().UTC()
	for rows.Next() {
		var i domain.LoanInstallment
		var due string
		if e = rows.Scan(&i.ID, &i.LoanID, &i.Position, &due, &i.PrincipalRial, &i.InterestFeeRial, &i.TotalDueRial); e != nil {
			return e
		}
		i.DueDate, e = time.Parse(time.RFC3339Nano, due)
		if e != nil {
			return e
		}
		installments = append(installments, i)
	}
	if e = rows.Err(); e != nil {
		return e
	}
	if e = rows.Close(); e != nil {
		return e
	}
	for _, i := range installments {
		if e = s.installmentPaid(ctx, &i); e != nil {
			return e
		}
		if i.RemainingRial > 0 && i.DueDate.Before(now) {
			i.OverdueRial = i.RemainingRial
			l.OverdueRial += i.RemainingRial
		}
		l.PaidPrincipalRial += i.PaidPrincipalRial
		l.PaidInterestRial += i.PaidInterestRial
		l.Installments = append(l.Installments, i)
	}
	l.RemainingPrincipalRial = l.PrincipalRial - l.PaidPrincipalRial
	if l.RemainingPrincipalRial < 0 {
		l.RemainingPrincipalRial = 0
	}
	l.RemainingInterestRial = l.InterestFeeRial - l.PaidInterestRial
	if l.RemainingInterestRial < 0 {
		l.RemainingInterestRial = 0
	}
	if l.RemainingPrincipalRial == 0 && l.RemainingInterestRial == 0 && l.Status == domain.LoanActive {
		l.Status = domain.LoanClosed
	}
	return rows.Err()
}
func (s *Store) installmentPaid(ctx context.Context, i *domain.LoanInstallment) error {
	if e := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.principal_rial),0),COALESCE(SUM(a.interest_rial),0) FROM loan_payment_allocations a JOIN loan_payments p ON p.id=a.payment_id WHERE a.installment_id=? AND p.status='posted'`, i.ID).Scan(&i.PaidPrincipalRial, &i.PaidInterestRial); e != nil {
		return e
	}
	i.PaidRial = i.PaidPrincipalRial + i.PaidInterestRial
	i.RemainingRial = i.TotalDueRial - i.PaidRial
	if i.RemainingRial < 0 {
		i.RemainingRial = 0
	}
	if i.RemainingRial == 0 {
		i.Status = "Paid"
	} else if i.PaidRial > 0 {
		i.Status = "Partially Paid"
	} else {
		i.Status = "Open"
	}
	return nil
}
func (s *Store) loadLoanPaymentAllocations(ctx context.Context, id string) ([]domain.LoanPaymentAllocation, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,payment_id,installment_id,position,principal_rial,interest_rial FROM loan_payment_allocations WHERE payment_id=? ORDER BY position`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.LoanPaymentAllocation
	for rows.Next() {
		var v domain.LoanPaymentAllocation
		if e = rows.Scan(&v.ID, &v.PaymentID, &v.InstallmentID, &v.Position, &v.PrincipalRial, &v.InterestRial); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func scanCheck(row scanner) (domain.Check, error) {
	var c domain.Check
	var direction, issue, due, created, updated string
	if e := row.Scan(&c.ID, &c.CheckNumber, &direction, &c.Bank, &c.Branch, &c.AccountDescriptor, &c.AmountRial, &issue, &due, &c.PayerPayee, &c.CustomerID, &c.SupplierID, &c.SourceType, &c.SourceID, &c.FinancialAccountID, &c.Notes, &c.Status, &created, &updated); e != nil {
		return c, e
	}
	c.Direction = domain.CheckDirection(direction)
	var e error
	c.IssueDate, e = time.Parse(time.RFC3339Nano, issue)
	if e != nil {
		return c, e
	}
	c.DueDate, e = time.Parse(time.RFC3339Nano, due)
	if e != nil {
		return c, e
	}
	c.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return c, e
	}
	c.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	return c, e
}
func scanLoan(row scanner) (domain.Loan, error) {
	var l domain.Loan
	var start, end, created, updated string
	if e := row.Scan(&l.ID, &l.LoanNumber, &l.Direction, &l.CounterpartyName, &l.CustomerID, &l.SupplierID, &l.PrincipalRial, &l.InterestFeeRial, &start, &end, &l.Status, &l.Notes, &l.FinancialAccountID, &l.JournalEntryID, &l.IdempotencyKey, &created, &updated); e != nil {
		return l, e
	}
	var e error
	l.StartDate, e = time.Parse(time.RFC3339Nano, start)
	if e != nil {
		return l, e
	}
	if end != "" {
		l.EndDate, e = parseNullableTime(end)
		if e != nil {
			return l, e
		}
	}
	l.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return l, e
	}
	l.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	return l, e
}
func scanLoanPayment(row scanner) (domain.LoanPayment, error) {
	var p domain.LoanPayment
	var paid, created string
	if e := row.Scan(&p.ID, &p.PaymentNumber, &p.LoanID, &p.FinancialAccountID, &p.AmountRial, &p.PrincipalRial, &p.InterestRial, &paid, &p.Notes, &p.Status, &p.JournalEntryID, &p.IdempotencyKey, &created); e != nil {
		return p, e
	}
	var e error
	p.PaidAt, e = time.Parse(time.RFC3339Nano, paid)
	if e != nil {
		return p, e
	}
	p.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	return p, e
}
