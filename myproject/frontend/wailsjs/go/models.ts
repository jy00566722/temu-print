export namespace main {
	
	export class LabelData {
	    serviceType: string;
	    phoneNumber: string;
	    itemNumber: string;
	    quantity: number;
	    totalItems: number;
	    warehouse: string;
	    shippingCrate: string;
	    currentTime: string;
	
	    static createFrom(source: any = {}) {
	        return new LabelData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceType = source["serviceType"];
	        this.phoneNumber = source["phoneNumber"];
	        this.itemNumber = source["itemNumber"];
	        this.quantity = source["quantity"];
	        this.totalItems = source["totalItems"];
	        this.warehouse = source["warehouse"];
	        this.shippingCrate = source["shippingCrate"];
	        this.currentTime = source["currentTime"];
	    }
	}

}

