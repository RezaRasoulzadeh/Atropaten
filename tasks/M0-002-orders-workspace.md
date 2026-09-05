# M0-002 — Build the Orders list and order workspace mock

## Objective

Extend the M0 desktop foundation with the first complete operational screen flow: an Orders list and an Order workspace that can be reached and exercised from the existing Atropaten shell.

This remains frontend-only mock functionality. The purpose is to establish the reusable interaction patterns and let the product flow be visually tested before persistence and domain implementation begin.

## Scope

### Orders list

Replace the current Orders placeholder with a polished, dense desktop Orders screen using representative mock data.

Include:

- Page toolbar/header with `New order` primary action
- Search/filter controls appropriate for orders
- Filters for at least commercial status, fulfillment status, payment status, and priority
- Dense sortable-looking order table with representative rows
- Columns covering order number, customer, created date, promised date, total, balance, fulfillment status, payment status, and priority where practical
- Clear overdue/late visual treatment without excessive color
- Row selection and opening an order workspace
- Useful empty/filter state where applicable

The three order state axes must remain visibly separate:

- Commercial state
- Fulfillment state
- Payment state

Do not collapse them into one generic status.

### Order workspace

Opening a representative order must show a proper order workspace rather than a modal.

Provide tabs for:

- Overview
- Items
- Production
- Payments
- Files
- History

Implement enough representative mock content in each tab to establish the intended layout and interaction pattern. The Overview and Items tabs should receive the most detail in this task.

The workspace should visibly include:

- Order number
- Customer
- Created date
- Promised date
- Priority
- Commercial / fulfillment / payment status
- Notes
- Totals summary
- Deposit / paid amount
- Remaining balance
- Representative order items

### Order items

Show that one order can contain unrelated services. Use at least two different example item types in the same order.

Each order item should expose enough information to preview future pricing/production integration, such as:

- Service name
- Key parameters/specification
- Quantity
- Selling price
- Estimated cost
- Profit or margin
- Production status

Do not implement the real pricing engine yet.

### New order interaction

Make the existing dashboard `New order` action and the Orders page `New order` action lead to a visible frontend-only new-order workspace or draft state instead of the current toast-only placeholder.

The draft interface should establish the basic composition for later M2 work without implementing persistence.

### Reusable UI patterns

Extract/reuse only clearly useful components and styles. This task should establish or improve reusable patterns for:

- Page toolbar
- Filter controls
- Dense data tables
- Tabs
- Detail/summary sections
- Status badges
- Form fields
- Empty state

Do not build a generic component framework.

## Visual requirements

Follow `docs/UI.md` and the existing M0 shell.

- Maintain current desktop density and visual language.
- Avoid oversized cards and large whitespace.
- Prefer a full-page workspace over nested modal dialogs.
- Order list should feel like operational business software, not a consumer web app.
- Keep financial numbers aligned and easy to scan.
- Preserve the existing shell/sidebar/status-bar behavior.

## Constraints

- Vue 3 + TypeScript only for this task.
- Use static/mock frontend data.
- No SQLite or persistence.
- No Go business/domain implementation.
- No router unless it is clearly justified; local application state is acceptable for M0.
- No external UI framework.
- Do not implement real pricing, inventory consumption, invoice posting, accounting, or payment allocation.
- Do not redesign the accepted M0 shell unless a small change is required to support this workflow.

## Acceptance criteria

1. Orders navigation opens a real Orders list rather than the generic placeholder.
2. Representative orders can be searched/filtered locally.
3. Commercial, fulfillment, and payment states are displayed independently.
4. Clicking/opening a representative order shows the order workspace.
5. Order workspace provides Overview, Items, Production, Payments, Files, and History tabs.
6. A representative order contains at least two unrelated service/item types.
7. Dashboard `New order` opens a visible new-order draft/workspace instead of only showing a toast.
8. Orders page `New order` opens the same draft flow.
9. The UI introduces reusable tabs/form/detail patterns suitable for later screens.
10. Existing dashboard and shell continue to work.
11. No persistence or authoritative business logic is introduced.
12. `npm run build` succeeds.
13. `git diff --check` succeeds.

## Manual visual validation

Run Atropaten and manually verify this flow:

1. Open Dashboard.
2. Click `Orders` in the sidebar.
3. Use at least one order filter/search control and verify the visible rows change.
4. Open an existing order.
5. Inspect every workspace tab.
6. Return to the Orders list.
7. Start a new order from the Orders screen.
8. Return to Dashboard and confirm its `New order` action opens the same draft flow.
9. Collapse/expand the sidebar while inside the Orders workflow.
10. Resize the window to normal desktop/laptop sizes and confirm there is no page-level horizontal overflow.

## Validation

```sh
cd frontend
npm run build
cd ..
git diff --check
```

Then run Wails locally for manual visual validation. On Linux with WebKitGTK 4.1, use the required Wails build tag.

## Relevant documentation

- `docs/UI.md`
- `docs/PRODUCT.md`
- `docs/DOMAIN.md`
- `docs/ARCHITECTURE.md`
- `MILESTONES.md`
- `CONTRIBUTING.md`
