package main

import (
	"Atropaten/internal/application"
	"fmt"
)

type AccountDTO struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Active      bool   `json:"active"`
	System      bool   `json:"system"`
	BalanceRial int64  `json:"balanceRial"`
}
type FinancialAccountDTO struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	LedgerAccountID string `json:"ledgerAccountId"`
	Details         string `json:"details"`
	Active          bool   `json:"active"`
	BalanceRial     int64  `json:"balanceRial"`
}
type JournalLineDTO struct {
	ID         string `json:"id"`
	AccountID  string `json:"accountId"`
	PartyType  string `json:"partyType"`
	PartyID    string `json:"partyId"`
	Memo       string `json:"memo"`
	Position   int    `json:"position"`
	DebitRial  int64  `json:"debitRial"`
	CreditRial int64  `json:"creditRial"`
}
type JournalEntryDTO struct {
	ID           string           `json:"id"`
	EntryNumber  string           `json:"entryNumber"`
	Description  string           `json:"description"`
	SourceType   string           `json:"sourceType"`
	SourceID     string           `json:"sourceId"`
	ReversalOfID string           `json:"reversalOfId"`
	PostedAt     string           `json:"postedAt"`
	CreatedAt    string           `json:"createdAt"`
	Lines        []JournalLineDTO `json:"lines"`
}
type PaymentAllocationInput struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	AmountRial int64  `json:"amountRial"`
}
type PaymentInputDTO struct {
	ID                 string                   `json:"id"`
	Direction          string                   `json:"direction"`
	Method             string                   `json:"method"`
	FinancialAccountID string                   `json:"financialAccountId"`
	CustomerID         string                   `json:"customerId"`
	SupplierID         string                   `json:"supplierId"`
	Reference          string                   `json:"reference"`
	Notes              string                   `json:"notes"`
	PostedAt           string                   `json:"postedAt"`
	IdempotencyKey     string                   `json:"idempotencyKey"`
	AmountRial         int64                    `json:"amountRial"`
	Allocations        []PaymentAllocationInput `json:"allocations"`
}
type PaymentAllocationDTO struct {
	ID         string `json:"id"`
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Position   int    `json:"position"`
	AmountRial int64  `json:"amountRial"`
	Reversed   bool   `json:"reversed"`
}
type PaymentDTO struct {
	ID                 string                 `json:"id"`
	PaymentNumber      string                 `json:"paymentNumber"`
	Direction          string                 `json:"direction"`
	Method             string                 `json:"method"`
	FinancialAccountID string                 `json:"financialAccountId"`
	CustomerID         string                 `json:"customerId"`
	SupplierID         string                 `json:"supplierId"`
	Reference          string                 `json:"reference"`
	Notes              string                 `json:"notes"`
	Status             string                 `json:"status"`
	JournalEntryID     string                 `json:"journalEntryId"`
	PostedAt           string                 `json:"postedAt"`
	CreatedAt          string                 `json:"createdAt"`
	AmountRial         int64                  `json:"amountRial"`
	Allocations        []PaymentAllocationDTO `json:"allocations"`
}

func (a *App) accountingService() (*application.AccountingService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.accounting == nil {
		return nil, fmt.Errorf("accounting service is not initialized")
	}
	return a.accounting, nil
}
func (a *App) ListAccounts() ([]AccountDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return nil, e
	}
	v, e := s.Accounts(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]AccountDTO, 0, len(v))
	for _, x := range v {
		out = append(out, AccountDTO{x.ID, x.Code, x.Name, x.Type, x.Active, x.System, x.BalanceRial})
	}
	return out, nil
}
func (a *App) ListFinancialAccounts() ([]FinancialAccountDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return nil, e
	}
	v, e := s.FinancialAccounts(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]FinancialAccountDTO, 0, len(v))
	for _, x := range v {
		out = append(out, FinancialAccountDTO{x.ID, x.Name, x.Type, x.LedgerAccountID, x.Details, x.Active, x.BalanceRial})
	}
	return out, nil
}
func (a *App) ListJournalEntries() ([]JournalEntryDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return nil, e
	}
	v, e := s.Journal(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]JournalEntryDTO, 0, len(v))
	for _, x := range v {
		j := JournalEntryDTO{ID: x.ID, EntryNumber: x.EntryNumber, Description: x.Description, SourceType: x.SourceType, SourceID: x.SourceID, ReversalOfID: x.ReversalOfID, PostedAt: x.PostedAt, CreatedAt: x.CreatedAt}
		for _, l := range x.Lines {
			j.Lines = append(j.Lines, JournalLineDTO{l.ID, l.AccountID, l.PartyType, l.PartyID, l.Memo, l.Position, l.DebitRial, l.CreditRial})
		}
		out = append(out, j)
	}
	return out, nil
}
func (a *App) ListPayments() ([]PaymentDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return nil, e
	}
	v, e := s.Payments(a.materialContext())
	if e != nil {
		return nil, e
	}
	return paymentDTOs(v), nil
}
func (a *App) CreatePayment(in PaymentInputDTO) (PaymentDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return PaymentDTO{}, e
	}
	v, e := s.CreatePayment(a.materialContext(), application.PaymentInput{ID: in.ID, Direction: in.Direction, Method: in.Method, FinancialAccountID: in.FinancialAccountID, CustomerID: in.CustomerID, SupplierID: in.SupplierID, Reference: in.Reference, Notes: in.Notes, PostedAt: in.PostedAt, IdempotencyKey: in.IdempotencyKey, AmountRial: in.AmountRial, Allocations: paymentAllocationInputs(in.Allocations)})
	return paymentDTO(v), e
}
func (a *App) ReversePayment(id, key string) (PaymentDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return PaymentDTO{}, e
	}
	v, e := s.ReversePayment(a.materialContext(), id, key)
	return paymentDTO(v), e
}
func paymentAllocationInputs(v []PaymentAllocationInput) []application.PaymentAllocationInput {
	out := make([]application.PaymentAllocationInput, 0, len(v))
	for _, x := range v {
		out = append(out, application.PaymentAllocationInput{TargetType: x.TargetType, TargetID: x.TargetID, AmountRial: x.AmountRial})
	}
	return out
}
func paymentDTO(v application.PaymentView) PaymentDTO {
	out := PaymentDTO{ID: v.ID, PaymentNumber: v.PaymentNumber, Direction: v.Direction, Method: v.Method, FinancialAccountID: v.FinancialAccountID, CustomerID: v.CustomerID, SupplierID: v.SupplierID, Reference: v.Reference, Notes: v.Notes, Status: v.Status, JournalEntryID: v.JournalEntryID, PostedAt: v.PostedAt, CreatedAt: v.CreatedAt, AmountRial: v.AmountRial}
	for _, a := range v.Allocations {
		out.Allocations = append(out.Allocations, PaymentAllocationDTO{a.ID, a.TargetType, a.TargetID, a.Position, a.AmountRial, a.Reversed})
	}
	return out
}
func paymentDTOs(v []application.PaymentView) []PaymentDTO {
	out := make([]PaymentDTO, 0, len(v))
	for _, p := range v {
		out = append(out, paymentDTO(p))
	}
	return out
}
