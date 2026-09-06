# M4-001 — Double-entry accounting foundation and payments

## Objective

Introduce Atropaten's authoritative financial ledger and first real payment flows without overbuilding a general accounting platform.

This vertical slice adds the chart of accounts, balanced immutable journal entries, cash/bank accounts, customer/supplier payments and allocations, and automatic posting for the commercial/purchase flows already implemented.

## Core invariants

- Double-entry accounting is authoritative.
- Every posted journal entry balances exactly in integer Rial: total debit == total credit.
- Posted journal entries and lines are immutable.
- Corrections use explicit reversing/correcting entries; never edit financial history.
- Business workflows post accounting through Go application services; Vue and Wails contain no journal logic.
- Retry/idempotency must never duplicate financial postings.
- All authoritative money remains integer Rial.
- Operational source records retain links to generated journal entries where useful for traceability.
- Archive and hard-delete remain distinct. Posted financial history is never hard-deleted.

## Scope

### 1. Safe v8 -> v9 migration

Preserve all existing catalog, customer, quote/order, supplier/purchase, inventory, reservation, and production data.

Add at minimum:
- accounts
- journal_entries
- journal_lines
- financial_accounts (cash/bank)
- payments
- payment_allocations
- source/posting idempotency/linkage metadata as needed

Foreign keys and constraints must protect ledger integrity.

### 2. Chart of accounts

Seed a compact default chart suitable for one print shop, including at least:
- Cash
- Bank
- Accounts Receivable
- Accounts Payable
- Inventory
- Sales Revenue
- Service Revenue
- Cost of Goods Sold / Material Cost
- Operating Expenses
- Loans Payable
- Loans Receivable or Other Receivable if needed
- Owner Equity
- Owner Drawings
- Retained Earnings
- Other Income
- Other Expense

Account model should support:
- stable ID/code
- name
- account type/category
- optional parent account
- active/archive state
- system/protected flag where useful

Do not implement arbitrary enterprise account dimensions.

### 3. Journal model

Journal entry:
- ID
- human-readable reference/number
- posting date/time
- description
- source type + source ID
- reversal linkage when applicable
- created timestamp

Journal line:
- ID
- journal entry ID
- account ID
- debit Rial
- credit Rial
- optional party/customer/supplier reference when useful
- memo
- deterministic position

Validation:
- at least two effective lines
- non-negative debit/credit
- each line uses debit or credit, not both
- exact balance required
- referenced accounts exist and are active/allowed for posting

### 4. Journal application service

Provide a small explicit Go posting API used by business workflows.

Requirements:
- transactionally create balanced entries
- idempotency by source/action key
- return existing posting on safe retry
- reverse an entry by creating the exact opposite entry
- no update/delete of posted entries
- queries by date/account/source/party as needed for UI

Do not create a formula/rules engine for accounting mappings.

### 5. Financial accounts / treasury basics

Persist operator-visible financial accounts:
- Cash register
- one or more Bank accounts

Fields:
- ID
- name
- type cash/bank
- linked ledger account ID
- optional bank/account/card/POS descriptive metadata
- active/archive state

Seed a default Cash financial account cleanly.

Balances must derive from posted journal lines, not a mutable balance column.

### 6. Payments

Payment model:
- ID
- direction incoming/outgoing
- method cash / bank_transfer / card / check / account_credit / other
- amount Rial
- date/time
- financial account when method requires it
- customer or supplier/party reference as appropriate
- reference
- notes
- status posted/reversed (or equivalent)
- created timestamp

Payment allocations:
- payment ID
- target type initially order/customer receivable or purchase/supplier payable as supported by current model
- target ID
- amount Rial
- deterministic position

Allow:
- partial payment
- multiple payments against one obligation
- one payment allocated across multiple eligible obligations if kept straightforward
- unallocated incoming customer money as customer credit
- unallocated outgoing supplier money as supplier advance/credit where modeled safely

Allocation totals cannot exceed payment amount.

### 7. Current order payment integration

There is no invoice domain yet, so M4-001 may treat confirmed/closed order commercial balances as the current receivable source until invoices arrive in a later slice.

Requirements:
- record customer payment
- allocate to one or more orders
- derive order paid amount and remaining amount
- derive payment status: Unpaid / Partially Paid / Paid / Overpaid/credit only where semantics are coherent
- do not let frontend directly set payment status
- preserve historical order selling totals

Automatic accounting for incoming allocated customer payment:
- Debit Cash/Bank
- Credit Accounts Receivable

If recording an unallocated customer deposit/credit, use a clear customer-credit liability or explicit interim account rather than falsely reducing a specific receivable. Add the minimal account needed if necessary.

### 8. Purchase/supplier payment integration

Posted purchases currently create inventory effects but no accounting.

Add accounting integration for new/post-migration business operations:

When a purchase is posted:
- Debit Inventory for landed inventory value
- Credit Accounts Payable for supplier obligation

Supplier payment allocation:
- Debit Accounts Payable
- Credit Cash/Bank

Cancellation/reversal of a posted purchase must create corresponding accounting reversal without deleting history.

Do not silently backfill historical v8 purchases with fabricated posting dates unless the migration/task implements an explicit deterministic opening/backfill strategy and tests it. Prefer preserving old data and clearly treating accounting activation as beginning at v9 if that is safer.

### 9. Sales recognition boundary

Do not prematurely invent invoice accounting.

For this task:
- payment/receivable ledger integration is required
- full sales revenue/COGS recognition timing may remain deferred to the invoice/delivery accounting slice
- document the chosen boundary explicitly so cash receipts are not accidentally treated as revenue twice later

### 10. Payment reversals

Posted payments cannot be hard-deleted.

Reversal must:
- create opposite journal effect
- release/reverse allocations
- preserve original payment
- be idempotent
- update derived order/purchase balances correctly

Draft/unposted financial records, if the implementation introduces any, may be hard-deleted when safe.

### 11. Accounting workspace

Replace the Accounting placeholder with a real workspace using the shared UI geometry.

Tabs/sections:
- Overview
- Accounts
- Transactions / Journal
- Payments

Overview:
- Cash
- Bank
- Accounts Receivable
- Accounts Payable
- current ledger summary available from implemented postings

Accounts:
- compact hierarchical/list view
- code/name/type
- derived balance
- active/archive where safe
- system accounts protected from destructive deletion

Journal:
- dense register
- date/reference/source/description/debit/credit
- inspector with lines
- clear reversal linkage

Payments:
- incoming/outgoing
- customer/supplier
- method
- financial account
- amount
- allocation summary
- status/date
- create payment workflow
- reverse action with explicit confirmation

### 12. Order and supplier UI integration

Orders:
- show Paid and Remaining values
- payment status derived from backend
- Payments tab becomes functional
- allow recording/allocating customer payments

Suppliers/Purchases:
- show payable/paid/remaining where applicable
- allow recording supplier payment from a suitable workflow

Use shared grouped Rial/Toman input/display everywhere.

### 13. UI/layout requirements

Follow project-wide rules:
- use full useful workspace width
- no clipped/overflowing inputs, selects, tables, or inspectors
- flex/grid children must shrink correctly
- forms rebalance before clipping
- shared sticky headers/action surfaces
- laptop and wide-desktop geometry
- Jalali user-facing dates
- destructive/reversal actions clearly differentiated

### 14. Tests

Add focused Go tests for at least:
- v8 -> v9 migration preservation
- default chart seeding and reopen idempotency
- balanced journal enforcement
- rejection of invalid/unbalanced entries
- immutable posted entries
- posting idempotency
- exact reversal entry
- cash/bank balance derivation
- incoming customer payment posting
- partial and multiple allocations
- unallocated customer credit behavior
- order paid/remaining/payment-state derivation
- purchase posting accounting for new v9 operations
- supplier payment posting/allocation
- payment reversal
- purchase cancellation accounting reversal
- retry does not duplicate journals/payments
- exact integer-Rial roundtrips

Existing tests must continue to pass.

## Non-goals

Do not implement yet:
- invoices/receipts/PDFs
- fiscal period closing
- owner capital/profit allocation
- checks lifecycle
- loans/installments
- recurring expenses
- payroll
- tax reporting
- bank reconciliation/import
- advanced manual journal editor beyond inspection/reversal if unnecessary
- full revenue/COGS recognition if invoice/delivery boundary is not yet defined
- financial reports beyond basic workspace summaries

## Acceptance criteria

- Existing v8 DB upgrades safely.
- A real balanced double-entry ledger exists.
- Posted financial history is immutable and reversible.
- Cash/bank balances derive from journal lines.
- Customer payments support partial allocation and update order payment state.
- Supplier payments update payable state.
- New posted purchases create Inventory/AP journal effects exactly once.
- Purchase cancellation reverses accounting exactly once.
- Payment reversal preserves history and restores derived balances.
- Accounting workspace is usable.
- Order Payments UI is functional.
- Grouped Rial/Toman and Jalali conventions remain consistent.
- Existing M1-M3 workflows remain intact.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.

## Manual verification for final integration pass

1. Open Accounting and inspect seeded accounts.
2. Create/confirm an order and record a partial cash payment.
3. Verify order paid/remaining and journal lines.
4. Record the remaining bank payment and verify Paid state.
5. Record an unallocated customer deposit and verify it is not falsely applied to an order.
6. Post a new supplier purchase and verify Inventory/AP posting.
7. Record supplier payment and verify AP decreases.
8. Reverse a payment and verify opposite journal plus restored balance.
9. Cancel a posted purchase and verify compensating accounting history.
10. Restart and verify persistence and derived balances.
