package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type OwnersRepository interface {
	ListOwners(context.Context, bool) ([]domain.Owner, error)
	GetOwner(context.Context, string) (domain.Owner, error)
	CreateOwner(context.Context, domain.Owner) (domain.Owner, error)
	UpdateOwnerShares(context.Context, string, int64, int64, string) (domain.Owner, error)
	ArchiveOwner(context.Context, string) error
	ReactivateOwner(context.Context, string) error
	DeleteOwner(context.Context, string) error
	OwnerBalances(context.Context, string) ([6]int64, error)
	ListOwnerTransactions(context.Context, string) ([]domain.OwnerTransaction, error)
	GetOwnerTransaction(context.Context, string) (domain.OwnerTransaction, error)
	CreateOwnerTransaction(context.Context, domain.OwnerTransaction) (domain.OwnerTransaction, error)
	ReverseOwnerTransaction(context.Context, string, string) (domain.OwnerTransaction, error)
	ListFiscalPeriods(context.Context) ([]domain.FiscalPeriod, error)
	GetFiscalPeriod(context.Context, string) (domain.FiscalPeriod, error)
	CreateFiscalPeriod(context.Context, domain.FiscalPeriod) (domain.FiscalPeriod, error)
	CloseFiscalPeriod(context.Context, string, string) (domain.FiscalPeriod, error)
}

type OwnersService struct {
	repository OwnersRepository
	now        func() time.Time
}

func NewOwnersService(r OwnersRepository) *OwnersService {
	return &OwnersService{repository: r, now: time.Now}
}

type OwnerView struct {
	ID, Name, Phone, Email, Notes  string
	Active                         bool
	OwnershipBPS, ProfitSharingBPS int64
	CapitalContributedRial         int64
	DrawingsRial                   int64
	CurrentBalanceRial             int64
	LoanPayableRial                int64
	LoanReceivableRial             int64
	AllocatedProfitLossRial        int64
	CreatedAt, UpdatedAt           string
}
type OwnerInput struct {
	ID, Name, Phone, Email, Notes  string
	OwnershipBPS, ProfitSharingBPS int64
}
type OwnerTransactionInput struct {
	ID, OwnerID, Type, FinancialAccountID, CategoryAccountID, Description, Notes, OccurredAt, IdempotencyKey string
	AmountRial                                                                                               int64
}
type OwnerTransactionView struct {
	ID, TransactionNumber, OwnerID, Type, FinancialAccountID, CategoryAccountID, Description, Notes, Status, JournalEntryID, IdempotencyKey string
	AmountRial                                                                                                                              int64
	OccurredAt, CreatedAt, UpdatedAt                                                                                                        string
}
type ProfitAllocationView struct {
	ID, PeriodID, OwnerID        string
	Position                     int
	ProfitSharingBPS, AmountRial int64
}
type FiscalPeriodView struct {
	ID, Name, Status, Notes, ClosingJournalEntryID, IdempotencyKey string
	StartDate, EndDate, ClosedAt, CreatedAt, UpdatedAt             string
	RevenueRial, COGSRial, ExpensesRial, ProfitLossRial            int64
	Allocations, PreviewAllocations                                []ProfitAllocationView
}

func (s *OwnersService) Owners(ctx context.Context, activeOnly bool) ([]OwnerView, error) {
	rows, err := s.repository.ListOwners(ctx, activeOnly)
	if err != nil {
		return nil, err
	}
	out := make([]OwnerView, 0, len(rows))
	for _, row := range rows {
		v := ownerView(row)
		balances, err := s.repository.OwnerBalances(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		v.CapitalContributedRial, v.DrawingsRial, v.CurrentBalanceRial = balances[0], balances[1], balances[2]
		v.LoanPayableRial, v.LoanReceivableRial, v.AllocatedProfitLossRial = balances[3], balances[4], balances[5]
		out = append(out, v)
	}
	return out, nil
}
func (s *OwnersService) Owner(ctx context.Context, id string) (OwnerView, error) {
	v, err := s.repository.GetOwner(ctx, id)
	if err != nil {
		return OwnerView{}, err
	}
	balances, err := s.repository.OwnerBalances(ctx, id)
	if err != nil {
		return OwnerView{}, err
	}
	out := ownerView(v)
	out.CapitalContributedRial, out.DrawingsRial, out.CurrentBalanceRial = balances[0], balances[1], balances[2]
	out.LoanPayableRial, out.LoanReceivableRial, out.AllocatedProfitLossRial = balances[3], balances[4], balances[5]
	return out, nil
}
func (s *OwnersService) CreateOwner(ctx context.Context, in OwnerInput) (OwnerView, error) {
	now := s.now().UTC()
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = mustID("OWN-")
	}
	v, err := s.repository.CreateOwner(ctx, domain.Owner{ID: id, Name: strings.TrimSpace(in.Name), Phone: strings.TrimSpace(in.Phone), Email: strings.TrimSpace(in.Email), Notes: strings.TrimSpace(in.Notes), OwnershipBPS: in.OwnershipBPS, ProfitSharingBPS: in.ProfitSharingBPS, Active: true, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return OwnerView{}, err
	}
	return s.Owner(ctx, v.ID)
}
func (s *OwnersService) UpdateShares(ctx context.Context, id string, ownership, profit int64, reason string) (OwnerView, error) {
	_, err := s.repository.UpdateOwnerShares(ctx, id, ownership, profit, strings.TrimSpace(reason))
	if err != nil {
		return OwnerView{}, err
	}
	return s.Owner(ctx, id)
}
func (s *OwnersService) Archive(ctx context.Context, id string) error {
	return s.repository.ArchiveOwner(ctx, id)
}
func (s *OwnersService) Reactivate(ctx context.Context, id string) error {
	return s.repository.ReactivateOwner(ctx, id)
}
func (s *OwnersService) Delete(ctx context.Context, id string) error {
	return s.repository.DeleteOwner(ctx, id)
}

func (s *OwnersService) Transactions(ctx context.Context, ownerID string) ([]OwnerTransactionView, error) {
	rows, err := s.repository.ListOwnerTransactions(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]OwnerTransactionView, 0, len(rows))
	for _, row := range rows {
		out = append(out, ownerTransactionView(row))
	}
	return out, nil
}
func (s *OwnersService) CreateTransaction(ctx context.Context, in OwnerTransactionInput) (OwnerTransactionView, error) {
	now := s.now().UTC()
	occurred, err := parseDate(in.OccurredAt, now)
	if err != nil {
		return OwnerTransactionView{}, err
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = mustID("OTX-")
	}
	v, err := s.repository.CreateOwnerTransaction(ctx, domain.OwnerTransaction{ID: id, OwnerID: strings.TrimSpace(in.OwnerID), Type: strings.TrimSpace(in.Type), AmountRial: in.AmountRial, OccurredAt: occurred, FinancialAccountID: strings.TrimSpace(in.FinancialAccountID), CategoryAccountID: strings.TrimSpace(in.CategoryAccountID), Description: strings.TrimSpace(in.Description), Notes: strings.TrimSpace(in.Notes), Status: domain.OwnerTxPosted, IdempotencyKey: strings.TrimSpace(in.IdempotencyKey), CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return OwnerTransactionView{}, err
	}
	return ownerTransactionView(v), nil
}
func (s *OwnersService) ReverseTransaction(ctx context.Context, id, key string) (OwnerTransactionView, error) {
	v, err := s.repository.ReverseOwnerTransaction(ctx, id, strings.TrimSpace(key))
	if err != nil {
		return OwnerTransactionView{}, err
	}
	return ownerTransactionView(v), nil
}

func (s *OwnersService) Periods(ctx context.Context) ([]FiscalPeriodView, error) {
	rows, err := s.repository.ListFiscalPeriods(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]FiscalPeriodView, 0, len(rows))
	for _, row := range rows {
		out = append(out, periodView(row))
	}
	return out, nil
}
func (s *OwnersService) Period(ctx context.Context, id string) (FiscalPeriodView, error) {
	v, err := s.repository.GetFiscalPeriod(ctx, id)
	if err != nil {
		return FiscalPeriodView{}, err
	}
	return periodView(v), nil
}
func (s *OwnersService) PreviewPeriod(ctx context.Context, id string) (FiscalPeriodView, error) {
	v, err := s.repository.GetFiscalPeriod(ctx, id)
	if err != nil {
		return FiscalPeriodView{}, err
	}
	owners, err := s.repository.ListOwners(ctx, true)
	if err != nil {
		return FiscalPeriodView{}, err
	}
	sort.Slice(owners, func(i, j int) bool { return owners[i].ID < owners[j].ID })
	alloc, err := domain.AllocateProfitLoss(v.ProfitLossRial, owners)
	if err != nil {
		return FiscalPeriodView{}, err
	}
	v.Allocations = alloc
	out := periodView(v)
	out.PreviewAllocations = out.Allocations
	out.Allocations = nil
	return out, nil
}
func (s *OwnersService) CreatePeriod(ctx context.Context, id, name, startDate, endDate, notes, key string) (FiscalPeriodView, error) {
	now := s.now().UTC()
	if strings.TrimSpace(id) == "" {
		id = mustID("FP-")
	}
	start, err := parseDate(startDate, now)
	if err != nil {
		return FiscalPeriodView{}, err
	}
	end, err := parseDate(endDate, start)
	if err != nil {
		return FiscalPeriodView{}, err
	}
	v, err := s.repository.CreateFiscalPeriod(ctx, domain.FiscalPeriod{ID: strings.TrimSpace(id), Name: strings.TrimSpace(name), StartDate: start, EndDate: end, Notes: strings.TrimSpace(notes), IdempotencyKey: strings.TrimSpace(key), Status: domain.FiscalPeriodOpen, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return FiscalPeriodView{}, err
	}
	return periodView(v), nil
}
func (s *OwnersService) ClosePeriod(ctx context.Context, id, key string) (FiscalPeriodView, error) {
	v, err := s.repository.CloseFiscalPeriod(ctx, id, strings.TrimSpace(key))
	if err != nil {
		return FiscalPeriodView{}, err
	}
	return periodView(v), nil
}

func ownerView(v domain.Owner) OwnerView {
	return OwnerView{ID: v.ID, Name: v.Name, Phone: v.Phone, Email: v.Email, Notes: v.Notes, Active: v.Active, OwnershipBPS: v.OwnershipBPS, ProfitSharingBPS: v.ProfitSharingBPS, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func ownerTransactionView(v domain.OwnerTransaction) OwnerTransactionView {
	return OwnerTransactionView{ID: v.ID, TransactionNumber: v.TransactionNumber, OwnerID: v.OwnerID, Type: v.Type, FinancialAccountID: v.FinancialAccountID, CategoryAccountID: v.CategoryAccountID, Description: v.Description, Notes: v.Notes, Status: v.Status, JournalEntryID: v.JournalEntryID, IdempotencyKey: v.IdempotencyKey, AmountRial: v.AmountRial, OccurredAt: v.OccurredAt.UTC().Format(time.RFC3339Nano), CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func periodView(v domain.FiscalPeriod) FiscalPeriodView {
	out := FiscalPeriodView{ID: v.ID, Name: v.Name, Status: v.Status, Notes: v.Notes, ClosingJournalEntryID: v.ClosingJournalEntryID, IdempotencyKey: v.IdempotencyKey, StartDate: v.StartDate.UTC().Format(time.RFC3339Nano), EndDate: v.EndDate.UTC().Format(time.RFC3339Nano), CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano), RevenueRial: v.RevenueRial, COGSRial: v.COGSRial, ExpensesRial: v.ExpensesRial, ProfitLossRial: v.ProfitLossRial}
	if v.ClosedAt != nil {
		out.ClosedAt = v.ClosedAt.UTC().Format(time.RFC3339Nano)
	}
	for _, a := range v.Allocations {
		out.Allocations = append(out.Allocations, ProfitAllocationView{ID: a.ID, PeriodID: a.PeriodID, OwnerID: a.OwnerID, Position: a.Position, ProfitSharingBPS: a.ProfitSharingBPS, AmountRial: a.AmountRial})
	}
	return out
}
