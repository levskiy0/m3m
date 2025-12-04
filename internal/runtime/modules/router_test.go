package modules

import (
	"testing"

	"github.com/dop251/goja"
)

func setupRouterTest(t *testing.T) (*RouterModule, *goja.Runtime) {
	t.Helper()

	vm := goja.New()
	router := NewRouterModule()
	router.SetVM(vm)

	// Register router in VM
	vm.Set("router", map[string]interface{}{
		"get":      router.Get,
		"post":     router.Post,
		"put":      router.Put,
		"delete":   router.Delete,
		"response": router.Response,
	})

	return router, vm
}

func TestRouterModule_Get(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/health', function(ctx) {
			return router.response(200, { status: 'ok' });
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/health",
	}

	resp, err := router.Handle("GET", "/health", ctx)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if resp.Status != 200 {
		t.Errorf("Response status = %d, want 200", resp.Status)
	}

	body, ok := resp.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Response body is not a map: %T", resp.Body)
	}

	if body["status"] != "ok" {
		t.Errorf("Response body status = %v, want 'ok'", body["status"])
	}
}

func TestRouterModule_Post(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.post('/users', function(ctx) {
			return router.response(201, { id: 1, name: ctx.body.name });
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "POST",
		Path:   "/users",
		Body:   map[string]interface{}{"name": "John"},
	}

	resp, err := router.Handle("POST", "/users", ctx)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if resp.Status != 201 {
		t.Errorf("Response status = %d, want 201", resp.Status)
	}
}

func TestRouterModule_Put(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.put('/users/:id', function(ctx) {
			return router.response(200, { id: ctx.params.id, updated: true });
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "PUT",
		Path:   "/users/123",
	}

	resp, err := router.Handle("PUT", "/users/123", ctx)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if resp.Status != 200 {
		t.Errorf("Response status = %d, want 200", resp.Status)
	}

	body, ok := resp.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Response body is not a map: %T", resp.Body)
	}

	if body["id"] != "123" {
		t.Errorf("Response body id = %v, want '123'", body["id"])
	}
}

func TestRouterModule_Delete(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.delete('/users/:id', function(ctx) {
			return router.response(204, null);
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "DELETE",
		Path:   "/users/123",
	}

	resp, err := router.Handle("DELETE", "/users/123", ctx)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if resp.Status != 204 {
		t.Errorf("Response status = %d, want 204", resp.Status)
	}
}

func TestRouterModule_PathParams(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/users/:userId/posts/:postId', function(ctx) {
			return router.response(200, {
				userId: ctx.params.userId,
				postId: ctx.params.postId
			});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/users/42/posts/7",
	}

	resp, err := router.Handle("GET", "/users/42/posts/7", ctx)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	body, ok := resp.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Response body is not a map: %T", resp.Body)
	}

	if body["userId"] != "42" {
		t.Errorf("userId = %v, want '42'", body["userId"])
	}
	if body["postId"] != "7" {
		t.Errorf("postId = %v, want '7'", body["postId"])
	}
}

func TestRouterModule_QueryParams(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/search', function(ctx) {
			return router.response(200, {
				q: ctx.query.q,
				page: ctx.query.page
			});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/search",
		Query: map[string]string{
			"q":    "test",
			"page": "1",
		},
	}

	resp, err := router.Handle("GET", "/search", ctx)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	body, ok := resp.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Response body is not a map: %T", resp.Body)
	}

	if body["q"] != "test" {
		t.Errorf("q = %v, want 'test'", body["q"])
	}
}

func TestRouterModule_Headers(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/auth', function(ctx) {
			return router.response(200, {
				auth: ctx.headers['Authorization']
			});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/auth",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
		},
	}

	resp, err := router.Handle("GET", "/auth", ctx)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	body, ok := resp.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Response body is not a map: %T", resp.Body)
	}

	if body["auth"] != "Bearer token123" {
		t.Errorf("auth = %v, want 'Bearer token123'", body["auth"])
	}
}

func TestRouterModule_RouteNotFound(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/exists', function(ctx) {
			return router.response(200, {});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/not-exists",
	}

	_, err = router.Handle("GET", "/not-exists", ctx)
	if err == nil {
		t.Error("Handle() should return error for non-existent route")
	}
}

func TestRouterModule_MethodNotAllowed(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/only-get', function(ctx) {
			return router.response(200, {});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{
		Method: "POST",
		Path:   "/only-get",
	}

	_, err = router.Handle("POST", "/only-get", ctx)
	if err == nil {
		t.Error("Handle() should return error for wrong method")
	}
}

func TestRouterModule_HasRoutes(t *testing.T) {
	router, vm := setupRouterTest(t)

	if router.HasRoutes() {
		t.Error("HasRoutes() should return false for empty router")
	}

	_, err := vm.RunString(`
		router.get('/test', function(ctx) {
			return router.response(200, {});
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	if !router.HasRoutes() {
		t.Error("HasRoutes() should return true after adding route")
	}
}

func TestRouterModule_Response(t *testing.T) {
	router, _ := setupRouterTest(t)

	resp := router.Response(201, map[string]interface{}{"id": 1})

	if resp["status"] != 201 {
		t.Errorf("Response status = %v, want 201", resp["status"])
	}

	body, ok := resp["body"].(map[string]interface{})
	if !ok {
		t.Fatalf("Response body is not a map: %T", resp["body"])
	}

	if body["id"] != 1 {
		t.Errorf("Response body id = %v, want 1", body["id"])
	}
}

func TestRouterModule_MultipleRoutes(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/a', function(ctx) {
			return router.response(200, { route: 'a' });
		});
		router.get('/b', function(ctx) {
			return router.response(200, { route: 'b' });
		});
		router.post('/a', function(ctx) {
			return router.response(201, { route: 'post-a' });
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register routes: %v", err)
	}

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/a", "a"},
		{"GET", "/b", "b"},
		{"POST", "/a", "post-a"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			ctx := &RequestContext{Method: tt.method, Path: tt.path}
			resp, err := router.Handle(tt.method, tt.path, ctx)
			if err != nil {
				t.Fatalf("Handle() error: %v", err)
			}

			body := resp.Body.(map[string]interface{})
			if body["route"] != tt.expected {
				t.Errorf("route = %v, want %s", body["route"], tt.expected)
			}
		})
	}
}

func TestRouterModule_PlainResponse(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/plain', function(ctx) {
			return { message: 'hello' };
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	ctx := &RequestContext{Method: "GET", Path: "/plain"}
	resp, err := router.Handle("GET", "/plain", ctx)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	// When returning a plain object without router.response(), status should be 200
	if resp.Status != 200 {
		t.Errorf("Response status = %d, want 200", resp.Status)
	}
}

func TestRouterModule_CaseInsensitiveMethod(t *testing.T) {
	router, vm := setupRouterTest(t)

	_, err := vm.RunString(`
		router.get('/test', function(ctx) {
			return router.response(200, { ok: true });
		});
	`)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Test with lowercase method
	ctx := &RequestContext{Method: "get", Path: "/test"}
	resp, err := router.Handle("get", "/test", ctx)
	if err != nil {
		t.Fatalf("Handle() with lowercase method error: %v", err)
	}

	if resp.Status != 200 {
		t.Errorf("Response status = %d, want 200", resp.Status)
	}
}
