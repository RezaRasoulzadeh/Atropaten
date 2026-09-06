package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type SupplierRepository interface {
	ListSuppliers(context.Context, bool) ([]domain.Supplier, error)
	GetSupplier(context.Context, string) (domain.Supplier, error)
	SaveSupplier(context.Context, domain.Supplier) error
	DeleteSupplier(context.Context, string) error
}

type SupplierView struct {
	ID, Name, Code, Phone, Email, Address, Notes, CreatedAt, UpdatedAt string
	Active                                                             bool
}
type SupplierInput struct{ Name, Code, Phone, Email, Address, Notes string }

type SuppliersService struct {
	repository SupplierRepository
	now        func() time.Time
}

func NewSuppliersService(r SupplierRepository) *SuppliersService {
	return &SuppliersService{repository: r, now: time.Now}
}
func (s *SuppliersService) List(ctx context.Context, archived bool) ([]SupplierView, error) {
	rows, e := s.repository.ListSuppliers(ctx, archived)
	if e != nil {
		return nil, e
	}
	out := make([]SupplierView, 0, len(rows))
	for _, v := range rows {
		out = append(out, supplierView(v))
	}
	return out, nil
}
func (s *SuppliersService) Get(ctx context.Context, id string) (SupplierView, error) {
	v, e := s.repository.GetSupplier(ctx, strings.TrimSpace(id))
	if e != nil {
		return SupplierView{}, e
	}
	return supplierView(v), nil
}
func (s *SuppliersService) Create(ctx context.Context, in SupplierInput) (SupplierView, error) {
	id, e := randomID("SUP-")
	if e != nil {
		return SupplierView{}, e
	}
	now := s.now().UTC()
	v := domain.Supplier{ID: id, Name: strings.TrimSpace(in.Name), Code: strings.TrimSpace(in.Code), Phone: strings.TrimSpace(in.Phone), Email: strings.TrimSpace(in.Email), Address: strings.TrimSpace(in.Address), Notes: strings.TrimSpace(in.Notes), Active: true, CreatedAt: now, UpdatedAt: now}
	if e = v.Validate(); e != nil {
		return SupplierView{}, e
	}
	if e = s.repository.SaveSupplier(ctx, v); e != nil {
		return SupplierView{}, e
	}
	return supplierView(v), nil
}
func (s *SuppliersService) Update(ctx context.Context, id string, in SupplierInput) (SupplierView, error) {
	v, e := s.repository.GetSupplier(ctx, strings.TrimSpace(id))
	if e != nil {
		return SupplierView{}, e
	}
	v.Name = strings.TrimSpace(in.Name)
	v.Code = strings.TrimSpace(in.Code)
	v.Phone = strings.TrimSpace(in.Phone)
	v.Email = strings.TrimSpace(in.Email)
	v.Address = strings.TrimSpace(in.Address)
	v.Notes = strings.TrimSpace(in.Notes)
	v.UpdatedAt = s.now().UTC()
	if e = v.Validate(); e != nil {
		return SupplierView{}, e
	}
	if e = s.repository.SaveSupplier(ctx, v); e != nil {
		return SupplierView{}, e
	}
	return supplierView(v), nil
}
func (s *SuppliersService) SetActive(ctx context.Context, id string, active bool) (SupplierView, error) {
	v, e := s.repository.GetSupplier(ctx, id)
	if e != nil {
		return SupplierView{}, e
	}
	v.Active = active
	v.UpdatedAt = s.now().UTC()
	if e = s.repository.SaveSupplier(ctx, v); e != nil {
		return SupplierView{}, e
	}
	return supplierView(v), nil
}
func (s *SuppliersService) Delete(ctx context.Context, id string) error {
	return s.repository.DeleteSupplier(ctx, strings.TrimSpace(id))
}
func supplierView(v domain.Supplier) SupplierView {
	return SupplierView{ID: v.ID, Name: v.Name, Code: v.Code, Phone: v.Phone, Email: v.Email, Address: v.Address, Notes: v.Notes, Active: v.Active, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}

type PurchaseRepository interface {
	ListPurchases(context.Context) ([]domain.Purchase, error)
	GetPurchase(context.Context, string) (domain.Purchase, error)
	SavePurchase(context.Context, domain.Purchase) error
	DeleteDraftPurchase(context.Context, string) error
	PostPurchase(context.Context, string) error
	CancelPurchase(context.Context, string) error
	ListInventoryMovements(context.Context, string) ([]domain.InventoryMovement, error)
	AdjustInventory(context.Context, string, domain.Quantity, int64, string) error
}
type PurchasePaymentLookup interface {
	PurchasePaymentSummary(context.Context, string) (int64, int64, error)
}
type PurchaseView struct {
	ID, PurchaseNumber, SupplierID, SupplierName, SupplierInvoiceNumber, PurchaseDate, Status, Notes string
	SubtotalRial, DiscountRial, ShippingRial, TaxRial, AdditionalCostsRial, TotalRial                int64
	PaidRial, RemainingRial                                                                          int64
	CreatedAt, UpdatedAt                                                                             string
	Items                                                                                            []PurchaseItemView
}
type PurchaseItemView struct {
	ID                                                                                                               string
	Position                                                                                                         int
	MaterialID, MaterialName, PurchaseUnit, ConsumptionUnit, PurchaseQuantity, ConversionFactor, ConsumptionQuantity string
	UnitAcquisitionCostRial, AllocatedAdditionalCostRial, LandedUnitCostRial, LineTotalRial                          int64
	Notes                                                                                                            string
}
type PurchaseInput struct {
	SupplierID, PurchaseDate, SupplierInvoiceNumber, Notes   string
	DiscountRial, ShippingRial, TaxRial, AdditionalCostsRial int64
}
type PurchaseItemInput struct{ MaterialID, PurchaseQuantity, UnitAcquisitionCostRial, Notes string }
type InventoryMovementView struct {
	ID, MaterialID, OccurredAt, MovementType, QuantityDelta, ReferenceType, ReferenceID, Note, CreatedAt string
	UnitCostRial, TotalCostRial                                                                          int64
}

type PurchasesService struct {
	repository PurchaseRepository
	suppliers  interface {
		GetSupplier(context.Context, string) (domain.Supplier, error)
	}
	materials MaterialRepository
	now       func() time.Time
}

func NewPurchasesService(r PurchaseRepository, s interface {
	GetSupplier(context.Context, string) (domain.Supplier, error)
}, m MaterialRepository) *PurchasesService {
	return &PurchasesService{repository: r, suppliers: s, materials: m, now: time.Now}
}
func (s *PurchasesService) List(ctx context.Context) ([]PurchaseView, error) {
	rows, e := s.repository.ListPurchases(ctx)
	if e != nil {
		return nil, e
	}
	out := make([]PurchaseView, 0, len(rows))
	for _, v := range rows {
		view := purchaseView(v)
		if lookup, ok := s.repository.(PurchasePaymentLookup); ok {
			view.PaidRial, view.RemainingRial, e = lookup.PurchasePaymentSummary(ctx, v.ID)
			if e != nil {
				return nil, e
			}
		}
		out = append(out, view)
	}
	return out, nil
}
func (s *PurchasesService) Get(ctx context.Context, id string) (PurchaseView, error) {
	v, e := s.repository.GetPurchase(ctx, id)
	if e != nil {
		return PurchaseView{}, e
	}
	view := purchaseView(v)
	if lookup, ok := s.repository.(PurchasePaymentLookup); ok {
		view.PaidRial, view.RemainingRial, e = lookup.PurchasePaymentSummary(ctx, v.ID)
		if e != nil {
			return PurchaseView{}, e
		}
	}
	return view, nil
}
func (s *PurchasesService) Create(ctx context.Context, in PurchaseInput) (PurchaseView, error) {
	p, e := s.newPurchase(ctx, in)
	if e != nil {
		return PurchaseView{}, e
	}
	if e = s.repository.SavePurchase(ctx, p); e != nil {
		return PurchaseView{}, e
	}
	return s.Get(ctx, p.ID)
}
func (s *PurchasesService) Update(ctx context.Context, id string, in PurchaseInput) (PurchaseView, error) {
	p, e := s.repository.GetPurchase(ctx, id)
	if e != nil {
		return PurchaseView{}, e
	}
	if p.Status != domain.PurchaseDraft {
		return PurchaseView{}, domain.ErrPurchaseNotDraft
	}
	if e = s.applyPurchaseInput(ctx, &p, in); e != nil {
		return PurchaseView{}, e
	}
	p.UpdatedAt = s.now().UTC()
	if e = s.repository.SavePurchase(ctx, p); e != nil {
		return PurchaseView{}, e
	}
	return purchaseView(p), nil
}
func (s *PurchasesService) newPurchase(ctx context.Context, in PurchaseInput) (domain.Purchase, error) {
	id, e := randomID("PUR-")
	if e != nil {
		return domain.Purchase{}, e
	}
	now := s.now().UTC()
	p := domain.Purchase{ID: id, PurchaseDate: now, Status: domain.PurchaseDraft, CreatedAt: now, UpdatedAt: now}
	if e = s.applyPurchaseInput(ctx, &p, in); e != nil {
		return domain.Purchase{}, e
	}
	return p, nil
}
func (s *PurchasesService) applyPurchaseInput(ctx context.Context, p *domain.Purchase, in PurchaseInput) error {
	if strings.TrimSpace(in.SupplierID) == "" {
		return fmt.Errorf("supplier is required")
	}
	sup, e := s.suppliers.GetSupplier(ctx, in.SupplierID)
	if e != nil {
		return e
	}
	p.SupplierID = sup.ID
	p.SupplierNameSnapshot = sup.Name
	p.SupplierCodeSnapshot = sup.Code
	p.SupplierInvoiceNumber = strings.TrimSpace(in.SupplierInvoiceNumber)
	p.Notes = strings.TrimSpace(in.Notes)
	p.DiscountRial = in.DiscountRial
	p.ShippingRial = in.ShippingRial
	p.TaxRial = in.TaxRial
	p.AdditionalCostsRial = in.AdditionalCostsRial
	if in.PurchaseDate != "" {
		t, pe := time.Parse(time.RFC3339, in.PurchaseDate)
		if pe != nil {
			t, pe = time.Parse(time.RFC3339Nano, in.PurchaseDate)
		}
		if pe != nil {
			t, pe = time.Parse("2006-01-02T15:04", in.PurchaseDate)
		}
		if pe != nil {
			t, pe = time.Parse("2006-01-02", in.PurchaseDate)
		}
		if pe != nil {
			return fmt.Errorf("purchase date: %w", pe)
		}
		p.PurchaseDate = t.UTC()
	}
	return recalculatePurchase(p)
}
func recalculatePurchase(p *domain.Purchase) error {
	var sub int64
	for i := range p.Items {
		p.Items[i].Position = i
		if p.Items[i].PurchaseQuantity < 0 || p.Items[i].UnitAcquisitionCostRial < 0 {
			return fmt.Errorf("purchase item values cannot be negative")
		}
		v, e := domain.MulQuantityRial(p.Items[i].PurchaseQuantity, p.Items[i].UnitAcquisitionCostRial)
		if e != nil {
			return e
		}
		p.Items[i].LineTotalRial = v
		if sub > int64(^uint64(0)>>1)-v {
			return fmt.Errorf("purchase subtotal is too large")
		}
		sub += v
	}
	if p.DiscountRial < 0 || p.ShippingRial < 0 || p.TaxRial < 0 || p.AdditionalCostsRial < 0 {
		return fmt.Errorf("purchase costs cannot be negative")
	}
	if p.DiscountRial > sub {
		return fmt.Errorf("discount cannot exceed subtotal")
	}
	total := sub - p.DiscountRial + p.ShippingRial + p.TaxRial + p.AdditionalCostsRial
	if total < 0 {
		return fmt.Errorf("purchase total is invalid")
	}
	p.SubtotalRial = sub
	p.TotalRial = total
	return nil
}
func (s *PurchasesService) AddItem(ctx context.Context, id string, in PurchaseItemInput) (PurchaseView, error) {
	p, e := s.repository.GetPurchase(ctx, id)
	if e != nil {
		return PurchaseView{}, e
	}
	if p.Status != domain.PurchaseDraft {
		return PurchaseView{}, domain.ErrPurchaseNotDraft
	}
	item, e := s.item(ctx, p.ID, in)
	if e != nil {
		return PurchaseView{}, e
	}
	p.Items = append(p.Items, item)
	p.UpdatedAt = s.now().UTC()
	if e = recalculatePurchase(&p); e != nil {
		return PurchaseView{}, e
	}
	if e = s.repository.SavePurchase(ctx, p); e != nil {
		return PurchaseView{}, e
	}
	return purchaseView(p), nil
}
func (s *PurchasesService) item(ctx context.Context, pid string, in PurchaseItemInput) (domain.PurchaseItem, error) {
	m, e := s.materials.Get(ctx, in.MaterialID)
	if e != nil {
		return domain.PurchaseItem{}, e
	}
	q, e := domain.ParseQuantity(in.PurchaseQuantity)
	if e != nil {
		return domain.PurchaseItem{}, e
	}
	cost, err := parseRialText(in.UnitAcquisitionCostRial)
	if err != nil {
		return domain.PurchaseItem{}, err
	}
	cons, e := domain.ConvertPurchaseQuantity(q, m.ConversionFactor)
	if e != nil {
		return domain.PurchaseItem{}, e
	}
	return domain.PurchaseItem{ID: mustID("PITM-"), PurchaseID: pid, MaterialID: m.ID, MaterialNameSnapshot: m.Name, PurchaseUnitSnapshot: m.PurchaseUnit, ConsumptionUnitSnapshot: m.ConsumptionUnit, PurchaseQuantity: q, ConversionFactorSnapshot: m.ConversionFactor, ConsumptionQuantity: cons, UnitAcquisitionCostRial: cost, Notes: strings.TrimSpace(in.Notes)}, nil
}
func (s *PurchasesService) UpdateItem(ctx context.Context, id, itemID string, in PurchaseItemInput) (PurchaseView, error) {
	p, e := s.repository.GetPurchase(ctx, id)
	if e != nil {
		return PurchaseView{}, e
	}
	if p.Status != domain.PurchaseDraft {
		return PurchaseView{}, domain.ErrPurchaseNotDraft
	}
	item, e := s.item(ctx, p.ID, in)
	if e != nil {
		return PurchaseView{}, e
	}
	item.ID = itemID
	found := false
	for i := range p.Items {
		if p.Items[i].ID == itemID {
			item.Position = i
			p.Items[i] = item
			found = true
		}
	}
	if !found {
		return PurchaseView{}, fmt.Errorf("purchase item not found")
	}
	p.UpdatedAt = s.now().UTC()
	if e = recalculatePurchase(&p); e != nil {
		return PurchaseView{}, e
	}
	if e = s.repository.SavePurchase(ctx, p); e != nil {
		return PurchaseView{}, e
	}
	return purchaseView(p), nil
}
func (s *PurchasesService) RemoveItem(ctx context.Context, id, itemID string) (PurchaseView, error) {
	p, e := s.repository.GetPurchase(ctx, id)
	if e != nil {
		return PurchaseView{}, e
	}
	if p.Status != domain.PurchaseDraft {
		return PurchaseView{}, domain.ErrPurchaseNotDraft
	}
	out := p.Items[:0]
	found := false
	for _, i := range p.Items {
		if i.ID == itemID {
			found = true
		} else {
			out = append(out, i)
		}
	}
	if !found {
		return PurchaseView{}, fmt.Errorf("purchase item not found")
	}
	p.Items = out
	p.UpdatedAt = s.now().UTC()
	if e = recalculatePurchase(&p); e != nil {
		return PurchaseView{}, e
	}
	if e = s.repository.SavePurchase(ctx, p); e != nil {
		return PurchaseView{}, e
	}
	return purchaseView(p), nil
}
func (s *PurchasesService) ReorderItems(ctx context.Context, id string, ids []string) (PurchaseView, error) {
	p, e := s.repository.GetPurchase(ctx, id)
	if e != nil {
		return PurchaseView{}, e
	}
	if p.Status != domain.PurchaseDraft {
		return PurchaseView{}, domain.ErrPurchaseNotDraft
	}
	by := map[string]domain.PurchaseItem{}
	for _, i := range p.Items {
		by[i.ID] = i
	}
	items := make([]domain.PurchaseItem, 0, len(ids))
	for _, id := range ids {
		i, ok := by[id]
		if !ok {
			return PurchaseView{}, fmt.Errorf("unknown purchase item %q", id)
		}
		items = append(items, i)
	}
	if len(items) != len(p.Items) {
		return PurchaseView{}, fmt.Errorf("all purchase items must be included")
	}
	p.Items = items
	p.UpdatedAt = s.now().UTC()
	if e = recalculatePurchase(&p); e != nil {
		return PurchaseView{}, e
	}
	if e = s.repository.SavePurchase(ctx, p); e != nil {
		return PurchaseView{}, e
	}
	return purchaseView(p), nil
}
func (s *PurchasesService) Post(ctx context.Context, id string) (PurchaseView, error) {
	if e := s.repository.PostPurchase(ctx, id); e != nil {
		return PurchaseView{}, e
	}
	return s.Get(ctx, id)
}
func (s *PurchasesService) Cancel(ctx context.Context, id string) (PurchaseView, error) {
	if e := s.repository.CancelPurchase(ctx, id); e != nil {
		return PurchaseView{}, e
	}
	return s.Get(ctx, id)
}
func (s *PurchasesService) DeleteDraft(ctx context.Context, id string) error {
	return s.repository.DeleteDraftPurchase(ctx, id)
}
func (s *PurchasesService) Movements(ctx context.Context, id string) ([]InventoryMovementView, error) {
	rows, e := s.repository.ListInventoryMovements(ctx, id)
	if e != nil {
		return nil, e
	}
	out := make([]InventoryMovementView, 0, len(rows))
	for _, v := range rows {
		out = append(out, InventoryMovementView{ID: v.ID, MaterialID: v.MaterialID, OccurredAt: v.OccurredAt.UTC().Format(time.RFC3339Nano), MovementType: v.MovementType, QuantityDelta: v.QuantityDelta.String(), UnitCostRial: v.UnitCostRial, TotalCostRial: v.TotalCostRial, ReferenceType: v.ReferenceType, ReferenceID: v.ReferenceID, Note: v.Note, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano)})
	}
	return out, nil
}
func (s *PurchasesService) Adjust(ctx context.Context, id, qty string, cost int64, note string) error {
	q, e := parseSignedQuantity(qty)
	if e != nil {
		return e
	}
	return s.repository.AdjustInventory(ctx, id, q, cost, note)
}
func purchaseView(p domain.Purchase) PurchaseView {
	out := PurchaseView{ID: p.ID, PurchaseNumber: p.PurchaseNumber, SupplierID: p.SupplierID, SupplierName: p.SupplierNameSnapshot, SupplierInvoiceNumber: p.SupplierInvoiceNumber, PurchaseDate: p.PurchaseDate.UTC().Format(time.RFC3339Nano), Status: p.Status, Notes: p.Notes, SubtotalRial: p.SubtotalRial, DiscountRial: p.DiscountRial, ShippingRial: p.ShippingRial, TaxRial: p.TaxRial, AdditionalCostsRial: p.AdditionalCostsRial, TotalRial: p.TotalRial, CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: p.UpdatedAt.UTC().Format(time.RFC3339Nano), Items: make([]PurchaseItemView, 0, len(p.Items))}
	for _, i := range p.Items {
		out.Items = append(out.Items, PurchaseItemView{ID: i.ID, Position: i.Position, MaterialID: i.MaterialID, MaterialName: i.MaterialNameSnapshot, PurchaseUnit: i.PurchaseUnitSnapshot, ConsumptionUnit: i.ConsumptionUnitSnapshot, PurchaseQuantity: i.PurchaseQuantity.String(), ConversionFactor: i.ConversionFactorSnapshot.String(), ConsumptionQuantity: i.ConsumptionQuantity.String(), UnitAcquisitionCostRial: i.UnitAcquisitionCostRial, AllocatedAdditionalCostRial: i.AllocatedAdditionalCostRial, LandedUnitCostRial: i.LandedUnitCostRial, LineTotalRial: i.LineTotalRial, Notes: i.Notes})
	}
	return out
}
func parseRial(v int64) (int64, error) {
	if v < 0 {
		return 0, fmt.Errorf("Rial amount cannot be negative")
	}
	return v, nil
}
func parseRialText(v string) (int64, error) {
	v = strings.ReplaceAll(strings.TrimSpace(v), ",", "")
	if v == "" {
		return 0, fmt.Errorf("Rial amount is required")
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Rial amount must be a whole number")
	}
	return parseRial(n)
}
func parseSignedQuantity(v string) (domain.Quantity, error) {
	v = strings.TrimSpace(v)
	negative := strings.HasPrefix(v, "-")
	if negative {
		v = strings.TrimPrefix(v, "-")
	}
	q, e := domain.ParseQuantity(v)
	if e != nil {
		return 0, e
	}
	if negative {
		return -q, nil
	}
	return q, nil
}
func mustID(prefix string) string {
	id, e := randomID(prefix)
	if e != nil {
		return prefix + fmt.Sprint(time.Now().UnixNano())
	}
	return id
}
