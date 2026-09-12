<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { CalendarClock, CheckCheck, ClipboardList, Factory, FileText, Package, Save, UserRound } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatQuantityInput } from '../../utils/quantity'
import type { OrderRecord } from '../../api/orders'
import type { useProductionWorkspace } from './useProductionWorkspace'

const props = defineProps<{
  workspace: ReturnType<typeof useProductionWorkspace>
  currencyUnit: CurrencyUnit
  orders: OrderRecord[]
  machines: any[]
}>()
const emit = defineEmits<{ cancel: [] }>()

const {
  busy,
  createMode,
  editing,
  saving,
  form,
  selected,
  selectedOrder,
  selectedItem,
  chooseOrder,
  chooseItem,
  create,
  update,
} = props.workspace

const activeStep = ref(1)
const validationAttempted = ref(false)

const isCreating = computed(() => createMode.value)
const title = computed(() => (isCreating.value ? 'Add production job' : 'Edit production job'))
const confirmedOrders = computed(() => props.orders.filter((order) => order.commercialStatus === 'Confirmed' && order.items.length))
const itemOptions = computed(() => [
  { label: 'Select an order item', value: '' },
  ...(selectedOrder.value?.items ?? []).map((item) => ({
    label: `${item.serviceName} · ${item.quantity} ${item.quantityUnit}`,
    value: item.id,
  })),
])
const orderOptions = computed(() => [
  { label: confirmedOrders.value.length ? 'Select a confirmed order' : 'No confirmed orders available', value: '' },
  ...confirmedOrders.value.map((order) => ({
    label: `${order.orderNumber} · ${order.customerName || 'Walk-in customer'}`,
    value: order.id,
  })),
])
const steps = computed(() => isCreating.value
  ? [
      { number: 1, title: 'Order', description: 'Confirmed order and customer' },
      { number: 2, title: 'Item & quantity', description: 'Service item and output' },
      { number: 3, title: 'Schedule', description: 'Machine, priority, date' },
      { number: 4, title: 'Review', description: 'Confirm before creating' },
    ]
  : [
      { number: 1, title: 'Job context', description: 'Protected order snapshot' },
      { number: 2, title: 'Schedule', description: 'Machine, priority, date' },
      { number: 3, title: 'Review', description: 'Confirm before saving' },
    ])
const currentStep = computed(() => steps.value.find((step) => step.number === activeStep.value) ?? steps.value[0])
const customerLabel = computed(() => selectedOrder.value?.customerName || 'Walk-in customer')
const orderLabel = computed(() => {
  const order = props.orders.find((item) => item.id === (isCreating.value ? form.value.orderId : selected.value?.orderId))
  return order?.orderNumber || 'No order selected'
})
const serviceLabel = computed(() => selectedItem.value?.serviceName || selected.value?.serviceName || 'No service selected')
const machineLabel = computed(() => props.machines.find((machine) => machine.id === form.value.assignedMachineId)?.name || 'Unassigned')
const quantityLabel = computed(() => `${form.value.quantity || '0'} ${form.value.quantityUnit || 'unit'}`)
const plannedLabel = computed(() => form.value.plannedAt ? form.value.plannedAt : 'No planned date')
const draftError = computed(() => {
  if (!isCreating.value) return ''
  if (!form.value.orderId || !confirmedOrders.value.some((order) => order.id === form.value.orderId)) return 'Choose a confirmed order.'
  if (!form.value.orderItemId || !selectedOrder.value?.items.some((item) => item.id === form.value.orderItemId)) return 'Choose an order item.'
  const quantity = Number(form.value.quantity.replace(/,/g, ''))
  if (!Number.isFinite(quantity) || quantity <= 0) return 'Enter a quantity greater than zero.'
  if (!form.value.quantityUnit.trim()) return 'Enter a quantity unit.'
  return ''
})

function stepClass(number: number) {
  if (number === activeStep.value) return 'wizard-step-active'
  if (number < activeStep.value) return 'wizard-step-complete'
  return 'wizard-step-idle'
}

function focusInvalid() {
  void nextTick(() => document.querySelector<HTMLElement>('#production-editor-wizard :invalid')?.focus())
}

function validateCurrentStep() {
  validationAttempted.value = true
  const formElement = document.querySelector<HTMLFormElement>('#production-editor-wizard')
  if (formElement && !formElement.reportValidity()) {
    focusInvalid()
    return false
  }
  if (isCreating.value && activeStep.value <= 2 && draftError.value) return false
  return true
}

function next() {
  if (busy.value || saving.value) return
  if (!validateCurrentStep()) return
  if (activeStep.value < steps.value.length) activeStep.value += 1
}

function previous() {
  if (activeStep.value > 1) activeStep.value -= 1
}

function jumpToStep(target: number) {
  if (target <= activeStep.value) {
    activeStep.value = target
    return
  }
  while (activeStep.value < target) {
    if (!validateCurrentStep()) return
    activeStep.value += 1
  }
}

function submit() {
  if (busy.value || saving.value) return
  validationAttempted.value = true
  if (activeStep.value < steps.value.length) {
    next()
    return
  }
  if (!validateCurrentStep() || (isCreating.value && draftError.value)) {
    if (isCreating.value && draftError.value) activeStep.value = !form.value.orderId || !confirmedOrders.value.some((order) => order.id === form.value.orderId) ? 1 : 2
    focusInvalid()
    return
  }
  void (isCreating.value ? create() : update())
}
</script>

<template>
  <div class="service-wizard w-full flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Production job editor">
    <header class="service-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0"><h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">{{ title }}</h1><p class="mt-1 text-xs leading-4 text-base-content/65">Create a production job from a confirmed order and schedule its work.</p><WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Production' }, { label: title, current: true }]" @navigate="emit('cancel')" /></div>
      <div class="flex shrink-0 items-center gap-2"><button class="btn btn-error" type="button" :disabled="busy || saving" @click="emit('cancel')">Cancel</button><button class="btn btn-success gap-2" type="submit" form="production-editor-wizard" :disabled="busy || saving"><Save :size="16" aria-hidden="true" />{{ saving ? (isCreating ? 'Creating…' : 'Saving…') : (isCreating ? 'Create job' : 'Save changes') }}</button></div>
    </header>

    <div class="service-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="service-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200"><nav aria-label="Production job setup steps" class="service-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible"><button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" @click="jumpToStep(step.number)"><span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span><span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ step.title }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ step.description }}</small></span></button></nav></aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="service-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
          <form id="production-editor-wizard" class="service-editor min-w-0" @submit.prevent="submit">
            <section v-if="activeStep === 1" class="min-w-0 space-y-6"><div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><ClipboardList :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">{{ isCreating ? 'Choose the order' : 'Job context' }}</h2><p class="text-sm text-base-content/60">{{ isCreating ? 'Production jobs can only be created from confirmed orders.' : 'The order, service, and item snapshot are protected after creation.' }}</p></div></div><template v-if="isCreating"><SelectField v-model="form.orderId" label="Confirmed order" aria-label="Confirmed order" :options="orderOptions" :invalid="validationAttempted && (!form.orderId || !confirmedOrders.some((order) => order.id === form.orderId))" @update:model-value="chooseOrder" /><p v-if="validationAttempted && !form.orderId" class="text-xs text-error">Choose a confirmed order before continuing.</p><EmptyState v-if="!confirmedOrders.length" compact title="No confirmed orders available" description="Confirm an order with at least one item before creating production work."><template #icon><ClipboardList :size="21" aria-hidden="true" /></template></EmptyState></template><div v-else class="grid min-w-0 gap-3 rounded-box border border-base-300 bg-base-100/45 p-4 sm:grid-cols-2"><div class="min-w-0"><span class="block text-xs text-base-content/55">Customer</span><strong class="mt-1 block truncate text-sm">{{ customerLabel }}</strong></div><div class="min-w-0"><span class="block text-xs text-base-content/55">Order</span><strong class="mt-1 block truncate text-sm">{{ props.workspace.jobOrder(selected!) }}</strong></div><div class="min-w-0 sm:col-span-2"><span class="block text-xs text-base-content/55">Service item</span><strong class="mt-1 block truncate text-sm">{{ serviceLabel }}</strong></div></div></section>

            <section v-else-if="activeStep === 2 && isCreating" class="min-w-0 space-y-6"><div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Package :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Item and quantity</h2><p class="text-sm text-base-content/60">Choose the service item and confirm the output quantity for this job.</p></div></div><SelectField v-model="form.orderItemId" label="Order item" aria-label="Order item" :options="itemOptions" :invalid="validationAttempted && !form.orderItemId" @update:model-value="chooseItem" /><div class="grid min-w-0 gap-4 sm:grid-cols-2"><FormField label="Quantity" required class="gap-1"><AppInput class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && (!form.quantity || draftError.includes('quantity')) }" :model-value="formatQuantityInput(form.quantity)" required inputmode="decimal" placeholder="120" @update:model-value="form.quantity = formatQuantityInput($event)" /></FormField><FormField label="Unit" required class="gap-1"><AppInput v-model="form.quantityUnit" class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && (!form.quantityUnit.trim() || draftError.includes('unit')) }" required placeholder="piece" /></FormField></div><p v-if="validationAttempted && draftError" class="text-xs text-error">{{ draftError }}</p></section>

            <section v-else-if="activeStep === 2 || (activeStep === 3 && isCreating)" class="min-w-0 space-y-6"><div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><CalendarClock :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Schedule the work</h2><p class="text-sm text-base-content/60">Assign the job and add the planning context operators need.</p></div></div><div class="grid min-w-0 gap-4 sm:grid-cols-2"><SelectField v-model="form.assignedMachineId" label="Machine" :options="[{ label: 'Unassigned', value: '' }, ...props.machines.map((machine) => ({ label: machine.name, value: machine.id }))]" /><SelectField v-model="form.priority" label="Priority" :options="['Urgent', 'High', 'Normal', 'Low'].map((value) => ({ label: value, value }))" /><FormField label="Planned date" class="gap-1 sm:col-span-2"><JalaliDatePicker v-model="form.plannedAt" placeholder="Set planned date" /></FormField><FormField label="Notes" class="gap-1 sm:col-span-2"><AppTextarea v-model="form.notes" class="textarea w-full min-w-0" rows="6" placeholder="Add production notes, handoff details, or quality instructions" /></FormField></div></section>

            <section v-else class="min-w-0 space-y-6"><div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><FileText :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Review and save</h2><p class="text-sm text-base-content/60">Confirm the job context and schedule before {{ isCreating ? 'creating' : 'saving' }}.</p></div></div><div class="rounded-box border border-base-300 bg-base-100/45 p-4"><dl class="divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Customer</dt><dd class="max-w-[65%] truncate text-end">{{ customerLabel }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Order</dt><dd class="max-w-[65%] truncate text-end">{{ orderLabel }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Service item</dt><dd class="max-w-[65%] truncate text-end">{{ serviceLabel }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Quantity</dt><dd>{{ quantityLabel }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Machine</dt><dd class="max-w-[65%] truncate text-end">{{ machineLabel }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Priority</dt><dd>{{ form.priority }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Planned date</dt><dd>{{ plannedLabel }}</dd></div></dl></div><div v-if="isCreating && draftError" class="rounded-box border border-error/40 bg-error/10 p-3 text-sm text-error">{{ draftError }} Return to the earlier step to fix it.</div></section>
          </form>
        </section>

        <aside class="service-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto"><div class="space-y-4"><div class="relative min-h-48 overflow-hidden rounded-box bg-base-100 p-5"><div class="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-transparent" aria-hidden="true"></div><div class="relative flex min-h-36 flex-col justify-end"><span class="mb-auto grid size-11 place-items-center rounded-box border border-base-300 bg-base-200/80 text-primary"><Factory :size="24" aria-hidden="true" /></span><div class="mt-6 flex items-end justify-between gap-3"><div class="min-w-0"><h2 class="truncate text-lg font-semibold">{{ selected?.jobNumber || 'New production job' }}</h2><p class="mt-1 truncate text-xs text-base-content/60">{{ serviceLabel }}</p></div><StatusBadge :label="isCreating ? 'Pending' : (selected?.status || 'Pending')" :tone="isCreating ? 'slate' : props.workspace.statusTone(selected?.status || 'Pending')" /></div></div></div><div class="divide-y divide-base-300 rounded-box border border-base-300 bg-base-100/35"><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Customer</span><strong class="max-w-[10rem] truncate text-end">{{ customerLabel }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Quantity</span><strong>{{ quantityLabel }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Machine</span><strong class="max-w-[10rem] truncate text-end">{{ machineLabel }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Priority</span><strong>{{ form.priority }}</strong></div></div><div class="flex gap-2 border-t border-base-300 pt-3 text-sm"><UserRound class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">{{ currentStep?.description }}. The job stays linked to the confirmed order item after saving.</p></div></div></aside>
      </div>
    </div>
    <footer class="flex min-w-0 items-center justify-between gap-3 border-t border-base-300 px-1 pt-3"><button class="btn btn-ghost btn-sm" type="button" :disabled="activeStep === 1 || busy || saving" @click="previous">Back</button><span class="text-xs text-base-content/55">Step {{ activeStep }} of {{ steps.length }}</span><button v-if="activeStep < steps.length" class="btn btn-primary btn-sm" type="button" @click="next">Continue</button><button v-else class="btn btn-success btn-sm gap-2" type="submit" form="production-editor-wizard" :disabled="busy || saving"><Save :size="14" aria-hidden="true" />{{ isCreating ? 'Create job' : 'Save changes' }}</button></footer>
  </div>
</template>
