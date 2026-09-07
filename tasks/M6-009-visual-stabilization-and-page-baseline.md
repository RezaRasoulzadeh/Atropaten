# M6-009 — Visual stabilization and page baseline

## Objective
Stop incremental UI patching and establish one reliable visual baseline across the entire frontend. Fix the concrete typography/control regressions first, then bring every workspace to the same structural quality as the current Dashboard and Orders main view. This is not a page-by-page redesign yet; it is a stabilization pass so later page revisions start from a clean, predictable system.

## Non-negotiable reference
The current Dashboard and Orders main view are the only acceptable baseline. Preserve their overall density, dark neutral surfaces, restrained Amber usage, sidebar/topbar proportions, compact controls, and information hierarchy. Do not invent a new visual language in this task.

## 1. Fix Vazirmatn rendering correctly
The current font integration contains manual metric overrides and asymmetric control padding. Remove the hacks instead of tuning around them.

Current known problems to eliminate:
- `frontend/src/styles/fonts.css` uses `ascent-override`, `descent-override`, and `line-gap-override`. Remove these metric overrides and let the actual Vazirmatn font metrics render normally unless a browser-specific, measured reason proves otherwise.
- `frontend/src/style.css` gives `.input, .select, .textarea` asymmetric `padding-block: 0.625rem 0.375rem`; remove that optical compensation.
- Do not solve vertical alignment through random per-component top/bottom padding, transforms, relative positioning, negative margins, or line-height exceptions.

Typography contract:
- Vazirmatn remains the global UI font, locally vendored.
- Use explicit, shared typography sizes/weights/line-heights for body, labels, headings, table text, badges, buttons and inputs.
- Controls of equal height must have visually centered text across button/input/select/textarea triggers.
- Text baselines must remain consistent when English, Persian, numbers, Jalali dates and icons appear together.
- SVG icons must align through flex/grid alignment, not by moving text vertically.
- Verify at 100% browser/WebView zoom. Do not accept a fix that only looks correct at one arbitrary element height.

## 2. Fix selects once, globally
Selects must not have centered option/trigger text unless the design explicitly calls for it. Standard form/select text is start-aligned.

Audit every select implementation, including native `<select>`, DaisyUI `.select`, custom select triggers and wrapper components.

Requirements:
- One shared select implementation/pattern for ordinary selects.
- Trigger/value text uses `text-align: start` and normal flex alignment.
- Dropdown indicator stays at the inline end.
- No generic `.btn` selector may accidentally center select values.
- Remove broad button/grid hacks that affect select triggers or other controls indirectly.
- Do not use `!important` as the normal solution.
- Native and custom select controls must share the same height, typography, radius, border and focus treatment as inputs.
- Placeholder/value/selected state must remain visually left/start aligned in LTR UI, while Persian content itself renders correctly.

## 3. Simplify global CSS aggressively
Treat `frontend/src/style.css` as infrastructure, not a second component library.

Keep only:
- Tailwind/DaisyUI/theme configuration;
- global font assignment;
- minimal shell geometry that DaisyUI/Tailwind cannot express cleanly;
- genuinely shared technical fixes with documented reasons.

Remove or migrate:
- global `.btn` reimplementations of DaisyUI layout;
- `:has()` layout tricks for button centering;
- redundant input/select/button styling already provided by DaisyUI;
- broad element overrides that create unexpected behavior in unrelated pages;
- page-specific visual fixes hidden in global CSS;
- stale M6-004/M6-005/M6-007 compatibility rules;
- duplicate component classes and dead CSS.

Prefer DaisyUI component classes plus local Tailwind layout utilities. Do not reproduce DaisyUI internals manually.

## 4. Establish shared page primitives
Before touching individual workspaces, define/repair the small set of primitives every page must use:
- `WorkspaceHeader`: eyebrow/title/description/actions;
- `SearchFilterBar`: flexible search, compact labelled filters, result count;
- `FormField`: compact label/help/error and a single control slot;
- shared Input, Select, Textarea and Button conventions;
- `DataTable` / register container;
- `AppPanel` / neutral content section;
- `StatusBadge`;
- `EmptyState`, `LoadingState`, `InlineAlert`;
- shared tabs and sticky bottom actions where needed;
- existing ToastHost, centralized toast API and ConfirmDialog.

These primitives should be thin wrappers around DaisyUI defaults. They must not become giant configurable abstractions.

## 5. Normalize every workspace to the baseline
Audit all major views and remove obvious visual breakage so each page is coherent before later detailed redesign tasks.

At minimum inspect:
- Dashboard
- Orders and order workspace
- Quotes and quote/configurator workspace
- Production
- Customers
- Services
- Materials
- Machines
- Purchases
- Suppliers
- Accounting
- Invoices
- Expenses
- Treasury/payments
- Checks
- Loans
- Owners
- Reports
- Settings
- print-preview entry surfaces

For every page fix baseline problems only:
- broken/misaligned header geometry;
- inconsistent toolbar/filter arrangement;
- clipped/overflowing forms;
- inconsistent control heights;
- oversized labels/headings;
- unnecessary cards/nested boxes;
- giant empty-state regions;
- random borders/backgrounds/shadows;
- inconsistent tab/panel/table spacing;
- inspector width/clipping problems;
- inconsistent action placement;
- incorrect select alignment;
- typography/baseline problems.

Do not creatively redesign business workflows yet. Preserve information and behavior. The goal is “clean and internally consistent”, not “final page design”.

## 6. Visual acceptance must be evidence-based
Do not report success from code inspection alone when a browser/runtime is available.

Use the M6-006 demo dataset and inspect populated screens when backend-connected Wails dev is available. At minimum visually check representative pages at 1024, 1280 and 1600 px widths.

Create `docs/M6-009-VISUAL-AUDIT.md` with a table for every major workspace containing:
- inspected/not inspected;
- populated/empty state checked;
- header/toolbar/forms/tables/inspector checked;
- remaining page-specific design issues deferred to later revisions.

If Wails/native rendering is unavailable, state exactly what was not visually verified. Do not claim “fixed” for a visual defect that was never rendered unless the defect was mechanically proven and documented.

## 7. Preserve product behavior
Do not change Go domain logic, persistence, accounting, inventory, pricing, posting or migrations except if a frontend compile/runtime contract requires a narrowly scoped correction.

Preserve:
- centralized toasts for meaningful success/error/warning/info;
- error normalization;
- confirmation dialogs;
- Jalali dates;
- Rial/Toman formatting;
- keyboard/navigation behavior;
- backend authority.

## Acceptance criteria
- Vazirmatn renders without metric-override hacks and without asymmetric control padding.
- Standard inputs/buttons/selects of the same size are visually vertically aligned.
- Ordinary selects are consistently start-aligned everywhere.
- Global CSS is substantially smaller/simpler and no longer reimplements DaisyUI controls.
- Dashboard and Orders remain acceptable and are not regressed.
- Every major workspace has been normalized to the same baseline structural system; obvious visual breakage is removed.
- Page-specific redesign remains deferred and recorded in the visual audit.
- No silent user-visible async outcomes are introduced.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.
- Run frontend tests if a test script exists; otherwise state that none exists.
- Commit and push to `origin/main` and return the final SHA plus the visual-audit summary.
