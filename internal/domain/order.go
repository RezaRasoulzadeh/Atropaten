package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var ErrOrderNotFound = errors.New("order not found")

type CommercialStatus string

const (
	CommercialDraft     CommercialStatus = "Draft"
	CommercialConfirmed CommercialStatus = "Confirmed"
	CommercialClosed    CommercialStatus = "Closed"
	CommercialCancelled CommercialStatus = "Cancelled"
)

type FulfillmentStatus string

const (
	FulfillmentPending      FulfillmentStatus = "Pending"
	FulfillmentInProduction FulfillmentStatus = "In Production"
	FulfillmentReady        FulfillmentStatus = "Ready"
	FulfillmentDelivered    FulfillmentStatus = "Delivered"
)

type PaymentStatus string

const (
	PaymentUnpaid        PaymentStatus = "Unpaid"
	PaymentPartiallyPaid PaymentStatus = "Partially Paid"
	PaymentPaid          PaymentStatus = "Paid"
)

type Priority string

const (
	PriorityUrgent Priority = "Urgent"
	PriorityHigh   Priority = "High"
	PriorityNormal Priority = "Normal"
	PriorityLow    Priority = "Low"
)

type Order struct {
	ID, OrderNumber, CustomerID, CustomerNameSnapshot        string
	CustomerPhoneSnapshot, Notes                             string
	CreatedAt, UpdatedAt                                     time.Time
	PromisedAt                                               *time.Time
	Priority                                                 Priority
	CommercialStatus                                         CommercialStatus
	FulfillmentStatus                                        FulfillmentStatus
	PaymentStatus                                            PaymentStatus
	SubtotalRial, DiscountRial, TotalRial, EstimatedCostRial int64
	Items                                                    []OrderItem
}

type OrderItem struct {
	ID, OrderID                                                    string
	Position                                                       int
	ServiceID, ServiceNameSnapshot, ServiceCodeSnapshot            string
	Quantity                                                       Quantity
	QuantityUnit                                                   string
	ResolvedParametersJSON, CostBreakdownJSON, PricingSnapshotJSON string
	EstimatedCostRial, SuggestedPriceRial, SellingPriceRial        int64
	Notes                                                          string
}

func NewOrder(id string, customerID string, now time.Time) Order {
	return Order{ID: strings.TrimSpace(id), CustomerID: strings.TrimSpace(customerID), CreatedAt: now.UTC(), UpdatedAt: now.UTC(), Priority: PriorityNormal, CommercialStatus: CommercialDraft, FulfillmentStatus: FulfillmentPending, PaymentStatus: PaymentUnpaid}
}

func (o Order) Validate() error {
	if o.ID == "" || o.OrderNumber == "" {
		return validationError("order", "id and order number are required")
	}
	if o.CreatedAt.IsZero() || o.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	if !validPriority(o.Priority) || !validCommercial(o.CommercialStatus) || !validFulfillment(o.FulfillmentStatus) || !validPayment(o.PaymentStatus) {
		return validationError("status", "contains an unsupported value")
	}
	if o.DiscountRial < 0 || o.SubtotalRial < 0 || o.TotalRial < 0 || o.EstimatedCostRial < 0 {
		return validationError("money", "cannot be negative")
	}
	if o.TotalRial != o.SubtotalRial-o.DiscountRial {
		return validationError("totalRial", "must equal subtotal less discount")
	}
	for i, item := range o.Items {
		if item.ID == "" || item.OrderID != o.ID || item.Position != i {
			return validationError("items", "must have contiguous positions and belong to the order")
		}
		if item.ServiceNameSnapshot == "" || item.SellingPriceRial < 0 || item.EstimatedCostRial < 0 || item.SuggestedPriceRial < 0 {
			return validationError("item", "contains invalid snapshot or money")
		}
	}
	return nil
}

func (o *Order) RecalculateTotals() error {
	var subtotal, estimated big.Int
	for _, item := range o.Items {
		subtotal.Add(&subtotal, big.NewInt(item.SellingPriceRial))
		estimated.Add(&estimated, big.NewInt(item.EstimatedCostRial))
	}
	if !subtotal.IsInt64() || !estimated.IsInt64() {
		return fmt.Errorf("order totals exceed Rial range")
	}
	o.SubtotalRial, o.EstimatedCostRial = subtotal.Int64(), estimated.Int64()
	if o.DiscountRial < 0 || o.DiscountRial > o.SubtotalRial {
		return fmt.Errorf("discount cannot exceed subtotal")
	}
	o.TotalRial = o.SubtotalRial - o.DiscountRial
	return nil
}

func ValidCommercialTransition(from, to CommercialStatus) bool {
	if from == to {
		return true
	}
	if from == CommercialDraft {
		return to == CommercialConfirmed || to == CommercialCancelled
	}
	if from == CommercialConfirmed {
		return to == CommercialClosed || to == CommercialCancelled
	}
	return false
}
func ValidFulfillmentTransition(from, to FulfillmentStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case FulfillmentPending:
		return to == FulfillmentInProduction || to == FulfillmentReady
	case FulfillmentInProduction:
		return to == FulfillmentReady
	case FulfillmentReady:
		return to == FulfillmentDelivered
	}
	return false
}
func ValidPriority(v Priority) bool {
	return v == PriorityUrgent || v == PriorityHigh || v == PriorityNormal || v == PriorityLow
}
func validPriority(v Priority) bool { return ValidPriority(v) }
func validCommercial(v CommercialStatus) bool {
	return v == CommercialDraft || v == CommercialConfirmed || v == CommercialClosed || v == CommercialCancelled
}
func validFulfillment(v FulfillmentStatus) bool {
	return v == FulfillmentPending || v == FulfillmentInProduction || v == FulfillmentReady || v == FulfillmentDelivered
}
func validPayment(v PaymentStatus) bool {
	return v == PaymentUnpaid || v == PaymentPartiallyPaid || v == PaymentPaid
}
