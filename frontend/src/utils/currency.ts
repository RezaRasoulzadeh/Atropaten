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
  if (unit === 'Rial') return String(amountRial)
  if (amountRial % 10 === 0) return String(amountRial / 10)
  return `${Math.trunc(amountRial / 10)}.${Math.abs(amountRial % 10)}`
}

export function parseMoneyInput(value: string, unit: CurrencyUnit): number | null {
  const normalized = value.trim().replaceAll(',', '')
  if (!/^\d+(?:\.\d+)?$/.test(normalized)) return null
  const [whole, fraction = ''] = normalized.split('.')
  if (unit === 'Rial' && fraction.length > 0) return null
  if (unit === 'Toman' && fraction.length > 1) return null
  try {
    const rial = BigInt(whole) * 10n + BigInt(fraction.padEnd(1, '0') || '0')
    const result = unit === 'Rial' ? BigInt(whole) : rial
    if (result > BigInt(Number.MAX_SAFE_INTEGER)) return null
    return Number(result)
  } catch {
    return null
  }
}
