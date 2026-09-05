<script setup lang="ts">
import { computed, ref } from 'vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import type { CommercialStatus, FulfillmentStatus, Order, PaymentStatus, Priority } from '../data/orders'
import { formatMoney, type CurrencyUnit } from '../utils/currency'
import { formatDate, formatDateTime } from '../utils/date'
import { ArrowRight, ArrowUpDown, ChevronDown, Plus, Search, SearchX } from 'lucide-vue-next'

type SortKey = 'id' | 'customer' | 'created' | 'promised' | 'total' | 'balance' | 'priority'
type BadgeTone = 'blue' | 'green' | 'amber' | 'red' | 'slate'

const props = defineProps<{ orders: Order[]; currencyUnit: CurrencyUnit }>()

const emit = defineEmits<{
  'open-order': [orderId: string]
  'new-order': []
}>()

const searchQuery = ref('')
const commercialFilter = ref<'All' | CommercialStatus>('All')
const fulfillmentFilter = ref<'All' | FulfillmentStatus>('All')
const paymentFilter = ref<'All' | PaymentStatus>('All')
const priorityFilter = ref<'All' | Priority>('All')
const selectedOrderId = ref<string | null>(null)
const sortKey = ref<SortKey>('promised')
const sortAscending = ref(true)

const filteredOrders = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const filtered = props.orders.filter((order) => {
    const matchesSearch = !query || [order.id, order.customer, order.itemSummary].some((value) => value.toLowerCase().includes(query))
    const matchesCommercial = commercialFilter.value === 'All' || order.commercialStatus === commercialFilter.value
    const matchesFulfillment = fulfillmentFilter.value === 'All' || order.fulfillmentStatus === fulfillmentFilter.value
    const matchesPayment = paymentFilter.value === 'All' || order.paymentStatus === paymentFilter.value
    const matchesPriority = priorityFilter.value === 'All' || order.priority === priorityFilter.value
    return matchesSearch && matchesCommercial && matchesFulfillment && matchesPayment && matchesPriority
  })

  return [...filtered].sort((left, right) => {
    const leftValue = sortValue(left, sortKey.value)
    const rightValue = sortValue(right, sortKey.value)
    const comparison = leftValue < rightValue ? -1 : leftValue > rightValue ? 1 : 0
    return sortAscending.value ? comparison : -comparison
  })
})

const hasActiveFilters = computed(() => Boolean(searchQuery.value || commercialFilter.value !== 'All' || fulfillmentFilter.value !== 'All' || paymentFilter.value !== 'All' || priorityFilter.value !== 'All'))

function sortValue(order: Order, key: SortKey): string | number {
  if (key === 'balance') return order.totalRial - order.paidRial
  if (key === 'total') return order.totalRial
  if (key === 'priority') return ['Urgent', 'High', 'Normal', 'Low'].indexOf(order.priority)
  if (key === 'id') return order.id
  if (key === 'customer') return order.customer
  if (key === 'created') return order.createdAt
  return order.promisedAt ?? ''
}

function sortOrders(key: SortKey) {
  if (sortKey.value === key) {
    sortAscending.value = !sortAscending.value
    return
  }
  sortKey.value = key
  sortAscending.value = true
}

function clearFilters() {
  searchQuery.value = ''
  commercialFilter.value = 'All'
  fulfillmentFilter.value = 'All'
  paymentFilter.value = 'All'
  priorityFilter.value = 'All'
}

function selectOrder(orderId: string) {
  selectedOrderId.value = orderId
}

function openOrder(orderId: string) {
  selectedOrderId.value = orderId
  emit('open-order', orderId)
}

function currency(amountRial: number) {
  return formatMoney(amountRial, props.currencyUnit)
}

function commercialTone(status: CommercialStatus): BadgeTone {
  return status === 'Confirmed' ? 'blue' : status === 'Closed' ? 'green' : status === 'Quoted' ? 'amber' : 'slate'
}

function fulfillmentTone(status: FulfillmentStatus): BadgeTone {
  return status === 'In production' ? 'blue' : status === 'Ready' || status === 'Delivered' ? 'green' : 'slate'
}

function paymentTone(status: PaymentStatus): BadgeTone {
  return status === 'Paid' ? 'green' : status === 'Partially paid' ? 'amber' : 'red'
}

function priorityTone(priority: Priority): BadgeTone {
  return priority === 'Urgent' ? 'red' : priority === 'High' ? 'amber' : priority === 'Normal' ? 'blue' : 'slate'
}
</script>

<template>
  <div class="orders-view">
    <WorkspaceStickyStack>
      <header class="workspace-heading orders-heading">
      <div>
        <p class="eyebrow">Sales / operational queue</p>
        <h1>Orders</h1>
        <p class="heading-description">Track every customer order from confirmation through delivery.</p>
      </div>
      <button class="button button-primary" type="button" @click="$emit('new-order')">
        <Plus class="button-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />
        New order
      </button>
      </header>

      <section class="orders-filter-bar panel" aria-label="Order filters">
      <label class="orders-search">
        <span class="sr-only">Search orders</span>
        <Search :size="16" :stroke-width="1.8" aria-hidden="true" />
        <input v-model="searchQuery" type="search" placeholder="Search order, customer, or item" autocomplete="off" />
      </label>
      <label class="filter-control">
        <span>Commercial</span>
        <span class="select-control"><select v-model="commercialFilter" aria-label="Filter by commercial status"><option>All</option><option>Draft</option><option>Quoted</option><option>Confirmed</option><option>Closed</option></select><ChevronDown class="select-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" /></span>
      </label>
      <label class="filter-control">
        <span>Fulfillment</span>
        <span class="select-control"><select v-model="fulfillmentFilter" aria-label="Filter by fulfillment status"><option>All</option><option>Pending</option><option>In production</option><option>Ready</option><option>Delivered</option></select><ChevronDown class="select-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" /></span>
      </label>
      <label class="filter-control">
        <span>Payment</span>
        <span class="select-control"><select v-model="paymentFilter" aria-label="Filter by payment status"><option>All</option><option>Unpaid</option><option>Partially paid</option><option>Paid</option></select><ChevronDown class="select-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" /></span>
      </label>
      <label class="filter-control">
        <span>Priority</span>
        <span class="select-control"><select v-model="priorityFilter" aria-label="Filter by priority"><option>All</option><option>Urgent</option><option>High</option><option>Normal</option><option>Low</option></select><ChevronDown class="select-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" /></span>
      </label>
      <button v-if="hasActiveFilters" class="filter-clear" type="button" @click="clearFilters">Clear</button>
      <span class="filter-result">{{ filteredOrders.length }} of {{ orders.length }} orders</span>
      </section>
    </WorkspaceStickyStack>

    <section class="orders-list-panel panel" aria-labelledby="orders-list-title">
      <header class="orders-list-heading">
        <div>
          <h2 id="orders-list-title" class="panel-title">All orders</h2>
          <p class="panel-subtitle">Select a row to inspect it, or open the order workspace.</p>
        </div>
        <span class="orders-state-key"><span class="state-key-dot"></span>State axes are independent</span>
      </header>

      <div v-if="filteredOrders.length" class="table-wrap orders-table-wrap">
        <table class="data-table orders-table">
          <thead>
            <tr>
              <th scope="col" :aria-sort="sortKey === 'id' ? (sortAscending ? 'ascending' : 'descending') : 'none'"><button type="button" @click="sortOrders('id')">Order <ArrowUpDown class="sort-icon" :size="13" :stroke-width="1.8" aria-hidden="true" /></button></th>
              <th scope="col" :aria-sort="sortKey === 'customer' ? (sortAscending ? 'ascending' : 'descending') : 'none'"><button type="button" @click="sortOrders('customer')">Customer <ArrowUpDown class="sort-icon" :size="13" :stroke-width="1.8" aria-hidden="true" /></button></th>
              <th scope="col">Items</th>
              <th scope="col" :aria-sort="sortKey === 'created' ? (sortAscending ? 'ascending' : 'descending') : 'none'"><button type="button" @click="sortOrders('created')">Created <ArrowUpDown class="sort-icon" :size="13" :stroke-width="1.8" aria-hidden="true" /></button></th>
              <th scope="col" :aria-sort="sortKey === 'promised' ? (sortAscending ? 'ascending' : 'descending') : 'none'"><button type="button" @click="sortOrders('promised')">Promised <ArrowUpDown class="sort-icon" :size="13" :stroke-width="1.8" aria-hidden="true" /></button></th>
              <th scope="col" class="numeric-column" :aria-sort="sortKey === 'total' ? (sortAscending ? 'ascending' : 'descending') : 'none'"><button type="button" @click="sortOrders('total')">Total <ArrowUpDown class="sort-icon" :size="13" :stroke-width="1.8" aria-hidden="true" /></button></th>
              <th scope="col" class="numeric-column" :aria-sort="sortKey === 'balance' ? (sortAscending ? 'ascending' : 'descending') : 'none'"><button type="button" @click="sortOrders('balance')">Balance <ArrowUpDown class="sort-icon" :size="13" :stroke-width="1.8" aria-hidden="true" /></button></th>
              <th scope="col">Commercial</th>
              <th scope="col">Fulfillment</th>
              <th scope="col">Payment</th>
              <th scope="col" :aria-sort="sortKey === 'priority' ? (sortAscending ? 'ascending' : 'descending') : 'none'"><button type="button" @click="sortOrders('priority')">Priority <ArrowUpDown class="sort-icon" :size="13" :stroke-width="1.8" aria-hidden="true" /></button></th>
              <th scope="col"><span class="sr-only">Open</span></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in filteredOrders" :key="order.id" :class="{ 'is-selected': selectedOrderId === order.id }" tabindex="0" @click="selectOrder(order.id)" @dblclick="openOrder(order.id)" @keydown.enter="openOrder(order.id)">
              <td><button class="order-link" type="button" @click.stop="openOrder(order.id)">{{ order.id }}</button></td>
              <td><span class="table-primary">{{ order.customer }}</span><span class="table-secondary">{{ order.customerDetail.split(' · ')[0] }}</span></td>
              <td><span class="table-primary item-summary">{{ order.itemSummary }}</span><span class="table-secondary">{{ order.items.length }} line items</span></td>
              <td>{{ formatDate(order.createdAt) }}</td>
              <td :class="{ 'text-danger': order.isOverdue }">{{ formatDateTime(order.promisedAt ?? '') }}</td>
              <td class="numeric-column table-money">{{ currency(order.totalRial) }}</td>
              <td class="numeric-column table-money" :class="{ 'text-danger': order.totalRial - order.paidRial > 0 }">{{ currency(order.totalRial - order.paidRial) }}</td>
              <td><StatusBadge :label="order.commercialStatus" :tone="commercialTone(order.commercialStatus)" /></td>
              <td><StatusBadge :label="order.fulfillmentStatus" :tone="fulfillmentTone(order.fulfillmentStatus)" /></td>
              <td><StatusBadge :label="order.paymentStatus" :tone="paymentTone(order.paymentStatus)" /></td>
              <td><StatusBadge :label="order.priority" :tone="priorityTone(order.priority)" /></td>
              <td><button class="row-open-button" type="button" aria-label="Open order" @click.stop="openOrder(order.id)"><ArrowRight :size="16" :stroke-width="1.8" aria-hidden="true" /></button></td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-else class="orders-empty">
        <div class="empty-workspace-icon" aria-hidden="true"><SearchX :size="21" :stroke-width="1.8" /></div>
        <h2>No orders match these filters</h2>
        <p>Try a different search or clear the filters to see the full operational queue.</p>
        <button class="button button-secondary" type="button" @click="clearFilters">Clear filters</button>
      </div>
    </section>
  </div>
</template>
