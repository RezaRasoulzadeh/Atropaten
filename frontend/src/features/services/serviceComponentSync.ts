import type { ComponentForm, ParameterForm } from './types'

export function isNumericParameter(parameter: ParameterForm | undefined) {
  return parameter?.type === 'integer' || parameter?.type === 'decimal'
}

export function isChoiceParameter(parameter: ParameterForm | undefined) {
  return parameter?.type === 'choice' && (Boolean(parameter.predefinedKey) || parameter.options.length > 0)
}

export function isMaterialParameter(parameter: ParameterForm | undefined) {
  return parameter?.type === 'material-reference' || (parameter?.type === 'choice' && Boolean(parameter.materialSource))
}

export function isMachineParameter(parameter: ParameterForm | undefined) {
  return parameter?.type === 'machine-reference'
}

function draftComponentID() {
  return `draft-component-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function suggestedComponent(type: 'material' | 'machine', name: string): ComponentForm {
  return {
    id: draftComponentID(),
    name,
    type,
    referenceId: '',
    usageMode: 'fixed',
    parameterKey: '',
    rateId: '',
    rateParameterKey: '',
    usageQuantity: '1',
    multiplier: '1',
    rateRial: 0,
    rateInput: '',
    percentage: '',
    rateBasis: '',
    enabled: true,
    notes: '',
    suggested: true,
  }
}

export function ensureSuggestedCostComponents(components: ComponentForm[], parameters: ParameterForm[]) {
  for (let index = components.length - 1; index >= 0; index -= 1) {
    const component = components[index]
    if (component.suggested && component.usageMode === 'parameter' && !component.parameterKey) components.splice(index, 1)
  }
  const candidates = parameters.filter((parameter) => isMaterialParameter(parameter) || isMachineParameter(parameter))
  if (!components.length && !candidates.length) {
    components.push(suggestedComponent('material', 'Material cost'))
    return false
  }
  for (const parameter of candidates) {
    const type = isMachineParameter(parameter) ? 'machine' : 'material'
    const alreadyLinked = components.some((component) => component.type === type && component.usageMode === 'parameter' && component.parameterKey === parameter.key)
    if (alreadyLinked) continue
    const placeholder = components.find((component) => component.suggested && component.type === type && component.usageMode === 'fixed' && !component.referenceId)
    if (placeholder) {
      placeholder.name = `${parameter.label || parameter.key} cost`
      placeholder.usageMode = 'parameter'
      placeholder.parameterKey = parameter.key
      continue
    }
    components.push(suggestedComponent(type, `${parameter.label || parameter.key} cost`))
    const added = components[components.length - 1]
    added.usageMode = 'parameter'
    added.parameterKey = parameter.key
  }
  return candidates.length > 0
}

export function reconcileCostComponents(components: ComponentForm[], parameters: ParameterForm[]) {
  const byKey = new Map(parameters.filter((parameter) => parameter.key.trim()).map((parameter) => [parameter.key, parameter]))
  for (const component of components) {
    const sourceParameter = byKey.get(component.parameterKey)
    if (component.type === 'material' && component.usageMode === 'parameter' && !isMaterialParameter(sourceParameter)) component.parameterKey = ''
    if (component.type === 'machine' && component.usageMode === 'parameter' && !isMachineParameter(sourceParameter)) component.parameterKey = ''
    if (component.type !== 'material' && component.type !== 'machine' && component.type !== 'overhead' && component.type !== 'waste' && component.usageMode === 'parameter' && !isNumericParameter(sourceParameter)) component.parameterKey = ''

    const rateParameter = byKey.get(component.rateParameterKey)
    if (component.rateParameterKey && !isChoiceParameter(rateParameter)) component.rateParameterKey = ''
    if (component.type === 'material' && component.usageMode === 'parameter' && component.referenceId) component.referenceId = ''
    if (component.type === 'machine' && component.usageMode === 'parameter' && component.referenceId) component.referenceId = ''
  }
}
