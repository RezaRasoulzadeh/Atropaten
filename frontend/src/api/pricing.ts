import { CalculateServicePrice, CalculateDraftServicePrice } from '../../wailsjs/go/main/App'
import type { main as mainTypes } from '../../wailsjs/go/models'

export type PrintLayout = { materialId: string; materialName: string; kind: string; quantity: string; consumedQuantity: string; unit: string; across: number; rows: number; itemsPerSheet: number; sheets: string; lengthMM: string; rotated: boolean; wastePercent: number; wasteCostRial: number; areaM2: string; originalLengthMM?: string }
export type PricingRecord = mainTypes.PricingDTO & { batchQuantity?: string; layouts?: PrintLayout[] }

export const pricingApi = {
  draft(input: unknown, parameters: Record<string,string>, quantity: string): Promise<PricingRecord> {
    return CalculateDraftServicePrice(input as mainTypes.ServiceInput, { parameters, quantity } as unknown as mainTypes.PricingRequest)
  },
  calculate(input: {
    quantity?: string
    serviceId: string
    parameters: Record<string, string>
    manualCosts?: Record<string, number>
    sellingPriceOverrideRial?: number | null
  }): Promise<PricingRecord> {
    return CalculateServicePrice(input as unknown as mainTypes.PricingRequest)
  },
}
