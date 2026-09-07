<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowDown, ArrowLeft, ArrowUp, PackageOpen, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import EmptyState from '../components/EmptyState.vue'
import JalaliDatePicker from '../components/JalaliDatePicker.vue'
import SectionPanel from '../components/SectionPanel.vue'
import SelectField from '../components/SelectField.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import WorkspaceTabs from '../components/WorkspaceTabs.vue'
import type { OrderItemPayload, OrderPayload, OrderRecord } from '../api/orders'
import { ordersApi } from '../api/orders'
import type { CurrencyUnit } from '../utils/currency'
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../utils/currency'
import { formatDateTime } from '../utils/date'
import DocumentMetadataPanel from '../components/DocumentMetadataPanel.vue'
import OrderItemConfigurator from '../components/OrderItemConfigurator.vue'
import OrderInvoicePanel from './OrderInvoicePanel.vue'
import OrderPaymentsPanel from './OrderPaymentsPanel.vue'
import OrderProductionPanel from './OrderProductionPanel.vue'

const props = defineProps<{
  order: OrderRecord
  currencyUnit: CurrencyUnit
  customers: any[]
  services: any[]
  materials: any[]
}>()

const emit = defineEmits<{
  back: []
  notify: [message: string]
  saved: [order: OrderRecord]
}>()

type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate'

const tab = ref('Overview')
const customerId = ref('')
const promisedAt = ref<string | null>(null)
const priority = ref('Normal')
const notes = ref('')
const discountText = ref('')
const editorOpen = ref(false)
const editingItem = ref<any | null>(null)
const saving = ref(false)

const tabs = ['Overview', 'Items', 'Production', 'Payments', 'Invoices', 'Files', 'History']
const commercialOptions = ['Draft', 'Confirmed', 'Closed', 'Cancelled'].map((value) => ({ label: value, value }))
const priorityOptions = ['Urgent', 'High', 'Normal', 'Low'].map((value) => ({ label: value, value }))
const customer = computed(() => props.customers.find((value) => value.id === customerId.value))
const customerOptions = computed(() => [
  { label: 'Walk-in customer', value: '' },
  ...props.customers
    .filter((value) => value.active || value.id === customerId.value)
    .map((value) => ({ label: value.name, value: value.id })),
])

watch(() => props.order, (order) => {
  customerId.value = order.customerId || ''
  promisedAt.value = order.promisedAt
  priority.value = order.priority
  notes.value = order.notes
  discountText.value = formatMoneyInput(order.discountRial, props.currencyUnit)
}, { immediate: true })

function tone(value: string): Tone {
  return value === 'Confirmed' || value === 'In Production'
    ? 'blue'
    : value === 'Closed' || value === 'Delivered' || value === 'Paid' || value === 'Ready'
      ? 'green'
      : value === 'Cancelled'
        ? 'red'
        : value === 'Partially Paid'
          ? 'amber'
          : 'slate'
}

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}

function payload(): OrderPayload {
  return {
    customerId: customerId.value,
    promisedAt: promisedAt.value,
    priority: priority.value,
    notes: notes.value,
    discountRial: parseMoneyInput(discountText.value, props.currencyUnit) || 0,
  }
}

async function saveMetadata() {
  saving.value = true
  try {
    const result = await ordersApi.update(props.order.id, payload())
    emit('saved', result)
    emit('notify', 'Order details saved')
  } catch (error) {
    emit('notify', String(error))
  } finally {
    saving.value = false
  }
}

async function updateDiscount() {
  saving.value = true
  try {
    const amount = parseMoneyInput(discountText.value, props.currencyUnit)
    if (amount === null) throw new Error('Enter a valid discount')
    const result = await ordersApi.discount(props.order.id, amount)
    emit('saved', result)
    emit('notify', 'Discount applied')
  } catch (error) {
    emit('notify', String(error))
  } finally {
    saving.value = false
  }
}

async function configured(input: OrderItemPayload) {
  saving.value = true
  const replacing = Boolean(editingItem.value)
  try {
    const result = replacing
      ? await ordersApi.replaceItem(props.order.id, editingItem.value.id, input)
      : await ordersApi.addItem(props.order.id, input)
    emit('saved', result)
    editorOpen.value = false
    editingItem.value = null
    emit('notify', replacing ? 'Item replaced' : 'Item added')
  } catch (error) {
    emit('notify', String(error))
  } finally {
    saving.value = false
  }
}

async function remove(item: any) {
  try {
    emit('saved', await ordersApi.removeItem(props.order.id, item.id))
    emit('notify', 'Item removed')
  } catch (error) {
    emit('notify', String(error))
  }
}

async function move(item: any, delta: number) {
  const items = [...props.order.items].sort((a, b) => a.position - b.position)
  const index = items.findIndex((value) => value.id === item.id)
  const next = index + delta
  if (next < 0 || next >= items.length) return
  ;[items[index], items[next]] = [items[next], items[index]]
  try {
    emit('saved', await ordersApi.reorderItems(props.order.id, items.map((value) => value.id)))
    emit('notify', 'Item order updated')
  } catch (error) {
    emit('notify', String(error))
  }
}

async function changeCommercial(value: string) {
  try {
    emit('saved', await ordersApi.commercialStatus(props.order.id, value))
    emit('notify', 'Commercial status updated')
  } catch (error) {
    emit('notify', String(error))
  }
}

function snapshot(item: any, key: string) {
  try {
    return JSON.parse(item[key] || '[]')
  } catch {
    return []
  }
}
</script>

<template>
  <div class="space-y-4">
    <WorkspaceStickyStack :flush="true">
      <header class="order-workspace-header">
        <div class="order-workspace-header__main">
          <button class="btn btn-ghost gap-2" type="button" @click="emit('back')">
            <ArrowLeft :size="16" aria-hidden="true" />
            <span>Orders</span>
          </button>
          <div class="order-workspace-header__title min-w-0">
            <p>Sales / order workspace</p>
            <div class="flex min-w-0 flex-wrap items-center gap-2">
              <h1 class="truncate">{{ order.orderNumber }}</h1>
              <StatusBadge :label="order.commercialStatus" :tone="tone(order.commercialStatus)" />
            </div>
            <p class="truncate">{{ order.customerName || 'Walk-in customer' }} · {{ order.items.length }} line items</p>
          </div>
        </div>

        <div class="order-workspace-header__actions flex w-full flex-wrap items-center justify-end gap-2 sm:w-auto">
          <SelectField
            class="w-40"
            :model-value="order.commercialStatus"
            :options="commercialOptions"
            aria-label="Commercial status"
            @update:model-value="changeCommercial"
          />
          <button class="btn btn-primary gap-2" type="button" :disabled="saving" @click="saveMetadata">
            {{ saving ? 'Saving…' : 'Save order' }}
          </button>
        </div>
      </header>

      <div class="grid gap-2 rounded-box border border-base-300 bg-base-100 p-3 sm:grid-cols-2 xl:grid-cols-4">
        <div class="min-w-0">
          <span class="block text-xs text-base-content/50">Customer</span>
          <strong class="block truncate text-sm">{{ order.customerName || 'Walk-in customer' }}</strong>
          <span class="block truncate text-xs text-base-content/60">{{ customer?.phone || order.customerPhone || 'No contact details' }}</span>
        </div>
        <div>
          <span class="block text-xs text-base-content/50">Created</span>
          <strong class="block text-sm">{{ formatDateTime(order.createdAt) }}</strong>
          <span class="block text-xs text-base-content/60">Jalali presentation</span>
        </div>
        <div>
          <span class="block text-xs text-base-content/50">Promised date</span>
          <JalaliDatePicker v-model="promisedAt" placeholder="Set promised date" />
        </div>
        <div class="min-w-0">
          <span class="block text-xs text-base-content/50">Priority</span>
          <SelectField v-model="priority" :options="priorityOptions" aria-label="Priority" />
        </div>
      </div>

      <WorkspaceTabs class="mt-2" :tabs="tabs" :active-tab="tab" @change="tab = $event" />
    </WorkspaceStickyStack>

    <section v-if="tab === 'Overview'" class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(18rem,26rem)]">
      <SectionPanel title="Order details" subtitle="Customer, delivery context, and notes.">
        <div class="grid gap-4 p-4">
          <SelectField v-model="customerId" label="Customer" :options="customerOptions" />
          <label class="form-control gap-1">
            <span class="text-xs text-base-content/60">Notes</span>
            <textarea v-model="notes" class="textarea textarea-bordered w-full min-w-0" rows="4" placeholder="Order notes" />
          </label>
          <div class="flex justify-end">
            <button class="btn btn-primary gap-2" type="button" :disabled="saving" @click="saveMetadata">
              {{ saving ? 'Saving…' : 'Save details' }}
            </button>
          </div>
        </div>
      </SectionPanel>

      <SectionPanel title="Order summary" subtitle="Derived from the saved order and pricing snapshots.">
        <div class="grid grid-cols-2 gap-3 p-4">
          <div class="rounded-box bg-base-200 p-3"><span class="block text-xs text-base-content/60">Subtotal</span><strong class="mt-1 block text-sm">{{ money(order.subtotalRial) }}</strong></div>
          <div class="rounded-box bg-base-200 p-3"><span class="block text-xs text-base-content/60">Discount</span><strong class="mt-1 block text-sm">{{ money(order.discountRial) }}</strong></div>
          <div class="rounded-box bg-base-200 p-3"><span class="block text-xs text-base-content/60">Total</span><strong class="mt-1 block text-primary text-lg">{{ money(order.totalRial) }}</strong></div>
          <div class="rounded-box bg-base-200 p-3"><span class="block text-xs text-base-content/60">Estimated cost</span><strong class="mt-1 block text-sm">{{ money(order.estimatedCostRial) }}</strong></div>
        </div>
        <div class="border-t border-base-300 p-4">
          <label class="form-control gap-1">
            <span class="text-xs text-base-content/60">Order discount</span>
            <input
              class="input input-bordered w-full min-w-0"
              :value="discountText"
              inputmode="decimal"
              :placeholder="`Amount in ${props.currencyUnit}`"
              @input="discountText = ($event.target as HTMLInputElement).value"
              @blur="discountText = formatMoneyInput(parseMoneyInput(discountText, props.currencyUnit) || 0, props.currencyUnit)"
            />
          </label>
          <button class="btn btn-outline mt-3 w-full" type="button" :disabled="saving" @click="updateDiscount">Apply discount</button>
        </div>
      </SectionPanel>
    </section>

    <SectionPanel v-else-if="tab === 'Items'" title="Configured items" subtitle="Each item keeps its accepted service and pricing snapshot.">
      <template #action>
        <button v-if="!editorOpen" class="btn btn-primary btn-sm gap-2" type="button" @click="editorOpen = true">
          <Plus :size="15" aria-hidden="true" />
          <span>Add service item</span>
        </button>
      </template>

      <OrderItemConfigurator
        v-if="editorOpen"
        :services="services"
        :materials="materials"
        :currency-unit="currencyUnit"
        :initial="editingItem"
        @configured="configured"
        @cancel="editorOpen = false; editingItem = null"
      />
      <EmptyState
        v-else-if="!order.items.length"
        title="No configured items"
        description="Add a service item to calculate pricing and prepare production."
      >
        <template #icon><PackageOpen :size="22" aria-hidden="true" /></template>
        <template #action>
          <button class="btn btn-primary btn-sm mt-3 gap-2" type="button" @click="editorOpen = true">
            <Plus :size="15" aria-hidden="true" />
            <span>Add service item</span>
          </button>
        </template>
      </EmptyState>
      <div v-else class="divide-y divide-base-300">
        <article v-for="item in [...order.items].sort((a, b) => a.position - b.position)" :key="item.id" class="grid min-w-0 gap-3 p-4 lg:grid-cols-[auto_minmax(0,1fr)_auto] lg:items-center">
          <div class="grid size-8 shrink-0 place-items-center rounded-box bg-base-200 text-xs font-semibold text-base-content/60">{{ item.position + 1 }}</div>
          <div class="min-w-0">
            <strong class="block truncate">{{ item.serviceName }}</strong>
            <span class="block text-sm text-base-content/70">{{ item.quantity }} {{ item.quantityUnit }}<template v-if="item.notes"> · {{ item.notes }}</template></span>
            <small class="block text-xs text-base-content/50">Accepted snapshot · {{ snapshot(item, 'resolvedParametersJson').length }} parameters · {{ snapshot(item, 'costBreakdownJson').length }} cost lines</small>
          </div>
          <div class="flex min-w-0 flex-wrap items-center justify-between gap-3 lg:justify-end">
            <div class="text-xs"><span class="block text-base-content/50">Cost {{ money(item.estimatedCostRial) }}</span><strong class="block text-sm text-primary">{{ money(item.sellingPriceRial) }}</strong></div>
            <div class="flex items-center gap-1">
              <button class="btn btn-ghost btn-square btn-sm" type="button" aria-label="Move item up" title="Move item up" @click="move(item, -1)"><ArrowUp :size="14" aria-hidden="true" /></button>
              <button class="btn btn-ghost btn-square btn-sm" type="button" aria-label="Move item down" title="Move item down" @click="move(item, 1)"><ArrowDown :size="14" aria-hidden="true" /></button>
              <button class="btn btn-ghost btn-square btn-sm" type="button" aria-label="Reconfigure item" title="Reconfigure item" @click="editingItem = item; editorOpen = true"><Pencil :size="14" aria-hidden="true" /></button>
              <button class="btn btn-ghost btn-square btn-sm text-error" type="button" aria-label="Remove item" title="Remove item" @click="remove(item)"><Trash2 :size="14" aria-hidden="true" /></button>
            </div>
          </div>
        </article>
      </div>
    </SectionPanel>

    <OrderInvoicePanel v-else-if="tab === 'Invoices'" :order="order" :currency-unit="currencyUnit" @notify="emit('notify', $event)" @saved="emit('saved', $event)" />
    <DocumentMetadataPanel v-else-if="tab === 'Files'" owner-type="order" :owner-id="order.id" :protected-context="order.commercialStatus !== 'Draft'" @notify="emit('notify', $event)" />
    <OrderProductionPanel v-else-if="tab === 'Production'" :order="order" :currency-unit="currencyUnit" />
    <OrderPaymentsPanel v-else-if="tab === 'Payments'" :order="order" :currency-unit="currencyUnit" @notify="emit('notify', $event)" @saved="emit('saved', $event)" />
    <SectionPanel v-else title="History" subtitle="Persisted order history will appear here.">
      <EmptyState title="History is not available yet" description="The saved order remains unchanged while this workspace is being expanded.">
        <template #icon><PackageOpen :size="22" aria-hidden="true" /></template>
      </EmptyState>
    </SectionPanel>
  </div>
</template>

<style scoped>
.order-workspace-header {
  align-items: center !important;
}

.order-workspace-header__main {
  display: flex !important;
  min-width: 0;
  flex: 1 1 auto;
  flex-direction: row !important;
  align-items: center;
  gap: 0.75rem;
}

.order-workspace-header__title {
  display: flex !important;
  min-width: 0;
  flex-direction: column !important;
  gap: 0.125rem;
}

.order-workspace-header__title > p:first-child {
  color: var(--color-primary);
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.35;
}

.order-workspace-header__title > p:last-child {
  color: color-mix(in oklab, var(--color-base-content) 65%, transparent);
  font-size: 0.8125rem;
  line-height: 1.5;
}

.order-workspace-header__actions {
  flex: 0 0 auto;
}

@media (max-width: 48rem) {
  .order-workspace-header {
    align-items: stretch !important;
  }

  .order-workspace-header__main {
    width: 100%;
  }

  .order-workspace-header__title {
    flex: 1 1 auto;
  }

  .order-workspace-header__actions {
    width: 100% !important;
    justify-content: flex-start !important;
  }

  .order-workspace-header__actions > :deep(.form-control) {
    flex: 1 1 10rem;
  }
}
</style>
