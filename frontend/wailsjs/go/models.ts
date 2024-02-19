export namespace cfg {
	
	export class FilterList {
	    name: string;
	    type: string;
	    url: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FilterList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.enabled = source["enabled"];
	    }
	}

}

