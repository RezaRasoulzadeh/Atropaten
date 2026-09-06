package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type ProductionRepository interface {
	ListProductionJobs(context.Context, string) ([]domain.ProductionJob, error)
	GetProductionJob(context.Context, string) (domain.ProductionJob, error)
	CreateProductionJob(context.Context, domain.ProductionJob) error
	UpdateProductionJob(context.Context, domain.ProductionJob) error
	TransitionProductionJob(context.Context, string, string) error
	DeleteProductionJob(context.Context, string) error
	CreateReservation(context.Context, domain.InventoryReservation) error
	UpdateReservation(context.Context, string, domain.Quantity) error
	ReleaseReservation(context.Context, string, string) error
	ListReservations(context.Context, string, string, string) ([]domain.InventoryReservation, error)
	RecordProductionConsumption(context.Context, string, string, string, domain.Quantity, domain.Quantity, string) (domain.ProductionConsumption, error)
	ReverseProductionConsumption(context.Context, string, string) error
	ListProductionConsumptions(context.Context, string) ([]domain.ProductionConsumption, error)
}

type ProductionJobInput struct {
	OrderID, OrderItemID, Quantity, QuantityUnit, AssignedMachineID, Priority, Notes string
	PlannedAt                                                                        *time.Time
}
type ReservationInput struct{ MaterialID, OrderID, OrderItemID, ProductionJobID, Quantity string }
type ConsumptionInput struct{ MaterialID, ConsumedQuantity, WasteQuantity, IdempotencyKey, Notes string }
type OutsourceInput struct {
	SupplierID, Description, SentAt, ExpectedReturnAt, ReceivedAt, Notes string
	QuotedCostRial, ActualCostRial                                       int64
}

type ProductionJobView struct {
	ID, JobNumber, OrderID, OrderItemID, ServiceName, Quantity, QuantityUnit, AssignedMachineID, Status, Priority, Notes, PlannedAt, StartedAt, CompletedAt, CreatedAt string
	EstimatedCostRial, ActualMaterialCostRial, ActualWasteCostRial, ActualOutsourcedCostRial, ActualTotalCostRial, OutsourceQuotedCostRial                             int64
	OutsourceSupplierID, OutsourceDescription, OutsourceSentAt, OutsourceExpectedReturnAt, OutsourceReceivedAt, OutsourceNotes                                         string
}
type ReservationView struct{ ID, MaterialID, OrderID, OrderItemID, ProductionJobID, Quantity, Status, CreatedAt, UpdatedAt string }
type ConsumptionView struct {
	ID, ProductionJobID, MaterialID, IdempotencyKey, ConsumedQuantity, WasteQuantity, Notes, CreatedAt string
	UnitCostRial, MaterialCostRial, WasteCostRial                                                      int64
}

type ProductionService struct {
	repository ProductionRepository
	now        func() time.Time
}

func NewProductionService(r ProductionRepository) *ProductionService {
	return &ProductionService{repository: r, now: time.Now}
}
func (s *ProductionService) List(ctx context.Context, status string) ([]ProductionJobView, error) {
	rows, e := s.repository.ListProductionJobs(ctx, status)
	if e != nil {
		return nil, e
	}
	out := make([]ProductionJobView, 0, len(rows))
	for _, v := range rows {
		out = append(out, productionJobView(v))
	}
	return out, nil
}
func (s *ProductionService) Get(ctx context.Context, id string) (ProductionJobView, error) {
	v, e := s.repository.GetProductionJob(ctx, strings.TrimSpace(id))
	if e != nil {
		return ProductionJobView{}, e
	}
	return productionJobView(v), nil
}
func (s *ProductionService) Create(ctx context.Context, in ProductionJobInput) (ProductionJobView, error) {
	id, e := randomID("JOB-")
	if e != nil {
		return ProductionJobView{}, e
	}
	var q domain.Quantity
	if strings.TrimSpace(in.Quantity) != "" {
		q, e = domain.ParseQuantity(in.Quantity)
		if e != nil {
			return ProductionJobView{}, e
		}
	}
	status := domain.ProductionPending
	if in.Priority == "" {
		in.Priority = string(domain.PriorityNormal)
	}
	j := domain.ProductionJob{ID: id, OrderID: strings.TrimSpace(in.OrderID), OrderItemID: strings.TrimSpace(in.OrderItemID), Quantity: q, QuantityUnit: strings.TrimSpace(in.QuantityUnit), AssignedMachineID: strings.TrimSpace(in.AssignedMachineID), Status: status, Priority: in.Priority, Notes: strings.TrimSpace(in.Notes), PlannedAt: in.PlannedAt, CreatedAt: s.now().UTC(), UpdatedAt: s.now().UTC()}
	if q > 0 {
		if e = j.Validate(); e != nil {
			return ProductionJobView{}, e
		}
	}
	if e = s.repository.CreateProductionJob(ctx, j); e != nil {
		return ProductionJobView{}, e
	}
	return s.Get(ctx, id)
}
func (s *ProductionService) Update(ctx context.Context, id string, in ProductionJobInput) (ProductionJobView, error) {
	j, e := s.repository.GetProductionJob(ctx, id)
	if e != nil {
		return ProductionJobView{}, e
	}
	if in.AssignedMachineID != "" {
		j.AssignedMachineID = in.AssignedMachineID
	}
	if in.Priority != "" {
		j.Priority = in.Priority
	}
	j.Notes = in.Notes
	if in.PlannedAt != nil {
		j.PlannedAt = in.PlannedAt
	}
	if e = s.repository.UpdateProductionJob(ctx, j); e != nil {
		return ProductionJobView{}, e
	}
	return s.Get(ctx, id)
}
func (s *ProductionService) Transition(ctx context.Context, id, status string) (ProductionJobView, error) {
	if !domain.ValidProductionStatus(status) {
		return ProductionJobView{}, domain.ErrProductionTransition
	}
	if e := s.repository.TransitionProductionJob(ctx, id, status); e != nil {
		return ProductionJobView{}, e
	}
	return s.Get(ctx, id)
}
func (s *ProductionService) Delete(ctx context.Context, id string) error {
	return s.repository.DeleteProductionJob(ctx, id)
}
func (s *ProductionService) Reserve(ctx context.Context, in ReservationInput) (ReservationView, error) {
	id, e := randomID("RES-")
	if e != nil {
		return ReservationView{}, e
	}
	q, e := domain.ParseQuantity(in.Quantity)
	if e != nil {
		return ReservationView{}, e
	}
	r := domain.InventoryReservation{ID: id, MaterialID: strings.TrimSpace(in.MaterialID), OrderID: strings.TrimSpace(in.OrderID), OrderItemID: strings.TrimSpace(in.OrderItemID), ProductionJobID: strings.TrimSpace(in.ProductionJobID), Quantity: q, Status: domain.ReservationActive, CreatedAt: s.now().UTC(), UpdatedAt: s.now().UTC()}
	if e = r.Validate(); e != nil {
		return ReservationView{}, e
	}
	if e = s.repository.CreateReservation(ctx, r); e != nil {
		return ReservationView{}, e
	}
	rows, e := s.repository.ListReservations(ctx, "", in.ProductionJobID, in.OrderID)
	if e != nil {
		return ReservationView{}, e
	}
	for _, v := range rows {
		if v.ID == id {
			return reservationView(v), nil
		}
	}
	return ReservationView{}, domain.ErrReservationNotFound
}
func (s *ProductionService) UpdateReservation(ctx context.Context, id, quantity string) (ReservationView, error) {
	q, e := domain.ParseQuantity(quantity)
	if e != nil {
		return ReservationView{}, e
	}
	if e = s.repository.UpdateReservation(ctx, id, q); e != nil {
		return ReservationView{}, e
	}
	rows, e := s.repository.ListReservations(ctx, "", "", "")
	if e != nil {
		return ReservationView{}, e
	}
	for _, v := range rows {
		if v.ID == id {
			return reservationView(v), nil
		}
	}
	return ReservationView{}, domain.ErrReservationNotFound
}
func (s *ProductionService) ReleaseReservation(ctx context.Context, id string) error {
	return s.repository.ReleaseReservation(ctx, id, domain.ReservationReleased)
}
func (s *ProductionService) Reservations(ctx context.Context, materialID, jobID, orderID string) ([]ReservationView, error) {
	rows, e := s.repository.ListReservations(ctx, materialID, jobID, orderID)
	if e != nil {
		return nil, e
	}
	out := make([]ReservationView, 0, len(rows))
	for _, v := range rows {
		out = append(out, reservationView(v))
	}
	return out, nil
}
func (s *ProductionService) Consume(ctx context.Context, jobID string, in ConsumptionInput) (ConsumptionView, error) {
	cons, e := domain.ParseQuantity(in.ConsumedQuantity)
	if e != nil {
		return ConsumptionView{}, e
	}
	waste, e := domain.ParseQuantity(in.WasteQuantity)
	if e != nil {
		return ConsumptionView{}, e
	}
	if in.IdempotencyKey == "" {
		return ConsumptionView{}, fmt.Errorf("idempotency key is required")
	}
	v, e := s.repository.RecordProductionConsumption(ctx, jobID, in.MaterialID, in.IdempotencyKey, cons, waste, in.Notes)
	if e != nil {
		return ConsumptionView{}, e
	}
	return consumptionView(v), nil
}
func (s *ProductionService) ReverseConsumption(ctx context.Context, id, reason string) error {
	return s.repository.ReverseProductionConsumption(ctx, id, reason)
}
func (s *ProductionService) Consumptions(ctx context.Context, jobID string) ([]ConsumptionView, error) {
	rows, e := s.repository.ListProductionConsumptions(ctx, jobID)
	if e != nil {
		return nil, e
	}
	out := make([]ConsumptionView, 0, len(rows))
	for _, v := range rows {
		out = append(out, consumptionView(v))
	}
	return out, nil
}
func (s *ProductionService) Outsource(ctx context.Context, id string, in OutsourceInput) (ProductionJobView, error) {
	j, e := s.repository.GetProductionJob(ctx, id)
	if e != nil {
		return ProductionJobView{}, e
	}
	j.OutsourceSupplierID = in.SupplierID
	j.OutsourceDescription = in.Description
	j.OutsourceSentAt = in.SentAt
	j.OutsourceExpectedReturnAt = in.ExpectedReturnAt
	j.OutsourceReceivedAt = in.ReceivedAt
	j.OutsourceNotes = in.Notes
	j.OutsourceQuotedCostRial = in.QuotedCostRial
	j.ActualOutsourcedCostRial = in.ActualCostRial
	if j.ActualOutsourcedCostRial < 0 || j.OutsourceQuotedCostRial < 0 {
		return ProductionJobView{}, fmt.Errorf("outsourced cost cannot be negative")
	}
	if e = s.repository.UpdateProductionJob(ctx, j); e != nil {
		return ProductionJobView{}, e
	}
	return s.Get(ctx, id)
}
func productionJobView(v domain.ProductionJob) ProductionJobView {
	return ProductionJobView{ID: v.ID, JobNumber: v.JobNumber, OrderID: v.OrderID, OrderItemID: v.OrderItemID, ServiceName: v.ServiceNameSnapshot, Quantity: v.Quantity.String(), QuantityUnit: v.QuantityUnit, AssignedMachineID: v.AssignedMachineID, Status: v.Status, Priority: v.Priority, Notes: v.Notes, PlannedAt: optionalTime(v.PlannedAt), StartedAt: optionalTime(v.StartedAt), CompletedAt: optionalTime(v.CompletedAt), CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), EstimatedCostRial: v.EstimatedCostRial, ActualMaterialCostRial: v.ActualMaterialCostRial, ActualWasteCostRial: v.ActualWasteCostRial, ActualOutsourcedCostRial: v.ActualOutsourcedCostRial, ActualTotalCostRial: v.ActualMaterialCostRial + v.ActualWasteCostRial + v.ActualOutsourcedCostRial, OutsourceQuotedCostRial: v.OutsourceQuotedCostRial, OutsourceSupplierID: v.OutsourceSupplierID, OutsourceDescription: v.OutsourceDescription, OutsourceSentAt: v.OutsourceSentAt, OutsourceExpectedReturnAt: v.OutsourceExpectedReturnAt, OutsourceReceivedAt: v.OutsourceReceivedAt, OutsourceNotes: v.OutsourceNotes}
}
func reservationView(v domain.InventoryReservation) ReservationView {
	return ReservationView{ID: v.ID, MaterialID: v.MaterialID, OrderID: v.OrderID, OrderItemID: v.OrderItemID, ProductionJobID: v.ProductionJobID, Quantity: v.Quantity.String(), Status: v.Status, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func consumptionView(v domain.ProductionConsumption) ConsumptionView {
	return ConsumptionView{ID: v.ID, ProductionJobID: v.ProductionJobID, MaterialID: v.MaterialID, IdempotencyKey: v.IdempotencyKey, ConsumedQuantity: v.ConsumedQuantity.String(), WasteQuantity: v.WasteQuantity.String(), UnitCostRial: v.UnitCostRial, MaterialCostRial: v.MaterialCostRial, WasteCostRial: v.WasteCostRial, Notes: v.Notes, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano)}
}
func optionalTime(v *time.Time) string {
	if v == nil {
		return ""
	}
	return v.UTC().Format(time.RFC3339Nano)
}
