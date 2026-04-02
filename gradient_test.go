package gocanvas

import (
	"image/color"
	"math"
	"testing"
)

func TestInterpolateStopsSingleStop(t *testing.T) {
	stops := []colorStop{{0.5, color.RGBA{255, 0, 0, 255}}}
	got := interpolateStops(stops, 0.0)
	if got != (color.RGBA{255, 0, 0, 255}) {
		t.Fatalf("expected red, got %v", got)
	}
}

func TestInterpolateStopsTwoStops(t *testing.T) {
	stops := []colorStop{
		{0, color.RGBA{0, 0, 0, 255}},
		{1, color.RGBA{254, 254, 254, 255}},
	}

	// At midpoint, expect ~127.
	got := interpolateStops(stops, 0.5)
	if got.R < 125 || got.R > 129 {
		t.Fatalf("expected R ~127, got %d", got.R)
	}

	// At 0, expect black.
	got = interpolateStops(stops, 0)
	if got != (color.RGBA{0, 0, 0, 255}) {
		t.Fatalf("expected black at 0, got %v", got)
	}

	// At 1, expect near-white.
	got = interpolateStops(stops, 1.0)
	if got.R < 253 {
		t.Fatalf("expected ~254 at 1.0, got %d", got.R)
	}
}

func TestInterpolateStopsClamp(t *testing.T) {
	stops := []colorStop{
		{0.2, color.RGBA{100, 0, 0, 255}},
		{0.8, color.RGBA{0, 100, 0, 255}},
	}
	// Below first stop.
	got := interpolateStops(stops, 0.0)
	if got != (color.RGBA{100, 0, 0, 255}) {
		t.Fatalf("expected first stop color for t<min, got %v", got)
	}
	// Above last stop.
	got = interpolateStops(stops, 1.0)
	if got != (color.RGBA{0, 100, 0, 255}) {
		t.Fatalf("expected last stop color for t>max, got %v", got)
	}
}

func TestLinearGradientHorizontal(t *testing.T) {
	g := NewLinearGradient(0, 0, 100, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	// At start.
	c0 := g.ColorAt(0, 50)
	if c0.R < 250 || c0.B > 5 {
		t.Fatalf("expected red at x=0, got %v", c0)
	}

	// At end.
	c1 := g.ColorAt(100, 50)
	if c1.B < 250 || c1.R > 5 {
		t.Fatalf("expected blue at x=100, got %v", c1)
	}

	// At midpoint.
	cm := g.ColorAt(50, 50)
	if cm.R < 120 || cm.R > 135 || cm.B < 120 || cm.B > 135 {
		t.Fatalf("expected roughly equal R/B at midpoint, got %v", cm)
	}
}

func TestLinearGradientVertical(t *testing.T) {
	g := NewLinearGradient(0, 0, 0, 200)
	g.AddColorStop(0, color.RGBA{0, 255, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	c0 := g.ColorAt(100, 0)
	if c0.G < 250 {
		t.Fatalf("expected green at top, got %v", c0)
	}
	c1 := g.ColorAt(100, 200)
	if c1.B < 250 {
		t.Fatalf("expected blue at bottom, got %v", c1)
	}
}

func TestRadialGradientConcentric(t *testing.T) {
	g := NewRadialGradient(50, 50, 0, 50, 50, 50)
	g.AddColorStop(0, color.RGBA{255, 255, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	// Center should be yellow.
	cc := g.ColorAt(50, 50)
	if cc.R < 250 || cc.G < 250 {
		t.Fatalf("expected yellow at center, got %v", cc)
	}

	// Edge should be blue.
	ce := g.ColorAt(100, 50)
	if ce.B < 250 {
		t.Fatalf("expected blue at edge, got %v", ce)
	}

	// Midpoint.
	cm := g.ColorAt(75, 50)
	if cm.R < 120 || cm.R > 135 || cm.B < 120 || cm.B > 135 {
		t.Fatalf("expected mix at midpoint, got %v", cm)
	}
}

func TestLinearGradientMultipleStops(t *testing.T) {
	g := NewLinearGradient(0, 0, 100, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(0.5, color.RGBA{0, 255, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	// At 25% should be red-green mix.
	c25 := g.ColorAt(25, 0)
	if c25.R < 120 || c25.G < 120 || c25.B > 10 {
		t.Fatalf("expected red+green mix at 25%%, got %v", c25)
	}

	// At 50% should be green.
	c50 := g.ColorAt(50, 0)
	if c50.G < 250 {
		t.Fatalf("expected green at 50%%, got %v", c50)
	}

	// At 75% should be green-blue mix.
	c75 := g.ColorAt(75, 0)
	if c75.G < 120 || c75.B < 120 || c75.R > 10 {
		t.Fatalf("expected green+blue mix at 75%%, got %v", c75)
	}
}

func TestCanvasFillRectGradient(t *testing.T) {
	c := New(100, 100)
	g := NewLinearGradient(0, 0, 100, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	c.SetFillGradient(g)
	c.FillRect(0, 0, 100, 100)

	// Check left edge is red.
	px := c.Image().RGBAAt(0, 50)
	if px.R < 240 {
		t.Fatalf("expected red on left, got %v", px)
	}

	// Check right edge is blue.
	px = c.Image().RGBAAt(99, 50)
	if px.B < 240 {
		t.Fatalf("expected blue on right, got %v", px)
	}
}

func TestSetFillColorClearsGradient(t *testing.T) {
	c := New(50, 50)
	g := NewLinearGradient(0, 0, 50, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	c.SetFillGradient(g)
	c.SetFillColor(color.RGBA{0, 255, 0, 255})
	c.FillRect(0, 0, 50, 50)

	// Should be solid green, not gradient.
	px := c.Image().RGBAAt(25, 25)
	if px.G < 250 || px.R > 5 || px.B > 5 {
		t.Fatalf("expected solid green, got %v", px)
	}
}

func TestCanvasGradientWithTransform(t *testing.T) {
	c := New(100, 100)
	g := NewLinearGradient(0, 0, 50, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	c.SetFillGradient(g)
	c.Scale(2, 2)
	c.FillRect(0, 0, 50, 50)

	// At pixel (0,50) in screen space maps to (0,25) in world space - should be red.
	px := c.Image().RGBAAt(0, 50)
	if px.R < 240 {
		t.Fatalf("expected red at left with 2x scale, got %v", px)
	}

	// At pixel (99,50) in screen space maps to (49.5,25) in world space - should be near blue.
	px = c.Image().RGBAAt(99, 50)
	if px.B < 240 {
		t.Fatalf("expected blue at right with 2x scale, got %v", px)
	}
}

func TestCanvasStrokeGradient(t *testing.T) {
	c := New(100, 100)
	g := NewLinearGradient(0, 0, 100, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	c.SetStrokeGradient(g)
	c.SetLineWidth(10)
	c.StrokeRect(10, 10, 80, 80)

	// Top-left area of the stroke should be red-ish.
	px := c.Image().RGBAAt(10, 10)
	if px.R < 200 {
		t.Fatalf("expected red-ish stroke at left, got %v", px)
	}

	// Top-right area should be blue-ish.
	px = c.Image().RGBAAt(90, 10)
	if px.B < 200 {
		t.Fatalf("expected blue-ish stroke at right, got %v", px)
	}
}

func TestCanvasFillPathGradient(t *testing.T) {
	c := New(100, 100)
	g := NewLinearGradient(0, 0, 100, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	c.SetFillGradient(g)
	c.BeginPath()
	c.Arc(50, 50, 40, 0, 2*math.Pi)
	c.Fill()

	// Center should be roughly purple.
	px := c.Image().RGBAAt(50, 50)
	if px.R < 100 || px.B < 100 {
		t.Fatalf("expected purple-ish at center of circle, got %v", px)
	}
}

func TestRadialGradientAddColorStopSorted(t *testing.T) {
	g := NewRadialGradient(0, 0, 0, 0, 0, 100)
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	// Stops should be sorted: 0 first, 1 second.
	if g.stops[0].Position != 0 || g.stops[1].Position != 1 {
		t.Fatalf("stops not sorted: %v", g.stops)
	}
}

func TestSaveRestorePreservesGradient(t *testing.T) {
	c := New(50, 50)
	g := NewLinearGradient(0, 0, 50, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	c.SetFillGradient(g)
	c.Save()
	c.SetFillColor(color.RGBA{0, 255, 0, 255})
	c.Restore()

	// After restore, gradient should be back.
	c.FillRect(0, 0, 50, 50)
	px := c.Image().RGBAAt(0, 25)
	if px.R < 240 {
		t.Fatalf("expected red after restore, got %v", px)
	}
}

func TestLerpColor(t *testing.T) {
	a := color.RGBA{0, 0, 0, 255}
	b := color.RGBA{200, 100, 50, 255}

	c := lerpColor(a, b, 0)
	if c != a {
		t.Fatalf("lerp at 0 should equal a, got %v", c)
	}

	c = lerpColor(a, b, 1)
	if c != b {
		t.Fatalf("lerp at 1 should equal b, got %v", c)
	}

	c = lerpColor(a, b, 0.5)
	if c.R != 100 || c.G != 50 || c.B != 25 {
		t.Fatalf("lerp at 0.5 unexpected: %v", c)
	}
}

func TestConicGradientBasic(t *testing.T) {
	g := NewConicGradient(50, 50, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	// Left side (angle = π → t = 1.0) should be blue.
	c0 := g.ColorAt(0, 50)
	if c0.B < 240 {
		t.Fatalf("expected blue at left (t=1), got %v", c0)
	}

	// Right side (angle = 0 → t = 0.5) should be midpoint.
	cm := g.ColorAt(100, 50)
	if cm.R < 100 || cm.R > 150 || cm.B < 100 || cm.B > 150 {
		t.Fatalf("expected mid-range at right, got %v", cm)
	}
}

func TestConicGradientRotation(t *testing.T) {
	g := NewConicGradient(50, 50, 180)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	// With 180 deg rotation, the right side (angle 0 → base t=0.5)
	// becomes t=0.5-0.5=0.0 → red.
	c0 := g.ColorAt(100, 50)
	if c0.R < 240 {
		t.Fatalf("expected red at right with 180 rotation, got %v", c0)
	}

	// Left side (angle π → base t=1.0) becomes t=1.0-0.5=0.5 → midpoint.
	cm := g.ColorAt(0, 50)
	if cm.R < 100 || cm.R > 150 || cm.B < 100 || cm.B > 150 {
		t.Fatalf("expected midpoint at left with 180 rotation, got %v", cm)
	}
}

func TestConicGradientMultipleStops(t *testing.T) {
	g := NewConicGradient(50, 50, 0)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(0.5, color.RGBA{0, 255, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})

	// Top (angle ≈ -π/2 → t ≈ 0.25) should be red-green mix.
	ct := g.ColorAt(50, 0)
	if ct.R < 100 || ct.G < 100 || ct.B > 20 {
		t.Fatalf("expected red+green mix at top, got %v", ct)
	}

	// Bottom (angle ≈ π/2 → t ≈ 0.75) should be green-blue mix.
	cb := g.ColorAt(50, 100)
	if cb.G < 100 || cb.B < 100 || cb.R > 20 {
		t.Fatalf("expected green+blue mix at bottom, got %v", cb)
	}
}
