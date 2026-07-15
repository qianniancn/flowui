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
	axis      layout.Axis
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
	size := image.Pt(gtx.Constraints.Max.X, thickness)
	if d.axis == layout.Vertical {
		size = image.Pt(thickness, gtx.Constraints.Max.Y)
	}
	size = gtx.Constraints.Constrain(size)
	drawLine(ctx, gtx, size, d.color, d.hasColor)
	return layout.Dimensions{Size: size}
}

type SeparatorWidget = DividerWidget

func Separator() SeparatorWidget {
	return DividerWidget{axis: layout.Vertical}
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
