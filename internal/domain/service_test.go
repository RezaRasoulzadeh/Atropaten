package domain

import (
	"testing"
	"time"
)

func TestServiceValidationSupportsGenericParameterTypes(t *testing.T) {
	now := time.Date(2024, time.August, 12, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-digital", ServiceDraft{
		Name: "Digital Print",
		Parameters: []ServiceParameterDraft{
			{ID: "P-quantity", Key: "quantity", Label: "Quantity", Type: ParameterInteger, Required: true, DefaultValue: "1"},
			{ID: "P-size", Key: "paper_size", Label: "Paper size", Type: ParameterChoice, Options: []string{"A4", "A3"}, DefaultValue: "A4"},
			{ID: "P-color", Key: "color_mode", Label: "Color mode", Type: ParameterChoice, Options: []string{"B&W", "Color"}},
			{ID: "P-sides", Key: "sides", Label: "Sides", Type: ParameterChoice, Options: []string{"Single", "Double"}},
			{ID: "P-paper", Key: "paper", Label: "Paper", Type: ParameterMaterialReference},
		},
	}, now)
	if err != nil || len(service.Parameters) != 5 {
		t.Fatalf("generic service rejected: service=%+v err=%v", service, err)
	}
	if service.Parameters[0].Position != 0 || service.Parameters[4].Position != 4 {
		t.Fatalf("parameter order was not assigned deterministically: %+v", service.Parameters)
	}
	_, err = NewService("SVC-design", ServiceDraft{
		Name:       "Graphic Design",
		Parameters: []ServiceParameterDraft{{ID: "P-hours", Key: "estimated_hours", Label: "Estimated hours", Type: ParameterDecimal, Required: true, DefaultValue: "0"}},
	}, now)
	if err != nil {
		t.Fatalf("decimal service rejected: %v", err)
	}
}

func TestServiceParameterValidation(t *testing.T) {
	now := time.Date(2024, time.August, 12, 0, 0, 0, 0, time.UTC)
	base := func(parameter ServiceParameterDraft) ServiceDraft {
		return ServiceDraft{Name: "Test", Parameters: []ServiceParameterDraft{parameter}}
	}
	for name, draft := range map[string]ServiceDraft{
		"duplicate keys": {Name: "Test", Parameters: []ServiceParameterDraft{
			{ID: "one", Key: "quantity", Label: "Quantity", Type: ParameterInteger},
			{ID: "two", Key: "quantity", Label: "Other", Type: ParameterInteger},
		}},
		"choice default":  base(ServiceParameterDraft{ID: "choice", Key: "size", Label: "Size", Type: ParameterChoice, Options: []string{"A4", "A3"}, DefaultValue: "Letter"}),
		"numeric bounds":  base(ServiceParameterDraft{ID: "decimal", Key: "hours", Label: "Hours", Type: ParameterDecimal, MinValue: quantityPointer(8), MaxValue: quantityPointer(2)}),
		"boolean default": base(ServiceParameterDraft{ID: "bool", Key: "lamination", Label: "Lamination", Type: ParameterBoolean, DefaultValue: "yes"}),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewService("SVC-invalid", draft, now); err == nil {
				t.Fatalf("invalid service accepted")
			}
		})
	}
}

func TestServiceCostComponentValidationIsGeneric(t *testing.T) {
	now := time.Date(2024, time.August, 12, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-costs", ServiceDraft{
		Name: "Configurable service",
		Parameters: []ServiceParameterDraft{
			{ID: "P-quantity", Key: "quantity", Label: "Quantity", Type: ParameterInteger},
			{ID: "P-hours", Key: "hours", Label: "Hours", Type: ParameterDecimal},
			{ID: "P-flag", Key: "proof", Label: "Proof", Type: ParameterBoolean},
		},
		Components: []ServiceCostComponentDraft{
			{ID: "C-material", Name: "Paper", Type: CostMaterial, ReferenceID: "MAT-1", UsageMode: UsageParameter, ParameterKey: "quantity", Multiplier: QuantityScale},
			{ID: "C-machine", Name: "Printer", Type: CostMachine, ReferenceID: "MAC-1", UsageMode: UsageFixed, Multiplier: QuantityScale},
			{ID: "C-labor", Name: "Labor", Type: CostLabor, UsageMode: UsageParameter, ParameterKey: "hours", Multiplier: QuantityScale, RateRial: 120000, RateBasis: RatePerHour},
			{ID: "C-outsourced", Name: "Outsource", Type: CostOutsourced, UsageMode: UsageFixed, Multiplier: QuantityScale, RateRial: 80000},
			{ID: "C-fixed", Name: "Setup", Type: CostFixed, UsageMode: UsageFixed, Multiplier: QuantityScale, RateRial: 50000},
			{ID: "C-overhead", Name: "Overhead", Type: CostOverhead, UsageMode: UsageFixed, Multiplier: QuantityScale, Percentage: 10 * QuantityScale},
			{ID: "C-waste", Name: "Waste", Type: CostWaste, UsageMode: UsageFixed, Multiplier: QuantityScale, Percentage: 2 * QuantityScale},
			{ID: "C-manual", Name: "Manual", Type: CostManual, UsageMode: UsageFixed, Multiplier: QuantityScale},
		},
	}, now)
	if err != nil {
		t.Fatalf("generic component configuration rejected: %v", err)
	}
	if len(service.Components) != 8 || service.Components[7].Position != 7 {
		t.Fatalf("component order not assigned: %+v", service.Components)
	}

	badReference := service
	badReference.Components = append([]ServiceCostComponent(nil), service.Components...)
	badReference.Components[0].ParameterKey = "proof"
	if err := badReference.Validate(); err == nil {
		t.Fatal("boolean parameter reference accepted")
	}
	badReference = service
	badReference.Components = append([]ServiceCostComponent(nil), service.Components...)
	badReference.Components[0].ParameterKey = "missing"
	if err := badReference.Validate(); err == nil {
		t.Fatal("missing parameter reference accepted")
	}
}

func quantityPointer(value int64) *Quantity {
	quantity := Quantity(value)
	return &quantity
}
