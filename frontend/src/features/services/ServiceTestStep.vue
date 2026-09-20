<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue'
import { pricingApi, type PricingRecord } from '../../api/pricing'
import { Calculator, CheckCircle2, Eye, FlaskConical, RotateCcw, TriangleAlert } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
import SelectField from '../../components/ui/SelectField.vue'
import type { MaterialRecord } from '../../api/materials'
import type { MachineRecord } from '../../api/machines'
import type { ServiceRecord } from '../../api/services'
import { formatMoney, roundMoneyUp, DEFAULT_MONETARY_ROUNDING_STEP_RIAL, type CurrencyUnit } from '../../utils/currency'
import { calculateServiceTest, isAutomaticVariationParameter, isMachineRateParameter, machineGroupOptions, machineRateOptions, outsourcedDefaultCostRial, outsourcedDefaultLines, testParameterLabel, visibleTestParameters, type TestPricingResult, type TestValues } from './serviceTestPricing'
import type { ParameterForm, ServiceForm } from './types'
import { translateUi } from '../../i18n'
import RollSizeFields from './RollSizeFields.vue'
import { ensureRollSizeInputs, materialOptionsForParameter, usesMaterialRollWidth, selectedRollMaterial, rollWidthValue, rollStockWidthValue } from './rollSizeInputs'

const props = defineProps<{
  form: ServiceForm
  parameters: ParameterForm[]
  materials: MaterialRecord[]
  machines: MachineRecord[]
  services: ServiceRecord[]
  values: TestValues
  currencyUnit: CurrencyUnit
  roundingStepRial?: number
}>()

const emit = defineEmits<{
  'update:result': [result: TestPricingResult]
}>()

const materialWidthMode = computed(() => usesMaterialRollWidth(props.form))
const selectedRoll = computed(() => selectedRollMaterial(props.form, props.materials, props.values))
const materialWidth = computed(() => rollWidthValue(selectedRoll.value, props.values.layout_margin_mm ?? props.parameters.find(p => p.key === 'layout_margin_mm')?.defaultValue ?? '0'))
const materialWidthLimit = computed(() => rollStockWidthValue(selectedRoll.value))
const inputParameters = computed(() => visibleTestParameters(props.form, props.parameters).filter(p =>
  !materialWidthMode.value || ![props.form.finishedSize?.widthParameterKey, props.form.finishedSize?.heightParameterKey].includes(p.key),
))

function defaultValue(parameter: ParameterForm) {
  if (parameter.defaultValue) return parameter.defaultValue
  if (parameter.type === 'boolean') return 'false'
  return ''
}

function syncValues() {
  const parameters = visibleTestParameters(props.form, props.parameters)
  const keys = new Set(parameters.map((parameter) => parameter.key).filter(Boolean))
  for (const parameter of parameters) {
    if (parameter.key && props.values[parameter.key] === undefined) props.values[parameter.key] = defaultValue(parameter)
  }
  for (const key of Object.keys(props.values)) if (!keys.has(key)) delete props.values[key]
}

function syncDynamicSelections() {
  for (const parameter of visibleTestParameters(props.form, props.parameters)) {
    if (!isAutomaticVariationParameter(props.form, parameter)) continue
    const options = valueOptions(parameter).filter((option) => option.value)
    if (options.length && !options.some((option) => option.value === props.values[parameter.key])) props.values[parameter.key] = options[0].value
  }
}

const draftPrice = ref<PricingRecord | null>(null)
const draftError = ref('')
let draftToken = 0
let draftTimer: ReturnType<typeof setTimeout> | undefined
onBeforeUnmount(() => { draftToken++; clearTimeout(draftTimer) })

function recalculate() {
  ensureRollSizeInputs(props.form)
  syncValues()
  syncDynamicSelections()
  if (materialWidthMode.value) {
    const key = props.form.finishedSize!.widthParameterKey
    if (!props.values[key] && materialWidth.value) props.values[key] = materialWidth.value
  }
  if (!props.form.finishedSize?.quantityParameterKey) {
    emit('update:result', calculateServiceTest(props.form, props.values, props.materials, props.machines, props.services, props.roundingStepRial)); return
  }
  const token = ++draftToken
  clearTimeout(draftTimer)
  draftPrice.value = null
  if (materialWidthMode.value && !materialWidth.value) {
    draftError.value = 'Select a roll material with a valid width. Edge margins must leave a positive printable width.'
    return
  }
  if (materialWidthMode.value) {
    const widthKey = props.form.finishedSize!.widthParameterKey
    const width = Number(props.values[widthKey])
    const limit = Number(materialWidthLimit.value)
    if (!Number.isFinite(width) || width <= 0 || !Number.isFinite(limit) || limit <= 0 || width > limit) {
      draftError.value = `Enter a positive custom width no greater than ${materialWidthLimit.value || 'the selected roll width'} mm.`
      return
    }
  }
  draftError.value = 'Updating layout…'
  draftTimer = setTimeout(async () => {
    try {
      const next = await pricingApi.draft(props.form, { ...props.values }, props.values[props.form.finishedSize!.quantityParameterKey] || '1')
      if (token !== draftToken) return
      draftPrice.value = next; draftError.value = ''
    } catch (error) { if (token === draftToken) draftError.value = String(error) }
  }, 250)
}

function reset() {
  for (const parameter of visibleTestParameters(props.form, props.parameters)) if (parameter.key) props.values[parameter.key] = defaultValue(parameter)
  recalculate()
}

function updateBoolean(key: string, event: Event) {
  props.values[key] = (event.target as HTMLInputElement).checked ? 'true' : 'false'
  recalculate()
}

function valueOptions(parameter: ParameterForm) {
	const selectLabel = translateUi(`Select ${translateUi(parameter.label.toLowerCase())}`)
	if (isMachineRateParameter(props.form, parameter)) return [{ label: selectLabel, value: '' }, ...machineRateOptions(props.form, parameter, props.parameters, props.machines, props.values)]
	if (parameter.type === 'choice') {
    if (parameter.materialSource) {
      return [{ label: selectLabel, value: '' }, ...materialOptionsForParameter(props.form, parameter, props.materials, props.values)]
    }
    return [{ label: selectLabel, value: '' }, ...(parameter.predefinedKey ? ((parameter as any).predefinedOptions || []).filter((option: any) => option.active !== false).map((option: any) => ({ label: option.label, value: option.code })) : parameter.options.map((value) => ({ label: value, value })))]
  }
  return [{ label: 'Select machine', value: '' }, ...machineGroupOptions(parameter, props.machines).map((item) => ({ label: `${item.name}${item.code ? ` · ${item.code}` : ''}`, value: item.id }))]
}

function onValueChanged() {
  recalculate()
}

watch(
  () => [props.parameters, props.form.components, props.form.pricingRule, props.form.finishedSize, props.materials, props.machines, props.services, props.values, props.roundingStepRial],
  recalculate,
  { deep: true, immediate: true },
)

const result = computed<TestPricingResult>(() => {
  if (!props.form.finishedSize?.quantityParameterKey) return calculateServiceTest(props.form, props.values, props.materials, props.machines, props.services, props.roundingStepRial)
  const p = draftPrice.value
  const outsourcedCost = outsourcedDefaultCostRial(props.form)
  const roundingStepRial = p?.roundingStepRial || props.roundingStepRial
  const step = roundingStepRial || DEFAULT_MONETARY_ROUNDING_STEP_RIAL
  const totalCostRial = roundMoneyUp((p?.estimatedCostRial || 0) + outsourcedCost, step)
  const lines = [
    ...(p?.components.map(c => ({ name: c.name, detail: c.explanation, amount: c.amountRial, missing: false })) || []),
    ...outsourcedDefaultLines(props.form),
  ]
  let sellingPriceRial = p?.effectiveSellingPriceRial || 0
  let markupRial = p ? p.suggestedSellingPriceRial - p.estimatedCostRial : 0
  if (p && props.form.pricingRule.type === 'markup') {
    markupRial = Math.ceil(totalCostRial * (Number(props.form.pricingRule.markupPercentage) || 0) / 100)
    sellingPriceRial = roundMoneyUp(totalCostRial + markupRial, step)
  } else if (p && props.form.pricingRule.type === 'fixed-margin') {
    sellingPriceRial = roundMoneyUp(totalCostRial + props.form.pricingRule.fixedMarginRial, step)
  }
  const profitRial = sellingPriceRial - totalCostRial
  return { lines, totalCostRial, markupRial, sellingPriceRial, pricingLabel: 'Complete batch', profitRial, marginPercentage: sellingPriceRial ? profitRial / sellingPriceRial * 100 : 0, belowCost: p ? sellingPriceRial < totalCostRial : false, hasMissing: !p }
})
watch(result, r => emit('update:result', r), {immediate:true})
</script>

<template>
  <section class="min-w-0 space-y-4" :aria-label='$t("Test service")'>
    <p v-if="form.finishedSize?.quantityParameterKey && draftError" class="text-sm text-warning">{{ draftError }}</p>
    <div v-for="layout in draftPrice?.layouts || []" :key="layout.materialId" class="rounded-box border border-primary/30 p-4 text-sm">
      <strong>{{ layout.materialName }}</strong> · {{ layout.consumedQuantity }} {{ $ui(layout.unit) }} · {{ layout.wastePercent.toFixed(1) }}{{ $t("% waste") }}
      <p v-if="layout.wasteCostRial > 0">{{ $t("Estimated material waste cost:") }} {{ formatMoney(layout.wasteCostRial, currencyUnit) }} {{ $t("(included in material cost)") }}</p>
      <p>{{ $ui(layout.rotated ? 'Rotate artwork 90°.' : 'Original orientation.') }} {{ $ui(layout.itemsPerSheet ? `${layout.itemsPerSheet} pieces/sheet · ${layout.sheets} sheets` : `${layout.across} across × ${layout.rows} rows · ${Number(layout.lengthMM)/1000} m of roll`) }}</p>
    </div>
    <div class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1.05fr)_minmax(18rem,0.95fr)]">
      <div class="min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-5">
        <div class="flex min-w-0 items-start justify-between gap-3 border-b border-base-300 pb-4">
          <div class="flex min-w-0 items-start gap-3">
            <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><FlaskConical :size="21" aria-hidden="true" /></span>
            <div class="min-w-0">
              <h2 class="text-lg font-semibold">{{ $t("Test service") }}</h2>
              <p class="mt-1 text-sm leading-5 text-base-content/60">{{ $t("Choose dynamic material and machine options, then preview the calculated cost and selling price.") }}</p>
            </div>
          </div>
          <button class="btn btn-outline btn-sm shrink-0 gap-2" type="button" @click="reset"><RotateCcw :size="14" aria-hidden="true" />{{ $t("Reset") }}</button>
        </div>

        <div v-if="materialWidthMode" class="mt-4 rounded-box border border-primary/30 bg-primary/5 p-4 space-y-3">
          <h3 class="font-semibold">{{ $t("Roll width & finished size") }}</h3>
          <RollSizeFields :width="values[form.finishedSize!.widthParameterKey] || ''" :height="values[form.finishedSize!.heightParameterKey] || ''" :max-width="materialWidthLimit" :width-invalid="Boolean(draftError && draftError.includes('custom width'))" :material-name="selectedRoll?.name" @width="values[form.finishedSize!.widthParameterKey] = $event; onValueChanged()" @height="values[form.finishedSize!.heightParameterKey] = $event; onValueChanged()" />
        </div>
        <div v-if="inputParameters.length" class="mt-4 space-y-3">
          <div v-for="parameter in inputParameters" :key="parameter.id" class="grid min-w-0 grid-cols-[2rem_minmax(0,1fr)] items-start gap-3">
            <span class="grid size-8 place-items-center rounded-box bg-base-300/60 text-base-content/70"><span v-if="parameter.type === 'integer' || parameter.type === 'decimal'" class="text-lg">#</span><span v-else-if="parameter.type === 'boolean'" class="text-sm">✓</span><span v-else class="text-base">◈</span></span>
            <FormField class="min-w-0 gap-1">
              <span class="text-sm text-base-content">{{ testParameterLabel(form, parameter, parameters) }}<em v-if="parameter.required" class="text-error"> *</em></span>
              <small class="text-xs text-base-content/50">{{ $ui(isAutomaticVariationParameter(form, parameter) ? 'Affects automatic variation pricing' : form.finishedSize?.quantityParameterKey ? 'Used to calculate the complete print layout' : 'Price can be entered on the order') }}</small>
              <AppInput v-if="parameter.type === 'integer' || parameter.type === 'decimal'" v-model="values[parameter.key]" class="input w-full min-w-0" :type="parameter.type === 'integer' ? 'number' : 'text'" :step="parameter.type === 'integer' ? '1' : 'any'" :min="parameter.minValue || undefined" :max="parameter.maxValue || undefined" inputmode="decimal" @update:model-value="onValueChanged" />
              <SelectField v-else-if="parameter.type === 'choice' || parameter.type === 'machine-reference'" v-model="values[parameter.key]" :aria-label="parameter.label" :options="valueOptions(parameter)" @update:model-value="onValueChanged" />
              <label v-else class="flex h-10 items-center gap-2 rounded-box border border-base-300 bg-base-100 px-3 text-sm"><input class="checkbox checkbox-sm" type="checkbox" :checked="values[parameter.key] === 'true'" @change="updateBoolean(parameter.key, $event)" />{{ $t("Enabled") }}</label>
              <small v-if="parameter.unit || parameter.minValue || parameter.maxValue" class="text-xs leading-5 text-base-content/55">{{ $ui(parameter.unit || form.defaultUnit) }}<span v-if="parameter.minValue"> {{ $t("· min") }} {{ parameter.minValue }}</span><span v-if="parameter.maxValue"> {{ $t("· max") }} {{ parameter.maxValue }}</span></small>
            </FormField>
          </div>
        </div>
        <div v-else class="mt-5 rounded-box border border-dashed border-base-300 p-6 text-center text-sm text-base-content/60">{{ $t("No automatic pricing inputs are configured. Grouped material and machine choices are tested here; other order-only fields are set when the order is created.") }}</div>
      </div>

      <div class="min-w-0 rounded-box border border-base-300 bg-base-200/20 p-4 sm:p-5">
        <div class="flex items-start gap-3 border-b border-base-300 pb-4">
          <span class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><Calculator :size="21" aria-hidden="true" /></span>
          <div><h2 class="text-lg font-semibold">{{ $t("Calculation result") }}</h2><p class="mt-1 text-sm leading-5 text-base-content/60">{{ $t("Real-time calculation based on your input.") }}</p></div>
        </div>
        <div class="mt-4 space-y-4">
          <div class="rounded-box border border-base-300 bg-base-100/30 p-3">
            <h3 class="text-sm font-semibold">{{ $t("Cost breakdown") }}</h3>
            <div v-if="result.lines.length" class="mt-3 divide-y divide-base-300/70">
              <div v-for="line in result.lines" :key="`${line.name}-${line.detail}`" class="flex min-w-0 items-center gap-2 py-2 first:pt-0 last:pb-0 text-sm"><span class="size-2 shrink-0 rounded-full" :class="line.missing ? 'bg-warning' : 'bg-primary'"></span><span class="min-w-0 flex-1"><span class="block truncate">{{ $ui(line.name) }}</span><small class="block truncate text-xs text-base-content/55">{{ $ui(line.detail) }}</small></span><span class="shrink-0 tabular-nums" :class="line.missing ? 'text-warning' : ''">{{ $ui(line.missing ? 'Needs setup' : formatMoney(line.amount, currencyUnit)) }}</span></div>
            </div>
            <p v-else class="mt-3 text-sm text-base-content/60">{{ $t("Add a cost component to calculate this service.") }}</p>
            <div class="mt-3 flex items-center justify-between gap-3 border-t border-base-300 pt-3 text-sm"><span class="font-semibold">{{ $t("Total cost") }}</span><strong class="tabular-nums">{{ formatMoney(result.totalCostRial, currencyUnit) }}</strong></div>
          </div>
          <div class="rounded-box border border-base-300 bg-base-100/30 p-3">
            <div class="flex items-center justify-between gap-3 text-sm"><span>{{ $ui(result.pricingLabel) }}</span><span v-if="form.pricingRule.type === 'markup'" class="tabular-nums">+ {{ formatMoney(result.markupRial, currencyUnit) }}</span><span v-else-if="form.pricingRule.type === 'fixed-margin'" class="tabular-nums">+ {{ formatMoney(form.pricingRule.fixedMarginRial, currencyUnit) }}</span><span v-else-if="form.pricingRule.type === 'fixed'" class="text-base-content/60">{{ $t("Direct price") }}</span></div>
            <div class="mt-3 flex items-center justify-between gap-3 rounded-box border border-success/25 px-3 py-3"><span class="font-semibold">{{ $ui(form.pricingRule.type === 'fixed' ? 'Fixed selling price' : 'Selling price') }}</span><strong class="text-lg tabular-nums" :class="result.hasMissing ? 'text-warning' : 'text-success'">{{ $ui(form.pricingRule.type === 'manual' ? 'Set in order' : result.hasMissing ? 'Needs setup' : formatMoney(result.sellingPriceRial, currencyUnit)) }}</strong></div>
          </div>
          <div class="flex items-start gap-3 rounded-box border p-3 text-sm" :class="result.hasMissing ? 'border-warning/30 text-warning' : form.pricingRule.type === 'manual' ? 'border-info/30 text-info' : result.belowCost ? 'border-warning/30 text-warning' : 'border-success/30 text-success'"><TriangleAlert v-if="result.hasMissing || result.belowCost" :size="18" class="mt-0.5 shrink-0" aria-hidden="true" /><Eye v-else-if="form.pricingRule.type === 'manual'" :size="18" class="mt-0.5 shrink-0" aria-hidden="true" /><CheckCircle2 v-else :size="18" class="mt-0.5 shrink-0" aria-hidden="true" /><div><strong>{{ $ui(result.hasMissing ? 'Pricing setup is incomplete' : form.pricingRule.type === 'manual' ? 'Price is entered on the order' : result.belowCost ? 'Price is below cost' : 'Price is above cost') }}</strong><p class="mt-1 text-xs leading-5 text-base-content/60">{{ $ui(result.hasMissing ? 'Configure the selected material or machine variation in the Pricing step.' : form.pricingRule.type === 'manual' ? 'This test shows the estimated cost; the operator sets the final selling price for each order.' : result.belowCost ? 'Review the pricing rule before saving this service.' : `Estimated margin: ${result.marginPercentage.toFixed(1)}%.`) }}</p></div></div>
          <div class="rounded-box border border-info/25 bg-info/5 p-3 text-xs leading-5 text-base-content/70">{{ $t("This is a test calculation. The final amount may vary with order conditions, discounts, or customer-specific rules.") }}</div>
        </div>
      </div>
    </div>
  </section>
</template>
