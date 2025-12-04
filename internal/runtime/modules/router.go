package modules

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/dop251/goja"
)

type RequestContext struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Params  map[string]string `json:"params"`
	Query   map[string]string `json:"query"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

type ResponseData struct {
	Status  int               `json:"status"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
}

type routeHandler struct {
	pattern *regexp.Regexp
	params  []string
	handler goja.Callable
	vm      *goja.Runtime
}

type RouterModule struct {
	routes map[string][]routeHandler
	mu     sync.RWMutex
	vm     *goja.Runtime
}

func NewRouterModule() *RouterModule {
	return &RouterModule{
		routes: map[string][]routeHandler{
			"GET":    {},
			"POST":   {},
			"PUT":    {},
			"DELETE": {},
		},
	}
}

func (r *RouterModule) SetVM(vm *goja.Runtime) {
	r.vm = vm
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

func (r *RouterModule) Response(status int, body interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status":  status,
		"body":    body,
		"headers": make(map[string]string),
	}
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

		// Extract params
		if ctx.Params == nil {
			ctx.Params = make(map[string]string)
		}
		for i, param := range h.params {
			if i+1 < len(matches) {
				ctx.Params[param] = matches[i+1]
			}
		}

		// Convert context to map for JavaScript
		ctxMap := map[string]interface{}{
			"method":  ctx.Method,
			"path":    ctx.Path,
			"params":  ctx.Params,
			"query":   ctx.Query,
			"headers": ctx.Headers,
			"body":    ctx.Body,
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

		// Parse result
		exported := result.Export()

		// Try to convert map to ResponseData
		if m, ok := exported.(map[string]interface{}); ok {
			resp := &ResponseData{
				Status:  200,
				Headers: make(map[string]string),
			}
			if status, ok := m["status"]; ok {
				switch v := status.(type) {
				case int64:
					resp.Status = int(v)
				case float64:
					resp.Status = int(v)
				case int:
					resp.Status = v
				}
			}
			if body, ok := m["body"]; ok {
				resp.Body = body
			}
			if headers, ok := m["headers"].(map[string]interface{}); ok {
				for k, v := range headers {
					if str, ok := v.(string); ok {
						resp.Headers[k] = str
					}
				}
			}
			return resp, nil
		}

		return &ResponseData{
			Status: 200,
			Body:   exported,
		}, nil
	}

	return nil, fmt.Errorf("route not found")
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

// GetSchema implements JSSchemaProvider
func (r *RouterModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "router",
		Description: "HTTP routing for creating API endpoints",
		Types: []JSTypeSchema{
			{
				Name:        "RequestContext",
				Description: "HTTP request context passed to route handlers",
				Fields: []JSParamSchema{
					{Name: "method", Type: "string", Description: "HTTP method"},
					{Name: "path", Type: "string", Description: "Request path"},
					{Name: "params", Type: "{ [key: string]: string }", Description: "URL path parameters"},
					{Name: "query", Type: "{ [key: string]: string }", Description: "Query string parameters"},
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Request headers"},
					{Name: "body", Type: "any", Description: "Request body"},
				},
			},
			{
				Name:        "ResponseData",
				Description: "HTTP response data",
				Fields: []JSParamSchema{
					{Name: "status", Type: "number", Description: "HTTP status code"},
					{Name: "body", Type: "any", Description: "Response body"},
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Response headers", Optional: true},
				},
			},
		},
		Methods: []JSMethodSchema{
			{
				Name:        "get",
				Description: "Register a GET route handler",
				Params: []JSParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "post",
				Description: "Register a POST route handler",
				Params: []JSParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "put",
				Description: "Register a PUT route handler",
				Params: []JSParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "delete",
				Description: "Register a DELETE route handler",
				Params: []JSParamSchema{
					{Name: "path", Type: "string", Description: "URL path pattern (supports :param)"},
					{Name: "handler", Type: "(ctx: RequestContext) => ResponseData | any", Description: "Route handler function"},
				},
			},
			{
				Name:        "response",
				Description: "Create a response object",
				Params: []JSParamSchema{
					{Name: "status", Type: "number", Description: "HTTP status code"},
					{Name: "body", Type: "any", Description: "Response body"},
				},
				Returns: &JSParamSchema{Type: "ResponseData"},
			},
		},
	}
}

// GetRouterSchema returns the router schema (static version)
func GetRouterSchema() JSModuleSchema {
	return (&RouterModule{}).GetSchema()
}
