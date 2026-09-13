package domain

import (
	"testing"
	"time"
)

func TestOrderTotalsUseIntegerRialAndDiscountValidation(t *testing.T) {
	o := NewOrder("ORDER-1", "CUS-1", time.Unix(0, 0))
	o.OrderNumber = "ORD-1001"
	o.Items = []OrderItem{{ID: "I-1", OrderID: o.ID, Position: 0, ServiceNameSnapshot: "A", SellingPriceRial: 125000001, EstimatedCostRial: 100000001}, {ID: "I-2", OrderID: o.ID, Position: 1, ServiceNameSnapshot: "B", SellingPriceRial: 2, EstimatedCostRial: 1}}
	o.DiscountRial = 10000000
	if err := o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if o.SubtotalRial != 125000003 || o.TotalRial != 115000003 || o.EstimatedCostRial != 100000002 {
		t.Fatalf("totals lost Rial precision: %+v", o)
	}
	o.DiscountRial = o.SubtotalRial + 1
	if err := o.RecalculateTotals(); err == nil {
		t.Fatal("discount above subtotal was accepted")
	}
}

func TestOrderStateAxesHaveBasicTransitions(t *testing.T) {
	statuses := []CommercialStatus{CommercialDraft, CommercialConfirmed, CommercialClosed, CommercialCancelled}
	for _, from := range statuses {
		for _, to := range statuses {
			if !ValidCommercialTransition(from, to) {
				t.Errorf("valid commercial transition %s -> %s was rejected", from, to)
			}
		}
	}
	for _, status := range statuses {
		if !ValidCommercialTransition(status, status) {
			t.Errorf("same-status assignment %s was rejected", status)
		}
	}
	if ValidCommercialTransition(CommercialStatus("Unknown"), CommercialDraft) {
		t.Error("invalid commercial source status was accepted")
	}
	if ValidCommercialTransition(CommercialDraft, CommercialStatus("Unknown")) {
		t.Error("invalid commercial target status was accepted")
	}
	// Fulfillment has its own operational rules: work can be reset to Pending.
	// This axis is separate from the permissive commercial status contract above.
	if !ValidFulfillmentTransition(FulfillmentPending, FulfillmentInProduction) || !ValidFulfillmentTransition(FulfillmentDelivered, FulfillmentPending) {
		t.Fatal("fulfillment transition rules are incorrect")
	}
}
