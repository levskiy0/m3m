package modules

import (
	"github.com/levskiy0/m3m/pkg/schema"
)

// JSModule interface for self-registering modules
type JSModule interface {
	schema.SchemaProvider
	// Name returns the module name as it appears in JavaScript (e.g., "validator", "crypto")
	Name() string
	// Register registers the module's methods into the JavaScript VM
	Register(vm interface{})
}

// GenerateAllTypeScript generates TypeScript from multiple schemas
func GenerateAllTypeScript(schemas []schema.ModuleSchema) string {
	return schema.GenerateAllTypeScript(schemas)
}
