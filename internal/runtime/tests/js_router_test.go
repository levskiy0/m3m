package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"m3m/internal/runtime/modules"
)

func TestJS_Router_BasicRoutes(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/test", function(ctx) {
			return {status: 200, body: "test response"};
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/test"}
	resp, err := routerModule.Handle("GET", "/test", ctx)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("Expected status 200, got %d", resp.Status)
	}
	if resp.Body != "test response" {
		t.Errorf("Expected 'test response', got %v", resp.Body)
	}
}

func TestJS_Router_PathParams(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/users/:id", function(ctx) {
			return {status: 200, body: {userId: ctx.params.id}};
		});

		router.get("/posts/:postId/comments/:commentId", function(ctx) {
			return {
				status: 200,
				body: {postId: ctx.params.postId, commentId: ctx.params.commentId}
			};
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/users/123"}
	resp, err := routerModule.Handle("GET", "/users/123", ctx)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	body := resp.Body.(map[string]interface{})
	if body["userId"] != "123" {
		t.Errorf("Expected userId=123, got %v", body["userId"])
	}

	ctx = &modules.RequestContext{Method: "GET", Path: "/posts/10/comments/5"}
	resp, err = routerModule.Handle("GET", "/posts/10/comments/5", ctx)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	body = resp.Body.(map[string]interface{})
	if body["postId"] != "10" || body["commentId"] != "5" {
		t.Errorf("Params not extracted correctly: %v", body)
	}
}

func TestJS_Router_QueryParams(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/search", function(ctx) {
			return {
				status: 200,
				body: {
					query: ctx.query.q || "empty",
					page: ctx.query.page || "1"
				}
			};
		});
	`)

	ctx := &modules.RequestContext{
		Method: "GET",
		Path:   "/search",
		Query:  map[string]string{"q": "test", "page": "5"},
	}
	resp, err := routerModule.Handle("GET", "/search", ctx)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	body := resp.Body.(map[string]interface{})
	if body["query"] != "test" {
		t.Errorf("Expected query=test, got %v", body["query"])
	}
}

func TestJS_Router_PostBody(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.post("/users", function(ctx) {
			if (!ctx.body) {
				return router.response(400, {error: "body required"});
			}
			return router.response(201, {
				created: true,
				name: ctx.body.name,
				email: ctx.body.email
			});
		});
	`)

	ctx := &modules.RequestContext{
		Method: "POST",
		Path:   "/users",
		Body:   map[string]interface{}{"name": "John", "email": "john@test.com"},
	}
	resp, err := routerModule.Handle("POST", "/users", ctx)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if resp.Status != 201 {
		t.Errorf("Expected status 201, got %d", resp.Status)
	}
	body := resp.Body.(map[string]interface{})
	if body["name"] != "John" {
		t.Errorf("Expected name=John, got %v", body["name"])
	}
}

func TestJS_Router_NotFound(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/exists", function(ctx) {
			return {status: 200, body: "ok"};
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/not-exists"}
	_, err := routerModule.Handle("GET", "/not-exists", ctx)
	if err == nil {
		t.Error("Should return error for non-existing route")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

func TestJS_Router_MethodNotAllowed(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/only-get", function(ctx) {
			return {status: 200, body: "ok"};
		});
	`)

	ctx := &modules.RequestContext{Method: "POST", Path: "/only-get"}
	_, err := routerModule.Handle("POST", "/only-get", ctx)
	if err == nil {
		t.Error("Should return error for wrong method")
	}
}

func TestJS_Router_HandlerThrows(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/crash", function(ctx) {
			throw new Error("Intentional crash");
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/crash"}
	_, err := routerModule.Handle("GET", "/crash", ctx)
	if err == nil {
		t.Error("Expected error when handler throws")
	}
}

func TestJS_Router_HandlerReturnsUndefined(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/undefined", function(ctx) {
			// no return
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/undefined"}
	resp, err := routerModule.Handle("GET", "/undefined", ctx)
	if err != nil {
		t.Logf("Got error (acceptable): %v", err)
		return
	}
	t.Logf("Got response: %+v", resp)
}

func TestJS_Router_HandlerReturnsNull(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/null", function(ctx) {
			return null;
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/null"}
	resp, err := routerModule.Handle("GET", "/null", ctx)
	if err != nil {
		t.Logf("Got error (acceptable): %v", err)
		return
	}
	t.Logf("Got response: %+v", resp)
}

func TestJS_Router_InfiniteLoop(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/loop", function(ctx) {
			var count = 0;
			while (count < 1000000) { count++; }
			return router.response(200, {count: count});
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/loop"}
	done := make(chan bool, 1)

	go func() {
		resp, _ := routerModule.Handle("GET", "/loop", ctx)
		if resp != nil {
			done <- true
		}
	}()

	select {
	case <-done:
		// OK
	case <-time.After(5 * time.Second):
		t.Error("Handler took too long - possible infinite loop issue")
	}
}

func TestJS_Router_SequentialMultipleRequests(t *testing.T) {
	// Note: Goja VM is not thread-safe, so concurrent access from multiple
	// goroutines is not supported. This test verifies that multiple sequential
	// requests work correctly.
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		var requestCount = 0;
		router.get("/sequential", function(ctx) {
			requestCount++;
			return router.response(200, {
				id: ctx.query.id || "unknown",
				count: requestCount
			});
		});
	`)

	// Run 100 sequential requests
	for i := 0; i < 100; i++ {
		ctx := &modules.RequestContext{
			Method: "GET",
			Path:   "/sequential",
			Query:  map[string]string{"id": fmt.Sprintf("%d", i)},
		}
		resp, err := routerModule.Handle("GET", "/sequential", ctx)
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
		if resp.Status != 200 {
			t.Errorf("Request %d: Expected 200, got %d", i, resp.Status)
		}
	}

	// Verify counter
	ctx := &modules.RequestContext{Method: "GET", Path: "/sequential"}
	resp, _ := routerModule.Handle("GET", "/sequential", ctx)
	body := resp.Body.(map[string]interface{})
	count := body["count"]
	// Handle both int64 and float64 since goja can return either
	var countVal float64
	switch v := count.(type) {
	case float64:
		countVal = v
	case int64:
		countVal = float64(v)
	}
	if countVal != 101 {
		t.Errorf("Expected 101 requests, got %v", count)
	}
}

func TestJS_Router_Headers(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/auth", function(ctx) {
			var auth = ctx.headers["Authorization"];
			if (!auth) {
				return router.response(401, {error: "No auth header"});
			}
			return router.response(200, {auth: auth});
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/auth", Headers: map[string]string{}}
	resp, _ := routerModule.Handle("GET", "/auth", ctx)
	if resp.Status != 401 {
		t.Errorf("Expected 401, got %d", resp.Status)
	}

	ctx = &modules.RequestContext{
		Method:  "GET",
		Path:    "/auth",
		Headers: map[string]string{"Authorization": "Bearer token123"},
	}
	resp, _ = routerModule.Handle("GET", "/auth", ctx)
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}
}

func TestJS_Router_ResponseHelper(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/helper", function(ctx) {
			return router.response(201, {message: "created"});
		});
	`)

	ctx := &modules.RequestContext{Method: "GET", Path: "/helper"}
	resp, err := routerModule.Handle("GET", "/helper", ctx)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if resp.Status != 201 {
		t.Errorf("Expected status 201, got %d", resp.Status)
	}
}

func TestJS_Router_AllMethods(t *testing.T) {
	h := NewJSTestHelper(t)
	routerModule := h.SetupRouter()

	h.MustRun(t, `
		router.get("/resource", function(ctx) {
			return router.response(200, {method: "GET"});
		});
		router.post("/resource", function(ctx) {
			return router.response(201, {method: "POST"});
		});
		router.put("/resource", function(ctx) {
			return router.response(200, {method: "PUT"});
		});
		router.delete("/resource", function(ctx) {
			return router.response(204, {method: "DELETE"});
		});
	`)

	methods := []struct {
		method string
		status int
	}{
		{"GET", 200},
		{"POST", 201},
		{"PUT", 200},
		{"DELETE", 204},
	}

	for _, m := range methods {
		t.Run(m.method, func(t *testing.T) {
			ctx := &modules.RequestContext{Method: m.method, Path: "/resource"}
			resp, err := routerModule.Handle(m.method, "/resource", ctx)
			if err != nil {
				t.Fatalf("Handle failed: %v", err)
			}
			if resp.Status != m.status {
				t.Errorf("Expected status %d, got %d", m.status, resp.Status)
			}
		})
	}
}
