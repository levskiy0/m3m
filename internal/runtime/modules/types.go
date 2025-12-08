package modules

import (
	"m3m/pkg/schema"
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
		GetSMTPSchema(),
	}
}

// GetBaseTypeDefinitions returns TypeScript definitions for Monaco IntelliSense
func GetBaseTypeDefinitions() string {
	return GenerateAllTypeScript(GetAllSchemas())
}
