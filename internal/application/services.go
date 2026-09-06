package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type ServiceRepository interface {
	ListServices(context.Context, bool) ([]domain.Service, error)
	GetService(context.Context, string) (domain.Service, error)
	SaveServiceDefinition(context.Context, domain.Service) error
}

type MaterialLookup interface {
	Get(context.Context, string) (domain.Material, error)
}

type MachineLookup interface {
	GetMachine(context.Context, string) (domain.Machine, error)
}

type ParameterInput struct {
	ID           string
	Key          string
	Label        string
	Type         string
	Required     bool
	DefaultValue string
	Options      []string
	MinValue     *string
	MaxValue     *string
	Unit         string
}

type ServiceInput struct {
	Name        string
	Code        string
	Category    string
	Description string
	Parameters  []ParameterInput
	Components  []CostComponentInput
}

type CostComponentInput struct {
	ID           string
	Name         string
	Type         string
	ReferenceID  string
	UsageMode    string
	ParameterKey string
	Multiplier   string
	RateRial     int64
	Percentage   string
	RateBasis    string
	Enabled      bool
	Notes        string
}

type ParameterView struct {
	ID           string
	Key          string
	Label        string
	Type         string
	Required     bool
	Position     int
	DefaultValue string
	Options      []string
	MinValue     *string
	MaxValue     *string
	Unit         string
	Active       bool
}

type ServiceView struct {
	ID          string
	Name        string
	Code        string
	Category    string
	Description string
	Active      bool
	CreatedAt   string
	UpdatedAt   string
	Parameters  []ParameterView
	Components  []CostComponentView
}

type CostComponentView struct {
	ID           string
	Name         string
	Type         string
	ReferenceID  string
	UsageMode    string
	ParameterKey string
	Multiplier   string
	RateRial     int64
	Percentage   string
	RateBasis    string
	Enabled      bool
	Position     int
	Notes        string
}

type ServicesService struct {
	repository ServiceRepository
	material   MaterialLookup
	machine    MachineLookup
	now        func() time.Time
}

func NewServicesService(repository ServiceRepository, material MaterialLookup, machine MachineLookup) *ServicesService {
	return &ServicesService{repository: repository, material: material, machine: machine, now: time.Now}
}

func (s *ServicesService) List(ctx context.Context, includeArchived bool) ([]ServiceView, error) {
	services, err := s.repository.ListServices(ctx, includeArchived)
	if err != nil {
		return nil, err
	}
	views := make([]ServiceView, 0, len(services))
	for _, service := range services {
		views = append(views, serviceView(service))
	}
	return views, nil
}

func (s *ServicesService) Get(ctx context.Context, id string) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(id))
	if err != nil {
		return ServiceView{}, err
	}
	return serviceView(service), nil
}

func (s *ServicesService) Create(ctx context.Context, input ServiceInput) (ServiceView, error) {
	id, err := newID("SVC-")
	if err != nil {
		return ServiceView{}, fmt.Errorf("create service id: %w", err)
	}
	draft, err := s.parseDraft(ctx, input, id, nil)
	if err != nil {
		return ServiceView{}, err
	}
	service, err := domain.NewService(id, draft, s.now())
	if err != nil {
		return ServiceView{}, err
	}
	if err := s.validateReferences(ctx, service); err != nil {
		return ServiceView{}, err
	}
	if err := s.repository.SaveServiceDefinition(ctx, service); err != nil {
		return ServiceView{}, err
	}
	return serviceView(service), nil
}

func (s *ServicesService) Update(ctx context.Context, id string, input ServiceInput) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(id))
	if err != nil {
		return ServiceView{}, err
	}
	draft, err := s.parseDraft(ctx, input, service.ID, service.Parameters)
	if err != nil {
		return ServiceView{}, err
	}
	if err := service.Update(draft, s.now()); err != nil {
		return ServiceView{}, err
	}
	if err := s.validateReferences(ctx, service); err != nil {
		return ServiceView{}, err
	}
	if err := s.repository.SaveServiceDefinition(ctx, service); err != nil {
		return ServiceView{}, err
	}
	return serviceView(service), nil
}

func (s *ServicesService) Archive(ctx context.Context, id string) (ServiceView, error) {
	return s.setActive(ctx, id, false)
}

func (s *ServicesService) Reactivate(ctx context.Context, id string) (ServiceView, error) {
	return s.setActive(ctx, id, true)
}

func (s *ServicesService) setActive(ctx context.Context, id string, active bool) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(id))
	if err != nil {
		return ServiceView{}, err
	}
	service.Active = active
	service.UpdatedAt = s.now().UTC()
	if err := s.repository.SaveServiceDefinition(ctx, service); err != nil {
		return ServiceView{}, err
	}
	return serviceView(service), nil
}

func (s *ServicesService) AddParameter(ctx context.Context, serviceID string, input ParameterInput) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(serviceID))
	if err != nil {
		return ServiceView{}, err
	}
	draft, err := s.parseParameter(ctx, input, service.ID, nil)
	if err != nil {
		return ServiceView{}, err
	}
	parameter := parameterFromDraft(service.ID, draft, len(service.Parameters), s.now())
	service.Parameters = append(service.Parameters, parameter)
	return s.saveDefinition(ctx, service)
}

func (s *ServicesService) UpdateParameter(ctx context.Context, serviceID, parameterID string, input ParameterInput) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(serviceID))
	if err != nil {
		return ServiceView{}, err
	}
	parameterIndex := -1
	for index, parameter := range service.Parameters {
		if parameter.ID == parameterID {
			parameterIndex = index
			break
		}
	}
	if parameterIndex < 0 {
		return ServiceView{}, domain.ErrParameterNotFound
	}
	input.ID = parameterID
	draft, err := s.parseParameter(ctx, input, service.ID, service.Parameters)
	if err != nil {
		return ServiceView{}, err
	}
	service.Parameters[parameterIndex] = parameterFromDraft(service.ID, draft, service.Parameters[parameterIndex].Position, s.now())
	return s.saveDefinition(ctx, service)
}

func (s *ServicesService) RemoveParameter(ctx context.Context, serviceID, parameterID string) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(serviceID))
	if err != nil {
		return ServiceView{}, err
	}
	filtered := service.Parameters[:0]
	found := false
	for _, parameter := range service.Parameters {
		if parameter.ID == parameterID {
			found = true
			continue
		}
		filtered = append(filtered, parameter)
	}
	if !found {
		return ServiceView{}, domain.ErrParameterNotFound
	}
	service.Parameters = filtered
	domain.NormalizeParameterOrder(service.Parameters)
	return s.saveDefinition(ctx, service)
}

func (s *ServicesService) ReorderParameters(ctx context.Context, serviceID string, parameterIDs []string) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(serviceID))
	if err != nil {
		return ServiceView{}, err
	}
	if len(parameterIDs) != len(service.Parameters) {
		return ServiceView{}, domain.ValidationError{Field: "parameterIDs", Message: "must include every parameter exactly once"}
	}
	byID := make(map[string]domain.ServiceParameter, len(service.Parameters))
	for _, parameter := range service.Parameters {
		byID[parameter.ID] = parameter
	}
	reordered := make([]domain.ServiceParameter, 0, len(parameterIDs))
	for _, id := range parameterIDs {
		parameter, exists := byID[id]
		if !exists {
			return ServiceView{}, domain.ValidationError{Field: "parameterIDs", Message: "contains an unknown parameter"}
		}
		if len(reordered) > 0 && reordered[len(reordered)-1].ID == id {
			return ServiceView{}, domain.ValidationError{Field: "parameterIDs", Message: "must not contain duplicates"}
		}
		reordered = append(reordered, parameter)
		delete(byID, id)
	}
	if len(byID) != 0 {
		return ServiceView{}, domain.ValidationError{Field: "parameterIDs", Message: "must include every parameter exactly once"}
	}
	service.Parameters = reordered
	domain.NormalizeParameterOrder(service.Parameters)
	return s.saveDefinition(ctx, service)
}

func (s *ServicesService) AddCostComponent(ctx context.Context, serviceID string, input CostComponentInput) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(serviceID))
	if err != nil {
		return ServiceView{}, err
	}
	draft, err := s.parseComponent(input)
	if err != nil {
		return ServiceView{}, err
	}
	service.Components = append(service.Components, componentFromDraft(service.ID, draft, len(service.Components), s.now()))
	return s.saveDefinition(ctx, service)
}

func (s *ServicesService) UpdateCostComponent(ctx context.Context, serviceID, componentID string, input CostComponentInput) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(serviceID))
	if err != nil {
		return ServiceView{}, err
	}
	componentIndex := -1
	for index, component := range service.Components {
		if component.ID == componentID {
			componentIndex = index
			break
		}
	}
	if componentIndex < 0 {
		return ServiceView{}, domain.ErrCostComponentNotFound
	}
	draft, err := s.parseComponent(input)
	if err != nil {
		return ServiceView{}, err
	}
	draft.ID = componentID
	service.Components[componentIndex] = componentFromDraft(service.ID, draft, service.Components[componentIndex].Position, s.now())
	return s.saveDefinition(ctx, service)
}

func (s *ServicesService) RemoveCostComponent(ctx context.Context, serviceID, componentID string) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(serviceID))
	if err != nil {
		return ServiceView{}, err
	}
	filtered := service.Components[:0]
	found := false
	for _, component := range service.Components {
		if component.ID == componentID {
			found = true
			continue
		}
		filtered = append(filtered, component)
	}
	if !found {
		return ServiceView{}, domain.ErrCostComponentNotFound
	}
	service.Components = filtered
	domain.NormalizeComponentOrder(service.Components)
	return s.saveDefinition(ctx, service)
}

func (s *ServicesService) ReorderCostComponents(ctx context.Context, serviceID string, componentIDs []string) (ServiceView, error) {
	service, err := s.repository.GetService(ctx, strings.TrimSpace(serviceID))
	if err != nil {
		return ServiceView{}, err
	}
	if len(componentIDs) != len(service.Components) {
		return ServiceView{}, domain.ValidationError{Field: "componentIDs", Message: "must include every component exactly once"}
	}
	byID := make(map[string]domain.ServiceCostComponent, len(service.Components))
	for _, component := range service.Components {
		byID[component.ID] = component
	}
	reordered := make([]domain.ServiceCostComponent, 0, len(componentIDs))
	for _, id := range componentIDs {
		component, exists := byID[id]
		if !exists {
			return ServiceView{}, domain.ValidationError{Field: "componentIDs", Message: "contains an unknown component"}
		}
		reordered = append(reordered, component)
		delete(byID, id)
	}
	if len(byID) != 0 {
		return ServiceView{}, domain.ValidationError{Field: "componentIDs", Message: "must include every component exactly once"}
	}
	service.Components = reordered
	domain.NormalizeComponentOrder(service.Components)
	return s.saveDefinition(ctx, service)
}

func (s *ServicesService) saveDefinition(ctx context.Context, service domain.Service) (ServiceView, error) {
	service.UpdatedAt = s.now().UTC()
	if err := service.Validate(); err != nil {
		return ServiceView{}, err
	}
	if err := s.validateReferences(ctx, service); err != nil {
		return ServiceView{}, err
	}
	if err := s.repository.SaveServiceDefinition(ctx, service); err != nil {
		return ServiceView{}, err
	}
	return serviceView(service), nil
}

func (s *ServicesService) validateReferences(ctx context.Context, service domain.Service) error {
	for _, parameter := range service.Parameters {
		if parameter.Type != domain.ParameterMaterialReference || parameter.DefaultValue == "" {
			continue
		}
		if s.material == nil {
			return fmt.Errorf("material-reference default cannot be checked")
		}
		material, err := s.material.Get(ctx, parameter.DefaultValue)
		if err != nil {
			return fmt.Errorf("parameter %q: material reference: %w", parameter.Key, err)
		}
		if !material.Active {
			return fmt.Errorf("parameter %q: material reference must be active", parameter.Key)
		}
	}
	for _, component := range service.Components {
		if component.Type == domain.CostMaterial {
			if s.material == nil {
				return fmt.Errorf("material component reference cannot be checked")
			}
			material, err := s.material.Get(ctx, component.ReferenceID)
			if err != nil {
				return fmt.Errorf("component %q: material reference: %w", component.Name, err)
			}
			if !material.Active {
				return fmt.Errorf("component %q: material reference must be active", component.Name)
			}
		}
		if component.Type == domain.CostMachine {
			if s.machine == nil {
				return fmt.Errorf("machine component reference cannot be checked")
			}
			machine, err := s.machine.GetMachine(ctx, component.ReferenceID)
			if err != nil {
				return fmt.Errorf("component %q: machine reference: %w", component.Name, err)
			}
			if !machine.Active {
				return fmt.Errorf("component %q: machine reference must be active", component.Name)
			}
		}
	}
	return nil
}

func (s *ServicesService) parseDraft(ctx context.Context, input ServiceInput, serviceID string, existing []domain.ServiceParameter) (domain.ServiceDraft, error) {
	existingIDs := make(map[string]struct{}, len(existing))
	for _, parameter := range existing {
		existingIDs[parameter.ID] = struct{}{}
	}
	parameters := make([]domain.ServiceParameterDraft, 0, len(input.Parameters))
	for _, parameterInput := range input.Parameters {
		if parameterInput.ID != "" {
			if _, exists := existingIDs[parameterInput.ID]; !exists && len(existing) > 0 {
				// A client-side draft parameter has no persisted identity yet.
				parameterInput.ID = ""
			}
		}
		parameter, err := s.parseParameter(ctx, parameterInput, serviceID, existing)
		if err != nil {
			return domain.ServiceDraft{}, err
		}
		parameters = append(parameters, parameter)
	}
	components := make([]domain.ServiceCostComponentDraft, 0, len(input.Components))
	for _, input := range input.Components {
		component, err := s.parseComponent(input)
		if err != nil {
			return domain.ServiceDraft{}, err
		}
		components = append(components, component)
	}
	return domain.ServiceDraft{Name: input.Name, Code: input.Code, Category: input.Category, Description: input.Description, Parameters: parameters, Components: components}, nil
}

func (s *ServicesService) parseComponent(input CostComponentInput) (domain.ServiceCostComponentDraft, error) {
	id := strings.TrimSpace(input.ID)
	if id == "" {
		var err error
		id, err = newID("CMP-")
		if err != nil {
			return domain.ServiceCostComponentDraft{}, fmt.Errorf("create component id: %w", err)
		}
	}
	multiplier := domain.Quantity(domain.QuantityScale)
	if strings.TrimSpace(input.Multiplier) != "" {
		parsed, err := domain.ParseQuantity(input.Multiplier)
		if err != nil {
			return domain.ServiceCostComponentDraft{}, domain.ValidationError{Field: "multiplier", Message: "must be a positive decimal with at most six fractional digits"}
		}
		multiplier = parsed
	}
	percentage := domain.Quantity(0)
	if strings.TrimSpace(input.Percentage) != "" {
		parsed, err := domain.ParseQuantity(input.Percentage)
		if err != nil {
			return domain.ServiceCostComponentDraft{}, domain.ValidationError{Field: "percentage", Message: "must be a non-negative decimal with at most six fractional digits"}
		}
		percentage = parsed
	}
	usageMode := domain.UsageMode(strings.ToLower(strings.TrimSpace(input.UsageMode)))
	if usageMode == "" {
		usageMode = domain.UsageFixed
	}
	return domain.ServiceCostComponentDraft{ID: id, Name: input.Name, Type: domain.CostComponentType(strings.ToLower(strings.TrimSpace(input.Type))), ReferenceID: input.ReferenceID, UsageMode: usageMode, ParameterKey: input.ParameterKey, Multiplier: multiplier, RateRial: input.RateRial, Percentage: percentage, RateBasis: input.RateBasis, Enabled: input.Enabled, Notes: input.Notes}, nil
}

func componentFromDraft(serviceID string, draft domain.ServiceCostComponentDraft, position int, now time.Time) domain.ServiceCostComponent {
	usageMode := domain.UsageMode(strings.ToLower(strings.TrimSpace(string(draft.UsageMode))))
	if usageMode == "" {
		usageMode = domain.UsageFixed
	}
	return domain.ServiceCostComponent{ID: draft.ID, ServiceID: serviceID, Name: strings.TrimSpace(draft.Name), Type: domain.CostComponentType(strings.ToLower(strings.TrimSpace(string(draft.Type)))), ReferenceID: strings.TrimSpace(draft.ReferenceID), UsageMode: usageMode, ParameterKey: strings.TrimSpace(draft.ParameterKey), Multiplier: draft.Multiplier, RateRial: draft.RateRial, Percentage: draft.Percentage, RateBasis: strings.TrimSpace(draft.RateBasis), Enabled: draft.Enabled, Position: position, Notes: strings.TrimSpace(draft.Notes), CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
}

func (s *ServicesService) parseParameter(_ context.Context, input ParameterInput, _ string, _ []domain.ServiceParameter) (domain.ServiceParameterDraft, error) {
	id := strings.TrimSpace(input.ID)
	if id == "" {
		var err error
		id, err = newID("PAR-")
		if err != nil {
			return domain.ServiceParameterDraft{}, fmt.Errorf("create parameter id: %w", err)
		}
	}
	minimum, err := parseOptionalQuantity(input.MinValue)
	if err != nil {
		return domain.ServiceParameterDraft{}, fmt.Errorf("minValue: %w", err)
	}
	maximum, err := parseOptionalQuantity(input.MaxValue)
	if err != nil {
		return domain.ServiceParameterDraft{}, fmt.Errorf("maxValue: %w", err)
	}
	defaultValue := strings.TrimSpace(input.DefaultValue)
	options := append([]string(nil), input.Options...)
	typeName := domain.ParameterType(strings.ToLower(strings.TrimSpace(input.Type)))
	switch typeName {
	case domain.ParameterInteger:
		if defaultValue != "" {
			parsed, parseErr := strconv.ParseInt(defaultValue, 10, 64)
			if parseErr != nil || parsed < 0 {
				return domain.ServiceParameterDraft{}, domain.ValidationError{Field: "defaultValue", Message: "must be a non-negative whole number"}
			}
			defaultValue = strconv.FormatInt(parsed, 10)
		}
		minimum, maximum, err = normalizeIntegerBounds(minimum, maximum)
		if err != nil {
			return domain.ServiceParameterDraft{}, err
		}
	case domain.ParameterDecimal:
		if defaultValue != "" {
			parsed, parseErr := domain.ParseQuantity(defaultValue)
			if parseErr != nil {
				return domain.ServiceParameterDraft{}, domain.ValidationError{Field: "defaultValue", Message: "must be a non-negative decimal with at most six fractional digits"}
			}
			defaultValue = parsed.String()
		}
	case domain.ParameterBoolean:
		if defaultValue != "" && defaultValue != "true" && defaultValue != "false" {
			return domain.ServiceParameterDraft{}, domain.ValidationError{Field: "defaultValue", Message: "must be true or false"}
		}
		minimum, maximum = nil, nil
	case domain.ParameterChoice, domain.ParameterMaterialReference:
		minimum, maximum = nil, nil
	default:
		options = nil
	}
	if typeName != domain.ParameterChoice {
		options = nil
	}
	return domain.ServiceParameterDraft{
		ID: id, Key: input.Key, Label: input.Label, Type: typeName, Required: input.Required,
		DefaultValue: defaultValue, Options: options, MinValue: minimum, MaxValue: maximum, Unit: input.Unit,
	}, nil
}

func parseOptionalQuantity(value *string) (*domain.Quantity, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	quantity, err := domain.ParseQuantity(*value)
	if err != nil {
		return nil, err
	}
	return &quantity, nil
}

func normalizeIntegerBounds(minimum, maximum *domain.Quantity) (*domain.Quantity, *domain.Quantity, error) {
	if minimum != nil && *minimum%domain.QuantityScale != 0 {
		return nil, nil, domain.ValidationError{Field: "minValue", Message: "must be a whole number for integer parameters"}
	}
	if maximum != nil && *maximum%domain.QuantityScale != 0 {
		return nil, nil, domain.ValidationError{Field: "maxValue", Message: "must be a whole number for integer parameters"}
	}
	return minimum, maximum, nil
}

func parameterFromDraft(serviceID string, draft domain.ServiceParameterDraft, position int, now time.Time) domain.ServiceParameter {
	return domain.ServiceParameter{
		ID: draft.ID, ServiceID: serviceID, Key: strings.TrimSpace(draft.Key), Label: strings.TrimSpace(draft.Label),
		Type: domain.ParameterType(strings.ToLower(strings.TrimSpace(string(draft.Type)))), Required: draft.Required,
		Position: position, DefaultValue: strings.TrimSpace(draft.DefaultValue), Options: append([]string(nil), draft.Options...),
		MinValue: draft.MinValue, MaxValue: draft.MaxValue, Unit: strings.TrimSpace(draft.Unit), Active: true,
		CreatedAt: now.UTC(), UpdatedAt: now.UTC(),
	}
}

func serviceView(service domain.Service) ServiceView {
	parameters := make([]ParameterView, 0, len(service.Parameters))
	for _, parameter := range service.Parameters {
		view := ParameterView{ID: parameter.ID, Key: parameter.Key, Label: parameter.Label, Type: string(parameter.Type), Required: parameter.Required, Position: parameter.Position, DefaultValue: parameter.DefaultValue, Options: append([]string(nil), parameter.Options...), Unit: parameter.Unit, Active: parameter.Active}
		if parameter.MinValue != nil {
			value := parameter.MinValue.String()
			view.MinValue = &value
		}
		if parameter.MaxValue != nil {
			value := parameter.MaxValue.String()
			view.MaxValue = &value
		}
		parameters = append(parameters, view)
	}
	components := make([]CostComponentView, 0, len(service.Components))
	for _, component := range service.Components {
		components = append(components, CostComponentView{ID: component.ID, Name: component.Name, Type: string(component.Type), ReferenceID: component.ReferenceID, UsageMode: string(component.UsageMode), ParameterKey: component.ParameterKey, Multiplier: component.Multiplier.String(), RateRial: component.RateRial, Percentage: component.Percentage.String(), RateBasis: component.RateBasis, Enabled: component.Enabled, Position: component.Position, Notes: component.Notes})
	}
	return ServiceView{ID: service.ID, Name: service.Name, Code: service.Code, Category: service.Category, Description: service.Description, Active: service.Active, CreatedAt: service.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: service.UpdatedAt.UTC().Format(time.RFC3339Nano), Parameters: parameters, Components: components}
}

func newID(prefix string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(bytes), nil
}
