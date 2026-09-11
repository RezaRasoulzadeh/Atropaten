<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, Factory, Layers3, Plus, Search } from 'lucide-vue-next'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import SearchField from '../../components/ui/SearchField.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import MachineDetailPanel from './MachineDetailPanel.vue'
import MachineEditorWizard from './MachineEditorWizard.vue'
import { useMachinesWorkspace, type MachineFilter } from './useMachinesWorkspace'

const props = defineProps<{ currencyUnit: CurrencyUnit }>()
const emit = defineEmits<{ notify: [message: string] }>()
const workspace = useMachinesWorkspace(props, emit)
const { machines, selectedId, selectedMachine, machineFilter, searchQuery, editorMode, isLoading, filteredMachines, startCreate, selectMachine } = workspace

const page = ref(1)
const pageSize = 10
const statusOptions: MachineFilter[] = ['All', 'Active', 'Archived']

const visibleMachines = computed(() => filteredMachines.value)
const pageCount = computed(() => Math.max(1, Math.ceil(visibleMachines.value.length / pageSize)))
const pagedMachines = computed(() => visibleMachines.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const pageNumbers = computed(() => Array.from({ length: pageCount.value }, (_, index) => index + 1))

function statusCount(status: MachineFilter) {
  if (status === 'All') return machines.value.length
  return machines.value.filter((machine) => status === 'Active' ? machine.active : !machine.active).length
}

function profileSummary(machine: typeof machines.value[number]) {
  const count = machine.rates?.length || 1
  return `${count} profile${count === 1 ? '' : 's'}`
}

function goToPage(value: number) {
  page.value = Math.min(Math.max(value, 1), pageCount.value)
}

function clearFilters() {
  searchQuery.value = ''
  machineFilter.value = 'All'
  page.value = 1
}

watch([searchQuery, machineFilter], () => { page.value = 1 })
watch(pageCount, (count) => { if (page.value > count) page.value = count })
watch([selectedId, editorMode], () => { void nextTick(() => document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' })) })
</script>

<template>
  <div v-if="editorMode" class="h-full min-h-0 w-full min-w-0">
    <MachineEditorWizard :workspace="workspace" :currency-unit="props.currencyUnit" @cancel="workspace.cancelEditor" />
  </div>

  <div v-else class="flex h-full min-h-0 min-w-0 flex-col overflow-hidden" aria-label="Machines workspace">
    <WorkspaceStickyStack class="shrink-0" :flush="true">
      <WorkspaceHeader :show-breadcrumb="true" title="Machines" description="Keep reusable equipment rates ready for service cost definitions.">
        <SearchField v-model="searchQuery" class="w-full min-w-0 sm:w-64" placeholder="Search machines…" aria-label="Search machines" />
        <button class="btn btn-primary w-full gap-2 sm:w-auto" type="button" @click="startCreate"><Plus :size="16" aria-hidden="true" />Add machine</button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <div class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(22rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.15fr)_minmax(24rem,0.85fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="machine-register flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100" aria-label="Machine register">
        <div class="shrink-0 border-b border-base-300 p-3 sm:p-4"><div class="flex min-w-0 flex-wrap items-center gap-3"><div class="flex min-w-0 flex-wrap items-center gap-2"><button v-for="status in statusOptions" :key="status" class="inline-flex h-9 items-center gap-2 rounded-box border px-3 text-sm transition-colors" :class="machineFilter === status ? 'border-primary bg-primary/10 text-primary' : 'border-base-300 text-base-content/70 hover:border-primary/50 hover:text-base-content'" type="button" @click="machineFilter = status"><span class="size-2 rounded-full" :class="status === 'Active' ? 'bg-success' : status === 'Archived' ? 'bg-base-content/35' : 'bg-primary'"></span>{{ status }}<span class="rounded-full bg-base-200 px-1.5 py-0.5 text-xs tabular-nums">{{ statusCount(status) }}</span></button></div></div></div>

        <div class="machine-register-table-head hidden grid-cols-[minmax(0,1.5fr)_minmax(7rem,0.8fr)_8rem_8rem_6rem_1.25rem] gap-3 border-b border-base-300 px-4 py-3 text-xs font-medium text-base-content/55"><span>Name</span><span>Category</span><span>Rate</span><span>Basis</span><span>Status</span><span></span></div>
        <LoadingState v-if="isLoading" label="Loading machines…" />
        <div v-else-if="pagedMachines.length" class="min-h-0 flex-1 overflow-y-auto divide-y divide-base-300">
          <button v-for="machine in pagedMachines" :key="machine.id" class="machine-register-row group grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200/60 focus-visible:bg-base-200/60 focus-visible:outline focus-visible:outline-1 focus-visible:outline-primary" :class="selectedId === machine.id ? 'bg-primary/10' : ''" type="button" @click="selectMachine(machine.id)">
            <span class="machine-register-mobile-card"><span class="machine-register-mobile-icon grid place-items-center rounded-box border border-base-300 bg-base-200 bg-cover bg-center text-primary" :style="machine.imagePath ? { backgroundImage: `url('${machine.imagePath}')` } : undefined"><Factory v-if="!machine.imagePath" :size="18" aria-hidden="true" /></span><span class="machine-register-mobile-identity min-w-0 self-center"><strong class="block truncate text-sm">{{ machine.name }}</strong><span class="block truncate text-xs text-base-content/60">{{ machine.code || 'No code' }}</span></span><StatusBadge class="machine-register-mobile-status justify-self-end self-center" :label="machine.active ? 'Active' : 'Archived'" :tone="machine.active ? 'green' : 'slate'" /><span class="register-row-arrow machine-register-mobile-arrow self-center text-base-content/45"><ChevronRight :size="17" aria-hidden="true" /></span><span class="machine-register-mobile-summary flex min-w-0 flex-wrap gap-x-3 gap-y-1 text-xs text-base-content/55"><span>{{ machine.category || 'Uncategorized' }}</span><span>{{ formatMoney(machine.rateRial, props.currencyUnit) }}</span><span>{{ profileSummary(machine) }}</span></span></span>
            <span class="machine-register-desktop-only machine-register-name flex min-w-0 items-center gap-3"><span class="grid size-9 shrink-0 place-items-center overflow-hidden rounded-box border border-base-300 bg-base-200 bg-cover bg-center text-primary" :style="machine.imagePath ? { backgroundImage: `url('${machine.imagePath}')` } : undefined"><Factory v-if="!machine.imagePath" :size="18" aria-hidden="true" /></span><span class="min-w-0"><strong class="block truncate text-sm">{{ machine.name }}</strong><span class="block truncate text-xs text-base-content/60">{{ machine.code || 'No code' }}</span></span></span>
            <span class="machine-register-desktop-only machine-register-category hidden truncate text-xs text-base-content/70">{{ machine.category || 'Uncategorized' }}</span>
            <span class="machine-register-desktop-only machine-register-rate hidden text-sm tabular-nums text-base-content/80">{{ formatMoney(machine.rateRial, props.currencyUnit) }}</span>
            <span class="machine-register-desktop-only machine-register-basis hidden truncate text-xs text-base-content/70">{{ workspace.basisLabel(machine.rateBasis) }}</span>
            <StatusBadge class="machine-register-desktop-only machine-register-status justify-self-end" :label="machine.active ? 'Active' : 'Archived'" :tone="machine.active ? 'green' : 'slate'" />
            <ChevronRight class="register-row-arrow machine-register-desktop-only machine-register-chevron hidden text-base-content/45" :size="17" aria-hidden="true" />
          </button>
        </div>
        <EmptyState v-else :title="machines.length ? 'No machines match this view' : 'No machines yet'" :description="machines.length ? 'Try another filter or search term.' : 'Add the first reusable rate input for production.'"><template #icon><Search :size="22" aria-hidden="true" /></template><template #action><button v-if="machines.length" class="btn btn-outline btn-sm" type="button" @click="clearFilters">Clear filters</button><button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="startCreate"><Plus :size="15" aria-hidden="true" />Create machine</button></template></EmptyState>
        <footer class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-4 py-3 text-xs text-base-content/60"><span>{{ visibleMachines.length ? `Showing ${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, visibleMachines.length)} of ${visibleMachines.length} machines` : '0 machines' }}</span><div class="flex items-center gap-1"><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === 1" aria-label="Previous page" @click="goToPage(page - 1)"><ChevronLeft :size="15" aria-hidden="true" /></button><button v-for="number in pageNumbers" :key="number" class="btn btn-xs min-w-8" :class="page === number ? 'btn-primary' : 'btn-ghost'" type="button" @click="goToPage(number)">{{ number }}</button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === pageCount" aria-label="Next page" @click="goToPage(page + 1)"><ChevronRight :size="15" aria-hidden="true" /></button></div></footer>
      </section>

      <MachineDetailPanel v-if="selectedMachine" :workspace="workspace" :currency-unit="props.currencyUnit" @edit="workspace.startEdit" @archive="workspace.setActive(false)" @reactivate="workspace.setActive(true)" @remove="workspace.remove" />
      <section v-else class="flex min-h-72 min-w-0 items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><EmptyState title="Select a machine" description="Choose a machine from the register to inspect rates and details."><template #icon><Layers3 :size="22" aria-hidden="true" /></template><template #action><button class="btn btn-primary btn-sm gap-2" type="button" @click="startCreate"><Plus :size="15" aria-hidden="true" />Create machine</button></template></EmptyState></section>
    </div>
  </div>
</template>
