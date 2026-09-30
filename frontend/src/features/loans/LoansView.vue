<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, CircleDollarSign, Plus, Search } from 'lucide-vue-next'
import EmptyState from '../../components/ui/EmptyState.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SearchField from '../../components/ui/SearchField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import { loansApi, type LoanRecord } from '../../api/loans'
import { useDynamicPagination } from '../../composables/useDynamicPagination'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { formatDate } from '../../utils/date'
import { formatLocalizedNumber } from '../../utils/number'
import { confirmAction } from '../../ui/feedback'
import LoanDetailPanel from './LoanDetailPanel.vue'

type LoanFilter = 'All' | 'Active' | 'Payable' | 'Receivable' | 'Overdue' | 'Closed'
type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate'

const props = defineProps<{ currencyUnit: CurrencyUnit; initialLoanId?: string | null }>()
const emit = defineEmits<{
  notify: [message: string]
  'new-loan': []
  'select-loan': [id: string]
  'manage-payments': [id: string, startNew: boolean]
  'edit-loan': [id: string]
  'loan-removed': [id: string]
}>()
const { busy, pageLoading, runAction, runLoad } = useWorkspaceActions()

const loans = ref<LoanRecord[]>([])
const query = ref('')
const status = ref<LoanFilter>('All')
const selectedLoanId = ref<string | null>(props.initialLoanId ?? null)
const statusOptions: LoanFilter[] = ['All', 'Active', 'Payable', 'Receivable', 'Overdue', 'Closed']

function loanMatchesFilter(loan: LoanRecord, filter: LoanFilter) {
  if (filter === 'All') return true
  if (filter === 'Payable') return loan.direction === 'payable'
  if (filter === 'Receivable') return loan.direction === 'receivable'
  if (filter === 'Overdue') return loan.overdueRial > 0
  return loan.status === filter
}

function statusCount(filter: LoanFilter) {
  return loans.value.filter((loan) => loanMatchesFilter(loan, filter)).length
}

const filteredLoans = computed(() => {
  const search = query.value.trim().toLowerCase()
  return loans.value.filter((loan) => {
    const matchesSearch = !search || [loan.loanNumber, loan.counterpartyName, loan.financialAccountId]
      .some((value) => String(value ?? '').toLowerCase().includes(search))
    return matchesSearch && loanMatchesFilter(loan, status.value)
  })
})
const selectedLoan = computed(() => loans.value.find((loan) => loan.id === selectedLoanId.value) ?? null)
const { page, pageSize, pageCount, pageNumbers, pagedItems: pagedLoans, goToPage } = useDynamicPagination(filteredLoans, { viewportSelector: '.loan-register-row' })
const pageSummary = computed(() => filteredLoans.value.length
  ? `Showing ${formatLocalizedNumber((page.value - 1) * pageSize.value + 1)}–${formatLocalizedNumber(Math.min(page.value * pageSize.value, filteredLoans.value.length))} of ${formatLocalizedNumber(filteredLoans.value.length)} loans`
  : `${formatLocalizedNumber(0)} loans`)

function tone(value: string): Tone {
  if (value === 'Closed') return 'green'
  if (value === 'Active') return 'blue'
  return 'slate'
}

function selectLoan(id: string) {
  selectedLoanId.value = id
  emit('select-loan', id)
}

function clearFilters() {
  query.value = ''
  status.value = 'All'
  page.value = 1
}

async function load() {
  return runLoad(async () => {
    try {
      loans.value = await loansApi.list()
    } catch (error) {
      reportError(error)
    }
  })
}

async function removeSelectedLoan() {
  const selected = selectedLoan.value
  if (!selected || !(await confirmAction({
    title: 'Remove loan',
    message: 'Remove this loan from the register and reverse its opening accounting entry? Posted payments must be removed first. Accounting and reversed payment history will be retained.',
    confirmLabel: 'Remove loan',
    danger: true,
  }))) return
  return runAction(async () => {
    try {
      await loansApi.remove(selected.id, `loan:remove:${selected.id}`)
      loans.value = loans.value.filter((loan) => loan.id !== selected.id)
      selectedLoanId.value = null
      emit('loan-removed', selected.id)
      emit('notify', 'Loan removed and its opening accounting entry reversed.')
    } catch (error) {
      reportError(error)
    }
  })
}

watch([query, status], () => { page.value = 1 })
watch(() => props.initialLoanId, (id) => {
  if (id) selectedLoanId.value = id
})
watch(filteredLoans, (visible) => {
  if (!visible.some((loan) => loan.id === selectedLoanId.value)) selectedLoanId.value = visible[0]?.id ?? null
}, { immediate: true })
onMounted(load)
</script>

<template>
  <div class="flex h-full min-h-0 min-w-0 flex-col overflow-hidden" :aria-label='$t("Loans workspace")'>
    <WorkspaceStickyStack class="shrink-0" :flush="true">
      <WorkspaceHeader :show-breadcrumb="true" :title='$t("Loans")' :description='$t("Track financing schedules, balances, and auditable payment activity.")'>
        <SearchField v-model="query" class="w-full min-w-0 sm:w-64" :placeholder='$t("Search loans…")' :aria-label='$t("Search loans")' />
        <button class="btn btn-primary w-full gap-2 sm:w-auto" type="button" @click="emit('new-loan')"><Plus :size="16" aria-hidden="true" />{{ $t("New loan") }}</button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <div class="grid min-h-0 min-w-0 flex-1 grid-rows-[minmax(22rem,auto)_auto] gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1.15fr)_minmax(24rem,0.85fr)] xl:grid-rows-1 xl:overflow-hidden">
      <section class="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100" :aria-label='$t("Loan register")'>
        <div class="shrink-0 border-b border-base-300 p-3 sm:p-4">
          <div class="flex min-w-0 flex-wrap items-center justify-between gap-3">
            <div class="flex min-w-0 flex-wrap items-center gap-2">
              <button v-for="filter in statusOptions" :key="filter" class="inline-flex h-9 items-center gap-2 rounded-box border px-3 text-sm transition-colors" :class="status === filter ? 'border-primary bg-primary/10 text-primary' : 'border-base-300 text-base-content/70 hover:border-primary/50 hover:text-base-content'" type="button" @click="status = filter">
                <span class="size-2 rounded-full" :class="filter === 'Overdue' ? 'bg-error' : filter === 'Closed' ? 'bg-success' : filter === 'Active' ? 'bg-info' : filter === 'All' ? 'bg-primary' : 'bg-base-content/35'"></span>
                {{ $ui(filter) }}
                <span class="rounded-full bg-base-200 px-1.5 py-0.5 text-xs tabular-nums">{{ formatLocalizedNumber(statusCount(filter)) }}</span>
              </button>
            </div>
            <button v-if="query || status !== 'All'" class="btn btn-ghost btn-sm" type="button" @click="clearFilters">{{ $t("Clear") }}</button>
          </div>
        </div>

        <div class="loan-register-table-head hidden gap-3 border-b border-base-300 px-4 py-3 text-xs font-medium text-base-content/55 md:grid"><span>{{ $t("Loan") }}</span><span>{{ $t("Counterparty") }}</span><span>{{ $t("Remaining") }}</span><span>{{ $t("Status") }}</span><span></span></div>
        <LoadingState v-if="pageLoading" :label='$t("Loading loans…")' />
        <div v-else-if="pagedLoans.length" class="min-h-0 flex-1 divide-y divide-base-300 overflow-y-auto">
          <button v-for="loan in pagedLoans" :key="loan.id" class="loan-register-row group grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200/60 focus-visible:bg-base-200/60 focus-visible:outline focus-visible:outline-1 focus-visible:outline-primary md:grid" :class="selectedLoanId === loan.id ? 'bg-primary/10' : ''" type="button" @click="selectLoan(loan.id)">
            <span class="flex min-w-0 items-center gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box border border-base-300 bg-base-200 text-primary"><CircleDollarSign :size="18" aria-hidden="true" /></span><span class="min-w-0"><strong class="block truncate text-sm">{{ loan.loanNumber }}</strong><span class="block truncate text-xs text-base-content/60">{{ $ui(loan.direction === 'payable' ? 'Payable / borrowed' : 'Receivable / lent') }} · {{ formatDate(loan.startDate) }}</span></span></span>
            <span class="hidden min-w-0 truncate text-xs text-base-content/70 md:block">{{ loan.counterpartyName }}</span>
            <span class="hidden text-sm font-semibold tabular-nums text-primary md:block">{{ formatMoney(loan.remainingPrincipalRial + loan.remainingInterestRial, currencyUnit) }}</span>
            <StatusBadge class="justify-self-end md:justify-self-start" :label="loan.overdueRial > 0 ? $t('Overdue') : loan.status" :tone="loan.overdueRial > 0 ? 'red' : tone(loan.status)" />
            <ChevronRight :size="17" class="register-row-arrow justify-self-end text-base-content/45" aria-hidden="true" />
            <span class="col-span-3 flex flex-wrap gap-x-3 gap-y-1 text-xs text-base-content/55 md:hidden"><span>{{ loan.counterpartyName }}</span><span>{{ formatMoney(loan.remainingPrincipalRial + loan.remainingInterestRial, currencyUnit) }}</span><span>{{ formatLocalizedNumber(loan.installments.length) }} {{ $t("installments") }}</span></span>
          </button>
        </div>
        <EmptyState v-else :title="$ui(loans.length ? 'No loans match this view' : 'No loans yet')" :description="$ui(loans.length ? 'Try another filter or search term.' : 'Open the first payable or receivable loan.')"><template #icon><Search :size="22" /></template><template #action><button v-if="loans.length" class="btn btn-outline btn-sm" type="button" @click="clearFilters">{{ $t("Clear filters") }}</button><button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="emit('new-loan')"><Plus :size="15" />{{ $t("Create loan") }}</button></template></EmptyState>
        <footer class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-base-300 bg-base-100 px-4 py-3 text-xs text-base-content/60"><span>{{ $ui(pageSummary) }}</span><div class="flex items-center gap-1"><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === 1" :aria-label='$t("Previous page")' @click="goToPage(page - 1)"><ChevronLeft :size="15" /></button><button v-for="number in pageNumbers" :key="number" class="btn btn-xs min-w-8" :class="page === number ? 'btn-primary' : 'btn-ghost'" type="button" @click="goToPage(number)">{{ formatLocalizedNumber(number) }}</button><button class="btn btn-ghost btn-xs btn-square" type="button" :disabled="page === pageCount" :aria-label='$t("Next page")' @click="goToPage(page + 1)"><ChevronRight :size="15" /></button></div></footer>
      </section>

      <LoanDetailPanel v-if="selectedLoan" :loan="selectedLoan" :currency-unit="currencyUnit" :busy="busy" @edit="emit('edit-loan', selectedLoan.id)" @manage-payments="emit('manage-payments', selectedLoan.id, false)" @remove="removeSelectedLoan" />
      <section v-else class="flex min-h-72 min-w-0 items-center justify-center rounded-box border border-dashed border-base-300 p-8 text-center"><EmptyState :title='$t("Select a loan")' :description='$t("Choose a loan from the register to inspect its balance, schedule, and payments.")'><template #icon><CircleDollarSign :size="22" /></template></EmptyState></section>
    </div>
  </div>
</template>

<style scoped>
.loan-register-table-head,
.loan-register-row {
  grid-template-columns: minmax(0, 1.2fr) minmax(8rem, 1fr) 8.5rem 7rem 1.25rem;
}

@media (max-width: 767px) {
  .loan-register-row {
    grid-template-columns: minmax(0, 1fr) auto auto;
  }
}
</style>
