<script setup lang="ts">
import PrintPreviewPanel from './PrintPreviewPanel.vue'


import FormGrid from '../../components/ui/FormGrid.vue';
import InlineAlert from '../../components/ui/InlineAlert.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import AppInput from '../../components/ui/AppInput.vue';
import FormField from '../../components/ui/FormField.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTable from '../../components/ui/DataTable.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref, watch } from 'vue';
import { BarChart3, FileText, Printer, RefreshCw } from 'lucide-vue-next';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import { reportsApi, type PrintDocumentRecord, type ReportRecord } from '../../api/reports';
import { formatMoney } from '../../utils/currency';
import { formatDate, formatDateTime, currentCanonicalDate } from '../../utils/date';
import { normalizeError } from '../../ui/feedback';
import SelectField from '../../components/ui/SelectField.vue';
const props = defineProps<{ currencyUnit: 'Rial' | 'Toman' }>();
const emit = defineEmits<{ notify: [message: string] }>();
const tabs = [
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
];
const activeTab = ref('profit_loss');
const end = ref<string | null>(currentCanonicalDate());
const start = ref<string | null>(new Date(Date.now() - 30 * 86400000).toISOString().slice(0, 10));
const report = ref<ReportRecord | null>(null);
const loading = ref(false);
const error = ref('');
const summary = computed(() => report.value?.summaries ?? []);
const money = (v: number) => formatMoney(v, props.currencyUnit);
const quantity = (v: number) => (v ? String(v / 1000000) : '0');
async function load() {
  loading.value = true;
  error.value = '';
  try {
    report.value = await reportsApi.report(activeTab.value, start.value || '', end.value || '');
  } catch (e) {
    error.value = normalizeError(e).message;
  } finally {
    loading.value = false;
  }
}
onMounted(load);
watch([activeTab, start, end], load);
</script>
<template>
  <div class="min-w-0 space-y-3">
    <WorkspaceStickyStack
      ><WorkspaceHeader
        title="Reports"
        eyebrow="Insights · authoritative queries"
        description="Reconciled views of journals, movements, production, and saved documents."
        ><div class="flex flex-wrap items-center gap-2">
          <button class="btn btn-ghost" type="button" @click="load">
            <RefreshCw :size="15" /> Refresh
          </button>
        </div></WorkspaceHeader
      ></WorkspaceStickyStack
    >
    <div class="min-w-0 space-y-3">
      <div role="tablist" class="flex flex-wrap items-center gap-2">
        <button
          class="btn btn-sm"
          v-for="tab in tabs"
          :key="tab[0]"
          :class="activeTab === tab[0] ? 'btn-primary' : 'btn-ghost'"
          type="button"
          @click="activeTab = tab[0]"
        >
          <BarChart3 :size="14" />{{ tab[1] }}
        </button>
      </div>
      <FormGrid
        ><FormField class="gap-1">From <JalaliDatePicker v-model="start" /></FormField
        ><FormField class="gap-1">To <JalaliDatePicker v-model="end" /></FormField
      ></FormGrid>
    </div>
    <InlineAlert v-if="error" tone="error">{{ error }}</InlineAlert>
    <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-5"><AppPanel v-for="item in summary" :key="item.key"><p class="text-xs text-base-content/60">{{item.label}}</p><strong class="block text-lg tabular-nums">{{money(item.amountRial)}}</strong><small v-if="item.count" class="text-xs text-base-content/50">{{item.count}} records</small></AppPanel></div>
    <AppPanel
      :title="tabs.find((t) => t[0] === activeTab)?.[1] || 'Report'"
      :subtitle="`${report?.startDate || ''} → ${report?.endDate || ''} · values supplied by Go`"
      ><div>
        <DataTable
          ><thead>
            <tr>
              <th>Source / name</th>
              <th>Category</th>
              <th>Status</th>
              <th class="text-end">Quantity</th>
              <th class="text-end">Amount</th>
              <th class="text-end">Secondary</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in report?.rows" :key="`${row.id}-${row.name}`">
              <DataTableCell
                ><span class="block font-medium">{{ row.name || row.id }}</span
                ><span v-if="row.referenceId" class="block text-xs text-base-content/60">{{
                  row.referenceId
                }}</span></DataTableCell
              ><DataTableCell>{{ row.category || '—' }}</DataTableCell
              ><DataTableCell
                ><StatusBadge v-if="row.status" :label="row.status" tone="slate" /><span
                  v-else
                  class="block text-xs text-base-content/60"
                  >—</span
                ></DataTableCell
              ><DataTableCell numeric
                >{{ row.quantityUnits ? quantity(row.quantityUnits) : '—'
                }}<span v-if="row.secondaryQuantityUnits">
                  / {{ quantity(row.secondaryQuantityUnits) }}</span
                ></DataTableCell
              ><DataTableCell numeric>{{ money(row.amountRial) }}</DataTableCell
              ><DataTableCell numeric>{{
                row.secondaryAmountRial ? money(row.secondaryAmountRial) : '—'
              }}</DataTableCell>
            </tr>
            <tr v-if="!loading && !report?.rows?.length">
              <DataTableCell colspan="6">No persisted records in this range.</DataTableCell>
            </tr>
          </tbody></DataTable
        >
      </div></AppPanel
    >
    <PrintPreviewPanel :currency-unit="currencyUnit" :start="start" :end="end" />
  </div>
</template>
