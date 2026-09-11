<script setup lang="ts">
import { computed } from 'vue';
import {
  Archive,
  ArrowLeft,
  Edit3,
  Package,
  RefreshCw,
  RotateCcw,
  Save,
  Trash2,
} from 'lucide-vue-next';
import AppInput from '../../components/ui/AppInput.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import DataTable from '../../components/ui/DataTable.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import FormField from '../../components/ui/FormField.vue';
import FormSection from '../../components/ui/FormSection.vue';
import InspectorSection from '../../components/layout/InspectorSection.vue';
import LoadingState from '../../components/ui/LoadingState.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import SelectField from '../../components/ui/SelectField.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney } from '../../utils/currency';
import type { useMaterialsWorkspace } from './useMaterialsWorkspace';

const props = defineProps<{
  workspace: ReturnType<typeof useMaterialsWorkspace>;
  currencyUnit: CurrencyUnit;
}>();
const emit = defineEmits<{ back: [] }>();

const {
  busy,
  selectedMaterial,
  editorMode,
  form,
  costDraft,
  isSaving,
  movements,
  isMovementsLoading,
  adjustmentQuantity,
  adjustmentCost,
  adjustmentNote,
  unitOptions,
  startEdit,
  cancelEditor,
  updateCost,
  saveMaterial,
  adjustStock,
  setActive,
  remove,
  unitLabel,
  dateLabel,
  updateAdjustmentCost,
} = props.workspace;

const isCreating = computed(() => editorMode.value === 'create');
const isEditing = computed(() => editorMode.value === 'edit');
</script>

<template>
  <div class="min-w-0 space-y-4" aria-label="Material workspace">
    <WorkspaceStickyStack :flush="true">
      <WorkspaceHeader
        :title="isCreating ? 'New material' : isEditing ? 'Edit material' : selectedMaterial?.name || 'Material'"
        :description="
          isCreating
            ? 'Create a stock item with units, conversion, and an optional opening cost.'
            : isEditing
              ? 'Update catalog details without rewriting inventory history.'
              : 'Review stock position, cost basis, movements, and catalog details.'
        "
      >
        <template v-if="selectedMaterial && !isCreating" #title-suffix>
          <StatusBadge :label="selectedMaterial.active ? 'Active' : 'Archived'" :tone="selectedMaterial.active ? 'green' : 'slate'" />
          <StatusBadge v-if="selectedMaterial.lowStock" label="Low stock" tone="amber" />
        </template>
        <button class="btn btn-ghost btn-sm gap-2" type="button" :disabled="busy" @click="emit('back')">
          <ArrowLeft :size="16" aria-hidden="true" />Materials
        </button>
        <template v-if="editorMode">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="cancelEditor">Cancel</button>
          <button class="btn btn-primary btn-sm gap-2" type="submit" form="material-editor" :disabled="busy || isSaving">
            <Save :size="14" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save material' }}
          </button>
        </template>
        <template v-else-if="selectedMaterial">
          <button class="btn btn-outline btn-error btn-sm gap-2" type="button" :disabled="busy" @click="remove">
            <Trash2 :size="14" aria-hidden="true" />Delete
          </button>
          <button v-if="selectedMaterial.active" class="btn btn-outline btn-warning btn-sm gap-2" type="button" :disabled="busy" @click="setActive(false)">
            <Archive :size="14" aria-hidden="true" />Archive
          </button>
          <button v-else class="btn btn-outline btn-success btn-sm gap-2" type="button" :disabled="busy" @click="setActive(true)">
            <RotateCcw :size="14" aria-hidden="true" />Reactivate
          </button>
          <button class="btn btn-outline btn-sm gap-2" type="button" :disabled="busy" @click="startEdit">
            <Edit3 :size="14" aria-hidden="true" />Edit
          </button>
        </template>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <AppPanel
      v-if="editorMode"
      :title="isCreating ? 'Material details' : 'Edit material'"
      :subtitle="isCreating ? 'Catalog fields and opening stock configuration.' : 'Catalog changes are saved without rewriting inventory movements.'"
    >
      <form id="material-editor" class="min-w-0 space-y-4" @submit.prevent="saveMaterial">
        <FormSection title="Identity" description="Use a clear name and stable catalog reference.">
          <FormField class="gap-1 sm:col-span-2">
            <span class="text-xs">Name</span>
            <AppInput v-model="form.name" class="input w-full min-w-0" type="text" placeholder="A4 80gsm Paper" autocomplete="off" />
          </FormField>
          <FormField class="gap-1">
            <span class="text-xs">SKU / code</span>
            <AppInput v-model="form.sku" class="input w-full min-w-0" type="text" placeholder="PAPER-A4" autocomplete="off" />
          </FormField>
          <FormField class="gap-1">
            <span class="text-xs">Category</span>
            <AppInput v-model="form.category" class="input w-full min-w-0" type="text" placeholder="Paper" autocomplete="off" />
          </FormField>
        </FormSection>

        <FormSection title="Units and opening stock" description="Keep purchase and production quantities explicit. Posted purchases set the pricing basis automatically.">
          <SelectField v-model="form.purchaseUnit" label="Purchase unit" :options="unitOptions.map((unit) => ({ label: unitLabel(unit), value: unit }))" />
          <SelectField v-model="form.consumptionUnit" label="Consumption unit" :options="unitOptions.map((unit) => ({ label: unitLabel(unit), value: unit }))" />
          <FormField class="gap-1 sm:col-span-2">
            <span class="text-xs">Conversion factor</span>
            <AppInput v-model="form.conversionFactor" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="500" />
            <small class="text-xs text-base-content/60">1 {{ form.purchaseUnit }} = {{ form.conversionFactor || '…' }} {{ form.consumptionUnit }}</small>
          </FormField>
          <FormField class="gap-1">
            <span class="text-xs">Opening physical stock</span>
            <AppInput v-model="form.physicalStock" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" :disabled="isEditing" />
            <small v-if="isEditing" class="text-xs text-base-content/60">Use Adjust stock for ledger movements.</small>
          </FormField>
          <FormField class="gap-1">
            <span class="text-xs">Reorder level</span>
            <AppInput v-model="form.reorderLevel" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" />
          </FormField>
          <FormField v-if="isCreating" class="gap-1 sm:col-span-2">
            <span class="text-xs">Opening unit cost / {{ form.consumptionUnit }} ({{ props.currencyUnit }})</span>
            <AppInput :model-value="costDraft" :money="props.currencyUnit" type="text" inputmode="decimal" placeholder="0" @update:model-value="updateCost" />
            <small class="text-xs text-base-content/60">Used for opening stock only. After a purchase is posted, pricing uses the highest landed purchase cost automatically.</small>
          </FormField>
        </FormSection>

        <FormSection title="Supplier and notes" description="Optional context for purchasing and storage.">
          <FormField class="gap-1 sm:col-span-2">
            <span class="text-xs">Preferred supplier <em>optional</em></span>
            <AppInput v-model="form.preferredSupplier" class="input w-full min-w-0" type="text" placeholder="Pars Paper" autocomplete="off" />
          </FormField>
          <FormField class="gap-1 sm:col-span-2">
            <span class="text-xs">Notes <em>optional</em></span>
            <AppTextarea v-model="form.notes" rows="3" placeholder="Storage or handling note" />
          </FormField>
        </FormSection>

      </form>
    </AppPanel>

    <template v-else-if="selectedMaterial">
      <AppPanel title="Material overview" subtitle="Persisted catalog and inventory position.">
        <div class="grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-5">
          <div class="rounded-box border border-base-300 bg-base-200/45 p-3">
            <span class="block text-xs text-base-content/60">Physical stock</span>
            <strong class="mt-1 block text-xl leading-6 tabular-nums">{{ selectedMaterial.physicalStock }}</strong>
            <span class="mt-1 block text-xs text-base-content/55">{{ selectedMaterial.consumptionUnit }}</span>
          </div>
          <div class="rounded-box border border-base-300 bg-base-200/45 p-3">
            <span class="block text-xs text-base-content/60">Available stock</span>
            <strong class="mt-1 block text-xl leading-6 tabular-nums">{{ selectedMaterial.availableStock }}</strong>
            <span class="mt-1 block text-xs text-base-content/55">{{ selectedMaterial.reservedStock }} reserved</span>
          </div>
          <div class="rounded-box border border-primary/25 bg-primary/5 p-3">
            <span class="block text-xs text-base-content/60">Pricing unit cost</span>
            <strong class="mt-1 block truncate text-sm font-semibold text-primary">{{ formatMoney(selectedMaterial.highestPurchaseUnitCostRial || selectedMaterial.averageUnitCostRial, props.currencyUnit) }}</strong>
            <span class="mt-1 block text-xs text-base-content/55">highest posted purchase</span>
          </div>
          <div class="rounded-box border border-base-300 bg-base-200/45 p-3">
            <span class="block text-xs text-base-content/60">Inventory average</span>
            <strong class="mt-1 block truncate text-sm font-semibold">{{ formatMoney(selectedMaterial.averageUnitCostRial, props.currencyUnit) }}</strong>
            <span class="mt-1 block text-xs text-base-content/55">valuation per {{ selectedMaterial.consumptionUnit }}</span>
          </div>
          <div class="rounded-box border border-base-300 bg-base-200/45 p-3">
            <span class="block text-xs text-base-content/60">Inventory value</span>
            <strong class="mt-1 block truncate text-sm font-semibold">{{ formatMoney(selectedMaterial.inventoryValueRial, props.currencyUnit) }}</strong>
            <span class="mt-1 block text-xs text-base-content/55">Updated {{ dateLabel(selectedMaterial.updatedAt) }}</span>
          </div>
        </div>

        <div class="grid min-w-0 gap-4 xl:grid-cols-2">
          <InspectorSection title="Catalog details" description="Units, reorder threshold, and reference data.">
            <dl class="grid min-w-0 gap-3 rounded-box border border-base-300 p-3 text-sm sm:grid-cols-2">
              <div>
                <dt class="text-xs text-base-content/60">SKU / category</dt>
                <dd class="mt-1 break-words font-medium">{{ selectedMaterial.sku || 'No SKU' }} · {{ selectedMaterial.category || 'Uncategorized' }}</dd>
              </div>
              <div>
                <dt class="text-xs text-base-content/60">Unit conversion</dt>
                <dd class="mt-1 font-medium">1 {{ unitLabel(selectedMaterial.purchaseUnit) }} = {{ selectedMaterial.conversionFactor }} {{ unitLabel(selectedMaterial.consumptionUnit) }}</dd>
              </div>
              <div>
                <dt class="text-xs text-base-content/60">Reorder level</dt>
                <dd class="mt-1 font-medium tabular-nums">{{ selectedMaterial.reorderLevel }} {{ selectedMaterial.consumptionUnit }}</dd>
              </div>
              <div>
                <dt class="text-xs text-base-content/60">Last updated</dt>
                <dd class="mt-1 font-medium">{{ dateLabel(selectedMaterial.updatedAt) }}</dd>
              </div>
            </dl>
          </InspectorSection>
          <InspectorSection title="Supplier and notes" description="Operational context kept with the catalog record.">
            <div class="space-y-3 rounded-box border border-base-300 p-3 text-sm">
              <p><span class="text-xs text-base-content/60">Preferred supplier</span><strong class="mt-1 block break-words">{{ selectedMaterial.preferredSupplier || 'Not specified' }}</strong></p>
              <p><span class="text-xs text-base-content/60">Notes</span><span class="mt-1 block whitespace-pre-wrap break-words">{{ selectedMaterial.notes || 'No notes recorded.' }}</span></p>
            </div>
          </InspectorSection>
        </div>
      </AppPanel>

      <AppPanel title="Adjust stock" subtitle="Each adjustment creates an immutable inventory movement; catalog cost history remains authoritative.">
        <div class="grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <FormField class="gap-1">
            <span class="text-xs">Quantity delta</span>
            <AppInput v-model="adjustmentQuantity" class="input w-full min-w-0" placeholder="−1 or 2.5" inputmode="decimal" />
            <small class="text-xs text-base-content/60">Positive adds stock; negative records a correction or return.</small>
          </FormField>
          <FormField class="gap-1">
            <span class="text-xs">Unit cost ({{ props.currencyUnit }})</span>
            <AppInput
              :model-value="adjustmentCost"
              :money="props.currencyUnit"
              class="input w-full min-w-0"
              inputmode="numeric"
              @update:model-value="updateAdjustmentCost"
            />
          </FormField>
          <FormField class="gap-1 sm:col-span-2 xl:col-span-1">
            <span class="text-xs">Reason / note</span>
            <AppInput v-model="adjustmentNote" class="input w-full min-w-0" placeholder="Count correction" />
          </FormField>
        </div>
        <div class="flex justify-end border-t border-base-300 pt-3">
          <button class="btn btn-primary gap-2" type="button" :disabled="busy" @click="adjustStock">
            <RefreshCw :size="15" aria-hidden="true" />Record movement
          </button>
        </div>
      </AppPanel>

      <AppPanel title="Movement history" subtitle="Inventory ledger history for this material.">
        <LoadingState v-if="isMovementsLoading" label="Loading movements…" />
        <DataTable v-else-if="movements.length" label="Material movement history">
          <thead>
            <tr>
              <th scope="col">Date / movement</th>
              <th scope="col" class="text-end">Quantity</th>
              <th scope="col" class="text-end">Unit cost</th>
              <th scope="col" class="text-end">Total value</th>
              <th scope="col">Note / source</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="movement in movements" :key="movement.id">
              <DataTableCell>
                <strong class="block">{{ movement.movementType }}</strong>
                <span class="block text-xs text-base-content/60">{{ dateLabel(movement.occurredAt) }}</span>
              </DataTableCell>
              <DataTableCell numeric :class="String(movement.quantityDelta).startsWith('-') ? 'text-error' : 'text-success'">
                {{ movement.quantityDelta }} {{ selectedMaterial.consumptionUnit }}
              </DataTableCell>
              <DataTableCell numeric>{{ formatMoney(movement.unitCostRial, props.currencyUnit) }}</DataTableCell>
              <DataTableCell numeric>{{ formatMoney(movement.totalCostRial, props.currencyUnit) }}</DataTableCell>
              <DataTableCell>
                <span class="block max-w-64 whitespace-normal break-words">{{ movement.note || movement.referenceType || '—' }}</span>
                <span v-if="movement.referenceType && movement.note" class="block text-xs text-base-content/55">{{ movement.referenceType }}</span>
              </DataTableCell>
            </tr>
          </tbody>
        </DataTable>
        <EmptyState v-else compact title="No inventory movements" description="Stock receipts, usage, and adjustments will appear here."><template #icon><Package :size="21" aria-hidden="true" /></template></EmptyState>
      </AppPanel>

    </template>
  </div>
</template>
