<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { CheckCheck, CheckCircle2, ClipboardList, Factory, Package, PackageOpen, Pause, Play, RotateCcw, UserRound, XCircle } from 'lucide-vue-next'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import AppInput from '../../components/ui/AppInput.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import FormField from '../../components/ui/FormField.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import ProductionCostBreakdown from './ProductionCostBreakdown.vue'
import { formatMoney } from '../../utils/currency'
import { formatQuantityInput, parseQuantityInput } from '../../utils/quantity'
import type { OrderRecord } from '../../api/orders'
import type { useProductionWorkspace } from './useProductionWorkspace'

const props = defineProps<{
  workspace: ReturnType<typeof useProductionWorkspace>
  currencyUnit: import('../../utils/currency').CurrencyUnit
  orders: OrderRecord[]
  materials: any[]
  machines: any[]
  suppliers: any[]
}>()

const emit = defineEmits<{ back: [] }>()
const {
  materialPlans, consumptions, stockMaterials, financialAccounts, materialsError, editingConsumptionId,
  outsourceQuantity, outsourceAccount, outsourceTotal, reservedConsumptionQuantity, canConsume, editableJob,
  suggestConsumption, updateConsumptionSlider, editConsumption, correctConsumption, updateReservation,
  busy,
  selected,
  selectedOrder,
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
  changeStatus,
  reserve,
  release,
  consume,
  consumeReservation,
  saveOutsource,
  jobOrder,
  statusTone,
  date,
} = props.workspace

const activeStep = ref(1)
const reservationDrafts = ref<Record<string,string>>({})
const usageExpanded = ref(false)
watch(reservations, rows => { reservationDrafts.value=Object.fromEntries(rows.map(r=>[r.id,r.quantity])) },{immediate:true})
const consumptionPercentage = computed(()=>reservedConsumptionQuantity.value>0 ? Math.min(100,Math.max(0,Number(parseQuantityInput(consumedQuantity.value || '0'))/reservedConsumptionQuantity.value*100)) : 0)
const activeReservations = computed(()=>reservations.value.filter(r=>r.status==='active'))
const selectedMaterial = computed(()=>stockMaterials.value.find(m=>m.id===consumptionMaterial.value))
function useReservation(r: typeof reservations.value[number]) { consumeReservation(r); usageExpanded.value=true }
function openConsumptionEdit(r: typeof consumptions.value[number]) { editConsumption(r); usageExpanded.value=true }

const steps = [
  { number: 1, title: 'Overview', description: 'Status, order, cost, and schedule' },
  { number: 2, title: 'Materials', description: 'Plan stock, adjust usage, and record waste' },
  { number: 3, title: 'Outsourcing', description: 'Track external production' },
]
const currentStep = computed(() => steps.find((step) => step.number === activeStep.value) ?? steps[0])
const machineName = computed(() => props.machines.find((machine) => machine.id === selected.value?.assignedMachineId)?.name || 'Unassigned')
const customerName = computed(() => selectedOrder.value?.customerName || 'Walk-in customer')
const statusActions = computed(() => {
  if (!selected.value) return []
  const statuses = ['Pending', 'Ready', 'In Progress', 'Paused', 'Completed', 'Cancelled', 'Failed'] as const
  return statuses.filter((status) => status !== selected.value?.status && (status !== 'Completed' || canConsume.value || (['In Progress','Paused'].includes(selected.value!.status)))).map((status) => ({
    status,
    label: status === 'In Progress'
      ? selected.value?.status === 'Paused' || selected.value?.status === 'Failed' ? 'Resume production' : 'Start production'
      : status === 'Completed' ? 'Complete job'
        : status === 'Cancelled' ? 'Cancel job'
          : status === 'Failed' ? 'Mark failed'
            : status === 'Paused' ? 'Pause production'
              : `Set ${status}`,
    kind: status === 'In Progress' || status === 'Completed' ? 'primary' : status === 'Cancelled' || status === 'Failed' ? 'danger' : 'secondary',
  }))
})

watch(() => selected.value?.id, () => { activeStep.value = 1 })

function money(value: number) {
  return formatMoney(value || 0, props.currencyUnit)
}

function stepClass(number: number) {
  if (number === activeStep.value) return 'wizard-step-active'
  if (number < activeStep.value) return 'wizard-step-complete'
  return 'wizard-step-idle'
}

function next() {
  if (activeStep.value < steps.length) activeStep.value += 1
}

function previous() {
  if (activeStep.value > 1) activeStep.value -= 1
}

</script>

<template>
  <div v-if="selected" class="service-wizard w-full flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Production management wizard">
    <header class="service-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0"><h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">{{ selected.jobNumber }}</h1><p class="mt-1 truncate text-xs leading-4 text-base-content/65">Manage {{ selected.serviceName }} for {{ jobOrder(selected) }}.</p><WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Production' }, { label: selected.jobNumber, current: true }]" @navigate="emit('back')" /></div>
      <div class="flex shrink-0 items-center gap-2"><button class="btn btn-success btn-sm" type="button" :disabled="busy" @click="emit('back')">Done</button></div>
    </header>

    <div class="service-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="service-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200"><nav aria-label="Production management steps" class="service-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible"><button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" @click="activeStep = step.number"><span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span><span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ step.title }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ step.description }}</small></span></button></nav></aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="service-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
          <section v-if="activeStep === 1" class="min-w-0 space-y-4"><div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><ClipboardList :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Production overview</h2><p class="text-sm text-base-content/60">Manage the job created from this confirmed order item.</p></div></div><div class="rounded-box border border-base-300 bg-base-100/45 p-4"><div class="flex flex-wrap items-center gap-3"><span class="text-xs text-base-content/60">Current status</span><StatusBadge :label="selected.status" :tone="statusTone(selected.status)" /></div><div v-if="statusActions.length" class="mt-3 flex flex-wrap items-center gap-2" role="group" aria-label="Available status actions"><span class="text-xs text-base-content/60">Move to</span><button v-for="action in statusActions" :key="action.status" type="button" class="btn btn-sm gap-2" :class="action.kind === 'primary' ? 'btn-primary' : action.kind === 'danger' ? 'btn-outline btn-error' : 'btn-outline'" :disabled="busy" @click="changeStatus(action.status)"><Play v-if="action.status === 'In Progress'" :size="14" aria-hidden="true" /><Pause v-else-if="action.status === 'Paused'" :size="14" aria-hidden="true" /><CheckCircle2 v-else-if="action.status === 'Completed'" :size="14" aria-hidden="true" /><XCircle v-else-if="action.status === 'Cancelled' || action.status === 'Failed'" :size="14" aria-hidden="true" /><CheckCircle2 v-else :size="14" aria-hidden="true" />{{ action.label }}</button></div><EmptyState v-else compact title="No further status transitions" description="This job is at its current terminal or completed state."><template #icon><CheckCircle2 :size="21" aria-hidden="true" /></template></EmptyState></div><div class="grid min-w-0 grid-cols-1 gap-4"><ProductionCostBreakdown :job="selected" :currency-unit="currencyUnit" /><div class="rounded-box border border-base-300 bg-base-100/45 p-4"><h3 class="text-sm font-semibold">Schedule</h3><dl class="mt-3 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-xs text-base-content/60">Started</dt><dd>{{ date(selected.startedAt) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-xs text-base-content/60">Completed</dt><dd>{{ date(selected.completedAt) }}</dd></div></dl></div></div><div class="rounded-box border border-base-300 bg-base-100/45 p-4"><h3 class="text-sm font-semibold">Order link</h3><p class="mt-2 text-sm leading-6 text-base-content/70">This production job is linked to <strong>{{ selected.serviceName }}</strong> on {{ jobOrder(selected) }}. Service, quantity, and customer context are managed from the order.</p></div></section>

          <section v-else-if="activeStep === 2" class="min-w-0 space-y-4">
            <header><h2 class="text-lg font-semibold">Order materials</h2><p class="mt-1 text-sm text-base-content/60">Linked to {{ selected.serviceName }} · {{ selected.quantity }} {{ selected.quantityUnit }}. Order edits update this plan; your adjustments are retained.</p></header>
            <p class="rounded-box border border-info/25 bg-info/5 p-3 text-sm leading-6">Review the total material allowance here. Available stock is reserved automatically. Completing the job records the remaining planned materials at inventory cost. You can record usage or waste early below; it will not be counted twice. Include waste in the total allowance.</p>
            <p v-if="materialsError" class="rounded-box border border-error/30 p-3 text-sm text-error" role="alert">{{ materialsError }}</p>
            <LoadingState v-if="reservationsLoading" label="Updating material quantities…" />
            <template v-else>
              <div v-for="plan in materialPlans" :key="plan.materialId" class="space-y-3 border-t border-base-300 pt-4">
                <div class="flex flex-wrap items-center justify-between gap-2"><h3 class="text-sm font-semibold">{{ plan.name }}</h3><span class="text-xs text-base-content/55">{{ plan.unit }}</span></div>
                <dl class="grid grid-cols-4 gap-2 text-xs">
                  <div><dt class="text-base-content/55">Order needs</dt><dd class="mt-1 text-sm tabular-nums">{{ plan.required }}</dd></div>
                  <div><dt class="text-base-content/55">Reserved</dt><dd class="mt-1 text-sm tabular-nums text-primary">{{ plan.reserved }}</dd></div>
                  <div><dt class="text-base-content/55">Used</dt><dd class="mt-1 text-sm tabular-nums">{{ plan.used }}</dd></div>
                  <div><dt class="text-base-content/55">Free stock</dt><dd class="mt-1 text-sm tabular-nums">{{ plan.available }}</dd></div>
                </dl>
                <p v-if="Number(plan.shortage)>0" class="text-xs text-warning" role="status">{{ plan.shortage }} {{ plan.unit }} still needed. Only available stock is reserved.</p>
              </div>
              <p v-if="!materialPlans.length && !materialsError" class="border-t border-base-300 pt-4 text-sm text-base-content/60">This service has no linked material requirement. Add material below if needed; reserved quantities will be recorded on completion.</p>
              <div class="border-t border-base-300 pt-4">
                <h3 class="text-sm font-semibold">Stock reserved for this job</h3>
                <p class="mt-1 text-xs text-base-content/55">Blocked from other jobs until used or released.</p>
                <div data-enter-scope v-for="reservation in activeReservations" :key="reservation.id" class="flex flex-wrap items-end gap-2 border-b border-base-300/60 py-3">
                  <FormField class="min-w-40 flex-1" :label="stockMaterials.find(m=>m.id===reservation.materialId)?.name || reservation.materialId"><AppInput v-model="reservationDrafts[reservation.id]" inputmode="decimal" :disabled="busy || !editableJob" /></FormField>
                  <div class="mb-1 flex gap-1"><button class="btn btn-outline btn-sm" :disabled="busy || !editableJob" data-enter-submit @click="updateReservation(reservation,reservationDrafts[reservation.id] || '0')">Update</button><button class="btn btn-primary btn-sm" :disabled="busy || !editableJob" @click="useReservation(reservation)">Use material</button><button class="btn btn-ghost btn-sm" :disabled="busy || !editableJob" @click="release(reservation)">Release</button></div>
                </div>
                <p v-if="!activeReservations.length" class="mt-3 text-xs text-base-content/55">No stock is currently reserved.</p>
              </div>
              <details data-enter-scope class="border-t border-base-300 pt-3">
                <summary class="cursor-pointer text-sm font-medium">Reserve additional material</summary>
                <FormGrid class="mt-3"><SelectField v-model="reservationMaterial" label="Material" :options="[{label:'Select material',value:''},...stockMaterials.map(m=>({label:m.name+' · free '+m.availableStock,value:m.id}))]" /><FormField label="Quantity"><AppInput v-model="reservationQuantity" inputmode="decimal" placeholder="Quantity" /></FormField></FormGrid>
                <button class="btn btn-outline btn-sm mt-3" :disabled="busy || !editableJob" data-enter-submit @click="reserve">Add reservation</button>
              </details>
            </template>
            <details data-enter-scope :open="usageExpanded" class="min-w-0 space-y-4 border-t border-base-300 pt-4" @toggle="usageExpanded=($event.target as HTMLDetailsElement).open">
            <summary class="cursor-pointer text-sm font-semibold">{{ editingConsumptionId ? 'Correct material usage' : 'Record early usage or waste (optional)' }}</summary>
            <p class="text-sm text-base-content/60">Use this job’s reservations first, then free stock. Enter only the additional quantity used in this entry.</p>
            <p v-if="!canConsume" class="text-sm text-warning">Start or resume in-house production from Overview before posting material usage.</p>
            <SelectField v-model="consumptionMaterial" label="Material" :options="[{label:'Select material',value:''},...stockMaterials.map(m=>({label:m.name,value:m.id}))]" @update:model-value="suggestConsumption" />
            <div v-if="selectedMaterial" class="flex flex-wrap gap-x-5 gap-y-1 text-xs text-base-content/60"><span>Reserved: {{ reservedConsumptionQuantity }} {{ selectedMaterial.consumptionUnit }}</span><span>Free stock: {{ selectedMaterial.availableStock }} {{ selectedMaterial.consumptionUnit }}</span></div>
            <FormGrid>
              <FormField label="Used quantity">
                <div class="relative"><AppInput v-model="consumedQuantity" class="pe-14" inputmode="decimal" /><button type="button" class="absolute inset-y-1 end-1 rounded px-2 text-xs font-medium text-primary hover:bg-primary/10" :disabled="busy || !canConsume" @click="consumedQuantity=String(reservedConsumptionQuantity)">Max</button></div>
              </FormField>
              <FormField label="Waste"><AppInput v-model="wasteQuantity" inputmode="decimal" /></FormField>
            </FormGrid>
            <div class="payment-slider my-3 w-full overflow-visible">
              <input class="range range-primary range-sm w-full" type="range" min="0" max="100" step="5" :value="consumptionPercentage" :disabled="busy || !canConsume || !reservedConsumptionQuantity" aria-label="Reserved material usage percentage" :aria-valuetext="Math.round(consumptionPercentage)+'% of reserved material'" @input="updateConsumptionSlider" />
              <div class="relative mx-2.5 mt-1 h-3 text-[10px] leading-3 text-base-content/40" aria-hidden="true"><span v-for="mark in [0,25,50,75,100]" :key="mark" class="absolute -translate-x-1/2" :style="{left:mark+'%'}">|</span></div>
              <div class="relative mx-2.5 mt-1 h-4 text-xs leading-4 text-base-content/50" aria-hidden="true"><span v-for="mark in [0,25,50,75,100]" :key="mark" class="absolute -translate-x-1/2" :style="{left:mark+'%'}">{{ mark }}</span></div>
            </div>
            <p class="text-xs text-base-content/55">The slider and Max use the reserved quantity. Type a larger quantity to use additional free stock.</p>
            <div class="flex flex-wrap gap-2"><button class="btn btn-primary btn-sm" :disabled="busy || !canConsume || !consumptionMaterial" data-enter-submit @click="consume">{{ editingConsumptionId ? 'Save correction' : 'Record usage' }}</button><button class="btn btn-ghost btn-sm" :disabled="busy" @click="suggestConsumption">{{ editingConsumptionId ? 'Cancel correction' : 'Use reserved quantity' }}</button></div>
            </details>
            <div class="border-t border-base-300 pt-4">
              <h3 class="text-sm font-semibold">Material usage history</h3>
              <p v-if="!consumptions.length" class="mt-3 text-sm text-base-content/55">No material has been used yet.</p>
              <div v-for="record in consumptions.slice().reverse()" :key="record.id" class="flex flex-wrap items-center justify-between gap-2 border-b border-base-300/60 py-3" :class="record.reversed ? 'opacity-50' : ''">
                <div class="min-w-0"><strong class="block text-sm">{{ stockMaterials.find(m=>m.id===record.materialId)?.name || record.materialId }}</strong><p class="mt-1 text-xs text-base-content/60">{{ record.consumedQuantity }} used · {{ record.wasteQuantity }} waste · {{ money(record.materialCostRial+record.wasteCostRial) }} <span v-if="record.reversed">· Returned / corrected</span></p></div>
                <div v-if="!record.reversed" class="flex gap-1"><button class="btn btn-ghost btn-sm" :disabled="busy || !canConsume" @click="openConsumptionEdit(record)">Edit</button><button class="btn btn-ghost btn-sm" :disabled="busy || !canConsume" @click="correctConsumption(record)">Return to stock</button></div>
              </div>
            </div>
          </section>

          <section v-else data-enter-scope class="min-w-0 space-y-4">
            <header><h2 class="text-lg font-semibold">Outsource production</h2><p class="mt-1 text-sm text-base-content/60">Choose how many of the {{ selected.quantity }} {{ selected.quantityUnit }} to outsource. Unit cost starts from the order estimate.</p></header>
            <FormGrid><FormField label="Outsourced quantity"><AppInput v-model="outsourceQuantity" inputmode="decimal" :disabled="busy || !editableJob" /></FormField><FormField :label="'Cost per '+selected.quantityUnit"><AppInput v-model="outsourceCost" :money="props.currencyUnit" inputmode="decimal" :disabled="busy || !editableJob" /></FormField></FormGrid>
            <div class="flex flex-wrap justify-between gap-3 border-y border-base-300 py-4"><div><span class="block text-xs text-base-content/55">In-house quantity</span><strong class="mt-1 block text-base tabular-nums">{{ Math.max(0,Number(selected.quantity)-Number(parseQuantityInput(outsourceQuantity || '0'))) }} {{ selected.quantityUnit }}</strong></div><div class="text-end"><span class="block text-xs text-base-content/55">Outsourcing expense</span><strong class="mt-1 block text-lg tabular-nums text-primary">{{ money(outsourceTotal) }}</strong></div></div>
            <FormGrid><SelectField v-model="outsourceSupplier" label="Supplier" :options="[{label:'Select supplier',value:''},...props.suppliers.map(s=>({label:s.name,value:s.id}))]" /><SelectField v-model="outsourceAccount" label="Expense payment account" :options="[{label:'Select account',value:''},...financialAccounts.map(a=>({label:a.name,value:a.id}))]" /></FormGrid>
            <FormField label="Scope / notes"><AppInput v-model="outsourceDescription" placeholder="External production details" /></FormField>
            <p class="text-sm leading-6 text-base-content/65">The outsourced share releases reserved stock and returns its share of used material. The expense replaces that material cost in the order margin. Further changes update the same expense.</p>
            <p class="text-xs text-base-content/55">Reducing outsourcing restores the in-house material plan. Completing the job records the remaining materials automatically.</p>
            <button class="btn btn-primary btn-sm" :disabled="busy || !editableJob" data-enter-submit @click="saveOutsource">Apply outsourcing</button>
            <p v-if="Number(selected.outsourceQuantity)>0" class="text-xs text-success">Saved: {{ selected.outsourceQuantity }} {{ selected.quantityUnit }} × {{ money(selected.outsourceUnitCostRial) }} = {{ money(selected.actualOutsourcedCostRial) }}</p>
          </section>
        </section>

        <aside class="service-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto"><div class="space-y-4"><div class="relative min-h-48 overflow-hidden rounded-box bg-base-100 p-5"><div class="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-transparent" aria-hidden="true"></div><div class="relative flex min-h-36 flex-col justify-end"><span class="mb-auto grid size-11 place-items-center rounded-box border border-base-300 bg-base-200/80 text-primary"><Factory :size="24" aria-hidden="true" /></span><div class="mt-6 flex items-end justify-between gap-3"><div class="min-w-0"><h2 class="truncate text-lg font-semibold">{{ selected.jobNumber }}</h2><p class="mt-1 truncate text-xs text-base-content/60">{{ selected.serviceName }}</p></div><StatusBadge :label="selected.status" :tone="statusTone(selected.status)" /></div></div></div><div class="divide-y divide-base-300 rounded-box border border-base-300 bg-base-100/35"><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Customer</span><strong class="max-w-[10rem] truncate text-end">{{ customerName }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Quantity</span><strong>{{ selected.quantity }} {{ selected.quantityUnit }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Machine</span><strong class="max-w-[10rem] truncate text-end">{{ machineName }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Step</span><strong>{{ currentStep.title }}</strong></div></div><div class="flex gap-2 border-t border-base-300 pt-3 text-sm"><UserRound class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">{{ currentStep.description }}. This job remains linked to the saved order item.</p></div></div></aside>
      </div>
    </div>

    <footer class="flex min-w-0 items-center justify-between gap-3 border-t border-base-300 px-1 pt-3"><button class="btn btn-ghost btn-sm" type="button" :disabled="activeStep === 1 || busy" @click="previous">Back</button><span class="text-xs text-base-content/55">Step {{ activeStep }} of {{ steps.length }}</span><button v-if="activeStep < steps.length" class="btn btn-primary btn-sm" type="button" @click="next">Continue</button><button v-else class="btn btn-primary btn-sm" type="button" @click="activeStep = 1">Back to overview</button></footer>
  </div>
</template>
