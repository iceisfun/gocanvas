package gocanvas

import (
	"math"
	"testing"
)

const epsilon = 1e-10

func approxEq(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func matrixApproxEq(a, b Matrix) bool {
	for i := range a {
		if !approxEq(a[i], b[i]) {
			return false
		}
	}
	return true
}

func TestIdentity(t *testing.T) {
	m := Identity()
	want := Matrix{1, 0, 0, 0, 1, 0}
	if m != want {
		t.Errorf("Identity() = %v, want %v", m, want)
	}
}

func TestIdentityMultiply(t *testing.T) {
	m := TranslateMatrix(10, 20)
	result := Identity().Multiply(m)
	if !matrixApproxEq(result, m) {
		t.Errorf("Identity * M = %v, want %v", result, m)
	}

	result = m.Multiply(Identity())
	if !matrixApproxEq(result, m) {
		t.Errorf("M * Identity = %v, want %v", result, m)
	}
}

func TestTransformPoint(t *testing.T) {
	m := TranslateMatrix(10, 20)
	x, y := m.TransformPoint(5, 3)
	if !approxEq(x, 15) || !approxEq(y, 23) {
		t.Errorf("TransformPoint(5,3) = (%v,%v), want (15,23)", x, y)
	}
}

func TestScaleTransformPoint(t *testing.T) {
	m := ScaleMatrix(2, 3)
	x, y := m.TransformPoint(5, 10)
	if !approxEq(x, 10) || !approxEq(y, 30) {
		t.Errorf("ScaleTransformPoint = (%v,%v), want (10,30)", x, y)
	}
}

func TestRotate90(t *testing.T) {
	m := RotateMatrix(math.Pi / 2)
	x, y := m.TransformPoint(1, 0)
	if !approxEq(x, 0) || !approxEq(y, 1) {
		t.Errorf("Rotate90 TransformPoint(1,0) = (%v,%v), want (0,1)", x, y)
	}
}

func TestTranslateScale(t *testing.T) {
	// Scale then translate: first apply scale, then translate.
	m := TranslateMatrix(10, 10).Multiply(ScaleMatrix(2, 2))
	x, y := m.TransformPoint(5, 5)
	if !approxEq(x, 20) || !approxEq(y, 20) {
		t.Errorf("Translate*Scale TransformPoint(5,5) = (%v,%v), want (20,20)", x, y)
	}
}

func TestInvert(t *testing.T) {
	m := TranslateMatrix(10, 20).Multiply(ScaleMatrix(2, 3))
	inv, ok := m.Invert()
	if !ok {
		t.Fatal("Invert returned false for invertible matrix")
	}

	result := m.Multiply(inv)
	if !matrixApproxEq(result, Identity()) {
		t.Errorf("M * M^-1 = %v, want Identity", result)
	}
}

func TestInvertSingular(t *testing.T) {
	m := Matrix{0, 0, 0, 0, 0, 0}
	_, ok := m.Invert()
	if ok {
		t.Error("Invert returned true for singular matrix")
	}
}
