<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import MasterDetail from '../../components/layout/MasterDetail.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import FormField from '../../components/ui/FormField.vue';
import { computed, ref, watch } from 'vue';
import { Mail, MapPin, Pencil, Phone, Plus, Trash2, UserRound, X } from 'lucide-vue-next';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import InspectorSection from '../../components/layout/InspectorSection.vue';
import FormSection from '../../components/ui/FormSection.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import LoadingState from '../../components/ui/LoadingState.vue';
import InlineAlert from '../../components/ui/InlineAlert.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import { customersApi, type CustomerPayload, type CustomerRecord } from '../../api/customers';
import { formatDateTime } from '../../utils/date';
import { confirmAction, normalizeError } from '../../ui/feedback';
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';

const props = defineProps<{ refreshKey?: number }>();
const emit = defineEmits<{ notify: [message: string] }>();
const customers = ref<CustomerRecord[]>([]);
const loading = ref(true);
const error = ref('');
const query = ref('');
const filter = ref<'Active' | 'Archived' | 'All'>('Active');
const selectedId = ref<string | null>(null);
const editing = ref(false);
const saving = ref(false);
const form = ref<CustomerPayload>({ name: '', phone: '', email: '', address: '', notes: '' });
const visible = computed(() =>
  customers.value.filter((customer) => {
    const matchesStatus =
      filter.value === 'All' || (filter.value === 'Active' ? customer.active : !customer.active);
    const q = query.value.trim().toLowerCase();
    return (
      matchesStatus &&
      (!q ||
        [customer.name, customer.phone, customer.email].some((value) =>
          value.toLowerCase().includes(q),
        ))
    );
  }),
);
const selected = computed(
  () => customers.value.find((customer) => customer.id === selectedId.value) ?? null,
);

async function load() {
  loading.value = true;
  error.value = '';
  try {
    customers.value = await customersApi.list(true);
    if (!selectedId.value && customers.value.length) select(customers.value[0]);
  } catch (value) {
    error.value = normalizeError(value).message;
  } finally {
    loading.value = false;
  }
}
function select(customer: CustomerRecord) {
  selectedId.value = customer.id;
  editing.value = false;
}
function newCustomer() {
  selectedId.value = null;
  editing.value = true;
  form.value = { name: '', phone: '', email: '', address: '', notes: '' };
}
function editCustomer() {
  if (!selected.value) return;
  form.value = {
    name: selected.value.name,
    phone: selected.value.phone,
    email: selected.value.email,
    address: selected.value.address,
    notes: selected.value.notes,
  };
  editing.value = true;
}
async function save() {
return runAction(async () => {
  saving.value = true;
  try {
    const result = selectedId.value
      ? await customersApi.update(selectedId.value, form.value)
      : await customersApi.create(form.value);
    customers.value = selectedId.value
      ? customers.value.map((customer) => (customer.id === result.id ? result : customer))
      : [result, ...customers.value];
    selectedId.value = result.id;
    editing.value = false;
    emit('notify', 'Customer saved');
  } catch (value) {
reportError(value);
  } finally {
    saving.value = false;
  }

});
}
async function toggle() {
return runAction(async () => {
  if (!selected.value) return;
  try {
    const result = selected.value.active
      ? await customersApi.archive(selected.value.id)
      : await customersApi.reactivate(selected.value.id);
    customers.value = customers.value.map((customer) =>
      customer.id === result.id ? result : customer,
    );
    emit('notify', result.active ? 'Customer reactivated' : 'Customer archived');
  } catch (value) {
reportError(value);
  }

});
}
async function remove() {
return runAction(async () => {
  if (
    !selected.value ||
    !(await confirmAction({
      title: 'Delete customer',
      message: 'Delete this customer permanently? Only unreferenced customers can be deleted.',
      confirmLabel: 'Delete customer',
      danger: true,
    }))
  )
    return;
  try {
    await customersApi.remove(selected.value.id);
    customers.value = customers.value.filter((customer) => customer.id !== selected.value!.id);
    selectedId.value = null;
    emit('notify', 'Customer deleted.');
  } catch (value) {
reportError(value);
  }

});
}
watch(() => props.refreshKey, load, { immediate: true });
</script>

<template>
  <div class="min-w-0 space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        eyebrow="Workspace / relationships"
        title="Customers"
        description="Keep customer contacts ready for every commercial workflow."
      >
        <button class="btn btn-primary" type="button" @click="newCustomer">
          <Plus :size="16" aria-hidden="true" />
          New customer
        </button>
      </WorkspaceHeader>
      <SearchFilterBar>
        <template #search
          ><SearchField
            v-model="query"
            label="Search customers"
            placeholder="Search name, phone, or email"
        /></template>
        <template #filters
          ><SelectField
            v-model="filter"
            label="Status"
            aria-label="Customer status"
            :options="['Active', 'Archived', 'All'].map((value) => ({ label: value, value }))"
        /></template>
        <template #count
          ><span>{{ visible.length }} customers</span></template
        >
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <InlineAlert v-if="error" :message="error" />

    <MasterDetail wide>
      <RegisterList
        title="Customer register"
        subtitle="Select a customer to inspect contact and account details."
        :count="visible.length"
      >
        <LoadingState v-if="loading" label="Loading customers…" />
        <EmptyState
          v-else-if="!visible.length"
          title="No customers in this view"
          description="Adjust the search or create a new customer."
        />
        <RegisterRow
          v-for="customer in visible"
          v-else
          :key="customer.id"
          :selected="selectedId === customer.id"
          @activate="select(customer)"
        >
          <template #icon><UserRound :size="17" aria-hidden="true" /></template>
          <template #identity>
            <div class="flex min-w-0 items-center justify-between gap-3">
              <strong class="block min-w-0 truncate text-sm font-semibold">{{ customer.name }}</strong>
              <StatusBadge
                class="shrink-0"
                :label="customer.active ? 'Active' : 'Archived'"
                :tone="customer.active ? 'green' : 'slate'"
              />
            </div>
          </template>
          <template #meta>
            <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2">
              <div class="min-w-0">
                <span class="block text-base-content/50">Phone</span>
                <span class="block truncate text-base-content/80">{{ customer.phone || 'No phone' }}</span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Email</span>
                <span class="block break-all text-base-content/80">{{ customer.email || 'No email' }}</span>
              </div>
            </div>
          </template>
        </RegisterRow>
      </RegisterList>

      <InspectorShell
        v-if="editing"
        :title="selectedId ? 'Edit customer' : 'New customer'"
        :sticky-footer="false"
        :subtitle="
          selectedId ? 'Update this customer profile.' : 'Add a contact for commercial workflows.'
        "
      >
        <template #header>
          <button
            class="btn btn-ghost btn-sm"
            type="button"
            aria-label="Cancel editing"
            @click="editing = false"
          >
            <X :size="16" />
          </button>
        </template>
        <form id="customer-editor" class="space-y-3" @submit.prevent="save">
          <FormSection title="Identity">
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs">Customer name</span
              ><AppInput
                v-model="form.name"
                class="input w-full min-w-0"
                required
                placeholder="e.g. Mehr Studio"
            /></FormField>
          </FormSection>
          <FormSection title="Contact">
            <FormField class="gap-1"
              ><span class="text-xs">Phone</span
              ><AppInput
                v-model="form.phone"
                class="input w-full min-w-0"
                placeholder="+98 …"
            /></FormField>
            <FormField class="gap-1"
              ><span class="text-xs">Email</span
              ><AppInput
                v-model="form.email"
                class="input w-full min-w-0"
                type="email"
                placeholder="contact@example.com"
            /></FormField>
          </FormSection>
          <FormSection title="Address and notes">
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs">Address</span
              ><AppTextarea
                v-model="form.address"
                rows="2"
              />
            </FormField>
            <FormField class="gap-1 sm:col-span-2"
              ><span class="text-xs">Notes</span
              ><AppTextarea
                v-model="form.notes"
                rows="3"
              />
            </FormField>
          </FormSection>
        </form>
        <template #footer>
          <div class="flex w-full justify-end gap-2">
            <button class="btn btn-ghost" type="button" :disabled="saving" @click="editing = false">
              Cancel
            </button>
            <button class="btn btn-primary" type="submit" form="customer-editor" :disabled="busy || saving">
              {{ saving ? 'Saving…' : 'Save customer' }}
            </button>
          </div>
        </template>
      </InspectorShell>

      <InspectorShell
        v-else-if="selected"
        :title="selected.name"
        :sticky-footer="false"
        subtitle="Customer profile and account details."
      >
        <template #header
          ><StatusBadge
            :label="selected.active ? 'Active' : 'Archived'"
            :tone="selected.active ? 'green' : 'slate'"
          /><button class="btn btn-ghost btn-sm" type="button" @click="editCustomer">
            <Pencil :size="14" />Edit
          </button></template
        >
        <div class="space-y-4">
          <InspectorSection title="Contact details">
          <section class="grid gap-2 rounded-box border border-base-300 bg-base-200/35 p-3 text-sm">
            <div class="flex items-start gap-2">
              <Phone :size="15" class="mt-0.5 shrink-0 text-base-content/60" /><span>{{
                selected.phone || 'No phone provided'
              }}</span>
            </div>
            <div class="flex items-start gap-2">
              <Mail :size="15" class="mt-0.5 shrink-0 text-base-content/60" /><span
                class="break-all"
                >{{ selected.email || 'No email provided' }}</span
              >
            </div>
            <div class="flex items-start gap-2">
              <MapPin :size="15" class="mt-0.5 shrink-0 text-base-content/60" /><span>{{
                selected.address || 'No address provided'
              }}</span>
            </div>
          </section>
          </InspectorSection>
          <InspectorSection v-if="selected.notes" title="Notes">
          <section class="rounded-box border border-base-300 p-3">
            <p class="whitespace-pre-wrap text-sm">{{ selected.notes }}</p>
          </section>
          </InspectorSection>
          <p class="text-xs text-base-content/55">
            Updated {{ formatDateTime(selected.updatedAt) }}
          </p>
        </div>
        <template #footer>
          <div class="flex w-full flex-wrap items-center justify-between gap-2">
            <button class="btn btn-ghost text-error" type="button" @click="remove" :disabled="busy">
              <Trash2 :size="14" />Delete
            </button>
            <button class="btn btn-ghost" type="button" @click="toggle" :disabled="busy">
              {{ selected.active ? 'Archive customer' : 'Reactivate customer' }}
            </button>
          </div>
        </template>
      </InspectorShell>

      <InspectorShell
        v-else
        title="Customer inspector"
        subtitle="Choose a customer from the register."
      >
        <EmptyState title="Select a customer" description="Choose a row to inspect its details." />
      </InspectorShell>
    </MasterDetail>
  </div>
</template>
