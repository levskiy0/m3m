package modules

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"m3m/internal/service"
	"m3m/pkg/schema"
)

type HTTPModule struct {
	client    *http.Client
	storage   *service.StorageService
	projectID string
}

// Name returns the module name for JavaScript
func (h *HTTPModule) Name() string {
	return "$http"
}

// Register registers the module into the JavaScript VM
func (h *HTTPModule) Register(vm interface{}) {
	runtime := vm.(*goja.Runtime)

	// Create response object with methods
	createResponse := func(resp *HTTPResponse) goja.Value {
		obj := runtime.NewObject()
		_ = obj.Set("status", resp.Status)
		_ = obj.Set("statusText", resp.StatusText)
		_ = obj.Set("headers", resp.Headers)
		_ = obj.Set("body", resp.Body)
		_ = obj.Set("json", func() interface{} { return resp.JSON() })
		_ = obj.Set("text", func() string { return resp.Text() })
		_ = obj.Set("buffer", func() string { return resp.Buffer() })
		return obj
	}

	// Wrap methods to return enhanced response objects
	runtime.Set(h.Name(), map[string]interface{}{
		"get": func(url string, options ...*HTTPOptions) goja.Value {
			return createResponse(h.Get(url, options...))
		},
		"post": func(url string, body interface{}, options ...*HTTPOptions) goja.Value {
			return createResponse(h.Post(url, body, options...))
		},
		"put": func(url string, body interface{}, options ...*HTTPOptions) goja.Value {
			return createResponse(h.Put(url, body, options...))
		},
		"delete": func(url string, options ...*HTTPOptions) goja.Value {
			return createResponse(h.Delete(url, options...))
		},
		"patch": func(url string, body interface{}, options ...*HTTPOptions) goja.Value {
			return createResponse(h.Patch(url, body, options...))
		},
		"head": func(url string, options ...*HTTPOptions) goja.Value {
			return createResponse(h.Head(url, options...))
		},
		"options": func(url string, options ...*HTTPOptions) goja.Value {
			return createResponse(h.Options(url, options...))
		},
		"postForm": func(url string, fields map[string]interface{}, options ...*HTTPOptions) goja.Value {
			return createResponse(h.PostForm(url, fields, options...))
		},
		"postFormURLEncoded": func(url string, fields map[string]string, options ...*HTTPOptions) goja.Value {
			return createResponse(h.PostFormURLEncoded(url, fields, options...))
		},
		"download": func(url string, storagePath string, options ...*HTTPOptions) goja.Value {
			return createResponse(h.Download(url, storagePath, options...))
		},
	})
}

type HTTPResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	bodyBytes  []byte            // internal: raw body bytes for buffer()
}

// JSON parses the response body as JSON
func (r *HTTPResponse) JSON() interface{} {
	var result interface{}
	if err := json.Unmarshal(r.bodyBytes, &result); err != nil {
		return nil
	}
	return result
}

// Text returns the response body as string
func (r *HTTPResponse) Text() string {
	return r.Body
}

// Buffer returns the response body as base64-encoded string
func (r *HTTPResponse) Buffer() string {
	return base64.StdEncoding.EncodeToString(r.bodyBytes)
}

type BasicAuth struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

type HTTPOptions struct {
	Headers         map[string]string `json:"headers"`
	Timeout         int               `json:"timeout"`         // in milliseconds
	FollowRedirects *bool             `json:"followRedirects"` // default: true
	MaxRedirects    int               `json:"maxRedirects"`    // default: 10
	Proxy           string            `json:"proxy"`           // proxy URL
	BasicAuth       *BasicAuth        `json:"basicAuth"`       // basic auth credentials
	BearerToken     string            `json:"bearerToken"`     // bearer token
	SkipTLSVerify   bool              `json:"skipTLSVerify"`   // skip TLS verification
}

type FormField struct {
	Value    string `json:"value"`    // string value
	Filename string `json:"filename"` // filename for file uploads
	Content  string `json:"content"`  // base64 encoded content for files
}

func NewHTTPModule(timeout time.Duration, storage *service.StorageService, projectID string) *HTTPModule {
	return &HTTPModule{
		client: &http.Client{
			Timeout: timeout,
		},
		storage:   storage,
		projectID: projectID,
	}
}

func (h *HTTPModule) buildClient(options *HTTPOptions) *http.Client {
	// Start with default timeout from base client
	timeout := h.client.Timeout
	if options != nil && options.Timeout > 0 {
		timeout = time.Duration(options.Timeout) * time.Millisecond
	}

	// Build transport
	transport := &http.Transport{}

	// Proxy support
	if options != nil && options.Proxy != "" {
		proxyURL, err := url.Parse(options.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	// Skip TLS verification (use with caution)
	if options != nil && options.SkipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Build redirect policy
	var checkRedirect func(*http.Request, []*http.Request) error
	if options != nil {
		followRedirects := true
		if options.FollowRedirects != nil {
			followRedirects = *options.FollowRedirects
		}

		if !followRedirects {
			checkRedirect = func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			}
		} else if options.MaxRedirects > 0 {
			maxRedirects := options.MaxRedirects
			checkRedirect = func(_ *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				return nil
			}
		}
	}

	return &http.Client{
		Timeout:       timeout,
		Transport:     transport,
		CheckRedirect: checkRedirect,
	}
}

func (h *HTTPModule) doRequest(method, urlStr string, body interface{}, options *HTTPOptions) *HTTPResponse {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return &HTTPResponse{Status: 0, StatusText: err.Error()}
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}

	// Set default content type for POST/PUT/PATCH
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Apply options
	h.applyOptions(req, options)

	client := h.buildClient(options)
	resp, err := client.Do(req)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}
	defer resp.Body.Close()

	return h.parseResponse(resp)
}

func (h *HTTPModule) applyOptions(req *http.Request, options *HTTPOptions) {
	if options == nil {
		return
	}

	// Set custom headers
	for k, v := range options.Headers {
		req.Header.Set(k, v)
	}

	// Basic auth
	if options.BasicAuth != nil {
		req.SetBasicAuth(options.BasicAuth.User, options.BasicAuth.Pass)
	}

	// Bearer token
	if options.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+options.BearerToken)
	}
}

func (h *HTTPModule) parseResponse(resp *http.Response) *HTTPResponse {
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
		bodyBytes:  respBody,
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

func (h *HTTPModule) Patch(url string, body interface{}, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return h.doRequest("PATCH", url, body, opts)
}

func (h *HTTPModule) Head(url string, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return h.doRequest("HEAD", url, nil, opts)
}

func (h *HTTPModule) Options(url string, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return h.doRequest("OPTIONS", url, nil, opts)
}

// PostForm sends a multipart form data request
// fields can be map[string]string for simple values or map[string]FormField for files
func (h *HTTPModule) PostForm(urlStr string, fields map[string]interface{}, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		switch v := value.(type) {
		case string:
			_ = writer.WriteField(key, v)
		case map[string]interface{}:
			// Handle file upload
			filename, _ := v["filename"].(string)
			content, _ := v["content"].(string)
			valueStr, _ := v["value"].(string)

			if filename != "" && content != "" {
				// File upload with base64 content
				part, err := writer.CreateFormFile(key, filename)
				if err != nil {
					return &HTTPResponse{Status: 0, StatusText: err.Error()}
				}
				decoded, err := base64.StdEncoding.DecodeString(content)
				if err != nil {
					return &HTTPResponse{Status: 0, StatusText: "failed to decode base64 content: " + err.Error()}
				}
				_, _ = part.Write(decoded)
			} else if valueStr != "" {
				_ = writer.WriteField(key, valueStr)
			}
		}
	}

	err := writer.Close()
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}

	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	h.applyOptions(req, opts)

	client := h.buildClient(opts)
	resp, err := client.Do(req)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}
	defer resp.Body.Close()

	return h.parseResponse(resp)
}

// Download downloads a file from URL and saves it to storage
func (h *HTTPModule) Download(urlStr string, storagePath string, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}

	h.applyOptions(req, opts)

	client := h.buildClient(opts)
	resp, err := client.Do(req)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{Status: resp.StatusCode, StatusText: err.Error()}
	}

	// Save to storage if storage service is available
	if h.storage != nil && h.projectID != "" {
		// Ensure directory exists
		dir := filepath.Dir(storagePath)
		if dir != "" && dir != "." {
			_ = h.storage.MkDir(h.projectID, dir)
		}

		err = h.storage.Write(h.projectID, storagePath, respBody)
		if err != nil {
			return &HTTPResponse{Status: resp.StatusCode, StatusText: "download succeeded but failed to save: " + err.Error()}
		}
	} else {
		return &HTTPResponse{Status: 0, StatusText: "storage service not available"}
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// Return response with saved file path info
	return &HTTPResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		Body:       storagePath, // Return the path where file was saved
		bodyBytes:  respBody,
	}
}

// PostFormURLEncoded sends an application/x-www-form-urlencoded request
func (h *HTTPModule) PostFormURLEncoded(urlStr string, fields map[string]string, options ...*HTTPOptions) *HTTPResponse {
	var opts *HTTPOptions
	if len(options) > 0 {
		opts = options[0]
	}

	data := url.Values{}
	for key, value := range fields {
		data.Set(key, value)
	}

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.applyOptions(req, opts)

	client := h.buildClient(opts)
	resp, err := client.Do(req)
	if err != nil {
		return &HTTPResponse{Status: 0, StatusText: err.Error()}
	}
	defer resp.Body.Close()

	return h.parseResponse(resp)
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
					{Name: "json", Type: "() => any", Description: "Parse body as JSON and return object"},
					{Name: "text", Type: "() => string", Description: "Return body as string"},
					{Name: "buffer", Type: "() => string", Description: "Return body as base64-encoded string"},
				},
			},
			{
				Name:        "BasicAuth",
				Description: "Basic authentication credentials",
				Fields: []schema.ParamSchema{
					{Name: "user", Type: "string", Description: "Username"},
					{Name: "pass", Type: "string", Description: "Password"},
				},
			},
			{
				Name:        "HTTPOptions",
				Description: "Options for HTTP requests",
				Fields: []schema.ParamSchema{
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Request headers", Optional: true},
					{Name: "timeout", Type: "number", Description: "Request timeout in milliseconds", Optional: true},
					{Name: "followRedirects", Type: "boolean", Description: "Follow redirects (default: true)", Optional: true},
					{Name: "maxRedirects", Type: "number", Description: "Maximum number of redirects to follow (default: 10)", Optional: true},
					{Name: "proxy", Type: "string", Description: "Proxy URL (e.g., 'http://proxy:8080')", Optional: true},
					{Name: "basicAuth", Type: "BasicAuth", Description: "Basic authentication credentials", Optional: true},
					{Name: "bearerToken", Type: "string", Description: "Bearer token for authorization", Optional: true},
					{Name: "skipTLSVerify", Type: "boolean", Description: "Skip TLS certificate verification (use with caution)", Optional: true},
				},
			},
			{
				Name:        "FormField",
				Description: "Form field for multipart uploads",
				Fields: []schema.ParamSchema{
					{Name: "value", Type: "string", Description: "String value for the field", Optional: true},
					{Name: "filename", Type: "string", Description: "Filename for file uploads", Optional: true},
					{Name: "content", Type: "string", Description: "Base64-encoded file content", Optional: true},
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
				Description: "Make a POST request with JSON body",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "body", Type: "any", Description: "Request body (will be JSON encoded)"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "put",
				Description: "Make a PUT request with JSON body",
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
			{
				Name:        "patch",
				Description: "Make a PATCH request with JSON body",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "body", Type: "any", Description: "Request body (will be JSON encoded)"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "head",
				Description: "Make a HEAD request (returns only headers, no body)",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "options",
				Description: "Make an OPTIONS request (check CORS and available methods)",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "postForm",
				Description: "Send multipart/form-data request with file support",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "fields", Type: "{ [key: string]: string | FormField }", Description: "Form fields. Use string for simple values, FormField object for files"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "postFormURLEncoded",
				Description: "Send application/x-www-form-urlencoded request",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to request"},
					{Name: "fields", Type: "{ [key: string]: string }", Description: "Form fields as key-value pairs"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse"},
			},
			{
				Name:        "download",
				Description: "Download file from URL and save to storage",
				Params: []schema.ParamSchema{
					{Name: "url", Type: "string", Description: "URL to download from"},
					{Name: "storagePath", Type: "string", Description: "Path in storage where file will be saved"},
					{Name: "options", Type: "HTTPOptions", Description: "Request options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "HTTPResponse", Description: "Response with body containing the saved file path"},
			},
		},
	}
}

// GetHTTPSchema returns the http schema (static version)
func GetHTTPSchema() schema.ModuleSchema {
	return (&HTTPModule{}).GetSchema()
}
