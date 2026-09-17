<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { AlertTriangle, Check, ChevronRight, CircleHelp, Factory, Gauge, Plus, Settings2, Trash2 } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
import type { MachineRecord } from '../../api/machines'
import type { ComponentForm, ParameterForm } from './types'

const props = defineProps<{
  components: ComponentForm[]
  parameters: ParameterForm[]
  machines: MachineRecord[]
  showErrors?: boolean
  required?: boolean
}>()

const selectedIndex = ref(0)

const activeMachines = computed(() => props.machines.filter((machine) => machine.active))
const machineComponents = computed(() => props.components.filter((component) => component.type === 'machine'))
const activeMachineComponent = computed(() => machineComponents.value[selectedIndex.value] || null)
const activeMachineParameter = computed(() => activeMachineComponent.value ? props.parameters.find((parameter) => parameter.key === activeMachineComponent.value?.parameterKey) || null : null)
const activeRateParameter = computed(() => activeMachineComponent.value?.rateParameterKey ? props.parameters.find((parameter) => parameter.key === activeMachineComponent.value?.rateParameterKey) || null : null)
const machineDefaultReady = computed(() => Boolean(activeMachineParameter.value && machineOptions(activeMachineParameter.value).some((machine) => machine.id === activeMachineParameter.value?.defaultValue)))
const rateDefaultReady = computed(() => !activeRateParameter.value || (Boolean(activeRateParameter.value.defaultValue) && activeRateParameter.value.options.includes(activeRateParameter.value.defaultValue)))

type RateOption = { value: string; label: string; machineCount: number; rates: number; selectorPredefinedKey: string }

function id(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function machineParam(component: ComponentForm | null) {
  return component ? props.parameters.find((parameter) => parameter.key === component.parameterKey) || null : null
}

function machineLabel(machine: MachineRecord) {
  return `${machine.name}${machine.code ? ` · ${machine.code}` : ''}`
}

function machineOptions(parameter: ParameterForm | null) {
  const configured = new Set(parameter?.options || [])
  return activeMachines.value.filter((machine) => !configured.size || configured.has(machine.id))
}

function rateValue(rate: any) {
  return String(rate.selectorValue || rate.name || rate.id || '').trim()
}

function rateOptionsForMachines(machines: MachineRecord[]): RateOption[] {
  const result = new Map<string, RateOption>()
  for (const machine of machines) {
    const rates = machine.rates?.length ? machine.rates.filter((rate: any) => rate.active) : [{ name: 'Standard', id: 'default', selectorValue: '', selectorPredefinedKey: '', rateRial: machine.rateRial, rateBasis: machine.rateBasis }]
    for (const rate of rates) {
      const value = rateValue(rate)
      if (!value) continue
      const existing = result.get(value)
      const selectorPredefinedKey = String(rate.selectorPredefinedKey || '').trim()
      if (existing) {
        existing.machineCount += 1
        existing.rates += 1
        // A shared rate value can come from both a predefined selector and a
        // custom parameter. In that case the parameter must stay generic.
        if (existing.selectorPredefinedKey !== selectorPredefinedKey) existing.selectorPredefinedKey = ''
      } else {
        result.set(value, { value, label: rate.name || value, machineCount: 1, rates: 1, selectorPredefinedKey })
      }
    }
  }
  return Array.from(result.values()).sort((left, right) => left.label.localeCompare(right.label))
}

function selectorKeyForRateOptions(options: RateOption[]) {
  // Machine-rate values are generated from machine profiles. They are not a
  // static predefined parameter, even when a profile happens to use the color
  // catalog as its selector source.
  void options
  return ''
}

function syncRateParameter(parameter: ParameterForm, machines: MachineRecord[]) {
  const options = rateOptionsForMachines(machines)
  const nextOptions = options.map((option) => option.value)
  if (parameter.options.join('\u001f') !== nextOptions.join('\u001f')) parameter.options = nextOptions

  const nextDefault = nextOptions.includes(parameter.defaultValue) ? parameter.defaultValue : nextOptions[0] || ''
  if (parameter.defaultValue !== nextDefault) parameter.defaultValue = nextDefault

  const nextPredefinedKey = selectorKeyForRateOptions(options)
  if (parameter.predefinedKey !== nextPredefinedKey) parameter.predefinedKey = nextPredefinedKey
}

const allRateOptions = computed(() => rateOptionsForMachines(machineOptions(activeMachineParameter.value)))

function componentIncomplete(component: ComponentForm) {
  const machineParameter = machineParam(component)
  const rateParameter = component.rateParameterKey ? props.parameters.find((parameter) => parameter.key === component.rateParameterKey) : null
  const machineDefaultReady = component.usageMode !== 'parameter' || Boolean(machineParameter && machineOptions(machineParameter).some((machine) => machine.id === machineParameter.defaultValue))
  const rateDefaultReady = !rateParameter || (rateParameter.type === 'choice' && rateParameter.options.length > 0 && rateParameter.options.includes(rateParameter.defaultValue))
  return !component.name.trim() || (component.usageMode === 'parameter' ? !machineParameter || machineParameter.type !== 'machine-reference' || !machineDefaultReady : !activeMachines.value.some((machine) => machine.id === component.referenceId)) || Boolean(component.rateParameterKey && (!rateParameter || !rateDefaultReady))
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
    defaultValue: activeMachines.value[0]?.id || '', options: activeMachines.value.map((machine) => machine.id), minValue: null, maxValue: null, unit: '',
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
      predefinedKey: selectorKeyForRateOptions(rateOptions),
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
    if (component.usageMode === 'parameter' && component.parameterKey && machineParam(component)) {
      const parameter = machineParam(component)
      const configured = parameter?.options.filter((value) => activeMachines.value.some((machine) => machine.id === value)) || []
      if (parameter && !configured.length) parameter.options = activeMachines.value.map((machine) => machine.id)
      if (parameter && !machineOptions(parameter).some((machine) => machine.id === parameter.defaultValue)) parameter.defaultValue = machineOptions(parameter)[0]?.id || ''
      const rateParameter = component.rateParameterKey ? props.parameters.find((item) => item.key === component.rateParameterKey) : null
      if (rateParameter) syncRateParameter(rateParameter, machineOptions(parameter))
      continue
    }
    const base = (component.name || 'machine').toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '') || 'machine'
    let key = base
    let suffix = 2
    while (props.parameters.some((parameter) => parameter.key === key)) key = `${base}_${suffix++}`
    props.parameters.push({ id: id('parameter'), key, label: component.name.replace(/\s+cost$/i, '') || 'Machine', type: 'machine-reference', required: true, defaultValue: component.referenceId || activeMachines.value[0]?.id || '', options: activeMachines.value.map((machine) => machine.id), minValue: null, maxValue: null, unit: '' })
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
  if (!parameter.options.includes(value)) parameter.options = [...parameter.options, value]
  parameter.defaultValue = value
}

function updateMachineOptions(parameter: ParameterForm, machineID: string, checked: boolean) {
  const options = new Set(parameter.options.length ? parameter.options : activeMachines.value.map((machine) => machine.id))
  if (checked) options.add(machineID)
  else {
    options.delete(machineID)
    if (!options.size) return
  }
  parameter.options = activeMachines.value.map((machine) => machine.id).filter((id) => options.has(id))
  if (!parameter.options.includes(parameter.defaultValue)) parameter.defaultValue = parameter.options[0] || ''
}

function updateRateOptions(parameter: ParameterForm, values: string[]) {
  parameter.options = values
  if (!values.includes(parameter.defaultValue)) parameter.defaultValue = values[0] || ''
}

function updateRateDefault(parameter: ParameterForm, value: string) {
  if (parameter.options.includes(value)) parameter.defaultValue = value
}

function selectMachine(index: number) {
  selectedIndex.value = index
}

watch(() => props.components.length, () => {
  if (selectedIndex.value >= machineComponents.value.length) selectedIndex.value = Math.max(0, machineComponents.value.length - 1)
})

normalizeMachineGroups()
watch(() => props.components, normalizeMachineGroups, { deep: true })
watch(() => props.machines, normalizeMachineGroups, { deep: true })
</script>

<template>
  <section class="min-w-0 space-y-4" :aria-label='$t("Service machine groups")'>
      <div class="flex min-w-0 flex-wrap items-start justify-between gap-3 rounded-box border border-primary/25 bg-primary/5 p-4">
      <div class="flex min-w-0 items-start gap-3"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Factory :size="20" aria-hidden="true" /></span><div class="min-w-0"><div class="flex flex-wrap items-center gap-2"><h2 class="text-base font-semibold">{{ $t("Machine groups") }}</h2><span class="badge badge-sm" :class="required ? 'badge-primary' : 'badge-ghost'">{{ $ui(required ? 'Required for this category' : 'Optional for this category') }}</span></div><p class="mt-1 max-w-3xl text-sm leading-5 text-base-content/65">{{ $t("Create a machine choice for the order, then connect it to the machine’s configured rate options. The selected rate and usage become part of the estimate.") }}</p></div></div>
      <button class="btn btn-primary btn-sm shrink-0 gap-2" type="button" :disabled="!activeMachines.length" @click="createMachineGroup"><Plus :size="15" aria-hidden="true" />{{ $t("Add machine group") }}</button>
    </div>

    <div v-if="!activeMachines.length && (required || machineComponents.length)" class="flex items-start gap-3 rounded-box border border-warning/30 bg-warning/10 p-4 text-sm"><AlertTriangle class="mt-0.5 shrink-0 text-warning" :size="18" aria-hidden="true" /><div><strong class="font-semibold">{{ $t("Add an active machine first") }}</strong><p class="mt-1 text-xs leading-5 text-base-content/65">{{ $t("Machine groups use the machines and rate profiles already configured in the Machines workspace.") }}</p></div></div>

    <div v-if="!machineComponents.length && !required" class="rounded-box border border-info/25 bg-info/5 p-4 text-sm leading-6 text-base-content/70">{{ $t("This category does not need machine setup. You can continue without adding a machine group, or add one if the service uses production equipment.") }}</div>

    <div v-if="!machineComponents.length" class="rounded-box border border-dashed border-primary/35 bg-base-100 p-8 text-center"><Gauge class="mx-auto text-primary" :size="28" aria-hidden="true" /><h3 class="mt-3 text-base font-semibold">{{ $t("Start with a machine group") }}</h3><p class="mx-auto mt-1 max-w-lg text-sm leading-5 text-base-content/60">{{ $t("For example, add a “Digital printer” group. Customers can choose the machine, and—when rate profiles exist—the configured rate options will be available too.") }}</p><button class="btn btn-primary btn-sm mt-4 gap-2" type="button" :disabled="!activeMachines.length" @click="createMachineGroup"><Plus :size="15" aria-hidden="true" />{{ $t("Add machine group") }}</button></div>

    <div v-else class="grid min-w-0 gap-4 lg:grid-cols-[minmax(15rem,0.72fr)_minmax(0,1.28fr)]">
      <div class="min-w-0 space-y-2"><div class="flex items-center justify-between px-1"><div><h3 class="text-sm font-semibold">{{ $t("Your machine groups") }}</h3><p class="mt-1 text-xs text-base-content/60">{{ $t("Each group controls one production machine.") }}</p></div><span class="badge badge-ghost text-xs">{{ machineComponents.length }}</span></div><button v-for="(component, index) in machineComponents" :key="component.id" class="flex w-full min-w-0 items-center gap-3 rounded-box border p-3 text-start transition-colors" :class="index === selectedIndex ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-100 hover:border-primary/45'" type="button" @click="selectMachine(index)"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Factory :size="18" aria-hidden="true" /></span><span class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ groupTitle(component) }}</strong><small class="mt-0.5 block truncate text-xs text-base-content/60">{{ groupSubtitle(component) }}</small></span><span v-if="!componentIncomplete(component)" class="grid size-6 shrink-0 place-items-center rounded-full bg-success/15 text-success"><Check :size="14" aria-hidden="true" /></span><ChevronRight class="shrink-0 text-base-content/45" :size="16" aria-hidden="true" /></button><button class="flex w-full items-center justify-center gap-2 rounded-box border border-dashed border-primary/45 px-3 py-3 text-sm text-primary hover:border-primary" type="button" :disabled="!activeMachines.length" @click="createMachineGroup"><Plus :size="15" aria-hidden="true" />{{ $t("Add machine group") }}</button></div>

      <div v-if="activeMachineComponent && activeMachineParameter" class="min-w-0 space-y-4 rounded-box border border-base-300 bg-base-100 p-4 sm:p-5"><div class="flex flex-wrap items-start justify-between gap-3 border-b border-base-300 pb-4"><div class="min-w-0"><p class="text-xs font-semibold uppercase tracking-wide text-primary">{{ $t("Machine group") }}</p><h3 class="mt-1 text-xl font-semibold">{{ $ui(activeMachineParameter.label || 'Untitled machine group') }}</h3><p class="mt-1 text-sm text-base-content/65">{{ $t("Set the order-facing machine choice and the rate behavior for this group.") }}</p></div><button class="btn btn-outline btn-error btn-sm gap-2" type="button" @click="removeMachineGroup"><Trash2 :size="14" aria-hidden="true" />{{ $t("Remove group") }}</button></div>
        <FormField class="gap-1"><span>{{ $t("Group title") }} <em class="text-error">*</em></span><AppInput :model-value="activeMachineParameter.label" :data-machine-group-title="activeMachineParameter.id" class="input w-full min-w-0" :class="{ 'input-error': showErrors && !activeMachineParameter.label.trim() }" :placeholder='$t("Digital printer")' @update:model-value="updateTitle(activeMachineComponent, $event)" /></FormField>
        <div class="grid min-w-0 gap-4 sm:grid-cols-2"><FormField class="gap-1"><span>{{ $t("Machine options") }}</span><div class="space-y-2 rounded-box border border-base-300 bg-base-200/25 p-3"><div v-for="machine in activeMachines" :key="machine.id" class="flex min-w-0 items-center gap-2 rounded-box border border-base-300 bg-base-100 px-3 py-2.5 text-sm"><label class="flex min-w-0 flex-1 items-center gap-2"><input class="checkbox checkbox-sm" type="checkbox" :checked="machineOptions(activeMachineParameter).some((option) => option.id === machine.id)" @change="updateMachineOptions(activeMachineParameter, machine.id, ($event.target as HTMLInputElement).checked)" /><span class="min-w-0 truncate">{{ machineLabel(machine) }}</span></label><label class="flex shrink-0 items-center gap-1.5 text-xs" :class="machineOptions(activeMachineParameter).some((option) => option.id === machine.id) ? 'text-primary' : 'text-base-content/35'"><input class="radio radio-primary radio-xs" type="radio" :name="`default-machine-${activeMachineParameter.id}`" :checked="activeMachineParameter.defaultValue === machine.id" :disabled="!machineOptions(activeMachineParameter).some((option) => option.id === machine.id)" @change="updateMachineDefault(activeMachineParameter, machine.id)" />{{ $t("Default") }}</label></div><p class="text-xs leading-5 text-base-content/60">{{ $t("Include the machines customers may choose, then mark one included machine as the default for estimation.") }}</p></div></FormField><div class="rounded-box border border-base-300 bg-base-200/25 p-3"><div class="flex items-start gap-2"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="16" aria-hidden="true" /><div><h4 class="text-sm font-semibold">{{ $t("Rate profiles") }}</h4><p class="mt-1 text-xs leading-5 text-base-content/65">{{ $t("Rates are read from each machine. Add profiles like Black & white or Full color in the Machines workspace.") }}</p></div></div></div></div>

        <div v-if="activeRateParameter" class="rounded-box border border-base-300"><div class="border-b border-base-300 px-4 py-3"><h4 class="text-sm font-semibold">{{ $t("Rate options") }}</h4><p class="mt-1 text-xs text-base-content/60">{{ $t("Include the profiles customers may choose, then mark one included profile as the default for estimation.") }}</p></div><div class="p-4"><FormField class="gap-1"><span>{{ $t("Rate choice title") }}</span><AppInput v-model="activeRateParameter.label" class="input w-full min-w-0" :placeholder='$t("Machine rate")' /></FormField><div class="mt-3 grid gap-2 sm:grid-cols-2"><div v-for="option in allRateOptions" :key="option.value" class="flex min-w-0 items-center gap-2 rounded-box border border-base-300 bg-base-100 px-3 py-2 text-sm"><label class="flex min-w-0 flex-1 items-center gap-2"><input class="checkbox checkbox-sm" type="checkbox" :checked="activeRateParameter.options.includes(option.value)" @change="updateRateOptions(activeRateParameter, ($event.target as HTMLInputElement).checked ? [...activeRateParameter.options, option.value] : activeRateParameter.options.filter((value) => value !== option.value))" /><span class="min-w-0 truncate">{{ option.label }}</span></label><label class="flex shrink-0 items-center gap-1.5 text-xs" :class="activeRateParameter.options.includes(option.value) ? 'text-primary' : 'text-base-content/35'"><input class="radio radio-primary radio-xs" type="radio" :name="`default-rate-${activeRateParameter.id}`" :checked="activeRateParameter.defaultValue === option.value" :disabled="!activeRateParameter.options.includes(option.value)" @change="updateRateDefault(activeRateParameter, option.value)" />{{ $t("Default") }}</label></div></div></div></div>
        <div class="rounded-box border border-base-300 bg-base-200/25 p-4"><div class="flex items-start gap-2"><Settings2 class="mt-0.5 shrink-0 text-primary" :size="16" aria-hidden="true" /><div><h4 class="text-sm font-semibold">{{ $t("Usage and cost basis") }}</h4><p class="mt-1 text-xs leading-5 text-base-content/65">{{ $t("Set how much machine time or output is used for one service unit. Machine rates come from the selected machine profile.") }}</p></div></div><div class="mt-3 grid min-w-0 gap-4 sm:grid-cols-2"><FormField class="gap-1"><span>{{ $t("Amount per service") }}</span><AppInput v-model="activeMachineComponent.usageQuantity" class="input w-full min-w-0" inputmode="decimal" placeholder="1" /></FormField><FormField class="gap-1"><span>{{ $t("Multiply by") }}</span><AppInput v-model="activeMachineComponent.multiplier" class="input w-full min-w-0" inputmode="decimal" placeholder="1" /></FormField></div><label class="mt-3 flex items-center gap-2 text-sm"><input v-model="activeMachineComponent.enabled" class="checkbox checkbox-sm" type="checkbox" />{{ $t("Include this machine cost in pricing") }}</label></div>
        <div v-if="machineDefaultReady && rateDefaultReady" class="rounded-box border border-success/25 bg-success/5 p-3 text-xs text-base-content/70"><strong class="font-medium text-success">{{ $t("Ready for pricing.") }}</strong> {{ $t("The selected defaults will be used in the overview estimate.") }}</div><div v-else class="rounded-box border border-warning/25 bg-warning/5 p-3 text-xs text-base-content/70"><strong class="font-medium text-warning">{{ $t("Finish the group setup.") }}</strong> {{ $t("Choose a default machine") }}{{ $ui(activeRateParameter ? ' and rate option' : '') }} {{ $t("before saving.") }}</div>
      </div>
    </div>

  </section>
</template>
