package modules

import (
	"fmt"
	"strings"
)

// JSParamSchema describes a method parameter
type JSParamSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "number", "boolean", "object", "array", "any", "void", custom type name
	Description string `json:"description"`
	Optional    bool   `json:"optional,omitempty"`
}

// JSMethodSchema describes a module method
type JSMethodSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Params      []JSParamSchema `json:"params,omitempty"`
	Returns     *JSParamSchema  `json:"returns,omitempty"` // nil means void
}

// JSTypeSchema describes a custom type/interface
type JSTypeSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Fields      []JSParamSchema `json:"fields"`
}

// JSNestedModuleSchema describes a nested module/namespace
type JSNestedModuleSchema struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Methods     []JSMethodSchema `json:"methods"`
}

// JSModuleSchema describes a complete module
type JSModuleSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Methods     []JSMethodSchema       `json:"methods"`
	Types       []JSTypeSchema         `json:"types,omitempty"`  // Custom types used by this module
	Nested      []JSNestedModuleSchema `json:"nested,omitempty"` // Nested namespaces/objects
}

// JSSchemaProvider interface for modules that provide schema
type JSSchemaProvider interface {
	GetSchema() JSModuleSchema
}

// JSModule interface for self-registering modules
type JSModule interface {
	JSSchemaProvider
	// Name returns the module name as it appears in JavaScript (e.g., "validator", "crypto")
	Name() string
	// Register registers the module's methods into the JavaScript VM
	Register(vm interface{})
}

// GenerateTypeScript generates TypeScript declaration from schema
func (s *JSModuleSchema) GenerateTypeScript() string {
	var sb strings.Builder

	// Generate custom types first
	for _, t := range s.Types {
		sb.WriteString(fmt.Sprintf("interface %s {\n", t.Name))
		for _, f := range t.Fields {
			optional := ""
			if f.Optional {
				optional = "?"
			}
			sb.WriteString(fmt.Sprintf("    %s%s: %s;\n", f.Name, optional, mapTypeToTS(f.Type)))
		}
		sb.WriteString("}\n\n")
	}

	// Generate module declaration
	sb.WriteString(fmt.Sprintf("// %s\n", s.Description))
	sb.WriteString(fmt.Sprintf("declare const %s: {\n", s.Name))

	for _, m := range s.Methods {
		// Add JSDoc comment
		sb.WriteString(fmt.Sprintf("    /** %s */\n", m.Description))

		// Build params
		params := make([]string, 0, len(m.Params))
		for _, p := range m.Params {
			optional := ""
			if p.Optional {
				optional = "?"
			}
			params = append(params, fmt.Sprintf("%s%s: %s", p.Name, optional, mapTypeToTS(p.Type)))
		}

		// Build return type
		returnType := "void"
		if m.Returns != nil {
			returnType = mapTypeToTS(m.Returns.Type)
		}

		sb.WriteString(fmt.Sprintf("    %s(%s): %s;\n", m.Name, strings.Join(params, ", "), returnType))
		sb.WriteString("\n")
	}

	// Generate nested modules
	for _, nested := range s.Nested {
		sb.WriteString(fmt.Sprintf("    /** %s */\n", nested.Description))
		sb.WriteString(fmt.Sprintf("    %s: {\n", nested.Name))

		for _, m := range nested.Methods {
			sb.WriteString(fmt.Sprintf("        /** %s */\n", m.Description))

			params := make([]string, 0, len(m.Params))
			for _, p := range m.Params {
				optional := ""
				if p.Optional {
					optional = "?"
				}
				params = append(params, fmt.Sprintf("%s%s: %s", p.Name, optional, mapTypeToTS(p.Type)))
			}

			returnType := "void"
			if m.Returns != nil {
				returnType = mapTypeToTS(m.Returns.Type)
			}

			sb.WriteString(fmt.Sprintf("        %s(%s): %s;\n", m.Name, strings.Join(params, ", "), returnType))
		}

		sb.WriteString("    };\n\n")
	}

	sb.WriteString("};\n")

	return sb.String()
}

// mapTypeToTS converts schema type to TypeScript type
func mapTypeToTS(t string) string {
	switch t {
	case "string":
		return "string"
	case "number", "int", "float", "int64", "float64":
		return "number"
	case "boolean", "bool":
		return "boolean"
	case "object":
		return "{ [key: string]: any }"
	case "array":
		return "any[]"
	case "any":
		return "any"
	case "void":
		return "void"
	case "string[]":
		return "string[]"
	case "number[]":
		return "number[]"
	default:
		// Custom type or complex type - return as-is
		return t
	}
}

// GenerateAllTypeScript generates TypeScript from multiple schemas
func GenerateAllTypeScript(schemas []JSModuleSchema) string {
	var sb strings.Builder

	sb.WriteString("// M3M Runtime API Type Definitions\n")
	sb.WriteString("// Auto-generated from module schemas\n\n")

	for _, schema := range schemas {
		sb.WriteString(schema.GenerateTypeScript())
		sb.WriteString("\n")
	}

	return sb.String()
}
