package modules

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dop251/goja"
	"m3m/pkg/schema"
)

type HTTPModule struct {
	client *http.Client
}

// Name returns the module name for JavaScript
func (h *HTTPModule) Name() string {
	return "$http"
}

// Register registers the module into the JavaScript VM
func (h *HTTPModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(h.Name(), map[string]interface{}{
		"get":    h.Get,
		"post":   h.Post,
		"put":    h.Put,
		"delete": h.Delete,
	})
}

type HTTPResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type HTTPOptions struct {
	Headers map[string]string `json:"headers"`
	Timeout int               `json:"timeout"` // in milliseconds
}

func NewHTTPModule(timeout time.Duration) *HTTPModule {
	return &HTTPModule{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (h *HTTPModule) doRequest(method, url string, body interface{}, options *HTTPOptions) *HTTPResponse {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return &HTTPResponse{Status: 0, StatusText: err.Error()}
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}

	// Set default content type for POST/PUT
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	if options != nil && options.Headers != nil {
		for k, v := range options.Headers {
			req.Header.Set(k, v)
		}
	}

	// Create client with custom timeout if specified
	client := h.client
	if options != nil && options.Timeout > 0 {
		client = &http.Client{
			Timeout: time.Duration(options.Timeout) * time.Millisecond,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{Status: resp.StatusCode, StatusText: err.Error()}
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &HTTPResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		Body:       string(respBody),
	}
}

func (h *HTTPModule) Get(url string, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return h.doRequest("GET", url, nil, opts)
}

func (h *HTTPModule) Post(url string, body interface{}, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return h.doRequest("POST", url, body, opts)
}

func (h *HTTPModule) Put(url string, body interface{}, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return h.doRequest("PUT", url, body, opts)
}

func (h *HTTPModule) Delete(url string, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return h.doRequest("DELETE", url, nil, opts)
}

// GetSchema implements JSSchemaProvider
func (h *HTTPModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$http",
		Description: "HTTP client for making external API requests",
		Types: []schema.TypeSchema{
			{
				Name:        "HTTPResponse",
				Description: "Response from an HTTP request",
				Fields: []schema.ParamSchema{
					{Name: "status", Type: "number", Description: "HTTP status code"},
					{Name: "statusText", Type: "string", Description: "HTTP status text"},
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Response headers"},
					{Name: "body", Type: "string", Description: "Response body as string"},
				},
			},
			{
				Name:        "HTTPOptions",
				Description: "Options for HTTP requests",
				Fields: []schema.ParamSchema{
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Request headers", Optional: true},
					{Name: "timeout", Type: "number", Description: "Request timeout in milliseconds", Optional: true},
				},
			},
		},
		Methods: []schema.MethodSchema{
			{
				Name:        "get",
				Description: "Make a GET request",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "post",
				Description: "Make a POST request",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "body", Type: "any", Description: "Request body (will be JSON encoded)"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "put",
				Description: "Make a PUT request",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "body", Type: "any", Description: "Request body (will be JSON encoded)"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "delete",
				Description: "Make a DELETE request",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
		},
	}
}

// GetHTTPSchema returns the http schema (static version)
func GetHTTPSchema() schema.ModuleSchema {
	return (&HTTPModule{}).GetSchema()
}
