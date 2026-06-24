export namespace app {
	
	export class CompletedJobResult {
	    code: number;
	    message: string;
	    jobs: task.CompletedJob[];
	
	    static createFrom(source: any = {}) {
	        return new CompletedJobResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.jobs = this.convertValues(source["jobs"], task.CompletedJob);
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
	export class PreviewResult {
	    code: number;
	    message: string;
	    index: number;
	    fileName: string;
	    sourceBase64: string;
	    targetBase64: string;
	    sourceSize: number;
	    targetSize: number;
	    ratio: string;
	
	    static createFrom(source: any = {}) {
	        return new PreviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.index = source["index"];
	        this.fileName = source["fileName"];
	        this.sourceBase64 = source["sourceBase64"];
	        this.targetBase64 = source["targetBase64"];
	        this.sourceSize = source["sourceSize"];
	        this.targetSize = source["targetSize"];
	        this.ratio = source["ratio"];
	    }
	}

}

export namespace dto {
	
	export class BasicResult {
	    code: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new BasicResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}
	export class ExportResult {
	    code: number;
	    message: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.path = source["path"];
	    }
	}
	export class FileInfo {
	    path: string;
	    name: string;
	    format: string;
	    size: number;
	    modTime: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.format = source["format"];
	        this.size = source["size"];
	        this.modTime = source["modTime"];
	    }
	}
	export class FilterInfo {
	    path: string;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new FilterInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.reason = source["reason"];
	    }
	}
	export class ScanResult {
	    code: number;
	    message: string;
	    totalFiles: number;
	    totalBytes: number;
	    supportedFiles: FileInfo[];
	    filteredFiles: FilterInfo[];
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.totalFiles = source["totalFiles"];
	        this.totalBytes = source["totalBytes"];
	        this.supportedFiles = this.convertValues(source["supportedFiles"], FileInfo);
	        this.filteredFiles = this.convertValues(source["filteredFiles"], FilterInfo);
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
	export class SelectDirResult {
	    code: number;
	    message: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new SelectDirResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.path = source["path"];
	    }
	}
	export class SelectPathsResult {
	    code: number;
	    message: string;
	    paths: string[];
	
	    static createFrom(source: any = {}) {
	        return new SelectPathsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.paths = source["paths"];
	    }
	}
	export class TaskStatusResult {
	    total: number;
	    done: number;
	    failed: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.done = source["done"];
	        this.failed = source["failed"];
	        this.status = source["status"];
	    }
	}

}

export namespace task {
	
	export class CompletedJob {
	    index: number;
	    fileName: string;
	    sourcePath: string;
	    targetPath: string;
	    oldSize: number;
	    newSize: number;
	    status: string;
	    message: string;
	    ratio: string;
	
	    static createFrom(source: any = {}) {
	        return new CompletedJob(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.fileName = source["fileName"];
	        this.sourcePath = source["sourcePath"];
	        this.targetPath = source["targetPath"];
	        this.oldSize = source["oldSize"];
	        this.newSize = source["newSize"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.ratio = source["ratio"];
	    }
	}

}

