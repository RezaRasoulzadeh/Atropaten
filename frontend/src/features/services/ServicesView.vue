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
import MasterDetail from '../../components/layout/MasterDetail.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTableRow from '../../components/ui/DataTableRow.vue';
import DataTable from '../../components/ui/DataTable.vue';
import FormField from '../../components/ui/FormField.vue';
import AppInput from '../../components/ui/AppInput.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref, watch } from 'vue';
import {
  Archive,
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
</script>

<template>
  <div class="min-w-0 space-y-3">
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
        /></template><template #count><span class="self-end pb-2"
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

    <MasterDetail :wide="!!editorMode" aria-label="Services workspace">
      <AppPanel
        title="Service register"
        subtitle="Select a service to inspect its operator parameters."
      >
        <template #action
          ><span>{{ filteredServices.length }} shown</span></template
        >
        <LoadingState v-if="isLoading" label="Loading records…" />
        <div v-else-if="filteredServices.length">
          <DataTable
            ><thead>
              <tr>
                <th scope="col">Service</th>
                <th scope="col">Category</th>
                <th scope="col">Parameters</th>
                <th scope="col">Status</th>
                <th scope="col">Updated</th>
              </tr>
            </thead>
            <tbody>
              <DataTableRow
                v-for="service in filteredServices"
                :key="service.id"
                :class="{ 'bg-base-300': selectedId === service.id }"
                @activate="selectService(service.id)"
                interactive
                ><DataTableCell
                  ><span class="block font-medium">{{ service.name }}</span
                  ><span class="block text-xs text-base-content/60">{{
                    service.code || 'No code'
                  }}</span></DataTableCell
                ><DataTableCell>{{ service.category || '—' }}</DataTableCell
                ><DataTableCell
                  ><span class="block font-medium"
                    >{{ service.parameters.length }}
                    {{ service.parameters.length === 1 ? 'parameter' : 'parameters' }}</span
                  ><span class="block text-xs text-base-content/60"
                    >{{
                      service.parameters.filter((parameter) => parameter.required).length
                    }}
                    required</span
                  ></DataTableCell
                ><DataTableCell
                  ><StatusBadge
                    :label="service.active ? 'Active' : 'Archived'"
                    :tone="service.active ? 'green' : 'slate'" /></DataTableCell
                ><DataTableCell>{{ dateLabel(service.updatedAt) }}</DataTableCell></DataTableRow
              >
            </tbody>
          </DataTable>
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
      </AppPanel>

      <AppPanel
        v-if="editorMode"
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
        <form @submit.prevent="saveService" class="min-w-0 space-y-3">
          <div v-if="formError" role="alert">{{ formError }}</div>
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
            ><textarea
              class="textarea w-full min-w-0"
              v-model="form.description"
              rows="2"
              placeholder="What this operation covers"
            ></textarea>
          </FormField>
          <div class="min-w-0 space-y-3">
            <div class="min-w-0 space-y-3">
              <h3 class="text-sm font-semibold">Parameters</h3>
              <p>Order is saved as shown and keys are stable references.</p>
            </div>
            <button class="btn btn-ghost" type="button" @click="addParameter">
              <ListPlus :size="15" :stroke-width="1.8" aria-hidden="true" />Add parameter
            </button>
          </div>
          <div v-if="form.parameters.length">
            <ServiceParameterEditor v-for="(parameter,index) in form.parameters" :key="parameter.id" :parameter="parameter" :index="index" :count="form.parameters.length" :materials="materials" @move="moveParameter(index,$event)" @remove="removeParameter(index)" @normalize="normalizeParameter(parameter)" @add-option="addOption(parameter)" @remove-option="removeOption(parameter,$event)" />
          </div>
          <div v-else class="min-w-0 space-y-3">
            <ListPlus :size="17" :stroke-width="1.8" aria-hidden="true" /><span
              >No parameters yet. Add one when the service needs operator input.</span
            >
          </div>
          <div class="min-w-0 space-y-3">
            <div class="min-w-0 space-y-3">
              <h3 class="text-sm font-semibold">Cost components</h3>
              <p>Ordered reusable inputs evaluated by the generic pricing preview.</p>
            </div>
            <button class="btn btn-ghost" type="button" @click="addComponent">
              <Plus :size="15" :stroke-width="1.8" aria-hidden="true" />Add component
            </button>
          </div>
          <div v-if="form.components.length">
            <ServiceCostEditor v-for="(component,index) in form.components" :key="component.id" :component="component" :index="index" :count="form.components.length" :materials="materials" :machines="machines" :parameters="numericParameters()" :currency-unit="currencyUnit" @move="moveComponent(index,$event)" @remove="removeComponent(index)" @change-type="updateComponentType(component)" @change-rate="updateComponentRate(component)" />
          </div>
          <div v-else class="min-w-0 space-y-3">
            <Plus :size="17" :stroke-width="1.8" aria-hidden="true" /><span
              >No cost components yet. Add reusable material, machine, labor, or other inputs.</span
            >
          </div>
          <div>
            <div class="min-w-0 space-y-3">
              <h3 class="text-sm font-semibold">Selling-price rule</h3>
              <p>
                Choose a generic suggestion rule; operators can override it in the live
                configurator.
              </p>
            </div>
          </div>
          <div v-if="form.pricingRule" class="min-w-0 space-y-3">
            <SelectField
              v-model="form.pricingRule.type"
              label="Rule"
              :options="[
                { label: 'Manual / enter at quote time', value: 'manual' },
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
              ><input
                class="input w-full min-w-0"
                :value="form.pricingRule.fixedPriceInput"
                inputmode="decimal"
                @input="
                  updateGroupedMoney(form.pricingRule, 'fixedPriceInput', 'fixedPriceRial', $event)
                "
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
              ><input
                class="input w-full min-w-0"
                :value="form.pricingRule.fixedMarginInput"
                inputmode="decimal"
                @input="
                  updateGroupedMoney(
                    form.pricingRule,
                    'fixedMarginInput',
                    'fixedMarginRial',
                    $event,
                  )
                "
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
                ><input
                  class="input w-full min-w-0"
                  :value="form.pricingRule.perUnitRateInput"
                  inputmode="decimal"
                  @input="
                    updateGroupedMoney(
                      form.pricingRule,
                      'perUnitRateInput',
                      'perUnitRateRial',
                      $event,
                    )
                  " /></FormField
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
                /><input
                  class="input w-full min-w-0"
                  :value="tier.priceInput"
                  inputmode="decimal"
                  placeholder="Price"
                  :aria-label="`Tier ${tierIndex + 1} price`"
                  @input="updateGroupedMoney(tier, 'priceInput', 'priceRial', $event)"
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
          <div class="flex flex-wrap items-center gap-2">
            <button class="btn btn-ghost" type="button" @click="cancelEditor">Cancel</button
            ><button class="btn btn-primary" type="submit" :disabled="busy || (isSaving)">
              <Save :size="15" :stroke-width="1.8" aria-hidden="true" />{{
                isSaving ? 'Saving…' : 'Save service'
              }}
            </button>
          </div>
        </form>
      </AppPanel>

      <InspectorShell
        v-else-if="selectedService"
        title="Service inspector"
        subtitle="Current persisted definition"
      >
        <template #action
          ><button
            class="btn btn-ghost"
            type="button"
            aria-label="Edit selected service"
            @click="startEdit"
          >
            <Edit3 :size="15" :stroke-width="1.8" aria-hidden="true" /></button
        ></template>
        <div class="min-w-0 space-y-3">
          <StatusBadge
            :label="selectedService.active ? 'Active' : 'Archived'"
            :tone="selectedService.active ? 'green' : 'slate'"
          /><span>{{ selectedService.parameters.length }} parameters</span>
        </div>
        <div class="min-w-0 space-y-3">
          <div><SlidersHorizontal :size="19" :stroke-width="1.8" aria-hidden="true" /></div>
          <div class="min-w-0 space-y-3">
            <h3 class="text-sm font-semibold">{{ selectedService.name }}</h3>
            <p>
              {{ selectedService.code || 'No code'
              }}<span v-if="selectedService.category"> · {{ selectedService.category }}</span>
            </p>
          </div>
        </div>
        <p v-if="selectedService.description">{{ selectedService.description }}</p>
        <div class="min-w-0 space-y-3">
          <div
            v-for="parameter in selectedService.parameters"
            :key="parameter.id"
            class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
          >
            <span>{{ String(parameter.position + 1).padStart(2, '0') }}</span>
            <div class="min-w-0 space-y-1">
              <strong>{{ parameter.label }}</strong
              ><small class="block text-xs leading-5 text-base-content/60"
                ><code>{{ parameter.key }}</code> · {{ typeLabel(parameter.type)
                }}<span v-if="parameter.required"> · required</span></small
              ><small
                v-if="parameter.type === 'choice'"
                class="block text-xs leading-5 text-base-content/60"
                >{{ parameter.options.join(' / ') }}</small
              ><small
                v-if="parameter.type === 'material-reference'"
                class="block text-xs leading-5 text-base-content/60"
                >Active material selection</small
              ><small
                v-if="parameter.defaultValue"
                class="block text-xs leading-5 text-base-content/60"
                >Default:
                {{
                  parameter.type === 'material-reference'
                    ? materials.find((material) => material.id === parameter.defaultValue)?.name ||
                      parameter.defaultValue
                    : parameter.defaultValue
                }}</small
              >
            </div>
          </div>
          <p v-if="!selectedService.parameters.length">No parameters configured.</p>
        </div>
        <div class="min-w-0 space-y-3">
          <div>
            Cost components <span>{{ selectedService.components.length }}</span>
          </div>
          <div
            v-for="component in selectedService.components"
            :key="component.id"
            class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
          >
            <span>{{ String(component.position + 1).padStart(2, '0') }}</span>
            <div class="min-w-0 space-y-1">
              <strong>{{ component.name }}</strong
              ><small class="block text-xs leading-5 text-base-content/60">{{
                componentSummary(component)
              }}</small>
            </div>
          </div>
          <p v-if="!selectedService.components.length">No cost components configured.</p>
        </div>
        <div>Updated {{ dateLabel(selectedService.updatedAt) }}</div>
        <div class="flex flex-wrap items-center gap-2">
          <button
            class="btn btn-ghost"
            v-if="selectedService.active"
            type="button"
            @click="setActive(false)"
           :disabled="busy">
            <Archive :size="15" :stroke-width="1.8" aria-hidden="true" />Archive</button
          ><button class="btn btn-ghost" v-else type="button" @click="setActive(true)" :disabled="busy">
            <RotateCcw :size="15" :stroke-width="1.8" aria-hidden="true" />Reactivate
          </button>
        </div>
      </InspectorShell>

      <InspectorShell v-else title="Service inspector" subtitle="Select a row to inspect it."
        ><div class="min-w-0 space-y-3">
          <SlidersHorizontal :size="20" :stroke-width="1.8" aria-hidden="true" />
          <p>Service details will appear here.</p>
          <button class="btn btn-primary" type="button" @click="startCreate">
            Create a service <Plus :size="14" :stroke-width="1.8" aria-hidden="true" />
          </button></div
      ></InspectorShell>
    </MasterDetail>
    <ServiceConfigurator
      v-if="selectedService && !editorMode && selectedService.active"
      :service="selectedService"
      :materials="materials"
      :currency-unit="props.currencyUnit"
    />
  </div>
</template>
