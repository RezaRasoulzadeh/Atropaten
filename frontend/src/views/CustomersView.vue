<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Mail, MapPin, Pencil, Phone, Plus, Search, Trash2, UserRound, X } from 'lucide-vue-next'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { customersApi, type CustomerPayload, type CustomerRecord } from '../api/customers'
import { formatDateTime } from '../utils/date'

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
  } catch (value) { error.value = String(value) } finally { loading.value = false }
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
  } catch (value) { emit('notify', String(value)) } finally { saving.value = false }
}
async function toggle() {
  if (!selected.value) return
  try {
    const result = selected.value.active ? await customersApi.archive(selected.value.id) : await customersApi.reactivate(selected.value.id)
    customers.value = customers.value.map((customer) => customer.id === result.id ? result : customer)
    emit('notify', result.active ? 'Customer reactivated' : 'Customer archived')
  } catch (value) { emit('notify', String(value)) }
}
async function remove() {
  if (!selected.value || !window.confirm('Delete this customer permanently? Only unreferenced customers can be deleted.')) return
  try {
    await customersApi.remove(selected.value.id)
    customers.value = customers.value.filter((customer) => customer.id !== selected.value!.id)
    selectedId.value = null
    emit('notify', 'Customer deleted.')
  } catch (value) { emit('notify', String(value)) }
}
watch(() => props.refreshKey, load, { immediate: true })
</script>

<template>
  <div class="customers-view">
    <WorkspaceStickyStack>
      <header class="workspace-heading"><div><p class="eyebrow">Workspace / relationships</p><h1>Customers</h1><p class="heading-description">Keep customer contacts ready for every commercial workflow.</p></div><button class="button button-primary" @click="newCustomer"><Plus :size="16" aria-hidden="true" />New customer</button></header>
      <section class="customers-toolbar panel"><label class="orders-search"><Search :size="16" aria-hidden="true" /><span class="sr-only">Search customers</span><input v-model="query" placeholder="Search name, phone, or email" /></label><label class="select-control"><select v-model="filter" aria-label="Customer status"><option>Active</option><option>Archived</option><option>All</option></select></label><span class="filter-result">{{ visible.length }} customers</span></section>
    </WorkspaceStickyStack>
    <div class="customers-layout">
      <section class="customers-list panel"><div v-if="loading" class="workspace-state">Loading customers…</div><div v-else-if="error" class="workspace-state state-error">{{ error }}</div><div v-else-if="!visible.length" class="workspace-state"><UserRound :size="22" /><strong>No customers in this view</strong><button class="button button-secondary" @click="newCustomer">Add customer</button></div><button v-for="customer in visible" v-else :key="customer.id" class="customer-row" :class="{ 'is-selected': selectedId === customer.id }" @click="select(customer)"><span class="customer-avatar"><UserRound :size="16" /></span><span class="customer-row-copy"><strong>{{ customer.name }}</strong><small>{{ customer.phone || customer.email || 'No contact details' }}</small></span><StatusBadge :label="customer.active ? 'Active' : 'Archived'" :tone="customer.active ? 'green' : 'slate'" /></button></section>
      <aside class="customer-inspector panel"><template v-if="editing"><header class="inspector-heading"><div><p class="section-kicker">{{ selectedId ? 'Edit customer' : 'New customer' }}</p><h2>{{ selectedId ? 'Customer details' : 'Create customer' }}</h2></div><button class="icon-button" aria-label="Cancel editing" @click="editing = false"><X :size="16" /></button></header><form class="customer-form" @submit.prevent="save"><label>Customer name<input v-model="form.name" required placeholder="e.g. Mehr Studio" /></label><div class="form-grid"><label>Phone<input v-model="form.phone" placeholder="+98 …" /></label><label>Email<input v-model="form.email" type="email" placeholder="contact@example.com" /></label></div><label>Address<textarea v-model="form.address" rows="2" /></label><label>Notes<textarea v-model="form.notes" rows="3" /></label><button class="button button-primary" :disabled="saving">{{ saving ? 'Saving…' : 'Save customer' }}</button></form></template><template v-else-if="selected"><header class="inspector-heading"><div><p class="section-kicker">Customer profile</p><h2>{{ selected.name }}</h2><StatusBadge :label="selected.active ? 'Active' : 'Archived'" :tone="selected.active ? 'green' : 'slate'" /></div><button class="button button-secondary" @click="editCustomer"><Pencil :size="14" />Edit</button></header><div class="customer-detail-list"><div><Phone :size="15" /><span>{{ selected.phone || 'No phone provided' }}</span></div><div><Mail :size="15" /><span>{{ selected.email || 'No email provided' }}</span></div><div><MapPin :size="15" /><span>{{ selected.address || 'No address provided' }}</span></div></div><p class="inspector-note">Updated {{ formatDateTime(selected.updatedAt) }}</p><div class="inspector-actions"><button class="button button-secondary" @click="toggle">{{ selected.active ? 'Archive customer' : 'Reactivate customer' }}</button><button class="button button-danger" @click="remove"><Trash2 :size="14" />Delete</button></div></template><div v-else class="workspace-state"><UserRound :size="24" /><strong>Select a customer</strong><span>Choose a row to inspect its details.</span></div></aside>
    </div>
  </div>
</template>
