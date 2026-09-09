<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import FormField from '../../components/ui/FormField.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction,pageLoading,runLoad}=useWorkspaceActions()

import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import RegisterList from '../../components/ui/RegisterList.vue'
import RegisterRow from '../../components/ui/RegisterRow.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import AppInput from '../../components/ui/AppInput.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref } from 'vue';
import { ArrowRightLeft, Edit3, Plus, ReceiptText, RotateCcw, Trash2 } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import {
  accountingApi,
  type AccountRecord,
  type ExpenseRecord,
  type FinancialAccountRecord,
  type TransferRecord,
} from '../../api/accounting';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../../utils/currency';
import { currentCanonicalDate, formatDateTime } from '../../utils/date';
import SelectField from '../../components/ui/SelectField.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import { purchasesApi, type PurchaseRecord } from '../../api/purchases';
import { confirmAction } from '../../ui/feedback';

const props = defineProps<{
  tab: string;
  currencyUnit: CurrencyUnit;
  accounts: AccountRecord[];
  financial: FinancialAccountRecord[];
  suppliers: any[];
}>();
const emit = defineEmits<{ notify: [string] }>();
const expenses = ref<ExpenseRecord[]>([]);
const purchases = ref<PurchaseRecord[]>([]);
const transfers = ref<TransferRecord[]>([]);
const editingExpense = ref<ExpenseRecord | null>(null);
const historyQuery = ref('');
const historyType = ref<'All' | 'Expenses' | 'Purchases'>('All');
const historyStatus = ref('All');

type HistoryEntry = {
  kind: 'Expense' | 'Purchase';
  id: string;
  number: string;
  description: string;
  accountId: string;
  date: string;
  amount: number;
  status: string;
  expense?: ExpenseRecord;
};
const expense = ref({
  expenseDate: '',
  categoryAccountId: 'ACC-EXP-OTHER',
  financialAccountId: 'FIN-CASH',
  description: '',
  payee: '',
  supplierId: '',
  amount: '',
  paymentMethod: 'cash',
});
const transfer = ref({
  sourceFinancialAccountId: 'FIN-CASH',
  destinationFinancialAccountId: 'FIN-BANK',
  amount: '',
  reference: '',
  transferDate: currentCanonicalDate(),
});
async function load() { return runLoad(async () => {
  try {
    [expenses.value, purchases.value, transfers.value] = await Promise.all([
      accountingApi.expenses(),
      purchasesApi.list(),
      accountingApi.transfers(),
    ]);
    const activeAccounts = props.financial.filter((account) => account.active);
    if (activeAccounts.length && !activeAccounts.some((account) => account.id === expense.value.financialAccountId)) {
      expense.value.financialAccountId = activeAccounts[0].id;
    }
    if (activeAccounts.length) {
      if (!activeAccounts.some((account) => account.id === transfer.value.sourceFinancialAccountId)) transfer.value.sourceFinancialAccountId = activeAccounts[0].id;
      if (!activeAccounts.some((account) => account.id === transfer.value.destinationFinancialAccountId)) transfer.value.destinationFinancialAccountId = activeAccounts.find((account) => account.id !== transfer.value.sourceFinancialAccountId)?.id || activeAccounts[0].id;
    }
  } catch (e) {
    reportError(e);
  }
}); }
onMounted(load);
function accountLabel(account: FinancialAccountRecord) {
  const identity = account.type === 'bank'
    ? [account.bankName, account.accountNumber && `Account ${account.accountNumber}`, account.cardNumber && `Card ${account.cardNumber}`].filter(Boolean).join(' · ')
    : 'Cash account';
  return identity ? `${account.name} · ${identity}` : account.name;
}
function financialAccountName(id: string) {
  const account = props.financial.find((value) => value.id === id);
  return account ? accountLabel(account) : 'Unknown account';
}
const historyEntries = computed<HistoryEntry[]>(() => [
  ...expenses.value.map((value) => ({
    kind: 'Expense' as const,
    id: value.id,
    number: value.expenseNumber,
    description: value.description || 'Expense',
    accountId: value.financialAccountId,
    date: value.expenseDate,
    amount: value.amountRial,
    status: value.status,
    expense: value,
  })),
  ...purchases.value.map((value) => ({
    kind: 'Purchase' as const,
    id: value.id,
    number: value.purchaseNumber,
    description: value.supplierName ? `Purchase from ${value.supplierName}` : 'Purchase',
    accountId: value.financialAccountId,
    date: value.purchaseDate,
    amount: value.totalRial,
    status: value.status,
  })),
].sort((a, b) => b.date.localeCompare(a.date)));
const filteredHistory = computed(() => {
  const query = historyQuery.value.trim().toLowerCase();
  return historyEntries.value.filter((entry) => {
    const matchesType = historyType.value === 'All' || entry.kind === historyType.value.slice(0, -1);
    const matchesStatus = historyStatus.value === 'All' || entry.status === historyStatus.value;
    const searchable = [entry.number, entry.description, entry.kind, financialAccountName(entry.accountId)].join(' ').toLowerCase();
    return matchesType && matchesStatus && (!query || searchable.includes(query));
  });
});
function clearHistoryFilters() {
  historyQuery.value = '';
  historyType.value = 'All';
  historyStatus.value = 'All';
}
function historyStatusTone(status: string) {
  return status === 'Posted' ? 'green' : status === 'Draft' ? 'amber' : 'slate';
}
function resetExpenseForm() {
  editingExpense.value = null;
  expense.value = { expenseDate: '', categoryAccountId: 'ACC-EXP-OTHER', financialAccountId: 'FIN-CASH', description: '', payee: '', supplierId: '', amount: '', paymentMethod: 'cash' };
  const activeAccounts = props.financial.filter((account) => account.active);
  if (activeAccounts.length) expense.value.financialAccountId = activeAccounts[0].id;
}
function startExpenseEdit(value: ExpenseRecord) {
  editingExpense.value = value;
  expense.value = {
    expenseDate: value.expenseDate,
    categoryAccountId: value.categoryAccountId,
    financialAccountId: value.financialAccountId,
    description: value.description,
    payee: value.payee,
    supplierId: value.supplierId,
    amount: formatMoneyInput(value.amountRial, props.currencyUnit),
    paymentMethod: value.paymentMethod,
  };
}
function cancelExpenseEdit() {
  resetExpenseForm();
}
async function removeExpense(value: ExpenseRecord) {
  return runAction(async () => {
    if (!(await confirmAction({
      title: 'Remove expense',
      message: `Remove ${value.expenseNumber}? The accounting effect will be reversed and retained in the ledger history.`,
      confirmLabel: 'Remove expense',
      danger: true,
    }))) return;
    try {
      await accountingApi.deleteExpense(value.id);
      expenses.value = expenses.value.filter((entry) => entry.id !== value.id);
      if (editingExpense.value?.id === value.id) resetExpenseForm();
      emit('notify', 'Expense removed.');
    } catch (e) {
      reportError(e);
    }
  });
}
async function createExpense() {
  return runAction(async () => {
    try {
      const amount = parseMoneyInput(expense.value.amount, props.currencyUnit);
      if (!amount || amount <= 0) throw new Error('Enter a positive amount');
      if (!props.financial.some((account) => account.id === expense.value.financialAccountId && account.active)) throw new Error('Select an active cash or bank account');
      const payload = {
        expenseDate: expense.value.expenseDate || undefined,
        categoryAccountId: expense.value.categoryAccountId,
        financialAccountId: expense.value.financialAccountId,
        description: expense.value.description,
        payee: expense.value.payee,
        supplierId: expense.value.supplierId,
        paymentMethod: expense.value.paymentMethod,
        amountRial: amount,
      };
      const v = editingExpense.value
        ? await accountingApi.updateExpense(editingExpense.value.id, payload)
        : await accountingApi.createExpense(payload);
      const wasEditing = Boolean(editingExpense.value);
      expenses.value = wasEditing
        ? expenses.value.map((entry) => (entry.id === v.id ? v : entry))
        : [v, ...expenses.value];
      resetExpenseForm();
      emit('notify', wasEditing ? 'Expense updated.' : 'Expense posted.');
    } catch (e) {
      reportError(e);
    }
  });
}
async function createTransfer() {
return runAction(async () => {
  try {
    const amount = parseMoneyInput(transfer.value.amount, props.currencyUnit);
    if (!amount || amount <= 0) throw new Error('Enter a positive amount');
    const v = await accountingApi.createTransfer({ ...transfer.value, amountRial: amount });
    transfers.value = [v, ...transfers.value];
    transfer.value = { ...transfer.value, amount: '', reference: '' };
    emit('notify', 'Transfer posted.');
  } catch (e) {
reportError(e);
  }

});
}
async function reverseTransfer(v: TransferRecord) {
return runAction(async () => {
  try {
    const next = await accountingApi.reverseTransfer(v.id);
    transfers.value = transfers.value.map((x) => (x.id === next.id ? next : x));
    emit('notify', 'Transfer reversed.');
  } catch (e) {
reportError(e);
  }

});
}
</script>
<template><div class="min-w-0 space-y-4"><LoadingState v-if="pageLoading" label="Loading records…" /><div v-show="!pageLoading" class="min-w-0 space-y-4">
<div class="flex min-w-0 items-start gap-3 rounded-box border border-base-300 bg-base-200/35 p-4">
  <span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><ArrowRightLeft :size="18" /></span>
  <div class="min-w-0"><p class="text-sm font-semibold">{{ tab === 'Expenses' ? 'Expense control' : 'Treasury transfers' }}</p><p class="mt-0.5 text-xs leading-5 text-base-content/60">{{ tab === 'Expenses' ? 'Track operating costs against the account that funded them.' : 'Move funds between financial accounts while keeping a complete audit trail.' }}</p></div>
</div>
<template v-if="tab==='Expenses'"><AppPanel :title="editingExpense ? 'Edit expense' : 'Post expense'" :subtitle="editingExpense ? 'Update the expense and its accounting entry.' : 'Record the category, payee and funding account.'"><form @submit.prevent="createExpense" class="min-w-0 space-y-3"><FormGrid :columns="3">
          <SelectField
            v-model="expense.categoryAccountId"
            label="Category"
            :options="
              props.accounts
                .filter((account) => account.type === 'expense' && account.code !== '5000')
                .map((account) => ({
                  label: `${account.code} · ${account.name}`,
                  value: account.id,
                }))
            "
          /><SelectField
            v-model="expense.financialAccountId"
            label="Financial account"
            :options="
              props.financial.filter((account) => account.active).map((account) => ({ label: accountLabel(account), value: account.id }))
            "
          /><FormField label="Description"><AppInput
            class="input w-full min-w-0"
            v-model="expense.description"
            placeholder="Description"
          /></FormField><FormField label="Payee"><AppInput
            class="input w-full min-w-0"
            v-model="expense.payee"
            placeholder="Payee"
          /></FormField><SelectField
            v-model="expense.supplierId"
            label="Supplier"
            :options="[
              { label: 'No supplier', value: '' },
              ...props.suppliers.map((supplier) => ({ label: supplier.name, value: supplier.id })),
            ]"
          /><FormField :label="`Amount (${props.currencyUnit})`"><AppInput
            class="input w-full min-w-0"
            v-model="expense.amount"
            :money="props.currencyUnit"
            required
            inputmode="numeric"
            :placeholder="`Amount (${props.currencyUnit})`"
          /></FormField><div class="flex flex-wrap items-center gap-2"><button v-if="editingExpense" class="btn btn-ghost" type="button" @click="cancelExpenseEdit">Cancel</button><button class="btn btn-primary" type="submit" :disabled="busy"><Plus :size="15" /> {{ editingExpense ? 'Save expense' : 'Post expense' }}</button></div>
        </FormGrid></form></AppPanel>
<SearchFilterBar>
  <template #search><SearchField v-model="historyQuery" label="Search expenses and purchases" placeholder="Search number, supplier, description, or account" /></template>
  <template #filters>
    <SelectField v-model="historyType" label="Record type" :options="['All', 'Expenses', 'Purchases'].map((value) => ({ label: value, value }))" />
    <SelectField v-model="historyStatus" label="Status" :options="['All', 'Posted', 'Reversed', 'Draft', 'Cancelled', 'Archived'].map((value) => ({ label: value, value }))" />
  </template>
  <template #count><span>{{ filteredHistory.length }} of {{ historyEntries.length }} records</span></template>
  <template #actions><button v-if="historyQuery || historyType !== 'All' || historyStatus !== 'All'" class="btn btn-primary btn-sm" type="button" @click="clearHistoryFilters">Clear filters</button></template>
</SearchFilterBar>
<RegisterList title="Expense and purchase history" subtitle="Expenses and supplier purchases share one searchable financial history." :count="filteredHistory.length">
  <div v-if="filteredHistory.length">
    <RegisterRow v-for="entry in filteredHistory" :key="`${entry.kind}-${entry.id}`" :interactive="false" sidecar>
      <template #icon><ArrowRightLeft :size="17" :stroke-width="1.8" aria-hidden="true" /></template>
      <template #identity>
        <div class="min-w-0">
          <div class="flex min-w-0 flex-wrap items-center gap-2">
            <StatusBadge :label="entry.kind" :tone="entry.kind === 'Expense' ? 'blue' : 'amber'" />
            <strong class="truncate text-sm">{{ entry.number }}</strong>
          </div>
          <p class="mt-1 break-words text-sm text-base-content/80">{{ entry.description }}</p>
        </div>
      </template>
      <template #meta>
        <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2">
          <div class="min-w-0"><span class="block text-base-content/50">Funding account</span><span class="block break-words text-base-content/80">{{ financialAccountName(entry.accountId) }}</span></div>
          <div><span class="block text-base-content/50">Date</span><span class="block text-base-content/80">{{ formatDateTime(entry.date) }}</span></div>
        </div>
      </template>
      <template #status><div class="flex flex-col items-end gap-2"><strong class="text-sm tabular-nums text-primary">{{ formatMoney(entry.amount, currencyUnit) }}</strong><StatusBadge :label="entry.status" :tone="historyStatusTone(entry.status)" /></div></template>
      <template #actions>
        <div v-if="entry.expense" class="flex flex-wrap justify-end gap-2">
          <button class="btn btn-outline btn-info btn-sm gap-1" type="button" @click.stop="startExpenseEdit(entry.expense)" :disabled="busy"><Edit3 :size="14" /> Edit</button>
          <button class="btn btn-outline btn-error btn-sm gap-1" type="button" @click.stop="removeExpense(entry.expense)" :disabled="busy"><Trash2 :size="14" /> Remove</button>
        </div>
        <span v-else class="whitespace-nowrap text-xs text-base-content/45">Read only</span>
      </template>
    </RegisterRow>
  </div>
  <EmptyState v-else title="No matching records" description="Try a different search or filter.">
    <template #icon><ReceiptText :size="21" aria-hidden="true" /></template>
    <template #action><button class="btn btn-primary btn-sm" type="button" @click="clearHistoryFilters">Clear filters</button></template>
  </EmptyState>
</RegisterList></template>
<template v-else><AppPanel title="Post transfer" subtitle="Move funds between cash and bank accounts."><form @submit.prevent="createTransfer" class="min-w-0 space-y-3"><FormGrid :columns="3">
          <SelectField
            v-model="transfer.sourceFinancialAccountId"
            label="From account"
            :options="
              props.financial.filter((account) => account.active).map((account) => ({
                label: `From ${accountLabel(account)}`,
                value: account.id,
              }))
            "
          /><SelectField
            v-model="transfer.destinationFinancialAccountId"
            label="To account"
            :options="
              props.financial.filter((account) => account.active).map((account) => ({ label: `To ${accountLabel(account)}`, value: account.id }))
            "
          /><FormField :label="`Amount (${props.currencyUnit})`"><AppInput
            class="input w-full min-w-0"
            v-model="transfer.amount"
            :money="props.currencyUnit"
            required
            inputmode="numeric"
            :placeholder="`Amount (${props.currencyUnit})`"
          /></FormField><FormField label="Reference"><AppInput
            class="input w-full min-w-0"
            v-model="transfer.reference"
            placeholder="Reference"
          /></FormField><button class="btn btn-primary" type="submit" :disabled="busy">
            <ArrowRightLeft :size="15" /> Post transfer
          </button>
        </FormGrid></form></AppPanel><AppPanel title="Transfer history" subtitle="Every transfer remains traceable and can be reversed." flush><DataTable v-if="transfers.length"><thead><tr><th>Transfer</th><th>Reference</th><th>Date</th><th class="text-end">Amount</th><th>Status</th><th>Actions</th></tr></thead><tbody><tr v-for="v in transfers" :key="v.id"><DataTableCell><strong>{{v.transferNumber}}</strong></DataTableCell><DataTableCell>{{v.reference || '—'}}</DataTableCell><DataTableCell>{{formatDateTime(v.transferDate)}}</DataTableCell><DataTableCell numeric>{{formatMoney(v.amountRial,currencyUnit)}}</DataTableCell><DataTableCell><StatusBadge :label="v.status" :tone="v.status==='Posted'?'green':'slate'" /></DataTableCell><DataTableCell><button v-if="v.status==='Posted'" class="btn btn-ghost btn-sm" @click="reverseTransfer(v)" :disabled="busy"><RotateCcw :size="14" /> Reverse</button></DataTableCell></tr></tbody></DataTable><EmptyState v-else compact title="No transfers" description="Transfers between financial accounts will appear here."><template #icon><ArrowRightLeft :size="21" aria-hidden="true" /></template></EmptyState></AppPanel></template>
</div></div></template>
