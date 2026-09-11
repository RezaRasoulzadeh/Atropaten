<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, Plus, Search, ShoppingCart } from 'lucide-vue-next'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import SearchField from '../../components/ui/SearchField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { SupplierRecord } from '../../api/suppliers'
import type { MaterialRecord } from '../../api/materials'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'
import PurchaseDetailPanel from './PurchaseDetailPanel.vue'
import PurchaseWorkspaceView from './PurchaseWorkspaceView.vue'
import { usePurchasesWorkspace, type PurchaseFilter } from './usePurchasesWorkspace'

const props = defineProps<{
  currencyUnit: CurrencyUnit
  suppliers: SupplierRecord[]
  materials: MaterialRecord[]
}>()
const emit = defineEmits<{ notify: [string] }>()
const workspace = usePurchasesWorkspace(props, emit)

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
} = workspace

const sortOrder = ref('date')
const page = ref(1)
const pageSize = 10
const statusOptions: PurchaseFilter[] = ['All', 'Draft', 'Posted', 'Archived']

const visibleRows = computed(() => [...filteredRows.value].sort((left, right) => {
  if (sortOrder.value === 'total') return right.totalRial - left.totalRial
  if (sortOrder.value === 'supplier') return `${left.supplierName}${left.purchaseNumber}`.localeCompare(`${right.supplierName}${right.purchaseNumber}`)
  return String(right.purchaseDate).localeCompare(String(left.purchaseDate))
}))
const pageCount = computed(() => Math.max(1, Math.ceil(visibleRows.value.length / pageSize)))
const pagedRows = computed(() => visibleRows.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const pageNumbers = computed(() => Array.from({ length: pageCount.value }, (_, index) => index + 1))

function statusTone(status: string) {
  return status === 'Posted' ? 'green' : status === 'Archived' || status === 'Cancelled' ? 'slate' : 'amber'
}

function statusCount(status: PurchaseFilter) {
  if (status === 'All') return rows.value.length
  return rows.value.filter((purchase) => purchase.status === status).length
}

function goToPage(value: number) {
  page.value = Math.min(Math.max(value, 1), pageCount.value)
}

function resetListFilters() {
  searchQuery.value = ''
  purchaseFilter.value = 'All'
  sortOrder.value = 'date'
  page.value = 1
}

watch([searchQuery, purchaseFilter, sortOrder], () => { page.value = 1 })
watch(pageCount, (count) => { if (page.value > count) page.value = count })
watch([selectedId, editing, createMode], () => {
  void nextTick(() => document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' }))
})
</script>

<template>
  <PurchaseWorkspaceView
    v-if="editing || createMode"
    :workspace="workspace"
    :currency-unit="props.currencyUnit"
    :suppliers="props.suppliers"
    :materials="props.materials"
    @back="backToPurchases"
  />

  <div v-else class="flex h-full min-h-0 min-w-0 flex-col overflow-hidden" aria-label="Purchases workspace">
    <WorkspaceStickyStack class="shrink-0" :flush="true">
      <WorkspaceHeader
        :show-breadcrumb="true"
        title="Purchases"
        description="Record supplier purchases, then post them into the immutable inventory ledger."
      >
        <SearchField v-model="searchQuery" class="w-full min-w-0 sm:w-64" placeholder="Search purchases…" aria-label="Search purchases" />
        <button class="btn btn-primary w-full gap-2 sm:w-auto" type="button" @click="startCreate"><Plus :size="16" aria-hidden="true" />New purchase</button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <div class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(22rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.15fr)_minmax(24rem,0.85fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="purchase-register flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100" aria-label="Purchase register">
        <div class="shrink-0 border-b border-base-300 p-3 sm:p-4">
          <div class="flex min-w-0 flex-wrap items-center justify-between gap-3">
            <div class="flex min-w-0 flex-wrap items-center gap-2">
              <button v-for="status in statusOptions" :key="status" class="inline-flex h-9 items-center gap-2 rounded-box border px-3 text-sm transition-colors" :class="purchaseFilter === status ? 'border-primary bg-primary/10 text-primary' : 'border-base-300 text-base-content/70 hover:border-primary/50 hover:text-base-content'" type="button" @click="purchaseFilter = status"><span class="size-2 rounded-full" :class="status === 'Posted' ? 'bg-success' : status === 'Archived' ? 'bg-base-content/35' : status === 'Draft' ? 'bg-warning' : 'bg-primary'"></span>{{ status }}<span class="rounded-full bg-base-200 px-1.5 py-0.5 text-xs tabular-nums">{{ statusCount(status) }}</span></button>
            </div>
            <span class="hidden h-6 w-px bg-base-300 sm:block" aria-hidden="true"></span>
            <div class="flex w-full min-w-0 flex-wrap items-center justify-end gap-2 sm:w-auto"><SelectField v-model="sortOrder" class="min-w-0 flex-1 sm:w-36 sm:flex-none" aria-label="Sort purchases" :options="[{ label: 'Sort by date', value: 'date' }, { label: 'Sort by total', value: 'total' }, { label: 'Sort by supplier', value: 'supplier' }]" /></div>
          </div>
        </div>

        <div class="purchase-register-table-head hidden grid-cols-[minmax(0,1.5fr)_minmax(7rem,0.8fr)_8rem_7rem_6rem_1.25rem] gap-3 border-b border-base-300 px-4 py-3 text-xs font-medium text-base-content/55"><span>Purchase</span><span>Supplier</span><span>Date</span><span>Total</span><span>Status</span><span></span></div>
        <LoadingState v-if="isLoading" label="Loading purchases…" />
        <div v-else-if="pagedRows.length" class="min-h-0 flex-1 overflow-y-auto divide-y divide-base-300">
          <button v-for="purchase in pagedRows" :key="purchase.id" class="purchase-register-row group grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200/60 focus-visible:bg-base-200/60 focus-visible:outline focus-visible:outline-1 focus-visible:outline-primary" :class="selectedId === purchase.id ? 'bg-primary/10' : ''" type="button" @click="select(purchase.id)">
            <span class="purchase-register-mobile-card">
              <span class="purchase-register-mobile-icon grid place-items-center rounded-box border border-base-300 bg-base-200 text-primary"><ShoppingCart :size="18" aria-hidden="true" /></span>
              <span class="purchase-register-mobile-identity min-w-0 self-center"><strong class="block truncate text-sm">{{ purchase.purchaseNumber }}</strong><span class="block truncate text-xs text-base-content/60">{{ purchase.supplierName || 'No supplier' }}<span v-if="purchase.supplierInvoiceNumber"> · {{ purchase.supplierInvoiceNumber }}</span></span></span>
              <StatusBadge class="purchase-register-mobile-status justify-self-end self-center" :label="purchase.status" :tone="statusTone(purchase.status)" />
              <ChevronRight :size="17" class="register-row-arrow purchase-register-mobile-arrow self-center text-base-content/45" aria-hidden="true" />
              <span class="purchase-register-mobile-summary flex min-w-0 flex-wrap gap-x-3 gap-y-1 text-xs text-base-content/55"><span>{{ formatDateTime(purchase.purchaseDate) }}</span><span>{{ purchase.items.length }} item{{ purchase.items.length === 1 ? '' : 's' }}</span><span>{{ formatMoney(purchase.totalRial, props.currencyUnit) }}</span></span>
            </span>
            <span class="purchase-register-name purchase-register-desktop-only flex min-w-0 items-center gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box border border-base-300 bg-base-200 text-primary"><ShoppingCart :size="18" aria-hidden="true" /></span><span class="min-w-0"><strong class="block truncate text-sm">{{ purchase.purchaseNumber }}</strong><span class="block truncate text-xs text-base-content/60">{{ purchase.supplierInvoiceNumber || 'No supplier invoice' }}</span></span></span>
            <span class="purchase-register-supplier purchase-register-desktop-only hidden truncate text-xs text-base-content/70">{{ purchase.supplierName || 'No supplier' }}</span>
            <span class="purchase-register-date purchase-register-desktop-only hidden truncate text-xs text-base-content/70">{{ formatDateTime(purchase.purchaseDate) }}</span>
            <span class="purchase-register-total purchase-register-desktop-only hidden text-sm tabular-nums text-base-content/80">{{ formatMoney(purchase.totalRial, props.currencyUnit) }}</span>
            <StatusBadge class="purchase-register-status purchase-register-desktop-only justify-self-start" :label="purchase.status" :tone="statusTone(purchase.status)" />
            <ChevronRight :size="17" class="register-row-arrow purchase-register-chevron purchase-register-desktop-only hidden text-base-content/45" aria-hidden="true" />
          </button>
        </div>
        <EmptyState v-else :title="rows.length ? 'No purchases match this view' : 'No purchases yet'" :description="rows.length ? 'Try another status or search term.' : 'Record the first supplier purchase to establish an inventory ledger.'"><template #icon><Search :size="22" aria-hidden="true" /></template><template #action><button v-if="rows.length" class="btn btn-outline btn-sm" type="button" @click="resetListFilters">Clear filters</button><button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="startCreate"><Plus :size="15" aria-hidden="true" />Record purchase</button></template></EmptyState>
        <footer class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-4 py-3 text-xs text-base-content/60"><span>{{ visibleRows.length ? `Showing ${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, visibleRows.length)} of ${visibleRows.length} purchases` : '0 purchases' }}</span><div class="flex items-center gap-1"><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === 1" aria-label="Previous page" @click="goToPage(page - 1)"><ChevronLeft :size="15" aria-hidden="true" /></button><button v-for="number in pageNumbers" :key="number" class="btn btn-xs min-w-8" :class="page === number ? 'btn-primary' : 'btn-ghost'" type="button" @click="goToPage(number)">{{ number }}</button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === pageCount" aria-label="Next page" @click="goToPage(page + 1)"><ChevronRight :size="15" aria-hidden="true" /></button></div></footer>
      </section>

      <PurchaseDetailPanel v-if="current" :workspace="workspace" :currency-unit="props.currencyUnit" />
      <section v-else class="flex min-h-72 min-w-0 items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><EmptyState title="Select a purchase" description="Choose a purchase from the register to inspect its items, payment, and posting status."><template #icon><ShoppingCart :size="22" aria-hidden="true" /></template></EmptyState></section>
    </div>
  </div>
</template>
