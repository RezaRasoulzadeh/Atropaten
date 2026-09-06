import {
  ArchiveService,
  CreateService,
  GetService,
  ListServices,
  ReactivateService,
  UpdateService,
  AddServiceCostComponent,
  RemoveServiceCostComponent,
  ReorderServiceCostComponents,
  UpdateServiceCostComponent,
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
}
export type ServiceCostComponentPayload = {
  id: string
  name: string
  type: string
  referenceId: string
  usageMode: string
  parameterKey: string
  multiplier: string
  rateRial: number
  percentage: string
  rateBasis: string
  enabled: boolean
  notes: string
}
export type ServicePayload = {
  name: string
  code: string
  category: string
  description: string
  parameters: ServiceParameterPayload[]
  components: ServiceCostComponentPayload[]
}

export const servicesApi = {
  list(includeArchived = true): Promise<ServiceRecord[]> {
    return ListServices(includeArchived)
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
  addComponent(id: string, input: ServiceCostComponentPayload): Promise<ServiceRecord> { return AddServiceCostComponent(id, input as unknown as mainTypes.ServiceCostComponentInput) },
  updateComponent(id: string, componentId: string, input: ServiceCostComponentPayload): Promise<ServiceRecord> { return UpdateServiceCostComponent(id, componentId, input as unknown as mainTypes.ServiceCostComponentInput) },
  removeComponent(id: string, componentId: string): Promise<ServiceRecord> { return RemoveServiceCostComponent(id, componentId) },
  reorderComponents(id: string, componentIds: string[]): Promise<ServiceRecord> { return ReorderServiceCostComponents(id, componentIds) },
}
