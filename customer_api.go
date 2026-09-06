package main

import (
	"Atropaten/internal/application"
	"fmt"
)

type CustomerInput struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Notes   string `json:"notes"`
}
type CustomerDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Address   string `json:"address"`
	Notes     string `json:"notes"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func (a *App) customerService() (*application.CustomersService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.customers == nil {
		return nil, fmt.Errorf("customers service is not initialized")
	}
	return a.customers, nil
}
func (a *App) ListCustomers(includeArchived bool) ([]CustomerDTO, error) {
	s, e := a.customerService()
	if e != nil {
		return nil, e
	}
	rows, e := s.List(a.materialContext(), includeArchived)
	if e != nil {
		return nil, e
	}
	out := make([]CustomerDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, customerDTO(r))
	}
	return out, nil
}
func (a *App) GetCustomer(id string) (CustomerDTO, error) {
	s, e := a.customerService()
	if e != nil {
		return CustomerDTO{}, e
	}
	r, e := s.Get(a.materialContext(), id)
	if e != nil {
		return CustomerDTO{}, e
	}
	return customerDTO(r), nil
}
func (a *App) CreateCustomer(i CustomerInput) (CustomerDTO, error) {
	s, e := a.customerService()
	if e != nil {
		return CustomerDTO{}, e
	}
	r, e := s.Create(a.materialContext(), application.CustomerInput{Name: i.Name, Phone: i.Phone, Email: i.Email, Address: i.Address, Notes: i.Notes})
	if e != nil {
		return CustomerDTO{}, e
	}
	return customerDTO(r), nil
}
func (a *App) UpdateCustomer(id string, i CustomerInput) (CustomerDTO, error) {
	s, e := a.customerService()
	if e != nil {
		return CustomerDTO{}, e
	}
	r, e := s.Update(a.materialContext(), id, application.CustomerInput{Name: i.Name, Phone: i.Phone, Email: i.Email, Address: i.Address, Notes: i.Notes})
	if e != nil {
		return CustomerDTO{}, e
	}
	return customerDTO(r), nil
}
func (a *App) ArchiveCustomer(id string) (CustomerDTO, error) { return a.setCustomerActive(id, false) }
func (a *App) ReactivateCustomer(id string) (CustomerDTO, error) {
	return a.setCustomerActive(id, true)
}
func (a *App) setCustomerActive(id string, active bool) (CustomerDTO, error) {
	s, e := a.customerService()
	if e != nil {
		return CustomerDTO{}, e
	}
	var r application.CustomerView
	if active {
		r, e = s.Reactivate(a.materialContext(), id)
	} else {
		r, e = s.Archive(a.materialContext(), id)
	}
	if e != nil {
		return CustomerDTO{}, e
	}
	return customerDTO(r), nil
}
func customerDTO(v application.CustomerView) CustomerDTO {
	return CustomerDTO{ID: v.ID, Name: v.Name, Phone: v.Phone, Email: v.Email, Address: v.Address, Notes: v.Notes, Active: v.Active, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
