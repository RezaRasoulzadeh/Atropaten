# M1-004 — Pricing engine + backend-connected service configurator

## Objective

Implement Atropaten's generic Go pricing engine and connect the existing Services catalog to a real backend-powered service configurator. Pricing must evaluate persisted service parameters and reusable cost components without service-specific backend code.

This task completes the core M1 catalog/pricing capability. Keep the model deterministic, explainable, and compatible with later quote/order snapshots.

## User-visible outcome

The user can select a persisted service, enter parameter values, and receive a live backend-calculated pricing result showing:

- resolved inputs
- material/machine/labor/outsourced/fixed/waste/overhead/manual cost breakdown
- estimated total cost
- suggested selling price
- operator-editable selling price
- profit and margin
- below-cost warning

The calculation must come from Go, not duplicated in Vue.

## Scope

### 1. Generic pricing request

Add a small application/domain request model that accepts:

- service ID
- parameter values keyed by persisted parameter key
- optional/manual cost values only where the existing component model explicitly allows them
- optional selling-price override for comparison/output, not as authoritative service configuration

Validate all inputs in Go against the selected persisted service definition.

Do not create backend types dedicated to specific services such as business cards or digital print.

### 2. Parameter resolution

Resolve and validate all existing M1-002 parameter types:

- integer
- decimal
- boolean
- choice
- material-reference

Rules:

- required parameters must be present
- defaults may be applied where defined
- integer/decimal bounds must be enforced
- choice values must belong to configured options
- material references must resolve to valid active materials where applicable
- fixed-scale decimal representation must remain exact; do not introduce binary floating point

Return useful structured validation errors suitable for the configurator UI.

### 3. Cost-component evaluation

Evaluate the generic M1-003 component types:

- material
- machine
- labor
- outsourced
- fixed
- overhead
- waste
- manual

Use persisted component ordering for the explanation output.

Each component result should expose enough information to explain the calculation, for example:

- component ID/name/type
- resolved usage quantity/basis
- resolved rate or source
- calculated Rial amount
- short human-readable explanation

Rules:

- authoritative money remains integer Rial
- authoritative decimal quantities/multipliers/percentages remain fixed-scale/non-lossy
- no floating-point arithmetic for money or authoritative decimal calculations
- define deterministic integer rounding behavior where multiplication/division produces fractional Rial; document and test it
- disabled components contribute zero and should either be omitted from the active breakdown or explicitly marked disabled consistently

### 4. Material component semantics

For a material component:

- resolve the referenced material
- derive usage from the component's configured fixed usage or numeric service parameter source and multiplier
- use the material's canonical average cost basis consistently
- convert units only through the existing material conversion contract
- do not mutate inventory

This is estimation only. Do not add reservations, movements, or production consumption.

### 5. Machine component semantics

For a machine component:

- resolve the referenced machine
- derive usage from fixed usage or numeric service parameter source and multiplier
- apply the persisted machine rate according to its rate basis
- include setup cost only if the existing component/configuration semantics explicitly opt into it; do not silently charge setup costs merely because a machine has one

If the M1-003 model does not yet encode explicit setup-cost inclusion, add the smallest generic field necessary rather than hardcoding behavior.

### 6. Labor / outsourced / fixed / manual components

Support generic fixed or usage-based calculation where the current component fields allow it.

Keep these component types generic. Do not introduce employee/payroll, supplier purchasing, or accounting behavior.

### 7. Waste and overhead

Implement percentage-based waste/overhead deterministically.

Define the calculation base explicitly in code and UI. Prefer simple generic semantics that are composable and non-recursive.

Recommended contract:

- waste percentage applies to eligible direct pre-waste cost accumulated before the waste component
- overhead percentage applies to eligible accumulated cost before the overhead component

If component ordering affects the base, make that behavior explicit and deterministic. Avoid recursive/self-referential percentage calculations.

### 8. Suggested selling price

Add a minimal generic pricing-rule mechanism sufficient for M1 acceptance.

Support at least:

- fixed selling price
- cost + markup percentage
- cost + fixed margin
- per-unit selling rate driven by a numeric parameter
- quantity tiers
- manual selling price

Persist rules as service configuration through a versioned migration if persistence changes are required.

Do not implement arbitrary executable expressions or a general scripting/formula language in this task.

Rules must remain generic and service-independent.

### 9. Pricing result

Return a deterministic backend result containing at minimum:

- service identity
- resolved parameter values
- ordered component breakdown
- estimated cost in Rial
- suggested selling price in Rial
- effective selling price in Rial if an operator override was supplied
- profit in Rial
- margin percentage using the project's fixed-scale decimal representation
- below-cost boolean
- warnings

Do not persist quote/order snapshots yet. M2 will own that.

### 10. Wails integration

Expose a thin pricing calculation operation to the frontend.

Requirements:

- no pricing formulas duplicated in Wails adapter methods
- no SQL in Wails methods
- no authoritative pricing/business rules in Vue
- typed request/result bindings

### 11. Backend-connected service configurator

Implement/replace the existing mock service configurator with a real vertical slice.

The UI must:

- select an active persisted service
- render controls dynamically from its parameter definitions
- support all current parameter types
- load material choices for material-reference parameters
- show validation errors near the relevant controls where practical
- invoke the Go pricing engine when inputs are complete/changed
- display ordered cost breakdown
- display estimated cost
- display suggested price
- allow operator selling-price override
- display profit and margin
- clearly warn when the effective selling price is below estimated cost
- allow resetting override back to suggested price
- follow existing shared sticky workspace/layout patterns
- use Lucide icons, Jalali conventions where dates appear, and existing restrained motion

Do not require raw JSON editing.

### 12. Representative configurations

The generic model and pricing engine must be able to represent and calculate at least these examples without service-specific backend code.

#### A. Digital / per-page print

Representative parameters:

- quantity: integer
- paper: material-reference
- paper_size: choice
- color_mode: choice
- sides: choice

Representative components may include:

- paper/material cost
- machine/page cost
- fixed setup cost
- waste percentage
- overhead percentage

#### B. Business cards

Use persisted parameters/components/rules to demonstrate a quantity-driven configuration. Exact imposition/yield geometry is not required in this task unless it already exists generically.

#### C. Graphic design / time-only work

Representative parameter:

- estimated_hours: decimal

Representative component:

- hourly labor rate × estimated_hours

This must work without any material or machine requirement.

### 13. Thousand separators — mandatory mini-task

Iranian Rial/Toman values are frequently large and must remain readable everywhere.

Implement a shared, project-wide thousands-grouping behavior for monetary UI.

Requirements:

- all read-only monetary labels/text should display thousands separators, e.g. `125,000,000`
- monetary input fields should also display/group thousands while editing
- grouping must work in both Rial and Toman presentation modes
- parsing grouped inputs must still resolve exactly to canonical integer Rial
- switching Rial/Toman must preserve the exact underlying Rial value
- negative values, if supported by an existing field, must format/parse correctly
- empty/partial editing states must remain usable
- do not use locale behavior that changes canonical persistence or introduces decimal ambiguity
- centralize behavior in shared currency utilities/components/directives rather than page-specific hacks
- migrate existing monetary inputs/read-only money displays in Dashboard, Materials, Machines, Services/configurator, Orders mock surfaces, and any other current screens using the shared money formatter/input helpers
- add focused frontend utility tests if a frontend test harness already exists cheaply; otherwise make the utility deterministic and cover relevant parsing/formatting behavior through the existing build plus targeted code-level checks where practical

The user-visible format may use ASCII comma grouping for now. Do not change canonical money storage.

### 14. Preserve architecture boundaries

Build on the existing M1 storage/domain/application boundaries.

Do not introduce:

- inventory movements
- reservations
- production consumption
- purchases/suppliers
- customers/quotes/orders persistence
- accounting journals
- arbitrary pricing scripts

## Migration expectations

If pricing-rule persistence or additional component semantics require schema changes:

- add the next versioned migration
- preserve all M1-001/002/003 data
- verify upgrade/reopen behavior
- keep migrations forward-only and deterministic

## Testing

Add focused deterministic Go tests for at least:

- parameter validation/default resolution
- integer and decimal parameter bounds
- choice validation
- material-reference validation
- material cost evaluation
- machine cost evaluation
- labor/time-only evaluation
- fixed and outsourced/manual cost evaluation where supported
- waste calculation
- overhead calculation
- component ordering/explanation
- exact integer-Rial arithmetic and documented rounding
- fixed-scale multiplier/percentage behavior
- pricing rules: fixed, markup %, fixed margin, per-unit, quantity tier, manual
- below-cost detection
- profit and margin
- representative Digital Print calculation
- representative Graphic Design calculation
- migration upgrade/reopen if schema changes

Prefer table-driven tests for pricing cases.

## Acceptance criteria

1. Pricing calculations are performed authoritatively in Go.
2. The engine is generic and contains no service-specific backend branches.
3. All supported persisted parameter types validate and resolve correctly.
4. Persisted cost components produce an ordered, explainable cost breakdown.
5. Money arithmetic remains integer Rial with deterministic tested rounding.
6. Decimal usage/multiplier/percentage math remains fixed-scale/non-lossy.
7. Material, machine, labor/time, fixed, waste, and overhead representative cases calculate correctly.
8. Suggested selling-price rules support fixed, markup %, fixed margin, per-unit, quantity tier, and manual modes.
9. Pricing result exposes cost, suggested price, effective price, profit, margin, warnings, and below-cost state.
10. Service configurator renders dynamically from persisted service definitions and calls the backend engine.
11. Operator can override selling price and immediately see profit/margin/below-cost impact.
12. Digital Print, Business Card, and Graphic Design examples can be configured without service-specific Go code.
13. Thousands separators are shown consistently on monetary read-only values and monetary inputs across the current app.
14. Grouped Rial/Toman input still round-trips to exact canonical Rial values.
15. Existing Materials, Services, Machines, Dashboard, and Orders surfaces remain functional.
16. Relevant Go tests pass.
17. Frontend build passes.
18. `git diff --check` passes.

## Manual verification

A full manual review may be deferred until milestone review, but implementation should remain runnable.

At minimum when Wails is available:

1. Open the service configurator.
2. Configure a persisted print service and observe backend cost breakdown.
3. Change quantity and confirm recalculation.
4. Override selling price below cost and confirm warning.
5. Configure a time-only design service and confirm it calculates without materials/machines.
6. Confirm large money values display with grouping in labels and editable inputs.
7. Switch Rial/Toman and confirm exact values remain stable.
8. Reopen Materials, Services, Machines, Dashboard, and Orders to check regressions.

## Validation

Run:

```sh
go test ./...
cd frontend
npm run build
cd ..
git diff --check
```

If Wails is available, smoke-test the configurator. If it is unavailable, state that explicitly rather than claiming manual verification.

## Relevant documentation

- `docs/PRODUCT.md`
- `docs/DOMAIN.md`
- `docs/ARCHITECTURE.md`
- `docs/UI.md`
- `MILESTONES.md`
- `CONTRIBUTING.md`
- `tasks/M1-001-materials-persistence-foundation.md`
- `tasks/M1-002-services-dynamic-parameters.md`
- `tasks/M1-003-machines-cost-components.md`
