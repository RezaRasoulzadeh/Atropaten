# M6-013 — Data surfaces: registers, tables, inspectors

## Objective
After geometry and typography/control metrics are stable, unify every list/register/table/detail surface around two deliberate patterns: rich entity registers and dense data tables.

## Reference
The existing Orders main register is the visual reference for rich rows: strong identity, muted secondary metadata, compact statuses, neutral dividers, restrained Amber, readable spacing, no zebra striping.

Do not copy its exact multi-line row structure everywhere. Use the pattern appropriate to each dataset.

## Pattern A — Rich register
Use for entity-centric queues and relationships such as Orders, Quotes, Customers, Production jobs, Suppliers where scanning entity identity/status is more important than column comparison.

Shared contract:
- neutral dividers, no card-per-row noise;
- clear primary identity;
- secondary metadata below/beside it in muted text;
- compact status badges;
- important amount/date/status aligned consistently;
- hover, keyboard focus and selected state consistent;
- whole-row activation semantics where appropriate;
- long text truncates/wraps intentionally;
- empty/loading/error states use shared primitives.

## Pattern B — Dense data table
Use for Materials, Machines, Purchases, Accounting ledger/accounts, Invoices, Checks, Loans, Owners finance tables, Reports and other column-comparison surfaces.

Shared contract:
- semantic table markup;
- no zebra striping;
- neutral horizontal dividers;
- compact readable row density;
- table headers visually subordinate but clear;
- numeric and money cells aligned consistently;
- quantities/dates/status columns sized intentionally;
- horizontal scroll only where genuinely necessary;
- sticky header only when useful and non-breaking;
- selected/hover/keyboard states consistent;
- primary/secondary text hierarchy within cells when needed.

## Inspectors/details
Unify detail views through `MasterDetail`, `InspectorShell`, `InspectorHeader`, and `InspectorSection` where appropriate.

Requirements:
- consistent inspector width and section spacing from M6-011;
- consistent identity/status/action header;
- no raw stacks of unrelated `<div>` blocks;
- read-only values and edit forms have clear hierarchy;
- lifecycle/destructive actions are separated from ordinary edits;
- history/movement/schedule sections use the shared table/register language.

## Migration scope
Review and migrate every major context. At minimum explicitly inspect:
- Orders (preserve accepted baseline);
- Quotes;
- Production;
- Customers;
- Services catalog/list;
- Materials;
- Machines;
- Purchases;
- Suppliers;
- Accounting Accounts/Journal/Payments;
- Invoices;
- Expenses/Transfers;
- Checks;
- Loans;
- Owners;
- Reports.

Do not leave a page using old zebra tables or ad-hoc row stacks just because it compiles.

## Shared components
Strengthen `RegisterList`, `RegisterRow`, `DataTable`, `DataTableRow`, `DataTableCell`, `StatusBadge`, and inspector primitives only as needed. Keep them thin and typed. Avoid a giant schema-driven table framework.

Delete dead/duplicate list/table wrappers after migration.

## Visual verification
Use M6-006 populated data and render all affected pages at 1024/1280/1600.

Check:
- long names;
- large Rial/Toman values;
- six-decimal quantities;
- many statuses;
- selected rows;
- hover/focus;
- table overflow;
- empty/no-result states;
- inspector open/closed states.

Update `docs/FRONTEND_UI_AUDIT.md` with which pattern each page uses and actual rendered status.

## Constraints
- Preserve M6-011 geometry and M6-012 typography/control metrics.
- Preserve backend/domain behavior.
- No new feature-wide global CSS.
- No page-specific visual system forks.

## Acceptance
- Every list-heavy page uses either the approved rich-register or dense-table language.
- No accidental zebra tables remain.
- Numeric/money comparison surfaces are aligned and readable.
- Inspectors are structurally consistent.
- Empty/loading/error/selection states are unified.
- Real screenshots cover every affected workspace.
- `go test ./...`, `cd frontend && npm run build`, `git diff --check` pass.
- Commit/push to main and report SHA.