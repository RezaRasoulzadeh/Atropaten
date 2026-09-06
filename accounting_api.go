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
type ExpenseDTO struct {
	ID                 string `json:"id"`
	ExpenseNumber      string `json:"expenseNumber"`
	ExpenseDate        string `json:"expenseDate"`
	CategoryAccountID  string `json:"categoryAccountId"`
	Payee              string `json:"payee"`
	SupplierID         string `json:"supplierId"`
	Description        string `json:"description"`
	AmountRial         int64  `json:"amountRial"`
	PaymentMethod      string `json:"paymentMethod"`
	FinancialAccountID string `json:"financialAccountId"`
	Notes              string `json:"notes"`
	Status             string `json:"status"`
	JournalEntryID     string `json:"journalEntryId"`
	IdempotencyKey     string `json:"idempotencyKey"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}
type ExpenseInputDTO struct {
	ID                 string `json:"id"`
	ExpenseDate        string `json:"expenseDate"`
	CategoryAccountID  string `json:"categoryAccountId"`
	Payee              string `json:"payee"`
	SupplierID         string `json:"supplierId"`
	Description        string `json:"description"`
	AmountRial         int64  `json:"amountRial"`
	PaymentMethod      string `json:"paymentMethod"`
	FinancialAccountID string `json:"financialAccountId"`
	Notes              string `json:"notes"`
	IdempotencyKey     string `json:"idempotencyKey"`
}
type TransferDTO struct {
	ID                            string `json:"id"`
	TransferNumber                string `json:"transferNumber"`
	SourceFinancialAccountID      string `json:"sourceFinancialAccountId"`
	DestinationFinancialAccountID string `json:"destinationFinancialAccountId"`
	AmountRial                    int64  `json:"amountRial"`
	TransferDate                  string `json:"transferDate"`
	Reference                     string `json:"reference"`
	Notes                         string `json:"notes"`
	Status                        string `json:"status"`
	JournalEntryID                string `json:"journalEntryId"`
	IdempotencyKey                string `json:"idempotencyKey"`
	CreatedAt                     string `json:"createdAt"`
	UpdatedAt                     string `json:"updatedAt"`
}
type TransferInputDTO struct {
	ID                            string `json:"id"`
	SourceFinancialAccountID      string `json:"sourceFinancialAccountId"`
	DestinationFinancialAccountID string `json:"destinationFinancialAccountId"`
	AmountRial                    int64  `json:"amountRial"`
	TransferDate                  string `json:"transferDate"`
	Reference                     string `json:"reference"`
	Notes                         string `json:"notes"`
	IdempotencyKey                string `json:"idempotencyKey"`
}
type CustomerFinancialDTO struct {
	CustomerID     string `json:"customerId"`
	ReceivableRial int64  `json:"receivableRial"`
	CreditRial     int64  `json:"creditRial"`
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
func (a *App) GetCustomerFinancialSummary(customerID string) (CustomerFinancialDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return CustomerFinancialDTO{}, e
	}
	v, e := s.CustomerFinancial(a.materialContext(), customerID)
	return CustomerFinancialDTO{CustomerID: v.CustomerID, ReceivableRial: v.ReceivableRial, CreditRial: v.CreditRial}, e
}
func (a *App) GetSupplierPayableBalance(supplierID string) (int64, error) {
	s, e := a.accountingService()
	if e != nil {
		return 0, e
	}
	return s.SupplierPayable(a.materialContext(), supplierID)
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
func (a *App) ListExpenses() ([]ExpenseDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return nil, e
	}
	v, e := s.Expenses(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]ExpenseDTO, 0, len(v))
	for _, x := range v {
		out = append(out, expenseDTO(x))
	}
	return out, nil
}
func (a *App) CreateExpense(in ExpenseInputDTO) (ExpenseDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return ExpenseDTO{}, e
	}
	v, e := s.CreateExpense(a.materialContext(), application.ExpenseInput{ID: in.ID, ExpenseDate: in.ExpenseDate, CategoryAccountID: in.CategoryAccountID, Payee: in.Payee, SupplierID: in.SupplierID, Description: in.Description, AmountRial: in.AmountRial, PaymentMethod: in.PaymentMethod, FinancialAccountID: in.FinancialAccountID, Notes: in.Notes, IdempotencyKey: in.IdempotencyKey})
	return expenseDTO(v), e
}
func (a *App) ReverseExpense(id, key string) (ExpenseDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return ExpenseDTO{}, e
	}
	v, e := s.ReverseExpense(a.materialContext(), id, key)
	return expenseDTO(v), e
}
func (a *App) ListTransfers() ([]TransferDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return nil, e
	}
	v, e := s.Transfers(a.materialContext())
	if e != nil {
		return nil, e
	}
	out := make([]TransferDTO, 0, len(v))
	for _, x := range v {
		out = append(out, transferDTO(x))
	}
	return out, nil
}
func (a *App) CreateTransfer(in TransferInputDTO) (TransferDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return TransferDTO{}, e
	}
	v, e := s.CreateTransfer(a.materialContext(), application.TransferInput{ID: in.ID, SourceFinancialAccountID: in.SourceFinancialAccountID, DestinationFinancialAccountID: in.DestinationFinancialAccountID, AmountRial: in.AmountRial, TransferDate: in.TransferDate, Reference: in.Reference, Notes: in.Notes, IdempotencyKey: in.IdempotencyKey})
	return transferDTO(v), e
}
func (a *App) ReverseTransfer(id, key string) (TransferDTO, error) {
	s, e := a.accountingService()
	if e != nil {
		return TransferDTO{}, e
	}
	v, e := s.ReverseTransfer(a.materialContext(), id, key)
	return transferDTO(v), e
}
func expenseDTO(v application.ExpenseView) ExpenseDTO {
	return ExpenseDTO{ID: v.ID, ExpenseNumber: v.ExpenseNumber, ExpenseDate: v.ExpenseDate, CategoryAccountID: v.CategoryAccountID, Payee: v.Payee, SupplierID: v.SupplierID, Description: v.Description, AmountRial: v.AmountRial, PaymentMethod: v.PaymentMethod, FinancialAccountID: v.FinancialAccountID, Notes: v.Notes, Status: v.Status, JournalEntryID: v.JournalEntryID, IdempotencyKey: v.IdempotencyKey, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func transferDTO(v application.TransferView) TransferDTO {
	return TransferDTO{ID: v.ID, TransferNumber: v.TransferNumber, SourceFinancialAccountID: v.SourceFinancialAccountID, DestinationFinancialAccountID: v.DestinationFinancialAccountID, AmountRial: v.AmountRial, TransferDate: v.TransferDate, Reference: v.Reference, Notes: v.Notes, Status: v.Status, JournalEntryID: v.JournalEntryID, IdempotencyKey: v.IdempotencyKey, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
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
