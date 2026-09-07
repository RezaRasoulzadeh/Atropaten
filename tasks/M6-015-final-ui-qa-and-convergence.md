# M6-015 — Final UI QA and convergence

## Objective
Perform the final repository-wide visual QA after M6-011 through M6-014. Do not introduce a new design direction. Find and fix remaining inconsistencies until the application reads as one coherent DaisyUI desktop ERP.

## Scope
Audit every major frontend context:
- Dashboard
- Orders register
- New Order / Order workspace
- Quotes / Quote workspace
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
- Treasury / transfers
- Checks
- Loans
- Owners
- Reports
- Settings
- print-preview entry surfaces

## Review dimensions
For every context verify all of the following:

### Position and geometry
- workspace header position;
- toolbar/filter/tab alignment;
- panel spacing;
- register/table width;
- inspector width and placement;
- sticky offsets;
- bottom action placement;
- no accidental horizontal overflow;
- no clipped popovers/forms;
- responsive behavior at 1024/1280/1600.

### Typography
- Vazirmatn actually loaded;
- natural metrics only;
- headings, labels, values, metadata and badges use the shared type scale;
- no page-local font/padding compensation;
- text appears optically aligned inside controls and buttons.

### Controls
- equivalent control heights;
- start-aligned selects;
- consistent icon size/alignment;
- editable focus = 1px dashed Amber with no ring/glow/shadow;
- validation states remain visible;
- buttons follow the same primary/secondary/danger hierarchy.

### Data surfaces
- rich register or dense table pattern chosen intentionally;
- no accidental zebra striping;
- neutral row dividers;
- consistent hover/focus/selection;
- numeric/money alignment;
- readable secondary metadata;
- consistent empty/loading/error states.

### Forms and feedback
- labels/help/errors coherent;
- no placeholder-only important fields;
- no silent meaningful action;
- submit/busy states correct;
- dialogs/toasts work and do not obscure controls;
- destructive actions clearly differentiated.

### Theme
- dark neutral surfaces;
- Amber used only for primary/action/active/focus emphasis;
- no Amber card/container borders;
- no gradients, glow, bevels or inset shadows;
- no competing page-specific themes.

## Evidence requirement
Use M6-006 populated data with the development browser bridge.

Render every context at 1024, 1280, and 1600 CSS px. Capture enough screenshots to verify the actual populated UI, not only empty shells.

The task is not complete while `docs/FRONTEND_UI_AUDIT.md` still has major pages marked Pending/Rebuild/Not rendered, except explicit native-only Windows/WebView/printing gaps that cannot be exercised in the environment.

Update the audit into a final status matrix with columns for:
- geometry;
- typography;
- controls;
- data surface;
- forms/feedback;
- rendered widths;
- remaining native-only gap.

Fix defects found during the pass rather than merely documenting them.

## Regression constraints
- Do not change backend/domain rules.
- Do not rewrite working Dashboard/Orders baseline without a concrete defect.
- Do not add broad global CSS to patch individual pages.
- Reuse established DaisyUI/shared primitives.
- Remove any dead temporary compatibility code or audit-only hacks that should not ship.
- Development-only browser bridge must remain isolated from production data/build behavior.

## Validation
Run:
- `go test ./...`
- `cd frontend && npm run build`
- `git diff --check`
- available browser interaction/visual audit scripts
- frontend tests if a script exists; otherwise report absence accurately.

No false claim of Windows/WebView2/native print validation.

## Acceptance
- All major web-renderable contexts are actually rendered and pass the final matrix.
- No known major visual inconsistency remains.
- Position, typography, controls, tables/registers, forms and feedback are unified.
- Remaining gaps are native-only or explicitly minor and documented.
- Commit/push to main and report final SHA and concise residual-gap list.