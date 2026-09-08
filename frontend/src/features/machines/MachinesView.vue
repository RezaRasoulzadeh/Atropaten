<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import InlineAlert from '../../components/ui/InlineAlert.vue';
import FormGrid from '../../components/ui/FormGrid.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import InspectorSection from '../../components/layout/InspectorSection.vue';
import MasterDetail from '../../components/layout/MasterDetail.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import FormField from '../../components/ui/FormField.vue';
import { computed, onMounted, ref, watch } from 'vue';
import { Archive, Edit3, Factory, Plus, RotateCcw, Save, X } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';
import { machinesApi, type MachineRecord, type MachinePayload } from '../../api/machines';
import {
  formatMoney,
  formatMoneyInput,
  parseMoneyInput,
  type CurrencyUnit,
} from '../../utils/currency';
import { formatDateTime } from '../../utils/date';

const props = defineProps<{ currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [message: string] }>();
type Filter = 'Active' | 'Archived' | 'All';
type Mode = 'create' | 'edit' | null;
type MachineForm = Omit<MachinePayload, 'rateRial' | 'setupCostRial'> & {
  rate: string;
  setupCost: string;
};

const machines = ref<MachineRecord[]>([]);
const selectedId = ref<string | null>(null);
const filter = ref<Filter>('Active');
const query = ref('');
const mode = ref<Mode>(null);
const form = ref<MachineForm>(emptyForm());
const loading = ref(false);
const saving = ref(false);
const error = ref('');
const formError = ref('');
const selectedMachine = computed(
  () => machines.value.find((machine) => machine.id === selectedId.value) ?? null,
);
const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  return machines.value.filter(
    (machine) =>
      (filter.value === 'All' || (filter.value === 'Active' ? machine.active : !machine.active)) &&
      (!q ||
        [machine.name, machine.code, machine.category, machine.rateBasis].some((value) =>
          value.toLowerCase().includes(q),
        )),
  );
});

onMounted(load);
watch(
  () => props.currencyUnit,
  () => {
    if (mode.value) {
      const machine = selectedMachine.value;
      if (machine) {
        form.value.rate = formatMoneyInput(machine.rateRial, props.currencyUnit);
        form.value.setupCost = formatMoneyInput(machine.setupCostRial, props.currencyUnit);
      }
    }
  },
);
watch(
  () => [form.value.rate, form.value.setupCost],
  ([rateValue, setupValue]) => {
    const parsedRate = parseMoneyInput(rateValue, props.currencyUnit);
    const parsedSetup = parseMoneyInput(setupValue, props.currencyUnit);
    if (parsedRate !== null) form.value.rate = formatMoneyInput(parsedRate, props.currencyUnit);
    if (parsedSetup !== null)
      form.value.setupCost = formatMoneyInput(parsedSetup, props.currencyUnit);
  },
);
function emptyForm(): MachineForm {
  return {
    name: '',
    code: '',
    category: '',
    rateBasis: 'hour',
    rate: '',
    setupCost: '',
    notes: '',
  };
}
function load() {
  loading.value = true;
  error.value = '';
  machinesApi
    .list(true)
    .then((data) => {
      machines.value = data;
      if (!selectedId.value && data.length) selectedId.value = data[0].id;
    })
    .catch((e) => {
      error.value = message(e, 'Machines could not be loaded.');
    })
    .finally(() => {
      loading.value = false;
    });
}
function select(id: string) {
  selectedId.value = id;
  mode.value = null;
  formError.value = '';
}
function startCreate() {
  mode.value = 'create';
  selectedId.value = null;
  form.value = emptyForm();
  formError.value = '';
}
function startEdit() {
  const machine = selectedMachine.value;
  if (!machine) return;
  form.value = {
    name: machine.name,
    code: machine.code,
    category: machine.category,
    rateBasis: machine.rateBasis,
    rate: formatMoneyInput(machine.rateRial, props.currencyUnit),
    setupCost: formatMoneyInput(machine.setupCostRial, props.currencyUnit),
    notes: machine.notes,
  };
  mode.value = 'edit';
  formError.value = '';
}
function cancel() {
  mode.value = null;
  formError.value = '';
}
function rate(value: string) {
  return parseMoneyInput(value, props.currencyUnit);
}
function date(value: string) {
  try {
    return formatDateTime(value);
  } catch {
    return 'Unknown date';
  }
}
function basisLabel(value: string) {
  return (
    ({ unit: 'Per unit / page', minute: 'Per minute', hour: 'Per hour' } as Record<string, string>)[
      value
    ] ?? value
  );
}
async function save() {
return runAction(async () => {
  formError.value = '';
  const parsedRate = rate(form.value.rate);
  const parsedSetup = form.value.setupCost.trim() === '' ? 0 : rate(form.value.setupCost);
  if (!form.value.name.trim()) {
    formError.value = 'Enter a machine name.';
    return;
  }
  if (parsedRate === null || parsedSetup === null) {
    formError.value = `Enter whole ${props.currencyUnit.toLowerCase()} amounts.`;
    return;
  }
  saving.value = true;
  const wasEditing = mode.value === 'edit';
  try {
    const payload = {
      name: form.value.name,
      code: form.value.code,
      category: form.value.category,
      rateBasis: form.value.rateBasis,
      rateRial: parsedRate,
      setupCostRial: parsedSetup,
      notes: form.value.notes,
    };
    const saved =
      mode.value === 'edit' && selectedId.value
        ? await machinesApi.update(selectedId.value, payload)
        : await machinesApi.create(payload);
    const index = machines.value.findIndex((item) => item.id === saved.id);
    if (index >= 0) machines.value.splice(index, 1, saved);
    else machines.value.push(saved);
    selectedId.value = saved.id;
    mode.value = null;
    emit('notify', wasEditing ? 'Machine updated.' : 'Machine created.');
  } catch (e) {
reportError(e);
    formError.value = message(e, 'Machine could not be saved.');
  } finally {
    saving.value = false;
  }

});
}
async function setActive(active: boolean) {
return runAction(async () => {
  const machine = selectedMachine.value;
  if (!machine) return;
  try {
    const updated = active
      ? await machinesApi.reactivate(machine.id)
      : await machinesApi.archive(machine.id);
    const index = machines.value.findIndex((item) => item.id === updated.id);
    if (index >= 0) machines.value.splice(index, 1, updated);
    emit('notify', active ? 'Machine reactivated.' : 'Machine archived.');
  } catch (e) {
reportError(e);
    error.value = message(e, 'Machine status could not be changed.');
  }

});
}
function message(errorValue: unknown, fallback: string) {
  return errorValue instanceof Error && errorValue.message
    ? errorValue.message
    : typeof errorValue === 'string'
      ? errorValue
      : fallback;
}
</script>

<template>
  <div class="min-w-0 space-y-3">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        title="Machines"
        eyebrow="Catalog / production inputs"
        description="Keep reusable equipment rates ready for service cost definitions."
        ><button class="btn btn-primary" type="button" @click="startCreate">
          <Plus :size="16" :stroke-width="1.8" aria-hidden="true" />New machine
        </button></WorkspaceHeader
      >
      <SearchFilterBar><template #search><SearchField
          v-model="query"
          label="Search machines"
          placeholder="Search machine, code, or category"
        /></template><template #filters><SelectField
          v-model="filter"
          label="Status"
          aria-label="Filter machines by status"
          :options="['Active', 'Archived', 'All'].map((value) => ({ label: value, value }))"
        /></template><template #count><span class="self-end pb-2">{{ filtered.length }} of {{ machines.length }} machines</span></template></SearchFilterBar>
    </WorkspaceStickyStack>
    <InlineAlert v-if="error" role="alert" class="min-w-0 space-y-3" tone="error"
      ><span>{{ error }}</span
      ><button
        class="btn btn-ghost"
        type="button"
        aria-label="Dismiss machines error"
        @click="error = ''"
      >
        <X :size="15" :stroke-width="1.8" aria-hidden="true" /></button
    ></InlineAlert>
    <MasterDetail aria-label="Machines workspace">
      <RegisterList
        title="Machine register"
        subtitle="Rates are stored as integer Rial; the toolbar controls display units."
        :count="filtered.length"
      >
        <LoadingState v-if="loading" label="Loading records…" />
        <div v-else-if="filtered.length">
          <RegisterRow
            v-for="machine in filtered"
            :key="machine.id"
            :selected="selectedId === machine.id"
            @activate="select(machine.id)"
          >
            <template #identity>
              <div class="flex min-w-0 items-center justify-between gap-3">
                <div class="min-w-0">
                  <strong class="block truncate text-sm">{{ machine.name }}</strong>
                  <span class="block truncate text-xs text-base-content/60">{{ machine.code || 'No code' }} · {{ machine.category || 'Uncategorized' }}</span>
                </div>
                <strong class="shrink-0 whitespace-nowrap text-sm tabular-nums">{{ formatMoney(machine.rateRial, props.currencyUnit) }}</strong>
              </div>
            </template>
            <template #meta>
              <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2">
                <div><span class="block text-base-content/50">Rate basis</span><span class="block text-base-content/80">{{ basisLabel(machine.rateBasis) }}</span></div>
                <div><span class="block text-base-content/50">Setup cost</span><span class="block text-base-content/80 tabular-nums">{{ formatMoney(machine.setupCostRial, props.currencyUnit) }}</span></div>
                <div><span class="block text-base-content/50">Notes</span><span class="block truncate text-base-content/80">{{ machine.notes || 'No notes' }}</span></div>
              </div>
            </template>
            <template #status><StatusBadge :label="machine.active ? 'Active' : 'Archived'" :tone="machine.active ? 'green' : 'slate'" /></template>
          </RegisterRow>
        </div>
        <div v-else class="min-w-0 space-y-3">
          <div><Factory :size="21" :stroke-width="1.8" /></div>
          <h2 class="text-base font-semibold">
            {{ machines.length ? 'No machines match this view' : 'No machines yet' }}
          </h2>
          <p>
            {{
              machines.length
                ? 'Try another status or search term.'
                : 'Add the first reusable rate input for production.'
            }}
          </p>
          <button
            class="btn btn-primary"
            v-if="!machines.length"
            type="button"
            @click="startCreate"
          >
            <Plus :size="15" :stroke-width="1.8" aria-hidden="true" />Create machine
          </button>
        </div>
      </RegisterList>
      <InspectorShell
        v-if="mode"
        :title="mode === 'create' ? 'New machine' : 'Edit machine'"
        subtitle="Save a reusable machine rate definition."
        ><template #action
          ><button
            class="btn btn-ghost"
            type="button"
            aria-label="Close machine editor"
            @click="cancel"
          >
            <X :size="16" :stroke-width="1.8" aria-hidden="true" /></button
        ></template>
        <form @submit.prevent="save" class="min-w-0 space-y-3">
          <div v-if="formError" role="alert">{{ formError }}</div>
          <FormField class="gap-1"
            ><span>Name</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="form.name"
              type="text"
              placeholder="Production printer" /></FormField
          ><FormGrid
            ><FormField class="gap-1"
              ><span>Code</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.code"
                type="text"
                placeholder="PRINTER-01" /></FormField
            ><FormField class="gap-1"
              ><span>Category</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.category"
                type="text"
                placeholder="Print production" /></FormField></FormGrid
          ><FormGrid
            ><SelectField
              v-model="form.rateBasis"
              label="Rate basis"
              :options="[
                { label: 'Per unit / page', value: 'unit' },
                { label: 'Per minute', value: 'minute' },
                { label: 'Per hour', value: 'hour' },
              ]" /><FormField class="gap-1"
              ><span>Rate ({{ props.currencyUnit }})</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.rate"
                type="text"
                inputmode="decimal"
                placeholder="0" /></FormField></FormGrid
          ><FormField class="gap-1"
            ><span>Setup / fixed cost ({{ props.currencyUnit }})</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="form.setupCost"
              type="text"
              inputmode="decimal"
              placeholder="Optional" /></FormField
          ><FormField class="gap-1"
            ><span>Notes</span
            ><AppTextarea
              v-model="form.notes"
              rows="3"
              placeholder="Capacity, operating notes, or rate context"
            />
          </FormField>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-ghost" type="button" @click="cancel">Cancel</button
            ><button class="btn btn-primary" type="submit" :disabled="busy || (saving)">
              <Save :size="15" :stroke-width="1.8" aria-hidden="true" />{{
                saving ? 'Saving…' : 'Save machine'
              }}
            </button>
          </div>
        </form></InspectorShell
      >
      <InspectorShell
        v-else-if="selectedMachine"
        title="Machine inspector"
        subtitle="Current persisted rate definition"
        ><template #action
          ><button
            class="btn btn-ghost"
            type="button"
            aria-label="Edit selected machine"
            @click="startEdit"
          >
            <Edit3 :size="15" :stroke-width="1.8" aria-hidden="true" /></button
        ></template>
        <div class="flex min-w-0 flex-wrap items-center gap-2">
          <StatusBadge
            :label="selectedMachine.active ? 'Active' : 'Archived'"
            :tone="selectedMachine.active ? 'green' : 'slate'"
          /><span>{{ basisLabel(selectedMachine.rateBasis) }}</span>
        </div>
        <div class="min-w-0 space-y-3">
          <div><Factory :size="19" :stroke-width="1.8" aria-hidden="true" /></div>
          <div class="min-w-0 space-y-3">
            <h3 class="text-sm font-semibold">{{ selectedMachine.name }}</h3>
            <p>
              {{ selectedMachine.code || 'No code'
              }}<span v-if="selectedMachine.category"> · {{ selectedMachine.category }}</span>
            </p>
          </div>
        </div>
        <InspectorSection title="Rate definition">
        <dl class="grid min-w-0 gap-2 text-sm">
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Rate</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ formatMoney(selectedMachine.rateRial, props.currencyUnit) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Setup / fixed</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ formatMoney(selectedMachine.setupCostRial, props.currencyUnit) }}
            </dd>
          </div>
        </dl>
        </InspectorSection>
        <p v-if="selectedMachine.notes">{{ selectedMachine.notes }}</p>
        <div>Updated {{ date(selectedMachine.updatedAt) }}</div>
        <div class="flex flex-wrap items-center gap-2">
          <button
            class="btn btn-ghost"
            v-if="selectedMachine.active"
            type="button"
            @click="setActive(false)"
           :disabled="busy">
            <Archive :size="15" :stroke-width="1.8" aria-hidden="true" />Archive</button
          ><button class="btn btn-ghost" v-else type="button" @click="setActive(true)" :disabled="busy">
            <RotateCcw :size="15" :stroke-width="1.8" aria-hidden="true" />Reactivate
          </button>
        </div></InspectorShell
      >
      <InspectorShell v-else title="Machine inspector" subtitle="Select a row to inspect it."
        ><div class="min-w-0 space-y-3">
          <Factory :size="20" :stroke-width="1.8" aria-hidden="true" />
          <p>Machine details will appear here.</p>
          <button class="btn btn-primary" type="button" @click="startCreate">
            Create a machine <Plus :size="14" :stroke-width="1.8" aria-hidden="true" />
          </button></div
      ></InspectorShell>
    </MasterDetail>
  </div>
</template>
