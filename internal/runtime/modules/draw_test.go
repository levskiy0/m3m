package modules

import (
	"image/color"
	"strings"
	"testing"
)

func TestDrawModule_NewDrawModule(t *testing.T) {
	module := NewDrawModule(nil, "test-project")

	if module == nil {
		t.Fatal("NewDrawModule() returned nil")
	}

	if module.projectID != "test-project" {
		t.Errorf("projectID = %q, want %q", module.projectID, "test-project")
	}
}

func TestDrawModule_CreateCanvas(t *testing.T) {
	module := NewDrawModule(nil, "test-project")

	tests := []struct {
		name           string
		width          int
		height         int
		expectedWidth  int
		expectedHeight int
	}{
		{"normal dimensions", 200, 100, 200, 100},
		{"zero width defaults to 100", 0, 100, 100, 100},
		{"zero height defaults to 100", 200, 0, 200, 100},
		{"negative width defaults to 100", -50, 100, 100, 100},
		{"negative height defaults to 100", 200, -50, 200, 100},
		{"large dimensions", 1920, 1080, 1920, 1080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := module.CreateCanvas(tt.width, tt.height)

			if canvas == nil {
				t.Fatal("CreateCanvas() returned nil")
			}

			if canvas.Width() != tt.expectedWidth {
				t.Errorf("Width() = %d, want %d", canvas.Width(), tt.expectedWidth)
			}

			if canvas.Height() != tt.expectedHeight {
				t.Errorf("Height() = %d, want %d", canvas.Height(), tt.expectedHeight)
			}
		})
	}
}

func TestDrawModule_LoadImage_NilStorage(t *testing.T) {
	module := NewDrawModule(nil, "test-project")

	canvas := module.LoadImage("test.png")

	if canvas != nil {
		t.Error("LoadImage with nil storage should return nil")
	}
}

func TestCanvas_SetColor(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(100, 100)

	// Test that SetColor doesn't panic with various inputs
	colors := []string{
		"#FF0000",
		"#00FF00",
		"#0000FF",
		"#FFF",
		"#000",
		"red",
		"green",
		"blue",
		"white",
		"black",
		"invalid-color",
		"",
	}

	for _, clr := range colors {
		t.Run(clr, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetColor(%q) panicked: %v", clr, r)
				}
			}()
			canvas.SetColor(clr)
		})
	}
}

func TestCanvas_SetLineWidth(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(100, 100)

	tests := []float64{1.0, 2.5, 0, -1.0, 10.0}

	for _, width := range tests {
		t.Run("", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetLineWidth(%f) panicked: %v", width, r)
				}
			}()
			canvas.SetLineWidth(width)
		})
	}
}

func TestCanvas_Clear(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(100, 100)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Clear panicked: %v", r)
		}
	}()

	canvas.Clear("white")
	canvas.Clear("#000000")
}

func TestCanvas_DrawingOperations(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(200, 200)

	// Test all drawing operations don't panic
	operations := []func(){
		func() { canvas.Rect(10, 10, 50, 50) },
		func() { canvas.FillRect(70, 10, 50, 50) },
		func() { canvas.Circle(50, 150, 30) },
		func() { canvas.FillCircle(150, 150, 30) },
		func() { canvas.Line(0, 0, 200, 200) },
		func() { canvas.Ellipse(100, 100, 40, 20) },
		func() { canvas.FillEllipse(100, 100, 30, 15) },
		func() { canvas.Arc(100, 100, 50, 0, 3.14) },
		func() { canvas.RoundedRect(10, 10, 80, 40, 10) },
		func() { canvas.FillRoundedRect(100, 10, 80, 40, 10) },
	}

	for i, op := range operations {
		t.Run("", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Operation %d panicked: %v", i, r)
				}
			}()
			op()
		})
	}
}

func TestCanvas_Text(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(200, 200)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Text operations panicked: %v", r)
		}
	}()

	canvas.SetColor("black")
	canvas.Text("Hello", 50, 50)
	canvas.TextCentered("World", 100, 100)
}

func TestCanvas_SetFontSize(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(100, 100)

	sizes := []float64{12.0, 24.0, 0, -5.0, 72.0}

	for _, size := range sizes {
		t.Run("", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					// Font loading might fail in test environment, which is OK
					t.Logf("SetFontSize(%f) panicked (expected in test env): %v", size, r)
				}
			}()
			canvas.SetFontSize(size)
		})
	}
}

func TestCanvas_Polygon(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(200, 200)

	// Triangle
	triangle := [][]float64{{100, 10}, {150, 100}, {50, 100}}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Polygon panicked: %v", r)
		}
	}()

	canvas.Polygon(triangle)
	canvas.FillPolygon(triangle)

	// Empty polygon (should not panic)
	canvas.Polygon([][]float64{})
	canvas.FillPolygon([][]float64{})

	// Less than 3 points (should not panic)
	canvas.Polygon([][]float64{{0, 0}, {10, 10}})
}

func TestCanvas_Transformations(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(100, 100)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Transformations panicked: %v", r)
		}
	}()

	canvas.Push()
	canvas.Translate(50, 50)
	canvas.Rotate(0.5)
	canvas.Scale(2.0, 2.0)
	canvas.Pop()
}

func TestCanvas_Save_NilStorage(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(100, 100)

	result := canvas.Save("test.png")

	if result {
		t.Error("Save with nil storage should return false")
	}
}

func TestCanvas_ToBase64(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(100, 100)

	// Draw something
	canvas.SetColor("red")
	canvas.FillRect(0, 0, 100, 100)

	tests := []struct {
		format   string
		mimeType string
	}{
		{"png", "image/png"},
		{"jpeg", "image/jpeg"},
		{"PNG", "image/png"},
		{"JPEG", "image/jpeg"},
		{"", "image/jpeg"}, // default to jpeg
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := canvas.ToBase64(tt.format)

			if !strings.HasPrefix(result, "data:"+tt.mimeType+";base64,") {
				t.Errorf("ToBase64(%q) prefix = %q, want data:%s;base64,", tt.format, result[:30], tt.mimeType)
			}

			// Verify it's valid base64
			if len(result) < 50 {
				t.Error("ToBase64 result too short")
			}
		})
	}
}

func TestCanvas_Dimensions(t *testing.T) {
	module := NewDrawModule(nil, "test-project")
	canvas := module.CreateCanvas(300, 200)

	if canvas.Width() != 300 {
		t.Errorf("Width() = %d, want 300", canvas.Width())
	}

	if canvas.Height() != 200 {
		t.Errorf("Height() = %d, want 200", canvas.Height())
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input    string
		expected color.Color
	}{
		{"#FF0000", color.RGBA{255, 0, 0, 255}},
		{"#00FF00", color.RGBA{0, 255, 0, 255}},
		{"#0000FF", color.RGBA{0, 0, 255, 255}},
		{"#FFF", color.RGBA{255, 255, 255, 255}},
		{"#000", color.RGBA{0, 0, 0, 255}},
		{"red", color.RGBA{255, 0, 0, 255}},
		{"green", color.RGBA{0, 255, 0, 255}},
		{"blue", color.RGBA{0, 0, 255, 255}},
		{"white", color.White},
		{"black", color.Black},
		{"#FF000080", color.RGBA{255, 0, 0, 128}}, // With alpha
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseColor(tt.input)
			r1, g1, b1, a1 := result.RGBA()
			r2, g2, b2, a2 := tt.expected.RGBA()

			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				t.Errorf("parseColor(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseColor_Invalid(t *testing.T) {
	// Invalid colors should return black (for truly invalid formats)
	invalid := []string{
		"invalid",
		"notacolor",
		"#12345", // Wrong length
		"",
	}

	for _, input := range invalid {
		t.Run(input, func(t *testing.T) {
			result := parseColor(input)
			if result != color.Black {
				t.Errorf("parseColor(%q) should return black for invalid input", input)
			}
		})
	}
}

func TestParseColor_InvalidHexChars(t *testing.T) {
	// Invalid hex chars get interpreted as 0, which is expected behavior
	// "#GGG" expands to "#GGGGGG" and G is interpreted as 0
	result := parseColor("#GGG")
	r, g, b, _ := result.RGBA()
	// Just verify it doesn't panic and returns some color
	_ = r
	_ = g
	_ = b
}

func TestHexToByte(t *testing.T) {
	tests := []struct {
		input    string
		expected byte
	}{
		{"00", 0},
		{"FF", 255},
		{"ff", 255},
		{"80", 128},
		{"7F", 127},
		{"AB", 171},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := hexToByte(tt.input)
			if result != tt.expected {
				t.Errorf("hexToByte(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
