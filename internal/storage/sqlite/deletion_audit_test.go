package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
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
	order := domain.NewOrder("ORD-delete-reference", protectedCustomer.ID, now)
	order.CustomerNameSnapshot = protectedCustomer.Name
	if err = s.CreateOrder(ctx, order); err != nil {
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
	historyOnly, _ := domain.NewMaterial("MAT-delete-history-only", domain.MaterialDraft{Name: "Reversed history", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.QuantityScale, PhysicalStock: 10 * domain.QuantityScale, AverageUnitCostRial: 25}, now)
	if err = s.Create(ctx, historyOnly); err != nil {
		t.Fatal(err)
	}
	if err = s.AdjustInventory(ctx, historyOnly.ID, -10*domain.QuantityScale, 25, "return all stock"); err != nil {
		t.Fatal("reverse history:", err)
	}
	if err = s.Delete(ctx, historyOnly.ID); err != nil {
		t.Fatal("reversed history should not block material purge:", err)
	}
	var movementHistory, materialRows int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM inventory_movements WHERE material_id=?`, historyOnly.ID).Scan(&movementHistory); err != nil {
		t.Fatal(err)
	}
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM materials WHERE id=?`, historyOnly.ID).Scan(&materialRows); err != nil {
		t.Fatal(err)
	}
	if movementHistory != 0 || materialRows != 0 {
		t.Fatalf("material history purge left rows: movements=%d materials=%d", movementHistory, materialRows)
	}
	reservedOnly, _ := domain.NewMaterial("MAT-delete-released-reservation", domain.MaterialDraft{Name: "Released reservation", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.QuantityScale}, now)
	if err = s.Create(ctx, reservedOnly); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`INSERT INTO inventory_reservations(id,material_id,quantity_units,status,created_at,updated_at) VALUES(?,?,?,?,?,?)`, "RES-delete-released", reservedOnly.ID, 1, "released", now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		t.Fatal("seed released reservation:", err)
	}
	if err = s.Delete(ctx, reservedOnly.ID); err != nil {
		t.Fatal("released reservation should not block material purge:", err)
	}
	var reservations int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM inventory_reservations WHERE id=?`, "RES-delete-released").Scan(&reservations); err != nil {
		t.Fatal(err)
	}
	if reservations != 0 {
		t.Fatalf("released reservation remains after material purge: %d", reservations)
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
	dependentService := domain.Service{ID: "SRV-dependent", Name: "Dependent service", Active: true, CreatedAt: now, UpdatedAt: now, Components: []domain.ServiceCostComponent{{ID: "COMP-service-reference", ServiceID: "SRV-dependent", Name: "Included service", Type: domain.CostService, ReferenceID: protectedService.ID, UsageMode: domain.UsageFixed, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true, CreatedAt: now, UpdatedAt: now}}}
	if err = s.SaveServiceDefinition(ctx, dependentService); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteService(ctx, protectedService.ID); !errors.Is(err, domain.ErrServiceDeleteProtected) {
		t.Fatalf("referenced service delete error=%v", err)
	}
}

func TestMaterialDeletionReportsProductionPlanDependency(t *testing.T) {
	s, _, job := productionFlowFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()
	planned, err := domain.NewMaterial("MAT-delete-planned", domain.MaterialDraft{Name: "Planned material", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.QuantityScale}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, planned); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`INSERT INTO production_material_plans(production_job_id,material_id,required_units,adjustment_units,reservation_id) VALUES(?,?,?,?,?)`, job.ID, planned.ID, 0, 0, "RES-delete-planned"); err != nil {
		t.Fatal("seed production plan:", err)
	}
	err = s.Delete(ctx, planned.ID)
	if !errors.Is(err, domain.ErrMaterialDeleteProtected) {
		t.Fatalf("production plan delete error=%v", err)
	}
	if strings.Contains(strings.ToLower(err.Error()), "foreign key") {
		t.Fatalf("production plan leaked raw foreign-key error: %v", err)
	}
}
