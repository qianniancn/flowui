package checkboxstyle

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
)

// checkboxIndicatorRenderer owns the default checkmark and indeterminate-bar
// geometry. The widget supplies only resolved foreground and state progress.
func checkboxIndicatorRenderer(gtx layout.Context, request theme.CheckboxIndicatorRenderRequest) layout.Dimensions {
	size := gtx.Constraints.Min
	selection := min(max(request.Selection, 0), 1)
	if selection <= 0 || size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}

	color := request.Foreground
	color.A = byte(float32(color.A)*selection + 0.5)
	if color.A == 0 {
		return layout.Dimensions{Size: size}
	}
	if request.Indeterminate {
		length := max(int(float32(size.X)*0.58*selection+0.5), 1)
		thickness := max(gtx.Dp(request.IndeterminateStroke), 1)
		center := image.Pt(size.X/2, size.Y/2)
		line := image.Rect(center.X-length/2, center.Y-thickness/2,
			center.X+(length+1)/2, center.Y+(thickness+1)/2)
		paint.FillShape(gtx.Ops, color, clip.UniformRRect(line, max(thickness/2, 1)).Op(gtx.Ops))
		return layout.Dimensions{Size: size}
	}

	width := float32(size.X)
	height := float32(size.Y)
	points := [3]f32.Point{
		f32.Pt(width*0.27, height*0.52),
		f32.Pt(width*0.43, height*0.68),
		f32.Pt(width*0.75, height*0.31),
	}
	path := draw.CheckPath(gtx.Ops, points, selection)
	stroke := clip.Stroke{Path: path, Width: float32(max(gtx.Dp(request.CheckStroke), 1))}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, color)
	stroke.Pop()
	return layout.Dimensions{Size: size}
}

var _ theme.CheckboxIndicatorRenderer = checkboxIndicatorRenderer
