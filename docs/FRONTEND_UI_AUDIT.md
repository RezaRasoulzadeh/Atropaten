# Frontend UI audit

Audit started 2026-09-08. Baseline: Dashboard and Orders register. Existing dirty changes to Customers, Materials, Order workspace and four primitives were present before this task and are preserved.

## System findings

Many former layout classes were removed without replacing their structure. The result is ungrouped forms, inline primary/secondary metadata and stacked inspectors. `form-control`, `label-text`, and `*-bordered` are stale DaisyUI-era classes. `WorkspaceStickyStack` styles arbitrary descendant headers, coupling typography to DOM position. `SectionPanel` merely proxies `AppPanel`. Table wrappers do not enforce density, keyboard selection or numeric alignment. Several finance workflows are compressed into one source line, concealing large components. Several failures only set local strings; some rejected promises are unhandled. Confirmation lacks native focus containment. Select attributes land on a wrapper rather than the control. Panels clip calendar popovers.

## Page inventory and target

All rows are source-audited. Render and remaining-defect columns will be updated with actual browser evidence, not inferred from compilation.

| Context | Baseline problems | Chosen pattern / reuse | Remove or replace | Rendered in this task | Remaining |
| --- | --- | --- | --- | --- | --- |
| Dashboard | Preserve density; date/filter overflow risk | KPI grid, AppPanel, WorkspaceHeader | Header descendant CSS | Pending | Browser pass |
| Orders | Good reference; duplicated register structure | Rich register, SearchFilterBar, StatusBadge | One-off row wrappers | Pending | Browser pass |
| New order / editor | Large coordinator; mixed summary/form geometry | Full editor, FormSection, InspectorShell, sticky actions | Nested summary containers | Pending | Browser pass |
| Quotes / editor | Raw header, zebra table, ungrouped editor | Rich register; full editor with summary | Zebra table, raw header | Pending | Rebuild |
| Production | Unstyled queue buttons; stacked forms; queue filter not wired | Rich register + inspector sections | Ad hoc queue and inspector wrappers | Pending | Rebuild |
| Customers | Existing register refactor; form metrics/overflow | Rich register + inspector | Repeated contact fields | Pending | Browser pass |
| Services | Large definition/parameter/component editor | Dense catalog + feature-owned editors | Giant single workflow view | Pending | Split and render |
| Materials | Existing table refactor; local editor groups | Dense numeric table + inspector | Repeated inventory field geometry | Pending | Browser pass |
| Machines | Zebra table; no inspector geometry | Dense rate table + inspector | Raw form/header wrappers | Pending | Rebuild |
| Purchases | Wide table; stacked editor; many line controls | Dense table + editor | Zebra, unstyled lines | Pending | Rebuild |
| Suppliers | Raw forms and metadata; no busy guard | Rich contact register + inspector | One-off contact editor | Pending | Rebuild |
| Accounting | Raw accounts/journal/payments | Dense ledger tables, tabs, KPI overview | Unstyled ledger divs | Pending | Rebuild |
| Invoices | Zebra; unreadable detail hierarchy | Dense table + snapshot inspector | Raw line/summary wrappers | Pending | Rebuild |
| Expenses | Placeholder-only fields; raw history | Dense history + compact form | Repeated posting form geometry | Pending | Rebuild |
| Treasury / transfers | Raw history and posting form | Dense history + compact form | Raw wrappers | Pending | Rebuild |
| Checks | Duplicate labels; unstyled lifecycle actions | Dense table + lifecycle inspector | Zebra and ad hoc detail wrappers | Pending | Rebuild |
| Loans | No load state; raw schedules/payment fields | Dense table + schedule inspector | Zebra, placeholder-only fields | Pending | Rebuild |
| Owners | Oversized multiflow view; unhandled reversal | Dense table + feature-owned finance panels | Raw rows/forms | Pending | Split and render |
| Reports | Wide raw tables; flat summaries | Dense table + compact filters/KPIs | Zebra; raw tabs | Pending | Rebuild |
| Settings | Ungrouped identity; silent backup errors | Grouped settings form + backup panel | Raw field stacks | Pending | Rebuild |
| Print entry / preview | ID input; unstyled document; whole-shell printing risk | Neutral preview entry + dedicated document | Raw print article | Pending | Browser print emulation; native unavailable |

## Verification plan

Render at 1024, 1280 and 1600 CSS pixels with the M6-006 dataset. Check long names, large money, decimal quantities, filters, empty/loading/error states, forms, focus borders, inspector overflow, dialogs and toasts. Run `go test ./...`, `npm run build` and `git diff --check`. There is no frontend test script in the baseline package.json.
