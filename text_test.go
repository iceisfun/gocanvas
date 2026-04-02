package gocanvas

import (
	"image/color"
	"math"
	"testing"
)

func TestMeasureText(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(16)
	if err != nil {
		t.Fatal(err)
	}

	c := New(200, 50)
	c.SetFont(face)

	m := c.MeasureText("Hello")
	if m.Width <= 0 {
		t.Errorf("MeasureText width = %v, want > 0", m.Width)
	}
	if m.Height <= 0 {
		t.Errorf("MeasureText height = %v, want > 0", m.Height)
	}
	if m.Ascent <= 0 {
		t.Errorf("MeasureText ascent = %v, want > 0", m.Ascent)
	}
}

func TestMeasureTextNoFont(t *testing.T) {
	c := New(500, 500)
	m := c.MeasureText("Hello")
	if m.Width != 0 || m.Height != 0 {
		t.Error("expected zero metrics with no font set")
	}
}

func TestFillText(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(20)
	if err != nil {
		t.Fatal(err)
	}

	c := New(200, 50)
	c.SetFont(face)
	c.SetFillColor(RGB(0, 0, 0))
	c.FillText("Hi", 10, 30)

	// Check that some pixels near the text position are not white.
	hasNonWhite := false
	for y := 15; y < 40; y++ {
		for x := 10; x < 60; x++ {
			px := c.Image().RGBAAt(x, y)
			if px != (color.RGBA{255, 255, 255, 255}) {
				hasNonWhite = true
				break
			}
		}
		if hasNonWhite {
			break
		}
	}
	if !hasNonWhite {
		t.Error("FillText: no visible pixels rendered")
	}
}

func TestStrokeText(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(30)
	if err != nil {
		t.Fatal(err)
	}

	c := New(200, 50)
	c.SetFont(face)
	c.SetStrokeColor(RGB(255, 0, 0))
	c.StrokeText("A", 10, 35)

	// Should have some red pixels.
	hasRed := false
	for y := 5; y < 45; y++ {
		for x := 5; x < 50; x++ {
			px := c.Image().RGBAAt(x, y)
			if px.R > 100 && px.G < 50 {
				hasRed = true
				break
			}
		}
		if hasRed {
			break
		}
	}
	if !hasRed {
		t.Error("StrokeText: no visible red pixels")
	}
}

func TestGoldenTextTransformed(t *testing.T) {
	f := loadTestFont(t)

	goldenTest(t, "text_transformed", 400, 350, func(c *Canvas) {
		face, err := f.NewFace(20)
		if err != nil {
			t.Fatal(err)
		}
		faceLg, err := f.NewFace(28)
		if err != nil {
			t.Fatal(err)
		}

		// Translated text.
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(0, 0, 0))
		c.Translate(20, 30)
		c.FillText("Translated", 0, 0)
		c.Restore()

		// Rotated text (30 degrees).
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(200, 0, 0))
		c.Translate(200, 60)
		c.Rotate(math.Pi / 6)
		c.FillText("Rotated 30", 0, 0)
		c.Restore()

		// Rotated text (90 degrees).
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(0, 0, 200))
		c.Translate(370, 30)
		c.Rotate(math.Pi / 2)
		c.FillText("Rot 90", 0, 0)
		c.Restore()

		// Scaled text (2x).
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(0, 150, 0))
		c.Translate(20, 100)
		c.Scale(2, 2)
		c.FillText("Big", 0, 0)
		c.Restore()

		// Rotated stroke text (45 degrees).
		c.Save()
		c.SetFont(faceLg)
		c.SetStrokeColor(RGB(180, 0, 180))
		c.Translate(200, 140)
		c.Rotate(math.Pi / 4)
		c.StrokeText("Outline", 0, 0)
		c.Restore()

		// Rotated + scaled text.
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(0, 100, 150))
		c.Translate(60, 200)
		c.Rotate(-math.Pi / 12)
		c.Scale(1.5, 1)
		c.FillText("Skewed", 0, 0)
		c.Restore()

		// Negative scale (mirrored text).
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(150, 100, 0))
		c.Translate(350, 200)
		c.Scale(-1, 1)
		c.FillText("Mirror", 0, 0)
		c.Restore()
	})
}

func TestGoldenTextShadow(t *testing.T) {
	f := loadTestFont(t)

	goldenTest(t, "text_shadow", 400, 300, func(c *Canvas) {
		face, err := f.NewFace(22)
		if err != nil {
			t.Fatal(err)
		}
		faceLg, err := f.NewFace(30)
		if err != nil {
			t.Fatal(err)
		}

		// Basic shadow.
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(0, 0, 0))
		c.SetShadowColor(color.RGBA{0, 0, 0, 150})
		c.SetShadowBlur(3)
		c.SetShadowOffset(4, 4)
		c.FillText("Shadow", 20, 35)
		c.Restore()

		// Red text with blue shadow.
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(200, 0, 0))
		c.SetShadowColor(color.RGBA{0, 0, 200, 180})
		c.SetShadowBlur(4)
		c.SetShadowOffset(3, 5)
		c.FillText("Color shadow", 200, 35)
		c.Restore()

		// Sharp shadow (no blur).
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(0, 0, 0))
		c.SetShadowColor(color.RGBA{255, 0, 0, 200})
		c.SetShadowBlur(0)
		c.SetShadowOffset(3, 3)
		c.FillText("Sharp shadow", 20, 90)
		c.Restore()

		// Stroked text with shadow.
		c.Save()
		c.SetFont(faceLg)
		c.SetStrokeColor(RGB(0, 0, 180))
		c.SetShadowColor(color.RGBA{0, 0, 0, 128})
		c.SetShadowBlur(5)
		c.SetShadowOffset(4, 4)
		c.StrokeText("Outline", 20, 150)
		c.Restore()

		// Rotated text with shadow.
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(0, 120, 0))
		c.SetShadowColor(color.RGBA{0, 0, 0, 160})
		c.SetShadowBlur(3)
		c.SetShadowOffset(4, 4)
		c.Translate(200, 140)
		c.Rotate(math.Pi / 6)
		c.FillText("Rotated+shadow", 0, 0)
		c.Restore()

		// Large blurry shadow.
		c.Save()
		c.SetFont(faceLg)
		c.SetFillColor(RGB(0, 0, 0))
		c.SetShadowColor(color.RGBA{0, 0, 0, 100})
		c.SetShadowBlur(8)
		c.SetShadowOffset(6, 6)
		c.FillText("Glow", 20, 240)
		c.Restore()

		// Scaled text with shadow.
		c.Save()
		c.SetFont(face)
		c.SetFillColor(RGB(100, 0, 150))
		c.SetShadowColor(color.RGBA{0, 0, 0, 140})
		c.SetShadowBlur(2)
		c.SetShadowOffset(3, 3)
		c.Translate(200, 240)
		c.Scale(1.5, 1.5)
		c.FillText("Scaled", 0, 0)
		c.Restore()
	})
}

func TestFitText(t *testing.T) {
	f := loadTestFont(t)

	c := New(200, 50)
	face, err := c.FitText("Hello World", 100, 30, f, 6, 72)
	if err != nil {
		t.Fatal(err)
	}

	c.SetFont(face)
	m := c.MeasureText("Hello World")

	if m.Width > 100 {
		t.Errorf("FitText: width %v exceeds maxWidth 100", m.Width)
	}
	if m.Height > 30 {
		t.Errorf("FitText: height %v exceeds maxHeight 30", m.Height)
	}
	if face.Size() < 6 {
		t.Errorf("FitText: size %v below minimum", face.Size())
	}
}

func TestFillTextFit(t *testing.T) {
	f := loadTestFont(t)

	t.Run("fits within box", func(t *testing.T) {
		c := New(200, 60)
		c.SetFillColor(RGB(0, 0, 0))
		if err := c.FillTextFit("Hello", 10, 10, 180, 40, f); err != nil {
			t.Fatal(err)
		}

		// Text should be rendered inside the box.
		pixels := countNonWhite(c, 10, 190, 10, 50)
		if pixels == 0 {
			t.Error("FillTextFit: no visible pixels inside box")
		}

		// Should not extend above or below the box (with some tolerance for antialiasing).
		above := countNonWhite(c, 0, 200, 0, 8)
		if above > 0 {
			t.Errorf("FillTextFit: %d pixels above box", above)
		}
	})

	t.Run("short text gets large", func(t *testing.T) {
		cSmallBox := New(200, 30)
		cSmallBox.SetFillColor(RGB(0, 0, 0))
		cSmallBox.FillTextFit("Hi", 10, 5, 180, 20, f)

		cLargeBox := New(200, 100)
		cLargeBox.SetFillColor(RGB(0, 0, 0))
		cLargeBox.FillTextFit("Hi", 10, 5, 180, 90, f)

		smallMaxY := maxNonWhiteY(cSmallBox)
		largeMaxY := maxNonWhiteY(cLargeBox)

		if largeMaxY <= smallMaxY {
			t.Errorf("larger box should produce taller text: small maxY=%d, large maxY=%d", smallMaxY, largeMaxY)
		}
	})

	t.Run("long text constrained by width", func(t *testing.T) {
		c := New(300, 200)
		c.SetFillColor(RGB(0, 0, 0))
		c.FillTextFit("This is a longer string of text", 10, 10, 200, 180, f)

		// Text should not exceed the box width.
		rightPixels := countNonWhite(c, 215, 300, 0, 200)
		if rightPixels > 0 {
			t.Errorf("FillTextFit: %d pixels beyond box right edge", rightPixels)
		}
	})
}

func TestGoldenFillTextFit(t *testing.T) {
	f := loadTestFont(t)

	goldenTest(t, "text_fit", 400, 300, func(c *Canvas) {
		// Draw boxes and fit text into them.
		c.SetStrokeColor(RGB(200, 200, 200))
		c.SetLineWidth(1)

		// Small box.
		c.StrokeRect(10, 10, 120, 30)
		c.SetFillColor(RGB(0, 0, 0))
		c.FillTextFit("Hello", 10, 10, 120, 30, f)

		// Wide box.
		c.SetStrokeColor(RGB(200, 200, 200))
		c.StrokeRect(10, 60, 380, 40)
		c.SetFillColor(RGB(200, 0, 0))
		c.FillTextFit("Wide box", 10, 60, 380, 40, f)

		// Tall narrow box — width should constrain.
		c.SetStrokeColor(RGB(200, 200, 200))
		c.StrokeRect(10, 120, 60, 160)
		c.SetFillColor(RGB(0, 0, 200))
		c.FillTextFit("Narrow", 10, 120, 60, 160, f)

		// Large square box.
		c.SetStrokeColor(RGB(200, 200, 200))
		c.StrokeRect(100, 120, 160, 160)
		c.SetFillColor(RGB(0, 120, 0))
		c.FillTextFit("Big", 100, 120, 160, 160, f)

		// Long text in a wide box.
		c.SetStrokeColor(RGB(200, 200, 200))
		c.StrokeRect(280, 120, 110, 50)
		c.SetFillColor(RGB(100, 0, 100))
		c.FillTextFit("Longer text here", 280, 120, 110, 50, f)
	})
}

func TestWordWrapBasic(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(16)
	if err != nil {
		t.Fatal(err)
	}

	c := New(400, 200)
	c.SetFont(face)

	lines := c.WordWrap("The quick brown fox jumps over the lazy dog", 100)
	if len(lines) < 2 {
		t.Errorf("WordWrap: expected multiple lines, got %d: %v", len(lines), lines)
	}
	for i, line := range lines {
		if line == "" {
			t.Errorf("WordWrap: line %d is empty", i)
		}
	}
}

func TestWordWrapNewlines(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(16)
	if err != nil {
		t.Fatal(err)
	}

	c := New(400, 200)
	c.SetFont(face)

	lines := c.WordWrap("Hello\nWorld", 1000)
	if len(lines) != 2 {
		t.Errorf("WordWrap with newline: expected 2 lines, got %d: %v", len(lines), lines)
	}
	if len(lines) == 2 {
		if lines[0] != "Hello" || lines[1] != "World" {
			t.Errorf("WordWrap with newline: got %v, want [Hello World]", lines)
		}
	}
}

func TestWordWrapLongWord(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(16)
	if err != nil {
		t.Fatal(err)
	}

	c := New(400, 200)
	c.SetFont(face)

	lines := c.WordWrap("Supercalifragilisticexpialidocious", 50)
	if len(lines) != 1 {
		t.Errorf("WordWrap long word: expected 1 line, got %d: %v", len(lines), lines)
	}
	if len(lines) > 0 && lines[0] != "Supercalifragilisticexpialidocious" {
		t.Errorf("WordWrap long word: got %q, want original word", lines[0])
	}
}

func TestWordWrapNoFont(t *testing.T) {
	c := New(400, 200)
	lines := c.WordWrap("Hello World", 100)
	if lines != nil {
		t.Error("WordWrap with no font: expected nil")
	}
}

func TestMeasureTextWrapped(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(16)
	if err != nil {
		t.Fatal(err)
	}

	c := New(400, 200)
	c.SetFont(face)

	w, h := c.MeasureTextWrapped("The quick brown fox jumps over the lazy dog", 100, 1.2)
	if w <= 0 {
		t.Errorf("MeasureTextWrapped width = %v, want > 0", w)
	}
	if h <= 0 {
		t.Errorf("MeasureTextWrapped height = %v, want > 0", h)
	}
}

// maxNonWhiteX returns the largest x coordinate of a non-white pixel.
func maxNonWhiteX(c *Canvas) int {
	b := c.Image().Bounds()
	maxX := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if c.Image().RGBAAt(x, y) != (color.RGBA{255, 255, 255, 255}) {
				if x > maxX {
					maxX = x
				}
			}
		}
	}
	return maxX
}

// countNonWhite counts pixels that are not pure white in the given rectangle.
func countNonWhite(c *Canvas, xMin, xMax, yMin, yMax int) int {
	count := 0
	for y := yMin; y < yMax; y++ {
		for x := xMin; x < xMax; x++ {
			if c.Image().RGBAAt(x, y) != (color.RGBA{255, 255, 255, 255}) {
				count++
			}
		}
	}
	return count
}

// maxNonWhiteY returns the largest y coordinate of a non-white pixel.
func maxNonWhiteY(c *Canvas) int {
	b := c.Image().Bounds()
	maxY := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if c.Image().RGBAAt(x, y) != (color.RGBA{255, 255, 255, 255}) {
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	return maxY
}
