export namespace main {
	
	export class AccountDTO {
	    id: string;
	    code: string;
	    name: string;
	    type: string;
	    active: boolean;
	    system: boolean;
	    balanceRial: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.active = source["active"];
	        this.system = source["system"];
	        this.balanceRial = source["balanceRial"];
	    }
	}
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
	export class CustomerFinancialDTO {
	    customerId: string;
	    receivableRial: number;
	    creditRial: number;
	
	    static createFrom(source: any = {}) {
	        return new CustomerFinancialDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.customerId = source["customerId"];
	        this.receivableRial = source["receivableRial"];
	        this.creditRial = source["creditRial"];
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
	export class ExpenseDTO {
	    id: string;
	    expenseNumber: string;
	    expenseDate: string;
	    categoryAccountId: string;
	    payee: string;
	    supplierId: string;
	    description: string;
	    amountRial: number;
	    paymentMethod: string;
	    financialAccountId: string;
	    notes: string;
	    status: string;
	    journalEntryId: string;
	    idempotencyKey: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ExpenseDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.expenseNumber = source["expenseNumber"];
	        this.expenseDate = source["expenseDate"];
	        this.categoryAccountId = source["categoryAccountId"];
	        this.payee = source["payee"];
	        this.supplierId = source["supplierId"];
	        this.description = source["description"];
	        this.amountRial = source["amountRial"];
	        this.paymentMethod = source["paymentMethod"];
	        this.financialAccountId = source["financialAccountId"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	        this.journalEntryId = source["journalEntryId"];
	        this.idempotencyKey = source["idempotencyKey"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ExpenseInputDTO {
	    id: string;
	    expenseDate: string;
	    categoryAccountId: string;
	    payee: string;
	    supplierId: string;
	    description: string;
	    amountRial: number;
	    paymentMethod: string;
	    financialAccountId: string;
	    notes: string;
	    idempotencyKey: string;
	
	    static createFrom(source: any = {}) {
	        return new ExpenseInputDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.expenseDate = source["expenseDate"];
	        this.categoryAccountId = source["categoryAccountId"];
	        this.payee = source["payee"];
	        this.supplierId = source["supplierId"];
	        this.description = source["description"];
	        this.amountRial = source["amountRial"];
	        this.paymentMethod = source["paymentMethod"];
	        this.financialAccountId = source["financialAccountId"];
	        this.notes = source["notes"];
	        this.idempotencyKey = source["idempotencyKey"];
	    }
	}
	export class FinancialAccountDTO {
	    id: string;
	    name: string;
	    type: string;
	    ledgerAccountId: string;
	    details: string;
	    active: boolean;
	    balanceRial: number;
	
	    static createFrom(source: any = {}) {
	        return new FinancialAccountDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.ledgerAccountId = source["ledgerAccountId"];
	        this.details = source["details"];
	        this.active = source["active"];
	        this.balanceRial = source["balanceRial"];
	    }
	}
	export class InventoryMovementDTO {
	    id: string;
	    materialId: string;
	    occurredAt: string;
	    movementType: string;
	    quantityDelta: string;
	    unitCostRial: number;
	    totalCostRial: number;
	    referenceType: string;
	    referenceId: string;
	    note: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new InventoryMovementDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.materialId = source["materialId"];
	        this.occurredAt = source["occurredAt"];
	        this.movementType = source["movementType"];
	        this.quantityDelta = source["quantityDelta"];
	        this.unitCostRial = source["unitCostRial"];
	        this.totalCostRial = source["totalCostRial"];
	        this.referenceType = source["referenceType"];
	        this.referenceId = source["referenceId"];
	        this.note = source["note"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class InventoryReservationDTO {
	    id: string;
	    materialId: string;
	    orderId: string;
	    orderItemId: string;
	    productionJobId: string;
	    quantity: string;
	    status: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new InventoryReservationDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.materialId = source["materialId"];
	        this.orderId = source["orderId"];
	        this.orderItemId = source["orderItemId"];
	        this.productionJobId = source["productionJobId"];
	        this.quantity = source["quantity"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class InventoryReservationInput {
	    materialId: string;
	    orderId: string;
	    orderItemId: string;
	    productionJobId: string;
	    quantity: string;
	
	    static createFrom(source: any = {}) {
	        return new InventoryReservationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.materialId = source["materialId"];
	        this.orderId = source["orderId"];
	        this.orderItemId = source["orderItemId"];
	        this.productionJobId = source["productionJobId"];
	        this.quantity = source["quantity"];
	    }
	}
	export class InvoiceItemDTO {
	    id: string;
	    orderItemId: string;
	    description: string;
	    serviceId: string;
	    quantityUnit: string;
	    notes: string;
	    position: number;
	    quantity: string;
	    unitPriceRial: number;
	    lineTotalRial: number;
	
	    static createFrom(source: any = {}) {
	        return new InvoiceItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.orderItemId = source["orderItemId"];
	        this.description = source["description"];
	        this.serviceId = source["serviceId"];
	        this.quantityUnit = source["quantityUnit"];
	        this.notes = source["notes"];
	        this.position = source["position"];
	        this.quantity = source["quantity"];
	        this.unitPriceRial = source["unitPriceRial"];
	        this.lineTotalRial = source["lineTotalRial"];
	    }
	}
	export class InvoiceDTO {
	    id: string;
	    invoiceNumber: string;
	    customerId: string;
	    customerName: string;
	    customerPhone: string;
	    orderId: string;
	    issueDate: string;
	    dueDate: string;
	    status: string;
	    notes: string;
	    subtotalRial: number;
	    discountRial: number;
	    totalRial: number;
	    paidRial: number;
	    remainingRial: number;
	    accountingJournalEntryId: string;
	    cogsJournalEntryId: string;
	    createdAt: string;
	    updatedAt: string;
	    items: InvoiceItemDTO[];
	
	    static createFrom(source: any = {}) {
	        return new InvoiceDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoiceNumber = source["invoiceNumber"];
	        this.customerId = source["customerId"];
	        this.customerName = source["customerName"];
	        this.customerPhone = source["customerPhone"];
	        this.orderId = source["orderId"];
	        this.issueDate = source["issueDate"];
	        this.dueDate = source["dueDate"];
	        this.status = source["status"];
	        this.notes = source["notes"];
	        this.subtotalRial = source["subtotalRial"];
	        this.discountRial = source["discountRial"];
	        this.totalRial = source["totalRial"];
	        this.paidRial = source["paidRial"];
	        this.remainingRial = source["remainingRial"];
	        this.accountingJournalEntryId = source["accountingJournalEntryId"];
	        this.cogsJournalEntryId = source["cogsJournalEntryId"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.items = this.convertValues(source["items"], InvoiceItemDTO);
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
	
	export class JournalLineDTO {
	    id: string;
	    accountId: string;
	    partyType: string;
	    partyId: string;
	    memo: string;
	    position: number;
	    debitRial: number;
	    creditRial: number;
	
	    static createFrom(source: any = {}) {
	        return new JournalLineDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.accountId = source["accountId"];
	        this.partyType = source["partyType"];
	        this.partyId = source["partyId"];
	        this.memo = source["memo"];
	        this.position = source["position"];
	        this.debitRial = source["debitRial"];
	        this.creditRial = source["creditRial"];
	    }
	}
	export class JournalEntryDTO {
	    id: string;
	    entryNumber: string;
	    description: string;
	    sourceType: string;
	    sourceId: string;
	    reversalOfId: string;
	    postedAt: string;
	    createdAt: string;
	    lines: JournalLineDTO[];
	
	    static createFrom(source: any = {}) {
	        return new JournalEntryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.entryNumber = source["entryNumber"];
	        this.description = source["description"];
	        this.sourceType = source["sourceType"];
	        this.sourceId = source["sourceId"];
	        this.reversalOfId = source["reversalOfId"];
	        this.postedAt = source["postedAt"];
	        this.createdAt = source["createdAt"];
	        this.lines = this.convertValues(source["lines"], JournalLineDTO);
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
	    reservedStock: string;
	    availableStock: string;
	    reorderLevel: string;
	    averageUnitCostRial: number;
	    inventoryValueRial: number;
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
	        this.reservedStock = source["reservedStock"];
	        this.availableStock = source["availableStock"];
	        this.reorderLevel = source["reorderLevel"];
	        this.averageUnitCostRial = source["averageUnitCostRial"];
	        this.inventoryValueRial = source["inventoryValueRial"];
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
	    paidRial: number;
	    remainingRial: number;
	    invoiceId: string;
	    invoiceStatus: string;
	    invoicedTotalRial: number;
	    quoteId: string;
	    productionJobCount: number;
	    completedProductionJobs: number;
	    inProgressProductionJobs: number;
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
	        this.paidRial = source["paidRial"];
	        this.remainingRial = source["remainingRial"];
	        this.invoiceId = source["invoiceId"];
	        this.invoiceStatus = source["invoiceStatus"];
	        this.invoicedTotalRial = source["invoicedTotalRial"];
	        this.quoteId = source["quoteId"];
	        this.productionJobCount = source["productionJobCount"];
	        this.completedProductionJobs = source["completedProductionJobs"];
	        this.inProgressProductionJobs = source["inProgressProductionJobs"];
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
	export class OutsourceInput {
	    supplierId: string;
	    description: string;
	    sentAt: string;
	    expectedReturnAt: string;
	    receivedAt: string;
	    notes: string;
	    quotedCostRial: number;
	    actualCostRial: number;
	
	    static createFrom(source: any = {}) {
	        return new OutsourceInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supplierId = source["supplierId"];
	        this.description = source["description"];
	        this.sentAt = source["sentAt"];
	        this.expectedReturnAt = source["expectedReturnAt"];
	        this.receivedAt = source["receivedAt"];
	        this.notes = source["notes"];
	        this.quotedCostRial = source["quotedCostRial"];
	        this.actualCostRial = source["actualCostRial"];
	    }
	}
	export class PaymentAllocationDTO {
	    id: string;
	    targetType: string;
	    targetId: string;
	    position: number;
	    amountRial: number;
	    reversed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PaymentAllocationDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.targetType = source["targetType"];
	        this.targetId = source["targetId"];
	        this.position = source["position"];
	        this.amountRial = source["amountRial"];
	        this.reversed = source["reversed"];
	    }
	}
	export class PaymentAllocationInput {
	    targetType: string;
	    targetId: string;
	    amountRial: number;
	
	    static createFrom(source: any = {}) {
	        return new PaymentAllocationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.targetType = source["targetType"];
	        this.targetId = source["targetId"];
	        this.amountRial = source["amountRial"];
	    }
	}
	export class PaymentDTO {
	    id: string;
	    paymentNumber: string;
	    direction: string;
	    method: string;
	    financialAccountId: string;
	    customerId: string;
	    supplierId: string;
	    reference: string;
	    notes: string;
	    status: string;
	    journalEntryId: string;
	    postedAt: string;
	    createdAt: string;
	    amountRial: number;
	    allocations: PaymentAllocationDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PaymentDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.paymentNumber = source["paymentNumber"];
	        this.direction = source["direction"];
	        this.method = source["method"];
	        this.financialAccountId = source["financialAccountId"];
	        this.customerId = source["customerId"];
	        this.supplierId = source["supplierId"];
	        this.reference = source["reference"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	        this.journalEntryId = source["journalEntryId"];
	        this.postedAt = source["postedAt"];
	        this.createdAt = source["createdAt"];
	        this.amountRial = source["amountRial"];
	        this.allocations = this.convertValues(source["allocations"], PaymentAllocationDTO);
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
	export class PaymentInputDTO {
	    id: string;
	    direction: string;
	    method: string;
	    financialAccountId: string;
	    customerId: string;
	    supplierId: string;
	    reference: string;
	    notes: string;
	    postedAt: string;
	    idempotencyKey: string;
	    amountRial: number;
	    allocations: PaymentAllocationInput[];
	
	    static createFrom(source: any = {}) {
	        return new PaymentInputDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.direction = source["direction"];
	        this.method = source["method"];
	        this.financialAccountId = source["financialAccountId"];
	        this.customerId = source["customerId"];
	        this.supplierId = source["supplierId"];
	        this.reference = source["reference"];
	        this.notes = source["notes"];
	        this.postedAt = source["postedAt"];
	        this.idempotencyKey = source["idempotencyKey"];
	        this.amountRial = source["amountRial"];
	        this.allocations = this.convertValues(source["allocations"], PaymentAllocationInput);
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
	
	
	export class ProductionConsumptionDTO {
	    id: string;
	    productionJobId: string;
	    materialId: string;
	    idempotencyKey: string;
	    consumedQuantity: string;
	    wasteQuantity: string;
	    notes: string;
	    createdAt: string;
	    unitCostRial: number;
	    materialCostRial: number;
	    wasteCostRial: number;
	
	    static createFrom(source: any = {}) {
	        return new ProductionConsumptionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.productionJobId = source["productionJobId"];
	        this.materialId = source["materialId"];
	        this.idempotencyKey = source["idempotencyKey"];
	        this.consumedQuantity = source["consumedQuantity"];
	        this.wasteQuantity = source["wasteQuantity"];
	        this.notes = source["notes"];
	        this.createdAt = source["createdAt"];
	        this.unitCostRial = source["unitCostRial"];
	        this.materialCostRial = source["materialCostRial"];
	        this.wasteCostRial = source["wasteCostRial"];
	    }
	}
	export class ProductionConsumptionInput {
	    materialId: string;
	    consumedQuantity: string;
	    wasteQuantity: string;
	    idempotencyKey: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new ProductionConsumptionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.materialId = source["materialId"];
	        this.consumedQuantity = source["consumedQuantity"];
	        this.wasteQuantity = source["wasteQuantity"];
	        this.idempotencyKey = source["idempotencyKey"];
	        this.notes = source["notes"];
	    }
	}
	export class ProductionJobDTO {
	    id: string;
	    jobNumber: string;
	    orderId: string;
	    orderItemId: string;
	    serviceName: string;
	    quantity: string;
	    quantityUnit: string;
	    assignedMachineId: string;
	    status: string;
	    priority: string;
	    notes: string;
	    plannedAt: string;
	    startedAt: string;
	    completedAt: string;
	    createdAt: string;
	    estimatedCostRial: number;
	    actualMaterialCostRial: number;
	    actualWasteCostRial: number;
	    actualOutsourcedCostRial: number;
	    actualTotalCostRial: number;
	    outsourceQuotedCostRial: number;
	    outsourceSupplierId: string;
	    outsourceDescription: string;
	    outsourceSentAt: string;
	    outsourceExpectedReturnAt: string;
	    outsourceReceivedAt: string;
	    outsourceNotes: string;
	
	    static createFrom(source: any = {}) {
	        return new ProductionJobDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.jobNumber = source["jobNumber"];
	        this.orderId = source["orderId"];
	        this.orderItemId = source["orderItemId"];
	        this.serviceName = source["serviceName"];
	        this.quantity = source["quantity"];
	        this.quantityUnit = source["quantityUnit"];
	        this.assignedMachineId = source["assignedMachineId"];
	        this.status = source["status"];
	        this.priority = source["priority"];
	        this.notes = source["notes"];
	        this.plannedAt = source["plannedAt"];
	        this.startedAt = source["startedAt"];
	        this.completedAt = source["completedAt"];
	        this.createdAt = source["createdAt"];
	        this.estimatedCostRial = source["estimatedCostRial"];
	        this.actualMaterialCostRial = source["actualMaterialCostRial"];
	        this.actualWasteCostRial = source["actualWasteCostRial"];
	        this.actualOutsourcedCostRial = source["actualOutsourcedCostRial"];
	        this.actualTotalCostRial = source["actualTotalCostRial"];
	        this.outsourceQuotedCostRial = source["outsourceQuotedCostRial"];
	        this.outsourceSupplierId = source["outsourceSupplierId"];
	        this.outsourceDescription = source["outsourceDescription"];
	        this.outsourceSentAt = source["outsourceSentAt"];
	        this.outsourceExpectedReturnAt = source["outsourceExpectedReturnAt"];
	        this.outsourceReceivedAt = source["outsourceReceivedAt"];
	        this.outsourceNotes = source["outsourceNotes"];
	    }
	}
	export class ProductionJobInput {
	    orderId: string;
	    orderItemId: string;
	    quantity: string;
	    quantityUnit: string;
	    assignedMachineId: string;
	    priority: string;
	    notes: string;
	    plannedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProductionJobInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.orderId = source["orderId"];
	        this.orderItemId = source["orderItemId"];
	        this.quantity = source["quantity"];
	        this.quantityUnit = source["quantityUnit"];
	        this.assignedMachineId = source["assignedMachineId"];
	        this.priority = source["priority"];
	        this.notes = source["notes"];
	        this.plannedAt = source["plannedAt"];
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
	export class PurchaseItemDTO {
	    id: string;
	    position: number;
	    materialId: string;
	    materialName: string;
	    purchaseUnit: string;
	    consumptionUnit: string;
	    purchaseQuantity: string;
	    conversionFactor: string;
	    consumptionQuantity: string;
	    unitAcquisitionCostRial: number;
	    allocatedAdditionalCostRial: number;
	    landedUnitCostRial: number;
	    lineTotalRial: number;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new PurchaseItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.position = source["position"];
	        this.materialId = source["materialId"];
	        this.materialName = source["materialName"];
	        this.purchaseUnit = source["purchaseUnit"];
	        this.consumptionUnit = source["consumptionUnit"];
	        this.purchaseQuantity = source["purchaseQuantity"];
	        this.conversionFactor = source["conversionFactor"];
	        this.consumptionQuantity = source["consumptionQuantity"];
	        this.unitAcquisitionCostRial = source["unitAcquisitionCostRial"];
	        this.allocatedAdditionalCostRial = source["allocatedAdditionalCostRial"];
	        this.landedUnitCostRial = source["landedUnitCostRial"];
	        this.lineTotalRial = source["lineTotalRial"];
	        this.notes = source["notes"];
	    }
	}
	export class PurchaseDTO {
	    id: string;
	    purchaseNumber: string;
	    supplierId: string;
	    supplierName: string;
	    supplierInvoiceNumber: string;
	    purchaseDate: string;
	    status: string;
	    notes: string;
	    subtotalRial: number;
	    discountRial: number;
	    shippingRial: number;
	    taxRial: number;
	    additionalCostsRial: number;
	    totalRial: number;
	    paidRial: number;
	    remainingRial: number;
	    createdAt: string;
	    updatedAt: string;
	    items: PurchaseItemDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PurchaseDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.purchaseNumber = source["purchaseNumber"];
	        this.supplierId = source["supplierId"];
	        this.supplierName = source["supplierName"];
	        this.supplierInvoiceNumber = source["supplierInvoiceNumber"];
	        this.purchaseDate = source["purchaseDate"];
	        this.status = source["status"];
	        this.notes = source["notes"];
	        this.subtotalRial = source["subtotalRial"];
	        this.discountRial = source["discountRial"];
	        this.shippingRial = source["shippingRial"];
	        this.taxRial = source["taxRial"];
	        this.additionalCostsRial = source["additionalCostsRial"];
	        this.totalRial = source["totalRial"];
	        this.paidRial = source["paidRial"];
	        this.remainingRial = source["remainingRial"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.items = this.convertValues(source["items"], PurchaseItemDTO);
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
	export class PurchaseInput {
	    supplierId: string;
	    purchaseDate: string;
	    supplierInvoiceNumber: string;
	    notes: string;
	    discountRial: number;
	    shippingRial: number;
	    taxRial: number;
	    additionalCostsRial: number;
	
	    static createFrom(source: any = {}) {
	        return new PurchaseInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supplierId = source["supplierId"];
	        this.purchaseDate = source["purchaseDate"];
	        this.supplierInvoiceNumber = source["supplierInvoiceNumber"];
	        this.notes = source["notes"];
	        this.discountRial = source["discountRial"];
	        this.shippingRial = source["shippingRial"];
	        this.taxRial = source["taxRial"];
	        this.additionalCostsRial = source["additionalCostsRial"];
	    }
	}
	
	export class PurchaseItemInput {
	    materialId: string;
	    purchaseQuantity: string;
	    unitAcquisitionCostRial: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new PurchaseItemInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.materialId = source["materialId"];
	        this.purchaseQuantity = source["purchaseQuantity"];
	        this.unitAcquisitionCostRial = source["unitAcquisitionCostRial"];
	        this.notes = source["notes"];
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
	
	
	export class SupplierDTO {
	    id: string;
	    name: string;
	    code: string;
	    phone: string;
	    email: string;
	    address: string;
	    notes: string;
	    active: boolean;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SupplierDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.code = source["code"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.address = source["address"];
	        this.notes = source["notes"];
	        this.active = source["active"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class SupplierInput {
	    name: string;
	    code: string;
	    phone: string;
	    email: string;
	    address: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new SupplierInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.code = source["code"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.address = source["address"];
	        this.notes = source["notes"];
	    }
	}
	export class TransferDTO {
	    id: string;
	    transferNumber: string;
	    sourceFinancialAccountId: string;
	    destinationFinancialAccountId: string;
	    amountRial: number;
	    transferDate: string;
	    reference: string;
	    notes: string;
	    status: string;
	    journalEntryId: string;
	    idempotencyKey: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TransferDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.transferNumber = source["transferNumber"];
	        this.sourceFinancialAccountId = source["sourceFinancialAccountId"];
	        this.destinationFinancialAccountId = source["destinationFinancialAccountId"];
	        this.amountRial = source["amountRial"];
	        this.transferDate = source["transferDate"];
	        this.reference = source["reference"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	        this.journalEntryId = source["journalEntryId"];
	        this.idempotencyKey = source["idempotencyKey"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class TransferInputDTO {
	    id: string;
	    sourceFinancialAccountId: string;
	    destinationFinancialAccountId: string;
	    amountRial: number;
	    transferDate: string;
	    reference: string;
	    notes: string;
	    idempotencyKey: string;
	
	    static createFrom(source: any = {}) {
	        return new TransferInputDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sourceFinancialAccountId = source["sourceFinancialAccountId"];
	        this.destinationFinancialAccountId = source["destinationFinancialAccountId"];
	        this.amountRial = source["amountRial"];
	        this.transferDate = source["transferDate"];
	        this.reference = source["reference"];
	        this.notes = source["notes"];
	        this.idempotencyKey = source["idempotencyKey"];
	    }
	}
	export class CheckDTO {
	    id!:string; checkNumber!:string; direction!:string; bank!:string; branch!:string; accountDescriptor!:string; payerPayee!:string; customerId!:string; supplierId!:string; sourceType!:string; sourceId!:string; financialAccountId!:string; notes!:string; status!:string; issueDate!:string; dueDate!:string; createdAt!:string; updatedAt!:string; amountRial!:number;
	    static createFrom(source:any={}) { return new CheckDTO(source); }
	    constructor(source:any={}) { if ('string'===typeof source) source=JSON.parse(source); Object.assign(this,source); }
	}
	export class CheckInputDTO extends CheckDTO {}
	export class CheckEventDTO { id!:string; checkId!:string; fromStatus!:string; toStatus!:string; note!:string; journalEntryId!:string; occurredAt!:string; static createFrom(source:any={}){return new CheckEventDTO(source)} constructor(source:any={}){if('string'===typeof source)source=JSON.parse(source);Object.assign(this,source)} }
	export class LoanInstallmentInputDTO { id!:string; dueDate!:string; principalRial!:number; interestFeeRial!:number; static createFrom(source:any={}){return new LoanInstallmentInputDTO(source)} constructor(source:any={}){if('string'===typeof source)source=JSON.parse(source);Object.assign(this,source)} }
	export class LoanDTO { id!:string; loanNumber!:string; direction!:string; counterpartyName!:string; customerId!:string; supplierId!:string; startDate!:string; endDate!:string; status!:string; notes!:string; financialAccountId!:string; journalEntryId!:string; idempotencyKey!:string; createdAt!:string; updatedAt!:string; principalRial!:number; interestFeeRial!:number; paidPrincipalRial!:number; paidInterestRial!:number; remainingPrincipalRial!:number; remainingInterestRial!:number; overdueRial!:number; installments!:Array<LoanInstallmentDTO>; static createFrom(source:any={}){return new LoanDTO(source)} constructor(source:any={}){if('string'===typeof source)source=JSON.parse(source);Object.assign(this,source)} }
	export class LoanInputDTO extends LoanDTO {}
	export class LoanInstallmentDTO extends LoanInstallmentInputDTO { totalDueRial!:number; paidPrincipalRial!:number; paidInterestRial!:number; paidRial!:number; remainingRial!:number; overdueRial!:number; position!:number; status!:string }
	export class LoanPaymentAllocationInputDTO { installmentId!:string; principalRial!:number; interestRial!:number; static createFrom(source:any={}){return new LoanPaymentAllocationInputDTO(source)} constructor(source:any={}){if('string'===typeof source)source=JSON.parse(source);Object.assign(this,source)} }
	export class LoanPaymentAllocationDTO extends LoanPaymentAllocationInputDTO { id!:string; paymentId!:string; position!:number }
	export class LoanPaymentDTO { id!:string; paymentNumber!:string; loanId!:string; financialAccountId!:string; paidAt!:string; notes!:string; status!:string; journalEntryId!:string; idempotencyKey!:string; amountRial!:number; principalRial!:number; interestRial!:number; allocations!:Array<LoanPaymentAllocationDTO>; static createFrom(source:any={}){return new LoanPaymentDTO(source)} constructor(source:any={}){if('string'===typeof source)source=JSON.parse(source);Object.assign(this,source)} }
	export class LoanPaymentInputDTO extends LoanPaymentDTO {}

}
