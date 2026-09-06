# M3-002 — Inventory reservations and production execution

## Objective

Complete M3 by connecting persisted Orders to material availability and production execution without violating the immutable inventory-ledger model established in M3-001.

This task adds inventory reservations, production jobs, actual consumption, waste, outsourcing metadata, and the real Production workspace.

## User-visible outcome

The operator can take a confirmed order into production, reserve required materials, see available versus reserved stock, start and complete production jobs, record actual consumption and waste, and inspect the resulting immutable inventory movements.

## Scope

### 1. Migration

Add the next versioned SQLite migration. It must upgrade the current v7 database without losing any M1/M2/M3-001 data.

Persist at minimum:

#### Inventory reservation
- ID
- material ID
- order ID
- optional order item ID
- optional production job ID
- quantity in canonical consumption units
- status (`active`, `released`, `consumed`, `cancelled` or equivalent)
- created/updated timestamps

Reservations are operational availability records, not inventory movements.

#### Production job
- ID
- order ID
- order item ID
- job number or stable human-facing identifier if useful
- quantity / production quantity snapshot
- assigned machine ID optional
- status
- priority
- planned/start/completed timestamps
- notes
- estimated cost snapshot references as appropriate
- actual material/labor/outsourced cost totals as available

#### Production consumption
Persist enough to record:
- production job ID
- material ID
- actual consumed quantity
- actual waste quantity
- unit cost snapshot or movement linkage
- notes
- created timestamp

Prefer deriving actual inventory value from the immutable movement ledger and explicit movement links rather than duplicating mutable balances.

#### Outsourcing metadata
For an outsourced production step, support lightweight metadata:
- supplier ID optional
- external description
- quoted/actual cost Rial
- sent date
- expected return date
- received date
- notes

Do not build a separate full subcontract-purchase subsystem yet.

### 2. Reservation model

Implement the invariant:

`available_stock = physical_stock - active_reserved_stock`

Requirements:
- physical stock continues to derive only from immutable inventory movements
- reservations never create inventory value/accounting effects
- reservations cannot make available stock invalid unless an explicit over-reservation policy is implemented; default should reject over-reservation
- reservation create/update/release operations are transactional
- cancelling/reducing an order/job releases reservations
- completing consumption converts reserved operational quantity into real inventory movements atomically where applicable
- no silent stock mutation

### 3. Estimated material requirements

Use persisted order-item snapshots and service configuration/cost-component data carefully.

For this task, derive reservable material quantities only when a deterministic material quantity is available from the saved order-item configuration/snapshot.

Requirements:
- do not reprice historical order items
- do not silently rerun current service pricing to mutate historical values
- if the saved snapshot contains enough resolved material quantity, use it
- if not, allow explicit operator reservation/production quantity entry rather than inventing a service-specific formula
- no service-specific Go branches

### 4. Order → production flow

Allow production creation only from coherent order states.

At minimum:
- confirmed commercial order can create production jobs
- fulfilment state moves consistently through Pending → In Production → Ready → Delivered
- creating a production job does not alter commercial pricing snapshots
- job status is independent from order commercial/payment status

Suggested production statuses:
- Pending
- Ready
- In Progress
- Paused
- Completed
- Cancelled
- Failed

Keep transition validation explicit and small; do not introduce a generic workflow framework.

### 5. Production job operations

Implement Go application/domain operations for:
- list/get jobs
- create job from order item
- assign/change machine
- update priority/notes
- transition status
- reserve materials
- edit/release reservations before consumption
- record actual consumption
- record waste
- record outsourced cost/status metadata
- complete job
- cancel job with reservation release

All multi-record operations must be transactional.

### 6. Actual consumption and inventory movements

Recording actual production material usage must create immutable inventory movements.

Use movement types such as:
- `production_consumption`
- `waste`

Requirements:
- consumption quantity must be positive
- waste quantity must be non-negative
- movement quantities use the M3-001 six-decimal fixed scale
- unit cost/value uses the current authoritative weighted-average inventory valuation rules at posting time unless the existing M3-001 model provides a stronger consistent mechanism
- posting is atomic with reservation reduction/consumption state
- retries/idempotency must not duplicate movements
- completed/posting history cannot be hard-deleted
- corrections use compensating movements, not history rewrite

### 7. Actual versus estimated cost

Production should expose estimated versus actual cost without mutating the commercial order-item snapshot.

At minimum:
- saved order item estimated cost remains immutable
- actual material cost is derived from posted production movements
- actual outsourced cost may be recorded
- machine/labor actual cost may remain simple/manual if no robust runtime usage data exists yet
- production job can expose actual total known cost

Do not replace historical estimated cost with actual cost.

### 8. Materials integration

Extend the Materials workspace to show:
- physical stock
- reserved stock
- available stock
- average cost
- inventory value
- low-stock state based on the appropriate available/physical rule documented by implementation
- reservation details/history where useful

Selecting a material should make reservation and movement effects understandable.

### 9. Orders integration

Extend Order workspace Production tab/surface to show:
- production jobs by item
- job status
- reserved materials
- assigned machine
- planned/actual timing
- actual consumption/waste
- actual known cost
- create/open job action

Do not collapse production lifecycle into order status.

### 10. Production workspace

Replace the placeholder Production page with a real persisted queue using the shared desktop workspace system.

Include:
- sticky header
- filters for Pending / Ready / In Progress / Paused / Completed / All
- search by order/customer/service
- promised date / lateness visibility
- priority
- assigned machine
- dense queue/table
- right inspector or dedicated workspace
- start/pause/resume/complete/cancel actions as valid
- reservations section
- actual consumption/waste section
- outsourcing section when applicable
- empty/loading/error/success states
- Jalali user-facing dates
- grouped Rial/Toman for costs

UI must obey project-wide layout rules:
- use available width
- no input/select/textarea clipping
- shrinkable flex/grid children
- inspector widths constrained sensibly
- laptop and wide desktop layouts both usable
- no feature-specific fixed positioning when shared sticky surfaces solve it

### 11. Delete/archive semantics

Follow the project-wide deletion rules.

- draft/unposted production records may be hard-deleted only when safe and when no immutable movement/history depends on them
- completed/posted production history must never be destructively deleted
- cancellation/reversal preserves posted history
- reservation rows may be purged only when they are purely transient and not needed for audit/history; otherwise mark released/cancelled
- referenced Machines, Materials, Suppliers, Orders remain protected by their existing rules

### 12. Wails adapters

Expose typed thin Wails bindings.

No SQL, stock arithmetic, reservation invariants, costing, or state-machine rules in Vue/Wails adapter code.

### 13. Tests

Add focused Go tests for at least:
- v7 → next migration preserving existing data
- reservation create/update/release persistence
- available stock = physical - reserved
- over-reservation rejection
- concurrent/transactional reservation integrity where practical
- production job CRUD/state transitions
- order fulfilment updates around production
- production consumption posts immutable movements
- waste posts separate immutable movement
- reservation consumption/release atomicity
- retry/idempotency does not double-post movements
- cancel before posting releases reservations
- completed job cannot be destructively deleted
- correction uses compensating movement rather than mutation
- actual material cost derived consistently from ledger valuation
- outsourced metadata persistence
- historical order pricing snapshot remains unchanged through production

Frontend build must remain clean.

## Non-goals

Do not implement yet:
- accounting journal entries
- COGS postings to accounting
- invoices/payments
- supplier payable accounting
- payroll/employees
- machine maintenance/counters/depreciation
- sophisticated production scheduling/optimization
- barcode scanning
- multi-location transfers
- customer notifications

## Acceptance criteria

- Existing v7 database upgrades safely.
- Reservations reduce available stock without changing physical stock.
- Production jobs are persisted and linked to order items.
- Actual consumption and waste create immutable inventory movements.
- Production completion preserves historical order pricing snapshots.
- Reservation release/consumption is transactional.
- Physical/reserved/available stock is visible and coherent in Materials.
- Orders expose production progress without collapsing status axes.
- Production workspace is real and usable.
- Outsourced production metadata can be recorded.
- Posted production history follows reversal/correction semantics rather than deletion.
- Existing M1/M2/M3-001 workflows remain intact.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.

## Manual verification for later integration pass

1. Open an existing confirmed order with a material-based item.
2. Create a production job.
3. Reserve material and verify available stock decreases while physical stock does not.
4. Start the job.
5. Record actual consumption and waste.
6. Complete the job.
7. Verify physical stock drops via immutable movements and reservation is cleared/consumed.
8. Verify order pricing snapshot is unchanged.
9. Verify Production queue/status and Materials reservation/availability views.
10. Restart and confirm persistence.
