# M6-002 — Backup, restore, Windows release, and upgrade hardening

## Objective

Complete the product-level M6 deliverables needed to ship Atropaten as a recoverable Windows desktop application.

This task adds authoritative backup/restore, upgrade compatibility checks, release-safe storage locations, Windows packaging configuration, and repeatable release smoke tests. It must preserve all existing M1-M6-001 behavior and data.

## Core invariants

- User data must survive application upgrades.
- Backup/restore must reproduce the authoritative SQLite database and managed attachment/document files included by Atropaten.
- Restore must never partially replace live application data.
- Restore must validate the backup before switching the application to it.
- Existing immutable financial/inventory/history invariants remain unchanged after backup/restore.
- Release builds must not depend on development tools being installed.
- Production data must not be stored inside the installed application directory.
- Database migrations remain forward-only and deterministic.

## Scope

### 1. Production data directory

Introduce one centralized Go path resolver for runtime application data.

Windows production data should live in an appropriate per-user application-data directory, not beside the executable and not in the source tree.

The resolver must cover at least:
- SQLite database
- managed attachments/artwork root if Atropaten owns those files
- backup default directory
- logs/config only if already appropriate to manage here

Development mode may retain a convenient local path, but the behavior must be explicit and testable.

Do not scatter OS path logic across services.

### 2. Managed attachments/files boundary

Inspect the existing attachment/proof implementation and define exactly which files Atropaten owns and therefore backs up.

Requirements:
- database metadata remains authoritative for attachment references
- backup includes managed files that are required to reproduce application state
- external user-referenced files that Atropaten intentionally does not own must not be silently copied unless the current architecture already treats them as managed
- missing referenced files should be reported clearly during backup validation rather than silently ignored

Document this boundary.

### 3. Backup format

Implement a versioned backup package suitable for a local Windows desktop application.

A backup should contain at least:
- a consistent SQLite database snapshot
- required managed attachment/artwork files
- manifest with backup format version
- application/schema version
- creation timestamp
- file list and integrity metadata/checksums

Prefer a single portable archive file unless there is a strong technical reason otherwise.

Do not copy a live SQLite file unsafely. Use SQLite's supported backup/checkpoint strategy so the captured database is internally consistent even when WAL mode is enabled.

### 4. Create backup

Add a Go backup service that:
- creates backup into a temporary destination first
- obtains a consistent DB snapshot
- copies managed files
- writes manifest/checksums
- validates the finished package
- atomically publishes/renames the completed backup
- cleans temporary artifacts on failure
- never mutates live application data

Return useful metadata:
- path
- created time
- database/schema version
- size
- managed-file count

### 5. Backup verification

Implement backup verification independently of restore.

Validate:
- archive structure
- manifest format/version
- checksums
- database can be opened
- SQLite integrity check succeeds
- schema version is known/supported
- required managed files are present

Corrupt/incomplete backups must be rejected with actionable errors.

### 6. Restore

Restore is destructive and must be explicit.

Required workflow:
1. select backup
2. validate completely before touching live data
3. restore into a temporary application-data location
4. open DB and run/validate supported migrations there if necessary
5. run SQLite integrity/foreign-key checks
6. verify managed files
7. close active DB/services safely
8. atomically switch current live data to restored data while retaining a rollback copy until success is established
9. reopen the application database/services
10. remove rollback copy only after successful reopen

If any stage fails, the original live data must remain usable.

If in-process hot restore is unsafe with current architecture, implement a controlled restart-required restore flow instead of pretending it is safe.

### 7. Upgrade/migration compatibility

Add explicit release upgrade tests covering representative older databases.

At minimum test opening/migrating database fixtures or generated states from important historical versions, including:
- pre-accounting version where practical
- v8 production
- v10 accounting/invoices
- v11 checks/loans
- v12 owners/fiscal periods
- current v13

Requirements:
- each supported old DB reaches current schema automatically
- user data remains intact
- migrations are idempotent on reopen
- unsupported future schema versions are rejected rather than modified

Do not invent lossy downgrade support.

### 8. Database startup hardening

On startup:
- create production data directory safely
- open DB
- reject future/unknown schema versions clearly
- apply forward migrations transactionally
- enforce foreign keys
- run lightweight sanity checks appropriate for startup

Do not run expensive full integrity scans on every launch unless justified.

### 9. Backup & Restore UI

Add a real Settings / Backup & Restore surface.

Show:
- current data location
- current schema/application version
- last backup information if available
- Create Backup
- Verify Backup
- Restore Backup

Restore UI must:
- explain that live data will be replaced
- require explicit confirmation
- show backup timestamp/version/size before confirmation
- report restart requirement when applicable
- clearly display failure without leaving the app in an ambiguous state

Use native Wails file/directory dialogs where appropriate rather than browser-only fake paths.

Keep Wails bindings thin; backup logic stays in Go application/storage services.

### 10. Windows packaging

Review and complete the Wails Windows build configuration.

Requirements:
- correct application/product name: Atropaten
- version source centralized/documented
- executable metadata reasonable
- application icon configured if a suitable project icon exists; otherwise keep this isolated so final artwork can be substituted without code changes
- installer/bundle configuration suitable for a normal Windows user
- installed app writes runtime data only to the production data directory
- uninstalling/reinstalling the app must not silently delete user data

Use Wails-supported packaging. Do not add a large custom installer framework unless required.

### 11. Release build commands/documentation

Add/update release documentation with exact commands for a clean Windows release build.

Document prerequisites and expected artifacts.

Where platform limitations prevent building the Windows installer in the current Linux/macOS Codex environment, still make all configuration/source changes possible and state exactly what remains to be executed on Windows.

Do not claim a Windows installer was validated unless it was actually built/run on Windows.

### 12. Release smoke-test harness/checklist

Add a concise repeatable release smoke test covering the representative shop workflow:
- fresh install/startup
- configure shop identity
- create material/service/customer/supplier
- purchase/post inventory
- create order and production job
- consume material/waste
- invoice
- customer payment
- supplier payment
- check lifecycle
- loan installment/payment where practical
- owner transaction
- report reconciliation spot checks
- print preview/document
- create backup
- restore backup
- restart
- verify state survives
- upgrade an older DB

Automate backend portions where practical. Keep unavoidable desktop/Windows checks as an explicit manual checklist.

### 13. Data integrity test after roundtrip

Add automated backup/restore roundtrip tests that populate representative cross-domain data and confirm after restore:
- row/history preservation
- journal totals remain balanced
- inventory reconstruction unchanged
- financial balances unchanged
- invoice/payment allocations unchanged
- checks/loans/owner records preserved
- settings preserved
- managed attachment content preserved where supported

Compare authoritative values, not incidental row ordering where ordering is not part of the contract.

### 14. Failure tests

Cover at least:
- corrupt archive
- missing manifest
- checksum mismatch
- missing managed file
- invalid SQLite DB
- future unsupported schema version
- restore failure before switch leaves live data untouched
- failed reopen rolls back to original data
- backup destination write failure cleans temporary files

### 15. Security/safety boundaries

- Sanitize archive paths during extraction; reject traversal such as `../`.
- Do not restore files outside the designated temporary/application-data roots.
- Do not trust archive manifest paths blindly.
- Use restrictive/default OS permissions where practical.
- Avoid exposing arbitrary filesystem deletion through Wails APIs.

## Non-goals

Do not implement in this task:
- cloud backup/sync
- multi-user/network database
- automatic updater service
- telemetry
- online account/login
- code signing certificate procurement
- Microsoft Store publishing
- macOS/Linux production packaging beyond keeping code portable
- final deep business/UI audit fixes unrelated to backup/release unless they block release build/startup

## Acceptance criteria

- Runtime production data location is centralized and safe for Windows installs.
- A backup captures a consistent DB and all Atropaten-managed files.
- Backup packages are versioned and integrity checked.
- Restore fully validates before replacing live data.
- Failed restore cannot destroy the original live dataset.
- Backup → restore roundtrip preserves representative cross-domain state.
- Supported historical DB versions migrate to current schema successfully.
- Future unknown DB schema is rejected safely.
- Settings contains usable Backup/Verify/Restore flows.
- Windows Wails packaging configuration is release-ready in source.
- Exact Windows release-build instructions exist.
- Release smoke checklist exists.
- `go test ./...` passes.
- `cd frontend && npm run build` passes.
- `git diff --check` passes.

## Final implementation report

Report separately:
- automated validations actually run
- whether Wails CLI was available
- whether a Windows build/installer was actually produced
- whether the Windows installer was actually launched/tested
- any release step that still requires a Windows machine

Do not report platform-specific validation as passed if it was not executed.
