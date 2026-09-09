<script setup lang="ts">
import { nextTick, watch } from 'vue';
import { Package, Plus } from 'lucide-vue-next';
import EmptyState from '../../components/ui/EmptyState.vue';
import LoadingState from '../../components/ui/LoadingState.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import SelectField from '../../components/ui/SelectField.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import MaterialWorkspaceView from './MaterialWorkspaceView.vue';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { useMaterialsWorkspace } from './useMaterialsWorkspace';

const props = defineProps<{ currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [message: string] }>();
const workspace = useMaterialsWorkspace(props, emit);

const {
  materials,
  selectedId,
  selectedMaterial,
  searchQuery,
  materialFilter,
  editorMode,
  isLoading,
  filteredMaterials,
  startCreate,
  selectMaterial,
  unitLabel,
} = workspace;

watch(
  [selectedId, editorMode],
  () => {
    void nextTick(() => {
      document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' });
    });
  },
);
</script>

<template>
  <div v-if="!selectedMaterial && !editorMode" class="min-w-0 space-y-3">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        title="Materials"
        eyebrow="Catalog / purchasing foundation"
        description="Keep physical stock, conversion units, and cost basis ready for production."
      >
        <button class="btn btn-primary" type="button" @click="startCreate">
          <Plus :size="16" :stroke-width="1.8" aria-hidden="true" />New material
        </button>
      </WorkspaceHeader>
      <SearchFilterBar>
        <template #search>
          <SearchField
            v-model="searchQuery"
            label="Search materials"
            placeholder="Search material, SKU, or category"
          />
        </template>
        <template #filters>
          <SelectField
            v-model="materialFilter"
            label="Status"
            aria-label="Filter materials by status"
            :options="['Active', 'Archived', 'All'].map((value) => ({ label: value, value }))"
          />
        </template>
        <template #count>
          <span class="whitespace-nowrap">{{ filteredMaterials.length }} of {{ materials.length }} materials</span>
        </template>
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <RegisterList
      title="Material register"
      subtitle="Select a material to inspect stock, cost basis, and movement history."
      :count="filteredMaterials.length"
    >
      <LoadingState v-if="isLoading" label="Loading materials…" />
      <EmptyState
        v-else-if="!filteredMaterials.length"
        :title="materials.length ? 'No materials match this view' : 'No materials yet'"
        :description="
          materials.length
            ? 'Try another status or search term.'
            : 'Create the first material to establish a production stock baseline.'
        "
      >
        <template #icon><Package :size="21" :stroke-width="1.8" aria-hidden="true" /></template>
        <template v-if="!materials.length" #action>
          <button class="btn btn-ghost" type="button" @click="startCreate">
            <Plus :size="15" :stroke-width="1.8" aria-hidden="true" />Create material
          </button>
        </template>
      </EmptyState>
      <template v-else>
        <RegisterRow
          v-for="material in filteredMaterials"
          :key="material.id"
          :selected="selectedId === material.id"
          @activate="selectMaterial(material.id)"
        >
          <template #icon><Package :size="17" :stroke-width="1.8" aria-hidden="true" /></template>
          <template #identity>
            <div class="flex min-w-0 items-start justify-between gap-3">
              <div class="min-w-0">
                <strong class="block truncate text-sm">{{ material.name }}</strong>
                <span class="block truncate text-xs text-base-content/60">
                  {{ material.sku || 'No SKU' }}<span v-if="material.category"> · {{ material.category }}</span>
                </span>
              </div>
              <div class="flex shrink-0 items-center gap-1.5">
                <StatusBadge v-if="material.lowStock" label="Low stock" tone="amber" />
                <StatusBadge :label="material.active ? 'Active' : 'Archived'" :tone="material.active ? 'green' : 'slate'" />
              </div>
            </div>
          </template>
          <template #meta>
            <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2 xl:grid-cols-4">
              <div class="min-w-0">
                <span class="block text-base-content/50">Stock</span>
                <span class="block text-base-content/80 tabular-nums">
                  {{ material.physicalStock }} physical · {{ material.availableStock }} available · {{ material.reservedStock }} reserved {{ material.consumptionUnit }}
                </span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Cost / value</span>
                <span class="block text-base-content/80 tabular-nums">
                  {{ formatMoney(material.averageUnitCostRial, props.currencyUnit) }} per {{ material.consumptionUnit }} · {{ formatMoney(material.inventoryValueRial, props.currencyUnit) }} total
                </span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Units</span>
                <span class="block text-base-content/80">1 {{ unitLabel(material.purchaseUnit) }} = {{ material.conversionFactor }} {{ unitLabel(material.consumptionUnit) }}</span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Reorder at</span>
                <span class="block text-base-content/80 tabular-nums">{{ material.reorderLevel }} {{ material.consumptionUnit }}</span>
              </div>
            </div>
          </template>
        </RegisterRow>
      </template>
    </RegisterList>
  </div>

  <MaterialWorkspaceView
    v-else
    :workspace="workspace"
    :currency-unit="props.currencyUnit"
    @back="workspace.backToMaterials"
  />
</template>
