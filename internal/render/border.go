package render

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// DrawRoundedBorder paints a rounded border inside rect.
func DrawRoundedBorder(gtx layout.Context, rect image.Rectangle, radius, width int, col color.NRGBA) {
	radius = min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)
	DrawRRectBorder(gtx, clip.UniformRRect(rect, radius), width, col)
}

// DrawRoundedInsetStroke paints a uniform rounded stroke inside rect.
// radius is applied to the inset path, and inset controls its position inside rect.
func DrawRoundedInsetStroke(gtx layout.Context, rect image.Rectangle, radius, width, inset int, col color.NRGBA) {
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	inset = max(inset, 0)
	borderRect := rect.Inset(inset)
	if borderRect.Empty() {
		return
	}
	radius = min(max(radius, 0), min(borderRect.Dx(), borderRect.Dy())/2)
	drawInsetStroke(gtx, clip.UniformRRect(borderRect, radius), width, col)
}

// DrawRRectBorder paints a border inside a rounded rectangle.
func DrawRRectBorder(gtx layout.Context, shape clip.RRect, width int, col color.NRGBA) {
	rect := shape.Rect
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	width = min(width, min(rect.Dx(), rect.Dy())/2)
	inset := max((width+1)/2, 1)
	shape.Rect = rect.Inset(inset)
	if shape.Rect.Empty() {
		return
	}
	shape.NW = max(shape.NW-inset, 0)
	shape.NE = max(shape.NE-inset, 0)
	shape.SE = max(shape.SE-inset, 0)
	shape.SW = max(shape.SW-inset, 0)
	drawInsetStroke(gtx, shape, width, col)
}

func drawInsetStroke(gtx layout.Context, shape clip.RRect, width int, col color.NRGBA) {
	stroke := clip.Stroke{Path: shape.Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

// DrawBottomBorder paints a straight border along the bottom edge of rect.
// The border is drawn inside the rectangle, matching DrawRRectBorder.
func DrawBottomBorder(gtx layout.Context, rect image.Rectangle, width int, col color.NRGBA) {
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	width = min(width, rect.Dy())
	paint.FillShape(gtx.Ops, col, clip.Rect{
		Min: image.Point{X: rect.Min.X, Y: rect.Max.Y - width},
		Max: rect.Max,
	}.Op())
}
