<script setup lang="ts">
import { computed } from 'vue'
import { Calculator, CircleHelp, Percent } from 'lucide-vue-next'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import type { ParameterForm, PricingRuleForm, ServiceForm } from './types'
import ServiceOverviewIdentity from './ServiceOverviewIdentity.vue'
import ServiceOverviewSection from './ServiceOverviewSection.vue'

type BreakdownItem = { name: string; amount: number; detail: string; missing: boolean }

const props = defineProps<{
  pricingRule: PricingRuleForm
  parameters: ParameterForm[]
  estimatedCostRial: number
  breakdown: BreakdownItem[]
  currencyUnit: CurrencyUnit
  form: ServiceForm
  active: boolean
}>()

const selectedQuantity = computed(() => {
  const parameter = props.parameters.find((item) => item.key === props.pricingRule.parameterKey)
  const value = Number(parameter?.defaultValue || 0)
  return Number.isFinite(value) ? value : 0
})
const markupAmount = computed(() => Math.ceil(props.estimatedCostRial * (Number(props.pricingRule.markupPercentage) || 0) / 100))
const selectedTier = computed(() => {
  let selected = props.pricingRule.tiers[0]
  for (const tier of props.pricingRule.tiers) if (Number(tier.minimumQuantity) <= selectedQuantity.value) selected = tier
  return selected
})
const sellingPrice = computed(() => {
  switch (props.pricingRule.type) {
    case 'markup': return props.estimatedCostRial + markupAmount.value
    case 'fixed-margin': return props.estimatedCostRial + props.pricingRule.fixedMarginRial
    case 'fixed': return props.pricingRule.fixedPriceRial
    case 'quantity-tiers': return selectedTier.value?.priceRial || 0
    case 'per-unit': return Math.ceil(selectedQuantity.value * props.pricingRule.perUnitRateRial)
    default: return 0
  }
})
const methodLabel = computed(() => ({ markup: 'Markup', 'fixed-margin': 'Fixed margin', fixed: 'Fixed price', 'quantity-tiers': 'Quantity tier', 'per-unit': 'Per-unit rate', manual: 'Manual price' }[props.pricingRule.type] || 'Pricing'))
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Live pricing preview">
    <ServiceOverviewIdentity :form="form" :active="active" />

    <ServiceOverviewSection title="Cost estimate" description="The estimated cost used by the selected pricing method.">
      <template #meta><Calculator :size="16" class="text-primary" aria-hidden="true" /></template>
      <div v-if="breakdown.length" class="mt-3 space-y-2">
        <div v-for="item in breakdown" :key="item.name" class="flex min-w-0 items-center gap-2 text-sm"><span class="size-2 shrink-0 rounded-full" :class="item.missing ? 'bg-warning' : 'bg-primary'"></span><span class="min-w-0 flex-1 truncate">{{ item.name }}</span><span class="shrink-0 tabular-nums" :class="item.missing ? 'text-warning' : ''">{{ item.missing ? 'Needs setup' : formatMoney(item.amount, currencyUnit) }}</span></div>
      </div>
      <div class="mt-3 flex items-center justify-between gap-3 border-t border-base-300/75 pt-2.5 text-sm"><span class="font-medium">Total cost</span><strong class="text-base tabular-nums">{{ formatMoney(estimatedCostRial, currencyUnit) }}</strong></div>
    </ServiceOverviewSection>

    <ServiceOverviewSection title="Selling price" description="The current selling-price result for this service.">
      <div class="flex items-center justify-between gap-3 text-sm"><span>{{ methodLabel }}{{ pricingRule.type === 'markup' ? ` (${pricingRule.markupPercentage || 0}%)` : '' }}</span><span v-if="pricingRule.type === 'markup'" class="tabular-nums">{{ formatMoney(markupAmount, currencyUnit) }}</span><span v-else-if="pricingRule.type === 'fixed-margin'" class="tabular-nums">{{ formatMoney(pricingRule.fixedMarginRial, currencyUnit) }}</span></div>
      <div class="mt-3 flex items-center justify-between gap-3 border-t border-base-300/75 pt-2.5"><span class="font-semibold">Selling price</span><strong class="text-lg text-success tabular-nums">{{ pricingRule.type === 'manual' ? 'Set in order' : formatMoney(sellingPrice, currencyUnit) }}</strong></div>
    </ServiceOverviewSection>

    <ServiceOverviewSection title="Pricing method" description="How the selling price is calculated.">
      <div class="flex items-start gap-3 text-sm"><span class="grid size-8 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Percent :size="16" aria-hidden="true" /></span><p class="text-xs leading-5 text-base-content/60">{{ pricingRule.type === 'manual' ? 'Operators choose the final price for each order.' : `The ${methodLabel.toLowerCase()} is applied after the estimated cost is calculated.` }}</p></div>
    </ServiceOverviewSection>
    <div class="flex items-start gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/65"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="15" aria-hidden="true" /><span>Try different parameter values in the Test step to see how the price changes.</span></div>
  </section>
</template>
