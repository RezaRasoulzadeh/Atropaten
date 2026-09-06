package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var (
	ErrSupplierNotFound        = errors.New("supplier not found")
	ErrPurchaseNotFound        = errors.New("purchase not found")
	ErrMovementNotFound        = errors.New("inventory movement not found")
	ErrSupplierDeleteProtected = errors.New("supplier has purchase history; archive it instead")
	ErrPurchaseNotDraft        = errors.New("only draft purchases can be edited or deleted")
	ErrPurchaseAlreadyPosted   = errors.New("purchase is already posted")
	ErrPurchaseCannotCancel    = errors.New("only posted purchases can be cancelled")
	ErrInsufficientStock       = errors.New("operation would make stock negative")
)

const (
	PurchaseDraft     = "Draft"
	PurchasePosted    = "Posted"
	PurchaseCancelled = "Cancelled"
)

type Supplier struct {
	ID, Name, Code, Phone, Email, Address, Notes string
	Active                                       bool
	CreatedAt, UpdatedAt                         time.Time
}

func (s Supplier) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return validationError("id", "is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		return validationError("name", "is required")
	}
	if s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	return nil
}

type Purchase struct {
	ID, PurchaseNumber, SupplierID, SupplierNameSnapshot, SupplierCodeSnapshot        string
	SupplierInvoiceNumber, Notes                                                      string
	PurchaseDate                                                                      time.Time
	Status                                                                            string
	SubtotalRial, DiscountRial, ShippingRial, TaxRial, AdditionalCostsRial, TotalRial int64
	CreatedAt, UpdatedAt                                                              time.Time
	Items                                                                             []PurchaseItem
}

type PurchaseItem struct {
	ID, PurchaseID                                                                          string
	Position                                                                                int
	MaterialID, MaterialNameSnapshot, PurchaseUnitSnapshot, ConsumptionUnitSnapshot         string
	PurchaseQuantity, ConversionFactorSnapshot, ConsumptionQuantity                         Quantity
	UnitAcquisitionCostRial, AllocatedAdditionalCostRial, LandedUnitCostRial, LineTotalRial int64
	Notes                                                                                   string
}

type InventoryMovement struct {
	ID, MaterialID, MovementType     string
	OccurredAt                       time.Time
	QuantityDelta                    Quantity
	UnitCostRial, TotalCostRial      int64
	ReferenceType, ReferenceID, Note string
	CreatedAt                        time.Time
}

type InventorySummary struct {
	PhysicalStock       Quantity
	AverageUnitCostRial int64
	InventoryValueRial  int64
}

// MulQuantityRial multiplies a fixed-scale quantity by an integer Rial amount.
// It rounds half up at the fixed-scale boundary, using only integers.
func MulQuantityRial(q Quantity, rial int64) (int64, error) {
	if q < 0 || rial < 0 {
		return 0, fmt.Errorf("quantity and Rial must be non-negative")
	}
	n := new(big.Int).Mul(big.NewInt(int64(q)), big.NewInt(rial))
	n.Add(n, big.NewInt(QuantityScale/2))
	n.Quo(n, big.NewInt(QuantityScale))
	if !n.IsInt64() {
		return 0, fmt.Errorf("Rial result is too large")
	}
	return n.Int64(), nil
}

// ConvertPurchaseQuantity converts purchase units to canonical consumption units.
// The result is exact at the existing six-decimal quantity scale.
func ConvertPurchaseQuantity(qty, factor Quantity) (Quantity, error) {
	if qty < 0 || factor <= 0 {
		return 0, fmt.Errorf("quantity must be non-negative and conversion factor positive")
	}
	n := new(big.Int).Mul(big.NewInt(int64(qty)), big.NewInt(int64(factor)))
	rem := new(big.Int)
	n.QuoRem(n, big.NewInt(QuantityScale), rem)
	if rem.Sign() != 0 {
		return 0, fmt.Errorf("converted quantity cannot be represented at six decimal places")
	}
	if !n.IsInt64() {
		return 0, fmt.Errorf("converted quantity is too large")
	}
	return Quantity(n.Int64()), nil
}

// AllocateLandedCosts allocates discount, shipping, tax and additional costs
// proportionally by acquisition value. Tax is included in inventory valuation.
// Any fixed-scale rounding remainder goes to the first non-zero line.
func AllocateLandedCosts(lineTotals []int64, discount, shipping, tax, additional int64) ([]int64, error) {
	if discount < 0 || shipping < 0 || tax < 0 || additional < 0 {
		return nil, fmt.Errorf("purchase costs cannot be negative")
	}
	var subtotal int64
	for _, v := range lineTotals {
		if v < 0 || (v > 0 && subtotal > int64(^uint64(0)>>1)-v) {
			return nil, fmt.Errorf("line total is invalid")
		}
		subtotal += v
	}
	if discount > subtotal {
		return nil, fmt.Errorf("discount cannot exceed subtotal")
	}
	landedExtra := shipping + tax + additional - discount
	alloc := make([]int64, len(lineTotals))
	if landedExtra == 0 || subtotal == 0 {
		return alloc, nil
	}
	first := -1
	var assigned int64
	for i, line := range lineTotals {
		if line > 0 && first < 0 {
			first = i
		}
		n := new(big.Int).Mul(big.NewInt(line), big.NewInt(landedExtra))
		n.Quo(n, big.NewInt(subtotal))
		if !n.IsInt64() {
			return nil, fmt.Errorf("allocation is too large")
		}
		alloc[i] = n.Int64()
		assigned += alloc[i]
	}
	if first >= 0 {
		alloc[first] += landedExtra - assigned
	}
	return alloc, nil
}

func WeightedAverage(existing InventorySummary, incomingQty Quantity, incomingValue int64) (InventorySummary, error) {
	if existing.PhysicalStock < 0 || incomingQty < 0 || incomingValue < 0 {
		return InventorySummary{}, fmt.Errorf("inventory values cannot be negative")
	}
	if incomingQty == 0 {
		return existing, nil
	}
	newQty := int64(existing.PhysicalStock) + int64(incomingQty)
	if newQty < 0 {
		return InventorySummary{}, fmt.Errorf("quantity overflow")
	}
	oldValue := existing.InventoryValueRial
	if oldValue < 0 {
		return InventorySummary{}, fmt.Errorf("inventory value cannot be negative")
	}
	total := new(big.Int).Add(big.NewInt(oldValue), big.NewInt(incomingValue))
	avg := new(big.Int).Mul(total, big.NewInt(QuantityScale))
	avg.Add(avg, big.NewInt(int64(newQty)/2))
	avg.Quo(avg, big.NewInt(newQty))
	value := total.Int64()
	if !avg.IsInt64() {
		return InventorySummary{}, fmt.Errorf("average cost is too large")
	}
	return InventorySummary{PhysicalStock: Quantity(newQty), AverageUnitCostRial: avg.Int64(), InventoryValueRial: value}, nil
}
