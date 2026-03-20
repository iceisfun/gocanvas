package gocanvas

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func goldenTest(t *testing.T, name string, width, height int, draw func(c *Canvas)) {
	t.Helper()

	c := New(width, height)
	draw(c)

	goldenPath := filepath.Join("testdata", name+".png")

	if *update {
		if err := c.SavePNG(goldenPath); err != nil {
			t.Fatalf("saving golden file: %v", err)
		}
		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	f, err := os.Open(goldenPath)
	if err != nil {
		// Generate golden file on first run.
		if os.IsNotExist(err) {
			if err := c.SavePNG(goldenPath); err != nil {
				t.Fatalf("saving golden file: %v", err)
			}
			t.Logf("generated golden file: %s (re-run to validate)", goldenPath)
			return
		}
		t.Fatalf("opening golden file: %v", err)
	}
	defer f.Close()

	golden, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decoding golden PNG: %v", err)
	}

	got := c.Image()
	goldenRGBA, ok := golden.(*image.RGBA)
	if !ok {
		// Convert to RGBA for comparison.
		b := golden.Bounds()
		goldenRGBA = image.NewRGBA(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				goldenRGBA.Set(x, y, golden.At(x, y))
			}
		}
	}

	if got.Bounds() != goldenRGBA.Bounds() {
		t.Fatalf("bounds mismatch: got %v, golden %v", got.Bounds(), goldenRGBA.Bounds())
	}

	diffCount := 0
	for i := range got.Pix {
		if got.Pix[i] != goldenRGBA.Pix[i] {
			diffCount++
		}
	}
	if diffCount > 0 {
		t.Errorf("%s: %d pixel component differences", name, diffCount)
		// Save actual for debugging.
		actualPath := filepath.Join("testdata", name+"_actual.png")
		_ = c.SavePNG(actualPath)
		t.Logf("saved actual output to %s", actualPath)
	}
}

func TestGoldenBasicShapes(t *testing.T) {
	goldenTest(t, "basic_shapes", 200, 200, func(c *Canvas) {
		// Red rectangle.
		c.SetFillColor(RGB(255, 0, 0))
		c.FillRect(10, 10, 60, 40)

		// Green circle.
		c.SetFillColor(RGB(0, 200, 0))
		c.BeginPath()
		c.path.Circle(150, 50, 30)
		c.Fill()

		// Blue stroked rectangle.
		c.SetStrokeColor(RGB(0, 0, 255))
		c.SetLineWidth(3)
		c.StrokeRect(10, 100, 80, 50)

		// Yellow filled ellipse.
		c.SetFillColor(RGB(255, 255, 0))
		c.BeginPath()
		c.path.Ellipse(150, 140, 40, 25)
		c.Fill()
	})
}

func TestGoldenTransforms(t *testing.T) {
	goldenTest(t, "transforms", 400, 500, func(c *Canvas) {
		// --- Row 1: Basic translate, scale, rotate ---

		// Translate.
		c.Save()
		c.Translate(20, 20)
		c.SetFillColor(RGB(200, 0, 0))
		c.FillRect(0, 0, 40, 30)
		c.Restore()

		// Uniform scale.
		c.Save()
		c.Translate(120, 35)
		c.Scale(2, 2)
		c.SetFillColor(RGB(0, 0, 200))
		c.FillRect(-10, -10, 20, 20)
		c.Restore()

		// Non-uniform scale (stretch).
		c.Save()
		c.Translate(220, 35)
		c.Scale(3, 1)
		c.SetFillColor(RGB(0, 150, 200))
		c.FillRect(-10, -10, 20, 20)
		c.Restore()

		// Rotate 30 degrees.
		c.Save()
		c.Translate(350, 35)
		c.Rotate(math.Pi / 6)
		c.SetFillColor(RGB(0, 180, 0))
		c.FillRect(-20, -12, 40, 24)
		c.Restore()

		// --- Row 2: Skew, negative scale (mirror), combined ---

		// Skew X.
		c.Save()
		c.Translate(50, 120)
		c.Transform(SkewMatrix(math.Pi/6, 0))
		c.SetFillColor(RGB(200, 100, 0))
		c.FillRect(-15, -15, 30, 30)
		c.Restore()

		// Skew Y.
		c.Save()
		c.Translate(160, 120)
		c.Transform(SkewMatrix(0, math.Pi/8))
		c.SetFillColor(RGB(200, 0, 200))
		c.FillRect(-15, -15, 30, 30)
		c.Restore()

		// Negative scale X (horizontal mirror).
		c.Save()
		c.Translate(270, 120)
		c.Scale(-1, 1)
		c.SetFillColor(RGB(100, 100, 0))
		// Draw an L-shape to make mirroring visible.
		c.BeginPath()
		c.MoveTo(0, -20)
		c.LineTo(25, -20)
		c.LineTo(25, -10)
		c.LineTo(10, -10)
		c.LineTo(10, 20)
		c.LineTo(0, 20)
		c.ClosePath()
		c.Fill()
		c.Restore()

		// Combined translate + rotate.
		c.Save()
		c.Translate(360, 120)
		c.Rotate(math.Pi / 4)
		c.SetFillColor(RGB(0, 100, 180))
		c.FillRect(-15, -15, 30, 30)
		c.Restore()

		// --- Row 3: Nested save/restore, SetTransform, ResetTransform ---

		// Nested transforms via save/restore.
		c.Save()
		c.Translate(50, 220)
		c.SetFillColor(RGB(180, 180, 180))
		c.FillRect(-20, -20, 40, 40) // Outer gray square.

		c.Save()
		c.Rotate(math.Pi / 4)
		c.SetFillColor(RGB(255, 80, 80))
		c.FillRect(-10, -10, 20, 20) // Inner rotated red square.
		c.Restore()
		c.Restore()

		// SetTransform: replace transform entirely.
		c.Save()
		c.Translate(999, 999) // This should be overridden.
		c.SetTransform(TranslateMatrix(160, 220).Multiply(RotateMatrix(math.Pi / 6)))
		c.SetFillColor(RGB(0, 200, 100))
		c.FillRect(-20, -10, 40, 20)
		c.Restore()

		// ResetTransform: draw at identity after transforms.
		c.Save()
		c.Translate(500, 500)
		c.Rotate(1.0)
		c.ResetTransform()
		c.SetFillColor(RGB(80, 80, 255))
		c.FillRect(270, 200, 40, 40) // Should appear at literal coords.
		c.Restore()

		// --- Row 4: Transforms on stroked paths and arcs ---

		// Scaled stroke.
		c.Save()
		c.Translate(50, 310)
		c.Scale(2, 1)
		c.SetStrokeColor(RGB(200, 0, 0))
		c.SetLineWidth(2)
		c.StrokeRect(-15, -15, 30, 30)
		c.Restore()

		// Rotated stroked triangle.
		c.Save()
		c.Translate(170, 310)
		c.Rotate(math.Pi / 5)
		c.SetStrokeColor(RGB(0, 0, 200))
		c.SetLineWidth(3)
		c.BeginPath()
		c.MoveTo(0, -20)
		c.LineTo(20, 15)
		c.LineTo(-20, 15)
		c.ClosePath()
		c.Stroke()
		c.Restore()

		// Translated and scaled arc (circle).
		c.Save()
		c.Translate(280, 310)
		c.Scale(1.5, 0.75)
		c.SetFillColor(RGB(255, 200, 0))
		c.BeginPath()
		c.Arc(0, 0, 20, 0, 2*math.Pi)
		c.ClosePath()
		c.Fill()
		c.Restore()

		// Rotated arc with stroke.
		c.Save()
		c.Translate(370, 310)
		c.Rotate(math.Pi / 3)
		c.SetStrokeColor(RGB(0, 150, 0))
		c.SetLineWidth(2)
		c.BeginPath()
		c.Arc(0, 0, 18, 0, math.Pi)
		c.Stroke()
		c.Restore()

		// --- Row 5: Composed transforms and complex paths ---

		// Scale then rotate then translate (all composed).
		c.Save()
		c.Translate(60, 420)
		c.Rotate(math.Pi / 8)
		c.Scale(1.5, 0.8)
		c.SetFillColor(RGB(100, 0, 150))
		c.FillRect(-20, -15, 40, 30)
		c.Restore()

		// Transform with bezier curve.
		c.Save()
		c.Translate(180, 420)
		c.Rotate(-math.Pi / 6)
		c.SetFillColor(RGB(0, 120, 120))
		c.BeginPath()
		c.MoveTo(-30, 0)
		c.BezierCurveTo(-30, -30, 30, -30, 30, 0)
		c.BezierCurveTo(30, 30, -30, 30, -30, 0)
		c.ClosePath()
		c.Fill()
		c.Restore()

		// Multiple sequential transforms (no save/restore).
		c.Save()
		c.Translate(320, 420)
		c.Scale(0.8, 0.8)
		c.Rotate(math.Pi / 12)
		c.Translate(0, -10)
		c.SetFillColor(RGB(200, 50, 50))
		c.FillRect(-15, -15, 30, 30)
		c.Restore()
	})
}

func TestGoldenStrokeStyles(t *testing.T) {
	goldenTest(t, "stroke_styles", 200, 200, func(c *Canvas) {
		c.SetStrokeColor(RGB(0, 0, 0))

		// Different line widths.
		widths := []float64{1, 2, 4, 8}
		for i, w := range widths {
			y := float64(20 + i*20)
			c.SetLineWidth(w)
			c.BeginPath()
			c.MoveTo(10, y)
			c.LineTo(190, y)
			c.Stroke()
		}

		// Different line caps.
		caps := []LineCap{CapButt, CapRound, CapSquare}
		c.SetLineWidth(6)
		c.SetStrokeColor(RGB(180, 0, 0))
		for i, cap := range caps {
			y := float64(120 + i*25)
			c.SetLineCap(cap)
			c.BeginPath()
			c.MoveTo(30, y)
			c.LineTo(170, y)
			c.Stroke()
		}
	})
}

func TestGoldenAlphaBlending(t *testing.T) {
	goldenTest(t, "alpha_blending", 200, 200, func(c *Canvas) {
		// Three overlapping semi-transparent circles.
		c.SetGlobalAlpha(0.5)

		c.SetFillColor(RGB(255, 0, 0))
		c.BeginPath()
		c.path.Circle(80, 80, 50)
		c.Fill()

		c.SetFillColor(RGB(0, 255, 0))
		c.BeginPath()
		c.path.Circle(120, 80, 50)
		c.Fill()

		c.SetFillColor(RGB(0, 0, 255))
		c.BeginPath()
		c.path.Circle(100, 120, 50)
		c.Fill()
	})
}

func TestGoldenComplexPath(t *testing.T) {
	goldenTest(t, "complex_path", 200, 200, func(c *Canvas) {
		// 5-pointed star.
		c.SetFillColor(RGB(255, 200, 0))
		c.SetStrokeColor(RGB(200, 100, 0))
		c.SetLineWidth(2)

		cx, cy := 100.0, 100.0
		outerR := 80.0
		innerR := 35.0

		c.BeginPath()
		for i := range 10 {
			angle := math.Pi/2 + float64(i)*math.Pi/5
			r := outerR
			if i%2 == 1 {
				r = innerR
			}
			x := cx + r*math.Cos(angle)
			y := cy - r*math.Sin(angle)
			if i == 0 {
				c.MoveTo(x, y)
			} else {
				c.LineTo(x, y)
			}
		}
		c.ClosePath()
		c.Fill()
		c.Stroke()
	})
}

func TestGoldenDashedLines(t *testing.T) {
	goldenTest(t, "dashed_lines", 200, 200, func(c *Canvas) {
		c.SetStrokeColor(RGB(0, 0, 0))
		c.SetLineWidth(2)

		// Dashed line.
		c.SetLineDash([]float64{10, 5})
		c.BeginPath()
		c.MoveTo(10, 20)
		c.LineTo(190, 20)
		c.Stroke()

		// Dotted line.
		c.SetLineDash([]float64{2, 4})
		c.BeginPath()
		c.MoveTo(10, 50)
		c.LineTo(190, 50)
		c.Stroke()

		// Dash-dot pattern.
		c.SetLineDash([]float64{10, 3, 2, 3})
		c.BeginPath()
		c.MoveTo(10, 80)
		c.LineTo(190, 80)
		c.Stroke()

		// Dashed rectangle.
		c.SetStrokeColor(RGB(0, 0, 200))
		c.SetLineDash([]float64{8, 4})
		c.SetLineWidth(2)
		c.StrokeRect(20, 110, 160, 70)
	})
}

func TestGoldenShadows(t *testing.T) {
	goldenTest(t, "shadows", 200, 200, func(c *Canvas) {
		// Shadow behind a red rectangle.
		c.SetShadowColor(color.RGBA{0, 0, 0, 128})
		c.SetShadowBlur(4)
		c.SetShadowOffset(4, 4)

		c.SetFillColor(RGB(220, 50, 50))
		c.FillRect(20, 20, 60, 40)

		// Shadow behind a blue circle.
		c.SetShadowColor(color.RGBA{0, 0, 100, 180})
		c.SetShadowBlur(6)
		c.SetShadowOffset(3, 5)

		c.SetFillColor(RGB(50, 50, 220))
		c.BeginPath()
		c.path.Circle(150, 60, 30)
		c.Fill()

		// No shadow for next shape.
		c.SetShadowColor(color.RGBA{})
		c.SetFillColor(RGB(50, 180, 50))
		c.FillRect(40, 120, 120, 50)
	})
}

func TestGoldenText(t *testing.T) {
	f := loadTestFont(t)

	goldenTest(t, "text_rendering", 300, 150, func(c *Canvas) {
		face, err := f.NewFace(24)
		if err != nil {
			t.Fatal(err)
		}
		c.SetFont(face)

		// Filled text.
		c.SetFillColor(RGB(0, 0, 0))
		c.FillText("Hello Canvas!", 10, 40)

		// Colored text.
		c.SetFillColor(RGB(200, 0, 0))
		face2, _ := f.NewFace(18)
		c.SetFont(face2)
		c.FillText("Red text at 18pt", 10, 80)

		// Stroked text.
		c.SetStrokeColor(RGB(0, 0, 200))
		face3, _ := f.NewFace(30)
		c.SetFont(face3)
		c.StrokeText("Outline", 10, 120)
	})
}

func TestGoldenAnnotations(t *testing.T) {
	f := loadTestFont(t)

	goldenTest(t, "annotations", 400, 300, func(c *Canvas) {
		style := DefaultAnnotStyle()
		style.Font = f

		// Draw bounding boxes with labels.
		DrawLabeledBox(c, "Person", 20, 40, 100, 150, style)

		style.StrokeColor = RGB(255, 0, 0)
		DrawLabeledBox(c, "Car", 200, 80, 150, 100, style)

		// Draw a polygon annotation.
		style.StrokeColor = RGB(0, 0, 255)
		style.FillColor = RGBA(0, 0, 255, 40)
		pts := []Point{{50, 250}, {150, 220}, {180, 280}, {80, 290}}
		DrawPolygon(c, pts, style)

		// Standalone label.
		style.FillColor = RGBA(0, 0, 0, 200)
		DrawLabel(c, "Score: 0.95", 200, 250, style)
	})
}
