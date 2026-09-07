// Invoice DTOs expose fixed-scale units as an integer string (1,000,000 = 1).
// Format with integer arithmetic so large and fractional quantities stay exact.
export function formatQuantityUnits(value: string): string {
  if (!/^-?\d+$/.test(value)) return value
  const units = BigInt(value)
  const absolute = units < 0n ? -units : units
  const fraction = (absolute % 1000000n).toString().padStart(6, '0').replace(/0+$/, '')
  return `${units < 0n ? '-' : ''}${absolute / 1000000n}${fraction ? `.${fraction}` : ''}`
}
