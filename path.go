package gocanvas

import "math"

// Point represents a 2D point.
type Point struct {
	X, Y float64
}

type pathOpType uint8

const (
	opMoveTo pathOpType = iota
	opLineTo
	opQuadTo
	opCubicTo
	opClose
)

type pathOp struct {
	op     pathOpType
	points [3]Point // usage depends on op type
}

// Path represents a 2D vector path composed of sub-paths.
type Path struct {
	ops []pathOp
}

// Reset clears the path.
func (p *Path) Reset() {
	p.ops = p.ops[:0]
}

// MoveTo starts a new sub-path at the given point.
func (p *Path) MoveTo(x, y float64) {
	p.ops = append(p.ops, pathOp{op: opMoveTo, points: [3]Point{{x, y}}})
}

// LineTo adds a line segment to the given point.
func (p *Path) LineTo(x, y float64) {
	p.ops = append(p.ops, pathOp{op: opLineTo, points: [3]Point{{x, y}}})
}

// QuadraticTo adds a quadratic Bezier curve.
func (p *Path) QuadraticTo(cx, cy, x, y float64) {
	p.ops = append(p.ops, pathOp{op: opQuadTo, points: [3]Point{{cx, cy}, {x, y}}})
}

// CubicTo adds a cubic Bezier curve.
func (p *Path) CubicTo(cx1, cy1, cx2, cy2, x, y float64) {
	p.ops = append(p.ops, pathOp{op: opCubicTo, points: [3]Point{{cx1, cy1}, {cx2, cy2}, {x, y}}})
}

// Close closes the current sub-path.
func (p *Path) Close() {
	p.ops = append(p.ops, pathOp{op: opClose})
}

// Rect adds a rectangular sub-path.
func (p *Path) Rect(x, y, w, h float64) {
	p.MoveTo(x, y)
	p.LineTo(x+w, y)
	p.LineTo(x+w, y+h)
	p.LineTo(x, y+h)
	p.Close()
}

// RoundRect adds a rounded rectangular sub-path. The radius is clamped
// to min(w/2, h/2) so the arcs never exceed the rectangle dimensions.
func (p *Path) RoundRect(x, y, w, h, radius float64) {
	// Clamp radius.
	maxR := math.Min(w/2, h/2)
	if radius > maxR {
		radius = maxR
	}
	if radius < 0 {
		radius = 0
	}

	// If radius is zero, fall back to a plain rectangle.
	if radius == 0 {
		p.Rect(x, y, w, h)
		return
	}

	// Build the rounded rect clockwise from the top-left arc.
	p.Arc(x+radius, y+radius, radius, math.Pi, 3*math.Pi/2)
	p.LineTo(x+w-radius, y)
	p.Arc(x+w-radius, y+radius, radius, -math.Pi/2, 0)
	p.LineTo(x+w, y+h-radius)
	p.Arc(x+w-radius, y+h-radius, radius, 0, math.Pi/2)
	p.LineTo(x+radius, y+h)
	p.Arc(x+radius, y+h-radius, radius, math.Pi/2, math.Pi)
	p.LineTo(x, y+radius)
	p.Close()
}

// Circle adds a circular sub-path using cubic Bezier approximation.
func (p *Path) Circle(cx, cy, r float64) {
	p.Ellipse(cx, cy, r, r)
}

// Ellipse adds an elliptical sub-path using cubic Bezier approximation.
func (p *Path) Ellipse(cx, cy, rx, ry float64) {
	// Approximate with 4 cubic Bezier curves.
	// Magic number: (4/3)*tan(π/8) ≈ 0.5522847498
	const k = 0.5522847498307936

	kx := rx * k
	ky := ry * k

	p.MoveTo(cx+rx, cy)
	p.CubicTo(cx+rx, cy+ky, cx+kx, cy+ry, cx, cy+ry)
	p.CubicTo(cx-kx, cy+ry, cx-rx, cy+ky, cx-rx, cy)
	p.CubicTo(cx-rx, cy-ky, cx-kx, cy-ry, cx, cy-ry)
	p.CubicTo(cx+kx, cy-ry, cx+rx, cy-ky, cx+rx, cy)
	p.Close()
}

// ArcTo adds an arc tangent to two lines. It draws a straight line from the
// current point toward (x1, y1), then an arc of the given radius tangent to
// the lines (current->p1) and (p1->p2), ending at the tangent point on the
// second line. This matches the HTML5 Canvas arcTo specification.
func (p *Path) ArcTo(x1, y1, x2, y2, radius float64) {
	// Determine current point from the last op.
	var p0x, p0y float64
	hasPoint := false
	for i := len(p.ops) - 1; i >= 0; i-- {
		switch p.ops[i].op {
		case opMoveTo, opLineTo:
			p0x, p0y = p.ops[i].points[0].X, p.ops[i].points[0].Y
			hasPoint = true
		case opQuadTo:
			p0x, p0y = p.ops[i].points[1].X, p.ops[i].points[1].Y
			hasPoint = true
		case opCubicTo:
			p0x, p0y = p.ops[i].points[2].X, p.ops[i].points[2].Y
			hasPoint = true
		case opClose:
			for j := i - 1; j >= 0; j-- {
				if p.ops[j].op == opMoveTo {
					p0x, p0y = p.ops[j].points[0].X, p.ops[j].points[0].Y
					hasPoint = true
					break
				}
			}
		}
		if hasPoint {
			break
		}
	}
	if !hasPoint {
		p.MoveTo(x1, y1)
		return
	}

	// Vectors from p1 to p0 and from p1 to p2.
	v1x, v1y := p0x-x1, p0y-y1
	v2x, v2y := x2-x1, y2-y1

	len1 := math.Sqrt(v1x*v1x + v1y*v1y)
	len2 := math.Sqrt(v2x*v2x + v2y*v2y)
	if len1 < 1e-10 || len2 < 1e-10 {
		p.LineTo(x1, y1)
		return
	}

	v1x /= len1
	v1y /= len1
	v2x /= len2
	v2y /= len2

	cross := v1x*v2y - v1y*v2x
	dot := v1x*v2x + v1y*v2y

	if math.Abs(cross) < 1e-10 {
		p.LineTo(x1, y1)
		return
	}

	halfAngle := math.Acos(clampF(dot, -1, 1)) / 2
	tanDist := radius / math.Tan(halfAngle)

	tp1x := x1 + v1x*tanDist
	tp1y := y1 + v1y*tanDist
	tp2x := x1 + v2x*tanDist
	tp2y := y1 + v2y*tanDist

	centerDist := radius / math.Sin(halfAngle)

	bx, by := v1x+v2x, v1y+v2y
	blen := math.Sqrt(bx*bx + by*by)
	if blen < 1e-10 {
		p.LineTo(x1, y1)
		return
	}
	bx /= blen
	by /= blen

	cx := x1 + bx*centerDist
	cy := y1 + by*centerDist

	startAngle := math.Atan2(tp1y-cy, tp1x-cx)
	endAngle := math.Atan2(tp2y-cy, tp2x-cx)

	if cross > 0 {
		for endAngle > startAngle {
			endAngle -= 2 * math.Pi
		}
	} else {
		for endAngle < startAngle {
			endAngle += 2 * math.Pi
		}
	}

	p.LineTo(tp1x, tp1y)
	p.EllipticArc(cx, cy, radius, radius, startAngle, endAngle)
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Arc adds an arc to the path. Angles are in radians.
func (p *Path) Arc(cx, cy, r, startAngle, endAngle float64) {
	p.EllipticArc(cx, cy, r, r, startAngle, endAngle)
}

// EllipticArc adds an elliptical arc to the path. Angles are in radians.
func (p *Path) EllipticArc(cx, cy, rx, ry, startAngle, endAngle float64) {
	// Normalize so we go from startAngle to endAngle counterclockwise.
	diff := endAngle - startAngle
	if diff == 0 {
		return
	}

	// Split into arcs of at most π/2 each.
	n := int(math.Ceil(math.Abs(diff) / (math.Pi / 2)))
	step := diff / float64(n)

	sx, sy := math.Cos(startAngle)*rx+cx, math.Sin(startAngle)*ry+cy
	if len(p.ops) == 0 {
		p.MoveTo(sx, sy)
	} else {
		p.LineTo(sx, sy)
	}

	for i := range n {
		a1 := startAngle + float64(i)*step
		a2 := a1 + step
		arcToCubic(p, cx, cy, rx, ry, a1, a2)
	}
}

// arcToCubic adds a single cubic Bezier approximation of an arc segment.
func arcToCubic(p *Path, cx, cy, rx, ry, a1, a2 float64) {
	alpha := (a2 - a1) / 2
	cosAlpha := math.Cos(alpha)
	sinAlpha := math.Sin(alpha)

	// Handle the unit arc, then transform.
	// Control points for a unit arc from -alpha to +alpha.
	k := (4.0 / 3.0) * (1.0 - cosAlpha) / sinAlpha
	if math.IsNaN(k) || math.IsInf(k, 0) {
		return
	}

	mid := (a1 + a2) / 2

	cos1 := math.Cos(a1)
	sin1 := math.Sin(a1)
	cos2 := math.Cos(a2)
	sin2 := math.Sin(a2)
	_ = mid

	cp1x := cx + rx*(cos1-k*sin1)
	cp1y := cy + ry*(sin1+k*cos1)
	cp2x := cx + rx*(cos2+k*sin2)
	cp2y := cy + ry*(sin2-k*cos2)
	ex := cx + rx*cos2
	ey := cy + ry*sin2

	p.CubicTo(cp1x, cp1y, cp2x, cp2y, ex, ey)
}

const defaultFlatness = 0.25

// flatten converts the path into flattened sub-paths (line segments only).
// Each sub-path is a slice of points.
func (p *Path) flatten(tolerance float64) [][]Point {
	if tolerance <= 0 {
		tolerance = defaultFlatness
	}

	var subPaths [][]Point
	var current []Point
	var startX, startY float64
	var curX, curY float64

	for _, op := range p.ops {
		switch op.op {
		case opMoveTo:
			if len(current) > 0 {
				subPaths = append(subPaths, current)
			}
			curX, curY = op.points[0].X, op.points[0].Y
			startX, startY = curX, curY
			current = []Point{{curX, curY}}

		case opLineTo:
			curX, curY = op.points[0].X, op.points[0].Y
			current = append(current, Point{curX, curY})

		case opQuadTo:
			cx, cy := op.points[0].X, op.points[0].Y
			ex, ey := op.points[1].X, op.points[1].Y
			current = flattenQuadratic(current, curX, curY, cx, cy, ex, ey, tolerance, 0)
			curX, curY = ex, ey

		case opCubicTo:
			c1x, c1y := op.points[0].X, op.points[0].Y
			c2x, c2y := op.points[1].X, op.points[1].Y
			ex, ey := op.points[2].X, op.points[2].Y
			current = flattenCubic(current, curX, curY, c1x, c1y, c2x, c2y, ex, ey, tolerance, 0)
			curX, curY = ex, ey

		case opClose:
			if len(current) > 0 {
				if curX != startX || curY != startY {
					current = append(current, Point{startX, startY})
				}
				curX, curY = startX, startY
			}
		}
	}

	if len(current) > 0 {
		subPaths = append(subPaths, current)
	}

	return subPaths
}

const maxFlattenDepth = 16

func flattenQuadratic(pts []Point, x0, y0, cx, cy, x1, y1, tol float64, depth int) []Point {
	if depth >= maxFlattenDepth || isQuadFlat(x0, y0, cx, cy, x1, y1, tol) {
		return append(pts, Point{x1, y1})
	}

	// De Casteljau subdivision at t=0.5.
	mx01 := (x0 + cx) / 2
	my01 := (y0 + cy) / 2
	mx12 := (cx + x1) / 2
	my12 := (cy + y1) / 2
	mx := (mx01 + mx12) / 2
	my := (my01 + my12) / 2

	pts = flattenQuadratic(pts, x0, y0, mx01, my01, mx, my, tol, depth+1)
	pts = flattenQuadratic(pts, mx, my, mx12, my12, x1, y1, tol, depth+1)
	return pts
}

func isQuadFlat(x0, y0, cx, cy, x1, y1, tol float64) bool {
	// Distance from control point to the line from start to end.
	dx := x1 - x0
	dy := y1 - y0
	d := math.Abs((cx-x0)*dy-(cy-y0)*dx) / math.Sqrt(dx*dx+dy*dy+1e-20)
	return d <= tol
}

func flattenCubic(pts []Point, x0, y0, c1x, c1y, c2x, c2y, x1, y1, tol float64, depth int) []Point {
	if depth >= maxFlattenDepth || isCubicFlat(x0, y0, c1x, c1y, c2x, c2y, x1, y1, tol) {
		return append(pts, Point{x1, y1})
	}

	// De Casteljau subdivision at t=0.5.
	m01x := (x0 + c1x) / 2
	m01y := (y0 + c1y) / 2
	m12x := (c1x + c2x) / 2
	m12y := (c1y + c2y) / 2
	m23x := (c2x + x1) / 2
	m23y := (c2y + y1) / 2
	m012x := (m01x + m12x) / 2
	m012y := (m01y + m12y) / 2
	m123x := (m12x + m23x) / 2
	m123y := (m12y + m23y) / 2
	mx := (m012x + m123x) / 2
	my := (m012y + m123y) / 2

	pts = flattenCubic(pts, x0, y0, m01x, m01y, m012x, m012y, mx, my, tol, depth+1)
	pts = flattenCubic(pts, mx, my, m123x, m123y, m23x, m23y, x1, y1, tol, depth+1)
	return pts
}

func isCubicFlat(x0, y0, c1x, c1y, c2x, c2y, x1, y1, tol float64) bool {
	dx := x1 - x0
	dy := y1 - y0
	invLen := 1.0 / math.Sqrt(dx*dx+dy*dy+1e-20)

	d1 := math.Abs((c1x-x0)*dy-(c1y-y0)*dx) * invLen
	d2 := math.Abs((c2x-x0)*dy-(c2y-y0)*dx) * invLen

	return d1 <= tol && d2 <= tol
}
