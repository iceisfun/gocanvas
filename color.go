package gocanvas

import "image/color"

// RGB returns an opaque color from 8-bit red, green, blue components.
func RGB(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// RGBA returns a color from 8-bit red, green, blue, alpha components.
func RGBA(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// Hex parses a hex color string. Supported formats:
//
//	"#RGB", "#RRGGBB", "#RRGGBBAA"
//
// The '#' prefix is optional. Returns black on invalid input.
func Hex(s string) color.RGBA {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}

	switch len(s) {
	case 3:
		r := hexNibble(s[0])
		g := hexNibble(s[1])
		b := hexNibble(s[2])
		if r < 0 || g < 0 || b < 0 {
			return color.RGBA{}
		}
		return color.RGBA{
			R: uint8(r | r<<4),
			G: uint8(g | g<<4),
			B: uint8(b | b<<4),
			A: 255,
		}
	case 6:
		r := hexByte(s[0], s[1])
		g := hexByte(s[2], s[3])
		b := hexByte(s[4], s[5])
		if r < 0 || g < 0 || b < 0 {
			return color.RGBA{}
		}
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	case 8:
		r := hexByte(s[0], s[1])
		g := hexByte(s[2], s[3])
		b := hexByte(s[4], s[5])
		a := hexByte(s[6], s[7])
		if r < 0 || g < 0 || b < 0 || a < 0 {
			return color.RGBA{}
		}
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
	default:
		return color.RGBA{}
	}
}

func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}

func hexByte(hi, lo byte) int {
	h := hexNibble(hi)
	l := hexNibble(lo)
	if h < 0 || l < 0 {
		return -1
	}
	return h<<4 | l
}
