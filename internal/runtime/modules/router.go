package modules

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/pkg/schema"
)

// CookieOptions defines options for setting cookies
type CookieOptions struct {
	MaxAge   int    `json:"maxAge"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"httpOnly"`
	SameSite string `json:"sameSite"` // "strict", "lax", "none"
}

// SetCookieData represents a cookie to be set
type SetCookieData struct {
	Name    string        `json:"name"`
	Value   string        `json:"value"`
	Options CookieOptions `json:"options"`
}

type RequestContext struct {
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Params    map[string]string `json:"params"`
	Query     map[string]string `json:"query"`
	Headers   map[string]string `json:"headers"`
	Body      interface{}       `json:"body"`
	IP        string            `json:"ip"`
	UserAgent string            `json:"userAgent"`
	Cookies   map[string]string `json:"cookies"`
}

// ResponseType indicates special response handling
type ResponseType string

const (
	ResponseTypeJSON     ResponseType = "json"
	ResponseTypeRedirect ResponseType = "redirect"
	ResponseTypeFile     ResponseType = "file"
	ResponseTypeRaw      ResponseType = "raw"
)

type ResponseData struct {
	Status      int               `json:"status"`
	Body        interface{}       `json:"body"`
	Headers     map[string]string `json:"headers"`
	Type        ResponseType      `json:"type"`
	RedirectURL string            `json:"redirectUrl,omitempty"`
	FilePath    string            `json:"filePath,omitempty"`
	SetCookies  []SetCookieData   `json:"setCookies,omitempty"`
	ContentType string            `json:"contentType,omitempty"`
}

// Middleware function type
type MiddlewareFunc func(ctx map[string]interface{}) (bool, error)

type routeHandler struct {
	pattern *regexp.Regexp
	params  []string
	handler goja.Callable
	vm      *goja.Runtime
}

type middlewareHandler struct {
	pathPrefix string         // empty for global middleware
	pattern    *regexp.Regexp // nil for global, compiled pattern for path-based
	handler    goja.Callable
	vm         *goja.Runtime
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Origins []string `json:"origins"`
	Methods []string `json:"methods"`
	Headers []string `json:"headers"`
}

type RouterModule struct {
	routes         map[string][]routeHandler
	middlewares    []middlewareHandler
	corsConfig     *CORSConfig
	mu             sync.RWMutex
	vm             *goja.Runtime
	hitCount       int64
	hitsByPath     map[string]int64
	hitsMu         sync.RWMutex
	handlerTimeout time.Duration // timeout for route handlers (0 = no timeout)
}

func NewRouterModule() *RouterModule {
	return &RouterModule{
		routes: map[string][]routeHandler{
			"GET":     {},
			"POST":    {},
			"PUT":     {},
			"DELETE":  {},
			"PATCH":   {},
			"HEAD":    {},
			"OPTIONS": {},
		},
		middlewares: []middlewareHandler{},
		hitsByPath:  make(map[string]int64),
	}
}

func (r *RouterModule) SetVM(vm *goja.Runtime) {
	r.vm = vm
}

// Name returns the module name for JavaScript
func (r *RouterModule) Name() string {
	return "$router"
}

// Register registers the module into the JavaScript VM
func (r *RouterModule) Register(vm interface{}) {
	v := vm.(*goja.Runtime)
	r.SetVM(v)
	v.Set(r.Name(), map[string]interface{}{
		"get":     r.Get,
		"post":    r.Post,
		"put":     r.Put,
		"delete":  r.Delete,
		"patch":   r.Patch,
		"head":    r.Head,
		"options": r.Options,
		"all":     r.All,
		"use":     r.Use,
		"group":   r.Group,
		"cors":    r.Cors,
	})
}

func (r *RouterModule) addRoute(method, path string, handler goja.Callable) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert path pattern to regex
	// e.g., /users/:id -> /users/([^/]+)
	params := []string{}
	pattern := path

	// Find all :param patterns
	paramRegex := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := paramRegex.FindAllStringSubmatch(path, -1)
	for _, match := range matches {
		params = append(params, match[1])
	}

	// Replace :param with capture group
	pattern = paramRegex.ReplaceAllString(pattern, `([^/]+)`)
	pattern = "^" + pattern + "$"

	re, err := regexp.Compile(pattern)
	if err != nil {
		return
	}

	r.routes[method] = append(r.routes[method], routeHandler{
		pattern: re,
		params:  params,
		handler: handler,
		vm:      r.vm,
	})
}

func (r *RouterModule) Get(path string, handler goja.Callable) {
	r.addRoute("GET", path, handler)
}

func (r *RouterModule) Post(path string, handler goja.Callable) {
	r.addRoute("POST", path, handler)
}

func (r *RouterModule) Put(path string, handler goja.Callable) {
	r.addRoute("PUT", path, handler)
}

func (r *RouterModule) Delete(path string, handler goja.Callable) {
	r.addRoute("DELETE", path, handler)
}

func (r *RouterModule) Patch(path string, handler goja.Callable) {
	r.addRoute("PATCH", path, handler)
}

func (r *RouterModule) Head(path string, handler goja.Callable) {
	r.addRoute("HEAD", path, handler)
}

func (r *RouterModule) Options(path string, handler goja.Callable) {
	r.addRoute("OPTIONS", path, handler)
}

// All registers a handler for all HTTP methods
func (r *RouterModule) All(path string, handler goja.Callable) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		r.addRoute(method, path, handler)
	}
}

// Use adds middleware. Can be called with just handler (global) or with path prefix and handler
func (r *RouterModule) Use(call goja.FunctionCall) goja.Value {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(call.Arguments) == 0 {
		return goja.Undefined()
	}

	var pathPrefix string
	var handler goja.Callable

	if len(call.Arguments) == 1 {
		// Global middleware: use(handler)
		h, ok := goja.AssertFunction(call.Arguments[0])
		if !ok {
			return goja.Undefined()
		}
		handler = h
		pathPrefix = ""
	} else {
		// Path-based middleware: use('/api', handler)
		pathPrefix = call.Arguments[0].String()
		h, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			return goja.Undefined()
		}
		handler = h
	}

	// Compile pattern for path prefix
	var pattern *regexp.Regexp
	if pathPrefix != "" {
		// Match paths that start with this prefix
		pattern = regexp.MustCompile("^" + regexp.QuoteMeta(pathPrefix))
	}

	r.middlewares = append(r.middlewares, middlewareHandler{
		pathPrefix: pathPrefix,
		pattern:    pattern,
		handler:    handler,
		vm:         r.vm,
	})

	return goja.Undefined()
}

// Group creates a route group with a common prefix
func (r *RouterModule) Group(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		return goja.Undefined()
	}

	prefix := call.Arguments[0].String()
	callback, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		return goja.Undefined()
	}

	// Create a group router object with methods that prepend the prefix
	groupRouter := r.vm.NewObject()

	groupRouter.Set("get", func(path string, handler goja.Callable) {
		r.addRoute("GET", prefix+path, handler)
	})
	groupRouter.Set("post", func(path string, handler goja.Callable) {
		r.addRoute("POST", prefix+path, handler)
	})
	groupRouter.Set("put", func(path string, handler goja.Callable) {
		r.addRoute("PUT", prefix+path, handler)
	})
	groupRouter.Set("delete", func(path string, handler goja.Callable) {
		r.addRoute("DELETE", prefix+path, handler)
	})
	groupRouter.Set("patch", func(path string, handler goja.Callable) {
		r.addRoute("PATCH", prefix+path, handler)
	})
	groupRouter.Set("head", func(path string, handler goja.Callable) {
		r.addRoute("HEAD", prefix+path, handler)
	})
	groupRouter.Set("options", func(path string, handler goja.Callable) {
		r.addRoute("OPTIONS", prefix+path, handler)
	})
	groupRouter.Set("all", func(path string, handler goja.Callable) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
		for _, method := range methods {
			r.addRoute(method, prefix+path, handler)
		}
	})

	// Call the callback with the group router
	callback(goja.Undefined(), groupRouter)

	return goja.Undefined()
}

// Cors configures CORS settings
func (r *RouterModule) Cors(call goja.FunctionCall) goja.Value {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(call.Arguments) == 0 {
		return goja.Undefined()
	}

	config := call.Arguments[0].Export()
	if configMap, ok := config.(map[string]interface{}); ok {
		cors := &CORSConfig{}

		if origins, ok := configMap["origins"]; ok {
			if originsSlice, ok := origins.([]interface{}); ok {
				for _, o := range originsSlice {
					cors.Origins = append(cors.Origins, fmt.Sprintf("%v", o))
				}
			}
		}

		if methods, ok := configMap["methods"]; ok {
			if methodsSlice, ok := methods.([]interface{}); ok {
				for _, m := range methodsSlice {
					cors.Methods = append(cors.Methods, fmt.Sprintf("%v", m))
				}
			}
		}

		if headers, ok := configMap["headers"]; ok {
			if headersSlice, ok := headers.([]interface{}); ok {
				for _, h := range headersSlice {
					cors.Headers = append(cors.Headers, fmt.Sprintf("%v", h))
				}
			}
		}

		r.corsConfig = cors
	}

	return goja.Undefined()
}

// GetCORSConfig returns the current CORS configuration
func (r *RouterModule) GetCORSConfig() *CORSConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.corsConfig
}

func (r *RouterModule) Handle(method, path string, ctx *RequestContext) (*ResponseData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	method = strings.ToUpper(method)
	handlers, ok := r.routes[method]
	if !ok {
		return nil, fmt.Errorf("method not allowed")
	}

	for _, h := range handlers {
		matches := h.pattern.FindStringSubmatch(path)
		if matches == nil {
			continue
		}

		// Track hit
		r.hitsMu.Lock()
		r.hitCount++
		routeKey := method + " " + path
		r.hitsByPath[routeKey]++
		r.hitsMu.Unlock()

		// Extract params
		if ctx.Params == nil {
			ctx.Params = make(map[string]string)
		}
		for i, param := range h.params {
			if i+1 < len(matches) {
				ctx.Params[param] = matches[i+1]
			}
		}

		// Create response accumulator for ctx methods
		respAccum := &ResponseData{
			Status:     200,
			Headers:    make(map[string]string),
			Type:       ResponseTypeJSON,
			SetCookies: []SetCookieData{},
		}

		// Build context map with extended properties and methods
		ctxMap := r.buildContextMap(ctx, respAccum, h.vm)

		// Run middleware chain
		if !r.runMiddleware(path, ctxMap, h.vm) {
			// Middleware returned false - abort
			return respAccum, nil
		}

		// Call handler with context as argument
		var result goja.Value
		var err error

		if h.vm != nil {
			ctxValue := h.vm.ToValue(ctxMap)
			result, err = h.handler(goja.Undefined(), ctxValue)
		} else {
			result, err = h.handler(goja.Undefined())
		}

		if err != nil {
			return nil, err
		}

		// If respAccum was modified by ctx methods (redirect, file, etc.), use it
		if respAccum.Type != ResponseTypeJSON || respAccum.RedirectURL != "" || respAccum.FilePath != "" {
			return respAccum, nil
		}

		// Parse result from handler return value
		return r.parseHandlerResult(result, respAccum)
	}

	return nil, fmt.Errorf("route not found")
}

// buildContextMap creates the JS context object with extended properties and methods
func (r *RouterModule) buildContextMap(ctx *RequestContext, respAccum *ResponseData, vm *goja.Runtime) map[string]interface{} {
	ctxMap := map[string]interface{}{
		"method":    ctx.Method,
		"path":      ctx.Path,
		"params":    ctx.Params,
		"query":     ctx.Query,
		"headers":   ctx.Headers,
		"body":      ctx.Body,
		"ip":        ctx.IP,
		"userAgent": ctx.UserAgent,
		"cookies":   ctx.Cookies,
	}

	// Add setCookie method
	ctxMap["setCookie"] = func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		name := call.Arguments[0].String()
		value := call.Arguments[1].String()

		cookie := SetCookieData{
			Name:  name,
			Value: value,
			Options: CookieOptions{
				Path: "/",
			},
		}

		if len(call.Arguments) > 2 {
			if opts := call.Arguments[2].Export(); opts != nil {
				if optsMap, ok := opts.(map[string]interface{}); ok {
					if maxAge, ok := optsMap["maxAge"]; ok {
						cookie.Options.MaxAge = toInt(maxAge)
					}
					if path, ok := optsMap["path"]; ok {
						cookie.Options.Path = fmt.Sprintf("%v", path)
					}
					if domain, ok := optsMap["domain"]; ok {
						cookie.Options.Domain = fmt.Sprintf("%v", domain)
					}
					if secure, ok := optsMap["secure"]; ok {
						cookie.Options.Secure = toBool(secure)
					}
					if httpOnly, ok := optsMap["httpOnly"]; ok {
						cookie.Options.HttpOnly = toBool(httpOnly)
					}
					if sameSite, ok := optsMap["sameSite"]; ok {
						cookie.Options.SameSite = fmt.Sprintf("%v", sameSite)
					}
				}
			}
		}

		respAccum.SetCookies = append(respAccum.SetCookies, cookie)
		return goja.Undefined()
	}

	// Add redirect method
	ctxMap["redirect"] = func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		url := call.Arguments[0].String()
		code := 302 // default redirect code
		if len(call.Arguments) > 1 {
			code = toInt(call.Arguments[1].Export())
		}

		respAccum.Type = ResponseTypeRedirect
		respAccum.RedirectURL = url
		respAccum.Status = code
		return goja.Undefined()
	}

	// Add response method - ctx.response(status, body, headers?)
	ctxMap["response"] = func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}

		status := toInt(call.Arguments[0].Export())
		body := call.Arguments[1].Export()

		respAccum.Status = status
		respAccum.Body = body

		// Check if body is string for raw response
		if _, isString := body.(string); isString {
			respAccum.Type = ResponseTypeRaw
		}

		if len(call.Arguments) > 2 {
			if headers := call.Arguments[2].Export(); headers != nil {
				if headersMap, ok := headers.(map[string]interface{}); ok {
					for k, v := range headersMap {
						respAccum.Headers[k] = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		return vm.ToValue(map[string]interface{}{
			"status":  status,
			"body":    body,
			"headers": respAccum.Headers,
		})
	}

	// Add file method
	ctxMap["file"] = func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		filePath := call.Arguments[0].String()
		respAccum.Type = ResponseTypeFile
		respAccum.FilePath = filePath
		respAccum.Status = 200
		return goja.Undefined()
	}

	return ctxMap
}

// runMiddleware executes the middleware chain
func (r *RouterModule) runMiddleware(path string, ctxMap map[string]interface{}, vm *goja.Runtime) bool {
	for _, mw := range r.middlewares {
		// Check if middleware applies to this path
		if mw.pattern != nil && !mw.pattern.MatchString(path) {
			continue
		}

		// Call middleware
		if mw.vm != nil {
			ctxValue := mw.vm.ToValue(ctxMap)
			result, err := mw.handler(goja.Undefined(), ctxValue)
			if err != nil {
				return false
			}

			// If middleware returns false, abort the chain
			if exported := result.Export(); exported != nil {
				if b, ok := exported.(bool); ok && !b {
					return false
				}
			}
		}
	}
	return true
}

// parseHandlerResult converts handler return value to ResponseData
func (r *RouterModule) parseHandlerResult(result goja.Value, respAccum *ResponseData) (*ResponseData, error) {
	exported := result.Export()

	// Try to convert map to ResponseData
	if m, ok := exported.(map[string]interface{}); ok {
		if status, ok := m["status"]; ok {
			respAccum.Status = toInt(status)
		}
		if body, ok := m["body"]; ok {
			respAccum.Body = body
		}
		if headers, ok := m["headers"].(map[string]interface{}); ok {
			for k, v := range headers {
				respAccum.Headers[k] = fmt.Sprintf("%v", v)
			}
		}
		// Check for special response types
		if respType, ok := m["type"]; ok {
			respAccum.Type = ResponseType(fmt.Sprintf("%v", respType))
		}
		if redirectURL, ok := m["redirectUrl"]; ok {
			respAccum.RedirectURL = fmt.Sprintf("%v", redirectURL)
		}
		if filePath, ok := m["filePath"]; ok {
			respAccum.FilePath = fmt.Sprintf("%v", filePath)
		}
		return respAccum, nil
	}

	respAccum.Body = exported
	return respAccum, nil
}

// Helper functions for type conversion
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		return 0
	default:
		return 0
	}
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val == "true" || val == "1"
	default:
		return false
	}
}

func (r *RouterModule) HasRoutes() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, handlers := range r.routes {
		if len(handlers) > 0 {
			return true
		}
	}
	return false
}

// RoutesCount returns the total number of registered routes
func (r *RouterModule) RoutesCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, handlers := range r.routes {
		count += len(handlers)
	}
	return count
}

// RoutesByMethod returns the count of routes per HTTP method
func (r *RouterModule) RoutesByMethod() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]int)
	for method, handlers := range r.routes {
		if len(handlers) > 0 {
			result[method] = len(handlers)
		}
	}
	return result
}

// HitCount returns total request count
func (r *RouterModule) HitCount() int64 {
	r.hitsMu.RLock()
	defer r.hitsMu.RUnlock()
	return r.hitCount
}

// HitsByPath returns hits per route
func (r *RouterModule) HitsByPath() map[string]int64 {
	r.hitsMu.RLock()
	defer r.hitsMu.RUnlock()
	result := make(map[string]int64)
	for k, v := range r.hitsByPath {
		result[k] = v
	}
	return result
}

// ResetHits resets hit counters and returns the last values
func (r *RouterModule) ResetHits() int64 {
	r.hitsMu.Lock()
	defer r.hitsMu.Unlock()
	count := r.hitCount
	r.hitCount = 0
	r.hitsByPath = make(map[string]int64)
	return count
}

// GetSchema implements JSSchemaProvider
func (r *RouterModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$router",
		Description: "HTTP routing for creating API endpoints with middleware, grouping, and CORS support",
		Types: []schema.TypeSchema{
			{
				Name:        "CookieOptions",
				Description: "Options for setting cookies",
				Fields: []schema.ParamSchema{
					{Name: "maxAge", Type: "number", Description: "Cookie max age in seconds", Optional: true},
					{Name: "path", Type: "string", Description: "Cookie path (default: /)", Optional: true},
					{Name: "domain", Type: "string", Description: "Cookie domain", Optional: true},
					{Name: "secure", Type: "boolean", Description: "HTTPS only", Optional: true},
					{Name: "httpOnly", Type: "boolean", Description: "HTTP only (not accessible via JS)", Optional: true},
					{Name: "sameSite", Type: "'strict' | 'lax' | 'none'", Description: "SameSite attribute", Optional: true},
				},
			},
			{
				Name:        "CORSConfig",
				Description: "CORS configuration",
				Fields: []schema.ParamSchema{
					{Name: "origins", Type: "string[]", Description: "Allowed origins (use ['*'] for all)"},
					{Name: "methods", Type: "string[]", Description: "Allowed HTTP methods", Optional: true},
					{Name: "headers", Type: "string[]", Description: "Allowed headers", Optional: true},
				},
			},
			{
				Name:        "RequestContext",
				Description: "HTTP request context passed to route handlers",
				Fields: []schema.ParamSchema{
					{Name: "method", Type: "string", Description: "HTTP method"},
					{Name: "path", Type: "string", Description: "Request path"},
					{Name: "params", Type: "{ [key: string]: string }", Description: "URL path parameters"},
					{Name: "query", Type: "{ [key: string]: string }", Description: "Query string parameters"},
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Request headers"},
					{Name: "body", Type: "any", Description: "Request body"},
					{Name: "ip", Type: "string", Description: "Client IP address"},
					{Name: "userAgent", Type: "string", Description: "Client User-Agent"},
					{Name: "cookies", Type: "{ [key: string]: string }", Description: "Request cookies"},
					{Name: "setCookie", Type: "(name: string, value: string, options?: CookieOptions) => void", Description: "Set a response cookie"},
					{Name: "redirect", Type: "(url: string, code?: number) => void", Description: "Redirect to URL (default code: 302)"},
					{Name: "response", Type: "(status: number, body: any, headers?: { [key: string]: string }) => ResponseData", Description: "Create and send response"},
					{Name: "file", Type: "(path: string) => void", Description: "Serve a file from storage"},
				},
			},
			{
				Name:        "ResponseData",
				Description: "HTTP response data",
				Fields: []schema.ParamSchema{
					{Name: "status", Type: "number", Description: "HTTP status code"},
					{Name: "body", Type: "any", Description: "Response body"},
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Response headers", Optional: true},
				},
			},
			{
				Name:        "GroupRouter",
				Description: "Router for grouped routes",
				Fields: []schema.ParamSchema{
					{Name: "get", Type: "(path: string, handler: RouteHandler) => void", Description: "Register GET route"},
					{Name: "post", Type: "(path: string, handler: RouteHandler) => void", Description: "Register POST route"},
					{Name: "put", Type: "(path: string, handler: RouteHandler) => void", Description: "Register PUT route"},
					{Name: "delete", Type: "(path: string, handler: RouteHandler) => void", Description: "Register DELETE route"},
					{Name: "patch", Type: "(path: string, handler: RouteHandler) => void", Description: "Register PATCH route"},
					{Name: "head", Type: "(path: string, handler: RouteHandler) => void", Description: "Register HEAD route"},
					{Name: "options", Type: "(path: string, handler: RouteHandler) => void", Description: "Register OPTIONS route"},
					{Name: "all", Type: "(path: string, handler: RouteHandler) => void", Description: "Register handler for all methods"},
				},
			},
		},
		Methods: []schema.MethodSchema{
			{
				Name:        "get",
				Description: "Register a GET route handler",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "post",
				Description: "Register a POST route handler",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "put",
				Description: "Register a PUT route handler",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "delete",
				Description: "Register a DELETE route handler",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "patch",
				Description: "Register a PATCH route handler",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "head",
				Description: "Register a HEAD route handler",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "options",
				Description: "Register an OPTIONS route handler",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "all",
				Description: "Register handler for all HTTP methods",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "use",
				Description: "Add middleware (global or path-based)",
				Params: []schema.ParamSchema{
					{Name: "pathOrHandler", Type: "string | ((ctx: RequestContext) => boolean | void)", Description: "Path prefix or middleware function"},
					{Name: "handler", Type: "((ctx: RequestContext) => boolean | void)", Description: "Middleware function (if path provided)", Optional: true},
				},
			},
			{
				Name:        "group",
				Description: "Create route group with common prefix",
				Params: []schema.ParamSchema{
					{Name: "prefix", Type: "string", Description: "URL path prefix for all routes in group"},
					{Name: "callback", Type: "(router: GroupRouter) => void", Description: "Function to define routes"},
				},
			},
			{
				Name:        "cors",
				Description: "Configure CORS settings",
				Params: []schema.ParamSchema{
					{Name: "config", Type: "CORSConfig", Description: "CORS configuration"},
				},
			},
		},
	}
}

// GetRouterSchema returns the router schema (static version)
func GetRouterSchema() schema.ModuleSchema {
	return (&RouterModule{}).GetSchema()
}
