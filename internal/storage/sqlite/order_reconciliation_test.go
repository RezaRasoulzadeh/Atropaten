package sqlite

import (
	"context"
	"reflect"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func saveEditedOrder(t *testing.T, s *Store, o domain.Order) domain.Order {
	t.Helper()
	if err := o.RecalculateTotals(); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveOrder(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	saved, err := s.GetOrder(context.Background(), o.ID)
	if err != nil {
		t.Fatal(err)
	}
	return saved
}

func assertOrderInventory(t *testing.T, s *Store, physical, reserved domain.Quantity) {
	t.Helper()
	state, err := s.InventoryState(context.Background(), "MAT-paper")
	if err != nil || state.PhysicalStock != physical*domain.QuantityScale || state.ReservedStock != reserved*domain.QuantityScale {
		t.Fatalf("inventory=%+v want=%d/%d err=%v", state, physical, reserved, err)
	}
}

func createOrderInvoice(t *testing.T, s *Store, orderID string, posted bool) domain.Invoice {
	t.Helper()
	ctx := context.Background()
	v, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	if posted {
		if err = s.PostInvoice(ctx, v.ID); err != nil {
			t.Fatal(err)
		}
	}
	invoice, err := s.GetInvoice(ctx, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	return invoice
}

func assertInvoiceHistory(t *testing.T, s *Store, old domain.Invoice, journal domain.JournalEntry) domain.Invoice {
	t.Helper()
	ctx := context.Background()
	v, err := s.GetInvoice(ctx, old.ID)
	if err != nil {
		t.Fatal(err)
	}
	if v.Status != domain.InvoiceVoided || !reflect.DeepEqual(v.Items, old.Items) || v.TotalRial != old.TotalRial || v.DiscountRial != old.DiscountRial || v.CustomerID != old.CustomerID || v.Notes != old.Notes || v.AccountingJournalEntryID != old.AccountingJournalEntryID {
		t.Fatalf("historical invoice changed: %+v", v)
	}
	gotJournal, err := s.GetJournalEntry(ctx, journal.ID)
	if err != nil || !reflect.DeepEqual(gotJournal, journal) {
		t.Fatalf("historical journal changed: %+v err=%v", gotJournal, err)
	}
	current, err := s.GetInvoiceForOrder(ctx, old.OrderID)
	if err != nil || current.ID == old.ID || current.Status == domain.InvoiceVoided {
		t.Fatalf("current invoice=%+v err=%v", current, err)
	}
	return current
}

func TestOrderEditsReconcileConfirmedProductionAndPartialUsage(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		used, quantity, reserved domain.Quantity
	}{
		{"increase", 0, 15, 30}, {"decrease", 0, 5, 10},
		{"partial increase", 8, 15, 22}, {"below partial usage", 12, 5, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, o, j := productionFlowFixture(t)
			ctx := context.Background()
			if tc.used > 0 {
				if err := s.TransitionProductionJob(ctx, j.ID, domain.ProductionInProgress); err != nil {
					t.Fatal(err)
				}
				if _, err := s.RecordProductionConsumption(ctx, j.ID, "MAT-paper", "partial", tc.used*domain.QuantityScale, 0, ""); err != nil {
					t.Fatal(err)
				}
			}
			history := productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements")
			o.Items[0].Quantity = tc.quantity * domain.QuantityScale
			o = saveEditedOrder(t, s, o)
			assertOrderInventory(t, s, 200-tc.used, tc.reserved)
			if productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements") != history {
				t.Fatal("order edit changed usage history")
			}
			snapshot := productionReopenSnapshot(t, s, "production_jobs", "production_material_plans", "inventory_reservations", "journal_entries")
			saveEditedOrder(t, s, o)
			if productionReopenSnapshot(t, s, "production_jobs", "production_material_plans", "inventory_reservations", "journal_entries") != snapshot {
				t.Fatal("retry changed production or accounting")
			}
		})
	}
}

func TestOrderEditsAfterCompletionReconcileOnlyOutstandingNeed(t *testing.T) {
	for _, tc := range []struct {
		name               string
		quantity, reserved domain.Quantity
		status             string
	}{
		{"additional work", 15, 10, domain.ProductionPending},
		{"less than produced", 5, 0, domain.ProductionCompleted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, o, j := completedProductionFixture(t)
			history := productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements")
			o.Items[0].Quantity = tc.quantity * domain.QuantityScale
			o = saveEditedOrder(t, s, o)
			assertOrderInventory(t, s, 180, tc.reserved)
			current, err := s.GetProductionJob(context.Background(), j.ID)
			if err != nil || current.Status != tc.status || current.ActualMaterialCostRial != 2000 {
				t.Fatalf("production=%+v err=%v", current, err)
			}
			if tc.status == domain.ProductionCompleted && (current.CompletedAt == nil || !current.CompletedAt.Equal(*j.CompletedAt)) {
				t.Fatal("completed historical timestamp lost")
			}
			if tc.status == domain.ProductionPending && (current.CompletedAt != nil || o.FulfillmentStatus != domain.FulfillmentInProduction) {
				t.Fatal("additional work not reflected in status")
			}
			if productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements") != history {
				t.Fatal("history rewritten")
			}
			assertProductionCOGS(t, s, 2000)
		})
	}
}

func TestOrderEditResnapshotsDraftInvoiceAndMetadata(t *testing.T) {
	s, o, _ := productionFlowFixture(t)
	ctx := context.Background()
	original := createOrderInvoice(t, s, o.ID, false)
	due := time.Now().UTC().AddDate(0, 0, 30)
	original.DueDate = &due
	if err := s.SaveInvoice(ctx, original); err != nil {
		t.Fatal(err)
	}
	o.Items[0].ServiceID = "changed-service"
	o.Items[0].ServiceNameSnapshot = "Changed description"
	o.Items[0].Quantity = 5 * domain.QuantityScale
	o.Items[0].QuantityUnit = "panel"
	o.Items[0].SellingPriceRial = 15000
	o.Items[0].Notes = "Line notes"
	o = saveEditedOrder(t, s, o)
	view, err := application.NewOrdersService(s, s, nil).Update(ctx, o.ID, application.OrderInput{Notes: "Commercial notes", DiscountRial: 3000})
	if err != nil || view.TotalRial != 12000 || view.InvoicedTotalRial != 12000 {
		t.Fatalf("view=%+v err=%v", view, err)
	}
	current, err := s.GetInvoiceForOrder(ctx, o.ID)
	if err != nil || current.ID != original.ID || current.Status != domain.InvoiceDraft || current.TotalRial != 12000 || current.DiscountRial != 3000 || current.Notes != "Commercial notes" || !current.DueDate.Equal(due) {
		t.Fatalf("draft=%+v err=%v", current, err)
	}
	line := current.Items[0]
	if line.DescriptionSnapshot != "Changed description" || line.ServiceID != "changed-service" || line.QuantityUnits != 5*domain.QuantityScale || line.QuantityUnit != "panel" || line.UnitPriceRial != 3000 || line.LineTotalRial != 15000 || line.Notes != "Line notes" {
		t.Fatalf("draft line=%+v", line)
	}
	if countRows(t, s, `SELECT COUNT(*) FROM journal_entries`) != 0 {
		t.Fatal("draft edit posted accounting")
	}
}

func TestPostedOrderEditReissuesWithoutLosingHistoryOrDuplicatingCOGS(t *testing.T) {
	s, o, j := completedProductionFixture(t)
	ctx := context.Background()
	old, err := s.GetInvoiceForOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := s.GetJournalEntry(ctx, old.AccountingJournalEntryID)
	if err != nil {
		t.Fatal(err)
	}
	o.Items[0].Quantity = 15 * domain.QuantityScale
	o.Items[0].SellingPriceRial = 15000
	o = saveEditedOrder(t, s, o)
	current := assertInvoiceHistory(t, s, old, journal)
	if current.TotalRial != 15000 || current.Items[0].QuantityUnits != 15*domain.QuantityScale {
		t.Fatalf("replacement=%+v", current)
	}
	assertProductionCOGS(t, s, 2000)
	for _, status := range []string{domain.ProductionInProgress, domain.ProductionCompleted, domain.ProductionCompleted} {
		if err = s.TransitionProductionJob(ctx, j.ID, status); err != nil {
			t.Fatal(err)
		}
	}
	assertOrderInventory(t, s, 170, 0)
	assertProductionCOGS(t, s, 3000)
	if countRows(t, s, `SELECT COUNT(*) FROM journal_entries WHERE source_type='invoice_cogs_adjustment'`) != 1 {
		t.Fatal("late usage did not adjust exactly once")
	}
	o, err = s.GetOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	before := productionReopenSnapshot(t, s, "invoices", "invoice_items", "journal_entries", "journal_lines", "invoice_replacements")
	saveEditedOrder(t, s, o)
	if productionReopenSnapshot(t, s, "invoices", "invoice_items", "journal_entries", "journal_lines", "invoice_replacements") != before {
		t.Fatal("same order save duplicated invoice reconciliation")
	}
	o.DiscountRial = 2000
	o = saveEditedOrder(t, s, o)
	latest, err := s.GetInvoiceForOrder(ctx, o.ID)
	if err != nil || latest.ID == current.ID || latest.TotalRial != 13000 {
		t.Fatalf("second replacement=%+v err=%v", latest, err)
	}
	if countRows(t, s, `SELECT COUNT(*) FROM invoice_replacements`) != 2 {
		t.Fatal("replacement chain is not one per edit")
	}
	assertProductionCOGS(t, s, 3000)
	var revenue int64
	if err = s.db.QueryRow(`SELECT SUM(credit_rial-debit_rial) FROM journal_lines WHERE account_id='ACC-REVENUE'`).Scan(&revenue); err != nil || revenue != 13000 {
		t.Fatalf("revenue=%d err=%v", revenue, err)
	}
}

func TestRemoveInvoicedItemRetainsProductionAndFinancialHistory(t *testing.T) {
	for _, posted := range []bool{false, true} {
		t.Run(map[bool]string{false: "draft", true: "posted"}[posted], func(t *testing.T) {
			s, o, j := productionFlowFixture(t)
			ctx := context.Background()
			if err := s.TransitionProductionJob(ctx, j.ID, domain.ProductionInProgress); err != nil {
				t.Fatal(err)
			}
			if _, err := s.RecordProductionConsumption(ctx, j.ID, "MAT-paper", "partial", 8*domain.QuantityScale, 0, ""); err != nil {
				t.Fatal(err)
			}
			old := createOrderInvoice(t, s, o.ID, posted)
			history := productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements")
			view, err := application.NewOrdersService(s, s, nil).RemoveItem(ctx, o.ID, o.Items[0].ID)
			if err != nil || len(view.Items) != 0 || view.TotalRial != 0 {
				t.Fatalf("removed view=%+v err=%v", view, err)
			}
			assertOrderInventory(t, s, 192, 0)
			if productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements") != history {
				t.Fatal("removal returned actual usage")
			}
			job, err := s.GetProductionJob(ctx, j.ID)
			if err != nil || job.Status != domain.ProductionCancelled || job.ActualMaterialCostRial != 800 {
				t.Fatalf("historical job=%+v err=%v", job, err)
			}
			current, err := s.GetInvoiceForOrder(ctx, o.ID)
			if err != nil || len(current.Items) != 0 || current.TotalRial != 0 {
				t.Fatalf("current invoice=%+v err=%v", current, err)
			}
			if posted {
				previous, err := s.GetInvoice(ctx, old.ID)
				if err != nil || previous.Status != domain.InvoiceVoided || !reflect.DeepEqual(old.Items, previous.Items) {
					t.Fatalf("removed historical lines=%+v err=%v", previous, err)
				}
				assertProductionCOGS(t, s, 800)
			} else if current.ID != old.ID || current.Status != domain.InvoiceDraft {
				t.Fatal("draft was replaced")
			}
		})
	}
}

func payOrderForEdit(t *testing.T, s *Store, id, customer string, amount int64, allocations ...domain.PaymentAllocation) domain.Payment {
	t.Helper()
	now := time.Now().UTC()
	p, err := s.CreatePayment(context.Background(), domain.Payment{ID: id, CustomerID: customer, Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", AmountRial: amount, PostedAt: now, CreatedAt: now, Allocations: allocations})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestOrderEditMovesExcessPaymentsToCreditAndPaymentCanStillReverse(t *testing.T) {
	for _, posted := range []bool{false, true} {
		t.Run(map[bool]string{false: "order deposit", true: "posted invoice"}[posted], func(t *testing.T) {
			s, o, _ := productionFlowFixture(t)
			ctx := context.Background()
			typ, target := "order", o.ID
			if posted {
				inv := createOrderInvoice(t, s, o.ID, true)
				typ, target = "invoice", inv.ID
			}
			payment := payOrderForEdit(t, s, "PAY-edit", "", 9000, domain.PaymentAllocation{TargetType: typ, TargetID: target, AmountRial: 9000})
			cashBefore, err := s.GetFinancialAccount(ctx, "FIN-CASH")
			if err != nil {
				t.Fatal(err)
			}
			paymentHistory := productionReopenSnapshot(t, s, "payments")
			view, err := application.NewOrdersService(s, s, nil).ApplyDiscount(ctx, o.ID, 4000)
			if err != nil || view.TotalRial != 6000 || view.PaidRial != 6000 || view.RemainingRial != 0 || view.PaymentStatus != string(domain.PaymentPaid) {
				t.Fatalf("payment view=%+v err=%v", view, err)
			}
			if productionReopenSnapshot(t, s, "payments") != paymentHistory {
				t.Fatal("cash payment rewritten")
			}
			_, credit, err := s.CustomerFinancialSummary(ctx, "")
			if err != nil || credit != 3000 {
				t.Fatalf("credit=%d err=%v", credit, err)
			}
			cashAfter, err := s.GetFinancialAccount(ctx, "FIN-CASH")
			if err != nil || cashAfter.BalanceRial != cashBefore.BalanceRial {
				t.Fatal("cash changed during edit")
			}
			snapshot := productionReopenSnapshot(t, s, "payment_allocations", "journal_entries", "invoices")
			if _, err = application.NewOrdersService(s, s, nil).ApplyDiscount(ctx, o.ID, 4000); err != nil {
				t.Fatal(err)
			}
			if productionReopenSnapshot(t, s, "payment_allocations", "journal_entries", "invoices") != snapshot {
				t.Fatal("retry duplicated credit")
			}
			if _, err = s.ReversePayment(ctx, payment.ID, "reverse-edited-payment"); err != nil {
				t.Fatal(err)
			}
			_, credit, err = s.CustomerFinancialSummary(ctx, "")
			if err != nil || credit != 0 {
				t.Fatalf("credit left after payment reversal=%d err=%v", credit, err)
			}
			paid, _, _, err := s.OrderPaymentSummary(ctx, o.ID)
			if err != nil || paid != 0 {
				t.Fatalf("paid after reversal=%d err=%v", paid, err)
			}
		})
	}
}

func TestOrderEditFailureRollsBackEveryDependentChange(t *testing.T) {
	for _, stage := range []string{"materials", "replacement", "credit"} {
		t.Run(stage, func(t *testing.T) {
			s, o, _ := completedProductionFixture(t)
			ctx := context.Background()
			inv, err := s.GetInvoiceForOrder(ctx, o.ID)
			if err != nil {
				t.Fatal(err)
			}
			payOrderForEdit(t, s, "PAY-rollback", "", 9000, domain.PaymentAllocation{TargetType: "invoice", TargetID: inv.ID, AmountRial: 9000})
			switch stage {
			case "materials":
				o.Items[0].CostBreakdownJSON = `[{"type":"material","enabled":true,"materialId":"missing","usageQuantity":"2"}]`
			case "replacement":
				_, err = s.db.Exec(`CREATE TRIGGER fail_replacement BEFORE INSERT ON invoices WHEN NEW.id LIKE 'INV-REISSUE-%' BEGIN SELECT RAISE(ABORT,'replacement failure'); END`)
			case "credit":
				_, err = s.db.Exec(`CREATE TRIGGER fail_credit BEFORE INSERT ON journal_entries WHEN NEW.source_type='payment_allocation_adjustment' BEGIN SELECT RAISE(ABORT,'credit failure'); END`)
			}
			if err != nil {
				t.Fatal(err)
			}
			o.Items[0].Quantity = 15 * domain.QuantityScale
			o.DiscountRial = 4000
			if err = o.RecalculateTotals(); err != nil {
				t.Fatal(err)
			}
			tables := []string{"orders", "order_items", "production_jobs", "production_material_plans", "inventory_reservations", "inventory_movements", "production_consumptions", "invoices", "invoice_items", "invoice_order_snapshots", "invoice_replacements", "payments", "payment_allocations", "journal_entries", "journal_lines", "invoice_number_sequences", "journal_number_sequences"}
			before := productionReopenSnapshot(t, s, tables...)
			for n := 0; n < 2; n++ {
				if err = s.SaveOrder(ctx, o); err == nil {
					t.Fatal("expected reconciliation failure")
				}
				if productionReopenSnapshot(t, s, tables...) != before {
					t.Fatal("failed order edit left partial changes")
				}
			}
		})
	}
}

func TestOrderEditsRemainPermissiveAcrossCommercialStates(t *testing.T) {
	for _, status := range []domain.CommercialStatus{domain.CommercialDraft, domain.CommercialConfirmed, domain.CommercialClosed, domain.CommercialCancelled} {
		t.Run(string(status), func(t *testing.T) {
			s, o, j := productionFlowFixture(t)
			ctx := context.Background()
			service := application.NewOrdersService(s, s, nil)
			if _, err := service.SetCommercialStatus(ctx, o.ID, string(status)); err != nil {
				t.Fatal(err)
			}
			o, err := s.GetOrder(ctx, o.ID)
			if err != nil {
				t.Fatal(err)
			}
			o.Items[0].Quantity = 15 * domain.QuantityScale
			o = saveEditedOrder(t, s, o)
			var required domain.Quantity
			if err = s.db.QueryRow(`SELECT required_units FROM production_material_plans WHERE production_job_id=?`, j.ID).Scan(&required); err != nil || required != 30*domain.QuantityScale {
				t.Fatalf("required=%s err=%v", required, err)
			}
			reserved := domain.Quantity(30)
			if status == domain.CommercialCancelled || status == domain.CommercialDraft {
				reserved = 0
			}
			assertOrderInventory(t, s, 200, reserved)
			if _, err = service.SetCommercialStatus(ctx, o.ID, string(domain.CommercialConfirmed)); err != nil {
				t.Fatal(err)
			}
			assertOrderInventory(t, s, 200, 30)
		})
	}
}

func TestOrderCustomerEditLeavesOriginalCustomersMoneyAsCredit(t *testing.T) {
	s, o, _ := productionFlowFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()
	for _, customer := range []domain.Customer{{ID: "CUS-old", Name: "Original", Phone: "111", Active: true, CreatedAt: now, UpdatedAt: now}, {ID: "CUS-new", Name: "New", Phone: "222", Active: true, CreatedAt: now, UpdatedAt: now}} {
		if err := s.SaveCustomer(ctx, customer); err != nil {
			t.Fatal(err)
		}
	}
	service := application.NewOrdersService(s, s, nil)
	if _, err := service.Update(ctx, o.ID, application.OrderInput{CustomerID: "CUS-old"}); err != nil {
		t.Fatal(err)
	}
	old := createOrderInvoice(t, s, o.ID, true)
	payOrderForEdit(t, s, "PAY-customer-change", "CUS-old", 10000, domain.PaymentAllocation{TargetType: "invoice", TargetID: old.ID, AmountRial: 9000})
	view, err := service.Update(ctx, o.ID, application.OrderInput{CustomerID: "CUS-new", Notes: "New billing customer"})
	if err != nil || view.CustomerID != "CUS-new" || view.PaidRial != 0 || view.RemainingRial != 10000 {
		t.Fatalf("customer edit=%+v err=%v", view, err)
	}
	current, err := s.GetInvoiceForOrder(ctx, o.ID)
	if err != nil || current.CustomerID != "CUS-new" || current.CustomerNameSnapshot != "New" || current.CustomerPhoneSnapshot != "222" {
		t.Fatalf("invoice customer=%+v err=%v", current, err)
	}
	receivable, credit, err := s.CustomerFinancialSummary(ctx, "CUS-old")
	if err != nil || receivable != 0 || credit != 10000 {
		t.Fatalf("old customer AR=%d credit=%d err=%v", receivable, credit, err)
	}
	receivable, credit, err = s.CustomerFinancialSummary(ctx, "CUS-new")
	if err != nil || receivable != 10000 || credit != 0 {
		t.Fatalf("new customer AR=%d credit=%d err=%v", receivable, credit, err)
	}
	if _, err = s.ReversePayment(ctx, "PAY-customer-change", ""); err != nil {
		t.Fatal(err)
	}
	_, credit, err = s.CustomerFinancialSummary(ctx, "CUS-old")
	if err != nil || credit != 0 {
		t.Fatalf("reversed credit=%d err=%v", credit, err)
	}
}

func TestOrderEditsAddReplaceAndRemoveCompletedItems(t *testing.T) {
	s, o, j := completedProductionFixture(t)
	ctx := context.Background()
	// Replacement uses the stable item/job identity and derives a new material
	// plan from the current resolved selection, without returning the old paper.
	material, err := domain.NewMaterial("MAT-new", domain.MaterialDraft{Name: "New stock", PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale, PhysicalStock: 100 * domain.QuantityScale, AverageUnitCostRial: 200}, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, material); err != nil {
		t.Fatal(err)
	}
	history := productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements")
	o.Items[0].ResolvedParametersJSON = `[{"key":"paper","materialId":"MAT-new"}]`
	o = saveEditedOrder(t, s, o)
	state, err := s.InventoryState(ctx, material.ID)
	if err != nil || state.ReservedStock != 20*domain.QuantityScale || state.PhysicalStock != 100*domain.QuantityScale {
		t.Fatalf("new material=%+v err=%v", state, err)
	}
	assertOrderInventory(t, s, 180, 0)
	// New item receives one outstanding job. Retrying must not create another.
	item := o.Items[0]
	item.ID, item.Position = "ITEM-added", 1
	o.Items = append(o.Items, item)
	o = saveEditedOrder(t, s, o)
	saveEditedOrder(t, s, o)
	if countRows(t, s, `SELECT COUNT(*) FROM production_jobs`) != 2 {
		t.Fatal("added item did not create exactly one job")
	}
	// Remove both current items; retain the consumed job and all historical links.
	o.Items = nil
	o = saveEditedOrder(t, s, o)
	if len(o.Items) != 0 {
		t.Fatal("historical items leaked into current order")
	}
	if productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements") != history {
		t.Fatal("replacement/removal rewrote usage")
	}
	job, err := s.GetProductionJob(ctx, j.ID)
	if err != nil || job.ActualMaterialCostRial != 2000 {
		t.Fatalf("removed historical job=%+v err=%v", job, err)
	}
	state, err = s.InventoryState(ctx, material.ID)
	if err != nil || state.ReservedStock != 0 {
		t.Fatalf("removed commitments=%+v err=%v", state, err)
	}
	assertProductionCOGS(t, s, 2000)
	snapshot := productionReopenSnapshot(t, s, "order_items", "invoices", "invoice_items", "journal_entries", "production_jobs", "inventory_reservations")
	saveEditedOrder(t, s, o)
	if productionReopenSnapshot(t, s, "order_items", "invoices", "invoice_items", "journal_entries", "production_jobs", "inventory_reservations") != snapshot {
		t.Fatal("retry changed retained history")
	}
}

func TestOrderQuantityBelowOutsourcedSharePreservesPaidExpense(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	j.OutsourceQuantity = 8 * domain.QuantityScale
	j.OutsourceUnitCostRial = 300
	j.OutsourceFinancialAccountID = "FIN-CASH"
	if err := s.UpdateProductionOutsourcing(ctx, j); err != nil {
		t.Fatal(err)
	}
	history := productionReopenSnapshot(t, s, "expenses", "journal_entries", "journal_lines", "inventory_movements", "production_consumptions")
	o.Items[0].Quantity = 5 * domain.QuantityScale
	saveEditedOrder(t, s, o)
	assertOrderInventory(t, s, 200, 0)
	if productionReopenSnapshot(t, s, "expenses", "journal_entries", "journal_lines", "inventory_movements", "production_consumptions") != history {
		t.Fatal("quantity edit reversed paid outsourcing")
	}
}

func TestRemoveCompletedOrderItemPreservesHistoricalJob(t *testing.T) {
	s, o, j := completedProductionFixture(t)
	ctx := context.Background()
	history := productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements")
	if _, err := application.NewOrdersService(s, s, nil).RemoveItem(ctx, o.ID, o.Items[0].ID); err != nil {
		t.Fatal(err)
	}
	job, err := s.GetProductionJob(ctx, j.ID)
	if err != nil || job.Status != domain.ProductionCompleted || job.CompletedAt == nil || !job.CompletedAt.Equal(*j.CompletedAt) {
		t.Fatalf("historical job=%+v err=%v", job, err)
	}
	if productionReopenSnapshot(t, s, "production_consumptions", "inventory_movements") != history {
		t.Fatal("removal rewrote completed material usage")
	}
	assertOrderInventory(t, s, 180, 0)
	assertProductionCOGS(t, s, 2000)
}

func TestCompletedOrderConfigurationEditReopensWorkWithoutReconsumingMaterial(t *testing.T) {
	s, o, j := completedProductionFixture(t)
	o.Items[0].ResolvedParametersJSON = `[{"key":"paper","materialId":"MAT-paper"},{"key":"finishing","value":"fold"}]`
	o = saveEditedOrder(t, s, o)
	job, err := s.GetProductionJob(context.Background(), j.ID)
	if err != nil || job.Status != domain.ProductionPending || job.CompletedAt != nil || o.FulfillmentStatus != domain.FulfillmentInProduction {
		t.Fatalf("configuration job=%+v err=%v", job, err)
	}
	assertOrderInventory(t, s, 180, 0)
	assertProductionCOGS(t, s, 2000)
}
