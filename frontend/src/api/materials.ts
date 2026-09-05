import {
  ArchiveMaterial,
  CreateMaterial,
  GetMaterial,
  ListMaterials,
  ReactivateMaterial,
  UpdateMaterial,
  type MaterialDTO,
  type MaterialInput,
} from '../../wailsjs/go/main/App'

export type MaterialRecord = MaterialDTO
export type MaterialPayload = MaterialInput

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
