import { getLocale } from '../i18n';

export function formatLocalizedNumber(value: number, options: Intl.NumberFormatOptions = {}): string {
  return new Intl.NumberFormat(getLocale() === 'fa' ? 'fa-IR' : 'en-US', options).format(value);
}

export function normalizeDigits(value: string | number): string {
  return String(value)
    .replace(/[۰-۹]/gu, (digit) => String('۰۱۲۳۴۵۶۷۸۹'.indexOf(digit)))
    .replace(/[٠-٩]/gu, (digit) => String('٠١٢٣٤٥٦٧٨٩'.indexOf(digit)))
    .replace(/٫/gu, '.')
    .replace(/٬/gu, ',');
}

export function localizeDigits(value: string | number): string {
  return getLocale() === 'fa' ? String(value).replace(/\d/gu, (digit) => '۰۱۲۳۴۵۶۷۸۹'[Number(digit)]) : String(value);
}
