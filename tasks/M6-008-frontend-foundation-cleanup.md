# M6-008 — Frontend foundation cleanup and dark theme

## Objective
Clean up the M6-007 frontend without redesigning individual pages. Establish a maintainable modular Vue component architecture and a single dark DaisyUI theme so subsequent page-by-page revisions can proceed on a stable foundation.

## Scope

### 1. DaisyUI-first styling
- Audit all frontend CSS, scoped styles, utility classes, and shared components. Remove obsolete legacy rules, duplicate declarations, specificity hacks, broad element overrides, and custom implementations of styling already supplied by DaisyUI.
- Use DaisyUI 5 default component classes, states, sizing, spacing, and appearance wherever possible. Do not recreate buttons, inputs, selects, tables, cards, badges, alerts, menus, tabs, or dialogs through custom CSS.
- Keep custom CSS only for application shell geometry, genuinely specialized widgets, necessary layout constraints, font loading, and documented theme tokens. Do not retain compatibility CSS solely to preserve obsolete visual designs.
- Avoid `!important`, deep selectors, arbitrary hardcoded colors, and page-specific overrides unless a documented technical necessity exists. Remove dead styles and unused components safely.
- Preserve functional behavior; do not perform a visual redesign of every workspace in this task.

### 2. Required dark Amber theme
- Replace the current light-only Atropaten theme with one centrally defined dark DaisyUI theme using **Amber as the primary accent**. This choice is fixed for this task; do not compare or substitute Cyan.
- Base surfaces must remain neutral dark gray/near-black. Amber is an accent, not a container color.
- **Do not use amber borders, amber outlines, amber box-shadows, amber glows, or amber-tinted card/panel/container borders.** Cards, panels, tables, inspectors, dialogs, sidebar regions, headers, toolbars, and other containers must use neutral DaisyUI base/neutral surfaces and neutral borders where borders are needed.
- Amber may be used for primary actions, selected navigation state, active tabs, focus indication, small emphasis/icon accents, links, and appropriate highlighted interactive states.
- Buttons, inputs, selects, textareas, checkboxes, radios, tabs, and other interactive controls may use visible borders according to DaisyUI defaults. Keep these borders neutral by default; use semantic/focus colors only for actual interaction or validation state.
- Use DaisyUI semantic tokens centrally: base-100/base-200/base-300, base-content, primary, primary-content, neutral, neutral-content, success, warning, error, and info. Avoid page-level hardcoded palette values.
- Keep success green, warning amber/yellow, error red, and info blue distinct from the primary accent where semantic meaning requires it.
- Ensure readable contrast, clear keyboard focus, disabled states, table hierarchy, form validation, toast severity, dialogs, and hover/active states.
- No decorative gradients, glow, bevels, inset shadows, colored card shadows, or decorative visual effects.
- Default the application to this dark Amber theme. Do not add a theme switcher or theme editor in this task.
- Update `docs/UI.md` with the exact theme rules so future page-by-page revisions use the same visual foundation.

### 3. Vazirmatn typography
- Use **Vazirmatn** as the primary application font, appropriate for Persian/RTL UI while remaining usable for Latin text and numbers.
- Download the required Vazirmatn webfont files from the official/upstream project during implementation and commit the required font assets into the frontend repository so the desktop app does not depend on Google Fonts, a CDN, or network access at runtime.
- Include only the weights actually used by the application; do not vendor the entire upstream font repository.
- Define local `@font-face` declarations in one central font stylesheet and apply Vazirmatn globally through the application/theme foundation.
- Prefer WOFF2 assets for runtime use. Preserve font licensing/attribution requirements by including the relevant license/notice in the repository when required by the upstream font license.
- Do not embed font binaries as base64 in CSS or source files.
- Ensure Persian text, Latin text, Rial/Toman values, table figures, inputs, buttons, toasts, and print-facing browser UI inherit the intended font consistently unless a specialized document-print requirement explicitly needs another font.

### 4. Modular frontend architecture
- Inspect the existing frontend tree before moving files. Establish clear ownership boundaries for app shell, shared UI components, composables, services/API bridge, types, utilities, domain features, and page/workspace components.
- Prefer a feature-oriented structure such as `src/app`, `src/components/ui`, `src/composables`, `src/services`, `src/types`, `src/utils`, and `src/features/<domain>` with local components/composables/types where appropriate. Adapt names to the actual repository; do not create empty folders or unnecessary abstraction layers.
- Split oversized components by responsibility. Extract reusable behavior and presentation only when genuinely shared. Avoid giant shared components, circular imports, catch-all utility files, and unnecessary barrel exports.
- Shared UI wrappers should be thin, typed, and built on DaisyUI defaults. Preserve existing component APIs where practical; otherwise migrate call sites consistently and remove obsolete aliases.
- Centralize common toast/error handling, confirmation, form-field integration, money/quantity/Jalali formatting, and Wails bridge access without duplicating domain logic in the frontend.
- Keep Go/domain/persistence authority unchanged. Do not introduce a new state-management or UI framework merely for this cleanup.

### 5. Preserve behavior and prepare page-by-page revision
- Preserve navigation, workspace routes, data loading, forms, validation, posting actions, keyboard behavior, sticky shell geometry, and existing backend contracts.
- Retain the M6-007 global toast requirement: meaningful success, error, warning, and information outcomes must not be silently swallowed. Keep centralized error normalization and shared confirmation behavior.
- Do not redesign Orders, Quotes, Production, Accounting, or other individual pages beyond what is required to migrate them to the cleaned foundation and maintain functionality.
- Do not add new business features, migrations, domain changes, or unrelated backend fixes.
- Create `docs/FRONTEND_ARCHITECTURE.md` documenting folder ownership, dependency direction, shared component conventions, styling rules, and how to add/revise a page.
- Update `docs/UI.md` and add `docs/FRONTEND_CLEANUP_AUDIT.md` recording removed CSS, moved components, retained exceptions, Amber theme implementation, Vazirmatn integration, and deferred page-specific issues.

## Acceptance criteria
- One centrally defined **dark Amber DaisyUI theme** is active and documented.
- Container surfaces remain neutral: no amber borders/shadows/glows on cards, panels, tables, dialogs, inspectors, toolbars, sidebar regions, or page containers.
- Standard interactive controls use DaisyUI defaults; buttons, inputs, selects, and related controls may use appropriate borders/focus states without introducing a competing custom control system.
- Vazirmatn is vendored locally from the official/upstream source, loaded via local WOFF2 assets, applied globally, and requires no runtime network request.
- Standard UI elements use DaisyUI defaults rather than competing custom styling systems.
- Obsolete CSS and unused compatibility layers are removed; remaining custom CSS has a clear purpose.
- Frontend folders and component boundaries are coherent, typed, and documented; no broken imports or circular dependencies introduced.
- Existing functional workflows and shared toast/confirmation/error behavior remain intact.
- Individual page redesign is explicitly deferred for subsequent tasks.
- `go test ./...`, `cd frontend && npm run build`, and `git diff --check` pass. Run available frontend tests; report absence of a test script rather than claiming tests ran.
- Perform available browser visual checks of the theme and shared components. Use M6-006 demo data for populated checks when a backend-connected runtime is available. Report native Wails/Windows/WebView limitations accurately.
- Commit and push to origin/main; report final SHA, changed architecture, removed CSS, theme implementation, Vazirmatn source/weights/assets, validation, and remaining limitations.
