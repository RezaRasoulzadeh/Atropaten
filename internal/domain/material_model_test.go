package domain

import (
	"testing"
	"time"
)

func testMaterial(t *testing.T, id string, kind MaterialKind, attrs ...MaterialAttributeValue) Material {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m, err := NewMaterial(id, MaterialDraft{Name: id, Kind: kind, PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: QuantityScale, Attributes: attrs}, now)
	if err != nil {
		t.Fatalf("material %s: %v", id, err)
	}
	return m
}

func TestMaterialResolverIntersectsStructuredSelections(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	attrs := func(grammage int64, finish string) []MaterialAttributeValue {
		return []MaterialAttributeValue{{Key: "grammage_gsm", ValueType: MaterialAttributeInteger, IntegerValue: grammage}, {Key: "finish", ValueType: MaterialAttributeEnum, EnumCode: finish}}
	}
	a := testMaterial(t, "MAT-matte", MaterialKindSheetStock, attrs(170, "matte")...)
	b := testMaterial(t, "MAT-gloss", MaterialKindSheetStock, attrs(170, "gloss")...)
	c := testMaterial(t, "MAT-other", MaterialKindSheetStock, attrs(120, "matte")...)
	service, err := NewService("SVC-structured", ServiceDraft{Name: "Structured", Parameters: []ServiceParameterDraft{
		{ID: "P-grammage", Key: "grammage", Label: "Grammage", Type: ParameterChoice, Required: true, MaterialSource: &MaterialParameterSource{AllowedKinds: []MaterialKind{MaterialKindSheetStock}, ExposedAttributeKey: "grammage_gsm"}},
		{ID: "P-finish", Key: "finish", Label: "Finish", Type: ParameterChoice, Required: true, MaterialSource: &MaterialParameterSource{AllowedKinds: []MaterialKind{MaterialKindSheetStock}, ExposedAttributeKey: "finish"}},
	}, Components: []ServiceCostComponentDraft{{ID: "C", Name: "Material", Type: CostMaterial, UsageMode: UsageParameter, ParameterKey: "grammage", Multiplier: QuantityScale, Enabled: true}}}, now)
	if err != nil {
		t.Fatal(err)
	}
	compatible, err := CompatibleMaterials(service, []Material{a, b, c}, map[string]string{"grammage": "170"})
	if err != nil || len(compatible) != 2 {
		t.Fatalf("compatible after grammage=%v err=%v", compatible, err)
	}
	compatible, err = CompatibleMaterials(service, []Material{a, b, c}, map[string]string{"grammage": "170", "finish": "matte"})
	if err != nil || len(compatible) != 1 || compatible[0].ID != a.ID {
		t.Fatalf("compatible after intersection=%v err=%v", compatible, err)
	}
	options, err := MaterialOptionsForParameter(service, "finish", []Material{a, b, c}, map[string]string{"grammage": "170"})
	if err != nil || len(options) != 2 || options[0].Value != "gloss" || options[1].Value != "matte" {
		t.Fatalf("derived options=%+v err=%v", options, err)
	}
}

func TestConfiguredMaterialVariantResolvesExactInventoryMaterial(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	material, err := NewMaterial("MAT-exact", MaterialDraft{Name: "Exact", Kind: MaterialKindSheetStock, PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: QuantityScale, Attributes: []MaterialAttributeValue{{Key: "finish", ValueType: MaterialAttributeEnum, EnumCode: "matte"}}}, now)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService("SVC-variant", ServiceDraft{Name: "Variant", Parameters: []ServiceParameterDraft{{ID: "P-finish", Key: "finish", Label: "Finish", Type: ParameterChoice, Required: true, MaterialSource: &MaterialParameterSource{AllowedKinds: []MaterialKind{MaterialKindSheetStock}, ExposedAttributeKey: "finish"}}}, MaterialVariants: []ServiceMaterialVariant{{ID: "VAR-1", MaterialID: material.ID, Values: map[string]string{"finish": "matte"}, Position: 0, Active: true}}}, now)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ResolveMaterialSelection(service, []Material{material}, map[string]string{"finish": "matte"})
	if err != nil || got.ID != material.ID {
		t.Fatalf("exact variant=%+v err=%v", got, err)
	}
	if _, err := ResolveMaterialSelection(service, []Material{material}, map[string]string{"finish": "gloss"}); err == nil {
		t.Fatal("unavailable variant was accepted")
	}
	options, err := MaterialOptionsForParameter(service, "finish", []Material{material}, nil)
	if err != nil || len(options) != 1 || options[0].Value != "matte" || len(options[0].MaterialIDs) != 1 {
		t.Fatalf("variant-derived options=%+v err=%v", options, err)
	}
}

func TestCompositeMaterialGroupResolvesSizeAndTypeCombination(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	attrs := func(width, height int64, finish string) []MaterialAttributeValue {
		return []MaterialAttributeValue{
			{Key: "width_mm", ValueType: MaterialAttributeDecimal, DecimalValue: Quantity(width) * QuantityScale},
			{Key: "height_mm", ValueType: MaterialAttributeDecimal, DecimalValue: Quantity(height) * QuantityScale},
			{Key: "finish", ValueType: MaterialAttributeEnum, EnumCode: finish},
		}
	}
	a4Matte := testMaterial(t, "MAT-a4-matte", MaterialKindSheetStock, attrs(210, 297, "matte")...)
	a4Gloss := testMaterial(t, "MAT-a4-gloss", MaterialKindSheetStock, attrs(210, 297, "gloss")...)
	a3Matte := testMaterial(t, "MAT-a3-matte", MaterialKindSheetStock, attrs(297, 420, "matte")...)
	service, err := NewService("SVC-composite", ServiceDraft{Name: "Composite", Parameters: []ServiceParameterDraft{
		{ID: "P-size", Key: "size", Label: "Paper size", Type: ParameterChoice, Required: true, MaterialSource: &MaterialParameterSource{AllowedKinds: []MaterialKind{MaterialKindSheetStock}, ExposedAttributeKeys: []string{"width_mm", "height_mm"}}},
		{ID: "P-type", Key: "type", Label: "Paper type", Type: ParameterChoice, Required: true, MaterialSource: &MaterialParameterSource{AllowedKinds: []MaterialKind{MaterialKindSheetStock}, ExposedAttributeKey: "finish"}},
	}, MaterialVariants: []ServiceMaterialVariant{
		{ID: "VAR-a4-matte", MaterialID: a4Matte.ID, Values: map[string]string{"size": "210\x1f297", "type": "matte"}, Position: 0, Active: true},
		{ID: "VAR-a4-gloss", MaterialID: a4Gloss.ID, Values: map[string]string{"size": "210\x1f297", "type": "gloss"}, Position: 1, Active: true},
		{ID: "VAR-a3-matte", MaterialID: a3Matte.ID, Values: map[string]string{"size": "297\x1f420", "type": "matte"}, Position: 2, Active: true},
	}}, now)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ResolveMaterialSelection(service, []Material{a4Matte, a4Gloss, a3Matte}, map[string]string{"size": "210\x1f297", "type": "gloss"})
	if err != nil || got.ID != a4Gloss.ID {
		t.Fatalf("composite variant=%+v err=%v", got, err)
	}
	options, err := MaterialOptionsForParameter(service, "size", []Material{a4Matte, a4Gloss, a3Matte}, map[string]string{"type": "matte"})
	if err != nil || len(options) != 2 {
		t.Fatalf("composite options=%+v err=%v", options, err)
	}
}

func TestMaterialResolverExcludesArchivedAndReportsEmpty(t *testing.T) {
	m := testMaterial(t, "MAT-archived", MaterialKindSheetStock, MaterialAttributeValue{Key: "grammage_gsm", ValueType: MaterialAttributeInteger, IntegerValue: 170})
	m.Active = false
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service, err := NewService("SVC-empty", ServiceDraft{Name: "Empty", Parameters: []ServiceParameterDraft{{ID: "P", Key: "grammage", Label: "Grammage", Type: ParameterChoice, MaterialSource: &MaterialParameterSource{AllowedKinds: []MaterialKind{MaterialKindSheetStock}, ExposedAttributeKey: "grammage_gsm"}}}}, now)
	if err != nil {
		t.Fatal(err)
	}
	got, err := CompatibleMaterials(service, []Material{m}, map[string]string{"grammage": "170"})
	if err != nil || len(got) != 0 {
		t.Fatalf("archived compatible=%v err=%v", got, err)
	}
	if _, err := ResolveMaterialSelection(service, []Material{m}, map[string]string{"grammage": "170"}); err == nil {
		t.Fatal("empty material result was not reported")
	}
}

func TestSheetYieldRules(t *testing.T) {
	result, err := CalculateSheetYield(320*QuantityScale, 450*QuantityScale, 90*QuantityScale, 50*QuantityScale, 27*QuantityScale, false)
	if err != nil || result.ItemsPerSheet != 27 || result.SheetsRequired != QuantityScale {
		t.Fatalf("normal sheet yield=%+v err=%v", result, err)
	}
	result, err = CalculateSheetYield(120*QuantityScale, 100*QuantityScale, 70*QuantityScale, 30*QuantityScale, 5*QuantityScale, true)
	if err != nil || result.ItemsPerSheet != 4 || !result.Rotated || result.SheetsRequired != 2*QuantityScale {
		t.Fatalf("rotated sheet yield=%+v err=%v", result, err)
	}
	if _, err := CalculateSheetYield(100*QuantityScale, 100*QuantityScale, 101*QuantityScale, 20*QuantityScale, QuantityScale, true); err == nil {
		t.Fatal("oversized finished item fit a sheet")
	}
}

func TestRollConsumptionRules(t *testing.T) {
	result, err := CalculateRollConsumption(1000*QuantityScale, 700*QuantityScale, 400*QuantityScale, 2*QuantityScale, true, "meter")
	if err != nil || result.Across != 2 || result.Rows != 1 || result.ConsumedQuantity != 700000 || !result.Rotated {
		t.Fatalf("rotated roll layout=%+v err=%v", result, err)
	}
	if _, err := CalculateRollConsumption(500*QuantityScale, 600*QuantityScale, 700*QuantityScale, QuantityScale, true, "meter"); err == nil {
		t.Fatal("oversized roll item fit")
	}
}

func TestFinishedSizeAndMaterialConsumptionUseExplicitDimensions(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	material := testMaterial(t, "MAT-A4", MaterialKindSheetStock,
		MaterialAttributeValue{Key: "width_mm", ValueType: MaterialAttributeDecimal, DecimalValue: 210 * QuantityScale},
		MaterialAttributeValue{Key: "height_mm", ValueType: MaterialAttributeDecimal, DecimalValue: 297 * QuantityScale},
	)
	service, err := NewService("SVC-cards", ServiceDraft{
		Name: "Cards",
		Parameters: []ServiceParameterDraft{
			{ID: "P-quantity", Key: "quantity", Label: "Quantity", Type: ParameterInteger},
			{ID: "P-size", Key: "size", Label: "Finished size", Type: ParameterChoice, Options: []string{"card"}},
		},
		FinishedSize: &ServiceFinishedSizeDefinition{ParameterKey: "size", QuantityParameterKey: "quantity", AllowRotation: true, Options: []FinishedSizeOption{{ID: "SIZE-card", Code: "card", Label: "Business card", WidthMM: 90 * QuantityScale, HeightMM: 50 * QuantityScale, Active: true}}},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	consumed, err := CalculateMaterialConsumption(material, service, map[string]ResolvedParameter{
		"quantity": {Key: "quantity", Type: ParameterInteger, Value: "27", Quantity: 27 * QuantityScale},
		"size":     {Key: "size", Type: ParameterChoice, Value: "card"},
	})
	if err != nil || consumed != 3*QuantityScale {
		t.Fatalf("source sheet consumption=%v err=%v", consumed, err)
	}
}
