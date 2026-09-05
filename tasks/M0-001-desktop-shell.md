# M0-001 — Build the Atropaten desktop shell

## Objective

Replace the default Wails/Vue starter interface with the initial Atropaten desktop application shell and reusable visual foundation defined in `docs/UI.md`.

This task is UI foundation only. Do not introduce SQLite, business persistence, accounting logic, or backend domain modeling yet.

## Scope

Implement the first usable desktop shell in `frontend/` using Vue 3 + TypeScript + Vite.

### Application shell

Create a Windows-first light desktop layout with:

- Collapsible left sidebar
- Top application bar
- Main workspace area
- Compact bottom status bar
- Responsive behavior appropriate for desktop/laptop window sizes

### Navigation

Add sidebar entries for:

- Dashboard
- Orders
- Production
- Customers
- Services
- Materials
- Purchases
- Suppliers
- Accounting
- Checks
- Loans
- Owners
- Reports
- Settings

Dashboard should be the initial active view.

Navigation may use local Vue state for this milestone. Do not add a router unless it materially improves the implementation.

### Dashboard

Build a representative dashboard using static/mock frontend data. Include:

- Sales KPI
- Profit KPI
- Receivables KPI
- Payables KPI
- Production status / active jobs
- Financial alerts such as due checks, installments, or overdue invoices
- Low-stock section
- Recent transactions
- Quick actions

The dashboard should prioritize operational attention over decorative charts.

### Reusable UI foundation

Create reusable components/styles for the patterns that upcoming screens will need, including at minimum:

- Sidebar/navigation item
- Toolbar/header area
- KPI/stat card
- Status badge
- Data table styling
- Section/panel surface
- Buttons
- Form controls

Avoid creating an oversized component framework. Extract components only where reuse is already clear.

### Visual direction

Follow `docs/UI.md`:

- Light theme
- Neutral light gray application background
- White work surfaces
- Subtle borders/shadows
- Compact desktop spacing
- Around 14px normal body text
- 36–42px dense table rows
- Moderate corner radius
- Strong hierarchy without oversized headings
- One primary accent
- Green/red mostly reserved for financial/status semantics
- Minimal or no gradients

Use CSS variables/design tokens for the core palette, spacing, radii, typography, and borders so later screens remain consistent.

### Starter cleanup

Remove the default Wails starter content and any starter assets/styles/components that are no longer used, including `HelloWorld` and the Wails logo if they become unused.

Do not remove files required by Wails itself.

## Constraints

- Vue 3 Composition API and TypeScript.
- No backend/database implementation in this task.
- No external UI framework unless already present in the project.
- Prefer native CSS and small local components.
- Do not add charting libraries just to decorate the dashboard.
- Keep the implementation understandable and easy to extend.
- Do not hardcode business behavior that belongs in later backend milestones.

## Acceptance criteria

1. `npm run build` in `frontend/` succeeds.
2. `wails dev` can launch the application successfully on the supported development setup.
3. No default Wails starter screen remains visible.
4. The application opens into the Atropaten dashboard shell.
5. Sidebar can collapse and expand.
6. All planned primary navigation entries are visible/reachable.
7. Dashboard includes representative operational and financial mock data.
8. Core design tokens and reusable UI patterns are established for later screens.
9. Layout remains usable at normal desktop/laptop window sizes without horizontal page-level overflow.
10. No SQLite, persistence layer, or real accounting/business implementation is added.

## Validation

Run at minimum:

```sh
cd frontend
npm run build
```

Then run the Wails development application using the platform-appropriate command. On Linux installations using WebKitGTK 4.1, use the required Wails build tag if necessary.

## Relevant documentation

- `docs/UI.md`
- `docs/PRODUCT.md`
- `docs/ARCHITECTURE.md`
- `ROADMAP.md`
- `MILESTONES.md`
