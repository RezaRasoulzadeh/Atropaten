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
import { Landmark, Plus, RotateCcw, WalletCards } from 'lucide-vue-next';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import AccountingExtrasView from './AccountingExtrasView.vue';
import InvoicesView from '../invoices/InvoicesView.vue';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import {
  accountingApi,
  type AccountRecord,
  type FinancialAccountRecord,
  type JournalEntryRecord,
  type PaymentRecord,
} from '../../api/accounting';
import SelectField from '../../components/ui/SelectField.vue';

const props = defineProps<{
  currencyUnit: CurrencyUnit;
  orders: any[];
  suppliers: any[];
  purchases: any[];
  customers: any[];
}>();
const emit = defineEmits<{ notify: [string]; refreshOrders: [] }>();
const tab = ref('Overview');
const accounts = ref<AccountRecord[]>([]);
const financial = ref<FinancialAccountRecord[]>([]);
const journal = ref<JournalEntryRecord[]>([]);
const payments = ref<PaymentRecord[]>([]);
const form = ref({
  direction: 'incoming',
  method: 'cash',
  financialAccountId: 'FIN-CASH',
  customerId: '',
  supplierId: '',
  targetId: '',
  amount: '',
  notes: '',
});
const money = (v: number) => formatMoney(v, props.currencyUnit);
const totalAssets = computed(() =>
  accounts.value.filter((a) => a.type === 'asset').reduce((n, a) => n + a.balanceRial, 0),
);
const receivable = computed(() => accounts.value.find((a) => a.code === '1100')?.balanceRial ?? 0);
const payable = computed(() => accounts.value.find((a) => a.code === '2000')?.balanceRial ?? 0);
async function load() { return runLoad(async () => {
  try {
    [accounts.value, financial.value, journal.value, payments.value] = await Promise.all([
      accountingApi.accounts(),
      accountingApi.financialAccounts(),
      accountingApi.journal(),
      accountingApi.payments(),
    ]);
    if (financial.value[0] && !financial.value.some((x) => x.id === form.value.financialAccountId))
      form.value.financialAccountId = financial.value[0].id;
  } catch (e) {
    reportError(e);
  }
}); }
onMounted(load);
async function create() {
return runAction(async () => {
  try {
    const amount = parseMoneyInput(form.value.amount, props.currencyUnit);
    if (amount === null || amount <= 0) throw new Error('Enter a positive Rial amount');
    await accountingApi.createPayment({
      direction: form.value.direction,
      method: form.value.method,
      financialAccountId: form.value.financialAccountId,
      customerId: form.value.customerId || undefined,
      supplierId: form.value.supplierId || undefined,
      amountRial: amount,
      notes: form.value.notes,
      allocations: form.value.targetId
        ? [
            {
              targetType: form.value.direction === 'incoming' ? 'invoice' : 'purchase',
              targetId: form.value.targetId,
              amountRial: amount,
            },
          ]
        : [],
    });
    form.value.amount = '';
    form.value.targetId = '';
    await load();
    emit('notify', 'Payment posted.');
  } catch (e) {
reportError(e);
  }

});
}
async function reverse(p: PaymentRecord) {
return runAction(async () => {
  try {
    await accountingApi.reversePayment(p.id);
    await load();
    emit('notify', 'Payment reversed.');
  } catch (e) {
reportError(e);
  }

});
}
function targetOptions() {
  return form.value.direction === 'incoming'
    ? props.orders.filter((o) => o.invoiceId)
    : props.purchases;
}
function targetLabel(x: any) {
  return form.value.direction === 'incoming'
    ? `${x.invoiceId || x.orderNumber} · ${x.customerName || 'Walk-in'}`
    : `${x.purchaseNumber} · ${x.supplierName}`;
}
const cashBalance = computed(() => accounts.value.find((a) => a.code === '1000')?.balanceRial ?? 0);
const bankBalance = computed(() => accounts.value.find((a) => a.code === '1010')?.balanceRial ?? 0);
const revenue = computed(() => accounts.value.find((a) => a.code === '4000')?.balanceRial ?? 0);
const expensesTotal = computed(() =>
  accounts.value
    .filter((a) => a.type === 'expense' && a.code !== '5000')
    .reduce((n, a) => n + a.balanceRial, 0),
);
</script>
<template>
<div class="space-y-4">
<WorkspaceStickyStack><WorkspaceHeader eyebrow="Finance / ledger" title="Accounting" description="Balances and transaction history from the posted ledger."><button class="btn btn-primary" @click="tab='Payments'"><Plus :size="15" />New payment</button></WorkspaceHeader><WorkspaceTabs :tabs="['Overview','Accounts','Journal','Payments','Invoices','Expenses','Transfers / Treasury']" :active-tab="tab" @change="tab=$event" /></WorkspaceStickyStack><LoadingState v-if="pageLoading" label="Loading records…" /><div v-show="!pageLoading" class="space-y-4">
<template v-if="tab==='Overview'">
<div class="grid gap-3 sm:grid-cols-3"><AppPanel v-for="metric in [{label:'Assets',value:totalAssets},{label:'Receivable',value:receivable},{label:'Payable',value:payable}]" :key="metric.label"><p class="text-xs text-base-content/60">{{metric.label}}</p><strong class="block text-xl font-semibold tabular-nums wrap-anywhere">{{money(metric.value)}}</strong></AppPanel></div>
<AppPanel title="Recent transactions" flush><DataTable><thead><tr><th>Entry</th><th>Description</th><th>Posted</th></tr></thead><tbody><tr v-for="entry in journal.slice(0,8)" :key="entry.id"><DataTableCell><strong>{{entry.entryNumber}}</strong></DataTableCell><DataTableCell>{{entry.description}}</DataTableCell><DataTableCell class="whitespace-nowrap">{{formatDateTime(entry.postedAt)}}</DataTableCell></tr></tbody></DataTable><EmptyState v-if="!journal.length" title="No journal entries" /></AppPanel>
</template>
<AppPanel v-else-if="tab==='Accounts'" title="Chart of accounts" flush><DataTable><thead><tr><th>Code</th><th>Account</th><th>Type</th><th class="text-end">Balance</th></tr></thead><tbody><tr v-for="a in accounts" :key="a.id"><DataTableCell>{{a.code}}</DataTableCell><DataTableCell><strong>{{a.name}}</strong></DataTableCell><DataTableCell>{{a.type}}</DataTableCell><DataTableCell numeric>{{money(a.balanceRial)}}</DataTableCell></tr></tbody></DataTable></AppPanel>
<AppPanel v-else-if="tab==='Journal'" title="Journal entries" subtitle="Separate debit and credit columns for every posted line." flush><DataTable><thead><tr><th>Entry / date</th><th>Account / memo</th><th class="text-end">Debit</th><th class="text-end">Credit</th></tr></thead><tbody><template v-for="entry in journal" :key="entry.id"><tr v-for="(line,index) in entry.lines" :key="line.id"><DataTableCell><template v-if="index===0"><strong class="block">{{entry.entryNumber}}</strong><span class="block text-xs text-base-content/60">{{formatDateTime(entry.postedAt)}}</span></template></DataTableCell><DataTableCell><strong class="block">{{accounts.find(a=>a.id===line.accountId)?.name || line.accountId}}</strong><span class="block text-xs text-base-content/60">{{line.memo || entry.description}}</span></DataTableCell><DataTableCell numeric>{{line.debitRial ? money(line.debitRial) : '—'}}</DataTableCell><DataTableCell numeric>{{line.creditRial ? money(line.creditRial) : '—'}}</DataTableCell></tr></template></tbody></DataTable><EmptyState v-if="!journal.length" title="No journal entries" /></AppPanel>
<template v-else-if="tab==='Payments'"><AppPanel title="Post payment" subtitle="Allocate a payment or leave an unallocated customer balance."><form @submit.prevent="create" class="min-w-0 space-y-3"><FormGrid>
          <SelectField
            v-model="form.direction"
            label="Direction"
            :options="[
              { label: 'Incoming · customer', value: 'incoming' },
              { label: 'Outgoing · supplier', value: 'outgoing' },
            ]"
          /><SelectField
            v-model="form.method"
            label="Method"
            :options="[
              { label: 'Cash', value: 'cash' },
              { label: 'Bank transfer', value: 'bank_transfer' },
              { label: 'Card', value: 'card' },
              { label: 'Check', value: 'check' },
              { label: 'Other', value: 'other' },
            ]"
          /><SelectField
            v-model="form.financialAccountId"
            label="Financial account"
            :options="financial.map((account) => ({ label: account.name, value: account.id }))"
          /><SelectField
            v-if="form.direction === 'incoming'"
            v-model="form.customerId"
            label="Customer"
            :options="[
              { label: 'Select customer', value: '' },
              ...props.customers.map((customer) => ({ label: customer.name, value: customer.id })),
            ]"
          /><SelectField
            v-else
            v-model="form.supplierId"
            label="Supplier"
            :options="[
              { label: 'Select supplier', value: '' },
              ...suppliers.map((supplier) => ({ label: supplier.name, value: supplier.id })),
            ]"
          /><SelectField
            v-model="form.targetId"
            label="Allocation"
            :options="[
              { label: 'Unallocated balance', value: '' },
              ...targetOptions().map((target) => ({
                label: targetLabel(target),
                value: target.id,
              })),
            ]"
          /><FormField :label="`Amount (${props.currencyUnit})`"><AppInput
            class="input w-full min-w-0"
            v-model="form.amount"
            :placeholder="`Amount (${props.currencyUnit})`"
            inputmode="numeric"
            @blur="
              form.amount = formatMoneyInput(
                parseMoneyInput(form.amount, props.currencyUnit) || 0,
                props.currencyUnit,
              )
            "
          /></FormField><FormField label="Notes"><AppTextarea
            v-model="form.notes"
            rows="2"
            placeholder="Notes"
          />
          ></FormField><button class="btn btn-primary" type="submit" :disabled="busy">Post payment</button>
        </FormGrid></form></AppPanel><AppPanel title="Payment history" flush><DataTable><thead><tr><th>Payment</th><th>Direction</th><th>Posted</th><th class="text-end">Amount</th><th>Status</th><th>Actions</th></tr></thead><tbody><tr v-for="p in payments" :key="p.id"><DataTableCell><strong>{{p.paymentNumber}}</strong></DataTableCell><DataTableCell>{{p.direction}}</DataTableCell><DataTableCell>{{formatDateTime(p.postedAt)}}</DataTableCell><DataTableCell numeric>{{money(p.amountRial)}}</DataTableCell><DataTableCell><StatusBadge :label="p.status" :tone="p.status==='posted'?'green':'slate'" /></DataTableCell><DataTableCell><button v-if="p.status==='posted'" class="btn btn-ghost btn-sm" @click="reverse(p)" :disabled="busy"><RotateCcw :size="14" />Reverse</button></DataTableCell></tr></tbody></DataTable><EmptyState v-if="!payments.length" title="No payments" /></AppPanel></template>
<InvoicesView v-else-if="tab==='Invoices'" :currency-unit="currencyUnit" :orders="orders" @notify="emit('notify',$event)" @refresh-orders="emit('refreshOrders')" />
<AccountingExtrasView v-else :tab="tab" :currency-unit="currencyUnit" :accounts="accounts" :financial="financial" :suppliers="suppliers" @notify="emit('notify',$event)" />
</div></div>
</template>
