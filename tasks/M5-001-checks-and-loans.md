# M5-001 — Checks, loans, and installments

## Objective

Add Atropaten's non-trivial treasury obligations: incoming/outgoing checks and loans/installments, fully integrated with the existing double-entry ledger and cash/bank model.

This task must preserve M4 accounting invariants: posted financial history is immutable, every journal entry balances exactly in integer Rial, and all corrections use explicit reversals/compensating entries.

## Scope

### 1. Safe v10 -> v11 migration

Preserve all existing customers, suppliers, orders, invoices, purchases, inventory, production, accounting, payments, expenses, and transfers.

Add persisted models for:
- checks
- check status history/events as needed
- loans
- loan installments
- loan payments/allocations or linkage to the existing payment/journal infrastructure
- idempotency/source linkage needed for accounting postings

### 2. Checks

Support both incoming and outgoing checks.

Check fields:
- ID
- direction incoming/outgoing
- check number
- bank
- branch / account descriptor where useful
- amount Rial
- issue date
- due date
- payer/payee name
- customer/supplier reference where applicable
- linked financial obligation/payment/source reference where applicable
- notes
- current status
- created/updated timestamps

Statuses must support the practical shop lifecycle.

Incoming example:
- Received
- Deposited
- Cleared
- Returned
- Cancelled
- Transferred/Endorsed if implemented safely

Outgoing example:
- Issued
- Delivered
- Cleared/Paid
- Returned/Rejected where applicable
- Cancelled

Use one coherent state machine rather than free-form strings.

### 3. Check history

Financially meaningful status changes must preserve history.

Requirements:
- no mutation that destroys prior lifecycle evidence
- record status events/history with timestamps and notes
- reversal/cancellation semantics explicit
- invalid transitions rejected in Go
- retries idempotent

### 4. Incoming check accounting

Use explicit accounting treatment integrated with the existing ledger.

At minimum, choose and document a coherent model such as:
- receiving a customer check may move value from Accounts Receivable to Checks Receivable
- depositing may move Checks Receivable to Deposited Checks / Bank-in-transit if needed
- clearing moves value to Bank
- returned check reverses the appropriate stage and restores receivable/obligation

Keep the account model compact. Add only system accounts actually required.

Do not count an uncleared check as bank cash.

### 5. Outgoing check accounting

Use explicit accounting treatment, for example:
- issuing/delivering against supplier payable may move AP into Checks Payable/Outstanding Checks
- clearing moves the amount out of Bank
- cancellation/return restores the proper liability or payable

Ensure the same obligation cannot be paid twice through both a normal payment and check flow.

### 6. Due/overdue views

Derived status/time logic must expose:
- overdue
- due today
- due this week
- upcoming
- cleared/closed history

Use Jalali presentation while storing canonical date/time.

### 7. Loans

Support loans where the shop is either borrower or lender.

Loan fields:
- ID
- direction payable/receivable
- counterparty name and optional customer/supplier/owner reference for future compatibility
- principal Rial
- interest/fee configuration
- start date
- optional maturity/end date
- notes
- status
- linked ledger accounts
- created/updated timestamps

Do not implement a generic financial-product engine.

### 8. Installment schedule

Persist deterministic installments:
- installment number/position
- due date
- principal component
- interest/fee component
- total due Rial
- paid amount
- remaining amount
- status

Allow:
- manual/custom schedules
- straightforward equal-installment generation if useful
- partial payments
- early payment
- overdue detection

All money remains integer Rial.

### 9. Loan accounting

For borrowed money:
- receipt: Debit Cash/Bank, Credit Loans Payable
- repayments: Debit Loans Payable for principal, Debit Interest/Finance Expense for interest/fees, Credit Cash/Bank

For money lent by the business:
- disbursement: Debit Loans Receivable, Credit Cash/Bank
- repayments received: Debit Cash/Bank, Credit Loans Receivable for principal and Credit Interest/Other Income for interest as appropriate

Use exact explicit postings and idempotency.

### 10. Loan payment integration

Reuse the existing financial account/payment infrastructure where sensible, but do not force loan semantics into invoice/payment allocations if that makes the model unclear.

Requirements:
- installment payment records exact principal/interest split
- partial installment payment supported
- repeated payment attempts do not duplicate journals
- reversal restores installment balance and creates opposite accounting entries
- total principal paid cannot exceed principal unless an explicit overpayment/credit path is modeled

### 11. Checks workspace

Replace the Checks placeholder with a real workspace.

Views/filters:
- Incoming
- Outgoing
- Due / Overdue
- Cleared / Closed
- All

Register columns should include:
- number
- direction
- counterparty
- bank
- amount
- due date
- status
- linked source/party

Inspector/editor:
- full check metadata
- lifecycle history
- valid next actions only
- explicit confirmation for cancellation/return/reversal-style actions

### 12. Loans workspace

Replace the Loans placeholder with a real workspace.

Views:
- Active
- Payable
- Receivable
- Overdue
- Closed

Loan inspector/workspace:
- principal
- remaining principal
- total interest/fees
- paid amount
- next due date
- overdue amount
- installment schedule
- payment history
- record payment
- reverse payment where allowed

### 13. Dashboard/attention integration

Add lightweight real attention data where practical:
- overdue checks
- checks due soon
- overdue loan installments
- next installment due

Do not rebuild the whole dashboard yet; full dashboard/reporting is M6.

### 14. Deletion semantics

- Draft/unposted checks or loans may be hard-deleted only when no posted/history dependency exists.
- Any check/loan with posted financial history must not be destructively deleted.
- Cancellation/reversal preserves history.
- Archive may hide inactive completed configuration records but must never substitute for Delete.

### 15. UI requirements

Follow shared project geometry:
- use available workspace width
- no clipped inputs/selects/tables/inspectors
- shrinkable grid/flex children
- responsive laptop + wide desktop layouts
- shared sticky headers/actions
- grouped Rial/Toman formatting
- Jalali dates
- clear visual distinction for due/overdue/cleared/cancelled states

### 16. Tests

Add focused Go tests covering at minimum:
- v10 -> v11 migration preservation
- valid/invalid check transitions
- incoming check receive/deposit/clear accounting
- returned incoming check restores proper receivable/account balance
- outgoing check issue/clear accounting
- outgoing cancellation/reversal
- uncleared checks excluded from bank cash balance
- check idempotency
- due/overdue classification
- loan creation and schedule persistence
- borrowed-loan initial posting
- lender-loan initial posting
- installment partial payment
- principal/interest split accounting
- overdue installment derivation
- early/final repayment
- loan payment reversal
- retry idempotency
- exact integer-Rial behavior
- reopen/persistence

Existing tests must continue to pass.

## Non-goals

Do not implement yet:
- owners/shares/capital/drawings
- owner loans/reimbursements beyond future-compatible references
- fiscal periods/closing
- profit allocation
- bank statement reconciliation/import
- check printing
- loan amortization products beyond the practical schedule described above
- advanced interest compounding engines
- reports beyond operational due/overdue views

## Acceptance criteria

- v10 DB upgrades safely to v11.
- Incoming/outgoing checks have a persisted controlled lifecycle and history.
- Check status changes create correct balanced accounting effects exactly once.
- Returned/cancelled checks restore the correct obligation/account without mutating history.
- Uncleared checks are not presented as available bank cash.
- Loans payable/receivable and installment schedules persist correctly.
- Loan payments split principal/interest explicitly and post balanced journals.
- Partial payments, overdue state, reversals, and retries work correctly.
- Checks and Loans workspaces are usable and responsive.
- Existing M1-M4 flows remain intact.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.
