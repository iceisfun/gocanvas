package gocanvas

import (
	"image"
	"image/color"
	"math"
)

// hasShadow returns true if the current state has an active shadow.
func (c *Canvas) hasShadow() bool {
	return c.state.shadowColor.A > 0 &&
		(c.state.shadowBlur > 0 || c.state.shadowOffsetX != 0 || c.state.shadowOffsetY != 0)
}

// renderWithShadow renders a shadow pass, then the actual content.
// The draw function rasterizes onto the given destination buffer.
func (c *Canvas) renderWithShadow(draw func(dst *image.RGBA)) {
	// Render shape to temporary buffer.
	tmp := image.NewRGBA(c.dst.Bounds())
	draw(tmp)

	// Apply blur if needed.
	if c.state.shadowBlur > 0 {
		radius := int(math.Ceil(c.state.shadowBlur))
		boxBlurRGBA(tmp, radius)
	}

	// Recolor: replace RGB with shadow color, keep blurred alpha.
	recolorAlpha(tmp, c.state.shadowColor)

	// Apply global alpha to shadow.
	shadowAlpha := c.state.globalAlpha

	// Composite shadow onto canvas at offset.
	offX := int(math.Round(c.state.shadowOffsetX))
	offY := int(math.Round(c.state.shadowOffsetY))
	compositeOver(c.dst, tmp, offX, offY, shadowAlpha)

	// Render actual content on top.
	draw(c.dst)
}

// recolorAlpha replaces all pixel colors with col, preserving the alpha channel.
func recolorAlpha(img *image.RGBA, col color.RGBA) {
	pix := img.Pix
	for i := 0; i < len(pix); i += 4 {
		a := pix[i+3]
		if a == 0 {
			continue
		}
		// Apply shadow color's alpha to the shape's alpha.
		ea := uint32(a) * uint32(col.A) / 255
		pix[i+0] = col.R
		pix[i+1] = col.G
		pix[i+2] = col.B
		pix[i+3] = uint8(ea)
	}
}

// compositeOver composites src onto dst at the given offset using source-over.
func compositeOver(dst, src *image.RGBA, offsetX, offsetY int, alpha float64) {
	sb := src.Bounds()
	db := dst.Bounds()

	for sy := sb.Min.Y; sy < sb.Max.Y; sy++ {
		dy := sy + offsetY
		if dy < db.Min.Y || dy >= db.Max.Y {
			continue
		}
		for sx := sb.Min.X; sx < sb.Max.X; sx++ {
			dx := sx + offsetX
			if dx < db.Min.X || dx >= db.Max.X {
				continue
			}

			sOff := src.PixOffset(sx, sy)
			sa := uint32(src.Pix[sOff+3])
			if sa == 0 {
				continue
			}

			// Apply external alpha.
			if alpha < 1.0 {
				sa = uint32(float64(sa) * alpha)
			}
			sr := uint32(src.Pix[sOff+0]) * sa / 255
			sg := uint32(src.Pix[sOff+1]) * sa / 255
			sb := uint32(src.Pix[sOff+2]) * sa / 255

			blendPixelPremul(dst, dx, dy, sr, sg, sb, sa, CompSourceOver)
		}
	}
}

// boxBlurRGBA applies a 3-pass box blur to approximate Gaussian blur.
func boxBlurRGBA(img *image.RGBA, radius int) {
	if radius <= 0 {
		return
	}
	// 3 passes of box blur approximates Gaussian.
	for range 3 {
		boxBlurHorizontal(img, radius)
		boxBlurVertical(img, radius)
	}
}

func boxBlurHorizontal(img *image.RGBA, radius int) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	div := 2*radius + 1

	tmp := make([]uint8, len(img.Pix))

	for y := range h {
		var sumR, sumG, sumB, sumA uint32

		// Initialize the window with clamped edge extension.
		for i := -radius; i <= radius; i++ {
			x := clampInt(i, 0, w-1)
			off := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)
			sumR += uint32(img.Pix[off+0])
			sumG += uint32(img.Pix[off+1])
			sumB += uint32(img.Pix[off+2])
			sumA += uint32(img.Pix[off+3])
		}

		for x := range w {
			off := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)
			tmp[off+0] = uint8(sumR / uint32(div))
			tmp[off+1] = uint8(sumG / uint32(div))
			tmp[off+2] = uint8(sumB / uint32(div))
			tmp[off+3] = uint8(sumA / uint32(div))

			// Slide window: add right, remove left.
			addX := clampInt(x+radius+1, 0, w-1)
			remX := clampInt(x-radius, 0, w-1)
			addOff := img.PixOffset(bounds.Min.X+addX, bounds.Min.Y+y)
			remOff := img.PixOffset(bounds.Min.X+remX, bounds.Min.Y+y)

			sumR += uint32(img.Pix[addOff+0]) - uint32(img.Pix[remOff+0])
			sumG += uint32(img.Pix[addOff+1]) - uint32(img.Pix[remOff+1])
			sumB += uint32(img.Pix[addOff+2]) - uint32(img.Pix[remOff+2])
			sumA += uint32(img.Pix[addOff+3]) - uint32(img.Pix[remOff+3])
		}
	}

	copy(img.Pix, tmp)
}

func boxBlurVertical(img *image.RGBA, radius int) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	div := 2*radius + 1

	tmp := make([]uint8, len(img.Pix))

	for x := range w {
		var sumR, sumG, sumB, sumA uint32

		for i := -radius; i <= radius; i++ {
			y := clampInt(i, 0, h-1)
			off := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)
			sumR += uint32(img.Pix[off+0])
			sumG += uint32(img.Pix[off+1])
			sumB += uint32(img.Pix[off+2])
			sumA += uint32(img.Pix[off+3])
		}

		for y := range h {
			off := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)
			tmp[off+0] = uint8(sumR / uint32(div))
			tmp[off+1] = uint8(sumG / uint32(div))
			tmp[off+2] = uint8(sumB / uint32(div))
			tmp[off+3] = uint8(sumA / uint32(div))

			addY := clampInt(y+radius+1, 0, h-1)
			remY := clampInt(y-radius, 0, h-1)
			addOff := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+addY)
			remOff := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+remY)

			sumR += uint32(img.Pix[addOff+0]) - uint32(img.Pix[remOff+0])
			sumG += uint32(img.Pix[addOff+1]) - uint32(img.Pix[remOff+1])
			sumB += uint32(img.Pix[addOff+2]) - uint32(img.Pix[remOff+2])
			sumA += uint32(img.Pix[addOff+3]) - uint32(img.Pix[remOff+3])
		}
	}

	copy(img.Pix, tmp)
}
