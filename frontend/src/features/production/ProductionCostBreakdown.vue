<script setup lang="ts">
import type { ProductionJobRecord } from '../../api/production'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'

const props = defineProps<{
  job: ProductionJobRecord
  currencyUnit: CurrencyUnit
  embedded?: boolean
}>()

function money(value: number) {
  return formatMoney(value || 0, props.currencyUnit)
}
</script>

<template>
  <section class="min-w-0" :class="embedded ? '' : 'rounded-box border border-base-300 bg-base-100/45 p-4'" aria-label="Production cost breakdown">
    <h3 class="text-sm font-semibold">Cost</h3>
    <dl class="mt-3 border-b border-base-300 pb-3 text-sm">
      <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
        <dt class="text-xs text-base-content/60">Original order estimate</dt>
        <dd class="whitespace-nowrap tabular-nums">{{ money(job.estimatedCostRial) }}</dd>
      </div>
    </dl>
    <dl class="mt-1 divide-y divide-base-300/70 text-sm">
      <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 py-2">
        <dt class="text-xs text-base-content/60">Consumed materials</dt>
        <dd class="whitespace-nowrap tabular-nums">{{ money(job.actualMaterialCostRial) }}</dd>
      </div>
      <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 py-2">
        <dt class="text-xs text-base-content/60">Material waste</dt>
        <dd class="whitespace-nowrap tabular-nums">{{ money(job.actualWasteCostRial) }}</dd>
      </div>
      <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 py-2">
        <dt class="text-xs text-base-content/60">
          Outsourcing
          <span v-if="Number(job.outsourceQuantity) > 0" class="mt-0.5 block text-base-content/45">
            {{ job.outsourceQuantity }} {{ job.quantityUnit }} × {{ money(job.outsourceUnitCostRial) }} / {{ job.quantityUnit }}
          </span>
        </dt>
        <dd class="whitespace-nowrap tabular-nums">{{ money(job.actualOutsourcedCostRial) }}</dd>
      </div>
      <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 py-2">
        <dt class="text-xs text-base-content/60">Remaining materials estimate</dt>
        <dd class="whitespace-nowrap tabular-nums">{{ money(job.remainingMaterialCostRial) }}</dd>
      </div>
      <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 py-2">
        <dt class="text-xs text-base-content/60">Machine, labor &amp; overhead estimate</dt>
        <dd class="whitespace-nowrap tabular-nums">{{ money(job.estimatedConversionCostRial) }}</dd>
      </div>
      <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 py-2 font-semibold">
        <dt class="text-xs">Expected total cost</dt>
        <dd class="whitespace-nowrap tabular-nums">{{ money(job.projectedCostRial) }}</dd>
      </div>
    </dl>
    <p class="mt-3 text-xs leading-5 text-base-content/55">
      Total includes the five components above; the original estimate is for comparison only.
      Material costs are net of returns and corrections. Machine, labor and overhead remain estimates; business profit uses posted expenses.
    </p>
  </section>
</template>
