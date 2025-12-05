package modules

import (
	"m3m/pkg/schema"
)

// Type aliases for backward compatibility
type JSParamSchema = schema.ParamSchema
type JSMethodSchema = schema.MethodSchema
type JSTypeSchema = schema.TypeSchema
type JSNestedModuleSchema = schema.NestedModuleSchema
type JSModuleSchema = schema.ModuleSchema
type JSSchemaProvider = schema.SchemaProvider

// JSModule interface for self-registering modules
type JSModule interface {
	JSSchemaProvider
	// Name returns the module name as it appears in JavaScript (e.g., "validator", "crypto")
	Name() string
	// Register registers the module's methods into the JavaScript VM
	Register(vm interface{})
}

// GenerateAllTypeScript generates TypeScript from multiple schemas
func GenerateAllTypeScript(schemas []JSModuleSchema) string {
	return schema.GenerateAllTypeScript(schemas)
}
