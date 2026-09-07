# M6-009 visual audit

Date: 2026-09-07

## Method and verification boundary

The shared shell and Dashboard were rendered in standalone Chrome at 1024, 1280, and 1600 CSS pixels after the corrective changes. The standalone Vite app has no Wails bridge, so API-backed pages show their existing normalized bridge error/empty states rather than M6-006 populated records. The Wails CLI and native WebView2 runtime are not installed in this workspace; populated demo-data rendering, native focus traversal, print handoff, and Windows WebView2 behavior remain unverified.

Every workspace below was inspected in source and through its shared component structure. “Not inspected” means not live-rendered in a backend-connected runtime; it does not mean the view was omitted from the corrective review.

The shared-control visual check also covers normal and focused states for the AppToolbar search input, currency select, `SelectField`, `AppTextarea`, date-picker input, and configurator inputs. The central contract is a 1px solid `base-300` border at rest and a 1px dashed Amber primary border on focus, with `outline: none` and `box-shadow: none`; the error/success color classes remain visible on focus. Ordinary buttons, cards, and containers are outside this selector set.

| Workspace | Live render | State checked | Header / toolbar | Forms / controls | Tables / panels | Inspector | Deferred page-specific design work |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Dashboard | Inspected: 1024 / 1280 / 1600 | Empty/error state; populated M6-006 unavailable | Checked | Checked | Checked | N/A | Real populated attention density and backend-connected KPI values |
| Orders | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Checked in source | Order workspace checked in source | Final register row density and order-detail information hierarchy |
| Quotes | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Checked in source | Quote workspace checked in source | Quote item editor and pricing-summary refinement |
| Production | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Checked in source | Job inspector checked in source | Reservation/consumption workflow grouping and dense job-row refinement |
| Customers | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register structure checked in source | Checked in source | Contact history and bulk relationship actions |
| Services | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register/configurator panels checked in source | Checked in source | Configurator step hierarchy and pricing-rule editor redesign |
| Materials | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register and movement panels checked in source | Checked in source | Inventory movement visualization and stock-adjustment flow |
| Machines | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register panel checked in source | Checked in source | Machine rate comparison and capacity details |
| Purchases | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register and line editor checked in source | Purchase editor checked in source | Purchase-line editing and totals layout |
| Suppliers | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register panel checked in source | Inspector/editor checked in source | Supplier activity and purchasing-history presentation |
| Accounting | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Tabs, journal, and history panels checked in source | Finance detail surfaces checked in source | Ledger navigation and account drill-down redesign |
| Invoices | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register checked in source | Invoice inspector checked in source | Receivables aging and line-level document treatment |
| Expenses | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked through Accounting extras source | Checked in source | History panel checked in source | N/A | Expense categorization and reversal history presentation |
| Treasury / payments | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked through Accounting extras source | Checked in source | History panel checked in source | Payment detail checked in source | Treasury reconciliation and allocation workflow |
| Checks | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register and lifecycle history checked in source | Check inspector checked in source | Lifecycle timeline and instrument actions |
| Loans | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Register and installment panels checked in source | Loan inspector checked in source | Schedule visualization and payment allocation redesign |
| Owners | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Owner/fiscal-period tables checked in source | Owner detail checked in source | Ownership, capital, and allocation information hierarchy |
| Reports | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Report table checked in source | Print-preview entry surface checked in source | Report-specific summaries and print-document styling |
| Settings | Not inspected: Wails bridge unavailable | Empty/error code path checked; populated unavailable | Checked in source | Checked in source | Backup/restore panel checked in source | N/A | Settings information architecture and backup progress states |
| Print-preview entry surfaces | Not inspected: native print runtime unavailable | Source/empty entry paths checked; populated documents unavailable | Checked in Reports and document metadata components | Checked in source | Document tables checked in source | Print document source checked in source | Native print layout, paper sizes, and Windows print dialog output |

## Corrective results

- Vazirmatn is still locally vendored and now uses its actual font metrics. The `ascent-override`, `descent-override`, and `line-gap-override` declarations were removed. The asymmetric control padding and global metric compensation were removed as well.
- `SelectField` is now the one ordinary-select pattern: a native DaisyUI `select select-bordered` with shared height, typography, border, focus behavior, and `text-start` value alignment. The old custom trigger and generic `.btn`/`:has()` centering rules are gone.
- Editable controls now override DaisyUI’s default focus outline/ring centrally: 1px dashed Amber primary border, no outline or shadow, and no box-size change. Validation state colors remain visible, and ordinary buttons are not included.
- The base cleanup reduced `frontend/src/style.css` from 209 to 117 lines (92 lines, 44%) before the shared focus contract was added. It is now 168 lines: still 41 lines smaller than the original, with only the centrally shared editable-control focus rules added back. Feature and other control geometry remains in DaisyUI and local Tailwind structure.
- The existing toast, confirmation, date, currency, and Wails API paths were preserved. No Go/domain files were changed.

## Remaining verification

Run the M6-006 demo-data fixture against a Wails-connected development process and repeat the populated checks at 1024, 1280, and 1600px. The deferred items above are later page-design work, not reasons to add more global CSS during this stabilization pass.
