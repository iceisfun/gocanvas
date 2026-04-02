package gocanvas

import (
	"image/color"
	"math"
	"sort"
)

// Gradient is the interface implemented by LinearGradient and RadialGradient.
type Gradient interface {
	// ColorAt returns the interpolated color at the given pixel position.
	ColorAt(x, y float64) color.RGBA
}

// colorStop is a position (0-1) plus an RGBA color.
type colorStop struct {
	Position float64
	Color    color.RGBA
}

// LinearGradient defines a linear color gradient between two points.
type LinearGradient struct {
	X0, Y0, X1, Y1 float64
	stops           []colorStop
}

// NewLinearGradient creates a linear gradient from (x0, y0) to (x1, y1).
func NewLinearGradient(x0, y0, x1, y1 float64) *LinearGradient {
	return &LinearGradient{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// AddColorStop adds a color stop at the given position (0-1).
func (g *LinearGradient) AddColorStop(position float64, col color.RGBA) {
	g.stops = append(g.stops, colorStop{Position: position, Color: col})
	sort.Slice(g.stops, func(i, j int) bool {
		return g.stops[i].Position < g.stops[j].Position
	})
}

// ColorAt returns the interpolated gradient color at the given pixel position.
func (g *LinearGradient) ColorAt(x, y float64) color.RGBA {
	if len(g.stops) == 0 {
		return color.RGBA{}
	}

	dx := g.X1 - g.X0
	dy := g.Y1 - g.Y0
	lenSq := dx*dx + dy*dy
	if lenSq < 1e-20 {
		return g.stops[0].Color
	}

	// Project (x,y) onto the gradient line to get position t.
	t := ((x-g.X0)*dx + (y-g.Y0)*dy) / lenSq
	return interpolateStops(g.stops, t)
}

// RadialGradient defines a radial color gradient between two circles.
type RadialGradient struct {
	CX0, CY0, R0 float64 // start circle
	CX1, CY1, R1 float64 // end circle
	stops         []colorStop
}

// NewRadialGradient creates a radial gradient between two circles.
func NewRadialGradient(cx0, cy0, r0, cx1, cy1, r1 float64) *RadialGradient {
	return &RadialGradient{CX0: cx0, CY0: cy0, R0: r0, CX1: cx1, CY1: cy1, R1: r1}
}

// AddColorStop adds a color stop at the given position (0-1).
func (g *RadialGradient) AddColorStop(position float64, col color.RGBA) {
	g.stops = append(g.stops, colorStop{Position: position, Color: col})
	sort.Slice(g.stops, func(i, j int) bool {
		return g.stops[i].Position < g.stops[j].Position
	})
}

// ColorAt returns the interpolated gradient color at the given pixel position.
func (g *RadialGradient) ColorAt(x, y float64) color.RGBA {
	if len(g.stops) == 0 {
		return color.RGBA{}
	}

	// For concentric circles (same center), use simple distance-based interpolation.
	// For the general two-circle case, find t such that the point lies on the
	// circle at center(t) = (1-t)*c0 + t*c1 with radius r(t) = (1-t)*r0 + t*r1.
	//
	// Solve: |p - center(t)|^2 = r(t)^2
	// This is a quadratic in t.

	dcx := g.CX1 - g.CX0
	dcy := g.CY1 - g.CY0
	dr := g.R1 - g.R0
	px := x - g.CX0
	py := y - g.CY0

	a := dcx*dcx + dcy*dcy - dr*dr
	b := 2*(px*dcx+py*dcy) - 2*g.R0*dr
	c := px*px + py*py - g.R0*g.R0

	var t float64
	if math.Abs(a) < 1e-20 {
		// Linear case.
		if math.Abs(b) < 1e-20 {
			return g.stops[0].Color
		}
		t = -c / b
	} else {
		disc := b*b - 4*a*c
		if disc < 0 {
			// Point is outside both circles — use the nearest stop.
			return g.stops[len(g.stops)-1].Color
		}
		sqrtDisc := math.Sqrt(disc)
		t1 := (-b + sqrtDisc) / (2 * a)
		t2 := (-b - sqrtDisc) / (2 * a)

		// Pick the largest t that yields a non-negative radius.
		t = math.Max(t1, t2)
		r1 := g.R0 + t*dr
		if r1 < 0 {
			t = math.Min(t1, t2)
		}
	}

	return interpolateStops(g.stops, t)
}

// interpolateStops finds the color at position t by interpolating between stops.
// Values outside 0-1 are clamped to the nearest stop color.
func interpolateStops(stops []colorStop, t float64) color.RGBA {
	if len(stops) == 1 || t <= stops[0].Position {
		return stops[0].Color
	}
	if t >= stops[len(stops)-1].Position {
		return stops[len(stops)-1].Color
	}

	// Find the two stops that bracket t.
	for i := 0; i < len(stops)-1; i++ {
		s0 := stops[i]
		s1 := stops[i+1]
		if t >= s0.Position && t <= s1.Position {
			span := s1.Position - s0.Position
			if span < 1e-20 {
				return s0.Color
			}
			f := (t - s0.Position) / span
			return lerpColor(s0.Color, s1.Color, f)
		}
	}

	return stops[len(stops)-1].Color
}

// ConicGradient defines a conic (angular) color gradient around a center point.
type ConicGradient struct {
	CX, CY   float64 // center
	Rotation float64 // rotation offset, normalized to 0-1
	stops    []colorStop
}

// NewConicGradient creates a conic gradient centered at (cx, cy) with the
// given rotation in degrees. Colors sweep angularly around the center.
func NewConicGradient(cx, cy, degrees float64) *ConicGradient {
	// Normalize degrees to [0, 360), then to [0, 1).
	deg := math.Mod(degrees, 360)
	if deg < 0 {
		deg += 360
	}
	return &ConicGradient{CX: cx, CY: cy, Rotation: deg / 360}
}

// AddColorStop adds a color stop at the given position (0-1).
func (g *ConicGradient) AddColorStop(position float64, col color.RGBA) {
	g.stops = append(g.stops, colorStop{Position: position, Color: col})
	sort.Slice(g.stops, func(i, j int) bool {
		return g.stops[i].Position < g.stops[j].Position
	})
}

// ColorAt returns the interpolated gradient color at the given pixel position.
func (g *ConicGradient) ColorAt(x, y float64) color.RGBA {
	if len(g.stops) == 0 {
		return color.RGBA{}
	}
	a := math.Atan2(y-g.CY, x-g.CX)
	// Normalize from [-π, π] to [0, 1].
	t := (a + math.Pi) / (2 * math.Pi)
	t -= g.Rotation
	if t < 0 {
		t += 1
	}
	return interpolateStops(g.stops, t)
}

// lerpColor linearly interpolates between two colors.
func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + t*(float64(b.R)-float64(a.R)) + 0.5),
		G: uint8(float64(a.G) + t*(float64(b.G)-float64(a.G)) + 0.5),
		B: uint8(float64(a.B) + t*(float64(b.B)-float64(a.B)) + 0.5),
		A: uint8(float64(a.A) + t*(float64(b.A)-float64(a.A)) + 0.5),
	}
}
