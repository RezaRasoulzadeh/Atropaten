<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction,pageLoading,runLoad}=useWorkspaceActions()

import InlineAlert from '../../components/ui/InlineAlert.vue';
import FormGrid from '../../components/ui/FormGrid.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import MasterDetail from '../../components/layout/MasterDetail.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTableRow from '../../components/ui/DataTableRow.vue';
import DataTable from '../../components/ui/DataTable.vue';
import AppInput from '../../components/ui/AppInput.vue';
import FormField from '../../components/ui/FormField.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref } from 'vue';
import { Landmark, Plus, RotateCcw, Trash2, X } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import { checksApi, type CheckEventRecord, type CheckRecord } from '../../api/checks';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { currentCanonicalDate, formatDateTime } from '../../utils/date';
import { normalizeError } from '../../ui/feedback';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';

const props = defineProps<{ currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [string] }>();
const rows = ref<CheckRecord[]>([]);
const selectedId = ref<string | null>(null);
const tab = ref('All');
const query = ref('');
const error = ref('');
const history = ref<CheckEventRecord[]>([]);
const createMode = ref(false);
const form = ref({
  direction: 'incoming',
  checkNumber: '',
  bank: '',
  branch: '',
  accountDescriptor: '',
  payerPayee: '',
  amountRial: '0',
  issueDate: currentCanonicalDate(),
  dueDate: currentCanonicalDate(),
  customerId: '',
  supplierId: '',
  financialAccountId: '',
  notes: '',
});
const selected = computed(() => rows.value.find((v) => v.id === selectedId.value) ?? null);
const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  return rows.value.filter((v) => {
    const due =
      Date.parse(v.dueDate) < Date.now() &&
      !['Cleared', 'Returned', 'Cancelled', 'Rejected'].includes(v.status);
    const match =
      tab.value === 'All' ||
      (tab.value === 'Incoming' && v.direction === 'incoming') ||
      (tab.value === 'Outgoing' && v.direction === 'outgoing') ||
      (tab.value === 'Due' &&
        Date.parse(v.dueDate) >= Date.now() &&
        !['Cleared', 'Returned', 'Cancelled', 'Rejected'].includes(v.status)) ||
      (tab.value === 'Overdue' && due) ||
      (tab.value === 'Cleared' && v.status === 'Cleared') ||
      (tab.value === 'Closed' && ['Returned', 'Cancelled', 'Rejected'].includes(v.status));
    return (
      match && (!q || [v.checkNumber, v.bank, v.payerPayee].join(' ').toLowerCase().includes(q))
    );
  });
});
const nextStatuses = computed(() => {
  if (!selected.value) return [];
  const incoming: { [key: string]: string[] } = {
    Draft: ['Received', 'Cancelled'],
    Received: ['Deposited', 'Returned', 'Cancelled'],
    Deposited: ['Cleared', 'Returned'],
    Cleared: ['Returned'],
    Returned: ['Cancelled'],
  };
  const outgoing: { [key: string]: string[] } = {
    Draft: ['Issued', 'Cancelled'],
    Issued: ['Delivered', 'Returned', 'Rejected', 'Cancelled'],
    Delivered: ['Cleared', 'Returned', 'Rejected'],
    Cleared: ['Returned', 'Rejected'],
    Returned: ['Cancelled'],
    Rejected: ['Cancelled'],
  };
  return (
    (selected.value.direction === 'incoming' ? incoming : outgoing)[selected.value.status] ?? []
  );
});
onMounted(load);
async function load() { return runLoad(async () => {
  try {
    rows.value = await checksApi.list();
    if (!selectedId.value && rows.value[0]) select(rows.value[0]);
  } catch (e) {
    error.value = normalizeError(e).message;
  }
}); }
async function select(v: CheckRecord) {
  selectedId.value = v.id;
  createMode.value = false;
  try {
    history.value = await checksApi.history(v.id);
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }
}
function begin() {
  createMode.value = true;
  selectedId.value = null;
  form.value = {
    ...form.value,
    direction: 'incoming',
    checkNumber: '',
    bank: '',
    branch: '',
    accountDescriptor: '',
    payerPayee: '',
    amountRial: '0',
    issueDate: currentCanonicalDate(),
    dueDate: currentCanonicalDate(),
    notes: '',
  };
}
async function create() {
return runAction(async () => {
  const amount = Number(form.value.amountRial.replaceAll(',', ''));
  if (!form.value.checkNumber || !form.value.bank || !form.value.payerPayee || !amount) {
    error.value = 'Check number, bank, payer/payee, and a positive amount are required.';
    return;
  }
  try {
    const v = await checksApi.create({ ...form.value, amountRial: amount });
    rows.value = [v, ...rows.value];
    await select(v);
    emit('notify', 'Check draft created.');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
async function transition(to: string) {
return runAction(async () => {
  if (!selected.value) return;
  try {
    const v = await checksApi.transition(selected.value.id, to);
    rows.value = rows.value.map((x) => (x.id === v.id ? v : x));
    await select(v);
    emit('notify', `Check moved to ${to}.`);
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
async function removeDraft() {
return runAction(async () => {
  if (!selected.value || selected.value.status !== 'Draft') return;
  try {
    await checksApi.deleteDraft(selected.value.id);
    rows.value = rows.value.filter((v) => v.id !== selected.value!.id);
    selectedId.value = null;
    history.value = [];
    emit('notify', 'Draft check deleted.');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
  }

});
}
function tone(s: string) {
  return s === 'Cleared'
    ? 'green'
    : s === 'Returned' || s === 'Rejected' || s === 'Cancelled'
      ? 'red'
      : s === 'Deposited' || s === 'Delivered'
        ? 'blue'
        : 'amber';
}
function date(v: string) {
  try {
    return formatDateTime(v);
  } catch {
    return '—';
  }
}
</script>
<template>
  <div class="min-w-0 space-y-3">
    <WorkspaceStickyStack
      ><WorkspaceHeader
        title="Checks"
        eyebrow="Finance / instruments"
        description="Controlled check lifecycles with uncleared instruments kept outside available bank cash."
        ><button class="btn btn-primary" @click="begin">
          <Plus :size="16" /> New check
        </button></WorkspaceHeader
      >
      <section class="min-w-0 space-y-4">
        <SearchField
          v-model="query"
          label="Search checks"
          placeholder="Search number, bank, or party"
        />
        <div class="flex flex-wrap items-center gap-2">
          <button
            class="btn btn-sm"
            v-for="value in ['Incoming', 'Outgoing', 'Due', 'Overdue', 'Cleared', 'Closed', 'All']"
            :key="value"
            :class="tab === value ? 'btn-primary' : 'btn-ghost'"
            @click="tab = value"
          >
            {{ value }}
          </button>
        </div>
        <span>{{ filtered.length }} shown</span>
      </section></WorkspaceStickyStack
    ><LoadingState v-if="pageLoading" label="Loading records…" /><div v-show="!pageLoading" class="space-y-4">
    <InlineAlert v-if="error" role="alert" class="flex flex-wrap items-center gap-2" tone="error"
      >{{ error }}
      <button class="btn btn-ghost" @click="error = ''" aria-label="Dismiss">
        <X :size="14" /></button
    ></InlineAlert>
    <MasterDetail
      ><AppPanel title="Check register" subtitle="Only valid next lifecycle actions are offered."
        ><div v-if="filtered.length">
          <DataTable
            ><thead>
              <tr>
                <th>Check</th>
                <th>Direction</th>
                <th>Party</th>
                <th>Due</th>
                <th class="text-end">Amount</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <DataTableRow
                v-for="v in filtered"
                :key="v.id"
                :class="{ 'bg-base-300': selectedId === v.id }"
                @activate="select(v)"
                interactive
                ><DataTableCell
                  ><span class="block font-medium">{{ v.checkNumber }}</span
                  ><span class="block text-xs text-base-content/60">{{
                    v.bank
                  }}</span></DataTableCell
                ><DataTableCell>{{
                  v.direction === 'incoming' ? 'Incoming' : 'Outgoing'
                }}</DataTableCell
                ><DataTableCell>{{ v.payerPayee }}</DataTableCell
                ><DataTableCell>{{ date(v.dueDate) }}</DataTableCell
                ><DataTableCell numeric>{{
                  formatMoney(v.amountRial, props.currencyUnit)
                }}</DataTableCell
                ><DataTableCell
                  ><StatusBadge :label="v.status" :tone="tone(v.status)" /></DataTableCell
              ></DataTableRow></tbody
          ></DataTable>
        </div>
        <div v-else class="min-w-0 space-y-3">
          <Landmark :size="22" />
          <p>No checks in this view.</p>
        </div></AppPanel
      >
      <InspectorShell
        v-if="createMode"
        title="New check"
        subtitle="Drafts have no financial effect until lifecycle posting."
        ><form @submit.prevent="create" class="min-w-0 space-y-3">
          <FormGrid
            ><FormField class="gap-1"
              ><span>Direction</span
              ><SelectField
                v-model="form.direction"
                label="Direction"
                :options="[
                  { label: 'Incoming', value: 'incoming' },
                  { label: 'Outgoing', value: 'outgoing' },
                ]" /></FormField
            ><FormField class="gap-1"
              ><span>Check number</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.checkNumber"
                required /></FormField
            ><FormField class="gap-1"
              ><span>Bank</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.bank"
                required /></FormField
            ><FormField class="gap-1"
              ><span>Payer / payee</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.payerPayee"
                required /></FormField
            ><FormField class="gap-1"
              ><span>Amount (Rial)</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.amountRial"
                inputmode="numeric"
                required /></FormField
            ><FormField class="gap-1"
              ><span>Branch</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.branch" /></FormField
            ><FormField class="gap-1"
              ><span>Issue date</span><JalaliDatePicker v-model="form.issueDate" /></FormField
            ><FormField class="gap-1"
              ><span>Due date</span
              ><JalaliDatePicker v-model="form.dueDate" /></FormField></FormGrid
          ><FormField class="gap-1"
            ><span>Notes</span
            ><textarea
              class="textarea w-full min-w-0"
              v-model="form.notes"
              rows="3"
            />
          </FormField>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-ghost" type="button" @click="createMode = false">Cancel</button
            ><button class="btn btn-primary"><Plus :size="15" /> Save draft</button>
          </div>
        </form></InspectorShell
      >
      <InspectorShell
        v-else-if="selected"
        title="Check inspector"
        :subtitle="`${selected.checkNumber} · ${selected.payerPayee}`"
        ><div class="min-w-0 space-y-3">
          <div class="min-w-0 space-y-3">
            <div><Landmark :size="19" /></div>
            <div class="min-w-0 space-y-3">
              <h3 class="text-sm font-semibold">{{ selected.bank }}</h3>
              <p>
                {{ selected.direction === 'incoming' ? 'Incoming' : 'Outgoing' }} · due
                {{ date(selected.dueDate) }}
              </p>
            </div>
          </div>
          <StatusBadge :label="selected.status" :tone="tone(selected.status)" />
          <dl class="grid min-w-0 gap-2 text-sm">
            <div
              class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
            >
              <dt class="text-xs text-base-content/60">Amount</dt>
              <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
                {{ formatMoney(selected.amountRial, props.currencyUnit) }}
              </dd>
            </div>
            <div
              class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
            >
              <dt class="text-xs text-base-content/60">Account</dt>
              <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
                {{ selected.financialAccountId || 'Default bank' }}
              </dd>
            </div>
          </dl>
          <div class="flex flex-wrap items-center gap-2">
            <button
              v-for="value in nextStatuses"
              :key="value"
              :class="value === 'Cancelled' ? 'btn-error' : 'btn-outline'"
              @click="transition(value)"
              class="btn btn-sm"
             :disabled="busy">
              <RotateCcw :size="14" /> {{ value }}
            </button>
          </div>
          <div class="min-w-0 space-y-3">
            <h3 class="text-sm font-semibold">Lifecycle history</h3>
            <div
              v-for="event in history"
              :key="event.id"
              class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
            >
              <span
                ><strong>{{ event.fromStatus }} → {{ event.toStatus }}</strong
                ><small class="block text-xs leading-5 text-base-content/60"
                  >{{ date(event.occurredAt)
                  }}<template v-if="event.note"> · {{ event.note }}</template></small
                ></span
              ><small class="block text-xs leading-5 text-base-content/60">{{
                event.journalEntryId ? 'Journaled' : 'State only'
              }}</small>
            </div>
            <div v-if="!history.length">No lifecycle events yet.</div>
          </div>
        </div></InspectorShell
      ><InspectorShell
        v-else
        title="Check inspector"
        subtitle="Select a check to inspect lifecycle and history."
        ><div>No check selected.</div></InspectorShell
      ></MasterDetail
    >
  </div></div>
</template>
