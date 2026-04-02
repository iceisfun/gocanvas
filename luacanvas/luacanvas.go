// Package luacanvas provides Lua scripting bindings for [gocanvas].
//
// It registers a "canvas" global on a [golua] VM that exposes canvas
// creation, image loading, font loading, drawing operations, transforms,
// and annotation helpers to Lua scripts.
//
// Gradients include linear, radial, and conic (canvas.conic_gradient).
// Fill rule can be set via set_fill_rule ("winding" or "evenodd").
// Convenience transforms rotate_about, scale_about, shear_about, and
// invert_y are available alongside the standard transform methods.
// Text support includes word_wrap and fill_text_wrapped for word-wrapped
// text layout and rendering.
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
// Each instance maintains its own image, font, and gradient registries,
// so multiple VMs can be used independently.
type Bindings struct {
	images      map[int64]image.Image
	nextImgID   int64
	fonts       map[int64]*gocanvas.Font
	nextFontID  int64
	gradients   map[int64]gocanvas.Gradient
	nextGradID  int64
}

// New creates a new Bindings instance.
func New() *Bindings {
	return &Bindings{
		images:    make(map[int64]image.Image),
		fonts:     make(map[int64]*gocanvas.Font),
		gradients: make(map[int64]gocanvas.Gradient),
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
	mod.SetString("linear_gradient", vm.NewNativeFunc(b.canvasLinearGradient))
	mod.SetString("radial_gradient", vm.NewNativeFunc(b.canvasRadialGradient))
	mod.SetString("conic_gradient", vm.NewNativeFunc(b.canvasConicGradient))
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
	t.SetString("set_fill_gradient", vm.NewNativeFunc(func(v *vm.VM) int {
		gradTbl := v.Get(2)
		if !gradTbl.IsTable() {
			panic(&vm.LuaError{Value: vm.NewString("set_fill_gradient: argument must be a gradient table")})
		}
		id := gradTbl.AsTable().Get(vm.NewString("_id")).AsInt()
		g := b.gradients[id]
		if g == nil {
			panic(&vm.LuaError{Value: vm.NewString("set_fill_gradient: invalid gradient reference")})
		}
		c.SetFillGradient(g)
		return 0
	}))
	t.SetString("set_stroke_gradient", vm.NewNativeFunc(func(v *vm.VM) int {
		gradTbl := v.Get(2)
		if !gradTbl.IsTable() {
			panic(&vm.LuaError{Value: vm.NewString("set_stroke_gradient: argument must be a gradient table")})
		}
		id := gradTbl.AsTable().Get(vm.NewString("_id")).AsInt()
		g := b.gradients[id]
		if g == nil {
			panic(&vm.LuaError{Value: vm.NewString("set_stroke_gradient: invalid gradient reference")})
		}
		c.SetStrokeGradient(g)
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
	t.SetString("set_line_dash", vm.NewNativeFunc(func(v *vm.VM) int {
		pat := v.Get(2)
		if pat.IsNil() || (pat.IsTable() && pat.AsTable().Len() == 0) {
			c.SetLineDash(nil)
			return 0
		}
		if !pat.IsTable() {
			panic(&vm.LuaError{Value: vm.NewString("set_line_dash: expected table or nil")})
		}
		tbl := pat.AsTable()
		n := tbl.Len()
		dash := make([]float64, n)
		for i := range n {
			dash[i] = tbl.Get(vm.NewInt(int64(i + 1))).AsFloat()
		}
		c.SetLineDash(dash)
		return 0
	}))
	t.SetString("set_line_dash_offset", vm.NewNativeFunc(func(v *vm.VM) int {
		c.SetLineDashOffset(v.Get(2).AsFloat())
		return 0
	}))
	t.SetString("set_shadow", vm.NewNativeFunc(func(v *vm.VM) int {
		opts := v.Get(2)
		if !opts.IsTable() {
			panic(&vm.LuaError{Value: vm.NewString("set_shadow: expected table")})
		}
		tbl := opts.AsTable()
		if col := tbl.Get(vm.NewString("color")); col.IsTable() {
			c.SetShadowColor(colorFromTable(col.AsTable()))
		}
		if blur := tbl.Get(vm.NewString("blur")); blur.IsNumber() {
			c.SetShadowBlur(blur.AsFloat())
		}
		if ox := tbl.Get(vm.NewString("offset_x")); ox.IsNumber() {
			if oy := tbl.Get(vm.NewString("offset_y")); oy.IsNumber() {
				c.SetShadowOffset(ox.AsFloat(), oy.AsFloat())
			}
		}
		return 0
	}))
	t.SetString("clear_shadow", vm.NewNativeFunc(func(v *vm.VM) int {
		c.SetShadowColor(color.RGBA{})
		c.SetShadowBlur(0)
		c.SetShadowOffset(0, 0)
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

	// Rounded rectangles.
	t.SetString("round_rect", vm.NewNativeFunc(func(v *vm.VM) int {
		c.RoundRect(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat(), v.Get(6).AsFloat())
		return 0
	}))
	t.SetString("fill_round_rect", vm.NewNativeFunc(func(v *vm.VM) int {
		c.FillRoundRect(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat(), v.Get(6).AsFloat())
		return 0
	}))
	t.SetString("stroke_round_rect", vm.NewNativeFunc(func(v *vm.VM) int {
		c.StrokeRoundRect(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat(), v.Get(6).AsFloat())
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
	t.SetString("arc_to", vm.NewNativeFunc(func(v *vm.VM) int {
		c.ArcTo(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat(), v.Get(6).AsFloat())
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
	t.SetString("rotate_about", vm.NewNativeFunc(func(v *vm.VM) int {
		c.RotateAbout(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat())
		return 0
	}))
	t.SetString("scale_about", vm.NewNativeFunc(func(v *vm.VM) int {
		c.ScaleAbout(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat())
		return 0
	}))
	t.SetString("shear_about", vm.NewNativeFunc(func(v *vm.VM) int {
		c.ShearAbout(v.Get(2).AsFloat(), v.Get(3).AsFloat(), v.Get(4).AsFloat(), v.Get(5).AsFloat())
		return 0
	}))
	t.SetString("invert_y", vm.NewNativeFunc(func(v *vm.VM) int {
		c.InvertY()
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

	// Text alignment.
	t.SetString("set_text_align", vm.NewNativeFunc(func(v *vm.VM) int {
		arg := v.Get(2)
		if !arg.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("set_text_align: expected \"left\", \"center\", or \"right\"")})
		}
		switch arg.AsString() {
		case "left":
			c.SetTextAlign(gocanvas.TextAlignLeft)
		case "center":
			c.SetTextAlign(gocanvas.TextAlignCenter)
		case "right":
			c.SetTextAlign(gocanvas.TextAlignRight)
		default:
			panic(&vm.LuaError{Value: vm.NewString("set_text_align: expected \"left\", \"center\", or \"right\"")})
		}
		return 0
	}))
	t.SetString("set_text_baseline", vm.NewNativeFunc(func(v *vm.VM) int {
		arg := v.Get(2)
		if !arg.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("set_text_baseline: expected \"alphabetic\", \"top\", \"middle\", or \"bottom\"")})
		}
		switch arg.AsString() {
		case "alphabetic":
			c.SetTextBaseline(gocanvas.TextBaselineAlphabetic)
		case "top":
			c.SetTextBaseline(gocanvas.TextBaselineTop)
		case "middle":
			c.SetTextBaseline(gocanvas.TextBaselineMiddle)
		case "bottom":
			c.SetTextBaseline(gocanvas.TextBaselineBottom)
		default:
			panic(&vm.LuaError{Value: vm.NewString("set_text_baseline: expected \"alphabetic\", \"top\", \"middle\", or \"bottom\"")})
		}
		return 0
	}))

	// Composite operations.
	t.SetString("set_composite_op", vm.NewNativeFunc(func(v *vm.VM) int {
		arg := v.Get(2)
		if !arg.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("set_composite_op: expected string")})
		}
		opMap := map[string]gocanvas.CompositeOp{
			"source-over":      gocanvas.CompSourceOver,
			"destination-over": gocanvas.CompDestinationOver,
			"source-in":        gocanvas.CompSourceIn,
			"destination-in":   gocanvas.CompDestinationIn,
			"source-out":       gocanvas.CompSourceOut,
			"destination-out":  gocanvas.CompDestinationOut,
			"lighter":          gocanvas.CompLighter,
			"copy":             gocanvas.CompCopy,
			"xor":              gocanvas.CompXOR,
			"multiply":         gocanvas.CompMultiply,
			"screen":           gocanvas.CompScreen,
		}
		op, ok := opMap[arg.AsString()]
		if !ok {
			panic(&vm.LuaError{Value: vm.NewString("set_composite_op: unknown operation: " + arg.AsString())})
		}
		c.SetCompositeOp(op)
		return 0
	}))

	// Fill rule.
	t.SetString("set_fill_rule", vm.NewNativeFunc(func(v *vm.VM) int {
		arg := v.Get(2)
		if !arg.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("set_fill_rule: expected \"winding\" or \"evenodd\"")})
		}
		switch arg.AsString() {
		case "winding":
			c.SetFillRule(gocanvas.FillRuleWinding)
		case "evenodd":
			c.SetFillRule(gocanvas.FillRuleEvenOdd)
		default:
			panic(&vm.LuaError{Value: vm.NewString("set_fill_rule: expected \"winding\" or \"evenodd\"")})
		}
		return 0
	}))

	// Clipping.
	t.SetString("clip", vm.NewNativeFunc(func(v *vm.VM) int {
		c.Clip()
		return 0
	}))
	t.SetString("reset_clip", vm.NewNativeFunc(func(v *vm.VM) int {
		c.ResetClip()
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

	// fit_text(text, w, h, font [, min_size, max_size])
	// Returns {width, height, ascent, descent, font_size} for the best fit.
	t.SetString("fit_text", vm.NewNativeFunc(func(v *vm.VM) int {
		text := v.Get(2)
		if !text.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("fit_text: first argument must be a string")})
		}
		w := v.Get(3).AsFloat()
		h := v.Get(4).AsFloat()

		fontTbl := v.Get(5)
		if !fontTbl.IsTable() {
			panic(&vm.LuaError{Value: vm.NewString("fit_text: fourth argument must be a font table")})
		}
		fontID := fontTbl.AsTable().Get(vm.NewString("_id")).AsInt()
		f := b.fonts[fontID]
		if f == nil {
			panic(&vm.LuaError{Value: vm.NewString("fit_text: invalid font reference")})
		}

		minSize := 1.0
		maxSize := h * 2
		if arg := v.Get(6); arg.IsNumber() {
			minSize = arg.AsFloat()
		}
		if arg := v.Get(7); arg.IsNumber() {
			maxSize = arg.AsFloat()
		}

		face, err := c.FitText(text.AsString(), w, h, f, minSize, maxSize)
		if err != nil {
			panic(&vm.LuaError{Value: vm.NewString(fmt.Sprintf("fit_text: %s", err))})
		}

		c.SetFont(face)
		m := c.MeasureText(text.AsString())

		result := vm.NewEmptyTable()
		result.SetString("width", vm.NewFloat(m.Width))
		result.SetString("height", vm.NewFloat(m.Height))
		result.SetString("ascent", vm.NewFloat(m.Ascent))
		result.SetString("descent", vm.NewFloat(m.Descent))
		result.SetString("font_size", vm.NewFloat(face.Size()))
		v.Set(0, vm.NewTable(result))
		return 1
	}))
	// fill_text_fit(text, x, y, w, h, font)
	// Finds the largest font size that fits and draws centered in the box.
	t.SetString("fill_text_fit", vm.NewNativeFunc(func(v *vm.VM) int {
		text := v.Get(2)
		if !text.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("fill_text_fit: first argument must be a string")})
		}

		fontTbl := v.Get(7)
		if !fontTbl.IsTable() {
			panic(&vm.LuaError{Value: vm.NewString("fill_text_fit: sixth argument must be a font table")})
		}
		fontID := fontTbl.AsTable().Get(vm.NewString("_id")).AsInt()
		f := b.fonts[fontID]
		if f == nil {
			panic(&vm.LuaError{Value: vm.NewString("fill_text_fit: invalid font reference")})
		}

		err := c.FillTextFit(text.AsString(),
			v.Get(3).AsFloat(), v.Get(4).AsFloat(),
			v.Get(5).AsFloat(), v.Get(6).AsFloat(), f)
		if err != nil {
			panic(&vm.LuaError{Value: vm.NewString(fmt.Sprintf("fill_text_fit: %s", err))})
		}
		return 0
	}))

	t.SetString("word_wrap", vm.NewNativeFunc(func(v *vm.VM) int {
		text := v.Get(2)
		if !text.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("word_wrap: first argument must be a string")})
		}
		width := v.Get(3).AsFloat()
		lines := c.WordWrap(text.AsString(), width)
		result := vm.NewEmptyTable()
		for i, line := range lines {
			result.Set(vm.NewInt(int64(i+1)), vm.NewString(line))
		}
		v.Set(0, vm.NewTable(result))
		return 1
	}))

	t.SetString("fill_text_wrapped", vm.NewNativeFunc(func(v *vm.VM) int {
		text := v.Get(2)
		if !text.IsString() {
			panic(&vm.LuaError{Value: vm.NewString("fill_text_wrapped: first argument must be a string")})
		}
		x := v.Get(3).AsFloat()
		y := v.Get(4).AsFloat()
		width := v.Get(5).AsFloat()
		lineSpacing := 1.2
		if arg := v.Get(6); arg.IsNumber() {
			lineSpacing = arg.AsFloat()
		}
		c.FillTextWrapped(text.AsString(), x, y, width, lineSpacing)
		return 0
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

	// Pixel access.
	t.SetString("set_pixel", vm.NewNativeFunc(func(v *vm.VM) int {
		x := int(v.Get(2).AsInt())
		y := int(v.Get(3).AsInt())
		r, g, b, a := colorArgs(v, 4)
		c.SetPixel(x, y, color.RGBA{r, g, b, a})
		return 0
	}))

	t.SetString("get_pixel", vm.NewNativeFunc(func(v *vm.VM) int {
		x := int(v.Get(2).AsInt())
		y := int(v.Get(3).AsInt())
		col := c.GetPixel(x, y)
		v.Set(0, vm.NewInt(int64(col.R)))
		v.Set(1, vm.NewInt(int64(col.G)))
		v.Set(2, vm.NewInt(int64(col.B)))
		v.Set(3, vm.NewInt(int64(col.A)))
		return 4
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

func (b *Bindings) canvasLinearGradient(v *vm.VM) int {
	x0 := v.Get(1).AsFloat()
	y0 := v.Get(2).AsFloat()
	x1 := v.Get(3).AsFloat()
	y1 := v.Get(4).AsFloat()
	g := gocanvas.NewLinearGradient(x0, y0, x1, y1)

	b.nextGradID++
	id := b.nextGradID
	b.gradients[id] = g

	t := vm.NewEmptyTable()
	t.SetString("_id", vm.NewInt(id))
	t.SetString("add_color_stop", vm.NewNativeFunc(func(v *vm.VM) int {
		// self=arg1, position=arg2, r=arg3, g=arg4, b=arg5, a=arg6 optional
		pos := v.Get(2).AsFloat()
		r := uint8(v.Get(3).AsInt())
		gv := uint8(v.Get(4).AsInt())
		bv := uint8(v.Get(5).AsInt())
		a := uint8(255)
		if arg := v.Get(6); arg.IsNumber() {
			a = uint8(arg.AsInt())
		}
		g.AddColorStop(pos, color.RGBA{R: r, G: gv, B: bv, A: a})
		return 0
	}))

	v.Set(0, vm.NewTable(t))
	return 1
}

func (b *Bindings) canvasRadialGradient(v *vm.VM) int {
	cx0 := v.Get(1).AsFloat()
	cy0 := v.Get(2).AsFloat()
	r0 := v.Get(3).AsFloat()
	cx1 := v.Get(4).AsFloat()
	cy1 := v.Get(5).AsFloat()
	r1 := v.Get(6).AsFloat()
	g := gocanvas.NewRadialGradient(cx0, cy0, r0, cx1, cy1, r1)

	b.nextGradID++
	id := b.nextGradID
	b.gradients[id] = g

	t := vm.NewEmptyTable()
	t.SetString("_id", vm.NewInt(id))
	t.SetString("add_color_stop", vm.NewNativeFunc(func(v *vm.VM) int {
		pos := v.Get(2).AsFloat()
		r := uint8(v.Get(3).AsInt())
		gv := uint8(v.Get(4).AsInt())
		bv := uint8(v.Get(5).AsInt())
		a := uint8(255)
		if arg := v.Get(6); arg.IsNumber() {
			a = uint8(arg.AsInt())
		}
		g.AddColorStop(pos, color.RGBA{R: r, G: gv, B: bv, A: a})
		return 0
	}))

	v.Set(0, vm.NewTable(t))
	return 1
}

func (b *Bindings) canvasConicGradient(v *vm.VM) int {
	cx := v.Get(1).AsFloat()
	cy := v.Get(2).AsFloat()
	deg := v.Get(3).AsFloat()
	g := gocanvas.NewConicGradient(cx, cy, deg)

	b.nextGradID++
	id := b.nextGradID
	b.gradients[id] = g

	t := vm.NewEmptyTable()
	t.SetString("_id", vm.NewInt(id))
	t.SetString("add_color_stop", vm.NewNativeFunc(func(v *vm.VM) int {
		pos := v.Get(2).AsFloat()
		r := uint8(v.Get(3).AsInt())
		gv := uint8(v.Get(4).AsInt())
		bv := uint8(v.Get(5).AsInt())
		a := uint8(255)
		if arg := v.Get(6); arg.IsNumber() {
			a = uint8(arg.AsInt())
		}
		g.AddColorStop(pos, color.RGBA{R: r, G: gv, B: bv, A: a})
		return 0
	}))

	v.Set(0, vm.NewTable(t))
	return 1
}
