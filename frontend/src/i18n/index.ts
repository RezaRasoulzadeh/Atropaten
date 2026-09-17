import { createI18n, type MessageResolver, type PathValue } from 'vue-i18n';
import { messages, type Locale } from './messages';

export type { Locale } from './messages';

const STORAGE_KEY = 'atropaten.locale';
const defaultLocale: Locale = 'en';
const dynamicMessageTemplates = Object.keys(messages.fa)
  .filter((key) => key.includes('${'))
  .map((key) => {
    const normalizedKey = key.trim().replace(/\s+/gu, ' ');
    const parts = normalizedKey.split(/\$\{[^}]+\}/u);
    const pattern = parts
      .map((part) => part.replace(/[.*+?^${}()|[\]\\]/gu, '\\$&'))
      .join('([\\s\\S]+?)');
    return {
      key,
      matcher: new RegExp(`^${pattern}$`, 'u'),
      literalLength: parts.join('').length,
      placeholderCount: parts.length - 1,
    };
  })
  .sort((left, right) => right.literalLength - left.literalLength || right.placeholderCount - left.placeholderCount);

const resolveLocaleMessage: MessageResolver = (source, path): PathValue => {
  if (!source || typeof source !== 'object') return null;
  const message = source as Record<string, unknown>;
  if (Object.prototype.hasOwnProperty.call(message, path)) return message[path] as PathValue;
  const resolved = path.split('.').reduce<unknown>((value, segment) => {
    if (value && typeof value === 'object' && Object.prototype.hasOwnProperty.call(value, segment)) {
      return (value as Record<string, unknown>)[segment];
    }
    return undefined;
  }, message);
  return (resolved as PathValue | undefined) ?? null;
};

function isLocale(value: unknown): value is Locale {
  return value === 'en' || value === 'fa';
}

function readSavedLocale(): Locale {
  try {
    const savedLocale = window.localStorage.getItem(STORAGE_KEY);
    return isLocale(savedLocale) ? savedLocale : defaultLocale;
  } catch {
    return defaultLocale;
  }
}

function applyDocumentLocale(locale: Locale) {
  if (typeof document === 'undefined') return;
  document.documentElement.lang = locale;
  document.documentElement.dir = locale === 'fa' ? 'rtl' : 'ltr';
  document.title = locale === 'fa' ? 'آتروپاتن | مدیریت چاپخانه' : 'Atropaten | Print shop control';
}

export const i18n = createI18n({
  legacy: false,
  locale: typeof window === 'undefined' ? defaultLocale : readSavedLocale(),
  fallbackLocale: defaultLocale,
  messages,
  messageResolver: resolveLocaleMessage,
});

export function setLocale(locale: Locale) {
  i18n.global.locale.value = locale;
  applyDocumentLocale(locale);
  try {
    window.localStorage.setItem(STORAGE_KEY, locale);
  } catch {
    // Keep the in-memory language change even when browser storage is unavailable.
  }
}

export function getLocale(): Locale {
  return i18n.global.locale.value as Locale;
}

export function translateUi(value: string | null | undefined): string;
export function translateUi<T>(value: T): T;
export function translateUi(value: unknown): unknown {
  if (typeof value !== 'string' || !value) return value ?? '';
  const leading = value.match(/^\s*/u)?.[0] ?? '';
  const trailing = value.match(/\s*$/u)?.[0] ?? '';
  const sourceText = value.trim();
  if (Object.prototype.hasOwnProperty.call(messages.fa, sourceText)) {
    return `${leading}${i18n.global.t(sourceText)}${trailing}`;
  }
  if (getLocale() === 'fa') {
    const normalizedText = sourceText.replace(/\s+/gu, ' ');
    for (const { key, matcher } of dynamicMessageTemplates) {
      const match = matcher.exec(normalizedText);
      if (!match) continue;
      const parameters = Object.fromEntries(match.slice(1).map((part, index) => [`arg${index}`, part]));
      return `${leading}${i18n.global.t(key, parameters)}${trailing}`;
    }
  }
  return value;
}

if (typeof window !== 'undefined') {
  applyDocumentLocale(getLocale());
  if (typeof window.addEventListener === 'function') {
    window.addEventListener('storage', (event) => {
      if (event.key === STORAGE_KEY && isLocale(event.newValue)) {
        i18n.global.locale.value = event.newValue;
        applyDocumentLocale(event.newValue);
      }
    });
  }
}
