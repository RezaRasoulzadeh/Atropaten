<script setup lang="ts">
import { computed } from 'vue'
import { Eye } from 'lucide-vue-next'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ParameterForm, ServiceForm } from './types'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import ServiceOverviewIdentity from './ServiceOverviewIdentity.vue'
import ServiceOverviewSection from './ServiceOverviewSection.vue'

const props = defineProps<{
  form: ServiceForm
  materials: MaterialRecord[]
  machines: MachineRecord[]
  currencyUnit?: CurrencyUnit
  showIdentity?: boolean
  showMaterialEstimate?: boolean
  active: boolean
}>()

const materialGroups = computed(() => props.form.parameters.filter((parameter) => parameter.type === 'choice' && parameter.materialSource && !parameter.materialSource.selectMaterial))
const selectedDefaultMaterial = computed(() => {
  if (!materialGroups.value.length || materialGroups.value.some((group) => !group.defaultValue)) return null
  const values = Object.fromEntries(materialGroups.value.map((group) => [group.key, group.defaultValue]))
  const variant = props.form.materialVariants.find((item) => item.active !== false && Object.keys(values).length === Object.keys(item.values).length && Object.keys(values).every((key) => item.values[key] === values[key]))
  return props.materials.find((material) => material.id === variant?.materialId && material.active) || null
})
const selectedDefaultMaterialCost = computed(() => selectedDefaultMaterial.value ? selectedDefaultMaterial.value.highestPurchaseUnitCostRial || selectedDefaultMaterial.value.averageUnitCostRial : 0)

function valueLabel(parameter: ParameterForm) {
  if (!parameter.defaultValue) return 'Not set'
  if (parameter.materialSource && !parameter.materialSource.selectMaterial) return parameter.defaultValue.split('\u001f').join(' × ')
  if (parameter.type === 'boolean') return parameter.defaultValue === 'true' ? 'Yes' : 'No'
  if (parameter.type === 'material-reference') {
    const material = props.materials.find((item) => item.id === parameter.defaultValue)
    return material ? `${material.name}${material.sku ? ` · ${material.sku}` : ''}` : 'Not set'
  }
  if (parameter.type === 'machine-reference') {
    const machine = props.machines.find((item) => item.id === parameter.defaultValue)
    return machine ? `${machine.name}${machine.code ? ` · ${machine.code}` : ''}` : 'Not set'
  }
  if (parameter.predefinedKey) {
    const option = (parameter as any).predefinedOptions?.find((item: any) => item.code === parameter.defaultValue)
    return option?.label || 'Not set'
  }
  return parameter.defaultValue
}
</script>

<template>
  <div class="min-w-0 space-y-4">
    <ServiceOverviewIdentity v-if="showIdentity !== false" :form="form" :active="active" />
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
    <ServiceOverviewSection v-if="showMaterialEstimate !== false && materialGroups.length" title="Estimated material cost" description="Based on the selected default in each material group.">
      <div v-if="selectedDefaultMaterial" class="flex min-w-0 items-center justify-between gap-3">
        <div class="min-w-0"><strong class="block truncate text-sm">{{ selectedDefaultMaterial.name }}</strong><small class="mt-1 block truncate text-xs text-base-content/55">Default material{{ selectedDefaultMaterial.sku ? ` · ${selectedDefaultMaterial.sku}` : '' }}</small></div>
        <strong class="shrink-0 text-sm tabular-nums">{{ formatMoney(selectedDefaultMaterialCost, currencyUnit || 'Rial') }}</strong>
      </div>
      <p v-else class="rounded-box border border-dashed border-warning/40 bg-warning/5 p-3 text-xs leading-5 text-warning">Choose a default for every material group and map that combination to an inventory material to see its estimate.</p>
    </ServiceOverviewSection>
  </div>
  <div class="mt-4 flex gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/70"><Eye class="mt-0.5 shrink-0 text-info" :size="16" aria-hidden="true" /><span>This is a live preview. Changes to parameters are reflected here immediately.</span></div>
</template>
