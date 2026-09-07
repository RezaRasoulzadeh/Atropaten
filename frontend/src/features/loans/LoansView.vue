<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction,pageLoading,runLoad}=useWorkspaceActions()

import InlineAlert from '../../components/ui/InlineAlert.vue';
import FormGrid from '../../components/ui/FormGrid.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import InspectorSection from '../../components/layout/InspectorSection.vue';
import MasterDetail from '../../components/layout/MasterDetail.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTableRow from '../../components/ui/DataTableRow.vue';
import DataTable from '../../components/ui/DataTable.vue';
import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref } from 'vue';
import { CircleDollarSign, Plus, RotateCcw, X } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import SelectField from '../../components/ui/SelectField.vue';
import { loansApi, type LoanPaymentRecord, type LoanRecord } from '../../api/loans';
import { accountingApi, type FinancialAccountRecord } from '../../api/accounting';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { currentCanonicalDate, formatDateTime } from '../../utils/date';
const props = defineProps<{ currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [string] }>();
const rows = ref<LoanRecord[]>([]);
const accounts = ref<FinancialAccountRecord[]>([]);
const payments = ref<LoanPaymentRecord[]>([]);
const selectedId = ref<string | null>(null);
const tab = ref('Active');
const error = ref('');
const createMode = ref(false);
const form = ref({
  direction: 'payable',
  counterpartyName: '',
  principalRial: '0',
  interestFeeRial: '0',
  startDate: currentCanonicalDate(),
  endDate: '',
  financialAccountId: '',
  installmentCount: '1',
  notes: '',
});
const pay = ref({
  principalRial: '0',
  interestRial: '0',
  financialAccountId: '',
  installmentId: '',
});
const selected = computed(() => rows.value.find((v) => v.id === selectedId.value) ?? null);
const filtered = computed(() =>
  rows.value.filter(
    (v) =>
      tab.value === 'All' ||
      (tab.value === 'Payable' && v.direction === 'payable') ||
      (tab.value === 'Receivable' && v.direction === 'receivable') ||
      (tab.value === 'Closed' && v.status === 'Closed') ||
      (tab.value === 'Overdue' && v.overdueRial > 0) ||
      (tab.value === 'Active' && v.status === 'Active' && v.overdueRial === 0),
  ),
);
onMounted(async () => {
  try {
    accounts.value = await accountingApi.financialAccounts();
    await load();
  } catch (e) {
reportError(e);
    error.value = String(e);
  }
});
async function load() { return runLoad(async () => {
  rows.value = await loansApi.list();
  if (!selectedId.value && rows.value[0]) select(rows.value[0]);
}); }
async function select(v: LoanRecord) {
  selectedId.value = v.id;
  createMode.value = false;
  try {
    payments.value = await loansApi.payments(v.id);
  } catch (e) {
reportError(e);
    error.value = String(e);
  }
}
function begin() {
  createMode.value = true;
  selectedId.value = null;
  form.value = {
    direction: 'payable',
    counterpartyName: '',
    principalRial: '0',
    interestFeeRial: '0',
    startDate: currentCanonicalDate(),
    endDate: '',
    financialAccountId: accounts.value[0]?.id ?? '',
    installmentCount: '1',
    notes: '',
  };
}
async function create() {
return runAction(async () => {
  const principal = Number(form.value.principalRial.replaceAll(',', ''));
  const interest = Number(form.value.interestFeeRial.replaceAll(',', ''));
  if (!form.value.counterpartyName || !principal || !form.value.financialAccountId) {
    error.value = 'Counterparty, account, and positive principal are required.';
    return;
  }
  try {
    const v = await loansApi.create({
      ...form.value,
      principalRial: principal,
      interestFeeRial: interest,
      installmentCount: Number(form.value.installmentCount),
    });
    rows.value = [v, ...rows.value];
    await select(v);
    emit('notify', 'Loan opened with a balanced journal entry.');
  } catch (e) {
reportError(e);
    error.value = String(e);
  }

});
}
async function recordPayment() {
return runAction(async () => {
  if (!selected.value) return;
  const principal = Number(pay.value.principalRial.replaceAll(',', ''));
  const interest = Number(pay.value.interestRial.replaceAll(',', ''));
  if ((!principal && !interest) || !pay.value.financialAccountId || !pay.value.installmentId) {
    error.value = 'Select an installment, account, and payment amount.';
    return;
  }
  try {
    await loansApi.createPayment({
      loanId: selected.value.id,
      financialAccountId: pay.value.financialAccountId,
      paidAt: currentCanonicalDate(),
      amountRial: principal + interest,
      principalRial: principal,
      interestRial: interest,
      allocations: [
        {
          installmentId: pay.value.installmentId,
          principalRial: principal,
          interestRial: interest,
        },
      ],
    });
    await load();
    const v = rows.value.find((x) => x.id === selected.value!.id);
    if (v) await select(v);
    pay.value = { ...pay.value, principalRial: '0', interestRial: '0' };
    emit('notify', 'Loan payment posted and allocated.');
  } catch (e) {
reportError(e);
    error.value = String(e);
  }

});
}
async function reversePayment(v: LoanPaymentRecord) {
return runAction(async () => {
  try {
    await loansApi.reversePayment(v.id);
    if (selected.value) await select(selected.value);
    emit('notify', 'Payment reversed with a compensating entry.');
  } catch (e) {
reportError(e);
    error.value = String(e);
  }

});
}
function tone(s: string) {
  return s === 'Closed' ? 'green' : s === 'Cancelled' ? 'red' : s === 'Active' ? 'blue' : 'amber';
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
        title="Loans"
        eyebrow="Finance / financing"
        description="Payable and receivable loans with persisted schedules, derived overdue balances, and reversible payments."
        ><button class="btn btn-primary" @click="begin">
          <Plus :size="16" /> New loan
        </button></WorkspaceHeader
      >
      <section class="min-w-0 space-y-4">
        <div class="flex flex-wrap items-center gap-2">
          <button
            class="btn btn-sm"
            v-for="value in ['Active', 'Payable', 'Receivable', 'Overdue', 'Closed', 'All']"
            :key="value"
            :class="tab === value ? 'btn-primary' : 'btn-ghost'"
            @click="tab = value"
          >
            {{ value }}
          </button>
        </div>
        <span>{{ filtered.length }} loans</span>
      </section></WorkspaceStickyStack
    ><LoadingState v-if="pageLoading" label="Loading records…" /><div v-show="!pageLoading" class="space-y-4"><InlineAlert v-if="error" role="alert" class="flex flex-wrap items-center gap-2" tone="error"
      >{{ error }}
      <button class="btn btn-ghost" @click="error = ''" aria-label="Dismiss">
        <X :size="14" /></button></InlineAlert
    ><MasterDetail
      ><AppPanel
        title="Loan register"
        subtitle="Principal, interest, remaining, and overdue values come from posted allocations."
        ><div v-if="filtered.length">
          <DataTable
            ><thead>
              <tr>
                <th>Loan</th>
                <th>Counterparty</th>
                <th>Type</th>
                <th class="text-end">Remaining</th>
                <th class="text-end">Overdue</th>
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
                  ><span class="block font-medium">{{ v.loanNumber }}</span
                  ><span class="block text-xs text-base-content/60">{{
                    date(v.startDate)
                  }}</span></DataTableCell
                ><DataTableCell>{{ v.counterpartyName }}</DataTableCell
                ><DataTableCell>{{
                  v.direction === 'payable' ? 'Payable' : 'Receivable'
                }}</DataTableCell
                ><DataTableCell numeric>{{
                  formatMoney(
                    v.remainingPrincipalRial + v.remainingInterestRial,
                    props.currencyUnit,
                  )
                }}</DataTableCell
                ><DataTableCell numeric>{{
                  formatMoney(v.overdueRial, props.currencyUnit)
                }}</DataTableCell
                ><DataTableCell
                  ><StatusBadge :label="v.status" :tone="tone(v.status)" /></DataTableCell
              ></DataTableRow></tbody
          ></DataTable>
        </div>
        <div v-else class="min-w-0 space-y-3">
          <CircleDollarSign :size="22" />
          <p>No loans in this view.</p>
        </div></AppPanel
      ><InspectorShell
        v-if="createMode"
        title="New loan"
        subtitle="Opening the loan posts principal to cash/bank and the loan balance."
        ><form @submit.prevent="create" class="min-w-0 space-y-3">
          <FormGrid
            ><SelectField
              v-model="form.direction"
              label="Type"
              :options="[
                { label: 'Payable / borrowed', value: 'payable' },
                { label: 'Receivable / lent', value: 'receivable' },
              ]" /><FormField class="gap-1"
              ><span>Counterparty</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.counterpartyName"
                required /></FormField
            ><FormField class="gap-1"
              ><span>Principal (Rial)</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.principalRial"
                inputmode="numeric" /></FormField
            ><FormField class="gap-1"
              ><span>Interest / fees (Rial)</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.interestFeeRial"
                inputmode="numeric" /></FormField
            ><FormField class="gap-1"
              ><span>Start date</span><JalaliDatePicker v-model="form.startDate" /></FormField
            ><FormField class="gap-1"
              ><span>Installments</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.installmentCount"
                type="number"
                min="1" /></FormField
            ><SelectField
              v-model="form.financialAccountId"
              label="Cash / bank account"
              :options="
                accounts.map((account) => ({ label: account.name, value: account.id }))
              " /></FormGrid
          ><FormField class="gap-1"
            ><span>Notes</span
            ><AppTextarea
              v-model="form.notes"
              rows="3"
            />
          </FormField>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-ghost" type="button" @click="createMode = false">Cancel</button
            ><button class="btn btn-primary">Open loan</button>
          </div>
        </form></InspectorShell
      ><InspectorShell
        v-else-if="selected"
        title="Loan inspector"
        :subtitle="`${selected.loanNumber} · ${selected.counterpartyName}`"
        ><div class="min-w-0 space-y-3">
          <div class="min-w-0 space-y-3">
            <div><CircleDollarSign :size="19" /></div>
            <div class="min-w-0 space-y-3">
              <h3 class="text-sm font-semibold">
                {{ selected.direction === 'payable' ? 'Payable loan' : 'Receivable loan' }}
              </h3>
              <p>
                {{ date(selected.startDate)
                }}<template v-if="selected.endDate"> → {{ date(selected.endDate) }}</template>
              </p>
            </div>
          </div>
          <StatusBadge :label="selected.status" :tone="tone(selected.status)" />
          <dl class="grid min-w-0 gap-2 text-sm">
            <div
              class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
            >
              <dt class="text-xs text-base-content/60">Principal remaining</dt>
              <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
                {{ formatMoney(selected.remainingPrincipalRial, props.currencyUnit) }}
              </dd>
            </div>
            <div
              class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
            >
              <dt class="text-xs text-base-content/60">Interest remaining</dt>
              <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
                {{ formatMoney(selected.remainingInterestRial, props.currencyUnit) }}
              </dd>
            </div>
            <div
              class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
            >
              <dt class="text-xs text-base-content/60">Overdue</dt>
              <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
                {{ formatMoney(selected.overdueRial, props.currencyUnit) }}
              </dd>
            </div>
          </dl>
          <InspectorSection title="Installment schedule">
            <DataTable v-if="selected.installments.length" label="Loan installment schedule">
              <thead><tr><th scope="col">Installment</th><th scope="col">Due</th><th scope="col" class="text-end">Paid</th><th scope="col" class="text-end">Remaining</th><th scope="col">Status</th></tr></thead>
              <tbody>
                <tr v-for="i in selected.installments" :key="i.id">
                  <DataTableCell><strong>#{{ i.position + 1 }}</strong></DataTableCell>
                  <DataTableCell>{{ date(i.dueDate) }}</DataTableCell>
                  <DataTableCell numeric>{{ formatMoney(i.paidRial, props.currencyUnit) }}</DataTableCell>
                  <DataTableCell numeric>{{ formatMoney(i.remainingRial, props.currencyUnit) }}</DataTableCell>
                  <DataTableCell><StatusBadge :label="i.status" :tone="i.status === 'Paid' ? 'green' : i.status === 'Overdue' ? 'red' : 'amber'" /></DataTableCell>
                </tr>
              </tbody>
            </DataTable>
            <EmptyState v-else title="No installments" description="This loan has no generated schedule." />
          </InspectorSection>
          <div class="min-w-0 space-y-3">
            <h3 class="text-sm font-semibold">Record payment</h3>
            <FormGrid
              ><SelectField
                v-model="pay.installmentId"
                label="Installment"
                :options="[
                  { label: 'Installment', value: '' },
                  ...selected.installments
                    .filter((installment) => installment.remainingRial > 0)
                    .map((installment) => ({
                      label: `#${installment.position + 1} · ${formatMoney(installment.remainingRial, props.currencyUnit)}`,
                      value: installment.id,
                    })),
                ]" /><SelectField
                v-model="pay.financialAccountId"
                label="Cash / bank account"
                :options="[
                  { label: 'Cash / bank account', value: '' },
                  ...accounts.map((account) => ({ label: account.name, value: account.id })),
                ]" /><FormField label="Principal (Rial)"><AppInput
                class="input w-full min-w-0"
                v-model="pay.principalRial"
                placeholder="Principal (Rial)"
                inputmode="numeric" /></FormField><FormField label="Interest (Rial)"><AppInput
                class="input w-full min-w-0"
                v-model="pay.interestRial"
                placeholder="Interest (Rial)"
                inputmode="numeric" /></FormField></FormGrid
            ><button class="btn btn-ghost" @click="recordPayment" :disabled="busy">Record payment</button>
          </div>
          <InspectorSection title="Payment history">
            <DataTable v-if="payments.length" label="Loan payment history">
              <thead><tr><th scope="col">Posted</th><th scope="col" class="text-end">Amount</th><th scope="col">Status</th><th scope="col">Actions</th></tr></thead>
              <tbody>
                <tr v-for="p in payments" :key="p.id">
                  <DataTableCell>{{ date(p.paidAt) }}</DataTableCell>
                  <DataTableCell numeric>{{ formatMoney(p.amountRial, props.currencyUnit) }}</DataTableCell>
                  <DataTableCell><StatusBadge :label="p.status" :tone="p.status === 'Posted' ? 'green' : 'slate'" /></DataTableCell>
                  <DataTableCell><button class="btn btn-ghost btn-sm" v-if="p.status === 'Posted'" @click="reversePayment(p)" :disabled="busy"><RotateCcw :size="14" /> Reverse</button></DataTableCell>
                </tr>
              </tbody>
            </DataTable>
            <EmptyState v-else title="No payments" description="Payments recorded against this loan will appear here." />
          </InspectorSection>
        </div></InspectorShell
      ><InspectorShell
        v-else
        title="Loan inspector"
        subtitle="Select a loan to inspect its schedule."
        ><div>No loan selected.</div></InspectorShell
      ></MasterDetail
    >
  </div></div>
</template>
