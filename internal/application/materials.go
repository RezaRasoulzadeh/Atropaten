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

type MaterialRepository interface {
		List(context.Context, bool) ([]domain.Material, error)
		Get(context.Context, string) (domain.Material, error)
		Create(context.Context, domain.Material) error
		Update(context.Context, domain.Material) error
}

type MaterialInput struct {
	Name                string
	SKU                 string
	Category            string
	PurchaseUnit        string
	ConsumptionUnit     string
	ConversionFactor    string
	PhysicalStock       string
	ReorderLevel        string
	AverageUnitCostRial int64
	PreferredSupplier   string
	Notes               string
}

type MaterialView struct {
	ID                    string
	Name                  string
	SKU                   string
	Category              string
	PurchaseUnit          string
	ConsumptionUnit       string
	ConversionFactor      string
	PhysicalStock         string
	ReorderLevel          string
	AverageUnitCostRial   int64
	PreferredSupplier     string
	Notes                 string
	Active                bool
	LowStock              bool
	CreatedAt             string
	UpdatedAt             string
}

type MaterialsService struct {
	repository MaterialRepository
	now        func() time.Time
}

func NewMaterialsService(repository MaterialRepository) *MaterialsService {
	return &MaterialsService{repository: repository, now: time.Now}
}

func (s *MaterialsService) List(ctx context.Context, includeArchived bool) ([]MaterialView, error) {
	materials, err := s.repository.List(ctx, includeArchived)
	if err != nil {
		return nil, err
	}
	views := make([]MaterialView, 0, len(materials))
	for _, material := range materials {
		views = append(views, toView(material))
	}
	return views, nil
}

func (s *MaterialsService) Get(ctx context.Context, id string) (MaterialView, error) {
	material, err := s.repository.Get(ctx, strings.TrimSpace(id))
	if err != nil {
		return MaterialView{}, err
	}
	return toView(material), nil
}

func (s *MaterialsService) Create(ctx context.Context, input MaterialInput) (MaterialView, error) {
	draft, err := parseDraft(input)
	if err != nil {
		return MaterialView{}, err
	}
	id, err := newMaterialID()
	if err != nil {
		return MaterialView{}, fmt.Errorf("create material id: %w", err)
	}
	material, err := domain.NewMaterial(id, draft, s.now())
	if err != nil {
		return MaterialView{}, err
	}
	if err := s.repository.Create(ctx, material); err != nil {
		return MaterialView{}, err
	}
	return toView(material), nil
}

func (s *MaterialsService) Update(ctx context.Context, id string, input MaterialInput) (MaterialView, error) {
	material, err := s.repository.Get(ctx, strings.TrimSpace(id))
	if err != nil {
		return MaterialView{}, err
	}
	draft, err := parseDraft(input)
	if err != nil {
		return MaterialView{}, err
	}
	if err := material.Update(draft, s.now()); err != nil {
		return MaterialView{}, err
	}
	if err := s.repository.Update(ctx, material); err != nil {
		return MaterialView{}, err
	}
	return toView(material), nil
}

func (s *MaterialsService) Archive(ctx context.Context, id string) (MaterialView, error) {
	return s.setActive(ctx, id, false)
}

func (s *MaterialsService) Reactivate(ctx context.Context, id string) (MaterialView, error) {
	return s.setActive(ctx, id, true)
}

func (s *MaterialsService) setActive(ctx context.Context, id string, active bool) (MaterialView, error) {
	material, err := s.repository.Get(ctx, strings.TrimSpace(id))
	if err != nil {
		return MaterialView{}, err
	}
	material.Active = active
	material.UpdatedAt = s.now().UTC()
	if err := s.repository.Update(ctx, material); err != nil {
		return MaterialView{}, err
	}
	return toView(material), nil
}

func parseDraft(input MaterialInput) (domain.MaterialDraft, error) {
	conversion, err := domain.ParseQuantity(input.ConversionFactor)
	if err != nil {
		return domain.MaterialDraft{}, fmt.Errorf("conversionFactor: %w", err)
	}
	stock, err := domain.ParseQuantity(input.PhysicalStock)
	if err != nil {
		return domain.MaterialDraft{}, fmt.Errorf("physicalStock: %w", err)
	}
	reorder, err := domain.ParseQuantity(input.ReorderLevel)
	if err != nil {
		return domain.MaterialDraft{}, fmt.Errorf("reorderLevel: %w", err)
	}
	if input.AverageUnitCostRial < 0 {
		return domain.MaterialDraft{}, domain.ValidationError{Field: "averageUnitCostRial", Message: "cannot be negative"}
	}
	return domain.MaterialDraft{
		Name: input.Name, SKU: input.SKU, Category: input.Category,
		PurchaseUnit: input.PurchaseUnit, ConsumptionUnit: input.ConsumptionUnit,
		ConversionFactor: conversion, PhysicalStock: stock, ReorderLevel: reorder,
		AverageUnitCostRial: input.AverageUnitCostRial,
		PreferredSupplier: input.PreferredSupplier, Notes: input.Notes,
	}, nil
}

func toView(material domain.Material) MaterialView {
	return MaterialView{
		ID: material.ID, Name: material.Name, SKU: material.SKU, Category: material.Category,
		PurchaseUnit: material.PurchaseUnit, ConsumptionUnit: material.ConsumptionUnit,
		ConversionFactor: material.ConversionFactor.String(), PhysicalStock: material.PhysicalStock.String(),
		ReorderLevel: material.ReorderLevel.String(), AverageUnitCostRial: material.AverageUnitCostRial,
		PreferredSupplier: material.PreferredSupplier, Notes: material.Notes, Active: material.Active,
		LowStock: material.LowStock(), CreatedAt: material.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: material.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func newMaterialID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "MAT-" + hex.EncodeToString(bytes), nil
}
