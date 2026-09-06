export namespace main {

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
	export class ServiceCostComponentDTO {
	    id: string; name: string; type: string; referenceId: string; usageMode: string; parameterKey: string; multiplier: string; rateRial: number; percentage: string; rateBasis: string; enabled: boolean; position: number; notes: string;
	    static createFrom(source: any = {}) { return new ServiceCostComponentDTO(source); }
	    constructor(source: any = {}) { if ('string' === typeof source) source = JSON.parse(source); this.id=source["id"]; this.name=source["name"]; this.type=source["type"]; this.referenceId=source["referenceId"]; this.usageMode=source["usageMode"]; this.parameterKey=source["parameterKey"]; this.multiplier=source["multiplier"]; this.rateRial=source["rateRial"]; this.percentage=source["percentage"]; this.rateBasis=source["rateBasis"]; this.enabled=source["enabled"]; this.position=source["position"]; this.notes=source["notes"]; }
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
	export class ServiceCostComponentInput {
	    id: string; name: string; type: string; referenceId: string; usageMode: string; parameterKey: string; multiplier: string; rateRial: number; percentage: string; rateBasis: string; enabled: boolean; notes: string;
	    static createFrom(source: any = {}) { return new ServiceCostComponentInput(source); }
	    constructor(source: any = {}) { if ('string' === typeof source) source = JSON.parse(source); this.id=source["id"]; this.name=source["name"]; this.type=source["type"]; this.referenceId=source["referenceId"]; this.usageMode=source["usageMode"]; this.parameterKey=source["parameterKey"]; this.multiplier=source["multiplier"]; this.rateRial=source["rateRial"]; this.percentage=source["percentage"]; this.rateBasis=source["rateBasis"]; this.enabled=source["enabled"]; this.notes=source["notes"]; }
	}
	export class MachineDTO {
	    id: string; name: string; code: string; category: string; rateBasis: string; rateRial: number; setupCostRial: number; notes: string; active: boolean; createdAt: string; updatedAt: string;
	    static createFrom(source: any = {}) { return new MachineDTO(source); }
	    constructor(source: any = {}) { if ('string' === typeof source) source = JSON.parse(source); this.id=source["id"]; this.name=source["name"]; this.code=source["code"]; this.category=source["category"]; this.rateBasis=source["rateBasis"]; this.rateRial=source["rateRial"]; this.setupCostRial=source["setupCostRial"]; this.notes=source["notes"]; this.active=source["active"]; this.createdAt=source["createdAt"]; this.updatedAt=source["updatedAt"]; }
	}
	export class MachineInput {
	    name: string; code: string; category: string; rateBasis: string; rateRial: number; setupCostRial: number; notes: string;
	    static createFrom(source: any = {}) { return new MachineInput(source); }
	    constructor(source: any = {}) { if ('string' === typeof source) source = JSON.parse(source); this.name=source["name"]; this.code=source["code"]; this.category=source["category"]; this.rateBasis=source["rateBasis"]; this.rateRial=source["rateRial"]; this.setupCostRial=source["setupCostRial"]; this.notes=source["notes"]; }
	}


}
