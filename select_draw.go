package flowui

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func drawSelectTrigger(gtx layout.Context, rect image.Rectangle, radius int, style selectStyle) {
	drawInputFrame(gtx, rect, radius, style.field)
}

func drawSelectIndicator(gtx layout.Context, theme *Theme, size image.Point, progress float32, col color.NRGBA) {
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	iconSize := float32(gtx.Dp(theme.Components.Select.IndicatorSize))
	if iconSize <= 0 {
		return
	}
	half := iconSize * 0.31
	stack := op.Affine(f32.AffineId().Rotate(center, progress*float32(math.Pi))).Push(gtx.Ops)
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X-half, center.Y-half*0.35))
	path.LineTo(f32.Pt(center.X, center.Y+half*0.55))
	path.LineTo(f32.Pt(center.X+half, center.Y-half*0.35))
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: max(dpFloat(gtx, theme.Components.Select.IndicatorStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
	stack.Pop()
}

func drawSelectPanel(gtx layout.Context, theme *Theme, rect image.Rectangle, radius int) {
	drawPopupSurface(gtx, theme, rect, radius)
}
