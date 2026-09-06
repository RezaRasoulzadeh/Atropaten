package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type MachineRepository interface {
	ListMachines(context.Context, bool) ([]domain.Machine, error)
	GetMachine(context.Context, string) (domain.Machine, error)
	SaveMachine(context.Context, domain.Machine) error
}

type MachineInput struct {
	Name          string
	Code          string
	Category      string
	RateBasis     string
	RateRial      int64
	SetupCostRial int64
	Notes         string
}

type MachineView struct {
	ID            string
	Name          string
	Code          string
	Category      string
	RateBasis     string
	RateRial      int64
	SetupCostRial int64
	Notes         string
	Active        bool
	CreatedAt     string
	UpdatedAt     string
}

type MachinesService struct {
	repository MachineRepository
	now        func() time.Time
}

func NewMachinesService(repository MachineRepository) *MachinesService {
	return &MachinesService{repository: repository, now: time.Now}
}

func (s *MachinesService) List(ctx context.Context, includeArchived bool) ([]MachineView, error) {
	machines, err := s.repository.ListMachines(ctx, includeArchived)
	if err != nil {
		return nil, err
	}
	views := make([]MachineView, 0, len(machines))
	for _, machine := range machines {
		views = append(views, machineView(machine))
	}
	return views, nil
}

func (s *MachinesService) Get(ctx context.Context, id string) (MachineView, error) {
	machine, err := s.repository.GetMachine(ctx, strings.TrimSpace(id))
	if err != nil {
		return MachineView{}, err
	}
	return machineView(machine), nil
}

func (s *MachinesService) Create(ctx context.Context, input MachineInput) (MachineView, error) {
	id, err := newID("MAC-")
	if err != nil {
		return MachineView{}, fmt.Errorf("create machine id: %w", err)
	}
	machine, err := domain.NewMachine(id, machineDraft(input), s.now())
	if err != nil {
		return MachineView{}, err
	}
	if err := s.repository.SaveMachine(ctx, machine); err != nil {
		return MachineView{}, err
	}
	return machineView(machine), nil
}

func (s *MachinesService) Update(ctx context.Context, id string, input MachineInput) (MachineView, error) {
	machine, err := s.repository.GetMachine(ctx, strings.TrimSpace(id))
	if err != nil {
		return MachineView{}, err
	}
	if err := machine.Update(machineDraft(input), s.now()); err != nil {
		return MachineView{}, err
	}
	if err := s.repository.SaveMachine(ctx, machine); err != nil {
		return MachineView{}, err
	}
	return machineView(machine), nil
}

func (s *MachinesService) Archive(ctx context.Context, id string) (MachineView, error) {
	return s.setActive(ctx, id, false)
}
func (s *MachinesService) Reactivate(ctx context.Context, id string) (MachineView, error) {
	return s.setActive(ctx, id, true)
}

func (s *MachinesService) setActive(ctx context.Context, id string, active bool) (MachineView, error) {
	machine, err := s.repository.GetMachine(ctx, strings.TrimSpace(id))
	if err != nil {
		return MachineView{}, err
	}
	machine.Active = active
	machine.UpdatedAt = s.now().UTC()
	if err := s.repository.SaveMachine(ctx, machine); err != nil {
		return MachineView{}, err
	}
	return machineView(machine), nil
}

func machineDraft(input MachineInput) domain.MachineDraft {
	return domain.MachineDraft{Name: input.Name, Code: input.Code, Category: input.Category, RateBasis: input.RateBasis, RateRial: input.RateRial, SetupCostRial: input.SetupCostRial, Notes: input.Notes}
}

func machineView(machine domain.Machine) MachineView {
	return MachineView{ID: machine.ID, Name: machine.Name, Code: machine.Code, Category: machine.Category, RateBasis: machine.RateBasis, RateRial: machine.RateRial, SetupCostRial: machine.SetupCostRial, Notes: machine.Notes, Active: machine.Active, CreatedAt: machine.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: machine.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
