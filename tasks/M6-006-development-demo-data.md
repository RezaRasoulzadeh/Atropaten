# M6-006 — Safe development demo data

## Objective
Create a realistic, reproducible, disposable dataset covering Atropaten's major workflows so the entire UI can be inspected with populated tables, long values, varied statuses, financial records, and historical activity. This is a development/testing facility, not production sample data or a new business feature.

## Safety and isolation
- Never seed the user's live database or overwrite existing records.
- Provide an explicit development-only CLI command or test fixture runner that creates a dedicated temporary/demo application-data root and SQLite database. Do not use the normal production data path.
- Require an explicit opt-in for generation and for resetting an existing demo dataset. Refuse unsafe paths and never delete an arbitrary directory.
- Do not expose a production Wails seed/reset API or a hidden production UI action.
- Mark the dataset clearly as DEMO in its shop identity and documentation. Use fictional names, contacts, references, and locally generated files only.
- Use a fixed seed and reference date for deterministic generation. Allow a documented optional reference date/seed if useful.
- Make generation reproducible and resettable without affecting real application data. Document how to launch the app against the demo root and return to normal data.

## Dataset coverage
Use existing application/domain services and authoritative posting workflows wherever practical. Do not directly fabricate ledger balances, inventory totals, payment statuses, or other derived business values. Where fixtures require direct persistence, use the existing validated repository boundaries and document why.

Generate realistic interconnected records across:
- Shop identity/settings and representative managed artwork/attachments/proof versions.
- Customers and suppliers, including short/long names, long notes, multiple contacts, active/archived records, and records with no activity.
- Materials, units/conversions, services, dynamic parameters, pricing rules, cost components, machines, and representative catalog configurations.
- Purchases, inventory movements, weighted-average costs, low/zero stock, reservations, adjustments, and historical consumption/waste.
- Quotes, mixed-item orders, production jobs, outsourcing, due/overdue/ready/completed statuses, and immutable pricing/cost snapshots.
- Invoices, partial/multiple payments, customer credit, supplier advances, expenses, cash/bank transfers, and journals.
- Incoming/outgoing checks in valid lifecycle states, loans/installments with paid/upcoming/overdue cases, owner capital/drawings/loans, fiscal periods and allocations where supported.
- Reports, statements, dashboard attention items, and printable documents backed by real persisted records.

Aim for useful density rather than millions of rows: dozens of customers/suppliers/catalog records, enough orders and transactions to populate several table pages, and a meaningful spread of statuses and dates. Include long text, large grouped Rial values, zero values, decimal quantities, empty optional fields, and realistic edge cases without violating domain invariants.

## UI inspection purpose
The dataset must make it possible to inspect every major workspace with populated content, filters, sorting, inspectors, tabs, detail views, and print previews. Include records that expose narrow columns, long labels, multi-line text, badges, dense tables, and form/inspector sizing problems. Preserve the ability to test genuine empty/no-result states by filtering or using a separate empty demo fixture.

Do not implement a new visual redesign or M6-005 feedback system in this task. Record any confirmed UI defects discovered during generation in a concise demo-data audit document, but keep unrelated UI repairs for the existing UI tasks.

## Verification
- Generation completes from a clean isolated database and produces the same deterministic dataset on repeat runs.
- Normal production data paths are never modified.
- Verify foreign keys, journal balance, inventory movement reconstruction, reservation availability, invoice/payment balances, and representative report reconciliation.
- Verify generated attachments exist under the demo managed-file root and can be included in a backup.
- Add focused tests for safety, determinism, and representative cross-domain invariants.
- Document exact generation, launch, reset, and cleanup commands in docs/DEMO_DATA.md.
- Run go test ./..., frontend production build, and git diff --check. State unavailable Wails/native validation explicitly.

## Delivery
Commit to main, push origin/main, and print the final SHA. Do not include generated demo databases, backups, or large binary fixtures in Git.