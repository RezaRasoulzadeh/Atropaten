package domain

import (
	"math"
	"testing"
)

func TestBannerLayoutAndWaste(t *testing.T) {
	material := Material{ID: "roll", Kind: MaterialKindRollMedia, ConsumptionUnit: "meter", Attributes: []MaterialAttributeValue{{Key: "width_mm", ValueType: MaterialAttributeDecimal, DecimalValue: 3200 * QuantityScale}}}
	service := Service{FinishedSize: &ServiceFinishedSizeDefinition{AllowCustom: true, AllowRotation: true, WidthParameterKey: "w", HeightParameterKey: "h", QuantityParameterKey: "q"}}
	parameters := map[string]ResolvedParameter{"w": {Value: "1000"}, "h": {Value: "3000"}, "q": {Quantity: QuantityScale}}
	layout, err := CalculatePrintLayout(material, service, parameters)
	if err != nil {
		t.Fatal(err)
	}
	if !layout.Rotated || layout.LengthMM != "1000" || layout.OriginalLengthMM != "3000" || layout.ConsumedQuantity != "1" || math.Abs(layout.WastePercent-6.25) > 0.001 {
		t.Fatalf("unexpected banner layout: %+v", layout)
	}
	if wasteCost, err := CalculateMaterialWasteCost(layout, 100_000); err != nil || wasteCost != 6_250 {
		t.Fatalf("waste cost = %d, %v; want 6250 Rial", wasteCost, err)
	}
	parameters["w"] = ResolvedParameter{Value: "3300"}
	if _, err := CalculatePrintLayout(material, service, parameters); err == nil {
		t.Fatal("finished width greater than the material width was accepted")
	}
	parameters["w"] = ResolvedParameter{Value: "1000"}
	parameters["q"] = ResolvedParameter{Quantity: 3 * QuantityScale}
	layout, err = CalculatePrintLayout(material, service, parameters)
	if err != nil || layout.LengthMM != "3000" || layout.Rotated {
		t.Fatalf("three banners should tie and retain orientation: %+v %v", layout, err)
	}
	service.FinishedSize.AllowRotation = false
	parameters["q"] = ResolvedParameter{Quantity: QuantityScale}
	layout, err = CalculatePrintLayout(material, service, parameters)
	if err != nil || layout.Rotated || layout.ConsumedQuantity != "3" {
		t.Fatalf("rotation lock: %+v %v", layout, err)
	}
	material.ConsumptionUnit = "square meter"
	service.FinishedSize.AllowRotation = true
	layout, err = CalculatePrintLayout(material, service, parameters)
	if err != nil || layout.ConsumedQuantity != "3.2" {
		t.Fatalf("area must include unused width: %+v %v", layout, err)
	}
}

func TestLayoutSheetMarginsAndInvalidInputs(t *testing.T) {
	material := Material{Kind: MaterialKindSheetStock, ConsumptionUnit: "sheet", Attributes: []MaterialAttributeValue{{Key: "width_mm", ValueType: MaterialAttributeDecimal, DecimalValue: 200 * QuantityScale}, {Key: "height_mm", ValueType: MaterialAttributeDecimal, DecimalValue: 200 * QuantityScale}}}
	service := Service{FinishedSize: &ServiceFinishedSizeDefinition{AllowCustom: true, AllowRotation: true, WidthParameterKey: "w", HeightParameterKey: "h", QuantityParameterKey: "q"}}
	parameters := map[string]ResolvedParameter{"w": {Value: "100"}, "h": {Value: "100"}, "q": {Quantity: 5 * QuantityScale}}
	layout, err := CalculatePrintLayout(material, service, parameters)
	if err != nil || layout.ItemsPerSheet != 4 || layout.Sheets != "2" || layout.WastePercent != 37.5 {
		t.Fatalf("sheet batch: %+v %v", layout, err)
	}
	parameters["layout_gap_mm"] = ResolvedParameter{Quantity: QuantityScale}
	layout, err = CalculatePrintLayout(material, service, parameters)
	if err != nil || layout.ItemsPerSheet != 1 || layout.Sheets != "5" {
		t.Fatalf("cutting gaps: %+v %v", layout, err)
	}
	parameters["layout_margin_mm"] = ResolvedParameter{Quantity: 101 * QuantityScale}
	if _, err = CalculatePrintLayout(material, service, parameters); err == nil {
		t.Fatal("expected no printable area")
	}
	delete(parameters, "layout_margin_mm")
	parameters["q"] = ResolvedParameter{Quantity: QuantityScale / 2}
	if _, err = CalculatePrintLayout(material, service, parameters); err == nil {
		t.Fatal("fractional pieces must fail")
	}
	if q, err := rollQuantity(1, 3200*QuantityScale, "meter"); err != nil || q != 1 {
		t.Fatalf("tiny positive consumption rounded to zero: %v %v", q, err)
	}
}
