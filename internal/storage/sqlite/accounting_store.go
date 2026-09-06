package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

func (s *Store) ListAccounts(ctx context.Context) ([]domain.Account, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,code,name,type,COALESCE(parent_id,''),active,system,created_at,updated_at FROM accounts ORDER BY code`)
	if err != nil {
		return nil, err
	}
	var out []domain.Account
	for rows.Next() {
		v, e := scanAccount(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, v)
	}
	if e := rows.Err(); e != nil {
		rows.Close()
		return nil, e
	}
	rows.Close()
	for i := range out {
		out[i].BalanceRial, err = s.accountBalance(ctx, out[i].ID, out[i].Type)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Store) GetAccount(ctx context.Context, id string) (domain.Account, error) {
	v, err := scanAccount(s.db.QueryRowContext(ctx, `SELECT id,code,name,type,COALESCE(parent_id,''),active,system,created_at,updated_at FROM accounts WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	if err != nil {
		return v, err
	}
	v.BalanceRial, err = s.accountBalance(ctx, v.ID, v.Type)
	return v, err
}

func (s *Store) accountBalance(ctx context.Context, id string, typ domain.AccountType) (int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT debit_rial,credit_rial FROM journal_lines WHERE account_id=?`, id)
	if err != nil {
		return 0, err
	}
	var raw big.Int
	for rows.Next() {
		var debit, credit int64
		if err := rows.Scan(&debit, &credit); err != nil {
			rows.Close()
			return 0, err
		}
		raw.Add(&raw, big.NewInt(debit))
		raw.Sub(&raw, big.NewInt(credit))
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()
	if typ == domain.AccountLiability || typ == domain.AccountEquity || typ == domain.AccountRevenue {
		raw.Neg(&raw)
	}
	if !raw.IsInt64() {
		return 0, fmt.Errorf("account balance is too large")
	}
	return raw.Int64(), nil
}

func (s *Store) ListJournalEntries(ctx context.Context) ([]domain.JournalEntry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,entry_number,posted_at,description,source_type,source_id,idempotency_key,COALESCE(reversal_of_id,''),created_at FROM journal_entries ORDER BY posted_at DESC,entry_number DESC`)
	if err != nil {
		return nil, err
	}
	var ids []string
	var out []domain.JournalEntry
	for rows.Next() {
		v, e := scanJournalEntry(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		ids = append(ids, v.ID)
		out = append(out, v)
	}
	if e := rows.Err(); e != nil {
		rows.Close()
		return nil, e
	}
	rows.Close()
	for i := range out {
		out[i].Lines, err = s.loadJournalLines(ctx, ids[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Store) GetJournalEntry(ctx context.Context, id string) (domain.JournalEntry, error) {
	v, err := scanJournalEntry(s.db.QueryRowContext(ctx, `SELECT id,entry_number,posted_at,description,source_type,source_id,idempotency_key,COALESCE(reversal_of_id,''),created_at FROM journal_entries WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.JournalEntry{}, domain.ErrJournalEntryNotFound
	}
	if err != nil {
		return v, err
	}
	v.Lines, err = s.loadJournalLines(ctx, id)
	return v, err
}

func (s *Store) loadJournalLines(ctx context.Context, id string) ([]domain.JournalLine, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,journal_entry_id,position,account_id,debit_rial,credit_rial,party_type,party_id,memo FROM journal_lines WHERE journal_entry_id=? ORDER BY position`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.JournalLine
	for rows.Next() {
		var v domain.JournalLine
		if err := rows.Scan(&v.ID, &v.JournalEntryID, &v.Position, &v.AccountID, &v.DebitRial, &v.CreditRial, &v.PartyType, &v.PartyID, &v.Memo); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) PostJournalEntry(ctx context.Context, entry domain.JournalEntry) (domain.JournalEntry, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.JournalEntry{}, err
	}
	v, err := s.postJournalTx(ctx, tx, entry)
	if err != nil {
		tx.Rollback()
		return domain.JournalEntry{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.JournalEntry{}, err
	}
	return v, nil
}

func (s *Store) postJournalTx(ctx context.Context, tx *sql.Tx, entry domain.JournalEntry) (domain.JournalEntry, error) {
	if err := entry.Validate(); err != nil {
		return domain.JournalEntry{}, err
	}
	if strings.TrimSpace(entry.IdempotencyKey) == "" {
		return domain.JournalEntry{}, fmt.Errorf("journal idempotency key is required")
	}
	var existingID string
	err := tx.QueryRowContext(ctx, `SELECT id FROM journal_entries WHERE idempotency_key=?`, entry.IdempotencyKey).Scan(&existingID)
	if err == nil {
		return scanJournalTx(ctx, tx, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domain.JournalEntry{}, err
	}
	var next int64
	if err = tx.QueryRowContext(ctx, `SELECT next_number FROM journal_number_sequences WHERE id=1`).Scan(&next); err != nil {
		return domain.JournalEntry{}, err
	}
	entry.EntryNumber = fmt.Sprintf("JE-%04d", next)
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = entry.PostedAt
	}
	_, err = tx.ExecContext(ctx, `UPDATE journal_number_sequences SET next_number=next_number+1 WHERE id=1`)
	if err != nil {
		return domain.JournalEntry{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO journal_entries(id,entry_number,posted_at,description,source_type,source_id,idempotency_key,reversal_of_id,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, entry.ID, entry.EntryNumber, entry.PostedAt.UTC().Format(time.RFC3339Nano), entry.Description, entry.SourceType, entry.SourceID, entry.IdempotencyKey, nullableString(entry.ReversalOfID), entry.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return domain.JournalEntry{}, err
	}
	for _, line := range entry.Lines {
		var active int
		if err = tx.QueryRowContext(ctx, `SELECT active FROM accounts WHERE id=?`, line.AccountID).Scan(&active); errors.Is(err, sql.ErrNoRows) {
			return domain.JournalEntry{}, domain.ErrAccountNotFound
		}
		if err != nil {
			return domain.JournalEntry{}, err
		}
		if active == 0 {
			return domain.JournalEntry{}, domain.ErrAccountInactive
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO journal_lines(id,journal_entry_id,position,account_id,debit_rial,credit_rial,party_type,party_id,memo) VALUES(?,?,?,?,?,?,?,?,?)`, line.ID, entry.ID, line.Position, line.AccountID, line.DebitRial, line.CreditRial, line.PartyType, line.PartyID, line.Memo)
		if err != nil {
			return domain.JournalEntry{}, err
		}
	}
	return entry, nil
}

func (s *Store) reverseJournalTx(ctx context.Context, tx *sql.Tx, id, key, description string, postedAt time.Time) (domain.JournalEntry, error) {
	original, err := scanJournalTx(ctx, tx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.JournalEntry{}, domain.ErrJournalEntryNotFound
		}
		return domain.JournalEntry{}, err
	}
	if description == "" {
		description = "Reversal of " + original.EntryNumber
	}
	return s.postJournalTx(ctx, tx, original.Reversal("REV-"+id, key, description, postedAt.UTC()))
}

func scanJournalTx(ctx context.Context, tx *sql.Tx, id string) (domain.JournalEntry, error) {
	var v domain.JournalEntry
	var posted, created string
	err := tx.QueryRowContext(ctx, `SELECT id,entry_number,posted_at,description,source_type,source_id,idempotency_key,COALESCE(reversal_of_id,''),created_at FROM journal_entries WHERE id=?`, id).Scan(&v.ID, &v.EntryNumber, &posted, &v.Description, &v.SourceType, &v.SourceID, &v.IdempotencyKey, &v.ReversalOfID, &created)
	if err != nil {
		return v, err
	}
	v.PostedAt, err = time.Parse(time.RFC3339Nano, posted)
	if err != nil {
		return v, err
	}
	v.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return v, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,journal_entry_id,position,account_id,debit_rial,credit_rial,party_type,party_id,memo FROM journal_lines WHERE journal_entry_id=? ORDER BY position`, id)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	for rows.Next() {
		var l domain.JournalLine
		if err = rows.Scan(&l.ID, &l.JournalEntryID, &l.Position, &l.AccountID, &l.DebitRial, &l.CreditRial, &l.PartyType, &l.PartyID, &l.Memo); err != nil {
			return v, err
		}
		v.Lines = append(v.Lines, l)
	}
	return v, rows.Err()
}

func (s *Store) ReverseJournalEntry(ctx context.Context, id, key, description string, postedAt time.Time) (domain.JournalEntry, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.JournalEntry{}, err
	}
	out, err := s.reverseJournalTx(ctx, tx, id, key, description, postedAt)
	if err != nil {
		tx.Rollback()
		return domain.JournalEntry{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.JournalEntry{}, err
	}
	return out, nil
}

func (s *Store) ListFinancialAccounts(ctx context.Context) ([]domain.FinancialAccount, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,type,ledger_account_id,details,active,created_at,updated_at FROM financial_accounts ORDER BY name,id`)
	if err != nil {
		return nil, err
	}
	var out []domain.FinancialAccount
	for rows.Next() {
		v, e := scanFinancialAccount(rows)
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
	rows.Close()
	for i := range out {
		out[i].BalanceRial, err = s.accountBalance(ctx, out[i].LedgerAccountID, domain.AccountAsset)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Store) ListPayments(ctx context.Context) ([]domain.Payment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,payment_number,direction,method,amount_rial,posted_at,financial_account_id,COALESCE(customer_id,''),COALESCE(supplier_id,''),reference,notes,status,journal_entry_id,idempotency_key,created_at FROM payments ORDER BY posted_at DESC,payment_number DESC`)
	if err != nil {
		return nil, err
	}
	var out []domain.Payment
	for rows.Next() {
		v, e := scanPayment(rows)
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
	rows.Close()
	for i := range out {
		out[i].Allocations, err = s.loadAllocations(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Store) GetPayment(ctx context.Context, id string) (domain.Payment, error) {
	v, err := scanPayment(s.db.QueryRowContext(ctx, `SELECT id,payment_number,direction,method,amount_rial,posted_at,financial_account_id,COALESCE(customer_id,''),COALESCE(supplier_id,''),reference,notes,status,journal_entry_id,idempotency_key,created_at FROM payments WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, domain.ErrPaymentNotFound
	}
	if err != nil {
		return v, err
	}
	v.Allocations, err = s.loadAllocations(ctx, id)
	return v, err
}
func (s *Store) loadAllocations(ctx context.Context, id string) ([]domain.PaymentAllocation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,payment_id,position,target_type,target_id,amount_rial,reversed FROM payment_allocations WHERE payment_id=? ORDER BY position`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PaymentAllocation
	for rows.Next() {
		var a domain.PaymentAllocation
		var rev int
		if err = rows.Scan(&a.ID, &a.PaymentID, &a.Position, &a.TargetType, &a.TargetID, &a.AmountRial, &rev); err != nil {
			return nil, err
		}
		a.Reversed = rev == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) CreatePayment(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Payment{}, err
	}
	if p.Status == "" {
		p.Status = domain.PaymentPosted
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = p.PostedAt
	}
	for i := range p.Allocations {
		p.Allocations[i].Position = i
		if p.Allocations[i].PaymentID == "" {
			p.Allocations[i].PaymentID = p.ID
		}
		if p.Allocations[i].ID == "" {
			p.Allocations[i].ID = fmt.Sprintf("%s-AL-%d", p.ID, i+1)
		}
	}
	if p.IdempotencyKey == "" {
		p.IdempotencyKey = p.ID
	}
	if err = p.Validate(); err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	var existing string
	err = tx.QueryRowContext(ctx, `SELECT id FROM payments WHERE idempotency_key=?`, p.IdempotencyKey).Scan(&existing)
	if err == nil {
		tx.Rollback()
		return s.GetPayment(ctx, existing)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return domain.Payment{}, err
	}
	var active int
	var ledger string
	if err = tx.QueryRowContext(ctx, `SELECT active,ledger_account_id FROM financial_accounts WHERE id=?`, p.FinancialAccountID).Scan(&active, &ledger); errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return domain.Payment{}, domain.ErrFinancialAccountNotFound
	}
	if err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	if active == 0 {
		tx.Rollback()
		return domain.Payment{}, domain.ErrAccountInactive
	}
	var allocTotal int64
	allocatedByTarget := map[string]int64{}
	for i := range p.Allocations {
		a := &p.Allocations[i]
		if (p.Direction == domain.PaymentIncoming && a.TargetType != "order" && a.TargetType != "invoice") || (p.Direction == domain.PaymentOutgoing && a.TargetType != "purchase") {
			tx.Rollback()
			return domain.Payment{}, domain.ErrPaymentInvalidParty
		}
		if a.TargetType == "order" {
			var customer string
			var targetTotal int64
			if err = tx.QueryRowContext(ctx, `SELECT COALESCE(customer_id,''),total_rial FROM orders WHERE id=?`, a.TargetID).Scan(&customer, &targetTotal); err != nil {
				tx.Rollback()
				return domain.Payment{}, domain.ErrAllocationTarget
			}
			if p.Direction == domain.PaymentIncoming && p.CustomerID != "" && customer != p.CustomerID {
				tx.Rollback()
				return domain.Payment{}, domain.ErrPaymentInvalidParty
			}
			var already int64
			if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.amount_rial),0) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type='order' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, a.TargetID).Scan(&already); err != nil {
				tx.Rollback()
				return domain.Payment{}, err
			}
			if already+allocatedByTarget[a.TargetID] > targetTotal-a.AmountRial {
				tx.Rollback()
				return domain.Payment{}, domain.ErrAllocationExceeded
			}
			allocatedByTarget[a.TargetID] += a.AmountRial
		} else if a.TargetType == "invoice" {
			var customer string
			var targetTotal int64
			if err = tx.QueryRowContext(ctx, `SELECT COALESCE(customer_id,''),total_rial FROM invoices WHERE id=? AND status IN ('Posted','Partially Paid','Paid')`, a.TargetID).Scan(&customer, &targetTotal); err != nil {
				tx.Rollback()
				return domain.Payment{}, domain.ErrAllocationTarget
			}
			if p.CustomerID != "" && customer != p.CustomerID {
				tx.Rollback()
				return domain.Payment{}, domain.ErrPaymentInvalidParty
			}
			var already int64
			if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.amount_rial),0) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type='invoice' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, a.TargetID).Scan(&already); err != nil {
				tx.Rollback()
				return domain.Payment{}, err
			}
			if already+allocatedByTarget[a.TargetID] > targetTotal-a.AmountRial {
				tx.Rollback()
				return domain.Payment{}, domain.ErrAllocationExceeded
			}
			allocatedByTarget[a.TargetID] += a.AmountRial
		} else {
			var supplier string
			var targetTotal int64
			if err = tx.QueryRowContext(ctx, `SELECT supplier_id,total_rial FROM purchases WHERE id=?`, a.TargetID).Scan(&supplier, &targetTotal); err != nil {
				tx.Rollback()
				return domain.Payment{}, domain.ErrAllocationTarget
			}
			if p.Direction == domain.PaymentOutgoing && p.SupplierID != "" && supplier != p.SupplierID {
				tx.Rollback()
				return domain.Payment{}, domain.ErrPaymentInvalidParty
			}
			var already int64
			if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.amount_rial),0) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type='purchase' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, a.TargetID).Scan(&already); err != nil {
				tx.Rollback()
				return domain.Payment{}, err
			}
			if already+allocatedByTarget[a.TargetID] > targetTotal-a.AmountRial {
				tx.Rollback()
				return domain.Payment{}, domain.ErrAllocationExceeded
			}
			allocatedByTarget[a.TargetID] += a.AmountRial
		}
		if allocTotal > p.AmountRial-a.AmountRial {
			tx.Rollback()
			return domain.Payment{}, domain.ErrAllocationExceeded
		}
		allocTotal += a.AmountRial
	}
	if p.Direction == domain.PaymentIncoming && p.CustomerID == "" && len(p.Allocations) == 0 {
		tx.Rollback()
		return domain.Payment{}, domain.ErrPaymentInvalidParty
	}
	if p.Direction == domain.PaymentOutgoing && p.SupplierID == "" {
		tx.Rollback()
		return domain.Payment{}, domain.ErrPaymentInvalidParty
	}
	lines := []domain.JournalLine{{ID: p.ID + "-L1", JournalEntryID: "JE-PAY-" + p.ID, Position: 0, AccountID: ledger, DebitRial: choose(p.Direction == domain.PaymentIncoming, p.AmountRial, 0), CreditRial: choose(p.Direction == domain.PaymentOutgoing, p.AmountRial, 0), PartyType: "", Memo: "Payment received/paid"}}
	pos := 1
	for _, a := range p.Allocations {
		acct := "ACC-AR"
		party := "customer"
		if p.Direction == domain.PaymentOutgoing {
			acct = "ACC-AP"
			party = "supplier"
		}
		lines = append(lines, domain.JournalLine{ID: fmt.Sprintf("%s-L%d", p.ID, pos+1), JournalEntryID: "JE-PAY-" + p.ID, Position: pos, AccountID: acct, CreditRial: choose(p.Direction == domain.PaymentIncoming, a.AmountRial, 0), DebitRial: choose(p.Direction == domain.PaymentOutgoing, a.AmountRial, 0), PartyType: party, PartyID: partyID(p), Memo: "Allocated to " + a.TargetID})
		pos++
	}
	if rest := p.AmountRial - allocTotal; rest > 0 {
		acct := "ACC-CUSTOMER-CREDIT"
		sideDebit := false
		if p.Direction == domain.PaymentOutgoing {
			acct = "ACC-OTHER-RECEIVABLE"
			sideDebit = true
		}
		lines = append(lines, domain.JournalLine{ID: fmt.Sprintf("%s-L%d", p.ID, pos+1), JournalEntryID: "JE-PAY-" + p.ID, Position: pos, AccountID: acct, DebitRial: choose(sideDebit, rest, 0), CreditRial: choose(!sideDebit, rest, 0), PartyType: chooseString(p.Direction == domain.PaymentIncoming, "customer", "supplier"), PartyID: partyID(p), Memo: "Unallocated balance"})
	}
	entry := domain.JournalEntry{ID: "JE-PAY-" + p.ID, Description: strings.TrimSpace(p.Notes), SourceType: "payment", SourceID: p.ID, IdempotencyKey: "payment:" + p.IdempotencyKey, PostedAt: p.PostedAt, CreatedAt: p.CreatedAt, Lines: lines}
	if entry.Description == "" {
		entry.Description = "Payment"
	}
	if _, err = s.postJournalTx(ctx, tx, entry); err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	var next int64
	if err = tx.QueryRowContext(ctx, `SELECT next_number FROM payment_number_sequences WHERE id=1`).Scan(&next); errors.Is(err, sql.ErrNoRows) {
		if _, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS payment_number_sequences(id INTEGER PRIMARY KEY CHECK(id=1),next_number INTEGER NOT NULL)`); err == nil {
			_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO payment_number_sequences VALUES(1,1001)`)
		}
		if err == nil {
			err = tx.QueryRowContext(ctx, `SELECT next_number FROM payment_number_sequences WHERE id=1`).Scan(&next)
		}
	}
	if err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	p.PaymentNumber = fmt.Sprintf("PAY-%04d", next)
	if _, err = tx.ExecContext(ctx, `UPDATE payment_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO payments(id,payment_number,direction,method,amount_rial,posted_at,financial_account_id,customer_id,supplier_id,reference,notes,status,journal_entry_id,idempotency_key,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, p.ID, p.PaymentNumber, p.Direction, p.Method, p.AmountRial, p.PostedAt.UTC().Format(time.RFC3339Nano), p.FinancialAccountID, nullableString(p.CustomerID), nullableString(p.SupplierID), p.Reference, p.Notes, p.Status, journalIDOr(p, "JE-PAY-"+p.ID), p.IdempotencyKey, p.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	for _, a := range p.Allocations {
		if _, err = tx.ExecContext(ctx, `INSERT INTO payment_allocations(id,payment_id,position,target_type,target_id,amount_rial,reversed) VALUES(?,?,?,?,?,?,0)`, a.ID, allocationPaymentID(a, p.ID), a.Position, a.TargetType, a.TargetID, a.AmountRial); err != nil {
			tx.Rollback()
			return domain.Payment{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.Payment{}, err
	}
	return s.GetPayment(ctx, p.ID)
}

func (s *Store) ReversePayment(ctx context.Context, id, key string) (domain.Payment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Payment{}, err
	}
	var status, journal string
	var posted string
	if err = tx.QueryRowContext(ctx, `SELECT status,journal_entry_id,posted_at FROM payments WHERE id=?`, id).Scan(&status, &journal, &posted); errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return domain.Payment{}, domain.ErrPaymentNotFound
	}
	if err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	if status == string(domain.PaymentReversedState) {
		tx.Rollback()
		return s.GetPayment(ctx, id)
	}
	if key == "" {
		key = "payment:reverse:" + id
	}
	original, err := scanJournalTx(ctx, tx, journal)
	if err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	if _, err = s.postJournalTx(ctx, tx, original.Reversal("REV-"+journal, key, "Payment reversal", time.Now().UTC())); err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE payments SET status='reversed' WHERE id=?`, id); err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE payment_allocations SET reversed=1 WHERE payment_id=?`, id); err != nil {
		tx.Rollback()
		return domain.Payment{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Payment{}, err
	}
	return s.GetPayment(ctx, id)
}

func (s *Store) OrderPaymentSummary(ctx context.Context, id string) (int64, int64, domain.PaymentStatus, error) {
	var paid, total int64
	if err := s.db.QueryRowContext(ctx, `SELECT total_rial FROM orders WHERE id=?`, id).Scan(&total); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, domain.PaymentUnpaid, domain.ErrOrderNotFound
		}
		return 0, 0, domain.PaymentUnpaid, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.amount_rial),0) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type='order' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, id).Scan(&paid); err != nil {
		return 0, 0, domain.PaymentUnpaid, err
	}
	remaining := total - paid
	if remaining < 0 {
		remaining = 0
	}
	status := domain.PaymentUnpaid
	if paid > 0 && paid < total {
		status = domain.PaymentPartiallyPaid
	}
	if paid >= total {
		status = domain.PaymentPaid
	}
	return paid, remaining, status, nil
}
func (s *Store) PurchasePaymentSummary(ctx context.Context, id string) (int64, int64, error) {
	var total, paid int64
	if err := s.db.QueryRowContext(ctx, `SELECT total_rial FROM purchases WHERE id=?`, id).Scan(&total); err != nil {
		return 0, 0, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.amount_rial),0) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type='purchase' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, id).Scan(&paid); err != nil {
		return 0, 0, err
	}
	remaining := total - paid
	if remaining < 0 {
		remaining = 0
	}
	return paid, remaining, nil
}

// CustomerFinancialSummary derives receivables and unapplied customer credit
// from posted journal lines; no customer balance is stored or mutated.
func (s *Store) CustomerFinancialSummary(ctx context.Context, customerID string) (int64, int64, error) {
	var receivable, credit int64
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(debit_rial-credit_rial),0) FROM journal_lines WHERE account_id='ACC-AR' AND party_type='customer' AND party_id=?`, customerID).Scan(&receivable); err != nil {
		return 0, 0, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(credit_rial-debit_rial),0) FROM journal_lines WHERE account_id='ACC-CUSTOMER-CREDIT' AND party_type='customer' AND party_id=?`, customerID).Scan(&credit); err != nil {
		return 0, 0, err
	}
	if receivable < 0 {
		receivable = 0
	}
	if credit < 0 {
		credit = 0
	}
	return receivable, credit, nil
}

// SupplierPayableBalance derives the supplier's open AP from posted journals.
func (s *Store) SupplierPayableBalance(ctx context.Context, supplierID string) (int64, error) {
	var payable int64
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(credit_rial-debit_rial),0) FROM journal_lines WHERE account_id='ACC-AP' AND party_type='supplier' AND party_id=?`, supplierID).Scan(&payable); err != nil {
		return 0, err
	}
	if payable < 0 {
		payable = 0
	}
	return payable, nil
}

func choose(ok bool, a, b int64) int64 {
	if ok {
		return a
	}
	return b
}
func chooseString(ok bool, a, b string) string {
	if ok {
		return a
	}
	return b
}
func partyID(p domain.Payment) string {
	if p.Direction == domain.PaymentIncoming {
		return p.CustomerID
	}
	return p.SupplierID
}
func scanAccount(row scanner) (domain.Account, error) {
	var v domain.Account
	var typ string
	var active, system int
	var c, u string
	if err := row.Scan(&v.ID, &v.Code, &v.Name, &typ, &v.ParentID, &active, &system, &c, &u); err != nil {
		return v, err
	}
	v.Type = domain.AccountType(typ)
	v.Active = active == 1
	v.System = system == 1
	var err error
	v.CreatedAt, err = time.Parse(time.RFC3339Nano, c)
	if err != nil {
		return v, err
	}
	v.UpdatedAt, err = time.Parse(time.RFC3339Nano, u)
	return v, err
}
func scanJournalEntry(row scanner) (domain.JournalEntry, error) {
	var v domain.JournalEntry
	var p, c string
	if err := row.Scan(&v.ID, &v.EntryNumber, &p, &v.Description, &v.SourceType, &v.SourceID, &v.IdempotencyKey, &v.ReversalOfID, &c); err != nil {
		return v, err
	}
	var err error
	v.PostedAt, err = time.Parse(time.RFC3339Nano, p)
	if err != nil {
		return v, err
	}
	v.CreatedAt, err = time.Parse(time.RFC3339Nano, c)
	return v, err
}
func scanFinancialAccount(row scanner) (domain.FinancialAccount, error) {
	var v domain.FinancialAccount
	var typ string
	var active int
	var c, u string
	if err := row.Scan(&v.ID, &v.Name, &typ, &v.LedgerAccountID, &v.Details, &active, &c, &u); err != nil {
		return v, err
	}
	v.Type = domain.FinancialAccountType(typ)
	v.Active = active == 1
	var err error
	v.CreatedAt, err = time.Parse(time.RFC3339Nano, c)
	if err != nil {
		return v, err
	}
	v.UpdatedAt, err = time.Parse(time.RFC3339Nano, u)
	return v, err
}
func scanPayment(row scanner) (domain.Payment, error) {
	var v domain.Payment
	var posted, created string
	if err := row.Scan(&v.ID, &v.PaymentNumber, &v.Direction, &v.Method, &v.AmountRial, &posted, &v.FinancialAccountID, &v.CustomerID, &v.SupplierID, &v.Reference, &v.Notes, &v.Status, &v.JournalEntryID, &v.IdempotencyKey, &created); err != nil {
		return v, err
	}
	var err error
	v.PostedAt, err = time.Parse(time.RFC3339Nano, posted)
	if err != nil {
		return v, err
	}
	v.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	return v, err
}
func journalIDOr(p domain.Payment, v string) string {
	if p.JournalEntryID != "" {
		return p.JournalEntryID
	}
	return v
}
func allocationPaymentID(a domain.PaymentAllocation, v string) string {
	if a.PaymentID != "" {
		return a.PaymentID
	}
	return v
}
