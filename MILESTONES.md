# Milestones

Milestones describe product outcomes, not arbitrary implementation batches.

## M0 - Desktop foundation

**Goal:** establish Atropaten's visual language and application navigation.

Deliverables:

- Application shell
- Sidebar/top bar/workspace
- Design tokens
- Reusable table, inspector, tabs, forms, badges, cards, and toolbar patterns
- Dashboard mock
- Orders list mock
- Order workspace mock
- Service configurator mock with live frontend-only sample calculation
- Materials table + inspector mock

Acceptance:

- Runs through `wails dev`
- Major screens are reachable from the shell
- Layout works at representative Windows desktop/laptop sizes
- Starter Wails UI/assets are removed
- UI establishes a coherent reusable component language

## M1 - Catalog and pricing engine

**Goal:** persist the shop's sellable services and cost inputs and calculate prices through Go.

Deliverables:

- SQLite/migrations foundation
- Money/quantity representation
- Materials and units
- Services and parameters
- Material/machine/labor/overhead/waste cost components
- Pricing rules
- Machines/rates
- Go pricing engine
- Service configurator connected to backend

Acceptance:

- Configurable examples such as business cards, per-page A3 printing, and time-only design work can be represented without service-specific backend code
- Pricing result explains its cost components
- Below-cost selling price can be detected
- Pricing calculations have deterministic Go tests

## M2 - Customers, quotes, and orders

**Goal:** create and preserve real sales work.

Deliverables:

- Customers
- Quotes
- Orders
- Multiple order items
- Historical configuration/cost/price snapshots
- Discounts and manual overrides
- Attachments/proof metadata

Acceptance:

- An order can contain unrelated services
- Reopening an old order shows its original calculation even after catalog changes
- Order commercial/fulfillment/payment states are not collapsed into one enum

## M3 - Purchasing, inventory, and production

**Goal:** connect physical material flow to order execution and actual cost.

Deliverables:

- Suppliers and purchases
- Inventory movement ledger
- Weighted-average costing
- Reservations
- Adjustments/stocktake
- Production jobs
- Actual consumption/waste
- Outsourcing
- Production queue

Acceptance:

- Inventory can be reconstructed from movements
- Reservation changes availability without posting consumption
- Production completion records actual material use/waste
- Purchase and production costing tests cover representative conversions and corrections

## M4 - Accounting and payments

**Goal:** make operational activity financially authoritative.

Deliverables:

- Chart of accounts
- Journal engine
- Invoices
- Cash/bank
- Payments and allocations
- Receivables/payables
- Expenses/transfers
- Purchase, inventory, sales, and COGS postings
- Reversal/correction support

Acceptance:

- Every posted journal entry balances
- Partial payments and customer credit work
- One payment can allocate across invoices
- Customer/supplier balances derive from authoritative transactions
- Routine operations require no manual journal entry

## M5 - Checks, loans, and owners

**Goal:** cover the shop's non-trivial treasury and partnership accounting.

Deliverables:

- Check lifecycle
- Loans/installments
- Owners and shares
- Capital/drawings
- Owner loans/reimbursements
- Fiscal periods
- Profit allocation/closing

Acceptance:

- Due/overdue obligations are visible
- Owner withdrawals cannot be classified as ordinary expense by the normal workflow
- Ownership and profit-sharing percentages can differ
- Closed-period profit allocation remains historically stable

## M6 - Reports and release

**Goal:** produce a complete, recoverable Windows application for daily use.

Deliverables:

- Real dashboard
- Financial/operational reports
- Printable quotes/invoices/receipts/statements
- Backup/restore
- Windows installer
- Upgrade/migration path
- Release smoke tests

Acceptance:

- Core reports reconcile to underlying transactions
- Backup followed by restore reproduces database and managed attachments
- Clean Windows installation works without development tooling
- Representative print-shop workflow succeeds end-to-end
