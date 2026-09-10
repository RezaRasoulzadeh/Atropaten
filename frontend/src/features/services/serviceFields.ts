import type {ComponentType} from './types'
export function componentNeedsRate(type: ComponentType) {
  return type === 'labor' || type === 'outsourced' || type === 'fixed' || type === 'manual';
}
export function componentNeedsReference(type: ComponentType) {
  return type === 'material' || type === 'machine' || type === 'service';
}
export function componentNeedsPercentage(type: ComponentType) {
  return type === 'overhead' || type === 'waste';
}
