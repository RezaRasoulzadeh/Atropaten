package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"Atropaten/internal/domain"
)

func (s *Store) reconcileJobCOGSTx(ctx context.Context, tx *sql.Tx, jobID string) error {
	var orderID string
	if err := tx.QueryRowContext(ctx, `SELECT order_id FROM production_jobs WHERE id=?`, jobID).Scan(&orderID); err != nil {
		return err
	}
	return s.reconcileOrderCOGSTx(ctx, tx, orderID)
}

// A posted order invoice recognizes subsequent usage/returns through additional
// journal entries. Estimates never enter the ledger and outsourcing is already
// recognized by its expense, so neither is counted as inventory COGS here.
func (s *Store) reconcileOrderCOGSTx(ctx context.Context, tx *sql.Tx, orderID string) error {
	var invoiceID string
	err := tx.QueryRowContext(ctx, `SELECT id FROM invoices WHERE order_id=? AND status IN ('Posted','Partially Paid','Paid') ORDER BY created_at,id LIMIT 1`, orderID).Scan(&invoiceID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	target, err := s.orderCOGSValueTx(ctx, tx, orderID)
	if err != nil {
		return err
	}
	var recognized int64
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(l.debit_rial-l.credit_rial),0) FROM journal_lines l JOIN journal_entries e ON e.id=l.journal_entry_id WHERE l.account_id='ACC-COGS' AND e.source_id=? AND e.source_type IN ('invoice_cogs','invoice_cogs_adjustment')`, invoiceID).Scan(&recognized); err != nil {
		return err
	}
	delta := target - recognized
	if delta == 0 {
		return nil
	}
	now := time.Now().UTC()
	id := fmt.Sprintf("JE-COGS-ADJ-%s-%d", invoiceID, now.UnixNano())
	debit, credit := delta, int64(0)
	if delta < 0 {
		debit, credit = 0, -delta
	}
	_, err = s.postJournalTx(ctx, tx, domain.JournalEntry{ID: id, Description: "Production cost adjustment for " + invoiceID, SourceType: "invoice_cogs_adjustment", SourceID: invoiceID, IdempotencyKey: id, PostedAt: now, CreatedAt: now, Lines: []domain.JournalLine{
		{ID: id + "-1", JournalEntryID: id, Position: 0, AccountID: "ACC-COGS", DebitRial: debit, CreditRial: credit, PartyType: "order", PartyID: orderID, Memo: "Net production material and waste cost"},
		{ID: id + "-2", JournalEntryID: id, Position: 1, AccountID: "ACC-INVENTORY", DebitRial: credit, CreditRial: debit, Memo: "Inventory cost adjustment"},
	}})
	return err
}

func (s *Store) reverseInvoiceCOGSAdjustmentsTx(ctx context.Context, tx *sql.Tx, invoiceID string) error {
	rows, err := tx.QueryContext(ctx, `SELECT id FROM journal_entries WHERE source_type='invoice_cogs_adjustment' AND source_id=? ORDER BY created_at,id`, invoiceID)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err = s.reverseJournalTx(ctx, tx, id, "invoice:cogs:void:"+id, "Void production cost adjustment", time.Now().UTC()); err != nil {
			return err
		}
	}
	return nil
}
