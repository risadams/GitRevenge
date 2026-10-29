export namespace app {
	
	export class CommitInfo {
	    commit: git.CommitDetail;
	    files: git.FileChange[];
	
	    static createFrom(source: any = {}) {
	        return new CommitInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.commit = this.convertValues(source["commit"], git.CommitDetail);
	        this.files = this.convertValues(source["files"], git.FileChange);
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
	export class GraphPage {
	    rows: graph.Row[];
	    refs: Record<string, Array<git.Ref>>;
	    total: number;
	    loading: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GraphPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = this.convertValues(source["rows"], graph.Row);
	        this.refs = this.convertValues(source["refs"], Array<git.Ref>, true);
	        this.total = source["total"];
	        this.loading = source["loading"];
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
	export class RepoState {
	    operation: string;
	    step: number;
	    total: number;
	    actions: string[];
	    conflicts: string[];
	
	    static createFrom(source: any = {}) {
	        return new RepoState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operation = source["operation"];
	        this.step = source["step"];
	        this.total = source["total"];
	        this.actions = source["actions"];
	        this.conflicts = source["conflicts"];
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
	export class WorkingTree {
	    status: git.Status;
	    signing: git.Signing;
	
	    static createFrom(source: any = {}) {
	        return new WorkingTree(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = this.convertValues(source["status"], git.Status);
	        this.signing = this.convertValues(source["signing"], git.Signing);
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

export namespace git {
	
	export class Signature {
	    name: string;
	    email: string;
	    time: number;
	    tz: string;
	
	    static createFrom(source: any = {}) {
	        return new Signature(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.email = source["email"];
	        this.time = source["time"];
	        this.tz = source["tz"];
	    }
	}
	export class CommitDetail {
	    id: string;
	    parents: string[];
	    author: Signature;
	    committer: Signature;
	    subject: string;
	    body: string;
	    signed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CommitDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.parents = source["parents"];
	        this.author = this.convertValues(source["author"], Signature);
	        this.committer = this.convertValues(source["committer"], Signature);
	        this.subject = source["subject"];
	        this.body = source["body"];
	        this.signed = source["signed"];
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
	export class Side {
	    exists: boolean;
	    text: string;
	    image?: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new Side(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exists = source["exists"];
	        this.text = source["text"];
	        this.image = source["image"];
	        this.size = source["size"];
	    }
	}
	export class ConflictFile {
	    path: string;
	    base: Side;
	    ours: Side;
	    theirs: Side;
	    merged: Side;
	    binary: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConflictFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.base = this.convertValues(source["base"], Side);
	        this.ours = this.convertValues(source["ours"], Side);
	        this.theirs = this.convertValues(source["theirs"], Side);
	        this.merged = this.convertValues(source["merged"], Side);
	        this.binary = source["binary"];
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
	export class FileChange {
	    status: string;
	    path: string;
	    oldPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new FileChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.path = source["path"];
	        this.oldPath = source["oldPath"];
	    }
	}
	export class FileContents {
	    path: string;
	    old: Side;
	    new: Side;
	    kind: string;
	    mime?: string;
	
	    static createFrom(source: any = {}) {
	        return new FileContents(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.old = this.convertValues(source["old"], Side);
	        this.new = this.convertValues(source["new"], Side);
	        this.kind = source["kind"];
	        this.mime = source["mime"];
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
	export class FileStatus {
	    path: string;
	    origPath?: string;
	    index: string;
	    workTree: string;
	    untracked: boolean;
	    conflicted: boolean;
	    submodule: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.origPath = source["origPath"];
	        this.index = source["index"];
	        this.workTree = source["workTree"];
	        this.untracked = source["untracked"];
	        this.conflicted = source["conflicted"];
	        this.submodule = source["submodule"];
	    }
	}
	export class Ref {
	    name: string;
	    short: string;
	    kind: string;
	    commit: string;
	    head: boolean;
	    upstream: string;
	    ahead: number;
	    behind: number;
	    gone: boolean;
	    subject?: string;
	
	    static createFrom(source: any = {}) {
	        return new Ref(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.short = source["short"];
	        this.kind = source["kind"];
	        this.commit = source["commit"];
	        this.head = source["head"];
	        this.upstream = source["upstream"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.gone = source["gone"];
	        this.subject = source["subject"];
	    }
	}
	export class Remote {
	    name: string;
	    fetchUrl: string;
	    pushUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new Remote(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.fetchUrl = source["fetchUrl"];
	        this.pushUrl = source["pushUrl"];
	    }
	}
	
	
	export class Signing {
	    enabled: boolean;
	    format: string;
	    key: string;
	
	    static createFrom(source: any = {}) {
	        return new Signing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.format = source["format"];
	        this.key = source["key"];
	    }
	}
	export class Status {
	    head: string;
	    oid: string;
	    upstream: string;
	    ahead: number;
	    behind: number;
	    files: FileStatus[];
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.head = source["head"];
	        this.oid = source["oid"];
	        this.upstream = source["upstream"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.files = this.convertValues(source["files"], FileStatus);
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

export namespace graph {
	
	export class Seg {
	    from: number;
	    to: number;
	    color: number;
	
	    static createFrom(source: any = {}) {
	        return new Seg(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.color = source["color"];
	    }
	}
	export class Row {
	    index: number;
	    col: number;
	    color: number;
	    up: Seg[];
	    down: Seg[];
	    width: number;
	    id: string;
	    parents: number;
	    author: string;
	    authorTime: number;
	    subject: string;
	
	    static createFrom(source: any = {}) {
	        return new Row(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.col = source["col"];
	        this.color = source["color"];
	        this.up = this.convertValues(source["up"], Seg);
	        this.down = this.convertValues(source["down"], Seg);
	        this.width = source["width"];
	        this.id = source["id"];
	        this.parents = source["parents"];
	        this.author = source["author"];
	        this.authorTime = source["authorTime"];
	        this.subject = source["subject"];
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
	export class SearchResult {
	    rows: number[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = source["rows"];
	        this.total = source["total"];
	    }
	}

}

