<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import FormField from '../../components/ui/FormField.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction,pageLoading,runLoad}=useWorkspaceActions()

import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import WorkspaceTabs from '../../components/layout/WorkspaceTabs.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref } from 'vue';
import { Archive, BookOpen, Landmark, Pencil, Plus, Save, Trash2, Users, WalletCards } from 'lucide-vue-next';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import AccountingExtrasView from './AccountingExtrasView.vue';
import ProfitAllocationView from './ProfitAllocationView.vue';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney } from '../../utils/currency';
import {
  accountingApi,
  type AccountRecord,
  type FinancialAccountRecord,
  type JournalEntryRecord,
} from '../../api/accounting';
import SelectField from '../../components/ui/SelectField.vue';
import { ownersApi, type OwnerRecord } from '../../api/owners';
import { confirmAction } from '../../ui/feedback';
import { formatDateTime } from '../../utils/date';

const props = defineProps<{
  currencyUnit: CurrencyUnit;
  suppliers: any[];
}>();
const emit = defineEmits<{ notify: [string] }>();
const tab = ref('Overview');
const accounts = ref<AccountRecord[]>([]);
const financial = ref<FinancialAccountRecord[]>([]);
const journal = ref<JournalEntryRecord[]>([]);
const owners = ref<OwnerRecord[]>([]);
const accountFormOpen = ref(false);
const accountForm = ref({
  id: '',
  name: '',
  type: 'bank' as 'cash' | 'bank',
  bankName: '',
  accountNumber: '',
  cardNumber: '',
  details: '',
  ownerIds: [] as string[],
});
const editingAccount = computed(() => !!accountForm.value.id);
const money = (v: number) => formatMoney(v, props.currencyUnit);
const totalAssets = computed(() =>
  accounts.value.filter((a) => a.type === 'asset').reduce((n, a) => n + a.balanceRial, 0),
);
const receivable = computed(() => accounts.value.find((a) => a.code === '1100')?.balanceRial ?? 0);
const payable = computed(() => accounts.value.find((a) => a.code === '2000')?.balanceRial ?? 0);
async function load() { return runLoad(async () => {
  try {
    [accounts.value, financial.value, owners.value, journal.value] = await Promise.all([
      accountingApi.accounts(),
      accountingApi.financialAccounts(),
      ownersApi.list(false),
      accountingApi.journal(),
    ]);
  } catch (e) {
    reportError(e);
  }
}); }
onMounted(load);
function startAccount(account?: FinancialAccountRecord) {
  accountForm.value = account
    ? { id: account.id, name: account.name, type: account.type as 'cash' | 'bank', bankName: account.bankName || '', accountNumber: account.accountNumber || '', cardNumber: account.cardNumber || '', details: account.details, ownerIds: [...(account.ownerIds || [])] }
    : { id: '', name: '', type: 'bank', bankName: '', accountNumber: '', cardNumber: '', details: '', ownerIds: [] };
  accountFormOpen.value = true;
  tab.value = 'Accounts';
}
function cancelAccountEdit() {
  accountFormOpen.value = false;
  accountForm.value = { id: '', name: '', type: 'bank', bankName: '', accountNumber: '', cardNumber: '', details: '', ownerIds: [] };
}
function ownerNames(ids: string[]) {
  return (ids || []).map((id) => owners.value.find((owner) => owner.id === id)?.name || id);
}
async function saveFinancialAccount() {
  return runAction(async () => {
    try {
      if (!accountForm.value.name.trim()) throw new Error('Enter an account name');
      if (accountForm.value.type === 'bank' && !accountForm.value.bankName.trim()) throw new Error('Enter a bank name');
      if (accountForm.value.type === 'bank' && !accountForm.value.accountNumber.trim() && !accountForm.value.cardNumber.trim()) throw new Error('Enter an account number or card number');
      const editing = editingAccount.value;
      const value = editing
        ? await accountingApi.updateFinancialAccount(accountForm.value as typeof accountForm.value & { id: string })
        : await accountingApi.createFinancialAccount(accountForm.value);
      financial.value = editing
        ? financial.value.map((account) => (account.id === value.id ? value : account))
        : [value, ...financial.value];
      cancelAccountEdit();
      emit('notify', editing ? 'Treasury account updated.' : 'Treasury account added.');
    } catch (e) {
      reportError(e);
    }
  });
}
async function archiveFinancialAccount(account: FinancialAccountRecord) {
  return runAction(async () => {
    if (!(await confirmAction({
      title: 'Archive treasury account',
      message: `Archive ${account.name}? It will remain in history but cannot be used for new transactions.`,
      confirmLabel: 'Archive account',
    }))) return;
    try {
      await accountingApi.archiveFinancialAccount(account.id);
      financial.value = financial.value.map((value) => value.id === account.id ? { ...value, active: false } : value);
      emit('notify', 'Treasury account archived.');
    } catch (e) {
      reportError(e);
    }
  });
}
async function removeFinancialAccount(account: FinancialAccountRecord) {
  return runAction(async () => {
    if (!(await confirmAction({
      title: 'Delete treasury account',
      message: `Delete ${account.name}? This is only possible when it has no financial history.`,
      confirmLabel: 'Delete account',
      danger: true,
    }))) return;
    try {
      await accountingApi.deleteFinancialAccount(account.id);
      financial.value = financial.value.filter((value) => value.id !== account.id);
      emit('notify', 'Treasury account deleted.');
    } catch (e) {
      reportError(e);
    }
  });
}
const cashBalance = computed(() => accounts.value.find((a) => a.code === '1000')?.balanceRial ?? 0);
const bankBalance = computed(() => accounts.value.find((a) => a.code === '1010')?.balanceRial ?? 0);
const revenue = computed(() => accounts.value.find((a) => a.code === '4000')?.balanceRial ?? 0);
const expensesTotal = computed(() =>
  accounts.value
    .filter((a) => a.type === 'expense' && a.code !== '5000')
    .reduce((n, a) => n + a.balanceRial, 0),
);
const recentTransactions = computed(() => journal.value.slice(0, 8));
function journalAmount(entry: JournalEntryRecord) {
  return entry.lines.reduce((total, line) => total + line.debitRial, 0);
}
function transactionSource(entry: JournalEntryRecord) {
  return entry.sourceType ? entry.sourceType.replaceAll('_', ' ') : 'Transaction';
}
</script>
<template>
<div class="min-w-0 space-y-4">
  <WorkspaceStickyStack>
    <WorkspaceHeader
      :show-breadcrumb="true"
      eyebrow="Finance / ledger"
      title="Accounting"
      description="Balances and treasury accounts for the business."
    >
    </WorkspaceHeader>
    <WorkspaceTabs
      :tabs="['Overview', 'Accounts', 'Expenses', 'Transfers / Treasury', 'Profit Allocation']"
      :active-tab="tab"
      @change="tab = $event"
    />
  </WorkspaceStickyStack>

  <LoadingState v-if="pageLoading" label="Loading records…" />
  <div v-show="!pageLoading" class="min-w-0 space-y-4">
    <template v-if="tab === 'Overview'">
      <section class="grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-label="Balance summary">
        <div class="flex min-h-32 min-w-0 flex-col rounded-box border border-base-300 bg-base-100 p-4">
          <div class="flex items-start justify-between gap-3">
            <p class="min-w-0 text-xs font-medium text-base-content/60">Total assets</p>
            <span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Landmark :size="18" /></span>
          </div>
          <strong class="mt-2 block whitespace-nowrap text-lg font-semibold leading-7 tabular-nums xl:text-xl">{{ money(totalAssets) }}</strong>
          <p class="mt-auto pt-2 text-xs text-base-content/55">All asset accounts</p>
        </div>
        <div class="flex min-h-32 min-w-0 flex-col rounded-box border border-base-300 bg-base-100 p-4">
          <div class="flex items-start justify-between gap-3">
            <p class="min-w-0 text-xs font-medium text-base-content/60">Cash on hand</p>
            <span class="grid size-9 shrink-0 place-items-center rounded-box bg-success/15 text-success"><WalletCards :size="18" /></span>
          </div>
          <strong class="mt-2 block whitespace-nowrap text-lg font-semibold leading-7 tabular-nums xl:text-xl">{{ money(cashBalance) }}</strong>
          <p class="mt-auto pt-2 text-xs text-base-content/55">Cash financial account</p>
        </div>
        <div class="flex min-h-32 min-w-0 flex-col rounded-box border border-base-300 bg-base-100 p-4">
          <div class="flex items-start justify-between gap-3">
            <p class="min-w-0 text-xs font-medium text-base-content/60">Receivable</p>
            <span class="grid size-9 shrink-0 place-items-center rounded-box bg-info/15 text-info"><BookOpen :size="18" /></span>
          </div>
          <strong class="mt-2 block whitespace-nowrap text-lg font-semibold leading-7 tabular-nums xl:text-xl">{{ money(receivable) }}</strong>
          <p class="mt-auto pt-2 text-xs text-base-content/55">Outstanding customer balance</p>
        </div>
        <div class="flex min-h-32 min-w-0 flex-col rounded-box border border-base-300 bg-base-100 p-4">
          <div class="flex items-start justify-between gap-3">
            <p class="min-w-0 text-xs font-medium text-base-content/60">Payable</p>
            <span class="grid size-9 shrink-0 place-items-center rounded-box bg-warning/15 text-warning"><Landmark :size="18" /></span>
          </div>
          <strong class="mt-2 block whitespace-nowrap text-lg font-semibold leading-7 tabular-nums xl:text-xl">{{ money(payable) }}</strong>
          <p class="mt-auto pt-2 text-xs text-base-content/55">Outstanding supplier balance</p>
        </div>
      </section>

      <section class="grid min-w-0 gap-4 xl:grid-cols-2">
        <AppPanel title="Financial snapshot" subtitle="Current totals from the accounting balances.">
          <dl class="divide-y divide-base-300 rounded-box border border-base-300 bg-base-200/30">
            <div class="flex items-center justify-between gap-3 px-3 py-2.5 text-sm"><dt class="text-base-content/65">Bank balance</dt><dd class="font-semibold tabular-nums">{{ money(bankBalance) }}</dd></div>
            <div class="flex items-center justify-between gap-3 px-3 py-2.5 text-sm"><dt class="text-base-content/65">Revenue</dt><dd class="font-semibold tabular-nums text-success">{{ money(revenue) }}</dd></div>
            <div class="flex items-center justify-between gap-3 px-3 py-2.5 text-sm"><dt class="text-base-content/65">Expenses</dt><dd class="font-semibold tabular-nums text-error">{{ money(expensesTotal) }}</dd></div>
          </dl>
          <p class="text-xs leading-5 text-base-content/55">Use Accounts to manage treasury accounts, or use the dedicated operational views for orders and invoices.</p>
        </AppPanel>
        <AppPanel title="Recent transactions" subtitle="The latest posted activity across the ledger." flush>
          <div v-if="recentTransactions.length" class="divide-y divide-base-300">
            <div v-for="entry in recentTransactions" :key="entry.id" class="flex min-w-0 items-center justify-between gap-3 px-4 py-3">
              <div class="min-w-0">
                <strong class="block truncate text-sm">{{ entry.description }}</strong>
                <span class="mt-1 block truncate text-xs text-base-content/55">{{ entry.entryNumber }} · {{ transactionSource(entry) }} · {{ formatDateTime(entry.postedAt) }}</span>
              </div>
              <span class="shrink-0 text-sm font-semibold tabular-nums">{{ money(journalAmount(entry)) }}</span>
            </div>
          </div>
          <EmptyState v-else title="No recent transactions" description="Posted accounting activity will appear here."><template #icon><BookOpen :size="21" aria-hidden="true" /></template></EmptyState>
        </AppPanel>
      </section>
    </template>

    <section v-else-if="tab === 'Accounts'" class="min-w-0 space-y-4">
      <AppPanel v-if="accountFormOpen" :title="editingAccount ? 'Edit treasury account' : 'Add treasury account'" :subtitle="editingAccount ? 'Update the account name, type, details, or attached owners.' : 'Create a cash or bank account for treasury movements.'">
        <form class="min-w-0 space-y-4" @submit.prevent="saveFinancialAccount">
          <section class="rounded-box border border-base-300 bg-base-200/20 p-3">
            <div class="mb-3"><h3 class="text-sm font-semibold">Account identity</h3><p class="mt-0.5 text-xs text-base-content/55">Choose the account type and give it a clear name.</p></div>
            <FormGrid :columns="2">
              <FormField label="Account name"><AppInput v-model="accountForm.name" class="input w-full min-w-0" required placeholder="e.g. Main business bank" /></FormField>
              <SelectField v-model="accountForm.type" label="Account type" :options="[{ label: 'Bank account', value: 'bank' }, { label: 'Cash account', value: 'cash' }]" />
            </FormGrid>
          </section>
          <section v-if="accountForm.type === 'bank'" class="rounded-box border border-base-300 bg-base-200/20 p-3">
            <div class="mb-3"><h3 class="text-sm font-semibold">Bank details</h3><p class="mt-0.5 text-xs text-base-content/55">Bank name is required; provide an account number or card number.</p></div>
            <FormGrid :columns="3">
              <FormField label="Bank name"><AppInput v-model="accountForm.bankName" class="input w-full min-w-0" required placeholder="e.g. Tejarat Bank" /></FormField>
              <FormField label="Account number"><AppInput v-model="accountForm.accountNumber" class="input w-full min-w-0" placeholder="Account number" /></FormField>
              <FormField label="Card number"><AppInput v-model="accountForm.cardNumber" class="input w-full min-w-0" inputmode="numeric" placeholder="Card number" /></FormField>
            </FormGrid>
          </section>
          <section class="rounded-box border border-base-300 bg-base-200/20 p-3">
            <div class="mb-3"><h3 class="text-sm font-semibold">Additional information</h3><p class="mt-0.5 text-xs text-base-content/55">Keep optional notes and ownership in one place.</p></div>
            <div class="space-y-4">
              <FormField label="Details"><AppTextarea v-model="accountForm.details" rows="3" placeholder="Branch, purpose, or other notes" /></FormField>
              <FormField label="Attached owners">
                <div v-if="owners.length" class="grid gap-2 rounded-box border border-base-300 bg-base-200/30 p-3 sm:grid-cols-2">
                  <label v-for="owner in owners" :key="owner.id" class="flex min-w-0 items-center gap-2 text-sm"><input v-model="accountForm.ownerIds" class="checkbox checkbox-sm" type="checkbox" :value="owner.id" /><span class="truncate">{{ owner.name }}</span></label>
                </div>
                <EmptyState v-else compact title="No owners available" description="Add owners from the Owners workspace to attach them here."><template #icon><Users :size="21" aria-hidden="true" /></template></EmptyState>
              </FormField>
            </div>
          </section>
          <div class="flex flex-wrap justify-end gap-2">
            <button class="btn btn-ghost" type="button" @click="cancelAccountEdit">Cancel</button>
            <button class="btn btn-primary" type="submit" :disabled="busy"><Save :size="15" /> {{ editingAccount ? 'Save changes' : 'Add account' }}</button>
          </div>
        </form>
      </AppPanel>
      <AppPanel title="Treasury accounts" subtitle="Cash and bank accounts available to payments, expenses, transfers, checks, loans, and owner movements." flush>
        <template #action><button class="btn btn-primary btn-sm" type="button" @click="startAccount()"><Plus :size="15" /> New account</button></template>
        <DataTable v-if="financial.length">
          <thead><tr><th>Account</th><th>Type</th><th>Owners</th><th class="text-end">Balance</th><th>Status</th><th class="text-end">Actions</th></tr></thead>
          <tbody>
            <tr v-for="account in financial" :key="account.id">
              <DataTableCell><strong class="block">{{ account.name }}</strong><span class="block max-w-72 truncate text-xs text-base-content/60">{{ account.type === 'bank' ? [account.bankName, account.accountNumber && `Account ${account.accountNumber}`, account.cardNumber && `Card ${account.cardNumber}`].filter(Boolean).join(' · ') || 'Bank details not set' : account.details || 'Cash account' }}</span></DataTableCell>
              <DataTableCell><StatusBadge :label="account.type === 'cash' ? 'Cash' : 'Bank'" :tone="account.type === 'cash' ? 'green' : 'blue'" /></DataTableCell>
              <DataTableCell><span v-if="ownerNames(account.ownerIds).length" class="text-xs">{{ ownerNames(account.ownerIds).join(' · ') }}</span><span v-else class="text-xs text-base-content/50">Business account</span></DataTableCell>
              <DataTableCell numeric>{{ money(account.balanceRial) }}</DataTableCell>
              <DataTableCell><StatusBadge :label="account.active ? 'Active' : 'Archived'" :tone="account.active ? 'green' : 'slate'" /></DataTableCell>
              <DataTableCell><div class="flex justify-end gap-2"><button class="btn btn-outline btn-info btn-sm" type="button" @click="startAccount(account)"><Pencil :size="14" /> Edit</button><button v-if="account.active" class="btn btn-outline btn-warning btn-sm" type="button" @click="archiveFinancialAccount(account)"><Archive :size="14" /> Archive</button><button v-if="account.id !== 'FIN-CASH'" class="btn btn-outline btn-error btn-sm" type="button" @click="removeFinancialAccount(account)"><Trash2 :size="14" /> Delete</button></div></DataTableCell>
            </tr>
          </tbody>
        </DataTable>
        <EmptyState v-else title="No treasury accounts" description="Add a cash or bank account to start tracking treasury activity.">
          <template #icon><Landmark :size="22" aria-hidden="true" /></template>
          <template #action><button class="btn btn-primary btn-sm gap-2" type="button" @click="startAccount()"><Plus :size="15" aria-hidden="true" /> Create account</button></template>
        </EmptyState>
      </AppPanel>
    </section>

    <ProfitAllocationView
      v-else-if="tab === 'Profit Allocation'"
      :currency-unit="props.currencyUnit"
      @notify="emit('notify', $event)"
    />
    <AccountingExtrasView v-else :tab="tab" :currency-unit="currencyUnit" :accounts="accounts" :financial="financial" :suppliers="suppliers" @notify="emit('notify', $event)" />
  </div>
</div>
</template>
