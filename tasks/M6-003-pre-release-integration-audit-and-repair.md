# M6-003 — Pre-release integration audit and repair

## Objective

Perform the first full-system pre-release audit of Atropaten after completion of M0–M6 implementation. This task is not a feature milestone. Its purpose is to validate the application's business logic, financial invariants, persistence, migration safety, deletion/reversal semantics, cross-domain flows, reports, backup/restore behavior, and UI integration as one coherent product, then fix every reproducible defect found within those areas.

Do not redesign working architecture for style. Prefer narrow, explicit repairs backed by tests.

## Scope

Audit the current v13 application from the perspective of a real print-shop workflow and verify that authoritative backend state remains internally consistent across all domains.

### 1. End-to-end business flows

Exercise and validate representative complete flows, including at minimum:

- customer → quote → order → production → invoice → payment → accounting
- supplier → purchase → inventory receipt → production consumption/waste → COGS
- partial and multiple customer payments
- supplier payment and payable reduction
- customer credit / unallocated deposit behavior
- expense posting and reversal
- cash/bank transfer and reversal
- incoming check lifecycle through clearing and return paths
- outgoing check lifecycle through clearing and cancellation/return paths
- payable and receivable loan creation, installments, partial repayment, final repayment, reversal
- owner capital contribution, drawing, owner-paid business expense, reimbursement, owner loans
- fiscal-period preview and closing
- backup → mutate live data → restore → reopen

Every flow must assert authoritative persisted results, not only UI state.

### 2. Accounting invariants

Audit all financial posting paths.

Required checks:

- every posted journal entry balances exactly in integer Rial
- no path can create a one-sided or zero-value effective journal entry
- posted journal entries and lines are immutable
- retries/idempotency keys cannot duplicate economic effects
- reversals create compensating history and never mutate/delete posted history
- order/invoice/payment status remains backend-derived
- purchase posting and cancellation remain balanced and idempotent
- invoice revenue and COGS recognition occur exactly once at the documented boundary
- check clearing/return paths never fabricate bank cash
- loan principal/interest allocation posts to the correct account classes
- owner drawings never become ordinary expenses
- owner loans remain distinct from capital
- fiscal-period closing does not double-count or permit posting/reversal into a closed period
- reports reconcile exactly to journal lines where they claim financial authority

Add invariant-style tests that scan all journal entries produced by integration scenarios.

### 3. Inventory and production invariants

Verify:

- physical stock reconstructs exactly from immutable movements
- reservations affect available stock but not physical stock or accounting consumption
- over-reservation remains impossible under updates/retries
- purchase receipt uses persisted landed cost and weighted-average behavior remains deterministic
- production consumption/waste uses exact persisted movement costs
- reversals/corrections use compensating movements
- no correction path silently edits historical movements
- COGS uses actual historical production/inventory cost, never current material/catalog price
- material conversions preserve fixed-scale quantity invariants

### 4. Deletion / purge / archive audit

Review every major CRUD domain against project deletion semantics.

For each relevant domain, verify explicitly whether records may be:

- archived/reactivated
- hard-deleted when genuinely unreferenced/draft
- protected from deletion once authoritative history exists
- corrected only through reversal/cancellation/void

Audit at least:

- customers
- suppliers
- materials
- services
- machines
- quotes
- orders
- purchases
- production jobs
- invoices
- payments
- expenses
- transfers
- checks
- loans
- owners
- fiscal periods

Fix any remaining archive-only alias where safe hard-delete is required, and any unsafe hard-delete path for protected history.

Add tests for both successful purge and blocked deletion.

### 5. Migration / persistence audit

Validate:

- clean database creation reaches v13 correctly
- sequential upgrades preserve invariants
- representative upgrades from v4/v5/v6/v8/v10/v11/v12/v13 where fixtures are practical
- schema reopen is deterministic
- foreign keys are enabled and checked
- migrations do not fabricate historical accounting events
- unknown future schema versions are rejected safely
- all SQLite transaction boundaries are atomic across cross-domain operations

Review any migration code that changes historical meaning or uses destructive table recreation.

### 6. Reports and derived balances

Cross-check reports against underlying authoritative sources.

Required reconciliations:

- P&L vs journal lines
- cash/bank vs journal lines
- AR/AP vs authoritative receivable/payable postings and allocations
- expenses vs journal lines
- inventory valuation vs reconstructed movements
- production consumption/waste vs movements/jobs
- customer/supplier statements vs authoritative financial history
- owner balances and fiscal-period allocation totals vs journals

Test exact boundary dates and empty periods.

### 7. Backup / restore safety audit

Review the M6-002 implementation beyond unit success.

Verify:

- SQLite Online Backup snapshot is actually used consistently
- archive traversal is impossible
- manifest/checksum coverage includes every managed file
- corrupt/missing archive contents fail before live switch
- rollback data survives until reopen succeeds
- failed reopen restores the previous live state
- database repository/service references are fully rebound after restore
- no stale service can continue referencing a closed pre-restore store
- backup/restore preserves cross-domain reconciliation results exactly

Add missing tests for any uncovered failure window.

### 8. UI integration audit

Do a code-level and runnable layout audit of all main workspaces.

Focus on functional consistency rather than cosmetic redesign:

- shared sticky header geometry
- bottom/status surfaces do not cover content
- no ordinary form control clips outside its panel
- flex/grid children use `min-width: 0` where needed
- inspectors/registers use available width sensibly
- no accidental large unused content regions
- responsive behavior at representative laptop and wide-desktop widths
- tables scroll horizontally only when genuinely needed
- Jalali dates are used consistently in user-facing surfaces
- grouped Rial/Toman formatting/input remains centralized
- no `$` or foreign currency remnants
- no frontend-owned accounting/status/business rules that should live in Go
- destructive actions have explicit confirmation
- invalid backend transitions are surfaced clearly

Repair obvious layout regressions found during this pass, especially previously deferred width/clipping problems.

### 9. API / Wails boundary audit

Verify:

- Wails methods remain thin adapters
- domain/accounting rules are not duplicated in Vue or DTO code
- generated bindings match current public App methods
- no stale manually appended duplicate declarations remain
- backend errors are not swallowed into misleading success UI

### 10. Concurrency / idempotency sanity

Where SQLite and current architecture permit practical testing, add focused tests for duplicate/retry attempts of:

- purchase posting
- invoice posting
- payment creation
- payment reversal
- check transition
- loan payment
- owner transaction
- fiscal close
- backup/restore switching

The goal is protection from duplicate economic effects, not broad concurrent-load engineering.

## Required outputs

1. Fix every reproducible issue found within this task's scope.
2. Add/expand automated tests for each repaired invariant.
3. Add `docs/PRE_RELEASE_AUDIT.md` containing:
   - areas audited
   - important invariants verified
   - defects found and fixed
   - any items that require native Windows/manual validation
   - any intentionally deferred non-blocking issues
4. Do not mark Windows packaging, WebView2, printing drivers, installer behavior, or native dialogs as validated unless actually tested on Windows.

## Non-goals

- new major product features
- payroll
- tax filing
- cloud sync
- SaaS/multi-user architecture
- broad visual redesign
- speculative refactors unrelated to a confirmed defect

## Validation

Run:

```bash
go test ./...
cd frontend && npm run build
cd .. && git diff --check
```

If Wails CLI is available, also run the appropriate local Wails build/dev validation. If unavailable, state that explicitly.

Review the complete final diff before finishing.

Commit to `main`, push `origin/main`, and print the final SHA.
