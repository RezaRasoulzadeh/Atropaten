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
		Name: "Graphic Design",
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
		"choice default": base(ServiceParameterDraft{ID: "choice", Key: "size", Label: "Size", Type: ParameterChoice, Options: []string{"A4", "A3"}, DefaultValue: "Letter"}),
		"numeric bounds": base(ServiceParameterDraft{ID: "decimal", Key: "hours", Label: "Hours", Type: ParameterDecimal, MinValue: quantityPointer(8), MaxValue: quantityPointer(2)}),
		"boolean default": base(ServiceParameterDraft{ID: "bool", Key: "lamination", Label: "Lamination", Type: ParameterBoolean, DefaultValue: "yes"}),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewService("SVC-invalid", draft, now); err == nil {
				t.Fatalf("invalid service accepted")
			}
		})
	}
}

func quantityPointer(value int64) *Quantity {
	quantity := Quantity(value)
	return &quantity
}
