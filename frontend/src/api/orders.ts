import {
  AddOrderItem,
  ApplyOrderDiscount,
  CreateOrder,
  DeleteOrder,
  GetOrder,
  ListOrders,
  RemoveOrderItem,
  ReplaceOrderItem,
  ReorderOrderItems,
  UpdateOrder,
  UpdateOrderCommercialStatus,
  UpdateOrderFulfillmentStatus,
  ArchiveOrder,
  UnarchiveOrder,
} from '../../wailsjs/go/main/App'
export interface OrderItemRecord {
  id: string
  position: number
  serviceId: string
  serviceName: string
  serviceCode: string
  quantity: string
  quantityUnit: string
  resolvedParametersJson: string
  costBreakdownJson: string
  pricingSnapshotJson: string
  estimatedCostRial: number
  suggestedPriceRial: number
  sellingPriceRial: number
  notes: string
}
export interface OrderRecord {
  id: string
  orderNumber: string
  customerId: string
  customerName: string
  customerPhone: string
  notes: string
  createdAt: string
  updatedAt: string
  promisedAt: string | null
  priority: string
  commercialStatus: string
  fulfillmentStatus: string
  paymentStatus: string
  archived: boolean
  invoiceId?: string
  invoiceStatus?: string
  invoicedTotalRial?: number
  subtotalRial: number
  discountRial: number
  totalRial: number
  paidRial: number
  remainingRial: number
  estimatedCostRial: number
  actualCostRial: number
  projectedCostRial: number
  marginRial: number
  marginPercentage: string
  productionJobCount: number
  completedProductionJobs: number
  inProgressProductionJobs: number
  items: OrderItemRecord[]
}
export interface OrderPayload {
  customerId: string
  promisedAt: string | null
  priority: string
  notes: string
  discountRial: number
}
export interface OrderItemPayload {
  serviceId: string
  parameters: Record<string, string>
  manualCosts: Record<string, number>
  sellingPriceOverrideRial: number | null
  quantity: string
  quantityUnit: string
  notes: string
}

function normalizeOrder(record: OrderRecord): OrderRecord {
  return { ...record, items: Array.isArray(record.items) ? record.items : [] }
}

export const ordersApi = {
  list(): Promise<OrderRecord[]> {
    return (ListOrders() as Promise<OrderRecord[]>).then((rows) => rows.map(normalizeOrder))
  },
  get(id: string): Promise<OrderRecord> {
    return (GetOrder(id) as Promise<OrderRecord>).then(normalizeOrder)
  },
  create(input: OrderPayload): Promise<OrderRecord> {
    return (
      CreateOrder(
        input as unknown as import('../../wailsjs/go/models').main.OrderInput,
      ) as Promise<OrderRecord>
    ).then(normalizeOrder)
  },
  update(id: string, input: OrderPayload): Promise<OrderRecord> {
    return (
      UpdateOrder(
        id,
        input as unknown as import('../../wailsjs/go/models').main.OrderInput,
      ) as Promise<OrderRecord>
    ).then(normalizeOrder)
  },
  remove(id: string): Promise<void> {
    return DeleteOrder(id)
  },
  addItem(id: string, input: OrderItemPayload): Promise<OrderRecord> {
    return (
      AddOrderItem(
        id,
        input as unknown as import('../../wailsjs/go/models').main.OrderItemInput,
      ) as Promise<OrderRecord>
    ).then(normalizeOrder)
  },
  replaceItem(id: string, itemId: string, input: OrderItemPayload): Promise<OrderRecord> {
    return (
      ReplaceOrderItem(
        id,
        itemId,
        input as unknown as import('../../wailsjs/go/models').main.OrderItemInput,
      ) as Promise<OrderRecord>
    ).then(normalizeOrder)
  },
  removeItem(id: string, itemId: string): Promise<OrderRecord> {
    return (RemoveOrderItem(id, itemId) as Promise<OrderRecord>).then(normalizeOrder)
  },
  reorderItems(id: string, ids: string[]): Promise<OrderRecord> {
    return (ReorderOrderItems(id, ids) as Promise<OrderRecord>).then(normalizeOrder)
  },
  discount(id: string, amount: number): Promise<OrderRecord> {
    return (ApplyOrderDiscount(id, amount) as Promise<OrderRecord>).then(normalizeOrder)
  },
  commercialStatus(id: string, status: string): Promise<OrderRecord> {
    return (UpdateOrderCommercialStatus(id, status) as Promise<OrderRecord>).then(normalizeOrder)
  },
  archive(id: string): Promise<OrderRecord> {
    return (ArchiveOrder(id) as Promise<OrderRecord>).then(normalizeOrder)
  },
  unarchive(id: string): Promise<OrderRecord> {
    return (UnarchiveOrder(id) as Promise<OrderRecord>).then(normalizeOrder)
  },
  fulfillmentStatus(id: string, status: string): Promise<OrderRecord> {
    return (UpdateOrderFulfillmentStatus(id, status) as Promise<OrderRecord>).then(normalizeOrder)
  },
}
