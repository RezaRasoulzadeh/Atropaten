export type ServiceFilter = 'Active' | 'Archived' | 'All';
export type EditorMode = 'create' | 'edit' | null;
export type ParameterType = 'integer' | 'decimal' | 'boolean' | 'choice' | 'material-reference' | 'machine-reference';
export type MaterialParameterSourceForm = {
  allowedKinds: string[];
  exposedAttributeKey: string;
  exposedAttributeKeys: string[];
  allowedValues: any[];
  selectMaterial: boolean;
  additionalFilters: any[];
};
export type MaterialAttributeDefinitionForm = {
  key: string;
  label: string;
  valueType: string;
  unit: string;
  applicableKinds: string[];
  active: boolean;
  position: number;
};
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
  predefinedKey?: string;
  materialSource?: MaterialParameterSourceForm;
};
export type ParameterTemplateSeed = Omit<ParameterForm, 'id'>;
export type ComponentType =
  | 'material'
  | 'machine'
  | 'service'
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
  rateId: string;
  rateParameterKey: string;
  usageQuantity: string;
  multiplier: string;
  rateRial: number;
  rateInput: string;
  percentage: string;
  rateBasis: string;
  enabled: boolean;
  notes: string;
  suggested?: boolean;
};
export type PricingTierForm = {
  position: number;
  minimumQuantity: string;
  priceRial: number;
  priceInput: string;
};
export type PricingVariationForm = {
  id: string;
  values: Record<string, string>;
  priceRial: number;
  priceInput: string;
  tiers: PricingTierForm[];
  position: number;
  active: boolean;
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
  variations: PricingVariationForm[];
};
export type FinishedSizeOptionForm = {
  id: string;
  code: string;
  label: string;
  widthMM: string;
  heightMM: string;
  position: number;
  active: boolean;
};
export type FinishedSizeForm = {
  parameterKey: string;
  quantityParameterKey: string;
  widthParameterKey: string;
  heightParameterKey: string;
  allowCustom: boolean;
  allowRotation: boolean;
  options: FinishedSizeOptionForm[];
};
export type ServiceForm = {
  name: string;
  code: string;
  category: string;
  description: string;
  imagePath: string;
  defaultUnit: string;
  defaultPriority: string;
  fulfillmentMode: 'in-house' | 'outsourced';
  defaultOutsourcedCostRial: number;
  defaultOutsourcedCostInput: string;
  defaultOutsourcedShippingRial: number;
  defaultOutsourcedShippingInput: string;
  parameters: ParameterForm[];
  components: ComponentForm[];
  pricingRule: PricingRuleForm;
  finishedSize: FinishedSizeForm | null;
  materialVariants: ServiceMaterialVariantForm[];
};

export type ServiceMaterialVariantForm = {
  id: string;
  materialId: string;
  values: Record<string, string>;
  sellingPriceRial: number;
  sellingPriceInput: string;
  position: number;
  active: boolean;
};

export type PredefinedParameterOption = {
  code: string;
  label: string;
  widthMM: string | null;
  heightMM: string | null;
  active: boolean;
  position: number;
};
export type PredefinedParameter = {
  key: string;
  label: string;
  valueType: string;
  unit: string;
  active: boolean;
  position: number;
  options: PredefinedParameterOption[];
};
