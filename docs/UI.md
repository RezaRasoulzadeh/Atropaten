# Atropaten frontend foundation

The frontend uses Tailwind CSS 4 and DaisyUI 5 with one custom dark `atropaten` theme. The application is dense and desktop-first, while preserving the existing Vue/Wails contracts and Go-owned domain behavior.

## Theme

`frontend/src/style.css` contains the only theme stylesheet. Surfaces use DaisyUI `base-100`, `base-200`, `base-300`, and `neutral` colors. Amber is the `primary` accent for actions, active navigation, links, and focus emphasis. Containers do not use amber borders, shadows, glows, gradients, bevels, or inset shadows.

Standard controls use DaisyUI classes (`btn`, `input`, `select`, `textarea`, `table`, `badge`, `alert`, `modal`, and `card`) and Tailwind layout utilities. There is no legacy feature CSS import or global control override.

Editable controls keep a solid neutral border at rest and use a same-width 1px dashed primary border on focus. Inputs, textareas, native selects, and custom select triggers have no outline, focus ring, shadow, glow, or layout shift; validation colors remain visible. This contract does not apply to cards, containers, or ordinary buttons.

Estedad is vendored locally as the single variable `frontend/src/assets/fonts/Estedad-VF.woff2` file. `frontend/src/styles/fonts.css` is the centralized local `@font-face` entry point, with the 100–900 weight axis and no metric overrides. No runtime font/CDN request is required.

## Commercial workflow

Orders are the single customer-sales workspace. There is no separate Quotes workspace in the target UI.

- **Draft Order** is the pre-commitment workspace for selecting a customer, configuring services, calculating and revising prices, editing line items, adding notes/attachments, and saving unfinished work.
- **Confirmed Order** represents customer commitment and preserves the accepted configuration, cost, and price snapshots.
- Fulfillment and payment remain independent state axes after confirmation.
- Saving a Draft Order must not imply inventory reservation, production creation, invoice issuance, or accounting posting.
- If the shop needs to hand a customer an estimate, the Draft Order may be printed/exported as an estimate/proposal without introducing a separate Quote entity or navigation destination.
- Legacy quote records may remain accessible only through explicit compatibility/history handling until migration is complete; they are not a primary workflow.

## Shared components and boundaries

Shared DaisyUI wrappers live in `frontend/src/components`: buttons, icon buttons, inputs, selects, textareas, form fields, messages, search/filter bars, tables, panels, badges, empty/loading states, alerts, confirmations, toasts, workspace tabs, sticky regions, and the application toolbar/sidebar primitives.

The frontend boundaries are documented in [FRONTEND_ARCHITECTURE.md](FRONTEND_ARCHITECTURE.md). Wails bridge modules remain under `frontend/src/api`, presentation utilities under `frontend/src/utils`, feedback services under `frontend/src/ui`, and domain workspaces under `frontend/src/views`.

## Layout rules

- Search/filter bars use a flexing search slot, same-row filters, and a compact result count. Labels are small and secondary when visible.
- Equivalent controls use the same DaisyUI size, border, radius, padding, and focus behavior.
- Tables use DaisyUI table primitives inside `overflow-x-auto` only when content is genuinely wide.
- Panels and inspectors use neutral dark surfaces. Empty states keep icon and text as one compact aligned group.
- Register/list rows, dense data tables, inspector shells, and compact form sections are local reusable primitives. They provide consistent workspace geometry without moving page-specific business layout into global CSS.
- Dense tables use aligned numeric cells, explicit dividers, and no zebra striping when stock, cost, or ledger values must be compared across rows.
- The shell uses Tailwind grid/flex utilities for the sidebar, toolbar, scrollable workspace, sticky page regions, and status strip.
- Desktop windows remain usable at narrower widths through wrapping and stacking; domain calculations, persistence, accounting, inventory, production, pricing, Jalali dates, money/quantity formatting, and Wails contracts are unchanged.

## Feedback preservation

`frontend/src/ui/feedback.ts` remains the central toast, error normalization, and confirmation service. `ToastHost` supports success, error, warning, and info outcomes. Inline page errors and confirmations remain part of the existing behavior; this foundation cleanup does not move domain rules into Vue.

## Cleanup scope

M6-008 removes the legacy global stylesheet, feature CSS files, scoped view CSS, old control class names, and obsolete compatibility selectors. The migration is foundation-only; individual page redesign and additional feedback behavior belong to later work.

See [FRONTEND_CLEANUP_AUDIT.md](FRONTEND_CLEANUP_AUDIT.md) and [DAISYUI_MIGRATION_AUDIT.md](DAISYUI_MIGRATION_AUDIT.md) for the detailed inventory and validation limits.
