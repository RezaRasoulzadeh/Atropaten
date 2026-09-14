package domain

import (
	"fmt"
	"math/big"
)

type MaterialConsumptionStrategy string

const (
	ConsumptionStrategySheet MaterialConsumptionStrategy = "sheet-yield"
	ConsumptionStrategyRoll  MaterialConsumptionStrategy = "roll-layout"
)

func ConsumptionStrategyForMaterialKind(kind MaterialKind) (MaterialConsumptionStrategy, bool) {
	switch kind {
	case MaterialKindSheetStock, MaterialKindBoard:
		return ConsumptionStrategySheet, true
	case MaterialKindRollMedia, MaterialKindFabric:
		return ConsumptionStrategyRoll, true
	default:
		return "", false
	}
}

func (m Material) Attribute(key string) (MaterialAttributeValue, bool) {
	for _, value := range m.Attributes {
		if value.Key == key {
			return value, true
		}
	}
	return MaterialAttributeValue{}, false
}

func (m Material) PhysicalDimensions() (widthMM, heightMM Quantity, ok bool) {
	width, widthOK := m.Attribute("width_mm")
	height, heightOK := m.Attribute("height_mm")
	if !widthOK || width.ValueType != MaterialAttributeDecimal || width.DecimalValue <= 0 {
		return 0, 0, false
	}
	if !heightOK || height.ValueType != MaterialAttributeDecimal || height.DecimalValue <= 0 {
		return 0, 0, false
	}
	return width.DecimalValue, height.DecimalValue, true
}

func (m Material) RollWidthMM() (Quantity, bool) {
	width, ok := m.Attribute("width_mm")
	return width.DecimalValue, ok && width.ValueType == MaterialAttributeDecimal && width.DecimalValue > 0
}

func CalculateMaterialConsumption(material Material, service Service, parameters map[string]ResolvedParameter) (Quantity, error) {
	if service.FinishedSize == nil {
		return 0, fmt.Errorf("service does not define finished dimensions")
	}
	quantityParameter, ok := parameters[service.FinishedSize.QuantityParameterKey]
	if !ok || quantityParameter.Quantity <= 0 {
		return 0, fmt.Errorf("finished quantity parameter %q is not resolved", service.FinishedSize.QuantityParameterKey)
	}
	selected := make(map[string]string, len(parameters))
	for key, value := range parameters {
		selected[key] = value.Value
	}
	width, height, err := service.ResolveFinishedDimensions(selected)
	if err != nil {
		return 0, err
	}
	strategy, ok := ConsumptionStrategyForMaterialKind(material.Kind)
	if !ok {
		return 0, fmt.Errorf("material kind %q has no consumption strategy", material.Kind)
	}
	switch strategy {
	case ConsumptionStrategySheet:
		sheetWidth, sheetHeight, ok := material.PhysicalDimensions()
		if !ok {
			return 0, fmt.Errorf("material %q has incomplete sheet dimensions", material.ID)
		}
		yield, err := CalculateSheetYield(sheetWidth, sheetHeight, width, height, quantityParameter.Quantity, service.FinishedSize.AllowRotation)
		if err != nil {
			return 0, err
		}
		return yield.SheetsRequired, nil
	case ConsumptionStrategyRoll:
		rollWidth, ok := material.RollWidthMM()
		if !ok {
			return 0, fmt.Errorf("material %q has incomplete roll width", material.ID)
		}
		layout, err := CalculateRollConsumption(rollWidth, width, height, quantityParameter.Quantity, service.FinishedSize.AllowRotation, material.ConsumptionUnit)
		if err != nil {
			return 0, err
		}
		return layout.ConsumedQuantity, nil
	default:
		return 0, fmt.Errorf("unsupported material consumption strategy %q", strategy)
	}
}

type SheetYield struct {
	ItemsPerSheet  int64
	SheetsRequired Quantity
	Rotated        bool
}

func CalculateSheetYield(sheetWidthMM, sheetHeightMM, itemWidthMM, itemHeightMM, requiredQuantity Quantity, allowRotation bool) (SheetYield, error) {
	if sheetWidthMM <= 0 || sheetHeightMM <= 0 || itemWidthMM <= 0 || itemHeightMM <= 0 {
		return SheetYield{}, fmt.Errorf("sheet and finished dimensions must be positive")
	}
	if requiredQuantity <= 0 {
		return SheetYield{}, fmt.Errorf("required finished quantity must be positive")
	}
	widthCount, heightCount := int64(sheetWidthMM/itemWidthMM), int64(sheetHeightMM/itemHeightMM)
	if heightCount > 0 && widthCount > int64(^uint64(0)>>1)/heightCount {
		return SheetYield{}, fmt.Errorf("sheet yield is too large")
	}
	normal := widthCount * heightCount
	best := normal
	rotated := false
	if allowRotation {
		rotatedYield := int64(sheetWidthMM/itemHeightMM) * int64(sheetHeightMM/itemWidthMM)
		if rotatedYield > best {
			best, rotated = rotatedYield, true
		}
	}
	if best <= 0 {
		return SheetYield{}, fmt.Errorf("finished dimensions do not fit on the selected sheet")
	}
	denominator := new(big.Int).Mul(big.NewInt(best), big.NewInt(QuantityScale))
	numerator := new(big.Int).Add(big.NewInt(int64(requiredQuantity)), new(big.Int).Sub(denominator, big.NewInt(1)))
	numerator.Quo(numerator, denominator)
	numerator.Mul(numerator, big.NewInt(QuantityScale))
	if !numerator.IsInt64() {
		return SheetYield{}, fmt.Errorf("required sheet quantity is too large")
	}
	return SheetYield{ItemsPerSheet: best, SheetsRequired: Quantity(numerator.Int64()), Rotated: rotated}, nil
}

type RollLayout struct {
	Across           int64
	Rows             int64
	ConsumedLengthMM Quantity
	ConsumedQuantity Quantity
	Rotated          bool
}

// CalculateRollConsumption lays rectangular pieces along a continuous roll.
// ConsumedQuantity is expressed in the material's existing consumption unit.
func CalculateRollConsumption(rollWidthMM, itemWidthMM, itemHeightMM, requiredQuantity Quantity, allowRotation bool, consumptionUnit string) (RollLayout, error) {
	if rollWidthMM <= 0 || itemWidthMM <= 0 || itemHeightMM <= 0 || requiredQuantity <= 0 {
		return RollLayout{}, fmt.Errorf("roll, finished dimensions, and quantity must be positive")
	}
	best, err := rollOrientation(rollWidthMM, itemWidthMM, itemHeightMM, requiredQuantity, consumptionUnit, false)
	if err != nil {
		best = RollLayout{}
	}
	if allowRotation && itemWidthMM != itemHeightMM {
		rotated, rotatedErr := rollOrientation(rollWidthMM, itemHeightMM, itemWidthMM, requiredQuantity, consumptionUnit, true)
		if rotatedErr == nil && (best.Across == 0 || rotated.ConsumedLengthMM < best.ConsumedLengthMM) {
			best = rotated
		}
	}
	if best.Across == 0 {
		return RollLayout{}, fmt.Errorf("finished dimensions do not fit within the selected roll width")
	}
	return best, nil
}

func rollOrientation(rollWidthMM, itemWidthMM, itemHeightMM, requiredQuantity Quantity, consumptionUnit string, rotated bool) (RollLayout, error) {
	across := int64(rollWidthMM / itemWidthMM)
	if across <= 0 {
		return RollLayout{}, fmt.Errorf("item width does not fit roll width")
	}
	rows := (int64(requiredQuantity) + across*QuantityScale - 1) / (across * QuantityScale)
	length := new(big.Int).Mul(big.NewInt(rows), big.NewInt(int64(itemHeightMM)))
	if !length.IsInt64() {
		return RollLayout{}, fmt.Errorf("roll length is too large")
	}
	lengthMM := Quantity(length.Int64())
	consumed, err := rollQuantity(lengthMM, rollWidthMM, consumptionUnit)
	if err != nil {
		return RollLayout{}, err
	}
	return RollLayout{Across: across, Rows: rows, ConsumedLengthMM: lengthMM, ConsumedQuantity: consumed, Rotated: rotated}, nil
}

func rollQuantity(lengthMM, widthMM Quantity, unit string) (Quantity, error) {
	switch NormalizeUnit(unit) {
	case "meter":
		return Quantity(int64(lengthMM) / 1000), nil
	case "square meter":
		n := new(big.Int).Mul(big.NewInt(int64(lengthMM)), big.NewInt(int64(widthMM)))
		n.Quo(n, big.NewInt(QuantityScale*1_000_000))
		if !n.IsInt64() {
			return 0, fmt.Errorf("roll area is too large")
		}
		return Quantity(n.Int64()), nil
	default:
		return 0, fmt.Errorf("roll material consumption unit %q must be meter or square meter", unit)
	}
}
