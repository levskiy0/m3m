package modules

import (
	"enco
	"fmt"
n"
	"fmt"
	
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
)

// TestHelper creates a Goja VM with modules registered for testing
type TestHelper struct {
	VM     *goja.Runtime
	Logger *LoggerModule
	logs   []string
	mu     sync.Mutex
}

// NewTestHelper creates a test VM with all modules registered
func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	h := &TestHelper{
		VM:   vm,
		logs: []string{},
	}

	// Mock logger that captures output
	h.registerMockLogger(vm)
	h.registerModules(vm)

	return h
}

func (h *TestHelper) registerMockLogger(vm *goja.Runtime) {
	capture := func(level string) func(args ...interface{}) {
		return func(args ...interface{}) {
			h.mu.Lock()
			defer h.mu.Unlock()
			msg := ""
			for _, a := range args {
				msg += toString(a) + " "
			}
			h.logs = append(h.logs, "["+level+"] "+strings.TrimSpace(msg))
		}
	}

	vm.Set("logger", map[string]interface{}{
		"debug": capture("DEBUG"),
		"info":  capture("INFO"),
		"warn":  capture("WARN"),
		"error": capture("ERROR"),
	})

	vm.Set("console", map[string]interface{}{
		"log":   capture("INFO"),
		"info":  capture("INFO"),
		"warn":  capture("WARN"),
		"error": capture("ERROR"),
		"debug": capture("DEBUG"),
	})
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *TestHelper) registerModules(vm *goja.Runtime) {
	// Crypto
	cryptoModule := NewCryptoModule()
	vm.Set("crypto", map[string]interface{}{
		"md5":         cryptoModule.MD5,
		"sha256":      cryptoModule.SHA256,
		"randomBytes": cryptoModule.RandomBytes,
	})

	// Encoding
	encodingModule := NewEncodingModule()
	vm.Set("encoding", map[string]interface{}{
		"base64Encode":  encodingModule.Base64Encode,
		"base64Decode":  encodingModule.Base64Decode,
		"jsonParse":     encodingModule.JSONParse,
		"jsonStringify": encodingModule.JSONStringify,
		"urlEncode":     encodingModule.URLEncode,
		"urlDecode":     encodingModule.URLDecode,
	})

	// Utils
	utilsModule := NewUtilsModule()
	vm.Set("utils", map[string]interface{}{
		"sleep":        utilsModule.Sleep,
		"random":       utilsModule.Random,
		"randomInt":    utilsModule.RandomInt,
		"uuid":         utilsModule.UUID,
		"slugify":      utilsModule.Slugify,
		"truncate":     utilsModule.Truncate,
		"capitalize":   utilsModule.Capitalize,
		"regexMatch":   utilsModule.RegexMatch,
		"regexReplace": utilsModule.RegexReplace,
		"formatDate":   utilsModule.FormatDate,
		"parseDate":    utilsModule.ParseDate,
		"timestamp":    utilsModule.Timestamp,
	})

	// Router
	routerModule := NewRouterModule()
	routerModule.SetVM(vm)
	vm.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"put":      routerModule.Put,
		"delete":   routerModule.Delete,
		"response": routerModule.Response,
	})

	// Env (mock)
	envModule := NewEnvModule(map[string]interface{}{
		"TEST_VAR":    "test_value",
		"NUMBER_VAR":  "42",
		"BOOL_VAR":    "true",
		"FLOAT_VAR":   "3.14",
		"EMPTY_VAR":   "",
		"JSON_VAR":    `{"key": "value"}`,
		"SPACES_VAR":  "  trimmed  ",
		"SPECIAL_VAR": "hello\nworld",
	})
	vm.Set("env", map[string]interface{}{
		"get":       envModule.Get,
		"has":       envModule.Has,
		"keys":      envModule.Keys,
		"getString": envModule.GetString,
		"getInt":    envModule.GetInt,
		"getFloat":  envModule.GetFloat,
		"getBool":   envModule.GetBool,
		"getAll":    envModule.GetAll,
	})

	// Delayed
	delayedModule := NewDelayedModule(5)
	vm.Set("delayed", map[string]interface{}{
		"run": delayedModule.Run,
	})
}

// Run executes JS code and returns the result
func (h *TestHelper) Run(code string) (goja.Value, error) {
	return h.VM.RunString(code)
}

// MustRun executes JS code and fails test on error
func (h *TestHelper) MustRun(t *testing.T, code string) goja.Value {
	t.Helper()
	result, err := h.Run(code)
	if err != nil {
		t.Fatalf("JS execution failed: %v\nCode: %s", err, code)
	}
	return result
}

// GetLogs returns captured log messages
func (h *TestHelper) GetLogs() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string{}, h.logs...)
}

// ClearLogs clears captured log messages
func (h *TestHelper) ClearLogs() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logs = []string{}
}

// ============== CRYPTO MODULE TESTS ==============

func TestJS_Crypto_MD5(t *testing.T) {
	h := NewTestHelper(t)
	cryptoModule := NewCryptoModule()

	tests := []struct {
		name  string
		input string
	}{
		{"simple string", "hello"},
		{"empty string", ""},
		{"unicode", "привет"},
		{"special chars", "<script>alert('xss')</script>"},
		{"numbers", "12345"},
		{"whitespace", "  spaces  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare JS result with Go module directly
			expected := cryptoModule.MD5(tt.input)
			result := h.MustRun(t, `crypto.md5("`+tt.input+`")`)
			if result.String() != expected {
				t.Errorf("MD5(%q): JS=%s, Go=%s", tt.input, result.String(), expected)
			}
		})
	}
}

func TestJS_Crypto_MD5_UnexpectedInput(t *testing.T) {
	h := NewTestHelper(t)

	// Test with number - should be converted to string by JS
	result, err := h.Run(`crypto.md5(123)`)
	if err != nil {
		t.Fatalf("Should handle number input: %v", err)
	}
	if result.String() == "" {
		t.Error("Should return a hash for number input")
	}

	// Test with null
	result, err = h.Run(`crypto.md5(null)`)
	if err != nil {
		t.Fatalf("Should handle null input: %v", err)
	}

	// Test with undefined
	result, err = h.Run(`crypto.md5(undefined)`)
	if err != nil {
		t.Fatalf("Should handle undefined input: %v", err)
	}

	// Test with object - JS converts to [object Object]
	result, err = h.Run(`crypto.md5({foo: "bar"})`)
	if err != nil {
		t.Fatalf("Should handle object input: %v", err)
	}

	// Test with array
	result, err = h.Run(`crypto.md5([1, 2, 3])`)
	if err != nil {
		t.Fatalf("Should handle array input: %v", err)
	}
}

func TestJS_Crypto_SHA256(t *testing.T) {
	h := NewTestHelper(t)

	result := h.MustRun(t, `crypto.sha256("hello")`)
	expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if result.String() != expected {
		t.Errorf("Expected %s, got %s", expected, result.String())
	}
}

func TestJS_Crypto_RandomBytes(t *testing.T) {
	h := NewTestHelper(t)

	// Test correct length
	result := h.MustRun(t, `crypto.randomBytes(16)`)
	if len(result.String()) != 32 { // hex encoding doubles length
		t.Errorf("Expected 32 chars, got %d", len(result.String()))
	}

	// Test uniqueness
	result1 := h.MustRun(t, `crypto.randomBytes(16)`)
	result2 := h.MustRun(t, `crypto.randomBytes(16)`)
	if result1.String() == result2.String() {
		t.Error("RandomBytes should produce unique values")
	}

	// Test zero length
	result = h.MustRun(t, `crypto.randomBytes(0)`)
	if result.String() != "" {
		t.Errorf("Zero length should return empty string, got %s", result.String())
	}

	// Test negative length
	result = h.MustRun(t, `crypto.randomBytes(-5)`)
	if result.String() != "" {
		t.Errorf("Negative length should return empty string, got %s", result.String())
	}
}

// ============== ENCODING MODULE TESTS ==============

func TestJS_Encoding_Base64(t *testing.T) {
	h := NewTestHelper(t)

	// Encode and decode roundtrip
	code := `
		var encoded = encoding.base64Encode("Hello, World!");
		var decoded = encoding.base64Decode(encoded);
		decoded;
	`
	result := h.MustRun(t, code)
	if result.String() != "Hello, World!" {
		t.Errorf("Base64 roundtrip failed: got %s", result.String())
	}
}

func TestJS_Encoding_Base64_InvalidInput(t *testing.T) {
	h := NewTestHelper(t)

	// Invalid base64 should return empty string
	result := h.MustRun(t, `encoding.base64Decode("not-valid-base64!!!")`)
	if result.String() != "" {
		t.Errorf("Invalid base64 should return empty string, got %s", result.String())
	}
}

func TestJS_Encoding_JSON(t *testing.T) {
	h := NewTestHelper(t)

	// Parse and stringify roundtrip
	code := `
		var obj = {name: "test", value: 42, nested: {foo: "bar"}};
		var str = encoding.jsonStringify(obj);
		var parsed = encoding.jsonParse(str);
		parsed.name + "-" + parsed.value + "-" + parsed.nested.foo;
	`
	result := h.MustRun(t, code)
	if result.String() != "test-42-bar" {
		t.Errorf("JSON roundtrip failed: got %s", result.String())
	}
}

func TestJS_Encoding_JSON_InvalidInput(t *testing.T) {
	h := NewTestHelper(t)

	// Invalid JSON should return null
	result := h.MustRun(t, `encoding.jsonParse("not valid json {")`)
	if !goja.IsNull(result) && !goja.IsUndefined(result) && result.Export() != nil {
		t.Errorf("Invalid JSON should return null, got %v", result.Export())
	}
}

func TestJS_Encoding_URL(t *testing.T) {
	h := NewTestHelper(t)

	// Encode special characters
	result := h.MustRun(t, `encoding.urlEncode("hello world & foo=bar")`)
	if result.String() != "hello+world+%26+foo%3Dbar" {
		t.Errorf("URL encode failed: got %s", result.String())
	}

	// Decode roundtrip
	code := `encoding.urlDecode(encoding.urlEncode("тест?foo=bar&baz=1"))`
	result = h.MustRun(t, code)
	if result.String() != "тест?foo=bar&baz=1" {
		t.Errorf("URL roundtrip failed: got %s", result.String())
	}
}

// ============== UTILS MODULE TESTS ==============

func TestJS_Utils_UUID(t *testing.T) {
	h := NewTestHelper(t)

	// Test format
	result := h.MustRun(t, `utils.uuid()`)
	uuid := result.String()

	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	if len(uuid) != 36 {
		t.Errorf("UUID should be 36 chars, got %d: %s", len(uuid), uuid)
	}
	if uuid[14] != '4' {
		t.Errorf("UUID v4 should have '4' at position 14, got %c", uuid[14])
	}

	// Test uniqueness
	uuid1 := h.MustRun(t, `utils.uuid()`).String()
	uuid2 := h.MustRun(t, `utils.uuid()`).String()
	if uuid1 == uuid2 {
		t.Error("UUIDs should be unique")
	}
}

func TestJS_Utils_Slugify(t *testing.T) {
	h := NewTestHelper(t)

	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"  Spaces  Around  ", "spaces-around"},
		{"Special!@#$%Chars", "specialchars"},
		{"Привет Мир", ""}, // Non-ASCII removed
		{"Mixed 123 Numbers", "mixed-123-numbers"},
		{"already-slugified", "already-slugified"},
		{"UPPERCASE", "uppercase"},
		{"multiple---dashes", "multiple-dashes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `utils.slugify("`+tt.input+`")`)
			if result.String() != tt.expected {
				t.Errorf("slugify(%q) = %q, expected %q", tt.input, result.String(), tt.expected)
			}
		})
	}
}

func TestJS_Utils_RandomInt(t *testing.T) {
	h := NewTestHelper(t)

	// Test range
	for i := 0; i < 100; i++ {
		result := h.MustRun(t, `utils.randomInt(5, 10)`)
		val := int(result.ToInteger())
		if val < 5 || val >= 10 {
			t.Errorf("randomInt(5, 10) = %d, expected [5, 10)", val)
		}
	}

	// Test min == max
	result := h.MustRun(t, `utils.randomInt(5, 5)`)
	if result.ToInteger() != 5 {
		t.Errorf("randomInt(5, 5) should return 5, got %d", result.ToInteger())
	}

	// Test min > max
	result = h.MustRun(t, `utils.randomInt(10, 5)`)
	if result.ToInteger() != 10 {
		t.Errorf("randomInt(10, 5) should return min (10), got %d", result.ToInteger())
	}
}

func TestJS_Utils_Truncate(t *testing.T) {
	h := NewTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"short text", `utils.truncate("hello", 10)`, "hello"},
		{"exact length", `utils.truncate("hello", 5)`, "hello"},
		{"truncated", `utils.truncate("hello world", 5)`, "hello..."},
		{"zero length", `utils.truncate("hello", 0)`, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

func TestJS_Utils_Timestamp(t *testing.T) {
	h := NewTestHelper(t)

	before := time.Now().UnixMilli()
	result := h.MustRun(t, `utils.timestamp()`)
	after := time.Now().UnixMilli()

	ts := result.ToInteger()
	if ts < before || ts > after {
		t.Errorf("Timestamp %d should be between %d and %d", ts, before, after)
	}
}

func TestJS_Utils_FormatDate(t *testing.T) {
	h := NewTestHelper(t)

	// Test with known timestamp (2024-01-15 12:30:45 UTC)
	ts := int64(1705322445000)

	result := h.MustRun(t, `utils.formatDate(1705322445000, "YYYY-MM-DD")`)
	if result.String() != "2024-01-15" {
		t.Errorf("formatDate failed: got %s", result.String())
	}

	// Test with invalid timestamp
	result = h.MustRun(t, `utils.formatDate(0, "YYYY-MM-DD")`)
	if result.String() != "1970-01-01" {
		t.Errorf("formatDate(0) should be Unix epoch, got %s", result.String())
	}

	_ = ts
}

// ============== ENV MODULE TESTS ==============

func TestJS_Env_Get(t *testing.T) {
	h := NewTestHelper(t)

	// Existing var
	result := h.MustRun(t, `env.get("TEST_VAR")`)
	if result.String() != "test_value" {
		t.Errorf("Expected 'test_value', got %s", result.String())
	}

	// Non-existing var - returns nil from Go which JS sees as null/undefined
	result = h.MustRun(t, `env.get("NOT_EXISTS")`)
	if !goja.IsUndefined(result) && !goja.IsNull(result) && result.Export() != nil {
		t.Errorf("Expected undefined/null for non-existing var, got %v", result.Export())
	}
}

func TestJS_Env_Has(t *testing.T) {
	h := NewTestHelper(t)

	result := h.MustRun(t, `env.has("TEST_VAR")`)
	if !result.ToBoolean() {
		t.Error("has(TEST_VAR) should return true")
	}

	result = h.MustRun(t, `env.has("NOT_EXISTS")`)
	if result.ToBoolean() {
		t.Error("has(NOT_EXISTS) should return false")
	}

	// Empty var should still exist
	result = h.MustRun(t, `env.has("EMPTY_VAR")`)
	if !result.ToBoolean() {
		t.Error("has(EMPTY_VAR) should return true")
	}
}

func TestJS_Env_TypedGetters(t *testing.T) {
	h := NewTestHelper(t)

	// getInt
	result := h.MustRun(t, `env.getInt("NUMBER_VAR", 0)`)
	if result.ToInteger() != 42 {
		t.Errorf("getInt should return 42, got %d", result.ToInteger())
	}

	result = h.MustRun(t, `env.getInt("NOT_EXISTS", 99)`)
	if result.ToInteger() != 99 {
		t.Errorf("getInt should return default 99, got %d", result.ToInteger())
	}

	result = h.MustRun(t, `env.getInt("TEST_VAR", 0)`) // not a number
	if result.ToInteger() != 0 {
		t.Errorf("getInt with non-numeric should return default, got %d", result.ToInteger())
	}

	// getFloat
	result = h.MustRun(t, `env.getFloat("FLOAT_VAR", 0)`)
	if result.ToFloat() != 3.14 {
		t.Errorf("getFloat should return 3.14, got %f", result.ToFloat())
	}

	// getBool
	result = h.MustRun(t, `env.getBool("BOOL_VAR", false)`)
	if !result.ToBoolean() {
		t.Error("getBool should return true")
	}

	result = h.MustRun(t, `env.getBool("NOT_EXISTS", true)`)
	if !result.ToBoolean() {
		t.Error("getBool should return default true")
	}
}

// ============== ROUTER MODULE TESTS ==============

func TestJS_Router_BasicRoutes(t *testing.T) {
	h := NewTestHelper(t)

	// Register routes
	h.MustRun(t, `
		router.get("/hello", function(ctx) {
			return {status: 200, body: "Hello World"};
		});

		router.post("/users", function(ctx) {
			return {status: 201, body: {created: true}};
		});
	`)

	// Get router module to test Handle
	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)

	// Re-register to our test router
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"put":      routerModule.Put,
		"delete":   routerModule.Delete,
		"response": routerModule.Response,
	})

	h.MustRun(t, `
		router.get("/test", function(ctx) {
			return {status: 200, body: "test response"};
		});
	`)

	// Test route handling
	ctx := &RequestContext{Method: "GET", Path: "/test"}
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
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"put":      routerModule.Put,
		"delete":   routerModule.Delete,
		"response": routerModule.Response,
	})

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

	// Test single param
	ctx := &RequestContext{Method: "GET", Path: "/users/123"}
	resp, err := routerModule.Handle("GET", "/users/123", ctx)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	body := resp.Body.(map[string]interface{})
	if body["userId"] != "123" {
		t.Errorf("Expected userId=123, got %v", body["userId"])
	}

	// Test multiple params
	ctx = &RequestContext{Method: "GET", Path: "/posts/10/comments/5"}
	resp, err = routerModule.Handle("GET", "/posts/10/comments/5", ctx)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	body = resp.Body.(map[string]interface{})
	if body["postId"] != "10" || body["commentId"] != "5" {
		t.Errorf("Params not extracted correctly: %v", body)
	}
}

func TestJS_Router_NotFound(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get": routerModule.Get,
	})

	h.MustRun(t, `
		router.get("/exists", function(ctx) {
			return {status: 200, body: "ok"};
		});
	`)

	ctx := &RequestContext{Method: "GET", Path: "/not-exists"}
	_, err := routerModule.Handle("GET", "/not-exists", ctx)
	if err == nil {
		t.Error("Should return error for non-existing route")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

func TestJS_Router_HandlerException(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get": routerModule.Get,
	})

	h.MustRun(t, `
		router.get("/error", function(ctx) {
			throw new Error("Something went wrong");
		});
	`)

	ctx := &RequestContext{Method: "GET", Path: "/error"}
	_, err := routerModule.Handle("GET", "/error", ctx)
	if err == nil {
		t.Error("Should return error when handler throws")
	}
}

// ============== DELAYED MODULE TESTS ==============

func TestJS_Delayed_Run(t *testing.T) {
	h := NewTestHelper(t)

	var executed bool
	var mu sync.Mutex

	// Register a trackable callback
	h.VM.Set("setExecuted", func() {
		mu.Lock()
		executed = true
		mu.Unlock()
	})

	h.MustRun(t, `
		delayed.run(function() {
			setExecuted();
		});
	`)

	// Wait for async execution
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if !executed {
		t.Error("Delayed callback should have executed")
	}
	mu.Unlock()
}

func TestJS_Delayed_PoolLimiting(t *testing.T) {
	// Test pool limiting at Go level
	delayedModule := NewDelayedModule(2) // Small pool

	var running int32
	var maxRunning int32
	var completed int32
	var mu sync.Mutex

	makeCallback := func() goja.Callable {
		return func(this goja.Value, args ...goja.Value) (goja.Value, error) {
			mu.Lock()
			running++
			if running > maxRunning {
				maxRunning = running
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			running--
			completed++
			mu.Unlock()
			return nil, nil
		}
	}

	for i := 0; i < 5; i++ {
		delayedModule.Run(makeCallback())
	}

	// Wait for completion
	for i := 0; i < 100; i++ {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		if completed >= 5 {
			mu.Unlock()
			break
		}
		mu.Unlock()
	}

	mu.Lock()
	if maxRunning > 2 {
		t.Errorf("Max concurrent should be 2 (pool size), got %d", maxRunning)
	}
	if completed < 5 {
		t.Logf("Note: Only %d/5 completed", completed)
	}
	mu.Unlock()
}

// ============== ERROR HANDLING TESTS ==============

func TestJS_SyntaxError(t *testing.T) {
	h := NewTestHelper(t)

	_, err := h.Run(`function broken( { }`)
	if err == nil {
		t.Error("Should return error for syntax error")
	}
}

func TestJS_RuntimeError(t *testing.T) {
	h := NewTestHelper(t)

	_, err := h.Run(`nonExistentFunction()`)
	if err == nil {
		t.Error("Should return error for undefined function call")
	}
}

func TestJS_TypeError(t *testing.T) {
	h := NewTestHelper(t)

	_, err := h.Run(`null.property`)
	if err == nil {
		t.Error("Should return error for null property access")
	}
}

// ============== COMPLEX INTEGRATION TESTS ==============

func TestJS_ComplexWorkflow(t *testing.T) {
	h := NewTestHelper(t)

	code := `
		// Simulate a real workflow
		var data = {
			users: [
				{id: 1, name: "Alice", email: "alice@test.com"},
				{id: 2, name: "Bob", email: "bob@test.com"}
			]
		};

		// Serialize and deserialize
		var json = encoding.jsonStringify(data);
		var parsed = encoding.jsonParse(json);

		// Process users
		var results = [];
		for (var i = 0; i < parsed.users.length; i++) {
			var user = parsed.users[i];
			results.push({
				id: user.id,
				slug: utils.slugify(user.name),
				hash: crypto.md5(user.email)
			});
		}

		encoding.jsonStringify(results);
	`

	result := h.MustRun(t, code)
	parsed := h.MustRun(t, `encoding.jsonParse('`+result.String()+`')`)

	arr := parsed.Export().([]interface{})
	if len(arr) != 2 {
		t.Errorf("Expected 2 results, got %d", len(arr))
	}

	first := arr[0].(map[string]interface{})
	if first["slug"] != "alice" {
		t.Errorf("Expected slug 'alice', got %v", first["slug"])
	}
}

func TestJS_ServicePattern(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"response": routerModule.Response,
	})

	// Simulate a typical service setup
	code := `
		// Configuration from env
		var apiKey = env.getString("TEST_VAR", "default");

		// API endpoint
		router.get("/api/status", function(ctx) {
			return router.response(200, {
				status: "ok",
				timestamp: utils.timestamp(),
				requestId: utils.uuid()
			});
		});

		router.post("/api/hash", function(ctx) {
			if (!ctx.body || !ctx.body.data) {
				return router.response(400, {error: "data required"});
			}
			return router.response(200, {
				md5: crypto.md5(ctx.body.data),
				sha256: crypto.sha256(ctx.body.data)
			});
		});

		"service registered";
	`

	result := h.MustRun(t, code)
	if result.String() != "service registered" {
		t.Errorf("Expected 'service registered', got %s", result.String())
	}

	// Test the registered routes
	ctx := &RequestContext{Method: "GET", Path: "/api/status"}
	resp, err := routerModule.Handle("GET", "/api/status", ctx)
	if err != nil {
		t.Fatalf("Status endpoint failed: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("Expected status 200, got %d", resp.Status)
	}

	body := resp.Body.(map[string]interface{})
	if body["status"] != "ok" {
		t.Errorf("Expected status ok, got %v", body["status"])
	}

	// Test POST with body
	ctx = &RequestContext{
		Method: "POST",
		Path:   "/api/hash",
		Body:   map[string]interface{}{"data": "test"},
	}
	resp, err = routerModule.Handle("POST", "/api/hash", ctx)
	if err != nil {
		t.Fatalf("Hash endpoint failed: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("Expected status 200, got %d", resp.Status)
	}

	body = resp.Body.(map[string]interface{})
	if body["md5"] != "098f6bcd4621d373cade4e832627b4f6" {
		t.Errorf("MD5 hash incorrect: %v", body["md5"])
	}
}

// ============== HTTP MODULE JS TESTS ==============

func TestJS_HTTP_Get(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "hello"}`))
	}))
	defer server.Close()

	h := NewTestHelper(t)

	// Register HTTP module
	httpModule := NewHTTPModule(30 * time.Second)
	h.VM.Set("http", map[string]interface{}{
		"get":    httpModule.Get,
		"post":   httpModule.Post,
		"put":    httpModule.Put,
		"delete": httpModule.Delete,
	})

	// Test GET request from JS
	code := fmt.Sprintf(`
		var resp = http.get("%s");
		JSON.stringify({
			status: resp.status,
			body: resp.body
		});
	`, server.URL)

	result := h.MustRun(t, code)
	if !strings.Contains(result.String(), `"status":200`) {
		t.Errorf("Expected status 200, got %s", result.String())
	}
	if !strings.Contains(result.String(), `"message":"hello"`) {
		t.Errorf("Expected hello message, got %s", result.String())
	}
}

func TestJS_HTTP_Post_WithBody(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	h := NewTestHelper(t)
	httpModule := NewHTTPModule(30 * time.Second)
	h.VM.Set("http", map[string]interface{}{
		"get":    httpModule.Get,
		"post":   httpModule.Post,
		"put":    httpModule.Put,
		"delete": httpModule.Delete,
	})

	code := fmt.Sprintf(`
		var resp = http.post("%s", {name: "test", value: 42});
		resp.status;
	`, server.URL)

	result := h.MustRun(t, code)
	if result.ToInteger() != 201 {
		t.Errorf("Expected status 201, got %d", result.ToInteger())
	}

	if !strings.Contains(receivedBody, `"name":"test"`) {
		t.Errorf("Body not serialized correctly: %s", receivedBody)
	}
}

func TestJS_HTTP_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := NewTestHelper(t)
	httpModule := NewHTTPModule(30 * time.Second)
	h.VM.Set("http", map[string]interface{}{
		"get":    httpModule.Get,
		"post":   httpModule.Post,
		"put":    httpModule.Put,
		"delete": httpModule.Delete,
	})

	// Test timeout - should fail with status 0
	code := fmt.Sprintf(`
		var resp = http.get("%s", {timeout: 50});
		resp.status;
	`, server.URL)

	result := h.MustRun(t, code)
	if result.ToInteger() != 0 {
		t.Errorf("Expected status 0 on timeout, got %d", result.ToInteger())
	}
}

func TestJS_HTTP_InvalidURL(t *testing.T) {
	h := NewTestHelper(t)
	httpModule := NewHTTPModule(30 * time.Second)
	h.VM.Set("http", map[string]interface{}{
		"get":    httpModule.Get,
		"post":   httpModule.Post,
		"put":    httpModule.Put,
		"delete": httpModule.Delete,
	})

	// Invalid URL should return status 0
	result := h.MustRun(t, `
		var resp = http.get("://invalid");
		resp.status;
	`)

	if result.ToInteger() != 0 {
		t.Errorf("Expected status 0 for invalid URL, got %d", result.ToInteger())
	}
}

func TestJS_HTTP_ConnectionRefused(t *testing.T) {
	h := NewTestHelper(t)
	httpModule := NewHTTPModule(1 * time.Second)
	h.VM.Set("http", map[string]interface{}{
		"get":    httpModule.Get,
		"post":   httpModule.Post,
		"put":    httpModule.Put,
		"delete": httpModule.Delete,
	})

	// Connection refused should return status 0
	result := h.MustRun(t, `
		var resp = http.get("http://127.0.0.1:59999");
		resp.status;
	`)

	if result.ToInteger() != 0 {
		t.Errorf("Expected status 0 for connection refused, got %d", result.ToInteger())
	}
}

// ============== UNEXPECTED INPUT TESTS ==============

func TestJS_Crypto_UnexpectedTypes(t *testing.T) {
	h := NewTestHelper(t)

	tests := []struct {
		name string
		code string
	}{
		{"number to md5", `crypto.md5(12345)`},
		{"boolean to md5", `crypto.md5(true)`},
		{"null to md5", `crypto.md5(null)`},
		{"undefined to md5", `crypto.md5(undefined)`},
		{"object to md5", `crypto.md5({foo: "bar"})`},
		{"array to md5", `crypto.md5([1, 2, 3])`},
		{"empty call", `crypto.md5()`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			// Should not crash - either return result or error
			if err != nil {
				t.Logf("Got error (acceptable): %v", err)
				return
			}
			// Should return some string (hash)
			if result.String() == "" && tt.name != "empty call" {
				t.Logf("Got empty result (acceptable for edge case)")
			}
		})
	}
}

func TestJS_Encoding_UnexpectedTypes(t *testing.T) {
	h := NewTestHelper(t)

	tests := []struct {
		name string
		code string
	}{
		{"number to base64", `encoding.base64Encode(12345)`},
		{"boolean to base64", `encoding.base64Encode(false)`},
		{"null to base64", `encoding.base64Encode(null)`},
		{"undefined to base64", `encoding.base64Encode(undefined)`},
		{"decode invalid", `encoding.base64Decode("!!invalid!!")`},
		{"parse invalid json", `encoding.jsonParse("not json {")`},
		{"stringify circular", `var obj = {}; obj.self = obj; encoding.jsonStringify(obj)`},
		{"url encode number", `encoding.urlEncode(12345)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			// Should not crash
			if err != nil {
				t.Logf("Got error (may be acceptable): %v", err)
				return
			}
			t.Logf("Got result: %v", result.Export())
		})
	}
}

func TestJS_Utils_UnexpectedTypes(t *testing.T) {
	h := NewTestHelper(t)

	tests := []struct {
		name string
		code string
	}{
		{"sleep negative", `utils.sleep(-100); "ok"`},
		{"randomInt swapped", `utils.randomInt(100, 0)`},
		{"randomInt same", `utils.randomInt(5, 5)`},
		{"truncate negative length", `utils.truncate("hello", -5)`},
		{"slugify empty", `utils.slugify("")`},
		{"regex invalid pattern", `utils.regexMatch("test", "[invalid")`},
		{"formatDate invalid", `utils.formatDate("not a number", "YYYY-MM-DD")`},
		{"parseDate invalid", `utils.parseDate("invalid", "YYYY-MM-DD")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			// Should not crash
			if err != nil {
				t.Logf("Got error (may be acceptable): %v", err)
				return
			}
			t.Logf("Got result: %v", result.Export())
		})
	}
}

// ============== FATAL ERROR AND PANIC RECOVERY TESTS ==============

func TestJS_Router_HandlerThrows(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"put":      routerModule.Put,
		"delete":   routerModule.Delete,
		"response": routerModule.Response,
	})

	// Register handler that throws
	h.MustRun(t, `
		router.get("/crash", function(ctx) {
			throw new Error("Intentional crash");
		});
	`)

	// Should return error, not crash
	ctx := &RequestContext{Method: "GET", Path: "/crash"}
	_, err := routerModule.Handle("GET", "/crash", ctx)
	if err == nil {
		t.Error("Expected error when handler throws")
	}
}

func TestJS_Router_HandlerReturnsUndefined(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"put":      routerModule.Put,
		"delete":   routerModule.Delete,
		"response": routerModule.Response,
	})

	// Handler returns nothing (undefined)
	h.MustRun(t, `
		router.get("/undefined", function(ctx) {
			// no return
		});
	`)

	ctx := &RequestContext{Method: "GET", Path: "/undefined"}
	resp, err := routerModule.Handle("GET", "/undefined", ctx)
	// Should handle gracefully
	if err != nil {
		t.Logf("Got error (acceptable): %v", err)
		return
	}
	t.Logf("Got response: %+v", resp)
}

func TestJS_Router_HandlerReturnsNull(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"put":      routerModule.Put,
		"delete":   routerModule.Delete,
		"response": routerModule.Response,
	})

	// Handler returns null
	h.MustRun(t, `
		router.get("/null", function(ctx) {
			return null;
		});
	`)

	ctx := &RequestContext{Method: "GET", Path: "/null"}
	resp, err := routerModule.Handle("GET", "/null", ctx)
	if err != nil {
		t.Logf("Got error (acceptable): %v", err)
		return
	}
	t.Logf("Got response: %+v", resp)
}

func TestJS_Router_InfiniteLoop(t *testing.T) {
	// Test that infinite loops don't hang forever (with timeout)
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"response": routerModule.Response,
	})

	// Register handler with potential infinite loop (limited by counter)
	h.MustRun(t, `
		router.get("/loop", function(ctx) {
			var count = 0;
			while (count < 1000000) { count++; }
			return router.response(200, {count: count});
		});
	`)

	// Should complete (with high count)
	ctx := &RequestContext{Method: "GET", Path: "/loop"}
	done := make(chan bool, 1)

	go func() {
		resp, _ := routerModule.Handle("GET", "/loop", ctx)
		if resp != nil {
			done <- true
		}
	}()

	select {
	case <-done:
		// OK - completed
	case <-time.After(5 * time.Second):
		t.Error("Handler took too long - possible infinite loop issue")
	}
}

func TestJS_Delayed_PanicRecovery(t *testing.T) {
	h := NewTestHelper(t)

	delayedModule := NewDelayedModule(5)
	h.VM.Set("delayed", map[string]interface{}{
		"run": delayedModule.Run,
	})

	var normalExecuted atomic.Bool

	// Set up tracking
	h.VM.Set("setExecuted", func() {
		normalExecuted.Store(true)
	})

	// Run task that panics
	h.MustRun(t, `
		delayed.run(function() {
			throw new Error("Crash in delayed task");
		});
	`)

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Run normal task - should still work
	h.MustRun(t, `
		delayed.run(function() {
			setExecuted();
		});
	`)

	time.Sleep(100 * time.Millisecond)

	if !normalExecuted.Load() {
		t.Error("Normal task should execute after panic in another task")
	}
}

func TestJS_Schedule_InvalidCron(t *testing.T) {
	h := NewTestHelper(t)

	logger := newTestLogger()
	scheduleModule := NewScheduleModule(logger)
	h.VM.Set("schedule", map[string]interface{}{
		"cron":   scheduleModule.Cron,
		"daily":  scheduleModule.Daily,
		"hourly": scheduleModule.Hourly,
	})

	// Invalid cron expression should not crash
	_, err := h.Run(`
		schedule.cron("invalid cron expression", function() {
			console.log("should not run");
		});
		"ok";
	`)

	if err != nil {
		t.Errorf("Should not error, got: %v", err)
	}

	// No jobs should be added
	if len(scheduleModule.jobs) != 0 {
		t.Errorf("Expected 0 jobs for invalid cron, got %d", len(scheduleModule.jobs))
	}
}

// ============== TYPE COERCION TESTS ==============

func TestJS_TypeCoercion(t *testing.T) {
	h := NewTestHelper(t)

	tests := []struct {
		name     string
		code     string
		validate func(result goja.Value) error
	}{
		{
			"crypto md5 of number becomes string",
			`crypto.md5(123)`,
			func(r goja.Value) error {
				if len(r.String()) != 32 {
					return fmt.Errorf("expected 32 char hash, got %d", len(r.String()))
				}
				return nil
			},
		},
		{
			"utils randomInt returns integer",
			`utils.randomInt(0, 10)`,
			func(r goja.Value) error {
				val := r.ToInteger()
				if val < 0 || val >= 10 {
					return fmt.Errorf("expected 0-9, got %d", val)
				}
				return nil
			},
		},
		{
			"utils timestamp returns number",
			`utils.timestamp()`,
			func(r goja.Value) error {
				val := r.ToInteger()
				// Should be after year 2020
				if val < 1577836800000 {
					return fmt.Errorf("timestamp too small: %d", val)
				}
				return nil
			},
		},
		{
			"env.get returns correct type",
			`typeof env.get("TEST_VAR")`,
			func(r goja.Value) error {
				if r.String() != "string" {
					return fmt.Errorf("expected string type, got %s", r.String())
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if err := tt.validate(result); err != nil {
				t.Error(err)
			}
		})
	}
}

// ============== CONCURRENT ACCESS TESTS ==============

func TestJS_ConcurrentRouterAccess(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"response": routerModule.Response,
	})

	h.MustRun(t, `
		router.get("/concurrent", function(ctx) {
			return router.response(200, {thread: ctx.query.id || "unknown"});
		});
	`)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Simulate concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := &RequestContext{
				Method: "GET",
				Path:   "/concurrent",
				Query:  map[string]string{"id": fmt.Sprintf("%d", id)},
			}
			_, err := routerModule.Handle("GET", "/concurrent", ctx)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent request failed: %v", err)
	}
}

func TestJS_ConcurrentDelayedTasks(t *testing.T) {
	delayedModule := NewDelayedModule(10)

	var counter atomic.Int32
	var wg sync.WaitGroup

	vm := goja.New()
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		counter.Add(1)
		return goja.Undefined()
	})

	callable, ok := goja.AssertFunction(handler)
	if !ok {
		t.Fatal("Failed to create callable")
	}

	// Run many tasks concurrently
	taskCount := 50
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			delayedModule.Run(callable)
		}()
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	if counter.Load() != int32(taskCount) {
		t.Errorf("Expected %d completed tasks, got %d", taskCount, counter.Load())
	}
}

// ============== EDGE CASE TESTS ==============

func TestJS_EdgeCases_EmptyStrings(t *testing.T) {
	h := NewTestHelper(t)

	tests := []struct {
		name string
		code string
	}{
		{"crypto md5 empty", `crypto.md5("")`},
		{"crypto sha256 empty", `crypto.sha256("")`},
		{"encoding base64 empty", `encoding.base64Encode("")`},
		{"encoding base64 decode empty", `encoding.base64Decode("")`},
		{"utils slugify empty", `utils.slugify("")`},
		{"utils truncate empty", `utils.truncate("", 10)`},
		{"utils capitalize empty", `utils.capitalize("")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			if err != nil {
				t.Errorf("Should handle empty string: %v", err)
				return
			}
			t.Logf("Result: %v", result.Export())
		})
	}
}

func TestJS_EdgeCases_LargeInputs(t *testing.T) {
	h := NewTestHelper(t)

	// Create large string in JS
	code := `
		var large = "";
		for (var i = 0; i < 10000; i++) {
			large += "x";
		}
		crypto.md5(large);
	`

	result := h.MustRun(t, code)
	if len(result.String()) != 32 {
		t.Errorf("Expected 32 char hash, got %d", len(result.String()))
	}
}

func TestJS_EdgeCases_UnicodeHandling(t *testing.T) {
	h := NewTestHelper(t)

	tests := []struct {
		name string
		code string
	}{
		{"crypto md5 unicode", `crypto.md5("привет мир")`},
		{"crypto sha256 emoji", `crypto.sha256("Hello 🌍")`},
		{"encoding base64 unicode", `encoding.base64Encode("日本語")`},
		{"utils slugify unicode", `utils.slugify("Привет World")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			if err != nil {
				t.Errorf("Should handle unicode: %v", err)
				return
			}
			t.Logf("Result: %v", result.Export())
		})
	}
}

func TestJS_EdgeCases_SpecialCharacters(t *testing.T) {
	h := NewTestHelper(t)

	// Test that special characters don't break anything
	tests := []struct {
		name string
		code string
	}{
		{"null bytes", `crypto.md5("hello\x00world")`},
		{"newlines", `crypto.md5("line1\nline2\rline3")`},
		{"tabs", `crypto.md5("col1\tcol2\tcol3")`},
		{"quotes", `crypto.md5("\"quoted\" and 'single'")`},
		{"backslashes", `crypto.md5("path\\to\\file")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			if err != nil {
				t.Errorf("Should handle special chars: %v", err)
				return
			}
			if len(result.String()) != 32 {
				t.Errorf("Expected 32 char hash, got %d", len(result.String()))
			}
		})
	}
}

// ============== REAL WORLD SCENARIO TESTS ==============

func TestJS_Scenario_APIAuthentication(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"response": routerModule.Response,
	})

	// Simulate API with authentication
	h.MustRun(t, `
		var API_KEY = env.get("TEST_VAR");

		router.get("/api/protected", function(ctx) {
			var authHeader = ctx.headers["Authorization"];
			if (!authHeader || authHeader !== "Bearer " + API_KEY) {
				return router.response(401, {error: "Unauthorized"});
			}
			return router.response(200, {data: "secret"});
		});
	`)

	// Test without auth
	ctx := &RequestContext{Method: "GET", Path: "/api/protected", Headers: map[string]string{}}
	resp, err := routerModule.Handle("GET", "/api/protected", ctx)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.Status != 401 {
		t.Errorf("Expected 401, got %d", resp.Status)
	}

	// Test with correct auth
	ctx = &RequestContext{
		Method:  "GET",
		Path:    "/api/protected",
		Headers: map[string]string{"Authorization": "Bearer test_value"},
	}
	resp, err = routerModule.Handle("GET", "/api/protected", ctx)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("Expected 200, got %d", resp.Status)
	}
}

func TestJS_Scenario_DataTransformation(t *testing.T) {
	h := NewTestHelper(t)

	// Complex data transformation pipeline
	code := `
		// Input data
		var rawData = [
			{name: "John Doe", email: "john@test.com", age: 25},
			{name: "Jane Smith", email: "jane@test.com", age: 30},
			{name: "Bob Johnson", email: "bob@test.com", age: 35}
		];

		// Transform
		var result = [];
		for (var i = 0; i < rawData.length; i++) {
			var item = rawData[i];
			result.push({
				id: crypto.md5(item.email).substring(0, 8),
				slug: utils.slugify(item.name),
				emailHash: crypto.sha256(item.email),
				ageGroup: item.age < 30 ? "young" : "adult",
				createdAt: utils.timestamp()
			});
		}

		encoding.jsonStringify(result);
	`

	result := h.MustRun(t, code)
	var parsed []interface{}
	if err := json.Unmarshal([]byte(result.String()), &parsed); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(parsed) != 3 {
		t.Errorf("Expected 3 items, got %d", len(parsed))
	}

	first := parsed[0].(map[string]interface{})
	if first["slug"] != "john-doe" {
		t.Errorf("Expected slug 'john-doe', got %v", first["slug"])
	}
}

func TestJS_Scenario_ErrorHandlingInService(t *testing.T) {
	h := NewTestHelper(t)

	routerModule := NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"response": routerModule.Response,
	})

	// Service with proper error handling
	h.MustRun(t, `
		router.post("/api/process", function(ctx) {
			try {
				if (!ctx.body) {
					return router.response(400, {error: "Body required"});
				}
				if (!ctx.body.data) {
					return router.response(400, {error: "data field required"});
				}

				var data = ctx.body.data;
				if (typeof data !== "string") {
					return router.response(400, {error: "data must be string"});
				}

				var result = {
					original: data,
					hash: crypto.md5(data),
					length: data.length,
					processed: true
				};

				return router.response(200, result);
			} catch (e) {
				return router.response(500, {error: "Internal error: " + e.message});
			}
		});
	`)

	// Test various error cases
	testCases := []struct {
		name           string
		body           interface{}
		expectedStatus int
	}{
		{"no body", nil, 400},
		{"empty body", map[string]interface{}{}, 400},
		{"wrong type", map[string]interface{}{"data": 123}, 400},
		{"valid", map[string]interface{}{"data": "test"}, 200},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &RequestContext{Method: "POST", Path: "/api/process", Body: tc.body}
			resp, err := routerModule.Handle("POST", "/api/process", ctx)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			if resp.Status != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.Status)
			}
		})
	}
}
