<script setup lang="ts">
import { ref } from 'vue'
import { ChevronDown, ChevronUp, Trash2 } from 'lucide-vue-next'
import FormField from '../../components/ui/FormField.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import AppInput from '../../components/ui/AppInput.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { ComponentForm, ParameterForm } from './types'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import type { CurrencyUnit } from '../../utils/currency'
import { componentNeedsPercentage, componentNeedsRate } from './serviceFields'

const props = defineProps<{
  component: ComponentForm
  index: number
  count: number
  materials: MaterialRecord[]
  machines: MachineRecord[]
  services: ServiceRecord[]
  parameters: ParameterForm[]
  currencyUnit: CurrencyUnit
  showErrors?: boolean
  hideSummary?: boolean
  currentServiceId?: string
}>()
const emit = defineEmits<{
  move: [direction: -1 | 1]
  remove: []
  changeType: []
  changeRate: [value: string]
}>()
const expanded = ref(true)

function syncExpanded(event: Event) {
  expanded.value = (event.target as HTMLDetailsElement).open
}
function typeLabel(type: string) {
  return {
    material: 'Material or paper',
    machine: 'Machine',
    service: 'Another service',
    labor: 'Labor',
    outsourced: 'Outsourced work',
    fixed: 'Fixed cost',
    overhead: 'Overhead percentage',
    waste: 'Waste percentage',
    manual: 'Manual cost',
  }[type] || 'Cost'
}
function materialSource(component: ComponentForm) {
  return component.referenceId || component.usageMode === 'fixed' ? 'fixed' : 'parameter'
}
function updateMaterialSource(component: ComponentForm, source: string) {
  if (source === 'parameter') {
    component.referenceId = ''
    component.usageMode = 'parameter'
    component.parameterKey = ''
    return
  }
  component.usageMode = component.referenceId ? component.usageMode : 'fixed'
  component.parameterKey = ''
}
function machineSource(component: ComponentForm) {
  return component.referenceId || component.usageMode === 'fixed' ? 'fixed' : 'parameter'
}
function updateMachineSource(component: ComponentForm, source: string) {
  if (source === 'parameter') {
    component.referenceId = ''
    component.usageMode = 'parameter'
    component.parameterKey = ''
    component.rateId = ''
    return
  }
  component.usageMode = component.referenceId ? component.usageMode : 'fixed'
  component.parameterKey = ''
}
function machineFor(component: ComponentForm) {
  if (component.usageMode === 'parameter' && !component.referenceId) return null
  return props.machines.find((machine) => machine.id === component.referenceId) || null
}
function machineRates(component: ComponentForm) {
  const machine = machineFor(component)
  if (!machine) return []
  const rates = machine.rates?.length
    ? machine.rates
    : [{ id: 'default', name: 'Standard', selectorValue: '', rateBasis: machine.rateBasis, rateRial: machine.rateRial, setupCostRial: machine.setupCostRial, active: true }]
  return rates.filter((rate) => rate.active)
}
function choiceParameters() {
  return props.parameters.filter((parameter) => parameter.type === 'choice' && parameter.options.length)
}
function updateRateParameter(component: ComponentForm, value: string) {
  component.rateParameterKey = value
  if (value) component.rateId = ''
}
function updateRateId(component: ComponentForm, value: string) {
  component.rateId = value
  if (value) component.rateParameterKey = ''
}
function numericParameters() {
  return props.parameters.filter((parameter) => parameter.type === 'integer' || parameter.type === 'decimal')
}
function materialParameters() {
  return props.parameters.filter((parameter) => parameter.type === 'material-reference' || parameter.type === 'choice')
}
function machineParameters() {
  return props.parameters.filter((parameter) => parameter.type === 'machine-reference' || parameter.type === 'choice')
}
function usesQuantitySource() {
  return !componentNeedsPercentage(props.component.type) && props.component.type !== 'manual'
}
function usesQuantityFields() {
  return usesQuantitySource() && (props.component.type !== 'material' || materialSource(props.component) === 'parameter' || props.component.usageMode === 'fixed')
}
function usesFixedQuantitySourceSelector() {
  if (props.component.type === 'material') return materialSource(props.component) === 'fixed'
  if (props.component.type === 'machine') return machineSource(props.component) === 'fixed'
  return true
}
</script>

<template>
  <article class="border-t border-base-300 first:border-t-0">
    <details class="group min-w-0" :open="hideSummary || expanded" @toggle="syncExpanded">
      <summary v-if="hideSummary" class="hidden" aria-hidden="true"></summary>
      <summary v-else class="flex min-w-0 cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 [&::-webkit-details-marker]:hidden">
        <div class="flex min-w-0 items-center gap-3">
          <span class="grid size-8 shrink-0 place-items-center rounded-full bg-primary/15 text-xs font-semibold text-primary tabular-nums">{{ String(index + 1).padStart(2, '0') }}</span>
          <div class="min-w-0">
            <strong class="block truncate text-sm">{{ component.name || 'New cost' }}</strong>
            <small class="block truncate text-xs text-base-content/60">{{ typeLabel(component.type) }} · {{ component.enabled ? 'Included in price' : 'Not included' }}</small>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="index === 0" :aria-label="`Move ${component.name || 'cost'} up`" @click.stop.prevent="emit('move', -1)"><ChevronUp :size="14" aria-hidden="true" /></button>
          <button class="btn btn-ghost btn-sm" type="button" :disabled="index === count - 1" :aria-label="`Move ${component.name || 'cost'} down`" @click.stop.prevent="emit('move', 1)"><ChevronDown :size="14" aria-hidden="true" /></button>
          <button class="btn btn-outline btn-error btn-sm" type="button" :aria-label="`Remove ${component.name || 'cost'}`" @click.stop.prevent="emit('remove')"><Trash2 :size="14" aria-hidden="true" /></button>
          <ChevronDown class="ml-1 shrink-0 transition-transform group-open:rotate-180" :size="16" aria-hidden="true" />
        </div>
      </summary>

      <div class="min-w-0 space-y-4 border-t border-base-300 bg-base-200/20 px-4 py-4">
        <div class="grid min-w-0 gap-3 rounded-box border border-primary/20 bg-primary/5 p-3 sm:grid-cols-[minmax(0,1fr)_minmax(14rem,0.7fr)]">
          <FormField class="gap-1"><span>What is this cost?</span><AppInput v-model="component.name" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !component.name.trim() }" type="text" required placeholder="Paper, printer, or finishing" /></FormField>
          <SelectField v-model="component.type" label="Cost type" :options="[
            { label: 'Material or paper', value: 'material' },
            { label: 'Machine', value: 'machine' },
            { label: 'Another service', value: 'service' },
            { label: 'Labor', value: 'labor' },
            { label: 'Outsourced work', value: 'outsourced' },
            { label: 'Fixed cost', value: 'fixed' },
            { label: 'Overhead percentage', value: 'overhead' },
            { label: 'Waste percentage', value: 'waste' },
            { label: 'Manual cost', value: 'manual' },
          ]" @update:model-value="emit('changeType')" />
        </div>

        <label class="flex items-start gap-2 rounded-box border border-base-300 bg-base-100 px-3 py-2.5 text-sm"><input class="checkbox mt-0.5" v-model="component.enabled" type="checkbox" /><span><strong class="block font-medium">Include this cost in pricing</strong><small class="mt-0.5 block text-xs text-base-content/60">Turn this off to keep the setup without charging for it yet.</small></span></label>

        <div v-if="component.type === 'material'" class="min-w-0 space-y-4 rounded-box border border-base-300 bg-base-100 p-3">
          <div><h4 class="text-sm font-semibold">Which material is used?</h4><p class="mt-1 text-xs leading-5 text-base-content/60">Choose one fixed stock item, or let the operator choose the material when placing the order.</p></div>
          <SelectField :model-value="materialSource(component)" label="Material selection" :options="[
            { label: 'Always use one material', value: 'fixed' },
            { label: 'Let the operator choose a material', value: 'parameter' },
          ]" @update:model-value="updateMaterialSource(component, $event)" />
          <SelectField v-if="materialSource(component) === 'fixed'" v-model="component.referenceId" label="Material" :invalid="props.showErrors && !component.referenceId" :options="[
            { label: 'Select an active material', value: '' },
            ...materials.map((material) => ({ label: `${material.name}${material.sku ? ` · ${material.sku}` : ''}`, value: material.id })),
          ]" />
          <div v-else class="space-y-2">
            <SelectField v-model="component.parameterKey" label="Which operator input chooses it?" :invalid="props.showErrors && !component.parameterKey" :options="[
              { label: 'Select a material or paper input', value: '' },
              ...materialParameters().map((parameter) => ({ label: `${parameter.label || parameter.key} · ${parameter.type === 'choice' ? 'choices' : 'materials'}`, value: parameter.key })),
            ]" />
            <p class="text-xs leading-5 text-base-content/60">Example: connect this cost to a “Paper size” input with choices A4 and A5.</p>
          </div>
        </div>

        <div v-else-if="component.type === 'machine'" class="min-w-0 space-y-4 rounded-box border border-base-300 bg-base-100 p-3">
          <div><h4 class="text-sm font-semibold">Which machine does the work?</h4><p class="mt-1 text-xs leading-5 text-base-content/60">Use one fixed machine, or let the operator choose a machine when placing the order.</p></div>
          <SelectField :model-value="machineSource(component)" label="Machine selection" :options="[
            { label: 'Always use one machine', value: 'fixed' },
            { label: 'Let the operator choose a machine', value: 'parameter' },
          ]" @update:model-value="updateMachineSource(component, $event)" />
          <SelectField v-if="machineSource(component) === 'fixed'" v-model="component.referenceId" label="Machine" :invalid="props.showErrors && !component.referenceId" :options="[
            { label: 'Select an active machine', value: '' },
            ...machines.map((machine) => ({ label: `${machine.name}${machine.code ? ` · ${machine.code}` : ''}`, value: machine.id })),
          ]" />
          <div v-else class="space-y-2">
            <SelectField v-model="component.parameterKey" label="Which operator input chooses it?" :invalid="props.showErrors && !component.parameterKey" :options="[
              { label: 'Select a machine input', value: '' },
              ...machineParameters().map((parameter) => ({ label: `${parameter.label || parameter.key} · ${parameter.type === 'choice' ? 'choices' : 'machines'}`, value: parameter.key })),
            ]" />
            <p class="text-xs leading-5 text-base-content/60">Example: connect this cost to a “Print method” input with machine choices.</p>
          </div>
          <div class="space-y-2 border-t border-base-300 pt-4">
            <SelectField v-if="machineSource(component) === 'fixed'" :model-value="component.rateId" label="Machine rate" :options="[
              { label: machineRates(component).length ? 'Use the machine standard rate' : 'No rate profiles configured', value: '' },
              ...machineRates(component).map((rate) => ({ label: `${rate.name} · ${rate.rateRial.toLocaleString()} / ${rate.rateBasis}`, value: rate.id })),
            ]" @update:model-value="updateRateId(component, $event)" />
            <SelectField :model-value="component.rateParameterKey" label="Rate varies with (optional)" :options="[
              { label: 'Use the selected machine rate', value: '' },
              ...choiceParameters().map((parameter) => ({ label: `${parameter.label || parameter.key} · matches a machine rate`, value: parameter.key })),
            ]" @update:model-value="updateRateParameter(component, $event)" />
            <p class="text-xs leading-5 text-base-content/60">For example, add “Black & white” and “Full color” rates to the machine, then connect this cost to the Color input. Each option is matched to the rate profile’s selector value or name.</p>
          </div>
        </div>

        <div v-else-if="component.type === 'service'" class="min-w-0 space-y-4 rounded-box border border-base-300 bg-base-100 p-3">
          <div><h4 class="text-sm font-semibold">Which service is included?</h4><p class="mt-1 text-xs leading-5 text-base-content/60">The selected service's estimated cost will be included in this service. Its own components are evaluated using their default values.</p></div>
          <SelectField v-model="component.referenceId" label="Service" :invalid="props.showErrors && !component.referenceId" :options="[
            { label: 'Select an active service', value: '' },
            ...services.filter((service) => service.active && service.id !== currentServiceId).map((service) => ({ label: `${service.name}${service.code ? ` · ${service.code}` : ''}`, value: service.id })),
          ]" />
          <p v-if="currentServiceId" class="text-xs leading-5 text-base-content/60">A service cannot include itself. Circular service dependencies are also rejected when you save.</p>
        </div>

        <div v-if="usesQuantitySource()" class="min-w-0 space-y-3 rounded-box border border-base-300 bg-base-100 p-3">
          <div><h4 class="text-sm font-semibold">How much is used?</h4><p class="mt-1 text-xs leading-5 text-base-content/60">Set the amount consumed for one service unit. Use an operator input when it changes with the order.</p></div>
          <SelectField v-if="usesFixedQuantitySourceSelector()" v-model="component.usageMode" label="Quantity comes from" :options="[
            { label: 'A fixed amount', value: 'fixed' },
            { label: 'An operator input', value: 'parameter' },
          ]" />
          <FormGrid v-if="usesQuantityFields()">
            <FormField class="gap-1"><span>Amount per service</span><AppInput v-model="component.usageQuantity" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !component.usageQuantity.trim() }" type="text" inputmode="decimal" placeholder="1" /></FormField>
            <FormField class="gap-1"><span>Multiply by</span><AppInput v-model="component.multiplier" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !component.multiplier.trim() }" type="text" inputmode="decimal" placeholder="1" /></FormField>
          </FormGrid>
          <SelectField v-if="component.usageMode === 'parameter' && usesFixedQuantitySourceSelector()" v-model="component.parameterKey" label="Which quantity input?" :invalid="props.showErrors && !component.parameterKey" :options="[
            { label: 'Select a quantity input', value: '' },
            ...numericParameters().map((parameter) => ({ label: `${parameter.label || parameter.key}${parameter.unit ? ` · ${parameter.unit}` : ''}`, value: parameter.key })),
          ]" />
        </div>

        <div v-if="componentNeedsRate(component.type)" class="min-w-0 space-y-3 rounded-box border border-base-300 bg-base-100 p-3">
          <div><h4 class="text-sm font-semibold">What does it cost?</h4><p class="mt-1 text-xs leading-5 text-base-content/60">Enter the rate for this cost. Material and machine rates come from their records.</p></div>
          <FormGrid>
            <FormField><span>{{ component.type === 'manual' ? 'Manual amount' : 'Rate' }} ({{ currencyUnit }})</span><AppInput :model-value="component.rateInput" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !component.rateInput.trim() }" :money="currencyUnit" type="text" inputmode="decimal" placeholder="0" @update:model-value="emit('changeRate', $event)" /></FormField>
            <SelectField v-if="component.type === 'labor' || component.type === 'outsourced'" v-model="component.rateBasis" label="Rate is charged per" :options="[{ label: 'Unit', value: 'unit' }, { label: 'Minute', value: 'minute' }, { label: 'Hour', value: 'hour' }]" />
          </FormGrid>
        </div>

        <div v-if="componentNeedsPercentage(component.type)" class="rounded-box border border-base-300 bg-base-100 p-3">
          <FormField><span>{{ component.type === 'waste' ? 'Waste percentage' : 'Overhead percentage' }}</span><AppInput v-model="component.percentage" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !component.percentage.trim() }" type="text" inputmode="decimal" placeholder="For example, 7" /><small class="text-xs leading-5 text-base-content/60">Applied to the costs that come before this item.</small></FormField>
        </div>

        <details class="rounded-box border border-base-300 bg-base-100 px-3 py-2">
          <summary class="cursor-pointer text-xs font-semibold text-base-content/70">Optional note</summary>
          <FormField class="mt-3"><span>Note for your team</span><AppInput v-model="component.notes" class="input w-full min-w-0" type="text" placeholder="Explain this cost or its source" /></FormField>
        </details>
      </div>
    </details>
  </article>
</template>
