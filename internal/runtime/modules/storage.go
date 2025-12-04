package modules

import (
	"m3m/internal/service"

	"github.com/dop251/goja"
)

type StorageModule struct {
	storage   *service.StorageService
	projectID string
}

func NewStorageModule(storage *service.StorageService, projectID string) *StorageModule {
	return &StorageModule{
		storage:   storage,
		projectID: projectID,
	}
}

// Name returns the module name for JavaScript
func (s *StorageModule) Name() string {
	return "storage"
}

// Register registers the module into the JavaScript VM
func (s *StorageModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(s.Name(), map[string]interface{}{
		"read":   s.Read,
		"write":  s.Write,
		"exists": s.Exists,
		"delete": s.Delete,
		"list":   s.List,
		"mkdir":  s.MkDir,
	})
}

func (s *StorageModule) Read(path string) string {
	data, err := s.storage.Read(s.projectID, path)
	if err != nil {
		return ""
	}
	return string(data)
}

func (s *StorageModule) Write(path string, content string) bool {
	err := s.storage.Write(s.projectID, path, []byte(content))
	return err == nil
}

func (s *StorageModule) Exists(path string) bool {
	return s.storage.Exists(s.projectID, path)
}

func (s *StorageModule) Delete(path string) bool {
	err := s.storage.Delete(s.projectID, path)
	return err == nil
}

func (s *StorageModule) List(path string) []string {
	files, err := s.storage.List(s.projectID, path)
	if err != nil {
		return []string{}
	}

	result := make([]string, len(files))
	for i, f := range files {
		result[i] = f.Name
	}
	return result
}

func (s *StorageModule) MkDir(path string) bool {
	err := s.storage.MkDir(s.projectID, path)
	return err == nil
}

// GetSchema implements JSSchemaProvider
func (s *StorageModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "storage",
		Description: "File storage operations for the project",
		Methods: []JSMethodSchema{
			{
				Name:        "read",
				Description: "Read file contents as string",
				Params:      []JSParamSchema{{Name: "path", Type: "string", Description: "File path relative to project storage"}},
				Returns:     &JSParamSchema{Type: "string"},
			},
			{
				Name:        "write",
				Description: "Write string content to file",
				Params: []JSParamSchema{
					{Name: "path", Type: "string", Description: "File path relative to project storage"},
					{Name: "content", Type: "string", Description: "Content to write"},
				},
				Returns: &JSParamSchema{Type: "boolean"},
			},
			{
				Name:        "exists",
				Description: "Check if file or directory exists",
				Params:      []JSParamSchema{{Name: "path", Type: "string", Description: "File path to check"}},
				Returns:     &JSParamSchema{Type: "boolean"},
			},
			{
				Name:        "delete",
				Description: "Delete file or directory",
				Params:      []JSParamSchema{{Name: "path", Type: "string", Description: "File path to delete"}},
				Returns:     &JSParamSchema{Type: "boolean"},
			},
			{
				Name:        "list",
				Description: "List files in directory",
				Params:      []JSParamSchema{{Name: "path", Type: "string", Description: "Directory path"}},
				Returns:     &JSParamSchema{Type: "string[]"},
			},
			{
				Name:        "mkdir",
				Description: "Create directory",
				Params:      []JSParamSchema{{Name: "path", Type: "string", Description: "Directory path to create"}},
				Returns:     &JSParamSchema{Type: "boolean"},
			},
		},
	}
}

// GetStorageSchema returns the storage schema (static version)
func GetStorageSchema() JSModuleSchema {
	return (&StorageModule{}).GetSchema()
}
