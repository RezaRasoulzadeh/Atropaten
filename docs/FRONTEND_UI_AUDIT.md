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

At the time of the M6-011 audit, the interaction harness reached the print-preview assertion but stopped because the preview child threw before exposing the expected two `.print-document` nodes. M6-014 isolated and fixed that render defect; the original assertion now passes.

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

Remaining geometry limitations are intentionally outside these tasks: the 1024px stacked master/detail layout requires page scrolling for long inspectors; dense Materials tables use local horizontal scrolling at narrow master widths; and native Wails/WebView window chrome was not rendered because the browser preview is the available runtime. Information architecture and workflow redesign remain out of scope for M6-012.

## M6-014 forms, editors, dialogs, overlays, and feedback

M6-014 was audited against the populated M6-006 bridge after the shared form/editor changes. The existing interaction harness completed 72 states (24 states at each of 1024px, 1280px, and 1600px) with no page errors and no `documentElement` or `main` overflow. Screenshots are in `/tmp/atropaten-ui-interactions/`; `results.json` records the width/state/overflow results.

### Interaction coverage

| Context/state | Rendered evidence | Result |
| --- | --- | --- |
| Customer create form | `customer-form-focus`, `validation-error-preserves-form`, `success-toast`, `confirmation-dialog`, `customer-empty` at all three widths | Focus contract measured as 1px dashed primary with no outline/shadow; failed save preserves values and shows an error; success shows a success toast; destructive confirmation is keyboard-cancellable. |
| Material editor | `material-editor` at all three widths | Shared form fields, stock/cost grouping, footer clearance, and long editor content remain contained. |
| Existing order / item editor | `order-editor`, `order-items`, `jalali-calendar` at all three widths | Metadata, notes, discount, item configuration entry, save actions, and the portaled Jalali calendar render without clipping. |
| Quote editor entry | `quote-editor` at all three widths | Quote register/workspace entry remains inside the accepted master/detail geometry; item configuration uses the shared editor primitive. |
| Services and configurator | `service-form`, `service-parameter-editor` at all three widths | Service description, parameter, pricing, cost, and manual-money inputs now use shared input/textarea primitives while retaining recalculation behavior. |
| Checks, loans, accounting, owners | `check-form`, `loan-form`, `loans-populated`, `accounting-*`, `owners-register` at all three widths | Lifecycle forms, Jalali fields, money controls, schedules, and table-backed details render without duplicate nested labels or page overflow. |
| Print preview | `print-preview` at all three widths; print media assertion | Both screen and teleported print documents render; `#app` is hidden and `.print-output` is visible in print media. |
| Busy/load/error feedback | `loading`, `page-error`, validation and success states at all three widths | Busy controls disable during the delayed mutation; load failure is visible as an error alert; error and success feedback are not console-only. |

The print-preview failure was a real UI/data-boundary defect, not a stale assertion. The M6-006 response can contain `statementLines: null` (and nullable allocation data), while `PrintDocument` read `.length` and iterated the value unconditionally. That child threw during render, so the expected two preview documents never mounted. The preview now treats nullable line collections as empty arrays; the original two-document assertion remains unchanged and passes.

The source audit covered New/existing Orders, Quotes, Customers, Materials and stock adjustment, Services/configurator/parameters/costs, Machines, Suppliers, Purchases, Production actions, accounting payment/invoice/expense/transfer flows, Checks lifecycle, Loans/schedules/payments, Owners/fiscal views, Reports/print preview, and Settings. Mutations consistently use the shared `runAction` busy guard or an equivalent local guard, errors go through `reportError`/`normalizeError` or the shared toast service, and no `window.confirm` remains. Destructive flows use `confirmAction` and the global native `ConfirmDialog`, which restores focus when the initiating element still exists and supports Escape cancellation.

The form/editor primitive cleanup replaced the remaining important raw textareas in Customers, Orders, Quotes, Materials, Machines, Purchases, Checks, Production, Accounting, Loans, Owners, Suppliers, Services, and the order-item configurator with `AppTextarea`. Money, quantity, service-pricing, and configurator text inputs were routed through `AppInput` while retaining the existing parsing and recalculation callbacks. Nested `FormField`/`SelectField` label structures were removed from Checks and Purchases. `AppTextarea` now forwards normal attributes centrally, and the confirmation dialog has explicit alert-dialog semantics without changing its native top-layer behavior.

### Remaining M6-014 exceptions

- The service definition and configurator remain information-dense because their parameter/component/pricing editors expose many domain fields; they are structurally grouped but are candidates for a later workflow redesign.
- Some low-level checkbox controls remain native inputs by design; they are not editable text controls and do not receive the dashed text-control focus treatment.
- The existing browser harness exercises representative mutations and source-audits the remaining action paths, but it does not submit every destructive/lifecycle action against a failure fixture. A future final QA pass should expand those backend-failure scenarios.
- Toasts from earlier intentionally injected failure states remain visible while the harness moves between workspaces; this is expected test-fixture persistence, not duplicate application feedback.
- Reports and print preview are now robust for nullable line collections, but broader print layout/content coverage beyond the representative invoice remains for final QA.

## M6-015 final repository-wide status

Audit date: 2026-09-08. This is the final M6 visual QA record. The accepted M6-011 geometry, M6-012 typography/control metrics, M6-013 data surfaces, and M6-014 form/feedback contracts were preserved. No new visual direction or information architecture was introduced.

### Evidence and status rules

- The populated M6-006 bridge was used for the repository-wide browser audits. The clean visual audit rendered 17 major workspaces at 1024px, 1280px, and 1600px: 51 renders total. Results and screenshots are in `/tmp/atropaten-m6-015-visual/`.
- The data-surface audit rechecked local table overflow, native select alignment, editable-control focus styling, and representative selected/focused/no-results states at all three widths. Results are in `/tmp/atropaten-m6-015-data/`.
- The typography/control audit rechecked `getBoundingClientRect()` and computed styles for toolbar, forms, date, money/quantity, inspector, dialog, and inline actions. Results and screenshots are in `/tmp/atropaten-m6-015-typography/`.
- The clean M6-014 interaction run covered 72 form/dialog/popover/feedback states (24 per width) in `/tmp/atropaten-ui-interactions/`. Its source was unchanged by the final product fix; the final M6-015 source changes affect only print output and audit harnesses. A later replay used the already-mutated failure-audit demo root and is not counted as a fresh clean-dataset run.
- `final-lifecycle-failure-audit.mjs` passed seven injected backend-failure cases: customer delete, material stock adjustment, purchase cancel, payment reversal, check transition, loan creation, and owner delete. Each case verified one request, cleared busy state, visible error feedback, no success feedback, and no page error.
- `final-print-audit.mjs` passed quote, invoice, payment receipt, customer statement, and supplier statement at 1280px. Screenshots and results are in `/tmp/atropaten-m6-015-print/`. Print media hides the application shell, shows only the prepared document, keeps paper/table backgrounds white, and preserves headings, metadata, lines/allocations, totals, long names, and statement balances.

| Context | Geometry | Typography | Controls | Data surface | Forms / feedback | Rendered widths | Native-only gap |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Dashboard | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Orders register | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| New Order / order workspace | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Quotes register | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Quote workspace | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Production | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Customers | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Services / configurator | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | Native OS text measurement only |
| Materials | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Machines | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Purchases | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Suppliers | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Accounting overview | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Accounting Accounts | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Accounting Journal | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Accounting Payments | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Invoices | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Expenses | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Transfers / Treasury | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Checks | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Loans | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Owners / fiscal | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Reports | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | None |
| Settings | PASS | PASS | PASS | PASS | PASS | 1024 / 1280 / 1600 | Native OS settings chrome only |
| Print preview: quote, invoice, receipt, statements | PASS | PASS | PASS | PASS | PASS | 1280 print media; entry at 1024 / 1280 / 1600 | Native printer dialog and WebView print handoff |

### Final findings and remaining gaps

The last product defect found was in print output, not in the workspace layouts: `PrintDocument` assumed nullable line/statement/allocation collections were always present and the screen preview omitted payment allocation and document metadata. The renderer now treats those collections safely, prints status/due/method/account/payment metadata, prints receipt allocations, and uses a white print-only canvas so dark application surfaces cannot bleed into paper output.

The Services/configurator surface is intentionally dense because parameters, material/cost components, pricing, and actions are all part of one domain workflow. It is grouped and aligned, but remains the clearest candidate for a future workflow redesign; no final-QA visual defect was found. The M6-006 service timestamp warning remains a fixture/domain validation issue: generated component literals omit timestamps even though the service API returns valid service timestamps. It is not a CSS or rendering failure and was not hidden.

Exact remaining gaps for later work are limited to native-environment verification and data-fixture fidelity: the Windows/Wails native printer dialog and WebView chrome were not available here; long multi-page pagination was checked through representative rendered content rather than a native printer; and the quote demo snapshot exposes `CUS-DEMO-03` instead of the customer display name. The backend fixture/data issue and the intentionally dense Services workflow remain outside this final visual convergence change. No known repository-wide geometry, typography, control, data-surface, form, feedback, theme, or print-preview defect remains from this audit.

## Estedad font replacement audit

Date: 2026-09-08. The application font is now the locally vendored Estedad variable WOFF2 (`Estedad`, weight axis 100–900). The previous Vazirmatn assets and declarations were removed after confirming no source references remained. The shared typography contracts were not changed: 24/32 workspace titles, 14/20 body/control text, 12/16 labels and metadata, 40px standard controls, 32px compact controls, natural metrics, and the existing dashed Amber editable-control focus contract remain in effect.

Chrome verified `document.fonts.status === "loaded"`, `document.fonts.check("14px Estedad") === true`, computed root/body `font-family: Estedad, system-ui, sans-serif`, and one loaded face with `font-weight: 100 900` at 1024px, 1280px, and 1600px. The existing typography/control audit passed with no page errors. The visual audit rendered Dashboard, Orders, Customers, Materials, Services, Accounting, Checks, Reports, and the remaining major workspaces at all three widths; screenshots are in `/tmp/atropaten-estedad-visual/` and metrics are in `/tmp/atropaten-estedad-typography/`.

Visual inspection found no shared metric change necessary: input/select/button heights remained 40px, compact controls 32px, labels and metadata retained 12/16, body/control text retained 14/20, and title/section baselines remained stable. No transform, negative margin, asymmetric padding, metric override, or page-specific compensation was introduced.
