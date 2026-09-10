<script setup lang="ts">
import { Eye, Printer } from 'lucide-vue-next'
import type { MaterialRecord } from '../../api/materials'
import type { ParameterForm, ServiceForm } from './types'

const props = defineProps<{
  form: ServiceForm
  materials: MaterialRecord[]
}>()

function valueLabel(parameter: ParameterForm) {
  if (!parameter.defaultValue) return 'Not set'
  if (parameter.type === 'boolean') return parameter.defaultValue === 'true' ? 'Yes' : 'No'
  if (parameter.type === 'material-reference') {
    const material = props.materials.find((item) => item.id === parameter.defaultValue)
    return material ? `${material.name}${material.sku ? ` · ${material.sku}` : ''}` : 'Not set'
  }
  return parameter.defaultValue
}
</script>

<template>
  <div class="min-w-0 space-y-4">
    <div class="relative flex min-h-32 overflow-hidden rounded-box bg-base-300 bg-cover bg-center" :style="form.imagePath ? { backgroundImage: `url('${form.imagePath}')` } : undefined">
      <div class="absolute inset-0 bg-gradient-to-t from-black/85 via-black/45 to-black/5" aria-hidden="true"></div>
      <div class="relative z-10 mt-auto flex min-w-0 items-end gap-3 p-3">
        <div v-if="!form.imagePath" class="grid size-10 shrink-0 place-items-center rounded-box bg-black/25 text-white/75"><Printer :size="21" aria-hidden="true" /></div>
        <div class="min-w-0 text-white"><strong class="block truncate text-sm">{{ form.name || 'Your service name' }}</strong><span class="mt-1 block truncate text-xs text-white/70">{{ form.code || 'SVC-001' }}<span v-if="form.category"> · {{ form.category }}</span></span></div>
      </div>
    </div>
    <h3 class="text-base font-semibold">Configured parameters</h3>
    <div v-if="form.parameters.length" class="space-y-3">
      <div v-for="parameter in form.parameters" :key="parameter.id" class="flex min-w-0 items-start justify-between gap-3 border-b border-base-300 pb-3 last:border-0 last:pb-0"><div class="min-w-0"><span class="block truncate text-sm">{{ parameter.label || 'Parameter' }}<em v-if="parameter.required" class="text-error"> *</em></span><small v-if="parameter.unit || parameter.minValue || parameter.maxValue" class="mt-1 block text-xs leading-4 text-base-content/55">{{ parameter.unit || form.defaultUnit }}<span v-if="parameter.minValue"> · min {{ parameter.minValue }}</span><span v-if="parameter.maxValue"> · max {{ parameter.maxValue }}</span></small></div><span class="max-w-[58%] break-words text-end text-sm text-base-content/75">{{ valueLabel(parameter) }}</span></div>
    </div>
    <div v-else class="text-sm text-base-content/60">Choose a parameter template to preview the order form.</div>
  </div>
  <div class="mt-4 flex gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/70"><Eye class="mt-0.5 shrink-0 text-info" :size="16" aria-hidden="true" /><span>This is a live preview. Changes to parameters are reflected here immediately.</span></div>
</template>
