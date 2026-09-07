package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
)

func TestReportsReconcileJournalLinesAndPersistedSettings(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "reporting.db"))
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
	if err = s.SaveShopSettings(ctx, settings); err != nil {
		t.Fatal(err)
	}
	saved, err := s.GetShopSettings(ctx)
	if err != nil || saved.ShopName != "Reconciled Shop" {
		t.Fatalf("settings=%+v err=%v", saved, err)
	}
	if _, err = s.Dashboard(ctx, when.Add(-time.Hour), when.Add(time.Hour)); err != nil {
		t.Fatal("dashboard query: ", err)
	}
	for _, kind := range []string{"receivables", "payables", "expenses", "sales_by_service", "customer_sales", "material_usage", "production"} {
		if _, err = s.Report(ctx, kind, when.Add(-time.Hour), when.Add(time.Hour)); err != nil {
			t.Fatalf("%s query: %v", kind, err)
		}
	}
	quote := domain.NewQuote("QUO-report", "", when)
	quote.CustomerNameSnapshot = "Historical Customer"
	quote.Items = []domain.QuoteItem{{ID: "QIT-report", QuoteID: quote.ID, Position: 0, ServiceNameSnapshot: "Saved service", Quantity: domain.QuantityScale, QuantityUnit: "unit", SellingPriceRial: 1234}}
	if err = quote.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err = s.CreateQuote(ctx, quote); err != nil {
		t.Fatal(err)
	}
	doc, err := s.PrintDocument(ctx, "quote", quote.ID, "", "", "")
	if err != nil || doc.CustomerName != "Historical Customer" || doc.TotalRial != 1234 || len(doc.Lines) != 1 {
		t.Fatalf("print document=%+v err=%v", doc, err)
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
