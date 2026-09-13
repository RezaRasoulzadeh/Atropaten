package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func completedProductionFixture(t *testing.T) (*Store, domain.Order, domain.ProductionJob) {
	t.Helper()
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	for _, status := range []string{domain.ProductionInProgress, domain.ProductionCompleted} {
		if err := s.TransitionProductionJob(ctx, j.ID, status); err != nil {
			t.Fatal(err)
		}
	}
	invoice, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	j, err = s.GetProductionJob(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	return s, o, j
}

// Capture complete rows, not just totals: reopening must preserve historical
// IDs, quantities, costs and timestamps as well as net inventory/accounting.
func productionReopenSnapshot(t *testing.T, s *Store, tables ...string) string {
	t.Helper()
	result := map[string][][]any{}
	for _, table := range tables {
		rows, err := s.db.Query(`SELECT * FROM ` + table + ` ORDER BY rowid`)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}
			if err = rows.Scan(pointers...); err != nil {
				t.Fatal(err)
			}
			result[table] = append(result[table], values)
		}
		if err = rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

var productionHistoryTables = []string{"production_consumptions", "inventory_movements", "journal_entries", "journal_lines", "expenses", "invoices"}
var productionReconciliationTables = []string{"production_jobs", "production_material_plans", "inventory_reservations", "orders"}

func assertReopenedProduction(t *testing.T, s *Store, before domain.ProductionJob, status string, quantity, required, reserved domain.Quantity, fulfillment domain.FulfillmentStatus) {
	t.Helper()
	ctx := context.Background()
	job, err := s.GetProductionJob(ctx, before.ID)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != status || job.CompletedAt != nil || job.Quantity != quantity || job.StartedAt == nil || !job.StartedAt.Equal(*before.StartedAt) {
		t.Fatalf("reopened job=%+v, before=%+v", job, before)
	}
	// Read stored allocations directly; ProductionMaterials would repair them.
	var actualRequired, actualReserved domain.Quantity
	if err = s.db.QueryRow(`SELECT required_units FROM production_material_plans WHERE production_job_id=? AND material_id='MAT-paper'`, job.ID).Scan(&actualRequired); err != nil {
		t.Fatal(err)
	}
	if err = s.db.QueryRow(`SELECT COALESCE(SUM(quantity_units),0) FROM inventory_reservations WHERE production_job_id=? AND material_id='MAT-paper' AND status='active'`, job.ID).Scan(&actualReserved); err != nil {
		t.Fatal(err)
	}
	if actualRequired != required || actualReserved != reserved {
		t.Fatalf("required=%s reserved=%s, want required=%s reserved=%s", actualRequired, actualReserved, required, reserved)
	}
	order, err := s.GetOrder(ctx, job.OrderID)
	if err != nil || order.FulfillmentStatus != fulfillment {
		t.Fatalf("fulfillment=%s want=%s err=%v", order.FulfillmentStatus, fulfillment, err)
	}
}

func assertProductionCOGS(t *testing.T, s *Store, want int64) {
	t.Helper()
	var actual int64
	if err := s.db.QueryRow(`SELECT COALESCE(SUM(debit_rial-credit_rial),0) FROM journal_lines WHERE account_id='ACC-COGS'`).Scan(&actual); err != nil || actual != want {
		t.Fatalf("COGS=%d want=%d err=%v", actual, want, err)
	}
}

func TestProductionReopeningPreservesActualsAndIsIdempotent(t *testing.T) {
	for _, status := range []string{domain.ProductionPending, domain.ProductionReady, domain.ProductionInProgress, domain.ProductionPaused, domain.ProductionFailed} {
		t.Run(status, func(t *testing.T) {
			s, _, j := completedProductionFixture(t)
			history := productionReopenSnapshot(t, s, productionHistoryTables...)
			if err := s.TransitionProductionJob(context.Background(), j.ID, status); err != nil {
				t.Fatal(err)
			}
			assertReopenedProduction(t, s, j, status, 10*domain.QuantityScale, 20*domain.QuantityScale, 0, domain.FulfillmentInProduction)
			if got := productionReopenSnapshot(t, s, productionHistoryTables...); got != history {
				t.Fatal("reopening changed historical consumption, inventory or accounting")
			}
			assertProductionCOGS(t, s, 2000)
			snapshot := productionReopenSnapshot(t, s, productionReconciliationTables...)
			if err := s.TransitionProductionJob(context.Background(), j.ID, status); err != nil {
				t.Fatal(err)
			}
			if productionReopenSnapshot(t, s, productionReconciliationTables...) != snapshot || productionReopenSnapshot(t, s, productionHistoryTables...) != history {
				t.Fatal("repeated reopening mutated state")
			}
		})
	}
}

func TestProductionReopeningRecomputesCurrentRequirements(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		quantity, required, reserved domain.Quantity
		usage                        string
	}{
		{"increased quantity", 15, 30, 10, "2"},
		{"decreased below actual usage", 5, 10, 0, "2"},
		{"changed material usage", 10, 30, 10, "3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, o, j := completedProductionFixture(t)
			// Seed a stale completed plan to exercise reopening independently
			// of the order-edit reconciliation path.
			if _, err := s.db.Exec(`UPDATE order_items SET quantity_units=?,cost_breakdown_json=? WHERE id=?`, tc.quantity*domain.QuantityScale, `[{"type":"material","enabled":true,"materialId":"MAT-paper","usageQuantity":"`+tc.usage+`"}]`, o.Items[0].ID); err != nil {
				t.Fatal(err)
			}
			history := productionReopenSnapshot(t, s, productionHistoryTables...)
			if err := s.TransitionProductionJob(context.Background(), j.ID, domain.ProductionInProgress); err != nil {
				t.Fatal(err)
			}
			assertReopenedProduction(t, s, j, domain.ProductionInProgress, tc.quantity*domain.QuantityScale, tc.required*domain.QuantityScale, tc.reserved*domain.QuantityScale, domain.FulfillmentInProduction)
			if productionReopenSnapshot(t, s, productionHistoryTables...) != history {
				t.Fatal("changed requirements rewrote history")
			}
			snapshot := productionReopenSnapshot(t, s, productionReconciliationTables...)
			if err := s.TransitionProductionJob(context.Background(), j.ID, domain.ProductionInProgress); err != nil {
				t.Fatal(err)
			}
			if productionReopenSnapshot(t, s, productionReconciliationTables...) != snapshot {
				t.Fatal("retry duplicated or changed reservations")
			}
		})
	}
}

func TestProductionReopeningAdditionalConsumptionAndRecompletion(t *testing.T) {
	s, o, j := completedProductionFixture(t)
	ctx := context.Background()
	if err := s.TransitionProductionJob(ctx, j.ID, domain.ProductionInProgress); err != nil {
		t.Fatal(err)
	}
	// An explicitly reopened job also follows the order-edit reconciliation path.
	o, err := s.GetOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	o.Items[0].Quantity = 15 * domain.QuantityScale
	if err = o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	assertReopenedProduction(t, s, j, domain.ProductionInProgress, 15*domain.QuantityScale, 30*domain.QuantityScale, 10*domain.QuantityScale, domain.FulfillmentInProduction)
	for i := 0; i < 2; i++ {
		if _, err = s.RecordProductionConsumption(ctx, j.ID, "MAT-paper", "reopen-additional", 3*domain.QuantityScale, domain.QuantityScale, "Additional run"); err != nil {
			t.Fatal(err)
		}
	}
	assertProductionCOGS(t, s, 2400)
	for i := 0; i < 2; i++ {
		if err = s.TransitionProductionJob(ctx, j.ID, domain.ProductionCompleted); err != nil {
			t.Fatal(err)
		}
	}
	state, err := s.InventoryState(ctx, "MAT-paper")
	if err != nil || state.PhysicalStock != 170*domain.QuantityScale || state.ReservedStock != 0 || state.InventoryValueRial != 17000 {
		t.Fatalf("final inventory=%+v err=%v", state, err)
	}
	assertProductionCOGS(t, s, 3000)
	if countRows(t, s, `SELECT COUNT(*) FROM production_consumptions`) != 3 || countRows(t, s, `SELECT COUNT(*) FROM journal_entries WHERE source_type='invoice_cogs_adjustment'`) != 2 {
		t.Fatal("additional usage/completion did not reconcile exactly once")
	}
	final, err := s.GetProductionJob(ctx, j.ID)
	if err != nil || final.CompletedAt == nil || final.ActualMaterialCostRial != 2900 || final.ActualWasteCostRial != 100 {
		t.Fatalf("final job=%+v err=%v", final, err)
	}
	o, err = s.GetOrder(ctx, o.ID)
	if err != nil || o.FulfillmentStatus != domain.FulfillmentReady {
		t.Fatalf("final order=%+v err=%v", o, err)
	}
}

func TestProductionReopeningUsesNetCorrections(t *testing.T) {
	s, _, j := completedProductionFixture(t)
	ctx := context.Background()
	usage, err := s.ListProductionConsumptions(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.ReverseProductionConsumption(ctx, usage[0].ID, "Actual material returned"); err != nil {
		t.Fatal(err)
	}
	history := productionReopenSnapshot(t, s, productionHistoryTables...)
	if err = s.TransitionProductionJob(ctx, j.ID, domain.ProductionInProgress); err != nil {
		t.Fatal(err)
	}
	assertReopenedProduction(t, s, j, domain.ProductionInProgress, 10*domain.QuantityScale, 20*domain.QuantityScale, 20*domain.QuantityScale, domain.FulfillmentInProduction)
	if productionReopenSnapshot(t, s, productionHistoryTables...) != history {
		t.Fatal("reopening modified corrected usage or COGS")
	}
	assertProductionCOGS(t, s, 0)
}

func TestProductionReopeningMultiJobFulfillment(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	item := o.Items[0]
	item.ID, item.Position = "ITEM-second", 1
	o.Items = append(o.Items, item)
	if err := o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveOrder(ctx, o); err != nil {
		t.Fatal(err)
	}
	second, err := s.GetProductionJob(ctx, "JOB-ORDER-"+item.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, jobID := range []string{j.ID, second.ID} {
		for _, status := range []string{domain.ProductionInProgress, domain.ProductionCompleted} {
			if err := s.TransitionProductionJob(ctx, jobID, status); err != nil {
				t.Fatal(err)
			}
		}
	}
	o, err = s.GetOrder(ctx, o.ID)
	if err != nil || o.FulfillmentStatus != domain.FulfillmentReady {
		t.Fatalf("completed order=%+v err=%v", o, err)
	}
	j, err = s.GetProductionJob(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	secondBefore, err := s.GetProductionJob(ctx, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, domain.ProductionPending); err != nil {
		t.Fatal(err)
	}
	assertReopenedProduction(t, s, j, domain.ProductionPending, 10*domain.QuantityScale, 20*domain.QuantityScale, 0, domain.FulfillmentInProduction)
	secondAfter, err := s.GetProductionJob(ctx, second.ID)
	if err != nil || secondAfter.Status != domain.ProductionCompleted || !secondAfter.CompletedAt.Equal(*secondBefore.CompletedAt) || !secondAfter.UpdatedAt.Equal(secondBefore.UpdatedAt) {
		t.Fatalf("other completed job changed: %+v err=%v", secondAfter, err)
	}
}

func TestProductionReopeningPreservesDeliveryAcrossCommercialStates(t *testing.T) {
	for _, commercial := range []domain.CommercialStatus{domain.CommercialConfirmed, domain.CommercialClosed, domain.CommercialCancelled, domain.CommercialDraft} {
		t.Run(string(commercial), func(t *testing.T) {
			s, o, j := completedProductionFixture(t)
			if _, err := s.db.Exec(`UPDATE orders SET commercial_status=?,fulfillment_status='Delivered' WHERE id=?`, commercial, o.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := s.db.Exec(`UPDATE order_items SET quantity_units=? WHERE id=?`, 15*domain.QuantityScale, o.Items[0].ID); err != nil {
				t.Fatal(err)
			}
			if err := s.TransitionProductionJob(context.Background(), j.ID, domain.ProductionReady); err != nil {
				t.Fatal(err)
			}
			assertReopenedProduction(t, s, j, domain.ProductionReady, 15*domain.QuantityScale, 30*domain.QuantityScale, 10*domain.QuantityScale, domain.FulfillmentDelivered)
		})
	}
}

func TestProductionReopeningFailureRollsBackEntireTransition(t *testing.T) {
	for _, failAt := range []string{"reservation", "fulfillment"} {
		t.Run(failAt, func(t *testing.T) {
			s, o, j := completedProductionFixture(t)
			ctx := context.Background()
			// The first material can be reserved. The second forces failure after
			// the first plan/reservation and job metadata have already changed.
			m, err := domain.NewMaterial("MAT-z", domain.MaterialDraft{Name: "Second material", PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale, AverageUnitCostRial: 100}, time.Now().UTC())
			if err != nil {
				t.Fatal(err)
			}
			if err = s.Create(ctx, m); err != nil {
				t.Fatal(err)
			}
			components := `[{"type":"material","enabled":true,"materialId":"MAT-paper","usageQuantity":"3"}]`
			if failAt == "reservation" {
				components = `[{"type":"material","enabled":true,"materialId":"MAT-paper","usageQuantity":"3"},{"type":"material","enabled":true,"materialId":"MAT-z","usageQuantity":"1"}]`
			} else if _, err = s.db.Exec(`CREATE TRIGGER reject_reopen_fulfillment BEFORE UPDATE OF fulfillment_status ON orders WHEN NEW.fulfillment_status='In Production' BEGIN SELECT RAISE(ABORT,'fulfillment failure'); END`); err != nil {
				t.Fatal(err)
			}
			if _, err = s.db.Exec(`UPDATE order_items SET cost_breakdown_json=? WHERE id=?`, components, o.Items[0].ID); err != nil {
				t.Fatal(err)
			}
			history := productionReopenSnapshot(t, s, productionHistoryTables...)
			state := productionReopenSnapshot(t, s, productionReconciliationTables...)
			for i := 0; i < 2; i++ {
				err = s.TransitionProductionJob(ctx, j.ID, domain.ProductionInProgress)
				if err == nil || (failAt == "reservation" && !errors.Is(err, domain.ErrReservationExceeded)) {
					t.Fatalf("reopen error=%v", err)
				}
				if productionReopenSnapshot(t, s, productionReconciliationTables...) != state || productionReopenSnapshot(t, s, productionHistoryTables...) != history {
					t.Fatal("failed reopen left partial changes")
				}
			}
		})
	}
}

func TestProductionReopeningReplacesMaterialSelectionWithoutReturningUsage(t *testing.T) {
	s, o, j := completedProductionFixture(t)
	ctx := context.Background()
	m, err := domain.NewMaterial("MAT-replacement", domain.MaterialDraft{Name: "Replacement", PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale, PhysicalStock: 50 * domain.QuantityScale, AverageUnitCostRial: 200}, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`UPDATE order_items SET resolved_parameters_json='[{"key":"paper","materialId":"MAT-replacement"}]' WHERE id=?`, o.Items[0].ID); err != nil {
		t.Fatal(err)
	}
	history := productionReopenSnapshot(t, s, productionHistoryTables...)
	if err = s.TransitionProductionJob(ctx, j.ID, domain.ProductionInProgress); err != nil {
		t.Fatal(err)
	}
	assertReopenedProduction(t, s, j, domain.ProductionInProgress, 10*domain.QuantityScale, 0, 0, domain.FulfillmentInProduction)
	stock, err := s.InventoryState(ctx, m.ID)
	if err != nil || stock.PhysicalStock != 50*domain.QuantityScale || stock.ReservedStock != 20*domain.QuantityScale {
		t.Fatalf("replacement inventory=%+v err=%v", stock, err)
	}
	if productionReopenSnapshot(t, s, productionHistoryTables...) != history {
		t.Fatal("material selection change rewrote actual usage")
	}
}

func TestProductionReopeningPreservesPlanOverrideAndOutsourcedShare(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	if err := s.SetProductionMaterialTarget(ctx, j.ID, "MAT-paper", 24*domain.QuantityScale); err != nil {
		t.Fatal(err)
	}
	j.OutsourceQuantity = 5 * domain.QuantityScale
	j.OutsourceUnitCostRial = 300
	j.OutsourceFinancialAccountID = "FIN-CASH"
	if err := s.UpdateProductionOutsourcing(ctx, j); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{domain.ProductionInProgress, domain.ProductionCompleted} {
		if err := s.TransitionProductionJob(ctx, j.ID, status); err != nil {
			t.Fatal(err)
		}
	}
	j, err := s.GetProductionJob(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	invoice, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`UPDATE order_items SET quantity_units=? WHERE id=?`, 20*domain.QuantityScale, o.Items[0].ID); err != nil {
		t.Fatal(err)
	}
	history := productionReopenSnapshot(t, s, productionHistoryTables...)
	if err = s.TransitionProductionJob(ctx, j.ID, domain.ProductionInProgress); err != nil {
		t.Fatal(err)
	}
	// Current requirement 40 + override 4, scaled to 15/20 in house = 33.
	// Twelve units were already consumed on the first completion, leaving 21.
	assertReopenedProduction(t, s, j, domain.ProductionInProgress, 20*domain.QuantityScale, 40*domain.QuantityScale, 21*domain.QuantityScale, domain.FulfillmentInProduction)
	if productionReopenSnapshot(t, s, productionHistoryTables...) != history {
		t.Fatal("reopening changed material usage or outsourcing expense")
	}
	// Inventory COGS is 1200; the existing outsourcing expense adds 1500.
	assertProductionCOGS(t, s, 2700)
}
