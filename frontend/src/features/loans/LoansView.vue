<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CircleDollarSign, Plus } from 'lucide-vue-next'
import AppPanel from '../../components/layout/AppPanel.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import DataTableRow from '../../components/ui/DataTableRow.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SearchField from '../../components/ui/SearchField.vue'
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { loansApi, type LoanRecord } from '../../api/loans'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'

const { pageLoading, runLoad } = useWorkspaceActions()
const props = defineProps<{ currencyUnit: CurrencyUnit }>()
const emit = defineEmits<{
  notify: [string]
  'open-loan': [string]
  'new-loan': []
}>()

const rows = ref<LoanRecord[]>([])
const query = ref('')
const view = ref('All')
const viewOptions = ['Active', 'Payable', 'Receivable', 'Overdue', 'Closed', 'All'].map((value) => ({
  label: value,
  value,
}))

const filtered = computed(() => {
  const search = query.value.trim().toLowerCase()
  return rows.value.filter((value) => {
    const matchesView =
      view.value === 'All' ||
      (view.value === 'Payable' && value.direction === 'payable') ||
      (view.value === 'Receivable' && value.direction === 'receivable') ||
      (view.value === 'Closed' && value.status === 'Closed') ||
      (view.value === 'Overdue' && value.overdueRial > 0) ||
      (view.value === 'Active' && value.status === 'Active' && value.overdueRial === 0)
    const matchesSearch =
      !search ||
      [value.loanNumber, value.counterpartyName, value.financialAccountId]
        .join(' ')
        .toLowerCase()
        .includes(search)
    return matchesView && matchesSearch
  })
})

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

function clearFilters() {
  query.value = ''
  view.value = 'All'
}

async function load() {
  return runLoad(async () => {
    try {
      rows.value = await loansApi.list()
    } catch (error) {
      reportError(error)
    }
  })
}

onMounted(load)
</script>

<template>
  <div class="min-w-0 space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        :show-breadcrumb="true"
        eyebrow="Finance / financing"
        title="Loans"
        description="Track payable and receivable loans, scheduled installments, overdue balances, and reversible payments."
      >
        <button class="btn btn-primary" type="button" @click="emit('new-loan')">
          <Plus :size="16" /> New loan
        </button>
      </WorkspaceHeader>

      <SearchFilterBar>
        <template #search>
          <SearchField v-model="query" label="Search loans" placeholder="Loan, counterparty, or account" />
        </template>
        <template #filters>
          <SelectField v-model="view" label="View" :options="viewOptions" />
        </template>
        <template #count><span>{{ filtered.length }} of {{ rows.length }} loans</span></template>
        <template #actions>
          <button v-if="query || view !== 'All'" class="btn btn-ghost btn-sm" type="button" @click="clearFilters">
            Clear
          </button>
        </template>
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <AppPanel title="Loan register" subtitle="Select a row to open the full loan workspace." :flush="true">
      <template #action><span class="text-xs text-base-content/60">{{ filtered.length }} shown</span></template>
      <LoadingState v-if="pageLoading" label="Loading loans…" />
      <EmptyState
        v-else-if="!filtered.length"
        title="No loans in this view"
        :description="rows.length ? 'Adjust the search or view filter.' : 'Open a loan to start tracking a financing schedule.'"
      >
        <template #icon><CircleDollarSign :size="22" aria-hidden="true" /></template>
        <template #action>
          <button v-if="rows.length" class="btn btn-primary btn-sm" type="button" @click="clearFilters">Clear filters</button>
          <button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="emit('new-loan')"><Plus :size="15" aria-hidden="true" /> Create loan</button>
        </template>
      </EmptyState>
      <DataTable v-else label="Loan register">
        <thead>
          <tr>
            <th scope="col" class="w-[17%]">Loan</th>
            <th scope="col" class="w-[23%]">Counterparty</th>
            <th scope="col" class="w-[14%]">Type</th>
            <th scope="col" class="w-[15%]">Started</th>
            <th scope="col" class="w-[18%] text-end">Remaining</th>
            <th scope="col" class="w-[13%]">Status</th>
          </tr>
        </thead>
        <tbody>
          <DataTableRow v-for="value in filtered" :key="value.id" interactive @activate="emit('open-loan', value.id)">
            <DataTableCell>
              <strong class="block whitespace-nowrap text-sm">{{ value.loanNumber }}</strong>
              <span class="mt-1 block text-xs text-base-content/55">{{ value.installments.length }} installments</span>
            </DataTableCell>
            <DataTableCell>
              <strong class="block max-w-56 truncate text-sm font-medium">{{ value.counterpartyName }}</strong>
              <span class="mt-1 block max-w-56 truncate text-xs text-base-content/55">{{ value.financialAccountId || 'No account' }}</span>
            </DataTableCell>
            <DataTableCell>
              <span class="block text-sm">{{ value.direction === 'payable' ? 'Payable' : 'Receivable' }}</span>
              <span class="mt-1 block text-xs text-base-content/55">{{ formatMoney(value.principalRial, props.currencyUnit) }} principal</span>
            </DataTableCell>
            <DataTableCell>
              <span class="block whitespace-nowrap text-sm">{{ date(value.startDate) }}</span>
              <span class="mt-1 block whitespace-nowrap text-xs text-base-content/55">{{ value.endDate ? `Ends ${date(value.endDate)}` : 'Open term' }}</span>
            </DataTableCell>
            <DataTableCell numeric>
              <strong class="text-sm text-primary">{{ formatMoney(value.remainingPrincipalRial + value.remainingInterestRial, props.currencyUnit) }}</strong>
              <span class="mt-1 block text-xs text-base-content/55">{{ value.overdueRial ? `${formatMoney(value.overdueRial, props.currencyUnit)} overdue` : 'On schedule' }}</span>
            </DataTableCell>
            <DataTableCell><StatusBadge :label="value.status" :tone="tone(value.status)" /></DataTableCell>
          </DataTableRow>
        </tbody>
      </DataTable>
    </AppPanel>
  </div>
</template>
