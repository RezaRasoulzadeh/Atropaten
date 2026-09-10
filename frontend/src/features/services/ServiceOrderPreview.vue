<script setup lang="ts">
import { Eye } from 'lucide-vue-next'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ParameterForm, ServiceForm } from './types'
import ServiceOverviewIdentity from './ServiceOverviewIdentity.vue'
import ServiceOverviewSection from './ServiceOverviewSection.vue'

const props = defineProps<{
  form: ServiceForm
  materials: MaterialRecord[]
  machines: MachineRecord[]
  active: boolean
}>()

function valueLabel(parameter: ParameterForm) {
  if (!parameter.defaultValue) return 'Not set'
  if (parameter.type === 'boolean') return parameter.defaultValue === 'true' ? 'Yes' : 'No'
  if (parameter.type === 'material-reference') {
    const material = props.materials.find((item) => item.id === parameter.defaultValue)
    return material ? `${material.name}${material.sku ? ` · ${material.sku}` : ''}` : 'Not set'
  }
  if (parameter.type === 'machine-reference') {
    const machine = props.machines.find((item) => item.id === parameter.defaultValue)
    return machine ? `${machine.name}${machine.code ? ` · ${machine.code}` : ''}` : 'Not set'
  }
  return parameter.defaultValue
}
</script>

<template>
  <div class="min-w-0 space-y-4">
    <ServiceOverviewIdentity :form="form" :active="active" />
    <ServiceOverviewSection title="Order fields" description="Values customers will see when ordering this service.">
      <div v-if="form.parameters.length" class="divide-y divide-base-300/70">
        <div v-for="parameter in form.parameters" :key="parameter.id" class="flex min-w-0 items-start justify-between gap-3 py-2.5 first:pt-0 last:pb-0">
          <div class="min-w-0">
            <span class="block truncate text-sm">{{ parameter.label || 'Parameter' }}<em v-if="parameter.required" class="text-error"> *</em></span>
            <small v-if="parameter.unit || parameter.minValue || parameter.maxValue" class="mt-1 block text-xs leading-4 text-base-content/55">{{ parameter.unit || form.defaultUnit }}<span v-if="parameter.minValue"> · min {{ parameter.minValue }}</span><span v-if="parameter.maxValue"> · max {{ parameter.maxValue }}</span></small>
          </div>
          <span class="max-w-[58%] break-words text-end text-sm text-base-content/75">{{ valueLabel(parameter) }}</span>
        </div>
      </div>
      <p v-else class="text-sm leading-5 text-base-content/60">Choose a parameter template to preview the order form.</p>
    </ServiceOverviewSection>
  </div>
  <div class="mt-4 flex gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/70"><Eye class="mt-0.5 shrink-0 text-info" :size="16" aria-hidden="true" /><span>This is a live preview. Changes to parameters are reflected here immediately.</span></div>
</template>
