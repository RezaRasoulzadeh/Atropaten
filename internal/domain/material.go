package domain

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// QuantityScale is intentionally an untyped constant so it can participate
// in arithmetic with both the Quantity domain type and int64 parser values.
const QuantityScale = 1_000_000

var (
	ErrMaterialNotFound = errors.New("material not found")
)

// Quantity is a fixed-scale decimal quantity. Six fractional digits are
// stored as an integer so inventory quantities never pass through float math.
type Quantity int64

type Material struct {
	ID                  string
	Name                string
	SKU                 string
	Category            string
	PurchaseUnit        string
	ConsumptionUnit     string
	ConversionFactor    Quantity
	PhysicalStock       Quantity
	ReservedStock       Quantity
	AvailableStock      Quantity
	ReorderLevel        Quantity
	AverageUnitCostRial int64
	PreferredSupplier   string
	Notes               string
	Active              bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type MaterialDraft struct {
	Name                string
	SKU                 string
	Category            string
	PurchaseUnit        string
	ConsumptionUnit     string
	ConversionFactor    Quantity
	PhysicalStock       Quantity
	ReorderLevel        Quantity
	AverageUnitCostRial int64
	PreferredSupplier   string
	Notes               string
}

var validUnits = []string{
	"piece", "sheet", "pack", "kilogram", "gram", "roll", "meter", "liter", "milliliter", "square meter",
}

func ValidUnits() []string {
	return append([]string(nil), validUnits...)
}

func NormalizeUnit(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func NewMaterial(id string, draft MaterialDraft, now time.Time) (Material, error) {
	material := Material{
		ID:                  strings.TrimSpace(id),
		Name:                strings.TrimSpace(draft.Name),
		SKU:                 strings.TrimSpace(draft.SKU),
		Category:            strings.TrimSpace(draft.Category),
		PurchaseUnit:        NormalizeUnit(draft.PurchaseUnit),
		ConsumptionUnit:     NormalizeUnit(draft.ConsumptionUnit),
		ConversionFactor:    draft.ConversionFactor,
		PhysicalStock:       draft.PhysicalStock,
		ReorderLevel:        draft.ReorderLevel,
		AverageUnitCostRial: draft.AverageUnitCostRial,
		PreferredSupplier:   strings.TrimSpace(draft.PreferredSupplier),
		Notes:               strings.TrimSpace(draft.Notes),
		Active:              true,
		CreatedAt:           now.UTC(),
		UpdatedAt:           now.UTC(),
	}
	if err := material.Validate(); err != nil {
		return Material{}, err
	}
	return material, nil
}

func (m *Material) Update(draft MaterialDraft, now time.Time) error {
	updated, err := NewMaterial(m.ID, draft, m.CreatedAt)
	if err != nil {
		return err
	}
	updated.Active = m.Active
	updated.UpdatedAt = now.UTC()
	*m = updated
	return nil
}

func (m Material) Validate() error {
	if m.ID == "" {
		return validationError("id", "is required")
	}
	if m.Name == "" {
		return validationError("name", "is required")
	}
	if !isValidUnit(m.PurchaseUnit) {
		return validationError("purchaseUnit", "must be a supported unit")
	}
	if !isValidUnit(m.ConsumptionUnit) {
		return validationError("consumptionUnit", "must be a supported unit")
	}
	if m.ConversionFactor <= 0 {
		return validationError("conversionFactor", "must be greater than zero")
	}
	if m.PhysicalStock < 0 {
		return validationError("physicalStock", "cannot be negative")
	}
	if m.ReorderLevel < 0 {
		return validationError("reorderLevel", "cannot be negative")
	}
	if m.AverageUnitCostRial < 0 {
		return validationError("averageUnitCostRial", "cannot be negative")
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	return nil
}

func (m Material) LowStock() bool {
	if m.ReservedStock == 0 && m.AvailableStock == 0 && m.PhysicalStock > 0 {
		return m.PhysicalStock <= m.ReorderLevel
	}
	return m.AvailableStock <= m.ReorderLevel
}

func isValidUnit(value string) bool {
	for _, unit := range validUnits {
		if value == unit {
			return true
		}
	}
	return false
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func validationError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}

func ParseQuantity(value string) (Quantity, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("quantity is required")
	}
	if strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") {
		return 0, fmt.Errorf("quantity must be non-negative")
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("quantity must be a decimal number")
	}
	if len(parts) == 2 && len(parts[1]) > 6 {
		return 0, fmt.Errorf("quantity supports at most six decimal places")
	}
	if _, err := strconv.ParseUint(parts[0], 10, 64); err != nil {
		return 0, fmt.Errorf("quantity must be a decimal number")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole > math.MaxInt64/QuantityScale {
		return 0, fmt.Errorf("quantity is too large")
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	for len(fraction) < 6 {
		fraction += "0"
	}
	frac := int64(0)
	if fraction != "" {
		frac, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("quantity must be a decimal number")
		}
	}
	if whole == math.MaxInt64/QuantityScale && frac > math.MaxInt64%QuantityScale {
		return 0, fmt.Errorf("quantity is too large")
	}
	return Quantity(whole*QuantityScale + frac), nil
}

func (q Quantity) String() string {
	if q == 0 {
		return "0"
	}
	whole := q / QuantityScale
	fraction := q % QuantityScale
	wholeValue := int64(whole)
	if fraction == 0 {
		return strconv.FormatInt(wholeValue, 10)
	}
	text := fmt.Sprintf("%06d", int64(fraction))
	text = strings.TrimRight(text, "0")
	return strconv.FormatInt(wholeValue, 10) + "." + text
}
