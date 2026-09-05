import { isValidJalaaliDate, toGregorian, toJalaali } from 'jalaali-js'

/**
 * Application dates stay calendar-neutral as ISO date strings or timestamps.
 * Jalali conversion is limited to the presentation and input boundary.
 */
export type CanonicalDate = string

export interface JalaliDate {
  year: number
  month: number
  day: number
}

const jalaliMonths = [
  'Farvardin',
  'Ordibehesht',
  'Khordad',
  'Tir',
  'Mordad',
  'Shahrivar',
  'Mehr',
  'Aban',
  'Azar',
  'Dey',
  'Bahman',
  'Esfand',
]

export const jalaliWeekdays = ['Sat', 'Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri']

const tehranDateParts = new Intl.DateTimeFormat('en-US', {
  calendar: 'gregory',
  day: 'numeric',
  hour: '2-digit',
  hour12: false,
  minute: '2-digit',
  month: 'numeric',
  timeZone: 'Asia/Tehran',
  year: 'numeric',
})

function pad(value: number): string {
  return String(value).padStart(2, '0')
}

function parseDateOnly(value: CanonicalDate): { year: number; month: number; day: number } | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return null
  return { year: Number(match[1]), month: Number(match[2]), day: Number(match[3]) }
}

function canonicalParts(value: CanonicalDate): { year: number; month: number; day: number; hour: number; minute: number } {
  const dateOnly = parseDateOnly(value)
  if (dateOnly) return { ...dateOnly, hour: 0, minute: 0 }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) throw new Error(`Invalid canonical date: ${value}`)

  const parts = Object.fromEntries(tehranDateParts.formatToParts(date).map(({ type, value: partValue }) => [type, partValue]))
  return {
    year: Number(parts.year),
    month: Number(parts.month),
    day: Number(parts.day),
    hour: Number(parts.hour),
    minute: Number(parts.minute),
  }
}

export function formatDate(value: CanonicalDate): string {
  const jalali = toJalaliDate(value)
  return `${jalali.day} ${jalaliMonths[jalali.month - 1]} ${jalali.year}`
}

export function formatDateTime(value: CanonicalDate): string {
  const { hour, minute } = canonicalParts(value)
  return `${formatDate(value)} · ${pad(hour)}:${pad(minute)}`
}

export function parseJalaliDate(value: string): CanonicalDate | null {
  const normalized = value
    .trim()
    .replace(/[۰-۹]/g, (digit) => String('۰۱۲۳۴۵۶۷۸۹'.indexOf(digit)))
  const match = /^(\d{4})[/-](\d{1,2})[/-](\d{1,2})$/.exec(normalized)
  if (!match) return null

  const jy = Number(match[1])
  const jm = Number(match[2])
  const jd = Number(match[3])
  if (!isValidJalaaliDate(jy, jm, jd)) return null

  const gregorian = toGregorian(jy, jm, jd)
  return `${gregorian.gy}-${pad(gregorian.gm)}-${pad(gregorian.gd)}`
}

export function toJalaliDate(value: CanonicalDate): JalaliDate {
  const { year, month, day } = canonicalParts(value)
  const jalali = toJalaali(year, month, day)
  return { year: jalali.jy, month: jalali.jm, day: jalali.jd }
}

export function fromJalaliDate(date: JalaliDate): CanonicalDate | null {
  if (!isValidJalaaliDate(date.year, date.month, date.day)) return null
  const gregorian = toGregorian(date.year, date.month, date.day)
  return `${gregorian.gy}-${pad(gregorian.gm)}-${pad(gregorian.gd)}`
}

export function jalaliMonthLength(year: number, month: number): number {
  if (month <= 6) return 31
  if (month <= 11) return 30
  return isValidJalaaliDate(year, month, 30) ? 30 : 29
}

export function jalaliMonthStartOffset(year: number, month: number): number {
  const canonical = fromJalaliDate({ year, month, day: 1 })
  if (!canonical) return 0
  const date = parseDateOnly(canonical)
  if (!date) return 0
  const weekday = new Date(Date.UTC(date.year, date.month - 1, date.day)).getUTCDay()
  return (weekday + 1) % 7
}

export function formatJalaliMonth(year: number, month: number): string {
  return `${jalaliMonths[month - 1]} ${year}`
}

export function currentCanonicalDate(): CanonicalDate {
  const { year, month, day } = canonicalParts(new Date().toISOString())
  return `${year}-${pad(month)}-${pad(day)}`
}
