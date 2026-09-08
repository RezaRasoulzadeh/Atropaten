<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import MasterDetail from '../../components/layout/MasterDetail.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import FormField from '../../components/ui/FormField.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import { computed, onMounted, ref, watch } from 'vue';
import {
  Archive,
  Check,
  Edit3,
  Package,
  Plus,
  RotateCcw,
  Save,
  X,
  RefreshCw,
} from 'lucide-vue-next';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import DataTable from '../../components/ui/DataTable.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import InspectorSection from '../../components/layout/InspectorSection.vue';
import FormSection from '../../components/ui/FormSection.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import LoadingState from '../../components/ui/LoadingState.vue';
import InlineAlert from '../../components/ui/InlineAlert.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import { materialsApi, type MaterialPayload, type MaterialRecord } from '../../api/materials';
import { purchasesApi } from '../../api/purchases';
import {
  formatMoney,
  formatMoneyInput,
  parseMoneyInput,
  type CurrencyUnit,
} from '../../utils/currency';
import { formatDateTime } from '../../utils/date';

const props = defineProps<{ currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [message: string] }>();

type MaterialFilter = 'Active' | 'Archived' | 'All';
type EditorMode = 'create' | 'edit' | null;
type MaterialForm = Omit<MaterialPayload, 'averageUnitCostRial'> & { averageUnitCostRial: number };

const materials = ref<MaterialRecord[]>([]);
const selectedId = ref<string | null>(null);
const searchQuery = ref('');
const materialFilter = ref<MaterialFilter>('Active');
const editorMode = ref<EditorMode>(null);
const form = ref<MaterialForm>(emptyForm());
const costDraft = ref('0');
const isLoading = ref(false);
const isSaving = ref(false);
const errorMessage = ref('');
const formError = ref('');
const movements = ref<any[]>([]);
const adjustmentQuantity = ref('');
const adjustmentCost = ref('0');
const adjustmentNote = ref('');

const unitOptions = [
  'piece',
  'sheet',
  'pack',
  'kilogram',
  'gram',
  'roll',
  'meter',
  'liter',
  'milliliter',
  'square meter',
];

const selectedMaterial = computed(
  () => materials.value.find((material) => material.id === selectedId.value) ?? null,
);
const filteredMaterials = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  return materials.value.filter((material) => {
    const matchesFilter =
      materialFilter.value === 'All' ||
      (materialFilter.value === 'Active' ? material.active : !material.active);
    const matchesSearch =
      !query ||
      [
        material.name,
        material.sku,
        material.category,
        material.purchaseUnit,
        material.consumptionUnit,
      ].some((value) => value.toLowerCase().includes(query));
    return matchesFilter && matchesSearch;
  });
});

watch(
  () => props.currencyUnit,
  () => {
    costDraft.value = formatMoneyInput(form.value.averageUnitCostRial, props.currencyUnit);
  },
);

onMounted(loadMaterials);

function emptyForm(): MaterialForm {
  return {
    name: '',
    sku: '',
    category: '',
    purchaseUnit: 'pack',
    consumptionUnit: 'sheet',
    conversionFactor: '500',
    physicalStock: '0',
    reorderLevel: '0',
    averageUnitCostRial: 0,
    preferredSupplier: '',
    notes: '',
  };
}

async function loadMaterials() {
  isLoading.value = true;
  errorMessage.value = '';
  try {
    materials.value = await materialsApi.list(true);
    if (!selectedId.value && materials.value.length) selectedId.value = materials.value[0].id;
  } catch (error) {
    errorMessage.value = errorMessageFrom(error, 'Materials could not be loaded.');
  } finally {
    isLoading.value = false;
  }
}

function selectMaterial(id: string) {
  selectedId.value = id;
  editorMode.value = null;
  formError.value = '';
  loadMovements(id);
}

async function loadMovements(id = selectedId.value ?? '') {
  if (!id) {
    movements.value = [];
    return;
  }
  try {
    movements.value = await purchasesApi.movements(id);
  } catch (error) {
    errorMessage.value = errorMessageFrom(error, 'Movement history could not be loaded.');
  }
}

async function adjustStock() {
return runAction(async () => {
  if (!selectedMaterial.value || !adjustmentQuantity.value.trim()) return;
  const cost = parseMoneyInput(adjustmentCost.value, props.currencyUnit);
  if (cost === null) {
    formError.value = 'Enter a whole adjustment cost.';
    return;
  }
  try {
    await purchasesApi.adjust(
      selectedMaterial.value.id,
      adjustmentQuantity.value,
      cost,
      adjustmentNote.value,
    );
    adjustmentQuantity.value = '';
    adjustmentNote.value = '';
    await loadMaterials();
    await loadMovements(selectedMaterial.value.id);
    emit('notify', 'Stock adjustment recorded as an immutable movement.');
  } catch (error) {
reportError(error);
    formError.value = errorMessageFrom(error, 'Stock adjustment could not be recorded.');
  }

});
}

function startCreate() {
  editorMode.value = 'create';
  selectedId.value = null;
  form.value = emptyForm();
  costDraft.value = formatMoneyInput(0, props.currencyUnit);
  formError.value = '';
}

function startEdit() {
  const material = selectedMaterial.value;
  if (!material) return;
  form.value = {
    name: material.name,
    sku: material.sku,
    category: material.category,
    purchaseUnit: material.purchaseUnit,
    consumptionUnit: material.consumptionUnit,
    conversionFactor: material.conversionFactor,
    physicalStock: material.physicalStock,
    reorderLevel: material.reorderLevel,
    averageUnitCostRial: material.averageUnitCostRial,
    preferredSupplier: material.preferredSupplier,
    notes: material.notes,
  };
  costDraft.value = formatMoneyInput(material.averageUnitCostRial, props.currencyUnit);
  editorMode.value = 'edit';
  formError.value = '';
}

function cancelEditor() {
  editorMode.value = null;
  formError.value = '';
}

function updateCost(value: string) {
  const parsed = parseMoneyInput(value, props.currencyUnit);
  if (parsed !== null) {
    form.value.averageUnitCostRial = parsed;
    costDraft.value = formatMoneyInput(parsed, props.currencyUnit);
  } else {
    costDraft.value = value;
  }
}

function onCostInput(value: string) {
  updateCost(value);
}

async function saveMaterial() {
return runAction(async () => {
  formError.value = '';
  const averageUnitCostRial = parseMoneyInput(costDraft.value, props.currencyUnit);
  if (averageUnitCostRial === null) {
    formError.value = `Enter a whole ${props.currencyUnit.toLowerCase()} amount.`;
    return;
  }
  form.value.averageUnitCostRial = averageUnitCostRial;
  isSaving.value = true;
  const wasEditing = editorMode.value === 'edit';
  try {
    const saved =
      editorMode.value === 'edit' && selectedId.value
        ? await materialsApi.update(selectedId.value, payload())
        : await materialsApi.create(payload());
    const existingIndex = materials.value.findIndex((material) => material.id === saved.id);
    if (existingIndex >= 0) materials.value.splice(existingIndex, 1, saved);
    else materials.value.push(saved);
    selectedId.value = saved.id;
    editorMode.value = null;
    emit('notify', wasEditing ? 'Material updated.' : 'Material created.');
  } catch (error) {
reportError(error);
    formError.value = errorMessageFrom(error, 'Material could not be saved.');
  } finally {
    isSaving.value = false;
  }

});
}

function payload(): MaterialPayload {
  return { ...form.value };
}

async function setActive(active: boolean) {
return runAction(async () => {
  const material = selectedMaterial.value;
  if (!material) return;
  try {
    const updated = active
      ? await materialsApi.reactivate(material.id)
      : await materialsApi.archive(material.id);
    const index = materials.value.findIndex((item) => item.id === updated.id);
    if (index >= 0) materials.value.splice(index, 1, updated);
    emit('notify', active ? 'Material reactivated.' : 'Material archived.');
  } catch (error) {
reportError(error);
    errorMessage.value = errorMessageFrom(error, 'Material status could not be changed.');
  }

});
}

function errorMessageFrom(error: unknown, fallback: string): string {
  return error instanceof Error && error.message
    ? error.message
    : typeof error === 'string'
      ? error
      : fallback;
}

function unitLabel(unit: string) {
  return unit.replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function dateLabel(value: string) {
  try {
    return formatDateTime(value);
  } catch {
    return 'Unknown date';
  }
}
</script>

<template>
  <div class="space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        eyebrow="Catalog / purchasing foundation"
        title="Materials"
        description="Keep physical stock, conversion units, and cost basis ready for production."
      >
        <button class="btn btn-primary" type="button" @click="startCreate">
          <Plus :size="16" :stroke-width="1.8" aria-hidden="true" />
          New material
        </button>
      </WorkspaceHeader>
      <SearchFilterBar>
        <template #search
          ><SearchField
            v-model="searchQuery"
            label="Search materials"
            placeholder="Search material, SKU, or category"
        /></template>
        <template #filters
          ><SelectField
            v-model="materialFilter"
            label="Status"
            aria-label="Filter materials by status"
            :options="['Active', 'Archived', 'All'].map((value) => ({ label: value, value }))"
        /></template>
        <template #count
          ><span>{{ filteredMaterials.length }} of {{ materials.length }} materials</span></template
        >
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <InlineAlert v-if="errorMessage" :message="errorMessage" @click="errorMessage = ''"
      ><template #action><X :size="15" :stroke-width="1.8" aria-hidden="true" /></template
    ></InlineAlert>

    <MasterDetail>
      <RegisterList
        title="Material register"
        subtitle="Scan stock, value and reorder state before opening the material inspector."
        :count="filteredMaterials.length"
      >
        <LoadingState v-if="isLoading" label="Loading materials…" />
        <div v-else-if="filteredMaterials.length">
          <RegisterRow
              v-for="material in filteredMaterials"
              :key="material.id"
              :selected="selectedId === material.id"
              @activate="selectMaterial(material.id)"
            >
              <template #identity>
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <strong class="block truncate text-sm">{{ material.name }}</strong>
                    <span class="block truncate text-xs text-base-content/60">
                      {{ material.sku || material.category || 'No SKU or category' }}
                    </span>
                  </div>
                  <strong class="shrink-0 whitespace-nowrap text-sm tabular-nums">
                    {{ material.availableStock }} {{ material.consumptionUnit }}
                  </strong>
                </div>
              </template>
              <template #meta>
                <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2">
                  <div>
                    <span class="block text-base-content/50">Stock</span>
                    <span class="block text-base-content/80 tabular-nums">
                      {{ material.physicalStock }} physical · {{ material.reservedStock }} reserved
                    </span>
                  </div>
                  <div>
                    <span class="block text-base-content/50">Cost / value</span>
                    <span class="block text-base-content/80 tabular-nums">
                      {{ formatMoney(material.averageUnitCostRial, props.currencyUnit) }} per {{ material.consumptionUnit }} · {{ formatMoney(material.inventoryValueRial, props.currencyUnit) }} total
                    </span>
                  </div>
                  <div>
                    <span class="block text-base-content/50">Units</span>
                    <span class="block text-base-content/80">{{ unitLabel(material.purchaseUnit) }} · 1 = {{ material.conversionFactor }} {{ material.consumptionUnit }}</span>
                  </div>
                  <div>
                    <span class="block text-base-content/50">Reorder at</span>
                    <span class="block text-base-content/80 tabular-nums">{{ material.reorderLevel }} {{ material.consumptionUnit }}</span>
                  </div>
                </div>
              </template>
              <template #status>
                <StatusBadge v-if="material.lowStock" label="Low stock" tone="amber" />
                <StatusBadge
                  :label="material.active ? 'Healthy' : 'Archived'"
                  :tone="material.active ? 'green' : 'slate'"
                />
              </template>
          </RegisterRow>
        </div>
        <EmptyState
          v-else
          :title="materials.length ? 'No materials match this view' : 'No materials yet'"
          :description="
            materials.length
              ? 'Try another status or search term.'
              : 'Create the first material to establish a production stock baseline.'
          "
        >
          <template #icon><Package :size="21" :stroke-width="1.8" aria-hidden="true" /></template>
          <template v-if="!materials.length" #action
            ><button class="btn btn-ghost" type="button" @click="startCreate">
              <Plus :size="15" :stroke-width="1.8" aria-hidden="true" />Create material
            </button></template
          >
        </EmptyState>
      </RegisterList>

      <InspectorShell
        v-if="editorMode"
        :title="editorMode === 'create' ? 'New material' : 'Edit material'"
        subtitle="Catalog fields and opening stock configuration."
      >
        <template #header
          ><button
            class="btn btn-ghost btn-sm"
            type="button"
            aria-label="Close material editor"
            @click="cancelEditor"
          >
            <X :size="16" :stroke-width="1.8" aria-hidden="true" /></button
        ></template>
        <form class="space-y-3" @submit.prevent="saveMaterial">
          <InlineAlert v-if="formError" :message="formError" />
          <FormSection title="Identity">
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs">Name</span
              ><AppInput
                v-model="form.name"
                class="input w-full min-w-0"
                type="text"
                placeholder="A4 80gsm Paper"
                autocomplete="off"
            /></FormField>
            <FormField class="gap-1"
              ><span class="text-xs">SKU / code</span
              ><AppInput
                v-model="form.sku"
                class="input w-full min-w-0"
                type="text"
                placeholder="PAPER-A4"
                autocomplete="off"
            /></FormField>
            <FormField class="gap-1"
              ><span class="text-xs">Category</span
              ><AppInput
                v-model="form.category"
                class="input w-full min-w-0"
                type="text"
                placeholder="Paper"
                autocomplete="off"
            /></FormField>
          </FormSection>
          <FormSection title="Units and stock">
            <SelectField
              v-model="form.purchaseUnit"
              label="Purchase unit"
              :options="unitOptions.map((unit) => ({ label: unitLabel(unit), value: unit }))"
            />
            <SelectField
              v-model="form.consumptionUnit"
              label="Consumption unit"
              :options="unitOptions.map((unit) => ({ label: unitLabel(unit), value: unit }))"
            />
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs">Conversion factor</span
              ><AppInput
                v-model="form.conversionFactor"
                class="input w-full min-w-0"
                type="text"
                inputmode="decimal"
                placeholder="500"
              /><small class="text-xs text-base-content/60"
                >1 {{ form.purchaseUnit }} = {{ form.conversionFactor || '…' }}
                {{ form.consumptionUnit }}</small
              ></FormField
            >
            <FormField class="gap-1"
              ><span class="text-xs">Physical stock</span
              ><AppInput
                v-model="form.physicalStock"
                class="input w-full min-w-0"
                type="text"
                inputmode="decimal"
                placeholder="0"
                :disabled="editorMode === 'edit'"
              /><small v-if="editorMode === 'edit'" class="text-xs text-base-content/60"
                >Use Adjust stock for ledger movements.</small
              ></FormField
            >
            <FormField class="gap-1"
              ><span class="text-xs">Reorder level</span
              ><AppInput
                v-model="form.reorderLevel"
                class="input w-full min-w-0"
                type="text"
                inputmode="decimal"
                placeholder="0"
            /></FormField>
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs"
                >Average cost / {{ form.consumptionUnit }} ({{ props.currencyUnit }})</span
              ><AppInput
                :model-value="costDraft"
                type="text"
                inputmode="decimal"
                placeholder="0"
                :disabled="editorMode === 'edit'"
                @update:model-value="onCostInput"
              /><small class="text-xs text-base-content/60"
                >Stored as integer Rial; use opening stock to establish catalog cost.</small
              ></FormField
            >
          </FormSection>
          <FormSection title="Supplier and notes">
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs">Preferred supplier <em>optional</em></span
              ><AppInput
                v-model="form.preferredSupplier"
                class="input w-full min-w-0"
                type="text"
                placeholder="Pars Paper"
                autocomplete="off"
            /></FormField>
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs">Notes <em>optional</em></span
              ><AppTextarea
                v-model="form.notes"
                rows="3"
                placeholder="Storage or handling note"
              />
            </FormField>
          </FormSection>
        </form>
        <template #footer
          ><button class="btn btn-ghost" type="button" @click="cancelEditor">Cancel</button
          ><button class="btn btn-primary" type="button" :disabled="busy || (isSaving)" @click="saveMaterial">
            <Save :size="15" :stroke-width="1.8" aria-hidden="true" />{{
              isSaving ? 'Saving…' : 'Save material'
            }}
          </button></template
        >
      </InspectorShell>

      <InspectorShell
        v-else-if="selectedMaterial"
        title="Material inspector"
        subtitle="Current persisted record."
      >
        <template #header
          ><StatusBadge
            :label="selectedMaterial.active ? 'Active' : 'Archived'"
            :tone="selectedMaterial.active ? 'green' : 'slate'"
          /><StatusBadge v-if="selectedMaterial.lowStock" label="Low stock" tone="amber" /><button
            class="btn btn-ghost btn-sm"
            type="button"
            aria-label="Edit selected material"
            @click="startEdit"
          >
            <Edit3 :size="15" :stroke-width="1.8" aria-hidden="true" />Edit
          </button></template
        >
        <div class="space-y-4">
          <section
            class="flex items-start gap-3 rounded-box border border-base-300 bg-base-200/35 p-3"
          >
            <div class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200">
              <Package :size="19" :stroke-width="1.8" aria-hidden="true" />
            </div>
            <div class="min-w-0">
              <h3 class="truncate font-semibold">{{ selectedMaterial.name }}</h3>
              <p class="mt-1 text-xs text-base-content/60">
                {{ selectedMaterial.sku || 'No SKU'
                }}<span v-if="selectedMaterial.category"> · {{ selectedMaterial.category }}</span>
              </p>
            </div>
          </section>
          <InspectorSection title="Stock and cost">
          <dl class="grid gap-2 rounded-box border border-base-300 p-3 text-sm sm:grid-cols-2">
            <div>
              <dt class="text-xs text-base-content/60">Unit conversion</dt>
              <dd class="mt-1 font-medium">
                1 {{ selectedMaterial.purchaseUnit }} = {{ selectedMaterial.conversionFactor }}
                {{ selectedMaterial.consumptionUnit }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-base-content/60">Physical stock</dt>
              <dd class="mt-1 font-medium tabular-nums">
                {{ selectedMaterial.physicalStock }} {{ selectedMaterial.consumptionUnit }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-base-content/60">Reserved stock</dt>
              <dd class="mt-1 font-medium tabular-nums">
                {{ selectedMaterial.reservedStock }} {{ selectedMaterial.consumptionUnit }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-base-content/60">Available stock</dt>
              <dd class="mt-1 font-medium tabular-nums">
                {{ selectedMaterial.availableStock }} {{ selectedMaterial.consumptionUnit }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-base-content/60">Reorder level</dt>
              <dd class="mt-1 font-medium tabular-nums">
                {{ selectedMaterial.reorderLevel }} {{ selectedMaterial.consumptionUnit }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-base-content/60">Average cost</dt>
              <dd class="mt-1 font-medium tabular-nums">
                {{ formatMoney(selectedMaterial.averageUnitCostRial, props.currencyUnit) }}
                <span class="text-xs font-normal text-base-content/60"
                  >/ {{ selectedMaterial.consumptionUnit }}</span
                >
              </dd>
            </div>
            <div class="sm:col-span-2">
              <dt class="text-xs text-base-content/60">Last updated</dt>
              <dd class="mt-1 font-medium">{{ dateLabel(selectedMaterial.updatedAt) }}</dd>
            </div>
          </dl>
          </InspectorSection>
          <p v-if="selectedMaterial.preferredSupplier" class="text-sm">
            <span class="text-xs text-base-content/60">Preferred supplier:</span>
            {{ selectedMaterial.preferredSupplier }}
          </p>
          <p
            v-if="selectedMaterial.notes"
            class="whitespace-pre-wrap rounded-box border border-base-300 p-3 text-sm"
          >
            {{ selectedMaterial.notes }}
          </p>
          <section class="rounded-box border border-base-300 bg-base-200/40 p-3">
            <div class="mb-3">
              <h3 class="text-sm font-semibold">Adjust stock</h3>
              <p class="mt-1 text-xs text-base-content/60">
                Positive adds stock; negative records a supplier return or correction.
              </p>
            </div>
            <div class="grid gap-3 sm:grid-cols-2">
              <FormField class="gap-1"
                ><span class="text-xs">Quantity delta</span
                ><AppInput
                  v-model="adjustmentQuantity"
                  class="input w-full min-w-0"
                  placeholder="−1 or 2.5"
                  inputmode="decimal"
              /></FormField>
              <FormField class="gap-1"
                ><span class="text-xs">Unit cost ({{ props.currencyUnit }})</span
                ><AppInput
                  v-model="adjustmentCost"
                  class="input w-full min-w-0"
                  inputmode="numeric"
              /></FormField>
              <FormField class="gap-1 sm:col-span-2"
                ><span class="text-xs">Reason / note</span
                ><AppInput
                  v-model="adjustmentNote"
                  class="input w-full min-w-0"
                  placeholder="Count correction"
              /></FormField>
            </div>
            <button class="btn btn-primary mt-3" type="button" @click="adjustStock" :disabled="busy">
              <RefreshCw :size="15" />Record movement
            </button>
          </section>
          <InspectorSection title="Recent movements" description="Inventory ledger history">
            <DataTable v-if="movements.length"
              ><thead>
                <tr>
                  <th scope="col">Movement</th>
                  <th scope="col" class="text-end text-end text-end">Quantity</th>
                  <th scope="col" class="text-end">Cost</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-base-300">
                <tr v-for="movement in movements.slice(0, 6)" :key="movement.id">
                  <DataTableCell>
                    <strong class="block">{{ movement.movementType }}</strong
                    ><span
                      class="text-xs text-base-content/60 block text-xs text-base-content/60"
                      >{{ dateLabel(movement.occurredAt) }}</span
                    >
                  </DataTableCell>
                  <DataTableCell
                    class="text-end tabular-nums"
                    :class="movement.quantityDelta.startsWith('-') ? 'text-error' : ''"
                    numeric
                  >
                    {{ movement.quantityDelta }} {{ selectedMaterial.consumptionUnit }}
                  </DataTableCell>
                  <DataTableCell class="whitespace-nowrap text-end tabular-nums" numeric>
                    {{ formatMoney(movement.totalCostRial, props.currencyUnit) }}
                  </DataTableCell>
                </tr>
              </tbody></DataTable
            >
            <p v-else class="rounded-box border border-base-300 p-3 text-xs text-base-content/60">
              No inventory movements yet.
            </p>
          </InspectorSection>
        </div>
        <template #footer
          ><button
            v-if="selectedMaterial.active"
            class="btn btn-ghost"
            type="button"
            @click="setActive(false)"
           :disabled="busy">
            <Archive :size="15" :stroke-width="1.8" aria-hidden="true" />Archive</button
          ><button v-else class="btn btn-ghost" type="button" @click="setActive(true)" :disabled="busy">
            <RotateCcw :size="15" :stroke-width="1.8" aria-hidden="true" />Reactivate
          </button></template
        >
      </InspectorShell>

      <InspectorShell v-else title="Material inspector" subtitle="Select a row to inspect it.">
        <EmptyState
          title="No material selected"
          description="Material details, stock adjustment, and movement history will appear here."
        >
          <template #icon><Check :size="20" :stroke-width="1.8" aria-hidden="true" /></template>
          <template #action
            ><button class="btn btn-ghost" type="button" @click="startCreate">
              Create a material <Plus :size="14" :stroke-width="1.8" aria-hidden="true" /></button
          ></template>
        </EmptyState>
      </InspectorShell>
    </MasterDetail>
  </div>
</template>
