<script setup lang="ts">
import { ref, watch } from 'vue'
import { CalendarDays, ClipboardList, Edit3, FileText, Package, UserRound } from 'lucide-vue-next'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { OrderRecord } from '../../api/orders'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'

type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate'
type DetailTab = 'overview' | 'items'

const props = defineProps<{
  order: OrderRecord
  currencyUnit: CurrencyUnit
  busy?: boolean
}>()

const emit = defineEmits<{ open: [] }>()
const activeTab = ref<DetailTab>('overview')

watch(() => props.order.id, () => { activeTab.value = 'overview' })

function orderStatus(order: OrderRecord) {
  if (order.commercialStatus === 'Cancelled') return 'Cancelled'
  if (order.commercialStatus === 'Closed') return 'Closed'
  if (order.fulfillmentStatus === 'Delivered') return 'Delivery'
  if (order.fulfillmentStatus === 'In Production' || order.fulfillmentStatus === 'Ready') return 'Production'
  return order.commercialStatus
}

function tone(value: string): Tone {
  if (value === 'Confirmed' || value === 'Production') return 'blue'
  if (value === 'Closed' || value === 'Delivery' || value === 'Delivered' || value === 'Paid' || value === 'Ready') return 'green'
  if (value === 'Cancelled') return 'red'
  if (value === 'Partially Paid' || value === 'Urgent') return 'amber'
  return 'slate'
}

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}

function itemCount(order: OrderRecord) {
  return Array.isArray(order.items) ? order.items.length : 0
}
</script>

<template>
  <section class="order-detail-panel h-auto min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:h-full xl:overflow-y-auto" aria-label="Order preview">
    <div class="border-b border-base-300 p-3 sm:p-5">
      <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300 sm:min-h-60">
        <div class="absolute inset-0 bg-gradient-to-br from-primary/20 via-base-300 to-base-200" aria-hidden="true"></div>
        <div class="absolute -end-10 -top-12 size-48 rounded-full bg-primary/10 blur-2xl" aria-hidden="true"></div>
        <div class="relative z-10 flex min-h-52 flex-col justify-between p-4 sm:min-h-60 sm:p-5">
          <div class="grid size-11 place-items-center rounded-box border border-primary/30 bg-base-100/70 text-primary">
            <ClipboardList :size="24" :stroke-width="1.7" aria-hidden="true" />
          </div>
          <div class="min-w-0">
            <div class="flex min-w-0 items-center gap-3">
              <h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ order.orderNumber }}</h2>
              <StatusBadge class="shrink-0" :label="orderStatus(order)" :tone="tone(orderStatus(order))" />
            </div>
            <p class="mt-2 truncate text-sm text-base-content/65">{{ order.customerName || 'Walk-in customer' }}</p>
            <p class="mt-1 text-xs text-base-content/50">Updated {{ formatDateTime(order.updatedAt || order.createdAt) }}</p>
          </div>
        </div>
      </div>

      <button class="btn btn-primary mt-2 w-full gap-2" type="button" :disabled="busy" @click="emit('open')">
        <Edit3 :size="15" aria-hidden="true" />Open order
      </button>

      <div class="mt-4 grid min-w-0 divide-y divide-base-300 border-y border-base-300 sm:mt-5 sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:first:pl-0">
          <FileText :size="20" class="shrink-0 text-primary" aria-hidden="true" />
          <div class="min-w-0"><span class="block text-xs text-base-content/55">Order total</span><strong class="block truncate text-sm tabular-nums">{{ money(order.totalRial) }}</strong></div>
        </div>
        <div class="flex items-center gap-3 py-3 sm:px-3">
          <UserRound :size="20" class="shrink-0 text-primary" aria-hidden="true" />
          <div class="min-w-0"><span class="block text-xs text-base-content/55">Payment</span><strong class="block truncate text-sm">{{ order.paymentStatus }}</strong></div>
        </div>
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:pr-0">
          <CalendarDays :size="20" class="shrink-0 text-primary" aria-hidden="true" />
          <div class="min-w-0"><span class="block text-xs text-base-content/55">Priority</span><strong class="block truncate text-sm">{{ order.priority }}</strong></div>
        </div>
      </div>
    </div>

    <nav class="flex min-w-0 overflow-x-auto border-b border-base-300 px-2" aria-label="Order preview tabs">
      <button class="shrink-0 border-b-2 px-3 py-3 text-sm transition-colors" :class="activeTab === 'overview' ? 'border-primary text-primary' : 'border-transparent text-base-content/65 hover:border-base-content/30 hover:text-base-content'" type="button" @click="activeTab = 'overview'">Overview</button>
      <button class="shrink-0 border-b-2 px-3 py-3 text-sm transition-colors" :class="activeTab === 'items' ? 'border-primary text-primary' : 'border-transparent text-base-content/65 hover:border-base-content/30 hover:text-base-content'" type="button" @click="activeTab = 'items'">Items <span class="ms-1 text-xs text-base-content/50">{{ itemCount(order) }}</span></button>
    </nav>

    <div class="min-w-0 p-3 sm:p-4">
      <div v-if="activeTab === 'overview'" class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4">
          <div class="flex items-start gap-3">
            <span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><UserRound :size="18" aria-hidden="true" /></span>
            <div><h3 class="text-sm font-semibold">Customer and delivery</h3><p class="mt-1 text-xs leading-5 text-base-content/60">The order context used by your team.</p></div>
          </div>
          <dl class="mt-4 divide-y divide-base-300/70 text-sm">
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Customer</dt><dd class="truncate text-end">{{ order.customerName || 'Walk-in customer' }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Contact</dt><dd class="truncate text-end">{{ order.customerPhone || 'No contact details' }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Created</dt><dd class="text-end">{{ formatDateTime(order.createdAt) }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Promised</dt><dd class="text-end">{{ order.promisedAt ? formatDateTime(order.promisedAt) : 'Not set' }}</dd></div>
            <div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Fulfillment</dt><dd class="text-end">{{ order.fulfillmentStatus }}</dd></div>
          </dl>
        </div>

        <div class="rounded-box border border-base-300 bg-base-200/20 p-4">
          <div class="flex items-center justify-between gap-3"><div class="flex items-center gap-2"><Package :size="17" class="text-primary" aria-hidden="true" /><h3 class="text-sm font-semibold">Production</h3></div><strong class="text-sm">{{ order.productionJobCount }} job{{ order.productionJobCount === 1 ? '' : 's' }}</strong></div>
          <div class="mt-3 flex flex-wrap gap-2 text-xs text-base-content/60"><span>{{ order.completedProductionJobs }} completed</span><span>·</span><span>{{ order.inProgressProductionJobs }} in progress</span></div>
        </div>

        <div v-if="order.notes" class="rounded-box border border-base-300 bg-base-200/20 p-4"><h3 class="text-sm font-semibold">Notes</h3><p class="mt-2 whitespace-pre-wrap text-sm leading-6 text-base-content/70">{{ order.notes }}</p></div>
      </div>

      <div v-else class="space-y-2">
        <div v-for="item in order.items" :key="item.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-2.5">
          <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Package :size="18" aria-hidden="true" /></span>
          <div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ item.serviceName || 'Service item' }}</strong><span class="block truncate text-xs text-base-content/55">{{ item.serviceCode || 'No code' }} · {{ item.quantity }} {{ item.quantityUnit }}</span></div>
          <span class="shrink-0 text-sm font-medium tabular-nums">{{ money(item.sellingPriceRial) }}</span>
        </div>
        <div v-if="!order.items.length" class="rounded-box border border-dashed border-base-300 p-8 text-center text-sm text-base-content/60">No configured items.</div>
        <div v-if="order.items.length" class="flex items-center justify-between gap-3 border-t border-base-300 pt-3 text-sm"><span class="text-base-content/60">Total</span><strong class="text-primary tabular-nums">{{ money(order.totalRial) }}</strong></div>
      </div>
    </div>
  </section>
</template>
