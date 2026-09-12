package sqlite

import (
	"Atropaten/internal/domain"
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) syncOrderItemsTx(ctx context.Context, tx *sql.Tx, order domain.Order) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,quantity_units,quantity_unit,cost_breakdown_json,resolved_parameters_json,service_id FROM order_items WHERE order_id=?`, order.ID)
	if err != nil {
		return err
	}
	type previous struct {
		quantity                        domain.Quantity
		unit, cost, parameters, service string
	}
	old := map[string]previous{}
	for rows.Next() {
		var id string
		var p previous
		if err = rows.Scan(&id, &p.quantity, &p.unit, &p.cost, &p.parameters, &p.service); err != nil {
			rows.Close()
			return err
		}
		old[id] = p
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, i := range order.Items {
		keep[i.ID] = true
	}
	for id := range old {
		if keep[id] {
			continue
		}
		// Document lines remain protected by their foreign keys. Production removal
		// uses the same compensation path as removing a job from the production view.
		jobs, err := tx.QueryContext(ctx, `SELECT id FROM production_jobs WHERE order_item_id=?`, id)
		if err != nil {
			return err
		}
		ids := []string{}
		for jobs.Next() {
			var job string
			if err = jobs.Scan(&job); err != nil {
				jobs.Close()
				return err
			}
			ids = append(ids, job)
		}
		err = jobs.Err()
		jobs.Close()
		if err != nil {
			return err
		}
		for _, job := range ids {
			if err = s.deleteProductionJobTx(ctx, tx, job); err != nil {
				return err
			}
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM inventory_reservations WHERE order_item_id=?`, id); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM order_items WHERE id=? AND order_id=?`, id, order.ID); err != nil {
			return fmt.Errorf("order item is referenced by an invoice: %w", err)
		}
	}
	// Move existing positions above their current maximum before upserting to
	// support reorder without deleting stable IDs used by jobs and invoices.
	var offset int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(display_order),0)+?+1 FROM order_items WHERE order_id=?`, len(order.Items), order.ID).Scan(&offset); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE order_items SET display_order=display_order+? WHERE order_id=?`, offset, order.ID); err != nil {
		return err
	}
	for _, i := range order.Items {
		_, err = tx.ExecContext(ctx, `INSERT INTO order_items(id,order_id,display_order,service_id,service_name_snapshot,service_code_snapshot,quantity_units,quantity_unit,resolved_parameters_json,cost_breakdown_json,pricing_snapshot_json,estimated_cost_rial,suggested_price_rial,selling_price_rial,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET display_order=excluded.display_order,service_id=excluded.service_id,service_name_snapshot=excluded.service_name_snapshot,service_code_snapshot=excluded.service_code_snapshot,quantity_units=excluded.quantity_units,quantity_unit=excluded.quantity_unit,resolved_parameters_json=excluded.resolved_parameters_json,cost_breakdown_json=excluded.cost_breakdown_json,pricing_snapshot_json=excluded.pricing_snapshot_json,estimated_cost_rial=excluded.estimated_cost_rial,suggested_price_rial=excluded.suggested_price_rial,selling_price_rial=excluded.selling_price_rial,notes=excluded.notes WHERE order_items.order_id=excluded.order_id`, i.ID, order.ID, i.Position, i.ServiceID, i.ServiceNameSnapshot, i.ServiceCodeSnapshot, i.Quantity, i.QuantityUnit, i.ResolvedParametersJSON, i.CostBreakdownJSON, i.PricingSnapshotJSON, i.EstimatedCostRial, i.SuggestedPriceRial, i.SellingPriceRial, i.Notes)
		if err != nil {
			return err
		}
		p, exists := old[i.ID]
		if !exists || (p.quantity == i.Quantity && p.unit == i.QuantityUnit && p.cost == i.CostBreakdownJSON && p.parameters == i.ResolvedParametersJSON && p.service == i.ServiceID) {
			continue
		}
		jobs, err := tx.QueryContext(ctx, `SELECT id FROM production_jobs WHERE order_item_id=? AND status NOT IN ('Completed','Cancelled')`, i.ID)
		if err != nil {
			return err
		}
		ids := []string{}
		for jobs.Next() {
			var id string
			if err = jobs.Scan(&id); err != nil {
				jobs.Close()
				return err
			}
			ids = append(ids, id)
		}
		err = jobs.Err()
		jobs.Close()
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err = syncProductionMaterialsTx(ctx, tx, id); err != nil {
				return err
			}
		}
	}
	return nil
}
