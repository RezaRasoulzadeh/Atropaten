import {
  ArchiveService,
  CreateService,
  GetService,
  ListServices,
  ReactivateService,
  UpdateService,
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
export type ServicePayload = {
  name: string
  code: string
  category: string
  description: string
  parameters: ServiceParameterPayload[]
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
}
