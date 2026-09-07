# M6-004 — UI system unification and UX hardening

## Objective

Unify Atropaten’s entire desktop UI into one coherent design system and remove accumulated layout, spacing, sizing, theme, state-feedback, and interaction inconsistencies across all workspaces.

This task is a product-wide UI/UX hardening pass, not a visual redesign and not a feature milestone. Preserve the existing information-dense Windows-first direction while making every screen feel like the same application.

The implementation must reduce one-off CSS and page-specific interaction behavior by introducing shared tokens, shared primitives, shared state patterns, and consistent workspace geometry.

## Scope

Audit and update all major workspaces and shared surfaces:

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
- Treasury/payment surfaces
- Checks
- Loans
- Owners
- Reports
- Settings
- Quotes/order editors/configurators
- print-preview entry surfaces where applicable
- global shell, sidebar, top bar, status bar, sticky headers, tabs, inspectors, forms, tables, and bottom action surfaces

## 1. Establish a single visual token system

Create or consolidate shared CSS variables/tokens for at least:

### Color

- application background
- surface
- raised surface
- subtle surface
- border
- stronger border/divider
- primary text
- secondary text
- muted text
- accent
- accent hover/pressed
- focus ring
- success
- warning
- danger/error
- informational
- disabled foreground/background/border
- selected row/control state
- hover row/control state

Use semantic tokens. Feature components should not invent arbitrary hex/rgb colors when an existing semantic token fits.

Green/red should remain semantic and not become decorative theme colors.

### Spacing

Define a compact spacing scale and use it consistently for:

- page/workspace gutters
- panel padding
- card padding
- section gaps
- form row gaps
- field label/control gaps
- table cell padding
- button icon/text gaps
- toolbar/action spacing
- tabs
- dialogs/confirmations

Remove arbitrary near-duplicate values where they create visible inconsistency.

### Typography

Centralize:

- font family
- base size
- compact metadata size
- labels
- table text
- titles
- section headings
- numerical/financial emphasis
- line heights
- font weights

Avoid oversized marketing-style headings. Keep the dense desktop ERP character documented in `docs/UI.md`.

### Shape/elevation

Centralize:

- border radius scale
- shadow/elevation scale
- border treatment
- separators

Do not use different card radii/shadows per feature without a functional reason.

### Sizing

Centralize dimensions for:

- buttons
- icon buttons
- inputs/selects/date controls
- table rows
- tabs
- sidebar items
- badges
- sticky surfaces
- status bar
- inspectors
- modals/confirmation panels if any

Controls serving the same purpose should have the same height throughout the application.

## 2. Unify global workspace geometry

Apply the documented workspace/sticky rules consistently everywhere.

Requirements:

- one shared workspace gutter system
- one page-header pattern
- one coordinated sticky-header/tab pattern
- one bottom-action pattern
- no page-specific fixed positioning when shared sticky primitives suffice
- no content hidden under sticky surfaces
- no unexplained large top gaps
- no nested vertical scrolling unless required by a specific component
- consistent z-index hierarchy
- consistent sticky borders/background/elevation
- correct scrolling/focus behavior

Refactor shared primitives where necessary rather than patching each page independently.

## 3. Width, padding, margin, overflow and responsive desktop cleanup

Audit every major page at representative widths:

- narrow laptop-like desktop
- ordinary desktop
- wide desktop

Fix:

- unused workspace width
- register/table panes that stop too early
- oversized inspectors
- undersized inspectors
- clipped inputs/selects/textareas
- controls extending outside panels
- accidental horizontal scroll in forms
- table overflow without deliberate scroll behavior
- inconsistent page margins
- inconsistent panel padding
- inconsistent section spacing
- two-column forms that do not collapse soon enough
- long labels/values/badges overflowing
- sticky surfaces overlapping content
- action bars hidden behind status bar

Use `min-width: 0`, `box-sizing: border-box`, sensible min/max widths, and responsive grid/flex behavior consistently.

Mobile redesign is out of scope; narrow desktop windows must still degrade gracefully.

## 4. Shared form-control system

Introduce or consolidate reusable field patterns for:

- text input
- numeric input
- money input
- quantity input
- select
- searchable select where already needed
- textarea
- checkbox/toggle
- Jalali date/date-time picker
- grouped fields
- field help/description
- inline action field

Every field must support a coherent state model:

- default
- hover
- focus
- populated
- disabled
- readonly
- required
- invalid
- warning where relevant
- success/confirmed only when meaningful

### Validation UX

Do not rely on raw backend errors dumped at page level when a specific field can be identified.

Implement a consistent model for:

- field-level validation messages directly beneath or adjacent to the field
- error border/state
- error icon only if useful and consistent
- preserving entered values after validation failure
- focusing or scrolling to the first invalid field on submit when practical
- clearing stale field errors after correction
- distinguishing client-side obvious validation from server/domain validation
- translating technical/backend validation messages into concise operator-facing text where the mapping is safe and unambiguous

Do not duplicate authoritative business rules in Vue. Frontend validation may catch obvious input-shape errors; backend/domain validation remains authoritative.

Examples of client-side validation suitable for UI:

- required field empty
- invalid numeric text
- malformed basic date range before submit
- zero/negative amount when the operation explicitly requires positive input

Examples that remain backend-authoritative:

- accounting balance/invariant
- lifecycle transitions
- ownership share coherence
- stock availability/reservation rules
- protected deletion
- closed fiscal periods
- historical reference protection

## 5. Application feedback and toast system

Replace inconsistent ad-hoc success/error notifications with one shared feedback system.

Support at least:

- success
- error
- warning
- informational

Requirements:

- consistent placement
- consistent sizing and spacing
- concise title/message hierarchy where useful
- semantic iconography
- appropriate duration
- errors should remain long enough to read
- destructive/important failures should not disappear too quickly
- duplicate rapid-fire toasts should be collapsed/deduplicated where practical
- allow manual dismissal
- success toasts should not obscure important controls
- screen-reader-friendly live-region behavior if practical in the current stack

Do not use toast as the only representation of a persistent page-blocking error.

## 6. Loading, submitting and async states

Every async user action must communicate progress coherently.

Standardize:

- initial page loading
- table loading
- inspector/detail loading
- submit/save loading
- destructive action loading
- backup/restore loading
- report loading
- print-document loading

Requirements:

- prevent duplicate submissions
- disable only the controls that must be disabled
- preserve context while loading
- use shared spinner/progress treatment
- do not replace entire complex screens with blank content for small background refreshes
- clearly distinguish loading from empty data
- never leave buttons stuck in loading state after failure

## 7. Empty, no-result and unavailable states

Provide consistent empty states for:

- truly empty domain
- filtered search with no matches
- unavailable due to prerequisite/configuration
- selected row removed/archived
- no report data in date range
- no transactions/history

Each state should explain what happened and provide the obvious next action when one exists.

Avoid decorative oversized illustrations; keep states compact and operational.

## 8. Error-state hierarchy

Implement a consistent hierarchy:

### Field error
For validation tied to one input.

### Form/operation error
For a failed save/post/action that affects the current form or panel.

### Page error
For failure to load the page’s core data.

### Toast error
For transient action feedback when page context remains valid.

### Critical confirmation/error
For destructive, financial, restore, close-period, reversal, cancellation, or other high-impact operations.

Raw Go/SQLite/Wails error strings should not normally be rendered directly to users when a safe friendly message can be produced. Keep enough technical detail available for debugging when necessary, but separate it from the primary operator-facing message.

## 9. Confirmation and destructive-action UX

Unify confirmation behavior for:

- Delete
- Cancel
- Void
- Reverse
- Restore backup
- Close fiscal period
- destructive status transitions where applicable

Requirements:

- action-specific title
- concise consequence description
- identify the affected record when available
- destructive primary button clearly styled as danger
- safe cancel/default focus behavior
- no generic `window.confirm` remaining on major workflows unless there is a documented reason
- disabled/protected delete should explain why when possible rather than simply fail silently

Archive and Delete must remain visually and semantically distinct.

## 10. Tables and data-density consistency

Unify all operational tables:

- header height
- row density
- typography
- numeric alignment
- money alignment
- hover state
- selected state
- status badges
- empty rows/state
- loading state
- sort affordance
- filter/action toolbar geometry
- horizontal overflow treatment

Money should align consistently and use grouped formatting.

Dates must use Jalali presentation through shared utilities.

Do not create page-specific status color mappings when shared semantic status primitives can cover them.

## 11. Buttons and actions

Define shared variants such as:

- primary
- secondary
- subtle/ghost
- danger
- icon-only
- compact table action

Standardize:

- height
- padding
- icon size
- icon/text gap
- hover
- pressed
- focus
- disabled
- loading

Primary actions should be visually obvious but not excessively large.

Avoid multiple competing primary buttons in the same action group.

## 12. Status badges and semantic states

Consolidate status badge styling across commercial, fulfillment, payment, production, check, loan, fiscal-period and other domain states.

Do not force unrelated meanings into identical colors if that would mislead the operator, but use one shared visual grammar.

Badge size, padding, radius, text weight and icon handling must be consistent.

## 13. Accessibility and keyboard UX

Without turning this task into a full accessibility certification, fix obvious desktop usability defects:

- visible keyboard focus
- sensible tab order
- labels associated with inputs
- icon-only buttons have accessible labels/tooltips
- disabled controls visually clear
- buttons use buttons, not clickable generic divs
- destructive confirmations keyboard-operable
- Escape closes dismissible transient surfaces where appropriate
- Enter submits forms only where predictable/safe
- color must not be the sole indicator of errors/status
- sufficient contrast for text, borders, disabled states, error messages and focus rings

Preserve reduced-motion support.

## 14. Theme consistency

Keep the product primarily light-themed for this release.

Unify all surfaces against the documented visual language:

- neutral/warm gray shell background
- white/subtle work surfaces
- restrained shadows
- subtle borders
- one accent
- minimal gradient usage

Remove local styles that visually contradict the rest of the application unless a domain-specific reason exists.

A dark theme is out of scope unless already implemented and trivial to keep consistent.

## 15. Centralize recurring UI primitives

Prefer shared components/utilities rather than repeated markup/CSS.

Audit whether the project should consolidate components such as:

- AppButton
- IconButton
- FormField / FieldMessage
- TextInput / SelectInput wrappers if useful
- MoneyInput
- QuantityInput
- StatusBadge
- AppToast / ToastHost
- InlineAlert
- EmptyState
- LoadingState / Spinner
- ConfirmDialog
- SectionPanel
- WorkspaceHeader
- WorkspaceTabs
- WorkspaceStickyStack
- WorkspaceBottomActions
- DataTable shell

Do not over-componentize trivial one-off markup. The goal is removal of visible/systemic inconsistency, not abstraction for its own sake.

## 16. Frontend error normalization

Add one central error-normalization utility for Wails/backend failures.

It should:

- accept unknown thrown values
- extract a stable readable message
- identify known validation/domain categories where safely possible
- map known protected-delete/conflict/not-found/validation cases to friendly UI copy
- preserve a fallback message
- avoid scattering `String(e)` throughout feature components

Do not parse unstable backend strings into new business logic. If backend error typing must be minimally improved to make UI handling reliable, keep that change narrow and preserve domain authority.

## 17. Audit every major screen

Do not stop after creating shared components. Migrate and inspect every major workspace.

For each workspace verify:

- header geometry
- tabs
- primary/secondary actions
- page padding
- panel padding
- form alignment
- input/control heights
- table density
- inspector width
- empty state
- loading state
- error state
- toast behavior
- delete/archive confirmation
- responsive desktop behavior
- no clipping/overflow
- Jalali date formatting
- grouped Rial/Toman formatting

## 18. Preserve domain and application behavior

This task must not rewrite accounting, inventory, production, pricing or persistence behavior merely to simplify the UI.

Do not:

- move authoritative calculations into Vue
- change accounting semantics
- change inventory semantics
- change posting/idempotency behavior
- change migrations unless narrowly required for a UI setting that already belongs in persistence
- replace controlled backend transitions with frontend guesses
- weaken confirmation/protection rules

## 19. Tests and verification

Add focused frontend tests where the project already has an appropriate testing mechanism, especially for shared utilities/components and error normalization.

At minimum, verify through implementation/build inspection:

- representative field validation
- failed submit preserves values
- loading prevents duplicate submit
- success/error/warning/info toast states
- protected delete shows useful failure
- destructive confirmation can cancel safely
- empty vs loading states are distinct
- money/date formatting uses shared utilities
- no known page still uses raw `String(error)` as its primary user-facing error path
- no major workflow still uses `window.confirm`

Do not add a large new frontend testing framework solely for this task unless clearly justified.

## 20. Documentation

Update `docs/UI.md` so the implemented design system is explicit rather than aspirational.

Document:

- token philosophy
- spacing scale
- typography hierarchy
- control sizes
- button variants
- feedback hierarchy
- form validation behavior
- toast behavior
- confirmation behavior
- loading/empty/error patterns
- shared workspace geometry
- responsive desktop acceptance widths

Create `docs/UI_UX_AUDIT.md` summarizing:

- screens audited
- shared primitives introduced/refactored
- major inconsistencies removed
- deferred visual/non-blocking issues
- any remaining manual Windows/WebView-specific UI checks

## Acceptance criteria

M6-004 is complete only when:

1. Major workspaces visibly share one layout/spacing/sizing system.
2. Page headers, tabs, panels, tables, forms and inspectors use consistent geometry.
3. Equivalent controls have equivalent sizes and states everywhere.
4. No major page has obvious unused-width, clipping, overflow or sticky-overlap defects at representative desktop widths.
5. Field-level validation is visually consistent and errors appear near the relevant input when possible.
6. Backend/domain failures use a central normalization path rather than ad-hoc `String(e)` rendering throughout the application.
7. Toasts use one shared success/error/warning/info system.
8. Loading/submitting states prevent duplicate actions and remain visually consistent.
9. Empty/no-result states are distinct from loading and failure states.
10. Major destructive workflows use the shared confirmation experience instead of browser `window.confirm`.
11. Archive and Delete remain distinct.
12. Money and user-facing dates remain consistently grouped and Jalali formatted.
13. Keyboard focus is visible and obvious accessibility defects in shared controls are corrected.
14. No accounting, inventory, production, reporting or persistence authority is moved into the frontend.
15. `docs/UI.md` reflects the final implemented rules.
16. `docs/UI_UX_AUDIT.md` records the product-wide pass.
17. Existing Go tests continue to pass.
18. Frontend production build passes.
19. `git diff --check` passes.

## Validation

Run:

```bash
go test ./...
cd frontend && npm run build
cd .. && git diff --check
```

If the existing frontend test suite is available, run it as well.

If Wails CLI is available, run an appropriate local Wails build/validation. If it is unavailable, state that explicitly.

Do not claim Windows/WebView native validation unless it was actually performed on Windows.

## Delivery

Commit directly to `main`, push `origin/main`, and print the final SHA.
