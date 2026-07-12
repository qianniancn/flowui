package input

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (t TextAreaWidget) layoutFrame(ctx *frame.Context, gtx layout.Context, state *inputState, style inputStyle, enabled bool, child layout.Widget) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.TextArea
	frameConstraints := gtx.Constraints
	if t.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	paddingX := gtx.Dp(tokens.PaddingX)
	paddingY := gtx.Dp(tokens.PaddingY)
	contentHeight := gtx.Sp(tokens.LineHeight) * t.resolvedRows()
	height := max(gtx.Dp(tokens.MinHeight), contentHeight+paddingY*2)
	height = min(max(height, frameConstraints.Min.Y), frameConstraints.Max.Y)
	frameConstraints.Min.Y = height
	frameConstraints.Max.Y = height

	macro := op.Record(gtx.Ops)
	childGtx := gtx
	innerWidth := max(frameConstraints.Max.X-paddingX*2, 0)
	minInnerWidth := min(max(frameConstraints.Min.X-paddingX*2, 0), innerWidth)
	innerHeight := max(height-paddingY*2, 0)
	childGtx.Constraints = layout.Constraints{
		Min: image.Pt(minInnerWidth, innerHeight),
		Max: image.Pt(innerWidth, innerHeight),
	}
	childDims := child(childGtx)
	call := macro.Stop()
	size := image.Pt(childDims.Size.X+paddingX*2, height)
	size = frameConstraints.Constrain(size)

	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(tokens.Radius), 1), min(size.X, size.Y)/2)
	ringWidth := state.RingWidth(gtx, style.RingWidth)

	opacity := paint.PushOpacity(gtx.Ops, style.Opacity)
	drawTextAreaFrame(gtx, activeTheme, rect, radius, style, ringWidth)
	offset := op.Offset(image.Pt(paddingX, paddingY)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	state.AddPointer(gtx, size, !enabled)
	return layout.Dimensions{Size: size}
}
