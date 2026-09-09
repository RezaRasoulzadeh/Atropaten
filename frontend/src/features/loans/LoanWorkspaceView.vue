<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ArrowLeft, CalendarDays, CircleDollarSign, Plus, RotateCcw } from 'lucide-vue-next'
import AppPanel from '../../components/layout/AppPanel.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import FormField from '../../components/ui/FormField.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { loansApi, type LoanPaymentRecord, type LoanRecord } from '../../api/loans'
import { accountingApi, type FinancialAccountRecord } from '../../api/accounting'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { currentCanonicalDate, formatDateTime } from '../../utils/date'
import { confirmAction, useToast } from '../../ui/feedback'

const props = defineProps<{ loanId: string | null; currencyUnit: CurrencyUnit; isNew?: boolean }>()
const emit = defineEmits<{ back: []; notify: [string] }>()
const { busy, runAction } = useWorkspaceActions()
const toast = useToast()

const loading = ref(false)
const creating = ref(Boolean(props.isNew))
const loan = ref<LoanRecord | null>(null)
const accounts = ref<FinancialAccountRecord[]>([])
const payments = ref<LoanPaymentRecord[]>([])
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
})
const pay = ref({
  principalRial: '0',
  interestRial: '0',
  financialAccountId: '',
  installmentId: '',
})

const directionOptions = [
  { label: 'Payable / borrowed', value: 'payable' },
  { label: 'Receivable / lent', value: 'receivable' },
]

function tone(value: string) {
  return value === 'Closed' ? 'green' : value === 'Active' ? 'blue' : 'amber'
}

function date(value: string) {
  try {
    return formatDateTime(value)
  } catch {
    return '—'
  }
}

function resetForm() {
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
  }
}

async function load() {
  loading.value = true
  try {
    accounts.value = await accountingApi.financialAccounts()
    if (creating.value) {
      resetForm()
      return
    }
    if (!props.loanId) return
    loan.value = await loansApi.get(props.loanId)
    payments.value = await loansApi.payments(props.loanId)
    pay.value.financialAccountId = accounts.value[0]?.id ?? ''
  } catch (error) {
    reportError(error)
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function create() {
  return runAction(async () => {
    const principal = Number(form.value.principalRial.replaceAll(',', ''))
    const interest = Number(form.value.interestFeeRial.replaceAll(',', ''))
    if (!form.value.counterpartyName || !principal || !form.value.financialAccountId) {
      toast.error('Counterparty, account, and positive principal are required.', 'Loans')
      return
    }
    try {
      loan.value = await loansApi.create({
        ...form.value,
        principalRial: principal,
        interestFeeRial: interest,
        installmentCount: Number(form.value.installmentCount),
      })
      payments.value = []
      creating.value = false
      emit('notify', 'Loan opened with a balanced journal entry.')
    } catch (error) {
      reportError(error)
    }
  })
}

async function refreshLoan() {
  if (!loan.value) return
  loan.value = await loansApi.get(loan.value.id)
  payments.value = await loansApi.payments(loan.value.id)
}

async function recordPayment() {
  return runAction(async () => {
    if (!loan.value) return
    const principal = Number(pay.value.principalRial.replaceAll(',', ''))
    const interest = Number(pay.value.interestRial.replaceAll(',', ''))
    if ((!principal && !interest) || !pay.value.financialAccountId || !pay.value.installmentId) {
      toast.error('Select an installment, account, and payment amount.', 'Loans')
      return
    }
    try {
      await loansApi.createPayment({
        loanId: loan.value.id,
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
      })
      await refreshLoan()
      pay.value = { ...pay.value, principalRial: '0', interestRial: '0' }
      emit('notify', 'Loan payment posted and allocated.')
    } catch (error) {
      reportError(error)
    }
  })
}

async function reversePayment(payment: LoanPaymentRecord) {
  return runAction(async () => {
    try {
      await loansApi.reversePayment(payment.id)
      await refreshLoan()
      emit('notify', 'Payment reversed with a compensating entry.')
    } catch (error) {
      reportError(error)
    }
  })
}
</script>

<template>
  <div class="min-w-0 space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        eyebrow="Finance / financing / loan"
        :title="loan?.loanNumber || (creating ? 'New loan' : 'Loan workspace')"
        :description="creating ? 'Open a payable or receivable loan with a persisted installment schedule.' : `${loan?.counterpartyName || 'Loan'} · ${loan?.direction === 'payable' ? 'Payable' : 'Receivable'}`"
      >
        <template #leading>
          <button class="btn btn-ghost btn-sm gap-1.5" type="button" @click="emit('back')">
            <ArrowLeft :size="16" aria-hidden="true" />
            <span class="hidden sm:inline">Loans</span>
          </button>
        </template>
        <StatusBadge v-if="loan" :label="loan.status" :tone="tone(loan.status)" />
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <LoadingState v-if="loading" label="Loading loan workspace…" />

    <form v-else-if="creating" class="min-w-0" @submit.prevent="create">
      <AppPanel title="Open a new loan" subtitle="Opening the loan posts principal to cash/bank and the loan balance.">
        <FormGrid>
          <SelectField v-model="form.direction" label="Type" :options="directionOptions" />
          <FormField label="Counterparty"><AppInput v-model="form.counterpartyName" class="input w-full min-w-0" required placeholder="Customer, supplier, or lender" /></FormField>
          <FormField label="Principal (Rial)"><AppInput v-model="form.principalRial" class="input w-full min-w-0" money="Rial" inputmode="numeric" /></FormField>
          <FormField label="Interest / fees (Rial)"><AppInput v-model="form.interestFeeRial" class="input w-full min-w-0" money="Rial" inputmode="numeric" /></FormField>
          <FormField label="Start date"><JalaliDatePicker v-model="form.startDate" /></FormField>
          <FormField label="End date"><JalaliDatePicker v-model="form.endDate" /></FormField>
          <FormField label="Installments"><AppInput v-model="form.installmentCount" class="input w-full min-w-0" type="number" min="1" /></FormField>
          <SelectField v-model="form.financialAccountId" label="Cash / bank account" :options="accounts.map((account) => ({ label: account.name, value: account.id }))" />
        </FormGrid>
        <FormField class="mt-4" label="Notes"><AppTextarea v-model="form.notes" rows="3" /></FormField>
        <div class="mt-5 flex flex-wrap justify-end gap-2">
          <button class="btn btn-ghost" type="button" @click="emit('back')">Cancel</button>
          <button class="btn btn-primary" type="submit" :disabled="busy"><Plus :size="15" /> Open loan</button>
        </div>
      </AppPanel>
    </form>

    <EmptyState v-else-if="!loan" title="Loan unavailable" description="This loan could not be loaded. Return to the loan register and try again.">
      <template #icon><CircleDollarSign :size="22" aria-hidden="true" /></template>
      <template #action><button class="btn btn-primary btn-sm" type="button" @click="emit('back')">Back to loans</button></template>
    </EmptyState>

    <template v-else>
      <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <div class="min-w-0 space-y-4">
          <AppPanel title="Loan overview" subtitle="Current financing position and persisted schedule.">
            <div class="grid min-w-0 gap-4 sm:grid-cols-[minmax(0,1fr)_18rem] sm:items-start">
              <div class="flex min-w-0 items-start gap-3">
                <div class="grid size-11 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><CircleDollarSign :size="23" aria-hidden="true" /></div>
                <div class="min-w-0">
                  <p class="text-xs font-medium uppercase tracking-wide text-primary/80">{{ loan.direction === 'payable' ? 'Payable loan' : 'Receivable loan' }}</p>
                  <p class="mt-1 text-2xl font-bold tabular-nums">{{ formatMoney(loan.principalRial + loan.interestFeeRial, props.currencyUnit) }}</p>
                  <p class="mt-1 text-xs text-base-content/60">{{ loan.counterpartyName }} · {{ loan.installments.length }} installments</p>
                </div>
              </div>
              <dl class="grid min-w-0 gap-2 rounded-box border border-base-300 bg-base-200/30 p-3 text-sm">
                <div class="flex items-center justify-between gap-3"><dt class="text-xs text-base-content/55">Started</dt><dd class="text-end text-xs">{{ date(loan.startDate) }}</dd></div>
                <div class="flex items-center justify-between gap-3"><dt class="text-xs text-base-content/55">End date</dt><dd class="text-end text-xs">{{ loan.endDate ? date(loan.endDate) : 'Open term' }}</dd></div>
                <div class="flex items-center justify-between gap-3"><dt class="text-xs text-base-content/55">Account</dt><dd class="max-w-36 truncate text-end text-xs">{{ loan.financialAccountId || 'No account' }}</dd></div>
              </dl>
            </div>
          </AppPanel>

          <AppPanel title="Installment schedule" :subtitle="`${loan.installments.length} scheduled installments`" :flush="true">
            <DataTable v-if="loan.installments.length" label="Loan installment schedule">
              <thead><tr><th scope="col">Installment</th><th scope="col">Due</th><th scope="col" class="text-end">Paid</th><th scope="col" class="text-end">Remaining</th><th scope="col">Status</th></tr></thead>
              <tbody>
                <tr v-for="installment in loan.installments" :key="installment.id">
                  <DataTableCell><strong>#{{ installment.position + 1 }}</strong><span class="mt-1 block text-xs text-base-content/55">{{ formatMoney(installment.totalDueRial, props.currencyUnit) }} due</span></DataTableCell>
                  <DataTableCell>{{ date(installment.dueDate) }}</DataTableCell>
                  <DataTableCell numeric>{{ formatMoney(installment.paidRial, props.currencyUnit) }}</DataTableCell>
                  <DataTableCell numeric>{{ formatMoney(installment.remainingRial, props.currencyUnit) }}</DataTableCell>
                  <DataTableCell><StatusBadge :label="installment.status" :tone="installment.status === 'Paid' ? 'green' : installment.status === 'Overdue' ? 'red' : 'amber'" /></DataTableCell>
                </tr>
              </tbody>
            </DataTable>
            <EmptyState v-else title="No installments" description="This loan has no generated schedule."><template #icon><CalendarDays :size="21" aria-hidden="true" /></template></EmptyState>
          </AppPanel>

          <AppPanel v-if="loan.notes" title="Notes" subtitle="Captured when this loan was opened."><p class="whitespace-pre-wrap text-sm leading-6">{{ loan.notes }}</p></AppPanel>
        </div>

        <aside class="min-w-0 space-y-4">
          <AppPanel title="Balance" subtitle="Derived from posted allocations.">
            <dl class="grid min-w-0 gap-1 text-sm">
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2"><dt class="text-xs text-base-content/60">Principal remaining</dt><dd class="text-end tabular-nums">{{ formatMoney(loan.remainingPrincipalRial, props.currencyUnit) }}</dd></div>
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-2"><dt class="text-xs text-base-content/60">Interest remaining</dt><dd class="text-end tabular-nums">{{ formatMoney(loan.remainingInterestRial, props.currencyUnit) }}</dd></div>
              <div class="flex items-center justify-between gap-4 border-b border-base-300 py-3 font-semibold"><dt>Remaining</dt><dd class="text-end tabular-nums text-primary">{{ formatMoney(loan.remainingPrincipalRial + loan.remainingInterestRial, props.currencyUnit) }}</dd></div>
              <div class="flex items-center justify-between gap-4 py-2"><dt class="text-xs text-base-content/60">Overdue</dt><dd class="text-end tabular-nums text-warning">{{ formatMoney(loan.overdueRial, props.currencyUnit) }}</dd></div>
            </dl>
          </AppPanel>

          <AppPanel title="Record payment" subtitle="Allocate a payment to one installment.">
            <FormGrid>
              <SelectField v-model="pay.installmentId" label="Installment" :options="[{ label: 'Select installment', value: '' }, ...loan.installments.filter((item) => item.remainingRial > 0).map((item) => ({ label: `#${item.position + 1} · ${formatMoney(item.remainingRial, props.currencyUnit)}`, value: item.id }))]" />
              <SelectField v-model="pay.financialAccountId" label="Cash / bank account" :options="[{ label: 'Select account', value: '' }, ...accounts.map((account) => ({ label: account.name, value: account.id }))]" />
              <FormField label="Principal (Rial)"><AppInput v-model="pay.principalRial" class="input w-full min-w-0" money="Rial" inputmode="numeric" /></FormField>
              <FormField label="Interest (Rial)"><AppInput v-model="pay.interestRial" class="input w-full min-w-0" money="Rial" inputmode="numeric" /></FormField>
            </FormGrid>
            <button class="btn btn-primary btn-sm mt-4 w-full" type="button" :disabled="busy" @click="recordPayment">Record payment</button>
          </AppPanel>

          <AppPanel title="Payment history" subtitle="Posted allocations and reversals.">
            <DataTable v-if="payments.length" label="Loan payment history">
              <thead><tr><th scope="col">Posted</th><th scope="col" class="text-end">Amount</th><th scope="col">Action</th></tr></thead>
              <tbody>
                <tr v-for="payment in payments" :key="payment.id">
                  <DataTableCell><span class="whitespace-nowrap">{{ date(payment.paidAt) }}</span><span class="mt-1 block text-xs text-base-content/55">{{ payment.paymentNumber }}</span></DataTableCell>
                  <DataTableCell numeric>{{ formatMoney(payment.amountRial, props.currencyUnit) }}</DataTableCell>
                  <DataTableCell><button v-if="payment.status === 'Posted'" class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="reversePayment(payment)"><RotateCcw :size="14" /> Reverse</button><StatusBadge v-else :label="payment.status" tone="slate" /></DataTableCell>
                </tr>
              </tbody>
            </DataTable>
            <EmptyState v-else title="No payments" description="Payments recorded against this loan will appear here."><template #icon><CircleDollarSign :size="21" aria-hidden="true" /></template></EmptyState>
          </AppPanel>
        </aside>
      </div>
    </template>
  </div>
</template>
