package modules

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/pkg/schema"
)

// RequireModule implements $require and $exports for multi-file support
type RequireModule struct {
	vm      *goja.Runtime
	files   map[string]string     // name -> code
	cache   map[string]goja.Value // name -> exports (cached loaded modules)
	loading map[string]bool       // name -> true (cycle detection)
	mu      sync.Mutex
}

// NewRequireModule creates a new require module with the given files
func NewRequireModule(vm *goja.Runtime, files []domain.CodeFile) *RequireModule {
	fileMap := make(map[string]string)
	for _, f := range files {
		fileMap[f.Name] = f.Code
	}
	return &RequireModule{
		vm:      vm,
		files:   fileMap,
		cache:   make(map[string]goja.Value),
		loading: make(map[string]bool),
	}
}

// Name returns the module name
func (r *RequireModule) Name() string {
	return "$require"
}

// Register registers $require function in the VM
func (r *RequireModule) Register(vm interface{}) {
	v := vm.(*goja.Runtime)
	v.Set("$require", r.Require)
}

// Require loads and returns exports from a module
func (r *RequireModule) Require(name string) (goja.Value, error) {
	r.mu.Lock()

	// Check cache
	if exports, ok := r.cache[name]; ok {
		r.mu.Unlock()
		return exports, nil
	}

	// Check for circular dependency
	if r.loading[name] {
		r.mu.Unlock()
		return nil, fmt.Errorf("circular dependency detected: %s", name)
	}

	// Check if file exists
	code, ok := r.files[name]
	if !ok {
		r.mu.Unlock()
		return nil, fmt.Errorf("module not found: %s", name)
	}

	r.loading[name] = true
	r.mu.Unlock()

	// Execute module code with local $exports
	exports, err := r.executeModule(name, code)

	r.mu.Lock()
	delete(r.loading, name)
	if err == nil {
		r.cache[name] = exports
	}
	r.mu.Unlock()

	return exports, err
}

// executeModule executes module code and returns its exports
func (r *RequireModule) executeModule(name, code string) (goja.Value, error) {
	// Wrap code in a function with local $exports
	// The module can call $exports({ key: value }) to export values
	wrappedCode := fmt.Sprintf(`
(function() {
	var __moduleExports = {};
	var $exports = function(obj) {
		if (typeof obj === 'object' && obj !== null) {
			Object.keys(obj).forEach(function(key) {
				__moduleExports[key] = obj[key];
			});
		}
	};
	%s
	return __moduleExports;
})()
`, code)

	result, err := r.vm.RunString(wrappedCode)
	if err != nil {
		return nil, fmt.Errorf("error in module %s: %w", name, err)
	}

	return result, nil
}

// GetSchema implements schema provider for documentation
func (r *RequireModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$require",
		Description: "Load and use code from other files in your service",
		Methods: []schema.MethodSchema{
			{
				Name:        "$require",
				Description: "Import exports from another file. Files are executed once and cached. Example: const { helper } = $require('utils');",
				Params: []schema.ParamSchema{
					{Name: "name", Type: "string", Description: "File name without extension (e.g., 'utils', 'helpers')"},
				},
				Returns: &schema.ParamSchema{Type: "object", Description: "Object containing exported values from the module"},
			},
		},
	}
}

// GetExportsSchema returns schema for $exports documentation
func GetExportsSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$exports",
		Description: "Export values from the current file for use by other files via $require",
		Methods: []schema.MethodSchema{
			{
				Name:        "$exports",
				Description: "Export an object with named values. Can be called multiple times. Example: $exports({ helper: myFunction, VERSION: '1.0' });",
				Params: []schema.ParamSchema{
					{Name: "exports", Type: "object", Description: "Object with key-value pairs to export"},
				},
			},
		},
	}
}
