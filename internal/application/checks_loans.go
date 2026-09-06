package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type CheckRepository interface {
	ListChecks(context.Context, string, string) ([]domain.Check, error)
	GetCheck(context.Context, string) (domain.Check, error)
	CreateCheck(context.Context, domain.Check) (domain.Check, error)
	DeleteDraftCheck(context.Context, string) error
	ChangeCheckStatus(context.Context, string, string, string, string) (domain.Check, error)
	ListCheckEvents(context.Context, string) ([]domain.CheckEvent, error)
}
type CheckView struct {
	ID, CheckNumber, Bank, Branch, AccountDescriptor, PayerPayee, CustomerID, SupplierID                         string
	Direction, IssueDate, DueDate, SourceType, SourceID, FinancialAccountID, Notes, Status, CreatedAt, UpdatedAt string
	AmountRial                                                                                                   int64
}
type CheckEventView struct{ ID, CheckID, FromStatus, ToStatus, Note, JournalEntryID, OccurredAt string }
type ChecksService struct {
	repository CheckRepository
	now        func() time.Time
}

func NewChecksService(r CheckRepository) *ChecksService {
	return &ChecksService{repository: r, now: time.Now}
}
func (s *ChecksService) List(ctx context.Context, direction, status string) ([]CheckView, error) {
	rows, e := s.repository.ListChecks(ctx, direction, status)
	if e != nil {
		return nil, e
	}
	out := make([]CheckView, 0, len(rows))
	for _, v := range rows {
		out = append(out, checkView(v))
	}
	return out, nil
}
func (s *ChecksService) Get(ctx context.Context, id string) (CheckView, error) {
	v, e := s.repository.GetCheck(ctx, id)
	if e != nil {
		return CheckView{}, e
	}
	return checkView(v), nil
}
func (s *ChecksService) Create(ctx context.Context, in CheckInput) (CheckView, error) {
	now := s.now().UTC()
	issue, e := parseDate(in.IssueDate, now)
	if e != nil {
		return CheckView{}, e
	}
	due, e := parseDate(in.DueDate, issue)
	if e != nil {
		return CheckView{}, e
	}
	direction := domain.CheckDirection(strings.TrimSpace(in.Direction))
	status := strings.TrimSpace(in.Status)
	if status == "" {
		status = domain.CheckDraft
	}
	v := domain.Check{ID: in.ID, Direction: direction, CheckNumber: in.CheckNumber, Bank: strings.TrimSpace(in.Bank), Branch: strings.TrimSpace(in.Branch), AccountDescriptor: strings.TrimSpace(in.AccountDescriptor), AmountRial: in.AmountRial, IssueDate: issue, DueDate: due, PayerPayee: strings.TrimSpace(in.PayerPayee), CustomerID: strings.TrimSpace(in.CustomerID), SupplierID: strings.TrimSpace(in.SupplierID), SourceType: strings.TrimSpace(in.SourceType), SourceID: strings.TrimSpace(in.SourceID), FinancialAccountID: strings.TrimSpace(in.FinancialAccountID), Notes: strings.TrimSpace(in.Notes), Status: status, CreatedAt: now, UpdatedAt: now}
	if v.ID == "" {
		v.ID = mustID("CHK-")
	}
	if e = v.Validate(); e != nil {
		return CheckView{}, e
	}
	v, e = s.repository.CreateCheck(ctx, v)
	if e != nil {
		return CheckView{}, e
	}
	return checkView(v), nil
}
func (s *ChecksService) Transition(ctx context.Context, id, to, note, key string) (CheckView, error) {
	v, e := s.repository.ChangeCheckStatus(ctx, id, strings.TrimSpace(to), strings.TrimSpace(note), strings.TrimSpace(key))
	if e != nil {
		return CheckView{}, e
	}
	return checkView(v), nil
}
func (s *ChecksService) DeleteDraft(ctx context.Context, id string) error {
	return s.repository.DeleteDraftCheck(ctx, id)
}
func (s *ChecksService) History(ctx context.Context, id string) ([]CheckEventView, error) {
	rows, e := s.repository.ListCheckEvents(ctx, id)
	if e != nil {
		return nil, e
	}
	out := make([]CheckEventView, 0, len(rows))
	for _, v := range rows {
		out = append(out, CheckEventView{v.ID, v.CheckID, v.FromStatus, v.ToStatus, v.Note, v.JournalEntryID, v.OccurredAt.UTC().Format(time.RFC3339Nano)})
	}
	return out, nil
}

type CheckInput struct {
	ID, CheckNumber, Direction, Bank, Branch, AccountDescriptor, PayerPayee, CustomerID, SupplierID, SourceType, SourceID, FinancialAccountID, Notes, IssueDate, DueDate, Status string
	AmountRial                                                                                                                                                                   int64
}

func checkView(v domain.Check) CheckView {
	return CheckView{ID: v.ID, CheckNumber: v.CheckNumber, Direction: string(v.Direction), Bank: v.Bank, Branch: v.Branch, AccountDescriptor: v.AccountDescriptor, PayerPayee: v.PayerPayee, CustomerID: v.CustomerID, SupplierID: v.SupplierID, AmountRial: v.AmountRial, IssueDate: v.IssueDate.UTC().Format(time.RFC3339Nano), DueDate: v.DueDate.UTC().Format(time.RFC3339Nano), SourceType: v.SourceType, SourceID: v.SourceID, FinancialAccountID: v.FinancialAccountID, Notes: v.Notes, Status: v.Status, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}

type LoanRepository interface {
	ListLoans(context.Context, string, string) ([]domain.Loan, error)
	GetLoan(context.Context, string) (domain.Loan, error)
	CreateLoan(context.Context, domain.Loan) (domain.Loan, error)
	ListLoanPayments(context.Context, string) ([]domain.LoanPayment, error)
	CreateLoanPayment(context.Context, domain.LoanPayment) (domain.LoanPayment, error)
	GetLoanPayment(context.Context, string) (domain.LoanPayment, error)
	ReverseLoanPayment(context.Context, string, string) (domain.LoanPayment, error)
}
type LoanInstallmentInput struct {
	ID                             string
	DueDate                        string
	PrincipalRial, InterestFeeRial int64
}
type LoanInput struct {
	ID, Direction, CounterpartyName, CustomerID, SupplierID, StartDate, EndDate, Notes, FinancialAccountID, IdempotencyKey string
	PrincipalRial, InterestFeeRial                                                                                         int64
	Installments                                                                                                           []LoanInstallmentInput
	InstallmentCount                                                                                                       int
}
type LoanView struct {
	ID, LoanNumber, Direction, CounterpartyName, CustomerID, SupplierID, StartDate, EndDate, Status, Notes, FinancialAccountID, JournalEntryID, IdempotencyKey, CreatedAt, UpdatedAt string
	PrincipalRial, InterestFeeRial, PaidPrincipalRial, PaidInterestRial, RemainingPrincipalRial, RemainingInterestRial, OverdueRial                                                  int64
	Installments                                                                                                                                                                     []LoanInstallmentView
}
type LoanInstallmentView struct {
	ID                                                                                                                      string
	Position                                                                                                                int
	DueDate                                                                                                                 string
	PrincipalRial, InterestFeeRial, TotalDueRial, PaidPrincipalRial, PaidInterestRial, PaidRial, RemainingRial, OverdueRial int64
	Status                                                                                                                  string
}
type LoanPaymentInput struct {
	ID, LoanID, FinancialAccountID, PaidAt, Notes, IdempotencyKey string
	AmountRial, PrincipalRial, InterestRial                       int64
	Allocations                                                   []LoanPaymentAllocationInput
}
type LoanPaymentAllocationInput struct {
	InstallmentID               string
	PrincipalRial, InterestRial int64
}
type LoanPaymentView struct {
	ID, PaymentNumber, LoanID, FinancialAccountID, PaidAt, Notes, Status, JournalEntryID, IdempotencyKey string
	AmountRial, PrincipalRial, InterestRial                                                              int64
	Allocations                                                                                          []LoanPaymentAllocationView
}
type LoanPaymentAllocationView struct {
	ID, PaymentID, InstallmentID string
	Position                     int
	PrincipalRial, InterestRial  int64
}
type LoansService struct {
	repository LoanRepository
	now        func() time.Time
}

func NewLoansService(r LoanRepository) *LoansService {
	return &LoansService{repository: r, now: time.Now}
}
func (s *LoansService) List(ctx context.Context, dir, status string) ([]LoanView, error) {
	rows, e := s.repository.ListLoans(ctx, dir, status)
	if e != nil {
		return nil, e
	}
	out := make([]LoanView, 0, len(rows))
	for _, v := range rows {
		out = append(out, loanView(v))
	}
	return out, nil
}
func (s *LoansService) Get(ctx context.Context, id string) (LoanView, error) {
	v, e := s.repository.GetLoan(ctx, id)
	if e != nil {
		return LoanView{}, e
	}
	return loanView(v), nil
}
func (s *LoansService) Create(ctx context.Context, in LoanInput) (LoanView, error) {
	now := s.now().UTC()
	start, e := parseDate(in.StartDate, now)
	if e != nil {
		return LoanView{}, e
	}
	var end *time.Time
	if strings.TrimSpace(in.EndDate) != "" {
		v, x := parseDate(in.EndDate, start)
		if x != nil {
			return LoanView{}, x
		}
		end = &v
	}
	count := in.InstallmentCount
	if count <= 0 {
		count = len(in.Installments)
	}
	if count <= 0 {
		count = 1
	}
	installments := make([]domain.LoanInstallment, 0, count)
	if len(in.Installments) > 0 {
		for n, x := range in.Installments {
			due, xerr := parseDate(x.DueDate, start)
			if xerr != nil {
				return LoanView{}, xerr
			}
			installments = append(installments, domain.LoanInstallment{ID: x.ID, LoanID: in.ID, Position: n, DueDate: due, PrincipalRial: x.PrincipalRial, InterestFeeRial: x.InterestFeeRial, TotalDueRial: x.PrincipalRial + x.InterestFeeRial})
		}
	} else {
		for n := 0; n < count; n++ {
			p := in.PrincipalRial / int64(count)
			if int64(n) < in.PrincipalRial%int64(count) {
				p++
			}
			interest := in.InterestFeeRial / int64(count)
			if int64(n) < in.InterestFeeRial%int64(count) {
				interest++
			}
			installments = append(installments, domain.LoanInstallment{ID: mustID("INST-"), LoanID: in.ID, Position: n, DueDate: start.AddDate(0, n+1, 0), PrincipalRial: p, InterestFeeRial: interest, TotalDueRial: p + interest})
		}
	}
	id := in.ID
	if id == "" {
		id = mustID("LOAN-")
	}
	for i := range installments {
		installments[i].LoanID = id
		if installments[i].ID == "" {
			installments[i].ID = mustID("INST-")
		}
	}
	l := domain.Loan{ID: id, Direction: strings.TrimSpace(in.Direction), CounterpartyName: strings.TrimSpace(in.CounterpartyName), CustomerID: strings.TrimSpace(in.CustomerID), SupplierID: strings.TrimSpace(in.SupplierID), PrincipalRial: in.PrincipalRial, InterestFeeRial: in.InterestFeeRial, StartDate: start, EndDate: end, Status: domain.LoanActive, Notes: strings.TrimSpace(in.Notes), FinancialAccountID: strings.TrimSpace(in.FinancialAccountID), IdempotencyKey: in.IdempotencyKey, CreatedAt: now, UpdatedAt: now, Installments: installments}
	if l.IdempotencyKey == "" {
		l.IdempotencyKey = l.ID
	}
	if e = l.Validate(); e != nil {
		return LoanView{}, e
	}
	v, e := s.repository.CreateLoan(ctx, l)
	if e != nil {
		return LoanView{}, e
	}
	return loanView(v), nil
}
func (s *LoansService) Payment(ctx context.Context, in LoanPaymentInput) (LoanPaymentView, error) {
	at, e := parseDate(in.PaidAt, s.now().UTC())
	if e != nil {
		return LoanPaymentView{}, e
	}
	id := in.ID
	if id == "" {
		id = mustID("LPAY-")
	}
	p := domain.LoanPayment{ID: id, LoanID: in.LoanID, FinancialAccountID: in.FinancialAccountID, AmountRial: in.AmountRial, PrincipalRial: in.PrincipalRial, InterestRial: in.InterestRial, PaidAt: at, Notes: in.Notes, Status: string(domain.PaymentPosted), IdempotencyKey: in.IdempotencyKey, CreatedAt: at}
	if p.IdempotencyKey == "" {
		p.IdempotencyKey = p.ID
	}
	for n, a := range in.Allocations {
		p.Allocations = append(p.Allocations, domain.LoanPaymentAllocation{ID: mustID("LPA-"), PaymentID: id, InstallmentID: a.InstallmentID, Position: n, PrincipalRial: a.PrincipalRial, InterestRial: a.InterestRial})
	}
	v, e := s.repository.CreateLoanPayment(ctx, p)
	if e != nil {
		return LoanPaymentView{}, e
	}
	return loanPaymentView(v), nil
}
func (s *LoansService) Payments(ctx context.Context, id string) ([]LoanPaymentView, error) {
	rows, e := s.repository.ListLoanPayments(ctx, id)
	if e != nil {
		return nil, e
	}
	out := make([]LoanPaymentView, 0, len(rows))
	for _, v := range rows {
		out = append(out, loanPaymentView(v))
	}
	return out, nil
}
func (s *LoansService) ReversePayment(ctx context.Context, id, key string) (LoanPaymentView, error) {
	v, e := s.repository.ReverseLoanPayment(ctx, id, key)
	if e != nil {
		return LoanPaymentView{}, e
	}
	return loanPaymentView(v), nil
}
func loanView(v domain.Loan) LoanView {
	out := LoanView{ID: v.ID, LoanNumber: v.LoanNumber, Direction: v.Direction, CounterpartyName: v.CounterpartyName, CustomerID: v.CustomerID, SupplierID: v.SupplierID, PrincipalRial: v.PrincipalRial, InterestFeeRial: v.InterestFeeRial, PaidPrincipalRial: v.PaidPrincipalRial, PaidInterestRial: v.PaidInterestRial, RemainingPrincipalRial: v.RemainingPrincipalRial, RemainingInterestRial: v.RemainingInterestRial, OverdueRial: v.OverdueRial, StartDate: v.StartDate.UTC().Format(time.RFC3339Nano), Status: v.Status, Notes: v.Notes, FinancialAccountID: v.FinancialAccountID, JournalEntryID: v.JournalEntryID, IdempotencyKey: v.IdempotencyKey, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
	if v.EndDate != nil {
		out.EndDate = v.EndDate.UTC().Format(time.RFC3339Nano)
	}
	for _, i := range v.Installments {
		out.Installments = append(out.Installments, LoanInstallmentView{ID: i.ID, Position: i.Position, DueDate: i.DueDate.UTC().Format(time.RFC3339Nano), PrincipalRial: i.PrincipalRial, InterestFeeRial: i.InterestFeeRial, TotalDueRial: i.TotalDueRial, PaidPrincipalRial: i.PaidPrincipalRial, PaidInterestRial: i.PaidInterestRial, PaidRial: i.PaidRial, RemainingRial: i.RemainingRial, OverdueRial: i.OverdueRial, Status: i.Status})
	}
	return out
}
func loanPaymentView(v domain.LoanPayment) LoanPaymentView {
	out := LoanPaymentView{ID: v.ID, PaymentNumber: v.PaymentNumber, LoanID: v.LoanID, FinancialAccountID: v.FinancialAccountID, AmountRial: v.AmountRial, PrincipalRial: v.PrincipalRial, InterestRial: v.InterestRial, PaidAt: v.PaidAt.UTC().Format(time.RFC3339Nano), Notes: v.Notes, Status: v.Status, JournalEntryID: v.JournalEntryID, IdempotencyKey: v.IdempotencyKey}
	for _, a := range v.Allocations {
		out.Allocations = append(out.Allocations, LoanPaymentAllocationView{ID: a.ID, PaymentID: a.PaymentID, InstallmentID: a.InstallmentID, Position: a.Position, PrincipalRial: a.PrincipalRial, InterestRial: a.InterestRial})
	}
	return out
}
func parseDate(value string, fallback time.Time) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return fallback.UTC(), nil
	}
	v, e := time.Parse(time.RFC3339, value)
	if e != nil {
		v, e = time.Parse("2006-01-02", value)
	}
	if e != nil {
		return time.Time{}, fmt.Errorf("invalid date: %w", e)
	}
	return v.UTC(), nil
}
