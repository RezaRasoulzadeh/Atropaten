<script setup lang="ts">
import FormGrid from '../../components/ui/FormGrid.vue'
import FormField from '../../components/ui/FormField.vue'
import AppInput from '../../components/ui/AppInput.vue'
import SelectField from '../../components/ui/SelectField.vue'
import { ref } from 'vue'
import {ChevronUp,ChevronDown,ChevronRight,Trash2,Plus} from 'lucide-vue-next'
import type {ComponentForm,ParameterForm} from './types'
import type {MaterialRecord} from '../../api/materials'
import type {MachineRecord} from '../../api/machines'
import type {CurrencyUnit} from '../../utils/currency'
import {componentNeedsRate,componentNeedsReference,componentNeedsPercentage} from './serviceFields'
const props=defineProps<{component:ComponentForm;index:number;count:number;materials:MaterialRecord[];machines:MachineRecord[];parameters:ParameterForm[];currencyUnit:CurrencyUnit}>()
const emit=defineEmits<{move:[direction:-1|1];remove:[];changeType:[];changeRate:[]}>()
const expanded=ref(props.index === 0)
function syncExpanded(event: Event) {
  expanded.value = (event.target as HTMLDetailsElement).open
}
</script>
<template><details class="group min-w-0 border-t border-base-300" :open="expanded" @toggle="syncExpanded">
              <summary class="flex min-w-0 cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 [&::-webkit-details-marker]:hidden">
                <div class="flex min-w-0 items-center gap-3">
                  <span class="grid size-7 shrink-0 place-items-center rounded-full bg-base-300 text-xs font-semibold tabular-nums">{{ String(index + 1).padStart(2, '0') }}</span>
                  <div class="min-w-0">
                    <strong class="block truncate">{{ component.name || 'Untitled component' }}</strong>
                    <small class="block truncate text-xs text-base-content/60">{{ component.type }} · {{ component.enabled ? 'Included' : 'Disabled' }}</small>
                  </div>
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <button
                    class="btn btn-ghost btn-sm"
                    type="button"
                    :disabled="index === 0"
                    :aria-label="`Move ${component.name || 'component'} up`"
                    @click.stop.prevent="emit('move',-1)"
                  >
                    <ChevronUp :size="14" :stroke-width="1.8" aria-hidden="true" /></button
                  ><button
                    class="btn btn-ghost btn-sm"
                    type="button"
                    :disabled="index === count - 1"
                    :aria-label="`Move ${component.name || 'component'} down`"
                    @click.stop.prevent="emit('move',1)"
                  >
                    <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" /></button
                  ><button
                    class="btn btn-outline btn-error btn-sm"
                    type="button"
                    :aria-label="`Remove ${component.name || 'component'}`"
                    @click.stop.prevent="emit('remove')"
                  >
                    <Trash2 :size="14" :stroke-width="1.8" aria-hidden="true" />
                  </button><ChevronRight class="shrink-0 transition-transform group-open:rotate-90" :size="15" :stroke-width="1.8" aria-hidden="true" />
                </div>
              </summary>
              <div class="min-w-0 space-y-3 border-t border-base-300 px-4 py-4">
              <FormGrid
                ><FormField class="gap-1"
                  ><span>Name</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="component.name"
                    type="text"
                    placeholder="Paper cost" /></FormField
                ><SelectField
                  v-model="component.type"
                  label="Type"
                  :options="[
                    { label: 'Material', value: 'material' },
                    { label: 'Machine', value: 'machine' },
                    { label: 'Labor', value: 'labor' },
                    { label: 'Outsourced', value: 'outsourced' },
                    { label: 'Fixed', value: 'fixed' },
                    { label: 'Overhead', value: 'overhead' },
                    { label: 'Waste', value: 'waste' },
                    { label: 'Manual', value: 'manual' },
                  ]"
                  @update:model-value="emit('changeType')"
              /></FormGrid>
              <FormField class="gap-1"
                ><input class="checkbox" v-model="component.enabled" type="checkbox" />Enabled for
                future pricing</FormField
              >
              <SelectField
                v-if="componentNeedsReference(component.type)"
                v-model="component.referenceId"
                :label="component.type === 'material' ? 'Material reference' : 'Machine reference'"
                :options="[
                  { label: `Select an active ${component.type}`, value: '' },
                  ...(component.type === 'material' ? materials : machines).map((item) => ({
                    label: item.name,
                    value: item.id,
                  })),
                ]"
              />
              <FormGrid v-if="!componentNeedsPercentage(component.type)"
                ><SelectField
                  v-model="component.usageMode"
                  label="Usage source"
                  :options="[
                    { label: 'Fixed quantity', value: 'fixed' },
                    { label: 'Numeric parameter', value: 'parameter' },
                  ]" /><FormField class="gap-1"
                  ><span>Fixed quantity</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="component.usageQuantity"
                    type="text"
                    inputmode="decimal"
                    placeholder="1" /></FormField
                ><FormField class="gap-1"
                  ><span>Multiplier</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="component.multiplier"
                    type="text"
                    inputmode="decimal"
                    placeholder="1" /></FormField
              ></FormGrid>
              <div
                class="gap-1"
                v-if="
                  component.usageMode === 'parameter' && !componentNeedsPercentage(component.type)
                "
                ><SelectField
                  v-model="component.parameterKey"
                  label="Usage parameter"
                  :options="[
                    { label: 'Select numeric parameter', value: '' },
                    ...parameters.map((parameter) => ({
                      label: `${parameter.label || parameter.key} · ${parameter.key}`,
                      value: parameter.key,
                    })),
                  ]"
                /><small
                  v-if="
                    component.parameterKey &&
                    !parameters.some(
                      (parameter) => parameter.key === component.parameterKey,
                    )
                  "
                  class="block text-xs leading-5 text-base-content/60"
                  >This reference is no longer numeric and will be rejected until corrected.</small
                ></div
              >
              <FormGrid v-if="componentNeedsRate(component.type)"
                ><FormField class="gap-1"
                  ><span>Rate ({{ currencyUnit }})</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="component.rateInput"
                    type="text"
                    inputmode="decimal"
                    placeholder="0"
                    @input="emit('changeRate')" /></FormField
                ><SelectField
                  v-if="component.type === 'labor' || component.type === 'outsourced'"
                  v-model="component.rateBasis"
                  label="Rate basis"
                  :options="[
                    { label: 'Per unit', value: 'unit' },
                    { label: 'Per minute', value: 'minute' },
                    { label: 'Per hour', value: 'hour' },
                  ]"
              /></FormGrid>
              <FormField class="gap-1" v-if="componentNeedsPercentage(component.type)"
                ><span>Percentage</span
                ><AppInput
                  class="input w-full min-w-0"
                  v-model="component.percentage"
                  type="text"
                  inputmode="decimal"
                  placeholder="10"
                /><small class="block text-xs leading-5 text-base-content/60"
                  >Stored as an exact fixed-scale percentage of the applicable future cost
                  basis.</small
                ></FormField
              >
              <FormField class="gap-1"
                ><span>Notes <em>optional</em></span
                ><AppInput
                  class="input w-full min-w-0"
                  v-model="component.notes"
                  type="text"
                  placeholder="Cost explanation or future basis"
              /></FormField>
              </div>
            </details></template>
