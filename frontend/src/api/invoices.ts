import { CreateInvoiceFromOrder, DeleteDraftInvoice, GetInvoice, ListInvoices, PostInvoice, VoidInvoice } from '../../wailsjs/go/main/App'

export interface InvoiceItemRecord { id:string; orderItemId:string; description:string; serviceId:string; quantityUnit:string; notes:string; position:number; quantity:string; unitPriceRial:number; lineTotalRial:number }
export interface InvoiceRecord { id:string; invoiceNumber:string; customerId:string; customerName:string; customerPhone:string; orderId:string; issueDate:string; dueDate:string; status:string; notes:string; subtotalRial:number; discountRial:number; totalRial:number; paidRial:number; remainingRial:number; accountingJournalEntryId:string; cogsJournalEntryId:string; createdAt:string; updatedAt:string; items:InvoiceItemRecord[] }
export const invoicesApi={
  list():Promise<InvoiceRecord[]>{return ListInvoices() as unknown as Promise<InvoiceRecord[]>},
  get(id:string):Promise<InvoiceRecord>{return GetInvoice(id) as unknown as Promise<InvoiceRecord>},
  createFromOrder(orderId:string):Promise<InvoiceRecord>{return CreateInvoiceFromOrder(orderId) as unknown as Promise<InvoiceRecord>},
  post(id:string):Promise<InvoiceRecord>{return PostInvoice(id) as unknown as Promise<InvoiceRecord>},
  void(id:string):Promise<InvoiceRecord>{return VoidInvoice(id) as unknown as Promise<InvoiceRecord>},
  deleteDraft(id:string):Promise<void>{return DeleteDraftInvoice(id)},
}
