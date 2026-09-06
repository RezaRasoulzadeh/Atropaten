package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const PercentageScale int64 = 10_000 // 100.00%, stored as fixed-point basis points.

var (
	ErrOwnerNotFound            = errors.New("owner not found")
	ErrOwnerProtected           = errors.New("owner has protected history")
	ErrOwnerInvalidPercentage   = errors.New("owner percentage must be between 0.00% and 100.00%")
	ErrOwnerShareTotal          = errors.New("active owner percentages must not exceed 100.00%")
	ErrOwnerTransaction         = errors.New("invalid owner transaction")
	ErrOwnerTransactionNotFound = errors.New("owner transaction not found")
	ErrPeriodNotFound           = errors.New("fiscal period not found")
	ErrPeriodOverlap            = errors.New("fiscal period overlaps an existing period")
	ErrPeriodClosed             = errors.New("fiscal period is closed")
	ErrPeriodState              = errors.New("fiscal period is not open")
	ErrPeriodOwnerShares        = errors.New("active owner profit-sharing percentages must total 100.00%")
	ErrPeriodAlreadyClosed      = errors.New("fiscal period has already been closed")
	ErrPeriodUnbalancedLedger   = errors.New("cannot close an unbalanced ledger")
)

type Owner struct {
	ID, Name, Phone, Email, Notes  string
	Active                         bool
	OwnershipBPS, ProfitSharingBPS int64
	CreatedAt, UpdatedAt           time.Time
}

type OwnerShareHistory struct {
	ID, OwnerID, Reason            string
	OwnershipBPS, ProfitSharingBPS int64
	EffectiveAt                    time.Time
}

const (
	OwnerTxCapitalContribution    = "capital_contribution"
	OwnerTxDrawing                = "drawing"
	OwnerTxPersonalExpense        = "owner_paid_expense"
	OwnerTxReimbursement          = "owner_reimbursement"
	OwnerTxLoanToBusiness         = "loan_to_business"
	OwnerTxLoanFromBusiness       = "loan_from_business"
	OwnerTxLoanRepaymentToOwner   = "loan_repayment_to_owner"
	OwnerTxLoanRepaymentFromOwner = "loan_repayment_from_owner"
	OwnerTxPosted                 = "Posted"
	OwnerTxReversed               = "Reversed"
)

type OwnerTransaction struct {
	ID, TransactionNumber, OwnerID, Type, FinancialAccountID, CategoryAccountID string
	AmountRial                                                                  int64
	OccurredAt                                                                  time.Time
	Description, Notes, Status, JournalEntryID, IdempotencyKey                  string
	CreatedAt, UpdatedAt                                                        time.Time
}

type FiscalPeriod struct {
	ID, Name, Status, Notes, ClosingJournalEntryID, IdempotencyKey string
	StartDate, EndDate                                             time.Time
	ClosedAt                                                       *time.Time
	CreatedAt, UpdatedAt                                           time.Time
	RevenueRial, COGSRial, ExpensesRial, ProfitLossRial            int64
	Allocations                                                    []ProfitAllocation
}

const (
	FiscalPeriodOpen    = "Open"
	FiscalPeriodClosing = "Closing"
	FiscalPeriodClosed  = "Closed"
)

type ProfitAllocation struct {
	ID, PeriodID, OwnerID string
	Position              int
	ProfitSharingBPS      int64
	AmountRial            int64
}

func ValidatePercentageBPS(v int64) error {
	if v < 0 || v > PercentageScale {
		return ErrOwnerInvalidPercentage
	}
	return nil
}
func (o Owner) Validate() error {
	if strings.TrimSpace(o.ID) == "" || strings.TrimSpace(o.Name) == "" || o.CreatedAt.IsZero() || o.UpdatedAt.IsZero() {
		return fmt.Errorf("owner id, name, and timestamps are required")
	}
	if err := ValidatePercentageBPS(o.OwnershipBPS); err != nil {
		return err
	}
	return ValidatePercentageBPS(o.ProfitSharingBPS)
}
func (t OwnerTransaction) Validate() error {
	valid := map[string]bool{OwnerTxCapitalContribution: true, OwnerTxDrawing: true, OwnerTxPersonalExpense: true, OwnerTxReimbursement: true, OwnerTxLoanToBusiness: true, OwnerTxLoanFromBusiness: true, OwnerTxLoanRepaymentToOwner: true, OwnerTxLoanRepaymentFromOwner: true}
	if t.ID == "" || t.OwnerID == "" || !valid[t.Type] || t.AmountRial <= 0 || t.OccurredAt.IsZero() || t.Status != OwnerTxPosted || t.IdempotencyKey == "" || t.CreatedAt.IsZero() || t.UpdatedAt.IsZero() {
		return ErrOwnerTransaction
	}
	if t.Type == OwnerTxPersonalExpense && t.CategoryAccountID == "" {
		return fmt.Errorf("owner-paid expense requires an expense category")
	}
	if t.Type != OwnerTxPersonalExpense && t.FinancialAccountID == "" {
		return fmt.Errorf("owner transaction requires a cash or bank account")
	}
	return nil
}
func (p FiscalPeriod) Validate() error {
	if p.ID == "" || strings.TrimSpace(p.Name) == "" || p.StartDate.IsZero() || p.EndDate.IsZero() || p.StartDate.After(p.EndDate) {
		return fmt.Errorf("fiscal period requires a name and valid date range")
	}
	if p.Status != FiscalPeriodOpen && p.Status != FiscalPeriodClosing && p.Status != FiscalPeriodClosed {
		return fmt.Errorf("invalid fiscal period status")
	}
	return nil
}
func AllocateProfitLoss(amount int64, owners []Owner) ([]ProfitAllocation, error) {
	if len(owners) == 0 {
		return nil, ErrPeriodOwnerShares
	}
	var total int64
	for _, o := range owners {
		if err := ValidatePercentageBPS(o.ProfitSharingBPS); err != nil {
			return nil, err
		}
		total += o.ProfitSharingBPS
	}
	if total != PercentageScale {
		return nil, ErrPeriodOwnerShares
	}
	allocations := make([]ProfitAllocation, 0, len(owners))
	absolute := new(big.Int).Abs(big.NewInt(amount))
	denominator := big.NewInt(PercentageScale)
	var assigned int64
	for i, o := range owners {
		product := new(big.Int).Mul(absolute, big.NewInt(o.ProfitSharingBPS))
		q := new(big.Int).Quo(product, denominator)
		value := q.Int64()
		if i == len(owners)-1 {
			value = absolute.Int64() - assigned
		} else {
			assigned += value
		}
		if amount < 0 {
			value = -value
		}
		allocations = append(allocations, ProfitAllocation{OwnerID: o.ID, Position: i, ProfitSharingBPS: o.ProfitSharingBPS, AmountRial: value})
	}
	return allocations, nil
}
