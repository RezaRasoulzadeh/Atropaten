<script setup lang="ts">
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import FormField from '../../components/ui/FormField.vue';
import { computed, nextTick, ref, watch } from 'vue';
import { Calculator, Plus } from 'lucide-vue-next';
import SelectField from '../../components/ui/SelectField.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import { pricingApi, type PricingRecord } from '../../api/pricing';
import type { OrderItemPayload } from '../../api/orders';
import {
  formatMoney,
  formatMoneyInput,
  parseMoneyInput,
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
const pricingPreview = ref<HTMLElement | null>(null);
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
const serviceOptions = computed(() =>
  props.services
    .filter((value) => value.active || value.id === serviceId.value)
    .map((value) => ({
      label: value.code ? `${value.name} · ${value.code}` : value.name,
      value: value.id,
    })),
);
const activeParams = computed(
  () => service.value?.parameters?.filter((value: any) => value.active !== false) || [],
);
const manualComponents = computed(
  () =>
    service.value?.components?.filter((value: any) => value.enabled && value.type === 'manual') ||
    [],
);

function initialize() {
  const initial = props.initial;
  serviceId.value = initial?.serviceId || props.services.find((value) => value.active)?.id || '';
  values.value = {};
  if (initial?.resolvedParametersJson) {
    try {
      for (const parameter of JSON.parse(initial.resolvedParametersJson))
        values.value[parameter.key] = parameter.value;
    } catch {
      // Invalid historical snapshots are ignored; the editor starts empty.
    }
  }
  quantity.value = initial?.quantity || '';
  unit.value = initial?.quantityUnit || 'unit';
  notes.value = initial?.notes || '';
  overrideText.value = '';
  manualTexts.value = {};
  manualCosts.value = {};
  pricing.value = null;
  lastWarningSignature = '';
}

watch(() => [props.initial, props.services], initialize, { immediate: true });
watch(serviceId, () => {
  values.value = {};
  pricing.value = null;
});

function setValue(key: string, value: string) {
  values.value[key] = value;
  pricing.value = null;
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
}

async function calculate() {
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
    pricing.value = next;
    const warning = next.belowCost ? 'Selling price is below estimated cost.' : '';
    if (warning !== lastWarningSignature) {
      if (warning) toast.warning(warning, 'Pricing');
      lastWarningSignature = warning;
    }
    await nextTick();
    pricingPreview.value?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  } catch (value) {
    toast.error(String(value).replace(/^Error:\s*/, ''), 'Pricing');
  } finally {
    calculating.value = false;
  }
}

function save() {
  if (props.busy) return;
  if (!pricing.value) {
    toast.warning('Calculate the item before adding it.', 'Pricing');
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
  <div class="border-t border-base-300 bg-base-200/40 p-4">
    <header class="mb-4 flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="text-xs font-semibold uppercase tracking-wide text-primary">
          {{ initial ? 'Reconfigure item' : 'Add service item' }}
        </p>
        <h3 class="text-base font-semibold">Configure service</h3>
        <p class="text-xs text-base-content/60">
          Pricing is calculated by the persisted service definition.
        </p>
      </div>
      <button class="btn btn-outline btn-sm" type="button" @click="emit('cancel')">Cancel</button>
    </header>

    <div class="grid gap-4 lg:grid-cols-2">
      <SelectField
        v-model="serviceId"
        label="Service"
        :options="serviceOptions"
        aria-label="Service"
      />

      <template v-for="parameter in activeParams" :key="parameter.id">
        <SelectField
          v-if="parameter.type === 'choice'"
          :model-value="values[parameter.key] || parameter.defaultValue || ''"
          :label="parameter.label"
          :options="[
            { label: 'Select…', value: '' },
            ...parameter.options.map((option: string) => ({ label: option, value: option })),
          ]"
          :aria-label="parameter.label"
          @update:model-value="setValue(parameter.key, $event)"
        />
        <FormField v-else-if="parameter.type === 'boolean'" class="justify-center gap-1">
          <span class="text-xs text-base-content/60">{{ parameter.label }}</span>
          <span
            class="flex h-10 items-center gap-2 rounded-field border border-base-300 bg-base-100 px-3 text-sm"
          >
            <input
              class="checkbox checkbox-sm"
              type="checkbox"
              :checked="values[parameter.key] === 'true'"
              @change="
                values[parameter.key] = ($event.target as HTMLInputElement).checked
                  ? 'true'
                  : 'false'
              "
            />
            <span>{{ values[parameter.key] === 'true' ? 'Enabled' : 'Disabled' }}</span>
          </span>
        </FormField>
        <SelectField
          v-else-if="parameter.type === 'material-reference'"
          :model-value="values[parameter.key] || ''"
          :label="parameter.label"
          :options="[
            { label: 'Select material…', value: '' },
            ...materials
              .filter((value: any) => value.active)
              .map((value: any) => ({ label: value.name, value: value.id })),
          ]"
          :aria-label="parameter.label"
          @update:model-value="setValue(parameter.key, $event)"
        />
        <SelectField
          v-else-if="parameter.type === 'machine-reference'"
          :model-value="values[parameter.key] || ''"
          :label="parameter.label"
          :options="[
            { label: 'Select machine…', value: '' },
            ...machines
              .filter((value: any) => value.active)
              .map((value: any) => ({ label: value.code ? `${value.name} · ${value.code}` : value.name, value: value.id })),
          ]"
          :aria-label="parameter.label"
          @update:model-value="setValue(parameter.key, $event)"
        />
        <FormField v-else class="gap-1">
          <span class="text-xs text-base-content/60">{{ parameter.label }}</span>
          <AppInput
            :model-value="values[parameter.key] || ''"
            :type="parameter.type === 'integer' ? 'number' : 'text'"
            :placeholder="parameter.defaultValue || parameter.type"
            @update:model-value="setValue(parameter.key, $event)"
          />
          <small v-if="parameter.unit" class="text-xs text-base-content/50">{{
            parameter.unit
          }}</small>
        </FormField>
      </template>

      <EmptyState v-if="!activeParams.length" compact class="lg:col-span-2" title="No dynamic parameters" description="This service uses its standard configuration."><template #icon><Calculator :size="21" aria-hidden="true" /></template></EmptyState>

      <FormField class="gap-1">
        <span class="text-xs text-base-content/60">Item quantity</span>
        <AppInput
          class="input w-full min-w-0"
          v-model="quantity"
          placeholder="Defaults from quantity parameter"
        />
      </FormField>
      <SelectField v-model="unit" label="Unit" :options="unitOptions" aria-label="Unit" />

      <template v-if="manualComponents.length">
        <FormField v-for="component in manualComponents" :key="component.id" class="gap-1">
          <span class="text-xs text-base-content/60">{{ component.name }}</span>
          <AppInput
            :model-value="manualTexts[component.id] || ''"
            :money="currencyUnit"
            :placeholder="`Amount in ${currencyUnit}`"
            @update:model-value="updateMoneyText($event, component.id)"
          />
        </FormField>
      </template>

      <FormField class="gap-1 lg:col-span-2">
        <span class="text-xs text-base-content/60">Notes</span>
        <AppTextarea
          v-model="notes"
          rows="2"
          placeholder="Item-specific notes"
        />
      </FormField>

      <div class="flex flex-wrap items-end gap-3 lg:col-span-2">
        <FormField class="min-w-52 flex-1 gap-1">
          <span class="text-xs text-base-content/60">Selling price override</span>
          <AppInput
            :model-value="overrideText"
            :money="currencyUnit"
            inputmode="decimal"
            :placeholder="`Optional ${currencyUnit} price`"
            @update:model-value="updateMoneyText($event, 'override')"
          />
        </FormField>
        <button
          class="btn btn-primary gap-2"
          type="button"
          :disabled="calculating || !serviceId"
          @click="calculate"
        >
          <Calculator :size="15" aria-hidden="true" />
          <span>{{ calculating ? 'Calculating…' : 'Calculate price' }}</span>
        </button>
      </div>
    </div>

    <div
      v-if="pricing"
      ref="pricingPreview"
      class="mt-4 rounded-box border border-base-300 bg-base-100 p-4"
    >
      <div class="flex flex-wrap items-center justify-between gap-2">
        <strong>Accepted pricing preview</strong>
      </div>
      <div class="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-4">
        <div>
          <span class="block text-xs text-base-content/60">Estimated cost</span
          ><strong class="block text-sm">{{
            formatMoney(pricing.estimatedCostRial, currencyUnit)
          }}</strong>
        </div>
        <div>
          <span class="block text-xs text-base-content/60">Suggested price</span
          ><strong class="block text-sm">{{
            formatMoney(pricing.suggestedSellingPriceRial, currencyUnit)
          }}</strong>
        </div>
        <div>
          <span class="block text-xs text-base-content/60">Effective price</span
          ><strong class="block text-sm text-primary">{{
            formatMoney(pricing.effectiveSellingPriceRial, currencyUnit)
          }}</strong>
        </div>
        <div>
          <span class="block text-xs text-base-content/60">Profit / margin</span
          ><strong class="block text-sm"
            >{{ formatMoney(pricing.profitRial, currencyUnit) }} ·
            {{ pricing.marginPercentage }}%</strong
          >
        </div>
      </div>
      <div v-if="pricing.components.length" class="mt-4 divide-y divide-base-300 rounded-box border border-base-300">
        <div
          v-for="component in pricing.components"
          :key="component.id"
          class="flex items-center justify-between gap-3 px-3 py-2 text-xs"
        >
          <span>{{ component.name }}</span>
          <strong>{{
            component.enabled ? formatMoney(component.amountRial, currencyUnit) : 'Disabled'
          }}</strong>
        </div>
      </div>
      <EmptyState v-else compact title="No cost breakdown yet" description="The calculated price has no cost components to display."><template #icon><Calculator :size="21" aria-hidden="true" /></template></EmptyState>
      <div class="mt-4 flex justify-end">
        <button class="btn btn-primary gap-2" type="button" @click="save" :disabled="busy || calculating">
          <Plus :size="15" aria-hidden="true" />
          <span>{{ initial ? 'Replace item' : `Add item to ${documentLabel}` }}</span>
        </button>
      </div>
    </div>
    <EmptyState v-else compact title="No pricing preview yet" description="Choose a service and calculate its price to preview this item."><template #icon><Calculator :size="21" aria-hidden="true" /></template></EmptyState>
  </div>
</template>
