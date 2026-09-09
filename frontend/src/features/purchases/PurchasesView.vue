<script setup lang="ts">
import { nextTick, watch } from 'vue';
import { Plus, ShoppingCart } from 'lucide-vue-next';
import EmptyState from '../../components/ui/EmptyState.vue';
import LoadingState from '../../components/ui/LoadingState.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import SelectField from '../../components/ui/SelectField.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import PurchaseWorkspaceView from './PurchaseWorkspaceView.vue';
import type { SupplierRecord } from '../../api/suppliers';
import type { MaterialRecord } from '../../api/materials';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import { usePurchasesWorkspace } from './usePurchasesWorkspace';

const props = defineProps<{
  currencyUnit: CurrencyUnit;
  suppliers: SupplierRecord[];
  materials: MaterialRecord[];
}>();
const emit = defineEmits<{ notify: [string] }>();
const workspace = usePurchasesWorkspace(props, emit);

const {
  rows,
  filteredRows,
  selectedId,
  current,
  createMode,
  searchQuery,
  purchaseFilter,
  isLoading,
  editing,
  startCreate,
  select,
  backToPurchases,
} = workspace;

watch(
  [selectedId, editing, createMode],
  () => {
    void nextTick(() => {
      document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' });
    });
  },
);

function statusTone(status: string) {
  return status === 'Posted' ? 'green' : status === 'Archived' || status === 'Cancelled' ? 'slate' : 'amber';
}
</script>

<template>
  <div v-if="!current" class="min-w-0 space-y-3">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        title="Purchases"
        eyebrow="Purchasing / inventory ledger"
        description="Record supplier purchases, then post them into the immutable inventory ledger."
      >
        <button class="btn btn-primary" type="button" @click="startCreate">
          <Plus :size="16" aria-hidden="true" />New purchase
        </button>
      </WorkspaceHeader>
      <SearchFilterBar>
        <template #search>
          <SearchField
            v-model="searchQuery"
            label="Search purchases"
            placeholder="Search purchase, supplier, or invoice"
          />
        </template>
        <template #filters>
          <SelectField
            v-model="purchaseFilter"
            label="Status"
            aria-label="Filter purchases by status"
            :options="['All', 'Draft', 'Posted', 'Cancelled', 'Archived'].map((value) => ({ label: value, value }))"
          />
        </template>
        <template #count>
          <span class="whitespace-nowrap">{{ filteredRows.length }} of {{ rows.length }} purchases</span>
        </template>
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <RegisterList
      title="Purchase register"
      subtitle="Open a purchase to edit drafts, manage items, or review posted history."
      :count="filteredRows.length"
    >
      <LoadingState v-if="isLoading" label="Loading purchases…" />
      <EmptyState
        v-else-if="!filteredRows.length"
        :title="rows.length ? 'No purchases match this view' : 'No purchases yet'"
        :description="
          rows.length
            ? 'Try another status or search term.'
            : 'Record the first supplier purchase to establish an inventory ledger.'
        "
      >
        <template #icon><ShoppingCart :size="21" :stroke-width="1.8" aria-hidden="true" /></template>
        <template v-if="!rows.length" #action>
          <button class="btn btn-ghost" type="button" @click="startCreate">
            <Plus :size="15" aria-hidden="true" />Record purchase
          </button>
        </template>
      </EmptyState>
      <template v-else>
        <RegisterRow
          v-for="purchase in filteredRows"
          :key="purchase.id"
          :selected="selectedId === purchase.id"
          @activate="select(purchase.id)"
        >
          <template #icon><ShoppingCart :size="17" :stroke-width="1.8" aria-hidden="true" /></template>
          <template #identity>
            <div class="flex min-w-0 items-start justify-between gap-3">
              <div class="min-w-0">
                <strong class="block truncate text-sm">{{ purchase.purchaseNumber }}</strong>
                <span class="block truncate text-xs text-base-content/60">
                  {{ purchase.supplierName || 'No supplier' }} · {{ purchase.supplierInvoiceNumber || 'No supplier invoice' }}
                </span>
              </div>
              <div class="flex shrink-0 items-center gap-1.5">
                <StatusBadge :label="purchase.status" :tone="statusTone(purchase.status)" />
              </div>
            </div>
          </template>
          <template #meta>
            <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2 xl:grid-cols-4">
              <div class="min-w-0">
                <span class="block text-base-content/50">Purchase date</span>
                <span class="block break-words text-base-content/80">{{ formatDateTime(purchase.purchaseDate) }}</span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Items</span>
                <span class="block text-base-content/80">{{ purchase.items.length }} line items</span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Paid</span>
                <span class="block break-words text-base-content/80 tabular-nums">{{ formatMoney(purchase.paidRial || 0, props.currencyUnit) }}</span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Total / remaining</span>
                <span class="block break-words text-base-content/80 tabular-nums">{{ formatMoney(purchase.totalRial, props.currencyUnit) }} / {{ formatMoney(purchase.remainingRial ?? purchase.totalRial, props.currencyUnit) }}</span>
              </div>
            </div>
          </template>
        </RegisterRow>
      </template>
    </RegisterList>
  </div>

  <PurchaseWorkspaceView
    v-else
    :workspace="workspace"
    :currency-unit="props.currencyUnit"
    :suppliers="props.suppliers"
    :materials="props.materials"
    @back="backToPurchases"
  />
</template>
