package gocanvas

import (
	"image"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Font wraps a parsed OpenType font.
type Font struct {
	font *opentype.Font
}

// LoadFont parses TTF or OTF font data and returns a Font.
func LoadFont(data []byte) (*Font, error) {
	f, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return &Font{font: f}, nil
}

// LoadFontFile reads a font file and returns a Font.
func LoadFontFile(path string) (*Font, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFont(data)
}

// NewFace creates a FontFace at the given size in points (1pt = 1px at 72 DPI).
func (f *Font) NewFace(size float64) (*FontFace, error) {
	face, err := opentype.NewFace(f.font, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	metrics := face.Metrics()

	return &FontFace{
		face:    face,
		font:    f,
		size:    size,
		metrics: metrics,
		cache:   make(map[rune]*glyphEntry),
	}, nil
}

// FontFace represents a font at a specific size with a glyph cache.
type FontFace struct {
	face    font.Face
	font    *Font
	size    float64
	metrics font.Metrics
	cache   map[rune]*glyphEntry
}

// Size returns the font size in points.
func (ff *FontFace) Size() float64 {
	return ff.size
}

// Ascent returns the font ascent (distance from baseline to top) in pixels.
func (ff *FontFace) Ascent() float64 {
	return fixedToFloat(ff.metrics.Ascent)
}

// Descent returns the font descent (distance from baseline to bottom) in pixels.
func (ff *FontFace) Descent() float64 {
	return fixedToFloat(ff.metrics.Descent)
}

// Height returns the line height in pixels.
func (ff *FontFace) Height() float64 {
	return fixedToFloat(ff.metrics.Height)
}

type glyphEntry struct {
	mask    *image.Alpha
	bounds  fixed.Rectangle26_6
	advance fixed.Int26_6
}

// glyph returns the cached glyph entry for the given rune, rasterizing on miss.
func (ff *FontFace) glyph(r rune) *glyphEntry {
	if e, ok := ff.cache[r]; ok {
		return e
	}

	bounds, advance, ok := ff.face.GlyphBounds(r)
	if !ok {
		// Use replacement character.
		bounds, advance, _ = ff.face.GlyphBounds('\uFFFD')
	}

	// Rasterize the glyph into an alpha mask.
	w := (bounds.Max.X - bounds.Min.X).Ceil()
	h := (bounds.Max.Y - bounds.Min.Y).Ceil()
	if w <= 0 || h <= 0 {
		e := &glyphEntry{
			mask:    image.NewAlpha(image.Rect(0, 0, 0, 0)),
			bounds:  bounds,
			advance: advance,
		}
		ff.cache[r] = e
		return e
	}

	mask := image.NewAlpha(image.Rect(0, 0, w, h))
	d := font.Drawer{
		Dst:  image.NewRGBA(image.Rect(0, 0, w, h)),
		Src:  image.White,
		Face: ff.face,
	}

	// Position so that the glyph origin aligns correctly.
	d.Dot = fixed.Point26_6{
		X: -bounds.Min.X,
		Y: -bounds.Min.Y,
	}

	// Draw the glyph and extract alpha.
	d.DrawString(string(r))

	// Extract alpha from the RGBA destination.
	rgba := d.Dst.(*image.RGBA)
	for i := range mask.Pix {
		mask.Pix[i] = rgba.Pix[i*4+3]
	}

	e := &glyphEntry{
		mask:    mask,
		bounds:  bounds,
		advance: advance,
	}
	ff.cache[r] = e
	return e
}

func fixedToFloat(v fixed.Int26_6) float64 {
	return float64(v) / 64.0
}

