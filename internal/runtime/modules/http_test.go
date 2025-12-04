package modules

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPModule_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Get(server.URL)

	if resp.Status != 200 {
		t.Errorf("Get() status = %d, want 200", resp.Status)
	}
	if resp.Body != `{"status":"ok"}` {
		t.Errorf("Get() body = %q, want {\"status\":\"ok\"}", resp.Body)
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Get() Content-Type header = %q, want application/json", resp.Headers["Content-Type"])
	}
}

func TestHTTPModule_Get_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Errorf("Authorization header = %q, want 'Bearer token123'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Custom") != "custom-value" {
			t.Errorf("X-Custom header = %q, want 'custom-value'", r.Header.Get("X-Custom"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	opts := &HTTPOptions{
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"X-Custom":      "custom-value",
		},
	}
	resp := httpMod.Get(server.URL, opts)

	if resp.Status != 200 {
		t.Errorf("Get() with headers status = %d, want 200", resp.Status)
	}
}

func TestHTTPModule_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)

		if data["name"] != "test" {
			t.Errorf("body.name = %v, want 'test'", data["name"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Post(server.URL, map[string]interface{}{"name": "test"})

	if resp.Status != 201 {
		t.Errorf("Post() status = %d, want 201", resp.Status)
	}
	if resp.Body != `{"id":1}` {
		t.Errorf("Post() body = %q, want {\"id\":1}", resp.Body)
	}
}

func TestHTTPModule_Post_WithNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != 0 {
			t.Errorf("Expected empty body, got %q", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Post(server.URL, nil)

	if resp.Status != 200 {
		t.Errorf("Post() with nil body status = %d, want 200", resp.Status)
	}
}

func TestHTTPModule_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)

		if data["id"] != float64(1) {
			t.Errorf("body.id = %v, want 1", data["id"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"updated":true}`))
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Put(server.URL, map[string]interface{}{"id": 1, "name": "updated"})

	if resp.Status != 200 {
		t.Errorf("Put() status = %d, want 200", resp.Status)
	}
}

func TestHTTPModule_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Delete(server.URL)

	if resp.Status != 204 {
		t.Errorf("Delete() status = %d, want 204", resp.Status)
	}
}

func TestHTTPModule_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	opts := &HTTPOptions{
		Timeout: 50, // 50ms timeout
	}
	resp := httpMod.Get(server.URL, opts)

	if resp.Status != 0 {
		t.Errorf("Get() with timeout should have status 0, got %d", resp.Status)
	}
	if resp.StatusText == "" {
		t.Errorf("Get() with timeout should have error in StatusText")
	}
}

func TestHTTPModule_InvalidURL(t *testing.T) {
	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Get("://invalid-url")

	if resp.Status != 0 {
		t.Errorf("Get() with invalid URL should have status 0, got %d", resp.Status)
	}
	if resp.StatusText == "" {
		t.Errorf("Get() with invalid URL should have error in StatusText")
	}
}

func TestHTTPModule_ConnectionRefused(t *testing.T) {
	httpMod := NewHTTPModule(1 * time.Second)
	// Connect to a port that's unlikely to be listening
	resp := httpMod.Get("http://127.0.0.1:59999")

	if resp.Status != 0 {
		t.Errorf("Get() to closed port should have status 0, got %d", resp.Status)
	}
}

func TestHTTPModule_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", 200},
		{"Created", 201},
		{"Bad Request", 400},
		{"Unauthorized", 401},
		{"Forbidden", 403},
		{"Not Found", 404},
		{"Internal Server Error", 500},
		{"Service Unavailable", 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			httpMod := NewHTTPModule(30 * time.Second)
			resp := httpMod.Get(server.URL)

			if resp.Status != tt.statusCode {
				t.Errorf("Get() status = %d, want %d", resp.Status, tt.statusCode)
			}
		})
	}
}

func TestHTTPModule_ResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.Header().Set("X-Another", "another-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Get(server.URL)

	if resp.Headers["X-Custom-Header"] != "custom-value" {
		t.Errorf("Response header X-Custom-Header = %q, want 'custom-value'", resp.Headers["X-Custom-Header"])
	}
	if resp.Headers["X-Another"] != "another-value" {
		t.Errorf("Response header X-Another = %q, want 'another-value'", resp.Headers["X-Another"])
	}
}

func TestHTTPModule_LargeResponse(t *testing.T) {
	largeBody := make([]byte, 1024*1024) // 1MB
	for i := range largeBody {
		largeBody[i] = 'x'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(largeBody)
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Get(server.URL)

	if resp.Status != 200 {
		t.Errorf("Get() large response status = %d, want 200", resp.Status)
	}
	if len(resp.Body) != len(largeBody) {
		t.Errorf("Get() large response body length = %d, want %d", len(resp.Body), len(largeBody))
	}
}

func TestHTTPModule_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	resp := httpMod.Get(server.URL)

	if resp.Status != 204 {
		t.Errorf("Get() empty response status = %d, want 204", resp.Status)
	}
	if resp.Body != "" {
		t.Errorf("Get() empty response body = %q, want empty", resp.Body)
	}
}

func TestHTTPModule_CustomContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Content-Type = %q, want 'text/plain'", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpMod := NewHTTPModule(30 * time.Second)
	opts := &HTTPOptions{
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}
	resp := httpMod.Post(server.URL, map[string]string{"test": "data"}, opts)

	// Custom Content-Type should override default
	if resp.Status != 200 {
		t.Errorf("Post() with custom Content-Type status = %d, want 200", resp.Status)
	}
}
