package imageview

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
)

func (i Widget) layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	i.applyConstraints(&gtx)
	macro := op.Record(gtx.Ops)
	dims := widget.Image{
		Src:      i.source,
		Fit:      i.gioFit(),
		Position: i.position.Direction(),
	}.Layout(gtx)
	call := macro.Stop()
	if dims.Size.X <= 0 || dims.Size.Y <= 0 {
		return dims
	}
	rect := image.Rectangle{Max: dims.Size}
	radius := min(gtx.Dp(i.radius), min(dims.Size.X, dims.Size.Y)/2)
	root := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	if i.source.Size() != (image.Point{}) && i.alt != "" {
		semantic.LabelOp(i.alt).Add(gtx.Ops)
	}
	opacity := float32(1)
	if i.hasOpacity {
		opacity = i.opacity
	}
	opacityStack := paint.PushOpacity(gtx.Ops, opacity)
	call.Add(gtx.Ops)
	opacityStack.Pop()
	root.Pop()
	return dims
}

func (i Widget) applyConstraints(gtx *layout.Context) {
	if i.hasWidth {
		width := min(max(gtx.Dp(i.width), gtx.Constraints.Min.X), gtx.Constraints.Max.X)
		gtx.Constraints.Min.X = width
		gtx.Constraints.Max.X = width
	}
	if i.hasHeight {
		height := min(max(gtx.Dp(i.height), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
		gtx.Constraints.Min.Y = height
		gtx.Constraints.Max.Y = height
	}
}

func (i Widget) gioFit() widget.Fit {
	switch i.fit {
	case FitContain:
		return widget.Contain
	case FitCover:
		return widget.Cover
	case FitFill:
		return widget.Fill
	case FitUnscaled:
		return widget.Unscaled
	default:
		return widget.ScaleDown
	}
}
