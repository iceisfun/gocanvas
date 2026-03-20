package gocanvas

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

// StrokeMode controls how line width interacts with transforms.
type StrokeMode uint8

const (
	// StrokeModeScreen applies line width in screen pixels.
	// The stroke width is constant regardless of the current transform.
	StrokeModeScreen StrokeMode = iota

	// StrokeModeWorld applies line width in world coordinates.
	// The stroke width scales with the current transform, matching
	// the HTML5 Canvas behavior.
	StrokeModeWorld
)

// TextAlign controls horizontal text alignment relative to the x coordinate.
type TextAlign uint8

const (
	// TextAlignLeft aligns the left edge of the text at x (default).
	TextAlignLeft TextAlign = iota

	// TextAlignCenter centers the text horizontally at x.
	TextAlignCenter

	// TextAlignRight aligns the right edge of the text at x.
	TextAlignRight
)

// TextBaseline controls vertical text alignment relative to the y coordinate.
type TextBaseline uint8

const (
	// TextBaselineAlphabetic positions y at the text baseline (default).
	TextBaselineAlphabetic TextBaseline = iota

	// TextBaselineTop positions y at the top of the em box.
	TextBaselineTop

	// TextBaselineMiddle positions y at the vertical center of the em box.
	TextBaselineMiddle

	// TextBaselineBottom positions y at the bottom of the em box.
	TextBaselineBottom
)

// CompositeOp specifies the compositing operation used when blending pixels.
type CompositeOp uint8

const (
	CompSourceOver      CompositeOp = iota // default: draw source over destination
	CompDestinationOver                    // draw behind existing content
	CompSourceIn                           // show source only where destination exists
	CompDestinationIn                      // keep destination only where source would draw
	CompSourceOut                          // show source only where destination is transparent
	CompDestinationOut                     // erase destination where source would draw
	CompLighter                            // additive blending
	CompCopy                               // replace entirely (no blending)
	CompXOR                                // XOR compositing
	CompMultiply                           // multiply channels
	CompScreen                             // screen blend
)

// drawState holds the mutable drawing state that is saved/restored.
type drawState struct {
	matrix      Matrix
	fill        color.RGBA
	stroke      color.RGBA
	lineWidth   float64
	lineCap     LineCap
	lineJoin    LineJoin
	miterLimit  float64
	globalAlpha float64
	strokeMode  StrokeMode

	// Dash pattern.
	lineDash       []float64
	lineDashOffset float64

	// Shadow.
	shadowColor   color.RGBA
	shadowBlur    float64
	shadowOffsetX float64
	shadowOffsetY float64

	// Font.
	fontFace     *FontFace
	textAlign    TextAlign
	textBaseline TextBaseline

	// Compositing.
	compositeOp CompositeOp

	// Clip mask (nil = no clipping).
	clip *image.Alpha

	// Gradients (nil = use solid color).
	fillGradient   Gradient
	strokeGradient Gradient
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
	cp := c.state
	if cp.lineDash != nil {
		dash := make([]float64, len(cp.lineDash))
		copy(dash, cp.lineDash)
		cp.lineDash = dash
	}
	if cp.clip != nil {
		dup := image.NewAlpha(cp.clip.Bounds())
		copy(dup.Pix, cp.clip.Pix)
		cp.clip = dup
	}
	c.stack = append(c.stack, cp)
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

// SetFillColor sets the fill color and clears any fill gradient.
func (c *Canvas) SetFillColor(col color.RGBA) {
	c.state.fill = col
	c.state.fillGradient = nil
}

// SetStrokeColor sets the stroke color and clears any stroke gradient.
func (c *Canvas) SetStrokeColor(col color.RGBA) {
	c.state.stroke = col
	c.state.strokeGradient = nil
}

// SetFillGradient sets a gradient as the fill style. The gradient coordinates
// are in the same coordinate space as drawing operations (world space).
// Clears any previously set fill color.
func (c *Canvas) SetFillGradient(g Gradient) {
	c.state.fillGradient = g
}

// SetStrokeGradient sets a gradient as the stroke style. The gradient
// coordinates are in the same coordinate space as drawing operations (world space).
// Clears any previously set stroke color.
func (c *Canvas) SetStrokeGradient(g Gradient) {
	c.state.strokeGradient = g
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

// SetStrokeMode sets how line width interacts with transforms.
func (c *Canvas) SetStrokeMode(mode StrokeMode) {
	c.state.strokeMode = mode
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

// --- Dash methods ---

// SetLineDash sets the dash pattern for stroke operations. Even indices are
// draw lengths, odd indices are gap lengths. An empty slice means solid.
// If the pattern has an odd length, it is doubled (per HTML5 Canvas spec).
func (c *Canvas) SetLineDash(pattern []float64) {
	if len(pattern) == 0 {
		c.state.lineDash = nil
		return
	}
	if len(pattern)%2 != 0 {
		pattern = append(pattern, pattern...)
	}
	dash := make([]float64, len(pattern))
	copy(dash, pattern)
	c.state.lineDash = dash
}

// LineDash returns the current dash pattern.
func (c *Canvas) LineDash() []float64 {
	if c.state.lineDash == nil {
		return nil
	}
	dash := make([]float64, len(c.state.lineDash))
	copy(dash, c.state.lineDash)
	return dash
}

// SetLineDashOffset sets the dash pattern offset.
func (c *Canvas) SetLineDashOffset(offset float64) {
	c.state.lineDashOffset = offset
}

// --- Shadow methods ---

// SetShadowColor sets the shadow color.
func (c *Canvas) SetShadowColor(col color.RGBA) {
	c.state.shadowColor = col
}

// SetShadowBlur sets the shadow blur radius.
func (c *Canvas) SetShadowBlur(radius float64) {
	if radius < 0 {
		radius = 0
	}
	c.state.shadowBlur = radius
}

// SetShadowOffset sets the shadow offset.
func (c *Canvas) SetShadowOffset(dx, dy float64) {
	c.state.shadowOffsetX = dx
	c.state.shadowOffsetY = dy
}

// --- Font methods ---

// SetFont sets the current font face for text rendering.
func (c *Canvas) SetFont(face *FontFace) {
	c.state.fontFace = face
}

// SetTextAlign sets the horizontal text alignment.
func (c *Canvas) SetTextAlign(align TextAlign) {
	c.state.textAlign = align
}

// SetTextBaseline sets the vertical text baseline.
func (c *Canvas) SetTextBaseline(baseline TextBaseline) {
	c.state.textBaseline = baseline
}

// SetCompositeOp sets the compositing operation for all drawing.
func (c *Canvas) SetCompositeOp(op CompositeOp) {
	c.state.compositeOp = op
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

// ArcTo adds an arc tangent to two lines defined by the current point,
// (x1, y1), and (x2, y2), with the given radius.
func (c *Canvas) ArcTo(x1, y1, x2, y2, radius float64) {
	c.path.ArcTo(x1, y1, x2, y2, radius)
}

// ClosePath closes the current sub-path.
func (c *Canvas) ClosePath() {
	c.path.Close()
}

// Rect adds a rectangular sub-path.
func (c *Canvas) Rect(x, y, w, h float64) {
	c.path.Rect(x, y, w, h)
}

// RoundRect adds a rounded rectangular sub-path to the current path.
func (c *Canvas) RoundRect(x, y, w, h, radius float64) {
	c.path.RoundRect(x, y, w, h, radius)
}

// --- Clipping ---

// Clip intersects the current clipping region with the current path.
// After calling Clip, only pixels inside the path will be affected by
// drawing operations. The current path is consumed.
func (c *Canvas) Clip() {
	subPaths := c.path.flatten(defaultFlatness)
	transformSubPaths(subPaths, c.state.matrix)
	edges := buildEdges(subPaths)

	mask := rasterizeMask(c.width, c.height, edges)

	if c.state.clip != nil {
		// Intersect: AND with existing clip.
		for i, a := range mask.Pix {
			mask.Pix[i] = uint8(uint32(a) * uint32(c.state.clip.Pix[i]) / 255)
		}
	}
	c.state.clip = mask
}

// ResetClip removes the current clipping region, so that drawing
// operations once again affect the entire canvas.
func (c *Canvas) ResetClip() {
	c.state.clip = nil
}

// --- Drawing methods ---

// clipDraw renders fn into a temporary buffer and composites the result
// through the current clip mask onto the canvas destination.
func (c *Canvas) clipDraw(fn func(dst *image.RGBA)) {
	if c.state.clip == nil {
		fn(c.dst)
		return
	}
	tmp := image.NewRGBA(c.dst.Bounds())
	fn(tmp)
	bounds := c.dst.Bounds()
	op := c.state.compositeOp
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			off := tmp.PixOffset(x, y)
			sa := uint32(tmp.Pix[off+3])
			if sa == 0 {
				continue
			}
			clipOff := c.state.clip.PixOffset(x, y)
			clipA := uint32(c.state.clip.Pix[clipOff])
			if clipA == 0 {
				continue
			}
			sr := uint32(tmp.Pix[off+0]) * clipA / 255
			sg := uint32(tmp.Pix[off+1]) * clipA / 255
			sb := uint32(tmp.Pix[off+2]) * clipA / 255
			sa = sa * clipA / 255
			blendPixelPremul(c.dst, x, y, sr, sg, sb, sa, op)
		}
	}
}

// fillEdges is the shared fill implementation respecting clip, shadow, gradient, and composite op.
func (c *Canvas) fillEdges(edges []edge) {
	op := c.state.compositeOp
	if c.state.fillGradient != nil {
		inv, ok := c.state.matrix.Invert()
		if !ok {
			return
		}
		if c.hasShadow() {
			c.renderWithShadow(func(dst *image.RGBA) {
				rasterizeGradientFill(dst, edges, c.state.fillGradient, inv, c.state.globalAlpha, op)
			})
		} else {
			c.clipDraw(func(dst *image.RGBA) {
				rasterizeGradientFill(dst, edges, c.state.fillGradient, inv, c.state.globalAlpha, op)
			})
		}
	} else {
		fillColor := c.applyAlpha(c.state.fill)
		if c.hasShadow() {
			c.renderWithShadow(func(dst *image.RGBA) {
				rasterizeFill(dst, edges, fillColor, op)
			})
		} else {
			c.clipDraw(func(dst *image.RGBA) {
				rasterizeFill(dst, edges, fillColor, op)
			})
		}
	}
}

// Fill fills the current path with the fill color or gradient.
func (c *Canvas) Fill() {
	subPaths := c.path.flatten(defaultFlatness)
	transformSubPaths(subPaths, c.state.matrix)
	edges := buildEdges(subPaths)
	c.fillEdges(edges)
}

// Stroke strokes the current path with the stroke color.
func (c *Canvas) Stroke() {
	subPaths := c.path.flatten(defaultFlatness)
	c.strokeSubPaths(subPaths)
}

// FillRect fills a rectangle without affecting the current path.
func (c *Canvas) FillRect(x, y, w, h float64) {
	var p Path
	p.Rect(x, y, w, h)
	subPaths := p.flatten(defaultFlatness)
	transformSubPaths(subPaths, c.state.matrix)
	edges := buildEdges(subPaths)
	c.fillEdges(edges)
}

// StrokeRect strokes a rectangle without affecting the current path.
func (c *Canvas) StrokeRect(x, y, w, h float64) {
	var p Path
	p.Rect(x, y, w, h)
	subPaths := p.flatten(defaultFlatness)
	c.strokeSubPaths(subPaths)
}

// FillRoundRect fills a rounded rectangle without affecting the current path.
func (c *Canvas) FillRoundRect(x, y, w, h, radius float64) {
	var p Path
	p.RoundRect(x, y, w, h, radius)
	subPaths := p.flatten(defaultFlatness)
	transformSubPaths(subPaths, c.state.matrix)
	edges := buildEdges(subPaths)
	c.fillEdges(edges)
}

// StrokeRoundRect strokes a rounded rectangle without affecting the current path.
func (c *Canvas) StrokeRoundRect(x, y, w, h, radius float64) {
	var p Path
	p.RoundRect(x, y, w, h, radius)
	subPaths := p.flatten(defaultFlatness)
	c.strokeSubPaths(subPaths)
}

// strokeSubPaths is the shared stroke implementation for both Stroke and StrokeRect.
// In StrokeModeScreen (default), line width is in screen pixels — constant regardless of transform.
// In StrokeModeWorld, line width is in world coordinates — scales with the transform.
func (c *Canvas) strokeSubPaths(subPaths [][]Point) {
	if c.state.strokeMode == StrokeModeWorld {
		// Stroke in world space, then transform to screen space.
		if len(c.state.lineDash) > 0 {
			subPaths = applyDash(subPaths, c.state.lineDash, c.state.lineDashOffset)
		}
		outlines := strokePath(subPaths, c.state.lineWidth, c.state.lineCap, c.state.lineJoin, c.state.miterLimit)
		transformSubPaths(outlines, c.state.matrix)
		c.rasterizeOutlines(outlines)
	} else {
		// Transform to screen space, then stroke at constant width.
		transformSubPaths(subPaths, c.state.matrix)
		if len(c.state.lineDash) > 0 {
			subPaths = applyDash(subPaths, c.state.lineDash, c.state.lineDashOffset)
		}
		outlines := strokePath(subPaths, c.state.lineWidth, c.state.lineCap, c.state.lineJoin, c.state.miterLimit)
		c.rasterizeOutlines(outlines)
	}
}

func (c *Canvas) rasterizeOutlines(outlines [][]Point) {
	edges := buildEdges(outlines)
	op := c.state.compositeOp
	if c.state.strokeGradient != nil {
		inv, ok := c.state.matrix.Invert()
		if !ok {
			return
		}
		if c.hasShadow() {
			c.renderWithShadow(func(dst *image.RGBA) {
				rasterizeGradientFill(dst, edges, c.state.strokeGradient, inv, c.state.globalAlpha, op)
			})
		} else {
			c.clipDraw(func(dst *image.RGBA) {
				rasterizeGradientFill(dst, edges, c.state.strokeGradient, inv, c.state.globalAlpha, op)
			})
		}
	} else {
		strokeColor := c.applyAlpha(c.state.stroke)
		if c.hasShadow() {
			c.renderWithShadow(func(dst *image.RGBA) {
				rasterizeFill(dst, edges, strokeColor, op)
			})
		} else {
			c.clipDraw(func(dst *image.RGBA) {
				rasterizeFill(dst, edges, strokeColor, op)
			})
		}
	}
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

// --- Pixel access ---

// SetPixel sets the color of a single pixel in screen coordinates.
// Coordinates outside the canvas bounds are ignored.
func (c *Canvas) SetPixel(x, y int, col color.RGBA) {
	if !(image.Point{x, y}).In(c.dst.Bounds()) {
		return
	}
	i := c.dst.PixOffset(x, y)
	c.dst.Pix[i+0] = col.R
	c.dst.Pix[i+1] = col.G
	c.dst.Pix[i+2] = col.B
	c.dst.Pix[i+3] = col.A
}

// GetPixel returns the color of a single pixel in screen coordinates.
// Coordinates outside the canvas bounds return transparent black.
func (c *Canvas) GetPixel(x, y int) color.RGBA {
	if !(image.Point{x, y}).In(c.dst.Bounds()) {
		return color.RGBA{}
	}
	i := c.dst.PixOffset(x, y)
	return color.RGBA{
		R: c.dst.Pix[i+0],
		G: c.dst.Pix[i+1],
		B: c.dst.Pix[i+2],
		A: c.dst.Pix[i+3],
	}
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
