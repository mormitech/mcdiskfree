export namespace disk {
	
	export class Usage {
	    total: number;
	    free: number;
	
	    static createFrom(source: any = {}) {
	        return new Usage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.free = source["free"];
	    }
	}
	export class Volume {
	    name: string;
	    path: string;
	    total: number;
	    free: number;
	
	    static createFrom(source: any = {}) {
	        return new Volume(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.total = source["total"];
	        this.free = source["free"];
	    }
	}

}

export namespace main {
	
	export class Node {
	    name: string;
	    path: string;
	    size: number;
	    isDir: boolean;
	    items: number;
	    children?: Node[];
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.isDir = source["isDir"];
	        this.items = source["items"];
	        this.children = this.convertValues(source["children"], Node);
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
	export class TrashResult {
	    trashed: string[];
	    failed: string[];
	
	    static createFrom(source: any = {}) {
	        return new TrashResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trashed = source["trashed"];
	        this.failed = source["failed"];
	    }
	}

}

