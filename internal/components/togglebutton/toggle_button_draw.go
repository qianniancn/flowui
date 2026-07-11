package togglebutton

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func drawToggleButtonSurface(gtx layout.Context, size image.Point, style toggleButtonStyle) {
	if size.X <= 0 || size.Y <= 0 || style.background.A == 0 {
		return
	}
	radius := toggleButtonRadius(gtx, size, style)
	paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
}

func drawToggleButtonFocus(gtx layout.Context, size image.Point, style toggleButtonStyle) {
	if size.X <= 0 || size.Y <= 0 || style.focus <= 0 {
		return
	}
	width := max(gtx.Dp(style.focusWidth), 1)
	offset := max(gtx.Dp(style.focusOffset), 0)
	expand := offset + max(width/2, 1)
	rect := image.Rectangle{Max: size}.Inset(-expand)
	radius := min(toggleButtonRadius(gtx, size, style)+expand, min(rect.Dx(), rect.Dy())/2)
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, max(radius, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func toggleButtonRadius(gtx layout.Context, size image.Point, style toggleButtonStyle) int {
	return min(max(gtx.Dp(style.radius), 0), min(size.X, size.Y)/2)
}
