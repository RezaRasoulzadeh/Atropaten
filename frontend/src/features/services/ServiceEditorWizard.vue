<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowLeft, ArrowRight, CircleHelp, FileText, ImagePlus, Package, Save, Upload, X } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import type { ServiceForm } from './types'

const props = defineProps<{
  form: ServiceForm
  editorMode: 'create' | 'edit'
  busy: boolean
  isSaving: boolean
  validationAttempted: boolean
  active: boolean
}>()
const emit = defineEmits<{
  cancel: []
  save: []
}>()

const activeStep = ref(1)
const imageInput = ref<HTMLInputElement | null>(null)
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
  const defaults = ['Print products', 'Finishing', 'Design', 'Packaging']
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
</script>

<template>
  <div class="service-wizard min-w-0 space-y-4" aria-label="Service editor">
    <header class="flex min-w-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 px-1 pt-4 pb-4">
      <div class="min-w-0">
        <h1 class="mt-2 text-2xl font-semibold tracking-tight text-primary">{{ editorMode === 'create' ? 'Add service' : 'Edit service' }}</h1>
        <p class="mt-1 text-sm text-base-content/65">Define a printing service that can be used in orders.</p>
        <WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Services' }, { label: editorMode === 'create' ? 'Add service' : 'Edit service', current: true }]" @navigate="emit('cancel')" />
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button class="btn btn-outline" type="button" :disabled="busy || isSaving" @click="emit('cancel')">Cancel</button>
        <button v-if="activeStep < steps.length" class="btn btn-primary gap-2" type="button" @click="next">Next <ArrowRight :size="16" aria-hidden="true" /></button>
        <button v-else class="btn btn-primary gap-2" type="button" :disabled="busy || isSaving" @click="emit('save')"><Save :size="16" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save service' }}</button>
      </div>
    </header>

    <div class="grid min-w-0 gap-4 xl:grid-cols-[16rem_minmax(0,1fr)_20rem]">
      <aside class="min-w-0 rounded-box border border-base-300 bg-base-200/35 p-3">
        <nav aria-label="Service setup steps" class="space-y-1">
          <button v-for="step in steps" :key="step.number" class="wizard-step w-full text-start" :class="stepClass(step.number)" type="button" @click="activeStep = step.number">
            <span class="wizard-step-number">{{ step.number < activeStep ? '✓' : step.number }}</span>
            <span class="min-w-0"><strong class="block text-sm">{{ step.title }}</strong><small class="mt-0.5 block text-xs leading-4 text-base-content/60">{{ step.description }}</small></span>
          </button>
        </nav>
        <button class="btn btn-outline mt-6 w-full justify-start gap-2" type="button" :disabled="busy || isSaving" @click="emit('save')"><Save :size="15" aria-hidden="true" />Save as draft</button>
      </aside>

      <main class="min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6">
        <form id="service-editor" class="service-editor min-w-0" @submit.prevent="next">
          <section v-if="activeStep === 1" class="min-w-0 space-y-6">
            <div class="flex items-start gap-3 border-b border-base-300 pb-4">
              <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileText :size="21" aria-hidden="true" /></span>
              <div><h2 class="text-lg font-semibold">Basic information</h2><p class="mt-1 text-sm text-base-content/65">Set the essential details for this service. You can configure other settings in the next steps.</p></div>
            </div>
            <div class="grid min-w-0 gap-4 sm:grid-cols-2">
              <FormField class="gap-1 sm:col-span-2"><span>Service name <em class="text-error">*</em></span><AppInput v-model="form.name" class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && !form.name.trim() }" type="text" required placeholder="Business card printing" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">A clear name shown to your team and customers.</small></FormField>
              <FormField class="gap-1"><span>Code <em class="text-error">*</em></span><AppInput v-model="form.code" class="input w-full min-w-0" type="text" required placeholder="SVC-001" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">An internal reference code.</small></FormField>
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
              <FormField class="gap-1 sm:col-span-2"><span>Service image <em>optional</em></span><input ref="imageInput" class="hidden" type="file" accept="image/png,image/jpeg,image/webp,image/gif" @change="onImageSelected" /><div class="flex flex-wrap items-center gap-3 rounded-box border border-dashed border-base-300 bg-base-100 p-3"><div v-if="imagePreview" class="relative size-20 shrink-0 overflow-hidden rounded-box border border-base-300 bg-base-200"><img :src="imagePreview" alt="Service preview" class="size-full object-cover" /><button class="btn btn-circle btn-error btn-xs absolute right-1 top-1" type="button" aria-label="Remove service image" @click="clearImage"><X :size="12" aria-hidden="true" /></button></div><div class="min-w-0"><strong class="block text-sm">{{ imagePreview ? 'Service image selected' : 'Add a service image' }}</strong><small class="mt-1 block text-xs leading-5 text-base-content/60">Use a clear image to recognize this service in lists and orders.</small></div><button class="btn btn-outline btn-sm ml-auto gap-2" type="button" @click="browseImage"><Upload :size="14" aria-hidden="true" />{{ imagePreview ? 'Replace image' : 'Choose image' }}</button></div></FormField>
              <FormField class="gap-1 sm:col-span-2"><span>Description</span><AppTextarea v-model="form.description" class="textarea w-full min-w-0" rows="6" maxlength="500" placeholder="High-quality printing with paper options and finishing." /><small class="text-xs leading-5 text-base-content/60">Add helpful details for your team and customers.</small></FormField>
            </div>
          </section>

          <section v-else class="flex min-h-[28rem] flex-col items-center justify-center text-center">
            <span class="grid size-14 place-items-center rounded-full bg-primary/15 text-primary"><CircleHelp :size="27" aria-hidden="true" /></span>
            <h2 class="mt-4 text-xl font-semibold">{{ steps[activeStep - 1].title }} is coming soon</h2>
            <p class="mt-2 max-w-md text-sm leading-6 text-base-content/65">This step is intentionally left for the next part of the service setup. Your existing service data is kept safe while we build it.</p>
            <button class="btn btn-outline mt-5 gap-2" type="button" @click="previous"><ArrowLeft :size="15" aria-hidden="true" />Back to basic information</button>
          </section>
        </form>
      </main>

      <aside class="min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4">
        <div class="flex items-start gap-3 border-b border-base-300 pb-4"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Package :size="19" aria-hidden="true" /></span><div><h2 class="text-base font-semibold">Live summary</h2><p class="mt-1 text-xs leading-5 text-base-content/60">A quick preview of this service.</p></div></div>
        <div class="mt-4 rounded-box border border-base-300 bg-base-100 p-4">
          <div class="relative grid aspect-[16/8] place-items-center overflow-hidden rounded-box bg-base-200 text-base-content/25"><img v-if="imagePreview" :src="imagePreview" alt="" class="size-full object-cover" /><ImagePlus v-else :size="42" stroke-width="1.2" aria-hidden="true" /></div>
          <h3 class="mt-4 break-words text-lg font-semibold">{{ form.name || 'Your service name' }}</h3>
          <p class="mt-1 break-words text-sm text-base-content/60">{{ form.code || 'SVC-001' }}<span v-if="form.category"> · {{ form.category }}</span></p>
          <StatusBadge class="mt-3" :label="active ? 'Active' : 'Archived'" :tone="active ? 'green' : 'slate'" />
          <dl class="mt-4 divide-y divide-base-300 rounded-box border border-base-300">
            <div class="flex items-center justify-between gap-3 px-3 py-2.5 text-sm"><dt class="text-base-content/60">Setup</dt><dd>{{ form.parameters.length ? `${form.parameters.length} input${form.parameters.length === 1 ? '' : 's'}` : 'Not configured yet' }}</dd></div>
            <div class="flex items-center justify-between gap-3 px-3 py-2.5 text-sm"><dt class="text-base-content/60">Costs</dt><dd>{{ form.components.length ? `${form.components.length} component${form.components.length === 1 ? '' : 's'}` : 'Not configured yet' }}</dd></div>
            <div class="flex items-center justify-between gap-3 px-3 py-2.5 text-sm"><dt class="text-base-content/60">Unit</dt><dd>{{ form.defaultUnit }}</dd></div>
            <div class="flex items-center justify-between gap-3 px-3 py-2.5 text-sm"><dt class="text-base-content/60">Default priority</dt><dd>{{ form.defaultPriority }}</dd></div>
          </dl>
          <p v-if="form.description" class="mt-4 border-t border-base-300 pt-4 text-sm leading-6 text-base-content/70">{{ form.description }}</p>
        </div>
        <div class="mt-4 flex gap-2 rounded-box border border-info/20 bg-info/5 p-3 text-sm"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70"><strong class="font-semibold text-info">Next step</strong><br />{{ activeStep === 1 ? 'Configure parameters that customers can choose when adding this service to an order.' : 'More setup options will be available here soon.' }}</p></div>
      </aside>
    </div>
  </div>
</template>
