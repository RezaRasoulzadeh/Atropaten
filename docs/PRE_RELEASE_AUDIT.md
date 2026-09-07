# Pre-release integration audit

Date: 2026-09-07

This audit covers the persisted M0–M6 product as implemented on schema v13. It is a repair record, not a new feature specification. The review followed the task priority order and exercised the existing SQLite/application boundaries with focused cross-domain tests.

## Audited areas

- Customer → quote → order → production → invoice → payment and journal flow, including immutable order/invoice snapshots and retry behavior.
- Supplier → purchase → movement ledger → production consumption/waste → historical-cost COGS, including supplier payments and purchase cancellation.
- Partial and multiple allocations, customer deposits/credit, expenses, transfers, checks, loans/installments, owner transactions, fiscal periods, and report queries.
- Persistence and migrations from representative v8, v10, v11, v12, and v13 databases, plus rejection of a future schema version.
- Draft deletion, protected posted/history records, archive versus delete APIs, compensating reversals, proof/attachment history, and managed-file backup/restore failure paths.
- Shared workspace/sticky surfaces and the current Wails-facing API/binding boundary by source inspection and frontend compilation.

## Invariants verified

- Posted journal entries balance exactly in integer Rial; posted lines are immutable and corrections append reversals.
- Inventory physical stock and value are reconstructed from immutable movements. Reservations affect availability only; consumption, waste, adjustments, purchase cancellation, and corrections append movements.
- Fixed-scale quantity conversion and weighted-average calculations do not use floating point and now reject overflowing intermediate or final Rial/quantity values.
- Historical quote/order/invoice/purchase/production snapshots are read from persisted records rather than current catalog values.
- Payment, purchase, invoice, expense, transfer, check, loan, owner, and fiscal-close operations are retry-safe in their existing idempotency boundaries.
- Uncleared checks do not become cleared bank cash; owner drawings remain equity/current-account effects rather than ordinary expenses; P&L/report totals come from journal lines.
- Backup validation checks archive safety, manifest/checksums, SQLite integrity/foreign keys/migrations, managed-file presence, and future-schema rejection before live replacement.

## Defects found and fixed

1. Landed-cost and weighted-average arithmetic could wrap an int64 intermediate or truncate an oversized `big.Int` result. All public inventory arithmetic now checks the wide intermediate and final Rial/quantity range; regression tests cover both quantity and value overflow.
2. An incoming payment could omit its customer while allocating to a customer order/invoice, producing an unowned AR journal line. The backend now derives the single allocation party and rejects cross-customer allocations; a regression verifies the AR party.
3. Purchase cancellation checked physical stock but not active reservations, allowing available stock to become invalid. Cancellation now preflights the complete batch against active reservations before inserting any compensating movement.
4. Dashboard revenue used invoice totals instead of the authoritative revenue journal. Dashboard revenue is now derived from `ACC-REVENUE` journal lines, with a reconciliation assertion.
5. Material-usage, service-sales actual cost, and production-performance reports used immutable consumption rows/job summary fields without applying compensating movement corrections. These report queries now derive actual material/waste cost from the movement ledger.
6. Customer, material, service, and machine catalog records had no safe purge boundary. Explicit backend delete APIs now purge only unreferenced records and return domain protection errors when authoritative history exists. Supplier deletion also protects payment, check, loan, and outsourced-production history. Archive remains a separate operation.
7. Attachment metadata could be deleted while referenced by proof history. Proof-linked attachment metadata is now protected, preserving the proof’s historical link.
8. Backup creation silently skipped managed files outside the application attachment root. It now fails explicitly (and rejects symlink/directory entries) instead of producing an incomplete backup.
9. New proof versions could be created directly as Approved/Rejected/Waiting, bypassing the controlled version workflow. The application boundary now requires Draft or Ready for a new version; later workflow changes use preserved version records and validated transitions.

## Verification performed

The regression suite includes the existing M0–M6 tests plus new tests for inventory overflow, reservation-safe purchase cancellation, allocation party derivation, catalog purge/protection, proof-linked attachment protection, external managed-file backup rejection, and journal-derived dashboard totals. The complete Go suite passed with a writable temporary cache.

The frontend production build and `git diff --check` are run as the final handoff validation for this audit.

## Remaining manual/native Windows validation

- Wails CLI was not available in this environment, so no local Wails build/bindings generation was run.
- Windows packaging, installer upgrade/uninstall behavior, WebView2 behavior, native file dialogs, filesystem permissions, and Windows path/rename semantics remain to be validated on Windows.
- Native WebView/browser print output and physical printer/PDF-driver behavior remain manual validation items.

No Windows installer, WebView2, native printing, or release-artifact claim is made by this audit.

## Intentionally deferred non-blocking issues

The product boundaries explicitly deferred in M6 and earlier remain deferred: backup/restore UI polish beyond the implemented workflow, release signing/auto-update, fiscal closing/reporting enhancements beyond the current journal-derived reports, payroll/tax filing, advanced scheduling, and full production/accounting expansion. These are not treated as audit failures for this release slice.
