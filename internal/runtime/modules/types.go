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
    /** Get a value by key, returns undefined if not found */
    get(key: string): any;

    /** Check if a key exists in the environment */
    has(key: string): boolean;

    /** Get all environment variable keys */
    keys(): string[];

    /** Get a string value with default fallback */
    getString(key: string, defaultValue: string): string;

    /** Get an integer value with default fallback */
    getInt(key: string, defaultValue: number): number;

    /** Get a float value with default fallback */
    getFloat(key: string, defaultValue: number): number;

    /** Get a boolean value with default fallback */
    getBool(key: string, defaultValue: boolean): boolean;

    /** Get all environment variables as a map */
    getAll(): { [key: string]: any };
};

// SMTP module
interface EmailOptions {
    from?: string;
    reply_to?: string;
    cc?: string[];
    bcc?: string[];
    headers?: { [key: string]: string };
    content_type?: 'text/plain' | 'text/html';
}

interface SMTPResult {
    success: boolean;
    error?: string;
}

declare const smtp: {
    /**
     * Send an email.
     * Requires env vars: SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM
     */
    send(to: string, subject: string, body: string, options?: EmailOptions): SMTPResult;

    /** Send an HTML email (convenience method) */
    sendHTML(to: string, subject: string, body: string): SMTPResult;
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

// Image module
interface ImageInfo {
    width: number;
    height: number;
    format: string;
}

declare const image: {
    /** Get image dimensions and format */
    info(path: string): ImageInfo | null;

    /** Resize image to exact dimensions */
    resize(src: string, dst: string, width: number, height: number): boolean;

    /** Resize image keeping aspect ratio (fit within bounds) */
    resizeKeepRatio(src: string, dst: string, maxWidth: number, maxHeight: number): boolean;

    /** Crop image to specified region */
    crop(src: string, dst: string, x: number, y: number, width: number, height: number): boolean;

    /** Create square thumbnail centered on image */
    thumbnail(src: string, dst: string, size: number): boolean;

    /** Read image as base64 data URI */
    readAsBase64(path: string): string;
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
interface GoalInfo {
    id: string;
    name: string;
    slug: string;
    type: 'counter' | 'daily_counter';
    description: string;
    color: string;
}

interface GoalStat {
    date: string;
    value: number;
}

declare const goals: {
    /** Increment a goal counter by the specified value (default 1) */
    increment(slug: string, value?: number): boolean;

    /** Get the current value of a goal (total for counter, today's value for daily_counter) */
    getValue(slug: string): number;

    /** Get statistics for a goal over a period of days (default 7) */
    getStats(slug: string, days?: number): GoalStat[];

    /** List all goals accessible by this project */
    list(): GoalInfo[];

    /** Get information about a specific goal by slug */
    get(slug: string): GoalInfo | null;
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

// Service lifecycle module
declare const service: {
    /**
     * Register a callback to be called during service initialization (boot phase).
     * Use this for setting up initial state, loading configuration, etc.
     */
    boot(callback: () => void): void;

    /**
     * Register a callback to be called when service is ready (start phase).
     * Use this for starting listeners, scheduling tasks, etc.
     */
    start(callback: () => void): void;

    /**
     * Register a callback to be called when service is stopping (shutdown phase).
     * Use this for cleanup, saving state, closing connections, etc.
     */
    shutdown(callback: () => void): void;
};
`
}
