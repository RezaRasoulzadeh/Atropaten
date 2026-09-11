package main

import (
	"fmt"

	"Atropaten/internal/application"
)

type MachineInput struct {
	Name          string             `json:"name"`
	Code          string             `json:"code"`
	Category      string             `json:"category"`
	ImagePath     string             `json:"imagePath"`
	RateBasis     string             `json:"rateBasis"`
	RateRial      int64              `json:"rateRial"`
	SetupCostRial int64              `json:"setupCostRial"`
	Notes         string             `json:"notes"`
	Rates         []MachineRateInput `json:"rates"`
}

type MachineRateInput struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	SelectorValue string `json:"selectorValue"`
	RateBasis     string `json:"rateBasis"`
	RateRial      int64  `json:"rateRial"`
	SetupCostRial int64  `json:"setupCostRial"`
	Active        bool   `json:"active"`
}

type MachineDTO struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Code          string           `json:"code"`
	Category      string           `json:"category"`
	ImagePath     string           `json:"imagePath"`
	RateBasis     string           `json:"rateBasis"`
	RateRial      int64            `json:"rateRial"`
	SetupCostRial int64            `json:"setupCostRial"`
	Notes         string           `json:"notes"`
	Active        bool             `json:"active"`
	CreatedAt     string           `json:"createdAt"`
	UpdatedAt     string           `json:"updatedAt"`
	Rates         []MachineRateDTO `json:"rates"`
}

type MachineRateDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	SelectorValue string `json:"selectorValue"`
	RateBasis     string `json:"rateBasis"`
	RateRial      int64  `json:"rateRial"`
	SetupCostRial int64  `json:"setupCostRial"`
	Active        bool   `json:"active"`
}

func (a *App) machineService() (*application.MachinesService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.machines == nil {
		return nil, fmt.Errorf("machines service is not initialized")
	}
	return a.machines, nil
}

func (a *App) ListMachines(includeArchived bool) ([]MachineDTO, error) {
	service, err := a.machineService()
	if err != nil {
		return nil, err
	}
	views, err := service.List(a.materialContext(), includeArchived)
	if err != nil {
		return nil, err
	}
	result := make([]MachineDTO, 0, len(views))
	for _, view := range views {
		result = append(result, machineDTO(view))
	}
	return result, nil
}

func (a *App) GetMachine(id string) (MachineDTO, error) {
	service, err := a.machineService()
	if err != nil {
		return MachineDTO{}, err
	}
	view, err := service.Get(a.materialContext(), id)
	if err != nil {
		return MachineDTO{}, err
	}
	return machineDTO(view), nil
}

func (a *App) CreateMachine(input MachineInput) (MachineDTO, error) {
	service, err := a.machineService()
	if err != nil {
		return MachineDTO{}, err
	}
	view, err := service.Create(a.materialContext(), application.MachineInput{Name: input.Name, Code: input.Code, Category: input.Category, ImagePath: input.ImagePath, RateBasis: input.RateBasis, RateRial: input.RateRial, SetupCostRial: input.SetupCostRial, Notes: input.Notes, Rates: machineRateInputs(input.Rates)})
	if err != nil {
		return MachineDTO{}, err
	}
	return machineDTO(view), nil
}

func (a *App) UpdateMachine(id string, input MachineInput) (MachineDTO, error) {
	service, err := a.machineService()
	if err != nil {
		return MachineDTO{}, err
	}
	view, err := service.Update(a.materialContext(), id, application.MachineInput{Name: input.Name, Code: input.Code, Category: input.Category, ImagePath: input.ImagePath, RateBasis: input.RateBasis, RateRial: input.RateRial, SetupCostRial: input.SetupCostRial, Notes: input.Notes, Rates: machineRateInputs(input.Rates)})
	if err != nil {
		return MachineDTO{}, err
	}
	return machineDTO(view), nil
}

func (a *App) ArchiveMachine(id string) (MachineDTO, error) {
	service, err := a.machineService()
	if err != nil {
		return MachineDTO{}, err
	}
	view, err := service.Archive(a.materialContext(), id)
	if err != nil {
		return MachineDTO{}, err
	}
	return machineDTO(view), nil
}
func (a *App) ReactivateMachine(id string) (MachineDTO, error) {
	service, err := a.machineService()
	if err != nil {
		return MachineDTO{}, err
	}
	view, err := service.Reactivate(a.materialContext(), id)
	if err != nil {
		return MachineDTO{}, err
	}
	return machineDTO(view), nil
}

func (a *App) DeleteMachine(id string) error {
	service, err := a.machineService()
	if err != nil {
		return err
	}
	return service.Delete(a.materialContext(), id)
}

func machineDTO(view application.MachineView) MachineDTO {
	rates := make([]MachineRateDTO, 0, len(view.Rates))
	for _, rate := range view.Rates {
		rates = append(rates, MachineRateDTO{ID: rate.ID, Name: rate.Name, SelectorValue: rate.SelectorValue, RateBasis: rate.RateBasis, RateRial: rate.RateRial, SetupCostRial: rate.SetupCostRial, Active: rate.Active})
	}
	return MachineDTO{ID: view.ID, Name: view.Name, Code: view.Code, Category: view.Category, ImagePath: view.ImagePath, RateBasis: view.RateBasis, RateRial: view.RateRial, SetupCostRial: view.SetupCostRial, Notes: view.Notes, Active: view.Active, CreatedAt: view.CreatedAt, UpdatedAt: view.UpdatedAt, Rates: rates}
}

func machineRateInputs(inputs []MachineRateInput) []application.MachineRateInput {
	rates := make([]application.MachineRateInput, 0, len(inputs))
	for _, rate := range inputs {
		rates = append(rates, application.MachineRateInput{ID: rate.ID, Name: rate.Name, SelectorValue: rate.SelectorValue, RateBasis: rate.RateBasis, RateRial: rate.RateRial, SetupCostRial: rate.SetupCostRial, Active: rate.Active})
	}
	return rates
}
