export namespace models {
	
	export class Language {
	    Tag: string;
	    Name: string;
	    Source: boolean;
	    Target: boolean;
	    Stable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Language(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Tag = source["Tag"];
	        this.Name = source["Name"];
	        this.Source = source["Source"];
	        this.Target = source["Target"];
	        this.Stable = source["Stable"];
	    }
	}

}

export namespace transport {
	
	export class TranslationDto {
	    Source: string;
	    Target: string;
	    DetectedSource: string;
	    Translation: string;
	    HasPrevious: boolean;
	    HasNext: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TranslationDto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Source = source["Source"];
	        this.Target = source["Target"];
	        this.DetectedSource = source["DetectedSource"];
	        this.Translation = source["Translation"];
	        this.HasPrevious = source["HasPrevious"];
	        this.HasNext = source["HasNext"];
	    }
	}
	export class LanguageDto {
	    PreferredSource: string;
	    PreferredTarget: string;
	    Detection?: models.Language;
	    SourceLanguages: models.Language[];
	    TargetLanguages: models.Language[];
	
	    static createFrom(source: any = {}) {
	        return new LanguageDto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PreferredSource = source["PreferredSource"];
	        this.PreferredTarget = source["PreferredTarget"];
	        this.Detection = this.convertValues(source["Detection"], models.Language);
	        this.SourceLanguages = this.convertValues(source["SourceLanguages"], models.Language);
	        this.TargetLanguages = this.convertValues(source["TargetLanguages"], models.Language);
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
	export class ProviderDto {
	    Current: string;
	    Providers: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProviderDto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Current = source["Current"];
	        this.Providers = source["Providers"];
	    }
	}
	export class FullDto {
	    Provider?: ProviderDto;
	    Languages?: LanguageDto;
	    Translation?: TranslationDto;
	
	    static createFrom(source: any = {}) {
	        return new FullDto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Provider = this.convertValues(source["Provider"], ProviderDto);
	        this.Languages = this.convertValues(source["Languages"], LanguageDto);
	        this.Translation = this.convertValues(source["Translation"], TranslationDto);
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

