package gocanvas

import (
	"testing"
)

func TestPathRect(t *testing.T) {
	var p Path
	p.Rect(0, 0, 10, 10)

	// moveTo + 3 lineTo + close = 5 ops
	if len(p.ops) != 5 {
		t.Errorf("Rect produced %d ops, want 5", len(p.ops))
	}
	if p.ops[0].op != opMoveTo {
		t.Error("first op should be moveTo")
	}
	if p.ops[4].op != opClose {
		t.Error("last op should be close")
	}
}

func TestPathCircle(t *testing.T) {
	var p Path
	p.Circle(50, 50, 25)

	// moveTo + 4 cubicTo + close = 6 ops
	if len(p.ops) != 6 {
		t.Errorf("Circle produced %d ops, want 6", len(p.ops))
	}
	if p.ops[0].op != opMoveTo {
		t.Error("first op should be moveTo")
	}
	for i := 1; i <= 4; i++ {
		if p.ops[i].op != opCubicTo {
			t.Errorf("op %d should be cubicTo, got %d", i, p.ops[i].op)
		}
	}
	if p.ops[5].op != opClose {
		t.Error("last op should be close")
	}
}

func TestPathReset(t *testing.T) {
	var p Path
	p.Rect(0, 0, 10, 10)
	p.Reset()
	if len(p.ops) != 0 {
		t.Errorf("Reset: ops len = %d, want 0", len(p.ops))
	}
}

func TestFlattenRect(t *testing.T) {
	var p Path
	p.Rect(0, 0, 10, 10)
	subPaths := p.flatten(defaultFlatness)
	if len(subPaths) != 1 {
		t.Fatalf("flatten rect: got %d sub-paths, want 1", len(subPaths))
	}
	// Should have 5 points: 4 corners + closing point back to start.
	if len(subPaths[0]) != 5 {
		t.Errorf("flatten rect: got %d points, want 5", len(subPaths[0]))
	}
}

func TestFlattenCubic(t *testing.T) {
	var p Path
	p.MoveTo(0, 0)
	p.CubicTo(0, 100, 100, 100, 100, 0)
	subPaths := p.flatten(0.5)
	if len(subPaths) != 1 {
		t.Fatalf("flatten cubic: got %d sub-paths, want 1", len(subPaths))
	}
	// Should produce more than 2 points (the curve is subdivided).
	if len(subPaths[0]) < 3 {
		t.Errorf("flatten cubic: only %d points, expected more", len(subPaths[0]))
	}
}
