package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestPurchasePostingCancellationAndSupplierProtection(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "inventory.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	supplier := domain.Supplier{ID: "SUP-test", Name: "Paper House", Active: true, CreatedAt: now, UpdatedAt: now}
	if err = store.SaveSupplier(ctx, supplier); err != nil {
		t.Fatal(err)
	}
	m, err := domain.NewMaterial("MAT-test", domain.MaterialDraft{Name: "Paper", PurchaseUnit: "pack", ConsumptionUnit: "sheet", ConversionFactor: 5 * domain.QuantityScale, PhysicalStock: 0, ReorderLevel: 0, AverageUnitCostRial: 0}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	p := domain.Purchase{ID: "PUR-test", SupplierID: supplier.ID, SupplierNameSnapshot: supplier.Name, PurchaseDate: now, Status: domain.PurchaseDraft, CreatedAt: now, UpdatedAt: now, Items: []domain.PurchaseItem{{ID: "PITM-test", MaterialID: m.ID, MaterialNameSnapshot: m.Name, PurchaseUnitSnapshot: m.PurchaseUnit, ConsumptionUnitSnapshot: m.ConsumptionUnit, PurchaseQuantity: 2 * domain.QuantityScale, ConversionFactorSnapshot: m.ConversionFactor, ConsumptionQuantity: 10 * domain.QuantityScale, UnitAcquisitionCostRial: 100, LineTotalRial: 200}}}
	p.SubtotalRial = 200
	p.TotalRial = 200
	if err = store.SavePurchase(ctx, p); err != nil {
		t.Fatal(err)
	}
	listed, err := store.ListPurchases(ctx)
	if err != nil || len(listed) != 1 || len(listed[0].Items) != 1 {
		t.Fatalf("listed purchase = %#v, err=%v", listed, err)
	}
	if err = store.PostPurchase(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if err = store.CancelInventoryMovement(ctx, "MOV-PITM-test"); !errors.Is(err, domain.ErrMovementCannotCancel) {
		t.Fatalf("purchase movement cancellation err=%v", err)
	}
	entries, err := store.ListJournalEntries(ctx)
	if err != nil || len(entries) != 1 || len(entries[0].Lines) != 2 {
		t.Fatalf("purchase journal entries=%d err=%v", len(entries), err)
	}
	got, err := store.Get(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.PhysicalStock != 10*domain.QuantityScale || got.AverageUnitCostRial != 20 || got.HighestPurchaseUnitCostRial != 20 {
		t.Fatalf("derived stock = %d, average cost=%d, highest purchase cost=%d", got.PhysicalStock, got.AverageUnitCostRial, got.HighestPurchaseUnitCostRial)
	}
	if err = store.DeleteDraftPurchase(ctx, p.ID); !errors.Is(err, domain.ErrPurchaseNotDraft) {
		t.Fatalf("posted delete err=%v", err)
	}
	if err = store.DeleteSupplier(ctx, supplier.ID); !errors.Is(err, domain.ErrSupplierDeleteProtected) {
		t.Fatalf("protected supplier delete err=%v", err)
	}
	if err = store.CancelPurchase(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if err = store.CancelPurchase(ctx, p.ID); err != nil {
		t.Fatal("idempotent cancellation:", err)
	}
	entries, err = store.ListJournalEntries(ctx)
	if err != nil || len(entries) != 2 {
		t.Fatalf("purchase reversal entries=%d err=%v", len(entries), err)
	}
	inventory, err := store.GetAccount(ctx, "ACC-INVENTORY")
	if err != nil || inventory.BalanceRial != 0 {
		t.Fatalf("inventory ledger after cancellation=%d err=%v", inventory.BalanceRial, err)
	}
	payable, err := store.GetAccount(ctx, "ACC-AP")
	if err != nil || payable.BalanceRial != 0 {
		t.Fatalf("payable ledger after cancellation=%d err=%v", payable.BalanceRial, err)
	}
	got, err = store.Get(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.PhysicalStock != 0 || got.AverageUnitCostRial != 0 {
		t.Fatalf("cancelled stock = %d, cost=%d", got.PhysicalStock, got.AverageUnitCostRial)
	}
	movs, err := store.ListInventoryMovements(ctx, m.ID)
	if err != nil || len(movs) != 2 {
		t.Fatalf("movements=%d err=%v", len(movs), err)
	}
	// Reposting the same item IDs must create fresh immutable movement IDs
	// instead of failing against the first posting/cancellation history.
	p.Status = domain.PurchaseDraft
	p.UpdatedAt = now.Add(time.Minute)
	if err = store.SavePurchase(ctx, p); err != nil {
		t.Fatal("reopen cancelled purchase:", err)
	}
	if err = store.PostPurchase(ctx, p.ID); err != nil {
		t.Fatal("repost cancelled purchase:", err)
	}
	if err = store.CancelPurchase(ctx, p.ID); err != nil {
		t.Fatal("recancel reposted purchase:", err)
	}
	movs, err = store.ListInventoryMovements(ctx, m.ID)
	if err != nil || len(movs) != 4 {
		t.Fatalf("movements after repost=%d err=%v", len(movs), err)
	}
}

func TestDraftDeleteAndManualAdjustmentUseLedger(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "adjustment.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	s := domain.Supplier{ID: "SUP-a", Name: "Supplier", Active: true, CreatedAt: now, UpdatedAt: now}
	if err = store.SaveSupplier(ctx, s); err != nil {
		t.Fatal(err)
	}
	m, _ := domain.NewMaterial("MAT-a", domain.MaterialDraft{Name: "Ink", PurchaseUnit: "liter", ConsumptionUnit: "milliliter", ConversionFactor: 1000 * domain.QuantityScale}, now)
	if err = store.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	p := domain.Purchase{ID: "PUR-draft", SupplierID: s.ID, SupplierNameSnapshot: s.Name, PurchaseDate: now, Status: domain.PurchaseDraft, CreatedAt: now, UpdatedAt: now}
	if err = store.SavePurchase(ctx, p); err != nil {
		t.Fatal(err)
	}
	if err = store.DeleteDraftPurchase(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = store.GetPurchase(ctx, p.ID); !errors.Is(err, domain.ErrPurchaseNotFound) {
		t.Fatalf("draft still exists: %v", err)
	}
	if err = store.AdjustInventory(ctx, m.ID, 2*domain.QuantityScale, 50, "count"); err != nil {
		t.Fatal(err)
	}
	if err = store.AdjustInventory(ctx, m.ID, -1*domain.QuantityScale, 0, "use"); err != nil {
		t.Fatal(err)
	}
	if err = store.AdjustInventory(ctx, m.ID, -2*domain.QuantityScale, 0, "too much"); !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("negative adjustment err=%v", err)
	}
	got, err := store.Get(ctx, m.ID)
	if err != nil || got.PhysicalStock != domain.QuantityScale {
		t.Fatalf("adjusted stock=%d err=%v", got.PhysicalStock, err)
	}
	movements, err := store.ListInventoryMovements(ctx, m.ID)
	if err != nil || len(movements) != 2 {
		t.Fatalf("manual movements=%d err=%v", len(movements), err)
	}
	if err = store.CancelInventoryMovement(ctx, movements[0].ID); err != nil {
		t.Fatalf("cancel manual movement: %v", err)
	}
	got, err = store.Get(ctx, m.ID)
	if err != nil || got.PhysicalStock != 2*domain.QuantityScale {
		t.Fatalf("stock after manual cancellation=%d err=%v", got.PhysicalStock, err)
	}
	movements, err = store.ListInventoryMovements(ctx, m.ID)
	if err != nil || len(movements) != 1 {
		t.Fatalf("visible movements after cancellation=%d err=%v", len(movements), err)
	}
	if err = store.CancelInventoryMovement(ctx, movements[0].ID); err != nil {
		t.Fatalf("cancel remaining manual movement: %v", err)
	}
	movements, err = store.ListInventoryMovements(ctx, m.ID)
	if err != nil || len(movements) != 0 {
		t.Fatalf("visible movements after full cancellation=%d err=%v", len(movements), err)
	}
}

func TestPostedPurchaseDeleteRemovesItsPaymentAndReturnsMaterialStock(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "purchase-delete.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)
	supplier := domain.Supplier{ID: "SUP-delete", Name: "Supplier", Active: true, CreatedAt: now, UpdatedAt: now}
	if err = store.SaveSupplier(ctx, supplier); err != nil {
		t.Fatal(err)
	}
	m, err := domain.NewMaterial("MAT-delete", domain.MaterialDraft{Name: "Stock material", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.QuantityScale}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	p := domain.Purchase{ID: "PUR-delete", SupplierID: supplier.ID, SupplierNameSnapshot: supplier.Name, PurchaseDate: now, Status: domain.PurchaseDraft, CreatedAt: now, UpdatedAt: now, Items: []domain.PurchaseItem{{ID: "PITM-delete", MaterialID: m.ID, MaterialNameSnapshot: m.Name, PurchaseUnitSnapshot: m.PurchaseUnit, ConsumptionUnitSnapshot: m.ConsumptionUnit, PurchaseQuantity: 5 * domain.QuantityScale, ConversionFactorSnapshot: domain.QuantityScale, ConsumptionQuantity: 5 * domain.QuantityScale, UnitAcquisitionCostRial: 100, LineTotalRial: 500}}, SubtotalRial: 500, TotalRial: 500}
	if err = store.SavePurchase(ctx, p); err != nil {
		t.Fatal(err)
	}
	if err = store.PostPurchase(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	payment := domain.Payment{ID: "PAY-delete", Direction: domain.PaymentOutgoing, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", SupplierID: supplier.ID, AmountRial: 500, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{ID: "AL-delete", TargetType: "purchase", TargetID: p.ID, AmountRial: 500}}}
	if _, err = store.CreatePayment(ctx, payment); err != nil {
		t.Fatal(err)
	}
	if blocked, depErr := store.PurchaseHasDependencies(ctx, p.ID); depErr != nil || blocked {
		t.Fatalf("payment should not block purchase deletion: blocked=%v err=%v", blocked, depErr)
	}
	if err = store.CancelPurchase(ctx, p.ID); err != nil {
		t.Fatal("return purchase stock:", err)
	}
	if err = store.DeletePurchase(ctx, p.ID); err != nil {
		t.Fatal("delete purchase:", err)
	}
	if _, err = store.GetPurchase(ctx, p.ID); !errors.Is(err, domain.ErrPurchaseNotFound) {
		t.Fatalf("purchase still exists: %v", err)
	}
	if _, err = store.GetPayment(ctx, payment.ID); !errors.Is(err, domain.ErrPaymentNotFound) {
		t.Fatalf("purchase payment still exists: %v", err)
	}
	got, err := store.Get(ctx, m.ID)
	value, valueErr := store.InventoryValue(ctx, m.ID)
	if err != nil || valueErr != nil || got.PhysicalStock != 0 || value != 0 {
		t.Fatalf("material inventory after purchase deletion: stock=%v value=%v err=%v valueErr=%v", got.PhysicalStock, value, err, valueErr)
	}
	var allocationCount, paymentCount int
	if err = store.db.QueryRow(`SELECT COUNT(*) FROM payment_allocations WHERE target_id=?`, p.ID).Scan(&allocationCount); err != nil {
		t.Fatal(err)
	}
	if err = store.db.QueryRow(`SELECT COUNT(*) FROM payments WHERE id=?`, payment.ID).Scan(&paymentCount); err != nil {
		t.Fatal(err)
	}
	if allocationCount != 0 || paymentCount != 0 {
		t.Fatalf("deleted purchase financial rows remain: allocations=%d payments=%d", allocationCount, paymentCount)
	}
}

func TestPurchaseCancellationCannotBreakActiveReservationAvailability(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "reserved-cancel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Date(2026, 1, 5, 8, 0, 0, 0, time.UTC)
	supplier := domain.Supplier{ID: "SUP-reserved-cancel", Name: "Supplier", Active: true, CreatedAt: now, UpdatedAt: now}
	if err = store.SaveSupplier(ctx, supplier); err != nil {
		t.Fatal(err)
	}
	m, err := domain.NewMaterial("MAT-reserved-cancel", domain.MaterialDraft{Name: "Stock", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.QuantityScale, PhysicalStock: 2 * domain.QuantityScale, AverageUnitCostRial: 10}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	p := domain.Purchase{ID: "PUR-reserved-cancel", SupplierID: supplier.ID, SupplierNameSnapshot: supplier.Name, PurchaseDate: now, Status: domain.PurchaseDraft, CreatedAt: now, UpdatedAt: now, Items: []domain.PurchaseItem{{ID: "PITM-reserved-cancel", MaterialID: m.ID, MaterialNameSnapshot: m.Name, PurchaseUnitSnapshot: m.PurchaseUnit, ConsumptionUnitSnapshot: m.ConsumptionUnit, PurchaseQuantity: 5 * domain.QuantityScale, ConversionFactorSnapshot: domain.QuantityScale, ConsumptionQuantity: 5 * domain.QuantityScale, UnitAcquisitionCostRial: 20, LineTotalRial: 100}}, SubtotalRial: 100, TotalRial: 100}
	if err = store.SavePurchase(ctx, p); err != nil {
		t.Fatal(err)
	}
	if err = store.PostPurchase(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if err = store.CreateReservation(ctx, domain.InventoryReservation{ID: "RES-reserved-cancel", MaterialID: m.ID, Quantity: 6 * domain.QuantityScale, Status: domain.ReservationActive}); err != nil {
		t.Fatal(err)
	}
	if err = store.CancelPurchase(ctx, p.ID); !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("cancellation reservation protection error=%v", err)
	}
	got, err := store.GetPurchase(ctx, p.ID)
	if err != nil || got.Status != domain.PurchasePosted {
		t.Fatalf("failed cancellation changed purchase: %+v, %v", got, err)
	}
	state, err := store.InventoryState(ctx, m.ID)
	if err != nil || state.PhysicalStock != 7*domain.QuantityScale || state.ReservedStock != 6*domain.QuantityScale {
		t.Fatalf("failed cancellation changed stock: %+v, %v", state, err)
	}
}
