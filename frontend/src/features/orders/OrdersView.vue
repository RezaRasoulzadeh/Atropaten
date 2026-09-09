<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, ref } from 'vue';
import { ClipboardList, Plus } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import SearchField from '../../components/ui/SearchField.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import SelectField from '../../components/ui/SelectField.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import type { OrderRecord } from '../../api/orders';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate';
const props = defineProps<{
  orders: OrderRecord[];
  currencyUnit: CurrencyUnit;
  loading?: boolean;
}>();
const emit = defineEmits<{ 'open-order': [id: string]; 'new-order': [] }>();
const query = ref('');
const status = ref('All');
const payment = ref('All');
const priority = ref('All');
const statusOptions = [
  { label: 'All', value: 'All' },
  { label: 'Draft', value: 'Draft' },
  { label: 'Confirmed', value: 'Confirmed' },
  { label: 'Production', value: 'Production' },
  { label: 'Delivery', value: 'Delivery' },
  { label: 'Closed', value: 'Closed' },
  { label: 'Cancelled', value: 'Cancelled' },
];
const paymentOptions = [
  { label: 'All', value: 'All' },
  { label: 'Unpaid', value: 'Unpaid' },
  { label: 'Partially Paid', value: 'Partially Paid' },
  { label: 'Paid', value: 'Paid' },
];
const priorityOptions = [
  { label: 'All', value: 'All' },
  { label: 'Urgent', value: 'Urgent' },
  { label: 'High', value: 'High' },
  { label: 'Normal', value: 'Normal' },
  { label: 'Low', value: 'Low' },
];
const orderItems = (order: OrderRecord) => (Array.isArray(order.items) ? order.items : []);
const filtered = computed(() =>
  props.orders.filter((o) => {
    const q = query.value.trim().toLowerCase();
    return (
      (!q ||
        [o.orderNumber, o.customerName, ...orderItems(o).map((i) => i.serviceName)].some((v) =>
          String(v ?? '')
            .toLowerCase()
            .includes(q),
        )) &&
      (status.value === 'All' || orderStatus(o) === status.value) &&
      (payment.value === 'All' || o.paymentStatus === payment.value) &&
      (priority.value === 'All' || o.priority === priority.value)
    );
  }),
);
function orderStatus(order: OrderRecord) {
  if (order.commercialStatus === 'Cancelled') return 'Cancelled';
  if (order.commercialStatus === 'Closed') return 'Closed';
  if (order.fulfillmentStatus === 'Delivered') return 'Delivery';
  if (order.fulfillmentStatus === 'In Production' || order.fulfillmentStatus === 'Ready') return 'Production';
  return order.commercialStatus;
}
function money(v: number) {
  return formatMoney(v, props.currencyUnit);
}
function tone(v: string): Tone {
  return v === 'Confirmed' || v === 'Production'
    ? 'blue'
    : v === 'Closed' || v === 'Delivery' || v === 'Delivered' || v === 'Paid' || v === 'Ready'
      ? 'green'
      : v === 'Cancelled'
        ? 'red'
        : v === 'Partially Paid'
          ? 'amber'
          : 'slate';
}
function itemSummary(o: OrderRecord) {
  return (
    orderItems(o)
      .map((i) => i.serviceName)
      .join(' · ') || 'No configured items'
  );
}
function clear() {
  query.value = '';
  status.value = payment.value = priority.value = 'All';
}
</script>
<template>
  <div class="space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        :show-breadcrumb="true"
        eyebrow="Sales / operational queue"
        title="Orders"
        description="Track every order from draft through production, delivery, and close."
      >
        <button class="btn btn-primary gap-2" type="button" @click="emit('new-order')">
          <Plus :size="16" :stroke-width="1.8" aria-hidden="true" />
          <span>New order</span>
        </button>
      </WorkspaceHeader>

      <SearchFilterBar>
        <template #search>
          <SearchField
            v-model="query"
            label="Search orders"
            placeholder="Order, customer, or service"
          />
        </template>
        <template #filters>
          <SelectField
            class="w-36"
            v-model="status"
            label="Order status"
            :options="statusOptions"
          />
          <SelectField class="w-32" v-model="payment" label="Payment" :options="paymentOptions" />
          <SelectField
            class="w-32"
            v-model="priority"
            label="Priority"
            :options="priorityOptions"
          />
        </template>
        <template #count
          ><span>{{ filtered.length }} of {{ props.orders.length }} orders</span></template
        >
        <template #actions>
          <button
            v-if="
              query ||
              status !== 'All' ||
              payment !== 'All' ||
              priority !== 'All'
            "
            class="btn btn-ghost btn-sm"
            type="button"
            @click="clear"
          >
            Clear
          </button>
        </template>
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <AppPanel
      title="All orders"
      subtitle="Open an order to inspect its accepted pricing snapshots."
      :flush="true"
    >
      <template #action>
        <span class="text-xs text-base-content/60">Draft · Confirmed · Production · Delivery</span>
      </template>

      <LoadingState v-if="loading" label="Loading records…" />
      <EmptyState
        v-else-if="!filtered.length"
        :title="props.orders.length ? 'No orders match these filters' : 'No persisted orders yet'"
        :description="props.orders.length ? 'Adjust the search or status filters to find an order.' : 'Create an order to start a commercial workflow.'"
      >
        <template #icon><ClipboardList :size="22" aria-hidden="true" /></template>
        <template #action>
          <button v-if="props.orders.length" class="btn btn-primary btn-sm" type="button" @click="clear">Clear filters</button>
          <button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="emit('new-order')"><Plus :size="15" aria-hidden="true" /> Create order</button>
        </template>
      </EmptyState>
      <div v-else>
        <div class="divide-y divide-base-300">
          <article
            v-for="order in filtered"
            :key="order.id"
            class="min-w-0 cursor-pointer p-3 transition-colors hover:bg-base-200"
            tabindex="0"
            @click="emit('open-order', order.id)"
            @keydown.enter="emit('open-order', order.id)"
          >
            <div class="flex min-w-0 items-center justify-between gap-3">
              <div class="flex min-w-0 items-center gap-2">
                <button
                  class="btn btn-ghost btn-sm -ms-2 h-8 min-h-8 shrink-0 px-2 font-semibold"
                  type="button"
                  @click.stop="emit('open-order', order.id)"
                >
                  {{ order.orderNumber }}
                </button>
                <p class="truncate text-sm font-semibold">
                  {{ order.customerName || 'Walk-in customer' }}
                </p>
              </div>
              <strong class="shrink-0 text-sm text-primary">{{ money(order.totalRial) }}</strong>
            </div>

            <div
              class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2 lg:grid-cols-4"
            >
              <div class="min-w-0">
                <span class="block text-base-content/50">Items</span>
                <span class="block break-words text-base-content/80">{{ itemSummary(order) }}</span>
                <span class="block text-base-content/50"
                  >{{ orderItems(order).length }} line items</span
                >
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Contact</span>
                <span class="block truncate text-base-content/80">{{
                  order.customerPhone || 'No contact details'
                }}</span>
              </div>
              <div>
                <span class="block text-base-content/50">Created</span>
                <span class="block text-base-content/80">{{
                  formatDateTime(order.createdAt)
                }}</span>
              </div>
              <div>
                <span class="block text-base-content/50">Promised</span>
                <span class="block text-base-content/80">{{
                  order.promisedAt ? formatDateTime(order.promisedAt) : '—'
                }}</span>
              </div>
            </div>

            <div class="mt-2 flex flex-wrap gap-1.5">
              <StatusBadge :label="orderStatus(order)" :tone="tone(orderStatus(order))" />
              <StatusBadge :label="order.paymentStatus" :tone="tone(order.paymentStatus)" />
              <StatusBadge :label="order.priority" :tone="tone(order.priority)" />
            </div>
          </article>
        </div>
      </div>
    </AppPanel>
  </div>
</template>
