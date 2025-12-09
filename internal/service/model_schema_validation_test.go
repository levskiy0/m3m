package service

import (
	"testing"

	"github.com/levskiy0/m3m/internal/domain"
)

func TestModelSchemaValidator_ValidateModel(t *testing.T) {
	validator := NewModelSchemaValidator()

	tests := []struct {
		name    string
		model   *domain.CreateModelRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid model",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString, Required: true},
					{Key: "email", Type: domain.FieldTypeString, Required: true},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			model: &domain.CreateModelRequest{
				Name: "",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty slug",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
			},
			wantErr: true,
			errMsg:  "slug is required",
		},
		{
			name: "invalid slug - starts with number",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "123users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
			},
			wantErr: true,
			errMsg:  "slug must start with a letter",
		},
		{
			name: "invalid slug - uppercase",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "Users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
			},
			wantErr: true,
			errMsg:  "slug must start with a letter",
		},
		{
			name: "valid slug with underscore",
			model: &domain.CreateModelRequest{
				Name: "User Profiles",
				Slug: "user_profiles",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
			},
			wantErr: false,
		},
		{
			name: "valid slug with hyphen",
			model: &domain.CreateModelRequest{
				Name: "User Profiles",
				Slug: "user-profiles",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
			},
			wantErr: false,
		},
		{
			name: "no fields",
			model: &domain.CreateModelRequest{
				Name:   "Users",
				Slug:   "users",
				Fields: []domain.ModelField{},
			},
			wantErr: true,
			errMsg:  "at least one field is required",
		},
		{
			name: "duplicate field keys",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
					{Key: "name", Type: domain.FieldTypeString},
				},
			},
			wantErr: true,
			errMsg:  "duplicate field key",
		},
		{
			name: "reserved field key - id",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "id", Type: domain.FieldTypeString},
				},
			},
			wantErr: true,
			errMsg:  "field key 'id' is reserved",
		},
		{
			name: "reserved field key - _id",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "_id", Type: domain.FieldTypeString},
				},
			},
			wantErr: true,
			errMsg:  "field key '_id' is reserved",
		},
		{
			name: "invalid field key - starts with number",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "1name", Type: domain.FieldTypeString},
				},
			},
			wantErr: true,
			errMsg:  "field key must start with a letter",
		},
		{
			name: "invalid field type",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: "invalid_type"},
				},
			},
			wantErr: true,
			errMsg:  "invalid field type",
		},
		{
			name: "ref type without ref_model",
			model: &domain.CreateModelRequest{
				Name: "Posts",
				Slug: "posts",
				Fields: []domain.ModelField{
					{Key: "author", Type: domain.FieldTypeRef, RefModel: ""},
				},
			},
			wantErr: true,
			errMsg:  "ref_model is required for ref type",
		},
		{
			name: "ref type with ref_model",
			model: &domain.CreateModelRequest{
				Name: "Posts",
				Slug: "posts",
				Fields: []domain.ModelField{
					{Key: "title", Type: domain.FieldTypeString, Required: true},
					{Key: "author", Type: domain.FieldTypeRef, RefModel: "users"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateModel(tt.model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestModelSchemaValidator_ValidateDefaultValue(t *testing.T) {
	validator := NewModelSchemaValidator()

	tests := []struct {
		name    string
		model   *domain.CreateModelRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid string default",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "status", Type: domain.FieldTypeString, DefaultValue: "active"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid string default (number)",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "status", Type: domain.FieldTypeString, DefaultValue: 123},
				},
			},
			wantErr: true,
			errMsg:  "default value must be a string",
		},
		{
			name: "valid number default",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "age", Type: domain.FieldTypeNumber, DefaultValue: 18},
				},
			},
			wantErr: false,
		},
		{
			name: "valid bool default",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "active", Type: domain.FieldTypeBool, DefaultValue: true},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid bool default (string)",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "active", Type: domain.FieldTypeBool, DefaultValue: "true"},
				},
			},
			wantErr: true,
			errMsg:  "default value must be a boolean",
		},
		{
			name: "valid date default",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "birth_date", Type: domain.FieldTypeDate, DefaultValue: "2000-01-01"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid date default format",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "birth_date", Type: domain.FieldTypeDate, DefaultValue: "01-01-2000"},
				},
			},
			wantErr: true,
			errMsg:  "expected date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateModel(tt.model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestModelSchemaValidator_ValidateTableConfig(t *testing.T) {
	validator := NewModelSchemaValidator()

	tests := []struct {
		name    string
		model   *domain.CreateModelRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid table config",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
					{Key: "email", Type: domain.FieldTypeString},
				},
				TableConfig: &domain.TableConfig{
					Columns:     []string{"name", "email"},
					Filters:     []string{"name"},
					SortColumns: []string{"name", "email"},
				},
			},
			wantErr: false,
		},
		{
			name: "table config with unknown column",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
				TableConfig: &domain.TableConfig{
					Columns: []string{"name", "unknown_field"},
				},
			},
			wantErr: true,
			errMsg:  "unknown field 'unknown_field'",
		},
		{
			name: "table config with unknown filter",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
				TableConfig: &domain.TableConfig{
					Columns: []string{"name"},
					Filters: []string{"unknown_filter"},
				},
			},
			wantErr: true,
			errMsg:  "unknown field 'unknown_filter'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateModel(tt.model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestModelSchemaValidator_ValidateFormConfig(t *testing.T) {
	validator := NewModelSchemaValidator()

	tests := []struct {
		name    string
		model   *domain.CreateModelRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid form config",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
					{Key: "bio", Type: domain.FieldTypeText},
				},
				FormConfig: &domain.FormConfig{
					FieldOrder:   []string{"name", "bio"},
					HiddenFields: []string{},
					FieldViews:   map[string]string{"bio": "tiptap"},
				},
			},
			wantErr: false,
		},
		{
			name: "form config with unknown field in order",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
				FormConfig: &domain.FormConfig{
					FieldOrder: []string{"name", "unknown"},
				},
			},
			wantErr: true,
			errMsg:  "unknown field 'unknown'",
		},
		{
			name: "form config with invalid view for field type",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "name", Type: domain.FieldTypeString},
				},
				FormConfig: &domain.FormConfig{
					FieldViews: map[string]string{"name": "tiptap"}, // tiptap is for text, not string
				},
			},
			wantErr: true,
			errMsg:  "invalid view 'tiptap' for field type 'string'",
		},
		{
			name: "valid text field with textarea view",
			model: &domain.CreateModelRequest{
				Name: "Users",
				Slug: "users",
				Fields: []domain.ModelField{
					{Key: "bio", Type: domain.FieldTypeText},
				},
				FormConfig: &domain.FormConfig{
					FieldViews: map[string]string{"bio": "textarea"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateModel(tt.model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
