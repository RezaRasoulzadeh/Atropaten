import { useWorkspaceActions } from '../../composables/useWorkspaceActions';
import { confirmAction, useToast } from '../../ui/feedback';
import { computed, onMounted, ref, watch } from 'vue';
import {
  productionApi,
  type ProductionJobRecord,
  type ReservationRecord,
} from '../../api/production';
import type { OrderRecord } from '../../api/orders';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoneyInput, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import { parseQuantityInput } from '../../utils/quantity';

export function useProductionWorkspace(
  props: {
    currencyUnit: CurrencyUnit;
    orders: OrderRecord[];
    materials: any[];
    machines: any[];
    suppliers: any[];
  },
  emit: (event: 'notify', message: string) => void,
) {
  const { busy, runAction } = useWorkspaceActions();
  const toast = useToast();
  const jobs = ref<ProductionJobRecord[]>([]);
  const selectedId = ref<string | null>(null);
  const statusFilter = ref('All');
  const searchQuery = ref('');
  const loading = ref(false);
  const createMode = ref(false);
  const editing = ref(false);
  const saving = ref(false);
  const form = ref({
    orderId: '',
    orderItemId: '',
    quantity: '',
    quantityUnit: '',
    assignedMachineId: '',
    priority: 'Normal',
    notes: '',
    plannedAt: null as string | null,
  });
  const reservations = ref<ReservationRecord[]>([]);
  const reservationsLoading = ref(false);
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
        (statusFilter.value === 'All' || j.status === statusFilter.value) &&
        (!q ||
          [
            j.jobNumber,
            j.serviceName,
            jobOrder(j),
            props.orders.find((o) => o.id === j.orderId)?.customerName ?? '',
          ].some((v) => v.toLowerCase().includes(q))),
    );
  });
  watch(
    () => props.currencyUnit,
    (unit, previousUnit) => {
      for (const input of [outsourceCost, outsourceQuotedCost]) {
        const amount = parseMoneyInput(input.value, previousUnit);
        if (amount !== null) input.value = formatMoneyInput(amount, unit);
      }
    },
  );
  onMounted(load);
  async function load() {
    loading.value = true;
    try {
      jobs.value = await productionApi.list('All');
      if (selectedId.value && !jobs.value.some((j) => j.id === selectedId.value))
        selectedId.value = null;
    } catch (e) {
      toast.error(message(e, 'Production queue could not be loaded.'));
    } finally {
      loading.value = false;
    }
  }
  async function select(id: string) {
    const differentJob = selectedId.value !== id;
    selectedId.value = id;
    createMode.value = false;
    editing.value = false;
    reservations.value = [];
    reservationsLoading.value = true;
    if (differentJob) {
      reservationMaterial.value = reservationQuantity.value = '';
      consumptionMaterial.value = consumedQuantity.value = wasteQuantity.value = '';
      consumptionKey.value = consumptionNotes.value = '';
      const job = selected.value;
      outsourceSupplier.value = job?.outsourceSupplierId || '';
      outsourceDescription.value = job?.outsourceDescription || '';
      outsourceCost.value = formatMoneyInput(
        job?.actualOutsourcedCostRial || 0,
        props.currencyUnit,
      );
      outsourceQuotedCost.value = formatMoneyInput(
        job?.outsourceQuotedCostRial || 0,
        props.currencyUnit,
      );
    }
    try {
      const result = await productionApi.reservations('', id, '');
      if (selectedId.value === id) reservations.value = result;
    } catch (e) {
      if (selectedId.value === id) {
        toast.error(message(e, 'Reservations could not be loaded.'));
      }
    } finally {
      if (selectedId.value === id) reservationsLoading.value = false;
    }
  }
  function beginCreate() {
    reservationsLoading.value = false;
    createMode.value = true;
    editing.value = false;
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
      plannedAt: null,
    };
  }
  function beginEdit() {
    const job = selected.value;
    if (!job || job.status === 'Completed' || job.status === 'Cancelled') return;
    form.value = {
      orderId: job.orderId,
      orderItemId: job.orderItemId,
      quantity: job.quantity,
      quantityUnit: job.quantityUnit,
      assignedMachineId: job.assignedMachineId,
      priority: job.priority,
      notes: job.notes,
      plannedAt: job.plannedAt || null,
    };
    editing.value = true;
  }
  function cancelEdit() {
    editing.value = false;
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
      if (!form.value.orderId || !form.value.orderItemId || !form.value.quantity) {
        toast.warning('Select an order, item, and quantity before creating the job.');
        return;
      }
      saving.value = true;
      try {
        const j = await productionApi.create({
          ...form.value,
          quantity: parseQuantityInput(form.value.quantity),
          plannedAt: null,
        });
        jobs.value.unshift(j);
        createMode.value = false;
        await select(j.id);
        emit('notify', 'Production job created from the saved order item.');
      } catch (e) {
        toast.error(message(e, 'Production job could not be created.'));
      } finally {
        saving.value = false;
      }
    });
  }
  async function update() {
    return runAction(async () => {
      if (!selected.value || !editing.value) return;
      saving.value = true;
      try {
        const j = await productionApi.update(selected.value.id, {
          ...form.value,
          quantity: parseQuantityInput(form.value.quantity),
          plannedAt: form.value.plannedAt
            ? form.value.plannedAt.includes('T')
              ? form.value.plannedAt
              : `${form.value.plannedAt}T00:00:00Z`
            : null,
        });
        replace(j);
        editing.value = false;
        emit('notify', 'Production job updated.');
      } catch (e) {
        toast.error(message(e, 'Production job could not be updated.'));
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
        toast.error(message(e, 'Production status could not be changed.'));
      }
    });
  }
  async function remove() {
    return runAction(async () => {
      if (!selected.value) return;
      const job = selected.value;
      if (
        !(await confirmAction({
          title: 'Delete production job?',
          message: `Delete ${job.jobNumber}? Reservations and consumption records will be removed, and inventory will receive compensating movements. This cannot be undone.`,
          confirmLabel: 'Delete permanently',
          danger: true,
        }))
      )
        return;
      try {
        await productionApi.delete(job.id);
        jobs.value = jobs.value.filter((j) => j.id !== job.id);
        selectedId.value = null;
        reservations.value = [];
        emit('notify', 'Production job deleted.');
      } catch (e) {
        toast.error(message(e, 'Production job could not be deleted.'));
      }
    });
  }
  async function reserve() {
    return runAction(async () => {
      if (!selected.value || !reservationMaterial.value || !reservationQuantity.value) {
        toast.warning('Select a material and enter a reservation quantity.');
        return;
      }
      try {
        await productionApi.reserve({
          materialId: reservationMaterial.value,
          orderId: selected.value.orderId,
          orderItemId: selected.value.orderItemId,
          productionJobId: selected.value.id,
          quantity: parseQuantityInput(reservationQuantity.value),
        });
        reservationQuantity.value = '';
        reservations.value = await productionApi.reservations('', selected.value.id, '');
        await load();
        emit('notify', 'Inventory reserved without creating a movement.');
      } catch (e) {
        toast.error(message(e, 'Reservation exceeds available stock or is invalid.'));
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
        toast.error(message(e, 'Reservation could not be released.'));
      }
    });
  }
  async function consume() {
    return runAction(async () => {
      if (!selected.value || !consumptionMaterial.value || (!consumedQuantity.value && !wasteQuantity.value)) {
        toast.warning('Select a material and enter consumed quantity or waste.');
        return;
      }
      try {
        await productionApi.consume(selected.value.id, {
          materialId: consumptionMaterial.value,
          consumedQuantity: parseQuantityInput(consumedQuantity.value || '0'),
          wasteQuantity: parseQuantityInput(wasteQuantity.value || '0'),
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
        toast.error(message(e, 'Consumption could not be posted.'));
      }
    });
  }
  async function saveOutsource() {
    return runAction(async () => {
      if (!selected.value) return;
      const cost = parseMoneyInput(outsourceCost.value, props.currencyUnit);
      const quoted = parseMoneyInput(outsourceQuotedCost.value, props.currencyUnit);
      if (cost === null || quoted === null) {
        toast.warning('Enter whole outsourced costs.');
        return;
      }
      try {
        const j = await productionApi.outsource(selected.value.id, {
          supplierId: outsourceSupplier.value,
          description: outsourceDescription.value,
          sentAt: selected.value.outsourceSentAt,
          expectedReturnAt: selected.value.outsourceExpectedReturnAt,
          receivedAt: selected.value.outsourceReceivedAt,
          notes: selected.value.outsourceNotes,
          quotedCostRial: quoted,
          actualCostRial: cost,
        });
        replace(j);
        emit('notify', 'Outsourced production metadata saved.');
      } catch (e) {
        toast.error(message(e, 'Outsourcing metadata could not be saved.'));
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
  return {
    busy,
    runAction,
    jobs,
    selectedId,
    statusFilter,
    searchQuery,
    loading,
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
    consumptionNotes,
    outsourceSupplier,
    outsourceDescription,
    outsourceCost,
    outsourceQuotedCost,
    selected,
    selectedOrder,
    selectedItem,
    visibleJobs,
    load,
    select,
    beginCreate,
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
    replace,
    jobOrder,
    jobContext,
    statusTone,
    message,
    date,
  };
}
