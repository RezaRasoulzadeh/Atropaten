<script setup lang="ts">
import { computed } from 'vue';
import { BarChart3, ListChecks } from 'lucide-vue-next';
import AppPanel from '../../components/layout/AppPanel.vue';
import DataTable from '../../components/ui/DataTable.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import type { DashboardRecord } from '../../api/reports';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { formatDate } from '../../utils/date';
const props = defineProps<{
  data: DashboardRecord;
  currencyUnit: CurrencyUnit;
}>();
const trend = computed(() => props.data.trend ?? []);
const pipeline = computed(() => props.data.pipeline ?? []);
const total = computed(() =>
  pipeline.value.reduce((sum, item) => sum + item.count, 0),
);
const maxCount = computed(() =>
  Math.max(1, ...pipeline.value.map((item) => item.count)),
);
const bounds = computed(() => {
  const values = trend.value.flatMap((item) => [
    item.revenueRial,
    item.grossProfitRial,
  ]);
  const min = Math.min(0, ...values);
  const max = Math.max(0, ...values);
  return { min, max: max === min ? min + 1 : max };
});
const hasActivity = computed(() =>
  trend.value.some(
    (item) => item.revenueRial !== 0 || item.grossProfitRial !== 0,
  ),
);
const x = (index: number) =>
  20 + (index * 560) / Math.max(1, trend.value.length - 1);
const y = (value: number) =>
  160 -
  ((value - bounds.value.min) / (bounds.value.max - bounds.value.min)) * 140;
const points = (key: 'revenueRial' | 'grossProfitRial') =>
  trend.value.map((item, index) => `${x(index)},${y(item[key])}`).join(' ');
const money = (value: number) => formatMoney(value, props.currencyUnit);
</script>
<template>
  <section
    class="grid min-w-0 items-start gap-4 xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]"
  >
    <AppPanel
      title="Sales & gross profit"
      subtitle="Daily totals for the selected period"
    >
      <div class="flex flex-wrap gap-4 text-xs">
        <span class="flex items-center gap-2"
          ><span class="h-0.5 w-5 bg-info"></span> Sales</span
        >
        <span class="flex items-center gap-2"
          ><span class="w-5 border-t-2 border-dashed border-success"></span>
          Gross profit</span
        >
      </div>
      <template v-if="hasActivity">
        <div class="flex justify-between gap-2 text-xs text-base-content/60">
          <span>{{ money(bounds.max) }}</span
          ><span>Daily {{ currencyUnit }}</span>
        </div>
        <svg
          viewBox="0 0 600 180"
          class="block w-full"
          role="img"
          aria-label="Daily sales and gross profit line chart. Exact values are available in the data table below."
        >
          <line
            x1="20"
            :y1="y(0)"
            x2="580"
            :y2="y(0)"
            stroke="var(--color-base-content)"
            stroke-opacity="0.25"
          />
          <polyline
            :points="points('revenueRial')"
            fill="none"
            stroke="var(--color-info)"
            stroke-width="2.5"
          />
          <polyline
            :points="points('grossProfitRial')"
            fill="none"
            stroke="var(--color-success)"
            stroke-width="2.5"
            stroke-dasharray="6 4"
          />
          <g v-for="(item, index) in trend" :key="item.date">
            <circle
              :cx="x(index)"
              :cy="y(item.revenueRial)"
              r="2.5"
              fill="var(--color-info)"
            >
              <title>
                {{ formatDate(item.date) }} · Sales
                {{ money(item.revenueRial) }} · Gross profit
                {{ money(item.grossProfitRial) }}
              </title>
            </circle>
          </g>
        </svg>
        <div class="flex justify-between gap-2 text-xs text-base-content/60">
          <span>{{ money(bounds.min) }}</span>
        </div>
        <div class="flex justify-between gap-2 text-xs text-base-content/60">
          <span>{{ formatDate(data.startDate) }}</span
          ><span>{{ formatDate(data.endDate) }}</span>
        </div>
      </template>
      <EmptyState
        v-else
        compact
        title="No activity in this period"
        description="Sales and gross profit will appear here once the selected period has activity."
      >
        <template #icon><BarChart3 :size="21" aria-hidden="true" /></template>
      </EmptyState>
      <details v-if="hasActivity" class="text-xs">
        <summary class="cursor-pointer py-1 text-primary">
          View chart data
        </summary>
        <DataTable label="Daily sales and gross profit"
          ><thead>
            <tr>
              <th>Date</th>
              <th class="text-end">Sales</th>
              <th class="text-end">Gross profit</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in trend" :key="item.date">
              <td>{{ formatDate(item.date) }}</td>
              <td class="text-end">{{ money(item.revenueRial) }}</td>
              <td class="text-end">{{ money(item.grossProfitRial) }}</td>
            </tr>
          </tbody></DataTable
        >
      </details>
    </AppPanel>
    <AppPanel
      title="Order pipeline"
      :subtitle="`${total} active orders · Current status`"
    >
      <div v-if="pipeline.length" class="space-y-5 py-3">
        <div v-for="item in pipeline" :key="item.status">
          <div class="mb-2 flex justify-between gap-2 text-xs">
            <span>{{ item.status }}</span
            ><strong>{{ item.count }}</strong>
          </div>
          <div
            class="h-3 overflow-hidden rounded bg-base-300"
            role="img"
            :aria-label="`${item.status}: ${item.count} orders`"
          >
            <div
              class="h-full rounded bg-info"
              :style="{ width: `${(item.count / maxCount) * 100}%` }"
            ></div>
          </div>
        </div>
      </div>
      <EmptyState
        v-else
        compact
        title="No active orders"
        description="Current order status distribution will appear here when orders are active."
      >
        <template #icon><ListChecks :size="21" aria-hidden="true" /></template>
      </EmptyState>
    </AppPanel>
  </section>
</template>
