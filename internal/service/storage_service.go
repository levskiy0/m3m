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
	ThumbURL    string    `json:"thumb_url,omitempty"`
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
			// Generate public URLs (without /api prefix)
			fileInfo.URL = fmt.Sprintf("%s/cdn/%s%s", s.config.Server.URI, projectID, filePath)
			fileInfo.DownloadURL = fmt.Sprintf("%s/cdn/download/%s%s", s.config.Server.URI, projectID, filePath)
			// Add thumb_url for images
			if strings.HasPrefix(fileInfo.MimeType, "image/") {
				fileInfo.ThumbURL = fmt.Sprintf("%s/cdn/resize/64x64/%s%s", s.config.Server.URI, projectID, filePath)
			}
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

	// Clean up resize cache for old path
	s.cleanupResizeCache(projectID, oldPath)

	return os.Rename(oldFullPath, newFullPath)
}

func (s *StorageService) Delete(projectID, relativePath string) error {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return err
	}

	// Clean up resize cache
	s.cleanupResizeCache(projectID, relativePath)

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

// GetPath returns the absolute filesystem path for a file in project storage
func (s *StorageService) GetPath(projectID, relativePath string) (string, error) {
	return s.resolvePath(projectID, relativePath)
}

// cleanupResizeCache removes all cached resized versions of a file or directory
func (s *StorageService) cleanupResizeCache(projectID, relativePath string) {
	tmpDir := filepath.Join(s.config.Storage.Path, projectID, "tmp", "resize")

	// Read all size directories (e.g., 64x64, 100x100)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return // No cache directory exists
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cachePath := filepath.Join(tmpDir, entry.Name(), relativePath)
		os.RemoveAll(cachePath) // Remove file or directory
	}
}

// GetResizedImage returns path to resized image, creating it if needed
// Images are cached in /storage/{project-id}/tmp/resize/{WxH}/...
func (s *StorageService) GetResizedImage(projectID, relativePath string, width, height int) (string, error) {
	// Build cache path
	sizeDir := fmt.Sprintf("%dx%d", width, height)
	cacheDir := filepath.Join(s.config.Storage.Path, projectID, "tmp", "resize", sizeDir)
	cachePath := filepath.Join(cacheDir, relativePath)

	// Check if cached version exists
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	// Get original file
	originalPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(originalPath); os.IsNotExist(err) {
		return "", ErrFileNotFound
	}

	// Open and decode original image
	file, err := os.Open(originalPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate new dimensions (object-contain - fit within bounds, preserve aspect ratio)
	origBounds := img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	ratio := float64(origWidth) / float64(origHeight)
	targetRatio := float64(width) / float64(height)

	var newWidth, newHeight int
	if ratio > targetRatio {
		// Image is wider - fit by width
		newWidth = width
		newHeight = int(float64(width) / ratio)
	} else {
		// Image is taller - fit by height
		newHeight = height
		newWidth = int(float64(height) * ratio)
	}

	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, origBounds, draw.Over, nil)

	// Create cache directory
	cacheFileDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheFileDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache dir: %w", err)
	}

	// Save to cache
	outFile, err := os.Create(cachePath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %w", err)
	}
	defer outFile.Close()

	switch format {
	case "png":
		err = png.Encode(outFile, resized)
	case "gif":
		err = png.Encode(outFile, resized) // Convert GIF to PNG for simplicity
	default:
		err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 85})
	}

	if err != nil {
		os.Remove(cachePath)
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	return cachePath, nil
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
