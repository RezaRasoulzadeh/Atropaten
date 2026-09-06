package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var (
	ErrAccountNotFound          = errors.New("account not found")
	ErrAccountInactive          = errors.New("account is inactive")
	ErrJournalEntryNotFound     = errors.New("journal entry not found")
	ErrJournalUnbalanced        = errors.New("journal entry is not balanced")
	ErrJournalTooFewLines       = errors.New("journal entry needs at least two lines")
	ErrJournalImmutable         = errors.New("posted journal history is immutable")
	ErrFinancialAccountNotFound = errors.New("financial account not found")
	ErrPaymentNotFound          = errors.New("payment not found")
	ErrPaymentReversed          = errors.New("payment is already reversed")
	ErrPaymentInvalidParty      = errors.New("payment party is invalid")
	ErrAllocationExceeded       = errors.New("payment allocation exceeds the payment amount")
	ErrAllocationTarget         = errors.New("payment allocation target not found")
)

type AccountType string

const (
	AccountAsset     AccountType = "asset"
	AccountLiability AccountType = "liability"
	AccountEquity    AccountType = "equity"
	AccountRevenue   AccountType = "revenue"
	AccountExpense   AccountType = "expense"
)

type Account struct {
	ID, Code, Name, ParentID string
	Type                     AccountType
	Active, System           bool
	BalanceRial              int64
	CreatedAt, UpdatedAt     time.Time
}

type JournalEntry struct {
	ID, EntryNumber, Description, SourceType, SourceID, IdempotencyKey, ReversalOfID string
	PostedAt, CreatedAt                                                              time.Time
	Lines                                                                            []JournalLine
}

type JournalLine struct {
	ID, JournalEntryID, AccountID, PartyType, PartyID, Memo string
	Position                                                int
	DebitRial, CreditRial                                   int64
}

func (j JournalEntry) Validate() error {
	if strings.TrimSpace(j.ID) == "" || strings.TrimSpace(j.Description) == "" || j.PostedAt.IsZero() {
		return fmt.Errorf("journal entry id, description, and posted time are required")
	}
	if len(j.Lines) < 2 {
		return ErrJournalTooFewLines
	}
	var debit, credit big.Int
	for i, line := range j.Lines {
		if line.AccountID == "" || line.Position != i {
			return fmt.Errorf("journal line %d has an invalid account or position", i)
		}
		if line.DebitRial < 0 || line.CreditRial < 0 || (line.DebitRial == 0) == (line.CreditRial == 0) {
			return fmt.Errorf("journal line %d must contain exactly one positive side", i)
		}
		debit.Add(&debit, big.NewInt(line.DebitRial))
		credit.Add(&credit, big.NewInt(line.CreditRial))
	}
	if debit.Cmp(&credit) != 0 {
		return ErrJournalUnbalanced
	}
	if !debit.IsInt64() {
		return fmt.Errorf("journal amount is too large")
	}
	return nil
}

type FinancialAccountType string

const (
	FinancialCash FinancialAccountType = "cash"
	FinancialBank FinancialAccountType = "bank"
)

type FinancialAccount struct {
	ID, Name, Details, LedgerAccountID string
	Type                               FinancialAccountType
	Active                             bool
	BalanceRial                        int64
	CreatedAt, UpdatedAt               time.Time
}

type PaymentDirection string

const (
	PaymentIncoming PaymentDirection = "incoming"
	PaymentOutgoing PaymentDirection = "outgoing"
)

type PaymentMethod string

const (
	PaymentCash         PaymentMethod = "cash"
	PaymentBankTransfer PaymentMethod = "bank_transfer"
	PaymentCard         PaymentMethod = "card"
	PaymentCheck        PaymentMethod = "check"
	PaymentOther        PaymentMethod = "other"
)

type PaymentState string

const (
	PaymentPosted        PaymentState = "posted"
	PaymentReversedState PaymentState = "reversed"
)

type Payment struct {
	ID, PaymentNumber, FinancialAccountID, CustomerID, SupplierID, Reference, Notes, IdempotencyKey, JournalEntryID string
	Direction                                                                                                       PaymentDirection
	Method                                                                                                          PaymentMethod
	AmountRial                                                                                                      int64
	PostedAt                                                                                                        time.Time
	Status                                                                                                          PaymentState
	CreatedAt                                                                                                       time.Time
	Allocations                                                                                                     []PaymentAllocation
}

type PaymentAllocation struct {
	ID, PaymentID, TargetType, TargetID string
	Position                            int
	AmountRial                          int64
	Reversed                            bool
}

func (p Payment) Validate() error {
	if p.ID == "" || p.AmountRial <= 0 || p.PostedAt.IsZero() || p.FinancialAccountID == "" {
		return fmt.Errorf("payment id, positive amount, posted time, and financial account are required")
	}
	if p.Direction != PaymentIncoming && p.Direction != PaymentOutgoing {
		return fmt.Errorf("unsupported payment direction")
	}
	if p.Method != PaymentCash && p.Method != PaymentBankTransfer && p.Method != PaymentCard && p.Method != PaymentCheck && p.Method != PaymentOther {
		return fmt.Errorf("unsupported payment method")
	}
	if p.Status != PaymentPosted && p.Status != PaymentReversedState {
		return fmt.Errorf("unsupported payment status")
	}
	var total big.Int
	for i, a := range p.Allocations {
		if a.Position != i || a.AmountRial <= 0 || a.TargetID == "" || (a.TargetType != "order" && a.TargetType != "purchase" && a.TargetType != "invoice") {
			return fmt.Errorf("invalid payment allocation %d", i)
		}
		total.Add(&total, big.NewInt(a.AmountRial))
	}
	if total.Cmp(big.NewInt(p.AmountRial)) > 0 {
		return ErrAllocationExceeded
	}
	return nil
}

func (j JournalEntry) Reversal(id, key, description string, postedAt time.Time) JournalEntry {
	r := JournalEntry{ID: id, Description: description, SourceType: j.SourceType, SourceID: j.SourceID, IdempotencyKey: key, ReversalOfID: j.ID, PostedAt: postedAt, CreatedAt: postedAt}
	for i, line := range j.Lines {
		r.Lines = append(r.Lines, JournalLine{ID: id + "-L" + fmt.Sprint(i+1), JournalEntryID: id, Position: i, AccountID: line.AccountID, DebitRial: line.CreditRial, CreditRial: line.DebitRial, PartyType: line.PartyType, PartyID: line.PartyID, Memo: "Reversal of " + j.ID})
	}
	return r
}
