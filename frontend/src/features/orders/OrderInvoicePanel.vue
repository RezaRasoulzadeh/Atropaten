<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import { onMounted, ref } from 'vue';
import { FileText, Plus, RotateCcw } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import { invoicesApi, type InvoiceRecord } from '../../api/invoices';
import { ordersApi, type OrderRecord } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney } from '../../utils/currency';
import { confirmAction } from '../../ui/feedback';
import EmptyState from '../../components/ui/EmptyState.vue';
const props = defineProps<{ order: OrderRecord; currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [string]; saved: [order: OrderRecord] }>();
const invoice = ref<InvoiceRecord | null>(null);
async function load() {
  try {
    if (props.order.invoiceId) invoice.value = await invoicesApi.get(props.order.invoiceId);
  } catch (e) {
    reportError(e);
  }
}
onMounted(load);
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
  <section class="min-w-0 space-y-4">
    <header class="border-b border-base-300 pb-4">
      <h2 class="text-sm font-semibold">Invoice</h2>
      <p class="mt-1 text-xs leading-4 text-base-content/60">Commercial snapshot and receivable status</p>
    </header>
    <div v-if="invoice" class="min-w-0 space-y-3">
      <div class="min-w-0 space-y-3">
        <div><FileText :size="19" /></div>
        <div class="min-w-0 space-y-3">
          <h3 class="text-sm font-semibold">{{ invoice.invoiceNumber }}</h3>
          <p>
            {{ invoice.items.length }} lines ·
            {{ formatMoney(invoice.totalRial, props.currencyUnit) }}
          </p>
        </div>
        <StatusBadge
          :label="invoice.status"
          :tone="
            invoice.status === 'Paid' || invoice.status === 'Posted'
              ? 'green'
              : invoice.status === 'Voided'
                ? 'slate'
                : 'amber'
          "
        />
      </div>
      <div class="min-w-0 space-y-3">
        <span>Paid</span><strong>{{ formatMoney(invoice.paidRial, props.currencyUnit) }}</strong>
      </div>
      <div class="min-w-0 space-y-3">
        <span>Remaining</span
        ><strong>{{ formatMoney(invoice.remainingRial, props.currencyUnit) }}</strong>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <button class="btn btn-primary" v-if="invoice.status === 'Draft'" @click="post" :disabled="busy">
          <Plus :size="15" /> Post invoice</button
        ><button
          class="btn btn-ghost"
          v-if="
            invoice.status === 'Posted' ||
            invoice.status === 'Partially Paid' ||
            invoice.status === 'Paid'
          "
          @click="reverse"
         :disabled="busy">
          <RotateCcw :size="15" /> Void
        </button>
      </div>
    </div>
    <EmptyState v-else compact title="No invoice linked" description="Create an invoice when this order is ready to bill.">
      <template #icon><FileText :size="22" aria-hidden="true" /></template>
      <template #action><button class="btn btn-primary btn-sm" type="button" @click="create" :disabled="busy"><Plus :size="15" /> Create invoice</button></template>
    </EmptyState>
  </section>
</template>
