package modules

import (
	"testing"
)

func TestEncodingModule_Base64Encode(t *testing.T) {
	enc := NewEncodingModule()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "aGVsbG8="},
		{"Hello World", "SGVsbG8gV29ybGQ="},
		{"", ""},
		{"test123", "dGVzdDEyMw=="},
		{"a", "YQ=="},
		{"user:password", "dXNlcjpwYXNzd29yZA=="},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := enc.Base64Encode(tt.input)
			if got != tt.expected {
				t.Errorf("Base64Encode(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEncodingModule_Base64Decode(t *testing.T) {
	enc := NewEncodingModule()

	tests := []struct {
		input    string
		expected string
	}{
		{"aGVsbG8=", "hello"},
		{"SGVsbG8gV29ybGQ=", "Hello World"},
		{"", ""},
		{"dGVzdDEyMw==", "test123"},
		{"YQ==", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := enc.Base64Decode(tt.input)
			if got != tt.expected {
				t.Errorf("Base64Decode(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEncodingModule_Base64Decode_Invalid(t *testing.T) {
	enc := NewEncodingModule()

	tests := []string{
		"!!!invalid!!!",
		"not-valid-base64",
		"===",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got := enc.Base64Decode(input)
			if got != "" {
				t.Errorf("Base64Decode(%q) = %q, want empty string for invalid input", input, got)
			}
		})
	}
}

func TestEncodingModule_Base64_Roundtrip(t *testing.T) {
	enc := NewEncodingModule()

	inputs := []string{
		"hello world",
		"special chars: !@#$%^&*()",
		"unicode: 日本語",
		"multi\nline\ntext",
		"",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			encoded := enc.Base64Encode(input)
			decoded := enc.Base64Decode(encoded)
			if decoded != input {
				t.Errorf("Base64 roundtrip failed for %q: got %q", input, decoded)
			}
		})
	}
}

func TestEncodingModule_JSONParse(t *testing.T) {
	enc := NewEncodingModule()

	tests := []struct {
		input     string
		checkFunc func(interface{}) bool
	}{
		{
			`{"name":"test","value":123}`,
			func(v interface{}) bool {
				m, ok := v.(map[string]interface{})
				if !ok {
					return false
				}
				return m["name"] == "test" && m["value"] == float64(123)
			},
		},
		{
			`[1, 2, 3]`,
			func(v interface{}) bool {
				arr, ok := v.([]interface{})
				if !ok || len(arr) != 3 {
					return false
				}
				return arr[0] == float64(1) && arr[1] == float64(2) && arr[2] == float64(3)
			},
		},
		{
			`"hello"`,
			func(v interface{}) bool {
				return v == "hello"
			},
		},
		{
			`123`,
			func(v interface{}) bool {
				return v == float64(123)
			},
		},
		{
			`true`,
			func(v interface{}) bool {
				return v == true
			},
		},
		{
			`null`,
			func(v interface{}) bool {
				return v == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := enc.JSONParse(tt.input)
			if !tt.checkFunc(got) {
				t.Errorf("JSONParse(%q) = %v, validation failed", tt.input, got)
			}
		})
	}
}

func TestEncodingModule_JSONParse_Invalid(t *testing.T) {
	enc := NewEncodingModule()

	tests := []string{
		"invalid json",
		"{malformed",
		"",
		"undefined",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got := enc.JSONParse(input)
			if got != nil {
				t.Errorf("JSONParse(%q) = %v, want nil for invalid JSON", input, got)
			}
		})
	}
}

func TestEncodingModule_JSONStringify(t *testing.T) {
	enc := NewEncodingModule()

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			"map",
			map[string]interface{}{"name": "test", "value": 123},
			`{"name":"test","value":123}`,
		},
		{
			"array",
			[]interface{}{1, 2, 3},
			`[1,2,3]`,
		},
		{
			"string",
			"hello",
			`"hello"`,
		},
		{
			"number",
			123,
			`123`,
		},
		{
			"bool",
			true,
			`true`,
		},
		{
			"nil",
			nil,
			`null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enc.JSONStringify(tt.input)
			if got != tt.expected {
				t.Errorf("JSONStringify(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEncodingModule_JSONStringify_Invalid(t *testing.T) {
	enc := NewEncodingModule()

	// Channels cannot be marshaled to JSON
	ch := make(chan int)
	got := enc.JSONStringify(ch)
	if got != "" {
		t.Errorf("JSONStringify(channel) = %q, want empty string", got)
	}
}

func TestEncodingModule_JSON_Roundtrip(t *testing.T) {
	enc := NewEncodingModule()

	inputs := []interface{}{
		map[string]interface{}{"key": "value", "num": float64(42)},
		[]interface{}{"a", "b", "c"},
		"simple string",
		float64(123.456),
		true,
	}

	for _, input := range inputs {
		t.Run("", func(t *testing.T) {
			json := enc.JSONStringify(input)
			parsed := enc.JSONParse(json)
			reparsed := enc.JSONStringify(parsed)
			if json != reparsed {
				t.Errorf("JSON roundtrip failed: %q != %q", json, reparsed)
			}
		})
	}
}

func TestEncodingModule_URLEncode(t *testing.T) {
	enc := NewEncodingModule()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello+world"},
		{"test=value", "test%3Dvalue"},
		{"a&b", "a%26b"},
		{"special chars: !@#$", "special+chars%3A+%21%40%23%24"},
		{"", ""},
		{"no-encoding-needed", "no-encoding-needed"},
		{"path/to/file", "path%2Fto%2Ffile"},
		{"query?param", "query%3Fparam"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := enc.URLEncode(tt.input)
			if got != tt.expected {
				t.Errorf("URLEncode(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEncodingModule_URLDecode(t *testing.T) {
	enc := NewEncodingModule()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello+world", "hello world"},
		{"test%3Dvalue", "test=value"},
		{"a%26b", "a&b"},
		{"", ""},
		{"no-encoding-needed", "no-encoding-needed"},
		{"path%2Fto%2Ffile", "path/to/file"},
		{"%E4%B8%AD%E6%96%87", "中文"}, // Unicode
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := enc.URLDecode(tt.input)
			if got != tt.expected {
				t.Errorf("URLDecode(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEncodingModule_URLDecode_Invalid(t *testing.T) {
	enc := NewEncodingModule()

	// Invalid percent encoding
	got := enc.URLDecode("%ZZ")
	if got != "" {
		t.Errorf("URLDecode(\"%%ZZ\") = %q, want empty string for invalid encoding", got)
	}
}

func TestEncodingModule_URL_Roundtrip(t *testing.T) {
	enc := NewEncodingModule()

	inputs := []string{
		"hello world",
		"test=value&other=123",
		"path/to/resource?query=1",
		"unicode: 日本語",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			encoded := enc.URLEncode(input)
			decoded := enc.URLDecode(encoded)
			if decoded != input {
				t.Errorf("URL roundtrip failed for %q: got %q", input, decoded)
			}
		})
	}
}
