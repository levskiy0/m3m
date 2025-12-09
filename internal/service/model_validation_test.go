package service

import (
	"testing"

	"github.com/levskiy0/m3m/internal/domain"
)

func TestDataValidator_Validate(t *testing.T) {
	model := &domain.Model{
		Fields: []domain.ModelField{
			{Key: "name", Type: domain.FieldTypeString, Required: true},
			{Key: "age", Type: domain.FieldTypeNumber, Required: false},
			{Key: "email", Type: domain.FieldTypeString, Required: true},
			{Key: "active", Type: domain.FieldTypeBool, Required: false},
			{Key: "score", Type: domain.FieldTypeFloat, Required: false},
			{Key: "birth_date", Type: domain.FieldTypeDate, Required: false},
			{Key: "created_at", Type: domain.FieldTypeDateTime, Required: false},
		},
	}

	validator := NewDataValidator(model)

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid data with all required fields",
			data: map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid data with all fields",
			data: map[string]interface{}{
				"name":       "John",
				"email":      "john@example.com",
				"age":        30,
				"active":     true,
				"score":      95.5,
				"birth_date": "1990-01-15",
				"created_at": "2024-01-15T10:30:00Z",
			},
			wantErr: false,
		},
		{
			name: "missing required field name",
			data: map[string]interface{}{
				"email": "john@example.com",
			},
			wantErr: true,
			errMsg:  "name: field is required",
		},
		{
			name: "missing required field email",
			data: map[string]interface{}{
				"name": "John",
			},
			wantErr: true,
			errMsg:  "email: field is required",
		},
		{
			name: "invalid type for age (expects number)",
			data: map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
				"age":   "thirty", // should be number
			},
			wantErr: true,
			errMsg:  "age: expected integer value",
		},
		{
			name: "invalid type for active (expects bool)",
			data: map[string]interface{}{
				"name":   "John",
				"email":  "john@example.com",
				"active": "yes", // should be bool
			},
			wantErr: true,
			errMsg:  "active: expected boolean value",
		},
		{
			name: "invalid date format",
			data: map[string]interface{}{
				"name":       "John",
				"email":      "john@example.com",
				"birth_date": "15-01-1990", // wrong format
			},
			wantErr: true,
			errMsg:  "birth_date: expected date format YYYY-MM-DD",
		},
		{
			name: "invalid datetime format",
			data: map[string]interface{}{
				"name":       "John",
				"email":      "john@example.com",
				"created_at": "2024/01/15 10:30", // wrong format
			},
			wantErr: true,
			errMsg:  "created_at: expected datetime format ISO 8601",
		},
		{
			name: "unknown field",
			data: map[string]interface{}{
				"name":    "John",
				"email":   "john@example.com",
				"unknown": "value",
			},
			wantErr: true,
			errMsg:  "unknown: unknown field",
		},
		{
			name: "float passed as number is OK (JSON parsing)",
			data: map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
				"age":   float64(30), // JSON numbers come as float64
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
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

func TestDataValidator_ValidatePartial(t *testing.T) {
	model := &domain.Model{
		Fields: []domain.ModelField{
			{Key: "name", Type: domain.FieldTypeString, Required: true},
			{Key: "age", Type: domain.FieldTypeNumber, Required: false},
		},
	}

	validator := NewDataValidator(model)

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "partial update with valid data",
			data:    map[string]interface{}{"name": "Jane"},
			wantErr: false,
		},
		{
			name:    "partial update - does not require all required fields",
			data:    map[string]interface{}{"age": 25},
			wantErr: false,
		},
		{
			name:    "partial update with unknown field",
			data:    map[string]interface{}{"unknown": "value"},
			wantErr: true,
		},
		{
			name:    "partial update with invalid type",
			data:    map[string]interface{}{"age": "twenty"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePartial(tt.data)
			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDataValidator_CoerceValue(t *testing.T) {
	model := &domain.Model{
		Fields: []domain.ModelField{
			{Key: "age", Type: domain.FieldTypeNumber},
			{Key: "score", Type: domain.FieldTypeFloat},
			{Key: "active", Type: domain.FieldTypeBool},
		},
	}

	validator := NewDataValidator(model)

	tests := []struct {
		name     string
		field    domain.ModelField
		value    interface{}
		expected interface{}
	}{
		{
			name:     "coerce float64 to int64 for number field",
			field:    domain.ModelField{Key: "age", Type: domain.FieldTypeNumber},
			value:    float64(25),
			expected: int64(25),
		},
		{
			name:     "coerce string to int64 for number field",
			field:    domain.ModelField{Key: "age", Type: domain.FieldTypeNumber},
			value:    "30",
			expected: int64(30),
		},
		{
			name:     "coerce int to float64 for float field",
			field:    domain.ModelField{Key: "score", Type: domain.FieldTypeFloat},
			value:    95,
			expected: float64(95),
		},
		{
			name:     "coerce string true to bool",
			field:    domain.ModelField{Key: "active", Type: domain.FieldTypeBool},
			value:    "true",
			expected: true,
		},
		{
			name:     "coerce string false to bool",
			field:    domain.ModelField{Key: "active", Type: domain.FieldTypeBool},
			value:    "false",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.CoerceValue(tt.field, tt.value)
			if result != tt.expected {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestDataValidator_RefValidation(t *testing.T) {
	model := &domain.Model{
		Fields: []domain.ModelField{
			{Key: "user_id", Type: domain.FieldTypeRef, Required: true, RefModel: "users"},
		},
	}

	validator := NewDataValidator(model)

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "valid ObjectID hex string",
			data:    map[string]interface{}{"user_id": "507f1f77bcf86cd799439011"},
			wantErr: false,
		},
		{
			name:    "invalid ObjectID hex string",
			data:    map[string]interface{}{"user_id": "invalid-id"},
			wantErr: true,
		},
		{
			name:    "empty ref required",
			data:    map[string]interface{}{"user_id": ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
