<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, Factory, Plus, Search } from 'lucide-vue-next'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import SearchField from '../../components/ui/SearchField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { OrderRecord } from '../../api/orders'
import type { CurrencyUnit } from '../../utils/currency'
import { confirmAction } from '../../ui/feedback'
import { useProductionWorkspace } from './useProductionWorkspace'
import ProductionJobEditorWizard from './ProductionJobEditorWizard.vue'
import ProductionJobWorkspaceView from './ProductionJobWorkspaceView.vue'

const props = defineProps<{
  currencyUnit: CurrencyUnit
  orders: OrderRecord[]
  materials: any[]
  machines: any[]
  suppliers: any[]
}>()
const emit = defineEmits<{ notify: [message: string] }>()
const workspace = useProductionWorkspace(props, emit)
const {
  jobs,
  selectedId,
  selected,
  statusFilter,
  searchQuery,
  loading,
  createMode,
  editing,
  visibleJobs,
  beginCreate,
  select,
  jobOrder,
  jobContext,
  statusTone,
} = workspace

const sortOrder = ref('updated')
const page = ref(1)
const pageSize = 10
const statusOptions = ['All', 'Pending', 'Ready', 'In Progress', 'Paused', 'Completed', 'Cancelled', 'Failed'] as const

const sortedJobs = computed(() => [...visibleJobs.value].sort((left, right) => {
  if (sortOrder.value === 'status') return `${left.status}${left.jobNumber}`.localeCompare(`${right.status}${right.jobNumber}`)
  if (sortOrder.value === 'job') return left.jobNumber.localeCompare(right.jobNumber)
  return String(right.createdAt).localeCompare(String(left.createdAt))
}))
const pageCount = computed(() => Math.max(1, Math.ceil(sortedJobs.value.length / pageSize)))
const pagedJobs = computed(() => sortedJobs.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const pageNumbers = computed(() => Array.from({ length: pageCount.value }, (_, index) => index + 1))

watch([searchQuery, statusFilter, sortOrder], () => { page.value = 1 })
watch(pageCount, (count) => { if (page.value > count) page.value = count })
watch([selectedId, createMode, editing], () => {
  void nextTick(() => document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' }))
})

function statusCount(status: string) {
  if (status === 'All') return jobs.value.length
  return jobs.value.filter((job) => job.status === status).length
}

function goToPage(value: number) {
  page.value = Math.min(Math.max(value, 1), pageCount.value)
}

function resetListFilters() {
  searchQuery.value = ''
  statusFilter.value = 'All'
  sortOrder.value = 'updated'
  page.value = 1
}

async function cancelEditor() {
  if (workspace.busy.value) return
  const wasEditing = editing.value
  if (createMode.value && (workspace.form.value.orderId || workspace.form.value.orderItemId || workspace.form.value.quantity || workspace.form.value.notes)) {
    const discard = await confirmAction({ title: 'Discard new production job?', message: 'The new job has not been saved.', confirmLabel: 'Discard draft', danger: true })
    if (!discard) return
  }
  if (editing.value) {
    const discard = await confirmAction({ title: 'Discard job changes?', message: 'Your unsaved production job changes will be lost.', confirmLabel: 'Discard changes', danger: true })
    if (!discard) return
    workspace.cancelEdit()
  }
  createMode.value = false
  if (!wasEditing) selectedId.value = null
}
</script>

<template>
  <div v-if="!selected && !createMode && !editing" class="flex h-full min-h-0 min-w-0 flex-col overflow-hidden" aria-label="Production workspace">
    <WorkspaceStickyStack class="shrink-0" :flush="true">
      <WorkspaceHeader :show-breadcrumb="true" title="Production" description="Schedule jobs, monitor progress, and keep material usage tied to confirmed orders.">
        <SearchField v-model="searchQuery" class="w-full min-w-0 sm:w-72" placeholder="Search production jobs…" aria-label="Search production jobs" />
        <button class="btn btn-primary w-full gap-2 sm:w-auto" type="button" @click="beginCreate"><Plus :size="16" aria-hidden="true" />Add production job</button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <div class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(22rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.15fr)_minmax(24rem,0.85fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100" aria-label="Production register">
        <div class="shrink-0 border-b border-base-300 p-3 sm:p-4"><div class="flex min-w-0 flex-wrap items-center justify-between gap-3"><div class="flex min-w-0 flex-wrap items-center gap-2"><button v-for="status in statusOptions" :key="status" class="inline-flex h-9 items-center gap-2 rounded-box border px-3 text-sm transition-colors" :class="statusFilter === status ? 'border-primary bg-primary/10 text-primary' : 'border-base-300 text-base-content/70 hover:border-primary/50 hover:text-base-content'" type="button" @click="statusFilter = status"><span class="size-2 rounded-full" :class="status === 'Completed' ? 'bg-success' : status === 'Cancelled' || status === 'Failed' ? 'bg-error' : status === 'All' ? 'bg-primary' : 'bg-info'"></span>{{ status }}<span class="rounded-full bg-base-200 px-1.5 py-0.5 text-xs tabular-nums">{{ statusCount(status) }}</span></button></div><span class="hidden h-6 w-px bg-base-300 sm:block" aria-hidden="true"></span><div class="flex w-full min-w-0 flex-wrap items-center justify-end gap-2 sm:w-auto"><SelectField v-model="sortOrder" class="min-w-0 flex-1 sm:w-40 sm:flex-none" aria-label="Sort production jobs" :options="[{ label: 'Recently created', value: 'updated' }, { label: 'Sort by job number', value: 'job' }, { label: 'Sort by status', value: 'status' }]" /></div></div></div>
        <div class="production-register-table-head hidden gap-3 border-b border-base-300 px-4 py-3 text-xs font-medium text-base-content/55 md:grid"><span>Job</span><span>Service / order</span><span>Schedule</span><span>Status</span><span></span></div>
        <LoadingState v-if="loading" label="Loading production jobs…" />
        <div v-else-if="pagedJobs.length" class="min-h-0 flex-1 overflow-y-auto divide-y divide-base-300"><button v-for="job in pagedJobs" :key="job.id" class="production-register-row group grid w-full min-w-0 items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200/60 focus-visible:bg-base-200/60 focus-visible:outline focus-visible:outline-1 focus-visible:outline-primary" :class="selectedId === job.id ? 'bg-primary/10' : ''" type="button" @click="select(job.id)"><span class="flex min-w-0 items-center gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box border border-base-300 bg-base-200 text-primary"><Factory :size="18" aria-hidden="true" /></span><span class="min-w-0"><strong class="block truncate text-sm">{{ job.jobNumber }}</strong><span class="block truncate text-xs text-base-content/60">{{ props.orders.find((order) => order.id === job.orderId)?.customerName || 'Walk-in customer' }}</span></span></span><span class="hidden min-w-0 truncate text-xs text-base-content/70 md:block">{{ job.serviceName }} · {{ jobOrder(job) }}</span><span class="hidden min-w-0 truncate text-xs text-base-content/70 md:block">{{ jobContext(job) }}</span><StatusBadge class="justify-self-end md:justify-self-start" :label="job.status" :tone="statusTone(job.status)" /><ChevronRight :size="17" class="register-row-arrow justify-self-end text-base-content/45" aria-hidden="true" /><span class="col-span-3 flex flex-wrap gap-x-3 gap-y-1 text-xs text-base-content/55 md:hidden"><span>{{ job.serviceName }}</span><span>{{ job.quantity }} {{ job.quantityUnit }}</span><span>{{ jobContext(job) }}</span></span></button></div>
        <EmptyState v-else :title="jobs.length ? 'No jobs match this view' : 'Production queue is empty'" :description="jobs.length ? 'Try another status filter or search term.' : 'Create a job from a confirmed order to start tracking production.'"><template #icon><Search :size="22" aria-hidden="true" /></template><template #action><button v-if="jobs.length" class="btn btn-outline btn-sm" type="button" @click="resetListFilters">Clear filters</button><button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="beginCreate"><Plus :size="15" aria-hidden="true" />Create job</button></template></EmptyState>
        <footer class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-4 py-3 text-xs text-base-content/60"><span>{{ sortedJobs.length ? `Showing ${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, sortedJobs.length)} of ${sortedJobs.length} jobs` : '0 jobs' }}</span><div class="flex items-center gap-1"><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === 1" aria-label="Previous page" @click="goToPage(page - 1)"><ChevronLeft :size="15" aria-hidden="true" /></button><button v-for="number in pageNumbers" :key="number" class="btn btn-xs min-w-8" :class="page === number ? 'btn-primary' : 'btn-ghost'" type="button" @click="goToPage(number)">{{ number }}</button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === pageCount" aria-label="Next page" @click="goToPage(page + 1)"><ChevronRight :size="15" aria-hidden="true" /></button></div></footer>
      </section>
      <section class="flex min-h-72 min-w-0 items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><EmptyState title="Open a production job" description="Select a job from the queue to manage workflow, reservations, consumption, and outsourcing."><template #icon><Factory :size="22" aria-hidden="true" /></template></EmptyState></section>
    </div>
  </div>

  <div v-else-if="createMode || editing" class="h-full min-h-0 w-full min-w-0">
    <ProductionJobEditorWizard :workspace="workspace" :currency-unit="props.currencyUnit" :orders="props.orders" :machines="props.machines" @cancel="cancelEditor" />
  </div>
  <ProductionJobWorkspaceView v-else v-bind="props" :workspace="workspace" @back="selectedId = null" />
</template>

<style scoped>
.production-register-table-head,
.production-register-row {
  grid-template-columns: minmax(0, 1.4fr) minmax(10rem, 1fr) minmax(10rem, 1fr) 7rem 1.25rem;
}

@media (max-width: 767px) {
  .production-register-row {
    grid-template-columns: minmax(0, 1fr) auto auto;
  }
}
</style>
