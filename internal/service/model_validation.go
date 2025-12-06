package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// IsValidationError checks if error is a ValidationErrors
func IsValidationError(err error) bool {
	var ve ValidationErrors
	return errors.As(err, &ve)
}

// DataValidator validates model data
type DataValidator struct {
	model *domain.Model
}

// NewDataValidator creates a new validator for a model
func NewDataValidator(model *domain.Model) *DataValidator {
	return &DataValidator{model: model}
}

// Validate validates data against model schema
func (v *DataValidator) Validate(data map[string]interface{}) error {
	var validationErrors []ValidationError

	// Build field map for quick lookup
	fieldMap := make(map[string]domain.ModelField)
	for _, f := range v.model.Fields {
		fieldMap[f.Key] = f
	}

	// Check required fields
	for _, field := range v.model.Fields {
		val, exists := data[field.Key]

		if field.Required {
			if !exists || val == nil || val == "" {
				validationErrors = append(validationErrors, ValidationError{
					Field:   field.Key,
					Message: "field is required",
				})
				continue
			}
		}

		// Skip validation if field is not present/empty and not required
		if !exists || val == nil || val == "" {
			continue
		}

		// Validate field type
		if err := v.validateFieldType(field, val); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   field.Key,
				Message: err.Error(),
			})
		}
	}

	// Check for unknown fields
	for key := range data {
		if _, exists := fieldMap[key]; !exists {
			validationErrors = append(validationErrors, ValidationError{
				Field:   key,
				Message: "unknown field",
			})
		}
	}

	if len(validationErrors) > 0 {
		return ValidationErrors{Errors: validationErrors}
	}

	return nil
}

// ValidatePartial validates partial data (for updates)
func (v *DataValidator) ValidatePartial(data map[string]interface{}) error {
	var validationErrors []ValidationError

	// Build field map for quick lookup
	fieldMap := make(map[string]domain.ModelField)
	for _, f := range v.model.Fields {
		fieldMap[f.Key] = f
	}

	// Only validate provided fields
	for key, val := range data {
		field, exists := fieldMap[key]
		if !exists {
			validationErrors = append(validationErrors, ValidationError{
				Field:   key,
				Message: "unknown field",
			})
			continue
		}

		// Skip nil values (removing the field)
		if val == nil {
			continue
		}

		// Validate field type
		if err := v.validateFieldType(field, val); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   key,
				Message: err.Error(),
			})
		}
	}

	if len(validationErrors) > 0 {
		return ValidationErrors{Errors: validationErrors}
	}

	return nil
}

// validateFieldType validates a single field value against its type
func (v *DataValidator) validateFieldType(field domain.ModelField, value interface{}) error {
	switch field.Type {
	case domain.FieldTypeString:
		return v.validateString(value)
	case domain.FieldTypeText:
		return v.validateString(value)
	case domain.FieldTypeNumber:
		return v.validateNumber(value)
	case domain.FieldTypeFloat:
		return v.validateFloat(value)
	case domain.FieldTypeBool:
		return v.validateBool(value)
	case domain.FieldTypeDocument:
		return v.validateDocument(value)
	case domain.FieldTypeFile:
		return v.validateString(value) // File is stored as string path/URL
	case domain.FieldTypeRef:
		return v.validateRef(value)
	case domain.FieldTypeDate:
		return v.validateDate(value)
	case domain.FieldTypeDateTime:
		return v.validateDateTime(value)
	default:
		return fmt.Errorf("unknown field type: %s", field.Type)
	}
}

func (v *DataValidator) validateString(value interface{}) error {
	switch value.(type) {
	case string:
		return nil
	default:
		return errors.New("expected string value")
	}
}

func (v *DataValidator) validateNumber(value interface{}) error {
	switch val := value.(type) {
	case int, int32, int64:
		return nil
	case float64:
		// JSON numbers come as float64, check if it's an integer
		if val == float64(int64(val)) {
			return nil
		}
		return errors.New("expected integer value, got float")
	case string:
		// Try to parse string as number
		if _, err := strconv.ParseInt(val, 10, 64); err == nil {
			return nil
		}
		return errors.New("expected integer value")
	default:
		return errors.New("expected integer value")
	}
}

func (v *DataValidator) validateFloat(value interface{}) error {
	switch val := value.(type) {
	case float32, float64, int, int32, int64:
		return nil
	case string:
		// Try to parse string as float
		if _, err := strconv.ParseFloat(val, 64); err == nil {
			return nil
		}
		return errors.New("expected numeric value")
	default:
		return errors.New("expected numeric value")
	}
}

func (v *DataValidator) validateBool(value interface{}) error {
	switch val := value.(type) {
	case bool:
		return nil
	case string:
		lower := strings.ToLower(val)
		if lower == "true" || lower == "false" || lower == "1" || lower == "0" {
			return nil
		}
		return errors.New("expected boolean value")
	case int, int32, int64, float64:
		return nil // Any number can be coerced to bool (0 = false, else = true)
	default:
		return errors.New("expected boolean value")
	}
}

func (v *DataValidator) validateDocument(value interface{}) error {
	switch value.(type) {
	case map[string]interface{}:
		return nil
	default:
		return errors.New("expected object value")
	}
}

func (v *DataValidator) validateRef(value interface{}) error {
	switch val := value.(type) {
	case string:
		// Must be a valid ObjectID hex string
		if _, err := primitive.ObjectIDFromHex(val); err != nil {
			return errors.New("expected valid reference ID")
		}
		return nil
	case primitive.ObjectID:
		return nil
	default:
		return errors.New("expected reference ID string")
	}
}

// Date format: YYYY-MM-DD
var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func (v *DataValidator) validateDate(value interface{}) error {
	switch val := value.(type) {
	case string:
		if !dateRegex.MatchString(val) {
			return errors.New("expected date format YYYY-MM-DD")
		}
		// Validate actual date
		if _, err := time.Parse("2006-01-02", val); err != nil {
			return errors.New("invalid date value")
		}
		return nil
	case time.Time:
		return nil
	default:
		return errors.New("expected date string (YYYY-MM-DD)")
	}
}

// DateTime format: ISO 8601 (YYYY-MM-DDTHH:MM or YYYY-MM-DDTHH:MM:SS with optional timezone)
var dateTimeRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}(:\d{2})?(Z|[+-]\d{2}:\d{2})?)?$`)

func (v *DataValidator) validateDateTime(value interface{}) error {
	switch val := value.(type) {
	case string:
		if !dateTimeRegex.MatchString(val) {
			return errors.New("expected datetime format ISO 8601 (YYYY-MM-DDTHH:MM or YYYY-MM-DDTHH:MM:SS)")
		}
		// Try parsing with various formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04", // Without seconds
			"2006-01-02",
		}
		for _, format := range formats {
			if _, err := time.Parse(format, val); err == nil {
				return nil
			}
		}
		return errors.New("invalid datetime value")
	case time.Time:
		return nil
	default:
		return errors.New("expected datetime string (ISO 8601)")
	}
}

// CoerceValue attempts to convert a value to the expected type
func (v *DataValidator) CoerceValue(field domain.ModelField, value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch field.Type {
	case domain.FieldTypeNumber:
		return v.coerceNumber(value)
	case domain.FieldTypeFloat:
		return v.coerceFloat(value)
	case domain.FieldTypeBool:
		return v.coerceBool(value)
	case domain.FieldTypeDate, domain.FieldTypeDateTime:
		return v.coerceDate(value)
	default:
		return value
	}
}

func (v *DataValidator) coerceNumber(value interface{}) interface{} {
	switch val := value.(type) {
	case float64:
		return int64(val)
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return value
}

func (v *DataValidator) coerceFloat(value interface{}) interface{} {
	switch val := value.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return value
}

func (v *DataValidator) coerceBool(value interface{}) interface{} {
	switch val := value.(type) {
	case string:
		lower := strings.ToLower(val)
		if lower == "true" || lower == "1" {
			return true
		}
		if lower == "false" || lower == "0" {
			return false
		}
	case int, int32, int64:
		return val != 0
	case float64:
		return val != 0
	}
	return value
}

func (v *DataValidator) coerceDate(value interface{}) interface{} {
	switch val := value.(type) {
	case string:
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t
			}
		}
	}
	return value
}

// ValidateAndCoerce validates and coerces data, returning processed data
func (v *DataValidator) ValidateAndCoerce(data map[string]interface{}) (map[string]interface{}, error) {
	if err := v.Validate(data); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	fieldMap := make(map[string]domain.ModelField)
	for _, f := range v.model.Fields {
		fieldMap[f.Key] = f
	}

	for key, value := range data {
		if field, exists := fieldMap[key]; exists {
			result[key] = v.CoerceValue(field, value)
		}
	}

	return result, nil
}
