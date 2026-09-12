package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"Atropaten/internal/domain"
)

func sumProductionCosts(values ...int64) (int64, error) {
	var total big.Int
	for _, v := range values {
		total.Add(&total, big.NewInt(v))
	}
	if !total.IsInt64() {
		return 0, fmt.Errorf("production cost is too large")
	}
	return total.Int64(), nil
}

// Conversion costs remain estimates until separately recorded as expenses.
// Material and percentage waste estimates are replaced by stock usage; overhead,
// labor, machines and other service costs retain their saved pricing basis.
func productionConversionCost(snapshot string, quantity domain.Quantity, fallback int64) (int64, error) {
	if !strings.HasPrefix(strings.TrimSpace(snapshot), "[") {
		return fallback, nil
	}
	var components []struct {
		Type    string `json:"type"`
		Enabled bool   `json:"enabled"`
		Amount  int64  `json:"amountRial"`
	}
	if err := json.Unmarshal([]byte(snapshot), &components); err != nil {
		return 0, err
	}
	if len(components) == 0 {
		return fallback, nil
	}
	var perUnit int64
	for _, c := range components {
		if !c.Enabled || c.Type == "material" || c.Type == "waste" {
			continue
		}
		if c.Amount < 0 {
			return 0, fmt.Errorf("negative production component cost")
		}
		var err error
		perUnit, err = sumProductionCosts(perUnit, c.Amount)
		if err != nil {
			return 0, err
		}
	}
	return domain.MulQuantityRial(quantity, perUnit)
}

func (s *Store) withProductionForecast(ctx context.Context, j domain.ProductionJob) (domain.ProductionJob, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return j, err
	}
	defer tx.Rollback()
	// Older open jobs may not have material plans yet. Reconcile in this read
	// transaction to forecast them accurately; rollback keeps reads non-mutating.
	if j.Status != domain.ProductionCompleted && j.Status != domain.ProductionCancelled {
		if err = syncProductionMaterialsTx(ctx, tx, j.ID); err != nil {
			return j, err
		}
	}
	var snapshot string
	if err = tx.QueryRowContext(ctx, `SELECT cost_breakdown_json FROM production_jobs WHERE id=?`, j.ID).Scan(&snapshot); err != nil {
		return j, err
	}
	fallback, err := scaleProductionQuantity(domain.Quantity(j.EstimatedCostRial), j.Quantity-j.OutsourceQuantity, j.Quantity)
	if err != nil {
		return j, err
	}
	j.EstimatedConversionCostRial, err = productionConversionCost(snapshot, j.Quantity-j.OutsourceQuantity, int64(fallback))
	if err != nil {
		return j, err
	}
	if j.Status == domain.ProductionCancelled {
		j.EstimatedConversionCostRial = 0
	}
	if j.Status != domain.ProductionCompleted && j.Status != domain.ProductionCancelled {
		remaining, err := remainingProductionMaterialsTx(ctx, tx, j.ID)
		if err != nil {
			return j, err
		}
		for id, q := range remaining {
			state, err := inventoryStateTx(ctx, tx, id)
			if err != nil {
				return j, err
			}
			var cost int64
			if state.PhysicalStock > 0 {
				value, e := scaleProductionQuantity(domain.Quantity(state.InventoryValueRial), q, state.PhysicalStock)
				cost, err = int64(value), e
			} else {
				// A shortage is not free material: retain a known acquisition rate
				// after stock is exhausted until the next purchase supplies a new rate.
				var rate int64
				if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(unit_cost_rial),0) FROM inventory_movements WHERE material_id=? AND quantity_delta_units>0`, id).Scan(&rate); err != nil {
					return j, err
				}
				cost, err = domain.MulQuantityRial(q, rate)
			}
			if err != nil {
				return j, err
			}
			j.RemainingMaterialCostRial, err = sumProductionCosts(j.RemainingMaterialCostRial, cost)
			if err != nil {
				return j, err
			}
		}
	}
	// Legacy snapshots without components cannot distinguish conversion costs.
	// Carry only the unaccounted part of the estimate, avoiding double counting.
	if !strings.HasPrefix(strings.TrimSpace(snapshot), "[") || strings.TrimSpace(snapshot) == "[]" {
		known, err := sumProductionCosts(j.ActualMaterialCostRial, j.ActualWasteCostRial, j.RemainingMaterialCostRial)
		if err != nil {
			return j, err
		}
		j.EstimatedConversionCostRial = int64(fallback) - known
		if j.EstimatedConversionCostRial < 0 || j.Status == domain.ProductionCancelled {
			j.EstimatedConversionCostRial = 0
		}
	}
	j.ProjectedCostRial, err = sumProductionCosts(j.ActualMaterialCostRial, j.ActualWasteCostRial, j.ActualOutsourcedCostRial, j.EstimatedConversionCostRial, j.RemainingMaterialCostRial)
	return j, err
}

// Unscheduled items keep their estimate. Cancelled work still carries costs that
// were actually incurred; a replacement job does not erase those losses.
func (s *Store) ProductionProjectedCostSummary(ctx context.Context, orderID string) (int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM production_jobs WHERE order_id=?`, orderID)
	if err != nil {
		return 0, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	var total int64
	if err = s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(estimated_cost_rial),0) FROM order_items i WHERE i.order_id=? AND NOT EXISTS(SELECT 1 FROM production_jobs p WHERE p.order_item_id=i.id AND p.status<>'Cancelled')`, orderID).Scan(&total); err != nil {
		return 0, err
	}
	for _, id := range ids {
		j, err := s.GetProductionJob(ctx, id)
		if err != nil {
			return 0, err
		}
		total, err = sumProductionCosts(total, j.ProjectedCostRial)
		if err != nil {
			return 0, err
		}
	}
	return total, nil
}
