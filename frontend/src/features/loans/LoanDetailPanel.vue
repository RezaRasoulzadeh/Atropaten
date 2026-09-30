<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { CalendarDays, CircleDollarSign, CreditCard, Edit3, Trash2 } from 'lucide-vue-next'
import EmptyState from '../../components/ui/EmptyState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { LoanRecord } from '../../api/loans'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { formatDate } from '../../utils/date'
import { formatLocalizedNumber } from '../../utils/number'

type DetailTab = 'overview' | 'schedule'
type Tone = 'blue' | 'green' | 'amber' | 'red' | 'slate'

const props = defineProps<{ loan: LoanRecord; currencyUnit: CurrencyUnit; busy?: boolean }>()
const emit = defineEmits<{
  edit: []
  remove: []
  'manage-payments': []
}>()
const activeTab = ref<DetailTab>('overview')

const totalRial = computed(() => props.loan.principalRial + props.loan.interestFeeRial)
const paidRial = computed(() => props.loan.paidPrincipalRial + props.loan.paidInterestRial)
const remainingRial = computed(() => props.loan.remainingPrincipalRial + props.loan.remainingInterestRial)
const progress = computed(() => totalRial.value > 0 ? Math.min(100, Math.round((paidRial.value / totalRial.value) * 100)) : 0)

function tone(value: string): Tone {
  if (value.toLowerCase() === 'closed' || value.toLowerCase() === 'paid') return 'green'
  if (value.toLowerCase() === 'overdue') return 'red'
  if (value.toLowerCase() === 'active' || value.toLowerCase() === 'posted') return 'blue'
  return 'slate'
}

watch(() => props.loan.id, () => { activeTab.value = 'overview' })
</script>

<template>
  <section class="h-auto min-h-0 min-w-0 overflow-visible rounded-box border border-base-300 bg-base-100 xl:h-full xl:overflow-y-auto" :aria-label='$t("Loan details")'>
    <div class="border-b border-base-300 p-3 sm:p-5">
      <div class="relative min-h-52 overflow-hidden rounded-box bg-base-300 sm:min-h-60">
        <div class="absolute inset-0 bg-gradient-to-br from-primary/20 via-base-300 to-base-200" aria-hidden="true"></div>
        <div class="absolute -end-10 -top-12 size-48 rounded-full bg-primary/10 blur-2xl" aria-hidden="true"></div>
        <div class="relative z-10 flex min-h-52 flex-col justify-between p-4 sm:min-h-60 sm:p-5">
          <div class="grid size-11 place-items-center rounded-box border border-primary/30 bg-base-100/70 text-primary"><CircleDollarSign :size="24" aria-hidden="true" /></div>
          <div class="min-w-0">
            <div class="flex min-w-0 items-center gap-3"><h2 class="min-w-0 truncate text-xl font-semibold sm:text-2xl">{{ loan.loanNumber }}</h2><StatusBadge class="shrink-0" :label="loan.status" :tone="tone(loan.status)" /></div>
            <p class="mt-2 truncate text-sm text-base-content/65">{{ loan.counterpartyName }}</p>
            <p class="mt-1 text-xs text-base-content/50">{{ $ui(loan.direction === 'payable' ? 'Payable / borrowed' : 'Receivable / lent') }} · {{ formatDate(loan.startDate) }}</p>
          </div>
        </div>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
        <button class="btn btn-outline btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('edit')"><Edit3 :size="14" />{{ $t("Edit loan") }}</button>
        <button class="btn btn-primary btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('manage-payments')"><CreditCard :size="14" />{{ $t("Manage payments") }}</button>
        <button class="btn btn-outline btn-error btn-sm w-full gap-2" type="button" :disabled="busy" @click="emit('remove')"><Trash2 :size="14" />{{ $t("Remove loan") }}</button>
      </div>

      <div class="mt-4 grid min-w-0 divide-y divide-base-300 border-y border-base-300 sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:first:ps-0"><CircleDollarSign :size="20" class="shrink-0 text-primary" /><div class="min-w-0"><span class="block text-xs text-base-content/55">{{ $t("Original total") }}</span><strong class="block truncate text-sm tabular-nums">{{ formatMoney(totalRial, currencyUnit) }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3"><CreditCard :size="20" class="shrink-0 text-success" /><div class="min-w-0"><span class="block text-xs text-base-content/55">{{ $t("Paid") }}</span><strong class="block truncate text-sm tabular-nums">{{ formatMoney(paidRial, currencyUnit) }}</strong></div></div>
        <div class="flex items-center gap-3 py-3 sm:px-3 sm:last:pe-0"><CalendarDays :size="20" class="shrink-0 text-warning" /><div class="min-w-0"><span class="block text-xs text-base-content/55">{{ $t("Remaining") }}</span><strong class="block truncate text-sm tabular-nums">{{ formatMoney(remainingRial, currencyUnit) }}</strong></div></div>
      </div>
    </div>

    <nav class="flex min-w-0 overflow-x-auto border-b border-base-300 px-2" :aria-label='$t("Loan detail tabs")'>
      <button v-for="tab in (['overview', 'schedule'] as DetailTab[])" :key="tab" class="shrink-0 border-b-2 px-3 py-3 text-sm capitalize transition-colors" :class="activeTab === tab ? 'border-primary text-primary' : 'border-transparent text-base-content/65 hover:text-base-content'" type="button" @click="activeTab = tab">{{ $ui(tab) }}<span v-if="tab === 'schedule'" class="ms-1 text-xs text-base-content/50">{{ formatLocalizedNumber(loan.installments.length) }}</span></button>
    </nav>

    <div class="min-w-0 p-3 sm:p-4">
      <div v-if="activeTab === 'overview'" class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4">
          <div class="flex items-center justify-between gap-3"><h3 class="text-sm font-semibold">{{ $t("Repayment progress") }}</h3><strong class="text-sm tabular-nums text-primary">{{ formatLocalizedNumber(progress) }}%</strong></div>
          <progress class="progress progress-primary mt-3 h-2 w-full" :value="progress" max="100"></progress>
          <dl class="mt-4 divide-y divide-base-300/70 text-sm">
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Principal remaining") }}</dt><dd class="tabular-nums">{{ formatMoney(loan.remainingPrincipalRial, currencyUnit) }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Interest remaining") }}</dt><dd class="tabular-nums">{{ formatMoney(loan.remainingInterestRial, currencyUnit) }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Overdue") }}</dt><dd class="tabular-nums" :class="loan.overdueRial ? 'text-error' : ''">{{ formatMoney(loan.overdueRial, currencyUnit) }}</dd></div>
          </dl>
        </div>
        <div class="rounded-box border border-base-300 bg-base-200/20 p-4">
          <h3 class="text-sm font-semibold">{{ $t("Loan information") }}</h3>
          <dl class="mt-3 divide-y divide-base-300/70 text-sm">
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Counterparty") }}</dt><dd class="truncate text-end">{{ loan.counterpartyName }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Start date") }}</dt><dd>{{ formatDate(loan.startDate) }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("End date") }}</dt><dd>{{ loan.endDate ? formatDate(loan.endDate) : $t('Open term') }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">{{ $t("Opening account") }}</dt><dd class="truncate text-end">{{ loan.financialAccountId }}</dd></div>
          </dl>
        </div>
        <div v-if="loan.notes" class="rounded-box border border-base-300 bg-base-200/20 p-4"><h3 class="text-sm font-semibold">{{ $t("Notes") }}</h3><p class="mt-2 whitespace-pre-wrap text-sm leading-6 text-base-content/70">{{ loan.notes }}</p></div>
      </div>

      <div v-else class="space-y-2">
        <div v-for="installment in loan.installments" :key="installment.id" class="rounded-box border border-base-300 bg-base-200/20 p-3">
          <div class="flex items-start justify-between gap-3"><div><strong class="text-sm">{{ $t("Installment") }} #{{ formatLocalizedNumber(installment.position + 1) }}</strong><p class="mt-1 text-xs text-base-content/55">{{ formatDate(installment.dueDate) }}</p></div><StatusBadge :label="installment.status" :tone="tone(installment.status)" /></div>
          <dl class="mt-3 grid grid-cols-3 gap-2 border-t border-base-300 pt-3 text-xs"><div><dt class="text-base-content/50">{{ $t("Due") }}</dt><dd class="mt-1 tabular-nums">{{ formatMoney(installment.totalDueRial, currencyUnit) }}</dd></div><div><dt class="text-base-content/50">{{ $t("Paid") }}</dt><dd class="mt-1 tabular-nums text-success">{{ formatMoney(installment.paidRial, currencyUnit) }}</dd></div><div><dt class="text-base-content/50">{{ $t("Remaining") }}</dt><dd class="mt-1 tabular-nums text-warning">{{ formatMoney(installment.remainingRial, currencyUnit) }}</dd></div></dl>
        </div>
        <EmptyState v-if="!loan.installments.length" :title='$t("No installments")' :description='$t("This loan has no generated schedule.")' />
      </div>
    </div>
  </section>
</template>
