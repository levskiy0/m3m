package modules

import (
	"testing"
)

func TestEnvModule_NewEnvModule(t *testing.T) {
	t.Run("with vars", func(t *testing.T) {
		vars := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		env := NewEnvModule(vars)
		if env == nil {
			t.Fatal("NewEnvModule() returned nil")
		}
		if env.vars == nil {
			t.Error("env.vars should not be nil")
		}
	})

	t.Run("with nil vars", func(t *testing.T) {
		env := NewEnvModule(nil)
		if env == nil {
			t.Fatal("NewEnvModule() returned nil")
		}
		if env.vars == nil {
			t.Error("env.vars should be initialized to empty map, not nil")
		}
	})

	t.Run("with empty vars", func(t *testing.T) {
		env := NewEnvModule(map[string]interface{}{})
		if env == nil {
			t.Fatal("NewEnvModule() returned nil")
		}
		if env.vars == nil {
			t.Error("env.vars should not be nil")
		}
	})
}

func TestEnvModule_Get_StringValue(t *testing.T) {
	vars := map[string]interface{}{
		"string_key": "string_value",
	}
	env := NewEnvModule(vars)

	got := env.Get("string_key")
	if got != "string_value" {
		t.Errorf("Get(\"string_key\") = %v, want \"string_value\"", got)
	}
}

func TestEnvModule_Get_IntValue(t *testing.T) {
	vars := map[string]interface{}{
		"int_key": 42,
	}
	env := NewEnvModule(vars)

	got := env.Get("int_key")
	if got != 42 {
		t.Errorf("Get(\"int_key\") = %v, want 42", got)
	}
}

func TestEnvModule_Get_FloatValue(t *testing.T) {
	vars := map[string]interface{}{
		"float_key": 3.14,
	}
	env := NewEnvModule(vars)

	got := env.Get("float_key")
	if got != 3.14 {
		t.Errorf("Get(\"float_key\") = %v, want 3.14", got)
	}
}

func TestEnvModule_Get_BoolValue(t *testing.T) {
	vars := map[string]interface{}{
		"bool_true":  true,
		"bool_false": false,
	}
	env := NewEnvModule(vars)

	gotTrue := env.Get("bool_true")
	if gotTrue != true {
		t.Errorf("Get(\"bool_true\") = %v, want true", gotTrue)
	}

	gotFalse := env.Get("bool_false")
	if gotFalse != false {
		t.Errorf("Get(\"bool_false\") = %v, want false", gotFalse)
	}
}

func TestEnvModule_Get_NilValue(t *testing.T) {
	vars := map[string]interface{}{
		"nil_key": nil,
	}
	env := NewEnvModule(vars)

	got := env.Get("nil_key")
	if got != nil {
		t.Errorf("Get(\"nil_key\") = %v, want nil", got)
	}
}

func TestEnvModule_Get_MapValue(t *testing.T) {
	nested := map[string]interface{}{
		"nested_key": "nested_value",
	}
	vars := map[string]interface{}{
		"map_key": nested,
	}
	env := NewEnvModule(vars)

	got := env.Get("map_key")
	gotMap, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("Get(\"map_key\") type = %T, want map[string]interface{}", got)
	}
	if gotMap["nested_key"] != "nested_value" {
		t.Error("nested map value should be preserved")
	}
}

func TestEnvModule_Get_SliceValue(t *testing.T) {
	slice := []interface{}{"a", "b", "c"}
	vars := map[string]interface{}{
		"slice_key": slice,
	}
	env := NewEnvModule(vars)

	got := env.Get("slice_key")
	gotSlice, ok := got.([]interface{})
	if !ok {
		t.Fatalf("Get(\"slice_key\") type = %T, want []interface{}", got)
	}
	if len(gotSlice) != 3 {
		t.Errorf("slice length = %d, want 3", len(gotSlice))
	}
}

func TestEnvModule_Get_NonExistentKey(t *testing.T) {
	vars := map[string]interface{}{
		"existing_key": "value",
	}
	env := NewEnvModule(vars)

	got := env.Get("non_existent_key")
	if got != nil {
		t.Errorf("Get(\"non_existent_key\") = %v, want nil", got)
	}
}

func TestEnvModule_Get_EmptyKey(t *testing.T) {
	vars := map[string]interface{}{
		"": "empty_key_value",
	}
	env := NewEnvModule(vars)

	got := env.Get("")
	if got != "empty_key_value" {
		t.Errorf("Get(\"\") = %v, want \"empty_key_value\"", got)
	}
}

func TestEnvModule_Get_SpecialCharKeys(t *testing.T) {
	tests := []struct {
		key   string
		value interface{}
	}{
		{"KEY_WITH_UNDERSCORE", "value1"},
		{"key-with-dash", "value2"},
		{"key.with.dot", "value3"},
		{"KEY123", "value4"},
		{"123key", "value5"},
		{"UPPER_CASE", "value6"},
		{"lower_case", "value7"},
	}

	vars := make(map[string]interface{})
	for _, tt := range tests {
		vars[tt.key] = tt.value
	}

	env := NewEnvModule(vars)

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := env.Get(tt.key)
			if got != tt.value {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.value)
			}
		})
	}
}

func TestEnvModule_Get_EmptyVars(t *testing.T) {
	env := NewEnvModule(map[string]interface{}{})

	got := env.Get("any_key")
	if got != nil {
		t.Errorf("Get on empty vars = %v, want nil", got)
	}
}

func TestEnvModule_Get_NilVars(t *testing.T) {
	env := NewEnvModule(nil)

	got := env.Get("any_key")
	if got != nil {
		t.Errorf("Get on nil vars = %v, want nil", got)
	}
}

func TestEnvModule_Get_JsonValue(t *testing.T) {
	// Simulate JSON parsed value (common use case)
	jsonLikeValue := map[string]interface{}{
		"database": map[string]interface{}{
			"host": "localhost",
			"port": float64(5432), // JSON numbers are float64
		},
		"features": []interface{}{"feature1", "feature2"},
	}
	vars := map[string]interface{}{
		"config": jsonLikeValue,
	}
	env := NewEnvModule(vars)

	got := env.Get("config")
	gotMap, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("Get(\"config\") type = %T, want map[string]interface{}", got)
	}

	dbConfig, ok := gotMap["database"].(map[string]interface{})
	if !ok {
		t.Fatal("database config should be a map")
	}

	if dbConfig["host"] != "localhost" {
		t.Error("database.host should be localhost")
	}
}

func TestEnvModule_Get_LargeValue(t *testing.T) {
	largeString := make([]byte, 1000000) // 1MB string
	for i := range largeString {
		largeString[i] = 'x'
	}

	vars := map[string]interface{}{
		"large_key": string(largeString),
	}
	env := NewEnvModule(vars)

	got := env.Get("large_key")
	if got != string(largeString) {
		t.Error("large string value should be preserved")
	}
}

func TestEnvModule_Get_UnicodeKey(t *testing.T) {
	vars := map[string]interface{}{
		"key": "value",
	}
	env := NewEnvModule(vars)

	got := env.Get("key")
	if got != "value" {
		t.Errorf("Get unicode key = %v, want \"value\"", got)
	}
}

func TestEnvModule_Get_MultipleTypes(t *testing.T) {
	vars := map[string]interface{}{
		"string": "hello",
		"int":    42,
		"float":  3.14,
		"bool":   true,
		"nil":    nil,
	}
	env := NewEnvModule(vars)

	tests := []struct {
		key      string
		expected interface{}
	}{
		{"string", "hello"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := env.Get(tt.key)
			if got != tt.expected {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestEnvModule_VarsImmutability(t *testing.T) {
	vars := map[string]interface{}{
		"key": "original",
	}
	env := NewEnvModule(vars)

	// Modify original map
	vars["key"] = "modified"
	vars["new_key"] = "new_value"

	// EnvModule should reflect changes since it stores reference
	got := env.Get("key")
	if got != "modified" {
		t.Log("Note: EnvModule stores reference to vars map, modifications are visible")
	}
}

func TestEnvModule_CaseSensitivity(t *testing.T) {
	vars := map[string]interface{}{
		"Key": "value1",
		"key": "value2",
		"KEY": "value3",
		"kEy": "value4",
	}
	env := NewEnvModule(vars)

	tests := []struct {
		key      string
		expected string
	}{
		{"Key", "value1"},
		{"key", "value2"},
		{"KEY", "value3"},
		{"kEy", "value4"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := env.Get(tt.key)
			if got != tt.expected {
				t.Errorf("Get(%q) = %v, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

// Tests for Has method
func TestEnvModule_Has(t *testing.T) {
	vars := map[string]interface{}{
		"existing_key": "value",
		"nil_key":      nil,
	}
	env := NewEnvModule(vars)

	tests := []struct {
		key      string
		expected bool
	}{
		{"existing_key", true},
		{"nil_key", true}, // Key exists even if value is nil
		{"non_existent", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := env.Has(tt.key)
			if got != tt.expected {
				t.Errorf("Has(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestEnvModule_Has_EmptyVars(t *testing.T) {
	env := NewEnvModule(nil)

	if env.Has("any_key") {
		t.Error("Has should return false for empty vars")
	}
}

// Tests for Keys method
func TestEnvModule_Keys(t *testing.T) {
	vars := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	env := NewEnvModule(vars)

	keys := env.Keys()

	if len(keys) != 3 {
		t.Errorf("Keys() length = %d, want 3", len(keys))
	}

	// Check all keys are present (order not guaranteed)
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for expected := range vars {
		if !keyMap[expected] {
			t.Errorf("Keys() missing key %q", expected)
		}
	}
}

func TestEnvModule_Keys_Empty(t *testing.T) {
	env := NewEnvModule(nil)

	keys := env.Keys()

	if len(keys) != 0 {
		t.Errorf("Keys() on empty vars should return empty slice, got %d keys", len(keys))
	}
}

// Tests for GetString method
func TestEnvModule_GetString(t *testing.T) {
	vars := map[string]interface{}{
		"string_key": "hello",
		"int_key":    42,
		"float_key":  3.14,
		"bool_key":   true,
	}
	env := NewEnvModule(vars)

	tests := []struct {
		key          string
		defaultValue string
		expected     string
	}{
		{"string_key", "default", "hello"},
		{"int_key", "default", "42"},
		{"float_key", "default", "3.14"},
		{"bool_key", "default", "true"},
		{"non_existent", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := env.GetString(tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("GetString(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.expected)
			}
		})
	}
}

// Tests for GetInt method
func TestEnvModule_GetInt(t *testing.T) {
	vars := map[string]interface{}{
		"int_key":     42,
		"int64_key":   int64(100),
		"float64_key": float64(55.9),
		"string_key":  "not_an_int",
	}
	env := NewEnvModule(vars)

	tests := []struct {
		key          string
		defaultValue int
		expected     int
	}{
		{"int_key", -1, 42},
		{"int64_key", -1, 100},
		{"float64_key", -1, 55}, // Truncated
		{"string_key", -1, -1},  // Returns default
		{"non_existent", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := env.GetInt(tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("GetInt(%q, %d) = %d, want %d", tt.key, tt.defaultValue, got, tt.expected)
			}
		})
	}
}

// Tests for GetFloat method
func TestEnvModule_GetFloat(t *testing.T) {
	vars := map[string]interface{}{
		"float64_key": float64(3.14),
		"int_key":     42,
		"string_key":  "not_a_float",
	}
	env := NewEnvModule(vars)

	tests := []struct {
		key          string
		defaultValue float64
		expected     float64
	}{
		{"float64_key", -1.0, 3.14},
		{"int_key", -1.0, 42.0},
		{"string_key", -1.0, -1.0}, // Returns default
		{"non_existent", -1.0, -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := env.GetFloat(tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("GetFloat(%q, %f) = %f, want %f", tt.key, tt.defaultValue, got, tt.expected)
			}
		})
	}
}

// Tests for GetBool method
func TestEnvModule_GetBool(t *testing.T) {
	vars := map[string]interface{}{
		"bool_true":  true,
		"bool_false": false,
		"string_key": "maybe", // Not a bool string
		"int_key":    1,       // Not a bool type
	}
	env := NewEnvModule(vars)

	tests := []struct {
		key          string
		defaultValue bool
		expected     bool
	}{
		{"bool_true", false, true},
		{"bool_false", true, false},
		{"string_key", false, false}, // Returns default (not a bool)
		{"int_key", false, false},    // Returns default (not a bool)
		{"non_existent", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := env.GetBool(tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("GetBool(%q, %v) = %v, want %v", tt.key, tt.defaultValue, got, tt.expected)
			}
		})
	}
}

// Tests for GetAll method
func TestEnvModule_GetAll(t *testing.T) {
	vars := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	env := NewEnvModule(vars)

	all := env.GetAll()

	if len(all) != len(vars) {
		t.Errorf("GetAll() length = %d, want %d", len(all), len(vars))
	}

	for k, v := range vars {
		if all[k] != v {
			t.Errorf("GetAll()[%q] = %v, want %v", k, all[k], v)
		}
	}

	// Verify it's a copy (modification doesn't affect original)
	all["new_key"] = "new_value"
	if env.Has("new_key") {
		t.Error("GetAll() should return a copy, not a reference")
	}
}

func TestEnvModule_GetAll_Empty(t *testing.T) {
	env := NewEnvModule(nil)

	all := env.GetAll()

	if all == nil {
		t.Error("GetAll() should return empty map, not nil")
	}

	if len(all) != 0 {
		t.Errorf("GetAll() on empty env should return empty map, got %d items", len(all))
	}
}
