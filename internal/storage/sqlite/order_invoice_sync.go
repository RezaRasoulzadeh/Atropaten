package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"Atropaten/internal/domain"
)

const orderItemsSelect = `SELECT id,order_id,display_order,service_id,service_name_snapshot,service_code_snapshot,quantity_units,quantity_unit,resolved_parameters_json,cost_breakdown_json,pricing_snapshot_json,estimated_cost_rial,suggested_price_rial,selling_price_rial,notes FROM order_items WHERE order_id=? AND removed_at IS NULL ORDER BY display_order,id`

func loadOrderTx(ctx context.Context, tx *sql.Tx, id string) (domain.Order, error) {
	o, err := scanOrder(tx.QueryRowContext(ctx, orderSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return o, domain.ErrOrderNotFound
	}
	if err != nil {
		return o, err
	}
	rows, err := tx.QueryContext(ctx, orderItemsSelect, id)
	if err != nil {
		return o, err
	}
	defer rows.Close()
	for rows.Next() {
		i, err := scanOrderItem(rows)
		if err != nil {
			return o, err
		}
		o.Items = append(o.Items, i)
	}
	return o, rows.Err()
}

// Retain configuration as well as visible invoice lines. Status, scheduling and
// timestamps are operational metadata and must not trigger financial reissues.
func orderCommercialSnapshot(o domain.Order) string {
	data, _ := json.Marshal(struct {
		CustomerID, Name, Phone, Notes string
		Discount                       int64
		Items                          []domain.OrderItem
	}{o.CustomerID, o.CustomerNameSnapshot, o.CustomerPhoneSnapshot, o.Notes, o.DiscountRial, o.Items})
	return string(data)
}

func saveInvoiceOrderSnapshotTx(ctx context.Context, tx *sql.Tx, invoiceID, orderID string) error {
	if orderID == "" {
		return nil
	}
	o, err := loadOrderTx(ctx, tx, orderID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO invoice_order_snapshots(invoice_id,snapshot_json) VALUES(?,?) ON CONFLICT(invoice_id) DO UPDATE SET snapshot_json=excluded.snapshot_json`, invoiceID, orderCommercialSnapshot(o))
	return err
}

func (s *Store) reconcileOrderInvoicesTx(ctx context.Context, tx *sql.Tx, o domain.Order) error {
	v, err := scanInvoice(tx.QueryRowContext(ctx, invoiceSelect+` WHERE order_id=? AND status<>'Voided'`, o.ID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = loadInvoiceItemsTx(ctx, tx, &v); err != nil {
		return err
	}
	updated, err := v.SnapshotOrder(o)
	if err != nil {
		return err
	}
	var snapshot string
	err = tx.QueryRowContext(ctx, `SELECT snapshot_json FROM invoice_order_snapshots WHERE invoice_id=?`, v.ID).Scan(&snapshot)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if reflect.DeepEqual(v, updated) && snapshot == orderCommercialSnapshot(o) {
		return nil
	}
	updated.UpdatedAt = time.Now().UTC()
	if v.Status == domain.InvoiceDraft {
		return s.saveInvoiceTx(ctx, tx, updated)
	}

	// Move direct allocations to the order before voiding. The receivable party
	// does not change, and the original allocation rows remain as reversed facts.
	if err = s.reconcileOrderAllocationsTx(ctx, tx, o, v.ID); err != nil {
		return err
	}
	if err = s.voidInvoiceTx(ctx, tx, v.ID); err != nil {
		return err
	}
	now := time.Now().UTC()
	id := fmt.Sprintf("INV-REISSUE-%x", sha256.Sum256([]byte(v.ID)))
	replacement, err := (domain.Invoice{ID: id, Status: domain.InvoiceDraft, IssueDate: now, DueDate: v.DueDate, CreatedAt: now, UpdatedAt: now}).SnapshotOrder(o)
	if err != nil {
		return err
	}
	if err = s.saveInvoiceTx(ctx, tx, replacement); err != nil {
		return err
	}
	if err = s.postInvoiceTx(ctx, tx, id, true); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO invoice_replacements(previous_invoice_id,replacement_invoice_id,created_at) VALUES(?,?,?)`, v.ID, id, now.Format(time.RFC3339Nano))
	return err
}

func loadInvoiceItemsTx(ctx context.Context, tx *sql.Tx, v *domain.Invoice) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,invoice_id,position,COALESCE(order_item_id,''),description_snapshot,service_id,quantity_units,quantity_unit,unit_price_rial,line_total_rial,notes FROM invoice_items WHERE invoice_id=? ORDER BY position,id`, v.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var i domain.InvoiceItem
		if err = rows.Scan(&i.ID, &i.InvoiceID, &i.Position, &i.OrderItemID, &i.DescriptionSnapshot, &i.ServiceID, &i.QuantityUnits, &i.QuantityUnit, &i.UnitPriceRial, &i.LineTotalRial, &i.Notes); err != nil {
			return err
		}
		v.Items = append(v.Items, i)
	}
	return rows.Err()
}

func (s *Store) reconcileOrderFinancialsTx(ctx context.Context, tx *sql.Tx, orderID string) error {
	o, err := loadOrderTx(ctx, tx, orderID)
	if err != nil {
		return err
	}
	if err = s.reconcileOrderInvoicesTx(ctx, tx, o); err != nil {
		return err
	}
	if err = s.reconcileOrderAllocationsTx(ctx, tx, o, ""); err != nil {
		return err
	}
	var paid int64
	if err = tx.QueryRowContext(ctx, orderPaidTotalSQL, orderID, orderID).Scan(&paid); err != nil {
		return err
	}
	status := domain.PaymentUnpaid
	if paid >= o.TotalRial {
		status = domain.PaymentPaid
	} else if paid > 0 {
		status = domain.PaymentPartiallyPaid
	}
	_, err = tx.ExecContext(ctx, `UPDATE orders SET payment_status=? WHERE id=?`, status, orderID)
	return err
}
