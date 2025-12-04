package modules

import (
	"strings"
	"testing"
)

func TestCryptoModule_MD5(t *testing.T) {
	crypto := NewCryptoModule()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"world", "7d793037a0760186574b0282f2f435e7"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"test123", "cc03e747a6afbbcbf8be7668acfebee5"},
		{"Hello World", "b10a8db164e0754105b7a99be72e3fe5"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := crypto.MD5(tt.input)
			if got != tt.expected {
				t.Errorf("MD5(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCryptoModule_MD5_Deterministic(t *testing.T) {
	crypto := NewCryptoModule()

	input := "test-deterministic"
	result1 := crypto.MD5(input)
	result2 := crypto.MD5(input)

	if result1 != result2 {
		t.Errorf("MD5 should be deterministic: %q != %q", result1, result2)
	}
}

func TestCryptoModule_SHA256(t *testing.T) {
	crypto := NewCryptoModule()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"world", "486ea46224d1bb4fb680f34f7c9ad96a8f24ec88be73ea8e5a6c65260e9cb8a7"},
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"test123", "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := crypto.SHA256(tt.input)
			if got != tt.expected {
				t.Errorf("SHA256(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCryptoModule_SHA256_Deterministic(t *testing.T) {
	crypto := NewCryptoModule()

	input := "test-deterministic"
	result1 := crypto.SHA256(input)
	result2 := crypto.SHA256(input)

	if result1 != result2 {
		t.Errorf("SHA256 should be deterministic: %q != %q", result1, result2)
	}
}

func TestCryptoModule_SHA256_Length(t *testing.T) {
	crypto := NewCryptoModule()

	got := crypto.SHA256("any input")
	if len(got) != 64 {
		t.Errorf("SHA256 output should be 64 hex chars, got %d", len(got))
	}
}

func TestCryptoModule_MD5_Length(t *testing.T) {
	crypto := NewCryptoModule()

	got := crypto.MD5("any input")
	if len(got) != 32 {
		t.Errorf("MD5 output should be 32 hex chars, got %d", len(got))
	}
}

func TestCryptoModule_RandomBytes(t *testing.T) {
	crypto := NewCryptoModule()

	tests := []struct {
		length int
	}{
		{8},
		{16},
		{32},
		{64},
		{1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := crypto.RandomBytes(tt.length)
			// RandomBytes returns hex string, so length is doubled
			expectedLen := tt.length * 2
			if len(got) != expectedLen {
				t.Errorf("RandomBytes(%d) length = %d, want %d", tt.length, len(got), expectedLen)
			}
		})
	}
}

func TestCryptoModule_RandomBytes_Unique(t *testing.T) {
	crypto := NewCryptoModule()

	results := make(map[string]bool)
	for i := 0; i < 100; i++ {
		got := crypto.RandomBytes(16)
		if results[got] {
			t.Errorf("RandomBytes(16) produced duplicate value: %s", got)
		}
		results[got] = true
	}
}

func TestCryptoModule_RandomBytes_HexChars(t *testing.T) {
	crypto := NewCryptoModule()

	got := crypto.RandomBytes(32)
	for _, c := range got {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("RandomBytes should only contain hex chars, found %q in %s", c, got)
		}
	}
}

func TestCryptoModule_RandomBytes_ZeroLength(t *testing.T) {
	crypto := NewCryptoModule()

	got := crypto.RandomBytes(0)
	if got != "" {
		t.Errorf("RandomBytes(0) = %q, want empty string", got)
	}
}
