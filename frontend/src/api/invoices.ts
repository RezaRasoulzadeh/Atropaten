import {
  CreateInvoiceFromOrder,
  DeleteDraftInvoice,
  GetInvoice,
  ListInvoices,
  PostInvoice,
  VoidInvoice,
} from '../../wailsjs/go/main/App'

export interface InvoiceItemRecord {
  id: string
  orderItemId: string
  description: string
  serviceId: string
  quantityUnit: string
  notes: string
  position: number
  quantity: string
  unitPriceRial: number
  lineTotalRial: number
}
export interface InvoiceRecord {
  id: string
  invoiceNumber: string
  customerId: string
  customerName: string
  customerPhone: string
  orderId: string
  issueDate: string
  dueDate: string
  status: string
  notes: string
  subtotalRial: number
  discountRial: number
  totalRial: number
  paidRial: number
  remainingRial: number
  accountingJournalEntryId: string
  cogsJournalEntryId: string
  createdAt: string
  updatedAt: string
  items: InvoiceItemRecord[]
}

function normalizeInvoice(record: InvoiceRecord): InvoiceRecord {
  return { ...record, items: Array.isArray(record.items) ? record.items : [] }
}

export const invoicesApi = {
  list(): Promise<InvoiceRecord[]> {
    return (ListInvoices() as unknown as Promise<InvoiceRecord[]>).then((rows) =>
      rows.map(normalizeInvoice),
    )
  },
  get(id: string): Promise<InvoiceRecord> {
    return (GetInvoice(id) as unknown as Promise<InvoiceRecord>).then(normalizeInvoice)
  },
  createFromOrder(orderId: string): Promise<InvoiceRecord> {
    return (CreateInvoiceFromOrder(orderId) as unknown as Promise<InvoiceRecord>).then(
      normalizeInvoice,
    )
  },
  post(id: string): Promise<InvoiceRecord> {
    return (PostInvoice(id) as unknown as Promise<InvoiceRecord>).then(normalizeInvoice)
  },
  void(id: string): Promise<InvoiceRecord> {
    return (VoidInvoice(id) as unknown as Promise<InvoiceRecord>).then(normalizeInvoice)
  },
  deleteDraft(id: string): Promise<void> {
    return DeleteDraftInvoice(id)
  },
}
