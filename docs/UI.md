# Atropaten UI system

Atropaten is a dense, Windows-first desktop ERP. The frontend is Vue 3 + Vite and uses Tailwind CSS with DaisyUI 5 as the styling foundation. Domain authority remains in Go/Wails; the frontend owns presentation, local form state, and feedback only.

## Theme and tokens

The single named DaisyUI theme is `atropaten`, configured at the top of `frontend/src/style.css`. It is a light theme with white work surfaces, a neutral application background, restrained blue primary accent, semantic green/amber/red/info colors, moderate radii, zero depth/noise, and no gradients or inset control shadows.

The existing semantic CSS tokens remain the application vocabulary: `--app-bg`, `--surface`, `--border`, `--text`, `--text-soft`, `--text-muted`, `--accent`, `--success`, `--warning`, `--danger`, `--info`, and the shared `--space-*` rhythm. Common controls are 36px high; compact actions are 32px; standard icon buttons are 34px.

## Reusable component contract

Shared components live in `frontend/src/components` and are exported from `frontend/src/ui/components.ts`:

- `AppButton` and `IconButton`: primary, secondary, ghost, danger, compact, disabled, and loading states with one icon/text alignment contract.
- `AppInput`, `AppSelect`, `AppTextarea`, `FormField`, and `FieldMessage`: DaisyUI-backed flat controls with consistent border, radius, padding, focus, disabled, helper, and error behavior.
- `SearchFilterBar`: one register toolbar geometry. The search slot flexes into remaining width; filters and the count remain compact peers on the same row and wrap before clipping.
- `DataTable`, `StatusBadge`, and `AppPanel`/`SectionPanel`: shared table shell, semantic status treatment, and compact section surface.
- `WorkspaceTabs`, `WorkspaceStickyStack`, and `WorkspaceBottomActions`: shared page geometry and sticky behavior.
- `EmptyState`, `LoadingState`, and `InlineAlert`: compact operational states rather than oversized placeholders.
- `ConfirmDialog`: shared promise-based confirmation service in `frontend/src/ui/feedback.ts`.
- `ToastHost`: one shell-level host using the central `useToast()` API.

Special money, quantity, and Jalali date controls continue to use their existing domain utilities/components. When a field is migrated, it receives the same `at-control` geometry without changing parsing, dates, amounts, or posting behavior.

Feature CSS is limited to workspace layout and domain-specific content. The former M6-004 `ux-hardening.css` control overrides were removed; its replacement contains only shared geometry and toolbar/layout selectors. Do not add broad selectors that restyle every input, select, button, or textarea outside the component contract.

## Toolbar and form rules

Every major register uses the same compact surface: search consumes available width, visible labels are small and muted, filters share the control baseline, and result counts align at the end. Oversized uppercase labels and page-specific toolbar geometry are not allowed.

Equivalent controls use the same height, border, radius, surface, padding, focus ring, disabled treatment, and no shadow. Forms use `min-width: 0`; inspector grids stack before controls clip. Tables may scroll horizontally only inside their table shell when the data is genuinely wide.

Panels are used for meaningful sections, not for every line of content. Empty states pair one small icon with text and an optional action. Dashed or oversized placeholder boxes are reserved for real drop/edit regions.

## Feedback and error handling

The shell mounts one `ToastHost`. The central API supports exactly `success`, `error`, `warning`, and `info`, with manual dismiss, severity-based timeout, deduplication, live-region semantics, and readable stacking above the status strip. User-triggered async outcomes must report success or failure. Page-blocking errors remain inline and also produce an error toast.

`normalizeError()` is the one frontend path for unknown thrown values. New feature code must not use `String(error)` for operator-facing messages, swallow a promise rejection, or leave a user-visible failure as a console-only event.

Destructive and financially significant actions use `confirmAction()` and `ConfirmDialog`, including delete draft, archive/reactivate where confirmation is required, post, void/reverse, close period, and backup restore. Browser `window.confirm` is not part of the application UI contract.

## Workspace geometry

The application shell owns the viewport: sidebar, top bar, scrollable workspace, and bottom status strip. `WorkspaceStickyStack` is sticky within the workspace scroll container and uses shared background, border, spacing, and z-index. `WorkspaceBottomActions` stays above the status strip. Register panes flex into remaining width beside deliberate inspectors; nested arbitrary fixed or overflow containers are defects.

Representative desktop acceptance widths are approximately 1024px, 1280–1440px, and 1600px+. Narrow desktop windows wrap filter rows and stack forms/registers before clipping; this is desktop degradation, not a mobile redesign. See [DAISYUI_MIGRATION_AUDIT.md](DAISYUI_MIGRATION_AUDIT.md) for the workspace inventory and remaining visual follow-up.

## Domain preservation

The UI migration does not move business rules into Vue and does not alter accounting, inventory, production, pricing, migrations, backup/restore, Jalali dates, grouped Rial/Toman formatting, or deletion/archive semantics. Existing Wails bindings and Go services remain authoritative.

Print previews retain their dedicated print surface and `@media print` rules. Their entry points use the same page headers, controls, and feedback system around the preview.
