package gocanvas

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

// drawState holds the mutable drawing state that is saved/restored.
type drawState struct {
	matrix     Matrix
	fill       color.RGBA
	stroke     color.RGBA
	lineWidth  float64
	lineCap    LineCap
	lineJoin   LineJoin
	miterLimit float64
	globalAlpha float64
}

// Canvas provides a 2D drawing surface with an HTML5 Canvas-like API.
type Canvas struct {
	dst    *image.RGBA
	width  int
	height int
	state  drawState
	stack  []drawState
	path   Path
}

// New creates a new Canvas with the given dimensions.
// The canvas is initialized with a white background.
func New(width, height int) *Canvas {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with white.
	for i := 0; i < len(dst.Pix); i += 4 {
		dst.Pix[i+0] = 255
		dst.Pix[i+1] = 255
		dst.Pix[i+2] = 255
		dst.Pix[i+3] = 255
	}

	return &Canvas{
		dst:    dst,
		width:  width,
		height: height,
		state: drawState{
			matrix:      Identity(),
			fill:        color.RGBA{0, 0, 0, 255},
			stroke:      color.RGBA{0, 0, 0, 255},
			lineWidth:   1,
			lineCap:     CapButt,
			lineJoin:    JoinMiter,
			miterLimit:  10,
			globalAlpha: 1.0,
		},
	}
}

// Width returns the canvas width in pixels.
func (c *Canvas) Width() int { return c.width }

// Height returns the canvas height in pixels.
func (c *Canvas) Height() int { return c.height }

// Save pushes the current drawing state onto the state stack.
func (c *Canvas) Save() {
	c.stack = append(c.stack, c.state)
}

// Restore pops the most recently saved drawing state. No-op if the stack is empty.
func (c *Canvas) Restore() {
	if len(c.stack) == 0 {
		return
	}
	c.state = c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
}

// --- Transform methods ---

// Translate applies a translation to the current transform.
func (c *Canvas) Translate(tx, ty float64) {
	c.state.matrix = c.state.matrix.Multiply(TranslateMatrix(tx, ty))
}

// Scale applies a scaling to the current transform.
func (c *Canvas) Scale(sx, sy float64) {
	c.state.matrix = c.state.matrix.Multiply(ScaleMatrix(sx, sy))
}

// Rotate applies a rotation (in radians) to the current transform.
func (c *Canvas) Rotate(radians float64) {
	c.state.matrix = c.state.matrix.Multiply(RotateMatrix(radians))
}

// Transform multiplies the current transform by the given matrix values.
func (c *Canvas) Transform(m Matrix) {
	c.state.matrix = c.state.matrix.Multiply(m)
}

// SetTransform replaces the current transform with the given matrix.
func (c *Canvas) SetTransform(m Matrix) {
	c.state.matrix = m
}

// ResetTransform sets the current transform to identity.
func (c *Canvas) ResetTransform() {
	c.state.matrix = Identity()
}

// --- Style setters ---

// SetFillColor sets the fill color.
func (c *Canvas) SetFillColor(col color.RGBA) {
	c.state.fill = col
}

// SetStrokeColor sets the stroke color.
func (c *Canvas) SetStrokeColor(col color.RGBA) {
	c.state.stroke = col
}

// SetLineWidth sets the line width for stroke operations.
func (c *Canvas) SetLineWidth(w float64) {
	c.state.lineWidth = w
}

// SetLineCap sets the line cap style.
func (c *Canvas) SetLineCap(cap LineCap) {
	c.state.lineCap = cap
}

// SetLineJoin sets the line join style.
func (c *Canvas) SetLineJoin(join LineJoin) {
	c.state.lineJoin = join
}

// SetMiterLimit sets the miter limit for miter joins.
func (c *Canvas) SetMiterLimit(limit float64) {
	c.state.miterLimit = limit
}

// SetGlobalAlpha sets the global alpha (opacity) for all drawing operations.
func (c *Canvas) SetGlobalAlpha(a float64) {
	if a < 0 {
		a = 0
	}
	if a > 1 {
		a = 1
	}
	c.state.globalAlpha = a
}

// --- Path methods ---

// BeginPath clears the current path.
func (c *Canvas) BeginPath() {
	c.path.Reset()
}

// MoveTo starts a new sub-path at the given point.
func (c *Canvas) MoveTo(x, y float64) {
	c.path.MoveTo(x, y)
}

// LineTo adds a line segment to the given point.
func (c *Canvas) LineTo(x, y float64) {
	c.path.LineTo(x, y)
}

// QuadraticCurveTo adds a quadratic Bezier curve.
func (c *Canvas) QuadraticCurveTo(cpx, cpy, x, y float64) {
	c.path.QuadraticTo(cpx, cpy, x, y)
}

// BezierCurveTo adds a cubic Bezier curve.
func (c *Canvas) BezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y float64) {
	c.path.CubicTo(cp1x, cp1y, cp2x, cp2y, x, y)
}

// Arc adds an arc to the current path.
func (c *Canvas) Arc(cx, cy, r, startAngle, endAngle float64) {
	c.path.Arc(cx, cy, r, startAngle, endAngle)
}

// ClosePath closes the current sub-path.
func (c *Canvas) ClosePath() {
	c.path.Close()
}

// Rect adds a rectangular sub-path.
func (c *Canvas) Rect(x, y, w, h float64) {
	c.path.Rect(x, y, w, h)
}

// --- Drawing methods ---

// Fill fills the current path with the fill color.
func (c *Canvas) Fill() {
	subPaths := c.path.flatten(defaultFlatness)
	transformSubPaths(subPaths, c.state.matrix)
	fillColor := c.applyAlpha(c.state.fill)
	edges := buildEdges(subPaths)
	rasterizeFill(c.dst, edges, fillColor)
}

// Stroke strokes the current path with the stroke color.
func (c *Canvas) Stroke() {
	subPaths := c.path.flatten(defaultFlatness)
	transformSubPaths(subPaths, c.state.matrix)

	// Convert stroke to fill outline.
	outlines := strokePath(subPaths, c.state.lineWidth, c.state.lineCap, c.state.lineJoin, c.state.miterLimit)

	strokeColor := c.applyAlpha(c.state.stroke)
	edges := buildEdges(outlines)
	rasterizeFill(c.dst, edges, strokeColor)
}

// FillRect fills a rectangle without affecting the current path.
func (c *Canvas) FillRect(x, y, w, h float64) {
	var p Path
	p.Rect(x, y, w, h)
	subPaths := p.flatten(defaultFlatness)
	transformSubPaths(subPaths, c.state.matrix)
	fillColor := c.applyAlpha(c.state.fill)
	edges := buildEdges(subPaths)
	rasterizeFill(c.dst, edges, fillColor)
}

// StrokeRect strokes a rectangle without affecting the current path.
func (c *Canvas) StrokeRect(x, y, w, h float64) {
	var p Path
	p.Rect(x, y, w, h)
	subPaths := p.flatten(defaultFlatness)
	transformSubPaths(subPaths, c.state.matrix)
	outlines := strokePath(subPaths, c.state.lineWidth, c.state.lineCap, c.state.lineJoin, c.state.miterLimit)
	strokeColor := c.applyAlpha(c.state.stroke)
	edges := buildEdges(outlines)
	rasterizeFill(c.dst, edges, strokeColor)
}

// ClearRect sets all pixels in the rectangle to transparent black.
func (c *Canvas) ClearRect(x, y, w, h float64) {
	// Transform corners.
	x0, y0 := c.state.matrix.TransformPoint(x, y)
	x1, y1 := c.state.matrix.TransformPoint(x+w, y+h)

	// Clamp to bounds.
	bounds := c.dst.Bounds()
	ix0 := clampInt(int(x0), bounds.Min.X, bounds.Max.X)
	iy0 := clampInt(int(y0), bounds.Min.Y, bounds.Max.Y)
	ix1 := clampInt(int(x1), bounds.Min.X, bounds.Max.X)
	iy1 := clampInt(int(y1), bounds.Min.Y, bounds.Max.Y)

	if ix0 > ix1 {
		ix0, ix1 = ix1, ix0
	}
	if iy0 > iy1 {
		iy0, iy1 = iy1, iy0
	}

	for py := iy0; py < iy1; py++ {
		for px := ix0; px < ix1; px++ {
			off := c.dst.PixOffset(px, py)
			c.dst.Pix[off+0] = 0
			c.dst.Pix[off+1] = 0
			c.dst.Pix[off+2] = 0
			c.dst.Pix[off+3] = 0
		}
	}
}

// applyAlpha applies globalAlpha to a color.
func (c *Canvas) applyAlpha(col color.RGBA) color.RGBA {
	if c.state.globalAlpha >= 1.0 {
		return col
	}
	a := float64(col.A) * c.state.globalAlpha
	return color.RGBA{R: col.R, G: col.G, B: col.B, A: uint8(a)}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// --- Image access ---

// Image returns the underlying RGBA image.
func (c *Canvas) Image() *image.RGBA {
	return c.dst
}

// WritePNG encodes the canvas to PNG format and writes it to w.
func (c *Canvas) WritePNG(w io.Writer) error {
	return png.Encode(w, c.dst)
}

// SavePNG saves the canvas as a PNG file.
func (c *Canvas) SavePNG(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := c.WritePNG(f); err != nil {
		return err
	}
	return f.Close()
}
