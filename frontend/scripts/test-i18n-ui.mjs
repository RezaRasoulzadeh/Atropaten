import assert from 'node:assert/strict'
import { createRequire } from 'node:module'
import { fileURLToPath } from 'node:url'
import { build } from 'esbuild'

const root = fileURLToPath(new URL('../', import.meta.url))
globalThis.window = {
  localStorage: { getItem: () => null, setItem: () => {} },
  addEventListener: () => {},
}

const result = await build({
  stdin: {
    resolveDir: root,
    loader: 'ts',
    contents: `
      import { setLocale, translateUi } from './src/i18n'
      export { setLocale, translateUi }
    `,
  },
  bundle: true,
  write: false,
  format: 'cjs',
  platform: 'node',
  external: ['vue', 'vue-i18n'],
})
const built = { exports: {} }
new Function('require', 'module', 'exports', result.outputFiles[0].text)(
  createRequire(import.meta.url),
  built,
  built.exports,
)

const { setLocale, translateUi } = built.exports
setLocale('fa')

const cases = [
  ['Last 30 days', '۳۰ روز گذشته'],
  ['Last 90 days', '۹۰ روز گذشته'],
  ['Last 365 days', '۳۶۵ روز گذشته'],
  ['Outstanding', 'معوق'],
  ['0 open invoices', '0 فاکتور تسویه‌نشده'],
  ['1 open invoice', '1 فاکتور تسویه‌نشده'],
  ['2 open invoices', '2 فاکتور تسویه‌نشده'],
  ['1 material option', '1 گزینهٔ مادهٔ اولیه'],
  ['2 material options', '2 گزینهٔ مادهٔ اولیه'],
  ['1 machine available', '1 دستگاه در دسترس'],
  ['1 machine rate available', '1 نرخ دستگاه در دسترس'],
  ['2 active orders · Current status', '2 سفارش فعال · وضعیت فعلی'],
  ['Open order ORD-1001', 'باز کردن سفارش ORD-1001'],
  ['Remove artwork.png', 'حذف artwork.png'],
  ['Preview artwork.png', 'پیش‌نمایش artwork.png'],
  ['3 line items', '3 آیتم فاکتور'],
  ['4 scheduled installments', '4 قسط زمان‌بندی‌شده'],
  ['Select cash or bank account', 'انتخاب حساب نقدی یا بانکی'],
  ['Always use one material', 'همیشه از یک مادهٔ اولیه استفاده بشه'],
  ['Configured usage / finished piece', 'مصرف تنظیم‌شده برای هر قطعهٔ نهایی'],
  ['Movements', 'گردش‌ها'],
  ['Data location', 'محل ذخیرهٔ داده‌ها'],
  ['Database', 'پایگاه داده'],
  ['Version / schema', 'نسخه / ساختار پایگاه داده'],
  ['Backups folder', 'پوشهٔ پشتیبان‌گیری'],
  ['No default choice', 'بدون گزینهٔ پیش‌فرض'],
  ['Select machine', 'انتخاب دستگاه'],
  ['Ink', 'جوهر'],
  ['Chemical', 'مواد شیمیایی'],
  ['free', 'آزاد'],
  ['On receipt', 'هنگام دریافت'],
  ['Cost per', 'هزینهٔ هر'],
  ['Accounts', 'حساب‌ها'],
  ['Transfers / Treasury', 'انتقال‌ها / خزانه‌داری'],
  ['Profit Allocation', 'تقسیم سود'],
  ['Payments', 'پرداخت‌ها'],
  ['posted', 'ثبت‌شده'],
  ['reversed', 'برگشت‌خورده'],
  ['Posted purchase PUR-1001', 'خرید ثبت‌شده PUR-1001'],
  ['Posted invoice INV-1001', 'فاکتور ثبت‌شده INV-1001'],
  ['receive check CHK-1001', 'دریافت چک CHK-1001'],
  ['invoice cogs', 'بهای تمام‌شدهٔ فاکتور'],
  ['payment allocation adjustment', 'اصلاح تخصیص پرداخت'],
]

for (const [source, expected] of cases) {
  assert.equal(translateUi(source), expected, `Incorrect Persian translation for: ${source}`)
}
assert.equal(
  translateUi(`${translateUi('Confirmed')}: 1 order`),
  'تأییدشده: 1 سفارش',
  'Dashboard chart labels must translate their status and singular order count',
)
assert.equal(
  translateUi(`${translateUi('Confirmed')}: 2 orders`),
  'تأییدشده: 2 سفارش',
  'Dashboard chart labels must translate plural order counts',
)
assert.equal(
  translateUi(`Select ${translateUi('Machine'.toLowerCase())}`),
  'انتخاب دستگاه',
  'Dynamic option prompts must translate their field labels',
)
assert.equal(
  translateUi('Roll printer — apply selected rate per'),
  'Roll printer — محاسبهٔ نرخ انتخاب‌شده به‌ازای',
  'Machine rate basis labels must translate their fixed phrase',
)

setLocale('en')
assert.equal(translateUi('Last 30 days'), 'Last 30 days', 'English labels must stay in English')
console.log(`UI translation smoke tests passed (${cases.length} Persian cases and English fallback).`)
