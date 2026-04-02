// Package gocanvas provides a 2D vector drawing API inspired by the
// HTML5 Canvas element. It renders to an in-memory [image.RGBA] with
// no CGo or platform dependencies.
//
// # Features
//
//   - Filled and stroked shapes: rectangles, rounded rectangles, circles,
//     ellipses, and arbitrary paths ([Canvas.FillRect], [Canvas.StrokeRect],
//     [Canvas.FillRoundRect], [Canvas.StrokeRoundRect], [Path.Circle],
//     [Path.Ellipse])
//   - Bezier curves: quadratic ([Path.QuadraticTo]) and cubic ([Path.CubicTo])
//   - Arc and arc-to primitives ([Path.Arc], [Path.ArcTo])
//   - Affine transforms: [Canvas.Translate], [Canvas.Rotate], [Canvas.Scale],
//     [Canvas.Transform], [Canvas.SetTransform], with convenience helpers
//     [Canvas.RotateAbout], [Canvas.ScaleAbout], [Canvas.ShearAbout], and
//     [Canvas.InvertY]
//   - Dash patterns ([Canvas.SetLineDash]) with configurable offset
//   - Shadow rendering with color, blur, and offset
//     ([Canvas.SetShadowColor], [Canvas.SetShadowBlur], [Canvas.SetShadowOffset])
//   - Linear, radial, and conic gradients ([LinearGradient], [RadialGradient],
//     [ConicGradient]) for both fill and stroke styles
//   - Font loading and text measurement ([Font], [FontFace], [Canvas.MeasureText])
//   - Text alignment ([TextAlign], [TextBaseline]) and auto-fit ([Canvas.FitText],
//     [Canvas.FillTextFit])
//   - Word wrapping ([Canvas.WordWrap], [Canvas.FillTextWrapped],
//     [Canvas.MeasureTextWrapped])
//   - Image drawing with source/destination rectangles ([Canvas.DrawImage])
//   - Pixel-level access ([Canvas.SetPixel], [Canvas.GetPixel])
//   - Compositing operations ([CompositeOp]): source-over, destination-over,
//     source-in, lighter, multiply, screen, XOR, and more
//   - Clip paths with winding or even-odd fill rules ([Canvas.Clip],
//     [FillRule], [FillRuleEvenOdd])
//   - Anti-aliased edge rendering
//   - Annotation helpers for labeled bounding boxes and polygons
//     ([DrawLabeledBox], [DrawAABB], [DrawLabel], [DrawPolygon])
//   - Stroke width in screen space or world space ([StrokeModeScreen],
//     [StrokeModeWorld])
//   - Color utilities: [RGB], [RGBA], [Hex]
//
// # Coordinate System
//
// The origin (0, 0) is at the top-left corner. X increases rightward,
// Y increases downward. Use [Canvas.InvertY] for math-style Y-up
// coordinates. All coordinates pass through the current affine
// transform matrix before rendering. See [Canvas.Translate],
// [Canvas.Rotate], [Canvas.Scale], and [Matrix] for transform
// operations.
//
// # State Management
//
// Drawing state includes the transform matrix, fill and stroke colors,
// line styles, font, shadow settings, gradients, composite operation,
// clip mask, and fill rule. Use [Canvas.Save] and [Canvas.Restore] to
// push and pop state. Saving creates a snapshot of the entire state;
// restoring discards the current state and returns to the saved one.
//
// # Basic Usage
//
//	c := gocanvas.New(800, 600)
//	c.SetFillColor(gocanvas.RGB(255, 0, 0))
//	c.FillRect(10, 10, 100, 50)
//	c.SavePNG("output.png")
//
// # Gradients
//
// Fill or stroke styles can use linear, radial, or conic gradients.
// Gradient coordinates are in world space and transform with the
// canvas matrix.
//
//	g := gocanvas.NewLinearGradient(0, 0, 800, 0)
//	g.AddColorStop(0, gocanvas.RGB(255, 0, 0))
//	g.AddColorStop(1, gocanvas.RGB(0, 0, 255))
//	c.SetFillGradient(g)
//	c.FillRect(0, 0, 800, 600)
//
// # Lua Scripting
//
// Optional Lua scripting bindings are available in the separate
// [github.com/iceisfun/gocanvas/luacanvas] module.
package gocanvas
