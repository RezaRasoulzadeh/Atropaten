import { isValidJalaaliDate, toGregorian, toJalaali } from 'jalaali-js'

/**
 * Application dates stay calendar-neutral as ISO date strings or timestamps.
 * Jalali conversion is limited to the presentation and input boundary.
 */
export type CanonicalDate = string

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
  const { year, month, day } = canonicalParts(value)
  const jalali = toJalaali(year, month, day)
  return `${jalali.jd} ${jalaliMonths[jalali.jm - 1]} ${jalali.jy}`
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
