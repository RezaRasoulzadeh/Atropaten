# Roadmap

The roadmap is ordered to establish usable vertical slices and validate the domain before committing to unnecessary complexity.

## Phase 0 - Foundation and UI language

- Establish project/product documentation
- Replace Wails starter UI with Atropaten application shell
- Define design tokens and reusable desktop components
- Implement sidebar, top bar, workspace layout, tables, inspectors, tabs, status badges, and form controls
- Build dashboard with representative local mock data
- Build order workspace and service configurator prototypes
- Validate desktop density and Windows-oriented interaction patterns

Exit condition: the major workflows can be evaluated visually and navigationally without relying on a final database schema.

## Phase 1 - Core persistence and catalog

- Select SQLite driver and migration approach
- Define money/decimal representation
- Establish migrations and transactional storage infrastructure
- Shop/settings foundation
- Party/customer/supplier foundation
- Materials, units, and conversion factors
- Services, dynamic parameters, cost components, and pricing rules
- Machines and usage rates

Exit condition: catalog data is persisted and the pricing engine can calculate explainable service estimates from real stored definitions.

## Phase 2 - Sales and pricing

- Quotes and quote items
- Orders and independent order items
- Pricing/cost snapshots
- Dynamic service configurator connected to Go pricing engine
- Discounts/manual selling-price override
- Customer order history
- Attachments and proof metadata

Exit condition: a real customer order can be configured, priced, saved, reopened, and remain historically stable after catalog changes.

## Phase 3 - Inventory and purchasing

- Supplier purchases
- Inventory movement ledger
- Weighted-average costing
- Reservations
- Stock adjustments and stocktake
- Waste and returns
- Low-stock warnings
- Optional location field from the beginning

Exit condition: purchasing and production demand produce auditable inventory quantities and costs.

## Phase 4 - Production

- Production jobs per order item
- Ready/in-progress/completed workflow
- Machine/employee assignment
- Estimated versus actual consumption
- Waste recording
- Outsourced production
- Proof approval flow
- Promised-date/late-order monitoring

Exit condition: an order can flow from confirmation through production and delivery with actual cost evidence.

## Phase 5 - Accounting and payments

- Chart of accounts
- Journal/posting engine
- Cash and bank accounts
- Invoices
- Payments and payment allocation
- Receivables/payables
- Purchase accounting
- Inventory/COGS accounting integration
- Expenses and transfers
- Reversal/correction operations

Exit condition: routine operational transactions automatically produce balanced, auditable accounting records and customer/supplier balances.

## Phase 6 - Treasury and ownership

- Incoming/outgoing checks and lifecycle
- Loans and installment schedules
- Multiple owners
- Capital contributions and drawings
- Owner loans/reimbursements
- Fiscal periods
- Profit/loss allocation and closing
- Owner statements

Exit condition: treasury obligations and partner equity can be managed without mixing owner activity with business revenue/expense.

## Phase 7 - Reporting, printing, and operations

- Dashboard backed by real data
- Operational and financial reports
- Quote/invoice/receipt printing
- Customer/supplier statements
- Owner statements
- Export where useful
- Backup/restore including managed attachments
- Application data-directory management
- Settings and numbering configuration

Exit condition: the application covers the daily operational loop and can be safely backed up/restored.

## Phase 8 - Windows release hardening

- Windows-native testing
- Installer/package configuration
- Printing validation against real shop printers
- Filesystem/path validation
- Performance on realistic datasets
- Upgrade/migration testing
- Backup/restore recovery testing
- Release build and smoke-test checklist

Exit condition: a clean Windows machine can install, operate, upgrade, back up, and restore Atropaten reliably.
