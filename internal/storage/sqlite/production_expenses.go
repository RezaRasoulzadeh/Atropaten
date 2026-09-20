package sqlite

import (
	"Atropaten/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Production, stock returns, expense, and balanced journal post in one transaction.
func (s *Store) reconcileOutsourceExpenseTx(ctx context.Context, tx *sql.Tx, j domain.ProductionJob) error {
	id := "EXP-OUTSOURCE-" + j.ID
	if j.ActualOutsourcedCostRial == 0 {
		return s.reverseOutsourceExpenseTx(ctx, tx, j.ID)
	}
	if j.OutsourcePaymentStatus == "" {
		j.OutsourcePaymentStatus = domain.OutsourcePaymentPaid
	}
	if !domain.ValidOutsourcePaymentStatus(j.OutsourcePaymentStatus) {
		return fmt.Errorf("unsupported outsourcing payment status")
	}
	var ledger, method string
	financialAccountID := j.OutsourceFinancialAccountID
	if j.OutsourcePaymentStatus == domain.OutsourcePaymentOnHold {
		if j.OutsourceSupplierID == "" {
			return fmt.Errorf("select a supplier for an outsourcing expense on hold")
		}
		ledger = "ACC-AP"
		method = string(domain.PaymentOther)
	} else if err := tx.QueryRowContext(ctx, `SELECT ledger_account_id,type FROM financial_accounts WHERE id=? AND active=1`, financialAccountID).Scan(&ledger, &method); err != nil {
		return fmt.Errorf("choose an active payment account for the outsourcing expense")
	}
	var category string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM accounts WHERE id='ACC-COGS' AND type='expense' AND active=1`).Scan(&category); err != nil {
		return domain.ErrAccountNotFound
	}
	var amount int64
	var account, supplier, notes, status, description, paymentStatus string
	err := tx.QueryRowContext(ctx, `SELECT amount_rial,COALESCE(financial_account_id,''),COALESCE(supplier_id,''),notes,status,description,COALESCE(payment_status,'paid') FROM expenses WHERE id=?`, id).Scan(&amount, &account, &supplier, &notes, &status, &description, &paymentStatus)
	exists := err == nil
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	desc := fmt.Sprintf("Outsourced production %s · %s × %d Rial · %s", j.JobNumber, j.OutsourceQuantity.String(), j.OutsourceUnitCostRial, j.OutsourceDescription)
	if exists && status == "Posted" && amount == j.ActualOutsourcedCostRial && account == financialAccountID && supplier == j.OutsourceSupplierID && notes == j.OutsourceNotes && description == desc && paymentStatus == j.OutsourcePaymentStatus {
		return nil
	}
	if err = s.reverseOutsourceExpenseTx(ctx, tx, j.ID); err != nil {
		return err
	}
	now := time.Now().UTC()
	stamp := now.Format(time.RFC3339Nano)
	journalID := fmt.Sprintf("JE-OUTSOURCE-%s-%d", j.ID, now.UnixNano())
	creditPartyType, creditPartyID, creditMemo := "", "", "Outsourcing payment"
	if j.OutsourcePaymentStatus == domain.OutsourcePaymentOnHold {
		creditPartyType, creditPartyID, creditMemo = "supplier", j.OutsourceSupplierID, "Outsourcing payable"
	}
	entry := domain.JournalEntry{ID: journalID, Description: desc, SourceType: "expense", SourceID: id, IdempotencyKey: journalID, PostedAt: now, CreatedAt: now, Lines: []domain.JournalLine{
		{ID: journalID + "-1", JournalEntryID: journalID, Position: 0, AccountID: category, DebitRial: j.ActualOutsourcedCostRial, PartyType: "supplier", PartyID: j.OutsourceSupplierID, Memo: desc},
		{ID: journalID + "-2", JournalEntryID: journalID, Position: 1, AccountID: ledger, CreditRial: j.ActualOutsourcedCostRial, PartyType: creditPartyType, PartyID: creditPartyID, Memo: creditMemo},
	}}
	if _, err = s.postJournalTx(ctx, tx, entry); err != nil {
		return err
	}
	if exists {
		_, err = tx.ExecContext(ctx, `UPDATE expenses SET amount_rial=?,financial_account_id=?,payment_method=?,payment_status=?,supplier_id=?,payee=?,description=?,notes=?,status='Posted',journal_entry_id=?,updated_at=? WHERE id=?`, j.ActualOutsourcedCostRial, nullableString(financialAccountID), method, j.OutsourcePaymentStatus, nullableString(j.OutsourceSupplierID), j.OutsourceSupplierID, desc, j.OutsourceNotes, journalID, stamp, id)
	} else {
		var n int64
		if err = tx.QueryRowContext(ctx, `SELECT next_number FROM expense_number_sequences WHERE id=1`).Scan(&n); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE expense_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO expenses(id,expense_number,expense_date,category_account_id,payee,supplier_id,description,amount_rial,payment_method,payment_status,financial_account_id,notes,status,journal_entry_id,idempotency_key,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,'Posted',?,?,?,?)`, id, fmt.Sprintf("EXP-%04d", n), stamp, category, j.OutsourceSupplierID, nullableString(j.OutsourceSupplierID), desc, j.ActualOutsourcedCostRial, method, j.OutsourcePaymentStatus, nullableString(financialAccountID), j.OutsourceNotes, journalID, "outsource:"+j.ID, stamp, stamp)
	}
	return err
}
