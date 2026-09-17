export function isRollService(service: any): boolean {
  return ['large format', 'banners & signage'].includes(String(service?.category || '').trim().toLowerCase()) ||
    (service?.parameters || []).some((p: any) => p.active !== false && p.materialSource?.allowedKinds?.length &&
      p.materialSource.allowedKinds.every((kind: string) => ['roll-media', 'fabric'].includes(kind)))
}

// Old services did not store a layout definition. Supply the same defaults as
// the pricing backend, without requiring a visit to the Materials step.
export function ensureRollSizeInputs(service: any) {
  if (!service || !isRollService(service)) return service
  if (service.finishedSize?.quantityParameterKey && service.finishedSize.allowCustom) return service
  const legacySizeKey = service.finishedSize?.allowCustom ? '' : service.finishedSize?.parameterKey || ''
  if (legacySizeKey) service.parameters = (service.parameters || []).filter((p: any) =>
    !(p.key === legacySizeKey && !p.materialSource && p.predefinedKey === 'print_size'))
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
  const canonical = (attribute: any) => {
    if (!attribute) return ''
    if (attribute.valueType === 'decimal') return String(attribute.decimalValue ?? '')
    if (attribute.valueType === 'integer') return String(attribute.integerValue ?? '')
    if (attribute.valueType === 'enum') return String(attribute.enumCode ?? '').trim()
    if (attribute.valueType === 'boolean') return attribute.booleanValue ? 'true' : 'false'
    return String(attribute.textValue ?? '').trim()
  }
  const matchesSource = (material: any, source: any) => {
    if (!source) return true
    if (source.allowedKinds?.length && !source.allowedKinds.includes(material.kind)) return false
    for (const filter of source.additionalFilters || []) {
      const attribute = material.attributes?.find((a: any) => a.key === filter.key)
      if (!attribute || attribute.valueType !== filter.value?.valueType || canonical(attribute) !== canonical(filter.value)) return false
    }
    const keys = source.exposedAttributeKeys?.length ? source.exposedAttributeKeys : source.exposedAttributeKey ? [source.exposedAttributeKey] : []
    for (const key of keys) {
      const attribute = material.attributes?.find((a: any) => a.key === key)
      if (!attribute) return false
      const allowed = (source.allowedValues || []).filter((value: any) => value.key === key)
      if (allowed.length && !allowed.some((value: any) => value.valueType === attribute.valueType && canonical(value) === canonical(attribute))) return false
    }
    return true
  }
  const candidates = materials.filter(m => m.active !== false && ['roll-media', 'fabric'].includes(m.kind) && (!ids.size || ids.has(m.id)))
    .filter(m => parameters.every((p: any) => {
      const source = p.materialSource
      if (!source) return true
      if (!matchesSource(m, source)) return false
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
  const attribute = material?.attributes?.find((a: any) => a.key === 'width_mm')
  const rawValue = attribute?.valueType === 'integer' ? attribute.integerValue : attribute?.decimalValue
  const rawWidth = Number(rawValue)
  const edge = Number(margin || '0')
  if (!Number.isFinite(rawWidth) || !Number.isFinite(edge) || edge < 0 || rawWidth <= 2 * edge) return ''
  return String(Math.round((rawWidth - 2 * edge) * 1e6) / 1e6)
}

export function rollStockWidthValue(material: any): string {
  const attribute = material?.attributes?.find((a: any) => a.key === 'width_mm')
  const rawValue = attribute?.valueType === 'integer' ? attribute.integerValue : attribute?.decimalValue
  const rawWidth = Number(rawValue)
  return Number.isFinite(rawWidth) && rawWidth > 0 ? String(rawWidth) : ''
}

export function materialOptionsForParameter(service: any, parameter: any, materials: any[], values: Record<string, string>) {
  if (!parameter?.materialSource) return []
  const materialParameters = (service.parameters || []).filter((item: any) => item.materialSource || item.type === 'material-reference')
  const canonical = (attribute: any) => {
    if (!attribute) return ''
    if (attribute.valueType === 'decimal') return String(attribute.decimalValue ?? '')
    if (attribute.valueType === 'integer') return String(attribute.integerValue ?? '')
    if (attribute.valueType === 'enum') return String(attribute.enumCode ?? '').trim()
    if (attribute.valueType === 'boolean') return attribute.booleanValue ? 'true' : 'false'
    return String(attribute.textValue ?? '').trim()
  }
  const keys = (item: any) => item.materialSource?.exposedAttributeKeys?.length ? item.materialSource.exposedAttributeKeys : item.materialSource?.exposedAttributeKey ? [item.materialSource.exposedAttributeKey] : []
  const sourceValue = (material: any, item: any) => item.type === 'material-reference' || item.materialSource?.selectMaterial
    ? material.id
    : keys(item).map((key: string) => canonical(material.attributes?.find((a: any) => a.key === key))).join('\u001f')
  const matchesSource = (material: any, item: any) => {
    const source = item.materialSource
    if (!source) return true
    if (source.allowedKinds?.length && !source.allowedKinds.includes(material.kind)) return false
    for (const filter of source.additionalFilters || []) {
      const attribute = material.attributes?.find((a: any) => a.key === filter.key)
      if (!attribute || attribute.valueType !== filter.value?.valueType || canonical(attribute) !== canonical(filter.value)) return false
    }
    for (const key of keys(item)) {
      const attribute = material.attributes?.find((a: any) => a.key === key)
      if (!attribute) return false
      const allowed = (source.allowedValues || []).filter((value: any) => value.key === key)
      if (allowed.length && !allowed.some((value: any) => value.valueType === attribute.valueType && canonical(value) === canonical(attribute))) return false
    }
    return true
  }
  const options = new Map<string, { label: string; value: string }>()
  const variants = (service.materialVariants || []).filter((variant: any) => variant.active !== false)
  for (const variant of variants) {
    const value = String(variant.values?.[parameter.key] || '').trim()
    if (!value) continue
    let material = materials.find((item: any) => item.active !== false && item.id === variant.materialId)
    if (!material || !materialParameters.every((item: any) => {
      const selected = String(values[item.key] || '').trim()
      return item.key === parameter.key || !selected || (item.type === 'material-reference' ? material.id === selected : String(variant.values?.[item.key] || '').trim() === selected)
    })) continue
    if (!matchesSource(material, parameter)) continue
    options.set(value, { value, label: material.name || value.split('\u001f').join(' × ') })
  }
  if (options.size) return Array.from(options.values()).sort((left, right) => left.label.localeCompare(right.label, undefined, { numeric: true }))
  for (const material of materials.filter((item: any) => item.active !== false)) {
    if (!materialParameters.every((item: any) => {
      if (!matchesSource(material, item)) return false
      if (item.key === parameter.key) return true
      const selected = String(values[item.key] || '').trim()
      return !selected || sourceValue(material, item) === selected
    })) continue
    const value = sourceValue(material, parameter)
    if (value && !options.has(value)) options.set(value, { value, label: material.name || value.split('\u001f').join(' × ') })
  }
  return Array.from(options.values()).sort((left, right) => left.label.localeCompare(right.label, undefined, { numeric: true }))
}
