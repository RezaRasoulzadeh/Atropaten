<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Archive, Calculator, Check, ChevronRight, CircleHelp, Edit3, Factory, FileText, Flag, Hash, Layers3, Package, RotateCcw, Ruler, Sparkles, Tag, Trash2 } from 'lucide-vue-next'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import { serviceEstimatedSellingPrice } from './serviceDefaultPricing'
import { serviceCategoryRequirements } from './serviceCategory'
import { findMaterialVariant, mappedMaterialOptionValues } from './serviceMaterialResolution'

type DetailTab = 'parameters' | 'components' | 'pricing' | 'info'

const props = defineProps<{
  service: ServiceRecord
  materials: MaterialRecord[]
  machines: MachineRecord[]
  services: ServiceRecord[]
  currencyUnit: CurrencyUnit
  busy?: boolean
}>()

const emit = defineEmits<{
  edit: []
  archive: []
  reactivate: []
  remove: []
}>()

const activeTab = ref<DetailTab>('parameters')
watch(() => props.service.id, () => { activeTab.value = 'parameters' })

const parameters = computed(() => props.service.parameters || [])
const components = computed(() => props.service.components || [])
const pricingRule = computed(() => props.service.pricingRule)

const estimatedPrice = computed(() => serviceEstimatedSellingPrice(props.service, props.materials, props.machines, props.services))

const categoryLabel = computed(() => props.service.category || 'Uncategorized')
const categoryRequirements = computed(() => serviceCategoryRequirements(props.service.category))
const materialComponentCount = computed(() => components.value.filter((component) => component.type === 'material').length)
const machineComponentCount = computed(() => components.value.filter((component) => component.type === 'machine').length)
const priceSummary = computed(() => {
  if (!pricingRule.value) return 'Not configured'
  if (pricingRule.value.type === 'manual') return 'Set in order'
  return props.service.finishedSize?.quantityParameterKey ? 'By size & quantity' : estimatedPrice.value !== null ? formatMoney(estimatedPrice.value, props.currencyUnit) : 'Needs setup'
})
const tabs: { id: DetailTab; label: string }[] = [
  { id: 'parameters', label: 'Parameters' },
  { id: 'components', label: 'Cost components' },
  { id: 'pricing', label: 'Pricing' },
  { id: 'info', label: 'Additional info' },
]

function parameterTypeLabel(type: string) {
  return ({ integer: 'Number', decimal: 'Decimal', boolean: 'Yes / no', choice: 'Choice', 'material-reference': 'Material', 'machine-reference': 'Machine' } as Record<string, string>)[type] || type
}

function parameterIcon(type: string) {
  if (type === 'integer') return Hash
  if (type === 'decimal') return Ruler
  if (type === 'boolean') return Check
  if (type === 'material-reference') return Package
  if (type === 'machine-reference') return Factory
  return Layers3
}

function componentTypeLabel(type: string) {
  return ({ material: 'Material', machine: 'Machine', service: 'Service', labor: 'Labor', outsourced: 'Outsourced', fixed: 'Fixed', overhead: 'Overhead', waste: 'Waste', manual: 'Manual' } as Record<string, string>)[type] || type
}

function sourceKeys(parameter: ServiceRecord['parameters'][number]) {
  const source = parameter.materialSource
  if (!source) return []
  return source.exposedAttributeKeys?.length ? source.exposedAttributeKeys : source.exposedAttributeKey ? [source.exposedAttributeKey] : []
}

function attributeValue(attribute: any) {
  if (!attribute) return ''
  if (attribute.valueType === 'decimal') return String(attribute.decimalValue ?? '')
  if (attribute.valueType === 'integer') return String(attribute.integerValue ?? '')
  if (attribute.valueType === 'enum') return String(attribute.enumCode ?? '').trim()
  if (attribute.valueType === 'boolean') return attribute.booleanValue ? 'true' : 'false'
  return String(attribute.textValue ?? '').trim()
}

function materialOptionCount(parameter: ServiceRecord['parameters'][number]) {
	const source = parameter.materialSource
	if (!source) return parameter.options?.length || 0
	const mappedValues = mappedMaterialOptionValues(props.service.materialVariants, parameter.key)
	if (mappedValues.size) return mappedValues.size
	const candidates = props.materials.filter((material) => material.active && (!source.allowedKinds?.length || source.allowedKinds.includes(material.kind)))
  if (source.selectMaterial) return candidates.length
  const keys = sourceKeys(parameter)
  const values = new Set(candidates.map((material) => {
    const parts = keys.map((key) => attributeValue(material.attributes?.find((attribute: any) => attribute.key === key)))
    return parts.length && parts.every(Boolean) ? parts.join('\u001f') : ''
  }).filter(Boolean))
  return values.size || source.allowedValues?.length || 0
}

function machineOptionCount(parameter: ServiceRecord['parameters'][number]) {
	const configured = new Set(parameter.options || [])
	return props.machines.filter((machine) => machine.active && (!configured.size || configured.has(machine.id))).length
}

function machineRateOptionCount(parameter: ServiceRecord['parameters'][number]) {
	const machineParameters = components.value
		.filter((component) => component.type === 'machine' && component.rateParameterKey === parameter.key)
		.map((component) => components.value.find((candidate) => candidate.type === 'machine' && candidate.parameterKey === component.parameterKey))
	const machineIDs = new Set(machineParameters.map((component) => {
		if (!component) return ''
		if (component.referenceId) return component.referenceId
		return parameters.value.find((item) => item.key === component.parameterKey)?.defaultValue || ''
	}).filter(Boolean))
	const candidates = props.machines.filter((machine) => machine.active && (!machineIDs.size || machineIDs.has(machine.id)))
	const rates = new Set(candidates.flatMap((machine) => machine.rates?.filter((rate) => rate.active).map((rate) => rate.id) || (machine.rateRial > 0 ? [`${machine.id}:standard`] : [])))
	return rates.size
}

function materialFor(component: ServiceRecord['components'][number]) {
  if (component.type !== 'material') return null
  if (component.usageMode === 'parameter' && !component.referenceId) {
    const parameter = parameters.value.find((item) => item.key === component.parameterKey)
    if (parameter?.type === 'material-reference' || parameter?.materialSource?.selectMaterial) {
      return props.materials.find((material) => material.id === parameter.defaultValue && material.active) || null
    }
    if (parameter?.materialSource) {
      const groups = parameters.value.filter((item) => item.type === 'choice' && item.materialSource && !item.materialSource.selectMaterial)
      const values = Object.fromEntries(groups.map((group) => [group.key, group.defaultValue]))
      if (groups.some((group) => !group.defaultValue)) return null
		const variant = findMaterialVariant(props.service.materialVariants, values)
      return props.materials.find((material) => material.id === variant?.materialId && material.active) || null
    }
    return null
  }
  return props.materials.find((material) => material.id === component.referenceId && material.active) || null
}

function componentReference(component: ServiceRecord['components'][number]) {
  if (component.type === 'material') return materialFor(component)?.name || (component.referenceId ? 'Selected material' : 'Choose a material')
  if (component.type === 'machine') {
    const machine = machineFor(component)
    const rate = machineRateFor(component)
    return machine ? `${machine.name}${rate?.name ? ` · ${rate.name}` : ''}` : (component.referenceId ? 'Selected machine' : 'Choose a machine')
  }
  if (component.type === 'service') return props.services.find((item) => item.id === component.referenceId)?.name || (component.referenceId ? 'Selected service' : 'Choose a service')
  if (component.type === 'overhead' || component.type === 'waste') return `${component.percentage || 0}% of previous costs`
  if (component.type === 'manual') return 'Entered when ordering'
  return component.rateRial ? formatMoney(component.rateRial, props.currencyUnit) : 'Configured cost'
}

function machineFor(component: ServiceRecord['components'][number]) {
  if (component.type !== 'machine') return null
  const selectedID = component.usageMode === 'parameter' && !component.referenceId
    ? parameters.value.find((parameter) => parameter.key === component.parameterKey)?.defaultValue
    : component.referenceId
  return props.machines.find((machine) => machine.id === selectedID && machine.active) || null
}

function normalizeRate(value: string) {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, '')
}

function machineRateFor(component: ServiceRecord['components'][number]) {
  const machine = machineFor(component)
  if (!machine) return null
  const rates = machine.rates?.length ? machine.rates.filter((rate) => rate.active) : [{ id: 'default', name: 'Standard', selectorValue: '', selectorPredefinedKey: '', rateRial: machine.rateRial, setupCostRial: machine.setupCostRial, active: true }]
  if (component.rateId) return rates.find((rate) => rate.id === component.rateId) || null
  if (component.rateParameterKey) {
    const parameter = parameters.value.find((item) => item.key === component.rateParameterKey)
    const wanted = normalizeRate(parameter?.defaultValue || '')
    return rates.find((rate) => rate.active && [rate.selectorValue, rate.name, rate.id].some((value) => normalizeRate(value) === wanted)) || null
  }
  return rates[0] || null
}

function parameterLabel(parameter: ServiceRecord['parameters'][number]) {
  if (parameter.type === 'machine-reference' && components.value.some((component) => component.type === 'machine' && component.parameterKey === parameter.key)) return 'Machine selection'
  if (parameter.type === 'choice' && components.value.some((component) => component.type === 'machine' && component.rateParameterKey === parameter.key)) return 'Machine rates'
  return parameter.label || parameter.key
}

function parameterSummary(parameter: ServiceRecord['parameters'][number]) {
  if (parameter.materialSource) return `${materialOptionCount(parameter)} material option${materialOptionCount(parameter) === 1 ? '' : 's'}`
  if (parameter.type === 'machine-reference' && components.value.some((component) => component.type === 'machine' && component.parameterKey === parameter.key)) {
    const count = machineOptionCount(parameter)
    return `${count} machine${count === 1 ? '' : 's'} available`
  }
	if (parameter.type === 'choice' && components.value.some((component) => component.type === 'machine' && component.rateParameterKey === parameter.key)) {
		const count = machineRateOptionCount(parameter)
		return `${count} machine rate${count === 1 ? '' : 's'} available`
	}
  if (parameter.type === 'choice') return `${parameter.options?.length || 0} options`
  if (parameter.unit) return parameter.unit
  return parameter.minValue || parameter.maxValue ? `${parameter.minValue || '—'} – ${parameter.maxValue || '∞'}` : parameterTypeLabel(parameter.type)
}

function componentAmount(component: ServiceRecord['components'][number]) {
  if (component.type === 'material') {
    const material = materialFor(component)
    return material ? material.highestPurchaseUnitCostRial || material.averageUnitCostRial || 0 : 0
  }
  if (component.type === 'machine') return machineRateFor(component)?.rateRial || 0
  return component.rateRial || 0
}

function componentIcon(type: string) {
  if (type === 'machine') return Factory
  if (type === 'service') return Sparkles
  if (type === 'material') return Package
  if (type === 'overhead' || type === 'waste') return Calculator
  return Tag
}

function pricingLabel(type?: string) {
  return ({ markup: 'Cost plus markup', 'fixed-margin': 'Cost plus fixed margin', fixed: 'Fixed price', 'quantity-tiers': 'Quantity tiers', 'per-unit': 'Per-unit parameter', manual: 'Manual price' } as Record<string, string>)[type || ''] || 'Not configured'
}

</script>

<template>
  <section class="service-detail-panel h-auto min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:h-full xl:overflow-y-auto" :aria-label='$t("Service details")'>
    <div class="border-b border-base-300 p-3 sm:p-5">
      <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300 bg-cover bg-center sm:min-h-60" :style="service.imagePath ? { backgroundImage: `url('${service.imagePath}')` } : undefined">
        <div class="absolute inset-0 bg-gradient-to-t from-black/95 via-black/65 to-black/10" aria-hidden="true"></div>
        <div v-if="!service.imagePath" class="absolute inset-0 grid place-items-center text-base-content/35"><Layers3 :size="42" :stroke-width="1.4" aria-hidden="true" /></div>
        <div class="relative z-10 flex min-h-52 items-end justify-start p-4 sm:min-h-60 sm:p-5">
          <div class="w-full min-w-0 text-start text-white">
            <div class="flex min-w-0 items-center justify-start gap-3"><h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ service.name }}</h2><StatusBadge class="shrink-0" :label="$ui(service.active ? 'Active' : 'Archived')" :tone="service.active ? 'green' : 'slate'" /></div>
            <div class="mt-2 flex min-w-0 flex-wrap items-center justify-start gap-x-3 gap-y-1 text-sm text-white/75"><span>{{ $ui(service.code || 'No code') }}</span><span class="size-1 rounded-full bg-white/50" aria-hidden="true"></span><span>{{ categoryLabel }}</span></div>
            <p v-if="service.description" class="mt-3 max-w-prose text-sm leading-6 text-white/75">{{ service.description }}</p>
          </div>
        </div>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
        <button class="btn btn-outline btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('edit')"><Edit3 :size="14" aria-hidden="true" />{{ $t("Edit") }}</button>
        <button v-if="service.active" class="btn btn-outline btn-warning btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('archive')"><Archive :size="14" aria-hidden="true" />{{ $t("Archive") }}</button>
        <button v-else class="btn btn-outline btn-success btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('reactivate')"><RotateCcw :size="14" aria-hidden="true" />{{ $t("Reactivate") }}</button>
        <button class="btn btn-outline btn-error btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('remove')"><Trash2 :size="14" aria-hidden="true" />{{ $t("Remove") }}</button>
      </div>

      <div class="mt-4 grid min-w-0 divide-y divide-base-300 border-y border-base-300 sm:mt-5 sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:first:pl-0"><Tag :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">{{ $t("Price preview") }}</span><strong class="block text-sm">{{ priceSummary }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3"><Package :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">{{ $t("Default unit") }}</span><strong class="block text-sm">{{ $ui(service.defaultUnit || 'piece') }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:pr-0"><Flag :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">{{ $t("Priority") }}</span><strong class="block text-sm">{{ $ui(service.defaultPriority || 'Normal') }}</strong></div></div>
      </div>
    </div>

    <nav class="flex min-w-0 overflow-x-auto border-b border-base-300 px-2" :aria-label='$t("Service details tabs")'>
      <button v-for="tab in tabs" :key="tab.id" class="shrink-0 border-b-2 px-3 py-3 text-sm transition-colors" :class="activeTab === tab.id ? 'border-primary text-primary' : 'border-transparent text-base-content/65 hover:border-base-content/30 hover:text-base-content'" type="button" @click="activeTab = tab.id">{{ $ui(tab.label) }}</button>
    </nav>

    <div class="min-w-0 p-3 sm:p-4">
      <div v-if="activeTab === 'parameters'" class="space-y-2">
        <div class="grid min-w-0 gap-2 pb-2 sm:grid-cols-2">
          <div class="flex min-w-0 items-start gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-2.5">
            <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Package :size="18" aria-hidden="true" /></span>
            <div class="min-w-0"><span class="block text-xs text-base-content/55">{{ $t("Material setup") }}</span><strong class="block text-sm">{{ $ui(categoryRequirements.material ? 'Required' : 'Optional') }}</strong><small class="block truncate text-xs text-base-content/55">{{ $ui(materialComponentCount ? `${materialComponentCount} component${materialComponentCount === 1 ? '' : 's'} configured` : categoryRequirements.material ? 'Not configured' : 'Not needed by category') }}</small></div>
          </div>
          <div class="flex min-w-0 items-start gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-2.5">
            <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Factory :size="18" aria-hidden="true" /></span>
            <div class="min-w-0"><span class="block text-xs text-base-content/55">{{ $t("Machine setup") }}</span><strong class="block text-sm">{{ $ui(categoryRequirements.machine ? 'Required' : 'Optional') }}</strong><small class="block truncate text-xs text-base-content/55">{{ $ui(machineComponentCount ? `${machineComponentCount} group${machineComponentCount === 1 ? '' : 's'} configured` : categoryRequirements.machine ? 'Not configured' : 'Not needed by category') }}</small></div>
          </div>
        </div>
        <div v-for="parameter in parameters" :key="parameter.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-2.5">
          <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><component :is="parameterIcon(parameter.type)" :size="19" aria-hidden="true" /></span>
          <div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ parameterLabel(parameter) }}</strong><span class="block truncate text-xs text-base-content/55">{{ parameterSummary(parameter) }}</span></div>
          <div class="flex shrink-0 items-center gap-2"><span v-if="parameter.required" class="text-xs text-error">{{ $t("Required") }}</span><ChevronRight :size="16" class="text-base-content/45" aria-hidden="true" /></div>
        </div>
        <div v-if="!parameters.length" class="rounded-box border border-dashed border-base-300 p-8 text-center text-sm text-base-content/60">{{ $t("No order inputs configured for this service.") }}</div>
      </div>

      <div v-else-if="activeTab === 'components'" class="space-y-2">
        <div v-for="component in components" :key="component.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-2.5">
          <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><component :is="componentIcon(component.type)" :size="18" aria-hidden="true" /></span>
          <div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ $ui(component.name || 'Cost component') }}</strong><span class="block truncate text-xs text-base-content/55">{{ $ui(componentTypeLabel(component.type)) }} · {{ componentReference(component) }}</span></div>
          <span v-if="!component.enabled" class="badge badge-ghost shrink-0 text-xs">{{ $t("Disabled") }}</span><span v-else-if="component.type !== 'overhead' && component.type !== 'waste' && componentAmount(component)" class="shrink-0 text-sm tabular-nums">{{ formatMoney(componentAmount(component), currencyUnit) }}</span>
        </div>
        <div v-if="!components.length" class="rounded-box border border-dashed border-base-300 p-8 text-center text-sm text-base-content/60">{{ $ui(categoryRequirements.material || categoryRequirements.machine ? 'Required cost setup is not configured yet.' : 'No cost components needed for this category.') }}</div>
      </div>

      <div v-else-if="activeTab === 'pricing'" class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Calculator :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">{{ $ui(pricingLabel(pricingRule?.type)) }}</h3><p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("How the selling price is calculated for this service.") }}</p></div></div>
          <div v-if="estimatedPrice !== null" class="mt-4 flex items-center justify-between gap-3 border-y border-base-300/70 py-3"><span class="text-sm text-base-content/60">{{ $t("Estimated selling price") }}</span><strong class="text-lg tabular-nums text-success">{{ formatMoney(estimatedPrice, currencyUnit) }}</strong></div>
          <div v-else-if="pricingRule && pricingRule.type !== 'manual'" class="mt-4 rounded-box border border-warning/30 bg-warning/5 px-3 py-2.5 text-sm text-warning">{{ $t("Set valid default values and cost references to calculate an estimate.") }}</div>
          <div v-else-if="pricingRule?.type === 'manual'" class="mt-4 rounded-box border border-base-300 px-3 py-2.5 text-sm text-base-content/60">{{ $t("The final price is entered manually for each order.") }}</div>
          <dl v-if="pricingRule" class="mt-4 divide-y divide-base-300/70 text-sm"><div v-if="pricingRule.type === 'fixed'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Selling price per unit") }}</dt><dd class="font-medium tabular-nums">{{ formatMoney(pricingRule.fixedPriceRial, currencyUnit) }}</dd></div><div v-if="pricingRule.type === 'markup'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Markup") }}</dt><dd class="font-medium">{{ pricingRule.markupPercentage }}%</dd></div><div v-if="pricingRule.type === 'fixed-margin'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Fixed margin") }}</dt><dd class="font-medium tabular-nums">{{ formatMoney(pricingRule.fixedMarginRial, currencyUnit) }}</dd></div><div v-if="pricingRule.type === 'per-unit'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Rate") }}</dt><dd class="font-medium tabular-nums">{{ formatMoney(pricingRule.perUnitRateRial, currencyUnit) }}</dd></div><div v-if="pricingRule.type === 'quantity-tiers'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Tiers") }}</dt><dd class="font-medium">{{ pricingRule.tiers?.length || 0 }} {{ $t("configured") }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">{{ $t("Cost components") }}</dt><dd class="font-medium">{{ components.length }}</dd></div></dl><p v-else class="mt-4 text-sm text-base-content/60">{{ $t("No pricing rule configured.") }}</p>
        </div>
      </div>

      <div v-else class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileText :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">{{ $t("Service information") }}</h3><p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("Catalog identity and default behavior.") }}</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Code") }}</dt><dd>{{ $ui(service.code || 'No code') }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Category") }}</dt><dd>{{ categoryLabel }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Created") }}</dt><dd>{{ formatDateTime(service.createdAt) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">{{ $t("Updated") }}</dt><dd>{{ formatDateTime(service.updatedAt) }}</dd></div></dl></div>
        <div v-if="service.description" class="rounded-box border border-base-300 bg-base-200/20 p-4"><h3 class="text-sm font-semibold">{{ $t("Description") }}</h3><p class="mt-2 text-sm leading-6 text-base-content/70">{{ service.description }}</p></div>
        <div class="flex items-start gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/60"><CircleHelp :size="15" class="mt-0.5 shrink-0 text-info" aria-hidden="true" /><span>{{ $t("Use Edit to change the service definition and its pricing setup.") }}</span></div>
      </div>
    </div>
  </section>
</template>
