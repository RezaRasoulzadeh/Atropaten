# M6-004 — UI system unification and layout hardening

## Objective

Unify Atropaten’s visual system and desktop workspace geometry across the entire product. This task is strictly about layout, theme, spacing, sizing, density, shared visual primitives, responsive desktop behavior, and removal of accumulated one-off CSS/layout inconsistencies.

Do not redesign the product from scratch and do not change business/accounting/inventory semantics.

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

## 1. Establish one semantic visual token system

Consolidate shared CSS variables/tokens for:

### Color

- application background
- surface / raised surface / subtle surface
- border / divider
- primary / secondary / muted text
- accent / hover / pressed
- focus ring
- success / warning / danger / info
- disabled states
- hover / selected states

Use semantic tokens. Feature components should not invent arbitrary color values when an existing token fits.

Keep green/red primarily semantic.

### Spacing

Define one compact spacing scale used consistently for:

- workspace gutters
- panel/card padding
- section gaps
- form rows
- labels/controls
- tables
- buttons
- toolbars
- tabs
- inspectors

Remove near-duplicate arbitrary spacing values that create visible inconsistency.

### Typography

Centralize:

- font family
- base text size
- metadata size
- labels
- table text
- page titles
- section headings
- numerical/financial emphasis
- line heights / weights

Preserve the dense desktop ERP direction. No oversized marketing-style headings.

### Shape/elevation

Centralize:

- border-radius scale
- shadow/elevation scale
- separators and borders

### Control sizing

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

Equivalent controls must have equivalent heights everywhere.

## 2. Unify global workspace geometry

Apply the documented workspace/sticky rules consistently:

- one workspace gutter system
- one page-header pattern
- one coordinated sticky-header/tab pattern
- one bottom-action pattern
- one z-index hierarchy
- consistent sticky border/background/elevation
- no unexplained top gaps
- no page-specific fixed positioning when shared sticky primitives suffice
- no content hidden under sticky surfaces
- no accidental nested vertical scrolling

Refactor shared primitives rather than patching each page independently where possible.

## 3. Width, padding, margin, overflow and responsive desktop cleanup

Audit every major page at:

- narrow laptop-like desktop
- normal desktop
- wide desktop

Fix:

- unused workspace width
- table/register panes that stop too early
- bad inspector widths
- clipped controls
- controls outside panels
- accidental horizontal scroll in forms
- bad two-column form collapse
- inconsistent page/panel padding
- inconsistent margins/gaps
- long labels/values/badges overflow
- sticky overlap
- hidden bottom action surfaces
- table overflow without intentional scrolling

Use `min-width: 0`, border-box sizing, deliberate min/max widths, and responsive grid/flex behavior consistently.

Mobile redesign is out of scope.

## 4. Standardize shared visual primitives

Consolidate and migrate shared components/styles for:

- WorkspaceHeader
- WorkspaceTabs
- WorkspaceStickyStack
- WorkspaceBottomActions
- SectionPanel
- AppButton / icon button variants
- form controls
- StatusBadge
- table shell
- inspector shell
- toolbar/filter rows
- compact cards

Do not over-componentize trivial markup. The goal is visible consistency and removal of repeated conflicting CSS.

## 5. Forms and controls — visual consistency only

Standardize visual treatment for:

- text input
- numeric input
- money input
- quantity input
- select
- textarea
- checkbox/toggle
- Jalali date/date-time picker
- grouped fields

Ensure:

- one control height system
- consistent label spacing
- consistent focus state
- consistent disabled/readonly appearance
- consistent required marker treatment
- all controls fit their panels
- two-column layouts collapse before clipping

Detailed validation/error UX belongs to M6-005.

## 6. Tables and density

Unify operational tables:

- header height
- row density
- cell padding
- typography
- numeric/money alignment
- selected/hover state
- status badge placement
- toolbar geometry
- horizontal overflow behavior

Money remains grouped consistently. User-facing dates remain Jalali.

## 7. Buttons, badges, actions

Standardize button variants:

- primary
- secondary
- ghost/subtle
- danger
- icon-only
- compact table action

Standardize height, padding, icon size, gap, focus, hover, disabled, and pressed states.

Consolidate status badge sizing, radius, weight, spacing, and semantic visual grammar across all domains.

## 8. Theme consistency

Keep the release primarily light-themed:

- neutral/warm gray shell background
- white/subtle surfaces
- restrained shadows
- subtle borders
- one accent
- minimal gradients

Remove local styles that visibly contradict the system unless there is a domain-specific reason.

## 9. Accessibility basics related to layout/components

Fix obvious shared-component issues:

- visible keyboard focus
- labels associated with fields
- icon buttons have accessible labels/tooltips
- disabled state visually clear
- buttons use actual button elements
- sufficient contrast
- reduced-motion preserved

Full UX/error-state behavior is M6-005.

## 10. Audit every major screen

Do not stop after introducing tokens/components. Migrate every major workspace and verify:

- page header geometry
- tabs
- primary/secondary actions
- page padding
- panel padding
- form alignment
- control heights
- table density
- inspector width
- responsive desktop behavior
- no clipping/overflow
- no accidental unused width
- no sticky overlap
- grouped Rial/Toman
- Jalali dates

## 11. Preserve behavior

Do not:

- change accounting semantics
- change inventory semantics
- change posting/idempotency behavior
- move authoritative calculations into Vue
- rewrite backend domain behavior to simplify layout

## 12. Documentation

Update `docs/UI.md` so the implemented design system is explicit:

- token philosophy
- spacing scale
- typography hierarchy
- control sizes
- button variants
- panel/table/inspector geometry
- sticky/workspace rules
- responsive desktop acceptance widths

Create `docs/UI_LAYOUT_AUDIT.md` documenting:

- screens audited
- visual primitives introduced/refactored
- layout inconsistencies fixed
- remaining manual Windows/WebView layout checks
- intentionally deferred visual issues

## Acceptance criteria

M6-004 is complete only when:

1. Major workspaces visibly share one layout/spacing/sizing/theme system.
2. Page headers, tabs, panels, tables, forms and inspectors use consistent geometry.
3. Equivalent controls have equivalent sizes everywhere.
4. No major page has obvious unused-width, clipping, overflow or sticky-overlap defects at representative desktop widths.
5. One semantic token system is used across the product instead of feature-local visual values.
6. Tables use consistent density and alignment.
7. Buttons/badges use shared variants and dimensions.
8. Money/date presentation remains grouped and Jalali.
9. No accounting/inventory/production/reporting authority moves into the frontend.
10. `docs/UI.md` reflects the implemented design system.
11. `docs/UI_LAYOUT_AUDIT.md` records the full pass.
12. Existing Go tests pass.
13. Frontend production build passes.
14. `git diff --check` passes.

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
