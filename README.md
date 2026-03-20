# gocanvas

A 2D drawing library for Go inspired by the HTML5 Canvas API. Renders to
in-memory RGBA images with no CGo or platform dependencies.

```go
import "github.com/iceisfun/gocanvas"
```

## Features

- Filled and stroked rectangles, paths, arcs, circles, and ellipses
- Quadratic and cubic Bezier curves
- Affine transforms (translate, scale, rotate, direct matrix)
- Line caps, joins, miter limits, and dash patterns
- Stroke modes: screen-space (constant width) or world-space (scales with transform)
- Global alpha compositing
- Shadow rendering with box blur
- TrueType/OpenType font loading, text measurement, and auto-fit
- Image drawing with source/destination rectangles (`DrawImage`)
- Annotation helpers for labeled bounding boxes and polygons
- Save/restore state stack

## Quick Start

```go
package main

import "github.com/iceisfun/gocanvas"

func main() {
    c := gocanvas.New(400, 300)

    c.SetFillColor(gocanvas.RGB(220, 50, 50))
    c.FillRect(20, 20, 120, 80)

    c.SetStrokeColor(gocanvas.RGB(50, 50, 220))
    c.SetLineWidth(3)
    c.StrokeRect(170, 20, 120, 80)

    c.SavePNG("output.png")
}
```

## Drawing Images

```go
// Load an image (any image.Image works).
f, _ := os.Open("photo.png")
img, _, _ := image.Decode(f)
f.Close()

// Draw a source sub-rect (sx, sy, sw, sh) into a dest rect (dx, dy, dw, dh).
c.DrawImage(img, 0, 0, 1024, 1024, 10, 10, 200, 200)

// Works with transforms.
c.Save()
c.Translate(300, 200)
c.Rotate(math.Pi / 4)
c.DrawImage(img, 0, 0, 1024, 1024, -100, -100, 200, 200)
c.Restore()
```

## Stroke Modes

By default, line width is in screen pixels and stays constant regardless of
the current transform (`StrokeModeScreen`). Set `StrokeModeWorld` to make
stroke width scale with the transform, matching HTML5 Canvas behavior.

```go
c.SetStrokeMode(gocanvas.StrokeModeWorld)
c.SetLineWidth(2)
c.Scale(3, 3)
c.StrokeRect(0, 0, 50, 50) // renders with a 6px line (2 * 3)
```

## Transforms

```go
c.Translate(100, 50)
c.Scale(2, 2)
c.Rotate(math.Pi / 4)

// Direct matrix access: | a b tx |
//                       | c d ty |
c.SetTransform(gocanvas.Matrix{a, b, tx, c, d, ty})
c.Transform(gocanvas.Matrix{a, b, tx, c, d, ty}) // multiply into current
c.ResetTransform()
```

## Text

```go
font, _ := gocanvas.LoadFontFile("DejaVuSans-Bold.ttf")

// Set font and draw
face, _ := font.NewFace(24)
c.SetFont(face)
c.FillText("hello", 10, 50)

// Measure text metrics
m := c.MeasureText("hello")
fmt.Println(m.Width, m.Height)

// Auto-fit: find the largest size that fits within a box
fitted, _ := c.FitText("hello", 200, 40, font, 1, 100)
c.SetFont(fitted)

// Or fit and draw in one call (vertically centered)
c.FillTextFit("hello", 10, 10, 200, 40, font)
```

## Annotations

```go
style := gocanvas.DefaultAnnotStyle()
font, _ := gocanvas.LoadFontFile("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf")
style.Font = font

gocanvas.DrawLabeledBox(c, "owl 0.97", 100, 50, 300, 400, style)
```

## Lua Scripting

Optional Lua bindings are available as a separate module:

```
go get github.com/iceisfun/gocanvas/luacanvas
```

See the [luacanvas README](luacanvas/README.md) for details.

## Dependencies

The core library depends only on `golang.org/x/image` for font rendering.
