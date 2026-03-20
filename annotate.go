package gocanvas

import (
	"image/color"
	"math"
)

// AnnotStyle configures the appearance of annotation primitives.
type AnnotStyle struct {
	StrokeColor color.RGBA // bounding box / polygon stroke color
	FillColor   color.RGBA // background color for labels
	TextColor   color.RGBA // text color
	Font        *Font      // font for text rendering
	FontSize    float64    // desired font size in points
	LineWidth   float64    // stroke width
	Padding     float64    // padding inside label boxes
}

// DefaultAnnotStyle returns a sensible default annotation style
// (green stroke, dark background, white text).
func DefaultAnnotStyle() AnnotStyle {
	return AnnotStyle{
		StrokeColor: color.RGBA{R: 0, G: 255, B: 0, A: 255},
		FillColor:   color.RGBA{R: 0, G: 0, B: 0, A: 180},
		TextColor:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
		FontSize:    14,
		LineWidth:   2,
		Padding:     4,
	}
}

// DrawAABB draws an axis-aligned bounding box, clamped to canvas bounds.
func DrawAABB(c *Canvas, x, y, w, h float64, style AnnotStyle) {
	// Clamp to canvas bounds.
	cw := float64(c.Width())
	ch := float64(c.Height())

	x = math.Max(0, x)
	y = math.Max(0, y)
	if x+w > cw {
		w = cw - x
	}
	if y+h > ch {
		h = ch - y
	}

	if w <= 0 || h <= 0 {
		return
	}

	c.Save()
	defer c.Restore()

	c.SetStrokeColor(style.StrokeColor)
	c.SetLineWidth(style.LineWidth)
	c.StrokeRect(x, y, w, h)
}

// DrawLabel draws a text label with a background at the given position,
// clamped to stay on screen.
func DrawLabel(c *Canvas, text string, x, y float64, style AnnotStyle) {
	if style.Font == nil {
		return
	}

	c.Save()
	defer c.Restore()

	face, err := style.Font.NewFace(style.FontSize)
	if err != nil {
		return
	}
	c.SetFont(face)

	m := c.MeasureText(text)
	pad := style.Padding
	boxW := m.Width + 2*pad
	boxH := m.Height + 2*pad

	cw := float64(c.Width())
	ch := float64(c.Height())

	// Clamp position to keep label on screen.
	if x+boxW > cw {
		x = cw - boxW
	}
	if x < 0 {
		x = 0
	}
	if y+boxH > ch {
		y = ch - boxH
	}
	if y < 0 {
		y = 0
	}

	// Draw background.
	c.SetFillColor(style.FillColor)
	c.FillRect(x, y, boxW, boxH)

	// Draw text.
	c.SetFillColor(style.TextColor)
	textX := x + pad
	textY := y + pad + m.Ascent
	c.FillText(text, textX, textY)
}

// DrawPolygon draws a polygon from the given points, with optional fill.
func DrawPolygon(c *Canvas, points []Point, style AnnotStyle) {
	if len(points) < 2 {
		return
	}

	c.Save()
	defer c.Restore()

	c.BeginPath()
	c.MoveTo(points[0].X, points[0].Y)
	for _, p := range points[1:] {
		c.LineTo(p.X, p.Y)
	}
	c.ClosePath()

	// Fill with semi-transparent version if fill color has alpha.
	if style.FillColor.A > 0 {
		c.SetFillColor(style.FillColor)
		c.Fill()
	}

	c.SetStrokeColor(style.StrokeColor)
	c.SetLineWidth(style.LineWidth)
	c.Stroke()
}

// DrawLabeledBox draws a bounding box with a text label.
// The label is placed above the box by default. If it would be off-screen,
// it is placed inside the box at the top. Text auto-fits if needed.
func DrawLabeledBox(c *Canvas, label string, x, y, w, h float64, style AnnotStyle) {
	// Draw the bounding box.
	DrawAABB(c, x, y, w, h, style)

	if style.Font == nil || label == "" {
		return
	}

	c.Save()
	defer c.Restore()

	// Determine the font face, fitting if needed.
	face, err := style.Font.NewFace(style.FontSize)
	if err != nil {
		return
	}

	pad := style.Padding
	maxLabelW := w - 2*pad
	if maxLabelW <= 0 {
		maxLabelW = w
	}

	c.SetFont(face)
	m := c.MeasureText(label)

	// Auto-fit if text is too wide.
	if m.Width > maxLabelW || m.Height > style.FontSize*2 {
		minSize := style.FontSize * 0.3
		fitted, err := c.FitText(label, maxLabelW, style.FontSize*2, style.Font, minSize, style.FontSize)
		if err == nil {
			face = fitted
			c.SetFont(face)
			m = c.MeasureText(label)
		}
	}

	boxW := m.Width + 2*pad
	boxH := m.Height + 2*pad

	// Determine label position.
	labelX := x
	labelY := y - boxH // above the box

	cw := float64(c.Width())
	ch := float64(c.Height())

	// Clamp horizontally.
	if labelX+boxW > cw {
		labelX = cw - boxW
	}
	if labelX < 0 {
		labelX = 0
	}

	// If label would be off-screen above, place inside the box at top.
	if labelY < 0 {
		labelY = y
	}

	// Clamp vertically.
	if labelY+boxH > ch {
		labelY = ch - boxH
	}
	if labelY < 0 {
		labelY = 0
	}

	// Draw label background.
	c.SetFillColor(style.FillColor)
	c.FillRect(labelX, labelY, boxW, boxH)

	// Draw label text.
	c.SetFillColor(style.TextColor)
	textX := labelX + pad
	textY := labelY + pad + m.Ascent
	c.FillText(label, textX, textY)
}
