import {
  ArchiveMaterial,
  CreateMaterial,
  DeleteMaterial,
  GetMaterial,
  ListMaterialAttributeDefinitions,
  ListPredefinedParameters,
  ListMaterials,
  ReactivateMaterial,
  UpdateMaterial,
} from '../../wailsjs/go/main/App'
import type { main as mainTypes } from '../../wailsjs/go/models'

export type MaterialRecord = mainTypes.MaterialDTO
export type MaterialAttributeDefinitionRecord = mainTypes.MaterialAttributeDefinitionDTO
export type PredefinedParameterRecord = mainTypes.PredefinedParameterDTO
export type MaterialAttributePayload = {
  key: string
  valueType: string
  decimalValue: string
  integerValue: number
  enumCode: string
  textValue: string
  booleanValue: boolean
}
export type MaterialPayload = {
  name: string
  sku: string
  kind: string
  attributes: MaterialAttributePayload[]
  purchaseUnit: string
  consumptionUnit: string
  conversionFactor: string
  physicalStock: string
  reorderLevel: string
  averageUnitCostRial: number
  preferredSupplier: string
  notes: string
}

export const materialsApi = {
  list(includeArchived = true): Promise<MaterialRecord[]> {
    return ListMaterials(includeArchived)
  },
  create(input: MaterialPayload): Promise<MaterialRecord> {
    return CreateMaterial(input as unknown as mainTypes.MaterialInput)
  },
  get(id: string): Promise<MaterialRecord> {
    return GetMaterial(id)
  },
  definitions(): Promise<MaterialAttributeDefinitionRecord[]> {
    return ListMaterialAttributeDefinitions()
  },
  predefinedParameters(): Promise<PredefinedParameterRecord[]> {
    return ListPredefinedParameters()
  },
  update(id: string, input: MaterialPayload): Promise<MaterialRecord> {
    return UpdateMaterial(id, input as unknown as mainTypes.MaterialInput)
  },
  archive(id: string): Promise<MaterialRecord> {
    return ArchiveMaterial(id)
  },
  reactivate(id: string): Promise<MaterialRecord> {
    return ReactivateMaterial(id)
  },
  remove(id: string): Promise<void> {
    return DeleteMaterial(id)
  },
}
