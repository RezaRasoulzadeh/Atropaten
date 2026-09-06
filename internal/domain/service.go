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
	ErrServiceNotFound       = errors.New("service not found")
	ErrParameterNotFound     = errors.New("service parameter not found")
	ErrCostComponentNotFound = errors.New("service cost component not found")
)

type ParameterType string

const (
	ParameterInteger           ParameterType = "integer"
	ParameterDecimal           ParameterType = "decimal"
	ParameterBoolean           ParameterType = "boolean"
	ParameterChoice            ParameterType = "choice"
	ParameterMaterialReference ParameterType = "material-reference"
)

var parameterKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type Service struct {
	ID          string
	Name        string
	Code        string
	Category    string
	Description string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Parameters  []ServiceParameter
	Components  []ServiceCostComponent
}

type ServiceDraft struct {
	Name        string
	Code        string
	Category    string
	Description string
	Parameters  []ServiceParameterDraft
	Components  []ServiceCostComponentDraft
}

type CostComponentType string

const (
	// Percentage components are definitions only in M1-003. M1-004 applies
	// overhead and waste to the enabled subtotal accumulated before them.
	CostMaterial   CostComponentType = "material"
	CostMachine    CostComponentType = "machine"
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
	ID           string
	ServiceID    string
	Name         string
	Type         CostComponentType
	ReferenceID  string
	UsageMode    UsageMode
	ParameterKey string
	Multiplier   Quantity
	RateRial     int64
	Percentage   Quantity
	RateBasis    string
	Enabled      bool
	Position     int
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ServiceCostComponentDraft struct {
	ID           string
	Name         string
	Type         CostComponentType
	ReferenceID  string
	UsageMode    UsageMode
	ParameterKey string
	Multiplier   Quantity
	RateRial     int64
	Percentage   Quantity
	RateBasis    string
	Enabled      bool
	Notes        string
}

type ServiceParameter struct {
	ID           string
	ServiceID    string
	Key          string
	Label        string
	Type         ParameterType
	Required     bool
	Position     int
	DefaultValue string
	Options      []string
	MinValue     *Quantity
	MaxValue     *Quantity
	Unit         string
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ServiceParameterDraft struct {
	ID           string
	Key          string
	Label        string
	Type         ParameterType
	Required     bool
	DefaultValue string
	Options      []string
	MinValue     *Quantity
	MaxValue     *Quantity
	Unit         string
}

func NewService(id string, draft ServiceDraft, now time.Time) (Service, error) {
	service := Service{
		ID:          strings.TrimSpace(id),
		Name:        strings.TrimSpace(draft.Name),
		Code:        strings.TrimSpace(draft.Code),
		Category:    strings.TrimSpace(draft.Category),
		Description: strings.TrimSpace(draft.Description),
		Active:      true,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC(),
	}
	service.Parameters = make([]ServiceParameter, len(draft.Parameters))
	for index, parameter := range draft.Parameters {
		service.Parameters[index] = parameterFromDraft(service.ID, parameter, index, now)
	}
	service.Components = make([]ServiceCostComponent, len(draft.Components))
	for index, component := range draft.Components {
		service.Components[index] = componentFromDraft(service.ID, component, index, now)
	}
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
	parameterTypes := make(map[string]ParameterType, len(s.Parameters))
	for _, parameter := range s.Parameters {
		parameterTypes[parameter.Key] = parameter.Type
	}
	for index, component := range s.Components {
		if component.UsageMode == UsageParameter {
			parameterType, exists := parameterTypes[component.ParameterKey]
			if !exists {
				return validationError(fmt.Sprintf("components[%d].parameterKey", index), "must reference an existing service parameter")
			}
			if parameterType != ParameterInteger && parameterType != ParameterDecimal {
				return validationError(fmt.Sprintf("components[%d].parameterKey", index), "must reference an integer or decimal parameter")
			}
		}
	}
	return nil
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
	if c.RateRial < 0 {
		return validationError("rateRial", "cannot be negative")
	}
	if c.Percentage < 0 || c.Percentage > Quantity(100*QuantityScale) {
		return validationError("percentage", "must be between 0 and 100")
	}
	switch c.Type {
	case CostMaterial, CostMachine:
		if strings.TrimSpace(c.ReferenceID) == "" {
			return validationError("referenceId", "is required")
		}
		if c.RateRial != 0 || c.Percentage != 0 {
			return validationError("rateRial", "is not supported for referenced components")
		}
	case CostLabor, CostOutsourced, CostFixed, CostManual:
		if c.ReferenceID != "" || c.Percentage != 0 {
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
		if len(p.Options) == 0 {
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
		if p.DefaultValue != "" {
			if _, exists := options[p.DefaultValue]; !exists {
				return validationError("defaultValue", "must belong to the choice options")
			}
		}
	case ParameterMaterialReference:
		if p.DefaultValue != "" && strings.TrimSpace(p.DefaultValue) == "" {
			return validationError("defaultValue", "must reference a material")
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
		MinValue: draft.MinValue, MaxValue: draft.MaxValue, Unit: strings.TrimSpace(draft.Unit), Active: true,
		CreatedAt: now.UTC(), UpdatedAt: now.UTC(),
	}
}

func componentFromDraft(serviceID string, draft ServiceCostComponentDraft, position int, now time.Time) ServiceCostComponent {
	usageMode := UsageMode(strings.ToLower(strings.TrimSpace(string(draft.UsageMode))))
	if usageMode == "" {
		usageMode = UsageFixed
	}
	return ServiceCostComponent{ID: draft.ID, ServiceID: serviceID, Name: strings.TrimSpace(draft.Name), Type: CostComponentType(strings.ToLower(strings.TrimSpace(string(draft.Type)))), ReferenceID: strings.TrimSpace(draft.ReferenceID), UsageMode: usageMode, ParameterKey: strings.TrimSpace(draft.ParameterKey), Multiplier: draft.Multiplier, RateRial: draft.RateRial, Percentage: draft.Percentage, RateBasis: strings.TrimSpace(draft.RateBasis), Enabled: draft.Enabled, Position: position, Notes: strings.TrimSpace(draft.Notes), CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
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
