# Atropaten release smoke checklist

Run this checklist on a clean Windows machine and once on an upgrade install.

- Install the NSIS artifact; verify the shortcut, product name/version, and that the data directory is outside the install directory.
- Create a customer, service, material, supplier, purchase, order, quote, invoice, payment, check, loan, owner, and fiscal period record.
- Configure an order, post production consumption/waste, post the invoice, allocate a partial payment, and confirm derived balances.
- Open Reports and Settings; confirm dashboard/report totals, Jalali dates, grouped Rial/Toman values, and no clipped desktop forms.
- Add an attachment/artwork file, create a backup, verify it, and record the archive path.
- Change a setting and add a transaction; restore the backup after explicit confirmation and verify the old settings, files, inventory, journals, invoices, payments, checks, loans, owners, and allocations return unchanged.
- Print a quote, invoice, payment receipt, customer statement, and supplier statement through the dedicated print layout; confirm A4 page breaks and hidden application chrome.
- Close and relaunch the app; confirm restored data is still present and the database is not beside the executable.
- Uninstall and reinstall; verify `%APPDATA%\Atropaten` and its backups/attachments remain intact.
- On upgrades, confirm the migration completes and the same cross-domain records remain usable.
