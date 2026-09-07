# Development demo data

The demo dataset is generated only by the standalone `cmd/demo-data` command. It never runs from the production Wails API and it never targets the normal application data directory.

## Safety contract

- Generation requires `--confirm-demo`.
- Reset/cleanup requires `--confirm-reset-demo` and a matching `.atropaten-demo.json` marker.
- `--root` is mandatory in the sense that the default is an explicit temporary path; choose a dedicated path and inspect the printed result before launching the app.
- The generator refuses the normal `Atropaten` data root, broad unsafe roots, and existing unmarked roots.
- The database, managed attachment placeholder, and backups directory are created below the demo root. Nothing is committed to Git.
- All records are fictional, use Rial integer amounts and fixed-scale decimal quantities, and use seed `6006` with reference date `2026-03-21` by default.

## Generate

From the repository root:

```bash
DEMO_ROOT="${TMPDIR:-/tmp}/atropaten-demo"
go run ./cmd/demo-data generate \
  --root "$DEMO_ROOT" \
  --seed 6006 \
  --reference-date 2026-03-21 \
  --confirm-demo
```

The command prints a JSON summary with the root, seed, reference date, and generated record counts. The fixture covers parties, catalog definitions, pricing snapshots, posted purchases and inventory movements, quotes, orders, production jobs/reservations/consumption, invoices, payments, expenses, transfers, checks, loans, owners, a fiscal period, settings, attachments, and a proof record.

## Launch against the demo root

The production app reads `ATROPATEN_DATA_DIR` only as its normal path-resolution override. Set it for the demo process, never globally:

```bash
ATROPATEN_DATA_DIR="$DEMO_ROOT" go run .
```

If the Wails CLI is installed, the equivalent development launch is:

```bash
ATROPATEN_DATA_DIR="$DEMO_ROOT" wails dev
```

The Wails CLI is not available in the current validation environment, so native Wails/WebView2 launch and Windows rendering were not claimed. The repository frontend build and Go tests are still run independently.

## Reset and clean up

Reset is explicit and only accepts a previously generated, marked demo root:

```bash
go run ./cmd/demo-data reset \
  --root "$DEMO_ROOT" \
  --confirm-reset-demo
```

`clean` is an alias for the same marker-checked removal:

```bash
go run ./cmd/demo-data clean \
  --root "$DEMO_ROOT" \
  --confirm-reset-demo
```

Reset removes only that marked demo root and its generated SQLite database, attachment placeholder, and backups directory. It cannot reset the normal production database.

## Reproducibility and coverage

Use the same seed and reference date to reproduce the same fixture graph and stable domain IDs. The generator uses existing store posting workflows for purchase inventory/accounting recognition, production consumption, invoice posting, payments, expenses, transfers, checks, loans, owner transactions, and fiscal-period records. Tests verify that the root is isolated, generation is opt-in, repeated seed/reference inputs produce identical summaries, and the expected cross-domain record counts exist.

The data intentionally includes long names and notes, large Rial amounts, decimal quantities, low/zero stock, active reservations, overdue and future dates, mixed commercial/fulfillment/payment statuses, outstanding receivables/payables, print-document inputs, and managed attachment/proof references.

## Demo-data UI audit notes

The fixture is useful for checking table density, filter/sort pagination, inspectors, detail tabs, status badges, long-text wrapping, Rial formatting, decimal quantity rendering, dashboard attention panels, and print-preview entry points. No M6-005 toast/loading/confirmation behavior is introduced here.

The audit also records the current environment limitation: the standalone Vite server cannot load Wails-bound services, and the Wails CLI is unavailable here. A backend-connected browser/native pass is therefore still required to verify the live populated pages and WebView print handoff. See [UI_LAYOUT_AUDIT.md](UI_LAYOUT_AUDIT.md).
