export namespace main {
	
	export class Document {
	    path: string;
	    name: string;
	    html: string;
	    css: string;
	    customCss: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new Document(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.html = source["html"];
	        this.css = source["css"];
	        this.customCss = source["customCss"];
	        this.error = source["error"];
	    }
	}

}

