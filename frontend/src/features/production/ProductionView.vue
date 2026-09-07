<script setup lang="ts">
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
import { watch } from 'vue'
import RegisterList from '../../components/ui/RegisterList.vue'
import RegisterRow from '../../components/ui/RegisterRow.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import InlineAlert from '../../components/ui/InlineAlert.vue';
import FormGrid from '../../components/ui/FormGrid.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import InspectorSection from '../../components/layout/InspectorSection.vue';
import MasterDetail from '../../components/layout/MasterDetail.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref } from 'vue';
import {
  CheckCircle2,
  Factory,
  PackageOpen,
  Plus,
  RotateCcw,
  Save,
  Trash2,
  X,
} from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';
import { productionApi, type ProductionJobRecord, type ReservationRecord } from '../../api/production';
import type { OrderRecord } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';

const props = defineProps<{
  currencyUnit: CurrencyUnit;
  orders: OrderRecord[];
  materials: any[];
  machines: any[];
  suppliers: any[];
}>();
const emit = defineEmits<{ notify: [message: string] }>();
import {useProductionWorkspace} from './useProductionWorkspace'
const {busy,runAction,jobs,selectedId,statusFilter,searchQuery,loading,error,createMode,saving,form,reservations,reservationMaterial,reservationQuantity,consumedQuantity,wasteQuantity,consumptionMaterial,consumptionKey,consumptionNotes,outsourceSupplier,outsourceDescription,outsourceCost,outsourceQuotedCost,selected,selectedOrder,selectedItem,visibleJobs,load,select,beginCreate,chooseOrder,create,changeStatus,remove,reserve,release,consume,saveOutsource,replace,jobOrder,jobContext,statusTone,message,date}=useProductionWorkspace(props,emit)
</script>

<template>
  <div class="min-w-0 space-y-3">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        title="Production"
        eyebrow="Operations / production ledger"
        description="Reserve material, record actual usage, and keep job progress independent from commercial status."
        ><button class="btn btn-primary" type="button" @click="beginCreate">
          <Plus :size="16" />New production job
        </button></WorkspaceHeader
      >
      <SearchFilterBar><template #search><SearchField
          v-model="searchQuery"
          label="Search production jobs"
          placeholder="Search job, order, customer, or service"
        /></template><template #filters><SelectField
          v-model="statusFilter"
          label="Queue"
          aria-label="Production status"
          :options="
            ['All', 'Pending', 'Ready', 'In Progress', 'Paused', 'Completed', 'Cancelled'].map(
              (value) => ({ label: value, value }),
            )
          "
        /></template><template #count><span class="self-end pb-2">{{ visibleJobs.length }} jobs</span></template></SearchFilterBar>
    </WorkspaceStickyStack>
    <InlineAlert v-if="error" role="alert" class="min-w-0 space-y-3" tone="error"
      ><span>{{ error }}</span
      ><button class="btn btn-ghost" type="button" @click="error = ''" aria-label="Dismiss">
        <X :size="15" /></button
    ></InlineAlert>
    <MasterDetail>
      <RegisterList title="Production queue" subtitle="Select a job to manage reservations and actual usage." :count="visibleJobs.length">
<LoadingState v-if="loading" label="Loading production jobs…" />
<EmptyState v-else-if="!visibleJobs.length" title="No production jobs in this view" description="Adjust the queue filter or create a job." />
<template v-else><RegisterRow v-for="job in visibleJobs" :key="job.id" :selected="selectedId === job.id" @activate="select(job.id)">
<template #icon><Factory :size="17" /></template>
<template #identity><strong class="block text-sm">{{job.jobNumber}} · {{job.serviceName}}</strong></template>
<template #meta><p class="mt-1 text-xs leading-5 text-base-content/60">{{jobOrder(job)}} · {{job.quantity}} {{job.quantityUnit}}</p><p class="text-xs leading-5 text-base-content/50">{{jobContext(job)}}</p></template>
<template #status><StatusBadge :label="job.status" :tone="statusTone(job.status)" /></template>
</RegisterRow></template>
</RegisterList>

      <InspectorShell
        v-if="createMode"
        title="New production job"
        subtitle="The service and cost are copied from the immutable order item snapshot."
      >
        <form @submit.prevent="create" class="min-w-0 space-y-3">
          <SelectField
            v-model="form.orderId"
            label="Confirmed order"
            :options="[
              { label: 'Select an order', value: '' },
              ...props.orders
                .filter((order) => order.commercialStatus === 'Confirmed')
                .map((order) => ({
                  label: `${order.orderNumber} · ${order.customerName || 'Walk-in'}`,
                  value: order.id,
                })),
            ]"
            @update:model-value="chooseOrder"
          /><SelectField
            v-model="form.orderItemId"
            label="Order item"
            :options="[
              { label: 'Select an item', value: '' },
              ...(selectedOrder?.items ?? []).map((item) => ({
                label: `${item.serviceName} · ${item.quantity} ${item.quantityUnit}`,
                value: item.id,
              })),
            ]"
          /><FormGrid
            ><FormField class="gap-1"
              ><span>Quantity</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.quantity"
                inputmode="decimal" /></FormField
            ><FormField class="gap-1"
              ><span>Unit</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.quantityUnit" /></FormField></FormGrid
          ><FormGrid
            ><SelectField
              v-model="form.priority"
              label="Priority"
              :options="
                ['Urgent', 'High', 'Normal', 'Low'].map((value) => ({ label: value, value }))
              " /><SelectField
              v-model="form.assignedMachineId"
              label="Machine"
              :options="[
                { label: 'Unassigned', value: '' },
                ...props.machines.map((machine) => ({ label: machine.name, value: machine.id })),
              ]" /></FormGrid
          ><FormField class="gap-1"
            ><span>Notes</span
            ><textarea
              class="textarea w-full min-w-0"
              v-model="form.notes"
              rows="3"
            ></textarea>
          </FormField>
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-ghost" type="button" @click="createMode = false">Cancel</button
            ><button class="btn btn-primary" type="submit" :disabled="busy || (saving)">
              <Save :size="15" />Create job
            </button>
          </div>
        </form>
      </InspectorShell>

      <InspectorShell
        v-else-if="selected"
        title="Job inspector"
        :subtitle="`${selected.serviceName} · ${jobOrder(selected)}`"
      >
        <div class="min-w-0 space-y-3">
          <StatusBadge :label="selected.status" :tone="statusTone(selected.status)" /><span>{{
            date(selected.createdAt)
          }}</span>
        </div>
        <div class="min-w-0 space-y-3">
          <Factory :size="19" />
          <div class="min-w-0 space-y-3">
            <h3 class="text-sm font-semibold">{{ selected.jobNumber }}</h3>
            <p>
              {{ selected.quantity }} {{ selected.quantityUnit }} · {{ selected.priority }} priority
            </p>
          </div>
        </div>
        <InspectorSection title="Cost and schedule">
        <dl class="grid min-w-0 gap-2 text-sm">
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Estimated cost</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ formatMoney(selected.estimatedCostRial, props.currencyUnit) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Actual cost</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ formatMoney(selected.actualTotalCostRial, props.currencyUnit) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Started</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ date(selected.startedAt) }}
            </dd>
          </div>
          <div
            class="grid min-w-0 grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)] gap-3 border-b border-base-300 py-2 last:border-0"
          >
            <dt class="text-xs text-base-content/60">Completed</dt>
            <dd class="min-w-0 text-end tabular-nums wrap-anywhere">
              {{ date(selected.completedAt) }}
            </dd>
          </div>
        </dl>
        </InspectorSection>
        <div class="flex flex-wrap items-center gap-2">
          <button
            class="btn btn-ghost"
            v-if="selected.status === 'Pending'"
            @click="changeStatus('Ready')"
           :disabled="busy">
            Mark ready</button
          ><button
            class="btn btn-ghost"
            v-if="
              selected.status === 'Pending' ||
              selected.status === 'Ready' ||
              selected.status === 'Paused'
            "
            @click="changeStatus('In Progress')"
           :disabled="busy">
            Start production</button
          ><button
            class="btn btn-ghost"
            v-if="selected.status === 'In Progress'"
            @click="changeStatus('Paused')"
           :disabled="busy">
            Pause</button
          ><button
            class="btn btn-ghost"
            v-if="selected.status === 'In Progress'"
            @click="changeStatus('Completed')"
           :disabled="busy">
            <CheckCircle2 :size="15" />Complete</button
          ><button
            class="btn btn-ghost"
            v-if="selected.status !== 'Completed' && selected.status !== 'Cancelled'"
            @click="changeStatus('Cancelled')"
           :disabled="busy">
            Cancel
          </button>
        </div>
        <InspectorSection title="Reservations">
          <div
            v-for="r in reservations"
            :key="r.id"
            class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
          >
            <span
              >{{ props.materials.find((m) => m.id === r.materialId)?.name ?? r.materialId }} ·
              {{ r.quantity }}</span
            ><button class="btn btn-ghost" v-if="r.status === 'active'" @click="release(r)" :disabled="busy">
              Release</button
            ><small v-else class="block text-xs leading-5 text-base-content/60">{{
              r.status
            }}</small>
          </div>
          <FormGrid
            ><SelectField
              v-model="reservationMaterial"
              label="Material"
              :options="[
                { label: 'Material', value: '' },
                ...props.materials.map((material) => ({
                  label: `${material.name} · available ${material.availableStock}`,
                  value: material.id,
                })),
              ]" /><FormField label="Qty"><AppInput
              class="input w-full min-w-0"
              v-model="reservationQuantity"
              placeholder="Qty"
              inputmode="decimal" /></FormField></FormGrid
          ><button class="btn btn-ghost" @click="reserve" :disabled="busy">
            <RotateCcw :size="15" />Reserve stock
          </button>
        </InspectorSection>
        <div class="min-w-0 space-y-3">
          <h3 class="text-sm font-semibold">Actual consumption & waste</h3>
          <FormGrid
            ><SelectField
              v-model="consumptionMaterial"
              label="Material"
              :options="[
                { label: 'Material', value: '' },
                ...props.materials.map((material) => ({
                  label: material.name,
                  value: material.id,
                })),
              ]" /><FormField label="Consumed"><AppInput
              class="input w-full min-w-0"
              v-model="consumedQuantity"
              placeholder="Consumed"
              inputmode="decimal" /></FormField><FormField label="Waste"><AppInput
              class="input w-full min-w-0"
              v-model="wasteQuantity"
              placeholder="Waste"
              inputmode="decimal" /></FormField></FormGrid
          ><FormField label="Idempotency key (optional)"><AppInput
            class="input w-full min-w-0"
            v-model="consumptionKey"
            placeholder="Idempotency key (optional)"
          /></FormField><button class="btn btn-primary" @click="consume" :disabled="busy">Post immutable movement</button>
        </div>
        <div class="min-w-0 space-y-3">
          <h3 class="text-sm font-semibold">Outsourced production</h3>
          <FormGrid
            ><SelectField
              v-model="outsourceSupplier"
              label="Supplier"
              :options="[
                { label: 'Supplier', value: '' },
                ...props.suppliers.map((supplier) => ({
                  label: supplier.name,
                  value: supplier.id,
                })),
              ]" /><FormField class="gap-1"
              ><span>Quoted ({{ props.currencyUnit }})</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="outsourceQuotedCost"
                placeholder="0"
                inputmode="numeric" /></FormField
            ><FormField class="gap-1"
              ><span>Actual ({{ props.currencyUnit }})</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="outsourceCost"
                placeholder="0"
                inputmode="numeric" /></FormField></FormGrid
          ><FormField label="Outsourced scope / return details"><AppInput
            class="input w-full min-w-0"
            v-model="outsourceDescription"
            placeholder="Outsourced scope / return details"
          /></FormField><button class="btn btn-primary" @click="saveOutsource" :disabled="busy">Save outsourcing</button>
        </div>
        <button
          class="btn btn-ghost"
          v-if="selected.status === 'Pending' || selected.status === 'Ready'"
          @click="remove"
         :disabled="busy">
          <Trash2 :size="14" />Delete unposted draft
        </button>
      </InspectorShell>
      <InspectorShell v-else title="Job inspector" subtitle="Select a production job to inspect it."
        ><PackageOpen :size="22" />
        <p>Production details will appear here.</p></InspectorShell
      >
    </MasterDetail>
  </div>
</template>
