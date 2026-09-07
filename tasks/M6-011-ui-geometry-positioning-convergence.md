# M6-011 — UI geometry and positioning convergence

## Objective
Finish the structural UI pass that the interrupted overall edit started. Make positioning, width, height, spacing, sticky regions, master-detail geometry, overflow, and responsive behavior consistent across the entire frontend before further visual polish.

This task is about **geometry only**. Do not spend time changing typography, colors, table styling, or page-specific aesthetics except where required to expose/fix a positioning defect.

## Current context
The frontend has already been reorganized into `src/app`, `src/components/ui`, `src/components/layout`, and `src/features`. Shared primitives now exist, but the repository-wide audit is still marked mostly pending. The goal is to make those primitives authoritative and eliminate ad-hoc page geometry.

## Required work

### Shell
Audit `AppShell.vue`, toolbar, sidebar/drawer, workspace scroll container, status/toast/dialog layers, and transitions.

Guarantee:
- exactly one intended vertical workspace scroller;
- no hidden page content behind toolbar/sticky regions;
- no body-level accidental scrolling;
- collapsed/expanded sidebar does not change content geometry unpredictably;
- drawer behavior is correct below desktop breakpoint;
- no horizontal shell overflow at 1024, 1280, or 1600 CSS px;
- toast/dialog layers do not shift normal layout.

### Workspace geometry
Make `WorkspaceHeader`, `WorkspaceStickyStack`, `WorkspaceTabs`, `AppToolbar`, `AppPanel`, `MasterDetail`, `InspectorShell`, `InspectorHeader`, `InspectorSection`, and `WorkspaceBottomActions` the only shared geometry contracts where applicable.

Eliminate page-specific copies of those patterns.

Standardize:
- workspace gutters;
- vertical rhythm between header, filters/tabs, panels and content;
- sticky offsets and z-index;
- panel padding;
- full-width vs master-detail content;
- inspector min/max width;
- main/register `min-width: 0` behavior;
- responsive collapse points;
- bottom action clearance;
- intentional horizontal scrolling only for genuinely wide data tables.

### Master/detail pages
Inspect Customers, Materials, Machines, Suppliers, Production, Checks, Loans, and other inspector-oriented workspaces.

At desktop widths:
- register/table should flex;
- inspector has a stable sensible width, not arbitrary per page;
- inspector cannot clip fields or overflow the viewport;
- selection/editor transitions must not resize the entire workspace unexpectedly.

At narrower widths:
- collapse to a usable single-column flow before either side becomes unusable;
- no controls clipped off-screen;
- no hidden action rows.

### Full-editor pages
Inspect New Order / Order workspace, Quote workspace, Services editor, Purchases editor, Accounting posting forms, Owners, Reports, Settings.

Ensure:
- coherent max/content width where appropriate;
- no narrow centered form floating in a huge unused workspace unless intentionally designed;
- grouped content aligns to a shared grid;
- sticky bottom actions remain reachable and never cover content;
- calendar/select/dialog popovers are not clipped by panel overflow.

## Verification
Use the development browser bridge and M6-006 demo data when available.

Render **every major workspace** at 1024, 1280, and 1600 px. For each, record:
- main horizontal overflow;
- panel/inspector clipping;
- sticky overlap;
- unused pathological width;
- action visibility;
- responsive collapse behavior.

Update `docs/FRONTEND_UI_AUDIT.md` with real rendered evidence. Do not mark a page complete from source inspection alone.

## Constraints
- Preserve the dark Amber theme.
- Preserve Vazirmatn and current control-focus contract.
- Preserve backend/domain/Wails behavior.
- Do not introduce new global feature CSS.
- Prefer Tailwind/DaisyUI and shared layout components.

## Acceptance
- No accidental workspace horizontal overflow at 1024/1280/1600.
- Shared geometry primitives are used consistently.
- No major page has clipped inspector/forms/actions.
- Sticky regions do not overlap content.
- Browser screenshots/evidence are recorded for all contexts.
- `go test ./...`, `cd frontend && npm run build`, `git diff --check` pass.
- Commit/push to main and report SHA plus any remaining geometry-only native gaps.