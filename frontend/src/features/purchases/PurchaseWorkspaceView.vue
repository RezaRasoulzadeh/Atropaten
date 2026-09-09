<script setup lang="ts">
import {
  ArrowDown,
  ArrowLeft,
  ArrowUp,
  Archive,
  Check,
  Edit3,
  Plus,
  Save,
  Trash2,
} from 'lucide-vue-next';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import FormField from '../../components/ui/FormField.vue';
import FormGrid from '../../components/ui/FormGrid.vue';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import SelectField from '../../components/ui/SelectField.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import type { usePurchasesWorkspace } from './usePurchasesWorkspace';

const props = defineProps<{
  workspace: ReturnType<typeof usePurchasesWorkspace>;
  currencyUnit: CurrencyUnit;
  suppliers: any[];
  materials: any[];
}>();
const emit = defineEmits<{ back: [] }>();

const {
  busy,
  current,
  createMode,
  editing,
  editingItemId,
  form,
  item,
  financial,
  startEdit,
  startItemEdit,
  cancelItemEdit,
  cancelEditor,
  save,
  addItem,
  removeItem,
  reorder,
  post,
  archivePurchase,
  removePurchase,
} = props.workspace;

function statusTone(status: string) {
  return status === 'Posted' ? 'green' : status === 'Archived' || status === 'Cancelled' ? 'slate' : 'amber';
}
function updatePurchaseMoney(field: 'discountRial' | 'shippingRial' | 'taxRial' | 'additionalCostsRial', value: string) {
  const parsed = parseMoneyInput(value, props.currencyUnit);
  if (parsed !== null) props.workspace.form.value[field] = parsed;
}
function financialAccountLabel(id: string) {
  const account = financial.value.find((value) => value.id === id);
  if (!account) return 'No account selected';
  const identity = account.type === 'bank'
    ? [account.bankName, account.accountNumber && `Account ${account.accountNumber}`, account.cardNumber && `Card ${account.cardNumber}`].filter(Boolean).join(' · ')
    : 'Cash account';
  return identity ? `${account.name} · ${identity}` : account.name;
}
</script>

<template>
  <div v-if="current" class="min-w-0 space-y-4" aria-label="Purchase workspace">
    <WorkspaceStickyStack :flush="true">
      <WorkspaceHeader
        :title="createMode === 'create' ? 'New purchase' : current.purchaseNumber"
        eyebrow="Purchasing / purchase workspace"
            :description="
          createMode === 'create'
            ? 'Build the purchase locally. Nothing is saved until you choose Save purchase.'
            : current.status === 'Draft'
            ? 'Review and complete this supplier purchase before posting inventory movements.'
            : current.status === 'Posted'
              ? 'Review the posted purchase and its immutable inventory history.'
              : current.status === 'Archived'
                ? 'Archived purchase history. It remains available for review.'
                : 'Review the purchase and its compensating inventory history.'
        "
      >
        <template #title-suffix>
          <StatusBadge :label="createMode ? 'Unsaved' : current.status" :tone="createMode ? 'slate' : statusTone(current.status)" />
        </template>
        <button class="btn btn-ghost btn-sm gap-2" type="button" :disabled="busy" @click="emit('back')">
          <ArrowLeft :size="16" aria-hidden="true" />Purchases
        </button>
        <template v-if="editing">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="cancelEditor">Cancel</button>
          <button class="btn btn-primary btn-sm gap-2" type="submit" form="purchase-editor" :disabled="busy">
            <Save :size="14" aria-hidden="true" />Save purchase
          </button>
        </template>
        <template v-else>
          <button class="btn btn-outline btn-error btn-sm gap-2" type="button" :disabled="busy" @click="removePurchase">
            <Trash2 :size="14" aria-hidden="true" />Delete
          </button>
          <button v-if="current.status !== 'Archived'" class="btn btn-outline btn-warning btn-sm gap-2" type="button" :disabled="busy" @click="archivePurchase">
            <Archive :size="14" aria-hidden="true" />Archive
          </button>
          <button v-if="current.status === 'Draft'" class="btn btn-primary btn-sm gap-2" type="button" :disabled="busy" @click="post">
            <Check :size="14" aria-hidden="true" />Post purchase
          </button>
          <button class="btn btn-outline btn-sm gap-2" type="button" :disabled="busy" @click="startEdit">
            <Edit3 :size="14" aria-hidden="true" />Edit
          </button>
        </template>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <AppPanel
      v-if="editing"
      title="Purchase details"
      :subtitle="createMode ? 'Nothing is saved until you choose Save purchase.' : 'Draft metadata and items can be changed before posting.'"
    >
      <form id="purchase-editor" class="min-w-0 space-y-4" @submit.prevent="save">
        <FormGrid>
          <SelectField
            v-model="form.supplierId"
            label="Supplier"
            :options="[
              { label: 'Select supplier', value: '' },
              ...props.suppliers.map((supplier) => ({ label: supplier.name, value: supplier.id })),
            ]"
          />
          <FormField class="gap-1">
            <span>Purchase date</span>
            <JalaliDatePicker v-model="form.purchaseDate" placeholder="Select Jalali date" />
          </FormField>
          <SelectField
            v-model="form.financialAccountId"
            label="Paid from account"
            :options="[
              { label: 'Select cash or bank account', value: '' },
              ...financial.filter((account) => account.active).map((account) => ({ label: financialAccountLabel(account.id), value: account.id })),
            ]"
          />
        </FormGrid>
        <FormField class="gap-1">
          <span>Supplier invoice</span>
          <AppInput v-model="form.supplierInvoiceNumber" class="input w-full min-w-0" />
        </FormField>
        <FormGrid>
          <FormField class="gap-1">
            <span>Discount ({{ props.currencyUnit }})</span>
            <AppInput :model-value="formatMoneyInput(form.discountRial, props.currencyUnit)" :money="props.currencyUnit" class="input w-full min-w-0" @update:model-value="updatePurchaseMoney('discountRial', $event)" />
          </FormField>
          <FormField class="gap-1">
            <span>Shipping ({{ props.currencyUnit }})</span>
            <AppInput :model-value="formatMoneyInput(form.shippingRial, props.currencyUnit)" :money="props.currencyUnit" class="input w-full min-w-0" @update:model-value="updatePurchaseMoney('shippingRial', $event)" />
          </FormField>
          <FormField class="gap-1">
            <span>Tax ({{ props.currencyUnit }})</span>
            <AppInput :model-value="formatMoneyInput(form.taxRial, props.currencyUnit)" :money="props.currencyUnit" class="input w-full min-w-0" @update:model-value="updatePurchaseMoney('taxRial', $event)" />
          </FormField>
          <FormField class="gap-1">
            <span>Additional costs ({{ props.currencyUnit }})</span>
            <AppInput :model-value="formatMoneyInput(form.additionalCostsRial, props.currencyUnit)" :money="props.currencyUnit" class="input w-full min-w-0" @update:model-value="updatePurchaseMoney('additionalCostsRial', $event)" />
          </FormField>
        </FormGrid>
        <FormField class="gap-1">
          <span>Notes</span>
          <AppTextarea v-model="form.notes" rows="3" />
        </FormField>
      </form>
    </AppPanel>

    <AppPanel title="Purchase items" subtitle="Items are converted into inventory quantities when the purchase is posted.">
      <div class="min-w-0 space-y-3">
        <div
          v-for="(line, index) in [...current.items].sort((a, b) => a.position - b.position)"
          :key="line.id"
          class="grid min-w-0 gap-3 border-b border-base-300 py-3 last:border-0 sm:grid-cols-[minmax(0,1fr)_auto_auto] sm:items-center"
        >
          <div class="min-w-0">
            <strong class="block break-words">{{ line.materialName }}</strong>
            <small class="mt-1 block break-words text-xs leading-5 text-base-content/60">
              {{ line.purchaseQuantity }} {{ line.purchaseUnit }} → {{ line.consumptionQuantity }} {{ line.consumptionUnit }}
              <span v-if="line.notes"> · {{ line.notes }}</span>
            </small>
          </div>
          <span class="text-sm tabular-nums sm:text-end">{{ formatMoney(line.lineTotalRial, props.currencyUnit) }}</span>
          <div v-if="current.status !== 'Archived'" class="flex flex-wrap justify-start gap-1 sm:justify-end">
            <button class="btn btn-outline btn-sm" type="button" :disabled="busy" aria-label="Edit item" @click="startItemEdit(line)">
              <Edit3 :size="13" aria-hidden="true" />
            </button>
            <button v-if="editing" class="btn btn-ghost btn-sm" type="button" :disabled="busy || index === 0" aria-label="Move line up" @click="reorder(index, -1)">
              <ArrowUp :size="13" aria-hidden="true" />
            </button>
            <button v-if="editing" class="btn btn-ghost btn-sm" type="button" :disabled="busy || index === current.items.length - 1" aria-label="Move line down" @click="reorder(index, 1)">
              <ArrowDown :size="13" aria-hidden="true" />
            </button>
            <button class="btn btn-outline btn-error btn-sm" type="button" :disabled="busy" aria-label="Remove item" @click="removeItem(line.id)">
              <Trash2 :size="13" aria-hidden="true" />
            </button>
          </div>
        </div>
        <p v-if="!current.items.length" class="rounded-box border border-dashed border-base-300 p-3 text-sm text-base-content/60">No items added yet.</p>
      </div>

      <form v-if="editing" class="mt-4 grid min-w-0 gap-3 border-t border-base-300 pt-4 sm:grid-cols-2 xl:grid-cols-[minmax(0,1.2fr)_minmax(0,0.7fr)_minmax(0,0.9fr)_auto]" @submit.prevent="addItem">
        <SelectField
          v-model="item.materialId"
          label="Material"
          :options="[
            { label: 'Select material', value: '' },
            ...props.materials.map((material) => ({ label: material.name, value: material.id })),
          ]"
        />
        <FormField class="gap-1">
          <span>Quantity</span>
          <AppInput v-model="item.purchaseQuantity" class="input w-full min-w-0" required inputmode="decimal" placeholder="Qty" />
        </FormField>
        <FormField class="gap-1">
          <span>Unit cost ({{ props.currencyUnit }})</span>
          <AppInput v-model="item.unitAcquisitionCostRial" :money="props.currencyUnit" class="input w-full min-w-0" required inputmode="numeric" placeholder="Unit cost" />
        </FormField>
        <div class="flex items-end gap-2">
          <button class="btn btn-outline btn-sm gap-2" type="submit" :disabled="busy">
            <Plus :size="14" aria-hidden="true" />{{ editingItemId ? 'Update item' : 'Add item' }}
          </button>
          <button v-if="editingItemId" class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="cancelItemEdit">Cancel</button>
        </div>
      </form>
    </AppPanel>

    <AppPanel title="Purchase summary" subtitle="Supplier, dates, and financial totals for this purchase.">
      <div class="grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <div class="min-w-0">
          <span class="block text-xs text-base-content/60">Supplier</span>
          <strong class="mt-1 block break-words text-sm">{{ current.supplierName || 'No supplier' }}</strong>
        </div>
        <div class="min-w-0">
          <span class="block text-xs text-base-content/60">Purchase date</span>
          <strong class="mt-1 block text-sm">{{ formatDateTime(current.purchaseDate) }}</strong>
        </div>
        <div class="min-w-0">
          <span class="block text-xs text-base-content/60">Invoice</span>
          <strong class="mt-1 block break-words text-sm">{{ current.supplierInvoiceNumber || 'No invoice number' }}</strong>
        </div>
        <div class="min-w-0">
          <span class="block text-xs text-base-content/60">Paid from</span>
          <strong class="mt-1 block break-words text-sm">{{ financialAccountLabel(current.financialAccountId) }}</strong>
        </div>
        <div class="min-w-0">
          <span class="block text-xs text-base-content/60">Payment</span>
          <strong class="mt-1 block text-sm tabular-nums">{{ formatMoney(current.paidRial || 0, props.currencyUnit) }} paid</strong>
          <span class="mt-1 block text-xs text-base-content/60 tabular-nums">{{ formatMoney(current.remainingRial ?? current.totalRial, props.currencyUnit) }} remaining</span>
        </div>
      </div>
      <dl class="mt-4 grid min-w-0 gap-2 border-t border-base-300 pt-3 text-sm sm:grid-cols-3">
        <div class="flex min-w-0 items-center justify-between gap-3 sm:block">
          <dt class="text-xs text-base-content/60">Subtotal</dt>
          <dd class="mt-1 tabular-nums sm:block">{{ formatMoney(current.subtotalRial, props.currencyUnit) }}</dd>
        </div>
        <div class="flex min-w-0 items-center justify-between gap-3 sm:block">
          <dt class="text-xs text-base-content/60">Discount / costs</dt>
          <dd class="mt-1 tabular-nums sm:block">{{ formatMoney(current.discountRial, props.currencyUnit) }} / {{ formatMoney(current.shippingRial + current.taxRial + current.additionalCostsRial, props.currencyUnit) }}</dd>
        </div>
        <div class="flex min-w-0 items-center justify-between gap-3 sm:block">
          <dt class="text-xs text-base-content/60">Total</dt>
          <dd class="mt-1 font-semibold tabular-nums sm:block">{{ formatMoney(current.totalRial, props.currencyUnit) }}</dd>
        </div>
      </dl>
      <p v-if="current.notes" class="mt-4 whitespace-pre-wrap break-words border-t border-base-300 pt-3 text-sm text-base-content/75">{{ current.notes }}</p>
    </AppPanel>
  </div>
</template>
