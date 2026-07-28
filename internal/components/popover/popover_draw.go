package popover

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawPopoverSurface(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int, style popoverStyle) {
	render.DrawShadow(gtx, rect, render.RoundedShadowCorners(theme.Components.Popover.Radius, theme.Components.Popover.Radius, theme.Components.Popover.Radius, theme.Components.Popover.Radius), render.ThemeShadow(theme.Shadows.Overlay, theme.Palette.OverlayShadowColor(), 1))
	paint.FillShape(gtx.Ops, style.surface, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawPopoverArrowShape(gtx layout.Context, placement overlay.PopoverPlacement, panel image.Point, arrow image.Point, style popoverStyle) {
	if arrow.X <= 0 || arrow.Y <= 0 {
		return
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	switch placement.Placement().Side {
	case overlay.SideTop:
		x := float32(panel.X) / 2
		y := float32(panel.Y)
		path.MoveTo(f32.Pt(x-float32(arrow.X)/2, y))
		path.LineTo(f32.Pt(x+float32(arrow.X)/2, y))
		path.LineTo(f32.Pt(x, y+float32(arrow.Y)))
	case overlay.SideLeft:
		x := float32(panel.X)
		y := float32(panel.Y) / 2
		path.MoveTo(f32.Pt(x, y-float32(arrow.X)/2))
		path.LineTo(f32.Pt(x, y+float32(arrow.X)/2))
		path.LineTo(f32.Pt(x+float32(arrow.Y), y))
	case overlay.SideRight:
		y := float32(panel.Y) / 2
		path.MoveTo(f32.Pt(0, y-float32(arrow.X)/2))
		path.LineTo(f32.Pt(0, y+float32(arrow.X)/2))
		path.LineTo(f32.Pt(-float32(arrow.Y), y))
	default:
		x := float32(panel.X) / 2
		path.MoveTo(f32.Pt(x-float32(arrow.X)/2, 0))
		path.LineTo(f32.Pt(x+float32(arrow.X)/2, 0))
		path.LineTo(f32.Pt(x, -float32(arrow.Y)))
	}
	paint.FillShape(gtx.Ops, style.surface, clip.Outline{Path: path.End()}.Op())
}
