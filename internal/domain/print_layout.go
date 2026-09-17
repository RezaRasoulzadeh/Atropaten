package domain

import (
	"fmt"
	"math"
	"math/big"
)

// PrintLayout is persisted with the price so production can use the same plan.
type PrintLayout struct {
	FinishedWidthMM  string  `json:"finishedWidthMM"`
	FinishedHeightMM string  `json:"finishedHeightMM"`
	GapMM            string  `json:"gapMM"`
	MarginMM         string  `json:"marginMM"`
	MaterialID       string  `json:"materialId"`
	MaterialName     string  `json:"materialName"`
	Kind             string  `json:"kind"`
	Quantity         string  `json:"quantity"`
	ConsumedQuantity string  `json:"consumedQuantity"`
	Unit             string  `json:"unit"`
	Across           int64   `json:"across"`
	Rows             int64   `json:"rows"`
	ItemsPerSheet    int64   `json:"itemsPerSheet"`
	Sheets           string  `json:"sheets"`
	LengthMM         string  `json:"lengthMM"`
	Rotated          bool    `json:"rotated"`
	WastePercent     float64 `json:"wastePercent"`
	WasteCostRial    int64   `json:"wasteCostRial"`
	AreaM2           string  `json:"areaM2"`
	OriginalLengthMM string  `json:"originalLengthMM,omitempty"`
}

func CalculatePrintLayout(material Material, service Service, parameters map[string]ResolvedParameter) (PrintLayout, error) {
	result := PrintLayout{MaterialID: material.ID, MaterialName: material.Name, Kind: string(material.Kind), Unit: material.ConsumptionUnit}
	if service.FinishedSize == nil {
		return result, fmt.Errorf("service does not define finished dimensions")
	}
	qty := parameters[service.FinishedSize.QuantityParameterKey].Quantity
	if qty <= 0 || qty%QuantityScale != 0 {
		return result, fmt.Errorf("finished quantity must be a positive whole number")
	}
	values := map[string]string{}
	for key, p := range parameters {
		values[key] = p.Value
	}
	width, height, err := service.ResolveFinishedDimensions(values)
	if err != nil {
		return result, err
	}
	// Gap is the cut/bleed allowance between pieces; margin is reserved on each edge.
	gap, margin := parameters["layout_gap_mm"].Quantity, parameters["layout_margin_mm"].Quantity
	if gap < 0 || margin < 0 || gap > 1000*QuantityScale || margin > 1000*QuantityScale {
		return result, fmt.Errorf("layout gap and margin must be between 0 and 1000 mm")
	}
	if width > Quantity(math.MaxInt64)-gap || height > Quantity(math.MaxInt64)-gap {
		return result, fmt.Errorf("finished dimensions are too large")
	}
	result.Quantity = qty.String()
	result.FinishedWidthMM, result.FinishedHeightMM, result.GapMM, result.MarginMM = width.String(), height.String(), gap.String(), margin.String()
	var area float64
	var areaQuantity Quantity
	switch material.Kind {
	case MaterialKindSheetStock, MaterialKindBoard:
		sw, sh, ok := material.PhysicalDimensions()
		if !ok {
			return result, fmt.Errorf("material needs physical sheet width and height")
		}
		if sw <= 2*margin || sh <= 2*margin {
			return result, fmt.Errorf("margins leave no printable sheet area")
		}
		if sw > Quantity(math.MaxInt64)-gap || sh > Quantity(math.MaxInt64)-gap {
			return result, fmt.Errorf("sheet dimensions are too large")
		}
		layout, e := CalculateSheetYield(sw-2*margin+gap, sh-2*margin+gap, width+gap, height+gap, qty, service.FinishedSize.AllowRotation)
		if e != nil {
			return result, e
		}
		if material.ConsumptionUnit != "sheet" && material.ConsumptionUnit != "piece" {
			return result, fmt.Errorf("sheet layout requires sheet or piece consumption units")
		}
		result.ItemsPerSheet, result.Sheets, result.Rotated = layout.ItemsPerSheet, layout.SheetsRequired.String(), layout.Rotated
		result.ConsumedQuantity = result.Sheets
		areaInt := new(big.Int).Mul(big.NewInt(int64(sw)), big.NewInt(int64(sh)))
		areaInt.Mul(areaInt, big.NewInt(int64(layout.SheetsRequired/QuantityScale)))
		denominator := big.NewInt(QuantityScale * 1_000_000)
		areaInt.Add(areaInt, new(big.Int).Sub(denominator, big.NewInt(1)))
		areaInt.Quo(areaInt, denominator)
		if !areaInt.IsInt64() {
			return result, fmt.Errorf("sheet area is too large")
		}
		areaQuantity = Quantity(areaInt.Int64())
		area = float64(sw) / float64(QuantityScale) * float64(sh) / float64(QuantityScale) * float64(layout.SheetsRequired/QuantityScale)
	case MaterialKindRollMedia, MaterialKindFabric:
		rw, ok := material.RollWidthMM()
		if !ok {
			return result, fmt.Errorf("material needs physical roll width")
		}
		if width > rw {
			return result, fmt.Errorf("finished width cannot exceed the selected material width")
		}
		if rw <= 2*margin || rw > Quantity(math.MaxInt64)-gap {
			return result, fmt.Errorf("invalid roll width or margins")
		}
		// Select using usable width, then charge the full stock width including waste.
		layout, e := CalculateRollConsumption(rw-2*margin+gap, width+gap, height+gap, qty, service.FinishedSize.AllowRotation, "meter")
		if e != nil {
			return result, e
		}
		if layout.ConsumedLengthMM > Quantity(math.MaxInt64)-2*margin {
			return result, fmt.Errorf("roll length is too large")
		}
		length := layout.ConsumedLengthMM - gap + 2*margin
		consumed, e := rollQuantity(length, rw, material.ConsumptionUnit)
		if e != nil {
			return result, e
		}
		result.Across, result.Rows, result.Rotated = layout.Across, layout.Rows, layout.Rotated
		result.LengthMM, result.ConsumedQuantity = length.String(), consumed.String()
		areaQuantity, e = rollQuantity(length, rw, "square meter")
		if e != nil {
			return result, e
		}
		original, e := CalculateRollConsumption(rw-2*margin+gap, width+gap, height+gap, qty, false, "meter")
		if e == nil && original.ConsumedLengthMM <= Quantity(math.MaxInt64)-2*margin {
			result.OriginalLengthMM = (original.ConsumedLengthMM - gap + 2*margin).String()
		}
		area = float64(rw) / float64(QuantityScale) * float64(length) / float64(QuantityScale)
	default:
		return result, fmt.Errorf("material kind %q does not support rectangular layouts", material.Kind)
	}
	finishedArea := float64(width) / float64(QuantityScale) * float64(height) / float64(QuantityScale) * float64(qty/QuantityScale)
	result.WastePercent = math.Max(0, 100*(1-finishedArea/area))
	result.AreaM2 = areaQuantity.String()
	return result, nil
}

// CalculateMaterialWasteCost allocates the consumed material cost to the
// unused area in a layout. The returned amount is informational: the full
// consumed material cost already includes this waste.
func CalculateMaterialWasteCost(layout PrintLayout, materialCostRial int64) (int64, error) {
	if materialCostRial < 0 {
		return 0, fmt.Errorf("material cost cannot be negative")
	}
	consumedArea, err := ParseQuantity(layout.AreaM2)
	if err != nil || consumedArea <= 0 {
		return 0, fmt.Errorf("layout material area must be positive")
	}
	width, err := ParseQuantity(layout.FinishedWidthMM)
	if err != nil || width <= 0 {
		return 0, fmt.Errorf("layout finished width must be positive")
	}
	height, err := ParseQuantity(layout.FinishedHeightMM)
	if err != nil || height <= 0 {
		return 0, fmt.Errorf("layout finished height must be positive")
	}
	quantity, err := ParseQuantity(layout.Quantity)
	if err != nil || quantity <= 0 {
		return 0, fmt.Errorf("layout quantity must be positive")
	}
	finishedArea := new(big.Int).Mul(big.NewInt(int64(width)), big.NewInt(int64(height)))
	finishedArea.Mul(finishedArea, big.NewInt(int64(quantity)))
	finishedArea.Quo(finishedArea, big.NewInt(QuantityScale*QuantityScale*1_000_000))
	consumed := big.NewInt(int64(consumedArea))
	if finishedArea.Cmp(consumed) >= 0 {
		return 0, nil
	}
	wastedArea := new(big.Int).Sub(consumed, finishedArea)
	numerator := new(big.Int).Mul(big.NewInt(materialCostRial), wastedArea)
	numerator.Add(numerator, new(big.Int).Sub(consumed, big.NewInt(1)))
	numerator.Quo(numerator, consumed)
	if !numerator.IsInt64() {
		return 0, fmt.Errorf("material waste cost exceeds Rial range")
	}
	return numerator.Int64(), nil
}
