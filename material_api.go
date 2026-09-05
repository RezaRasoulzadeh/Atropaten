package main

import (
	"context"
	"fmt"

	"Atropaten/internal/application"
)

// MaterialInput is the Wails transport shape. Quantities are decimal strings
// at the UI boundary; the application layer converts them to fixed-scale ints.
type MaterialInput struct {
	Name                string `json:"name"`
	SKU                 string `json:"sku"`
	Category            string `json:"category"`
	PurchaseUnit        string `json:"purchaseUnit"`
	ConsumptionUnit     string `json:"consumptionUnit"`
	ConversionFactor    string `json:"conversionFactor"`
	PhysicalStock       string `json:"physicalStock"`
	ReorderLevel        string `json:"reorderLevel"`
	AverageUnitCostRial int64  `json:"averageUnitCostRial"`
	PreferredSupplier   string `json:"preferredSupplier"`
	Notes               string `json:"notes"`
}

type MaterialDTO struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	SKU                 string `json:"sku"`
	Category            string `json:"category"`
	PurchaseUnit        string `json:"purchaseUnit"`
	ConsumptionUnit     string `json:"consumptionUnit"`
	ConversionFactor    string `json:"conversionFactor"`
	PhysicalStock       string `json:"physicalStock"`
	ReorderLevel        string `json:"reorderLevel"`
	AverageUnitCostRial int64  `json:"averageUnitCostRial"`
	PreferredSupplier   string `json:"preferredSupplier"`
	Notes               string `json:"notes"`
	Active              bool   `json:"active"`
	LowStock            bool   `json:"lowStock"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

func (a *App) materialService() (*application.MaterialsService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.materials == nil {
		return nil, fmt.Errorf("materials service is not initialized")
	}
	return a.materials, nil
}

func (a *App) materialContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

func (a *App) ListMaterials(includeArchived bool) ([]MaterialDTO, error) {
	service, err := a.materialService()
	if err != nil {
		return nil, err
	}
	views, err := service.List(a.materialContext(), includeArchived)
	if err != nil {
		return nil, err
	}
	result := make([]MaterialDTO, 0, len(views))
	for _, view := range views {
		result = append(result, materialDTO(view))
	}
	return result, nil
}

func (a *App) GetMaterial(id string) (MaterialDTO, error) {
	service, err := a.materialService()
	if err != nil {
		return MaterialDTO{}, err
	}
	view, err := service.Get(a.materialContext(), id)
	if err != nil {
		return MaterialDTO{}, err
	}
	return materialDTO(view), nil
}

func (a *App) CreateMaterial(input MaterialInput) (MaterialDTO, error) {
	service, err := a.materialService()
	if err != nil {
		return MaterialDTO{}, err
	}
	view, err := service.Create(a.materialContext(), applicationInput(input))
	if err != nil {
		return MaterialDTO{}, err
	}
	return materialDTO(view), nil
}

func (a *App) UpdateMaterial(id string, input MaterialInput) (MaterialDTO, error) {
	service, err := a.materialService()
	if err != nil {
		return MaterialDTO{}, err
	}
	view, err := service.Update(a.materialContext(), id, applicationInput(input))
	if err != nil {
		return MaterialDTO{}, err
	}
	return materialDTO(view), nil
}

func (a *App) ArchiveMaterial(id string) (MaterialDTO, error) {
	service, err := a.materialService()
	if err != nil {
		return MaterialDTO{}, err
	}
	view, err := service.Archive(a.materialContext(), id)
	if err != nil {
		return MaterialDTO{}, err
	}
	return materialDTO(view), nil
}

func (a *App) ReactivateMaterial(id string) (MaterialDTO, error) {
	service, err := a.materialService()
	if err != nil {
		return MaterialDTO{}, err
	}
	view, err := service.Reactivate(a.materialContext(), id)
	if err != nil {
		return MaterialDTO{}, err
	}
	return materialDTO(view), nil
}

func materialDTO(view application.MaterialView) MaterialDTO {
	return MaterialDTO{
		ID: view.ID, Name: view.Name, SKU: view.SKU, Category: view.Category,
		PurchaseUnit: view.PurchaseUnit, ConsumptionUnit: view.ConsumptionUnit,
		ConversionFactor: view.ConversionFactor, PhysicalStock: view.PhysicalStock,
		ReorderLevel: view.ReorderLevel, AverageUnitCostRial: view.AverageUnitCostRial,
		PreferredSupplier: view.PreferredSupplier, Notes: view.Notes, Active: view.Active,
		LowStock: view.LowStock, CreatedAt: view.CreatedAt, UpdatedAt: view.UpdatedAt,
	}
}

func applicationInput(input MaterialInput) application.MaterialInput {
	return application.MaterialInput{
		Name: input.Name, SKU: input.SKU, Category: input.Category,
		PurchaseUnit: input.PurchaseUnit, ConsumptionUnit: input.ConsumptionUnit,
		ConversionFactor: input.ConversionFactor, PhysicalStock: input.PhysicalStock,
		ReorderLevel: input.ReorderLevel, AverageUnitCostRial: input.AverageUnitCostRial,
		PreferredSupplier: input.PreferredSupplier, Notes: input.Notes,
	}
}
