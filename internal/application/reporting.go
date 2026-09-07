package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type ReportingRepository interface {
	Report(context.Context, string, time.Time, time.Time) (domain.Report, error)
	Dashboard(context.Context, time.Time, time.Time) (domain.Dashboard, error)
	PrintDocument(context.Context, string, string, string, string, string) (domain.PrintDocument, error)
	GetShopSettings(context.Context) (domain.ShopSettings, error)
	SaveShopSettings(context.Context, domain.ShopSettings) error
}

type ReportingService struct {
	repository ReportingRepository
	now        func() time.Time
}

func NewReportingService(r ReportingRepository) *ReportingService {
	return &ReportingService{repository: r, now: time.Now}
}

func reportingDates(now time.Time, start, end string) (time.Time, time.Time, error) {
	var err error
	var s, e time.Time
	if start == "" {
		s = now.AddDate(0, 0, -30)
	} else {
		s, err = time.Parse("2006-01-02", start)
		if err != nil {
			return s, e, fmt.Errorf("start date: %w", err)
		}
	}
	if end == "" {
		e = now
	} else {
		e, err = time.Parse("2006-01-02", end)
		if err != nil {
			return s, e, fmt.Errorf("end date: %w", err)
		}
	}
	if e.Before(s) {
		return s, e, fmt.Errorf("end date must not be before start date")
	}
	return s, e, nil
}
func (s *ReportingService) Report(ctx context.Context, kind, start, end string) (domain.Report, error) {
	a, b, e := reportingDates(s.now(), start, end)
	if e != nil {
		return domain.Report{}, e
	}
	return s.repository.Report(ctx, kind, a, b)
}
func (s *ReportingService) Dashboard(ctx context.Context, start, end string) (domain.Dashboard, error) {
	a, b, e := reportingDates(s.now(), start, end)
	if e != nil {
		return domain.Dashboard{}, e
	}
	return s.repository.Dashboard(ctx, a, b)
}
func (s *ReportingService) PrintDocument(ctx context.Context, kind, id, start, end, partyID string) (domain.PrintDocument, error) {
	if start == "" && end == "" {
		a, b, _ := reportingDates(s.now(), "", "")
		start = a.Format("2006-01-02")
		end = b.Format("2006-01-02")
	}
	return s.repository.PrintDocument(ctx, kind, id, start, end, partyID)
}
func (s *ReportingService) ShopSettings(ctx context.Context) (domain.ShopSettings, error) {
	return s.repository.GetShopSettings(ctx)
}
func (s *ReportingService) SaveShopSettings(ctx context.Context, v domain.ShopSettings) error {
	if strings.TrimSpace(v.ShopName) == "" {
		return fmt.Errorf("shop name is required")
	}
	return s.repository.SaveShopSettings(ctx, v)
}
