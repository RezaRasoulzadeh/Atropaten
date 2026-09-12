<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, ClipboardList, Layers3, Plus, Search } from 'lucide-vue-next'
import EmptyState from '../../components/ui/EmptyState.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SearchField from '../../components/ui/SearchField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import OrderDetailPanel from './OrderDetailPanel.vue'
import { ordersApi, type OrderRecord } from '../../api/orders'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'
import { useWorkspaceActions } from '../../composables/useWorkspaceActions'
import { confirmAction } from '../../ui/feedback'

type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate'
type OrderFilter = 'All' | 'Draft' | 'Confirmed' | 'Production' | 'Delivery' | 'Closed' | 'Cancelled' | 'Archived'

const props = defineProps<{
  orders: OrderRecord[]
  currencyUnit: CurrencyUnit
  loading?: boolean
}>()

const emit = defineEmits<{
  'edit-order': [id: string]
  'new-order': []
  'order-updated': [order: OrderRecord]
  'order-removed': [id: string]
  notify: [message: string]
}>()
const { busy, runAction } = useWorkspaceActions()

const query = ref('')
const status = ref<OrderFilter>('All')
const selectedOrderId = ref<string | null>(null)
const page = ref(1)
const pageSize = 10

const statusOptions: OrderFilter[] = ['All', 'Draft', 'Confirmed', 'Production', 'Delivery', 'Closed', 'Cancelled', 'Archived']

const orderItems = (order: OrderRecord) => (Array.isArray(order.items) ? order.items : [])

function orderStatus(order: OrderRecord): OrderFilter {
  if (order.archived) return 'Archived'
  if (order.commercialStatus === 'Cancelled') return 'Cancelled'
  if (order.commercialStatus === 'Closed') return 'Closed'
  if (order.fulfillmentStatus === 'Delivered') return 'Delivery'
  if (order.fulfillmentStatus === 'In Production' || order.fulfillmentStatus === 'Ready') return 'Production'
  return order.commercialStatus as OrderFilter
}

function statusCount(filter: OrderFilter) {
  if (filter === 'All') return props.orders.length
  return props.orders.filter((order) => orderStatus(order) === filter).length
}

const filteredOrders = computed(() => {
  const search = query.value.trim().toLowerCase()
  return props.orders.filter((order) => {
    const matchesSearch =
      !search ||
      [order.orderNumber, order.customerName, order.customerPhone, ...orderItems(order).map((item) => item.serviceName)].some((value) =>
        String(value ?? '').toLowerCase().includes(search),
      )

    return matchesSearch && (status.value === 'All' || orderStatus(order) === status.value)
  })
})

const visibleOrders = computed(() => filteredOrders.value)

const pageCount = computed(() => Math.max(1, Math.ceil(visibleOrders.value.length / pageSize)))
const pagedOrders = computed(() => visibleOrders.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const selectedOrder = computed(() => props.orders.find((order) => order.id === selectedOrderId.value) ?? null)
const pageNumbers = computed(() => Array.from({ length: pageCount.value }, (_, index) => index + 1))
const pageSummary = computed(() => visibleOrders.value.length
  ? 'Showing ' + ((page.value - 1) * pageSize + 1) + '–' + Math.min(page.value * pageSize, visibleOrders.value.length) + ' of ' + visibleOrders.value.length + ' orders'
  : '0 orders')

function tone(value: string): Tone {
  if (value === 'Confirmed' || value === 'Production') return 'blue'
  if (value === 'Closed' || value === 'Delivery' || value === 'Delivered' || value === 'Paid' || value === 'Ready') return 'green'
  if (value === 'Cancelled') return 'red'
  if (value === 'Partially Paid' || value === 'Urgent') return 'amber'
  return 'slate'
}

function itemSummary(order: OrderRecord) {
  return orderItems(order).map((item) => item.serviceName).join(' · ') || 'No configured items'
}

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}

function clearFilters() {
  query.value = ''
  status.value = 'All'
  page.value = 1
}

function goToPage(value: number) {
  page.value = Math.min(Math.max(value, 1), pageCount.value)
}

watch([query, status], () => { page.value = 1 })
watch(pageCount, (count) => { if (page.value > count) page.value = count })
watch(visibleOrders, (orders) => {
  if (!orders.some((order) => order.id === selectedOrderId.value)) selectedOrderId.value = orders[0]?.id ?? null
}, { immediate: true })

function selectOrder(id: string) {
  selectedOrderId.value = id
}

function editSelectedOrder() {
  if (selectedOrder.value) emit('edit-order', selectedOrder.value.id)
}

async function archiveSelectedOrder() {
  return runAction(async () => {
    const order = selectedOrder.value
    if (!order || order.archived) return
    const updated = await ordersApi.archive(order.id)
    emit('order-updated', updated)
    emit('notify', 'Order archived.')
  })
}

async function unarchiveSelectedOrder() {
  return runAction(async () => {
    const order = selectedOrder.value
    if (!order || !order.archived) return
    const updated = await ordersApi.unarchive(order.id)
    emit('order-updated', updated)
    emit('notify', 'Order unarchived.')
  })
}

async function removeSelectedOrder() {
  return runAction(async () => {
    const order = selectedOrder.value
    if (!order || !(await confirmAction({
      title: 'Delete order',
      message: 'Delete this order and its production jobs, draft invoices, and document links? Reserved and consumed stock will be returned, outsourcing expenses reversed, and posted invoices voided. Accounting history and files on disk are kept. Active payments must be reversed first.',
      confirmLabel: 'Delete order',
      danger: true,
    }))) return
    await ordersApi.remove(order.id)
    selectedOrderId.value = null
    emit('order-removed', order.id)
    emit('notify', 'Order deleted.')
  })
}
</script>

<template>
  <div class="flex h-full min-h-0 min-w-0 flex-col overflow-hidden" aria-label="Orders workspace">
    <WorkspaceStickyStack class="shrink-0" :flush="true">
      <WorkspaceHeader
        :show-breadcrumb="true"
        title="Orders"
        description="Track every order from draft through production, delivery, and close."
      >
        <SearchField v-model="query" class="w-full min-w-0 sm:w-64" placeholder="Search orders…" aria-label="Search orders" />
        <button class="btn btn-primary w-full gap-2 sm:w-auto" type="button" @click="emit('new-order')">
          <Plus :size="16" aria-hidden="true" />New order
        </button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <div class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(22rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.15fr)_minmax(24rem,0.85fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100" aria-label="Order register">
      <div class="shrink-0 border-b border-base-300 p-3 sm:p-4">
        <div class="flex min-w-0 flex-wrap items-center justify-between gap-3">
          <div class="flex min-w-0 flex-wrap items-center gap-2">
            <button
              v-for="filter in statusOptions"
              :key="filter"
              class="inline-flex h-9 items-center gap-2 rounded-box border px-3 text-sm transition-colors"
              :class="status === filter ? 'border-primary bg-primary/10 text-primary' : 'border-base-300 text-base-content/70 hover:border-primary/50 hover:text-base-content'"
              type="button"
              @click="status = filter"
            >
              <span class="size-2 rounded-full" :class="filter === 'Cancelled' ? 'bg-error' : filter === 'All' ? 'bg-primary' : filter === 'Closed' || filter === 'Delivery' ? 'bg-success' : filter === 'Production' || filter === 'Confirmed' ? 'bg-info' : 'bg-base-content/35'"></span>
              {{ filter }}
              <span class="rounded-full bg-base-200 px-1.5 py-0.5 text-xs tabular-nums">{{ statusCount(filter) }}</span>
            </button>
          </div>
          <button v-if="query || status !== 'All'" class="btn btn-ghost btn-sm" type="button" @click="clearFilters">Clear</button>
        </div>
      </div>

      <div class="order-register-table-head hidden grid-cols-[minmax(0,1.35fr)_minmax(10rem,1fr)_minmax(0,1.25fr)_8.5rem_7rem_1.25rem] gap-3 border-b border-base-300 px-4 py-3 text-xs font-medium text-base-content/55 md:grid">
        <span>Order</span><span>Customer</span><span>Items</span><span>Total</span><span>Status</span><span></span>
      </div>

      <LoadingState v-if="loading" label="Loading orders…" />
      <div v-else-if="pagedOrders.length" class="min-h-0 flex-1 overflow-y-auto divide-y divide-base-300">
        <button
          v-for="order in pagedOrders"
          :key="order.id"
          class="order-register-row group grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200/60 focus-visible:bg-base-200/60 focus-visible:outline focus-visible:outline-1 focus-visible:outline-primary md:grid-cols-[minmax(0,1.35fr)_minmax(10rem,1fr)_minmax(0,1.25fr)_8.5rem_7rem_1.25rem]"
          type="button"
          :class="selectedOrderId === order.id ? 'bg-primary/10' : ''"
          @click="selectOrder(order.id)"
        >
          <span class="order-register-name flex min-w-0 items-center gap-3">
            <span class="grid size-9 shrink-0 place-items-center rounded-box border border-base-300 bg-base-200 text-primary"><ClipboardList :size="18" aria-hidden="true" /></span>
            <span class="min-w-0">
              <strong class="block truncate text-sm">{{ order.orderNumber }}</strong>
              <span class="block truncate text-xs text-base-content/60">{{ order.customerName || 'Walk-in customer' }}</span>
            </span>
          </span>
          <span class="hidden min-w-0 truncate text-xs text-base-content/70 md:block">{{ order.customerPhone || 'No contact details' }}</span>
          <span class="hidden min-w-0 truncate text-xs text-base-content/70 md:block">
            {{ itemSummary(order) }} · {{ orderItems(order).length }} line item{{ orderItems(order).length === 1 ? '' : 's' }}
          </span>
          <span class="hidden text-sm font-semibold tabular-nums text-primary md:block">{{ money(order.totalRial) }}</span>
          <StatusBadge class="justify-self-end md:justify-self-start" :label="orderStatus(order)" :tone="tone(orderStatus(order))" />
          <ChevronRight :size="17" class="register-row-arrow justify-self-end text-base-content/45" aria-hidden="true" />
          <span class="col-span-3 flex flex-wrap gap-x-3 gap-y-1 text-xs text-base-content/55 md:hidden">
            <span>{{ itemSummary(order) }}</span>
            <span>{{ money(order.totalRial) }}</span>
            <span>{{ order.paymentStatus }} · {{ order.priority }}</span>
            <span>{{ formatDateTime(order.createdAt) }}</span>
          </span>
        </button>
      </div>
      <EmptyState
        v-else
        :title="props.orders.length ? 'No orders match this view' : 'No orders yet'"
        :description="props.orders.length ? 'Try another status or search filter.' : 'Create the first order to start a commercial workflow.'"
      >
        <template #icon><Search :size="22" aria-hidden="true" /></template>
        <template #action>
          <button v-if="props.orders.length" class="btn btn-primary btn-sm" type="button" @click="clearFilters">Clear filters</button>
          <button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="emit('new-order')"><Plus :size="15" aria-hidden="true" />Create order</button>
        </template>
      </EmptyState>

      <footer class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-4 py-3 text-xs text-base-content/60">
        <span>{{ pageSummary }}</span>
        <div class="flex items-center gap-1">
          <button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === 1" aria-label="Previous page" @click="goToPage(page - 1)"><ChevronLeft :size="15" aria-hidden="true" /></button>
          <button v-for="number in pageNumbers" :key="number" class="btn btn-xs min-w-8" :class="page === number ? 'btn-primary' : 'btn-ghost'" type="button" @click="goToPage(number)">{{ number }}</button>
          <button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === pageCount" aria-label="Next page" @click="goToPage(page + 1)"><ChevronRight :size="15" aria-hidden="true" /></button>
        </div>
      </footer>
      </section>

      <OrderDetailPanel v-if="selectedOrder" :order="selectedOrder" :currency-unit="props.currencyUnit" :busy="busy" @edit="editSelectedOrder" @archive="archiveSelectedOrder" @unarchive="unarchiveSelectedOrder" @remove="removeSelectedOrder" />
      <section v-else class="flex min-h-72 min-w-0 items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center">
        <EmptyState title="Select an order" description="Choose an order from the register to preview its customer, status, totals, and items.">
          <template #icon><Layers3 :size="22" aria-hidden="true" /></template>
        </EmptyState>
      </section>
    </div>
  </div>
</template>

<style scoped>
.order-register-table-head,
.order-register-row {
  grid-template-columns: minmax(0, 1.35fr) minmax(10rem, 1fr) minmax(0, 1.25fr) 8.5rem 7rem 1.25rem;
}

@media (max-width: 767px) {
  .order-register-row {
    grid-template-columns: minmax(0, 1fr) auto auto;
  }
}
</style>
