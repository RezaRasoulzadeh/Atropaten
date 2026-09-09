import { computed, onMounted, ref } from 'vue';
import { purchasesApi, type PurchaseItemRecord, type PurchasePayload, type PurchaseRecord } from '../../api/purchases';
import type { SupplierRecord } from '../../api/suppliers';
import type { MaterialRecord } from '../../api/materials';
import { accountingApi, type FinancialAccountRecord } from '../../api/accounting';
import { useWorkspaceActions } from '../../composables/useWorkspaceActions';
import { confirmAction, useToast } from '../../ui/feedback';
import { currentCanonicalDate } from '../../utils/date';
import { formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency';
import { formatQuantityUnits } from '../../utils/quantity';

export type PurchaseFilter = 'All' | 'Draft' | 'Posted' | 'Cancelled' | 'Archived';

type PurchasesProps = {
  currencyUnit: CurrencyUnit;
  suppliers: SupplierRecord[];
  materials: MaterialRecord[];
};

type PurchasesEmit = (event: 'notify', message: string) => void;

export function usePurchasesWorkspace(props: PurchasesProps, emit: PurchasesEmit) {
  const { busy, runAction } = useWorkspaceActions();
  const toast = useToast();
  const rows = ref<PurchaseRecord[]>([]);
  const financial = ref<FinancialAccountRecord[]>([]);
  const selectedId = ref<string | null>(null);
  const searchQuery = ref('');
  const purchaseFilter = ref<PurchaseFilter>('All');
  const isLoading = ref(false);
  const editing = ref(false);
  const createMode = ref<'create' | null>(null);
  const editingItemId = ref<string | null>(null);
  const draftItems = ref<PurchaseItemRecord[]>([]);
  const form = ref<PurchasePayload>(emptyForm());
  const item = ref({
    materialId: '',
    purchaseQuantity: '1',
    unitAcquisitionCostRial: '0',
    notes: '',
  });

  const current = computed(() => {
    if (createMode.value) return draftPurchase();
    return rows.value.find((value) => value.id === selectedId.value) ?? null;
  });
  const filteredRows = computed(() => {
    const query = searchQuery.value.trim().toLowerCase();
    return rows.value.filter((value) => {
      const matchesStatus = purchaseFilter.value === 'All' || value.status === purchaseFilter.value;
      const matchesSearch =
        !query ||
        [value.purchaseNumber, value.supplierName, value.supplierInvoiceNumber]
          .join(' ')
          .toLowerCase()
          .includes(query);
      return matchesStatus && matchesSearch;
    });
  });

  onMounted(load);

  function emptyForm(): PurchasePayload {
    return {
      supplierId: props.suppliers[0]?.id ?? '',
      financialAccountId: 'FIN-CASH',
      purchaseDate: currentCanonicalDate(),
      supplierInvoiceNumber: '',
      notes: '',
      discountRial: 0,
      shippingRial: 0,
      taxRial: 0,
      additionalCostsRial: 0,
    };
  }

  function dateOnly(value: string) {
    return value.slice(0, 10);
  }

  function copyForm(value: PurchaseRecord): PurchasePayload {
    return {
      supplierId: value.supplierId,
      financialAccountId: value.financialAccountId || '',
      purchaseDate: dateOnly(value.purchaseDate),
      supplierInvoiceNumber: value.supplierInvoiceNumber,
      notes: value.notes,
      discountRial: value.discountRial,
      shippingRial: value.shippingRial,
      taxRial: value.taxRial,
      additionalCostsRial: value.additionalCostsRial,
    };
  }

  function resetItem() {
    editingItemId.value = null;
    item.value = {
      materialId: '',
      purchaseQuantity: '1',
      unitAcquisitionCostRial: formatMoneyInput(0, props.currencyUnit),
      notes: '',
    };
  }

  function fixedQuantity(value: string): bigint | null {
    const normalized = value.trim().replaceAll(',', '');
    if (!/^\d+(?:\.\d+)?$/.test(normalized)) return null;
    const [whole, fraction = ''] = normalized.split('.');
    if (fraction.length > 6) return null;
    try {
      return BigInt(whole) * 1_000_000n + BigInt(fraction.padEnd(6, '0') || '0');
    } catch {
      return null;
    }
  }

  function localItem(): { item: PurchaseItemRecord } | { error: string } {
    if (!item.value.materialId) return { error: 'Select a material.' };
    const material = props.materials.find((value) => value.id === item.value.materialId);
    const quantity = fixedQuantity(item.value.purchaseQuantity);
    const cost = parseMoneyInput(item.value.unitAcquisitionCostRial, props.currencyUnit);
    if (!material) return { error: 'The selected material is no longer available.' };
    if (quantity === null || quantity <= 0n) return { error: 'Enter a quantity greater than zero.' };
    if (cost === null || cost < 0) return { error: 'Enter a valid non-negative unit cost.' };
    const conversionText = String(material.conversionFactor ?? '').trim();
    const conversion = fixedQuantity(conversionText);
    if (conversion === null || conversion <= 0n) return { error: 'The selected material has an invalid unit conversion.' };
    const converted = quantity * conversion;
    if (converted % 1_000_000n !== 0n) {
      return { error: 'The selected material conversion cannot represent this quantity.' };
    }
    const lineTotal = Number((quantity * BigInt(cost) + 500_000n) / 1_000_000n);
    return { item: {
      id: editingItemId.value || `draft-item-${Date.now()}-${Math.random().toString(16).slice(2)}`,
      position: draftItems.value.length,
      materialId: material.id,
      materialName: material.name,
      purchaseUnit: material.purchaseUnit,
      consumptionUnit: material.consumptionUnit,
      purchaseQuantity: formatQuantityUnits(String(quantity)),
      conversionFactor: material.conversionFactor,
      consumptionQuantity: formatQuantityUnits(String(converted / 1_000_000n)),
      unitAcquisitionCostRial: cost,
      allocatedAdditionalCostRial: 0,
      landedUnitCostRial: 0,
      lineTotalRial: lineTotal,
      notes: item.value.notes.trim(),
    } };
  }

  function draftPurchase(): PurchaseRecord {
    const supplier = props.suppliers.find((value) => value.id === form.value.supplierId);
    const subtotalRial = draftItems.value.reduce((total, value) => total + value.lineTotalRial, 0);
    const totalRial = subtotalRial - Number(form.value.discountRial || 0) + Number(form.value.shippingRial || 0) + Number(form.value.taxRial || 0) + Number(form.value.additionalCostsRial || 0);
    const now = new Date().toISOString();
    return {
      id: 'new-purchase',
      purchaseNumber: 'New purchase',
      supplierId: form.value.supplierId,
      financialAccountId: form.value.financialAccountId,
      supplierName: supplier?.name || 'No supplier selected',
      supplierInvoiceNumber: form.value.supplierInvoiceNumber,
      purchaseDate: `${form.value.purchaseDate}T00:00:00Z`,
      status: 'Draft',
      notes: form.value.notes,
      subtotalRial,
      discountRial: Number(form.value.discountRial || 0),
      shippingRial: Number(form.value.shippingRial || 0),
      taxRial: Number(form.value.taxRial || 0),
      additionalCostsRial: Number(form.value.additionalCostsRial || 0),
      totalRial,
      paidRial: 0,
      remainingRial: totalRial,
      createdAt: now,
      updatedAt: now,
      items: draftItems.value,
    };
  }

  async function load() {
    isLoading.value = true;
    try {
      [rows.value, financial.value] = await Promise.all([
        purchasesApi.list(),
        accountingApi.financialAccounts(),
      ]);
      if (!financial.value.some((account) => account.id === form.value.financialAccountId && account.active)) {
        form.value.financialAccountId = financial.value.find((account) => account.active)?.id ?? '';
      }
    } catch (error) {
      toast.error(errorMessageFrom(error, 'Purchases could not be loaded.'), 'Purchases');
    } finally {
      isLoading.value = false;
    }
  }

  function select(id: string) {
    const value = rows.value.find((purchase) => purchase.id === id);
    if (!value) return;
    selectedId.value = value.id;
    editing.value = false;
    createMode.value = null;
    draftItems.value = [];
    form.value = copyForm(value);
    resetItem();
  }

  function startEdit() {
    const value = current.value;
    if (!value || createMode.value) return;
    form.value = copyForm(value);
    draftItems.value = [];
    resetItem();
    editing.value = true;
  }

  function startItemEdit(line: PurchaseItemRecord) {
    editing.value = true;
    editingItemId.value = line.id;
    item.value = {
      materialId: line.materialId,
      purchaseQuantity: line.purchaseQuantity,
      unitAcquisitionCostRial: formatMoneyInput(line.unitAcquisitionCostRial, props.currencyUnit),
      notes: line.notes,
    };
  }

  function cancelItemEdit() {
    resetItem();
  }

  function startCreate() {
    selectedId.value = null;
    createMode.value = 'create';
    draftItems.value = [];
    form.value = emptyForm();
    resetItem();
    editing.value = true;
  }

  function cancelEditor() {
    if (createMode.value) {
      backToPurchases();
      return;
    }
    editing.value = false;
    if (current.value) form.value = copyForm(current.value);
  }

  function backToPurchases() {
    selectedId.value = null;
    editing.value = false;
    createMode.value = null;
    draftItems.value = [];
    form.value = emptyForm();
    resetItem();
  }

  async function save() {
    return runAction(async () => {
      if (!current.value) return;
      if (!financial.value.some((account) => account.id === form.value.financialAccountId && account.active)) {
        toast.error('Select an active cash or bank account for this purchase.', 'Purchases');
        return;
      }
      if (createMode.value) {
        if (!props.suppliers.length) {
          toast.error('Create a supplier before recording a purchase.', 'Purchases');
          return;
        }
        let created: PurchaseRecord | null = null;
        try {
          created = await purchasesApi.create(form.value);
          for (const line of draftItems.value) {
            created = await purchasesApi.addItem(created.id, {
              materialId: line.materialId,
              purchaseQuantity: line.purchaseQuantity,
              unitAcquisitionCostRial: String(line.unitAcquisitionCostRial),
              notes: line.notes,
            });
          }
          rows.value = [created, ...rows.value];
          selectedId.value = created.id;
          form.value = copyForm(created);
          createMode.value = null;
          draftItems.value = [];
          editing.value = false;
          emit('notify', 'Purchase created.');
        } catch (error) {
          if (created) {
            try {
              await purchasesApi.remove(created.id);
            } catch {
              // Keep the original failure as the user-facing message.
            }
          }
          toast.error(errorMessageFrom(error, 'Purchase could not be saved.'), 'Purchases');
        }
        return;
      }
      try {
        replace(await purchasesApi.update(current.value.id, form.value));
        editing.value = false;
        emit('notify', 'Purchase saved.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Purchase could not be saved.'), 'Purchases');
      }
    });
  }

  async function addItem() {
    return runAction(async () => {
      if (!current.value) return;
      if (createMode.value) {
        const isUpdating = Boolean(editingItemId.value);
        const result = localItem();
        if ('error' in result) {
          toast.error(result.error, 'Purchases');
          return;
        }
        draftItems.value = (editingItemId.value
          ? draftItems.value.map((line) => (line.id === editingItemId.value ? result.item : line))
          : [...draftItems.value, result.item]
        ).map((line, index) => ({ ...line, position: index }));
        resetItem();
        emit('notify', isUpdating ? 'Purchase item updated in the unsaved purchase.' : 'Purchase item added to the unsaved purchase.');
        return;
      }
      try {
        const isUpdating = Boolean(editingItemId.value);
        const cost = parseMoneyInput(item.value.unitAcquisitionCostRial, props.currencyUnit);
        if (cost === null) {
          toast.error('Enter a valid unit cost.', 'Purchases');
          return;
        }
        const payload = {
            ...item.value,
            unitAcquisitionCostRial: String(cost),
        };
        replace(editingItemId.value
          ? await purchasesApi.updateItem(current.value.id, editingItemId.value, payload)
          : await purchasesApi.addItem(current.value.id, payload));
        resetItem();
        emit('notify', isUpdating ? 'Purchase item updated.' : 'Purchase item added.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Purchase item could not be added.'), 'Purchases');
      }
    });
  }

  async function removeItem(id: string) {
    return runAction(async () => {
      if (!current.value) return;
      if (createMode.value) {
        draftItems.value = draftItems.value.filter((line) => line.id !== id).map((line, index) => ({ ...line, position: index }));
        emit('notify', 'Purchase item removed from the unsaved purchase.');
        return;
      }
      try {
        replace(await purchasesApi.removeItem(current.value.id, id));
        emit('notify', 'Purchase item removed.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Purchase item could not be removed.'), 'Purchases');
      }
    });
  }

  async function reorder(index: number, direction: number) {
    return runAction(async () => {
      if (!current.value) return;
      const items = [...current.value.items].sort((a, b) => a.position - b.position);
      const target = index + direction;
      if (target < 0 || target >= items.length) return;
      [items[index], items[target]] = [items[target], items[index]];
      if (createMode.value) {
        draftItems.value = items.map((line, position) => ({ ...line, position }));
        emit('notify', 'Purchase items reordered.');
        return;
      }
      try {
        replace(await purchasesApi.reorder(current.value.id, items.map((value) => value.id)));
        emit('notify', 'Purchase items reordered.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Purchase items could not be reordered.'), 'Purchases');
      }
    });
  }

  async function post() {
    return runAction(async () => {
      if (
        !current.value ||
        !(await confirmAction({
          title: 'Post purchase',
          message: 'Post this purchase and create inventory movements?',
          confirmLabel: 'Post purchase',
        }))
      )
        return;
      try {
        replace(await purchasesApi.post(current.value.id));
        editing.value = false;
        emit('notify', 'Purchase posted.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Purchase could not be posted.'), 'Purchases');
      }
    });
  }

  async function archivePurchase() {
    return runAction(async () => {
      if (
        !current.value ||
        !(await confirmAction({
          title: 'Archive purchase',
          message: 'Archive this purchase? It will remain in history and can no longer be used in active workflows.',
          confirmLabel: 'Archive purchase',
        }))
      )
        return;
      try {
        replace(await purchasesApi.archive(current.value.id));
        editing.value = false;
        emit('notify', 'Purchase archived.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Purchase could not be archived.'), 'Purchases');
      }
    });
  }

  async function removePurchase() {
    return runAction(async () => {
      if (
        !current.value ||
        !(await confirmAction({
          title: 'Delete purchase',
          message: 'Delete this purchase permanently? If inventory, production, or payment history depends on it, deletion will be refused and you can archive it instead.',
          confirmLabel: 'Delete purchase',
          danger: true,
        }))
      )
        return;
      const id = current.value.id;
      try {
        await purchasesApi.remove(id);
        rows.value = rows.value.filter((value) => value.id !== id);
        backToPurchases();
        emit('notify', 'Purchase deleted.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Purchase could not be deleted. Archive it instead.'), 'Purchases');
      }
    });
  }

  function replace(value: PurchaseRecord) {
    rows.value = rows.value.map((purchase) => (purchase.id === value.id ? value : purchase));
    selectedId.value = value.id;
    form.value = copyForm(value);
  }

  function errorMessageFrom(error: unknown, fallback: string) {
    return error instanceof Error && error.message
      ? error.message
      : typeof error === 'string'
        ? error
        : fallback;
  }

  return {
    busy,
    rows,
    financial,
    filteredRows,
    selectedId,
    current,
    searchQuery,
    purchaseFilter,
    isLoading,
    editing,
    editingItemId,
    createMode,
    form,
    item,
    select,
    startCreate,
    startEdit,
    startItemEdit,
    cancelItemEdit,
    cancelEditor,
    backToPurchases,
    save,
    addItem,
    removeItem,
    reorder,
    post,
    archivePurchase,
    removePurchase,
  };
}
