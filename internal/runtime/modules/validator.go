package modules

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-playground/validator/v10"
	"github.com/levskiy0/m3m/pkg/schema"
)

// ValidatorModule provides data validation utilities using go-playground/validator
type ValidatorModule struct {
	validate *validator.Validate
}

// NewValidatorModule creates a new validator module
func NewValidatorModule() *ValidatorModule {
	return &ValidatorModule{
		validate: validator.New(),
	}
}

// Name returns the module name for JavaScript
func (v *ValidatorModule) Name() string {
	return "$validator"
}

// Register registers the module into the JavaScript VM
func (v *ValidatorModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(v.Name(), map[string]interface{}{
		"struct":         v.Struct,
		"var":            v.Var,
		"isValid":        v.IsValid,
		"isEmail":        v.IsEmail,
		"isURL":          v.IsURL,
		"isUUID":         v.IsUUID,
		"isUUIDv4":       v.IsUUIDv4,
		"isNumeric":      v.IsNumeric,
		"isAlpha":        v.IsAlpha,
		"isAlphanumeric": v.IsAlphanumeric,
		"isJSON":         v.IsJSON,
		"isBase64":       v.IsBase64,
		"isIP":           v.IsIP,
		"isIPv4":         v.IsIPv4,
		"isIPv6":         v.IsIPv6,
		"isCIDR":         v.IsCIDR,
		"isMAC":          v.IsMAC,
		"isHexColor":     v.IsHexColor,
		"isRGBColor":     v.IsRGBColor,
		"isRGBAColor":    v.IsRGBAColor,
		"isLatitude":     v.IsLatitude,
		"isLongitude":    v.IsLongitude,
		"isCreditCard":   v.IsCreditCard,
		"isISBN":         v.IsISBN,
		"contains":       v.Contains,
		"startsWith":     v.StartsWith,
		"endsWith":       v.EndsWith,
		"minLength":      v.MinLength,
		"maxLength":      v.MaxLength,
		"length":         v.Length,
		"lengthBetween":  v.LengthBetween,
		"min":            v.Min,
		"max":            v.Max,
		"between":        v.Between,
		"matches":        v.Matches,
		"oneOf":          v.OneOf,
		"notEmpty":       v.NotEmpty,
		"required":       v.Required,
	})
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Struct validates a struct/object against validation tags
// Example: validator.struct({email: "test@example.com", age: 25}, {email: "required,email", age: "required,min=18"})
func (v *ValidatorModule) Struct(data map[string]interface{}, rules map[string]string) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []ValidationError{}}

	for field, rule := range rules {
		value, exists := data[field]
		if !exists {
			if strings.Contains(rule, "required") {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   field,
					Tag:     "required",
					Value:   "",
					Message: field + " is required",
				})
			}
			continue
		}

		// Validate single value
		err := v.validate.Var(value, rule)
		if err != nil {
			result.Valid = false
			for _, e := range err.(validator.ValidationErrors) {
				result.Errors = append(result.Errors, ValidationError{
					Field:   field,
					Tag:     e.Tag(),
					Value:   toString(value),
					Message: formatValidationMessage(field, e),
				})
			}
		}
	}

	return result
}

// Var validates a single value against rules
// Example: validator.var("test@example.com", "required,email")
func (v *ValidatorModule) Var(value interface{}, rules string) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []ValidationError{}}

	err := v.validate.Var(value, rules)
	if err != nil {
		result.Valid = false
		for _, e := range err.(validator.ValidationErrors) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "",
				Tag:     e.Tag(),
				Value:   toString(value),
				Message: formatValidationMessageSimple(e),
			})
		}
	}

	return result
}

// IsValid performs quick validation and returns boolean
func (v *ValidatorModule) IsValid(value interface{}, rules string) bool {
	return v.validate.Var(value, rules) == nil
}

// IsEmail checks if value is a valid email
func (v *ValidatorModule) IsEmail(value string) bool {
	return v.validate.Var(value, "email") == nil
}

// IsURL checks if value is a valid URL
func (v *ValidatorModule) IsURL(value string) bool {
	return v.validate.Var(value, "url") == nil
}

// IsUUID checks if value is a valid UUID (any version)
func (v *ValidatorModule) IsUUID(value string) bool {
	return v.validate.Var(value, "uuid") == nil
}

// IsUUIDv4 checks if value is a valid UUID v4
func (v *ValidatorModule) IsUUIDv4(value string) bool {
	return v.validate.Var(value, "uuid4") == nil
}

// IsNumeric checks if value contains only numeric characters
func (v *ValidatorModule) IsNumeric(value string) bool {
	return v.validate.Var(value, "numeric") == nil
}

// IsAlpha checks if value contains only alphabetic characters
func (v *ValidatorModule) IsAlpha(value string) bool {
	return v.validate.Var(value, "alpha") == nil
}

// IsAlphanumeric checks if value contains only alphanumeric characters
func (v *ValidatorModule) IsAlphanumeric(value string) bool {
	return v.validate.Var(value, "alphanum") == nil
}

// IsJSON checks if value is valid JSON
func (v *ValidatorModule) IsJSON(value string) bool {
	return v.validate.Var(value, "json") == nil
}

// IsBase64 checks if value is valid base64
func (v *ValidatorModule) IsBase64(value string) bool {
	return v.validate.Var(value, "base64") == nil
}

// IsIP checks if value is valid IP address (v4 or v6)
func (v *ValidatorModule) IsIP(value string) bool {
	return v.validate.Var(value, "ip") == nil
}

// IsIPv4 checks if value is valid IPv4 address
func (v *ValidatorModule) IsIPv4(value string) bool {
	return v.validate.Var(value, "ipv4") == nil
}

// IsIPv6 checks if value is valid IPv6 address
func (v *ValidatorModule) IsIPv6(value string) bool {
	return v.validate.Var(value, "ipv6") == nil
}

// IsCIDR checks if value is valid CIDR notation
func (v *ValidatorModule) IsCIDR(value string) bool {
	return v.validate.Var(value, "cidr") == nil
}

// IsMAC checks if value is valid MAC address
func (v *ValidatorModule) IsMAC(value string) bool {
	return v.validate.Var(value, "mac") == nil
}

// IsHexColor checks if value is valid hex color
func (v *ValidatorModule) IsHexColor(value string) bool {
	return v.validate.Var(value, "hexcolor") == nil
}

// IsRGBColor checks if value is valid RGB color
func (v *ValidatorModule) IsRGBColor(value string) bool {
	return v.validate.Var(value, "rgb") == nil
}

// IsRGBAColor checks if value is valid RGBA color
func (v *ValidatorModule) IsRGBAColor(value string) bool {
	return v.validate.Var(value, "rgba") == nil
}

// IsLatitude checks if value is valid latitude
func (v *ValidatorModule) IsLatitude(value interface{}) bool {
	return v.validate.Var(value, "latitude") == nil
}

// IsLongitude checks if value is valid longitude
func (v *ValidatorModule) IsLongitude(value interface{}) bool {
	return v.validate.Var(value, "longitude") == nil
}

// IsCreditCard checks if value is valid credit card number
func (v *ValidatorModule) IsCreditCard(value string) bool {
	return v.validate.Var(value, "credit_card") == nil
}

// IsISBN checks if value is valid ISBN (10 or 13)
func (v *ValidatorModule) IsISBN(value string) bool {
	return v.validate.Var(value, "isbn") == nil
}

// Contains checks if value contains the substring
func (v *ValidatorModule) Contains(value, substring string) bool {
	return v.validate.Var(value, "contains="+substring) == nil
}

// StartsWith checks if value starts with the prefix
func (v *ValidatorModule) StartsWith(value, prefix string) bool {
	return v.validate.Var(value, "startswith="+prefix) == nil
}

// EndsWith checks if value ends with the suffix
func (v *ValidatorModule) EndsWith(value, suffix string) bool {
	return v.validate.Var(value, "endswith="+suffix) == nil
}

// MinLength checks if string length is at least min
func (v *ValidatorModule) MinLength(value string, min int) bool {
	return len(value) >= min
}

// MaxLength checks if string length is at most max
func (v *ValidatorModule) MaxLength(value string, max int) bool {
	return len(value) <= max
}

// Length checks if string length is exactly length
func (v *ValidatorModule) Length(value string, length int) bool {
	return len(value) == length
}

// LengthBetween checks if string length is between min and max (inclusive)
func (v *ValidatorModule) LengthBetween(value string, min, max int) bool {
	l := len(value)
	return l >= min && l <= max
}

// Min checks if number is at least min
func (v *ValidatorModule) Min(value interface{}, min float64) bool {
	num, ok := toFloat64(value)
	if !ok {
		return false
	}
	return num >= min
}

// Max checks if number is at most max
func (v *ValidatorModule) Max(value interface{}, max float64) bool {
	num, ok := toFloat64(value)
	if !ok {
		return false
	}
	return num <= max
}

// Between checks if number is between min and max (inclusive)
func (v *ValidatorModule) Between(value interface{}, min, max float64) bool {
	num, ok := toFloat64(value)
	if !ok {
		return false
	}
	return num >= min && num <= max
}

// Matches checks if value matches the regex pattern
func (v *ValidatorModule) Matches(value, pattern string) bool {
	matched, err := regexp.MatchString(pattern, value)
	return err == nil && matched
}

// OneOf checks if value is one of the allowed values
func (v *ValidatorModule) OneOf(value interface{}, allowed []interface{}) bool {
	for _, a := range allowed {
		if reflect.DeepEqual(value, a) {
			return true
		}
	}
	return false
}

// NotEmpty checks if value is not empty (works for strings, arrays, maps)
func (v *ValidatorModule) NotEmpty(value interface{}) bool {
	if value == nil {
		return false
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() > 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() > 0
	default:
		return true
	}
}

// Required checks if value is not nil/undefined
func (v *ValidatorModule) Required(value interface{}) bool {
	return value != nil
}

// Helper functions

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return ""
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

func formatValidationMessage(field string, e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "url":
		return field + " must be a valid URL"
	case "min":
		return field + " must be at least " + e.Param()
	case "max":
		return field + " must be at most " + e.Param()
	case "len":
		return field + " must be exactly " + e.Param() + " characters"
	case "uuid":
		return field + " must be a valid UUID"
	case "numeric":
		return field + " must contain only numbers"
	case "alpha":
		return field + " must contain only letters"
	case "alphanum":
		return field + " must contain only letters and numbers"
	default:
		return field + " failed validation: " + e.Tag()
	}
}

func formatValidationMessageSimple(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "value is required"
	case "email":
		return "must be a valid email"
	case "url":
		return "must be a valid URL"
	case "min":
		return "must be at least " + e.Param()
	case "max":
		return "must be at most " + e.Param()
	case "len":
		return "must be exactly " + e.Param() + " characters"
	case "uuid":
		return "must be a valid UUID"
	case "numeric":
		return "must contain only numbers"
	case "alpha":
		return "must contain only letters"
	case "alphanum":
		return "must contain only letters and numbers"
	default:
		return "failed validation: " + e.Tag()
	}
}

// GetSchema implements JSSchemaProvider
func (v *ValidatorModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$validator",
		Description: "Data validation utilities powered by go-playground/validator",
		Types: []schema.TypeSchema{
			{
				Name:        "ValidationError",
				Description: "Single validation error",
				Fields: []schema.ParamSchema{
					{Name: "field", Type: "string", Description: "Field name that failed validation"},
					{Name: "tag", Type: "string", Description: "Validation rule that failed"},
					{Name: "value", Type: "string", Description: "The value that failed validation"},
					{Name: "message", Type: "string", Description: "Human readable error message"},
				},
			},
			{
				Name:        "ValidationResult",
				Description: "Result of validation",
				Fields: []schema.ParamSchema{
					{Name: "valid", Type: "boolean", Description: "Whether validation passed"},
					{Name: "errors", Type: "ValidationError[]", Description: "Array of validation errors"},
				},
			},
		},
		Methods: []schema.MethodSchema{
			// Core validation
			{
				Name:        "struct",
				Description: "Validate object against rules. Rules use go-playground/validator syntax (e.g., 'required,email', 'min=1,max=100')",
				Params: []schema.ParamSchema{
					{Name: "data", Type: "object", Description: "Object to validate"},
					{Name: "rules", Type: "object", Description: "Object mapping field names to validation rules"},
				},
				Returns: &schema.ParamSchema{Type: "ValidationResult", Description: "Validation result with errors"},
			},
			{
				Name:        "var",
				Description: "Validate a single value against rules",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "any", Description: "Value to validate"},
					{Name: "rules", Type: "string", Description: "Validation rules (e.g., 'required,email')"},
				},
				Returns: &schema.ParamSchema{Type: "ValidationResult", Description: "Validation result with errors"},
			},
			{
				Name:        "isValid",
				Description: "Quick validation that returns boolean",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "any", Description: "Value to validate"},
					{Name: "rules", Type: "string", Description: "Validation rules"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			// String validators
			{
				Name:        "isEmail",
				Description: "Check if value is a valid email address",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isURL",
				Description: "Check if value is a valid URL",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isUUID",
				Description: "Check if value is a valid UUID (any version)",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isUUIDv4",
				Description: "Check if value is a valid UUID v4",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isNumeric",
				Description: "Check if value contains only numeric characters",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isAlpha",
				Description: "Check if value contains only alphabetic characters",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isAlphanumeric",
				Description: "Check if value contains only alphanumeric characters",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isJSON",
				Description: "Check if value is valid JSON",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isBase64",
				Description: "Check if value is valid base64",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			// Network validators
			{
				Name:        "isIP",
				Description: "Check if value is valid IP address (v4 or v6)",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isIPv4",
				Description: "Check if value is valid IPv4 address",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isIPv6",
				Description: "Check if value is valid IPv6 address",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isCIDR",
				Description: "Check if value is valid CIDR notation",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isMAC",
				Description: "Check if value is valid MAC address",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			// Color validators
			{
				Name:        "isHexColor",
				Description: "Check if value is valid hex color (e.g., #fff, #ffffff)",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isRGBColor",
				Description: "Check if value is valid RGB color",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isRGBAColor",
				Description: "Check if value is valid RGBA color",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			// Geo validators
			{
				Name:        "isLatitude",
				Description: "Check if value is valid latitude (-90 to 90)",
				Params:      []schema.ParamSchema{{Name: "value", Type: "number"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isLongitude",
				Description: "Check if value is valid longitude (-180 to 180)",
				Params:      []schema.ParamSchema{{Name: "value", Type: "number"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			// Other validators
			{
				Name:        "isCreditCard",
				Description: "Check if value is valid credit card number (Luhn algorithm)",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "isISBN",
				Description: "Check if value is valid ISBN (10 or 13)",
				Params:      []schema.ParamSchema{{Name: "value", Type: "string"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			// String comparison
			{
				Name:        "contains",
				Description: "Check if value contains the substring",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "string"},
					{Name: "substring", Type: "string"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "startsWith",
				Description: "Check if value starts with the prefix",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "string"},
					{Name: "prefix", Type: "string"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "endsWith",
				Description: "Check if value ends with the suffix",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "string"},
					{Name: "suffix", Type: "string"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			// Length validators
			{
				Name:        "minLength",
				Description: "Check if string length is at least min",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "string"},
					{Name: "min", Type: "number"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "maxLength",
				Description: "Check if string length is at most max",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "string"},
					{Name: "max", Type: "number"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "length",
				Description: "Check if string length is exactly length",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "string"},
					{Name: "length", Type: "number"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "lengthBetween",
				Description: "Check if string length is between min and max (inclusive)",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "string"},
					{Name: "min", Type: "number"},
					{Name: "max", Type: "number"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			// Number validators
			{
				Name:        "min",
				Description: "Check if number is at least min",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "number"},
					{Name: "min", Type: "number"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "max",
				Description: "Check if number is at most max",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "number"},
					{Name: "max", Type: "number"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "between",
				Description: "Check if number is between min and max (inclusive)",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "number"},
					{Name: "min", Type: "number"},
					{Name: "max", Type: "number"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			// Pattern and choice validators
			{
				Name:        "matches",
				Description: "Check if value matches the regex pattern",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "string"},
					{Name: "pattern", Type: "string", Description: "Regular expression pattern"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "oneOf",
				Description: "Check if value is one of the allowed values",
				Params: []schema.ParamSchema{
					{Name: "value", Type: "any"},
					{Name: "allowed", Type: "any[]", Description: "Array of allowed values"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			// General validators
			{
				Name:        "notEmpty",
				Description: "Check if value is not empty (works for strings, arrays, objects)",
				Params:      []schema.ParamSchema{{Name: "value", Type: "any"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "required",
				Description: "Check if value is not null/undefined",
				Params:      []schema.ParamSchema{{Name: "value", Type: "any"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
		},
	}
}

// GetValidatorSchema returns the validator schema (static version)
func GetValidatorSchema() schema.ModuleSchema {
	return (&ValidatorModule{}).GetSchema()
}
