<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import AppPanel from '../../components/layout/AppPanel.vue'
import RegisterList from '../../components/ui/RegisterList.vue'
import RegisterRow from '../../components/ui/RegisterRow.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import InspectorShell from '../../components/layout/InspectorShell.vue'
import MasterDetail from '../../components/layout/MasterDetail.vue'
import FormGrid from '../../components/ui/FormGrid.vue';
import FormField from '../../components/ui/FormField.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import { computed, ref, watch } from 'vue';
import { ArrowLeft, ArrowDown, ArrowUp, Pencil, Plus, Trash2 } from 'lucide-vue-next';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import WorkspaceTabs from '../../components/layout/WorkspaceTabs.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import OrderItemConfigurator from '../sales/OrderItemConfigurator.vue';
import DocumentMetadataPanel from '../documents/DocumentMetadataPanel.vue';
import type { QuoteRecord, QuotePayload } from '../../api/quotes';
import { quotesApi } from '../../api/quotes';
import type { OrderItemPayload } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import SelectField from '../../components/ui/SelectField.vue';
const props = defineProps<{
  quote: QuoteRecord;
  currencyUnit: CurrencyUnit;
  customers: any[];
  services: any[];
  materials: any[];
}>();
const emit = defineEmits<{
  back: [];
  notify: [message: string];
  saved: [quote: QuoteRecord];
  converted: [orderId: string];
}>();
const tab = ref('Overview');
const customerId = ref('');
const expiryDate = ref<string | null>(null);
const notes = ref('');
const discountText = ref('');
const editorOpen = ref(false);
const editingItem = ref<any | null>(null);
const saving = ref(false);
const tabs = ['Overview', 'Items', 'Files'];
const customer = computed(() => props.customers.find((c) => c.id === customerId.value));
watch(
  () => props.quote,
  (q) => {
    customerId.value = q.customerId || '';
    expiryDate.value = q.expiryDate;
    notes.value = q.notes;
    discountText.value = formatMoneyInput(q.discountRial, props.currencyUnit);
  },
  { immediate: true },
);
function tone(v: string) {
  return v === 'Accepted' || v === 'Converted'
    ? 'green'
    : v === 'Sent'
      ? 'blue'
      : v === 'Rejected'
        ? 'red'
        : v === 'Expired'
          ? 'amber'
          : 'slate';
}
function money(v: number) {
  return formatMoney(v, props.currencyUnit);
}
function payload(): QuotePayload {
  return {
    customerId: customerId.value,
    expiryDate: expiryDate.value,
    notes: notes.value,
    discountRial: parseMoneyInput(discountText.value, props.currencyUnit) || 0,
  };
}
async function saveMetadata() {
return runAction(async () => {
  saving.value = true;
  try {
    emit('saved', await quotesApi.update(props.quote.id, payload()));
    emit('notify', 'Quote details saved');
  } catch (e) {
reportError(e);
  } finally {
    saving.value = false;
  }

});
}
async function configured(input: OrderItemPayload) {
return runAction(async () => {
  saving.value = true;
  try {
    const result = editingItem.value
      ? await quotesApi.replaceItem(props.quote.id, editingItem.value.id, input)
      : await quotesApi.addItem(props.quote.id, input);
    emit('saved', result);
    editorOpen.value = false;
    editingItem.value = null;
    emit('notify', editingItem.value ? 'Item replaced' : 'Item added');
  } catch (e) {
reportError(e);
  } finally {
    saving.value = false;
  }

});
}
async function remove(item: any) {
return runAction(async () => {
  try {
    emit('saved', await quotesApi.removeItem(props.quote.id, item.id));
    emit('notify', 'Item removed');
  } catch (e) {
reportError(e);
  }

});
}
async function move(item: any, delta: number) {
return runAction(async () => {
  const items = [...props.quote.items].sort((a, b) => a.position - b.position);
  const i = items.findIndex((x) => x.id === item.id);
  const n = i + delta;
  if (n < 0 || n >= items.length) return;
  [items[i], items[n]] = [items[n], items[i]];
  try {
    emit(
      'saved',
      await quotesApi.reorderItems(
        props.quote.id,
        items.map((x) => x.id),
      ),
    );
  } catch (e) {
reportError(e);
  }

});
}
async function changeStatus(value: string) {
return runAction(async () => {
  try {
    emit('saved', await quotesApi.status(props.quote.id, value));
    emit('notify', 'Quote status updated');
  } catch (err) {
reportError(err);
  }

});
}
async function discount() {
return runAction(async () => {
  try {
    const amount = parseMoneyInput(discountText.value, props.currencyUnit);
    if (amount === null) throw new Error('Enter a valid discount');
    emit('saved', await quotesApi.discount(props.quote.id, amount));
    emit('notify', 'Discount applied');
  } catch (e) {
reportError(e);
  }

});
}
async function convert() {
return runAction(async () => {
  if (props.quote.convertedOrderId) return;
  try {
    const result = await quotesApi.convert(props.quote.id);
    emit('saved', result);
    emit('converted', result.convertedOrderId);
    emit('notify', 'Quote converted to order');
  } catch (e) {
reportError(e);
  }

});
}
</script>
<template>
<div class="space-y-4">
<WorkspaceStickyStack flush>
<WorkspaceHeader eyebrow="Sales / quote workspace" :title="quote.quoteNumber" :description="quote.customerName || 'Walk-in customer'">
<button class="btn btn-ghost" @click="emit('back')"><ArrowLeft :size="16" />Quotes</button>
<StatusBadge :label="quote.status" :tone="tone(quote.status)" />
<SelectField :model-value="quote.status" aria-label="Quote status" :options="['Draft','Sent','Accepted','Rejected','Expired','Converted'].map(value=>({label:value,value}))" @update:model-value="changeStatus" />
<button v-if="quote.status==='Accepted' && !quote.convertedOrderId" class="btn btn-primary" :disabled="busy || (saving)" @click="convert">Convert to order</button>
</WorkspaceHeader>
<WorkspaceTabs :tabs="tabs" :active-tab="tab" @change="tab=$event" />
</WorkspaceStickyStack>
<MasterDetail v-if="tab==='Overview'">
<AppPanel title="Quote details" subtitle="Customer and commercial context">
<form id="quote-details" class="space-y-4" @submit.prevent="saveMetadata">
<FormGrid><SelectField v-model="customerId" label="Customer" :disabled="quote.status!=='Draft'" :options="[{label:'Walk-in customer',value:''},...customers.filter((x:any)=>x.active||x.id===customerId).map((c:any)=>({label:c.name,value:c.id}))]" />
<FormField label="Expiry date"><JalaliDatePicker v-model="expiryDate" /></FormField></FormGrid>
<FormField label="Notes"><AppTextarea v-model="notes" rows="4" :disabled="quote.status!=='Draft'" /></FormField>
<div class="flex items-center justify-between gap-3 border-t border-base-300 pt-3"><span class="text-xs text-base-content/60">Created {{formatDateTime(quote.createdAt)}}</span><button class="btn btn-primary" :disabled="saving || quote.status!=='Draft'">Save details</button></div>
</form>
</AppPanel>
<InspectorShell title="Quote summary" :subtitle="quote.items.length+' saved items'">
<dl class="space-y-3 text-sm">
<div class="flex justify-between gap-3"><dt class="text-base-content/60">Subtotal</dt><dd class="tabular-nums">{{money(quote.subtotalRial)}}</dd></div>
<div class="flex justify-between gap-3"><dt class="text-base-content/60">Discount</dt><dd class="tabular-nums">{{money(quote.discountRial)}}</dd></div>
<div class="flex justify-between gap-3 border-t border-base-300 pt-3 font-semibold"><dt>Total</dt><dd class="tabular-nums">{{money(quote.totalRial)}}</dd></div>
<div class="flex justify-between gap-3"><dt class="text-base-content/60">Estimated cost</dt><dd class="tabular-nums">{{money(quote.estimatedCostRial)}}</dd></div>
</dl>
<form class="space-y-3 border-t border-base-300 pt-3" @submit.prevent="discount"><FormField :label="'Discount ('+currencyUnit+')'"><AppInput v-model="discountText" inputmode="decimal" :disabled="quote.status!=='Draft'" /></FormField><button class="btn btn-outline w-full" :disabled="quote.status!=='Draft' || saving">Apply discount</button></form>
<p v-if="quote.convertedOrderId" class="text-xs text-base-content/60 wrap-anywhere">Linked order: {{quote.convertedOrderId}}</p>
</InspectorShell>
</MasterDetail>
<template v-else-if="tab==='Items'">
<AppPanel v-if="editorOpen" title="Configure service item"><OrderItemConfigurator :services="services" :materials="materials" :currency-unit="currencyUnit" document-label="quote" :initial="editingItem" :busy="saving" @configured="configured" @cancel="editorOpen=false;editingItem=null" /></AppPanel>
<RegisterList v-else title="Configured items" subtitle="Each item keeps its accepted pricing snapshot." :count="quote.items.length">
<template #action><button class="btn btn-primary btn-sm" :disabled="quote.status!=='Draft'" @click="editorOpen=true"><Plus :size="15" />Add service item</button></template>
<EmptyState v-if="!quote.items.length" title="No configured items" description="Add a service item to prepare this quote." />
<RegisterRow v-for="item in [...quote.items].sort((a,b)=>a.position-b.position)" :key="item.id" :interactive="false">
<template #identity><strong class="block text-sm">{{item.serviceName}}</strong></template>
<template #meta><p class="mt-1 text-xs text-base-content/60">{{item.quantity}} {{item.quantityUnit}} · {{item.notes || 'Accepted snapshot'}}</p></template>
<template #status><strong class="whitespace-nowrap text-sm tabular-nums">{{money(item.sellingPriceRial)}}</strong></template>
<template #actions><div class="flex flex-wrap items-center gap-1">
<button class="btn btn-ghost btn-square btn-sm" aria-label="Move item up" :disabled="busy || (quote.status!=='Draft')" @click="move(item,-1)"><ArrowUp :size="14" /></button>
<button class="btn btn-ghost btn-square btn-sm" aria-label="Move item down" :disabled="busy || (quote.status!=='Draft')" @click="move(item,1)"><ArrowDown :size="14" /></button>
<button class="btn btn-ghost btn-square btn-sm" aria-label="Reconfigure item" :disabled="quote.status!=='Draft'" @click="editingItem=item;editorOpen=true"><Pencil :size="14" /></button>
<button class="btn btn-ghost btn-square btn-sm text-error" aria-label="Remove item" :disabled="busy || (quote.status!=='Draft')" @click="remove(item)"><Trash2 :size="14" /></button>
</div></template>
</RegisterRow>
</RegisterList>
</template>
<DocumentMetadataPanel v-else owner-type="quote" :owner-id="quote.id" :protected-context="quote.status!=='Draft'" @notify="emit('notify',$event)" />
</div>
</template>
