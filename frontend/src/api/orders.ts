import { AddOrderItem, ApplyOrderDiscount, CreateOrder, GetOrder, ListOrders, RemoveOrderItem, ReplaceOrderItem, ReorderOrderItems, UpdateOrder, UpdateOrderCommercialStatus, UpdateOrderFulfillmentStatus } from '../../wailsjs/go/main/App'
export interface OrderItemRecord { id:string; position:number; serviceId:string; serviceName:string; serviceCode:string; quantity:string; quantityUnit:string; resolvedParametersJson:string; costBreakdownJson:string; pricingSnapshotJson:string; estimatedCostRial:number; suggestedPriceRial:number; sellingPriceRial:number; notes:string }
export interface OrderRecord { id:string; orderNumber:string; customerId:string; customerName:string; customerPhone:string; notes:string; createdAt:string; updatedAt:string; promisedAt:string|null; priority:string; commercialStatus:string; fulfillmentStatus:string; paymentStatus:string; invoiceId?:string; invoiceStatus?:string; invoicedTotalRial?:number; subtotalRial:number; discountRial:number; totalRial:number; estimatedCostRial:number; productionJobCount:number; completedProductionJobs:number; inProgressProductionJobs:number; items:OrderItemRecord[] }
export interface OrderPayload { customerId:string; promisedAt:string|null; priority:string; notes:string; discountRial:number }
export interface OrderItemPayload { serviceId:string; parameters:Record<string,string>; manualCosts:Record<string,number>; sellingPriceOverrideRial:number|null; quantity:string; quantityUnit:string; notes:string }
export const ordersApi = {
  list():Promise<OrderRecord[]> { return ListOrders() as Promise<OrderRecord[]> },
  get(id:string):Promise<OrderRecord> { return GetOrder(id) as Promise<OrderRecord> },
  create(input:OrderPayload):Promise<OrderRecord> { return CreateOrder(input as unknown as import('../../wailsjs/go/models').main.OrderInput) as Promise<OrderRecord> },
  update(id:string,input:OrderPayload):Promise<OrderRecord> { return UpdateOrder(id,input as unknown as import('../../wailsjs/go/models').main.OrderInput) as Promise<OrderRecord> },
  addItem(id:string,input:OrderItemPayload):Promise<OrderRecord> { return AddOrderItem(id,input as unknown as import('../../wailsjs/go/models').main.OrderItemInput) as Promise<OrderRecord> },
  replaceItem(id:string,itemId:string,input:OrderItemPayload):Promise<OrderRecord> { return ReplaceOrderItem(id,itemId,input as unknown as import('../../wailsjs/go/models').main.OrderItemInput) as Promise<OrderRecord> },
  removeItem(id:string,itemId:string):Promise<OrderRecord> { return RemoveOrderItem(id,itemId) as Promise<OrderRecord> },
  reorderItems(id:string,ids:string[]):Promise<OrderRecord> { return ReorderOrderItems(id,ids) as Promise<OrderRecord> },
  discount(id:string,amount:number):Promise<OrderRecord> { return ApplyOrderDiscount(id,amount) as Promise<OrderRecord> },
  commercialStatus(id:string,status:string):Promise<OrderRecord> { return UpdateOrderCommercialStatus(id,status) as Promise<OrderRecord> },
  fulfillmentStatus(id:string,status:string):Promise<OrderRecord> { return UpdateOrderFulfillmentStatus(id,status) as Promise<OrderRecord> },
}
