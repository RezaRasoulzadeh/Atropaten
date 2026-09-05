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
