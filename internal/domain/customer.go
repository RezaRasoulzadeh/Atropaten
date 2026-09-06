package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrCustomerNotFound = errors.New("customer not found")

type Customer struct {
	ID        string
	Name      string
	Phone     string
	Email     string
	Address   string
	Notes     string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CustomerDraft struct {
	Name, Phone, Email, Address, Notes string
}

func NewCustomer(id string, draft CustomerDraft, now time.Time) (Customer, error) {
	c := Customer{ID: strings.TrimSpace(id), Active: true, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	c.apply(draft)
	if err := c.Validate(); err != nil {
		return Customer{}, err
	}
	return c, nil
}

func (c *Customer) Update(draft CustomerDraft, now time.Time) error {
	c.apply(draft)
	c.UpdatedAt = now.UTC()
	return c.Validate()
}

func (c *Customer) apply(d CustomerDraft) {
	c.Name = strings.TrimSpace(d.Name)
	c.Phone = strings.TrimSpace(d.Phone)
	c.Email = strings.TrimSpace(d.Email)
	c.Address = strings.TrimSpace(d.Address)
	c.Notes = strings.TrimSpace(d.Notes)
}

func (c Customer) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return validationError("id", "is required")
	}
	if c.Name == "" {
		return validationError("name", "is required")
	}
	if c.CreatedAt.IsZero() || c.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	return nil
}
