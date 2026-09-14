<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { AlertTriangle, Check, ChevronRight, CircleHelp, Factory, Gauge, Plus, Settings2, Trash2 } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { MachineRecord } from '../../api/machines'
import type { MaterialRecord } from '../../api/materials'
import type { ServiceRecord } from '../../api/services'
import type { ComponentForm, ParameterForm } from './types'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import ServiceCostEditor from './ServiceCostEditor.vue'

const props = defineProps<{
  components: ComponentForm[]
  parameters: ParameterForm[]
  machines: MachineRecord[]
  materials: MaterialRecord[]
  services: ServiceRecord[]
  currencyUnit: CurrencyUnit
  showErrors?: boolean
  currentServiceId?: string
}>()

const selectedIndex = ref(0)
const selectedSupportingIndex = ref(0)

const activeMachines = computed(() => props.machines.filter((machine) => machine.active))
const machineComponents = computed(() => props.components.filter((component) => component.type === 'machine'))
const supportingComponents = computed(() => props.components.filter((component) => component.type !== 'machine'))
const activeMachineComponent = computed(() => machineComponents.value[selectedIndex.value] || null)
const activeMachineParameter = computed(() => activeMachineComponent.value ? props.parameters.find((parameter) => parameter.key === activeMachineComponent.value?.parameterKey) || null : null)
const activeRateParameter = computed(() => activeMachineComponent.value?.rateParameterKey ? props.parameters.find((parameter) => parameter.key === activeMachineComponent.value?.rateParameterKey) || null : null)
const activeSupportingComponent = computed(() => supportingComponents.value[selectedSupportingIndex.value] || null)

type RateOption = { value: string; label: string; machineCount: number; rates: number }

function id(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function machineParam(component: ComponentForm | null) {
  return component ? props.parameters.find((parameter) => parameter.key === component.parameterKey) || null : null
}

function machineLabel(machine: MachineRecord) {
  return `${machine.name}${machine.code ? ` · ${machine.code}` : ''}`
}

function rateValue(rate: any) {
  return String(rate.selectorValue || rate.name || rate.id || '').trim()
}

function rateOptionsForMachines(machines: MachineRecord[]): RateOption[] {
  const result = new Map<string, RateOption>()
  for (const machine of machines) {
    const rates = machine.rates?.length ? machine.rates.filter((rate: any) => rate.active) : [{ name: 'Standard', id: 'default', selectorValue: '', rateRial: machine.rateRial, rateBasis: machine.rateBasis }]
    for (const rate of rates) {
      const value = rateValue(rate)
      if (!value) continue
      const existing = result.get(value)
      if (existing) {
        existing.machineCount += 1
        existing.rates += 1
      } else {
        result.set(value, { value, label: rate.name || value, machineCount: 1, rates: 1 })
      }
    }
  }
  return Array.from(result.values()).sort((left, right) => left.label.localeCompare(right.label))
}

const allRateOptions = computed(() => rateOptionsForMachines(activeMachines.value))

function componentIncomplete(component: ComponentForm) {
  const machineParameter = machineParam(component)
  const rateParameter = component.rateParameterKey ? props.parameters.find((parameter) => parameter.key === component.rateParameterKey) : null
  return !component.name.trim() || (component.usageMode === 'parameter' ? !machineParameter || machineParameter.type !== 'machine-reference' : !activeMachines.value.some((machine) => machine.id === component.referenceId)) || Boolean(component.rateParameterKey && (!rateParameter || rateParameter.type !== 'choice' || !rateParameter.options.length))
}

function groupTitle(component: ComponentForm) {
  return machineParam(component)?.label || component.name || 'Untitled machine group'
}

function groupSubtitle(component: ComponentForm) {
  const parameter = machineParam(component)
  const rateParameter = component.rateParameterKey ? props.parameters.find((item) => item.key === component.rateParameterKey) : null
  return `${parameter ? 'Choose machine in order' : 'Fixed machine'}${rateParameter ? ` · ${rateParameter.label}` : ' · standard rate'}`
}

function createMachineGroup() {
  if (!activeMachines.value.length) return
  const baseKey = 'machine'
  let machineKey = baseKey
  let suffix = 2
  while (props.parameters.some((parameter) => parameter.key === machineKey)) machineKey = `${baseKey}_${suffix++}`
  const machineParameter: ParameterForm = {
    id: id('parameter'), key: machineKey, label: 'Machine', type: 'machine-reference', required: true,
    defaultValue: activeMachines.value[0]?.id || '', options: [], minValue: null, maxValue: null, unit: '',
  }
  props.parameters.push(machineParameter)

  const rateOptions = allRateOptions.value
  let rateParameter: ParameterForm | null = null
  if (rateOptions.length) {
    let rateKey = 'machine_rate'
    let rateSuffix = 2
    while (props.parameters.some((parameter) => parameter.key === rateKey)) rateKey = `machine_rate_${rateSuffix++}`
    rateParameter = {
      id: id('parameter'), key: rateKey, label: 'Machine rate', type: 'choice', required: true,
      defaultValue: rateOptions[0].value, options: rateOptions.map((option) => option.value), minValue: null, maxValue: null, unit: '',
    }
    props.parameters.push(rateParameter)
  }

  const component: ComponentForm = {
    id: id('component'), name: 'Machine cost', type: 'machine', referenceId: '', usageMode: 'parameter', parameterKey: machineKey,
    rateId: '', rateParameterKey: rateParameter?.key || '', usageQuantity: '1', multiplier: '1', rateRial: 0, rateInput: '', percentage: '', rateBasis: '', enabled: true, notes: '',
  }
  props.components.push(component)
  selectedIndex.value = machineComponents.value.length - 1
  void nextTick(() => document.querySelector<HTMLInputElement>(`[data-machine-group-title="${machineParameter.id}"]`)?.focus())
}

function normalizeMachineGroups() {
  for (const component of props.components.filter((item) => item.type === 'machine')) {
    if (component.usageMode === 'parameter' && component.parameterKey && machineParam(component)) continue
    const base = (component.name || 'machine').toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '') || 'machine'
    let key = base
    let suffix = 2
    while (props.parameters.some((parameter) => parameter.key === key)) key = `${base}_${suffix++}`
    props.parameters.push({ id: id('parameter'), key, label: component.name.replace(/\s+cost$/i, '') || 'Machine', type: 'machine-reference', required: true, defaultValue: component.referenceId || '', options: [], minValue: null, maxValue: null, unit: '' })
    component.usageMode = 'parameter'
    component.parameterKey = key
    component.referenceId = ''
    component.rateId = ''
  }
}

function removeMachineGroup() {
  const component = activeMachineComponent.value
  if (!component) return
  const machineParameterKey = component.parameterKey
  const rateParameterKey = component.rateParameterKey
  const componentIndex = props.components.indexOf(component)
  if (componentIndex >= 0) props.components.splice(componentIndex, 1)
  for (const key of [machineParameterKey, rateParameterKey]) {
    if (!key) continue
    const index = props.parameters.findIndex((parameter) => parameter.key === key)
    if (index >= 0) props.parameters.splice(index, 1)
  }
  selectedIndex.value = Math.max(0, Math.min(selectedIndex.value, machineComponents.value.length - 1))
}

function updateTitle(component: ComponentForm, value: string) {
  const parameter = machineParam(component)
  if (parameter) parameter.label = value
  component.name = value ? `${value} cost` : 'Machine cost'
}

function updateMachineDefault(parameter: ParameterForm, value: string) {
  parameter.defaultValue = value
}

function updateRateOptions(parameter: ParameterForm, values: string[]) {
  parameter.options = values
  if (!values.includes(parameter.defaultValue)) parameter.defaultValue = values[0] || ''
}

function selectMachine(index: number) {
  selectedIndex.value = index
}

function selectSupporting(index: number) {
  selectedSupportingIndex.value = index
}

function addSupportingCost() {
  props.components.push({ id: id('component'), name: '', type: 'fixed', referenceId: '', usageMode: 'fixed', parameterKey: '', rateId: '', rateParameterKey: '', usageQuantity: '1', multiplier: '1', rateRial: 0, rateInput: '', percentage: '', rateBasis: '', enabled: true, notes: '' })
  selectedSupportingIndex.value = supportingComponents.value.length - 1
}

function removeSupportingCost() {
  const component = activeSupportingComponent.value
  if (!component) return
  const index = props.components.indexOf(component)
  if (index >= 0) props.components.splice(index, 1)
  selectedSupportingIndex.value = Math.max(0, Math.min(selectedSupportingIndex.value, supportingComponents.value.length - 1))
}

function moveComponent(component: ComponentForm, direction: -1 | 1) {
  const index = props.components.indexOf(component)
  const target = index + direction
  if (index < 0 || target < 0 || target >= props.components.length) return
  props.components.splice(index, 1)
  props.components.splice(target, 0, component)
}

watch(() => props.components.length, () => {
  if (selectedIndex.value >= machineComponents.value.length) selectedIndex.value = Math.max(0, machineComponents.value.length - 1)
  if (selectedSupportingIndex.value >= supportingComponents.value.length) selectedSupportingIndex.value = Math.max(0, supportingComponents.value.length - 1)
})

normalizeMachineGroups()
watch(() => props.components, normalizeMachineGroups, { deep: true })
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Service machine groups">
    <div class="flex min-w-0 flex-wrap items-start justify-between gap-3 rounded-box border border-primary/25 bg-primary/5 p-4">
      <div class="flex min-w-0 items-start gap-3"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Factory :size="20" aria-hidden="true" /></span><div class="min-w-0"><h2 class="text-base font-semibold">Machine groups</h2><p class="mt-1 max-w-3xl text-sm leading-5 text-base-content/65">Create a machine choice for the order, then connect it to the machine’s configured rate options. The selected rate and usage become part of the estimate.</p></div></div>
      <button class="btn btn-primary btn-sm shrink-0 gap-2" type="button" :disabled="!activeMachines.length" @click="createMachineGroup"><Plus :size="15" aria-hidden="true" />Add machine group</button>
    </div>

    <div v-if="!activeMachines.length" class="flex items-start gap-3 rounded-box border border-warning/30 bg-warning/10 p-4 text-sm"><AlertTriangle class="mt-0.5 shrink-0 text-warning" :size="18" aria-hidden="true" /><div><strong class="font-semibold">Add an active machine first</strong><p class="mt-1 text-xs leading-5 text-base-content/65">Machine groups use the machines and rate profiles already configured in the Machines workspace.</p></div></div>

    <div v-if="!machineComponents.length" class="rounded-box border border-dashed border-primary/35 bg-base-100 p-8 text-center"><Gauge class="mx-auto text-primary" :size="28" aria-hidden="true" /><h3 class="mt-3 text-base font-semibold">Start with a machine group</h3><p class="mx-auto mt-1 max-w-lg text-sm leading-5 text-base-content/60">For example, add a “Digital printer” group. Customers can choose the machine, and—when rate profiles exist—the configured rate options will be available too.</p><button class="btn btn-primary btn-sm mt-4 gap-2" type="button" :disabled="!activeMachines.length" @click="createMachineGroup"><Plus :size="15" aria-hidden="true" />Add machine group</button></div>

    <div v-else class="grid min-w-0 gap-4 lg:grid-cols-[minmax(15rem,0.72fr)_minmax(0,1.28fr)]">
      <div class="min-w-0 space-y-2"><div class="flex items-center justify-between px-1"><div><h3 class="text-sm font-semibold">Your machine groups</h3><p class="mt-1 text-xs text-base-content/60">Each group controls one production machine.</p></div><span class="badge badge-ghost text-xs">{{ machineComponents.length }}</span></div><button v-for="(component, index) in machineComponents" :key="component.id" class="flex w-full min-w-0 items-center gap-3 rounded-box border p-3 text-start transition-colors" :class="index === selectedIndex ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-100 hover:border-primary/45'" type="button" @click="selectMachine(index)"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Factory :size="18" aria-hidden="true" /></span><span class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ groupTitle(component) }}</strong><small class="mt-0.5 block truncate text-xs text-base-content/60">{{ groupSubtitle(component) }}</small></span><span v-if="!componentIncomplete(component)" class="grid size-6 shrink-0 place-items-center rounded-full bg-success/15 text-success"><Check :size="14" aria-hidden="true" /></span><ChevronRight class="shrink-0 text-base-content/45" :size="16" aria-hidden="true" /></button><button class="flex w-full items-center justify-center gap-2 rounded-box border border-dashed border-primary/45 px-3 py-3 text-sm text-primary hover:border-primary" type="button" :disabled="!activeMachines.length" @click="createMachineGroup"><Plus :size="15" aria-hidden="true" />Add machine group</button></div>

      <div v-if="activeMachineComponent && activeMachineParameter" class="min-w-0 space-y-4 rounded-box border border-base-300 bg-base-100 p-4 sm:p-5"><div class="flex flex-wrap items-start justify-between gap-3 border-b border-base-300 pb-4"><div class="min-w-0"><p class="text-xs font-semibold uppercase tracking-wide text-primary">Machine group</p><h3 class="mt-1 text-xl font-semibold">{{ activeMachineParameter.label || 'Untitled machine group' }}</h3><p class="mt-1 text-sm text-base-content/65">Set the order-facing machine choice and the rate behavior for this group.</p></div><button class="btn btn-outline btn-error btn-sm gap-2" type="button" @click="removeMachineGroup"><Trash2 :size="14" aria-hidden="true" />Remove group</button></div>
        <FormField class="gap-1"><span>Group title <em class="text-error">*</em></span><AppInput :model-value="activeMachineParameter.label" :data-machine-group-title="activeMachineParameter.id" class="input w-full min-w-0" :class="{ 'input-error': showErrors && !activeMachineParameter.label.trim() }" placeholder="Digital printer" @update:model-value="updateTitle(activeMachineComponent, $event)" /></FormField>
        <div class="grid min-w-0 gap-4 sm:grid-cols-2"><FormField class="gap-1"><span>Default machine</span><SelectField :model-value="activeMachineParameter.defaultValue" label="" :options="[{ label: 'Choose a machine…', value: '' }, ...activeMachines.map((machine) => ({ label: machineLabel(machine), value: machine.id }))]" @update:model-value="updateMachineDefault(activeMachineParameter, $event)" /><small class="text-xs leading-5 text-base-content/60">Customers can choose any active machine in the order. This is the default suggestion.</small></FormField><div class="rounded-box border border-base-300 bg-base-200/25 p-3"><div class="flex items-start gap-2"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="16" aria-hidden="true" /><div><h4 class="text-sm font-semibold">Rate profiles</h4><p class="mt-1 text-xs leading-5 text-base-content/65">Rates are read from each machine. Add profiles like Black &amp; white or Full color in the Machines workspace.</p></div></div></div></div>

        <div v-if="activeRateParameter" class="rounded-box border border-base-300"><div class="border-b border-base-300 px-4 py-3"><h4 class="text-sm font-semibold">Rate option</h4><p class="mt-1 text-xs text-base-content/60">Customers can choose one configured rate profile for this machine group.</p></div><div class="p-4"><FormField class="gap-1"><span>Rate choice title</span><AppInput v-model="activeRateParameter.label" class="input w-full min-w-0" placeholder="Machine rate" /></FormField><div class="mt-3 grid gap-2 sm:grid-cols-2"><label v-for="option in allRateOptions" :key="option.value" class="flex min-w-0 items-center gap-2 rounded-box border border-base-300 bg-base-100 px-3 py-2 text-sm"><input class="checkbox checkbox-sm" type="checkbox" :checked="activeRateParameter.options.includes(option.value)" @change="updateRateOptions(activeRateParameter, ($event.target as HTMLInputElement).checked ? [...activeRateParameter.options, option.value] : activeRateParameter.options.filter((value) => value !== option.value))" /><span class="min-w-0 flex-1 truncate">{{ option.label }}</span><small class="text-xs text-base-content/50">{{ option.machineCount }} machine{{ option.machineCount === 1 ? '' : 's' }}</small></label></div></div></div>
        <div class="rounded-box border border-base-300 bg-base-200/25 p-4"><div class="flex items-start gap-2"><Settings2 class="mt-0.5 shrink-0 text-primary" :size="16" aria-hidden="true" /><div><h4 class="text-sm font-semibold">Usage and cost basis</h4><p class="mt-1 text-xs leading-5 text-base-content/65">Set how much machine time or output is used for one service unit. Machine rates come from the selected machine profile.</p></div></div><div class="mt-3 grid min-w-0 gap-4 sm:grid-cols-2"><FormField class="gap-1"><span>Amount per service</span><AppInput v-model="activeMachineComponent.usageQuantity" class="input w-full min-w-0" inputmode="decimal" placeholder="1" /></FormField><FormField class="gap-1"><span>Multiply by</span><AppInput v-model="activeMachineComponent.multiplier" class="input w-full min-w-0" inputmode="decimal" placeholder="1" /></FormField></div><label class="mt-3 flex items-center gap-2 text-sm"><input v-model="activeMachineComponent.enabled" class="checkbox checkbox-sm" type="checkbox" />Include this machine cost in pricing</label></div>
        <div v-if="activeMachineParameter.defaultValue && activeRateParameter?.defaultValue" class="rounded-box border border-success/25 bg-success/5 p-3 text-xs text-base-content/70"><strong class="font-medium text-success">Ready for pricing.</strong> The default machine and rate option are configured.</div><div v-else class="rounded-box border border-warning/25 bg-warning/5 p-3 text-xs text-base-content/70"><strong class="font-medium text-warning">Finish the group setup.</strong> Choose a default machine{{ activeRateParameter ? ' and rate option' : '' }} before saving.</div>
      </div>
    </div>

    <section v-if="supportingComponents.length" class="rounded-box border border-base-300 bg-base-100"><div class="flex flex-wrap items-center justify-between gap-3 border-b border-base-300 px-4 py-3"><div><h3 class="text-sm font-semibold">Supporting costs</h3><p class="mt-1 text-xs text-base-content/60">Optional labor, fixed, overhead, or other costs for this service.</p></div><button class="btn btn-outline btn-sm gap-2" type="button" @click="addSupportingCost"><Plus :size="14" aria-hidden="true" />Add supporting cost</button></div><div class="grid min-w-0 gap-4 p-4 lg:grid-cols-[minmax(13rem,0.72fr)_minmax(0,1.28fr)]"><div class="min-w-0 space-y-2"><button v-for="(component, index) in supportingComponents" :key="component.id" class="flex w-full min-w-0 items-center gap-2 rounded-box border p-3 text-start" :class="index === selectedSupportingIndex ? 'border-primary bg-primary/10' : 'border-base-300 hover:border-primary/45'" type="button" @click="selectSupporting(index)"><span class="min-w-0 flex-1 truncate text-sm">{{ component.name || 'Supporting cost' }}</span><ChevronRight :size="16" class="shrink-0 text-base-content/45" aria-hidden="true" /></button><button class="flex w-full items-center justify-center gap-2 rounded-box border border-dashed border-base-300 px-3 py-2.5 text-xs text-base-content/65 hover:border-primary hover:text-primary" type="button" @click="addSupportingCost"><Plus :size="14" aria-hidden="true" />Add supporting cost</button></div><div v-if="activeSupportingComponent" class="min-w-0 rounded-box border border-base-300"><div class="flex items-center justify-between border-b border-base-300 px-4 py-3"><h4 class="text-sm font-semibold">Edit supporting cost</h4><button class="btn btn-outline btn-error btn-xs gap-1" type="button" @click="removeSupportingCost"><Trash2 :size="13" aria-hidden="true" />Remove</button></div><ServiceCostEditor class="border-0" hide-summary :component="activeSupportingComponent" :index="selectedSupportingIndex" :count="supportingComponents.length" :materials="materials" :machines="machines" :services="services" :current-service-id="currentServiceId" :parameters="parameters" :currency-unit="currencyUnit" :show-errors="showErrors" /></div></div></section><button v-else class="flex w-full items-center justify-center gap-2 rounded-box border border-dashed border-base-300 px-3 py-3 text-sm text-base-content/65 hover:border-primary hover:text-primary" type="button" @click="addSupportingCost"><Plus :size="15" aria-hidden="true" />Add supporting cost</button>
  </section>
</template>
