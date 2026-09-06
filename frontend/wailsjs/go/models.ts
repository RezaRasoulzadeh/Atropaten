export namespace main {

	export class AttachmentDTO {
	    id: string;
	    ownerType: string;
	    ownerId: string;
	    fileName: string;
	    path: string;
	    mimeType: string;
	    checksum: string;
	    category: string;
	    notes: string;
	    sizeBytes?: number;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new AttachmentDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ownerType = source["ownerType"];
	        this.ownerId = source["ownerId"];
	        this.fileName = source["fileName"];
	        this.path = source["path"];
	        this.mimeType = source["mimeType"];
	        this.checksum = source["checksum"];
	        this.category = source["category"];
	        this.notes = source["notes"];
	        this.sizeBytes = source["sizeBytes"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class CustomerDTO {
	    id: string;
	    name: string;
	    phone: string;
	    email: string;
	    address: string;
	    notes: string;
	    active: boolean;
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new CustomerDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.address = source["address"];
	        this.notes = source["notes"];
	        this.active = source["active"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class CustomerInput {
	    name: string;
	    phone: string;
	    email: string;
	    address: string;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new CustomerInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.address = source["address"];
	        this.notes = source["notes"];
	    }
	}
	export class MachineDTO {
	    id: string;
	    name: string;
	    code: string;
	    category: string;
	    rateBasis: string;
	    rateRial: number;
	    setupCostRial: number;
	    notes: string;
	    active: boolean;
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new MachineDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.code = source["code"];
	        this.category = source["category"];
	        this.rateBasis = source["rateBasis"];
	        this.rateRial = source["rateRial"];
	        this.setupCostRial = source["setupCostRial"];
	        this.notes = source["notes"];
	        this.active = source["active"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class MachineInput {
	    name: string;
	    code: string;
	    category: string;
	    rateBasis: string;
	    rateRial: number;
	    setupCostRial: number;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new MachineInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.code = source["code"];
	        this.category = source["category"];
	        this.rateBasis = source["rateBasis"];
	        this.rateRial = source["rateRial"];
	        this.setupCostRial = source["setupCostRial"];
	        this.notes = source["notes"];
	    }
	}
	export class MaterialDTO {
	    id: string;
	    name: string;
	    sku: string;
	    category: string;
	    purchaseUnit: string;
	    consumptionUnit: string;
	    conversionFactor: string;
	    physicalStock: string;
	    reorderLevel: string;
	    averageUnitCostRial: number;
	    preferredSupplier: string;
	    notes: string;
	    active: boolean;
	    lowStock: boolean;
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new MaterialDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sku = source["sku"];
	        this.category = source["category"];
	        this.purchaseUnit = source["purchaseUnit"];
	        this.consumptionUnit = source["consumptionUnit"];
	        this.conversionFactor = source["conversionFactor"];
	        this.physicalStock = source["physicalStock"];
	        this.reorderLevel = source["reorderLevel"];
	        this.averageUnitCostRial = source["averageUnitCostRial"];
	        this.preferredSupplier = source["preferredSupplier"];
	        this.notes = source["notes"];
	        this.active = source["active"];
	        this.lowStock = source["lowStock"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class MaterialInput {
	    name: string;
	    sku: string;
	    category: string;
	    purchaseUnit: string;
	    consumptionUnit: string;
	    conversionFactor: string;
	    physicalStock: string;
	    reorderLevel: string;
	    averageUnitCostRial: number;
	    preferredSupplier: string;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new MaterialInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.sku = source["sku"];
	        this.category = source["category"];
	        this.purchaseUnit = source["purchaseUnit"];
	        this.consumptionUnit = source["consumptionUnit"];
	        this.conversionFactor = source["conversionFactor"];
	        this.physicalStock = source["physicalStock"];
	        this.reorderLevel = source["reorderLevel"];
	        this.averageUnitCostRial = source["averageUnitCostRial"];
	        this.preferredSupplier = source["preferredSupplier"];
	        this.notes = source["notes"];
	    }
	}
	export class OrderItemDTO {
	    id: string;
	    position: number;
	    serviceId: string;
	    serviceName: string;
	    serviceCode: string;
	    quantity: string;
	    quantityUnit: string;
	    resolvedParametersJson: string;
	    costBreakdownJson: string;
	    pricingSnapshotJson: string;
	    estimatedCostRial: number;
	    suggestedPriceRial: number;
	    sellingPriceRial: number;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new OrderItemDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.position = source["position"];
	        this.serviceId = source["serviceId"];
	        this.serviceName = source["serviceName"];
	        this.serviceCode = source["serviceCode"];
	        this.quantity = source["quantity"];
	        this.quantityUnit = source["quantityUnit"];
	        this.resolvedParametersJson = source["resolvedParametersJson"];
	        this.costBreakdownJson = source["costBreakdownJson"];
	        this.pricingSnapshotJson = source["pricingSnapshotJson"];
	        this.estimatedCostRial = source["estimatedCostRial"];
	        this.suggestedPriceRial = source["suggestedPriceRial"];
	        this.sellingPriceRial = source["sellingPriceRial"];
	        this.notes = source["notes"];
	    }
	}
	export class OrderDTO {
	    id: string;
	    orderNumber: string;
	    customerId: string;
	    customerName: string;
	    customerPhone: string;
	    notes: string;
	    createdAt: string;
	    updatedAt: string;
	    promisedAt?: string;
	    priority: string;
	    commercialStatus: string;
	    fulfillmentStatus: string;
	    paymentStatus: string;
	    subtotalRial: number;
	    discountRial: number;
	    totalRial: number;
	    estimatedCostRial: number;
	    quoteId: string;
	    items: OrderItemDTO[];

	    static createFrom(source: any = {}) {
	        return new OrderDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.orderNumber = source["orderNumber"];
	        this.customerId = source["customerId"];
	        this.customerName = source["customerName"];
	        this.customerPhone = source["customerPhone"];
	        this.notes = source["notes"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.promisedAt = source["promisedAt"];
	        this.priority = source["priority"];
	        this.commercialStatus = source["commercialStatus"];
	        this.fulfillmentStatus = source["fulfillmentStatus"];
	        this.paymentStatus = source["paymentStatus"];
	        this.subtotalRial = source["subtotalRial"];
	        this.discountRial = source["discountRial"];
	        this.totalRial = source["totalRial"];
	        this.estimatedCostRial = source["estimatedCostRial"];
	        this.quoteId = source["quoteId"];
	        this.items = this.convertValues(source["items"], OrderItemDTO);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OrderInput {
	    customerId: string;
	    promisedAt?: string;
	    priority: string;
	    notes: string;
	    discountRial: number;

	    static createFrom(source: any = {}) {
	        return new OrderInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.customerId = source["customerId"];
	        this.promisedAt = source["promisedAt"];
	        this.priority = source["priority"];
	        this.notes = source["notes"];
	        this.discountRial = source["discountRial"];
	    }
	}

	export class OrderItemInput {
	    serviceId: string;
	    parameters: Record<string, string>;
	    manualCosts: Record<string, number>;
	    sellingPriceOverrideRial?: number;
	    quantity: string;
	    quantityUnit: string;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new OrderItemInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceId = source["serviceId"];
	        this.parameters = source["parameters"];
	        this.manualCosts = source["manualCosts"];
	        this.sellingPriceOverrideRial = source["sellingPriceOverrideRial"];
	        this.quantity = source["quantity"];
	        this.quantityUnit = source["quantityUnit"];
	        this.notes = source["notes"];
	    }
	}
	export class PricingComponentDTO {
	    id: string;
	    name: string;
	    type: string;
	    enabled: boolean;
	    usageQuantity: string;
	    rateRial: number;
	    percentage: string;
	    amountRial: number;
	    explanation: string;

	    static createFrom(source: any = {}) {
	        return new PricingComponentDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	        this.usageQuantity = source["usageQuantity"];
	        this.rateRial = source["rateRial"];
	        this.percentage = source["percentage"];
	        this.amountRial = source["amountRial"];
	        this.explanation = source["explanation"];
	    }
	}
	export class ResolvedParameterDTO {
	    key: string;
	    label: string;
	    type: string;
	    value: string;
	    quantity: string;
	    materialId: string;
	    unit: string;

	    static createFrom(source: any = {}) {
	        return new ResolvedParameterDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.value = source["value"];
	        this.quantity = source["quantity"];
	        this.materialId = source["materialId"];
	        this.unit = source["unit"];
	    }
	}
	export class PricingDTO {
	    serviceId: string;
	    serviceName: string;
	    serviceCode: string;
	    parameters: ResolvedParameterDTO[];
	    components: PricingComponentDTO[];
	    estimatedCostRial: number;
	    suggestedSellingPriceRial: number;
	    effectiveSellingPriceRial: number;
	    profitRial: number;
	    marginPercentage: string;
	    warnings: string[];
	    belowCost: boolean;

	    static createFrom(source: any = {}) {
	        return new PricingDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceId = source["serviceId"];
	        this.serviceName = source["serviceName"];
	        this.serviceCode = source["serviceCode"];
	        this.parameters = this.convertValues(source["parameters"], ResolvedParameterDTO);
	        this.components = this.convertValues(source["components"], PricingComponentDTO);
	        this.estimatedCostRial = source["estimatedCostRial"];
	        this.suggestedSellingPriceRial = source["suggestedSellingPriceRial"];
	        this.effectiveSellingPriceRial = source["effectiveSellingPriceRial"];
	        this.profitRial = source["profitRial"];
	        this.marginPercentage = source["marginPercentage"];
	        this.warnings = source["warnings"];
	        this.belowCost = source["belowCost"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PricingRequest {
	    serviceId: string;
	    parameters: Record<string, string>;
	    manualCosts: Record<string, number>;
	    sellingPriceOverrideRial?: number;

	    static createFrom(source: any = {}) {
	        return new PricingRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceId = source["serviceId"];
	        this.parameters = source["parameters"];
	        this.manualCosts = source["manualCosts"];
	        this.sellingPriceOverrideRial = source["sellingPriceOverrideRial"];
	    }
	}
	export class PricingTierDTO {
	    position: number;
	    minimumQuantity: string;
	    priceRial: number;

	    static createFrom(source: any = {}) {
	        return new PricingTierDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.position = source["position"];
	        this.minimumQuantity = source["minimumQuantity"];
	        this.priceRial = source["priceRial"];
	    }
	}
	export class PricingRuleDTO {
	    id: string;
	    type: string;
	    fixedPriceRial: number;
	    markupPercentage: string;
	    fixedMarginRial: number;
	    perUnitRateRial: number;
	    parameterKey: string;
	    tiers: PricingTierDTO[];

	    static createFrom(source: any = {}) {
	        return new PricingRuleDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.fixedPriceRial = source["fixedPriceRial"];
	        this.markupPercentage = source["markupPercentage"];
	        this.fixedMarginRial = source["fixedMarginRial"];
	        this.perUnitRateRial = source["perUnitRateRial"];
	        this.parameterKey = source["parameterKey"];
	        this.tiers = this.convertValues(source["tiers"], PricingTierDTO);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PricingTierInput {
	    position: number;
	    minimumQuantity: string;
	    priceRial: number;

	    static createFrom(source: any = {}) {
	        return new PricingTierInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.position = source["position"];
	        this.minimumQuantity = source["minimumQuantity"];
	        this.priceRial = source["priceRial"];
	    }
	}
	export class PricingRuleInput {
	    id: string;
	    type: string;
	    fixedPriceRial: number;
	    markupPercentage: string;
	    fixedMarginRial: number;
	    perUnitRateRial: number;
	    parameterKey: string;
	    tiers: PricingTierInput[];

	    static createFrom(source: any = {}) {
	        return new PricingRuleInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.fixedPriceRial = source["fixedPriceRial"];
	        this.markupPercentage = source["markupPercentage"];
	        this.fixedMarginRial = source["fixedMarginRial"];
	        this.perUnitRateRial = source["perUnitRateRial"];
	        this.parameterKey = source["parameterKey"];
	        this.tiers = this.convertValues(source["tiers"], PricingTierInput);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}


	export class ProofDTO {
	    id: string;
	    ownerType: string;
	    ownerId: string;
	    attachmentId: string;
	    status: string;
	    versionLabel: string;
	    approverNote: string;
	    internalNote: string;
	    preparedAt?: string;
	    approvedAt?: string;
	    rejectedAt?: string;
	    createdAt?: string;

	    static createFrom(source: any = {}) {
	        return new ProofDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ownerType = source["ownerType"];
	        this.ownerId = source["ownerId"];
	        this.attachmentId = source["attachmentId"];
	        this.status = source["status"];
	        this.versionLabel = source["versionLabel"];
	        this.approverNote = source["approverNote"];
	        this.internalNote = source["internalNote"];
	        this.preparedAt = source["preparedAt"];
	        this.approvedAt = source["approvedAt"];
	        this.rejectedAt = source["rejectedAt"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class QuoteItemDTO {
	    id: string;
	    position: number;
	    serviceId: string;
	    serviceName: string;
	    serviceCode: string;
	    quantity: string;
	    quantityUnit: string;
	    resolvedParametersJson: string;
	    costBreakdownJson: string;
	    pricingSnapshotJson: string;
	    estimatedCostRial: number;
	    suggestedPriceRial: number;
	    sellingPriceRial: number;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new QuoteItemDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.position = source["position"];
	        this.serviceId = source["serviceId"];
	        this.serviceName = source["serviceName"];
	        this.serviceCode = source["serviceCode"];
	        this.quantity = source["quantity"];
	        this.quantityUnit = source["quantityUnit"];
	        this.resolvedParametersJson = source["resolvedParametersJson"];
	        this.costBreakdownJson = source["costBreakdownJson"];
	        this.pricingSnapshotJson = source["pricingSnapshotJson"];
	        this.estimatedCostRial = source["estimatedCostRial"];
	        this.suggestedPriceRial = source["suggestedPriceRial"];
	        this.sellingPriceRial = source["sellingPriceRial"];
	        this.notes = source["notes"];
	    }
	}
	export class QuoteDTO {
	    id: string;
	    quoteNumber: string;
	    customerId: string;
	    customerName: string;
	    customerPhone: string;
	    notes: string;
	    createdAt: string;
	    updatedAt: string;
	    expiryDate?: string;
	    status: string;
	    subtotalRial: number;
	    discountRial: number;
	    totalRial: number;
	    estimatedCostRial: number;
	    convertedOrderId: string;
	    items: QuoteItemDTO[];

	    static createFrom(source: any = {}) {
	        return new QuoteDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.quoteNumber = source["quoteNumber"];
	        this.customerId = source["customerId"];
	        this.customerName = source["customerName"];
	        this.customerPhone = source["customerPhone"];
	        this.notes = source["notes"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.expiryDate = source["expiryDate"];
	        this.status = source["status"];
	        this.subtotalRial = source["subtotalRial"];
	        this.discountRial = source["discountRial"];
	        this.totalRial = source["totalRial"];
	        this.estimatedCostRial = source["estimatedCostRial"];
	        this.convertedOrderId = source["convertedOrderId"];
	        this.items = this.convertValues(source["items"], QuoteItemDTO);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QuoteInput {
	    customerId: string;
	    expiryDate?: string;
	    notes: string;
	    discountRial: number;

	    static createFrom(source: any = {}) {
	        return new QuoteInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.customerId = source["customerId"];
	        this.expiryDate = source["expiryDate"];
	        this.notes = source["notes"];
	        this.discountRial = source["discountRial"];
	    }
	}


	export class ServiceCostComponentDTO {
	    id: string;
	    name: string;
	    type: string;
	    referenceId: string;
	    usageMode: string;
	    parameterKey: string;
	    usageQuantity: string;
	    multiplier: string;
	    rateRial: number;
	    percentage: string;
	    rateBasis: string;
	    enabled: boolean;
	    position: number;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new ServiceCostComponentDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.referenceId = source["referenceId"];
	        this.usageMode = source["usageMode"];
	        this.parameterKey = source["parameterKey"];
	        this.usageQuantity = source["usageQuantity"];
	        this.multiplier = source["multiplier"];
	        this.rateRial = source["rateRial"];
	        this.percentage = source["percentage"];
	        this.rateBasis = source["rateBasis"];
	        this.enabled = source["enabled"];
	        this.position = source["position"];
	        this.notes = source["notes"];
	    }
	}
	export class ServiceCostComponentInput {
	    id: string;
	    name: string;
	    type: string;
	    referenceId: string;
	    usageMode: string;
	    parameterKey: string;
	    usageQuantity: string;
	    multiplier: string;
	    rateRial: number;
	    percentage: string;
	    rateBasis: string;
	    enabled: boolean;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new ServiceCostComponentInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.referenceId = source["referenceId"];
	        this.usageMode = source["usageMode"];
	        this.parameterKey = source["parameterKey"];
	        this.usageQuantity = source["usageQuantity"];
	        this.multiplier = source["multiplier"];
	        this.rateRial = source["rateRial"];
	        this.percentage = source["percentage"];
	        this.rateBasis = source["rateBasis"];
	        this.enabled = source["enabled"];
	        this.notes = source["notes"];
	    }
	}
	export class ServiceParameterDTO {
	    id: string;
	    key: string;
	    label: string;
	    type: string;
	    required: boolean;
	    position: number;
	    defaultValue: string;
	    options: string[];
	    minValue?: string;
	    maxValue?: string;
	    unit: string;
	    active: boolean;

	    static createFrom(source: any = {}) {
	        return new ServiceParameterDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.required = source["required"];
	        this.position = source["position"];
	        this.defaultValue = source["defaultValue"];
	        this.options = source["options"];
	        this.minValue = source["minValue"];
	        this.maxValue = source["maxValue"];
	        this.unit = source["unit"];
	        this.active = source["active"];
	    }
	}
	export class ServiceDTO {
	    id: string;
	    name: string;
	    code: string;
	    category: string;
	    description: string;
	    active: boolean;
	    createdAt: string;
	    updatedAt: string;
	    parameters: ServiceParameterDTO[];
	    components: ServiceCostComponentDTO[];
	    pricingRule?: PricingRuleDTO;

	    static createFrom(source: any = {}) {
	        return new ServiceDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.code = source["code"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.active = source["active"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.parameters = this.convertValues(source["parameters"], ServiceParameterDTO);
	        this.components = this.convertValues(source["components"], ServiceCostComponentDTO);
	        this.pricingRule = this.convertValues(source["pricingRule"], PricingRuleDTO);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ServiceParameterInput {
	    id: string;
	    key: string;
	    label: string;
	    type: string;
	    required: boolean;
	    defaultValue: string;
	    options: string[];
	    minValue?: string;
	    maxValue?: string;
	    unit: string;

	    static createFrom(source: any = {}) {
	        return new ServiceParameterInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.required = source["required"];
	        this.defaultValue = source["defaultValue"];
	        this.options = source["options"];
	        this.minValue = source["minValue"];
	        this.maxValue = source["maxValue"];
	        this.unit = source["unit"];
	    }
	}
	export class ServiceInput {
	    name: string;
	    code: string;
	    category: string;
	    description: string;
	    parameters: ServiceParameterInput[];
	    components: ServiceCostComponentInput[];
	    pricingRule?: PricingRuleInput;

	    static createFrom(source: any = {}) {
	        return new ServiceInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.code = source["code"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.parameters = this.convertValues(source["parameters"], ServiceParameterInput);
	        this.components = this.convertValues(source["components"], ServiceCostComponentInput);
	        this.pricingRule = this.convertValues(source["pricingRule"], PricingRuleInput);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}


}
