import {
  ArchiveMaterial,
  CreateMaterial,
  GetMaterial,
  ListMaterials,
  ReactivateMaterial,
  UpdateMaterial,
} from '../../wailsjs/go/main/App'
import type { main as mainTypes } from '../../wailsjs/go/models'

export type MaterialRecord = mainTypes.MaterialDTO
export type MaterialPayload = mainTypes.MaterialInput

export const materialsApi = {
  list(includeArchived = true): Promise<MaterialRecord[]> {
    return ListMaterials(includeArchived)
  },
  create(input: MaterialPayload): Promise<MaterialRecord> {
    return CreateMaterial(input)
  },
  get(id: string): Promise<MaterialRecord> {
    return GetMaterial(id)
  },
  update(id: string, input: MaterialPayload): Promise<MaterialRecord> {
    return UpdateMaterial(id, input)
  },
  archive(id: string): Promise<MaterialRecord> {
    return ArchiveMaterial(id)
  },
  reactivate(id: string): Promise<MaterialRecord> {
    return ReactivateMaterial(id)
  },
}
