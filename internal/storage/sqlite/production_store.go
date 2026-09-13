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

func (s *Store) inventoryState(ctx context.Context, materialID string) (domain.InventorySummary, error) {
	var quantity, value, reserved int64
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_delta_units),0),COALESCE(SUM(total_cost_rial),0),COALESCE((SELECT SUM(quantity_units) FROM inventory_reservations WHERE material_id=? AND status='active'),0) FROM inventory_movements WHERE material_id=?`, materialID, materialID).Scan(&quantity, &value, &reserved); err != nil {
		return domain.InventorySummary{}, err
	}
	return inventoryStateValues(quantity, value, reserved)
}

func inventoryStateTx(ctx context.Context, tx *sql.Tx, materialID string) (domain.InventorySummary, error) {
	var quantity, value, reserved int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_delta_units),0),COALESCE(SUM(total_cost_rial),0),COALESCE((SELECT SUM(quantity_units) FROM inventory_reservations WHERE material_id=? AND status='active'),0) FROM inventory_movements WHERE material_id=?`, materialID, materialID).Scan(&quantity, &value, &reserved); err != nil {
		return domain.InventorySummary{}, err
	}
	return inventoryStateValues(quantity, value, reserved)
}

func inventoryStateValues(quantity, value, reserved int64) (domain.InventorySummary, error) {
	if quantity < 0 {
		return domain.InventorySummary{}, fmt.Errorf("inventory ledger is negative")
	}
	if reserved < 0 || reserved > quantity {
		return domain.InventorySummary{}, fmt.Errorf("active reservations exceed physical stock")
	}
	avg := int64(0)
	if quantity > 0 {
		x := new(big.Int).Mul(big.NewInt(value), big.NewInt(domain.QuantityScale))
		x.Add(x, big.NewInt(quantity/2))
		x.Quo(x, big.NewInt(quantity))
		if !x.IsInt64() {
			return domain.InventorySummary{}, fmt.Errorf("average cost is too large")
		}
		avg = x.Int64()
	}
	return domain.InventorySummary{PhysicalStock: domain.Quantity(quantity), ReservedStock: domain.Quantity(reserved), AvailableStock: domain.Quantity(quantity - reserved), AverageUnitCostRial: avg, InventoryValueRial: value}, nil
}

func (s *Store) InventoryState(ctx context.Context, materialID string) (domain.InventorySummary, error) {
	return s.inventoryState(ctx, materialID)
}

func (s *Store) ProductionSummary(ctx context.Context, orderID string) (int, int, int, error) {
	var total, completed, inProgress int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(CASE WHEN status='Completed' THEN 1 ELSE 0 END),0), COALESCE(SUM(CASE WHEN status IN ('In Progress','Paused') THEN 1 ELSE 0 END),0) FROM production_jobs WHERE order_id=?`, orderID).Scan(&total, &completed, &inProgress)
	return total, completed, inProgress, err
}

func (s *Store) ProductionCostSummary(ctx context.Context, orderID string) (int64, error) {
	var cost int64
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(COALESCE((SELECT SUM(CASE WHEN im.movement_type IN ('production_consumption','waste') THEN -im.total_cost_rial ELSE 0 END) FROM inventory_movements im JOIN production_consumptions pc ON pc.id=im.reference_id WHERE pc.production_job_id=pj.id AND im.reference_type IN ('production_consumption','production_correction')),0)+pj.actual_outsourced_cost_rial),0) FROM production_jobs pj WHERE pj.order_id=?`, orderID).Scan(&cost)
	return cost, err
}

func (s *Store) ListReservations(ctx context.Context, materialID, jobID, orderID string) ([]domain.InventoryReservation, error) {
	query := `SELECT id,material_id,COALESCE(order_id,''),COALESCE(order_item_id,''),COALESCE(production_job_id,''),quantity_units,status,created_at,updated_at FROM inventory_reservations WHERE 1=1`
	args := []any{}
	if materialID != "" {
		query += ` AND material_id=?`
		args = append(args, materialID)
	}
	if jobID != "" {
		query += ` AND production_job_id=?`
		args = append(args, jobID)
	}
	if orderID != "" {
		query += ` AND order_id=?`
		args = append(args, orderID)
	}
	query += ` ORDER BY created_at,id`
	rows, e := s.db.QueryContext(ctx, query, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.InventoryReservation{}
	for rows.Next() {
		v, e := scanReservation(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) CreateReservation(ctx context.Context, r domain.InventoryReservation) error {
	if err := r.Validate(); err != nil {
		return err
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	if r.ProductionJobID != "" {
		var orderID, itemID, status string
		var qty, outsource domain.Quantity
		if e = tx.QueryRowContext(ctx, `SELECT order_id,order_item_id,status,quantity_units,outsource_quantity_units FROM production_jobs WHERE id=?`, r.ProductionJobID).Scan(&orderID, &itemID, &status, &qty, &outsource); e != nil {
			return fail(e)
		}
		if status == "Completed" || status == "Cancelled" || qty <= outsource {
			return fail(domain.ErrProductionNotEditable)
		}
		if (r.OrderID != "" && r.OrderID != orderID) || (r.OrderItemID != "" && r.OrderItemID != itemID) {
			return fail(fmt.Errorf("reservation must belong to the production order item"))
		}
		r.OrderID = orderID
		r.OrderItemID = itemID
	}
	state, e := inventoryStateTx(ctx, tx, r.MaterialID)
	if e != nil {
		return fail(e)
	}
	if r.Quantity > state.AvailableStock {
		return fail(domain.ErrReservationExceeded)
	}
	if r.Status == "" {
		r.Status = domain.ReservationActive
	}
	now := time.Now().UTC()
	if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_reservations(id,material_id,order_id,order_item_id,production_job_id,quantity_units,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, r.ID, r.MaterialID, nullableString(r.OrderID), nullableString(r.OrderItemID), nullableString(r.ProductionJobID), r.Quantity, r.Status, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); e != nil {
		return fail(e)
	}
	return tx.Commit()
}

func (s *Store) UpdateReservation(ctx context.Context, id string, quantity domain.Quantity) error {
	if quantity <= 0 {
		return fmt.Errorf("reservation quantity must be positive")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var material, status string
	if e = tx.QueryRowContext(ctx, `SELECT material_id,status FROM inventory_reservations WHERE id=?`, id).Scan(&material, &status); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrReservationNotFound)
	}
	if e != nil {
		return fail(e)
	}
	if status != domain.ReservationActive {
		return fail(fmt.Errorf("only active reservations can be edited"))
	}
	state, e := inventoryStateTx(ctx, tx, material)
	if e != nil {
		return fail(e)
	}
	var old domain.Quantity
	if e = tx.QueryRowContext(ctx, `SELECT quantity_units FROM inventory_reservations WHERE id=?`, id).Scan(&old); e != nil {
		return fail(e)
	}
	if quantity > state.AvailableStock+old {
		return fail(domain.ErrReservationExceeded)
	}
	if e = adjustManagedReservationTx(ctx, tx, id, quantity); e != nil {
		return fail(e)
	}
	_, e = tx.ExecContext(ctx, `UPDATE inventory_reservations SET quantity_units=?,updated_at=? WHERE id=?`, quantity, time.Now().UTC().Format(time.RFC3339Nano), id)
	if e != nil {
		return fail(e)
	}
	return tx.Commit()
}

func (s *Store) ReleaseReservation(ctx context.Context, id, status string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = adjustManagedReservationTx(ctx, tx, id, 0); e != nil {
		return e
	}
	if status != "released" && status != "cancelled" {
		status = domain.ReservationReleased
	}
	res, e := tx.ExecContext(ctx, `UPDATE inventory_reservations SET status=?,updated_at=? WHERE id=? AND status='active'`, status, time.Now().UTC().Format(time.RFC3339Nano), id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrReservationNotFound
	}
	return tx.Commit()
}

func (s *Store) ListProductionJobs(ctx context.Context, status string) ([]domain.ProductionJob, error) {
	q := `SELECT id,job_number,order_id,order_item_id,service_name_snapshot,quantity_units,quantity_unit,COALESCE(assigned_machine_id,''),status,priority,planned_at,started_at,completed_at,notes,estimated_cost_rial,actual_material_cost_rial,actual_waste_cost_rial,actual_outsourced_cost_rial,COALESCE(outsource_supplier_id,''),outsource_description,outsource_quoted_cost_rial,COALESCE(outsource_sent_at,''),COALESCE(outsource_expected_return_at,''),COALESCE(outsource_received_at,''),outsource_notes,created_at,updated_at FROM production_jobs`
	args := []any{}
	if status != "" && status != "All" {
		q += ` WHERE status=?`
		args = append(args, status)
	}
	q += ` ORDER BY CASE priority WHEN 'Urgent' THEN 0 WHEN 'High' THEN 1 WHEN 'Normal' THEN 2 ELSE 3 END,planned_at,id`
	rows, e := s.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	out := []domain.ProductionJob{}
	for rows.Next() {
		v, e := scanProductionJob(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, v)
	}
	if e = rows.Err(); e != nil {
		rows.Close()
		return nil, e
	}
	if e = rows.Close(); e != nil {
		return nil, e
	}
	for i := range out {
		out[i], e = s.withProductionActuals(ctx, out[i])
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
func (s *Store) GetProductionJob(ctx context.Context, id string) (domain.ProductionJob, error) {
	v, e := scanProductionJob(s.db.QueryRowContext(ctx, `SELECT id,job_number,order_id,order_item_id,service_name_snapshot,quantity_units,quantity_unit,COALESCE(assigned_machine_id,''),status,priority,planned_at,started_at,completed_at,notes,estimated_cost_rial,actual_material_cost_rial,actual_waste_cost_rial,actual_outsourced_cost_rial,COALESCE(outsource_supplier_id,''),outsource_description,outsource_quoted_cost_rial,COALESCE(outsource_sent_at,''),COALESCE(outsource_expected_return_at,''),COALESCE(outsource_received_at,''),outsource_notes,created_at,updated_at FROM production_jobs WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.ProductionJob{}, domain.ErrProductionJobNotFound
	}
	if e != nil {
		return v, e
	}
	return s.withProductionActuals(ctx, v)
}
func (s *Store) withProductionActuals(ctx context.Context, j domain.ProductionJob) (domain.ProductionJob, error) {
	if e := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN m.movement_type='production_consumption' THEN -m.total_cost_rial ELSE 0 END),0),COALESCE(SUM(CASE WHEN m.movement_type='waste' THEN -m.total_cost_rial ELSE 0 END),0) FROM inventory_movements m JOIN production_consumptions c ON c.id=m.reference_id WHERE c.production_job_id=? AND m.reference_type IN ('production_consumption','production_correction')`, j.ID).Scan(&j.ActualMaterialCostRial, &j.ActualWasteCostRial); e != nil {
		return j, e
	}
	if err := s.db.QueryRowContext(ctx, `SELECT outsource_quantity_units,outsource_unit_cost_rial,outsource_financial_account_id FROM production_jobs WHERE id=?`, j.ID).Scan(&j.OutsourceQuantity, &j.OutsourceUnitCostRial, &j.OutsourceFinancialAccountID); err != nil {
		return j, err
	}
	return s.withProductionForecast(ctx, j)
}

func (s *Store) CreateProductionJob(ctx context.Context, j domain.ProductionJob) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var commercial, fulfillment string
	if e = tx.QueryRowContext(ctx, `SELECT commercial_status,fulfillment_status FROM orders WHERE id=?`, j.OrderID).Scan(&commercial, &fulfillment); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrOrderNotFound)
	}
	if e != nil {
		return fail(e)
	}
	if commercial != string(domain.CommercialConfirmed) {
		return fail(fmt.Errorf("production requires a confirmed order"))
	}
	var service string
	var qty int64
	var unit string
	var estimated int64
	if e = tx.QueryRowContext(ctx, `SELECT service_name_snapshot,quantity_units,quantity_unit,estimated_cost_rial FROM order_items WHERE id=? AND order_id=? AND removed_at IS NULL`, j.OrderItemID, j.OrderID).Scan(&service, &qty, &unit, &estimated); e != nil {
		return fail(fmt.Errorf("order item: %w", e))
	}
	j.ServiceNameSnapshot = service
	j.Quantity = domain.Quantity(qty)
	j.QuantityUnit = unit
	var duplicate int
	if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM production_jobs WHERE order_item_id=? AND status<>'Cancelled'`, j.OrderItemID).Scan(&duplicate); e != nil {
		return fail(e)
	}
	if duplicate > 0 {
		return fail(fmt.Errorf("this order item already has a production job"))
	}
	if j.EstimatedCostRial == 0 {
		j.EstimatedCostRial = estimated
	}
	if e = j.Validate(); e != nil {
		return fail(e)
	}
	var number int64
	if e = tx.QueryRowContext(ctx, `SELECT next_number FROM production_number_sequences WHERE id=1`).Scan(&number); e != nil {
		return fail(e)
	}
	j.JobNumber = fmt.Sprintf("JOB-%04d", number)
	if _, e = tx.ExecContext(ctx, `UPDATE production_number_sequences SET next_number=next_number+1 WHERE id=1`); e != nil {
		return fail(e)
	}
	now := time.Now().UTC()
	j.CreatedAt = now
	j.UpdatedAt = now
	if _, e = tx.ExecContext(ctx, `INSERT INTO production_jobs(id,job_number,order_id,order_item_id,service_name_snapshot,quantity_units,quantity_unit,assigned_machine_id,status,priority,planned_at,started_at,completed_at,notes,estimated_cost_rial,actual_material_cost_rial,actual_waste_cost_rial,actual_outsourced_cost_rial,outsource_supplier_id,outsource_description,outsource_quoted_cost_rial,outsource_sent_at,outsource_expected_return_at,outsource_received_at,outsource_notes,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, j.ID, j.JobNumber, j.OrderID, j.OrderItemID, j.ServiceNameSnapshot, j.Quantity, j.QuantityUnit, nullableString(j.AssignedMachineID), j.Status, j.Priority, nullableTime(j.PlannedAt), nullableTime(j.StartedAt), nullableTime(j.CompletedAt), j.Notes, j.EstimatedCostRial, 0, 0, 0, nullableString(j.OutsourceSupplierID), j.OutsourceDescription, j.OutsourceQuotedCostRial, nullableString(j.OutsourceSentAt), nullableString(j.OutsourceExpectedReturnAt), nullableString(j.OutsourceReceivedAt), j.OutsourceNotes, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); e != nil {
		return fail(e)
	}
	if e = reserveOrderMaterialsTx(ctx, tx, j); e != nil {
		return fail(fmt.Errorf("reserve order materials: %w", e))
	}
	if fulfillment == string(domain.FulfillmentPending) {
		if _, e = tx.ExecContext(ctx, `UPDATE orders SET fulfillment_status=?,updated_at=? WHERE id=?`, domain.FulfillmentInProduction, now.Format(time.RFC3339Nano), j.OrderID); e != nil {
			return fail(e)
		}
	}
	return tx.Commit()
}

func (s *Store) UpdateProductionJob(ctx context.Context, j domain.ProductionJob) error {
	if !domain.ValidProductionStatus(j.Status) {
		return domain.ErrProductionTransition
	}
	res, e := s.db.ExecContext(ctx, `UPDATE production_jobs SET assigned_machine_id=?,priority=?,notes=?,actual_outsourced_cost_rial=?,outsource_supplier_id=?,outsource_description=?,outsource_quoted_cost_rial=?,outsource_sent_at=?,outsource_expected_return_at=?,outsource_received_at=?,outsource_notes=?,updated_at=? WHERE id=? AND status NOT IN ('Completed','Cancelled')`, nullableString(j.AssignedMachineID), j.Priority, j.Notes, j.ActualOutsourcedCostRial, nullableString(j.OutsourceSupplierID), j.OutsourceDescription, j.OutsourceQuotedCostRial, nullableString(j.OutsourceSentAt), nullableString(j.OutsourceExpectedReturnAt), nullableString(j.OutsourceReceivedAt), j.OutsourceNotes, time.Now().UTC().Format(time.RFC3339Nano), j.ID)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrProductionNotEditable
	}
	return nil
}

func (s *Store) UpdateProductionOutsourcing(ctx context.Context, j domain.ProductionJob) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var qty, oldOutsource domain.Quantity
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT quantity_units,outsource_quantity_units,status FROM production_jobs WHERE id=?`, j.ID).Scan(&qty, &oldOutsource, &status); err != nil {
		return err
	}
	if status == "Cancelled" || status == "Completed" {
		return domain.ErrProductionNotEditable
	}
	if j.OutsourceQuantity < 0 || j.OutsourceQuantity > qty || j.OutsourceUnitCostRial < 0 {
		return fmt.Errorf("outsourced quantity must be between zero and the job quantity")
	}
	total, err := domain.MulQuantityRial(j.OutsourceQuantity, j.OutsourceUnitCostRial)
	if err != nil {
		return err
	}
	if j.OutsourceQuantity > 0 && total <= 0 {
		return fmt.Errorf("enter an outsource unit cost greater than zero")
	}
	j.ActualOutsourcedCostRial = total
	if j.OutsourceQuantity > oldOutsource {
		// Preserve original movements; compensate and re-post only the retained share.
		rows, err := tx.QueryContext(ctx, `SELECT c.id,c.material_id,c.consumed_quantity_units,c.waste_quantity_units,c.notes FROM production_consumptions c WHERE c.production_job_id=? AND NOT EXISTS(SELECT 1 FROM inventory_movements m WHERE m.reference_id=c.id AND m.reference_type='production_correction') ORDER BY c.created_at,c.id`, j.ID)
		if err != nil {
			return err
		}
		type usage struct {
			id, material, note string
			consumed, waste    domain.Quantity
		}
		items := []usage{}
		for rows.Next() {
			var u usage
			if err = rows.Scan(&u.id, &u.material, &u.consumed, &u.waste, &u.note); err != nil {
				rows.Close()
				return err
			}
			items = append(items, u)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for _, u := range items {
			cons, err := scaleProductionQuantity(u.consumed, qty-j.OutsourceQuantity, qty-oldOutsource)
			if err != nil {
				return err
			}
			waste, err := scaleProductionQuantity(u.waste, qty-j.OutsourceQuantity, qty-oldOutsource)
			if err != nil {
				return err
			}
			if err = s.reverseProductionConsumptionTx(ctx, tx, u.id, "Material returned for outsourced share"); err != nil {
				return err
			}
			if cons+waste > 0 {
				if _, err = s.recordProductionConsumptionTx(ctx, tx, j.ID, u.material, fmt.Sprintf("outsource:%s:%d", u.id, j.OutsourceQuantity), cons, waste, u.note); err != nil {
					return err
				}
			}
		}
		// Manual allocations are scaled too; order-derived allocations are reconciled below.
		rows, err = tx.QueryContext(ctx, `SELECT id,quantity_units FROM inventory_reservations WHERE production_job_id=? AND status='active' AND id NOT IN (SELECT reservation_id FROM production_material_plans WHERE production_job_id=?)`, j.ID, j.ID)
		if err != nil {
			return err
		}
		type allocation struct {
			id  string
			qty domain.Quantity
		}
		allocations := []allocation{}
		for rows.Next() {
			var a allocation
			if err = rows.Scan(&a.id, &a.qty); err != nil {
				rows.Close()
				return err
			}
			allocations = append(allocations, a)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for _, a := range allocations {
			q, err := scaleProductionQuantity(a.qty, qty-j.OutsourceQuantity, qty-oldOutsource)
			if err != nil {
				return err
			}
			if q == 0 {
				_, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='released' WHERE id=?`, a.id)
			} else {
				_, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET quantity_units=? WHERE id=?`, q, a.id)
			}
			if err != nil {
				return err
			}
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.ExecContext(ctx, `UPDATE production_jobs SET outsource_quantity_units=?,outsource_unit_cost_rial=?,outsource_financial_account_id=?,actual_outsourced_cost_rial=?,outsource_supplier_id=?,outsource_description=?,outsource_quoted_cost_rial=?,outsource_sent_at=?,outsource_expected_return_at=?,outsource_received_at=?,outsource_notes=?,updated_at=? WHERE id=?`, j.OutsourceQuantity, j.OutsourceUnitCostRial, j.OutsourceFinancialAccountID, total, nullableString(j.OutsourceSupplierID), j.OutsourceDescription, j.OutsourceQuotedCostRial, nullableString(j.OutsourceSentAt), nullableString(j.OutsourceExpectedReturnAt), nullableString(j.OutsourceReceivedAt), j.OutsourceNotes, now, j.ID)
	if err != nil {
		return err
	}
	if err = syncProductionMaterialsTx(ctx, tx, j.ID); err != nil {
		return err
	}
	if err = s.reconcileOutsourceExpenseTx(ctx, tx, j); err != nil {
		return err
	}
	if err = s.reconcileJobCOGSTx(ctx, tx, j.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ensureOrderMaterials(ctx context.Context, jobID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = syncProductionMaterialsTx(ctx, tx, jobID); err != nil {
		return err
	}
	return tx.Commit()
}

func reserveOrderMaterialsTx(ctx context.Context, tx *sql.Tx, j domain.ProductionJob) error {
	return syncProductionMaterialsTx(ctx, tx, j.ID)
}

func multiplyProductionQuantity(left, right domain.Quantity) (domain.Quantity, error) {
	value := new(big.Int).Mul(big.NewInt(int64(left)), big.NewInt(int64(right)))
	value.Add(value, big.NewInt(domain.QuantityScale/2))
	value.Quo(value, big.NewInt(domain.QuantityScale))
	if !value.IsInt64() {
		return 0, fmt.Errorf("quantity is too large")
	}
	return domain.Quantity(value.Int64()), nil
}

func (s *Store) TransitionProductionJob(ctx context.Context, id, status string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var current, orderID string
	if e = tx.QueryRowContext(ctx, `SELECT status,order_id FROM production_jobs WHERE id=?`, id).Scan(&current, &orderID); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrProductionJobNotFound)
	}
	if e != nil {
		return fail(e)
	}
	if !domain.ValidProductionTransition(current, status) {
		return fail(domain.ErrProductionTransition)
	}
	if current == status {
		return tx.Commit()
	}
	// Reopening is explicit and cannot silently replace another job for this item.
	if current == domain.ProductionCancelled && status != domain.ProductionCancelled {
		var duplicates int
		if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM production_jobs WHERE order_item_id=(SELECT order_item_id FROM production_jobs WHERE id=?) AND id<>? AND status<>'Cancelled'`, id, id).Scan(&duplicates); e != nil {
			return fail(e)
		}
		if duplicates > 0 {
			return fail(fmt.Errorf("this order item already has an active production job"))
		}
	}
	now := time.Now().UTC()
	var started, completed any
	if status == domain.ProductionInProgress {
		started = now.Format(time.RFC3339Nano)
	}
	if status == domain.ProductionCompleted {
		if current != domain.ProductionInProgress && current != domain.ProductionPaused {
			return fail(fmt.Errorf("start production before completing the job"))
		}
		if e = s.completeProductionMaterialsTx(ctx, tx, id); e != nil {
			return fail(e)
		}
		completed = now.Format(time.RFC3339Nano)
		if _, e = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='released',updated_at=? WHERE production_job_id=? AND status='active'`, now.Format(time.RFC3339Nano), id); e != nil {
			return fail(e)
		}
	}
	if status == domain.ProductionCancelled {
		if _, e = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='cancelled',updated_at=? WHERE production_job_id=? AND status='active'`, now.Format(time.RFC3339Nano), id); e != nil {
			return fail(e)
		}
	}
	if _, e = tx.ExecContext(ctx, `UPDATE production_jobs SET status=?,started_at=COALESCE(started_at,?),completed_at=?,updated_at=?,produced_quantity_units=CASE WHEN ?='Completed' THEN MAX(produced_quantity_units,quantity_units) ELSE produced_quantity_units END WHERE id=?`, status, started, completed, now.Format(time.RFC3339Nano), status, id); e != nil {
		return fail(e)
	}
	if current == domain.ProductionCompleted && status != domain.ProductionCancelled {
		// Synchronize after changing status so the completed-job guard no longer
		// applies. Historical usage/corrections remain the authoritative actuals.
		if e = reopenProductionMaterialsTx(ctx, tx, id); e != nil {
			return fail(e)
		}
	}
	if e = syncProductionFulfillmentTx(ctx, tx, orderID, now); e != nil {
		return fail(e)
	}
	if e = s.reconcileJobCOGSTx(ctx, tx, id); e != nil {
		return fail(e)
	}
	return tx.Commit()
}

func (s *Store) DeleteProductionJob(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = s.deleteProductionJobTx(ctx, tx, id); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) deleteProductionJobTx(ctx context.Context, tx *sql.Tx, id string) error {
	var e error
	fail := func(e error) error { return e }
	var orderID string
	if e = tx.QueryRowContext(ctx, `SELECT order_id FROM production_jobs WHERE id=?`, id).Scan(&orderID); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrProductionJobNotFound)
	}
	if e != nil {
		return fail(e)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	rows, e := tx.QueryContext(ctx, `SELECT pc.id,pc.material_id,pc.consumed_quantity_units,pc.waste_quantity_units,pc.unit_cost_rial,pc.material_cost_rial,pc.waste_cost_rial,EXISTS(SELECT 1 FROM inventory_movements im WHERE im.reference_type='production_correction' AND im.reference_id=pc.id) FROM production_consumptions pc WHERE pc.production_job_id=? ORDER BY pc.created_at,pc.id`, id)
	if e != nil {
		return fail(e)
	}
	type consumption struct {
		id, material                                   string
		consumed, waste, cost, materialCost, wasteCost int64
		corrected                                      bool
	}
	consumptions := make([]consumption, 0)
	for rows.Next() {
		var item consumption
		var corrected int
		if e = rows.Scan(&item.id, &item.material, &item.consumed, &item.waste, &item.cost, &item.materialCost, &item.wasteCost, &corrected); e != nil {
			rows.Close()
			return fail(e)
		}
		item.corrected = corrected == 1
		consumptions = append(consumptions, item)
	}
	if e = rows.Err(); e != nil {
		rows.Close()
		return fail(e)
	}
	rows.Close()
	for _, item := range consumptions {
		if item.corrected {
			continue
		}
		if item.consumed > 0 {
			total := item.materialCost
			if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-DELETE-"+item.id+"-C", item.material, now, "production_consumption", item.consumed, item.cost, total, "production_correction", item.id, "Compensating movement for deleted production job "+id, now); e != nil {
				return fail(e)
			}
		}
		if item.waste > 0 {
			total := item.wasteCost
			if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-DELETE-"+item.id+"-W", item.material, now, "waste", item.waste, item.cost, total, "production_correction", item.id, "Compensating movement for deleted production job "+id, now); e != nil {
				return fail(e)
			}
		}
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM production_consumptions WHERE production_job_id=?`, id); e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM inventory_reservations WHERE production_job_id=?`, id); e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM production_jobs WHERE id=?`, id); e != nil {
		return fail(e)
	}
	if e = s.reverseOutsourceExpenseTx(ctx, tx, id); e != nil {
		return e
	}
	return s.reconcileOrderCOGSTx(ctx, tx, orderID)
}

func (s *Store) RecordProductionConsumption(ctx context.Context, jobID, materialID, key string, consumed, waste domain.Quantity, note string) (domain.ProductionConsumption, error) {
	if key == "" || consumed < 0 || waste < 0 || waste > domain.Quantity(int64(^uint64(0)>>1))-consumed || consumed+waste <= 0 {
		return domain.ProductionConsumption{}, fmt.Errorf("consumption quantities and idempotency key are required")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return domain.ProductionConsumption{}, e
	}
	defer tx.Rollback()
	result, e := s.recordProductionConsumptionTx(ctx, tx, jobID, materialID, key, consumed, waste, note)
	if e != nil {
		return result, e
	}
	if e = s.reconcileJobCOGSTx(ctx, tx, jobID); e != nil {
		return domain.ProductionConsumption{}, e
	}
	if e = tx.Commit(); e != nil {
		return domain.ProductionConsumption{}, e
	}
	return result, nil
}
func (s *Store) recordProductionConsumptionTx(ctx context.Context, tx *sql.Tx, jobID, materialID, key string, consumed, waste domain.Quantity, note string) (domain.ProductionConsumption, error) {
	var e error
	fail := func(e error) (domain.ProductionConsumption, error) { return domain.ProductionConsumption{}, e }
	var existing domain.ProductionConsumption
	var existingCreated string
	e = tx.QueryRowContext(ctx, `SELECT id,production_job_id,material_id,idempotency_key,consumed_quantity_units,waste_quantity_units,unit_cost_rial,material_cost_rial,waste_cost_rial,notes,created_at FROM production_consumptions WHERE production_job_id=? AND idempotency_key=?`, jobID, key).Scan(&existing.ID, &existing.ProductionJobID, &existing.MaterialID, &existing.IdempotencyKey, &existing.ConsumedQuantity, &existing.WasteQuantity, &existing.UnitCostRial, &existing.MaterialCostRial, &existing.WasteCostRial, &existing.Notes, &existingCreated)
	if e == nil {
		existing.CreatedAt, _ = time.Parse(time.RFC3339Nano, existingCreated)
		return existing, nil
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return fail(e)
	}
	var status string
	if e = tx.QueryRowContext(ctx, `SELECT status FROM production_jobs WHERE id=?`, jobID).Scan(&status); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrProductionJobNotFound)
	}
	if e != nil {
		return fail(e)
	}
	if status != domain.ProductionInProgress && status != domain.ProductionPaused {
		return fail(fmt.Errorf("job must be in progress to record usage"))
	}
	var qty, outsource domain.Quantity
	if e = tx.QueryRowContext(ctx, `SELECT quantity_units,outsource_quantity_units FROM production_jobs WHERE id=?`, jobID).Scan(&qty, &outsource); e != nil {
		return fail(e)
	}
	if qty <= outsource {
		return fail(fmt.Errorf("this job is fully outsourced; reduce outsourcing before using material"))
	}
	var reservedForJob int64
	if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_units),0) FROM inventory_reservations WHERE production_job_id=? AND material_id=? AND status='active'`, jobID, materialID).Scan(&reservedForJob); e != nil {
		return fail(e)
	}
	state, e := inventoryStateTx(ctx, tx, materialID)
	if e != nil {
		return fail(e)
	}
	// Reservations keep stock unavailable to other workflows, but they do
	// not make consumption mandatory. This job may use its own reservation
	// first and then consume any stock that is still unreserved.
	availableToJob := reservedForJob + int64(state.AvailableStock)
	if int64(consumed+waste) > availableToJob {
		return fail(domain.ErrInsufficientStock)
	}
	unitCost := state.AverageUnitCostRial
	// Allocate from the exact inventory value. Multiplying a rounded unit rate
	// can overdraw value or leave stranded Rial when the last units are used.
	totalCost, e := scaleProductionQuantity(domain.Quantity(state.InventoryValueRial), consumed+waste, state.PhysicalStock)
	if e != nil {
		return fail(e)
	}
	materialCost, e := scaleProductionQuantity(totalCost, consumed, consumed+waste)
	if e != nil {
		return fail(e)
	}
	matCost, wasteCost := int64(materialCost), int64(totalCost-materialCost)
	id := fmt.Sprintf("PC-%d", time.Now().UnixNano())
	now := time.Now().UTC()
	if _, e = tx.ExecContext(ctx, `INSERT INTO production_consumptions(id,production_job_id,material_id,idempotency_key,consumed_quantity_units,waste_quantity_units,unit_cost_rial,material_cost_rial,waste_cost_rial,notes,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, id, jobID, materialID, key, consumed, waste, unitCost, matCost, wasteCost, note, now.Format(time.RFC3339Nano)); e != nil {
		return fail(e)
	}
	if consumed > 0 {
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-"+id+"-CONSUMED", materialID, now.Format(time.RFC3339Nano), "production_consumption", -consumed, unitCost, -matCost, "production_consumption", id, "Production consumption", now.Format(time.RFC3339Nano)); e != nil {
			return fail(e)
		}
	}
	if waste > 0 {
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-"+id+"-WASTE", materialID, now.Format(time.RFC3339Nano), "waste", -waste, unitCost, -wasteCost, "production_consumption", id, "Production waste", now.Format(time.RFC3339Nano)); e != nil {
			return fail(e)
		}
	}
	remaining := int64(consumed + waste)
	resRows, e := tx.QueryContext(ctx, `SELECT id,quantity_units FROM inventory_reservations WHERE production_job_id=? AND material_id=? AND status='active' ORDER BY created_at,id`, jobID, materialID)
	if e != nil {
		return fail(e)
	}
	type rr struct {
		id string
		q  int64
	}
	var reservations []rr
	for resRows.Next() {
		var r rr
		if e = resRows.Scan(&r.id, &r.q); e != nil {
			resRows.Close()
			return fail(e)
		}
		reservations = append(reservations, r)
	}
	resRows.Close()
	for _, r := range reservations {
		if remaining == 0 {
			break
		}
		take := r.q
		if take > remaining {
			take = remaining
		}
		remaining -= take
		if take == r.q {
			if _, e = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='consumed',updated_at=? WHERE id=?`, now.Format(time.RFC3339Nano), r.id); e != nil {
				return fail(e)
			}
		} else if _, e = tx.ExecContext(ctx, `UPDATE inventory_reservations SET quantity_units=quantity_units-?,updated_at=? WHERE id=?`, take, now.Format(time.RFC3339Nano), r.id); e != nil {
			return fail(e)
		}
	}
	return domain.ProductionConsumption{ID: id, ProductionJobID: jobID, MaterialID: materialID, IdempotencyKey: key, ConsumedQuantity: consumed, WasteQuantity: waste, UnitCostRial: unitCost, MaterialCostRial: matCost, WasteCostRial: wasteCost, Notes: note, CreatedAt: now}, nil
}

func (s *Store) ReverseProductionConsumption(ctx context.Context, id, reason string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = s.reverseProductionConsumptionTx(ctx, tx, id, reason); e != nil {
		return e
	}
	var jobID string
	if e = tx.QueryRowContext(ctx, `SELECT production_job_id FROM production_consumptions WHERE id=?`, id).Scan(&jobID); e != nil {
		return e
	}
	if e = s.reconcileJobCOGSTx(ctx, tx, jobID); e != nil {
		return e
	}
	return tx.Commit()
}
func (s *Store) reverseProductionConsumptionTx(ctx context.Context, tx *sql.Tx, id, reason string) error {
	var e error
	fail := func(e error) error { return e }
	var job, material string
	var consumed, waste, cost, materialCost, wasteCost int64
	if e = tx.QueryRowContext(ctx, `SELECT production_job_id,material_id,consumed_quantity_units,waste_quantity_units,unit_cost_rial,material_cost_rial,waste_cost_rial FROM production_consumptions WHERE id=?`, id).Scan(&job, &material, &consumed, &waste, &cost, &materialCost, &wasteCost); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrConsumptionNotFound)
	}
	if e != nil {
		return fail(e)
	}
	var n int
	if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_movements WHERE reference_type='production_correction' AND reference_id=?`, id).Scan(&n); e != nil {
		return fail(e)
	}
	if n > 0 {
		return fail(fmt.Errorf("consumption is already corrected"))
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if consumed > 0 {
		total := materialCost
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-CORRECT-"+id+"-C", material, now, "production_consumption", consumed, cost, total, "production_correction", id, reason, now); e != nil {
			return fail(e)
		}
	}
	if waste > 0 {
		total := wasteCost
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-CORRECT-"+id+"-W", material, now, "waste", waste, cost, total, "production_correction", id, reason, now); e != nil {
			return fail(e)
		}
	}
	_ = job
	return nil
}

func (s *Store) ListProductionConsumptions(ctx context.Context, jobID string) ([]domain.ProductionConsumption, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,production_job_id,material_id,idempotency_key,consumed_quantity_units,waste_quantity_units,unit_cost_rial,material_cost_rial,waste_cost_rial,notes,created_at,EXISTS(SELECT 1 FROM inventory_movements m WHERE m.reference_id=production_consumptions.id AND m.reference_type='production_correction') FROM production_consumptions WHERE production_job_id=? ORDER BY created_at,id`, jobID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.ProductionConsumption{}
	for rows.Next() {
		var v domain.ProductionConsumption
		var at string
		if e = rows.Scan(&v.ID, &v.ProductionJobID, &v.MaterialID, &v.IdempotencyKey, &v.ConsumedQuantity, &v.WasteQuantity, &v.UnitCostRial, &v.MaterialCostRial, &v.WasteCostRial, &v.Notes, &at, &v.Reversed); e != nil {
			return nil, e
		}
		v.CreatedAt, e = time.Parse(time.RFC3339Nano, at)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func scanReservation(row scanner) (domain.InventoryReservation, error) {
	var v domain.InventoryReservation
	var q int64
	var created, updated string
	if e := row.Scan(&v.ID, &v.MaterialID, &v.OrderID, &v.OrderItemID, &v.ProductionJobID, &q, &v.Status, &created, &updated); e != nil {
		return v, e
	}
	v.Quantity = domain.Quantity(q)
	var e error
	v.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return v, e
	}
	v.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	return v, e
}
func scanProductionJob(row scanner) (domain.ProductionJob, error) {
	var v domain.ProductionJob
	var q int64
	var planned, started, completed, outsourceSent, outsourceExpected, outsourceReceived sql.NullString
	var outsourceQuoted int64
	var created, updated string
	if e := row.Scan(&v.ID, &v.JobNumber, &v.OrderID, &v.OrderItemID, &v.ServiceNameSnapshot, &q, &v.QuantityUnit, &v.AssignedMachineID, &v.Status, &v.Priority, &planned, &started, &completed, &v.Notes, &v.EstimatedCostRial, &v.ActualMaterialCostRial, &v.ActualWasteCostRial, &v.ActualOutsourcedCostRial, &v.OutsourceSupplierID, &v.OutsourceDescription, &outsourceQuoted, &outsourceSent, &outsourceExpected, &outsourceReceived, &v.OutsourceNotes, &created, &updated); e != nil {
		return v, e
	}
	v.Quantity = domain.Quantity(q)
	v.PlannedAt = parseOptionalTime(planned.String)
	v.StartedAt = parseOptionalTime(started.String)
	v.CompletedAt = parseOptionalTime(completed.String)
	v.OutsourceSentAt = outsourceSent.String
	v.OutsourceQuotedCostRial = outsourceQuoted
	v.OutsourceExpectedReturnAt = outsourceExpected.String
	v.OutsourceReceivedAt = outsourceReceived.String
	var e error
	v.CreatedAt, e = time.Parse(time.RFC3339Nano, created)
	if e != nil {
		return v, e
	}
	v.UpdatedAt, e = time.Parse(time.RFC3339Nano, updated)
	return v, e
}
func parseOptionalTime(v string) *time.Time {
	if v == "" {
		return nil
	}
	t, e := time.Parse(time.RFC3339Nano, v)
	if e != nil {
		return nil
	}
	return &t
}
