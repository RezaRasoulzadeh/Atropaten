import {
  GetProductionMaterials,
  SetProductionMaterialTarget,
  UpdateProductionConsumption,
  CreateInventoryReservation,
  CreateProductionJob,
  DeleteProductionJob,
  GetProductionJob,
  ListInventoryReservations,
  ListProductionConsumptions,
  ListProductionJobs,
  RecordProductionConsumption,
  ReleaseInventoryReservation,
  ReverseProductionConsumption,
  UpdateInventoryReservation,
  UpdateProductionJob,
  UpdateProductionJobStatus,
  UpdateProductionOutsourcing,
} from '../../wailsjs/go/main/App'

export interface ProductionJobRecord {
  outsourceQuantity: string
  outsourceUnitCostRial: number
  outsourceFinancialAccountId: string
  id: string
  jobNumber: string
  orderId: string
  orderItemId: string
  serviceName: string
  quantity: string
  quantityUnit: string
  assignedMachineId: string
  status: string
  priority: string
  notes: string
  plannedAt: string
  startedAt: string
  completedAt: string
  createdAt: string
  estimatedCostRial: number
  actualMaterialCostRial: number
  actualWasteCostRial: number
  actualTotalCostRial: number
  actualOutsourcedCostRial: number
  outsourceQuotedCostRial: number
  outsourceSupplierId: string
  outsourceDescription: string
  outsourceSentAt: string
  outsourceExpectedReturnAt: string
  outsourceReceivedAt: string
  outsourceNotes: string
}
export interface ReservationRecord {
  id: string
  materialId: string
  orderId: string
  orderItemId: string
  productionJobId: string
  quantity: string
  status: string
  createdAt: string
  updatedAt: string
}
export interface ConsumptionRecord {
  reversed: boolean
  id: string
  productionJobId: string
  materialId: string
  idempotencyKey: string
  consumedQuantity: string
  wasteQuantity: string
  notes: string
  createdAt: string
  unitCostRial: number
  materialCostRial: number
  wasteCostRial: number
}
export interface ProductionJobPayload {
  orderId: string
  orderItemId: string
  quantity: string
  quantityUnit: string
  assignedMachineId: string
  priority: string
  notes: string
  plannedAt: string | null
}
export interface ReservationPayload {
  materialId: string
  orderId: string
  orderItemId: string
  productionJobId: string
  quantity: string
}
export interface ConsumptionPayload {
  materialId: string
  consumedQuantity: string
  wasteQuantity: string
  idempotencyKey: string
  notes: string
}
export interface OutsourcePayload {
  quantity: string
  unitCostRial: number
  financialAccountId: string
  supplierId: string
  description: string
  sentAt: string
  expectedReturnAt: string
  receivedAt: string
  notes: string
  quotedCostRial: number
  actualCostRial: number
}
export const productionApi = {
  materials(jobId: string): Promise<ProductionMaterialRecord[]> {
    return GetProductionMaterials(jobId)
  },
  setMaterialTarget(jobId: string, materialId: string, quantity: string) {
    return SetProductionMaterialTarget(jobId, materialId, quantity)
  },
  updateConsumption(id: string, value: ConsumptionPayload) {
    return UpdateProductionConsumption(id, value as any)
  },
  list(status = 'All') {
    return ListProductionJobs(status) as unknown as Promise<ProductionJobRecord[]>
  },
  get(id: string) {
    return GetProductionJob(id) as unknown as Promise<ProductionJobRecord>
  },
  create(v: ProductionJobPayload) {
    return CreateProductionJob(v as any) as unknown as Promise<ProductionJobRecord>
  },
  update(id: string, v: ProductionJobPayload) {
    return UpdateProductionJob(id, v as any) as unknown as Promise<ProductionJobRecord>
  },
  status(id: string, status: string) {
    return UpdateProductionJobStatus(id, status) as unknown as Promise<ProductionJobRecord>
  },
  delete(id: string) {
    return DeleteProductionJob(id)
  },
  reserve(v: ReservationPayload) {
    return CreateInventoryReservation(v as any) as unknown as Promise<ReservationRecord>
  },
  updateReservation(id: string, quantity: string) {
    return UpdateInventoryReservation(id, quantity) as unknown as Promise<ReservationRecord>
  },
  releaseReservation(id: string) {
    return ReleaseInventoryReservation(id)
  },
  reservations(materialId = '', jobId = '', orderId = '') {
    return ListInventoryReservations(materialId, jobId, orderId) as unknown as Promise<
      ReservationRecord[]
    >
  },
  consume(jobId: string, v: ConsumptionPayload) {
    return RecordProductionConsumption(jobId, v as any) as unknown as Promise<ConsumptionRecord>
  },
  reverseConsumption(id: string, reason: string) {
    return ReverseProductionConsumption(id, reason)
  },
  consumptions(jobId: string) {
    return ListProductionConsumptions(jobId) as unknown as Promise<ConsumptionRecord[]>
  },
  outsource(id: string, v: OutsourcePayload) {
    return UpdateProductionOutsourcing(id, v as any) as unknown as Promise<ProductionJobRecord>
  },
}

export interface ProductionMaterialRecord {
  materialId: string
  name: string
  unit: string
  reservationId: string
  required: string
  planned: string
  reserved: string
  used: string
  available: string
  shortage: string
}
