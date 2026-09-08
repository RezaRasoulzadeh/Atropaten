<script setup lang="ts">
import DashboardCharts from './DashboardCharts.vue';
import DataTable from '../../components/ui/DataTable.vue';
import SelectField from '../../components/ui/SelectField.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import {
  BarChart3,
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
import { reportsApi, type DashboardRecord } from '../../api/reports';
import { formatMoney } from '../../utils/currency';
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
        :eyebrow="`${formatDate(data?.startDate || start)} → ${formatDate(data?.endDate || end)}`"
        title="Dashboard"
        description="Sales performance and orders to follow up."
      >
        <SelectField
          :model-value="period"
          :disabled="loading"
          :options="[
            { label: 'Last 30 days', value: '30' },
            { label: 'Last 90 days', value: '90' },
            { label: 'Last 365 days', value: '365' },
          ]"
          aria-label="Dashboard period"
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
          /><span>Refresh</span></button
        ><button
          class="btn btn-primary gap-2"
          type="button"
          @click="emit('newOrder')"
        >
          <FilePlus2 :size="15" :stroke-width="1.8" aria-hidden="true" /><span
            >New order</span
          >
        </button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>
    <div class="mt-4 space-y-4">
      <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <KpiCard
          :value="money(data?.revenueRial || 0)"
          detail="Sales in selected period"
          title="Sales"
          :trend="loading ? 'Loading' : 'Period'"
          :icon="TrendingUp"
          accent="blue"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.grossProfitRial || 0)"
          detail="Sales less cost of goods"
          title="Gross profit"
          trend="Period"
          :icon="BarChart3"
          accent="green"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.receivableRial || 0)"
          :detail="`${data?.openInvoiceCount || 0} open invoices`"
          title="Receivables"
          trend="Outstanding"
          :icon="HandCoins"
          accent="amber"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.payableRial || 0)"
          detail="Outstanding supplier balance"
          title="Payables"
          trend="Supplier"
          :icon="ReceiptText"
          accent="red"
          :loading="initialLoading"
        />
      </section>
      <div
        v-if="initialLoading"
        class="grid gap-4 xl:grid-cols-2"
        aria-label="Loading dashboard charts"
      >
        <div v-for="n in 2" :key="n" class="skeleton h-72 bg-base-300/70"></div>
      </div>
      <DashboardCharts
        v-else-if="data"
        :data="data"
        :currency-unit="currencyUnit"
      />
      <AppPanel
        title="Orders needing attention"
        subtitle="Unfinished confirmed orders and orders with reference attachments"
        :flush="true"
      >
        <template #action
          ><button
            class="btn btn-ghost btn-sm"
            @click="emit('navigate', 'Orders')"
          >
            All orders
          </button></template
        >
        <p v-if="initialLoading" class="p-4 text-sm text-base-content/60">
          Loading orders…
        </p>
        <DataTable
          v-else-if="data?.ordersNeedingAttention?.length"
          label="Orders needing attention"
        >
          <thead>
            <tr>
              <th>Order / Customer</th>
              <th>Follow-up</th>
              <th>Fulfillment</th>
              <th>Due date</th>
              <th class="text-end">Total</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in data.ordersNeedingAttention" :key="order.id">
              <td>
                <button
                  :aria-label="`Open order ${order.orderNumber}`"
                  class="text-start font-semibold text-primary hover:underline"
                  @click="emit('openOrder', order.id)"
                >
                  {{ order.orderNumber }}</button
                ><span
                  class="block max-w-60 truncate text-xs text-base-content/60"
                  :title="order.customer"
                  >{{ order.customer || 'Walk-in customer' }}</span
                >
              </td>
              <td>
                <div class="flex flex-wrap gap-1">
                  <StatusBadge
                    v-if="order.commercialStatus === 'Confirmed'"
                    label="Confirmed"
                    tone="blue"
                  /><StatusBadge
                    v-if="order.referenceCount"
                    :label="`${order.referenceCount} reference ${order.referenceCount === 1 ? 'file' : 'files'}`"
                    tone="amber"
                  />
                </div>
              </td>
              <td>
                <StatusBadge :label="order.fulfillmentStatus" tone="slate" />
              </td>
              <td>
                {{
                  order.dueDate ? formatDate(order.dueDate) : 'Not scheduled'
                }}
              </td>
              <td class="text-end">{{ money(order.totalRial) }}</td>
            </tr>
          </tbody>
        </DataTable>
        <p v-else-if="data" class="p-4 text-sm text-base-content/60">
          No orders need follow-up.
        </p>
        <p v-else class="p-4 text-sm text-base-content/60">
          Order data is unavailable. Refresh to try again.
        </p>
      </AppPanel>
      <RegisterList
        title="Production queue"
        subtitle="Active production jobs"
        :count="data?.production?.length || 0"
      >
        <div
          v-if="initialLoading"
          class="space-y-3 p-4"
          aria-label="Loading production jobs"
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
                    job.customer || 'Walk-in customer'
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
        <p v-else class="px-4 py-6 text-sm text-base-content/60">
          No active production jobs.
        </p>
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
