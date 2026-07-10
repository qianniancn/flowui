package surface

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
)

func (s SurfaceWidget) layout(ctx *frame.Context, gtx layout.Context, style surfaceStyle) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	var dims layout.Dimensions
	func() {
		background := ctx.BackgroundColor()
		if style.background.A != 0 {
			background = style.background
		}
		restore := frame.PushColors(ctx, style.foreground, background)
		defer restore()
		if s.child != nil {
			dims = s.child.Layout(ctx, gtx)
		}
	}()
	content := macro.Stop()

	dims.Size = gtx.Constraints.Constrain(dims.Size)
	rect := image.Rectangle{Max: dims.Size}
	radius := min(max(gtx.Dp(s.radius), 0), min(dims.Size.X, dims.Size.Y)/2)
	if s.shadow && !rect.Empty() {
		shapeRadius := s.radius
		render.DrawShadow(gtx, rect, render.RoundedShadowCorners(shapeRadius, shapeRadius, shapeRadius, shapeRadius), render.SurfaceShadow(frame.ActiveTheme(ctx).Palette.SurfaceShadow))
	}
	if style.background.A != 0 && !rect.Empty() {
		paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	content.Add(gtx.Ops)
	return dims
}
