# Vue 3 + TypeScript + Vite

This template should help get you started developing with Vue 3 and TypeScript in Vite. The template uses Vue
3 `<script setup>` SFCs, check out
the [script setup docs](https://v3.vuejs.org/api/sfc-script-setup.html#sfc-script-setup) to learn more.

## Demo

Run these commands from the repository root. The demo dataset is isolated under `/tmp/atropaten-demo` by default.

```fish
set DEMO_ROOT /tmp/atropaten-demo
if set -q TMPDIR
    set DEMO_ROOT "$TMPDIR/atropaten-demo"
end

# Generate the fictional demo dataset.
go run ./cmd/demo-data generate \
  --root "$DEMO_ROOT" \
  --seed 6006 \
  --reference-date 2026-03-21 \
  --confirm-demo

# Launch the app against the demo dataset.
env ATROPATEN_DATA_DIR="$DEMO_ROOT" wails dev

# Reset or remove the demo dataset when finished.
go run ./cmd/demo-data reset \
  --root "$DEMO_ROOT" \
  --confirm-reset-demo
```

See [docs/DEMO_DATA.md](../docs/DEMO_DATA.md) for the safety contract and the `clean` command.

## Recommended IDE Setup

- [VS Code](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar)

## Type Support For `.vue` Imports in TS

Since TypeScript cannot handle type information for `.vue` imports, they are shimmed to be a generic Vue component type
by default. In most cases this is fine if you don't really care about component prop types outside of templates.
However, if you wish to get actual prop types in `.vue` imports (for example to get props validation when using
manual `h(...)` calls), you can enable Volar's Take Over mode by following these steps:

1. Run `Extensions: Show Built-in Extensions` from VS Code's command palette, look
   for `TypeScript and JavaScript Language Features`, then right click and select `Disable (Workspace)`. By default,
   Take Over mode will enable itself if the default TypeScript extension is disabled.
2. Reload the VS Code window by running `Developer: Reload Window` from the command palette.

You can learn more about Take Over mode [here](https://github.com/johnsoncodehk/volar/discussions/471).
