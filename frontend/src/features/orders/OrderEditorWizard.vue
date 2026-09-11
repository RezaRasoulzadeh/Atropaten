<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowLeft, CheckCheck, ClipboardCheck, CircleHelp, Edit3, FileText, Package, Plus, Save, Trash2, UserRound } from 'lucide-vue-next'
import FormField from '../../components/ui/FormField.vue'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import SelectField from '../../components/ui/SelectField.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import FormSection from '../../components/ui/FormSection.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import OrderItemConfigurator from '../sales/OrderItemConfigurator.vue'
import type { OrderItemPayload, OrderRecord } from '../../api/orders'
import { ordersApi } from '../../api/orders'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../../utils/currency'
import { reportError } from '../../composables/useWorkspaceActions'

type DraftItem = { id: string; payload: OrderItemPayload }

const props = defineProps<{
  order: OrderRecord
  customers: any[]
  services: any[]
  materials: any[]
  machines: any[]
  currencyUnit: CurrencyUnit
  busy?: boolean
}>()

const emit = defineEmits<{
  cancel: []
  saved: [order: OrderRecord]
  notify: [message: string]
}>()

const activeStep = ref(1)
const customerId = ref(props.order.customerId || '')
const promisedAt = ref<string | null>(props.order.promisedAt)
const priority = ref(props.order.priority || 'Normal')
const notes = ref(props.order.notes || '')
const discountText = ref('')
const draftItems = ref<DraftItem[]>([])
const editingIndex = ref<number | null>(null)
const itemEditorOpen = ref(false)
const saving = ref(false)
const validationAttempted = ref(false)

const steps = [
  { number: 1, title: 'Customer', description: 'Customer, priority, delivery' },
  { number: 2, title: 'Items', description: 'Services and quantities' },
  { number: 3, title: 'Pricing', description: 'Review totals and discount' },
  { number: 4, title: 'Review', description: 'Confirm before saving' },
]
const priorityOptions = ['Urgent', 'High', 'Normal', 'Low'].map((value) => ({ label: value, value }))
const customerOptions = computed(() => [
  { label: 'Walk-in customer', value: '' },
  ...props.customers
    .filter((value) => value.active || value.id === customerId.value)
    .map((value) => ({ label: value.name, value: value.id })),
])
const selectedCustomer = computed(() => props.customers.find((value) => value.id === customerId.value))
const draftTotal = computed(() => props.order.totalRial || 0)
const draftItemTotal = computed(() => draftItems.value.reduce((total, item) => {
  const service = props.services.find((value) => value.id === item.payload.serviceId)
  const fixed = service?.pricingRule?.fixedPriceRial || 0
  const quantity = Number(item.payload.quantity) || 1
  return total + fixed * quantity
}, 0))
const discountRial = computed(() => parseMoneyInput(discountText.value, props.currencyUnit) || 0)
const previewTotal = computed(() => Math.max(draftItemTotal.value - discountRial.value, 0) || draftTotal.value)

function stepClass(number: number) {
  if (number === activeStep.value) return 'wizard-step-active'
  if (number < activeStep.value) return 'wizard-step-complete'
  return 'wizard-step-idle'
}

function serviceName(item: DraftItem) {
  return props.services.find((value) => value.id === item.payload.serviceId)?.name || 'Service item'
}

function serviceCode(item: DraftItem) {
  return props.services.find((value) => value.id === item.payload.serviceId)?.code || 'No code'
}

function itemInitial(item: DraftItem | undefined) {
  if (!item) return undefined
  return {
    serviceId: item.payload.serviceId,
    quantity: item.payload.quantity,
    quantityUnit: item.payload.quantityUnit,
    notes: item.payload.notes,
    resolvedParametersJson: JSON.stringify(Object.entries(item.payload.parameters).map(([key, value]) => ({ key, value }))),
  }
}

function openItemEditor(index: number | null = null) {
  editingIndex.value = index
  itemEditorOpen.value = true
}

function saveItem(payload: OrderItemPayload) {
  if (editingIndex.value === null) {
    draftItems.value.push({ id: `draft-item-${Date.now()}-${Math.random().toString(16).slice(2)}`, payload })
  } else {
    draftItems.value[editingIndex.value] = { ...draftItems.value[editingIndex.value], payload }
  }
  editingIndex.value = null
  itemEditorOpen.value = false
}

function removeItem(index: number) {
  draftItems.value.splice(index, 1)
}

function next() {
  validationAttempted.value = true
  if (activeStep.value === 1) {
    const form = document.querySelector<HTMLFormElement>('#order-editor-wizard')
    if (form && !form.reportValidity()) return
  }
  if (activeStep.value < steps.length) activeStep.value += 1
}

function previous() {
  if (activeStep.value > 1) activeStep.value -= 1
}

function payload() {
  return {
    customerId: customerId.value,
    promisedAt: promisedAt.value,
    priority: priority.value,
    notes: notes.value,
    discountRial: discountRial.value,
  }
}

async function save() {
  if (saving.value || props.busy) return
  saving.value = true
  let saved: OrderRecord | null = null
  try {
    saved = await ordersApi.create(payload())
    for (const item of draftItems.value) {
      saved = await ordersApi.addItem(saved.id, item.payload)
    }
    emit('saved', saved)
    emit('notify', 'Order created')
  } catch (error) {
    if (saved) emit('saved', saved)
    reportError(error)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="order-wizard w-full min-w-0 flex min-h-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Order editor">
    <header class="order-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">Add order</h1>
        <p class="mt-1 text-xs leading-4 text-base-content/65">Create a customer order, configure its services, and review the price before saving.</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Orders' }, { label: 'Add order', current: true }]" @navigate="emit('cancel')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="saving || busy" @click="emit('cancel')">Cancel</button>
        <button class="btn btn-success gap-2" type="button" :disabled="saving || busy" @click="save"><Save :size="16" aria-hidden="true" />{{ saving ? 'Saving…' : 'Save order' }}</button>
      </div>
    </header>

    <div class="order-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="order-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200">
        <nav class="order-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible" aria-label="Order setup steps">
          <button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" @click="activeStep = step.number">
            <span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span>
            <span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ step.title }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ step.description }}</small></span>
          </button>
        </nav>
      </aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="order-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
          <form id="order-editor-wizard" class="min-w-0" @submit.prevent="next">
            <section v-if="activeStep === 1" class="min-w-0 space-y-6">
              <FormSection title="Customer and delivery" description="Choose who this order is for and set the operational context.">
                <FormField class="gap-1 sm:col-span-2"><span>Customer</span><SelectField v-model="customerId" :options="customerOptions" aria-label="Customer" /></FormField>
                <FormField class="gap-1"><span>Priority</span><SelectField v-model="priority" :options="priorityOptions" aria-label="Priority" /></FormField>
                <FormField class="gap-1"><span>Promised date</span><JalaliDatePicker v-model="promisedAt" placeholder="Set promised date" /></FormField>
                <FormField class="gap-1 sm:col-span-2"><span>Order notes</span><AppTextarea v-model="notes" rows="6" placeholder="Delivery instructions, preferences, or internal notes" /></FormField>
              </FormSection>
            </section>

            <section v-else-if="activeStep === 2" class="min-w-0 space-y-4">
              <div class="flex flex-wrap items-start justify-between gap-3"><div><h2 class="text-lg font-semibold">Configure order items</h2><p class="mt-1 text-sm text-base-content/60">Add services and calculate their accepted pricing before creating the order.</p></div><button v-if="!itemEditorOpen" class="btn btn-primary btn-sm gap-2" type="button" @click="openItemEditor()"><Plus :size="15" aria-hidden="true" />Add service</button></div>
              <OrderItemConfigurator v-if="itemEditorOpen" :services="services" :materials="materials" :machines="machines" :currency-unit="currencyUnit" :initial="itemInitial(editingIndex === null ? undefined : draftItems[editingIndex])" :busy="saving || busy" @configured="saveItem" @cancel="itemEditorOpen = false; editingIndex = null" />
              <div v-else-if="draftItems.length" class="space-y-2">
                <div v-for="(item, index) in draftItems" :key="item.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-100 px-3 py-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Package :size="18" aria-hidden="true" /></span><div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ serviceName(item) }}</strong><span class="block truncate text-xs text-base-content/55">{{ serviceCode(item) }} · {{ item.payload.quantity || '1' }} {{ item.payload.quantityUnit }}</span></div><button class="btn btn-ghost btn-square btn-sm" type="button" aria-label="Edit service item" @click="openItemEditor(index)"><Edit3 :size="14" aria-hidden="true" /></button><button class="btn btn-ghost btn-error btn-square btn-sm" type="button" aria-label="Remove service item" @click="removeItem(index)"><Trash2 :size="14" aria-hidden="true" /></button></div>
              </div>
              <section v-else class="flex min-h-56 flex-col items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><span class="grid size-12 place-items-center rounded-full bg-primary/15 text-primary"><Package :size="23" aria-hidden="true" /></span><h3 class="mt-3 text-sm font-semibold">No services added</h3><p class="mt-1 max-w-sm text-xs leading-5 text-base-content/60">Add at least one service when the order is ready for pricing and production.</p><button class="btn btn-primary btn-sm mt-4 gap-2" type="button" @click="openItemEditor()"><Plus :size="15" aria-hidden="true" />Add service</button></section>
            </section>

            <section v-else-if="activeStep === 3" class="min-w-0 space-y-5">
              <div><h2 class="text-lg font-semibold">Review pricing</h2><p class="mt-1 text-sm text-base-content/60">Pricing is calculated from the configured service items after the order is saved.</p></div>
              <div class="grid gap-3 sm:grid-cols-3"><div class="rounded-box border border-base-300 bg-base-100 p-4"><span class="block text-xs text-base-content/60">Items</span><strong class="mt-1 block text-lg">{{ draftItems.length }}</strong></div><div class="rounded-box border border-base-300 bg-base-100 p-4"><span class="block text-xs text-base-content/60">Estimated preview</span><strong class="mt-1 block text-sm tabular-nums">{{ previewTotal ? formatMoney(previewTotal, currencyUnit) : 'Calculated on save' }}</strong></div><FormField class="gap-1 rounded-box border border-base-300 bg-base-100 p-4"><span class="text-xs text-base-content/60">Discount</span><AppInput v-model="discountText" :money="currencyUnit" inputmode="decimal" placeholder="Optional discount" @blur="discountText = discountRial ? formatMoneyInput(discountRial, currencyUnit) : ''" /></FormField></div>
              <div class="rounded-box border border-warning/30 bg-warning/5 p-4 text-sm leading-6 text-base-content/70"><strong class="text-warning">Pricing checkpoint.</strong> Each service item must be calculated in the item configurator. The backend then stores the accepted pricing snapshot on the order.</div>
              <button class="btn btn-outline btn-sm gap-2" type="button" @click="activeStep = 2"><Edit3 :size="14" aria-hidden="true" />Edit items</button>
            </section>

            <section v-else class="min-w-0 space-y-5">
              <div><h2 class="text-lg font-semibold">Confirm order</h2><p class="mt-1 text-sm text-base-content/60">Check the order context before creating this draft order.</p></div>
              <dl class="divide-y divide-base-300 rounded-box border border-base-300 bg-base-100 px-4"><div class="flex justify-between gap-4 py-3 text-sm"><dt class="text-base-content/60">Customer</dt><dd class="text-end">{{ selectedCustomer?.name || 'Walk-in customer' }}</dd></div><div class="flex justify-between gap-4 py-3 text-sm"><dt class="text-base-content/60">Priority</dt><dd>{{ priority }}</dd></div><div class="flex justify-between gap-4 py-3 text-sm"><dt class="text-base-content/60">Promised date</dt><dd>{{ promisedAt ? 'Set' : 'Not set' }}</dd></div><div class="flex justify-between gap-4 py-3 text-sm"><dt class="text-base-content/60">Services</dt><dd>{{ draftItems.length }}</dd></div></dl>
              <div v-if="!draftItems.length" class="rounded-box border border-dashed border-warning/40 bg-warning/5 p-4 text-sm leading-6 text-base-content/70"><strong class="text-warning">No service items yet.</strong> You can save this as an empty draft and add items later.</div>
              <div v-if="notes" class="rounded-box border border-base-300 bg-base-100 p-4"><h3 class="text-sm font-semibold">Notes</h3><p class="mt-2 whitespace-pre-wrap text-sm leading-6 text-base-content/70">{{ notes }}</p></div>
            </section>
          </form>
        </section>

        <aside class="order-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto">
          <div class="space-y-4">
            <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300"><div class="absolute inset-0 bg-gradient-to-br from-primary/20 via-base-300 to-base-200" aria-hidden="true"></div><div class="relative z-10 flex min-h-52 flex-col justify-between p-4"><div class="grid size-11 place-items-center rounded-box border border-primary/30 bg-base-100/70 text-primary"><ClipboardCheck :size="23" aria-hidden="true" /></div><div><h2 class="truncate text-xl font-semibold">{{ selectedCustomer?.name || 'New order' }}</h2><p class="mt-1 text-sm text-base-content/60">Draft order</p><StatusBadge class="mt-3" label="Draft" tone="slate" /></div></div></div>
            <div class="rounded-box border border-base-300 bg-base-100 p-4"><h3 class="text-xs font-semibold uppercase tracking-wide text-base-content/60">Order summary</h3><dl class="mt-3 divide-y divide-base-300/70"><div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Customer</dt><dd class="max-w-[10rem] truncate">{{ selectedCustomer?.name || 'Walk-in customer' }}</dd></div><div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Items</dt><dd>{{ draftItems.length }}</dd></div><div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Priority</dt><dd>{{ priority }}</dd></div><div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Total</dt><dd class="font-semibold tabular-nums text-primary">{{ previewTotal ? formatMoney(previewTotal, currencyUnit) : 'Calculated on save' }}</dd></div></dl></div>
            <div v-if="draftItems.length" class="rounded-box border border-base-300 bg-base-100 p-4"><div class="flex items-center gap-2"><FileText :size="17" class="text-primary" aria-hidden="true" /><h3 class="text-sm font-semibold">Configured services</h3></div><div class="mt-3 space-y-2"><div v-for="item in draftItems" :key="item.id" class="flex min-w-0 items-center gap-2 text-xs"><Package :size="14" class="shrink-0 text-base-content/50" aria-hidden="true" /><span class="min-w-0 flex-1 truncate">{{ serviceName(item) }}</span><span class="shrink-0 text-base-content/55">{{ item.payload.quantity || '1' }} {{ item.payload.quantityUnit }}</span></div></div></div>
            <div class="flex items-start gap-2 border-t border-base-300 pt-3 text-sm"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">The order starts as a draft. Confirm it later when the customer and pricing are ready.</p></div>
          </div>
        </aside>
      </div>
    </div>

    <footer class="flex min-w-0 items-center justify-between gap-3 border-t border-base-300 px-1 pt-3"><button class="btn btn-ghost btn-sm gap-2" type="button" :disabled="activeStep === 1 || busy || saving" @click="previous"><ArrowLeft :size="14" aria-hidden="true" />Back</button><span class="text-xs text-base-content/55">Step {{ activeStep }} of {{ steps.length }}</span><button v-if="activeStep < steps.length" class="btn btn-primary btn-sm" type="button" @click="next">Continue</button><button v-else class="btn btn-success btn-sm gap-2" type="button" :disabled="busy || saving" @click="save"><Save :size="14" aria-hidden="true" />{{ saving ? 'Saving…' : 'Create order' }}</button></footer>
  </div>
</template>
