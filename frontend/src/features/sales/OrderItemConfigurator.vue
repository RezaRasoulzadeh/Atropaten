<script setup lang="ts">
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import FormField from '../../components/ui/FormField.vue';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { Calculator, Layers3, Plus, X } from 'lucide-vue-next';
import SelectField from '../../components/ui/SelectField.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import { pricingApi, type PricingRecord } from '../../api/pricing';
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

const props = withDefaults(
  defineProps<{
    services: any[];
    materials: any[];
    machines: any[];
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
  configured: [payload: OrderItemPayload];
  cancel: [];
}>();

const serviceId = ref('');
const values = ref<Record<string, string>>({});
const quantity = ref('');
const unit = ref('unit');
const notes = ref('');
const overrideText = ref('');
const manualTexts = ref<Record<string, string>>({});
const pricing = ref<PricingRecord | null>(null);
const toast = useToast();
let lastWarningSignature = '';
const calculating = ref(false);
const manualCosts = ref<Record<string, number>>({});
const overrideInvalid = computed(() => overrideText.value.trim() !== '' && parseMoneyInput(overrideText.value, props.currencyUnit) === null);
const missingRequiredParameters = computed(() => activeParams.value.some((parameter: any) => parameter.required && !String(values.value[parameter.key] ?? '').trim()));
const quantityInvalid = computed(() => {
  const parsed = Number(quantity.value);
  return !quantity.value.trim() || !Number.isFinite(parsed) || parsed <= 0;
});
const canAdd = computed(() => Boolean(serviceId.value && pricing.value && !calculating.value && !overrideInvalid.value && !missingRequiredParameters.value && !quantityInvalid.value));
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

const service = computed(() => props.services.find((value) => value.id === serviceId.value));
const activeParams = computed(
  () => service.value?.parameters?.filter((value: any) => value.active !== false) || [],
);
const manualComponents = computed(
  () =>
    service.value?.components?.filter((value: any) => value.enabled && value.type === 'manual') ||
    [],
);
const totalEstimatedCost = computed(() =>
  pricing.value
    ? pricing.value.roundingStepRial !== 1000
      ? roundMoneyUp(Math.round(pricing.value.estimatedCostRial * quantityNumber.value), pricing.value.roundingStepRial)
      : Math.round(pricing.value.estimatedCostRial * quantityNumber.value)
    : 0,
);
const totalSuggestedPrice = computed(() =>
  pricing.value ? sellingPriceTotal(pricing.value.suggestedSellingPriceRial, quantityNumber.value, pricing.value.roundingStepRial) : 0,
);
const totalEffectivePrice = computed(() =>
  pricing.value ? sellingPriceTotal(pricing.value.effectiveSellingPriceRial, quantityNumber.value, pricing.value.roundingStepRial) : 0,
);
const totalProfit = computed(() => totalEffectivePrice.value - totalEstimatedCost.value);

function initialize() {
  initializing = true;
  const initial = props.initial;
  const selectedServiceId = initial?.serviceId || props.presetServiceId || props.services.find((value) => value.active)?.id || '';
  const selectedService = props.services.find((value) => value.id === selectedServiceId);
  serviceId.value = selectedServiceId;
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
  overrideText.value = '';
  manualTexts.value = {};
  manualCosts.value = {};
  pricing.value = null;
  lastWarningSignature = '';
  initializing = false;
  scheduleCalculate();
}

watch(() => [props.initial, props.presetServiceId, props.services], initialize, { immediate: true });
watch(serviceId, (next, previous) => {
  if (initializing || !previous || next === previous) return;
  const selectedService = props.services.find((value) => value.id === next);
  values.value = Object.fromEntries(
    (selectedService?.parameters || [])
      .filter((parameter: any) => parameter.key)
      .map((parameter: any) => [parameter.key, parameter.defaultValue || (parameter.type === 'boolean' ? 'false' : '')]),
  );
  quantity.value = '1';
  unit.value = selectedService?.defaultUnit || 'piece';
  manualTexts.value = {};
  manualCosts.value = {};
  pricing.value = null;
  lastWarningSignature = '';
  scheduleCalculate();
});

function setValue(key: string, value: string) {
  values.value[key] = value;
}

function parameterIsMissing(parameter: any) {
  return Boolean(parameter.required && !String(values.value[parameter.key] ?? '').trim());
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

function scheduleCalculate() {
  if (calculationTimer) clearTimeout(calculationTimer);
  if (!serviceId.value || overrideInvalid.value) {
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
    const next = await pricingApi.calculate({
      serviceId: serviceId.value,
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
  [serviceId, values, manualCosts, overrideText],
  scheduleCalculate,
  { deep: true },
);

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
  if (missingRequiredParameters.value) {
    toast.warning('Complete the required service options before adding the item.', 'Order item');
    return;
  }
  if (quantityInvalid.value) {
    toast.warning('Enter a quantity greater than zero.', 'Order item');
    return;
  }
  if (!pricing.value) {
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
    notes: notes.value,
  });
}
</script>

<template>
  <div data-enter-scope class="flex max-h-[min(88vh,52rem)] min-h-0 flex-col bg-base-100">
    <header class="flex shrink-0 items-start justify-between gap-3 border-b border-base-300 bg-base-100 p-4">
      <div class="min-w-0">
        <p class="text-xs font-semibold uppercase tracking-wide text-primary">
          {{ initial ? 'Reconfigure item' : 'Add service item' }}
        </p>
        <h3 class="text-base font-semibold">Configure service</h3>
        <p class="mt-1 text-xs leading-5 text-base-content/60">Fill the required details. The price updates automatically.</p>
      </div>
      <button class="btn btn-ghost btn-square btn-sm shrink-0" type="button" aria-label="Close service item editor" title="Close" @click="emit('cancel')"><X :size="18" aria-hidden="true" /></button>
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
              <span class="block text-[0.68rem] font-semibold uppercase tracking-wide text-primary">Selected service</span>
              <strong class="block truncate text-sm">{{ service?.name || 'Service unavailable' }}</strong>
              <span class="block truncate text-xs text-base-content/55">{{ service?.code || service?.category || 'Close and choose another service' }}</span>
            </div>
            <span class="badge badge-ghost shrink-0 text-xs">{{ service?.defaultUnit || unit }}</span>
          </div>
          <p v-if="!service" class="mt-2 text-xs text-error">This service is unavailable. Close this window and choose another service.</p>
        </section>

        <section v-if="activeParams.length" class="rounded-box border border-base-300 bg-base-100 p-4">
          <div class="flex items-start justify-between gap-3 border-b border-base-300 pb-3">
            <div>
              <h4 class="text-sm font-semibold">Service options</h4>
              <p class="mt-1 text-xs leading-5 text-base-content/60">Use the customer’s requested specifications.</p>
            </div>
            <span class="badge badge-ghost shrink-0 text-xs">{{ activeParams.length }} option{{ activeParams.length === 1 ? '' : 's' }}</span>
          </div>
          <div class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2">
            <template v-for="parameter in activeParams" :key="parameter.id">
              <SelectField
                v-if="parameter.type === 'choice'"
                :model-value="values[parameter.key] || parameter.defaultValue || ''"
                :label="parameter.label"
                :invalid="parameterIsMissing(parameter)"
                :options="[
                  { label: 'Select…', value: '' },
                  ...parameter.options.map((option: string) => ({ label: option, value: option })),
                ]"
                :aria-label="parameter.label"
                @update:model-value="setValue(parameter.key, $event)"
              />
              <FormField v-else-if="parameter.type === 'boolean'" class="justify-center gap-1">
                <span class="text-xs text-base-content/60">{{ parameter.label }}<em v-if="parameter.required" class="text-error"> *</em></span>
                <span class="flex h-10 items-center gap-2 rounded-field border border-base-300 bg-base-100 px-3 text-sm">
                  <input class="checkbox checkbox-sm" type="checkbox" :checked="values[parameter.key] === 'true'" @change="values[parameter.key] = ($event.target as HTMLInputElement).checked ? 'true' : 'false'" />
                  <span>{{ values[parameter.key] === 'true' ? 'Enabled' : 'Disabled' }}</span>
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
                  ...machines.filter((value: any) => value.active).map((value: any) => ({ label: value.code ? `${value.name} · ${value.code}` : value.name, value: value.id })),
                ]"
                :aria-label="parameter.label"
                @update:model-value="setValue(parameter.key, $event)"
              />
              <FormField v-else class="gap-1">
                <span class="text-xs text-base-content/60">{{ parameter.label }}<em v-if="parameter.required" class="text-error"> *</em></span>
                <AppInput
                  :model-value="values[parameter.key] || ''"
                  :class="{ 'input-error': parameterIsMissing(parameter) }"
                  :type="parameter.type === 'integer' ? 'number' : 'text'"
                  :placeholder="parameter.defaultValue || parameter.type"
                  @update:model-value="setValue(parameter.key, $event)"
                />
                <small v-if="parameter.unit" class="text-xs text-base-content/50">{{ parameter.unit }}</small>
              </FormField>
            </template>
          </div>
          <p v-if="missingRequiredParameters" class="mt-3 text-xs text-warning">Complete the required options to add this item.</p>
        </section>
        <section v-else class="rounded-box border border-dashed border-base-300 bg-base-100 p-4 text-sm text-base-content/60">
          This service uses its standard configuration. No additional service options are required.
        </section>

        <section class="rounded-box border border-base-300 bg-base-100 p-4">
          <div class="border-b border-base-300 pb-3">
            <h4 class="text-sm font-semibold">Item details</h4>
            <p class="mt-1 text-xs leading-5 text-base-content/60">Set the quantity and any notes for this order line.</p>
          </div>
          <div class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2">
            <FormField class="gap-1">
              <span class="text-xs text-base-content/60">Item quantity *</span>
              <AppInput v-model="quantity" class="input w-full min-w-0" :class="{ 'input-error': quantityInvalid }" inputmode="decimal" placeholder="1" />
              <small v-if="quantityInvalid" class="text-xs text-error">Enter a quantity greater than zero.</small>
            </FormField>
            <SelectField v-model="unit" label="Unit" :options="unitOptions" aria-label="Unit" />
            <FormField class="gap-1 sm:col-span-2">
              <span class="text-xs text-base-content/60">Item notes <em class="font-normal text-base-content/50">optional</em></span>
              <AppTextarea v-model="notes" rows="3" placeholder="Item-specific notes or instructions" />
            </FormField>
          </div>
        </section>

        <section v-if="manualComponents.length || service" class="rounded-box border border-base-300 bg-base-100 p-4">
          <div class="border-b border-base-300 pb-3">
            <h4 class="text-sm font-semibold">Pricing adjustments</h4>
            <p class="mt-1 text-xs leading-5 text-base-content/60">Optional overrides. Pricing updates automatically after every change.</p>
          </div>
          <div class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2">
            <FormField v-for="component in manualComponents" :key="component.id" class="gap-1">
              <span class="text-xs text-base-content/60">{{ component.name }}</span>
              <AppInput :model-value="manualTexts[component.id] || ''" :money="currencyUnit" :placeholder="`Amount in ${currencyUnit}`" @update:model-value="updateMoneyText($event, component.id)" />
            </FormField>
            <FormField class="gap-1" :class="manualComponents.length ? 'sm:col-span-2' : ''">
              <span class="text-xs text-base-content/60">Selling price override <em class="font-normal text-base-content/50">optional</em></span>
              <AppInput :model-value="overrideText" :class="{ 'input-error': overrideInvalid }" :money="currencyUnit" inputmode="decimal" :placeholder="`Optional ${currencyUnit} price`" @update:model-value="updateMoneyText($event, 'override')" />
              <small v-if="overrideInvalid" class="text-xs text-error">Enter a valid amount.</small>
            </FormField>
          </div>
        </section>
      </div>

      <aside class="min-w-0 self-start rounded-box border border-base-300 bg-base-100 p-4 xl:sticky xl:top-0">
        <div class="flex items-start justify-between gap-3 border-b border-base-300 pb-3">
          <div>
            <h4 class="text-sm font-semibold">Live price preview</h4>
            <p class="mt-1 text-xs leading-5 text-base-content/60">Updates automatically as inputs change.</p>
          </div>
          <span class="badge badge-ghost shrink-0 gap-1 text-xs" aria-live="polite"><Calculator :size="13" aria-hidden="true" />{{ calculating ? 'Updating…' : pricing ? 'Updated' : 'Waiting' }}</span>
        </div>

        <div v-if="pricing" class="mt-4 space-y-4" :class="{ 'opacity-60': calculating }">
          <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-1">
            <div class="rounded-box border border-base-300 bg-base-200/35 p-3"><span class="block text-xs text-base-content/60">Estimated cost · order total</span><strong class="mt-1 block text-sm">{{ formatMoney(totalEstimatedCost, currencyUnit) }}</strong><span class="mt-0.5 block text-[0.68rem] text-base-content/50">{{ formatMoney(pricing.estimatedCostRial, currencyUnit) }} / {{ unit }}</span></div>
            <div class="rounded-box border border-base-300 bg-base-200/35 p-3"><span class="block text-xs text-base-content/60">Suggested price · order total</span><strong class="mt-1 block text-sm">{{ formatMoney(totalSuggestedPrice, currencyUnit) }}</strong><span class="mt-0.5 block text-[0.68rem] text-base-content/50">{{ formatMoney(pricing.suggestedSellingPriceRial, currencyUnit) }} / {{ unit }}</span></div>
            <div class="rounded-box border border-primary/30 bg-primary/5 p-3"><span class="block text-xs text-base-content/60">Effective price · order total</span><strong class="mt-1 block text-lg text-primary">{{ formatMoney(totalEffectivePrice, currencyUnit) }}</strong><span class="mt-0.5 block text-[0.68rem] text-base-content/60">{{ quantity || '1' }} {{ unit }} · {{ formatMoney(pricing.effectiveSellingPriceRial, currencyUnit) }} / {{ unit }}</span></div>
            <div class="rounded-box border border-base-300 bg-base-200/35 p-3"><span class="block text-xs text-base-content/60">Profit / margin</span><strong class="mt-1 block text-sm" :class="totalProfit < 0 ? 'text-error' : 'text-success'">{{ formatMoney(totalProfit, currencyUnit) }} · {{ pricing.marginPercentage }}%</strong></div>
          </div>
          <p class="text-xs text-base-content/60">Selling prices round up to the next 100 tomans (1,000 rials).</p>
          <div v-if="pricing.belowCost" class="rounded-box border border-warning/30 bg-warning/5 p-3 text-xs leading-5 text-warning">The selling price is below the estimated cost. Review the override or service pricing before adding.</div>
          <div v-if="pricing.components.length" class="divide-y divide-base-300 rounded-box border border-base-300">
            <div v-for="component in pricing.components" :key="component.id" class="flex items-center justify-between gap-3 px-3 py-2 text-xs"><span class="min-w-0 truncate">{{ component.name }}</span><strong class="shrink-0">{{ component.enabled ? formatMoney(component.amountRial, currencyUnit) : 'Disabled' }}</strong></div>
          </div>
        </div>
        <div v-else-if="calculating" class="mt-4 rounded-box border border-dashed border-primary/40 bg-primary/5 p-5 text-center text-sm text-base-content/70" aria-live="polite">Updating the price preview…</div>
        <EmptyState v-else compact class="mt-4" title="Price preview pending" description="Complete the required details to see the order total."><template #icon><Calculator :size="21" aria-hidden="true" /></template></EmptyState>

      </aside>
    </div>
    </div>
    <footer class="flex shrink-0 items-center justify-between gap-3 border-t border-base-300 bg-base-100 p-3 sm:p-4">
      <button class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="emit('cancel')">Cancel</button>
      <div class="flex min-w-0 items-center gap-3">
        <span class="hidden truncate text-xs text-base-content/55 sm:block" aria-live="polite">{{ calculating ? 'Updating price…' : pricing ? `Total for ${quantity || '1'} ${unit}` : service ? 'Enter item details' : 'Service unavailable' }}</span>
        <button class="btn btn-primary btn-sm gap-2" type="button" data-enter-submit @click="save" :disabled="busy || !canAdd">
          <Plus :size="15" aria-hidden="true" />
          <span>{{ initial ? 'Replace item' : `Add item to ${documentLabel}` }}</span>
        </button>
      </div>
    </footer>
  </div>
</template>
