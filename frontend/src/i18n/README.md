# UI translations

The interface uses English (`en`) and Persian (`fa`). Keep application labels in `messages.ts` and the Persian string catalogs, then render template copy through `$t()` or `$ui()` so locale changes also update dynamic labels and fallback text.

The Persian wording follows established Persian software localization for common interface patterns. GNOME Persian translations and the Persian Localization Lab glossary are useful references:

- https://l10n.gnome.org/languages/fa/gnome-46/ui/
- https://wiki.localizationlab.org/index.php/Persian_Localization_Lab_Unified_Glossary

Use this project glossary consistently:

- `item` → `آیتم` for order, service, and purchase entries.
- `line item` → `آیتم فاکتور`.
- `service` → `خدمت`.
- `material` → `مادهٔ اولیه` (plural: `مواد اولیه`).
- `waste` → `پرت` in print and production flows.
- `supplier` → `تأمین‌کننده`.
- `save` → `ذخیره`; `post` → `ثبت` for accounting or inventory actions.
- `pre-invoice` → `پیش‌فاکتور`.

Use Persian punctuation and half-spaces where they make the label easier to read. Keep customer, supplier, service, and material names, identifiers, and other user-entered values unchanged.
