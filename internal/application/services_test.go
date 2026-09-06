package application

import (
	"context"
	"testing"

	"Atropaten/internal/domain"
)

type serviceRepositoryStub struct{ service domain.Service }

func (r *serviceRepositoryStub) ListServices(_ context.Context, _ bool) ([]domain.Service, error) {
	if r.service.ID == "" {
		return []domain.Service{}, nil
	}
	return []domain.Service{r.service}, nil
}
func (r *serviceRepositoryStub) GetService(_ context.Context, _ string) (domain.Service, error) {
	if r.service.ID == "" {
		return domain.Service{}, domain.ErrServiceNotFound
	}
	return r.service, nil
}
func (r *serviceRepositoryStub) SaveServiceDefinition(_ context.Context, service domain.Service) error {
	r.service = service
	return nil
}

type materialLookupStub struct{ material domain.Material }

func (m materialLookupStub) Get(_ context.Context, id string) (domain.Material, error) {
	if m.material.ID == "" || m.material.ID != id {
		return domain.Material{}, domain.ErrMaterialNotFound
	}
	return m.material, nil
}

type machineLookupStub struct{ machine domain.Machine }

func (m machineLookupStub) GetMachine(_ context.Context, id string) (domain.Machine, error) {
	if m.machine.ID == "" || m.machine.ID != id {
		return domain.Machine{}, domain.ErrMachineNotFound
	}
	return m.machine, nil
}

func TestServiceComponentReferencesAndBrokenParameterUpdateAreRejected(t *testing.T) {
	repository := &serviceRepositoryStub{}
	machine := domain.Machine{ID: "MAC-1", Name: "Printer", Active: true}
	material := domain.Material{ID: "MAT-1", Name: "Paper", Active: true}
	service := NewServicesService(repository, materialLookupStub{material: material}, machineLookupStub{machine: machine})
	created, err := service.Create(context.Background(), ServiceInput{
		Name:       "Digital Print",
		Parameters: []ParameterInput{{ID: "P-quantity", Key: "quantity", Label: "Quantity", Type: string(domain.ParameterInteger)}},
		Components: []CostComponentInput{{ID: "C-paper", Name: "Paper", Type: string(domain.CostMaterial), ReferenceID: "MAT-1", UsageMode: string(domain.UsageParameter), ParameterKey: "quantity", Multiplier: "1"}, {ID: "C-printer", Name: "Printer", Type: string(domain.CostMachine), ReferenceID: "MAC-1", UsageMode: string(domain.UsageFixed), Multiplier: "1"}},
	})
	if err != nil {
		t.Fatalf("create service with valid references: %v", err)
	}
	if len(created.Components) != 2 {
		t.Fatalf("components = %+v", created.Components)
	}
	_, err = service.Update(context.Background(), created.ID, ServiceInput{Name: "Digital Print", Components: []CostComponentInput{{ID: "C-paper", Name: "Paper", Type: string(domain.CostMaterial), ReferenceID: "MAT-1", UsageMode: string(domain.UsageParameter), ParameterKey: "quantity", Multiplier: "1"}}})
	if err == nil {
		t.Fatal("broken parameter reference was accepted")
	}
	if repository.service.Name != "Digital Print" || len(repository.service.Components) != 2 {
		t.Fatal("invalid update changed persisted service")
	}
	_, err = service.Create(context.Background(), ServiceInput{Name: "Invalid machine", Components: []CostComponentInput{{Name: "Printer", Type: string(domain.CostMachine), ReferenceID: "MAC-missing", UsageMode: string(domain.UsageFixed), Multiplier: "1"}}})
	if err == nil {
		t.Fatal("missing machine reference was accepted")
	}
}
