package tests

import (
	"testing"
	"time"

	"m3m/internal/runtime/modules"

	"github.com/dop251/goja"
)

// ============== CRYPTO MODULE TESTS ==============

func TestJS_Crypto_MD5(t *testing.T) {
	h := NewJSTestHelper(t)
	cryptoModule := modules.NewCryptoModule()

	tests := []struct {
		name  string
		input string
	}{
		{"simple string", "hello"},
		{"empty string", ""},
		{"unicode", "привет"},
		{"special chars", "<script>alert('xss')</script>"},
		{"numbers", "12345"},
		{"whitespace", "  spaces  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := cryptoModule.MD5(tt.input)
			result := h.MustRun(t, `crypto.md5("`+tt.input+`")`)
			if result.String() != expected {
				t.Errorf("MD5(%q): JS=%s, Go=%s", tt.input, result.String(), expected)
			}
		})
	}
}

func TestJS_Crypto_UnexpectedTypes(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name string
		code string
	}{
		{"number to md5", `crypto.md5(12345)`},
		{"boolean to md5", `crypto.md5(true)`},
		{"null to md5", `crypto.md5(null)`},
		{"undefined to md5", `crypto.md5(undefined)`},
		{"object to md5", `crypto.md5({foo: "bar"})`},
		{"array to md5", `crypto.md5([1, 2, 3])`},
		{"empty call", `crypto.md5()`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			if err != nil {
				t.Logf("Got error (acceptable): %v", err)
				return
			}
			if result.String() == "" && tt.name != "empty call" {
				t.Logf("Got empty result (acceptable for edge case)")
			}
		})
	}
}

func TestJS_Crypto_SHA256(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `crypto.sha256("hello")`)
	expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if result.String() != expected {
		t.Errorf("Expected %s, got %s", expected, result.String())
	}
}

func TestJS_Crypto_RandomBytes(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `crypto.randomBytes(16)`)
	if len(result.String()) != 32 {
		t.Errorf("Expected 32 chars, got %d", len(result.String()))
	}

	result1 := h.MustRun(t, `crypto.randomBytes(16)`)
	result2 := h.MustRun(t, `crypto.randomBytes(16)`)
	if result1.String() == result2.String() {
		t.Error("RandomBytes should produce unique values")
	}

	result = h.MustRun(t, `crypto.randomBytes(0)`)
	if result.String() != "" {
		t.Errorf("Zero length should return empty string")
	}

	result = h.MustRun(t, `crypto.randomBytes(-5)`)
	if result.String() != "" {
		t.Errorf("Negative length should return empty string")
	}
}

// ============== ENCODING MODULE TESTS ==============

func TestJS_Encoding_Base64(t *testing.T) {
	h := NewJSTestHelper(t)

	code := `
		var encoded = encoding.base64Encode("Hello, World!");
		var decoded = encoding.base64Decode(encoded);
		decoded;
	`
	result := h.MustRun(t, code)
	if result.String() != "Hello, World!" {
		t.Errorf("Base64 roundtrip failed: got %s", result.String())
	}
}

func TestJS_Encoding_Base64_InvalidInput(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `encoding.base64Decode("not-valid-base64!!!")`)
	if result.String() != "" {
		t.Errorf("Invalid base64 should return empty string, got %s", result.String())
	}
}

func TestJS_Encoding_JSON(t *testing.T) {
	h := NewJSTestHelper(t)

	code := `
		var obj = {name: "test", value: 42, nested: {foo: "bar"}};
		var str = encoding.jsonStringify(obj);
		var parsed = encoding.jsonParse(str);
		parsed.name + "-" + parsed.value + "-" + parsed.nested.foo;
	`
	result := h.MustRun(t, code)
	if result.String() != "test-42-bar" {
		t.Errorf("JSON roundtrip failed: got %s", result.String())
	}
}

func TestJS_Encoding_JSON_InvalidInput(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `encoding.jsonParse("not valid json {")`)
	if !goja.IsNull(result) && !goja.IsUndefined(result) && result.Export() != nil {
		t.Errorf("Invalid JSON should return null, got %v", result.Export())
	}
}

func TestJS_Encoding_URL(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `encoding.urlEncode("hello world & foo=bar")`)
	if result.String() != "hello+world+%26+foo%3Dbar" {
		t.Errorf("URL encode failed: got %s", result.String())
	}

	code := `encoding.urlDecode(encoding.urlEncode("тест?foo=bar&baz=1"))`
	result = h.MustRun(t, code)
	if result.String() != "тест?foo=bar&baz=1" {
		t.Errorf("URL roundtrip failed: got %s", result.String())
	}
}

// ============== UTILS MODULE TESTS ==============

func TestJS_Utils_UUID(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `utils.uuid()`)
	uuid := result.String()

	if len(uuid) != 36 {
		t.Errorf("UUID should be 36 chars, got %d: %s", len(uuid), uuid)
	}
	if uuid[14] != '4' {
		t.Errorf("UUID v4 should have '4' at position 14, got %c", uuid[14])
	}

	uuid1 := h.MustRun(t, `utils.uuid()`).String()
	uuid2 := h.MustRun(t, `utils.uuid()`).String()
	if uuid1 == uuid2 {
		t.Error("UUIDs should be unique")
	}
}

func TestJS_Utils_Slugify(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"  Spaces  Around  ", "spaces-around"},
		{"Special!@#$%Chars", "specialchars"},
		{"Mixed 123 Numbers", "mixed-123-numbers"},
		{"already-slugified", "already-slugified"},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.MustRun(t, `utils.slugify("`+tt.input+`")`)
			if result.String() != tt.expected {
				t.Errorf("slugify(%q) = %q, expected %q", tt.input, result.String(), tt.expected)
			}
		})
	}
}

func TestJS_Utils_RandomInt(t *testing.T) {
	h := NewJSTestHelper(t)

	for i := 0; i < 100; i++ {
		result := h.MustRun(t, `utils.randomInt(5, 10)`)
		val := int(result.ToInteger())
		if val < 5 || val >= 10 {
			t.Errorf("randomInt(5, 10) = %d, expected [5, 10)", val)
		}
	}

	result := h.MustRun(t, `utils.randomInt(5, 5)`)
	if result.ToInteger() != 5 {
		t.Errorf("randomInt(5, 5) should return 5, got %d", result.ToInteger())
	}

	result = h.MustRun(t, `utils.randomInt(10, 5)`)
	if result.ToInteger() != 10 {
		t.Errorf("randomInt(10, 5) should return min (10), got %d", result.ToInteger())
	}
}

func TestJS_Utils_Truncate(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"short text", `utils.truncate("hello", 10)`, "hello"},
		{"exact length", `utils.truncate("hello", 5)`, "hello"},
		{"truncated", `utils.truncate("hello world", 5)`, "hello..."},
		{"zero length", `utils.truncate("hello", 0)`, "..."},
		{"negative length", `utils.truncate("hello", -5)`, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.MustRun(t, tt.code)
			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

func TestJS_Utils_Timestamp(t *testing.T) {
	h := NewJSTestHelper(t)

	before := time.Now().UnixMilli()
	result := h.MustRun(t, `utils.timestamp()`)
	after := time.Now().UnixMilli()

	ts := result.ToInteger()
	if ts < before || ts > after {
		t.Errorf("Timestamp %d should be between %d and %d", ts, before, after)
	}
}

// ============== ENV MODULE TESTS ==============

func TestJS_Env_Get(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `env.get("TEST_VAR")`)
	if result.String() != "test_value" {
		t.Errorf("Expected 'test_value', got %s", result.String())
	}

	result = h.MustRun(t, `env.get("NOT_EXISTS")`)
	if !goja.IsUndefined(result) && !goja.IsNull(result) && result.Export() != nil {
		t.Errorf("Expected undefined/null for non-existing var, got %v", result.Export())
	}
}

func TestJS_Env_Has(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `env.has("TEST_VAR")`)
	if !result.ToBoolean() {
		t.Error("has(TEST_VAR) should return true")
	}

	result = h.MustRun(t, `env.has("NOT_EXISTS")`)
	if result.ToBoolean() {
		t.Error("has(NOT_EXISTS) should return false")
	}

	result = h.MustRun(t, `env.has("EMPTY_VAR")`)
	if !result.ToBoolean() {
		t.Error("has(EMPTY_VAR) should return true")
	}
}

func TestJS_Env_TypedGetters(t *testing.T) {
	h := NewJSTestHelper(t)

	result := h.MustRun(t, `env.getInt("NUMBER_VAR", 0)`)
	if result.ToInteger() != 42 {
		t.Errorf("getInt should return 42, got %d", result.ToInteger())
	}

	result = h.MustRun(t, `env.getInt("NOT_EXISTS", 99)`)
	if result.ToInteger() != 99 {
		t.Errorf("getInt should return default 99, got %d", result.ToInteger())
	}

	result = h.MustRun(t, `env.getFloat("FLOAT_VAR", 0)`)
	if result.ToFloat() != 3.14 {
		t.Errorf("getFloat should return 3.14, got %f", result.ToFloat())
	}

	result = h.MustRun(t, `env.getBool("BOOL_VAR", false)`)
	if !result.ToBoolean() {
		t.Error("getBool should return true")
	}
}

// ============== EDGE CASES ==============

func TestJS_EdgeCases_EmptyStrings(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name string
		code string
	}{
		{"crypto md5 empty", `crypto.md5("")`},
		{"crypto sha256 empty", `crypto.sha256("")`},
		{"encoding base64 empty", `encoding.base64Encode("")`},
		{"encoding base64 decode empty", `encoding.base64Decode("")`},
		{"utils slugify empty", `utils.slugify("")`},
		{"utils truncate empty", `utils.truncate("", 10)`},
		{"utils capitalize empty", `utils.capitalize("")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			if err != nil {
				t.Errorf("Should handle empty string: %v", err)
				return
			}
			t.Logf("Result: %v", result.Export())
		})
	}
}

func TestJS_EdgeCases_UnicodeHandling(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name string
		code string
	}{
		{"crypto md5 unicode", `crypto.md5("привет мир")`},
		{"crypto sha256 emoji", `crypto.sha256("Hello 🌍")`},
		{"encoding base64 unicode", `encoding.base64Encode("日本語")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.Run(tt.code)
			if err != nil {
				t.Errorf("Should handle unicode: %v", err)
				return
			}
			t.Logf("Result: %v", result.Export())
		})
	}
}

func TestJS_EdgeCases_LargeInputs(t *testing.T) {
	h := NewJSTestHelper(t)

	code := `
		var large = "";
		for (var i = 0; i < 10000; i++) {
			large += "x";
		}
		crypto.md5(large);
	`

	result := h.MustRun(t, code)
	if len(result.String()) != 32 {
		t.Errorf("Expected 32 char hash, got %d", len(result.String()))
	}
}
