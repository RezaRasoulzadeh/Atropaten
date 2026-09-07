package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type CustomerRepository interface {
	ListCustomers(context.Context, bool) ([]domain.Customer, error)
	GetCustomer(context.Context, string) (domain.Customer, error)
	SaveCustomer(context.Context, domain.Customer) error
}
type CustomerInput struct{ Name, Phone, Email, Address, Notes string }
type CustomerView struct {
	ID, Name, Phone, Email, Address, Notes string
	Active                                 bool
	CreatedAt, UpdatedAt                   string
}
type CustomersService struct {
	repository CustomerRepository
	now        func() time.Time
}

func NewCustomersService(repository CustomerRepository) *CustomersService {
	return &CustomersService{repository: repository, now: time.Now}
}
func (s *CustomersService) List(ctx context.Context, includeArchived bool) ([]CustomerView, error) {
	rows, err := s.repository.ListCustomers(ctx, includeArchived)
	if err != nil {
		return nil, err
	}
	out := make([]CustomerView, 0, len(rows))
	for _, row := range rows {
		out = append(out, customerView(row))
	}
	return out, nil
}
func (s *CustomersService) Get(ctx context.Context, id string) (CustomerView, error) {
	row, err := s.repository.GetCustomer(ctx, strings.TrimSpace(id))
	if err != nil {
		return CustomerView{}, err
	}
	return customerView(row), nil
}
func (s *CustomersService) Create(ctx context.Context, input CustomerInput) (CustomerView, error) {
	id, err := randomID("CUS-")
	if err != nil {
		return CustomerView{}, err
	}
	row, err := domain.NewCustomer(id, domain.CustomerDraft{Name: input.Name, Phone: input.Phone, Email: input.Email, Address: input.Address, Notes: input.Notes}, s.now())
	if err != nil {
		return CustomerView{}, err
	}
	if err := s.repository.SaveCustomer(ctx, row); err != nil {
		return CustomerView{}, err
	}
	return customerView(row), nil
}
func (s *CustomersService) Update(ctx context.Context, id string, input CustomerInput) (CustomerView, error) {
	row, err := s.repository.GetCustomer(ctx, strings.TrimSpace(id))
	if err != nil {
		return CustomerView{}, err
	}
	if err := row.Update(domain.CustomerDraft{Name: input.Name, Phone: input.Phone, Email: input.Email, Address: input.Address, Notes: input.Notes}, s.now()); err != nil {
		return CustomerView{}, err
	}
	if err := s.repository.SaveCustomer(ctx, row); err != nil {
		return CustomerView{}, err
	}
	return customerView(row), nil
}
func (s *CustomersService) Archive(ctx context.Context, id string) (CustomerView, error) {
	return s.setActive(ctx, id, false)
}
func (s *CustomersService) Reactivate(ctx context.Context, id string) (CustomerView, error) {
	return s.setActive(ctx, id, true)
}
func (s *CustomersService) Delete(ctx context.Context, id string) error {
	repository, ok := s.repository.(interface {
		DeleteCustomer(context.Context, string) error
	})
	if !ok {
		return fmt.Errorf("customer deletion is not supported")
	}
	return repository.DeleteCustomer(ctx, strings.TrimSpace(id))
}
func (s *CustomersService) setActive(ctx context.Context, id string, active bool) (CustomerView, error) {
	row, err := s.repository.GetCustomer(ctx, strings.TrimSpace(id))
	if err != nil {
		return CustomerView{}, err
	}
	row.Active = active
	row.UpdatedAt = s.now().UTC()
	if err := s.repository.SaveCustomer(ctx, row); err != nil {
		return CustomerView{}, err
	}
	return customerView(row), nil
}
func customerView(c domain.Customer) CustomerView {
	return CustomerView{ID: c.ID, Name: c.Name, Phone: c.Phone, Email: c.Email, Address: c.Address, Notes: c.Notes, Active: c.Active, CreatedAt: c.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: c.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func randomID(prefix string) (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + strings.ToUpper(hex.EncodeToString(b)), nil
}
