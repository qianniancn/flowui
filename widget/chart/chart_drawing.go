package chart

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// StrokeLine draws a line segment when its width and color are visible.
func StrokeLine(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA) {
	if width <= 0 || col.A == 0 {
		return
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

// PointRect returns a square marker centered at center.
func PointRect(center f32.Point, diameter int) image.Rectangle {
	half := float32(diameter) / 2
	return image.Rect(
		int(math.Round(float64(center.X-half))),
		int(math.Round(float64(center.Y-half))),
		int(math.Round(float64(center.X+half))),
		int(math.Round(float64(center.Y+half))),
	)
}

// ClampLabelPosition keeps a label rectangle inside plot.
func ClampLabelPosition(position, size image.Point, plot image.Rectangle) image.Point {
	position.X = min(max(position.X, plot.Min.X), max(plot.Max.X-size.X, plot.Min.X))
	position.Y = min(max(position.Y, plot.Min.Y), max(plot.Max.Y-size.Y, plot.Min.Y))
	return position
}
