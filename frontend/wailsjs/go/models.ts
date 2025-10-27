export namespace cfg {
	
	export enum UpdatePolicyType {
	    AUTOMATIC = "automatic",
	    PROMPT = "prompt",
	    DISABLED = "disabled",
	}

}

export namespace filter {
	
	export class List {
	    name: string;
	    type: string;
	    url: string;
	    enabled: boolean;
	    trusted: boolean;
	    locales: string[];
	
	    static createFrom(source: any = {}) {
	        return new List(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.enabled = source["enabled"];
	        this.trusted = source["trusted"];
	        this.locales = source["locales"];
	    }
	}

}

export namespace options {
	
	export class SecondInstanceData {
	    Args: string[];
	    WorkingDirectory: string;
	
	    static createFrom(source: any = {}) {
	        return new SecondInstanceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Args = source["Args"];
	        this.WorkingDirectory = source["WorkingDirectory"];
	    }
	}

}

