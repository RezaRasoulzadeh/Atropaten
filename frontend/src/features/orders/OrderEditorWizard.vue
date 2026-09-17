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
import type { PricingRecord } from '../../api/pricing'
import { ordersApi } from '../../api/orders'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney, formatMoneyInput, parseMoneyInput, sellingPriceTotal } from '../../utils/currency'
import { reportError } from '../../composables/useWorkspaceActions'

type DraftItem = { id: string; payload: OrderItemPayload; preview?: PricingRecord }

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
const hasCompletePricePreview = computed(() => draftItems.value.length > 0 && draftItems.value.every((item) => Boolean(item.preview)))
const draftItemTotal = computed(() => draftItems.value.reduce((total, item) => {
  if (!item.preview) return total
  const quantity = item.preview.batchQuantity ? 1 : Number(item.payload.quantity) || 1
  return total + sellingPriceTotal(item.preview.effectiveSellingPriceRial, quantity, item.preview.roundingStepRial)
}, 0))
const discountRial = computed(() => parseMoneyInput(discountText.value, props.currencyUnit) || 0)
const previewTotal = computed(() => hasCompletePricePreview.value ? Math.max(draftItemTotal.value - discountRial.value, 0) : null)

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

function saveItem(payload: OrderItemPayload, preview?: PricingRecord) {
  if (editingIndex.value === null) {
    draftItems.value.push({ id: `draft-item-${Date.now()}-${Math.random().toString(16).slice(2)}`, payload, preview })
  } else {
    draftItems.value[editingIndex.value] = { ...draftItems.value[editingIndex.value], payload, preview }
  }
  editingIndex.value = null
  itemEditorOpen.value = false
}

function removeItem(index: number) {
  draftItems.value.splice(index, 1)
}

function submit() {
  if (props.busy || saving.value || itemEditorOpen.value) return
  if (activeStep.value < steps.length) next()
  else void save()
}

function next() {
  if (props.busy || saving.value) return
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
    items: draftItems.value.map((item) => item.payload),
  }
}

async function save() {
  if (saving.value || props.busy || itemEditorOpen.value) return
  saving.value = true
  try {
    const saved = await ordersApi.create(payload())
    emit('saved', saved)
    emit('notify', 'Order created')
  } catch (error) {
    reportError(error)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="order-wizard w-full min-w-0 flex min-h-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" :aria-label='$t("Order editor")'>
    <header class="order-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">{{ $t("Add order") }}</h1>
        <p class="mt-1 text-xs leading-4 text-base-content/65">{{ $t("Create a customer order, configure its services, and review the price before saving.") }}</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Orders' }, { label: 'Add order', current: true }]" @navigate="emit('cancel')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="saving || busy" @click="emit('cancel')">{{ $t("Cancel") }}</button>
        <button class="btn btn-success gap-2" type="button" :disabled="saving || busy || itemEditorOpen" @click="save"><Save :size="16" aria-hidden="true" />{{ $ui(saving ? 'Saving…' : 'Save order') }}</button>
      </div>
    </header>

    <div class="order-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="order-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200">
        <nav class="order-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible" :aria-label='$t("Order setup steps")'>
          <button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" :disabled="itemEditorOpen && step.number !== 2" @click="activeStep = step.number">
            <span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span>
            <span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ $ui(step.title) }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ $ui(step.description) }}</small></span>
          </button>
        </nav>
      </aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="order-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
          <form id="order-editor-wizard" class="min-w-0" :aria-busy="busy || saving" @submit.prevent="submit">
            <section v-if="activeStep === 1" class="min-w-0 space-y-6">
              <FormSection :title='$t("Customer and delivery")' :description='$t("Choose who this order is for and set the operational context.")'>
                <FormField class="gap-1 sm:col-span-2"><span>{{ $t("Customer") }}</span><SelectField v-model="customerId" :options="customerOptions" :aria-label='$t("Customer")' /></FormField>
                <FormField class="gap-1"><span>{{ $t("Priority") }}</span><SelectField v-model="priority" :options="priorityOptions" :aria-label='$t("Priority")' /></FormField>
                <FormField class="gap-1"><span>{{ $t("Promised date") }}</span><JalaliDatePicker v-model="promisedAt" :placeholder='$t("Set promised date")' /></FormField>
                <FormField class="gap-1 sm:col-span-2"><span>{{ $t("Order notes") }}</span><AppTextarea v-model="notes" rows="6" :placeholder='$t("Delivery instructions, preferences, or internal notes")' /></FormField>
              </FormSection>
            </section>

            <section v-else-if="activeStep === 2" class="min-w-0 space-y-4">
              <div class="flex flex-wrap items-start justify-between gap-3"><div><h2 class="text-lg font-semibold">{{ $t("Configure order items") }}</h2><p class="mt-1 text-sm text-base-content/60">{{ $t("Add services and calculate their accepted pricing before creating the order.") }}</p></div><button v-if="!itemEditorOpen" class="btn btn-primary btn-sm gap-2" type="button" @click="openItemEditor()"><Plus :size="15" aria-hidden="true" />{{ $t("Add service") }}</button></div>
              <OrderItemConfigurator v-if="itemEditorOpen" :services="services" :materials="materials" :machines="machines" :currency-unit="currencyUnit" :initial="itemInitial(editingIndex === null ? undefined : draftItems[editingIndex])" :busy="saving || busy" @configured="saveItem" @cancel="itemEditorOpen = false; editingIndex = null" />
              <div v-else-if="draftItems.length" class="space-y-2">
                <div v-for="(item, index) in draftItems" :key="item.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-100 px-3 py-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Package :size="18" aria-hidden="true" /></span><div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ serviceName(item) }}</strong><span class="block truncate text-xs text-base-content/55">{{ serviceCode(item) }} · {{ $ui(item.payload.quantity || '1') }} {{ item.payload.quantityUnit }}</span></div><button class="btn btn-ghost btn-square btn-sm" type="button" :aria-label='$t("Edit service item")' @click="openItemEditor(index)"><Edit3 :size="14" aria-hidden="true" /></button><button class="btn btn-ghost btn-error btn-square btn-sm" type="button" :aria-label='$t("Remove service item")' @click="removeItem(index)"><Trash2 :size="14" aria-hidden="true" /></button></div>
              </div>
              <section v-else class="flex min-h-56 flex-col items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><span class="grid size-12 place-items-center rounded-full bg-primary/15 text-primary"><Package :size="23" aria-hidden="true" /></span><h3 class="mt-3 text-sm font-semibold">{{ $t("No services added") }}</h3><p class="mt-1 max-w-sm text-xs leading-5 text-base-content/60">{{ $t("Add at least one service when the order is ready for pricing and production.") }}</p><button class="btn btn-primary btn-sm mt-4 gap-2" type="button" @click="openItemEditor()"><Plus :size="15" aria-hidden="true" />{{ $t("Add service") }}</button></section>
            </section>

            <section v-else-if="activeStep === 3" class="min-w-0 space-y-5">
              <div><h2 class="text-lg font-semibold">{{ $t("Review pricing") }}</h2><p class="mt-1 text-sm text-base-content/60">{{ $t("Review the item previews and discount before the server validates and saves the order.") }}</p></div>
              <div class="grid gap-3 sm:grid-cols-3"><div class="rounded-box border border-base-300 bg-base-100 p-4"><span class="block text-xs text-base-content/60">{{ $t("Items") }}</span><strong class="mt-1 block text-lg">{{ draftItems.length }}</strong></div><div class="rounded-box border border-base-300 bg-base-100 p-4"><span class="block text-xs text-base-content/60">{{ $t("Estimated preview") }}</span><strong class="mt-1 block text-sm tabular-nums">{{ $ui(previewTotal !== null ? formatMoney(previewTotal, currencyUnit) : 'Calculated on save') }}</strong></div><FormField class="gap-1 rounded-box border border-base-300 bg-base-100 p-4"><span class="text-xs text-base-content/60">{{ $t("Discount") }}</span><AppInput v-model="discountText" :money="currencyUnit" inputmode="decimal" :placeholder='$t("Optional discount")' @blur="discountText = discountRial ? formatMoneyInput(discountRial, currencyUnit) : ''" /></FormField></div>
              <div class="rounded-box border border-warning/30 bg-warning/5 p-4 text-sm leading-6 text-base-content/70"><strong class="text-warning">{{ $t("Pricing checkpoint.") }}</strong> {{ $t("Item previews are shown above. The server validates and saves all configured service snapshots together.") }}</div>
              <button class="btn btn-outline btn-sm gap-2" type="button" @click="activeStep = 2"><Edit3 :size="14" aria-hidden="true" />{{ $t("Edit items") }}</button>
            </section>

            <section v-else class="min-w-0 space-y-5">
              <div><h2 class="text-lg font-semibold">{{ $t("Confirm order") }}</h2><p class="mt-1 text-sm text-base-content/60">{{ $t("Check the order context before creating this draft order.") }}</p></div>
              <dl class="divide-y divide-base-300 rounded-box border border-base-300 bg-base-100 px-4"><div class="flex justify-between gap-4 py-3 text-sm"><dt class="text-base-content/60">{{ $t("Customer") }}</dt><dd class="text-end">{{ $ui(selectedCustomer?.name || 'Walk-in customer') }}</dd></div><div class="flex justify-between gap-4 py-3 text-sm"><dt class="text-base-content/60">{{ $t("Priority") }}</dt><dd>{{ $ui(priority) }}</dd></div><div class="flex justify-between gap-4 py-3 text-sm"><dt class="text-base-content/60">{{ $t("Promised date") }}</dt><dd>{{ $ui(promisedAt ? 'Scheduled' : 'Not set') }}</dd></div><div class="flex justify-between gap-4 py-3 text-sm"><dt class="text-base-content/60">{{ $t("Services") }}</dt><dd>{{ draftItems.length }}</dd></div></dl>
              <div v-if="!draftItems.length" class="rounded-box border border-dashed border-warning/40 bg-warning/5 p-4 text-sm leading-6 text-base-content/70"><strong class="text-warning">{{ $t("No service items yet.") }}</strong> {{ $t("You can save this as an empty draft and add items later.") }}</div>
              <div v-if="notes" class="rounded-box border border-base-300 bg-base-100 p-4"><h3 class="text-sm font-semibold">{{ $t("Notes") }}</h3><p class="mt-2 whitespace-pre-wrap text-sm leading-6 text-base-content/70">{{ notes }}</p></div>
            </section>
          </form>
        </section>

        <aside class="order-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto">
          <div class="space-y-4">
            <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300"><div class="absolute inset-0 bg-gradient-to-br from-primary/20 via-base-300 to-base-200" aria-hidden="true"></div><div class="relative z-10 flex min-h-52 flex-col justify-between p-4"><div class="grid size-11 place-items-center rounded-box border border-primary/30 bg-base-100/70 text-primary"><ClipboardCheck :size="23" aria-hidden="true" /></div><div><h2 class="truncate text-xl font-semibold">{{ $ui(selectedCustomer?.name || 'New order') }}</h2><p class="mt-1 text-sm text-base-content/60">{{ $t("Draft order") }}</p><StatusBadge class="mt-3" :label='$t("Draft")' tone="slate" /></div></div></div>
            <div class="rounded-box border border-base-300 bg-base-100 p-4"><h3 class="text-xs font-semibold uppercase tracking-wide text-base-content/60">{{ $t("Order summary") }}</h3><dl class="mt-3 divide-y divide-base-300/70"><div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">{{ $t("Customer") }}</dt><dd class="max-w-[10rem] truncate">{{ $ui(selectedCustomer?.name || 'Walk-in customer') }}</dd></div><div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">{{ $t("Items") }}</dt><dd>{{ draftItems.length }}</dd></div><div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">{{ $t("Priority") }}</dt><dd>{{ $ui(priority) }}</dd></div><div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">{{ $t("Total") }}</dt><dd class="font-semibold tabular-nums text-primary">{{ $ui(previewTotal !== null ? formatMoney(previewTotal, currencyUnit) : 'Calculated on save') }}</dd></div></dl></div>
            <div v-if="draftItems.length" class="rounded-box border border-base-300 bg-base-100 p-4"><div class="flex items-center gap-2"><FileText :size="17" class="text-primary" aria-hidden="true" /><h3 class="text-sm font-semibold">{{ $t("Configured services") }}</h3></div><div class="mt-3 space-y-2"><div v-for="item in draftItems" :key="item.id" class="flex min-w-0 items-center gap-2 text-xs"><Package :size="14" class="shrink-0 text-base-content/50" aria-hidden="true" /><span class="min-w-0 flex-1 truncate">{{ serviceName(item) }}</span><span class="shrink-0 text-base-content/55">{{ $ui(item.payload.quantity || '1') }} {{ item.payload.quantityUnit }}</span></div></div></div>
            <div class="flex items-start gap-2 border-t border-base-300 pt-3 text-sm"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">{{ $t("The order starts as a draft. Confirm it later when the customer and pricing are ready.") }}</p></div>
          </div>
        </aside>
      </div>
    </div>

    <footer class="flex min-w-0 items-center justify-between gap-3 border-t border-base-300 px-1 pt-3"><button class="btn btn-ghost btn-sm gap-2" type="button" :disabled="activeStep === 1 || busy || saving || itemEditorOpen" @click="previous"><ArrowLeft class="rtl-directional-arrow" :size="14" aria-hidden="true" />{{ $t("Back") }}</button><span class="text-xs text-base-content/55">{{ $t("Step") }} {{ activeStep }} {{ $t("of") }} {{ steps.length }}</span><button v-if="activeStep < steps.length" class="btn btn-primary btn-sm" type="button" :disabled="itemEditorOpen" @click="next">{{ $t("Continue") }}</button><button v-else class="btn btn-success btn-sm gap-2" type="button" :disabled="busy || saving || itemEditorOpen" @click="save"><Save :size="14" aria-hidden="true" />{{ $ui(saving ? 'Saving…' : 'Create order') }}</button></footer>
  </div>
</template>
