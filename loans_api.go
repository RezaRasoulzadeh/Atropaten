package main

import (
	"Atropaten/internal/application"
	"fmt"
)

type LoanInstallmentDTO struct {
	ID                string `json:"id"`
	Position          int    `json:"position"`
	DueDate           string `json:"dueDate"`
	PrincipalRial     int64  `json:"principalRial"`
	InterestFeeRial   int64  `json:"interestFeeRial"`
	TotalDueRial      int64  `json:"totalDueRial"`
	PaidPrincipalRial int64  `json:"paidPrincipalRial"`
	PaidInterestRial  int64  `json:"paidInterestRial"`
	PaidRial          int64  `json:"paidRial"`
	RemainingRial     int64  `json:"remainingRial"`
	OverdueRial       int64  `json:"overdueRial"`
	Status            string `json:"status"`
}
type LoanDTO struct {
	ID                     string               `json:"id"`
	LoanNumber             string               `json:"loanNumber"`
	Direction              string               `json:"direction"`
	CounterpartyName       string               `json:"counterpartyName"`
	CustomerID             string               `json:"customerId"`
	SupplierID             string               `json:"supplierId"`
	StartDate              string               `json:"startDate"`
	EndDate                string               `json:"endDate"`
	Status                 string               `json:"status"`
	Notes                  string               `json:"notes"`
	FinancialAccountID     string               `json:"financialAccountId"`
	JournalEntryID         string               `json:"journalEntryId"`
	IdempotencyKey         string               `json:"idempotencyKey"`
	CreatedAt              string               `json:"createdAt"`
	UpdatedAt              string               `json:"updatedAt"`
	PrincipalRial          int64                `json:"principalRial"`
	InterestFeeRial        int64                `json:"interestFeeRial"`
	PaidPrincipalRial      int64                `json:"paidPrincipalRial"`
	PaidInterestRial       int64                `json:"paidInterestRial"`
	RemainingPrincipalRial int64                `json:"remainingPrincipalRial"`
	RemainingInterestRial  int64                `json:"remainingInterestRial"`
	OverdueRial            int64                `json:"overdueRial"`
	Installments           []LoanInstallmentDTO `json:"installments"`
}
type LoanInstallmentInputDTO struct {
	ID              string `json:"id"`
	DueDate         string `json:"dueDate"`
	PrincipalRial   int64  `json:"principalRial"`
	InterestFeeRial int64  `json:"interestFeeRial"`
}
type LoanInputDTO struct {
	ID                 string                    `json:"id"`
	Direction          string                    `json:"direction"`
	CounterpartyName   string                    `json:"counterpartyName"`
	CustomerID         string                    `json:"customerId"`
	SupplierID         string                    `json:"supplierId"`
	StartDate          string                    `json:"startDate"`
	EndDate            string                    `json:"endDate"`
	Notes              string                    `json:"notes"`
	FinancialAccountID string                    `json:"financialAccountId"`
	IdempotencyKey     string                    `json:"idempotencyKey"`
	PrincipalRial      int64                     `json:"principalRial"`
	InterestFeeRial    int64                     `json:"interestFeeRial"`
	InstallmentCount   int                       `json:"installmentCount"`
	Installments       []LoanInstallmentInputDTO `json:"installments"`
}
type LoanPaymentAllocationDTO struct {
	ID            string `json:"id"`
	PaymentID     string `json:"paymentId"`
	InstallmentID string `json:"installmentId"`
	Position      int    `json:"position"`
	PrincipalRial int64  `json:"principalRial"`
	InterestRial  int64  `json:"interestRial"`
}
type LoanPaymentDTO struct {
	ID                 string                     `json:"id"`
	PaymentNumber      string                     `json:"paymentNumber"`
	LoanID             string                     `json:"loanId"`
	FinancialAccountID string                     `json:"financialAccountId"`
	PaidAt             string                     `json:"paidAt"`
	Notes              string                     `json:"notes"`
	Status             string                     `json:"status"`
	JournalEntryID     string                     `json:"journalEntryId"`
	IdempotencyKey     string                     `json:"idempotencyKey"`
	AmountRial         int64                      `json:"amountRial"`
	PrincipalRial      int64                      `json:"principalRial"`
	InterestRial       int64                      `json:"interestRial"`
	Allocations        []LoanPaymentAllocationDTO `json:"allocations"`
}
type LoanPaymentAllocationInputDTO struct {
	InstallmentID string `json:"installmentId"`
	PrincipalRial int64  `json:"principalRial"`
	InterestRial  int64  `json:"interestRial"`
}
type LoanPaymentInputDTO struct {
	ID                 string                          `json:"id"`
	LoanID             string                          `json:"loanId"`
	FinancialAccountID string                          `json:"financialAccountId"`
	PaidAt             string                          `json:"paidAt"`
	Notes              string                          `json:"notes"`
	IdempotencyKey     string                          `json:"idempotencyKey"`
	AmountRial         int64                           `json:"amountRial"`
	PrincipalRial      int64                           `json:"principalRial"`
	InterestRial       int64                           `json:"interestRial"`
	Allocations        []LoanPaymentAllocationInputDTO `json:"allocations"`
}

func (a *App) loansService() (*application.LoansService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.loans == nil {
		return nil, fmt.Errorf("loans service is not initialized")
	}
	return a.loans, nil
}
func (a *App) ListLoans(direction, status string) ([]LoanDTO, error) {
	s, e := a.loansService()
	if e != nil {
		return nil, e
	}
	v, e := s.List(a.materialContext(), direction, status)
	if e != nil {
		return nil, e
	}
	o := make([]LoanDTO, 0, len(v))
	for _, x := range v {
		o = append(o, loanDTO(x))
	}
	return o, nil
}
func (a *App) GetLoan(id string) (LoanDTO, error) {
	s, e := a.loansService()
	if e != nil {
		return LoanDTO{}, e
	}
	v, e := s.Get(a.materialContext(), id)
	return loanDTO(v), e
}
func (a *App) CreateLoan(in LoanInputDTO) (LoanDTO, error) {
	s, e := a.loansService()
	if e != nil {
		return LoanDTO{}, e
	}
	x := application.LoanInput{ID: in.ID, Direction: in.Direction, CounterpartyName: in.CounterpartyName, CustomerID: in.CustomerID, SupplierID: in.SupplierID, StartDate: in.StartDate, EndDate: in.EndDate, Notes: in.Notes, FinancialAccountID: in.FinancialAccountID, IdempotencyKey: in.IdempotencyKey, PrincipalRial: in.PrincipalRial, InterestFeeRial: in.InterestFeeRial, InstallmentCount: in.InstallmentCount}
	for _, i := range in.Installments {
		x.Installments = append(x.Installments, application.LoanInstallmentInput{ID: i.ID, DueDate: i.DueDate, PrincipalRial: i.PrincipalRial, InterestFeeRial: i.InterestFeeRial})
	}
	v, e := s.Create(a.materialContext(), x)
	return loanDTO(v), e
}
func (a *App) ListLoanPayments(id string) ([]LoanPaymentDTO, error) {
	s, e := a.loansService()
	if e != nil {
		return nil, e
	}
	v, e := s.Payments(a.materialContext(), id)
	if e != nil {
		return nil, e
	}
	o := make([]LoanPaymentDTO, 0, len(v))
	for _, x := range v {
		o = append(o, loanPaymentDTO(x))
	}
	return o, nil
}
func (a *App) CreateLoanPayment(in LoanPaymentInputDTO) (LoanPaymentDTO, error) {
	s, e := a.loansService()
	if e != nil {
		return LoanPaymentDTO{}, e
	}
	x := application.LoanPaymentInput{ID: in.ID, LoanID: in.LoanID, FinancialAccountID: in.FinancialAccountID, PaidAt: in.PaidAt, Notes: in.Notes, IdempotencyKey: in.IdempotencyKey, AmountRial: in.AmountRial, PrincipalRial: in.PrincipalRial, InterestRial: in.InterestRial}
	for _, i := range in.Allocations {
		x.Allocations = append(x.Allocations, application.LoanPaymentAllocationInput{InstallmentID: i.InstallmentID, PrincipalRial: i.PrincipalRial, InterestRial: i.InterestRial})
	}
	v, e := s.Payment(a.materialContext(), x)
	return loanPaymentDTO(v), e
}
func (a *App) ReverseLoanPayment(id, key string) (LoanPaymentDTO, error) {
	s, e := a.loansService()
	if e != nil {
		return LoanPaymentDTO{}, e
	}
	v, e := s.ReversePayment(a.materialContext(), id, key)
	return loanPaymentDTO(v), e
}
func loanDTO(v application.LoanView) LoanDTO {
	o := LoanDTO{ID: v.ID, LoanNumber: v.LoanNumber, Direction: v.Direction, CounterpartyName: v.CounterpartyName, CustomerID: v.CustomerID, SupplierID: v.SupplierID, StartDate: v.StartDate, EndDate: v.EndDate, Status: v.Status, Notes: v.Notes, FinancialAccountID: v.FinancialAccountID, JournalEntryID: v.JournalEntryID, IdempotencyKey: v.IdempotencyKey, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, PrincipalRial: v.PrincipalRial, InterestFeeRial: v.InterestFeeRial, PaidPrincipalRial: v.PaidPrincipalRial, PaidInterestRial: v.PaidInterestRial, RemainingPrincipalRial: v.RemainingPrincipalRial, RemainingInterestRial: v.RemainingInterestRial, OverdueRial: v.OverdueRial}
	for _, i := range v.Installments {
		o.Installments = append(o.Installments, LoanInstallmentDTO{ID: i.ID, Position: i.Position, DueDate: i.DueDate, PrincipalRial: i.PrincipalRial, InterestFeeRial: i.InterestFeeRial, TotalDueRial: i.TotalDueRial, PaidPrincipalRial: i.PaidPrincipalRial, PaidInterestRial: i.PaidInterestRial, PaidRial: i.PaidRial, RemainingRial: i.RemainingRial, OverdueRial: i.OverdueRial, Status: i.Status})
	}
	return o
}
func loanPaymentDTO(v application.LoanPaymentView) LoanPaymentDTO {
	o := LoanPaymentDTO{ID: v.ID, PaymentNumber: v.PaymentNumber, LoanID: v.LoanID, FinancialAccountID: v.FinancialAccountID, PaidAt: v.PaidAt, Notes: v.Notes, Status: v.Status, JournalEntryID: v.JournalEntryID, IdempotencyKey: v.IdempotencyKey, AmountRial: v.AmountRial, PrincipalRial: v.PrincipalRial, InterestRial: v.InterestRial}
	for _, i := range v.Allocations {
		o.Allocations = append(o.Allocations, LoanPaymentAllocationDTO{ID: i.ID, PaymentID: i.PaymentID, InstallmentID: i.InstallmentID, Position: i.Position, PrincipalRial: i.PrincipalRial, InterestRial: i.InterestRial})
	}
	return o
}
