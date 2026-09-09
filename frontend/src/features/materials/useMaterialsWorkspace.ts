import { computed, onMounted, ref, watch } from 'vue';
import { materialsApi, type MaterialPayload, type MaterialRecord } from '../../api/materials';
import { purchasesApi } from '../../api/purchases';
import { useWorkspaceActions } from '../../composables/useWorkspaceActions';
import { confirmAction, useToast } from '../../ui/feedback';
import {
  formatMoneyInput,
  parseMoneyInput,
  type CurrencyUnit,
} from '../../utils/currency';
import { formatDateTime } from '../../utils/date';

export type MaterialFilter = 'Active' | 'Archived' | 'All';
export type EditorMode = 'create' | 'edit' | null;
export type MaterialForm = Omit<MaterialPayload, 'averageUnitCostRial'> & {
  averageUnitCostRial: number;
};

type MaterialsProps = { currencyUnit: CurrencyUnit };
type MaterialsEmit = (event: 'notify', message: string) => void;

export function useMaterialsWorkspace(props: MaterialsProps, emit: MaterialsEmit) {
  const { busy, runAction } = useWorkspaceActions();
  const toast = useToast();

  const materials = ref<MaterialRecord[]>([]);
  const selectedId = ref<string | null>(null);
  const searchQuery = ref('');
  const materialFilter = ref<MaterialFilter>('Active');
  const editorMode = ref<EditorMode>(null);
  const form = ref<MaterialForm>(emptyForm());
  const costDraft = ref('0');
  const isLoading = ref(false);
  const isSaving = ref(false);
  const movements = ref<any[]>([]);
  const isMovementsLoading = ref(false);
  const adjustmentQuantity = ref('');
  const adjustmentCost = ref('0');
  const adjustmentNote = ref('');

  const unitOptions = [
    'piece',
    'sheet',
    'pack',
    'kilogram',
    'gram',
    'roll',
    'meter',
    'liter',
    'milliliter',
    'square meter',
  ];

  const selectedMaterial = computed(
    () => materials.value.find((material) => material.id === selectedId.value) ?? null,
  );
  const filteredMaterials = computed(() => {
    const query = searchQuery.value.trim().toLowerCase();
    return materials.value.filter((material) => {
      const matchesFilter =
        materialFilter.value === 'All' ||
        (materialFilter.value === 'Active' ? material.active : !material.active);
      const matchesSearch =
        !query ||
        [
          material.name,
          material.sku,
          material.category,
          material.purchaseUnit,
          material.consumptionUnit,
        ].some((value) => value.toLowerCase().includes(query));
      return matchesFilter && matchesSearch;
    });
  });

  onMounted(loadMaterials);

  watch(
    () => props.currencyUnit,
    () => {
      costDraft.value = formatMoneyInput(form.value.averageUnitCostRial, props.currencyUnit);
    },
  );

  watch(
    selectedId,
    (id) => {
      void loadMovements(id ?? '');
    },
    { immediate: true },
  );

  function emptyForm(): MaterialForm {
    return {
      name: '',
      sku: '',
      category: '',
      purchaseUnit: 'pack',
      consumptionUnit: 'sheet',
      conversionFactor: '500',
      physicalStock: '0',
      reorderLevel: '0',
      averageUnitCostRial: 0,
      preferredSupplier: '',
      notes: '',
    };
  }

  async function loadMaterials() {
    isLoading.value = true;
    try {
      materials.value = await materialsApi.list(true);
    } catch (error) {
      toast.error(errorMessageFrom(error, 'Materials could not be loaded.'), 'Materials');
    } finally {
      isLoading.value = false;
    }
  }

  async function loadMovements(id = selectedId.value ?? '') {
    if (!id) {
      movements.value = [];
      return;
    }
    isMovementsLoading.value = true;
    try {
      movements.value = await purchasesApi.movements(id);
    } catch (error) {
      toast.error(errorMessageFrom(error, 'Movement history could not be loaded.'), 'Materials');
    } finally {
      isMovementsLoading.value = false;
    }
  }

  function selectMaterial(id: string) {
    selectedId.value = id;
    editorMode.value = null;
  }

  function startCreate() {
    editorMode.value = 'create';
    selectedId.value = null;
    form.value = emptyForm();
    costDraft.value = formatMoneyInput(0, props.currencyUnit);
    adjustmentQuantity.value = '';
    adjustmentNote.value = '';
  }

  function startEdit() {
    const material = selectedMaterial.value;
    if (!material) return;
    form.value = {
      name: material.name,
      sku: material.sku,
      category: material.category,
      purchaseUnit: material.purchaseUnit,
      consumptionUnit: material.consumptionUnit,
      conversionFactor: material.conversionFactor,
      physicalStock: material.physicalStock,
      reorderLevel: material.reorderLevel,
      averageUnitCostRial: material.averageUnitCostRial,
      preferredSupplier: material.preferredSupplier,
      notes: material.notes,
    };
    costDraft.value = formatMoneyInput(material.averageUnitCostRial, props.currencyUnit);
    editorMode.value = 'edit';
  }

  function cancelEditor() {
    editorMode.value = null;
  }

  function backToMaterials() {
    selectedId.value = null;
    editorMode.value = null;
    movements.value = [];
  }

  function updateCost(value: string) {
    const parsed = parseMoneyInput(value, props.currencyUnit);
    if (parsed !== null) {
      form.value.averageUnitCostRial = parsed;
      costDraft.value = formatMoneyInput(parsed, props.currencyUnit);
    } else {
      costDraft.value = value;
    }
  }

  function updateAdjustmentCost(value: string) {
    const parsed = parseMoneyInput(value, props.currencyUnit);
    adjustmentCost.value = parsed === null ? value : formatMoneyInput(parsed, props.currencyUnit);
  }

  async function saveMaterial() {
    return runAction(async () => {
      const averageUnitCostRial = parseMoneyInput(costDraft.value, props.currencyUnit);
      if (averageUnitCostRial === null) {
        toast.error(`Enter a whole ${props.currencyUnit.toLowerCase()} amount.`, 'Materials');
        return;
      }
      form.value.averageUnitCostRial = averageUnitCostRial;
      isSaving.value = true;
      const wasEditing = editorMode.value === 'edit';
      try {
        const saved =
          wasEditing && selectedId.value
            ? await materialsApi.update(selectedId.value, payload())
            : await materialsApi.create(payload());
        const existingIndex = materials.value.findIndex((material) => material.id === saved.id);
        if (existingIndex >= 0) materials.value.splice(existingIndex, 1, saved);
        else materials.value.push(saved);
        selectedId.value = saved.id;
        editorMode.value = null;
        emit('notify', wasEditing ? 'Material updated.' : 'Material created.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Material could not be saved.'), 'Materials');
      } finally {
        isSaving.value = false;
      }
    });
  }

  function payload(): MaterialPayload {
    return { ...form.value };
  }

  async function adjustStock() {
    return runAction(async () => {
      const material = selectedMaterial.value;
      if (!material || !adjustmentQuantity.value.trim()) {
        toast.error('Enter a quantity adjustment.', 'Materials');
        return;
      }
      const cost = parseMoneyInput(adjustmentCost.value, props.currencyUnit);
      if (cost === null) {
        toast.error('Enter a whole adjustment cost.', 'Materials');
        return;
      }
      try {
        await purchasesApi.adjust(material.id, adjustmentQuantity.value, cost, adjustmentNote.value);
        adjustmentQuantity.value = '';
        adjustmentNote.value = '';
        await loadMaterials();
        await loadMovements(material.id);
        emit('notify', 'Stock adjustment recorded as an immutable movement.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Stock adjustment could not be recorded.'), 'Materials');
      }
    });
  }

  async function setActive(active: boolean) {
    return runAction(async () => {
      const material = selectedMaterial.value;
      if (!material) return;
      try {
        const updated = active
          ? await materialsApi.reactivate(material.id)
          : await materialsApi.archive(material.id);
        const index = materials.value.findIndex((item) => item.id === updated.id);
        if (index >= 0) materials.value.splice(index, 1, updated);
        emit('notify', active ? 'Material reactivated.' : 'Material archived.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Material status could not be changed.'), 'Materials');
      }
    });
  }

  async function remove() {
    return runAction(async () => {
      const material = selectedMaterial.value;
      if (
        !material ||
        !(await confirmAction({
          title: 'Delete material',
          message:
            'Delete this material permanently? Materials with inventory or production history must be archived instead.',
          confirmLabel: 'Delete material',
          danger: true,
        }))
      ) {
        return;
      }
      try {
        await materialsApi.remove(material.id);
        materials.value = materials.value.filter((item) => item.id !== material.id);
        backToMaterials();
        emit('notify', 'Material deleted.');
      } catch (error) {
        toast.error(errorMessageFrom(error, 'Material could not be deleted.'), 'Materials');
      }
    });
  }

  function errorMessageFrom(error: unknown, fallback: string): string {
    return error instanceof Error && error.message
      ? error.message
      : typeof error === 'string'
        ? error
        : fallback;
  }

  function unitLabel(unit: string) {
    return unit.replace(/\b\w/g, (letter) => letter.toUpperCase());
  }

  function dateLabel(value: string) {
    try {
      return formatDateTime(value);
    } catch {
      return 'Unknown date';
    }
  }

  return {
    busy,
    materials,
    selectedId,
    selectedMaterial,
    searchQuery,
    materialFilter,
    editorMode,
    form,
    costDraft,
    isLoading,
    isSaving,
    movements,
    isMovementsLoading,
    adjustmentQuantity,
    adjustmentCost,
    adjustmentNote,
    unitOptions,
    filteredMaterials,
    loadMaterials,
    selectMaterial,
    startCreate,
    startEdit,
    cancelEditor,
    backToMaterials,
    updateCost,
    updateAdjustmentCost,
    saveMaterial,
    adjustStock,
    setActive,
    remove,
    unitLabel,
    dateLabel,
  };
}
