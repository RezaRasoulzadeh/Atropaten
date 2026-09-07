<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Mail, MapPin, Pencil, Phone, Plus, Trash2, UserRound, X } from 'lucide-vue-next'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { customersApi, type CustomerPayload, type CustomerRecord } from '../api/customers'
import { formatDateTime } from '../utils/date'
import { confirmAction, normalizeError } from '../ui/feedback'
import SearchFilterBar from '../components/SearchFilterBar.vue'
import SearchField from '../components/SearchField.vue'
import SelectField from '../components/SelectField.vue'

const props = defineProps<{ refreshKey?: number }>()
const emit = defineEmits<{ notify: [message: string] }>()
const customers = ref<CustomerRecord[]>([])
const loading = ref(true)
const error = ref('')
const query = ref('')
const filter = ref<'Active' | 'Archived' | 'All'>('Active')
const selectedId = ref<string | null>(null)
const editing = ref(false)
const saving = ref(false)
const form = ref<CustomerPayload>({ name: '', phone: '', email: '', address: '', notes: '' })
const visible = computed(() => customers.value.filter((customer) => {
  const matchesStatus = filter.value === 'All' || (filter.value === 'Active' ? customer.active : !customer.active)
  const q = query.value.trim().toLowerCase()
  return matchesStatus && (!q || [customer.name, customer.phone, customer.email].some((value) => value.toLowerCase().includes(q)))
}))
const selected = computed(() => customers.value.find((customer) => customer.id === selectedId.value) ?? null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    customers.value = await customersApi.list(true)
    if (!selectedId.value && customers.value.length) select(customers.value[0])
  } catch (value) { error.value = normalizeError(value).message } finally { loading.value = false }
}
function select(customer: CustomerRecord) { selectedId.value = customer.id; editing.value = false }
function newCustomer() { selectedId.value = null; editing.value = true; form.value = { name: '', phone: '', email: '', address: '', notes: '' } }
function editCustomer() {
  if (!selected.value) return
  form.value = { name: selected.value.name, phone: selected.value.phone, email: selected.value.email, address: selected.value.address, notes: selected.value.notes }
  editing.value = true
}
async function save() {
  saving.value = true
  try {
    const result = selectedId.value ? await customersApi.update(selectedId.value, form.value) : await customersApi.create(form.value)
    customers.value = selectedId.value ? customers.value.map((customer) => customer.id === result.id ? result : customer) : [result, ...customers.value]
    selectedId.value = result.id
    editing.value = false
    emit('notify', 'Customer saved')
  } catch (value) { emit('notify', `Error: ${normalizeError(value).message}`) } finally { saving.value = false }
}
async function toggle() {
  if (!selected.value) return
  try {
    const result = selected.value.active ? await customersApi.archive(selected.value.id) : await customersApi.reactivate(selected.value.id)
    customers.value = customers.value.map((customer) => customer.id === result.id ? result : customer)
    emit('notify', result.active ? 'Customer reactivated' : 'Customer archived')
  } catch (value) { emit('notify', `Error: ${normalizeError(value).message}`) }
}
async function remove() {
  if (!selected.value || !(await confirmAction({ title: 'Delete customer', message: 'Delete this customer permanently? Only unreferenced customers can be deleted.', confirmLabel: 'Delete customer', danger: true }))) return
  try {
    await customersApi.remove(selected.value.id)
    customers.value = customers.value.filter((customer) => customer.id !== selected.value!.id)
    selectedId.value = null
    emit('notify', 'Customer deleted.')
  } catch (value) { emit('notify', `Error: ${normalizeError(value).message}`) }
}
watch(() => props.refreshKey, load, { immediate: true })
</script>

<template>
  <div>
    <WorkspaceStickyStack>
      <header><div><p>Workspace / relationships</p><h1>Customers</h1><p>Keep customer contacts ready for every commercial workflow.</p></div><button class="btn btn-primary" type="button" @click="newCustomer"><Plus :size="16" aria-hidden="true" />New customer</button></header>
      <SearchFilterBar><template #search><SearchField v-model="query" label="Search customers" placeholder="Search name, phone, or email" /></template><template #filters><SelectField v-model="filter" label="Status" aria-label="Customer status" :options="['Active', 'Archived', 'All'].map((value) => ({ label: value, value }))" /></template><template #count><span>{{ visible.length }} customers</span></template></SearchFilterBar>
    </WorkspaceStickyStack>
    <div class="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(20rem,28rem)]">
      <section class="card min-w-0 overflow-hidden border border-base-300 bg-base-100 shadow-none"><div v-if="loading">Loading customers…</div><div v-else-if="error">{{ error }}</div><div v-else-if="!visible.length"><UserRound :size="22" /><strong>No customers in this view</strong><button class="btn btn-ghost" @click="newCustomer">Add customer</button></div><button v-for="customer in visible" v-else :key="customer.id" :class="{ 'bg-base-300': selectedId === customer.id }" @click="select(customer)"><span><UserRound :size="16" /></span><span><strong>{{ customer.name }}</strong><small>{{ customer.phone || customer.email || 'No contact details' }}</small></span><StatusBadge :label="customer.active ? 'Active' : 'Archived'" :tone="customer.active ? 'green' : 'slate'" /></button></section>
      <aside><template v-if="editing"><header><div><p>{{ selectedId ? 'Edit customer' : 'New customer' }}</p><h2>{{ selectedId ? 'Customer details' : 'Create customer' }}</h2></div><button class="btn btn-ghost" aria-label="Cancel editing" @click="editing = false"><X :size="16" /></button></header><form @submit.prevent="save"><label class="form-control gap-1">Customer name<input class="input input-bordered w-full min-w-0" v-model="form.name" required placeholder="e.g. Mehr Studio" /></label><div><label class="form-control gap-1">Phone<input class="input input-bordered w-full min-w-0" v-model="form.phone" placeholder="+98 …" /></label><label class="form-control gap-1">Email<input class="input input-bordered w-full min-w-0" v-model="form.email" type="email" placeholder="contact@example.com" /></label></div><label class="form-control gap-1">Address<textarea class="textarea textarea-bordered w-full min-w-0" v-model="form.address" rows="2" /></label><label class="form-control gap-1">Notes<textarea class="textarea textarea-bordered w-full min-w-0" v-model="form.notes" rows="3" /></label><button class="btn btn-ghost" :disabled="saving">{{ saving ? 'Saving…' : 'Save customer' }}</button></form></template><template v-else-if="selected"><header><div><p>Customer profile</p><h2>{{ selected.name }}</h2><StatusBadge :label="selected.active ? 'Active' : 'Archived'" :tone="selected.active ? 'green' : 'slate'" /></div><button class="btn btn-ghost" @click="editCustomer"><Pencil :size="14" />Edit</button></header><div><div><Phone :size="15" /><span>{{ selected.phone || 'No phone provided' }}</span></div><div><Mail :size="15" /><span>{{ selected.email || 'No email provided' }}</span></div><div><MapPin :size="15" /><span>{{ selected.address || 'No address provided' }}</span></div></div><p>Updated {{ formatDateTime(selected.updatedAt) }}</p><div><button class="btn btn-ghost" @click="toggle">{{ selected.active ? 'Archive customer' : 'Reactivate customer' }}</button><button class="btn btn-ghost" @click="remove"><Trash2 :size="14" />Delete</button></div></template><div v-else><UserRound :size="24" /><strong>Select a customer</strong><span>Choose a row to inspect its details.</span></div></aside>
    </div>
  </div>
</template>
