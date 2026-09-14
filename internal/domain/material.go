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
	ErrMaterialNotFound        = errors.New("material not found")
	ErrMaterialDeleteProtected = errors.New("material has active inventory or operational dependencies; archive it instead")
)

// Quantity is a fixed-scale decimal quantity. Six fractional digits are
// stored as an integer so inventory quantities never pass through float math.
type Quantity int64

// MaterialKind is a stable behavioral classification. Display categories remain
// available for organization, but application decisions must use this code.
type MaterialKind string

const (
	MaterialKindSheetStock        MaterialKind = "sheet-stock"
	MaterialKindRollMedia         MaterialKind = "roll-media"
	MaterialKindBoard             MaterialKind = "board"
	MaterialKindInk               MaterialKind = "ink"
	MaterialKindLaminationFilm    MaterialKind = "lamination-film"
	MaterialKindAdhesive          MaterialKind = "adhesive"
	MaterialKindFabric            MaterialKind = "fabric"
	MaterialKindPackaging         MaterialKind = "packaging"
	MaterialKindChemical          MaterialKind = "chemical"
	MaterialKindGenericConsumable MaterialKind = "generic-consumable"
)

var materialKinds = []MaterialKind{
	MaterialKindSheetStock, MaterialKindRollMedia, MaterialKindBoard,
	MaterialKindInk, MaterialKindLaminationFilm, MaterialKindAdhesive,
	MaterialKindFabric, MaterialKindPackaging, MaterialKindChemical,
	MaterialKindGenericConsumable,
}

func ValidMaterialKinds() []MaterialKind { return append([]MaterialKind(nil), materialKinds...) }

func IsValidMaterialKind(value MaterialKind) bool {
	for _, kind := range materialKinds {
		if value == kind {
			return true
		}
	}
	return false
}

type MaterialAttributeValueType string

const (
	MaterialAttributeDecimal MaterialAttributeValueType = "decimal"
	MaterialAttributeInteger MaterialAttributeValueType = "integer"
	MaterialAttributeEnum    MaterialAttributeValueType = "enum"
	MaterialAttributeText    MaterialAttributeValueType = "text"
	MaterialAttributeBoolean MaterialAttributeValueType = "boolean"
)

type MaterialAttributeEnumOption struct {
	Code     string
	Label    string
	Active   bool
	Position int
}

// MaterialAttributeDefinition describes a typed, reusable specification. The
// definition is persisted independently from materials so new specifications do
// not require another materials-table column.
type MaterialAttributeDefinition struct {
	Key             string
	Label           string
	ValueType       MaterialAttributeValueType
	Unit            string
	ApplicableKinds []MaterialKind
	EnumOptions     []MaterialAttributeEnumOption
	Active          bool
	Position        int
}

type MaterialAttributeValue struct {
	Key          string
	ValueType    MaterialAttributeValueType
	DecimalValue Quantity
	IntegerValue int64
	EnumCode     string
	TextValue    string
	BooleanValue bool
}

func (v MaterialAttributeValue) Canonical() string {
	switch v.ValueType {
	case MaterialAttributeDecimal:
		return v.DecimalValue.String()
	case MaterialAttributeInteger:
		return strconv.FormatInt(v.IntegerValue, 10)
	case MaterialAttributeEnum:
		return strings.TrimSpace(v.EnumCode)
	case MaterialAttributeBoolean:
		if v.BooleanValue {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(v.TextValue)
	}
}

func (v MaterialAttributeValue) Validate() error {
	if strings.TrimSpace(v.Key) == "" {
		return validationError("attribute.key", "is required")
	}
	switch v.ValueType {
	case MaterialAttributeDecimal, MaterialAttributeInteger, MaterialAttributeText, MaterialAttributeBoolean:
	case MaterialAttributeEnum:
		if strings.TrimSpace(v.EnumCode) == "" {
			return validationError("attribute.enumCode", "is required")
		}
	default:
		return validationError("attribute.valueType", "is not supported")
	}
	return nil
}

type MaterialAttributeFilter struct {
	Key   string
	Value MaterialAttributeValue
}

// MaterialParameterSource makes a service parameter explicitly material-backed.
// ExposedAttributeKey derives customer options from actual material attributes;
// SelectMaterial exposes explicit material IDs instead.
type MaterialParameterSource struct {
	AllowedKinds        []MaterialKind
	ExposedAttributeKey string
	AllowedValues       []MaterialAttributeValue
	SelectMaterial      bool
	AdditionalFilters   []MaterialAttributeFilter
}

type Material struct {
	ID                          string
	Name                        string
	SKU                         string
	Category                    string
	Kind                        MaterialKind
	Attributes                  []MaterialAttributeValue
	PurchaseUnit                string
	ConsumptionUnit             string
	ConversionFactor            Quantity
	PhysicalStock               Quantity
	ReservedStock               Quantity
	AvailableStock              Quantity
	ReorderLevel                Quantity
	AverageUnitCostRial         int64
	HighestPurchaseUnitCostRial int64
	PreferredSupplier           string
	Notes                       string
	Active                      bool
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}

type MaterialDraft struct {
	Name                string
	SKU                 string
	Category            string
	Kind                MaterialKind
	Attributes          []MaterialAttributeValue
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
	kind := draft.Kind
	if kind == "" {
		kind = MaterialKindGenericConsumable
	}
	material := Material{
		ID:                  strings.TrimSpace(id),
		Name:                strings.TrimSpace(draft.Name),
		SKU:                 strings.TrimSpace(draft.SKU),
		Category:            strings.TrimSpace(draft.Category),
		Kind:                kind,
		Attributes:          append([]MaterialAttributeValue(nil), draft.Attributes...),
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
	if m.Kind == "" {
		return validationError("kind", "is required")
	}
	if !IsValidMaterialKind(m.Kind) {
		return validationError("kind", "must be a supported material kind")
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
	seenAttributes := make(map[string]struct{}, len(m.Attributes))
	for index, attribute := range m.Attributes {
		if err := attribute.Validate(); err != nil {
			return validationError(fmt.Sprintf("attributes[%d]", index), err.Error())
		}
		if attribute.Key == "width_mm" || attribute.Key == "height_mm" || attribute.Key == "length_mm" {
			if attribute.ValueType != MaterialAttributeDecimal || attribute.DecimalValue <= 0 {
				return validationError(fmt.Sprintf("attributes[%d]", index), "physical dimensions must be positive decimal millimetres")
			}
		}
		if attribute.Key == "grammage_gsm" || attribute.Key == "thickness_micron" {
			if attribute.ValueType != MaterialAttributeInteger || attribute.IntegerValue <= 0 {
				return validationError(fmt.Sprintf("attributes[%d]", index), "specification must be a positive integer")
			}
		}
		if _, exists := seenAttributes[attribute.Key]; exists {
			return validationError("attributes", "must not contain duplicate keys")
		}
		seenAttributes[attribute.Key] = struct{}{}
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
