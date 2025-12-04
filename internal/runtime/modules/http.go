package modules

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type HTTPModule struct {
	client *http.Client
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
