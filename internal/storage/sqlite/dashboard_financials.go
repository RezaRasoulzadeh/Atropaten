package sqlite

import (
	"context"
	"time"

	"Atropaten/internal/domain"
)

// Dashboard sales recognize confirmed/closed orders on their order date. Their
// current production forecast supplies cost, so invoice posting, payments and
// stock journals never count the same order twice. Standalone ledger activity
// retains its recorded basis. Both cards and charts use this single series.
func (s *Store) dashboardFinancialTrend(ctx context.Context, start, end time.Time, d *domain.Dashboard) error {
	from, until := reportWindow(start, end)
	rows, err := s.db.QueryContext(ctx, `SELECT id,created_at,total_rial FROM orders
	 WHERE commercial_status IN ('Confirmed','Closed') AND created_at>=? AND created_at<?
	 ORDER BY created_at,id`, from, until)
	if err != nil {
		return err
	}
	type sale struct {
		id, at  string
		revenue int64
	}
	var sales []sale
	for rows.Next() {
		var v sale
		if err = rows.Scan(&v.id, &v.at, &v.revenue); err != nil {
			rows.Close()
			return err
		}
		sales = append(sales, v)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	daily := map[string]domain.DashboardTrend{}
	add := func(stamp string, revenue, profit int64) error {
		at, err := time.Parse(time.RFC3339Nano, stamp)
		if err != nil {
			return err
		}
		date := at.In(start.Location()).Format(time.DateOnly)
		point := daily[date]
		point.RevenueRial, err = sumProductionCosts(point.RevenueRial, revenue)
		if err != nil {
			return err
		}
		point.GrossProfitRial, err = sumProductionCosts(point.GrossProfitRial, profit)
		if err != nil {
			return err
		}
		daily[date] = point
		return nil
	}
	for _, v := range sales {
		cost, err := s.ProductionProjectedCostSummary(ctx, v.id)
		if err != nil {
			return err
		}
		if err = add(v.at, v.revenue, v.revenue-cost); err != nil {
			return err
		}
	}
	rows, err = s.db.QueryContext(ctx, `SELECT je.posted_at,
	 SUM(CASE WHEN jl.account_id='ACC-REVENUE' THEN jl.credit_rial-jl.debit_rial ELSE 0 END),
	 SUM(jl.credit_rial-jl.debit_rial)
	 FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id
	 LEFT JOIN invoices i ON i.id=je.source_id AND je.source_type IN ('invoice','invoice_cogs','invoice_cogs_adjustment')
	 LEFT JOIN production_jobs pj ON je.source_type='expense' AND je.source_id='EXP-OUTSOURCE-'||pj.id
	 WHERE (jl.account_id IN ('ACC-REVENUE','ACC-COGS') OR
	 (je.source_type='expense' AND je.source_id LIKE 'EXP-OUTSOURCE-%' AND jl.account_id='ACC-EXP-OTHER'))
	 AND COALESCE(i.order_id,pj.order_id,'')=''
	 AND NOT EXISTS(SELECT 1 FROM deleted_order_records deleted WHERE deleted.invoice_id=i.id)
	 AND je.posted_at>=? AND je.posted_at<?
	 GROUP BY je.id ORDER BY je.posted_at,je.id`, from, until)
	if err != nil {
		return err
	}
	for rows.Next() {
		var at string
		var revenue, profit int64
		if err = rows.Scan(&at, &revenue, &profit); err != nil {
			rows.Close()
			return err
		}
		if err = add(at, revenue, profit); err != nil {
			rows.Close()
			return err
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	d.RevenueRial, d.GrossProfitRial = 0, 0
	d.Trend = nil
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		key := date.Format(time.DateOnly)
		point := daily[key]
		point.Date = key
		d.Trend = append(d.Trend, point)
		d.RevenueRial, err = sumProductionCosts(d.RevenueRial, point.RevenueRial)
		if err != nil {
			return err
		}
		d.GrossProfitRial, err = sumProductionCosts(d.GrossProfitRial, point.GrossProfitRial)
		if err != nil {
			return err
		}
	}
	return nil
}
