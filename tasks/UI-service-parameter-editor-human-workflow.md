# Service parameter editor — human-friendly workflow

## Goal
Redesign only the **Parameters** part of Service Add/Edit so configuring multiple service parameters is fast and understandable for a print-shop operator, while preserving the existing parameter domain model and persistence behavior.

## Problem
The current parameter editor exposes each parameter as a dense schema/configuration form. This represents the data correctly but becomes difficult to scan and manage when a service has several parameters. The user should think in terms of order options such as **Paper type**, **Quantity**, **Color**, **Finishing**, etc., not implementation/schema fields.

## Target interaction
Use a master/detail parameter editor inside the Service workspace.

### Parameter list
- Show all parameters as compact rows/cards in an ordered list on the left/upper portion of the Parameters workspace.
- Each row should primarily show the human label, a human-readable type summary, required/optional state when useful, and a concise configuration summary.
- Examples: `Paper type — Choice · 5 options`, `Quantity — Number · 50–50,000`, `Material — Material`, `Duplex — Yes / No`.
- Make ordering obvious and easy. Preserve the existing ordering semantics.
- Selecting a parameter opens that one parameter in the editor; do not expand every parameter inline.
- `Add parameter` creates/selects a new parameter and focuses its editor.
- Removal remains explicit and destructive.

### Selected parameter editor
- Edit only the selected parameter at a time.
- Lead with human concepts: **Label**, **Type**, **Required**, then type-specific settings.
- Keep existing supported domain types unless the current backend/model already supports more.
- Translate technical enum names into human-facing labels where possible.
- Show only fields relevant to the selected type.
- Choice parameters: make options a compact editable/reorderable list rather than a long stack of generic form rows.
- Numeric parameters: group min/max/default/unit together.
- Boolean parameters: simple default selection.
- Material-reference parameters: simple default-material selection.
- Preserve stable parameter keys, but de-emphasize the key as an advanced/internal field. If safe with current behavior, generate a sensible key for new parameters from the label and allow it to be edited under an Advanced disclosure. Do not change existing keys automatically.

### Order preview
- Add a compact live preview beside the editor when space permits, showing how the current service parameters will actually appear when adding the service to an Order.
- Preview controls should be representative only and must not mutate order/domain data.
- At narrower workspace widths, move the preview below rather than squeezing the editor.

## UX constraints
- Multiple parameters must remain first-class; do not simplify the domain by reducing the number of parameters.
- Avoid nested accordion-after-accordion interaction.
- Avoid exposing all configuration fields for all parameters simultaneously.
- Optimize for scanning: an operator should understand a service's parameter structure from the compact list without opening every item.
- Respect Atropaten's existing dark DaisyUI visual language, typography/control metrics, neutral panels, and Amber action/focus treatment.
- Keep the layout responsive to the actual available workspace width.
- Preserve keyboard accessibility and existing form/error behavior.

## Scope
Primary files expected:
- `frontend/src/features/services/ServicesView.vue`
- `frontend/src/features/services/ServiceParameterEditor.vue`
- supporting Services frontend files only if required.

Do not redesign Cost components in this task. That will be handled separately after the Parameters interaction is accepted.

Do not change backend/domain contracts, parameter persistence, pricing semantics, service lifecycle behavior, or unrelated UI.

## Review contract
The user is live-reviewing the running Wails app. Implement the UI change and stop for visual/interaction review. Do not spend time on builds, automated tests, diff audits, commits/pushes, documentation cleanup, or unrelated work unless explicitly requested.