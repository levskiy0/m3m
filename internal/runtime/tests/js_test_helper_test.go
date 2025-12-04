package tests

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"m3m/internal/runtime/modules"

	"github.com/dop251/goja"
)

// JSTestHelper creates a Goja VM with modules registered for testing
type JSTestHelper struct {
	VM   *goja.Runtime
	logs []string
	mu   sync.Mutex
}

// NewJSTestHelper creates a test VM with all modules registered
func NewJSTestHelper(t *testing.T) *JSTestHelper {
	t.Helper()
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	h := &JSTestHelper{
		VM:   vm,
		logs: []string{},
	}

	h.registerMockLogger(vm)
	h.registerModules(vm)

	return h
}

func (h *JSTestHelper) registerMockLogger(vm *goja.Runtime) {
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

func (h *JSTestHelper) registerModules(vm *goja.Runtime) {
	// Crypto
	cryptoModule := modules.NewCryptoModule()
	vm.Set("crypto", map[string]interface{}{
		"md5":         cryptoModule.MD5,
		"sha256":      cryptoModule.SHA256,
		"randomBytes": cryptoModule.RandomBytes,
	})

	// Encoding
	encodingModule := modules.NewEncodingModule()
	vm.Set("encoding", map[string]interface{}{
		"base64Encode":  encodingModule.Base64Encode,
		"base64Decode":  encodingModule.Base64Decode,
		"jsonParse":     encodingModule.JSONParse,
		"jsonStringify": encodingModule.JSONStringify,
		"urlEncode":     encodingModule.URLEncode,
		"urlDecode":     encodingModule.URLDecode,
	})

	// Utils
	utilsModule := modules.NewUtilsModule()
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

	// Env (mock)
	envModule := modules.NewEnvModule(map[string]interface{}{
		"TEST_VAR":   "test_value",
		"API_KEY":    "secret123",
		"NUMBER_VAR": "42",
		"BOOL_VAR":   "true",
		"FLOAT_VAR":  "3.14",
		"EMPTY_VAR":  "",
		"JSON_VAR":   `{"key": "value"}`,
		"DB_HOST":    "localhost",
		"DB_PORT":    "5432",
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
	delayedModule := modules.NewDelayedModule(5)
	vm.Set("delayed", map[string]interface{}{
		"run": delayedModule.Run,
	})

	// HTTP
	httpModule := modules.NewHTTPModule(30 * time.Second)
	vm.Set("http", map[string]interface{}{
		"get":    httpModule.Get,
		"post":   httpModule.Post,
		"put":    httpModule.Put,
		"delete": httpModule.Delete,
	})
}

// Run executes JS code and returns the result
func (h *JSTestHelper) Run(code string) (goja.Value, error) {
	return h.VM.RunString(code)
}

// MustRun executes JS code and fails test on error
func (h *JSTestHelper) MustRun(t *testing.T, code string) goja.Value {
	t.Helper()
	result, err := h.Run(code)
	if err != nil {
		t.Fatalf("JS execution failed: %v\nCode: %s", err, code)
	}
	return result
}

// GetLogs returns captured log messages
func (h *JSTestHelper) GetLogs() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string{}, h.logs...)
}

// ClearLogs clears captured log messages
func (h *JSTestHelper) ClearLogs() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logs = []string{}
}

// SetupRouter creates and registers a router module
func (h *JSTestHelper) SetupRouter() *modules.RouterModule {
	routerModule := modules.NewRouterModule()
	routerModule.SetVM(h.VM)
	h.VM.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"put":      routerModule.Put,
		"delete":   routerModule.Delete,
		"response": routerModule.Response,
	})
	return routerModule
}
