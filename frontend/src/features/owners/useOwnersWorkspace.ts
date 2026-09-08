import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
import { computed, onMounted, ref } from 'vue';
import {
  ownersApi,
  type FiscalPeriodRecord,
  type OwnerRecord,
  type OwnerTransactionRecord,
} from '../../api/owners';
import { accountingApi, type FinancialAccountRecord } from '../../api/accounting';
import { formatMoney, formatMoneyInput, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import { confirmAction, useToast } from '../../ui/feedback';
export type Tab =
  | 'Overview'
  | 'Owners'
  | 'Transactions'
  | 'Capital & Drawings'
  | 'Loans / Current Accounts'
  | 'Profit Allocation';
export function useOwnersWorkspace(props:{ currencyUnit: 'Rial' | 'Toman' },emit:(event:'notify',message:string)=>void){
const {busy,runAction,pageLoading,runLoad}=useWorkspaceActions()
const toast = useToast();
const tabs: Tab[] = [
  'Overview',
  'Owners',
  'Transactions',
  'Capital & Drawings',
  'Loans / Current Accounts',
  'Profit Allocation',
];
const tab = ref<Tab>('Overview');
const owners = ref<OwnerRecord[]>([]);
const transactions = ref<OwnerTransactionRecord[]>([]);
const periods = ref<FiscalPeriodRecord[]>([]);
const accounts = ref<FinancialAccountRecord[]>([]);
const selectedOwner = ref<string>('');
const selectedPeriod = ref<string>('');
const editing = ref(false);
const saving = ref(false);
const ownerForm = ref({
  name: '',
  phone: '',
  email: '',
  notes: '',
  ownershipBps: 0,
  profitSharingBps: 0,
});
const shareForm = ref({ ownershipBps: 0, profitSharingBps: 0 });
const txForm = ref({
  type: 'capital_contribution',
  amountText: '',
  financialAccountId: '',
  categoryAccountId: 'ACC-EXP-OTHER',
  description: '',
  occurredAt: null as string | null,
  notes: '',
});
const periodForm = ref({
  name: '',
  startDate: null as string | null,
  endDate: null as string | null,
  notes: '',
});
const currentOwner = computed(() => owners.value.find((x) => x.id === selectedOwner.value) ?? null);
const selectedPeriodRow = computed(
  () => periods.value.find((x) => x.id === selectedPeriod.value) ?? null,
);
const filteredTransactions = computed(() =>
  transactions.value.filter((x) => !selectedOwner.value || x.ownerId === selectedOwner.value),
);
const ownerName = (id: string) => owners.value.find((x) => x.id === id)?.name ?? id;
const money = (v: number) => formatMoney(v, props.currencyUnit);
async function load() { return runLoad(async () => {
  try {
    [owners.value, periods.value, accounts.value] = await Promise.all([
      ownersApi.list(false),
      ownersApi.periods(),
      accountingApi.financialAccounts(),
    ]);
    if (!selectedOwner.value) selectedOwner.value = owners.value[0]?.id ?? '';
    if (!selectedPeriod.value) selectedPeriod.value = periods.value[0]?.id ?? '';
    syncShares();
    await loadTransactions();
  } catch (e) {
    reportError(e);
  }
}); }
function syncShares() {
  if (currentOwner.value)
    shareForm.value = {
      ownershipBps: currentOwner.value.ownershipBps,
      profitSharingBps: currentOwner.value.profitSharingBps,
    };
}
async function loadTransactions() {
  try {
    transactions.value = await ownersApi.transactions(selectedOwner.value);
  } catch (e) {
    reportError(e);
  }
}
function startOwner() {
  ownerForm.value = {
    name: '',
    phone: '',
    email: '',
    notes: '',
    ownershipBps: 0,
    profitSharingBps: 0,
  };
  editing.value = true;
  tab.value = 'Owners';
}
async function saveOwner() {
return runAction(async () => {
  saving.value = true;
  try {
    const v = await ownersApi.create(ownerForm.value);
    owners.value = [v, ...owners.value];
    selectedOwner.value = v.id;
    editing.value = false;
    emit('notify', 'Owner saved.');
  } catch (e) {
    reportError(e);
  } finally {
    saving.value = false;
  }

});
}
async function saveShares() {
return runAction(async () => {
  if (!currentOwner.value) return;
  try {
    const v = await ownersApi.updateShares(currentOwner.value.id, shareForm.value);
    owners.value = owners.value.map((x) => (x.id === v.id ? v : x));
    emit('notify', 'Owner shares updated.');
  } catch (e) {
    reportError(e);
  }

});
}
async function setActive() {
return runAction(async () => {
  if (!currentOwner.value) return;
  try {
    if (currentOwner.value.active) await ownersApi.archive(currentOwner.value.id);
    else await ownersApi.reactivate(currentOwner.value.id);
    await load();
  } catch (e) {
    reportError(e);
  }

});
}
async function deleteOwner() {
return runAction(async () => {
  if (
    !currentOwner.value ||
    !(await confirmAction({
      title: 'Delete owner',
      message: 'Delete this owner permanently when safe?',
      confirmLabel: 'Delete owner',
      danger: true,
    }))
  )
    return;
  try {
    await ownersApi.delete(currentOwner.value.id);
    owners.value = owners.value.filter((x) => x.id !== currentOwner.value!.id);
    selectedOwner.value = owners.value[0]?.id ?? '';
    emit('notify', 'Owner deleted.');
  } catch (e) {
    reportError(e);
  }

});
}
async function postTransaction() {
return runAction(async () => {
  if (!selectedOwner.value) return;
  const amount = parseMoneyInput(txForm.value.amountText, props.currencyUnit);
  if (amount === null || amount <= 0) {
    toast.error('Enter a valid positive amount.', 'Owners');
    return;
  }
  try {
    await ownersApi.createTransaction({
      ...txForm.value,
      ownerId: selectedOwner.value,
      amountRial: amount,
    });
    await load();
    emit('notify', 'Owner transaction posted.');
  } catch (e) {
    reportError(e);
  }

});
}
async function createPeriod() {
return runAction(async () => {
  try {
    const v = await ownersApi.createPeriod(periodForm.value);
    periods.value = [v, ...periods.value];
    selectedPeriod.value = v.id;
    emit('notify', 'Fiscal period created.');
  } catch (e) {
    reportError(e);
  }

});
}
async function previewPeriod() {
return runAction(async () => {
  if (!selectedPeriod.value) return;
  try {
    const v = await ownersApi.previewPeriod(selectedPeriod.value);
    periods.value = periods.value.map((x) => (x.id === v.id ? v : x));
  } catch (e) {
    reportError(e);
  }

});
}
async function closePeriod() {
return runAction(async () => {
  if (
    !selectedPeriod.value ||
    !(await confirmAction({
      title: 'Close fiscal period',
      message: 'Close this fiscal period permanently?',
      confirmLabel: 'Close period',
      danger: true,
    }))
  )
    return;
  try {
    const v = await ownersApi.closePeriod(selectedPeriod.value);
    periods.value = periods.value.map((x) => (x.id === v.id ? v : x));
    emit('notify', 'Fiscal period closed.');
  } catch (e) {
    reportError(e);
  }

});
}
onMounted(load);
async function reverseTransaction(id: string) {
return runAction(async () => { try { await ownersApi.reverseTransaction(id); await load(); emit('notify','Owner transaction reversed.'); } catch(error) { reportError(error); } 
});
}
return {busy,runAction,pageLoading,runLoad,tabs,tab,owners,transactions,periods,accounts,selectedOwner,selectedPeriod,editing,saving,ownerForm,shareForm,txForm,periodForm,currentOwner,selectedPeriodRow,filteredTransactions,ownerName,money,load,syncShares,loadTransactions,startOwner,saveOwner,saveShares,setActive,deleteOwner,postTransaction,createPeriod,previewPeriod,closePeriod,reverseTransaction}
}
