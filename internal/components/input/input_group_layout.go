package input

import (
	"image"
	"image/color"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type recordedInputGroupChild struct {
	call op.CallOp
	size image.Point
}

func (g InputGroupWidget) layoutFrame(ctx *frame.Context, gtx layout.Context, state *inputGroupState, style inputGroupStyle, enabled bool, editor layout.Widget) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.InputGroup
	constraints := gtx.Constraints
	if g.fullWidth {
		constraints.Min.X = constraints.Max.X
	}
	minHeight := min(gtx.Dp(tokens.MinHeight), constraints.Max.Y)
	padding := gtx.Dp(tokens.PaddingX)
	prefixLeft, prefixRight := padding, padding
	if g.hasPrefixPadding {
		prefixLeft = gtx.Dp(g.prefixLeftPadding)
		prefixRight = gtx.Dp(g.prefixRightPadding)
	}
	suffixLeft, suffixRight := padding, padding
	if g.hasSuffixPadding {
		suffixLeft = gtx.Dp(g.suffixLeftPadding)
		suffixRight = gtx.Dp(g.suffixRightPadding)
	}
	divider := max(gtx.Dp(tokens.DividerWidth), 0)

	prefix := recordInputGroupWidget(ctx, gtx, g.prefix, style.Placeholder, style.Background)
	suffix := recordInputGroupWidget(ctx, gtx, g.suffix, style.Placeholder, style.Background)
	prefixWidth := inputGroupSlotWidth(prefix, prefixLeft, prefixRight, divider, g.prefix != nil)
	suffixWidth := inputGroupSlotWidth(suffix, suffixLeft, suffixRight, divider, g.suffix != nil)
	leftPadding := padding
	if g.prefix != nil {
		leftPadding = 0
	}
	rightPadding := padding
	if g.suffix != nil {
		rightPadding = 0
	}

	outerFixed := prefixWidth + suffixWidth + leftPadding + rightPadding
	maxEditorWidth := max(constraints.Max.X-outerFixed, 0)
	minEditorWidth := max(constraints.Min.X-outerFixed, 0)
	minEditorWidth = min(minEditorWidth, maxEditorWidth)
	editor = insetInputGroupEditor(editor, leftPadding, rightPadding)
	editorChild := recordInputGroupLayout(gtx, editor, layout.Constraints{
		Min: image.Pt(minEditorWidth+leftPadding+rightPadding, 0),
		Max: image.Pt(maxEditorWidth+leftPadding+rightPadding, constraints.Max.Y),
	})

	size := image.Pt(
		prefixWidth+editorChild.size.X+suffixWidth,
		max(minHeight, max(prefix.size.Y, max(editorChild.size.Y, suffix.size.Y))),
	)
	size = constraints.Constrain(size)
	radius := min(max(gtx.Dp(tokens.Radius), 1), min(size.X, size.Y)/2)
	ringWidth := state.RingWidth(gtx, style.RingWidth)

	opacity := paint.PushOpacity(gtx.Ops, style.Opacity)
	drawInputGroupFrame(gtx, frame.ActiveTheme(ctx), image.Rectangle{Max: size}, radius, style, ringWidth)
	x := 0
	if g.prefix != nil {
		addRecordedInputGroupChild(gtx, prefix, image.Pt(prefixLeft, max((size.Y-prefix.size.Y)/2, 0)))
		drawInputGroupDivider(gtx, prefixWidth-divider, size.Y, divider, style)
		x += prefixWidth
	}
	addRecordedInputGroupChild(gtx, editorChild, image.Pt(x, max((size.Y-editorChild.size.Y)/2, 0)))
	x += editorChild.size.X
	if g.suffix != nil {
		drawInputGroupDivider(gtx, x, size.Y, divider, style)
		addRecordedInputGroupChild(gtx, suffix, image.Pt(x+divider+suffixLeft, max((size.Y-suffix.size.Y)/2, 0)))
	}
	opacity.Pop()
	addInputGroupPointer(gtx, &state.State, size, !enabled)
	return layout.Dimensions{Size: size}
}

func insetInputGroupEditor(editor layout.Widget, left, right int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: unit.Dp(left), Right: unit.Dp(right)}.Layout(gtx, editor)
	}
}

func recordInputGroupWidget(ctx *frame.Context, gtx layout.Context, widget frame.Widget, foreground, background color.NRGBA) recordedInputGroupChild {
	if widget == nil {
		return recordedInputGroupChild{}
	}
	restore := frame.PushColors(ctx, foreground, background)
	defer restore()
	return recordInputGroupLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Point{}
		return widget.Layout(ctx, gtx)
	}, layout.Constraints{Max: gtx.Constraints.Max})
}

func recordInputGroupLayout(gtx layout.Context, widget layout.Widget, constraints layout.Constraints) recordedInputGroupChild {
	macro := op.Record(gtx.Ops)
	childGtx := gtx
	childGtx.Constraints = constraints
	dims := widget(childGtx)
	return recordedInputGroupChild{call: macro.Stop(), size: dims.Size}
}

func inputGroupSlotWidth(child recordedInputGroupChild, left, right, divider int, present bool) int {
	if !present {
		return 0
	}
	return child.size.X + left + right + divider
}

func addRecordedInputGroupChild(gtx layout.Context, child recordedInputGroupChild, offset image.Point) {
	stack := op.Offset(offset).Push(gtx.Ops)
	child.call.Add(gtx.Ops)
	stack.Pop()
}

func addInputGroupPointer(gtx layout.Context, state event.Tag, size image.Point, disabled bool) {
	if disabled {
		return
	}
	area := clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, state)
	pass.Pop()
	area.Pop()
}
