// Package luacanvas provides Lua scripting bindings for [gocanvas].
//
// It registers a "canvas" global on a [golua] VM that exposes canvas
// creation, image loading, font loading, drawing operations, transforms,
// and annotation helpers to Lua scripts.
//
// Basic usage:
//
//	v := vm.New()
//	stdlib.Open(v)
//
//	b := luacanvas.New()
//	b.Register(v)
//
//	// run a compiled Lua chunk that uses the canvas.* API
//	v.Run(proto)
package luacanvas

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/iceisfun/gocanvas"
	"github.com/iceisfun/golua/vm"
)

// Bindings manages the state for canvas Lua bindings.
// Each instance maintains its own image and font registries,
// so multiple VMs can be used independently.
type Bindings struct {
	images     map[int64]image.Image
	nextImgID  int64
	fonts      map[int64]*gocanvas.Font
	nextFontID int64
}

// New creates a new Bindings instance.
func New() *Bindings {
	return &Bindings{
		images: make(map[int64]image.Image),
		fonts:  make(map[int64]*gocanvas.Font),
	}
}

// Register registers the "canvas" global module on the given VM.
// After calling Register, Lua scripts can use canvas.new(), canvas.load_image(),
// and canvas.load_font().
func (b *Bindings) Register(v *vm.VM) {
	mod := vm.NewEmptyTable()
	mod.SetString("new", vm.NewNativeFunc(b.canvasNew))
	mod.SetString("load_image", vm.NewNativeFunc(b.canvasLoadImage))
	mod.SetString("load_font", vm.NewNativeFunc(b.canvasLoadFont))
	v.SetGlobal("canvas", vm.NewTable(mod))
}

func (b *Bindings) canvasNew(v *vm.VM) int {
	w := v.Get(1)
	h := v.Get(2)
	if !w.IsNumber() || !h.IsNumber() {
		panic(&vm.LuaError{Value: vm.NewString("canvas.new: expected (width, height)")})
	}
	c := gocanvas.New(int(w.AsInt()), int(h.AsInt()))
	v.Set(0, vm.NewTable(b.canvasToLua(c)))
	return 1
}

func (b *Bindings) canvasLoadImage(v *vm.VM) int {
	path := v.Get(1)
	if !path.IsString() {
		panic(&vm.LuaError{Value: vm.NewString("canvas.load_image: expected string path")})
	}

	f, err := os.Open(path.AsString())
	if err != nil {
		panic(&vm.LuaError{Value: vm.NewString(fmt.Sprintf("canvas.load_image: %s", err))})
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(&vm.LuaError{Value: vm.NewString(fmt.Sprintf("canvas.load_image: %s", err))})
	}

	b.nextImgID++
	b.images[b.nextImgID] = img

	bounds := img.Bounds()
	t := vm.NewEmptyTable()
	t.SetString("width", vm.NewInt(int64(bounds.Dx())))
	t.SetString("height", vm.NewInt(int64(bounds.Dy())))
	t.SetString("_id", vm.NewInt(b.nextImgID))

	v.Set(0, vm.NewTable(t))
	return 1
}

func (b *Bindings) canvasLoadFont(v *vm.VM) int {
	path := v.Get(1)
	if !path.IsString() {
		panic(&vm.LuaError{Value: vm.NewString("canvas.load_font: expected string path")})
	}

	f, err := gocanvas.LoadFontFile(path.AsString())
	if err != nil {
		panic(&vm.LuaError{Value: vm.NewString(fmt.Sprintf("canvas.load_font: %s", err))})
	}

	b.nextFontID++
	b.fonts[b.nextFontID] = f

	t := vm.NewEmptyTable()
	t.SetString("_id", vm.NewInt(b.nextFontID))

	v.Set(0, vm.NewTable(t))
	return 1
}

func (b *Bindings) canvasToLua(c *gocanvas.Canvas) *vm.Table {
	t := vm.NewEmptyTable()

	t.SetString("width", vm.NewInt(int64(c.Width())))
	t.SetString("height", vm.NewInt(int64(c.Height())))

	// Style.
	t.SetString("set_fill_color", vm.NewNativeFunc(func(v *vm.VM) int {
		r, g, b, a := colorArgs(v, 2)
		c.SetFillColor(color.RGBA{R: r, G: g, B: b, A: a})
		return 0
	}))
	t.SetString("set_stroke_color", vm.NewNativeFunc(func(v *vm.VM) int {
		r, g, b, a := colorArgs(v, 2)
		c.SetStrokeColor(color.RGBA{R: r, G: g, B: b, A: a})
		return 0
	}))
	t.SetString("set_line_width", vm.NewNativeFunc(func(v *vm.VM) int {
		c.SetLineWidth(v.Get(2).AsFloat())
		return 0
	}))
	t.SetString("set_global_alpha", vm.NewNativeFunc(func(v *vm.VM) int {
		c.SetGlobalAlpha(v.Get(2).AsFloat())
		return 0
	}))

	// Rectangles.
	t.SetString("fill_rect", vm.NewNativeFunc(func(v *vm.VM) int {
		c.FillRect(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat())
		return 0
	}))
	t.SetString("stroke_rect", vm.NewNativeFunc(func(v *vm.VM) int {
		c.StrokeRect(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat())
		return 0
	}))
	t.SetString("clear_rect", vm.NewNativeFunc(func(v *vm.VM) int {
		c.ClearRect(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat())
		return 0
	}))

	// Path.
	t.SetString("begin_path", vm.NewNativeFunc(func(v *vm.VM) int {
		c.BeginPath()
		return 0
	}))
	t.SetString("move_to", vm.NewNativeFunc(func(v *vm.VM) int {
		c.MoveTo(v.Get(2).AsFloat(), v.Get(3).AsFloat())
		return 0
	}))
	t.SetString("line_to", vm.NewNativeFunc(func(v *vm.VM) int {
		c.LineTo(v.Get(2).AsFloat(), v.Get(3).AsFloat())
		return 0
	}))
	t.SetString("arc", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Arc(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat(), v.Get(6).AsFloat())
		return 0
	}))
	t.SetString("rect", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Rect(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat())
		return 0
	}))
	t.SetString("close_path", vm.NewNativeFunc(func(v *vm.VM) int {
		c.ClosePath()
		return 0
	}))

	// Drawing.
	t.SetString("fill", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Fill()
		return 0
	}))
	t.SetString("stroke", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Stroke()
		return 0
	}))

	// Transform.
	t.SetString("save", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Save()
		return 0
	}))
	t.SetString("restore", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Restore()
		return 0
	}))
	t.SetString("translate", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Translate(v.Get(2).AsFloat(), v.Get(3).AsFloat())
		return 0
	}))
	t.SetString("scale", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Scale(v.Get(2).AsFloat(), v.Get(3).AsFloat())
		return 0
	}))
	t.SetString("rotate", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Rotate(v.Get(2).AsFloat())
		return 0
	}))
	t.SetString("set_transform", vm.NewNativeFunc(func(v *vm.VM) int {
		c.SetTransform(gocanvas.Matrix{
			v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(),
			v.Get(5).AsFloat(), v.Get(6).AsFloat(), v.Get(7).AsFloat(),
		})
		return 0
	}))
	t.SetString("transform", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Transform(gocanvas.Matrix{
			v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(),
			v.Get(5).AsFloat(), v.Get(6).AsFloat(), v.Get(7).AsFloat(),
		})
		return 0
	}))
	t.SetString("reset_transform", vm.NewNativeFunc(func(v *vm.VM) int {
		c.ResetTransform()
		return 0
	}))
	t.SetString("set_stroke_mode", vm.NewNativeFunc(func(v *vm.VM) int {
		mode := v.Get(2)
		if !mode.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("set_stroke_mode: expected \"screen\" or \"world\"")})
		}
		switch mode.AsString() {
		case "screen":
			c.SetStrokeMode(gocanvas.StrokeModeScreen)
		case "world":
			c.SetStrokeMode(gocanvas.StrokeModeWorld)
		default:
			panic(&vm.LuaError{Value: vm.NewString("set_stroke_mode: expected \"screen\" or \"world\"")})
		}
		return 0
	}))

	// Image drawing.
	t.SetString("draw_image", vm.NewNativeFunc(func(v *vm.VM) int {
		imgTbl := v.Get(2)
		if !imgTbl.IsTable() {
			panic(&vm.LuaError{Value: vm.NewString("draw_image: first argument must be an image table")})
		}

		id := imgTbl.AsTable().Get(vm.NewString("_id")).AsInt()
		img := b.images[id]
		if img == nil {
			panic(&vm.LuaError{Value: vm.NewString("draw_image: invalid image reference")})
		}

		c.DrawImage(img,
			v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat(), v.Get(6).AsFloat(),
			v.Get(7).AsFloat(), v.Get(8).AsFloat(), v.Get(9).AsFloat(), v.Get(10).AsFloat(),
		)
		return 0
	}))

	// Text.
	t.SetString("set_font", vm.NewNativeFunc(func(v *vm.VM) int {
		fontTbl := v.Get(2)
		if !fontTbl.IsTable() {
			panic(&vm.LuaError{Value: vm.NewString("set_font: first argument must be a font table")})
		}
		size := v.Get(3)
		if !size.IsNumber() {
			panic(&vm.LuaError{Value: vm.NewString("set_font: second argument must be font size")})
		}
		fontID := fontTbl.AsTable().Get(vm.NewString("_id")).AsInt()
		f := b.fonts[fontID]
		if f == nil {
			panic(&vm.LuaError{Value: vm.NewString("set_font: invalid font reference")})
		}
		face, err := f.NewFace(size.AsFloat())
		if err != nil {
			panic(&vm.LuaError{Value: vm.NewString(fmt.Sprintf("set_font: %s", err))})
		}
		c.SetFont(face)
		return 0
	}))
	t.SetString("fill_text", vm.NewNativeFunc(func(v *vm.VM) int {
		text := v.Get(2)
		if !text.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("fill_text: first argument must be a string")})
		}
		c.FillText(text.AsString(), v.Get(3).AsFloat(), v.Get(4).AsFloat())
		return 0
	}))
	t.SetString("stroke_text", vm.NewNativeFunc(func(v *vm.VM) int {
		text := v.Get(2)
		if !text.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("stroke_text: first argument must be a string")})
		}
		c.StrokeText(text.AsString(), v.Get(3).AsFloat(), v.Get(4).AsFloat())
		return 0
	}))
	t.SetString("measure_text", vm.NewNativeFunc(func(v *vm.VM) int {
		text := v.Get(2)
		if !text.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("measure_text: first argument must be a string")})
		}

		// If an options table is provided with font+font_size, use those
		// temporarily. Otherwise use the current canvas font.
		var restore func()
		if opts := v.Get(3); opts.IsTable() {
			tbl := opts.AsTable()
			fontVal := tbl.Get(vm.NewString("font"))
			sizeVal := tbl.Get(vm.NewString("font_size"))
			if fontVal.IsTable() && sizeVal.IsNumber() {
				fontID := fontVal.AsTable().Get(vm.NewString("_id")).AsInt()
				if f := b.fonts[fontID]; f != nil {
					face, err := f.NewFace(sizeVal.AsFloat())
					if err == nil {
						c.Save()
						c.SetFont(face)
						restore = func() { c.Restore() }
					}
				}
			}
		}

		m := c.MeasureText(text.AsString())

		if restore != nil {
			restore()
		}

		result := vm.NewEmptyTable()
		result.SetString("width", vm.NewFloat(m.Width))
		result.SetString("height", vm.NewFloat(m.Height))
		result.SetString("ascent", vm.NewFloat(m.Ascent))
		result.SetString("descent", vm.NewFloat(m.Descent))
		v.Set(0, vm.NewTable(result))
		return 1
	}))

	// Annotations.
	t.SetString("draw_labeled_box", vm.NewNativeFunc(func(v *vm.VM) int {
		label := v.Get(2)
		if !label.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("draw_labeled_box: first argument must be a string")})
		}

		x := v.Get(3).AsFloat()
		y := v.Get(4).AsFloat()
		w := v.Get(5).AsFloat()
		h := v.Get(6).AsFloat()

		style := gocanvas.DefaultAnnotStyle()

		if styleTbl := v.Get(7); styleTbl.IsTable() {
			tbl := styleTbl.AsTable()
			if fontVal := tbl.Get(vm.NewString("font")); fontVal.IsTable() {
				fontID := fontVal.AsTable().Get(vm.NewString("_id")).AsInt()
				if f := b.fonts[fontID]; f != nil {
					style.Font = f
				}
			}
			if v := tbl.Get(vm.NewString("font_size")); v.IsNumber() {
				style.FontSize = v.AsFloat()
			}
			if v := tbl.Get(vm.NewString("line_width")); v.IsNumber() {
				style.LineWidth = v.AsFloat()
			}
			if v := tbl.Get(vm.NewString("padding")); v.IsNumber() {
				style.Padding = v.AsFloat()
			}
			if v := tbl.Get(vm.NewString("stroke_color")); v.IsTable() {
				style.StrokeColor = colorFromTable(v.AsTable())
			}
			if v := tbl.Get(vm.NewString("fill_color")); v.IsTable() {
				style.FillColor = colorFromTable(v.AsTable())
			}
			if v := tbl.Get(vm.NewString("text_color")); v.IsTable() {
				style.TextColor = colorFromTable(v.AsTable())
			}
		}

		gocanvas.DrawLabeledBox(c, label.AsString(), x, y, w, h, style)
		return 0
	}))

	// Output.
	t.SetString("save_png", vm.NewNativeFunc(func(v *vm.VM) int {
		path := v.Get(2)
		if !path.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("save_png: expected string path")})
		}
		if err := c.SavePNG(path.AsString()); err != nil {
			panic(&vm.LuaError{Value: vm.NewString(fmt.Sprintf("save_png: %s", err))})
		}
		return 0
	}))

	return t
}

func colorArgs(v *vm.VM, start int) (uint8, uint8, uint8, uint8) {
	r := uint8(v.Get(start).AsInt())
	g := uint8(v.Get(start + 1).AsInt())
	b := uint8(v.Get(start + 2).AsInt())
	a := uint8(255)
	if arg := v.Get(start + 3); arg.IsNumber() {
		a = uint8(arg.AsInt())
	}
	return r, g, b, a
}

func colorFromTable(tbl vm.LuaTable) color.RGBA {
	r := uint8(tbl.Get(vm.NewInt(1)).AsInt())
	g := uint8(tbl.Get(vm.NewInt(2)).AsInt())
	b := uint8(tbl.Get(vm.NewInt(3)).AsInt())
	a := uint8(255)
	if v := tbl.Get(vm.NewInt(4)); v.IsNumber() {
		a = uint8(v.AsInt())
	}
	return color.RGBA{R: r, G: g, B: b, A: a}
}
