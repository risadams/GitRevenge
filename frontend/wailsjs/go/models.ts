export namespace app {
	
	export class Environment {
	    gitPath: string;
	    gitVersion: string;
	    lfsVersion: string;
	    problem: string;
	    minVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gitPath = source["gitPath"];
	        this.gitVersion = source["gitVersion"];
	        this.lfsVersion = source["lfsVersion"];
	        this.problem = source["problem"];
	        this.minVersion = source["minVersion"];
	    }
	}
	export class FailedRepo {
	    path: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new FailedRepo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.error = source["error"];
	    }
	}
	export class RecentRepo {
	    path: string;
	    name: string;
	    // Go type: time
	    lastOpened: any;
	    exists: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RecentRepo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.lastOpened = this.convertValues(source["lastOpened"], null);
	        this.exists = source["exists"];
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
	export class RepoTab {
	    path: string;
	    name: string;
	    head: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoTab(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.head = source["head"];
	    }
	}
	export class SessionState {
	    tabs: RepoTab[];
	    active: string;
	    failed: FailedRepo[];
	
	    static createFrom(source: any = {}) {
	        return new SessionState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tabs = this.convertValues(source["tabs"], RepoTab);
	        this.active = source["active"];
	        this.failed = this.convertValues(source["failed"], FailedRepo);
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

