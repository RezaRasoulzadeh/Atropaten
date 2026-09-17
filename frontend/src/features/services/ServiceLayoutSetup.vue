<script setup lang="ts">
import { computed } from 'vue'
import SelectField from '../../components/ui/SelectField.vue'
import AppInput from '../../components/ui/AppInput.vue'
import type { ServiceForm } from './types'
import { isRollService } from './rollSizeInputs'
const props = defineProps<{ form: ServiceForm }>()
const enabled = computed(() => Boolean(props.form.finishedSize?.quantityParameterKey))
const automaticRoll = computed(() => isRollService(props.form) && props.form.finishedSize?.allowCustom)
function parameter(key: string, label: string, unit: string, value: string) {
  let p = props.form.parameters.find(p => p.key === key)
  if (!p) {
    p = { id: crypto.randomUUID(), key, label, type: key === 'layout_quantity' ? 'integer' : 'decimal', required: true, defaultValue: value, options: [], minValue: key === 'layout_gap_mm' || key === 'layout_margin_mm' ? '0' : key === 'layout_quantity' ? '1' : '0.001', maxValue: null, unit }
    props.form.parameters.push(p)
  }
  p.required = true
  return p
}
function enable() {
  parameter('layout_quantity', 'Finished pieces', 'piece', '1')
  parameter('layout_gap_mm', 'Space between pieces (mm)', 'mm', '0')
  parameter('layout_margin_mm', 'Edge margin (mm)', 'mm', '0')
  if (props.form.finishedSize) { props.form.finishedSize.quantityParameterKey = 'layout_quantity'; return }
  parameter('finished_width_mm', 'Finished width (mm)', 'mm', '1000')
  parameter('finished_height_mm', 'Finished height (mm)', 'mm', '3000')
  props.form.finishedSize = { parameterKey: '', quantityParameterKey: 'layout_quantity', widthParameterKey: 'finished_width_mm', heightParameterKey: 'finished_height_mm', allowCustom: true, allowRotation: true, options: [] }
}
function disable() {
  if (props.form.finishedSize) props.form.finishedSize.quantityParameterKey = ''
  for (const key of ['layout_quantity', 'layout_gap_mm', 'layout_margin_mm']) {
    const p = props.form.parameters.find(p => p.key === key)
    if (p) p.required = false
  }
  for (const component of props.form.components) {
    if (component.type === 'machine' && ['sheet','meter','square meter'].includes(component.rateBasis)) component.rateBasis = ''
  }
}
</script>
<template>
  <section class="space-y-4 rounded-box border border-primary/35 bg-primary/5 p-4" :aria-label='$t("Print layout setup")'>
    <div class="flex items-center justify-between gap-3"><div><h3 class="font-semibold">{{ $t("Print size & material layout") }}</h3><p class="mt-1 text-xs text-base-content/60">{{ $t("Custom banners, roll rotation, waste, and multiple pieces per sheet.") }}</p></div><span v-if="automaticRoll" class="badge badge-outline">{{ $t("Automatic for rolls") }}</span><input v-else class="toggle toggle-primary" type="checkbox" :aria-label='$t("Calculate material layout")' :checked="enabled" @change="($event.target as HTMLInputElement).checked ? enable() : disable()" /></div>
    <div v-if="!enabled" class="space-y-3">
      <p class="text-sm text-base-content/70">{{ $t("Enable layout to collect finished dimensions in orders and calculate material consumption for the entire quantity.") }}</p>
      <button class="btn btn-primary btn-sm" type="button" @click="enable">{{ $t("Set up roll / sheet layout") }}</button>
      <p v-if="['large format', 'banners & signage'].includes(form.category.trim().toLowerCase())" class="text-xs text-warning">{{ $t("Recommended for this category: rotation and waste need a configured print layout.") }}</p>
    </div>
    <template v-if="enabled && form.finishedSize">
      <p class="text-sm">{{ $ui(automaticRoll ? 'The exact material mapped in the grouped options above supplies the roll width. Enter custom height / length in Test and order items (1000 mm = 1 m).' : form.finishedSize.allowCustom ? 'Customers enter width and height in millimetres (1000 mm = 1 m).' : 'Uses the configured finished-size choices.') }} {{ $t("Quantity means finished pieces. Fixed prices and manual overrides apply to the entire batch.") }}</p>
      <label class="flex items-center gap-2 text-sm"><input v-model="form.finishedSize.allowRotation" class="checkbox checkbox-sm" type="checkbox" />{{ $t("Allow 90° rotation to reduce consumption") }}</label>
      <p class="text-xs text-base-content/60">{{ $t("Disable rotation for directional fabric, grain, or artwork. Layout uses a rectangular grid; include bleed in the finished dimensions or spacing.") }}</p>
      <div class="grid gap-3 sm:grid-cols-2">
        <label v-if="form.finishedSize.allowCustom && !automaticRoll" class="space-y-1 text-sm"><span>{{ $t("Default finished width (mm)") }}</span><AppInput :model-value="form.parameters.find(p => p.key === form.finishedSize?.widthParameterKey)?.defaultValue || ''" type="number" min="0.001" step="any" @update:model-value="parameter(form.finishedSize!.widthParameterKey, 'Finished width (mm)', 'mm', '').defaultValue = $event" /></label>
        <label v-if="form.finishedSize.allowCustom" class="space-y-1 text-sm"><span>{{ $t("Default finished height (mm)") }}</span><AppInput :model-value="form.parameters.find(p => p.key === form.finishedSize?.heightParameterKey)?.defaultValue || ''" type="number" min="0.001" step="any" @update:model-value="parameter(form.finishedSize!.heightParameterKey, 'Finished height (mm)', 'mm', '').defaultValue = $event" /></label>
        <label v-for="key in ['layout_gap_mm', 'layout_margin_mm']" :key="key" class="space-y-1 text-sm"><span>{{ $ui(key === 'layout_gap_mm' ? 'Default space between pieces (mm)' : 'Default edge margin (mm)') }}</span><AppInput :model-value="form.parameters.find(p => p.key === key)?.defaultValue || '0'" type="number" min="0" max="1000" @update:model-value="parameter(key, key === 'layout_gap_mm' ? 'Space between pieces (mm)' : 'Edge margin (mm)', 'mm', '0').defaultValue = $event" /></label>
      </div>
      <SelectField v-for="component in form.components.filter(c => c.type === 'machine')" :key="component.id" v-model="component.rateBasis" :label="$ui(`${component.name || $ui('Machine')} — apply selected rate per`)" :options="[{label: 'Configured usage / finished piece', value: ''}, {label: 'Consumed sheet',value:'sheet'}, {label: 'Consumed roll metre',value:'meter'}, {label: 'Consumed square metre (includes waste)',value:'square meter'}]" />
      <p class="text-xs text-base-content/60">{{ $t("Use the selected machine profile’s amount for the basis above. A machine profile set to Per meter or Per square meter automatically uses this layout when no override is selected. Time-based profiles should use configured usage. Material stock must use sheet/piece units for sheets, or metre/square metre units for rolls.") }}</p>
    </template>
  </section>
</template>
