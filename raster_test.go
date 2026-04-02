package gocanvas

import (
	"image"
	"image/color"
	"testing"
)

func TestBuildEdgesTriangle(t *testing.T) {
	tri := [][]Point{{
		{0, 0}, {10, 0}, {5, 10}, {0, 0},
	}}
	edges := buildEdges(tri)
	// The horizontal edge (0,0)-(10,0) should be skipped.
	// Remaining: (10,0)-(5,10), (5,10)-(0,0) = 2 edges
	if len(edges) != 2 {
		t.Errorf("buildEdges triangle: got %d edges, want 2", len(edges))
	}
}

func TestRasterizeFillRect(t *testing.T) {
	dst := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Fill with white first.
	for i := range dst.Pix {
		dst.Pix[i] = 255
	}

	// Rect from (2,2) to (6,6).
	rect := [][]Point{{
		{2, 2}, {6, 2}, {6, 6}, {2, 6}, {2, 2},
	}}
	edges := buildEdges(rect)
	red := color.RGBA{255, 0, 0, 255}
	rasterizeFill(dst, edges, red, CompSourceOver, FillRuleWinding)

	// Check a pixel inside the rect.
	got := dst.RGBAAt(3, 3)
	if got != red {
		t.Errorf("pixel (3,3) = %v, want red", got)
	}

	// Check a pixel outside the rect.
	got = dst.RGBAAt(1, 1)
	if got != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("pixel (1,1) = %v, want white", got)
	}
}

func TestBlendPixelFullAlpha(t *testing.T) {
	dst := image.NewRGBA(image.Rect(0, 0, 1, 1))
	dst.SetRGBA(0, 0, color.RGBA{100, 100, 100, 255})

	blendPixel(dst, 0, 0, color.RGBA{255, 0, 0, 255}, CompSourceOver)
	got := dst.RGBAAt(0, 0)
	want := color.RGBA{255, 0, 0, 255}
	if got != want {
		t.Errorf("blendPixel full alpha: got %v, want %v", got, want)
	}
}

func TestBlendPixelHalfAlpha(t *testing.T) {
	dst := image.NewRGBA(image.Rect(0, 0, 1, 1))
	dst.SetRGBA(0, 0, color.RGBA{0, 0, 0, 255})

	// Blend 50% red onto black.
	blendPixel(dst, 0, 0, color.RGBA{255, 0, 0, 128}, CompSourceOver)
	got := dst.RGBAAt(0, 0)

	// Expected: src premul = (128, 0, 0, 128)
	// out = src + dst * (1 - srcA/255)
	// out.R = 128*255/255 + 0*(255-128)/255 = 128
	// out.A = 128*255/255 + 255*(255-128)/255 = 128 + 127 = 255
	if got.R < 126 || got.R > 130 {
		t.Errorf("blendPixel half alpha: R = %d, want ~128", got.R)
	}
	if got.A != 255 {
		t.Errorf("blendPixel half alpha: A = %d, want 255", got.A)
	}
}

func TestEvenOddFillRule(t *testing.T) {
	// Two concentric rectangles drawn in the same winding direction.
	// With even-odd rule, the inner rect should create a hole.
	c := New(50, 50)
	c.SetFillColor(color.RGBA{255, 0, 0, 255})
	c.SetFillRule(FillRuleEvenOdd)

	c.BeginPath()
	// Outer rect: 5,5 to 45,45
	c.Rect(5, 5, 40, 40)
	// Inner rect: 15,15 to 35,35
	c.Rect(15, 15, 20, 20)
	c.Fill()

	// Center pixel (25,25) should NOT be filled (hole) — should be white.
	center := c.GetPixel(25, 25)
	if center.R != 255 || center.G != 255 || center.B != 255 {
		t.Errorf("EvenOdd center pixel = %v, want white (hole)", center)
	}

	// Pixel in outer ring (10,10) should be red.
	outer := c.GetPixel(10, 10)
	if outer.R != 255 || outer.G != 0 || outer.B != 0 {
		t.Errorf("EvenOdd outer pixel = %v, want red", outer)
	}
}

func TestWindingFillRule(t *testing.T) {
	// Same geometry as TestEvenOddFillRule but with winding rule.
	// Both rects wound the same way, so center should be filled.
	c := New(50, 50)
	c.SetFillColor(color.RGBA{255, 0, 0, 255})
	c.SetFillRule(FillRuleWinding)

	c.BeginPath()
	// Outer rect: 5,5 to 45,45
	c.Rect(5, 5, 40, 40)
	// Inner rect: 15,15 to 35,35
	c.Rect(15, 15, 20, 20)
	c.Fill()

	// Center pixel (25,25) should be red (filled).
	center := c.GetPixel(25, 25)
	if center.R != 255 || center.G != 0 || center.B != 0 {
		t.Errorf("Winding center pixel = %v, want red", center)
	}

	// Pixel in outer ring (10,10) should also be red.
	outer := c.GetPixel(10, 10)
	if outer.R != 255 || outer.G != 0 || outer.B != 0 {
		t.Errorf("Winding outer pixel = %v, want red", outer)
	}
}

func TestStrokePathHorizontalLine(t *testing.T) {
	sp := [][]Point{{
		{0, 10}, {20, 10},
	}}
	outlines := strokePath(sp, 4, CapButt, JoinMiter, 10)
	if len(outlines) != 1 {
		t.Fatalf("strokePath: got %d outlines, want 1", len(outlines))
	}
	// The outline should have points above and below y=10, roughly at y=8 and y=12.
	var minY, maxY float64
	minY, maxY = 1e10, -1e10
	for _, p := range outlines[0] {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	if minY > 9 || maxY < 11 {
		t.Errorf("strokePath: y range [%v, %v], expected to span around y=10±2", minY, maxY)
	}
}
