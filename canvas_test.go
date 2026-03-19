package gocanvas

import (
	"image/color"
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
