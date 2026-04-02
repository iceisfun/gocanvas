# Examples

Lua scripts demonstrating gocanvas features. Each script uses the gocanvas Lua bindings to produce a PNG image.

## Running

From the project root:

```
go run github.com/iceisfun/gocanvas/luacanvas/cmd/gocanvas examples/basic.lua
```

The output is written to `out.png` in the working directory.

## Example Index

| Name | Description | Key APIs |
|------|-------------|----------|
| [basic.lua](basic.lua) | Filled and stroked rectangles, a circle, a triangle, and semi-transparent overlay | `canvas.new`, `set_fill_color`, `fill_rect`, `stroke_rect`, `arc`, `begin_path`, `move_to`, `line_to`, `close_path`, `fill` |
| [wheel.lua](wheel.lua) | Color wheel made of pie-slice arcs with a white ring outline | `arc`, `move_to`, `close_path`, `fill`, `set_stroke_color`, `stroke` |
| [concave.lua](concave.lua) | Concave polygons (star, arrow, L-shape) drawn over a loaded image | `canvas.load_image`, `draw_image`, `begin_path`, `move_to`, `line_to`, `close_path`, `fill`, `stroke` |
| [dashed.lua](dashed.lua) | Dashed and dot-dash stroke patterns on polygons, plus glowing dashed lines using shadows | `set_line_dash`, `set_shadow`, `clear_shadow`, `arc`, `stroke` |
| [drawimage.lua](drawimage.lua) | Loading an image and drawing it scaled, cropped, tiled as thumbnails, and with reduced alpha | `canvas.load_image`, `draw_image`, `set_global_alpha` |
| [transform.lua](transform.lua) | Coordinate transforms: scale, rotate, translate, and a fan of rotated image copies | `save`, `restore`, `translate`, `scale`, `rotate`, `draw_image`, `set_global_alpha` |
| [gradient.lua](gradient.lua) | Linear and radial gradients used as fill and stroke styles, including a rotated gradient rect | `canvas.linear_gradient`, `canvas.radial_gradient`, `add_color_stop`, `set_fill_gradient`, `set_stroke_gradient` |
| [roundrect.lua](roundrect.lua) | UI-style dashboard with panels, cards, gauge bars, and buttons using rounded rectangles | `fill_round_rect`, `stroke_round_rect` |
| [arcto.lua](arcto.lua) | Rounded corners via `arc_to`: rounded rectangles, triangles, hexagons, and a sharp-vs-rounded star comparison | `arc_to`, `begin_path`, `move_to`, `close_path`, `fill`, `stroke` |
| [textalign.lua](textalign.lua) | Grid showing all combinations of text horizontal alignment and baseline positioning | `canvas.load_font`, `set_font`, `set_text_align`, `set_text_baseline`, `fill_text` |
| [bb.lua](bb.lua) | Object-detection-style labeled bounding boxes drawn over a source image | `canvas.load_image`, `canvas.load_font`, `draw_image`, `draw_labeled_box` |
| [mandelbrot.lua](mandelbrot.lua) | Pixel-by-pixel Mandelbrot set rendering with smooth HSV coloring | `canvas.new`, `set_pixel`, `save_png` |
