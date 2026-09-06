package domain

import (
	"testing"
	"time"
)

func TestProductionTransitionsKeepCommercialConcernsSeparate(t *testing.T) {
	valid := []struct{ from, to string }{
		{ProductionPending, ProductionReady}, {ProductionReady, ProductionInProgress},
		{ProductionInProgress, ProductionPaused}, {ProductionPaused, ProductionInProgress},
		{ProductionInProgress, ProductionCompleted}, {ProductionPending, ProductionCancelled},
	}
	for _, tt := range valid {
		if !ValidProductionTransition(tt.from, tt.to) { t.Errorf("%s -> %s should be valid", tt.from, tt.to) }
	}
	if ValidProductionTransition(ProductionCompleted, ProductionInProgress) || ValidProductionTransition(ProductionPending, "Paid") {
		t.Fatal("invalid production transition accepted")
	}
	if ValidProductionStatus("Paid") { t.Fatal("commercial status accepted as production status") }
}

func TestReservationAndProductionValidation(t *testing.T) {
	now := time.Now().UTC()
	job := ProductionJob{ID:"JOB-1",OrderID:"ORD-1",OrderItemID:"ITEM-1",Quantity:1,QuantityUnit:"sheet",Status:ProductionPending,Priority:string(PriorityNormal),CreatedAt:now,UpdatedAt:now}
	if err := job.Validate(); err != nil { t.Fatal(err) }
	job.Status = "Paid"
	if err := job.Validate(); err == nil { t.Fatal("unsupported production status accepted") }
	r := InventoryReservation{ID:"RES-1",MaterialID:"MAT-1",Quantity:1,Status:ReservationActive}
	if err := r.Validate(); err != nil { t.Fatal(err) }
	r.Quantity = 0
	if err := r.Validate(); err == nil { t.Fatal("zero reservation accepted") }
}
