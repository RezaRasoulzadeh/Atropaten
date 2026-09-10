<script setup lang="ts">
import { Eye, Pencil } from 'lucide-vue-next'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import ServiceOverviewIdentity from './ServiceOverviewIdentity.vue'
import ServiceOverviewSection from './ServiceOverviewSection.vue'
import type { ParameterForm, ServiceForm } from './types'
import type { TestPricingResult, TestValues } from './serviceTestPricing'

const props = defineProps<{
  form: ServiceForm
  active: boolean
  parameters: ParameterForm[]
  values: TestValues
  materials: MaterialRecord[]
  machines: MachineRecord[]
  result: TestPricingResult | null
  currencyUnit: CurrencyUnit
}>()

defineEmits<{ edit: [] }>()

function valueLabel(parameter: ParameterForm) {
  const value = props.values[parameter.key]
  if (!value) return 'Not set'
  if (parameter.type === 'boolean') return value === 'true' ? 'Yes' : 'No'
  if (parameter.type === 'material-reference') {
    const material = props.materials.find((item) => item.id === value)
    return material ? `${material.name}${material.sku ? ` · ${material.sku}` : ''}` : 'Not set'
  }
  if (parameter.type === 'machine-reference') {
    const machine = props.machines.find((item) => item.id === value)
    return machine ? `${machine.name}${machine.code ? ` · ${machine.code}` : ''}` : 'Not set'
  }
  return value
}

function quantityLabel() {
  const parameter = props.parameters.find((item) => item.key === props.form.pricingRule.parameterKey) || props.parameters.find((item) => item.key === 'quantity')
  const value = parameter ? props.values[parameter.key] || parameter.defaultValue : ''
  return value ? `${value} ${parameter?.unit || props.form.defaultUnit}` : 'one unit'
}
</script>

<template>
  <div class="min-w-0 space-y-4">
    <ServiceOverviewIdentity :form="form" :active="active" />
    <ServiceOverviewSection title="Selected options" description="Values used for this test calculation.">
      <template #meta><button class="btn btn-ghost btn-xs gap-1" type="button" @click="$emit('edit')"><Pencil :size="13" aria-hidden="true" />Edit</button></template>
      <div v-if="parameters.length" class="divide-y divide-base-300/70">
        <div v-for="parameter in parameters" :key="parameter.id" class="flex min-w-0 items-start justify-between gap-3 py-2.5 first:pt-0 last:pb-0 text-sm"><span class="min-w-0 truncate">{{ parameter.label || 'Parameter' }}<em v-if="parameter.required" class="text-error"> *</em></span><span class="max-w-[58%] break-words text-end text-base-content/75">{{ valueLabel(parameter) }}</span></div>
      </div>
      <p v-else class="text-sm text-base-content/60">No operator parameters configured.</p>
    </ServiceOverviewSection>
    <ServiceOverviewSection title="Estimated price" description="The result for the current test values.">
      <div v-if="result" class="space-y-3">
        <div class="flex items-center justify-between gap-3"><strong class="text-xl text-success tabular-nums">{{ formatMoney(result.sellingPriceRial, currencyUnit) }}</strong><span class="badge badge-ghost text-xs">{{ result.pricingLabel }}</span></div>
        <p class="text-sm text-base-content/65">for {{ quantityLabel() }}</p>
        <div class="flex items-center justify-between gap-3 border-t border-base-300 pt-3 text-sm"><span>Total cost</span><span class="tabular-nums">{{ formatMoney(result.totalCostRial, currencyUnit) }}</span></div>
      </div>
      <p v-else class="text-sm text-base-content/60">Enter test values to preview the price.</p>
    </ServiceOverviewSection>
    <div class="flex items-start gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/65"><Eye class="mt-0.5 shrink-0 text-info" :size="15" aria-hidden="true" /><span>This is a live test preview. Change values in the form to see the result update.</span></div>
  </div>
</template>
