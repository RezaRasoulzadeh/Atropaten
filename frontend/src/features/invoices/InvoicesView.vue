<script setup lang="ts">
import { formatQuantityUnits } from '../../utils/quantity'
import LoadingState from '../../components/ui/LoadingState.vue'
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import InlineAlert from '../../components/ui/InlineAlert.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import MasterDetail from '../../components/layout/MasterDetail.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTableRow from '../../components/ui/DataTableRow.vue';
import DataTable from '../../components/ui/DataTable.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref } from 'vue';
import { FileText, Plus, RotateCcw, Trash2, X } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import { invoicesApi, type InvoiceRecord } from '../../api/invoices';
import type { OrderRecord } from '../../api/orders';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import { confirmAction, normalizeError } from '../../ui/feedback';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';

const props = defineProps<{ currencyUnit: CurrencyUnit; orders: OrderRecord[] }>();
const emit = defineEmits<{ notify: [string]; refreshOrders: [] }>();
const rows = ref<InvoiceRecord[]>([]);
const selected = ref<string | null>(null);
const query = ref('');
const status = ref('All');
const error = ref('');
const loading = ref(false);
const current = computed(() => rows.value.find((v) => v.id === selected.value) ?? null);
const filtered = computed(() =>
  rows.value.filter(
    (v) =>
      (status.value === 'All' || v.status === status.value) &&
      (!query.value.trim() ||
        [v.invoiceNumber, v.customerName, v.orderId]
          .join(' ')
          .toLowerCase()
          .includes(query.value.trim().toLowerCase())),
  ),
);
function tone(value: string) {
  return value === 'Paid' || value === 'Posted'
    ? 'green'
    : value === 'Partially Paid'
      ? 'blue'
      : value === 'Voided'
        ? 'slate'
        : 'amber';
}
async function load() {
  loading.value = true;
  try {
    rows.value = await invoicesApi.list();
    if (!selected.value && rows.value[0]) selected.value = rows.value[0].id;
  } catch (e) {
    error.value = normalizeError(e).message;
  } finally {
    loading.value = false;
  }
}
onMounted(load);
async function create(orderId: string) {
return runAction(async () => {
  try {
    const value = await invoicesApi.createFromOrder(orderId);
    rows.value = [value, ...rows.value];
    selected.value = value.id;
    emit('notify', 'Draft invoice created from the saved order snapshot.');
    emit('refreshOrders');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
async function post() {
return runAction(async () => {
  if (!current.value) return;
  try {
    const value = await invoicesApi.post(current.value.id);
    rows.value = rows.value.map((v) => (v.id === value.id ? value : v));
    emit('notify', 'Invoice posted with AR, revenue, and eligible actual COGS.');
    emit('refreshOrders');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
async function reverse() {
return runAction(async () => {
  if (
    !current.value ||
    !(await confirmAction({
      title: 'Void invoice',
      message: 'Void this invoice with reversing journal entries?',
      confirmLabel: 'Void invoice',
      danger: true,
    }))
  )
    return;
  try {
    const value = await invoicesApi.void(current.value.id);
    rows.value = rows.value.map((v) => (v.id === value.id ? value : v));
    emit('notify', 'Invoice voided with history preserved.');
    emit('refreshOrders');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
async function remove() {
return runAction(async () => {
  if (
    !current.value ||
    !(await confirmAction({
      title: 'Delete draft invoice',
      message: 'Delete this draft invoice permanently?',
      confirmLabel: 'Delete invoice',
      danger: true,
    }))
  )
    return;
  try {
    await invoicesApi.deleteDraft(current.value.id);
    rows.value = rows.value.filter((v) => v.id !== current.value!.id);
    selected.value = null;
    emit('notify', 'Draft invoice deleted.');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
</script>
<template>
  <div class="min-w-0 space-y-3">
    <WorkspaceStickyStack
      ><WorkspaceHeader
        title="Invoices"
        eyebrow="Finance / receivables"
        description="Invoice saved order snapshots once; derive receivables from allocations."
        ><span>{{ rows.length }} invoices</span></WorkspaceHeader
      >
      <SearchFilterBar><template #search><SearchField
          v-model="query"
          label="Search invoices"
          placeholder="Search invoice, customer, or order"
        /></template><template #filters><SelectField
          v-model="status"
          label="Status"
          :options="
            ['All', 'Draft', 'Posted', 'Partially Paid', 'Paid', 'Voided'].map((value) => ({
              label: value,
              value,
            }))
          "
        /></template><template #count><span class="self-end pb-2">{{ filtered.length }} shown</span></template></SearchFilterBar></WorkspaceStickyStack
    >
    <InlineAlert v-if="error" role="alert" class="flex flex-wrap items-center gap-2" tone="error"
      >{{ error }}
      <button class="btn btn-ghost" @click="error = ''" aria-label="Dismiss">
        <X :size="14" /></button
    ></InlineAlert>
    <MasterDetail
      ><AppPanel
        title="Invoice register"
        subtitle="Invoice status and amounts are derived in Go from posted allocations."
        ><LoadingState v-if="loading" label="Loading records…" />
        <div v-else-if="filtered.length">
          <DataTable
            ><thead>
              <tr>
                <th>Invoice</th>
                <th>Customer</th>
                <th>Order</th>
                <th>Date</th>
                <th class="text-end">Total</th>
                <th class="text-end">Remaining</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <DataTableRow
                v-for="value in filtered"
                :key="value.id"
                :class="{ 'bg-base-300': selected === value.id }"
                @activate="selected = value.id"
                interactive
                ><DataTableCell
                  ><span>{{ value.invoiceNumber }}</span></DataTableCell
                ><DataTableCell>{{ value.customerName || 'Walk-in customer' }}</DataTableCell
                ><DataTableCell>{{ value.orderId || '—' }}</DataTableCell
                ><DataTableCell>{{ formatDateTime(value.issueDate) }}</DataTableCell
                ><DataTableCell numeric>{{
                  formatMoney(value.totalRial, props.currencyUnit)
                }}</DataTableCell
                ><DataTableCell numeric>{{
                  formatMoney(value.remainingRial, props.currencyUnit)
                }}</DataTableCell
                ><DataTableCell
                  ><StatusBadge :label="value.status" :tone="tone(value.status)" /></DataTableCell
              ></DataTableRow></tbody
          ></DataTable>
        </div>
        <div v-else class="min-w-0 space-y-3">
          <FileText :size="22" /><strong>{{
            rows.length ? 'No invoices match this filter' : 'No invoices yet'
          }}</strong>
        </div></AppPanel
      >
      <InspectorShell
        v-if="current"
        title="Invoice inspector"
        subtitle="Exact immutable line and accounting snapshots."
        ><div class="min-w-0 space-y-3">
          <div class="min-w-0 space-y-3">
            <div><FileText :size="19" /></div>
            <div class="min-w-0 space-y-3">
              <h3 class="text-sm font-semibold">{{ current.invoiceNumber }}</h3>
              <p>
                {{ current.customerName || 'Walk-in customer' }} ·
                {{ current.orderId || 'No order link' }}
              </p>
            </div>
          </div>
          <StatusBadge :label="current.status" :tone="tone(current.status)" />
          <div>
            <div
              v-for="line in current.items"
              :key="line.id"
              class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
            >
              <span
                ><strong>{{ line.description }}</strong
                ><small class="block text-xs leading-5 text-base-content/60"
                  >{{ formatQuantityUnits(line.quantity) }} {{ line.quantityUnit }} · snapshot
                  {{ formatMoney(line.unitPriceRial, props.currencyUnit) }}</small
                ></span
              ><strong>{{ formatMoney(line.lineTotalRial, props.currencyUnit) }}</strong>
            </div>
          </div>
          <dl class="grid min-w-0 gap-2 text-sm">
            <div
              class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
            >
              <dt class="text-xs text-base-content/60">Subtotal</dt>
              <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
                {{ formatMoney(current.subtotalRial, props.currencyUnit) }}
              </dd>
            </div>
            <div
              class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
            >
              <dt class="text-xs text-base-content/60">Paid</dt>
              <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
                {{ formatMoney(current.paidRial, props.currencyUnit) }}
              </dd>
            </div>
            <div
              class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
            >
              <dt class="text-xs text-base-content/60">Remaining</dt>
              <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
                {{ formatMoney(current.remainingRial, props.currencyUnit) }}
              </dd>
            </div>
          </dl>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-primary" v-if="current.status === 'Draft'" @click="post" :disabled="busy">
              <Plus :size="15" /> Post invoice</button
            ><button class="btn btn-ghost" v-if="current.status === 'Draft'" @click="remove" :disabled="busy">
              <Trash2 :size="15" /> Delete draft</button
            ><button
              class="btn btn-ghost"
              v-if="['Posted', 'Partially Paid', 'Paid'].includes(current.status)"
              @click="reverse"
             :disabled="busy">
              <RotateCcw :size="15" /> Void / reverse
            </button>
          </div>
        </div></InspectorShell
      ><InspectorShell
        v-else
        title="Invoice inspector"
        subtitle="Select an invoice or create one from a saved order."
        ><div>No invoice selected.</div></InspectorShell
      ></MasterDetail
    >
    <AppPanel
      title="Orders ready to invoice"
      subtitle="Creating an invoice copies stored order pricing snapshots exactly."
      ><div class="min-w-0 space-y-3">
        <div
          v-for="order in props.orders.filter((value) => !value.invoiceId && value.totalRial > 0)"
          :key="order.id"
          class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
        >
          <span
            ><strong
              >{{ order.orderNumber }} · {{ order.customerName || 'Walk-in customer' }}</strong
            ><small class="block text-xs leading-5 text-base-content/60"
              >{{ formatMoney(order.totalRial, props.currencyUnit) }} ·
              {{ order.items.length }} items</small
            ></span
          ><button class="btn btn-primary" @click="create(order.id)" :disabled="busy">
            <Plus :size="14" /> Create invoice
          </button>
        </div>
        <div v-if="!props.orders.some((value) => !value.invoiceId && value.totalRial > 0)">
          All priced orders already have invoices.
        </div>
      </div></AppPanel
    >
  </div>
</template>
