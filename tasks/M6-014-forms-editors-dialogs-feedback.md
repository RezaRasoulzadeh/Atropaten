# M6-014 — Forms, editors, dialogs, and feedback unification

## Objective
After geometry, typography/control metrics, and data surfaces are stable, unify every create/edit/posting workflow so forms and interaction feedback behave and look like one application.

## Form system
Use shared `FormField`, `FormGrid`, `FormSection`, `AppInput`, `AppTextarea`, `SelectField`, Jalali controls, money/quantity helpers, and layout primitives.

Standardize:
- label placement and size;
- help/error text placement;
- field spacing;
- required/optional treatment;
- responsive 1/2/3-column field grids;
- full-width fields only when content warrants it;
- action placement;
- create/edit/view mode transitions;
- disabled/read-only presentation;
- preservation of entered values on validation failure;
- first-invalid focus where practical;
- duplicate-submit prevention and busy state.

Do not use placeholder text as the only label for important fields.

## Workflow coverage
Explicitly inspect and repair the main interactive editors/forms:
- New Order and existing Order editor;
- New Quote / Quote editor;
- Customer create/edit;
- Material create/edit and stock adjustment;
- Service definition, parameters, cost components, pricing/configurator;
- Machine create/edit;
- Supplier create/edit;
- Purchase editor and line items;
- Production job actions/consumption/waste/outsourcing;
- Accounting payment posting;
- Invoices;
- Expenses;
- Treasury/transfers;
- Checks and lifecycle actions;
- Loans, schedules and payments;
- Owners and fiscal/profit actions;
- Reports filters/preview input;
- Settings and backup/restore.

Break oversized feature components further where doing so materially improves ownership and visual consistency. Do not split merely to reduce line count.

## Dialogs and destructive actions
All major confirmation actions must use the shared `ConfirmDialog`.

Verify:
- correct focus on open;
- Escape/cancel behavior;
- keyboard containment/focus return;
- clear consequence text;
- danger styling only for destructive actions;
- archive/delete/reverse/void/close/restore semantics are not visually conflated.

No `window.confirm`.

## Toasts/errors
Preserve and complete centralized feedback:
- meaningful successful user action -> success toast;
- failed user action -> error toast;
- warning/info requiring attention -> appropriate toast;
- passive data refresh should not spam toasts;
- page-blocking load errors may remain inline;
- no swallowed promise rejection or console-only failure;
- normalize unknown Wails/backend errors centrally.

Audit every async user action for feedback and busy state.

## Popovers and specialized controls
Verify Jalali calendar, selects, dialogs, configurators and other overlays:
- no clipping;
- correct z-index;
- logical keyboard behavior;
- visual alignment with the shared form system;
- no accidental width overflow.

## Visual verification
Use M6-006 demo data and browser automation at 1024/1280/1600.

For representative forms test:
- initial state;
- focused field;
- validation failure with values preserved;
- submitting state;
- success toast;
- backend error toast;
- confirmation dialog;
- long labels/values;
- narrow-width wrapping.

Update `docs/FRONTEND_UI_AUDIT.md` with interaction-state coverage per page.

## Constraints
- Preserve previous geometry, typography, and data-surface contracts.
- Preserve backend/domain authority and accounting/inventory semantics.
- No new styling framework or global feature CSS.

## Acceptance
- Major forms use a coherent shared system.
- No important form relies on placeholder-only labeling.
- No silent meaningful user action remains.
- Busy, error, success, confirm and validation states are consistent.
- Specialized controls/popovers are not clipped.
- Actual rendered interaction evidence exists.
- `go test ./...`, `cd frontend && npm run build`, `git diff --check` pass.
- Commit/push to main and report SHA.