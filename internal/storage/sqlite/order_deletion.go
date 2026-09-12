package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"Atropaten/internal/domain"
)

// Remove operational records atomically, but retain immutable financial and
// inventory history with compensating entries. Never reverse an active payment
// implicitly: it may also allocate money to unrelated orders or invoices.
func (s *Store) deleteOrderTx(ctx context.Context, tx *sql.Tx, id string) error {
	var number string
	if err := tx.QueryRowContext(ctx, `SELECT order_number FROM orders WHERE id=?`, id).Scan(&number); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}
	var active int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_allocations a
		JOIN payments p ON p.id=a.payment_id WHERE a.reversed=0 AND p.status='posted'
		AND ((a.target_type='order' AND a.target_id=?) OR
		(a.target_type='invoice' AND a.target_id IN (SELECT id FROM invoices WHERE order_id=?)))`, id, id).Scan(&active); err != nil {
		return err
	}
	if active > 0 {
		return domain.ErrOrderDeleteProtected
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
			if _, err = tx.ExecContext(ctx, `DELETE FROM invoice_items WHERE invoice_id=?`, invoiceID); err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, `DELETE FROM invoices WHERE id=?`, invoiceID); err != nil {
				return err
			}
			continue
		}
		// Void before removing jobs, so their cost reconciliation cannot post
		// further COGS against an invoice being removed from operations.
		if err = s.voidInvoiceTx(ctx, tx, invoiceID); err != nil {
			return fmt.Errorf("void linked invoice: %w", err)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE invoice_items SET order_item_id=NULL WHERE invoice_id=? AND order_item_id IN (SELECT id FROM order_items WHERE order_id=?)`, invoiceID, id); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE invoices SET order_id=NULL WHERE id=?`, invoiceID); err != nil {
			return err
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
