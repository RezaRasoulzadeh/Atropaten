<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { BarChart3, Printer, RefreshCw } from 'lucide-vue-next'
import AppPanel from '../../components/layout/AppPanel.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import FormField from '../../components/ui/FormField.vue'
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import ReportPrintDocument from './ReportPrintDocument.vue'
import { reportsApi, type ReportRecord } from '../../api/reports'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { currentCanonicalDate, formatDateTime } from '../../utils/date'
import { normalizeError, useToast } from '../../ui/feedback'

const props = defineProps<{ currencyUnit: CurrencyUnit }>()

const reportTabs = [
  ['profit_loss', 'P&L'],
  ['cash_bank', 'Cash / Bank'],
  ['receivables', 'Receivables'],
  ['payables', 'Payables'],
  ['expenses', 'Expenses'],
  ['inventory', 'Inventory'],
  ['sales_by_service', 'Sales by service'],
  ['customer_sales', 'Customer sales'],
  ['material_usage', 'Material use / waste'],
  ['production', 'Production'],
] as const

const activeTab = ref('profit_loss')
const end = ref<string | null>(currentCanonicalDate())
const start = ref<string | null>(new Date(Date.now() - 30 * 86400000).toISOString().slice(0, 10))
const report = ref<ReportRecord | null>(null)
const loading = ref(false)
const toast = useToast()
const summary = computed(() => report.value?.summaries ?? [])
const activeLabel = computed(
  () => reportTabs.find(([value]) => value === activeTab.value)?.[1] ?? 'Report',
)
const rowCount = computed(() => report.value?.rows?.length ?? 0)

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}

function quantity(value: number) {
  return value ? String(value / 1000000) : '0'
}

async function load() {
  loading.value = true
  try {
    const value = await reportsApi.report(activeTab.value, start.value || '', end.value || '')
    report.value = {
      ...value,
      summaries: Array.isArray(value.summaries) ? value.summaries : [],
      rows: Array.isArray(value.rows) ? value.rows : [],
    }
  } catch (error) {
    toast.error(normalizeError(error).message, 'Reports')
  } finally {
    loading.value = false
  }
}

function printReport() {
  if (!report.value || loading.value) return
  window.print()
}

onMounted(load)
watch([activeTab, start, end], load)
</script>

<template>
  <div class="min-w-0 space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        :show-breadcrumb="true"
        eyebrow="Insights / authoritative queries"
        title="Reports"
        description="Reconciled views of journals, movements, production, and saved documents."
      >
        <button class="btn btn-ghost btn-sm" type="button" :disabled="loading" @click="load">
          <RefreshCw :size="15" /> Refresh
        </button>
        <button class="btn btn-primary btn-sm" type="button" :disabled="!report || loading" @click="printReport">
          <Printer :size="15" /> Print / save PDF
        </button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>

    <SearchFilterBar>
      <template #search>
        <div class="flex h-10 min-w-0 items-center gap-2 px-1">
          <BarChart3 :size="17" class="shrink-0 text-primary" aria-hidden="true" />
          <div class="min-w-0">
            <strong class="block truncate text-sm">{{ activeLabel }}</strong>
            <span class="block truncate text-xs text-base-content/55">Choose a reporting period; the report updates automatically.</span>
          </div>
        </div>
      </template>
      <template #filters>
        <SelectField
          v-model="activeTab"
          label="Report"
          :options="reportTabs.map(([value, label]) => ({ label, value }))"
        />
        <FormField label="From"><JalaliDatePicker v-model="start" /></FormField>
        <FormField label="To"><JalaliDatePicker v-model="end" /></FormField>
      </template>
      <template #count><span>{{ rowCount }} rows</span></template>
    </SearchFilterBar>

    <LoadingState v-if="loading" label="Loading report…" />
    <template v-else>
      <section v-if="summary.length" class="grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-5" aria-label="Report summary">
        <article v-for="item in summary" :key="item.key" class="min-w-0 rounded-box border border-base-300 bg-base-100 p-4">
          <div class="flex items-start justify-between gap-3">
            <p class="min-w-0 truncate text-xs font-medium text-base-content/60">{{ item.label }}</p>
            <span class="grid size-8 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><BarChart3 :size="15" /></span>
          </div>
          <strong class="mt-3 block truncate text-lg tabular-nums">{{ money(item.amountRial) }}</strong>
          <span v-if="item.count" class="mt-1 block text-xs text-base-content/50">{{ item.count }} records</span>
        </article>
      </section>

      <AppPanel :title="activeLabel" :subtitle="`${report?.startDate || ''} → ${report?.endDate || ''} · values supplied by Go`" :flush="true">
        <template #action><span class="text-xs text-base-content/55">{{ rowCount }} rows</span></template>
        <DataTable label="Report data">
          <thead>
            <tr><th scope="col">Source / name</th><th scope="col">Category</th><th scope="col">Date</th><th scope="col">Status</th><th scope="col" class="text-end">Quantity</th><th scope="col" class="text-end">Amount</th><th scope="col" class="text-end">Secondary</th></tr>
          </thead>
          <tbody>
            <tr v-for="row in report?.rows" :key="`${row.id}-${row.name}`">
              <DataTableCell><span class="block max-w-72 truncate font-medium">{{ row.name || row.id }}</span><span v-if="row.secondaryName" class="mt-1 block max-w-72 truncate text-xs text-base-content/55">{{ row.secondaryName }}</span><span v-if="row.referenceId" class="mt-1 block text-xs text-base-content/55">{{ row.referenceId }}</span></DataTableCell>
              <DataTableCell>{{ row.category || '—' }}</DataTableCell>
              <DataTableCell>{{ row.date ? formatDateTime(row.date) : '—' }}</DataTableCell>
              <DataTableCell><StatusBadge v-if="row.status" :label="row.status" tone="slate" /><span v-else class="text-xs text-base-content/55">—</span></DataTableCell>
              <DataTableCell numeric>{{ row.quantityUnits ? quantity(row.quantityUnits) : '—' }}<span v-if="row.secondaryQuantityUnits"> / {{ quantity(row.secondaryQuantityUnits) }}</span></DataTableCell>
              <DataTableCell numeric>{{ money(row.amountRial) }}</DataTableCell>
              <DataTableCell numeric>{{ row.secondaryAmountRial ? money(row.secondaryAmountRial) : '—' }}</DataTableCell>
            </tr>
            <tr v-if="!report?.rows?.length"><DataTableCell colspan="7"><EmptyState compact title="No records in this period" description="Try another reporting range."><template #icon><BarChart3 :size="21" aria-hidden="true" /></template></EmptyState></DataTableCell></tr>
          </tbody>
        </DataTable>
      </AppPanel>
    </template>

    <Teleport to="body">
      <div v-if="report" class="print-output">
        <ReportPrintDocument :report="report" :title="activeLabel" :currency-unit="props.currencyUnit" />
      </div>
    </Teleport>
  </div>
</template>
