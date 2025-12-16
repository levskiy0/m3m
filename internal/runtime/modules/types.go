package modules

import (
	"github.com/levskiy0/m3m/pkg/schema"
)

// GetAllSchemas returns all module schemas
func GetAllSchemas() []schema.ModuleSchema {
	return []schema.ModuleSchema{
		GetLoggerSchema(),
		GetRouterSchema(),
		GetScheduleSchema(),
		GetEnvSchema(),
		GetStorageSchema(),
		GetDatabaseSchema(),
		GetGoalsSchema(),
		GetHTTPSchema(),
		GetCryptoSchema(),
		GetEncodingSchema(),
		GetUtilsSchema(),
		GetValidatorSchema(),
		GetDelayedSchema(),
		GetServiceSchema(),
		GetImageSchema(),
		GetDrawSchema(),
		GetMailSchema(),
		GetHookSchema(),
		GetUISchema(),
		GetRequireSchema(),
		GetExportsSchema(),
	}
}

// GetRequireSchema returns the $require module schema
func GetRequireSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$require",
		Description: "Load and use code from other files in your service. Files are executed once and cached.",
		Methods: []schema.MethodSchema{
			{
				Name:        "$require",
				Description: "Import exports from another file. The file is executed once and its exports are cached for subsequent calls.",
				Params: []schema.ParamSchema{
					{Name: "name", Type: "string", Description: "File name without extension (e.g., 'utils', 'helpers')"},
				},
				Returns: &schema.ParamSchema{Type: "object", Description: "Object containing exported values from the module"},
			},
		},
	}
}

// GetBaseTypeDefinitions returns TypeScript definitions for Monaco IntelliSense
func GetBaseTypeDefinitions() string {
	return GenerateAllTypeScript(GetAllSchemas())
}
