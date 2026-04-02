package gocanvas

import (
	"image/color"
	"math"
	"math/rand"
	"testing"
)

func BenchmarkFillRect(b *testing.B) {
	c := New(1000, 1000)
	c.SetFillColor(color.RGBA{255, 0, 0, 255})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.FillRect(100, 100, 800, 800)
	}
}

func BenchmarkFillRectSmall(b *testing.B) {
	c := New(1000, 1000)
	c.SetFillColor(color.RGBA{255, 0, 0, 255})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.FillRect(400, 400, 10, 10)
	}
}

func BenchmarkStrokeRect(b *testing.B) {
	c := New(1000, 1000)
	c.SetStrokeColor(color.RGBA{0, 0, 255, 255})
	c.SetLineWidth(2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.StrokeRect(100, 100, 800, 800)
	}
}

func BenchmarkCircleFill(b *testing.B) {
	c := New(1000, 1000)
	rnd := rand.New(rand.NewSource(99))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := rnd.Float64() * 1000
		y := rnd.Float64() * 1000
		c.BeginPath()
		c.Arc(x, y, 10, 0, 2*math.Pi)
		c.SetFillColor(color.RGBA{uint8(i % 256), 0, 0, 255})
		c.Fill()
	}
}

func BenchmarkLinearGradientFill(b *testing.B) {
	c := New(1000, 1000)
	g := NewLinearGradient(0, 0, 1000, 1000)
	g.AddColorStop(0, color.RGBA{255, 0, 0, 255})
	g.AddColorStop(0.5, color.RGBA{0, 255, 0, 255})
	g.AddColorStop(1, color.RGBA{0, 0, 255, 255})
	c.SetFillGradient(g)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.FillRect(0, 0, 1000, 1000)
	}
}

func BenchmarkPathComplexFill(b *testing.B) {
	c := New(500, 500)
	c.SetFillColor(color.RGBA{255, 0, 0, 255})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.BeginPath()
		c.MoveTo(250, 10)
		for j := 0; j < 20; j++ {
			angle := float64(j) * 2 * math.Pi / 20
			r := 200.0
			if j%2 == 1 {
				r = 80
			}
			c.LineTo(250+r*math.Cos(angle), 250+r*math.Sin(angle))
		}
		c.ClosePath()
		c.Fill()
	}
}

func BenchmarkTransformedFillRect(b *testing.B) {
	c := New(1000, 1000)
	c.SetFillColor(color.RGBA{0, 255, 0, 255})
	c.Translate(500, 500)
	c.Rotate(0.3)
	c.Scale(1.5, 1.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.FillRect(-100, -100, 200, 200)
	}
}

func BenchmarkSetPixel(b *testing.B) {
	c := New(1000, 1000)
	col := color.RGBA{255, 128, 0, 255}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetPixel(i%1000, (i/1000)%1000, col)
	}
}
