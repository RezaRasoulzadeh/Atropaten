# M1-002 — Services and dynamic parameters

## Objective

Add persisted **Services + dynamic parameters** as the next M1 vertical slice.

A service must describe a sellable print-shop operation without requiring service-specific backend code. The user should be able to define a service, define its configurable parameters, edit/archive it, restart Atropaten, and see the same configuration preserved.

This task builds on M1-001's SQLite, migration, Go domain/application, Wails, money, quantity, and shared UI foundations.

## User-visible outcome

After this task, the user can open **Services** and manage persisted service definitions such as:

- Digital Print
- Business Cards
- Lamination
- Cutting
- Graphic Design

Each service can have an ordered set of configurable parameters, for example:

- `quantity` — integer, required
- `paper_size` — choice: A4 / A3
- `color_mode` — choice: B&W / Color
- `sides` — choice: Single / Double
- `lamination` — boolean
- `design_hours` — decimal

The Services workspace must be a real backend-connected workflow, not mock data.

## Scope

### 1. Persistence and migrations

Extend the existing SQLite migration system; do not replace or bypass it.

Persist services and their parameter definitions with versioned migrations.

At minimum, preserve:

### Service

- ID
- name
- optional code/SKU-like identifier
- optional category
- description/notes
- active/archived state
- created/updated timestamps

### Service parameter

- ID
- service ID
- stable key/name used internally, e.g. `paper_size`
- user-facing label, e.g. `Paper size`
- parameter type
- required flag
- display/order position
- optional default value where applicable
- optional choice options for choice parameters
- optional numeric minimum/maximum where applicable
- optional unit/suffix label where useful
- active state if needed for safe future editing

Use foreign keys and deterministic ordering.

Do not store parameter configuration in opaque JSON if ordinary relational rows provide clearer validation/query/update behavior. Small structured fields such as choice options may use a simple representation if justified, but keep the model explicit and easy to migrate.

### 2. Supported parameter types

Implement these generic parameter kinds:

- integer
- decimal
- boolean
- choice
- material-reference

The domain must remain generic. Do not create backend types such as `PaperSizeParameter`, `ColorModeParameter`, `BusinessCardService`, etc.

Rules:

- parameter key must be non-empty and stable enough to be referenced by future pricing/order snapshots
- keys must be unique within a service
- labels must be non-empty
- choice parameters require at least one valid option
- default choice value, if set, must belong to the available options
- numeric min/max must be valid and `min <= max`
- boolean defaults must be valid booleans
- material-reference parameters represent selecting a persisted material but do not yet create service material-consumption rules
- display order must be deterministic

Use the same non-binary-floating-point discipline established for authoritative numeric values. If decimal parameter metadata/defaults need persistence, use the existing fixed-scale/decimal approach or another explicit exact representation consistent with M1-001.

### 3. Domain model and application services

Add focused Go domain/application operations for at minimum:

- list services
- get service with ordered parameters
- create service
- update service metadata
- archive/reactivate service
- add parameter
- update parameter
- remove/archive parameter
- reorder parameters

A service create/update operation may accept the whole parameter collection transactionally if that produces a simpler and safer interface. Prefer atomic replacement/update over sequences that can leave half-edited service definitions.

Validation belongs in Go.

Keep repository interfaces/application services independent from Wails.

### 4. Transactional integrity

Editing a service and its parameter definitions must not leave partial state.

Use transactions for multi-row service-definition writes. A failed validation/write must preserve the prior valid definition.

Add tests for rollback/atomicity where meaningful.

### 5. Wails integration

Expose thin bindings required by the Services workspace.

Requirements:

- no SQL in Wails methods
- no authoritative validation duplicated in Vue
- typed request/response shapes
- clear errors suitable for the UI
- reuse the existing startup/store/application wiring rather than creating an independent database connection or service locator

### 6. Services workspace

Replace any mock/decorative Services page with a persisted workspace following the project-wide UI system.

Use a dense master/detail layout:

- sticky page header
- `New service` primary action
- search
- active/archived/all filter
- dense service list/table on the left/main area
- right inspector/editor for the selected service
- clear empty/loading/error states
- archive/reactivate actions
- success feedback

Suggested service-list fields:

- Service
- Category
- Parameters count
- Status
- Updated

Use existing shared components/styles where practical. Do not introduce a separate visual language.

### 7. Service editor

The service editor must support editing service metadata and parameter definitions in one coherent workflow.

Parameter editor requirements:

- add parameter
- change label/key/type
- required toggle
- default value where applicable
- integer/decimal min/max where applicable
- choice option management
- material-reference type visible as a selectable parameter kind
- reorder parameters
- remove parameter before save
- clear parameter-type-specific controls

Do not require the user to edit JSON manually.

When a parameter changes type, normalize/remove incompatible old configuration so invalid hidden state is not persisted.

### 8. Material-reference parameter behavior

For `material-reference` parameters:

- load selectable active materials from the existing Materials application layer/binding as needed
- the parameter definition itself may optionally constrain or default to a material if the domain design supports it cleanly
- do not implement consumption quantity, waste, reservations, inventory movements, or cost calculation in this task

The purpose here is only to establish that a service may ask the operator to select a material dynamically.

Avoid coupling the service domain directly to Vue-generated shapes.

### 9. Generic example configurations

The implementation must prove the schema/model is generic by supporting at least these persisted examples without service-specific backend code:

#### Digital Print

- quantity — integer, required, minimum 1
- paper_size — choice: A4, A3
- color_mode — choice: B&W, Color
- sides — choice: Single, Double
- paper — material-reference

#### Graphic Design

- estimated_hours — decimal, required, minimum 0

#### Lamination

- quantity — integer, required
- size — choice

These examples may be entered manually during verification; do not hardcode them as built-in service classes.

### 10. Editing and historical safety direction

M2 will snapshot resolved service configuration into quotes/orders. This task does not implement order snapshots yet, but avoid designs that make that difficult.

Parameter identity/key should be stable enough that future order-item snapshots can preserve the exact parameter/value pair even when the catalog changes later.

Do not add premature version-history infrastructure unless required for correctness now.

### 11. Existing project-wide rules

Preserve all established conventions:

- app shell owns viewport
- workspace owns normal vertical scrolling
- shared sticky workspace/header geometry
- shared bottom/action surfaces where appropriate
- Lucide icons
- Jalali date presentation
- Rial/Toman behavior where money appears
- restrained motion
- dense professional desktop layout
- no one-off fixed/sticky hacks
- Go owns domain/application rules
- SQLite remains infrastructure
- Wails is a thin adapter

## Testing

Add focused Go tests for:

- fresh migration adds service/parameter schema correctly
- reopen preserves services and ordered parameters
- create/read/update/archive/reactivate service
- parameter key uniqueness within service
- valid/invalid choice parameter rules
- numeric min/max validation
- exact decimal/fixed-scale parameter metadata round-trip where applicable
- deterministic parameter ordering
- transactional service-definition update/rollback
- material-reference parameter accepts valid definition without implementing consumption logic

Frontend build must pass.

## Non-goals

Do **not** implement in this task:

- pricing rules
- cost components
- pricing engine
- service material consumption requirements
- machine requirements
- labor costing
- overhead/waste calculations
- quotes/orders snapshots
- inventory reservations or movements
- production jobs
- accounting
- service-specific Go structs/algorithms
- advanced service-definition version history

## Acceptance criteria

1. Existing SQLite database upgrades through a new versioned migration without losing M1-001 material data.
2. Services persist across application restarts.
3. Each service has an ordered persisted parameter collection.
4. Supported parameter kinds are integer, decimal, boolean, choice, and material-reference.
5. Go enforces service/parameter validation.
6. Parameter keys are unique per service.
7. Choice/default/min/max rules are enforced correctly.
8. Multi-row service definition writes are transactional.
9. Wails bindings contain no SQL/business-rule duplication.
10. Services UI is connected to the backend and contains no authoritative mock dataset.
11. User can create/edit/archive/reactivate services.
12. User can add/edit/remove/reorder parameters without editing raw JSON.
13. Generic Digital Print and Graphic Design definitions can be represented without service-specific backend code.
14. Existing Materials data/workflow remains intact after migration.
15. Existing Dashboard/Orders/Materials navigation remains functional.
16. Relevant Go tests pass.
17. Frontend build passes.
18. `git diff --check` passes.

## Manual verification

Run Atropaten and verify:

1. Confirm the existing material created during M1-001 is still present after migration.
2. Open Services.
3. Create `Digital Print`.
4. Add parameters for quantity, paper size, color mode, sides, and material reference.
5. Reorder at least two parameters and save.
6. Close/reopen the editor and verify ordering/defaults/options persist.
7. Edit one choice list and save.
8. Archive the service and verify filters.
9. Reactivate it.
10. Create `Graphic Design` with a decimal `estimated_hours` parameter.
11. Restart Atropaten and confirm both definitions persist.
12. Confirm Dashboard, Orders, and Materials still work.

## Validation

Run:

```sh
go test ./...
cd frontend
npm run build
cd ..
git diff --check
```

If Wails is available, run the application and perform the manual verification above.

Before finishing, review the complete diff for architecture boundary leaks, migration safety, duplicate validation, or service-specific logic.

## Relevant documentation

- `docs/PRODUCT.md`
- `docs/DOMAIN.md`
- `docs/ARCHITECTURE.md`
- `docs/UI.md`
- `MILESTONES.md`
- `CONTRIBUTING.md`
- `tasks/M1-001-materials-persistence-foundation.md`
