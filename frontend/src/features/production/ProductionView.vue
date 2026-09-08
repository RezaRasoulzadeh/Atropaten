<script setup lang="ts">
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import LoadingState from '../../components/ui/LoadingState.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import ProductionJobWorkspaceView from './ProductionJobWorkspaceView.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import { Factory, Plus } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';
import type { OrderRecord } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';

const props = defineProps<{
  currencyUnit: CurrencyUnit;
  orders: OrderRecord[];
  materials: any[];
  machines: any[];
  suppliers: any[];
}>();
const emit = defineEmits<{ notify: [message: string] }>();
import { useProductionWorkspace } from './useProductionWorkspace';
const workspace = useProductionWorkspace(props, emit);
const {
  jobs,
  selectedId,
  selected,
  statusFilter,
  searchQuery,
  loading,
  createMode,
  visibleJobs,
  select,
  beginCreate,
  jobOrder,
  jobContext,
  statusTone,
} = workspace;
function back() {
  selectedId.value = null;
  createMode.value = false;
}
</script>

<template>
  <div v-if="!selected && !createMode" class="min-w-0 space-y-3">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        title="Production"
        eyebrow="Operations / production ledger"
        description="Reserve material, record actual usage, and keep job progress independent from commercial status."
        ><button class="btn btn-primary" type="button" @click="beginCreate">
          <Plus :size="16" />New production job
        </button></WorkspaceHeader
      >
      <SearchFilterBar
        ><template #search
          ><SearchField
            v-model="searchQuery"
            label="Search production jobs"
            placeholder="Search job, order, customer, or service" /></template
        ><template #filters
          ><SelectField
            v-model="statusFilter"
            label="Queue"
            aria-label="Production status"
            :options="
              ['All', 'Pending', 'Ready', 'In Progress', 'Paused', 'Completed', 'Cancelled', 'Failed'].map(
                (value) => ({ label: value, value }),
              )
            " /></template
        ><template #count
          ><span>{{ visibleJobs.length }} jobs</span></template
        ></SearchFilterBar
      >
    </WorkspaceStickyStack>
    <RegisterList
      title="Production queue"
      subtitle="Open a job to manage reservations and actual usage."
      :count="visibleJobs.length"
    >
      <LoadingState v-if="loading" label="Loading production jobs…" />
      <EmptyState
        v-else-if="!visibleJobs.length"
        class="min-h-48 flex-col text-center"
        :title="jobs.length ? 'No jobs match this queue' : 'Production queue is empty'"
        :description="
          jobs.length
            ? 'Try another queue filter to find a production job.'
            : 'Create a job from a confirmed order to start tracking production.'
        "
      >
        <template #icon><Factory :size="32" :stroke-width="1.5" aria-hidden="true" /></template>
        <template v-if="!jobs.length" #action>
          <button class="btn btn-outline btn-sm mt-3 gap-2" type="button" @click="beginCreate">
            <Plus :size="14" aria-hidden="true" /> New production job
          </button>
        </template>
      </EmptyState>
      <template v-else
        ><RegisterRow v-for="job in visibleJobs" :key="job.id" @activate="select(job.id)">
          <template #icon><Factory :size="17" /></template>
          <template #identity>
            <div class="flex min-w-0 items-start justify-between gap-3">
              <div class="min-w-0">
                <strong class="text-sm">{{ job.jobNumber }}</strong
                ><span class="ms-2 text-sm text-base-content/60">{{
                  props.orders.find((order) => order.id === job.orderId)?.customerName ||
                  'Walk-in customer'
                }}</span>
              </div>
              <StatusBadge :label="job.status" :tone="statusTone(job.status)" />
            </div>
          </template>
          <template #meta>
            <div class="mt-2 grid min-w-0 gap-x-4 gap-y-2 text-xs sm:grid-cols-2 xl:grid-cols-3">
              <div class="min-w-0">
                <span class="block text-base-content/50">Service</span
                ><span class="block break-words">{{ job.serviceName }}</span>
              </div>
              <div>
                <span class="block text-base-content/50">Order / Quantity</span
                ><span>{{ jobOrder(job) }} · {{ job.quantity }} {{ job.quantityUnit }}</span>
              </div>
              <div class="min-w-0">
                <span class="block text-base-content/50">Schedule / Machine</span
                ><span class="block break-words">{{ jobContext(job) }}</span>
              </div>
            </div>
          </template>
        </RegisterRow></template
      >
    </RegisterList>
  </div>
  <ProductionJobWorkspaceView v-else v-bind="props" :workspace="workspace" @back="back" />
</template>
