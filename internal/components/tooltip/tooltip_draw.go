package tooltip

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawTooltipSurface(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int, style tooltipStyle) {
	radii := theme.Components.Tooltip.Radius
	render.DrawShadow(gtx, rect, render.RoundedShadowCorners(radii, radii, radii, radii), render.ThemeShadow(theme.Shadows.Overlay, theme.Palette.OverlayShadowColor(), 1))
	roundRect := clip.UniformRRect(rect, radius)
	paint.FillShape(gtx.Ops, style.surface, roundRect.Op(gtx.Ops))
	borderWidth := gtx.Dp(theme.Components.Tooltip.BorderWidth)
	if borderWidth > 0 && style.border.A > 0 {
		stroke := clip.Stroke{Path: roundRect.Path(gtx.Ops), Width: float32(borderWidth)}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, style.border)
		stroke.Pop()
	}
}
