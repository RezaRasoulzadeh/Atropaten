<script setup lang="ts">
import type { ReportRecord } from '../../api/reports'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'

const props = defineProps<{ report: ReportRecord; title: string; currencyUnit: CurrencyUnit }>()

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}

function quantity(value: number) {
  return value ? String(value / 1000000) : '0'
}
</script>

<template>
  <article class="report-print-document space-y-5 bg-white p-6 text-sm text-black">
    <header class="flex flex-wrap items-start justify-between gap-5 border-b border-slate-300 pb-4">
      <div>
        <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Atropaten · Reports</p>
        <h1 class="mt-1 text-2xl font-bold">{{ title }}</h1>
        <p class="mt-1 text-xs text-slate-600">{{ report.startDate }} → {{ report.endDate }}</p>
      </div>
      <div class="text-end text-xs text-slate-600">
        <p>Generated {{ formatDateTime(new Date().toISOString()) }}</p>
        <p>{{ report.rows.length }} rows · {{ report.summaries.length }} summary values</p>
      </div>
    </header>

    <section v-if="report.summaries.length" class="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <div v-for="item in report.summaries" :key="item.key" class="rounded border border-slate-300 p-3">
        <p class="text-xs text-slate-600">{{ item.label }}</p>
        <strong class="mt-1 block tabular-nums">{{ money(item.amountRial) }}</strong>
        <span v-if="item.count" class="mt-1 block text-xs text-slate-500">{{ item.count }} records</span>
      </div>
    </section>

    <table class="w-full border-collapse text-xs">
      <thead>
        <tr class="border-b-2 border-slate-400 text-start">
          <th class="px-2 py-2 text-start">Source / name</th>
          <th class="px-2 py-2 text-start">Category</th>
          <th class="px-2 py-2 text-start">Date</th>
          <th class="px-2 py-2 text-start">Status</th>
          <th class="px-2 py-2 text-end">Quantity</th>
          <th class="px-2 py-2 text-end">Amount</th>
          <th class="px-2 py-2 text-end">Secondary</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in report.rows" :key="`${row.id}-${row.name}`" class="border-b border-slate-200 align-top">
          <td class="px-2 py-2"><strong>{{ row.name || row.id }}</strong><span v-if="row.secondaryName" class="block text-slate-600">{{ row.secondaryName }}</span><span v-if="row.referenceId" class="block text-slate-500">{{ row.referenceId }}</span></td>
          <td class="px-2 py-2">{{ row.category || '—' }}</td>
          <td class="whitespace-nowrap px-2 py-2">{{ row.date ? formatDateTime(row.date) : '—' }}</td>
          <td class="px-2 py-2">{{ row.status || '—' }}</td>
          <td class="whitespace-nowrap px-2 py-2 text-end tabular-nums">{{ row.quantityUnits ? quantity(row.quantityUnits) : '—' }}<span v-if="row.secondaryQuantityUnits"> / {{ quantity(row.secondaryQuantityUnits) }}</span></td>
          <td class="whitespace-nowrap px-2 py-2 text-end tabular-nums">{{ money(row.amountRial) }}</td>
          <td class="whitespace-nowrap px-2 py-2 text-end tabular-nums">{{ row.secondaryAmountRial ? money(row.secondaryAmountRial) : '—' }}</td>
        </tr>
        <tr v-if="!report.rows.length"><td colspan="7" class="px-2 py-5 text-center text-slate-500">No records in this period.</td></tr>
      </tbody>
    </table>
  </article>
</template>

<style scoped>
@media print {
  .report-print-document {
    min-height: 100vh;
    padding: 0;
    color: #000;
    background: #fff;
    font-size: 9pt;
  }

  .report-print-document tr {
    break-inside: avoid;
  }
}
</style>
