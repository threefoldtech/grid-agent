export namespace main {
	
	export class Step {
	    type: string;
	    progressText: string;
	    exportPrefix: string;
	    commandID: string;
	    content: string;
	    output: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new Step(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.progressText = source["progressText"];
	        this.exportPrefix = source["exportPrefix"];
	        this.commandID = source["commandID"];
	        this.content = source["content"];
	        this.output = source["output"];
	        this.error = source["error"];
	    }
	}
	export class Message {
	    role: string;
	    content: string;
	    timestamp: string;
	    requestID: string;
	    steps: Step[];
	    isCommand: boolean;
	    output: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = source["timestamp"];
	        this.requestID = source["requestID"];
	        this.steps = this.convertValues(source["steps"], Step);
	        this.isCommand = source["isCommand"];
	        this.output = source["output"];
	        this.error = source["error"];
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
	export class Profile {
	    id: string;
	    name: string;
	    instructions: string;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.instructions = source["instructions"];
	    }
	}
	export class Settings {
	    mnemonics: string;
	    network: string;
	    geminiApiKey: string;
	    model: string;
	    theme: string;
	    isConfigured: boolean;
	    profiles: Profile[];
	    activeProfileID: string;
	    enableExportSummary: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mnemonics = source["mnemonics"];
	        this.network = source["network"];
	        this.geminiApiKey = source["geminiApiKey"];
	        this.model = source["model"];
	        this.theme = source["theme"];
	        this.isConfigured = source["isConfigured"];
	        this.profiles = this.convertValues(source["profiles"], Profile);
	        this.activeProfileID = source["activeProfileID"];
	        this.enableExportSummary = source["enableExportSummary"];
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
	
	export class UpdateInfo {
	    updateAvailable: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    releaseURL: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.updateAvailable = source["updateAvailable"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseURL = source["releaseURL"];
	    }
	}
	export class VersionMismatchInfo {
	    hasMismatch: boolean;
	    appVersion: string;
	    tfcmdVersion: string;
	    releaseURL: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionMismatchInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasMismatch = source["hasMismatch"];
	        this.appVersion = source["appVersion"];
	        this.tfcmdVersion = source["tfcmdVersion"];
	        this.releaseURL = source["releaseURL"];
	    }
	}

}

