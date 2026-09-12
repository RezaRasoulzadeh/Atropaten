<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import { nextTick, ref, watch } from 'vue';
import { FileText, Plus, Printer, RotateCcw } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import { invoicesApi, type InvoiceRecord } from '../../api/invoices';
import { ordersApi, type OrderRecord } from '../../api/orders';
import { reportsApi, type ShopSettingsRecord } from '../../api/reports';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney } from '../../utils/currency';
import { confirmAction } from '../../ui/feedback';
import EmptyState from '../../components/ui/EmptyState.vue';
import InvoicePrintDocument from '../invoices/InvoicePrintDocument.vue';
const props = defineProps<{ order: OrderRecord; currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [string]; saved: [order: OrderRecord] }>();
const invoice = ref<InvoiceRecord | null>(null);
const shopSettings = ref<ShopSettingsRecord | null>(null);
const printMode = ref<'pre' | 'final'>('final');
async function load() {
  try {
    const id = props.order.invoiceId;
    const loaded = id ? await invoicesApi.get(id) : null;
    if (id === props.order.invoiceId) invoice.value = loaded;
  } catch (e) {
    reportError(e);
  }
}
watch(() => [props.order.id, props.order.invoiceId, props.order.paidRial, props.order.remainingRial, props.order.invoiceStatus], load, { immediate: true });
async function create() {
return runAction(async () => {
  try {
    invoice.value = await invoicesApi.createFromOrder(props.order.id);
    emit('notify', 'Draft invoice created from this order.');
    emit('saved', await ordersApi.get(props.order.id));
  } catch (e) {
    reportError(e);
  }

});
}
async function post() {
return runAction(async () => {
  if (!invoice.value) return;
  try {
    invoice.value = await invoicesApi.post(invoice.value.id);
    emit('notify', 'Invoice posted.');
    emit('saved', await ordersApi.get(props.order.id));
  } catch (e) {
    reportError(e);
  }

});
}
async function printInvoice(mode: 'pre' | 'final') {
return runAction(async () => {
  if (!invoice.value) return;
  try {
    if (!shopSettings.value) shopSettings.value = await reportsApi.settings();
    printMode.value = mode;
    await nextTick();
    window.print();
  } catch (e) {
    reportError(e);
  }
});
}
async function reverse() {
return runAction(async () => {
  if (
    !invoice.value ||
    !(await confirmAction({
      title: 'Void invoice',
      message: 'Void this invoice with a reversal?',
      confirmLabel: 'Void invoice',
      danger: true,
    }))
  )
    return;
  try {
    invoice.value = await invoicesApi.void(invoice.value.id);
    emit('notify', 'Invoice voided with history preserved.');
    emit('saved', await ordersApi.get(props.order.id));
  } catch (e) {
    reportError(e);
  }

});
}
</script>
<template>
  <section class="invoice-step min-w-0">
    <header class="flex min-w-0 flex-wrap items-start justify-between gap-3 border-b border-base-300 pb-3">
      <div class="flex min-w-0 items-center gap-3">
        <span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary">
          <FileText :size="18" aria-hidden="true" />
        </span>
        <div class="min-w-0">
          <h2 class="text-sm font-semibold leading-5">Invoice</h2>
          <p class="mt-0.5 text-xs leading-4 text-base-content/60">Create, print, or post the commercial document.</p>
        </div>
      </div>
      <StatusBadge
        v-if="invoice"
        :label="invoice.status"
        :tone="invoice.status === 'Paid' || invoice.status === 'Posted' ? 'green' : invoice.status === 'Voided' ? 'slate' : 'amber'"
      />
    </header>

    <div v-if="invoice" class="min-w-0 pt-4">
      <div class="grid min-w-0 gap-4 border-b border-base-300 pb-4 lg:grid-cols-[minmax(0,1fr)_minmax(20rem,0.9fr)] lg:items-center">
        <div class="flex min-w-0 items-center gap-3">
          <span class="grid size-11 shrink-0 place-items-center rounded-box border border-primary/25 bg-primary/10 text-primary">
            <FileText :size="22" aria-hidden="true" />
          </span>
          <div class="min-w-0">
            <p class="text-[11px] font-medium uppercase tracking-wide text-primary/75">{{ invoice.status === 'Draft' ? 'Pre-invoice' : 'Final invoice' }}</p>
            <h3 class="mt-0.5 truncate text-base font-semibold">{{ invoice.invoiceNumber }}</h3>
            <p class="mt-1 text-xs text-base-content/60">{{ invoice.items.length }} line{{ invoice.items.length === 1 ? '' : 's' }} · {{ invoice.customerName || 'Walk-in customer' }}</p>
          </div>
        </div>
        <dl class="grid min-w-0 grid-cols-3 gap-3 lg:border-s lg:ps-4">
          <div class="min-w-0">
            <dt class="text-[11px] text-base-content/55">Total</dt>
            <dd class="mt-1 truncate text-sm font-semibold tabular-nums">{{ formatMoney(invoice.totalRial, props.currencyUnit) }}</dd>
          </div>
          <div class="min-w-0">
            <dt class="text-[11px] text-base-content/55">Paid</dt>
            <dd class="mt-1 truncate text-sm font-semibold tabular-nums text-success">{{ formatMoney(invoice.paidRial, props.currencyUnit) }}</dd>
          </div>
          <div class="min-w-0">
            <dt class="text-[11px] text-base-content/55">Remaining</dt>
            <dd class="mt-1 truncate text-sm font-semibold tabular-nums text-warning">{{ formatMoney(invoice.remainingRial, props.currencyUnit) }}</dd>
          </div>
        </dl>
      </div>

      <div class="flex min-w-0 flex-wrap items-center justify-between gap-3 pt-4">
        <p class="text-xs leading-4 text-base-content/60">
          {{ invoice.status === 'Draft' ? 'This document is not posted to accounting.' : 'This document is posted and ready as the final invoice.' }}
        </p>
        <div class="flex flex-wrap items-center gap-2">
          <button v-if="invoice.status === 'Draft'" class="btn btn-ghost btn-sm gap-1.5" type="button" @click="printInvoice('pre')" :disabled="busy">
            <Printer :size="14" aria-hidden="true" /> Print pre-invoice
          </button>
          <button v-if="invoice.status === 'Draft'" class="btn btn-primary btn-sm gap-1.5" type="button" @click="post" :disabled="busy">
            <Plus :size="14" aria-hidden="true" /> Post invoice
          </button>
          <button v-if="invoice.status === 'Posted' || invoice.status === 'Partially Paid' || invoice.status === 'Paid'" class="btn btn-primary btn-sm gap-1.5" type="button" @click="printInvoice('final')" :disabled="busy">
            <Printer :size="14" aria-hidden="true" /> Print final invoice
          </button>
          <button v-if="invoice.status === 'Posted' || invoice.status === 'Partially Paid' || invoice.status === 'Paid'" class="btn btn-ghost btn-sm gap-1.5" type="button" @click="reverse" :disabled="busy">
            <RotateCcw :size="14" aria-hidden="true" /> Void
          </button>
        </div>
      </div>
    </div>
    <EmptyState v-else compact title="No invoice linked" description="Create an invoice when this order is ready to bill.">
      <template #icon><FileText :size="22" aria-hidden="true" /></template>
      <template #action><button class="btn btn-primary btn-sm" type="button" @click="create" :disabled="busy"><Plus :size="15" aria-hidden="true" /> Create invoice</button></template>
    </EmptyState>
    <Teleport to="body">
      <div v-if="invoice" class="print-output">
        <InvoicePrintDocument :invoice="invoice" :shop="shopSettings" :currency-unit="props.currencyUnit" :document-type="printMode" />
      </div>
    </Teleport>
  </section>
</template>
