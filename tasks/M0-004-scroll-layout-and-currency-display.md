# M0-004 — Fix scrollable workspace behavior and currency display rules

## Objective

Refine Atropaten's desktop layout so page-level actions remain available while content scrolls, and establish the application's currency display rule for Rial/Toman without changing the canonical stored monetary unit.

This is a UI-foundation task. Preserve the current visual language and existing screen behavior.

## Part 1 — Scrollable workspace and sticky/fixed actions

Audit the current shell, Dashboard, Orders list, and Order workspace and correct the scroll ownership/layout model.

### Layout rule

The main application shell should remain fixed to the desktop window. Scrolling should happen inside the intended workspace/content region rather than causing important page controls to disappear unnecessarily.

Use sticky/fixed positioning deliberately:

- Global shell elements such as sidebar, top application toolbar, and bottom status bar should remain visible.
- Page headers that contain primary actions, filters, search, or important context should remain sticky within the workspace when useful.
- Dashboard Quick Actions currently located at the bottom of the content should become a persistent action surface at the bottom of the workspace instead of requiring the user to scroll to reach them.
- The persistent Quick Actions surface must not cover dashboard content. Reserve appropriate layout space/padding for it.
- Orders page primary actions/filter controls should remain accessible while the order table scrolls where practical.
- Order workspace header/context and important primary actions should remain accessible while tab content scrolls where practical.
- Do not make every small element sticky. Sticky behavior should be limited to controls that materially help desktop workflows.

### Scrolling constraints

- Avoid nested scrollbars unless a contained region genuinely requires one.
- Avoid page-level horizontal scrolling at representative desktop/laptop widths.
- Tables may use their own horizontal overflow only when necessary for dense data.
- Sticky elements must respect the top toolbar and bottom status/action areas and must not overlap each other.
- Sticky surfaces should have proper backgrounds/borders/shadows so content scrolling underneath remains visually clear.
- Preserve keyboard focus visibility and accessibility.

### Dashboard Quick Actions

Refactor Quick Actions into a persistent bottom action bar/surface inside the workspace.

Requirements:

- Always visible while Dashboard is active.
- Sits above the global bottom status bar.
- Does not obscure dashboard content.
- Keeps the existing actions and behavior.
- Matches the current restrained desktop UI rather than looking like a mobile floating action bar.
- Adapts cleanly when the sidebar is collapsed.

## Part 2 — Rial/Toman display preference

Atropaten is an Iranian print-shop application. Replace the current dollar-based mock currency presentation with Rial/Toman semantics.

### Canonical monetary unit

All persisted monetary data must be represented canonically in **Iranian Rial (IRR)**.

This is a project invariant for future backend/persistence work:

- Stored monetary values are Rial.
- Business calculations operate on canonical Rial values unless explicitly documented otherwise.
- Toman is a presentation/input convenience only.
- `1 toman = 10 rial`.
- Never store a mixture of Rial and Toman values in domain/persistence data.

This task does not need to introduce SQLite persistence if persistence is not already part of the current milestone. It must establish the frontend representation and code structure so later backend work can keep Rial canonical.

### User display preference

Add a simple user-facing currency-unit preference with two options:

- Toman
- Rial

The default may be Toman for the current UI unless existing project documentation specifies otherwise.

For this M0 task, local frontend state is sufficient. Do not introduce backend persistence solely for this preference.

### Formatting

Create one reusable money formatting/conversion utility or composable rather than scattering `/ 10`, `* 10`, suffixes, or formatting logic across components.

Requirements:

- Input to the formatter is always a canonical Rial amount.
- Rial display shows the Rial amount.
- Toman display divides the canonical Rial amount by 10.
- Use readable thousands separators.
- Use a clear unit label/suffix such as `ریال` / `تومان` or an appropriate consistent English-language equivalent if the current interface remains English.
- Do not use `$` for Atropaten monetary values.
- Dashboard KPIs, status bar balances, alerts, recent transactions, Orders mock data, Order totals, payments, and other currently implemented monetary surfaces must use the shared formatter.
- Switching the display preference updates all visible monetary values consistently.

### Mock data rule

Convert existing mock monetary constants so their source values are canonical Rial amounts. Components must not contain values preformatted as dollar strings.

Prefer numeric values such as:

```ts
amountRial: 248_600_000
```

rather than:

```ts
amount: '$24,860'
```

The exact mock values may be adjusted to realistic Iranian print-shop amounts as long as the UI remains useful for visual testing.

## Architecture constraints

- Keep page/view components separated; do not move screen logic back into a monolithic `App.vue`.
- Keep scroll/layout primitives reusable where practical.
- Keep money formatting centralized.
- Do not add a state-management library solely for the currency preference.
- Do not add SQLite or backend persistence solely for this task.
- Do not introduce floating/mobile-oriented UI patterns inconsistent with `docs/UI.md`.

## Acceptance criteria

1. Global application chrome stays visible while workspace content scrolls.
2. Dashboard Quick Actions remain visible without scrolling to the end of the dashboard.
3. Persistent Quick Actions do not overlap or hide dashboard content.
4. Important page header actions remain accessible during scrolling where appropriate.
5. Orders and Order workspace remain usable with long content/table data.
6. No accidental nested-scroll or overlapping sticky-region regressions are introduced.
7. Existing desktop/laptop layouts do not gain page-level horizontal overflow.
8. Dollar signs are removed from currently implemented Atropaten business data.
9. User can switch displayed monetary unit between Toman and Rial.
10. All currently implemented monetary displays update consistently when the unit changes.
11. Canonical mock/source monetary values are Rial.
12. Rial/Toman conversion and formatting logic is centralized and not duplicated across screens.
13. The code/documentation makes the invariant explicit: persisted money is always Rial; Toman is presentation/input only.
14. Existing navigation and actions continue to work.
15. `npm run build` and `git diff --check` pass.

## Visual verification

Run the app and inspect at minimum:

- Dashboard at top, middle, and bottom scroll positions
- Persistent Quick Actions
- Expanded/collapsed sidebar
- Orders list with enough rows to scroll
- Order workspace with long tab content
- Page-header sticky behavior
- Bottom status bar interaction with sticky/fixed surfaces
- Rial display
- Toman display

Verify that no sticky/fixed surface obscures meaningful content and that switching currency units does not change underlying canonical values.

## Validation

```sh
cd frontend
npm run build
cd ..
git diff --check
```

## Relevant documentation

- `docs/UI.md`
- `docs/ARCHITECTURE.md`
- `docs/DOMAIN.md`
- `CONTRIBUTING.md`
