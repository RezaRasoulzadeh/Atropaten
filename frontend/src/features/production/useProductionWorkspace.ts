import { useWorkspaceActions } from '../../composables/useWorkspaceActions';
import { confirmAction, useToast } from '../../ui/feedback';
import { computed, onMounted, ref, watch } from 'vue';
import {
  productionApi,
  type ProductionJobRecord,
  type ReservationRecord,
  type ConsumptionRecord,
  type ProductionMaterialRecord,
} from '../../api/production';
import { materialsApi, type MaterialRecord } from '../../api/materials';
import { accountingApi, type FinancialAccountRecord } from '../../api/accounting';
import { ordersApi, type OrderRecord } from '../../api/orders';
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
  emit: {
    (event: 'notify', value: string): void
    (event: 'order-updated', value: OrderRecord): void
  },
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
  const materialPlans = ref<ProductionMaterialRecord[]>([]);
  const consumptions = ref<ConsumptionRecord[]>([]);
  const stockMaterials = ref<MaterialRecord[]>([]);
  const financialAccounts = ref<FinancialAccountRecord[]>([]);
  const materialsError = ref('');
  const editingConsumptionId = ref<string | null>(null);
  const outsourceQuantity = ref('');
  const outsourceAccount = ref('');
  const initializedJobId = ref('');
  let selectionRequest = 0;
  const consumptionLimit = computed(() => {
    const id = consumptionMaterial.value;
    const material = stockMaterials.value.find(m => m.id === id);
    const reserved = reservations.value.filter(r => r.materialId === id && r.status === 'active').reduce((n,r) => n+Number(r.quantity),0);
    const existing = consumptions.value.find(c => c.id === editingConsumptionId.value && !c.reversed);
    return Math.max(0,Number(material?.availableStock || 0)+reserved+Number(existing?.consumedQuantity || 0)+Number(existing?.wasteQuantity || 0)-Number(parseQuantityInput(wasteQuantity.value || '0') || 0));
  });
  const outsourceTotal = computed(() => Math.round(Number(parseQuantityInput(outsourceQuantity.value || '0')) * (parseMoneyInput(outsourceCost.value,props.currencyUnit) || 0)));
  const canConsume = computed(() => !!selected.value && ['In Progress','Paused'].includes(selected.value.status) && Number(selected.value.outsourceQuantity || 0) < Number(selected.value.quantity));
  const editableJob = computed(() => !!selected.value && !['Completed','Cancelled'].includes(selected.value.status));

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
      const id = jobs.value.some(j => j.id === selectedId.value) ? selectedId.value : jobs.value[0]?.id;
      if (id) await select(id);
      else { selectedId.value = null; reservations.value=[]; materialPlans.value=[]; consumptions.value=[]; }
    } catch (e) { toast.error(message(e,'Production queue could not be loaded.')); }
    finally { loading.value = false; }
  }
  async function select(id: string) {
    const request = ++selectionRequest;
    selectedId.value = id;
    createMode.value = editing.value = false;
    reservationsLoading.value = true;
    materialsError.value = '';
    try {
      // Reconcile first, then load the reservations and balances it produced.
      let plans: ProductionMaterialRecord[] = [];
      try { plans = await productionApi.materials(id); }
      catch(e) { if(request === selectionRequest) materialsError.value = message(e,'Order materials could not be synchronized.'); }
      const [job, reserved, used, stock, accounts] = await Promise.all([
        productionApi.get(id), productionApi.reservations('',id,''), productionApi.consumptions(id),
        materialsApi.list(), accountingApi.financialAccounts(),
      ]);
      if (request !== selectionRequest) return;
      replace(job);
      materialPlans.value = plans;
      reservations.value = reserved;
      consumptions.value = used;
      stockMaterials.value = stock;
      financialAccounts.value = accounts.filter(a => a.active);
      if (initializedJobId.value !== id) {
        initializedJobId.value = id;
        reservationMaterial.value = reservationQuantity.value = '';
        editingConsumptionId.value = null;
        consumptionKey.value = consumptionNotes.value = '';
        consumptionMaterial.value = reserved.find(r => r.status === 'active')?.materialId || plans[0]?.materialId || '';
        suggestConsumption();
        resetOutsourceDraft();
      }
    } catch(e) {
      if(request === selectionRequest) { materialsError.value=message(e,'Production details could not be loaded.'); toast.error(materialsError.value); }
    } finally { if(request === selectionRequest) reservationsLoading.value=false; }
  }
  function resetOutsourceDraft() {
    const job=selected.value;
    outsourceSupplier.value=job?.outsourceSupplierId || '';
    outsourceDescription.value=job?.outsourceDescription || '';
    outsourceQuantity.value=Number(job?.outsourceQuantity || 0)>0 ? job!.outsourceQuantity : job?.quantity || '';
    const unitCost=Number(job?.outsourceQuantity || 0)>0 ? job!.outsourceUnitCostRial : Math.round((job?.estimatedCostRial || 0)/Number(job?.quantity || 1));
    outsourceCost.value=formatMoneyInput(unitCost,props.currencyUnit);
    outsourceQuotedCost.value=formatMoneyInput(job?.estimatedCostRial || 0,props.currencyUnit);
    outsourceAccount.value=job?.outsourceFinancialAccountId || financialAccounts.value.find(a => a.type==='cash')?.id || financialAccounts.value[0]?.id || '';
  }
  function suggestConsumption() {
    const amount=reservations.value.filter(r=>r.materialId===consumptionMaterial.value && r.status==='active').reduce((n,r)=>n+Number(r.quantity),0);
    consumedQuantity.value=String(amount);
    wasteQuantity.value='0';
    consumptionKey.value='';
    editingConsumptionId.value=null;
  }
  function updateConsumptionSlider(event: Event) {
    const percent=Number((event.target as HTMLInputElement).value);
    consumedQuantity.value=String(Math.round(consumptionLimit.value*percent/100*1000000)/1000000);
  }
  function editConsumption(record: ConsumptionRecord) {
    if(record.reversed)return;
    consumptionMaterial.value=record.materialId;
    consumedQuantity.value=record.consumedQuantity;
    wasteQuantity.value=record.wasteQuantity;
    consumptionNotes.value=record.notes;
    editingConsumptionId.value=record.id;
    consumptionKey.value=crypto.randomUUID();
  }
  async function saveMaterialTarget(plan: ProductionMaterialRecord, value: string) {
    return runAction(async()=> {
      if(!selected.value)return;
      try {
        await productionApi.setMaterialTarget(selected.value.id,plan.materialId,parseQuantityInput(value));
        await select(selected.value.id);
        emit('notify','Material plan updated.');
      } catch(e){toast.error(message(e,'Material plan could not be updated.'));}
    });
  }
  async function updateReservation(record: ReservationRecord, value: string) {
    return runAction(async()=>{
      try { await productionApi.updateReservation(record.id,parseQuantityInput(value)); if(selected.value)await select(selected.value.id); }
      catch(e){toast.error(message(e,'Reservation could not be updated.'));}
    });
  }
  async function correctConsumption(record: ConsumptionRecord) {
    return runAction(async()=>{
      try {
        await productionApi.updateConsumption(record.id,{materialId:record.materialId,consumedQuantity:'0',wasteQuantity:'0',idempotencyKey:crypto.randomUUID(),notes:'Material returned by operator'});
        if(selected.value){await select(selected.value.id);await refreshOrder(selected.value.orderId);}
      }catch(e){toast.error(message(e,'Consumption could not be returned.'));}
    });
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
  function chooseItem() {
    const item = selectedOrder.value?.items.find((x) => x.id === form.value.orderItemId);
    form.value.quantity = item?.quantity ?? '';
    form.value.quantityUnit = item?.quantityUnit ?? '';
  }
  function plannedAtInput() {
    if (!form.value.plannedAt) return null;
    return form.value.plannedAt.includes('T')
      ? form.value.plannedAt
      : `${form.value.plannedAt}T00:00:00Z`;
  }
  function validCreateDraft() {
    const order = props.orders.find((item) => item.id === form.value.orderId);
    const item = order?.items.find((entry) => entry.id === form.value.orderItemId);
    const quantity = Number(parseQuantityInput(form.value.quantity));
    if (!order || order.commercialStatus !== 'Confirmed') return 'Choose a confirmed order.';
    if (!item) return 'Choose an order item.';
    if (!Number.isFinite(quantity) || quantity <= 0) return 'Enter a quantity greater than zero.';
    if (!form.value.quantityUnit.trim()) return 'Enter a quantity unit.';
    return '';
  }
  async function create() {
    return runAction(async () => {
      const draftError = validCreateDraft();
      if (draftError) {
        toast.warning(draftError);
        return;
      }
      saving.value = true;
      try {
        const j = await productionApi.create({
          ...form.value,
          quantity: parseQuantityInput(form.value.quantity),
          plannedAt: plannedAtInput(),
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
          plannedAt: plannedAtInput(),
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
        await refreshOrder(j.orderId);
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
        await refreshOrder(job.orderId);
        await load();
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
        if (selected.value) await select(selected.value.id);
      } catch (e) {
        toast.error(message(e, 'Reservation could not be released.'));
      }
    });
  }
  async function consumeReservation(r: ReservationRecord) {
    if (r.status !== 'active') return;
    consumptionMaterial.value = r.materialId;
    consumedQuantity.value = r.quantity;
    wasteQuantity.value = '0';
    editingConsumptionId.value = null;
    consumptionKey.value = '';
  }
  async function consume() {
    return runAction(async () => {
      if (!selected.value || !consumptionMaterial.value || (!consumedQuantity.value && !wasteQuantity.value)) {
        toast.warning('Select a material and enter consumed quantity or waste.');
        return;
      }
      try {
        const payload = {
          materialId: consumptionMaterial.value,
          consumedQuantity: parseQuantityInput(consumedQuantity.value || '0'),
          wasteQuantity: parseQuantityInput(wasteQuantity.value || '0'),
          idempotencyKey: consumptionKey.value ||= crypto.randomUUID(),
          notes: consumptionNotes.value,
        };
        if (editingConsumptionId.value) await productionApi.updateConsumption(editingConsumptionId.value,payload);
        else await productionApi.consume(selected.value.id,payload);
        editingConsumptionId.value=null;
        consumedQuantity.value = '';
        wasteQuantity.value = '';
        consumptionKey.value = '';
        consumptionNotes.value = '';
        await load();
        if (selected.value) await select(selected.value.id);
        await refreshOrder(selected.value?.orderId || '');
        emit('notify', 'Material usage saved; stock and reservations updated.');
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
          quantity: parseQuantityInput(outsourceQuantity.value || '0'),
          unitCostRial: cost,
          financialAccountId: outsourceAccount.value,
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
        await refreshOrder(j.orderId);
        await select(j.id);
        resetOutsourceDraft();
        emit('notify', 'Outsourcing saved; stock, expense, and order cost updated.');
      } catch (e) {
        toast.error(message(e, 'Outsourcing metadata could not be saved.'));
      }
    });
  }
  function replace(j: ProductionJobRecord) {
    const i = jobs.value.findIndex((x) => x.id === j.id);
    if (i >= 0) jobs.value.splice(i, 1, j);
  }
  async function refreshOrder(orderId: string) {
    if (!orderId) return;
    try {
      emit('order-updated', await ordersApi.get(orderId));
    } catch (e) {
      toast.warning(message(e, 'Order cost summary could not be refreshed.'));
    }
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
    materialPlans, consumptions, stockMaterials, financialAccounts, materialsError, editingConsumptionId,
    outsourceQuantity, outsourceAccount, outsourceTotal, consumptionLimit, canConsume, editableJob,
    suggestConsumption, updateConsumptionSlider, editConsumption, correctConsumption, saveMaterialTarget, updateReservation, resetOutsourceDraft,
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
    chooseItem,
    create,
    update,
    beginEdit,
    cancelEdit,
    changeStatus,
    remove,
    reserve,
    release,
    consumeReservation,
    consume,
    saveOutsource,
    replace,
    refreshOrder,
    jobOrder,
    jobContext,
    statusTone,
    message,
    date,
  };
}
