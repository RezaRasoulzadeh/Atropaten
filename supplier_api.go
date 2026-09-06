package main

import (
	"Atropaten/internal/application"
	"fmt"
)

type SupplierInput struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Notes   string `json:"notes"`
}
type SupplierDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Address   string `json:"address"`
	Notes     string `json:"notes"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func (a *App) supplierService() (*application.SuppliersService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.suppliers == nil {
		return nil, fmt.Errorf("suppliers service is not initialized")
	}
	return a.suppliers, nil
}
func (a *App) ListSuppliers(archived bool) ([]SupplierDTO, error) {
	s, e := a.supplierService()
	if e != nil {
		return nil, e
	}
	rows, e := s.List(a.materialContext(), archived)
	if e != nil {
		return nil, e
	}
	out := make([]SupplierDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, supplierDTO(v))
	}
	return out, nil
}
func (a *App) GetSupplier(id string) (SupplierDTO, error) {
	s, e := a.supplierService()
	if e != nil {
		return SupplierDTO{}, e
	}
	v, e := s.Get(a.materialContext(), id)
	return supplierDTO(v), e
}
func (a *App) CreateSupplier(i SupplierInput) (SupplierDTO, error) {
	s, e := a.supplierService()
	if e != nil {
		return SupplierDTO{}, e
	}
	v, e := s.Create(a.materialContext(), application.SupplierInput(i))
	return supplierDTO(v), e
}
func (a *App) UpdateSupplier(id string, i SupplierInput) (SupplierDTO, error) {
	s, e := a.supplierService()
	if e != nil {
		return SupplierDTO{}, e
	}
	v, e := s.Update(a.materialContext(), id, application.SupplierInput(i))
	return supplierDTO(v), e
}
func (a *App) ArchiveSupplier(id string) (SupplierDTO, error) { return a.setSupplierActive(id, false) }
func (a *App) ReactivateSupplier(id string) (SupplierDTO, error) {
	return a.setSupplierActive(id, true)
}
func (a *App) setSupplierActive(id string, active bool) (SupplierDTO, error) {
	s, e := a.supplierService()
	if e != nil {
		return SupplierDTO{}, e
	}
	v, e := s.SetActive(a.materialContext(), id, active)
	return supplierDTO(v), e
}
func (a *App) DeleteSupplier(id string) error {
	s, e := a.supplierService()
	if e != nil {
		return e
	}
	return s.Delete(a.materialContext(), id)
}
func supplierDTO(v application.SupplierView) SupplierDTO {
	return SupplierDTO{ID: v.ID, Name: v.Name, Code: v.Code, Phone: v.Phone, Email: v.Email, Address: v.Address, Notes: v.Notes, Active: v.Active, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
