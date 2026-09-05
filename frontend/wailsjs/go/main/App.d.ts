export interface MaterialInput {
  name: string
  sku: string
  category: string
  purchaseUnit: string
  consumptionUnit: string
  conversionFactor: string
  physicalStock: string
  reorderLevel: string
  averageUnitCostRial: number
  preferredSupplier: string
  notes: string
}

export interface MaterialDTO {
  id: string
  name: string
  sku: string
  category: string
  purchaseUnit: string
  consumptionUnit: string
  conversionFactor: string
  physicalStock: string
  reorderLevel: string
  averageUnitCostRial: number
  preferredSupplier: string
  notes: string
  active: boolean
  lowStock: boolean
  createdAt: string
  updatedAt: string
}

export function ListMaterials(includeArchived: boolean): Promise<MaterialDTO[]>
export function GetMaterial(id: string): Promise<MaterialDTO>
export function CreateMaterial(input: MaterialInput): Promise<MaterialDTO>
export function UpdateMaterial(id: string, input: MaterialInput): Promise<MaterialDTO>
export function ArchiveMaterial(id: string): Promise<MaterialDTO>
export function ReactivateMaterial(id: string): Promise<MaterialDTO>
