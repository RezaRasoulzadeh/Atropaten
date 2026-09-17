<script setup lang="ts">
import { ref } from 'vue'
import { ChevronDown, ChevronUp, Plus, Trash2 } from 'lucide-vue-next'
import FormField from '../../components/ui/FormField.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import AppInput from '../../components/ui/AppInput.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { ParameterForm } from './types'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'

const props = defineProps<{
  parameter: ParameterForm
  index: number
  count: number
  materials: MaterialRecord[]
  machines: MachineRecord[]
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
    'machine-reference': 'Machine',
  }[type] || 'Input'
}
function updateLabel(value: string) {
  props.parameter.label = value
  emit('labelChange')
}
function updateOption(index: number, previous: string, value: string) {
  props.parameter.options[index] = value
  if (props.parameter.defaultValue === previous) props.parameter.defaultValue = value
}
function setMaterialBacked(event: Event) {
  const enabled = (event.target as HTMLInputElement).checked
  props.parameter.materialSource = enabled ? {
    allowedKinds: ['sheet-stock'],
    exposedAttributeKey: 'grammage_gsm',
    exposedAttributeKeys: ['grammage_gsm'],
    allowedValues: [],
    selectMaterial: false,
    additionalFilters: [],
  } : undefined
  props.parameter.options = []
  props.parameter.defaultValue = ''
  emit('normalize')
}
</script>

<template>
  <article class="border-t border-base-300 first:border-t-0">
    <details class="group min-w-0" :open="expanded" @toggle="syncExpanded">
      <summary class="flex min-w-0 cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 [&::-webkit-details-marker]:hidden">
        <div class="flex min-w-0 items-center gap-3">
          <span class="grid size-8 shrink-0 place-items-center rounded-full bg-primary/15 text-xs font-semibold text-primary tabular-nums">{{ $ui(String(index + 1).padStart(2, '0')) }}</span>
          <div class="min-w-0">
            <strong class="block truncate text-sm">{{ $ui(parameter.label || 'New operator input') }}</strong>
            <small class="block truncate text-xs text-base-content/60">{{ $ui(typeLabel(parameter.type)) }} · {{ $ui(parameter.required ? 'Required' : 'Optional') }}</small>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="index === 0" :aria-label="$ui(`Move ${parameter.label || $ui('input')} up`)" @click.stop.prevent="emit('move', -1)"><ChevronUp :size="14" aria-hidden="true" /></button>
          <button class="btn btn-ghost btn-sm" type="button" :disabled="index === count - 1" :aria-label="$ui(`Move ${parameter.label || $ui('input')} down`)" @click.stop.prevent="emit('move', 1)"><ChevronDown :size="14" aria-hidden="true" /></button>
          <button class="btn btn-outline btn-error btn-sm" type="button" :aria-label="$ui(`Remove ${parameter.label || $ui('input')}`)" @click.stop.prevent="emit('remove')"><Trash2 :size="14" aria-hidden="true" /></button>
          <ChevronDown class="ml-1 shrink-0 transition-transform group-open:rotate-180" :size="16" aria-hidden="true" />
        </div>
      </summary>

      <div class="min-w-0 space-y-4 border-t border-base-300 bg-base-200/20 px-4 py-4">
        <div class="grid min-w-0 gap-3 rounded-box border border-primary/20 bg-primary/5 p-3 sm:grid-cols-[minmax(0,1fr)_minmax(13rem,0.7fr)]">
          <FormField class="gap-1">
            <span>{{ $t("What should the operator enter?") }}</span>
            <AppInput :model-value="parameter.label" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !parameter.label.trim() }" type="text" required :placeholder='$t("Paper size, quantity, or color")' autocomplete="off" @update:model-value="updateLabel" />
          </FormField>
          <SelectField v-model="parameter.type" :label='$t("Answer type")' :invalid="props.showErrors && !parameter.type" :options="[
            { label: 'Quantity or count', value: 'integer' },
            { label: 'Decimal measurement', value: 'decimal' },
            { label: 'Choose from options', value: 'choice' },
            { label: 'Choose a material or paper', value: 'material-reference' },
            { label: 'Choose a machine', value: 'machine-reference' },
            { label: 'Yes / no choice', value: 'boolean' },
          ]" @update:model-value="emit('normalize')" />
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3 rounded-box border border-base-300 bg-base-100 px-3 py-2.5">
          <label class="flex items-center gap-2 text-sm"><input class="checkbox" v-model="parameter.required" type="checkbox" />{{ $t("Required when ordering") }}</label>
          <span class="text-xs text-base-content/60">{{ $ui(parameter.required ? 'The order cannot continue without an answer.' : 'The operator may leave this blank.') }}</span>
        </div>

        <FormGrid v-if="parameter.type === 'integer' || parameter.type === 'decimal'">
          <FormField class="gap-1"><span>{{ $t("Default value") }} <em>{{ $t("optional") }}</em></span><AppInput v-model="parameter.defaultValue" class="input w-full min-w-0" type="text" inputmode="decimal" :placeholder='$t("For example, 1")' /></FormField>
          <FormField class="gap-1"><span>{{ $t("Unit") }} <em>{{ $t("optional") }}</em></span><AppInput v-model="parameter.unit" class="input w-full min-w-0" type="text" :placeholder='$t("sheets, hours, or mm")' /></FormField>
          <FormField class="gap-1"><span>{{ $t("Minimum") }} <em>{{ $t("optional") }}</em></span><AppInput v-model="parameter.minValue" class="input w-full min-w-0" type="text" inputmode="decimal" :placeholder='$t("No minimum")' /></FormField>
          <FormField class="gap-1"><span>{{ $t("Maximum") }} <em>{{ $t("optional") }}</em></span><AppInput v-model="parameter.maxValue" class="input w-full min-w-0" type="text" inputmode="decimal" :placeholder='$t("No maximum")' /></FormField>
        </FormGrid>

        <div v-else-if="parameter.type === 'boolean'" class="rounded-box border border-base-300 bg-base-100 p-3">
          <SelectField v-model="parameter.defaultValue" :label='$t("Default answer")' :options="[{ label: 'No default answer', value: '' }, { label: 'Yes', value: 'true' }, { label: 'No', value: 'false' }]" />
        </div>

        <div v-else-if="parameter.type === 'material-reference'" class="rounded-box border border-base-300 bg-base-100 p-3">
          <SelectField v-model="parameter.defaultValue" :label='$t("Default material")' :options="[
            { label: 'No default material', value: '' },
            ...materials.map((material) => ({ label: `${material.name}${material.sku ? ` · ${material.sku}` : ''}`, value: material.id })),
          ]" />
          <p class="mt-2 text-xs leading-5 text-base-content/60">{{ $t("Use this input when the operator should choose the paper or stock item used by the service.") }}</p>
        </div>

        <div v-else-if="parameter.type === 'machine-reference'" class="rounded-box border border-base-300 bg-base-100 p-3">
          <SelectField v-model="parameter.defaultValue" :label='$t("Default machine")' :options="[
            { label: 'No default machine', value: '' },
            ...machines.map((machine) => ({ label: `${machine.name}${machine.code ? ` · ${machine.code}` : ''}`, value: machine.id })),
          ]" />
          <p class="mt-2 text-xs leading-5 text-base-content/60">{{ $t("Use this input when the operator should choose the machine used by the service.") }}</p>
        </div>

        <div v-else-if="parameter.type === 'choice'" class="space-y-3">
          <label class="flex items-center gap-2 rounded-box border border-base-300 bg-base-100 px-3 py-2.5 text-sm"><input class="checkbox checkbox-sm" type="checkbox" :checked="Boolean(parameter.materialSource)" @change="setMaterialBacked" />{{ $t("Derive choices from compatible inventory materials") }}</label>
          <div v-if="parameter.materialSource" class="min-w-0 space-y-3 rounded-box border border-primary/25 bg-primary/5 p-3">
          <div class="flex items-start gap-3"><span class="mt-0.5 text-primary">◈</span><div><h4 class="text-sm font-semibold">{{ $t("Inventory-backed options") }}</h4><p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("Options are derived from active materials and resolve through material identity, not labels.") }}</p></div></div>
          <div class="grid min-w-0 gap-3 sm:grid-cols-2">
            <SelectField :model-value="parameter.materialSource.allowedKinds[0] || ''" :label='$t("Allowed material kind")' :options="[
              { label: 'Choose kind…', value: '' },
              { label: 'Sheet stock', value: 'sheet-stock' },
              { label: 'Roll media', value: 'roll-media' },
              { label: 'Board', value: 'board' },
              { label: 'Fabric', value: 'fabric' },
              { label: 'Packaging', value: 'packaging' },
              { label: 'Generic consumable', value: 'generic-consumable' },
            ]" @update:model-value="parameter.materialSource.allowedKinds = $event ? [$event] : []" />
            <SelectField :model-value="parameter.materialSource.exposedAttributeKey" :label='$t("Exposed specification")' :options="[
              { label: 'Choose specification…', value: '' },
              { label: 'Grammage (gsm)', value: 'grammage_gsm' },
              { label: 'Material subtype', value: 'material_subtype' },
              { label: 'Finish', value: 'finish' },
              { label: 'Coating', value: 'coating' },
              { label: 'Color', value: 'color' },
            ]" @update:model-value="parameter.materialSource.exposedAttributeKey = $event" />
          </div>
          <label class="flex items-center gap-2 text-sm"><input v-model="parameter.materialSource.selectMaterial" class="checkbox checkbox-sm" type="checkbox" />{{ $t("Also require explicit material selection when multiple physical materials match") }}</label>
          <p class="text-xs leading-5 text-base-content/60">{{ $t("Allowed values are populated from real compatible inventory materials at order time.") }}</p>
          </div>
          <div v-else class="min-w-0 space-y-3 rounded-box border border-base-300 bg-base-100 p-3">
            <div class="flex flex-wrap items-center justify-between gap-2"><div><h4 class="text-sm font-semibold">{{ $t("Choices the operator can pick") }}</h4><p class="mt-1 text-xs text-base-content/60">{{ $t("For arbitrary non-inventory choices, add stable option values.") }}</p></div><button class="btn btn-outline btn-sm" type="button" @click="emit('addOption')"><Plus :size="14" aria-hidden="true" />{{ $t("Add choice") }}</button></div>
            <div v-if="parameter.options.length" class="grid min-w-0 gap-2 sm:grid-cols-2"><div v-for="(option, optionIndex) in parameter.options" :key="`${parameter.id}-${optionIndex}`" class="flex min-w-0 items-center gap-2"><AppInput :model-value="parameter.options[optionIndex]" class="input w-full min-w-0" type="text" required :aria-label="$ui(`Choice ${optionIndex + 1}`)" :placeholder='$t("A4")' @update:model-value="updateOption(optionIndex, option, $event)" /><button class="btn btn-outline btn-error btn-sm shrink-0" type="button" :aria-label="$ui(`Remove choice ${optionIndex + 1}`)" @click="emit('removeOption', optionIndex)"><Trash2 :size="13" aria-hidden="true" /></button></div></div>
            <p v-else class="rounded-box border border-dashed border-base-300 p-3 text-sm text-base-content/60">{{ $t("Add at least one choice.") }}</p>
            <SelectField v-if="parameter.options.length" v-model="parameter.defaultValue" :label='$t("Default choice")' :options="[{ label: 'No default choice', value: '' }, ...parameter.options.map((option) => ({ label: option, value: option }))]" />
          </div>
        </div>
        <div v-else class="min-w-0 space-y-3 rounded-box border border-base-300 bg-base-100 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div><h4 class="text-sm font-semibold">{{ $t("Choices the operator can pick") }}</h4><p class="mt-1 text-xs text-base-content/60">{{ $t("For paper sizes, add choices such as A4 and A5.") }}</p></div>
            <button class="btn btn-outline btn-sm" type="button" @click="emit('addOption')"><Plus :size="14" aria-hidden="true" />{{ $t("Add choice") }}</button>
          </div>
          <div v-if="parameter.options.length" class="grid min-w-0 gap-2 sm:grid-cols-2">
            <div v-for="(option, optionIndex) in parameter.options" :key="`${parameter.id}-${optionIndex}`" class="flex min-w-0 items-center gap-2">
              <AppInput :model-value="parameter.options[optionIndex]" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !parameter.options[optionIndex].trim() }" type="text" required :aria-label="$ui(`Choice ${optionIndex + 1}`)" :placeholder='$t("A4")' @update:model-value="updateOption(optionIndex, option, $event)" />
              <button class="btn btn-outline btn-error btn-sm shrink-0" type="button" :aria-label="$ui(`Remove choice ${optionIndex + 1}`)" @click="emit('removeOption', optionIndex)"><Trash2 :size="13" aria-hidden="true" /></button>
            </div>
          </div>
          <p v-else class="rounded-box border border-dashed border-base-300 p-3 text-sm text-base-content/60">{{ $t("Add at least one choice.") }}</p>
          <SelectField v-if="parameter.options.length" v-model="parameter.defaultValue" :label='$t("Default choice")' :options="[{ label: 'No default choice', value: '' }, ...parameter.options.map((option) => ({ label: option, value: option }))]" />
        </div>

        <details class="rounded-box border border-base-300 bg-base-100 px-3 py-2">
          <summary class="cursor-pointer text-xs font-semibold text-base-content/70">{{ $t("Advanced: internal key") }}</summary>
          <FormField class="mt-3 gap-1"><span>{{ $t("Internal key") }}</span><AppInput v-model="parameter.key" class="input w-full min-w-0" :class="{ 'input-error': props.showErrors && !parameter.key.trim() }" type="text" required :placeholder='$t("Generated from the label")' autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">{{ $t("This is generated automatically and is only used internally by pricing.") }}</small></FormField>
        </details>
      </div>
    </details>
  </article>
</template>
