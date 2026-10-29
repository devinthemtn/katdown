export namespace main {
	
	export class CodeCSS {
	    dark: string;
	    light: string;
	
	    static createFrom(source: any = {}) {
	        return new CodeCSS(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dark = source["dark"];
	        this.light = source["light"];
	    }
	}
	export class Document {
	    path: string;
	    name: string;
	    html: string;
	    css: string;
	    customCss: boolean;
	    theme: string;
	    codeCss: CodeCSS;
	    error: string;
	    configError: string;
	
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
	        this.theme = source["theme"];
	        this.codeCss = this.convertValues(source["codeCss"], CodeCSS);
	        this.error = source["error"];
	        this.configError = source["configError"];
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

