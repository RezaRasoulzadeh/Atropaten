<script setup lang="ts">
import { ref, watch } from 'vue'
import { AlertTriangle, Calculator, LoaderCircle, RotateCcw } from 'lucide-vue-next'
import type { ServiceRecord } from '../api/services'
import type { MaterialRecord } from '../api/materials'
import { pricingApi, type PricingRecord } from '../api/pricing'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../utils/currency'
import SelectField from './SelectField.vue'

const props = defineProps<{ service: ServiceRecord; materials: MaterialRecord[]; currencyUnit: CurrencyUnit }>()
const values = ref<Record<string, string>>({})
const overrideText = ref('')
const overrideRial = ref<number | null>(null)
const manualCosts = ref<Record<string, number>>({})
const manualTexts = ref<Record<string, string>>({})
const result = ref<PricingRecord | null>(null)
const loading = ref(false)
const error = ref('')
let requestToken = 0

function resetValues() {
  const next: Record<string, string> = {}
  for (const parameter of props.service.parameters) next[parameter.key] = parameter.defaultValue || ''
  values.value = next
  overrideText.value = ''
  overrideRial.value = null
  manualCosts.value = {}
  manualTexts.value = {}
  result.value = null
  void calculate()
}

watch(() => props.service.id, resetValues, { immediate: true })
watch(values, () => { void calculate() }, { deep: true })
watch(() => props.currencyUnit, () => {
  if (overrideRial.value !== null) overrideText.value = formatMoneyInput(overrideRial.value, props.currencyUnit)
})

async function calculate() {
  const token = ++requestToken
  loading.value = true
  error.value = ''
  try {
    const next = await pricingApi.calculate({ serviceId: props.service.id, parameters: values.value, manualCosts: manualCosts.value, sellingPriceOverrideRial: overrideRial.value })
    if (token === requestToken) result.value = next
  } catch (cause) {
    if (token === requestToken) { result.value = null; error.value = cause instanceof Error ? cause.message : 'Pricing could not be calculated.' }
  } finally {
    if (token === requestToken) loading.value = false
  }
}

function updateManual(componentId: string, event: Event) {
  const value = (event.target as HTMLInputElement).value
  manualTexts.value[componentId] = value
  if (value.trim() === '') { delete manualCosts.value[componentId]; void calculate(); return }
  const parsed = parseMoneyInput(value, props.currencyUnit)
  if (parsed !== null) { manualCosts.value[componentId] = parsed; manualTexts.value[componentId] = formatMoneyInput(parsed, props.currencyUnit); void calculate() }
}

function updateOverride(event: Event) {
  const input = event.target as HTMLInputElement
  const parsed = input.value.trim() === '' ? null : parseMoneyInput(input.value, props.currencyUnit)
  if (input.value.trim() !== '' && parsed === null) { overrideText.value = input.value; return }
  overrideRial.value = parsed
  overrideText.value = parsed === null ? '' : formatMoneyInput(parsed, props.currencyUnit)
  void calculate()
}

function resetOverride() {
  overrideRial.value = null
  overrideText.value = ''
  void calculate()
}

function updateBoolean(key: string, event: Event) {
  values.value[key] = (event.target as HTMLInputElement).checked ? 'true' : 'false'
}

function money(value: number) { return formatMoney(value, props.currencyUnit) }
function signedMoney(value: number) { return `${value < 0 ? '−' : ''}${money(Math.abs(value))}` }
function typeLabel(type: string) { return ({ integer: 'Integer', decimal: 'Decimal', boolean: 'Boolean', choice: 'Choice', 'material-reference': 'Material' } as Record<string, string>)[type] ?? type }
</script>

<template>
  <section aria-label="Service pricing configurator">
    <header><div><p>Live pricing preview</p><h2><Calculator :size="17" :stroke-width="1.8" aria-hidden="true" />{{ service.name }}</h2><p>Resolve the persisted parameters and inspect the ordered cost explanation.</p></div><span v-if="loading"><LoaderCircle :size="15" :stroke-width="1.8" />Calculating</span></header>
    <div>
      <div>
        <div><h3>Parameters</h3><span>{{ service.parameters.length }} inputs</span></div>
        <div v-if="service.parameters.length">
          <label class="form-control gap-1" v-for="parameter in service.parameters" :key="parameter.id"><span>{{ parameter.label }}<em v-if="parameter.required">required</em></span>
            <input class="input input-bordered w-full min-w-0" v-if="parameter.type === 'integer' || parameter.type === 'decimal'" v-model="values[parameter.key]" type="text" inputmode="decimal" :placeholder="parameter.defaultValue || 'Enter value'" />
            <SelectField v-else-if="parameter.type === 'choice'" v-model="values[parameter.key]" :label="parameter.label" :options="[{ label: `Select ${parameter.label.toLowerCase()}`, value: '' }, ...parameter.options.map((option) => ({ label: option, value: option }))]" />
            <SelectField v-else-if="parameter.type === 'material-reference'" v-model="values[parameter.key]" :label="parameter.label" :options="[{ label: 'Select material', value: '' }, ...materials.map((material) => ({ label: `${material.name}${material.sku ? ` · ${material.sku}` : ''}`, value: material.id }))]" />
            <span v-else><input class="checkbox" :checked="values[parameter.key] === 'true'" type="checkbox" @change="updateBoolean(parameter.key, $event)" /> Enabled</span>
            <small v-if="parameter.unit || parameter.minValue || parameter.maxValue">{{ parameter.unit }}<span v-if="parameter.minValue"> · min {{ parameter.minValue }}</span><span v-if="parameter.maxValue"> · max {{ parameter.maxValue }}</span></small>
          </label>
        </div>
        <div v-else>This service has no operator parameters.</div>
      </div>
      <div>
        <div><h3>Price position</h3><span v-if="result">{{ result.marginPercentage }}% margin</span></div>
        <div v-if="error" role="alert"><AlertTriangle :size="15" :stroke-width="1.8" />{{ error }}</div>
        <div v-if="result"><div><span>Estimated cost</span><strong>{{ money(result.estimatedCostRial) }}</strong></div><div><span>Suggested price</span><strong>{{ money(result.suggestedSellingPriceRial) }}</strong></div><div><span>Effective price</span><strong>{{ money(result.effectiveSellingPriceRial) }}</strong></div><div><span>Profit</span><strong :class="{ 'text-error': result.profitRial < 0 }">{{ signedMoney(result.profitRial) }}</strong></div></div>
        <label class="form-control gap-1"><span>Selling price override <em>optional</em></span><div><input class="input input-bordered w-full min-w-0" :value="overrideText" type="text" inputmode="decimal" :placeholder="`Use ${typeLabel('fixed')} rule suggestion`" @input="updateOverride" /><button class="btn btn-ghost" v-if="overrideText" type="button" aria-label="Clear selling price override" @click="resetOverride"><RotateCcw :size="14" :stroke-width="1.8" /></button></div></label>
        <div v-if="result?.belowCost"><AlertTriangle :size="15" :stroke-width="1.8" /><span>Selling price is below estimated cost.</span></div>
        <div v-if="service.components.some((component) => component.type === 'manual')"><div><h3>Manual costs</h3><span>Optional inputs</span></div><label class="form-control gap-1" v-for="component in service.components.filter((item) => item.type === 'manual')" :key="component.id"><span>{{ component.name }}</span><input class="input input-bordered w-full min-w-0" :value="manualTexts[component.id] || ''" inputmode="decimal" placeholder="0" @input="updateManual(component.id, $event)" /></label></div>
      </div>
    </div>
    <div v-if="result"><div><h3>Ordered cost breakdown</h3><span>{{ result.components.filter((component) => component.enabled).length }} enabled components</span></div><div><table class="table table-zebra w-full"><thead><tr><th>Component</th><th>Basis</th><th>Explanation</th><th>Amount</th></tr></thead><tbody><tr v-for="component in result.components" :key="component.id" :class="{ 'opacity-50': !component.enabled }"><td><span>{{ component.name }}</span><span>{{ component.type }}</span></td><td>{{ component.enabled ? component.usageQuantity : '—' }}<span v-if="component.percentage !== '0'"> · {{ component.percentage }}%</span></td><td>{{ component.explanation }}</td><td>{{ money(component.amountRial) }}</td></tr></tbody></table></div></div>
    <div v-if="result?.warnings.length"><AlertTriangle :size="15" :stroke-width="1.8" /><span v-for="warning in result.warnings" :key="warning">{{ warning }}</span></div>
  </section>
</template>
