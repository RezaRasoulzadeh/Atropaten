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
  costDraft,
  isSaving,
  selectedMaterial,
  unitOptions,
  unitLabel,
  updateCost,
  saveMaterial,
} = props.workspace

const activeStep = ref(1)
const validationAttempted = ref(false)
const steps = [
  { number: 1, title: 'Basic', description: 'Name, category, code' },
  { number: 2, title: 'Stock setup', description: 'Units, conversion, stock' },
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
const costValue = computed(() => {
  const value = Number(String(costDraft.value).replace(/[^0-9.-]/g, ''))
  return Number.isFinite(value) && value > 0 ? formatMoney(value, props.currencyUnit) : 'Not set'
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
  if (!validateCurrentStep()) return
  if (activeStep.value < steps.length) activeStep.value += 1
}

function previous() {
  if (activeStep.value > 1) activeStep.value -= 1
}

function submit() {
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
  <div class="service-wizard flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Material editor">
    <header class="service-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-semibold tracking-tight text-primary">{{ title }}</h1>
        <p class="mt-1 text-sm text-base-content/65">Define stock, units, and purchasing details for this material.</p>
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
          <form id="material-editor-wizard" class="service-editor min-w-0" @submit.prevent="submit">
            <section v-if="activeStep === 1" class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4">
                <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Package :size="21" aria-hidden="true" /></span>
                <div><h2 class="text-lg font-semibold">Basic details</h2><p class="text-sm text-base-content/60">Identify the material in your catalog.</p></div>
              </div>
              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <FormField class="gap-1 sm:col-span-2"><span>Name <em class="text-error">*</em></span><AppInput v-model="form.name" class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && !form.name.trim() }" type="text" required placeholder="A4 80gsm Paper" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Use the name your team will recognize when selecting material.</small></FormField>
                <FormField class="gap-1"><span>SKU / code</span><AppInput v-model="form.sku" class="input w-full min-w-0" type="text" placeholder="PAPER-A4" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Optional internal reference.</small></FormField>
                <FormField class="gap-1"><span>Category</span><AppInput v-model="form.category" class="input w-full min-w-0" type="text" placeholder="Paper" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Helps organize materials in lists and reports.</small></FormField>
              </div>
            </section>

            <section v-else-if="activeStep === 2" class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4">
                <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Warehouse :size="21" aria-hidden="true" /></span>
                <div><h2 class="text-lg font-semibold">Stock setup</h2><p class="text-sm text-base-content/60">Tell the system how this material is bought and consumed.</p></div>
              </div>
              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <SelectField v-model="form.purchaseUnit" label="Purchase unit" :options="unitSelectOptions" />
                <SelectField v-model="form.consumptionUnit" label="Consumption unit" :options="unitSelectOptions" />
                <FormField class="gap-1 sm:col-span-2"><span>Conversion factor</span><AppInput v-model="form.conversionFactor" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="500" /><small class="text-xs leading-5 text-base-content/60">{{ conversionValue }}</small></FormField>
                <FormField class="gap-1"><span>Opening physical stock</span><AppInput v-model="form.physicalStock" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" :disabled="isEditing" /><small class="text-xs leading-5 text-base-content/60">{{ isEditing ? 'Use Adjust stock for later ledger movements.' : 'Starting quantity in the consumption unit.' }}</small></FormField>
                <FormField class="gap-1"><span>Reorder level</span><AppInput v-model="form.reorderLevel" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" /><small class="text-xs leading-5 text-base-content/60">Warn when available stock reaches this level.</small></FormField>
              </div>
            </section>

            <section v-else class="min-w-0 space-y-6">
              <div class="flex items-center gap-3 border-b border-base-300 pb-4">
                <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Database :size="21" aria-hidden="true" /></span>
                <div><h2 class="text-lg font-semibold">Purchasing details</h2><p class="text-sm text-base-content/60">Keep the cost basis and operational context with the material.</p></div>
              </div>
              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <FormField v-if="isCreating" class="gap-1 sm:col-span-2"><span>Opening unit cost / {{ form.consumptionUnit }} ({{ props.currencyUnit }})</span><AppInput :model-value="costDraft" :money="props.currencyUnit" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" @update:model-value="updateCost" /><small class="text-xs leading-5 text-base-content/60">Used for opening stock only. Posted purchases become the authoritative pricing basis.</small></FormField>
                <div v-else class="rounded-box border border-base-300 bg-base-100/45 p-4 sm:col-span-2"><span class="block text-xs text-base-content/60">Current pricing basis</span><strong class="mt-1 block text-lg font-semibold text-primary">{{ selectedMaterial ? formatMoney(selectedMaterial.highestPurchaseUnitCostRial || selectedMaterial.averageUnitCostRial, props.currencyUnit) : costValue }}</strong><p class="mt-1 text-xs leading-5 text-base-content/60">Pricing uses the highest landed purchase cost when available.</p></div>
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
              <div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Unit cost</span><strong class="text-primary">{{ costValue }}</strong></div>
              <div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Supplier</span><strong class="max-w-[9rem] truncate text-end">{{ form.preferredSupplier || 'Not specified' }}</strong></div>
            </div>

            <div class="flex gap-2 border-t border-base-300 pt-3 text-sm"><Package class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">{{ activeStep === 1 ? 'Start with the catalog identity, then continue to stock setup.' : activeStep === 2 ? 'Conversion keeps purchase quantities consistent with production usage.' : 'Save the material when its purchasing details are ready.' }}</p></div>
          </div>
        </aside>
      </div>
    </div>
  </div>
</template>
