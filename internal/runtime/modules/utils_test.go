package modules

import (
	"strings"
	"testing"
	"time"
)

func TestUtilsModule_Sleep(t *testing.T) {
	utils := NewUtilsModule()

	start := time.Now()
	utils.Sleep(50)
	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Errorf("Sleep() should wait at least 50ms, got %v", elapsed)
	}
}

func TestUtilsModule_Random(t *testing.T) {
	utils := NewUtilsModule()

	for i := 0; i < 100; i++ {
		r := utils.Random()
		if r < 0 || r >= 1 {
			t.Errorf("Random() = %v, want value in [0, 1)", r)
		}
	}
}

func TestUtilsModule_RandomInt(t *testing.T) {
	utils := NewUtilsModule()

	tests := []struct {
		name     string
		min, max int
	}{
		{"positive range", 1, 10},
		{"zero to positive", 0, 100},
		{"negative range", -10, -1},
		{"cross zero", -5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				got := utils.RandomInt(tt.min, tt.max)
				if got < tt.min || got >= tt.max {
					t.Errorf("RandomInt(%d, %d) = %d, want value in [%d, %d)", tt.min, tt.max, got, tt.min, tt.max)
				}
			}
		})
	}
}

func TestUtilsModule_RandomInt_EqualMinMax(t *testing.T) {
	utils := NewUtilsModule()

	got := utils.RandomInt(5, 5)
	if got != 5 {
		t.Errorf("RandomInt(5, 5) = %d, want 5", got)
	}
}

func TestUtilsModule_RandomInt_MinGreaterThanMax(t *testing.T) {
	utils := NewUtilsModule()

	got := utils.RandomInt(10, 5)
	if got != 10 {
		t.Errorf("RandomInt(10, 5) = %d, want 10 (min)", got)
	}
}

func TestUtilsModule_UUID(t *testing.T) {
	utils := NewUtilsModule()

	uuid1 := utils.UUID()
	uuid2 := utils.UUID()

	if len(uuid1) != 36 {
		t.Errorf("UUID() = %q, want 36 character string", uuid1)
	}

	if uuid1 == uuid2 {
		t.Errorf("UUID() should return unique values, got same: %s", uuid1)
	}

	// Check UUID format (8-4-4-4-12)
	parts := strings.Split(uuid1, "-")
	if len(parts) != 5 {
		t.Errorf("UUID() format should be 8-4-4-4-12, got %q", uuid1)
	}
}

func TestUtilsModule_Slugify(t *testing.T) {
	utils := NewUtilsModule()

	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"This is a TEST", "this-is-a-test"},
		{"Special @#$% Characters!", "special-characters"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"---Leading and trailing---", "leading-and-trailing"},
		{"123 Numbers", "123-numbers"},
		{"", ""},
		{"Already-a-slug", "already-a-slug"},
		{"UPPERCASE", "uppercase"},
		{"mixed CASE with 123", "mixed-case-with-123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := utils.Slugify(tt.input)
			if got != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestUtilsModule_Truncate(t *testing.T) {
	utils := NewUtilsModule()

	tests := []struct {
		text     string
		length   int
		expected string
	}{
		{"Hello World", 5, "Hello..."},
		{"Short", 10, "Short"},
		{"Exact", 5, "Exact"},
		{"", 5, ""},
		{"Test", 0, "..."},
		{"LongString", 4, "Long..."},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := utils.Truncate(tt.text, tt.length)
			if got != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.text, tt.length, got, tt.expected)
			}
		})
	}
}

func TestUtilsModule_Capitalize(t *testing.T) {
	utils := NewUtilsModule()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"HELLO", "HELLO"},
		{"hello world", "Hello world"},
		{"", ""},
		{"a", "A"},
		{"123abc", "123abc"},
		{"тест", "Тест"}, // Unicode support
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := utils.Capitalize(tt.input)
			if got != tt.expected {
				t.Errorf("Capitalize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestUtilsModule_RegexMatch(t *testing.T) {
	utils := NewUtilsModule()

	tests := []struct {
		text     string
		pattern  string
		expected []string
	}{
		{"abc123def456", `\d+`, []string{"123", "456"}},
		{"hello world", `\w+`, []string{"hello", "world"}},
		{"no match here", `\d+`, []string{}},
		{"test@email.com", `\w+@\w+\.\w+`, []string{"test@email.com"}},
		{"", `\w+`, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := utils.RegexMatch(tt.text, tt.pattern)
			if len(got) != len(tt.expected) {
				t.Errorf("RegexMatch(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("RegexMatch(%q, %q)[%d] = %q, want %q", tt.text, tt.pattern, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestUtilsModule_RegexMatch_InvalidPattern(t *testing.T) {
	utils := NewUtilsModule()

	got := utils.RegexMatch("test", "[invalid")
	if len(got) != 0 {
		t.Errorf("RegexMatch with invalid pattern should return empty slice, got %v", got)
	}
}

func TestUtilsModule_RegexReplace(t *testing.T) {
	utils := NewUtilsModule()

	tests := []struct {
		text        string
		pattern     string
		replacement string
		expected    string
	}{
		{"hello123world", `\d+`, "-", "hello-world"},
		{"abc ABC", `[a-z]+`, "X", "X ABC"},
		{"test", `\d+`, "X", "test"},
		{"a  b   c", `\s+`, " ", "a b c"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := utils.RegexReplace(tt.text, tt.pattern, tt.replacement)
			if got != tt.expected {
				t.Errorf("RegexReplace(%q, %q, %q) = %q, want %q", tt.text, tt.pattern, tt.replacement, got, tt.expected)
			}
		})
	}
}

func TestUtilsModule_RegexReplace_InvalidPattern(t *testing.T) {
	utils := NewUtilsModule()

	got := utils.RegexReplace("test", "[invalid", "X")
	if got != "test" {
		t.Errorf("RegexReplace with invalid pattern should return original text, got %q", got)
	}
}

func TestUtilsModule_FormatDate(t *testing.T) {
	utils := NewUtilsModule()

	// Use current time to avoid timezone issues
	now := time.Now()
	timestamp := now.UnixMilli()

	tests := []struct {
		format   string
		expected string
	}{
		{"YYYY-MM-DD", now.Format("2006-01-02")},
		{"DD/MM/YYYY", now.Format("02/01/2006")},
		{"YYYY", now.Format("2006")},
		{"HH:mm:ss", now.Format("15:04:05")},
		{"YYYY-MM-DD HH:mm:ss", now.Format("2006-01-02 15:04:05")},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got := utils.FormatDate(timestamp, tt.format)
			if got != tt.expected {
				t.Errorf("FormatDate(%d, %q) = %q, want %q", timestamp, tt.format, got, tt.expected)
			}
		})
	}
}

func TestUtilsModule_ParseDate(t *testing.T) {
	utils := NewUtilsModule()

	tests := []struct {
		text   string
		format string
	}{
		{"2024-01-15", "YYYY-MM-DD"},
		{"15/01/2024", "DD/MM/YYYY"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := utils.ParseDate(tt.text, tt.format)
			if got == 0 {
				t.Errorf("ParseDate(%q, %q) = 0, want non-zero timestamp", tt.text, tt.format)
			}
		})
	}
}

func TestUtilsModule_ParseDate_Invalid(t *testing.T) {
	utils := NewUtilsModule()

	got := utils.ParseDate("invalid-date", "YYYY-MM-DD")
	if got != 0 {
		t.Errorf("ParseDate with invalid date should return 0, got %d", got)
	}
}

func TestUtilsModule_Timestamp(t *testing.T) {
	utils := NewUtilsModule()

	before := time.Now().UnixMilli()
	got := utils.Timestamp()
	after := time.Now().UnixMilli()

	if got < before || got > after {
		t.Errorf("Timestamp() = %d, should be between %d and %d", got, before, after)
	}
}
