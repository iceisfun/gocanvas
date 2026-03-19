package gocanvas

import (
	"flag"
	"image"
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
	goldenTest(t, "transforms", 200, 200, func(c *Canvas) {
		// Translated rectangle.
		c.Save()
		c.Translate(50, 50)
		c.SetFillColor(RGB(200, 0, 0))
		c.FillRect(0, 0, 40, 30)
		c.Restore()

		// Scaled rectangle.
		c.Save()
		c.Translate(150, 30)
		c.Scale(2, 1.5)
		c.SetFillColor(RGB(0, 0, 200))
		c.FillRect(-10, -10, 20, 20)
		c.Restore()

		// Rotated rectangle.
		c.Save()
		c.Translate(100, 150)
		c.Rotate(math.Pi / 6)
		c.SetFillColor(RGB(0, 180, 0))
		c.FillRect(-25, -15, 50, 30)
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
