<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { Archive, ChevronLeft, ChevronRight, Edit3, FileText, MapPin, Mail, Phone, Plus, RotateCcw, Save, Search, Trash2, Truck } from 'lucide-vue-next'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import SearchField from '../../components/ui/SearchField.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { suppliersApi, type SupplierPayload, type SupplierRecord } from '../../api/suppliers'
import { formatDateTime } from '../../utils/date'
import { confirmAction } from '../../ui/feedback'

const emit = defineEmits<{ notify: [message: string] }>()
const { busy, runAction } = useWorkspaceActions()

type SupplierFilter = 'All' | 'Active' | 'Archived'

const suppliers = ref<SupplierRecord[]>([])
const selectedId = ref<string | null>(null)
const searchQuery = ref('')
const supplierFilter = ref<SupplierFilter>('All')
const page = ref(1)
const pageSize = 10
const isLoading = ref(false)
const editing = ref(false)
const form = ref<SupplierPayload>(emptyForm())
const statusOptions: SupplierFilter[] = ['All', 'Active', 'Archived']

const selectedSupplier = computed(() => suppliers.value.find((supplier) => supplier.id === selectedId.value) ?? null)
const filteredSuppliers = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return suppliers.value.filter((supplier) => {
    const matchesFilter = supplierFilter.value === 'All' || (supplierFilter.value === 'Active' ? supplier.active : !supplier.active)
    const matchesSearch = !query || [supplier.name, supplier.code, supplier.phone, supplier.email, supplier.address, supplier.notes]
      .some((value) => value.toLowerCase().includes(query))
    return matchesFilter && matchesSearch
  })
})
const visibleSuppliers = computed(() => filteredSuppliers.value)
const pageCount = computed(() => Math.max(1, Math.ceil(visibleSuppliers.value.length / pageSize)))
const pagedSuppliers = computed(() => visibleSuppliers.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const pageNumbers = computed(() => Array.from({ length: pageCount.value }, (_, index) => index + 1))

onMounted(loadSuppliers)

watch([searchQuery, supplierFilter], () => { page.value = 1 })
watch(pageCount, (count) => { if (page.value > count) page.value = count })
watch([selectedId, editing], () => {
  void nextTick(() => document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' }))
})

function emptyForm(): SupplierPayload {
  return { name: '', code: '', phone: '', email: '', address: '', notes: '' }
}

function statusCount(status: SupplierFilter) {
  if (status === 'All') return suppliers.value.length
  return suppliers.value.filter((supplier) => status === 'Active' ? supplier.active : !supplier.active).length
}

async function loadSuppliers() {
  isLoading.value = true
  try {
    const data = await suppliersApi.list(true)
    suppliers.value = data
    if (!selectedId.value) selectedId.value = data.find((supplier) => supplier.active)?.id ?? data[0]?.id ?? null
  } catch (error) {
    reportError(error)
  } finally {
    isLoading.value = false
  }
}

function selectSupplier(id: string) {
  selectedId.value = id
  editing.value = false
}

function startCreate() {
  selectedId.value = null
  form.value = emptyForm()
  editing.value = true
}

function startEdit() {
  const supplier = selectedSupplier.value
  if (!supplier) return
  form.value = {
    name: supplier.name,
    code: supplier.code,
    phone: supplier.phone,
    email: supplier.email,
    address: supplier.address,
    notes: supplier.notes,
  }
  editing.value = true
}

function resetListFilters() {
  searchQuery.value = ''
  supplierFilter.value = 'All'
  page.value = 1
}

function goToPage(value: number) {
  page.value = Math.min(Math.max(value, 1), pageCount.value)
}

async function save() {
  return runAction(async () => {
    try {
      const supplier = selectedId.value
        ? await suppliersApi.update(selectedId.value, form.value)
        : await suppliersApi.create(form.value)
      suppliers.value = suppliers.value.some((item) => item.id === supplier.id)
        ? suppliers.value.map((item) => item.id === supplier.id ? supplier : item)
        : [supplier, ...suppliers.value]
      selectedId.value = supplier.id
      editing.value = false
      emit('notify', 'Supplier saved.')
    } catch (error) {
      reportError(error)
    }
  })
}

async function setActive(active: boolean) {
  return runAction(async () => {
    if (!selectedSupplier.value) return
    try {
      const supplier = active
        ? await suppliersApi.reactivate(selectedSupplier.value.id)
        : await suppliersApi.archive(selectedSupplier.value.id)
      suppliers.value = suppliers.value.map((item) => item.id === supplier.id ? supplier : item)
      emit('notify', active ? 'Supplier reactivated.' : 'Supplier archived.')
    } catch (error) {
      reportError(error)
    }
  })
}

async function remove() {
  return runAction(async () => {
    if (!selectedSupplier.value || !(await confirmAction({
      title: 'Delete supplier',
      message: 'Delete this supplier permanently?',
      confirmLabel: 'Delete supplier',
      danger: true,
    }))) return
    try {
      const deletedId = selectedSupplier.value.id
      await suppliersApi.remove(deletedId)
      suppliers.value = suppliers.value.filter((item) => item.id !== deletedId)
      selectedId.value = null
      emit('notify', 'Supplier deleted.')
    } catch (error) {
      reportError(error)
    }
  })
}
</script>

<template>
  <div v-if="editing" class="flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Supplier editor">
    <header class="flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-semibold tracking-tight text-primary">{{ selectedId ? 'Edit supplier' : 'Add supplier' }}</h1>
        <p class="mt-1 text-sm text-base-content/65">Keep supplier contact and purchasing information organized.</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Suppliers' }, { label: selectedId ? 'Edit supplier' : 'Add supplier', current: true }]" @navigate="editing = false" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="busy" @click="editing = false">Cancel</button>
        <button class="btn btn-success gap-2" type="button" :disabled="busy" @click="save"><Save :size="16" aria-hidden="true" />{{ busy ? 'Saving…' : 'Save supplier' }}</button>
      </div>
    </header>

    <div class="min-h-0 min-w-0 flex-1 overflow-y-auto">
      <form class="grid w-full min-w-0 gap-6 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6" @submit.prevent="save">
        <section class="space-y-1"><h2 class="text-base font-semibold">Supplier identity</h2><p class="text-xs leading-5 text-base-content/60">Use a clear name and an optional code for purchasing documents.</p></section>
        <div class="grid min-w-0 gap-4 sm:grid-cols-2">
          <FormField label="Supplier name" required class="gap-1 sm:col-span-2"><AppInput v-model="form.name" class="input w-full min-w-0" required placeholder="Paper supplier" autocomplete="off" /></FormField>
          <FormField label="Code" class="gap-1"><AppInput v-model="form.code" class="input w-full min-w-0" placeholder="SUP-001" autocomplete="off" /></FormField>
          <FormField label="Phone" class="gap-1"><AppInput v-model="form.phone" class="input w-full min-w-0" type="tel" placeholder="+98…" autocomplete="tel" /></FormField>
          <FormField label="Email" class="gap-1 sm:col-span-2"><AppInput v-model="form.email" class="input w-full min-w-0" type="email" placeholder="contact@supplier.com" autocomplete="email" /></FormField>
          <FormField label="Address" class="gap-1 sm:col-span-2"><AppTextarea v-model="form.address" class="textarea w-full min-w-0" rows="3" placeholder="Supplier address" /></FormField>
          <FormField label="Notes" class="gap-1 sm:col-span-2"><AppTextarea v-model="form.notes" class="textarea w-full min-w-0" rows="5" placeholder="Payment terms, delivery details, or internal notes" /></FormField>
        </div>
        <div class="flex items-center gap-2 border-t border-base-300 pt-4"><span class="grid size-8 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileText :size="16" aria-hidden="true" /></span><p class="text-xs leading-5 text-base-content/60">Supplier details are stored locally and can be updated at any time.</p></div>
      </form>
    </div>
  </div>

  <div v-else class="flex h-full min-h-0 min-w-0 flex-col overflow-hidden" aria-label="Suppliers workspace">
    <WorkspaceStickyStack class="shrink-0" :flush="true">
      <WorkspaceHeader :show-breadcrumb="true" title="Suppliers" description="Manage supplier contacts and purchasing relationships for your shop.">
        <SearchField v-model="searchQuery" class="w-full min-w-0 sm:w-64" placeholder="Search suppliers…" aria-label="Search suppliers" />
        <button class="btn btn-primary w-full gap-2 sm:w-auto" type="button" @click="startCreate"><Plus :size="16" aria-hidden="true" />Add supplier</button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <div class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(22rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.15fr)_minmax(24rem,0.85fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100" aria-label="Supplier register">
        <div class="shrink-0 border-b border-base-300 p-3 sm:p-4"><div class="flex min-w-0 flex-wrap items-center gap-3"><div class="flex min-w-0 flex-wrap items-center gap-2"><button v-for="status in statusOptions" :key="status" class="inline-flex h-9 items-center gap-2 rounded-box border px-3 text-sm transition-colors" :class="supplierFilter === status ? 'border-primary bg-primary/10 text-primary' : 'border-base-300 text-base-content/70 hover:border-primary/50 hover:text-base-content'" type="button" @click="supplierFilter = status"><span class="size-2 rounded-full" :class="status === 'Active' ? 'bg-success' : status === 'Archived' ? 'bg-base-content/35' : 'bg-primary'"></span>{{ status }}<span class="rounded-full bg-base-200 px-1.5 py-0.5 text-xs tabular-nums">{{ statusCount(status) }}</span></button></div></div></div>
        <div class="supplier-register-table-head hidden gap-3 border-b border-base-300 px-4 py-3 text-xs font-medium text-base-content/55 md:grid"><span>Supplier</span><span>Contact</span><span>Location</span><span>Status</span><span></span></div>
        <LoadingState v-if="isLoading" label="Loading suppliers…" />
        <div v-else-if="pagedSuppliers.length" class="min-h-0 flex-1 overflow-y-auto divide-y divide-base-300"><button v-for="supplier in pagedSuppliers" :key="supplier.id" class="supplier-register-row group grid w-full min-w-0 items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200/60 focus-visible:bg-base-200/60 focus-visible:outline focus-visible:outline-1 focus-visible:outline-primary" :class="selectedId === supplier.id ? 'bg-primary/10' : ''" type="button" @click="selectSupplier(supplier.id)"><span class="flex min-w-0 items-center gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box border border-base-300 bg-base-200 text-primary"><Truck :size="18" aria-hidden="true" /></span><span class="min-w-0"><strong class="block truncate text-sm">{{ supplier.name }}</strong><span class="block truncate text-xs text-base-content/60">{{ supplier.code || 'No code' }}<span v-if="supplier.notes"> · {{ supplier.notes }}</span></span></span></span><span class="hidden min-w-0 truncate text-xs text-base-content/70 md:block">{{ supplier.phone || supplier.email || 'No contact' }}</span><span class="hidden min-w-0 truncate text-xs text-base-content/70 md:block">{{ supplier.address || 'No address' }}</span><StatusBadge class="justify-self-end md:justify-self-start" :label="supplier.active ? 'Active' : 'Archived'" :tone="supplier.active ? 'green' : 'slate'" /><ChevronRight :size="17" class="register-row-arrow justify-self-end text-base-content/45" aria-hidden="true" /><span class="col-span-3 flex flex-wrap gap-x-3 gap-y-1 text-xs text-base-content/55 md:hidden"><span>{{ supplier.phone || supplier.email || 'No contact' }}</span><span>{{ supplier.address || 'No address' }}</span></span></button></div>
        <EmptyState v-else :title="suppliers.length ? 'No suppliers match this view' : 'No suppliers yet'" :description="suppliers.length ? 'Try another filter or search term.' : 'Add the first supplier to start tracking purchasing relationships.'"><template #icon><Search :size="22" aria-hidden="true" /></template><template #action><button v-if="suppliers.length" class="btn btn-outline btn-sm" type="button" @click="resetListFilters">Clear filters</button><button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="startCreate"><Plus :size="15" aria-hidden="true" />Create supplier</button></template></EmptyState>
        <footer class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-4 py-3 text-xs text-base-content/60"><span>{{ visibleSuppliers.length ? `Showing ${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, visibleSuppliers.length)} of ${visibleSuppliers.length} suppliers` : '0 suppliers' }}</span><div class="flex items-center gap-1"><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === 1" aria-label="Previous page" @click="goToPage(page - 1)"><ChevronLeft :size="15" aria-hidden="true" /></button><button v-for="number in pageNumbers" :key="number" class="btn btn-xs min-w-8" :class="page === number ? 'btn-primary' : 'btn-ghost'" type="button" @click="goToPage(number)">{{ number }}</button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === pageCount" aria-label="Next page" @click="goToPage(page + 1)"><ChevronRight :size="15" aria-hidden="true" /></button></div></footer>
      </section>

      <section v-if="selectedSupplier" class="supplier-detail-panel min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:h-full xl:overflow-y-auto" aria-label="Supplier details">
        <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300 sm:min-h-60"><div class="absolute inset-0 bg-gradient-to-br from-primary/25 via-base-300 to-base-300"></div><div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 via-black/40 to-transparent p-5 pt-20 text-white sm:p-6 sm:pt-24"><div class="flex min-w-0 items-center justify-start gap-3"><h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ selectedSupplier.name }}</h2><StatusBadge class="shrink-0" :label="selectedSupplier.active ? 'Active' : 'Archived'" :tone="selectedSupplier.active ? 'green' : 'slate'" /></div><div class="mt-2 flex min-w-0 flex-wrap items-center justify-start gap-x-3 gap-y-1 text-sm text-white/75"><span>{{ selectedSupplier.code || 'No code' }}</span><span class="size-1 rounded-full bg-white/50" aria-hidden="true"></span><span>Supplier</span></div></div></div>
        <div class="flex flex-wrap items-center gap-2 border-b border-base-300 p-4"><button class="btn btn-primary btn-sm gap-2" type="button" @click="startEdit"><Edit3 :size="14" aria-hidden="true" />Edit supplier</button><button class="btn btn-outline btn-sm" :class="selectedSupplier.active ? 'btn-warning' : 'btn-success'" type="button" :disabled="busy" @click="setActive(!selectedSupplier.active)"><RotateCcw v-if="!selectedSupplier.active" :size="14" /><Archive v-else :size="14" />{{ selectedSupplier.active ? 'Archive' : 'Reactivate' }}</button><button class="btn btn-outline btn-error btn-sm gap-2" type="button" :disabled="busy" @click="remove"><Trash2 :size="14" aria-hidden="true" />Delete</button></div>
        <div class="space-y-4 p-4 sm:p-5"><div class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Phone :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Contact details</h3><p class="mt-1 text-xs leading-5 text-base-content/60">How to reach this supplier for orders and delivery.</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex items-start justify-between gap-3 py-2"><dt class="flex items-center gap-2 text-base-content/60"><Phone :size="14" aria-hidden="true" />Phone</dt><dd class="min-w-0 text-end tabular-nums wrap-anywhere">{{ selectedSupplier.phone || '—' }}</dd></div><div class="flex items-start justify-between gap-3 py-2"><dt class="flex items-center gap-2 text-base-content/60"><Mail :size="14" aria-hidden="true" />Email</dt><dd class="min-w-0 text-end wrap-anywhere">{{ selectedSupplier.email || '—' }}</dd></div><div class="flex items-start justify-between gap-3 py-2 last:pb-0"><dt class="flex items-center gap-2 text-base-content/60"><MapPin :size="14" aria-hidden="true" />Address</dt><dd class="min-w-0 max-w-[65%] text-end wrap-anywhere">{{ selectedSupplier.address || '—' }}</dd></div></dl></div><div v-if="selectedSupplier.notes" class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileText :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Notes</h3><p class="mt-2 text-sm leading-6 text-base-content/70">{{ selectedSupplier.notes }}</p></div></div></div><div class="rounded-box border border-base-300 bg-base-200/20 p-4"><h3 class="text-sm font-semibold">Supplier information</h3><dl class="mt-3 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Code</dt><dd>{{ selectedSupplier.code || 'No code' }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Created</dt><dd>{{ formatDateTime(selectedSupplier.createdAt) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Updated</dt><dd>{{ formatDateTime(selectedSupplier.updatedAt) }}</dd></div></dl></div><div class="flex items-start gap-2 border-t border-base-300 pt-3 text-xs leading-5 text-base-content/60"><Truck :size="15" class="mt-0.5 shrink-0 text-info" aria-hidden="true" /><span>Use Edit supplier to update contact details without changing purchasing history.</span></div></div>
      </section>
      <section v-else class="flex min-h-72 min-w-0 items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><EmptyState title="Select a supplier" description="Choose a supplier from the register to inspect contact and purchasing information."><template #icon><Truck :size="22" aria-hidden="true" /></template></EmptyState></section>
    </div>
  </div>
</template>

<style scoped>
.supplier-register-table-head,
.supplier-register-row {
  grid-template-columns: minmax(0, 1.5fr) minmax(8rem, 0.8fr) minmax(8rem, 0.8fr) 6rem 1.25rem;
}

@media (max-width: 767px) {
  .supplier-register-row {
    grid-template-columns: minmax(0, 1fr) auto auto;
  }
}
</style>
