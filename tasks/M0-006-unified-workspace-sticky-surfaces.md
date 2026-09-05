# M0-006 — Unified workspace geometry and sticky action surfaces

## Objective

Refine the current Atropaten workspace so page headers, sticky controls, persistent bottom actions, tabs, and scroll ownership behave as one coherent project-wide system.

The current implementation has too much unused whitespace between the global top bar and some page content, and different feature surfaces can drift toward one-off sticky/fixed behavior. This task must establish the shared implementation that future pages will reuse.

This is a UI-foundation task, not a visual redesign.

## Primary design correction

Use the space directly below the global top bar for the page header/action surface instead of leaving a large empty band before the page title and controls.

For example, in the current New Order Draft workspace, the section context/back action, page title, state metadata, and actions such as `More` / `Save draft` should participate in the shared top workspace header rather than appearing after excessive empty top spacing.

The target should feel like a dense desktop application:

```text
Global top bar
────────────────────────────────────
Sticky page header / page actions
Optional sticky tabs / filters
────────────────────────────────────
Scrollable page content
────────────────────────────────────
Optional persistent page actions
Global bottom status strip
```

## 1. Establish reusable workspace primitives

Create or refactor shared frontend components/styles for the common geometry. Use names appropriate to the existing codebase, but responsibilities should cover:

- page/workspace container
- page header
- page title/context region
- page action region
- sticky top stack
- optional sticky tabs/filter row
- scrollable page body
- persistent bottom action surface

Do not force every page to use every region. The structure must compose cleanly.

Avoid feature-specific duplicated layout CSS.

## 2. Scroll ownership

The application shell must own the viewport.

Stable shell surfaces:

- global sidebar
- global top bar
- global bottom status strip

Normal vertical scrolling belongs to the main workspace/body region between those shell surfaces.

Requirements:

- no document/body scrolling for normal desktop use
- no accidental double vertical scrollbars
- no feature page independently pretending to own the viewport
- inspectors/tables may scroll internally only when there is a clear UX reason
- scrollbars should appear on the expected content area

## 3. Shared sticky page header

Implement a reusable sticky page-header/action surface immediately below the global top bar.

It should support:

- breadcrumb/back navigation or section context
- page title
- compact description or metadata when useful
- status badges
- primary action
- secondary/overflow actions
- filters or controls where a page needs them

The page header should remain visible while long page content scrolls when persistent access materially improves the workflow.

Use one shared offset and z-index strategy rather than feature-specific magic numbers.

The visual style must remain consistent across pages:

- background compatible with the current workspace background/surface system
- subtle lower border/elevation only when needed to distinguish a scrolled sticky state
- compact desktop spacing
- no oversized floating-card appearance
- existing typography hierarchy preserved

## 4. Sticky tabs / secondary controls

For complex workspaces such as Orders, tabs may form a coordinated sticky stack below the page header.

Requirements:

- header and tabs must not overlap
- offsets must derive from shared layout variables/components
- dropdowns/menus must appear above sticky surfaces correctly
- focus targets must not be hidden
- switching tabs must not create layout jumps

Do not make every tab bar sticky automatically; use it where long content benefits from persistent workspace navigation.

## 5. Persistent bottom actions

Implement one shared pattern for important actions that should stay available near the bottom of a page.

Dashboard Quick Actions are the first explicit case.

Requirements:

- Quick Actions must remain accessible while dashboard content scrolls
- the action surface must sit above the global bottom status strip
- page content must reserve enough bottom space so nothing is hidden behind it
- avoid large floating overlays
- use the same border/background/elevation language as the rest of the application
- future pages should be able to reuse this pattern

Do not use unrelated `position: fixed` implementations inside individual feature components.

## 6. Apply the system to current screens

At minimum migrate and visually verify:

### Dashboard

- compact top workspace header
- remove unnecessary top whitespace
- Quick Actions use the shared persistent-bottom pattern
- dashboard content scrolls in the workspace only

### Orders list

- page title, filters, New Order action, and relevant page controls use the common header/sticky geometry
- long order content does not make actions disappear unnecessarily

### New Order Draft / Order workspace

- remove the excessive blank area above the current page title
- move page context/title/actions into the shared top workspace header
- preserve `More`, `Save draft`, status context, and back navigation
- coordinate tabs with the sticky header when scrolling
- ensure order forms/cards scroll beneath the sticky stack rather than moving the whole shell

### Existing additional implemented views

Any other current page using a top header/action area should be migrated if required to prevent a second competing layout pattern.

## 7. Unified design rule

Do not merely patch current screens. The resulting components/CSS must be suitable for all future major Atropaten pages.

Future pages such as Production, Materials, Customers, Services, Purchases, Accounting, Checks, Loans, Owners, Reports, and Settings should be able to adopt the same workspace composition without new sticky/fixed layout logic.

Keep these rules aligned with `docs/UI.md` and `CONTRIBUTING.md`.

## Constraints

- Preserve existing Lucide icon usage.
- Preserve Rial/Toman behavior.
- Preserve Jalali/date work already present when this task is implemented.
- Preserve motion/reduced-motion behavior.
- Preserve current navigation and order interactions.
- No business/domain/backend changes.
- No UI framework addition.
- No redesign of cards/tables/forms solely for this task.
- Avoid arbitrary CSS `top`, `bottom`, and `z-index` values scattered across feature files.

## Acceptance criteria

1. The global sidebar, top bar, and bottom status strip remain stable while page content scrolls.
2. Normal page scrolling happens inside the shared workspace rather than the document/body.
3. The large unused top whitespace visible in the current Order workspace is removed.
4. Dashboard, Orders list, and Order workspace use the same page-header/action geometry.
5. Important page actions remain available through the shared sticky page-header pattern where appropriate.
6. Order workspace tabs coordinate correctly with the sticky page header and do not overlap it.
7. Dashboard Quick Actions remain accessible through the shared persistent-bottom action pattern.
8. Persistent-bottom actions do not cover page content or the global bottom status strip.
9. Sticky surfaces use shared offsets/z-index/style conventions rather than page-specific constants.
10. There are no accidental nested/double vertical scrollbars in the primary tested screens.
11. Dropdowns, menus, focus states, and interactive controls render above/below sticky surfaces correctly.
12. The system is implemented as reusable layout primitives/styles suitable for future pages.
13. Existing navigation, order flows, currency behavior, dates, and icons continue to work.
14. `npm run build` and `git diff --check` pass.

## Visual verification

Run the application and test at several realistic desktop window heights, especially a smaller laptop-height window.

Verify:

- Dashboard at top and after significant scrolling
- Dashboard Quick Actions while scrolling
- Orders list at top and after scrolling
- New Order Draft at top and after scrolling past multiple sections
- Order workspace tabs while scrolling
- More/dropdown actions from sticky regions
- global bottom status strip with persistent page actions present
- collapsed and expanded sidebar

The UI should feel like one continuous desktop shell, not separate web pages placed inside the application.

## Validation

```sh
cd frontend
npm run build
cd ..
git diff --check
```

## Relevant documentation

- `docs/UI.md`
- `CONTRIBUTING.md`
- `tasks/M0-004-scroll-layout-and-currency-display.md`
