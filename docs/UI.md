# Atropaten frontend foundation

The frontend uses Tailwind CSS 4 and DaisyUI 5 with one custom dark `atropaten` theme. The application is dense and desktop-first, while preserving the existing Vue/Wails contracts and Go-owned domain behavior.

## Theme

`frontend/src/style.css` contains the only theme stylesheet. Surfaces use DaisyUI `base-100`, `base-200`, `base-300`, and `neutral` colors. Amber is the `primary` accent for actions, active navigation, links, and focus emphasis. Containers do not use amber borders, shadows, glows, gradients, bevels, or inset shadows.

Standard controls use DaisyUI classes (`btn`, `input`, `select`, `textarea`, `table`, `badge`, `alert`, `modal`, and `card`) and Tailwind layout utilities. There is no legacy feature CSS import or global control override.

Vazirmatn is vendored locally in `frontend/src/assets/fonts/vazirmatn/` with only regular, semibold, and bold files. `frontend/src/styles/fonts.css` is the centralized local `@font-face` entry point; `OFL.txt` records the upstream SIL Open Font License notice. No runtime font/CDN request is required.

## Shared components and boundaries

Shared DaisyUI wrappers live in `frontend/src/components`: buttons, icon buttons, inputs, selects, textareas, form fields, messages, search/filter bars, tables, panels, badges, empty/loading states, alerts, confirmations, toasts, workspace tabs, sticky regions, and the application toolbar/sidebar primitives.

The frontend boundaries are documented in [FRONTEND_ARCHITECTURE.md](FRONTEND_ARCHITECTURE.md). Wails bridge modules remain under `frontend/src/api`, presentation utilities under `frontend/src/utils`, feedback services under `frontend/src/ui`, and domain workspaces under `frontend/src/views`.

## Layout rules

- Search/filter bars use a flexing search slot, same-row filters, and a compact result count. Labels are small and secondary when visible.
- Equivalent controls use the same DaisyUI size, border, radius, padding, and focus behavior.
- Tables use DaisyUI table primitives inside `overflow-x-auto` only when content is genuinely wide.
- Panels and inspectors use neutral dark surfaces. Empty states keep icon and text as one compact aligned group.
- The shell uses Tailwind grid/flex utilities for the sidebar, toolbar, scrollable workspace, sticky page regions, and status strip.
- Desktop windows remain usable at narrower widths through wrapping and stacking; domain calculations, persistence, accounting, inventory, production, pricing, Jalali dates, money/quantity formatting, and Wails contracts are unchanged.

## Feedback preservation

`frontend/src/ui/feedback.ts` remains the central toast, error normalization, and confirmation service. `ToastHost` supports success, error, warning, and info outcomes. Inline page errors and confirmations remain part of the existing behavior; this foundation cleanup does not move domain rules into Vue.

## Cleanup scope

M6-008 removes the legacy global stylesheet, feature CSS files, scoped view CSS, old control class names, and obsolete compatibility selectors. The migration is foundation-only; individual page redesign and additional feedback behavior belong to later work.

See [FRONTEND_CLEANUP_AUDIT.md](FRONTEND_CLEANUP_AUDIT.md) and [DAISYUI_MIGRATION_AUDIT.md](DAISYUI_MIGRATION_AUDIT.md) for the detailed inventory and validation limits.
