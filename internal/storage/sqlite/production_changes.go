package sqlite

import (
	"Atropaten/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *Store) SetProductionMaterialTarget(ctx context.Context, jobID, materialID string, target domain.Quantity) error {
	if target < 0 {
		return fmt.Errorf("planned quantity cannot be negative")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var qty, outsource, required domain.Quantity
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT j.quantity_units,j.outsource_quantity_units,p.required_units,j.status FROM production_jobs j JOIN production_material_plans p ON p.production_job_id=j.id WHERE j.id=? AND p.material_id=?`, jobID, materialID).Scan(&qty, &outsource, &required, &status); err != nil {
		return err
	}
	if status == "Completed" || status == "Cancelled" || qty <= outsource {
		return domain.ErrProductionNotEditable
	}
	fullTarget, err := scaleProductionQuantity(target, qty, qty-outsource)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE production_material_plans SET adjustment_units=? WHERE production_job_id=? AND material_id=?`, fullTarget-required, jobID, materialID); err != nil {
		return err
	}
	if err = syncProductionMaterialsTx(ctx, tx, jobID); err != nil {
		return err
	}
	return tx.Commit()
}

// Keep edits to a managed reservation across reloads. Adjustment is against
// the order requirement, so later order changes still increase/decrease it.
func adjustManagedReservationTx(ctx context.Context, tx *sql.Tx, id string, remaining domain.Quantity) error {
	var jobID, materialID string
	var required, qty, outsource, old domain.Quantity
	err := tx.QueryRowContext(ctx, `SELECT p.production_job_id,p.material_id,p.required_units,j.quantity_units,j.outsource_quantity_units,r.quantity_units FROM production_material_plans p JOIN production_jobs j ON j.id=p.production_job_id JOIN inventory_reservations r ON r.id=p.reservation_id WHERE p.reservation_id=?`, id).Scan(&jobID, &materialID, &required, &qty, &outsource, &old)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if qty <= outsource {
		return domain.ErrProductionNotEditable
	}
	var used, other domain.Quantity
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(-m.quantity_delta_units),0) FROM inventory_movements m JOIN production_consumptions c ON c.id=m.reference_id WHERE c.production_job_id=? AND c.material_id=? AND m.reference_type IN ('production_consumption','production_correction')`, jobID, materialID).Scan(&used); err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_units),0) FROM inventory_reservations WHERE production_job_id=? AND material_id=? AND status='active' AND id<>?`, jobID, materialID, id).Scan(&other); err != nil {
		return err
	}
	target, err := scaleProductionQuantity(used+other+remaining, qty, qty-outsource)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE production_material_plans SET adjustment_units=? WHERE production_job_id=? AND material_id=?`, target-required, jobID, materialID)
	return err
}

func (s *Store) UpdateProductionConsumption(ctx context.Context, id, key string, consumed, waste domain.Quantity, note string) error {
	if key == "" || consumed < 0 || waste < 0 || waste > domain.Quantity(int64(^uint64(0)>>1))-consumed {
		return fmt.Errorf("invalid consumption quantities or request key")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var jobID, materialID, status string
	if err = tx.QueryRowContext(ctx, `SELECT c.production_job_id,c.material_id,j.status FROM production_consumptions c JOIN production_jobs j ON j.id=c.production_job_id WHERE c.id=?`, id).Scan(&jobID, &materialID, &status); err != nil {
		return err
	}
	if status == "Cancelled" || status == "Completed" {
		return domain.ErrProductionNotEditable
	}
	var exists int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM production_consumptions WHERE production_job_id=? AND idempotency_key=?`, jobID, key).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	if err = s.reverseProductionConsumptionTx(ctx, tx, id, "Operator correction"); err != nil {
		return err
	}
	if consumed+waste > 0 {
		if _, err = s.recordProductionConsumptionTx(ctx, tx, jobID, materialID, key, consumed, waste, note); err != nil {
			return err
		}
	}
	if err = syncProductionMaterialsTx(ctx, tx, jobID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) reverseOutsourceExpenseTx(ctx context.Context, tx *sql.Tx, jobID string) error {
	id := "EXP-OUTSOURCE-" + jobID
	var status, journalID string
	err := tx.QueryRowContext(ctx, `SELECT status,journal_entry_id FROM expenses WHERE id=?`, id).Scan(&status, &journalID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if status != "Posted" {
		return nil
	}
	now := time.Now().UTC()
	if _, err = s.reverseJournalTx(ctx, tx, journalID, "outsource:reverse:"+journalID, "Reverse outsourcing "+jobID, now); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE expenses SET status='Reversed',updated_at=? WHERE id=?`, now.Format(time.RFC3339Nano), id)
	return err
}
