package sqlite

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestReportsReconcileJournalLinesAndPersistedSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reporting.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	when := time.Date(2026, 2, 10, 8, 0, 0, 0, time.UTC)
	entry := domain.JournalEntry{ID: "JE-REPORT-1", Description: "Report activity", SourceType: "test", SourceID: "REPORT", IdempotencyKey: "report-activity", PostedAt: when, CreatedAt: when, Lines: []domain.JournalLine{
		{ID: "JE-REPORT-1-L1", JournalEntryID: "JE-REPORT-1", Position: 0, AccountID: "ACC-CASH", DebitRial: 700},
		{ID: "JE-REPORT-1-L2", JournalEntryID: "JE-REPORT-1", Position: 1, AccountID: "ACC-REVENUE", CreditRial: 1000},
		{ID: "JE-REPORT-1-L3", JournalEntryID: "JE-REPORT-1", Position: 2, AccountID: "ACC-COGS", DebitRial: 200},
		{ID: "JE-REPORT-1-L4", JournalEntryID: "JE-REPORT-1", Position: 3, AccountID: "ACC-EXP-OTHER", DebitRial: 100},
	}}
	if _, err = s.PostJournalEntry(ctx, entry); err != nil {
		t.Fatal(err)
	}
	report, err := s.Report(ctx, "profit_loss", when.Add(-time.Hour), when.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int64{}
	for _, summary := range report.Summaries {
		got[summary.Key] = summary.AmountRial
	}
	if got["revenue"] != 1000 || got["cogs"] != 200 || got["expenses"] != 100 || got["net_profit"] != 700 {
		t.Fatalf("P&L summaries=%v", got)
	}
	cash, err := s.Report(ctx, "cash_bank", when.Add(-time.Hour), when.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	var cashBalance int64
	for _, row := range cash.Rows {
		if row.ID == "FIN-CASH" {
			cashBalance = row.AmountRial
		}
	}
	if cashBalance != 700 {
		t.Fatalf("cash rows=%+v", cash.Rows)
	}
	settings, err := s.GetShopSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	settings.ShopName = "Reconciled Shop"
	settings.MonetaryRoundingStepRial = 10000
	if err = s.SaveShopSettings(ctx, settings); err != nil {
		t.Fatal(err)
	}
	saved, err := s.GetShopSettings(ctx)
	if err != nil || saved.ShopName != "Reconciled Shop" || saved.MonetaryRoundingStepRial != 10000 {
		t.Fatalf("settings=%+v err=%v", saved, err)
	}
	dashboard, err := s.Dashboard(ctx, when.Add(-time.Hour), when.Add(time.Hour))
	if err != nil {
		t.Fatal("dashboard query: ", err)
	}
	if dashboard.RevenueRial != 1000 || dashboard.GrossProfitRial != 800 {
		t.Fatalf("dashboard financial totals=%+v, want journal-derived revenue/profit", dashboard)
	}
	for _, kind := range []string{"receivables", "payables", "expenses", "sales_by_service", "customer_sales", "material_usage", "production"} {
		if _, err = s.Report(ctx, kind, when.Add(-time.Hour), when.Add(time.Hour)); err != nil {
			t.Fatalf("%s query: %v", kind, err)
		}
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	persisted, err := reopened.GetShopSettings(ctx)
	if err != nil || persisted.MonetaryRoundingStepRial != 10000 {
		t.Fatalf("persisted rounding setting=%+v err=%v", persisted, err)
	}
}

func TestInventoryReportUsesMovementLedger(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "inventory-report.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 2, 11, 8, 0, 0, 0, time.UTC)
	if _, err = s.db.ExecContext(ctx, `INSERT INTO materials(id,name,purchase_unit,consumption_unit,conversion_factor_units,physical_stock_units,reorder_level_units,average_unit_cost_rial,created_at,updated_at) VALUES('MAT-report','Paper','pack','sheet',1000000,0,1000000,0,?,?)`, now.Format(time.RFC3339), now.Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES('MOV-report','MAT-report',?,'purchase',2000000,50,100,'purchase','PUR-report','test',?)`, now.Format(time.RFC3339), now.Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}
	report, err := s.Report(ctx, "inventory", now, now)
	if err != nil || len(report.Rows) != 1 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
	if report.Rows[0].QuantityUnits != 2000000 || report.Rows[0].AmountRial != 100 || report.Rows[0].SecondaryAmountRial != 50 {
		t.Fatalf("movement-derived row=%+v", report.Rows[0])
	}
}

func TestMaterialUsageReportAppliesCompensatingConsumptionMovements(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "usage-correction-report.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 2, 12, 8, 0, 0, 0, time.UTC)
	stamp := now.Format(time.RFC3339Nano)
	if _, err = s.db.ExecContext(ctx, `INSERT INTO materials(id,name,purchase_unit,consumption_unit,conversion_factor_units,physical_stock_units,reorder_level_units,average_unit_cost_rial,created_at,updated_at) VALUES('MAT-usage-correction','Paper','pack','sheet',1000000,0,0,0,?,?)`, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	for _, movement := range []struct {
		id, typ        string
		quantity, cost int64
	}{
		{"MOV-usage-original", "production_consumption", -2 * domain.QuantityScale, -200},
		{"MOV-usage-correction", "production_consumption", domain.QuantityScale, 100},
	} {
		if _, err = s.db.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, movement.id, "MAT-usage-correction", stamp, movement.typ, movement.quantity, 100, movement.cost, "production_correction", movement.id, "audit", stamp); err != nil {
			t.Fatal(err)
		}
	}
	report, err := s.Report(ctx, "material_usage", now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil || len(report.Rows) != 1 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
	if report.Rows[0].QuantityUnits != domain.QuantityScale || report.Rows[0].AmountRial != 100 {
		t.Fatalf("correction-aware usage row=%+v", report.Rows[0])
	}
}

func TestDashboardOrderFollowUpAndPipeline(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "dashboard.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		id          string
		commercial  domain.CommercialStatus
		fulfillment domain.FulfillmentStatus
		refs        int
		want        bool
	}{
		{"confirmed", domain.CommercialConfirmed, domain.FulfillmentPending, 0, true},
		{"reference-draft", domain.CommercialDraft, domain.FulfillmentPending, 2, true},
		{"both", domain.CommercialConfirmed, domain.FulfillmentInProduction, 2, true},
		{"plain-draft", domain.CommercialDraft, domain.FulfillmentPending, 0, false},
		{"delivered", domain.CommercialConfirmed, domain.FulfillmentDelivered, 1, false},
		{"cancelled", domain.CommercialCancelled, domain.FulfillmentPending, 1, false},
		{"closed", domain.CommercialClosed, domain.FulfillmentPending, 1, false},
	}
	for _, tc := range cases {
		o := domain.NewOrder(tc.id, "", now)
		o.CommercialStatus = tc.commercial
		o.FulfillmentStatus = tc.fulfillment
		if tc.id == "both" {
			o.PromisedAt = &now
		}
		if err := s.CreateOrder(ctx, o); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < tc.refs; i++ {
			a := domain.Attachment{ID: fmt.Sprintf("%s-%d", tc.id, i), OwnerType: domain.AttachmentOrder, OwnerID: tc.id, FileName: "reference.txt", Path: "reference.txt", Category: domain.AttachmentCategory("reference"), CreatedAt: now}
			if err := s.SaveAttachment(ctx, a); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Artwork alone must not qualify a draft.
	if err := s.SaveAttachment(ctx, domain.Attachment{ID: "art", OwnerType: domain.AttachmentOrder, OwnerID: "plain-draft", FileName: "art.txt", Path: "art.txt", Category: domain.AttachmentArtwork, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	d, err := s.Dashboard(ctx, now, now)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]domain.DashboardOrder{}
	for _, o := range d.OrdersNeedingAttention {
		got[o.ID] = o
	}
	if len(d.OrdersNeedingAttention) != 3 {
		t.Fatalf("follow-up rows=%+v", d.OrdersNeedingAttention)
	}
	for _, tc := range cases {
		o, ok := got[tc.id]
		if ok != tc.want || ok && o.ReferenceCount != tc.refs {
			t.Errorf("case %s: row=%+v present=%v", tc.id, o, ok)
		}
	}
	if d.OrdersNeedingAttention[0].ID != "both" {
		t.Fatal("scheduled orders should precede unscheduled orders")
	}
	counts := map[string]int{}
	for _, p := range d.Pipeline {
		counts[p.Status] = p.Count
	}
	if counts["Draft"] != 2 || counts["Pending"] != 1 || counts["In Production"] != 1 {
		t.Fatalf("pipeline=%v", counts)
	}
}

func TestDashboardTrendReconcilesAndIncludesZeroDays(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "trend.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	start := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	for i, amount := range []int64{1000, -300, 5000} {
		date := start.AddDate(0, 0, i*2)
		id := fmt.Sprintf("trend-%d", i)
		revenue := domain.JournalLine{ID: id + "-r", JournalEntryID: id, Position: 0, AccountID: "ACC-REVENUE"}
		cash := domain.JournalLine{ID: id + "-c", JournalEntryID: id, Position: 1, AccountID: "ACC-CASH"}
		if amount > 0 {
			revenue.CreditRial = amount
			cash.DebitRial = amount
		} else {
			revenue.DebitRial = -amount
			cash.CreditRial = -amount
		}
		_, err := s.PostJournalEntry(ctx, domain.JournalEntry{ID: id, Description: "Trend and reversal", SourceType: "test", SourceID: id, IdempotencyKey: id, PostedAt: date, CreatedAt: date, Lines: []domain.JournalLine{revenue, cash}})
		if err != nil {
			t.Fatal(err)
		}
	}
	d, err := s.Dashboard(ctx, start, start.AddDate(0, 0, 2))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Trend) != 3 || d.Trend[1].RevenueRial != 0 || d.Trend[2].RevenueRial != -300 {
		t.Fatalf("trend=%+v", d.Trend)
	}
	var revenue, profit int64
	for _, point := range d.Trend {
		revenue += point.RevenueRial
		profit += point.GrossProfitRial
	}
	if revenue != 700 || revenue != d.RevenueRial || profit != d.GrossProfitRial {
		t.Fatalf("trend does not reconcile: %+v", d)
	}
}
