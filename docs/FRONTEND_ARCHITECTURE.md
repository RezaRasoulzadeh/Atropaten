# Frontend architecture

The target is a desktop ERP built on Vue 3, Tailwind 4 and DaisyUI 5. Go services own domain calculations, validation and posting. The generated Wails bridge remains unchanged.

## Boundaries

- `src/app`: application shell and navigation.
- `src/components/ui`: typed controls, forms, tables, registers and feedback surfaces.
- `src/components/layout`: workspace headers, tabs, panels and inspector geometry.
- `src/features/<domain>`: workspace views and feature-owned editors/panels.
- `src/composables`: presentation lifecycle and user-action state only.
- `src/api`: existing typed Wails service adapters; no additional service layer.
- `src/ui/feedback.ts`: centralized notifications, normalization and confirmation.
- `src/utils`: existing Jalali date, fixed quantity and Rial/Toman presentation.

No empty directories or generic page schema. Features compose small primitives. Tables retain semantic table elements and explicit numeric columns; entity registers use keyboard-accessible row activation. Inspector content belongs to the feature, while its header, sections and width belong to shared layout.

## Refactoring sequence

Audit all contexts before changing them. First establish controls and geometry; then migrate in batches: Customers/Materials/Orders; Quotes/Production; Suppliers/Machines/Services; Purchases/Invoices; Accounting/Expenses/Treasury; Checks/Loans/Owners; Reports/Settings/print surfaces. Preserve existing unsaved frontend edits and review them as part of the migration.

The browser verification bridge is development-only, uses a newly generated marked M6-006 root, and is excluded from ordinary Go builds. It must never connect to production application data. Browser rendering is not native Wails or WebView2 validation.
