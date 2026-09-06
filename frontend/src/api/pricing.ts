import { CalculateServicePrice } from '../../wailsjs/go/main/App'
import type { main as mainTypes } from '../../wailsjs/go/models'

export type PricingRecord = mainTypes.PricingDTO

export const pricingApi = {
  calculate(input: { serviceId: string; parameters: Record<string, string>; manualCosts?: Record<string, number>; sellingPriceOverrideRial?: number | null }): Promise<PricingRecord> {
    return CalculateServicePrice(input as unknown as mainTypes.PricingRequest)
  },
}
