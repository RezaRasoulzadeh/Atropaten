# Architecture

## Technology baseline

- Go backend
- Wails 2 desktop/runtime bridge
- Vue 3 + TypeScript frontend
- Vite frontend tooling
- SQLite local database
- Windows as primary deployment target

## Architectural goal

Keep business logic independent of Wails, Vue, and SQLite. Wails methods are adapters into application use cases, not the domain implementation itself.

Proposed backend structure:

```text
Atropaten/
├── cmd/
├── internal/
│   ├── domain/
│   ├── application/
│   ├── accounting/
│   ├── pricing/
│   ├── inventory/
│   ├── storage/
│   └── platform/
├── frontend/
├── migrations/
├── docs/
└── assets/
```

This may evolve with implementation. Do not create packages merely to mirror this diagram when they contain no meaningful boundary.

## Layers

### Domain

Business types, invariants, calculations, and state transitions that do not depend on persistence or UI.

Examples: pricing calculations, inventory movement rules, order state transitions, accounting balancing rules, owner allocation rules.

### Application

Use cases coordinating domain operations and transaction boundaries.

Examples: create order, add order item, receive purchase, start production, record consumption, issue invoice, record payment, deposit check, close fiscal period.

### Infrastructure

SQLite repositories, filesystem attachment storage, backup/restore, OS integration, printing/export adapters, and other external concerns.

### Wails adapter

A thin API exposed to the frontend. It validates transport-level input, calls application use cases, and returns frontend DTOs/errors.

### Frontend

Vue owns presentation state, interaction, routing/workspaces, filtering, table configuration, and client-side display concerns. Authoritative financial or inventory calculations belong in Go.

## Persistence principles

The final schema will be designed after the primary UI/workflows and domain boundaries stabilize.

Required properties:

- Explicit schema migrations
- Foreign keys enabled
- Transactional multi-record operations
- Monetary values stored without floating-point rounding errors
- Immutable movement/journal history after posting
- Historical snapshots for order item pricing/costing
- Database constraints for invariants where practical
- Backup/restore tested as a product feature

The exact SQLite driver and migration library are intentionally undecided until persistence implementation begins.

## Money and quantities

Do not use binary floating point for authoritative money calculations. Choose a consistent integer minor-unit or decimal representation before persistence implementation.

Quantities need decimal support because paper can be sheets while ink, roll media, area, length, and labor may be fractional.

## Accounting boundary

Every operation with financial consequences should expose its accounting effect through one central posting mechanism. Business modules should not directly manipulate account balances.

Balances are derived from journal lines. Posted journal entries must balance debits and credits.

For M4-002, sales and COGS recognition occurs when an invoice is posted. The
sales entry uses the invoice's saved Rial total; the optional COGS entry uses
only the net actual production-consumption and waste movement costs already
linked to that order. Later catalog changes or later movements never reprice
or backfill an invoice automatically; corrections are reversing entries.

For M5, incoming checks use Accounts Receivable → Checks Receivable → Checks
in Transit → Bank as the lifecycle advances. Outgoing checks use Accounts
Payable → Checks Payable → Bank. Only the clearing step touches a bank ledger,
so uncleared instruments are never available cash; returns and rejections are
compensating journal entries. Loans post principal to Cash/Bank against Loans
Payable or Loans Receivable, while repayments split integer-Rial principal and
interest lines explicitly. Check and loan posting keys are persisted so
retries cannot create duplicate financial history.

For M5-002, owner percentages are stored as integer basis points (10,000 =
100.00%), with ownership and profit-sharing tracked independently. Owner
balances are derived from party-tagged journal lines; drawings use a dedicated
equity account and owner-paid expenses credit the owner current account.
Fiscal-period P&L is derived from posted journal lines only. Closing snapshots
the active profit-sharing basis points and allocates integer Rial using floor
division; the deterministic remainder goes to the final owner in ascending
owner-ID order. Closing creates one balanced allocation journal and immutable
allocation rows, and closed-period posting/reversal is rejected.

## Inventory boundary

Inventory quantity is derived from movements. Business modules request inventory operations; they do not directly edit a current-stock number as the source of truth.

Cached/projected balances may be added for performance only if they can be rebuilt from authoritative history.

## Pricing boundary

The pricing engine accepts a service definition, resolved parameters, current cost inputs, and relevant rules. It returns an explainable calculation result containing component costs, estimated cost, suggested price, and production quantities.

The resulting calculation is snapshotted into the order item.

## Files

Application-managed files should use a predictable data directory. Store file metadata and hashes in SQLite, with bytes on disk. Backup/restore must include both database and managed files.

## Testing priorities

Most tests should target pure Go domain/application logic without launching Wails.

High-value invariant tests include:

- Pricing and production-yield calculations
- Weighted-average inventory costing
- Reservation versus consumption behavior
- Balanced accounting postings
- Partial/multi-invoice payment allocation
- Owner capital/drawing/loan separation
- Historical snapshot stability
- Reversal/correction workflows
- Fiscal-period closing rules

Frontend tests should focus on critical interaction/state behavior rather than duplicating backend business-rule tests.
