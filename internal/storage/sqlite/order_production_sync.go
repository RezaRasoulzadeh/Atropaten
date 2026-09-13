package sqlite

import (
	"Atropaten/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Job creation puts an order into production, including Pending/Ready jobs.
// Readiness requires every item to be completed and no unfinished live jobs.
// Delivery remains an operator-controlled fact independent of production work.
func syncProductionFulfillmentTx(ctx context.Context, tx *sql.Tx, orderID string, now time.Time) error {
	var items, incomplete, live, unfinished int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(CASE WHEN NOT EXISTS(SELECT 1 FROM production_jobs p WHERE p.order_item_id=i.id AND p.status='Completed') THEN 1 ELSE 0 END),0) FROM order_items i WHERE i.order_id=? AND i.removed_at IS NULL`, orderID).Scan(&items, &incomplete); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(CASE WHEN status<>'Completed' THEN 1 ELSE 0 END),0) FROM production_jobs WHERE order_id=? AND status<>'Cancelled' AND order_item_id IN (SELECT id FROM order_items WHERE removed_at IS NULL)`, orderID).Scan(&live, &unfinished); err != nil {
		return err
	}
	status := domain.FulfillmentPending
	if items > 0 && incomplete == 0 && unfinished == 0 {
		status = domain.FulfillmentReady
	} else if live > 0 {
		status = domain.FulfillmentInProduction
	}
	_, err := tx.ExecContext(ctx, `UPDATE orders SET fulfillment_status=?,updated_at=? WHERE id=? AND fulfillment_status IN ('Pending','In Production','Ready') AND fulfillment_status<>?`, status, now.UTC().Format(time.RFC3339Nano), orderID, status)
	return err
}

func (s *Store) syncOrderItemsTx(ctx context.Context, tx *sql.Tx, order domain.Order) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,quantity_units,quantity_unit,cost_breakdown_json,resolved_parameters_json,service_id FROM order_items WHERE order_id=? AND removed_at IS NULL`, order.ID)
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
		// Keep the item and job identities for immutable invoice and consumption
		// history. Only current commitments are cancelled on removal.
		now := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err = tx.ExecContext(ctx, `UPDATE order_items SET removed_at=? WHERE id=? AND order_id=?`, now, id, order.ID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE production_jobs SET status='Cancelled',updated_at=? WHERE order_item_id=? AND status NOT IN ('Completed','Cancelled')`, now, id); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='cancelled',updated_at=? WHERE order_item_id=? AND status='active'`, now, id); err != nil {
			return err
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
		p, exists := old[i.ID]
		productionChanged := exists && (p.quantity != i.Quantity || p.unit != i.QuantityUnit || p.cost != i.CostBreakdownJSON || p.parameters != i.ResolvedParametersJSON || p.service != i.ServiceID)

		_, err = tx.ExecContext(ctx, `INSERT INTO order_items(id,order_id,display_order,service_id,service_name_snapshot,service_code_snapshot,quantity_units,quantity_unit,resolved_parameters_json,cost_breakdown_json,pricing_snapshot_json,estimated_cost_rial,suggested_price_rial,selling_price_rial,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET display_order=excluded.display_order,service_id=excluded.service_id,service_name_snapshot=excluded.service_name_snapshot,service_code_snapshot=excluded.service_code_snapshot,quantity_units=excluded.quantity_units,quantity_unit=excluded.quantity_unit,resolved_parameters_json=excluded.resolved_parameters_json,cost_breakdown_json=excluded.cost_breakdown_json,pricing_snapshot_json=excluded.pricing_snapshot_json,estimated_cost_rial=excluded.estimated_cost_rial,suggested_price_rial=excluded.suggested_price_rial,selling_price_rial=excluded.selling_price_rial,notes=excluded.notes,removed_at=NULL WHERE order_items.order_id=excluded.order_id`, i.ID, order.ID, i.Position, i.ServiceID, i.ServiceNameSnapshot, i.ServiceCodeSnapshot, i.Quantity, i.QuantityUnit, i.ResolvedParametersJSON, i.CostBreakdownJSON, i.PricingSnapshotJSON, i.EstimatedCostRial, i.SuggestedPriceRial, i.SellingPriceRial, i.Notes)
		if err != nil {
			return err
		}
		if !exists {
			var scheduled bool
			if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM production_jobs WHERE order_id=?)`, order.ID).Scan(&scheduled); err != nil {
				return err
			}
			if scheduled {
				if err = createOrderProductionTx(ctx, tx, order, i); err != nil {
					return err
				}
			}
		}
		if productionChanged {
			jobs, err := orderDependencyIDs(ctx, tx, `SELECT id FROM production_jobs WHERE order_item_id=? AND status<>'Cancelled'`, i.ID)
			if err != nil {
				return err
			}
			for _, id := range jobs {
				if err = reconcileEditedProductionTx(ctx, tx, id, p.service != i.ServiceID || p.unit != i.QuantityUnit || p.parameters != i.ResolvedParametersJSON); err != nil {
					return err
				}
			}
		}

	}
	// Historical rows occupy deterministic positions after current lines, so
	// retries do not keep shifting their positions or collide with added lines.
	removed, err := orderDependencyIDs(ctx, tx, `SELECT id FROM order_items WHERE order_id=? AND removed_at IS NOT NULL ORDER BY removed_at,id`, order.ID)
	if err != nil {
		return err
	}
	for n, id := range removed {
		if _, err = tx.ExecContext(ctx, `UPDATE order_items SET display_order=? WHERE id=?`, len(order.Items)+n, id); err != nil {
			return err
		}
	}
	return syncProductionFulfillmentTx(ctx, tx, order.ID, time.Now().UTC())
}

func createOrderProductionTx(ctx context.Context, tx *sql.Tx, o domain.Order, item domain.OrderItem) error {
	var number int64
	if err := tx.QueryRowContext(ctx, `SELECT next_number FROM production_number_sequences WHERE id=1`).Scan(&number); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE production_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
		return err
	}
	id := "JOB-ORDER-" + item.ID
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := tx.ExecContext(ctx, `INSERT INTO production_jobs(id,job_number,order_id,order_item_id,service_name_snapshot,quantity_units,quantity_unit,status,priority,estimated_cost_rial,created_at,updated_at) VALUES(?,?,?,?,?,?,?,'Pending',?,?,?,?)`, id, fmt.Sprintf("JOB-%04d", number), o.ID, item.ID, item.ServiceNameSnapshot, item.Quantity, item.QuantityUnit, o.Priority, item.EstimatedCostRial, now, now)
	if err != nil {
		return err
	}
	return reconcileProductionMaterialsTx(ctx, tx, id, false, true)
}

func reconcileEditedProductionTx(ctx context.Context, tx *sql.Tx, jobID string, configurationChanged bool) error {
	var status string
	var produced domain.Quantity
	if err := tx.QueryRowContext(ctx, `SELECT status,produced_quantity_units FROM production_jobs WHERE id=?`, jobID).Scan(&status, &produced); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='released',updated_at=? WHERE production_job_id=? AND status='active'`, now, jobID); err != nil {
		return err
	}
	if err := reconcileProductionMaterialsTx(ctx, tx, jobID, false, true); err != nil {
		return err
	}
	if status == domain.ProductionCompleted {
		var quantity domain.Quantity
		if err := tx.QueryRowContext(ctx, `SELECT quantity_units FROM production_jobs WHERE id=?`, jobID).Scan(&quantity); err != nil {
			return err
		}
		remaining, err := remainingProductionMaterialsTx(ctx, tx, jobID)
		if err != nil {
			return err
		}
		if quantity > produced || configurationChanged || len(remaining) > 0 {
			_, err = tx.ExecContext(ctx, `UPDATE production_jobs SET status='Pending',completed_at=NULL,updated_at=? WHERE id=?`, now, jobID)
			return err
		}
	}
	return nil
}

// Commercial cancellation/drafting releases commitments; restoring a commercial
// order restores only the unconsumed plan, without fabricating production usage.
func syncOrderCommercialProductionTx(ctx context.Context, tx *sql.Tx, orderID string) error {
	jobs, err := orderDependencyIDs(ctx, tx, `SELECT p.id FROM production_jobs p JOIN order_items i ON i.id=p.order_item_id WHERE p.order_id=? AND p.status NOT IN ('Completed','Cancelled') AND i.removed_at IS NULL`, orderID)
	if err != nil {
		return err
	}
	for _, id := range jobs {
		if err = reconcileProductionMaterialsTx(ctx, tx, id, false, true); err != nil {
			return err
		}
	}
	return nil
}
