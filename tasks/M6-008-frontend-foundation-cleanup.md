# M6-008 — Frontend foundation cleanup and dark theme

## Objective
Clean up the M6-007 frontend without redesigning individual pages. Establish a maintainable modular Vue component architecture and a single dark DaisyUI theme so subsequent page-by-page revisions can proceed on a stable foundation.

## Scope

### 1. DaisyUI-first styling
- Audit all frontend CSS, scoped styles, utility classes, and shared components. Remove obsolete legacy rules, duplicate declarations, specificity hacks, broad element overrides, and custom implementations of styling already supplied by DaisyUI.
- Use DaisyUI 5 default component classes, states, sizing, spacing, and appearance wherever possible. Do not recreate buttons, inputs, selects, tables, cards, badges, alerts, menus, tabs, or dialogs through custom CSS.
- Keep custom CSS only for application shell geometry, genuinely specialized widgets, necessary layout constraints, and documented theme tokens. Do not retain compatibility CSS solely to preserve obsolete visual designs.
- Avoid `!important`, deep selectors, arbitrary hardcoded colors, and page-specific overrides unless a documented technical necessity exists. Remove dead styles and unused components safely.
- Preserve functional behavior; do not perform a visual redesign of every workspace in this task.

### 2. Dark theme
- Replace the current light-only Atropaten theme with a coherent dark theme. Use DaisyUI's dark theme as the foundation, with one restrained accent: Amber or Cyan. Choose the better fit after comparing both in a small theme preview; do not mix both as competing primary accents.
- Define the theme centrally using DaisyUI 5's supported theme configuration and semantic color tokens. Use base-100/base-200/base-300, base-content, primary, secondary, accent, neutral, and semantic success/warning/error/info colors.
- Ensure readable contrast, visible focus, disabled states, table hierarchy, form validation, toast severity, dialogs, and keyboard navigation. No decorative gradients, glow, bevels, inset shadows, or custom visual effects.
- Default the application to the new dark theme. Keep theme selection infrastructure simple; do not build an elaborate theme editor or unrelated customization system.
- Document the selected accent, theme tokens, and rules for future page revisions in docs/UI.md.

### 3. Modular frontend architecture
- Inspect the existing frontend tree before moving files. Establish clear ownership boundaries for app shell, shared UI components, composables, services/API bridge, types, utilities, domain features, and page/workspace components.
- Prefer a feature-oriented structure such as `src/app`, `src/components/ui`, `src/composables`, `src/services`, `src/types`, `src/utils`, and `src/features/<domain>` with local components/composables/types where appropriate. Adapt names to the actual repository; do not create empty folders or unnecessary abstraction layers.
- Split oversized components by responsibility. Extract reusable behavior and presentation only when genuinely shared. Avoid giant shared components, circular imports, catch-all utility files, and unnecessary barrel exports.
- Shared UI wrappers should be thin, typed, and built on DaisyUI defaults. Preserve existing component APIs where practical; otherwise migrate call sites consistently and remove obsolete aliases.
- Centralize common toast/error handling, confirmation, form-field integration, money/quantity/Jalali formatting, and Wails bridge access without duplicating domain logic in the frontend.
- Keep Go/domain/persistence authority unchanged. Do not introduce a new state-management or UI framework merely for this cleanup.

### 4. Preserve behavior and prepare page-by-page revision
- Preserve navigation, workspace routes, data loading, forms, validation, posting actions, keyboard behavior, sticky shell geometry, and existing backend contracts.
- Retain the M6-007 global toast requirement: meaningful success, error, warning, and information outcomes must not be silently swallowed. Keep centralized error normalization and shared confirmation behavior.
- Do not redesign Orders, Quotes, Production, Accounting, or other individual pages beyond what is required to migrate them to the cleaned foundation and maintain functionality.
- Do not add new business features, migrations, domain changes, or unrelated backend fixes.
- Create `docs/FRONTEND_ARCHITECTURE.md` documenting folder ownership, dependency direction, shared component conventions, styling rules, and how to add/revise a page.
- Update `docs/UI.md` and add `docs/FRONTEND_CLEANUP_AUDIT.md` recording removed CSS, moved components, retained exceptions, theme choice, and deferred page-specific issues.

## Acceptance criteria
- One centrally defined dark DaisyUI theme with Amber or Cyan primary accent is active and documented.
- Standard UI elements use DaisyUI defaults rather than competing custom styling systems.
- Obsolete CSS and unused compatibility layers are removed; remaining custom CSS has a clear purpose.
- Frontend folders and component boundaries are coherent, typed, and documented; no broken imports or circular dependencies introduced.
- Existing functional workflows and shared toast/confirmation/error behavior remain intact.
- Individual page redesign is explicitly deferred for subsequent tasks.
- `go test ./...`, `cd frontend && npm run build`, and `git diff --check` pass. Run available frontend tests; report absence of a test script rather than claiming tests ran.
- Perform available browser visual checks of the theme and shared components. Use M6-006 demo data for populated checks when a backend-connected runtime is available. Report native Wails/Windows/WebView limitations accurately.
- Commit and push to origin/main; report final SHA, changed architecture, theme choice, removed CSS, validation, and remaining limitations.
