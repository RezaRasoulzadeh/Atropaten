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
  const formatted = new Intl.NumberFormat('en-US', { maximumFractionDigits: 0 }).format(amount)
  return `${formatted} ${currencyLabels[unit]}`
}

export function formatSignedMoney(amountRial: number, unit: CurrencyUnit, sign: '+' | '−'): string {
  return `${sign}${formatMoney(amountRial, unit)}`
}
