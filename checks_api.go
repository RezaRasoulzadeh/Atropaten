package main

import (
	"Atropaten/internal/application"
	"fmt"
)

type CheckDTO struct {
	ID                 string `json:"id"`
	CheckNumber        string `json:"checkNumber"`
	Direction          string `json:"direction"`
	Bank               string `json:"bank"`
	Branch             string `json:"branch"`
	AccountDescriptor  string `json:"accountDescriptor"`
	PayerPayee         string `json:"payerPayee"`
	CustomerID         string `json:"customerId"`
	SupplierID         string `json:"supplierId"`
	SourceType         string `json:"sourceType"`
	SourceID           string `json:"sourceId"`
	FinancialAccountID string `json:"financialAccountId"`
	Notes              string `json:"notes"`
	Status             string `json:"status"`
	IssueDate          string `json:"issueDate"`
	DueDate            string `json:"dueDate"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	AmountRial         int64  `json:"amountRial"`
}
type CheckInputDTO struct {
	ID                 string `json:"id"`
	CheckNumber        string `json:"checkNumber"`
	Direction          string `json:"direction"`
	Bank               string `json:"bank"`
	Branch             string `json:"branch"`
	AccountDescriptor  string `json:"accountDescriptor"`
	PayerPayee         string `json:"payerPayee"`
	CustomerID         string `json:"customerId"`
	SupplierID         string `json:"supplierId"`
	SourceType         string `json:"sourceType"`
	SourceID           string `json:"sourceId"`
	FinancialAccountID string `json:"financialAccountId"`
	Notes              string `json:"notes"`
	IssueDate          string `json:"issueDate"`
	DueDate            string `json:"dueDate"`
	Status             string `json:"status"`
	AmountRial         int64  `json:"amountRial"`
}
type CheckEventDTO struct {
	ID             string `json:"id"`
	CheckID        string `json:"checkId"`
	FromStatus     string `json:"fromStatus"`
	ToStatus       string `json:"toStatus"`
	Note           string `json:"note"`
	JournalEntryID string `json:"journalEntryId"`
	OccurredAt     string `json:"occurredAt"`
}

func (a *App) checksService() (*application.ChecksService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.checks == nil {
		return nil, fmt.Errorf("checks service is not initialized")
	}
	return a.checks, nil
}
func (a *App) ListChecks(direction, status string) ([]CheckDTO, error) {
	s, err := a.checksService()
	if err != nil {
		return nil, err
	}
	rows, err := s.List(a.materialContext(), direction, status)
	if err != nil {
		return nil, err
	}
	out := make([]CheckDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, checkDTO(v))
	}
	return out, nil
}
func (a *App) GetCheck(id string) (CheckDTO, error) {
	s, err := a.checksService()
	if err != nil {
		return CheckDTO{}, err
	}
	v, err := s.Get(a.materialContext(), id)
	return checkDTO(v), err
}
func (a *App) CreateCheck(in CheckInputDTO) (CheckDTO, error) {
	s, err := a.checksService()
	if err != nil {
		return CheckDTO{}, err
	}
	v, err := s.Create(a.materialContext(), application.CheckInput{ID: in.ID, CheckNumber: in.CheckNumber, Direction: in.Direction, Bank: in.Bank, Branch: in.Branch, AccountDescriptor: in.AccountDescriptor, PayerPayee: in.PayerPayee, CustomerID: in.CustomerID, SupplierID: in.SupplierID, SourceType: in.SourceType, SourceID: in.SourceID, FinancialAccountID: in.FinancialAccountID, Notes: in.Notes, IssueDate: in.IssueDate, DueDate: in.DueDate, Status: in.Status, AmountRial: in.AmountRial})
	return checkDTO(v), err
}
func (a *App) TransitionCheck(id, to, note, key string) (CheckDTO, error) {
	s, err := a.checksService()
	if err != nil {
		return CheckDTO{}, err
	}
	v, err := s.Transition(a.materialContext(), id, to, note, key)
	return checkDTO(v), err
}
func (a *App) DeleteDraftCheck(id string) error {
	s, err := a.checksService()
	if err != nil {
		return err
	}
	return s.DeleteDraft(a.materialContext(), id)
}
func (a *App) ListCheckEvents(id string) ([]CheckEventDTO, error) {
	s, err := a.checksService()
	if err != nil {
		return nil, err
	}
	rows, err := s.History(a.materialContext(), id)
	if err != nil {
		return nil, err
	}
	out := make([]CheckEventDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, CheckEventDTO{v.ID, v.CheckID, v.FromStatus, v.ToStatus, v.Note, v.JournalEntryID, v.OccurredAt})
	}
	return out, nil
}
func checkDTO(v application.CheckView) CheckDTO {
	return CheckDTO{ID: v.ID, CheckNumber: v.CheckNumber, Direction: v.Direction, Bank: v.Bank, Branch: v.Branch, AccountDescriptor: v.AccountDescriptor, PayerPayee: v.PayerPayee, CustomerID: v.CustomerID, SupplierID: v.SupplierID, SourceType: v.SourceType, SourceID: v.SourceID, FinancialAccountID: v.FinancialAccountID, Notes: v.Notes, Status: v.Status, IssueDate: v.IssueDate, DueDate: v.DueDate, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, AmountRial: v.AmountRial}
}
