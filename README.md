# Atropaten

Atropaten is a local desktop management application for small print shops. It combines sales, production, inventory, pricing, accounting, customer/supplier management, checks, loans, owner accounting, and reporting in one Windows-first application.

## Stack

- Go
- Wails 2
- Vue 3
- TypeScript
- Vite
- SQLite

## Product principles

- Local-first desktop application; no SaaS or multi-tenant assumptions.
- Windows is the primary deployment target.
- Fast operational workflows are more important than decorative UI.
- Accounting is double-entry internally, while routine operations post journal entries automatically.
- Historical order, cost, inventory, and accounting records are immutable or corrected through compensating records.
- Pricing is rule-based and composable instead of hardcoded per service.
- Estimated cost, actual cost, and selling price remain distinct.
- Owners/partners are first-class accounting entities.

## Project documents

- [Product](docs/PRODUCT.md)
- [Domain](docs/DOMAIN.md)
- [Architecture](docs/ARCHITECTURE.md)
- [UI](docs/UI.md)
- [Roadmap](ROADMAP.md)
- [Milestones](MILESTONES.md)
- [Contributing](CONTRIBUTING.md)

## Development

```sh
wails dev
```

## Build

```sh
wails build
```

## Status

Initial product and architecture definition. Implementation begins with the desktop shell and core operational workflows before the final persistence schema is locked down.
