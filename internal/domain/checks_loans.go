package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var (
	ErrCheckNotFound       = errors.New("check not found")
	ErrCheckTransition     = errors.New("invalid check status transition")
	ErrCheckProtected      = errors.New("check history cannot be deleted")
	ErrCheckObligationPaid = errors.New("the linked obligation already has an active payment or check")
	ErrLoanNotFound        = errors.New("loan not found")
	ErrLoanProtected       = errors.New("loan history cannot be deleted")
	ErrLoanPaymentNotFound = errors.New("loan payment not found")
	ErrLoanPaymentExceeded = errors.New("loan payment exceeds the remaining principal")
)

type CheckDirection string

const (
	CheckIncoming CheckDirection = "incoming"
	CheckOutgoing CheckDirection = "outgoing"
)

const (
	CheckDraft     = "Draft"
	CheckReceived  = "Received"
	CheckDeposited = "Deposited"
	CheckCleared   = "Cleared"
	CheckReturned  = "Returned"
	CheckCancelled = "Cancelled"
	CheckIssued    = "Issued"
	CheckDelivered = "Delivered"
	CheckRejected  = "Rejected"
)

func ValidCheckStatus(direction CheckDirection, status string) bool {
	if direction == CheckIncoming {
		switch status {
		case CheckDraft, CheckReceived, CheckDeposited, CheckCleared, CheckReturned, CheckCancelled:
			return true
		}
	}
	if direction == CheckOutgoing {
		switch status {
		case CheckDraft, CheckIssued, CheckDelivered, CheckCleared, CheckReturned, CheckRejected, CheckCancelled:
			return true
		}
	}
	return false
}

func ValidCheckTransition(direction CheckDirection, from, to string) bool {
	// A retry is handled by the persisted idempotency key in the store; a new
	// event that leaves the state unchanged is never a valid transition.
	if from == to {
		return false
	}
	if direction == CheckIncoming {
		switch from {
		case CheckDraft:
			return to == CheckReceived || to == CheckCancelled
		case CheckReceived:
			return to == CheckDeposited || to == CheckReturned || to == CheckCancelled
		case CheckDeposited:
			return to == CheckCleared || to == CheckReturned
		case CheckCleared:
			return to == CheckReturned
		case CheckReturned:
			return to == CheckCancelled
		}
	}
	if direction == CheckOutgoing {
		switch from {
		case CheckDraft:
			return to == CheckIssued || to == CheckCancelled
		case CheckIssued:
			return to == CheckDelivered || to == CheckReturned || to == CheckRejected || to == CheckCancelled
		case CheckDelivered:
			return to == CheckCleared || to == CheckReturned || to == CheckRejected
		case CheckCleared:
			return to == CheckReturned || to == CheckRejected
		case CheckReturned, CheckRejected:
			return to == CheckCancelled
		}
	}
	return false
}

type Check struct {
	ID, CheckNumber, Bank, Branch, AccountDescriptor, PayerPayee, CustomerID, SupplierID string
	Direction                                                                            CheckDirection
	AmountRial                                                                           int64
	IssueDate, DueDate                                                                   time.Time
	SourceType, SourceID, FinancialAccountID, Notes, Status                              string
	CreatedAt, UpdatedAt                                                                 time.Time
}

type CheckEvent struct {
	ID, CheckID, FromStatus, ToStatus, Note, JournalEntryID, IdempotencyKey string
	OccurredAt                                                              time.Time
}

func (c Check) Validate() error {
	if c.ID == "" || c.CheckNumber == "" || c.Bank == "" || c.AmountRial <= 0 || c.IssueDate.IsZero() || c.DueDate.IsZero() || c.Status == "" || !ValidCheckStatus(c.Direction, c.Status) || c.CreatedAt.IsZero() || c.UpdatedAt.IsZero() {
		return fmt.Errorf("check identity, bank, positive amount, dates, status, and timestamps are required")
	}
	if c.Direction == CheckIncoming && c.PayerPayee == "" {
		return fmt.Errorf("incoming check payer is required")
	}
	if c.Direction == CheckOutgoing && c.PayerPayee == "" {
		return fmt.Errorf("outgoing check payee is required")
	}
	return nil
}

const (
	LoanPayable    = "payable"
	LoanReceivable = "receivable"
	LoanDraft      = "Draft"
	LoanActive     = "Active"
	LoanClosed     = "Closed"
	LoanCancelled  = "Cancelled"
)

type Loan struct {
	ID, LoanNumber, Direction, CounterpartyName, CustomerID, SupplierID                             string
	PrincipalRial, InterestFeeRial                                                                  int64
	StartDate                                                                                       time.Time
	EndDate                                                                                         *time.Time
	Status, Notes, FinancialAccountID, JournalEntryID, IdempotencyKey                               string
	PaidPrincipalRial, PaidInterestRial, RemainingPrincipalRial, RemainingInterestRial, OverdueRial int64
	CreatedAt, UpdatedAt                                                                            time.Time
	Installments                                                                                    []LoanInstallment
}

type LoanInstallment struct {
	ID, LoanID                                                                string
	Position                                                                  int
	DueDate                                                                   time.Time
	PrincipalRial, InterestFeeRial, TotalDueRial                              int64
	PaidPrincipalRial, PaidInterestRial, PaidRial, RemainingRial, OverdueRial int64
	Status                                                                    string
}

type LoanPayment struct {
	ID, PaymentNumber, LoanID, FinancialAccountID, Notes, Status, JournalEntryID, IdempotencyKey string
	AmountRial, PrincipalRial, InterestRial                                                      int64
	PaidAt, CreatedAt                                                                            time.Time
	Allocations                                                                                  []LoanPaymentAllocation
}

type LoanPaymentAllocation struct {
	ID, PaymentID, InstallmentID string
	Position                     int
	PrincipalRial, InterestRial  int64
}

func (l Loan) Validate() error {
	if l.ID == "" || l.CounterpartyName == "" || (l.Direction != LoanPayable && l.Direction != LoanReceivable) || l.PrincipalRial <= 0 || l.InterestFeeRial < 0 || l.StartDate.IsZero() || l.FinancialAccountID == "" || l.Status != LoanActive || l.CreatedAt.IsZero() || l.UpdatedAt.IsZero() {
		return fmt.Errorf("loan requires direction, counterparty, positive principal, start date, account, and Active status")
	}
	var principal, interest big.Int
	for i, installment := range l.Installments {
		if installment.ID == "" || installment.LoanID != l.ID || installment.Position != i || installment.DueDate.IsZero() || installment.PrincipalRial < 0 || installment.InterestFeeRial < 0 || installment.TotalDueRial != installment.PrincipalRial+installment.InterestFeeRial {
			return fmt.Errorf("invalid loan installment %d", i)
		}
		principal.Add(&principal, big.NewInt(installment.PrincipalRial))
		interest.Add(&interest, big.NewInt(installment.InterestFeeRial))
	}
	if !principal.IsInt64() || principal.Int64() != l.PrincipalRial || !interest.IsInt64() || interest.Int64() != l.InterestFeeRial {
		return fmt.Errorf("installment schedule does not equal loan principal and interest")
	}
	return nil
}

func (p LoanPayment) Validate() error {
	if p.ID == "" || p.LoanID == "" || p.FinancialAccountID == "" || p.AmountRial <= 0 || p.PrincipalRial < 0 || p.InterestRial < 0 || p.AmountRial != p.PrincipalRial+p.InterestRial || p.PaidAt.IsZero() || p.Status != string(PaymentPosted) {
		return fmt.Errorf("loan payment requires balanced positive principal/interest and posted date")
	}
	var principal, interest big.Int
	for i, allocation := range p.Allocations {
		if allocation.ID == "" || allocation.PaymentID != p.ID || allocation.Position != i || allocation.InstallmentID == "" || allocation.PrincipalRial < 0 || allocation.InterestRial < 0 {
			return fmt.Errorf("invalid loan payment allocation %d", i)
		}
		principal.Add(&principal, big.NewInt(allocation.PrincipalRial))
		interest.Add(&interest, big.NewInt(allocation.InterestRial))
	}
	if !principal.IsInt64() || principal.Int64() != p.PrincipalRial || !interest.IsInt64() || interest.Int64() != p.InterestRial {
		return fmt.Errorf("loan payment allocations do not equal payment split")
	}
	return nil
}

func NormalizeCounterparty(v string) string { return strings.TrimSpace(v) }
