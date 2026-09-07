<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Archive, Edit3, Factory, Plus, RotateCcw, Save, X } from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import SearchField from '../components/SearchField.vue'
import SelectField from '../components/SelectField.vue'
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
  <div>
    <WorkspaceStickyStack>
      <header><div><p>Catalog / production inputs</p><h1>Machines</h1><p>Keep reusable equipment rates ready for service cost definitions.</p></div><button class="btn btn-ghost" type="button" @click="startCreate"><Plus :size="16" :stroke-width="1.8" aria-hidden="true" />New machine</button></header>
      <section class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_12rem_auto] sm:items-end" aria-label="Machine filters"><SearchField v-model="query" label="Search machines" placeholder="Search machine, code, or category"/><SelectField v-model="filter" label="Status" aria-label="Filter machines by status" :options="['Active', 'Archived', 'All'].map((value) => ({ label: value, value }))"/><span class="self-end pb-2">{{ filtered.length }} of {{ machines.length }} machines</span></section>
    </WorkspaceStickyStack>
    <div v-if="error" role="alert"><span>{{ error }}</span><button class="btn btn-ghost" type="button" aria-label="Dismiss machines error" @click="error = ''"><X :size="15" :stroke-width="1.8" aria-hidden="true" /></button></div>
    <section aria-label="Machines workspace">
      <SectionPanel title="Machine register" subtitle="Rates are stored as integer Rial; the toolbar controls display units."><template #action><span>{{ filtered.length }} shown</span></template><div v-if="loading"><Factory :size="21" :stroke-width="1.8" aria-hidden="true" /><p>Loading machines…</p></div><div v-else-if="filtered.length"><table class="table table-zebra w-full"><thead><tr><th>Machine</th><th>Category</th><th>Rate basis</th><th>Rate</th><th>Status</th></tr></thead><tbody><tr v-for="machine in filtered" :key="machine.id" :class="{ 'bg-base-300': selectedId === machine.id }" tabindex="0" @click="select(machine.id)" @keydown.enter="select(machine.id)"><td><span>{{ machine.name }}</span><span>{{ machine.code || 'No code' }}</span></td><td>{{ machine.category || '—' }}</td><td>{{ basisLabel(machine.rateBasis) }}</td><td>{{ formatMoney(machine.rateRial, props.currencyUnit) }}<span>{{ machine.rateBasis }}</span></td><td><StatusBadge :label="machine.active ? 'Active' : 'Archived'" :tone="machine.active ? 'green' : 'slate'" /></td></tr></tbody></table></div><div v-else><div><Factory :size="21" :stroke-width="1.8" /></div><h2>{{ machines.length ? 'No machines match this view' : 'No machines yet' }}</h2><p>{{ machines.length ? 'Try another status or search term.' : 'Add the first reusable rate input for production.' }}</p><button class="btn btn-ghost" v-if="!machines.length" type="button" @click="startCreate"><Plus :size="15" :stroke-width="1.8" aria-hidden="true" />Create machine</button></div></SectionPanel>
      <SectionPanel v-if="mode" :title="mode === 'create' ? 'New machine' : 'Edit machine'" subtitle="Save a reusable machine rate definition."><template #action><button class="btn btn-ghost" type="button" aria-label="Close machine editor" @click="cancel"><X :size="16" :stroke-width="1.8" aria-hidden="true" /></button></template><form @submit.prevent="save"><div v-if="formError" role="alert">{{ formError }}</div><label class="form-control gap-1"><span>Name</span><input class="input input-bordered w-full min-w-0" v-model="form.name" type="text" placeholder="Production printer" /></label><div><label class="form-control gap-1"><span>Code</span><input class="input input-bordered w-full min-w-0" v-model="form.code" type="text" placeholder="PRINTER-01" /></label><label class="form-control gap-1"><span>Category</span><input class="input input-bordered w-full min-w-0" v-model="form.category" type="text" placeholder="Print production" /></label></div><div><SelectField v-model="form.rateBasis" label="Rate basis" :options="[{ label: 'Per unit / page', value: 'unit' }, { label: 'Per minute', value: 'minute' }, { label: 'Per hour', value: 'hour' }]"/><label class="form-control gap-1"><span>Rate ({{ props.currencyUnit }})</span><input class="input input-bordered w-full min-w-0" v-model="form.rate" type="text" inputmode="decimal" placeholder="0" /></label></div><label class="form-control gap-1"><span>Setup / fixed cost ({{ props.currencyUnit }})</span><input class="input input-bordered w-full min-w-0" v-model="form.setupCost" type="text" inputmode="decimal" placeholder="Optional" /></label><label class="form-control gap-1"><span>Notes</span><textarea class="textarea textarea-bordered w-full min-w-0" v-model="form.notes" rows="3" placeholder="Capacity, operating notes, or rate context"></textarea></label><div><button class="btn btn-ghost" type="button" @click="cancel">Cancel</button><button class="btn btn-ghost" type="submit" :disabled="saving"><Save :size="15" :stroke-width="1.8" aria-hidden="true" />{{ saving ? 'Saving…' : 'Save machine' }}</button></div></form></SectionPanel>
      <SectionPanel v-else-if="selectedMachine" title="Machine inspector" subtitle="Current persisted rate definition"><template #action><button class="btn btn-ghost" type="button" aria-label="Edit selected machine" @click="startEdit"><Edit3 :size="15" :stroke-width="1.8" aria-hidden="true" /></button></template><div><StatusBadge :label="selectedMachine.active ? 'Active' : 'Archived'" :tone="selectedMachine.active ? 'green' : 'slate'" /><span>{{ basisLabel(selectedMachine.rateBasis) }}</span></div><div><div><Factory :size="19" :stroke-width="1.8" aria-hidden="true" /></div><div><h3>{{ selectedMachine.name }}</h3><p>{{ selectedMachine.code || 'No code' }}<span v-if="selectedMachine.category"> · {{ selectedMachine.category }}</span></p></div></div><dl><div><dt>Rate</dt><dd>{{ formatMoney(selectedMachine.rateRial, props.currencyUnit) }}</dd></div><div><dt>Setup / fixed</dt><dd>{{ formatMoney(selectedMachine.setupCostRial, props.currencyUnit) }}</dd></div></dl><p v-if="selectedMachine.notes">{{ selectedMachine.notes }}</p><div>Updated {{ date(selectedMachine.updatedAt) }}</div><div><button class="btn btn-ghost" v-if="selectedMachine.active" type="button" @click="setActive(false)"><Archive :size="15" :stroke-width="1.8" aria-hidden="true" />Archive</button><button class="btn btn-ghost" v-else type="button" @click="setActive(true)"><RotateCcw :size="15" :stroke-width="1.8" aria-hidden="true" />Reactivate</button></div></SectionPanel>
      <SectionPanel v-else title="Machine inspector" subtitle="Select a row to inspect it."><div><Factory :size="20" :stroke-width="1.8" aria-hidden="true" /><p>Machine details will appear here.</p><button class="btn btn-ghost" type="button" @click="startCreate">Create a machine <Plus :size="14" :stroke-width="1.8" aria-hidden="true" /></button></div></SectionPanel>
    </section>
  </div>
</template>
