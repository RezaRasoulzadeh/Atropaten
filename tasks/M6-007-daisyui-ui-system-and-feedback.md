# M6-007 — DaisyUI UI system and feedback unification

## Status

This task supersedes the unfinished intent of M6-004 and M6-005. Do not continue those tasks independently after this task begins.

## Objective

Replace Atropaten's accumulated ad-hoc visual/control styling with a coherent Vue UI system built on Tailwind CSS + DaisyUI, while preserving the existing Wails/Vue application architecture and all domain behavior.

DaisyUI is the styling/component foundation, not permission to apply its defaults blindly. Build one compact Atropaten theme and a reusable application component layer so feature screens do not invent their own buttons, inputs, filter bars, badges, cards, confirmations, toasts, loading states, or error presentation.

The result must remain a dense, modern, Windows-first desktop ERP rather than a generic website or a stock DaisyUI demo.

## 1. Introduce Tailwind CSS + DaisyUI safely

The current frontend is Vue 3 + Vite and does not currently depend on Tailwind or DaisyUI.

- Add compatible Tailwind CSS and DaisyUI dependencies and configure them correctly for the existing Vite frontend.
- Do not replace Vue, Vite, Wails bindings, Jalali utilities, Lucide, or application architecture.
- Keep the build simple and local; do not add a heavyweight UI framework alongside DaisyUI.
- Remove obsolete CSS only after its migrated replacement exists.
- Avoid running old global CSS and DaisyUI against the same controls in ways that cause specificity conflicts.
- Do not keep broad selectors that globally restyle every `input`, `select`, `button`, etc. behind DaisyUI components.

## 2. Create one custom Atropaten DaisyUI theme

Define a named Atropaten light theme using DaisyUI semantic theme variables/tokens.

Visual direction:

- neutral/warm-gray application background
- clean white/subtle work surfaces
- one restrained primary accent
- green/red reserved mainly for semantic success/danger/financial meaning
- subtle borders
- restrained shadow/elevation
- modern flat controls
- moderate radius
- compact density
- approximately 14px normal UI text
- approximately 36–40px common control/table-row height where practical
- no marketing-style oversized headings
- no decorative gradients unless already justified

Explicitly forbidden:

- inset/inner shadows on inputs/selects/textareas
- beveled/recessed controls
- Windows-98-style control treatment
- arbitrary feature-specific color systems
- multiple competing visual themes

## 3. Build reusable application components around DaisyUI

Do not scatter large DaisyUI class strings independently across every feature screen. Create a small reusable component layer that owns the application's visual contract.

At minimum provide/refactor reusable components or equivalent primitives for:

- `AppButton`
- `IconButton`
- `AppInput`
- `AppSelect`
- `AppTextarea`
- `FormField`
- `FieldMessage`
- `MoneyInput` integration
- `QuantityInput` integration
- `DateInput` / Jalali-aware field integration where existing behavior permits
- `SearchFilterBar`
- `FilterField`
- `DataTable` shell/primitives
- `StatusBadge`
- `AppPanel` / section surface
- `WorkspaceHeader`
- `WorkspaceTabs`
- `WorkspaceStickyStack`
- `WorkspaceBottomActions`
- `EmptyState`
- `LoadingState` / spinner
- `InlineAlert`
- `ConfirmDialog`
- `ToastHost` and toast API/service

Do not over-componentize trivial markup. The point is that equivalent controls and states must have one implementation and one visual contract.

## 4. Unify search/filter toolbars everywhere

The compact existing search/filter row previously identified by review is the reference behavior:

- search field consumes remaining width
- filters sit beside it with consistent control height
- result count aligns cleanly on the same row
- toolbar has one coherent surface, padding, border, radius, and spacing
- no accidental internal shadows
- no random page-specific label geometry

If a filter/input requires a visible label:

- label must be smaller than the input/select value text
- label is secondary/muted, never dominant
- do not use large uppercase labels
- keep vertical space compact

Create one reusable `SearchFilterBar`/filter pattern and migrate every major workspace to it instead of maintaining independent toolbar CSS.

## 5. Form and control visual rules

All standard inputs/selects/textareas must use DaisyUI-backed shared components with a flat modern appearance.

Requirements:

- same height for equivalent controls
- same border/radius/background/padding
- consistent hover/focus/disabled/readonly/invalid states
- subtle focus ring
- no inner shadow
- no clipping outside inspectors/panels
- labels consistently smaller/secondary to entered values
- helper/error text uses a compact smaller size
- two-column forms collapse before controls clip
- `min-width: 0` and box sizing are correct throughout flex/grid layouts

Audit special money/quantity/date controls so they visually match the same system without losing their existing domain behavior.

## 6. Buttons and icon alignment

Use shared DaisyUI-backed button variants:

- primary
- secondary
- ghost/subtle
- danger
- icon-only
- compact/table action

All buttons must have consistent:

- height
- padding
- radius
- icon size
- icon/text gap
- loading state
- disabled state
- focus state

Lucide icon and button text must render as one aligned group. Do not allow the broken appearance where an icon is visually detached from its text.

## 7. Panels, cards and empty states

Use surfaces intentionally.

- Do not wrap every piece of content in a rounded box merely because DaisyUI provides cards.
- Keep panels compact and information-dense.
- Empty states should be compact and operational.
- Icon + message + optional action should be one aligned composition.
- Remove oversized dashed placeholder boxes unless the border communicates an actual drop zone or editable region.
- Avoid decorative empty-state illustrations.

Migrate the Parameters/Cost Components empty states and similar screens to the shared pattern.

## 8. Tables, inspectors and dense workspace layout

Unify tables through shared primitives:

- header geometry
- row density
- selected/hover state
- sort affordance
- money/numeric alignment
- status badges
- horizontal overflow only for genuinely wide data
- loading/empty/error representation

Unify inspector/register layouts:

- register consumes available width
- inspector has deliberate min/default/max width
- no fields outside the inspector
- no large unused workspace region caused by arbitrary fixed widths
- narrow desktop windows degrade gracefully

Preserve shared sticky page/header/tab/bottom-action geometry.

## 9. Toast system is mandatory and global

This is a hard product requirement: user-visible application outcomes must not be silent.

Implement one global reusable toast system using DaisyUI visual primitives and a central Vue API/store/composable.

Support exactly these semantic types:

- success
- error
- warning
- info

Requirements:

- one `ToastHost` mounted at the application shell level
- one central API/composable used by feature screens
- consistent placement and stacking
- concise message/title hierarchy
- Lucide semantic icons
- manual dismiss
- appropriate timeout by severity
- errors/warnings remain long enough to read
- duplicate rapid-fire messages deduplicate/collapse where reasonable
- multiple concurrent messages remain readable
- toast must not cover persistent bottom actions or status surfaces
- accessibility live-region semantics where practical

### No silent outcomes

Every user-triggered asynchronous operation must report its outcome through toast unless the action is purely transient/local and has no meaningful outcome.

At minimum:

- successful create/save/update => success toast
- successful delete/archive/reactivate => success toast
- successful post/void/reverse/cancel/close/restore/backup => success toast
- failed operation => error toast
- backend/domain warning surfaced to the user => warning toast
- informational operation/system notice => info toast

There must be no swallowed promise rejection, `console.error`-only failure, or silent catch on a user-visible workflow.

A page-blocking/load error must still appear inline in context, but it ALSO gets an error toast when first encountered. Toast does not replace persistent contextual error UI.

Do not spam toasts for passive normal rendering, every table refresh, hover, or expected validation keystroke.

## 10. Central error normalization

Create one central frontend error-normalization path for Wails/backend errors.

It must:

- accept unknown thrown values
- extract a stable readable operator-facing message
- preserve useful technical detail separately when needed
- recognize reliable typed/category information if available
- provide safe fallbacks
- map known protected-delete/not-found/conflict/validation cases to friendly copy where reliable
- avoid repeating `String(e)` and ad-hoc catch behavior throughout pages

Do not parse unstable backend English strings into new business logic. Narrow backend error typing may be added if genuinely needed for reliable UI classification, but domain authority stays in Go.

## 11. Field validation and form feedback

Reusable fields must support:

- required
- invalid
- disabled
- readonly
- help text
- inline field error

Rules:

- obvious shape validation may happen in Vue
- domain/accounting/inventory/lifecycle validation remains backend-authoritative
- preserve entered values after failed submit
- clear stale errors after correction
- focus/scroll to first invalid field when practical
- submit buttons enter loading state and prevent duplicate submission
- backend form failure gets contextual error + toast

## 12. Confirmation dialogs

Replace major browser `window.confirm` usage with one reusable DaisyUI-backed `ConfirmDialog`.

Use it for destructive/high-impact actions such as:

- Delete
- Cancel
- Void
- Reverse
- fiscal close
- backup restore
- other destructive financial/state transitions

Dialog requirements:

- identify affected record when practical
- action-specific consequence copy
- clear danger styling for destructive action
- safe cancel/default focus
- Escape support when safe
- keyboard operable
- no ambiguity between Archive and Delete

## 13. Loading, empty, unavailable and error states

Create shared patterns and migrate major screens.

Distinguish clearly:

- initial loading
- background refresh
- submitting
- empty domain
- no search results
- prerequisite/unavailable state
- page error
- operation error

No blank screen during ordinary async work. Do not confuse empty data with failed loading.

## 14. Audit and migrate every major workspace

Do not stop after implementing shared components.

Audit/migrate:

- Dashboard
- Orders
- Quotes/configurators
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
- print-preview entry surfaces where applicable

For every workspace verify:

- DaisyUI/shared components actually used
- search/filter toolbar consistency
- field/label hierarchy
- buttons/icon alignment
- forms and inspectors fit
- no inset shadows
- table density
- status badges
- responsive desktop width behavior
- empty/loading/error states
- destructive confirmations
- success/error/warning/info toast coverage for user actions
- grouped Rial/Toman preserved
- Jalali user-facing dates preserved

## 15. Remove conflicting legacy UI CSS

Once screens are migrated:

- delete obsolete global UI rules that conflict with Tailwind/DaisyUI
- remove old broad input/select/button styling
- remove duplicated page-specific component styling now owned by shared primitives
- preserve only CSS still needed for application shell geometry, print layouts, specialized domain components, or behavior not sensibly expressed through Tailwind/DaisyUI
- avoid `!important` except where a documented unavoidable integration boundary requires it

Do not solve conflicts by stacking another layer of overrides on top of M6-004 CSS.

## 16. Preserve application behavior

Do not change business/domain semantics to simplify UI migration.

Do not move authoritative calculations/status transitions into Vue.

Preserve:

- accounting behavior
- inventory behavior
- production behavior
- pricing behavior
- migration/schema behavior
- backup/restore behavior
- immutable history
- deletion/archive distinctions
- Jalali date handling
- grouped Rial/Toman handling

## 17. Demo data compatibility

M6-006 provides/targets realistic development demo data. Ensure the redesigned UI works with dense populated data, long values, varied statuses, large monetary amounts, and empty/no-result cases.

If M6-006 is already implemented, use its demo dataset during visual inspection. If not, do not block this task on it, but keep the UI compatible with that task.

## 18. Documentation

Update `docs/UI.md` to document the implemented DaisyUI-based system, including:

- Atropaten DaisyUI theme
- shared application components
- density/control sizing
- toolbar/filter pattern
- label hierarchy
- toast semantics and mandatory feedback rule
- field validation/error hierarchy
- confirmation behavior
- loading/empty/error patterns
- permitted remaining custom CSS

Create `docs/DAISYUI_MIGRATION_AUDIT.md` documenting:

- screens migrated
- shared components introduced
- legacy CSS removed
- remaining custom CSS and why
- toast coverage audit
- any remaining visual/native Windows checks

## Acceptance criteria

This task is complete only when:

1. Tailwind CSS + DaisyUI are installed/configured and the frontend builds cleanly.
2. One custom Atropaten DaisyUI theme is the visual foundation.
3. Major screens use reusable shared application components rather than independent ad-hoc styling.
4. Standard inputs/selects/textareas have no inset/inner shadow or beveled appearance.
5. Search/filter bars are consistent across major workspaces.
6. Labels are visually secondary and smaller than control values.
7. Icon/text alignment is correct in buttons and empty states.
8. Tables/forms/panels/inspectors follow one density and geometry system.
9. All four toast types exist: success/error/warning/info.
10. User-triggered async failures are never silent.
11. User-triggered meaningful successes are never silent.
12. Warnings/info outcomes that require operator awareness use toast.
13. Page-blocking errors remain contextual in addition to toast.
14. A central error-normalization utility replaces scattered raw error rendering/catches.
15. Major destructive workflows use the shared confirmation dialog instead of browser confirm.
16. Loading/submitting prevents duplicate actions and distinguishes empty/error states.
17. No broad legacy CSS competes with DaisyUI on migrated controls.
18. Grouped Rial/Toman and Jalali presentation remain intact.
19. No accounting/inventory/production/pricing/persistence authority moves to Vue.
20. Every major workspace is included in the migration audit.
21. Go tests, frontend production build, and `git diff --check` pass.

## Validation

Run:

```bash
go test ./...
cd frontend
npm run build
cd ..
git diff --check
```

Run any existing frontend tests if available.

If Wails CLI is available, run an appropriate local Wails validation. If unavailable, state that explicitly.

Do not claim Windows/WebView native validation unless actually performed on Windows.

Visual verification is required where the available environment permits it; a successful build alone is not proof that the migration looks correct.

## Delivery

Commit directly to `main`, push `origin/main`, and print the final SHA.