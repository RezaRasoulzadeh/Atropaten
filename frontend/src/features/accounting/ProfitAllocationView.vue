<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CalendarDays, CheckCircle2, Plus, RefreshCw, Users } from 'lucide-vue-next'
import AppPanel from '../../components/layout/AppPanel.vue'
import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import DataTableRow from '../../components/ui/DataTableRow.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import FormField from '../../components/ui/FormField.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import AppInput from '../../components/ui/AppInput.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { ownersApi, type FiscalPeriodRecord, type OwnerRecord } from '../../api/owners'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'
import { confirmAction } from '../../ui/feedback'

const props = defineProps<{ currencyUnit: CurrencyUnit }>()
const emit = defineEmits<{ notify: [string] }>()
const { busy, pageLoading, runAction, runLoad } = useWorkspaceActions()

const periods = ref<FiscalPeriodRecord[]>([])
const owners = ref<OwnerRecord[]>([])
const selectedId = ref<string | null>(null)
const form = ref({ name: '', startDate: null as string | null, endDate: null as string | null })
const selected = computed(() => periods.value.find((value) => value.id === selectedId.value) ?? null)

function normalizePeriod(period: FiscalPeriodRecord): FiscalPeriodRecord {
  return {
    ...period,
    allocations: Array.isArray(period.allocations) ? period.allocations : [],
    previewAllocations: Array.isArray(period.previewAllocations) ? period.previewAllocations : [],
  }
}

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}

function ownerName(id: string) {
  return owners.value.find((owner) => owner.id === id)?.name ?? id
}

function periodTone(status: string) {
  return status === 'Closed' ? 'slate' : 'amber'
}

async function load() {
  return runLoad(async () => {
    try {
      const [periodRows, ownerRows] = await Promise.all([ownersApi.periods(), ownersApi.list(true)])
      periods.value = periodRows.map(normalizePeriod)
      owners.value = ownerRows
      if (!selectedId.value) selectedId.value = periods.value[0]?.id ?? null
    } catch (error) {
      reportError(error)
    }
  })
}

onMounted(load)

async function createPeriod() {
  return runAction(async () => {
    try {
      if (!form.value.name.trim() || !form.value.startDate || !form.value.endDate) {
        throw new Error('Period name, start date, and end date are required')
      }
      const value = await ownersApi.createPeriod(form.value)
      periods.value = [normalizePeriod(value), ...periods.value]
      selectedId.value = value.id
      form.value = { name: '', startDate: null, endDate: null }
      emit('notify', 'Fiscal period created.')
    } catch (error) {
      reportError(error)
    }
  })
}

async function preview() {
  return runAction(async () => {
    if (!selected.value) return
    try {
      const value = await ownersApi.previewPeriod(selected.value.id)
      periods.value = periods.value.map((period) => (period.id === value.id ? normalizePeriod(value) : period))
      emit('notify', 'Profit allocation preview refreshed.')
    } catch (error) {
      reportError(error)
    }
  })
}

async function close() {
  return runAction(async () => {
    if (
      !selected.value ||
      !(await confirmAction({
        title: 'Close fiscal period',
        message: 'Close this fiscal period permanently?',
        confirmLabel: 'Close period',
        danger: true,
      }))
    )
      return
    try {
      const value = await ownersApi.closePeriod(selected.value.id)
      periods.value = periods.value.map((period) => (period.id === value.id ? normalizePeriod(value) : period))
      emit('notify', 'Fiscal period closed.')
    } catch (error) {
      reportError(error)
    }
  })
}
</script>

<template>
  <div class="min-w-0 space-y-4">
    <LoadingState v-if="pageLoading" label="Loading fiscal periods…" />
    <template v-else>
      <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <AppPanel title="Fiscal periods" subtitle="Close once; the profit-sharing snapshot is preserved forever." :flush="true">
          <DataTable v-if="periods.length" label="Fiscal periods">
            <thead><tr><th scope="col">Period</th><th scope="col">Dates</th><th scope="col">Status</th><th scope="col" class="text-end">Profit / loss</th></tr></thead>
            <tbody>
              <DataTableRow v-for="period in periods" :key="period.id" interactive :selected="selectedId === period.id" @activate="selectedId = period.id">
                <DataTableCell><strong>{{ period.name }}</strong><span class="mt-1 block text-xs text-base-content/55">{{ period.allocations.length }} allocations</span></DataTableCell>
                <DataTableCell><span class="block whitespace-nowrap">{{ formatDateTime(period.startDate) }}</span><span class="mt-1 block whitespace-nowrap text-xs text-base-content/55">to {{ formatDateTime(period.endDate) }}</span></DataTableCell>
                <DataTableCell><StatusBadge :label="period.status" :tone="periodTone(period.status)" /></DataTableCell>
                <DataTableCell numeric><strong class="text-primary">{{ money(period.profitLossRial) }}</strong></DataTableCell>
              </DataTableRow>
            </tbody>
          </DataTable>
          <EmptyState v-else title="No fiscal periods" description="Create a period to preview and close profit allocation."><template #icon><CalendarDays :size="22" aria-hidden="true" /></template></EmptyState>
        </AppPanel>

        <AppPanel title="New fiscal period" subtitle="Use the ledger snapshot to calculate owner allocations.">
          <form class="min-w-0 space-y-4" @submit.prevent="createPeriod">
            <FormField label="Period name"><AppInput v-model="form.name" class="input w-full min-w-0" required placeholder="1405" /></FormField>
            <FormGrid>
              <FormField label="Start date"><JalaliDatePicker v-model="form.startDate" /></FormField>
              <FormField label="End date"><JalaliDatePicker v-model="form.endDate" /></FormField>
            </FormGrid>
            <button class="btn btn-primary w-full" type="submit" :disabled="busy"><Plus :size="15" /> Create period</button>
          </form>
        </AppPanel>
      </div>

      <AppPanel v-if="selected" title="Closing preview" subtitle="Revenue − COGS − expenses, calculated from journal lines only.">
        <template #action>
          <StatusBadge :label="selected.status" :tone="periodTone(selected.status)" />
        </template>
        <div class="grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-box border border-base-300 bg-base-200/30 p-3"><span class="block text-xs text-base-content/55">Revenue</span><strong class="mt-1 block tabular-nums">{{ money(selected.revenueRial) }}</strong></div>
          <div class="rounded-box border border-base-300 bg-base-200/30 p-3"><span class="block text-xs text-base-content/55">COGS</span><strong class="mt-1 block tabular-nums">{{ money(selected.cogsRial) }}</strong></div>
          <div class="rounded-box border border-base-300 bg-base-200/30 p-3"><span class="block text-xs text-base-content/55">Expenses</span><strong class="mt-1 block tabular-nums">{{ money(selected.expensesRial) }}</strong></div>
          <div class="rounded-box border border-primary/20 bg-primary/5 p-3"><span class="block text-xs text-primary/70">Profit / loss</span><strong class="mt-1 block tabular-nums text-primary">{{ money(selected.profitLossRial) }}</strong></div>
        </div>
        <div class="mt-4 flex flex-wrap justify-end gap-2">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="preview"><RefreshCw :size="14" /> Refresh preview</button>
          <button v-if="selected.status === 'Open'" class="btn btn-primary btn-sm" type="button" :disabled="busy" @click="close"><CheckCircle2 :size="14" /> Close period</button>
        </div>
        <div class="mt-5 border-t border-base-300 pt-4">
          <h3 class="text-sm font-semibold">Owner allocation</h3>
          <div v-if="(selected.previewAllocations?.length ? selected.previewAllocations : selected.allocations).length" class="mt-2 divide-y divide-base-300 rounded-box border border-base-300 bg-base-200/20">
            <div v-for="allocation in (selected.previewAllocations?.length ? selected.previewAllocations : selected.allocations)" :key="allocation.ownerId" class="flex min-w-0 flex-wrap items-center justify-between gap-3 px-3 py-2.5 text-sm">
              <span class="min-w-0 truncate">{{ ownerName(allocation.ownerId) }} <span class="text-xs text-base-content/55">· {{ allocation.profitSharingBps / 100 }}%</span></span>
              <strong class="shrink-0 tabular-nums">{{ money(allocation.amountRial) }}</strong>
            </div>
          </div>
          <EmptyState v-else title="No owner allocations" description="Add active owners with profit shares to calculate allocation."><template #icon><Users :size="21" aria-hidden="true" /></template></EmptyState>
        </div>
      </AppPanel>
      <EmptyState v-else title="Select a fiscal period" description="Choose a period above to preview its profit allocation."><template #icon><CalendarDays :size="22" aria-hidden="true" /></template></EmptyState>
    </template>
  </div>
</template>
