package gocanvas

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// newRedImage creates a 4x4 solid red *image.RGBA for testing.
func newRedImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = 255
		img.Pix[i+1] = 0
		img.Pix[i+2] = 0
		img.Pix[i+3] = 255
	}
	return img
}

// newCheckerImage creates a 4x4 image: top-left 2x2 red, rest blue.
func newCheckerImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			off := img.PixOffset(x, y)
			if x < 2 && y < 2 {
				img.Pix[off+0] = 255
				img.Pix[off+1] = 0
				img.Pix[off+2] = 0
				img.Pix[off+3] = 255
			} else {
				img.Pix[off+0] = 0
				img.Pix[off+1] = 0
				img.Pix[off+2] = 255
				img.Pix[off+3] = 255
			}
		}
	}
	return img
}

func TestDrawImage_Basic(t *testing.T) {
	c := New(10, 10)
	src := newRedImage()

	// Draw full source to (2,2) at same size.
	c.DrawImage(src, 0, 0, 4, 4, 2, 2, 4, 4)

	// Pixel at (3,3) should be red (inside drawn area).
	r, g, b, _ := c.dst.At(3, 3).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("expected red at (3,3), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Pixel at (0,0) should still be white (outside drawn area).
	r, g, b, _ = c.dst.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("expected white at (0,0), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawImage_SourceSubrect(t *testing.T) {
	c := New(10, 10)
	src := newCheckerImage()

	// Draw only the top-left 2x2 (red) portion, scaled to 4x4 on canvas.
	c.DrawImage(src, 0, 0, 2, 2, 0, 0, 4, 4)

	// Should be red throughout the 4x4 area.
	r, g, b, _ := c.dst.At(1, 1).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("expected red at (1,1), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Draw only the bottom-right 2x2 (blue) portion, at (5,0).
	c.DrawImage(src, 2, 2, 2, 2, 5, 0, 4, 4)

	r, g, b, _ = c.dst.At(6, 1).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("expected blue at (6,1), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawImage_Scale(t *testing.T) {
	c := New(20, 20)
	src := newRedImage() // 4x4

	// Draw 4x4 source scaled to 16x16 on canvas.
	c.DrawImage(src, 0, 0, 4, 4, 2, 2, 16, 16)

	// Center of scaled area should be red.
	r, g, b, _ := c.dst.At(10, 10).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("expected red at (10,10), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Outside at (0,0) should be white.
	r, g, b, _ = c.dst.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("expected white at (0,0), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawImage_WithTransform(t *testing.T) {
	c := New(20, 20)
	src := newRedImage()

	// Translate then draw.
	c.Save()
	c.Translate(5, 5)
	c.DrawImage(src, 0, 0, 4, 4, 0, 0, 4, 4)
	c.Restore()

	// Should be red at (6,6) (translated by 5,5).
	r, g, b, _ := c.dst.At(6, 6).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("expected red at (6,6), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// (1,1) should still be white.
	r, g, b, _ = c.dst.At(1, 1).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("expected white at (1,1), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawImage_WithRotation(t *testing.T) {
	c := New(30, 30)
	src := newRedImage()

	// Rotate 90 degrees around center of canvas.
	c.Save()
	c.Translate(15, 15)
	c.Rotate(math.Pi / 2)
	c.Translate(-15, -15)
	c.DrawImage(src, 0, 0, 4, 4, 13, 13, 4, 4)
	c.Restore()

	// After 90° rotation around (15,15), the rect (13,13)-(17,17)
	// maps roughly to (13,13)-(17,17) rotated. The center (15,15) stays fixed.
	// Check that at least some non-white pixels exist in the expected area.
	found := false
	for y := 10; y < 20; y++ {
		for x := 10; x < 20; x++ {
			r, _, _, _ := c.dst.At(x, y).RGBA()
			if r>>8 == 255 {
				_, g, b, _ := c.dst.At(x, y).RGBA()
				if g>>8 == 0 && b>>8 == 0 {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected red pixels in rotated area, found none")
	}
}

func TestDrawImage_GlobalAlpha(t *testing.T) {
	c := New(10, 10)
	src := newRedImage()

	c.SetGlobalAlpha(0.5)
	c.DrawImage(src, 0, 0, 4, 4, 0, 0, 4, 4)

	// Pixel should be a blend of red (50% alpha) over white.
	r, g, b, _ := c.dst.At(1, 1).RGBA()
	rr := r >> 8
	gg := g >> 8
	bb := b >> 8

	// Red ~255, green ~128, blue ~128 (red over white at 50%).
	if rr < 200 || gg < 100 || gg > 160 || bb < 100 || bb > 160 {
		t.Errorf("unexpected alpha-blended color at (1,1): (%d, %d, %d)", rr, gg, bb)
	}
}

func TestDrawImage_ZeroDimensions(t *testing.T) {
	c := New(10, 10)
	src := newRedImage()

	// Should be no-ops (zero or negative dimensions).
	c.DrawImage(src, 0, 0, 0, 4, 0, 0, 4, 4)
	c.DrawImage(src, 0, 0, 4, 0, 0, 0, 4, 4)
	c.DrawImage(src, 0, 0, 4, 4, 0, 0, 0, 4)
	c.DrawImage(src, 0, 0, 4, 4, 0, 0, 4, 0)

	// Should still be all white.
	r, g, b, _ := c.dst.At(1, 1).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("expected white at (1,1), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawImage_GenericImage(t *testing.T) {
	// Use image.NRGBA to exercise the generic path.
	src := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			src.SetNRGBA(x, y, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	c := New(10, 10)
	c.DrawImage(src, 0, 0, 4, 4, 0, 0, 4, 4)

	r, g, b, _ := c.dst.At(1, 1).RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 {
		t.Errorf("expected green at (1,1), got (%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawImage_SemitransparentSource(t *testing.T) {
	// Source with 50% alpha red.
	src := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			src.SetNRGBA(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 128})
		}
	}

	c := New(10, 10)
	c.DrawImage(src, 0, 0, 4, 4, 0, 0, 4, 4)

	// Result should be red blended over white at ~50%.
	r, g, b, _ := c.dst.At(1, 1).RGBA()
	rr := r >> 8
	gg := g >> 8
	bb := b >> 8

	if rr < 200 || gg < 100 || gg > 160 || bb < 100 || bb > 160 {
		t.Errorf("unexpected blended color at (1,1): (%d, %d, %d)", rr, gg, bb)
	}
}
