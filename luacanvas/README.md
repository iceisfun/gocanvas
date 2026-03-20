# luacanvas

Lua scripting bindings for [gocanvas](https://github.com/iceisfun/gocanvas),
powered by [golua](https://github.com/iceisfun/golua).

```go
import "github.com/iceisfun/gocanvas/luacanvas"
```

This is a separate Go module so the core canvas library stays free of the
golua dependency. Install it independently:

```
go get github.com/iceisfun/gocanvas/luacanvas
```

## Go Usage

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

## CLI

A ready-to-use command is included:

```
go install github.com/iceisfun/gocanvas/luacanvas/cmd/gocanvas@latest
gocanvas script.lua
```

Or from a clone of the repo:

```
go run ./luacanvas/cmd/gocanvas/ examples/basic.lua
```

## Lua API

### Module Functions

| Function | Description |
|----------|-------------|
| `canvas.new(width, height)` | Create a new canvas. Returns a canvas object. |
| `canvas.load_image(path)` | Load a PNG or JPEG. Returns `{width, height, _id}`. |
| `canvas.load_font(path)` | Load a TTF/OTF font. Returns `{_id}`. |

### Canvas Methods

All methods use `:` syntax (e.g. `c:fill_rect(0, 0, 100, 50)`).

#### Style

| Method | Description |
|--------|-------------|
| `c:set_fill_color(r, g, b [, a])` | Set fill color (0-255). Alpha defaults to 255. |
| `c:set_stroke_color(r, g, b [, a])` | Set stroke color. |
| `c:set_line_width(w)` | Set stroke line width. |
| `c:set_global_alpha(a)` | Set global opacity (0.0-1.0). |
| `c:set_stroke_mode(mode)` | `"screen"` (constant width) or `"world"` (scales with transform). |
| `c:set_line_dash(pattern)` | Set dash pattern, e.g. `{15, 8}`. Pass `{}` or `nil` for solid. |
| `c:set_line_dash_offset(offset)` | Set the dash pattern offset. |
| `c:set_shadow(opts)` | Set shadow: `{color={r,g,b,a}, blur=N, offset_x=N, offset_y=N}`. |
| `c:clear_shadow()` | Remove shadow. |

#### Shapes

| Method | Description |
|--------|-------------|
| `c:fill_rect(x, y, w, h)` | Fill a rectangle. |
| `c:stroke_rect(x, y, w, h)` | Stroke a rectangle. |
| `c:clear_rect(x, y, w, h)` | Clear a rectangle to transparent black. |

#### Path

| Method | Description |
|--------|-------------|
| `c:begin_path()` | Start a new path. |
| `c:move_to(x, y)` | Move to a point. |
| `c:line_to(x, y)` | Line to a point. |
| `c:arc(cx, cy, r, start, end)` | Add an arc (angles in radians). |
| `c:rect(x, y, w, h)` | Add a rectangle sub-path. |
| `c:close_path()` | Close the current sub-path. |
| `c:fill()` | Fill the current path. |
| `c:stroke()` | Stroke the current path. |

#### Transform

| Method | Description |
|--------|-------------|
| `c:save()` | Push the drawing state. |
| `c:restore()` | Pop the drawing state. |
| `c:translate(tx, ty)` | Translate. |
| `c:scale(sx, sy)` | Scale. |
| `c:rotate(radians)` | Rotate. |
| `c:set_transform(a, b, tx, c, d, ty)` | Replace the current matrix. |
| `c:transform(a, b, tx, c, d, ty)` | Multiply into the current matrix. |
| `c:reset_transform()` | Reset to identity. |

#### Text

| Method | Description |
|--------|-------------|
| `c:set_font(font, size)` | Set the active font face. |
| `c:fill_text(text, x, y)` | Draw filled text at the baseline origin. |
| `c:stroke_text(text, x, y)` | Draw outlined text. |
| `c:measure_text(text [, opts])` | Returns `{width, height, ascent, descent}`. |
| `c:fit_text(text, w, h, font [, min, max])` | Find largest size that fits. Returns `{width, height, ascent, descent, font_size}`. |
| `c:fill_text_fit(text, x, y, w, h, font)` | Auto-fit and draw text centered in a box. |

`measure_text` accepts an optional table `{font=, font_size=}` to measure
without changing the canvas state:

```lua
local m = c:measure_text("hello", {font = font, font_size = 24})
print(m.width, m.height)
```

`fit_text` binary-searches for the largest font size that fits within
`w x h`, sets the canvas font to the result, and returns metrics:

```lua
local m = c:fit_text("Hello World", 300, 50, font)
print(m.font_size, m.width, m.height)
-- font is now set, so fill_text uses the fitted size
c:fill_text("Hello World", x, y + m.ascent)
```

#### Images

```lua
local img = canvas.load_image("photo.png")
c:draw_image(img, sx, sy, sw, sh, dx, dy, dw, dh)
```

| Parameter | Description |
|-----------|-------------|
| `sx, sy, sw, sh` | Source rectangle within the image. |
| `dx, dy, dw, dh` | Destination rectangle on the canvas. |

Transforms, global alpha, and shadows are applied.

#### Annotations

```lua
local font = canvas.load_font("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf")
c:draw_labeled_box("label", x, y, w, h, {
    font = font,
    font_size = 14,
    line_width = 2,
    padding = 4,
    stroke_color = {0, 255, 0},
    fill_color = {0, 0, 0, 180},
    text_color = {255, 255, 255},
})
```

All style fields are optional and default to green stroke, dark background,
white text.

#### Output

| Method | Description |
|--------|-------------|
| `c:save_png(path)` | Save the canvas as a PNG file. |

## Examples

See the [`examples/`](../examples/) directory:

- **basic.lua** - Shapes, colors, and transparency.
- **drawimage.lua** - Loading images, sub-rects, scaling, alpha.
- **transform.lua** - Rotate, scale, and combined transforms.
- **bb.lua** - ML-style bounding box annotations with labels.
- **concave.lua** - Concave polygon fill and stroke.
- **dashed.lua** - Dash patterns, dot-dash lines, and glowing dashed strokes.
