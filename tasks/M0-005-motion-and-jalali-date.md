# M0-005 — Smooth UI motion and Jalali date foundation

## Objective

Polish the existing Atropaten desktop experience with restrained, smooth motion and establish a shared Jalali (Solar Hijri / Persian) date presentation layer for the Iranian print-shop workflow.

This remains a frontend-foundation task. Preserve the existing design, information density, navigation, and business behavior.

## 1. Motion system

Add subtle motion where it improves spatial continuity or interaction feedback.

### Apply motion to appropriate existing interactions

At minimum inspect and improve:

- Sidebar collapse / expand
- View and workspace transitions
- Dashboard / Orders navigation
- Opening and closing inspectors, panels, dropdowns, menus, or overlays that currently exist
- Tabs and active-state changes
- Sticky / persistent action surfaces from M0-004
- Hover / pressed / focus feedback for interactive cards and buttons where useful
- Status/filter changes where a transition improves comprehension

### Motion rules

- Motion must be fast and restrained for a dense desktop application.
- Prefer CSS transitions and Vue's built-in transition primitives.
- Do not add an animation framework unless there is a strong technical need.
- Favor opacity and transform transitions over expensive layout animation.
- Avoid excessive entrance animations, bouncing, spring effects, decorative motion, or anything that slows repetitive operational work.
- Do not animate large tables row-by-row.
- Avoid visible layout jumps when sidebar, sticky regions, or workspaces change state.
- Respect `prefers-reduced-motion: reduce`; important interactions must remain fully usable with motion disabled.
- Centralize common duration/easing values as CSS design tokens rather than scattering arbitrary timing values.

Suggested motion scale should generally remain around 120–250ms depending on interaction size.

## 2. Jalali date presentation

Atropaten is intended for an Iranian shop, so user-facing operational dates should support the Jalali / Solar Hijri calendar.

### Rules

- Establish one shared date formatting/parsing boundary rather than formatting dates ad hoc inside components.
- User-facing dates should default to Jalali presentation.
- Use Persian/Iranian conventions suitable for the UI while preserving the existing English application language unless/until localization is implemented separately.
- Keep canonical application date/time values calendar-neutral and suitable for persistence/interchange; do **not** design future database storage around formatted Jalali strings.
- Conversion to Jalali belongs at the presentation/input boundary.
- Do not silently reinterpret timestamps when converting calendars.
- Keep date-only values distinct from timestamps where that distinction matters.

### Current UI migration

Replace hardcoded/mock Gregorian date presentation throughout currently implemented screens where the date represents an application/business date, including as applicable:

- Dashboard current date
- Orders list
- Order workspace
- promised / due dates
- transaction dates
- production dates
- any other visible mock business dates

Relative labels such as `Today` may remain where appropriate, but detailed/explicit date presentation should use the shared Jalali formatter.

### Date input readiness

Prepare the frontend date API so future forms can accept/select Jalali dates without forcing domain/database models to store Jalali-formatted text.

A full custom date-picker component is **not** required in this task unless an existing implemented workflow already requires date entry. Do not build a large calendar component just to satisfy this foundation task.

If a small dependency is needed for reliable Jalali conversion, choose a focused, maintained library rather than implementing calendar arithmetic manually. Keep it isolated behind Atropaten's own date utility API so the rest of the UI is not coupled directly to a third-party library.

## 3. Architecture

Create shared frontend primitives/utilities for:

- motion timing/easing tokens
- Jalali date formatting
- conversion needed by existing UI
- future date input integration

Components should consume these shared boundaries rather than each inventing formatting or transition rules.

Do not introduce backend persistence or SQLite in this task.

## Acceptance criteria

1. Sidebar collapse/expand feels smooth and does not visibly jump.
2. Major existing view/workspace changes have restrained transition feedback where useful.
3. Existing panels/tabs/interactive surfaces use consistent motion rather than unrelated timings.
4. Motion remains fast enough for repetitive desktop use.
5. `prefers-reduced-motion` is respected.
6. Common motion timing/easing is represented by shared design tokens.
7. User-facing explicit business dates on implemented screens use Jalali presentation.
8. Jalali formatting is centralized behind a shared utility/API.
9. Canonical date values are not represented as formatted Jalali strings in application models.
10. Future persistence can use standard canonical dates/timestamps without depending on the UI calendar.
11. Existing navigation, scrolling, sticky surfaces, currency switching, and order workflows continue to work.
12. `npm run build` and `git diff --check` pass.

## Visual verification

Run Atropaten and manually inspect at minimum:

- Sidebar expand/collapse
- Dashboard ↔ Orders navigation
- Order list ↔ order workspace/draft workflow
- Tabs and existing interactive panels
- Sticky actions while scrolling
- Dashboard date
- Order/promised/due dates
- Transaction dates
- Reduced-motion behavior if practical

Motion should make state changes easier to follow without making the application feel animated for its own sake.

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
- `CONTRIBUTING.md`
