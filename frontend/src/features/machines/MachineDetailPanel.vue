<script setup lang="ts">
import { ref, watch } from 'vue'
import { Archive, Edit3, Factory, FileText, Gauge, Layers3, RotateCcw, Tag, Trash2 } from 'lucide-vue-next'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'
import type { useMachinesWorkspace } from './useMachinesWorkspace'

type MachineTab = 'overview' | 'rates' | 'details'

const props = defineProps<{
  workspace: ReturnType<typeof useMachinesWorkspace>
  currencyUnit: CurrencyUnit
}>()
const emit = defineEmits<{ edit: []; archive: []; reactivate: []; remove: [] }>()

const { busy, selectedMachine, basisLabel } = props.workspace
const activeTab = ref<MachineTab>('overview')
watch(() => selectedMachine.value?.id, () => { activeTab.value = 'overview' })

const rates = () => selectedMachine.value?.rates || []
const standardRate = () => selectedMachine.value?.rateRial || 0
const rateCount = () => rates().length || 1
function dateLabel(value: string) {
  try { return formatDateTime(value) } catch { return 'Unknown date' }
}
</script>

<template>
  <section v-if="selectedMachine" class="machine-detail-panel h-auto min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:h-full xl:overflow-y-auto" aria-label="Machine details">
    <header class="border-b border-base-300 p-3 sm:p-5">
      <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300 bg-cover bg-center sm:min-h-60" :style="selectedMachine.imagePath ? { backgroundImage: `url('${selectedMachine.imagePath}')` } : undefined">
        <div class="absolute inset-0 bg-gradient-to-l from-black/95 via-black/65 to-black/10" aria-hidden="true"></div>
        <div v-if="!selectedMachine.imagePath" class="absolute inset-0 grid place-items-center text-base-content/30"><Factory :size="42" aria-hidden="true" /></div>
        <div class="relative z-10 flex min-h-52 items-end justify-start p-4 text-white sm:min-h-60 sm:p-5">
          <div class="w-full min-w-0 text-start">
            <div class="flex min-w-0 items-center justify-start gap-3"><h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ selectedMachine.name }}</h2><StatusBadge class="shrink-0" :label="selectedMachine.active ? 'Active' : 'Archived'" :tone="selectedMachine.active ? 'green' : 'slate'" /></div>
            <div class="mt-2 flex min-w-0 flex-wrap items-center justify-start gap-x-3 gap-y-1 text-sm text-white/75"><span>{{ selectedMachine.code || 'No code' }}</span><span class="size-1 rounded-full bg-white/50" aria-hidden="true"></span><span>{{ selectedMachine.category || 'Uncategorized' }}</span></div>
          </div>
        </div>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
        <button class="btn btn-outline btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('edit')"><Edit3 :size="14" aria-hidden="true" />Edit</button>
        <button v-if="selectedMachine.active" class="btn btn-outline btn-warning btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('archive')"><Archive :size="14" aria-hidden="true" />Archive</button>
        <button v-else class="btn btn-outline btn-success btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('reactivate')"><RotateCcw :size="14" aria-hidden="true" />Reactivate</button>
        <button class="btn btn-outline btn-error btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('remove')"><Trash2 :size="14" aria-hidden="true" />Remove</button>
      </div>

      <div class="mt-4 grid min-w-0 divide-y divide-base-300 border-y border-base-300 sm:mt-5 sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:first:pl-0"><Gauge :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Standard rate</span><strong class="block text-sm tabular-nums">{{ formatMoney(standardRate(), currencyUnit) }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3"><Tag :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Rate basis</span><strong class="block text-sm">{{ basisLabel(selectedMachine.rateBasis) }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:pr-0"><Layers3 :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Rate profiles</span><strong class="block text-sm">{{ rateCount() }}</strong></div></div>
      </div>
    </header>

    <nav class="flex min-w-0 overflow-x-auto border-b border-base-300 px-2" aria-label="Machine details tabs">
      <button v-for="tab in [{ id: 'overview', label: 'Overview' }, { id: 'rates', label: 'Rate profiles' }, { id: 'details', label: 'Additional info' }]" :key="tab.id" class="shrink-0 border-b-2 px-3 py-3 text-sm transition-colors" :class="activeTab === tab.id ? 'border-primary text-primary' : 'border-transparent text-base-content/65 hover:border-base-content/30 hover:text-base-content'" type="button" @click="activeTab = tab.id as MachineTab">{{ tab.label }}</button>
    </nav>

    <div class="min-w-0 p-3 sm:p-4">
      <div v-if="activeTab === 'overview'" class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Gauge :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Rate definition</h3><p class="mt-1 text-xs leading-5 text-base-content/60">The standard rate is used when no matching profile is selected.</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Standard rate</dt><dd class="font-medium tabular-nums">{{ formatMoney(selectedMachine.rateRial, currencyUnit) }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Setup / fixed cost</dt><dd class="font-medium tabular-nums">{{ formatMoney(selectedMachine.setupCostRial, currencyUnit) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Basis</dt><dd class="text-end">{{ basisLabel(selectedMachine.rateBasis) }}</dd></div></dl></div>
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Layers3 :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Available profiles</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Profiles can be linked to service parameters such as color or media mode.</p></div></div><div class="mt-4 space-y-2"><div v-for="rate in rates()" :key="rate.id" class="flex min-w-0 items-center justify-between gap-3 rounded-box border border-base-300 px-3 py-2.5"><div class="min-w-0"><strong class="block truncate text-sm">{{ rate.name }}</strong><span class="block truncate text-xs text-base-content/55">{{ rate.selectorValue || 'Default profile' }} · {{ basisLabel(rate.rateBasis) }}</span></div><span class="shrink-0 text-sm tabular-nums">{{ formatMoney(rate.rateRial, currencyUnit) }}</span></div><EmptyState v-if="!rates().length" compact title="No profiles configured" description="The standard rate will be used." /></div></div>
      </div>

      <div v-else-if="activeTab === 'rates'" class="space-y-2">
        <div v-for="rate in rates()" :key="rate.id" class="flex min-w-0 items-center gap-3 rounded-box border border-base-300 bg-base-200/20 px-3 py-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Layers3 :size="18" aria-hidden="true" /></span><div class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ rate.name }}</strong><span class="block truncate text-xs text-base-content/55">{{ rate.selectorValue || 'Matches any value' }} · {{ basisLabel(rate.rateBasis) }}</span></div><div class="shrink-0 text-end"><strong class="block text-sm tabular-nums">{{ formatMoney(rate.rateRial, currencyUnit) }}</strong><span class="block text-xs text-base-content/55">setup {{ formatMoney(rate.setupCostRial, currencyUnit) }}</span></div></div>
      </div>

      <div v-else class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileText :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Machine information</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Catalog identity and operating context.</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Code</dt><dd class="break-words text-end">{{ selectedMachine.code || 'No code' }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Category</dt><dd class="break-words text-end">{{ selectedMachine.category || 'Uncategorized' }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Created</dt><dd class="text-end">{{ dateLabel(selectedMachine.createdAt) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Updated</dt><dd class="text-end">{{ dateLabel(selectedMachine.updatedAt) }}</dd></div></dl></div>
        <div v-if="selectedMachine.notes" class="rounded-box border border-base-300 bg-base-200/20 p-4"><h3 class="text-sm font-semibold">Notes</h3><p class="mt-2 whitespace-pre-wrap break-words text-sm leading-6 text-base-content/70">{{ selectedMachine.notes }}</p></div>
      </div>
    </div>
  </section>
</template>
