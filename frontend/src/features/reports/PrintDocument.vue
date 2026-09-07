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
<div class="grid gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(14rem,auto)]"><div><strong>{{document.customerName || document.supplierName}}</strong><p>{{document.customerContact}}</p><p v-if="document.reference">Reference: {{document.reference}}</p></div><dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-4 gap-y-1 text-end text-xs"><template v-if="document.status"><dt class="text-base-content/60">Status</dt><dd>{{document.status}}</dd></template><template v-if="document.dueDate"><dt class="text-base-content/60">Due</dt><dd>{{formatDate(document.dueDate)}}</dd></template><template v-if="document.method"><dt class="text-base-content/60">Method</dt><dd>{{document.method}}</dd></template><template v-if="document.accountName"><dt class="text-base-content/60">Account</dt><dd>{{document.accountName}}</dd></template><template v-if="document.paymentStatus"><dt class="text-base-content/60">Payment</dt><dd>{{document.paymentStatus}}</dd></template></dl></div>
<div class="overflow-x-auto"><table v-if="document.lines?.length" class="table table-sm"><thead><tr><th>Description</th><th class="text-end">Quantity</th><th class="text-end">Unit price</th><th class="text-end">Total</th></tr></thead><tbody><tr v-for="(line,index) in (document.lines || [])" :key="index"><td>{{line.description}}</td><td class="text-end tabular-nums">{{line.quantityUnits/1000000}} {{line.unit}}</td><td class="text-end whitespace-nowrap tabular-nums">{{formatMoney(line.unitPriceRial,currencyUnit)}}</td><td class="text-end whitespace-nowrap tabular-nums">{{formatMoney(line.lineTotalRial,currencyUnit)}}</td></tr></tbody></table>
<table v-if="document.statementLines?.length" class="table table-sm"><thead><tr><th>Date</th><th>Reference</th><th>Description</th><th class="text-end">Debit</th><th class="text-end">Credit</th><th class="text-end">Balance</th></tr></thead><tbody><tr v-for="(line,index) in (document.statementLines || [])" :key="index"><td>{{line.date ? formatDateTime(line.date) : '—'}}</td><td>{{line.reference}}</td><td>{{line.description}}</td><td class="text-end tabular-nums">{{formatMoney(line.debitRial,currencyUnit)}}</td><td class="text-end tabular-nums">{{formatMoney(line.creditRial,currencyUnit)}}</td><td class="text-end tabular-nums">{{formatMoney(line.balanceRial,currencyUnit)}}</td></tr></tbody></table>
<table v-if="document.allocations?.length" class="table table-sm"><thead><tr><th>Allocation</th><th>Reference</th><th class="text-end">Amount</th></tr></thead><tbody><tr v-for="(allocation,index) in (document.allocations || [])" :key="index"><td>{{allocation.targetType}}</td><td>{{allocation.reference}}</td><td class="text-end whitespace-nowrap tabular-nums">{{formatMoney(allocation.amountRial,currencyUnit)}}</td></tr></tbody></table></div>
<div class="flex justify-end gap-5 border-t border-base-300 pt-3"><span>Total</span><strong class="tabular-nums">{{formatMoney(document.totalRial || document.amountRial,currencyUnit)}}</strong></div>
<p class="text-xs leading-5">{{document.notes || document.shop.documentFooter}}</p>
</article>
</template>
<style scoped>
@media print {
 .print-document, .print-document table, .print-document th, .print-document td { background: white; color: black; }
 .print-document { padding: 0; font-size: 10pt; }
 .print-document th, .print-document td { border-bottom: 1px solid #ccc; }
 .print-document p { color: inherit; }
 .print-document tr { break-inside: avoid; }
}
</style>
