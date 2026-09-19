<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowLeft, CheckCheck, CircleHelp, Save, Upload, X } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import ServiceMaterialsStep from './ServiceMaterialsStep.vue'
import ServiceLayoutSetup from './ServiceLayoutSetup.vue'
import ServiceOrderPreview from './ServiceOrderPreview.vue'
import ServiceCostBreakdownPreview from './ServiceCostBreakdownPreview.vue'
import ServicePricingStep from './ServicePricingStep.vue'
import ServicePricingPreview from './ServicePricingPreview.vue'
import ServiceTestStep from './ServiceTestStep.vue'
import ServiceTestPreview from './ServiceTestPreview.vue'
import ServiceOverviewIdentity from './ServiceOverviewIdentity.vue'
import ServiceOverviewSection from './ServiceOverviewSection.vue'
import type { MaterialAttributeDefinitionRecord, MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import { formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency'
import ServiceMachinesStep from './ServiceMachinesStep.vue'
import type { PredefinedParameter, ServiceForm } from './types'
import { outsourcedDefaultCostRial, outsourcedDefaultLines, type TestPricingResult, type TestValues } from './serviceTestPricing'
import { ensureSuggestedCostComponents, reconcileCostComponents } from './serviceComponentSync'
import { SERVICE_CATEGORIES, serviceCategoryRequirements, serviceCategorySupportsLayout } from './serviceCategory'
import { ensureRollSizeInputs } from './rollSizeInputs'

const props = defineProps<{
  form: ServiceForm
  editorMode: 'create' | 'edit'
  busy: boolean
  isSaving: boolean
  validationAttempted: boolean
  active: boolean
  materials: MaterialRecord[]
  attributeDefinitions: MaterialAttributeDefinitionRecord[]
  machines: MachineRecord[]
  services: ServiceRecord[]
  serviceId?: string
  currencyUnit: CurrencyUnit
  predefinedParameters: PredefinedParameter[]
}>()
const emit = defineEmits<{
  cancel: []
  save: []
}>()

const activeStep = ref(1)
watch(() => props.form, form => ensureRollSizeInputs(form), { immediate: true, deep: true })
const imageInput = ref<HTMLInputElement | null>(null)
const codeWasEdited = ref(Boolean(props.form.code.trim()))
const lastGeneratedCode = ref('')
const pricingCostEstimate = ref(0)
const pricingBreakdown = ref<Array<{ name: string; amount: number; detail: string; missing: boolean }>>([])
const testValues = ref<TestValues>({})
const testResult = ref<TestPricingResult | null>(null)
const imagePreview = computed(() => props.form.imagePath || '')
const categoryRequirements = computed(() => serviceCategoryRequirements(props.form.category))
type ServiceStepKey = 'basic' | 'materials' | 'machines' | 'outsourcing' | 'pricing' | 'test'
const isOutsourced = computed(() => props.form.fulfillmentMode === 'outsourced')
const defaultOutsourcedCostInvalid = computed(() => isOutsourced.value && props.form.defaultOutsourcedCostRial <= 0)
const defaultOutsourcedShippingInvalid = computed(() => props.form.defaultOutsourcedShippingRial < 0)
const effectivePricingCostEstimate = computed(() => isOutsourced.value ? outsourcedDefaultCostRial(props.form) : pricingCostEstimate.value)
const effectivePricingBreakdown = computed(() => isOutsourced.value ? outsourcedDefaultLines(props.form) : pricingBreakdown.value)
const steps = computed(() => isOutsourced.value
  ? [
      { number: 1, key: 'basic' as ServiceStepKey, title: 'Basic', description: 'Name, category, description' },
      { number: 2, key: 'outsourcing' as ServiceStepKey, title: 'Outsourcing', description: 'External fulfillment and order costs' },
      { number: 3, key: 'pricing' as ServiceStepKey, title: 'Pricing', description: 'How the price is calculated' },
      { number: 4, key: 'test' as ServiceStepKey, title: 'Test', description: 'Try it with real values' },
    ]
  : [
      { number: 1, key: 'basic' as ServiceStepKey, title: 'Basic', description: 'Name, category, description' },
      { number: 2, key: 'materials' as ServiceStepKey, title: 'Materials', description: 'Group stock choices by order options' },
      { number: 3, key: 'machines' as ServiceStepKey, title: 'Machines', description: 'Group machine rates and costs' },
      { number: 4, key: 'pricing' as ServiceStepKey, title: 'Pricing', description: 'How the price is calculated' },
      { number: 5, key: 'test' as ServiceStepKey, title: 'Test', description: 'Try it with real values' },
    ])
const activeStepKey = computed<ServiceStepKey>(() => steps.value[activeStep.value - 1]?.key || 'basic')
watch(isOutsourced, () => {
  if (activeStep.value > steps.value.length) activeStep.value = steps.value.length
})

function submit() {
  if (props.busy || props.isSaving) return
  if (activeStep.value < steps.value.length) next()
  else emit('save')
}

function goToStep(number: number) {
  reconcileCostComponents(props.form.components, props.form.parameters)
  const step = steps.value[number - 1]
  if (step?.key === 'materials' || step?.key === 'machines') ensureSuggestedCostComponents(props.form.components, props.form.parameters)
  activeStep.value = number
}
function requestSave() {
  reconcileCostComponents(props.form.components, props.form.parameters)
  ensureSuggestedCostComponents(props.form.components, props.form.parameters)
  emit('save')
}

function next() {
  if (props.busy || props.isSaving) return
  if (activeStep.value === 1) {
    const formElement = document.querySelector<HTMLFormElement>('#service-editor')
    if (formElement && !formElement.reportValidity()) return
  }
  if (activeStep.value < steps.value.length) goToStep(activeStep.value + 1)
}
function previous() {
  if (activeStep.value > 1) goToStep(activeStep.value - 1)
}
function browseImage() {
  imageInput.value?.click()
}
function onImageSelected(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !file.type.startsWith('image/')) return
  const reader = new FileReader()
  reader.onload = () => {
    if (typeof reader.result === 'string') props.form.imagePath = reader.result
  }
  reader.readAsDataURL(file)
  if (imageInput.value) imageInput.value.value = ''
}
function clearImage() {
  props.form.imagePath = ''
}
function updateOutsourcedDefaultMoney(text: string, field: 'cost' | 'shipping') {
  const inputKey = field === 'cost' ? 'defaultOutsourcedCostInput' : 'defaultOutsourcedShippingInput'
  const valueKey = field === 'cost' ? 'defaultOutsourcedCostRial' : 'defaultOutsourcedShippingRial'
  props.form[inputKey] = text
  const parsed = parseMoneyInput(text, props.currencyUnit)
  if (parsed !== null && parsed >= 0) {
    props.form[valueKey] = parsed
    props.form[inputKey] = formatMoneyInput(parsed, props.currencyUnit)
  }
}
function categoryOptions() {
  const defaults: string[] = [...SERVICE_CATEGORIES]
  const current = props.form.category.trim()
  return [
    { label: 'Select a category', value: '' },
    ...defaults.map((value) => ({ label: value, value })),
    ...(current && !defaults.includes(current) ? [{ label: current, value: current }] : []),
  ]
}
function stepClass(number: number) {
  if (number === activeStep.value) return 'wizard-step-active'
  if (number < activeStep.value) return 'wizard-step-complete'
  return 'wizard-step-idle'
}
function stepDescription(number: number, fallback: string) {
  const key = steps.value[number - 1]?.key
  if (key === 'materials') {
    const materialGroups = props.form.parameters.filter((parameter) => parameter.type === 'choice' && parameter.materialSource).length
    return `${materialGroups} material group${materialGroups === 1 ? '' : 's'} · ${props.form.parameters.length} order input${props.form.parameters.length === 1 ? '' : 's'}`
  }
  if (key === 'machines') {
    const machineGroups = props.form.components.filter((component) => component.type === 'machine').length
    return `${machineGroups} machine group${machineGroups === 1 ? '' : 's'} · defaults drive estimate`
  }
  if (key === 'outsourcing') return 'Set default supplier cost and shipping'
  if (key === 'pricing') return props.form.pricingRule?.type === 'manual' ? 'Manual price setup' : 'Selling price rules'
  if (key === 'test') return testResult.value ? 'Test result available' : 'Try it with real values'
  return fallback
}
function generateServiceCode(name: string) {
  const words = name
    .normalize('NFKC')
    .trim()
    .split(/[^\p{L}\p{N}]+/u)
    .filter(Boolean)
  if (!words.length) return ''
  const value = words.length === 1 ? words[0].slice(0, 4) : words.slice(0, 4).map((word) => word[0]).join('')
  return value.toLocaleUpperCase()
}
function markCodeEdited() {
  codeWasEdited.value = true
}
watch(
  () => props.form.name,
  (name) => {
    if (codeWasEdited.value) return
    const generated = generateServiceCode(name)
    if (!generated || !props.form.code || props.form.code === lastGeneratedCode.value) {
      props.form.code = generated
      lastGeneratedCode.value = generated
    }
  },
  { immediate: true },
)
watch(
  () => props.form.parameters,
  () => {
    if (activeStepKey.value !== 'materials' && activeStepKey.value !== 'machines') return
    const hasCostParameter = props.form.parameters.some((parameter) => parameter.type === 'material-reference' || parameter.type === 'machine-reference' || (parameter.type === 'choice' && Boolean(parameter.materialSource)))
    if (hasCostParameter) ensureSuggestedCostComponents(props.form.components, props.form.parameters)
  },
  { deep: true },
)
</script>

<template>
  <div class="service-wizard w-full flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" :aria-label='$t("Service editor")'>
    <header class="service-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">{{ $ui(editorMode === 'create' ? 'Add service' : 'Edit service') }}</h1>
        <p class="mt-1 text-xs leading-4 text-base-content/65">{{ $t("Define a printing service that can be used in orders.") }}</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Services' }, { label: editorMode === 'create' ? 'Add service' : 'Edit service', current: true }]" @navigate="emit('cancel')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="busy || isSaving" @click="emit('cancel')">{{ $t("Cancel") }}</button>
        <button v-if="editorMode === 'create'" class="btn btn-primary gap-2" type="button" :disabled="busy || isSaving" @click="requestSave"><Save :size="15" aria-hidden="true" />{{ $t("Save as draft") }}</button>
        <button class="btn btn-success gap-2" type="button" :disabled="busy || isSaving" @click="requestSave"><Save :size="16" aria-hidden="true" />{{ $ui(isSaving ? 'Saving…' : 'Save') }}</button>
      </div>
    </header>

    <div class="service-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="service-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200">
        <nav :aria-label='$t("Service setup steps")' class="service-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible">
          <button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" @click="goToStep(step.number)">
            <span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span>
            <span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ $ui(step.title) }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ $ui(stepDescription(step.number, step.description)) }}</small></span>
          </button>
        </nav>
      </aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="service-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
        <form id="service-editor" class="service-editor min-w-0" :aria-busy="busy || isSaving" @submit.prevent="submit">
          <section v-if="activeStepKey === 'basic'" class="min-w-0 space-y-6">
            <div class="grid min-w-0 gap-4 sm:grid-cols-2">
              <FormField class="gap-1 sm:col-span-2"><span>{{ $t("Service name") }} <em class="text-error">*</em></span><AppInput v-model="form.name" class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && !form.name.trim() }" type="text" required :placeholder='$t("Business card printing")' autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">{{ $t("A clear name shown to your team and customers.") }}</small></FormField>
              <FormField class="gap-1"><span>{{ $t("Code") }} <em class="text-error">*</em></span><AppInput v-model="form.code" class="input w-full min-w-0" type="text" required :placeholder='$t("Auto-generated")' autocomplete="off" @update:model-value="markCodeEdited" /><small class="text-xs leading-5 text-base-content/60">{{ $t("Generated from the service name. Edit it to use a custom code.") }}</small></FormField>
              <FormField class="gap-1"><span>{{ $t("Category") }} <em class="text-error">*</em></span><SelectField v-model="form.category" :options="categoryOptions()" :aria-label='$t("Service category")' /><small class="text-xs leading-5 text-base-content/60">{{ $t("Helps organize services in lists and reports.") }}</small></FormField>
              <FormField class="gap-1 sm:col-span-2"><span>{{ $t("Fulfillment") }} <em class="text-error">*</em></span><SelectField v-model="form.fulfillmentMode" :aria-label='$t("Service fulfillment")' :options="[
                { label: 'In-house production', value: 'in-house' },
                { label: 'Outsourced service', value: 'outsourced' },
              ]" /><small class="text-xs leading-5 text-base-content/60">{{ $ui(form.fulfillmentMode === 'outsourced' ? 'An external supplier fulfills this service. Set its default supplier cost in the Outsourcing step; order lines can override it when a quote changes.' : 'Your team fulfills this service using the configured materials, machines, and production flow.') }}</small></FormField>
              <FormField class="gap-1"><span>{{ $t("Default unit") }} <em class="text-error">*</em></span><SelectField v-model="form.defaultUnit" :aria-label='$t("Default service unit")' :options="[
                { label: 'Piece', value: 'piece' },
                { label: 'Sheet', value: 'sheet' },
                { label: 'Set', value: 'set' },
                { label: 'Hour', value: 'hour' },
                { label: 'Minute', value: 'minute' },
              ]" /><small class="text-xs leading-5 text-base-content/60">{{ $t("Used when an order does not specify a unit.") }}</small></FormField>
              <FormField class="gap-1"><span>{{ $t("Default priority") }}</span><SelectField v-model="form.defaultPriority" :aria-label='$t("Default service priority")' :options="[
                { label: 'Urgent', value: 'Urgent' },
                { label: 'High', value: 'High' },
                { label: 'Normal', value: 'Normal' },
                { label: 'Low', value: 'Low' },
              ]" /><small class="text-xs leading-5 text-base-content/60">{{ $t("Used for new order suggestions.") }}</small></FormField>
              <FormField class="gap-1 sm:col-span-2"><span>{{ $t("Service image") }} <em>{{ $t("optional") }}</em></span><input ref="imageInput" class="hidden" type="file" accept="image/png,image/jpeg,image/webp,image/gif" @change="onImageSelected" /><div class="flex flex-wrap items-center gap-3 rounded-box border border-dashed border-base-300 bg-base-100 p-3"><div v-if="imagePreview" class="relative size-20 shrink-0 overflow-hidden rounded-box border border-base-300 bg-base-200"><img :src="imagePreview" :alt='$t("Service preview")' class="size-full object-cover" /><button class="btn btn-circle btn-error btn-xs absolute right-1 top-1" type="button" :aria-label='$t("Remove service image")' @click.stop.prevent="clearImage"><X :size="12" aria-hidden="true" /></button></div><div class="min-w-0"><strong class="block text-sm">{{ $ui(imagePreview ? 'Service image selected' : 'Add a service image') }}</strong><small class="mt-1 block text-xs leading-5 text-base-content/60">{{ $t("Use a clear image to recognize this service in lists and orders.") }}</small></div><button class="btn btn-outline btn-sm ml-auto gap-2" type="button" @click.stop.prevent="browseImage"><Upload :size="14" aria-hidden="true" />{{ $ui(imagePreview ? 'Replace image' : 'Choose image') }}</button></div></FormField>
              <FormField class="gap-1 sm:col-span-2"><span>{{ $t("Description") }}</span><AppTextarea v-model="form.description" class="textarea w-full min-w-0" rows="6" maxlength="500" :placeholder='$t("High-quality printing with paper options and finishing.")' /><small class="text-xs leading-5 text-base-content/60">{{ $t("Add helpful details for your team and customers.") }}</small></FormField>
            </div>
          </section>

          <section v-else-if="activeStepKey === 'materials'">
          <ServiceMaterialsStep
            :category="form.category"
            :parameters="form.parameters"
            :materials="materials"
            :attribute-definitions="attributeDefinitions"
            :material-variants="form.materialVariants"
            :show-errors="validationAttempted"
            :currency-unit="currencyUnit"
            :required="categoryRequirements.material"
          />
          <ServiceLayoutSetup v-if="serviceCategorySupportsLayout(form.category)" :form="form" />
          </section>
          <section v-else-if="activeStepKey === 'machines'">
          <ServiceMachinesStep
            :components="form.components"
            :parameters="form.parameters"
            :machines="machines"
            :show-errors="validationAttempted"
            :required="categoryRequirements.machine"
          />
          </section>
          <section v-else-if="activeStepKey === 'outsourcing'" class="min-w-0 space-y-5">
            <div>
              <h2 class="text-xl font-semibold">{{ $t("Outsourced service setup") }}</h2>
              <p class="mt-1 max-w-2xl text-sm leading-6 text-base-content/65">{{ $t("This service is fulfilled outside the shop. Set the default supplier cost here; it is reused automatically when the service is added to an order.") }}</p>
            </div>
            <div class="rounded-box border border-info/30 bg-info/10 p-4 text-sm leading-6">
              <strong class="block">{{ $t("Service cost defaults") }}</strong>
              <span class="text-base-content/70">{{ $t("The default item cost is stored on the service and included in estimated cost. Order-time values remain optional overrides for supplier quotes.") }}</span>
            </div>
            <div class="grid min-w-0 gap-4 sm:grid-cols-2">
              <FormField class="gap-1">
                <span>{{ $t("Default outsourced item cost") }} <em class="text-error">*</em></span>
                <AppInput :model-value="form.defaultOutsourcedCostInput" :class="{ 'input-error': validationAttempted && defaultOutsourcedCostInvalid }" :money="currencyUnit" inputmode="decimal" :placeholder="$ui(`${$ui(currencyUnit)} cost`)" required @update:model-value="updateOutsourcedDefaultMoney($event, 'cost')" />
                <small class="text-xs leading-5 text-base-content/60">{{ $t("Used automatically for new order items unless an order-specific supplier quote is entered.") }}</small>
                <small v-if="validationAttempted && defaultOutsourcedCostInvalid" class="text-xs text-error">{{ $t("Enter the default outsourced cost for this service.") }}</small>
              </FormField>
              <FormField class="gap-1">
                <span>{{ $t("Default outsourced shipping") }} <em class="font-normal text-base-content/50">{{ $t("optional") }}</em></span>
                <AppInput :model-value="form.defaultOutsourcedShippingInput" :class="{ 'input-error': validationAttempted && defaultOutsourcedShippingInvalid }" :money="currencyUnit" inputmode="decimal" :placeholder="$ui(`Optional ${$ui(currencyUnit)} shipping`)" @update:model-value="updateOutsourcedDefaultMoney($event, 'shipping')" />
                <small class="text-xs leading-5 text-base-content/60">{{ $t("Applied to new order items when shipping is predictable.") }}</small>
              </FormField>
            </div>
          </section>
          <ServicePricingStep
            v-else-if="activeStepKey === 'pricing'"
            :pricing-rule="form.pricingRule"
            :parameters="form.parameters"
            :components="form.components"
            :materials="materials"
            :machines="machines"
            :material-variants="form.materialVariants"
            :currency-unit="currencyUnit"
            :estimated-cost-rial="effectivePricingCostEstimate"
            :show-errors="validationAttempted"
          />
          <ServiceTestStep
            v-else-if="activeStepKey === 'test'"
            :form="form"
            :parameters="form.parameters"
            :materials="materials"
            :machines="machines"
            :services="services"
            :values="testValues"
            :currency-unit="currencyUnit"
            @update:result="testResult = $event"
          />
          <section v-else class="flex min-h-[28rem] flex-col items-center justify-center text-center">
            <span class="grid size-14 place-items-center rounded-full bg-primary/15 text-primary"><CircleHelp :size="27" aria-hidden="true" /></span>
            <h2 class="mt-4 text-xl font-semibold">{{ $ui(steps[activeStep - 1].title) }} {{ $t("is coming soon") }}</h2>
            <p class="mt-2 max-w-md text-sm leading-6 text-base-content/65">{{ $t("This step is intentionally left for the next part of the service setup. Your existing service data is kept safe while we build it.") }}</p>
            <button class="btn btn-outline mt-5 gap-2" type="button" @click="previous"><ArrowLeft class="rtl-directional-arrow" :size="15" aria-hidden="true" />{{ $t("Back to basic information") }}</button>
          </section>
        </form>
        </section>

        <aside class="service-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto">
        <ServiceCostBreakdownPreview v-if="activeStepKey === 'materials' || activeStepKey === 'machines'" :form="form" :active="active" :components="form.components" :parameters="form.parameters" :materials="materials" :machines="machines" :services="services" :currency-unit="currencyUnit" @update:total="pricingCostEstimate = $event" @update:breakdown="pricingBreakdown = $event" />
        <ServiceOrderPreview v-if="activeStepKey === 'materials' || activeStepKey === 'machines'" :form="form" :materials="materials" :machines="machines" :currency-unit="currencyUnit" :show-identity="false" :show-material-estimate="false" :active="active" />
        <ServicePricingPreview v-if="activeStepKey === 'pricing'" :form="form" :active="active" :pricing-rule="form.pricingRule" :parameters="form.parameters" :materials="materials" :machines="machines" :estimated-cost-rial="effectivePricingCostEstimate" :breakdown="effectivePricingBreakdown" :currency-unit="currencyUnit" />
        <ServiceTestPreview v-if="activeStepKey === 'test'" :form="form" :active="active" :parameters="form.parameters" :values="testValues" :materials="materials" :machines="machines" :result="testResult" :currency-unit="currencyUnit" @edit="goToStep(isOutsourced ? 3 : 2)" />
        <template v-if="activeStepKey === 'basic'">
        <div class="space-y-4">
          <ServiceOverviewIdentity :form="form" :active="active" />
          <ServiceOverviewSection :title='$t("Configuration")' :description='$t("Current defaults for this service.")'>
            <dl class="divide-y divide-base-300/70">
              <div class="flex items-center justify-between gap-3 py-2.5 first:pt-0 text-sm"><dt class="text-base-content/60">{{ $t("Setup") }}</dt><dd>{{ $ui(form.parameters.length ? `${form.parameters.length} input${form.parameters.length === 1 ? '' : 's'}` : 'Not configured yet') }}</dd></div>
              <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">{{ $t("Costs") }}</dt><dd>{{ $ui(form.components.length ? `${form.components.length} component${form.components.length === 1 ? '' : 's'}` : 'Not configured yet') }}</dd></div>
              <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">{{ $t("Unit") }}</dt><dd>{{ $ui(form.defaultUnit) }}</dd></div>
              <div class="flex items-center justify-between gap-3 py-2.5 last:pb-0 text-sm"><dt class="text-base-content/60">{{ $t("Default priority") }}</dt><dd>{{ $ui(form.defaultPriority) }}</dd></div>
            </dl>
          </ServiceOverviewSection>
          <ServiceOverviewSection v-if="form.description" :title='$t("Description")'>
            <p class="text-sm leading-6 text-base-content/70">{{ form.description }}</p>
          </ServiceOverviewSection>
        </div>
        <div class="mt-4 flex gap-2 border-t border-base-300 pt-3 text-sm"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">{{ $t("Configure materials and order options that customers can choose when adding this service to an order.") }}</p></div>
        </template>
        </aside>
      </div>
    </div>
    <footer class="flex min-w-0 items-center justify-between gap-3 border-t border-base-300 px-1 pt-3"><button class="btn btn-ghost btn-sm" type="button" :disabled="activeStep === 1 || busy || isSaving" @click="previous">{{ $t("Back") }}</button><span class="text-xs text-base-content/55">{{ $t("Step") }} {{ activeStep }} {{ $t("of") }} {{ steps.length }}</span><button v-if="activeStep < steps.length" class="btn btn-primary btn-sm" type="button" @click="next">{{ $t("Continue") }}</button><button v-else class="btn btn-success btn-sm gap-2" type="button" :disabled="busy || isSaving" @click="requestSave"><Save :size="14" aria-hidden="true" />{{ $ui(isSaving ? 'Saving…' : 'Save service') }}</button></footer>
  </div>
</template>
