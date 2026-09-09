# Service parameter editor — human-friendly multiple-parameter workflow

## Goal
Redesign only the **Parameters** portion of Service Add/Edit so a print-shop operator can configure and understand many parameters without reading a long sequence of schema forms.

Preserve the current parameter domain model, ordering, validation, persistence, pricing references, and Service save behavior.

## Current baseline on main
The latest Services UI already improved terminology and creation:

- `ServicesView.vue` uses a human-facing `Add an operator input…` selector.
- New parameters can be created directly by type through `addParameter(type)`.
- Current supported types are `integer`, `decimal`, `choice`, `material-reference`, and `boolean`.
- New parameter labels can synchronize their generated key through `syncParameterKey(parameter)`.
- Validation state is surfaced through `validationAttempted` / `show-errors`.
- Parameters still render as a `v-for` of `ServiceParameterEditor` inside one bordered container.
- Each `ServiceParameterEditor` is still an expandable details/schema editor, so a service with several parameters becomes a long accordion-like document.

Keep the good human-facing changes above. Replace only the interaction model that becomes complex with multiple items.

## Target UX
Use a **master/detail parameter builder**.

The operator should see the service's parameter structure first and edit one parameter at a time.

### 1. Compact parameter list
Create a compact ordered list of all parameters.

Each item should be understandable without opening it and should show:

- Human label as the primary text.
- Human-readable type, never raw enum text when a clearer label exists.
- Concise configuration summary.
- Required/optional state only where useful.
- Clear selected state.
- Reorder affordance.
- Explicit remove action, visually secondary until used.

Examples:

- `Quantity` — **Number** · `50–50,000 sheets`
- `Paper type` — **Choices** · `5 options`
- `Paper` — **Material** · `Default: Gloss 300gsm`
- `Duplex` — **Yes / No** · `Default: Yes`

Do not display internal keys as primary list information.

Selecting a list item edits that parameter. Do not expand multiple parameter forms simultaneously.

### 2. Add parameter interaction
Keep the current type-first creation concept because it is already more human-friendly.

`Add parameter` should present the existing types with operator-facing names:

- Quantity / whole number
- Decimal measurement
- Choices
- Material / paper
- Yes / No

After choosing a type:

- create the parameter using the existing `addParameter(type)` behavior,
- select the new parameter immediately,
- focus the useful first field (normally Label),
- do not create a second nested dialog unless clearly necessary.

### 3. Selected parameter editor
Show one focused editor beside or below the list.

Lead with fields in this order:

1. Label
2. Type
3. Required
4. Type-specific configuration
5. Advanced/internal settings

The editor must adapt to type rather than showing a generic schema.

#### Integer / decimal
Present the related settings as one understandable group:

- minimum
- maximum
- default
- unit/suffix

Use human labels such as `Minimum`, `Maximum`, `Default`, `Unit`.

#### Choice
This is especially important for print services.

Show a compact option builder:

- one option per concise row,
- direct inline editing,
- easy Add option,
- easy reorder,
- delete as an icon/action on the row,
- default choice visibly identifiable/selectable.

Do not make five choices look like five separate full forms.

#### Boolean
Use a simple default control:

- Not set
- Yes
- No

Do not expose `true` / `false` as operator-facing wording.

#### Material reference
Use the existing material records and show:

- optional default material,
- clear material name/SKU presentation,
- short explanation that the operator may choose an active material at order time.

### 4. Parameter key
The key remains necessary because pricing/components may reference it.

Preserve existing keys exactly during edit unless the user changes them.

For new parameters, preserve the current automatic label → key synchronization behavior where safe.

Move the key out of the primary workflow into an **Advanced** disclosure or secondary area. Explain it briefly as an internal stable identifier.

Once the user explicitly edits/overrides a key, do not silently regenerate it from later label changes.

Do not break current pricing-rule or component references to parameter keys.

### 5. Live Order preview
Add a compact **Order preview** for the Parameters section.

Its purpose is not decoration; it lets the operator immediately understand what the configured parameter list becomes during Order entry.

Show the parameters in their current order using representative controls:

- integer/decimal → numeric-looking input
- choice → select
- material reference → material select
- boolean → Yes/No control

Use defaults where configured.

The preview is local UI state only. It must not create/update an Order or invoke unrelated persistence.

The preview should update immediately when label, ordering, type, choices, default, required state, etc. change.

### 6. Responsive composition
At useful desktop width:

`[ compact parameter list ] [ selected parameter editor ] [ compact order preview ]`

Do not force three columns when the available Service workspace is too narrow.

At medium widths, prefer:

`[ parameter list ] [ selected editor ]`
`[ order preview below ]`

At narrow widths, stack all three cleanly.

Base this on actual available workspace width, not blindly on viewport `lg/xl` breakpoints.

## Interaction details
- One selected parameter at a time.
- Maintain a valid selection when parameters are added, removed, or reordered.
- If the selected parameter is removed, select the nearest remaining item.
- Preserve existing ordering semantics and `moveParameter` behavior.
- Preserve current validation/error messages. A validation error in a non-selected parameter should make that list item visibly indicate an issue and allow the user to select it.
- Preserve keyboard accessibility.
- Avoid nested accordions and avoid modal-per-parameter interaction.
- Do not require a separate Save Parameter persistence operation: the Service remains the atomic saved entity unless the current architecture already requires otherwise.

## Visual direction
Stay inside Atropaten's established UI:

- dark DaisyUI surfaces,
- neutral borders/panels,
- Amber only for active/action/focus emphasis,
- compact desktop density,
- existing typography and control heights,
- minimal shadows/effects,
- clear selected item state.

The visual concept discussed with the user is a step-like Service editor with a compact parameter list, focused editor, and live Order preview. Implement the concept using the existing Atropaten component language rather than introducing a separate visual system.

## Scope
Primary expected files:

- `frontend/src/features/services/ServicesView.vue`
- `frontend/src/features/services/ServiceParameterEditor.vue`
- `frontend/src/features/services/useServicesWorkspace.ts` only for local editor-selection/key-sync helpers if required
- small Services-only presentation component(s) may be added if they make the structure clearer

Do **not** redesign Cost components in this task. Cost components are the next separate UX pass after Parameters is accepted.

Do not change:

- Go/backend/domain contracts
- persisted parameter schema
- service save semantics
- pricing semantics
- component formulas/references
- service lifecycle behavior
- unrelated workspaces

## Review contract
The user is live-reviewing the running Wails application.

Implement this Parameters UX, then stop so the user can interact with it.

Do not spend time on builds, automated tests, diff audits, git commits/pushes, documentation cleanup, or unrelated refactoring unless explicitly requested.