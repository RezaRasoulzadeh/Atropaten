package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrMachineNotFound = errors.New("machine not found")

const (
	RatePerUnit   = "unit"
	RatePerMinute = "minute"
	RatePerHour   = "hour"
)

type Machine struct {
	ID            string
	Name          string
	Code          string
	Category      string
	RateBasis     string
	RateRial      int64
	SetupCostRial int64
	Notes         string
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type MachineDraft struct {
	Name          string
	Code          string
	Category      string
	RateBasis     string
	RateRial      int64
	SetupCostRial int64
	Notes         string
}

func SupportedRateBases() []string { return []string{RatePerUnit, RatePerMinute, RatePerHour} }

func NewMachine(id string, draft MachineDraft, now time.Time) (Machine, error) {
	machine := Machine{ID: strings.TrimSpace(id), Name: strings.TrimSpace(draft.Name), Code: strings.TrimSpace(draft.Code), Category: strings.TrimSpace(draft.Category), RateBasis: strings.ToLower(strings.TrimSpace(draft.RateBasis)), RateRial: draft.RateRial, SetupCostRial: draft.SetupCostRial, Notes: strings.TrimSpace(draft.Notes), Active: true, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if err := machine.Validate(); err != nil {
		return Machine{}, err
	}
	return machine, nil
}

func (m *Machine) Update(draft MachineDraft, now time.Time) error {
	updated, err := NewMachine(m.ID, draft, m.CreatedAt)
	if err != nil {
		return err
	}
	updated.Active = m.Active
	updated.UpdatedAt = now.UTC()
	*m = updated
	return nil
}

func (m Machine) Validate() error {
	if strings.TrimSpace(m.ID) == "" {
		return validationError("id", "is required")
	}
	if strings.TrimSpace(m.Name) == "" {
		return validationError("name", "is required")
	}
	validBasis := false
	for _, basis := range SupportedRateBases() {
		if m.RateBasis == basis {
			validBasis = true
			break
		}
	}
	if !validBasis {
		return validationError("rateBasis", "must be unit, minute, or hour")
	}
	if m.RateRial < 0 {
		return validationError("rateRial", "cannot be negative")
	}
	if m.SetupCostRial < 0 {
		return validationError("setupCostRial", "cannot be negative")
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	return nil
}
