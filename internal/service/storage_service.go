package service

import (
	"archive/zip"
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

	// Resolve symlinks to prevent symlink-based path traversal
	// Only check if the path exists (for reads/deletes)
	if realPath, err := filepath.EvalSymlinks(fullPath); err == nil {
		realBasePath, _ := filepath.EvalSymlinks(basePath)
		if realBasePath == "" {
			realBasePath = basePath
		}
		if !strings.HasPrefix(realPath, realBasePath) {
			return "", ErrInvalidPath
		}
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

// Stat returns detailed file information
func (s *StorageService) Stat(projectID, relativePath string) (*FileInfo, error) {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	fileInfo := &FileInfo{
		Name:      info.Name(),
		Path:      relativePath,
		IsDir:     info.IsDir(),
		Size:      info.Size(),
		UpdatedAt: info.ModTime(),
	}

	if !info.IsDir() {
		fileInfo.MimeType = getMimeType(info.Name())
		fileInfo.URL = fmt.Sprintf("%s/cdn/%s%s", s.config.Server.URI, projectID, relativePath)
		fileInfo.DownloadURL = fmt.Sprintf("%s/cdn/download/%s%s", s.config.Server.URI, projectID, relativePath)
		if strings.HasPrefix(fileInfo.MimeType, "image/") {
			fileInfo.ThumbURL = fmt.Sprintf("%s/cdn/resize/64x64/%s%s", s.config.Server.URI, projectID, relativePath)
		}
	}

	return fileInfo, nil
}

// Append adds content to the end of a file
func (s *StorageService) Append(projectID, relativePath string, content []byte) error {
	fullPath, err := s.resolvePath(projectID, relativePath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	return err
}

// Copy copies a file or directory
func (s *StorageService) Copy(projectID, srcPath, dstPath string) error {
	srcFullPath, err := s.resolvePath(projectID, srcPath)
	if err != nil {
		return err
	}

	dstFullPath, err := s.resolvePath(projectID, dstPath)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(srcFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}

	if srcInfo.IsDir() {
		return s.copyDir(srcFullPath, dstFullPath)
	}
	return s.copyFile(srcFullPath, dstFullPath)
}

func (s *StorageService) copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *StorageService) copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := s.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := s.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Move moves a file or directory
func (s *StorageService) Move(projectID, srcPath, dstPath string) error {
	srcFullPath, err := s.resolvePath(projectID, srcPath)
	if err != nil {
		return err
	}

	dstFullPath, err := s.resolvePath(projectID, dstPath)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dstFullPath), 0755); err != nil {
		return err
	}

	// Clean up resize cache for source path
	s.cleanupResizeCache(projectID, srcPath)

	return os.Rename(srcFullPath, dstFullPath)
}

// Glob returns files matching a glob pattern
func (s *StorageService) Glob(projectID, pattern string) ([]string, error) {
	basePath := s.getProjectStoragePath(projectID)
	fullPattern := filepath.Join(basePath, pattern)

	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, err
	}

	// Convert to relative paths
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		relPath, err := filepath.Rel(basePath, match)
		if err != nil {
			continue
		}
		result = append(result, relPath)
	}

	return result, nil
}

// Zip creates a zip archive from source paths
func (s *StorageService) Zip(projectID string, srcPaths []string, dstPath string) error {
	dstFullPath, err := s.resolvePath(projectID, dstPath)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dstFullPath), 0755); err != nil {
		return err
	}

	zipFile, err := os.Create(dstFullPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, srcPath := range srcPaths {
		srcFullPath, err := s.resolvePath(projectID, srcPath)
		if err != nil {
			return err
		}

		info, err := os.Stat(srcFullPath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := s.addDirToZip(zipWriter, srcFullPath, srcPath); err != nil {
				return err
			}
		} else {
			if err := s.addFileToZip(zipWriter, srcFullPath, srcPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *StorageService) addFileToZip(zipWriter *zip.Writer, filePath, archivePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = archivePath
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func (s *StorageService) addDirToZip(zipWriter *zip.Writer, dirPath, archivePath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())
		entryArchivePath := filepath.Join(archivePath, entry.Name())

		if entry.IsDir() {
			if err := s.addDirToZip(zipWriter, entryPath, entryArchivePath); err != nil {
				return err
			}
		} else {
			if err := s.addFileToZip(zipWriter, entryPath, entryArchivePath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Unzip extracts a zip archive
func (s *StorageService) Unzip(projectID, srcPath, dstPath string) error {
	srcFullPath, err := s.resolvePath(projectID, srcPath)
	if err != nil {
		return err
	}

	dstFullPath, err := s.resolvePath(projectID, dstPath)
	if err != nil {
		return err
	}

	reader, err := zip.OpenReader(srcFullPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(dstFullPath, file.Name)

		// Check for path traversal
		if !strings.HasPrefix(filePath, filepath.Clean(dstFullPath)+string(os.PathSeparator)) {
			return ErrInvalidPath
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// GetURL returns the public URL for a file
func (s *StorageService) GetURL(projectID, relativePath string) string {
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}
	return fmt.Sprintf("%s/cdn/%s%s", s.config.Server.URI, projectID, relativePath)
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
