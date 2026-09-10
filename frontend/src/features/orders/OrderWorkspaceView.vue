<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, ref, watch } from 'vue';
import { ArrowDown, ArrowLeft, ArrowUp, CheckCircle2, PackageOpen, Pencil, Play, Plus, Trash2, XCircle } from 'lucide-vue-next';
import EmptyState from '../../components/ui/EmptyState.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import FormSection from '../../components/ui/FormSection.vue';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import SelectField from '../../components/ui/SelectField.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import WorkspaceTabs from '../../components/layout/WorkspaceTabs.vue';
import type { OrderItemPayload, OrderPayload, OrderRecord } from '../../api/orders';
import { ordersApi } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import { confirmAction } from '../../ui/feedback';
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
const saving = ref(false);
const isNew = computed(() => props.isNew || props.order.id.startsWith('new-order-'));

const tabs = ['Overview', 'Items', 'Production', 'Payments', 'Invoices', 'Files', 'History'];
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
const orderStatus = computed(() => {
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

function payload(): OrderPayload {
  return {
    customerId: customerId.value,
    promisedAt: promisedAt.value,
    priority: priority.value,
    notes: notes.value,
    discountRial: parseMoneyInput(discountText.value, props.currencyUnit) || 0,
  };
}

async function saveMetadata(returnToOrders = false) {
return runAction(async () => {
  saving.value = true;
  try {
    const result = isNew.value
      ? await ordersApi.create(payload())
      : await ordersApi.update(props.order.id, payload());
    emit('saved', result);
    emit('notify', 'Order details saved');
    if (returnToOrders) emit('back');
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
    const amount = parseMoneyInput(discountText.value, props.currencyUnit);
    if (amount === null) throw new Error('Enter a valid discount');
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

async function removeOrder() {
  return runAction(async () => {
    if (isNew.value) {
      emit('back');
      return;
    }
    if (!(await confirmAction({
      title: 'Delete order',
      message: 'Delete this order permanently? Orders with financial, production, or document history cannot be deleted.',
      confirmLabel: 'Delete order',
      danger: true,
    }))) return;
    try {
      await ordersApi.remove(props.order.id);
      emit('removed', props.order.id);
      emit('notify', 'Order deleted.');
    } catch (error) {
      reportError(error);
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
    editorOpen.value = false;
    editingItem.value = null;
    emit('notify', replacing ? 'Item replaced' : 'Item added');
  } catch (error) {
reportError(error);
  } finally {
    saving.value = false;
  }

});
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

async function move(item: any, delta: number) {
return runAction(async () => {
  const items = [...props.order.items].sort((a, b) => a.position - b.position);
  const index = items.findIndex((value) => value.id === item.id);
  const next = index + delta;
  if (next < 0 || next >= items.length) return;
  [items[index], items[next]] = [items[next], items[index]];
  try {
    emit(
      'saved',
      await ordersApi.reorderItems(
        props.order.id,
        items.map((value) => value.id),
      ),
    );
    emit('notify', 'Item order updated');
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

function snapshot(item: any, key: string) {
  try {
    return JSON.parse(item[key] || '[]');
  } catch {
    return [];
  }
}
</script>

<template>
  <div class="space-y-4">
    <WorkspaceStickyStack :flush="true">
      <WorkspaceHeader
        eyebrow="Sales / order workspace"
        :title="order.orderNumber"
        :description="`${order.customerName || 'Walk-in customer'} · ${order.items.length} line items`"
      >
        <template #title-suffix>
          <StatusBadge :label="orderStatus" :tone="tone(orderStatus)" />
        </template>
        <button class="btn btn-ghost btn-sm gap-2" type="button" @click="emit('back')">
          <ArrowLeft :size="16" aria-hidden="true" /><span>Orders</span>
        </button>
        <button
          v-if="!isNew"
          class="btn btn-outline btn-error btn-sm gap-2"
          type="button"
          :disabled="busy || saving"
          @click="removeOrder"
        >
          <Trash2 :size="15" aria-hidden="true" />
          Delete order
        </button>
        <button
          class="btn btn-primary btn-sm gap-2"
          type="button"
          :disabled="busy || saving"
          @click="saveMetadata(true)"
        >
          {{ saving ? 'Saving…' : 'Save order' }}
        </button>
      </WorkspaceHeader>

      <div
        class="grid gap-2 border-y border-base-300 bg-base-100 p-3 sm:grid-cols-2 xl:grid-cols-4"
      >
        <div class="min-w-0">
          <span class="block text-xs text-base-content/50">Customer</span>
          <strong class="block truncate text-sm">{{
            order.customerName || 'Walk-in customer'
          }}</strong>
          <span class="block truncate text-xs text-base-content/60">{{
            customer?.phone || order.customerPhone || 'No contact details'
          }}</span>
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

      <WorkspaceTabs
        class="mt-2"
        :tabs="isNew ? ['Overview'] : tabs"
        :active-tab="tab"
        @change="tab = $event"
      />
    </WorkspaceStickyStack>

    <AppPanel
      v-if="tab === 'Overview' && !isNew"
      title="Order workflow"
      subtitle="Move the order through its commercial and fulfillment stages."
    >
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
    </AppPanel>

    <section
      v-if="tab === 'Overview'"
      class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(18rem,26rem)]"
    >
      <AppPanel title="Order details" subtitle="Customer, delivery context, and notes.">
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
          <div class="flex justify-end">
            <button
              class="btn btn-primary gap-2"
              type="button"
              :disabled="busy || (saving)"
              @click="() => saveMetadata()"
            >
              {{ saving ? 'Saving…' : 'Save details' }}
            </button>
          </div>
        </div>
      </AppPanel>

      <AppPanel
        title="Order summary"
        subtitle="Derived from the saved order and pricing snapshots."
      >
        <div class="grid grid-cols-2 gap-3">
          <div class="rounded-box bg-base-200 p-3">
            <span class="block text-xs text-base-content/60">Subtotal</span
            ><strong class="mt-1 block text-sm">{{ money(order.subtotalRial) }}</strong>
          </div>
          <div class="rounded-box bg-base-200 p-3">
            <span class="block text-xs text-base-content/60">Discount</span
            ><strong class="mt-1 block text-sm">{{ money(order.discountRial) }}</strong>
          </div>
          <div class="rounded-box bg-base-200 p-3">
            <span class="block text-xs text-base-content/60">Total</span
            ><strong class="mt-1 block text-primary text-lg">{{ money(order.totalRial) }}</strong>
          </div>
          <div class="rounded-box bg-base-200 p-3">
            <span class="block text-xs text-base-content/60">Estimated cost</span
            ><strong class="mt-1 block text-sm">{{ money(order.estimatedCostRial) }}</strong>
          </div>
        </div>
        <div class="mt-4 border-t border-base-300 pt-4">
          <FormField class="gap-1">
            <span class="text-xs text-base-content/60">Order discount</span>
            <AppInput
              v-model="discountText"
              :money="props.currencyUnit"
              inputmode="decimal"
              :placeholder="`Amount in ${props.currencyUnit}`"
              @blur="
                discountText = formatMoneyInput(
                  parseMoneyInput(discountText, props.currencyUnit) || 0,
                  props.currencyUnit,
                )
              "
            />
          </FormField>
          <button
            class="btn btn-outline mt-3 w-full"
            type="button"
            :disabled="isNew || busy || (saving)"
            @click="updateDiscount"
          >
            Apply discount
          </button>
        </div>
      </AppPanel>
    </section>

    <template v-else-if="tab === 'Items'">
      <AppPanel
        v-if="editorOpen"
        title="Configure service item"
        subtitle="Pricing is calculated from the persisted service definition."
      >
        <OrderItemConfigurator
          :services="services"
          :materials="materials"
          :machines="machines"
          :currency-unit="currencyUnit"
          :initial="editingItem" :busy="busy"
          @configured="configured"
          @cancel="
            editorOpen = false;
            editingItem = null;
          "
        />
      </AppPanel>
      <RegisterList
        v-else
        title="Configured items"
        subtitle="Accepted service and pricing snapshots for this order."
        :count="order.items.length"
      >
        <template #action
          ><button class="btn btn-primary btn-sm gap-2" type="button" @click="editorOpen = true">
            <Plus :size="15" aria-hidden="true" />Add service item
          </button></template
        >
        <EmptyState
          v-if="!order.items.length"
          title="No configured items"
          description="Add a service item to calculate pricing and prepare production."
        >
          <template #icon><PackageOpen :size="22" aria-hidden="true" /></template>
          <template #action
            ><button
              class="btn btn-primary btn-sm mt-3 gap-2"
              type="button"
              @click="editorOpen = true"
            >
              <Plus :size="15" aria-hidden="true" />Add service item
            </button></template
          >
        </EmptyState>
        <RegisterRow :interactive="false" :sidecar="true"
          v-for="item in [...order.items].sort((a, b) => a.position - b.position)"
          v-else
          :key="item.id"
        >
          <template #icon
            ><span class="text-xs font-semibold text-base-content/60">{{
              item.position + 1
            }}</span></template
          >
          <template #identity
            ><strong class="block truncate text-sm">{{ item.serviceName }}</strong></template
          >
          <template #meta
            ><span class="block text-sm text-base-content/70"
              >{{ item.quantity }} {{ item.quantityUnit
              }}<template v-if="item.notes"> · {{ item.notes }}</template></span
            ><small class="block text-xs text-base-content/50"
              >Accepted snapshot · {{ snapshot(item, 'resolvedParametersJson').length }} parameters
              · {{ snapshot(item, 'costBreakdownJson').length }} cost lines</small
            ></template
          >
          <template #status
            ><div class="text-end text-xs">
              <span class="block text-base-content/50"
                >Cost {{ money(item.estimatedCostRial) }}</span
              ><strong class="block text-sm text-primary">{{
                money(item.sellingPriceRial)
              }}</strong>
            </div></template
          >
          <template #actions>
            <div class="flex items-center gap-1" @click.stop>
              <button
                class="btn btn-ghost btn-square btn-sm"
                type="button"
                aria-label="Move item up"
                title="Move item up"
                @click="move(item, -1)"
               :disabled="busy">
                <ArrowUp :size="14" aria-hidden="true" />
              </button>
              <button
                class="btn btn-ghost btn-square btn-sm"
                type="button"
                aria-label="Move item down"
                title="Move item down"
                @click="move(item, 1)"
               :disabled="busy">
                <ArrowDown :size="14" aria-hidden="true" />
              </button>
              <button
                class="btn btn-ghost btn-square btn-sm"
                type="button"
                aria-label="Reconfigure item"
                title="Reconfigure item"
                @click="
                  editingItem = item;
                  editorOpen = true;
                "
              >
                <Pencil :size="14" aria-hidden="true" />
              </button>
              <button
                class="btn btn-outline btn-error btn-square btn-sm"
                type="button"
                aria-label="Remove item"
                title="Remove item"
                @click="remove(item)"
               :disabled="busy">
                <Trash2 :size="14" aria-hidden="true" />
              </button>
            </div>
          </template>
        </RegisterRow>
      </RegisterList>
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
    />
    <OrderPaymentsPanel
      v-else-if="tab === 'Payments'"
      :order="order"
      :currency-unit="currencyUnit"
      @notify="emit('notify', $event)"
      @saved="emit('saved', $event)"
    />
    <AppPanel v-else title="Order history" subtitle="A chronological record of important order changes.">
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
    </AppPanel>
  </div>
</template>
