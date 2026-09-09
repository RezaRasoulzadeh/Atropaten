<script setup lang="ts">
import { computed, ref } from 'vue';
import {
  ArrowLeft,
  CheckCircle2,
  Pause,
  Pencil,
  Play,
  RotateCcw,
  Save,
  Trash2,
  XCircle,
} from 'lucide-vue-next';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import WorkspaceTabs from '../../components/layout/WorkspaceTabs.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import FormGrid from '../../components/ui/FormGrid.vue';
import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import SelectField from '../../components/ui/SelectField.vue';
import JalaliDatePicker from '../../components/ui/JalaliDatePicker.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import LoadingState from '../../components/ui/LoadingState.vue';
import { confirmAction } from '../../ui/feedback';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { formatQuantityInput } from '../../utils/quantity';
import type { OrderRecord } from '../../api/orders';
import type { useProductionWorkspace } from './useProductionWorkspace';
const props = defineProps<{
  workspace: ReturnType<typeof useProductionWorkspace>;
  currencyUnit: CurrencyUnit;
  orders: OrderRecord[];
  materials: any[];
  machines: any[];
  suppliers: any[];
}>();
const emit = defineEmits<{ back: [] }>();
const {
  busy,
  selected,
  selectedOrder,
  createMode,
  editing,
  saving,
  form,
  reservations,
  reservationsLoading,
  reservationMaterial,
  reservationQuantity,
  consumedQuantity,
  wasteQuantity,
  consumptionMaterial,
  consumptionKey,
  outsourceSupplier,
  outsourceDescription,
  outsourceCost,
  outsourceQuotedCost,
  chooseOrder,
  create,
  update,
  beginEdit,
  cancelEdit,
  changeStatus,
  remove,
  reserve,
  release,
  consume,
  saveOutsource,
  jobOrder,
  statusTone,
  date,
} = props.workspace;
const tab = ref('Overview');
const canEdit = computed(
  () => !!selected.value && selected.value.status !== 'Completed' && selected.value.status !== 'Cancelled',
);
const productionStatuses = [
  'Pending',
  'Ready',
  'In Progress',
  'Paused',
  'Completed',
  'Cancelled',
  'Failed',
] as const;
const statusActions = computed(() => {
  if (!selected.value) return [];
  return productionStatuses
    .filter((status) => status !== selected.value?.status)
    .map((status) => ({
      status,
      label:
        status === 'In Progress'
          ? selected.value?.status === 'Paused' || selected.value?.status === 'Failed'
            ? 'Resume production'
            : 'Start production'
          : status === 'Completed'
            ? 'Complete job'
            : status === 'Cancelled'
              ? 'Cancel job'
              : status === 'Failed'
                ? 'Mark failed'
                : status === 'Paused'
                  ? 'Pause production'
                  : `Set ${status}`,
      kind: status === 'In Progress' || status === 'Completed' ? 'primary' : status === 'Cancelled' || status === 'Failed' ? 'danger' : 'secondary',
    }));
});
async function back() {
  if (busy.value) return;
  if (createMode.value && (form.value.orderId || form.value.notes || form.value.quantity)) {
    if (
      !(await confirmAction({
        title: 'Discard new production job?',
        message: 'The new job has not been saved.',
        confirmLabel: 'Discard draft',
        danger: true,
      }))
    )
      return;
  }
  if (editing.value) {
    if (
      !(await confirmAction({
        title: 'Discard job changes?',
        message: 'Your unsaved production job changes will be lost.',
        confirmLabel: 'Discard changes',
        danger: true,
      }))
    )
      return;
    cancelEdit();
  }
  emit('back');
}
</script>
<template>
  <div class="min-w-0 space-y-4">
    <WorkspaceStickyStack :flush="true">
      <WorkspaceHeader
        eyebrow="Operations / production workspace"
        :title="createMode ? 'New production job' : selected?.jobNumber || 'Production job'"
        :description="
          createMode
            ? 'Create a job from a confirmed order item.'
            : `${selected?.serviceName} · ${selected ? jobOrder(selected) : ''}`
        "
      >
        <template v-if="selected && !createMode" #title-suffix
          ><StatusBadge :label="selected.status" :tone="statusTone(selected.status)"
        /></template>
        <button class="btn btn-ghost btn-sm gap-2" type="button" :disabled="busy" @click="back">
          <ArrowLeft :size="16" aria-hidden="true" />Production
        </button>
        <template v-if="createMode"
          ><button class="btn btn-ghost btn-sm" :disabled="busy" @click="back">Cancel</button
            ><button
              form="production-create"
              type="submit"
            class="btn btn-primary btn-sm"
            :disabled="busy || saving"
          >
            <Save :size="15" />{{ saving ? 'Creating…' : 'Create job' }}
          </button></template
        >
        <template v-else-if="selected">
          <button
            class="btn btn-outline btn-error btn-sm gap-2"
            type="button"
            :disabled="busy"
            title="Delete this production job and its dependent operational records"
            @click="remove"
          >
            <Trash2 :size="14" />Delete job
          </button>
          <button
            v-if="editing"
            class="btn btn-ghost btn-sm"
            type="button"
            :disabled="busy || saving"
            @click="cancelEdit"
          >
            Cancel editing
          </button>
          <button
            v-else-if="canEdit"
            class="btn btn-outline btn-sm gap-2"
            type="button"
            :disabled="busy"
            @click="beginEdit"
          >
            <Pencil :size="14" />Edit job
          </button>
          <button
            v-if="editing"
            form="production-edit"
            class="btn btn-primary btn-sm gap-2"
            type="submit"
            :disabled="busy || saving"
          >
            <Save :size="14" />{{ saving ? 'Saving…' : 'Save changes' }}
          </button>
        </template>
      </WorkspaceHeader>
      <div
        v-if="selected && !createMode"
        class="grid gap-2 border-y border-base-300 bg-base-100 p-3 sm:grid-cols-2 xl:grid-cols-4"
      >
        <div class="min-w-0">
          <span class="block text-xs text-base-content/50">Customer / Order</span
          ><strong class="block truncate text-sm">{{
            selectedOrder?.customerName || 'Walk-in customer'
          }}</strong
          ><span class="text-xs text-base-content/60">{{ jobOrder(selected) }}</span>
        </div>
        <div>
          <span class="block text-xs text-base-content/50">Quantity / Priority</span
          ><strong class="block text-sm">{{ selected.quantity }} {{ selected.quantityUnit }}</strong
          ><span class="text-xs text-base-content/60">{{ selected.priority }}</span>
        </div>
        <div class="min-w-0">
          <span class="block text-xs text-base-content/50">Machine</span
          ><strong class="block truncate text-sm">{{
            props.machines.find((machine) => machine.id === selected?.assignedMachineId)?.name ||
            'Unassigned'
          }}</strong>
        </div>
        <div>
          <span class="block text-xs text-base-content/50">Promised date</span
          ><strong class="block text-sm">{{ date(selectedOrder?.promisedAt || '') }}</strong>
        </div>
      </div>
      <WorkspaceTabs
        v-if="!createMode && !editing"
        class="mt-2"
        :tabs="['Overview', 'Reservations', 'Consumption', 'Outsourcing']"
        :active-tab="tab"
        @change="tab = $event"
      />
    </WorkspaceStickyStack>
    <AppPanel
      v-if="createMode"
      title="Job details"
      subtitle="Service and estimated cost use the saved order item snapshot."
    >
      <form id="production-create" @submit.prevent="create" class="min-w-0 space-y-3">
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
              :model-value="formatQuantityInput(form.quantity)"
              @update:model-value="form.quantity = formatQuantityInput($event)"
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
              ['Urgent', 'High', 'Normal', 'Low'].map((value) => ({
                label: value,
                value,
              }))
            " /><SelectField
            v-model="form.assignedMachineId"
            label="Machine"
            :options="[
              { label: 'Unassigned', value: '' },
              ...props.machines.map((machine) => ({
                label: machine.name,
                value: machine.id,
              })),
            ]" /></FormGrid
        ><FormField class="gap-1"
          ><span>Notes</span><AppTextarea v-model="form.notes" rows="3" />
        </FormField>
      </form>
    </AppPanel>
    <template v-else-if="selected">
      <AppPanel
        v-if="editing"
        title="Edit production job"
        subtitle="Order, service, and quantity stay linked to the saved order item."
      >
        <form id="production-edit" @submit.prevent="update" class="min-w-0 space-y-4">
          <div class="grid gap-3 rounded-box border border-base-300 bg-base-200 p-3 sm:grid-cols-2">
            <div class="min-w-0">
              <span class="block text-xs text-base-content/60">Customer / Order</span>
              <strong class="block truncate text-sm">{{ selectedOrder?.customerName || 'Walk-in customer' }}</strong>
              <span class="text-xs text-base-content/60">{{ jobOrder(selected) }}</span>
            </div>
            <div class="min-w-0">
              <span class="block text-xs text-base-content/60">Service / Quantity</span>
              <strong class="block truncate text-sm">{{ selected.serviceName }}</strong>
              <span class="text-xs text-base-content/60">{{ selected.quantity }} {{ selected.quantityUnit }}</span>
            </div>
          </div>
          <FormGrid>
            <SelectField
              v-model="form.assignedMachineId"
              label="Machine"
              :options="[
                { label: 'Unassigned', value: '' },
                ...props.machines.map((machine) => ({ label: machine.name, value: machine.id })),
              ]"
            />
            <SelectField
              v-model="form.priority"
              label="Priority"
              :options="['Urgent', 'High', 'Normal', 'Low'].map((value) => ({ label: value, value }))"
            />
          </FormGrid>
          <FormField label="Planned date">
            <JalaliDatePicker v-model="form.plannedAt" placeholder="Set planned date" />
          </FormField>
          <FormField label="Notes">
            <AppTextarea v-model="form.notes" rows="4" placeholder="Add production notes" />
          </FormField>
        </form>
      </AppPanel>
      <template v-else>
      <AppPanel v-show="tab === 'Overview'" title="Job workflow" subtitle="Move the job through its available production stages.">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2 text-sm">
            <span class="text-xs text-base-content/60">Current status</span>
            <StatusBadge :label="selected.status" :tone="statusTone(selected.status)" />
          </div>
          <div v-if="statusActions.length" class="flex flex-wrap items-center gap-2" role="group" aria-label="Available status actions">
            <span class="text-xs text-base-content/60">Move to</span>
            <button
              v-for="action in statusActions"
              :key="action.status"
              type="button"
              class="btn btn-sm gap-2"
              :class="
                action.kind === 'primary'
                  ? 'btn-primary'
                  : action.kind === 'danger'
                    ? 'btn-outline btn-error'
                    : 'btn-outline'
              "
              :disabled="busy"
              @click="changeStatus(action.status)"
            >
              <Play v-if="action.status === 'In Progress'" :size="14" aria-hidden="true" />
              <Pause v-else-if="action.status === 'Paused'" :size="14" aria-hidden="true" />
              <CheckCircle2 v-else-if="action.status === 'Completed'" :size="14" aria-hidden="true" />
              <XCircle v-else-if="action.status === 'Cancelled' || action.status === 'Failed'" :size="14" aria-hidden="true" />
              <CheckCircle2 v-else :size="14" aria-hidden="true" />
              {{ action.label }}
            </button>
          </div>
          <span v-else class="text-xs text-base-content/60">No further status transitions are available.</span>
        </div>
      </AppPanel>
      <section v-show="tab === 'Overview'" class="grid min-w-0 gap-4 xl:grid-cols-2">
        <AppPanel title="Cost and schedule">
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
        </AppPanel>
        <AppPanel title="Job context"
          ><dl class="space-y-3 text-sm">
            <div>
              <dt class="text-xs text-base-content/60">Created</dt>
              <dd>{{ date(selected.createdAt) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-base-content/60">Planned</dt>
              <dd>{{ date(selected.plannedAt) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-base-content/60">Notes</dt>
              <dd class="whitespace-pre-wrap break-words">
                {{ selected.notes || 'No notes.' }}
              </dd>
            </div>
          </dl></AppPanel
        >
      </section>
      <AppPanel v-show="tab === 'Reservations'" title="Reservations">
        <LoadingState v-if="reservationsLoading" label="Loading reservations…" />
        <p v-else-if="!reservations.length" class="text-sm text-base-content/60">
          No reservations for this job.
        </p>
        <div
          v-for="r in reservations"
          :key="r.id"
          class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
        >
          <span
            >{{ props.materials.find((m) => m.id === r.materialId)?.name ?? r.materialId }} ·
            {{ r.quantity }}</span
          ><button
            class="btn btn-ghost"
            v-if="r.status === 'active'"
            @click="release(r)"
            :disabled="busy"
          >
            Release</button
          ><small v-else class="block text-xs leading-5 text-base-content/60">{{ r.status }}</small>
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
            ]" /><FormField label="Qty"
            ><AppInput
              class="input w-full min-w-0"
              :model-value="formatQuantityInput(reservationQuantity)"
              @update:model-value="reservationQuantity = formatQuantityInput($event)"
              placeholder="Qty"
              inputmode="decimal" /></FormField></FormGrid
        ><button class="btn btn-ghost" @click="reserve" :disabled="busy">
          <RotateCcw :size="15" />Reserve stock
        </button>
      </AppPanel>
      <AppPanel v-show="tab === 'Consumption'" title="Actual consumption & waste">
        <FormGrid :columns="3"
          ><SelectField
            v-model="consumptionMaterial"
            label="Material"
            :options="[
              { label: 'Material', value: '' },
              ...props.materials.map((material) => ({
                label: material.name,
                value: material.id,
              })),
            ]" /><FormField label="Consumed"
            ><AppInput
              class="input w-full min-w-0"
              :model-value="formatQuantityInput(consumedQuantity)"
              @update:model-value="consumedQuantity = formatQuantityInput($event)"
              placeholder="Consumed"
              inputmode="decimal" /></FormField
          ><FormField label="Waste"
            ><AppInput
              class="input w-full min-w-0"
              :model-value="formatQuantityInput(wasteQuantity)"
              @update:model-value="wasteQuantity = formatQuantityInput($event)"
              placeholder="Waste"
              inputmode="decimal" /></FormField></FormGrid
        ><FormField label="Posting reference (optional)"
          ><AppInput
            class="input w-full min-w-0"
            v-model="consumptionKey"
            placeholder="Posting reference (optional)" /></FormField
        ><button class="btn btn-primary" @click="consume" :disabled="busy">
          Record consumption
        </button>
      </AppPanel>
      <AppPanel v-show="tab === 'Outsourcing'" title="Outsourced production">
        <FormGrid :columns="3"
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
              :money="props.currencyUnit"
              placeholder="0"
              inputmode="numeric" /></FormField
          ><FormField class="gap-1"
            ><span>Actual ({{ props.currencyUnit }})</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="outsourceCost"
              :money="props.currencyUnit"
              placeholder="0"
              inputmode="numeric" /></FormField></FormGrid
        ><FormField label="Outsourced scope / return details"
          ><AppInput
            class="input w-full min-w-0"
            v-model="outsourceDescription"
            placeholder="Outsourced scope / return details" /></FormField
        ><button class="btn btn-primary" @click="saveOutsource" :disabled="busy">
          Save outsourcing
        </button>
      </AppPanel>
      </template>
    </template>
  </div>
</template>
