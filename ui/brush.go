package ui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

// Brush is a resolved solid or gradient fill.
type Brush = render.Brush

// LinearGradient creates a linear gradient from its ordered stops.
func LinearGradient(stops ...GradientStop) Gradient {
	return style.LinearGradient(stops...)
}

// ColorStop creates a gradient stop at a normalized offset.
func ColorStop(offset float32, value ColorSource) GradientStop {
	return style.ColorStop(offset, value)
}

// ResolveColor resolves source using the active theme, or the default theme
// when ctx is nil.
func ResolveColor(ctx *Context, source ColorSource) (color.NRGBA, bool) {
	return styleruntime.ResolveColor(ctx, source)
}

// ResolveBrush resolves source using the active theme, or the default theme
// when ctx is nil.
func ResolveBrush(ctx *Context, source PaintSource) (Brush, bool) {
	return styleruntime.ResolveBrush(ctx, source)
}

// DrawBrush fills rect with brush and clips it to a pixel radius.
func DrawBrush(gtx layout.Context, rect image.Rectangle, radius int, brush Brush) {
	render.DrawBrush(gtx, rect, radius, brush)
}

// DrawBrushRRect fills a rounded rectangle with brush.
func DrawBrushRRect(gtx layout.Context, shape clip.RRect, brush Brush) {
	render.DrawBrushRRect(gtx, shape, brush)
}
