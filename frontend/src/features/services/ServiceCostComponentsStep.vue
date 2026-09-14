<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Calculator, ChevronRight, Plus, Sparkles, Trash2 } from 'lucide-vue-next'
import ServiceCostEditor from './ServiceCostEditor.vue'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import type { ComponentForm, ComponentType, ParameterForm } from './types'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency'
import { ensureSuggestedCostComponents, isMachineParameter, isMaterialParameter, reconcileCostComponents } from './serviceComponentSync'

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
    rateId: '',
    rateParameterKey: '',
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
    component.rateId = ''
    component.rateParameterKey = ''
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

function syncSuggestedComponents() {
  suggestedFromParameters.value = ensureSuggestedCostComponents(props.components, props.parameters)
  if (props.components.length && selectedIndex.value >= props.components.length) selectedIndex.value = props.components.length - 1
}

function addOtherComponent(type: ComponentType) {
  if (!type) return
  const component = emptyComponent(type)
  normalizeComponent(component)
  props.components.push(component)
  selectedIndex.value = props.components.length - 1
  showAddMenu.value = false
}

function addMachineGroup() {
  let machineParameter = props.parameters.find((parameter) => isMachineParameter(parameter))
  if (!machineParameter) {
    const baseKey = 'machine'
    let key = baseKey
    let suffix = 2
    while (props.parameters.some((parameter) => parameter.key === key)) key = `${baseKey}_${suffix++}`
    machineParameter = {
      id: `draft-parameter-${Date.now()}-${Math.random().toString(16).slice(2)}`,
      key,
      label: 'Machine',
      type: 'machine-reference',
      required: true,
      defaultValue: '',
      options: [],
      minValue: null,
      maxValue: null,
      unit: '',
    }
    props.parameters.push(machineParameter)
  }
  const component = emptyComponent('machine')
  component.name = `${machineParameter.label || 'Machine'} cost`
  component.usageMode = 'parameter'
  component.parameterKey = machineParameter.key
  props.components.push(component)
  selectedIndex.value = props.components.length - 1
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
  if (component.type === 'machine') {
    if (component.usageMode === 'parameter' && !component.referenceId) return `${props.parameters.find((parameter) => parameter.key === component.parameterKey)?.label || 'Parameter selected'}${component.rateParameterKey ? ` · ${props.parameters.find((parameter) => parameter.key === component.rateParameterKey)?.label || 'rate'}` : ''}`
    const machine = props.machines.find((item) => item.id === component.referenceId)
    const rate = machine?.rates?.find((item) => item.id === component.rateId)
    return `${machine?.name || 'Choose machine'}${rate ? ` · ${rate.name}` : component.rateParameterKey ? ` · ${props.parameters.find((parameter) => parameter.key === component.rateParameterKey)?.label || 'rate'}` : ''}`
  }
  if (component.type === 'service') return props.services.find((service) => service.id === component.referenceId)?.name || 'Choose service'
  if (component.type === 'overhead' || component.type === 'waste') return `${component.percentage || '0'}%`
  if (component.type === 'labor' || component.type === 'outsourced' || component.type === 'fixed' || component.type === 'manual') return component.rateRial ? formatMoney(component.rateRial, props.currencyUnit) : 'Set rate'
  return typeLabel(component.type)
}

function componentIsIncomplete(component: ComponentForm) {
  if (!component.name.trim()) return true
  if (component.type === 'material') {
    return component.usageMode === 'parameter'
      ? !isMaterialParameter(props.parameters.find((parameter) => parameter.key === component.parameterKey))
      : !props.materials.some((material) => material.id === component.referenceId && material.active)
  }
  if (component.type === 'machine') {
    return component.usageMode === 'parameter'
      ? !isMachineParameter(props.parameters.find((parameter) => parameter.key === component.parameterKey))
      : !props.machines.some((machine) => machine.id === component.referenceId && machine.active)
  }
  if (component.type === 'service') return !props.services.some((service) => service.id === component.referenceId && service.active && service.id !== props.currentServiceId)
  if (component.type === 'overhead' || component.type === 'waste') return !component.percentage.trim()
  return (component.type === 'labor' || component.type === 'outsourced' || component.type === 'fixed' || component.type === 'manual') && !component.rateInput.trim()
}

watch(
  () => props.components.length,
  (length) => {
    if (!length) selectedIndex.value = 0
    else if (selectedIndex.value >= length) selectedIndex.value = length - 1
  },
)

onMounted(syncSuggestedComponents)
watch(
  () => props.parameters,
  () => {
    reconcileCostComponents(props.components, props.parameters)
    syncSuggestedComponents()
  },
  { deep: true },
)
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Service machines and grouped costs">
    <div class="flex flex-wrap items-center justify-between gap-3 rounded-box border border-primary/25 bg-primary/5 p-3.5">
      <div class="flex min-w-0 items-start gap-3">
        <span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Calculator :size="18" aria-hidden="true" /></span>
        <div><h2 class="text-sm font-semibold">Machines and grouped costs</h2><p class="mt-1 max-w-2xl text-xs leading-5 text-base-content/65">Choose a machine from an order input, connect its rate to options such as color or mode, and add supporting costs only when needed.</p></div>
      </div>
      <button class="btn btn-primary btn-sm gap-2" type="button" @click="addMachineGroup"><Plus :size="15" aria-hidden="true" />Add machine group</button>
    </div>
    <div v-if="suggestedFromParameters" class="flex items-start gap-2 rounded-box border border-primary/20 bg-primary/5 px-3 py-2.5 text-xs leading-5 text-base-content/70"><Sparkles class="mt-0.5 shrink-0 text-primary" :size="15" aria-hidden="true" /><span>We created cost rows from explicit inventory-backed material or machine inputs. Review each row and add any fixed, labor, finishing, or overhead costs separately.</span></div>

    <div class="grid min-w-0 gap-4 lg:grid-cols-[minmax(13rem,0.82fr)_minmax(0,1.18fr)]">
      <div class="min-w-0 space-y-2">
        <div class="relative flex flex-wrap items-end justify-between gap-2 px-1"><div><h3 class="text-sm font-semibold">Machine and cost groups</h3><p class="mt-1 text-xs text-base-content/60">Select a group to configure its machine, rate, and usage.</p></div><div class="flex min-w-0 flex-wrap items-center justify-end gap-2"><span class="badge badge-ghost text-xs">{{ components.length }}</span><button class="btn btn-outline btn-sm gap-2" type="button" :aria-expanded="showAddMenu" @click="showAddMenu = !showAddMenu"><Plus :size="15" aria-hidden="true" />Add another cost</button><div v-if="showAddMenu" class="absolute right-0 top-full z-20 grid w-72 gap-1 rounded-box border border-base-300 bg-base-100 p-2 shadow-xl"><button class="flex items-start gap-2 rounded-box border border-primary/20 bg-primary/5 px-2.5 py-2 text-start hover:bg-primary/10" type="button" @click="addMachineGroup"><span class="grid size-7 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Calculator :size="15" aria-hidden="true" /></span><span class="min-w-0"><strong class="block text-sm font-medium">Machine group</strong><small class="block text-xs text-base-content/60">Machine selection plus parameter-based rates</small></span></button><button v-for="item in componentTypes.filter((entry) => entry.value !== 'machine')" :key="item.value" class="flex items-start gap-2 rounded-box px-2.5 py-2 text-start hover:bg-base-200" type="button" @click="addOtherComponent(item.value)"><span class="grid size-7 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Calculator :size="15" aria-hidden="true" /></span><span class="min-w-0"><strong class="block text-sm font-medium">{{ item.label }}</strong><small class="block text-xs text-base-content/60">{{ item.value === 'material' ? 'Inventory stock used by the service' : item.value === 'labor' ? 'Internal work charged by time or unit' : item.value === 'fixed' ? 'A flat amount per service' : item.value === 'overhead' || item.value === 'waste' ? 'A percentage added to the estimate' : 'A linked or manually entered cost' }}</small></span></button></div></div></div>
        <div v-if="components.length" class="space-y-2">
          <button v-for="(component, index) in components" :key="component.id" class="flex w-full min-w-0 items-center gap-2 rounded-box border p-3 text-start transition-colors" :class="index === selectedIndex ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-100 hover:border-primary/45'" type="button" @click="selectedIndex = index">
            <span class="grid size-8 shrink-0 place-items-center rounded-box bg-base-200 text-base-content/65"><Calculator :size="16" aria-hidden="true" /></span>
            <span class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ component.name || 'New cost' }}</strong><small class="mt-0.5 block truncate text-xs text-base-content/60">{{ typeLabel(component.type) }} · {{ sourceLabel(component) }}</small><small v-if="componentIsIncomplete(component)" class="mt-1 block text-xs font-medium text-warning">Needs setup</small></span>
            <ChevronRight class="shrink-0 text-base-content/45" :size="16" aria-hidden="true" />
          </button>
        </div>
        <div v-else class="rounded-box border border-dashed border-base-300 p-5 text-center"><Calculator class="mx-auto text-primary" :size="22" aria-hidden="true" /><p class="mt-2 text-sm font-medium">No machine or supporting costs yet</p><p class="mt-1 text-xs leading-5 text-base-content/60">Add a machine group first, then add supporting costs only when the service needs them.</p><button class="btn btn-outline btn-sm mt-3 gap-2" type="button" @click="showAddMenu = true"><Plus :size="14" aria-hidden="true" />Add a group</button></div>
      </div>

      <div v-if="activeComponent" class="min-w-0 rounded-box border border-base-300 bg-base-100">
        <div class="flex items-start justify-between gap-3 border-b border-base-300 px-4 py-4"><div class="min-w-0"><h3 class="text-lg font-semibold">{{ activeComponent.type === 'machine' ? 'Edit machine group' : 'Edit supporting cost' }}</h3><p class="mt-1 text-sm text-base-content/65">Configure how this group contributes to the service price.</p></div><div class="flex shrink-0 items-center gap-1"><button class="btn btn-ghost btn-sm" type="button" :disabled="selectedIndex === 0" aria-label="Move group up" @click="moveComponent(-1)">↑</button><button class="btn btn-ghost btn-sm" type="button" :disabled="selectedIndex === components.length - 1" aria-label="Move group down" @click="moveComponent(1)">↓</button><button class="btn btn-outline btn-error btn-sm" type="button" aria-label="Delete group" @click="removeComponent"><Trash2 :size="14" aria-hidden="true" /></button></div></div>
        <ServiceCostEditor :key="activeComponent.id" class="border-0" hide-summary :component="activeComponent" :index="selectedIndex" :count="components.length" :materials="materials" :machines="machines" :services="services" :current-service-id="currentServiceId" :parameters="parameters" :currency-unit="currencyUnit" :show-errors="showErrors" @remove="removeComponent" @move="moveComponent" @change-type="changeType(activeComponent)" @change-rate="changeRate(activeComponent, $event)" />
      </div>
    </div>
  </section>
</template>
