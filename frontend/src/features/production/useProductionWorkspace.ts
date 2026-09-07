import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
import { watch } from 'vue'
import { computed, onMounted, ref } from 'vue';
import { productionApi, type ProductionJobRecord, type ReservationRecord } from '../../api/production';
import type { OrderRecord } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';

export function useProductionWorkspace(props:{
  currencyUnit: CurrencyUnit;
  orders: OrderRecord[];
  materials: any[];
  machines: any[];
  suppliers: any[];
},emit:(event:'notify',message:string)=>void){
const {busy,runAction}=useWorkspaceActions()
const jobs = ref<ProductionJobRecord[]>([]);
const selectedId = ref<string | null>(null);
const statusFilter = ref('All');
const searchQuery = ref('');
const loading = ref(false);
const error = ref('');
const createMode = ref(false);
const saving = ref(false);
const form = ref({
  orderId: '',
  orderItemId: '',
  quantity: '',
  quantityUnit: '',
  assignedMachineId: '',
  priority: 'Normal',
  notes: '',
});
const reservations = ref<ReservationRecord[]>([]);
const reservationMaterial = ref('');
const reservationQuantity = ref('');
const consumedQuantity = ref('');
const wasteQuantity = ref('');
const consumptionMaterial = ref('');
const consumptionKey = ref('');
const consumptionNotes = ref('');
const outsourceSupplier = ref('');
const outsourceDescription = ref('');
const outsourceCost = ref('0');
const outsourceQuotedCost = ref('0');
const selected = computed(() => jobs.value.find((j) => j.id === selectedId.value) ?? null);
const selectedOrder = computed(
  () =>
    props.orders.find(
      (o) => o.id === (createMode.value ? form.value.orderId : selected.value?.orderId),
    ) ?? null,
);
const selectedItem = computed(
  () =>
    selectedOrder.value?.items.find(
      (i) => i.id === (createMode.value ? form.value.orderItemId : selected.value?.orderItemId),
    ) ?? null,
);
const visibleJobs = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  return jobs.value.filter(
    (j) =>
      !q ||
      [
        j.jobNumber,
        j.serviceName,
        jobOrder(j),
        props.orders.find((o) => o.id === j.orderId)?.customerName ?? '',
      ].some((v) => v.toLowerCase().includes(q)),
  );
});
onMounted(load);
watch(statusFilter, load);
async function load() {
  loading.value = true;
  error.value = '';
  try {
    jobs.value = await productionApi.list(statusFilter.value);
    if (selectedId.value && !jobs.value.some((j) => j.id === selectedId.value))
      selectedId.value = null;
    if (!selectedId.value && jobs.value.length) select(jobs.value[0].id);
  } catch (e) {
    error.value = message(e, 'Production queue could not be loaded.');
  } finally {
    loading.value = false;
  }
}
async function select(id: string) {
  selectedId.value = id;
  createMode.value = false;
  try {
    reservations.value = await productionApi.reservations('', id, '');
  } catch (e) {
reportError(e);
    error.value = message(e, 'Reservations could not be loaded.');
  }
}
function beginCreate() {
  createMode.value = true;
  selectedId.value = null;
  reservations.value = [];
  form.value = {
    orderId: '',
    orderItemId: '',
    quantity: '',
    quantityUnit: '',
    assignedMachineId: '',
    priority: 'Normal',
    notes: '',
  };
}
function chooseOrder() {
  const o = props.orders.find((x) => x.id === form.value.orderId);
  const i = o?.items[0];
  form.value.orderItemId = i?.id ?? '';
  form.value.quantity = i?.quantity ?? '';
  form.value.quantityUnit = i?.quantityUnit ?? '';
}
async function create() {
return runAction(async () => {
  if (!form.value.orderId || !form.value.orderItemId || !form.value.quantity) return;
  saving.value = true;
  try {
    const j = await productionApi.create({ ...form.value, plannedAt: null });
    jobs.value.unshift(j);
    createMode.value = false;
    await select(j.id);
    emit('notify', 'Production job created from the saved order item.');
  } catch (e) {
reportError(e);
    error.value = message(e, 'Production job could not be created.');
  } finally {
    saving.value = false;
  }

});
}
async function changeStatus(status: string) {
return runAction(async () => {
  if (!selected.value) return;
  try {
    const j = await productionApi.status(selected.value.id, status);
    replace(j);
    await select(j.id);
    emit('notify', `Job moved to ${status}.`);
  } catch (e) {
reportError(e);
    error.value = message(e, 'Production status could not be changed.');
  }

});
}
async function remove() {
return runAction(async () => {
  if (!selected.value) return;
  try {
    await productionApi.delete(selected.value.id);
    jobs.value = jobs.value.filter((j) => j.id !== selected.value!.id);
    selectedId.value = null;
    reservations.value = [];
    emit('notify', 'Unposted production draft deleted.');
  } catch (e) {
reportError(e);
    error.value = message(e, 'Production history cannot be deleted.');
  }

});
}
async function reserve() {
return runAction(async () => {
  if (!selected.value || !reservationMaterial.value || !reservationQuantity.value) return;
  try {
    await productionApi.reserve({
      materialId: reservationMaterial.value,
      orderId: selected.value.orderId,
      orderItemId: selected.value.orderItemId,
      productionJobId: selected.value.id,
      quantity: reservationQuantity.value,
    });
    reservationQuantity.value = '';
    reservations.value = await productionApi.reservations('', selected.value.id, '');
    await load();
    emit('notify', 'Inventory reserved without creating a movement.');
  } catch (e) {
reportError(e);
    error.value = message(e, 'Reservation exceeds available stock or is invalid.');
  }

});
}
async function release(r: ReservationRecord) {
return runAction(async () => {
  try {
    await productionApi.releaseReservation(r.id);
    if (selected.value)
      reservations.value = await productionApi.reservations('', selected.value.id, '');
  } catch (e) {
reportError(e);
    error.value = message(e, 'Reservation could not be released.');
  }

});
}
async function consume() {
return runAction(async () => {
  if (
    !selected.value ||
    !consumptionMaterial.value ||
    (!consumedQuantity.value && !wasteQuantity.value)
  )
    return;
  try {
    await productionApi.consume(selected.value.id, {
      materialId: consumptionMaterial.value,
      consumedQuantity: consumedQuantity.value || '0',
      wasteQuantity: wasteQuantity.value || '0',
      idempotencyKey: consumptionKey.value || `ui-${Date.now()}`,
      notes: consumptionNotes.value,
    });
    consumedQuantity.value = '';
    wasteQuantity.value = '';
    consumptionKey.value = '';
    consumptionNotes.value = '';
    await load();
    if (selected.value) await select(selected.value.id);
    emit('notify', 'Consumption posted as immutable inventory movement(s).');
  } catch (e) {
reportError(e);
    error.value = message(e, 'Consumption could not be posted.');
  }

});
}
async function saveOutsource() {
return runAction(async () => {
  if (!selected.value) return;
  const cost = parseMoneyInput(outsourceCost.value, props.currencyUnit);
  const quoted = parseMoneyInput(outsourceQuotedCost.value, props.currencyUnit);
  if (cost === null || quoted === null) {
    error.value = 'Enter whole outsourced costs.';
    return;
  }
  try {
    const j = await productionApi.outsource(selected.value.id, {
      supplierId: outsourceSupplier.value,
      description: outsourceDescription.value,
      sentAt: '',
      expectedReturnAt: '',
      receivedAt: '',
      notes: '',
      quotedCostRial: quoted,
      actualCostRial: cost,
    });
    replace(j);
    emit('notify', 'Outsourced production metadata saved.');
  } catch (e) {
reportError(e);
    error.value = message(e, 'Outsourcing metadata could not be saved.');
  }

});
}
function replace(j: ProductionJobRecord) {
  const i = jobs.value.findIndex((x) => x.id === j.id);
  if (i >= 0) jobs.value.splice(i, 1, j);
}
function jobOrder(j: ProductionJobRecord) {
  return props.orders.find((o) => o.id === j.orderId)?.orderNumber ?? j.orderId;
}
function jobContext(j: ProductionJobRecord) {
  const o = props.orders.find((x) => x.id === j.orderId);
  return `${o?.promisedAt ? `Due ${date(o.promisedAt)}` : 'No promised date'} · ${props.machines.find((m) => m.id === j.assignedMachineId)?.name ?? 'Unassigned'}`;
}
function statusTone(s: string) {
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
function message(e: unknown, fallback: string) {
  return e instanceof Error && e.message ? e.message : typeof e === 'string' ? e : fallback;
}
function date(v: string) {
  try {
    return v ? formatDateTime(v) : '—';
  } catch {
    return '—';
  }
}
return {busy,runAction,jobs,selectedId,statusFilter,searchQuery,loading,error,createMode,saving,form,reservations,reservationMaterial,reservationQuantity,consumedQuantity,wasteQuantity,consumptionMaterial,consumptionKey,consumptionNotes,outsourceSupplier,outsourceDescription,outsourceCost,outsourceQuotedCost,selected,selectedOrder,selectedItem,visibleJobs,load,select,beginCreate,chooseOrder,create,changeStatus,remove,reserve,release,consume,saveOutsource,replace,jobOrder,jobContext,statusTone,message,date}
}