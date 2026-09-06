package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var ErrQuoteNotFound = errors.New("quote not found")

type QuoteStatus string

const (
	QuoteDraft     QuoteStatus = "Draft"
	QuoteSent      QuoteStatus = "Sent"
	QuoteAccepted  QuoteStatus = "Accepted"
	QuoteRejected  QuoteStatus = "Rejected"
	QuoteExpired   QuoteStatus = "Expired"
	QuoteConverted QuoteStatus = "Converted"
)

type Quote struct {
	ID, QuoteNumber, CustomerID, CustomerNameSnapshot, CustomerPhoneSnapshot string
	CreatedAt, UpdatedAt                                                     time.Time
	ExpiryDate                                                               *time.Time
	Status                                                                   QuoteStatus
	Notes                                                                    string
	SubtotalRial, DiscountRial, TotalRial, EstimatedCostRial                 int64
	ConvertedOrderID                                                         string
	Items                                                                    []QuoteItem
}

type QuoteItem struct {
	ID, QuoteID                                                    string
	Position                                                       int
	ServiceID, ServiceNameSnapshot, ServiceCodeSnapshot            string
	Quantity                                                       Quantity
	QuantityUnit                                                   string
	ResolvedParametersJSON, CostBreakdownJSON, PricingSnapshotJSON string
	EstimatedCostRial, SuggestedPriceRial, SellingPriceRial        int64
	Notes                                                          string
}

func NewQuote(id, customerID string, now time.Time) Quote {
	return Quote{ID: strings.TrimSpace(id), CustomerID: strings.TrimSpace(customerID), CreatedAt: now.UTC(), UpdatedAt: now.UTC(), Status: QuoteDraft}
}

func (q Quote) Validate() error {
	if q.ID == "" || q.QuoteNumber == "" {
		return validationError("quote", "id and quote number are required")
	}
	if q.CreatedAt.IsZero() || q.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	if !ValidQuoteStatus(q.Status) {
		return validationError("status", "contains an unsupported value")
	}
	if q.DiscountRial < 0 || q.SubtotalRial < 0 || q.TotalRial < 0 || q.EstimatedCostRial < 0 || q.DiscountRial > q.SubtotalRial {
		return validationError("money", "contains an invalid amount")
	}
	if q.TotalRial != q.SubtotalRial-q.DiscountRial {
		return validationError("totalRial", "must equal subtotal less discount")
	}
	if q.Status == QuoteConverted && q.ConvertedOrderID == "" {
		return validationError("convertedOrderID", "is required for converted quotes")
	}
	for i, item := range q.Items {
		if item.ID == "" || item.QuoteID != q.ID || item.Position != i {
			return validationError("items", "must have contiguous positions and belong to the quote")
		}
		if item.ServiceNameSnapshot == "" || item.SellingPriceRial < 0 || item.EstimatedCostRial < 0 || item.SuggestedPriceRial < 0 {
			return validationError("item", "contains invalid snapshot or money")
		}
	}
	return nil
}

func (q *Quote) RecalculateTotals() error {
	var subtotal, estimated big.Int
	for _, item := range q.Items {
		subtotal.Add(&subtotal, big.NewInt(item.SellingPriceRial))
		estimated.Add(&estimated, big.NewInt(item.EstimatedCostRial))
	}
	if !subtotal.IsInt64() || !estimated.IsInt64() {
		return fmt.Errorf("quote totals exceed Rial range")
	}
	q.SubtotalRial, q.EstimatedCostRial = subtotal.Int64(), estimated.Int64()
	if q.DiscountRial < 0 || q.DiscountRial > q.SubtotalRial {
		return fmt.Errorf("discount cannot exceed subtotal")
	}
	q.TotalRial = q.SubtotalRial - q.DiscountRial
	return nil
}

func ValidQuoteStatus(v QuoteStatus) bool {
	return v == QuoteDraft || v == QuoteSent || v == QuoteAccepted || v == QuoteRejected || v == QuoteExpired || v == QuoteConverted
}
func ValidQuoteTransition(from, to QuoteStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case QuoteDraft:
		return to == QuoteSent || to == QuoteRejected || to == QuoteExpired
	case QuoteSent:
		return to == QuoteAccepted || to == QuoteRejected || to == QuoteExpired
	case QuoteAccepted:
		return to == QuoteRejected
	}
	return false
}
