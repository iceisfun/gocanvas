# gocanvas

A 2D drawing library for Go inspired by the HTML5 Canvas API. Renders to
in-memory RGBA images with no CGo or platform dependencies.

```go
import "github.com/iceisfun/gocanvas"
```

## Features

- Filled and stroked rectangles, rounded rectangles, paths, arcs, circles, and ellipses
- Quadratic and cubic Bezier curves, arcTo
- Affine transforms (translate, scale, rotate, direct matrix)
- Line caps, joins, miter limits, and dash patterns
- Stroke modes: screen-space (constant width) or world-space (scales with transform)
- Composite operations (source-over, destination-over, lighter, multiply, screen, xor, etc.)
- Anti-aliased edges (8x sub-pixel sampling)
- Clip paths
- Global alpha compositing
- Shadow rendering with box blur
- TrueType/OpenType font loading, text measurement, auto-fit, text alignment
- Image drawing with source/destination rectangles (`DrawImage`)
- Annotation helpers for labeled bounding boxes and polygons
- Linear and radial gradients for fill and stroke
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

## Rounded Rectangles

```go
c.FillRoundRect(10, 10, 200, 100, 15)    // filled with corner radius 15
c.StrokeRoundRect(10, 10, 200, 100, 15)  // stroked outline

// Or as a path for more control
c.BeginPath()
c.RoundRect(10, 10, 200, 100, 15)
c.Fill()
```

## Clip Paths

```go
c.Save()
c.BeginPath()
c.Arc(200, 150, 100, 0, math.Pi*2)
c.Clip()                            // only draw inside the circle
c.FillRect(0, 0, 400, 300)         // clipped to circle
c.Restore()                         // restores previous clip
c.ResetClip()                       // or explicitly remove clip
```

## Composite Operations

```go
c.SetCompositeOp(gocanvas.CompMultiply)      // multiply blend
c.SetCompositeOp(gocanvas.CompScreen)        // screen blend
c.SetCompositeOp(gocanvas.CompLighter)       // additive
c.SetCompositeOp(gocanvas.CompDestinationOver) // draw behind
c.SetCompositeOp(gocanvas.CompSourceOver)    // default
```

## Text

```go
font, _ := gocanvas.LoadFontFile("DejaVuSans-Bold.ttf")

// Set font and draw
face, _ := font.NewFace(24)
c.SetFont(face)
c.FillText("hello", 10, 50)

// Text alignment
c.SetTextAlign(gocanvas.TextAlignCenter)
c.SetTextBaseline(gocanvas.TextBaselineMiddle)
c.FillText("centered", 200, 150)

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

## Gradients

Fill and stroke operations support linear and radial gradients. Gradient
coordinates are in world space and transform with the current matrix.

```go
// Linear gradient from left to right.
lg := gocanvas.NewLinearGradient(0, 0, 400, 0)
lg.AddColorStop(0, gocanvas.RGB(255, 0, 0))
lg.AddColorStop(1, gocanvas.RGB(0, 0, 255))
c.SetFillGradient(lg)
c.FillRect(0, 0, 400, 300)

// Radial gradient (spotlight effect).
rg := gocanvas.NewRadialGradient(200, 150, 10, 200, 150, 150)
rg.AddColorStop(0, gocanvas.RGB(255, 255, 200))
rg.AddColorStop(1, gocanvas.RGBA(0, 0, 0, 0))
c.SetFillGradient(rg)
c.FillRect(0, 0, 400, 300)

// Setting a solid color clears the gradient.
c.SetFillColor(gocanvas.RGB(0, 0, 0))
```

## Lua Scripting

Optional Lua bindings are available as a separate module:

```
go get github.com/iceisfun/gocanvas/luacanvas
```

See the [luacanvas README](luacanvas/README.md) for details.

## Dependencies

The core library depends only on `golang.org/x/image` for font rendering.
