# Dashboard charts and notifications

The dashboard retains its four financial summaries and production queue. It replaces Needs attention, Low stock, and Recent payments with daily sales/gross-profit lines and an active-order pipeline. The toolbar bell is available with either expanded or collapsed navigation and contains the existing operational alerts, low-stock items, and recent payment activity.

## Behavior and boundaries

- Period selection offers the last 30, 90, or 365 days. Sales and gross profit are read-only Go/SQLite projections from the same integer-Rial journal entries as the dashboard totals. Missing days have zero values, and reversals remain negative. Dates and currencies use the shared Jalali and Rial/Toman presentation utilities. An expandable semantic data table exposes exact chart values.
- The pipeline shows current unfinished orders, independent of the financial period. Draft orders form their own group; other active orders use fulfillment status. Cancelled, closed, and delivered orders are excluded.
- Orders needing attention includes unfinished confirmed orders OR unfinished orders with an order-owned `reference` attachment. One order appears once, with its reference-file count. Artwork, proof, and other attachment categories alone do not qualify a draft. Cancelled, closed, and delivered orders are excluded. Scheduled orders come first, ordered by promised date. Clicking the order number loads and opens the existing order workspace.
- Notifications refresh on mount, navigation, window focus, opening the bell, manual refresh, and every minute. They use the current date independently of the dashboard's selected historical period. Clicking an item marks it read and opens its domain workspace. Unread filtering and Mark all read are available; Escape and outside clicks close the panel.
- Notification read state lasts for the application session. Reading a notification does not change financial, inventory, or order state. A changed stock quantity or obligation amount creates a new unread presentation item. The existing backend limits on the source alert/payment lists remain in place; this is a current-activity inbox, not an unlimited notification archive.
- Errors preserve the last loaded dashboard/inbox and show an explicit stale-data message. Empty data has distinct empty states. Toasts and confirmations remain in the central feedback service.
- The existing `GetDashboard` method has additive DTO fields and regenerated Wails models. No schema migration, financial posting, inventory mutation, or production API was introduced. Shared workspace headers, panels, semantic tables, badges, controls, and theme tokens own layout and styling.

## Verification

Passed `npm run build --prefix frontend` and targeted reporting tests:

```sh
go test ./internal/storage/sqlite -run 'TestDashboard|TestReportsReconcile' -count=1
```

The added Go tests cover reference counts/deduplication, irrelevant attachment categories, excluded lifecycle states, due-date sorting, pipeline counts, zero-activity days, negative reversals, selected-period boundaries, and reconciliation with dashboard totals.

The browser audit used a newly generated marked disposable demo root with reference date 2026-09-08 and a reference attachment added only to its draft order. Screenshots were inspected at 1024 and 1600 pixels. Checks passed for viewport/workspace overflow, notification read/filter behavior, Escape dismissal, direct order navigation, period selection, chart-data disclosure, stale-data errors, and empty states. Browser checks are not native Wails/WebView2 validation.

The reusable audit is `frontend/scripts/dashboard-audit.mjs`. It requires the existing opt-in `ui_preview` backend on 127.0.0.1:4174, Vite on 127.0.0.1:5178, a recent marked demo fixture with active orders/activity, and Playwright supplied through `PLAYWRIGHT_MODULE`. `CHROME_PATH` and `DASHBOARD_AUDIT_OUTPUT` are optional overrides. The audit injects error/empty responses only in its own browser context.

`go test ./...` still fails on three pre-existing tests, reproduced using the unchanged reporting files in a separate baseline copy:

- `TestCatalogDeletionPurgesOnlyUnreferencedRecords`: referenced service deletion unexpectedly succeeds.
- `TestAssertIsolatedRootRejectsEmptyAndProductionRoot`: broad temporary root accepted.
- `TestOrderStateAxesHaveBasicTransitions`: commercial transition expectation fails.

Those behaviors are outside this dashboard change. The existing unrelated edit to the primary theme color was preserved.
