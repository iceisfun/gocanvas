package gocanvas

import (
	"image"
	"image/color"
	"math"
	"sort"
)

// LineCap specifies the shape at the end of a stroked line.
type LineCap uint8

const (
	CapButt   LineCap = iota // flat edge flush with endpoint
	CapRound                 // semicircle
	CapSquare                // extends by half line width
)

// LineJoin specifies the shape at the junction of two line segments.
type LineJoin uint8

const (
	JoinMiter LineJoin = iota // sharp corner
	JoinRound                 // arc
	JoinBevel                 // flat cut
)

// edge represents a line segment for the scanline rasterizer.
type edge struct {
	x0, y0, x1, y1 float64
	dir             int // +1 if y increases, -1 if y decreases
}

// buildEdges converts flattened sub-paths into a list of edges for rasterization.
// Horizontal edges are skipped. Edges are oriented so that y0 < y1.
func buildEdges(subPaths [][]Point) []edge {
	var edges []edge
	for _, sp := range subPaths {
		for i := 0; i < len(sp)-1; i++ {
			p0 := sp[i]
			p1 := sp[i+1]
			if p0.Y == p1.Y {
				continue // skip horizontal edges
			}
			e := edge{}
			if p0.Y < p1.Y {
				e = edge{p0.X, p0.Y, p1.X, p1.Y, 1}
			} else {
				e = edge{p1.X, p1.Y, p0.X, p0.Y, -1}
			}
			edges = append(edges, e)
		}
	}
	return edges
}

// aaSamples is the number of vertical sub-pixel samples used for anti-aliasing.
const aaSamples = 8

// rasterizeFill fills edges onto the destination image using non-zero winding rule
// with anti-aliased edges.
func rasterizeFill(dst *image.RGBA, edges []edge, fill color.RGBA, op CompositeOp) {
	if len(edges) == 0 {
		return
	}

	bounds := dst.Bounds()

	// Find y range.
	yMin := math.Inf(1)
	yMax := math.Inf(-1)
	for _, e := range edges {
		if e.y0 < yMin {
			yMin = e.y0
		}
		if e.y1 > yMax {
			yMax = e.y1
		}
	}

	startY := int(math.Floor(yMin))
	endY := int(math.Ceil(yMax))
	if startY < bounds.Min.Y {
		startY = bounds.Min.Y
	}
	if endY > bounds.Max.Y {
		endY = bounds.Max.Y
	}

	// Find x range for coverage buffer sizing.
	xMinF := math.Inf(1)
	xMaxF := math.Inf(-1)
	for _, e := range edges {
		if e.x0 < xMinF {
			xMinF = e.x0
		}
		if e.x1 < xMinF {
			xMinF = e.x1
		}
		if e.x0 > xMaxF {
			xMaxF = e.x0
		}
		if e.x1 > xMaxF {
			xMaxF = e.x1
		}
	}
	xMinPx := int(math.Floor(xMinF)) - 1
	xMaxPx := int(math.Ceil(xMaxF)) + 1
	if xMinPx < bounds.Min.X {
		xMinPx = bounds.Min.X
	}
	if xMaxPx > bounds.Max.X {
		xMaxPx = bounds.Max.X
	}

	covLen := xMaxPx - xMinPx
	if covLen <= 0 {
		return
	}
	coverage := make([]uint32, covLen)

	// Pre-compute premultiplied source color.
	sa := uint32(fill.A)
	sr := uint32(fill.R) * sa / 255
	sg := uint32(fill.G) * sa / 255
	sb := uint32(fill.B) * sa / 255

	// Sub-scanline offsets within a pixel row.
	var subOffsets [aaSamples]float64
	for i := range aaSamples {
		subOffsets[i] = (float64(i) + 0.5) / float64(aaSamples)
	}

	var intercepts []edgeIntercept
	for y := startY; y < endY; y++ {
		for i := range coverage {
			coverage[i] = 0
		}

		for _, subOff := range subOffsets {
			scanY := float64(y) + subOff

			intercepts = intercepts[:0]
			for i := range edges {
				e := &edges[i]
				if scanY < e.y0 || scanY >= e.y1 {
					continue
				}
				t := (scanY - e.y0) / (e.y1 - e.y0)
				x := e.x0 + t*(e.x1-e.x0)
				intercepts = append(intercepts, edgeIntercept{x: x, dir: e.dir})
			}

			sort.Slice(intercepts, func(i, j int) bool {
				return intercepts[i].x < intercepts[j].x
			})

			winding := 0
			for i := 0; i < len(intercepts)-1; i++ {
				winding += intercepts[i].dir
				if winding != 0 {
					leftX := intercepts[i].x
					rightX := intercepts[i+1].x

					pxLeft := int(math.Floor(leftX))
					pxRight := int(math.Floor(rightX))

					if pxLeft < xMinPx {
						pxLeft = xMinPx
					}
					if pxRight >= xMaxPx {
						pxRight = xMaxPx - 1
					}

					for px := pxLeft; px <= pxRight; px++ {
						cLeft := float64(px)
						cRight := float64(px + 1)

						spanLeft := leftX
						if spanLeft < cLeft {
							spanLeft = cLeft
						}
						spanRight := rightX
						if spanRight > cRight {
							spanRight = cRight
						}

						if spanRight > spanLeft {
							frac := spanRight - spanLeft
							coverage[px-xMinPx] += uint32(frac * 256)
						}
					}
				}
			}
		}

		// Emit pixels with accumulated coverage.
		maxCov := uint32(aaSamples * 256)
		for i := range coverage {
			if coverage[i] == 0 {
				continue
			}

			px := i + xMinPx
			if px < bounds.Min.X || px >= bounds.Max.X {
				continue
			}

			cov := coverage[i]
			if cov >= maxCov {
				blendPixelPremul(dst, px, y, sr, sg, sb, sa, op)
			} else {
				covAlpha := sa * cov / maxCov
				covR := sr * cov / maxCov
				covG := sg * cov / maxCov
				covB := sb * cov / maxCov
				blendPixelPremul(dst, px, y, covR, covG, covB, covAlpha, op)
			}
		}
	}
}

type edgeIntercept struct {
	x   float64
	dir int
}

// blendPixelPremul composites a premultiplied source pixel onto the destination
// using the specified composite operation.
func blendPixelPremul(dst *image.RGBA, x, y int, sr, sg, sb, sa uint32, op CompositeOp) {
	off := dst.PixOffset(x, y)
	if off < 0 || off+3 >= len(dst.Pix) {
		return
	}

	// Fast path for source-over with full/zero alpha.
	if op == CompSourceOver {
		if sa == 255 {
			dst.Pix[off+0] = uint8(sr)
			dst.Pix[off+1] = uint8(sg)
			dst.Pix[off+2] = uint8(sb)
			dst.Pix[off+3] = 255
			return
		}
		if sa == 0 {
			return
		}
	}

	dr := uint32(dst.Pix[off+0])
	dg := uint32(dst.Pix[off+1])
	db := uint32(dst.Pix[off+2])
	da := uint32(dst.Pix[off+3])

	var outR, outG, outB, outA uint32

	switch op {
	case CompSourceOver:
		invA := 255 - sa
		outR = (sr*255 + dr*invA) / 255
		outG = (sg*255 + dg*invA) / 255
		outB = (sb*255 + db*invA) / 255
		outA = (sa*255 + da*invA) / 255

	case CompDestinationOver:
		invA := 255 - da
		outR = (dr*255 + sr*invA) / 255
		outG = (dg*255 + sg*invA) / 255
		outB = (db*255 + sb*invA) / 255
		outA = (da*255 + sa*invA) / 255

	case CompSourceIn:
		outR = sr * da / 255
		outG = sg * da / 255
		outB = sb * da / 255
		outA = sa * da / 255

	case CompDestinationIn:
		outR = dr * sa / 255
		outG = dg * sa / 255
		outB = db * sa / 255
		outA = da * sa / 255

	case CompSourceOut:
		invDA := 255 - da
		outR = sr * invDA / 255
		outG = sg * invDA / 255
		outB = sb * invDA / 255
		outA = sa * invDA / 255

	case CompDestinationOut:
		invSA := 255 - sa
		outR = dr * invSA / 255
		outG = dg * invSA / 255
		outB = db * invSA / 255
		outA = da * invSA / 255

	case CompLighter:
		outR = sr + dr
		if outR > 255 {
			outR = 255
		}
		outG = sg + dg
		if outG > 255 {
			outG = 255
		}
		outB = sb + db
		if outB > 255 {
			outB = 255
		}
		outA = sa + da
		if outA > 255 {
			outA = 255
		}

	case CompCopy:
		outR = sr
		outG = sg
		outB = sb
		outA = sa

	case CompXOR:
		invDA := 255 - da
		invSA := 255 - sa
		outR = (sr*invDA + dr*invSA) / 255
		outG = (sg*invDA + dg*invSA) / 255
		outB = (sb*invDA + db*invSA) / 255
		outA = (sa*invDA + da*invSA) / 255

	case CompMultiply:
		outR = sr * dr / 255
		outG = sg * dg / 255
		outB = sb * db / 255
		invA := 255 - sa
		outA = (sa*255 + da*invA) / 255

	case CompScreen:
		outR = sr + dr - sr*dr/255
		outG = sg + dg - sg*dg/255
		outB = sb + db - sb*db/255
		invA := 255 - sa
		outA = (sa*255 + da*invA) / 255

	default:
		invA := 255 - sa
		outR = (sr*255 + dr*invA) / 255
		outG = (sg*255 + dg*invA) / 255
		outB = (sb*255 + db*invA) / 255
		outA = (sa*255 + da*invA) / 255
	}

	dst.Pix[off+0] = uint8(outR)
	dst.Pix[off+1] = uint8(outG)
	dst.Pix[off+2] = uint8(outB)
	dst.Pix[off+3] = uint8(outA)
}

// rasterizeMask produces an alpha mask from edges using non-zero winding rule.
func rasterizeMask(width, height int, edges []edge) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, width, height))
	if len(edges) == 0 {
		return mask
	}

	yMin := math.Inf(1)
	yMax := math.Inf(-1)
	for _, e := range edges {
		if e.y0 < yMin {
			yMin = e.y0
		}
		if e.y1 > yMax {
			yMax = e.y1
		}
	}

	startY := int(math.Floor(yMin))
	endY := int(math.Ceil(yMax))
	if startY < 0 {
		startY = 0
	}
	if endY > height {
		endY = height
	}

	var intercepts []edgeIntercept
	for y := startY; y < endY; y++ {
		scanY := float64(y) + 0.5
		intercepts = intercepts[:0]
		for i := range edges {
			e := &edges[i]
			if scanY < e.y0 || scanY >= e.y1 {
				continue
			}
			t := (scanY - e.y0) / (e.y1 - e.y0)
			x := e.x0 + t*(e.x1-e.x0)
			intercepts = append(intercepts, edgeIntercept{x: x, dir: e.dir})
		}
		sort.Slice(intercepts, func(i, j int) bool {
			return intercepts[i].x < intercepts[j].x
		})
		winding := 0
		for i := 0; i < len(intercepts)-1; i++ {
			winding += intercepts[i].dir
			if winding != 0 {
				xStart := int(math.Ceil(intercepts[i].x - 0.5))
				xEnd := int(math.Ceil(intercepts[i+1].x - 0.5))
				if xStart < 0 {
					xStart = 0
				}
				if xEnd > width {
					xEnd = width
				}
				for x := xStart; x < xEnd; x++ {
					mask.Pix[y*mask.Stride+x] = 255
				}
			}
		}
	}
	return mask
}

// blendPixel blends a single non-premultiplied color onto the destination.
func blendPixel(dst *image.RGBA, x, y int, src color.RGBA, op CompositeOp) {
	sa := uint32(src.A)
	sr := uint32(src.R) * sa / 255
	sg := uint32(src.G) * sa / 255
	sb := uint32(src.B) * sa / 255
	blendPixelPremul(dst, x, y, sr, sg, sb, sa, op)
}

// strokePath converts a stroked polyline into a filled polygon outline.
func strokePath(subPaths [][]Point, width float64, cap LineCap, join LineJoin, miterLimit float64) [][]Point {
	halfW := width / 2
	var result [][]Point

	for _, sp := range subPaths {
		if len(sp) < 2 {
			continue
		}

		closed := false
		if len(sp) >= 3 && sp[0].X == sp[len(sp)-1].X && sp[0].Y == sp[len(sp)-1].Y {
			closed = true
		}

		var left, right []Point

		for i := 0; i < len(sp)-1; i++ {
			p0 := sp[i]
			p1 := sp[i+1]

			dx := p1.X - p0.X
			dy := p1.Y - p0.Y
			l := math.Sqrt(dx*dx + dy*dy)
			if l < 1e-10 {
				continue
			}

			// Perpendicular unit vector scaled by half width.
			nx := -dy / l * halfW
			ny := dx / l * halfW

			l0 := Point{p0.X + nx, p0.Y + ny}
			l1 := Point{p1.X + nx, p1.Y + ny}
			r0 := Point{p0.X - nx, p0.Y - ny}
			r1 := Point{p1.X - nx, p1.Y - ny}

			if i == 0 {
				left = append(left, l0)
				right = append(right, r0)
			} else {
				// Join with previous segment.
				addJoin(&left, l0, join, miterLimit, halfW, sp[i-1], p0, p1, 1)
				addJoin(&right, r0, join, miterLimit, halfW, sp[i-1], p0, p1, -1)
			}

			left = append(left, l1)
			right = append(right, r1)
		}

		if closed {
			// Join last segment to first.
			if len(sp) >= 3 {
				p0 := sp[0]
				p1 := sp[1]
				dx := p1.X - p0.X
				dy := p1.Y - p0.Y
				l := math.Sqrt(dx*dx + dy*dy)
				if l >= 1e-10 {
					nx := -dy / l * halfW
					ny := dx / l * halfW
					l0 := Point{p0.X + nx, p0.Y + ny}
					r0 := Point{p0.X - nx, p0.Y - ny}

					addJoin(&left, l0, join, miterLimit, halfW, sp[len(sp)-2], sp[0], sp[1], 1)
					addJoin(&right, r0, join, miterLimit, halfW, sp[len(sp)-2], sp[0], sp[1], -1)
				}
			}
		} else {
			// Add end caps.
			left = addCap(left, sp[len(sp)-1], sp[len(sp)-2], cap, halfW, false)
			right = addCap(right, sp[0], sp[1], cap, halfW, true)
		}

		// Build outline: left forward + right reversed.
		outline := make([]Point, 0, len(left)+len(right)+1)
		outline = append(outline, left...)
		for i := len(right) - 1; i >= 0; i-- {
			outline = append(outline, right[i])
		}
		if len(outline) > 0 {
			outline = append(outline, outline[0])
		}

		result = append(result, outline)
	}

	return result
}

// addJoin adds join geometry between two consecutive offset segments.
func addJoin(pts *[]Point, next Point, join LineJoin, miterLimit, halfW float64, _, curr, nextSeg Point, _ int) {
	switch join {
	case JoinBevel:
		*pts = append(*pts, next)
	case JoinRound:
		if len(*pts) == 0 {
			*pts = append(*pts, next)
			return
		}
		last := (*pts)[len(*pts)-1]
		addArcPoints(pts, curr, last, next, halfW)
	case JoinMiter:
		if len(*pts) == 0 {
			*pts = append(*pts, next)
			return
		}
		last := (*pts)[len(*pts)-1]

		mx := (last.X + next.X) / 2
		my := (last.Y + next.Y) / 2
		dx := mx - curr.X
		dy := my - curr.Y
		miterLen := math.Sqrt(dx*dx + dy*dy)

		if miterLen > miterLimit*halfW {
			*pts = append(*pts, next)
		} else {
			ix, iy, ok := lineIntersection(
				(*pts)[len(*pts)-2], last,
				next, Point{next.X + (nextSeg.X - curr.X), next.Y + (nextSeg.Y - curr.Y)},
			)
			if ok {
				*pts = append(*pts, Point{ix, iy})
			}
			*pts = append(*pts, next)
		}
	}
}

func lineIntersection(p1, p2, p3, p4 Point) (float64, float64, bool) {
	d1x := p2.X - p1.X
	d1y := p2.Y - p1.Y
	d2x := p4.X - p3.X
	d2y := p4.Y - p3.Y

	denom := d1x*d2y - d1y*d2x
	if math.Abs(denom) < 1e-10 {
		return 0, 0, false
	}

	t := ((p3.X-p1.X)*d2y - (p3.Y-p1.Y)*d2x) / denom
	return p1.X + t*d1x, p1.Y + t*d1y, true
}

func addArcPoints(pts *[]Point, center, from, to Point, radius float64) {
	a1 := math.Atan2(from.Y-center.Y, from.X-center.X)
	a2 := math.Atan2(to.Y-center.Y, to.X-center.X)

	diff := a2 - a1
	if diff > math.Pi {
		diff -= 2 * math.Pi
	} else if diff < -math.Pi {
		diff += 2 * math.Pi
	}

	steps := int(math.Ceil(math.Abs(diff) / (math.Pi / 8)))
	steps = max(steps, 1)

	step := diff / float64(steps)
	for i := 1; i <= steps; i++ {
		a := a1 + float64(i)*step
		*pts = append(*pts, Point{
			X: center.X + radius*math.Cos(a),
			Y: center.Y + radius*math.Sin(a),
		})
	}
}

func addCap(pts []Point, endpoint, adjacent Point, cap LineCap, halfW float64, _ bool) []Point {
	dx := endpoint.X - adjacent.X
	dy := endpoint.Y - adjacent.Y
	l := math.Sqrt(dx*dx + dy*dy)
	if l < 1e-10 {
		return pts
	}
	dx /= l
	dy /= l

	switch cap {
	case CapButt:
		// No extra geometry needed.
	case CapSquare:
		ext := Point{endpoint.X + dx*halfW, endpoint.Y + dy*halfW}
		nx := -dy * halfW
		ny := dx * halfW
		pts = append(pts,
			Point{ext.X + nx, ext.Y + ny},
			Point{ext.X - nx, ext.Y - ny},
		)
	case CapRound:
		nx := -dy * halfW
		ny := dx * halfW
		center := endpoint
		start := Point{endpoint.X + nx, endpoint.Y + ny}
		end := Point{endpoint.X - nx, endpoint.Y - ny}
		_ = start

		a1 := math.Atan2(ny, nx)
		steps := 8
		for i := 0; i <= steps; i++ {
			a := a1 + math.Pi*float64(i)/float64(steps)
			pts = append(pts, Point{
				X: center.X + halfW*math.Cos(a),
				Y: center.Y + halfW*math.Sin(a),
			})
		}
		_ = end
	}

	return pts
}

// rasterizeGradientFill fills edges using a gradient for color sampling.
func rasterizeGradientFill(dst *image.RGBA, edges []edge, grad Gradient, inv Matrix, alpha float64, op CompositeOp) {
	if len(edges) == 0 {
		return
	}

	bounds := dst.Bounds()

	yMin := math.Inf(1)
	yMax := math.Inf(-1)
	for _, e := range edges {
		if e.y0 < yMin {
			yMin = e.y0
		}
		if e.y1 > yMax {
			yMax = e.y1
		}
	}

	startY := int(math.Floor(yMin))
	endY := int(math.Ceil(yMax))
	if startY < bounds.Min.Y {
		startY = bounds.Min.Y
	}
	if endY > bounds.Max.Y {
		endY = bounds.Max.Y
	}

	var intercepts []edgeIntercept
	for y := startY; y < endY; y++ {
		scanY := float64(y) + 0.5

		intercepts = intercepts[:0]
		for i := range edges {
			e := &edges[i]
			if scanY < e.y0 || scanY >= e.y1 {
				continue
			}
			t := (scanY - e.y0) / (e.y1 - e.y0)
			x := e.x0 + t*(e.x1-e.x0)
			intercepts = append(intercepts, edgeIntercept{x: x, dir: e.dir})
		}

		sort.Slice(intercepts, func(i, j int) bool {
			return intercepts[i].x < intercepts[j].x
		})

		winding := 0
		for i := 0; i < len(intercepts)-1; i++ {
			winding += intercepts[i].dir
			if winding != 0 {
				xStart := int(math.Ceil(intercepts[i].x - 0.5))
				xEnd := int(math.Ceil(intercepts[i+1].x - 0.5))
				if xStart < bounds.Min.X {
					xStart = bounds.Min.X
				}
				if xEnd > bounds.Max.X {
					xEnd = bounds.Max.X
				}
				for x := xStart; x < xEnd; x++ {
					gx, gy := inv.TransformPoint(float64(x)+0.5, float64(y)+0.5)
					col := grad.ColorAt(gx, gy)
					if alpha < 1.0 {
						col.A = uint8(float64(col.A) * alpha)
					}
					sa := uint32(col.A)
					sr := uint32(col.R) * sa / 255
					sg := uint32(col.G) * sa / 255
					sb := uint32(col.B) * sa / 255
					blendPixelPremul(dst, x, y, sr, sg, sb, sa, op)
				}
			}
		}
	}
}

// transformSubPaths transforms all points in sub-paths by the given matrix.
func transformSubPaths(subPaths [][]Point, m Matrix) {
	for _, sp := range subPaths {
		for i := range sp {
			sp[i].X, sp[i].Y = m.TransformPoint(sp[i].X, sp[i].Y)
		}
	}
}
