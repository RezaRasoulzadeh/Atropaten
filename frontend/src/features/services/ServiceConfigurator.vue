<script setup lang="ts">
import AppInput from '../../components/ui/AppInput.vue';
import FormField from '../../components/ui/FormField.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTable from '../../components/ui/DataTable.vue';
import { ref, watch } from 'vue';
import { Calculator, LoaderCircle, RotateCcw } from 'lucide-vue-next';
import type { ServiceRecord } from '../../api/services';
import type { MaterialRecord } from '../../api/materials';
import type { MachineRecord } from '../../api/machines';
import { pricingApi, type PricingRecord } from '../../api/pricing';
import {
  formatMoney,
  formatMoneyInput,
  parseMoneyInput,
  type CurrencyUnit,
} from '../../utils/currency';
import SelectField from '../../components/ui/SelectField.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import { useToast } from '../../ui/feedback';

const props = defineProps<{
  service: ServiceRecord;
  materials: MaterialRecord[];
  machines: MachineRecord[];
  currencyUnit: CurrencyUnit;
}>();
const values = ref<Record<string, string>>({});
const overrideText = ref('');
const overrideRial = ref<number | null>(null);
const manualCosts = ref<Record<string, number>>({});
const manualTexts = ref<Record<string, string>>({});
const result = ref<PricingRecord | null>(null);
const loading = ref(false);
const toast = useToast();
let lastWarningSignature = '';
let requestToken = 0;

function resetValues() {
  const next: Record<string, string> = {};
  for (const parameter of props.service.parameters)
    next[parameter.key] = parameter.defaultValue || '';
  values.value = next;
  overrideText.value = '';
  overrideRial.value = null;
  manualCosts.value = {};
  manualTexts.value = {};
  result.value = null;
  lastWarningSignature = '';
  void calculate();
}

watch(() => props.service.id, resetValues, { immediate: true });
watch(
  values,
  () => {
    void calculate();
  },
  { deep: true },
);
watch(
  () => props.currencyUnit,
  () => {
    if (overrideRial.value !== null)
      overrideText.value = formatMoneyInput(overrideRial.value, props.currencyUnit);
  },
);

async function calculate() {
  const token = ++requestToken;
  loading.value = true;
  try {
    const next = await pricingApi.calculate({
      serviceId: props.service.id,
      parameters: values.value,
      manualCosts: manualCosts.value,
      sellingPriceOverrideRial: overrideRial.value,
    });
    if (token === requestToken) {
      result.value = next;
      const warnings = next.warnings ?? [];
      const signature = [...(next.belowCost ? ['Selling price is below estimated cost.'] : []), ...warnings].join('\u0000');
      if (signature !== lastWarningSignature) {
        if (next.belowCost) toast.warning('Selling price is below estimated cost.', 'Pricing');
        for (const warning of warnings) toast.warning(warning, 'Pricing');
        lastWarningSignature = signature;
      }
    }
  } catch (cause) {
    if (token === requestToken) {
      result.value = null;
      toast.error(cause instanceof Error ? cause.message : 'Pricing could not be calculated.', 'Pricing');
    }
  } finally {
    if (token === requestToken) loading.value = false;
  }
}

function updateManual(componentId: string, value: string) {
  manualTexts.value[componentId] = value;
  if (value.trim() === '') {
    delete manualCosts.value[componentId];
    void calculate();
    return;
  }
  const parsed = parseMoneyInput(value, props.currencyUnit);
  if (parsed !== null) {
    manualCosts.value[componentId] = parsed;
    manualTexts.value[componentId] = formatMoneyInput(parsed, props.currencyUnit);
    void calculate();
  }
}

function updateOverride(value: string) {
  const parsed = value.trim() === '' ? null : parseMoneyInput(value, props.currencyUnit);
  if (value.trim() !== '' && parsed === null) {
    overrideText.value = value;
    return;
  }
  overrideRial.value = parsed;
  overrideText.value = parsed === null ? '' : formatMoneyInput(parsed, props.currencyUnit);
  void calculate();
}

function resetOverride() {
  overrideRial.value = null;
  overrideText.value = '';
  void calculate();
}

function updateBoolean(key: string, event: Event) {
  values.value[key] = (event.target as HTMLInputElement).checked ? 'true' : 'false';
}

function money(value: number) {
  return formatMoney(value, props.currencyUnit);
}
function signedMoney(value: number) {
  return `${value < 0 ? '−' : ''}${money(Math.abs(value))}`;
}
function typeLabel(type: string) {
  return (
    (
      {
        integer: 'Integer',
        decimal: 'Decimal',
        boolean: 'Boolean',
        choice: 'Choice',
        'material-reference': 'Material',
        'machine-reference': 'Machine',
      } as Record<string, string>
    )[type] ?? type
  );
}
</script>

<template>
  <section aria-label="Service pricing configurator" class="min-w-0 space-y-5">
    <header class="flex min-w-0 flex-wrap items-start justify-between gap-3 border-b border-base-300 pb-4">
      <div class="min-w-0">
        <p class="text-xs font-semibold uppercase tracking-wide text-primary">Live pricing preview</p>
        <h2 class="mt-1 flex min-w-0 items-center gap-2 text-base font-semibold">
          <Calculator :size="17" :stroke-width="1.8" aria-hidden="true" />
          <span class="truncate">{{ service.name }}</span>
        </h2>
        <p class="mt-1 text-xs leading-5 text-base-content/60">Resolve the persisted parameters and inspect the ordered cost explanation.</p>
      </div>
      <span v-if="loading" class="badge badge-ghost shrink-0 gap-1 text-xs">
        <LoaderCircle :size="13" :stroke-width="1.8" aria-hidden="true" />Calculating
      </span>
    </header>
    <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(19rem,0.8fr)]">
      <div class="min-w-0 space-y-4 rounded-box border border-base-300 bg-base-200/25 p-4">
        <div class="flex min-w-0 items-start justify-between gap-3">
          <div>
            <h3 class="text-sm font-semibold">Parameters</h3>
            <p class="mt-1 text-xs text-base-content/60">Values used to calculate this service.</p>
          </div>
          <span class="badge badge-ghost shrink-0 text-xs">{{ service.parameters.length }} inputs</span>
        </div>
        <div v-if="service.parameters.length" class="min-w-0 space-y-3">
          <FormField class="gap-1" v-for="parameter in service.parameters" :key="parameter.id"
            ><span>{{ parameter.label }}<em v-if="parameter.required">required</em></span>
            <AppInput
              class="input w-full min-w-0"
              v-if="parameter.type === 'integer' || parameter.type === 'decimal'"
              v-model="values[parameter.key]"
              type="text"
              inputmode="decimal"
              :placeholder="parameter.defaultValue || 'Enter value'"
            />
            <SelectField
              v-else-if="parameter.type === 'choice'"
              v-model="values[parameter.key]"
              :aria-label="parameter.label"
              :options="[
                { label: `Select ${parameter.label.toLowerCase()}`, value: '' },
                ...parameter.options.map((option) => ({ label: option, value: option })),
              ]"
            />
            <SelectField
              v-else-if="parameter.type === 'material-reference'"
              v-model="values[parameter.key]"
              :aria-label="parameter.label"
              :options="[
                { label: 'Select material', value: '' },
                ...materials.map((material) => ({
                  label: `${material.name}${material.sku ? ` · ${material.sku}` : ''}`,
                  value: material.id,
                })),
              ]"
            />
            <SelectField
              v-else-if="parameter.type === 'machine-reference'"
              v-model="values[parameter.key]"
              :aria-label="parameter.label"
              :options="[
                { label: 'Select machine', value: '' },
                ...machines.filter((machine) => machine.active).map((machine) => ({
                  label: `${machine.name}${machine.code ? ` · ${machine.code}` : ''}`,
                  value: machine.id,
                })),
              ]"
            />
            <span v-else
              ><input
                class="checkbox"
                :checked="values[parameter.key] === 'true'"
                type="checkbox"
                @change="updateBoolean(parameter.key, $event)"
              />
              Enabled</span
            >
            <small
              v-if="parameter.unit || parameter.minValue || parameter.maxValue"
              class="block text-xs leading-5 text-base-content/60"
              >{{ parameter.unit
              }}<span v-if="parameter.minValue"> · min {{ parameter.minValue }}</span
              ><span v-if="parameter.maxValue"> · max {{ parameter.maxValue }}</span></small
            >
          </FormField>
        </div>
        <EmptyState v-else compact title="No operator parameters" description="This service can be priced without additional operator input."><template #icon><Calculator :size="21" aria-hidden="true" /></template></EmptyState>
      </div>
      <div class="min-w-0 space-y-4 rounded-box border border-base-300 bg-base-200/25 p-4">
        <div class="flex min-w-0 items-start justify-between gap-3">
          <div>
          <h3 class="text-sm font-semibold">Price position</h3>
          <p class="mt-1 text-xs text-base-content/60">Estimated economics for the current inputs.</p>
          </div>
          <span v-if="result" class="badge badge-ghost shrink-0 text-xs">{{ result.marginPercentage }}% margin</span>
        </div>
        <div v-if="result" class="grid min-w-0 gap-2 sm:grid-cols-2 xl:grid-cols-1">
          <div class="rounded-box border border-base-300 bg-base-100 p-3">
            <span class="block text-xs text-base-content/60">Estimated cost</span><strong class="mt-1 block text-sm">{{ money(result.estimatedCostRial) }}</strong>
          </div>
          <div class="rounded-box border border-base-300 bg-base-100 p-3">
            <span class="block text-xs text-base-content/60">Suggested price</span><strong class="mt-1 block text-sm text-primary">{{ money(result.suggestedSellingPriceRial) }}</strong>
          </div>
          <div class="rounded-box border border-primary/30 bg-primary/5 p-3">
            <span class="block text-xs text-base-content/60">Effective price</span><strong class="mt-1 block text-sm text-primary">{{ money(result.effectiveSellingPriceRial) }}</strong>
          </div>
          <div class="rounded-box border border-base-300 bg-base-100 p-3">
            <span class="block text-xs text-base-content/60">Profit</span><strong class="mt-1 block text-sm" :class="{ 'text-error': result.profitRial < 0, 'text-success': result.profitRial > 0 }">{{ signedMoney(result.profitRial) }}</strong>
          </div>
        </div>
        <EmptyState v-else compact title="No price preview yet" description="Enter the current inputs and calculate a price to see the estimate."><template #icon><Calculator :size="21" aria-hidden="true" /></template></EmptyState>
        <div class="rounded-box border border-base-300 bg-base-100 p-3"
          ><span class="text-xs font-semibold">Selling price override <em class="font-normal text-base-content/55">optional</em></span>
          <div class="mt-2 flex min-w-0 items-end gap-2">
            <FormField :label="`Use ${typeLabel('fixed')} rule suggestion`"><AppInput
              :model-value="overrideText"
              :money="props.currencyUnit"
              type="text"
              inputmode="decimal"
              :placeholder="`Use ${typeLabel('fixed')} rule suggestion`"
              @update:model-value="updateOverride"
            /></FormField><button
              class="btn btn-ghost btn-sm"
              v-if="overrideText"
              type="button"
              aria-label="Clear selling price override"
              @click="resetOverride"
            >
              <RotateCcw :size="14" :stroke-width="1.8" />
            </button></div
        ></div>
        <div
          v-if="service.components.some((component) => component.type === 'manual')"
          class="min-w-0 space-y-3 rounded-box border border-base-300 bg-base-100 p-3"
        >
          <div class="min-w-0">
            <h3 class="text-sm font-semibold">Manual costs</h3>
            <span class="mt-1 block text-xs text-base-content/60">Optional inputs</span>
          </div>
          <FormField
            class="gap-1"
            v-for="component in service.components.filter((item) => item.type === 'manual')"
            :key="component.id"
            ><span>{{ component.name }}</span
            ><AppInput
              :model-value="manualTexts[component.id] || ''"
              :money="props.currencyUnit"
              inputmode="decimal"
              placeholder="0"
              @update:model-value="updateManual(component.id, $event)"
          /></FormField>
        </div>
      </div>
    </div>
    <div v-if="result" class="min-w-0 space-y-3">
      <div class="min-w-0 space-y-3">
        <h3 class="text-sm font-semibold">Ordered cost breakdown</h3>
        <span
          >{{ result.components.filter((component) => component.enabled).length }} enabled
          components</span
        >
      </div>
      <div v-if="result.components.length">
        <DataTable
          ><thead>
            <tr>
              <th>Component</th>
              <th>Basis</th>
              <th>Explanation</th>
              <th class="text-end">Amount</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="component in result.components"
              :key="component.id"
              :class="{ 'opacity-50': !component.enabled }"
            >
              <DataTableCell
                ><span class="block font-medium">{{ component.name }}</span
                ><span class="block text-xs text-base-content/60">{{
                  component.type
                }}</span></DataTableCell
              ><DataTableCell numeric
                >{{ component.enabled ? component.usageQuantity : '—'
                }}<span v-if="component.percentage !== '0'">
                  · {{ component.percentage }}%</span
                ></DataTableCell
              ><DataTableCell>{{ component.explanation }}</DataTableCell
              ><DataTableCell numeric>{{ money(component.amountRial) }}</DataTableCell>
            </tr>
          </tbody></DataTable
        >
      </div>
      <EmptyState v-else compact title="No cost breakdown yet" description="The calculated service has no cost components to display."><template #icon><Calculator :size="21" aria-hidden="true" /></template></EmptyState>
    </div>
  </section>
</template>
