export type CommercialStatus = 'Draft' | 'Quoted' | 'Confirmed' | 'Closed'
export type FulfillmentStatus = 'Pending' | 'In production' | 'Ready' | 'Delivered'
export type PaymentStatus = 'Unpaid' | 'Partially paid' | 'Paid'
export type Priority = 'Urgent' | 'High' | 'Normal' | 'Low'
export type ProductionStatus = 'Queued' | 'In progress' | 'Ready to print' | 'Delivered' | 'Waiting for proof'

export interface OrderItem {
  id: string
  service: string
  specification: string
  quantity: number
  unit: string
  sellingPriceRial: number
  estimatedCostRial: number
  productionStatus: ProductionStatus
}

export interface Order {
  id: string
  customer: string
  customerDetail: string
  created: string
  promised: string
  priority: Priority
  commercialStatus: CommercialStatus
  fulfillmentStatus: FulfillmentStatus
  paymentStatus: PaymentStatus
  itemSummary: string
  totalRial: number
  paidRial: number
  notes: string
  items: OrderItem[]
}

export const mockOrders: Order[] = [
  {
    id: 'ORD-1048',
    customer: 'Mehr Studio',
    customerDetail: 'mehr.studio@example.com · +98 21 4451 2080',
    created: '12 Aug 2024',
    promised: 'Today · 15:30',
    priority: 'High',
    commercialStatus: 'Confirmed',
    fulfillmentStatus: 'In production',
    paymentStatus: 'Partially paid',
    itemSummary: 'Business cards · A3 menus',
    totalRial: 92_000_000,
    paidRial: 46_000_000,
    notes: 'Call Mehr Studio before trimming. They approved the first proof by email.',
    items: [
      { id: 'item-1048-1', service: 'Business cards', specification: '500 pcs · 350gsm matte · full color · double-sided', quantity: 500, unit: 'pcs', sellingPriceRial: 54_000_000, estimatedCostRial: 31_200_000, productionStatus: 'In progress' },
      { id: 'item-1048-2', service: 'A3 menu printing', specification: '40 sheets · 300gsm gloss · full color · folded', quantity: 40, unit: 'sheets', sellingPriceRial: 38_000_000, estimatedCostRial: 21_400_000, productionStatus: 'Ready to print' },
    ],
  },
  {
    id: 'ORD-1045',
    customer: 'Arman Foods',
    customerDetail: 'orders@armanfoods.example · +98 21 8834 7110',
    created: '12 Aug 2024',
    promised: 'Today · 17:00',
    priority: 'Normal',
    commercialStatus: 'Confirmed',
    fulfillmentStatus: 'Ready',
    paymentStatus: 'Paid',
    itemSummary: 'A3 menus · window stickers',
    totalRial: 64_000_000,
    paidRial: 64_000_000,
    notes: 'Pack menus flat. Stickers should be separated by branch.',
    items: [
      { id: 'item-1045-1', service: 'A3 menu printing', specification: '80 sheets · 250gsm matte · full color · folded', quantity: 80, unit: 'sheets', sellingPriceRial: 42_000_000, estimatedCostRial: 24_800_000, productionStatus: 'Ready to print' },
      { id: 'item-1045-2', service: 'Window sticker set', specification: '6 pcs · clear vinyl · outdoor laminate', quantity: 6, unit: 'pcs', sellingPriceRial: 22_000_000, estimatedCostRial: 11_600_000, productionStatus: 'Delivered' },
    ],
  },
  {
    id: 'ORD-1042',
    customer: 'Nika Events',
    customerDetail: 'studio@nikaevents.example · +98 21 2268 1904',
    created: '11 Aug 2024',
    promised: '13 Aug · 10:00',
    priority: 'Urgent',
    commercialStatus: 'Quoted',
    fulfillmentStatus: 'Pending',
    paymentStatus: 'Unpaid',
    itemSummary: 'Event banners · artwork setup',
    totalRial: 148_000_000,
    paidRial: 0,
    notes: 'Waiting for the final sponsor lockup before sending to print.',
    items: [
      { id: 'item-1042-1', service: 'Event banners', specification: '3 pcs · 90 × 200cm · 440gsm PVC · eyelets', quantity: 3, unit: 'pcs', sellingPriceRial: 126_000_000, estimatedCostRial: 78_400_000, productionStatus: 'Queued' },
      { id: 'item-1042-2', service: 'Artwork setup', specification: '3 layouts · preflight and print-ready export', quantity: 3, unit: 'layouts', sellingPriceRial: 22_000_000, estimatedCostRial: 8_000_000, productionStatus: 'Waiting for proof' },
    ],
  },
  {
    id: 'ORD-1039',
    customer: 'Pendar Clinic',
    customerDetail: 'admin@pendarclinic.example · +98 21 8872 4022',
    created: '31 Jul 2024',
    promised: 'Overdue · 2 hours',
    priority: 'High',
    commercialStatus: 'Confirmed',
    fulfillmentStatus: 'In production',
    paymentStatus: 'Unpaid',
    itemSummary: 'Appointment cards · envelopes',
    totalRial: 245_000_000,
    paidRial: 0,
    notes: 'Clinic requested delivery before the afternoon appointments.',
    items: [
      { id: 'item-1039-1', service: 'Appointment cards', specification: '1,000 pcs · 300gsm matte · 1/1 color', quantity: 1000, unit: 'pcs', sellingPriceRial: 185_000_000, estimatedCostRial: 114_600_000, productionStatus: 'In progress' },
      { id: 'item-1039-2', service: 'Branded envelopes', specification: '1,000 pcs · C5 · self-seal · 1/0 color', quantity: 1000, unit: 'pcs', sellingPriceRial: 60_000_000, estimatedCostRial: 32_200_000, productionStatus: 'Queued' },
    ],
  },
  {
    id: 'ORD-1037',
    customer: 'Novin Architects',
    customerDetail: 'hello@novinarch.example · +98 21 2201 1196',
    created: '29 Jul 2024',
    promised: '09 Aug · 12:00',
    priority: 'Normal',
    commercialStatus: 'Closed',
    fulfillmentStatus: 'Delivered',
    paymentStatus: 'Paid',
    itemSummary: 'Presentation boards · binding',
    totalRial: 112_000_000,
    paidRial: 112_000_000,
    notes: 'Delivered to the reception desk.',
    items: [
      { id: 'item-1037-1', service: 'Presentation boards', specification: '12 boards · A2 · foamboard mount · matte', quantity: 12, unit: 'boards', sellingPriceRial: 96_000_000, estimatedCostRial: 51_200_000, productionStatus: 'Delivered' },
      { id: 'item-1037-2', service: 'Perfect binding', specification: '1 portfolio · 64 pages · soft cover', quantity: 1, unit: 'book', sellingPriceRial: 16_000_000, estimatedCostRial: 7_200_000, productionStatus: 'Delivered' },
    ],
  },
  {
    id: 'ORD-1032',
    customer: 'Cafe Saba',
    customerDetail: 'saba.cafe@example.com · +98 21 6671 0340',
    created: '26 Jul 2024',
    promised: '02 Aug · 16:00',
    priority: 'Low',
    commercialStatus: 'Closed',
    fulfillmentStatus: 'Delivered',
    paymentStatus: 'Paid',
    itemSummary: 'Loyalty cards · counter sign',
    totalRial: 38_500_000,
    paidRial: 38_500_000,
    notes: 'Repeat order. Use the saved Cafe Saba artwork.',
    items: [
      { id: 'item-1032-1', service: 'Loyalty cards', specification: '250 pcs · 350gsm matte · rounded corners', quantity: 250, unit: 'pcs', sellingPriceRial: 28_500_000, estimatedCostRial: 14_200_000, productionStatus: 'Delivered' },
      { id: 'item-1032-2', service: 'Counter sign', specification: '1 pc · A4 acrylic · full color insert', quantity: 1, unit: 'pc', sellingPriceRial: 10_000_000, estimatedCostRial: 4_200_000, productionStatus: 'Delivered' },
    ],
  },
]

export function createDraftOrder(): Order {
  return {
    id: 'DRAFT-NEW',
    customer: 'Select customer',
    customerDetail: 'Choose an existing customer or add a new one',
    created: 'Today · 12 Aug 2024',
    promised: 'Set promised date',
    priority: 'Normal',
    commercialStatus: 'Draft',
    fulfillmentStatus: 'Pending',
    paymentStatus: 'Unpaid',
    itemSummary: 'No services added yet',
    totalRial: 0,
    paidRial: 0,
    notes: '',
    items: [],
  }
}
