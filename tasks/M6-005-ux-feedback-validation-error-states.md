# M6-005 — UX feedback, validation, and interaction-state hardening

## Objective

Harden Atropaten’s product-wide interaction behavior after the M6-004 visual/layout unification pass.

This task focuses on validation UX, frontend error handling, toast/feedback behavior, loading/submitting states, empty/error/unavailable states, destructive confirmations, protected-action feedback, keyboard interaction, and consistency of operator-facing state transitions.

Do not redesign layouts or change accounting/inventory/business semantics except for minimal narrow backend error typing needed to present reliable UI feedback.

## Scope

Audit all major workspaces and high-impact workflows:

- Dashboard
- Orders
- Production
- Customers
- Services
- Materials
- Purchases
- Suppliers
- Accounting
- Invoices
- Expenses
- Treasury/payments
- Checks
- Loans
- Owners
- Reports
- Settings
- Quotes/configurators
- backup/restore
- fiscal close
- cancellation/reversal/void/delete/archive workflows

## 1. Central frontend error normalization

Create one shared error-normalization utility for Wails/backend failures.

It must:

- accept unknown thrown values
- extract a stable readable message
- distinguish known validation/conflict/not-found/protected-delete style failures where safely possible
- map technical failures to concise operator-facing copy where mapping is stable
- preserve a fallback message
- optionally retain technical detail for debugging without making it the primary UI copy
- replace scattered `String(e)`/raw-error rendering in major workspaces

Do not parse unstable backend strings into business logic. If narrow backend error categories/types are required for reliable handling, keep the backend change minimal and preserve domain authority.

## 2. Shared form validation UX

Implement one coherent field-validation pattern.

Each field must support relevant states:

- default
- focused
- disabled
- readonly
- required
- invalid
- warning where meaningful

Requirements:

- field errors shown beside/below the relevant control
- invalid state visible without relying only on color
- preserve entered values after submit failure
- clear stale field errors after correction where safe
- focus/scroll to first invalid field when practical
- distinguish frontend shape validation from backend/domain validation
- do not duplicate authoritative domain rules in Vue

Suitable frontend validation includes:

- required value missing
- invalid numeric text
- clearly malformed date range
- explicitly positive-only amount <= 0

Backend remains authoritative for:

- stock/reservation rules
- lifecycle transitions
- journal/accounting invariants
- ownership-share coherence
- protected deletion
- closed periods
- historical-reference protection

## 3. Toast and transient feedback system

Replace inconsistent notifications with one shared toast/feedback system.

Support:

- success
- error
- warning
- information

Requirements:

- consistent position and width
- consistent title/message hierarchy
- semantic iconography
- sensible duration
- errors remain readable long enough
- important/destructive failures do not disappear immediately
- manual dismiss
- deduplicate repeated rapid-fire messages where practical
- do not obscure primary controls
- screen-reader live-region behavior where practical

A toast must not be the only representation of a page-blocking or persistent error.

## 4. Loading and submitting states

Standardize async behavior for:

- initial page load
- list/table load
- inspector/detail load
- save/post
- delete/archive/reactivate
- cancellation/reversal/void
- backup/restore
- report load
- print-document load
- fiscal close

Requirements:

- prevent duplicate submission
- disable only necessary controls
- preserve screen context while background work runs
- use shared spinner/progress treatment
- never confuse loading with empty data
- never leave a button stuck after failure
- show action-specific pending text where useful

## 5. Empty, filtered-empty, unavailable, and error states

Implement shared state patterns for:

- truly empty domain
- filtered search with no results
- unavailable because prerequisite/configuration is missing
- selected record removed/archived
- no report data in selected range
- no transaction/history rows
- page load failure
- partial-panel load failure

Each state should explain what happened and expose the obvious next action where one exists.

Keep states compact and operational rather than decorative.

## 6. Error-state hierarchy

Use a consistent hierarchy:

### Field error
Validation tied to one input.

### Form/operation error
Save/post/action failed while the surrounding page remains usable.

### Page error
Core page data could not load.

### Toast error
Transient action feedback when current page context remains valid.

### Critical confirmation/failure
High-impact financial/destructive workflow such as restore, fiscal close, reversal, cancellation, void, or delete.

Raw Go/SQLite/Wails error strings should not normally be the primary user-facing message.

## 7. Confirmation system

Replace major `window.confirm` usage with one shared confirmation experience.

Cover:

- Delete
- Archive where consequence deserves confirmation
- Cancel
- Void
- Reverse
- Restore backup
- Close fiscal period
- destructive lifecycle transitions

Requirements:

- action-specific title
- concise consequence explanation
- record identity when available
- danger styling for destructive primary action
- safe Cancel default/focus behavior
- keyboard operable
- Escape closes when safe
- loading state during confirmed async action
- protected/blocked operation explains why

Archive and Delete must remain visibly and semantically distinct.

## 8. Protected-action UX

When backend rejects an operation because of authoritative history or a domain invariant, make the reason understandable.

Examples:

- referenced catalog record cannot be deleted
- closed-period posting blocked
- stock/reservation prevents cancellation
- proof-linked attachment cannot be removed
- settled invoice cannot be voided directly

Do not silently substitute Archive for Delete or otherwise alter requested semantics.

## 9. Keyboard/focus interaction hardening

Fix obvious desktop interaction defects:

- visible focus
- predictable tab order
- first invalid field focus after failed validation when useful
- confirmation dialog focus containment/restoration
- icon buttons accessible by keyboard
- Enter submits only where safe/predictable
- Escape dismisses transient/dialog surfaces where safe
- tooltips/accessible labels for icon-only controls

Do not turn this into a full accessibility certification task.

## 10. Operation-specific feedback

Review financial and destructive workflows so success/failure messages describe what actually happened.

Examples:

- payment posted
- payment reversed
- check transitioned
- loan payment posted/reversed
- purchase cancelled
- invoice voided
- owner transaction posted/reversed
- fiscal period closed
- backup created/verified/restored
- record archived/deleted/reactivated

Avoid vague generic “Saved” feedback for materially different operations.

## 11. Audit every major screen

For every major workspace verify:

- initial loading state
- empty/no-result state
- page error state
- field error treatment
- failed-submit behavior
- success feedback
- action failure feedback
- duplicate-submit prevention
- destructive confirmation
- protected-delete feedback
- archive/delete distinction
- disabled/read-only clarity
- keyboard focus
- no raw `String(error)` as primary UI feedback

## 12. Preserve domain behavior

Do not:

- move accounting/inventory/pricing logic into Vue
- duplicate backend transition rules
- weaken backend validation
- alter idempotency boundaries
- change deletion semantics to make UX simpler
- silently auto-correct financially meaningful input

## 13. Tests and verification

Use the project’s existing frontend test mechanisms if present. Do not introduce a large new testing framework solely for this task.

Add focused tests where appropriate for:

- error normalization
- obvious field validation
- failed submit preserves values
- loading blocks duplicate submission
- toast variants
- confirm cancel/confirm paths
- protected delete feedback
- loading vs empty vs error distinction

At minimum verify by source/build audit that:

- major `String(e)` usage is removed from operator-facing paths
- major `window.confirm` usage is removed
- major async mutations have duplicate-submit protection
- all major workspaces expose distinct loading/empty/error behavior

## 14. Documentation

Update `docs/UI.md` with final implemented UX rules for:

- validation behavior
- error hierarchy
- toast behavior
- loading/submitting behavior
- empty/unavailable states
- confirmation behavior
- protected-action feedback
- keyboard/focus conventions

Create `docs/UI_UX_STATE_AUDIT.md` documenting:

- screens/workflows audited
- shared feedback/state primitives introduced
- raw-error/confirmation inconsistencies removed
- remaining manual Windows/WebView interaction checks
- intentionally deferred non-blocking UX issues

## Acceptance criteria

M6-005 is complete only when:

1. Major frontend failures use one central normalization path.
2. Field-level validation is consistent and preserves entered values after failure.
3. Major forms prevent duplicate async submission.
4. Toasts use one shared success/error/warning/info system.
5. Loading, empty, unavailable, and error states are visibly distinct.
6. Major destructive/financial workflows use shared confirmation instead of browser `window.confirm`.
7. Protected actions explain why they are blocked.
8. Archive and Delete remain distinct.
9. Keyboard focus and dialog interaction are predictable.
10. Major workspaces no longer rely on raw `String(e)` as primary feedback.
11. No authoritative domain logic is moved into Vue.
12. `docs/UI.md` contains the implemented UX-state rules.
13. `docs/UI_UX_STATE_AUDIT.md` records the full pass.
14. Existing Go tests pass.
15. Frontend production build passes.
16. `git diff --check` passes.

## Validation

```bash
go test ./...
cd frontend && npm run build
cd .. && git diff --check
```

Run existing frontend tests if available.

If Wails CLI is unavailable, state that explicitly. Do not claim Windows/WebView native validation unless actually performed on Windows.

## Delivery

Commit directly to `main`, push `origin/main`, and print the final SHA.
