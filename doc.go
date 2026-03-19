// Package gocanvas provides a 2D vector drawing API inspired by the
// HTML5 Canvas element. It renders to an in-memory RGBA image with
// zero external dependencies.
//
// Basic usage:
//
//	c := gocanvas.New(800, 600)
//	c.SetFillColor(gocanvas.RGB(255, 0, 0))
//	c.FillRect(10, 10, 100, 50)
//	c.SavePNG("output.png")
package gocanvas
