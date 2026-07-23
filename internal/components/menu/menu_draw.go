package menu

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func drawMenuSeparator(gtx layout.Context, size image.Point, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, col, clip.Rect{Max: size}.Op())
}

func drawMenuDot(gtx layout.Context, size image.Point, dotSize int, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	dot := min(max(dotSize, 1), min(size.X, size.Y))
	rect := image.Rect((size.X-dot)/2, (size.Y-dot)/2, (size.X+dot)/2, (size.Y+dot)/2)
	paint.FillShape(gtx.Ops, col, clip.Ellipse(rect).Op(gtx.Ops))
}
