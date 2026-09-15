export function isRollService(service: any): boolean {
  return ['large format', 'banners & signage'].includes(String(service?.category || '').trim().toLowerCase()) ||
    (service?.parameters || []).some((p: any) => p.active !== false && p.materialSource?.allowedKinds?.length &&
      p.materialSource.allowedKinds.every((kind: string) => ['roll-media', 'fabric'].includes(kind)))
}

// Old services did not store a layout definition. Supply the same defaults as
// the pricing backend, without requiring a visit to the Materials step.
export function ensureRollSizeInputs(service: any) {
  if (!service || !isRollService(service)) return service
  if (service.finishedSize?.quantityParameterKey) return service
  const size = { parameterKey: '', widthParameterKey: 'finished_width_mm', heightParameterKey: 'finished_height_mm',
    allowCustom: true, allowRotation: true, options: [], ...(service.finishedSize?.allowCustom ? service.finishedSize : {}), quantityParameterKey: 'layout_quantity' }
  service.parameters ||= []
  for (const [key, label, unit, value] of [
    [size.widthParameterKey, 'Finished width (mm)', 'mm', ''],
    [size.heightParameterKey, 'Custom height / length (mm)', 'mm', '1000'],
    ['layout_quantity', 'Finished pieces', 'piece', '1'],
    ['layout_gap_mm', 'Space between pieces (mm)', 'mm', '0'],
    ['layout_margin_mm', 'Edge margin (mm)', 'mm', '0'],
  ]) {
    const existing = service.parameters.find((p: any) => p.key === key)
    if (existing) { existing.required = true; continue }
    service.parameters.push({ id: service.id ? `${service.id}-${key}` : crypto.randomUUID(), key, label, unit,
      type: key === 'layout_quantity' ? 'integer' : 'decimal', required: true, active: true,
      defaultValue: value, options: [], minValue: key === 'layout_quantity' ? '1' : key.startsWith('layout_') ? '0' : '0.001', maxValue: null })
  }
  service.finishedSize = size
  return service
}

// Resolve stock from saved identities/options, never from display labels.
export function usesMaterialRollWidth(service: any): boolean {
  return Boolean(service?.finishedSize?.allowCustom && service.finishedSize.quantityParameterKey &&
    isRollService(service))
}

export function selectedRollMaterial(service: any, materials: any[], values: Record<string, string>) {
  if (!usesMaterialRollWidth(service)) return null
  const parameters = service.parameters || []
  const selected = Object.fromEntries(parameters.map((p: any) => [p.key, values[p.key] ?? p.defaultValue ?? '']))
  const ids = new Set<string>()
  for (const c of service.components || []) {
    if (c.type !== 'material' || c.enabled === false) continue
    if (c.referenceId) ids.add(c.referenceId)
    else {
      const p = parameters.find((p: any) => p.key === c.parameterKey)
      if (p?.type === 'material-reference' || p?.materialSource?.selectMaterial) {
        if (selected[p.key]) ids.add(selected[p.key])
      }
    }
  }
  const variants = (service.materialVariants || []).filter((v: any) => v.active !== false)
  if (variants.length) {
    const variant = variants.find((v: any) => Object.entries(v.values || {}).every(([key, value]) => selected[key] === value))
    if (!variant) return null
    ids.add(variant.materialId)
  }
  const candidates = materials.filter(m => m.active !== false && ['roll-media', 'fabric'].includes(m.kind) && (!ids.size || ids.has(m.id)))
    .filter(m => parameters.every((p: any) => {
      const source = p.materialSource
      if (!source) return true
      if (source.allowedKinds?.length && !source.allowedKinds.includes(m.kind)) return false
      if (source.selectMaterial) return !selected[p.key] || selected[p.key] === m.id
      const keys = source.exposedAttributeKeys?.length ? source.exposedAttributeKeys : source.exposedAttributeKey ? [source.exposedAttributeKey] : []
      if (!keys.length || !selected[p.key]) return true
      const canonical = keys.map((key: string) => {
        const a = m.attributes?.find((a: any) => a.key === key)
        if (!a) return ''
        return String(a.valueType === 'decimal' ? a.decimalValue : a.valueType === 'integer' ? a.integerValue : a.valueType === 'enum' ? a.enumCode : a.valueType === 'boolean' ? a.booleanValue : a.textValue)
      }).join('\u001f')
      return canonical === selected[p.key]
    }))
  return candidates.length === 1 ? candidates[0] : null
}

export function rollWidthValue(material: any, margin: string): string {
  const rawWidth = Number(material?.attributes?.find((a: any) => a.key === 'width_mm')?.decimalValue)
  const edge = Number(margin || '0')
  if (!Number.isFinite(rawWidth) || !Number.isFinite(edge) || edge < 0 || rawWidth <= 2 * edge) return ''
  return String(Math.round((rawWidth - 2 * edge) * 1e6) / 1e6)
}
