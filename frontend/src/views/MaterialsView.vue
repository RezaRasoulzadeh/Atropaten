<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  Archive,
  Check,
  Edit3,
  Package,
  Plus,
  RotateCcw,
  Save,
  Search,
  X,
} from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import { materialsApi, type MaterialPayload, type MaterialRecord } from '../api/materials'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../utils/currency'
import { formatDateTime } from '../utils/date'

const props = defineProps<{ currencyUnit: CurrencyUnit }>()
const emit = defineEmits<{ notify: [message: string] }>()

type MaterialFilter = 'Active' | 'Archived' | 'All'
type EditorMode = 'create' | 'edit' | null
type MaterialForm = Omit<MaterialPayload, 'averageUnitCostRial'> & { averageUnitCostRial: number }

const materials = ref<MaterialRecord[]>([])
const selectedId = ref<string | null>(null)
const searchQuery = ref('')
const materialFilter = ref<MaterialFilter>('Active')
const editorMode = ref<EditorMode>(null)
const form = ref<MaterialForm>(emptyForm())
const costDraft = ref('0')
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const formError = ref('')

const unitOptions = ['piece', 'sheet', 'pack', 'kilogram', 'gram', 'roll', 'meter', 'liter', 'milliliter', 'square meter']

const selectedMaterial = computed(() => materials.value.find((material) => material.id === selectedId.value) ?? null)
const filteredMaterials = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return materials.value.filter((material) => {
    const matchesFilter = materialFilter.value === 'All' || (materialFilter.value === 'Active' ? material.active : !material.active)
    const matchesSearch = !query || [material.name, material.sku, material.category, material.purchaseUnit, material.consumptionUnit].some((value) => value.toLowerCase().includes(query))
    return matchesFilter && matchesSearch
  })
})

watch(() => props.currencyUnit, () => {
  costDraft.value = formatMoneyInput(form.value.averageUnitCostRial, props.currencyUnit)
})

onMounted(loadMaterials)

function emptyForm(): MaterialForm {
  return {
    name: '', sku: '', category: '', purchaseUnit: 'pack', consumptionUnit: 'sheet',
    conversionFactor: '500', physicalStock: '0', reorderLevel: '0', averageUnitCostRial: 0,
    preferredSupplier: '', notes: '',
  }
}

async function loadMaterials() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    materials.value = await materialsApi.list(true)
    if (!selectedId.value && materials.value.length) selectedId.value = materials.value[0].id
  } catch (error) {
    errorMessage.value = errorMessageFrom(error, 'Materials could not be loaded.')
  } finally {
    isLoading.value = false
  }
}

function selectMaterial(id: string) {
  selectedId.value = id
  editorMode.value = null
  formError.value = ''
}

function startCreate() {
  editorMode.value = 'create'
  selectedId.value = null
  form.value = emptyForm()
  costDraft.value = formatMoneyInput(0, props.currencyUnit)
  formError.value = ''
}

function startEdit() {
  const material = selectedMaterial.value
  if (!material) return
  form.value = {
    name: material.name, sku: material.sku, category: material.category,
    purchaseUnit: material.purchaseUnit, consumptionUnit: material.consumptionUnit,
    conversionFactor: material.conversionFactor, physicalStock: material.physicalStock,
    reorderLevel: material.reorderLevel, averageUnitCostRial: material.averageUnitCostRial,
    preferredSupplier: material.preferredSupplier, notes: material.notes,
  }
  costDraft.value = formatMoneyInput(material.averageUnitCostRial, props.currencyUnit)
  editorMode.value = 'edit'
  formError.value = ''
}

function cancelEditor() {
  editorMode.value = null
  formError.value = ''
}

function updateCost(value: string) {
  costDraft.value = value
  const parsed = parseMoneyInput(value, props.currencyUnit)
  if (parsed !== null) form.value.averageUnitCostRial = parsed
}

function onCostInput(event: Event) {
  updateCost((event.target as HTMLInputElement).value)
}

async function saveMaterial() {
  formError.value = ''
  const averageUnitCostRial = parseMoneyInput(costDraft.value, props.currencyUnit)
  if (averageUnitCostRial === null) {
    formError.value = `Enter a whole ${props.currencyUnit.toLowerCase()} amount.`
    return
  }
  form.value.averageUnitCostRial = averageUnitCostRial
  isSaving.value = true
  const wasEditing = editorMode.value === 'edit'
  try {
    const saved = editorMode.value === 'edit' && selectedId.value
      ? await materialsApi.update(selectedId.value, payload())
      : await materialsApi.create(payload())
    const existingIndex = materials.value.findIndex((material) => material.id === saved.id)
    if (existingIndex >= 0) materials.value.splice(existingIndex, 1, saved)
    else materials.value.push(saved)
    selectedId.value = saved.id
    editorMode.value = null
    emit('notify', wasEditing ? 'Material updated.' : 'Material created.')
  } catch (error) {
    formError.value = errorMessageFrom(error, 'Material could not be saved.')
  } finally {
    isSaving.value = false
  }
}

function payload(): MaterialPayload {
  return { ...form.value }
}

async function setActive(active: boolean) {
  const material = selectedMaterial.value
  if (!material) return
  try {
    const updated = active ? await materialsApi.reactivate(material.id) : await materialsApi.archive(material.id)
    const index = materials.value.findIndex((item) => item.id === updated.id)
    if (index >= 0) materials.value.splice(index, 1, updated)
    emit('notify', active ? 'Material reactivated.' : 'Material archived.')
  } catch (error) {
    errorMessage.value = errorMessageFrom(error, 'Material status could not be changed.')
  }
}

function errorMessageFrom(error: unknown, fallback: string): string {
  return error instanceof Error && error.message ? error.message : typeof error === 'string' ? error : fallback
}

function unitLabel(unit: string) {
  return unit.replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function dateLabel(value: string) {
  try {
    return formatDateTime(value)
  } catch {
    return 'Unknown date'
  }
}
</script>

<template>
  <div class="materials-view">
    <WorkspaceStickyStack>
      <header class="workspace-heading materials-heading">
        <div>
          <p class="eyebrow">Catalog / purchasing foundation</p>
          <h1>Materials</h1>
          <p class="heading-description">Keep physical stock, conversion units, and cost basis ready for production.</p>
        </div>
        <button class="button button-primary" type="button" @click="startCreate">
          <Plus class="button-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />
          New material
        </button>
      </header>

      <section class="materials-filter-bar panel" aria-label="Material filters">
        <label class="materials-search">
          <span class="sr-only">Search materials</span>
          <Search :size="16" :stroke-width="1.8" aria-hidden="true" />
          <input v-model="searchQuery" type="search" placeholder="Search material, SKU, or category" autocomplete="off" />
        </label>
        <label class="filter-control materials-status-filter">
          <span>Status</span>
          <span class="select-control"><select v-model="materialFilter" aria-label="Filter materials by status"><option>Active</option><option>Archived</option><option>All</option></select></span>
        </label>
        <span class="filter-result">{{ filteredMaterials.length }} of {{ materials.length }} materials</span>
      </section>
    </WorkspaceStickyStack>

    <div v-if="errorMessage" class="materials-error" role="alert">
      <span>{{ errorMessage }}</span>
      <button class="icon-button" type="button" aria-label="Dismiss materials error" @click="errorMessage = ''"><X :size="15" :stroke-width="1.8" aria-hidden="true" /></button>
    </div>

    <section class="materials-layout" aria-label="Materials workspace">
      <SectionPanel title="Material register" subtitle="Select a material to inspect its stock and cost basis." class="materials-table-panel">
        <template #action>
          <span class="materials-table-count">{{ filteredMaterials.length }} shown</span>
        </template>
        <div v-if="isLoading" class="materials-empty"><Package :size="21" :stroke-width="1.8" aria-hidden="true" /><p>Loading materials…</p></div>
        <div v-else-if="filteredMaterials.length" class="table-wrap materials-table-wrap">
          <table class="data-table materials-table">
            <thead>
              <tr><th scope="col">Material</th><th scope="col">Units</th><th scope="col">Physical stock</th><th scope="col">Available</th><th scope="col" class="numeric-column">Average cost</th><th scope="col">Reorder</th><th scope="col">State</th></tr>
            </thead>
            <tbody>
              <tr v-for="material in filteredMaterials" :key="material.id" :class="{ 'is-selected': selectedId === material.id }" tabindex="0" @click="selectMaterial(material.id)" @keydown.enter="selectMaterial(material.id)">
                <td><span class="table-primary">{{ material.name }}</span><span class="table-secondary">{{ material.sku || material.category || 'No SKU or category' }}</span></td>
                <td><span class="table-primary">{{ unitLabel(material.purchaseUnit) }}</span><span class="table-secondary">1 = {{ material.conversionFactor }} {{ material.consumptionUnit }}</span></td>
                <td><span class="table-primary">{{ material.physicalStock }} {{ material.consumptionUnit }}</span><span class="table-secondary">Current physical</span></td>
                <td><span class="table-primary">{{ material.physicalStock }} {{ material.consumptionUnit }}</span><span class="table-secondary">No reservations</span></td>
                <td class="numeric-column table-money">{{ formatMoney(material.averageUnitCostRial, props.currencyUnit) }}<span class="table-secondary">per {{ material.consumptionUnit }}</span></td>
                <td>{{ material.reorderLevel }} {{ material.consumptionUnit }}</td>
                <td><StatusBadge v-if="material.lowStock" label="Low stock" tone="amber" /><StatusBadge v-else :label="material.active ? 'Healthy' : 'Archived'" :tone="material.active ? 'green' : 'slate'" /></td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="materials-empty materials-empty-register">
          <div class="empty-workspace-icon" aria-hidden="true"><Package :size="21" :stroke-width="1.8" /></div>
          <h2>{{ materials.length ? 'No materials match this view' : 'No materials yet' }}</h2>
          <p>{{ materials.length ? 'Try another status or search term.' : 'Create the first material to establish a production stock baseline.' }}</p>
          <button v-if="!materials.length" class="button button-secondary" type="button" @click="startCreate"><Plus class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Create material</button>
        </div>
      </SectionPanel>

      <SectionPanel v-if="editorMode" :title="editorMode === 'create' ? 'New material' : 'Edit material'" subtitle="Validation and persistence run in the application layer." class="material-inspector material-editor">
        <template #action><button class="icon-button" type="button" aria-label="Close material editor" @click="cancelEditor"><X :size="16" :stroke-width="1.8" aria-hidden="true" /></button></template>
        <form class="material-form" @submit.prevent="saveMaterial">
          <div v-if="formError" class="form-error" role="alert">{{ formError }}</div>
          <label class="form-field form-field-wide"><span>Name</span><input v-model="form.name" type="text" placeholder="A4 80gsm Paper" autocomplete="off" /></label>
          <div class="material-form-grid">
            <label class="form-field"><span>SKU / code</span><input v-model="form.sku" type="text" placeholder="PAPER-A4" autocomplete="off" /></label>
            <label class="form-field"><span>Category</span><input v-model="form.category" type="text" placeholder="Paper" autocomplete="off" /></label>
          </div>
          <div class="material-form-grid">
            <label class="form-field"><span>Purchase unit</span><select v-model="form.purchaseUnit"><option v-for="unit in unitOptions" :key="unit" :value="unit">{{ unitLabel(unit) }}</option></select></label>
            <label class="form-field"><span>Consumption unit</span><select v-model="form.consumptionUnit"><option v-for="unit in unitOptions" :key="unit" :value="unit">{{ unitLabel(unit) }}</option></select></label>
          </div>
          <label class="form-field form-field-wide"><span>Conversion factor</span><input v-model="form.conversionFactor" type="text" inputmode="decimal" placeholder="500" /><small class="field-help">1 {{ form.purchaseUnit }} = {{ form.conversionFactor || '…' }} {{ form.consumptionUnit }}</small></label>
          <div class="material-form-grid">
            <label class="form-field"><span>Physical stock</span><input v-model="form.physicalStock" type="text" inputmode="decimal" placeholder="0" /></label>
            <label class="form-field"><span>Reorder level</span><input v-model="form.reorderLevel" type="text" inputmode="decimal" placeholder="0" /></label>
          </div>
          <label class="form-field form-field-wide"><span>Average cost / {{ form.consumptionUnit }} ({{ props.currencyUnit }})</span><input :value="costDraft" type="text" inputmode="decimal" placeholder="0" @input="onCostInput" /><small class="field-help">Stored as integer Rial; display follows the toolbar preference.</small></label>
          <label class="form-field form-field-wide"><span>Preferred supplier <em>optional</em></span><input v-model="form.preferredSupplier" type="text" placeholder="Pars Paper" autocomplete="off" /></label>
          <label class="form-field form-field-wide"><span>Notes <em>optional</em></span><textarea v-model="form.notes" rows="3" placeholder="Storage or handling note"></textarea></label>
          <div class="material-form-actions"><button class="button button-secondary" type="button" @click="cancelEditor">Cancel</button><button class="button button-primary" type="submit" :disabled="isSaving"><Save class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save material' }}</button></div>
        </form>
      </SectionPanel>

      <SectionPanel v-else-if="selectedMaterial" title="Material inspector" subtitle="Current persisted record" class="material-inspector">
        <template #action><button class="icon-button" type="button" aria-label="Edit selected material" @click="startEdit"><Edit3 :size="15" :stroke-width="1.8" aria-hidden="true" /></button></template>
        <div class="inspector-status"><StatusBadge :label="selectedMaterial.active ? 'Active' : 'Archived'" :tone="selectedMaterial.active ? 'green' : 'slate'" /><StatusBadge v-if="selectedMaterial.lowStock" label="Low stock" tone="amber" /></div>
        <div class="inspector-heading"><div class="material-inspector-icon"><Package :size="19" :stroke-width="1.8" aria-hidden="true" /></div><div><h3>{{ selectedMaterial.name }}</h3><p>{{ selectedMaterial.sku || 'No SKU' }}<span v-if="selectedMaterial.category"> · {{ selectedMaterial.category }}</span></p></div></div>
        <dl class="inspector-details">
          <div><dt>Unit conversion</dt><dd>1 {{ selectedMaterial.purchaseUnit }} = {{ selectedMaterial.conversionFactor }} {{ selectedMaterial.consumptionUnit }}</dd></div>
          <div><dt>Physical stock</dt><dd>{{ selectedMaterial.physicalStock }} {{ selectedMaterial.consumptionUnit }}</dd></div>
          <div><dt>Available / physical</dt><dd>{{ selectedMaterial.physicalStock }} {{ selectedMaterial.consumptionUnit }} <small>No reservations yet</small></dd></div>
          <div><dt>Reorder level</dt><dd>{{ selectedMaterial.reorderLevel }} {{ selectedMaterial.consumptionUnit }}</dd></div>
          <div><dt>Average cost</dt><dd>{{ formatMoney(selectedMaterial.averageUnitCostRial, props.currencyUnit) }} <small>per {{ selectedMaterial.consumptionUnit }}</small></dd></div>
          <div><dt>Last updated</dt><dd>{{ dateLabel(selectedMaterial.updatedAt) }}</dd></div>
        </dl>
        <p v-if="selectedMaterial.notes" class="inspector-note">{{ selectedMaterial.notes }}</p>
        <div class="inspector-actions"><button v-if="selectedMaterial.active" class="button button-secondary" type="button" @click="setActive(false)"><Archive class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Archive</button><button v-else class="button button-secondary" type="button" @click="setActive(true)"><RotateCcw class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Reactivate</button></div>
      </SectionPanel>

      <SectionPanel v-else title="Material inspector" subtitle="Select a row to inspect it." class="material-inspector material-inspector-empty">
        <div class="inspector-empty-copy"><Check :size="20" :stroke-width="1.8" aria-hidden="true" /><p>Material details will appear here.</p><button class="text-button" type="button" @click="startCreate">Create a material <Plus :size="14" :stroke-width="1.8" aria-hidden="true" /></button></div>
      </SectionPanel>
    </section>
  </div>
</template>
