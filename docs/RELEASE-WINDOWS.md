# Windows build and data-safety contract

The application version is centralized in `internal/platform/paths.go` as
`ApplicationVersion`. Keep it aligned with the `info.productVersion` value in
`wails.json` when cutting a release. The Windows metadata template consumes
that project information for the executable and installer properties.

The current release metadata is `0.1.0` and the current database schema is
version 44. The schema migration adds service-level outsourced cost defaults
with safe zero values for existing databases, so upgrades preserve previous
orders and service data.

On a Windows build host:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
cd frontend
npm ci
cd ..
wails doctor
wails build --clean --platform windows/amd64 --nsis
```

The expected artifacts are the application executable and NSIS installer under
`build/bin/`. The installer only installs and removes files under its install
directory. It must not delete `%APPDATA%\Atropaten`, which contains the live
database, backups, attachments, and artwork. Reinstall and uninstall therefore
leave user data available for a later installation or manual restore.

Before signing off the Windows release candidate, verify the order flow for an
outsourced service: create the service with a default supplier item cost and
optional shipping, confirm those defaults are prefilled when adding an order
item, and confirm that order-specific overrides remain optional. Also verify
that in-house paper-printing and roll-material services continue through their
existing materials and machines steps.

This repository is developed on Linux; Windows compilation, WebView2 behavior,
installer execution, and print-driver behavior must be validated on Windows
before a release claim is made.

## GitHub release candidate

Pushing a version tag such as `v0.1.0` runs
`.github/workflows/windows-release.yml` on a Windows build host. The workflow
checks that the tag matches both the product version and the application
version, runs the Go and frontend checks, builds the x64 NSIS installer, and
attaches the installer and its SHA-256 checksum to a **draft** GitHub release.
The draft stays unpublished until the release smoke checklist below has been
completed on Windows 10.

After committing and pushing the release source to `main`, create and push a
tag matching `info.productVersion` in `wails.json`:

```powershell
$version = (Get-Content .\wails.json -Raw | ConvertFrom-Json).info.productVersion
$tag = "v$version"
git tag $tag
git push origin main
git push origin $tag
```

The workflow intentionally does not publish the draft automatically. Install
and test it on a clean Windows 10 machine using [the release smoke checklist](RELEASE-SMOKE.md),
then publish the draft from GitHub Releases. The installer is unsigned; signing
is not configured in this repository.

The Wails v2.15 target supports Windows 10/11 x64. Its WebView2 runtime
dependency is handled by the NSIS installer. Windows 10 itself reached the end
of standard Microsoft support on October 14, 2025; organizations still using
it should account for their applicable Extended Security Updates coverage.
