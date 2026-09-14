package main

import (
	"context"
	"fmt"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

// MaterialInput is the Wails transport shape. Quantities are decimal strings
// at the UI boundary; the application layer converts them to fixed-scale ints.
type MaterialInput struct {
	Name                string                   `json:"name"`
	SKU                 string                   `json:"sku"`
	Category            string                   `json:"category"`
	Kind                string                   `json:"kind"`
	Attributes          []MaterialAttributeInput `json:"attributes"`
	PurchaseUnit        string                   `json:"purchaseUnit"`
	ConsumptionUnit     string                   `json:"consumptionUnit"`
	ConversionFactor    string                   `json:"conversionFactor"`
	PhysicalStock       string                   `json:"physicalStock"`
	ReorderLevel        string                   `json:"reorderLevel"`
	AverageUnitCostRial int64                    `json:"averageUnitCostRial"`
	PreferredSupplier   string                   `json:"preferredSupplier"`
	Notes               string                   `json:"notes"`
}

type MaterialAttributeInput struct {
	Key          string `json:"key"`
	ValueType    string `json:"valueType"`
	DecimalValue string `json:"decimalValue"`
	IntegerValue int64  `json:"integerValue"`
	EnumCode     string `json:"enumCode"`
	TextValue    string `json:"textValue"`
	BooleanValue bool   `json:"booleanValue"`
}

type MaterialAttributeDTO struct {
	Key          string `json:"key"`
	ValueType    string `json:"valueType"`
	DecimalValue string `json:"decimalValue"`
	IntegerValue int64  `json:"integerValue"`
	EnumCode     string `json:"enumCode"`
	TextValue    string `json:"textValue"`
	BooleanValue bool   `json:"booleanValue"`
}

type MaterialDTO struct {
	ID                          string                 `json:"id"`
	Name                        string                 `json:"name"`
	SKU                         string                 `json:"sku"`
	Category                    string                 `json:"category"`
	Kind                        string                 `json:"kind"`
	Attributes                  []MaterialAttributeDTO `json:"attributes"`
	PurchaseUnit                string                 `json:"purchaseUnit"`
	ConsumptionUnit             string                 `json:"consumptionUnit"`
	ConversionFactor            string                 `json:"conversionFactor"`
	PhysicalStock               string                 `json:"physicalStock"`
	ReservedStock               string                 `json:"reservedStock"`
	AvailableStock              string                 `json:"availableStock"`
	ReorderLevel                string                 `json:"reorderLevel"`
	AverageUnitCostRial         int64                  `json:"averageUnitCostRial"`
	HighestPurchaseUnitCostRial int64                  `json:"highestPurchaseUnitCostRial"`
	InventoryValueRial          int64                  `json:"inventoryValueRial"`
	PreferredSupplier           string                 `json:"preferredSupplier"`
	Notes                       string                 `json:"notes"`
	Active                      bool                   `json:"active"`
	LowStock                    bool                   `json:"lowStock"`
	CreatedAt                   string                 `json:"createdAt"`
	UpdatedAt                   string                 `json:"updatedAt"`
}

type MaterialAttributeDefinitionDTO struct {
	Key             string                               `json:"key"`
	Label           string                               `json:"label"`
	ValueType       string                               `json:"valueType"`
	Unit            string                               `json:"unit"`
	ApplicableKinds []string                             `json:"applicableKinds"`
	EnumOptions     []MaterialAttributeEnumOptionDTO `json:"enumOptions"`
	Active          bool                                 `json:"active"`
	Position        int                                  `json:"position"`
}

type MaterialAttributeEnumOptionDTO struct {
	Code string `json:"code"`
	Label string `json:"label"`
	Active bool `json:"active"`
	Position int `json:"position"`
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

func (a *App) ListMaterialAttributeDefinitions() ([]MaterialAttributeDefinitionDTO, error) {
	service, err := a.materialService()
	if err != nil {
		return nil, err
	}
	definitions, err := service.ListAttributeDefinitions(a.materialContext())
	if err != nil {
		return nil, err
	}
	result := make([]MaterialAttributeDefinitionDTO, 0, len(definitions))
	for _, definition := range definitions {
		dto := MaterialAttributeDefinitionDTO{Key: definition.Key, Label: definition.Label, ValueType: definition.ValueType, Unit: definition.Unit, ApplicableKinds: definition.ApplicableKinds, Active: definition.Active, Position: definition.Position}
		for _, option := range definition.EnumOptions { dto.EnumOptions = append(dto.EnumOptions, MaterialAttributeEnumOptionDTO{Code: option.Code, Label: option.Label, Active: option.Active, Position: option.Position}) }
		result = append(result, dto)
	}
	return result, nil
}

func (a *App) CreateMaterial(input MaterialInput) (MaterialDTO, error) {
	service, err := a.materialService()
	if err != nil {
		return MaterialDTO{}, err
	}
	converted, err := applicationInput(input)
	if err != nil {
		return MaterialDTO{}, err
	}
	view, err := service.Create(a.materialContext(), converted)
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
	converted, err := applicationInput(input)
	if err != nil {
		return MaterialDTO{}, err
	}
	view, err := service.Update(a.materialContext(), id, converted)
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

func (a *App) DeleteMaterial(id string) error {
	service, err := a.materialService()
	if err != nil {
		return err
	}
	return service.Delete(a.materialContext(), id)
}

func materialDTO(view application.MaterialView) MaterialDTO {
	dto := MaterialDTO{
		ID: view.ID, Name: view.Name, SKU: view.SKU, Category: view.Category,
		Kind:         view.Kind,
		PurchaseUnit: view.PurchaseUnit, ConsumptionUnit: view.ConsumptionUnit,
		ConversionFactor: view.ConversionFactor, PhysicalStock: view.PhysicalStock,
		ReservedStock: view.ReservedStock, AvailableStock: view.AvailableStock,
		ReorderLevel: view.ReorderLevel, AverageUnitCostRial: view.AverageUnitCostRial, HighestPurchaseUnitCostRial: view.HighestPurchaseUnitCostRial, InventoryValueRial: view.InventoryValueRial,
		PreferredSupplier: view.PreferredSupplier, Notes: view.Notes, Active: view.Active,
		LowStock: view.LowStock, CreatedAt: view.CreatedAt, UpdatedAt: view.UpdatedAt,
	}
	for _, value := range view.Attributes {
		dto.Attributes = append(dto.Attributes, materialAttributeDTO(value))
	}
	return dto
}

func applicationInput(input MaterialInput) (application.MaterialInput, error) {
	attributes := make([]domain.MaterialAttributeValue, 0, len(input.Attributes))
	for _, raw := range input.Attributes {
		value := domain.MaterialAttributeValue{Key: raw.Key, ValueType: domain.MaterialAttributeValueType(raw.ValueType), IntegerValue: raw.IntegerValue, EnumCode: raw.EnumCode, TextValue: raw.TextValue, BooleanValue: raw.BooleanValue}
		if value.ValueType == domain.MaterialAttributeDecimal {
			parsed, err := domain.ParseQuantity(raw.DecimalValue)
			if err != nil {
				return application.MaterialInput{}, fmt.Errorf("attribute %q decimalValue: %w", raw.Key, err)
			}
			value.DecimalValue = parsed
		}
		attributes = append(attributes, value)
	}
	return application.MaterialInput{
		Name: input.Name, SKU: input.SKU, Category: input.Category,
		Kind: input.Kind, Attributes: attributes,
		PurchaseUnit: input.PurchaseUnit, ConsumptionUnit: input.ConsumptionUnit,
		ConversionFactor: input.ConversionFactor, PhysicalStock: input.PhysicalStock,
		ReorderLevel: input.ReorderLevel, AverageUnitCostRial: input.AverageUnitCostRial,
		PreferredSupplier: input.PreferredSupplier, Notes: input.Notes,
	}, nil
}

func materialAttributeDTO(value domain.MaterialAttributeValue) MaterialAttributeDTO {
	return MaterialAttributeDTO{Key: value.Key, ValueType: string(value.ValueType), DecimalValue: value.DecimalValue.String(), IntegerValue: value.IntegerValue, EnumCode: value.EnumCode, TextValue: value.TextValue, BooleanValue: value.BooleanValue}
}
