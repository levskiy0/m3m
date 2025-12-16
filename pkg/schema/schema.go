// Package schema provides schema definitions for M3M runtime modules and plugins.
// This package is designed to be imported by both internal modules and external plugins.
package schema

import (
	"fmt"
	"strings"
)

// ParamSchema describes a method parameter or field
type ParamSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "number", "boolean", "object", "array", "any", "void", custom type name
	Description string `json:"description"`
	Optional    bool   `json:"optional,omitempty"`
	Variadic    bool   `json:"variadic,omitempty"` // true for rest parameters (...args)
}

// MethodSchema describes a module method
type MethodSchema struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Params      []ParamSchema `json:"params,omitempty"`
	Returns     *ParamSchema  `json:"returns,omitempty"` // nil means void
}

// TypeSchema describes a custom type/interface
type TypeSchema struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Fields      []ParamSchema `json:"fields"`
}

// NestedModuleSchema describes a nested module/namespace
type NestedModuleSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Methods     []MethodSchema `json:"methods"`
}

// ModuleSchema describes a complete module
type ModuleSchema struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Methods     []MethodSchema       `json:"methods"`
	Types       []TypeSchema         `json:"types,omitempty"`      // Custom types used by this module
	Nested      []NestedModuleSchema `json:"nested,omitempty"`     // Nested namespaces/objects
	RawTypes    string               `json:"rawTypes,omitempty"`   // Raw TypeScript for complex types
	IsFunction  bool                 `json:"isFunction,omitempty"` // If true, generate as standalone function instead of object with methods
}

// SchemaProvider interface for modules that provide schema
type SchemaProvider interface {
	GetSchema() ModuleSchema
}

// GenerateTypeScript generates TypeScript declaration from schema
func (s *ModuleSchema) GenerateTypeScript() string {
	var sb strings.Builder

	// Output raw types first (for complex type definitions)
	if s.RawTypes != "" {
		sb.WriteString(s.RawTypes)
		sb.WriteString("\n\n")
	}

	// Generate custom types first
	for _, t := range s.Types {
		if t.Description != "" {
			sb.WriteString(fmt.Sprintf("/** %s */\n", t.Description))
		}
		sb.WriteString(fmt.Sprintf("interface %s {\n", t.Name))
		for _, f := range t.Fields {
			optional := ""
			if f.Optional {
				optional = "?"
			}
			if f.Description != "" {
				sb.WriteString(fmt.Sprintf("    /** %s */\n", f.Description))
			}
			sb.WriteString(fmt.Sprintf("    %s%s: %s;\n", f.Name, optional, MapTypeToTS(f.Type)))
		}
		sb.WriteString("}\n\n")
	}

	// Generate module declaration
	if s.Description != "" {
		sb.WriteString(fmt.Sprintf("/** %s */\n", s.Description))
	}
	sb.WriteString(fmt.Sprintf("declare const %s: {\n", s.Name))

	for _, m := range s.Methods {
		// Add JSDoc comment
		if m.Description != "" {
			sb.WriteString(fmt.Sprintf("    /** %s */\n", m.Description))
		}

		// Build params
		params := make([]string, 0, len(m.Params))
		for _, p := range m.Params {
			optional := ""
			if p.Optional {
				optional = "?"
			}
			if p.Variadic {
				// Rest parameter: ...args: any[]
				params = append(params, fmt.Sprintf("...%s: %s[]", p.Name, MapTypeToTS(p.Type)))
			} else {
				params = append(params, fmt.Sprintf("%s%s: %s", p.Name, optional, MapTypeToTS(p.Type)))
			}
		}

		// Build return type
		returnType := "void"
		if m.Returns != nil {
			returnType = MapTypeToTS(m.Returns.Type)
		}

		sb.WriteString(fmt.Sprintf("    %s(%s): %s;\n", m.Name, strings.Join(params, ", "), returnType))
		sb.WriteString("\n")
	}

	// Generate nested modules
	for _, nested := range s.Nested {
		if nested.Description != "" {
			sb.WriteString(fmt.Sprintf("    /** %s */\n", nested.Description))
		}
		sb.WriteString(fmt.Sprintf("    %s: {\n", nested.Name))

		for _, m := range nested.Methods {
			if m.Description != "" {
				sb.WriteString(fmt.Sprintf("        /** %s */\n", m.Description))
			}

			params := make([]string, 0, len(m.Params))
			for _, p := range m.Params {
				optional := ""
				if p.Optional {
					optional = "?"
				}
				if p.Variadic {
					params = append(params, fmt.Sprintf("...%s: %s[]", p.Name, MapTypeToTS(p.Type)))
				} else {
					params = append(params, fmt.Sprintf("%s%s: %s", p.Name, optional, MapTypeToTS(p.Type)))
				}
			}

			returnType := "void"
			if m.Returns != nil {
				returnType = MapTypeToTS(m.Returns.Type)
			}

			sb.WriteString(fmt.Sprintf("        %s(%s): %s;\n", m.Name, strings.Join(params, ", "), returnType))
		}

		sb.WriteString("    };\n\n")
	}

	sb.WriteString("};\n")

	return sb.String()
}

// MapTypeToTS converts schema type to TypeScript type
func MapTypeToTS(t string) string {
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
func GenerateAllTypeScript(schemas []ModuleSchema) string {
	var sb strings.Builder

	sb.WriteString("// M3M Runtime API Type Definitions\n")
	sb.WriteString("// Auto-generated from module schemas\n\n")

	for _, s := range schemas {
		sb.WriteString(s.GenerateTypeScript())
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateMarkdown generates Markdown documentation from schema
func (s *ModuleSchema) GenerateMarkdown() string {
	var sb strings.Builder

	// Module header
	sb.WriteString(fmt.Sprintf("## %s\n\n", s.Name))
	if s.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", s.Description))
	}

	// Types section
	if len(s.Types) > 0 {
		sb.WriteString("### Types\n\n")
		for _, t := range s.Types {
			sb.WriteString(fmt.Sprintf("#### %s\n\n", t.Name))
			if t.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", t.Description))
			}
			sb.WriteString("| Field | Type | Description |\n")
			sb.WriteString("|-------|------|-------------|\n")
			for _, f := range t.Fields {
				optional := ""
				if f.Optional {
					optional = " *(optional)*"
				}
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s%s |\n", f.Name, f.Type, f.Description, optional))
			}
			sb.WriteString("\n")
		}
	}

	// Methods section
	if len(s.Methods) > 0 {
		sb.WriteString("### Methods\n\n")
		for _, m := range s.Methods {
			// Build params signature
			params := make([]string, 0, len(m.Params))
			for _, p := range m.Params {
				optional := ""
				if p.Optional {
					optional = "?"
				}
				params = append(params, fmt.Sprintf("%s%s: %s", p.Name, optional, p.Type))
			}

			returnType := "void"
			if m.Returns != nil {
				returnType = m.Returns.Type
			}

			sb.WriteString(fmt.Sprintf("#### `%s(%s): %s`\n\n", m.Name, strings.Join(params, ", "), returnType))
			if m.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", m.Description))
			}

			// Parameters table
			if len(m.Params) > 0 {
				sb.WriteString("**Parameters:**\n\n")
				sb.WriteString("| Name | Type | Description |\n")
				sb.WriteString("|------|------|-------------|\n")
				for _, p := range m.Params {
					optional := ""
					if p.Optional {
						optional = " *(optional)*"
					}
					sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s%s |\n", p.Name, p.Type, p.Description, optional))
				}
				sb.WriteString("\n")
			}
		}
	}

	// Nested modules
	for _, nested := range s.Nested {
		sb.WriteString(fmt.Sprintf("### %s.%s\n\n", s.Name, nested.Name))
		if nested.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", nested.Description))
		}

		for _, m := range nested.Methods {
			params := make([]string, 0, len(m.Params))
			for _, p := range m.Params {
				optional := ""
				if p.Optional {
					optional = "?"
				}
				params = append(params, fmt.Sprintf("%s%s: %s", p.Name, optional, p.Type))
			}

			returnType := "void"
			if m.Returns != nil {
				returnType = m.Returns.Type
			}

			sb.WriteString(fmt.Sprintf("#### `%s(%s): %s`\n\n", m.Name, strings.Join(params, ", "), returnType))
			if m.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", m.Description))
			}
		}
	}

	return sb.String()
}

// GenerateAllMarkdown generates Markdown documentation from multiple schemas
func GenerateAllMarkdown(schemas []ModuleSchema) string {
	var sb strings.Builder

	sb.WriteString("# M3M JavaScript Runtime API\n\n")
	sb.WriteString("Auto-generated documentation for M3M runtime modules.\n\n")
	sb.WriteString("All modules are available globally with `$` prefix (e.g., `$router`, `$database`).\n\n")

	// Table of contents
	sb.WriteString("## Table of Contents\n\n")
	for _, s := range schemas {
		anchor := strings.ToLower(strings.ReplaceAll(s.Name, "$", ""))
		sb.WriteString(fmt.Sprintf("- [%s](#%s) - %s\n", s.Name, anchor, truncateString(s.Description, 60)))
	}
	sb.WriteString("\n---\n\n")

	for _, s := range schemas {
		sb.WriteString(s.GenerateMarkdown())
		sb.WriteString("\n---\n\n")
	}

	return sb.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
