package domain

import (
	"reflect"
	"testing"
)

func TestLegacyRollSizeDefaults(t *testing.T) {
	projected := Service{Category: "Large format", Parameters: []ServiceParameter{{Key: "paper", PredefinedKey: PredefinedParameterPrintSize}}, FinishedSize: &ServiceFinishedSizeDefinition{ParameterKey: "paper", QuantityParameterKey: "quantity"}}
	if !projected.WithRollSizeDefaults().FinishedSize.AllowCustom {
		t.Fatal("legacy print-size projection must not hide custom roll length")
	}
	if len(projected.WithRollSizeDefaults().Parameters) != 5 {
		t.Fatal("legacy predefined print-size input must not remain as a hidden required input")
	}
	for _, parameter := range projected.WithRollSizeDefaults().Parameters {
		if parameter.Key == "paper" {
			t.Fatal("legacy predefined print-size input must not remain as a hidden required input")
		}
	}
	for _, category := range []string{"Large format", "Banners & signage", "Paper printing"} {
		t.Run(category, func(t *testing.T) {
			legacy := Service{ID: "roll-service", Category: category, Parameters: []ServiceParameter{
				{Key: "paper", Active: true, MaterialSource: &MaterialParameterSource{AllowedKinds: []MaterialKind{MaterialKindRollMedia}}},
			}}
			configured := legacy.WithRollSizeDefaults()
			if configured.FinishedSize == nil || !configured.FinishedSize.AllowCustom || configured.FinishedSize.QuantityParameterKey != "layout_quantity" || len(configured.Parameters) != 6 {
				t.Fatalf("legacy roll missing automatic inputs: %+v", configured)
			}
			if legacy.FinishedSize != nil || len(legacy.Parameters) != 1 {
				t.Fatal("normalizing a read mutated the stored definition")
			}
			if !reflect.DeepEqual(configured, configured.WithRollSizeDefaults()) {
				t.Fatal("normalization must be idempotent")
			}
			if configured.FinishedSize.ParameterKey != "" || len(configured.FinishedSize.Options) != 0 {
				t.Fatal("legacy predefined size must become material-driven custom sizing")
			}
		})
	}
	for _, category := range []string{"Design", "Finishing", "Paper printing"} {
		service := Service{Category: category}
		if service.WithRollSizeDefaults().FinishedSize != nil {
			t.Fatalf("unexpected roll fields for %s", category)
		}
	}
}
