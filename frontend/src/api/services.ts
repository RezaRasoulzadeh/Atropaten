import {
  ArchiveService,
  CreateService,
  DeleteService,
  GetService,
  ListServices,
  ReactivateService,
  UpdateService,
  AddServiceCostComponent,
  RemoveServiceCostComponent,
  ReorderServiceCostComponents,
  UpdateServiceCostComponent,
  GetServiceMaterialOptions,
  ListPredefinedParameters,
} from '../../wailsjs/go/main/App'
import type { main as mainTypes } from '../../wailsjs/go/models'

export type ServiceRecord = mainTypes.ServiceDTO
export type ServiceParameterPayload = {
  id: string
  key: string
  label: string
  type: string
  required: boolean
  defaultValue: string
  options: string[]
  minValue: string | null
  maxValue: string | null
  unit: string
  predefinedKey?: string
  materialSource?: any
}
export type ServiceCostComponentPayload = {
  id: string
  name: string
  type: string
  referenceId: string
  usageMode: string
  parameterKey: string
  rateId: string
  rateParameterKey: string
  usageQuantity: string
  multiplier: string
  rateRial: number
  percentage: string
  rateBasis: string
  enabled: boolean
  notes: string
}
export type ServiceMaterialVariantPayload = {
  id: string
  materialId: string
  values: Record<string, string>
  sellingPriceRial: number
  position: number
  active: boolean
}
export type PricingTierPayload = { position: number; minimumQuantity: string; priceRial: number }
export type PricingRulePayload = {
  id: string
  type: string
  fixedPriceRial: number
  markupPercentage: string
  fixedMarginRial: number
  perUnitRateRial: number
  parameterKey: string
  tiers: PricingTierPayload[]
}
export type FinishedSizePayload = {
  parameterKey: string
  quantityParameterKey: string
  widthParameterKey: string
  heightParameterKey: string
  allowCustom: boolean
  allowRotation: boolean
  options: Array<{ id: string; code: string; label: string; widthMM: string; heightMM: string; position: number; active: boolean }>
}
export type ServicePayload = {
  name: string
  code: string
  category: string
  description: string
  imagePath: string
  defaultUnit: string
  defaultPriority: string
  parameters: ServiceParameterPayload[]
  components: ServiceCostComponentPayload[]
  pricingRule: PricingRulePayload | null
  finishedSize: FinishedSizePayload | null
  materialVariants?: ServiceMaterialVariantPayload[]
}

export const servicesApi = {
  list(includeArchived = true): Promise<ServiceRecord[]> {
    return ListServices(includeArchived)
  },
  materialOptions(id: string, selected: Record<string, string>) {
    return GetServiceMaterialOptions(id, selected)
  },
  predefinedParameters() {
    return ListPredefinedParameters()
  },
  get(id: string): Promise<ServiceRecord> {
    return GetService(id)
  },
  create(input: ServicePayload): Promise<ServiceRecord> {
    return CreateService(input as unknown as mainTypes.ServiceInput)
  },
  update(id: string, input: ServicePayload): Promise<ServiceRecord> {
    return UpdateService(id, input as unknown as mainTypes.ServiceInput)
  },
  archive(id: string): Promise<ServiceRecord> {
    return ArchiveService(id)
  },
  reactivate(id: string): Promise<ServiceRecord> {
    return ReactivateService(id)
  },
  remove(id: string): Promise<void> {
    return DeleteService(id)
  },
  addComponent(id: string, input: ServiceCostComponentPayload): Promise<ServiceRecord> {
    return AddServiceCostComponent(id, input as unknown as mainTypes.ServiceCostComponentInput)
  },
  updateComponent(
    id: string,
    componentId: string,
    input: ServiceCostComponentPayload,
  ): Promise<ServiceRecord> {
    return UpdateServiceCostComponent(
      id,
      componentId,
      input as unknown as mainTypes.ServiceCostComponentInput,
    )
  },
  removeComponent(id: string, componentId: string): Promise<ServiceRecord> {
    return RemoveServiceCostComponent(id, componentId)
  },
  reorderComponents(id: string, componentIds: string[]): Promise<ServiceRecord> {
    return ReorderServiceCostComponents(id, componentIds)
  },
}
