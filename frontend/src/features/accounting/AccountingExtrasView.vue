<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import FormField from '../../components/ui/FormField.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction,pageLoading,runLoad}=useWorkspaceActions()

import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import InlineAlert from '../../components/ui/InlineAlert.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { onMounted, ref } from 'vue';
import { ArrowRightLeft, Plus, RotateCcw } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import {
  accountingApi,
  type AccountRecord,
  type ExpenseRecord,
  type FinancialAccountRecord,
  type TransferRecord,
} from '../../api/accounting';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, parseMoneyInput } from '../../utils/currency';
import { currentCanonicalDate, formatDateTime } from '../../utils/date';
import { normalizeError } from '../../ui/feedback';
import SelectField from '../../components/ui/SelectField.vue';

const props = defineProps<{
  tab: string;
  currencyUnit: CurrencyUnit;
  accounts: AccountRecord[];
  financial: FinancialAccountRecord[];
  suppliers: any[];
}>();
const emit = defineEmits<{ notify: [string] }>();
const expenses = ref<ExpenseRecord[]>([]);
const transfers = ref<TransferRecord[]>([]);
const error = ref('');
const expense = ref({
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
    [expenses.value, transfers.value] = await Promise.all([
      accountingApi.expenses(),
      accountingApi.transfers(),
    ]);
  } catch (e) {
    error.value = normalizeError(e).message;
  }
}); }
onMounted(load);
async function createExpense() {
return runAction(async () => {
  try {
    const amount = parseMoneyInput(expense.value.amount, props.currencyUnit);
    if (!amount || amount <= 0) throw new Error('Enter a positive amount');
    const v = await accountingApi.createExpense({ ...expense.value, amountRial: amount });
    expenses.value = [v, ...expenses.value];
    expense.value = { ...expense.value, description: '', payee: '', amount: '' };
    emit('notify', 'Expense posted.');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
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
    error.value = normalizeError(e).message;
  }

});
}
async function reverseExpense(v: ExpenseRecord) {
return runAction(async () => {
  try {
    const next = await accountingApi.reverseExpense(v.id);
    expenses.value = expenses.value.map((x) => (x.id === next.id ? next : x));
    emit('notify', 'Expense reversed.');
  } catch (e) {
reportError(e);
    error.value = normalizeError(e).message;
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
    error.value = normalizeError(e).message;
  }

});
}
</script>
<template><div class="space-y-4"><LoadingState v-if="pageLoading" label="Loading records…" /><div v-show="!pageLoading" class="space-y-4">
<InlineAlert v-if="error" :message="error" />
<template v-if="tab==='Expenses'"><AppPanel title="Post expense" subtitle="Record the category, payee and funding account."><form @submit.prevent="createExpense" class="min-w-0 space-y-3"><FormGrid>
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
              props.financial.map((account) => ({ label: account.name, value: account.id }))
            "
          /><FormField label="Description"><AppInput
            class="input w-full min-w-0"
            v-model="expense.description"
            required
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
            required
            inputmode="numeric"
            :placeholder="`Amount (${props.currencyUnit})`"
          /></FormField><button class="btn btn-primary" type="submit" :disabled="busy"><Plus :size="15" /> Post expense</button>
        </FormGrid></form></AppPanel>
<AppPanel title="Expense history" flush><DataTable><thead><tr><th>Expense</th><th>Description</th><th>Date</th><th class="text-end">Amount</th><th>Status</th><th>Actions</th></tr></thead><tbody><tr v-for="v in expenses" :key="v.id"><DataTableCell><strong>{{v.expenseNumber}}</strong></DataTableCell><DataTableCell>{{v.description}}</DataTableCell><DataTableCell>{{formatDateTime(v.expenseDate)}}</DataTableCell><DataTableCell numeric>{{formatMoney(v.amountRial,currencyUnit)}}</DataTableCell><DataTableCell><StatusBadge :label="v.status" :tone="v.status==='Posted'?'green':'slate'" /></DataTableCell><DataTableCell><button v-if="v.status==='Posted'" class="btn btn-ghost btn-sm" @click="reverseExpense(v)" :disabled="busy"><RotateCcw :size="14" />Reverse</button></DataTableCell></tr></tbody></DataTable><EmptyState v-if="!expenses.length" title="No expenses" /></AppPanel></template>
<template v-else><AppPanel title="Post transfer" subtitle="Move funds between cash and bank accounts."><form @submit.prevent="createTransfer" class="min-w-0 space-y-3"><FormGrid>
          <SelectField
            v-model="transfer.sourceFinancialAccountId"
            label="From account"
            :options="
              props.financial.map((account) => ({
                label: `From ${account.name}`,
                value: account.id,
              }))
            "
          /><SelectField
            v-model="transfer.destinationFinancialAccountId"
            label="To account"
            :options="
              props.financial.map((account) => ({ label: `To ${account.name}`, value: account.id }))
            "
          /><FormField :label="`Amount (${props.currencyUnit})`"><AppInput
            class="input w-full min-w-0"
            v-model="transfer.amount"
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
        </FormGrid></form></AppPanel><AppPanel title="Transfer history" flush><DataTable><thead><tr><th>Transfer</th><th>Reference</th><th>Date</th><th class="text-end">Amount</th><th>Status</th><th>Actions</th></tr></thead><tbody><tr v-for="v in transfers" :key="v.id"><DataTableCell><strong>{{v.transferNumber}}</strong></DataTableCell><DataTableCell>{{v.reference || '—'}}</DataTableCell><DataTableCell>{{formatDateTime(v.transferDate)}}</DataTableCell><DataTableCell numeric>{{formatMoney(v.amountRial,currencyUnit)}}</DataTableCell><DataTableCell><StatusBadge :label="v.status" :tone="v.status==='Posted'?'green':'slate'" /></DataTableCell><DataTableCell><button v-if="v.status==='Posted'" class="btn btn-ghost btn-sm" @click="reverseTransfer(v)" :disabled="busy"><RotateCcw :size="14" />Reverse</button></DataTableCell></tr></tbody></DataTable><EmptyState v-if="!transfers.length" title="No transfers" /></AppPanel></template>
</div></div></template>
