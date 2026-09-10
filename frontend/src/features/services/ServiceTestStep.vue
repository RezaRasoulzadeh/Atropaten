<script setup lang="ts">
import { computed, watch } from 'vue'
import { Calculator, CheckCircle2, FlaskConical, RotateCcw, TriangleAlert } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { calculateServiceTest, type TestPricingResult, type TestValues } from './serviceTestPricing'
import type { ParameterForm, ServiceForm } from './types'

const props = defineProps<{
  form: ServiceForm
  parameters: ParameterForm[]
  materials: MaterialRecord[]
  machines: MachineRecord[]
  services: ServiceRecord[]
  values: TestValues
  currencyUnit: CurrencyUnit
}>()

const emit = defineEmits<{
  'update:result': [result: TestPricingResult]
}>()

function defaultValue(parameter: ParameterForm) {
  if (parameter.defaultValue) return parameter.defaultValue
  if (parameter.type === 'boolean') return 'false'
  return ''
}

function syncValues() {
  const keys = new Set(props.parameters.map((parameter) => parameter.key).filter(Boolean))
  for (const parameter of props.parameters) {
    if (parameter.key && props.values[parameter.key] === undefined) props.values[parameter.key] = defaultValue(parameter)
  }
  for (const key of Object.keys(props.values)) if (!keys.has(key)) delete props.values[key]
}

function recalculate() {
  syncValues()
  emit('update:result', calculateServiceTest(props.form, props.values, props.materials, props.machines, props.services))
}

function reset() {
  for (const parameter of props.parameters) if (parameter.key) props.values[parameter.key] = defaultValue(parameter)
  recalculate()
}

function updateBoolean(key: string, event: Event) {
  props.values[key] = (event.target as HTMLInputElement).checked ? 'true' : 'false'
  recalculate()
}

function valueOptions(parameter: ParameterForm) {
  if (parameter.type === 'choice') return [{ label: `Select ${parameter.label.toLowerCase()}`, value: '' }, ...parameter.options.map((value) => ({ label: value, value }))]
  if (parameter.type === 'material-reference') return [{ label: 'Select material', value: '' }, ...props.materials.filter((item) => item.active).map((item) => ({ label: `${item.name}${item.sku ? ` · ${item.sku}` : ''}`, value: item.id }))]
  return [{ label: 'Select machine', value: '' }, ...props.machines.filter((item) => item.active).map((item) => ({ label: `${item.name}${item.code ? ` · ${item.code}` : ''}`, value: item.id }))]
}

function onValueChanged() {
  recalculate()
}

watch(
  () => [props.parameters, props.form.components, props.form.pricingRule, props.materials, props.machines, props.services, props.values],
  recalculate,
  { deep: true, immediate: true },
)

const result = computed<TestPricingResult>(() => calculateServiceTest(props.form, props.values, props.materials, props.machines, props.services))
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Test service">
    <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1.05fr)_minmax(18rem,0.95fr)]">
      <div class="min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-5">
        <div class="flex min-w-0 items-start justify-between gap-3 border-b border-base-300 pb-4">
          <div class="flex min-w-0 items-start gap-3">
            <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FlaskConical :size="21" aria-hidden="true" /></span>
            <div class="min-w-0">
              <h2 class="text-lg font-semibold">Test service</h2>
              <p class="mt-1 text-sm leading-5 text-base-content/60">Configure the service like a customer would. See the calculated cost and selling price.</p>
            </div>
          </div>
          <button class="btn btn-outline btn-sm shrink-0 gap-2" type="button" @click="reset"><RotateCcw :size="14" aria-hidden="true" />Reset</button>
        </div>

        <div v-if="parameters.length" class="mt-4 space-y-3">
          <div v-for="parameter in parameters" :key="parameter.id" class="grid min-w-0 grid-cols-[2rem_minmax(0,1fr)] items-start gap-3">
            <span class="grid size-8 place-items-center rounded-box bg-base-300/60 text-base-content/70"><span v-if="parameter.type === 'integer' || parameter.type === 'decimal'" class="text-lg">#</span><span v-else-if="parameter.type === 'boolean'" class="text-sm">✓</span><span v-else class="text-base">◈</span></span>
            <FormField class="min-w-0 gap-1">
              <span class="text-sm text-base-content">{{ parameter.label || 'Parameter' }}<em v-if="parameter.required" class="text-error"> *</em></span>
              <AppInput v-if="parameter.type === 'integer' || parameter.type === 'decimal'" v-model="values[parameter.key]" class="input w-full min-w-0" :type="parameter.type === 'integer' ? 'number' : 'text'" :step="parameter.type === 'integer' ? '1' : 'any'" :min="parameter.minValue || undefined" :max="parameter.maxValue || undefined" inputmode="decimal" @update:model-value="onValueChanged" />
              <SelectField v-else-if="parameter.type === 'choice' || parameter.type === 'material-reference' || parameter.type === 'machine-reference'" v-model="values[parameter.key]" :aria-label="parameter.label" :options="valueOptions(parameter)" @update:model-value="onValueChanged" />
              <label v-else class="flex h-10 items-center gap-2 rounded-box border border-base-300 bg-base-100 px-3 text-sm"><input class="checkbox checkbox-sm" type="checkbox" :checked="values[parameter.key] === 'true'" @change="updateBoolean(parameter.key, $event)" />Enabled</label>
              <small v-if="parameter.unit || parameter.minValue || parameter.maxValue" class="text-xs leading-5 text-base-content/55">{{ parameter.unit || form.defaultUnit }}<span v-if="parameter.minValue"> · min {{ parameter.minValue }}</span><span v-if="parameter.maxValue"> · max {{ parameter.maxValue }}</span></small>
            </FormField>
          </div>
        </div>
        <div v-else class="mt-5 rounded-box border border-dashed border-base-300 p-6 text-center text-sm text-base-content/60">This service has no customer parameters. The test uses its configured cost and pricing rules.</div>
      </div>

      <div class="min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-5">
        <div class="flex items-start gap-3 border-b border-base-300 pb-4">
          <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Calculator :size="21" aria-hidden="true" /></span>
          <div><h2 class="text-lg font-semibold">Calculation result</h2><p class="mt-1 text-sm leading-5 text-base-content/60">Real-time calculation based on your input.</p></div>
        </div>
        <div class="mt-4 space-y-4">
          <div class="rounded-box border border-base-300 bg-base-100/30 p-3">
            <h3 class="text-sm font-semibold">Cost breakdown</h3>
            <div v-if="result.lines.length" class="mt-3 divide-y divide-base-300/70">
              <div v-for="line in result.lines" :key="`${line.name}-${line.detail}`" class="flex min-w-0 items-center gap-2 py-2 first:pt-0 last:pb-0 text-sm"><span class="size-2 shrink-0 rounded-full" :class="line.missing ? 'bg-warning' : 'bg-primary'"></span><span class="min-w-0 flex-1"><span class="block truncate">{{ line.name }}</span><small class="block truncate text-xs text-base-content/55">{{ line.detail }}</small></span><span class="shrink-0 tabular-nums" :class="line.missing ? 'text-warning' : ''">{{ line.missing ? 'Needs setup' : formatMoney(line.amount, currencyUnit) }}</span></div>
            </div>
            <p v-else class="mt-3 text-sm text-base-content/60">Add a cost component to calculate this service.</p>
            <div class="mt-3 flex items-center justify-between gap-3 border-t border-base-300 pt-3 text-sm"><span class="font-semibold">Total cost</span><strong class="tabular-nums">{{ formatMoney(result.totalCostRial, currencyUnit) }}</strong></div>
          </div>
          <div class="rounded-box border border-base-300 bg-base-100/30 p-3">
            <div class="flex items-center justify-between gap-3 text-sm"><span>{{ result.pricingLabel }}</span><span v-if="form.pricingRule.type === 'markup'" class="tabular-nums">+ {{ formatMoney(result.markupRial, currencyUnit) }}</span><span v-else-if="form.pricingRule.type === 'fixed-margin'" class="tabular-nums">+ {{ formatMoney(form.pricingRule.fixedMarginRial, currencyUnit) }}</span><span v-else-if="form.pricingRule.type === 'fixed'" class="text-base-content/60">Direct price</span></div>
            <div class="mt-3 flex items-center justify-between gap-3 rounded-box border border-success/25 px-3 py-3"><span class="font-semibold">{{ form.pricingRule.type === 'fixed' ? 'Fixed selling price' : 'Selling price' }}</span><strong class="text-lg text-success tabular-nums">{{ formatMoney(result.sellingPriceRial, currencyUnit) }}</strong></div>
          </div>
          <div class="flex items-start gap-3 rounded-box border p-3 text-sm" :class="result.belowCost ? 'border-warning/30 text-warning' : 'border-success/30 text-success'"><TriangleAlert v-if="result.belowCost" :size="18" class="mt-0.5 shrink-0" aria-hidden="true" /><CheckCircle2 v-else :size="18" class="mt-0.5 shrink-0" aria-hidden="true" /><div><strong>{{ result.belowCost ? 'Price is below cost' : 'Price is above cost' }}</strong><p class="mt-1 text-xs leading-5 text-base-content/60">{{ result.belowCost ? 'Review the pricing rule before saving this service.' : `Estimated margin: ${result.marginPercentage.toFixed(1)}%.` }}</p></div></div>
          <div class="rounded-box border border-info/25 bg-info/5 p-3 text-xs leading-5 text-base-content/70">This is a test calculation. The final amount may vary with order conditions, discounts, or customer-specific rules.</div>
        </div>
      </div>
    </div>
  </section>
</template>
