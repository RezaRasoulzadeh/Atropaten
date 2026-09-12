<script setup lang="ts">
import { ref, watch } from 'vue'
import { Archive, Edit3, FileText, Package, RefreshCw, RotateCcw, Tag, Trash2 } from 'lucide-vue-next'
import EmptyState from '../../components/ui/EmptyState.vue'
import FormField from '../../components/ui/FormField.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import type { useMaterialsWorkspace } from './useMaterialsWorkspace'

type MaterialTab = 'overview' | 'movements' | 'details'

const props = defineProps<{
  workspace: ReturnType<typeof useMaterialsWorkspace>
  currencyUnit: CurrencyUnit
}>()

const {
  busy,
  selectedMaterial,
  isMovementsLoading,
  movements,
  adjustmentQuantity,
  adjustmentCost,
  adjustmentNote,
  updateAdjustmentCost,
  adjustStock,
  cancelMovement,
  startEdit,
  setActive,
  remove,
  unitLabel,
  dateLabel,
} = props.workspace

const activeTab = ref<MaterialTab>('overview')
watch(() => selectedMaterial.value?.id, () => { activeTab.value = 'overview' })

function pricingUnitCost() {
  if (!selectedMaterial.value) return 0
  return selectedMaterial.value.highestPurchaseUnitCostRial || selectedMaterial.value.averageUnitCostRial
}
</script>

<template>
  <section v-if="selectedMaterial" class="material-detail-panel h-auto min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:h-full xl:overflow-y-auto" aria-label="Material details">
    <header class="border-b border-base-300 p-3 sm:p-5">
      <div class="rounded-box bg-base-200/70 p-4 sm:p-5">
        <div class="flex min-w-0 items-start gap-3">
          <span class="grid size-11 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Package :size="24" aria-hidden="true" /></span>
          <div class="min-w-0 flex-1">
            <div class="flex min-w-0 flex-wrap items-center justify-between gap-2">
              <h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ selectedMaterial.name }}</h2>
              <div class="flex shrink-0 flex-wrap gap-1.5"><StatusBadge v-if="selectedMaterial.lowStock" label="Low stock" tone="amber" /><StatusBadge :label="selectedMaterial.active ? 'Active' : 'Archived'" :tone="selectedMaterial.active ? 'green' : 'slate'" /></div>
            </div>
            <p class="mt-1 truncate text-sm text-base-content/65">{{ selectedMaterial.sku || 'No SKU' }}<span v-if="selectedMaterial.category"> · {{ selectedMaterial.category }}</span></p>
            <p v-if="selectedMaterial.notes" class="mt-3 line-clamp-2 text-sm leading-5 text-base-content/65">{{ selectedMaterial.notes }}</p>
          </div>
        </div>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
        <button class="btn btn-outline btn-sm w-full gap-2" type="button" :disabled="busy" @click="startEdit"><Edit3 :size="14" aria-hidden="true" />Edit</button>
        <button v-if="selectedMaterial.active" class="btn btn-outline btn-warning btn-sm w-full gap-2" type="button" :disabled="busy" @click="setActive(false)"><Archive :size="14" aria-hidden="true" />Archive</button>
        <button v-else class="btn btn-outline btn-success btn-sm w-full gap-2" type="button" :disabled="busy" @click="setActive(true)"><RotateCcw :size="14" aria-hidden="true" />Reactivate</button>
        <button class="btn btn-outline btn-error btn-sm w-full gap-2" type="button" :disabled="busy" @click="remove"><Trash2 :size="14" aria-hidden="true" />Remove</button>
      </div>

      <div class="mt-4 grid min-w-0 divide-y divide-base-300 border-y border-base-300 sm:mt-5 sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:first:pl-0"><Package :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Available stock</span><strong class="block text-sm tabular-nums">{{ selectedMaterial.availableStock }} {{ selectedMaterial.consumptionUnit }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3"><Tag :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Unit cost</span><strong class="block text-sm tabular-nums">{{ formatMoney(pricingUnitCost(), currencyUnit) }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:pr-0"><FileText :size="20" class="shrink-0 text-primary" aria-hidden="true" /><div><span class="block text-xs text-base-content/55">Inventory value</span><strong class="block text-sm tabular-nums">{{ formatMoney(selectedMaterial.inventoryValueRial, currencyUnit) }}</strong></div></div>
      </div>
    </header>

    <nav class="flex min-w-0 overflow-x-auto border-b border-base-300 px-2" aria-label="Material details tabs">
      <button v-for="tab in [{ id: 'overview', label: 'Overview' }, { id: 'movements', label: 'Movements' }, { id: 'details', label: 'Additional info' }]" :key="tab.id" class="shrink-0 border-b-2 px-3 py-3 text-sm transition-colors" :class="activeTab === tab.id ? 'border-primary text-primary' : 'border-transparent text-base-content/65 hover:border-base-content/30 hover:text-base-content'" type="button" @click="activeTab = tab.id as MaterialTab">{{ tab.label }}</button>
    </nav>

    <div class="min-w-0 p-3 sm:p-4">
      <div v-if="activeTab === 'overview'" class="space-y-4">
        <div class="grid min-w-0 gap-3 sm:grid-cols-2">
          <div class="rounded-box border border-base-300 p-3"><span class="block text-xs text-base-content/55">Physical stock</span><strong class="mt-1 block text-xl tabular-nums">{{ selectedMaterial.physicalStock }}</strong><span class="text-xs text-base-content/55">{{ selectedMaterial.consumptionUnit }}</span></div>
          <div class="rounded-box border border-base-300 p-3"><span class="block text-xs text-base-content/55">Reserved stock</span><strong class="mt-1 block text-xl tabular-nums">{{ selectedMaterial.reservedStock }}</strong><span class="text-xs text-base-content/55">{{ selectedMaterial.consumptionUnit }}</span></div>
          <div class="rounded-box border border-base-300 p-3"><span class="block text-xs text-base-content/55">Reorder level</span><strong class="mt-1 block text-xl tabular-nums">{{ selectedMaterial.reorderLevel }}</strong><span class="text-xs text-base-content/55">{{ selectedMaterial.consumptionUnit }}</span></div>
          <div class="rounded-box border border-base-300 p-3"><span class="block text-xs text-base-content/55">Inventory average</span><strong class="mt-1 block text-sm tabular-nums">{{ formatMoney(selectedMaterial.averageUnitCostRial, currencyUnit) }}</strong><span class="text-xs text-base-content/55">per {{ selectedMaterial.consumptionUnit }}</span></div>
        </div>

        <div data-enter-scope class="rounded-box border border-base-300 p-4">
          <div class="flex items-start justify-between gap-3"><div><h3 class="text-sm font-semibold">Adjust stock</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Record a correction or movement without changing purchase history.</p></div><RefreshCw :size="18" class="shrink-0 text-primary" aria-hidden="true" /></div>
          <div class="mt-4 grid min-w-0 gap-3 sm:grid-cols-2">
            <FormField class="gap-1"><span class="text-xs">Quantity delta</span><input v-model="adjustmentQuantity" class="input input-sm w-full min-w-0" placeholder="−1 or 2.5" inputmode="decimal" /><small class="text-xs leading-5 text-base-content/55">Positive adds stock; negative records a correction.</small></FormField>
            <FormField class="gap-1"><span class="text-xs">Unit cost ({{ currencyUnit }})</span><input :value="adjustmentCost" class="input input-sm w-full min-w-0" inputmode="numeric" @input="updateAdjustmentCost(($event.target as HTMLInputElement).value)" /></FormField>
            <FormField class="gap-1 sm:col-span-2"><span class="text-xs">Reason / note</span><input v-model="adjustmentNote" class="input input-sm w-full min-w-0" placeholder="Count correction" /></FormField>
          </div>
          <div class="mt-3 flex justify-end border-t border-base-300 pt-3"><button class="btn btn-primary btn-sm gap-2" type="button" :disabled="busy" data-enter-submit @click="adjustStock"><RefreshCw :size="14" aria-hidden="true" />Record movement</button></div>
        </div>
      </div>

      <div v-else-if="activeTab === 'movements'" class="space-y-3">
        <LoadingState v-if="isMovementsLoading" label="Loading movements…" />
        <div v-else-if="movements.length" class="space-y-2" role="list" aria-label="Material movement history">
          <article v-for="movement in movements" :key="movement.id" class="rounded-box border border-base-300 bg-base-200/20 p-3 text-sm transition-colors hover:bg-base-200/45" role="listitem">
            <div class="material-movement-grid min-w-0">
              <div class="material-movement-kind min-w-0"><strong class="block truncate">{{ movement.movementType }}</strong><span class="block truncate text-xs text-base-content/55">{{ dateLabel(movement.occurredAt) }}</span></div>
              <div class="material-movement-metric min-w-0"><span class="material-movement-mobile-label text-xs text-base-content/55">Quantity</span><span class="tabular-nums" :class="String(movement.quantityDelta).startsWith('-') ? 'text-error' : 'text-success'">{{ movement.quantityDelta }} {{ selectedMaterial.consumptionUnit }}</span></div>
              <div class="material-movement-metric min-w-0"><span class="material-movement-mobile-label text-xs text-base-content/55">Unit cost</span><span class="tabular-nums">{{ formatMoney(movement.unitCostRial, currencyUnit) }}</span></div>
              <div class="material-movement-metric min-w-0"><span class="material-movement-mobile-label text-xs text-base-content/55">Total value</span><span class="font-medium tabular-nums">{{ formatMoney(movement.totalCostRial, currencyUnit) }}</span></div>
            </div>
            <div class="material-movement-source mt-2 flex min-w-0 items-start justify-between gap-3 border-t border-base-300/70 pt-2 text-xs"><span class="shrink-0 text-base-content/50">Note / source</span><span class="flex min-w-0 flex-wrap items-center justify-end gap-2 break-words text-end text-base-content/65"><span>{{ movement.note || movement.referenceType || '—' }}</span><button v-if="movement.referenceType === 'manual_adjustment'" class="btn btn-outline btn-error btn-xs gap-1" type="button" :disabled="busy" @click.stop="cancelMovement(movement.id)"><RotateCcw :size="12" aria-hidden="true" />Cancel</button></span></div>
          </article>
        </div>
        <EmptyState v-else compact title="No inventory movements" description="Stock receipts, usage, and adjustments will appear here."><template #icon><Package :size="21" aria-hidden="true" /></template></EmptyState>
      </div>

      <div v-else class="space-y-3">
        <div class="rounded-box border border-base-300 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileText :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Material information</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Catalog identity and unit configuration.</p></div></div><dl class="mt-4 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">SKU / category</dt><dd class="break-words text-end">{{ selectedMaterial.sku || 'No SKU' }} · {{ selectedMaterial.category || 'Uncategorized' }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Unit conversion</dt><dd class="break-words text-end">1 {{ unitLabel(selectedMaterial.purchaseUnit) }} = {{ selectedMaterial.conversionFactor }} {{ unitLabel(selectedMaterial.consumptionUnit) }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Preferred supplier</dt><dd class="break-words text-end">{{ selectedMaterial.preferredSupplier || 'Not specified' }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Updated</dt><dd class="text-end">{{ dateLabel(selectedMaterial.updatedAt) }}</dd></div></dl></div>
        <div v-if="selectedMaterial.notes" class="rounded-box border border-base-300 p-4"><h3 class="text-sm font-semibold">Notes</h3><p class="mt-2 whitespace-pre-wrap break-words text-sm leading-6 text-base-content/70">{{ selectedMaterial.notes }}</p></div>
      </div>
    </div>
  </section>
</template>
