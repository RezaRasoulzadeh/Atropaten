import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import type { ComponentForm, ParameterForm, ServiceForm } from './types'

export type TestValues = Record<string, string>
export type TestPricingLine = { name: string; detail: string; amount: number; missing: boolean }
export type TestPricingResult = {
  lines: TestPricingLine[]
  totalCostRial: number
  markupRial: number
  sellingPriceRial: number
  pricingLabel: string
  profitRial: number
  marginPercentage: number
  belowCost: boolean
  hasMissing: boolean
}

function number(value: string | undefined, fallback = 0) {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

function normalize(value: string | undefined) {
  return String(value || '').toLowerCase().replace(/[^a-z0-9]+/g, '')
}

function materialFor(component: ComponentForm, parameters: ParameterForm[], values: TestValues, materials: MaterialRecord[]) {
  if (component.type !== 'material') return null
  if (component.usageMode === 'parameter' && !component.referenceId) {
    const parameter = parameters.find((item) => item.key === component.parameterKey)
    const wanted = normalize(values[component.parameterKey] || parameter?.defaultValue)
    return materials.find((material) => [material.id, material.name, material.sku].some((value) => normalize(value) === wanted))
      || materials.find((material) => [material.name, material.sku].some((value) => normalize(value).includes(wanted) || wanted.includes(normalize(value))))
      || null
  }
  return materials.find((material) => material.id === component.referenceId) || null
}

function machineFor(component: ComponentForm, parameters: ParameterForm[], values: TestValues, machines: MachineRecord[]) {
  if (component.type !== 'machine') return null
  if (component.usageMode === 'parameter' && !component.referenceId) {
    const parameter = parameters.find((item) => item.key === component.parameterKey)
    const wanted = normalize(values[component.parameterKey] || parameter?.defaultValue)
    return machines.find((machine) => [machine.id, machine.name, machine.code].some((value) => normalize(value) === wanted))
      || machines.find((machine) => [machine.name, machine.code].some((value) => normalize(value).includes(wanted) || wanted.includes(normalize(value))))
      || null
  }
  return machines.find((machine) => machine.id === component.referenceId) || null
}

function machineRate(component: ComponentForm, machine: MachineRecord, parameters: ParameterForm[], values: TestValues) {
  const rates = machine.rates?.length ? machine.rates : [{ id: 'default', name: 'Standard', selectorValue: '', rateBasis: machine.rateBasis, rateRial: machine.rateRial, setupCostRial: machine.setupCostRial, active: true }]
  if (component.rateId) return rates.find((rate) => rate.id === component.rateId && rate.active) || null
  if (component.rateParameterKey) {
    const parameter = parameters.find((item) => item.key === component.rateParameterKey)
    const wanted = normalize(values[component.rateParameterKey] || parameter?.defaultValue)
    return rates.find((rate) => rate.active && [rate.selectorValue, rate.name, rate.id].some((value) => normalize(value) === wanted)) || null
  }
  return rates.find((rate) => rate.active) || rates[0] || null
}

function usageFor(component: ComponentForm, parameters: ParameterForm[], values: TestValues) {
  if (component.usageMode === 'parameter' && !((component.type === 'material' || component.type === 'machine') && !component.referenceId)) {
    return Math.max(0, number(values[component.parameterKey] || parameters.find((item) => item.key === component.parameterKey)?.defaultValue, number(component.usageQuantity, 1)))
  }
  return Math.max(0, number(component.usageQuantity, 1))
}

function nestedServiceCost(service: ServiceRecord, materials: MaterialRecord[], machines: MachineRecord[], services: ServiceRecord[], visited = new Set<string>()) {
  if (visited.has(service.id)) return 0
  const next = new Set(visited).add(service.id)
  const parameters = service.parameters.map((parameter) => ({ ...parameter, type: parameter.type as ParameterForm['type'], minValue: parameter.minValue ?? null, maxValue: parameter.maxValue ?? null })) as ParameterForm[]
  const values: TestValues = Object.fromEntries(parameters.map((parameter) => [parameter.key, parameter.defaultValue || '']))
  return calculateComponents(service.components as unknown as ComponentForm[], parameters, values, materials, machines, services, next).totalCostRial
}

function calculateComponents(components: ComponentForm[], parameters: ParameterForm[], values: TestValues, materials: MaterialRecord[], machines: MachineRecord[], services: ServiceRecord[], visited = new Set<string>()) {
  let running = 0
  const lines: TestPricingLine[] = []
  for (const component of components.filter((item) => item.enabled)) {
    if (component.type === 'overhead' || component.type === 'waste') {
      const amount = Math.round(running * Math.max(0, number(component.percentage)) / 100)
      running += amount
      lines.push({ name: component.name || 'Cost component', detail: `${component.percentage || 0}% of previous costs`, amount, missing: false })
      continue
    }
    const material = materialFor(component, parameters, values, materials)
    const machine = machineFor(component, parameters, values, machines)
    const service = component.type === 'service' ? services.find((item) => item.id === component.referenceId && item.active) : null
    const rate = machine ? machineRate(component, machine, parameters, values) : null
    const base = component.type === 'material' ? material ? (material.highestPurchaseUnitCostRial || material.averageUnitCostRial) : 0
      : component.type === 'machine' ? rate?.rateRial || 0
      : component.type === 'service' && service ? nestedServiceCost(service, materials, machines, services, visited) : component.rateRial
    const amount = Math.round(base * usageFor(component, parameters, values) * Math.max(0, number(component.multiplier, 1)))
    const missing = (component.type === 'material' && !material) || (component.type === 'machine' && (!machine || !rate)) || (component.type === 'service' && !service)
    const detail = component.type === 'material' ? material?.name || 'Choose a material' : component.type === 'machine' ? `${machine?.name || 'Choose a machine'}${rate ? ` · ${rate.name}` : ''}` : component.type === 'service' ? service?.name || 'Choose a service' : 'Included in estimate'
    running += amount
    lines.push({ name: component.name || 'Cost component', detail, amount, missing })
  }
  return { lines, totalCostRial: running }
}

export function calculateServiceTest(form: ServiceForm, values: TestValues, materials: MaterialRecord[], machines: MachineRecord[], services: ServiceRecord[]): TestPricingResult {
  const costs = calculateComponents(form.components, form.parameters, values, materials, machines, services)
  const rule = form.pricingRule
  const quantity = number(values[rule.parameterKey] || form.parameters.find((parameter) => parameter.key === rule.parameterKey)?.defaultValue)
  const markupRial = rule.type === 'markup' ? Math.ceil(costs.totalCostRial * number(rule.markupPercentage) / 100) : 0
  let sellingPriceRial = 0
  let pricingLabel = 'Manual price'
  if (rule.type === 'markup') { sellingPriceRial = costs.totalCostRial + markupRial; pricingLabel = `Cost + markup (${rule.markupPercentage || 0}%)` }
  else if (rule.type === 'fixed-margin') { sellingPriceRial = costs.totalCostRial + rule.fixedMarginRial; pricingLabel = 'Cost + fixed margin' }
  else if (rule.type === 'fixed') { sellingPriceRial = rule.fixedPriceRial; pricingLabel = 'Fixed price' }
  else if (rule.type === 'per-unit') { sellingPriceRial = Math.ceil(quantity * rule.perUnitRateRial); pricingLabel = 'Per-unit parameter' }
  else if (rule.type === 'quantity-tiers') {
    const tier = [...rule.tiers].sort((a, b) => Number(a.minimumQuantity) - Number(b.minimumQuantity)).filter((item) => Number(item.minimumQuantity) <= quantity).at(-1)
    sellingPriceRial = tier?.priceRial || 0
    pricingLabel = 'Quantity tiers'
  }
  const profitRial = sellingPriceRial - costs.totalCostRial
  return { ...costs, markupRial, sellingPriceRial, pricingLabel, profitRial, marginPercentage: sellingPriceRial ? profitRial / sellingPriceRial * 100 : 0, belowCost: rule.type !== 'manual' && sellingPriceRial < costs.totalCostRial, hasMissing: costs.lines.some((line) => line.missing) }
}
