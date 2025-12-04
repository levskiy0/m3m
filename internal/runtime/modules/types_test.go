package modules

import (
	"strings"
	"testing"
)

func TestGetAllSchemas(t *testing.T) {
	schemas := GetAllSchemas()

	// Verify we have all expected modules
	expectedModules := []string{
		"logger", "console", "router", "schedule", "env", "storage", "database",
		"goals", "http", "crypto", "encoding", "utils", "delayed",
		"service", "image", "draw", "smtp",
	}

	if len(schemas) != len(expectedModules) {
		t.Errorf("Expected %d schemas, got %d", len(expectedModules), len(schemas))
	}

	moduleNames := make(map[string]bool)
	for _, s := range schemas {
		moduleNames[s.Name] = true
	}

	for _, expected := range expectedModules {
		if !moduleNames[expected] {
			t.Errorf("Missing module schema: %s", expected)
		}
	}
}

func TestGetBaseTypeDefinitions(t *testing.T) {
	ts := GetBaseTypeDefinitions()

	// Check header
	if !strings.Contains(ts, "M3M Runtime API Type Definitions") {
		t.Error("Missing header comment")
	}

	// Check that all modules are declared
	expectedDeclarations := []string{
		"declare const logger:",
		"declare const router:",
		"declare const schedule:",
		"declare const env:",
		"declare const storage:",
		"declare const database:",
		"declare const goals:",
		"declare const http:",
		"declare const crypto:",
		"declare const encoding:",
		"declare const utils:",
		"declare const delayed:",
		"declare const service:",
		"declare const image:",
		"declare const draw:",
		"declare const smtp:",
		"declare const console:",
	}

	for _, decl := range expectedDeclarations {
		if !strings.Contains(ts, decl) {
			t.Errorf("Missing declaration: %s", decl)
		}
	}
}

func TestGeneratedTypeScript_ContainsTypes(t *testing.T) {
	ts := GetBaseTypeDefinitions()

	// Check for custom types
	expectedTypes := []string{
		"interface RequestContext",
		"interface ResponseData",
		"interface HTTPResponse",
		"interface HTTPOptions",
		"interface GoalInfo",
		"interface GoalStatInfo",
		"interface Collection",
		"interface ImageInfo",
		"interface Canvas",
		"interface EmailOptions",
		"interface SMTPResult",
	}

	for _, typ := range expectedTypes {
		if !strings.Contains(ts, typ) {
			t.Errorf("Missing type: %s", typ)
		}
	}
}

func TestGeneratedTypeScript_ContainsMethods(t *testing.T) {
	ts := GetBaseTypeDefinitions()

	// Check for key methods
	expectedMethods := []string{
		"debug(",
		"info(",
		"get(",
		"post(",
		"daily(",
		"cron(",
		"increment(",
		"collection(",
		"md5(",
		"sha256(",
		"base64Encode(",
		"uuid(",
		"boot(",
		"createCanvas(",
		"send(",
	}

	for _, method := range expectedMethods {
		if !strings.Contains(ts, method) {
			t.Errorf("Missing method: %s", method)
		}
	}
}

func TestSchemaProvider_AllModulesImplement(t *testing.T) {
	// Test that all static schema getters work
	schemas := []func() JSModuleSchema{
		GetLoggerSchema,
		GetConsoleSchema,
		GetRouterSchema,
		GetScheduleSchema,
		GetEnvSchema,
		GetStorageSchema,
		GetDatabaseSchema,
		GetGoalsSchema,
		GetHTTPSchema,
		GetCryptoSchema,
		GetEncodingSchema,
		GetUtilsSchema,
		GetDelayedSchema,
		GetServiceSchema,
		GetImageSchema,
		GetDrawSchema,
		GetSMTPSchema,
	}

	for i, getter := range schemas {
		schema := getter()
		if schema.Name == "" {
			t.Errorf("Schema %d has empty name", i)
		}
		if schema.Description == "" {
			t.Errorf("Schema %d (%s) has empty description", i, schema.Name)
		}
		if len(schema.Methods) == 0 {
			t.Errorf("Schema %d (%s) has no methods", i, schema.Name)
		}
	}
}
