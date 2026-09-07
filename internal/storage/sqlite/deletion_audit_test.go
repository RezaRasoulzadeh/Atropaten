package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestCatalogDeletionPurgesOnlyUnreferencedRecords(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "deletion.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 1, 6, 8, 0, 0, 0, time.UTC)

	customer, _ := domain.NewCustomer("CUS-delete-safe", domain.CustomerDraft{Name: "Safe customer"}, now)
	if err = s.SaveCustomer(ctx, customer); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteCustomer(ctx, customer.ID); err != nil {
		t.Fatal("safe customer purge:", err)
	}
	if _, err = s.GetCustomer(ctx, customer.ID); !errors.Is(err, domain.ErrCustomerNotFound) {
		t.Fatalf("customer remains after purge: %v", err)
	}

	protectedCustomer, _ := domain.NewCustomer("CUS-delete-protected", domain.CustomerDraft{Name: "Protected customer"}, now)
	if err = s.SaveCustomer(ctx, protectedCustomer); err != nil {
		t.Fatal(err)
	}
	quote := domain.NewQuote("QUO-delete-reference", protectedCustomer.ID, now)
	quote.Items = []domain.QuoteItem{{ID: "QITEM-delete-reference", QuoteID: quote.ID, Position: 0, ServiceID: "SRV-delete-protected", ServiceNameSnapshot: "Historical service", Quantity: domain.QuantityScale, QuantityUnit: "unit", ResolvedParametersJSON: "{}", CostBreakdownJSON: "[]", PricingSnapshotJSON: "{}", SellingPriceRial: 100}}
	if err = quote.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.CreateQuote(ctx, quote); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteCustomer(ctx, protectedCustomer.ID); !errors.Is(err, domain.ErrCustomerDeleteProtected) {
		t.Fatalf("referenced customer delete error=%v", err)
	}

	mat, _ := domain.NewMaterial("MAT-delete-safe", domain.MaterialDraft{Name: "Safe material", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.QuantityScale}, now)
	if err = s.Create(ctx, mat); err != nil {
		t.Fatal(err)
	}
	if err = s.Delete(ctx, mat.ID); err != nil {
		t.Fatal("safe material purge:", err)
	}
	protectedMat, _ := domain.NewMaterial("MAT-delete-protected", domain.MaterialDraft{Name: "Referenced material", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.QuantityScale}, now)
	if err = s.Create(ctx, protectedMat); err != nil {
		t.Fatal(err)
	}
	if err = s.AdjustInventory(ctx, protectedMat.ID, domain.QuantityScale, 10, "audit reference"); err != nil {
		t.Fatal(err)
	}
	if err = s.Delete(ctx, protectedMat.ID); !errors.Is(err, domain.ErrMaterialDeleteProtected) {
		t.Fatalf("referenced material delete error=%v", err)
	}

	safeMachine, _ := domain.NewMachine("MAC-delete-safe", domain.MachineDraft{Name: "Safe machine", RateBasis: domain.RatePerUnit}, now)
	if err = s.SaveMachine(ctx, safeMachine); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteMachine(ctx, safeMachine.ID); err != nil {
		t.Fatal("safe machine purge:", err)
	}
	protectedMachine, _ := domain.NewMachine("MAC-delete-protected", domain.MachineDraft{Name: "Referenced machine", RateBasis: domain.RatePerUnit}, now)
	if err = s.SaveMachine(ctx, protectedMachine); err != nil {
		t.Fatal(err)
	}
	machineService := domain.Service{ID: "SRV-machine-reference", Name: "Machine recipe", Active: true, CreatedAt: now, UpdatedAt: now, Components: []domain.ServiceCostComponent{{ID: "COMP-machine-reference", ServiceID: "SRV-machine-reference", Name: "Printer", Type: domain.CostMachine, ReferenceID: protectedMachine.ID, UsageMode: domain.UsageFixed, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true, CreatedAt: now, UpdatedAt: now}}}
	if err = s.SaveServiceDefinition(ctx, machineService); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteMachine(ctx, protectedMachine.ID); !errors.Is(err, domain.ErrMachineDeleteProtected) {
		t.Fatalf("referenced machine delete error=%v", err)
	}

	safeService := domain.Service{ID: "SRV-delete-safe", Name: "Safe service", Active: true, CreatedAt: now, UpdatedAt: now}
	if err = s.SaveServiceDefinition(ctx, safeService); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteService(ctx, safeService.ID); err != nil {
		t.Fatal("safe service purge:", err)
	}
	protectedService := domain.Service{ID: "SRV-delete-protected", Name: "Historical service", Active: true, CreatedAt: now, UpdatedAt: now}
	if err = s.SaveServiceDefinition(ctx, protectedService); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteService(ctx, protectedService.ID); !errors.Is(err, domain.ErrServiceDeleteProtected) {
		t.Fatalf("referenced service delete error=%v", err)
	}
}
