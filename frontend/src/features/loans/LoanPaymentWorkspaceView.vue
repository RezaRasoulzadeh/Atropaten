<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CircleDollarSign, CreditCard, Edit3, Plus, RotateCcw, Save, Trash2, X } from 'lucide-vue-next'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import FormField from '../../components/ui/FormField.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import { accountingApi, type FinancialAccountRecord } from '../../api/accounting'
import { loansApi, type LoanPaymentPayload, type LoanPaymentRecord, type LoanRecord } from '../../api/loans'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { currentCanonicalDate, formatDate } from '../../utils/date'
import { formatLocalizedNumber } from '../../utils/number'
import { confirmAction, useToast } from '../../ui/feedback'

type EditorMode = 'none' | 'new' | 'edit'
type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate'

const props = defineProps<{ loanId: string; currencyUnit: CurrencyUnit; startNew?: boolean }>()
const emit = defineEmits<{ back: []; notify: [message: string] }>()
const { busy, runAction } = useWorkspaceActions()
const toast = useToast()
const loading = ref(false)
const loan = ref<LoanRecord | null>(null)
const accounts = ref<FinancialAccountRecord[]>([])
const payments = ref<LoanPaymentRecord[]>([])
const mode = ref<EditorMode>(props.startNew ? 'new' : 'none')
const editing = ref<LoanPaymentRecord | null>(null)
const pendingPaymentId = ref('')
const form = ref(emptyForm())

function emptyForm() {
  return {
    installmentId: '',
    financialAccountId: '',
    paidAt: currentCanonicalDate(),
    notes: '',
  }
}

const activeAccounts = computed(() => accounts.value.filter((account) => account.active && (account.type === 'cash' || account.type === 'bank')))
const accountOptions = computed(() => [{ label: 'Select account', value: '' }, ...activeAccounts.value.map((account) => ({ label: account.name, value: account.id }))])
const selectableInstallments = computed(() => (loan.value?.installments ?? [])
  .filter((installment) => installment.remainingRial > 0 || installment.id === editing.value?.allocations[0]?.installmentId))
const selectedInstallment = computed(() => selectableInstallments.value.find((installment) => installment.id === form.value.installmentId) ?? null)
const automaticAllocation = computed(() => {
  const installment = selectedInstallment.value
  if (!installment) return { principal: 0, interest: 0, total: 0 }
  const existingAllocation = editing.value?.allocations[0]
  if (existingAllocation?.installmentId === installment.id) {
    return {
      principal: editing.value?.principalRial ?? existingAllocation.principalRial,
      interest: editing.value?.interestRial ?? existingAllocation.interestRial,
      total: editing.value?.amountRial ?? existingAllocation.principalRial + existingAllocation.interestRial,
    }
  }
  const principal = Math.max(0, installment.principalRial - installment.paidPrincipalRial)
  const interest = Math.max(0, installment.interestFeeRial - installment.paidInterestRial)
  return { principal, interest, total: principal + interest }
})
const postedCount = computed(() => payments.value.filter((payment) => payment.status.toLowerCase() === 'posted').length)
const remainingRial = computed(() => (loan.value?.remainingPrincipalRial ?? 0) + (loan.value?.remainingInterestRial ?? 0))

function tone(value: string): Tone {
  if (value.toLowerCase() === 'posted' || value.toLowerCase() === 'active') return 'blue'
  if (value.toLowerCase() === 'closed') return 'green'
  return 'slate'
}

function installmentPaymentAmount(installment: LoanRecord['installments'][number]) {
  const allocation = editing.value?.allocations[0]
  return allocation?.installmentId === installment.id ? editing.value?.amountRial ?? installment.remainingRial : installment.remainingRial
}

function resetEditor(nextMode: EditorMode = 'none') {
  mode.value = nextMode
  editing.value = null
  pendingPaymentId.value = ''
  form.value = emptyForm()
  form.value.financialAccountId = activeAccounts.value[0]?.id ?? ''
}

async function load() {
  loading.value = true
  try {
    const [record, financialAccounts, loanPayments] = await Promise.all([
      loansApi.get(props.loanId),
      accountingApi.financialAccounts(),
      loansApi.payments(props.loanId),
    ])
    loan.value = record
    accounts.value = financialAccounts
    payments.value = loanPayments
    resetEditor(props.startNew && record.status !== 'Closed' ? 'new' : 'none')
  } catch (error) {
    reportError(error)
  } finally {
    loading.value = false
  }
}

async function refresh() {
  const [record, loanPayments] = await Promise.all([loansApi.get(props.loanId), loansApi.payments(props.loanId)])
  loan.value = record
  payments.value = loanPayments
}

function newPayment() {
  if (loan.value?.status === 'Closed') return
  resetEditor('new')
}

function editPayment(payment: LoanPaymentRecord) {
  if (payment.status.toLowerCase() !== 'posted' || payment.allocations.length !== 1) return
  mode.value = 'edit'
  editing.value = payment
  pendingPaymentId.value = ''
  const allocation = payment.allocations[0]
  form.value = {
    installmentId: allocation.installmentId,
    financialAccountId: payment.financialAccountId,
    paidAt: payment.paidAt.slice(0, 10),
    notes: payment.notes,
  }
}

function makePaymentId() {
  return `LPAY-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

async function save() {
  return runAction(async () => {
    if (!loan.value) return
    if (!form.value.installmentId || !form.value.financialAccountId || !form.value.paidAt) {
      toast.error('Select an installment, payment date, and active cash or bank account.', 'Loans')
      return
    }
    const { principal, interest, total } = automaticAllocation.value
    if (total <= 0) {
      toast.error('The selected installment has no remaining balance.', 'Loans')
      return
    }
    if (!pendingPaymentId.value) pendingPaymentId.value = makePaymentId()
    const payload: LoanPaymentPayload = {
      id: pendingPaymentId.value,
      idempotencyKey: pendingPaymentId.value,
      loanId: loan.value.id,
      financialAccountId: form.value.financialAccountId,
      paidAt: form.value.paidAt,
      amountRial: total,
      principalRial: principal,
      interestRial: interest,
      notes: form.value.notes.trim(),
      allocations: [{ installmentId: form.value.installmentId, principalRial: principal, interestRial: interest }],
    }
    try {
      if (mode.value === 'edit' && editing.value) await loansApi.updatePayment(editing.value.id, payload)
      else await loansApi.createPayment(payload)
      const message = mode.value === 'edit' ? 'Loan payment updated with an auditable correcting entry.' : 'Loan payment posted and allocated.'
      await refresh()
      resetEditor('none')
      emit('notify', message)
    } catch (error) {
      reportError(error)
    }
  })
}

async function removePayment(payment: LoanPaymentRecord) {
  if (!(await confirmAction({
    title: 'Remove loan payment',
    message: 'Reverse this payment and restore the loan balance? The original posting stays in accounting history and a compensating entry is created.',
    confirmLabel: 'Remove payment',
    danger: true,
  }))) return
  return runAction(async () => {
    try {
      await loansApi.reversePayment(payment.id, `loan-payment:remove:${payment.id}`)
      await refresh()
      if (editing.value?.id === payment.id) resetEditor('none')
      emit('notify', 'Loan payment removed with a compensating accounting entry.')
    } catch (error) {
      reportError(error)
    }
  })
}

onMounted(load)
</script>

<template>
  <div class="loan-payment-wizard w-full flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" :aria-label='$t("Loan payment manager")'>
    <header class="loan-payment-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">{{ $t("Loan payments") }}</h1>
        <p class="mt-1 text-xs leading-4 text-base-content/65">{{ $ui(loan ? `${loan.loanNumber} · ${loan.counterpartyName}` : 'Create, correct, or remove loan payments.') }}</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Loans' }, { label: 'Loan payments', current: true }]" @navigate="emit('back')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="busy" @click="emit('back')">{{ $t("Cancel") }}</button>
        <button class="btn btn-success gap-2" type="button" :disabled="loading || busy || loan?.status === 'Closed'" @click="newPayment"><Plus :size="16" />{{ $t("New payment") }}</button>
      </div>
    </header>

    <LoadingState v-if="loading" :label='$t("Loading loan payments…")' />
    <div v-else-if="loan" class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(20rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.1fr)_minmax(24rem,0.9fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100">
        <div class="grid shrink-0 grid-cols-3 divide-x border-b border-base-300 bg-base-200/20">
          <div class="min-w-0 p-3"><span class="block text-xs text-base-content/55">{{ $t("Remaining") }}</span><strong class="mt-1 block truncate text-sm tabular-nums text-primary">{{ formatMoney(remainingRial, currencyUnit) }}</strong></div>
          <div class="min-w-0 p-3"><span class="block text-xs text-base-content/55">{{ $t("Posted payments") }}</span><strong class="mt-1 block text-sm tabular-nums">{{ formatLocalizedNumber(postedCount) }}</strong></div>
          <div class="min-w-0 p-3"><span class="block text-xs text-base-content/55">{{ $t("Status") }}</span><StatusBadge class="mt-1" :label="loan.status" :tone="tone(loan.status)" /></div>
        </div>
        <div v-if="payments.length" class="min-h-0 flex-1 divide-y divide-base-300 overflow-y-auto">
          <article v-for="payment in payments" :key="payment.id" class="p-4" :class="payment.status.toLowerCase() === 'reversed' ? 'bg-base-200/30 opacity-70' : ''">
            <div class="flex min-w-0 items-start justify-between gap-3"><div class="min-w-0"><div class="flex items-center gap-2"><strong class="truncate text-sm">{{ payment.paymentNumber }}</strong><StatusBadge :label="payment.status" :tone="tone(payment.status)" /></div><p class="mt-1 text-xs text-base-content/55">{{ formatDate(payment.paidAt) }} · {{ accounts.find((account) => account.id === payment.financialAccountId)?.name || payment.financialAccountId }}</p></div><strong class="shrink-0 text-sm tabular-nums text-primary">{{ formatMoney(payment.amountRial, currencyUnit) }}</strong></div>
            <div class="mt-3 grid grid-cols-2 gap-2 text-xs text-base-content/60"><span>{{ $t("Principal") }}: {{ formatMoney(payment.principalRial, currencyUnit) }}</span><span>{{ $t("Interest") }}: {{ formatMoney(payment.interestRial, currencyUnit) }}</span></div>
            <p v-if="payment.notes" class="mt-2 truncate text-xs text-base-content/55">{{ payment.notes }}</p>
            <div v-if="payment.status.toLowerCase() === 'posted'" class="mt-3 flex justify-end gap-2 border-t border-base-300 pt-3"><button class="btn btn-outline btn-xs gap-1" type="button" :disabled="busy || payment.allocations.length !== 1" @click="editPayment(payment)"><Edit3 :size="13" />{{ $t("Edit") }}</button><button class="btn btn-outline btn-error btn-xs gap-1" type="button" :disabled="busy" @click="removePayment(payment)"><Trash2 :size="13" />{{ $t("Remove") }}</button></div>
            <div v-else class="mt-3 flex items-center gap-1 border-t border-base-300 pt-3 text-xs text-base-content/50"><RotateCcw :size="12" />{{ $t("Accounting reversal retained") }}</div>
          </article>
        </div>
        <EmptyState v-else :title='$t("No payments")' :description='$t("Record the first payment for this loan.")'><template #icon><CreditCard :size="22" /></template><template #action><button class="btn btn-primary btn-sm gap-2" type="button" :disabled="loan.status === 'Closed'" @click="newPayment"><Plus :size="14" />{{ $t("Record payment") }}</button></template></EmptyState>
      </section>

      <section class="min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:overflow-y-auto">
        <form v-if="mode !== 'none'" data-enter-scope class="p-4 sm:p-5" @submit.prevent="save">
          <div class="flex items-start justify-between gap-3 border-b border-base-300 pb-4"><div><h2 class="font-semibold">{{ $t(mode === 'edit' ? 'Edit payment' : 'Record payment') }}</h2><p class="mt-1 text-sm leading-6 text-base-content/60">{{ $t(mode === 'edit' ? 'Saving creates a correcting accounting entry and preserves the original.' : 'Allocate this payment to one installment.') }}</p></div><button class="btn btn-ghost btn-sm btn-square" type="button" :aria-label='$t("Close editor")' @click="resetEditor('none')"><X :size="16" /></button></div>
          <div class="mt-5">
            <div class="flex items-center justify-between gap-3"><h3 class="text-sm font-semibold">{{ $t("Choose installment") }}</h3><span class="text-xs text-base-content/55">{{ $t("The remaining principal and interest are calculated automatically.") }}</span></div>
            <div class="mt-3 grid gap-2">
              <button v-for="installment in selectableInstallments" :key="installment.id" class="rounded-box border p-3 text-start transition-colors" :class="form.installmentId === installment.id ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-200/20 hover:border-primary/50'" type="button" @click="form.installmentId = installment.id">
                <span class="flex items-start justify-between gap-3"><span><strong class="block text-sm">{{ $t("Installment") }} #{{ formatLocalizedNumber(installment.position + 1) }}</strong><span class="mt-1 block text-xs text-base-content/55">{{ formatDate(installment.dueDate) }}</span></span><strong class="text-sm tabular-nums text-primary">{{ formatMoney(installmentPaymentAmount(installment), currencyUnit) }}</strong></span>
              </button>
            </div>
            <p v-if="!selectableInstallments.length" class="mt-3 rounded-box border border-dashed border-base-300 p-4 text-center text-sm text-base-content/60">{{ $t("All installments are fully paid.") }}</p>
          </div>

          <div v-if="selectedInstallment" class="mt-4 rounded-box border border-primary/30 bg-primary/5 p-4">
            <div class="flex items-center justify-between gap-3"><h3 class="text-sm font-semibold">{{ $t("Automatic payment allocation") }}</h3><strong class="text-base tabular-nums text-primary">{{ formatMoney(automaticAllocation.total, currencyUnit) }}</strong></div>
            <dl class="mt-3 grid grid-cols-2 gap-3 text-sm"><div><dt class="text-xs text-base-content/55">{{ $t("Principal") }}</dt><dd class="mt-1 tabular-nums">{{ formatMoney(automaticAllocation.principal, currencyUnit) }}</dd></div><div><dt class="text-xs text-base-content/55">{{ $t("Interest") }}</dt><dd class="mt-1 tabular-nums">{{ formatMoney(automaticAllocation.interest, currencyUnit) }}</dd></div></dl>
          </div>

          <div class="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-1 2xl:grid-cols-2">
            <SelectField v-model="form.financialAccountId" :label='$t("Cash / bank account")' :options="accountOptions" />
            <FormField :label='$t("Payment date")'><JalaliDatePicker v-model="form.paidAt" /></FormField>
            <FormField class="sm:col-span-2 xl:col-span-1 2xl:col-span-2" :label='$t("Notes")'><AppTextarea v-model="form.notes" rows="3" /></FormField>
          </div>
          <div class="mt-5 flex justify-end gap-2"><button class="btn btn-ghost" type="button" @click="resetEditor('none')">{{ $t("Cancel") }}</button><button class="btn btn-primary gap-2" type="submit" :disabled="busy || !selectedInstallment"><Save :size="15" />{{ $t(mode === 'edit' ? 'Save payment changes' : 'Record payment') }}</button></div>
        </form>
        <div v-else class="flex min-h-72 h-full items-center justify-center p-8 text-center"><EmptyState :title='$t("Select a payment")' :description='$t("Edit an existing payment or create a new one.")'><template #icon><CircleDollarSign :size="22" /></template><template #action><button class="btn btn-primary btn-sm gap-2" type="button" :disabled="loan.status === 'Closed'" @click="newPayment"><Plus :size="14" />{{ $t("New payment") }}</button></template></EmptyState></div>
      </section>
    </div>
  </div>
</template>
