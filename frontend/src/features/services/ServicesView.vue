<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, Layers3, Plus, Search } from 'lucide-vue-next'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import SearchField from '../../components/ui/SearchField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import type { ServiceFilter } from './types'
import ServiceEditorWizard from './ServiceEditorWizard.vue'
import ServiceDetailPanel from './ServiceDetailPanel.vue'
import { serviceEstimatedSellingPrice } from './serviceDefaultPricing'
import { useServicesWorkspace } from './useServicesWorkspace'

const emit = defineEmits<{ notify: [message: string] }>()
const props = defineProps<{ currencyUnit: CurrencyUnit }>()

const {
  busy,
  services,
  materials,
  machines,
  selectedId,
  searchQuery,
  serviceFilter,
  editorMode,
  form,
  isLoading,
  isSaving,
  validationAttempted,
  selectedService,
  filteredServices,
  startCreate,
  startEdit,
  selectService,
  cancelEditor,
  saveService,
  setActive,
  remove,
} = useServicesWorkspace(props, emit)

const sortOrder = ref('name')
const page = ref(1)
const pageSize = 10
const statusOptions: ServiceFilter[] = ['All', 'Active', 'Archived']

const visibleServices = computed(() => {
  const items = filteredServices.value
  return [...items].sort((left, right) => {
    if (sortOrder.value === 'updated') return String(right.updatedAt).localeCompare(String(left.updatedAt))
    if (sortOrder.value === 'category') return `${left.category}${left.name}`.localeCompare(`${right.category}${right.name}`)
    return left.name.localeCompare(right.name)
  })
})

const pageCount = computed(() => Math.max(1, Math.ceil(visibleServices.value.length / pageSize)))
const pagedServices = computed(() => visibleServices.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const pageNumbers = computed(() => Array.from({ length: pageCount.value }, (_, index) => index + 1))

function statusCount(status: ServiceFilter) {
  if (status === 'All') return services.value.length
  return services.value.filter((service) => status === 'Active' ? service.active : !service.active).length
}

function goToPage(value: number) {
  page.value = Math.min(Math.max(value, 1), pageCount.value)
}

function estimatedPrice(service: typeof services.value[number]) {
  return serviceEstimatedSellingPrice(service, materials.value, machines.value, services.value)
}

function resetListFilters() {
  searchQuery.value = ''
  serviceFilter.value = 'All' as ServiceFilter
  sortOrder.value = 'name'
  page.value = 1
}

watch([searchQuery, serviceFilter, sortOrder], () => { page.value = 1 })
watch(pageCount, (count) => { if (page.value > count) page.value = count })

watch(
  [() => selectedId.value, () => editorMode.value],
  () => {
    void nextTick(() => document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' }))
  },
)

</script>

<template>
  <ServiceEditorWizard
    v-if="editorMode"
    :form="form"
    :editor-mode="editorMode"
    :busy="busy"
    :is-saving="isSaving"
    :validation-attempted="validationAttempted"
    :active="editorMode === 'edit' ? (selectedService?.active ?? true) : true"
    :materials="materials"
    :machines="machines"
    :services="services"
    :service-id="selectedService?.id || ''"
    :currency-unit="props.currencyUnit"
    @cancel="cancelEditor"
    @save="saveService"
  />

  <div v-else class="flex h-full min-h-0 min-w-0 flex-col overflow-hidden" aria-label="Services workspace">
    <WorkspaceStickyStack class="shrink-0" :flush="true">
      <WorkspaceHeader
        :show-breadcrumb="true"
        title="Services"
        eyebrow="Catalog / sellable operations"
        description="Manage reusable printing services, customer parameters, and pricing rules."
      >
        <SearchField v-model="searchQuery" class="w-full min-w-0 sm:w-64" placeholder="Search services…" aria-label="Search services" />
        <button class="btn btn-primary w-full gap-2 sm:w-auto" type="button" @click="startCreate"><Plus :size="16" aria-hidden="true" />Add service</button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <div class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(22rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.15fr)_minmax(24rem,0.85fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100" aria-label="Service register">
        <div class="shrink-0 border-b border-base-300 p-3 sm:p-4">
          <div class="flex min-w-0 flex-wrap items-center justify-between gap-3">
            <div class="flex min-w-0 flex-wrap items-center gap-2">
              <button v-for="status in statusOptions" :key="status" class="inline-flex h-9 items-center gap-2 rounded-box border px-3 text-sm transition-colors" :class="serviceFilter === status ? 'border-primary bg-primary/10 text-primary' : 'border-base-300 text-base-content/70 hover:border-primary/50 hover:text-base-content'" type="button" @click="serviceFilter = status"><span class="size-2 rounded-full" :class="status === 'Active' ? 'bg-success' : status === 'Archived' ? 'bg-base-content/35' : 'bg-primary'"></span>{{ status }}<span class="rounded-full bg-base-200 px-1.5 py-0.5 text-xs tabular-nums">{{ statusCount(status) }}</span></button>
            </div>
            <span class="hidden h-6 w-px bg-base-300 sm:block" aria-hidden="true"></span>
            <div class="flex w-full min-w-0 flex-wrap items-center justify-end gap-2 sm:w-auto"><SelectField v-model="sortOrder" class="min-w-0 flex-1 sm:w-36 sm:flex-none" aria-label="Sort services" :options="[{ label: 'Sort by name', value: 'name' }, { label: 'Sort by category', value: 'category' }, { label: 'Recently updated', value: 'updated' }]" /></div>
          </div>
        </div>

        <div class="hidden grid-cols-[minmax(0,1.5fr)_minmax(7rem,0.8fr)_7.5rem_6rem_1.25rem] gap-3 border-b border-base-300 px-4 py-3 text-xs font-medium text-base-content/55 md:grid"><span>Name</span><span>Category</span><span>Last price</span><span>Status</span><span></span></div>
        <LoadingState v-if="isLoading" label="Loading services…" />
        <div v-else-if="pagedServices.length" class="min-h-0 flex-1 overflow-y-auto divide-y divide-base-300">
          <button v-for="service in pagedServices" :key="service.id" class="group grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200/60 focus-visible:bg-base-200/60 focus-visible:outline focus-visible:outline-1 focus-visible:outline-primary md:grid-cols-[minmax(0,1.5fr)_minmax(7rem,0.8fr)_7.5rem_6rem_1.25rem]" :class="selectedId === service.id ? 'bg-primary/10' : ''" type="button" @click="selectService(service.id)">
            <span class="flex min-w-0 items-center gap-3"><span class="size-9 shrink-0 overflow-hidden rounded-box border border-base-300 bg-base-200 bg-cover bg-center" :style="service.imagePath ? { backgroundImage: `url('${service.imagePath}')` } : undefined"><Layers3 v-if="!service.imagePath" :size="18" class="m-2 text-base-content/50" aria-hidden="true" /></span><span class="min-w-0"><strong class="block truncate text-sm">{{ service.name }}</strong><span class="block truncate text-xs text-base-content/60">{{ service.code || 'No code' }}<span v-if="service.description"> · {{ service.description }}</span></span></span></span>
            <span class="hidden truncate text-xs text-base-content/70 md:block">{{ service.category || 'Uncategorized' }}</span>
            <span class="hidden text-sm tabular-nums text-base-content/80 md:block">{{ estimatedPrice(service) !== null ? formatMoney(estimatedPrice(service) ?? 0, props.currencyUnit) : 'Needs setup' }}</span>
            <StatusBadge class="justify-self-end md:justify-self-start" :label="service.active ? 'Active' : 'Archived'" :tone="service.active ? 'green' : 'slate'" />
            <ChevronRight :size="17" class="register-row-arrow justify-self-end text-base-content/45" aria-hidden="true" />
            <span class="col-span-3 flex flex-wrap gap-x-3 gap-y-1 text-xs text-base-content/55 md:hidden"><span>{{ service.category || 'Uncategorized' }}</span><span>{{ estimatedPrice(service) !== null ? formatMoney(estimatedPrice(service) ?? 0, props.currencyUnit) : 'Needs setup' }}</span></span>
          </button>
        </div>
        <EmptyState v-else :title="services.length ? 'No services match this view' : 'No services yet'" :description="services.length ? 'Try another filter or search term.' : 'Create the first reusable operation for your shop.'">
          <template #icon><Search :size="22" aria-hidden="true" /></template>
          <template #action><button v-if="services.length" class="btn btn-outline btn-sm" type="button" @click="resetListFilters">Clear filters</button><button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="startCreate"><Plus :size="15" aria-hidden="true" />Create service</button></template>
        </EmptyState>
        <footer class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-4 py-3 text-xs text-base-content/60"><span>{{ visibleServices.length ? `Showing ${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, visibleServices.length)} of ${visibleServices.length} services` : '0 services' }}</span><div class="flex items-center gap-1"><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === 1" aria-label="Previous page" @click="goToPage(page - 1)"><ChevronLeft :size="15" aria-hidden="true" /></button><button v-for="number in pageNumbers" :key="number" class="btn btn-xs min-w-8" :class="page === number ? 'btn-primary' : 'btn-ghost'" type="button" @click="goToPage(number)">{{ number }}</button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === pageCount" aria-label="Next page" @click="goToPage(page + 1)"><ChevronRight :size="15" aria-hidden="true" /></button></div></footer>
      </section>

      <ServiceDetailPanel
        v-if="selectedService"
        :service="selectedService"
        :materials="materials"
        :machines="machines"
        :services="services"
        :currency-unit="props.currencyUnit"
        :busy="busy"
        @edit="startEdit"
        @archive="setActive(false)"
        @reactivate="setActive(true)"
        @remove="remove"
      />
      <section v-else class="flex min-h-72 min-w-0 items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><EmptyState title="Select a service" description="Choose a service from the register to inspect its setup and pricing."><template #icon><Layers3 :size="22" aria-hidden="true" /></template></EmptyState></section>
    </div>
  </div>
</template>
