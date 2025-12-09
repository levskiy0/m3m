package modules

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/dop251/goja"
	"github.com/fogleman/gg"

	"github.com/levskiy0/m3m/internal/service"
	"github.com/levskiy0/m3m/pkg/schema"
)

// Canvas represents a drawing canvas in JS
type Canvas struct {
	ctx       *gg.Context
	width     int
	height    int
	projectID string
	storage   *service.StorageService
}

// DrawModule provides canvas drawing operations
type DrawModule struct {
	storage   *service.StorageService
	projectID string
}

// NewDrawModule creates a new draw module
func NewDrawModule(storage *service.StorageService, projectID string) *DrawModule {
	return &DrawModule{
		storage:   storage,
		projectID: projectID,
	}
}

// Name returns the module name for JavaScript
func (d *DrawModule) Name() string {
	return "$draw"
}

// Register registers the module into the JavaScript VM
func (d *DrawModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(d.Name(), map[string]interface{}{
		"createCanvas": d.CreateCanvas,
		"loadImage":    d.LoadImage,
	})
}

// CreateCanvas creates a new canvas with the specified dimensions
func (d *DrawModule) CreateCanvas(width, height int) *Canvas {
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 100
	}

	ctx := gg.NewContext(width, height)

	return &Canvas{
		ctx:       ctx,
		width:     width,
		height:    height,
		projectID: d.projectID,
		storage:   d.storage,
	}
}

// LoadImage loads an image from storage and creates a canvas from it
func (d *DrawModule) LoadImage(path string) *Canvas {
	if d.storage == nil {
		return nil
	}

	data, err := d.storage.Read(d.projectID, path)
	if err != nil {
		return nil
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	bounds := img.Bounds()
	ctx := gg.NewContextForImage(img)

	return &Canvas{
		ctx:       ctx,
		width:     bounds.Dx(),
		height:    bounds.Dy(),
		projectID: d.projectID,
		storage:   d.storage,
	}
}

// SetColor sets the drawing color (hex string like "#FF0000" or "red")
func (c *Canvas) SetColor(colorStr string) {
	clr := parseColor(colorStr)
	c.ctx.SetColor(clr)
}

// SetLineWidth sets the line width for drawing
func (c *Canvas) SetLineWidth(width float64) {
	if width <= 0 {
		width = 1
	}
	c.ctx.SetLineWidth(width)
}

// Clear clears the canvas with the specified color
func (c *Canvas) Clear(colorStr string) {
	clr := parseColor(colorStr)
	c.ctx.SetColor(clr)
	c.ctx.Clear()
}

// Rect draws a rectangle outline
func (c *Canvas) Rect(x, y, width, height float64) {
	c.ctx.DrawRectangle(x, y, width, height)
	c.ctx.Stroke()
}

// FillRect draws a filled rectangle
func (c *Canvas) FillRect(x, y, width, height float64) {
	c.ctx.DrawRectangle(x, y, width, height)
	c.ctx.Fill()
}

// Circle draws a circle outline
func (c *Canvas) Circle(x, y, radius float64) {
	c.ctx.DrawCircle(x, y, radius)
	c.ctx.Stroke()
}

// FillCircle draws a filled circle
func (c *Canvas) FillCircle(x, y, radius float64) {
	c.ctx.DrawCircle(x, y, radius)
	c.ctx.Fill()
}

// Line draws a line between two points
func (c *Canvas) Line(x1, y1, x2, y2 float64) {
	c.ctx.DrawLine(x1, y1, x2, y2)
	c.ctx.Stroke()
}

// Ellipse draws an ellipse outline
func (c *Canvas) Ellipse(x, y, rx, ry float64) {
	c.ctx.DrawEllipse(x, y, rx, ry)
	c.ctx.Stroke()
}

// FillEllipse draws a filled ellipse
func (c *Canvas) FillEllipse(x, y, rx, ry float64) {
	c.ctx.DrawEllipse(x, y, rx, ry)
	c.ctx.Fill()
}

// Arc draws an arc
func (c *Canvas) Arc(x, y, r, angle1, angle2 float64) {
	c.ctx.DrawArc(x, y, r, angle1, angle2)
	c.ctx.Stroke()
}

// Text draws text at the specified position
func (c *Canvas) Text(text string, x, y float64) {
	c.ctx.DrawString(text, x, y)
}

// TextCentered draws text centered at the specified position
func (c *Canvas) TextCentered(text string, x, y float64) {
	c.ctx.DrawStringAnchored(text, x, y, 0.5, 0.5)
}

// SetFontSize sets the font size
func (c *Canvas) SetFontSize(size float64) {
	if size <= 0 {
		size = 12
	}
	// Use built-in font
	c.ctx.LoadFontFace("/System/Library/Fonts/Helvetica.ttc", size)
}

// RoundedRect draws a rounded rectangle outline
func (c *Canvas) RoundedRect(x, y, w, h, r float64) {
	c.ctx.DrawRoundedRectangle(x, y, w, h, r)
	c.ctx.Stroke()
}

// FillRoundedRect draws a filled rounded rectangle
func (c *Canvas) FillRoundedRect(x, y, w, h, r float64) {
	c.ctx.DrawRoundedRectangle(x, y, w, h, r)
	c.ctx.Fill()
}

// Polygon draws a polygon from points
func (c *Canvas) Polygon(points [][]float64) {
	if len(points) < 3 {
		return
	}

	c.ctx.MoveTo(points[0][0], points[0][1])
	for i := 1; i < len(points); i++ {
		if len(points[i]) >= 2 {
			c.ctx.LineTo(points[i][0], points[i][1])
		}
	}
	c.ctx.ClosePath()
	c.ctx.Stroke()
}

// FillPolygon draws a filled polygon from points
func (c *Canvas) FillPolygon(points [][]float64) {
	if len(points) < 3 {
		return
	}

	c.ctx.MoveTo(points[0][0], points[0][1])
	for i := 1; i < len(points); i++ {
		if len(points[i]) >= 2 {
			c.ctx.LineTo(points[i][0], points[i][1])
		}
	}
	c.ctx.ClosePath()
	c.ctx.Fill()
}

// Save saves the canvas to storage
func (c *Canvas) Save(path string) bool {
	if c.storage == nil {
		return false
	}

	var buf bytes.Buffer
	var err error

	// Determine format from extension
	lowerPath := strings.ToLower(path)
	if strings.HasSuffix(lowerPath, ".png") {
		err = png.Encode(&buf, c.ctx.Image())
	} else {
		err = jpeg.Encode(&buf, c.ctx.Image(), &jpeg.Options{Quality: 85})
	}

	if err != nil {
		return false
	}

	return c.storage.Write(c.projectID, path, buf.Bytes()) == nil
}

// ToBase64 returns the canvas as a base64 data URI
func (c *Canvas) ToBase64(format string) string {
	var buf bytes.Buffer
	var mimeType string

	format = strings.ToLower(format)
	if format == "png" {
		png.Encode(&buf, c.ctx.Image())
		mimeType = "image/png"
	} else {
		jpeg.Encode(&buf, c.ctx.Image(), &jpeg.Options{Quality: 85})
		mimeType = "image/jpeg"
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:" + mimeType + ";base64," + encoded
}

// Width returns the canvas width
func (c *Canvas) Width() int {
	return c.width
}

// Height returns the canvas height
func (c *Canvas) Height() int {
	return c.height
}

// Translate applies a translation transformation
func (c *Canvas) Translate(x, y float64) {
	c.ctx.Translate(x, y)
}

// Rotate applies a rotation transformation (in radians)
func (c *Canvas) Rotate(angle float64) {
	c.ctx.Rotate(angle)
}

// Scale applies a scale transformation
func (c *Canvas) Scale(sx, sy float64) {
	c.ctx.Scale(sx, sy)
}

// Push saves the current transformation state
func (c *Canvas) Push() {
	c.ctx.Push()
}

// Pop restores the previously saved transformation state
func (c *Canvas) Pop() {
	c.ctx.Pop()
}

// parseColor parses a color string to color.Color
func parseColor(s string) color.Color {
	s = strings.TrimSpace(strings.ToLower(s))

	// Named colors
	namedColors := map[string]color.Color{
		"black":   color.Black,
		"white":   color.White,
		"red":     color.RGBA{255, 0, 0, 255},
		"green":   color.RGBA{0, 255, 0, 255},
		"blue":    color.RGBA{0, 0, 255, 255},
		"yellow":  color.RGBA{255, 255, 0, 255},
		"cyan":    color.RGBA{0, 255, 255, 255},
		"magenta": color.RGBA{255, 0, 255, 255},
		"orange":  color.RGBA{255, 165, 0, 255},
		"purple":  color.RGBA{128, 0, 128, 255},
		"pink":    color.RGBA{255, 192, 203, 255},
		"brown":   color.RGBA{165, 42, 42, 255},
		"gray":    color.RGBA{128, 128, 128, 255},
		"grey":    color.RGBA{128, 128, 128, 255},
	}

	if clr, ok := namedColors[s]; ok {
		return clr
	}

	// Hex color
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	}

	// Support 3-char hex (#RGB)
	if len(s) == 3 {
		s = string(s[0]) + string(s[0]) + string(s[1]) + string(s[1]) + string(s[2]) + string(s[2])
	}

	// Parse 6-char hex (#RRGGBB)
	if len(s) == 6 {
		r := hexToByte(s[0:2])
		g := hexToByte(s[2:4])
		b := hexToByte(s[4:6])
		return color.RGBA{r, g, b, 255}
	}

	// Parse 8-char hex (#RRGGBBAA)
	if len(s) == 8 {
		r := hexToByte(s[0:2])
		g := hexToByte(s[2:4])
		b := hexToByte(s[4:6])
		a := hexToByte(s[6:8])
		return color.RGBA{r, g, b, a}
	}

	// Default to black
	return color.Black
}

// hexToByte converts a hex string to byte
func hexToByte(s string) byte {
	var result byte
	for _, c := range s {
		result *= 16
		if c >= '0' && c <= '9' {
			result += byte(c - '0')
		} else if c >= 'a' && c <= 'f' {
			result += byte(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			result += byte(c - 'A' + 10)
		}
	}
	return result
}

// GetSchema implements JSSchemaProvider
func (d *DrawModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$draw",
		Description: "Canvas-based drawing and graphics creation",
		Types: []schema.TypeSchema{
			{
				Name:        "Canvas",
				Description: "A drawing canvas with various drawing methods",
				Fields: []schema.ParamSchema{
					{Name: "setColor", Type: "(color: string) => void", Description: "Set drawing color (hex or named)"},
					{Name: "setLineWidth", Type: "(width: number) => void", Description: "Set line width"},
					{Name: "clear", Type: "(color: string) => void", Description: "Clear canvas with color"},
					{Name: "rect", Type: "(x: number, y: number, w: number, h: number) => void", Description: "Draw rectangle outline"},
					{Name: "fillRect", Type: "(x: number, y: number, w: number, h: number) => void", Description: "Draw filled rectangle"},
					{Name: "circle", Type: "(x: number, y: number, r: number) => void", Description: "Draw circle outline"},
					{Name: "fillCircle", Type: "(x: number, y: number, r: number) => void", Description: "Draw filled circle"},
					{Name: "line", Type: "(x1: number, y1: number, x2: number, y2: number) => void", Description: "Draw a line"},
					{Name: "ellipse", Type: "(x: number, y: number, rx: number, ry: number) => void", Description: "Draw ellipse outline"},
					{Name: "fillEllipse", Type: "(x: number, y: number, rx: number, ry: number) => void", Description: "Draw filled ellipse"},
					{Name: "arc", Type: "(x: number, y: number, r: number, a1: number, a2: number) => void", Description: "Draw an arc"},
					{Name: "text", Type: "(text: string, x: number, y: number) => void", Description: "Draw text"},
					{Name: "textCentered", Type: "(text: string, x: number, y: number) => void", Description: "Draw centered text"},
					{Name: "setFontSize", Type: "(size: number) => void", Description: "Set font size"},
					{Name: "roundedRect", Type: "(x: number, y: number, w: number, h: number, r: number) => void", Description: "Draw rounded rectangle outline"},
					{Name: "fillRoundedRect", Type: "(x: number, y: number, w: number, h: number, r: number) => void", Description: "Draw filled rounded rectangle"},
					{Name: "polygon", Type: "(points: number[][]) => void", Description: "Draw polygon outline"},
					{Name: "fillPolygon", Type: "(points: number[][]) => void", Description: "Draw filled polygon"},
					{Name: "save", Type: "(path: string) => boolean", Description: "Save canvas to storage"},
					{Name: "toBase64", Type: "(format: string) => string", Description: "Export as base64 data URI"},
					{Name: "width", Type: "() => number", Description: "Get canvas width"},
					{Name: "height", Type: "() => number", Description: "Get canvas height"},
					{Name: "translate", Type: "(x: number, y: number) => void", Description: "Apply translation"},
					{Name: "rotate", Type: "(angle: number) => void", Description: "Apply rotation (radians)"},
					{Name: "scale", Type: "(sx: number, sy: number) => void", Description: "Apply scale"},
					{Name: "push", Type: "() => void", Description: "Save transformation state"},
					{Name: "pop", Type: "() => void", Description: "Restore transformation state"},
				},
			},
		},
		Methods: []schema.MethodSchema{
			{
				Name:        "createCanvas",
				Description: "Create a new drawing canvas",
				Params: []schema.ParamSchema{
					{Name: "width", Type: "number", Description: "Canvas width in pixels"},
					{Name: "height", Type: "number", Description: "Canvas height in pixels"},
				},
				Returns: &schema.ParamSchema{Type: "Canvas"},
			},
			{
				Name:        "loadImage",
				Description: "Load an image from storage as a canvas",
				Params:      []schema.ParamSchema{{Name: "path", Type: "string", Description: "Path to image in storage"}},
				Returns:     &schema.ParamSchema{Type: "Canvas | null"},
			},
		},
	}
}

// GetDrawSchema returns the draw schema (static version)
func GetDrawSchema() schema.ModuleSchema {
	return (&DrawModule{}).GetSchema()
}
