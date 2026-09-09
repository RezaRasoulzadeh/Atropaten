<script setup lang="ts">
import ServiceParameterEditor from './ServiceParameterEditor.vue'
import ServiceCostEditor from './ServiceCostEditor.vue'
import {componentNeedsRate,componentNeedsReference,componentNeedsPercentage} from './serviceFields'
import type {ServiceFilter,EditorMode,ParameterType,ParameterForm,ComponentType,ComponentForm,PricingTierForm,PricingRuleForm,ServiceForm} from './types'
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
import FormGrid from '../../components/ui/FormGrid.vue';
import InspectorShell from '../../components/layout/InspectorShell.vue';
import InspectorSection from '../../components/layout/InspectorSection.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppTextarea from '../../components/ui/AppTextarea.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import {
  Archive,
  ArrowLeft,
  ChevronDown,
  ChevronUp,
  Edit3,
  ListPlus,
  Package,
  Plus,
  RotateCcw,
  Save,
  SlidersHorizontal,
  Trash2,
} from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';
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
import ServiceConfigurator from './ServiceConfigurator.vue';

const emit = defineEmits<{ notify: [message: string] }>();
const props = defineProps<{ currencyUnit: CurrencyUnit }>();
import {useServicesWorkspace} from './useServicesWorkspace'
const {busy,runAction,services,materials,machines,selectedId,searchQuery,serviceFilter,editorMode,form,isLoading,isSaving,validationAttempted,selectedService,filteredServices,emptyForm,emptyParameter,emptyComponent,loadServices,selectService,startCreate,startEdit,numericParameters,normalizeComponent,updateComponentType,addComponent,removeComponent,moveComponent,updateComponentRate,updateGroupedMoney,normalizePricingRule,addPricingTier,removePricingTier,componentSummary,cancelEditor,addParameter,removeParameter,moveParameter,syncParameterKey,normalizeParameter,addOption,removeOption,saveService,setActive,remove,typeLabel,dateLabel}=useServicesWorkspace(props,emit)
const newInputType = ref<ParameterType | ''>('');
const newCostType = ref<ComponentType | ''>('');
function addInput(type: string) {
  if (!type) return;
  addParameter(type as ParameterType);
  newInputType.value = '';
}
function addCost(type: string) {
  if (!type) return;
  addComponent(type as ComponentType);
  newCostType.value = '';
}
function backToServices() {
  selectedId.value = null;
  cancelEditor();
}
watch(
  [() => selectedId.value, () => editorMode.value],
  () => {
    void nextTick(() => {
      document.querySelector('main')?.scrollTo({ top: 0, behavior: 'auto' });
    });
  },
);
</script>

<template>
  <div v-if="!selectedService && !editorMode" class="min-w-0 space-y-3">
      <WorkspaceStickyStack>
        <WorkspaceHeader
          title="Services"
          eyebrow="Catalog / sellable operations"
          description="Define reusable work with operator-facing parameters for future pricing."
          ><button class="btn btn-primary" type="button" @click="startCreate">
            <Plus :size="16" :stroke-width="1.8" aria-hidden="true" />New service
          </button></WorkspaceHeader
        >
        <SearchFilterBar><template #search><SearchField
            v-model="searchQuery"
            label="Search services"
            placeholder="Search service, code, or category"
          /></template><template #filters><SelectField
            v-model="serviceFilter"
            label="Status"
            aria-label="Filter services by status"
            :options="['Active', 'Archived', 'All'].map((value) => ({ label: value, value }))"
          /></template><template #count><span class="whitespace-nowrap"
            >{{ filteredServices.length }} of {{ services.length }} services</span
          ></template></SearchFilterBar>
      </WorkspaceStickyStack>

      <RegisterList
        title="Service register"
        subtitle="Select a service to inspect its operator parameters."
        :count="filteredServices.length"
      >
        <LoadingState v-if="isLoading" label="Loading records…" />
        <div v-else-if="filteredServices.length">
          <RegisterRow
            v-for="service in filteredServices"
            :key="service.id"
            :selected="selectedId === service.id"
            @activate="selectService(service.id)"
          >
            <template #icon><Package :size="17" :stroke-width="1.8" aria-hidden="true" /></template>
            <template #identity>
              <div class="flex min-w-0 items-center justify-between gap-3">
                <div class="min-w-0">
                  <strong class="block truncate text-sm">{{ service.name }}</strong>
                  <span class="block truncate text-xs text-base-content/60">{{ service.code || 'No code' }} · {{ service.category || 'Uncategorized' }}</span>
                </div>
                <div class="flex shrink-0 items-center gap-2">
                  <StatusBadge :label="service.active ? 'Active' : 'Archived'" :tone="service.active ? 'green' : 'slate'" />
                  <span class="hidden text-xs text-base-content/60 md:inline">{{ dateLabel(service.updatedAt) }}</span>
                </div>
              </div>
            </template>
            <template #meta>
              <div class="mt-2 grid min-w-0 grid-cols-1 gap-x-5 gap-y-1.5 text-xs sm:grid-cols-2">
                <div><span class="block text-base-content/50">Configuration</span><span class="block text-base-content/80">{{ service.parameters.length }} {{ service.parameters.length === 1 ? 'parameter' : 'parameters' }} · {{ service.parameters.filter((parameter) => parameter.required).length }} required</span></div>
                <div><span class="block text-base-content/50">Pricing inputs</span><span class="block text-base-content/80">{{ service.components.length }} cost {{ service.components.length === 1 ? 'component' : 'components' }}</span></div>
              </div>
            </template>
          </RegisterRow>
        </div>
        <EmptyState
          v-else
          :title="services.length ? 'No services match this view' : 'No services yet'"
          :description="services.length ? 'Try another status or search term.' : 'Create the first reusable operation for your shop.'"
        >
          <template #icon><SlidersHorizontal :size="22" :stroke-width="1.8" aria-hidden="true" /></template>
          <template v-if="!services.length" #action><button class="btn btn-primary btn-sm" type="button" @click="startCreate"><Plus :size="15" :stroke-width="1.8" aria-hidden="true" />Create service</button></template>
        </EmptyState>
      </RegisterList>
    </div>

  <div v-else class="min-w-0 space-y-4" aria-label="Service workspace">
      <WorkspaceStickyStack :flush="true">
        <WorkspaceHeader
          :title="editorMode === 'create' ? 'New service' : selectedService?.name || 'Service'"
          eyebrow="Catalog / service workspace"
          :description="
            editorMode === 'create'
              ? 'Create a reusable service for orders and production.'
              : 'Review the service definition, test its inputs, and manage its lifecycle.'
          "
        >
          <template v-if="selectedService && !editorMode" #title-suffix>
            <StatusBadge
              :label="selectedService.active ? 'Active' : 'Archived'"
              :tone="selectedService.active ? 'green' : 'slate'"
            />
          </template>
          <button class="btn btn-ghost btn-sm gap-2" type="button" aria-label="Back to services" @click="backToServices">
            <ArrowLeft :size="17" :stroke-width="1.8" aria-hidden="true" />
            <span class="hidden sm:inline">Services</span>
          </button>
          <template v-if="editorMode">
            <button class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="cancelEditor">Cancel</button>
            <button class="btn btn-primary btn-sm gap-2" type="submit" form="service-editor" :disabled="busy || isSaving">
              <Save :size="15" :stroke-width="1.8" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save service' }}
            </button>
          </template>
          <template v-else-if="selectedService">
            <button class="btn btn-outline btn-error btn-sm gap-2" type="button" :disabled="busy" @click="remove">
              <Trash2 :size="14" :stroke-width="1.8" aria-hidden="true" />Delete
            </button>
            <button v-if="selectedService.active" class="btn btn-outline btn-warning btn-sm gap-2" type="button" :disabled="busy" @click="setActive(false)">
              <Archive :size="14" :stroke-width="1.8" aria-hidden="true" />Archive
            </button>
            <button v-else class="btn btn-outline btn-success btn-sm gap-2" type="button" :disabled="busy" @click="setActive(true)">
              <RotateCcw :size="14" :stroke-width="1.8" aria-hidden="true" />Reactivate
            </button>
            <button class="btn btn-outline btn-sm gap-2" type="button" :disabled="busy" @click="startEdit">
              <Edit3 :size="14" :stroke-width="1.8" aria-hidden="true" />Edit
            </button>
          </template>
        </WorkspaceHeader>
      </WorkspaceStickyStack>

      <AppPanel
        v-if="editorMode"
        class="service-editor-panel"
        :title="editorMode === 'create' ? 'New service' : 'Edit service'"
        subtitle="The full definition saves atomically with its parameters."
      >
        <form id="service-editor" @submit.prevent="saveService" class="service-editor min-w-0 space-y-5">
          <section class="rounded-box border border-base-300 bg-base-200/25 p-4">
            <div class="mb-4">
              <h3 class="text-sm font-semibold">Service identity</h3>
              <p class="mt-1 text-xs text-base-content/60">Give this reusable operation a clear name and catalog reference.</p>
            </div>
            <div class="space-y-3">
          <FormField class="gap-1"
            ><span>Name</span
            ><AppInput
              class="input w-full min-w-0"
              v-model="form.name"
              type="text"
              placeholder="Digital Print"
              required
              autocomplete="off"
          /></FormField>
          <FormGrid
            ><FormField class="gap-1"
              ><span>Code</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.code"
                type="text"
                placeholder="PRINT"
                autocomplete="off" /></FormField
            ><FormField class="gap-1"
              ><span>Category</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.category"
                type="text"
                placeholder="Production"
                autocomplete="off" /></FormField
          ></FormGrid>
          <FormField class="gap-1"
            ><span>Description / notes</span
            ><AppTextarea
              v-model="form.description"
              rows="2"
              placeholder="What this operation covers"
            />
          </FormField>
            </div>
          </section>
          <section class="min-w-0 space-y-3">
            <header class="flex min-w-0 flex-wrap items-start justify-between gap-3">
              <div class="min-w-0">
                <h3 class="text-sm font-semibold">Parameters</h3>
                <p class="mt-1 text-xs text-base-content/60">Add the questions an operator must answer when using this service.</p>
              </div>
              <SelectField
                v-model="newInputType"
                aria-label="Add an operator input"
                :options="[
                  { label: 'Add an operator input…', value: '' },
                  { label: 'Quantity or count', value: 'integer' },
                  { label: 'Decimal measurement', value: 'decimal' },
                  { label: 'Choose from options', value: 'choice' },
                  { label: 'Choose a material or paper', value: 'material-reference' },
                  { label: 'Yes / no choice', value: 'boolean' },
                ]"
                @update:model-value="addInput"
              />
            </header>
          <div v-if="form.parameters.length" class="overflow-hidden rounded-box border border-base-300 bg-base-100">
            <ServiceParameterEditor v-for="(parameter,index) in form.parameters" :key="parameter.id" :parameter="parameter" :index="index" :count="form.parameters.length" :materials="materials" :show-errors="validationAttempted" @move="moveParameter(index,$event)" @remove="removeParameter(index)" @normalize="normalizeParameter(parameter)" @label-change="syncParameterKey(parameter)" @add-option="addOption(parameter)" @remove-option="removeOption(parameter,$event)" />
          </div>
          <EmptyState v-else compact title="No parameters yet" description="Add one when the service needs operator input.">
            <template #icon><ListPlus :size="20" aria-hidden="true" /></template>
            <template #action><SelectField
              v-model="newInputType"
              aria-label="Add an operator input"
              :options="[
                { label: 'Add an operator input…', value: '' },
                { label: 'Quantity or count', value: 'integer' },
                { label: 'Decimal measurement', value: 'decimal' },
                { label: 'Choose from options', value: 'choice' },
                { label: 'Choose a material or paper', value: 'material-reference' },
                { label: 'Yes / no choice', value: 'boolean' },
              ]"
              @update:model-value="addInput"
            /></template>
          </EmptyState>
          </section>

          <section class="min-w-0 space-y-3">
            <header class="flex min-w-0 flex-wrap items-start justify-between gap-3">
              <div class="min-w-0">
                <h3 class="text-sm font-semibold">Cost components</h3>
                <p class="mt-1 text-xs text-base-content/60">Add the things that make this service cost money, such as paper, a machine, or labor.</p>
              </div>
              <SelectField
                v-model="newCostType"
                aria-label="Add a cost"
                :options="[
                  { label: 'Add a cost…', value: '' },
                  { label: 'Material or paper', value: 'material' },
                  { label: 'Machine', value: 'machine' },
                  { label: 'Labor', value: 'labor' },
                  { label: 'Outsourced work', value: 'outsourced' },
                  { label: 'Fixed cost', value: 'fixed' },
                  { label: 'Overhead percentage', value: 'overhead' },
                  { label: 'Waste percentage', value: 'waste' },
                  { label: 'Manual cost', value: 'manual' },
                ]"
                @update:model-value="addCost"
              />
            </header>
          <div v-if="form.components.length" class="overflow-hidden rounded-box border border-base-300 bg-base-100">
            <ServiceCostEditor v-for="(component,index) in form.components" :key="component.id" :component="component" :index="index" :count="form.components.length" :materials="materials" :machines="machines" :parameters="form.parameters" :currency-unit="currencyUnit" :show-errors="validationAttempted" @move="moveComponent(index,$event)" @remove="removeComponent(index)" @change-type="updateComponentType(component)" @change-rate="updateComponentRate(component)" />
          </div>
          <EmptyState v-else compact title="No cost components yet" description="Add reusable material, machine, labor, or other inputs.">
            <template #icon><Plus :size="20" aria-hidden="true" /></template>
            <template #action><SelectField
              v-model="newCostType"
              aria-label="Add a cost"
              :options="[
                { label: 'Add a cost…', value: '' },
                { label: 'Material or paper', value: 'material' },
                { label: 'Machine', value: 'machine' },
                { label: 'Labor', value: 'labor' },
                { label: 'Outsourced work', value: 'outsourced' },
                { label: 'Fixed cost', value: 'fixed' },
                { label: 'Overhead percentage', value: 'overhead' },
                { label: 'Waste percentage', value: 'waste' },
                { label: 'Manual cost', value: 'manual' },
              ]"
              @update:model-value="addCost"
            /></template>
          </EmptyState>
          </section>

          <section v-if="form.pricingRule" class="min-w-0 space-y-3 rounded-box border border-base-300 bg-base-200/25 p-4">
            <div>
              <h3 class="text-sm font-semibold">Selling-price rule</h3>
              <p class="mt-1 text-xs text-base-content/60">Choose a generic suggestion rule; operators can override it in the live configurator.</p>
            </div>
          <div class="min-w-0 space-y-3">
            <SelectField
              v-model="form.pricingRule.type"
              label="Rule"
              :options="[
                { label: 'Manual / enter at order time', value: 'manual' },
                { label: 'Fixed selling price', value: 'fixed' },
                { label: 'Cost plus markup %', value: 'markup' },
                { label: 'Cost plus fixed margin', value: 'fixed-margin' },
                { label: 'Per-unit parameter', value: 'per-unit' },
                { label: 'Quantity tiers', value: 'quantity-tiers' },
              ]"
              @update:model-value="normalizePricingRule"
            />
            <FormField class="gap-1" v-if="form.pricingRule.type === 'fixed'"
              ><span>Fixed price ({{ props.currencyUnit }})</span
              ><AppInput
                :model-value="form.pricingRule.fixedPriceInput"
                :money="props.currencyUnit"
                inputmode="decimal"
                @update:model-value="updateGroupedMoney(form.pricingRule, 'fixedPriceInput', 'fixedPriceRial', $event)"
            /></FormField>
            <FormField class="gap-1" v-if="form.pricingRule.type === 'markup'"
              ><span>Markup percentage</span
              ><AppInput
                class="input w-full min-w-0"
                v-model="form.pricingRule.markupPercentage"
                inputmode="decimal"
                placeholder="20"
            /></FormField>
            <FormField class="gap-1" v-if="form.pricingRule.type === 'fixed-margin'"
              ><span>Fixed margin ({{ props.currencyUnit }})</span
              ><AppInput
                :model-value="form.pricingRule.fixedMarginInput"
                :money="props.currencyUnit"
                inputmode="decimal"
                @update:model-value="updateGroupedMoney(form.pricingRule, 'fixedMarginInput', 'fixedMarginRial', $event)"
            /></FormField>
            <template v-if="form.pricingRule.type === 'per-unit'"
              ><SelectField
                v-model="form.pricingRule.parameterKey"
                label="Numeric parameter"
                :options="[
                  { label: 'Select parameter', value: '' },
                  ...numericParameters().map((parameter) => ({
                    label: parameter.label || parameter.key,
                    value: parameter.key,
                  })),
                ]" /><FormField class="gap-1"
                ><span>Rate / unit ({{ props.currencyUnit }})</span
                ><AppInput
                  :model-value="form.pricingRule.perUnitRateInput"
                  :money="props.currencyUnit"
                  inputmode="decimal"
                  @update:model-value="updateGroupedMoney(form.pricingRule, 'perUnitRateInput', 'perUnitRateRial', $event)" /></FormField
            ></template>
            <div v-if="form.pricingRule.type === 'quantity-tiers'" class="min-w-0 space-y-3">
              <div class="min-w-0 space-y-3">
                <span>Quantity tiers</span
                ><button class="btn btn-ghost" type="button" @click="addPricingTier">
                  <Plus :size="14" :stroke-width="1.8" />Add tier
                </button>
              </div>
              <div
                v-for="(tier, tierIndex) in form.pricingRule.tiers"
                :key="tier.position"
                class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
              >
                <AppInput
                  class="input w-full min-w-0"
                  v-model="tier.minimumQuantity"
                  inputmode="decimal"
                  :placeholder="tierIndex === 0 ? '0' : '10'"
                  :aria-label="`Tier ${tierIndex + 1} minimum quantity`"
                /><AppInput
                  :model-value="tier.priceInput"
                  :money="props.currencyUnit"
                  inputmode="decimal"
                  placeholder="Price"
                  :aria-label="`Tier ${tierIndex + 1} price`"
                  @update:model-value="updateGroupedMoney(tier, 'priceInput', 'priceRial', $event)"
                /><button
                  class="btn btn-outline btn-error"
                  type="button"
                  :aria-label="`Remove tier ${tierIndex + 1}`"
                  @click="removePricingTier(tierIndex)"
                >
                  <Trash2 :size="13" :stroke-width="1.8" />
                </button>
              </div>
              <p v-if="!form.pricingRule.tiers.length">Add a zero-minimum tier before saving.</p>
            </div>
            <p v-if="form.pricingRule.type === 'manual'">
              The live configurator will show cost and require an entered selling price.
            </p>
          </div>
          </section>
        </form>
      </AppPanel>

      <InspectorShell
        v-else-if="selectedService"
        title="Service overview"
        subtitle="Definition used by orders and production."
      >
        <div class="grid min-w-0 gap-3 sm:grid-cols-3">
          <div class="rounded-box border border-base-300 bg-base-200/45 p-3">
            <span class="block text-xs text-base-content/60">Operator inputs</span>
            <strong class="mt-1 block text-xl leading-6 tabular-nums">{{ selectedService.parameters.length }}</strong>
            <span class="mt-1 block text-xs text-base-content/55">{{ selectedService.parameters.filter((parameter) => parameter.required).length }} required</span>
          </div>
          <div class="rounded-box border border-base-300 bg-base-200/45 p-3">
            <span class="block text-xs text-base-content/60">Cost components</span>
            <strong class="mt-1 block text-xl leading-6 tabular-nums">{{ selectedService.components.length }}</strong>
            <span class="mt-1 block text-xs text-base-content/55">Ordered pricing inputs</span>
          </div>
          <div class="rounded-box border border-base-300 bg-base-200/45 p-3">
            <span class="block text-xs text-base-content/60">Price rule</span>
            <strong class="mt-1 block truncate text-sm font-semibold">{{ selectedService.pricingRule?.type || 'Manual' }}</strong>
            <span class="mt-1 block text-xs text-base-content/55">Used in order pricing</span>
          </div>
        </div>

        <div
          v-if="selectedService.description"
          class="rounded-box border border-primary/25 bg-primary/5 p-3"
        >
          <span class="block text-xs font-semibold text-primary">Service notes</span>
          <p class="mt-1 text-sm leading-5">{{ selectedService.description }}</p>
        </div>

        <div class="grid min-w-0 gap-4 xl:grid-cols-2">
          <InspectorSection title="Operator parameters" description="Inputs available when this service is used in an order.">
            <div v-if="selectedService.parameters.length" class="min-w-0 space-y-2">
              <div
                v-for="parameter in selectedService.parameters"
                :key="parameter.id"
                class="flex min-w-0 items-start gap-3 rounded-box border border-base-300 bg-base-200/30 p-3"
              >
                <span class="grid size-7 shrink-0 place-items-center rounded-full bg-primary/15 text-xs font-semibold text-primary">
                  {{ String(parameter.position + 1).padStart(2, '0') }}
                </span>
                <div class="min-w-0 flex-1">
                  <div class="flex min-w-0 flex-wrap items-center gap-2">
                    <strong class="wrap-anywhere">{{ parameter.label || parameter.key }}</strong>
                    <span v-if="parameter.required" class="text-[0.6875rem] font-semibold text-primary">Required</span>
                  </div>
                  <small class="mt-1 block break-words text-xs leading-5 text-base-content/60">
                    <code>{{ parameter.key }}</code> · {{ typeLabel(parameter.type) }}<span v-if="parameter.unit"> · {{ parameter.unit }}</span>
                  </small>
                  <small v-if="parameter.type === 'choice'" class="mt-1 block break-words text-xs leading-5 text-base-content/60">
                    {{ parameter.options.join(' / ') || 'No choices configured' }}
                  </small>
                  <small v-if="parameter.defaultValue" class="mt-1 block break-words text-xs leading-5 text-base-content/60">
                    Default: {{ parameter.type === 'material-reference' ? materials.find((material) => material.id === parameter.defaultValue)?.name || parameter.defaultValue : parameter.defaultValue }}
                  </small>
                </div>
              </div>
            </div>
            <EmptyState v-else compact title="No parameters configured" description="Add parameters in the service editor when operators need input."><template #icon><ListPlus :size="21" aria-hidden="true" /></template></EmptyState>
          </InspectorSection>

          <InspectorSection title="Cost components" description="Ordered inputs used by the pricing engine.">
            <div v-if="selectedService.components.length" class="min-w-0 space-y-2">
              <div
                v-for="component in selectedService.components"
                :key="component.id"
                class="flex min-w-0 items-start gap-3 rounded-box border border-base-300 bg-base-200/30 p-3"
              >
                <span class="grid size-7 shrink-0 place-items-center rounded-full bg-base-300 text-xs font-semibold tabular-nums">
                  {{ String(component.position + 1).padStart(2, '0') }}
                </span>
                <div class="min-w-0 flex-1">
                  <strong class="block wrap-anywhere">{{ component.name || typeLabel(component.type) }}</strong>
                  <small class="mt-1 block break-words text-xs leading-5 text-base-content/60">{{ componentSummary(component) }}</small>
                </div>
              </div>
            </div>
            <EmptyState v-else compact title="No cost components configured" description="Add reusable inputs in the service editor."><template #icon><Plus :size="21" aria-hidden="true" /></template></EmptyState>
          </InspectorSection>
        </div>

        <div class="border-t border-base-300 pt-3 text-xs text-base-content/55">Updated {{ dateLabel(selectedService.updatedAt) }}</div>
      </InspectorShell>

      <div v-if="selectedService && !editorMode && selectedService.active" class="min-w-0 rounded-box border border-base-300 bg-base-100 p-4">
        <ServiceConfigurator
          :service="selectedService"
          :materials="materials"
          :currency-unit="props.currencyUnit"
        />
      </div>
  </div>
</template>
