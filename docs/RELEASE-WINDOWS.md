# Windows build and data-safety contract

The application version is centralized in `internal/platform/paths.go` as
`ApplicationVersion`. Keep it aligned with the `info.productVersion` value in
`wails.json` when cutting a release. The Windows metadata template consumes
that project information for the executable and installer properties.

On a Windows build host:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
cd frontend
npm ci
cd ..
wails doctor
wails build --clean --target windows/amd64 --nsis
```

The expected artifacts are the application executable and NSIS installer under
`build/bin/`. The installer only installs and removes files under its install
directory. It must not delete `%APPDATA%\Atropaten`, which contains the live
database, backups, attachments, and artwork. Reinstall and uninstall therefore
leave user data available for a later installation or manual restore.

This repository is developed on Linux; Windows compilation, WebView2 behavior,
installer execution, and print-driver behavior must be validated on Windows
before a release claim is made.
