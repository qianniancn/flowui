package tooltip

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
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

func drawTooltipArrow(gtx layout.Context, placement overlay.PopoverPlacement, panel image.Point, anchor float32, size, borderWidth int, style tooltipStyle) {
	if size <= 0 {
		return
	}
	span := float32(size)
	depth := span * 2 / 3
	half := span / 2
	centerX := anchor
	centerY := anchor
	inset := min(float32(max(borderWidth, 0)), depth)
	var from, control0, control1, to f32.Point
	switch placement.Placement().Side {
	case overlay.SideTop:
		y := float32(panel.Y) - inset
		from = f32.Pt(centerX-half, y)
		control0 = f32.Pt(centerX-half+span*0.457069, y+depth)
		control1 = f32.Pt(centerX+span*0.041667, y+depth)
		to = f32.Pt(centerX+half, y)
	case overlay.SideBottom:
		y := inset
		from = f32.Pt(centerX-half, y)
		control0 = f32.Pt(centerX-half+span*0.457069, y-depth)
		control1 = f32.Pt(centerX+span*0.041667, y-depth)
		to = f32.Pt(centerX+half, y)
	case overlay.SideLeft:
		x := float32(panel.X) - inset
		from = f32.Pt(x, centerY-half)
		control0 = f32.Pt(x+depth, centerY-half+span*0.457069)
		control1 = f32.Pt(x+depth, centerY+span*0.041667)
		to = f32.Pt(x, centerY+half)
	case overlay.SideRight:
		x := inset
		from = f32.Pt(x, centerY-half)
		control0 = f32.Pt(x-depth, centerY-half+span*0.457069)
		control1 = f32.Pt(x-depth, centerY+span*0.041667)
		to = f32.Pt(x, centerY+half)
	}

	var fillPath clip.Path
	fillPath.Begin(gtx.Ops)
	fillPath.MoveTo(from)
	fillPath.CubeTo(control0, control1, to)
	fillPath.Close()
	paint.FillShape(gtx.Ops, style.surface, clip.Outline{Path: fillPath.End()}.Op())
	if borderWidth <= 0 || style.border.A == 0 {
		return
	}
	var curve clip.Path
	curve.Begin(gtx.Ops)
	curve.MoveTo(from)
	curve.CubeTo(control0, control1, to)
	stroke := clip.Stroke{Path: curve.End(), Width: float32(borderWidth)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, style.border)
	stroke.Pop()
}
