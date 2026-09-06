package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"Atropaten/internal/domain"
)

func (s *Store) ListExpenses(ctx context.Context) ([]domain.Expense, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,expense_number,expense_date,category_account_id,COALESCE(payee,''),COALESCE(supplier_id,''),description,amount_rial,payment_method,financial_account_id,notes,status,journal_entry_id,idempotency_key,created_at,updated_at FROM expenses ORDER BY expense_date DESC,expense_number DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Expense
	for rows.Next() {
		v, e := scanExpense(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) GetExpense(ctx context.Context, id string) (domain.Expense, error) {
	v, e := scanExpense(s.db.QueryRowContext(ctx, `SELECT id,expense_number,expense_date,category_account_id,COALESCE(payee,''),COALESCE(supplier_id,''),description,amount_rial,payment_method,financial_account_id,notes,status,journal_entry_id,idempotency_key,created_at,updated_at FROM expenses WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.Expense{}, domain.ErrExpenseNotFound
	}
	return v, e
}
func (s *Store) CreateExpense(ctx context.Context, v domain.Expense) (domain.Expense, error) {
	if v.Status == "" {
		v.Status = "Posted"
	}
	if v.UpdatedAt.IsZero() {
		v.UpdatedAt = time.Now().UTC()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = v.UpdatedAt
	}
	if v.ExpenseDate.IsZero() {
		v.ExpenseDate = v.CreatedAt
	}
	if v.IdempotencyKey == "" {
		v.IdempotencyKey = v.ID
	}
	if err := v.Validate(); err != nil {
		return domain.Expense{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Expense{}, err
	}
	fail := func(e error) (domain.Expense, error) { tx.Rollback(); return domain.Expense{}, e }
	var existing string
	if err = tx.QueryRowContext(ctx, `SELECT id FROM expenses WHERE idempotency_key=?`, v.IdempotencyKey).Scan(&existing); err == nil {
		tx.Rollback()
		return s.GetExpense(ctx, existing)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fail(err)
	}
	var active int
	var ledger string
	if err = tx.QueryRowContext(ctx, `SELECT active,ledger_account_id FROM financial_accounts WHERE id=?`, v.FinancialAccountID).Scan(&active, &ledger); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrFinancialAccountNotFound)
	} else if err != nil {
		return fail(err)
	}
	if active == 0 {
		return fail(domain.ErrAccountInactive)
	}
	var typ string
	if err = tx.QueryRowContext(ctx, `SELECT type FROM accounts WHERE id=? AND active=1`, v.CategoryAccountID).Scan(&typ); err != nil {
		return fail(domain.ErrAccountNotFound)
	}
	if typ != "expense" {
		return fail(fmt.Errorf("expense category must be an expense account"))
	}
	je := "JE-EXP-" + v.ID
	entry := domain.JournalEntry{ID: je, Description: v.Description, SourceType: "expense", SourceID: v.ID, IdempotencyKey: "expense:" + v.IdempotencyKey, PostedAt: v.ExpenseDate, CreatedAt: v.CreatedAt, Lines: []domain.JournalLine{{ID: je + "-L1", JournalEntryID: je, Position: 0, AccountID: v.CategoryAccountID, DebitRial: v.AmountRial, PartyType: "supplier", PartyID: v.SupplierID, Memo: v.Description}, {ID: je + "-L2", JournalEntryID: je, Position: 1, AccountID: ledger, CreditRial: v.AmountRial, Memo: "Cash/Bank expense payment"}}}
	if _, err = s.postJournalTx(ctx, tx, entry); err != nil {
		return fail(err)
	}
	var n int64
	if err = tx.QueryRowContext(ctx, `SELECT next_number FROM expense_number_sequences WHERE id=1`).Scan(&n); err != nil {
		return fail(err)
	}
	v.ExpenseNumber = fmt.Sprintf("EXP-%04d", n)
	if _, err = tx.ExecContext(ctx, `UPDATE expense_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO expenses(id,expense_number,expense_date,category_account_id,payee,supplier_id,description,amount_rial,payment_method,financial_account_id,notes,status,journal_entry_id,idempotency_key,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.ExpenseNumber, v.ExpenseDate.UTC().Format(time.RFC3339Nano), v.CategoryAccountID, v.Payee, nullableString(v.SupplierID), v.Description, v.AmountRial, v.PaymentMethod, v.FinancialAccountID, v.Notes, v.Status, je, v.IdempotencyKey, v.CreatedAt.UTC().Format(time.RFC3339Nano), v.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.Expense{}, err
	}
	return s.GetExpense(ctx, v.ID)
}
func (s *Store) ReverseExpense(ctx context.Context, id, key string) (domain.Expense, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Expense{}, err
	}
	fail := func(e error) (domain.Expense, error) { tx.Rollback(); return domain.Expense{}, e }
	var status, je string
	if err = tx.QueryRowContext(ctx, `SELECT status,journal_entry_id FROM expenses WHERE id=?`, id).Scan(&status, &je); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrExpenseNotFound)
	} else if err != nil {
		return fail(err)
	}
	if status == "Reversed" {
		tx.Rollback()
		return s.GetExpense(ctx, id)
	}
	if key == "" {
		key = "expense:reverse:" + id
	}
	if _, err = s.reverseJournalTx(ctx, tx, je, key, "Reversal of expense "+id, time.Now().UTC()); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE expenses SET status='Reversed',updated_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), id); err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.Expense{}, err
	}
	return s.GetExpense(ctx, id)
}

func (s *Store) ListTransfers(ctx context.Context) ([]domain.FinancialTransfer, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,transfer_number,source_financial_account_id,destination_financial_account_id,amount_rial,transfer_date,reference,notes,status,journal_entry_id,idempotency_key,created_at,updated_at FROM financial_transfers ORDER BY transfer_date DESC,transfer_number DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.FinancialTransfer
	for rows.Next() {
		v, e := scanTransfer(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) GetTransfer(ctx context.Context, id string) (domain.FinancialTransfer, error) {
	v, e := scanTransfer(s.db.QueryRowContext(ctx, `SELECT id,transfer_number,source_financial_account_id,destination_financial_account_id,amount_rial,transfer_date,reference,notes,status,journal_entry_id,idempotency_key,created_at,updated_at FROM financial_transfers WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.FinancialTransfer{}, domain.ErrTransferNotFound
	}
	return v, e
}
func (s *Store) CreateTransfer(ctx context.Context, v domain.FinancialTransfer) (domain.FinancialTransfer, error) {
	if v.Status == "" {
		v.Status = "Posted"
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = v.TransferDate
	}
	if v.UpdatedAt.IsZero() {
		v.UpdatedAt = v.CreatedAt
	}
	if v.IdempotencyKey == "" {
		v.IdempotencyKey = v.ID
	}
	if err := v.Validate(); err != nil {
		return domain.FinancialTransfer{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.FinancialTransfer{}, err
	}
	fail := func(e error) (domain.FinancialTransfer, error) { tx.Rollback(); return domain.FinancialTransfer{}, e }
	var existing string
	if err = tx.QueryRowContext(ctx, `SELECT id FROM financial_transfers WHERE idempotency_key=?`, v.IdempotencyKey).Scan(&existing); err == nil {
		tx.Rollback()
		return s.GetTransfer(ctx, existing)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fail(err)
	}
	var source, dest string
	var active int
	if err = tx.QueryRowContext(ctx, `SELECT active,ledger_account_id FROM financial_accounts WHERE id=?`, v.SourceFinancialAccountID).Scan(&active, &source); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrFinancialAccountNotFound)
	} else if err != nil {
		return fail(err)
	}
	if active == 0 {
		return fail(domain.ErrAccountInactive)
	}
	if err = tx.QueryRowContext(ctx, `SELECT active,ledger_account_id FROM financial_accounts WHERE id=?`, v.DestinationFinancialAccountID).Scan(&active, &dest); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrFinancialAccountNotFound)
	} else if err != nil {
		return fail(err)
	}
	if active == 0 {
		return fail(domain.ErrAccountInactive)
	}
	je := "JE-TRF-" + v.ID
	entry := domain.JournalEntry{ID: je, Description: "Transfer " + v.Reference, SourceType: "transfer", SourceID: v.ID, IdempotencyKey: "transfer:" + v.IdempotencyKey, PostedAt: v.TransferDate, CreatedAt: v.CreatedAt, Lines: []domain.JournalLine{{ID: je + "-L1", JournalEntryID: je, Position: 0, AccountID: dest, DebitRial: v.AmountRial, Memo: "Transfer destination"}, {ID: je + "-L2", JournalEntryID: je, Position: 1, AccountID: source, CreditRial: v.AmountRial, Memo: "Transfer source"}}}
	if _, err = s.postJournalTx(ctx, tx, entry); err != nil {
		return fail(err)
	}
	var n int64
	if err = tx.QueryRowContext(ctx, `SELECT next_number FROM transfer_number_sequences WHERE id=1`).Scan(&n); err != nil {
		return fail(err)
	}
	v.TransferNumber = fmt.Sprintf("TRF-%04d", n)
	if _, err = tx.ExecContext(ctx, `UPDATE transfer_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO financial_transfers(id,transfer_number,source_financial_account_id,destination_financial_account_id,amount_rial,transfer_date,reference,notes,status,journal_entry_id,idempotency_key,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.TransferNumber, v.SourceFinancialAccountID, v.DestinationFinancialAccountID, v.AmountRial, v.TransferDate.UTC().Format(time.RFC3339Nano), v.Reference, v.Notes, v.Status, je, v.IdempotencyKey, v.CreatedAt.UTC().Format(time.RFC3339Nano), v.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.FinancialTransfer{}, err
	}
	return s.GetTransfer(ctx, v.ID)
}
func (s *Store) ReverseTransfer(ctx context.Context, id, key string) (domain.FinancialTransfer, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.FinancialTransfer{}, err
	}
	fail := func(e error) (domain.FinancialTransfer, error) { tx.Rollback(); return domain.FinancialTransfer{}, e }
	var status, je string
	if err = tx.QueryRowContext(ctx, `SELECT status,journal_entry_id FROM financial_transfers WHERE id=?`, id).Scan(&status, &je); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrTransferNotFound)
	} else if err != nil {
		return fail(err)
	}
	if status == "Reversed" {
		tx.Rollback()
		return s.GetTransfer(ctx, id)
	}
	if key == "" {
		key = "transfer:reverse:" + id
	}
	if _, err = s.reverseJournalTx(ctx, tx, je, key, "Reversal of transfer "+id, time.Now().UTC()); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE financial_transfers SET status='Reversed',updated_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), id); err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return domain.FinancialTransfer{}, err
	}
	return s.GetTransfer(ctx, id)
}

func scanExpense(row scanner) (domain.Expense, error) {
	var v domain.Expense
	var date, created, updated string
	if err := row.Scan(&v.ID, &v.ExpenseNumber, &date, &v.CategoryAccountID, &v.Payee, &v.SupplierID, &v.Description, &v.AmountRial, &v.PaymentMethod, &v.FinancialAccountID, &v.Notes, &v.Status, &v.JournalEntryID, &v.IdempotencyKey, &created, &updated); err != nil {
		return v, err
	}
	var err error
	v.ExpenseDate, err = time.Parse(time.RFC3339Nano, date)
	if err != nil {
		return v, err
	}
	v.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return v, err
	}
	v.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	return v, err
}
func scanTransfer(row scanner) (domain.FinancialTransfer, error) {
	var v domain.FinancialTransfer
	var date, created, updated string
	if err := row.Scan(&v.ID, &v.TransferNumber, &v.SourceFinancialAccountID, &v.DestinationFinancialAccountID, &v.AmountRial, &date, &v.Reference, &v.Notes, &v.Status, &v.JournalEntryID, &v.IdempotencyKey, &created, &updated); err != nil {
		return v, err
	}
	var err error
	v.TransferDate, err = time.Parse(time.RFC3339Nano, date)
	if err != nil {
		return v, err
	}
	v.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return v, err
	}
	v.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	return v, err
}
