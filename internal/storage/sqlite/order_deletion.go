package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"Atropaten/internal/domain"
)

// Remove operational records atomically, retaining immutable financial and
// inventory history with compensating entries. Payments used only by this
// order are reversed as part of the same transaction; mixed payments retain
// their cash history and release this order's allocations to customer credit.
func (s *Store) deleteOrderTx(ctx context.Context, tx *sql.Tx, id string) error {
	var number string
	if err := tx.QueryRowContext(ctx, `SELECT order_number FROM orders WHERE id=?`, id).Scan(&number); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}
	if err := s.reconcileOrderPaymentsForDeletionTx(ctx, tx, id); err != nil {
		return fmt.Errorf("reverse linked payments: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO deleted_order_records(id,order_number,invoice_id,deleted_at) VALUES(?,?,(SELECT id FROM invoices WHERE order_id=? AND status<>'Draft'),?)`, id, number, id, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	invoices, err := orderDependencyIDs(ctx, tx, `SELECT id FROM invoices WHERE order_id=?`, id)
	if err != nil {
		return err
	}
	for _, invoiceID := range invoices {
		var status string
		if err = tx.QueryRowContext(ctx, `SELECT status FROM invoices WHERE id=?`, invoiceID).Scan(&status); err != nil {
			return err
		}
		if status == domain.InvoiceDraft {
			if err = s.deleteInvoiceRecordTx(ctx, tx, invoiceID); err != nil {
				return fmt.Errorf("remove draft invoice: %w", err)
			}
			continue
		}
		// Void before removing jobs, so their cost reconciliation cannot post
		// further COGS against an invoice being removed from operations.
		if err = s.voidInvoiceTx(ctx, tx, invoiceID); err != nil {
			return fmt.Errorf("void linked invoice: %w", err)
		}
		if err = s.deleteInvoiceRecordTx(ctx, tx, invoiceID); err != nil {
			return fmt.Errorf("remove voided invoice: %w", err)
		}
	}
	jobs, err := orderDependencyIDs(ctx, tx, `SELECT id FROM production_jobs WHERE order_id=?`, id)
	if err != nil {
		return err
	}
	for _, jobID := range jobs {
		if err = s.deleteProductionJobTx(ctx, tx, jobID); err != nil {
			return fmt.Errorf("remove linked production: %w", err)
		}
	}
	for _, query := range []string{
		`DELETE FROM inventory_reservations WHERE order_id=?`,
		`DELETE FROM proofs WHERE owner_type='order' AND owner_id=?`,
		`DELETE FROM attachments WHERE owner_type='order' AND owner_id=?`,
		`DELETE FROM orders WHERE id=?`,
	} {
		if _, err = tx.ExecContext(ctx, query, id); err != nil {
			return fmt.Errorf("remove order dependencies: %w", err)
		}
	}
	// Reversed allocations retain the original order ID as an audit reference;
	// deleted_order_records preserves its human-readable number. Files are left
	// on disk deliberately: imported paths may point to shared original files.
	return nil
}

func (s *Store) deleteInvoiceRecordTx(ctx context.Context, tx *sql.Tx, invoiceID string) error {
	// Journal rows and payment rows are historical records. Detach their
	// foreign-key references before removing the invoice snapshot itself.
	if _, err := tx.ExecContext(ctx, `DELETE FROM invoice_replacements WHERE previous_invoice_id=? OR replacement_invoice_id=?`, invoiceID, invoiceID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM invoice_order_snapshots WHERE invoice_id=?`, invoiceID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM invoice_items WHERE invoice_id=?`, invoiceID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE invoices SET accounting_journal_entry_id=NULL,cogs_journal_entry_id=NULL WHERE id=?`, invoiceID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM invoices WHERE id=?`, invoiceID)
	return err
}

func (s *Store) reconcileOrderPaymentsForDeletionTx(ctx context.Context, tx *sql.Tx, orderID string) error {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT a.payment_id FROM payment_allocations a JOIN payments p ON p.id=a.payment_id
		WHERE a.reversed=0 AND p.status='posted' AND ((a.target_type='order' AND a.target_id=?) OR
		(a.target_type='invoice' AND a.target_id IN (SELECT id FROM invoices WHERE order_id=?))) ORDER BY a.payment_id`, orderID, orderID)
	if err != nil {
		return err
	}
	var paymentIDs []string
	for rows.Next() {
		var paymentID string
		if err = rows.Scan(&paymentID); err != nil {
			rows.Close()
			return err
		}
		paymentIDs = append(paymentIDs, paymentID)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, paymentID := range paymentIDs {
		var unrelated int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_allocations a WHERE a.payment_id=? AND a.reversed=0 AND NOT (
			a.target_type='order' AND a.target_id=? OR a.target_type='invoice' AND a.target_id IN (SELECT id FROM invoices WHERE order_id=?))`, paymentID, orderID, orderID).Scan(&unrelated); err != nil {
			return err
		}
		if unrelated == 0 {
			if err = s.reversePaymentTx(ctx, tx, paymentID, "payment:order-delete:"+orderID+":"+paymentID); err != nil {
				return err
			}
			continue
		}
		if err = s.releaseOrderPaymentAllocationsTx(ctx, tx, orderID, paymentID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) releaseOrderPaymentAllocationsTx(ctx context.Context, tx *sql.Tx, orderID, paymentID string) error {
	rows, err := tx.QueryContext(ctx, `SELECT a.id,a.amount_rial,COALESCE(p.customer_id,''),COALESCE(o.customer_id,'')
		FROM payment_allocations a JOIN payments p ON p.id=a.payment_id JOIN orders o ON o.id=?
		WHERE a.payment_id=? AND a.reversed=0 AND (a.target_type='order' AND a.target_id=? OR
		(a.target_type='invoice' AND a.target_id IN (SELECT id FROM invoices WHERE order_id=?)))`, orderID, paymentID, orderID, orderID)
	if err != nil {
		return err
	}
	type allocation struct {
		id, customer, orderCustomer string
		amount                      int64
	}
	var allocations []allocation
	for rows.Next() {
		var a allocation
		if err = rows.Scan(&a.id, &a.amount, &a.customer, &a.orderCustomer); err != nil {
			rows.Close()
			return err
		}
		allocations = append(allocations, a)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, a := range allocations {
		if _, err = tx.ExecContext(ctx, `UPDATE payment_allocations SET reversed=1 WHERE id=? AND reversed=0`, a.id); err != nil {
			return err
		}
		customer := a.customer
		if customer == "" {
			customer = a.orderCustomer
		}
		jid := "JE-ORDER-DELETE-CREDIT-" + a.id
		now := time.Now().UTC()
		if _, err = s.postJournalTx(ctx, tx, domain.JournalEntry{ID: jid, Description: "Release allocation for deleted order", SourceType: "payment_allocation_adjustment", SourceID: paymentID, IdempotencyKey: jid, PostedAt: now, CreatedAt: now, Lines: []domain.JournalLine{
			{ID: jid + "-AR", JournalEntryID: jid, Position: 0, AccountID: "ACC-AR", DebitRial: a.amount, PartyType: "customer", PartyID: customer, Memo: "Deleted order allocation"},
			{ID: jid + "-CREDIT", JournalEntryID: jid, Position: 1, AccountID: "ACC-CUSTOMER-CREDIT", CreditRial: a.amount, PartyType: "customer", PartyID: customer, Memo: "Unallocated customer credit"},
		}}); err != nil {
			return err
		}
	}
	return nil
}

func orderDependencyIDs(ctx context.Context, tx *sql.Tx, query, id string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var value string
		if err = rows.Scan(&value); err != nil {
			return nil, err
		}
		ids = append(ids, value)
	}
	return ids, rows.Err()
}
