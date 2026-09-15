export type ServiceCategoryRequirements = {
  material: boolean
  machine: boolean
}

export const SERVICE_CATEGORIES = [
  'Print products',
  'Paper printing',
  'Large format',
  'Banners & signage',
  'Cards & stationery',
  'Flyers & brochures',
  'Labels & stickers',
  'Finishing',
  'Design',
  'Packaging',
] as const

const CATEGORY_REQUIREMENTS: Record<string, ServiceCategoryRequirements> = {
  'print products': { material: true, machine: true },
  'paper printing': { material: true, machine: true },
  'large format': { material: true, machine: true },
  'banners & signage': { material: true, machine: true },
  'cards & stationery': { material: true, machine: true },
  'flyers & brochures': { material: true, machine: true },
  'labels & stickers': { material: true, machine: true },
  // Finishing work varies: some jobs are manual and others use stock or equipment.
  finishing: { material: false, machine: false },
  // Design is intentionally time/labor-only by default.
  design: { material: false, machine: false },
  // Packaging can consume stock without requiring production equipment.
  packaging: { material: true, machine: false },
}

export function serviceCategoryRequirements(category: string): ServiceCategoryRequirements {
  return CATEGORY_REQUIREMENTS[category.trim().toLocaleLowerCase()] || { material: false, machine: false }
}

const LAYOUT_CATEGORIES = new Set([
  'print products',
  'paper printing',
  'large format',
  'banners & signage',
  'cards & stationery',
  'flyers & brochures',
  'labels & stickers',
])

export function serviceCategorySupportsLayout(category: string): boolean {
  return LAYOUT_CATEGORIES.has(category.trim().toLowerCase())
}
