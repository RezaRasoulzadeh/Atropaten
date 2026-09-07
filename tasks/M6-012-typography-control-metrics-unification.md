# M6-012 — Typography and control metrics unification

## Objective
After M6-011 geometry is complete, make typography, font metrics, control heights, text/icon alignment, labels, selects, and focus states visually consistent across the entire application.

This task is specifically about **font and control metrics**, not page redesign.

## Typography baseline
Use locally vendored Vazirmatn everywhere.

Do not add or restore:
- `ascent-override`;
- `descent-override`;
- `line-gap-override`;
- `transform: translateY(...)` font fixes;
- negative margins to visually move text;
- asymmetric top/bottom padding as font compensation.

Use the font's real metrics and solve alignment by consistent control sizing, flex alignment, line-height, icon dimensions, and padding.

Define and apply a small coherent type scale for:
- workspace title;
- section/panel title;
- body;
- control value;
- label;
- secondary metadata/help text;
- table header;
- badge/status text.

Avoid arbitrary page-local font sizes. Labels must remain visually subordinate to field values.

## Control contract
Audit all inputs, textareas, native selects, search fields, date fields, money/quantity fields, custom triggers, buttons, icon buttons and tabs.

Use DaisyUI defaults as the foundation and converge on shared heights/density.

Requirements:
- equivalent controls have identical visual height;
- control text is optically centered with Vazirmatn without metric hacks;
- native select text is start-aligned, never horizontally centered;
- select arrow/indicator has stable alignment and does not displace text;
- button icon + label combinations are visually centered without `:has()` grid tricks;
- icons use consistent common sizes/stroke widths by role;
- text fields, select values, buttons and toolbar controls align on the same baseline when placed in one row;
- placeholder and disabled states remain readable;
- no arbitrary per-page control padding.

## Focus and validation
Preserve the required editable-control focus contract:
- rest: 1px solid neutral border;
- focus: 1px dashed Amber primary border;
- same border width, no layout shift;
- no outline, ring, glow or shadow;
- validation error/success color remains visible.

Do not apply dashed focus treatment to cards or containers. Buttons may retain an appropriate keyboard-visible DaisyUI focus treatment unless it conflicts with the theme; do not remove keyboard visibility globally.

## Browser metric check
Add/extend the interaction/visual audit so representative controls are measured using `getBoundingClientRect()` and computed style at 1024/1280/1600.

At minimum compare:
- toolbar search/select/button;
- normal form input/select/button;
- textarea;
- Jalali date field;
- money field;
- dialog controls;
- table/register action buttons.

Inspect screenshots, not only numeric box metrics. Specifically look for Vazirmatn appearing lower or higher inside controls.

## Page coverage
Audit every feature page for local font/line-height/height classes that violate the shared system. Replace them with shared primitives or coherent Tailwind classes.

Update `docs/FRONTEND_UI_AUDIT.md` with typography/control status for every context.

## Constraints
- M6-011 geometry must remain stable.
- Do not redesign page information architecture.
- Preserve dark Amber theme and backend behavior.
- No new global CSS beyond genuinely global font/control infrastructure.

## Acceptance
- Vazirmatn uses natural metrics only.
- No visible systemic vertical text misalignment in controls.
- Selects are consistently start-aligned.
- Equivalent controls share height/density.
- Labels/metadata/type hierarchy are coherent across pages.
- Focus contract is consistent without layout shift.
- Real browser screenshots cover representative controls at 1024/1280/1600.
- `go test ./...`, `cd frontend && npm run build`, `git diff --check` pass.
- Commit/push to main and report SHA.