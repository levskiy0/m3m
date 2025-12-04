package modules

// GetBaseTypeDefinitions returns TypeScript definitions for Monaco IntelliSense
func GetBaseTypeDefinitions() string {
	return `
// M3M Runtime API Type Definitions

// Logger module
declare const logger: {
    debug(...args: any[]): void;
    info(...args: any[]): void;
    warn(...args: any[]): void;
    error(...args: any[]): void;
};

// Console (alias for logger)
declare const console: {
    log(...args: any[]): void;
    info(...args: any[]): void;
    warn(...args: any[]): void;
    error(...args: any[]): void;
    debug(...args: any[]): void;
};

// Router module
interface RequestContext {
    method: string;
    path: string;
    params: { [key: string]: string };
    query: { [key: string]: string };
    headers: { [key: string]: string };
    body: any;
}

interface ResponseData {
    status: number;
    body: any;
    headers?: { [key: string]: string };
}

declare const router: {
    get(path: string, handler: (ctx: RequestContext) => ResponseData | any): void;
    post(path: string, handler: (ctx: RequestContext) => ResponseData | any): void;
    put(path: string, handler: (ctx: RequestContext) => ResponseData | any): void;
    delete(path: string, handler: (ctx: RequestContext) => ResponseData | any): void;
    response(status: number, body: any): ResponseData;
};

// Schedule module
declare const schedule: {
    daily(handler: () => void): void;
    hourly(handler: () => void): void;
    cron(expression: string, handler: () => void): void;
};

// Environment module
declare const env: {
    get(key: string): any;
};

// Storage module
declare const storage: {
    read(path: string): string;
    write(path: string, content: string): boolean;
    exists(path: string): boolean;
    delete(path: string): boolean;
    list(path: string): string[];
    mkdir(path: string): boolean;
};

// Database module
interface Collection {
    find(filter?: { [key: string]: any }): { [key: string]: any }[];
    findOne(filter?: { [key: string]: any }): { [key: string]: any } | null;
    insert(data: { [key: string]: any }): { [key: string]: any } | null;
    update(id: string, data: { [key: string]: any }): boolean;
    delete(id: string): boolean;
    count(filter?: { [key: string]: any }): number;
}

declare const database: {
    collection(name: string): Collection;
};

// Goals module
declare const goals: {
    increment(slug: string, value?: number): boolean;
};

// HTTP module
interface HTTPResponse {
    status: number;
    statusText: string;
    headers: { [key: string]: string };
    body: string;
}

interface HTTPOptions {
    headers?: { [key: string]: string };
    timeout?: number;
}

declare const http: {
    get(url: string, options?: HTTPOptions): HTTPResponse;
    post(url: string, body: any, options?: HTTPOptions): HTTPResponse;
    put(url: string, body: any, options?: HTTPOptions): HTTPResponse;
    delete(url: string, options?: HTTPOptions): HTTPResponse;
};

// Crypto module
declare const crypto: {
    md5(data: string): string;
    sha256(data: string): string;
    randomBytes(length: number): string;
};

// Encoding module
declare const encoding: {
    base64Encode(data: string): string;
    base64Decode(data: string): string;
    jsonParse(data: string): any;
    jsonStringify(data: any): string;
    urlEncode(data: string): string;
    urlDecode(data: string): string;
};

// Utils module
declare const utils: {
    sleep(ms: number): void;
    random(): number;
    randomInt(min: number, max: number): number;
    uuid(): string;
    slugify(text: string): string;
    truncate(text: string, length: number): string;
    capitalize(text: string): string;
    regexMatch(text: string, pattern: string): string[];
    regexReplace(text: string, pattern: string, replacement: string): string;
    formatDate(timestamp: number, format: string): string;
    parseDate(text: string, format: string): number;
    timestamp(): number;
};

// Delayed module
declare const delayed: {
    run(handler: () => void): void;
};
`
}
