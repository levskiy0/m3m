package service

import (
	"fmt"
	"regexp"
	"strings"

	"m3m/internal/domain"
)

// ModelSchemaValidator validates model schema definitions
type ModelSchemaValidator struct{}

// NewModelSchemaValidator creates a new schema validator
func NewModelSchemaValidator() *ModelSchemaValidator {
	return &ModelSchemaValidator{}
}

// ValidateModel validates a model schema
func (v *ModelSchemaValidator) ValidateModel(model *domain.CreateModelRequest) error {
	var errors []ValidationError

	// Validate name
	if strings.TrimSpace(model.Name) == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "name is required"})
	}

	// Validate slug
	if err := v.validateSlug(model.Slug); err != nil {
		errors = append(errors, ValidationError{Field: "slug", Message: err.Error()})
	}

	// Validate fields
	if len(model.Fields) == 0 {
		errors = append(errors, ValidationError{Field: "fields", Message: "at least one field is required"})
	} else {
		fieldErrors := v.validateFields(model.Fields)
		errors = append(errors, fieldErrors...)
	}

	// Validate table config if provided
	if model.TableConfig != nil {
		configErrors := v.validateTableConfig(model.TableConfig, model.Fields)
		errors = append(errors, configErrors...)
	}

	// Validate form config if provided
	if model.FormConfig != nil {
		configErrors := v.validateFormConfig(model.FormConfig, model.Fields)
		errors = append(errors, configErrors...)
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateModelUpdate validates model update request
func (v *ModelSchemaValidator) ValidateModelUpdate(req *domain.UpdateModelRequest, existingModel *domain.Model) error {
	var errors []ValidationError

	// Validate name if provided
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "name cannot be empty"})
	}

	// Validate fields if provided
	if req.Fields != nil {
		if len(*req.Fields) == 0 {
			errors = append(errors, ValidationError{Field: "fields", Message: "at least one field is required"})
		} else {
			fieldErrors := v.validateFields(*req.Fields)
			errors = append(errors, fieldErrors...)
		}
	}

	// Get the fields to validate config against
	fields := existingModel.Fields
	if req.Fields != nil {
		fields = *req.Fields
	}

	// Validate table config if provided
	if req.TableConfig != nil {
		configErrors := v.validateTableConfig(req.TableConfig, fields)
		errors = append(errors, configErrors...)
	}

	// Validate form config if provided
	if req.FormConfig != nil {
		configErrors := v.validateFormConfig(req.FormConfig, fields)
		errors = append(errors, configErrors...)
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// Slug validation: lowercase letters, numbers, underscores, hyphens
var slugRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func (v *ModelSchemaValidator) validateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug is required")
	}
	if len(slug) < 2 {
		return fmt.Errorf("slug must be at least 2 characters")
	}
	if len(slug) > 64 {
		return fmt.Errorf("slug must be at most 64 characters")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug must start with a letter and contain only lowercase letters, numbers, underscores, and hyphens")
	}
	return nil
}

// Field key validation: letters, numbers, underscores (must start with letter)
var fieldKeyRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// Reserved field keys
var reservedFields = map[string]bool{
	"_id":         true,
	"_model_id":   true,
	"_created_at": true,
	"_updated_at": true,
	"id":          true,
	"created_at":  true,
	"updated_at":  true,
}

func (v *ModelSchemaValidator) validateFields(fields []domain.ModelField) []ValidationError {
	var errors []ValidationError
	seenKeys := make(map[string]bool)

	validTypes := map[domain.FieldType]bool{
		domain.FieldTypeString:   true,
		domain.FieldTypeText:     true,
		domain.FieldTypeNumber:   true,
		domain.FieldTypeFloat:    true,
		domain.FieldTypeBool:     true,
		domain.FieldTypeDocument: true,
		domain.FieldTypeFile:     true,
		domain.FieldTypeRef:      true,
		domain.FieldTypeDate:     true,
		domain.FieldTypeDateTime: true,
	}

	for i, field := range fields {
		fieldPath := fmt.Sprintf("fields[%d]", i)

		// Validate key
		if field.Key == "" {
			errors = append(errors, ValidationError{
				Field:   fieldPath + ".key",
				Message: "field key is required",
			})
		} else {
			// Check key format
			if !fieldKeyRegex.MatchString(field.Key) {
				errors = append(errors, ValidationError{
					Field:   fieldPath + ".key",
					Message: "field key must start with a letter and contain only letters, numbers, and underscores",
				})
			}

			// Check reserved fields
			if reservedFields[field.Key] {
				errors = append(errors, ValidationError{
					Field:   fieldPath + ".key",
					Message: fmt.Sprintf("field key '%s' is reserved", field.Key),
				})
			}

			// Check uniqueness
			if seenKeys[field.Key] {
				errors = append(errors, ValidationError{
					Field:   fieldPath + ".key",
					Message: fmt.Sprintf("duplicate field key '%s'", field.Key),
				})
			}
			seenKeys[field.Key] = true
		}

		// Validate type
		if !validTypes[field.Type] {
			errors = append(errors, ValidationError{
				Field:   fieldPath + ".type",
				Message: fmt.Sprintf("invalid field type '%s'", field.Type),
			})
		}

		// Validate ref model for ref type
		if field.Type == domain.FieldTypeRef && field.RefModel == "" {
			errors = append(errors, ValidationError{
				Field:   fieldPath + ".ref_model",
				Message: "ref_model is required for ref type",
			})
		}

		// Validate default value type matches field type
		if field.DefaultValue != nil {
			if err := v.validateDefaultValue(field); err != nil {
				errors = append(errors, ValidationError{
					Field:   fieldPath + ".default_value",
					Message: err.Error(),
				})
			}
		}
	}

	return errors
}

func (v *ModelSchemaValidator) validateDefaultValue(field domain.ModelField) error {
	value := field.DefaultValue

	switch field.Type {
	case domain.FieldTypeString, domain.FieldTypeText, domain.FieldTypeFile:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("default value must be a string")
		}
	case domain.FieldTypeNumber:
		switch value.(type) {
		case int, int32, int64, float64:
			// OK
		default:
			return fmt.Errorf("default value must be a number")
		}
	case domain.FieldTypeFloat:
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// OK
		default:
			return fmt.Errorf("default value must be a number")
		}
	case domain.FieldTypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("default value must be a boolean")
		}
	case domain.FieldTypeDocument:
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("default value must be an object")
		}
	case domain.FieldTypeDate, domain.FieldTypeDateTime:
		if str, ok := value.(string); ok {
			// Allow special $now value for current date/time
			if str == "$now" {
				return nil
			}
			// Validate date format
			validator := &DataValidator{}
			if field.Type == domain.FieldTypeDate {
				if err := validator.validateDate(str); err != nil {
					return err
				}
			} else {
				if err := validator.validateDateTime(str); err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("default value must be a date string")
		}
	}

	return nil
}

// System fields that can be used in table config columns and sorting
var allowedSystemFields = map[string]bool{
	"_created_at": true,
	"_updated_at": true,
}

func (v *ModelSchemaValidator) validateTableConfig(config *domain.TableConfig, fields []domain.ModelField) []ValidationError {
	var errors []ValidationError
	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f.Key] = true
	}

	// Validate columns (allow system fields for display)
	for i, col := range config.Columns {
		if !fieldMap[col] && !allowedSystemFields[col] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("table_config.columns[%d]", i),
				Message: fmt.Sprintf("unknown field '%s'", col),
			})
		}
	}

	// Validate filters (only user-defined fields)
	for i, filter := range config.Filters {
		if !fieldMap[filter] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("table_config.filters[%d]", i),
				Message: fmt.Sprintf("unknown field '%s'", filter),
			})
		}
	}

	// Validate sort columns (allow system fields for sorting)
	for i, sortCol := range config.SortColumns {
		if !fieldMap[sortCol] && !allowedSystemFields[sortCol] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("table_config.sort_columns[%d]", i),
				Message: fmt.Sprintf("unknown field '%s'", sortCol),
			})
		}
	}

	// Build field type map for searchable validation
	fieldTypeMap := make(map[string]domain.FieldType)
	for _, f := range fields {
		fieldTypeMap[f.Key] = f.Type
	}

	// Validate searchable fields (only string and text types allowed)
	for i, searchField := range config.Searchable {
		if !fieldMap[searchField] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("table_config.searchable[%d]", i),
				Message: fmt.Sprintf("unknown field '%s'", searchField),
			})
			continue
		}
		// Check that searchable field is string or text type
		fieldType := fieldTypeMap[searchField]
		if fieldType != domain.FieldTypeString && fieldType != domain.FieldTypeText {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("table_config.searchable[%d]", i),
				Message: fmt.Sprintf("field '%s' must be string or text type to be searchable", searchField),
			})
		}
	}

	return errors
}

func (v *ModelSchemaValidator) validateFormConfig(config *domain.FormConfig, fields []domain.ModelField) []ValidationError {
	var errors []ValidationError
	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f.Key] = true
	}

	// Validate field order
	for i, fieldKey := range config.FieldOrder {
		if !fieldMap[fieldKey] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("form_config.field_order[%d]", i),
				Message: fmt.Sprintf("unknown field '%s'", fieldKey),
			})
		}
	}

	// Validate hidden fields
	for i, fieldKey := range config.HiddenFields {
		if !fieldMap[fieldKey] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("form_config.hidden_fields[%d]", i),
				Message: fmt.Sprintf("unknown field '%s'", fieldKey),
			})
		}
	}

	// Validate field views
	validViews := map[string]map[string]bool{
		"string":   {"input": true, "select": true},
		"text":     {"textarea": true, "tiptap": true, "markdown": true},
		"number":   {"input": true, "slider": true},
		"float":    {"input": true, "slider": true},
		"bool":     {"checkbox": true, "switch": true},
		"date":     {"datepicker": true, "input": true},
		"datetime": {"datetimepicker": true, "input": true},
		"file":     {"file": true, "image": true},
		"ref":      {"select": true, "combobox": true},
		"document": {"json": true},
	}

	// Build field type map
	fieldTypeMap := make(map[string]domain.FieldType)
	for _, f := range fields {
		fieldTypeMap[f.Key] = f.Type
	}

	for fieldKey, view := range config.FieldViews {
		if !fieldMap[fieldKey] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("form_config.field_views.%s", fieldKey),
				Message: fmt.Sprintf("unknown field '%s'", fieldKey),
			})
			continue
		}

		fieldType := fieldTypeMap[fieldKey]
		if allowedViews, ok := validViews[string(fieldType)]; ok {
			if !allowedViews[view] {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("form_config.field_views.%s", fieldKey),
					Message: fmt.Sprintf("invalid view '%s' for field type '%s'", view, fieldType),
				})
			}
		}
	}

	return errors
}
