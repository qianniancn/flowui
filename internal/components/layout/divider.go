package layoutui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type DividerWidget struct {
	thickness unit.Dp
	color     color.NRGBA
	hasColor  bool
}

func Divider() DividerWidget {
	return DividerWidget{}
}

func (d DividerWidget) Thickness(dp int) DividerWidget {
	d.thickness = unit.Dp(dp)
	return d
}

func (d DividerWidget) Color(c color.NRGBA) DividerWidget {
	d.color = c
	d.hasColor = true
	return d
}

func (d DividerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	thickness := lineThickness(gtx, d.thickness)
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, thickness))
	drawLine(ctx, gtx, size, d.color, d.hasColor)
	return layout.Dimensions{Size: size}
}

type SeparatorWidget struct {
	thickness unit.Dp
	color     color.NRGBA
	hasColor  bool
}

func Separator() SeparatorWidget {
	return SeparatorWidget{}
}

func (s SeparatorWidget) Thickness(dp int) SeparatorWidget {
	s.thickness = unit.Dp(dp)
	return s
}

func (s SeparatorWidget) Color(c color.NRGBA) SeparatorWidget {
	s.color = c
	s.hasColor = true
	return s
}

func (s SeparatorWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	thickness := lineThickness(gtx, s.thickness)
	size := gtx.Constraints.Constrain(image.Pt(thickness, gtx.Constraints.Max.Y))
	drawLine(ctx, gtx, size, s.color, s.hasColor)
	return layout.Dimensions{Size: size}
}

func lineThickness(gtx layout.Context, thickness unit.Dp) int {
	if thickness == 0 {
		thickness = 1
	}
	return max(gtx.Dp(thickness), 1)
}

func drawLine(ctx *frame.Context, gtx layout.Context, size image.Point, c color.NRGBA, hasColor bool) {
	if !hasColor {
		c = frame.ActiveTheme(ctx).Palette.Foreground
		c.A = 0x33
	}
	paint.FillShape(gtx.Ops, c, clip.Rect{Max: size}.Op())
}
