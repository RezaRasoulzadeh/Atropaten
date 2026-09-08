<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import FormGrid from '../../components/ui/FormGrid.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import MasterDetail from '../../components/layout/MasterDetail.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import { computed, onMounted, ref } from 'vue';
import { ArrowDown, ArrowUp, Check, Plus, Save, ShoppingCart, Trash2, X } from 'lucide-vue-next';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import SelectField from '../../components/ui/SelectField.vue';
import { purchasesApi, type PurchasePayload, type PurchaseRecord } from '../../api/purchases';
import type { SupplierRecord } from '../../api/suppliers';
import type { MaterialRecord } from '../../api/materials';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { currentCanonicalDate, formatDateTime } from '../../utils/date';
import { confirmAction, useToast } from '../../ui/feedback';

const props = defineProps<{
  currencyUnit: CurrencyUnit;
  suppliers: SupplierRecord[];
  materials: MaterialRecord[];
}>();
const emit = defineEmits<{ notify: [string] }>();
const rows = ref<PurchaseRecord[]>([]);
const selectedId = ref<string | null>(null);
const loading = ref(false);
const toast = useToast();
const editing = ref(false);
const form = ref<PurchasePayload>(blank());
const item = ref({
  materialId: '',
  purchaseQuantity: '1',
  unitAcquisitionCostRial: '0',
  notes: '',
});
const current = computed(() => rows.value.find((value) => value.id === selectedId.value) ?? null);

function blank(): PurchasePayload {
  return {
    supplierId: props.suppliers[0]?.id ?? '',
    purchaseDate: currentCanonicalDate(),
    supplierInvoiceNumber: '',
    notes: '',
    discountRial: 0,
    shippingRial: 0,
    taxRial: 0,
    additionalCostsRial: 0,
  };
}
function dateOnly(value: string) {
  return value.slice(0, 10);
}
onMounted(load);
async function load() {
  loading.value = true;
  try {
    rows.value = await purchasesApi.list();
  } catch (value) {
    reportError(value);
  } finally {
    loading.value = false;
  }
}
function select(value: PurchaseRecord) {
  selectedId.value = value.id;
  editing.value = value.status === 'Draft';
  form.value = {
    supplierId: value.supplierId,
    purchaseDate: dateOnly(value.purchaseDate),
    supplierInvoiceNumber: value.supplierInvoiceNumber,
    notes: value.notes,
    discountRial: value.discountRial,
    shippingRial: value.shippingRial,
    taxRial: value.taxRial,
    additionalCostsRial: value.additionalCostsRial,
  };
}
async function create() {
return runAction(async () => {
  if (!props.suppliers.length) {
    toast.error('Create a supplier before recording a purchase.', 'Purchases');
    return;
  }
  try {
    const value = await purchasesApi.create(blank());
    rows.value = [value, ...rows.value];
    select(value);
    emit('notify', 'Draft purchase created.');
  } catch (value) {
reportError(value);
  }

});
}
async function save() {
return runAction(async () => {
  if (!current.value) return;
  try {
    replace(await purchasesApi.update(current.value.id, form.value));
    emit('notify', 'Purchase saved.');
  } catch (value) {
reportError(value);
  }

});
}
async function addItem() {
return runAction(async () => {
  if (!current.value) return;
  try {
    replace(await purchasesApi.addItem(current.value.id, item.value));
    item.value = { materialId: '', purchaseQuantity: '1', unitAcquisitionCostRial: '0', notes: '' };
  } catch (value) {
reportError(value);
  }

});
}
async function removeItem(id: string) {
return runAction(async () => {
  if (!current.value) return;
  try {
    replace(await purchasesApi.removeItem(current.value.id, id));
  } catch (value) {
reportError(value);
  }

});
}
async function reorder(index: number, direction: number) {
return runAction(async () => {
  if (!current.value) return;
  const items = [...current.value.items].sort((a, b) => a.position - b.position);
  const target = index + direction;
  if (target < 0 || target >= items.length) return;
  [items[index], items[target]] = [items[target], items[index]];
  try {
    replace(
      await purchasesApi.reorder(
        current.value.id,
        items.map((value) => value.id),
      ),
    );
  } catch (value) {
reportError(value);
  }

});
}
async function post() {
return runAction(async () => {
  if (
    !current.value ||
    !(await confirmAction({
      title: 'Post purchase',
      message: 'Post this purchase and create inventory movements?',
      confirmLabel: 'Post purchase',
    }))
  )
    return;
  try {
    replace(await purchasesApi.post(current.value.id));
    editing.value = false;
    emit('notify', 'Purchase posted.');
  } catch (value) {
reportError(value);
  }

});
}
async function cancel() {
return runAction(async () => {
  if (
    !current.value ||
    !(await confirmAction({
      title: 'Cancel purchase',
      message: 'Cancel this posted purchase? This creates compensating movements.',
      confirmLabel: 'Cancel purchase',
      danger: true,
    }))
  )
    return;
  try {
    replace(await purchasesApi.cancel(current.value.id));
    editing.value = false;
    emit('notify', 'Purchase cancelled with history preserved.');
  } catch (value) {
reportError(value);
  }

});
}
async function removeDraft() {
return runAction(async () => {
  if (
    !current.value ||
    !(await confirmAction({
      title: 'Delete purchase draft',
      message: 'Delete this draft permanently?',
      confirmLabel: 'Delete draft',
      danger: true,
    }))
  )
    return;
  try {
    await purchasesApi.deleteDraft(current.value.id);
    rows.value = rows.value.filter((value) => value.id !== current.value!.id);
    selectedId.value = null;
    emit('notify', 'Draft deleted.');
  } catch (value) {
reportError(value);
  }

});
}
function replace(value: PurchaseRecord) {
  rows.value = rows.value.map((item) => (item.id === value.id ? value : item));
  selectedId.value = value.id;
  form.value = {
    supplierId: value.supplierId,
    purchaseDate: dateOnly(value.purchaseDate),
    supplierInvoiceNumber: value.supplierInvoiceNumber,
    notes: value.notes,
    discountRial: value.discountRial,
    shippingRial: value.shippingRial,
    taxRial: value.taxRial,
    additionalCostsRial: value.additionalCostsRial,
  };
}
</script>

<template>
  <div class="min-w-0 space-y-3">
    <WorkspaceStickyStack
      ><WorkspaceHeader
        title="Purchases"
        eyebrow="Catalog / inventory"
        description="Record supplier purchases, then post them into the immutable inventory ledger."
        ><button class="btn btn-primary" @click="create" :disabled="busy">
          <Plus :size="16" /> New purchase
        </button></WorkspaceHeader
      ></WorkspaceStickyStack
    >
    <MasterDetail>
      <RegisterList
        title="Purchase register"
        subtitle="Drafts are editable; posted history is protected."
        :count="rows.length"
        ><LoadingState v-if="loading" label="Loading records…" />
        <div v-else-if="rows.length">
          <RegisterRow
            v-for="value in rows"
            :key="value.id"
            :selected="selectedId === value.id"
            @activate="select(value)"
          >
            <template #identity>
              <div class="flex min-w-0 items-center justify-between gap-3">
                <div class="min-w-0">
                  <strong class="block truncate text-sm">{{ value.purchaseNumber }}</strong>
                  <span class="block truncate text-xs text-base-content/60">{{ value.supplierName }} · {{ value.supplierInvoiceNumber || 'No supplier invoice' }}</span>
                </div>
                <strong class="shrink-0 whitespace-nowrap text-sm tabular-nums">{{ formatMoney(value.totalRial, props.currencyUnit) }}</strong>
              </div>
            </template>
            <template #meta>
              <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2">
                <div><span class="block text-base-content/50">Purchase date</span><span class="block text-base-content/80">{{ formatDateTime(value.purchaseDate) }}</span></div>
                <div><span class="block text-base-content/50">Paid</span><span class="block text-base-content/80 tabular-nums">{{ formatMoney(value.paidRial || 0, props.currencyUnit) }}</span></div>
                <div><span class="block text-base-content/50">Remaining</span><span class="block text-base-content/80 tabular-nums">{{ formatMoney(value.remainingRial ?? value.totalRial, props.currencyUnit) }}</span></div>
                <div><span class="block text-base-content/50">Items</span><span class="block text-base-content/80">{{ value.items.length }} line items</span></div>
              </div>
            </template>
            <template #status><StatusBadge :label="value.status" :tone="value.status === 'Posted' ? 'green' : value.status === 'Cancelled' ? 'slate' : 'amber'" /></template>
          </RegisterRow>
        </div>
        <div v-else class="min-w-0 space-y-3">
          <ShoppingCart :size="22" />
          <p>No purchases yet.</p>
          <button class="btn btn-ghost" @click="create" :disabled="busy"><Plus :size="15" /> Record purchase</button>
        </div>
      </RegisterList>
      <InspectorShell
        v-if="current"
        title="Purchase editor"
        :subtitle="
          current.status === 'Draft' ? 'Draft purchase' : 'Historical purchase · read only'
        "
        ><div class="min-w-0 space-y-3">
          <StatusBadge
            :label="current.status"
            :tone="
              current.status === 'Posted'
                ? 'green'
                : current.status === 'Cancelled'
                  ? 'slate'
                  : 'amber'
            "
          /><span>{{ current.purchaseNumber }}</span>
        </div>
        <form v-if="editing" @submit.prevent="save" class="min-w-0 space-y-3">
          <FormGrid
            ><SelectField
                v-model="form.supplierId"
                label="Supplier"
                :options="[
                  { label: 'Select supplier', value: '' },
                  ...props.suppliers.map((supplier) => ({
                    label: supplier.name,
                    value: supplier.id,
                  })),
                ]" />
            <FormField class="gap-1"
              ><span>Purchase date</span
              ><JalaliDatePicker
                v-model="form.purchaseDate"
                placeholder="Select Jalali date" /></FormField></FormGrid
          ><FormField class="gap-1"
            ><span>Supplier invoice</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="form.supplierInvoiceNumber" /></FormField
          ><FormGrid
            ><FormField class="gap-1"
              ><span>Discount Rial</span
              ><AppInput
                class="input w-full min-w-0"
                v-model.number="form.discountRial"
                type="number"
                min="0" /></FormField
            ><FormField class="gap-1"
              ><span>Shipping Rial</span
              ><AppInput
                class="input w-full min-w-0"
                v-model.number="form.shippingRial"
                type="number"
                min="0" /></FormField
            ><FormField class="gap-1"
              ><span>Tax Rial</span
              ><AppInput
                class="input w-full min-w-0"
                v-model.number="form.taxRial"
                type="number"
                min="0" /></FormField
            ><FormField class="gap-1"
              ><span>Additional Rial</span
              ><AppInput
                class="input w-full min-w-0"
                v-model.number="form.additionalCostsRial"
                type="number"
                min="0" /></FormField></FormGrid
          ><FormField class="gap-1"
            ><span>Notes</span
            ><AppTextarea
              v-model="form.notes"
              rows="2"
            /></FormField
          ><button class="btn btn-primary" type="submit" :disabled="busy"><Save :size="15" /> Save metadata</button>
        </form>
        <div class="min-w-0 space-y-3">
          <h3 class="text-sm font-semibold">
            Items <span>{{ current.items.length }}</span>
          </h3>
          <div
            v-for="(line, index) in [...current.items].sort((a, b) => a.position - b.position)"
            :key="line.id"
            class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
          >
            <div class="min-w-0 space-y-1">
              <strong>{{ line.materialName }}</strong
              ><small class="block text-xs leading-5 text-base-content/60"
                >{{ line.purchaseQuantity }} {{ line.purchaseUnit }} →
                {{ line.consumptionQuantity }} {{ line.consumptionUnit }}</small
              >
            </div>
            <span>{{ formatMoney(line.lineTotalRial, props.currencyUnit) }}</span
            ><span v-if="editing"
              ><button
                class="btn btn-ghost"
                :disabled="busy || (index === 0)"
                aria-label="Move line up"
                @click="reorder(index, -1)"
              >
                <ArrowUp :size="13" /></button
              ><button
                class="btn btn-ghost"
                :disabled="busy || (index === current.items.length - 1)"
                aria-label="Move line down"
                @click="reorder(index, 1)"
              >
                <ArrowDown :size="13" /></button></span
            ><button
              class="btn btn-outline btn-error"
              v-if="editing"
              @click="removeItem(line.id)"
              aria-label="Remove item"
             :disabled="busy">
              <Trash2 :size="14" />
            </button>
          </div>
          <form v-if="editing" @submit.prevent="addItem" class="min-w-0 space-y-3">
            <SelectField
              v-model="item.materialId"
              label="Material"
              :options="[
                { label: 'Material', value: '' },
                ...props.materials.map((material) => ({
                  label: material.name,
                  value: material.id,
                })),
              ]"
            /><FormField label="Qty"><AppInput
              class="input w-full min-w-0"
              v-model="item.purchaseQuantity"
              required
              placeholder="Qty"
              inputmode="decimal"
            /></FormField><FormField label="Unit cost Rial"><AppInput
              class="input w-full min-w-0"
              v-model="item.unitAcquisitionCostRial"
              required
              placeholder="Unit cost Rial"
              inputmode="numeric"
            /></FormField><button class="btn btn-ghost"><Plus :size="15" /> Add line</button>
          </form>
        </div>
        <dl class="grid min-w-0 gap-2 text-sm">
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Subtotal</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ formatMoney(current.subtotalRial, props.currencyUnit) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Discount / costs</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ formatMoney(current.discountRial, props.currencyUnit) }} /
              {{
                formatMoney(
                  current.shippingRial + current.taxRial + current.additionalCostsRial,
                  props.currencyUnit,
                )
              }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Total</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ formatMoney(current.totalRial, props.currencyUnit) }}
            </dd>
          </div>
        </dl>
        <div class="flex flex-wrap items-center gap-2">
          <button class="btn btn-primary" v-if="current.status === 'Draft'" @click="post" :disabled="busy">
            <Check :size="15" /> Post purchase</button
          ><button class="btn btn-outline btn-error" v-if="current.status === 'Draft'" @click="removeDraft" :disabled="busy">
            <Trash2 :size="15" /> Delete draft</button
          ><button class="btn btn-ghost" v-if="current.status === 'Posted'" @click="cancel" :disabled="busy">
            Cancel / reverse
          </button>
        </div></InspectorShell
      >
      <InspectorShell
        v-else
        title="Purchase inspector"
        subtitle="Select a purchase or start a new draft."
        ><div class="min-w-0 space-y-3">
          <ShoppingCart :size="20" />
          <p>Purchase details will appear here.</p>
        </div></InspectorShell
      >
    </MasterDetail>
  </div>
</template>
