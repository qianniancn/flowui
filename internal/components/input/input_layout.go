package input

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (i InputWidget) layoutFrame(ctx *frame.Context, gtx layout.Context, state *field.State, style field.Style, child layout.Widget) layout.Dimensions {
	frameConstraints := gtx.Constraints
	if i.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	height := min(gtx.Dp(frame.ActiveTheme(ctx).Spacing.ControlHeight), frameConstraints.Max.Y)
	frameConstraints.Min.Y = min(max(frameConstraints.Min.Y, height), frameConstraints.Max.Y)

	macro := op.Record(gtx.Ops)
	childGtx := gtx
	left := gtx.Dp(frame.ActiveTheme(ctx).Components.Input.PaddingX)
	right := gtx.Dp(frame.ActiveTheme(ctx).Components.Input.PaddingX)
	horizontalPadding := left + right
	maxX := max(frameConstraints.Max.X-horizontalPadding, 0)
	minX := min(max(frameConstraints.Min.X-horizontalPadding, 0), maxX)
	childGtx.Constraints = layout.Constraints{
		Min: image.Pt(minX, 0),
		Max: image.Pt(maxX, frameConstraints.Max.Y),
	}
	childDims := child(childGtx)
	call := macro.Stop()
	size := image.Pt(childDims.Size.X+left+right, childDims.Size.Y)
	size = frameConstraints.Constrain(size)

	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(frame.ActiveTheme(ctx).Shape.ControlRadius), 1), min(size.X, size.Y)/2)

	field.DrawFrame(gtx, rect, radius, style)
	offset := image.Pt(left, max((size.Y-childDims.Size.Y)/2, 0))
	stack := op.Offset(offset).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()
	state.AddPointer(gtx, size, i.disabled)
	return layout.Dimensions{Size: size}
}
