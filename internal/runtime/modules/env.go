package modules

import (
	"fmt"
)

type EnvModule struct {
	vars map[string]interface{}
}

func NewEnvModule(vars map[string]interface{}) *EnvModule {
	if vars == nil {
		vars = make(map[string]interface{})
	}
	return &EnvModule{vars: vars}
}

// Get returns the value for the given key, or nil if not found
func (e *EnvModule) Get(key string) interface{} {
	return e.vars[key]
}

// Has returns true if the key exists in the environment
func (e *EnvModule) Has(key string) bool {
	_, ok := e.vars[key]
	return ok
}

// Keys returns all environment variable keys
func (e *EnvModule) Keys() []string {
	keys := make([]string, 0, len(e.vars))
	for k := range e.vars {
		keys = append(keys, k)
	}
	return keys
}

// GetString returns the value as a string, or defaultValue if not found or not a string
func (e *EnvModule) GetString(key string, defaultValue string) string {
	val, ok := e.vars[key]
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
	val, ok := e.vars[key]
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
	default:
		return defaultValue
	}
}

// GetFloat returns the value as a float64, or defaultValue if not found or not convertible
func (e *EnvModule) GetFloat(key string, defaultValue float64) float64 {
	val, ok := e.vars[key]
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
	default:
		return defaultValue
	}
}

// GetBool returns the value as a bool, or defaultValue if not found or not a bool
func (e *EnvModule) GetBool(key string, defaultValue bool) bool {
	val, ok := e.vars[key]
	if !ok {
		return defaultValue
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

// GetAll returns a copy of all environment variables
func (e *EnvModule) GetAll() map[string]interface{} {
	result := make(map[string]interface{}, len(e.vars))
	for k, v := range e.vars {
		result[k] = v
	}
	return result
}

// GetSchema implements JSSchemaProvider
func (e *EnvModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "env",
		Description: "Access to environment variables configured for the project",
		Methods: []JSMethodSchema{
			{
				Name:        "get",
				Description: "Get a value by key, returns undefined if not found",
				Params:      []JSParamSchema{{Name: "key", Type: "string", Description: "Environment variable key"}},
				Returns:     &JSParamSchema{Type: "any"},
			},
			{
				Name:        "has",
				Description: "Check if a key exists in the environment",
				Params:      []JSParamSchema{{Name: "key", Type: "string", Description: "Environment variable key"}},
				Returns:     &JSParamSchema{Type: "boolean"},
			},
			{
				Name:        "keys",
				Description: "Get all environment variable keys",
				Returns:     &JSParamSchema{Type: "string[]"},
			},
			{
				Name:        "getString",
				Description: "Get a string value with default fallback",
				Params: []JSParamSchema{
					{Name: "key", Type: "string", Description: "Environment variable key"},
					{Name: "defaultValue", Type: "string", Description: "Default value if not found"},
				},
				Returns: &JSParamSchema{Type: "string"},
			},
			{
				Name:        "getInt",
				Description: "Get an integer value with default fallback",
				Params: []JSParamSchema{
					{Name: "key", Type: "string", Description: "Environment variable key"},
					{Name: "defaultValue", Type: "number", Description: "Default value if not found"},
				},
				Returns: &JSParamSchema{Type: "number"},
			},
			{
				Name:        "getFloat",
				Description: "Get a float value with default fallback",
				Params: []JSParamSchema{
					{Name: "key", Type: "string", Description: "Environment variable key"},
					{Name: "defaultValue", Type: "number", Description: "Default value if not found"},
				},
				Returns: &JSParamSchema{Type: "number"},
			},
			{
				Name:        "getBool",
				Description: "Get a boolean value with default fallback",
				Params: []JSParamSchema{
					{Name: "key", Type: "string", Description: "Environment variable key"},
					{Name: "defaultValue", Type: "boolean", Description: "Default value if not found"},
				},
				Returns: &JSParamSchema{Type: "boolean"},
			},
			{
				Name:        "getAll",
				Description: "Get all environment variables as a map",
				Returns:     &JSParamSchema{Type: "{ [key: string]: any }"},
			},
		},
	}
}

// GetEnvSchema returns the env schema (static version)
func GetEnvSchema() JSModuleSchema {
	return (&EnvModule{}).GetSchema()
}
