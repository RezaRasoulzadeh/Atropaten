<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import { computed, onMounted, ref } from 'vue';
import { Factory, PackageOpen } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import {
  productionApi,
  type ConsumptionRecord,
  type ProductionJobRecord,
  type ReservationRecord,
} from '../../api/production';
import type { OrderRecord } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import EmptyState from '../../components/ui/EmptyState.vue';
const props = defineProps<{ order: OrderRecord; currencyUnit: CurrencyUnit }>();
const jobs = ref<ProductionJobRecord[]>([]);
const reserved = ref<Record<string, ReservationRecord[]>>({});
const usage = ref<Record<string, ConsumptionRecord[]>>({});
const progress = computed(() =>
  jobs.value.length
    ? Math.round(
        (jobs.value.filter((j) => j.status === 'Completed').length / jobs.value.length) * 100,
      )
    : 0,
);
onMounted(async () => {
  try {
    jobs.value = (await productionApi.list('All')).filter((j) => j.orderId === props.order.id);
    await Promise.all(
      jobs.value.map(async (j) => {
        reserved.value[j.id] = await productionApi.reservations('', j.id, '');
        usage.value[j.id] = await productionApi.consumptions(j.id);
      }),
    );
  } catch (e) {
    reportError(e);
  }
});
function reservedCount(id: string) {
  return (reserved.value[id] ?? []).filter((r) => r.status === 'active').length;
}
function usageCount(id: string) {
  const rows = usage.value[id] ?? [];
  return rows.reduce((n, r) => n + Number(r.consumedQuantity) + Number(r.wasteQuantity), 0);
}
function tone(s: string) {
  return s === 'Completed'
    ? 'green'
    : s === 'Cancelled' || s === 'Failed'
      ? 'red'
      : s === 'In Progress' || s === 'Ready'
        ? 'blue'
        : s === 'Paused'
          ? 'amber'
          : 'slate';
}
</script>
<template>
  <section class="min-w-0 space-y-4">
    <header class="rounded-box border border-base-300 bg-base-100 p-4">
      <div class="flex min-w-0 flex-wrap items-start justify-between gap-3">
        <div class="min-w-0">
          <h2 class="text-base font-semibold">Production jobs</h2>
          <p class="mt-1 text-xs text-base-content/60">
            Track execution independently from payment and commercial status.
          </p>
        </div>
        <div class="shrink-0 text-end">
          <strong class="block text-lg tabular-nums">{{ jobs.length }}</strong>
          <span class="text-xs text-base-content/60">{{ progress }}% completed</span>
        </div>
      </div>
      <div class="mt-4 h-2 overflow-hidden rounded-full bg-base-300" aria-hidden="true">
        <div
          class="h-full rounded-full bg-primary transition-[width] duration-300"
          :style="{ width: `${progress}%` }"
        ></div>
      </div>
    </header>
    <div v-if="jobs.length" class="space-y-3">
      <div
        v-for="job in jobs"
        :key="job.id"
        class="min-w-0 rounded-box border border-base-300 bg-base-100 p-4"
      >
        <div class="flex min-w-0 flex-wrap items-start justify-between gap-4">
          <div class="flex min-w-0 items-start gap-3">
            <div class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary">
              <Factory :size="17" aria-hidden="true" />
            </div>
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <strong class="truncate text-sm">{{ job.jobNumber }}</strong>
                <StatusBadge :label="job.status" :tone="tone(job.status)" />
              </div>
              <p class="mt-1 truncate text-sm text-base-content/80">{{ job.serviceName }}</p>
            </div>
          </div>
          <div class="shrink-0 text-end">
            <span class="block text-xs text-base-content/50">Actual cost</span>
            <strong class="block text-sm tabular-nums">{{
              formatMoney(job.actualTotalCostRial, props.currencyUnit)
            }}</strong>
          </div>
        </div>
        <div class="mt-4 grid min-w-0 gap-3 border-t border-base-300 pt-3 text-xs sm:grid-cols-3">
          <div class="min-w-0">
            <span class="block text-base-content/50">Quantity</span>
            <strong class="block truncate font-medium text-base-content/80"
              >{{ job.quantity }} {{ job.quantityUnit }}</strong
            >
          </div>
          <div class="min-w-0">
            <span class="block text-base-content/50">Created</span>
            <strong class="block truncate font-medium text-base-content/80">{{
              job.createdAt ? formatDateTime(job.createdAt) : '—'
            }}</strong>
          </div>
          <div class="min-w-0">
            <span class="block text-base-content/50">Material activity</span>
            <strong class="block truncate font-medium text-base-content/80"
              >{{ reservedCount(job.id) }} reservations · {{ usageCount(job.id) }} used / waste</strong
            >
          </div>
        </div>
      </div>
    </div>
    <EmptyState
      v-else
      title="No production jobs linked"
      description="Confirm the order, then create a job from the Production workspace."
    >
      <template #icon><PackageOpen :size="24" :stroke-width="1.6" aria-hidden="true" /></template>
    </EmptyState>
  </section>
</template>
