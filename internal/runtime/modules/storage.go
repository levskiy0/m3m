package modules

import (
	"encoding/base64"

	"m3m/internal/service"

	"github.com/dop251/goja"
	"m3m/pkg/schema"
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
	return "$storage"
}

// Register registers the module into the JavaScript VM
func (s *StorageModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(s.Name(), map[string]interface{}{
		"read":       s.Read,
		"readBase64": s.ReadBase64,
		"write":      s.Write,
		"append":     s.Append,
		"exists":     s.Exists,
		"delete":     s.Delete,
		"list":       s.List,
		"mkdir":      s.MkDir,
		"getPath":    s.GetPath,
		"stat":       s.Stat,
		"size":       s.Size,
		"mimeType":   s.MimeType,
		"copy":       s.Copy,
		"move":       s.Move,
		"rename":     s.Rename,
		"glob":       s.Glob,
		"getUrl":     s.GetUrl,
		"zip":        s.Zip,
		"unzip":      s.Unzip,
		"tmp": map[string]interface{}{
			"read":       s.TmpRead,
			"readBase64": s.TmpReadBase64,
			"write":      s.TmpWrite,
			"append":     s.TmpAppend,
			"exists":     s.TmpExists,
			"delete":     s.TmpDelete,
			"list":       s.TmpList,
			"mkdir":      s.TmpMkDir,
			"getPath":    s.TmpGetPath,
			"stat":       s.TmpStat,
			"size":       s.TmpSize,
			"mimeType":   s.TmpMimeType,
			"copy":       s.TmpCopy,
			"move":       s.TmpMove,
			"rename":     s.TmpRename,
			"glob":       s.TmpGlob,
			"getUrl":     s.TmpGetUrl,
			"zip":        s.TmpZip,
			"unzip":      s.TmpUnzip,
		},
	})
}

func (s *StorageModule) Read(path string) string {
	data, err := s.storage.Read(s.projectID, path)
	if err != nil {
		return ""
	}
	return string(data)
}

func (s *StorageModule) ReadBase64(path string) string {
	data, err := s.storage.Read(s.projectID, path)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
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

// Tmp methods - wrap main methods with "tmp/" prefix

func (s *StorageModule) TmpRead(path string) string {
	return s.Read("tmp/" + path)
}

func (s *StorageModule) TmpReadBase64(path string) string {
	return s.ReadBase64("tmp/" + path)
}

func (s *StorageModule) TmpWrite(path string, content string) bool {
	return s.Write("tmp/"+path, content)
}

func (s *StorageModule) TmpExists(path string) bool {
	return s.Exists("tmp/" + path)
}

func (s *StorageModule) TmpDelete(path string) bool {
	return s.Delete("tmp/" + path)
}

func (s *StorageModule) TmpList(path string) []string {
	return s.List("tmp/" + path)
}

func (s *StorageModule) TmpMkDir(path string) bool {
	return s.MkDir("tmp/" + path)
}

func (s *StorageModule) GetPath(path string) string {
	fullPath, err := s.storage.GetPath(s.projectID, path)
	if err != nil {
		return ""
	}
	return fullPath
}

func (s *StorageModule) TmpGetPath(path string) string {
	return s.GetPath("tmp/" + path)
}

// New methods

func (s *StorageModule) Append(path string, content string) bool {
	err := s.storage.Append(s.projectID, path, []byte(content))
	return err == nil
}

func (s *StorageModule) TmpAppend(path string, content string) bool {
	return s.Append("tmp/"+path, content)
}

func (s *StorageModule) Stat(path string) map[string]interface{} {
	info, err := s.storage.Stat(s.projectID, path)
	if err != nil {
		return nil
	}
	return map[string]interface{}{
		"name":      info.Name,
		"path":      info.Path,
		"size":      info.Size,
		"isDir":     info.IsDir,
		"mimeType":  info.MimeType,
		"updatedAt": info.UpdatedAt.Unix(),
		"url":       info.URL,
	}
}

func (s *StorageModule) TmpStat(path string) map[string]interface{} {
	return s.Stat("tmp/" + path)
}

func (s *StorageModule) Size(path string) int64 {
	info, err := s.storage.Stat(s.projectID, path)
	if err != nil {
		return -1
	}
	return info.Size
}

func (s *StorageModule) TmpSize(path string) int64 {
	return s.Size("tmp/" + path)
}

func (s *StorageModule) MimeType(path string) string {
	info, err := s.storage.Stat(s.projectID, path)
	if err != nil {
		return ""
	}
	return info.MimeType
}

func (s *StorageModule) TmpMimeType(path string) string {
	return s.MimeType("tmp/" + path)
}

func (s *StorageModule) Copy(src string, dst string) bool {
	err := s.storage.Copy(s.projectID, src, dst)
	return err == nil
}

func (s *StorageModule) TmpCopy(src string, dst string) bool {
	return s.Copy("tmp/"+src, "tmp/"+dst)
}

func (s *StorageModule) Move(src string, dst string) bool {
	err := s.storage.Move(s.projectID, src, dst)
	return err == nil
}

func (s *StorageModule) TmpMove(src string, dst string) bool {
	return s.Move("tmp/"+src, "tmp/"+dst)
}

func (s *StorageModule) Rename(oldPath string, newPath string) bool {
	err := s.storage.Rename(s.projectID, oldPath, newPath)
	return err == nil
}

func (s *StorageModule) TmpRename(oldPath string, newPath string) bool {
	return s.Rename("tmp/"+oldPath, "tmp/"+newPath)
}

func (s *StorageModule) Glob(pattern string) []string {
	matches, err := s.storage.Glob(s.projectID, pattern)
	if err != nil {
		return []string{}
	}
	return matches
}

func (s *StorageModule) TmpGlob(pattern string) []string {
	return s.Glob("tmp/" + pattern)
}

func (s *StorageModule) GetUrl(path string) string {
	return s.storage.GetURL(s.projectID, path)
}

func (s *StorageModule) TmpGetUrl(path string) string {
	return s.GetUrl("tmp/" + path)
}

func (s *StorageModule) Zip(srcPaths []string, dstPath string) bool {
	err := s.storage.Zip(s.projectID, srcPaths, dstPath)
	return err == nil
}

func (s *StorageModule) TmpZip(srcPaths []string, dstPath string) bool {
	// Prefix all src paths with tmp/
	tmpSrcPaths := make([]string, len(srcPaths))
	for i, p := range srcPaths {
		tmpSrcPaths[i] = "tmp/" + p
	}
	return s.Zip(tmpSrcPaths, "tmp/"+dstPath)
}

func (s *StorageModule) Unzip(srcPath string, dstPath string) bool {
	err := s.storage.Unzip(s.projectID, srcPath, dstPath)
	return err == nil
}

func (s *StorageModule) TmpUnzip(srcPath string, dstPath string) bool {
	return s.Unzip("tmp/"+srcPath, "tmp/"+dstPath)
}

// GetSchema implements JSSchemaProvider
func (s *StorageModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$storage",
		Description: "File storage operations for the project",
		Methods: []schema.MethodSchema{
			{
				Name:        "read",
				Description: "Read file contents as string",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to project storage"}},
				Returns:     &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "readBase64",
				Description: "Read file contents as base64 encoded string (for binary files like images)",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to project storage"}},
				Returns:     &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "write",
				Description: "Write string content to file",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "File path relative to project storage"},
					{Name: "content", Type: "string", Description: "Content to write"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "append",
				Description: "Append string content to end of file",
				Params: []schema.ParamSchema{
					{Name: "path", Type: "string", Description: "File path relative to project storage"},
					{Name: "content", Type: "string", Description: "Content to append"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "exists",
				Description: "Check if file or directory exists",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path to check"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "delete",
				Description: "Delete file or directory",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path to delete"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "list",
				Description: "List files in directory",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "Directory path"}},
				Returns:     &schema.ParamSchema{Type: "string[]"},
			},
			{
				Name:        "mkdir",
				Description: "Create directory",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "Directory path to create"}},
				Returns:     &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "getPath",
				Description: "Get absolute filesystem path for a file",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to project storage"}},
				Returns:     &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "stat",
				Description: "Get detailed file information",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to project storage"}},
				Returns:     &schema.ParamSchema{Type: "{ name: string, path: string, size: number, isDir: boolean, mimeType: string, updatedAt: number, url: string } | null"},
			},
			{
				Name:        "size",
				Description: "Get file size in bytes",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to project storage"}},
				Returns:     &schema.ParamSchema{Type: "number"},
			},
			{
				Name:        "mimeType",
				Description: "Get file MIME type",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to project storage"}},
				Returns:     &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "copy",
				Description: "Copy file or directory",
				Params: []schema.ParamSchema{
					{Name: "src", Type: "string", Description: "Source path"},
					{Name: "dst", Type: "string", Description: "Destination path"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "move",
				Description: "Move file or directory",
				Params: []schema.ParamSchema{
					{Name: "src", Type: "string", Description: "Source path"},
					{Name: "dst", Type: "string", Description: "Destination path"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "rename",
				Description: "Rename file or directory",
				Params: []schema.ParamSchema{
					{Name: "oldPath", Type: "string", Description: "Current path"},
					{Name: "newPath", Type: "string", Description: "New path"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "glob",
				Description: "Find files matching a glob pattern",
				Params:      []schema.ParamSchema{{Name: "pattern", Type: "string", Description: "Glob pattern (e.g., '**/*.json')"}},
				Returns:     &schema.ParamSchema{Type: "string[]"},
			},
			{
				Name:        "getUrl",
				Description: "Get public URL for a file",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to project storage"}},
				Returns:     &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "zip",
				Description: "Create a zip archive from files/directories",
				Params: []schema.ParamSchema{
					{Name: "srcPaths", Type: "string[]", Description: "Array of source paths to archive"},
					{Name: "dstPath", Type: "string", Description: "Destination zip file path"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
			{
				Name:        "unzip",
				Description: "Extract a zip archive",
				Params: []schema.ParamSchema{
					{Name: "srcPath", Type: "string", Description: "Source zip file path"},
					{Name: "dstPath", Type: "string", Description: "Destination directory path"},
				},
				Returns: &schema.ParamSchema{Type: "boolean"},
			},
		},
		Nested: []schema.NestedModuleSchema{
			{
				Name:        "tmp",
				Description: "Temporary storage operations (files stored in tmp/ directory)",
				Methods: []schema.MethodSchema{
					{
						Name:        "read",
						Description: "Read file contents from tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to tmp storage"}},
						Returns:     &schema.ParamSchema{Type: "string"},
					},
					{
						Name:        "readBase64",
						Description: "Read file contents as base64 encoded string from tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to tmp storage"}},
						Returns:     &schema.ParamSchema{Type: "string"},
					},
					{
						Name:        "write",
						Description: "Write string content to tmp storage",
						Params: []schema.ParamSchema{
							{Name: "path", Type: "string", Description: "File path relative to tmp storage"},
							{Name: "content", Type: "string", Description: "Content to write"},
						},
						Returns: &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "append",
						Description: "Append string content to end of file in tmp storage",
						Params: []schema.ParamSchema{
							{Name: "path", Type: "string", Description: "File path relative to tmp storage"},
							{Name: "content", Type: "string", Description: "Content to append"},
						},
						Returns: &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "exists",
						Description: "Check if file exists in tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path to check"}},
						Returns:     &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "delete",
						Description: "Delete file from tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path to delete"}},
						Returns:     &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "list",
						Description: "List files in tmp directory",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "Directory path"}},
						Returns:     &schema.ParamSchema{Type: "string[]"},
					},
					{
						Name:        "mkdir",
						Description: "Create directory in tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "Directory path to create"}},
						Returns:     &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "getPath",
						Description: "Get absolute filesystem path for a file in tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to tmp storage"}},
						Returns:     &schema.ParamSchema{Type: "string"},
					},
					{
						Name:        "stat",
						Description: "Get detailed file information from tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to tmp storage"}},
						Returns:     &schema.ParamSchema{Type: "{ name: string, path: string, size: number, isDir: boolean, mimeType: string, updatedAt: number, url: string } | null"},
					},
					{
						Name:        "size",
						Description: "Get file size in bytes from tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to tmp storage"}},
						Returns:     &schema.ParamSchema{Type: "number"},
					},
					{
						Name:        "mimeType",
						Description: "Get file MIME type from tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to tmp storage"}},
						Returns:     &schema.ParamSchema{Type: "string"},
					},
					{
						Name:        "copy",
						Description: "Copy file or directory within tmp storage",
						Params: []schema.ParamSchema{
							{Name: "src", Type: "string", Description: "Source path"},
							{Name: "dst", Type: "string", Description: "Destination path"},
						},
						Returns: &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "move",
						Description: "Move file or directory within tmp storage",
						Params: []schema.ParamSchema{
							{Name: "src", Type: "string", Description: "Source path"},
							{Name: "dst", Type: "string", Description: "Destination path"},
						},
						Returns: &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "rename",
						Description: "Rename file or directory in tmp storage",
						Params: []schema.ParamSchema{
							{Name: "oldPath", Type: "string", Description: "Current path"},
							{Name: "newPath", Type: "string", Description: "New path"},
						},
						Returns: &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "glob",
						Description: "Find files matching a glob pattern in tmp storage",
						Params:      []schema.ParamSchema{{Name: "pattern", Type: "string", Description: "Glob pattern (e.g., '**/*.json')"}},
						Returns:     &schema.ParamSchema{Type: "string[]"},
					},
					{
						Name:        "getUrl",
						Description: "Get public URL for a file in tmp storage",
						Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "File path relative to tmp storage"}},
						Returns:     &schema.ParamSchema{Type: "string"},
					},
					{
						Name:        "zip",
						Description: "Create a zip archive from files/directories in tmp storage",
						Params: []schema.ParamSchema{
							{Name: "srcPaths", Type: "string[]", Description: "Array of source paths to archive"},
							{Name: "dstPath", Type: "string", Description: "Destination zip file path"},
						},
						Returns: &schema.ParamSchema{Type: "boolean"},
					},
					{
						Name:        "unzip",
						Description: "Extract a zip archive in tmp storage",
						Params: []schema.ParamSchema{
							{Name: "srcPath", Type: "string", Description: "Source zip file path"},
							{Name: "dstPath", Type: "string", Description: "Destination directory path"},
						},
						Returns: &schema.ParamSchema{Type: "boolean"},
					},
				},
			},
		},
	}
}

// GetStorageSchema returns the storage schema (static version)
func GetStorageSchema() schema.ModuleSchema {
	return (&StorageModule{}).GetSchema()
}
