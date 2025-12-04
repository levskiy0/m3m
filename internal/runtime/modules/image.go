package modules

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"

	"m3m/internal/service"
)

type ImageModule struct {
	storage   *service.StorageService
	projectID string
}

type ImageInfo struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
}

func NewImageModule(storage *service.StorageService, projectID string) *ImageModule {
	return &ImageModule{
		storage:   storage,
		projectID: projectID,
	}
}

// Info returns information about an image (width, height, format)
func (m *ImageModule) Info(path string) *ImageInfo {
	data, err := m.storage.Read(m.projectID, path)
	if err != nil {
		return nil
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	bounds := img.Bounds()
	return &ImageInfo{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: format,
	}
}

// Resize resizes an image to the specified dimensions and saves to destination
func (m *ImageModule) Resize(src, dst string, width, height int) bool {
	data, err := m.storage.Read(m.projectID, src)
	if err != nil {
		return false
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Encode to bytes
	var buf bytes.Buffer
	if err := m.encodeImage(&buf, resized, format, dst); err != nil {
		return false
	}

	// Write to storage
	return m.storage.Write(m.projectID, dst, buf.Bytes()) == nil
}

// ResizeKeepRatio resizes an image keeping aspect ratio (fit within bounds)
func (m *ImageModule) ResizeKeepRatio(src, dst string, maxWidth, maxHeight int) bool {
	info := m.Info(src)
	if info == nil {
		return false
	}

	// Calculate new dimensions keeping aspect ratio
	ratio := float64(info.Width) / float64(info.Height)
	newWidth := maxWidth
	newHeight := int(float64(maxWidth) / ratio)

	if newHeight > maxHeight {
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * ratio)
	}

	return m.Resize(src, dst, newWidth, newHeight)
}

// Crop crops an image to the specified region
func (m *ImageModule) Crop(src, dst string, x, y, width, height int) bool {
	data, err := m.storage.Read(m.projectID, src)
	if err != nil {
		return false
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}

	bounds := img.Bounds()
	// Validate crop bounds
	if x < 0 || y < 0 || x+width > bounds.Dx() || y+height > bounds.Dy() {
		return false
	}

	// Create cropped image
	cropped := image.NewRGBA(image.Rect(0, 0, width, height))
	cropRect := image.Rect(x, y, x+width, y+height)
	draw.Copy(cropped, image.Point{}, img, cropRect, draw.Over, nil)

	// Encode to bytes
	var buf bytes.Buffer
	if err := m.encodeImage(&buf, cropped, format, dst); err != nil {
		return false
	}

	// Write to storage
	return m.storage.Write(m.projectID, dst, buf.Bytes()) == nil
}

// Thumbnail creates a square thumbnail
func (m *ImageModule) Thumbnail(src, dst string, size int) bool {
	data, err := m.storage.Read(m.projectID, src)
	if err != nil {
		return false
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Calculate crop region for center square
	var cropRect image.Rectangle
	if w > h {
		offset := (w - h) / 2
		cropRect = image.Rect(offset, 0, offset+h, h)
	} else {
		offset := (h - w) / 2
		cropRect = image.Rect(0, offset, w, offset+w)
	}

	// First crop to square
	square := image.NewRGBA(image.Rect(0, 0, cropRect.Dx(), cropRect.Dy()))
	draw.Copy(square, image.Point{}, img, cropRect, draw.Over, nil)

	// Then resize to target size
	thumb := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), square, square.Bounds(), draw.Over, nil)

	// Encode to bytes
	var buf bytes.Buffer
	if err := m.encodeImage(&buf, thumb, format, dst); err != nil {
		return false
	}

	// Write to storage
	return m.storage.Write(m.projectID, dst, buf.Bytes()) == nil
}

// encodeImage encodes an image to the appropriate format
func (m *ImageModule) encodeImage(buf *bytes.Buffer, img image.Image, originalFormat, dstPath string) error {
	// Determine format from destination extension, falling back to original format
	ext := strings.ToLower(filepath.Ext(dstPath))
	format := originalFormat

	switch ext {
	case ".png":
		format = "png"
	case ".jpg", ".jpeg":
		format = "jpeg"
	}

	switch format {
	case "png":
		return png.Encode(buf, img)
	default:
		return jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
	}
}

// ReadAsBase64 reads an image and returns it as base64 data URI
func (m *ImageModule) ReadAsBase64(path string) string {
	data, err := m.storage.Read(m.projectID, path)
	if err != nil {
		return ""
	}

	// Detect format
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ""
	}

	// Determine MIME type
	mimeType := "image/jpeg"
	switch format {
	case "png":
		mimeType = "image/png"
	case "gif":
		mimeType = "image/gif"
	case "webp":
		mimeType = "image/webp"
	}

	// Encode to base64
	encoded := m.base64Encode(data)
	return "data:" + mimeType + ";base64," + encoded
}

func (m *ImageModule) base64Encode(data []byte) string {
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder
	result.Grow(((len(data) + 2) / 3) * 4)

	for i := 0; i < len(data); i += 3 {
		var n uint32
		remaining := len(data) - i
		switch remaining {
		case 1:
			n = uint32(data[i]) << 16
		case 2:
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8
		default:
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8 | uint32(data[i+2])
		}

		result.WriteByte(base64Table[(n>>18)&0x3F])
		result.WriteByte(base64Table[(n>>12)&0x3F])
		if remaining > 1 {
			result.WriteByte(base64Table[(n>>6)&0x3F])
		} else {
			result.WriteByte('=')
		}
		if remaining > 2 {
			result.WriteByte(base64Table[n&0x3F])
		} else {
			result.WriteByte('=')
		}
	}

	return result.String()
}

// Ensure image module is registered with proper file handling
func init() {
	// Register decoders for common formats
	// PNG and JPEG are registered by default when importing image/png and image/jpeg
	// Additional formats can be registered here if needed
	_ = os.Getenv("") // Prevent unused import error
}
