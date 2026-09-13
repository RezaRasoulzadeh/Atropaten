package domain

import (
	"fmt"
	"math/big"
)

// SnapshotOrder replaces only order-owned commercial fields. Invoice identity,
// dates and journal references belong to the document's lifecycle.
func (invoice Invoice) SnapshotOrder(order Order) (Invoice, error) {
	invoice.OrderID = order.ID
	invoice.CustomerID = order.CustomerID
	invoice.CustomerNameSnapshot = order.CustomerNameSnapshot
	if invoice.CustomerNameSnapshot == "" {
		invoice.CustomerNameSnapshot = "Walk-in customer"
	}
	invoice.CustomerPhoneSnapshot = order.CustomerPhoneSnapshot
	invoice.Notes = order.Notes
	invoice.SubtotalRial, invoice.DiscountRial, invoice.TotalRial = order.SubtotalRial, order.DiscountRial, order.TotalRial
	ids := map[string]string{}
	for _, item := range invoice.Items {
		ids[item.OrderItemID] = item.ID
	}
	invoice.Items = nil
	for position, item := range order.Items {
		id := ids[item.ID]
		if id == "" {
			id = invoice.ID + "-LINE-" + item.ID
		}
		// SellingPriceRial is the whole order line. Retain that exact total;
		// the displayed per-unit rate is rounded to the nearest Rial.
		price := new(big.Int).Mul(big.NewInt(item.SellingPriceRial), big.NewInt(QuantityScale))
		if item.Quantity > 0 {
			price.Add(price, big.NewInt(int64(item.Quantity)/2))
			price.Quo(price, big.NewInt(int64(item.Quantity)))
		}
		if !price.IsInt64() {
			return Invoice{}, fmt.Errorf("invoice unit price is too large")
		}
		unitPrice := price.Int64()
		invoice.Items = append(invoice.Items, InvoiceItem{ID: id, InvoiceID: invoice.ID, OrderItemID: item.ID, Position: position, DescriptionSnapshot: item.ServiceNameSnapshot, ServiceID: item.ServiceID, QuantityUnits: int64(item.Quantity), QuantityUnit: item.QuantityUnit, UnitPriceRial: unitPrice, LineTotalRial: item.SellingPriceRial, Notes: item.Notes})
	}
	return invoice, nil
}
