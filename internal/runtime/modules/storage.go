package modules

import (
	"m3m/internal/service"
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
