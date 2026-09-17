import { getLocale } from '../i18n';

export function formatLocalizedNumber(value: number, options: Intl.NumberFormatOptions = {}): string {
  return new Intl.NumberFormat(getLocale() === 'fa' ? 'fa-IR' : 'en-US', options).format(value);
}

export function localizeDigits(value: string | number): string {
  return getLocale() === 'fa' ? String(value).replace(/\d/gu, (digit) => '۰۱۲۳۴۵۶۷۸۹'[Number(digit)]) : String(value);
}
