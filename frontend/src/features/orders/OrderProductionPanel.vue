<script setup lang="ts">
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import { computed, onMounted, ref, watch } from 'vue';
import { CheckCheck, Factory, PackageOpen, Plus } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import ProductionCostBreakdown from '../production/ProductionCostBreakdown.vue';
import {
  productionApi,
  type ConsumptionRecord,
  type ProductionJobRecord,
  type ReservationRecord,
} from '../../api/production';
import { ordersApi, type OrderRecord } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import EmptyState from '../../components/ui/EmptyState.vue';
const props = defineProps<{ order: OrderRecord; currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [message: string]; saved: [order: OrderRecord] }>();
const jobs = ref<ProductionJobRecord[]>([]);
const reserved = ref<Record<string, ReservationRecord[]>>({});
const usage = ref<Record<string, ConsumptionRecord[]>>({});
const addingItemId = ref<string | null>(null);
const orderItems = computed(() =>
  (Array.isArray(props.order.items) ? props.order.items : []).slice().sort((a, b) => a.position - b.position),
);
const canAddToProduction = computed(() =>
  !props.order.archived &&
  !['Cancelled', 'Closed'].includes(props.order.commercialStatus) &&
  props.order.fulfillmentStatus !== 'Delivered',
);
const progress = computed(() =>
  jobs.value.length
    ? Math.round(
        (jobs.value.filter((j) => j.status === 'Completed').length / jobs.value.length) * 100,
      )
    : 0,
);
async function load() {
  try {
    jobs.value = (await productionApi.list('All')).filter((j) => j.orderId === props.order.id);
    await Promise.all(
      jobs.value.map(async (j) => {
        await productionApi.materials(j.id);
        reserved.value[j.id] = await productionApi.reservations('', j.id, '');
        usage.value[j.id] = await productionApi.consumptions(j.id);
      }),
    );
  } catch (e) {
    reportError(e);
  }
}
onMounted(load);
watch(() => [props.order.id,props.order.updatedAt], load);

function itemJob(itemId: string) {
  return jobs.value.find((job) => job.orderItemId === itemId && job.status !== 'Cancelled')
    ?? jobs.value.find((job) => job.orderItemId === itemId);
}

function itemActionLabel(item: OrderRecord['items'][number]) {
  const job = itemJob(item.id);
  if (job?.status === 'Cancelled') return 'Add again';
  if (job) return job.status;
  return 'Add to production';
}

async function addItemToProduction(item: OrderRecord['items'][number]) {
  if (addingItemId.value || busy.value || !canAddToProduction.value || itemJob(item.id)?.status === 'Completed') return;
  return runAction(async () => {
    addingItemId.value = item.id;
    const wasDraft = props.order.commercialStatus === 'Draft';
    let jobCreated = false;
    try {
      if (wasDraft) await ordersApi.commercialStatus(props.order.id, 'Confirmed');
      const job = await productionApi.create({
        orderId: props.order.id,
        orderItemId: item.id,
        quantity: item.quantity,
        quantityUnit: item.quantityUnit,
        assignedMachineId: '',
        priority: props.order.priority || 'Normal',
        notes: '',
        plannedAt: null,
      });
      jobCreated = true;
      jobs.value = [...jobs.value, job];
      reserved.value[job.id] = await productionApi.reservations('',job.id,'');
      usage.value[job.id] = [];
      emit('saved', await ordersApi.get(props.order.id));
      emit('notify', wasDraft ? `${item.serviceName} added to production and order confirmed.` : `${item.serviceName} added to production.`);
    } catch (e) {
      if (wasDraft && !jobCreated) {
        try {
          await ordersApi.commercialStatus(props.order.id, 'Draft');
        } catch {
          // Keep the original error visible; the order status can be corrected manually.
        }
      }
      reportError(e);
    } finally {
      addingItemId.value = null;
    }
  });
}
function reservedCount(id: string) {
  return (reserved.value[id] ?? []).filter((r) => r.status === 'active').length;
}
function usageCount(id: string) {
  const rows = usage.value[id] ?? [];
  return rows.filter(r=>!r.reversed).reduce((n, r) => n + Number(r.consumedQuantity) + Number(r.wasteQuantity), 0);
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
    <header class="border-b border-base-300 pb-4">
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
    <div v-if="orderItems.length" class="space-y-2 pt-4">
      <div class="flex min-w-0 flex-wrap items-center justify-between gap-2">
        <div>
          <h3 class="text-sm font-semibold">Order items</h3>
          <p class="mt-0.5 text-xs text-base-content/60">Add each service to the production queue when it is ready.</p>
        </div>
        <span v-if="props.order.commercialStatus === 'Draft'" class="text-[11px] text-base-content/55">Adding a job confirms the order.</span>
      </div>
      <div
        v-for="item in orderItems"
        :key="item.id"
        class="flex min-w-0 flex-wrap items-center gap-3 rounded-box border border-base-300 bg-base-100/55 p-3"
      >
        <span class="grid size-8 shrink-0 place-items-center rounded-box bg-base-200 text-xs font-semibold text-base-content/60">{{ item.position + 1 }}</span>
        <div class="min-w-0 flex-1">
          <strong class="block truncate text-sm">{{ item.serviceName }}</strong>
          <span class="mt-0.5 block truncate text-xs text-base-content/55">{{ item.quantity }} {{ item.quantityUnit }}</span>
        </div>
        <div class="flex shrink-0 items-center gap-2">
          <template v-if="itemJob(item.id)?.status === 'Cancelled'">
            <StatusBadge label="Cancelled" tone="red" />
            <button class="btn btn-outline btn-warning btn-sm gap-1.5" type="button" :disabled="busy || addingItemId !== null || !canAddToProduction" @click="addItemToProduction(item)">
              <Plus :size="14" aria-hidden="true" />{{ addingItemId === item.id ? 'Adding…' : 'Add again' }}
            </button>
          </template>
          <StatusBadge v-else-if="itemJob(item.id)" :label="itemJob(item.id)!.status" :tone="tone(itemJob(item.id)!.status)" />
          <button v-else class="btn btn-primary btn-sm gap-1.5" type="button" :disabled="busy || addingItemId !== null || !canAddToProduction" @click="addItemToProduction(item)">
            <Plus :size="14" aria-hidden="true" />{{ addingItemId === item.id ? 'Adding…' : 'Add to production' }}
          </button>
        </div>
      </div>
      <p v-if="!canAddToProduction" class="text-xs text-warning">This order is closed, cancelled, archived, or delivered, so new production jobs cannot be added.</p>
    </div>
    <EmptyState
      v-else
      compact
      title="No order items"
      description="Add a service item before sending work to production."
    >
      <template #icon><PackageOpen :size="24" :stroke-width="1.6" aria-hidden="true" /></template>
    </EmptyState>

    <div v-if="jobs.length" class="space-y-3 border-t border-base-300 pt-4">
      <div class="flex items-center justify-between gap-2">
        <div>
          <h3 class="text-sm font-semibold">Production activity</h3>
          <p class="mt-0.5 text-xs text-base-content/60">Execution, cost, and material activity for this order.</p>
        </div>
        <span class="badge badge-ghost text-xs">{{ jobs.length }} job{{ jobs.length === 1 ? '' : 's' }}</span>
      </div>
      <div
        v-for="job in jobs"
        :key="job.id"
        class="min-w-0 rounded-box border border-base-300 bg-base-100/55 p-3"
      >
        <div class="flex min-w-0 flex-wrap items-start justify-between gap-3">
          <div class="flex min-w-0 items-start gap-2.5">
            <div class="grid size-8 shrink-0 place-items-center rounded-box bg-base-200 text-primary">
              <Factory :size="16" aria-hidden="true" />
            </div>
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <strong class="truncate text-sm">{{ job.jobNumber }}</strong>
                <StatusBadge :label="job.status" :tone="tone(job.status)" />
              </div>
              <p class="mt-0.5 truncate text-xs text-base-content/70">{{ job.serviceName }} · {{ job.quantity }} {{ job.quantityUnit }}</p>
            </div>
          </div>
          <div class="shrink-0 text-end">
            <span class="block text-[11px] text-base-content/50">Expected total cost</span>
            <strong class="block text-sm tabular-nums">{{ formatMoney(job.projectedCostRial, props.currencyUnit) }}</strong>
          </div>
        </div>
        <details class="mt-3 border-t border-base-300 pt-2">
          <summary class="cursor-pointer text-xs text-base-content/65">Cost breakdown · includes machine, labor &amp; overhead estimates</summary>
          <ProductionCostBreakdown class="mt-2" embedded :job="job" :currency-unit="props.currencyUnit" />
        </details>
        <div class="mt-3 grid min-w-0 grid-cols-2 gap-3 border-t border-base-300 pt-3 text-xs sm:grid-cols-3">
          <div class="min-w-0"><span class="block text-base-content/50">Created</span><strong class="mt-0.5 block truncate font-medium text-base-content/80">{{ job.createdAt ? formatDateTime(job.createdAt) : '—' }}</strong></div>
          <div class="min-w-0"><span class="block text-base-content/50">Reservations</span><strong class="mt-0.5 block truncate font-medium text-base-content/80">{{ reservedCount(job.id) }} active</strong></div>
          <div class="min-w-0"><span class="block text-base-content/50">Used / waste</span><strong class="mt-0.5 block truncate font-medium text-base-content/80">{{ usageCount(job.id) }}</strong></div>
        </div>
      </div>
    </div>
  </section>
</template>
