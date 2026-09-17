<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue'
import { pricingApi, type PricingRecord } from '../../api/pricing'
import { CircleHelp } from 'lucide-vue-next'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import type { ComponentForm, ParameterForm, ServiceForm } from './types'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import ServiceOverviewIdentity from './ServiceOverviewIdentity.vue'
import ServiceOverviewSection from './ServiceOverviewSection.vue'
import { collapseGroupedMaterialComponents } from './serviceComponentSync'
import { serviceCategoryRequirements } from './serviceCategory'
import { findMaterialVariant } from './serviceMaterialResolution'
import { ensureRollSizeInputs, rollWidthValue, selectedRollMaterial, usesMaterialRollWidth } from './rollSizeInputs'

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
		if (parameter?.type === 'material-reference' || parameter?.materialSource?.selectMaterial) {
			return props.materials.find((material) => material.id === parameter.defaultValue && material.active) || null
		}
		if (parameter?.materialSource) {
			const groups = parameters.filter((item) => item.type === 'choice' && item.materialSource && !item.materialSource.selectMaterial)
			const values = Object.fromEntries(groups.map((group) => [group.key, group.defaultValue]))
			if (groups.some((group) => !group.defaultValue)) return null
			const variant = findMaterialVariant(props.form.materialVariants, values)
			return props.materials.find((material) => material.id === variant?.materialId && material.active) || null
		}
		return null
  }
  return props.materials.find((material) => material.id === component.referenceId) || null
}

function machineFor(component: ComponentForm, parameters = props.parameters) {
  if (component.type !== 'machine') return null
  if (component.usageMode === 'parameter' && !component.referenceId) {
    const parameter = parameters.find((item) => item.key === component.parameterKey)
    if (!parameter?.defaultValue) return null
    return props.machines.find((machine) => machine.id === parameter.defaultValue && machine.active) || null
  }
  return props.machines.find((machine) => machine.id === component.referenceId && machine.active) || null
}

function machineRateFor(component: ComponentForm, parameters = props.parameters) {
  const machine = machineFor(component, parameters)
  if (!machine) return null
  const rates = machine.rates?.length ? machine.rates : [{ id: 'default', name: 'Standard', selectorValue: '', selectorPredefinedKey: '', rateRial: machine.rateRial, active: true }]
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
  for (const component of collapseGroupedMaterialComponents(service.components as unknown as ComponentForm[], parameters).filter((item) => item.enabled)) {
    if (component.type === 'overhead' || component.type === 'waste') {
      running += running * Math.max(0, numeric(component.percentage, 0)) / 100
      continue
    }
    const usage = usageFor(component as unknown as ComponentForm, parameters) * Math.max(0, numeric(component.multiplier))
    running += baseAmount(component as unknown as ComponentForm, parameters, nextVisited) * usage
  }
  return running
}

const effectiveComponents = computed(() => collapseGroupedMaterialComponents(props.components, props.parameters))
const categoryRequirements = computed(() => serviceCategoryRequirements(props.form.category))

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

const layoutPrice = ref<PricingRecord | null>(null)
const layoutMessage = ref('')
let previewToken = 0
let previewTimer: ReturnType<typeof setTimeout> | undefined
onBeforeUnmount(() => { previewToken++; clearTimeout(previewTimer) })
watch(() => props.form, () => {
  const token = ++previewToken
  clearTimeout(previewTimer)
  layoutPrice.value = null
  if (!props.form.finishedSize?.quantityParameterKey) return
  ensureRollSizeInputs(props.form)
  layoutMessage.value = 'Calculating batch layout…'
  previewTimer = setTimeout(async () => {
    const values = Object.fromEntries(props.form.parameters.map(p => [p.key, p.defaultValue]))
    if (usesMaterialRollWidth(props.form)) {
      const materialComponent = effectiveComponents.value.find((component) => component.type === 'material' && component.enabled)
      const selectedMaterial = (materialComponent ? materialFor(materialComponent, props.parameters) : null) || selectedRollMaterial(props.form, props.materials, values)
      const widthKey = props.form.finishedSize?.widthParameterKey
      const width = rollWidthValue(selectedMaterial, values.layout_margin_mm || '0')
      if (widthKey && width) values[widthKey] = width
    }
    const heightKey = props.form.finishedSize?.heightParameterKey
    if (heightKey && !values[heightKey]) {
      if (token === previewToken) layoutMessage.value = 'Set a default custom height / length to preview layout pricing.'
      return
    }
    try {
      const price = await pricingApi.draft(props.form, values, values[props.form.finishedSize!.quantityParameterKey] || '1')
      if (token === previewToken) { layoutPrice.value = price; layoutMessage.value = '' }
    } catch(e) {
      if (token !== previewToken) return
      const message = String(e)
      layoutMessage.value = message.includes('Finished width') || message.includes('Finished height')
        ? 'Set valid default finished dimensions to preview layout pricing.'
        : message
    }
  },250)
},{deep:true,immediate:true})
const rows = computed<BreakdownRow[]>(() => {
  if (props.form.finishedSize?.quantityParameterKey) return effectiveComponents.value.filter(c => c.enabled).map(component => {
    const line = layoutPrice.value?.components.find(c => c.id === component.id)
    return {component, amount:line?.amountRial || 0, detail:line?.explanation || layoutMessage.value, missing:!line}
  })
  let running = 0
  return effectiveComponents.value.filter((component) => component.enabled).map((component) => {
    if (component.type === 'overhead' || component.type === 'waste') {
      const percentage = Math.max(0, numeric(component.percentage, 0))
      const amount = running * percentage / 100
      running += amount
      return { component, amount, detail: `${percentage}% of previous costs`, missing: false }
    }
    const amount = baseAmount(component) * usageFor(component) * Math.max(0, numeric(component.multiplier))
    running += amount
    const missing = (component.type === 'material' && categoryRequirements.value.material && !materialFor(component)) || (component.type === 'machine' && categoryRequirements.value.machine && !machineFor(component)) || (component.type === 'service' && !props.services.some((service) => service.id === component.referenceId && service.active))
    const detail = component.type === 'material' ? (materialFor(component)?.name || 'Choose a material') : component.type === 'machine' ? `${machineFor(component)?.name || 'Choose a machine'}${machineRateFor(component)?.name ? ` · ${machineRateFor(component)?.name}` : ''}` : component.type === 'service' ? (props.services.find((service) => service.id === component.referenceId)?.name || 'Choose a service') : 'Included in estimate'
    const rateMissing = component.type === 'machine' && categoryRequirements.value.machine && !!machineFor(component) && !machineRateFor(component)
    return { component, amount, detail, missing: missing || rateMissing }
  })
})

const subtotal = computed(() => rows.value.filter((row) => row.component.type !== 'overhead' && row.component.type !== 'waste').reduce((total, row) => total + row.amount, 0))
const total = computed(() => rows.value.reduce((sum, row) => sum + row.amount, 0))
const defaultEstimateSummary = computed(() => effectiveComponents.value
  .filter((component) => component.enabled && (component.type === 'material' || component.type === 'machine'))
  .map((component) => {
    if (component.type === 'material') return `Material: ${materialFor(component)?.name || 'not set'}`
    const machine = machineFor(component)
    const rate = machine ? machineRateFor(component) : null
    return `Machine: ${machine?.name || 'not set'}${rate?.name ? ` · ${rate.name}` : ''}`
  }))
watch(total, (value) => emit('update:total', Math.round(value)), { immediate: true })
watch(rows, (value) => emit('update:breakdown', value.map((row) => ({ name: row.component.name || 'Cost component', amount: Math.round(row.amount), detail: row.detail, missing: row.missing }))), { immediate: true })
</script>

<template>
  <section class="min-w-0 space-y-4" :aria-label='$t("Cost breakdown preview")'>
    <ServiceOverviewIdentity :form="form" :active="active" />

    <ServiceOverviewSection :title='$t("Cost estimate")' :description="$ui(form.finishedSize?.quantityParameterKey ? 'Complete batch cost from the default dimensions and quantity.' : 'Current cost per service unit from the configured components and their selected defaults.')">
      <div v-if="defaultEstimateSummary.length" class="mb-3 rounded-box border border-primary/20 bg-primary/5 px-3 py-2.5 text-xs leading-5 text-base-content/70"><strong class="font-medium text-primary">{{ $t("Defaults used for estimate:") }}</strong> {{ $ui(defaultEstimateSummary.join(' · ')) }}</div>
      <div v-if="rows.length" class="space-y-2">
        <div class="divide-y divide-base-300/70">
          <div v-for="row in rows" :key="row.component.id" class="flex min-w-0 items-center gap-2 py-2.5 first:pt-0 last:pb-0">
            <span class="size-2 shrink-0 rounded-full" :class="row.missing ? 'bg-warning' : 'bg-primary'"></span>
            <div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ $ui(row.component.name || 'Cost component') }}</strong><small class="block truncate text-xs text-base-content/55">{{ $ui(row.detail) }}</small></div>
            <span class="shrink-0 text-sm tabular-nums" :class="row.missing ? 'text-warning' : ''">{{ $ui(row.missing ? 'Needs setup' : formatMoney(row.amount, currencyUnit)) }}</span>
          </div>
        </div>
        <div class="flex items-center justify-between gap-3 border-t border-base-300/75 pt-2.5 text-sm"><span class="font-medium">{{ $t("Subtotal") }}</span><strong class="tabular-nums">{{ formatMoney(subtotal, currencyUnit) }}</strong></div>
        <div class="flex items-center justify-between gap-3 border-t border-base-300/75 pt-2.5 text-sm"><span class="font-semibold">{{ $t("Estimated cost") }}</span><strong class="text-base text-success tabular-nums">{{ formatMoney(total, currencyUnit) }}</strong></div>
      </div>
      <p v-else class="rounded-box border border-dashed border-base-300/80 p-4 text-center text-sm leading-5 text-base-content/60">{{ $t("Add a cost component to see its estimated breakdown.") }}</p>
    </ServiceOverviewSection>

    <div class="flex items-start gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/65"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="15" aria-hidden="true" /><span>{{ $t("Material costs use the highest recorded purchase cost when available. This is a setup preview; final pricing is calculated after the service is saved.") }}</span></div>
  </section>
</template>
