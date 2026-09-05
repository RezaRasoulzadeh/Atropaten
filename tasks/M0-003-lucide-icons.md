# M0-003 — Standardize the UI on Lucide icons

## Objective

Adopt `lucide-vue-next` as Atropaten's standard UI icon set and replace the temporary/custom/text-symbol icons introduced during the desktop foundation work.

This is a focused visual-foundation task. Do not redesign the application or change business behavior.

## Scope

### Dependency

Add `lucide-vue-next` to the frontend dependencies using npm so `package.json` and the lockfile remain consistent.

### Icon migration

Replace existing application UI icons with appropriate Lucide Vue components wherever Lucide provides a suitable semantic icon, including:

- Sidebar navigation
- Sidebar collapse/expand control
- Search
- Shop selector affordance
- Notifications
- KPI cards
- Production/dashboard actions
- Financial alerts
- Low-stock indicators where an icon is appropriate
- Recent transactions
- Quick actions
- Buttons and toolbar actions
- Orders/workspace UI that exists on `main` when this task is implemented

Replace Unicode/text glyphs used as pseudo-icons such as arrows, chevrons, ellipses, plus signs, stars, information marks, and similar symbols when a Lucide icon is appropriate.

### AppIcon cleanup

Remove `AppIcon.vue` if it is no longer necessary after the migration. If a small wrapper remains genuinely useful, it must use Lucide internally rather than maintain a custom SVG/icon catalog.

Do not create another parallel icon system.

### Consistency

Establish consistent icon conventions:

- Default navigation/action icon size around 16–18px
- Compact icons may use 14–16px
- Larger visual indicators may use 18–20px
- Use Lucide's normal stroke style consistently; do not mix arbitrary stroke widths without a specific visual reason
- Icons inherit semantic foreground color from their surrounding component where practical
- Decorative icons must not create unnecessary accessibility noise
- Icon-only interactive controls require an accessible label/title

Preserve the existing restrained Atropaten visual language. Lucide icons should improve clarity rather than make the interface more decorative.

## Constraints

- Use `lucide-vue-next`, not copied Lucide SVG markup.
- Do not add a second icon library.
- Do not redesign layouts merely to accommodate icons.
- Do not replace meaningful text labels with icon-only controls unless the action remains immediately understandable and accessible.
- Keep bundle imports tree-shakeable by importing only icons that are used.
- Preserve existing UI behavior.

## Acceptance criteria

1. `lucide-vue-next` is an explicit frontend dependency.
2. Primary application navigation uses Lucide icons.
3. Toolbar, dashboard, quick-action, and existing order-workspace icons use Lucide wherever appropriate.
4. Temporary Unicode/text pseudo-icons are removed wherever Lucide has an appropriate equivalent.
5. There is no duplicate custom icon catalog alongside Lucide.
6. Icon sizing and stroke appearance are visually consistent across the application.
7. Icon-only buttons remain accessible.
8. Existing shell/sidebar/navigation interactions continue to work.
9. The application remains visually consistent with `docs/UI.md`.
10. `npm run build` and `git diff --check` pass.

## Visual verification

Run Atropaten and inspect at minimum:

- Expanded and collapsed sidebar
- Top toolbar
- Dashboard KPI cards
- Production queue and alerts
- Quick actions
- Orders list/workspace if already implemented

Confirm icons are crisp, aligned, semantically appropriate, and do not change existing spacing/layout unexpectedly.

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
