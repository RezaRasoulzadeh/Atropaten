<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch, type Component } from 'vue'
import {
  AlertTriangle,
  Check,
  ChevronRight,
  Download,
  FileJson,
  FolderOpen,
  GripVertical,
  Hash,
  Layers3,
  Plus,
  Ruler,
  Save,
  Settings2,
  SlidersHorizontal,
  Sparkles,
  Square,
  Trash2,
  X,
} from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { MaterialRecord } from '../../api/materials'
import type { ParameterForm, ParameterTemplateSeed, ParameterType } from './types'

const props = defineProps<{
  parameters: ParameterForm[]
  category: string
  defaultUnit: string
  materials: MaterialRecord[]
  templateScope: string
  showErrors?: boolean
}>()

const selectedIndex = ref(0)
const emit = defineEmits<{
  applyTemplate: [parameters: ParameterTemplateSeed[]]
}>()

type TemplateSource = 'built-in' | 'draft' | 'custom'

type TemplateOption = {
  id: string
  title: string
  description: string
  icon?: Component
  parameters: ParameterTemplateSeed[]
  source: TemplateSource
}

type StoredTemplate = Omit<TemplateOption, 'icon' | 'source'> & {
  source: Exclude<TemplateSource, 'built-in'>
  updatedAt: string
}

const customTemplates = ref<StoredTemplate[]>([])
const draftTemplate = ref<StoredTemplate | null>(null)
const chooserOpen = ref(false)
const replacementOpen = ref(false)
const pendingTemplate = ref<TemplateOption | null>(null)
const templateName = ref('')
const importInput = ref<HTMLInputElement | null>(null)
const importError = ref('')
const templateDialogOpen = computed(() => chooserOpen.value || replacementOpen.value)

const customTemplatesKey = 'atropaten:service-parameter-templates'
const draftTemplateKey = computed(() => `atropaten:service-parameter-draft:${props.templateScope || 'new-service'}`)

function cloneParameters(parameters: ParameterTemplateSeed[] | ParameterForm[]): ParameterTemplateSeed[] {
  return parameters.map((parameter) => ({
    key: parameter.key,
    label: parameter.label,
    type: parameter.type,
    required: parameter.required,
    defaultValue: parameter.defaultValue,
    options: [...parameter.options],
    minValue: parameter.minValue,
    maxValue: parameter.maxValue,
    unit: parameter.unit,
  }))
}

function readStoredTemplates() {
  if (typeof window === 'undefined') return
  try {
    const saved = JSON.parse(window.localStorage.getItem(customTemplatesKey) || '[]')
    customTemplates.value = Array.isArray(saved) ? saved.filter((item) => item && Array.isArray(item.parameters)) : []
    const draft = JSON.parse(window.localStorage.getItem(draftTemplateKey.value) || 'null')
    draftTemplate.value = draft && Array.isArray(draft.parameters) ? draft : null
  } catch {
    customTemplates.value = []
    draftTemplate.value = null
  }
}

function persistCustomTemplates() {
  if (typeof window !== 'undefined') window.localStorage.setItem(customTemplatesKey, JSON.stringify(customTemplates.value))
}

function persistDraftTemplate() {
  if (typeof window !== 'undefined' && draftTemplate.value) window.localStorage.setItem(draftTemplateKey.value, JSON.stringify(draftTemplate.value))
}

watch(() => props.templateScope, readStoredTemplates, { immediate: true })

watch(
  templateDialogOpen,
  (open) => {
    if (typeof document !== 'undefined') document.body.classList.toggle('service-template-dialog-open', open)
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  if (typeof document !== 'undefined') document.body.classList.remove('service-template-dialog-open')
})

const activeParameter = computed(() => props.parameters[selectedIndex.value] || null)
const activeMaterials = computed(() => props.materials.filter((material) => material.active))

watch(
  () => props.parameters.length,
  (length) => {
    if (!length) selectedIndex.value = 0
    else if (selectedIndex.value >= length) selectedIndex.value = length - 1
  },
)

function id() {
  return `draft-parameter-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function blankParameter(type: ParameterType = 'choice'): ParameterForm {
  return {
    id: id(),
    key: '',
    label: '',
    type,
    required: false,
    defaultValue: '',
    options: type === 'choice' ? ['Option 1'] : [],
    minValue: null,
    maxValue: null,
    unit: '',
  }
}

function addParameter() {
  props.parameters.push(blankParameter())
  selectedIndex.value = props.parameters.length - 1
}

function removeParameter(index: number) {
  props.parameters.splice(index, 1)
  if (selectedIndex.value >= props.parameters.length) selectedIndex.value = Math.max(0, props.parameters.length - 1)
}

function moveParameter(index: number, direction: -1 | 1) {
  const target = index + direction
  if (target < 0 || target >= props.parameters.length) return
  const [parameter] = props.parameters.splice(index, 1)
  props.parameters.splice(target, 0, parameter)
  selectedIndex.value = target
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
    parameter.defaultValue = activeMaterials.value.some((material) => material.id === parameter.defaultValue) ? parameter.defaultValue : ''
  } else if (parameter.type === 'integer' || parameter.type === 'decimal') {
    const valid = parameter.type === 'integer' ? /^\d+$/.test(parameter.defaultValue) : /^\d+(?:\.\d{1,6})?$/.test(parameter.defaultValue)
    if (parameter.defaultValue && !valid) parameter.defaultValue = ''
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

function typeLabel(type: ParameterType) {
  return {
    integer: 'Number',
    decimal: 'Decimal',
    boolean: 'Yes / no',
    choice: 'Choice',
    'material-reference': 'Material',
  }[type]
}

function typeIcon(type: ParameterType) {
  return {
    integer: Hash,
    decimal: Ruler,
    boolean: Check,
    choice: Layers3,
    'material-reference': Square,
  }[type]
}

function parameterSummary(parameter: ParameterForm) {
  if (parameter.type === 'integer' || parameter.type === 'decimal') {
    const range = [parameter.minValue, parameter.maxValue].filter(Boolean).join(' – ')
    return `${typeLabel(parameter.type)}${range ? ` · ${range}` : ''}${parameter.unit ? ` ${parameter.unit}` : ''}`
  }
  if (parameter.type === 'material-reference') return `Material · ${activeMaterials.value.length} available`
  if (parameter.type === 'choice') return `Choice · ${parameter.options.length} option${parameter.options.length === 1 ? '' : 's'}`
  return `${typeLabel(parameter.type)} · ${parameter.required ? 'Required' : 'Optional'}`
}

function seed(label: string, key: string, type: ParameterType, values: Partial<ParameterTemplateSeed> = {}): ParameterTemplateSeed {
  return {
    key,
    label,
    type,
    required: false,
    defaultValue: '',
    options: [],
    minValue: null,
    maxValue: null,
    unit: '',
    ...values,
  }
}

const templates = computed(() => {
  const category = props.category.toLowerCase()

  const paperPrintingParameters = (businessCard = false): ParameterTemplateSeed[] => [
    seed('Quantity', 'quantity', 'integer', { required: true, defaultValue: '100', minValue: '1', maxValue: '50000', unit: props.defaultUnit }),
    seed('Paper type', 'paper_type', 'choice', {
      required: true,
      defaultValue: businessCard ? '350gsm silk coated' : '120gsm uncoated',
      options: businessCard
        ? ['250gsm matte coated', '300gsm matte coated', '350gsm silk coated', '400gsm premium coated', '300gsm kraft', '300gsm recycled']
        : ['80gsm uncoated', '120gsm uncoated', '170gsm matte coated', '250gsm matte coated', '300gsm matte coated', '350gsm silk coated', '300gsm kraft', '300gsm recycled'],
    }),
    seed('Size', 'size', 'choice', {
      required: true,
      defaultValue: businessCard ? '90 × 50 mm (Standard)' : 'A4 (210 × 297 mm)',
      options: businessCard
        ? ['90 × 50 mm (Standard)', '85 × 55 mm (Standard)', '90 × 55 mm (European)', 'Custom']
        : ['A6 (105 × 148 mm)', 'A5 (148 × 210 mm)', 'A4 (210 × 297 mm)', 'A3 (297 × 420 mm)', 'A2 (420 × 594 mm)', 'Custom'],
    }),
    seed('Color', 'color', 'choice', { required: true, defaultValue: 'Full color (2 sides)', options: ['Full color (2 sides)', 'Full color (1 side)', 'Black & white (2 sides)', 'Black & white (1 side)', 'Spot color'] }),
    seed('Finishing', 'finishing', 'choice', { required: true, defaultValue: 'Standard trim', options: ['Standard trim', 'Crease', 'Fold', 'Die cut', 'UV coating', 'Foil stamping', 'Embossing'] }),
    seed('Corners', 'corners', 'choice', { defaultValue: 'Square', options: ['Square', 'Rounded 2 corners', 'Rounded 4 corners', 'Custom die cut'] }),
    seed('Lamination', 'lamination', 'choice', { defaultValue: 'None', options: ['None', 'Matte', 'Gloss', 'Soft touch', 'Anti-scratch'] }),
    seed('Turnaround', 'turnaround', 'choice', { required: true, defaultValue: 'Standard (3–5 business days)', options: ['Standard (3–5 business days)', 'Express (1–2 business days)', 'Same day (subject to approval)', 'Scheduled date'] }),
  ]

  const largeFormatParameters = (banner = false): ParameterTemplateSeed[] => [
    seed('Quantity', 'quantity', 'integer', { required: true, defaultValue: '1', minValue: '1', maxValue: '1000', unit: props.defaultUnit }),
    seed('Finished width', 'finished_width', 'decimal', { required: true, defaultValue: banner ? '200' : '100', minValue: '1', unit: banner ? 'cm' : 'cm' }),
    seed('Finished height', 'finished_height', 'decimal', { required: true, defaultValue: banner ? '100' : '140', minValue: '1', unit: banner ? 'cm' : 'cm' }),
    seed('Print unit', 'print_unit', 'choice', { required: true, defaultValue: 'cm', options: ['mm', 'cm', 'm', 'inch', 'ft'] }),
    seed(banner ? 'Banner material' : 'Substrate', 'substrate', 'choice', {
      required: true,
      defaultValue: banner ? '13oz matte vinyl' : '5mm foam board',
      options: banner
        ? ['13oz matte vinyl', '15oz blockout vinyl', '18oz heavy-duty vinyl', 'Mesh vinyl', 'Fabric banner']
        : ['Self-adhesive vinyl', '5mm foam board', '3mm PVC board', 'Acrylic', 'Aluminum composite', 'Fabric', 'Canvas'],
    }),
    seed('Color', 'color', 'choice', { required: true, defaultValue: 'Full color (1 side)', options: ['Full color (1 side)', 'Full color (2 sides)', 'Black & white (1 side)'] }),
    seed('Finishing', 'finishing', 'choice', { required: true, defaultValue: banner ? 'Hemmed edges' : 'Trim to size', options: banner ? ['Raw cut', 'Hemmed edges', 'Hemmed + grommets', 'Pole pockets', 'Reinforced corners'] : ['Trim to size', 'Contour cut', 'Mounted', 'Laminated', 'Folded'] }),
    ...(banner ? [
      seed('Grommets', 'grommets', 'choice', { defaultValue: 'Standard every 60 cm', options: ['None', 'Corners only', 'Standard every 60 cm', 'Every 30 cm', 'Custom placement'] }),
      seed('Pole pockets', 'pole_pockets', 'choice', { defaultValue: 'None', options: ['None', 'Top', 'Bottom', 'Top and bottom', 'Left and right'] }),
      seed('Wind slits', 'wind_slits', 'choice', { defaultValue: 'None', options: ['None', 'Corners only', 'Every 60 cm'] }),
    ] : []),
    seed('Turnaround', 'turnaround', 'choice', { required: true, defaultValue: 'Standard (3–5 business days)', options: ['Standard (3–5 business days)', 'Express (1–2 business days)', 'Same day (subject to approval)', 'Scheduled date'] }),
  ]

  const brochureParameters: ParameterTemplateSeed[] = [
    ...paperPrintingParameters(false),
    seed('Pages', 'pages', 'integer', { required: true, defaultValue: '4', minValue: '2', maxValue: '256', unit: 'pages' }),
    seed('Folding', 'folding', 'choice', { defaultValue: 'None', options: ['None', 'Half fold', 'Tri-fold', 'Z-fold', 'Gate fold', 'Accordion fold'] }),
    seed('Binding', 'binding', 'choice', { defaultValue: 'None', options: ['None', 'Saddle stitch', 'Perfect bound', 'Spiral bound', 'Wire-o bound'] }),
  ]

  const labelParameters: ParameterTemplateSeed[] = [
    seed('Quantity', 'quantity', 'integer', { required: true, defaultValue: '100', minValue: '1', maxValue: '100000', unit: props.defaultUnit }),
    seed('Size', 'size', 'choice', { required: true, defaultValue: '100 × 50 mm', options: ['50 × 30 mm', '100 × 50 mm', 'A6', 'A5', 'Custom'] }),
    seed('Stock', 'stock', 'choice', { required: true, defaultValue: 'White adhesive paper', options: ['White adhesive paper', 'Clear adhesive film', 'Kraft adhesive paper', 'White waterproof vinyl', 'Silver adhesive film'] }),
    seed('Color', 'color', 'choice', { required: true, defaultValue: 'Full color', options: ['Full color', 'Black & white', 'Spot color'] }),
    seed('Shape', 'shape', 'choice', { required: true, defaultValue: 'Rectangle', options: ['Rectangle', 'Circle', 'Oval', 'Rounded rectangle', 'Custom die cut'] }),
    seed('Finish', 'finish', 'choice', { defaultValue: 'Matte', options: ['Matte', 'Gloss', 'Clear', 'Soft touch'] }),
    seed('Turnaround', 'turnaround', 'choice', { required: true, defaultValue: 'Standard (3–5 business days)', options: ['Standard (3–5 business days)', 'Express (1–2 business days)', 'Same day (subject to approval)'] }),
  ]

  const posterParameters: ParameterTemplateSeed[] = [
    ...paperPrintingParameters(false),
    seed('Orientation', 'orientation', 'choice', { defaultValue: 'Portrait', options: ['Portrait', 'Landscape', 'Square'] }),
    seed('Mounting', 'mounting', 'choice', { defaultValue: 'Unmounted', options: ['Unmounted', 'Foam board', 'PVC board', 'Acrylic', 'Aluminum composite'] }),
  ]

  const envelopeParameters: ParameterTemplateSeed[] = [
    seed('Quantity', 'quantity', 'integer', { required: true, defaultValue: '100', minValue: '1', maxValue: '100000', unit: props.defaultUnit }),
    seed('Envelope size', 'envelope_size', 'choice', { required: true, defaultValue: 'DL (110 × 220 mm)', options: ['C6 (114 × 162 mm)', 'DL (110 × 220 mm)', 'C5 (162 × 229 mm)', 'C4 (229 × 324 mm)', 'Custom'] }),
    seed('Paper stock', 'paper_stock', 'choice', { required: true, defaultValue: '120gsm white', options: ['90gsm white', '120gsm white', '120gsm recycled', 'Kraft', 'Color paper'] }),
    seed('Color', 'color', 'choice', { required: true, defaultValue: 'Full color (1 side)', options: ['Full color (1 side)', 'Full color (2 sides)', 'Black & white (1 side)', 'No printing'] }),
    seed('Window', 'window', 'choice', { defaultValue: 'No window', options: ['No window', 'Left window', 'Right window', 'Custom window'] }),
    seed('Closure', 'closure', 'choice', { defaultValue: 'Gummed flap', options: ['Gummed flap', 'Self-adhesive strip', 'Peel and seal'] }),
    seed('Turnaround', 'turnaround', 'choice', { required: true, defaultValue: 'Standard (3–5 business days)', options: ['Standard (3–5 business days)', 'Express (1–2 business days)'] }),
  ]

  const textileParameters: ParameterTemplateSeed[] = [
    seed('Quantity', 'quantity', 'integer', { required: true, defaultValue: '10', minValue: '1', maxValue: '10000', unit: props.defaultUnit }),
    seed('Product', 'product', 'choice', { required: true, defaultValue: 'T-shirt', options: ['T-shirt', 'Hoodie', 'Polo shirt', 'Tote bag', 'Cap', 'Other garment'] }),
    seed('Print area', 'print_area', 'choice', { required: true, defaultValue: 'Front', options: ['Front', 'Back', 'Front and back', 'Sleeve', 'Custom'] }),
    seed('Print method', 'print_method', 'choice', { required: true, defaultValue: 'DTF', options: ['DTF', 'DTG', 'Screen printing', 'Sublimation', 'Heat transfer', 'Embroidery'] }),
    seed('Color', 'color', 'choice', { required: true, defaultValue: 'Full color', options: ['Full color', 'Black & white', 'Spot colors'] }),
    seed('Finishing', 'finishing', 'choice', { defaultValue: 'Standard', options: ['Standard', 'Fold and pack', 'Individual packaging', 'Tag removal and replacement'] }),
    seed('Turnaround', 'turnaround', 'choice', { required: true, defaultValue: 'Standard (3–5 business days)', options: ['Standard (3–5 business days)', 'Express (1–2 business days)'] }),
  ]

  if (category.includes('banner')) {
    return [{ id: 'banner', title: 'Banner printing', description: 'Prebuilt choices for vinyl and fabric banners, including mounting and outdoor finishing.', icon: Square, parameters: largeFormatParameters(true) }]
  }
  if (category.includes('large') || category.includes('wide') || category.includes('sign')) {
    return [{ id: 'large-format', title: 'Large format printing', description: 'A flexible setup for posters, boards, displays, wall graphics, and oversized prints.', icon: Ruler, parameters: largeFormatParameters() }]
  }
  if (category.includes('brochure') || category.includes('booklet') || category.includes('flyer')) {
    return [{ id: 'brochure', title: 'Flyer and brochure printing', description: 'Paper stock plus folding, page count, and binding options for handouts and booklets.', icon: Layers3, parameters: brochureParameters }]
  }
  if (category.includes('label') || category.includes('sticker')) {
    return [{ id: 'labels', title: 'Labels and stickers', description: 'A ready setup for adhesive products with stock, shape, color, and finish choices.', icon: Square, parameters: labelParameters }]
  }
  if (category.includes('envelope')) {
    return [{ id: 'envelopes', title: 'Envelope printing', description: 'Common envelope sizes, paper stocks, windows, closures, and print coverage.', icon: Square, parameters: envelopeParameters }]
  }
  if (category.includes('textile') || category.includes('apparel') || category.includes('garment')) {
    return [{ id: 'textile', title: 'Textile and apparel printing', description: 'A flexible setup for garments, print areas, decoration methods, and fulfillment.', icon: Layers3, parameters: textileParameters }]
  }
  if (category.includes('poster')) {
    return [{ id: 'posters', title: 'Poster printing', description: 'Sheet printing with poster orientation and mounting choices for displays.', icon: Ruler, parameters: posterParameters }]
  }
  if (category.includes('card') || category.includes('stationery')) {
    return [{ id: 'business-cards', title: 'Cards and stationery', description: 'A focused setup for cards and small-format stationery with stock, size, and finishing choices.', icon: Square, parameters: paperPrintingParameters(true) }]
  }
  if (category.includes('print') || category.includes('product') || category.includes('paper')) {
    return [
      {
        id: 'paper-printing',
        title: 'Paper printing',
        description: 'The common parameter set for flyers, leaflets, posters, documents, and other sheet products.',
        icon: Sparkles,
        parameters: paperPrintingParameters(false),
      },
      {
        id: 'business-cards',
        title: 'Cards and stationery',
        description: 'A focused small-format setup for business cards, postcards, and stationery.',
        icon: Square,
        parameters: paperPrintingParameters(true),
      },
      {
        id: 'large-format',
        title: 'Large format',
        description: 'Dimensions and substrate choices for posters, boards, displays, and oversized prints.',
        icon: Ruler,
        parameters: largeFormatParameters(),
      },
      {
        id: 'banner',
        title: 'Banners and signage',
        description: 'Outdoor-ready dimensions and installation choices such as hems, grommets, and pole pockets.',
        icon: Square,
        parameters: largeFormatParameters(true),
      },
      {
        id: 'posters',
        title: 'Poster printing',
        description: 'Sheet printing with mounting and orientation choices for posters and displays.',
        icon: Ruler,
        parameters: posterParameters,
      },
      {
        id: 'envelopes',
        title: 'Envelope printing',
        description: 'Envelope sizes, stock, windows, closure types, and print coverage.',
        icon: Square,
        parameters: envelopeParameters,
      },
      {
        id: 'textile',
        title: 'Textile printing',
        description: 'Garments, print areas, decoration methods, and fulfillment options.',
        icon: Layers3,
        parameters: textileParameters,
      },
    ]
  }
  if (category.includes('finish')) {
    return [{
      id: 'finishing',
      title: 'Finishing service',
      description: 'Capture the quantity, finish choice, and delivery speed for finishing work.',
      icon: Layers3,
      parameters: [
        seed('Quantity', 'quantity', 'integer', { required: true, defaultValue: '1', minValue: '1', unit: props.defaultUnit }),
        seed('Finish type', 'finish_type', 'choice', { required: true, defaultValue: 'Matte', options: ['Matte', 'Glossy', 'Soft touch', 'None'] }),
        seed('Turnaround', 'turnaround', 'choice', { required: true, defaultValue: 'Standard', options: ['Standard', 'Express'] }),
      ],
    }]
  }
  if (category.includes('pack')) {
    return [{
      id: 'packaging',
      title: 'Packaging service',
      description: 'Start with quantity, stock, size, and delivery inputs for packaging work.',
      icon: Square,
      parameters: [
        seed('Quantity', 'quantity', 'integer', { required: true, defaultValue: '1', minValue: '1', unit: props.defaultUnit }),
        seed('Material type', 'material_type', 'choice', { required: true, defaultValue: 'Kraft board', options: ['Kraft board', 'Coated board', 'Recycled board', 'Corrugated board'] }),
        seed('Size', 'size', 'choice', { required: true, defaultValue: 'Standard', options: ['Small', 'Standard', 'Large', 'Custom'] }),
      ],
    }]
  }
  return [{
    id: 'simple-service',
    title: 'Simple service',
    description: 'A small starting point for any service. Add more inputs whenever you need them.',
    icon: SlidersHorizontal,
    parameters: [
      seed('Quantity', 'quantity', 'integer', { required: true, defaultValue: '1', minValue: '1', unit: props.defaultUnit }),
      seed('Option', 'option', 'choice', { options: ['Standard', 'Express'], defaultValue: 'Standard' }),
    ],
  }]
})

const availableTemplates = computed<TemplateOption[]>(() => [
  ...(draftTemplate.value ? [{ ...draftTemplate.value, source: 'draft' as const, icon: FileJson }] : []),
  ...templates.value.map((template) => ({ ...template, source: 'built-in' as const })),
  ...customTemplates.value.map((template) => ({ ...template, source: 'custom' as const, icon: FileJson })),
])

function saveCurrentAsDraft() {
  if (!props.parameters.length) return
  draftTemplate.value = {
    id: `draft-${props.templateScope || 'new-service'}`,
    title: 'Temporary draft',
    description: 'Your previous parameter setup for this service.',
    parameters: cloneParameters(props.parameters),
    source: 'draft',
    updatedAt: new Date().toISOString(),
  }
  persistDraftTemplate()
}

function saveCurrentAsCustom(name = templateName.value) {
  if (!props.parameters.length) return
  const trimmedName = name.trim() || 'My service template'
  const template: StoredTemplate = {
    id: `custom-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    title: trimmedName,
    description: 'A reusable parameter setup created in this service editor.',
    parameters: cloneParameters(props.parameters),
    source: 'custom',
    updatedAt: new Date().toISOString(),
  }
  customTemplates.value = [template, ...customTemplates.value]
  persistCustomTemplates()
}

function saveCurrentAsCustomWithPrompt() {
  if (!props.parameters.length || typeof window === 'undefined') return
  const name = window.prompt('Name this reusable template', `${props.category || 'Service'} setup`)
  if (name?.trim()) saveCurrentAsCustom(name)
}

function openChooser() {
  importError.value = ''
  chooserOpen.value = true
}

function openImport() {
  importInput.value?.click()
}

function requestApply(template: TemplateOption) {
  if (!props.parameters.length) {
    emit('applyTemplate', cloneParameters(template.parameters))
    chooserOpen.value = false
    return
  }
  pendingTemplate.value = template
  templateName.value = `${props.category || 'Service'} setup`
  chooserOpen.value = false
  replacementOpen.value = true
}

function replaceWithPending(action: 'replace' | 'draft' | 'custom') {
  if (!pendingTemplate.value) return
  if (action === 'draft') saveCurrentAsDraft()
  if (action === 'custom') saveCurrentAsCustom()
  emit('applyTemplate', cloneParameters(pendingTemplate.value.parameters))
  pendingTemplate.value = null
  replacementOpen.value = false
}

function deleteCustomTemplate(template: TemplateOption) {
  customTemplates.value = customTemplates.value.filter((item) => item.id !== template.id)
  persistCustomTemplates()
}

function exportTemplate(template: TemplateOption) {
  if (typeof window === 'undefined') return
  const payload = {
    format: 'atropaten.service-parameter-template',
    version: 1,
    name: template.title,
    description: template.description,
    parameters: cloneParameters(template.parameters),
  }
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${template.title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') || 'service-template'}.json`
  link.click()
  URL.revokeObjectURL(url)
}

function normalizeImportedParameters(raw: unknown): ParameterTemplateSeed[] {
  if (!Array.isArray(raw)) throw new Error('The template must contain a parameters array.')
  const allowedTypes: ParameterType[] = ['integer', 'decimal', 'boolean', 'choice', 'material-reference']
  const result = raw.map((item, index) => {
    if (!item || typeof item !== 'object') throw new Error(`Parameter ${index + 1} is invalid.`)
    const source = item as Record<string, unknown>
    const type = allowedTypes.includes(source.type as ParameterType) ? source.type as ParameterType : 'choice'
    const label = String(source.label || `Parameter ${index + 1}`).trim()
    const key = String(source.key || label.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '') || `parameter_${index + 1}`)
    const options = Array.isArray(source.options) ? source.options.map(String).map((value) => value.trim()).filter(Boolean) : []
    return seed(label, key, type, {
      required: source.required === true,
      defaultValue: source.defaultValue == null ? '' : String(source.defaultValue),
      options,
      minValue: source.minValue == null || source.minValue === '' ? null : String(source.minValue),
      maxValue: source.maxValue == null || source.maxValue === '' ? null : String(source.maxValue),
      unit: source.unit == null ? '' : String(source.unit),
    })
  })
  if (!result.length) throw new Error('The template does not contain any parameters.')
  return result
}

async function importTemplate(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  try {
    const parsed = JSON.parse(await file.text()) as Record<string, unknown> | unknown[]
    const source = Array.isArray(parsed) ? {} : parsed
    const parameters = normalizeImportedParameters(Array.isArray(parsed) ? parsed : source.parameters)
    const title = String(Array.isArray(parsed) ? file.name.replace(/\.json$/i, '') : source.name || file.name.replace(/\.json$/i, '')).trim() || 'Imported template'
    const imported: StoredTemplate = {
      id: `custom-${Date.now()}-${Math.random().toString(16).slice(2)}`,
      title,
      description: String(Array.isArray(parsed) ? 'Imported from a JSON template.' : source.description || 'Imported from a JSON template.'),
      parameters,
      source: 'custom',
      updatedAt: new Date().toISOString(),
    }
    customTemplates.value = [imported, ...customTemplates.value]
    persistCustomTemplates()
    importError.value = ''
  } catch (error) {
    importError.value = error instanceof Error ? error.message : 'Could not import this JSON template.'
  }
}

function applyTemplate(template: TemplateOption) {
  requestApply(template)
}
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Service parameters">
    <div v-if="!parameters.length" class="space-y-4">
      <div class="rounded-box border border-dashed border-primary/35 bg-primary/5 p-4">
        <div class="flex items-start gap-3">
          <span class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Sparkles :size="18" aria-hidden="true" /></span>
          <div><h3 class="text-sm font-semibold">Start with a service template</h3><p class="mt-1 text-xs leading-5 text-base-content/65">We prepared a starting point from the category <strong class="text-base-content">{{ category || 'you selected' }}</strong>. You can change every input after applying it.</p></div>
        </div>
      </div>
      <div class="grid min-w-0 gap-2.5">
        <article v-for="template in templates" :key="template.id" class="flex min-w-0 flex-col rounded-box border border-base-300 bg-base-100 p-3">
          <div class="flex items-start gap-2.5"><span class="grid size-8 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><component :is="template.icon" :size="17" aria-hidden="true" /></span><div class="min-w-0"><h3 class="truncate text-sm font-semibold">{{ template.title }}</h3><p class="mt-1 line-clamp-2 text-xs leading-4 text-base-content/60">{{ template.description }}</p></div></div>
          <div class="mt-3 flex flex-wrap gap-1"><span v-for="parameter in template.parameters.slice(0, 5)" :key="parameter.key" class="badge badge-ghost px-2 text-[0.68rem]">{{ parameter.label }}</span><span v-if="template.parameters.length > 5" class="badge badge-ghost px-2 text-[0.68rem]">+{{ template.parameters.length - 5 }}</span></div>
          <button class="btn btn-primary btn-xs mt-3 self-start gap-1.5" type="button" @click="applyTemplate({ ...template, source: 'built-in' })"><Check :size="13" aria-hidden="true" />Use template</button>
        </article>
      </div>
      <button class="btn btn-ghost btn-xs" type="button" @click="addParameter">Start with a blank parameter</button>
    </div>

    <div v-else class="grid min-w-0 gap-4 lg:grid-cols-[minmax(13rem,0.78fr)_minmax(0,1.3fr)]">
      <div class="min-w-0 space-y-2">
        <button v-for="(parameter, index) in parameters" :key="parameter.id" class="group flex w-full min-w-0 items-center gap-2 rounded-box border p-3 text-start transition-colors" :class="index === selectedIndex ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-100 hover:border-primary/45'" type="button" @click="selectedIndex = index">
          <GripVertical class="shrink-0 text-base-content/40" :size="16" aria-hidden="true" />
          <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-base-content/75"><component :is="typeIcon(parameter.type)" :size="19" aria-hidden="true" /></span>
          <span class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ parameter.label || 'New parameter' }}</strong><small class="mt-0.5 block truncate text-xs text-base-content/60">{{ parameterSummary(parameter) }}</small></span>
          <ChevronRight class="shrink-0 text-base-content/45" :size="17" aria-hidden="true" />
        </button>
        <div class="grid gap-2 sm:grid-cols-2 lg:grid-cols-1">
          <button class="flex w-full items-center justify-center gap-2 rounded-box border border-dashed border-base-300 px-3 py-3 text-sm text-base-content/65 transition-colors hover:border-primary hover:text-primary" type="button" @click="addParameter"><Plus :size="15" aria-hidden="true" />Add parameter</button>
          <button class="flex w-full items-center justify-center gap-2 rounded-box border border-dashed border-base-300 px-3 py-3 text-sm text-base-content/65 transition-colors hover:border-primary hover:text-primary" type="button" @click="openChooser"><FolderOpen :size="15" aria-hidden="true" />Choose template</button>
        </div>
      </div>

      <div v-if="activeParameter" class="min-w-0 rounded-box border border-base-300 bg-base-100 p-4 sm:p-5">
        <div class="flex min-w-0 flex-wrap items-start justify-between gap-3 border-b border-base-300 pb-4">
          <div class="min-w-0"><h3 class="text-xl font-semibold">Edit parameter</h3><p class="mt-1 text-sm text-base-content/65">Configure how this parameter appears to customers.</p></div>
          <button class="btn btn-outline btn-error btn-sm gap-2" type="button" @click="removeParameter(selectedIndex)"><Trash2 :size="14" aria-hidden="true" />Delete</button>
        </div>
        <div class="mt-5 grid min-w-0 gap-4 sm:grid-cols-2">
          <FormField class="gap-1"><span>Label <em class="text-error">*</em></span><AppInput v-model="activeParameter.label" class="input w-full min-w-0" :class="{ 'input-error': showErrors && !activeParameter.label.trim() }" required placeholder="Quantity, paper type, or color" /></FormField>
          <FormField class="gap-1"><span>Type <em class="text-error">*</em></span><SelectField v-model="activeParameter.type" label="" :options="[
            { label: 'Number', value: 'integer' },
            { label: 'Decimal measurement', value: 'decimal' },
            { label: 'Choose from options', value: 'choice' },
            { label: 'Choose a material or paper', value: 'material-reference' },
            { label: 'Yes / no choice', value: 'boolean' },
          ]" @update:model-value="normalizeParameter(activeParameter)" /></FormField>
        </div>
        <div class="mt-4 flex flex-wrap items-center justify-between gap-3 rounded-box border border-base-300 bg-base-200/35 px-3 py-2.5"><label class="flex items-center gap-2 text-sm"><input v-model="activeParameter.required" class="checkbox" type="checkbox" />Required when ordering</label><span class="text-xs text-base-content/60">Customers must provide a value.</span></div>

        <div v-if="activeParameter.type === 'integer' || activeParameter.type === 'decimal'" class="mt-4 grid min-w-0 gap-4 sm:grid-cols-2">
          <FormField class="gap-1"><span>Unit <em class="text-base-content/45">optional</em></span><AppInput v-model="activeParameter.unit" class="input w-full min-w-0" placeholder="piece, sheet, mm" /></FormField>
          <FormField class="gap-1"><span>Default value <em class="text-base-content/45">optional</em></span><AppInput v-model="activeParameter.defaultValue" class="input w-full min-w-0" inputmode="decimal" placeholder="100" /></FormField>
          <FormField class="gap-1"><span>Minimum value</span><AppInput v-model="activeParameter.minValue" class="input w-full min-w-0" inputmode="decimal" placeholder="No minimum" /></FormField>
          <FormField class="gap-1"><span>Maximum value</span><AppInput v-model="activeParameter.maxValue" class="input w-full min-w-0" inputmode="decimal" placeholder="No maximum" /></FormField>
        </div>
        <div v-else-if="activeParameter.type === 'material-reference'" class="mt-4 rounded-box border border-base-300 bg-base-200/25 p-3"><SelectField v-model="activeParameter.defaultValue" label="Default material" :options="[{ label: 'No default material', value: '' }, ...activeMaterials.map((material) => ({ label: `${material.name}${material.sku ? ` · ${material.sku}` : ''}`, value: material.id }))]" /><p class="mt-2 text-xs leading-5 text-base-content/60">Customers will choose from active materials. Add materials in the Materials view to expand this list.</p></div>
        <div v-else-if="activeParameter.type === 'boolean'" class="mt-4 rounded-box border border-base-300 bg-base-200/25 p-3"><SelectField v-model="activeParameter.defaultValue" label="Default answer" :options="[{ label: 'No default answer', value: '' }, { label: 'Yes', value: 'true' }, { label: 'No', value: 'false' }]" /></div>
        <div v-else class="mt-4 rounded-box border border-base-300 bg-base-200/25 p-3"><div class="flex flex-wrap items-center justify-between gap-2"><div><h4 class="text-sm font-semibold">Options customers can choose</h4><p class="mt-1 text-xs text-base-content/60">Add values such as A4, A5, or Matte.</p></div><button class="btn btn-outline btn-sm gap-2" type="button" @click="addOption(activeParameter)"><Plus :size="14" aria-hidden="true" />Add option</button></div><div v-if="activeParameter.options.length" class="mt-3 grid min-w-0 gap-2 sm:grid-cols-2"><div v-for="(option, optionIndex) in activeParameter.options" :key="`${activeParameter.id}-${optionIndex}`" class="flex min-w-0 items-center gap-2"><AppInput v-model="activeParameter.options[optionIndex]" class="input w-full min-w-0" :class="{ 'input-error': showErrors && !activeParameter.options[optionIndex].trim() }" required :aria-label="`Option ${optionIndex + 1}`" placeholder="A4" /><button class="btn btn-outline btn-error btn-sm shrink-0" type="button" :aria-label="`Remove option ${optionIndex + 1}`" @click="removeOption(activeParameter, optionIndex)"><Trash2 :size="13" aria-hidden="true" /></button></div></div><p v-else class="mt-3 rounded-box border border-dashed border-base-300 p-3 text-sm text-base-content/60">No options yet.</p><SelectField v-if="activeParameter.options.length" v-model="activeParameter.defaultValue" class="mt-3" label="Default option" :options="[{ label: 'No default option', value: '' }, ...activeParameter.options.map((option) => ({ label: option, value: option }))]" /></div>
        <details class="mt-4 rounded-box border border-base-300 px-3 py-2"><summary class="flex cursor-pointer list-none items-center gap-2 text-xs font-semibold text-base-content/70 [&::-webkit-details-marker]:hidden"><Settings2 :size="14" aria-hidden="true" />Advanced settings</summary><FormField class="mt-3 gap-1"><span>Internal key</span><AppInput v-model="activeParameter.key" class="input w-full min-w-0" placeholder="Generated from the label" /><small class="text-xs leading-5 text-base-content/60">Used internally by pricing. Leave it unchanged unless you know why it needs a custom key.</small></FormField></details>
      </div>
    </div>

    <input ref="importInput" class="hidden" type="file" accept="application/json,.json" @change="importTemplate" />

    <div v-if="chooserOpen" class="fixed inset-0 z-50 grid place-items-center bg-black/60 p-4" role="dialog" aria-modal="true" aria-labelledby="template-chooser-title" @click.self="chooserOpen = false">
      <div class="flex max-h-[min(44rem,calc(100vh-2rem))] w-full max-w-5xl min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-200 shadow-2xl">
        <header class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 px-4 py-3 sm:px-5">
          <div class="flex min-w-0 items-center gap-2.5"><span class="grid size-8 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FileJson :size="17" aria-hidden="true" /></span><div class="min-w-0"><h3 id="template-chooser-title" class="truncate text-base font-semibold">Choose a parameter template</h3><p class="text-xs text-base-content/60">Start from a prepared setup, import JSON, or reuse your own.</p></div></div>
          <div class="flex flex-wrap items-center gap-2">
            <button v-if="parameters.length" class="btn btn-ghost btn-xs gap-1.5" type="button" @click="saveCurrentAsDraft"><Save :size="13" aria-hidden="true" />Save temporary draft</button>
            <button v-if="parameters.length" class="btn btn-ghost btn-xs gap-1.5" type="button" @click="saveCurrentAsCustomWithPrompt"><FileJson :size="13" aria-hidden="true" />Save reusable</button>
            <button class="btn btn-outline btn-xs gap-1.5" type="button" @click="openImport"><FolderOpen :size="13" aria-hidden="true" />Import JSON</button>
            <button class="btn btn-ghost btn-square btn-sm" type="button" aria-label="Close template chooser" @click="chooserOpen = false"><X :size="16" aria-hidden="true" /></button>
          </div>
        </header>

        <div class="min-h-0 overflow-y-auto p-4 sm:p-5">
          <p v-if="importError" class="mb-3 flex items-start gap-2 rounded-box border border-error/35 bg-error/10 p-3 text-sm text-error"><AlertTriangle class="mt-0.5 shrink-0" :size="15" aria-hidden="true" />{{ importError }}</p>
          <div class="grid min-w-0 gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
            <article v-for="template in availableTemplates" :key="`${template.source}-${template.id}`" class="flex min-w-0 flex-col rounded-box border border-base-300 bg-base-100 p-3 transition-colors hover:border-primary/60">
              <div class="flex min-w-0 items-start gap-2.5"><span class="grid size-8 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><component :is="template.icon || FileJson" :size="17" aria-hidden="true" /></span><div class="min-w-0"><div class="flex items-center gap-1.5"><h4 class="truncate text-sm font-semibold">{{ template.title }}</h4><span v-if="template.source === 'draft'" class="badge badge-warning badge-xs shrink-0">Draft</span><span v-else-if="template.source === 'custom'" class="badge badge-info badge-xs shrink-0">Saved</span></div><p class="mt-1 line-clamp-2 text-xs leading-4 text-base-content/60">{{ template.description }}</p></div></div>
              <div class="mt-3 flex flex-wrap gap-1"><span v-for="parameter in template.parameters.slice(0, 4)" :key="parameter.key" class="badge badge-ghost px-2 text-[0.68rem]">{{ parameter.label }}</span><span v-if="template.parameters.length > 4" class="badge badge-ghost px-2 text-[0.68rem]">+{{ template.parameters.length - 4 }}</span></div>
              <div class="mt-3 flex items-center justify-between gap-2"><button class="btn btn-primary btn-xs gap-1.5" type="button" @click="applyTemplate(template)"><Check :size="13" aria-hidden="true" />Use template</button><div class="flex items-center gap-1"><button class="btn btn-ghost btn-square btn-xs" type="button" title="Export JSON" aria-label="Export JSON" @click="exportTemplate(template)"><Download :size="13" aria-hidden="true" /></button><button v-if="template.source === 'custom'" class="btn btn-ghost btn-square btn-xs text-error" type="button" title="Delete saved template" aria-label="Delete saved template" @click="deleteCustomTemplate(template)"><Trash2 :size="13" aria-hidden="true" /></button></div></div>
            </article>
          </div>
          <div v-if="!availableTemplates.length" class="rounded-box border border-dashed border-base-300 p-8 text-center text-sm text-base-content/60">No templates are available yet. Import a JSON template or build parameters manually.</div>
        </div>
      </div>
    </div>

    <div v-if="replacementOpen && pendingTemplate" class="fixed inset-0 z-[60] grid place-items-center bg-black/60 p-4" role="dialog" aria-modal="true" aria-labelledby="replace-template-title">
      <div class="w-full max-w-lg rounded-box border border-base-300 bg-base-200 p-5 shadow-2xl">
        <div class="flex items-start gap-3"><span class="grid size-9 shrink-0 place-items-center rounded-box bg-warning/15 text-warning"><AlertTriangle :size="19" aria-hidden="true" /></span><div><h3 id="replace-template-title" class="text-base font-semibold">Replace current parameters?</h3><p class="mt-1 text-sm leading-5 text-base-content/65">Choosing <strong class="text-base-content">{{ pendingTemplate.title }}</strong> will remove the current setup and replace its temporary draft. Save it first if you may want to return to it.</p></div></div>
        <FormField class="mt-4 gap-1"><span>Reusable template name <em class="text-base-content/45">optional</em></span><AppInput v-model="templateName" class="input w-full min-w-0" placeholder="My print setup" /></FormField>
        <div class="mt-5 flex flex-wrap justify-end gap-2"><button class="btn btn-ghost btn-sm" type="button" @click="replacementOpen = false; pendingTemplate = null">Cancel</button><button class="btn btn-outline btn-sm" type="button" @click="replaceWithPending('replace')">Replace without saving</button><button class="btn btn-outline btn-sm gap-1.5" type="button" @click="replaceWithPending('draft')"><Save :size="14" aria-hidden="true" />Save draft & replace</button><button class="btn btn-primary btn-sm gap-1.5" type="button" @click="replaceWithPending('custom')"><FileJson :size="14" aria-hidden="true" />Save reusable & replace</button></div>
      </div>
    </div>
  </section>
</template>
