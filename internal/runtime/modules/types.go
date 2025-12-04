package modules

// GetAllSchemas returns all module schemas
func GetAllSchemas() []JSModuleSchema {
	return []JSModuleSchema{
		GetLoggerSchema(),
		GetConsoleSchema(),
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
