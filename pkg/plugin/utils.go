package plugin

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"
)

// IsFilePath checks if the string looks like a file path
func IsFilePath(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Absolute paths or relative paths starting with . or /
	if s[0] == '/' || s[0] == '.' {
		return true
	}
	// Check for file extension and not URL
	if filepath.Ext(s) != "" && !IsURL(s) {
		return true
	}
	return false
}

// IsURL checks if the string is an HTTP(S) URL
func IsURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}

// IsBase64 checks if the string appears to be base64 encoded data
func IsBase64(s string) bool {
	if len(s) < 100 {
		return false
	}
	// Try to decode first 100 chars to check if valid base64
	_, err := base64.StdEncoding.DecodeString(s[:100])
	return err == nil
}

// ResolvePath resolves a path relative to a base storage path.
// It includes path traversal protection.
func ResolvePath(basePath, path string) (string, error) {
	if basePath == "" || filepath.IsAbs(path) || !IsFilePath(path) {
		return path, nil
	}

	cleanBase := filepath.Clean(basePath)

	// If path already starts with basePath, don't add it again
	if strings.HasPrefix(path, cleanBase+"/") || strings.HasPrefix(path, cleanBase+string(filepath.Separator)) {
		return filepath.Clean(path), nil
	}

	// Join and clean the path
	resolved := filepath.Join(basePath, path)
	resolved = filepath.Clean(resolved)

	// Ensure the resolved path is still within the base path (prevent path traversal)
	if !strings.HasPrefix(resolved, cleanBase) {
		return "", fmt.Errorf("path traversal detected: %s", path)
	}

	return resolved, nil
}

// MustResolvePath resolves a path and returns the original if there's an error.
// Use this for backward compatibility where errors were ignored.
func MustResolvePath(basePath, path string) string {
	resolved, err := ResolvePath(basePath, path)
	if err != nil {
		return path
	}
	return resolved
}
