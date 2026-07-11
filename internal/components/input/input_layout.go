package input

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (i InputWidget) layoutFrame(ctx *frame.Context, gtx layout.Context, state *inputState, style inputStyle, enabled bool, child layout.Widget) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Input
	frameConstraints := gtx.Constraints
	if i.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	height := min(max(gtx.Dp(tokens.Height), frameConstraints.Min.Y), frameConstraints.Max.Y)
	frameConstraints.Min.Y = height
	frameConstraints.Max.Y = height

	macro := op.Record(gtx.Ops)
	childGtx := gtx
	left := gtx.Dp(tokens.PaddingX)
	right := gtx.Dp(tokens.PaddingX)
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
	radius := min(max(gtx.Dp(tokens.Radius), 1), min(size.X, size.Y)/2)
	ringWidth := state.RingWidth(gtx, style.RingWidth)

	opacity := paint.PushOpacity(gtx.Ops, style.Opacity)
	drawInputFrame(gtx, activeTheme, rect, radius, style, ringWidth)
	offset := image.Pt(left, max((size.Y-childDims.Size.Y)/2, 0))
	stack := op.Offset(offset).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()
	opacity.Pop()
	state.AddPointer(gtx, size, !enabled)
	return layout.Dimensions{Size: size}
}
