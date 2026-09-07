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
  RefreshCw,
} from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import { materialsApi, type MaterialPayload, type MaterialRecord } from '../api/materials'
import { purchasesApi } from '../api/purchases'
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
const movements = ref<any[]>([])
const adjustmentQuantity = ref('')
const adjustmentCost = ref('0')
const adjustmentNote = ref('')

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
  loadMovements(id)
}

async function loadMovements(id = selectedId.value ?? '') {
  if (!id) { movements.value = []; return }
  try { movements.value = await purchasesApi.movements(id) } catch (error) { errorMessage.value = errorMessageFrom(error, 'Movement history could not be loaded.') }
}

async function adjustStock() {
  if (!selectedMaterial.value || !adjustmentQuantity.value.trim()) return
  const cost = parseMoneyInput(adjustmentCost.value, props.currencyUnit)
  if (cost === null) { formError.value = 'Enter a whole adjustment cost.'; return }
  try {
    await purchasesApi.adjust(selectedMaterial.value.id, adjustmentQuantity.value, cost, adjustmentNote.value)
    adjustmentQuantity.value = ''; adjustmentNote.value = ''; await loadMaterials(); await loadMovements(selectedMaterial.value.id)
    emit('notify', 'Stock adjustment recorded as an immutable movement.')
  } catch (error) { formError.value = errorMessageFrom(error, 'Stock adjustment could not be recorded.') }
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
  const parsed = parseMoneyInput(value, props.currencyUnit)
  if (parsed !== null) {
    form.value.averageUnitCostRial = parsed
    costDraft.value = formatMoneyInput(parsed, props.currencyUnit)
  } else {
    costDraft.value = value
  }
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
  <div>
    <WorkspaceStickyStack>
      <header>
        <div>
          <p>Catalog / purchasing foundation</p>
          <h1>Materials</h1>
          <p>Keep physical stock, conversion units, and cost basis ready for production.</p>
        </div>
        <button class="btn btn-ghost" type="button" @click="startCreate">
          <Plus :size="16" :stroke-width="1.8" aria-hidden="true" />
          New material
        </button>
      </header>

      <section aria-label="Material filters">
        <label class="form-control gap-1">
          <span>Search materials</span>
          <Search :size="16" :stroke-width="1.8" aria-hidden="true" />
          <input class="input input-bordered w-full min-w-0" v-model="searchQuery" type="search" placeholder="Search material, SKU, or category" autocomplete="off" />
        </label>
        <label class="form-control gap-1">
          <span>Status</span>
          <span><select class="select select-bordered w-full min-w-0" v-model="materialFilter" aria-label="Filter materials by status"><option>Active</option><option>Archived</option><option>All</option></select></span>
        </label>
        <span>{{ filteredMaterials.length }} of {{ materials.length }} materials</span>
      </section>
    </WorkspaceStickyStack>

    <div v-if="errorMessage" role="alert">
      <span>{{ errorMessage }}</span>
      <button class="btn btn-ghost" type="button" aria-label="Dismiss materials error" @click="errorMessage = ''"><X :size="15" :stroke-width="1.8" aria-hidden="true" /></button>
    </div>

    <section aria-label="Materials workspace">
      <SectionPanel title="Material register" subtitle="Select a material to inspect its stock and cost basis.">
        <template #action>
          <span>{{ filteredMaterials.length }} shown</span>
        </template>
        <div v-if="isLoading"><Package :size="21" :stroke-width="1.8" aria-hidden="true" /><p>Loading materials…</p></div>
        <div v-else-if="filteredMaterials.length">
          <table class="table table-zebra w-full">
            <thead>
              <tr><th scope="col">Material</th><th scope="col">Units</th><th scope="col">Physical</th><th scope="col">Available</th><th scope="col">Average cost</th><th scope="col">Inventory value</th><th scope="col">Reorder</th><th scope="col">State</th></tr>
            </thead>
            <tbody>
              <tr v-for="material in filteredMaterials" :key="material.id" :class="{ 'bg-base-300': selectedId === material.id }" tabindex="0" @click="selectMaterial(material.id)" @keydown.enter="selectMaterial(material.id)">
                <td><span>{{ material.name }}</span><span>{{ material.sku || material.category || 'No SKU or category' }}</span></td>
                <td><span>{{ unitLabel(material.purchaseUnit) }}</span><span>1 = {{ material.conversionFactor }} {{ material.consumptionUnit }}</span></td>
                <td><span>{{ material.physicalStock }} {{ material.consumptionUnit }}</span><span>Current physical</span></td>
                <td><span>{{ material.availableStock }} {{ material.consumptionUnit }}</span><span>{{ material.reservedStock }} reserved</span></td>
                <td>{{ formatMoney(material.averageUnitCostRial, props.currencyUnit) }}<span>per {{ material.consumptionUnit }}</span></td>
                <td>{{ formatMoney(material.inventoryValueRial, props.currencyUnit) }}</td>
                <td>{{ material.reorderLevel }} {{ material.consumptionUnit }}</td>
                <td><StatusBadge v-if="material.lowStock" label="Low stock" tone="amber" /><StatusBadge v-else :label="material.active ? 'Healthy' : 'Archived'" :tone="material.active ? 'green' : 'slate'" /></td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else>
          <div aria-hidden="true"><Package :size="21" :stroke-width="1.8" /></div>
          <h2>{{ materials.length ? 'No materials match this view' : 'No materials yet' }}</h2>
          <p>{{ materials.length ? 'Try another status or search term.' : 'Create the first material to establish a production stock baseline.' }}</p>
          <button class="btn btn-ghost" v-if="!materials.length" type="button" @click="startCreate"><Plus :size="15" :stroke-width="1.8" aria-hidden="true" />Create material</button>
        </div>
      </SectionPanel>

      <SectionPanel v-if="editorMode" :title="editorMode === 'create' ? 'New material' : 'Edit material'" subtitle="Validation and persistence run in the application layer.">
        <template #action><button class="btn btn-ghost" type="button" aria-label="Close material editor" @click="cancelEditor"><X :size="16" :stroke-width="1.8" aria-hidden="true" /></button></template>
        <form @submit.prevent="saveMaterial">
          <div v-if="formError" role="alert">{{ formError }}</div>
          <label class="form-control gap-1"><span>Name</span><input class="input input-bordered w-full min-w-0" v-model="form.name" type="text" placeholder="A4 80gsm Paper" autocomplete="off" /></label>
          <div>
            <label class="form-control gap-1"><span>SKU / code</span><input class="input input-bordered w-full min-w-0" v-model="form.sku" type="text" placeholder="PAPER-A4" autocomplete="off" /></label>
            <label class="form-control gap-1"><span>Category</span><input class="input input-bordered w-full min-w-0" v-model="form.category" type="text" placeholder="Paper" autocomplete="off" /></label>
          </div>
          <div>
            <label class="form-control gap-1"><span>Purchase unit</span><select class="select select-bordered w-full min-w-0" v-model="form.purchaseUnit"><option v-for="unit in unitOptions" :key="unit" :value="unit">{{ unitLabel(unit) }}</option></select></label>
            <label class="form-control gap-1"><span>Consumption unit</span><select class="select select-bordered w-full min-w-0" v-model="form.consumptionUnit"><option v-for="unit in unitOptions" :key="unit" :value="unit">{{ unitLabel(unit) }}</option></select></label>
          </div>
          <label class="form-control gap-1"><span>Conversion factor</span><input class="input input-bordered w-full min-w-0" v-model="form.conversionFactor" type="text" inputmode="decimal" placeholder="500" /><small>1 {{ form.purchaseUnit }} = {{ form.conversionFactor || '…' }} {{ form.consumptionUnit }}</small></label>
          <div>
            <label class="form-control gap-1"><span>Physical stock</span><input class="input input-bordered w-full min-w-0" v-model="form.physicalStock" type="text" inputmode="decimal" placeholder="0" :disabled="editorMode === 'edit'" /><small v-if="editorMode === 'edit'">Use Adjust Stock to create a ledger movement.</small></label>
            <label class="form-control gap-1"><span>Reorder level</span><input class="input input-bordered w-full min-w-0" v-model="form.reorderLevel" type="text" inputmode="decimal" placeholder="0" /></label>
          </div>
          <label class="form-control gap-1"><span>Average cost / {{ form.consumptionUnit }} ({{ props.currencyUnit }})</span><input class="input input-bordered w-full min-w-0" :value="costDraft" type="text" inputmode="decimal" placeholder="0" @input="onCostInput" :disabled="editorMode === 'edit'" /><small>Stored as integer Rial; edit catalog cost only when establishing opening stock.</small></label>
          <label class="form-control gap-1"><span>Preferred supplier <em>optional</em></span><input class="input input-bordered w-full min-w-0" v-model="form.preferredSupplier" type="text" placeholder="Pars Paper" autocomplete="off" /></label>
          <label class="form-control gap-1"><span>Notes <em>optional</em></span><textarea class="textarea textarea-bordered w-full min-w-0" v-model="form.notes" rows="3" placeholder="Storage or handling note"></textarea></label>
          <div><button class="btn btn-ghost" type="button" @click="cancelEditor">Cancel</button><button class="btn btn-ghost" type="submit" :disabled="isSaving"><Save :size="15" :stroke-width="1.8" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save material' }}</button></div>
        </form>
      </SectionPanel>

      <SectionPanel v-else-if="selectedMaterial" title="Material inspector" subtitle="Current persisted record">
        <template #action><button class="btn btn-ghost" type="button" aria-label="Edit selected material" @click="startEdit"><Edit3 :size="15" :stroke-width="1.8" aria-hidden="true" /></button></template>
        <div><StatusBadge :label="selectedMaterial.active ? 'Active' : 'Archived'" :tone="selectedMaterial.active ? 'green' : 'slate'" /><StatusBadge v-if="selectedMaterial.lowStock" label="Low stock" tone="amber" /></div>
        <div><div><Package :size="19" :stroke-width="1.8" aria-hidden="true" /></div><div><h3>{{ selectedMaterial.name }}</h3><p>{{ selectedMaterial.sku || 'No SKU' }}<span v-if="selectedMaterial.category"> · {{ selectedMaterial.category }}</span></p></div></div>
        <dl>
          <div><dt>Unit conversion</dt><dd>1 {{ selectedMaterial.purchaseUnit }} = {{ selectedMaterial.conversionFactor }} {{ selectedMaterial.consumptionUnit }}</dd></div>
          <div><dt>Physical stock</dt><dd>{{ selectedMaterial.physicalStock }} {{ selectedMaterial.consumptionUnit }}</dd></div>
          <div><dt>Reserved stock</dt><dd>{{ selectedMaterial.reservedStock }} {{ selectedMaterial.consumptionUnit }}</dd></div>
          <div><dt>Available stock</dt><dd>{{ selectedMaterial.availableStock }} {{ selectedMaterial.consumptionUnit }} <small>Physical less active reservations</small></dd></div>
          <div><dt>Reorder level</dt><dd>{{ selectedMaterial.reorderLevel }} {{ selectedMaterial.consumptionUnit }}</dd></div>
          <div><dt>Average cost</dt><dd>{{ formatMoney(selectedMaterial.averageUnitCostRial, props.currencyUnit) }} <small>per {{ selectedMaterial.consumptionUnit }}</small></dd></div>
          <div><dt>Last updated</dt><dd>{{ dateLabel(selectedMaterial.updatedAt) }}</dd></div>
        </dl>
        <p v-if="selectedMaterial.notes">{{ selectedMaterial.notes }}</p>
          <div><button class="btn btn-ghost" v-if="selectedMaterial.active" type="button" @click="setActive(false)"><Archive :size="15" :stroke-width="1.8" aria-hidden="true" />Archive</button><button class="btn btn-ghost" v-else type="button" @click="setActive(true)"><RotateCcw :size="15" :stroke-width="1.8" aria-hidden="true" />Reactivate</button></div>
        <div><h3>Adjust stock</h3><p>Positive adds stock; negative records a supplier return or correction.</p><div><label class="form-control gap-1"><span>Quantity delta</span><input class="input input-bordered w-full min-w-0" v-model="adjustmentQuantity" placeholder="−1 or 2.5" inputmode="decimal" /></label><label class="form-control gap-1"><span>Unit cost ({{ props.currencyUnit }})</span><input class="input input-bordered w-full min-w-0" v-model="adjustmentCost" inputmode="numeric" /></label></div><label class="form-control gap-1"><span>Reason / note</span><input class="input input-bordered w-full min-w-0" v-model="adjustmentNote" placeholder="Count correction" /></label><button class="btn btn-ghost" type="button" @click="adjustStock"><RefreshCw :size="15"/> Record movement</button></div>
        <div><h3>Recent movements</h3><div v-if="movements.length" v-for="movement in movements.slice(0,6)" :key="movement.id"><span><strong>{{ movement.movementType }}</strong><small>{{ dateLabel(movement.occurredAt) }}</small></span><span :class="{'text-danger':movement.quantityDelta.startsWith('-')}">{{ movement.quantityDelta }} {{ selectedMaterial.consumptionUnit }}</span><span>{{ formatMoney(movement.totalCostRial, props.currencyUnit) }}</span></div><p v-else>No inventory movements yet.</p></div>
      </SectionPanel>

      <SectionPanel v-else title="Material inspector" subtitle="Select a row to inspect it.">
        <div><Check :size="20" :stroke-width="1.8" aria-hidden="true" /><p>Material details will appear here.</p><button class="btn btn-ghost" type="button" @click="startCreate">Create a material <Plus :size="14" :stroke-width="1.8" aria-hidden="true" /></button></div>
      </SectionPanel>
    </section>
  </div>
</template>
