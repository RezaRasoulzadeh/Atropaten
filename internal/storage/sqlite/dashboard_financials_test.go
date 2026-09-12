package sqlite

import (
	"context"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestDashboardConfirmedOrderProductionAndInvoiceFlow(t *testing.T) {
	s, order, job := productionFlowFixture(t)
	ctx := context.Background()
	order.Items[0].CostBreakdownJSON = `[{"type":"material","enabled":true,"materialId":"MAT-paper","usageQuantity":"2","amountRial":200},{"type":"machine","enabled":true,"amountRial":100}]`
	order.Items[0].EstimatedCostRial = 3000
	order.DiscountRial = 1000
	if err := order.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	created, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	invoice, err := s.GetInvoice(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	invoice.IssueDate = yesterday.Add(12 * time.Hour)
	if err = s.SaveInvoice(ctx, invoice); err != nil {
		t.Fatal(err)
	}
	check := func(revenue, profit int64) domain.Dashboard {
		t.Helper()
		d, err := s.Dashboard(ctx, yesterday, today)
		if err != nil {
			t.Fatal(err)
		}
		var chartRevenue, chartProfit int64
		for _, point := range d.Trend {
			chartRevenue += point.RevenueRial
			chartProfit += point.GrossProfitRial
		}
		if d.RevenueRial != revenue || d.GrossProfitRial != profit || chartRevenue != revenue || chartProfit != profit {
			t.Fatalf("totals=%d/%d chart=%d/%d want=%d/%d trend=%+v", d.RevenueRial, d.GrossProfitRial, chartRevenue, chartProfit, revenue, profit, d.Trend)
		}
		return d
	}
	check(9000, 6000) // Confirmation recognizes sales even while the invoice is Draft.
	if err = s.PostInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	d := check(9000, 6000)
	if d.Trend[1].RevenueRial != 9000 || d.Trend[1].GrossProfitRial != 6000 || d.Trend[0].GrossProfitRial != 0 {
		t.Fatalf("sale day=%+v", d.Trend)
	}
	if err = s.SetProductionMaterialTarget(ctx, job.ID, "MAT-paper", 30*domain.QuantityScale); err != nil {
		t.Fatal(err)
	}
	check(9000, 5000) // Extra planned material changes the forecast immediately.
	job, err = s.GetProductionJob(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	job.OutsourceQuantity = 5 * domain.QuantityScale
	job.OutsourceUnitCostRial = 600
	job.OutsourceFinancialAccountID = "FIN-CASH"
	if err = s.UpdateProductionOutsourcing(ctx, job); err != nil {
		t.Fatal(err)
	}
	check(9000, 4000) // 3000 outsourced + 1500 materials + 500 conversion.
	if err = s.TransitionProductionJob(ctx, job.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, job.ID, "Completed"); err != nil {
		t.Fatal(err)
	}
	check(9000, 4000) // Posting actual stock costs must not count them twice.
	view, err := application.NewOrdersService(s, s, nil).Get(ctx, order.ID)
	if err != nil || view.MarginRial != 4000 || view.ProjectedCostRial != 5000 {
		t.Fatalf("order=%+v err=%v", view, err)
	}
	if err = s.VoidInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	d = check(9000, 4000) // Voiding a document does not cancel the confirmed order.
	if d.Trend[0].RevenueRial != 0 || d.Trend[0].GrossProfitRial != 0 || d.Trend[1].RevenueRial != 9000 || d.Trend[1].GrossProfitRial != 4000 {
		t.Fatalf("invoice dates must not move the order's sale: %+v", d.Trend)
	}
	d, err = s.Dashboard(ctx, yesterday, yesterday)
	if err != nil || d.RevenueRial != 0 || d.GrossProfitRial != 0 {
		t.Fatalf("invoice-only window=%+v err=%v", d, err)
	}
}

func TestDashboardKeepsUnlinkedLedgerCostsAlongsideInvoiceForecast(t *testing.T) {
	s, order, _ := productionFlowFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()
	invoice, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	_, err = s.PostJournalEntry(ctx, domain.JournalEntry{ID: "JE-mixed", Description: "Other sales and cost", SourceType: "test", SourceID: "mixed", IdempotencyKey: "mixed", PostedAt: now, CreatedAt: now, Lines: []domain.JournalLine{
		{ID: "mixed-cash", JournalEntryID: "JE-mixed", Position: 0, AccountID: "ACC-CASH", DebitRial: 500},
		{ID: "mixed-cost", JournalEntryID: "JE-mixed", Position: 1, AccountID: "ACC-COGS", DebitRial: 100},
		{ID: "mixed-sale", JournalEntryID: "JE-mixed", Position: 2, AccountID: "ACC-REVENUE", CreditRial: 600},
	}})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Dashboard(ctx, now.Truncate(24*time.Hour), now.Truncate(24*time.Hour))
	if err != nil || d.RevenueRial != 10600 || d.GrossProfitRial != 8500 || len(d.Trend) != 1 || d.Trend[0].RevenueRial != 10600 || d.Trend[0].GrossProfitRial != 8500 {
		t.Fatalf("mixed activity=%+v err=%v", d, err)
	}
}

func TestDashboardFollowsOrderStatusAndDeletion(t *testing.T) {
	s, order, _ := productionFlowFixture(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(24 * time.Hour)
	check := func(want int64) {
		t.Helper()
		d, err := s.Dashboard(ctx, now, now)
		if err != nil || d.RevenueRial != want {
			t.Fatalf("sales=%d want=%d err=%v", d.RevenueRial, want, err)
		}
	}
	for _, status := range []domain.CommercialStatus{domain.CommercialDraft, domain.CommercialConfirmed, domain.CommercialClosed, domain.CommercialCancelled, domain.CommercialConfirmed} {
		order.CommercialStatus = status
		if err := s.SaveOrderMetadata(ctx, order); err != nil {
			t.Fatal(err)
		}
		want := int64(0)
		if status == domain.CommercialConfirmed || status == domain.CommercialClosed {
			want = order.TotalRial
		}
		check(want)
	}
	invoice, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, invoice.ID); err != nil {
		t.Fatal(err)
	}
	check(order.TotalRial)
	if err = s.DeleteOrder(ctx, order.ID); err != nil {
		t.Fatal(err)
	}
	check(0)
	d, err := s.Dashboard(ctx, now, now)
	if err != nil || d.GrossProfitRial != 0 {
		t.Fatalf("deleted order left profit=%d err=%v", d.GrossProfitRial, err)
	}
}
