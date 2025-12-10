package modules

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/pkg/schema"
)

type EnvModule struct {
	getVars func() map[string]interface{}
}

// NewEnvModule creates a new env module with a getter function for lazy loading.
// This allows env vars to be updated without restarting the runtime.
func NewEnvModule(getVars func() map[string]interface{}) *EnvModule {
	if getVars == nil {
		getVars = func() map[string]interface{} { return make(map[string]interface{}) }
	}
	return &EnvModule{getVars: getVars}
}

// Name returns the module name for JavaScript
func (e *EnvModule) Name() string {
	return "$env"
}

// Register registers the module into the JavaScript VM
func (e *EnvModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(e.Name(), map[string]interface{}{
		"get":       e.Get,
		"has":       e.Has,
		"keys":      e.Keys,
		"getString": e.GetString,
		"getInt":    e.GetInt,
		"getFloat":  e.GetFloat,
		"getBool":   e.GetBool,
		"getAll":    e.GetAll,
	})
}

// Get returns the value for the given key, or nil if not found
func (e *EnvModule) Get(key string) interface{} {
	return e.getVars()[key]
}

// Has returns true if the key exists in the environment
func (e *EnvModule) Has(key string) bool {
	_, ok := e.getVars()[key]
	return ok
}

// Keys returns all environment variable keys
func (e *EnvModule) Keys() []string {
	vars := e.getVars()
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	return keys
}

// GetString returns the value as a string, or defaultValue if not found or not a string
func (e *EnvModule) GetString(key string, defaultValue string) string {
	val, ok := e.getVars()[key]
	if !ok {
		return defaultValue
	}
	if str, ok := val.(string); ok {
		return str
	}
	// Try to convert to string
	return fmt.Sprintf("%v", val)
}

// GetInt returns the value as an int, or defaultValue if not found or not convertible
func (e *EnvModule) GetInt(key string, defaultValue int) int {
	val, ok := e.getVars()[key]
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i
		}
		return defaultValue
	default:
		return defaultValue
	}
}

// GetFloat returns the value as a float64, or defaultValue if not found or not convertible
func (e *EnvModule) GetFloat(key string, defaultValue float64) float64 {
	val, ok := e.getVars()[key]
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f
		}
		return defaultValue
	default:
		return defaultValue
	}
}

// GetBool returns the value as a bool, or defaultValue if not found or not a bool
func (e *EnvModule) GetBool(key string, defaultValue bool) bool {
	val, ok := e.getVars()[key]
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		if s == "true" || s == "1" || s == "yes" || s == "on" {
			return true
		}
		if s == "false" || s == "0" || s == "no" || s == "off" {
			return false
		}
		return defaultValue
	default:
		return defaultValue
	}
}

// GetAll returns a copy of all environment variables
func (e *EnvModule) GetAll() map[string]interface{} {
	vars := e.getVars()
	result := make(map[string]interface{}, len(vars))
	for k, v := range vars {
		result[k] = v
	}
	return result
}

// GetSchema implements JSSchemaProvider
func (e *EnvModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$env",
		Description: "Access to environment variables configured for the project",
		Methods: []schema.MethodSchema{
			{
				Name:        "get",
				Description: "Get a value by key, returns undefined if not found",
				Params:      []schema.ParamSchema{{Name: "key", Type: "string", Description: "Environment variable key"}},
				Returns:     &schema.ParamSchema{Type: "any"},
			},
			{
				Name:        "has",
				Description: "Check if a key exists in the environment",
				Params:      []schema.ParamSchema{{Name: "key", Type: "string", Description: "Environment variable key"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "keys",
				Description: "Get all environment variable keys",
				Returns:     &schema.ParamSchema{Type: "string[]"},
			},
			{
				Name:        "getString",
				Description: "Get a string value with default fallback",
				Params: []schema.ParamSchema{
					{Name: "key", Type: "string", Description: "Environment variable key"},
					{Name: "defaultValue", Type: "string", Description: "Default value if not found"},
				},
				Returns: &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "getInt",
				Description: "Get an integer value with default fallback",
				Params: []schema.ParamSchema{
					{Name: "key", Type: "string", Description: "Environment variable key"},
					{Name: "defaultValue", Type: "number", Description: "Default value if not found"},
				},
				Returns: &schema.ParamSchema{Type: "number"},
			},
			{
				Name:        "getFloat",
				Description: "Get a float value with default fallback",
				Params: []schema.ParamSchema{
					{Name: "key", Type: "string", Description: "Environment variable key"},
					{Name: "defaultValue", Type: "number", Description: "Default value if not found"},
				},
				Returns: &schema.ParamSchema{Type: "number"},
			},
			{
				Name:        "getBool",
				Description: "Get a boolean value with default fallback",
				Params: []schema.ParamSchema{
					{Name: "key", Type: "string", Description: "Environment variable key"},
					{Name: "defaultValue", Type: "boolean", Description: "Default value if not found"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "getAll",
				Description: "Get all environment variables as a map",
				Returns:     &schema.ParamSchema{Type: "{ [key: string]: any }"},
			},
		},
	}
}

// GetEnvSchema returns the env schema (static version)
func GetEnvSchema() schema.ModuleSchema {
	return (&EnvModule{}).GetSchema()
}
