# M6-010 — Orders creation, Customers, and Materials UI rebuild

## Objective
Rebuild three related frontend contexts to a production-quality visual baseline using the now-stable dark Amber DaisyUI foundation:

1. New Order / Order creation and editing workflow
2. Customers workspace
3. Materials workspace

This is intentionally larger than a single-page polish task but still bounded. Do not touch unrelated workspaces except for shared primitives extracted because these three screens need them.

Dashboard and the Orders main register are the visual references. Preserve their density, hierarchy, neutral surfaces, Amber accent behavior, typography, control focus behavior, and restrained use of borders.

## Global visual rules

- Dark neutral surfaces. No Amber borders/shadows/glows on cards, panels, tables, inspectors, toolbars, dialogs, or containers.
- Amber only for primary actions, active state emphasis, links, selected navigation/tabs, and focused editable controls.
- Inputs, textareas, selects and custom select triggers: 1px solid neutral border at rest, 1px dashed Amber border on focus, no outline/ring/shadow/glow/layout shift.
- Vazirmatn remains the global font using its natural metrics; no metric overrides, vertical transforms, negative margins, asymmetric padding hacks, or local typography compensation.
- Use DaisyUI defaults for standard controls and Tailwind for page geometry. Do not add broad new global CSS.
- Preserve all existing toast, confirmation, error normalization, Jalali date, Rial/Toman, quantity, Wails and backend/domain behavior.

## Shared register/table direction

Use Orders main register as the reference language, not as a literal row template.

Create or refine shared primitives needed by these three contexts so future pages can reuse them:

- `RegisterList` / `RegisterRow` style primitives for rich entity rows with primary identity, secondary metadata, status badges, actions, selection, hover, keyboard activation and neutral row dividers.
- `DataTable` primitives for genuinely tabular numeric data with consistent density, typography, border treatment, header hierarchy, alignment, selection and horizontal overflow.
- Shared inspector/editor panel structure with explicit header, content, actions and responsive width behavior.
- Shared form section / field grid patterns for compact desktop forms.

Do not over-abstract. Extract only patterns actually shared by these three contexts. Avoid giant configurable components and avoid page-specific CSS classes pretending to be reusable.

## Context 1 — New Order / order creation workflow

The current Orders main register is acceptable; do not redesign it except for shared-component integration where needed.

The New Order / order editor workflow is currently considered unusable and must be rebuilt.

Inspect the actual order creation/editing code path, including `OrderWorkspaceView.vue`, configurator usage, item editing, customer selection, commercial/fulfillment/payment/priority metadata, promised date, notes, totals, validation, save/confirm/cancel actions and related subpanels.

Requirements:

- Establish a clear workspace hierarchy: page header, order identity/status, customer/order metadata, item list, item configurator/editor, totals/summary and persistent actions.
- The workflow must be understandable without guessing which panel is primary.
- New-order mode and existing-order mode must have deliberate but consistent layouts.
- Customer selection must be compact and readable; walk-in state must be obvious.
- Order metadata controls must use shared field geometry and start-aligned selects.
- Item list should use a readable register/table pattern depending on the data. Each item needs clear service identity, configuration summary, quantity, unit/line price and edit/remove affordances.
- Configurator/editor must not visually compete with the item list. Use sections/tabs/collapsible structure only where it reduces complexity.
- Totals, override/margin warnings and pricing information must be visually grouped and numerically aligned.
- Primary action must be obvious. Secondary/destructive actions must be visually subordinate.
- Validation errors must appear near the relevant field/section and also preserve the existing toast/error contract where appropriate.
- Prevent clipping, hidden bottom actions, accidental nested scrolling and unusable narrow columns.
- Keep existing business rules and backend authority unchanged.

## Context 2 — Customers workspace

Replace the current raw/unstructured list + inspector implementation with the shared register/inspector pattern.

Requirements:

- Use `WorkspaceHeader` + shared `SearchFilterBar` geometry consistent with Orders.
- Customer list should use rich register rows, not a generic zebra table.
- Each row should expose name as primary identity, contact summary, active/archive state and optionally one useful secondary detail without clutter.
- Clear hover, selected and keyboard-focus states.
- Right-side inspector on desktop with explicit min/max width and `min-width: 0`; collapse below the register at narrower widths rather than clipping.
- Inspector detail view and create/edit view must share the same shell.
- Form fields grouped logically: identity, contact, address, notes.
- Primary Save action uses primary styling. Archive/reactivate and Delete are clearly separated; Delete remains protected by confirmation and backend constraints.
- Loading, error, empty, no-results and submitting states use shared components rather than raw text blocks.
- Preserve customer data and behavior exactly.

## Context 3 — Materials workspace

Rebuild the current oversized Materials workspace into smaller visual components while preserving all inventory behavior.

Requirements:

- Split the monolithic page by responsibility where useful: register, inspector/summary, material editor, stock movement history, stock adjustment.
- Use the shared toolbar/header system.
- Material register should use the new shared `DataTable` because physical/available/reserved quantities, average cost, inventory value and reorder status benefit from aligned numeric columns.
- Do not use zebra striping. Use neutral row dividers, restrained hover/selected backgrounds and the same text hierarchy as Orders.
- Primary material identity should still be easy to scan: name + SKU/category, with secondary metadata visually subordinate.
- Numeric columns must align consistently; money uses existing grouped formatter and quantity formatting remains authoritative.
- Status/low-stock badges must remain compact.
- Inspector should summarize stock, units, cost basis, supplier and notes without becoming a long raw text wall.
- Editor forms should be split into compact logical sections rather than one uninterrupted form.
- Stock adjustment must be visually distinct from catalog editing because it creates immutable ledger movements.
- Movement history must use a clear data-table/list presentation with date, type, quantity, cost/value and note/source where available.
- Preserve reservation/physical/available semantics, immutable movement behavior, weighted-average costing and backend authority.

## Architecture boundaries

- Prefer extracting components into sensible feature folders for these contexts, e.g. `features/orders`, `features/customers`, `features/materials`, plus `components/ui` for genuinely shared primitives.
- Do not perform a full-repo folder migration in this task.
- Move only files touched by these contexts when doing so improves ownership and readability.
- Avoid creating wrappers that only rename one DaisyUI class with no behavioral/layout value.

## Visual verification

This task is not complete from source inspection or a successful build alone.

Where runtime access exists:

- Use M6-006 demo data.
- Render and inspect at approximately 1024, 1280 and 1600 px widths.
- Explicitly inspect:
  - New Order before adding items
  - New Order with multiple configured items
  - Customer register + selected inspector
  - Customer create/edit mode
  - Materials populated register
  - Material selected inspector
  - Material edit mode
  - Stock adjustment and movement history
  - focus treatment for inputs/selects/textareas
  - long names/notes, large money values and decimal quantities
- Do not claim Windows/WebView/Wails-native validation if unavailable.

Record screenshots or concrete visual observations in `docs/M6-010-UI-AUDIT.md`. For each of the three contexts, record what was actually rendered, remaining defects, and any deferred redesign.

## Non-goals

- Do not redesign Dashboard or Orders main register beyond shared-component integration.
- Do not redesign Quotes, Production, Services, Machines, Purchases, Suppliers, Accounting, Invoices, Checks, Loans, Owners, Reports or Settings yet.
- Do not add business features, database migrations, new accounting logic or inventory behavior.
- Do not add a second UI framework or global CSS compatibility layer.

## Acceptance criteria

- New Order / order editing is visually usable and coherent at representative desktop widths.
- Customers is rebuilt around the shared register + inspector/editor pattern.
- Materials is rebuilt around the shared data-table + inspector/editor + movement pattern.
- Orders main register remains visually intact.
- Shared register/table/inspector/form primitives are reusable without becoming over-generalized.
- No zebra tables in these rebuilt contexts.
- No Amber card/container borders, shadows or glows.
- Editable focus contract remains dashed Amber border with no outline/ring/shadow.
- No font metric hacks are reintroduced.
- Existing domain behavior and backend contracts remain unchanged.
- `go test ./...`, `cd frontend && npm run build`, and `git diff --check` pass.
- Run any available frontend tests; if no frontend test script exists, state that explicitly.
- Update `docs/UI.md` only where shared UI conventions changed and create `docs/M6-010-UI-AUDIT.md`.
- Commit and push to `origin/main`; report final SHA, extracted shared primitives, page/component changes, validation, visual checks and remaining limitations.
