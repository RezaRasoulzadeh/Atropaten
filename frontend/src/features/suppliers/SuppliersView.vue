<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import RegisterList from '../../components/ui/RegisterList.vue'
import RegisterRow from '../../components/ui/RegisterRow.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import InlineAlert from '../../components/ui/InlineAlert.vue';
import FormGrid from '../../components/ui/FormGrid.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import InspectorSection from '../../components/layout/InspectorSection.vue';
import MasterDetail from '../../components/layout/MasterDetail.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTableRow from '../../components/ui/DataTableRow.vue';
import DataTable from '../../components/ui/DataTable.vue';
import AppInput from '../../components/ui/AppInput.vue';
import FormField from '../../components/ui/FormField.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref } from 'vue';
import { Archive, Edit3, Plus, RotateCcw, Save, Trash2, Truck, X } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import { suppliersApi, type SupplierPayload, type SupplierRecord } from '../../api/suppliers';
import { formatDateTime } from '../../utils/date';
import { confirmAction, normalizeError } from '../../ui/feedback';
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';
const emit = defineEmits<{ notify: [string] }>();
type Filter = 'Active' | 'Archived' | 'All';
const rows = ref<SupplierRecord[]>([]);
const selected = ref<string | null>(null);
const filter = ref<Filter>('Active');
const search = ref('');
const editing = ref(false);
const form = ref<SupplierPayload>(empty());
const error = ref('');
const loading = ref(false);
const current = computed(() => rows.value.find((v) => v.id === selected.value) ?? null);
const filtered = computed(() =>
  rows.value.filter(
    (v) =>
      (filter.value === 'All' || (filter.value === 'Active' ? v.active : !v.active)) &&
      (!search.value ||
        [v.name, v.code, v.phone, v.email]
          .join(' ')
          .toLowerCase()
          .includes(search.value.toLowerCase())),
  ),
);
onMounted(load);
function empty(): SupplierPayload {
  return { name: '', code: '', phone: '', email: '', address: '', notes: '' };
}
async function load() {
  loading.value = true;
  try {
    rows.value = await suppliersApi.list(true);
    if (!selected.value && rows.value[0]) selected.value = rows.value[0].id;
  } catch (e) {
    error.value = normalizeError(e).message;
  } finally {
    loading.value = false;
  }
}
function startCreate() {
  selected.value = null;
  form.value = empty();
  editing.value = true;
}
function startEdit() {
  if (!current.value) return;
  form.value = {
    name: current.value.name,
    code: current.value.code,
    phone: current.value.phone,
    email: current.value.email,
    address: current.value.address,
    notes: current.value.notes,
  };
  editing.value = true;
}
async function save() {
return runAction(async () => {
  try {
    const v =
      editing.value && selected.value
        ? await suppliersApi.update(selected.value, form.value)
        : ((await suppliersApi.create(form.value)) as unknown as SupplierRecord);
    rows.value = rows.value.some((x) => x.id === v.id)
      ? rows.value.map((x) => (x.id === v.id ? v : x))
      : [v, ...rows.value];
    selected.value = v.id;
    editing.value = false;
    emit('notify', 'Supplier saved.');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
async function setActive(active: boolean) {
return runAction(async () => {
  if (!current.value) return;
  try {
    const v = active
      ? await suppliersApi.reactivate(current.value.id)
      : await suppliersApi.archive(current.value.id);
    rows.value = rows.value.map((x) => (x.id === v.id ? v : x));
    emit('notify', active ? 'Supplier reactivated.' : 'Supplier archived.');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
async function remove() {
return runAction(async () => {
  if (
    !current.value ||
    !(await confirmAction({
      title: 'Delete supplier',
      message: 'Delete this supplier permanently?',
      confirmLabel: 'Delete supplier',
      danger: true,
    }))
  )
    return;
  try {
    await suppliersApi.remove(current.value.id);
    rows.value = rows.value.filter((x) => x.id !== current.value!.id);
    selected.value = null;
    emit('notify', 'Supplier deleted.');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
</script>
<template>
  <div class="min-w-0 space-y-3">
    <WorkspaceStickyStack
      ><WorkspaceHeader
        title="Suppliers"
        eyebrow="Catalog / purchasing"
        description="Keep supplier contacts and purchasing history safe and readable."
        ><button class="btn btn-primary" type="button" @click="startCreate">
          <Plus :size="16" /> New supplier
        </button></WorkspaceHeader
      ><SearchFilterBar
        ><template #search
          ><SearchField
            v-model="search"
            label="Search suppliers"
            placeholder="Search suppliers" /></template
        ><template #filters
          ><SelectField
            v-model="filter"
            label="Status"
            :options="
              ['Active', 'Archived', 'All'].map((value) => ({ label: value, value }))
            " /></template
        ><template #count
          ><span>{{ filtered.length }} shown</span></template
        ></SearchFilterBar
      ></WorkspaceStickyStack
    ><InlineAlert v-if="error" role="alert" class="flex flex-wrap items-center gap-2" tone="error"
      >{{ error }}
      <button class="btn btn-ghost" @click="error = ''" aria-label="Dismiss">
        <X :size="14" /></button></InlineAlert
    ><MasterDetail
      ><RegisterList title="Supplier register" subtitle="Select a supplier to inspect contact details." :count="filtered.length">
<LoadingState v-if="loading" label="Loading suppliers…" />
<EmptyState v-else-if="!filtered.length" title="No suppliers in this view" description="Change the filters or add a supplier." />
<template v-else><RegisterRow v-for="v in filtered" :key="v.id" :selected="selected === v.id" @activate="selected = v.id; editing = false">
<template #icon><Truck :size="17" /></template><template #identity><strong class="block text-sm">{{v.name}}</strong></template><template #meta><p class="mt-1 text-xs leading-5 text-base-content/60">{{v.code || 'No code'}} · {{v.phone || v.email || 'No contact details'}}</p></template><template #status><StatusBadge :label="v.active ? 'Active' : 'Archived'" :tone="v.active ? 'green' : 'slate'" /></template>
</RegisterRow></template></RegisterList><InspectorShell
        v-if="editing"
        title="Supplier editor"
        subtitle="Persisted in the local database."
        ><form @submit.prevent="save" class="min-w-0 space-y-3">
          <FormField class="gap-1"
            ><span>Name</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="form.name"
              required /></FormField
          ><FormGrid
            ><FormField class="gap-1"
              ><span>Code</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.code" /></FormField
            ><FormField class="gap-1"
              ><span>Phone</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.phone" /></FormField></FormGrid
          ><FormField class="gap-1"
            ><span>Email</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="form.email"
              type="email" /></FormField
          ><FormField class="gap-1"
            ><span>Address</span
            ><textarea
              class="textarea w-full min-w-0"
              v-model="form.address"
              rows="2"
            /></FormField
          ><FormField class="gap-1"
            ><span>Notes</span
            ><textarea
              class="textarea w-full min-w-0"
              v-model="form.notes"
              rows="3"
            />
          </FormField>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-ghost" type="button" @click="editing = false">Cancel</button
            ><button class="btn btn-primary"><Save :size="15" /> Save</button>
          </div>
        </form></InspectorShell
      ><InspectorShell
        v-else-if="current"
        title="Supplier inspector"
        subtitle="Archive and delete have different meanings."
        ><div class="min-w-0 space-y-3">
          <div><Truck :size="19" /></div>
          <div class="min-w-0 space-y-3">
            <h3 class="text-sm font-semibold">{{ current.name }}</h3>
            <p>{{ current.code || 'No code' }}</p>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <StatusBadge
            :label="current.active ? 'Active' : 'Archived'"
            :tone="current.active ? 'green' : 'slate'"
          />
        </div>
        <InspectorSection title="Contact details">
        <dl class="grid min-w-0 gap-2 text-sm">
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Phone</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">{{ current.phone || '—' }}</dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Email</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">{{ current.email || '—' }}</dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Address</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ current.address || '—' }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Updated</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ formatDateTime(current.updatedAt) }}
            </dd>
          </div>
        </dl>
        </InspectorSection>
        <div class="flex flex-wrap items-center gap-2">
          <button class="btn btn-ghost" @click="startEdit"><Edit3 :size="15" /> Edit</button
          ><button class="btn btn-ghost" @click="setActive(!current.active)" :disabled="busy">
            <RotateCcw v-if="!current.active" :size="15" /><Archive v-else :size="15" />
            {{ current.active ? 'Archive' : 'Reactivate' }}</button
          ><button class="btn btn-ghost" @click="remove" :disabled="busy"><Trash2 :size="15" /> Delete</button>
        </div></InspectorShell
      ><InspectorShell v-else title="Supplier inspector" subtitle="Select a row to inspect it."
        ><div class="min-w-0 space-y-3">
          <Truck :size="20" />
          <p>Supplier details will appear here.</p>
        </div></InspectorShell
      ></MasterDetail
    >
  </div>
</template>
