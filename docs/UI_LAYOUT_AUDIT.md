# M6-004 UI layout audit

Date: 2026-09-07

## Scope and method

This pass audited the shared shell, every navigation workspace, the order/quote editors, configurators, finance surfaces, settings, and printable-preview entry surface. The audit focused on visual/layout behavior: theme tokens, spacing, typography, control sizing, table density, forms, inspectors, page headers/tabs, sticky geometry, workspace width, clipping/overflow, and narrow-to-wide desktop behavior. Domain calculations and persistence code were not changed.

## Screens audited

| Area | Checked surfaces |
| --- | --- |
| Global shell | Sidebar expanded/collapsed geometry, top bar, search, shop/currency controls, status bar, workspace scroll container |
| Dashboard | Sticky header/actions, KPI cards, operational tables, attention/stock/activity panels, persistent quick actions |
| Orders | Register/filter toolbar, operational table, selected rows, order header/meta strip, tabs, overview, items, production, payments, files, history |
| Quotes | Quote register, quote workspace, item editor, service configurator, pricing summary, metadata/proof surfaces |
| Catalog | Services/configurator, materials/register/inspector, machines/register/inspector |
| Purchasing | Purchases register/editor, suppliers register/inspector, purchase lines and totals |
| Operations | Production queue, job inspector, reservation/consumption/outsourcing sections |
| Finance | Accounting tabs, accounts/journal/payments, invoices/register/inspector, checks, loans, treasury/payment-related panels |
| Ownership and insight | Owners/fiscal periods, reports/filter tabs/summary/table/print preview |
| Setup | Settings/document identity and backup/restore surfaces |

The representative width targets are narrow laptop (about 1024px), normal desktop (1280–1440px), and wide desktop (1600px+). The app is intentionally desktop-first; no mobile redesign was attempted.

## Primitives introduced or refactored

- Expanded the root semantic token system for surfaces, borders, text, interaction, semantic status, disabled state, spacing, type scale, controls, table density, inspector width, radii, shadows, and z-index.
- Added a final shared layout layer at `frontend/src/views/ux-hardening.css`, loaded after feature styles so old view-specific rules converge on the shared system.
- Standardized page headings, panel headers, buttons, icon-button hit areas, controls/focus states, badges, table shells, tabs, inspector geometry, and bottom action surfaces.
- Added compatibility aliases for legacy feature styles (`--border`, `--border-subtle`, `--surface-muted`, and `--surface-raised`) so older views remain on the semantic system rather than silently rendering with missing values.
- Generalized `WorkspaceTabs` accessibility labeling and tab-list semantics.

## Inconsistencies fixed

- Removed the extra nested workspace gutter from Customers, which previously made the register/inspector stop early and waste available width.
- Replaced independent register/inspector column widths with one explicit inspector width model and a shared stacking breakpoint.
- Removed feature-level sticky inspector positioning that could overlap the shared sticky header stack.
- Unified control heights, focus treatment, disabled treatment, panel borders/radii/shadows, tab geometry, badge geometry, table header/row density, numeric alignment, and selected-row treatment.
- Added `min-width: 0`, border-box sizing, intentional table-only horizontal scrolling, and wrapping/collapse rules to the shared workspace layer.
- Reduced the page-title scale and capped workspace gutters to preserve dense desktop ERP behavior at wide widths.
- Aligned finance, report, settings, ownership, and editor surfaces with the same legacy-compatible semantic tokens and shared form/table primitives.

## Remaining manual Windows/WebView checks

These checks require running the packaged application in the target native environment and were not claimed here:

- Windows WebView2 rendering at 1024px, 1280px, 1440px, and 1600px with both sidebar states.
- Native focus traversal through every inspector, date picker, table row, tab strip, and bottom action surface.
- WebView2 scrollbar rendering, date-picker popover placement near the bottom/right viewport edges, and print-preview handoff.
- Actual Windows print dialog/page output for invoice, quote, receipt, and statement previews.
- Native Wails window resize behavior while an inspector form or wide table is active.

## Intentionally deferred

- Toast, loading, confirmation, and validation/error-state behavior beyond preserving the existing markup and functionality; these belong to M6-005.
- Mobile-specific redesign.
- Any accounting, inventory, production, pricing, reporting, posting, persistence, or other domain-semantic changes.

## Corrective follow-up after the initial M6-004 pass

The first M6-004 implementation exposed three visual regressions during screenshot review:

- The shared page-tab selector also styled tabs nested inside Checks and Loans filter bars, making those toolbars too tall and misaligned.
- The shared control rule used `surface-subtle` for every input/select/textarea, which made ordinary controls look recessed beside older scoped rules.
- Production retained a one-off unlabeled Queue span, so its filter row did not use the same labeled-control geometry as the reference toolbar.

The corrective pass gave the existing toolbar classes one explicit shared contract, migrated Production’s Queue field into a labeled filter control, made controls flat with one focus treatment, and reduced Parameters/Cost Components empty states to compact icon-text groups. The repair removes the conflicting empty-state and control declarations rather than adding another page-specific exception.

No browser, screenshot harness, Wails runtime, or Windows WebView2 environment is available in this workspace. I therefore verified the DOM/CSS structure and production build, but did not claim live visual inspection at the requested desktop widths; those remain manual checks for the native/browser environment.
