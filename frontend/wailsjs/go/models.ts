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

}

