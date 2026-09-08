<script setup lang="ts">
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onBeforeUnmount, onMounted, ref, watch, type Component } from 'vue';
import {
  BarChart3,
  CircleAlert,
  FilePlus2,
  HandCoins,
  ReceiptText,
  RefreshCw,
  TriangleAlert,
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
import { formatDate, formatDateTime, currentCanonicalDate } from '../../utils/date';
import { normalizeError } from '../../ui/feedback';
const props = defineProps<{ currencyUnit: 'Rial' | 'Toman' }>();
const emit = defineEmits<{ navigate: [view: string]; newOrder: []; notify: [message: string] }>();
const data = ref<DashboardRecord | null>(null);
const loading = ref(false);
const error = ref('');
const refreshAnimation = ref<'once' | 'infinite' | ''>('');
let refreshTimer: ReturnType<typeof setTimeout> | undefined;
let refreshClearTimer: ReturnType<typeof setTimeout> | undefined;
const end = currentCanonicalDate();
const start = new Date(Date.now() - 30 * 86400000).toISOString().slice(0, 10);
const money = (v: number) => formatMoney(v, props.currencyUnit);
async function load() {
  loading.value = true;
  refreshAnimation.value = 'once';
  if (refreshTimer) clearTimeout(refreshTimer);
  if (refreshClearTimer) clearTimeout(refreshClearTimer);
  refreshTimer = setTimeout(() => {
    if (loading.value) refreshAnimation.value = 'infinite';
  }, 700);
  const startedAt = Date.now();
  try {
    data.value = await reportsApi.dashboard(start, end);
  } catch (e) {
    error.value = normalizeError(e).message;
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
watch(() => props.currencyUnit, load);
onBeforeUnmount(() => {
  if (refreshTimer) clearTimeout(refreshTimer);
  if (refreshClearTimer) clearTimeout(refreshClearTimer);
});
const initialLoading = computed(() => loading.value && !data.value);
const attention = computed(() => data.value?.attention ?? []);
</script>
<template>
  <div>
    <WorkspaceStickyStack>
      <WorkspaceHeader
        :eyebrow="`${data?.startDate || start} → ${data?.endDate || end}`"
        title="Good morning"
        description="Here is what needs your attention today."
      >
        <button class="btn btn-outline btn-primary" type="button" @click="load">
          <RefreshCw
            :class="{
              'refresh-once': refreshAnimation === 'once',
              'animate-spin': refreshAnimation === 'infinite',
            }"
            :size="15"
          /><span>Refresh</span></button
        ><button class="btn btn-primary gap-2" type="button" @click="emit('newOrder')">
          <FilePlus2 :size="15" :stroke-width="1.8" aria-hidden="true" /><span>New order</span>
        </button>
      </WorkspaceHeader>
    </WorkspaceStickyStack>
    <p v-if="error">{{ error }}</p>
    <div class="mt-4 space-y-4">
      <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <KpiCard
          :value="money(data?.revenueRial || 0)"
          detail="Posted invoices in period"
          title="Sales"
          :trend="loading ? 'Loading' : 'Persisted'"
          :icon="TrendingUp"
          accent="blue"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.grossProfitRial || 0)"
          detail="Journal-derived P&L"
          title="Gross profit"
          trend="Journal"
          :icon="BarChart3"
          accent="green"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.receivableRial || 0)"
          :detail="`${data?.openInvoiceCount || 0} open invoices`"
          title="Receivables"
          trend="Authoritative"
          :icon="HandCoins"
          accent="amber"
          :loading="initialLoading"
        />
        <KpiCard
          :value="money(data?.payableRial || 0)"
          detail="Journal-derived"
          title="Payables"
          trend="Supplier"
          :icon="ReceiptText"
          accent="red"
          :loading="initialLoading"
        />
      </section>
      <section class="grid gap-4 xl:grid-cols-2">
        <AppPanel
          title="Needs attention"
          subtitle="Due obligations and operational exceptions"
          :flush="true"
        >
          <div
            v-if="initialLoading"
            class="min-h-72 space-y-3 p-4"
            aria-label="Loading attention items"
          >
            <div v-for="row in 6" :key="row" class="skeleton h-10 w-full bg-base-300/70"></div>
          </div>
          <div v-else class="divide-y divide-base-300">
            <button
              v-for="item in attention"
              :key="`${item.kind}-${item.detail}`"
              class="flex w-full items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200"
              type="button"
              @click="emit('notify', item.detail)"
            >
              <CircleAlert class="shrink-0 text-warning" :size="16" />
              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-semibold">{{ item.label }}</span>
                <span class="block truncate text-xs text-base-content/60"
                  >{{ item.detail
                  }}<template v-if="item.date"> · {{ formatDate(item.date) }}</template></span
                >
              </span>
              <span class="shrink-0 text-xs font-semibold text-warning">{{
                item.amountRial ? money(item.amountRial) : '—'
              }}</span>
            </button>
            <p v-if="!attention.length" class="px-4 py-6 text-sm text-base-content/60">
              No attention items from persisted data.
            </p>
          </div>
        </AppPanel>
        <RegisterList
          title="Production queue"
          subtitle="Persisted active jobs"
          :count="data?.production?.length || 0"
        >
          <div v-if="initialLoading" class="space-y-3 p-4" aria-label="Loading production jobs">
            <div v-for="row in 3" :key="row" class="skeleton h-10 w-full bg-base-300/70"></div>
          </div>
          <div v-else-if="data?.production?.length">
            <RegisterRow v-for="job in data.production" :key="job.id" :interactive="false">
              <template #identity>
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <strong class="block truncate text-sm">{{ job.orderNumber || job.id }}</strong>
                    <span class="block truncate text-xs text-base-content/60">{{ job.customer || 'Walk-in customer' }}</span>
                  </div>
                  <StatusBadge :label="job.status" tone="blue" />
                </div>
              </template>
              <template #meta>
                <div class="mt-2 min-w-0 text-xs text-base-content/80">
                  <span class="block truncate">{{ job.service }}</span>
                </div>
              </template>
            </RegisterRow>
          </div>
          <p v-else class="px-4 py-6 text-sm text-base-content/60">No active production jobs.</p>
        </RegisterList>
      </section>
      <section class="grid gap-4 xl:grid-cols-2">
        <AppPanel title="Low stock" subtitle="Movement-derived availability" :flush="true">
          <div
            v-if="initialLoading"
            class="min-h-28 space-y-3 p-4"
            aria-label="Loading low stock items"
          >
            <div v-for="row in 2" :key="row" class="skeleton h-8 w-full bg-base-300/70"></div>
          </div>
          <div v-else class="divide-y divide-base-300">
            <div
              v-for="item in data?.lowStock"
              :key="item.id"
              class="flex items-center gap-3 px-4 py-3"
            >
              <TriangleAlert class="shrink-0 text-warning" :size="16" />
              <span class="min-w-0 flex-1 truncate text-sm font-semibold">{{ item.name }}</span>
              <span class="shrink-0 text-xs text-base-content/70"
                >{{ item.availableUnits / 1000000 }} {{ item.unit }}</span
              >
              <span class="hidden shrink-0 text-xs text-base-content/50 sm:inline"
                >Reorder at {{ item.reorderLevelUnits / 1000000 }}</span
              >
            </div>
            <p v-if="!data?.lowStock?.length" class="px-4 py-6 text-sm text-base-content/60">
              No low-stock materials.
            </p>
          </div>
        </AppPanel>
        <AppPanel
          title="Recent payments"
          subtitle="Latest persisted financial activity"
          :flush="true"
        >
          <div
            v-if="initialLoading"
            class="min-h-28 space-y-3 p-4"
            aria-label="Loading recent payments"
          >
            <div v-for="row in 2" :key="row" class="skeleton h-8 w-full bg-base-300/70"></div>
          </div>
          <div v-else class="divide-y divide-base-300">
            <div
              v-for="item in data?.recentActivity"
              :key="item.id"
              class="flex items-center gap-3 px-4 py-3"
            >
              <span class="shrink-0 text-info"><HandCoins :size="16" /></span>
              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-semibold">{{ item.label }}</span>
                <span class="block truncate text-xs text-base-content/60"
                  >{{ item.detail }} · {{ formatDateTime(item.date) }}</span
                >
              </span>
              <span
                class="shrink-0 text-sm font-semibold"
                :class="item.direction === 'incoming' ? 'text-success' : 'text-error'"
                >{{ money(item.amountRial) }}</span
              >
            </div>
            <p v-if="!data?.recentActivity?.length" class="px-4 py-6 text-sm text-base-content/60">
              No payments recorded.
            </p>
          </div>
        </AppPanel>
      </section>
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
