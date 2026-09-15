package domain

import (
	"reflect"
	"testing"
)

func TestLegacyRollSizeDefaults(t *testing.T) {
	projected := Service{Category: "Large format", FinishedSize: &ServiceFinishedSizeDefinition{ParameterKey: "paper"}}
	if !projected.WithRollSizeDefaults().FinishedSize.AllowCustom {
		t.Fatal("legacy print-size projection must not hide custom roll length")
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
		})
	}
	for _, category := range []string{"Design", "Finishing", "Paper printing"} {
		service := Service{Category: category}
		if service.WithRollSizeDefaults().FinishedSize != nil {
			t.Fatalf("unexpected roll fields for %s", category)
		}
	}
}
