# M5-002 — Owners, fiscal periods, capital, drawings, and profit allocation

## Objective

Complete M5 by making multi-owner accounting first-class and adding fiscal-period closing/profit allocation without weakening the existing immutable double-entry ledger.

Atropaten must distinguish business expenses, owner drawings, owner loans, capital contributions, reimbursements, and profit allocation correctly.

## Core invariants

- Ownership percentage and profit-sharing percentage are separate concepts.
- Owner withdrawals are not ordinary business expenses.
- Owner personal purchases for the business must be represented as either owner payable/loan or capital contribution, never silently as revenue or miscellaneous income.
- Profit belongs to the business until an explicit fiscal-period allocation/closing operation.
- Closed-period allocations remain historically stable.
- All financial effects use balanced immutable journal entries in integer Rial.
- Posted owner/fiscal transactions are never mutated or deleted; corrections use reversal/correcting entries.
- Retry/idempotency must not duplicate postings or allocations.
- Vue/Wails contain no accounting rules.

## Scope

### 1. Safe v11 -> v12 migration

Preserve all M1-M5-001 data.

Add at minimum:
- owners
- ownership/profit-sharing history or effective-dated shares as needed
- owner financial transactions
- fiscal periods
- fiscal period closing/allocation records
- profit allocation lines
- source/idempotency/reversal metadata

Do not rewrite prior ledger history during migration.

### 2. Owners

Owner fields:
- ID
- name
- optional contact/notes
- active/archive state
- current ownership percentage
- current profit-sharing percentage
- created/updated timestamps

Rules:
- ownership percentages across active owners should be validated coherently when changed
- profit-sharing percentages may differ from ownership percentages
- percentages use deterministic fixed-scale representation, not float
- preserve historical changes sufficiently so prior period allocation remains explainable

### 3. Owner ledger accounts

Create/provision owner-specific accounts as needed, such as:
- Owner Capital
- Owner Drawings
- Owner Current / Owner Payable
- Owner Loan Payable
- Owner Loan Receivable
- Profit Allocation / retained earnings linkage as needed

Avoid creating an excessive chart. Use explicit account linkage per owner where practical.

### 4. Owner transactions

Support at least:

#### Capital contribution
Owner puts cash into business:
- Dr Cash/Bank
- Cr Owner Capital

Owner contributes inventory/material personally purchased for business:
- Dr Inventory (or appropriate asset)
- Cr Owner Capital OR Owner Payable/Loan depending on operator choice

#### Drawing / personal withdrawal
- Dr Owner Drawings / owner current
- Cr Cash/Bank

Must never hit ordinary expense accounts by the normal workflow.

#### Business expense paid personally by owner
- Dr appropriate Expense
- Cr Owner Payable / owner current

Later reimbursement:
- Dr Owner Payable
- Cr Cash/Bank

#### Owner lends money to business
- Dr Cash/Bank
- Cr Owner Loan Payable

Repayment:
- Dr Owner Loan Payable
- Cr Cash/Bank

#### Business lends money to owner
- Dr Owner Loan Receivable
- Cr Cash/Bank

Repayment received:
- Dr Cash/Bank
- Cr Owner Loan Receivable

All flows:
- explicit type
- amount Rial
- date/time
- linked financial account when relevant
- memo/reference
- journal linkage
- reversal support
- idempotency

### 5. Owner balances

Derive, never store as mutable truth:
- capital contributed
- drawings
- owner payable/current balance
- owner loan payable
- owner loan receivable
- allocated profit/loss
- current capital/equity summary

The UI should make these categories visibly distinct.

### 6. Fiscal periods

Fiscal period fields:
- ID
- name/label
- start date
- end date
- status Open / Closing / Closed
- closed timestamp
- closing journal/allocation linkage
- notes

Rules:
- periods cannot overlap
- a period must have start <= end
- only one coherent active/open period where applicable
- posted transactions inside a Closed period cannot be edited/reversed by ordinary workflows without an explicit post-close correction policy
- keep the policy simple and explicit; do not silently reopen periods

### 7. Period profit calculation

For a period, derive profit/loss from authoritative journal accounts:
- revenue + other income
- minus COGS/material cost
- minus operating/other expenses

Do not calculate profit from UI aggregates or order totals.

The calculation must reconcile to ledger lines and be deterministic in integer Rial.

### 8. Profit/loss allocation

At closing, allocate period profit or loss according to the period's captured profit-sharing percentages.

Requirements:
- snapshot each owner's allocation percentage into the closing record
- deterministic integer-Rial rounding
- any remainder must be assigned by a documented deterministic rule
- allocation remains historically stable if owner percentages change later
- closing is idempotent
- closing cannot run twice
- closed periods cannot overlap future/other periods

Accounting treatment should be explicit and balanced. Use Retained Earnings / Current Year Earnings and owner equity/current accounts consistently; document the chosen model.

### 9. Closing workflow

Before closing:
- validate period boundaries
- validate owner profit-sharing percentages sum correctly
- verify ledger is balanced
- calculate period P&L
- show a preview of owner allocation

Closing:
- execute atomically
- create required closing/allocation journal entries
- persist allocation snapshot
- mark period Closed

Do not implement complex statutory accounting close procedures beyond what the shop needs.

### 10. Reversal / correction policy

- Owner transactions with posted journal effects reverse through compensating entries.
- A Closed fiscal period cannot simply be deleted or reopened by ordinary delete/edit actions.
- If a correction to a closed period is required, use a clearly explicit correction/reopen mechanism only if implemented safely; otherwise reject and require a later-period correcting entry.
- Never mutate historical profit allocation rows.

### 11. Owners workspace

Replace the Owners placeholder with a real persisted workspace.

Sections/tabs:
- Overview
- Owners
- Transactions
- Capital & Drawings
- Loans / Current Accounts
- Profit Allocation

Show per owner:
- ownership %
- profit-sharing %
- capital
- drawings
- payable/current
- loans to/from business
- allocated profit/loss
- current equity summary

Actions:
- create/archive/reactivate owner
- update future/current shares safely
- record capital contribution
- record drawing
- record owner-paid expense
- reimburse owner
- record owner loan to/from business
- repayments
- reverse eligible transaction

### 12. Fiscal periods UI

Add a fiscal-period surface under Accounting or Owners.

Show:
- period
- dates
- status
- revenue
- COGS
- expenses
- profit/loss
- closing date

Closing workflow:
- preview calculated P&L
- preview exact owner allocations
- explicit confirmation
- clear errors if percentages or period state are invalid

### 13. Accounting integration

Extend Accounting workspace/journal inspection so owner and closing sources are clearly identifiable.

No duplicate balances stored in frontend.

### 14. UI/layout requirements

Follow global rules:
- shared sticky workspace geometry
- use available width
- no clipped fields/inspectors
- shrinkable grids/flex children
- grouped Rial/Toman money
- Jalali user-facing dates
- compact professional desktop layout
- destructive vs reversal actions clearly distinguished

### 15. Tests

Add focused Go tests for at least:
- v11 -> v12 migration preservation
- owner percentage fixed-scale validation
- ownership and profit-sharing independence
- capital contribution postings
- drawing postings and guarantee it does not hit ordinary expense
- owner-paid business expense + reimbursement
- owner loan payable/receivable flows
- transaction reversal/idempotency
- derived owner balances
- fiscal period overlap rejection
- closed-period protection
- period P&L derived from journal lines
- deterministic profit allocation rounding/remainder
- profit-sharing snapshot stability after later owner percentage changes
- closing idempotency
- closing balanced journal entries
- loss allocation behavior
- database reopen persistence

Existing tests must continue to pass.

## Non-goals

Do not implement yet:
- payroll
- tax filing/reporting
- statutory corporate accounting workflows
- complex consolidation
- multi-company/multi-shop ownership
- automatic dividend/legal distribution workflows
- full reporting suite
- backup/restore
- Windows installer

## Acceptance criteria

- Existing v11 DB upgrades safely.
- Multiple owners can have distinct ownership and profit-sharing percentages.
- Capital, drawings, owner-paid expenses, reimbursements, and owner loans are represented correctly through balanced journals.
- Drawings are not ordinary expenses.
- Owner balances derive from ledger history.
- Fiscal periods are persisted and non-overlapping.
- Period P&L derives from authoritative journals.
- Closing snapshots owner profit-sharing percentages and produces stable historical allocation.
- Closing and transaction posting are idempotent.
- Closed history is protected from destructive mutation.
- Owners and fiscal-period UI are usable.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.
