<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Calculator, ChevronRight, Plus, Sparkles, Trash2 } from 'lucide-vue-next'
import SelectField from '../../components/ui/SelectField.vue'
import ServiceCostEditor from './ServiceCostEditor.vue'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import type { ComponentForm, ComponentType, ParameterForm } from './types'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency'

const props = defineProps<{
  components: ComponentForm[]
  parameters: ParameterForm[]
  materials: MaterialRecord[]
  machines: MachineRecord[]
  services: ServiceRecord[]
  currencyUnit: CurrencyUnit
  showErrors?: boolean
  currentServiceId?: string
}>()

const selectedIndex = ref(0)
const otherType = ref<ComponentType | ''>('')
const showAddMenu = ref(false)
const suggestedFromParameters = ref(false)

const activeComponent = computed(() => props.components[selectedIndex.value] || null)

const componentTypes: Array<{ label: string; value: ComponentType }> = [
  { label: 'Material or paper', value: 'material' },
  { label: 'Machine', value: 'machine' },
  { label: 'Another service', value: 'service' },
  { label: 'Labor', value: 'labor' },
  { label: 'Outsourced work', value: 'outsourced' },
  { label: 'Fixed cost', value: 'fixed' },
  { label: 'Overhead percentage', value: 'overhead' },
  { label: 'Waste percentage', value: 'waste' },
  { label: 'Manual cost', value: 'manual' },
]

function id() {
  return `draft-component-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function emptyComponent(type: ComponentType): ComponentForm {
  return {
    id: id(),
    name: '',
    type,
    referenceId: '',
    usageMode: 'fixed',
    parameterKey: '',
    usageQuantity: '1',
    multiplier: '1',
    rateRial: 0,
    rateInput: '',
    percentage: '',
    rateBasis: type === 'labor' || type === 'outsourced' ? 'hour' : '',
    enabled: true,
    notes: '',
  }
}

function defaultName(type: ComponentType) {
  return {
    material: 'Material cost',
    machine: 'Machine cost',
    service: 'Service cost',
    labor: 'Labor cost',
    outsourced: 'Outsourced work',
    fixed: 'Fixed cost',
    overhead: 'Overhead',
    waste: 'Waste allowance',
    manual: 'Manual cost',
  }[type]
}

function normalizeComponent(component: ComponentForm) {
  if (!component.name.trim()) component.name = defaultName(component.type)
  if (component.type === 'material' || component.type === 'machine') {
    component.rateRial = 0
    component.rateInput = ''
    component.percentage = ''
    component.rateBasis = ''
  } else if (component.type === 'service') {
    component.referenceId = ''
    component.usageMode = 'fixed'
    component.parameterKey = ''
    component.rateRial = 0
    component.rateInput = ''
    component.percentage = ''
    component.rateBasis = ''
  } else if (component.type === 'overhead' || component.type === 'waste') {
    component.referenceId = ''
    component.usageMode = 'fixed'
    component.parameterKey = ''
    component.rateRial = 0
    component.rateInput = ''
    component.multiplier = '1'
    component.rateBasis = ''
  } else {
    component.referenceId = ''
    component.usageMode = 'fixed'
    component.parameterKey = ''
    component.percentage = ''
    if (component.type !== 'labor' && component.type !== 'outsourced') component.rateBasis = ''
  }
}

function costParameter(parameter: ParameterForm) {
  if (parameter.type !== 'choice' && parameter.type !== 'material-reference') return false
  return /paper|stock|substrate|material|media|finish|lamination|ink|color/i.test(`${parameter.key} ${parameter.label}`)
}

function addSuggestedComponents() {
  if (props.components.length) return
  const candidates = props.parameters.filter(costParameter)
  if (!candidates.length) return
  for (const parameter of candidates) {
    props.components.push({
      ...emptyComponent('material'),
      name: `${parameter.label || parameter.key} cost`,
      usageMode: 'parameter',
      parameterKey: parameter.key,
    })
  }
  suggestedFromParameters.value = true
  selectedIndex.value = 0
}

function addOtherComponent(type = otherType.value) {
  if (!type) return
  const component = emptyComponent(type)
  normalizeComponent(component)
  props.components.push(component)
  selectedIndex.value = props.components.length - 1
  otherType.value = ''
  showAddMenu.value = false
}

function removeComponent() {
  if (!activeComponent.value) return
  props.components.splice(selectedIndex.value, 1)
  if (selectedIndex.value >= props.components.length) selectedIndex.value = Math.max(0, props.components.length - 1)
}

function moveComponent(direction: -1 | 1) {
  const target = selectedIndex.value + direction
  if (target < 0 || target >= props.components.length) return
  const [component] = props.components.splice(selectedIndex.value, 1)
  props.components.splice(target, 0, component)
  selectedIndex.value = target
}

function changeType(component: ComponentForm) {
  normalizeComponent(component)
}

function changeRate(component: ComponentForm, value: string) {
  component.rateInput = value
  const parsed = parseMoneyInput(value, props.currencyUnit)
  if (parsed !== null) {
    component.rateRial = parsed
    component.rateInput = formatMoneyInput(parsed, props.currencyUnit)
  }
}

function typeLabel(type: ComponentType) {
  return componentTypes.find((item) => item.value === type)?.label || 'Cost'
}

function sourceLabel(component: ComponentForm) {
  if (component.type === 'material') {
    if (component.usageMode === 'parameter') {
      return props.parameters.find((parameter) => parameter.key === component.parameterKey)?.label || 'Parameter selected'
    }
    return props.materials.find((material) => material.id === component.referenceId)?.name || 'Choose material'
  }
  if (component.type === 'machine') return props.machines.find((machine) => machine.id === component.referenceId)?.name || 'Choose machine'
  if (component.type === 'service') return props.services.find((service) => service.id === component.referenceId)?.name || 'Choose service'
  if (component.type === 'overhead' || component.type === 'waste') return `${component.percentage || '0'}%`
  if (component.type === 'labor' || component.type === 'outsourced' || component.type === 'fixed' || component.type === 'manual') return component.rateRial ? formatMoney(component.rateRial, props.currencyUnit) : 'Set rate'
  return typeLabel(component.type)
}

watch(
  () => props.components.length,
  (length) => {
    if (!length) selectedIndex.value = 0
    else if (selectedIndex.value >= length) selectedIndex.value = length - 1
  },
)

onMounted(addSuggestedComponents)
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Service cost components">
    <div v-if="suggestedFromParameters" class="flex items-start gap-2 rounded-box border border-primary/20 bg-primary/5 px-3 py-2.5 text-xs leading-5 text-base-content/70"><Sparkles class="mt-0.5 shrink-0 text-primary" :size="15" aria-hidden="true" /><span>We created material-based cost rows from your paper, finish, and material parameters. Review each row and add any machine, labor, or fixed cost separately.</span></div>

    <div class="grid min-w-0 gap-4 lg:grid-cols-[minmax(13rem,0.82fr)_minmax(0,1.18fr)]">
      <div class="min-w-0 space-y-2">
        <div class="flex flex-wrap items-end justify-between gap-2 px-1"><div><h3 class="text-sm font-semibold">Cost register</h3><p class="mt-1 text-xs text-base-content/60">Select a row to configure it.</p></div><div class="flex min-w-0 flex-wrap items-center justify-end gap-2"><span class="badge badge-ghost text-xs">{{ components.length }}</span><button class="btn btn-outline btn-sm gap-2" type="button" @click="showAddMenu = !showAddMenu"><Plus :size="15" aria-hidden="true" />Add other component</button><SelectField v-if="showAddMenu" v-model="otherType" class="w-48" label="" aria-label="Other component type" :options="[{ label: 'Choose component type', value: '' }, ...componentTypes]" @update:model-value="addOtherComponent" /></div></div>
        <div v-if="components.length" class="space-y-2">
          <button v-for="(component, index) in components" :key="component.id" class="flex w-full min-w-0 items-center gap-2 rounded-box border p-3 text-start transition-colors" :class="index === selectedIndex ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-100 hover:border-primary/45'" type="button" @click="selectedIndex = index">
            <span class="grid size-8 shrink-0 place-items-center rounded-box bg-base-200 text-base-content/65"><Calculator :size="16" aria-hidden="true" /></span>
            <span class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ component.name || 'New cost' }}</strong><small class="mt-0.5 block truncate text-xs text-base-content/60">{{ typeLabel(component.type) }} · {{ sourceLabel(component) }}</small></span>
            <ChevronRight class="shrink-0 text-base-content/45" :size="16" aria-hidden="true" />
          </button>
        </div>
        <div v-else class="rounded-box border border-dashed border-base-300 p-5 text-center"><Calculator class="mx-auto text-primary" :size="22" aria-hidden="true" /><p class="mt-2 text-sm font-medium">No cost components yet</p><p class="mt-1 text-xs leading-5 text-base-content/60">Add a machine, labor, material, or another cost to build the estimate.</p><button class="btn btn-outline btn-sm mt-3 gap-2" type="button" @click="showAddMenu = true"><Plus :size="14" aria-hidden="true" />Add a component</button></div>
      </div>

      <div v-if="activeComponent" class="min-w-0 rounded-box border border-base-300 bg-base-100">
        <div class="flex items-start justify-between gap-3 border-b border-base-300 px-4 py-4"><div class="min-w-0"><h3 class="text-lg font-semibold">Edit component</h3><p class="mt-1 text-sm text-base-content/65">Configure how this cost contributes to the service price.</p></div><div class="flex shrink-0 items-center gap-1"><button class="btn btn-ghost btn-sm" type="button" :disabled="selectedIndex === 0" aria-label="Move component up" @click="moveComponent(-1)">↑</button><button class="btn btn-ghost btn-sm" type="button" :disabled="selectedIndex === components.length - 1" aria-label="Move component down" @click="moveComponent(1)">↓</button><button class="btn btn-outline btn-error btn-sm" type="button" aria-label="Delete component" @click="removeComponent"><Trash2 :size="14" aria-hidden="true" /></button></div></div>
        <ServiceCostEditor :key="activeComponent.id" class="border-0" hide-summary :component="activeComponent" :index="selectedIndex" :count="components.length" :materials="materials" :machines="machines" :services="services" :current-service-id="currentServiceId" :parameters="parameters" :currency-unit="currencyUnit" :show-errors="showErrors" @remove="removeComponent" @move="moveComponent" @change-type="changeType(activeComponent)" @change-rate="changeRate(activeComponent, $event)" />
      </div>
    </div>
  </section>
</template>
