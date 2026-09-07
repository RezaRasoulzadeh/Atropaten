<script setup lang="ts">
import type { PrintDocumentRecord } from '../../api/reports'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { formatDate, formatDateTime } from '../../utils/date'
defineProps<{ document: PrintDocumentRecord; currencyUnit: CurrencyUnit }>()
</script>
<template>
<article class="print-document space-y-5 bg-base-100 p-5 text-sm">
<header class="flex flex-wrap justify-between gap-4 border-b border-base-300 pb-4">
<div><h2 class="text-lg font-bold">{{document.shop.shopName}}</h2><p class="text-xs text-base-content/60">{{document.shop.shopSubtitle}}</p><p class="mt-2">{{document.shop.address}} · {{document.shop.phone}}</p></div>
<div class="space-y-1 text-end"><strong class="block capitalize">{{document.kind.replaceAll('_',' ')}}</strong><span class="block">{{document.number}}</span><span class="block">{{formatDate(document.date)}}</span></div>
</header>
<div><strong>{{document.customerName || document.supplierName}}</strong><p>{{document.customerContact}}</p><p v-if="document.reference">Reference: {{document.reference}}</p></div>
<div class="overflow-x-auto"><table v-if="document.lines.length" class="table table-sm"><thead><tr><th>Description</th><th class="text-end">Quantity</th><th class="text-end">Unit price</th><th class="text-end">Total</th></tr></thead><tbody><tr v-for="(line,index) in document.lines" :key="index"><td>{{line.description}}</td><td class="text-end tabular-nums">{{line.quantityUnits/1000000}} {{line.unit}}</td><td class="text-end whitespace-nowrap tabular-nums">{{formatMoney(line.unitPriceRial,currencyUnit)}}</td><td class="text-end whitespace-nowrap tabular-nums">{{formatMoney(line.lineTotalRial,currencyUnit)}}</td></tr></tbody></table>
<table v-if="document.statementLines.length" class="table table-sm"><thead><tr><th>Date</th><th>Reference</th><th>Description</th><th class="text-end">Debit</th><th class="text-end">Credit</th><th class="text-end">Balance</th></tr></thead><tbody><tr v-for="(line,index) in document.statementLines" :key="index"><td>{{line.date ? formatDateTime(line.date) : '—'}}</td><td>{{line.reference}}</td><td>{{line.description}}</td><td class="text-end tabular-nums">{{formatMoney(line.debitRial,currencyUnit)}}</td><td class="text-end tabular-nums">{{formatMoney(line.creditRial,currencyUnit)}}</td><td class="text-end tabular-nums">{{formatMoney(line.balanceRial,currencyUnit)}}</td></tr></tbody></table></div>
<div class="flex justify-end gap-5 border-t border-base-300 pt-3"><span>Total</span><strong class="tabular-nums">{{formatMoney(document.totalRial || document.amountRial,currencyUnit)}}</strong></div>
<p class="text-xs leading-5">{{document.notes || document.shop.documentFooter}}</p>
</article>
</template>
<style scoped>
@media print {
 .print-document { background: white; color: black; padding: 0; font-size: 10pt; }
 .print-document th, .print-document td { color: black; border-bottom: 1px solid #ccc; }
 .print-document p { color: inherit; }
 .print-document tr { break-inside: avoid; }
}
</style>
