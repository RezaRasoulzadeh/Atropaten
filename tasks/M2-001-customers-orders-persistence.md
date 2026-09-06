# M2-001 — Customers and persisted orders

## Objective

Replace the remaining mock order/customer workflow with the first persisted commercial vertical slice: customers, orders, mixed order items, dynamic service configuration snapshots, pricing snapshots, and a usable Orders workspace.

Build on M1 without coupling historical orders to mutable catalog definitions.

## User-visible outcome

The operator can create a customer, create an order with multiple unrelated service items, configure each item using the existing generic service configurator/pricing engine, save it, reopen it after restart, and see the exact commercial/pricing configuration that was accepted at order time.

## Scope

### 1. Migration

Add the next versioned SQLite migration. It must upgrade an existing M1 database without losing Materials, Services, Machines, pricing rules, or cost components.

Persist at minimum:

#### Customer
- ID
- name
- optional phone
- optional email
- optional address
- optional notes
- active/archived
- created/updated timestamps

Keep this model compatible with a future generalized Party/customer/supplier model, but do not build the supplier domain now.

#### Order
- ID
- stable human-facing order number
- optional customer ID
- created date/time
- optional promised date/time
- priority
- commercial status
- fulfillment status
- payment status placeholder/derived initial state
- notes
- subtotal Rial
- discount Rial
- selling total Rial
- estimated cost Rial
- created/updated timestamps

#### Order item
- ID
- order ID
- position
- service ID reference when still available
- immutable service name/code snapshot
- quantity where applicable
- resolved parameter snapshot
- immutable cost/pricing breakdown snapshot
- estimated cost Rial
- suggested price Rial
- selling price Rial
- notes

Use normalized relational storage where it materially benefits querying/integrity, but immutable structured snapshots may use JSON if clearly bounded, validated at creation, and treated as historical values rather than an editable domain store.

### 2. Historical snapshot invariant

An accepted/saved order item must not change when later:
- a service is renamed
- parameters change
- materials or machine rates change
- cost components change
- pricing rules change
- the service is archived

Store enough resolved information to render the historical item and explain its price later without rerunning current catalog pricing.

Do not silently recalculate persisted historical items on read.

### 3. Order number generation

Implement a simple deterministic local order-number strategy suitable for one shop.

Requirements:
- unique under concurrent/transactional creation
- human readable
- no dependence on frontend-generated IDs
- do not build a generalized document numbering framework yet

### 4. Application/domain operations

Implement clear Go operations for:

Customers:
- list/get/create/update/archive/reactivate
- search/filter support as appropriate

Orders:
- list/get/create/update draft order metadata
- add configured item
- edit/replace a draft item by recalculating from current catalog inputs and then replacing its snapshot
- remove/reorder draft items
- apply order-level discount
- update commercial/fulfillment state with basic valid transitions
- persist atomically

Use the M1 pricing service as the authoritative calculator when creating/reconfiguring an item.

Do not duplicate pricing formulas in Wails or Vue.

### 5. Money and totals

All authoritative money remains integer Rial.

Order totals must be calculated in Go:
- subtotal = sum item selling prices
- discount must be non-negative and cannot produce an invalid negative total
- selling total = subtotal - discount
- estimated cost = sum item estimated costs

Use the shared frontend grouped Rial/Toman formatting/input behavior from M1-004 everywhere.

### 6. Customer validation

At minimum:
- name required
- trim text consistently
- phone/email may remain lightweight strings; do not add brittle country-specific validation yet
- archived customers remain readable from historical orders

### 7. Order state axes

Keep separate axes rather than one overloaded status.

Commercial:
- Draft
- Confirmed
- Closed
- Cancelled

Fulfillment:
- Pending
- In Production
- Ready
- Delivered

Payment for this task:
- Unpaid
- Partially Paid
- Paid

M2-001 does not implement payments yet. New orders start Unpaid; do not expose arbitrary payment-state editing that pretends accounting exists.

Enforce only transitions necessary for a coherent first workflow. Avoid an overengineered state-machine framework.

### 8. Wails adapters

Expose typed thin bindings for customers and orders.

No SQL, pricing formulas, total calculation, snapshot construction, or authoritative state validation in Wails handlers.

### 9. Customers workspace

Replace the placeholder Customers page with a real persisted vertical slice using the shared workspace geometry.

Include:
- sticky page header + New customer
- search
- active/archived/all filter
- dense customer list/table
- selection
- right inspector/editor
- create/edit/archive/reactivate
- empty/loading/error/success states
- Jalali user-facing timestamps where appropriate

Do not build full customer financial ledgers yet.

### 10. Orders workspace

Replace the current mock Orders data/workflow with persisted backend data while preserving/refining the existing M0 Orders UX.

Include:
- sticky header + New order
- search/filter
- dense order register
- customer selection
- promised date using shared Jalali date UI
- priority
- notes
- separate commercial/fulfillment/payment status display
- order workspace with item list and totals
- create/save/reopen

Remove authoritative mock order data once persisted data owns this workflow.

### 11. Add-item service configurator

Reuse the generic M1 service configuration/pricing behavior when adding an order item.

Flow:
1. choose active service
2. render its dynamic parameters
3. calculate via Go pricing engine
4. show breakdown/suggested price/profit/below-cost warning
5. allow selling-price override
6. add item by persisting the resolved immutable snapshot

Support multiple unrelated services in one order.

Do not create service-specific Vue or Go branches.

### 12. Draft editing and catalog changes

When editing an existing draft item:
- show its saved snapshot first
- explicit reconfiguration may use the current service definition and pricing engine
- saving reconfiguration replaces that item's snapshot atomically
- merely opening the order must never mutate/reprice it

### 13. Tests

Add focused Go tests for at least:
- M1 v4 → M2 migration preserving catalog data
- customer create/read/update/archive/reactivate persistence
- unique order-number generation
- mixed order item persistence and deterministic ordering
- exact Rial totals/discount roundtrip
- immutable item snapshot after service/pricing/catalog changes
- order reopen without repricing
- item add/update/remove/reorder transaction behavior
- failed item pricing/snapshot write rolls back
- basic commercial/fulfillment transition validation
- archived customer/service historical readability

Frontend build must remain clean.

## Non-goals

Do not implement yet:
- quotes
- invoices
- payments/payment allocations
- customer financial ledger
- suppliers/purchases
- inventory reservations/movements
- production jobs/consumption/waste
- accounting journal entries
- attachments/artwork
- proof approval workflow
- printing/PDF documents
- multi-location

## Acceptance criteria

- Existing M1 database upgrades without data loss.
- Customers are persisted and fully usable from the desktop UI.
- Orders are persisted; mock authoritative order data is removed.
- One order can contain multiple unrelated service items.
- Item configuration and pricing use the Go M1 engine.
- Saved items contain immutable resolved parameter/cost/price snapshots.
- Catalog changes do not alter historical saved items.
- Order totals are authoritative Go integer-Rial values.
- Grouped Rial/Toman formatting remains consistent on labels and money inputs.
- Separate commercial/fulfillment/payment statuses are represented.
- Orders survive restart and reopen without repricing.
- Existing Materials, Services, Machines, and pricing workflows remain intact.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.

## Manual verification for later integration pass

1. Confirm existing Materials/Services/Machines remain.
2. Create a customer.
3. Create an order for that customer.
4. Add two unrelated configured services.
5. Override one selling price.
6. Apply an order discount.
7. Save and reopen the order.
8. Change a service pricing rule.
9. Reopen the existing order and confirm its saved values are unchanged.
10. Restart the app and confirm persistence.
