package gocanvas

import (
	"image"
	"image/color"
	"testing"
)

func TestHasShadow(t *testing.T) {
	c := New(10, 10)

	if c.hasShadow() {
		t.Error("expected no shadow by default")
	}

	c.SetShadowColor(RGBA(0, 0, 0, 128))
	c.SetShadowBlur(5)
	if !c.hasShadow() {
		t.Error("expected shadow after setting color and blur")
	}
}

func TestShadowOffset(t *testing.T) {
	c := New(50, 50)
	c.SetShadowColor(RGBA(0, 0, 0, 255))
	c.SetShadowOffset(5, 5)
	c.SetShadowBlur(0) // no blur for precise test

	c.SetFillColor(RGB(255, 0, 0))
	c.FillRect(10, 10, 10, 10)

	// Shadow should appear at offset (15,15) to (25,25).
	shadow := c.Image().RGBAAt(20, 20)
	// Should not be white (shadow is there).
	if shadow == (color.RGBA{255, 255, 255, 255}) {
		t.Error("expected shadow at offset position")
	}

	// Original shape at (15, 15) should be red.
	shape := c.Image().RGBAAt(15, 15)
	if shape.R < 200 {
		t.Errorf("expected red shape, got %v", shape)
	}
}

func TestBoxBlur(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Set a single white pixel in the center.
	img.SetRGBA(5, 5, color.RGBA{255, 255, 255, 255})

	boxBlurRGBA(img, 1)

	// The center pixel should have been spread out (no longer 255).
	got := img.RGBAAt(5, 5)
	if got.R == 255 || got.R == 0 {
		t.Errorf("expected blurred value, got R=%d", got.R)
	}

	// An adjacent pixel should have some value.
	adj := img.RGBAAt(4, 5)
	if adj.R == 0 {
		t.Error("expected blur spread to adjacent pixel")
	}
}

func TestCompositeOver(t *testing.T) {
	dst := image.NewRGBA(image.Rect(0, 0, 10, 10))
	src := image.NewRGBA(image.Rect(0, 0, 5, 5))

	// Fill src with red.
	for i := 0; i < len(src.Pix); i += 4 {
		src.Pix[i+0] = 255
		src.Pix[i+3] = 255
	}

	compositeOver(dst, src, 3, 3, 1.0)

	got := dst.RGBAAt(4, 4)
	if got.R != 255 || got.A != 255 {
		t.Errorf("compositeOver: got %v at (4,4), want red", got)
	}

	// Outside offset region should be transparent.
	got = dst.RGBAAt(1, 1)
	if got.A != 0 {
		t.Errorf("compositeOver: got %v at (1,1), want transparent", got)
	}
}
