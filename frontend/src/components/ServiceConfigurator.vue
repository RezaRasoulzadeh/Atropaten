<script setup lang="ts">
import { ref, watch } from 'vue'
import { AlertTriangle, Calculator, LoaderCircle, RotateCcw } from 'lucide-vue-next'
import type { ServiceRecord } from '../api/services'
import type { MaterialRecord } from '../api/materials'
import { pricingApi, type PricingRecord } from '../api/pricing'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../utils/currency'

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
  <section class="service-configurator panel" aria-label="Service pricing configurator">
    <header class="configurator-header"><div><p class="eyebrow">Live pricing preview</p><h2><Calculator :size="17" :stroke-width="1.8" aria-hidden="true" />{{ service.name }}</h2><p>Resolve the persisted parameters and inspect the ordered cost explanation.</p></div><span v-if="loading" class="configurator-loading"><LoaderCircle :size="15" :stroke-width="1.8" class="spin" />Calculating</span></header>
    <div class="configurator-grid">
      <div class="configurator-parameters">
        <div class="configurator-section-heading"><h3>Parameters</h3><span>{{ service.parameters.length }} inputs</span></div>
        <div v-if="service.parameters.length" class="configurator-fields">
          <label v-for="parameter in service.parameters" :key="parameter.id" class="form-field"><span>{{ parameter.label }}<em v-if="parameter.required">required</em></span>
            <input v-if="parameter.type === 'integer' || parameter.type === 'decimal'" v-model="values[parameter.key]" type="text" inputmode="decimal" :placeholder="parameter.defaultValue || 'Enter value'" />
            <select v-else-if="parameter.type === 'choice'" v-model="values[parameter.key]"><option value="">Select {{ parameter.label.toLowerCase() }}</option><option v-for="option in parameter.options" :key="option" :value="option">{{ option }}</option></select>
            <select v-else-if="parameter.type === 'material-reference'" v-model="values[parameter.key]"><option value="">Select material</option><option v-for="material in materials" :key="material.id" :value="material.id">{{ material.name }}{{ material.sku ? ` · ${material.sku}` : '' }}</option></select>
            <span v-else class="checkbox-control"><input :checked="values[parameter.key] === 'true'" type="checkbox" @change="updateBoolean(parameter.key, $event)" /> Enabled</span>
            <small v-if="parameter.unit || parameter.minValue || parameter.maxValue" class="field-help">{{ parameter.unit }}<span v-if="parameter.minValue"> · min {{ parameter.minValue }}</span><span v-if="parameter.maxValue"> · max {{ parameter.maxValue }}</span></small>
          </label>
        </div>
        <div v-else class="configurator-empty">This service has no operator parameters.</div>
      </div>
      <div class="configurator-summary">
        <div class="configurator-section-heading"><h3>Price position</h3><span v-if="result">{{ result.marginPercentage }}% margin</span></div>
        <div v-if="error" class="configurator-error" role="alert"><AlertTriangle :size="15" :stroke-width="1.8" />{{ error }}</div>
        <div v-if="result" class="price-summary-grid"><div><span>Estimated cost</span><strong>{{ money(result.estimatedCostRial) }}</strong></div><div><span>Suggested price</span><strong>{{ money(result.suggestedSellingPriceRial) }}</strong></div><div class="price-effective"><span>Effective price</span><strong>{{ money(result.effectiveSellingPriceRial) }}</strong></div><div><span>Profit</span><strong :class="{ 'is-negative': result.profitRial < 0 }">{{ signedMoney(result.profitRial) }}</strong></div></div>
        <label class="form-field override-field"><span>Selling price override <em>optional</em></span><div class="input-with-action"><input :value="overrideText" type="text" inputmode="decimal" :placeholder="`Use ${typeLabel('fixed')} rule suggestion`" @input="updateOverride" /><button v-if="overrideText" class="icon-button icon-button-small" type="button" aria-label="Clear selling price override" @click="resetOverride"><RotateCcw :size="14" :stroke-width="1.8" /></button></div></label>
        <div v-if="result?.belowCost" class="below-cost-warning"><AlertTriangle :size="15" :stroke-width="1.8" /><span>Selling price is below estimated cost.</span></div>
        <div v-if="service.components.some((component) => component.type === 'manual')" class="manual-costs"><div class="configurator-section-heading"><h3>Manual costs</h3><span>Optional inputs</span></div><label v-for="component in service.components.filter((item) => item.type === 'manual')" :key="component.id" class="form-field"><span>{{ component.name }}</span><input :value="manualTexts[component.id] || ''" inputmode="decimal" placeholder="0" @input="updateManual(component.id, $event)" /></label></div>
      </div>
    </div>
    <div v-if="result" class="configurator-breakdown"><div class="configurator-section-heading"><h3>Ordered cost breakdown</h3><span>{{ result.components.filter((component) => component.enabled).length }} enabled components</span></div><div class="table-wrap"><table class="data-table pricing-breakdown-table"><thead><tr><th>Component</th><th>Basis</th><th>Explanation</th><th class="numeric-column">Amount</th></tr></thead><tbody><tr v-for="component in result.components" :key="component.id" :class="{ 'is-disabled': !component.enabled }"><td><span class="table-primary">{{ component.name }}</span><span class="table-secondary">{{ component.type }}</span></td><td>{{ component.enabled ? component.usageQuantity : '—' }}<span v-if="component.percentage !== '0'" class="table-secondary"> · {{ component.percentage }}%</span></td><td class="table-secondary">{{ component.explanation }}</td><td class="numeric-column table-money">{{ money(component.amountRial) }}</td></tr></tbody></table></div></div>
    <div v-if="result?.warnings.length" class="configurator-warnings"><AlertTriangle :size="15" :stroke-width="1.8" /><span v-for="warning in result.warnings" :key="warning">{{ warning }}</span></div>
  </section>
</template>
