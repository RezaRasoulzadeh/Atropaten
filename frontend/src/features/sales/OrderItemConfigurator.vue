<script setup lang="ts">
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import FormField from '../../components/ui/FormField.vue';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { Calculator, Layers3, Plus, X } from 'lucide-vue-next';
import SelectField from '../../components/ui/SelectField.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import { pricingApi, type PricingRecord } from '../../api/pricing';
import { servicesApi } from '../../api/services';
import type { OrderItemPayload } from '../../api/orders';
import {
  formatMoney,
  formatMoneyInput,
  parseMoneyInput,
  roundMoneyUp,
  sellingPriceTotal,
  type CurrencyUnit,
} from '../../utils/currency';
import { useToast } from '../../ui/feedback';
import { serviceCategoryRequirements } from '../services/serviceCategory';
import { serviceDefaultPricingResult } from '../services/serviceDefaultPricing';
import RollSizeFields from '../services/RollSizeFields.vue';
import { ensureRollSizeInputs, usesMaterialRollWidth, selectedRollMaterial, rollWidthValue, rollStockWidthValue } from '../services/rollSizeInputs';

const props = withDefaults(
  defineProps<{
    services: any[];
    materials: any[];
    machines: any[];
    suppliers?: any[];
    currencyUnit: CurrencyUnit;
    initial?: any;
    presetServiceId?: string;
    documentLabel?: string;
    busy?: boolean;
  }>(),
  {
    documentLabel: 'order',
  },
);

const emit = defineEmits<{
  configured: [payload: OrderItemPayload, preview?: PricingRecord];
  cancel: [];
}>();

const serviceId = ref('');
const values = ref<Record<string, string>>({});
const quantity = ref('');
const unit = ref('unit');
const notes = ref('');
const outsourcedCostText = ref('');
const outsourcedShippingText = ref('');
const outsourcedSupplierId = ref('');
const outsourcedNotes = ref('');
const overrideText = ref('');
const manualTexts = ref<Record<string, string>>({});
const pricing = ref<PricingRecord | null>(null);
const materialOptions = ref<Record<string, Array<{ label: string; value: string }>>>({});
const materialMessages = ref<Record<string, string>>({});
const toast = useToast();
let lastWarningSignature = '';
const calculating = ref(false);
const manualCosts = ref<Record<string, number>>({});
const overrideInvalid = computed(() => overrideText.value.trim() !== '' && parseMoneyInput(overrideText.value, props.currencyUnit) === null);
const isOutsourcedService = computed(() => service.value?.fulfillmentMode === 'outsourced');
const outsourcedCostInvalid = computed(() => {
  if (!outsourcedCostText.value.trim()) return false;
  const parsed = parseMoneyInput(outsourcedCostText.value, props.currencyUnit);
  return parsed === null || parsed < 0;
});
const outsourcedShippingInvalid = computed(() => {
  if (!outsourcedShippingText.value.trim()) return false;
  const parsed = parseMoneyInput(outsourcedShippingText.value, props.currencyUnit);
  return parsed === null || parsed < 0;
});
const missingRequiredParameters = computed(() => activeParams.value.some((parameter: any) => parameterIsMissing(parameter)));
const quantityInvalid = computed(() => {
  const parsed = Number(quantity.value);
  const layoutService = service.value?.finishedSize?.quantityParameterKey;
  return !quantity.value.trim() || !Number.isFinite(parsed) || parsed <= 0 || (Boolean(layoutService) && !Number.isInteger(parsed));
});
const requiresMaterialOrMachineResolution = computed(() => Boolean(
  activeParams.value.some((parameter: any) => isMaterialParameter(parameter) || isMachineParameter(parameter) || isMachineRateParameter(parameter)) ||
  (service.value?.components || []).some((component: any) => component.enabled !== false && (component.type === 'material' || component.type === 'machine')),
));
const canAddWithoutPricingPreview = computed(() => Boolean(service.value && !requiresMaterialOrMachineResolution.value));
const canAdd = computed(() => Boolean(
  serviceId.value &&
  (pricing.value || canAddWithoutPricingPreview.value) &&
  (!calculating.value || canAddWithoutPricingPreview.value) &&
  !overrideInvalid.value &&
  !outsourcedCostInvalid.value &&
  !outsourcedShippingInvalid.value &&
  !customWidthInvalid.value &&
  !missingRequiredParameters.value &&
  !quantityInvalid.value,
));
const quantityNumber = computed(() => {
  const parsed = Number(quantity.value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
});
let calculationTimer: ReturnType<typeof setTimeout> | undefined;
let requestToken = 0;
let initializing = false;
const unitOptions = [
  'piece',
  'sheet',
  'pack',
  'roll',
  'meter',
  'liter',
  'kilogram',
  'hour',
  'job',
  'unit',
].map((value) => ({ label: value, value }));

function configuredService(id: string) {
  const normalizedID = String(id || '').trim();
  const source = props.services.find(value => String(value?.id || '').trim() === normalizedID);
  return source ? ensureRollSizeInputs({ ...source, parameters: (source.parameters || []).map((p: any) => ({ ...p })) }) : undefined;
}
const service = computed(() => configuredService(serviceId.value));
const activeParams = computed(
  () => service.value?.parameters?.filter((value: any) => value.active !== false) || [],
);
const layoutEnabled = computed(() => Boolean(service.value?.finishedSize?.quantityParameterKey));
const materialWidthMode = computed(() => usesMaterialRollWidth(service.value));
const selectedRoll = computed(() => selectedRollMaterial(service.value, props.materials, values.value));
const materialWidth = computed(() => rollWidthValue(selectedRoll.value, values.value.layout_margin_mm ?? service.value?.parameters?.find((p: any) => p.key === 'layout_margin_mm')?.defaultValue ?? '0'));
const materialWidthLimit = computed(() => rollStockWidthValue(selectedRoll.value));
const customWidthInvalid = computed(() => {
  if (!materialWidthMode.value) return false;
  const widthKey = service.value?.finishedSize?.widthParameterKey;
  const width = Number(widthKey ? values.value[widthKey] : '');
  const limit = Number(materialWidthLimit.value);
  return !Number.isFinite(width) || width <= 0 || !Number.isFinite(limit) || limit <= 0 || width > limit;
});
const layoutParams = computed(() => {
  if (!layoutEnabled.value) return [];
  const size = service.value.finishedSize;
  const keys = new Set([size.parameterKey, size.widthParameterKey, size.heightParameterKey, 'layout_gap_mm', 'layout_margin_mm'].filter(Boolean));
  return activeParams.value.filter((p: any) => keys.has(p.key));
});
const optionParams = computed(() => activeParams.value.filter((p: any) =>
  p.key !== service.value?.finishedSize?.quantityParameterKey && !layoutParams.value.includes(p),
));
const needsLayoutSetup = computed(() => !layoutEnabled.value && ['large format', 'banners & signage'].includes(String(service.value?.category || '').trim().toLowerCase()));
const categoryRequirements = computed(() => serviceCategoryRequirements(service.value?.category || ''));
const hasMaterialSetup = computed(() => (service.value?.components || []).some((component: any) => component.type === 'material'));
const hasMachineSetup = computed(() => (service.value?.components || []).some((component: any) => component.type === 'machine'));
const manualComponents = computed(
  () =>
    service.value?.components?.filter((value: any) => value.enabled && value.type === 'manual') ||
    [],
);
const priceMultiplier = computed(() => pricing.value?.batchQuantity ? 1 : quantityNumber.value);
const totalEstimatedCost = computed(() =>
  pricing.value
    ? pricing.value.roundingStepRial !== 1000
      ? roundMoneyUp(Math.round(pricing.value.estimatedCostRial * priceMultiplier.value), pricing.value.roundingStepRial)
      : Math.round(pricing.value.estimatedCostRial * priceMultiplier.value)
    : 0,
);
const totalOutsourcedCost = computed(() => {
  if (!isOutsourcedService.value) return 0;
  const item = parseMoneyInput(outsourcedCostText.value, props.currencyUnit) || 0;
  const shipping = parseMoneyInput(outsourcedShippingText.value, props.currencyUnit) || 0;
  return item + shipping;
});
const totalEstimatedCostWithOutsourcing = computed(() => totalEstimatedCost.value + totalOutsourcedCost.value);
const supplierOptions = computed(() => {
  const currentId = outsourcedSupplierId.value;
  return [
    { label: 'Select supplier…', value: '' },
    ...(props.suppliers || [])
      .filter((supplier: any) => supplier.active !== false || supplier.id === currentId)
      .map((supplier: any) => ({ label: supplier.code ? `${supplier.name} · ${supplier.code}` : supplier.name, value: supplier.id })),
  ];
});
const totalSuggestedPrice = computed(() =>
  pricing.value ? sellingPriceTotal(pricing.value.suggestedSellingPriceRial, priceMultiplier.value, pricing.value.roundingStepRial) : 0,
);
const totalEffectivePrice = computed(() =>
  pricing.value ? sellingPriceTotal(pricing.value.effectiveSellingPriceRial, priceMultiplier.value, pricing.value.roundingStepRial) : 0,
);
const totalProfit = computed(() => totalEffectivePrice.value - totalEstimatedCostWithOutsourcing.value);

function initialize() {
  initializing = true;
  const initial = props.initial;
  const selectedServiceId = String(initial?.serviceId || props.presetServiceId || props.services.find((value) => value?.active !== false)?.id || '').trim();
  const selectedService = configuredService(selectedServiceId);
  serviceId.value = String(selectedService?.id || '');
  const initialValues: Record<string, string> = {};
  if (initial?.resolvedParametersJson) {
    try {
      for (const parameter of JSON.parse(initial.resolvedParametersJson))
        initialValues[parameter.key] = parameter.value;
    } catch {
      // Invalid historical snapshots are ignored; the editor starts empty.
    }
  }
  for (const parameter of selectedService?.parameters || []) {
    if (parameter.key && initialValues[parameter.key] === undefined) {
      initialValues[parameter.key] = parameter.defaultValue || (parameter.type === 'boolean' ? 'false' : '');
    }
  }
  values.value = initialValues;
  quantity.value = initial?.quantity || '1';
  unit.value = initial?.quantityUnit || selectedService?.defaultUnit || 'piece';
  notes.value = initial?.notes || '';
  const hasInitialOutsourcedCost = initial?.outsourcedCostRial !== undefined && initial?.outsourcedCostRial !== null;
  const hasInitialOutsourcedShipping = initial?.outsourcedShippingRial !== undefined && initial?.outsourcedShippingRial !== null;
  outsourcedCostText.value = hasInitialOutsourcedCost
    ? (initial.outsourcedCostRial > 0 ? formatMoneyInput(initial.outsourcedCostRial, props.currencyUnit) : '')
    : (selectedService?.fulfillmentMode === 'outsourced' && selectedService.defaultOutsourcedCostRial > 0 ? formatMoneyInput(selectedService.defaultOutsourcedCostRial, props.currencyUnit) : '');
  outsourcedShippingText.value = hasInitialOutsourcedShipping
    ? (initial.outsourcedShippingRial > 0 ? formatMoneyInput(initial.outsourcedShippingRial, props.currencyUnit) : '')
    : (selectedService?.fulfillmentMode === 'outsourced' && selectedService.defaultOutsourcedShippingRial > 0 ? formatMoneyInput(selectedService.defaultOutsourcedShippingRial, props.currencyUnit) : '');
  outsourcedSupplierId.value = initial?.outsourcedSupplierId || '';
  outsourcedNotes.value = initial?.outsourcedNotes || '';
  overrideText.value = '';
  manualTexts.value = {};
  manualCosts.value = {};
  pricing.value = null;
  lastWarningSignature = '';
  initializing = false;
  scheduleCalculate();
}

watch(() => [props.initial, props.presetServiceId, props.services], initialize, { immediate: true });
watch([serviceId, values, () => props.services, () => props.materials], () => {
  syncMachineRateSelections();
  void refreshMaterialOptions();
}, { deep: true, immediate: true });
watch(serviceId, (next, previous) => {
  if (initializing || !previous || next === previous) return;
  const selectedService = configuredService(next);
  values.value = Object.fromEntries(
    (selectedService?.parameters || [])
      .filter((parameter: any) => parameter.key)
      .map((parameter: any) => [parameter.key, parameter.defaultValue || (parameter.type === 'boolean' ? 'false' : '')]),
  );
  quantity.value = '1';
  unit.value = selectedService?.defaultUnit || 'piece';
  manualTexts.value = {};
  manualCosts.value = {};
  outsourcedCostText.value = selectedService?.fulfillmentMode === 'outsourced' && selectedService.defaultOutsourcedCostRial > 0
    ? formatMoneyInput(selectedService.defaultOutsourcedCostRial, props.currencyUnit)
    : '';
  outsourcedShippingText.value = selectedService?.fulfillmentMode === 'outsourced' && selectedService.defaultOutsourcedShippingRial > 0
    ? formatMoneyInput(selectedService.defaultOutsourcedShippingRial, props.currencyUnit)
    : '';
  outsourcedSupplierId.value = '';
  outsourcedNotes.value = '';
  pricing.value = null;
  lastWarningSignature = '';
  scheduleCalculate();
});

function setValue(key: string, value: string) {
  values.value[key] = value;
}

function isMaterialParameter(parameter: any) {
  return parameter.type === 'material-reference' || Boolean(parameter.materialSource);
}

function isMachineParameter(parameter: any) {
  return parameter.type === 'machine-reference';
}

function isMachineRateParameter(parameter: any) {
  return parameter.type === 'choice' && (service.value?.components || []).some(
    (component: any) => component.type === 'machine' && component.rateParameterKey === parameter.key,
  );
}

function parameterIsRequired(parameter: any) {
  if (!parameter.required) return false;
  if (isMaterialParameter(parameter)) return categoryRequirements.value.material || hasMaterialSetup.value;
  if (isMachineParameter(parameter) || isMachineRateParameter(parameter)) return categoryRequirements.value.machine || hasMachineSetup.value;
  return true;
}

function configuredMachineOptions(parameter: any) {
  const configured = new Set(parameter.options || []);
  if (!configured.size) {
    const fallback = props.machines.find((machine: any) => machine.active && machine.id === parameter.defaultValue);
    return fallback ? [fallback] : [];
  }
  return props.machines.filter((machine: any) => machine.active && configured.has(machine.id));
}

function machineRateOptions(parameter: any) {
  const allowed = new Set(parameter.options || []);
  const machineComponents = (service.value?.components || [])
    .filter((component: any) => component.type === 'machine' && component.rateParameterKey === parameter.key)
  const selectedMachines = machineComponents
    .map((component: any) => {
      if (component.usageMode !== 'parameter' || !component.parameterKey) {
        return props.machines.find((machine: any) => machine.id === component.referenceId && machine.active);
      }
      const machineParameter = (service.value?.parameters || []).find((item: any) => item.key === component.parameterKey);
      const selectedID = values.value[component.parameterKey] || machineParameter?.defaultValue;
      return props.machines.find((machine: any) => machine.id === selectedID && machine.active);
    })
    .filter(Boolean);
  const candidates = selectedMachines.length
    ? selectedMachines
    : machineComponents.flatMap((component: any) => {
      const machineParameter = (service.value?.parameters || []).find((item: any) => item.key === component.parameterKey);
      return machineParameter ? configuredMachineOptions(machineParameter) : [];
    });
  const options = new Map<string, { label: string; value: string }>();
  for (const machine of candidates as any[]) {
    const rates = machine.rates?.length
      ? machine.rates.filter((rate: any) => rate.active)
      : [{ id: 'default', name: 'Standard', selectorValue: '', selectorPredefinedKey: '', rateRial: machine.rateRial, setupCostRial: machine.setupCostRial, rateBasis: machine.rateBasis, active: true }];
    for (const rate of rates) {
      const value = String(rate.selectorValue || rate.name || rate.id || '').trim();
      if (!value || (allowed.size && !allowed.has(value))) continue;
      if (!options.has(value)) options.set(value, { value, label: rate.name || value });
    }
  }
  return Array.from(options.values()).sort((left, right) => left.label.localeCompare(right.label));
}

function parameterOptions(parameter: any) {
  if (parameter.materialSource) return availableMaterialOptions(parameter);
  if (isMachineRateParameter(parameter)) return machineRateOptions(parameter);
  if (parameter.predefinedKey) {
    const configured = new Set(parameter.options || []);
    return (parameter.predefinedOptions || [])
      .filter((option: any) => option.active !== false && (!configured.size || configured.has(option.code)))
      .map((option: any) => ({ label: option.label, value: option.code }));
  }
  return (parameter.options || []).map((option: string) => ({ label: option, value: option }));
}

function canonicalMaterialAttribute(attribute: any) {
  if (!attribute) return '';
  if (attribute.valueType === 'decimal') {
    const value = String(attribute.decimalValue ?? '').trim();
    const numeric = Number(value);
    return Number.isFinite(numeric) ? String(numeric) : value;
  }
  if (attribute.valueType === 'integer') return String(attribute.integerValue ?? '').trim();
  if (attribute.valueType === 'enum') return String(attribute.enumCode ?? '').trim();
  if (attribute.valueType === 'boolean') return attribute.booleanValue ? 'true' : 'false';
  return String(attribute.textValue ?? '').trim();
}

function materialAttribute(material: any, key: string) {
  return (material.attributes || []).find((attribute: any) => attribute.key === key);
}

function materialSourceKeys(parameter: any) {
  const source = parameter.materialSource;
  if (!source) return [];
  return source.exposedAttributeKeys?.length
    ? source.exposedAttributeKeys
    : source.exposedAttributeKey
      ? [source.exposedAttributeKey]
      : [];
}

function materialSourceValue(material: any, parameter: any) {
  if (parameter.materialSource?.selectMaterial) return material.id;
  const values = materialSourceKeys(parameter).map((key: string) => canonicalMaterialAttribute(materialAttribute(material, key)));
  return values.every(Boolean) ? values.join('\u001f') : '';
}

function materialMatchesSource(material: any, parameter: any) {
  const source = parameter.materialSource;
  if (!source) return true;
  if (source.allowedKinds?.length && !source.allowedKinds.includes(material.kind)) return false;
  for (const filter of source.additionalFilters || []) {
    if (canonicalMaterialAttribute(materialAttribute(material, filter.key)) !== canonicalMaterialAttribute(filter.value)) return false;
  }
  for (const key of materialSourceKeys(parameter)) {
    if (!materialAttribute(material, key)) return false;
  }
  for (const allowed of source.allowedValues || []) {
    const attribute = materialAttribute(material, allowed.key);
    if (attribute && canonicalMaterialAttribute(attribute) === canonicalMaterialAttribute(allowed)) continue;
    if (materialSourceKeys(parameter).includes(allowed.key)) return false;
  }
  return true;
}

function inventoryMaterialOptions(parameter: any) {
  if (!parameter.materialSource) return [];
  const materialParameters = activeParams.value.filter((item: any) => isMaterialParameter(item));
  const options = new Map<string, { label: string; value: string }>();
  for (const material of props.materials.filter((item: any) => item.active)) {
    if (!materialParameters.every((item: any) => materialMatchesSource(material, item))) continue;
    const compatible = materialParameters.every((item: any) => {
      if (item.key === parameter.key) return true;
      const selected = String(values.value[item.key] || '').trim();
      if (!selected) return true;
      if (item.type === 'material-reference') return material.id === selected;
      return materialSourceValue(material, item) === selected;
    });
    if (!compatible) continue;
    const value = materialSourceValue(material, parameter);
    if (!value || options.has(value)) continue;
    options.set(value, { value, label: material.name || value });
  }
  return Array.from(options.values()).sort((left, right) => left.label.localeCompare(right.label, undefined, { numeric: true }));
}

function derivedMaterialOptions(parameter: any) {
  const saved = savedMaterialOptions(parameter);
  return saved.length ? saved : inventoryMaterialOptions(parameter);
}

function availableMaterialOptions(parameter: any) {
  const loaded = materialOptions.value[parameter.key] || [];
  return loaded.length ? loaded : derivedMaterialOptions(parameter);
}

function savedMaterialOptions(parameter: any) {
  const variants = (service.value?.materialVariants || []).filter((variant: any) => variant.active !== false);
  if (!variants.length) return [];
  const materialParameters = activeParams.value.filter((item: any) => isMaterialParameter(item));
  const options = new Map<string, { label: string; value: string; materialIds: string[] }>();
  for (const variant of variants) {
    const value = String(variant.values?.[parameter.key] || '').trim();
    if (!value) continue;
    const compatible = materialParameters.every((item: any) => {
      if (item.key === parameter.key) return true;
      const selected = String(values.value[item.key] || '').trim();
      return !selected || String(variant.values?.[item.key] || '').trim() === selected;
    });
    if (!compatible) continue;
    const material = props.materials.find((item: any) => item.active && item.id === variant.materialId);
    const label = material?.name || value;
    const existing = options.get(value);
    if (existing) {
      if (!existing.materialIds.includes(variant.materialId)) existing.materialIds.push(variant.materialId);
    } else {
      options.set(value, { value, label, materialIds: variant.materialId ? [variant.materialId] : [] });
    }
  }
  return Array.from(options.values()).map(({ label, value }) => ({ label, value }));
}

function materialOptionLabel(option: any) {
  const names = (option.materialIds || [])
    .map((id: string) => props.materials.find((material: any) => material.active && material.id === id)?.name)
    .filter(Boolean);
  return names.length ? Array.from(new Set(names)).join(' / ') : option.label;
}

function syncMaterialSelections() {
  for (const parameter of activeParams.value) {
    if (!parameter.materialSource) continue;
    const options = availableMaterialOptions(parameter);
    if (values.value[parameter.key] && !options.some((option) => option.value === values.value[parameter.key])) values.value[parameter.key] = '';
    if (!values.value[parameter.key] && parameter.defaultValue && options.some((option) => option.value === parameter.defaultValue)) values.value[parameter.key] = parameter.defaultValue;
  }
}

let materialOptionsToken = 0;
async function refreshMaterialOptions() {
  const selected = serviceId.value;
  if (!selected) { materialOptions.value = {}; materialMessages.value = {}; return; }
  if (!activeParams.value.some((parameter: any) => parameter.materialSource)) {
    materialOptions.value = {};
    materialMessages.value = {};
    return;
  }
  const token = ++materialOptionsToken;
  materialOptions.value = Object.fromEntries(
    activeParams.value
      .filter((parameter: any) => parameter.materialSource)
      .map((parameter: any) => [parameter.key, derivedMaterialOptions(parameter)]),
  );
  syncMaterialSelections();
  try {
    const materialValues = Object.fromEntries(
      activeParams.value
        .filter((parameter: any) => isMaterialParameter(parameter))
        .map((parameter: any) => [parameter.key, values.value[parameter.key] || '']),
    );
    const groups = await servicesApi.materialOptions(selected, materialValues);
    if (token !== materialOptionsToken || selected !== serviceId.value) return;
    const next: Record<string, Array<{ label: string; value: string }>> = {};
    const messages: Record<string, string> = {};
    for (const group of groups as any[]) {
      const liveOptions = (group.options || []).map((option: any) => ({ label: materialOptionLabel(option), value: option.value }));
      const fallbackOptions = materialOptions.value[group.parameterKey] || [];
      next[group.parameterKey] = liveOptions.length ? liveOptions : fallbackOptions;
      if (group.message && !next[group.parameterKey].length) messages[group.parameterKey] = group.message;
    }
    if (Object.keys(next).length) materialOptions.value = next;
    materialMessages.value = messages;
    syncMaterialSelections();
  } catch {
    // Keep the service's saved material variants available when the live
    // inventory lookup is temporarily unavailable.
    materialMessages.value = {};
    syncMaterialSelections();
  }
}

function parameterIsMissing(parameter: any) {
  return Boolean(parameterIsRequired(parameter) && !String(values.value[parameter.key] ?? '').trim());
}

function syncMachineRateSelections() {
  for (const parameter of activeParams.value.filter((item: any) => isMachineRateParameter(item))) {
    const options = machineRateOptions(parameter);
    if (!options.length || options.some((option) => option.value === values.value[parameter.key])) continue;
    const fallback = parameter.defaultValue && options.some((option) => option.value === parameter.defaultValue)
      ? parameter.defaultValue
      : options[0].value;
    if (values.value[parameter.key] !== fallback) values.value[parameter.key] = fallback;
  }
}

function updateMoneyText(text: string, key: string) {
  if (key === 'override') {
    overrideText.value = text;
    const parsed = parseMoneyInput(text, props.currencyUnit);
    if (parsed !== null) overrideText.value = formatMoneyInput(parsed, props.currencyUnit);
    return;
  }
  manualTexts.value[key] = text;
  const parsed = parseMoneyInput(text, props.currencyUnit);
  if (parsed !== null) manualCosts.value[key] = parsed;
  else delete manualCosts.value[key];
}

function updateOutsourcedMoney(text: string, key: 'cost' | 'shipping') {
  const target = key === 'cost' ? outsourcedCostText : outsourcedShippingText;
  target.value = text;
  const parsed = parseMoneyInput(text, props.currencyUnit);
  if (parsed !== null) target.value = formatMoneyInput(parsed, props.currencyUnit);
}

function scheduleCalculate() {
  requestToken += 1;
  pricing.value = null;
  calculating.value = true;
  const quantityKey = service.value?.finishedSize?.quantityParameterKey;
  if (materialWidthMode.value) {
    const widthKey = service.value.finishedSize.widthParameterKey;
    if (!values.value[widthKey] && materialWidth.value) values.value[widthKey] = materialWidth.value;
  }
  if (quantityKey && values.value[quantityKey] !== quantity.value) values.value[quantityKey] = quantity.value;
  if (calculationTimer) clearTimeout(calculationTimer);
  if (!serviceId.value || overrideInvalid.value || (materialWidthMode.value && (!materialWidth.value || customWidthInvalid.value))) {
    requestToken += 1;
    pricing.value = null;
    calculating.value = false;
    return;
  }
  calculationTimer = setTimeout(() => {
    calculationTimer = undefined;
    void calculate();
  }, 250);
}

function dependencyFreePricingPreview(): PricingRecord | null {
  const selectedService = service.value;
  if (!selectedService || activeParams.value.length || requiresMaterialOrMachineResolution.value) return null;
  const result = serviceDefaultPricingResult(selectedService, props.materials, props.machines, props.services);
  if (!result) return null;
  const override = overrideText.value.trim() ? parseMoneyInput(overrideText.value, props.currencyUnit) : null;
  const effective = override ?? result.sellingPriceRial;
  return {
    serviceId: selectedService.id,
    serviceName: selectedService.name,
    serviceCode: selectedService.code,
    parameters: [],
    components: result.lines.map((line, index) => ({
      id: `${selectedService.id}-preview-${index}`,
      name: line.name,
      type: '',
      referenceId: '',
      materialId: '',
      parameterKey: '',
      enabled: !line.missing,
      usageQuantity: '1',
      rateRial: line.amount,
      percentage: '',
      amountRial: line.amount,
      explanation: line.detail,
    })),
    estimatedCostRial: result.totalCostRial,
    suggestedSellingPriceRial: result.sellingPriceRial,
    effectiveSellingPriceRial: effective,
    profitRial: effective - result.totalCostRial,
    marginPercentage: effective ? String(((effective - result.totalCostRial) / effective) * 100) : '0',
    warnings: result.hasMissing ? ['Some service costs need configuration.'] : [],
    belowCost: effective < result.totalCostRial,
    roundingStepRial: 1000,
    finishedWidthMM: '',
    finishedHeightMM: '',
  } as unknown as PricingRecord;
}

async function calculate() {
  if (!serviceId.value || overrideInvalid.value) return;
  const token = ++requestToken;
  calculating.value = true;
  try {
    const override = overrideText.value.trim()
      ? parseMoneyInput(overrideText.value, props.currencyUnit)
      : null;
    if (overrideText.value.trim() && override === null)
      throw new Error('Enter a valid selling price');
    const localPreview = dependencyFreePricingPreview();
    if (localPreview) {
      pricing.value = localPreview;
      return;
    }
    const next = await pricingApi.calculate({
      serviceId: serviceId.value,
      quantity: quantity.value,
      parameters: values.value,
      manualCosts: manualCosts.value,
      sellingPriceOverrideRial: override,
    });
    if (token !== requestToken) return;
    pricing.value = next;
    const warning = next.belowCost ? 'Selling price is below estimated cost.' : '';
    if (warning !== lastWarningSignature) {
      if (warning) toast.warning(warning, 'Pricing');
      lastWarningSignature = warning;
    }
  } catch (value) {
    if (token === requestToken) {
      pricing.value = null;
      toast.error(String(value).replace(/^Error:\s*/, ''), 'Pricing');
    }
  } finally {
    if (token === requestToken) calculating.value = false;
  }
}

watch(
  [serviceId, values, quantity, manualCosts, overrideText],
  scheduleCalculate,
  { deep: true },
);
watch(materialWidth, scheduleCalculate);

onBeforeUnmount(() => {
  if (calculationTimer) clearTimeout(calculationTimer);
  requestToken += 1;
});

function save() {
  if (props.busy) return;
  if (!serviceId.value) {
    toast.warning('The selected service is unavailable. Close this window and choose another service.', 'Order item');
    return;
  }
  if (overrideInvalid.value) {
    toast.warning('Enter a valid selling price override.', 'Pricing');
    return;
  }
  if (outsourcedCostInvalid.value || outsourcedShippingInvalid.value) {
    toast.warning('Enter valid outsourced cost and shipping amounts.', 'Order item');
    return;
  }
  if (missingRequiredParameters.value) {
    toast.warning('Complete the required service options before adding the item.', 'Order item');
    return;
  }
  if (quantityInvalid.value) {
    toast.warning('Enter a quantity greater than zero.', 'Order item');
    return;
  }
  if (customWidthInvalid.value) {
    toast.warning('Enter a positive custom width no greater than the selected material width.', 'Order item');
    return;
  }
  if (!pricing.value && !canAddWithoutPricingPreview.value) {
    toast.warning('Wait for the live price preview before adding the item.', 'Pricing');
    return;
  }
  const override = overrideText.value.trim()
    ? parseMoneyInput(overrideText.value, props.currencyUnit)
    : null;
  emit('configured', {
    serviceId: serviceId.value,
    parameters: { ...values.value },
    manualCosts: { ...manualCosts.value },
    sellingPriceOverrideRial: override,
    quantity: quantity.value,
    quantityUnit: unit.value,
    outsourcedCostRial: isOutsourcedService.value ? (parseMoneyInput(outsourcedCostText.value, props.currencyUnit) || 0) : 0,
    outsourcedShippingRial: isOutsourcedService.value ? (parseMoneyInput(outsourcedShippingText.value, props.currencyUnit) || 0) : 0,
    outsourcedSupplierId: isOutsourcedService.value ? outsourcedSupplierId.value : '',
    outsourcedNotes: isOutsourcedService.value ? outsourcedNotes.value : '',
    notes: notes.value,
  }, pricing.value || undefined);
}
</script>

<template>
  <div data-enter-scope class="flex max-h-[min(88vh,52rem)] min-h-0 flex-col bg-base-100">
    <header class="flex shrink-0 items-start justify-between gap-3 border-b border-base-300 bg-base-100 p-4">
      <div class="min-w-0">
        <p class="text-xs font-semibold uppercase tracking-wide text-primary">
          {{ $ui(initial ? 'Reconfigure item' : 'Add service item') }}
        </p>
        <h3 class="text-base font-semibold">{{ $t("Configure service") }}</h3>
        <p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("Fill the required details. The price updates automatically.") }}</p>
      </div>
      <button class="btn btn-ghost btn-square btn-sm shrink-0" type="button" :aria-label='$t("Close service item editor")' :title='$t("Close")' @click="emit('cancel')"><X :size="18" aria-hidden="true" /></button>
    </header>

    <div class="min-h-0 flex-1 overflow-y-auto p-4">
      <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(18rem,0.72fr)]">
      <div class="min-w-0 space-y-4">
        <section class="rounded-box border border-base-300 bg-base-100 p-3">
          <div class="flex min-w-0 items-center gap-3">
            <span class="grid size-10 shrink-0 place-items-center overflow-hidden rounded-box border border-base-300 bg-base-200 bg-cover bg-center text-primary" :style="service?.imagePath ? { backgroundImage: `url('${service.imagePath}')` } : undefined">
              <Layers3 v-if="!service?.imagePath" :size="18" aria-hidden="true" />
            </span>
            <div class="min-w-0 flex-1">
              <span class="block text-[0.68rem] font-semibold uppercase tracking-wide text-primary">{{ $t("Selected service") }}</span>
              <strong class="block truncate text-sm">{{ $ui(service?.name || 'Service unavailable') }}</strong>
              <span class="block truncate text-xs text-base-content/55">{{ $ui(service?.code || service?.category || 'Close and choose another service') }}</span>
            </div>
            <span class="badge badge-ghost shrink-0 text-xs">{{ $ui(service?.defaultUnit || unit) }}</span>
          </div>
          <p v-if="!service" class="mt-2 text-xs text-error">{{ $t("This service is unavailable. Close this window and choose another service.") }}</p>
        </section>

        <section v-if="needsLayoutSetup" class="rounded-box border border-warning/40 bg-warning/5 p-4 text-sm" role="status">
          <h4 class="font-semibold">{{ $t("Print layout is not configured") }}</h4>
          <p class="mt-2">{{ $t("Edit this service → Materials → Set up roll / sheet layout, then save it. Finished size, roll rotation, and waste calculations will appear here once enabled.") }}</p>
        </section>

        <section v-if="layoutEnabled" class="space-y-4 rounded-box border border-primary/35 bg-primary/5 p-4" :aria-label='$t("Finished size and layout")'>
          <div><h4 class="font-semibold">{{ $t("Finished size & layout") }}</h4><p class="mt-1 text-xs text-base-content/65">{{ $ui(materialWidthMode ? 'Enter the custom finished width and height in mm. Width is capped at the selected roll width.' : 'Enter the finished dimensions in mm: 1000 mm = 1 m.') }} {{ $t("The preview calculates stock for the whole item quantity.") }}</p></div>
          <RollSizeFields v-if="materialWidthMode" :width="values[service.finishedSize.widthParameterKey] || ''" :height="values[service.finishedSize.heightParameterKey] || ''" :max-width="materialWidthLimit" :width-invalid="customWidthInvalid" :material-name="selectedRoll?.name" @width="setValue(service.finishedSize.widthParameterKey, $event)" @height="setValue(service.finishedSize.heightParameterKey, $event)" />
          <div class="grid gap-4 sm:grid-cols-2">
            <template v-for="parameter in layoutParams.filter((p: any) => !materialWidthMode || ![service.finishedSize.widthParameterKey, service.finishedSize.heightParameterKey].includes(p.key))" :key="parameter.id">
              <SelectField v-if="parameter.type === 'choice'" :model-value="values[parameter.key] || ''" :label="parameter.label" :invalid="parameterIsMissing(parameter)" :options="[{label: 'Select finished size…', value: ''}, ...parameterOptions(parameter)]" @update:model-value="setValue(parameter.key, $event)" />
              <FormField v-else class="gap-1"><span>{{ parameter.label }}</span><AppInput :model-value="values[parameter.key] || ''" type="number" min="0" step="any" :aria-label="parameter.label" @update:model-value="setValue(parameter.key, $event)" /></FormField>
            </template>
          </div>
          <p class="text-xs text-base-content/65">{{ $ui(service.finishedSize.allowRotation ? 'Automatic rotation enabled — the preview shows the most efficient grid orientation.' : 'Rotation locked — the original artwork orientation is preserved.') }}</p>
        </section>

        <section v-if="optionParams.length" class="rounded-box border border-base-300 bg-base-100 p-4">
          <div class="flex items-start justify-between gap-3 border-b border-base-300 pb-3">
            <div>
              <h4 class="text-sm font-semibold">{{ $t("Service options") }}</h4>
              <p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("Use the customer’s requested specifications.") }}</p>
            </div>
            <span class="badge badge-ghost shrink-0 text-xs">{{ optionParams.length }} {{ $t("option") }}{{ $ui(optionParams.length === 1 ? '' : 's') }}</span>
          </div>
          <div class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2">
            <template v-for="parameter in optionParams" :key="parameter.id">
              <SelectField
                v-if="parameter.type === 'choice'"
                :model-value="values[parameter.key] || parameter.defaultValue || ''"
                :label="parameter.label"
                :invalid="parameterIsMissing(parameter)"
                :options="[
                  { label: 'Select…', value: '' },
                  ...parameterOptions(parameter),
                ]"
                :aria-label="parameter.label"
                @update:model-value="setValue(parameter.key, $event)"
              />
              <FormField v-else-if="parameter.type === 'boolean'" class="justify-center gap-1">
                <span class="text-xs text-base-content/60">{{ parameter.label }}<em v-if="parameterIsRequired(parameter)" class="text-error"> *</em></span>
                <span class="flex h-10 items-center gap-2 rounded-field border border-base-300 bg-base-100 px-3 text-sm">
                  <input class="checkbox checkbox-sm" type="checkbox" :checked="values[parameter.key] === 'true'" @change="values[parameter.key] = ($event.target as HTMLInputElement).checked ? 'true' : 'false'" />
                  <span>{{ $ui(values[parameter.key] === 'true' ? 'Enabled' : 'Disabled') }}</span>
                </span>
              </FormField>
              <SelectField
                v-else-if="parameter.type === 'material-reference'"
                :model-value="values[parameter.key] || ''"
                :label="parameter.label"
                :invalid="parameterIsMissing(parameter)"
                :options="[
                  { label: 'Select material…', value: '' },
                  ...materials.filter((value: any) => value.active).map((value: any) => ({ label: value.name, value: value.id })),
                ]"
                :aria-label="parameter.label"
                @update:model-value="setValue(parameter.key, $event)"
              />
              <SelectField
                v-else-if="parameter.type === 'machine-reference'"
                :model-value="values[parameter.key] || ''"
                :label="parameter.label"
                :invalid="parameterIsMissing(parameter)"
                :options="[
                  { label: 'Select machine…', value: '' },
                  ...configuredMachineOptions(parameter).map((value: any) => ({ label: value.code ? `${value.name} · ${value.code}` : value.name, value: value.id })),
                ]"
                :aria-label="parameter.label"
                @update:model-value="setValue(parameter.key, $event)"
              />
              <FormField v-else class="gap-1">
                <span class="text-xs text-base-content/60">{{ parameter.label }}<em v-if="parameterIsRequired(parameter)" class="text-error"> *</em></span>
                <AppInput
                  :model-value="values[parameter.key] || ''"
                  :class="{ 'input-error': parameterIsMissing(parameter) }"
                  :type="parameter.type === 'integer' ? 'number' : 'text'"
                  :placeholder="$ui(parameter.defaultValue || parameter.type)"
                  @update:model-value="setValue(parameter.key, $event)"
                />
                <small v-if="parameter.unit" class="text-xs text-base-content/50">{{ $ui(parameter.unit) }}</small>
              </FormField>
              <p v-if="parameter.materialSource && materialMessages[parameter.key]" class="text-xs leading-5 text-warning sm:col-span-2">{{ materialMessages[parameter.key] }}</p>
            </template>
          </div>
          <p v-if="missingRequiredParameters" class="mt-3 text-xs text-warning">{{ $t("Complete the required options to add this item.") }}</p>
        </section>
        <section v-else-if="!layoutEnabled && !needsLayoutSetup" class="rounded-box border border-dashed border-base-300 bg-base-100 p-4 text-sm text-base-content/60">
          {{ $t("This service uses its standard configuration. No additional service options are required.") }}
        </section>

        <section class="rounded-box border border-base-300 bg-base-100 p-4">
          <div class="border-b border-base-300 pb-3">
            <h4 class="text-sm font-semibold">{{ $t("Item details") }}</h4>
            <p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("Set the quantity and any notes for this order line.") }}</p>
          </div>
          <div class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2">
            <FormField class="gap-1">
              <span class="text-xs text-base-content/60">{{ $t("Item quantity *") }}</span>
              <AppInput v-model="quantity" class="input w-full min-w-0" :class="{ 'input-error': quantityInvalid }" inputmode="decimal" placeholder="1" />
              <small v-if="quantityInvalid" class="text-xs text-error">{{ $ui(service?.finishedSize?.quantityParameterKey ? 'Enter a positive whole number of finished pieces.' : 'Enter a quantity greater than zero.') }}</small>
            </FormField>
            <SelectField v-model="unit" :label='$t("Unit")' :options="unitOptions" :aria-label='$t("Unit")' />
            <FormField class="gap-1 sm:col-span-2">
              <span class="text-xs text-base-content/60">{{ $t("Item notes") }} <em class="font-normal text-base-content/50">{{ $t("optional") }}</em></span>
              <AppTextarea v-model="notes" rows="3" :placeholder='$t("Item-specific notes or instructions")' />
            </FormField>
          </div>
        </section>

        <section v-if="isOutsourcedService" class="rounded-box border border-info/35 bg-info/5 p-4">
          <div class="border-b border-info/20 pb-3">
            <h4 class="text-sm font-semibold">{{ $t("Outsourced fulfillment") }}</h4>
            <p class="mt-1 text-xs leading-5 text-base-content/65">{{ $t("Optional order-specific overrides. Service defaults are prefilled; shipping is optional. Both values are recorded in estimated cost and do not change the customer's selling price.") }}</p>
          </div>
          <div class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2">
            <SelectField v-if="supplierOptions.length > 1" v-model="outsourcedSupplierId" :label='$t("Supplier")' :options="supplierOptions" :aria-label='$t("Outsourced supplier")' />
            <FormField class="gap-1" :class="(suppliers || []).length ? '' : 'sm:col-span-2'">
              <span class="text-xs text-base-content/60">{{ $t("Outsourced item cost") }} <em class="font-normal text-base-content/50">{{ $t("optional · total") }}</em></span>
              <AppInput :model-value="outsourcedCostText" :class="{ 'input-error': outsourcedCostInvalid }" :money="currencyUnit" inputmode="decimal" :placeholder="$ui(`Optional ${$ui(currencyUnit)} cost`)" @update:model-value="updateOutsourcedMoney($event, 'cost')" />
              <small v-if="outsourcedCostInvalid" class="text-xs text-error">{{ $t("Enter a valid amount.") }}</small>
            </FormField>
            <FormField class="gap-1">
              <span class="text-xs text-base-content/60">{{ $t("Shipping") }} <em class="font-normal text-base-content/50">{{ $t("optional · total") }}</em></span>
              <AppInput :model-value="outsourcedShippingText" :class="{ 'input-error': outsourcedShippingInvalid }" :money="currencyUnit" inputmode="decimal" :placeholder="$ui(`Optional ${$ui(currencyUnit)} shipping`)" @update:model-value="updateOutsourcedMoney($event, 'shipping')" />
              <small v-if="outsourcedShippingInvalid" class="text-xs text-error">{{ $t("Enter a valid amount.") }}</small>
            </FormField>
            <FormField class="gap-1 sm:col-span-2">
              <span class="text-xs text-base-content/60">{{ $t("Outsourcing notes") }} <em class="font-normal text-base-content/50">{{ $t("optional") }}</em></span>
              <AppTextarea v-model="outsourcedNotes" rows="2" :placeholder='$t("Vendor quote, delivery note, or handoff details")' />
            </FormField>
          </div>
        </section>

        <section v-if="manualComponents.length || service" class="rounded-box border border-base-300 bg-base-100 p-4">
          <div class="border-b border-base-300 pb-3">
            <h4 class="text-sm font-semibold">{{ $t("Pricing adjustments") }}</h4>
            <p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("Optional overrides. Pricing updates automatically after every change.") }}</p>
          </div>
          <div class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2">
            <FormField v-for="component in manualComponents" :key="component.id" class="gap-1">
              <span class="text-xs text-base-content/60">{{ component.name }}</span>
              <AppInput :model-value="manualTexts[component.id] || ''" :money="currencyUnit" :placeholder="$ui(`Amount in ${$ui(currencyUnit)}`)" @update:model-value="updateMoneyText($event, component.id)" />
            </FormField>
            <FormField class="gap-1" :class="manualComponents.length ? 'sm:col-span-2' : ''">
              <span class="text-xs text-base-content/60">{{ $t("Selling price override") }} <em class="font-normal text-base-content/50">{{ $t("optional") }}</em></span>
              <AppInput :model-value="overrideText" :class="{ 'input-error': overrideInvalid }" :money="currencyUnit" inputmode="decimal" :placeholder="$ui(`Optional ${$ui(currencyUnit)} price`)" @update:model-value="updateMoneyText($event, 'override')" />
              <small v-if="overrideInvalid" class="text-xs text-error">{{ $t("Enter a valid amount.") }}</small>
            </FormField>
          </div>
        </section>
      </div>

      <aside class="min-w-0 self-start rounded-box border border-base-300 bg-base-100 p-4 xl:sticky xl:top-0">
        <div class="flex items-start justify-between gap-3 border-b border-base-300 pb-3">
          <div>
            <h4 class="text-sm font-semibold">{{ $t("Live price preview") }}</h4>
            <p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("Updates automatically as inputs change.") }}</p>
          </div>
          <span class="badge badge-ghost shrink-0 gap-1 text-xs" aria-live="polite"><Calculator :size="13" aria-hidden="true" />{{ $ui(calculating ? 'Updating…' : pricing ? 'Updated' : 'Waiting') }}</span>
        </div>

        <div v-if="pricing" class="mt-4 space-y-4" :class="{ 'opacity-60': calculating }">
          <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-1">
            <div class="rounded-box border border-base-300 bg-base-200/35 p-3"><span class="block text-xs text-base-content/60">{{ $t("Estimated cost · order total") }}</span><strong class="mt-1 block text-sm">{{ formatMoney(totalEstimatedCostWithOutsourcing, currencyUnit) }}</strong><span class="mt-0.5 block text-[0.68rem] text-base-content/50">{{ formatMoney(pricing.estimatedCostRial, currencyUnit) }} / {{ $ui(pricing.batchQuantity ? 'batch' : unit) }}<template v-if="totalOutsourcedCost"> · + {{ formatMoney(totalOutsourcedCost, currencyUnit) }} {{ $t("outsourced") }}</template></span></div>
            <div class="rounded-box border border-base-300 bg-base-200/35 p-3"><span class="block text-xs text-base-content/60">{{ $t("Suggested price · order total") }}</span><strong class="mt-1 block text-sm">{{ formatMoney(totalSuggestedPrice, currencyUnit) }}</strong><span class="mt-0.5 block text-[0.68rem] text-base-content/50">{{ formatMoney(pricing.suggestedSellingPriceRial, currencyUnit) }} / {{ $ui(unit) }}</span></div>
            <div v-for="layout in pricing.layouts || []" :key="layout.materialId" class="rounded-box border border-primary/30 bg-primary/5 p-3 text-sm space-y-2">
              <strong>{{ layout.materialName }} {{ $t("· production layout") }}</strong>
              <p v-if="layout.itemsPerSheet">{{ layout.itemsPerSheet }} {{ $t("pieces per sheet ·") }} {{ layout.sheets }} {{ $t("sheets") }}</p>
              <p v-else>{{ layout.across }} {{ $t("across ×") }} {{ layout.rows }} {{ $t("rows ·") }} {{ Number(layout.lengthMM) / 1000 }} {{ $t("m of roll") }}</p>
              <p>{{ layout.consumedQuantity }} {{ $ui(layout.unit) }} {{ $t("consumed ·") }} {{ layout.wastePercent.toFixed(1) }}{{ $t("% waste") }}</p>
              <p v-if="layout.wasteCostRial > 0">{{ $t("Estimated material waste cost:") }} {{ formatMoney(layout.wasteCostRial, currencyUnit) }} {{ $t("(included in material cost)") }}</p>
              <p v-if="layout.rotated" class="text-primary">{{ $t("Rotate artwork 90° for this layout.") }}<span v-if="layout.originalLengthMM"> {{ $t("Roll usage drops from") }} {{ Number(layout.originalLengthMM) / 1000 }} {{ $t("m to") }} {{ Number(layout.lengthMM) / 1000 }} {{ $t("m.") }}</span></p>
              <p v-else>{{ $t("Original orientation uses the least material (or rotation is disabled).") }}</p>
            </div>
            <div class="rounded-box border border-primary/30 bg-primary/5 p-3"><span class="block text-xs text-base-content/60">{{ $t("Effective price · order total") }}</span><strong class="mt-1 block text-lg text-primary">{{ formatMoney(totalEffectivePrice, currencyUnit) }}</strong><span class="mt-0.5 block text-[0.68rem] text-base-content/60">{{ $ui(quantity || '1') }} {{ $ui(unit) }} · {{ formatMoney(pricing.effectiveSellingPriceRial, currencyUnit) }} / {{ $ui(pricing.batchQuantity ? 'batch' : unit) }}</span></div>
            <div class="rounded-box border border-base-300 bg-base-200/35 p-3"><span class="block text-xs text-base-content/60">{{ $t("Profit / margin") }}</span><strong class="mt-1 block text-sm" :class="totalProfit < 0 ? 'text-error' : 'text-success'">{{ formatMoney(totalProfit, currencyUnit) }} · {{ pricing.marginPercentage }}%</strong></div>
          </div>
          <p class="text-xs text-base-content/60">{{ $t("Selling prices round up to the next 100 tomans (1,000 rials).") }}</p>
          <div v-if="pricing.belowCost" class="rounded-box border border-warning/30 bg-warning/5 p-3 text-xs leading-5 text-warning">{{ $t("The selling price is below the estimated cost. Review the override or service pricing before adding.") }}</div>
          <div v-if="pricing.components.length" class="divide-y divide-base-300 rounded-box border border-base-300">
            <div v-for="component in pricing.components" :key="component.id" class="flex items-center justify-between gap-3 px-3 py-2 text-xs"><span class="min-w-0 truncate">{{ component.name }}</span><strong class="shrink-0">{{ $ui(component.enabled ? formatMoney(component.amountRial, currencyUnit) : 'Disabled') }}</strong></div>
          </div>
        </div>
        <div v-else-if="calculating" class="mt-4 rounded-box border border-dashed border-primary/40 bg-primary/5 p-5 text-center text-sm text-base-content/70" aria-live="polite">{{ $t("Updating the price preview…") }}</div>
        <EmptyState v-else compact class="mt-4" :title='$t("Price preview pending")' :description='$t("Complete the required details to see the order total.")'><template #icon><Calculator :size="21" aria-hidden="true" /></template></EmptyState>

      </aside>
    </div>
    </div>
    <footer class="flex shrink-0 items-center justify-between gap-3 border-t border-base-300 bg-base-100 p-3 sm:p-4">
      <button class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="emit('cancel')">{{ $t("Cancel") }}</button>
      <div class="flex min-w-0 items-center gap-3">
        <span class="hidden truncate text-xs text-base-content/55 sm:block" aria-live="polite">{{ $ui(calculating ? 'Updating price…' : pricing ? `Total for ${quantity || '1'} ${$ui(unit)}` : service ? 'Enter item details' : 'Service unavailable') }}</span>
        <button class="btn btn-primary btn-sm gap-2" type="button" data-enter-submit @click="save" :disabled="busy || !canAdd">
          <Plus :size="15" aria-hidden="true" />
          <span>{{ $ui(initial ? 'Replace item' : `Add item to ${$ui(documentLabel)}`) }}</span>
        </button>
      </div>
    </footer>
  </div>
</template>
