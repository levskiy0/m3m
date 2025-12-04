package modules

import (
	"os"
	"path/filepath"
	"testing"

	"m3m/internal/config"
	"m3m/internal/service"
)

func newTestStorageService(t *testing.T) (*service.StorageService, string) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			Path: tmpDir,
		},
		Logging: config.LoggingConfig{
			Path: filepath.Join(tmpDir, "logs"),
		},
	}
	return service.NewStorageService(cfg), tmpDir
}

func createTestProjectDir(t *testing.T, tmpDir, projectID string) {
	projectPath := filepath.Join(tmpDir, projectID, "storage")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}
}

// Minimal valid PNG (1x1 transparent pixel)
var minimalPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

func TestImageModule_Info_NonExistentFile(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	result := module.Info("nonexistent.png")
	if result != nil {
		t.Error("Expected nil for non-existent file")
	}
}

func TestImageModule_Info_ValidPNG(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	// Write test PNG file
	if err := storage.Write(projectID, "test.png", minimalPNG); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result := module.Info("test.png")
	if result == nil {
		t.Fatal("Expected info for valid PNG")
	}
	if result.Width != 1 || result.Height != 1 {
		t.Errorf("Expected 1x1, got %dx%d", result.Width, result.Height)
	}
	if result.Format != "png" {
		t.Errorf("Expected png format, got %s", result.Format)
	}
}

func TestImageModule_Info_InvalidImage(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	// Write invalid image data
	if err := storage.Write(projectID, "invalid.png", []byte("not an image")); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result := module.Info("invalid.png")
	if result != nil {
		t.Error("Expected nil for invalid image data")
	}
}

func TestImageModule_Resize_NonExistentFile(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	result := module.Resize("nonexistent.png", "output.png", 100, 100)
	if result {
		t.Error("Expected false for non-existent source file")
	}
}

func TestImageModule_Resize_ValidPNG(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	if err := storage.Write(projectID, "source.png", minimalPNG); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result := module.Resize("source.png", "output.png", 50, 50)
	if !result {
		t.Error("Expected true for valid resize operation")
	}

	// Verify output was created
	info := module.Info("output.png")
	if info == nil {
		t.Fatal("Expected output file to be created")
	}
	if info.Width != 50 || info.Height != 50 {
		t.Errorf("Expected 50x50, got %dx%d", info.Width, info.Height)
	}
}

func TestImageModule_ResizeKeepRatio_NonExistentFile(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	result := module.ResizeKeepRatio("nonexistent.png", "output.png", 100, 100)
	if result {
		t.Error("Expected false for non-existent source file")
	}
}

func TestImageModule_Crop_NonExistentFile(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	result := module.Crop("nonexistent.png", "output.png", 0, 0, 50, 50)
	if result {
		t.Error("Expected false for non-existent source file")
	}
}

func TestImageModule_Crop_OutOfBounds(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	if err := storage.Write(projectID, "source.png", minimalPNG); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Try to crop beyond image bounds (image is 1x1)
	result := module.Crop("source.png", "output.png", 0, 0, 100, 100)
	if result {
		t.Error("Expected false for crop bounds exceeding image size")
	}
}

func TestImageModule_Crop_NegativeCoords(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	if err := storage.Write(projectID, "source.png", minimalPNG); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Try to crop with negative coordinates
	result := module.Crop("source.png", "output.png", -10, -10, 50, 50)
	if result {
		t.Error("Expected false for negative crop coordinates")
	}
}

func TestImageModule_Thumbnail_NonExistentFile(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	result := module.Thumbnail("nonexistent.png", "output.png", 100)
	if result {
		t.Error("Expected false for non-existent source file")
	}
}

func TestImageModule_Thumbnail_ValidPNG(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	if err := storage.Write(projectID, "source.png", minimalPNG); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result := module.Thumbnail("source.png", "thumb.png", 32)
	if !result {
		t.Error("Expected true for valid thumbnail operation")
	}

	// Verify thumbnail was created with correct size
	info := module.Info("thumb.png")
	if info == nil {
		t.Fatal("Expected thumbnail file to be created")
	}
	if info.Width != 32 || info.Height != 32 {
		t.Errorf("Expected 32x32, got %dx%d", info.Width, info.Height)
	}
}

func TestImageModule_ReadAsBase64_NonExistentFile(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	result := module.ReadAsBase64("nonexistent.png")
	if result != "" {
		t.Error("Expected empty string for non-existent file")
	}
}

func TestImageModule_ReadAsBase64_ValidPNG(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	if err := storage.Write(projectID, "test.png", minimalPNG); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result := module.ReadAsBase64("test.png")
	if result == "" {
		t.Error("Expected base64 data for valid file")
	}
	// Check data URI prefix
	expected := "data:image/png;base64,"
	if len(result) < len(expected) || result[:len(expected)] != expected {
		t.Errorf("Expected data URI prefix '%s', got: %s", expected, result[:min(len(result), 30)])
	}
}

func TestImageModule_ReadAsBase64_InvalidImage(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	// Write invalid image data
	if err := storage.Write(projectID, "invalid.png", []byte("not an image")); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result := module.ReadAsBase64("invalid.png")
	if result != "" {
		t.Error("Expected empty string for invalid image data")
	}
}

func TestImageModule_FormatConversion(t *testing.T) {
	storage, tmpDir := newTestStorageService(t)
	projectID := "test-project"
	createTestProjectDir(t, tmpDir, projectID)
	module := NewImageModule(storage, projectID)

	if err := storage.Write(projectID, "source.png", minimalPNG); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Resize PNG to JPEG
	result := module.Resize("source.png", "output.jpg", 50, 50)
	if !result {
		t.Error("Expected true for PNG to JPEG conversion")
	}

	// Verify output was created
	if !storage.Exists(projectID, "output.jpg") {
		t.Error("Expected JPEG output file to exist")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
