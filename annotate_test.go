package gocanvas

import (
	"image/color"
	"testing"
)

func TestDrawAABB(t *testing.T) {
	c := New(100, 100)
	style := DefaultAnnotStyle()
	DrawAABB(c, 10, 10, 80, 80, style)

	// Border pixel should be green.
	got := c.Image().RGBAAt(10, 10)
	if got.G < 100 {
		t.Errorf("DrawAABB border pixel = %v, expected green", got)
	}
}

func TestDrawAABBClamping(t *testing.T) {
	c := New(100, 100)
	style := DefaultAnnotStyle()

	// Box extends beyond canvas — should not panic.
	DrawAABB(c, -10, -10, 200, 200, style)
}

func TestDrawLabel(t *testing.T) {
	f := loadTestFont(t)
	c := New(200, 100)
	style := DefaultAnnotStyle()
	style.Font = f

	DrawLabel(c, "Test", 10, 10, style)

	// Should have dark background pixels somewhere in the label region.
	hasDark := false
	for y := 10; y < 40; y++ {
		for x := 10; x < 100; x++ {
			px := c.Image().RGBAAt(x, y)
			if px.A > 200 && px.R < 80 && px.G < 80 && px.B < 80 {
				hasDark = true
				break
			}
		}
		if hasDark {
			break
		}
	}
	if !hasDark {
		t.Error("DrawLabel: no dark background pixels found")
	}
}

func TestDrawLabelClamping(t *testing.T) {
	f := loadTestFont(t)
	c := New(100, 100)
	style := DefaultAnnotStyle()
	style.Font = f

	// Label near edge — should clamp and not panic.
	DrawLabel(c, "Overflow Test", 80, 5, style)
}

func TestDrawPolygon(t *testing.T) {
	c := New(100, 100)
	style := DefaultAnnotStyle()
	style.FillColor = RGBA(255, 0, 0, 128)

	pts := []Point{{10, 10}, {90, 10}, {50, 90}}
	DrawPolygon(c, pts, style)

	// Interior should have some red from fill.
	got := c.Image().RGBAAt(50, 40)
	if got == (color.RGBA{255, 255, 255, 255}) {
		t.Error("DrawPolygon: interior is white, expected fill")
	}
}

func TestDrawLabeledBox(t *testing.T) {
	f := loadTestFont(t)
	c := New(300, 200)
	style := DefaultAnnotStyle()
	style.Font = f

	DrawLabeledBox(c, "Object", 50, 50, 100, 80, style)

	// Box border should be visible.
	got := c.Image().RGBAAt(50, 50)
	if got == (color.RGBA{255, 255, 255, 255}) {
		t.Error("DrawLabeledBox: no box visible at (50,50)")
	}
}

func TestDrawLabeledBoxTopClamping(t *testing.T) {
	f := loadTestFont(t)
	c := New(300, 200)
	style := DefaultAnnotStyle()
	style.Font = f

	// Box at top — label should be placed inside.
	DrawLabeledBox(c, "Top", 10, 5, 100, 80, style)
}
