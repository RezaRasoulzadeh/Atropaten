package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"Atropaten/internal/domain"
)

const ownerSelect = `SELECT id,name,phone,email,notes,active,ownership_bps,profit_sharing_bps,created_at,updated_at FROM owners`
const ownerTxSelect = `SELECT id,transaction_number,owner_id,type,amount_rial,occurred_at,COALESCE(financial_account_id,''),COALESCE(category_account_id,''),description,notes,status,journal_entry_id,idempotency_key,created_at,updated_at FROM owner_transactions`
const periodSelect = `SELECT id,name,start_date,end_date,status,COALESCE(closed_at,''),COALESCE(closing_journal_entry_id,''),idempotency_key,notes,created_at,updated_at FROM fiscal_periods`

func (s *Store) ListOwners(ctx context.Context, activeOnly bool) ([]domain.Owner, error) {
	q := ownerSelect
	if activeOnly {
		q += ` WHERE active=1`
	}
	q += ` ORDER BY name,id`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	var out []domain.Owner
	for rows.Next() {
		v, e := scanOwner(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, v)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	return out, nil
}
func (s *Store) GetOwner(ctx context.Context, id string) (domain.Owner, error) {
	v, err := scanOwner(s.db.QueryRowContext(ctx, ownerSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Owner{}, domain.ErrOwnerNotFound
	}
	return v, err
}
func (s *Store) CreateOwner(ctx context.Context, o domain.Owner) (domain.Owner, error) {
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = o.CreatedAt
	}
	if err := o.Validate(); err != nil {
		return domain.Owner{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Owner{}, err
	}
	fail := func(e error) (domain.Owner, error) { tx.Rollback(); return domain.Owner{}, e }
	var exists string
	if err = tx.QueryRowContext(ctx, `SELECT id FROM owners WHERE id=?`, o.ID).Scan(&exists); err == nil {
		tx.Rollback()
		return s.GetOwner(ctx, o.ID)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fail(err)
	}
	if err = validateOwnerTotalsTx(ctx, tx, o.ID, o.OwnershipBPS, o.ProfitSharingBPS); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO owners(id,name,phone,email,notes,active,ownership_bps,profit_sharing_bps,created_at,updated_at) VALUES(?,?,?,?,?,1,?,?,?,?)`, o.ID, o.Name, o.Phone, o.Email, o.Notes, o.OwnershipBPS, o.ProfitSharingBPS, o.CreatedAt.UTC().Format(time.RFC3339Nano), o.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO owner_share_history(id,owner_id,ownership_bps,profit_sharing_bps,effective_at,reason) VALUES(?,?,?,?,?,?)`, "OSH-"+o.ID+"-"+fmt.Sprint(o.CreatedAt.UnixNano()), o.ID, o.OwnershipBPS, o.ProfitSharingBPS, o.CreatedAt.UTC().Format(time.RFC3339Nano), "Initial shares"); err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.Owner{}, err
	}
	return s.GetOwner(ctx, o.ID)
}
func (s *Store) UpdateOwnerShares(ctx context.Context, id string, ownership, profit int64, reason string) (domain.Owner, error) {
	if err := domain.ValidatePercentageBPS(ownership); err != nil {
		return domain.Owner{}, err
	}
	if err := domain.ValidatePercentageBPS(profit); err != nil {
		return domain.Owner{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Owner{}, err
	}
	fail := func(e error) (domain.Owner, error) { tx.Rollback(); return domain.Owner{}, e }
	var active int
	if err = tx.QueryRowContext(ctx, `SELECT active FROM owners WHERE id=?`, id).Scan(&active); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrOwnerNotFound)
	} else if err != nil {
		return fail(err)
	}
	if active == 0 {
		return fail(domain.ErrOwnerProtected)
	}
	if err = validateOwnerTotalsTx(ctx, tx, id, ownership, profit); err != nil {
		return fail(err)
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE owners SET ownership_bps=?,profit_sharing_bps=?,updated_at=? WHERE id=?`, ownership, profit, now.Format(time.RFC3339Nano), id); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO owner_share_history(id,owner_id,ownership_bps,profit_sharing_bps,effective_at,reason) VALUES(?,?,?,?,?,?)`, "OSH-"+id+"-"+fmt.Sprint(now.UnixNano()), id, ownership, profit, now.Format(time.RFC3339Nano), reason); err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.Owner{}, err
	}
	return s.GetOwner(ctx, id)
}
func (s *Store) ArchiveOwner(ctx context.Context, id string) error {
	return s.setOwnerActive(ctx, id, false)
}
func (s *Store) ReactivateOwner(ctx context.Context, id string) error {
	return s.setOwnerActive(ctx, id, true)
}
func (s *Store) setOwnerActive(ctx context.Context, id string, active bool) error {
	n := 0
	value := 0
	if active {
		value = 1
	}
	res, err := s.db.ExecContext(ctx, `UPDATE owners SET active=?,updated_at=? WHERE id=?`, value, time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return err
	}
	n64, err := res.RowsAffected()
	if err != nil {
		return err
	}
	n = int(n64)
	if n == 0 {
		return domain.ErrOwnerNotFound
	}
	if active {
		owners, err := s.ListOwners(ctx, true)
		if err != nil {
			return err
		}
		var own, profit int64
		for _, o := range owners {
			own += o.OwnershipBPS
			profit += o.ProfitSharingBPS
		}
		if own > domain.PercentageScale || profit > domain.PercentageScale {
			_, _ = s.db.ExecContext(ctx, `UPDATE owners SET active=0 WHERE id=?`, id)
			return domain.ErrOwnerShareTotal
		}
	}
	return nil
}
func (s *Store) DeleteOwner(ctx context.Context, id string) error {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM owner_transactions WHERE owner_id=?`, id).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return domain.ErrOwnerProtected
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM profit_allocations WHERE owner_id=?`, id).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return domain.ErrOwnerProtected
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM owner_share_history WHERE owner_id=?`, id); err != nil {
		tx.Rollback()
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM owners WHERE id=?`, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows == 0 {
		tx.Rollback()
		return domain.ErrOwnerNotFound
	}
	return tx.Commit()
}

func validateOwnerTotalsTx(ctx context.Context, tx *sql.Tx, id string, ownership, profit int64) error {
	var own, total int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(ownership_bps),0),COALESCE(SUM(profit_sharing_bps),0) FROM owners WHERE active=1 AND id<>?`, id).Scan(&own, &total); err != nil {
		return err
	}
	if own+ownership > domain.PercentageScale || total+profit > domain.PercentageScale {
		return domain.ErrOwnerShareTotal
	}
	return nil
}

// OwnerBalances is deliberately derived from journal lines. It returns capital, drawings, current/payable, loan payable, loan receivable, and allocated profit/loss.
func (s *Store) OwnerBalances(ctx context.Context, id string) ([6]int64, error) {
	var out [6]int64
	queries := []struct {
		account        string
		positiveCredit bool
	}{{"ACC-OWNER-CAPITAL", true}, {"ACC-OWNER-DRAWINGS", false}, {"ACC-OWNER-CURRENT", true}, {"ACC-OWNER-LOAN-PAYABLE", true}, {"ACC-OWNER-LOAN-RECEIVABLE", false}}
	for i, q := range queries {
		var debit, credit int64
		if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(debit_rial),0),COALESCE(SUM(credit_rial),0) FROM journal_lines WHERE account_id=? AND party_id=?`, q.account, id).Scan(&debit, &credit); err != nil {
			return out, err
		}
		if q.positiveCredit {
			out[i] = credit - debit
		} else {
			out[i] = debit - credit
		}
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount_rial),0) FROM profit_allocations WHERE owner_id=?`, id).Scan(&out[5]); err != nil {
		return out, err
	}
	return out, nil
}

func (s *Store) ListOwnerTransactions(ctx context.Context, ownerID string) ([]domain.OwnerTransaction, error) {
	q := ownerTxSelect
	args := []any{}
	if ownerID != "" {
		q += ` WHERE owner_id=?`
		args = append(args, ownerID)
	}
	q += ` ORDER BY occurred_at DESC,transaction_number DESC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.OwnerTransaction
	for rows.Next() {
		v, e := scanOwnerTransaction(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) CreateOwnerTransaction(ctx context.Context, t domain.OwnerTransaction) (domain.OwnerTransaction, error) {
	if t.Status == "" {
		t.Status = domain.OwnerTxPosted
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = t.OccurredAt
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = t.CreatedAt
	}
	if t.IdempotencyKey == "" {
		t.IdempotencyKey = t.ID
	}
	if err := t.Validate(); err != nil {
		return domain.OwnerTransaction{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.OwnerTransaction{}, err
	}
	fail := func(e error) (domain.OwnerTransaction, error) { tx.Rollback(); return domain.OwnerTransaction{}, e }
	var existing string
	if err = tx.QueryRowContext(ctx, `SELECT id FROM owner_transactions WHERE idempotency_key=?`, t.IdempotencyKey).Scan(&existing); err == nil {
		tx.Rollback()
		return s.GetOwnerTransaction(ctx, existing)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fail(err)
	}
	var active int
	if err = tx.QueryRowContext(ctx, `SELECT active FROM owners WHERE id=?`, t.OwnerID).Scan(&active); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrOwnerNotFound)
	} else if err != nil {
		return fail(err)
	}
	if active == 0 {
		return fail(domain.ErrOwnerProtected)
	}
	var ledger, typ string
	if t.FinancialAccountID != "" {
		if err = tx.QueryRowContext(ctx, `SELECT active,ledger_account_id,type FROM financial_accounts WHERE id=?`, t.FinancialAccountID).Scan(&active, &ledger, &typ); errors.Is(err, sql.ErrNoRows) {
			return fail(domain.ErrFinancialAccountNotFound)
		} else if err != nil {
			return fail(err)
		}
		if active == 0 {
			return fail(domain.ErrAccountInactive)
		}
		if typ != "cash" && typ != "bank" {
			return fail(fmt.Errorf("owner transaction requires a cash or bank account"))
		}
	}
	if t.Type == domain.OwnerTxPersonalExpense {
		var accountType string
		if err = tx.QueryRowContext(ctx, `SELECT type FROM accounts WHERE id=? AND active=1`, t.CategoryAccountID).Scan(&accountType); err != nil {
			return fail(domain.ErrAccountNotFound)
		}
		if accountType != "expense" {
			return fail(fmt.Errorf("owner-paid expense category must be an expense account"))
		}
	}
	if len(t.Description) == 0 {
		t.Description = "Owner transaction"
	}
	var lines []domain.JournalLine
	je := "JE-OWNER-" + t.ID
	ownerLine := func(account string, debit, credit int64, memo string) domain.JournalLine {
		return domain.JournalLine{AccountID: account, DebitRial: debit, CreditRial: credit, PartyType: "owner", PartyID: t.OwnerID, Memo: memo}
	}
	cashLine := func(debit, credit int64) domain.JournalLine {
		return domain.JournalLine{AccountID: ledger, DebitRial: debit, CreditRial: credit, Memo: "Cash/Bank owner transaction"}
	}
	switch t.Type {
	case domain.OwnerTxCapitalContribution:
		lines = []domain.JournalLine{ownerLine("ACC-OWNER-CAPITAL", 0, t.AmountRial, "Owner capital contribution"), cashLine(t.AmountRial, 0)}
	case domain.OwnerTxDrawing:
		lines = []domain.JournalLine{ownerLine("ACC-OWNER-DRAWINGS", t.AmountRial, 0, "Owner drawing"), cashLine(0, t.AmountRial)}
	case domain.OwnerTxPersonalExpense:
		lines = []domain.JournalLine{ownerLine(t.CategoryAccountID, t.AmountRial, 0, "Business expense paid by owner"), ownerLine("ACC-OWNER-CURRENT", 0, t.AmountRial, "Owner current payable")}
	case domain.OwnerTxReimbursement:
		lines = []domain.JournalLine{ownerLine("ACC-OWNER-CURRENT", t.AmountRial, 0, "Owner reimbursement"), cashLine(0, t.AmountRial)}
	case domain.OwnerTxLoanToBusiness:
		lines = []domain.JournalLine{ownerLine("ACC-OWNER-LOAN-PAYABLE", 0, t.AmountRial, "Owner loan to business"), cashLine(t.AmountRial, 0)}
	case domain.OwnerTxLoanFromBusiness:
		lines = []domain.JournalLine{ownerLine("ACC-OWNER-LOAN-RECEIVABLE", t.AmountRial, 0, "Loan from business to owner"), cashLine(0, t.AmountRial)}
	case domain.OwnerTxLoanRepaymentToOwner:
		lines = []domain.JournalLine{ownerLine("ACC-OWNER-LOAN-PAYABLE", t.AmountRial, 0, "Repay owner loan"), cashLine(0, t.AmountRial)}
	case domain.OwnerTxLoanRepaymentFromOwner:
		lines = []domain.JournalLine{cashLine(t.AmountRial, 0), ownerLine("ACC-OWNER-LOAN-RECEIVABLE", 0, t.AmountRial, "Owner loan repayment")}
	}
	for i := range lines {
		lines[i].ID = je + fmt.Sprintf("-L%d", i+1)
		lines[i].JournalEntryID = je
		lines[i].Position = i
	}
	if _, err = s.postJournalTx(ctx, tx, domain.JournalEntry{ID: je, Description: t.Description, SourceType: "owner_transaction", SourceID: t.ID, IdempotencyKey: "owner:" + t.IdempotencyKey, PostedAt: t.OccurredAt, CreatedAt: t.CreatedAt, Lines: lines}); err != nil {
		return fail(err)
	}
	var n int64
	if err = tx.QueryRowContext(ctx, `SELECT next_number FROM owner_transaction_number_sequences WHERE id=1`).Scan(&n); err != nil {
		return fail(err)
	}
	t.TransactionNumber = fmt.Sprintf("OWN-%04d", n)
	if _, err = tx.ExecContext(ctx, `UPDATE owner_transaction_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO owner_transactions(id,transaction_number,owner_id,type,amount_rial,occurred_at,financial_account_id,category_account_id,description,notes,status,journal_entry_id,idempotency_key,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, t.ID, t.TransactionNumber, t.OwnerID, t.Type, t.AmountRial, t.OccurredAt.UTC().Format(time.RFC3339Nano), nullableString(t.FinancialAccountID), nullableString(t.CategoryAccountID), t.Description, t.Notes, t.Status, je, t.IdempotencyKey, t.CreatedAt.UTC().Format(time.RFC3339Nano), t.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.OwnerTransaction{}, err
	}
	return s.GetOwnerTransaction(ctx, t.ID)
}
func (s *Store) GetOwnerTransaction(ctx context.Context, id string) (domain.OwnerTransaction, error) {
	v, e := scanOwnerTransaction(s.db.QueryRowContext(ctx, ownerTxSelect+` WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.OwnerTransaction{}, domain.ErrOwnerTransactionNotFound
	}
	return v, e
}
func (s *Store) ReverseOwnerTransaction(ctx context.Context, id, key string) (domain.OwnerTransaction, error) {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return domain.OwnerTransaction{}, e
	}
	fail := func(x error) (domain.OwnerTransaction, error) { tx.Rollback(); return domain.OwnerTransaction{}, x }
	var status, je string
	if e = tx.QueryRowContext(ctx, `SELECT status,journal_entry_id FROM owner_transactions WHERE id=?`, id).Scan(&status, &je); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrOwnerTransactionNotFound)
	} else if e != nil {
		return fail(e)
	}
	if status == domain.OwnerTxReversed {
		tx.Rollback()
		return s.GetOwnerTransaction(ctx, id)
	}
	if key == "" {
		key = "owner:reverse:" + id
	}
	if _, e = s.reverseJournalTx(ctx, tx, je, key, "Reverse owner transaction", time.Now().UTC()); e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `UPDATE owner_transactions SET status='Reversed',updated_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), id); e != nil {
		return fail(e)
	}
	if e = tx.Commit(); e != nil {
		return domain.OwnerTransaction{}, e
	}
	return s.GetOwnerTransaction(ctx, id)
}

func (s *Store) ListFiscalPeriods(ctx context.Context) ([]domain.FiscalPeriod, error) {
	rows, e := s.db.QueryContext(ctx, periodSelect+` ORDER BY start_date`)
	if e != nil {
		return nil, e
	}
	var out []domain.FiscalPeriod
	for rows.Next() {
		v, x := scanPeriod(rows)
		if x != nil {
			rows.Close()
			return nil, x
		}
		out = append(out, v)
	}
	if e = rows.Err(); e != nil {
		rows.Close()
		return nil, e
	}
	rows.Close()
	for i := range out {
		if e = s.populatePeriod(ctx, &out[i]); e != nil {
			return nil, e
		}
	}
	return out, nil
}
func (s *Store) GetFiscalPeriod(ctx context.Context, id string) (domain.FiscalPeriod, error) {
	v, e := scanPeriod(s.db.QueryRowContext(ctx, periodSelect+` WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.FiscalPeriod{}, domain.ErrPeriodNotFound
	}
	if e != nil {
		return v, e
	}
	return v, s.populatePeriod(ctx, &v)
}
func (s *Store) CreateFiscalPeriod(ctx context.Context, p domain.FiscalPeriod) (domain.FiscalPeriod, error) {
	if p.Status == "" {
		p.Status = domain.FiscalPeriodOpen
	}
	if p.Status != domain.FiscalPeriodOpen {
		return domain.FiscalPeriod{}, domain.ErrPeriodState
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = p.CreatedAt
	}
	if p.IdempotencyKey == "" {
		p.IdempotencyKey = p.ID
	}
	if err := p.Validate(); err != nil {
		return domain.FiscalPeriod{}, err
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return domain.FiscalPeriod{}, e
	}
	fail := func(x error) (domain.FiscalPeriod, error) { tx.Rollback(); return domain.FiscalPeriod{}, x }
	var exists string
	if e = tx.QueryRowContext(ctx, `SELECT id FROM fiscal_periods WHERE idempotency_key=?`, p.IdempotencyKey).Scan(&exists); e == nil {
		tx.Rollback()
		return s.GetFiscalPeriod(ctx, exists)
	} else if !errors.Is(e, sql.ErrNoRows) {
		return fail(e)
	}
	var n int
	if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM fiscal_periods WHERE NOT (end_date < ? OR start_date > ?)`, p.StartDate.UTC().Format(time.RFC3339Nano), p.EndDate.UTC().Format(time.RFC3339Nano)).Scan(&n); e != nil {
		return fail(e)
	}
	if n > 0 {
		return fail(domain.ErrPeriodOverlap)
	}
	if _, e = tx.ExecContext(ctx, `INSERT INTO fiscal_periods(id,name,start_date,end_date,status,idempotency_key,notes,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, p.ID, p.Name, p.StartDate.UTC().Format(time.RFC3339Nano), p.EndDate.UTC().Format(time.RFC3339Nano), p.Status, p.IdempotencyKey, p.Notes, p.CreatedAt.UTC().Format(time.RFC3339Nano), p.UpdatedAt.UTC().Format(time.RFC3339Nano)); e != nil {
		return fail(e)
	}
	if e = tx.Commit(); e != nil {
		return domain.FiscalPeriod{}, e
	}
	return s.GetFiscalPeriod(ctx, p.ID)
}
func (s *Store) CloseFiscalPeriod(ctx context.Context, id, key string) (domain.FiscalPeriod, error) {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return domain.FiscalPeriod{}, e
	}
	fail := func(x error) (domain.FiscalPeriod, error) { tx.Rollback(); return domain.FiscalPeriod{}, x }
	var p domain.FiscalPeriod
	var start, end, closed, closing, created, updated string
	if e = tx.QueryRowContext(ctx, periodSelect+` WHERE id=?`, id).Scan(&p.ID, &p.Name, &start, &end, &p.Status, &closed, &closing, &p.IdempotencyKey, &p.Notes, &created, &updated); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrPeriodNotFound)
	} else if e != nil {
		return fail(e)
	}
	p.StartDate, e = time.Parse(time.RFC3339Nano, start)
	if e != nil {
		return fail(e)
	}
	p.EndDate, e = time.Parse(time.RFC3339Nano, end)
	if e != nil {
		return fail(e)
	}
	p.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return fail(e)
	}
	p.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	if e != nil {
		return fail(e)
	}
	if closed != "" {
		v, _ := time.Parse(time.RFC3339Nano, closed)
		p.ClosedAt = &v
	}
	p.ClosingJournalEntryID = closing
	if p.Status == domain.FiscalPeriodClosed {
		tx.Rollback()
		return s.GetFiscalPeriod(ctx, id)
	}
	if p.Status != domain.FiscalPeriodOpen {
		return fail(domain.ErrPeriodState)
	}
	owners, e := ownersTx(ctx, tx)
	if e != nil {
		return fail(e)
	}
	sort.Slice(owners, func(i, j int) bool { return owners[i].ID < owners[j].ID })
	var totalShares int64
	for _, o := range owners {
		totalShares += o.ProfitSharingBPS
	}
	if totalShares != domain.PercentageScale {
		return fail(domain.ErrPeriodOwnerShares)
	}
	revenue, cogs, expenses, profit, e := periodProfitTx(ctx, tx, p.StartDate, p.EndDate)
	if e != nil {
		return fail(e)
	}
	p.RevenueRial, p.COGSRial, p.ExpensesRial, p.ProfitLossRial = revenue, cogs, expenses, profit
	var debit, credit int64
	if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(debit_rial),0),COALESCE(SUM(credit_rial),0) FROM journal_lines`).Scan(&debit, &credit); e != nil {
		return fail(e)
	}
	if debit != credit {
		return fail(domain.ErrPeriodUnbalancedLedger)
	}
	allocations, e := domain.AllocateProfitLoss(profit, owners)
	if e != nil {
		return fail(e)
	}
	now := time.Now().UTC()
	if key == "" {
		key = "period:close:" + id
	}
	if profit != 0 {
		je := "JE-CLOSE-" + id
		lines := []domain.JournalLine{}
		for _, a := range allocations {
			if a.AmountRial == 0 {
				continue
			}
			if profit > 0 {
				lines = append(lines, domain.JournalLine{AccountID: "ACC-RETAINED-EARNINGS", DebitRial: abs64(a.AmountRial), PartyType: "owner", PartyID: a.OwnerID, Memo: "Allocate period profit"})
				lines = append(lines, domain.JournalLine{AccountID: "ACC-OWNER-CURRENT", CreditRial: abs64(a.AmountRial), PartyType: "owner", PartyID: a.OwnerID, Memo: "Owner profit allocation"})
			} else {
				lines = append(lines, domain.JournalLine{AccountID: "ACC-OWNER-CURRENT", DebitRial: abs64(a.AmountRial), PartyType: "owner", PartyID: a.OwnerID, Memo: "Allocate period loss"})
				lines = append(lines, domain.JournalLine{AccountID: "ACC-RETAINED-EARNINGS", CreditRial: abs64(a.AmountRial), PartyType: "owner", PartyID: a.OwnerID, Memo: "Owner loss allocation"})
			}
		}
		for i := range lines {
			lines[i].ID = je + fmt.Sprintf("-L%d", i+1)
			lines[i].JournalEntryID = je
			lines[i].Position = i
		}
		if _, e = s.postJournalTxWithPeriod(ctx, tx, domain.JournalEntry{ID: je, Description: "Close fiscal period " + p.Name, SourceType: "profit_allocation", SourceID: p.ID, IdempotencyKey: key, PostedAt: p.EndDate, CreatedAt: now, Lines: lines}, true); e != nil {
			return fail(e)
		}
		p.ClosingJournalEntryID = je
	}
	for _, a := range allocations {
		a.ID = "PAL-" + p.ID + "-" + a.OwnerID
		if _, e = tx.ExecContext(ctx, `INSERT INTO profit_allocations(id,period_id,owner_id,position,profit_sharing_bps,amount_rial) VALUES(?,?,?,?,?,?)`, a.ID, p.ID, a.OwnerID, a.Position, a.ProfitSharingBPS, a.AmountRial); e != nil {
			return fail(e)
		}
	}
	if _, e = tx.ExecContext(ctx, `UPDATE fiscal_periods SET status='Closed',closed_at=?,closing_journal_entry_id=?,updated_at=? WHERE id=? AND status='Open'`, now.Format(time.RFC3339Nano), nullableString(p.ClosingJournalEntryID), now.Format(time.RFC3339Nano), id); e != nil {
		return fail(e)
	}
	if e = tx.Commit(); e != nil {
		return domain.FiscalPeriod{}, e
	}
	return s.GetFiscalPeriod(ctx, id)
}

func (s *Store) populatePeriod(ctx context.Context, p *domain.FiscalPeriod) error {
	var err error
	p.RevenueRial, p.COGSRial, p.ExpensesRial, p.ProfitLossRial, err = periodProfitDB(ctx, s.db, p.StartDate, p.EndDate)
	if err != nil {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,period_id,owner_id,position,profit_sharing_bps,amount_rial FROM profit_allocations WHERE period_id=? ORDER BY position`, p.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var a domain.ProfitAllocation
		if err := rows.Scan(&a.ID, &a.PeriodID, &a.OwnerID, &a.Position, &a.ProfitSharingBPS, &a.AmountRial); err != nil {
			return err
		}
		p.Allocations = append(p.Allocations, a)
	}
	return rows.Err()
}
func periodProfitTx(ctx context.Context, tx *sql.Tx, start, end time.Time) (int64, int64, int64, int64, error) {
	endExclusive := end.UTC().Add(24 * time.Hour)
	var revenue, cogs, expenses int64
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN a.type='revenue' THEN jl.credit_rial-jl.debit_rial ELSE 0 END),0),COALESCE(SUM(CASE WHEN a.id='ACC-COGS' THEN jl.debit_rial-jl.credit_rial ELSE 0 END),0),COALESCE(SUM(CASE WHEN a.type='expense' AND a.id<>'ACC-COGS' THEN jl.debit_rial-jl.credit_rial ELSE 0 END),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id JOIN accounts a ON a.id=jl.account_id WHERE je.posted_at>=? AND je.posted_at<? AND je.source_type<>'profit_allocation'`, start.UTC().Format(time.RFC3339Nano), endExclusive.Format(time.RFC3339Nano)).Scan(&revenue, &cogs, &expenses)
	return revenue, cogs, expenses, revenue - cogs - expenses, err
}
func periodProfitDB(ctx context.Context, db *sql.DB, start, end time.Time) (int64, int64, int64, int64, error) {
	endExclusive := end.UTC().Add(24 * time.Hour)
	var revenue, cogs, expenses int64
	err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN a.type='revenue' THEN jl.credit_rial-jl.debit_rial ELSE 0 END),0),COALESCE(SUM(CASE WHEN a.id='ACC-COGS' THEN jl.debit_rial-jl.credit_rial ELSE 0 END),0),COALESCE(SUM(CASE WHEN a.type='expense' AND a.id<>'ACC-COGS' THEN jl.debit_rial-jl.credit_rial ELSE 0 END),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id JOIN accounts a ON a.id=jl.account_id WHERE je.posted_at>=? AND je.posted_at<? AND je.source_type<>'profit_allocation'`, start.UTC().Format(time.RFC3339Nano), endExclusive.Format(time.RFC3339Nano)).Scan(&revenue, &cogs, &expenses)
	return revenue, cogs, expenses, revenue - cogs - expenses, err
}
func ownersTx(ctx context.Context, tx *sql.Tx) ([]domain.Owner, error) {
	rows, e := tx.QueryContext(ctx, ownerSelect+` WHERE active=1`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.Owner
	for rows.Next() {
		v, x := scanOwner(rows)
		if x != nil {
			return nil, x
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func scanOwner(row scanner) (domain.Owner, error) {
	var o domain.Owner
	var active int
	var created, updated string
	if e := row.Scan(&o.ID, &o.Name, &o.Phone, &o.Email, &o.Notes, &active, &o.OwnershipBPS, &o.ProfitSharingBPS, &created, &updated); e != nil {
		return o, e
	}
	o.Active = active != 0
	var e error
	o.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return o, e
	}
	o.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	return o, e
}
func scanOwnerTransaction(row scanner) (domain.OwnerTransaction, error) {
	var t domain.OwnerTransaction
	var occurred, created, updated string
	if e := row.Scan(&t.ID, &t.TransactionNumber, &t.OwnerID, &t.Type, &t.AmountRial, &occurred, &t.FinancialAccountID, &t.CategoryAccountID, &t.Description, &t.Notes, &t.Status, &t.JournalEntryID, &t.IdempotencyKey, &created, &updated); e != nil {
		return t, e
	}
	var e error
	t.OccurredAt, e = time.Parse(time.RFC3339Nano, occurred)
	if e != nil {
		return t, e
	}
	t.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return t, e
	}
	t.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	return t, e
}
func scanPeriod(row scanner) (domain.FiscalPeriod, error) {
	var p domain.FiscalPeriod
	var start, end, closed, closing, created, updated string
	if e := row.Scan(&p.ID, &p.Name, &start, &end, &p.Status, &closed, &closing, &p.IdempotencyKey, &p.Notes, &created, &updated); e != nil {
		return p, e
	}
	var e error
	p.StartDate, e = time.Parse(time.RFC3339Nano, start)
	if e != nil {
		return p, e
	}
	p.EndDate, e = time.Parse(time.RFC3339Nano, end)
	if e != nil {
		return p, e
	}
	if closed != "" {
		v, x := time.Parse(time.RFC3339Nano, closed)
		if x != nil {
			return p, x
		}
		p.ClosedAt = &v
	}
	p.ClosingJournalEntryID = closing
	p.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return p, e
	}
	p.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	return p, e
}
func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
