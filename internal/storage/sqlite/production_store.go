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
	_, e = tx.ExecContext(ctx, `UPDATE inventory_reservations SET quantity_units=?,updated_at=? WHERE id=?`, quantity, time.Now().UTC().Format(time.RFC3339Nano), id)
	if e != nil {
		return fail(e)
	}
	return tx.Commit()
}

func (s *Store) ReleaseReservation(ctx context.Context, id, status string) error {
	if status != "released" && status != "cancelled" {
		status = domain.ReservationReleased
	}
	res, e := s.db.ExecContext(ctx, `UPDATE inventory_reservations SET status=?,updated_at=? WHERE id=? AND status='active'`, status, time.Now().UTC().Format(time.RFC3339Nano), id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrReservationNotFound
	}
	return nil
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
	if e := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN m.movement_type='production_consumption' THEN -m.total_cost_rial ELSE 0 END),0),COALESCE(SUM(CASE WHEN m.movement_type='waste' THEN -m.total_cost_rial ELSE 0 END),0) FROM inventory_movements m JOIN production_consumptions c ON c.id=m.reference_id WHERE c.production_job_id=? AND m.reference_type='production_consumption'`, j.ID).Scan(&j.ActualMaterialCostRial, &j.ActualWasteCostRial); e != nil {
		return j, e
	}
	return j, nil
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
	if e = tx.QueryRowContext(ctx, `SELECT service_name_snapshot,quantity_units,quantity_unit,estimated_cost_rial FROM order_items WHERE id=? AND order_id=?`, j.OrderItemID, j.OrderID).Scan(&service, &qty, &unit, &estimated); e != nil {
		return fail(fmt.Errorf("order item: %w", e))
	}
	j.ServiceNameSnapshot = service
	if j.Quantity <= 0 {
		j.Quantity = domain.Quantity(qty)
	}
	if j.QuantityUnit == "" {
		j.QuantityUnit = unit
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
	now := time.Now().UTC()
	var started, completed any
	if status == domain.ProductionInProgress {
		started = now.Format(time.RFC3339Nano)
	}
	if status == domain.ProductionCompleted {
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
	if _, e = tx.ExecContext(ctx, `UPDATE production_jobs SET status=?,started_at=COALESCE(?,started_at),completed_at=COALESCE(?,completed_at),updated_at=? WHERE id=?`, status, started, completed, now.Format(time.RFC3339Nano), id); e != nil {
		return fail(e)
	}
	if status == domain.ProductionCompleted {
		var open int
		if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM production_jobs WHERE order_id=? AND status NOT IN ('Completed','Cancelled')`, orderID).Scan(&open); e != nil {
			return fail(e)
		}
		if open == 0 {
			if _, e = tx.ExecContext(ctx, `UPDATE orders SET fulfillment_status=?,updated_at=? WHERE id=? AND fulfillment_status IN (?,?)`, domain.FulfillmentReady, now.Format(time.RFC3339Nano), orderID, domain.FulfillmentPending, domain.FulfillmentInProduction); e != nil {
				return fail(e)
			}
		}
	}
	return tx.Commit()
}

func (s *Store) DeleteProductionJob(ctx context.Context, id string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var status string
	if e = tx.QueryRowContext(ctx, `SELECT status FROM production_jobs WHERE id=?`, id).Scan(&status); errors.Is(e, sql.ErrNoRows) {
		return fail(domain.ErrProductionJobNotFound)
	}
	if e != nil {
		return fail(e)
	}
	if status != domain.ProductionPending && status != domain.ProductionReady {
		return fail(domain.ErrProductionHistoryProtected)
	}
	var n int
	if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM production_consumptions WHERE production_job_id=?`, id).Scan(&n); e != nil {
		return fail(e)
	}
	if n > 0 {
		return fail(domain.ErrProductionHistoryProtected)
	}
	if _, e = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='cancelled',updated_at=? WHERE production_job_id=? AND status='active'`, time.Now().UTC().Format(time.RFC3339Nano), id); e != nil {
		return fail(e)
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM production_jobs WHERE id=?`, id); e != nil {
		return fail(e)
	}
	return tx.Commit()
}

func (s *Store) RecordProductionConsumption(ctx context.Context, jobID, materialID, key string, consumed, waste domain.Quantity, note string) (domain.ProductionConsumption, error) {
	if key == "" || consumed < 0 || waste < 0 || waste > domain.Quantity(int64(^uint64(0)>>1))-consumed || consumed+waste <= 0 {
		return domain.ProductionConsumption{}, fmt.Errorf("consumption quantities and idempotency key are required")
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return domain.ProductionConsumption{}, e
	}
	fail := func(x error) (domain.ProductionConsumption, error) {
		_ = tx.Rollback()
		return domain.ProductionConsumption{}, x
	}
	var existing domain.ProductionConsumption
	var existingCreated string
	e = tx.QueryRowContext(ctx, `SELECT id,production_job_id,material_id,idempotency_key,consumed_quantity_units,waste_quantity_units,unit_cost_rial,material_cost_rial,waste_cost_rial,notes,created_at FROM production_consumptions WHERE production_job_id=? AND idempotency_key=?`, jobID, key).Scan(&existing.ID, &existing.ProductionJobID, &existing.MaterialID, &existing.IdempotencyKey, &existing.ConsumedQuantity, &existing.WasteQuantity, &existing.UnitCostRial, &existing.MaterialCostRial, &existing.WasteCostRial, &existing.Notes, &existingCreated)
	if e == nil {
		existing.CreatedAt, _ = time.Parse(time.RFC3339Nano, existingCreated)
		_ = tx.Rollback()
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
	var reserved int64
	if e = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_units),0) FROM inventory_reservations WHERE production_job_id=? AND material_id=? AND status='active'`, jobID, materialID).Scan(&reserved); e != nil {
		return fail(e)
	}
	if int64(consumed+waste) > reserved {
		return fail(domain.ErrReservationExceeded)
	}
	state, e := inventoryStateTx(ctx, tx, materialID)
	if e != nil {
		return fail(e)
	}
	if consumed+waste > state.PhysicalStock {
		return fail(domain.ErrInsufficientStock)
	}
	unitCost := state.AverageUnitCostRial
	matCost, e := domain.MulQuantityRial(consumed, unitCost)
	if e != nil {
		return fail(e)
	}
	wasteCost, e := domain.MulQuantityRial(waste, unitCost)
	if e != nil {
		return fail(e)
	}
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
	if e = tx.Commit(); e != nil {
		return domain.ProductionConsumption{}, e
	}
	return domain.ProductionConsumption{ID: id, ProductionJobID: jobID, MaterialID: materialID, IdempotencyKey: key, ConsumedQuantity: consumed, WasteQuantity: waste, UnitCostRial: unitCost, MaterialCostRial: matCost, WasteCostRial: wasteCost, Notes: note, CreatedAt: now}, nil
}

func (s *Store) ReverseProductionConsumption(ctx context.Context, id, reason string) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	fail := func(x error) error { _ = tx.Rollback(); return x }
	var job, material string
	var consumed, waste, cost int64
	if e = tx.QueryRowContext(ctx, `SELECT production_job_id,material_id,consumed_quantity_units,waste_quantity_units,unit_cost_rial FROM production_consumptions WHERE id=?`, id).Scan(&job, &material, &consumed, &waste, &cost); errors.Is(e, sql.ErrNoRows) {
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
		total, _ := domain.MulQuantityRial(domain.Quantity(consumed), cost)
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-CORRECT-"+id+"-C", material, now, "production_consumption", consumed, cost, total, "production_correction", id, reason, now); e != nil {
			return fail(e)
		}
	}
	if waste > 0 {
		total, _ := domain.MulQuantityRial(domain.Quantity(waste), cost)
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-CORRECT-"+id+"-W", material, now, "waste", waste, cost, total, "production_correction", id, reason, now); e != nil {
			return fail(e)
		}
	}
	_ = job
	return tx.Commit()
}

func (s *Store) ListProductionConsumptions(ctx context.Context, jobID string) ([]domain.ProductionConsumption, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,production_job_id,material_id,idempotency_key,consumed_quantity_units,waste_quantity_units,unit_cost_rial,material_cost_rial,waste_cost_rial,notes,created_at FROM production_consumptions WHERE production_job_id=? ORDER BY created_at,id`, jobID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.ProductionConsumption{}
	for rows.Next() {
		var v domain.ProductionConsumption
		var at string
		if e = rows.Scan(&v.ID, &v.ProductionJobID, &v.MaterialID, &v.IdempotencyKey, &v.ConsumedQuantity, &v.WasteQuantity, &v.UnitCostRial, &v.MaterialCostRial, &v.WasteCostRial, &v.Notes, &at); e != nil {
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
