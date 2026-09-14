package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestMachinePersistenceAndLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "machines.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2024, time.August, 12, 7, 0, 0, 0, time.UTC)
	machine, err := domain.NewMachine("MAC-printer", domain.MachineDraft{Name: "Production Printer", Code: "PR-01", Category: "Digital print", ImagePath: "data:image/png;base64,machine", RateBasis: domain.RatePerUnit, RateRial: 123456789, SetupCostRial: 987654321, Rates: []domain.MachineRate{{ID: "standard", Name: "Standard", RateBasis: domain.RatePerUnit, RateRial: 123456789, SetupCostRial: 987654321, Active: true}, {ID: "full-color", Name: "Full color", SelectorValue: "full-color", SelectorPredefinedKey: domain.PredefinedParameterColor, RateBasis: domain.RatePerUnit, RateRial: 456789123, Active: true}}, Notes: "Main production line"}, now)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := store.SaveMachine(ctx, machine); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	got, err := store.GetMachine(ctx, machine.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.RateRial != 123456789 || got.SetupCostRial != 987654321 || got.RateBasis != domain.RatePerUnit || got.ImagePath != "data:image/png;base64,machine" {
		t.Fatalf("machine rate changed: %+v", got)
	}
	if len(got.Rates) != 2 || got.Rates[1].Name != "Full color" || got.Rates[1].RateRial != 456789123 || got.Rates[1].SelectorPredefinedKey != domain.PredefinedParameterColor || got.Rates[1].SelectorValue != "full-color" {
		t.Fatalf("machine rate profiles changed: %+v", got.Rates)
	}
	got.Active = false
	got.UpdatedAt = now.Add(time.Hour)
	if err := store.SaveMachine(ctx, got); err != nil {
		t.Fatal(err)
	}
	active, err := store.ListMachines(ctx, false)
	if err != nil || len(active) != 0 {
		t.Fatalf("active machines = %+v, err=%v", active, err)
	}
	got.Active = true
	if err := store.SaveMachine(ctx, got); err != nil {
		t.Fatal(err)
	}
	active, err = store.ListMachines(ctx, false)
	if err != nil || len(active) != 1 {
		t.Fatalf("reactivated machines = %+v, err=%v", active, err)
	}
}

func TestMachinePredefinedSelectorRejectsUnknownOption(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "machine-selector-validation.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	machine, err := domain.NewMachine("MAC-invalid-selector", domain.MachineDraft{Name: "Invalid selector", RateBasis: domain.RatePerUnit, Rates: []domain.MachineRate{{ID: "default", Name: "Standard", RateBasis: domain.RatePerUnit, Active: true}, {ID: "invalid", Name: "Invalid", SelectorValue: "not-in-catalog", SelectorPredefinedKey: domain.PredefinedParameterColor, RateBasis: domain.RatePerUnit, Active: true}}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveMachine(context.Background(), machine); err == nil {
		t.Fatal("expected unknown predefined selector option to be rejected")
	}
}
