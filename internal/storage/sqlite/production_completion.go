package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"Atropaten/internal/domain"
)

// Remaining material includes the plan (less usage already recorded) and any
// additional allocations. The same quantities drive forecasting and completion.
func remainingProductionMaterialsTx(ctx context.Context, tx *sql.Tx, jobID string) (map[string]domain.Quantity, error) {
	rows, err := tx.QueryContext(ctx, `SELECT p.material_id,p.required_units,p.adjustment_units,j.quantity_units,j.outsource_quantity_units,
	 COALESCE((SELECT SUM(-m.quantity_delta_units) FROM inventory_movements m JOIN production_consumptions c ON c.id=m.reference_id WHERE c.production_job_id=j.id AND c.material_id=p.material_id AND m.reference_type IN ('production_consumption','production_correction')),0)
	 FROM production_material_plans p JOIN production_jobs j ON j.id=p.production_job_id WHERE j.id=?`, jobID)
	if err != nil {
		return nil, err
	}
	remaining := map[string]domain.Quantity{}
	for rows.Next() {
		var id string
		var required, adjustment, qty, outsource, used domain.Quantity
		if err = rows.Scan(&id, &required, &adjustment, &qty, &outsource, &used); err != nil {
			rows.Close()
			return nil, err
		}
		target := required + adjustment
		if target < 0 || required == 0 {
			target = 0
		}
		target, err = scaleProductionQuantity(target, qty-outsource, qty)
		if err != nil {
			rows.Close()
			return nil, err
		}
		if target > used {
			remaining[id] = target - used
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = tx.QueryContext(ctx, `SELECT material_id,SUM(quantity_units) FROM inventory_reservations WHERE production_job_id=? AND status='active' GROUP BY material_id`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var reserved domain.Quantity
		if err = rows.Scan(&id, &reserved); err != nil {
			return nil, err
		}
		if reserved > remaining[id] {
			remaining[id] = reserved
		}
	}
	return remaining, rows.Err()
}

func (s *Store) completeProductionMaterialsTx(ctx context.Context, tx *sql.Tx, jobID string) error {
	if err := syncProductionMaterialsTx(ctx, tx, jobID); err != nil {
		return err
	}
	remaining, err := remainingProductionMaterialsTx(ctx, tx, jobID)
	if err != nil {
		return err
	}
	for materialID, quantity := range remaining {
		if quantity <= 0 {
			continue
		}
		key := fmt.Sprintf("complete:%s:%s:%d", jobID, materialID, time.Now().UnixNano())
		if _, err = s.recordProductionConsumptionTx(ctx, tx, jobID, materialID, key, quantity, 0, "Planned material recorded on completion"); err != nil {
			return fmt.Errorf("cannot complete job: material %s: %w", materialID, err)
		}
	}
	return nil
}
