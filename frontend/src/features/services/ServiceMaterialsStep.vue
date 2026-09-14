<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { AlertTriangle, Check, ChevronRight, CircleHelp, Layers3, Package, Plus, Trash2 } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { MaterialAttributeDefinitionRecord, MaterialRecord } from '../../api/materials'
import type { MaterialParameterSourceForm, ParameterForm, ServiceMaterialVariantForm } from './types'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'

const props = defineProps<{
  parameters: ParameterForm[]
  materials: MaterialRecord[]
  attributeDefinitions: MaterialAttributeDefinitionRecord[]
  materialVariants: ServiceMaterialVariantForm[]
  currencyUnit?: CurrencyUnit
  showErrors?: boolean
}>()

const selectedIndex = ref(0)
const focusedGroupTitle = ref('')

const materialGroups = computed(() => props.parameters.filter((parameter) => parameter.type === 'choice' && parameter.materialSource))
const activeMaterials = computed(() => props.materials.filter((material) => material.active))
const activeGroup = computed(() => materialGroups.value[selectedIndex.value] || null)
const materialKinds = computed(() => Array.from(new Set(activeMaterials.value.map((material) => material.kind).filter(Boolean))))

type MaterialOption = { value: string; label: string; materialIds: string[] }
type CombinationRow = { key: string; values: Record<string, string>; variant: ServiceMaterialVariantForm | null; candidates: MaterialRecord[] }

function newID(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function sourceKeys(parameter: ParameterForm) {
  const source = parameter.materialSource
  if (!source) return []
  return source.exposedAttributeKeys?.length ? source.exposedAttributeKeys : source.exposedAttributeKey ? [source.exposedAttributeKey] : []
}

function attributeDefinition(key: string) {
  return props.attributeDefinitions.find((definition) => definition.key === key)
}

function attributeLabel(key: string) {
  const definition = attributeDefinition(key)
  return definition ? `${definition.label}${definition.unit ? ` (${definition.unit})` : ''}` : key
}

function kindLabel(kind: string) {
  return kind.split('-').map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join(' ')
}

function attributeValue(material: MaterialRecord, key: string): any | null {
  return (material.attributes || []).find((attribute: any) => attribute.key === key) || null
}

function canonicalAttribute(attribute: any) {
  if (!attribute) return ''
  if (attribute.valueType === 'decimal') return String(attribute.decimalValue ?? '')
  if (attribute.valueType === 'integer') return String(attribute.integerValue ?? '')
  if (attribute.valueType === 'enum') return String(attribute.enumCode ?? '').trim()
  if (attribute.valueType === 'boolean') return attribute.booleanValue ? 'true' : 'false'
  return String(attribute.textValue ?? '').trim()
}

function displayAttribute(attribute: any, key: string) {
  const value = canonicalAttribute(attribute)
  const definition = attributeDefinition(key)
  return definition?.unit ? `${value} ${definition.unit}` : value
}

function sourceValue(material: MaterialRecord, parameter: ParameterForm) {
  const values = sourceKeys(parameter).map((key) => canonicalAttribute(attributeValue(material, key)))
  return values.every(Boolean) ? values.join('\u001f') : ''
}

function sourceLabel(material: MaterialRecord, parameter: ParameterForm) {
  return sourceKeys(parameter).map((key) => displayAttribute(attributeValue(material, key), key)).join(' × ')
}

function materialCost(material: MaterialRecord) {
  return material.highestPurchaseUnitCostRial || material.averageUnitCostRial || 0
}

function materialLabel(material: MaterialRecord) {
  const cost = props.currencyUnit ? ` · ${formatMoney(materialCost(material), props.currencyUnit)}` : ''
  return `${material.name}${material.sku ? ` · ${material.sku}` : ''}${cost}`
}

function materialsFor(parameter: ParameterForm) {
  const source = parameter.materialSource
  const keys = sourceKeys(parameter)
  return activeMaterials.value.filter((material) => {
    if (source?.allowedKinds?.length && !source.allowedKinds.includes(material.kind)) return false
    return keys.every((key) => Boolean(attributeValue(material, key)))
  })
}

function groupOptions(parameter: ParameterForm): MaterialOption[] {
  const options = new Map<string, MaterialOption>()
  for (const material of materialsFor(parameter)) {
    const value = sourceValue(material, parameter)
    if (!value) continue
    const existing = options.get(value)
    if (existing) existing.materialIds.push(material.id)
    else options.set(value, { value, label: sourceLabel(material, parameter), materialIds: [material.id] })
  }
  return Array.from(options.values()).sort((left, right) => left.label.localeCompare(right.label, undefined, { numeric: true }))
}

function selectedGroupOptionCount(parameter: ParameterForm) {
  return groupOptions(parameter).length
}

function setGroupDefault(parameter: ParameterForm, value: string) {
  if (groupOptions(parameter).some((option) => option.value === value)) parameter.defaultValue = value
}

function isDefaultCombination(row: CombinationRow) {
  return materialGroups.value.every((group) => group.defaultValue && group.defaultValue === row.values[group.key])
}

function combinations(groups: ParameterForm[]): Array<{ key: string; values: Record<string, string> }> {
  if (!groups.length) return []
  let rows: Array<{ key: string; values: Record<string, string> }> = [{ key: '', values: {} }]
  for (const group of groups) {
    const options = groupOptions(group)
    rows = rows.flatMap((row) => options.map((option) => ({
      key: row.key ? `${row.key}\u001e${group.key}=${option.value}` : `${group.key}=${option.value}`,
      values: { ...row.values, [group.key]: option.value },
    })))
  }
  return rows
}

function findVariant(values: Record<string, string>) {
  return props.materialVariants.find((variant) => {
    const keys = Object.keys(values)
    return keys.length === Object.keys(variant.values).length && keys.every((key) => variant.values[key] === values[key])
  }) || null
}

function candidatesFor(values: Record<string, string>) {
  return activeMaterials.value.filter((material) => materialGroups.value.every((group) => {
    const selected = values[group.key]
    if (!selected) return false
    return sourceValue(material, group) === selected
  }))
}

const combinationRows = computed<CombinationRow[]>(() => combinations(materialGroups.value).map((row) => ({
  ...row,
  variant: findVariant(row.values),
  candidates: candidatesFor(row.values),
})))

function syncVariants() {
  for (const group of materialGroups.value) {
    const options = groupOptions(group)
    if (!options.some((option) => option.value === group.defaultValue)) group.defaultValue = options[0]?.value || ''
  }
  const rows = combinations(materialGroups.value)
  const current = new Map(props.materialVariants.map((variant) => [variantKey(variant.values), variant]))
  const next: ServiceMaterialVariantForm[] = []
  for (const row of rows) {
    const candidates = candidatesFor(row.values)
    const existing = current.get(variantKey(row.values))
    const materialId = existing && candidates.some((material) => material.id === existing.materialId)
      ? existing.materialId
      : candidates.length === 1 ? candidates[0].id : existing?.materialId || ''
    next.push(existing || { id: newID('variant'), materialId, values: { ...row.values }, sellingPriceRial: 0, sellingPriceInput: '', position: next.length, active: true })
    next[next.length - 1].values = { ...row.values }
    next[next.length - 1].materialId = materialId
    next[next.length - 1].position = next.length - 1
    next[next.length - 1].active = true
  }
  props.materialVariants.splice(0, props.materialVariants.length, ...next)
}

function variantKey(values: Record<string, string>) {
  return Object.keys(values).sort().map((key) => `${key}=${values[key]}`).join('\u001e')
}

function createGroup() {
  const baseKey = 'material_option'
  let key = baseKey
  let suffix = 2
  while (props.parameters.some((parameter) => parameter.key === key)) key = `${baseKey}_${suffix++}`
  const preferredKeys = props.attributeDefinitions.some((definition) => definition.key === 'width_mm') && props.attributeDefinitions.some((definition) => definition.key === 'height_mm')
    ? ['width_mm', 'height_mm']
    : props.attributeDefinitions.filter((definition) => definition.active).slice(0, 1).map((definition) => definition.key)
  const source: MaterialParameterSourceForm = {
    allowedKinds: materialKinds.value.includes('sheet-stock') ? ['sheet-stock'] : [],
    exposedAttributeKey: preferredKeys[0] || '',
    exposedAttributeKeys: preferredKeys,
    allowedValues: [],
    selectMaterial: false,
    additionalFilters: [],
  }
  const parameter: ParameterForm = {
    id: newID('parameter'),
    key,
    label: 'Paper size',
    type: 'choice',
    required: true,
    defaultValue: '',
    options: [],
    minValue: null,
    maxValue: null,
    unit: '',
    materialSource: source,
  }
  props.parameters.push(parameter)
  selectedIndex.value = materialGroups.value.length - 1
  focusedGroupTitle.value = parameter.id
  void nextTick(() => document.querySelector<HTMLInputElement>(`[data-material-group-title="${parameter.id}"]`)?.focus())
}

function removeGroup() {
  if (!activeGroup.value) return
  const index = props.parameters.indexOf(activeGroup.value)
  if (index < 0) return
  props.parameters.splice(index, 1)
  selectedIndex.value = Math.max(0, Math.min(selectedIndex.value, materialGroups.value.length - 1))
  syncVariants()
}

function selectGroup(index: number) {
  selectedIndex.value = index
}

function setGroupKind(parameter: ParameterForm, value: string) {
  if (!parameter.materialSource) return
  parameter.materialSource.allowedKinds = value ? [value] : []
  syncVariants()
}

function toggleAttribute(parameter: ParameterForm, key: string, event: Event) {
  if (!parameter.materialSource) return
  const checked = (event.target as HTMLInputElement).checked
  const keys = sourceKeys(parameter).filter((item) => item !== key)
  if (checked) keys.push(key)
  if (!keys.length) keys.push(key)
  parameter.materialSource.exposedAttributeKeys = keys
  parameter.materialSource.exposedAttributeKey = keys[0] || ''
  syncVariants()
}

function setVariantMaterial(variant: ServiceMaterialVariantForm | null, materialId: string) {
  if (variant) variant.materialId = materialId
}

function groupNeedsSetup(parameter: ParameterForm) {
  const options = groupOptions(parameter)
  return !parameter.label.trim() || !sourceKeys(parameter).length || !options.length || !options.some((option) => option.value === parameter.defaultValue)
}

watch(() => props.parameters, syncVariants, { deep: true })
watch(() => props.materials, syncVariants, { deep: true })

syncVariants()
</script>

<template>
  <section class="min-w-0 space-y-4" aria-label="Service material option groups">
    <div class="flex min-w-0 flex-wrap items-start justify-between gap-3 rounded-box border border-primary/25 bg-primary/5 p-4">
      <div class="flex min-w-0 items-start gap-3">
        <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Package :size="20" aria-hidden="true" /></span>
        <div class="min-w-0"><h2 class="text-base font-semibold">Material option groups</h2><p class="mt-1 max-w-3xl text-sm leading-5 text-base-content/65">Create the choices customers will see—such as Paper size or Paper type—and map every combination to the exact inventory material and cost used for pricing.</p></div>
      </div>
      <button class="btn btn-primary btn-sm shrink-0 gap-2" type="button" :disabled="!activeMaterials.length" @click="createGroup"><Plus :size="15" aria-hidden="true" />Add grouped option</button>
    </div>

    <div v-if="!activeMaterials.length" class="flex items-start gap-3 rounded-box border border-warning/30 bg-warning/10 p-4 text-sm"><AlertTriangle class="mt-0.5 shrink-0 text-warning" :size="18" aria-hidden="true" /><div><strong class="font-semibold">Add active materials first</strong><p class="mt-1 text-xs leading-5 text-base-content/65">Material groups are built from inventory records. Add sizes, types, and costs in Materials, then return here to group them.</p></div></div>

    <div v-if="!materialGroups.length" class="rounded-box border border-dashed border-primary/35 bg-base-100 p-8 text-center">
      <Layers3 class="mx-auto text-primary" :size="28" aria-hidden="true" /><h3 class="mt-3 text-base font-semibold">Start with your first material group</h3><p class="mx-auto mt-1 max-w-lg text-sm leading-5 text-base-content/60">For example, add “Paper size”, choose Width and Height, and the available sizes will be created from your active inventory. Add “Paper type” the same way.</p>
      <button class="btn btn-primary btn-sm mt-4 gap-2" type="button" :disabled="!activeMaterials.length" @click="createGroup"><Plus :size="15" aria-hidden="true" />Add grouped option</button>
    </div>

    <div v-else class="grid min-w-0 gap-4 lg:grid-cols-[minmax(15rem,0.72fr)_minmax(0,1.28fr)]">
      <div class="min-w-0 space-y-2">
        <div class="flex items-center justify-between px-1"><div><h3 class="text-sm font-semibold">Your groups</h3><p class="mt-1 text-xs text-base-content/60">Each group becomes one customer choice.</p></div><span class="badge badge-ghost text-xs">{{ materialGroups.length }}</span></div>
        <button v-for="(group, index) in materialGroups" :key="group.id" class="flex w-full min-w-0 items-center gap-3 rounded-box border p-3 text-start transition-colors" :class="index === selectedIndex ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-100 hover:border-primary/45'" type="button" @click="selectGroup(index)">
          <span class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-primary"><Layers3 :size="18" aria-hidden="true" /></span><span class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ group.label || 'Untitled group' }}</strong><small class="mt-0.5 block truncate text-xs text-base-content/60">{{ selectedGroupOptionCount(group) }} option{{ selectedGroupOptionCount(group) === 1 ? '' : 's' }} · {{ sourceKeys(group).map(attributeLabel).join(' + ') || 'Choose inventory attributes' }}</small></span><span v-if="!groupNeedsSetup(group)" class="grid size-6 shrink-0 place-items-center rounded-full bg-success/15 text-success"><Check :size="14" aria-hidden="true" /></span><ChevronRight class="shrink-0 text-base-content/45" :size="16" aria-hidden="true" />
        </button>
        <button class="flex w-full items-center justify-center gap-2 rounded-box border border-dashed border-primary/45 px-3 py-3 text-sm text-primary hover:border-primary" type="button" :disabled="!activeMaterials.length" @click="createGroup"><Plus :size="15" aria-hidden="true" />Add grouped option</button>
      </div>

      <div v-if="activeGroup" class="min-w-0 space-y-4 rounded-box border border-base-300 bg-base-100 p-4 sm:p-5">
        <div class="flex flex-wrap items-start justify-between gap-3 border-b border-base-300 pb-4"><div class="min-w-0"><p class="text-xs font-semibold uppercase tracking-wide text-primary">Grouped option</p><h3 class="mt-1 text-xl font-semibold">{{ activeGroup.label || 'Untitled group' }}</h3><p class="mt-1 text-sm text-base-content/65">Name the choice and select the inventory attributes that define its options.</p></div><button class="btn btn-outline btn-error btn-sm gap-2" type="button" @click="removeGroup"><Trash2 :size="14" aria-hidden="true" />Remove group</button></div>
        <div class="grid min-w-0 gap-4 sm:grid-cols-2">
          <FormField class="gap-1"><span>Group title <em class="text-error">*</em></span><AppInput v-model="activeGroup.label" :data-material-group-title="activeGroup.id" class="input w-full min-w-0" :class="{ 'input-error': showErrors && !activeGroup.label.trim() }" placeholder="Paper size" /></FormField>
          <FormField class="gap-1"><span>Material kind</span><SelectField :model-value="activeGroup.materialSource?.allowedKinds?.[0] || ''" label="" :options="[{ label: 'All material kinds', value: '' }, ...materialKinds.map((kind) => ({ label: kindLabel(kind), value: kind }))]" @update:model-value="setGroupKind(activeGroup, $event)" /></FormField>
        </div>

        <div class="rounded-box border border-base-300 bg-base-200/25 p-3.5"><div class="flex items-start gap-2"><CircleHelp class="mt-0.5 shrink-0 text-info" :size="16" aria-hidden="true" /><div><h4 class="text-sm font-semibold">What defines an option?</h4><p class="mt-1 text-xs leading-5 text-base-content/65">Choose one attribute for a simple group, or multiple attributes for a combined value. Paper size uses Width + Height; paper type can use Material subtype or Finish.</p></div></div><div class="mt-3 grid gap-2 sm:grid-cols-2"><label v-for="definition in attributeDefinitions.filter((item) => item.active)" :key="definition.key" class="flex min-w-0 items-center gap-2 rounded-box border border-base-300 bg-base-100 px-3 py-2 text-sm"><input class="checkbox checkbox-sm" type="checkbox" :checked="sourceKeys(activeGroup).includes(definition.key)" @change="toggleAttribute(activeGroup, definition.key, $event)" /><span class="min-w-0 truncate">{{ definition.label }}<small v-if="definition.unit" class="ml-1 text-xs text-base-content/55">({{ definition.unit }})</small></span></label></div></div>

        <div class="rounded-box border border-base-300"><div class="flex flex-wrap items-start justify-between gap-2 border-b border-base-300 px-4 py-3"><div><h4 class="text-sm font-semibold">Options from inventory</h4><p class="mt-1 text-xs text-base-content/60">{{ selectedGroupOptionCount(activeGroup) }} distinct value{{ selectedGroupOptionCount(activeGroup) === 1 ? '' : 's' }} found in active materials. Mark one as the default for the estimate.</p></div><span class="badge badge-ghost text-xs">Automatic</span></div><div v-if="sourceKeys(activeGroup).length && selectedGroupOptionCount(activeGroup)" class="divide-y divide-base-300/70"><div v-for="option in groupOptions(activeGroup)" :key="option.value" class="flex min-w-0 items-center justify-between gap-3 px-4 py-3"><div class="min-w-0"><strong class="block truncate text-sm">{{ option.label }}</strong><small class="block truncate text-xs text-base-content/60">{{ option.materialIds.length }} matching material{{ option.materialIds.length === 1 ? '' : 's' }}</small></div><label class="flex shrink-0 items-center gap-1.5 text-xs" :class="activeGroup.defaultValue === option.value ? 'font-medium text-primary' : 'text-base-content/60'"><input class="radio radio-primary radio-sm" type="radio" :name="`default-material-${activeGroup.id}`" :checked="activeGroup.defaultValue === option.value" @change="setGroupDefault(activeGroup, option.value)" />Default</label></div></div><div v-else class="p-4 text-sm text-warning">Choose at least one inventory attribute with active materials.</div></div>

        <div class="rounded-box border border-primary/30 bg-primary/5"><div class="border-b border-primary/15 px-4 py-3"><h4 class="text-sm font-semibold">Material combination pricing</h4><p class="mt-1 text-xs leading-5 text-base-content/65">Select the exact inventory material for every combination. Its recorded unit cost will be used in estimation.</p></div><div v-if="combinationRows.length" class="divide-y divide-base-300/70"><div v-for="row in combinationRows" :key="row.key" class="grid min-w-0 gap-3 px-4 py-3 sm:grid-cols-[minmax(0,1fr)_minmax(13rem,0.9fr)]"><div class="min-w-0"><div class="flex flex-wrap gap-1.5"><span v-for="group in materialGroups" :key="group.key" class="badge badge-ghost max-w-full truncate text-xs">{{ group.label }}: {{ groupOptions(group).find((option) => option.value === row.values[group.key])?.label || row.values[group.key] }}</span><span v-if="isDefaultCombination(row)" class="badge badge-primary badge-outline text-xs">Default estimate</span></div><small class="mt-1 block text-xs text-base-content/55">{{ row.candidates.length }} compatible material{{ row.candidates.length === 1 ? '' : 's' }} found</small></div><SelectField :model-value="row.variant?.materialId || ''" label="Exact inventory material" :options="[{ label: row.candidates.length ? 'Choose material…' : 'No compatible material', value: '' }, ...row.candidates.map((material) => ({ label: materialLabel(material), value: material.id }))]" @update:model-value="setVariantMaterial(row.variant, $event)" /></div></div><div v-else class="p-5 text-sm text-warning">No combinations can be generated yet. Check the selected attributes and active inventory.</div></div>
        <p v-if="combinationRows.length && combinationRows.some((row) => !row.variant?.materialId)" class="text-xs leading-5 text-warning">Complete each combination before saving. Pricing will remain unavailable for an unmapped material choice.</p>
      </div>
    </div>
  </section>
</template>
