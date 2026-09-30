<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CalendarDays, CircleDollarSign, Landmark, Save, UserRound } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SelectField from '../../components/ui/SelectField.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import { accountingApi, type FinancialAccountRecord } from '../../api/accounting'
import { loansApi, type LoanPayload, type LoanRecord } from '../../api/loans'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { formatMoney, formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency'
import { currentCanonicalDate, formatDate } from '../../utils/date'
import { formatLocalizedNumber } from '../../utils/number'
import { useToast } from '../../ui/feedback'

const props = defineProps<{ loanId: string | null; currencyUnit: CurrencyUnit; isNew?: boolean }>()
const emit = defineEmits<{ back: []; notify: [message: string]; created: [loan: LoanRecord]; updated: [loan: LoanRecord] }>()
const { busy, runAction } = useWorkspaceActions()
const toast = useToast()
const loading = ref(false)
const accounts = ref<FinancialAccountRecord[]>([])
const originalLoan = ref<LoanRecord | null>(null)
const paymentHistory = ref(false)
const pendingUpdateKey = ref('')
const form = ref({
  direction: 'payable',
  counterpartyName: '',
  principal: formatMoneyInput(0, props.currencyUnit),
  interest: formatMoneyInput(0, props.currencyUnit),
  startDate: currentCanonicalDate(),
  endDate: '',
  financialAccountId: '',
  installmentCount: '1',
  notes: '',
})

const directionOptions = [
  { label: 'Payable / borrowed', value: 'payable' },
  { label: 'Receivable / lent', value: 'receivable' },
]
const isNew = computed(() => props.isNew || !props.loanId)
const financialLocked = computed(() => !isNew.value && paymentHistory.value)
const activeAccounts = computed(() => accounts.value.filter((account) => account.active && (account.type === 'cash' || account.type === 'bank')))
const accountOptions = computed(() => accounts.value.filter((account) => account.active || account.id === form.value.financialAccountId).map((account) => ({ label: `${account.name} · ${account.type}`, value: account.id })))
const principalRial = computed(() => parseMoneyInput(form.value.principal, props.currencyUnit) ?? 0)
const interestRial = computed(() => parseMoneyInput(form.value.interest, props.currencyUnit) ?? 0)
const totalRial = computed(() => principalRial.value + interestRial.value)
const installmentCount = computed(() => Math.max(1, Number(form.value.installmentCount) || 1))
const installmentEstimate = computed(() => Math.ceil(totalRial.value / installmentCount.value))

async function load() {
  loading.value = true
  try {
    accounts.value = await accountingApi.financialAccounts()
    if (isNew.value || !props.loanId) {
      form.value.financialAccountId = activeAccounts.value[0]?.id ?? ''
    } else {
      const [record, payments] = await Promise.all([loansApi.get(props.loanId), loansApi.payments(props.loanId)])
      originalLoan.value = record
      paymentHistory.value = payments.length > 0
      form.value = {
        direction: record.direction,
        counterpartyName: record.counterpartyName,
        principal: formatMoneyInput(record.principalRial, props.currencyUnit),
        interest: formatMoneyInput(record.interestFeeRial, props.currencyUnit),
        startDate: record.startDate.slice(0, 10),
        endDate: record.endDate ? record.endDate.slice(0, 10) : '',
        financialAccountId: record.financialAccountId,
        installmentCount: String(record.installments.length || 1),
        notes: record.notes,
      }
    }
  } catch (error) {
    reportError(error)
  } finally {
    loading.value = false
  }
}

async function save() {
  return runAction(async () => {
    const principal = parseMoneyInput(form.value.principal, props.currencyUnit)
    const interest = parseMoneyInput(form.value.interest, props.currencyUnit)
    const count = Number(form.value.installmentCount)
    if (!form.value.counterpartyName.trim()) {
      toast.error('Enter the customer, supplier, or lender name.', 'Loans')
      return
    }
    if (principal === null || principal <= 0 || interest === null || interest < 0) {
      toast.error('Enter a positive principal and a valid interest or fee amount.', 'Loans')
      return
    }
    if (!Number.isInteger(count) || count < 1) {
      toast.error('Installment count must be a positive whole number.', 'Loans')
      return
    }
    if (!financialLocked.value && !activeAccounts.value.some((account) => account.id === form.value.financialAccountId)) {
      toast.error('Select an active cash or bank account.', 'Loans')
      return
    }
    try {
      const payload: LoanPayload = {
        direction: form.value.direction,
        counterpartyName: form.value.counterpartyName.trim(),
        principalRial: principal,
        interestFeeRial: interest,
        startDate: form.value.startDate,
        endDate: form.value.endDate,
        financialAccountId: form.value.financialAccountId,
        installmentCount: count,
        notes: form.value.notes.trim(),
      }
      const original = originalLoan.value
      const financialChanged = Boolean(original && (
        original.direction !== form.value.direction ||
        original.principalRial !== principal ||
        original.interestFeeRial !== interest ||
        original.startDate.slice(0, 10) !== form.value.startDate ||
        original.financialAccountId !== form.value.financialAccountId ||
        original.installments.length !== count
      ))
      if (original && !financialChanged) {
        payload.installmentCount = undefined
        payload.installments = original.installments.map((installment) => ({
          id: installment.id,
          dueDate: installment.dueDate,
          principalRial: installment.principalRial,
          interestFeeRial: installment.interestFeeRial,
        }))
      }
      if (isNew.value) {
        const created = await loansApi.create(payload)
        emit('notify', 'Loan opened with a balanced journal entry.')
        emit('created', created)
      } else if (props.loanId) {
        if (!pendingUpdateKey.value) pendingUpdateKey.value = `loan-update-${props.loanId}-${Date.now()}`
        const updated = await loansApi.update(props.loanId, { ...payload, idempotencyKey: pendingUpdateKey.value })
        pendingUpdateKey.value = ''
        emit('notify', paymentHistory.value ? 'Loan details updated. Financial terms remain locked because payment history exists.' : 'Loan updated with a correcting accounting entry.')
        emit('updated', updated)
      }
    } catch (error) {
      reportError(error)
    }
  })
}

onMounted(load)
</script>

<template>
  <div class="loan-wizard w-full flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" :aria-label='$t("Loan editor")'>
    <header class="loan-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">{{ $ui(isNew ? 'Add loan' : 'Edit loan') }}</h1>
        <p class="mt-1 text-xs leading-4 text-base-content/65">{{ $t(isNew ? 'Open a payable or receivable financing agreement with an automatic installment schedule.' : 'Update this loan agreement and its repayment details.') }}</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Loans' }, { label: isNew ? 'Add loan' : 'Edit loan', current: true }]" @navigate="emit('back')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="busy" @click="emit('back')">{{ $t("Cancel") }}</button>
        <button class="btn btn-success gap-2" type="submit" form="loan-editor" :disabled="busy || !accountOptions.length"><Save :size="16" aria-hidden="true" />{{ $ui(busy ? 'Saving…' : isNew ? 'Save loan' : 'Save changes') }}</button>
      </div>
    </header>

    <LoadingState v-if="loading" :label='$t("Preparing loan editor…")' />
    <form v-else id="loan-editor" class="grid min-h-0 min-w-0 flex-1 grid-rows-[auto_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1fr)_22rem] xl:grid-rows-1 xl:overflow-hidden" @submit.prevent="save">
      <section class="min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:overflow-y-auto">
        <div class="border-b border-base-300 p-4 sm:p-5"><div class="flex items-start gap-3"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><CircleDollarSign :size="21" /></span><div><h2 class="font-semibold">{{ $t("Financing details") }}</h2><p class="mt-1 text-sm leading-6 text-base-content/60">{{ $t("Define the counterparty, opening amount, repayment term, and account used for the opening journal.") }}</p></div></div></div>

        <div class="space-y-6 p-4 sm:p-5">
          <fieldset class="space-y-3"><legend class="text-sm font-semibold">{{ $t("Identity") }}</legend><div class="grid gap-4 sm:grid-cols-2"><SelectField v-model="form.direction" :label='$t("Loan type")' :options="directionOptions" :disabled="financialLocked" /><FormField :label='$t("Counterparty")'><AppInput v-model="form.counterpartyName" required :placeholder='$t("Customer, supplier, or lender")' /></FormField></div></fieldset>

          <fieldset class="space-y-3 border-t border-base-300 pt-5"><legend class="text-sm font-semibold">{{ $t("Amounts and account") }}</legend><div class="grid gap-4 sm:grid-cols-2"><FormField :label="$ui(`Principal (${currencyUnit})`)"><AppInput v-model="form.principal" :money="currencyUnit" inputmode="numeric" :disabled="financialLocked" /></FormField><FormField :label="$ui(`Interest / fees (${currencyUnit})`)"><AppInput v-model="form.interest" :money="currencyUnit" inputmode="numeric" :disabled="financialLocked" /></FormField><SelectField v-model="form.financialAccountId" class="sm:col-span-2" :label='$t("Opening cash / bank account")' :options="accountOptions" :disabled="financialLocked" /></div><p v-if="financialLocked" class="text-xs text-warning">{{ $t("Financial terms are locked because this loan has payment history. Counterparty, end date, and notes can still be edited.") }}</p><p v-else-if="!accountOptions.length" class="text-xs text-error">{{ $t("Create an active cash or bank account before opening a loan.") }}</p></fieldset>

          <fieldset class="space-y-3 border-t border-base-300 pt-5"><legend class="text-sm font-semibold">{{ $t("Repayment schedule") }}</legend><div class="grid gap-4 sm:grid-cols-3"><FormField :label='$t("Start date")'><JalaliDatePicker v-model="form.startDate" :disabled="financialLocked" /></FormField><FormField :label='$t("End date (optional)")'><JalaliDatePicker v-model="form.endDate" /></FormField><FormField :label='$t("Installment count")'><AppInput v-model="form.installmentCount" type="number" inputmode="numeric" min="1" step="1" :disabled="financialLocked" /></FormField></div><p class="text-xs leading-5 text-base-content/55">{{ $t("The principal and interest are distributed across monthly installments. Any remainder is added to the earliest installments.") }}</p></fieldset>

          <FormField class="border-t border-base-300 pt-5" :label='$t("Notes")'><AppTextarea v-model="form.notes" rows="4" :placeholder='$t("Agreement details, reference, or internal notes")' /></FormField>
        </div>
      </section>

      <aside class="min-w-0 space-y-4 xl:overflow-y-auto">
        <section class="overflow-hidden rounded-box border border-base-300 bg-base-100">
          <div class="bg-gradient-to-br from-primary/20 via-base-200 to-base-300 p-5"><div class="grid size-11 place-items-center rounded-box border border-primary/30 bg-base-100/70 text-primary"><Landmark :size="23" /></div><p class="mt-8 text-xs font-medium uppercase tracking-wide text-primary/80">{{ $ui(form.direction === 'payable' ? 'Payable / borrowed' : 'Receivable / lent') }}</p><h2 class="mt-1 truncate text-xl font-semibold">{{ form.counterpartyName || $t('New counterparty') }}</h2></div>
          <div class="p-4">
            <h3 class="text-sm font-semibold">{{ $t("Opening summary") }}</h3>
            <dl class="mt-3 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Principal") }}</dt><dd class="tabular-nums">{{ formatMoney(principalRial, currencyUnit) }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Interest / fees") }}</dt><dd class="tabular-nums">{{ formatMoney(interestRial, currencyUnit) }}</dd></div><div class="flex justify-between gap-3 py-3 font-semibold"><dt>{{ $t("Total scheduled") }}</dt><dd class="tabular-nums text-primary">{{ formatMoney(totalRial, currencyUnit) }}</dd></div></dl>
          </div>
        </section>

        <section class="rounded-box border border-base-300 bg-base-100 p-4"><div class="flex items-center gap-2"><CalendarDays :size="18" class="text-primary" /><h3 class="text-sm font-semibold">{{ $t("Schedule preview") }}</h3></div><dl class="mt-3 space-y-2 text-sm"><div class="flex justify-between gap-3"><dt class="text-base-content/60">{{ $t("Starts") }}</dt><dd>{{ formatDate(form.startDate) }}</dd></div><div class="flex justify-between gap-3"><dt class="text-base-content/60">{{ $t("Installments") }}</dt><dd>{{ formatLocalizedNumber(installmentCount) }}</dd></div><div class="flex justify-between gap-3"><dt class="text-base-content/60">{{ $t("About each") }}</dt><dd class="tabular-nums">{{ formatMoney(installmentEstimate, currencyUnit) }}</dd></div></dl></section>

        <section class="rounded-box border border-base-300 bg-base-100 p-4"><div class="flex items-start gap-3"><UserRound :size="18" class="mt-0.5 shrink-0 text-primary" /><p class="text-xs leading-5 text-base-content/60">{{ $t("Opening the loan immediately posts a balanced accounting entry. Later payment edits and removals preserve the original history through correcting entries.") }}</p></div></section>

      </aside>
    </form>
  </div>
</template>
