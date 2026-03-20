package gocanvas

import (
	"image/color"
	"math"
	"testing"
)

func TestNewCanvas(t *testing.T) {
	c := New(100, 50)
	if c.Width() != 100 || c.Height() != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50", c.Width(), c.Height())
	}

	// Should be white.
	got := c.Image().RGBAAt(0, 0)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("initial pixel = %v, want white", got)
	}
}

func TestSaveRestore(t *testing.T) {
	c := New(10, 10)
	origFill := c.state.fill
	origMatrix := c.state.matrix

	c.Save()
	c.SetFillColor(RGB(255, 0, 0))
	c.Translate(10, 10)

	if c.state.fill == origFill {
		t.Error("fill should have changed after SetFillColor")
	}

	c.Restore()

	if c.state.fill != origFill {
		t.Errorf("fill after restore = %v, want %v", c.state.fill, origFill)
	}
	if c.state.matrix != origMatrix {
		t.Errorf("matrix after restore = %v, want %v", c.state.matrix, origMatrix)
	}
}

func TestRestoreEmptyStack(t *testing.T) {
	c := New(10, 10)
	origFill := c.state.fill
	c.Restore() // should be no-op
	if c.state.fill != origFill {
		t.Error("Restore on empty stack should be no-op")
	}
}

func TestFillRect(t *testing.T) {
	c := New(20, 20)
	c.SetFillColor(RGB(255, 0, 0))
	c.FillRect(5, 5, 10, 10)

	// Check a pixel inside.
	got := c.Image().RGBAAt(10, 10)
	if got.R != 255 || got.G != 0 || got.B != 0 {
		t.Errorf("FillRect interior pixel = %v, want red", got)
	}

	// Check a pixel outside.
	got = c.Image().RGBAAt(2, 2)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("FillRect exterior pixel = %v, want white", got)
	}
}

func TestClearRect(t *testing.T) {
	c := New(20, 20)
	c.SetFillColor(RGB(255, 0, 0))
	c.FillRect(0, 0, 20, 20)
	c.ClearRect(5, 5, 10, 10)

	// Cleared region should be transparent.
	got := c.Image().RGBAAt(10, 10)
	if got.A != 0 {
		t.Errorf("ClearRect pixel alpha = %d, want 0", got.A)
	}

	// Outside region should still be red.
	got = c.Image().RGBAAt(2, 2)
	if got.R != 255 {
		t.Errorf("outside ClearRect R = %d, want 255", got.R)
	}
}

func TestStrokeRect(t *testing.T) {
	c := New(30, 30)
	c.SetStrokeColor(RGB(0, 0, 255))
	c.SetLineWidth(2)
	c.StrokeRect(5, 5, 20, 20)

	// A pixel on the border should be blue.
	got := c.Image().RGBAAt(5, 5)
	if got.B == 0 {
		t.Error("StrokeRect border pixel should have blue component")
	}

	// Interior should be white (not filled).
	got = c.Image().RGBAAt(15, 15)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("StrokeRect interior = %v, want white", got)
	}
}

func TestTranslatedFillRect(t *testing.T) {
	c := New(30, 30)
	c.SetFillColor(RGB(0, 255, 0))
	c.Translate(10, 10)
	c.FillRect(0, 0, 5, 5)

	// The rect should appear at (10,10)-(15,15).
	got := c.Image().RGBAAt(12, 12)
	if got.G != 255 {
		t.Errorf("translated rect pixel = %v, want green", got)
	}

	// Original (0,0) should be white.
	got = c.Image().RGBAAt(2, 2)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("origin pixel = %v, want white", got)
	}
}

func TestScaledFillRect(t *testing.T) {
	c := New(40, 40)
	c.SetFillColor(RGB(255, 0, 0))
	c.Scale(2, 2)
	c.FillRect(5, 5, 5, 5)

	// Rect at (5,5)-(10,10) scaled 2x should appear at (10,10)-(20,20).
	got := c.Image().RGBAAt(15, 15)
	if got.R != 255 || got.G != 0 || got.B != 0 {
		t.Errorf("scaled rect interior = %v, want red", got)
	}

	// Original unscaled position (7,7) should still be white.
	got = c.Image().RGBAAt(7, 7)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("outside scaled rect = %v, want white", got)
	}

	// Past the scaled rect (25,25) should be white.
	got = c.Image().RGBAAt(25, 25)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("beyond scaled rect = %v, want white", got)
	}
}

func TestNonUniformScale(t *testing.T) {
	c := New(60, 30)
	c.SetFillColor(RGB(0, 0, 255))
	c.Scale(3, 1)
	c.FillRect(5, 5, 10, 10)

	// X is scaled 3x: rect should span x=[15,45], y=[5,15].
	got := c.Image().RGBAAt(30, 10)
	if got.B != 255 {
		t.Errorf("stretched rect interior = %v, want blue", got)
	}

	// Y is not scaled: y=20 should be outside.
	got = c.Image().RGBAAt(30, 20)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("below stretched rect = %v, want white", got)
	}
}

func TestRotate90FillRect(t *testing.T) {
	c := New(40, 40)
	c.SetFillColor(RGB(0, 255, 0))
	c.Translate(20, 20)
	c.Rotate(math.Pi / 2) // 90 degrees CCW
	c.FillRect(0, 0, 10, 5)

	// After 90 deg rotation around (20,20): rect (0,0)-(10,5)
	// maps to roughly (20,20) extending left and down.
	// Point (0,5) in local -> rotated: (-5, 0) + translate -> (15, 20).
	got := c.Image().RGBAAt(17, 22)
	if got.G != 255 {
		t.Errorf("rotated 90 rect pixel = %v, want green", got)
	}

	// Far corner should be white.
	got = c.Image().RGBAAt(35, 35)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("outside rotated rect = %v, want white", got)
	}
}

func TestSkewTransform(t *testing.T) {
	c := New(60, 40)
	c.SetFillColor(RGB(200, 100, 0))
	c.Translate(10, 10)
	c.Transform(SkewMatrix(math.Pi/4, 0)) // 45 degree X skew
	c.FillRect(0, 0, 10, 10)

	// With X skew of 45 deg, bottom-left corner shifts right by ~10px.
	// Check that a pixel at the skewed position is filled.
	got := c.Image().RGBAAt(18, 18)
	if got.R < 150 {
		t.Errorf("skewed rect pixel = %v, want orange-ish", got)
	}

	// Top-left (10,10) should still be filled (start of rect).
	got = c.Image().RGBAAt(11, 11)
	if got.R < 150 {
		t.Errorf("skewed rect top-left = %v, want filled", got)
	}
}

func TestNegativeScale(t *testing.T) {
	c := New(40, 20)
	c.SetFillColor(RGB(255, 0, 0))

	// Draw at right side using negative X scale (mirror).
	c.Translate(30, 0)
	c.Scale(-1, 1)
	c.FillRect(0, 5, 10, 10)

	// Mirrored: rect should appear at x=[20,30], y=[5,15].
	got := c.Image().RGBAAt(25, 10)
	if got.R != 255 {
		t.Errorf("mirrored rect pixel = %v, want red", got)
	}

	// Left side should be white.
	got = c.Image().RGBAAt(5, 10)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("left of mirrored rect = %v, want white", got)
	}
}

func TestSetTransformOverrides(t *testing.T) {
	c := New(30, 30)
	c.SetFillColor(RGB(0, 0, 255))

	// Apply some transform, then override with SetTransform.
	c.Translate(100, 100)
	c.Rotate(1.0)
	c.SetTransform(TranslateMatrix(5, 5))
	c.FillRect(0, 0, 10, 10)

	// Should appear at (5,5)-(15,15), ignoring previous translate/rotate.
	got := c.Image().RGBAAt(10, 10)
	if got.B != 255 {
		t.Errorf("SetTransform rect = %v, want blue", got)
	}

	got = c.Image().RGBAAt(2, 2)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("outside SetTransform rect = %v, want white", got)
	}
}

func TestResetTransformRestoresIdentity(t *testing.T) {
	c := New(30, 30)
	c.SetFillColor(RGB(255, 0, 255))

	c.Translate(100, 100)
	c.Scale(5, 5)
	c.Rotate(2.0)
	c.ResetTransform()
	c.FillRect(5, 5, 10, 10)

	// Should draw at literal (5,5)-(15,15) since transform was reset.
	got := c.Image().RGBAAt(10, 10)
	if got.R != 255 || got.B != 255 {
		t.Errorf("ResetTransform rect = %v, want magenta", got)
	}

	got = c.Image().RGBAAt(2, 2)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("outside ResetTransform rect = %v, want white", got)
	}
}

func TestNestedSaveRestoreTransforms(t *testing.T) {
	c := New(60, 60)

	// Level 0: identity.
	c.Save()
	c.Translate(10, 10)

	// Level 1: translated by (10,10).
	c.Save()
	c.Translate(20, 20)

	// Level 2: translated by (30,30) total.
	c.SetFillColor(RGB(255, 0, 0))
	c.FillRect(0, 0, 10, 10) // Should appear at (30,30)-(40,40).

	c.Restore() // Back to level 1 (10,10).
	c.SetFillColor(RGB(0, 255, 0))
	c.FillRect(0, 0, 5, 5) // Should appear at (10,10)-(15,15).

	c.Restore() // Back to level 0 (identity).
	c.SetFillColor(RGB(0, 0, 255))
	c.FillRect(0, 0, 5, 5) // Should appear at (0,0)-(5,5).

	// Check nested red rect.
	got := c.Image().RGBAAt(35, 35)
	if got.R != 255 {
		t.Errorf("nested level 2 = %v, want red", got)
	}

	// Check level 1 green rect.
	got = c.Image().RGBAAt(12, 12)
	if got.G != 255 {
		t.Errorf("nested level 1 = %v, want green", got)
	}

	// Check level 0 blue rect.
	got = c.Image().RGBAAt(2, 2)
	if got.B != 255 {
		t.Errorf("nested level 0 = %v, want blue", got)
	}
}

func TestTransformComposition(t *testing.T) {
	// Verify that sequential transforms compose correctly.
	// Translate(10,0) then Scale(2,2): matrix = T * S.
	// Point (x,y) -> (2x+10, 2y). Rect (0,0)-(10,10) -> (10,0)-(30,20).
	c := New(60, 30)
	c.SetFillColor(RGB(255, 0, 0))
	c.Translate(10, 0)
	c.Scale(2, 2)
	c.FillRect(0, 0, 10, 10)

	// Interior at (20,10) should be red.
	got := c.Image().RGBAAt(20, 10)
	if got.R != 255 {
		t.Errorf("composed transform interior = %v, want red", got)
	}

	// Before the rect (x=5) should be white.
	got = c.Image().RGBAAt(5, 10)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("before composed rect = %v, want white", got)
	}

	// After the rect (x=35) should be white.
	got = c.Image().RGBAAt(35, 10)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("after composed rect = %v, want white", got)
	}
}

func TestTransformAppliesToStroke(t *testing.T) {
	c := New(40, 40)
	c.SetStrokeColor(RGB(255, 0, 0))
	c.SetLineWidth(2)

	c.Translate(20, 0)
	c.BeginPath()
	c.MoveTo(0, 0)
	c.LineTo(0, 40)
	c.Stroke()

	// The vertical line should appear at x=20.
	got := c.Image().RGBAAt(20, 20)
	if got.R == 0 {
		t.Error("translated stroke should have red at x=20")
	}

	// x=5 should be white.
	got = c.Image().RGBAAt(5, 20)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("left of translated stroke = %v, want white", got)
	}
}

func TestTransformAppliesToClearRect(t *testing.T) {
	c := New(30, 30)
	c.SetFillColor(RGB(255, 0, 0))
	c.FillRect(0, 0, 30, 30)

	c.Translate(10, 10)
	c.ClearRect(0, 0, 10, 10)

	// Cleared region at (10,10)-(20,20) should be transparent.
	got := c.Image().RGBAAt(15, 15)
	if got.A != 0 {
		t.Errorf("translated ClearRect pixel alpha = %d, want 0", got.A)
	}

	// Outside cleared region should still be red.
	got = c.Image().RGBAAt(5, 5)
	if got.R != 255 {
		t.Errorf("outside translated ClearRect = %v, want red", got)
	}
}

func TestTransformCustomMatrix(t *testing.T) {
	// Use Transform() to apply a manual matrix.
	c := New(40, 40)
	c.SetFillColor(RGB(0, 200, 0))

	// Apply a 2x scale + translate(5,5) via raw matrix.
	c.Transform(Matrix{2, 0, 5, 0, 2, 5})
	c.FillRect(0, 0, 10, 10)

	// Rect should appear at (5,5)-(25,25).
	got := c.Image().RGBAAt(15, 15)
	if got.G < 180 {
		t.Errorf("custom matrix rect interior = %v, want green", got)
	}

	got = c.Image().RGBAAt(2, 2)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("outside custom matrix rect = %v, want white", got)
	}
}

func TestSaveRestorePreservesTransformIndependently(t *testing.T) {
	c := New(10, 10)

	c.Save()
	c.Translate(5, 5)
	c.SetFillColor(RGB(255, 0, 0))
	saved := c.state.matrix
	c.Restore()

	// After restore, matrix should be identity again.
	if c.state.matrix != Identity() {
		t.Errorf("matrix after restore = %v, want identity", c.state.matrix)
	}

	// The saved matrix should have had the translation.
	if saved == Identity() {
		t.Error("saved matrix should not be identity")
	}
}

func TestGlobalAlpha(t *testing.T) {
	c := New(10, 10)
	c.SetGlobalAlpha(0.5)
	c.SetFillColor(RGB(255, 0, 0))
	c.FillRect(0, 0, 10, 10)

	got := c.Image().RGBAAt(5, 5)
	// With 50% alpha red over white, we expect approximately R=255, G=128, B=128.
	if got.R < 200 || got.G < 100 || got.G > 160 {
		t.Errorf("global alpha pixel = %v, expected blended red-on-white", got)
	}
}
