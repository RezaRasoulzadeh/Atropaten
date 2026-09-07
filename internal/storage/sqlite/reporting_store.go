package sqlite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

func reportWindow(start, end time.Time) (string, string) {
	s, e := domain.ReportRange(start, end)
	return s.Format(time.RFC3339Nano), e.Format(time.RFC3339Nano)
}

func (s *Store) GetShopSettings(ctx context.Context) (domain.ShopSettings, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key,value FROM shop_settings ORDER BY key`)
	if err != nil {
		return domain.ShopSettings{}, err
	}
	defer rows.Close()
	v := domain.ShopSettings{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return v, err
		}
		switch key {
		case "shop_name":
			v.ShopName = value
		case "shop_subtitle":
			v.ShopSubtitle = value
		case "phone":
			v.Phone = value
		case "address":
			v.Address = value
		case "email":
			v.Email = value
		case "website":
			v.Website = value
		case "registration_id":
			v.RegistrationID = value
		case "tax_id":
			v.TaxID = value
		case "logo_path":
			v.LogoPath = value
		case "document_footer":
			v.DocumentFooter = value
		case "document_notes":
			v.DocumentNotes = value
		}
	}
	return v, rows.Err()
}

func (s *Store) SaveShopSettings(ctx context.Context, v domain.ShopSettings) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	values := map[string]string{
		"shop_name": v.ShopName, "shop_subtitle": v.ShopSubtitle, "phone": v.Phone, "address": v.Address,
		"email": v.Email, "website": v.Website, "registration_id": v.RegistrationID, "tax_id": v.TaxID,
		"logo_path": v.LogoPath, "document_footer": v.DocumentFooter, "document_notes": v.DocumentNotes,
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for key, value := range values {
		if _, err = tx.ExecContext(ctx, `INSERT INTO shop_settings(key,value,updated_at) VALUES(?,?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value,updated_at=excluded.updated_at`, key, strings.TrimSpace(value), now); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Report(ctx context.Context, kind string, start, end time.Time) (domain.Report, error) {
	from, until := reportWindow(start, end)
	r := domain.Report{Kind: kind, StartDate: start.Format("2006-01-02"), EndDate: end.Format("2006-01-02"), Rows: []domain.ReportRow{}, Summaries: []domain.ReportSummary{}}
	switch kind {
	case "profit_loss":
		return s.reportProfitLoss(ctx, r, from, until)
	case "cash_bank":
		return s.reportCashBank(ctx, r, from, until)
	case "receivables":
		return s.reportReceivables(ctx, r, from, until)
	case "payables":
		return s.reportPayables(ctx, r, from, until)
	case "expenses":
		return s.reportExpenses(ctx, r, from, until)
	case "inventory":
		return s.reportInventory(ctx, r)
	case "sales_by_service":
		return s.reportSalesByService(ctx, r, from, until)
	case "customer_sales":
		return s.reportCustomerSales(ctx, r, from, until)
	case "material_usage":
		return s.reportMaterialUsage(ctx, r, from, until)
	case "production":
		return s.reportProduction(ctx, r, from, until)
	default:
		return r, fmt.Errorf("unsupported report kind %q", kind)
	}
}

func (s *Store) reportProfitLoss(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,a.name,a.type,COALESCE(SUM(jl.credit_rial-jl.debit_rial),0),COALESCE(SUM(jl.debit_rial-jl.credit_rial),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id JOIN accounts a ON a.id=jl.account_id WHERE je.posted_at>=? AND je.posted_at<? AND je.source_type<>'profit_allocation' GROUP BY a.id,a.name,a.type ORDER BY a.code,a.id`, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var revenue, cogs, expenses int64
	for rows.Next() {
		var id, name, typ string
		var signed, debit int64
		if err := rows.Scan(&id, &name, &typ, &signed, &debit); err != nil {
			return r, err
		}
		amount := signed
		if typ == "expense" {
			amount = -signed
		}
		if typ == "revenue" {
			revenue += signed
		}
		if id == "ACC-COGS" {
			cogs += -signed
		}
		if typ == "expense" && id != "ACC-COGS" {
			expenses += -signed
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, Category: typ, AmountRial: amount, SecondaryAmountRial: debit})
	}
	if err := rows.Err(); err != nil {
		return r, err
	}
	return addSummaries(r,
		domain.ReportSummary{Key: "revenue", Label: "Revenue", AmountRial: revenue},
		domain.ReportSummary{Key: "cogs", Label: "COGS", AmountRial: cogs},
		domain.ReportSummary{Key: "gross_profit", Label: "Gross profit", AmountRial: revenue - cogs},
		domain.ReportSummary{Key: "expenses", Label: "Operating expenses", AmountRial: expenses},
		domain.ReportSummary{Key: "net_profit", Label: "Net profit / loss", AmountRial: revenue - cogs - expenses}), nil
}

func (s *Store) reportCashBank(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT fa.id,fa.name,fa.type,COALESCE(SUM(CASE WHEN je.posted_at<? THEN jl.debit_rial-jl.credit_rial ELSE 0 END),0),COALESCE(SUM(CASE WHEN je.posted_at>=? AND je.posted_at<? THEN jl.debit_rial ELSE 0 END),0),COALESCE(SUM(CASE WHEN je.posted_at>=? AND je.posted_at<? THEN jl.credit_rial ELSE 0 END),0) FROM financial_accounts fa JOIN accounts a ON a.id=fa.ledger_account_id LEFT JOIN journal_lines jl ON jl.account_id=a.id LEFT JOIN journal_entries je ON je.id=jl.journal_entry_id GROUP BY fa.id,fa.name,fa.type ORDER BY fa.type,fa.name,fa.id`, from, from, until, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var opening, inflows, outflows int64
	for rows.Next() {
		var id, name, typ string
		var open, in, out int64
		if err := rows.Scan(&id, &name, &typ, &open, &in, &out); err != nil {
			return r, err
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, Category: typ, AmountRial: open + in - out, SecondaryAmountRial: open, TertiaryAmountRial: in})
		opening += open
		inflows += in
		outflows += out
	}
	if err := rows.Err(); err != nil {
		return r, err
	}
	return addSummaries(r, domain.ReportSummary{Key: "opening", Label: "Opening balance", AmountRial: opening}, domain.ReportSummary{Key: "inflows", Label: "Inflows", AmountRial: inflows}, domain.ReportSummary{Key: "outflows", Label: "Outflows", AmountRial: outflows}, domain.ReportSummary{Key: "closing", Label: "Closing balance", AmountRial: opening + inflows - outflows}), nil
}

func (s *Store) reportReceivables(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT COALESCE(jl.party_id,''),COALESCE(c.name,'Unassigned customer'),COALESCE(SUM(jl.debit_rial-jl.credit_rial),0),COALESCE(SUM(CASE WHEN je.posted_at>=? AND je.posted_at<? THEN jl.debit_rial ELSE 0 END),0),COALESCE((SELECT SUM(pa.amount_rial) FROM payment_allocations pa JOIN payments p ON p.id=pa.payment_id WHERE pa.target_type IN ('invoice','order') AND pa.reversed=0 AND p.status='posted' AND p.customer_id=jl.party_id),0),COALESCE((SELECT SUM(l2.credit_rial-l2.debit_rial) FROM journal_lines l2 WHERE l2.account_id='ACC-CUSTOMER-CREDIT' AND l2.party_id=jl.party_id),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id LEFT JOIN customers c ON c.id=jl.party_id WHERE jl.account_id='ACC-AR' GROUP BY jl.party_id,c.name ORDER BY c.name,jl.party_id`, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var outstanding, charges, paid, credits int64
	for rows.Next() {
		var id, name string
		var balance, charge, payment, credit int64
		if err := rows.Scan(&id, &name, &balance, &charge, &payment, &credit); err != nil {
			return r, err
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, AmountRial: balance, SecondaryAmountRial: payment, TertiaryAmountRial: credit})
		outstanding += balance
		charges += charge
		paid += payment
		credits += credit
	}
	if err := rows.Err(); err != nil {
		return r, err
	}
	return addSummaries(r, domain.ReportSummary{Key: "receivable", Label: "Outstanding receivable", AmountRial: outstanding}, domain.ReportSummary{Key: "charges", Label: "Period receivable", AmountRial: charges}, domain.ReportSummary{Key: "allocated", Label: "Allocated payments", AmountRial: paid}, domain.ReportSummary{Key: "credit", Label: "Customer credit", AmountRial: credits}), nil
}

func (s *Store) reportPayables(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT COALESCE(jl.party_id,''),COALESCE(s.name,'Unassigned supplier'),COALESCE(SUM(jl.credit_rial-jl.debit_rial),0),COALESCE(SUM(CASE WHEN je.posted_at>=? AND je.posted_at<? THEN jl.credit_rial ELSE 0 END),0),COALESCE((SELECT SUM(pa.amount_rial) FROM payment_allocations pa JOIN payments p ON p.id=pa.payment_id WHERE pa.target_type='purchase' AND pa.reversed=0 AND p.status='posted' AND p.supplier_id=jl.party_id),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id LEFT JOIN suppliers s ON s.id=jl.party_id WHERE jl.account_id='ACC-AP' GROUP BY jl.party_id,s.name ORDER BY s.name,jl.party_id`, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var outstanding, charges, paid int64
	for rows.Next() {
		var id, name string
		var balance, charge, payment int64
		if err := rows.Scan(&id, &name, &balance, &charge, &payment); err != nil {
			return r, err
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, AmountRial: balance, SecondaryAmountRial: charge, TertiaryAmountRial: payment})
		outstanding += balance
		charges += charge
		paid += payment
	}
	if err := rows.Err(); err != nil {
		return r, err
	}
	return addSummaries(r, domain.ReportSummary{Key: "payable", Label: "Outstanding payable", AmountRial: outstanding}, domain.ReportSummary{Key: "charges", Label: "Period payable", AmountRial: charges}, domain.ReportSummary{Key: "allocated", Label: "Allocated payments", AmountRial: paid}), nil
}

func (s *Store) reportExpenses(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,a.name,COALESCE(SUM(jl.debit_rial-jl.credit_rial),0),COUNT(DISTINCT je.id) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id JOIN accounts a ON a.id=jl.account_id WHERE a.type='expense' AND a.id<>'ACC-COGS' AND je.posted_at>=? AND je.posted_at<? GROUP BY a.id,a.name ORDER BY a.name,a.id`, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var total int64
	for rows.Next() {
		var id, name string
		var amount, count int64
		if err := rows.Scan(&id, &name, &amount, &count); err != nil {
			return r, err
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, AmountRial: amount, SecondaryAmountRial: count})
		total += amount
	}
	return addSummaries(r, domain.ReportSummary{Key: "total", Label: "Operating expenses", AmountRial: total}), rows.Err()
}

func (s *Store) reportInventory(ctx context.Context, r domain.Report) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT m.id,m.name,m.consumption_unit,m.reorder_level_units,COALESCE(SUM(im.quantity_delta_units),0),COALESCE(SUM(im.total_cost_rial),0),COALESCE((SELECT SUM(ir.quantity_units) FROM inventory_reservations ir WHERE ir.material_id=m.id AND ir.status='active'),0) FROM materials m LEFT JOIN inventory_movements im ON im.material_id=m.id GROUP BY m.id,m.name,m.consumption_unit,m.reorder_level_units ORDER BY lower(m.name),m.id`)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var value int64
	for rows.Next() {
		var id, name, unit string
		var reorder, qty, cost, res int64
		if err := rows.Scan(&id, &name, &unit, &reorder, &qty, &cost, &res); err != nil {
			return r, err
		}
		avg := int64(0)
		if qty > 0 {
			avg = (cost*domain.QuantityScale + qty/2) / qty
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, Category: unit, AmountRial: cost, SecondaryAmountRial: avg, QuantityUnits: qty, SecondaryQuantityUnits: res, TertiaryAmountRial: qty - res})
		value += cost
	}
	return addSummaries(r, domain.ReportSummary{Key: "value", Label: "Inventory value", AmountRial: value}), rows.Err()
}

func (s *Store) reportSalesByService(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT COALESCE(ii.service_id,''),ii.description_snapshot,SUM(ii.quantity_units),SUM(ii.line_total_rial),COALESCE(SUM(oi.estimated_cost_rial),0),COALESCE(SUM(pc.actual_cost_rial),0),COUNT(DISTINCT i.id) FROM invoice_items ii JOIN invoices i ON i.id=ii.invoice_id AND i.status IN ('Posted','Partially Paid','Paid') LEFT JOIN (SELECT id AS order_item_id,estimated_cost_rial FROM order_items) oi ON oi.order_item_id=ii.order_item_id LEFT JOIN (SELECT pj.order_item_id,SUM(pc.material_cost_rial+pc.waste_cost_rial) actual_cost_rial FROM production_jobs pj JOIN production_consumptions pc ON pc.production_job_id=pj.id GROUP BY pj.order_item_id) pc ON pc.order_item_id=ii.order_item_id WHERE i.issue_date>=? AND i.issue_date<? GROUP BY ii.service_id,ii.description_snapshot ORDER BY ii.description_snapshot,ii.service_id`, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var revenue, estimatedTotal, actualTotal int64
	for rows.Next() {
		var id, name string
		var qty, sales, estimated, actual, count int64
		if err := rows.Scan(&id, &name, &qty, &sales, &estimated, &actual, &count); err != nil {
			return r, err
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, AmountRial: sales, SecondaryAmountRial: estimated, TertiaryAmountRial: actual, QuantityUnits: qty})
		revenue += sales
		estimatedTotal += estimated
		actualTotal += actual
	}
	return addSummaries(r, domain.ReportSummary{Key: "revenue", Label: "Invoiced revenue", AmountRial: revenue}, domain.ReportSummary{Key: "estimated_cost", Label: "Estimated cost", AmountRial: estimatedTotal}, domain.ReportSummary{Key: "actual_cost", Label: "Actual cost", AmountRial: actualTotal}, domain.ReportSummary{Key: "gross_profit", Label: "Gross profit", AmountRial: revenue - actualTotal}), rows.Err()
}

func (s *Store) reportCustomerSales(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT COALESCE(i.customer_id,''),i.customer_name_snapshot,SUM(i.total_rial),COUNT(*),COALESCE((SELECT SUM(pa.amount_rial) FROM payment_allocations pa JOIN payments p ON p.id=pa.payment_id WHERE pa.target_type='invoice' AND pa.target_id IN (SELECT id FROM invoices WHERE customer_id=i.customer_id) AND pa.reversed=0 AND p.status='posted'),0),COALESCE((SELECT SUM(jl.debit_rial-jl.credit_rial) FROM journal_lines jl WHERE jl.account_id='ACC-AR' AND jl.party_id=i.customer_id),0) FROM invoices i WHERE i.status IN ('Posted','Partially Paid','Paid') AND i.issue_date>=? AND i.issue_date<? GROUP BY i.customer_id,i.customer_name_snapshot ORDER BY i.customer_name_snapshot,i.customer_id`, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var total int64
	for rows.Next() {
		var id, name string
		var sales, count, paid, balance int64
		if err := rows.Scan(&id, &name, &sales, &count, &paid, &balance); err != nil {
			return r, err
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, AmountRial: sales, SecondaryAmountRial: paid, TertiaryAmountRial: balance})
		total += sales
	}
	return addSummaries(r, domain.ReportSummary{Key: "revenue", Label: "Customer sales", AmountRial: total}), rows.Err()
}

func (s *Store) reportMaterialUsage(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT COALESCE(pc.material_id,''),COALESCE(m.name,'Unknown material'),SUM(pc.consumed_quantity_units),SUM(pc.waste_quantity_units),SUM(pc.material_cost_rial),SUM(pc.waste_cost_rial) FROM production_consumptions pc LEFT JOIN materials m ON m.id=pc.material_id WHERE pc.created_at>=? AND pc.created_at<? GROUP BY pc.material_id,m.name ORDER BY m.name,pc.material_id`, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var cost int64
	for rows.Next() {
		var id, name string
		var consumed, waste, mat, wasteCost int64
		if err := rows.Scan(&id, &name, &consumed, &waste, &mat, &wasteCost); err != nil {
			return r, err
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, Name: name, QuantityUnits: consumed, SecondaryQuantityUnits: waste, AmountRial: mat, SecondaryAmountRial: wasteCost})
		cost += mat + wasteCost
	}
	return addSummaries(r, domain.ReportSummary{Key: "cost", Label: "Material and waste cost", AmountRial: cost}), rows.Err()
}

func (s *Store) reportProduction(ctx context.Context, r domain.Report, from, until string) (domain.Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,COALESCE(order_id,''),service_name_snapshot,status,estimated_cost_rial,COALESCE(actual_material_cost_rial+actual_waste_cost_rial+actual_outsourced_cost_rial,0),created_at FROM production_jobs WHERE created_at>=? AND created_at<? ORDER BY created_at,id`, from, until)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	var estimated, actual int64
	for rows.Next() {
		var id, order, service, status, date string
		var e, a int64
		if err := rows.Scan(&id, &order, &service, &status, &e, &a, &date); err != nil {
			return r, err
		}
		r.Rows = append(r.Rows, domain.ReportRow{ID: id, ReferenceID: order, Name: service, Status: status, Date: date, AmountRial: e, SecondaryAmountRial: a})
		estimated += e
		actual += a
	}
	return addSummaries(r, domain.ReportSummary{Key: "estimated", Label: "Estimated production cost", AmountRial: estimated}, domain.ReportSummary{Key: "actual", Label: "Actual production cost", AmountRial: actual}), rows.Err()
}

func addSummaries(r domain.Report, values ...domain.ReportSummary) domain.Report {
	r.Summaries = append(r.Summaries, values...)
	return r
}

func (s *Store) Dashboard(ctx context.Context, start, end time.Time) (domain.Dashboard, error) {
	from, until := reportWindow(start, end)
	d := domain.Dashboard{StartDate: start.Format("2006-01-02"), EndDate: end.Format("2006-01-02"), Attention: []domain.DashboardAttention{}, LowStock: []domain.DashboardLowStock{}, Production: []domain.DashboardProduction{}, RecentActivity: []domain.DashboardActivity{}}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_rial),0) FROM invoices WHERE status IN ('Posted','Partially Paid','Paid') AND issue_date>=? AND issue_date<?`, from, until).Scan(&d.RevenueRial); err != nil {
		return d, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(jl.debit_rial-jl.credit_rial),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE jl.account_id='ACC-CASH'`).Scan(&d.CashRial); err != nil {
		return d, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(jl.debit_rial-jl.credit_rial),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE jl.account_id='ACC-BANK'`).Scan(&d.BankRial); err != nil {
		return d, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(debit_rial-credit_rial),0) FROM journal_lines WHERE account_id='ACC-AR'`).Scan(&d.ReceivableRial); err != nil {
		return d, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(credit_rial-debit_rial),0) FROM journal_lines WHERE account_id='ACC-AP'`).Scan(&d.PayableRial); err != nil {
		return d, err
	}
	var revenue, cogs int64
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(credit_rial-debit_rial),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE jl.account_id='ACC-REVENUE' AND je.posted_at>=? AND je.posted_at<?`, from, until).Scan(&revenue); err != nil {
		return d, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(debit_rial-credit_rial),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE jl.account_id='ACC-COGS' AND je.posted_at>=? AND je.posted_at<?`, from, until).Scan(&cogs); err != nil {
		return d, err
	}
	d.GrossProfitRial = revenue - cogs
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoices WHERE status IN ('Posted','Partially Paid')`).Scan(&d.OpenInvoiceCount)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE promised_at>=? AND promised_at<? AND commercial_status<>'Cancelled'`, from, until).Scan(&d.DueOrderCount)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE promised_at<? AND promised_at IS NOT NULL AND fulfillment_status<>'Delivered' AND commercial_status<>'Cancelled'`, from).Scan(&d.OverdueOrderCount)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE fulfillment_status='In Production'`).Scan(&d.InProductionCount)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE fulfillment_status='Ready'`).Scan(&d.ReadyOrderCount)
	attentionRows, err := s.db.QueryContext(ctx, `SELECT 'Overdue invoice',invoice_number,COALESCE(due_date,''),remaining FROM (SELECT i.invoice_number,i.due_date,i.total_rial-COALESCE((SELECT SUM(pa.amount_rial) FROM payment_allocations pa JOIN payments p ON p.id=pa.payment_id WHERE pa.target_type='invoice' AND pa.target_id=i.id AND pa.reversed=0 AND p.status='posted'),0) remaining FROM invoices i WHERE i.status IN ('Posted','Partially Paid') AND i.due_date IS NOT NULL) WHERE due_date<? AND remaining>0 ORDER BY due_date,invoice_number LIMIT 10`, from)
	if err != nil {
		return d, err
	}
	for attentionRows.Next() {
		var a domain.DashboardAttention
		if err := attentionRows.Scan(&a.Label, &a.Detail, &a.Date, &a.AmountRial); err != nil {
			attentionRows.Close()
			return d, err
		}
		a.Kind = "invoice"
		d.Attention = append(d.Attention, a)
	}
	attentionRows.Close()
	checkRows, err := s.db.QueryContext(ctx, `SELECT CASE WHEN due_date<? THEN 'Overdue check' ELSE 'Check due soon' END,check_number,due_date,amount_rial,direction FROM checks WHERE status IN ('Received','Deposited','Issued','Delivered') AND due_date<? ORDER BY due_date,check_number LIMIT 10`, from, until)
	if err != nil {
		return d, err
	}
	for checkRows.Next() {
		var a domain.DashboardAttention
		if err := checkRows.Scan(&a.Label, &a.Detail, &a.Date, &a.AmountRial, &a.Kind); err != nil {
			checkRows.Close()
			return d, err
		}
		d.Attention = append(d.Attention, a)
	}
	checkRows.Close()
	nowText := time.Now().UTC().Format(time.RFC3339Nano)
	loanRows, err := s.db.QueryContext(ctx, `SELECT 'Overdue loan installment',loans.loan_number,loan_installments.due_date,loan_installments.total_due_rial-COALESCE(paid.amount_rial,0),'loan' FROM loan_installments JOIN loans ON loans.id=loan_installments.loan_id LEFT JOIN (SELECT lpa.installment_id,SUM(lpa.principal_rial+lpa.interest_rial) amount_rial FROM loan_payment_allocations lpa JOIN loan_payments lp ON lp.id=lpa.payment_id AND lp.status='posted' GROUP BY lpa.installment_id) paid ON paid.installment_id=loan_installments.id WHERE loans.status='Active' AND loan_installments.due_date<? AND loan_installments.total_due_rial>COALESCE(paid.amount_rial,0) ORDER BY loan_installments.due_date,loans.loan_number LIMIT 10`, nowText)
	if err != nil {
		return d, err
	}
	for loanRows.Next() {
		var a domain.DashboardAttention
		if err := loanRows.Scan(&a.Label, &a.Detail, &a.Date, &a.AmountRial, &a.Kind); err != nil {
			loanRows.Close()
			return d, err
		}
		d.Attention = append(d.Attention, a)
	}
	loanRows.Close()
	rows, err := s.db.QueryContext(ctx, `SELECT m.id,m.name,m.consumption_unit,m.reorder_level_units,COALESCE(SUM(im.quantity_delta_units),0),COALESCE(SUM(im.total_cost_rial),0),COALESCE((SELECT SUM(quantity_units) FROM inventory_reservations WHERE material_id=m.id AND status='active'),0) FROM materials m LEFT JOIN inventory_movements im ON im.material_id=m.id GROUP BY m.id,m.name,m.consumption_unit,m.reorder_level_units HAVING COALESCE(SUM(im.quantity_delta_units),0)-COALESCE((SELECT SUM(quantity_units) FROM inventory_reservations WHERE material_id=m.id AND status='active'),0)<=m.reorder_level_units ORDER BY m.name,m.id LIMIT 10`)
	if err != nil {
		return d, err
	}
	for rows.Next() {
		var id, name, unit string
		var reorder, qty, cost, res int64
		if err := rows.Scan(&id, &name, &unit, &reorder, &qty, &cost, &res); err != nil {
			rows.Close()
			return d, err
		}
		avg := int64(0)
		if qty > 0 {
			avg = (cost*domain.QuantityScale + qty/2) / qty
		}
		d.LowStock = append(d.LowStock, domain.DashboardLowStock{ID: id, Name: name, Unit: unit, AvailableUnits: qty - res, ReorderLevelUnits: reorder, AverageCostRial: avg, ValueRial: cost})
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT production_jobs.id,COALESCE(production_jobs.order_id,''),COALESCE(orders.order_number,''),COALESCE(orders.customer_name_snapshot,''),production_jobs.service_name_snapshot,production_jobs.status,COALESCE(production_jobs.planned_at,'') FROM production_jobs LEFT JOIN orders ON orders.id=production_jobs.order_id WHERE production_jobs.status NOT IN ('Completed','Cancelled') ORDER BY production_jobs.planned_at,production_jobs.id LIMIT 10`)
	if err != nil {
		return d, err
	}
	for rows.Next() {
		var j domain.DashboardProduction
		if err := rows.Scan(&j.ID, &j.OrderID, &j.OrderNumber, &j.Customer, &j.Service, &j.Status, &j.DueDate); err != nil {
			rows.Close()
			return d, err
		}
		d.Production = append(d.Production, j)
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT p.id,p.posted_at,CASE WHEN p.direction='incoming' THEN 'Payment received' ELSE 'Payment sent' END,COALESCE(p.reference,''),p.direction,p.amount_rial FROM payments p ORDER BY p.posted_at DESC,p.id DESC LIMIT 10`)
	if err != nil {
		return d, err
	}
	for rows.Next() {
		var a domain.DashboardActivity
		if err := rows.Scan(&a.ID, &a.Date, &a.Label, &a.Detail, &a.Direction, &a.AmountRial); err != nil {
			rows.Close()
			return d, err
		}
		d.RecentActivity = append(d.RecentActivity, a)
	}
	rows.Close()
	return d, nil
}

func (s *Store) PrintDocument(ctx context.Context, kind, id, start, end, partyID string) (domain.PrintDocument, error) {
	shop, err := s.GetShopSettings(ctx)
	if err != nil {
		return domain.PrintDocument{}, err
	}
	doc := domain.PrintDocument{Kind: kind, Shop: shop, Lines: []domain.PrintLine{}, StatementLines: []domain.StatementLine{}, Allocations: []domain.PrintAllocation{}}
	switch kind {
	case "quote":
		q, e := s.GetQuote(ctx, id)
		if e != nil {
			return doc, e
		}
		doc.Number = q.QuoteNumber
		doc.Date = q.CreatedAt.Format(time.RFC3339)
		if q.ExpiryDate != nil {
			doc.DueDate = q.ExpiryDate.Format(time.RFC3339)
		}
		doc.Status = string(q.Status)
		doc.CustomerName = q.CustomerNameSnapshot
		doc.CustomerContact = q.CustomerPhoneSnapshot
		doc.SubtotalRial = q.SubtotalRial
		doc.DiscountRial = q.DiscountRial
		doc.TotalRial = q.TotalRial
		doc.Notes = q.Notes
		for _, i := range q.Items {
			doc.Lines = append(doc.Lines, domain.PrintLine{Description: i.ServiceNameSnapshot, QuantityUnits: int64(i.Quantity), Unit: i.QuantityUnit, UnitPriceRial: i.SellingPriceRial, LineTotalRial: i.SellingPriceRial})
		}
		return doc, nil
	case "invoice":
		i, e := s.GetInvoice(ctx, id)
		if e != nil {
			return doc, e
		}
		doc.Number = i.InvoiceNumber
		doc.Date = i.IssueDate.Format(time.RFC3339)
		if i.DueDate != nil {
			doc.DueDate = i.DueDate.Format(time.RFC3339)
		}
		doc.Status = i.Status
		doc.CustomerName = i.CustomerNameSnapshot
		doc.CustomerContact = i.CustomerPhoneSnapshot
		doc.Reference = i.OrderID
		doc.SubtotalRial = i.SubtotalRial
		doc.DiscountRial = i.DiscountRial
		doc.TotalRial = i.TotalRial
		doc.PaidRial = i.PaidRial
		doc.RemainingRial = i.RemainingRial
		doc.PaymentStatus = i.Status
		doc.Notes = i.Notes
		for _, x := range i.Items {
			doc.Lines = append(doc.Lines, domain.PrintLine{Description: x.DescriptionSnapshot, QuantityUnits: x.QuantityUnits, Unit: x.QuantityUnit, UnitPriceRial: x.UnitPriceRial, LineTotalRial: x.LineTotalRial})
		}
		return doc, nil
	case "payment_receipt":
		p, e := s.GetPayment(ctx, id)
		if e != nil {
			return doc, e
		}
		doc.Number = p.PaymentNumber
		doc.Date = p.PostedAt.Format(time.RFC3339)
		doc.Status = string(p.Status)
		doc.AmountRial = p.AmountRial
		doc.Method = string(p.Method)
		doc.Reference = p.Reference
		doc.Notes = p.Notes
		for _, a := range p.Allocations {
			doc.Allocations = append(doc.Allocations, domain.PrintAllocation{TargetType: a.TargetType, Reference: a.TargetID, AmountRial: a.AmountRial})
		}
		var account string
		_ = s.db.QueryRowContext(ctx, `SELECT name FROM financial_accounts WHERE id=?`, p.FinancialAccountID).Scan(&account)
		doc.AccountName = account
		_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(name,'') FROM customers WHERE id=?`, p.CustomerID).Scan(&doc.CustomerName)
		_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(name,'') FROM suppliers WHERE id=?`, p.SupplierID).Scan(&doc.SupplierName)
		return doc, nil
	case "customer_statement":
		return s.statement(ctx, doc, "ACC-AR", partyID, start, end)
	case "supplier_statement":
		return s.statement(ctx, doc, "ACC-AP", partyID, start, end)
	default:
		return doc, fmt.Errorf("unsupported print document kind %q", kind)
	}
}

func (s *Store) statement(ctx context.Context, doc domain.PrintDocument, account, party, start, end string) (domain.PrintDocument, error) {
	if party == "" {
		return doc, errors.New("party is required")
	}
	doc.Number = "Statement"
	if end != "" {
		doc.Date = end + "T00:00:00Z"
	} else {
		doc.Date = time.Now().UTC().Format(time.RFC3339)
	}
	from := start
	until := end
	if until != "" {
		if t, e := time.Parse("2006-01-02", until); e == nil {
			until = t.Add(24 * time.Hour).Format(time.RFC3339)
		}
	}
	if from != "" {
		if t, e := time.Parse("2006-01-02", from); e == nil {
			from = t.Format(time.RFC3339)
		}
	}
	var opening int64
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(debit_rial-credit_rial),0) FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE jl.account_id=? AND jl.party_id=? AND (?='' OR je.posted_at<?)`, account, party, start, start).Scan(&opening); err != nil {
		return doc, err
	}
	doc.StatementLines = append(doc.StatementLines, domain.StatementLine{Description: "Opening balance", BalanceRial: opening})
	rows, err := s.db.QueryContext(ctx, `SELECT je.posted_at,je.entry_number,je.description,jl.debit_rial,jl.credit_rial FROM journal_lines jl JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE jl.account_id=? AND jl.party_id=? AND (?='' OR je.posted_at>=?) AND (?='' OR je.posted_at<?) ORDER BY je.posted_at,je.entry_number,jl.position`, account, party, start, from, end, until)
	if err != nil {
		return doc, err
	}
	defer rows.Close()
	balance := opening
	for rows.Next() {
		var date, ref, desc string
		var debit, credit int64
		if err := rows.Scan(&date, &ref, &desc, &debit, &credit); err != nil {
			return doc, err
		}
		balance += debit - credit
		doc.StatementLines = append(doc.StatementLines, domain.StatementLine{Date: date, Reference: ref, Description: desc, DebitRial: debit, CreditRial: credit, BalanceRial: balance})
	}
	doc.TotalRial = balance
	if account == "ACC-AR" {
		_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(name,'') FROM customers WHERE id=?`, party).Scan(&doc.CustomerName)
	} else {
		_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(name,'') FROM suppliers WHERE id=?`, party).Scan(&doc.SupplierName)
	}
	return doc, rows.Err()
}
