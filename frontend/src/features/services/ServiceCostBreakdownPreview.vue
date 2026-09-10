<script setup lang="ts">
import { computed, watch } from 'vue'
import { CircleHelp } from 'lucide-vue-next'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import type { ComponentForm, ParameterForm, ServiceForm } from './types'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import ServiceOverviewIdentity from './ServiceOverviewIdentity.vue'
import ServiceOverviewSection from './ServiceOverviewSection.vue'

const props = defineProps<{
  components: ComponentForm[]
  parameters: ParameterForm[]
  materials: MaterialRecord[]
  machines: MachineRecord[]
  services: ServiceRecord[]
  currencyUnit: CurrencyUnit
  form: ServiceForm
  active: boolean
}>()
const emit = defineEmits<{
  'update:total': [value: number]
  'update:breakdown': [value: Array<{ name: string; amount: number; detail: string; missing: boolean }>]
}>()

type BreakdownRow = {
  component: ComponentForm
  amount: number
  detail: string
  missing: boolean
}

function numeric(value: string, fallback = 1) {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

function normalize(value: string) {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, '')
}

function materialFor(component: ComponentForm, parameters = props.parameters) {
  if (component.type !== 'material') return null
  if (component.usageMode === 'parameter') {
    const parameter = parameters.find((item) => item.key === component.parameterKey)
    if (!parameter?.defaultValue) return null
    const wanted = normalize(parameter.defaultValue)
    if (!wanted) return null
    return props.materials.find((material) => [material.id, material.name, material.sku].some((value) => normalize(value) === wanted))
      || props.materials.find((material) => [material.name, material.sku].some((value) => normalize(value).includes(wanted) || wanted.includes(normalize(value))))
      || null
  }
  return props.materials.find((material) => material.id === component.referenceId) || null
}

function machineFor(component: ComponentForm, parameters = props.parameters) {
  if (component.type !== 'machine') return null
  if (component.usageMode === 'parameter' && !component.referenceId) {
    const parameter = parameters.find((item) => item.key === component.parameterKey)
    if (!parameter?.defaultValue) return null
    const wanted = normalize(parameter.defaultValue)
    if (!wanted) return null
    return props.machines.find((machine) => [machine.id, machine.name, machine.code].some((value) => normalize(value) === wanted))
      || props.machines.find((machine) => [machine.name, machine.code].some((value) => normalize(value).includes(wanted) || wanted.includes(normalize(value))))
      || null
  }
  return props.machines.find((machine) => machine.id === component.referenceId) || null
}

function machineRateFor(component: ComponentForm, parameters = props.parameters) {
  const machine = machineFor(component, parameters)
  if (!machine) return null
  const rates = machine.rates || []
  if (component.rateId) return rates.find((rate) => rate.id === component.rateId && rate.active) || null
  if (component.rateParameterKey) {
    const parameter = parameters.find((item) => item.key === component.rateParameterKey)
    const wanted = normalize(parameter?.defaultValue || '')
    return rates.find((rate) => rate.active && [rate.selectorValue, rate.name, rate.id].some((value) => normalize(value) === wanted)) || null
  }
  return rates.find((rate) => rate.active) || (rates.length ? rates[0] : { rateRial: machine.rateRial, name: 'Standard' })
}

function usageFor(component: ComponentForm, parameters = props.parameters) {
  if (component.usageMode === 'parameter' && !((component.type === 'material' || component.type === 'machine') && !component.referenceId)) {
    const parameter = parameters.find((item) => item.key === component.parameterKey)
    return Math.max(0, numeric(parameter?.defaultValue || component.usageQuantity))
  }
  return Math.max(0, numeric(component.usageQuantity))
}

function serviceCost(service: ServiceRecord, visited = new Set<string>()) {
  if (visited.has(service.id)) return 0
  const nextVisited = new Set(visited).add(service.id)
  const parameters = service.parameters.map((parameter) => ({
    ...parameter,
    type: parameter.type as ParameterForm['type'],
    minValue: parameter.minValue ?? null,
    maxValue: parameter.maxValue ?? null,
  })) as ParameterForm[]
  let running = 0
  for (const component of service.components.filter((item) => item.enabled)) {
    if (component.type === 'overhead' || component.type === 'waste') {
      running += running * Math.max(0, numeric(component.percentage, 0)) / 100
      continue
    }
    const usage = usageFor(component as unknown as ComponentForm, parameters) * Math.max(0, numeric(component.multiplier))
    running += baseAmount(component as unknown as ComponentForm, parameters, nextVisited) * usage
  }
  return running
}

function baseAmount(component: ComponentForm, parameters = props.parameters, visited = new Set<string>()) {
  if (component.type === 'material') {
    const material = materialFor(component, parameters)
    return material ? material.highestPurchaseUnitCostRial || material.averageUnitCostRial : 0
  }
  if (component.type === 'machine') return machineRateFor(component, parameters)?.rateRial || 0
  if (component.type === 'service') {
    const service = props.services.find((item) => item.id === component.referenceId && item.active)
    return service ? serviceCost(service, visited) : 0
  }
  if (component.type === 'labor' || component.type === 'outsourced' || component.type === 'fixed' || component.type === 'manual') return Math.max(0, component.rateRial)
  return 0
}

const rows = computed<BreakdownRow[]>(() => {
  let running = 0
  return props.components.filter((component) => component.enabled).map((component) => {
    if (component.type === 'overhead' || component.type === 'waste') {
      const percentage = Math.max(0, numeric(component.percentage, 0))
      const amount = running * percentage / 100
      running += amount
      return { component, amount, detail: `${percentage}% of previous costs`, missing: false }
    }
    const amount = baseAmount(component) * usageFor(component) * Math.max(0, numeric(component.multiplier))
    running += amount
    const missing = (component.type === 'material' && !materialFor(component)) || (component.type === 'machine' && !machineFor(component)) || (component.type === 'service' && !props.services.some((service) => service.id === component.referenceId && service.active))
    const detail = component.type === 'material' ? (materialFor(component)?.name || 'Choose a material') : component.type === 'machine' ? `${machineFor(component)?.name || 'Choose a machine'}${machineRateFor(component)?.name ? ` · ${machineRateFor(component)?.name}` : ''}` : component.type === 'service' ? (props.services.find((service) => service.id === component.referenceId)?.name || 'Choose a service') : 'Included in estimate'
    const rateMissing = component.type === 'machine' && !!machineFor(component) && !machineRateFor(component)
    return { component, amount, detail, missing: missing || rateMissing }
  })
})

const subtotal = computed(() => rows.value.filter((row) => row.component.type !== 'overhead' && row.component.type !== 'waste').reduce((total, row) => total + row.amount, 0))
const total = computed(() => rows.value.reduce((sum, row) => sum + row.amount, 0))
watch(total, (value) => emit('update:total', Math.round(value)), { immediate: true })
watch(rows, (value) => emit('update:breakdown', value.map((row) => ({ name: row.component.name || 'Cost component', amount: Math.round(row.amount), detail: row.detail, missing: row.missing }))), { immediate: true })
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Cost breakdown preview">
    <ServiceOverviewIdentity :form="form" :active="active" />

    <ServiceOverviewSection title="Cost estimate" description="Current cost per service unit from the configured components.">
      <div v-if="rows.length" class="space-y-2">
        <div class="divide-y divide-base-300/70">
          <div v-for="row in rows" :key="row.component.id" class="flex min-w-0 items-center gap-2 py-2.5 first:pt-0 last:pb-0">
            <span class="size-2 shrink-0 rounded-full" :class="row.missing ? 'bg-warning' : 'bg-primary'"></span>
            <div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ row.component.name || 'Cost component' }}</strong><small class="block truncate text-xs text-base-content/55">{{ row.detail }}</small></div>
            <span class="shrink-0 text-sm tabular-nums" :class="row.missing ? 'text-warning' : ''">{{ row.missing ? 'Needs setup' : formatMoney(row.amount, currencyUnit) }}</span>
          </div>
        </div>
        <div class="flex items-center justify-between gap-3 border-t border-base-300/75 pt-2.5 text-sm"><span class="font-medium">Subtotal</span><strong class="tabular-nums">{{ formatMoney(subtotal, currencyUnit) }}</strong></div>
        <div class="flex items-center justify-between gap-3 border-t border-base-300/75 pt-2.5 text-sm"><span class="font-semibold">Estimated cost</span><strong class="text-base text-success tabular-nums">{{ formatMoney(total, currencyUnit) }}</strong></div>
      </div>
      <p v-else class="rounded-box border border-dashed border-base-300/80 p-4 text-center text-sm leading-5 text-base-content/60">Add a cost component to see its estimated breakdown.</p>
    </ServiceOverviewSection>

    <div class="flex items-start gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/65"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="15" aria-hidden="true" /><span>Material costs use the highest recorded purchase cost when available. This is a setup preview; final pricing is calculated after the service is saved.</span></div>
  </section>
</template>
