package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type ProductionMaterial struct {
	MaterialID    string `json:"materialId"`
	Name          string `json:"name"`
	Unit          string `json:"unit"`
	ReservationID string `json:"reservationId"`
	Required      string `json:"required"`
	Planned       string `json:"planned"`
	Reserved      string `json:"reserved"`
	Used          string `json:"used"`
	Available     string `json:"available"`
	Shortage      string `json:"shortage"`
}

// Order quantities and resolved selections are authoritative. Service definitions
// are used only to recover identity from legacy snapshots lacking material IDs.
func orderMaterialRequirementsTx(ctx context.Context, tx *sql.Tx, itemID string) (map[string]domain.Quantity, error) {
	var componentsJSON, paramsJSON, serviceID string
	var quantity domain.Quantity
	if err := tx.QueryRowContext(ctx, `SELECT cost_breakdown_json,resolved_parameters_json,service_id,quantity_units FROM order_items WHERE id=?`, itemID).Scan(&componentsJSON, &paramsJSON, &serviceID, &quantity); err != nil {
		return nil, err
	}
	var components []struct {
		ID, Name, Type, MaterialID, ReferenceID, ParameterKey, UsageQuantity string
		Enabled                                                              bool
	}
	// Explicit tags are needed for the saved camelCase field names.
	var raw []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Type          string `json:"type"`
		MaterialID    string `json:"materialId"`
		ReferenceID   string `json:"referenceId"`
		ParameterKey  string `json:"parameterKey"`
		UsageQuantity string `json:"usageQuantity"`
		Enabled       bool   `json:"enabled"`
	}
	if strings.HasPrefix(strings.TrimSpace(componentsJSON), "[") {
		if err := json.Unmarshal([]byte(componentsJSON), &raw); err != nil {
			return nil, fmt.Errorf("read order materials: %w", err)
		}
	}
	for _, c := range raw {
		components = append(components, struct {
			ID, Name, Type, MaterialID, ReferenceID, ParameterKey, UsageQuantity string
			Enabled                                                              bool
		}{c.ID, c.Name, c.Type, c.MaterialID, c.ReferenceID, c.ParameterKey, c.UsageQuantity, c.Enabled})
	}
	var parameters []struct {
		Key        string `json:"key"`
		MaterialID string `json:"materialId"`
	}
	if strings.HasPrefix(strings.TrimSpace(paramsJSON), "[") {
		if err := json.Unmarshal([]byte(paramsJSON), &parameters); err != nil {
			return nil, err
		}
	}
	selected := map[string]string{}
	for _, p := range parameters {
		if p.MaterialID != "" {
			selected[p.Key] = p.MaterialID
		}
	}
	materialCount := 0
	for _, c := range components {
		if c.Enabled && c.Type == "material" {
			materialCount++
		}
	}
	required := map[string]domain.Quantity{}
	for _, c := range components {
		if !c.Enabled || c.Type != "material" {
			continue
		}
		id, key := c.MaterialID, c.ParameterKey
		if id == "" {
			id = c.ReferenceID
		}
		if id == "" && key == "" {
			var reference, parameter string
			err := tx.QueryRowContext(ctx, `SELECT reference_id,parameter_key FROM service_cost_components WHERE id=? AND service_id=? AND component_type='material'`, c.ID, serviceID).Scan(&reference, &parameter)
			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}
			id, key = reference, parameter
		}
		if id == "" {
			id = selected[key]
		}
		if id == "" && materialCount == 1 && len(selected) == 1 {
			for _, value := range selected {
				id = value
			}
		}
		if id == "" {
			return nil, fmt.Errorf("material %q on this order needs a material selection", c.Name)
		}
		usage, err := domain.ParseQuantity(c.UsageQuantity)
		if err != nil || usage < 0 {
			return nil, fmt.Errorf("invalid usage for material %q", c.Name)
		}
		total, err := multiplyProductionQuantity(usage, quantity)
		if err != nil {
			return nil, err
		}
		sum := new(big.Int).Add(big.NewInt(int64(required[id])), big.NewInt(int64(total)))
		if !sum.IsInt64() {
			return nil, fmt.Errorf("material quantity is too large")
		}
		required[id] = domain.Quantity(sum.Int64())
	}
	return required, nil
}

func scaleProductionQuantity(q, numerator, denominator domain.Quantity) (domain.Quantity, error) {
	if denominator <= 0 {
		return 0, fmt.Errorf("production quantity must be positive")
	}
	n := new(big.Int).Mul(big.NewInt(int64(q)), big.NewInt(int64(numerator)))
	n.Add(n, big.NewInt(int64(denominator)/2))
	n.Quo(n, big.NewInt(int64(denominator)))
	if !n.IsInt64() {
		return 0, fmt.Errorf("material quantity is too large")
	}
	return domain.Quantity(n.Int64()), nil
}

func syncProductionMaterialsTx(ctx context.Context, tx *sql.Tx, jobID string) error {
	var itemID, status, commercial string
	var qty, outsource domain.Quantity
	if err := tx.QueryRowContext(ctx, `SELECT p.order_item_id,p.status,i.quantity_units,p.outsource_quantity_units,o.commercial_status FROM production_jobs p JOIN order_items i ON i.id=p.order_item_id JOIN orders o ON o.id=p.order_id WHERE p.id=?`, jobID).Scan(&itemID, &status, &qty, &outsource, &commercial); err != nil {
		return err
	}
	if status == "Cancelled" || status == "Completed" || commercial == "Cancelled" || commercial == "Closed" || commercial == "Draft" {
		return nil
	}
	if qty <= 0 || outsource > qty {
		return fmt.Errorf("order quantity cannot be below the outsourced quantity")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE production_jobs SET cost_breakdown_json=(SELECT cost_breakdown_json FROM order_items WHERE id=?) WHERE id=?`, itemID, jobID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE production_jobs SET quantity_units=?,quantity_unit=(SELECT quantity_unit FROM order_items WHERE id=?),service_name_snapshot=(SELECT service_name_snapshot FROM order_items WHERE id=?),estimated_cost_rial=(SELECT estimated_cost_rial FROM order_items WHERE id=?) WHERE id=?`, qty, itemID, itemID, itemID, jobID); err != nil {
		return err
	}
	required, err := orderMaterialRequirementsTx(ctx, tx, itemID)
	if err != nil {
		return err
	}
	// Retain plans whose material was removed so their unused reservation is released.
	rows, err := tx.QueryContext(ctx, `SELECT material_id FROM production_material_plans WHERE production_job_id=?`, jobID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		if _, ok := required[id]; !ok {
			required[id] = 0
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(required))
	for id := range required {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, id := range ids {
		reservationID := "RES-AUTO-" + jobID + "-" + id
		_, err = tx.ExecContext(ctx, `INSERT INTO production_material_plans(production_job_id,material_id,required_units,reservation_id) VALUES(?,?,?,?) ON CONFLICT(production_job_id,material_id) DO UPDATE SET required_units=excluded.required_units`, jobID, id, required[id], reservationID)
		if err != nil {
			return err
		}
		// Retire the old per-component auto reservations when upgrading a job.
		if _, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='released',updated_at=? WHERE production_job_id=? AND material_id=? AND status='active' AND substr(id,1,length(?))=?`, now, jobID, id, "RES-"+jobID+"-", "RES-"+jobID+"-"); err != nil {
			return err
		}
		var adjustment domain.Quantity
		if err = tx.QueryRowContext(ctx, `SELECT adjustment_units FROM production_material_plans WHERE production_job_id=? AND material_id=?`, jobID, id).Scan(&adjustment); err != nil {
			return err
		}
		target := required[id] + adjustment
		if target < 0 || required[id] == 0 {
			target = 0
		}
		target, err = scaleProductionQuantity(target, qty-outsource, qty)
		if err != nil {
			return err
		}
		var used, manual domain.Quantity
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(-m.quantity_delta_units),0) FROM inventory_movements m JOIN production_consumptions c ON c.id=m.reference_id WHERE c.production_job_id=? AND c.material_id=? AND m.reference_type IN ('production_consumption','production_correction')`, jobID, id).Scan(&used); err != nil {
			return err
		}
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_units),0) FROM inventory_reservations WHERE production_job_id=? AND material_id=? AND status='active' AND id<>?`, jobID, id, reservationID).Scan(&manual); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='released',updated_at=? WHERE id=? AND status='active'`, now, reservationID); err != nil {
			return err
		}
		wanted := target - used - manual
		if wanted < 0 {
			wanted = 0
		}
		state, err := inventoryStateTx(ctx, tx, id)
		if err != nil {
			return err
		}
		if wanted > state.AvailableStock {
			wanted = state.AvailableStock
		}
		if wanted > 0 {
			_, err = tx.ExecContext(ctx, `INSERT INTO inventory_reservations(id,material_id,order_id,order_item_id,production_job_id,quantity_units,status,created_at,updated_at) SELECT ?,?,order_id,order_item_id,id,?,'active',?,? FROM production_jobs WHERE id=? ON CONFLICT(id) DO UPDATE SET quantity_units=excluded.quantity_units,status='active',updated_at=excluded.updated_at`, reservationID, id, wanted, now, now, jobID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) ProductionMaterials(ctx context.Context, jobID string) ([]ProductionMaterial, error) {
	if err := s.ensureOrderMaterials(ctx, jobID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.material_id,m.name,m.consumption_unit,p.reservation_id,p.required_units,p.adjustment_units,j.quantity_units,j.outsource_quantity_units FROM production_material_plans p JOIN materials m ON m.id=p.material_id JOIN production_jobs j ON j.id=p.production_job_id WHERE p.production_job_id=? ORDER BY m.name,p.material_id`, jobID)
	if err != nil {
		return nil, err
	}
	type plan struct {
		material                             ProductionMaterial
		required, adjustment, qty, outsource domain.Quantity
	}
	plans := []plan{}
	for rows.Next() {
		var p plan
		if err = rows.Scan(&p.material.MaterialID, &p.material.Name, &p.material.Unit, &p.material.ReservationID, &p.required, &p.adjustment, &p.qty, &p.outsource); err != nil {
			rows.Close()
			return nil, err
		}
		plans = append(plans, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := []ProductionMaterial{}
	for _, p := range plans {
		target := p.required + p.adjustment
		if target < 0 || p.required == 0 {
			target = 0
		}
		target, err = scaleProductionQuantity(target, p.qty-p.outsource, p.qty)
		if err != nil {
			return nil, err
		}
		var used, reserved domain.Quantity
		if err = s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(-m.quantity_delta_units),0) FROM inventory_movements m JOIN production_consumptions c ON c.id=m.reference_id WHERE c.production_job_id=? AND c.material_id=? AND m.reference_type IN ('production_consumption','production_correction')`, jobID, p.material.MaterialID).Scan(&used); err != nil {
			return nil, err
		}
		if err = s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_units),0) FROM inventory_reservations WHERE production_job_id=? AND material_id=? AND status='active'`, jobID, p.material.MaterialID).Scan(&reserved); err != nil {
			return nil, err
		}
		stock, err := s.InventoryState(ctx, p.material.MaterialID)
		if err != nil {
			return nil, err
		}
		shortage := target - used - reserved
		if shortage < 0 {
			shortage = 0
		}
		p.material.Required = p.required.String()
		p.material.Planned = target.String()
		p.material.Reserved = reserved.String()
		p.material.Used = used.String()
		p.material.Available = stock.AvailableStock.String()
		p.material.Shortage = shortage.String()
		result = append(result, p.material)
	}
	return result, nil
}
