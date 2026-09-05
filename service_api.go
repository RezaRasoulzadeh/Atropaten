package main

import (
	"fmt"

	"Atropaten/internal/application"
)

type ServiceParameterInput struct {
	ID           string   `json:"id"`
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	DefaultValue string   `json:"defaultValue"`
	Options      []string `json:"options"`
	MinValue     *string  `json:"minValue"`
	MaxValue     *string  `json:"maxValue"`
	Unit         string   `json:"unit"`
}

type ServiceInput struct {
	Name        string                 `json:"name"`
	Code        string                 `json:"code"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Parameters  []ServiceParameterInput `json:"parameters"`
}

type ServiceParameterDTO struct {
	ID           string   `json:"id"`
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	Position     int      `json:"position"`
	DefaultValue string   `json:"defaultValue"`
	Options      []string `json:"options"`
	MinValue     *string  `json:"minValue"`
	MaxValue     *string  `json:"maxValue"`
	Unit         string   `json:"unit"`
	Active       bool     `json:"active"`
}

type ServiceDTO struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Code        string               `json:"code"`
	Category    string               `json:"category"`
	Description string               `json:"description"`
	Active      bool                 `json:"active"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
	Parameters  []ServiceParameterDTO `json:"parameters"`
}

func (a *App) serviceService() (*application.ServicesService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.services == nil {
		return nil, fmt.Errorf("services service is not initialized")
	}
	return a.services, nil
}

func (a *App) ListServices(includeArchived bool) ([]ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return nil, err
	}
	views, err := service.List(a.materialContext(), includeArchived)
	if err != nil {
		return nil, err
	}
	result := make([]ServiceDTO, 0, len(views))
	for _, view := range views {
		result = append(result, serviceDTO(view))
	}
	return result, nil
}

func (a *App) GetService(id string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Get(a.materialContext(), id)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) CreateService(input ServiceInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Create(a.materialContext(), applicationServiceInput(input))
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) UpdateService(id string, input ServiceInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Update(a.materialContext(), id, applicationServiceInput(input))
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) ArchiveService(id string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Archive(a.materialContext(), id)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) ReactivateService(id string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.Reactivate(a.materialContext(), id)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) AddServiceParameter(id string, input ServiceParameterInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.AddParameter(a.materialContext(), id, applicationParameterInput(input))
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) UpdateServiceParameter(id, parameterID string, input ServiceParameterInput) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.UpdateParameter(a.materialContext(), id, parameterID, applicationParameterInput(input))
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) RemoveServiceParameter(id, parameterID string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.RemoveParameter(a.materialContext(), id, parameterID)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func (a *App) ReorderServiceParameters(id string, parameterIDs []string) (ServiceDTO, error) {
	service, err := a.serviceService()
	if err != nil {
		return ServiceDTO{}, err
	}
	view, err := service.ReorderParameters(a.materialContext(), id, parameterIDs)
	if err != nil {
		return ServiceDTO{}, err
	}
	return serviceDTO(view), nil
}

func applicationServiceInput(input ServiceInput) application.ServiceInput {
	parameters := make([]application.ParameterInput, 0, len(input.Parameters))
	for _, parameter := range input.Parameters {
		parameters = append(parameters, applicationParameterInput(parameter))
	}
	return application.ServiceInput{Name: input.Name, Code: input.Code, Category: input.Category, Description: input.Description, Parameters: parameters}
}

func applicationParameterInput(input ServiceParameterInput) application.ParameterInput {
	return application.ParameterInput{ID: input.ID, Key: input.Key, Label: input.Label, Type: input.Type, Required: input.Required, DefaultValue: input.DefaultValue, Options: input.Options, MinValue: input.MinValue, MaxValue: input.MaxValue, Unit: input.Unit}
}

func serviceDTO(view application.ServiceView) ServiceDTO {
	parameters := make([]ServiceParameterDTO, 0, len(view.Parameters))
	for _, parameter := range view.Parameters {
		parameters = append(parameters, ServiceParameterDTO{ID: parameter.ID, Key: parameter.Key, Label: parameter.Label, Type: parameter.Type, Required: parameter.Required, Position: parameter.Position, DefaultValue: parameter.DefaultValue, Options: parameter.Options, MinValue: parameter.MinValue, MaxValue: parameter.MaxValue, Unit: parameter.Unit, Active: parameter.Active})
	}
	return ServiceDTO{ID: view.ID, Name: view.Name, Code: view.Code, Category: view.Category, Description: view.Description, Active: view.Active, CreatedAt: view.CreatedAt, UpdatedAt: view.UpdatedAt, Parameters: parameters}
}
