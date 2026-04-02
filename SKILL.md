---
name: gocanvas
description: GoCanvas — 2D vector drawing library for Go with an HTML5 Canvas-like API. Covers canvas creation, shapes, paths, transforms, gradients, text rendering, image drawing, annotations, compositing, and Lua scripting bindings.
license: MIT
compatibility: claude-code, opencode
metadata:
  language: go
  domain: 2d-graphics
---

# GoCanvas Skill

Use this when helping someone who imported `github.com/iceisfun/gocanvas` and wants to draw 2D graphics, or who is writing Lua scripts that use the `canvas` global.

## SKILLS

Copy-paste block for an AI assistant:

```text
SKILLS:
- GoCanvas is a pure-Go 2D vector drawing library inspired by the HTML5 Canvas API. No CGo, no platform dependencies.
- Module: github.com/iceisfun/gocanvas. Lua bindings: github.com/iceisfun/gocanvas/luacanvas.
- Core workflow: gocanvas.New(w, h) -> set styles -> draw shapes/text/images -> c.SavePNG("out.png").
- Canvas is initialized with a white background, black fill/stroke, line width 1, identity transform.
- Colors use image/color.RGBA. Helpers: gocanvas.RGB(r,g,b), gocanvas.RGBA(r,g,b,a), gocanvas.Hex("#RRGGBB").
- Drawing state (transform, colors, line style, font, clip, shadow, gradient, composite op) is saved/restored with Save()/Restore().
- Paths: BeginPath, MoveTo, LineTo, QuadraticCurveTo, BezierCurveTo, Arc, ArcTo, Rect, RoundRect, ClosePath, then Fill() or Stroke().
- Convenience shapes that don't affect the current path: FillRect, StrokeRect, ClearRect, FillRoundRect, StrokeRoundRect.
- Transforms: Translate, Scale, Rotate, Transform(Matrix), SetTransform(Matrix), ResetTransform.
- Matrix is [6]float64 laid out as [a, b, tx, c, d, ty]. Constructors: Identity, TranslateMatrix, ScaleMatrix, RotateMatrix, SkewMatrix.
- StrokeMode: StrokeModeScreen (default, constant pixel width) or StrokeModeWorld (stroke scales with transform).
- Gradients: NewLinearGradient(x0,y0,x1,y1) or NewRadialGradient(cx0,cy0,r0,cx1,cy1,r1), add stops with AddColorStop(pos, color), apply with SetFillGradient/SetStrokeGradient.
- Text: LoadFontFile(path) -> font.NewFace(size) -> c.SetFont(face) -> c.FillText/StrokeText/MeasureText. FitText finds largest size that fits a box. FillTextFit auto-fits and draws.
- TextAlign: TextAlignLeft (default), TextAlignCenter, TextAlignRight. TextBaseline: TextBaselineAlphabetic (default), TextBaselineTop, TextBaselineMiddle, TextBaselineBottom.
- DrawImage(img, sx,sy,sw,sh, dx,dy,dw,dh) draws a source sub-rect into a dest rect, respecting transforms and alpha.
- Shadows: SetShadowColor, SetShadowBlur (3-pass box blur), SetShadowOffset.
- Compositing: SetCompositeOp with CompSourceOver (default), CompDestinationOver, CompSourceIn, CompDestinationIn, CompSourceOut, CompDestinationOut, CompLighter, CompCopy, CompXOR, CompMultiply, CompScreen.
- Clipping: build a path then call Clip(). ResetClip() removes the clip. Clips are intersected (narrowing).
- Dash patterns: SetLineDash([]float64{draw, gap, ...}), SetLineDashOffset(offset). Empty slice = solid.
- LineCap: CapButt (default), CapRound, CapSquare. LineJoin: JoinMiter (default), JoinRound, JoinBevel. SetMiterLimit for miter joins.
- Pixel access: SetPixel(x, y, color), GetPixel(x, y) in screen coordinates.
- Annotations (ML/vision helpers): DrawLabeledBox, DrawAABB, DrawLabel, DrawPolygon with AnnotStyle.
- Anti-aliasing: 8x sub-pixel sampling for all filled/stroked paths.
- Output: c.SavePNG(path), c.WritePNG(writer), c.Image() returns *image.RGBA.
- Lua bindings: luacanvas.New().Register(vm) exposes a "canvas" global. Methods use snake_case and colon syntax (c:fill_rect, c:set_fill_color).
- In Lua, colors are r,g,b[,a] integers 0-255. Gradients are created with canvas.linear_gradient/canvas.radial_gradient. Fonts via canvas.load_font(path).
```

## What You Usually Need To Know

GoCanvas follows the HTML5 Canvas API conventions closely. If you know the browser Canvas API, you already know the mental model: set state, build paths, fill or stroke.

Key differences from HTML5 Canvas:
- Colors are `color.RGBA` structs, not CSS strings.
- `StrokeModeScreen` (default) keeps stroke width constant in pixels regardless of transform. Use `StrokeModeWorld` for HTML5-like behavior where strokes scale.
- Font loading is explicit: load a TTF/OTF file, create a face at a specific size.
- No `fillStyle = "red"` — use `SetFillColor(gocanvas.RGB(255, 0, 0))`.

## Go API Quick Reference

### Canvas Lifecycle

```go
c := gocanvas.New(800, 600)       // white background, black fill/stroke
c.Width()                          // 800
c.Height()                         // 600
c.SavePNG("output.png")           // write to file
c.WritePNG(w)                     // write to io.Writer
c.Image()                         // *image.RGBA
```

### Colors

```go
gocanvas.RGB(255, 0, 0)           // opaque red
gocanvas.RGBA(0, 0, 255, 128)    // semi-transparent blue
gocanvas.Hex("#FF6600")           // hex parsing
gocanvas.Hex("#F60")              // short hex
```

### Style

```go
c.SetFillColor(gocanvas.RGB(255, 0, 0))
c.SetStrokeColor(gocanvas.RGB(0, 0, 255))
c.SetLineWidth(3)
c.SetLineCap(gocanvas.CapRound)      // CapButt, CapRound, CapSquare
c.SetLineJoin(gocanvas.JoinRound)    // JoinMiter, JoinRound, JoinBevel
c.SetMiterLimit(10)
c.SetGlobalAlpha(0.5)
c.SetStrokeMode(gocanvas.StrokeModeWorld)
c.SetLineDash([]float64{10, 5})      // 10px dash, 5px gap
c.SetLineDashOffset(3)
c.SetCompositeOp(gocanvas.CompMultiply)
```

### Shapes (don't affect current path)

```go
c.FillRect(x, y, w, h)
c.StrokeRect(x, y, w, h)
c.ClearRect(x, y, w, h)              // transparent black
c.FillRoundRect(x, y, w, h, radius)
c.StrokeRoundRect(x, y, w, h, radius)
```

### Paths

```go
c.BeginPath()
c.MoveTo(x, y)
c.LineTo(x, y)
c.QuadraticCurveTo(cpx, cpy, x, y)
c.BezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y)
c.Arc(cx, cy, r, startAngle, endAngle)    // radians
c.ArcTo(x1, y1, x2, y2, radius)           // tangent arc
c.Rect(x, y, w, h)
c.RoundRect(x, y, w, h, radius)
c.ClosePath()
c.Fill()
c.Stroke()
```

### Transforms

```go
c.Save()                               // push state
c.Restore()                            // pop state
c.Translate(tx, ty)
c.Scale(sx, sy)
c.Rotate(radians)
c.Transform(matrix)                    // multiply into current
c.SetTransform(matrix)                 // replace current
c.ResetTransform()                     // identity
```

### Matrix

```go
m := gocanvas.Identity()
m = gocanvas.TranslateMatrix(10, 20)
m = gocanvas.ScaleMatrix(2, 2)
m = gocanvas.RotateMatrix(math.Pi / 4)
m = gocanvas.SkewMatrix(0.5, 0)
m = m.Multiply(n)                      // applies n first, then m
x, y := m.TransformPoint(px, py)
inv, ok := m.Invert()
```

### Gradients

```go
// Linear
g := gocanvas.NewLinearGradient(x0, y0, x1, y1)
g.AddColorStop(0.0, gocanvas.RGB(255, 0, 0))
g.AddColorStop(1.0, gocanvas.RGB(0, 0, 255))
c.SetFillGradient(g)

// Radial
g := gocanvas.NewRadialGradient(cx0, cy0, r0, cx1, cy1, r1)
g.AddColorStop(0.0, gocanvas.RGB(255, 255, 255))
g.AddColorStop(1.0, gocanvas.RGB(0, 0, 0))
c.SetStrokeGradient(g)
```

### Text

```go
f, _ := gocanvas.LoadFontFile("font.ttf")
face, _ := f.NewFace(24)              // 24pt (1pt = 1px at 72 DPI)
c.SetFont(face)
c.SetTextAlign(gocanvas.TextAlignCenter)
c.SetTextBaseline(gocanvas.TextBaselineMiddle)
c.FillText("Hello", x, y)
c.StrokeText("Hello", x, y)
m := c.MeasureText("Hello")           // TextMetrics{Width, Height, Ascent, Descent}

// Auto-fit text to a bounding box
face, _ = c.FitText("Hello", maxW, maxH, f, minSize, maxSize)
c.FillTextFit("Hello", x, y, w, h, f) // finds size and draws centered
```

### Image Drawing

```go
// img is any image.Image (loaded via image.Decode, etc.)
c.DrawImage(img, sx, sy, sw, sh, dx, dy, dw, dh)
// sx,sy,sw,sh = source rectangle in img
// dx,dy,dw,dh = destination rectangle on canvas
// Respects current transform, globalAlpha, and shadow
```

### Shadows

```go
c.SetShadowColor(gocanvas.RGBA(0, 0, 0, 128))
c.SetShadowBlur(10)
c.SetShadowOffset(5, 5)
// All subsequent draws cast shadows until cleared:
c.SetShadowColor(color.RGBA{})
c.SetShadowBlur(0)
c.SetShadowOffset(0, 0)
```

### Clipping

```go
c.BeginPath()
c.Arc(200, 200, 100, 0, 2*math.Pi)
c.Clip()          // only pixels inside the circle are drawn to now
// ... draw ...
c.ResetClip()     // remove clipping
```

### Pixel Access

```go
c.SetPixel(x, y, gocanvas.RGB(255, 0, 0))
col := c.GetPixel(x, y)   // color.RGBA
```

### Annotations

```go
style := gocanvas.DefaultAnnotStyle() // green stroke, dark bg, white text
style.Font = f
gocanvas.DrawLabeledBox(c, "Person 95%", x, y, w, h, style)
gocanvas.DrawAABB(c, x, y, w, h, style)
gocanvas.DrawLabel(c, "label", x, y, style)
gocanvas.DrawPolygon(c, points, style) // []gocanvas.Point
```

## Lua API Quick Reference

The Lua bindings mirror the Go API with snake_case naming. Canvas methods use colon syntax.

### Module Functions

```lua
local c = canvas.new(800, 600)
local img = canvas.load_image("photo.png")   -- {width, height, _id}
local font = canvas.load_font("font.ttf")    -- {_id}
local g = canvas.linear_gradient(x0, y0, x1, y1)
local g = canvas.radial_gradient(cx0, cy0, r0, cx1, cy1, r1)
```

### Canvas Methods

```lua
-- Style
c:set_fill_color(r, g, b [, a])          -- 0-255
c:set_stroke_color(r, g, b [, a])
c:set_fill_gradient(gradient)
c:set_stroke_gradient(gradient)
c:set_line_width(w)
c:set_global_alpha(a)                     -- 0.0-1.0
c:set_line_dash({10, 5})                  -- or {} for solid
c:set_line_dash_offset(offset)
c:set_stroke_mode("screen")              -- or "world"
c:set_text_align("left")                 -- "left", "center", "right"
c:set_text_baseline("alphabetic")        -- "alphabetic", "top", "middle", "bottom"
c:set_composite_op("source-over")        -- "source-over", "destination-over", "source-in",
                                          -- "destination-in", "source-out", "destination-out",
                                          -- "lighter", "copy", "xor", "multiply", "screen"

-- Shadows
c:set_shadow({color={0,0,0,128}, blur=10, offset_x=5, offset_y=5})
c:clear_shadow()

-- Shapes
c:fill_rect(x, y, w, h)
c:stroke_rect(x, y, w, h)
c:clear_rect(x, y, w, h)
c:fill_round_rect(x, y, w, h, radius)
c:stroke_round_rect(x, y, w, h, radius)

-- Paths
c:begin_path()
c:move_to(x, y)
c:line_to(x, y)
c:arc(cx, cy, r, start_angle, end_angle)
c:arc_to(x1, y1, x2, y2, radius)
c:rect(x, y, w, h)
c:round_rect(x, y, w, h, radius)
c:close_path()
c:fill()
c:stroke()

-- Transforms
c:save()
c:restore()
c:translate(tx, ty)
c:scale(sx, sy)
c:rotate(radians)
c:set_transform(a, b, tx, c, d, ty)
c:transform(a, b, tx, c, d, ty)
c:reset_transform()

-- Clipping
c:clip()
c:reset_clip()

-- Text
c:set_font(font, size)
c:fill_text("text", x, y)
c:stroke_text("text", x, y)
local m = c:measure_text("text")          -- {width, height, ascent, descent}
local m = c:measure_text("text", {font=f, font_size=24})
local m = c:fit_text("text", w, h, font [, min, max])  -- {width, height, ascent, descent, font_size}
c:fill_text_fit("text", x, y, w, h, font)

-- Image drawing
c:draw_image(img, sx, sy, sw, sh, dx, dy, dw, dh)

-- Pixel access
c:set_pixel(x, y, r, g, b [, a])
local r, g, b, a = c:get_pixel(x, y)

-- Annotations
c:draw_labeled_box("label", x, y, w, h, {
    font = font,
    font_size = 14,
    line_width = 2,
    padding = 4,
    stroke_color = {0, 255, 0},
    fill_color = {0, 0, 0, 180},
    text_color = {255, 255, 255},
})

-- Gradients
g:add_color_stop(position, r, g, b [, a])

-- Output
c:save_png("out.png")
```

## Complete Lua Example

```lua
local c = canvas.new(400, 300)

-- Background gradient
local bg = canvas.linear_gradient(0, 0, 0, 300)
bg:add_color_stop(0.0, 100, 150, 255)
bg:add_color_stop(1.0, 20, 40, 100)
c:set_fill_gradient(bg)
c:fill_rect(0, 0, 400, 300)

-- Rounded rectangle with shadow
c:set_shadow({color={0,0,0,100}, blur=8, offset_x=3, offset_y=3})
c:set_fill_color(255, 255, 255)
c:fill_round_rect(30, 30, 150, 100, 12)
c:clear_shadow()

-- Circle with radial gradient
local orb = canvas.radial_gradient(300, 150, 10, 300, 150, 60)
orb:add_color_stop(0.0, 255, 200, 100)
orb:add_color_stop(1.0, 200, 50, 0, 0)
c:set_fill_gradient(orb)
c:begin_path()
c:arc(300, 150, 60, 0, math.pi * 2)
c:fill()

-- Dashed stroke
c:set_stroke_color(255, 255, 255, 180)
c:set_line_width(2)
c:set_line_dash({8, 4})
c:stroke_rect(20, 20, 360, 260)

c:save_png("out.png")
```

## Complete Go Example

```go
package main

import (
    "math"

    "github.com/iceisfun/gocanvas"
)

func main() {
    c := gocanvas.New(400, 300)

    // Gradient background
    bg := gocanvas.NewLinearGradient(0, 0, 0, 300)
    bg.AddColorStop(0.0, gocanvas.RGB(100, 150, 255))
    bg.AddColorStop(1.0, gocanvas.RGB(20, 40, 100))
    c.SetFillGradient(bg)
    c.FillRect(0, 0, 400, 300)

    // Rounded rectangle with shadow
    c.SetShadowColor(gocanvas.RGBA(0, 0, 0, 100))
    c.SetShadowBlur(8)
    c.SetShadowOffset(3, 3)
    c.SetFillColor(gocanvas.RGB(255, 255, 255))
    c.FillRoundRect(30, 30, 150, 100, 12)
    c.SetShadowColor(gocanvas.RGBA(0, 0, 0, 0))
    c.SetShadowBlur(0)
    c.SetShadowOffset(0, 0)

    // Circle
    c.SetFillColor(gocanvas.RGB(255, 100, 50))
    c.BeginPath()
    c.Arc(300, 150, 60, 0, 2*math.Pi)
    c.Fill()

    // Dashed border
    c.SetStrokeColor(gocanvas.RGBA(255, 255, 255, 180))
    c.SetLineWidth(2)
    c.SetLineDash([]float64{8, 4})
    c.StrokeRect(20, 20, 360, 260)

    c.SavePNG("output.png")
}
```

## Embedding Lua Scripts in Go

```go
package main

import (
    "log"
    "os"

    "github.com/iceisfun/gocanvas/luacanvas"
    "github.com/iceisfun/golua/compiler"
    "github.com/iceisfun/golua/parser"
    "github.com/iceisfun/golua/stdlib"
    "github.com/iceisfun/golua/vm"
)

func main() {
    source, _ := os.ReadFile("script.lua")
    block, _ := parser.Parse("script.lua", string(source))
    proto, _ := compiler.Compile("script.lua", block)

    v := vm.New()
    stdlib.Open(v)

    b := luacanvas.New()
    b.Register(v)

    if _, err := v.Run(proto); err != nil {
        log.Fatal(err)
    }
}
```

## Common Patterns

### Rotated Shape

```go
c.Save()
c.Translate(cx, cy)       // move origin to center of rotation
c.Rotate(angle)           // rotate
c.FillRect(-w/2, -h/2, w, h)  // draw centered at origin
c.Restore()
```

### Circular Clip

```go
c.Save()
c.BeginPath()
c.Arc(cx, cy, r, 0, 2*math.Pi)
c.Clip()
c.DrawImage(img, 0, 0, iw, ih, cx-r, cy-r, r*2, r*2)
c.Restore()  // restores clip state too
```

### ML Bounding Box Overlay

```go
style := gocanvas.DefaultAnnotStyle()
style.Font, _ = gocanvas.LoadFontFile("font.ttf")
// Draw image as background
c.DrawImage(photo, 0, 0, pw, ph, 0, 0, cw, ch)
// Overlay detections
for _, det := range detections {
    style.StrokeColor = det.Color
    gocanvas.DrawLabeledBox(c, det.Label, det.X, det.Y, det.W, det.H, style)
}
```
