package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrReservationNotFound        = errors.New("reservation not found")
	ErrProductionJobNotFound      = errors.New("production job not found")
	ErrReservationExceeded        = errors.New("reservation exceeds available stock")
	ErrProductionNotEditable      = errors.New("production job history is protected")
	ErrProductionTransition       = errors.New("invalid production status transition")
	ErrProductionHistoryProtected = errors.New("production history cannot be deleted")
	ErrConsumptionNotFound        = errors.New("production consumption not found")
)

const (
	ReservationActive    = "active"
	ReservationReleased  = "released"
	ReservationConsumed  = "consumed"
	ReservationCancelled = "cancelled"
	ProductionPending    = "Pending"
	ProductionReady      = "Ready"
	ProductionInProgress = "In Progress"
	ProductionPaused     = "Paused"
	ProductionCompleted  = "Completed"
	ProductionCancelled  = "Cancelled"
	ProductionFailed     = "Failed"
)

type InventoryReservation struct {
	ID, MaterialID, OrderID, OrderItemID, ProductionJobID string
	Quantity                                              Quantity
	Status                                                string
	CreatedAt, UpdatedAt                                  time.Time
}

type ProductionJob struct {
	ID, JobNumber, OrderID, OrderItemID                                                                                        string
	ServiceNameSnapshot                                                                                                        string
	Quantity                                                                                                                   Quantity
	QuantityUnit, AssignedMachineID, Status, Priority, Notes                                                                   string
	PlannedAt, StartedAt, CompletedAt                                                                                          *time.Time
	EstimatedCostRial, ActualMaterialCostRial, ActualWasteCostRial, ActualOutsourcedCostRial, OutsourceQuotedCostRial          int64
	OutsourceSupplierID, OutsourceDescription, OutsourceSentAt, OutsourceExpectedReturnAt, OutsourceReceivedAt, OutsourceNotes string
	CreatedAt, UpdatedAt                                                                                                       time.Time
}

type ProductionConsumption struct {
	ID, ProductionJobID, MaterialID, IdempotencyKey string
	ConsumedQuantity, WasteQuantity                 Quantity
	UnitCostRial, MaterialCostRial, WasteCostRial   int64
	Notes                                           string
	CreatedAt                                       time.Time
}

func ValidProductionTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case ProductionPending:
		return to == ProductionReady || to == ProductionInProgress || to == ProductionCancelled
	case ProductionReady:
		return to == ProductionInProgress || to == ProductionCancelled
	case ProductionInProgress:
		return to == ProductionPaused || to == ProductionCompleted || to == ProductionCancelled || to == ProductionFailed
	case ProductionPaused:
		return to == ProductionInProgress || to == ProductionCancelled
	case ProductionFailed:
		return to == ProductionInProgress || to == ProductionCancelled
	}
	return false
}

func ValidProductionStatus(status string) bool {
	switch status {
	case ProductionPending, ProductionReady, ProductionInProgress, ProductionPaused, ProductionCompleted, ProductionCancelled, ProductionFailed:
		return true
	default:
		return false
	}
}

func (j ProductionJob) Validate() error {
	if strings.TrimSpace(j.ID) == "" || strings.TrimSpace(j.OrderID) == "" || strings.TrimSpace(j.OrderItemID) == "" {
		return validationError("job", "id, order, and order item are required")
	}
	if j.Quantity <= 0 {
		return validationError("quantity", "must be positive")
	}
	if !ValidProductionStatus(j.Status) {
		return validationError("status", "is unsupported")
	}
	if j.EstimatedCostRial < 0 || j.ActualMaterialCostRial < 0 || j.ActualWasteCostRial < 0 || j.ActualOutsourcedCostRial < 0 || j.OutsourceQuotedCostRial < 0 {
		return validationError("cost", "cannot be negative")
	}
	if j.CreatedAt.IsZero() || j.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	return nil
}

func (r InventoryReservation) Validate() error {
	if r.ID == "" || r.MaterialID == "" {
		return validationError("reservation", "id and material are required")
	}
	if r.Quantity <= 0 {
		return validationError("quantity", "must be positive")
	}
	if r.Status != ReservationActive && r.Status != ReservationReleased && r.Status != ReservationConsumed && r.Status != ReservationCancelled {
		return fmt.Errorf("unsupported reservation status")
	}
	return nil
}
