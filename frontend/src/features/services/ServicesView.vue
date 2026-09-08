<script setup lang="ts">
import ServiceParameterEditor from './ServiceParameterEditor.vue'
import ServiceCostEditor from './ServiceCostEditor.vue'
import {componentNeedsRate,componentNeedsReference,componentNeedsPercentage} from './serviceFields'
import type {ServiceFilter,EditorMode,ParameterType,ParameterForm,ComponentType,ComponentForm,PricingTierForm,PricingRuleForm,ServiceForm} from './types'
import LoadingState from '../../components/ui/LoadingState.vue'
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
  X,
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
const {busy,runAction,services,materials,machines,selectedId,searchQuery,serviceFilter,editorMode,form,isLoading,isSaving,errorMessage,formError,selectedService,filteredServices,emptyForm,emptyParameter,emptyComponent,loadServices,selectService,startCreate,startEdit,numericParameters,normalizeComponent,updateComponentType,addComponent,removeComponent,moveComponent,updateComponentRate,updateGroupedMoney,normalizePricingRule,addPricingTier,removePricingTier,componentSummary,cancelEditor,addParameter,removeParameter,moveParameter,normalizeParameter,addOption,removeOption,saveService,setActive,typeLabel,dateLabel,errorMessageFrom}=useServicesWorkspace(props,emit)
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
  <div class="min-w-0 space-y-3">
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

      <div v-if="errorMessage" role="alert" class="min-w-0 space-y-3">
        <span>{{ errorMessage }}</span
        ><button
          class="btn btn-ghost"
          type="button"
          aria-label="Dismiss services error"
          @click="errorMessage = ''"
        >
          <X :size="15" :stroke-width="1.8" aria-hidden="true" />
        </button>
      </div>

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
        <div v-else class="min-w-0 space-y-3">
          <div aria-hidden="true"><SlidersHorizontal :size="21" :stroke-width="1.8" /></div>
          <h2 class="text-base font-semibold">
            {{ services.length ? 'No services match this view' : 'No services yet' }}
          </h2>
          <p>
            {{
              services.length
                ? 'Try another status or search term.'
                : 'Create the first reusable operation for your shop.'
            }}
          </p>
          <button
            class="btn btn-primary"
            v-if="!services.length"
            type="button"
            @click="startCreate"
          >
            <Plus :size="15" :stroke-width="1.8" aria-hidden="true" />Create service
          </button>
        </div>
      </RegisterList>
    </div>

    <div v-else class="min-w-0 space-y-4" aria-label="Service workspace">
      <WorkspaceStickyStack>
        <WorkspaceHeader
          :title="editorMode === 'create' ? 'New service' : selectedService?.name || 'Service'"
          eyebrow="Catalog / service workspace"
          :description="
            editorMode === 'create'
              ? 'Create a reusable service for orders and production.'
              : 'Review the service definition, test its inputs, and manage its lifecycle.'
          "
        >
          <template #leading>
            <button
              class="btn btn-ghost btn-sm"
              type="button"
              aria-label="Back to services"
              @click="backToServices"
            >
              <ArrowLeft :size="17" :stroke-width="1.8" aria-hidden="true" />
              <span class="hidden sm:inline">Services</span>
            </button>
          </template>
        </WorkspaceHeader>
      </WorkspaceStickyStack>

      <div v-if="errorMessage" role="alert" class="min-w-0 space-y-3">
        <span>{{ errorMessage }}</span
        ><button
          class="btn btn-ghost"
          type="button"
          aria-label="Dismiss services error"
          @click="errorMessage = ''"
        >
          <X :size="15" :stroke-width="1.8" aria-hidden="true" />
        </button>
      </div>

      <AppPanel
        v-if="editorMode"
        class="service-editor-panel"
        :title="editorMode === 'create' ? 'New service' : 'Edit service'"
        subtitle="The full definition saves atomically with its parameters."
      >
        <template #action
          ><button
            class="btn btn-ghost"
            type="button"
            aria-label="Close service editor"
            @click="cancelEditor"
          >
            <X :size="16" :stroke-width="1.8" aria-hidden="true" /></button
        ></template>
        <form @submit.prevent="saveService" class="min-w-0 space-y-5">
          <div v-if="formError" role="alert" class="rounded-box border border-error/35 bg-error/10 p-3 text-sm text-error">
            {{ formError }}
          </div>
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
                <p class="mt-1 text-xs text-base-content/60">Order is saved as shown and keys are stable references.</p>
              </div>
              <button class="btn btn-outline btn-sm shrink-0" type="button" @click="addParameter">
                <ListPlus :size="14" :stroke-width="1.8" aria-hidden="true" />Add parameter
              </button>
            </header>
          <div v-if="form.parameters.length" class="overflow-hidden rounded-box border border-base-300 bg-base-100">
            <ServiceParameterEditor v-for="(parameter,index) in form.parameters" :key="parameter.id" :parameter="parameter" :index="index" :count="form.parameters.length" :materials="materials" @move="moveParameter(index,$event)" @remove="removeParameter(index)" @normalize="normalizeParameter(parameter)" @add-option="addOption(parameter)" @remove-option="removeOption(parameter,$event)" />
          </div>
          <div v-else class="flex min-w-0 items-center gap-2 rounded-box border border-dashed border-base-300 p-4 text-sm text-base-content/60">
            <ListPlus :size="17" :stroke-width="1.8" aria-hidden="true" /><span
              >No parameters yet. Add one when the service needs operator input.</span
            >
          </div>
          </section>

          <section class="min-w-0 space-y-3">
            <header class="flex min-w-0 flex-wrap items-start justify-between gap-3">
              <div class="min-w-0">
                <h3 class="text-sm font-semibold">Cost components</h3>
                <p class="mt-1 text-xs text-base-content/60">Ordered reusable inputs evaluated by the generic pricing preview.</p>
              </div>
              <button class="btn btn-outline btn-sm shrink-0" type="button" @click="addComponent">
                <Plus :size="14" :stroke-width="1.8" aria-hidden="true" />Add component
              </button>
            </header>
          <div v-if="form.components.length" class="overflow-hidden rounded-box border border-base-300 bg-base-100">
            <ServiceCostEditor v-for="(component,index) in form.components" :key="component.id" :component="component" :index="index" :count="form.components.length" :materials="materials" :machines="machines" :parameters="numericParameters()" :currency-unit="currencyUnit" @move="moveComponent(index,$event)" @remove="removeComponent(index)" @change-type="updateComponentType(component)" @change-rate="updateComponentRate(component)" />
          </div>
          <div v-else class="flex min-w-0 items-center gap-2 rounded-box border border-dashed border-base-300 p-4 text-sm text-base-content/60">
            <Plus :size="17" :stroke-width="1.8" aria-hidden="true" /><span
              >No cost components yet. Add reusable material, machine, labor, or other inputs.</span
            >
          </div>
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
                  inputmode="decimal"
                  placeholder="Price"
                  :aria-label="`Tier ${tierIndex + 1} price`"
                  @update:model-value="updateGroupedMoney(tier, 'priceInput', 'priceRial', $event)"
                /><button
                  class="btn btn-ghost"
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
          <div class="flex flex-wrap items-center justify-end gap-2 border-t border-base-300 pt-4">
            <button class="btn btn-ghost" type="button" @click="cancelEditor">Cancel</button
            ><button class="btn btn-primary" type="submit" :disabled="busy || isSaving">
              <Save :size="15" :stroke-width="1.8" aria-hidden="true" />{{
                isSaving ? 'Saving…' : 'Save service'
              }}
            </button>
          </div>
        </form>
      </AppPanel>

      <InspectorShell
        v-else-if="selectedService"
        title="Service overview"
        subtitle="Definition used by orders and production."
      >
        <template #action>
          <StatusBadge
            :label="selectedService.active ? 'Active' : 'Archived'"
            :tone="selectedService.active ? 'green' : 'slate'"
          />
          <button class="btn btn-primary btn-sm" type="button" @click="startEdit">
            <Edit3 :size="14" :stroke-width="1.8" aria-hidden="true" />Edit
          </button>
        </template>

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
            <p v-else class="rounded-box border border-dashed border-base-300 p-3 text-sm text-base-content/60">No parameters configured.</p>
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
            <p v-else class="rounded-box border border-dashed border-base-300 p-3 text-sm text-base-content/60">No cost components configured.</p>
          </InspectorSection>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3 border-t border-base-300 pt-3">
          <span class="text-xs text-base-content/55">Updated {{ dateLabel(selectedService.updatedAt) }}</span>
          <button
            class="btn btn-ghost btn-sm"
            v-if="selectedService.active"
            type="button"
            @click="setActive(false)"
            :disabled="busy"
          >
            <Archive :size="14" :stroke-width="1.8" aria-hidden="true" />Deactivate service
          </button>
          <button class="btn btn-ghost btn-sm" v-else type="button" @click="setActive(true)" :disabled="busy">
            <RotateCcw :size="14" :stroke-width="1.8" aria-hidden="true" />Reactivate service
          </button>
        </div>
      </InspectorShell>

      <div v-if="selectedService && !editorMode && selectedService.active" class="min-w-0 rounded-box border border-base-300 bg-base-100 p-4">
        <ServiceConfigurator
          :service="selectedService"
          :materials="materials"
          :currency-unit="props.currencyUnit"
        />
      </div>
    </div>
  </div>
</template>
