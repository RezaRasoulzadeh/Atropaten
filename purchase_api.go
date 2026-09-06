package main

import (
	"Atropaten/internal/application"
	"fmt"
)

type PurchaseInput struct {
	SupplierID            string `json:"supplierId"`
	PurchaseDate          string `json:"purchaseDate"`
	SupplierInvoiceNumber string `json:"supplierInvoiceNumber"`
	Notes                 string `json:"notes"`
	DiscountRial          int64  `json:"discountRial"`
	ShippingRial          int64  `json:"shippingRial"`
	TaxRial               int64  `json:"taxRial"`
	AdditionalCostsRial   int64  `json:"additionalCostsRial"`
}
type PurchaseItemInput struct {
	MaterialID              string `json:"materialId"`
	PurchaseQuantity        string `json:"purchaseQuantity"`
	UnitAcquisitionCostRial string `json:"unitAcquisitionCostRial"`
	Notes                   string `json:"notes"`
}
type PurchaseDTO struct {
	ID                    string            `json:"id"`
	PurchaseNumber        string            `json:"purchaseNumber"`
	SupplierID            string            `json:"supplierId"`
	SupplierName          string            `json:"supplierName"`
	SupplierInvoiceNumber string            `json:"supplierInvoiceNumber"`
	PurchaseDate          string            `json:"purchaseDate"`
	Status                string            `json:"status"`
	Notes                 string            `json:"notes"`
	SubtotalRial          int64             `json:"subtotalRial"`
	DiscountRial          int64             `json:"discountRial"`
	ShippingRial          int64             `json:"shippingRial"`
	TaxRial               int64             `json:"taxRial"`
	AdditionalCostsRial   int64             `json:"additionalCostsRial"`
	TotalRial             int64             `json:"totalRial"`
	PaidRial              int64             `json:"paidRial"`
	RemainingRial         int64             `json:"remainingRial"`
	CreatedAt             string            `json:"createdAt"`
	UpdatedAt             string            `json:"updatedAt"`
	Items                 []PurchaseItemDTO `json:"items"`
}
type PurchaseItemDTO struct {
	ID                          string `json:"id"`
	Position                    int    `json:"position"`
	MaterialID                  string `json:"materialId"`
	MaterialName                string `json:"materialName"`
	PurchaseUnit                string `json:"purchaseUnit"`
	ConsumptionUnit             string `json:"consumptionUnit"`
	PurchaseQuantity            string `json:"purchaseQuantity"`
	ConversionFactor            string `json:"conversionFactor"`
	ConsumptionQuantity         string `json:"consumptionQuantity"`
	UnitAcquisitionCostRial     int64  `json:"unitAcquisitionCostRial"`
	AllocatedAdditionalCostRial int64  `json:"allocatedAdditionalCostRial"`
	LandedUnitCostRial          int64  `json:"landedUnitCostRial"`
	LineTotalRial               int64  `json:"lineTotalRial"`
	Notes                       string `json:"notes"`
}
type InventoryMovementDTO struct {
	ID            string `json:"id"`
	MaterialID    string `json:"materialId"`
	OccurredAt    string `json:"occurredAt"`
	MovementType  string `json:"movementType"`
	QuantityDelta string `json:"quantityDelta"`
	UnitCostRial  int64  `json:"unitCostRial"`
	TotalCostRial int64  `json:"totalCostRial"`
	ReferenceType string `json:"referenceType"`
	ReferenceID   string `json:"referenceId"`
	Note          string `json:"note"`
	CreatedAt     string `json:"createdAt"`
}

func (a *App) purchaseService() (*application.PurchasesService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.purchases == nil {
		return nil, fmt.Errorf("purchases service is not initialized")
	}
	return a.purchases, nil
}
func (a *App) ListPurchases() ([]PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return nil, e
	}
	rows, e := s.List(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]PurchaseDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, purchaseDTO(v))
	}
	return out, nil
}
func (a *App) GetPurchase(id string) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.Get(a.materialContext(), id)
	return purchaseDTO(v), e
}
func (a *App) CreatePurchase(i PurchaseInput) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.Create(a.materialContext(), application.PurchaseInput(i))
	return purchaseDTO(v), e
}
func (a *App) UpdatePurchase(id string, i PurchaseInput) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.Update(a.materialContext(), id, application.PurchaseInput(i))
	return purchaseDTO(v), e
}
func (a *App) AddPurchaseItem(id string, i PurchaseItemInput) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.AddItem(a.materialContext(), id, application.PurchaseItemInput(i))
	return purchaseDTO(v), e
}
func (a *App) UpdatePurchaseItem(id, itemID string, i PurchaseItemInput) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.UpdateItem(a.materialContext(), id, itemID, application.PurchaseItemInput(i))
	return purchaseDTO(v), e
}
func (a *App) RemovePurchaseItem(id, itemID string) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.RemoveItem(a.materialContext(), id, itemID)
	return purchaseDTO(v), e
}
func (a *App) ReorderPurchaseItems(id string, ids []string) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.ReorderItems(a.materialContext(), id, ids)
	return purchaseDTO(v), e
}
func (a *App) PostPurchase(id string) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.Post(a.materialContext(), id)
	return purchaseDTO(v), e
}
func (a *App) CancelPurchase(id string) (PurchaseDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return PurchaseDTO{}, e
	}
	v, e := s.Cancel(a.materialContext(), id)
	return purchaseDTO(v), e
}
func (a *App) DeleteDraftPurchase(id string) error {
	s, e := a.purchaseService()
	if e != nil {
		return e
	}
	return s.DeleteDraft(a.materialContext(), id)
}
func (a *App) ListMaterialMovements(id string) ([]InventoryMovementDTO, error) {
	s, e := a.purchaseService()
	if e != nil {
		return nil, e
	}
	rows, e := s.Movements(a.materialContext(), id)
	if e != nil {
		return nil, e
	}
	out := make([]InventoryMovementDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, InventoryMovementDTO{ID: v.ID, MaterialID: v.MaterialID, OccurredAt: v.OccurredAt, MovementType: v.MovementType, QuantityDelta: v.QuantityDelta, UnitCostRial: v.UnitCostRial, TotalCostRial: v.TotalCostRial, ReferenceType: v.ReferenceType, ReferenceID: v.ReferenceID, Note: v.Note, CreatedAt: v.CreatedAt})
	}
	return out, nil
}
func (a *App) AdjustMaterialStock(id, qty string, cost int64, note string) error {
	s, e := a.purchaseService()
	if e != nil {
		return e
	}
	return s.Adjust(a.materialContext(), id, qty, cost, note)
}
func purchaseDTO(v application.PurchaseView) PurchaseDTO {
	out := PurchaseDTO{ID: v.ID, PurchaseNumber: v.PurchaseNumber, SupplierID: v.SupplierID, SupplierName: v.SupplierName, SupplierInvoiceNumber: v.SupplierInvoiceNumber, PurchaseDate: v.PurchaseDate, Status: v.Status, Notes: v.Notes, SubtotalRial: v.SubtotalRial, DiscountRial: v.DiscountRial, ShippingRial: v.ShippingRial, TaxRial: v.TaxRial, AdditionalCostsRial: v.AdditionalCostsRial, TotalRial: v.TotalRial, PaidRial: v.PaidRial, RemainingRial: v.RemainingRial, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Items: make([]PurchaseItemDTO, 0, len(v.Items))}
	for _, i := range v.Items {
		out.Items = append(out.Items, PurchaseItemDTO{ID: i.ID, Position: i.Position, MaterialID: i.MaterialID, MaterialName: i.MaterialName, PurchaseUnit: i.PurchaseUnit, ConsumptionUnit: i.ConsumptionUnit, PurchaseQuantity: i.PurchaseQuantity, ConversionFactor: i.ConversionFactor, ConsumptionQuantity: i.ConsumptionQuantity, UnitAcquisitionCostRial: i.UnitAcquisitionCostRial, AllocatedAdditionalCostRial: i.AllocatedAdditionalCostRial, LandedUnitCostRial: i.LandedUnitCostRial, LineTotalRial: i.LineTotalRial, Notes: i.Notes})
	}
	return out
}
