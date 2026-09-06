# M6-001 — Real dashboard, reconciled reports, and printable documents

## Objective

Turn Atropaten's accumulated operational and financial data into a trustworthy daily dashboard, reconciled reports, and printable customer/supplier documents without duplicating business logic in the frontend.

This task does not perform final packaging/release hardening; that follows in M6-002.

## Core rules

- Reports derive from authoritative persisted operational/accounting data, never duplicated mutable summary columns where a source ledger exists.
- Financial reports must reconcile to journal entries.
- Inventory reports must reconcile to inventory movements.
- Production reports must derive from production jobs/consumption/waste.
- Historical financial and order snapshots remain immutable.
- All authoritative money remains integer Rial.
- User-facing money uses the shared grouped Rial/Toman utilities.
- User-facing dates use the shared Jalali utilities.
- Report calculations belong in Go application/domain/query services; Vue/Wails remain thin.
- Use deterministic ordering and integer arithmetic.

## Scope

### 1. Reporting/query service

Add a dedicated read/query layer for dashboard/reporting needs. It may use optimized SQL, but must not bypass domain/accounting meaning.

Support explicit date ranges and sensible defaults.

Avoid loading entire tables into Vue and calculating business totals client-side.

### 2. Real dashboard

Replace remaining mock/sample dashboard data with persisted data.

Dashboard should prioritize attention and daily operations rather than decorative charts.

Include at minimum:
- today's / selected-period sales revenue from posted invoices
- cash balance
- bank balance
- total receivables
- total payables
- outstanding customer balances
- overdue checks
- checks due soon
- overdue loan installments
- low-stock materials
- production jobs requiring attention
- orders due/overdue / in production / ready where current data supports it
- recent payments/financial activity

Quick actions should navigate to real workflows.

No fake/sample financial numbers may remain in production dashboard code.

### 3. Financial reports

Add a Reports workspace with a practical report selector/date filters.

Implement at minimum:

#### Profit & Loss
- Revenue
- COGS
- Gross profit
- Operating expenses
- Other income/expense where applicable
- Net profit/loss

Must derive strictly from journal lines/account classifications and reconcile to fiscal-period preview logic.

#### Cash / Bank
- opening balance for range
- inflows
- outflows
- closing balance
- breakdown by financial account

Must reconcile to account ledger activity.

#### Accounts Receivable
- customer
- invoiced/receivable
- allocated payments
- customer credit
- outstanding
- aging buckets if practical from current invoice dates/due dates

#### Accounts Payable
- supplier
- purchases/payable
- allocated payments
- advances/credits if represented
- outstanding

#### Expenses
- by expense account/category
- totals by period
- transaction drilldown

#### Inventory valuation
- material
- physical quantity
- reserved quantity
- available quantity
- average cost
- inventory value

Must reconcile to inventory movement-derived stock and weighted-average values.

### 4. Operational reports

Implement useful report views for:

#### Sales by service
- quantity/jobs/orders where meaningful
- invoiced revenue
- estimated cost
- actual cost where available
- gross profit / margin

Use immutable order/invoice snapshots, not current catalog pricing.

#### Customer sales
- customer revenue
- outstanding receivable
- payment totals
- order/invoice count

#### Material consumption and waste
- material
- consumed quantity
- waste quantity
- actual material cost
- selected period

Derive from immutable production/inventory movements.

#### Production performance
- jobs completed
- jobs pending/in progress/failed
- estimated vs actual cost
- outsourced cost
- due/late indicators where supported

### 5. Drilldown and reconciliation

Reports must provide useful drilldown to source rows where practical.

Examples:
- P&L account total -> journal lines/entries
- AR customer -> invoices/payments
- AP supplier -> purchases/payments
- inventory value -> material/movement history
- material waste -> production jobs/consumption records

Do not introduce report-only balances that can drift from authoritative data.

### 6. Printable document model

Create a reusable print-document data layer separate from Vue screen state.

At minimum support:
- Quote
- Invoice
- Payment receipt
- Customer statement
- Supplier statement

Documents must use saved historical snapshots/transactions.

Do not reprice or recalculate old quotes/orders/invoices from current catalog definitions.

### 7. Shop/company settings required for documents

Persist minimal shop identity/settings if not already present:
- shop name
- optional legal/display subtitle
- phone
- address
- optional email
- optional website/social
- optional registration/tax identifiers as free display fields
- logo file path/managed attachment reference if practical
- default document footer/notes

Do not build a large settings system yet.

### 8. Printable Quote

Include at minimum:
- quote number/date/expiry
- customer identity/contact
- line items using immutable quote snapshots
- quantities
- prices
- discount
- total
- notes
- shop identity

### 9. Printable Invoice

Include at minimum:
- invoice number/date/status
- customer identity/contact
- source order reference
- immutable invoice items
- subtotal/discount/total
- paid amount
- remaining amount
- payment status
- notes/shop identity

### 10. Payment receipt

Include:
- receipt/payment number
- date
- customer/supplier as applicable
- amount
- method
- financial account label where appropriate
- allocations/references
- notes
- reversal state clearly visible if reversed

### 11. Customer statement

For a selected range:
- opening balance where deterministically available
- invoices/receivable charges
- payments
- credits/reversals
- running balance
- closing balance

Must reconcile to authoritative customer financial summary.

### 12. Supplier statement

For a selected range:
- opening payable where deterministically available
- purchases/payable charges
- supplier payments
- reversals/credits where represented
- running balance
- closing balance

Must reconcile to authoritative supplier payable balance.

### 13. Print/export behavior

Provide a usable print-preview flow suitable for Windows desktop.

Preferred implementation:
- dedicated clean print layout/component
- browser/WebView print mechanism for printing/PDF where reliable
- print CSS that removes application chrome
- A4-friendly layout
- deterministic pagination for normal document sizes

Do not add a heavy external PDF generation stack unless necessary.

Printed content must remain readable in grayscale and should not rely on color alone.

### 14. Reports workspace UI

Add/complete navigation for Reports.

Suggested sections:
- Overview
- Profit & Loss
- Cash / Bank
- Receivables
- Payables
- Expenses
- Inventory
- Sales
- Production

Requirements:
- date-range filters
- compact summary cards
- dense tables
- sorting where useful
- source drilldown
- printable/exportable presentation where practical
- shared sticky workspace geometry
- no clipped controls/tables/inspectors

### 15. Dashboard/report performance

Avoid obvious N+1 query patterns for lists/report periods.

Representative shop-sized databases should load dashboard/report summaries promptly.

Use indexes in a migration only when justified by report/query paths.

If schema/index changes are needed, add a safe v12 -> v13 migration and preservation tests. If no schema changes are required, do not create an artificial migration.

### 16. Tests

Add focused Go tests for at least:
- dashboard numbers from persisted data
- P&L reconciliation to journal lines
- cash/bank opening + movement + closing reconciliation
- AR totals reconcile to invoices/payments/credits
- AP totals reconcile to purchases/payments
- inventory valuation reconciles to movement-derived stock
- sales report uses historical snapshots rather than current catalog pricing
- material consumption/waste reporting
- production estimated-vs-actual reporting
- date-range boundaries
- reversals excluded/included correctly according to net financial effect
- customer statement running/closing balance
- supplier statement running/closing balance
- printable quote/invoice data uses immutable snapshots
- receipt reversal state
- migration preservation if a migration is introduced

Existing tests must continue to pass.

## Non-goals

Do not implement yet:
- final backup/restore workflow
- Windows installer/release packaging
- auto-update
- release signing
- final end-to-end deep review
- major UI redesign
- tax filing/report submission
- payroll
- external cloud sync
- analytics warehouse

## Acceptance criteria

- Dashboard contains real persisted operational/financial data only.
- P&L reconciles to journal entries.
- Cash/Bank report reconciles to financial-account ledger balances.
- AR/AP reports reconcile to authoritative transactions.
- Inventory valuation reconciles to inventory movements/current derived valuation.
- Sales/production/material reports use historical persisted facts.
- Quote, Invoice, Payment Receipt, Customer Statement, and Supplier Statement have usable print previews.
- Printable documents do not reprice historical data.
- Reports and print views follow Jalali/grouped-money conventions.
- Existing M1-M5 workflows remain intact.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.

## Manual verification deferred to final integration pass

The final project review will deeply validate report reconciliation, business flows, printing, UI geometry, and accounting correctness before release. For this task, perform implementation-level checks and automated validation only.
