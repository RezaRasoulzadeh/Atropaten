<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, Package, Plus, Search } from 'lucide-vue-next'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import SearchField from '../../components/ui/SearchField.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import MaterialEditorWizard from './MaterialEditorWizard.vue'
import MaterialDetailPanel from './MaterialDetailPanel.vue'
import { useMaterialsWorkspace, type MaterialFilter } from './useMaterialsWorkspace'

const props = defineProps<{ currencyUnit: CurrencyUnit }>()
const emit = defineEmits<{ notify: [message: string] }>()
const workspace = useMaterialsWorkspace(props, emit)
const { materials, selectedId, selectedMaterial, searchQuery, materialFilter, editorMode, isLoading, filteredMaterials, startCreate, selectMaterial } = workspace

const page = ref(1)
const pageSize = 10
const statusOptions: MaterialFilter[] = ['All', 'Active', 'Archived']

const visibleMaterials = computed(() => filteredMaterials.value)
const pageCount = computed(() => Math.max(1, Math.ceil(visibleMaterials.value.length / pageSize)))
const pagedMaterials = computed(() => visibleMaterials.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const pageNumbers = computed(() => Array.from({ length: pageCount.value }, (_, index) => index + 1))

function statusCount(status: MaterialFilter) {
  if (status === 'All') return materials.value.length
  return materials.value.filter((material) => status === 'Active' ? material.active : !material.active).length
}
function unitCost(material: typeof materials.value[number]) {
  return material.highestPurchaseUnitCostRial || material.averageUnitCostRial
}
function goToPage(value: number) { page.value = Math.min(Math.max(value, 1), pageCount.value) }
function clearFilters() {
  searchQuery.value = ''
  materialFilter.value = 'All'
  page.value = 1
}

watch([searchQuery, materialFilter], () => { page.value = 1 })
watch(pageCount, (count) => { if (page.value > count) page.value = count })
watch([selectedId, editorMode], () => { void nextTick(() => document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' })) })
</script>

<template>
  <div v-if="editorMode" class="h-full min-h-0 w-full min-w-0">
    <MaterialEditorWizard :workspace="workspace" :currency-unit="props.currencyUnit" @cancel="workspace.cancelEditor" />
  </div>

  <div v-else class="flex h-full min-h-0 min-w-0 flex-col overflow-hidden" aria-label="Materials workspace">
    <WorkspaceStickyStack class="shrink-0" :flush="true">
      <WorkspaceHeader :show-breadcrumb="true" title="Materials" description="Keep physical stock, conversion units, and cost basis ready for production.">
        <SearchField v-model="searchQuery" class="w-full min-w-0 sm:w-64" placeholder="Search materials…" aria-label="Search materials" />
        <button class="btn btn-primary w-full gap-2 sm:w-auto" type="button" @click="startCreate"><Plus :size="16" aria-hidden="true" />Add material</button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <div class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(22rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.15fr)_minmax(24rem,0.85fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="material-register flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100" aria-label="Material register">
        <div class="shrink-0 border-b border-base-300 p-3 sm:p-4">
          <div class="flex min-w-0 flex-wrap items-center justify-between gap-3">
            <div class="flex min-w-0 flex-wrap items-center gap-2"><button v-for="status in statusOptions" :key="status" class="inline-flex h-9 items-center gap-2 rounded-box border px-3 text-sm transition-colors" :class="materialFilter === status ? 'border-primary bg-primary/10 text-primary' : 'border-base-300 text-base-content/70 hover:border-primary/50 hover:text-base-content'" type="button" @click="materialFilter = status"><span class="size-2 rounded-full" :class="status === 'Active' ? 'bg-success' : status === 'Archived' ? 'bg-base-content/35' : 'bg-primary'"></span>{{ status }}<span class="rounded-full bg-base-200 px-1.5 py-0.5 text-xs tabular-nums">{{ statusCount(status) }}</span></button></div>
          </div>
        </div>

        <div class="material-register-table-head hidden grid-cols-[minmax(0,1.5fr)_minmax(7rem,0.8fr)_8rem_7rem_6rem_1.25rem] gap-3 border-b border-base-300 px-4 py-3 text-xs font-medium text-base-content/55"><span>Name</span><span>Category</span><span>Unit cost</span><span>Stock</span><span>Status</span><span></span></div>
        <LoadingState v-if="isLoading" label="Loading materials…" />
        <div v-else-if="pagedMaterials.length" class="min-h-0 flex-1 overflow-y-auto divide-y divide-base-300">
          <button v-for="material in pagedMaterials" :key="material.id" class="material-register-row group grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200/60 focus-visible:bg-base-200/60 focus-visible:outline focus-visible:outline-1 focus-visible:outline-primary" :class="selectedId === material.id ? 'bg-primary/10' : ''" type="button" @click="selectMaterial(material.id)">
            <span class="material-register-mobile-card">
              <span class="material-register-mobile-icon grid place-items-center rounded-box border border-base-300 bg-base-200 text-primary"><Package :size="18" aria-hidden="true" /></span>
              <span class="material-register-mobile-identity min-w-0 self-center"><strong class="block truncate text-sm">{{ material.name }}</strong><span class="block truncate text-xs text-base-content/60">{{ material.sku || 'No SKU' }}<span v-if="material.consumptionUnit"> · {{ material.consumptionUnit }}</span></span></span>
              <StatusBadge class="material-register-mobile-status justify-self-end self-center" :label="material.active ? 'Active' : 'Archived'" :tone="material.active ? 'green' : 'slate'" />
              <ChevronRight :size="17" class="register-row-arrow material-register-mobile-arrow self-center text-base-content/45" aria-hidden="true" />
              <span class="material-register-mobile-summary flex min-w-0 flex-wrap gap-x-3 gap-y-1 text-xs text-base-content/55"><span>{{ material.category || 'Uncategorized' }}</span><span>{{ formatMoney(unitCost(material), props.currencyUnit) }} / {{ material.consumptionUnit }}</span><span>{{ material.availableStock }} available</span></span>
            </span>
            <span class="material-register-name material-register-desktop-only flex min-w-0 items-center gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box border border-base-300 bg-base-200 text-primary"><Package :size="18" aria-hidden="true" /></span><span class="min-w-0"><strong class="block truncate text-sm">{{ material.name }}</strong><span class="block truncate text-xs text-base-content/60">{{ material.sku || 'No SKU' }}<span v-if="material.consumptionUnit"> · {{ material.consumptionUnit }}</span></span></span></span>
            <span class="material-register-category material-register-desktop-only hidden truncate text-xs text-base-content/70">{{ material.category || 'Uncategorized' }}</span>
            <span class="material-register-cost material-register-desktop-only hidden text-sm tabular-nums text-base-content/80">{{ formatMoney(unitCost(material), props.currencyUnit) }}</span>
            <span class="material-register-stock material-register-desktop-only hidden truncate text-xs text-base-content/70">{{ material.availableStock }} {{ material.consumptionUnit }}</span>
            <StatusBadge class="material-register-status material-register-desktop-only justify-self-end" :label="material.active ? 'Active' : 'Archived'" :tone="material.active ? 'green' : 'slate'" />
            <ChevronRight :size="17" class="register-row-arrow material-register-chevron material-register-desktop-only hidden text-base-content/45" aria-hidden="true" />
          </button>
        </div>
        <EmptyState v-else :title="materials.length ? 'No materials match this view' : 'No materials yet'" :description="materials.length ? 'Try another filter or search term.' : 'Create the first material to establish a production stock baseline.'"><template #icon><Search :size="22" aria-hidden="true" /></template><template #action><button v-if="materials.length" class="btn btn-outline btn-sm" type="button" @click="clearFilters">Clear filters</button><button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="startCreate"><Plus :size="15" aria-hidden="true" />Create material</button></template></EmptyState>
        <footer class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-4 py-3 text-xs text-base-content/60"><span>{{ visibleMaterials.length ? `Showing ${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, visibleMaterials.length)} of ${visibleMaterials.length} materials` : '0 materials' }}</span><div class="flex items-center gap-1"><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === 1" aria-label="Previous page" @click="goToPage(page - 1)"><ChevronLeft :size="15" aria-hidden="true" /></button><button v-for="number in pageNumbers" :key="number" class="btn btn-xs min-w-8" :class="page === number ? 'btn-primary' : 'btn-ghost'" type="button" @click="goToPage(number)">{{ number }}</button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === pageCount" aria-label="Next page" @click="goToPage(page + 1)"><ChevronRight :size="15" aria-hidden="true" /></button></div></footer>
      </section>

      <MaterialDetailPanel v-if="selectedMaterial" :workspace="workspace" :currency-unit="props.currencyUnit" />
      <section v-else class="flex min-h-72 min-w-0 items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><EmptyState title="Select a material" description="Choose a material from the register to inspect stock and cost details."><template #icon><Package :size="22" aria-hidden="true" /></template></EmptyState></section>
    </div>
  </div>
</template>
