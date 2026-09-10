<script setup lang="ts">
import { computed } from 'vue'
import { BarChart3, Calculator, CircleHelp, List, Percent, Plus, Tag, Trash2 } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../../utils/currency'
import type { ParameterForm, PricingRuleForm } from './types'

const props = defineProps<{
  pricingRule: PricingRuleForm
  parameters: ParameterForm[]
  currencyUnit: CurrencyUnit
  estimatedCostRial: number
  showErrors?: boolean
}>()

const numericParameters = computed(() => props.parameters.filter((parameter) => parameter.type === 'integer' || parameter.type === 'decimal'))
const methods = [
  { type: 'markup', title: 'Cost + markup', description: 'Add a percentage on top of the total cost.', icon: Percent },
  { type: 'fixed-margin', title: 'Cost + fixed margin', description: 'Add a fixed amount to the total cost.', icon: Calculator },
  { type: 'fixed', title: 'Fixed price', description: 'Use the same price regardless of cost.', icon: Tag },
  { type: 'quantity-tiers', title: 'Quantity tiers', description: 'Set different prices based on quantity.', icon: BarChart3 },
  { type: 'per-unit', title: 'Per-unit parameter', description: 'Multiply a rate by a numeric parameter.', icon: Calculator },
  { type: 'manual', title: 'Manual price', description: 'Set the price manually in each order.', icon: List },
]

function setType(type: string) {
  props.pricingRule.type = type
  if (type !== 'per-unit' && type !== 'quantity-tiers') props.pricingRule.parameterKey = ''
  if (type !== 'quantity-tiers') props.pricingRule.tiers = []
}

function updateMoney(field: 'fixedPriceInput' | 'fixedMarginInput' | 'perUnitRateInput', value: string) {
  props.pricingRule[field] = value
  const valueField = field === 'fixedPriceInput' ? 'fixedPriceRial' : field === 'fixedMarginInput' ? 'fixedMarginRial' : 'perUnitRateRial'
  const parsed = parseMoneyInput(value, props.currencyUnit)
  if (parsed !== null) {
    props.pricingRule[valueField] = parsed
    props.pricingRule[field] = formatMoneyInput(parsed, props.currencyUnit)
  }
}

function addTier() {
  const last = props.pricingRule.tiers.at(-1)
  props.pricingRule.tiers.push({
    position: props.pricingRule.tiers.length,
    minimumQuantity: last ? String(Math.max(1, Number(last.minimumQuantity) + 10)) : '0',
    priceRial: 0,
    priceInput: formatMoneyInput(0, props.currencyUnit),
  })
}

function removeTier(index: number) {
  props.pricingRule.tiers.splice(index, 1)
  props.pricingRule.tiers.forEach((tier, position) => { tier.position = position })
}

function updateTierPrice(index: number, value: string) {
  const tier = props.pricingRule.tiers[index]
  if (!tier) return
  tier.priceInput = value
  const parsed = parseMoneyInput(value, props.currencyUnit)
  if (parsed !== null) {
    tier.priceRial = parsed
    tier.priceInput = formatMoneyInput(parsed, props.currencyUnit)
  }
}
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Service pricing">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold">Pricing method</h2>
        <p class="mt-1 text-sm text-base-content/65">Choose how the selling price is calculated for this service.</p>
      </div>
      <div class="flex items-center gap-2 rounded-box border border-base-300 bg-base-100 px-3 py-2 text-xs text-base-content/65">
        <span>Estimated cost</span><strong class="text-sm text-base-content">{{ formatMoney(estimatedCostRial, currencyUnit) }}</strong>
      </div>
    </div>

    <div class="grid min-w-0 gap-2 sm:grid-cols-2 xl:grid-cols-5">
      <button v-for="method in methods" :key="method.type" class="min-w-0 rounded-box border p-3 text-start transition-colors" :class="pricingRule.type === method.type ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-100 hover:border-primary/50'" type="button" @click="setType(method.type)">
        <div class="flex items-center justify-between gap-2"><span class="grid size-9 place-items-center rounded-box bg-base-200 text-primary"><component :is="method.icon" :size="20" aria-hidden="true" /></span><span class="size-4 rounded-full border-2" :class="pricingRule.type === method.type ? 'border-primary bg-primary' : 'border-base-content/40'"></span></div>
        <strong class="mt-3 block text-sm">{{ method.title }}</strong>
        <small class="mt-1 block text-xs leading-4 text-base-content/60">{{ method.description }}</small>
      </button>
    </div>

    <div v-if="pricingRule.type === 'markup'" class="rounded-box border border-base-300 bg-base-100 p-4">
      <h3 class="text-base font-semibold">Cost + markup settings</h3><p class="mt-1 text-sm text-base-content/65">The selling price is calculated by adding a percentage to the total cost.</p>
      <FormGrid class="mt-4">
        <FormField class="gap-1"><span>Markup percentage <em class="text-error">*</em></span><div class="join w-full"><AppInput v-model="pricingRule.markupPercentage" class="input join-item w-full min-w-0" :class="{ 'input-error': showErrors && !pricingRule.markupPercentage.trim() }" type="text" inputmode="decimal" placeholder="30" /><span class="join-item grid w-12 place-items-center border border-base-300 bg-base-200 text-sm">%</span></div><small class="text-xs leading-5 text-base-content/60">For example, 30% markup on 50,000 {{ currencyUnit }} adds 15,000 {{ currencyUnit }}.</small></FormField>
        <div class="flex items-end text-sm text-base-content/70">Estimated selling price: <strong class="ml-1 text-primary">{{ formatMoney(Math.ceil(estimatedCostRial * (1 + (Number(pricingRule.markupPercentage) || 0) / 100)), currencyUnit) }}</strong></div>
      </FormGrid>
    </div>

    <div v-else-if="pricingRule.type === 'fixed-margin'" class="rounded-box border border-base-300 bg-base-100 p-4">
      <h3 class="text-base font-semibold">Fixed margin settings</h3><p class="mt-1 text-sm text-base-content/65">Add a fixed amount to the estimated cost for every service unit.</p>
      <FormField class="mt-4 max-w-md gap-1"><span>Fixed margin ({{ currencyUnit }}) <em class="text-error">*</em></span><AppInput :model-value="pricingRule.fixedMarginInput" class="input w-full" :class="{ 'input-error': showErrors && !pricingRule.fixedMarginInput.trim() }" :money="currencyUnit" type="text" inputmode="numeric" placeholder="15,000" @update:model-value="updateMoney('fixedMarginInput', $event)" /></FormField>
    </div>

    <div v-else-if="pricingRule.type === 'fixed'" class="rounded-box border border-base-300 bg-base-100 p-4">
      <h3 class="text-base font-semibold">Fixed price settings</h3><p class="mt-1 text-sm text-base-content/65">Set the selling price charged for one service unit. It is compared with the estimated cost, but it does not add to or replace the cost calculation.</p>
      <FormField class="mt-4 max-w-md gap-1"><span>Selling price per service unit ({{ currencyUnit }}) <em class="text-error">*</em></span><AppInput :model-value="pricingRule.fixedPriceInput" class="input w-full" :class="{ 'input-error': showErrors && !pricingRule.fixedPriceInput.trim() }" :money="currencyUnit" type="text" inputmode="numeric" placeholder="87,800" @update:model-value="updateMoney('fixedPriceInput', $event)" /></FormField>
    </div>

    <div v-else-if="pricingRule.type === 'quantity-tiers'" class="rounded-box border border-base-300 bg-base-100 p-4">
      <div class="flex flex-wrap items-start justify-between gap-3"><div><h3 class="text-base font-semibold">Quantity tier settings</h3><p class="mt-1 text-sm text-base-content/65">Select a quantity input and set the price that applies from each minimum quantity.</p></div><button class="btn btn-outline btn-sm gap-2" type="button" @click="addTier"><Plus :size="14" aria-hidden="true" />Add tier</button></div>
      <SelectField v-model="pricingRule.parameterKey" class="mt-4 max-w-md" label="Quantity parameter" :invalid="showErrors && !pricingRule.parameterKey" :options="[{ label: 'Select a numeric parameter', value: '' }, ...numericParameters.map((parameter) => ({ label: `${parameter.label || parameter.key}${parameter.unit ? ` · ${parameter.unit}` : ''}`, value: parameter.key }))]" />
      <div v-if="pricingRule.tiers.length" class="mt-4 space-y-2">
        <div v-for="(tier, index) in pricingRule.tiers" :key="tier.position" class="grid min-w-0 items-end gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]"><FormField class="gap-1"><span>From quantity</span><AppInput v-model="tier.minimumQuantity" class="input w-full" inputmode="numeric" placeholder="0" /></FormField><FormField class="gap-1"><span>Price ({{ currencyUnit }})</span><AppInput :model-value="tier.priceInput" class="input w-full" :money="currencyUnit" inputmode="numeric" placeholder="0" @update:model-value="updateTierPrice(index, $event)" /></FormField><button class="btn btn-outline btn-error btn-square" type="button" aria-label="Remove tier" @click="removeTier(index)"><Trash2 :size="14" aria-hidden="true" /></button></div>
      </div>
      <p v-else class="mt-4 rounded-box border border-dashed border-base-300 p-3 text-sm text-base-content/60">Add a zero-quantity tier before saving.</p>
    </div>

    <div v-else-if="pricingRule.type === 'per-unit'" class="rounded-box border border-base-300 bg-base-100 p-4">
      <h3 class="text-base font-semibold">Per-unit parameter settings</h3><p class="mt-1 text-sm text-base-content/65">Multiply a price rate by a numeric parameter such as quantity, hours, or finished area.</p>
      <FormGrid class="mt-4"><SelectField v-model="pricingRule.parameterKey" label="Numeric parameter" :invalid="showErrors && !pricingRule.parameterKey" :options="[{ label: 'Select a numeric parameter', value: '' }, ...numericParameters.map((parameter) => ({ label: `${parameter.label || parameter.key}${parameter.unit ? ` · ${parameter.unit}` : ''}`, value: parameter.key }))]" /><FormField class="gap-1"><span>Rate per unit ({{ currencyUnit }}) <em class="text-error">*</em></span><AppInput :model-value="pricingRule.perUnitRateInput" class="input w-full" :class="{ 'input-error': showErrors && !pricingRule.perUnitRateInput.trim() }" :money="currencyUnit" type="text" inputmode="numeric" placeholder="1,000" @update:model-value="updateMoney('perUnitRateInput', $event)" /></FormField></FormGrid>
    </div>

    <div v-else class="rounded-box border border-base-300 bg-base-100 p-4"><h3 class="text-base font-semibold">Manual price</h3><p class="mt-1 text-sm leading-6 text-base-content/65">The operator will enter the selling price when adding this service to an order. The estimated cost remains available for comparison.</p></div>

    <details class="rounded-box border border-base-300 bg-base-100 px-4 py-3">
      <summary class="flex cursor-pointer list-none items-center gap-2 text-sm font-semibold [&::-webkit-details-marker]:hidden"><CircleHelp :size="17" class="text-primary" aria-hidden="true" />Advanced settings</summary>
      <p class="mt-3 text-sm leading-6 text-base-content/65">Minimum price, maximum price, and custom rounding rules can be added here as the pricing model grows. The selected pricing method and its values are saved now.</p>
    </details>

    <div class="flex items-start gap-2 rounded-box border border-info/20 bg-info/5 p-3 text-sm leading-6"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p><strong class="font-semibold text-info">How it works</strong><br />The total cost from your cost components is calculated first, then the selected pricing method is applied to suggest a selling price.</p></div>
  </section>
</template>
