# Product Definition

## Purpose

Atropaten manages the day-to-day operation and finances of a single physical print shop. It should replace disconnected spreadsheets, notebooks, calculators, and ad-hoc accounting with one fast local desktop workflow.

## Core flow

`Materials -> Services -> Orders -> Production -> Payments -> Accounting`

Purchasing feeds inventory. Inventory and production feed actual cost. Sales and payments feed accounting. Reporting is derived from those records rather than maintained separately.

## Users

The initial product targets a small shop where owners and employees may perform several roles. Role/permission support should therefore remain possible, but the application must not require enterprise-style administration for normal use.

## Functional areas

### Catalog

- Materials and consumables
- Services and service variants
- Dynamic service parameters
- Pricing rules
- Machines and usage rates

### Inventory and purchasing

- Purchase units and consumption units with conversion factors
- Weighted-average material costing by default
- Immutable stock movement ledger
- Physical, reserved, and available quantities
- Purchases and supplier balances
- Waste, adjustments, returns, and stocktakes
- Optional inventory locations

### Sales

- Customers
- Quotes
- Orders containing multiple unrelated items
- Invoices and receipts
- Deposits and customer credit
- Partial and multi-method payments
- Discounts and manual price overrides

### Production

- Production jobs per order item
- Estimated versus actual consumption
- Waste tracking
- Machine and employee assignment
- Outsourced production
- Artwork/file management
- Proof approval

### Finance

- Double-entry accounting
- Cash and bank accounts
- Accounts receivable/payable
- Expenses
- Incoming and outgoing checks
- Loans and installments
- Transfers
- Automatic journal generation for routine operations

### Ownership

Support multiple owners/partners with independent:

- Ownership percentage
- Profit-sharing percentage
- Capital contributions
- Drawings
- Owner loans
- Business expenses paid personally
- Personal expenses paid by the business
- Profit/loss allocations
- Capital balances and statements

Business operating profit must be calculated before owner allocation.

### Reporting

- Daily sales and cash flow
- Revenue, expenses, gross profit, and net profit
- Receivables and payables
- Cash and bank balances
- Upcoming/overdue checks and installments
- Inventory valuation and low stock
- Material consumption and waste
- Profitability by service/order/customer
- Owner capital and allocation statements

## Important invariants

1. Historical cost and price snapshots must not change when current catalog costs or pricing rules change.
2. Inventory corrections create movements; they do not rewrite history.
3. Accounting corrections use reversing/correcting entries rather than destructive edits after posting.
4. Routine business actions generate balanced journal entries automatically.
5. Order commercial, fulfillment, and payment states remain separate.
6. Production reservations do not reduce accounting inventory until consumption occurs.
7. Large artwork and attachments live on disk; the database stores managed metadata and references.
8. Derived balances should be calculated from authoritative transactions rather than independently editable totals.

## Non-goals for the initial product

- SaaS/multi-tenant operation
- Multiple independent companies in one installation
- E-commerce storefront
- Cloud synchronization
- Full payroll/HR suite
- Enterprise manufacturing/MRP
- Mandatory lot/serial tracking

These may be reconsidered only when a concrete business requirement appears.
