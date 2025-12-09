package tests

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/levskiy0/m3m/internal/runtime/modules"

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

	vm.Set("$logger", map[string]interface{}{
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
	// Stateless modules - use self-registration
	modules.NewCryptoModule().Register(vm)
	modules.NewEncodingModule().Register(vm)
	modules.NewUtilsModule().Register(vm)
	modules.NewValidatorModule().Register(vm)
	modules.NewDelayedModule(5).Register(vm)
	modules.NewHTTPModule(30 * time.Second).Register(vm)

	// Env (mock) - use self-registration
	modules.NewEnvModule(map[string]interface{}{
		"TEST_VAR":   "test_value",
		"API_KEY":    "secret123",
		"NUMBER_VAR": "42",
		"BOOL_VAR":   "true",
		"FLOAT_VAR":  "3.14",
		"EMPTY_VAR":  "",
		"JSON_VAR":   `{"key": "value"}`,
		"DB_HOST":    "localhost",
		"DB_PORT":    "5432",
	}).Register(vm)
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
	routerModule.Register(h.VM)
	return routerModule
}
