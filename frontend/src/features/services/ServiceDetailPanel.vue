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
import { serviceDefaultPricingResult, serviceEstimatedSellingPrice } from './serviceDefaultPricing'

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

function componentReference(component: ServiceRecord['components'][number]) {
  if (component.type === 'material') return props.materials.find((item) => item.id === component.referenceId)?.name || (component.referenceId ? 'Selected material' : 'Choose a material')
  if (component.type === 'machine') return props.machines.find((item) => item.id === component.referenceId)?.name || (component.referenceId ? 'Selected machine' : 'Choose a machine')
  if (component.type === 'service') return props.services.find((item) => item.id === component.referenceId)?.name || (component.referenceId ? 'Selected service' : 'Choose a service')
  if (component.type === 'overhead' || component.type === 'waste') return `${component.percentage || 0}% of previous costs`
  if (component.type === 'manual') return 'Entered when ordering'
  return component.rateRial ? formatMoney(component.rateRial, props.currencyUnit) : 'Configured cost'
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
  <section class="service-detail-panel h-auto min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:h-full xl:overflow-y-auto" aria-label="Service details">
    <div class="border-b border-base-300 p-3 sm:p-5">
      <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300 bg-cover bg-center sm:min-h-60" :style="service.imagePath ? { backgroundImage: `url('${service.imagePath}')` } : undefined">
        <div class="absolute inset-0 bg-gradient-to-l from-black/95 via-black/65 to-black/10" aria-hidden="true"></div>
        <div v-if="!service.imagePath" class="absolute inset-0 grid place-items-center text-base-content/35"><Layers3 :size="42" :stroke-width="1.4" aria-hidden="true" /></div>
        <div class="relative z-10 flex min-h-52 items-end justify-start p-4 sm:min-h-60 sm:p-5">
          <div class="w-full min-w-0 text-start text-white">
            <div class="flex min-w-0 items-center justify-start gap-3"><h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ service.name }}</h2><StatusBadge class="shrink-0" :label="service.active ? 'Active' : 'Archived'" :tone="service.active ? 'green' : 'slate'" /></div>
            <div class="mt-2 flex min-w-0 flex-wrap items-center justify-start gap-x-3 gap-y-1 text-sm text-white/75"><span>{{ service.code || 'No code' }}</span><span class="size-1 rounded-full bg-white/50" aria-hidden="true"></span><span>{{ categoryLabel }}</span></div>
            <p v-if="service.description" class="mt-3 max-w-prose text-sm leading-6 text-white/75">{{ service.description }}</p>
          </div>
        </div>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
        <button class="btn btn-outline btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('edit')"><Edit3 :size="14" aria-hidden="true" />Edit</button>
        <button v-if="service.active" class="btn btn-outline btn-warning btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('archive')"><Archive :size="14" aria-hidden="true" />Archive</button>
        <button v-else class="btn btn-outline btn-success btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('reactivate')"><RotateCcw :size="14" aria-hidden="true" />Reactivate</button>
        <button class="btn btn-outline btn-error btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('remove')"><Trash2 :size="14" aria-hidden="true" />Remove</button>
      </div>

      <div class="mt-4 grid min-w-0 divide-y divide-base-300 border-y border-base-300 sm:mt-5 sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:first:pl-0"><Tag :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Last price</span><strong class="block text-sm">{{ estimatedPrice !== null ? formatMoney(estimatedPrice, currencyUnit) : 'Needs setup' }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3"><Package :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Default unit</span><strong class="block text-sm">{{ service.defaultUnit || 'piece' }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:pr-0"><Flag :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Priority</span><strong class="block text-sm">{{ service.defaultPriority || 'Normal' }}</strong></div></div>
      </div>
    </div>

    <nav class="flex min-w-0 overflow-x-auto border-b border-base-300 px-2" aria-label="Service details tabs">
      <button v-for="tab in tabs" :key="tab.id" class="shrink-0 border-b-2 px-3 py-3 text-sm transition-colors" :class="activeTab === tab.id ? 'border-primary text-primary' : 'border-transparent text-base-content/65 hover:border-base-content/30 hover:text-base-content'" type="button" @click="activeTab = tab.id">{{ tab.label }}</button>
    </nav>

    <div class="min-w-0 p-3 sm:p-4">
      <div v-if="activeTab === 'parameters'" class="space-y-2">
        <div v-for="parameter in parameters" :key="parameter.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-2.5">
          <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><component :is="parameterIcon(parameter.type)" :size="19" aria-hidden="true" /></span>
          <div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ parameter.label || parameter.key }}</strong><span class="block truncate text-xs text-base-content/55">{{ parameterTypeLabel(parameter.type) }}<span v-if="parameter.type === 'choice'"> · {{ parameter.options?.length || 0 }} options</span><span v-else-if="parameter.unit"> · {{ parameter.unit }}</span><span v-else-if="parameter.minValue || parameter.maxValue"> · {{ parameter.minValue || '—' }} – {{ parameter.maxValue || '∞' }}</span></span></div>
          <div class="flex shrink-0 items-center gap-2"><span v-if="parameter.required" class="text-xs text-error">Required</span><ChevronRight :size="16" class="text-base-content/45" aria-hidden="true" /></div>
        </div>
        <div v-if="!parameters.length" class="rounded-box border border-dashed border-base-300 p-8 text-center text-sm text-base-content/60">No customer parameters configured.</div>
      </div>

      <div v-else-if="activeTab === 'components'" class="space-y-2">
        <div v-for="component in components" :key="component.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-2.5">
          <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><component :is="componentIcon(component.type)" :size="18" aria-hidden="true" /></span>
          <div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ component.name || 'Cost component' }}</strong><span class="block truncate text-xs text-base-content/55">{{ componentTypeLabel(component.type) }} · {{ componentReference(component) }}</span></div>
          <span v-if="!component.enabled" class="badge badge-ghost shrink-0 text-xs">Disabled</span><span v-else-if="component.type !== 'overhead' && component.type !== 'waste' && component.rateRial" class="shrink-0 text-sm tabular-nums">{{ formatMoney(component.rateRial, currencyUnit) }}</span>
        </div>
        <div v-if="!components.length" class="rounded-box border border-dashed border-base-300 p-8 text-center text-sm text-base-content/60">No cost components configured.</div>
      </div>

      <div v-else-if="activeTab === 'pricing'" class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Calculator :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">{{ pricingLabel(pricingRule?.type) }}</h3><p class="mt-1 text-xs leading-5 text-base-content/60">How the selling price is calculated for this service.</p></div></div>
          <div v-if="estimatedPrice !== null" class="mt-4 flex items-center justify-between gap-3 border-y border-base-300/70 py-3"><span class="text-sm text-base-content/60">Estimated selling price</span><strong class="text-lg tabular-nums text-success">{{ formatMoney(estimatedPrice, currencyUnit) }}</strong></div>
          <div v-else-if="pricingRule && pricingRule.type !== 'manual'" class="mt-4 rounded-box border border-warning/30 bg-warning/5 px-3 py-2.5 text-sm text-warning">Set valid default values and cost references to calculate an estimate.</div>
          <div v-else-if="pricingRule?.type === 'manual'" class="mt-4 rounded-box border border-base-300 px-3 py-2.5 text-sm text-base-content/60">The final price is entered manually for each order.</div>
          <dl v-if="pricingRule" class="mt-4 divide-y divide-base-300/70 text-sm"><div v-if="pricingRule.type === 'fixed'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Selling price per unit</dt><dd class="font-medium tabular-nums">{{ formatMoney(pricingRule.fixedPriceRial, currencyUnit) }}</dd></div><div v-if="pricingRule.type === 'markup'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Markup</dt><dd class="font-medium">{{ pricingRule.markupPercentage }}%</dd></div><div v-if="pricingRule.type === 'fixed-margin'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Fixed margin</dt><dd class="font-medium tabular-nums">{{ formatMoney(pricingRule.fixedMarginRial, currencyUnit) }}</dd></div><div v-if="pricingRule.type === 'per-unit'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Rate</dt><dd class="font-medium tabular-nums">{{ formatMoney(pricingRule.perUnitRateRial, currencyUnit) }}</dd></div><div v-if="pricingRule.type === 'quantity-tiers'" class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Tiers</dt><dd class="font-medium">{{ pricingRule.tiers?.length || 0 }} configured</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Cost components</dt><dd class="font-medium">{{ components.length }}</dd></div></dl><p v-else class="mt-4 text-sm text-base-content/60">No pricing rule configured.</p>
        </div>
      </div>

      <div v-else class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileText :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Service information</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Catalog identity and default behavior.</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Code</dt><dd>{{ service.code || 'No code' }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Category</dt><dd>{{ categoryLabel }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Created</dt><dd>{{ formatDateTime(service.createdAt) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Updated</dt><dd>{{ formatDateTime(service.updatedAt) }}</dd></div></dl></div>
        <div v-if="service.description" class="rounded-box border border-base-300 bg-base-200/20 p-4"><h3 class="text-sm font-semibold">Description</h3><p class="mt-2 text-sm leading-6 text-base-content/70">{{ service.description }}</p></div>
        <div class="flex items-start gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/60"><CircleHelp :size="15" class="mt-0.5 shrink-0 text-info" aria-hidden="true" /><span>Use Edit to change the service definition and its pricing setup.</span></div>
      </div>
    </div>
  </section>
</template>
