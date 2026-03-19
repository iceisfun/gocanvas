package gocanvas

import "math"

// Matrix represents a 2D affine transformation as a 2x3 matrix.
//
// The layout is:
//
//	| a  b  tx |    [0]=a  [1]=b  [2]=tx
//	| c  d  ty |    [3]=c  [4]=d  [5]=ty
//	| 0  0   1 |    (implicit)
type Matrix [6]float64

// Identity returns the identity transformation matrix.
func Identity() Matrix {
	return Matrix{1, 0, 0, 0, 1, 0}
}

// TranslateMatrix returns a translation matrix.
func TranslateMatrix(tx, ty float64) Matrix {
	return Matrix{1, 0, tx, 0, 1, ty}
}

// ScaleMatrix returns a scaling matrix.
func ScaleMatrix(sx, sy float64) Matrix {
	return Matrix{sx, 0, 0, 0, sy, 0}
}

// RotateMatrix returns a rotation matrix for the given angle in radians.
func RotateMatrix(radians float64) Matrix {
	s, c := math.Sincos(radians)
	return Matrix{c, -s, 0, s, c, 0}
}

// SkewMatrix returns a skew matrix.
func SkewMatrix(sx, sy float64) Matrix {
	return Matrix{1, math.Tan(sx), 0, math.Tan(sy), 1, 0}
}

// Multiply returns the composition m * n.
// This applies n first, then m.
func (m Matrix) Multiply(n Matrix) Matrix {
	return Matrix{
		m[0]*n[0] + m[1]*n[3],
		m[0]*n[1] + m[1]*n[4],
		m[0]*n[2] + m[1]*n[5] + m[2],
		m[3]*n[0] + m[4]*n[3],
		m[3]*n[1] + m[4]*n[4],
		m[3]*n[2] + m[4]*n[5] + m[5],
	}
}

// TransformPoint applies the transformation to a point.
func (m Matrix) TransformPoint(x, y float64) (float64, float64) {
	return m[0]*x + m[1]*y + m[2], m[3]*x + m[4]*y + m[5]
}

// Determinant returns the determinant of the 2x2 linear portion.
func (m Matrix) Determinant() float64 {
	return m[0]*m[4] - m[1]*m[3]
}

// Invert returns the inverse of the matrix. Returns the zero matrix and false
// if the matrix is singular.
func (m Matrix) Invert() (Matrix, bool) {
	det := m.Determinant()
	if det == 0 {
		return Matrix{}, false
	}
	invDet := 1.0 / det
	return Matrix{
		m[4] * invDet,
		-m[1] * invDet,
		(m[1]*m[5] - m[4]*m[2]) * invDet,
		-m[3] * invDet,
		m[0] * invDet,
		(m[3]*m[2] - m[0]*m[5]) * invDet,
	}, true
}
