package domain

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrServiceNotFound        = errors.New("service not found")
	ErrServiceDeleteProtected = errors.New("service has order, invoice, or service dependency history; archive it instead")
	ErrParameterNotFound      = errors.New("service parameter not found")
	ErrCostComponentNotFound  = errors.New("service cost component not found")
)

type ParameterType string

const (
	ParameterInteger           ParameterType = "integer"
	ParameterDecimal           ParameterType = "decimal"
	ParameterBoolean           ParameterType = "boolean"
	ParameterChoice            ParameterType = "choice"
	ParameterMaterialReference ParameterType = "material-reference"
	ParameterMachineReference  ParameterType = "machine-reference"
)

var parameterKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type Service struct {
	ID              string
	Name            string
	Code            string
	Category        string
	Description     string
	ImagePath       string
	DefaultUnit     string
	DefaultPriority Priority
	Active          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Parameters      []ServiceParameter
	Components      []ServiceCostComponent
	PricingRule     *ServicePricingRule
	FinishedSize    *ServiceFinishedSizeDefinition
	// MaterialVariants contains explicit service-time mappings from a complete
	// set of material-backed option values to one physical inventory material.
	// It is optional so legacy services continue to resolve through their
	// existing MaterialSource constraints.
	MaterialVariants []ServiceMaterialVariant
}

type ServiceDraft struct {
	Name             string
	Code             string
	Category         string
	Description      string
	ImagePath        string
	DefaultUnit      string
	DefaultPriority  Priority
	Parameters       []ServiceParameterDraft
	Components       []ServiceCostComponentDraft
	PricingRule      *ServicePricingRuleDraft
	FinishedSize     *ServiceFinishedSizeDefinition
	MaterialVariants []ServiceMaterialVariant
}

// ServiceMaterialVariant is the immutable selection rule used by new orders.
// Values are keyed by service parameter key and contain canonical option
// values (not display labels). A variant must resolve to exactly one material.
type ServiceMaterialVariant struct {
	ID               string
	MaterialID       string
	Values           map[string]string
	SellingPriceRial int64
	Position         int
	Active           bool
}

type FinishedSizeOption struct {
	ID       string
	Code     string
	Label    string
	WidthMM  Quantity
	HeightMM Quantity
	Position int
	Active   bool
}

type ServiceFinishedSizeDefinition struct {
	ParameterKey         string
	QuantityParameterKey string
	WidthParameterKey    string
	HeightParameterKey   string
	AllowCustom          bool
	AllowRotation        bool
	Options              []FinishedSizeOption
}

type CostComponentType string

const (
	// Percentage components are definitions only in M1-003. M1-004 applies
	// overhead and waste to the enabled subtotal accumulated before them.
	CostMaterial   CostComponentType = "material"
	CostMachine    CostComponentType = "machine"
	CostService    CostComponentType = "service"
	CostLabor      CostComponentType = "labor"
	CostOutsourced CostComponentType = "outsourced"
	CostFixed      CostComponentType = "fixed"
	CostOverhead   CostComponentType = "overhead"
	CostWaste      CostComponentType = "waste"
	CostManual     CostComponentType = "manual"
)

type UsageMode string

const (
	UsageFixed     UsageMode = "fixed"
	UsageParameter UsageMode = "parameter"
)

type ServiceCostComponent struct {
	ID               string
	ServiceID        string
	Name             string
	Type             CostComponentType
	ReferenceID      string
	UsageMode        UsageMode
	ParameterKey     string
	RateID           string
	RateParameterKey string
	UsageQuantity    Quantity
	Multiplier       Quantity
	RateRial         int64
	Percentage       Quantity
	RateBasis        string
	Enabled          bool
	Position         int
	Notes            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ServiceCostComponentDraft struct {
	ID               string
	Name             string
	Type             CostComponentType
	ReferenceID      string
	UsageMode        UsageMode
	ParameterKey     string
	RateID           string
	RateParameterKey string
	UsageQuantity    Quantity
	Multiplier       Quantity
	RateRial         int64
	Percentage       Quantity
	RateBasis        string
	Enabled          bool
	Notes            string
}

type ServiceParameter struct {
	ID                string
	ServiceID         string
	Key               string
	Label             string
	Type              ParameterType
	Required          bool
	Position          int
	DefaultValue      string
	Options           []string
	MinValue          *Quantity
	MaxValue          *Quantity
	Unit              string
	PredefinedKey     string
	PredefinedOptions []PredefinedParameterOption
	MaterialSource    *MaterialParameterSource
	Active            bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PricingRuleType string

const (
	PricingFixed       PricingRuleType = "fixed"
	PricingMarkup      PricingRuleType = "markup"
	PricingFixedMargin PricingRuleType = "fixed-margin"
	PricingPerUnit     PricingRuleType = "per-unit"
	PricingTiers       PricingRuleType = "quantity-tiers"
	PricingVariation   PricingRuleType = "variation"
	PricingManual      PricingRuleType = "manual"
)

type ServicePricingRule struct {
	ID               string
	ServiceID        string
	Type             PricingRuleType
	FixedPriceRial   int64
	MarkupPercentage Quantity
	FixedMarginRial  int64
	PerUnitRateRial  int64
	ParameterKey     string
	Tiers            []ServicePricingTier
	Variations       []ServicePricingVariation
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ServicePricingRuleDraft struct {
	ID               string
	Type             PricingRuleType
	FixedPriceRial   int64
	MarkupPercentage Quantity
	FixedMarginRial  int64
	PerUnitRateRial  int64
	ParameterKey     string
	Tiers            []ServicePricingTierDraft
	Variations       []ServicePricingVariationDraft
}

type ServicePricingTier struct {
	Position        int
	MinimumQuantity Quantity
	PriceRial       int64
}

type ServicePricingTierDraft struct {
	Position        int
	MinimumQuantity Quantity
	PriceRial       int64
}

// ServicePricingVariation is a selling-price row for one combination of
// customer-facing options (for example paper size + paper type + machine).
// Quantity-tier rules keep their tier prices inside the row so every
// combination can have its own quantity curve.
type ServicePricingVariation struct {
	ID        string
	Values    map[string]string
	PriceRial int64
	Tiers     []ServicePricingTier
	Position  int
	Active    bool
}

type ServicePricingVariationDraft struct {
	ID        string
	Values    map[string]string
	PriceRial int64
	Tiers     []ServicePricingTierDraft
	Position  int
	Active    bool
}

type ServiceParameterDraft struct {
	ID                string
	Key               string
	Label             string
	Type              ParameterType
	Required          bool
	DefaultValue      string
	Options           []string
	MinValue          *Quantity
	MaxValue          *Quantity
	Unit              string
	PredefinedKey     string
	PredefinedOptions []PredefinedParameterOption
	MaterialSource    *MaterialParameterSource
}

func NewService(id string, draft ServiceDraft, now time.Time) (Service, error) {
	service := Service{
		ID:              strings.TrimSpace(id),
		Name:            strings.TrimSpace(draft.Name),
		Code:            strings.TrimSpace(draft.Code),
		Category:        strings.TrimSpace(draft.Category),
		Description:     strings.TrimSpace(draft.Description),
		ImagePath:       strings.TrimSpace(draft.ImagePath),
		DefaultUnit:     strings.TrimSpace(draft.DefaultUnit),
		DefaultPriority: draft.DefaultPriority,
		Active:          true,
		CreatedAt:       now.UTC(),
		UpdatedAt:       now.UTC(),
	}
	if service.DefaultUnit == "" {
		service.DefaultUnit = "piece"
	}
	if service.DefaultPriority == "" {
		service.DefaultPriority = PriorityNormal
	}
	service.Parameters = make([]ServiceParameter, len(draft.Parameters))
	for index, parameter := range draft.Parameters {
		service.Parameters[index] = parameterFromDraft(service.ID, parameter, index, now)
	}
	service.Components = make([]ServiceCostComponent, len(draft.Components))
	for index, component := range draft.Components {
		service.Components[index] = componentFromDraft(service.ID, component, index, now)
	}
	if draft.PricingRule != nil {
		rule := pricingRuleFromDraft(service.ID, *draft.PricingRule, now)
		service.PricingRule = &rule
	}
	if draft.FinishedSize != nil {
		finishedSize := *draft.FinishedSize
		finishedSize.Options = append([]FinishedSizeOption(nil), draft.FinishedSize.Options...)
		service.FinishedSize = &finishedSize
	}
	service.MaterialVariants = cloneMaterialVariants(draft.MaterialVariants)
	if err := service.Validate(); err != nil {
		return Service{}, err
	}
	return service, nil
}

func (s *Service) Update(draft ServiceDraft, now time.Time) error {
	updated, err := NewService(s.ID, draft, now)
	if err != nil {
		return err
	}
	updated.Active = s.Active
	updated.CreatedAt = s.CreatedAt
	updated.UpdatedAt = now.UTC()
	for index := range updated.Parameters {
		for _, existing := range s.Parameters {
			if existing.ID != "" && existing.ID == updated.Parameters[index].ID {
				updated.Parameters[index].CreatedAt = existing.CreatedAt
				updated.Parameters[index].Active = existing.Active
			}
		}
		updated.Parameters[index].ServiceID = s.ID
		updated.Parameters[index].UpdatedAt = now.UTC()
	}
	for index := range updated.Components {
		for _, existing := range s.Components {
			if existing.ID != "" && existing.ID == updated.Components[index].ID {
				updated.Components[index].CreatedAt = existing.CreatedAt
			}
		}
		updated.Components[index].ServiceID = s.ID
		updated.Components[index].UpdatedAt = now.UTC()
	}
	if updated.PricingRule != nil && s.PricingRule != nil {
		updated.PricingRule.ID = s.PricingRule.ID
		updated.PricingRule.CreatedAt = s.PricingRule.CreatedAt
	}
	*s = updated
	return nil
}

func (s Service) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return validationError("id", "is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		return validationError("name", "is required")
	}
	if s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	keys := make(map[string]struct{}, len(s.Parameters))
	ids := make(map[string]struct{}, len(s.Parameters))
	for index, parameter := range s.Parameters {
		if parameter.ServiceID != s.ID {
			return validationError(fmt.Sprintf("parameters[%d].serviceId", index), "must match the service")
		}
		if parameter.Position != index {
			return validationError(fmt.Sprintf("parameters[%d].position", index), "must be deterministic")
		}
		if _, exists := keys[parameter.Key]; exists {
			return validationError("parameter.key", "must be unique within a service")
		}
		if _, exists := ids[parameter.ID]; exists {
			return validationError("parameter.id", "must be unique within a service")
		}
		keys[parameter.Key] = struct{}{}
		ids[parameter.ID] = struct{}{}
		if err := parameter.Validate(); err != nil {
			return fmt.Errorf("parameter %q: %w", parameter.Key, err)
		}
	}
	componentIDs := make(map[string]struct{}, len(s.Components))
	for index, component := range s.Components {
		if component.ServiceID != s.ID {
			return validationError(fmt.Sprintf("components[%d].serviceId", index), "must match the service")
		}
		if component.Position != index {
			return validationError(fmt.Sprintf("components[%d].position", index), "must be deterministic")
		}
		if _, exists := componentIDs[component.ID]; exists {
			return validationError("component.id", "must be unique within a service")
		}
		componentIDs[component.ID] = struct{}{}
		if err := component.Validate(); err != nil {
			return fmt.Errorf("component %q: %w", component.Name, err)
		}
	}
	if s.FinishedSize != nil {
		finished := s.FinishedSize
		parameterTypes := make(map[string]ParameterType, len(s.Parameters))
		for _, parameter := range s.Parameters {
			parameterTypes[parameter.Key] = parameter.Type
		}
		if finished.ParameterKey != "" && parameterTypes[finished.ParameterKey] != ParameterChoice {
			return validationError("finishedSize.parameterKey", "must reference a choice parameter")
		}
		if finished.QuantityParameterKey != "" && parameterTypes[finished.QuantityParameterKey] != ParameterInteger && parameterTypes[finished.QuantityParameterKey] != ParameterDecimal {
			return validationError("finishedSize.quantityParameterKey", "must reference a numeric quantity parameter")
		}
		if finished.AllowCustom {
			if finished.WidthParameterKey == "" || finished.HeightParameterKey == "" {
				return validationError("finishedSize", "custom sizes require width and height parameter keys")
			}
			for _, key := range []string{finished.WidthParameterKey, finished.HeightParameterKey} {
				if parameterTypes[key] != ParameterInteger && parameterTypes[key] != ParameterDecimal {
					return validationError("finishedSize", "custom dimensions must reference numeric parameters")
				}
			}
		}
		if !finished.AllowCustom && len(finished.Options) == 0 {
			return validationError("finishedSize.options", "requires at least one predefined size when custom sizes are disabled")
		}
		if len(finished.Options) > 0 {
			if finished.ParameterKey == "" {
				return validationError("finishedSize.parameterKey", "is required when predefined sizes are configured")
			}
			choiceValues := map[string]struct{}{}
			for _, parameter := range s.Parameters {
				if parameter.Key == finished.ParameterKey {
					for _, option := range parameter.Options {
						choiceValues[option] = struct{}{}
					}
				}
			}
			for _, option := range finished.Options {
				if _, exists := choiceValues[option.Code]; !exists {
					return validationError("finishedSize.options", "codes must belong to the referenced choice parameter")
				}
			}
		}
		seen := map[string]struct{}{}
		for index, option := range finished.Options {
			if option.Position != index {
				return validationError("finishedSize.options.position", "must be deterministic")
			}
			if strings.TrimSpace(option.ID) == "" || strings.TrimSpace(option.Code) == "" || strings.TrimSpace(option.Label) == "" {
				return validationError(fmt.Sprintf("finishedSize.options[%d]", index), "requires id, code, and label")
			}
			if option.WidthMM <= 0 || option.HeightMM <= 0 {
				return validationError(fmt.Sprintf("finishedSize.options[%d]", index), "dimensions must be positive")
			}
			if _, exists := seen[option.Code]; exists {
				return validationError("finishedSize.options", "codes must be unique")
			}
			seen[option.Code] = struct{}{}
		}
	}
	variantIDs := make(map[string]struct{}, len(s.MaterialVariants))
	variantKeys := make(map[string]struct{}, len(s.MaterialVariants))
	variantParameterTypes := parameterTypesForService(s.Parameters)
	materialParameterKeys := make(map[string]struct{})
	for _, parameter := range s.Parameters {
		if parameter.MaterialSource != nil || parameter.Type == ParameterMaterialReference {
			materialParameterKeys[parameter.Key] = struct{}{}
		}
	}
	for index, variant := range s.MaterialVariants {
		if strings.TrimSpace(variant.ID) == "" || strings.TrimSpace(variant.MaterialID) == "" {
			return validationError(fmt.Sprintf("materialVariants[%d]", index), "requires id and materialId")
		}
		if variant.Position != index {
			return validationError("materialVariants.position", "must be deterministic")
		}
		if _, exists := variantIDs[variant.ID]; exists {
			return validationError("materialVariants.id", "must be unique")
		}
		variantIDs[variant.ID] = struct{}{}
		if variant.SellingPriceRial < 0 {
			return validationError(fmt.Sprintf("materialVariants[%d].sellingPriceRial", index), "cannot be negative")
		}
		if len(variant.Values) == 0 {
			return validationError(fmt.Sprintf("materialVariants[%d].values", index), "must contain at least one option value")
		}
		parts := make([]string, 0, len(variant.Values))
		for key, value := range variant.Values {
			if _, exists := variantParameterTypes[key]; !exists {
				return validationError(fmt.Sprintf("materialVariants[%d].values", index), "must reference service parameters")
			}
			if _, exists := materialParameterKeys[key]; !exists {
				return validationError(fmt.Sprintf("materialVariants[%d].values", index), "must reference material-backed parameters")
			}
			if strings.TrimSpace(value) == "" {
				return validationError(fmt.Sprintf("materialVariants[%d].values", index), "cannot contain empty values")
			}
			parts = append(parts, key+"="+value)
		}
		sort.Strings(parts)
		variantKey := strings.Join(parts, "\x1f")
		if _, exists := variantKeys[variantKey]; exists {
			return validationError("materialVariants", "must not contain duplicate option combinations")
		}
		variantKeys[variantKey] = struct{}{}
	}
	if s.PricingRule != nil {
		if s.PricingRule.ServiceID != s.ID {
			return validationError("pricingRule.serviceId", "must match the service")
		}
		if err := s.PricingRule.Validate(s.Parameters); err != nil {
			return fmt.Errorf("pricing rule: %w", err)
		}
		if s.PricingRule.Type == PricingVariation && len(s.PricingRule.Variations) == 0 {
			if len(s.MaterialVariants) == 0 {
				return validationError("pricingRule", "variation pricing requires pricing variations")
			}
			for index, variant := range s.MaterialVariants {
				if variant.Active && variant.SellingPriceRial <= 0 {
					return validationError(fmt.Sprintf("materialVariants[%d].sellingPriceRial", index), "must be greater than zero for legacy variation pricing")
				}
			}
		}
	}
	parameterTypes := make(map[string]ParameterType, len(s.Parameters))
	for _, parameter := range s.Parameters {
		parameterTypes[parameter.Key] = parameter.Type
	}
	for index, component := range s.Components {
		if component.RateParameterKey != "" {
			rateParameterType, rateParameterExists := parameterTypes[component.RateParameterKey]
			if component.Type != CostMachine || !rateParameterExists || rateParameterType != ParameterChoice {
				return validationError(fmt.Sprintf("components[%d].rateParameterKey", index), "must reference a choice parameter on a machine component")
			}
		}
		if component.UsageMode != UsageParameter {
			continue
		}
		parameterType, exists := parameterTypes[component.ParameterKey]
		if !exists {
			return validationError(fmt.Sprintf("components[%d].parameterKey", index), "must reference an existing service parameter")
		}
		if component.Type == CostMaterial {
			if component.ReferenceID == "" && parameterType != ParameterMaterialReference {
				materialParameter := false
				for _, parameter := range s.Parameters {
					if parameter.Key == component.ParameterKey && parameter.Type == ParameterChoice && parameter.MaterialSource != nil {
						materialParameter = true
						break
					}
				}
				if !materialParameter {
					return validationError(fmt.Sprintf("components[%d].parameterKey", index), "must reference an explicit material parameter")
				}
			}
			if component.ReferenceID != "" && parameterType != ParameterInteger && parameterType != ParameterDecimal {
				return validationError(fmt.Sprintf("components[%d].parameterKey", index), "must reference an integer or decimal parameter")
			}
			continue
		}
		if component.Type == CostMachine {
			if component.ReferenceID == "" && parameterType != ParameterMachineReference {
				return validationError(fmt.Sprintf("components[%d].parameterKey", index), "must reference an explicit machine parameter")
			}
			if component.ReferenceID != "" && parameterType != ParameterInteger && parameterType != ParameterDecimal {
				return validationError(fmt.Sprintf("components[%d].parameterKey", index), "must reference an integer or decimal parameter")
			}
			continue
		}
		if parameterType != ParameterInteger && parameterType != ParameterDecimal {
			return validationError(fmt.Sprintf("components[%d].parameterKey", index), "must reference an integer or decimal parameter")
		}
	}
	return nil
}

func parameterTypesForService(parameters []ServiceParameter) map[string]ParameterType {
	result := make(map[string]ParameterType, len(parameters))
	for _, parameter := range parameters {
		result[parameter.Key] = parameter.Type
	}
	return result
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func cloneMaterialVariants(variants []ServiceMaterialVariant) []ServiceMaterialVariant {
	if len(variants) == 0 {
		return nil
	}
	result := make([]ServiceMaterialVariant, len(variants))
	for index, variant := range variants {
		result[index] = variant
		result[index].Values = make(map[string]string, len(variant.Values))
		for key, value := range variant.Values {
			result[index].Values[key] = value
		}
	}
	return result
}

func (c ServiceCostComponent) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return validationError("id", "is required")
	}
	if strings.TrimSpace(c.Name) == "" {
		return validationError("name", "is required")
	}
	if c.Position < 0 {
		return validationError("position", "cannot be negative")
	}
	if c.CreatedAt.IsZero() || c.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	if c.Multiplier <= 0 {
		return validationError("multiplier", "must be greater than zero")
	}
	if c.UsageQuantity < 0 {
		return validationError("usageQuantity", "cannot be negative")
	}
	if c.RateRial < 0 {
		return validationError("rateRial", "cannot be negative")
	}
	if c.Percentage < 0 || c.Percentage > Quantity(100*QuantityScale) {
		return validationError("percentage", "must be between 0 and 100")
	}
	switch c.Type {
	case CostMaterial:
		if c.UsageMode == UsageFixed && strings.TrimSpace(c.ReferenceID) == "" {
			return validationError("referenceId", "is required")
		}
		if c.RateRial != 0 || c.Percentage != 0 {
			return validationError("rateRial", "is not supported for referenced components")
		}
		if c.RateID != "" || c.RateParameterKey != "" {
			return validationError("rateId", "is only supported for machine components")
		}
	case CostMachine:
		if c.UsageMode == UsageFixed && strings.TrimSpace(c.ReferenceID) == "" {
			return validationError("referenceId", "is required")
		}
		if c.RateRial != 0 || c.Percentage != 0 {
			return validationError("rateRial", "is not supported for referenced components")
		}
	case CostService:
		if strings.TrimSpace(c.ReferenceID) == "" {
			return validationError("referenceId", "is required")
		}
		if c.RateRial != 0 || c.Percentage != 0 {
			return validationError("rateRial", "is not supported for referenced components")
		}
		if c.RateID != "" || c.RateParameterKey != "" {
			return validationError("rateId", "is only supported for machine components")
		}
	case CostLabor, CostOutsourced, CostFixed, CostManual:
		if c.ReferenceID != "" || c.Percentage != 0 || c.RateID != "" || c.RateParameterKey != "" {
			return validationError("referenceId", "is not supported for this component type")
		}
		if c.Type == CostLabor {
			validBasis := false
			for _, basis := range SupportedRateBases() {
				if c.RateBasis == basis {
					validBasis = true
					break
				}
			}
			if !validBasis {
				return validationError("rateBasis", "must be unit, minute, or hour for labor")
			}
		}
	case CostOverhead, CostWaste:
		if c.Percentage <= 0 {
			return validationError("percentage", "must be greater than zero")
		}
		if c.ReferenceID != "" || c.RateRial != 0 || c.UsageMode != UsageFixed || c.ParameterKey != "" {
			return validationError("percentage", "must be the only cost input for this component type")
		}
	default:
		return validationError("type", "is not supported")
	}
	if c.Type != CostOverhead && c.Type != CostWaste {
		if c.UsageMode != UsageFixed && c.UsageMode != UsageParameter {
			return validationError("usageMode", "must be fixed or parameter")
		}
		if c.UsageMode == UsageFixed && c.ParameterKey != "" {
			return validationError("parameterKey", "must be empty for fixed usage")
		}
		if c.UsageMode == UsageParameter && strings.TrimSpace(c.ParameterKey) == "" {
			return validationError("parameterKey", "is required for parameter usage")
		}
	}
	return nil
}

func (r ServicePricingRule) Validate(parameters []ServiceParameter) error {
	if strings.TrimSpace(r.ID) == "" {
		return validationError("id", "is required")
	}
	if strings.TrimSpace(r.ServiceID) == "" {
		return validationError("serviceId", "is required")
	}
	if r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	if r.FixedPriceRial < 0 || r.FixedMarginRial < 0 || r.PerUnitRateRial < 0 {
		return validationError("pricing", "money values cannot be negative")
	}
	if r.MarkupPercentage < 0 || r.MarkupPercentage > Quantity(1000*QuantityScale) {
		return validationError("markupPercentage", "must be between 0 and 1000")
	}
	parameterTypes := make(map[string]ParameterType, len(parameters))
	for _, parameter := range parameters {
		parameterTypes[parameter.Key] = parameter.Type
	}
	switch r.Type {
	case PricingFixed, PricingMarkup, PricingFixedMargin, PricingManual:
	case PricingVariation:
		if len(r.Variations) > 0 {
			if err := validatePricingVariations(r.Variations, parameterTypes, false); err != nil {
				return err
			}
		}
	case PricingPerUnit:
		return validateNumericPricingParameter(r.ParameterKey, parameterTypes)
	case PricingTiers:
		if err := validateNumericPricingParameter(r.ParameterKey, parameterTypes); err != nil {
			return err
		}
		if len(r.Variations) > 0 {
			if err := validatePricingVariations(r.Variations, parameterTypes, true); err != nil {
				return err
			}
		} else {
			if err := validatePricingTiers(r.Tiers, "tiers"); err != nil {
				return err
			}
		}
	default:
		return validationError("type", "is not supported")
	}
	return nil
}

func validatePricingVariations(variations []ServicePricingVariation, parameterTypes map[string]ParameterType, quantityTiers bool) error {
	if len(variations) == 0 {
		return validationError("variations", "must contain at least one variation")
	}
	seenIDs := make(map[string]struct{}, len(variations))
	seenValues := make(map[string]struct{}, len(variations))
	for index, variation := range variations {
		if strings.TrimSpace(variation.ID) == "" {
			return validationError(fmt.Sprintf("variations[%d].id", index), "is required")
		}
		if variation.Position != index {
			return validationError("variations.position", "must be deterministic")
		}
		if _, exists := seenIDs[variation.ID]; exists {
			return validationError("variations.id", "must be unique")
		}
		seenIDs[variation.ID] = struct{}{}
		if len(variation.Values) == 0 {
			return validationError(fmt.Sprintf("variations[%d].values", index), "must contain at least one option")
		}
		parts := make([]string, 0, len(variation.Values))
		for key, value := range variation.Values {
			if _, exists := parameterTypes[key]; !exists {
				return validationError(fmt.Sprintf("variations[%d].values", index), "must reference service parameters")
			}
			if strings.TrimSpace(value) == "" {
				return validationError(fmt.Sprintf("variations[%d].values", index), "cannot contain empty values")
			}
			parts = append(parts, key+"="+value)
		}
		sort.Strings(parts)
		key := strings.Join(parts, "\x1f")
		if _, exists := seenValues[key]; exists {
			return validationError("variations", "must not contain duplicate option combinations")
		}
		seenValues[key] = struct{}{}
		if !quantityTiers && variation.PriceRial <= 0 {
			return validationError(fmt.Sprintf("variations[%d].priceRial", index), "must be greater than zero")
		}
		if variation.PriceRial < 0 {
			return validationError(fmt.Sprintf("variations[%d].priceRial", index), "cannot be negative")
		}
		if quantityTiers {
			if err := validatePricingTiers(variation.Tiers, fmt.Sprintf("variations[%d].tiers", index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func validatePricingTiers(tiers []ServicePricingTier, field string) error {
	if len(tiers) == 0 {
		return validationError(field, "must contain at least one tier")
	}
	for index, tier := range tiers {
		if tier.Position != index {
			return validationError(field+".position", "must be deterministic")
		}
		if tier.MinimumQuantity < 0 || tier.PriceRial < 0 {
			return validationError(field, "cannot contain negative values")
		}
		if index > 0 && tier.MinimumQuantity <= tiers[index-1].MinimumQuantity {
			return validationError(field, "must be ordered by increasing minimum quantity")
		}
	}
	if tiers[0].MinimumQuantity != 0 {
		return validationError(field+"[0].minimumQuantity", "must be zero")
	}
	return nil
}

func validateNumericPricingParameter(key string, parameterTypes map[string]ParameterType) error {
	if strings.TrimSpace(key) == "" {
		return validationError("parameterKey", "is required")
	}
	typeName, exists := parameterTypes[key]
	if !exists || (typeName != ParameterInteger && typeName != ParameterDecimal) {
		return validationError("parameterKey", "must reference an integer or decimal parameter")
	}
	return nil
}

func (p ServiceParameter) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return validationError("id", "is required")
	}
	if !parameterKeyPattern.MatchString(p.Key) {
		return validationError("key", "must use lowercase letters, numbers, and underscores")
	}
	if strings.TrimSpace(p.Label) == "" {
		return validationError("label", "is required")
	}
	if p.Position < 0 {
		return validationError("position", "cannot be negative")
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	if p.Type != ParameterChoice && len(p.Options) > 0 {
		return validationError("options", "are only supported for choice parameters")
	}
	if p.PredefinedKey != "" {
		if p.Type != ParameterChoice {
			return validationError("predefinedKey", "is only supported for choice parameters")
		}
		if !IsSupportedPredefinedParameter(p.PredefinedKey) {
			return validationError("predefinedKey", "is not supported")
		}
		if p.MaterialSource != nil {
			return validationError("predefinedKey", "cannot be combined with a material source")
		}
		if p.DefaultValue != "" {
			found := false
			for _, option := range p.PredefinedOptions {
				if option.Code == p.DefaultValue && option.Active {
					found = true
					break
				}
			}
			if len(p.PredefinedOptions) > 0 && !found {
				return validationError("defaultValue", "must belong to the predefined parameter options")
			}
		}
	}
	if p.MaterialSource != nil {
		if p.Type != ParameterChoice {
			return validationError("materialSource", "is only supported for choice parameters")
		}
		attributeKeys := p.MaterialSource.AttributeKeys()
		if len(p.MaterialSource.AllowedKinds) == 0 && len(attributeKeys) == 0 && !p.MaterialSource.SelectMaterial {
			return validationError("materialSource", "must expose an attribute or select materials explicitly")
		}
		seenAttributeKeys := make(map[string]struct{}, len(attributeKeys))
		for index, key := range attributeKeys {
			key = strings.TrimSpace(key)
			if key == "" {
				return validationError(fmt.Sprintf("materialSource.exposedAttributeKeys[%d]", index), "is required")
			}
			if _, exists := seenAttributeKeys[key]; exists {
				return validationError("materialSource.exposedAttributeKeys", "must be unique")
			}
			seenAttributeKeys[key] = struct{}{}
		}
		for index, kind := range p.MaterialSource.AllowedKinds {
			if !IsValidMaterialKind(kind) {
				return validationError(fmt.Sprintf("materialSource.allowedKinds[%d]", index), "is not a supported material kind")
			}
		}
		if len(attributeKeys) == 0 && len(p.MaterialSource.AllowedValues) > 0 {
			return validationError("materialSource.allowedValues", "require an exposed attribute")
		}
		for index, value := range p.MaterialSource.AllowedValues {
			if err := value.Validate(); err != nil {
				return validationError(fmt.Sprintf("materialSource.allowedValues[%d]", index), err.Error())
			}
			if len(attributeKeys) > 0 && !containsString(attributeKeys, value.Key) {
				return validationError(fmt.Sprintf("materialSource.allowedValues[%d].key", index), "must match the exposed attribute")
			}
		}
		for index, filter := range p.MaterialSource.AdditionalFilters {
			if strings.TrimSpace(filter.Key) == "" {
				return validationError(fmt.Sprintf("materialSource.additionalFilters[%d]", index), "requires an attribute key")
			}
			if err := filter.Value.Validate(); err != nil {
				return validationError(fmt.Sprintf("materialSource.additionalFilters[%d]", index), err.Error())
			}
		}
	}
	if p.Type != ParameterInteger && p.Type != ParameterDecimal && (p.MinValue != nil || p.MaxValue != nil) {
		return validationError("bounds", "are only supported for numeric parameters")
	}
	switch p.Type {
	case ParameterInteger:
		if err := validateIntegerValue("defaultValue", p.DefaultValue); err != nil {
			return err
		}
		if err := validateNumericBounds(p.MinValue, p.MaxValue); err != nil {
			return err
		}
		if p.MinValue != nil && !isWholeQuantity(*p.MinValue) {
			return validationError("minValue", "must be a whole number for integer parameters")
		}
		if p.MaxValue != nil && !isWholeQuantity(*p.MaxValue) {
			return validationError("maxValue", "must be a whole number for integer parameters")
		}
		if p.DefaultValue != "" {
			defaultValue, err := ParseQuantity(p.DefaultValue)
			if err != nil {
				return validationError("defaultValue", "is outside the supported quantity range")
			}
			if err := validateDefaultBounds(defaultValue, p.MinValue, p.MaxValue); err != nil {
				return err
			}
		}
	case ParameterDecimal:
		if err := validateDecimalValue("defaultValue", p.DefaultValue); err != nil {
			return err
		}
		if err := validateNumericBounds(p.MinValue, p.MaxValue); err != nil {
			return err
		}
		if p.DefaultValue != "" {
			defaultValue, err := ParseQuantity(p.DefaultValue)
			if err != nil {
				return validationError("defaultValue", "is outside the supported quantity range")
			}
			if err := validateDefaultBounds(defaultValue, p.MinValue, p.MaxValue); err != nil {
				return err
			}
		}
	case ParameterBoolean:
		if p.DefaultValue != "" && p.DefaultValue != "true" && p.DefaultValue != "false" {
			return validationError("defaultValue", "must be true or false")
		}
	case ParameterChoice:
		if len(p.Options) == 0 && p.MaterialSource == nil && p.PredefinedKey == "" {
			return validationError("options", "must contain at least one option")
		}
		options := make(map[string]struct{}, len(p.Options))
		for index, option := range p.Options {
			option = strings.TrimSpace(option)
			if option == "" {
				return validationError(fmt.Sprintf("options[%d]", index), "cannot be empty")
			}
			if _, exists := options[option]; exists {
				return validationError("options", "must not contain duplicates")
			}
			options[option] = struct{}{}
		}
		if p.DefaultValue != "" && p.MaterialSource == nil && p.PredefinedKey == "" {
			if _, exists := options[p.DefaultValue]; !exists {
				return validationError("defaultValue", "must belong to the choice options")
			}
		}
	case ParameterMaterialReference:
		if p.DefaultValue != "" && strings.TrimSpace(p.DefaultValue) == "" {
			return validationError("defaultValue", "must reference a material")
		}
	case ParameterMachineReference:
		if p.DefaultValue != "" && strings.TrimSpace(p.DefaultValue) == "" {
			return validationError("defaultValue", "must reference a machine")
		}
	default:
		return validationError("type", "is not supported")
	}
	return nil
}

func parameterFromDraft(serviceID string, draft ServiceParameterDraft, position int, now time.Time) ServiceParameter {
	options := make([]string, len(draft.Options))
	for index, option := range draft.Options {
		options[index] = strings.TrimSpace(option)
	}
	return ServiceParameter{
		ID: draft.ID, ServiceID: serviceID, Key: strings.TrimSpace(draft.Key), Label: strings.TrimSpace(draft.Label),
		Type: ParameterType(strings.ToLower(strings.TrimSpace(string(draft.Type)))), Required: draft.Required,
		Position: position, DefaultValue: strings.TrimSpace(draft.DefaultValue), Options: options,
		PredefinedKey: strings.TrimSpace(draft.PredefinedKey), PredefinedOptions: append([]PredefinedParameterOption(nil), draft.PredefinedOptions...), MaterialSource: draft.MaterialSource,
		MinValue: draft.MinValue, MaxValue: draft.MaxValue, Unit: strings.TrimSpace(draft.Unit), Active: true,
		CreatedAt: now.UTC(), UpdatedAt: now.UTC(),
	}
}

func componentFromDraft(serviceID string, draft ServiceCostComponentDraft, position int, now time.Time) ServiceCostComponent {
	usageMode := UsageMode(strings.ToLower(strings.TrimSpace(string(draft.UsageMode))))
	if usageMode == "" {
		usageMode = UsageFixed
	}
	usageQuantity := draft.UsageQuantity
	if usageQuantity == 0 {
		usageQuantity = Quantity(QuantityScale)
	}
	return ServiceCostComponent{ID: draft.ID, ServiceID: serviceID, Name: strings.TrimSpace(draft.Name), Type: CostComponentType(strings.ToLower(strings.TrimSpace(string(draft.Type)))), ReferenceID: strings.TrimSpace(draft.ReferenceID), UsageMode: usageMode, ParameterKey: strings.TrimSpace(draft.ParameterKey), RateID: strings.TrimSpace(draft.RateID), RateParameterKey: strings.TrimSpace(draft.RateParameterKey), UsageQuantity: usageQuantity, Multiplier: draft.Multiplier, RateRial: draft.RateRial, Percentage: draft.Percentage, RateBasis: strings.TrimSpace(draft.RateBasis), Enabled: draft.Enabled, Position: position, Notes: strings.TrimSpace(draft.Notes), CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
}

func pricingRuleFromDraft(serviceID string, draft ServicePricingRuleDraft, now time.Time) ServicePricingRule {
	tiers := make([]ServicePricingTier, len(draft.Tiers))
	for index, tier := range draft.Tiers {
		tiers[index] = ServicePricingTier{Position: index, MinimumQuantity: tier.MinimumQuantity, PriceRial: tier.PriceRial}
	}
	id := draft.ID
	if id == "" {
		id = "pricing-" + serviceID
	}
	variations := make([]ServicePricingVariation, len(draft.Variations))
	for index, variation := range draft.Variations {
		variationTiers := make([]ServicePricingTier, len(variation.Tiers))
		for tierIndex, tier := range variation.Tiers {
			variationTiers[tierIndex] = ServicePricingTier{Position: tierIndex, MinimumQuantity: tier.MinimumQuantity, PriceRial: tier.PriceRial}
		}
		values := make(map[string]string, len(variation.Values))
		for key, value := range variation.Values {
			values[key] = value
		}
		variations[index] = ServicePricingVariation{ID: variation.ID, Values: values, PriceRial: variation.PriceRial, Tiers: variationTiers, Position: index, Active: variation.Active}
	}
	return ServicePricingRule{ID: id, ServiceID: serviceID, Type: PricingRuleType(strings.ToLower(strings.TrimSpace(string(draft.Type)))), FixedPriceRial: draft.FixedPriceRial, MarkupPercentage: draft.MarkupPercentage, FixedMarginRial: draft.FixedMarginRial, PerUnitRateRial: draft.PerUnitRateRial, ParameterKey: strings.TrimSpace(draft.ParameterKey), Tiers: tiers, Variations: variations, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
}

func validateNumericBounds(minimum, maximum *Quantity) error {
	if minimum != nil && *minimum < 0 {
		return validationError("minValue", "cannot be negative")
	}
	if maximum != nil && *maximum < 0 {
		return validationError("maxValue", "cannot be negative")
	}
	if minimum != nil && maximum != nil && *minimum > *maximum {
		return validationError("minValue", "must be less than or equal to maxValue")
	}
	return nil
}

func validateDefaultBounds(value Quantity, minimum, maximum *Quantity) error {
	if minimum != nil && value < *minimum {
		return validationError("defaultValue", "must be greater than or equal to minValue")
	}
	if maximum != nil && value > *maximum {
		return validationError("defaultValue", "must be less than or equal to maxValue")
	}
	return nil
}

func validateIntegerValue(field, value string) error {
	if value == "" {
		return nil
	}
	if strings.Contains(value, ".") {
		return validationError(field, "must be a whole number")
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return validationError(field, "must be a non-negative whole number")
	}
	return nil
}

func validateDecimalValue(field, value string) error {
	if value == "" {
		return nil
	}
	quantity, err := ParseQuantity(value)
	if err != nil {
		return validationError(field, "must be a non-negative decimal with at most six fractional digits")
	}
	if quantity < 0 {
		return validationError(field, "cannot be negative")
	}
	return nil
}

func isWholeQuantity(value Quantity) bool {
	return value%QuantityScale == 0
}

func NormalizeParameterOrder(parameters []ServiceParameter) {
	sort.SliceStable(parameters, func(left, right int) bool {
		return parameters[left].Position < parameters[right].Position
	})
	for index := range parameters {
		parameters[index].Position = index
	}
}

func NormalizeComponentOrder(components []ServiceCostComponent) {
	sort.SliceStable(components, func(left, right int) bool { return components[left].Position < components[right].Position })
	for index := range components {
		components[index].Position = index
	}
}
