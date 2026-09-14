<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { ArrowLeft, CheckCheck, Database, Package, Save, Warehouse } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import type { useMaterialsWorkspace } from './useMaterialsWorkspace'

const props = defineProps<{
  workspace: ReturnType<typeof useMaterialsWorkspace>
  currencyUnit: CurrencyUnit
}>()
const emit = defineEmits<{ cancel: [] }>()

const {
  busy,
  editorMode,
  form,
  attributeDefinitions,
  predefinedParameters,
  costDraft,
  isSaving,
  selectedMaterial,
  unitOptions,
  unitLabel,
  updateCost,
  saveMaterial,
} = props.workspace

const dimensionKeys = new Set(['width_mm', 'height_mm', 'length_mm'])
const specificationGroupOrder = ['Classification', 'Physical properties', 'Finish and appearance', 'Adhesive', 'Other specifications']
const printSizeDefinition = computed(() => predefinedParameters.value.find((definition) => definition.key === 'print_size'))
const physicalSizeOptions = computed(() => (printSizeDefinition.value?.options || [])
  .filter((option) => option.active && option.widthMM && option.heightMM)
  .map((option) => ({ label: `${option.label} (${option.widthMM} × ${option.heightMM} mm)`, value: option.code })))
const physicalSizeValue = computed(() => {
  const width = String(attributeValue('width_mm'))
  const height = String(attributeValue('height_mm'))
  return (printSizeDefinition.value?.options || []).find((option) => option.widthMM === width && option.heightMM === height)?.code || ''
})

const applicableDefinitions = computed(() => attributeDefinitions.value
  .filter((definition) => definition.active && definition.applicableKinds.includes(form.value.kind))
  .sort((left, right) => left.position - right.position || left.key.localeCompare(right.key)))
const specificationGroups = computed(() => {
  const groups = new Map<string, typeof applicableDefinitions.value>()
  for (const definition of applicableDefinitions.value) {
    if (dimensionKeys.has(definition.key)) continue
    const group = definition.key === 'material_subtype'
      ? 'Classification'
      : ['grammage_gsm', 'thickness_micron'].includes(definition.key)
        ? 'Physical properties'
        : ['finish', 'coating', 'color'].includes(definition.key)
          ? 'Finish and appearance'
          : definition.key === 'adhesive_type' ? 'Adhesive' : 'Other specifications'
    const definitions = groups.get(group) || []
    definitions.push(definition)
    groups.set(group, definitions)
  }
  return specificationGroupOrder
    .filter((group) => groups.has(group))
    .map((group) => ({ name: group, definitions: groups.get(group)! }))
})

const kindOptions = [
  ['sheet-stock', 'Sheet stock'], ['roll-media', 'Roll media'], ['board', 'Board'], ['ink', 'Ink'],
  ['lamination-film', 'Lamination film'], ['adhesive', 'Adhesive'], ['fabric', 'Fabric'],
  ['packaging', 'Packaging'], ['chemical', 'Chemical'], ['generic-consumable', 'Generic consumable'],
].map(([value, label]) => ({ value, label }))

function attributeValue(key: string) {
  const attribute = form.value.attributes.find((item) => item.key === key)
  if (!attribute) return ''
  switch (attribute.valueType) {
    case 'decimal': return attribute.decimalValue || ''
    case 'integer': return attribute.integerValue > 0 ? attribute.integerValue.toString() : ''
    case 'enum': return attribute.enumCode || ''
    case 'text': return attribute.textValue || ''
    case 'boolean': return attribute.booleanValue
    default: return ''
  }
}
function setAttribute(key: string, valueType: string, value: string | boolean) {
  if (valueType !== 'boolean' && String(value).trim() === '') {
    form.value.attributes = form.value.attributes.filter((item) => item.key !== key)
    return
  }
  let attribute = form.value.attributes.find((item) => item.key === key)
  if (!attribute) {
    attribute = { key, valueType, decimalValue: '', integerValue: 0, enumCode: '', textValue: '', booleanValue: false }
    form.value.attributes.push(attribute)
  }
  if (valueType === 'decimal') attribute.decimalValue = String(value)
  else if (valueType === 'integer') attribute.integerValue = Number(value) || 0
  else if (valueType === 'enum') attribute.enumCode = String(value)
  else if (valueType === 'text') attribute.textValue = String(value)
  else if (valueType === 'boolean') attribute.booleanValue = Boolean(value)
}
function enumOptions(definition: { enumOptions: { code: string; label: string; active: boolean; position: number }[] }) {
  return [
    { label: 'Not specified', value: '' },
    ...[...(definition.enumOptions || [])]
      .filter((option) => option.active)
      .sort((left, right) => left.position - right.position || left.code.localeCompare(right.code))
      .map((option) => ({ label: option.label, value: option.code })),
  ]
}
function showDimensions() { return applicableDefinitions.value.some((definition) => definition.key === 'width_mm') }
function showHeight() { return applicableDefinitions.value.some((definition) => definition.key === 'height_mm') }
function showLength() { return applicableDefinitions.value.some((definition) => definition.key === 'length_mm') }
function selectPhysicalSize(code: string) {
  const option = printSizeDefinition.value?.options.find((item) => item.code === code)
  if (!option?.widthMM || !option.heightMM) return
  setAttribute('width_mm', 'decimal', option.widthMM)
  setAttribute('height_mm', 'decimal', option.heightMM)
}
function updateKind(kind: string) {
  form.value.kind = kind
  const applicableKeys = new Set(attributeDefinitions.value
    .filter((definition) => definition.active && definition.applicableKinds.includes(kind))
    .map((definition) => definition.key))
  form.value.attributes = form.value.attributes.filter((attribute) => applicableKeys.has(attribute.key))
}

const activeStep = ref(1)
const validationAttempted = ref(false)
const steps = [
  { number: 1, title: 'Identity & kind', description: 'Name, code, material type' },
  { number: 2, title: 'Specifications & stock', description: 'Size, properties, units' },
  { number: 3, title: 'Purchasing', description: 'Cost, supplier, notes' },
]

const isCreating = computed(() => editorMode.value === 'create')
const isEditing = computed(() => editorMode.value === 'edit')
const title = computed(() => (isCreating.value ? 'Add material' : 'Edit material'))
const unitSelectOptions = computed(() => unitOptions.map((unit) => ({ label: unitLabel(unit), value: unit })))
const materialTitle = computed(() => form.value.name.trim() || 'Your material')
const materialCode = computed(() => form.value.sku.trim() || 'No SKU')
const materialCategory = computed(() => form.value.category.trim() || 'Uncategorized')
const stockValue = computed(() => `${form.value.physicalStock || '0'} ${unitLabel(form.value.consumptionUnit)}`)
const conversionValue = computed(() => `1 ${unitLabel(form.value.purchaseUnit)} = ${form.value.conversionFactor || '…'} ${unitLabel(form.value.consumptionUnit)}`)
const pricingUnitCostValue = computed(() => {
  const value = selectedMaterial.value?.highestPurchaseUnitCostRial || form.value.averageUnitCostRial
  return value > 0 ? formatMoney(value, props.currencyUnit) : 'Not set'
})

function stepClass(number: number) {
  if (number === activeStep.value) return 'wizard-step-active'
  if (number < activeStep.value) return 'wizard-step-complete'
  return 'wizard-step-idle'
}

function focusInvalid(formElement: HTMLFormElement) {
  const invalid = formElement.querySelector<HTMLElement>(':invalid')
  invalid?.focus()
}

function validateCurrentStep() {
  validationAttempted.value = true
  const formElement = document.querySelector<HTMLFormElement>('#material-editor-wizard')
  if (!formElement) return true
  const valid = formElement.reportValidity()
  if (!valid) focusInvalid(formElement)
  return valid
}

function next() {
  if (busy.value || isSaving.value) return
  if (!validateCurrentStep()) return
  if (activeStep.value < steps.length) activeStep.value += 1
}

function previous() {
  if (activeStep.value > 1) activeStep.value -= 1
}

function onFormSubmit(event: SubmitEvent) {
  if (busy.value || isSaving.value) return
  // Explicit Save buttons save; Enter advances until the final step.
  if (!event.submitter && activeStep.value < steps.length) next()
  else submit()
}

function submit() {
  if (busy.value || isSaving.value) return
  validationAttempted.value = true
  if (!form.value.name.trim()) {
    activeStep.value = 1
    void nextTick(() => {
      document.querySelector<HTMLInputElement>('#material-editor-wizard input')?.focus()
    })
    return
  }
  const formElement = document.querySelector<HTMLFormElement>('#material-editor-wizard')
  if (formElement && !formElement.reportValidity()) {
    focusInvalid(formElement)
    return
  }
  void saveMaterial()
}
</script>

<template>
  <div class="service-wizard w-full flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Material editor">
    <header class="service-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-bold leading-8 tracking-tight text-primary">{{ title }}</h1>
        <p class="mt-1 text-xs leading-4 text-base-content/65">Define stock, units, and purchasing details for this material.</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Materials' }, { label: title, current: true }]" @navigate="emit('cancel')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-error" type="button" :disabled="busy || isSaving" @click="emit('cancel')">Cancel</button>
        <button class="btn btn-success gap-2" type="submit" form="material-editor-wizard" :disabled="busy || isSaving">
          <Save :size="16" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save' }}
        </button>
      </div>
    </header>

    <div class="service-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="service-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200">
        <nav aria-label="Material setup steps" class="service-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible">
          <button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" @click="activeStep = step.number">
            <span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span>
            <span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ step.title }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ step.description }}</small></span>
          </button>
        </nav>
      </aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="service-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
          <form id="material-editor-wizard" class="service-editor min-w-0" :aria-busy="busy || isSaving" @submit.prevent="onFormSubmit">
            <section v-if="activeStep === 1" class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4">
                <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Package :size="21" aria-hidden="true" /></span>
                <div><h2 class="text-lg font-semibold">Identify the material</h2><p class="text-sm text-base-content/60">Start by choosing what kind of inventory material this is.</p></div>
              </div>
              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <FormField class="gap-1 sm:col-span-2"><span>Name <em class="text-error">*</em></span><AppInput v-model="form.name" class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && !form.name.trim() }" type="text" required placeholder="A4 80gsm Paper" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Use the name your team will recognize when selecting material.</small></FormField>
                <FormField class="gap-1"><span>SKU / code</span><AppInput v-model="form.sku" class="input w-full min-w-0" type="text" placeholder="PAPER-A4" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Optional internal reference.</small></FormField>
                <FormField class="gap-1"><span>Category</span><AppInput v-model="form.category" class="input w-full min-w-0" type="text" placeholder="Paper" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Helps organize materials in lists and reports.</small></FormField>
                <FormField class="gap-1 sm:col-span-2"><span>Material kind <em class="text-error">*</em></span><SelectField v-model="form.kind" :options="kindOptions" aria-label="Material kind" @update:model-value="updateKind" /><small class="text-xs leading-5 text-base-content/60">Stable kind used for compatibility and consumption behavior; category remains an organizational label.</small></FormField>
              </div>
            </section>

            <section v-else-if="activeStep === 2" class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4">
                <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Warehouse :size="21" aria-hidden="true" /></span>
                <div><h2 class="text-lg font-semibold">Specifications &amp; stock</h2><p class="text-sm text-base-content/60">Set the applicable physical properties first, then define how this material is bought and consumed.</p></div>
              </div>
              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <SelectField v-model="form.purchaseUnit" label="Purchase unit" :options="unitSelectOptions" />
                <SelectField v-model="form.consumptionUnit" label="Consumption unit" :options="unitSelectOptions" />
                <FormField class="gap-1 sm:col-span-2"><span>Conversion factor</span><AppInput v-model="form.conversionFactor" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="500" /><small class="text-xs leading-5 text-base-content/60">{{ conversionValue }}</small></FormField>
                <FormField class="gap-1"><span>Opening physical stock</span><AppInput v-model="form.physicalStock" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" :disabled="isEditing" /><small class="text-xs leading-5 text-base-content/60">{{ isEditing ? 'Use Adjust stock for later ledger movements.' : 'Starting quantity in the consumption unit.' }}</small></FormField>
                <FormField class="gap-1"><span>Reorder level</span><AppInput v-model="form.reorderLevel" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" /><small class="text-xs leading-5 text-base-content/60">Warn when available stock reaches this level.</small></FormField>
              </div>
              <div v-if="showDimensions()" class="space-y-3 rounded-box border border-primary/25 bg-primary/5 p-4">
                <div><h3 class="text-sm font-semibold">Physical inventory specifications</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Stored canonically in millimetres. These describe the source inventory material, not the customer’s finished size.</p></div>
                <SelectField v-if="(form.kind === 'sheet-stock' || form.kind === 'board') && physicalSizeOptions.length" :model-value="physicalSizeValue" label="Common source size" :options="[{ label: 'Enter custom dimensions', value: '' }, ...physicalSizeOptions]" @update:model-value="selectPhysicalSize" />
                <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                  <FormField class="gap-1"><span>Physical width (mm) <em class="text-error">*</em></span><AppInput :model-value="String(attributeValue('width_mm'))" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="320" @update:model-value="setAttribute('width_mm', 'decimal', $event)" /></FormField>
                  <FormField v-if="showHeight()" class="gap-1"><span>Physical height (mm) <em class="text-error">*</em></span><AppInput :model-value="String(attributeValue('height_mm'))" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="450" @update:model-value="setAttribute('height_mm', 'decimal', $event)" /></FormField>
                  <FormField v-else-if="showLength()" class="gap-1"><span>Optional roll length (mm)</span><AppInput :model-value="String(attributeValue('length_mm'))" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="Optional" @update:model-value="setAttribute('length_mm', 'decimal', $event)" /></FormField>
                </div>
              </div>
              <div class="space-y-4 rounded-box border border-base-300 bg-base-100 p-4">
                <div>
                  <h3 class="text-sm font-semibold">Material specifications</h3>
                  <p class="mt-1 text-xs leading-5 text-base-content/60">Complete known values; migrated materials intentionally remain blank until verified. These fields are driven by the selected material kind.</p>
                </div>
                <div v-if="!specificationGroups.length" class="rounded-box border border-dashed border-base-300 px-3 py-4 text-sm text-base-content/60">This material kind has no additional specifications yet.</div>
                <div v-for="group in specificationGroups" :key="group.name" class="space-y-3">
                  <h4 class="text-xs font-semibold uppercase tracking-wide text-base-content/55">{{ group.name }}</h4>
                  <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                    <FormField v-for="definition in group.definitions" :key="definition.key" class="gap-1">
                      <span>{{ definition.label }}<span v-if="definition.unit"> ({{ definition.unit }})</span></span>
                      <SelectField v-if="definition.valueType === 'enum'" :model-value="String(attributeValue(definition.key))" :options="enumOptions(definition)" :aria-label="definition.label" @update:model-value="setAttribute(definition.key, definition.valueType, $event)" />
                      <div v-else-if="definition.valueType === 'boolean'" class="flex h-10 items-center gap-2"><input class="checkbox" type="checkbox" :checked="Boolean(attributeValue(definition.key))" :aria-label="definition.label" @change="setAttribute(definition.key, definition.valueType, ($event.target as HTMLInputElement).checked)" /><span class="text-sm">Enabled</span></div>
                      <AppInput v-else :model-value="String(attributeValue(definition.key))" class="input w-full min-w-0" :type="definition.valueType === 'text' ? 'text' : 'number'" :min="definition.valueType === 'integer' || definition.valueType === 'decimal' ? 0 : undefined" :step="definition.valueType === 'decimal' ? 'any' : undefined" @update:model-value="setAttribute(definition.key, definition.valueType, $event)" />
                    </FormField>
                  </div>
                </div>
              </div>
            </section>

            <section v-else class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4">
                <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Database :size="21" aria-hidden="true" /></span>
                <div><h2 class="text-lg font-semibold">Purchasing details</h2><p class="text-sm text-base-content/60">Keep the cost basis and operational context with the material.</p></div>
              </div>
              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <FormField v-if="isCreating" class="gap-1 sm:col-span-2"><span>Opening unit cost / {{ form.consumptionUnit }} ({{ props.currencyUnit }})</span><AppInput :model-value="costDraft" :money="props.currencyUnit" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" @update:model-value="updateCost" /><small class="text-xs leading-5 text-base-content/60">Used for opening stock only. Posted purchases become the authoritative pricing basis.</small></FormField>
                <div v-else class="rounded-box border border-base-300 bg-base-100/45 p-4 sm:col-span-2"><span class="block text-xs text-base-content/60">Current pricing basis</span><strong class="mt-1 block text-lg font-semibold text-primary">{{ pricingUnitCostValue }}</strong><p class="mt-1 text-xs leading-5 text-base-content/60">Pricing uses the highest landed purchase cost when available.</p></div>
                <FormField class="gap-1 sm:col-span-2"><span>Preferred supplier <em class="text-base-content/45">optional</em></span><AppInput v-model="form.preferredSupplier" class="input w-full min-w-0" type="text" placeholder="Pars Paper" autocomplete="off" /></FormField>
                <FormField class="gap-1 sm:col-span-2"><span>Notes <em class="text-base-content/45">optional</em></span><AppTextarea v-model="form.notes" class="textarea w-full min-w-0" rows="6" placeholder="Storage, handling, or purchasing notes" /></FormField>
              </div>
            </section>
          </form>
        </section>

        <aside class="service-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto">
          <div class="space-y-4">
            <div class="relative min-h-48 overflow-hidden rounded-box bg-base-100 p-5">
              <div class="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-transparent" aria-hidden="true"></div>
              <div class="relative flex min-h-36 flex-col justify-end">
                <span class="mb-auto grid size-11 place-items-center rounded-box border border-base-300 bg-base-200/80 text-primary"><Package :size="24" aria-hidden="true" /></span>
                <div class="mt-6 flex items-end justify-between gap-3"><div class="min-w-0"><h2 class="truncate text-lg font-semibold">{{ materialTitle }}</h2><p class="mt-1 truncate text-xs text-base-content/60">{{ materialCode }} · {{ materialCategory }}</p></div><StatusBadge label="Active" tone="green" /></div>
              </div>
            </div>

            <div class="divide-y divide-base-300 rounded-box border border-base-300 bg-base-100/35">
              <div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Purchase unit</span><strong>{{ unitLabel(form.purchaseUnit) }}</strong></div>
              <div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Conversion</span><strong class="text-end text-xs">{{ conversionValue }}</strong></div>
              <div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Available stock</span><strong>{{ stockValue }}</strong></div>
              <div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Pricing unit cost</span><strong class="text-primary">{{ pricingUnitCostValue }}</strong></div>
              <div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Supplier</span><strong class="max-w-[9rem] truncate text-end">{{ form.preferredSupplier || 'Not specified' }}</strong></div>
            </div>

            <div class="flex gap-2 border-t border-base-300 pt-3 text-sm"><Package class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">{{ activeStep === 1 ? 'Start with the catalog identity, then continue to stock setup.' : activeStep === 2 ? 'Conversion keeps purchase quantities consistent with production usage.' : 'Save the material when its purchasing details are ready.' }}</p></div>
          </div>
        </aside>
      </div>
    </div>
    <footer class="flex min-w-0 items-center justify-between gap-3 border-t border-base-300 px-1 pt-3"><button class="btn btn-ghost btn-sm" type="button" :disabled="activeStep === 1 || busy || isSaving" @click="previous">Back</button><span class="text-xs text-base-content/55">Step {{ activeStep }} of {{ steps.length }}</span><button v-if="activeStep < steps.length" class="btn btn-primary btn-sm" type="button" @click="next">Continue</button><button v-else class="btn btn-success btn-sm gap-2" type="submit" form="material-editor-wizard" :disabled="busy || isSaving"><Save :size="14" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save material' }}</button></footer>
  </div>
</template>
