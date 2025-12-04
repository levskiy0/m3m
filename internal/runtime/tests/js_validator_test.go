package tests

import (
	"testing"
)

// ============== VALIDATOR MODULE TESTS ==============

func TestJS_Validator_IsEmail(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co", true},
		{"user+tag@example.org", true},
		{"invalid", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
		{"test@example", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isEmail("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isEmail(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsURL(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com", true},
		{"http://example.com/path", true},
		{"https://example.com:8080/path?query=1", true},
		{"ftp://files.example.com", true},
		{"invalid", false},
		{"example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isURL("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isURL(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsUUID(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"123e4567-e89b-12d3-a456-426614174000", true},
		{"invalid-uuid", false},
		{"550e8400-e29b-41d4-a716", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isUUID("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isUUID(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsNumeric(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"12345", true},
		{"0", true},
		{"123.456", true}, // numeric allows decimal numbers
		{"12a34", false},
		{"", false},
		{"-123", true}, // numeric allows negative numbers
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isNumeric("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isNumeric(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsAlpha(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"Hello", true},
		{"HelloWorld", true},
		{"hello123", false},
		{"hello world", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isAlpha("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isAlpha(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsAlphanumeric(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"hello123", true},
		{"Hello", true},
		{"12345", true},
		{"hello_world", false},
		{"hello-world", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isAlphanumeric("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isAlphanumeric(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsJSON(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid object", `{"key": "value"}`, true},
		{"valid array", `[1, 2, 3]`, true},
		{"valid string", `"hello"`, true},
		{"valid number", `123`, true},
		{"invalid", `{invalid}`, false},
		{"empty", ``, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use template literal for JSON strings
			code := "$validator.isJSON(`" + tt.input + "`)"
			result := h.MustRun(t, code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isJSON(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsIP(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.0", true},
		{"::1", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"invalid", false},
		{"256.256.256.256", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isIP("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isIP(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsIPv4(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.0", true},
		{"::1", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isIPv4("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isIPv4(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsHexColor(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected bool
	}{
		{"#fff", true},
		{"#ffffff", true},
		{"#FFF", true},
		{"#FFFFFF", true},
		{"fff", false},
		{"#gggggg", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `$validator.isHexColor("`+tt.input+`")`)
			if result.ToBoolean() != tt.expected {
				t.Errorf("isHexColor(%q) = %v, expected %v", tt.input, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsLatitude(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid positive", `$validator.isLatitude(45.5)`, true},
		{"valid negative", `$validator.isLatitude(-45.5)`, true},
		{"valid zero", `$validator.isLatitude(0)`, true},
		{"valid max", `$validator.isLatitude(90)`, true},
		{"valid min", `$validator.isLatitude(-90)`, true},
		{"invalid over", `$validator.isLatitude(91)`, false},
		{"invalid under", `$validator.isLatitude(-91)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsLongitude(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid positive", `$validator.isLongitude(120.5)`, true},
		{"valid negative", `$validator.isLongitude(-120.5)`, true},
		{"valid zero", `$validator.isLongitude(0)`, true},
		{"valid max", `$validator.isLongitude(180)`, true},
		{"valid min", `$validator.isLongitude(-180)`, true},
		{"invalid over", `$validator.isLongitude(181)`, false},
		{"invalid under", `$validator.isLongitude(-181)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_MinLength(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid", `$validator.minLength("hello", 3)`, true},
		{"exact", `$validator.minLength("hello", 5)`, true},
		{"invalid", `$validator.minLength("hi", 5)`, false},
		{"empty", `$validator.minLength("", 1)`, false},
		{"zero min", `$validator.minLength("", 0)`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_MaxLength(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid", `$validator.maxLength("hi", 5)`, true},
		{"exact", `$validator.maxLength("hello", 5)`, true},
		{"invalid", `$validator.maxLength("hello world", 5)`, false},
		{"empty", `$validator.maxLength("", 5)`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_LengthBetween(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid", `$validator.lengthBetween("hello", 3, 10)`, true},
		{"exact min", `$validator.lengthBetween("hello", 5, 10)`, true},
		{"exact max", `$validator.lengthBetween("hello", 3, 5)`, true},
		{"too short", `$validator.lengthBetween("hi", 3, 10)`, false},
		{"too long", `$validator.lengthBetween("hello world", 3, 5)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_Min(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid", `$validator.min(10, 5)`, true},
		{"exact", `$validator.min(5, 5)`, true},
		{"invalid", `$validator.min(3, 5)`, false},
		{"negative valid", `$validator.min(-5, -10)`, true},
		{"float valid", `$validator.min(5.5, 5.0)`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_Max(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid", `$validator.max(3, 5)`, true},
		{"exact", `$validator.max(5, 5)`, true},
		{"invalid", `$validator.max(10, 5)`, false},
		{"negative valid", `$validator.max(-10, -5)`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_Between(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid", `$validator.between(7, 5, 10)`, true},
		{"exact min", `$validator.between(5, 5, 10)`, true},
		{"exact max", `$validator.between(10, 5, 10)`, true},
		{"too small", `$validator.between(3, 5, 10)`, false},
		{"too big", `$validator.between(15, 5, 10)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_Matches(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"simple match", `$validator.matches("hello", "^hello$")`, true},
		{"partial match", `$validator.matches("hello world", "hello")`, true},
		{"number pattern", `$validator.matches("abc123", "^[a-z]+[0-9]+$")`, true},
		{"no match", `$validator.matches("hello", "^world$")`, false},
		{"email pattern", `$validator.matches("test@example.com", "^[a-z]+@[a-z]+\\.[a-z]+$")`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_Contains(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"contains", `$validator.contains("hello world", "world")`, true},
		{"start", `$validator.contains("hello world", "hello")`, true},
		{"not found", `$validator.contains("hello world", "foo")`, false},
		{"empty substring", `$validator.contains("hello", "")`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_StartsWith(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"starts with", `$validator.startsWith("hello world", "hello")`, true},
		{"does not start", `$validator.startsWith("hello world", "world")`, false},
		{"empty prefix", `$validator.startsWith("hello", "")`, true},
		{"exact match", `$validator.startsWith("hello", "hello")`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_EndsWith(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"ends with", `$validator.endsWith("hello world", "world")`, true},
		{"does not end", `$validator.endsWith("hello world", "hello")`, false},
		{"empty suffix", `$validator.endsWith("hello", "")`, true},
		{"exact match", `$validator.endsWith("hello", "hello")`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_OneOf(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"string found", `$validator.oneOf("apple", ["apple", "banana", "cherry"])`, true},
		{"string not found", `$validator.oneOf("grape", ["apple", "banana", "cherry"])`, false},
		{"number found", `$validator.oneOf(2, [1, 2, 3])`, true},
		{"number not found", `$validator.oneOf(5, [1, 2, 3])`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_NotEmpty(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"non-empty string", `$validator.notEmpty("hello")`, true},
		{"empty string", `$validator.notEmpty("")`, false},
		{"non-empty array", `$validator.notEmpty([1, 2, 3])`, true},
		{"empty array", `$validator.notEmpty([])`, false},
		{"non-empty object", `$validator.notEmpty({a: 1})`, true},
		{"empty object", `$validator.notEmpty({})`, false},
		{"null", `$validator.notEmpty(null)`, false},
		{"number", `$validator.notEmpty(42)`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_Required(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"string", `$validator.required("hello")`, true},
		{"number", `$validator.required(42)`, true},
		{"empty string", `$validator.required("")`, true}, // empty string is not null
		{"zero", `$validator.required(0)`, true},
		{"false", `$validator.required(false)`, true},
		{"null", `$validator.required(null)`, false},
		{"undefined", `$validator.required(undefined)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_IsValid(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid email", `$validator.isValid("test@example.com", "email")`, true},
		{"invalid email", `$validator.isValid("invalid", "email")`, false},
		{"valid required", `$validator.isValid("hello", "required")`, true},
		{"combined rules", `$validator.isValid("test@example.com", "required,email")`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.ToBoolean() != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.code, result.ToBoolean(), tt.expected)
			}
		})
	}
}

func TestJS_Validator_Var(t *testing.T) {
	h := NewJSTestHelper(t)

	code := `
		var result = $validator.var("test@example.com", "required,email");
		result.valid;
	`
	result := h.MustRun(t, code)
	if !result.ToBoolean() {
		t.Error("$validator.var should return valid=true for valid email")
	}

	code = `
		var result = $validator.var("invalid", "email");
		result.valid;
	`
	result = h.MustRun(t, code)
	if result.ToBoolean() {
		t.Error("$validator.var should return valid=false for invalid email")
	}

	code = `
		var result = $validator.var("invalid", "email");
		result.errors.length > 0;
	`
	result = h.MustRun(t, code)
	if !result.ToBoolean() {
		t.Error("$validator.var should return errors for invalid input")
	}
}

func TestJS_Validator_Struct(t *testing.T) {
	h := NewJSTestHelper(t)

	// Valid struct
	code := `
		var result = $validator.struct(
			{email: "test@example.com", age: 25},
			{email: "required,email", age: "required,min=18"}
		);
		result.valid;
	`
	result := h.MustRun(t, code)
	if !result.ToBoolean() {
		t.Error("$validator.struct should return valid=true for valid data")
	}

	// Invalid email
	code = `
		var result = $validator.struct(
			{email: "invalid", age: 25},
			{email: "required,email", age: "required,min=18"}
		);
		result.valid;
	`
	result = h.MustRun(t, code)
	if result.ToBoolean() {
		t.Error("$validator.struct should return valid=false for invalid email")
	}

	// Missing required field
	code = `
		var result = $validator.struct(
			{age: 25},
			{email: "required,email", age: "required,min=18"}
		);
		result.valid;
	`
	result = h.MustRun(t, code)
	if result.ToBoolean() {
		t.Error("$validator.struct should return valid=false for missing required field")
	}

	// Check error details
	code = `
		var result = $validator.struct(
			{email: "invalid"},
			{email: "required,email"}
		);
		result.errors[0].field;
	`
	result = h.MustRun(t, code)
	if result.String() != "email" {
		t.Errorf("Error field should be 'email', got %s", result.String())
	}

	// Age below minimum
	code = `
		var result = $validator.struct(
			{email: "test@example.com", age: 15},
			{email: "required,email", age: "required,min=18"}
		);
		result.valid;
	`
	result = h.MustRun(t, code)
	if result.ToBoolean() {
		t.Error("$validator.struct should return valid=false for age below minimum")
	}
}

func TestJS_Validator_ComplexScenario(t *testing.T) {
	h := NewJSTestHelper(t)

	// Real-world validation scenario
	code := `
		var userData = {
			username: "johndoe123",
			email: "john@example.com",
			password: "secret123",
			age: 25,
			website: "https://example.com"
		};

		var rules = {
			username: "required,alphanum,min=3,max=20",
			email: "required,email",
			password: "required,min=8",
			age: "required,min=18",
			website: "url"
		};

		var result = $validator.struct(userData, rules);
		result.valid;
	`
	result := h.MustRun(t, code)
	if !result.ToBoolean() {
		t.Error("Complex validation should pass for valid data")
	}

	// Test with invalid data
	code = `
		var userData = {
			username: "j",              // too short
			email: "invalid-email",     // invalid
			password: "short",          // too short
			age: 15,                    // below 18
			website: "not-a-url"        // invalid
		};

		var rules = {
			username: "required,min=3,max=20",
			email: "required,email",
			password: "required,min=8",
			age: "required,min=18",
			website: "url"
		};

		var result = $validator.struct(userData, rules);
		result.errors.length;
	`
	result = h.MustRun(t, code)
	errorCount := result.ToInteger()
	if errorCount < 4 {
		t.Errorf("Expected at least 4 errors, got %d", errorCount)
	}
}
