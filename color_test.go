package gocanvas

import (
	"image/color"
	"testing"
)

func TestRGB(t *testing.T) {
	c := RGB(10, 20, 30)
	want := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	if c != want {
		t.Errorf("RGB(10,20,30) = %v, want %v", c, want)
	}
}

func TestRGBA(t *testing.T) {
	c := RGBA(10, 20, 30, 128)
	want := color.RGBA{R: 10, G: 20, B: 30, A: 128}
	if c != want {
		t.Errorf("RGBA(10,20,30,128) = %v, want %v", c, want)
	}
}

func TestHex(t *testing.T) {
	tests := []struct {
		input string
		want  color.RGBA
	}{
		{"#FF0000", color.RGBA{255, 0, 0, 255}},
		{"#ff0000", color.RGBA{255, 0, 0, 255}},
		{"FF0000", color.RGBA{255, 0, 0, 255}},
		{"#F00", color.RGBA{255, 0, 0, 255}},
		{"#00FF0080", color.RGBA{0, 255, 0, 128}},
		{"#000", color.RGBA{0, 0, 0, 255}},
		{"#FFF", color.RGBA{255, 255, 255, 255}},
		{"", color.RGBA{}},
		{"#GG0000", color.RGBA{}},
		{"#12345", color.RGBA{}},
	}

	for _, tt := range tests {
		got := Hex(tt.input)
		if got != tt.want {
			t.Errorf("Hex(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
