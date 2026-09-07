# M6-004 — UI system unification and UX hardening

## Objective

Perform a complete frontend UI/UX consistency pass across Atropaten after M0–M6 feature completion.

This task is not a visual redesign. Preserve the established Windows-first dense ERP direction, but remove accumulated one-off layout, spacing, sizing, styling, feedback, and interaction inconsistencies so the application behaves and looks like one coherent desktop product.

The result must establish reusable UI primitives/tokens and migrate all major workspaces to them.

## Scope

Audit and repair the full frontend, including:

- App shell
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
- Treasury
- Checks
- Loans
- Owners
- Reports
- Settings
- Quote/order inspectors and configurators
- Print-preview entry points where they share application UI

Do not change authoritative business/accounting logic unless a UI defect exposes a real integration bug. Keep backend behavior intact.

## 1. Design tokens and global visual system

Replace repeated arbitrary values with a coherent shared token system where practical.

Centralize at minimum:

- application/surface/background colors
- text colors and muted text hierarchy
- primary accent
- success/warning/error/info semantic colors
- border colors
- focus-ring treatment
- shadows/elevation
- radii
- spacing scale
- control heights
- icon sizes
- typography sizes/weights/line heights
- table row heights
- sticky offsets/z-index layers
- animation durations/easing

Do not introduce multiple competing token systems.

Preserve the light neutral desktop theme defined by `docs/UI.md`:

- neutral/warm-gray application background
- white or subtly differentiated surfaces
- restrained borders/shadows
- one main accent
- green/red primarily for meaningful status/financial semantics
- approximately 14px regular UI text
- compact dense desktop controls/tables

Remove accidental theme variations between pages.

## 2. Layout and geometry unification

All major pages must use the same workspace geometry.

Unify:

- workspace horizontal gutters
- top spacing below the global topbar
- page-header height/padding
- title/subtitle spacing
- page-level actions
- tab-strip height/padding
- section gaps
- card/panel padding
- table containers
- inspector/editor widths
- form spacing
- bottom-action spacing
- sticky offsets
- content max/min widths

Do not leave page-specific arbitrary margins/paddings unless structurally necessary.

Use the existing shared sticky workspace components or improve them centrally instead of implementing page-local sticky rules.

The shell remains responsible for the viewport. Feature pages must not invent competing viewport/fixed-height behavior.

## 3. Width, clipping, overflow and responsive desktop behavior

Perform an explicit full-app layout audit.

Requirements:

- use available workspace width intentionally
- registers/tables flex into available space
- right inspectors have coherent shared sizing rules
- apply `min-width: 0` correctly to shrinking grid/flex children
- inputs/selects/textareas never extend outside their panel
- no controls disappear under sticky headers/status bars
- two-column forms collapse/rebalance before clipping
- long text wraps/truncates intentionally
- genuine wide tables may scroll horizontally
- ordinary forms must not require horizontal scrolling
- no unexplained narrow content column inside a wide workspace
- avoid nested vertical scroll containers unless necessary

Validate representative widths around:

- 1280×720
- 1366×768
- 1440×900
- 1920×1080

Mobile optimization is not required, but narrower desktop windows must remain usable.

## 4. Shared component system

Identify duplicated UI patterns and replace them with reusable components or shared classes/utilities where this reduces inconsistency.

Standardize at minimum:

- buttons: primary / secondary / ghost / destructive / icon
- inputs
- textareas
- selects
- checkboxes/toggles where present
- field labels
- field help text
- field errors
- form groups
- panels/cards
- badges/status pills
- tabs
- tables/registers
- empty states
- loading states
- error states
- confirmation surfaces
- toast/notification system
- inline banners/alerts
- inspector headers/actions
- destructive-action confirmation pattern

Do not over-abstract feature-specific controls into generic components when doing so harms readability.

## 5. Form UX and validation states

Audit every user-editable form.

Requirements:

- visible label for every meaningful field
- required fields indicated consistently
- invalid fields receive a clear error visual state
- field-level error text should appear next to the relevant input where possible
- do not rely only on a generic top-of-form error
- preserve user-entered data after validation failure
- validation should happen at sensible moments; avoid noisy validation while the user is midway through typing
- submit buttons show pending/busy state
- prevent accidental duplicate submission while pending
- failed submission returns controls to usable state
- successful submission produces consistent feedback
- parse/format errors for Rial/Toman should be understandable
- date errors should identify invalid/missing ranges clearly
- percentage/quantity constraints should explain allowed values
- backend/domain errors should be translated into useful operator-facing messages where practical without hiding authoritative errors

Use `aria-invalid`, `aria-describedby`, labels, and accessible focus behavior where appropriate.

When submit fails, focus or scroll to the first actionable invalid field where practical.

## 6. Toast and notification system

Replace ad-hoc notification behavior with one consistent toast system.

Support semantic states:

- success
- error
- warning
- info

Requirements:

- shared visual treatment
- shared placement
- consistent duration
- error/warning messages should remain visible long enough to read
- allow manual dismissal when appropriate
- avoid duplicate identical toasts from one operation
- do not use success toast for failed/partial operations
- important destructive/financial failures should not disappear before the operator can read them
- toast text should state the completed/failed action clearly

Do not use toasts as the only place for field-validation errors.

## 7. Loading, empty and error states

Every async workspace and data surface must explicitly handle:

- initial loading
- refreshing/reloading
- empty data
- query/filter with no results
- backend failure
- partial action failure

Avoid blank panels that look broken.

Standardize loading indicators and skeleton/spinner usage. Keep loading presentation compact; this is desktop ERP software.

Empty-state messages should explain the state and expose the relevant primary action where useful.

Error states should provide a retry action where retry is meaningful.

## 8. Buttons, actions and destructive UX

Audit all action placement and button hierarchy.

Requirements:

- one visually dominant primary action per local context where practical
- secondary actions visually quieter
- destructive actions clearly destructive and separated from routine actions
- destructive confirmation must name the affected entity/action
- financial/reversal/close/restore actions require explicit confirmation where already required by domain rules
- disabled states must be visually obvious and semantically disabled
- pending actions must not be clickable repeatedly
- icon-only buttons need tooltips or accessible labels
- recurring action locations should be consistent between workspaces

Archive and Delete must remain distinct and must never be visually or behaviorally conflated.

## 9. Tables and registers

Unify table/register styling and interactions across the application.

Standardize:

- header height
- row height
- font sizing
- hover state
- selected row state
- borders/dividers
- numeric alignment
- money alignment
- status badge placement
- empty state
- overflow behavior
- sticky header behavior where useful
- row action placement

Money should align consistently and use centralized grouped Rial/Toman formatting.

Dates remain Jalali in user-facing UI.

Avoid multiple unrelated table visual languages.

## 10. Inspectors and editors

Unify all table + inspector/editor experiences.

Requirements:

- shared inspector width strategy
- consistent header and close/back behavior
- consistent section padding/gaps
- consistent footer/action placement
- form controls never clip
- inspector may scroll independently only where deliberate and must not conflict with page scroll/sticky surfaces
- dirty/unsaved state should be clear where edits are staged rather than immediately persisted

## 11. Status and semantic language

Audit badges, colors, wording, capitalization, and status labels.

Requirements:

- same semantic status uses same color treatment everywhere
- do not overload red/green decoratively
- status capitalization/wording should be consistent
- avoid raw backend enum strings when a human label is expected
- warnings, errors, overdue, cancelled, reversed, archived, draft, posted, paid, closed etc. must be visually distinguishable without becoming visually noisy

## 12. Typography and density

Normalize typography throughout the app.

Establish shared styles for:

- page title
- section title
- body text
- secondary metadata
- form label
- field help/error
- table text
- KPI/financial number
- eyebrow/context label

Remove oversized headings and inconsistent weights.

Keep the product compact and information-dense, not spacious marketing UI.

## 13. Focus, keyboard and accessibility basics

This is not a full accessibility certification task, but repair obvious desktop usability gaps.

Requirements:

- visible keyboard focus for interactive controls
- sensible tab order
- no focus trapped behind sticky/floating surfaces
- dialogs/confirmations return focus appropriately where controlled by app components
- icon-only actions have accessible names
- buttons must be actual buttons, not generic clickable elements where avoidable
- form labels correctly associate with fields
- disabled controls expose correct semantics
- status/errors should not depend only on color

Preserve reduced-motion support.

## 14. Error-message normalization

Create a small frontend error-normalization layer if one does not already exist.

It should safely turn Wails/backend errors into operator-facing text while preserving enough detail for troubleshooting.

Do not leak implementation garbage such as `[object Object]`.

Known domain errors such as protected deletion, invalid state transitions, duplicate/idempotent operations, closed-period protection, insufficient stock, reservation conflicts, invalid allocations, or referenced history should receive clear messages where practical.

Unexpected errors may retain a concise technical detail after a human-readable prefix.

## 15. Theme cleanup

Audit the complete CSS/theme implementation.

Remove or consolidate:

- duplicate component styles
- contradictory media queries
- dead styles from earlier mock UI
- arbitrary one-off colors
- arbitrary one-off radii/shadows
- page-local copies of global patterns
- unnecessary inline styles
- stale classes no longer used

Do not change the product to dark theme and do not add theme switching in this task.

## 16. Existing domain formatting rules

Preserve and enforce everywhere:

- canonical backend money: integer Rial
- centralized grouped Rial/Toman UI display/input
- exact 1 Toman = 10 Rial conversion
- Jalali user-facing dates
- canonical persisted dates remain unchanged
- no frontend accounting calculations replacing backend authority

## 17. Verification

Add/update frontend tests where the project has test infrastructure for reusable UI behavior.

At minimum verify important shared utilities/components for:

- money input error state
- normalized backend errors
- toast semantic variants
- duplicate-submit prevention/pending state where practical
- key form validation helpers

Do not add a large new frontend testing framework solely for this task if none exists. Prefer focused tests using current infrastructure.

Manually/source-audit every major workspace for adoption of the unified patterns.

## Acceptance criteria

- All major workspaces visibly belong to one coherent application.
- Global spacing, padding, typography, control sizes, card/panel geometry, tabs, buttons, and tables are consistent.
- Shared layout/sticky/width rules are used rather than page-local approximations.
- No obvious clipped inputs, accidental horizontal form overflow, or content hidden under sticky surfaces at representative desktop widths.
- Forms have consistent invalid/pending/success/failure states.
- Backend failures never render as meaningless raw objects.
- A single coherent toast system supports success/error/warning/info.
- Loading, empty, no-results, and backend-error states are explicit across major data workspaces.
- Destructive and financial actions use consistent confirmation and pending behavior.
- Archive and Delete remain distinct.
- Rial/Toman and Jalali behavior remain centralized and correct.
- No business/accounting authority moves into Vue.
- Frontend production build passes.
- Existing Go tests remain passing.

## Non-goals

- New business features
- Backend/accounting redesign
- Mobile-first redesign
- Dark mode/theme switching
- Brand/logo redesign
- Major dashboard information architecture redesign
- New charts for decoration
- Windows installer validation
- Code signing
- Auto-update

## Validation

Run:

```bash
go test ./...
cd frontend && npm run build
cd .. && git diff --check
```

Run existing frontend tests if configured.

If Wails CLI is available, run an appropriate local Wails build/dev validation. If unavailable, state that explicitly.

Do not claim Windows-native visual validation unless it was actually performed on Windows.

## Handoff

Update `docs/UI.md` and/or `CONTRIBUTING.md` only where needed to document the final reusable UI system and rules established by this task.

Commit to `main`, push `origin/main`, and print the final SHA.
