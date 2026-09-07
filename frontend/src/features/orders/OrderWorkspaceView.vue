<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, ref, watch } from 'vue';
import { ArrowDown, ArrowLeft, ArrowUp, PackageOpen, Pencil, Plus, Trash2 } from 'lucide-vue-next';
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
}>();

const emit = defineEmits<{
  back: [];
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

const tabs = ['Overview', 'Items', 'Production', 'Payments', 'Invoices', 'Files', 'History'];
const commercialOptions = ['Draft', 'Confirmed', 'Closed', 'Cancelled'].map((value) => ({
  label: value,
  value,
}));
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
  return value === 'Confirmed' || value === 'In Production'
    ? 'blue'
    : value === 'Closed' || value === 'Delivered' || value === 'Paid' || value === 'Ready'
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

async function saveMetadata() {
return runAction(async () => {
  saving.value = true;
  try {
    const result = await ordersApi.update(props.order.id, payload());
    emit('saved', result);
    emit('notify', 'Order details saved');
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

async function changeCommercial(value: string) {
return runAction(async () => {
  try {
    emit('saved', await ordersApi.commercialStatus(props.order.id, value));
    emit('notify', 'Commercial status updated');
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
        <button class="btn btn-ghost btn-sm gap-2" type="button" @click="emit('back')">
          <ArrowLeft :size="16" aria-hidden="true" /><span>Orders</span>
        </button>
        <StatusBadge :label="order.commercialStatus" :tone="tone(order.commercialStatus)" />
        <SelectField
          class="w-40"
          :model-value="order.commercialStatus"
          :options="commercialOptions"
          aria-label="Commercial status"
          @update:model-value="changeCommercial"
        />
        <button
          class="btn btn-primary gap-2"
          type="button"
          :disabled="busy || (saving)"
          @click="saveMetadata"
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

      <WorkspaceTabs class="mt-2" :tabs="tabs" :active-tab="tab" @change="tab = $event" />
    </WorkspaceStickyStack>

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
              @click="saveMetadata"
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
            :disabled="busy || (saving)"
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
        <RegisterRow :interactive="false"
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
                class="btn btn-ghost btn-square btn-sm text-error"
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
    <AppPanel v-else title="History" subtitle="Persisted order history will appear here.">
      <EmptyState
        title="History is not available yet"
        description="The saved order remains unchanged while this workspace is being expanded."
      >
        <template #icon><PackageOpen :size="22" aria-hidden="true" /></template>
      </EmptyState>
    </AppPanel>
  </div>
</template>
