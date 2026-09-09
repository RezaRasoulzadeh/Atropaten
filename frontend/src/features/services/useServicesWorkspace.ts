import {componentNeedsRate,componentNeedsReference,componentNeedsPercentage} from './serviceFields'
import type {ServiceFilter,EditorMode,ParameterType,ParameterForm,ComponentType,ComponentForm,PricingTierForm,PricingRuleForm,ServiceForm} from './types'
import {useWorkspaceActions} from '../../composables/useWorkspaceActions'
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { materialsApi, type MaterialRecord } from '../../api/materials';
import { servicesApi, type ServiceRecord } from '../../api/services';
import { machinesApi, type MachineRecord } from '../../api/machines';
import {
  formatMoney,
  formatMoneyInput,
  parseMoneyInput,
  type CurrencyUnit,
} from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import { confirmAction, useToast } from '../../ui/feedback';

export function useServicesWorkspace(props:{ currencyUnit: CurrencyUnit },emit:(event:'notify',message:string)=>void){
const {busy,runAction}=useWorkspaceActions()
const toast = useToast();
const services = ref<ServiceRecord[]>([]);
const materials = ref<MaterialRecord[]>([]);
const machines = ref<MachineRecord[]>([]);
const selectedId = ref<string | null>(null);
const searchQuery = ref('');
const serviceFilter = ref<ServiceFilter>('Active');
const editorMode = ref<EditorMode>(null);
const form = ref<ServiceForm>(emptyForm());
const isLoading = ref(false);
const isSaving = ref(false);
const validationAttempted = ref(false);
const selectedService = computed(
  () => services.value.find((service) => service.id === selectedId.value) ?? null,
);
const filteredServices = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  return services.value.filter((service) => {
    const matchesFilter =
      serviceFilter.value === 'All' ||
      (serviceFilter.value === 'Active' ? service.active : !service.active);
    const matchesSearch =
      !query ||
      [service.name, service.code, service.category, service.description].some((value) =>
        value.toLowerCase().includes(query),
      );
    return matchesFilter && matchesSearch;
  });
});
onMounted(loadServices);
watch(
  () => props.currencyUnit,
  () => {
    for (const component of form.value.components)
      component.rateInput = formatMoneyInput(component.rateRial, props.currencyUnit);
    const rule = form.value.pricingRule;
    if (rule) {
      rule.fixedPriceInput = formatMoneyInput(rule.fixedPriceRial, props.currencyUnit);
      rule.fixedMarginInput = formatMoneyInput(rule.fixedMarginRial, props.currencyUnit);
      rule.perUnitRateInput = formatMoneyInput(rule.perUnitRateRial, props.currencyUnit);
      for (const tier of rule.tiers)
        tier.priceInput = formatMoneyInput(tier.priceRial, props.currencyUnit);
    }
  },
);
function emptyForm(): ServiceForm {
  return {
    name: '',
    code: '',
    category: '',
    description: '',
    parameters: [],
    components: [],
    pricingRule: {
      id: '',
      type: 'manual',
      fixedPriceRial: 0,
      fixedPriceInput: '',
      markupPercentage: '20',
      fixedMarginRial: 0,
      fixedMarginInput: '',
      perUnitRateRial: 0,
      perUnitRateInput: '',
      parameterKey: '',
      tiers: [],
    },
  };
}
function emptyParameter(type: ParameterType = 'integer'): ParameterForm {
  return {
    id: `draft-parameter-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    key: '',
    label: '',
    type,
    required: false,
    defaultValue: '',
    options: [],
    minValue: null,
    maxValue: null,
    unit: '',
  };
}
function emptyComponent(type: ComponentType = 'fixed'): ComponentForm {
  return {
    id: `draft-component-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    name: '',
    type,
    referenceId: '',
    usageMode: 'fixed',
    parameterKey: '',
    usageQuantity: '1',
    multiplier: '1',
    rateRial: 0,
    rateInput: '',
    percentage: '',
    rateBasis: 'hour',
    enabled: true,
    notes: '',
  };
}
async function loadServices() {
  isLoading.value = true;
  try {
    const [serviceData, materialData, machineData] = await Promise.all([
      servicesApi.list(true),
      materialsApi.list(false),
      machinesApi.list(false),
    ]);
    services.value = serviceData;
    materials.value = materialData;
    machines.value = machineData;
  } catch (error) {
    toast.error(errorMessageFrom(error, 'Services could not be loaded.'), 'Services');
  } finally {
    isLoading.value = false;
  }
}
function selectService(id: string) {
  selectedId.value = id;
  editorMode.value = null;
}
function startCreate() {
  editorMode.value = 'create';
  selectedId.value = null;
  form.value = emptyForm();
  validationAttempted.value = false;
}
function startEdit() {
  const service = selectedService.value;
  if (!service) return;
  form.value = {
    name: service.name,
    code: service.code,
    category: service.category,
    description: service.description,
    parameters: service.parameters.map((parameter) => ({
      id: parameter.id,
      key: parameter.key,
      label: parameter.label,
      type: parameter.type as ParameterType,
      required: parameter.required,
      defaultValue: parameter.defaultValue,
      options: Array.isArray(parameter.options) ? [...parameter.options] : [],
      minValue: parameter.minValue ?? null,
      maxValue: parameter.maxValue ?? null,
      unit: parameter.unit,
    })),
    components: service.components.map((component) => ({
      id: component.id,
      name: component.name,
      type: component.type as ComponentType,
      referenceId: component.referenceId,
      usageMode: component.usageMode as 'fixed' | 'parameter',
      parameterKey: component.parameterKey,
      usageQuantity: component.usageQuantity || '1',
      multiplier: component.multiplier,
      rateRial: component.rateRial,
      rateInput: formatMoneyInput(component.rateRial, props.currencyUnit),
      percentage: component.percentage,
      rateBasis: component.rateBasis || 'hour',
      enabled: component.enabled,
      notes: component.notes,
    })),
    pricingRule: service.pricingRule
      ? {
          id: service.pricingRule.id,
          type: service.pricingRule.type,
          fixedPriceRial: service.pricingRule.fixedPriceRial,
          fixedPriceInput: formatMoneyInput(service.pricingRule.fixedPriceRial, props.currencyUnit),
          markupPercentage: service.pricingRule.markupPercentage,
          fixedMarginRial: service.pricingRule.fixedMarginRial,
          fixedMarginInput: formatMoneyInput(
            service.pricingRule.fixedMarginRial,
            props.currencyUnit,
          ),
          perUnitRateRial: service.pricingRule.perUnitRateRial,
          perUnitRateInput: formatMoneyInput(
            service.pricingRule.perUnitRateRial,
            props.currencyUnit,
          ),
          parameterKey: service.pricingRule.parameterKey,
          tiers: (Array.isArray(service.pricingRule.tiers) ? service.pricingRule.tiers : []).map((tier) => ({
            position: tier.position,
            minimumQuantity: tier.minimumQuantity,
            priceRial: tier.priceRial,
            priceInput: formatMoneyInput(tier.priceRial, props.currencyUnit),
          })),
        }
      : null,
  };
  validationAttempted.value = false;
  editorMode.value = 'edit';
}
function numericParameters() {
  return form.value.parameters.filter(
    (parameter) => parameter.type === 'integer' || parameter.type === 'decimal',
  );
}
function defaultComponentName(type: ComponentType) {
  return {
    material: 'Material cost',
    machine: 'Machine cost',
    labor: 'Labor cost',
    outsourced: 'Outsourced work',
    fixed: 'Fixed cost',
    overhead: 'Overhead',
    waste: 'Waste allowance',
    manual: 'Manual cost',
  }[type];
}
function normalizeComponent(component: ComponentForm) {
  const defaultNames = Object.values({
    material: 'Material cost',
    machine: 'Machine cost',
    labor: 'Labor cost',
    outsourced: 'Outsourced work',
    fixed: 'Fixed cost',
    overhead: 'Overhead',
    waste: 'Waste allowance',
    manual: 'Manual cost',
  });
  if (!component.name.trim() || defaultNames.includes(component.name.trim()))
    component.name = defaultComponentName(component.type);
  if (component.type === 'material' || component.type === 'machine') {
    component.rateRial = 0;
    component.rateInput = '';
    component.percentage = '';
    component.rateBasis = '';
    component.usageMode = 'fixed';
    component.parameterKey = '';
  } else if (component.type === 'overhead' || component.type === 'waste') {
    component.referenceId = '';
    component.usageMode = 'fixed';
    component.parameterKey = '';
    component.rateRial = 0;
    component.rateInput = '';
    component.multiplier = '1';
    component.rateBasis = '';
  } else {
    component.referenceId = '';
    component.usageMode = 'fixed';
    component.parameterKey = '';
    component.percentage = '';
    if (component.type !== 'labor' && component.type !== 'outsourced') component.rateBasis = '';
  }
}
function updateComponentType(component: ComponentForm) {
  normalizeComponent(component);
}
function addComponent(type: ComponentType = 'fixed') {
  form.value.components.push(emptyComponent(type));
  normalizeComponent(form.value.components[form.value.components.length - 1]);
}
function removeComponent(index: number) {
  form.value.components.splice(index, 1);
}
function moveComponent(index: number, direction: -1 | 1) {
  const target = index + direction;
  if (target < 0 || target >= form.value.components.length) return;
  const [component] = form.value.components.splice(index, 1);
  form.value.components.splice(target, 0, component);
}
function updateComponentRate(component: ComponentForm) {
  const value = component.rateInput;
  const parsed = parseMoneyInput(value, props.currencyUnit);
  if (parsed !== null) {
    component.rateRial = parsed;
    component.rateInput = formatMoneyInput(parsed, props.currencyUnit);
  }
}
function updateGroupedMoney(target: any, textKey: string, valueKey: string, value: string) {
  target[textKey] = value;
  const parsed = parseMoneyInput(value, props.currencyUnit);
  if (parsed !== null) {
    target[valueKey] = parsed;
    target[textKey] = formatMoneyInput(parsed, props.currencyUnit);
  }
}
function normalizePricingRule() {
  const rule = form.value.pricingRule;
  if (!rule) return;
  if (rule.type !== 'per-unit' && rule.type !== 'quantity-tiers') rule.parameterKey = '';
  if (rule.type !== 'quantity-tiers') rule.tiers = [];
}
function addPricingTier() {
  const rule = form.value.pricingRule;
  if (rule)
    rule.tiers.push({
      position: rule.tiers.length,
      minimumQuantity: rule.tiers.length ? '10' : '0',
      priceRial: 0,
      priceInput: '',
    });
}
function removePricingTier(index: number) {
  const rule = form.value.pricingRule;
  if (!rule) return;
  rule.tiers.splice(index, 1);
  rule.tiers.forEach((tier, position) => {
    tier.position = position;
  });
}
function componentSummary(component: {
  type: string;
  name: string;
  referenceId: string;
  usageMode: string;
  parameterKey: string;
  multiplier: string;
  rateRial: number;
  percentage: string;
  enabled: boolean;
  rateBasis: string;
}) {
  const source =
    component.type === 'material'
      ? component.usageMode === 'parameter' && !component.referenceId
        ? `selected by ${component.parameterKey || 'material parameter'}`
        : materials.value.find((item) => item.id === component.referenceId)?.name || 'Material'
      : component.type === 'machine'
        ? machines.value.find((item) => item.id === component.referenceId)?.name || 'Machine'
        : component.type;
  if (component.type === 'overhead' || component.type === 'waste')
    return `${component.name || typeLabel(component.type)} · ${component.percentage}%${component.enabled ? '' : ' · disabled'}`;
  const usage =
    component.type === 'material' && component.usageMode === 'parameter' && !component.referenceId
      ? `${component.parameterKey || 'material'} · ${component.multiplier}× quantity`
      : component.usageMode === 'parameter'
      ? `${component.parameterKey} × ${component.multiplier}`
      : `fixed × ${component.multiplier}`;
  const rate = componentNeedsRate(component.type as ComponentType)
    ? ` · ${formatMoney(component.rateRial, props.currencyUnit)}${component.rateBasis ? ` / ${component.rateBasis}` : ''}`
    : '';
  return `${component.name || typeLabel(component.type)} · ${source} · ${usage}${rate}${component.enabled ? '' : ' · disabled'}`;
}
function cancelEditor() {
  editorMode.value = null;
  validationAttempted.value = false;
}
function addParameter(type: ParameterType = 'integer') {
  form.value.parameters.push(emptyParameter(type));
}
function syncParameterKey(parameter: ParameterForm) {
  if (parameter.key.trim()) return;
  const base = parameter.label
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '') || 'input';
  let key = base;
  let suffix = 2;
  while (form.value.parameters.some((item) => item !== parameter && item.key === key)) {
    key = `${base}_${suffix++}`;
  }
  parameter.key = key;
}
function removeParameter(index: number) {
  form.value.parameters.splice(index, 1);
}
function moveParameter(index: number, direction: -1 | 1) {
  const target = index + direction;
  if (target < 0 || target >= form.value.parameters.length) return;
  const [parameter] = form.value.parameters.splice(index, 1);
  form.value.parameters.splice(target, 0, parameter);
}
function normalizeParameter(parameter: ParameterForm) {
  if (parameter.type === 'choice') {
    parameter.minValue = null;
    parameter.maxValue = null;
    parameter.defaultValue = parameter.options.includes(parameter.defaultValue)
      ? parameter.defaultValue
      : '';
    return;
  }
  parameter.options = [];
  if (parameter.type !== 'integer' && parameter.type !== 'decimal') {
    parameter.minValue = null;
    parameter.maxValue = null;
  }
  if (parameter.type === 'boolean') {
    parameter.defaultValue =
      parameter.defaultValue === 'true' || parameter.defaultValue === 'false'
        ? parameter.defaultValue
        : '';
  } else if (parameter.type === 'material-reference') {
    parameter.defaultValue = materials.value.some(
      (material) => material.id === parameter.defaultValue,
    )
      ? parameter.defaultValue
      : '';
  } else if (parameter.type === 'integer' || parameter.type === 'decimal') {
    const decimalDefault = /^\d+(?:\.\d{1,6})?$/.test(parameter.defaultValue);
    const integerDefault = /^\d+$/.test(parameter.defaultValue);
    if (
      (parameter.type === 'integer' && !integerDefault) ||
      (parameter.type === 'decimal' && !decimalDefault)
    )
      parameter.defaultValue = '';
    if (parameter.type === 'integer') {
      if (parameter.minValue && !/^\d+$/.test(parameter.minValue)) parameter.minValue = null;
      if (parameter.maxValue && !/^\d+$/.test(parameter.maxValue)) parameter.maxValue = null;
    }
  } else {
    parameter.defaultValue = '';
  }
}
function addOption(parameter: ParameterForm) {
  parameter.options.push(`Option ${parameter.options.length + 1}`);
}
function removeOption(parameter: ParameterForm, index: number) {
  const removed = parameter.options.splice(index, 1)[0];
  if (parameter.defaultValue === removed) parameter.defaultValue = '';
}
async function saveService() {
return runAction(async () => {
  validationAttempted.value = true;
  for (const parameter of form.value.parameters) syncParameterKey(parameter);
  await nextTick();
  const firstInvalid = document.querySelector<HTMLElement>('#service-editor [aria-invalid="true"], #service-editor :invalid');
  firstInvalid?.focus();
  if (!form.value.name.trim()) {
    toast.error('Enter a service name.', 'Services');
    return;
  }
  isSaving.value = true;
  const wasEditing = editorMode.value === 'edit';
  try {
    const payload = {
      name: form.value.name,
      code: form.value.code,
      category: form.value.category,
      description: form.value.description,
      parameters: form.value.parameters.map((parameter) => ({ ...parameter })),
      components: form.value.components.map((component) => ({
        id: component.id,
        name: component.name,
        type: component.type,
        referenceId: component.referenceId,
        usageMode: component.usageMode,
        parameterKey: component.parameterKey,
        usageQuantity: component.usageQuantity,
        multiplier: component.multiplier,
        rateRial: component.rateRial,
        percentage: component.percentage,
        rateBasis: component.rateBasis,
        enabled: component.enabled,
        notes: component.notes,
      })),
      pricingRule: form.value.pricingRule
        ? {
            id: form.value.pricingRule.id,
            type: form.value.pricingRule.type,
            fixedPriceRial: form.value.pricingRule.fixedPriceRial,
            markupPercentage: form.value.pricingRule.markupPercentage,
            fixedMarginRial: form.value.pricingRule.fixedMarginRial,
            perUnitRateRial: form.value.pricingRule.perUnitRateRial,
            parameterKey: form.value.pricingRule.parameterKey,
            tiers: form.value.pricingRule.tiers.map((tier) => ({
              position: tier.position,
              minimumQuantity: tier.minimumQuantity,
              priceRial: tier.priceRial,
            })),
          }
        : null,
    };
    const saved =
      wasEditing && selectedId.value
        ? await servicesApi.update(selectedId.value, payload)
        : await servicesApi.create(payload);
    const existingIndex = services.value.findIndex((service) => service.id === saved.id);
    if (existingIndex >= 0) services.value.splice(existingIndex, 1, saved);
    else services.value.push(saved);
    selectedId.value = saved.id;
    editorMode.value = null;
    emit('notify', wasEditing ? 'Service updated.' : 'Service created.');
  } catch (error) {
    toast.error(errorMessageFrom(error, 'Service could not be saved.'), 'Services');
  } finally {
    isSaving.value = false;
  }

});
}
async function setActive(active: boolean) {
return runAction(async () => {
  const service = selectedService.value;
  if (!service) return;
  try {
    const updated = active
      ? await servicesApi.reactivate(service.id)
      : await servicesApi.archive(service.id);
    const index = services.value.findIndex((item) => item.id === updated.id);
    if (index >= 0) services.value.splice(index, 1, updated);
    emit('notify', active ? 'Service reactivated.' : 'Service deactivated.');
  } catch (error) {
    toast.error(errorMessageFrom(error, 'Service status could not be changed.'), 'Services');
  }

});
}
async function remove() {
  return runAction(async () => {
    const service = selectedService.value;
    if (
      !service ||
      !(await confirmAction({
        title: 'Delete service',
        message: 'Delete this service permanently? Services used by orders or invoices will be archived instead.',
        confirmLabel: 'Delete service',
        danger: true,
      }))
    )
      return;
    try {
      await servicesApi.remove(service.id);
      services.value = services.value.filter((item) => item.id !== service.id);
      selectedId.value = null;
      editorMode.value = null;
      emit('notify', 'Service deleted.');
    } catch (error) {
      if (!isDeletionProtected(error)) {
        toast.error(errorMessageFrom(error, 'Service could not be deleted.'), 'Services');
        return;
      }
      try {
        const archived = await servicesApi.archive(service.id);
        const index = services.value.findIndex((item) => item.id === archived.id);
        if (index >= 0) services.value.splice(index, 1, archived);
        emit('notify', 'Service is in use, so it was archived instead.');
      } catch (archiveError) {
        toast.error(errorMessageFrom(archiveError, 'Service could not be deleted or archived.'), 'Services');
      }
    }
  });
}
function isDeletionProtected(error: unknown) {
  const detail = errorMessageFrom(error, '').toLowerCase();
  return detail.includes('archive it instead') || detail.includes('delete protected');
}
function typeLabel(type: string) {
  return (
    {
      integer: 'Integer',
      decimal: 'Decimal',
      boolean: 'Boolean',
      choice: 'Choice',
      'material-reference': 'Material reference',
    }[type] ?? type
  );
}
function dateLabel(value: string) {
  try {
    return formatDateTime(value);
  } catch {
    return 'Unknown date';
  }
}
function errorMessageFrom(error: unknown, fallback: string): string {
  return error instanceof Error && error.message
    ? error.message
    : typeof error === 'string'
      ? error
      : fallback;
}
return {busy,runAction,services,materials,machines,selectedId,searchQuery,serviceFilter,editorMode,form,isLoading,isSaving,validationAttempted,selectedService,filteredServices,emptyForm,emptyParameter,emptyComponent,loadServices,selectService,startCreate,startEdit,numericParameters,normalizeComponent,updateComponentType,addComponent,removeComponent,moveComponent,updateComponentRate,updateGroupedMoney,normalizePricingRule,addPricingTier,removePricingTier,componentSummary,cancelEditor,addParameter,removeParameter,moveParameter,syncParameterKey,normalizeParameter,addOption,removeOption,saveService,setActive,remove,typeLabel,dateLabel,errorMessageFrom}
}
