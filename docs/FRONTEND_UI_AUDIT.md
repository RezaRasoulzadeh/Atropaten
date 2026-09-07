# Frontend UI geometry and typography audit

Audit date: 2026-09-08. This audit preserves the accepted M6-011 geometry baseline and adds the M6-012 repository-wide typography and control-metrics evidence.

## Runtime and dataset

The checks used the repository's opt-in `ui_preview` bridge against the isolated M6-006 dataset:

- Demo root: `/tmp/atropaten-m6-010-demo`
- Seed: `6006`
- Reference date: `2026-03-21`
- Records: 4 orders, 5 quotes, 4 customers, 3 services, 4 materials, 3 machines, 3 purchases, 3 suppliers, 2 production jobs, 2 invoices, 2 checks, 1 loan, 2 owners, 1 fiscal period, and populated accounting/report data.
- Render widths: 1024, 1280, and 1600 CSS pixels; viewport height 1000 for the audit harness.
- Register screenshots: `/tmp/atropaten-m6-011-ui-audit-final/{width}-{workspace}.png`.
- Order workspace screenshots: `/tmp/atropaten-m6-011-order/{width}-{overview|items}.png`.
- M6-012 control screenshots: `/tmp/atropaten-m6-012-controls-final/{width}-{dashboard|customer-form|material-form|order-detail|jalali-calendar|inspector-jalali-calendar|confirmation-dialog}.png`.
- M6-012 metric results: `/tmp/atropaten-m6-012-controls-final/results.json`.

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

## M6-013 data-surface convergence

M6-013 keeps the accepted geometry and typography baselines and standardizes data-heavy workspaces around two patterns only: the Orders-style rich register and the shared semantic dense table. No new global styling layer was added. The shared `DataTable`, `DataTableRow`, `DataTableCell`, `RegisterList`, `RegisterRow`, `StatusBadge`, `MasterDetail`, `InspectorShell`, `InspectorHeader`, and `InspectorSection` primitives provide the common surface language.

| Workspace | Chosen surface pattern | Inspector/detail treatment | Rendered evidence |
| --- | --- | --- | --- |
| Orders | Rich register (baseline preserved) | Existing order detail preserved | `/tmp/atropaten-m6-013-final/{width}-orders.png` |
| Quotes | Rich register | Shared inspector/editor shell | `/tmp/atropaten-m6-013-final/{width}-quotes.png` |
| Production | Rich register | Shared inspector with cost/schedule and reservation sections | `/tmp/atropaten-m6-013-after/{width}-production.png` |
| Customers | Rich register | Contact and notes sections in `InspectorShell` | `/tmp/atropaten-m6-013-after/{width}-customers.png` |
| Suppliers | Rich register | Contact details section in `InspectorShell` | `/tmp/atropaten-m6-013-after/{width}-suppliers.png` |
| Services catalog | Dense table | Parameter and cost-component inspector sections | `/tmp/atropaten-m6-013-after/{width}-services.png` |
| Materials | Dense table | Stock/cost and movement sections; movement ledger remains a dense table | `/tmp/atropaten-m6-013-after/{width}-materials.png` |
| Machines | Dense table | Rate definition inspector section | `/tmp/atropaten-m6-013-after/{width}-machines.png` |
| Purchases | Dense table | Existing editor/inspector retained; item and amount hierarchy stays inside the shell | `/tmp/atropaten-m6-013-after/{width}-purchases.png` |
| Accounting accounts/journal/payments | Dense tables | Existing tab/panel containment retained | `/tmp/atropaten-m6-013-final/{width}-accounting.png` |
| Invoices | Dense table | Invoice lines and totals use inspector sections; ready-to-invoice queue remains a shared register-like action surface | `/tmp/atropaten-m6-013-after/{width}-invoices.png` |
| Expenses/transfers/treasury | Dense tables | Existing Accounting tab panels retained | `/tmp/atropaten-m6-013-final/{width}-accounting.png` |
| Checks | Dense table | Lifecycle history converted to a dense table in the inspector | `/tmp/atropaten-m6-013-after/{width}-checks.png` |
| Loans | Dense table | Installment schedule and payment history converted to dense tables | `/tmp/atropaten-m6-013-after/{width}-loans.png` |
| Owners finance/history | Dense tables | Existing owner inspector and transaction tables retained | `/tmp/atropaten-m6-013-final/{width}-owners.png` |
| Reports | Dense table | Report result table stays inside its panel | `/tmp/atropaten-m6-013-final/{width}-reports.png` |

### Rendered data-surface checks

The full visual audit rendered every affected context at 1024px, 1280px, and 1600px with the populated M6-006 dataset. It measured `documentElement` and `main` overflow, table scroll/client widths, and native select alignment. All 45 context/width renders reported no page-level or main-region overflow, and all native selects reported `text-align: start`. Dense table overflow was confined to `.data-table` where the master/detail column could not show all deliberate comparison columns; this was checked on Materials, Purchases, Invoices, Checks, and the supporting inspector schedules.

The focused state audit is in `/tmp/atropaten-m6-013-data-surfaces-final/results.json` with screenshots for selected, hover, keyboard-focus, and no-results states for representative registers at all three widths. It also checked computed editable-control focus styles: 1px dashed Amber border, no box shadow, and no layout-changing border width. Representative populated checks included selected rows, long Persian names, large Toman/Rial values, quantities with unit metadata, multiple status badges, empty/no-results states, open inspectors, and local table scrolling. Screenshots were visually inspected for Services, Materials, Machines, Purchases, Invoices, Checks, Loans, Owners, Reports, Customers, Production, and the preserved Orders/Quotes baseline.

The dense-table primitive now prevents character-level wrapping in numeric/status columns, keeps status badges intact, aligns numeric cells at the end, and makes table width explicit so scrolling remains local. Lifecycle, installment, payment, invoice-line, stock-movement, and inspector metadata surfaces use the same neutral-divider/table hierarchy. `rg` found no `table-zebra`, `:has()` alignment workaround, or font metric override in frontend source.

Remaining defects for later redesign are page-specific workflow and form concerns owned by M6-014: purchase editing is still a combined editor rather than a redesigned form flow, production/service configuration controls remain information-dense, and the 1024px stacked inspectors require normal page scrolling for long content. The M6-006 Services fixture continues to emit its existing timestamp validation alert; it is a dataset/domain fixture issue, not a data-surface rendering failure. No table aesthetic or business workflow redesign was started here.

## M6-012 typography and control metrics

The same populated M6-006 browser bridge rendered the representative control states at 1024px, 1280px, and 1600px. The control audit collected `getBoundingClientRect()` and computed styles for each state, then screenshots were inspected for Vazirmatn optical centering and mixed-row baseline alignment.

| Representative control | Rendered metric at all three widths | Visual result |
| --- | --- | --- |
| Toolbar search, currency select, and icon action | 40px height; 14px / 20px; 1px solid neutral border at rest | Aligned on one row; select values remain start-aligned. |
| Normal form input and select | 40px height; 14px / 20px | Matches toolbar controls and shared field labels. |
| Textarea | 80px baseline fixture height; 14px / 20px | Natural row sizing remains intact without font compensation. |
| Jalali date field | 40px height; 14px / 20px | Calendar icon is centered with the shared small-button geometry. |
| Money / quantity field | 40px height; 14px / 20px | Numeric values align with adjacent fields without offsets. |
| Focused editable control | 40px; 1px dashed primary border; no outline or shadow | No layout shift; validation color selectors remain authoritative. |
| Dialog controls and register rows | Dialog buttons use the shared 14px / 20px button text; register rows preserve the compact list baseline | Keyboard-visible dialog focus and register action alignment remain readable. |

Shared type scale now uses a 24px / 32px workspace title, 14px / 20px section and body text, and 12px / 16px labels, metadata, help text, table headers, and status badges. Standard controls use the shared 40px density; small buttons/tabs use 32px. The normal button line-height now matches fields without changing M6-011 workspace gutters or master/detail widths.

The local Vazirmatn files use their natural font metrics. The earlier vertical drift came from `ascent-override`, `descent-override`, and `line-gap-override` declarations that changed the browser's font box; those declarations remain absent, and no transform, negative margin, or asymmetric padding compensation was added.

The focused-control check covered inputs, selects, textareas, search fields, date fields, money/quantity fields, inspector controls, dialog buttons, and register/table contexts. Native selects reported `text-align: start` at every workspace and width. Ordinary buttons/cards/containers do not receive the dashed editable-control treatment.

`InspectorShell` was explicitly tested with the Checks editor's Jalali calendar at all three widths. The M6-011 footer clearance and 24rem/30rem geometry remain unchanged; the shell no longer clips floating content, and the shared calendar is portaled to the document layer with viewport collision positioning. Native select menus remain browser top-layer UI, while the confirmation dialog remains a native top-layer dialog.

Every major workspace in the M6-011 table was re-rendered after the shared metric changes. The populated Dashboard and Orders register remain the visual baseline; Customers, Materials, order detail/editor, production, services, machines, purchases, suppliers, accounting tabs, invoices, checks, loans, owners, reports, and settings were checked for shared header/control/table type alignment at all required widths. No page-level overflow or select alignment regression was observed.

Remaining exceptions are intentional or page-specific: textarea heights still follow their requested row counts, some dense table/register content wraps at narrow widths, the existing print-preview interaction assertion still needs separate investigation, and detailed workflow/page redesign remains outside M6-012.

## Validation and remaining geometry work

Completed for this task:

- `npm run build`
- Populated visual audit at 1024/1280/1600 for every listed workspace
- Populated order overview/items render at 1024/1280/1600
- No page-level horizontal overflow in 51 register renders
- Shared select alignment remained start-aligned in all rendered workspaces

Remaining geometry limitations are intentionally outside these tasks: the 1024px stacked master/detail layout requires page scrolling for long inspectors; dense Materials tables use local horizontal scrolling at narrow master widths; the print-preview interaction assertion needs separate investigation; and native Wails/WebView window chrome was not rendered because the browser preview is the available runtime. Information architecture and workflow redesign remain out of scope for M6-012.
