<script setup lang="ts">
import DashboardCharts from './DashboardCharts.vue';
import DataTable from '../../components/ui/DataTable.vue';
import SelectField from '../../components/ui/SelectField.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import {
  BarChart3,
  ClipboardList,
  Factory,
  FilePlus2,
  HandCoins,
  ReceiptText,
  RefreshCw,
  TrendingUp,
} from 'lucide-vue-next';
import KpiCard from '../../components/ui/KpiCard.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import { reportsApi, type DashboardRecord } from '../../api/reports';
import { formatMoney } from '../../utils/currency';
import { formatLocalizedNumber } from '../../utils/number';
import { formatDate, currentCanonicalDate } from '../../utils/date';
import { normalizeError, useToast } from '../../ui/feedback';
const props = defineProps<{ currencyUnit: 'Rial' | 'Toman' }>();
const emit = defineEmits<{
  navigate: [view: string];
  newOrder: [];
  notify: [message: string];
  openOrder: [id: string];
}>();
const data = ref<DashboardRecord | null>(null);
const loading = ref(false);
const toast = useToast();
const refreshAnimation = ref<'once' | 'infinite' | ''>('');
let refreshTimer: ReturnType<typeof setTimeout> | undefined;
let refreshClearTimer: ReturnType<typeof setTimeout> | undefined;
const period = ref('30');
const end = ref(currentCanonicalDate());
const start = computed(() => {
  const date = new Date(`${end.value}T00:00:00Z`);
  date.setUTCDate(date.getUTCDate() - Number(period.value) + 1);
  return date.toISOString().slice(0, 10);
});
const money = (v: number) => formatMoney(v, props.currencyUnit);
const localizedCount = (v: number) => formatLocalizedNumber(v);
async function load() {
  if (loading.value) return;
  loading.value = true;
  end.value = currentCanonicalDate();
  refreshAnimation.value = 'once';
  if (refreshTimer) clearTimeout(refreshTimer);
  if (refreshClearTimer) clearTimeout(refreshClearTimer);
  refreshTimer = setTimeout(() => {
    if (loading.value) refreshAnimation.value = 'infinite';
  }, 700);
  const startedAt = Date.now();
  try {
    data.value = await reportsApi.dashboard(start.value, end.value);
  } catch (e) {
    toast.error(normalizeError(e).message, 'Dashboard');
  } finally {
    loading.value = false;
    if (refreshTimer) clearTimeout(refreshTimer);
    refreshTimer = undefined;
    const remaining = Math.max(0, 700 - (Date.now() - startedAt));
    refreshClearTimer = setTimeout(() => {
      refreshAnimation.value = '';
    }, remaining);
  }
}
onMounted(load);
watch(period, load);
onBeforeUnmount(() => {
  if (refreshTimer) clearTimeout(refreshTimer);
  if (refreshClearTimer) clearTimeout(refreshClearTimer);
});
const initialLoading = computed(() => loading.value && !data.value);
</script>
<template>
  <div>
    <WorkspaceStickyStack>
      <WorkspaceHeader
        :eyebrow="$ui(`${formatDate(data?.startDate || start)} → ${formatDate(data?.endDate || end)}`)"
        :title='$t("Dashboard")'
        :description='$t("Sales performance and orders to follow up.")'
      >
        <SelectField
          :model-value="period"
          :disabled="loading"
          :options="[
            { label: 'Last 30 days', value: '30' },
            { label: 'Last 90 days', value: '90' },
            { label: 'Last 365 days', value: '365' },
          ]"
          :aria-label='$t("Dashboard period")'
          @update:model-value="period = $event"
        />
        <button
          :disabled="loading"
          class="btn btn-outline btn-primary"
          type="button"
          @click="load"
        >
          <RefreshCw
            :class="{
              'refresh-once': refreshAnimation === 'once',
              'animate-spin': refreshAnimation === 'infinite',
            }"
            :size="15"
          /><span>{{ $t("Refresh") }}</span></button
        ><button
          class="btn btn-primary gap-2"
          type="button"
          @click="emit('newOrder')"
        >
          <FilePlus2 :size="15" :stroke-width="1.8" aria-hidden="true" /><span
            >{{ $t("New order") }}</span
          >
        </button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>
    <div class="mt-4 space-y-4">
      <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <KpiCard
          :value="money(data?.revenueRial || 0)"
          :detail='$t("Confirmed and closed orders, after discounts")'
          :title='$t("Sales")'
          :trend="loading ? 'Loading' : 'Period'"
          :icon="TrendingUp"
          accent="blue"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.grossProfitRial || 0)"
          :detail='$t("Sales less expected order cost")'
          :title='$t("Gross profit")'
          trend="Period"
          :icon="BarChart3"
          accent="green"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.receivableRial || 0)"
          :detail="$ui(`${localizedCount(data?.openInvoiceCount || 0)} open invoice${data?.openInvoiceCount === 1 ? '' : 's'}`)"
          :title='$t("Receivables")'
          trend="Outstanding"
          :icon="HandCoins"
          accent="amber"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.payableRial || 0)"
          :detail='$t("Outstanding supplier balance")'
          :title='$t("Payables")'
          trend="Supplier"
          :icon="ReceiptText"
          accent="red"
          :loading="initialLoading"
        />
      </section>
      <p class="text-xs text-base-content/60">
        {{ $t("Confirmed and closed orders count as sales on their order date. Draft and cancelled orders are excluded. Gross profit uses current expected production costs. Invoices and payments do not count an order again.") }}
      </p>
      <div
        v-if="initialLoading"
        class="grid gap-4 xl:grid-cols-2"
        :aria-label='$t("Loading dashboard charts")'
      >
        <div v-for="n in 2" :key="n" class="skeleton h-72 bg-base-300/70"></div>
      </div>
      <DashboardCharts
        v-else-if="data"
        :data="data"
        :currency-unit="currencyUnit"
      />
      <AppPanel
        :title='$t("Orders needing attention")'
        :subtitle='$t("Unfinished confirmed orders and orders with reference attachments")'
        :flush="true"
      >
        <template #action
          ><button
            class="btn btn-ghost btn-sm"
            @click="emit('navigate', 'Orders')"
          >
            {{ $t("All orders") }}
          </button></template
        >
        <p v-if="initialLoading" class="p-4 text-sm text-base-content/60">
          {{ $t("Loading orders…") }}
        </p>
        <DataTable
          v-else-if="data?.ordersNeedingAttention?.length"
          :label='$t("Orders needing attention")'
        >
          <thead>
            <tr>
              <th>{{ $t("Order / Customer") }}</th>
              <th>{{ $t("Follow-up") }}</th>
              <th>{{ $t("Fulfillment") }}</th>
              <th>{{ $t("Due date") }}</th>
              <th class="text-end">{{ $t("Total") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in data.ordersNeedingAttention" :key="order.id">
              <td>
                <button
                  :aria-label="$ui(`Open order ${order.orderNumber}`)"
                  class="text-start font-semibold text-primary hover:underline"
                  @click="emit('openOrder', order.id)"
                >
                  {{ order.orderNumber }}</button
                ><span
                  class="block max-w-60 truncate text-xs text-base-content/60"
                  :title="order.customer"
                  >{{ $ui(order.customer || 'Walk-in customer') }}</span
                >
              </td>
              <td>
                <div class="flex flex-wrap gap-1">
                  <StatusBadge
                    v-if="order.commercialStatus === 'Confirmed'"
                    :label='$t("Confirmed")'
                    tone="blue"
                  /><StatusBadge
                    v-if="order.referenceCount"
                    :label="$ui(`${order.referenceCount} reference ${order.referenceCount === 1 ? 'file' : 'files'}`)"
                    tone="amber"
                  />
                </div>
              </td>
              <td>
                <StatusBadge :label="order.fulfillmentStatus" tone="slate" />
              </td>
              <td>
                {{
                  $ui(order.dueDate ? formatDate(order.dueDate) : 'Not scheduled')
                }}
              </td>
              <td class="text-end">{{ money(order.totalRial) }}</td>
            </tr>
          </tbody>
        </DataTable>
        <EmptyState
          v-else-if="data"
          compact
          :title='$t("No follow-up orders")'
          :description='$t("Orders needing attention will appear here.")'
        >
          <template #icon><ClipboardList :size="21" aria-hidden="true" /></template>
          <template #action><button class="btn btn-primary btn-sm" type="button" @click="emit('navigate', 'Orders')">{{ $t("View orders") }}</button></template>
        </EmptyState>
        <EmptyState
          v-else
          compact
          :title='$t("Order data unavailable")'
          :description='$t("Refresh the dashboard to try loading the order summary again.")'
        >
          <template #icon><RefreshCw :size="21" aria-hidden="true" /></template>
          <template #action><button class="btn btn-primary btn-sm" type="button" @click="load">{{ $t("Refresh dashboard") }}</button></template>
        </EmptyState>
      </AppPanel>
      <RegisterList
        :title='$t("Production queue")'
        :subtitle='$t("Active production jobs")'
        :count="data?.production?.length || 0"
      >
        <div
          v-if="initialLoading"
          class="space-y-3 p-4"
          :aria-label='$t("Loading production jobs")'
        >
          <div
            v-for="row in 3"
            :key="row"
            class="skeleton h-10 w-full bg-base-300/70"
          ></div>
        </div>
        <div v-else-if="data?.production?.length">
          <RegisterRow
            v-for="job in data.production"
            :key="job.id"
            :interactive="false"
          >
            <template #identity>
              <div
                class="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-start gap-x-3 gap-y-1"
              >
                <div class="flex min-w-0 items-center gap-2">
                  <strong class="shrink-0 truncate text-sm">{{
                    job.orderNumber || job.id
                  }}</strong>
                  <span class="min-w-0 truncate text-xs text-base-content/60">{{
                    $ui(job.customer || 'Walk-in customer')
                  }}</span>
                </div>
                <StatusBadge
                  class="self-start"
                  :label="job.status"
                  tone="blue"
                />
                <span class="min-w-0 truncate text-xs text-base-content/80">{{
                  job.service
                }}</span>
              </div>
            </template>
          </RegisterRow>
        </div>
        <EmptyState
          v-else
          compact
          :title='$t("No active production jobs")'
          :description='$t("Active jobs will appear here once production is scheduled.")'
        >
          <template #icon><Factory :size="21" aria-hidden="true" /></template>
          <template #action><button class="btn btn-primary btn-sm" type="button" @click="emit('navigate', 'Production')">{{ $t("View production") }}</button></template>
        </EmptyState>
      </RegisterList>
    </div>
  </div>
</template>

<style scoped>
@keyframes refresh-spin-once {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

.refresh-once {
  animation: refresh-spin-once 0.7s linear 1;
}
</style>
