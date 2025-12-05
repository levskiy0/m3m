package service

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/image/draw"

	"m3m/internal/config"
)

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrDirectoryNotFound = errors.New("directory not found")
	ErrInvalidPath       = errors.New("invalid path")
)

type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	IsDir       bool      `json:"is_dir"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mime_type"`
	UpdatedAt   time.Time `json:"updated_at"`
	URL         string    `json:"url,omitempty"`
	DownloadURL string    `json:"download_url,omitempty"`
}

type StorageService struct {
	config *config.Config
}

func NewStorageService(config *config.Config) *StorageService {
	return &StorageService{
		config: config,
	}
}

func (s *StorageService) getProjectStoragePath(projectID string) string {
	return filepath.Join(s.config.Storage.Path, projectID, "storage")
}

func (s *StorageService) resolvePath(projectID, relativePath string) (string, error) {
	basePath := s.getProjectStoragePath(projectID)
	fullPath := filepath.Join(basePath, relativePath)

	// Prevent directory traversal
	if !strings.HasPrefix(fullPath, basePath) {
		return "", ErrInvalidPath
	}

	return fullPath, nil
}

func (s *StorageService) List(projectID, relativePath string) ([]FileInfo, error) {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDirectoryNotFound
		}
		return nil, err
	}

	files := make([]FileInfo, 0)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(relativePath, entry.Name())
		fileInfo := FileInfo{
			Name:      entry.Name(),
			Path:      filePath,
			IsDir:     entry.IsDir(),
			Size:      info.Size(),
			UpdatedAt: info.ModTime(),
		}

		if !entry.IsDir() {
			fileInfo.MimeType = getMimeType(entry.Name())
			// Generate public URLs
			fileInfo.URL = fmt.Sprintf("%s/api/cdn/%s%s", s.config.Server.URI, projectID, filePath)
			fileInfo.DownloadURL = fmt.Sprintf("%s/api/download/%s%s", s.config.Server.URI, projectID, filePath)
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

func (s *StorageService) MkDir(projectID, relativePath string) error {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return err
	}

	return os.MkdirAll(fullPath, 0755)
}

func (s *StorageService) Upload(projectID, relativePath string, file *multipart.FileHeader) error {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (s *StorageService) Download(projectID, relativePath string) (string, error) {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", ErrFileNotFound
	}

	return fullPath, nil
}

func (s *StorageService) Read(projectID, relativePath string) ([]byte, error) {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(fullPath)
}

func (s *StorageService) Write(projectID, relativePath string, content []byte) error {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(fullPath, content, 0644)
}

func (s *StorageService) Rename(projectID, oldPath, newPath string) error {
	oldFullPath, err := s.resolvePath(projectID, oldPath)
	if err != nil {
		return err
	}

	newFullPath, err := s.resolvePath(projectID, newPath)
	if err != nil {
		return err
	}

	return os.Rename(oldFullPath, newFullPath)
}

func (s *StorageService) Delete(projectID, relativePath string) error {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return err
	}

	return os.RemoveAll(fullPath)
}

func (s *StorageService) Exists(projectID, relativePath string) bool {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return false
	}

	_, err = os.Stat(fullPath)
	return err == nil
}

func (s *StorageService) GenerateThumbnail(projectID, relativePath string, width, height int) ([]byte, error) {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Create thumbnail
	thumb := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Encode to bytes
	var buf strings.Builder
	writer := &writerAdapter{&buf}

	switch format {
	case "png":
		err = png.Encode(writer, thumb)
	default:
		err = jpeg.Encode(writer, thumb, &jpeg.Options{Quality: 80})
	}

	if err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

type writerAdapter struct {
	builder *strings.Builder
}

func (w *writerAdapter) Write(p []byte) (n int, err error) {
	return w.builder.Write(p)
}

func (s *StorageService) GetLogsPath(projectID string) string {
	return filepath.Join(s.config.Logging.Path, projectID)
}

func (s *StorageService) CreateLogFile(projectID string) (string, error) {
	logsPath := s.GetLogsPath(projectID)
	if err := os.MkdirAll(logsPath, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s.log", uuid.New().String())
	return filepath.Join(logsPath, filename), nil
}

func (s *StorageService) ClearLogs(projectID string) error {
	logsPath := s.GetLogsPath(projectID)
	return os.RemoveAll(logsPath)
}

// GetStorageSize calculates total size of project storage folder in bytes
func (s *StorageService) GetStorageSize(projectID string) (int64, error) {
	storagePath := s.getProjectStoragePath(projectID)

	var totalSize int64
	err := filepath.Walk(storagePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil // Directory doesn't exist yet, return 0
			}
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return totalSize, nil
}

func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".yaml": "text/yaml",
		".yml":  "text/yaml",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
