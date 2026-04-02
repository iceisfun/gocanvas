package gocanvas

import (
	"image"
	"image/color"
	"math"
	"strings"
	"unicode"
)

// TextMetrics contains measurements for rendered text.
type TextMetrics struct {
	Width   float64 // total advance width
	Ascent  float64 // distance from baseline to top
	Descent float64 // distance from baseline to bottom
	Height  float64 // Ascent + Descent
}

// MeasureText measures the given text with the current font face.
// Returns zero metrics if no font is set.
func (c *Canvas) MeasureText(text string) TextMetrics {
	ff := c.state.fontFace
	if ff == nil {
		return TextMetrics{}
	}
	return measureText(ff, text)
}

// measureText measures text with a specific font face.
func measureText(ff *FontFace, text string) TextMetrics {
	var totalAdvance float64
	for _, r := range text {
		g := ff.glyph(r)
		totalAdvance += fixedToFloat(g.advance)
	}

	ascent := ff.Ascent()
	descent := ff.Descent()

	return TextMetrics{
		Width:   totalAdvance,
		Ascent:  ascent,
		Descent: descent,
		Height:  ascent + descent,
	}
}

// alignText adjusts the (x, y) position based on textAlign and textBaseline.
func (c *Canvas) alignText(ff *FontFace, text string, x, y float64) (float64, float64) {
	if c.state.textAlign != TextAlignLeft {
		m := measureText(ff, text)
		switch c.state.textAlign {
		case TextAlignCenter:
			x -= m.Width / 2
		case TextAlignRight:
			x -= m.Width
		}
	}

	if c.state.textBaseline != TextBaselineAlphabetic {
		ascent := ff.Ascent()
		descent := ff.Descent()
		switch c.state.textBaseline {
		case TextBaselineTop:
			y += ascent
		case TextBaselineMiddle:
			y += (ascent - descent) / 2
		case TextBaselineBottom:
			y -= descent
		}
	}

	return x, y
}

// FillText renders text filled with the current fill color at the given position.
// The position (x, y) is adjusted by the current textAlign and textBaseline settings.
func (c *Canvas) FillText(text string, x, y float64) {
	ff := c.state.fontFace
	if ff == nil {
		return
	}

	x, y = c.alignText(ff, text, x, y)
	fillColor := c.applyAlpha(c.state.fill)

	if c.hasShadow() {
		c.renderWithShadow(func(dst *image.RGBA) {
			drawText(dst, ff, text, x, y, fillColor, c.state.matrix)
		})
	} else {
		drawText(c.dst, ff, text, x, y, fillColor, c.state.matrix)
	}
}

// StrokeText renders text outlined with the current stroke color at the given position.
// The position (x, y) is adjusted by the current textAlign and textBaseline settings.
func (c *Canvas) StrokeText(text string, x, y float64) {
	ff := c.state.fontFace
	if ff == nil {
		return
	}

	x, y = c.alignText(ff, text, x, y)
	strokeColor := c.applyAlpha(c.state.stroke)

	if c.hasShadow() {
		c.renderWithShadow(func(dst *image.RGBA) {
			drawTextStroke(dst, ff, text, x, y, strokeColor, c.state.matrix)
		})
	} else {
		drawTextStroke(c.dst, ff, text, x, y, strokeColor, c.state.matrix)
	}
}

// FitText determines the maximum font size that fits the given text within
// maxWidth and maxHeight. It searches between minSize and maxSize using binary search.
// Returns the font face at the fitted size.
func (c *Canvas) FitText(text string, maxWidth, maxHeight float64, f *Font, minSize, maxSize float64) (*FontFace, error) {
	return fitText(text, maxWidth, maxHeight, f, minSize, maxSize)
}

// FillTextFit finds the largest font size that fits the text within the
// rectangle (x, y, w, h) and draws it filled with the current fill color.
// Horizontal placement within the box respects the current textAlign setting.
// Vertical placement within the box respects the current textBaseline setting.
func (c *Canvas) FillTextFit(text string, x, y, w, h float64, f *Font) error {
	face, err := fitText(text, w, h, f, 1, h*2)
	if err != nil {
		return err
	}

	m := measureText(face, text)

	// Compute tx so alignment places the text within the box.
	var tx float64
	switch c.state.textAlign {
	case TextAlignCenter:
		tx = x + w/2
	case TextAlignRight:
		tx = x + w
	default:
		tx = x
	}

	// Compute ty so baseline places the text vertically centered in the box.
	var ty float64
	switch c.state.textBaseline {
	case TextBaselineTop:
		ty = y + (h-m.Height)/2
	case TextBaselineMiddle:
		ty = y + h/2
	case TextBaselineBottom:
		ty = y + (h+m.Height)/2
	default: // Alphabetic
		ty = y + (h-m.Height)/2 + m.Ascent
	}

	c.SetFont(face)
	c.FillText(text, tx, ty)
	return nil
}

// WordWrap splits text into lines that fit within maxWidth using the current font.
// Returns nil if no font is set.
func (c *Canvas) WordWrap(text string, maxWidth float64) []string {
	ff := c.state.fontFace
	if ff == nil {
		return nil
	}
	return wordWrap(ff, text, maxWidth)
}

func splitOnSpace(x string) []string {
	var result []string
	pi := 0
	ps := false
	for i, c := range x {
		s := unicode.IsSpace(c)
		if s != ps && i > 0 {
			result = append(result, x[pi:i])
			pi = i
		}
		ps = s
	}
	result = append(result, x[pi:])
	return result
}

func wordWrap(ff *FontFace, s string, width float64) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		fields := splitOnSpace(line)
		if len(fields)%2 == 1 {
			fields = append(fields, "")
		}
		x := ""
		for i := 0; i < len(fields); i += 2 {
			m := measureText(ff, x+fields[i])
			if m.Width > width {
				if x == "" {
					result = append(result, fields[i])
					x = ""
					continue
				} else {
					result = append(result, x)
					x = ""
				}
			}
			x += fields[i] + fields[i+1]
		}
		if x != "" {
			result = append(result, x)
		}
	}
	for i, line := range result {
		result[i] = strings.TrimSpace(line)
	}
	return result
}

// FillTextWrapped renders word-wrapped text within the given width.
// lineSpacing is a multiplier on the font height (1.0 = single space, 1.5 = 1.5x, etc).
// Text alignment respects the current textAlign setting.
func (c *Canvas) FillTextWrapped(text string, x, y, width, lineSpacing float64) {
	ff := c.state.fontFace
	if ff == nil {
		return
	}
	lines := wordWrap(ff, text, width)
	lineHeight := ff.Ascent() + ff.Descent()
	spacing := lineHeight * lineSpacing

	for i, line := range lines {
		ly := y + float64(i)*spacing + ff.Ascent()
		var lx float64
		switch c.state.textAlign {
		case TextAlignCenter:
			lx = x + width/2
		case TextAlignRight:
			lx = x + width
		default:
			lx = x
		}
		c.FillText(line, lx, ly)
	}
}

// MeasureTextWrapped measures the dimensions of word-wrapped text.
func (c *Canvas) MeasureTextWrapped(text string, width, lineSpacing float64) (w, h float64) {
	ff := c.state.fontFace
	if ff == nil {
		return 0, 0
	}
	lines := wordWrap(ff, text, width)
	lineHeight := ff.Ascent() + ff.Descent()
	spacing := lineHeight * lineSpacing

	maxW := 0.0
	for _, line := range lines {
		m := measureText(ff, line)
		if m.Width > maxW {
			maxW = m.Width
		}
	}
	h = float64(len(lines)) * spacing
	return maxW, h
}

func fitText(text string, maxWidth, maxHeight float64, f *Font, minSize, maxSize float64) (*FontFace, error) {
	var bestFace *FontFace

	lo, hi := minSize, maxSize
	for hi-lo > 0.5 {
		mid := (lo + hi) / 2
		face, err := f.NewFace(mid)
		if err != nil {
			return nil, err
		}

		m := measureText(face, text)
		if m.Width <= maxWidth && m.Height <= maxHeight {
			bestFace = face
			lo = mid
		} else {
			hi = mid
		}
	}

	if bestFace == nil {
		var err error
		bestFace, err = f.NewFace(lo)
		if err != nil {
			return nil, err
		}
	}

	return bestFace, nil
}

// drawText renders text glyphs onto the destination with fill color, respecting the transform.
func drawText(dst *image.RGBA, ff *FontFace, text string, x, y float64, col color.RGBA, m Matrix) {
	curX := x
	for _, r := range text {
		g := ff.glyph(r)
		if g.mask.Bounds().Empty() {
			curX += fixedToFloat(g.advance)
			continue
		}

		// Glyph origin in text space.
		gx := curX + fixedToFloat(g.bounds.Min.X)
		gy := y + fixedToFloat(g.bounds.Min.Y)

		drawGlyphTransformed(dst, g.mask, gx, gy, col, m)
		curX += fixedToFloat(g.advance)
	}
}

// drawTextStroke renders text outlines by drawing only edge pixels of glyph masks.
func drawTextStroke(dst *image.RGBA, ff *FontFace, text string, x, y float64, col color.RGBA, m Matrix) {
	curX := x
	for _, r := range text {
		g := ff.glyph(r)
		if g.mask.Bounds().Empty() {
			curX += fixedToFloat(g.advance)
			continue
		}

		gx := curX + fixedToFloat(g.bounds.Min.X)
		gy := y + fixedToFloat(g.bounds.Min.Y)

		drawGlyphStrokeTransformed(dst, g.mask, gx, gy, col, m)
		curX += fixedToFloat(g.advance)
	}
}

// drawGlyphTransformed composites a glyph mask onto dst with transform-aware positioning.
// Uses inverse mapping to avoid gaps when scaling or rotating.
func drawGlyphTransformed(dst *image.RGBA, mask *image.Alpha, gx, gy float64, col color.RGBA, m Matrix) {
	inv, ok := m.Invert()
	if !ok {
		return
	}

	dstBounds := dst.Bounds()
	startX, startY, endX, endY := glyphDestBounds(mask.Bounds(), gx, gy, m, dstBounds)

	for dy := startY; dy < endY; dy++ {
		for dx := startX; dx < endX; dx++ {
			sx, sy := inv.TransformPoint(float64(dx)+0.5, float64(dy)+0.5)
			mx := int(math.Floor(sx - gx))
			my := int(math.Floor(sy - gy))

			a := maskAt(mask, mx, my)
			if a == 0 {
				continue
			}
			blendGlyphPixel(dst, dx, dy, col, a)
		}
	}
}

// drawGlyphStrokeTransformed renders only edge pixels of a glyph for a stroke effect.
// Uses inverse mapping to avoid gaps when scaling or rotating.
func drawGlyphStrokeTransformed(dst *image.RGBA, mask *image.Alpha, gx, gy float64, col color.RGBA, m Matrix) {
	inv, ok := m.Invert()
	if !ok {
		return
	}

	b := mask.Bounds()
	dstBounds := dst.Bounds()
	startX, startY, endX, endY := glyphDestBounds(b, gx, gy, m, dstBounds)

	for dy := startY; dy < endY; dy++ {
		for dx := startX; dx < endX; dx++ {
			sx, sy := inv.TransformPoint(float64(dx)+0.5, float64(dy)+0.5)
			mx := int(math.Floor(sx - gx))
			my := int(math.Floor(sy - gy))

			a := maskAt(mask, mx, my)
			if a == 0 {
				continue
			}

			// Check if this is an edge pixel (any neighbor is transparent).
			isEdge := mx <= b.Min.X || mx >= b.Max.X-1 || my <= b.Min.Y || my >= b.Max.Y-1
			if !isEdge {
				for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
					if maskAt(mask, mx+d[0], my+d[1]) == 0 {
						isEdge = true
						break
					}
				}
			}
			if !isEdge {
				continue
			}

			blendGlyphPixel(dst, dx, dy, col, a)
		}
	}
}

// glyphDestBounds computes the destination bounding box for a transformed glyph,
// clamped to the destination image bounds.
func glyphDestBounds(b image.Rectangle, gx, gy float64, m Matrix, dstBounds image.Rectangle) (startX, startY, endX, endY int) {
	corners := [4][2]float64{
		{gx + float64(b.Min.X), gy + float64(b.Min.Y)},
		{gx + float64(b.Max.X), gy + float64(b.Min.Y)},
		{gx + float64(b.Min.X), gy + float64(b.Max.Y)},
		{gx + float64(b.Max.X), gy + float64(b.Max.Y)},
	}

	minDX, minDY := math.Inf(1), math.Inf(1)
	maxDX, maxDY := math.Inf(-1), math.Inf(-1)
	for _, c := range corners {
		dx, dy := m.TransformPoint(c[0], c[1])
		minDX = min(minDX, dx)
		minDY = min(minDY, dy)
		maxDX = max(maxDX, dx)
		maxDY = max(maxDY, dy)
	}

	startX = max(int(math.Floor(minDX)), dstBounds.Min.X)
	startY = max(int(math.Floor(minDY)), dstBounds.Min.Y)
	endX = min(int(math.Ceil(maxDX))+1, dstBounds.Max.X)
	endY = min(int(math.Ceil(maxDY))+1, dstBounds.Max.Y)
	return
}

// maskAt returns the alpha value at (mx, my) in the mask, or 0 if out of bounds.
func maskAt(mask *image.Alpha, mx, my int) uint8 {
	b := mask.Bounds()
	if mx < b.Min.X || mx >= b.Max.X || my < b.Min.Y || my >= b.Max.Y {
		return 0
	}
	return mask.AlphaAt(mx, my).A
}

// blendGlyphPixel blends a colored pixel with the given alpha coverage.
func blendGlyphPixel(dst *image.RGBA, x, y int, col color.RGBA, coverage uint8) {
	if coverage == 0 {
		return
	}
	sa := uint32(col.A) * uint32(coverage) / 255
	sr := uint32(col.R) * sa / 255
	sg := uint32(col.G) * sa / 255
	sb := uint32(col.B) * sa / 255
	blendPixelPremul(dst, x, y, sr, sg, sb, sa, CompSourceOver)
}
