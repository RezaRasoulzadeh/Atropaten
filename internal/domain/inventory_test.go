package domain

import (
	"math"
	"testing"
)

func TestInventoryMathIsFixedScaleAndDeterministic(t *testing.T) {
	q := Quantity(1_250_000)
	got, err := ConvertPurchaseQuantity(q, Quantity(2_500_000))
	if err != nil || got != Quantity(3_125_000) {
		t.Fatalf("conversion = %d, err=%v", got, err)
	}
	line, err := MulQuantityRial(Quantity(1_500_000), 123)
	if err != nil || line != 185 {
		t.Fatalf("line total = %d, err=%v", line, err)
	}
	alloc, err := AllocateLandedCosts([]int64{100, 200, 300}, 0, 0, 0, 10)
	if err != nil || alloc[0]+alloc[1]+alloc[2] != 10 || alloc[0] != 2 || alloc[1] != 3 || alloc[2] != 5 {
		t.Fatalf("allocation = %#v, err=%v", alloc, err)
	}
	discount, err := AllocateLandedCosts([]int64{100, 200}, 30, 0, 0, 0)
	if err != nil || discount[0]+discount[1] != -30 {
		t.Fatalf("discount allocation = %#v, err=%v", discount, err)
	}
}

func TestWeightedAverageZeroAndMultiplePurchases(t *testing.T) {
	zero, err := WeightedAverage(InventorySummary{}, Quantity(2_000_000), 200)
	if err != nil || zero.PhysicalStock != Quantity(2_000_000) || zero.AverageUnitCostRial != 100 {
		t.Fatalf("zero stock average = %+v, err=%v", zero, err)
	}
	got, err := WeightedAverage(zero, Quantity(1_000_000), 150)
	if err != nil || got.AverageUnitCostRial != 117 || got.InventoryValueRial != 350 {
		t.Fatalf("weighted average = %+v, err=%v", got, err)
	}
}

func TestPurchaseConversionRejectsLossyFixedScaleResult(t *testing.T) {
	if _, err := ConvertPurchaseQuantity(1, 1); err == nil {
		t.Fatal("lossy conversion was silently truncated")
	}
}

func TestInventoryMathRejectsInt64IntermediateOverflow(t *testing.T) {
	if _, err := AllocateLandedCosts([]int64{1}, 0, math.MaxInt64, 1, 0); err == nil {
		t.Fatal("landed-cost intermediate overflow was accepted")
	}
	if _, err := WeightedAverage(InventorySummary{PhysicalStock: Quantity(math.MaxInt64), InventoryValueRial: 1}, 1, 1); err == nil {
		t.Fatal("quantity overflow was accepted")
	}
	if _, err := WeightedAverage(InventorySummary{PhysicalStock: 1, InventoryValueRial: math.MaxInt64}, 1, 1); err == nil {
		t.Fatal("inventory-value overflow was accepted")
	}
}
