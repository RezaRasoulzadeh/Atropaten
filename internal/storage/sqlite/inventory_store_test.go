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
	entries, err := store.ListJournalEntries(ctx)
	if err != nil || len(entries) != 1 || len(entries[0].Lines) != 2 {
		t.Fatalf("purchase journal entries=%d err=%v", len(entries), err)
	}
	got, err := store.Get(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.PhysicalStock != 10*domain.QuantityScale || got.AverageUnitCostRial != 20 {
		t.Fatalf("derived stock = %d, cost=%d", got.PhysicalStock, got.AverageUnitCostRial)
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
}
