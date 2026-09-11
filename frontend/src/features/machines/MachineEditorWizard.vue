<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { CheckCheck, Factory, FileText, Gauge, Layers3, Save, Upload, X } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceBreadcrumb from '../../components/layout/WorkspaceBreadcrumb.vue'
import type { CurrencyUnit } from '../../utils/currency'
import { formatMoney } from '../../utils/currency'
import type { useMachinesWorkspace } from './useMachinesWorkspace'

const props = defineProps<{
  workspace: ReturnType<typeof useMachinesWorkspace>
  currencyUnit: CurrencyUnit
}>()
const emit = defineEmits<{ cancel: [] }>()

const { busy, editorMode, form, isSaving, basisLabel, emptyRate, saveMachine } = props.workspace
const activeStep = ref(1)
const validationAttempted = ref(false)
const imageInput = ref<HTMLInputElement | null>(null)
const steps = [
  { number: 1, title: 'Basic', description: 'Name, category, basis' },
  { number: 2, title: 'Rate profiles', description: 'Standard and alternatives' },
  { number: 3, title: 'Notes & review', description: 'Operating context' },
]
const isCreating = computed(() => editorMode.value === 'create')
const title = computed(() => isCreating.value ? 'Add machine' : 'Edit machine')
const machineTitle = computed(() => form.value.name.trim() || 'Your machine')
const machineCode = computed(() => form.value.code.trim() || 'No code')
const machineCategory = computed(() => form.value.category.trim() || 'Uncategorized')
const profileCount = computed(() => form.value.rates.length + 1)
const imagePreview = computed(() => form.value.imagePath || '')

function stepClass(number: number) {
  if (number === activeStep.value) return 'wizard-step-active'
  if (number < activeStep.value) return 'wizard-step-complete'
  return 'wizard-step-idle'
}

function validateCurrentStep() {
  validationAttempted.value = true
  const formElement = document.querySelector<HTMLFormElement>('#machine-editor-wizard')
  if (!formElement) return true
  const valid = formElement.reportValidity()
  if (!valid) formElement.querySelector<HTMLElement>(':invalid')?.focus()
  return valid
}

function next() {
  if (!validateCurrentStep()) return
  if (activeStep.value < steps.length) activeStep.value += 1
}

function submit() {
  validationAttempted.value = true
  if (!form.value.name.trim() || !form.value.rate.trim()) {
    activeStep.value = !form.value.name.trim() ? 1 : 2
    void nextTick(() => document.querySelector<HTMLInputElement>('#machine-editor-wizard input')?.focus())
    return
  }
  const formElement = document.querySelector<HTMLFormElement>('#machine-editor-wizard')
  if (formElement && !formElement.reportValidity()) {
    formElement.querySelector<HTMLElement>(':invalid')?.focus()
    return
  }
  void saveMachine()
}

function addRate() {
  form.value.rates.push(emptyRate(form.value.rates.length))
}

function browseImage() {
  imageInput.value?.click()
}

function onImageSelected(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !file.type.startsWith('image/')) return
  const reader = new FileReader()
  reader.onload = () => {
    if (typeof reader.result === 'string') form.value.imagePath = reader.result
  }
  reader.readAsDataURL(file)
  if (imageInput.value) imageInput.value.value = ''
}

function clearImage() {
  form.value.imagePath = ''
}
</script>

<template>
  <div class="service-wizard flex min-h-0 min-w-0 flex-col gap-4 overflow-visible xl:h-full xl:overflow-hidden" aria-label="Machine editor">
    <header class="service-wizard-header flex min-w-0 shrink-0 flex-wrap items-end justify-between gap-4 border-b border-base-300 bg-base-200 px-1 pt-4 pb-4">
      <div class="min-w-0"><h1 class="mt-2 text-2xl font-semibold tracking-tight text-primary">{{ title }}</h1><p class="mt-1 text-sm text-base-content/65">Define reusable equipment and rates for service costing.</p><WorkspaceBreadcrumb class="mt-2" :items="[{ label: 'Machines' }, { label: title, current: true }]" @navigate="emit('cancel')" /></div>
      <div class="flex shrink-0 items-center gap-2"><button class="btn btn-error" type="button" :disabled="busy || isSaving" @click="emit('cancel')">Cancel</button><button class="btn btn-success gap-2" type="submit" form="machine-editor-wizard" :disabled="busy || isSaving"><Save :size="16" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save' }}</button></div>
    </header>

    <div class="service-wizard-main flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-visible xl:overflow-hidden">
      <aside class="service-wizard-steps flex min-w-0 shrink-0 flex-col border-b border-base-300 pb-3 xl:sticky xl:top-0 xl:z-20 xl:bg-base-200"><nav aria-label="Machine setup steps" class="service-wizard-step-nav flex min-w-0 gap-1 overflow-x-auto pb-1 xl:overflow-visible"><button v-for="step in steps" :key="step.number" class="wizard-step w-auto min-w-[11rem] shrink-0 text-start xl:min-w-0 xl:flex-1" :class="stepClass(step.number)" type="button" @click="activeStep = step.number"><span class="wizard-step-number"><CheckCheck v-if="step.number < activeStep" :size="17" :stroke-width="2.2" aria-hidden="true" /><span v-else>{{ step.number }}</span></span><span class="min-w-0"><strong class="block truncate whitespace-nowrap text-sm">{{ step.title }}</strong><small class="mt-0.5 block truncate whitespace-nowrap text-xs leading-4 text-base-content/60">{{ step.description }}</small></span></button></nav></aside>

      <div class="grid min-h-0 min-w-0 flex-1 gap-4 overflow-visible xl:grid-cols-[minmax(0,1fr)_20rem] xl:overflow-hidden">
        <section class="service-wizard-form-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-6 xl:overflow-y-auto">
          <form id="machine-editor-wizard" class="service-editor min-w-0" @submit.prevent="submit">
            <section v-if="activeStep === 1" class="min-w-0 space-y-6"><div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Factory :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Basic details</h2><p class="text-sm text-base-content/60">Identify the equipment and choose its default billing basis.</p></div></div><div class="grid min-w-0 gap-4 sm:grid-cols-2"><FormField class="gap-1 sm:col-span-2"><span>Name <em class="text-error">*</em></span><AppInput v-model="form.name" class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && !form.name.trim() }" type="text" required placeholder="Full color printer" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Use a name your team will recognize when selecting a machine.</small></FormField><FormField class="gap-1"><span>Code</span><AppInput v-model="form.code" class="input w-full min-w-0" type="text" placeholder="PRINTER-01" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Optional internal reference.</small></FormField><FormField class="gap-1"><span>Category</span><AppInput v-model="form.category" class="input w-full min-w-0" type="text" placeholder="Digital printing" autocomplete="off" /><small class="text-xs leading-5 text-base-content/60">Helps organize production equipment.</small></FormField><SelectField v-model="form.rateBasis" label="Default rate basis" :options="[{ label: 'Per unit / page', value: 'unit' }, { label: 'Per minute', value: 'minute' }, { label: 'Per hour', value: 'hour' }]" /><div class="rounded-box border border-base-300 bg-base-100/45 p-3 text-sm sm:col-span-1"><span class="block text-xs text-base-content/55">How it works</span><p class="mt-1 leading-5 text-base-content/70">Add alternative profiles in the next step when one machine has different rates for color, media, or mode.</p></div><FormField class="gap-1 sm:col-span-2"><span>Machine image <em class="text-base-content/45">optional</em></span><input ref="imageInput" class="hidden" type="file" accept="image/png,image/jpeg,image/webp,image/gif" @change="onImageSelected" /><div class="flex flex-wrap items-center gap-3 rounded-box border border-dashed border-base-300 bg-base-100/45 p-3"><div v-if="imagePreview" class="relative size-20 shrink-0 overflow-hidden rounded-box border border-base-300 bg-base-200"><img :src="imagePreview" alt="Machine preview" class="size-full object-cover" /><button class="btn btn-circle btn-error btn-xs absolute right-1 top-1" type="button" aria-label="Remove machine image" @click.stop.prevent="clearImage"><X :size="12" aria-hidden="true" /></button></div><div class="min-w-0"><strong class="block text-sm">{{ imagePreview ? 'Machine image selected' : 'Add a machine image' }}</strong><small class="mt-1 block text-xs leading-5 text-base-content/60">Use a clear image to recognize this machine in lists and costing screens.</small></div><button class="btn btn-outline btn-sm ml-auto gap-2" type="button" @click.stop.prevent="browseImage"><Upload :size="14" aria-hidden="true" />{{ imagePreview ? 'Replace image' : 'Choose image' }}</button></div></FormField></div></section>

            <section v-else-if="activeStep === 2" class="min-w-0 space-y-6"><div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Gauge :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Rate profiles</h2><p class="text-sm text-base-content/60">Set the standard rate and add alternatives linked to service parameters.</p></div></div><div class="rounded-box border border-base-300 bg-base-100/35 p-4"><div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><Layers3 :size="18" aria-hidden="true" /></span><div><h3 class="text-sm font-semibold">Standard profile</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Used when no alternate profile matches the order parameter.</p></div></div><div class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2"><FormField class="gap-1"><span>Rate <em class="text-error">*</em></span><AppInput v-model="form.rate" :money="props.currencyUnit" class="input w-full min-w-0" :class="{ 'input-error': validationAttempted && !form.rate.trim() }" type="text" inputmode="decimal" required placeholder="0" /><small class="text-xs text-base-content/60">{{ basisLabel(form.rateBasis) }}</small></FormField><FormField class="gap-1"><span>Setup / fixed cost <em class="text-base-content/45">optional</em></span><AppInput v-model="form.setupCost" :money="props.currencyUnit" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" /></FormField></div></div><div class="rounded-box border border-base-300 bg-base-100/35 p-4"><div class="flex items-start justify-between gap-3"><div><h3 class="text-sm font-semibold">Alternative profiles</h3><p class="mt-1 text-xs leading-5 text-base-content/60">Examples: Black &amp; white, Full color, or Matte mode.</p></div><button class="btn btn-outline btn-sm shrink-0 gap-2" type="button" @click="addRate"><Layers3 :size="14" aria-hidden="true" />Add profile</button></div><div v-if="form.rates.length" class="mt-4 space-y-3"><div v-for="(item, index) in form.rates" :key="item.id" class="rounded-box border border-base-300 p-3"><div class="flex items-center justify-between gap-2"><strong class="text-sm">Profile {{ index + 2 }}</strong><button class="btn btn-ghost btn-xs text-error" type="button" @click="form.rates.splice(index, 1)">Remove</button></div><div class="mt-3 grid min-w-0 gap-3 sm:grid-cols-2"><FormField class="gap-1"><span>Profile name <em class="text-error">*</em></span><AppInput v-model="item.name" class="input w-full min-w-0" type="text" required placeholder="Full color" /></FormField><FormField class="gap-1"><span>Matches parameter value</span><AppInput v-model="item.selectorValue" class="input w-full min-w-0" type="text" placeholder="Full color" /></FormField><SelectField v-model="item.rateBasis" label="Rate basis" :options="[{ label: 'Per unit / page', value: 'unit' }, { label: 'Per minute', value: 'minute' }, { label: 'Per hour', value: 'hour' }]" /><FormField class="gap-1"><span>Rate <em class="text-error">*</em></span><AppInput v-model="item.rate" :money="props.currencyUnit" class="input w-full min-w-0" type="text" inputmode="decimal" required placeholder="0" /></FormField><FormField class="gap-1 sm:col-span-2"><span>Setup / fixed cost optional</span><AppInput v-model="item.setupCost" :money="props.currencyUnit" class="input w-full min-w-0" type="text" inputmode="decimal" placeholder="0" /></FormField></div></div></div><p v-else class="mt-4 rounded-box border border-dashed border-base-300 p-4 text-sm text-base-content/60">No alternatives yet. The standard rate will be used.</p></div></section>

            <section v-else class="min-w-0 space-y-6"><div class="flex items-center gap-3 border-b border-base-300 pb-4"><span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/10 text-primary"><FileText :size="21" aria-hidden="true" /></span><div><h2 class="text-lg font-semibold">Notes &amp; review</h2><p class="text-sm text-base-content/60">Add operating context and confirm the rate setup before saving.</p></div></div><FormField class="gap-1"><span>Notes <em class="text-base-content/45">optional</em></span><AppTextarea v-model="form.notes" class="textarea w-full min-w-0" rows="7" placeholder="Capacity, operating notes, or rate context" /><small class="text-xs leading-5 text-base-content/60">These notes stay with the machine and help operators understand its rates.</small></FormField><div class="rounded-box border border-base-300 bg-base-100/45 p-4"><h3 class="text-sm font-semibold">Ready to save?</h3><dl class="mt-3 divide-y divide-base-300/70 text-sm"><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Machine</dt><dd class="max-w-[60%] truncate text-end">{{ machineTitle }}</dd></div><div class="flex justify-between gap-3 py-2"><dt class="text-base-content/60">Default basis</dt><dd class="text-end">{{ basisLabel(form.rateBasis) }}</dd></div><div class="flex justify-between gap-3 py-2 last:pb-0"><dt class="text-base-content/60">Profiles</dt><dd>{{ profileCount }}</dd></div></dl></div></section>
          </form>
        </section>

        <aside class="service-wizard-preview-panel min-h-0 min-w-0 rounded-box border border-base-300 bg-base-200/35 p-4 xl:overflow-y-auto"><div class="space-y-4"><div class="relative min-h-48 overflow-hidden rounded-box bg-base-100 bg-cover bg-center p-5" :style="imagePreview ? { backgroundImage: `url('${imagePreview}')` } : undefined"><div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/45 to-black/10" aria-hidden="true"></div><div v-if="!imagePreview" class="absolute inset-0 grid place-items-center text-base-content/30"><Factory :size="42" aria-hidden="true" /></div><div class="relative z-10 flex min-h-36 flex-col justify-end text-white"><span v-if="!imagePreview" class="mb-auto grid size-11 place-items-center rounded-box border border-white/15 bg-black/20 text-white/80"><Factory :size="24" aria-hidden="true" /></span><div class="mt-auto flex min-w-0 items-end justify-between gap-3"><div class="min-w-0"><h2 class="truncate text-lg font-semibold">{{ machineTitle }}</h2><p class="mt-1 truncate text-xs text-white/70">{{ machineCode }} · {{ machineCategory }}</p></div><StatusBadge label="Active" tone="green" /></div></div></div><div class="divide-y divide-base-300 rounded-box border border-base-300 bg-base-100/35"><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Standard rate</span><strong class="text-primary">{{ form.rate ? formatMoney(Number(String(form.rate).replace(/[^0-9.-]/g, '')), props.currencyUnit) : 'Not set' }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Rate basis</span><strong class="max-w-[10rem] truncate text-end">{{ basisLabel(form.rateBasis) }}</strong></div><div class="flex items-center justify-between gap-3 px-3 py-3 text-sm"><span class="text-base-content/60">Profiles</span><strong>{{ profileCount }}</strong></div></div><div class="flex gap-2 border-t border-base-300 pt-3 text-sm"><Factory class="mt-0.5 shrink-0 text-info" :size="17" aria-hidden="true" /><p class="leading-5 text-base-content/70">{{ activeStep === 1 ? 'Start with the equipment identity and default billing basis.' : activeStep === 2 ? 'Alternative profiles let one machine price different production modes.' : 'Review the setup and save when the rates are ready.' }}</p></div></div></aside>
      </div>
    </div>
  </div>
</template>
