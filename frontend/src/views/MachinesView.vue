<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Archive, ChevronDown, Edit3, Factory, Plus, RotateCcw, Save, Search, X } from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import { machinesApi, type MachineRecord, type MachinePayload } from '../api/machines'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../utils/currency'
import { formatDateTime } from '../utils/date'

const props = defineProps<{ currencyUnit: CurrencyUnit }>()
const emit = defineEmits<{ notify: [message: string] }>()
type Filter = 'Active' | 'Archived' | 'All'
type Mode = 'create' | 'edit' | null
type MachineForm = Omit<MachinePayload, 'rateRial' | 'setupCostRial'> & { rate: string; setupCost: string }

const machines = ref<MachineRecord[]>([])
const selectedId = ref<string | null>(null)
const filter = ref<Filter>('Active')
const query = ref('')
const mode = ref<Mode>(null)
const form = ref<MachineForm>(emptyForm())
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const formError = ref('')
const selectedMachine = computed(() => machines.value.find((machine) => machine.id === selectedId.value) ?? null)
const filtered = computed(() => { const q = query.value.trim().toLowerCase(); return machines.value.filter((machine) => (filter.value === 'All' || (filter.value === 'Active' ? machine.active : !machine.active)) && (!q || [machine.name, machine.code, machine.category, machine.rateBasis].some((value) => value.toLowerCase().includes(q)))) })

onMounted(load)
watch(() => props.currencyUnit, () => { if (mode.value) { const machine = selectedMachine.value; if (machine) { form.value.rate = formatMoneyInput(machine.rateRial, props.currencyUnit); form.value.setupCost = formatMoneyInput(machine.setupCostRial, props.currencyUnit) } } })
watch(() => [form.value.rate, form.value.setupCost], ([rateValue, setupValue]) => {
  const parsedRate = parseMoneyInput(rateValue, props.currencyUnit)
  const parsedSetup = parseMoneyInput(setupValue, props.currencyUnit)
  if (parsedRate !== null) form.value.rate = formatMoneyInput(parsedRate, props.currencyUnit)
  if (parsedSetup !== null) form.value.setupCost = formatMoneyInput(parsedSetup, props.currencyUnit)
})
function emptyForm(): MachineForm { return { name: '', code: '', category: '', rateBasis: 'hour', rate: '', setupCost: '', notes: '' } }
function load() { loading.value = true; error.value = ''; machinesApi.list(true).then((data) => { machines.value = data; if (!selectedId.value && data.length) selectedId.value = data[0].id }).catch((e) => { error.value = message(e, 'Machines could not be loaded.') }).finally(() => { loading.value = false }) }
function select(id: string) { selectedId.value = id; mode.value = null; formError.value = '' }
function startCreate() { mode.value = 'create'; selectedId.value = null; form.value = emptyForm(); formError.value = '' }
function startEdit() { const machine = selectedMachine.value; if (!machine) return; form.value = { name: machine.name, code: machine.code, category: machine.category, rateBasis: machine.rateBasis, rate: formatMoneyInput(machine.rateRial, props.currencyUnit), setupCost: formatMoneyInput(machine.setupCostRial, props.currencyUnit), notes: machine.notes }; mode.value = 'edit'; formError.value = '' }
function cancel() { mode.value = null; formError.value = '' }
function rate(value: string) { return parseMoneyInput(value, props.currencyUnit) }
function date(value: string) { try { return formatDateTime(value) } catch { return 'Unknown date' } }
function basisLabel(value: string) { return ({ unit: 'Per unit / page', minute: 'Per minute', hour: 'Per hour' } as Record<string, string>)[value] ?? value }
async function save() {
  formError.value = ''; const parsedRate = rate(form.value.rate); const parsedSetup = form.value.setupCost.trim() === '' ? 0 : rate(form.value.setupCost)
  if (!form.value.name.trim()) { formError.value = 'Enter a machine name.'; return }
  if (parsedRate === null || parsedSetup === null) { formError.value = `Enter whole ${props.currencyUnit.toLowerCase()} amounts.`; return }
  saving.value = true
  const wasEditing = mode.value === 'edit'
  try {
    const payload = { name: form.value.name, code: form.value.code, category: form.value.category, rateBasis: form.value.rateBasis, rateRial: parsedRate, setupCostRial: parsedSetup, notes: form.value.notes }
    const saved = mode.value === 'edit' && selectedId.value ? await machinesApi.update(selectedId.value, payload) : await machinesApi.create(payload)
    const index = machines.value.findIndex((item) => item.id === saved.id); if (index >= 0) machines.value.splice(index, 1, saved); else machines.value.push(saved)
    selectedId.value = saved.id; mode.value = null; emit('notify', wasEditing ? 'Machine updated.' : 'Machine created.')
  } catch (e) { formError.value = message(e, 'Machine could not be saved.') } finally { saving.value = false }
}
async function setActive(active: boolean) { const machine = selectedMachine.value; if (!machine) return; try { const updated = active ? await machinesApi.reactivate(machine.id) : await machinesApi.archive(machine.id); const index = machines.value.findIndex((item) => item.id === updated.id); if (index >= 0) machines.value.splice(index, 1, updated); emit('notify', active ? 'Machine reactivated.' : 'Machine archived.') } catch (e) { error.value = message(e, 'Machine status could not be changed.') } }
function message(errorValue: unknown, fallback: string) { return errorValue instanceof Error && errorValue.message ? errorValue.message : typeof errorValue === 'string' ? errorValue : fallback }
</script>

<template>
  <div class="machines-view">
    <WorkspaceStickyStack>
      <header class="workspace-heading machines-heading"><div><p class="eyebrow">Catalog / production inputs</p><h1>Machines</h1><p class="heading-description">Keep reusable equipment rates ready for service cost definitions.</p></div><button class="button button-primary" type="button" @click="startCreate"><Plus class="button-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />New machine</button></header>
      <section class="machines-filter-bar panel" aria-label="Machine filters"><label class="services-search"><span class="sr-only">Search machines</span><Search :size="16" :stroke-width="1.8" aria-hidden="true" /><input v-model="query" type="search" placeholder="Search machine, code, or category" /></label><label class="filter-control machines-status-filter"><span>Status</span><span class="select-control"><select v-model="filter" aria-label="Filter machines by status"><option>Active</option><option>Archived</option><option>All</option></select><ChevronDown class="select-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" /></span></label><span class="filter-result">{{ filtered.length }} of {{ machines.length }} machines</span></section>
    </WorkspaceStickyStack>
    <div v-if="error" class="services-error" role="alert"><span>{{ error }}</span><button class="icon-button" type="button" aria-label="Dismiss machines error" @click="error = ''"><X :size="15" :stroke-width="1.8" aria-hidden="true" /></button></div>
    <section class="machines-layout" aria-label="Machines workspace">
      <SectionPanel title="Machine register" subtitle="Rates are stored as integer Rial; the toolbar controls display units." class="machines-table-panel"><template #action><span class="services-table-count">{{ filtered.length }} shown</span></template><div v-if="loading" class="services-empty"><Factory :size="21" :stroke-width="1.8" aria-hidden="true" /><p>Loading machines…</p></div><div v-else-if="filtered.length" class="table-wrap machines-table-wrap"><table class="data-table machines-table"><thead><tr><th>Machine</th><th>Category</th><th>Rate basis</th><th>Rate</th><th>Status</th></tr></thead><tbody><tr v-for="machine in filtered" :key="machine.id" :class="{ 'is-selected': selectedId === machine.id }" tabindex="0" @click="select(machine.id)" @keydown.enter="select(machine.id)"><td><span class="table-primary">{{ machine.name }}</span><span class="table-secondary">{{ machine.code || 'No code' }}</span></td><td>{{ machine.category || '—' }}</td><td>{{ basisLabel(machine.rateBasis) }}</td><td class="table-money">{{ formatMoney(machine.rateRial, props.currencyUnit) }}<span class="table-secondary">{{ machine.rateBasis }}</span></td><td><StatusBadge :label="machine.active ? 'Active' : 'Archived'" :tone="machine.active ? 'green' : 'slate'" /></td></tr></tbody></table></div><div v-else class="services-empty services-empty-register"><div class="empty-workspace-icon"><Factory :size="21" :stroke-width="1.8" /></div><h2>{{ machines.length ? 'No machines match this view' : 'No machines yet' }}</h2><p>{{ machines.length ? 'Try another status or search term.' : 'Add the first reusable rate input for production.' }}</p><button v-if="!machines.length" class="button button-secondary" type="button" @click="startCreate"><Plus class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Create machine</button></div></SectionPanel>
      <SectionPanel v-if="mode" :title="mode === 'create' ? 'New machine' : 'Edit machine'" subtitle="Save a reusable machine rate definition." class="service-inspector machine-editor"><template #action><button class="icon-button" type="button" aria-label="Close machine editor" @click="cancel"><X :size="16" :stroke-width="1.8" aria-hidden="true" /></button></template><form class="service-form" @submit.prevent="save"><div v-if="formError" class="form-error" role="alert">{{ formError }}</div><label class="form-field form-field-wide"><span>Name</span><input v-model="form.name" type="text" placeholder="Production printer" /></label><div class="service-form-grid"><label class="form-field"><span>Code</span><input v-model="form.code" type="text" placeholder="PRINTER-01" /></label><label class="form-field"><span>Category</span><input v-model="form.category" type="text" placeholder="Print production" /></label></div><div class="service-form-grid"><label class="form-field"><span>Rate basis</span><span class="select-control"><select v-model="form.rateBasis"><option value="unit">Per unit / page</option><option value="minute">Per minute</option><option value="hour">Per hour</option></select><ChevronDown class="select-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" /></span></label><label class="form-field"><span>Rate ({{ props.currencyUnit }})</span><input v-model="form.rate" type="text" inputmode="decimal" placeholder="0" /></label></div><label class="form-field form-field-wide"><span>Setup / fixed cost ({{ props.currencyUnit }})</span><input v-model="form.setupCost" type="text" inputmode="decimal" placeholder="Optional" /></label><label class="form-field form-field-wide"><span>Notes</span><textarea v-model="form.notes" rows="3" placeholder="Capacity, operating notes, or rate context"></textarea></label><div class="service-form-actions"><button class="button button-secondary" type="button" @click="cancel">Cancel</button><button class="button button-primary" type="submit" :disabled="saving"><Save class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />{{ saving ? 'Saving…' : 'Save machine' }}</button></div></form></SectionPanel>
      <SectionPanel v-else-if="selectedMachine" title="Machine inspector" subtitle="Current persisted rate definition" class="service-inspector"><template #action><button class="icon-button" type="button" aria-label="Edit selected machine" @click="startEdit"><Edit3 :size="15" :stroke-width="1.8" aria-hidden="true" /></button></template><div class="inspector-status"><StatusBadge :label="selectedMachine.active ? 'Active' : 'Archived'" :tone="selectedMachine.active ? 'green' : 'slate'" /><span class="parameter-summary">{{ basisLabel(selectedMachine.rateBasis) }}</span></div><div class="service-inspector-heading"><div class="service-inspector-icon"><Factory :size="19" :stroke-width="1.8" aria-hidden="true" /></div><div><h3>{{ selectedMachine.name }}</h3><p>{{ selectedMachine.code || 'No code' }}<span v-if="selectedMachine.category"> · {{ selectedMachine.category }}</span></p></div></div><dl class="machine-details"><div><dt>Rate</dt><dd>{{ formatMoney(selectedMachine.rateRial, props.currencyUnit) }}</dd></div><div><dt>Setup / fixed</dt><dd>{{ formatMoney(selectedMachine.setupCostRial, props.currencyUnit) }}</dd></div></dl><p v-if="selectedMachine.notes" class="service-description">{{ selectedMachine.notes }}</p><div class="inspector-meta">Updated {{ date(selectedMachine.updatedAt) }}</div><div class="inspector-actions"><button v-if="selectedMachine.active" class="button button-secondary" type="button" @click="setActive(false)"><Archive class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Archive</button><button v-else class="button button-secondary" type="button" @click="setActive(true)"><RotateCcw class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Reactivate</button></div></SectionPanel>
      <SectionPanel v-else title="Machine inspector" subtitle="Select a row to inspect it." class="service-inspector service-inspector-empty"><div class="inspector-empty-copy"><Factory :size="20" :stroke-width="1.8" aria-hidden="true" /><p>Machine details will appear here.</p><button class="text-button" type="button" @click="startCreate">Create a machine <Plus :size="14" :stroke-width="1.8" aria-hidden="true" /></button></div></SectionPanel>
    </section>
  </div>
</template>
