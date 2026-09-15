type MaterialVariantLike = {
  active?: boolean
  materialId: string
  values: Record<string, string>
}

function sameValue(left: unknown, right: unknown) {
  return String(left ?? '').trim() === String(right ?? '').trim()
}

/** Resolve a mapped material without assuming every variant has every input. */
export function findMaterialVariant<T extends MaterialVariantLike>(variants: T[] | undefined, selected: Record<string, string>) {
  const active = (variants || []).filter((variant) => variant.active !== false && Object.keys(variant.values || {}).length > 0)
  const matches = active.filter((variant) => Object.entries(variant.values || {}).every(([key, value]) => sameValue(selected[key], value)))
  if (!matches.length) return null

  const exact = matches.find((variant) => Object.keys(variant.values || {}).length === Object.keys(selected).length)
  return exact || (matches.length === 1 ? matches[0] : null)
}

export function mappedMaterialOptionValues<T extends MaterialVariantLike>(variants: T[] | undefined, parameterKey: string) {
  return new Set(
    (variants || [])
      .filter((variant) => variant.active !== false)
      .map((variant) => variant.values?.[parameterKey] || '')
      .filter(Boolean),
  )
}
