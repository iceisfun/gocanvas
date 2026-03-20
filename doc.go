// Package gocanvas provides a 2D vector drawing API inspired by the
// HTML5 Canvas element. It renders to an in-memory RGBA image with
// no CGo or platform dependencies.
//
// Features include filled and stroked shapes, Bezier curves, affine
// transforms, dash patterns, shadow rendering, font loading and text
// measurement, image drawing with source/destination rectangles, and
// annotation helpers for labeled bounding boxes.
//
// Stroke width can operate in screen space (constant pixel width) or
// world space (scales with the current transform) via [StrokeModeScreen]
// and [StrokeModeWorld].
//
// Basic usage:
//
//	c := gocanvas.New(800, 600)
//	c.SetFillColor(gocanvas.RGB(255, 0, 0))
//	c.FillRect(10, 10, 100, 50)
//	c.SavePNG("output.png")
//
// Optional Lua scripting bindings are available in the separate
// [github.com/iceisfun/gocanvas/luacanvas] module.
package gocanvas
