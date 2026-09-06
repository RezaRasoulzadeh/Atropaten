# Domain Model

This document defines conceptual boundaries and invariants. It is intentionally not a final database schema.

## Domain map

```text
Shop
├── Catalog
│   ├── Materials
│   ├── Services
│   ├── Pricing Rules
│   └── Machines
├── Inventory
│   ├── Stock
│   ├── Purchases
│   ├── Consumption
│   └── Adjustments / Waste
├── Sales
│   ├── Customers
│   ├── Quotes
│   ├── Orders
│   ├── Production
│   ├── Invoices
│   └── Payments
├── Suppliers
│   ├── Suppliers
│   ├── Purchases
│   └── Payables
├── Finance
│   ├── Accounts
│   ├── Transactions
│   ├── Checks
│   ├── Loans
│   ├── Expenses
│   └── Journal
└── Management
    ├── Owners
    ├── Dashboard
    ├── Reports
    ├── Backup / Restore
    └── Settings
```

## Core conceptual entities

```text
Shop
Owner
OwnershipShare
FiscalPeriod

Party
├── Customer
├── Supplier
└── OwnerReference

Material
MaterialUnit
MaterialPurchase
InventoryMovement
InventoryReservation

Machine
MachineUsageRate
MachineMaintenance

Service
ServiceVariant
ServiceParameter
ServicePriceRule
ServiceMaterialRequirement
ServiceMachineRequirement
ServiceLaborRequirement

Quote
QuoteItem

Order
OrderItem
ProductionJob
ProductionConsumption
WasteRecord
Attachment
ProofApproval

Invoice
InvoiceLine
Payment
PaymentAllocation
Expense

Account
JournalEntry
JournalLine

CashAccount
BankAccount
Check
Loan
LoanInstallment

OwnerCapitalTransaction
OwnerDrawing
OwnerLoan
ProfitAllocation
```

## Deletion, archive, and historical records

Atropaten must distinguish **Archive** from **Delete**. They are not aliases.

For ordinary user-managed master data such as customers, suppliers, materials, services, machines, reusable pricing definitions, and similar records:

- **Archive** is reversible and only removes the record from normal active workflows.
- **Delete** means a real hard delete from persistent storage. It must not silently turn into archive/soft-delete behavior.
- Every major CRUD surface should expose a proper Delete action where deletion is semantically safe.
- Deleting an unreferenced record should purge it and its purely-owned child configuration transactionally instead of leaving dead rows indefinitely.
- If another authoritative record still references it, deletion must either cascade only through data that has no independent historical meaning, or be rejected with a clear explanation of what prevents deletion.
- Never silently orphan references, silently null important links, or silently archive when the user requested Delete.

Historical/posted business records follow stronger rules:

- Persisted order/quote snapshots, posted inventory movements, journal entries, payments, invoices, production consumption, and other authoritative history must not be hard-deleted once they have business/accounting meaning.
- Draft/unposted transactional records may be hard-deleted when no protected downstream record depends on them.
- Posted/committed history is corrected through cancellation, reversal, compensating records, or explicit void workflows rather than destructive deletion.

The intent is to avoid database clutter from disposable master/configuration records while preserving auditability of authoritative history.

## Service and pricing model

A service is a configurable calculation model, not merely a fixed-price catalog item.

```text
Service
├── Parameters
├── Cost Components
├── Pricing Rules
└── Output Rules
```

Parameter types may include integer, decimal, boolean, choice, and material reference.

Cost components may include material consumption, machine usage, labor/time, outsourced service, fixed cost, percentage overhead, waste, and manual/custom cost.

Pricing methods may include fixed price, per-unit price, quantity tiers, cost plus markup percentage, cost plus fixed margin, formula-based pricing, and manual pricing.

Calculation pipeline:

```text
Input parameters
-> Validate
-> Calculate production quantities
-> Resolve material consumption
-> Resolve machine/labor costs
-> Apply waste
-> Calculate base cost
-> Apply overhead
-> Calculate suggested price
-> Apply discount/manual override
-> Snapshot result
```

Every order item stores the resolved configuration and cost/price snapshot used when it was sold.

## Inventory model

Materials distinguish purchase units from consumption units. Example: paper may be purchased by pack and consumed by sheet.

Default valuation is weighted average cost.

Inventory movements are immutable and include purchase, production consumption, waste, supplier return, customer return where applicable, adjustment, transfer, and opening balance.

Conceptually:

```text
physical_stock = sum(posted physical movements)
reserved_stock = sum(active reservations)
available_stock = physical_stock - reserved_stock
```

Reservations affect availability but not accounting inventory value. Actual production consumption posts the inventory reduction and related cost.

## Sales and production states

Avoid one combined status enum.

Commercial state:

`Draft -> Quoted -> Confirmed -> Closed` with cancellation where appropriate.

Fulfillment state:

`Pending -> In Production -> Ready -> Delivered`

Payment state is derived:

`Unpaid | Partially Paid | Paid | Overpaid | Refunded`

Production jobs have their own operational state:

`Pending -> Ready -> In Progress -> Completed`, with Paused, Cancelled, or Failed when required.

## Payments

Payments are independent transactions. Payment allocations connect payments to invoices, allowing one payment to cover multiple invoices and one invoice to receive multiple payments.

An incoming payment not yet allocated to an invoice is customer credit. A deposit therefore does not need a special financial model.

## Accounting

Atropaten uses double-entry accounting internally.

Initial account groups include:

- Cash
- Bank
- Accounts Receivable
- Accounts Payable
- Inventory
- Sales Revenue
- Service Revenue
- Cost of Goods Sold
- Operating Expenses
- Loans Payable/Receivable
- Owner Equity
- Owner Drawings
- Retained Earnings
- Other Income/Expense

Normal workflows post journal entries automatically. Manual journal entry is an advanced accounting operation, not the normal way to record sales, purchases, or payments.

## Owners

Ownership percentage and profit-sharing percentage are separate concepts.

Owner capital, drawings, loans, reimbursements, and profit allocations must never be confused with ordinary business income or expense.

Operational costing and business accounting profit are calculated first. Profit/loss is allocated to owners only through an explicit fiscal-period closing/allocation process.

## Attachments

Large files are stored in a managed application directory, not as SQLite blobs. Metadata should include ownership/reference, kind, original filename, managed path, content hash, size, creation time, and notes.
