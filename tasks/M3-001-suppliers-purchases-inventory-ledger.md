# M3-001 — Suppliers, purchases, and inventory movement ledger

## Objective

Start M3 by connecting supplier purchasing to authoritative inventory quantity and weighted-average material cost.

This task adds persisted suppliers, purchases, purchase items, immutable inventory movements, weighted-average costing, and the usable Suppliers/Purchases/Materials UI needed to exercise the flow.

Do not implement production consumption/reservations yet; those belong to the next M3 task.

## User-visible outcome

The operator can create suppliers, record a material purchase, see inventory quantity and weighted-average cost update immediately, inspect movement history, correct/remove draft purchases safely, post a purchase, and reopen everything after restart.

## Scope

### 1. Migration

Add the next versioned SQLite migration upgrading v6 safely without losing M1/M2 data.

Persist at minimum:

#### Supplier
- ID
- name
- optional code
- phone
- email
- address
- notes
- active/archived
- created/updated timestamps

#### Purchase
- ID
- human-facing unique purchase number
- supplier ID
- optional supplier invoice number
- purchase date
- status
- notes
- subtotal Rial
- discount Rial
- shipping Rial
- tax Rial
- additional costs Rial
- total Rial
- created/updated timestamps

#### Purchase item
- ID
- purchase ID
- position
- material ID
- immutable material name/unit snapshot for historical readability
- purchase quantity in the material's purchase unit
- conversion factor snapshot
- resulting consumption-unit quantity
- unit acquisition cost Rial
- allocated additional-cost Rial where applicable
- resulting landed/unit inventory cost Rial
- line total Rial
- notes

#### Inventory movement
Immutable ledger rows including at least:
- ID
- material ID
- timestamp
- type
- quantity delta in canonical consumption unit/fixed-scale quantity
- unit cost Rial
- total cost Rial
- reference type
- reference ID
- note
- created timestamp

Initial movement types required in this task:
- opening_balance
- purchase
- adjustment
- supplier_return

Keep future types compatible with production consumption, waste, customer return, and transfer.

### 2. Inventory invariant

Inventory is derived from movement history.

Do not make a mutable `current_stock` field the authoritative source of truth.

The material read model may expose cached/derived:
- physical stock
- average cost
- inventory value
- low-stock state

but these must reconcile to authoritative movements.

### 3. Weighted-average costing

Use weighted-average cost as the default valuation method.

On posted inventory-increasing purchases:

`new_average_cost = (existing_inventory_value + incoming_inventory_value) / (existing_quantity + incoming_quantity)`

Requirements:
- all authoritative money is integer Rial
- quantities use the existing fixed-scale exact representation
- deterministic rounding in Go
- no binary floating point
- zero-stock edge cases tested
- supplier returns reduce quantity/value coherently and must not silently create impossible negative stock

Document and test the precise rounding rule.

### 4. Purchase units and conversion

Materials already distinguish/are expected to distinguish purchase and consumption concepts.

This task must support purchase quantity conversion into canonical consumption quantity using an exact fixed-scale conversion factor.

Examples:
- paper purchased by pack, consumed by sheet
- lamination purchased by roll, consumed by meter

If M1 material schema lacks fields required for clean purchase/consumption unit handling, extend it via migration without breaking existing materials.

Do not use binary floats.

### 5. Purchase lifecycle

Use a small explicit lifecycle:
- Draft
- Posted
- Cancelled

Draft:
- editable
- items can be added/removed/reordered
- can be hard-deleted because it has not affected inventory history

Posted:
- creates immutable purchase inventory movements atomically
- purchase/item content becomes historically protected
- must not be hard-deleted
- corrections use explicit reversal/cancellation behavior rather than rewriting/removing posted movements

Cancelled posted purchase:
- create compensating inventory movements atomically
- preserve original purchase and original movements
- do not delete history

Do not implement accounting/AP effects yet.

### 6. Deletion semantics

Follow the global project deletion rule.

Suppliers:
- provide Archive and Delete as separate actions
- hard-delete a supplier only when it has no protected historical references
- if purchases/history reference it, reject deletion clearly and suggest archive instead
- deletion must actually purge the database row, not mark archived

Draft purchases:
- hard-delete purchase and owned draft items

Posted/cancelled purchases and inventory movements:
- never hard-delete through normal UI
- use cancellation/compensating history

Materials:
- do not add unsafe material deletion that would break service/order/purchase/inventory history
- if a Delete action exists, reject it when protected references exist

### 7. Additional acquisition costs

Support purchase-level:
- shipping
- tax
- additional costs
- discount

Allocate acquisition costs that belong in inventory valuation across purchase items using a simple deterministic rule.

Prefer proportional allocation by line acquisition value.

Define clearly whether tax is included in landed inventory value; keep the first implementation simple and documented.

Rounding allocation must be deterministic and preserve the exact purchase-level allocated total by assigning any remainder predictably.

### 8. Application/domain operations

Suppliers:
- list/get/create/update/archive/reactivate/delete

Purchases:
- list/get/create/update draft metadata
- add/update/remove/reorder draft items
- recalculate draft totals in Go
- post atomically
- cancel posted purchase atomically with compensating movements
- hard-delete draft purchase

Inventory:
- list material movements
- create manual stock adjustment with explicit reason/note
- read material inventory summary

Keep business rules in Go, Wails thin.

### 9. Materials integration

Extend the real Materials workspace so each material shows authoritative inventory information:
- physical stock
- average cost
- inventory value
- low-stock state
- purchase unit / consumption unit where applicable

Inspector should show recent inventory movements and a clear Adjust Stock action.

An adjustment must create a movement; it must not directly overwrite stock.

Use grouped Rial/Toman money formatting and exact quantity formatting.

### 10. Suppliers workspace

Replace the placeholder Suppliers page with a real persisted workspace.

Include:
- shared sticky header + New supplier
- search
- Active / Archived / All
- dense register/table
- right inspector/editor
- create/edit/archive/reactivate/delete
- deletion confirmation
- clear error when deletion is blocked by history
- loading/empty/error/success states
- Jalali timestamps where shown

Follow the global layout rules:
- use available workspace width
- grid/flex children must be shrinkable
- no clipped/overflowing inputs
- inspector width must remain usable on laptop and wide desktop windows
- no page-specific scroll hacks

### 11. Purchases workspace

Replace the placeholder Purchases page with a real persisted workflow.

Include:
- sticky header + New purchase
- search/filter by supplier/status/date where practical
- dense purchase register
- supplier
- purchase/invoice number
- purchase date
- total
- status
- right inspector or dedicated purchase workspace

Draft editor:
- supplier selection
- purchase date
- supplier invoice number
- notes
- material line items
- quantity
- purchase unit
- conversion preview
- unit acquisition cost
- shipping/tax/additional costs/discount
- totals
- add/edit/remove/reorder lines
- Post Purchase action
- Delete Draft action

Posted view:
- read-only historical lines
- inventory effects summary
- Cancel/Reversal action with confirmation
- no destructive delete

Use Jalali dates in the UI and canonical date/time persistence.

### 12. Inventory movements UI

In Materials inspector or a focused subview show recent movements with:
- date
- type
- quantity delta
- unit cost
- total cost
- reference
- note

Keep this compact and reusable; do not build full warehouse reporting yet.

### 13. Referential/historical safety

- Archiving/deleting catalog/supplier entities must not make historical purchases unreadable.
- Posted purchase items retain enough snapshots to render history even if mutable supplier/material metadata later changes.
- Inventory movement rows are immutable.
- Reversal/cancellation creates compensating movement rows.

### 14. Tests

Add focused Go tests for at least:
- v6 → next migration preserving M1/M2 data
- supplier CRUD/archive/reactivate/delete
- supplier delete blocked by purchase history
- draft purchase CRUD and hard delete
- purchase item deterministic ordering
- exact quantity conversion
- exact integer-Rial totals
- deterministic acquisition-cost allocation with rounding remainder
- weighted-average costing across multiple purchases
- zero-stock weighted-average edge case
- posting atomically creates movement rows and updates derived inventory summary
- failed posting rolls back purchase/movement effects
- posted purchase cannot be destructively deleted
- cancellation creates compensating movements and preserves original rows
- supplier return behavior and negative-stock rejection
- stock adjustment creates movement rather than direct stock mutation
- inventory reconstruction from movements
- reopen persistence

Frontend build must remain clean.

## Non-goals

Do not implement yet:
- inventory reservations
- production jobs
- production consumption/waste
- outsourcing production
- transfers/multiple locations
- supplier payments/accounts payable
- invoices/accounting journal entries
- checks
- purchase-order approval workflow
- lot/batch tracking
- barcode scanning
- advanced landed-cost allocation methods

## Acceptance criteria

- Existing v6 database upgrades without data loss.
- Suppliers are fully persisted and usable.
- Safe hard-delete exists where allowed; archive is not used as fake deletion.
- Draft purchases can be created, edited, deleted, and posted.
- Posted purchases generate immutable inventory movements atomically.
- Cancelling a posted purchase preserves history through compensating movements.
- Material stock is derived from movements.
- Weighted-average cost is calculated deterministically in Go.
- Purchase-unit to consumption-unit conversion is exact/fixed-scale.
- Materials UI displays real stock/average cost/movement history.
- Purchases and Suppliers use the shared responsive desktop workspace without clipping/overflow.
- Grouped Rial/Toman formatting remains consistent.
- Existing Customers/Quotes/Orders/Services/Machines remain intact.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.

## Manual verification for later integration pass

1. Confirm existing M1/M2 data remains after migration.
2. Create a supplier.
3. Record a draft purchase with two materials.
4. Delete a draft line and confirm it is actually removed.
5. Post the purchase.
6. Open Materials and verify stock/average cost/movements.
7. Record a second purchase at a different cost and verify weighted average changes.
8. Restart and verify persistence.
9. Attempt to delete the referenced supplier and verify deletion is rejected clearly.
10. Cancel the posted purchase and verify compensating movements/history remain.
