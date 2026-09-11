<script setup lang="ts">
import { ref, watch } from 'vue'
import { Archive, Check, Edit3, FileText, Receipt, RotateCcw, ShoppingCart, Trash2 } from 'lucide-vue-next'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'
import type { usePurchasesWorkspace } from './usePurchasesWorkspace'

type PurchaseTab = 'overview' | 'items' | 'details'

const props = defineProps<{
  workspace: ReturnType<typeof usePurchasesWorkspace>
  currencyUnit: CurrencyUnit
}>()

const { busy, current, financial, startEdit, post, archivePurchase, unarchivePurchase, removePurchase } = props.workspace
const activeTab = ref<PurchaseTab>('overview')
watch(() => current.value?.id, () => { activeTab.value = 'overview' })

function statusTone(status: string) {
  return status === 'Posted' ? 'green' : status === 'Archived' || status === 'Cancelled' ? 'slate' : 'amber'
}

function financialAccountLabel(id: string) {
  const account = financial.value.find((value) => value.id === id)
  if (!account) return 'No account selected'
  const identity = account.type === 'bank'
    ? [account.bankName, account.accountNumber && `Account ${account.accountNumber}`, account.cardNumber && `Card ${account.cardNumber}`].filter(Boolean).join(' · ')
    : 'Cash account'
  return identity ? `${account.name} · ${identity}` : account.name
}
</script>

<template>
  <section v-if="current" class="purchase-detail-panel h-auto min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:h-full xl:overflow-y-auto" aria-label="Purchase details">
    <header class="border-b border-base-300 p-3 sm:p-5">
      <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300 sm:min-h-60">
        <div class="absolute inset-0 bg-gradient-to-t from-black/95 via-black/65 to-black/10" aria-hidden="true"></div>
        <div class="absolute inset-0 grid place-items-center text-white/10"><ShoppingCart :size="92" :stroke-width="1" aria-hidden="true" /></div>
        <div class="relative z-10 flex min-h-52 items-end justify-start p-4 text-white sm:min-h-60 sm:p-5"><div class="w-full min-w-0 text-start"><div class="flex min-w-0 items-center justify-start gap-3"><h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ current.purchaseNumber }}</h2><StatusBadge class="shrink-0" :label="current.status" :tone="statusTone(current.status)" /></div><div class="mt-2 flex min-w-0 flex-wrap items-center justify-start gap-x-3 gap-y-1 text-sm text-white/75"><span>{{ current.supplierName || 'No supplier' }}</span><span class="size-1 rounded-full bg-white/50" aria-hidden="true"></span><span>{{ current.items.length }} item{{ current.items.length === 1 ? '' : 's' }}</span></div></div></div>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
        <button class="btn btn-outline btn-sm w-full gap-2" type="button" :disabled="busy" @click="startEdit"><Edit3 :size="14" aria-hidden="true" />Edit</button>
        <button v-if="current.status === 'Draft'" class="btn btn-primary btn-sm w-full gap-2" type="button" :disabled="busy" @click="post"><Check :size="14" aria-hidden="true" />Post</button>
        <button v-if="current.status !== 'Archived'" class="btn btn-outline btn-warning btn-sm w-full gap-2" type="button" :disabled="busy" @click="archivePurchase"><Archive :size="14" aria-hidden="true" />Archive</button>
        <button v-else class="btn btn-outline btn-success btn-sm w-full gap-2" type="button" :disabled="busy" @click="unarchivePurchase"><RotateCcw :size="14" aria-hidden="true" />Unarchive</button>
        <button class="btn btn-outline btn-error btn-sm w-full gap-2" type="button" :disabled="busy" @click="removePurchase"><Trash2 :size="14" aria-hidden="true" />Delete</button>
      </div>

      <div class="mt-4 grid min-w-0 divide-y divide-base-300 border-y border-base-300 sm:mt-5 sm:grid-cols-3 sm:divide-x sm:divide-y-0"><div class="flex items-center gap-3 py-3 sm:px-3 sm:first:pl-0"><Receipt :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Total</span><strong class="block text-sm tabular-nums">{{ formatMoney(current.totalRial, currencyUnit) }}</strong></div></div><div class="flex items-center gap-3 py-3 sm:px-3"><ShoppingCart :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Items</span><strong class="block text-sm">{{ current.items.length }}</strong></div></div><div class="flex items-center gap-3 py-3 sm:px-3 sm:pr-0"><RotateCcw :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Remaining</span><strong class="block text-sm tabular-nums">{{ formatMoney(current.remainingRial ?? current.totalRial, currencyUnit) }}</strong></div></div></div>
    </header>

    <nav class="flex min-w-0 overflow-x-auto border-b border-base-300 px-2" aria-label="Purchase details tabs"><button v-for="tab in [{ id: 'overview', label: 'Overview' }, { id: 'items', label: 'Items' }, { id: 'details', label: 'Additional info' }]" :key="tab.id" class="shrink-0 border-b-2 px-3 py-3 text-sm transition-colors" :class="activeTab === tab.id ? 'border-primary text-primary' : 'border-transparent text-base-content/65 hover:border-base-content/30 hover:text-base-content'" type="button" @click="activeTab = tab.id as PurchaseTab">{{ tab.label }}</button></nav>

    <div class="min-w-0 p-3 sm:p-4">
      <div v-if="activeTab === 'overview'" class="space-y-3">
        <div class="rounded-box border border-base-300 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Receipt :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Purchase summary</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Financial totals and posting state for this supplier purchase.</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Subtotal</dt><dd class="font-medium tabular-nums">{{ formatMoney(current.subtotalRial, currencyUnit) }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Discount</dt><dd class="font-medium tabular-nums">{{ formatMoney(current.discountRial, currencyUnit) }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Additional costs</dt><dd class="font-medium tabular-nums">{{ formatMoney(current.shippingRial + current.taxRial + current.additionalCostsRial, currencyUnit) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Total</dt><dd class="font-semibold tabular-nums text-primary">{{ formatMoney(current.totalRial, currencyUnit) }}</dd></div></dl></div>
        <div class="rounded-box border border-base-300 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><ShoppingCart :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Payment status</h3><p class="mt-1 text-xs leading-5 text-base-content/60">The account and amount associated with this purchase.</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Paid</dt><dd class="font-medium tabular-nums">{{ formatMoney(current.paidRial || 0, currencyUnit) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Remaining</dt><dd class="font-medium tabular-nums">{{ formatMoney(current.remainingRial ?? current.totalRial, currencyUnit) }}</dd></div></dl></div>
      </div>

      <div v-else-if="activeTab === 'items'" class="space-y-2"><div v-for="line in [...current.items].sort((a, b) => a.position - b.position)" :key="line.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><ShoppingCart :size="18" aria-hidden="true" /></span><div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ line.materialName }}</strong><span class="block truncate text-xs text-base-content/55">{{ line.purchaseQuantity }} {{ line.purchaseUnit }} → {{ line.consumptionQuantity }} {{ line.consumptionUnit }}</span></div><span class="shrink-0 text-sm tabular-nums">{{ formatMoney(line.lineTotalRial, currencyUnit) }}</span></div><EmptyState v-if="!current.items.length" compact title="No items added" description="This purchase does not contain any material lines."><template #icon><ShoppingCart :size="21" aria-hidden="true" /></template></EmptyState></div>

      <div v-else class="space-y-3"><div class="rounded-box border border-base-300 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileText :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Purchase information</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Supplier, date, account, and invoice details.</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Supplier</dt><dd class="break-words text-end">{{ current.supplierName || 'No supplier' }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Purchase date</dt><dd class="text-end">{{ formatDateTime(current.purchaseDate) }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Invoice</dt><dd class="break-words text-end">{{ current.supplierInvoiceNumber || 'No invoice number' }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Paid from</dt><dd class="break-words text-end">{{ financialAccountLabel(current.financialAccountId) }}</dd></div></dl></div><div v-if="current.notes" class="rounded-box border border-base-300 p-4"><h3 class="text-sm font-semibold">Notes</h3><p class="mt-2 whitespace-pre-wrap break-words text-sm leading-6 text-base-content/70">{{ current.notes }}</p></div></div>
    </div>
  </section>
</template>
