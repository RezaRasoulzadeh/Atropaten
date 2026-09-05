<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
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
  Search,
  SlidersHorizontal,
  Trash2,
  X,
} from 'lucide-vue-next'
import SectionPanel from '../components/SectionPanel.vue'
import StatusBadge from '../components/StatusBadge.vue'
import WorkspaceStickyStack from '../components/WorkspaceStickyStack.vue'
import { materialsApi, type MaterialRecord } from '../api/materials'
import { servicesApi, type ServiceRecord } from '../api/services'
import { formatDateTime } from '../utils/date'

const emit = defineEmits<{ notify: [message: string] }>()

type ServiceFilter = 'Active' | 'Archived' | 'All'
type EditorMode = 'create' | 'edit' | null
type ParameterType = 'integer' | 'decimal' | 'boolean' | 'choice' | 'material-reference'
type ParameterForm = {
  id: string
  key: string
  label: string
  type: ParameterType
  required: boolean
  defaultValue: string
  options: string[]
  minValue: string | null
  maxValue: string | null
  unit: string
}
type ServiceForm = { name: string; code: string; category: string; description: string; parameters: ParameterForm[] }

const services = ref<ServiceRecord[]>([])
const materials = ref<MaterialRecord[]>([])
const selectedId = ref<string | null>(null)
const searchQuery = ref('')
const serviceFilter = ref<ServiceFilter>('Active')
const editorMode = ref<EditorMode>(null)
const form = ref<ServiceForm>(emptyForm())
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const formError = ref('')

const selectedService = computed(() => services.value.find((service) => service.id === selectedId.value) ?? null)
const filteredServices = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return services.value.filter((service) => {
    const matchesFilter = serviceFilter.value === 'All' || (serviceFilter.value === 'Active' ? service.active : !service.active)
    const matchesSearch = !query || [service.name, service.code, service.category, service.description].some((value) => value.toLowerCase().includes(query))
    return matchesFilter && matchesSearch
  })
})

onMounted(loadServices)

function emptyForm(): ServiceForm {
  return { name: '', code: '', category: '', description: '', parameters: [] }
}

function emptyParameter(): ParameterForm {
  return { id: `draft-parameter-${Date.now()}-${Math.random().toString(16).slice(2)}`, key: '', label: '', type: 'integer', required: false, defaultValue: '', options: [], minValue: null, maxValue: null, unit: '' }
}

async function loadServices() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const [serviceData, materialData] = await Promise.all([servicesApi.list(true), materialsApi.list(false)])
    services.value = serviceData
    materials.value = materialData
    if (!selectedId.value && services.value.length) selectedId.value = services.value[0].id
  } catch (error) {
    errorMessage.value = errorMessageFrom(error, 'Services could not be loaded.')
  } finally {
    isLoading.value = false
  }
}

function selectService(id: string) {
  selectedId.value = id
  editorMode.value = null
  formError.value = ''
}

function startCreate() {
  editorMode.value = 'create'
  selectedId.value = null
  form.value = emptyForm()
  formError.value = ''
}

function startEdit() {
  const service = selectedService.value
  if (!service) return
  form.value = {
    name: service.name, code: service.code, category: service.category, description: service.description,
    parameters: service.parameters.map((parameter) => ({
      id: parameter.id, key: parameter.key, label: parameter.label, type: parameter.type as ParameterType,
      required: parameter.required, defaultValue: parameter.defaultValue, options: [...parameter.options],
      minValue: parameter.minValue ?? null, maxValue: parameter.maxValue ?? null, unit: parameter.unit,
    })),
  }
  editorMode.value = 'edit'
  formError.value = ''
}

function cancelEditor() {
  editorMode.value = null
  formError.value = ''
}

function addParameter() {
  form.value.parameters.push(emptyParameter())
}

function removeParameter(index: number) {
  form.value.parameters.splice(index, 1)
}

function moveParameter(index: number, direction: -1 | 1) {
  const target = index + direction
  if (target < 0 || target >= form.value.parameters.length) return
  const [parameter] = form.value.parameters.splice(index, 1)
  form.value.parameters.splice(target, 0, parameter)
}

function normalizeParameter(parameter: ParameterForm) {
  if (parameter.type === 'choice') {
    parameter.minValue = null
    parameter.maxValue = null
    parameter.defaultValue = parameter.options.includes(parameter.defaultValue) ? parameter.defaultValue : ''
    return
  }
  parameter.options = []
  if (parameter.type !== 'integer' && parameter.type !== 'decimal') {
    parameter.minValue = null
    parameter.maxValue = null
  }
  if (parameter.type === 'boolean') {
    parameter.defaultValue = parameter.defaultValue === 'true' || parameter.defaultValue === 'false' ? parameter.defaultValue : ''
  } else if (parameter.type === 'material-reference') {
    parameter.defaultValue = materials.value.some((material) => material.id === parameter.defaultValue) ? parameter.defaultValue : ''
  } else if (parameter.type === 'integer' || parameter.type === 'decimal') {
    const decimalDefault = /^\d+(?:\.\d{1,6})?$/.test(parameter.defaultValue)
    const integerDefault = /^\d+$/.test(parameter.defaultValue)
    if ((parameter.type === 'integer' && !integerDefault) || (parameter.type === 'decimal' && !decimalDefault)) parameter.defaultValue = ''
    if (parameter.type === 'integer') {
      if (parameter.minValue && !/^\d+$/.test(parameter.minValue)) parameter.minValue = null
      if (parameter.maxValue && !/^\d+$/.test(parameter.maxValue)) parameter.maxValue = null
    }
  } else {
    parameter.defaultValue = ''
  }
}

function addOption(parameter: ParameterForm) {
  parameter.options.push(`Option ${parameter.options.length + 1}`)
}

function removeOption(parameter: ParameterForm, index: number) {
  const removed = parameter.options.splice(index, 1)[0]
  if (parameter.defaultValue === removed) parameter.defaultValue = ''
}

async function saveService() {
  formError.value = ''
  if (!form.value.name.trim()) {
    formError.value = 'Enter a service name.'
    return
  }
  isSaving.value = true
  const wasEditing = editorMode.value === 'edit'
  try {
    const payload = {
      name: form.value.name,
      code: form.value.code,
      category: form.value.category,
      description: form.value.description,
      parameters: form.value.parameters.map((parameter) => ({ ...parameter })),
    }
    const saved = wasEditing && selectedId.value
      ? await servicesApi.update(selectedId.value, payload)
      : await servicesApi.create(payload)
    const existingIndex = services.value.findIndex((service) => service.id === saved.id)
    if (existingIndex >= 0) services.value.splice(existingIndex, 1, saved)
    else services.value.push(saved)
    selectedId.value = saved.id
    editorMode.value = null
    emit('notify', wasEditing ? 'Service updated.' : 'Service created.')
  } catch (error) {
    formError.value = errorMessageFrom(error, 'Service could not be saved.')
  } finally {
    isSaving.value = false
  }
}

async function setActive(active: boolean) {
  const service = selectedService.value
  if (!service) return
  try {
    const updated = active ? await servicesApi.reactivate(service.id) : await servicesApi.archive(service.id)
    const index = services.value.findIndex((item) => item.id === updated.id)
    if (index >= 0) services.value.splice(index, 1, updated)
    emit('notify', active ? 'Service reactivated.' : 'Service archived.')
  } catch (error) {
    errorMessage.value = errorMessageFrom(error, 'Service status could not be changed.')
  }
}

function typeLabel(type: string) {
  return { integer: 'Integer', decimal: 'Decimal', boolean: 'Boolean', choice: 'Choice', 'material-reference': 'Material reference' }[type] ?? type
}

function dateLabel(value: string) {
  try { return formatDateTime(value) } catch { return 'Unknown date' }
}

function errorMessageFrom(error: unknown, fallback: string): string {
  return error instanceof Error && error.message ? error.message : typeof error === 'string' ? error : fallback
}
</script>

<template>
  <div class="services-view">
    <WorkspaceStickyStack>
      <header class="workspace-heading services-heading">
        <div>
          <p class="eyebrow">Catalog / sellable operations</p>
          <h1>Services</h1>
          <p class="heading-description">Define reusable work with operator-facing parameters for future pricing.</p>
        </div>
        <button class="button button-primary" type="button" @click="startCreate"><Plus class="button-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />New service</button>
      </header>
      <section class="services-filter-bar panel" aria-label="Service filters">
        <label class="services-search"><span class="sr-only">Search services</span><Search :size="16" :stroke-width="1.8" aria-hidden="true" /><input v-model="searchQuery" type="search" placeholder="Search service, code, or category" autocomplete="off" /></label>
        <label class="filter-control services-status-filter"><span>Status</span><span class="select-control"><select v-model="serviceFilter" aria-label="Filter services by status"><option>Active</option><option>Archived</option><option>All</option></select><ChevronDown class="select-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" /></span></label>
        <span class="filter-result">{{ filteredServices.length }} of {{ services.length }} services</span>
      </section>
    </WorkspaceStickyStack>

    <div v-if="errorMessage" class="services-error" role="alert"><span>{{ errorMessage }}</span><button class="icon-button" type="button" aria-label="Dismiss services error" @click="errorMessage = ''"><X :size="15" :stroke-width="1.8" aria-hidden="true" /></button></div>

    <section class="services-layout" aria-label="Services workspace">
      <SectionPanel title="Service register" subtitle="Select a service to inspect its operator parameters." class="services-table-panel">
        <template #action><span class="services-table-count">{{ filteredServices.length }} shown</span></template>
        <div v-if="isLoading" class="services-empty"><SlidersHorizontal :size="21" :stroke-width="1.8" aria-hidden="true" /><p>Loading services…</p></div>
        <div v-else-if="filteredServices.length" class="table-wrap services-table-wrap">
          <table class="data-table services-table"><thead><tr><th scope="col">Service</th><th scope="col">Category</th><th scope="col">Parameters</th><th scope="col">Status</th><th scope="col">Updated</th></tr></thead>
            <tbody><tr v-for="service in filteredServices" :key="service.id" :class="{ 'is-selected': selectedId === service.id }" tabindex="0" @click="selectService(service.id)" @keydown.enter="selectService(service.id)"><td><span class="table-primary">{{ service.name }}</span><span class="table-secondary">{{ service.code || 'No code' }}</span></td><td>{{ service.category || '—' }}</td><td><span class="table-primary">{{ service.parameters.length }} {{ service.parameters.length === 1 ? 'parameter' : 'parameters' }}</span><span class="table-secondary">{{ service.parameters.filter((parameter) => parameter.required).length }} required</span></td><td><StatusBadge :label="service.active ? 'Active' : 'Archived'" :tone="service.active ? 'green' : 'slate'" /></td><td>{{ dateLabel(service.updatedAt) }}</td></tr></tbody>
          </table>
        </div>
        <div v-else class="services-empty services-empty-register"><div class="empty-workspace-icon" aria-hidden="true"><SlidersHorizontal :size="21" :stroke-width="1.8" /></div><h2>{{ services.length ? 'No services match this view' : 'No services yet' }}</h2><p>{{ services.length ? 'Try another status or search term.' : 'Create the first reusable operation for your shop.' }}</p><button v-if="!services.length" class="button button-secondary" type="button" @click="startCreate"><Plus class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Create service</button></div>
      </SectionPanel>

      <SectionPanel v-if="editorMode" :title="editorMode === 'create' ? 'New service' : 'Edit service'" subtitle="The full definition saves atomically with its parameters." class="service-inspector service-editor">
        <template #action><button class="icon-button" type="button" aria-label="Close service editor" @click="cancelEditor"><X :size="16" :stroke-width="1.8" aria-hidden="true" /></button></template>
        <form class="service-form" @submit.prevent="saveService">
          <div v-if="formError" class="form-error" role="alert">{{ formError }}</div>
          <label class="form-field form-field-wide"><span>Name</span><input v-model="form.name" type="text" placeholder="Digital Print" autocomplete="off" /></label>
          <div class="service-form-grid"><label class="form-field"><span>Code</span><input v-model="form.code" type="text" placeholder="PRINT" autocomplete="off" /></label><label class="form-field"><span>Category</span><input v-model="form.category" type="text" placeholder="Production" autocomplete="off" /></label></div>
          <label class="form-field form-field-wide"><span>Description / notes</span><textarea v-model="form.description" rows="2" placeholder="What this operation covers"></textarea></label>
          <div class="parameter-editor-heading"><div><h3>Parameters</h3><p>Order is saved as shown and keys are stable references.</p></div><button class="button button-secondary button-compact" type="button" @click="addParameter"><ListPlus class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Add parameter</button></div>
          <div v-if="form.parameters.length" class="parameter-editor-list">
            <article v-for="(parameter, index) in form.parameters" :key="parameter.id" class="parameter-editor-card">
              <header class="parameter-card-header"><span class="parameter-number">{{ String(index + 1).padStart(2, '0') }}</span><strong>{{ parameter.label || 'Untitled parameter' }}</strong><div class="parameter-order-actions"><button class="icon-button icon-button-small" type="button" :disabled="index === 0" :aria-label="`Move ${parameter.label || 'parameter'} up`" @click="moveParameter(index, -1)"><ChevronUp :size="14" :stroke-width="1.8" aria-hidden="true" /></button><button class="icon-button icon-button-small" type="button" :disabled="index === form.parameters.length - 1" :aria-label="`Move ${parameter.label || 'parameter'} down`" @click="moveParameter(index, 1)"><ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" /></button><button class="icon-button icon-button-small danger-icon" type="button" :aria-label="`Remove ${parameter.label || 'parameter'}`" @click="removeParameter(index)"><Trash2 :size="14" :stroke-width="1.8" aria-hidden="true" /></button></div></header>
              <div class="service-form-grid"><label class="form-field"><span>Key</span><input v-model="parameter.key" type="text" placeholder="paper_size" autocomplete="off" /></label><label class="form-field"><span>Label</span><input v-model="parameter.label" type="text" placeholder="Paper size" autocomplete="off" /></label></div>
              <div class="service-form-grid parameter-type-row"><label class="form-field"><span>Type</span><select v-model="parameter.type" @change="normalizeParameter(parameter)"><option value="integer">Integer</option><option value="decimal">Decimal</option><option value="boolean">Boolean</option><option value="choice">Choice</option><option value="material-reference">Material reference</option></select></label><label class="form-field parameter-required"><span>Required</span><span class="checkbox-control"><input v-model="parameter.required" type="checkbox" />Required input</span></label></div>
              <div v-if="parameter.type === 'integer' || parameter.type === 'decimal'" class="service-form-grid"><label class="form-field"><span>Minimum</span><input v-model="parameter.minValue" type="text" inputmode="decimal" placeholder="Optional" /></label><label class="form-field"><span>Maximum</span><input v-model="parameter.maxValue" type="text" inputmode="decimal" placeholder="Optional" /></label></div>
              <label v-if="parameter.type === 'integer' || parameter.type === 'decimal'" class="form-field form-field-wide"><span>Default value</span><input v-model="parameter.defaultValue" type="text" inputmode="decimal" placeholder="Optional" /></label>
              <label v-else-if="parameter.type === 'boolean'" class="form-field form-field-wide"><span>Default value</span><select v-model="parameter.defaultValue"><option value="">Not set</option><option value="true">True</option><option value="false">False</option></select></label>
              <label v-else-if="parameter.type === 'choice'" class="form-field form-field-wide"><span>Default choice</span><select v-model="parameter.defaultValue"><option value="">Not set</option><option v-for="option in parameter.options" :key="option" :value="option">{{ option }}</option></select></label>
              <label v-else class="form-field form-field-wide"><span>Default material <em>optional</em></span><select v-model="parameter.defaultValue"><option value="">No default material</option><option v-for="material in materials" :key="material.id" :value="material.id">{{ material.name }}{{ material.sku ? ` · ${material.sku}` : '' }}</option></select><small class="field-help">Operators may select an active material later; consumption is not configured here.</small></label>
              <div v-if="parameter.type === 'choice'" class="choice-options"><div class="choice-options-heading"><span>Choice options</span><button class="text-button" type="button" @click="addOption(parameter)"><Plus :size="14" :stroke-width="1.8" aria-hidden="true" />Add option</button></div><div v-for="(option, optionIndex) in parameter.options" :key="`${parameter.id}-${optionIndex}`" class="choice-option-row"><input v-model="parameter.options[optionIndex]" type="text" :aria-label="`Choice option ${optionIndex + 1}`" placeholder="A4" @input="parameter.defaultValue = parameter.defaultValue === option ? parameter.options[optionIndex] : parameter.defaultValue" /><button class="icon-button icon-button-small" type="button" :aria-label="`Remove choice option ${optionIndex + 1}`" @click="removeOption(parameter, optionIndex)"><Trash2 :size="13" :stroke-width="1.8" aria-hidden="true" /></button></div><p v-if="!parameter.options.length" class="field-help">Add at least one option before saving.</p></div>
              <label class="form-field form-field-wide"><span>Unit / suffix <em>optional</em></span><input v-model="parameter.unit" type="text" placeholder="sheets, hours, or mm" /></label>
            </article>
          </div>
          <div v-else class="parameter-empty-inline"><ListPlus :size="17" :stroke-width="1.8" aria-hidden="true" /><span>No parameters yet. Add one when the service needs operator input.</span></div>
          <div class="service-form-actions"><button class="button button-secondary" type="button" @click="cancelEditor">Cancel</button><button class="button button-primary" type="submit" :disabled="isSaving"><Save class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />{{ isSaving ? 'Saving…' : 'Save service' }}</button></div>
        </form>
      </SectionPanel>

      <SectionPanel v-else-if="selectedService" title="Service inspector" subtitle="Current persisted definition" class="service-inspector">
        <template #action><button class="icon-button" type="button" aria-label="Edit selected service" @click="startEdit"><Edit3 :size="15" :stroke-width="1.8" aria-hidden="true" /></button></template>
        <div class="inspector-status"><StatusBadge :label="selectedService.active ? 'Active' : 'Archived'" :tone="selectedService.active ? 'green' : 'slate'" /><span class="parameter-summary">{{ selectedService.parameters.length }} parameters</span></div>
        <div class="service-inspector-heading"><div class="service-inspector-icon"><SlidersHorizontal :size="19" :stroke-width="1.8" aria-hidden="true" /></div><div><h3>{{ selectedService.name }}</h3><p>{{ selectedService.code || 'No code' }}<span v-if="selectedService.category"> · {{ selectedService.category }}</span></p></div></div>
        <p v-if="selectedService.description" class="service-description">{{ selectedService.description }}</p>
        <div class="inspector-parameter-list"><div v-for="parameter in selectedService.parameters" :key="parameter.id" class="inspector-parameter"><span class="parameter-position">{{ String(parameter.position + 1).padStart(2, '0') }}</span><div class="inspector-parameter-copy"><strong>{{ parameter.label }}</strong><small><code>{{ parameter.key }}</code> · {{ typeLabel(parameter.type) }}<span v-if="parameter.required"> · required</span></small><small v-if="parameter.type === 'choice'">{{ parameter.options.join(' / ') }}</small><small v-if="parameter.type === 'material-reference'">Active material selection</small><small v-if="parameter.defaultValue">Default: {{ parameter.type === 'material-reference' ? (materials.find((material) => material.id === parameter.defaultValue)?.name || parameter.defaultValue) : parameter.defaultValue }}</small></div></div><p v-if="!selectedService.parameters.length" class="inspector-no-parameters">No parameters configured.</p></div>
        <div class="inspector-meta">Updated {{ dateLabel(selectedService.updatedAt) }}</div>
        <div class="inspector-actions"><button v-if="selectedService.active" class="button button-secondary" type="button" @click="setActive(false)"><Archive class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Archive</button><button v-else class="button button-secondary" type="button" @click="setActive(true)"><RotateCcw class="button-icon" :size="15" :stroke-width="1.8" aria-hidden="true" />Reactivate</button></div>
      </SectionPanel>

      <SectionPanel v-else title="Service inspector" subtitle="Select a row to inspect it." class="service-inspector service-inspector-empty"><div class="inspector-empty-copy"><SlidersHorizontal :size="20" :stroke-width="1.8" aria-hidden="true" /><p>Service details will appear here.</p><button class="text-button" type="button" @click="startCreate">Create a service <Plus :size="14" :stroke-width="1.8" aria-hidden="true" /></button></div></SectionPanel>
    </section>
  </div>
</template>
