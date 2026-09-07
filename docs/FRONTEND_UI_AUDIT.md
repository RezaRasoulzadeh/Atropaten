# Frontend UI geometry audit

Audit date: 2026-09-08. This audit covers M6-011 only: position, widths, heights, spacing, sticky placement, workspace gutters, panel geometry, master/detail proportions, inspector sizing, bottom actions, overflow, and responsive collapse.

## Runtime and dataset

The checks used the repository's opt-in `ui_preview` bridge against the isolated M6-006 dataset:

- Demo root: `/tmp/atropaten-m6-010-demo`
- Seed: `6006`
- Reference date: `2026-03-21`
- Records: 4 orders, 5 quotes, 4 customers, 3 services, 4 materials, 3 machines, 3 purchases, 3 suppliers, 2 production jobs, 2 invoices, 2 checks, 1 loan, 2 owners, 1 fiscal period, and populated accounting/report data.
- Render widths: 1024, 1280, and 1600 CSS pixels; viewport height 1000 for the audit harness.
- Register screenshots: `/tmp/atropaten-m6-011-ui-audit-final/{width}-{workspace}.png`.
- Order workspace screenshots: `/tmp/atropaten-m6-011-order/{width}-{overview|items}.png`.

The browser harness navigated every major workspace at all three widths and measured `documentElement` and `main` overflow. All 51 register renders reported `overflow: false` and `mainOverflow: false`. Native select measurements reported `text-align: start` for every rendered select.

## Shared geometry decisions

- `MasterDetail` now converges at the `xl` breakpoint (1280px): 24rem inspector by default and 30rem for wide editors. At 1024px it remains a single-column stack; this keeps the master column usable at the narrow desktop width.
- `InspectorShell` keeps the existing sticky footer placement but reserves bottom space in its body, so a long inspector cannot hide its last fields or movement/history content underneath the action bar.
- `WorkspaceStickyStack` was verified at `top: 65px` in the order editor at 1024, 1280, and 1600px, directly below the fixed-height application toolbar.
- Existing workspace gutters remain 1rem below the large-desktop shell breakpoint and 1.5rem at 1024px and above. No page-level horizontal overflow was introduced by the breakpoint change.
- Dense tables keep local horizontal scrolling where required. The Materials table measured 850px scroll width against a 710px master column at 1024px and 566px at 1280px; at 1600px it fit its 886px column. This is intentional local table overflow, not document overflow.

## Rendered workspace evidence

| Workspace/context | 1024px | 1280px | 1600px | Geometry result / remaining issue |
| --- | --- | --- | --- | --- |
| Dashboard | Rendered, populated | Rendered, populated | Rendered, populated | Two-column KPI and attention regions hold; no page overflow. |
| Orders register | Rendered, 4 records | Rendered, 4 records | Rendered, 4 records | Preserved as the baseline; register width and filter row converge cleanly. |
| New order / order overview | Rendered, ORD-1004 | Rendered, ORD-1004 | Rendered, ORD-1004 | Header, metadata strip, tabs, details, and totals use stable gutters; overview becomes two columns at 1280px. |
| Order items | Rendered, populated | Rendered, populated | Rendered, populated | Item list and persistent workspace header stay within the main column; no overflow. |
| Quotes / quote editor entry | Rendered, populated register | Rendered, populated register | Rendered, populated register | Register geometry converges; editor state was additionally exercised at 1024px by the interaction harness. |
| Production | Rendered, 2 jobs | Rendered, 2 jobs | Rendered, 2 jobs | Stacked at 1024px, master/detail at 1280px and 1600px; 24rem inspector remains readable. |
| Customers | Rendered, 3 active records | Rendered, 3 active records | Rendered, 3 active records | Master/detail now uses the same 1280px convergence point; 1024px remains stacked. |
| Services | Rendered, 3 records | Rendered, 3 records | Rendered, 3 records | Table and inspector align at 1280px; fixture emitted an existing timestamp alert, not a layout error. |
| Materials | Rendered, 4 records | Rendered, 4 records | Rendered, 4 records | Local table scroll at 1024/1280; inspector column aligns at 1280/1600 and reserves footer space. |
| Machines | Rendered, 3 records | Rendered, 3 records | Rendered, 3 records | Register/inspector collapse and side-by-side proportions are consistent. |
| Purchases | Rendered, 3 records | Rendered, 3 records | Rendered, 3 records | Wide register remains locally contained; master/detail converges at 1280px. |
| Suppliers | Rendered, 3 records | Rendered, 3 records | Rendered, 3 records | Contact register and inspector use the shared breakpoint and no page overflow. |
| Accounting | Rendered, populated overview | Rendered, populated overview | Rendered, populated overview | Tabs remain in the sticky workspace header; ledger width remains contained. |
| Invoices | Rendered, 2 records | Rendered, 2 records | Rendered, 2 records | Register and inspector stack at 1024 and split from 1280; no page overflow. |
| Expenses | Rendered through Accounting / Expenses tab | Rendered through Accounting / Expenses tab | Rendered through Accounting / Expenses tab | Tab content remains inside the Accounting workspace geometry. |
| Treasury / transfers | Rendered through Accounting / Transfers tab | Rendered through Accounting / Transfers tab | Rendered through Accounting / Transfers tab | Tab content remains inside the Accounting workspace geometry. |
| Checks | Rendered, 2 records | Rendered, 2 records | Rendered, 2 records | Filter row, register, and inspector remain contained; 1280 uses side-by-side detail. |
| Loans | Rendered with register/inspector state | Rendered with register/inspector state | Rendered with register/inspector state | 1024 stack is intentional; 1280/1600 use the shared inspector width. |
| Owners / fiscal periods | Rendered with populated tables | Rendered with populated tables | Rendered with populated tables | Multiple tables stay within their panels; no document overflow. |
| Reports / print entry | Rendered, populated reports | Rendered, populated reports | Rendered, populated reports | Report tables remain locally contained; print preview geometry was entered by the interaction harness. |
| Settings | Rendered, populated settings | Rendered, populated settings | Rendered, populated settings | Settings master/detail uses the same responsive boundary. |

## Additional interaction geometry checks

The interaction harness rendered customer form focus/error states, the material editor, order overview/items, quote editor entry, accounting tabs, loan/check/service forms, Jalali calendar, confirmation dialog, and toast states. Order overview/items were independently rendered at all three widths. The focused order measurements were:

| Width | Sticky stack top | Sticky stack width | Document overflow | Main overflow |
| --- | ---: | ---: | --- | --- |
| 1024 | 65px | 744px | false | false |
| 1280 | 65px | 1000px | false | false |
| 1600 | 65px | 1320px | false | false |

The full interaction harness reached the print-preview assertion but stopped because the existing assertion expected two `.print-document` nodes while the rendered preview did not expose that count. This is recorded as an existing print-preview test limitation; no print CSS or document structure was changed in M6-011.

## Validation and remaining geometry work

Completed for this task:

- `npm run build`
- Populated visual audit at 1024/1280/1600 for every listed workspace
- Populated order overview/items render at 1024/1280/1600
- No page-level horizontal overflow in 51 register renders
- Shared select alignment remained start-aligned in all rendered workspaces

Remaining geometry limitations are intentionally outside M6-011: the 1024px stacked master/detail layout requires page scrolling for long inspectors; dense Materials tables use local horizontal scrolling at narrow master widths; the print-preview interaction assertion needs separate investigation; and native Wails/WebView window chrome was not rendered because the browser preview is the available runtime. Typography, color, table styling, and workflow redesign remain out of scope for this task.
