/**
 * Canonical money values are always stored and calculated as Iranian Rial.
 * Toman is a presentation/input preference only: 1 toman = 10 rial.
 */
export type CurrencyUnit = 'Toman' | 'Rial'

const currencyLabels: Record<CurrencyUnit, string> = {
  Toman: 'IRT',
  Rial: 'IRR',
}

export function convertRial(amountRial: number, unit: CurrencyUnit): number {
  return unit === 'Toman' ? amountRial / 10 : amountRial
}

export function formatMoney(amountRial: number, unit: CurrencyUnit): string {
  const amount = convertRial(amountRial, unit)
  const hasFractionalToman = unit === 'Toman' && amountRial % 10 !== 0
  const formatted = new Intl.NumberFormat('en-US', { maximumFractionDigits: hasFractionalToman ? 1 : 0 }).format(amount)
  return `${formatted} ${currencyLabels[unit]}`
}

export function formatSignedMoney(amountRial: number, unit: CurrencyUnit, sign: '+' | '−'): string {
  return `${sign}${formatMoney(amountRial, unit)}`
}

export function formatMoneyInput(amountRial: number, unit: CurrencyUnit): string {
  const sign = amountRial < 0 ? '-' : ''
  const absolute = Math.abs(amountRial)
  if (unit === 'Rial') return `${sign}${groupInteger(String(absolute))}`
  if (absolute % 10 === 0) return `${sign}${groupInteger(String(absolute / 10))}`
  return `${sign}${groupInteger(String(Math.trunc(absolute / 10)))}.${absolute % 10}`
}

export function parseMoneyInput(value: string, unit: CurrencyUnit): number | null {
  const normalized = value.trim().replaceAll(',', '')
  if (!/^-?\d+(?:\.\d+)?$/.test(normalized)) return null
  const negative = normalized.startsWith('-')
  const unsigned = negative ? normalized.slice(1) : normalized
  const [whole, fraction = ''] = unsigned.split('.')
  if (unit === 'Rial' && fraction.length > 0) return null
  if (unit === 'Toman' && fraction.length > 1) return null
  try {
    const rial = BigInt(whole) * 10n + BigInt(fraction.padEnd(1, '0') || '0')
    const unsignedResult = unit === 'Rial' ? BigInt(whole) : rial
    if (unsignedResult > BigInt(Number.MAX_SAFE_INTEGER)) return null
    const result = negative ? -unsignedResult : unsignedResult
    return Number(result)
  } catch {
    return null
  }
}

/** Groups an already validated integer portion without changing its value. */
export function groupInteger(value: string): string {
  const sign = value.startsWith('-') ? '-' : ''
  const digits = sign ? value.slice(1) : value
  return `${sign}${digits.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}`
}
