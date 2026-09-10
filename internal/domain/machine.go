package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMachineNotFound        = errors.New("machine not found")
	ErrMachineDeleteProtected = errors.New("machine has production history; archive it instead")
)

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
	Rates         []MachineRate
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type MachineRate struct {
	ID            string
	Name          string
	SelectorValue string
	RateBasis     string
	RateRial      int64
	SetupCostRial int64
	Active        bool
}

type MachineDraft struct {
	Name          string
	Code          string
	Category      string
	RateBasis     string
	RateRial      int64
	SetupCostRial int64
	Notes         string
	Rates         []MachineRate
}

func SupportedRateBases() []string { return []string{RatePerUnit, RatePerMinute, RatePerHour} }

func NewMachine(id string, draft MachineDraft, now time.Time) (Machine, error) {
	rates := append([]MachineRate(nil), draft.Rates...)
	if len(rates) == 0 {
		rates = []MachineRate{{ID: "default", Name: "Standard", RateBasis: draft.RateBasis, RateRial: draft.RateRial, SetupCostRial: draft.SetupCostRial, Active: true}}
	}
	for index := range rates {
		rates[index].ID = strings.TrimSpace(rates[index].ID)
		if rates[index].ID == "" {
			rates[index].ID = fmt.Sprintf("rate-%d", index+1)
		}
		if rates[index].Name == "" {
			rates[index].Name = fmt.Sprintf("Rate %d", index+1)
		}
		rates[index].RateBasis = strings.ToLower(strings.TrimSpace(rates[index].RateBasis))
		rates[index].Name = strings.TrimSpace(rates[index].Name)
		rates[index].SelectorValue = strings.TrimSpace(rates[index].SelectorValue)
	}
	activeRate := false
	for _, rate := range rates {
		if rate.Active {
			activeRate = true
			break
		}
	}
	if !activeRate {
		rates[0].Active = true
	}
	primary := rates[0]
	for _, rate := range rates {
		if rate.Active {
			primary = rate
			break
		}
	}
	machine := Machine{ID: strings.TrimSpace(id), Name: strings.TrimSpace(draft.Name), Code: strings.TrimSpace(draft.Code), Category: strings.TrimSpace(draft.Category), RateBasis: primary.RateBasis, RateRial: primary.RateRial, SetupCostRial: primary.SetupCostRial, Notes: strings.TrimSpace(draft.Notes), Active: true, Rates: rates, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
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
	seenRateIDs := map[string]bool{}
	for index, rate := range m.Rates {
		if err := rate.Validate(); err != nil {
			return fmt.Errorf("rates[%d]: %w", index, err)
		}
		if seenRateIDs[rate.ID] {
			return validationError(fmt.Sprintf("rates[%d].id", index), "must be unique")
		}
		seenRateIDs[rate.ID] = true
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		return validationError("timestamps", "are required")
	}
	return nil
}

func (r MachineRate) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return validationError("id", "is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return validationError("name", "is required")
	}
	validBasis := false
	for _, basis := range SupportedRateBases() {
		if r.RateBasis == basis {
			validBasis = true
			break
		}
	}
	if !validBasis {
		return validationError("rateBasis", "must be unit, minute, or hour")
	}
	if r.RateRial < 0 {
		return validationError("rateRial", "cannot be negative")
	}
	if r.SetupCostRial < 0 {
		return validationError("setupCostRial", "cannot be negative")
	}
	return nil
}

func (m Machine) Rate(rateID string) (MachineRate, bool) {
	for _, rate := range m.Rates {
		if rate.Active && rate.ID == strings.TrimSpace(rateID) {
			return rate, true
		}
	}
	if strings.TrimSpace(rateID) == "" {
		for _, rate := range m.Rates {
			if rate.Active {
				return rate, true
			}
		}
	}
	if len(m.Rates) == 0 && rateID == "" {
		return MachineRate{ID: "legacy", Name: "Standard", RateBasis: m.RateBasis, RateRial: m.RateRial, SetupCostRial: m.SetupCostRial, Active: true}, true
	}
	return MachineRate{}, false
}
