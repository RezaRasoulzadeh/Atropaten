# UI Direction

## Goal

Atropaten should feel like a modern desktop productivity/accounting application, not a marketing website wrapped in a desktop window.

The interface is Windows-first, information-dense, keyboard-friendly, and optimized for repeated daily shop operations.

## Visual language

- Primary light theme
- Neutral light/warm-gray application background
- White or subtly differentiated work surfaces
- One primary accent color
- Green/red reserved mainly for meaningful financial/status semantics
- Subtle borders and shadows
- Moderate corner radius
- Minimal gradients

## Implemented design system

M6-004 uses one semantic token layer in `frontend/src/style.css`, with the shared workspace rules in `frontend/src/views/ux-hardening.css`. Feature views may add domain-specific layout, but shared colors, spacing, control geometry, table density, and sticky surfaces should use these tokens first.

### Tokens

The light theme is built from these semantic groups:

- Surfaces: `--app-bg`, `--surface`, `--surface-raised`, `--surface-subtle`, `--surface-muted`, and `--surface-hover`.
- Structure: `--border`, `--border-subtle`, `--line`, and `--line-strong`.
- Text: `--text`, `--text-soft`, and `--text-muted`.
- Interaction: `--accent`, `--accent-strong`, `--accent-hover`, `--accent-pressed`, and `--focus-ring`.
- Meaning: `--success`, `--warning`, `--danger`, and `--info`, each with a soft background token.
- Disabled: `--disabled-bg` and `--disabled-text`.

Green and red remain reserved for meaningful status, financial direction, warnings, and destructive actions. New feature styles should not add arbitrary color values when one of these semantic tokens fits.

### Spacing and typography

The compact desktop rhythm is `--space-1` through `--space-8`: 4, 8, 12, 16, 20, 24, 28, and 32px. Use these for workspace gutters, panel padding, form rows, table cells, tabs, toolbars, and inspector sections. The workspace gutter is responsive within a deliberate 20–40px range so narrow laptop windows do not waste horizontal space.

The UI uses Nunito with a 14px body size. Metadata is 12px, labels are 11px, table text is 12px, section headings are 14px, and page titles are capped at 24px. This preserves a dense ERP/productivity hierarchy rather than introducing marketing-style headings.

### Controls, buttons, badges, and tables

- Standard inputs, selects, date controls, and normal buttons are 36px high (`--control-height`); compact controls are 32px.
- Controls use a flat surface with a subtle border and restrained focus ring. Recessed/inset treatment is not part of the visual language.
- Icon buttons use a 34px square hit area and must have an accessible label when they have no visible text.
- Primary, secondary, ghost/text, and danger actions use the shared `.button-*` variants. Destructive actions use `.button-danger`; archive and delete remain visually and semantically distinct.
- `StatusBadge` uses the same pill geometry and semantic tones across orders, inventory, production, finance, and purchasing.
- Operational table headers are 36px and normal rows are 48px. Tables use compact 12px text, consistent cell padding, tabular numeric alignment, and deliberate horizontal scrolling only inside `.table-wrap` for genuinely wide registers.

### Panels, inspectors, forms, and workspace geometry

`SectionPanel` is the shared bordered surface. Panel headers use a 60px minimum height and 12/16px padding. Register + inspector screens use `minmax(0, 1fr) var(--inspector-width)`, where the inspector is explicitly sized with `--inspector-width`; both grid children have `min-width: 0` so controls cannot escape their panel. At narrower desktop widths the two columns become one column before clipping begins.

Forms use full-width, border-box controls, consistent label/control gaps, and two-column grids that collapse at the shared desktop threshold. Ordinary forms and inspectors do not create their own horizontal scroll containers. Long table data may scroll inside its table shell, while long labels and values wrap or truncate intentionally.

All major views use the same page-header pattern. `WorkspaceStickyStack` is sticky relative to the workspace scroll container below the global top bar, with the shared background, border, elevation, and z-index. Tabs use the shared `.workspace-tabs`/`.workspace-tab` geometry and remain horizontally scrollable when a desktop window is narrow. `WorkspaceBottomActions` sits above the status strip with the same panel treatment. Feature components should not introduce independent fixed offsets or z-index values.

### Responsive desktop acceptance widths

The target checks are representative desktop widths, not mobile layouts:

- Narrow laptop: approximately 1024px viewport width, including collapsed-sidebar operation.
- Normal desktop: approximately 1280–1440px, with register panes consuming remaining width beside explicit inspectors.
- Wide desktop: approximately 1600px and above, with workspace gutters capped rather than leaving a floating narrow column.

At the narrow breakpoint, filter rows wrap, two-column forms collapse, and register/inspector layouts stack before fields clip. Mobile redesign remains out of scope.

Search/filter bars use the same explicit toolbar contract across registers: the search field flexes into remaining width, labeled filters keep a compact secondary label above the control, and the result count remains aligned with the controls. Toolbar tabs used as filters are compact segmented choices and do not inherit the full page-tab strip geometry.
- Approximately 14px normal UI text
- Compact table rows, approximately 36-42px
- Strong typographic hierarchy without oversized headings

## Application shell

Primary navigation:

```text
Dashboard
Orders
Production
Customers
Services
Materials
Purchases
Suppliers
Accounting
Checks
Loans
Owners
Reports
Settings
```

The left sidebar is collapsible. A top bar provides global search and application/shop controls. A compact bottom status area may expose high-value financial balances such as cash, bank, receivable, and payable.

Frequently used operational areas appear before finance and administration.

## Global workspace and sticky-surface rules

All major pages must use the same workspace structure rather than inventing independent page shells.

The application shell owns the viewport. The global sidebar, top bar, and bottom status strip remain visually stable. The page workspace between them owns normal vertical scrolling.

Every substantial page should use a shared page-header/action pattern positioned directly below the global top bar. This region should consume the otherwise unused whitespace at the top of the workspace and contain the page identity and high-value actions instead of leaving a large decorative gap before content.

The standard page header should normally contain, as applicable:

- breadcrumb/back navigation or section context
- page title and compact supporting metadata/status
- primary action
- secondary actions or overflow menu
- page-level filters or controls when they are important enough to remain accessible

When the workspace scrolls, high-value page actions should remain available using the shared sticky page-header/action surface. Sticky elements must be positioned relative to the correct scroll container and offset below the global top bar; they must not cover content, tabs, dropdowns, or focus targets.

Bottom actions that are operationally important throughout a page, such as Dashboard Quick Actions, should use the shared persistent-bottom action pattern above the global bottom status strip rather than appearing as ordinary content at the end of a long page.

Use these patterns consistently across Dashboard, Orders, Production, Customers, Services, Materials, Purchases, Suppliers, Accounting, Checks, Loans, Owners, Reports, and Settings wherever the page needs persistent controls.

Do not create multiple visually different sticky bars for different features. Shared sticky surfaces should have consistent background, border, shadow/elevation, spacing, z-index behavior, and transition behavior. Prefer subtle separation from content rather than floating-card styling.

Avoid:

- large unused whitespace between the global top bar and page header/content
- `position: fixed` inside feature components when a sticky shared surface is sufficient
- unrelated page-specific sticky offsets or z-index values
- nested vertical scroll containers without a specific functional reason
- content hidden underneath sticky headers or the bottom status strip
- action bars that disappear merely because the user scrolled a long table/form

Tables, inspectors, tabs, and forms should participate in the same workspace geometry. A page with tabs may keep the page header and, where useful, the tab strip sticky as a coordinated stack rather than implementing unrelated sticky positions.

## Workspace fit, width, and overflow rules

Desktop pages must use the available workspace width deliberately. A large unused right/left region is a layout defect unless that whitespace is part of a specific reading-width decision.

For register + inspector/editor layouts:

- the register/table pane should flex to consume remaining width
- the inspector must have an explicit sensible width/min-width/max-width
- neither pane may force controls outside its own box
- grid/flex children that contain forms/tables must use `min-width: 0` where required so they can actually shrink
- inputs, selects, textareas, and grouped field rows must fit the inspector width and use `box-sizing: border-box`
- two-column form rows must collapse or rebalance before fields clip or overflow
- long labels, values, badges, and table content must wrap, truncate, or scroll intentionally; accidental clipping is not acceptable
- horizontal scrolling is acceptable for genuinely wide data tables, but ordinary forms/inspectors should not require it
- the workspace should not leave a narrow content column floating inside a much wider desktop viewport without a deliberate max-width reason

Every major page must be checked at representative desktop window widths, including narrower laptop-like sizes and wide desktop sizes. Layout acceptance is not only “build passes”; the page must use width coherently and no interactive control may be clipped, hidden under another surface, or rendered outside its panel.

The screenshot pattern to avoid is a register that stops early while unused workspace remains beside it, combined with an inspector whose form controls extend beyond and get clipped by the panel/window boundary.

## Interaction patterns

Prefer:

- Dense sortable/filterable tables
- Resizable columns
- Persistent table/column configuration
- Right-side inspectors for quick detail/edit operations
- Tabs within complex workspaces
- Context menus
- Keyboard shortcuts
- Inline editing where safe
- Explicit confirmation for destructive or financially significant operations

Avoid excessive modal dialogs and navigation to separate pages for simple inspections.

Deletion must be a distinct destructive action from Archive. When a record can be safely hard-deleted, expose Delete with explicit confirmation. If deletion is blocked because protected history references the record, explain why rather than silently archiving it.

## Dashboard

The dashboard answers: **What requires attention now?**

Primary content:

- Sales
- Profit
- Receivables
- Payables
- Production queue/status
- Late/promised orders
- Checks due/overdue
- Loan installments due
- Overdue customer invoices
- Low inventory
- Recent transactions

Charts are secondary to actionable operational information.

## Orders list

A dense table with:

- Order number
- Customer
- Item summary
- Total
- Remaining balance
- Promised date
- Commercial status
- Fulfillment status

Provide search and filters for status, customer, and date. Opening an order enters the order workspace.

## Order workspace

Header contains order number, customer, status, promised date, and primary actions.

Tabs:

- Overview
- Items
- Production
- Payments
- Files
- History

The workspace shows total, paid, and remaining balance. Authorized users can see estimated/actual cost, profit, and margin. Cost/profit information should be hideable quickly when a customer can see the screen.

## Service configurator

This is one of the most important interfaces in the product.

Layout should combine dynamic service parameters with a live calculation panel.

Example parameters:

- Quantity
- Paper/material
- Size
- Color mode
- Single/double sided
- Lamination
- Cutting/finishing
- Design source

Live calculation explains:

- Material cost
- Printing/machine cost
- Finishing cost
- Labor
- Waste
- Other/outsourced cost
- Estimated total cost
- Suggested selling price
- Editable selling price
- Profit
- Margin

Changing a parameter updates the calculation immediately. Below-cost overrides should produce a clear warning without unnecessarily preventing the operator from proceeding.

## Materials

Use the reusable **table + inspector** pattern.

Table columns initially include material, physical stock, reserved, available, average cost, and low-stock status.

Selecting a row opens an inspector showing current quantities, average cost, recent movements, supplier information, and actions such as Purchase and Adjust.

## Production

Production should prioritize work sequencing and promised delivery. Operators need to see what is pending, ready, in progress, outsourced, late, and completed without opening every order.

## Accounting

Workspace tabs:

- Overview
- Accounts
- Transactions
- Journal
- Expenses

Overview emphasizes cash, bank, receivables, payables, and current-period result.

The account view uses a hierarchical chart-of-accounts presentation with balances. Journal entry UI is primarily for inspection and advanced corrections because normal workflows post automatically.

## Checks

Use a due-date-oriented interface grouped by urgency, for example:

- Overdue
- This week
- Later

Allow Incoming, Outgoing, and All views. Show party, amount, bank/reference, due date, and current lifecycle status prominently.

## Owners

Show each owner's ownership share, profit share, and capital position.

Provide focused views/actions for:

- Capital
- Drawings
- Loans
- Profit distribution
- Statement/history

## Responsive expectations

This is desktop software. Optimize first for common laptop/desktop window sizes rather than mobile breakpoints. Narrow windows should degrade gracefully, but mobile UI is not an initial product target.
