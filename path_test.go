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

func TestPathRoundRect(t *testing.T) {
	var p Path
	p.RoundRect(0, 0, 100, 50, 10)

	hasMove := false
	hasCubic := false
	hasLine := false
	hasClose := false
	for _, op := range p.ops {
		switch op.op {
		case opMoveTo:
			hasMove = true
		case opLineTo:
			hasLine = true
		case opCubicTo:
			hasCubic = true
		case opClose:
			hasClose = true
		}
	}
	if !hasMove {
		t.Error("RoundRect should have moveTo")
	}
	if !hasCubic {
		t.Error("RoundRect should have cubicTo (from arcs)")
	}
	if !hasLine {
		t.Error("RoundRect should have lineTo (edges)")
	}
	if !hasClose {
		t.Error("RoundRect should have close")
	}
}

func TestPathRoundRectZeroRadius(t *testing.T) {
	var p1 Path
	p1.RoundRect(10, 20, 100, 50, 0)

	var p2 Path
	p2.Rect(10, 20, 100, 50)

	if len(p1.ops) != len(p2.ops) {
		t.Errorf("RoundRect(r=0) produced %d ops, Rect produced %d ops", len(p1.ops), len(p2.ops))
	}
}

func TestPathRoundRectClampedRadius(t *testing.T) {
	var p Path
	p.RoundRect(0, 0, 20, 10, 100)

	if len(p.ops) == 0 {
		t.Error("RoundRect with clamped radius should produce ops")
	}
	if p.ops[len(p.ops)-1].op != opClose {
		t.Error("RoundRect should end with close")
	}
}

func TestFlattenRoundRect(t *testing.T) {
	var p Path
	p.RoundRect(0, 0, 100, 50, 10)
	subPaths := p.flatten(defaultFlatness)
	if len(subPaths) != 1 {
		t.Fatalf("flatten round rect: got %d sub-paths, want 1", len(subPaths))
	}
	if len(subPaths[0]) < 10 {
		t.Errorf("flatten round rect: only %d points, expected more due to arcs", len(subPaths[0]))
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
