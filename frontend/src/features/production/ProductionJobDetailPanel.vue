<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUpRight, CheckCircle2, Clock3, Factory, FileText, Package, Trash2, UserRound } from 'lucide-vue-next'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { CurrencyUnit } from '../../utils/currency'
import ProductionCostBreakdown from './ProductionCostBreakdown.vue'
import type { useProductionWorkspace } from './useProductionWorkspace'

const props = defineProps<{
  workspace: ReturnType<typeof useProductionWorkspace>
  currencyUnit: CurrencyUnit
  machines: any[]
  busy?: boolean
}>()

const emit = defineEmits<{
  open: []
  remove: []
}>()

const { selected, selectedOrder, selectedItem, jobOrder, statusTone, date } = props.workspace

const machineName = computed(
  () => props.machines.find((machine) => machine.id === selected.value?.assignedMachineId)?.name || 'Unassigned',
)
const customerName = computed(() => selectedOrder.value?.customerName || 'Walk-in customer')

</script>

<template>
  <section class="production-detail-panel min-w-0" aria-label="Production job details">
    <div class="border-b border-base-300 p-3 sm:p-5">
      <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300 sm:min-h-60">
        <div class="absolute inset-0 bg-gradient-to-br from-primary/20 via-base-300 to-base-200" aria-hidden="true"></div>
        <div class="absolute -end-10 -top-12 size-48 rounded-full bg-primary/10 blur-2xl" aria-hidden="true"></div>
        <div class="relative z-10 flex min-h-52 flex-col justify-between p-4 sm:min-h-60 sm:p-5">
          <div class="grid size-11 place-items-center rounded-box border border-primary/30 bg-base-100/70 text-primary"><Factory :size="24" :stroke-width="1.7" aria-hidden="true" /></div>
          <div class="min-w-0">
            <div class="flex min-w-0 items-center gap-3"><h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ selected?.jobNumber }}</h2><StatusBadge class="shrink-0" :label="selected?.status || 'Pending'" :tone="statusTone(selected?.status || 'Pending')" /></div>
            <p class="mt-2 truncate text-sm text-base-content/75">{{ selected?.serviceName }}</p>
            <p class="mt-1 truncate text-xs text-base-content/55">{{ jobOrder(selected!) }} · {{ customerName }}</p>
          </div>
        </div>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-2">
        <button class="btn btn-primary btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('open')"><ArrowUpRight :size="14" aria-hidden="true" />Manage production</button>
        <button class="btn btn-outline btn-error btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('remove')"><Trash2 :size="14" aria-hidden="true" />Remove</button>
      </div>

      <div class="mt-4 grid min-w-0 divide-y divide-base-300 border-y border-base-300 sm:mt-5 sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:first:pl-0"><UserRound :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div class="min-w-0"><span class="block text-xs text-base-content/55">Customer / order</span><strong class="block truncate text-sm">{{ customerName }}</strong><span class="block truncate text-xs text-base-content/60">{{ jobOrder(selected!) }}</span></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3"><Package :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div class="min-w-0"><span class="block text-xs text-base-content/55">Quantity</span><strong class="block truncate text-sm">{{ selected?.quantity }} {{ selected?.quantityUnit }}</strong><span class="block truncate text-xs text-base-content/60">{{ selected?.priority }} priority</span></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:pr-0"><Factory :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div class="min-w-0"><span class="block text-xs text-base-content/55">Machine</span><strong class="block truncate text-sm">{{ machineName }}</strong><span class="block truncate text-xs text-base-content/60">{{ selected?.plannedAt ? date(selected.plannedAt) : 'Not scheduled' }}</span></div></div>
      </div>
    </div>

    <div class="min-w-0 space-y-3 p-3 sm:p-4">
      <div class="rounded-box border border-base-300 bg-base-200/20 p-4">
        <div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><CheckCircle2 :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Job overview</h3><p class="mt-1 text-xs leading-5 text-base-content/60">The saved production snapshot and current workflow state.</p></div></div>
        <dl class="mt-4 divide-y divide-base-300/70 text-sm">
          <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Service item</dt><dd class="truncate text-end">{{ selected?.serviceName }}</dd></div>
          <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Order item</dt><dd class="truncate text-end">{{ selectedItem?.serviceCode || 'Saved order item' }}</dd></div>
          <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Created</dt><dd class="text-end">{{ date(selected?.createdAt || '') }}</dd></div>
          <div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Status</dt><dd><StatusBadge :label="selected?.status || 'Pending'" :tone="statusTone(selected?.status || 'Pending')" /></dd></div>
        </dl>
      </div>

      <ProductionCostBreakdown v-if="selected" :job="selected" :currency-unit="currencyUnit" />

      <div class="rounded-box border border-base-300 bg-base-200/20 p-4">
        <div class="flex items-center gap-2"><FileText :size="17" class="text-primary" aria-hidden="true" /><h3 class="text-sm font-semibold">Schedule</h3></div>
        <dl class="mt-3 divide-y divide-base-300/70 text-sm">
          <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Started</dt><dd class="text-end">{{ date(selected?.startedAt || '') }}</dd></div>
          <div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Completed</dt><dd class="text-end">{{ date(selected?.completedAt || '') }}</dd></div>
        </dl>
      </div>

      <div class="rounded-box border border-base-300 bg-base-200/20 p-4">
        <div class="flex items-center gap-2"><Clock3 :size="17" class="text-primary" aria-hidden="true" /><h3 class="text-sm font-semibold">Production tracking</h3></div>
        <p class="mt-3 text-sm leading-6 text-base-content/65">Review the Materials plan, start production, then complete the job to record remaining stock usage. Waste and corrections are available in Materials.</p>
      </div>

      <div v-if="selected?.notes" class="rounded-box border border-base-300 bg-base-200/20 p-4"><h3 class="text-sm font-semibold">Notes</h3><p class="mt-2 whitespace-pre-wrap text-sm leading-6 text-base-content/70">{{ selected.notes }}</p></div>
    </div>
  </section>
</template>
