package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"Atropaten/internal/domain"
)

const invoiceSelect = `SELECT id,invoice_number,COALESCE(customer_id,''),customer_name_snapshot,customer_phone_snapshot,COALESCE(order_id,''),issue_date,due_date,status,notes,subtotal_rial,discount_rial,total_rial,COALESCE(accounting_journal_entry_id,''),COALESCE(cogs_journal_entry_id,''),created_at,updated_at FROM invoices`

func (s *Store) ListInvoices(ctx context.Context) ([]domain.Invoice, error) {
	rows, err := s.db.QueryContext(ctx, invoiceSelect+` ORDER BY issue_date DESC,invoice_number DESC`)
	if err != nil {
		return nil, err
	}
	var out []domain.Invoice
	for rows.Next() {
		v, e := scanInvoice(rows)
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
		if err = s.loadInvoiceItems(ctx, &out[i]); err != nil {
			return nil, err
		}
		if err = s.withInvoicePayment(ctx, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (s *Store) GetInvoice(ctx context.Context, id string) (domain.Invoice, error) {
	v, err := scanInvoice(s.db.QueryRowContext(ctx, invoiceSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Invoice{}, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return v, err
	}
	if err = s.loadInvoiceItems(ctx, &v); err != nil {
		return v, err
	}
	err = s.withInvoicePayment(ctx, &v)
	return v, err
}
func (s *Store) GetInvoiceForOrder(ctx context.Context, orderID string) (domain.Invoice, error) {
	v, err := scanInvoice(s.db.QueryRowContext(ctx, invoiceSelect+` WHERE order_id=?`, orderID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Invoice{}, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return v, err
	}
	if err = s.loadInvoiceItems(ctx, &v); err != nil {
		return v, err
	}
	return v, s.withInvoicePayment(ctx, &v)
}
func (s *Store) OrderInvoiceSummary(ctx context.Context, orderID string) (string, string, int64, int64, int64, error) {
	var id, status string
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT id,status,total_rial FROM invoices WHERE order_id=?`, orderID).Scan(&id, &status, &total); errors.Is(err, sql.ErrNoRows) {
		return "", "", 0, 0, 0, domain.ErrInvoiceNotFound
	} else if err != nil {
		return "", "", 0, 0, 0, err
	}
	var paid int64
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.amount_rial),0) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type='invoice' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, id).Scan(&paid); err != nil {
		return "", "", 0, 0, 0, err
	}
	remaining := total - paid
	if remaining < 0 {
		remaining = 0
	}
	if status != "Voided" {
		if paid >= total {
			status = domain.InvoicePaid
		} else if paid > 0 {
			status = domain.InvoicePartiallyPaid
		}
	}
	return id, status, total, paid, remaining, nil
}
func (s *Store) withInvoicePayment(ctx context.Context, v *domain.Invoice) error {
	if v.Status == domain.InvoiceDraft || v.Status == domain.InvoiceVoided {
		v.PaidRial = 0
		v.RemainingRial = v.TotalRial
		return nil
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.amount_rial),0) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type='invoice' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, v.ID).Scan(&v.PaidRial); err != nil {
		return err
	}
	v.RemainingRial = v.TotalRial - v.PaidRial
	if v.RemainingRial < 0 {
		v.RemainingRial = 0
	}
	if v.PaidRial >= v.TotalRial {
		v.Status = domain.InvoicePaid
	} else if v.PaidRial > 0 {
		v.Status = domain.InvoicePartiallyPaid
	}
	return nil
}
func (s *Store) loadInvoiceItems(ctx context.Context, v *domain.Invoice) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,invoice_id,position,COALESCE(order_item_id,''),description_snapshot,service_id,quantity_units,quantity_unit,unit_price_rial,line_total_rial,notes FROM invoice_items WHERE invoice_id=? ORDER BY position,id`, v.ID)
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

func (s *Store) SaveInvoice(ctx context.Context, v domain.Invoice) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if v.Status != domain.InvoiceDraft {
		return domain.ErrInvoiceNotDraft
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	fail := func(e error) error { tx.Rollback(); return e }
	var status string
	if e := tx.QueryRowContext(ctx, `SELECT status FROM invoices WHERE id=?`, v.ID).Scan(&status); e == nil && status != domain.InvoiceDraft {
		return fail(domain.ErrInvoiceNotDraft)
	} else if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return fail(e)
	}
	if v.InvoiceNumber == "" {
		var n int64
		if err = tx.QueryRowContext(ctx, `SELECT next_number FROM invoice_number_sequences WHERE id=1`).Scan(&n); err != nil {
			return fail(err)
		}
		v.InvoiceNumber = fmt.Sprintf("INV-%04d", n)
		if _, err = tx.ExecContext(ctx, `UPDATE invoice_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
			return fail(err)
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO invoices(id,invoice_number,customer_id,customer_name_snapshot,customer_phone_snapshot,order_id,issue_date,due_date,status,notes,subtotal_rial,discount_rial,total_rial,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET customer_id=excluded.customer_id,customer_name_snapshot=excluded.customer_name_snapshot,customer_phone_snapshot=excluded.customer_phone_snapshot,order_id=excluded.order_id,issue_date=excluded.issue_date,due_date=excluded.due_date,status=excluded.status,notes=excluded.notes,subtotal_rial=excluded.subtotal_rial,discount_rial=excluded.discount_rial,total_rial=excluded.total_rial,updated_at=excluded.updated_at`, v.ID, v.InvoiceNumber, nullableString(v.CustomerID), v.CustomerNameSnapshot, v.CustomerPhoneSnapshot, nullableString(v.OrderID), formatTime(&v.IssueDate), nullableTime(v.DueDate), v.Status, v.Notes, v.SubtotalRial, v.DiscountRial, v.TotalRial, formatTime(&v.CreatedAt), formatTime(&v.UpdatedAt))
	if err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM invoice_items WHERE invoice_id=?`, v.ID); err != nil {
		return fail(err)
	}
	for _, i := range v.Items {
		if _, err = tx.ExecContext(ctx, `INSERT INTO invoice_items(id,invoice_id,position,order_item_id,description_snapshot,service_id,quantity_units,quantity_unit,unit_price_rial,line_total_rial,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, i.ID, v.ID, i.Position, nullableString(i.OrderItemID), i.DescriptionSnapshot, i.ServiceID, i.QuantityUnits, i.QuantityUnit, i.UnitPriceRial, i.LineTotalRial, i.Notes); err != nil {
			return fail(err)
		}
	}
	return tx.Commit()
}
func (s *Store) DeleteDraftInvoice(ctx context.Context, id string) error {
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM invoices WHERE id=?`, id).Scan(&status); errors.Is(err, sql.ErrNoRows) {
		return domain.ErrInvoiceNotFound
	} else if err != nil {
		return err
	}
	if status != domain.InvoiceDraft {
		return domain.ErrInvoiceProtected
	}
	var allocations int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_allocations WHERE target_type='invoice' AND target_id=?`, id).Scan(&allocations); err != nil {
		return err
	}
	if allocations > 0 {
		return domain.ErrInvoiceProtected
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM invoices WHERE id=?`, id)
	return err
}

func (s *Store) PostInvoice(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	fail := func(e error) error { tx.Rollback(); return e }
	var v domain.Invoice
	var issue, created, updated string
	var due sql.NullString
	if err = tx.QueryRowContext(ctx, invoiceSelect+` WHERE id=?`, id).Scan(&v.ID, &v.InvoiceNumber, &v.CustomerID, &v.CustomerNameSnapshot, &v.CustomerPhoneSnapshot, &v.OrderID, &issue, &due, &v.Status, &v.Notes, &v.SubtotalRial, &v.DiscountRial, &v.TotalRial, &v.AccountingJournalEntryID, &v.COGSJournalEntryID, &created, &updated); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrInvoiceNotFound)
	}
	if err != nil {
		return fail(err)
	}
	if v.Status != domain.InvoiceDraft {
		return tx.Commit()
	}
	v.IssueDate, _ = time.Parse(time.RFC3339Nano, issue)
	v.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	v.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	if due.Valid {
		v.DueDate, _ = parseNullableTime(due.String)
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,invoice_id,position,COALESCE(order_item_id,''),description_snapshot,service_id,quantity_units,quantity_unit,unit_price_rial,line_total_rial,notes FROM invoice_items WHERE invoice_id=? ORDER BY position,id`, id)
	if err != nil {
		return fail(err)
	}
	for rows.Next() {
		var i domain.InvoiceItem
		if err = rows.Scan(&i.ID, &i.InvoiceID, &i.Position, &i.OrderItemID, &i.DescriptionSnapshot, &i.ServiceID, &i.QuantityUnits, &i.QuantityUnit, &i.UnitPriceRial, &i.LineTotalRial, &i.Notes); err != nil {
			rows.Close()
			return fail(err)
		}
		v.Items = append(v.Items, i)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return fail(err)
	}
	if v.TotalRial <= 0 {
		return fail(fmt.Errorf("invoice total must be positive before posting"))
	}
	now := time.Now().UTC()
	je := "JE-INV-" + id
	entry := domain.JournalEntry{ID: je, Description: "Posted invoice " + v.InvoiceNumber, SourceType: "invoice", SourceID: id, IdempotencyKey: "invoice:post:" + id, PostedAt: v.IssueDate, CreatedAt: now, Lines: []domain.JournalLine{{ID: je + "-L1", JournalEntryID: je, Position: 0, AccountID: "ACC-AR", DebitRial: v.TotalRial, PartyType: "customer", PartyID: v.CustomerID, Memo: "Invoice receivable"}, {ID: je + "-L2", JournalEntryID: je, Position: 1, AccountID: "ACC-REVENUE", CreditRial: v.TotalRial, PartyType: "customer", PartyID: v.CustomerID, Memo: "Sales/service revenue"}}}
	if _, err = s.postJournalTx(ctx, tx, entry); err != nil {
		return fail(err)
	}
	cogs, err := s.orderCOGSValueTx(ctx, tx, v.OrderID)
	if err != nil {
		return fail(err)
	}
	cogsID := ""
	if cogs > 0 && v.OrderID != "" {
		cogsID = "JE-COGS-INV-" + id
		ce := domain.JournalEntry{ID: cogsID, Description: "COGS for invoice " + v.InvoiceNumber, SourceType: "invoice_cogs", SourceID: id, IdempotencyKey: "invoice:cogs:" + id, PostedAt: v.IssueDate, CreatedAt: now, Lines: []domain.JournalLine{{ID: cogsID + "-L1", JournalEntryID: cogsID, Position: 0, AccountID: "ACC-COGS", DebitRial: cogs, PartyType: "order", PartyID: v.OrderID, Memo: "Actual material consumption and waste"}, {ID: cogsID + "-L2", JournalEntryID: cogsID, Position: 1, AccountID: "ACC-INVENTORY", CreditRial: cogs, PartyType: "order", PartyID: v.OrderID, Memo: "Actual inventory movement cost"}}}
		if _, err = s.postJournalTx(ctx, tx, ce); err != nil {
			return fail(err)
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE invoices SET status='Posted',accounting_journal_entry_id=?,cogs_journal_entry_id=?,updated_at=? WHERE id=?`, je, nullableString(cogsID), now.Format(time.RFC3339Nano), id); err != nil {
		return fail(err)
	}
	return tx.Commit()
}

// COGS recognition boundary: invoice posting recognizes the exact net cost of
// production consumption and waste movements already linked to this order.
// It never consults current catalog prices and never backfills later movements.
func (s *Store) orderCOGSValueTx(ctx context.Context, tx *sql.Tx, orderID string) (int64, error) {
	if orderID == "" {
		return 0, nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT m.total_cost_rial FROM inventory_movements m JOIN production_consumptions c ON c.id=m.reference_id JOIN production_jobs j ON j.id=c.production_job_id WHERE j.order_id=? AND m.reference_type IN ('production_consumption','production_correction')`, orderID)
	if err != nil {
		return 0, err
	}
	var total big.Int
	for rows.Next() {
		var cost int64
		if err = rows.Scan(&cost); err != nil {
			rows.Close()
			return 0, err
		}
		total.Sub(&total, big.NewInt(cost))
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()
	if total.Sign() < 0 {
		return 0, nil
	}
	if !total.IsInt64() {
		return 0, fmt.Errorf("COGS amount is too large")
	}
	return total.Int64(), nil
}
func (s *Store) VoidInvoice(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	fail := func(e error) error { tx.Rollback(); return e }
	var status, je, cogs string
	if err = tx.QueryRowContext(ctx, `SELECT status,COALESCE(accounting_journal_entry_id,''),COALESCE(cogs_journal_entry_id,'') FROM invoices WHERE id=?`, id).Scan(&status, &je, &cogs); errors.Is(err, sql.ErrNoRows) {
		return fail(domain.ErrInvoiceNotFound)
	}
	if err != nil {
		return fail(err)
	}
	if status == domain.InvoiceVoided {
		return tx.Commit()
	}
	if status != domain.InvoicePosted {
		return fail(domain.ErrInvoiceCannotVoid)
	}
	var n int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_allocations WHERE target_type='invoice' AND target_id=? AND reversed=0`, id).Scan(&n); err != nil {
		return fail(err)
	}
	if n > 0 {
		return fail(domain.ErrInvoiceCannotVoid)
	}
	if je != "" {
		if _, err = s.reverseJournalTx(ctx, tx, je, "invoice:void:"+id, "Void invoice "+id, time.Now().UTC()); err != nil {
			return fail(err)
		}
	}
	if cogs != "" {
		if _, err = s.reverseJournalTx(ctx, tx, cogs, "invoice:cogs:void:"+id, "Void COGS for invoice "+id, time.Now().UTC()); err != nil {
			return fail(err)
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE invoices SET status='Voided',updated_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), id); err != nil {
		return fail(err)
	}
	return tx.Commit()
}

func scanInvoice(row scanner) (domain.Invoice, error) {
	var v domain.Invoice
	var issue, created, updated string
	var due sql.NullString
	if err := row.Scan(&v.ID, &v.InvoiceNumber, &v.CustomerID, &v.CustomerNameSnapshot, &v.CustomerPhoneSnapshot, &v.OrderID, &issue, &due, &v.Status, &v.Notes, &v.SubtotalRial, &v.DiscountRial, &v.TotalRial, &v.AccountingJournalEntryID, &v.COGSJournalEntryID, &created, &updated); err != nil {
		return v, err
	}
	var err error
	v.IssueDate, err = time.Parse(time.RFC3339Nano, issue)
	if err != nil {
		return v, err
	}
	if due.Valid {
		v.DueDate, err = parseNullableTime(due.String)
		if err != nil {
			return v, err
		}
	}
	v.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return v, err
	}
	v.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	return v, err
}
func parseNullableTime(v string) (*time.Time, error) {
	if v == "" {
		return nil, nil
	}
	t, e := time.Parse(time.RFC3339Nano, v)
	return &t, e
}
func formatTime(v *time.Time) any {
	if v == nil || v.IsZero() {
		return nil
	}
	return v.UTC().Format(time.RFC3339Nano)
}
