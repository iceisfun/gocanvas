package gocanvas

import (
	"image/color"
	"testing"
)

func TestApplyDashSimple(t *testing.T) {
	// Horizontal line from (0,0) to (20,0) with 5-on/5-off pattern.
	sp := [][]Point{{{0, 0}, {20, 0}}}
	result := applyDash(sp, []float64{5, 5}, 0)

	// Expect 2 dash segments: (0,0)-(5,0) and (10,0)-(15,0).
	if len(result) != 2 {
		t.Fatalf("applyDash: got %d segments, want 2", len(result))
	}
}

func TestApplyDashOffset(t *testing.T) {
	sp := [][]Point{{{0, 0}, {20, 0}}}
	result := applyDash(sp, []float64{5, 5}, 5)

	// Offset of 5 starts in the gap, so first dash starts at x=5.
	if len(result) < 1 {
		t.Fatal("applyDash with offset: no segments")
	}
}

func TestApplyDashOddPattern(t *testing.T) {
	// Odd pattern [3] should become [3, 3].
	sp := [][]Point{{{0, 0}, {18, 0}}}
	result := applyDash(sp, []float64{3}, 0)

	// 18 units / (3+3) = 3 full cycles = 3 dash segments.
	if len(result) != 3 {
		t.Errorf("applyDash odd pattern: got %d segments, want 3", len(result))
	}
}

func TestStrokeWithDash(t *testing.T) {
	c := New(100, 20)
	c.SetStrokeColor(RGB(0, 0, 0))
	c.SetLineWidth(2)
	c.SetLineDash([]float64{10, 5})

	c.BeginPath()
	c.MoveTo(0, 10)
	c.LineTo(100, 10)
	c.Stroke()

	// Check that there are gaps (white pixels in the middle of the line).
	// At x=12 (within first dash) should be black.
	got := c.Image().RGBAAt(5, 10)
	if got == (color.RGBA{255, 255, 255, 255}) {
		t.Error("expected dash segment at x=5, got white")
	}

	// Somewhere in a gap should be white.
	// First gap starts at x=10, spans to x=15.
	got = c.Image().RGBAAt(12, 10)
	if got.R < 200 {
		t.Errorf("expected gap at x=12, got %v", got)
	}
}

func TestSaveRestoreDash(t *testing.T) {
	c := New(10, 10)
	c.SetLineDash([]float64{5, 5})
	c.Save()
	c.SetLineDash(nil)
	if c.LineDash() != nil {
		t.Error("expected nil dash after setting nil")
	}
	c.Restore()
	dash := c.LineDash()
	if len(dash) != 2 || dash[0] != 5 || dash[1] != 5 {
		t.Errorf("expected [5,5] after restore, got %v", dash)
	}
}
