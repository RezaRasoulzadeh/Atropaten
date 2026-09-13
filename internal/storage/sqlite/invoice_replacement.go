package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"Atropaten/internal/domain"
)

// Voiding releases invoice allocations as customer credit. The payment and its
// original journal remain untouched; only the allocation and receivable claim
// are adjusted. Reversing the payment later reverses this adjustment too.
func (s *Store) releaseInvoiceAllocationsTx(ctx context.Context, tx *sql.Tx, invoiceID string) error {
	rows, err := tx.QueryContext(ctx, `SELECT a.id,a.payment_id,a.amount_rial,COALESCE(p.customer_id,''),COALESCE(i.customer_id,'') FROM payment_allocations a JOIN payments p ON p.id=a.payment_id JOIN invoices i ON i.id=a.target_id WHERE a.target_type='invoice' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, invoiceID)
	if err != nil {
		return err
	}
	type allocation struct {
		id, payment, customer, invoiceCustomer string
		amount                                 int64
	}
	var allocations []allocation
	for rows.Next() {
		var a allocation
		if err = rows.Scan(&a.id, &a.payment, &a.amount, &a.customer, &a.invoiceCustomer); err != nil {
			rows.Close()
			return err
		}
		if a.customer == "" {
			a.customer = a.invoiceCustomer
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
		now := time.Now().UTC()
		jid := "JE-INVOICE-CREDIT-" + a.id
		_, err = s.postJournalTx(ctx, tx, domain.JournalEntry{
			ID: jid, Description: "Release allocation for voided invoice " + invoiceID,
			SourceType: "payment_allocation_adjustment", SourceID: a.payment,
			IdempotencyKey: jid, PostedAt: now, CreatedAt: now,
			Lines: []domain.JournalLine{
				{ID: jid + "-AR", JournalEntryID: jid, Position: 0, AccountID: "ACC-AR", DebitRial: a.amount, PartyType: "customer", PartyID: a.customer, Memo: "Voided invoice allocation"},
				{ID: jid + "-CREDIT", JournalEntryID: jid, Position: 1, AccountID: "ACC-CUSTOMER-CREDIT", CreditRial: a.amount, PartyType: "customer", PartyID: a.customer, Memo: "Unallocated customer credit"},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func linkInvoiceReplacementTx(ctx context.Context, tx *sql.Tx, invoiceID, orderID string) error {
	if orderID == "" {
		return nil
	}
	var previous string
	err := tx.QueryRowContext(ctx, `SELECT id FROM invoices WHERE order_id=? AND status='Voided' AND id<>? AND NOT EXISTS(SELECT 1 FROM invoice_replacements r WHERE r.previous_invoice_id=invoices.id) ORDER BY updated_at DESC,id DESC LIMIT 1`, orderID, invoiceID).Scan(&previous)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO invoice_replacements(previous_invoice_id,replacement_invoice_id,created_at) VALUES(?,?,?)`, previous, invoiceID, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("link invoice replacement: %w", err)
	}
	return nil
}
