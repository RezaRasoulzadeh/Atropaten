<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import { computed, ref, watch } from 'vue';
import { CheckCheck, CheckCircle2, Layers3, PackageOpen, Pencil, Play, Plus, Save, Trash2, XCircle } from 'lucide-vue-next';
import EmptyState from '../../components/ui/EmptyState.vue';
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue';
import SearchField from '../../components/ui/SearchField.vue';
import FormSection from '../../components/ui/FormSection.vue';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import SelectField from '../../components/ui/SelectField.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import type { OrderItemPayload, OrderPayload, OrderRecord } from '../../api/orders';
import { ordersApi } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, formatMoneyInput, formatMoneyInputWhileTyping, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import DocumentMetadataPanel from '../documents/DocumentMetadataPanel.vue';
import OrderItemConfigurator from '../sales/OrderItemConfigurator.vue';
import OrderInvoicePanel from './OrderInvoicePanel.vue';
import OrderPaymentsPanel from './OrderPaymentsPanel.vue';
import OrderProductionPanel from './OrderProductionPanel.vue';

const props = defineProps<{
  order: OrderRecord;
  currencyUnit: CurrencyUnit;
  customers: any[];
  services: any[];
  materials: any[];
  machines: any[];
  isNew?: boolean;
}>();

const emit = defineEmits<{
  back: [];
  removed: [id: string];
  notify: [message: string];
  saved: [order: OrderRecord];
}>();

type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate';

const tab = ref('Overview');
const customerId = ref('');
const promisedAt = ref<string | null>(null);
const priority = ref('Normal');
const notes = ref('');
const discountText = ref('');
const editorOpen = ref(false);
const editingItem = ref<any | null>(null);
const serviceQuery = ref('');
const selectedServiceForEditor = ref('');
const saving = ref(false);
const isNew = computed(() => props.isNew || props.order.id.startsWith('new-order-'));

const steps = [
  { number: 1, tab: 'Overview', title: 'Details', description: 'Customer, delivery, notes' },
  { number: 2, tab: 'Items', title: 'Items', description: 'Services and quantities' },
  { number: 3, tab: 'Files', title: 'Files', description: 'Artwork and references' },
  { number: 4, tab: 'Payments', title: 'Payments', description: 'Receipts and balance' },
  { number: 5, tab: 'Invoices', title: 'Invoices', description: 'Commercial documents' },
  { number: 6, tab: 'Production', title: 'Production', description: 'Jobs and progress' },
  { number: 7, tab: 'History', title: 'Workflow & history', description: 'Status changes and activity' },
];
const priorityOptions = ['Urgent', 'High', 'Normal', 'Low'].map((value) => ({
  label: value,
  value,
}));
const customer = computed(() => props.customers.find((value) => value.id === customerId.value));
const customerOptions = computed(() => [
  { label: 'Walk-in customer', value: '' },
  ...props.customers
    .filter((value) => value.active || value.id === customerId.value)
    .map((value) => ({ label: value.name, value: value.id })),
]);
const filteredCatalogServices = computed(() => {
  const query = serviceQuery.value.trim().toLowerCase();
  return props.services.filter((service) => {
    if (!service.active && service.id !== selectedServiceForEditor.value) return false;
    if (!query) return true;
    return [service.name, service.code, service.category].some((value) => String(value || '').toLowerCase().includes(query));
  });
});
const orderStatus = computed(() => {
  if (props.order.archived) return 'Archived';
  if (props.order.commercialStatus === 'Cancelled') return 'Cancelled';
  if (props.order.commercialStatus === 'Closed') return 'Closed';
  if (props.order.fulfillmentStatus === 'Delivered') return 'Delivery';
  if (props.order.fulfillmentStatus === 'In Production' || props.order.fulfillmentStatus === 'Ready') return 'Production';
  return props.order.commercialStatus;
});
const orderStatusActions = computed(() => {
  const labels: Record<string, string> = {
    Draft: 'Set draft',
    Confirmed: 'Confirm order',
    Production: 'Start production',
    Delivery: 'Mark delivered',
    Closed: 'Close order',
    Cancelled: 'Cancel order',
  };
  return ['Draft', 'Confirmed', 'Production', 'Delivery', 'Closed', 'Cancelled']
    .filter((status) => status !== orderStatus.value)
    .map((status) => ({
      status,
      label: labels[status],
      kind: status === 'Cancelled' ? 'danger' : status === 'Draft' ? 'secondary' : 'primary',
    }));
});

watch(
  () => props.order,
  (order) => {
    customerId.value = order.customerId || '';
    promisedAt.value = order.promisedAt;
    priority.value = order.priority;
    notes.value = order.notes;
    discountText.value = formatMoneyInput(order.discountRial, props.currencyUnit);
  },
  { immediate: true },
);

function tone(value: string): Tone {
  return value === 'Confirmed' || value === 'Production'
    ? 'blue'
    : value === 'Closed' || value === 'Delivery' || value === 'Delivered' || value === 'Paid' || value === 'Ready'
      ? 'green'
      : value === 'Cancelled'
        ? 'red'
        : value === 'Partially Paid'
          ? 'amber'
          : 'slate';
}

function money(value: number) {
  return formatMoney(value, props.currencyUnit);
}

function stepClass(number: number) {
  if (number === activeStepNumber.value) return 'wizard-step-active';
  if (number < activeStepNumber.value) return 'wizard-step-complete';
  return 'wizard-step-idle';
}

const activeStepNumber = computed(() => steps.find((step) => step.tab === tab.value)?.number || 1);

function selectStep(step: (typeof steps)[number]) {
  if (isNew.value && step.number > 1) return;
  tab.value = step.tab;
}

function previousStep() {
  const previous = steps[activeStepNumber.value - 2];
  if (previous) tab.value = previous.tab;
}

async function nextStep() {
  if (isNew.value && activeStepNumber.value === 1) {
    const savedOrder = await saveMetadata();
    if (savedOrder) tab.value = 'Items';
    return;
  }
  const next = steps[activeStepNumber.value];
  if (next) tab.value = next.tab;
}

function updateDiscountInput(value: string) {
  const trimmed = value.trim();
  if (trimmed.endsWith('%')) {
    const percent = trimmed.slice(0, -1).replaceAll(',', '').trim();
    discountText.value = /^\d*(?:\.\d*)?$/.test(percent) ? `${percent}%` : value;
    return;
  }
  discountText.value = formatMoneyInputWhileTyping(value, props.currencyUnit);
}

function parseDiscountRial(value = discountText.value): number | null {
  const trimmed = value.trim();
  if (!trimmed) return 0;
  const percentage = /^(\d+(?:\.\d+)?)\s*%$/.exec(trimmed.replaceAll(',', ''));
  if (percentage) {
    const percent = Number(percentage[1]);
    if (!Number.isFinite(percent) || percent < 0 || percent > 100) return null;
    return Math.round((props.order.subtotalRial * percent) / 100);
  }
  return parseMoneyInput(trimmed, props.currencyUnit);
}

function normalizeDiscountInput() {
  const trimmed = discountText.value.trim();
  if (!trimmed) {
    discountText.value = '';
    return;
  }
  const percentage = /^(\d+(?:\.\d+)?)\s*%$/.exec(trimmed.replaceAll(',', ''));
  if (percentage) {
    const percent = Number(percentage[1]);
    if (Number.isFinite(percent) && percent <= 100) discountText.value = `${percent}%`;
    return;
  }
  const amount = parseMoneyInput(trimmed, props.currencyUnit);
  if (amount !== null) discountText.value = formatMoneyInput(amount, props.currencyUnit);
}

function payload(): OrderPayload {
  return {
    customerId: customerId.value,
    promisedAt: promisedAt.value,
    priority: priority.value,
    notes: notes.value,
    discountRial: parseDiscountRial() || 0,
  };
}

async function saveMetadata(closeAfterSave = false) {
  if (saving.value || busy.value) return;
return runAction(async () => {
  saving.value = true;
  try {
    const discount = parseDiscountRial();
    if (discount === null || discount > props.order.subtotalRial) throw new Error('Enter a valid discount up to the order subtotal');
    const creating = isNew.value;
    const result = creating
      ? await ordersApi.create(payload())
      : await ordersApi.update(props.order.id, payload());
    emit('saved', result);
    emit('notify', creating ? 'Order created.' : 'Order details saved.');
    if (closeAfterSave) emit('back');
    return result;
  } catch (error) {
reportError(error);
  } finally {
    saving.value = false;
  }

});
}

async function updateDiscount() {
return runAction(async () => {
  saving.value = true;
  try {
    const amount = parseDiscountRial();
    if (amount === null || amount > props.order.subtotalRial) throw new Error('Enter a valid discount up to the order subtotal');
    const result = await ordersApi.discount(props.order.id, amount);
    emit('saved', result);
    emit('notify', 'Discount applied');
  } catch (error) {
reportError(error);
  } finally {
    saving.value = false;
  }

});
}

async function configured(input: OrderItemPayload) {
return runAction(async () => {
  saving.value = true;
  const replacing = Boolean(editingItem.value);
  try {
    const result = replacing
      ? await ordersApi.replaceItem(props.order.id, editingItem.value.id, input)
      : await ordersApi.addItem(props.order.id, input);
    emit('saved', result);
    closeItemEditor();
    emit('notify', replacing ? 'Item replaced' : 'Item added');
  } catch (error) {
reportError(error);
  } finally {
    saving.value = false;
  }

});
}

function openItemEditor(serviceId = '') {
  editingItem.value = null;
  selectedServiceForEditor.value = serviceId;
  editorOpen.value = true;
}

function editItem(item: any) {
  editingItem.value = item;
  selectedServiceForEditor.value = '';
  editorOpen.value = true;
}

function closeItemEditor() {
  editorOpen.value = false;
  editingItem.value = null;
  selectedServiceForEditor.value = '';
}

async function remove(item: any) {
return runAction(async () => {
  try {
    emit('saved', await ordersApi.removeItem(props.order.id, item.id));
    emit('notify', 'Item removed');
  } catch (error) {
reportError(error);
  }

});
}

async function changeOrderStatus(value: string) {
return runAction(async () => {
  try {
    let result: OrderRecord;
    if (value === 'Production' || value === 'Delivery') {
      if (props.order.commercialStatus !== 'Confirmed') {
        await ordersApi.commercialStatus(props.order.id, 'Confirmed');
      }
      if (value === 'Production' && props.order.fulfillmentStatus !== 'Pending' && props.order.fulfillmentStatus !== 'In Production') {
        await ordersApi.fulfillmentStatus(props.order.id, 'Pending');
      }
      result = await ordersApi.fulfillmentStatus(
        props.order.id,
        value === 'Production' ? 'In Production' : 'Delivered',
      );
    } else {
      result = await ordersApi.commercialStatus(props.order.id, value);
      if ((value === 'Draft' || value === 'Confirmed') && props.order.fulfillmentStatus !== 'Pending') {
        result = await ordersApi.fulfillmentStatus(props.order.id, 'Pending');
      }
    }
    emit('saved', result);
    emit('notify', 'Order status updated');
  } catch (error) {
reportError(error);
  }

});
}

</script>

<template>
  <div class="order-wizard w-full flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Order editor">
    <header class="order-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">{{ isNew ? 'Add order' : 'Edit order' }}</h1>
        <p class="mt-1 text-xs leading-4 text-base-content/65">Create and manage the customer order through its workflow.</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Orders' }, { label: isNew ? 'Add order' : 'Edit order', current: true }]" @navigate="emit('back')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="busy || saving" @click="emit('back')">Cancel</button>
        <button class="btn btn-success gap-2" type="button" :disabled="busy || saving" @click="saveMetadata(true)"><Save :size="16" aria-hidden="true" />{{ saving ? 'Saving…' : 'Save order' }}</button>
      </div>
    </header>

    <div class="order-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="order-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200">
        <nav aria-label="Order setup steps" class="order-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible">
          <button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" :disabled="isNew && step.number > 1" @click="selectStep(step)">
            <span class="wizard-step-number"><CheckCheck v-if="step.number < activeStepNumber" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span>
            <span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ step.title }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ step.description }}</small></span>
          </button>
        </nav>
      </aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="order-wizard-form-panel min-h-0 min-w-0 space-y-4 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:h-full xl:overflow-y-auto">
    <section
      v-if="tab === 'Overview'"
      class="min-w-0 space-y-6"
    >
        <div class="space-y-3">
          <FormSection
            title="Customer and delivery"
            description="Keep the commercial identity and promised date together."
          >
            <SelectField v-model="customerId" label="Customer" :options="customerOptions" />
            <SelectField
              v-model="priority"
              label="Priority"
              :options="priorityOptions"
              aria-label="Priority"
            />
            <div class="gap-1">
              <span class="text-xs">Promised date</span
              ><JalaliDatePicker v-model="promisedAt" placeholder="Set promised date" />
            </div>
          </FormSection>
          <FormSection title="Notes">
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs">Order notes</span
              ><AppTextarea
                v-model="notes"
                rows="4"
                placeholder="Order notes"
              />
            </FormField>
          </FormSection>
        </div>
    </section>

    <template v-else-if="tab === 'Items'">
      <div class="grid min-h-0 min-w-0 gap-4 xl:h-full xl:grid-cols-2 xl:divide-x xl:divide-base-300">
        <section class="flex min-h-80 min-w-0 flex-col overflow-hidden xl:h-full xl:pe-4" aria-label="Service catalog">
          <div class="sticky top-0 z-10 shrink-0 border-b border-base-300 bg-base-200/20 pb-3">
            <div class="flex items-center justify-between gap-2">
              <div><h2 class="text-sm font-semibold">Services</h2><p class="mt-1 text-xs text-base-content/60">Choose a service to add.</p></div>
              <span class="badge badge-ghost text-xs">{{ filteredCatalogServices.length }}</span>
            </div>
            <SearchField v-model="serviceQuery" class="mt-3" placeholder="Search services…" aria-label="Search services" />
          </div>
          <div class="min-h-0 flex-1 space-y-2 overflow-y-auto pt-3">
            <button v-for="service in filteredCatalogServices" :key="service.id" class="flex w-full min-w-0 items-center gap-2 rounded-box border border-base-300 bg-base-100 p-2.5 text-start transition-colors hover:border-primary/50 hover:bg-primary/5" type="button" @click="openItemEditor(service.id)">
              <span class="grid size-10 shrink-0 place-items-center overflow-hidden rounded-box border border-base-300 bg-base-200 bg-cover bg-center text-primary" :style="service.imagePath ? { backgroundImage: `url('${service.imagePath}')` } : undefined"><Layers3 v-if="!service.imagePath" :size="18" aria-hidden="true" /></span>
              <span class="min-w-0"><strong class="block truncate text-sm">{{ service.name }}</strong><span class="mt-0.5 block truncate text-xs text-base-content/55">{{ service.code || service.category || 'Service' }}</span><span class="mt-0.5 block text-xs text-primary">Add item</span></span>
            </button>
            <EmptyState v-if="!filteredCatalogServices.length" compact title="No services found" description="Try another search or add an active service first."><template #icon><Layers3 :size="21" aria-hidden="true" /></template></EmptyState>
          </div>
        </section>

        <section class="flex min-h-80 min-w-0 flex-col overflow-hidden xl:h-full xl:ps-4" aria-label="Order items">
          <div class="flex shrink-0 items-center justify-between gap-3 border-b border-base-300 pb-3"><div><h2 class="text-sm font-semibold">Order items</h2><p class="mt-1 text-xs text-base-content/60">Edit or remove configured services.</p></div><span class="badge badge-ghost text-xs">{{ order.items.length }} item{{ order.items.length === 1 ? '' : 's' }}</span></div>
          <div class="min-h-0 flex-1 space-y-2 overflow-y-auto pt-3">
            <div v-for="item in [...order.items].sort((a, b) => a.position - b.position)" :key="item.id" class="flex min-w-0 items-center gap-2 rounded-box border border-base-300 bg-base-200/20 p-2.5">
              <span class="grid size-8 shrink-0 place-items-center rounded-box bg-base-200 text-xs font-semibold text-base-content/60">{{ item.position + 1 }}</span>
              <span class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ item.serviceName }}</strong><span class="block truncate text-xs text-base-content/55">{{ item.quantity }} {{ item.quantityUnit }}<template v-if="item.notes"> · {{ item.notes }}</template></span></span>
              <span class="shrink-0 text-end"><strong class="block text-sm text-primary">{{ money(item.sellingPriceRial) }}</strong><span class="block text-[0.68rem] text-base-content/50">{{ money(item.estimatedCostRial) }} cost</span></span>
              <div class="flex shrink-0 items-center gap-0.5"><button class="btn btn-ghost btn-square btn-xs" type="button" aria-label="Edit order item" title="Edit item" :disabled="busy" @click="editItem(item)"><Pencil :size="14" aria-hidden="true" /></button><button class="btn btn-ghost btn-error btn-square btn-xs" type="button" aria-label="Remove order item" title="Remove item" :disabled="busy" @click="remove(item)"><Trash2 :size="14" aria-hidden="true" /></button></div>
            </div>
            <EmptyState v-if="!order.items.length" compact title="No items yet" description="Choose a service from the catalog to add the first item."><template #icon><PackageOpen :size="21" aria-hidden="true" /></template></EmptyState>
          </div>
        </section>
      </div>
    </template>

    <OrderInvoicePanel
      v-else-if="tab === 'Invoices'"
      :order="order"
      :currency-unit="currencyUnit"
      @notify="emit('notify', $event)"
      @saved="emit('saved', $event)"
    />
    <DocumentMetadataPanel
      v-else-if="tab === 'Files'"
      owner-type="order"
      :owner-id="order.id"
      :protected-context="order.commercialStatus !== 'Draft'"
      @notify="emit('notify', $event)"
    />
    <OrderProductionPanel
      v-else-if="tab === 'Production'"
      :order="order"
      :currency-unit="currencyUnit"
      @notify="emit('notify', $event)"
      @saved="emit('saved', $event)"
    />
    <template v-else-if="tab === 'Payments'">
      <section class="min-w-0 border-b border-base-300 pb-4">
        <header class="mb-3">
          <h2 class="text-sm font-semibold leading-5">Order discount</h2>
          <p class="mt-1 text-xs leading-4 text-base-content/60">Enter a rial amount or a percentage of the order subtotal.</p>
        </header>
        <div class="flex max-w-2xl items-end gap-2">
          <FormField class="min-w-0 flex-1 gap-1">
            <span class="text-xs text-base-content/60">Discount</span>
            <div class="flex min-w-0 items-center gap-2">
              <AppInput
                class="flex-1"
                :model-value="discountText"
                inputmode="decimal"
                placeholder="Amount or 10%"
                @update:model-value="updateDiscountInput"
                @blur="normalizeDiscountInput"
              />
            </div>
          </FormField>
          <button
            class="btn btn-outline shrink-0"
            type="button"
            :disabled="isNew || busy || saving"
            @click="updateDiscount"
          >
            Apply
          </button>
        </div>
      </section>
      <OrderPaymentsPanel
        :order="order"
        :currency-unit="currencyUnit"
        @notify="emit('notify', $event)"
        @saved="emit('saved', $event)"
      />
    </template>
    <template v-else-if="tab === 'History'">
      <section v-if="!isNew" class="min-w-0 border-b border-base-300 pb-4">
        <header class="mb-3">
          <h2 class="text-sm font-semibold leading-5">Order workflow</h2>
          <p class="mt-1 text-xs leading-4 text-base-content/60">Move the order through its commercial and fulfillment stages.</p>
        </header>
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2 text-sm">
            <span class="text-xs text-base-content/60">Current status</span>
            <StatusBadge :label="orderStatus" :tone="tone(orderStatus)" />
          </div>
          <div v-if="orderStatusActions.length" class="flex flex-wrap items-center gap-2" role="group" aria-label="Available order status actions">
            <span class="text-xs text-base-content/60">Move to</span>
            <button
              v-for="action in orderStatusActions"
              :key="action.status"
              class="btn btn-sm gap-2"
              :class="action.kind === 'primary' ? 'btn-primary' : action.kind === 'danger' ? 'btn-outline btn-error' : 'btn-outline'"
              type="button"
              :disabled="busy"
              @click="changeOrderStatus(action.status)"
            >
              <Play v-if="action.status === 'Production'" :size="14" aria-hidden="true" />
              <XCircle v-else-if="action.status === 'Cancelled'" :size="14" aria-hidden="true" />
              <CheckCircle2 v-else :size="14" aria-hidden="true" />
              {{ action.label }}
            </button>
          </div>
          <EmptyState v-else compact title="No further workflow actions" description="This order is at its current terminal or completed state."><template #icon><CheckCircle2 :size="21" aria-hidden="true" /></template></EmptyState>
        </div>
      </section>
      <section class="min-w-0">
        <header class="mb-3">
          <h2 class="text-sm font-semibold leading-5">Order history</h2>
          <p class="mt-1 text-xs leading-4 text-base-content/60">A chronological record of important order changes.</p>
        </header>
      <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,0.6fr)]">
        <div class="relative space-y-3 ps-6 before:absolute before:inset-y-2 before:start-2 before:w-px before:bg-base-300">
          <div class="relative rounded-box border border-base-300 bg-base-200/35 p-3">
            <span class="absolute -start-[1.58rem] top-4 grid size-4 place-items-center rounded-full border-2 border-base-100 bg-primary ring-1 ring-primary/30"></span>
            <span class="block text-xs text-base-content/50">Created</span>
            <strong class="mt-1 block text-sm">Order {{ order.orderNumber }} was created</strong>
            <span class="mt-1 block text-xs text-base-content/55">{{ formatDateTime(order.createdAt) }}</span>
          </div>
          <div class="relative rounded-box border border-base-300 bg-base-200/35 p-3">
            <span class="absolute -start-[1.58rem] top-4 grid size-4 place-items-center rounded-full border-2 border-base-100 bg-base-300 ring-1 ring-base-300"></span>
            <span class="block text-xs text-base-content/50">Last updated</span>
            <strong class="mt-1 block text-sm">Order details were last saved</strong>
            <span class="mt-1 block text-xs text-base-content/55">{{ formatDateTime(order.updatedAt) }}</span>
          </div>
        </div>
        <div class="rounded-box border border-dashed border-base-300 bg-base-200/25 p-4">
          <div class="flex items-start gap-3">
            <div class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary">
              <PackageOpen :size="18" aria-hidden="true" />
            </div>
            <div>
              <strong class="block text-sm">Detailed history is coming next</strong>
              <p class="mt-2 text-xs leading-5 text-base-content/60">
                Status changes, item updates, payments, production, and corrections will appear here as a preserved timeline.
              </p>
            </div>
          </div>
          <div class="mt-4 border-t border-base-300 pt-3">
            <span class="block text-xs text-base-content/50">Current order status</span>
            <StatusBadge class="mt-2" :label="orderStatus" :tone="tone(orderStatus)" />
          </div>
        </div>
      </div>
      </section>
    </template>

        </section>

        <aside class="order-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:h-full xl:overflow-y-auto">
          <div class="space-y-4">
            <section class="rounded-box border border-base-300 bg-base-100 p-4">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <span class="block text-xs text-base-content/55">Order overview</span>
                  <strong class="mt-1 block truncate text-lg">{{ isNew ? 'New order' : order.orderNumber }}</strong>
                </div>
                <StatusBadge :label="orderStatus" :tone="tone(orderStatus)" />
              </div>
              <div class="mt-4 border-t border-base-300 pt-3">
                <span class="block text-xs text-base-content/55">Customer</span>
                <strong class="mt-1 block truncate text-sm">{{ customer?.name || order.customerName || 'Walk-in customer' }}</strong>
                <span class="mt-1 block truncate text-xs text-base-content/60">{{ customer?.phone || order.customerPhone || 'No contact details' }}</span>
              </div>
            </section>

            <section class="rounded-box border border-base-300 bg-base-100 p-4">
              <h2 class="text-sm font-semibold">Order summary</h2>
              <p class="mt-1 text-xs leading-5 text-base-content/60">Current totals and fulfillment context.</p>
              <dl class="mt-4 divide-y divide-base-300/70">
                <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Items</dt><dd>{{ order.items.length }}</dd></div>
                <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Priority</dt><dd>{{ priority }}</dd></div>
                <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Promised</dt><dd>{{ promisedAt ? formatDateTime(promisedAt) : 'Not set' }}</dd></div>
                <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Subtotal</dt><dd>{{ money(order.subtotalRial) }}</dd></div>
                <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Discount</dt><dd>{{ money(order.discountRial) }}</dd></div>
                <div class="flex items-center justify-between gap-3 py-2.5 text-sm font-semibold"><dt>Total</dt><dd class="text-primary">{{ money(order.totalRial) }}</dd></div>
                <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Paid</dt><dd class="text-success">{{ money(order.paidRial || 0) }}</dd></div>
                <div class="flex items-center justify-between gap-3 py-2.5 text-sm font-semibold"><dt>Remaining</dt><dd class="text-warning">{{ money(order.remainingRial ?? order.totalRial) }}</dd></div>
              </dl>
            </section>

            <section v-if="order.items.length" class="rounded-box border border-base-300 bg-base-100 p-4">
              <h2 class="text-sm font-semibold">Configured services</h2>
              <ul class="mt-3 divide-y divide-base-300/70">
                <li v-for="item in [...order.items].sort((a, b) => a.position - b.position)" :key="item.id" class="flex items-start justify-between gap-3 py-2.5 first:pt-0 last:pb-0">
                  <span class="min-w-0"><strong class="block truncate text-sm">{{ item.serviceName }}</strong><small class="mt-0.5 block text-xs text-base-content/55">{{ item.quantity }} {{ item.quantityUnit }}</small></span>
                  <strong class="shrink-0 text-sm text-primary">{{ money(item.sellingPriceRial) }}</strong>
                </li>
              </ul>
            </section>

            <div class="border-t border-base-300 pt-3 text-xs leading-5 text-base-content/65">
              <span class="block">Created {{ formatDateTime(order.createdAt) }}</span>
              <span class="block">Updated {{ formatDateTime(order.updatedAt) }}</span>
            </div>
          </div>
        </aside>
      </div>
    </div>

    <footer class="flex min-w-0 items-center justify-between gap-3 border-t border-base-300 px-1 pt-3">
      <button class="btn btn-ghost btn-sm" type="button" :disabled="activeStepNumber === 1 || busy || saving" @click="previousStep">Back</button>
      <span class="text-xs text-base-content/55">Step {{ activeStepNumber }} of {{ steps.length }}</span>
      <button v-if="activeStepNumber < steps.length" class="btn btn-primary btn-sm" type="button" :disabled="busy || saving" @click="nextStep">Continue</button>
      <button v-else class="btn btn-success btn-sm gap-2" type="button" :disabled="busy || saving" @click="saveMetadata(true)"><Save :size="14" aria-hidden="true" />{{ saving ? 'Saving…' : 'Save order' }}</button>
    </footer>

    <div v-if="editorOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/65 p-3 sm:p-6" role="dialog" aria-modal="true" aria-label="Configure order service item" @click.self="closeItemEditor">
      <div class="w-full max-w-4xl overflow-hidden rounded-box border border-base-300 bg-base-100 shadow-2xl">
        <OrderItemConfigurator
          :services="services"
          :materials="materials"
          :machines="machines"
          :currency-unit="currencyUnit"
          :initial="editingItem"
          :preset-service-id="selectedServiceForEditor"
          :busy="busy || saving"
          @configured="configured"
          @cancel="closeItemEditor"
        />
      </div>
    </div>
  </div>
</template>
