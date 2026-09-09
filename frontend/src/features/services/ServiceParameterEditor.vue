<script setup lang="ts">
import { ref } from 'vue'
import { ChevronDown, ChevronUp, Plus, Trash2 } from 'lucide-vue-next'
import FormField from '../../components/ui/FormField.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import AppInput from '../../components/ui/AppInput.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { ParameterForm } from './types'
import type { MaterialRecord } from '../../api/materials'

const props = defineProps<{
  parameter: ParameterForm
  index: number
  count: number
  materials: MaterialRecord[]
  showErrors?: boolean
}>()
const emit = defineEmits<{
  move: [direction: -1 | 1]
  remove: []
  normalize: []
  addOption: []
  removeOption: [index: number]
  labelChange: []
}>()
const expanded = ref(props.index === 0)

function syncExpanded(event: Event) {
  expanded.value = (event.target as HTMLDetailsElement).open
}
function typeLabel(type: string) {
  return {
    integer: 'Quantity or count',
    decimal: 'Decimal measurement',
    boolean: 'Yes / no choice',
    choice: 'Choice from a list',
    'material-reference': 'Material or paper',
  }[type] || 'Input'
}
</script>

<template>
  <article class="border-t border-base-300 first:border-t-0">
    <details class="group min-w-0" :open="expanded" @toggle="syncExpanded">
      <summary class="flex min-w-0 cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 [&::-webkit-details-marker]:hidden">
        <div class="flex min-w-0 items-center gap-3">
          <span class="grid size-8 shrink-0 place-items-center rounded-full bg-primary/15 text-xs font-semibold text-primary tabular-nums">{{ String(index + 1).padStart(2, '0') }}</span>
          <div class="min-w-0">
            <strong class="block truncate text-sm">{{ parameter.label || 'New operator input' }}</strong>
            <small class="block truncate text-xs text-base-content/60">{{ typeLabel(parameter.type) }} · {{ parameter.required ? 'Required' : 'Optional' }}</small>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="index === 0" :aria-label="`Move ${parameter.label || 'input'} up`" @click.stop.prevent="emit('move', -1)"><ChevronUp :size="14" aria-hidden="true" /></button>
          <button class="btn btn-ghost btn-sm" type="button" :disabled="index === count - 1" :aria-label="`Move ${parameter.label || 'input'} down`" @click.stop.prevent="emit('move', 1)"><ChevronDown :size="14" aria-hidden="true" /></button>
          <button class="btn btn-outline btn-error btn-sm" type="button" :aria-label="`Remove ${parameter.label || 'input'}`" @click.stop.prevent="emit('remove')"><Trash2 :size="14" aria-hidden="true" /></button>
          <ChevronDown class="ml-1 shrink-0 transition-transform group-open:rotate-180" :size="16" aria-hidden="true" />
        </div>
      </summary>

      <div class="min-w-0 space-y-4 border-t border-base-300 bg-base-200/20 px-4 py-4">
        <div class="grid min-w-0 gap-3 rounded-box border border-primary/20 bg-primary/5 p-3 sm:grid-cols-[minmax(0,1fr)_minmax(13rem,0.7fr)]">
          <FormField class="gap-1">
            <span>What should the operator enter?</span>
            <AppInput v-model="parameter.label" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !parameter.label.trim() }" type="text" required placeholder="Paper size, quantity, or color" autocomplete="off" @input="emit('labelChange')" />
          </FormField>
          <SelectField v-model="parameter.type" label="Answer type" :invalid="props.showErrors && !parameter.type" :options="[
            { label: 'Quantity or count', value: 'integer' },
            { label: 'Decimal measurement', value: 'decimal' },
            { label: 'Choose from options', value: 'choice' },
            { label: 'Choose a material or paper', value: 'material-reference' },
            { label: 'Yes / no choice', value: 'boolean' },
          ]" @update:model-value="emit('normalize')" />
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3 rounded-box border border-base-300 bg-base-100 px-3 py-2.5">
          <label class="flex items-center gap-2 text-sm"><input class="checkbox" v-model="parameter.required" type="checkbox" />Required when ordering</label>
          <span class="text-xs text-base-content/60">{{ parameter.required ? 'The order cannot continue without an answer.' : 'The operator may leave this blank.' }}</span>
        </div>

        <FormGrid v-if="parameter.type === 'integer' || parameter.type === 'decimal'">
          <FormField class="gap-1"><span>Default value <em>optional</em></span><AppInput v-model="parameter.defaultValue" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="For example, 1" /></FormField>
          <FormField class="gap-1"><span>Unit <em>optional</em></span><AppInput v-model="parameter.unit" class="input w-full min-w-0" type="text" placeholder="sheets, hours, or mm" /></FormField>
          <FormField class="gap-1"><span>Minimum <em>optional</em></span><AppInput v-model="parameter.minValue" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="No minimum" /></FormField>
          <FormField class="gap-1"><span>Maximum <em>optional</em></span><AppInput v-model="parameter.maxValue" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="No maximum" /></FormField>
        </FormGrid>

        <div v-else-if="parameter.type === 'boolean'" class="rounded-box border border-base-300 bg-base-100 p-3">
          <SelectField v-model="parameter.defaultValue" label="Default answer" :options="[{ label: 'No default answer', value: '' }, { label: 'Yes', value: 'true' }, { label: 'No', value: 'false' }]" />
        </div>

        <div v-else-if="parameter.type === 'material-reference'" class="rounded-box border border-base-300 bg-base-100 p-3">
          <SelectField v-model="parameter.defaultValue" label="Default material" :options="[
            { label: 'No default material', value: '' },
            ...materials.map((material) => ({ label: `${material.name}${material.sku ? ` · ${material.sku}` : ''}`, value: material.id })),
          ]" />
          <p class="mt-2 text-xs leading-5 text-base-content/60">Use this input when the operator should choose the paper or stock item used by the service.</p>
        </div>

        <div v-else class="min-w-0 space-y-3 rounded-box border border-base-300 bg-base-100 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div><h4 class="text-sm font-semibold">Choices the operator can pick</h4><p class="mt-1 text-xs text-base-content/60">For paper sizes, add choices such as A4 and A5.</p></div>
            <button class="btn btn-outline btn-sm" type="button" @click="emit('addOption')"><Plus :size="14" aria-hidden="true" />Add choice</button>
          </div>
          <div v-if="parameter.options.length" class="grid min-w-0 gap-2 sm:grid-cols-2">
            <div v-for="(option, optionIndex) in parameter.options" :key="`${parameter.id}-${optionIndex}`" class="flex min-w-0 items-center gap-2">
              <AppInput v-model="parameter.options[optionIndex]" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !parameter.options[optionIndex].trim() }" type="text" required :aria-label="`Choice ${optionIndex + 1}`" placeholder="A4" @input="parameter.defaultValue = parameter.defaultValue === option ? parameter.options[optionIndex] : parameter.defaultValue" />
              <button class="btn btn-outline btn-error btn-sm shrink-0" type="button" :aria-label="`Remove choice ${optionIndex + 1}`" @click="emit('removeOption', optionIndex)"><Trash2 :size="13" aria-hidden="true" /></button>
            </div>
          </div>
          <p v-else class="rounded-box border border-dashed border-base-300 p-3 text-sm text-base-content/60">Add at least one choice.</p>
          <SelectField v-if="parameter.options.length" v-model="parameter.defaultValue" label="Default choice" :options="[{ label: 'No default choice', value: '' }, ...parameter.options.map((option) => ({ label: option, value: option }))]" />
        </div>

        <details class="rounded-box border border-base-300 bg-base-100 px-3 py-2">
          <summary class="cursor-pointer text-xs font-semibold text-base-content/70">Advanced: internal key</summary>
          <FormField class="mt-3 gap-1"><span>Internal key</span><AppInput v-model="parameter.key" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !parameter.key.trim() }" type="text" required placeholder="Generated from the label" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">This is generated automatically and is only used internally by pricing.</small></FormField>
        </details>
      </div>
    </details>
  </article>
</template>
