package gocanvas

import (
	"image"
	"math"
)

// DrawImage draws a region of the source image onto the canvas.
// (sx, sy, sw, sh) defines the source rectangle within img.
// (dx, dy, dw, dh) defines the destination rectangle on the canvas.
// The current transform, globalAlpha, and shadow settings are applied.
func (c *Canvas) DrawImage(img image.Image, sx, sy, sw, sh, dx, dy, dw, dh float64) {
	if sw <= 0 || sh <= 0 || dw <= 0 || dh <= 0 {
		return
	}

	alpha := c.state.globalAlpha

	if c.hasShadow() {
		c.renderWithShadow(func(dst *image.RGBA) {
			drawImageTransformed(dst, img, sx, sy, sw, sh, dx, dy, dw, dh, alpha, c.state.matrix)
		})
	} else {
		drawImageTransformed(c.dst, img, sx, sy, sw, sh, dx, dy, dw, dh, alpha, c.state.matrix)
	}
}

func drawImageTransformed(dst *image.RGBA, img image.Image, sx, sy, sw, sh, dx, dy, dw, dh, alpha float64, m Matrix) {
	inv, ok := m.Invert()
	if !ok {
		return
	}

	dstBounds := dst.Bounds()

	// Compute bounding box of the transformed destination rectangle.
	corners := [4][2]float64{
		{dx, dy},
		{dx + dw, dy},
		{dx, dy + dh},
		{dx + dw, dy + dh},
	}

	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, c := range corners {
		px, py := m.TransformPoint(c[0], c[1])
		minX = min(minX, px)
		minY = min(minY, py)
		maxX = max(maxX, px)
		maxY = max(maxY, py)
	}

	startX := max(int(math.Floor(minX)), dstBounds.Min.X)
	startY := max(int(math.Floor(minY)), dstBounds.Min.Y)
	endX := min(int(math.Ceil(maxX))+1, dstBounds.Max.X)
	endY := min(int(math.Ceil(maxY))+1, dstBounds.Max.Y)

	imgBounds := img.Bounds()

	// Fast path for *image.RGBA sources.
	if rgba, ok := img.(*image.RGBA); ok {
		drawImageRGBA(dst, rgba, sx, sy, sw, sh, dx, dy, dw, dh, alpha, inv, startX, startY, endX, endY, imgBounds)
		return
	}

	// Generic path for any image.Image.
	for py := startY; py < endY; py++ {
		for px := startX; px < endX; px++ {
			cx, cy := inv.TransformPoint(float64(px)+0.5, float64(py)+0.5)

			u := (cx - dx) / dw
			v := (cy - dy) / dh
			if u < 0 || u >= 1 || v < 0 || v >= 1 {
				continue
			}

			ix := int(math.Floor(sx + u*sw))
			iy := int(math.Floor(sy + v*sh))
			if ix < imgBounds.Min.X || ix >= imgBounds.Max.X || iy < imgBounds.Min.Y || iy >= imgBounds.Max.Y {
				continue
			}

			r, g, b, a := img.At(ix, iy).RGBA()
			// RGBA() returns premultiplied values in [0, 65535].
			sa := uint32(a >> 8)
			if sa == 0 {
				continue
			}
			sr := uint32(r >> 8)
			sg := uint32(g >> 8)
			sb := uint32(b >> 8)

			if alpha < 1.0 {
				sa = uint32(float64(sa) * alpha)
				sr = uint32(float64(sr) * alpha)
				sg = uint32(float64(sg) * alpha)
				sb = uint32(float64(sb) * alpha)
			}

			blendPixelPremul(dst, px, py, sr, sg, sb, sa)
		}
	}
}

func drawImageRGBA(dst, src *image.RGBA, sx, sy, sw, sh, dx, dy, dw, dh, alpha float64, inv Matrix, startX, startY, endX, endY int, imgBounds image.Rectangle) {
	for py := startY; py < endY; py++ {
		for px := startX; px < endX; px++ {
			cx, cy := inv.TransformPoint(float64(px)+0.5, float64(py)+0.5)

			u := (cx - dx) / dw
			v := (cy - dy) / dh
			if u < 0 || u >= 1 || v < 0 || v >= 1 {
				continue
			}

			ix := int(math.Floor(sx + u*sw))
			iy := int(math.Floor(sy + v*sh))
			if ix < imgBounds.Min.X || ix >= imgBounds.Max.X || iy < imgBounds.Min.Y || iy >= imgBounds.Max.Y {
				continue
			}

			off := src.PixOffset(ix, iy)
			sa := uint32(src.Pix[off+3])
			if sa == 0 {
				continue
			}
			// image.RGBA stores premultiplied values.
			sr := uint32(src.Pix[off+0])
			sg := uint32(src.Pix[off+1])
			sb := uint32(src.Pix[off+2])

			if alpha < 1.0 {
				sa = uint32(float64(sa) * alpha)
				sr = uint32(float64(sr) * alpha)
				sg = uint32(float64(sg) * alpha)
				sb = uint32(float64(sb) * alpha)
			}

			blendPixelPremul(dst, px, py, sr, sg, sb, sa)
		}
	}
}
