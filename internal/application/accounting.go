package application

import (
	"context"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type AccountingRepository interface {
	ListAccounts(context.Context) ([]domain.Account, error)
	GetAccount(context.Context, string) (domain.Account, error)
	ListJournalEntries(context.Context) ([]domain.JournalEntry, error)
	GetJournalEntry(context.Context, string) (domain.JournalEntry, error)
	ListFinancialAccounts(context.Context) ([]domain.FinancialAccount, error)
	ListPayments(context.Context) ([]domain.Payment, error)
	GetPayment(context.Context, string) (domain.Payment, error)
	CreatePayment(context.Context, domain.Payment) (domain.Payment, error)
	ReversePayment(context.Context, string, string) (domain.Payment, error)
}

type AccountingService struct {
	repository AccountingRepository
	now        func() time.Time
}

func NewAccountingService(r AccountingRepository) *AccountingService {
	return &AccountingService{repository: r, now: time.Now}
}

type AccountView struct {
	ID, Code, Name, Type string
	Active, System       bool
	BalanceRial          int64
}
type FinancialAccountView struct {
	ID, Name, Type, LedgerAccountID, Details string
	Active                                   bool
	BalanceRial                              int64
}
type JournalLineView struct {
	ID, AccountID, PartyType, PartyID, Memo string
	Position                                int
	DebitRial, CreditRial                   int64
}
type JournalEntryView struct {
	ID, EntryNumber, Description, SourceType, SourceID, ReversalOfID, PostedAt, CreatedAt string
	Lines                                                                                 []JournalLineView
}
type PaymentAllocationView struct {
	ID, TargetType, TargetID string
	Position                 int
	AmountRial               int64
	Reversed                 bool
}
type PaymentView struct {
	ID, PaymentNumber, Direction, Method, FinancialAccountID, CustomerID, SupplierID, Reference, Notes, Status, JournalEntryID, PostedAt, CreatedAt string
	AmountRial                                                                                                                                      int64
	Allocations                                                                                                                                     []PaymentAllocationView
}
type PaymentInput struct {
	ID, Direction, Method, FinancialAccountID, CustomerID, SupplierID, Reference, Notes, PostedAt, IdempotencyKey string
	AmountRial                                                                                                    int64
	Allocations                                                                                                   []PaymentAllocationInput
}
type PaymentAllocationInput struct {
	TargetType, TargetID string
	AmountRial           int64
}

func (s *AccountingService) Accounts(ctx context.Context) ([]AccountView, error) {
	v, e := s.repository.ListAccounts(ctx)
	if e != nil {
		return nil, e
	}
	out := make([]AccountView, 0, len(v))
	for _, a := range v {
		out = append(out, AccountView{a.ID, a.Code, a.Name, string(a.Type), a.Active, a.System, a.BalanceRial})
	}
	return out, nil
}
func (s *AccountingService) Account(ctx context.Context, id string) (AccountView, error) {
	v, e := s.repository.GetAccount(ctx, id)
	if e != nil {
		return AccountView{}, e
	}
	return AccountView{v.ID, v.Code, v.Name, string(v.Type), v.Active, v.System, v.BalanceRial}, nil
}
func (s *AccountingService) FinancialAccounts(ctx context.Context) ([]FinancialAccountView, error) {
	v, e := s.repository.ListFinancialAccounts(ctx)
	if e != nil {
		return nil, e
	}
	out := make([]FinancialAccountView, 0, len(v))
	for _, a := range v {
		out = append(out, FinancialAccountView{a.ID, a.Name, string(a.Type), a.LedgerAccountID, a.Details, a.Active, a.BalanceRial})
	}
	return out, nil
}
func (s *AccountingService) Journal(ctx context.Context) ([]JournalEntryView, error) {
	v, e := s.repository.ListJournalEntries(ctx)
	if e != nil {
		return nil, e
	}
	out := make([]JournalEntryView, 0, len(v))
	for _, j := range v {
		out = append(out, journalView(j))
	}
	return out, nil
}
func (s *AccountingService) Payments(ctx context.Context) ([]PaymentView, error) {
	v, e := s.repository.ListPayments(ctx)
	if e != nil {
		return nil, e
	}
	out := make([]PaymentView, 0, len(v))
	for _, p := range v {
		out = append(out, paymentView(p))
	}
	return out, nil
}
func (s *AccountingService) Payment(ctx context.Context, id string) (PaymentView, error) {
	v, e := s.repository.GetPayment(ctx, id)
	if e != nil {
		return PaymentView{}, e
	}
	return paymentView(v), nil
}
func (s *AccountingService) CreatePayment(ctx context.Context, in PaymentInput) (PaymentView, error) {
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = mustID("PAY-")
	}
	posted := s.now().UTC()
	if strings.TrimSpace(in.PostedAt) != "" {
		v, e := time.Parse(time.RFC3339, in.PostedAt)
		if e != nil {
			return PaymentView{}, e
		}
		posted = v.UTC()
	}
	p := domain.Payment{ID: id, Direction: domain.PaymentDirection(in.Direction), Method: domain.PaymentMethod(in.Method), FinancialAccountID: strings.TrimSpace(in.FinancialAccountID), CustomerID: strings.TrimSpace(in.CustomerID), SupplierID: strings.TrimSpace(in.SupplierID), Reference: strings.TrimSpace(in.Reference), Notes: strings.TrimSpace(in.Notes), AmountRial: in.AmountRial, PostedAt: posted, Status: domain.PaymentPosted, IdempotencyKey: strings.TrimSpace(in.IdempotencyKey), CreatedAt: posted}
	for i, a := range in.Allocations {
		p.Allocations = append(p.Allocations, domain.PaymentAllocation{ID: mustID("PAL-"), PaymentID: id, Position: i, TargetType: strings.TrimSpace(a.TargetType), TargetID: strings.TrimSpace(a.TargetID), AmountRial: a.AmountRial})
	}
	v, e := s.repository.CreatePayment(ctx, p)
	if e != nil {
		return PaymentView{}, e
	}
	return paymentView(v), nil
}
func (s *AccountingService) ReversePayment(ctx context.Context, id, key string) (PaymentView, error) {
	v, e := s.repository.ReversePayment(ctx, id, key)
	if e != nil {
		return PaymentView{}, e
	}
	return paymentView(v), nil
}
func journalView(j domain.JournalEntry) JournalEntryView {
	v := JournalEntryView{ID: j.ID, EntryNumber: j.EntryNumber, Description: j.Description, SourceType: j.SourceType, SourceID: j.SourceID, ReversalOfID: j.ReversalOfID, PostedAt: j.PostedAt.UTC().Format(time.RFC3339Nano), CreatedAt: j.CreatedAt.UTC().Format(time.RFC3339Nano)}
	for _, l := range j.Lines {
		v.Lines = append(v.Lines, JournalLineView{l.ID, l.AccountID, l.PartyType, l.PartyID, l.Memo, l.Position, l.DebitRial, l.CreditRial})
	}
	return v
}
func paymentView(p domain.Payment) PaymentView {
	v := PaymentView{ID: p.ID, PaymentNumber: p.PaymentNumber, Direction: string(p.Direction), Method: string(p.Method), FinancialAccountID: p.FinancialAccountID, CustomerID: p.CustomerID, SupplierID: p.SupplierID, Reference: p.Reference, Notes: p.Notes, Status: string(p.Status), JournalEntryID: p.JournalEntryID, PostedAt: p.PostedAt.UTC().Format(time.RFC3339Nano), CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339Nano), AmountRial: p.AmountRial}
	for _, a := range p.Allocations {
		v.Allocations = append(v.Allocations, PaymentAllocationView{a.ID, a.TargetType, a.TargetID, a.Position, a.AmountRial, a.Reversed})
	}
	return v
}
