# M2-002 — Quotes, attachments, and proof metadata

## Objective

Complete M2 by adding persisted quotes, quote-to-order conversion, and lightweight attachment/proof metadata while preserving the same immutable commercial snapshot rules established by M2-001.

This task should finish the Customers / Quotes / Orders milestone without entering invoicing, accounting, inventory, or production execution.

## User-visible outcome

The operator can create a quote for a customer, configure multiple unrelated service items using the existing Go pricing engine, save and reopen it without repricing, convert it into a real order, and record artwork/proof metadata for quotes/orders.

## Scope

### 1. Migration

Add the next versioned SQLite migration. It must upgrade an existing v5 database without losing M1 or M2-001 data.

Persist at minimum:

#### Quote
- ID
- unique human-facing quote number
- optional customer ID
- created date/time
- optional expiry date
- status
- notes
- subtotal Rial
- discount Rial
- selling total Rial
- estimated cost Rial
- created/updated timestamps

#### Quote item
- ID
- quote ID
- position
- source service ID when still available
- immutable service name/code snapshot
- immutable resolved parameter snapshot
- immutable pricing/cost breakdown snapshot
- estimated cost Rial
- suggested price Rial
- selling price Rial
- notes

#### Attachment metadata
Support association with quote or order.

Persist:
- ID
- owner type (`quote` or `order`)
- owner ID
- logical file name
- relative/managed file path or external path representation
- optional MIME type
- optional size
- optional hash/checksum if practical
- category/type (artwork, proof, reference, other)
- notes
- created timestamp

Do not store attachment binary data in SQLite.

#### Proof metadata
Keep proof approval lightweight but explicit.

Persist enough for:
- order/quote association
- optional related attachment ID
- status
- version/label
- prepared timestamp
- approved/rejected timestamp
- optional approver/customer note
- internal note

Suggested proof states:
- Draft
- Ready
- Waiting Customer Approval
- Approved
- Rejected

No workflow engine is required.

### 2. Quote snapshot invariant

Quotes use the same historical-snapshot rules as M2-001 orders.

After save, opening a quote must not silently recalculate when current Services, Materials, Machines, cost components, pricing rules, or material costs change.

Reconfiguration must be an explicit user action that replaces the quote item's snapshot atomically.

### 3. Quote numbering

Use the existing document-numbering approach from M2-001 where possible, with a separate quote prefix/sequence if appropriate.

Requirements:
- human readable
- unique
- transactionally safe
- generated in Go/storage layer, not frontend
- do not create an over-general document numbering framework

### 4. Quote operations

Implement Go application/domain operations for:
- list/get/create/update quote metadata
- add/edit/remove/reorder quote items
- order-level discount
- status changes
- convert quote to order atomically

Suggested quote statuses:
- Draft
- Sent
- Accepted
- Rejected
- Expired
- Converted

Keep transition rules simple and coherent.

### 5. Quote-to-order conversion

Conversion is important and must be historically safe.

Requirements:
- conversion copies the quote's saved item snapshots exactly; it must not rerun current pricing
- order totals must match the accepted quote at conversion time
- preserve customer, notes, discount, and relevant item notes
- create a new real order using the existing M2-001 order model
- record quote → order linkage
- mark quote Converted only after the order is created successfully
- operation must be atomic
- repeat conversion must not accidentally create duplicate orders

Do not invent production/payment/accounting effects.

### 6. Pricing integration

For creating or explicitly reconfiguring a quote item:
- select an active service
- render generic dynamic parameters
- use the existing Go pricing engine
- support price override
- show cost breakdown/profit/below-cost warning
- persist immutable snapshot

No duplicated Vue pricing formulas and no service-specific branches.

### 7. Totals and currency

All authoritative money remains integer Rial.

Quote totals are Go-owned and follow the same rules as orders.

Use the shared grouped Rial/Toman formatting and input behavior everywhere.

### 8. Quotes workspace

Add a real Quotes workspace to the main navigation using the unified desktop workspace system.

Include:
- sticky header + New quote
- search/filter
- dense quote register
- customer
- quote number
- created/expiry date
- status
- total
- right inspector or full quote workspace
- create/save/reopen
- archive is not necessary if statuses adequately preserve history
- loading/empty/error states
- Jalali display/input for user-facing dates

Reuse existing Orders/customer UI patterns rather than creating a parallel design language.

### 9. Quote editor/workspace

Support:
- customer selection
- expiry date
- notes
- item list
- add/edit/remove/reorder items
- live item configurator via backend
- subtotal/discount/total/cost/profit summary
- status controls
- Convert to Order action when valid

After conversion, surface the linked order and prevent accidental duplicate conversion.

### 10. Attachments metadata UI

Add a lightweight Files/Attachments surface to Quote and Order workspaces.

For this milestone, it is acceptable to manage metadata/path references without implementing a sophisticated managed-file store.

Support:
- add metadata/path reference
- category
- display name
- notes
- remove metadata entry from draft/non-protected contexts
- open/reveal action only if existing Wails capabilities make it clean; otherwise omit

If implementing file picking is straightforward with current Wails, use a thin adapter. Do not block the task on platform-specific file-management complexity.

Do not put filesystem business rules in Vue.

### 11. Proof metadata UI

Add a compact Proofs surface for Quotes/Orders where appropriate.

Support:
- create proof record/version
- associate attachment optionally
- mark Ready / Waiting Approval / Approved / Rejected
- record approval/rejection note and timestamp
- preserve prior proof versions rather than overwriting history

This is metadata only. Do not implement customer portal, email, notifications, or online approval.

### 12. Customer integration

Quotes and orders should both show under the selected customer context where useful.

Do not implement customer financial ledger yet.

### 13. Historical safety

Archiving or changing a customer/service must not make historical quotes/orders unreadable.

Quote-to-order conversion must operate from stored quote snapshots, not current catalog definitions.

### 14. Tests

Add focused Go tests for at least:
- v5 → next migration preserving M1/M2-001 data
- quote CRUD and persistence
- deterministic quote item ordering
- exact Rial totals and discount
- immutable quote snapshot after catalog/pricing changes
- explicit reconfiguration replacing only the selected draft quote item snapshot
- unique quote numbering
- quote status validation
- atomic quote-to-order conversion
- conversion copies stored snapshots without repricing
- repeated conversion does not duplicate an order
- rollback if order creation fails during conversion
- attachment metadata persistence
- proof version/history persistence
- proof status/timestamp validation where applicable
- historical readability after service/customer archive

Frontend build must remain clean.

## Non-goals

Do not implement yet:
- invoices
- receipts
- payments
- journal entries/accounting
- supplier/purchase flows
- inventory reservation or movements
- production jobs
- actual consumption/waste
- customer portal
- email/SMS
- cloud storage
- full document rendering/PDF printing
- image preview/editor
- online proof approval

## Acceptance criteria

- Existing v5 database upgrades without data loss.
- Quotes are persisted and usable from a real Quotes workspace.
- Quotes support multiple unrelated service items.
- Quote items use the existing Go pricing engine only when initially configured or explicitly reconfigured.
- Reopening saved quotes does not reprice them.
- Quote-to-order conversion copies immutable saved snapshots without using current catalog pricing.
- Conversion is atomic and idempotent enough to prevent accidental duplicate orders.
- Attachment metadata is persisted for quotes/orders without SQLite BLOB storage.
- Proof versions and approval metadata are persisted historically.
- Jalali dates and grouped Rial/Toman behavior remain consistent.
- Existing Customers, Orders, Materials, Services, Machines, and pricing workflows remain intact.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.

## Manual verification for later integration pass

1. Create a customer and quote.
2. Add two unrelated service items.
3. Save/reopen the quote.
4. Change a catalog pricing rule and confirm quote stays unchanged.
5. Add artwork/proof metadata.
6. Record a proof as Waiting Customer Approval then Approved.
7. Convert quote to order.
8. Verify the new order exactly matches the saved quote prices/configuration.
9. Try conversion again and verify no duplicate order is created.
10. Restart and confirm persistence.
