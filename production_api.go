package main

import (
	"fmt"
	"time"

	"Atropaten/internal/application"
)

type ProductionJobInput struct {
	OrderID           string  `json:"orderId"`
	OrderItemID       string  `json:"orderItemId"`
	Quantity          string  `json:"quantity"`
	QuantityUnit      string  `json:"quantityUnit"`
	AssignedMachineID string  `json:"assignedMachineId"`
	Priority          string  `json:"priority"`
	Notes             string  `json:"notes"`
	PlannedAt         *string `json:"plannedAt"`
}

type ProductionJobDTO struct {
	ID                        string `json:"id"`
	JobNumber                 string `json:"jobNumber"`
	OrderID                   string `json:"orderId"`
	OrderItemID               string `json:"orderItemId"`
	ServiceName               string `json:"serviceName"`
	Quantity                  string `json:"quantity"`
	QuantityUnit              string `json:"quantityUnit"`
	AssignedMachineID         string `json:"assignedMachineId"`
	Status                    string `json:"status"`
	Priority                  string `json:"priority"`
	Notes                     string `json:"notes"`
	PlannedAt                 string `json:"plannedAt"`
	StartedAt                 string `json:"startedAt"`
	CompletedAt               string `json:"completedAt"`
	CreatedAt                 string `json:"createdAt"`
	EstimatedCostRial         int64  `json:"estimatedCostRial"`
	ActualMaterialCostRial    int64  `json:"actualMaterialCostRial"`
	ActualWasteCostRial       int64  `json:"actualWasteCostRial"`
	ActualOutsourcedCostRial  int64  `json:"actualOutsourcedCostRial"`
	ActualTotalCostRial       int64  `json:"actualTotalCostRial"`
	OutsourceQuotedCostRial   int64  `json:"outsourceQuotedCostRial"`
	OutsourceSupplierID       string `json:"outsourceSupplierId"`
	OutsourceDescription      string `json:"outsourceDescription"`
	OutsourceSentAt           string `json:"outsourceSentAt"`
	OutsourceExpectedReturnAt string `json:"outsourceExpectedReturnAt"`
	OutsourceReceivedAt       string `json:"outsourceReceivedAt"`
	OutsourceNotes            string `json:"outsourceNotes"`
}

type InventoryReservationInput struct {
	MaterialID      string `json:"materialId"`
	OrderID         string `json:"orderId"`
	OrderItemID     string `json:"orderItemId"`
	ProductionJobID string `json:"productionJobId"`
	Quantity        string `json:"quantity"`
}
type InventoryReservationDTO struct {
	ID              string `json:"id"`
	MaterialID      string `json:"materialId"`
	OrderID         string `json:"orderId"`
	OrderItemID     string `json:"orderItemId"`
	ProductionJobID string `json:"productionJobId"`
	Quantity        string `json:"quantity"`
	Status          string `json:"status"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}
type ProductionConsumptionInput struct {
	MaterialID       string `json:"materialId"`
	ConsumedQuantity string `json:"consumedQuantity"`
	WasteQuantity    string `json:"wasteQuantity"`
	IdempotencyKey   string `json:"idempotencyKey"`
	Notes            string `json:"notes"`
}
type ProductionConsumptionDTO struct {
	ID               string `json:"id"`
	ProductionJobID  string `json:"productionJobId"`
	MaterialID       string `json:"materialId"`
	IdempotencyKey   string `json:"idempotencyKey"`
	ConsumedQuantity string `json:"consumedQuantity"`
	WasteQuantity    string `json:"wasteQuantity"`
	Notes            string `json:"notes"`
	CreatedAt        string `json:"createdAt"`
	UnitCostRial     int64  `json:"unitCostRial"`
	MaterialCostRial int64  `json:"materialCostRial"`
	WasteCostRial    int64  `json:"wasteCostRial"`
}
type OutsourceInput struct {
	SupplierID       string `json:"supplierId"`
	Description      string `json:"description"`
	SentAt           string `json:"sentAt"`
	ExpectedReturnAt string `json:"expectedReturnAt"`
	ReceivedAt       string `json:"receivedAt"`
	Notes            string `json:"notes"`
	QuotedCostRial   int64  `json:"quotedCostRial"`
	ActualCostRial   int64  `json:"actualCostRial"`
}

func (a *App) productionService() (*application.ProductionService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.production == nil {
		return nil, fmt.Errorf("production service is not initialized")
	}
	return a.production, nil
}

func (a *App) ListProductionJobs(status string) ([]ProductionJobDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return nil, e
	}
	rows, e := s.List(a.materialContext(), status)
	if e != nil {
		return nil, e
	}
	out := make([]ProductionJobDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, productionJobDTO(v))
	}
	return out, nil
}
func (a *App) GetProductionJob(id string) (ProductionJobDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return ProductionJobDTO{}, e
	}
	v, e := s.Get(a.materialContext(), id)
	return productionJobDTO(v), e
}
func (a *App) CreateProductionJob(i ProductionJobInput) (ProductionJobDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return ProductionJobDTO{}, e
	}
	var planned *time.Time
	if i.PlannedAt != nil && *i.PlannedAt != "" {
		t, pe := time.Parse(time.RFC3339Nano, *i.PlannedAt)
		if pe != nil {
			return ProductionJobDTO{}, fmt.Errorf("plannedAt: %w", pe)
		}
		planned = &t
	}
	v, e := s.Create(a.materialContext(), application.ProductionJobInput{OrderID: i.OrderID, OrderItemID: i.OrderItemID, Quantity: i.Quantity, QuantityUnit: i.QuantityUnit, AssignedMachineID: i.AssignedMachineID, Priority: i.Priority, Notes: i.Notes, PlannedAt: planned})
	return productionJobDTO(v), e
}
func (a *App) UpdateProductionJob(id string, i ProductionJobInput) (ProductionJobDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return ProductionJobDTO{}, e
	}
	var planned *time.Time
	if i.PlannedAt != nil && *i.PlannedAt != "" {
		t, pe := time.Parse(time.RFC3339Nano, *i.PlannedAt)
		if pe != nil {
			return ProductionJobDTO{}, pe
		}
		planned = &t
	}
	v, e := s.Update(a.materialContext(), id, application.ProductionJobInput{AssignedMachineID: i.AssignedMachineID, Priority: i.Priority, Notes: i.Notes, PlannedAt: planned})
	return productionJobDTO(v), e
}
func (a *App) UpdateProductionJobStatus(id, status string) (ProductionJobDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return ProductionJobDTO{}, e
	}
	v, e := s.Transition(a.materialContext(), id, status)
	return productionJobDTO(v), e
}
func (a *App) DeleteProductionJob(id string) error {
	s, e := a.productionService()
	if e != nil {
		return e
	}
	return s.Delete(a.materialContext(), id)
}
func (a *App) CreateInventoryReservation(i InventoryReservationInput) (InventoryReservationDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return InventoryReservationDTO{}, e
	}
	v, e := s.Reserve(a.materialContext(), application.ReservationInput{MaterialID: i.MaterialID, OrderID: i.OrderID, OrderItemID: i.OrderItemID, ProductionJobID: i.ProductionJobID, Quantity: i.Quantity})
	return reservationDTO(v), e
}
func (a *App) UpdateInventoryReservation(id, quantity string) (InventoryReservationDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return InventoryReservationDTO{}, e
	}
	v, e := s.UpdateReservation(a.materialContext(), id, quantity)
	return reservationDTO(v), e
}
func (a *App) ReleaseInventoryReservation(id string) error {
	s, e := a.productionService()
	if e != nil {
		return e
	}
	return s.ReleaseReservation(a.materialContext(), id)
}
func (a *App) ListInventoryReservations(materialID, jobID, orderID string) ([]InventoryReservationDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return nil, e
	}
	rows, e := s.Reservations(a.materialContext(), materialID, jobID, orderID)
	if e != nil {
		return nil, e
	}
	out := make([]InventoryReservationDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, reservationDTO(v))
	}
	return out, nil
}
func (a *App) RecordProductionConsumption(jobID string, i ProductionConsumptionInput) (ProductionConsumptionDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return ProductionConsumptionDTO{}, e
	}
	v, e := s.Consume(a.materialContext(), jobID, application.ConsumptionInput{MaterialID: i.MaterialID, ConsumedQuantity: i.ConsumedQuantity, WasteQuantity: i.WasteQuantity, IdempotencyKey: i.IdempotencyKey, Notes: i.Notes})
	return consumptionDTO(v), e
}
func (a *App) ReverseProductionConsumption(id, reason string) error {
	s, e := a.productionService()
	if e != nil {
		return e
	}
	return s.ReverseConsumption(a.materialContext(), id, reason)
}
func (a *App) ListProductionConsumptions(jobID string) ([]ProductionConsumptionDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return nil, e
	}
	rows, e := s.Consumptions(a.materialContext(), jobID)
	if e != nil {
		return nil, e
	}
	out := make([]ProductionConsumptionDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, consumptionDTO(v))
	}
	return out, nil
}
func (a *App) UpdateProductionOutsourcing(id string, i OutsourceInput) (ProductionJobDTO, error) {
	s, e := a.productionService()
	if e != nil {
		return ProductionJobDTO{}, e
	}
	v, e := s.Outsource(a.materialContext(), id, application.OutsourceInput{SupplierID: i.SupplierID, Description: i.Description, SentAt: i.SentAt, ExpectedReturnAt: i.ExpectedReturnAt, ReceivedAt: i.ReceivedAt, Notes: i.Notes, QuotedCostRial: i.QuotedCostRial, ActualCostRial: i.ActualCostRial})
	return productionJobDTO(v), e
}

func productionJobDTO(v application.ProductionJobView) ProductionJobDTO {
	return ProductionJobDTO{ID: v.ID, JobNumber: v.JobNumber, OrderID: v.OrderID, OrderItemID: v.OrderItemID, ServiceName: v.ServiceName, Quantity: v.Quantity, QuantityUnit: v.QuantityUnit, AssignedMachineID: v.AssignedMachineID, Status: v.Status, Priority: v.Priority, Notes: v.Notes, PlannedAt: v.PlannedAt, StartedAt: v.StartedAt, CompletedAt: v.CompletedAt, CreatedAt: v.CreatedAt, EstimatedCostRial: v.EstimatedCostRial, ActualMaterialCostRial: v.ActualMaterialCostRial, ActualWasteCostRial: v.ActualWasteCostRial, ActualTotalCostRial: v.ActualTotalCostRial, OutsourceQuotedCostRial: v.OutsourceQuotedCostRial, OutsourceSupplierID: v.OutsourceSupplierID, OutsourceDescription: v.OutsourceDescription, OutsourceSentAt: v.OutsourceSentAt, OutsourceExpectedReturnAt: v.OutsourceExpectedReturnAt, OutsourceReceivedAt: v.OutsourceReceivedAt, OutsourceNotes: v.OutsourceNotes}
}
func reservationDTO(v application.ReservationView) InventoryReservationDTO {
	return InventoryReservationDTO{ID: v.ID, MaterialID: v.MaterialID, OrderID: v.OrderID, OrderItemID: v.OrderItemID, ProductionJobID: v.ProductionJobID, Quantity: v.Quantity, Status: v.Status, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func consumptionDTO(v application.ConsumptionView) ProductionConsumptionDTO {
	return ProductionConsumptionDTO{ID: v.ID, ProductionJobID: v.ProductionJobID, MaterialID: v.MaterialID, IdempotencyKey: v.IdempotencyKey, ConsumedQuantity: v.ConsumedQuantity, WasteQuantity: v.WasteQuantity, Notes: v.Notes, CreatedAt: v.CreatedAt, UnitCostRial: v.UnitCostRial, MaterialCostRial: v.MaterialCostRial, WasteCostRial: v.WasteCostRial}
}
