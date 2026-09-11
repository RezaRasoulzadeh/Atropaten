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

func (s *Store) ListSuppliers(ctx context.Context, includeArchived bool) ([]domain.Supplier, error) {
	q := `SELECT id,name,code,phone,email,address,notes,active,created_at,updated_at FROM suppliers`
	if !includeArchived {
		q += ` WHERE active=1`
	}
	q += ` ORDER BY lower(name),id`
	rows, e := s.db.QueryContext(ctx, q)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Supplier{}
	for rows.Next() {
		v, e := scanSupplier(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) GetSupplier(ctx context.Context, id string) (domain.Supplier, error) {
	v, e := scanSupplier(s.db.QueryRowContext(ctx, `SELECT id,name,code,phone,email,address,notes,active,created_at,updated_at FROM suppliers WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.Supplier{}, domain.ErrSupplierNotFound
	}
	return v, e
}
func (s *Store) SaveSupplier(ctx context.Context, v domain.Supplier) error {
	res, e := s.db.ExecContext(ctx, `UPDATE suppliers SET name=?,code=?,phone=?,email=?,address=?,notes=?,active=?,updated_at=? WHERE id=?`, v.Name, v.Code, v.Phone, v.Email, v.Address, v.Notes, boolToInt(v.Active), v.UpdatedAt.UTC().Format(time.RFC3339Nano), v.ID)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		_, e = s.db.ExecContext(ctx, `INSERT INTO suppliers(id,name,code,phone,email,address,notes,active,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, v.ID, v.Name, v.Code, v.Phone, v.Email, v.Address, v.Notes, boolToInt(v.Active), v.CreatedAt.UTC().Format(time.RFC3339Nano), v.UpdatedAt.UTC().Format(time.RFC3339Nano))
	}
	return e
}
func (s *Store) DeleteSupplier(ctx context.Context, id string) error {
	var n int
	for _, query := range []string{
		`SELECT COUNT(*) FROM purchases WHERE supplier_id=?`,
		`SELECT COUNT(*) FROM payments WHERE supplier_id=?`,
		`SELECT COUNT(*) FROM checks WHERE supplier_id=?`,
		`SELECT COUNT(*) FROM loans WHERE supplier_id=?`,
		`SELECT COUNT(*) FROM production_jobs WHERE outsource_supplier_id=?`,
	} {
		if e := s.db.QueryRowContext(ctx, query, id).Scan(&n); e != nil {
			return e
		}
		if n > 0 {
			return domain.ErrSupplierDeleteProtected
		}
	}
	res, e := s.db.ExecContext(ctx, `DELETE FROM suppliers WHERE id=?`, id)
	if e != nil {
		return e
	}
	n64, _ := res.RowsAffected()
	if n64 == 0 {
		return domain.ErrSupplierNotFound
	}
	return nil
}

func (s *Store) ListPurchases(ctx context.Context) ([]domain.Purchase, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,purchase_number,supplier_id,supplier_name_snapshot,supplier_code_snapshot,supplier_invoice_number,COALESCE(financial_account_id,''),purchase_date,status,notes,subtotal_rial,discount_rial,shipping_rial,tax_rial,additional_costs_rial,total_rial,created_at,updated_at,archived FROM purchases ORDER BY purchase_date DESC,purchase_number DESC`)
	if e != nil {
		return nil, e
	}
	out := []domain.Purchase{}
	for rows.Next() {
		p, e := scanPurchase(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	if e := rows.Err(); e != nil {
		return nil, e
	}
	if e := rows.Close(); e != nil {
		return nil, e
	}
	for i := range out {
		if e := s.loadPurchaseItems(ctx, &out[i]); e != nil {
			return nil, e
		}
	}
	return out, nil
}
func (s *Store) GetPurchase(ctx context.Context, id string) (domain.Purchase, error) {
	p, e := scanPurchase(s.db.QueryRowContext(ctx, `SELECT id,purchase_number,supplier_id,supplier_name_snapshot,supplier_code_snapshot,supplier_invoice_number,COALESCE(financial_account_id,''),purchase_date,status,notes,subtotal_rial,discount_rial,shipping_rial,tax_rial,additional_costs_rial,total_rial,created_at,updated_at,archived FROM purchases WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.Purchase{}, domain.ErrPurchaseNotFound
	}
	if e != nil {
		return p, e
	}
	if e = s.loadPurchaseItems(ctx, &p); e != nil {
		return p, e
	}
	return p, nil
}
func (s *Store) SavePurchase(ctx context.Context, p domain.Purchase) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var existingStatus string
	statusErr := tx.QueryRowContext(ctx, `SELECT status FROM purchases WHERE id=?`, p.ID).Scan(&existingStatus)
	if statusErr == nil && existingStatus != domain.PurchaseDraft && existingStatus != domain.PurchaseCancelled {
		return fail(domain.ErrPurchaseNotDraft)
	}
	if statusErr != nil && !errors.Is(statusErr, sql.ErrNoRows) {
		return fail(statusErr)
	}
	var number string
	if p.PurchaseNumber == "" {
		var next int64
		if e = tx.QueryRowContext(ctx, `SELECT next_number FROM purchase_number_sequences WHERE id=1`).Scan(&next); e != nil {
			return fail(e)
		}
		number = fmt.Sprintf("PUR-%04d", next)
		if _, e = tx.ExecContext(ctx, `UPDATE purchase_number_sequences SET next_number=next_number+1 WHERE id=1`); e != nil {
			return fail(e)
		}
	} else {
		number = p.PurchaseNumber
	}
	archived := 0
	if p.Archived {
		archived = 1
	}
	var financialAccount any
	if p.FinancialAccountID != "" {
		financialAccount = p.FinancialAccountID
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO purchases(id,purchase_number,supplier_id,supplier_name_snapshot,supplier_code_snapshot,supplier_invoice_number,financial_account_id,purchase_date,status,notes,subtotal_rial,discount_rial,shipping_rial,tax_rial,additional_costs_rial,total_rial,created_at,updated_at,archived) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET supplier_id=excluded.supplier_id,supplier_name_snapshot=excluded.supplier_name_snapshot,supplier_code_snapshot=excluded.supplier_code_snapshot,supplier_invoice_number=excluded.supplier_invoice_number,financial_account_id=excluded.financial_account_id,purchase_date=excluded.purchase_date,status=excluded.status,notes=excluded.notes,subtotal_rial=excluded.subtotal_rial,discount_rial=excluded.discount_rial,shipping_rial=excluded.shipping_rial,tax_rial=excluded.tax_rial,additional_costs_rial=excluded.additional_costs_rial,total_rial=excluded.total_rial,updated_at=excluded.updated_at,archived=excluded.archived`, p.ID, number, p.SupplierID, p.SupplierNameSnapshot, p.SupplierCodeSnapshot, p.SupplierInvoiceNumber, financialAccount, p.PurchaseDate.UTC().Format(time.RFC3339Nano), p.Status, p.Notes, p.SubtotalRial, p.DiscountRial, p.ShippingRial, p.TaxRial, p.AdditionalCostsRial, p.TotalRial, p.CreatedAt.UTC().Format(time.RFC3339Nano), p.UpdatedAt.UTC().Format(time.RFC3339Nano), archived)
	if e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM purchase_items WHERE purchase_id=?`, p.ID); e != nil {
		return fail(e)
	}
	for i, item := range p.Items {
		if _, e = tx.ExecContext(ctx, `INSERT INTO purchase_items(id,purchase_id,position,material_id,material_name_snapshot,purchase_unit_snapshot,consumption_unit_snapshot,purchase_quantity_units,conversion_factor_units,consumption_quantity_units,unit_acquisition_cost_rial,allocated_additional_cost_rial,landed_unit_cost_rial,line_total_rial,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, item.ID, p.ID, i, item.MaterialID, item.MaterialNameSnapshot, item.PurchaseUnitSnapshot, item.ConsumptionUnitSnapshot, item.PurchaseQuantity, item.ConversionFactorSnapshot, item.ConsumptionQuantity, item.UnitAcquisitionCostRial, item.AllocatedAdditionalCostRial, item.LandedUnitCostRial, item.LineTotalRial, item.Notes); e != nil {
			return fail(e)
		}
	}
	if e = tx.Commit(); e != nil {
		return e
	}
	return nil
}
func (s *Store) DeleteDraftPurchase(ctx context.Context, id string) error {
	var status string
	e := s.db.QueryRowContext(ctx, `SELECT status FROM purchases WHERE id=?`, id).Scan(&status)
	if errors.Is(e, sql.ErrNoRows) {
		return domain.ErrPurchaseNotFound
	}
	if e != nil {
		return e
	}
	if status != domain.PurchaseDraft {
		return domain.ErrPurchaseNotDraft
	}
	_, e = s.db.ExecContext(ctx, `DELETE FROM purchases WHERE id=?`, id)
	return e
}

func (s *Store) ArchivePurchase(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE purchases SET archived=1,updated_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.ErrPurchaseNotFound
	}
	return nil
}

func (s *Store) PurchaseHasDependencies(ctx context.Context, id string) (bool, error) {
	queries := []string{
		`SELECT COUNT(*) FROM payment_allocations WHERE target_type='purchase' AND target_id=?`,
		`SELECT COUNT(*) FROM production_consumptions pc JOIN purchase_items pi ON pi.material_id=pc.material_id WHERE pi.purchase_id=?`,
	}
	for _, query := range queries {
		var count int
		if err := s.db.QueryRowContext(ctx, query, id).Scan(&count); err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) PurchaseItemHasDependencies(ctx context.Context, purchaseID, itemID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM production_consumptions pc JOIN purchase_items pi ON pi.material_id=pc.material_id WHERE pi.purchase_id=? AND pi.id=?`, purchaseID, itemID).Scan(&count)
	return count > 0, err
}

func (s *Store) DeletePurchase(ctx context.Context, id string) error {
	blocked, err := s.PurchaseHasDependencies(ctx, id)
	if err != nil {
		return err
	}
	if blocked {
		return domain.ErrPurchaseDeleteProtected
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM purchases WHERE id=?`, id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.ErrPurchaseNotFound
	}
	return nil
}

func (s *Store) PostPurchase(ctx context.Context, id string) error {
	return s.changePurchaseStatus(ctx, id, domain.PurchasePosted)
}
func (s *Store) changePurchaseStatus(ctx context.Context, id, status string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var p domain.Purchase
	var purchaseDate, created, updated string
	e = tx.QueryRowContext(ctx, `SELECT id,purchase_number,supplier_id,supplier_name_snapshot,supplier_code_snapshot,supplier_invoice_number,COALESCE(financial_account_id,''),purchase_date,status,notes,subtotal_rial,discount_rial,shipping_rial,tax_rial,additional_costs_rial,total_rial,created_at,updated_at,archived FROM purchases WHERE id=?`, id).Scan(&p.ID, &p.PurchaseNumber, &p.SupplierID, &p.SupplierNameSnapshot, &p.SupplierCodeSnapshot, &p.SupplierInvoiceNumber, &p.FinancialAccountID, &purchaseDate, &p.Status, &p.Notes, &p.SubtotalRial, &p.DiscountRial, &p.ShippingRial, &p.TaxRial, &p.AdditionalCostsRial, &p.TotalRial, &created, &updated, &p.Archived)
	if errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrPurchaseNotFound)
	}
	if e != nil {
		return fail(e)
	}
	if p.Status != domain.PurchaseDraft {
		if status == domain.PurchasePosted && p.Status == domain.PurchasePosted {
			return tx.Commit()
		}
		return fail(domain.ErrPurchaseAlreadyPosted)
	}
	if e = purchaseTime(&p, purchaseDate, created, updated); e != nil {
		return fail(e)
	}
	rows, e := tx.QueryContext(ctx, `SELECT id,purchase_id,position,material_id,material_name_snapshot,purchase_unit_snapshot,consumption_unit_snapshot,purchase_quantity_units,conversion_factor_units,consumption_quantity_units,unit_acquisition_cost_rial,allocated_additional_cost_rial,landed_unit_cost_rial,line_total_rial,notes FROM purchase_items WHERE purchase_id=? ORDER BY position,id`, id)
	if e != nil {
		return fail(e)
	}
	for rows.Next() {
		var i domain.PurchaseItem
		e = scanPurchaseItem(rows, &i)
		if e != nil {
			rows.Close()
			return fail(e)
		}
		p.Items = append(p.Items, i)
	}
	rows.Close()
	if e = rows.Err(); e != nil {
		return fail(e)
	}
	lines := make([]int64, len(p.Items))
	for i := range p.Items {
		lines[i] = p.Items[i].LineTotalRial
	}
	alloc, e := domain.AllocateLandedCosts(lines, p.DiscountRial, p.ShippingRial, p.TaxRial, p.AdditionalCostsRial)
	if e != nil {
		return fail(e)
	}
	now := time.Now().UTC()
	for i := range p.Items {
		item := &p.Items[i]
		item.AllocatedAdditionalCostRial = alloc[i]
		total := item.LineTotalRial + alloc[i]
		if total < 0 {
			return fail(fmt.Errorf("landed cost cannot be negative for item %s", item.ID))
		}
		item.LandedUnitCostRial = 0
		if item.ConsumptionQuantity > 0 {
			v, e := quantityUnitCost(total, item.ConsumptionQuantity)
			if e != nil {
				return fail(e)
			}
			item.LandedUnitCostRial = v
		}
		if _, e = tx.ExecContext(ctx, `UPDATE purchase_items SET allocated_additional_cost_rial=?,landed_unit_cost_rial=? WHERE id=?`, item.AllocatedAdditionalCostRial, item.LandedUnitCostRial, item.ID); e != nil {
			return fail(e)
		}
		if item.ConsumptionQuantity == 0 {
			continue
		}
		movementID, movementErr := uniqueInventoryMovementID(ctx, tx, "MOV-"+item.ID)
		if movementErr != nil {
			return fail(movementErr)
		}
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, movementID, item.MaterialID, p.PurchaseDate.UTC().Format(time.RFC3339Nano), "purchase", item.ConsumptionQuantity, item.LandedUnitCostRial, total, "purchase", p.ID, "Posted purchase", now.Format(time.RFC3339Nano)); e != nil {
			return fail(e)
		}
	}
	// Inventory/AP recognition is part of this same transaction. It is deliberately
	// limited to purchase posting; revenue, invoices and COGS remain a later slice.
	var inventoryValue int64
	for _, item := range p.Items {
		if inventoryValue > int64(^uint64(0)>>1)-(item.LineTotalRial+item.AllocatedAdditionalCostRial) {
			return fail(fmt.Errorf("purchase accounting amount is too large"))
		}
		inventoryValue += item.LineTotalRial + item.AllocatedAdditionalCostRial
	}
	if inventoryValue <= 0 {
		return fail(fmt.Errorf("posted purchase must have a positive inventory value"))
	}
	entryID := "JE-PUR-" + p.ID
	entry := domain.JournalEntry{ID: entryID, Description: "Posted purchase " + p.PurchaseNumber, SourceType: "purchase", SourceID: p.ID, IdempotencyKey: "purchase:post:" + p.ID, PostedAt: p.PurchaseDate.UTC(), CreatedAt: now, Lines: []domain.JournalLine{
		{ID: entryID + "-L1", JournalEntryID: entryID, Position: 0, AccountID: "ACC-INVENTORY", DebitRial: inventoryValue, PartyType: "supplier", PartyID: p.SupplierID, Memo: "Inventory received"},
		{ID: entryID + "-L2", JournalEntryID: entryID, Position: 1, AccountID: "ACC-AP", CreditRial: inventoryValue, PartyType: "supplier", PartyID: p.SupplierID, Memo: "Supplier payable"},
	}}
	if _, e = s.postJournalTx(ctx, tx, entry); e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `UPDATE purchases SET status=?,accounting_journal_entry_id=?,updated_at=? WHERE id=?`, status, entryID, now.Format(time.RFC3339Nano), id); e != nil {
		return fail(e)
	}
	return tx.Commit()
}
func (s *Store) CancelPurchase(ctx context.Context, id string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var status, journalID string
	if e = tx.QueryRowContext(ctx, `SELECT status,COALESCE(accounting_journal_entry_id,'') FROM purchases WHERE id=?`, id).Scan(&status, &journalID); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrPurchaseNotFound)
	}
	if e != nil {
		return fail(e)
	}
	if status != domain.PurchasePosted {
		if status == domain.PurchaseCancelled {
			return tx.Commit()
		}
		return fail(domain.ErrPurchaseCannotCancel)
	}
	rows, e := tx.QueryContext(ctx, `SELECT m.id,m.material_id,m.quantity_delta_units,m.unit_cost_rial,m.total_cost_rial,m.occurred_at
		FROM inventory_movements m
		WHERE m.reference_type='purchase' AND m.reference_id=?
		  AND NOT EXISTS (
			SELECT 1 FROM inventory_movements c
			WHERE c.reference_type='purchase_cancel'
			  AND (c.id='MOV-CANCEL-' || m.id OR c.id LIKE 'MOV-CANCEL-' || m.id || '-%')
		  )
		ORDER BY m.id`, id)
	if e != nil {
		return fail(e)
	}
	type mv struct {
		id, m string
		q     int64
		c, t  int64
		at    string
	}
	var movements []mv
	for rows.Next() {
		var v mv
		if e = rows.Scan(&v.id, &v.m, &v.q, &v.c, &v.t, &v.at); e != nil {
			rows.Close()
			return fail(e)
		}
		movements = append(movements, v)
	}
	rows.Close()
	// Cancellation removes physical stock but does not release unrelated
	// reservations. Preflight the complete batch so the transaction cannot
	// leave active reservations above physical stock.
	cancelByMaterial := map[string]int64{}
	for _, v := range movements {
		if v.q < 0 || cancelByMaterial[v.m] > int64(^uint64(0)>>1)-v.q {
			return fail(domain.ErrInsufficientStock)
		}
		cancelByMaterial[v.m] += v.q
	}
	for materialID, cancelQty := range cancelByMaterial {
		state, stateErr := inventoryStateTx(ctx, tx, materialID)
		if stateErr != nil {
			return fail(stateErr)
		}
		if domain.Quantity(cancelQty) > state.PhysicalStock || state.ReservedStock > state.PhysicalStock-domain.Quantity(cancelQty) {
			return fail(domain.ErrInsufficientStock)
		}
	}
	remaining := map[string]int64{}
	for _, v := range movements {
		var stock int64
		if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_delta_units),0) FROM inventory_movements WHERE material_id=?`, v.m).Scan(&stock); e != nil {
			return fail(e)
		}
		if _, ok := remaining[v.m]; !ok {
			remaining[v.m] = stock
		}
		if remaining[v.m]-v.q < 0 {
			return fail(domain.ErrInsufficientStock)
		}
		remaining[v.m] -= v.q
		movementID, movementErr := uniqueInventoryMovementID(ctx, tx, "MOV-CANCEL-"+v.id)
		if movementErr != nil {
			return fail(movementErr)
		}
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, movementID, v.m, time.Now().UTC().Format(time.RFC3339Nano), "supplier_return", -v.q, v.c, -v.t, "purchase_cancel", id, "Cancelled purchase", time.Now().UTC().Format(time.RFC3339Nano)); e != nil {
			return fail(e)
		}
	}
	if journalID != "" {
		if _, e = s.reverseJournalTx(ctx, tx, journalID, "purchase:cancel:"+id, "Cancellation of purchase "+id, time.Now().UTC()); e != nil {
			return fail(e)
		}
	}
	if _, e = tx.ExecContext(ctx, `UPDATE purchases SET status='Cancelled',updated_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), id); e != nil {
		return fail(e)
	}
	return tx.Commit()
}

// uniqueInventoryMovementID preserves the stable movement IDs used by the
// original posting path, but gives a reopened/reposted purchase a fresh ID
// when old immutable history already occupies that ID.
func uniqueInventoryMovementID(ctx context.Context, tx *sql.Tx, preferred string) (string, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_movements WHERE id=?`, preferred).Scan(&count); err != nil {
		return "", err
	}
	if count == 0 {
		return preferred, nil
	}
	for attempt := 0; attempt < 10; attempt++ {
		candidate := fmt.Sprintf("%s-%d-%d", preferred, time.Now().UnixNano(), attempt)
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_movements WHERE id=?`, candidate).Scan(&count); err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique inventory movement ID")
}

func (s *Store) ListInventoryMovements(ctx context.Context, materialID string) ([]domain.InventoryMovement, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at
		FROM inventory_movements m
		WHERE m.material_id=?
		  AND NOT EXISTS (
			SELECT 1 FROM inventory_movements c
			WHERE c.reference_type='manual_adjustment_cancel'
			  AND ((m.reference_type='manual_adjustment' AND c.reference_id=m.id)
			   OR (m.reference_type='manual_adjustment_cancel' AND c.reference_id=m.reference_id))
		  )
		ORDER BY occurred_at DESC,id DESC`, materialID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.InventoryMovement{}
	for rows.Next() {
		v, e := scanMovement(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) CancelInventoryMovement(ctx context.Context, id string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var materialID, referenceType string
	var quantity, unitCost, total int64
	if e = tx.QueryRowContext(ctx, `SELECT material_id,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type FROM inventory_movements WHERE id=?`, id).Scan(&materialID, &quantity, &unitCost, &total, &referenceType); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrMovementNotFound)
	}
	if e != nil {
		return fail(e)
	}
	if referenceType != "manual_adjustment" {
		return fail(domain.ErrMovementCannotCancel)
	}
	var canceled int
	if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_movements WHERE reference_type='manual_adjustment_cancel' AND reference_id=?`, id).Scan(&canceled); e != nil {
		return fail(e)
	}
	if canceled > 0 {
		return fail(domain.ErrMovementAlreadyCanceled)
	}
	state, e := inventoryStateTx(ctx, tx, materialID)
	if e != nil {
		return fail(e)
	}
	reversedQuantity := -quantity
	remainingStock := int64(state.PhysicalStock) + reversedQuantity
	if remainingStock < 0 || int64(state.ReservedStock) > remainingStock {
		return fail(domain.ErrInsufficientStock)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	movementID, e := uniqueInventoryMovementID(ctx, tx, "MOV-CANCEL-"+id)
	if e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, movementID, materialID, now, "adjustment", reversedQuantity, unitCost, -total, "manual_adjustment_cancel", id, "Cancelled manual adjustment", now); e != nil {
		return fail(e)
	}
	return tx.Commit()
}
func (s *Store) AdjustInventory(ctx context.Context, materialID string, qty domain.Quantity, cost int64, note string) error {
	if cost < 0 {
		return fmt.Errorf("unit cost cannot be negative")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var stock int64
	if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_delta_units),0) FROM inventory_movements WHERE material_id=?`, materialID).Scan(&stock); e != nil {
		return fail(e)
	}
	newStock := new(big.Int).Add(big.NewInt(stock), big.NewInt(int64(qty)))
	if !newStock.IsInt64() || newStock.Sign() < 0 {
		return fail(domain.ErrInsufficientStock)
	}
	if qty < 0 && cost == 0 {
		var value int64
		if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_cost_rial),0) FROM inventory_movements WHERE material_id=?`, materialID).Scan(&value); e != nil {
			return fail(e)
		}
		if stock > 0 {
			var a, e2 = quantityUnitCost(value, domain.Quantity(stock))
			if e2 != nil {
				return fail(e2)
			}
			cost = a
		}
	}
	total, e := domain.MulQuantityRial(absQuantity(qty), cost)
	if e != nil {
		return fail(e)
	}
	if qty < 0 {
		total = -total
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-ADJ-"+fmt.Sprint(time.Now().UnixNano()), materialID, now, "adjustment", qty, cost, total, "manual_adjustment", "", note, now)
	if e != nil {
		return fail(e)
	}
	return tx.Commit()
}

func roundDivide(n, d int64) (int64, error) {
	if d <= 0 {
		return 0, fmt.Errorf("division denominator must be positive")
	}
	x := new(big.Int).Add(big.NewInt(n), big.NewInt(d/2))
	x.Quo(x, big.NewInt(d))
	if !x.IsInt64() {
		return 0, fmt.Errorf("cost is too large")
	}
	return x.Int64(), nil
}

func quantityUnitCost(total int64, quantity domain.Quantity) (int64, error) {
	if quantity <= 0 {
		return 0, fmt.Errorf("quantity must be positive")
	}
	x := new(big.Int).Mul(big.NewInt(total), big.NewInt(domain.QuantityScale))
	if x.Sign() >= 0 {
		x.Add(x, big.NewInt(int64(quantity)/2))
	} else {
		x.Sub(x, big.NewInt(int64(quantity)/2))
	}
	x.Quo(x, big.NewInt(int64(quantity)))
	if !x.IsInt64() {
		return 0, fmt.Errorf("unit cost is too large")
	}
	return x.Int64(), nil
}
func absQuantity(q domain.Quantity) domain.Quantity {
	if q < 0 {
		return -q
	}
	return q
}
func purchaseTime(p *domain.Purchase, purchaseDate, created, updated string) error {
	var e error
	p.PurchaseDate, e = time.Parse(time.RFC3339Nano, purchaseDate)
	if e != nil {
		return e
	}
	p.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return e
	}
	p.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	return e
}
func scanSupplier(row scanner) (domain.Supplier, error) {
	var v domain.Supplier
	var active int
	var c, u string
	if e := row.Scan(&v.ID, &v.Name, &v.Code, &v.Phone, &v.Email, &v.Address, &v.Notes, &active, &c, &u); e != nil {
		return v, e
	}
	v.Active = active == 1
	var e error
	v.CreatedAt, e = time.Parse(time.RFC3339Nano, c)
	if e != nil {
		return v, e
	}
	v.UpdatedAt, e = time.Parse(time.RFC3339Nano, u)
	return v, e
}
func scanPurchase(row scanner) (domain.Purchase, error) {
	var p domain.Purchase
	var d, c, u string
	if e := row.Scan(&p.ID, &p.PurchaseNumber, &p.SupplierID, &p.SupplierNameSnapshot, &p.SupplierCodeSnapshot, &p.SupplierInvoiceNumber, &p.FinancialAccountID, &d, &p.Status, &p.Notes, &p.SubtotalRial, &p.DiscountRial, &p.ShippingRial, &p.TaxRial, &p.AdditionalCostsRial, &p.TotalRial, &c, &u, &p.Archived); e != nil {
		return p, e
	}
	return p, purchaseTime(&p, d, c, u)
}
func (s *Store) loadPurchaseItems(ctx context.Context, p *domain.Purchase) error {
	rows, e := s.db.QueryContext(ctx, `SELECT id,purchase_id,position,material_id,material_name_snapshot,purchase_unit_snapshot,consumption_unit_snapshot,purchase_quantity_units,conversion_factor_units,consumption_quantity_units,unit_acquisition_cost_rial,allocated_additional_cost_rial,landed_unit_cost_rial,line_total_rial,notes FROM purchase_items WHERE purchase_id=? ORDER BY position,id`, p.ID)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var i domain.PurchaseItem
		if e = scanPurchaseItem(rows, &i); e != nil {
			return e
		}
		p.Items = append(p.Items, i)
	}
	return rows.Err()
}
func scanPurchaseItem(row scanner, i *domain.PurchaseItem) error {
	var q, f, c int64
	if e := row.Scan(&i.ID, &i.PurchaseID, &i.Position, &i.MaterialID, &i.MaterialNameSnapshot, &i.PurchaseUnitSnapshot, &i.ConsumptionUnitSnapshot, &q, &f, &c, &i.UnitAcquisitionCostRial, &i.AllocatedAdditionalCostRial, &i.LandedUnitCostRial, &i.LineTotalRial, &i.Notes); e != nil {
		return e
	}
	i.PurchaseQuantity = domain.Quantity(q)
	i.ConversionFactorSnapshot = domain.Quantity(f)
	i.ConsumptionQuantity = domain.Quantity(c)
	return nil
}
func scanMovement(row scanner) (domain.InventoryMovement, error) {
	var v domain.InventoryMovement
	var at, c string
	var q int64
	if e := row.Scan(&v.ID, &v.MaterialID, &at, &v.MovementType, &q, &v.UnitCostRial, &v.TotalCostRial, &v.ReferenceType, &v.ReferenceID, &v.Note, &c); e != nil {
		return v, e
	}
	v.QuantityDelta = domain.Quantity(q)
	var e error
	v.OccurredAt, e = time.Parse(time.RFC3339Nano, at)
	if e != nil {
		return v, e
	}
	v.CreatedAt, e = time.Parse(time.RFC3339Nano, c)
	return v, e
}
