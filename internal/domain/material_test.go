package domain

import (
	"testing"
	"time"
)

func TestParseQuantityRoundTripsWithoutFloat(t *testing.T) {
	cases := map[string]string{"500": "500", "1.25": "1.25", "0.000001": "0.000001", "12.500000": "12.5"}
	for value, expected := range cases {
		quantity, err := ParseQuantity(value)
		if err != nil {
			t.Fatalf("ParseQuantity(%q): %v", value, err)
		}
		if expected != quantity.String() {
			t.Fatalf("ParseQuantity(%q) round-tripped as %q, want %q", value, quantity.String(), expected)
		}
	}
}

func TestMaterialValidation(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	_, err := NewMaterial("MAT-test", MaterialDraft{
		Name: "A4 paper", PurchaseUnit: "pack", ConsumptionUnit: "sheet", ConversionFactor: 500 * QuantityScale,
	}, now)
	if err != nil {
		t.Fatalf("valid material rejected: %v", err)
	}
	_, err = NewMaterial("MAT-test", MaterialDraft{
		Name: "", PurchaseUnit: "pack", ConsumptionUnit: "sheet", ConversionFactor: QuantityScale,
	}, now)
	if err == nil {
		t.Fatal("empty name accepted")
	}
	_, err = NewMaterial("MAT-test", MaterialDraft{
		Name: "A4 paper", PurchaseUnit: "unknown", ConsumptionUnit: "sheet", ConversionFactor: QuantityScale,
	}, now)
	if err == nil {
		t.Fatal("unknown unit accepted")
	}
}
