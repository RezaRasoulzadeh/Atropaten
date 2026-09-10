<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowLeft, CheckCheck, CircleHelp, Save, Upload, X } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import ServiceParametersStep from './ServiceParametersStep.vue'
import ServiceOrderPreview from './ServiceOrderPreview.vue'
import ServiceCostBreakdownPreview from './ServiceCostBreakdownPreview.vue'
import ServicePricingStep from './ServicePricingStep.vue'
import ServicePricingPreview from './ServicePricingPreview.vue'
import ServiceTestStep from './ServiceTestStep.vue'
import ServiceTestPreview from './ServiceTestPreview.vue'
import ServiceOverviewIdentity from './ServiceOverviewIdentity.vue'
import ServiceOverviewSection from './ServiceOverviewSection.vue'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import type { CurrencyUnit } from '../../utils/currency'
import ServiceCostComponentsStep from './ServiceCostComponentsStep.vue'
import type { ParameterTemplateSeed, ServiceForm } from './types'
import type { TestPricingResult, TestValues } from './serviceTestPricing'

const props = defineProps<{
  form: ServiceForm
  editorMode: 'create' | 'edit'
  busy: boolean
  isSaving: boolean
  validationAttempted: boolean
  active: boolean
  materials: MaterialRecord[]
  machines: MachineRecord[]
  services: ServiceRecord[]
  serviceId?: string
  currencyUnit: CurrencyUnit
}>()
const emit = defineEmits<{
  cancel: []
  save: []
}>()

const activeStep = ref(1)
const imageInput = ref<HTMLInputElement | null>(null)
const codeWasEdited = ref(Boolean(props.form.code.trim()))
const lastGeneratedCode = ref('')
const pricingCostEstimate = ref(0)
const pricingBreakdown = ref<Array<{ name: string; amount: number; detail: string; missing: boolean }>>([])
const testValues = ref<TestValues>({})
const testResult = ref<TestPricingResult | null>(null)
const imagePreview = computed(() => props.form.imagePath || '')
const steps = [
  { number: 1, title: 'Basic', description: 'Name, category, description' },
  { number: 2, title: 'Parameters', description: 'Options customers can choose' },
  { number: 3, title: 'Cost components', description: 'Material, machine, labor, etc.' },
  { number: 4, title: 'Pricing', description: 'How the price is calculated' },
  { number: 5, title: 'Test', description: 'Try it with real values' },
]

function next() {
  if (activeStep.value === 1) {
    const formElement = document.querySelector<HTMLFormElement>('#service-editor')
    if (formElement && !formElement.reportValidity()) return
  }
  if (activeStep.value < steps.length) activeStep.value += 1
}
function previous() {
  if (activeStep.value > 1) activeStep.value -= 1
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
function categoryOptions() {
  const defaults = [
    'Print products',
    'Paper printing',
    'Large format',
    'Banners & signage',
    'Cards & stationery',
    'Flyers & brochures',
    'Labels & stickers',
    'Finishing',
    'Design',
    'Packaging',
  ]
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
function applyParameterTemplate(parameters: ParameterTemplateSeed[]) {
  props.form.parameters = parameters.map((parameter, index) => ({
    ...parameter,
    id: `draft-parameter-${Date.now()}-${index}-${Math.random().toString(16).slice(2)}`,
  }))
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
</script>

<template>
  <div class="service-wizard flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Service editor">
    <header class="service-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-semibold tracking-tight text-primary">{{ editorMode === 'create' ? 'Add service' : 'Edit service' }}</h1>
        <p class="mt-1 text-sm text-base-content/65">Define a printing service that can be used in orders.</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Services' }, { label: editorMode === 'create' ? 'Add service' : 'Edit service', current: true }]" @navigate="emit('cancel')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="busy || isSaving" @click="emit('cancel')">Cancel</button>
        <button v-if="editorMode === 'create'" class="btn btn-primary gap-2" type="button" :disabled="busy || isSaving" @click="emit('save')"><Save :size="15" aria-hidden="true" />Save as draft</button>
        <button class="btn btn-success gap-2" type="button" :disabled="busy || isSaving" @click="emit('save')"><Save :size="16" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save' }}</button>
      </div>
    </header>

    <div class="service-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="service-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200">
        <nav aria-label="Service setup steps" class="service-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible">
          <button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" @click="activeStep = step.number">
            <span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span>
            <span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ step.title }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ step.description }}</small></span>
          </button>
        </nav>
      </aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="service-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
        <form id="service-editor" class="service-editor min-w-0" @submit.prevent="next">
          <section v-if="activeStep === 1" class="min-w-0 space-y-6">
            <div class="grid min-w-0 gap-4 sm:grid-cols-2">
              <FormField class="gap-1 sm:col-span-2"><span>Service name <em class="text-error">*</em></span><AppInput v-model="form.name" class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && !form.name.trim() }" type="text" required placeholder="Business card printing" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">A clear name shown to your team and customers.</small></FormField>
              <FormField class="gap-1"><span>Code <em class="text-error">*</em></span><AppInput v-model="form.code" class="input w-full min-w-0" type="text" required placeholder="Auto-generated" autocomplete="off" @update:model-value="markCodeEdited" /><small class="text-xs leading-5 text-base-content/60">Generated from the service name. Edit it to use a custom code.</small></FormField>
              <FormField class="gap-1"><span>Category <em class="text-error">*</em></span><SelectField v-model="form.category" :options="categoryOptions()" aria-label="Service category" /><small class="text-xs leading-5 text-base-content/60">Helps organize services in lists and reports.</small></FormField>
              <FormField class="gap-1"><span>Default unit <em class="text-error">*</em></span><SelectField v-model="form.defaultUnit" aria-label="Default service unit" :options="[
                { label: 'Piece', value: 'piece' },
                { label: 'Sheet', value: 'sheet' },
                { label: 'Set', value: 'set' },
                { label: 'Hour', value: 'hour' },
                { label: 'Minute', value: 'minute' },
              ]" /><small class="text-xs leading-5 text-base-content/60">Used when an order does not specify a unit.</small></FormField>
              <FormField class="gap-1"><span>Default priority</span><SelectField v-model="form.defaultPriority" aria-label="Default service priority" :options="[
                { label: 'Urgent', value: 'Urgent' },
                { label: 'High', value: 'High' },
                { label: 'Normal', value: 'Normal' },
                { label: 'Low', value: 'Low' },
              ]" /><small class="text-xs leading-5 text-base-content/60">Used for new order suggestions.</small></FormField>
              <FormField class="gap-1 sm:col-span-2"><span>Service image <em>optional</em></span><input ref="imageInput" class="hidden" type="file" accept="image/png,image/jpeg,image/webp,image/gif" @change="onImageSelected" /><div class="flex flex-wrap items-center gap-3 rounded-box border border-dashed border-base-300 bg-base-100 p-3"><div v-if="imagePreview" class="relative size-20 shrink-0 overflow-hidden rounded-box border border-base-300 bg-base-200"><img :src="imagePreview" alt="Service preview" class="size-full object-cover" /><button class="btn btn-circle btn-error btn-xs absolute right-1 top-1" type="button" aria-label="Remove service image" @click.stop.prevent="clearImage"><X :size="12" aria-hidden="true" /></button></div><div class="min-w-0"><strong class="block text-sm">{{ imagePreview ? 'Service image selected' : 'Add a service image' }}</strong><small class="mt-1 block text-xs leading-5 text-base-content/60">Use a clear image to recognize this service in lists and orders.</small></div><button class="btn btn-outline btn-sm ml-auto gap-2" type="button" @click.stop.prevent="browseImage"><Upload :size="14" aria-hidden="true" />{{ imagePreview ? 'Replace image' : 'Choose image' }}</button></div></FormField>
              <FormField class="gap-1 sm:col-span-2"><span>Description</span><AppTextarea v-model="form.description" class="textarea w-full min-w-0" rows="6" maxlength="500" placeholder="High-quality printing with paper options and finishing." /><small class="text-xs leading-5 text-base-content/60">Add helpful details for your team and customers.</small></FormField>
            </div>
          </section>

          <ServiceParametersStep
            v-else-if="activeStep === 2"
            :parameters="form.parameters"
            :category="form.category"
            :default-unit="form.defaultUnit"
            :materials="materials"
            :machines="machines"
            :template-scope="form.code || form.name || 'new-service'"
            :show-errors="validationAttempted"
            @apply-template="applyParameterTemplate"
          />
          <ServiceCostComponentsStep
            v-else-if="activeStep === 3"
            :components="form.components"
            :parameters="form.parameters"
            :materials="materials"
            :machines="machines"
            :services="services"
            :current-service-id="serviceId"
            :currency-unit="currencyUnit"
            :show-errors="validationAttempted"
          />
          <ServicePricingStep
            v-else-if="activeStep === 4"
            :pricing-rule="form.pricingRule"
            :parameters="form.parameters"
            :currency-unit="currencyUnit"
            :estimated-cost-rial="pricingCostEstimate"
            :show-errors="validationAttempted"
          />
          <ServiceTestStep
            v-else-if="activeStep === 5"
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
            <h2 class="mt-4 text-xl font-semibold">{{ steps[activeStep - 1].title }} is coming soon</h2>
            <p class="mt-2 max-w-md text-sm leading-6 text-base-content/65">This step is intentionally left for the next part of the service setup. Your existing service data is kept safe while we build it.</p>
            <button class="btn btn-outline mt-5 gap-2" type="button" @click="previous"><ArrowLeft :size="15" aria-hidden="true" />Back to basic information</button>
          </section>
        </form>
        </section>

        <aside class="service-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto">
        <ServiceOrderPreview v-if="activeStep === 2" :form="form" :materials="materials" :machines="machines" :active="active" />
        <ServiceCostBreakdownPreview v-show="activeStep === 3" :form="form" :active="active" :components="form.components" :parameters="form.parameters" :materials="materials" :machines="machines" :services="services" :currency-unit="currencyUnit" @update:total="pricingCostEstimate = $event" @update:breakdown="pricingBreakdown = $event" />
        <ServicePricingPreview v-if="activeStep === 4" :form="form" :active="active" :pricing-rule="form.pricingRule" :parameters="form.parameters" :estimated-cost-rial="pricingCostEstimate" :breakdown="pricingBreakdown" :currency-unit="currencyUnit" />
        <ServiceTestPreview v-if="activeStep === 5" :form="form" :active="active" :parameters="form.parameters" :values="testValues" :materials="materials" :machines="machines" :result="testResult" :currency-unit="currencyUnit" @edit="activeStep = 2" />
        <template v-if="activeStep === 1">
        <div class="space-y-4">
          <ServiceOverviewIdentity :form="form" :active="active" />
          <ServiceOverviewSection title="Configuration" description="Current defaults for this service.">
            <dl class="divide-y divide-base-300/70">
              <div class="flex items-center justify-between gap-3 py-2.5 first:pt-0 text-sm"><dt class="text-base-content/60">Setup</dt><dd>{{ form.parameters.length ? `${form.parameters.length} input${form.parameters.length === 1 ? '' : 's'}` : 'Not configured yet' }}</dd></div>
              <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Costs</dt><dd>{{ form.components.length ? `${form.components.length} component${form.components.length === 1 ? '' : 's'}` : 'Not configured yet' }}</dd></div>
              <div class="flex items-center justify-between gap-3 py-2.5 text-sm"><dt class="text-base-content/60">Unit</dt><dd>{{ form.defaultUnit }}</dd></div>
              <div class="flex items-center justify-between gap-3 py-2.5 last:pb-0 text-sm"><dt class="text-base-content/60">Default priority</dt><dd>{{ form.defaultPriority }}</dd></div>
            </dl>
          </ServiceOverviewSection>
          <ServiceOverviewSection v-if="form.description" title="Description">
            <p class="text-sm leading-6 text-base-content/70">{{ form.description }}</p>
          </ServiceOverviewSection>
        </div>
        <div class="mt-4 flex gap-2 border-t border-base-300 pt-3 text-sm"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">Configure parameters that customers can choose when adding this service to an order.</p></div>
        </template>
        </aside>
      </div>
    </div>
  </div>
</template>
