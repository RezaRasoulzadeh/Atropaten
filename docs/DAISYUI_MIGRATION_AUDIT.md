# DaisyUI migration audit

## Scope

M6-007 introduced Tailwind CSS 4, DaisyUI 5, the named `atropaten` light theme, reusable Vue primitives, central error normalization, a global toast host, and promise-based confirmation. The pass intentionally preserved Go/Wails authority and domain behavior.

## Shared-system findings and fixes

| Area | Root cause | M6-007 result |
| --- | --- | --- |
| Inputs/selects/textareas | M6-004 broad workspace selectors competed with feature CSS and DaisyUI | Removed the old `ux-hardening.css` control layer; shared `at-control` is flat, border-only, and shadow-free |
| Search/filter bars | Multiple feature-specific flex geometries and label placements | One shared toolbar geometry; Customers and Suppliers use `SearchFilterBar`, and the same contract covers all register class variants |
| Buttons | Raw button classes mixed icon spacing and disabled states | Added `AppButton`/`IconButton` variants and aligned icon/text content |
| Empty states | Feature-specific large placeholder treatment | Added compact `EmptyState`, `LoadingState`, and `InlineAlert` primitives |
| Destructive operations | Browser confirms and ad hoc strings | Added central `ConfirmDialog`/`confirmAction`; migrated major invoice, purchase, customer, supplier, owner, order-invoice, and restore actions |
| Feedback | App-local single message with no severity and scattered raw catches | Added `useToast`, deduplication, severity timeouts, live-region host, and `normalizeError` |

## Major-workspace audit

| Workspace | Toolbar / header | Table / inspector | Feedback / confirmation | Notes |
| --- | --- | --- | --- | --- |
| Dashboard | Existing shared shell header | KPI/panel surfaces retained | shell toast path available | Attention cards remain domain-driven |
| Orders | Existing order filter geometry uses shared toolbar contract | register + order workspace retained | order workspace emits through shell | Pricing/production authority unchanged |
| Quotes/configurators | Existing workspace/header | configurator panels retained | shell feedback path available | Parameters and cost sections retain domain behavior |
| Production | Shared geometry-only toolbar rules remove the broken toolbar layout | production inspectors remain width-constrained | shell feedback path available | No production semantics changed |
| Customers | Migrated to `SearchFilterBar` and `AppButton` | list + inspector retained | shared delete confirmation | Labels are secondary |
| Services | Shared toolbar class contract | configurator/inspector retained | shell feedback path available | Existing pricing calculations untouched |
| Materials | Shared toolbar class contract | table + stock inspector retained | shell feedback path available | Inventory ledger untouched |
| Purchases | Shared header and toolbar geometry | register + editor/inspector retained | post/cancel/delete use shared confirmation | Posting remains Go-authoritative |
| Suppliers | Migrated to `SearchFilterBar` and `AppButton` | register + inspector retained | shared delete confirmation | Archive and delete remain distinct |
| Accounting | Shared panel/table geometry | payment and accounting tables retained | inline errors plus shell path | Journal authority unchanged |
| Invoices | Shared toolbar geometry | register + invoice inspector retained | void/delete use shared confirmation | Snapshot/accounting semantics unchanged |
| Expenses | Shared panel/table geometry | accounting extras retained | shell feedback path available | Posting/reversal remains Go-authoritative |
| Treasury/payments | Shared panel/table geometry | payment panels retained | shell feedback path available | Allocations unchanged |
| Checks | Shared panel/table geometry | register/history inspector retained | shell feedback path available | State transitions unchanged |
| Loans | Shared panel/table geometry | loan/installment inspector retained | shell feedback path available | Loan journal authority unchanged |
| Owners | Shared tabs/header geometry | owner/period inspectors retained | delete/close use shared confirmation | Basis-point semantics unchanged |
| Reports | Shared page/sticky geometry | report tables/print preview retained | inline + shell feedback path | Print entry remains available |
| Settings | Shared page/panel geometry | settings/backup panels retained | restore uses shared confirmation | Backup/restore Wails API unchanged |
| Print preview | Existing dedicated print surface | preview remains isolated for print | host remains outside print surface | No native print claims made |

## Visual verification

The frontend production build was inspected through its generated CSS and representative narrow/wide desktop layout rules. A Wails CLI/native Windows WebView2 runtime is not available in this Linux workspace, so native WebView rendering, Windows font rasterization, and packaged print dialogs were not validated. No claim of native validation is made.

## Follow-up

Some lower-level feature forms still contain legacy markup while they are progressively routed through the shared geometry contract. Their domain behavior is intentionally unchanged. Future UI work should replace those raw controls with `AppInput`, `AppSelect`, `AppTextarea`, `FormField`, and `SearchFilterBar` rather than adding selectors to global CSS.
