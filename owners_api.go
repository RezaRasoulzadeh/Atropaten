package main

import (
	"fmt"

	"Atropaten/internal/application"
)

type OwnerInputDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	Notes            string `json:"notes"`
	OwnershipBPS     int64  `json:"ownershipBps"`
	ProfitSharingBPS int64  `json:"profitSharingBps"`
}
type OwnerDTO struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	Phone                   string `json:"phone"`
	Email                   string `json:"email"`
	Notes                   string `json:"notes"`
	Active                  bool   `json:"active"`
	OwnershipBPS            int64  `json:"ownershipBps"`
	ProfitSharingBPS        int64  `json:"profitSharingBps"`
	CapitalContributedRial  int64  `json:"capitalContributedRial"`
	DrawingsRial            int64  `json:"drawingsRial"`
	CurrentBalanceRial      int64  `json:"currentBalanceRial"`
	LoanPayableRial         int64  `json:"loanPayableRial"`
	LoanReceivableRial      int64  `json:"loanReceivableRial"`
	AllocatedProfitLossRial int64  `json:"allocatedProfitLossRial"`
	CreatedAt               string `json:"createdAt"`
	UpdatedAt               string `json:"updatedAt"`
}
type OwnerShareInputDTO struct {
	OwnershipBPS     int64  `json:"ownershipBps"`
	ProfitSharingBPS int64  `json:"profitSharingBps"`
	Reason           string `json:"reason"`
}
type OwnerTransactionInputDTO struct {
	ID                 string `json:"id"`
	OwnerID            string `json:"ownerId"`
	Type               string `json:"type"`
	FinancialAccountID string `json:"financialAccountId"`
	CategoryAccountID  string `json:"categoryAccountId"`
	Description        string `json:"description"`
	Notes              string `json:"notes"`
	OccurredAt         string `json:"occurredAt"`
	IdempotencyKey     string `json:"idempotencyKey"`
	AmountRial         int64  `json:"amountRial"`
}
type OwnerTransactionDTO struct {
	ID                 string `json:"id"`
	TransactionNumber  string `json:"transactionNumber"`
	OwnerID            string `json:"ownerId"`
	Type               string `json:"type"`
	FinancialAccountID string `json:"financialAccountId"`
	CategoryAccountID  string `json:"categoryAccountId"`
	Description        string `json:"description"`
	Notes              string `json:"notes"`
	Status             string `json:"status"`
	JournalEntryID     string `json:"journalEntryId"`
	IdempotencyKey     string `json:"idempotencyKey"`
	OccurredAt         string `json:"occurredAt"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	AmountRial         int64  `json:"amountRial"`
}
type FiscalPeriodInputDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate"`
	Notes          string `json:"notes"`
	IdempotencyKey string `json:"idempotencyKey"`
}
type ProfitAllocationDTO struct {
	ID               string `json:"id"`
	PeriodID         string `json:"periodId"`
	OwnerID          string `json:"ownerId"`
	Position         int    `json:"position"`
	ProfitSharingBPS int64  `json:"profitSharingBps"`
	AmountRial       int64  `json:"amountRial"`
}
type FiscalPeriodDTO struct {
	ID                    string                `json:"id"`
	Name                  string                `json:"name"`
	Status                string                `json:"status"`
	Notes                 string                `json:"notes"`
	ClosingJournalEntryID string                `json:"closingJournalEntryId"`
	IdempotencyKey        string                `json:"idempotencyKey"`
	StartDate             string                `json:"startDate"`
	EndDate               string                `json:"endDate"`
	ClosedAt              string                `json:"closedAt"`
	CreatedAt             string                `json:"createdAt"`
	UpdatedAt             string                `json:"updatedAt"`
	RevenueRial           int64                 `json:"revenueRial"`
	COGSRial              int64                 `json:"cogsRial"`
	ExpensesRial          int64                 `json:"expensesRial"`
	ProfitLossRial        int64                 `json:"profitLossRial"`
	Allocations           []ProfitAllocationDTO `json:"allocations"`
	PreviewAllocations    []ProfitAllocationDTO `json:"previewAllocations"`
}

func (a *App) ownersService() (*application.OwnersService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.owners == nil {
		return nil, fmt.Errorf("owners service is not initialized")
	}
	return a.owners, nil
}
func (a *App) ListOwners(activeOnly bool) ([]OwnerDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return nil, e
	}
	v, e := s.Owners(a.materialContext(), activeOnly)
	if e != nil {
		return nil, e
	}
	out := make([]OwnerDTO, 0, len(v))
	for _, x := range v {
		out = append(out, ownerDTO(x))
	}
	return out, nil
}
func (a *App) GetOwner(id string) (OwnerDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return OwnerDTO{}, e
	}
	v, e := s.Owner(a.materialContext(), id)
	if e != nil {
		return OwnerDTO{}, e
	}
	return ownerDTO(v), nil
}
func (a *App) CreateOwner(i OwnerInputDTO) (OwnerDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return OwnerDTO{}, e
	}
	v, e := s.CreateOwner(a.materialContext(), application.OwnerInput{ID: i.ID, Name: i.Name, Phone: i.Phone, Email: i.Email, Notes: i.Notes, OwnershipBPS: i.OwnershipBPS, ProfitSharingBPS: i.ProfitSharingBPS})
	if e != nil {
		return OwnerDTO{}, e
	}
	return ownerDTO(v), nil
}
func (a *App) UpdateOwnerShares(id string, i OwnerShareInputDTO) (OwnerDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return OwnerDTO{}, e
	}
	v, e := s.UpdateShares(a.materialContext(), id, i.OwnershipBPS, i.ProfitSharingBPS, i.Reason)
	if e != nil {
		return OwnerDTO{}, e
	}
	return ownerDTO(v), nil
}
func (a *App) ArchiveOwner(id string) error {
	s, e := a.ownersService()
	if e != nil {
		return e
	}
	return s.Archive(a.materialContext(), id)
}
func (a *App) ReactivateOwner(id string) error {
	s, e := a.ownersService()
	if e != nil {
		return e
	}
	return s.Reactivate(a.materialContext(), id)
}
func (a *App) DeleteOwner(id string) error {
	s, e := a.ownersService()
	if e != nil {
		return e
	}
	return s.Delete(a.materialContext(), id)
}
func (a *App) ListOwnerTransactions(ownerID string) ([]OwnerTransactionDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return nil, e
	}
	v, e := s.Transactions(a.materialContext(), ownerID)
	if e != nil {
		return nil, e
	}
	out := make([]OwnerTransactionDTO, 0, len(v))
	for _, x := range v {
		out = append(out, ownerTransactionDTO(x))
	}
	return out, nil
}
func (a *App) CreateOwnerTransaction(i OwnerTransactionInputDTO) (OwnerTransactionDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return OwnerTransactionDTO{}, e
	}
	v, e := s.CreateTransaction(a.materialContext(), application.OwnerTransactionInput{ID: i.ID, OwnerID: i.OwnerID, Type: i.Type, FinancialAccountID: i.FinancialAccountID, CategoryAccountID: i.CategoryAccountID, Description: i.Description, Notes: i.Notes, OccurredAt: i.OccurredAt, IdempotencyKey: i.IdempotencyKey, AmountRial: i.AmountRial})
	if e != nil {
		return OwnerTransactionDTO{}, e
	}
	return ownerTransactionDTO(v), nil
}
func (a *App) ReverseOwnerTransaction(id, key string) (OwnerTransactionDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return OwnerTransactionDTO{}, e
	}
	v, e := s.ReverseTransaction(a.materialContext(), id, key)
	if e != nil {
		return OwnerTransactionDTO{}, e
	}
	return ownerTransactionDTO(v), nil
}
func (a *App) ListFiscalPeriods() ([]FiscalPeriodDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return nil, e
	}
	v, e := s.Periods(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]FiscalPeriodDTO, 0, len(v))
	for _, x := range v {
		out = append(out, periodDTO(x))
	}
	return out, nil
}
func (a *App) GetFiscalPeriod(id string) (FiscalPeriodDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return FiscalPeriodDTO{}, e
	}
	v, e := s.Period(a.materialContext(), id)
	if e != nil {
		return FiscalPeriodDTO{}, e
	}
	return periodDTO(v), nil
}
func (a *App) PreviewFiscalPeriod(id string) (FiscalPeriodDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return FiscalPeriodDTO{}, e
	}
	v, e := s.PreviewPeriod(a.materialContext(), id)
	if e != nil {
		return FiscalPeriodDTO{}, e
	}
	return periodDTO(v), nil
}
func (a *App) CreateFiscalPeriod(i FiscalPeriodInputDTO) (FiscalPeriodDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return FiscalPeriodDTO{}, e
	}
	v, e := s.CreatePeriod(a.materialContext(), i.ID, i.Name, i.StartDate, i.EndDate, i.Notes, i.IdempotencyKey)
	if e != nil {
		return FiscalPeriodDTO{}, e
	}
	return periodDTO(v), nil
}
func (a *App) CloseFiscalPeriod(id, key string) (FiscalPeriodDTO, error) {
	s, e := a.ownersService()
	if e != nil {
		return FiscalPeriodDTO{}, e
	}
	v, e := s.ClosePeriod(a.materialContext(), id, key)
	if e != nil {
		return FiscalPeriodDTO{}, e
	}
	return periodDTO(v), nil
}

func ownerDTO(v application.OwnerView) OwnerDTO {
	return OwnerDTO{ID: v.ID, Name: v.Name, Phone: v.Phone, Email: v.Email, Notes: v.Notes, Active: v.Active, OwnershipBPS: v.OwnershipBPS, ProfitSharingBPS: v.ProfitSharingBPS, CapitalContributedRial: v.CapitalContributedRial, DrawingsRial: v.DrawingsRial, CurrentBalanceRial: v.CurrentBalanceRial, LoanPayableRial: v.LoanPayableRial, LoanReceivableRial: v.LoanReceivableRial, AllocatedProfitLossRial: v.AllocatedProfitLossRial, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func ownerTransactionDTO(v application.OwnerTransactionView) OwnerTransactionDTO {
	return OwnerTransactionDTO{ID: v.ID, TransactionNumber: v.TransactionNumber, OwnerID: v.OwnerID, Type: v.Type, FinancialAccountID: v.FinancialAccountID, CategoryAccountID: v.CategoryAccountID, Description: v.Description, Notes: v.Notes, Status: v.Status, JournalEntryID: v.JournalEntryID, IdempotencyKey: v.IdempotencyKey, OccurredAt: v.OccurredAt, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, AmountRial: v.AmountRial}
}
func allocationDTO(v application.ProfitAllocationView) ProfitAllocationDTO {
	return ProfitAllocationDTO{ID: v.ID, PeriodID: v.PeriodID, OwnerID: v.OwnerID, Position: v.Position, ProfitSharingBPS: v.ProfitSharingBPS, AmountRial: v.AmountRial}
}
func periodDTO(v application.FiscalPeriodView) FiscalPeriodDTO {
	out := FiscalPeriodDTO{ID: v.ID, Name: v.Name, Status: v.Status, Notes: v.Notes, ClosingJournalEntryID: v.ClosingJournalEntryID, IdempotencyKey: v.IdempotencyKey, StartDate: v.StartDate, EndDate: v.EndDate, ClosedAt: v.ClosedAt, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, RevenueRial: v.RevenueRial, COGSRial: v.COGSRial, ExpensesRial: v.ExpensesRial, ProfitLossRial: v.ProfitLossRial}
	for _, x := range v.Allocations {
		out.Allocations = append(out.Allocations, allocationDTO(x))
	}
	for _, x := range v.PreviewAllocations {
		out.PreviewAllocations = append(out.PreviewAllocations, allocationDTO(x))
	}
	return out
}
