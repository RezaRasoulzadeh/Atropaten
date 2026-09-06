# M4-002 — Invoices, expenses, transfers, and operational accounting postings

## Objective

Complete M4 by adding the missing authoritative financial workflows on top of the v9 ledger: invoices, invoice allocations, sales/COGS posting, expenses, internal cash/bank transfers, and customer/supplier balance derivation.

This task must preserve the accounting foundation from M4-001 and avoid introducing a second source of truth.

## Core invariants

- The double-entry journal remains authoritative.
- Posted invoices, expenses, transfers, and journal history are immutable.
- Corrections use explicit reversal/void/correcting transactions.
- Every posting is transactionally idempotent.
- All authoritative money remains integer Rial.
- Customer/supplier balances derive from authoritative posted transactions and allocations, not mutable balance columns.
- Vue/Wails remain thin; business/accounting rules live in Go.

## Scope

### 1. Safe v9 -> v10 migration

Preserve all prior data.

Add at minimum:
- invoices
- invoice_items or immutable invoice snapshot lines
- invoice/order linkage
- payment allocation support to invoices
- expenses
- financial transfers
- posting/reversal linkage as needed

Use foreign keys and uniqueness constraints to protect idempotency and history.

### 2. Invoices

Add persisted invoices with:
- ID
- unique human-facing invoice number
- customer
- optional source order
- issue date
- due date
- status Draft / Posted / Partially Paid / Paid / Voided as appropriate
- notes
- subtotal Rial
- discount Rial
- total Rial
- paid Rial derived from allocations
- remaining Rial derived
- created/updated timestamps

Invoice lines must preserve immutable commercial snapshots:
- description/service snapshot
- quantity
- unit/selling amount
- line total
- optional order item/source linkage
- tax field only if already supported cleanly; do not invent tax complexity

A posted invoice must never silently reprice from current catalog data.

### 3. Order -> invoice flow

Support creating an invoice from an order.

Requirements:
- copy saved order item commercial snapshots, not current service pricing
- preserve order/customer/notes/discount where appropriate
- allow one invoice per order initially unless a simple multi-invoice model is already clearly supported
- prevent accidental duplicate invoice creation
- atomic linkage
- do not mutate historical order pricing

If multiple invoices per order are straightforward with existing architecture, support them; otherwise explicitly enforce one invoice per order for this milestone.

### 4. Sales posting

Define and implement the authoritative sales recognition boundary.

When an invoice is Posted:
- Debit Accounts Receivable
- Credit Sales Revenue / Service Revenue using a deterministic mapping

Do not post cash as revenue directly.

If order items distinguish service/material sales in a useful generic way, map to Sales vs Service Revenue explicitly. Otherwise use a single deterministic revenue account for now and document it.

Invoice void/reversal:
- create exact opposite journal effect
- preserve invoice and original posting
- reverse eligible allocations or reject void while settled if safer and clearly explained
- be idempotent

### 5. COGS / inventory posting

Integrate production/inventory with accounting without mutating historical inventory movements.

For inventory-backed material consumption that has become economically final, post:
- Debit COGS / Material Cost
- Credit Inventory

Choose one explicit recognition trigger and document it. Preferred trigger:
- invoice posting or final delivery/closure, provided the implementation can deterministically identify posted production consumption tied to the invoiced order.

Requirements:
- no duplicate COGS posting for same consumption
- use actual inventory movement costs, not current material price
- waste accounting treatment must be explicit and deterministic
- reversal/void must create compensating journal entries without altering movements

If consumption is not yet available for an order, do not fabricate COGS.

### 6. Payment allocation migration to invoices

Extend M4-001 payment allocation so customer payments can allocate to invoices.

Requirements:
- one payment can allocate across multiple invoices
- partial payments supported
- invoice paid/remaining/status derived from active allocations
- customer credit remains supported
- allocations cannot exceed eligible outstanding amount unless explicitly representing credit
- reversing payment reverses invoice allocation effects

Existing order allocations must remain readable and valid after migration. Do not silently corrupt or discard them.

New UI should favor invoice allocation for posted invoices.

### 7. Receivables/payables derivation

Provide authoritative queries for:
- customer receivable balance
- customer credit balance
- supplier payable balance
- per-invoice outstanding
- per-purchase outstanding

Balances must derive from posted sources/payments/reversals.

Do not maintain manually updated customer/supplier balance columns.

### 8. Expenses

Persist expenses with:
- ID
- date
- category/account
- payee/optional supplier
- description
- amount Rial
- payment method
- financial account
- notes
- status Posted / Reversed
- journal linkage

Posting:
- Debit selected Expense account
- Credit Cash/Bank

Support a small useful set of expense categories via accounts, not hardcoded UI-only labels.

Examples:
- Rent
- Electricity/utilities
- Internet
- Repairs/maintenance
- Transport
- Software
- Salaries/wages placeholder account if needed, without implementing payroll
- Tax/fees
- Other expense

Posted expense cannot be hard-deleted. Reverse it instead.

### 9. Cash/bank transfers

Add internal financial-account transfers.

Transfer:
- source cash/bank account
- destination cash/bank account
- amount Rial
- date
- reference
- notes

Posting:
- Debit destination ledger account
- Credit source ledger account

Requirements:
- source != destination
- positive amount
- transactionally idempotent
- immutable after posting
- reversal creates opposite entry
- no income/expense effect

### 10. Cash/bank management UI

Extend Accounting workspace with a Treasury section or equivalent:
- financial accounts
- derived balances
- transfers
- recent account movements

Allow adding/editing/archive of non-system financial-account metadata where safe.

Do not hard-delete financial accounts referenced by posted history.

### 11. Invoices workspace

Add a real Invoices workspace reachable from navigation or Accounting/Sales flow.

Required views:
- invoice register
- search/filter/status
- create/open invoice from order
- invoice workspace/inspector
- line snapshot display
- issue/due dates
- totals/paid/remaining
- post action
- void/reversal action
- payment allocation visibility
- customer/history context

Follow shared sticky/workspace rules.

### 12. Order integration

Orders should show:
- invoice linkage/status
- invoiced total
- paid/remaining based on invoice/payment state where an invoice exists
- action to create/open invoice

Do not allow duplicate financial state calculations in Vue.

### 13. Supplier/purchase integration

Purchases should continue to expose:
- total
- paid
- remaining
- supplier payable state

Customer and supplier views should gain a lightweight financial summary where practical:
- receivable/payable
- credit
- recent financial activity

### 14. Accounting workspace completion

Accounting should now expose:
- Overview
- Accounts
- Journal
- Payments
- Invoices
- Expenses
- Transfers/Treasury

Overview should derive:
- Cash
- Bank
- Accounts Receivable
- Accounts Payable
- Sales revenue for implemented period query
- Expenses for implemented period query

Do not implement full financial statements yet.

### 15. Deletion/reversal rules

- Draft invoice may hard-delete if it has no protected downstream references.
- Posted invoice cannot hard-delete.
- Posted expense cannot hard-delete.
- Posted transfer cannot hard-delete.
- Payment/journal history cannot hard-delete.
- Financial account referenced by posted history cannot hard-delete.
- Protected delete requests return clear domain errors, never silently archive.

### 16. UI/layout requirements

Apply project-wide rules:
- full useful workspace width
- no accidental overflow/clipping
- `min-width: 0` where needed
- responsive form grids/inspectors
- deliberate horizontal scroll only for wide registers
- shared sticky headers/actions
- grouped Rial/Toman inputs/displays
- Jalali business dates
- explicit confirmations for post/void/reverse/destructive actions

### 17. Tests

Add focused Go tests for at least:
- v9 -> v10 migration preservation
- invoice numbering and persistence
- order -> invoice snapshot immutability
- duplicate conversion protection
- invoice posting AR/revenue journal
- invoice posting idempotency
- invoice void/reversal
- invoice payment partial/multi-allocation
- payment reversal restoring invoice outstanding
- customer receivable/credit derivation
- supplier payable derivation remains correct
- expense posting and reversal
- transfer posting and reversal
- transfer has no P&L effect
- COGS posting uses actual inventory movement cost
- COGS idempotency
- COGS reversal/void compensation
- protected deletion behavior
- exact integer Rial behavior

Existing tests must continue to pass.

## Non-goals

Do not implement yet:
- printable PDF layouts (M6)
- tax return/reporting engine
- recurring expense scheduler
- payroll
- checks lifecycle
- loans/installments
- owner accounting/profit allocation
- fiscal period close
- bank reconciliation/import
- full balance sheet/income statement reports

## Acceptance criteria

- Existing v9 DB migrates safely.
- Orders can produce immutable invoices.
- Posted invoice creates balanced AR/revenue accounting exactly once.
- Actual production consumption can create deterministic Inventory/COGS accounting exactly once at the chosen recognition boundary.
- Customer payments allocate across invoices and derive outstanding balances.
- Customer/supplier balances derive from authoritative transactions.
- Expenses post to expense + cash/bank correctly.
- Transfers move value between financial accounts without affecting profit.
- Posted financial records use reversal rather than destructive deletion.
- Accounting/Invoices UI is usable and follows shared layout conventions.
- Existing M1-M4-001 workflows remain intact.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.

## Manual verification for final integration pass

1. Create/confirm an order and create an invoice from it.
2. Modify catalog pricing and confirm the old invoice remains unchanged.
3. Post invoice and inspect AR/revenue journal.
4. Allocate one payment across two invoices and verify outstanding balances.
5. Reverse payment and verify balances restore.
6. Complete production consumption and verify COGS/Inventory posting at the documented trigger.
7. Record a cash expense and verify expense/cash journal.
8. Transfer money from Cash to Bank and verify no income/expense impact.
9. Void/reverse representative posted records and inspect compensating entries.
10. Restart and verify persistence and derived balances.
