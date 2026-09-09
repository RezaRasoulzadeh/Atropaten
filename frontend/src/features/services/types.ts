export type ServiceFilter = 'Active' | 'Archived' | 'All';
export type EditorMode = 'create' | 'edit' | null;
export type ParameterType = 'integer' | 'decimal' | 'boolean' | 'choice' | 'material-reference';
export type ParameterForm = {
  id: string;
  key: string;
  label: string;
  type: ParameterType;
  required: boolean;
  defaultValue: string;
  options: string[];
  minValue: string | null;
  maxValue: string | null;
  unit: string;
};
export type ComponentType =
  | 'material'
  | 'machine'
  | 'labor'
  | 'outsourced'
  | 'fixed'
  | 'overhead'
  | 'waste'
  | 'manual';
export type ComponentForm = {
  id: string;
  name: string;
  type: ComponentType;
  referenceId: string;
  usageMode: 'fixed' | 'parameter';
  parameterKey: string;
  usageQuantity: string;
  multiplier: string;
  rateRial: number;
  rateInput: string;
  percentage: string;
  rateBasis: string;
  enabled: boolean;
  notes: string;
};
export type PricingTierForm = {
  position: number;
  minimumQuantity: string;
  priceRial: number;
  priceInput: string;
};
export type PricingRuleForm = {
  id: string;
  type: string;
  fixedPriceRial: number;
  fixedPriceInput: string;
  markupPercentage: string;
  fixedMarginRial: number;
  fixedMarginInput: string;
  perUnitRateRial: number;
  perUnitRateInput: string;
  parameterKey: string;
  tiers: PricingTierForm[];
};
export type ServiceForm = {
  name: string;
  code: string;
  category: string;
  description: string;
  imagePath: string;
  defaultUnit: string;
  defaultPriority: string;
  parameters: ParameterForm[];
  components: ComponentForm[];
  pricingRule: PricingRuleForm;
};
