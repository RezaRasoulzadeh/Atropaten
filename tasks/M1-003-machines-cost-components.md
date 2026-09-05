# M1-003 — Machines, rates, and reusable cost components

## Objective

Add the persisted cost-input layer required before the Go pricing engine: reusable Machines with canonical rates plus generic Service cost components for material, machine, labor, outsourced, fixed, overhead, waste, and manual/custom costs.

This task must remain configuration-driven. Do not implement service-specific pricing code or the final pricing engine yet.

## User-visible outcome

After this task, the user can:

- open **Machines**, create/edit/archive machine definitions and rates, restart Atropaten, and see them persist;
- open **Services**, attach ordered cost components to a service, edit/reorder/remove them, and see readable cost configuration in the inspector;
- configure representative services such as digital printing and graphic design without hardcoded service-specific backend models.

No live final quote/selling-price calculation is required in this task. M1-004 will execute these definitions through the Go pricing engine.

## Scope

### 1. Versioned persistence migration

Extend the existing SQLite migration system from M1-001/M1-002.

Add persistence for:

- machines
- machine rate definitions as needed by the chosen compact model
- ordered service cost components

Requirements:

- preserve all existing Materials and Services data;
- migration must be safe on fresh and existing databases;
- foreign keys enabled and enforced;
- service cost-component collection writes are transactional;
- machine archive/reactivation must not silently corrupt existing service references;
- no runtime schema recreation or destructive reset.

### 2. Machine domain model

Implement a focused reusable Machine model.

Minimum fields:

- ID
- name
- optional code
- category/type label
- rate basis
- rate in canonical integer Rial
- optional setup/fixed cost in integer Rial if useful to avoid service-specific hacks
- notes
- active/archived
- created/updated timestamps

Supported rate bases should remain generic and small. At minimum support concepts equivalent to:

- per unit/page/cycle
- per minute
- per hour

Do not build maintenance history, counters, depreciation, consumable inventory, or production assignment yet.

Validation belongs in Go:

- non-empty name
- supported rate basis
- non-negative canonical Rial rates/costs

### 3. Machine application operations

Add application-layer operations for at minimum:

- list machines
- get machine
- create machine
- update machine
- archive machine
- reactivate machine

Wails methods must be thin adapters over these operations.

### 4. Generic service cost-component model

Extend persisted Service configuration with an ordered list of reusable cost components.

Supported component types:

1. **material**
2. **machine**
3. **labor**
4. **outsourced**
5. **fixed**
6. **overhead**
7. **waste**
8. **manual**

Keep the model generic. Do not create backend structs such as `DigitalPrintPaperCost` or `BusinessCardMachineCost`.

Every component should have common metadata where appropriate:

- ID
- type
- label/name
- enabled state
- order/position
- optional notes

Then type-specific configuration should be explicit and typed enough for validation and the future pricing engine.

### 5. Material cost components

A material component must be able to reference a persisted Material and define how future pricing obtains the consumed quantity.

For this task, support a compact generic configuration that can represent at least:

- fixed material quantity per service item/order item;
- quantity sourced from a numeric service parameter;
- a multiplier/factor applied to that quantity;
- optional waste percentage or separate generic waste component.

Use the existing fixed-scale quantity representation. No authoritative binary floating point.

The persisted component must preserve a stable material reference by ID, not only by name.

Do not post inventory consumption yet.

### 6. Machine cost components

A machine component must reference a persisted Machine and define how machine usage is obtained in a generic way.

Support at minimum:

- fixed usage quantity;
- usage sourced from a numeric service parameter;
- multiplier/factor.

Examples the model should be able to represent later:

- 100 pages × printer per-page rate;
- 20 minutes × cutter per-minute rate;
- 1.5 hours × design workstation/internal machine hourly rate if desired.

Do not calculate final cost in this task beyond any small preview needed to make editing understandable. The authoritative pricing evaluation belongs in M1-004.

### 7. Labor cost components

Support generic labor configuration with canonical integer-Rial rates.

At minimum represent:

- rate basis: per unit, per minute, or per hour;
- rate in Rial;
- usage fixed or sourced from a numeric service parameter;
- optional multiplier.

Do not create employee/payroll domain entities yet.

### 8. Outsourced, fixed, manual components

Support simple canonical Rial definitions:

- **outsourced:** fixed amount or parameter-scaled amount;
- **fixed:** fixed Rial amount;
- **manual:** operator-entered amount intended for future runtime input/override.

Keep the distinction explicit for later cost explanation/reporting.

### 9. Overhead component

Support percentage overhead without binary floating point.

Represent percentages using the existing fixed-scale decimal approach or another explicit integer/fixed-scale representation.

The component must define a clear future basis, at minimum one of:

- percentage of accumulated prior cost;
- percentage of material cost;
- percentage of subtotal before overhead.

Choose and document a small deterministic contract. Avoid an expression language in this task.

### 10. Waste component

Support explicit waste configuration generically.

At minimum allow a percentage/factor that the future pricing engine can apply to a referenced material component or material subtotal.

Use fixed-scale decimal representation and validate non-negative values.

Do not mutate inventory or create waste ledger entries yet.

### 11. Parameter references

Where a component sources usage from a Service parameter:

- reference it by stable parameter key or ID according to the current M1-002 model;
- validate the referenced parameter exists;
- only numeric parameter types may be used for numeric usage;
- service update must reject broken component references atomically;
- renaming/changing a parameter must not leave silently invalid components.

Prefer a design that makes future service editing predictable and does not require repairing hidden JSON.

### 12. Service application and persistence integration

Extend the existing service aggregate/application flow rather than creating a separate disconnected cost-component service unless there is a strong architectural reason.

Requirements:

- service + parameters + cost components can be loaded as one coherent persisted configuration;
- updates that alter parameters/components are transactional;
- ordering round-trips exactly;
- validation happens in Go before persistence;
- archive/reactivate semantics remain consistent with M1-002.

### 13. Wails integration

Expose only the operations needed by Machines and Services UI.

Requirements:

- no SQL in Wails methods;
- no authoritative business validation in Vue;
- clear error messages suitable for UI feedback;
- typed generated bindings;
- reuse the existing application/store wiring and single database lifecycle.

### 14. Machines workspace

Add a real **Machines** workspace using the established shared UI system.

Include:

- sticky page header and `New machine` action;
- search;
- active/archived/all filter;
- dense table/list;
- right inspector/editor;
- create/edit/archive/reactivate;
- rate basis and Rial/Toman-aware rate inputs/display;
- notes/status/timestamps;
- loading/empty/error/success states.

Suggested table columns:

- Machine
- Type/category
- Rate basis
- Rate
- Setup/fixed cost if implemented
- Status

Use Lucide icons, shared workspace geometry, Jalali timestamps, and existing currency utilities.

### 15. Services cost-component editor

Extend the current Services inspector/editor.

The user must be able to:

- add a cost component;
- select its type;
- edit type-specific fields through normal controls;
- select referenced Materials/Machines where relevant;
- select eligible numeric service parameters as usage sources;
- enable/disable components;
- reorder components;
- remove components;
- see a readable summary of configured cost sources when not editing.

Do not require raw JSON editing.

Keep dense desktop ergonomics. If the editor becomes too tall, use coordinated tabs/sections inside the right inspector rather than introducing large modal flows.

### 16. Representative configurations

The model/UI should be able to persist configurations equivalent to:

#### Digital Print

Parameters from M1-002:

- quantity: integer
- paper_size: choice
- color_mode: choice
- sides: choice
- paper: material-reference

Cost components can include:

- paper material cost using quantity-derived usage/factor;
- printer machine cost using quantity-derived usage;
- fixed setup cost;
- waste percentage;
- overhead percentage.

It is acceptable if the exact sheet-yield formula is deferred to M1-004, provided the generic component structure does not block it.

#### Graphic Design

Parameter:

- estimated_hours: decimal

Cost components:

- labor hourly rate using `estimated_hours`;
- optional fixed overhead.

### 17. Money and decimal invariants

Preserve existing M1 foundations:

- authoritative money is integer Rial;
- Toman is input/presentation only;
- no binary floating point for persisted authoritative rates, quantities, multipliers, percentages, or decimal parameter metadata;
- round trips through Go/SQLite must be exact within the documented fixed scale.

### 18. Existing UI/architecture conventions

Preserve:

- single app-owned DB lifecycle;
- domain/application/infrastructure/Wails boundaries;
- Lucide icons;
- unified sticky workspace geometry;
- workspace-owned scrolling;
- Rial/Toman preference;
- Jalali presentation;
- restrained motion;
- dense table + right-inspector visual language.

Do not regress Dashboard, Orders, Materials, or Services.

## Testing

Add focused Go tests for at minimum:

- migration upgrade from an M1-002 database;
- fresh migration and reopen;
- machine create/read/update/archive/reactivate persistence;
- exact integer-Rial machine/labor/fixed-rate round trips;
- service cost-component persistence and exact ordering;
- each supported component type validation;
- material/machine foreign-reference validation;
- numeric parameter-reference validation;
- broken parameter-reference update rejection;
- transaction rollback when a component in the collection is invalid;
- exact fixed-scale multiplier/percentage round trips;
- existing M1-001/M1-002 persistence tests continue passing.

Frontend build must pass.

## Non-goals

Do **not** implement in this task:

- final Go pricing calculation/evaluation engine;
- formula/expression language;
- live authoritative selling-price calculation;
- pricing rules/markup/tier evaluation;
- order/quote item snapshots;
- purchases or supplier domain;
- inventory movement/consumption posting;
- production jobs;
- machine maintenance/counters/depreciation;
- employees/payroll;
- accounting journal entries;
- actual-vs-estimated production costing.

## Acceptance criteria

1. Existing M1-001/M1-002 databases migrate without data loss.
2. Machines persist through versioned SQLite migrations and app restart.
3. Machine rates/costs use canonical integer Rial.
4. Services persist ordered generic cost components.
5. All required component types are representable and validated in Go.
6. Material and machine references are stable persisted IDs.
7. Numeric parameter usage references are validated against the owning service.
8. Service parameter/component updates are atomic.
9. Fixed-scale percentages/multipliers round-trip without binary-float loss.
10. Machines UI provides create/edit/archive/reactivate/search/filter workflows.
11. Services UI provides usable add/edit/remove/reorder cost-component workflows without raw JSON.
12. Rial/Toman and Jalali presentation conventions are preserved.
13. Existing Dashboard, Orders, Materials, and Services workflows remain functional.
14. Relevant Go tests pass.
15. Frontend build passes.
16. `git diff --check` passes.

## Manual verification

1. Run Atropaten against the existing development database containing Materials and Services.
2. Confirm previous Materials and Services still exist after migration.
3. Create a machine such as `Production Printer` with a per-page rate.
4. Close/reopen Atropaten and verify it persists.
5. Edit an existing service or create `Digital Print`.
6. Add a material component referencing an existing paper Material.
7. Add a machine component referencing `Production Printer`.
8. Add fixed, waste, and overhead components.
9. Reorder components, save, reopen, and confirm exact ordering/configuration persists.
10. Create/edit `Graphic Design` with an `estimated_hours` decimal parameter and labor hourly component referencing that parameter.
11. Archive/reactivate the machine and confirm existing service configuration remains readable and explicit.
12. Confirm Dashboard, Orders, Materials, and other existing pages still render/navigate correctly.

## Validation

Run:

```sh
go test ./...
cd frontend
npm run build
cd ..
git diff --check
```

If Wails is available, also run the application and exercise the manual flow.

If Go is unavailable in the implementation environment, state that explicitly and do not claim Go tests passed.

## Relevant documentation

- `docs/PRODUCT.md`
- `docs/DOMAIN.md`
- `docs/ARCHITECTURE.md`
- `docs/UI.md`
- `MILESTONES.md`
- `CONTRIBUTING.md`
- `tasks/M1-001-materials-persistence-foundation.md`
- `tasks/M1-002-services-dynamic-parameters.md`
