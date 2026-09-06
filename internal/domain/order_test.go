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
	if !ValidCommercialTransition(CommercialDraft, CommercialConfirmed) || ValidCommercialTransition(CommercialClosed, CommercialDraft) {
		t.Fatal("commercial transition rules are incorrect")
	}
	if !ValidFulfillmentTransition(FulfillmentPending, FulfillmentInProduction) || ValidFulfillmentTransition(FulfillmentDelivered, FulfillmentPending) {
		t.Fatal("fulfillment transition rules are incorrect")
	}
}
