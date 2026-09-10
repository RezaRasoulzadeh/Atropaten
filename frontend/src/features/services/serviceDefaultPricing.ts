import type { MachineRecord } from '../../api/machines'
import type { MaterialRecord } from '../../api/materials'
import type { ServiceRecord } from '../../api/services'
import { calculateServiceTest, type TestPricingResult, type TestValues } from './serviceTestPricing'
import type { ComponentForm, ParameterForm, PricingRuleForm, ServiceForm } from './types'

function formFromService(service: ServiceRecord): ServiceForm {
  const parameters: ParameterForm[] = (service.parameters || []).map((parameter) => ({
    id: parameter.id,
    key: parameter.key,
    label: parameter.label,
    type: parameter.type as ParameterForm['type'],
    required: parameter.required,
    defaultValue: parameter.defaultValue || '',
    options: Array.isArray(parameter.options) ? [...parameter.options] : [],
    minValue: parameter.minValue ?? null,
    maxValue: parameter.maxValue ?? null,
    unit: parameter.unit || '',
  }))

  const components: ComponentForm[] = (service.components || []).map((component) => ({
    id: component.id,
    name: component.name,
    type: component.type as ComponentForm['type'],
    referenceId: component.referenceId || '',
    usageMode: component.usageMode as ComponentForm['usageMode'],
    parameterKey: component.parameterKey || '',
    rateId: component.rateId || '',
    rateParameterKey: component.rateParameterKey || '',
    usageQuantity: component.usageQuantity || '1',
    multiplier: component.multiplier || '1',
    rateRial: component.rateRial || 0,
    rateInput: '',
    percentage: component.percentage || '',
    rateBasis: component.rateBasis || 'unit',
    enabled: component.enabled,
    notes: component.notes || '',
  }))

  const sourceRule = service.pricingRule
  const pricingRule: PricingRuleForm = sourceRule
    ? {
        id: sourceRule.id,
        type: sourceRule.type,
        fixedPriceRial: sourceRule.fixedPriceRial || 0,
        fixedPriceInput: '',
        markupPercentage: sourceRule.markupPercentage || '',
        fixedMarginRial: sourceRule.fixedMarginRial || 0,
        fixedMarginInput: '',
        perUnitRateRial: sourceRule.perUnitRateRial || 0,
        perUnitRateInput: '',
        parameterKey: sourceRule.parameterKey || '',
        tiers: (sourceRule.tiers || []).map((tier) => ({
          position: tier.position,
          minimumQuantity: tier.minimumQuantity || '0',
          priceRial: tier.priceRial || 0,
          priceInput: '',
        })),
      }
    : {
        id: '',
        type: 'manual',
        fixedPriceRial: 0,
        fixedPriceInput: '',
        markupPercentage: '',
        fixedMarginRial: 0,
        fixedMarginInput: '',
        perUnitRateRial: 0,
        perUnitRateInput: '',
        parameterKey: '',
        tiers: [],
      }

  return {
    name: service.name,
    code: service.code,
    category: service.category,
    description: service.description,
    imagePath: service.imagePath || '',
    defaultUnit: service.defaultUnit || 'piece',
    defaultPriority: service.defaultPriority || 'Normal',
    parameters,
    components,
    pricingRule,
  }
}

function defaultValues(service: ServiceForm): TestValues {
  return Object.fromEntries(
    service.parameters
      .filter((parameter) => parameter.key)
      .map((parameter) => [
        parameter.key,
        parameter.defaultValue || (parameter.type === 'boolean' ? 'false' : ''),
      ]),
  )
}

/** Calculate the service using the defaults saved in its parameters. */
export function serviceDefaultPricingResult(
  service: ServiceRecord,
  materials: MaterialRecord[],
  machines: MachineRecord[],
  services: ServiceRecord[],
): TestPricingResult | null {
  if (!service.pricingRule || service.pricingRule.type === 'manual') return null
  const form = formFromService(service)
  return calculateServiceTest(form, defaultValues(form), materials, machines, services)
}

/** Return a usable final selling price, or null when defaults cannot produce one. */
export function serviceEstimatedSellingPrice(
  service: ServiceRecord,
  materials: MaterialRecord[],
  machines: MachineRecord[],
  services: ServiceRecord[],
): number | null {
  const result = serviceDefaultPricingResult(service, materials, machines, services)
  if (!result || result.sellingPriceRial <= 0) return null

  const requiresCompleteCost = service.pricingRule?.type === 'markup' || service.pricingRule?.type === 'fixed-margin'
  if (requiresCompleteCost && result.hasMissing) return null
  return result.sellingPriceRial
}
