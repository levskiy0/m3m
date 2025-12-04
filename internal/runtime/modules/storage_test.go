package modules

import (
	"os"
	"path/filepath"
	"testing"

	"m3m/internal/config"
	"m3m/internal/service"
)

func setupStorageTest(t *testing.T) (*StorageModule, string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "m3m-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		Storage: config.StorageConfig{
			Path: tmpDir,
		},
		Logging: config.LoggingConfig{
			Path: filepath.Join(tmpDir, "logs"),
		},
	}

	storageService := service.NewStorageService(cfg)
	projectID := "test-project-123"

	// Create project storage directory
	projectPath := filepath.Join(tmpDir, projectID, "storage")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project storage dir: %v", err)
	}

	module := NewStorageModule(storageService, projectID)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return module, projectPath, cleanup
}

func TestStorageModule_Write_Read(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	content := "Hello, World!"
	path := "test.txt"

	// Write
	ok := module.Write(path, content)
	if !ok {
		t.Error("Write() returned false")
	}

	// Read
	got := module.Read(path)
	if got != content {
		t.Errorf("Read() = %q, want %q", got, content)
	}
}

func TestStorageModule_Write_NestedPath(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	content := "nested content"
	path := "dir1/dir2/file.txt"

	ok := module.Write(path, content)
	if !ok {
		t.Error("Write() to nested path returned false")
	}

	got := module.Read(path)
	if got != content {
		t.Errorf("Read() nested path = %q, want %q", got, content)
	}
}

func TestStorageModule_Read_NonExistent(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	got := module.Read("non-existent.txt")
	if got != "" {
		t.Errorf("Read() non-existent = %q, want empty string", got)
	}
}

func TestStorageModule_Exists(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Create file
	module.Write("exists.txt", "content")

	// Test exists
	if !module.Exists("exists.txt") {
		t.Error("Exists() returned false for existing file")
	}

	// Test non-existent
	if module.Exists("does-not-exist.txt") {
		t.Error("Exists() returned true for non-existent file")
	}
}

func TestStorageModule_Delete(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Create and delete file
	module.Write("to-delete.txt", "content")

	ok := module.Delete("to-delete.txt")
	if !ok {
		t.Error("Delete() returned false")
	}

	if module.Exists("to-delete.txt") {
		t.Error("File still exists after Delete()")
	}
}

func TestStorageModule_Delete_NonExistent(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Delete should succeed even for non-existent files
	ok := module.Delete("non-existent.txt")
	if !ok {
		t.Error("Delete() of non-existent file returned false")
	}
}

func TestStorageModule_Delete_Directory(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Create directory with files
	module.Write("dir/file1.txt", "content1")
	module.Write("dir/file2.txt", "content2")

	// Delete directory
	ok := module.Delete("dir")
	if !ok {
		t.Error("Delete() directory returned false")
	}

	if module.Exists("dir") {
		t.Error("Directory still exists after Delete()")
	}
}

func TestStorageModule_MkDir(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	ok := module.MkDir("new-dir")
	if !ok {
		t.Error("MkDir() returned false")
	}

	if !module.Exists("new-dir") {
		t.Error("Directory doesn't exist after MkDir()")
	}
}

func TestStorageModule_MkDir_Nested(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	ok := module.MkDir("a/b/c/d")
	if !ok {
		t.Error("MkDir() nested returned false")
	}

	if !module.Exists("a/b/c/d") {
		t.Error("Nested directory doesn't exist after MkDir()")
	}
}

func TestStorageModule_List(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Create files
	module.Write("file1.txt", "content1")
	module.Write("file2.txt", "content2")
	module.MkDir("subdir")

	files := module.List("")
	if len(files) != 3 {
		t.Errorf("List() returned %d items, want 3", len(files))
	}

	// Check files are present
	hasFile1, hasFile2, hasSubdir := false, false, false
	for _, f := range files {
		switch f {
		case "file1.txt":
			hasFile1 = true
		case "file2.txt":
			hasFile2 = true
		case "subdir":
			hasSubdir = true
		}
	}

	if !hasFile1 || !hasFile2 || !hasSubdir {
		t.Errorf("List() missing expected items: file1=%v file2=%v subdir=%v", hasFile1, hasFile2, hasSubdir)
	}
}

func TestStorageModule_List_Subdirectory(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Create files in subdirectory
	module.Write("subdir/file1.txt", "content1")
	module.Write("subdir/file2.txt", "content2")

	files := module.List("subdir")
	if len(files) != 2 {
		t.Errorf("List(subdir) returned %d items, want 2", len(files))
	}
}

func TestStorageModule_List_Empty(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	files := module.List("")
	if len(files) != 0 {
		t.Errorf("List() on empty dir returned %d items, want 0", len(files))
	}
}

func TestStorageModule_List_NonExistent(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	files := module.List("non-existent-dir")
	if len(files) != 0 {
		t.Errorf("List() on non-existent dir returned %d items, want 0", len(files))
	}
}

func TestStorageModule_WriteOverwrite(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Write initial content
	module.Write("file.txt", "initial")

	// Overwrite
	module.Write("file.txt", "overwritten")

	got := module.Read("file.txt")
	if got != "overwritten" {
		t.Errorf("Read() after overwrite = %q, want 'overwritten'", got)
	}
}

func TestStorageModule_EmptyContent(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Write empty content
	ok := module.Write("empty.txt", "")
	if !ok {
		t.Error("Write() empty content returned false")
	}

	got := module.Read("empty.txt")
	if got != "" {
		t.Errorf("Read() empty file = %q, want empty string", got)
	}

	if !module.Exists("empty.txt") {
		t.Error("Empty file doesn't exist")
	}
}

func TestStorageModule_LargeContent(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Create 1MB content
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('a' + (i % 26))
	}

	ok := module.Write("large.txt", string(largeContent))
	if !ok {
		t.Error("Write() large content returned false")
	}

	got := module.Read("large.txt")
	if got != string(largeContent) {
		t.Error("Read() large file content mismatch")
	}
}

func TestStorageModule_SpecialCharacters(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	content := "Special chars: \n\t\r unicode: 日本語 emoji: 🎉"
	module.Write("special.txt", content)

	got := module.Read("special.txt")
	if got != content {
		t.Errorf("Read() special chars = %q, want %q", got, content)
	}
}

func TestStorageModule_BinaryLikeContent(t *testing.T) {
	module, _, cleanup := setupStorageTest(t)
	defer cleanup()

	// Create content with null bytes
	content := "before\x00after"
	module.Write("binary.txt", content)

	got := module.Read("binary.txt")
	if got != content {
		t.Errorf("Read() binary-like content mismatch: len=%d want=%d", len(got), len(content))
	}
}
