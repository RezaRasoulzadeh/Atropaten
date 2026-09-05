# M1-001 — Materials persistence foundation

## Objective

Introduce Atropaten's first real persisted vertical slice: SQLite-backed materials with migrations, canonical money/quantity representation, Go application/domain boundaries, Wails bindings, and a fully usable Materials workspace.

This task establishes the persistence conventions future M1+ features will build on. Keep the implementation simple and explicit.

## User-visible outcome

After this task, the user can open **Materials**, create/edit/archive materials, restart Atropaten, and see the same data still present.

The Materials page must replace any purely decorative/mock implementation with a functional table + inspector/editor using the existing project-wide workspace/sticky design system.

## Scope

### 1. SQLite foundation

Add the SQLite persistence foundation required by this slice.

Requirements:

- Use SQLite as defined in `docs/ARCHITECTURE.md`.
- Add versioned migrations; do not rely on runtime schema recreation or deleting a database.
- Ensure database initialization and migration application happen during app startup through infrastructure code, not inside Vue or Wails handlers.
- Choose a project-appropriate local application-data database path rather than placing authoritative runtime data in source-controlled project files.
- Keep database opening/migration logic isolated from domain/application logic.
- Enable foreign keys.
- Use transactions for multi-record writes where applicable.

Do not introduce an ORM unless there is a strong concrete benefit. Prefer `database/sql` and a focused SQLite driver.

### 2. Canonical money representation

Establish the Go representation for authoritative monetary values.

Rules:

- Authoritative money is stored as integer **Rial**.
- Do not use `float32`/`float64` for stored or authoritative monetary calculations.
- Toman remains presentation/input only (`1 toman = 10 rial`).
- Reuse the existing frontend currency preference/formatter behavior rather than creating a second currency model.
- Persist material cost values in Rial integers.

Keep the money type/utility small. Avoid speculative generalized currency frameworks.

### 3. Canonical quantity representation

Establish an explicit representation for material quantities suitable for inventory/pricing work.

Materials need:

- purchase unit
- consumption unit
- conversion factor between purchase and consumption units
- current stock expressed canonically enough for future inventory movements
- reorder/minimum stock level

Do not use binary floating point for authoritative quantity values if it can cause cumulative inventory/costing drift. Use a simple decimal/fixed-scale representation appropriate for Go + SQLite and document the chosen scale/contract in code.

The first implementation should support common examples such as:

- sheet → sheet (`A4 paper`)
- pack → sheet (`500 sheets per ream/pack`)
- kilogram → gram
- roll → meter
- piece → piece

Do not hardcode the system to those examples.

### 4. Material domain model

Implement a focused Material model with at minimum:

- ID
- name
- optional SKU/code
- category/type label where useful
- purchase unit
- consumption unit
- conversion factor
- physical/current stock
- reorder level
- average unit cost in Rial, based on the canonical consumption unit or with the basis made explicit
- optional preferred supplier name/reference placeholder if persistence of suppliers is not implemented yet
- notes
- active/archived state
- created/updated timestamps

Do not build purchasing, inventory movement ledger, suppliers, weighted-average purchase posting, or production consumption yet. Those come in later vertical slices.

### 5. Application services

Create application-layer operations for at minimum:

- list materials
- get material
- create material
- update material
- archive/reactivate material

Validation belongs in Go, including sensible rules for:

- non-empty name
- valid units
- positive conversion factor
- non-negative stock/reorder/cost values where appropriate

Keep Wails methods thin adapters over these application operations.

### 6. Wails integration

Expose only the operations the Materials UI needs.

Requirements:

- No SQL in Wails adapter methods.
- No authoritative business validation in Vue.
- Return clear structured errors/messages suitable for UI display.
- Keep bindings straightforward and typed.

### 7. Materials workspace

Implement the real Materials UI as a vertical slice, consistent with `docs/UI.md` and the shared sticky workspace system.

Use the established **dense table + right inspector/editor** pattern.

The page should include:

- sticky page header with title/context and primary `New material` action
- search/filter control
- active/archived filtering
- dense materials table
- selection state
- right-side inspector/editor for selected material
- clear empty state
- create-material workflow
- edit/save workflow
- archive/reactivate action
- success/error feedback

Suggested table columns:

- Material
- Purchase unit
- Consumption unit
- Physical stock
- Available/physical indicator (for now available may equal physical because reservations are not implemented yet; do not pretend reservations exist)
- Average cost
- Reorder level
- Low-stock status

### 8. Inspector/editor behavior

The inspector/editor should expose the fields implemented in this task and provide visible explanation of unit conversion.

Example:

`1 pack = 500 sheets`

When currency display is set to Toman, cost fields/display may show Toman through the existing presentation preference, but values sent to/stored by the backend must resolve to canonical Rial integers.

When Jalali presentation is appropriate for created/updated timestamps, use the existing shared date utility.

### 9. Low-stock state

For this task, low-stock can be derived simply from:

`physical stock <= reorder level`

Do not persist a redundant low-stock boolean.

### 10. Existing project-wide UI rules

This task must use the existing shared conventions established in M0:

- Lucide icons
- unified sticky workspace geometry
- intentional scroll ownership
- Rial/Toman presentation
- Jalali user-facing date presentation
- restrained motion
- consistent table/form/card/badge styling

Do not introduce a visually separate Materials design system.

## Persistence and migration expectations

At minimum include:

- migration metadata/version tracking
- initial materials schema migration
- startup migration application
- repository/store implementation
- tests that verify migration/open behavior on a temporary database
- tests that verify material persistence survives close/reopen

The exact schema is implementation-defined, but it must preserve integer Rial values and the chosen canonical quantity representation without lossy float conversion.

## Testing

Add focused Go tests for:

- migration application on a fresh database
- reopening an existing migrated database
- create/read/update/archive material persistence
- material validation
- money round-tripping exactly as integer Rial
- quantity/conversion round-tripping without binary-float drift

Frontend tests are optional unless an existing frontend testing setup makes them cheap, but the frontend build must pass.

## Non-goals

Do **not** implement in this task:

- purchases
- suppliers as a full domain
- inventory movement ledger
- weighted-average cost updates from purchases
- reservations
- production consumption
- service material requirements
- pricing engine
- accounting journal entries
- multi-location inventory behavior beyond avoiding schema/design choices that make it impossible later

## Acceptance criteria

1. Atropaten initializes a local SQLite database through versioned migrations.
2. Restarting the app preserves material records.
3. Money is authoritative integer Rial end-to-end in Go/SQLite.
4. Material quantity/conversion values use a non-lossy canonical representation suitable for inventory work.
5. Go owns material validation and application operations.
6. Wails methods are thin adapters with no SQL/business-rule duplication.
7. Materials page lists persisted materials from the backend.
8. User can create a material from the running UI.
9. User can edit a material and observe the persisted update after restart.
10. User can archive/reactivate materials.
11. Search/filter and active/archived filtering work.
12. Low-stock status is derived correctly.
13. Materials UI follows the shared sticky workspace/table/inspector patterns.
14. Rial/Toman presentation works without changing canonical stored Rial values.
15. Existing Dashboard/Orders behavior remains functional.
16. Relevant Go tests pass.
17. Frontend build passes.
18. `git diff --check` passes.

## Manual verification

Run Atropaten and verify:

1. Open Materials.
2. Create `A4 80gsm Paper` with purchase unit `pack`, consumption unit `sheet`, conversion `500`, stock and reorder values, and an average cost.
3. Save it and see it in the table immediately.
4. Switch Rial/Toman display and confirm only presentation changes.
5. Edit the material and save.
6. Archive it and verify active filtering.
7. Reactivate it.
8. Close and reopen Atropaten.
9. Confirm the material and edited values are still present.
10. Confirm Dashboard and Orders still render and navigate correctly.

## Validation

Run the relevant Go test suite plus:

```sh
cd frontend
npm run build
cd ..
git diff --check
```

If Wails is available in the implementation environment, also run the application and exercise the manual flow above.

## Relevant documentation

- `docs/PRODUCT.md`
- `docs/DOMAIN.md`
- `docs/ARCHITECTURE.md`
- `docs/UI.md`
- `MILESTONES.md`
- `CONTRIBUTING.md`
