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
	paddingDp := tokens.PaddingX
	verticalPaddingDp := unit.Dp(0)
	if g.multiline {
		verticalPaddingDp = tokens.TextAreaPaddingY
	}
	verticalPadding := gtx.Dp(verticalPaddingDp)
	prefixLeftDp, prefixRightDp := paddingDp, paddingDp
	if g.hasPrefixPadding {
		prefixLeftDp = g.prefixLeftPadding
		prefixRightDp = g.prefixRightPadding
	}
	prefixLeft, prefixRight := gtx.Dp(prefixLeftDp), gtx.Dp(prefixRightDp)
	suffixLeftDp, suffixRightDp := paddingDp, paddingDp
	if g.hasSuffixPadding {
		suffixLeftDp = g.suffixLeftPadding
		suffixRightDp = g.suffixRightPadding
	}
	suffixLeft, suffixRight := gtx.Dp(suffixLeftDp), gtx.Dp(suffixRightDp)
	divider := max(gtx.Dp(tokens.DividerWidth), 0)

	prefix := recordInputGroupWidget(ctx, gtx, g.prefix, style.Placeholder, style.Background)
	suffix := recordInputGroupWidget(ctx, gtx, g.suffix, style.Placeholder, style.Background)
	prefixWidth := inputGroupSlotWidth(prefix, prefixLeft, prefixRight, divider, g.prefix != nil)
	suffixWidth := inputGroupSlotWidth(suffix, suffixLeft, suffixRight, divider, g.suffix != nil)
	leftPaddingDp := paddingDp
	if g.prefix != nil {
		leftPaddingDp = 0
	}
	rightPaddingDp := paddingDp
	if g.suffix != nil {
		rightPaddingDp = 0
	}
	leftPadding, rightPadding := gtx.Dp(leftPaddingDp), gtx.Dp(rightPaddingDp)

	outerFixed := prefixWidth + suffixWidth + leftPadding + rightPadding
	maxEditorWidth := max(constraints.Max.X-outerFixed, 0)
	minEditorWidth := max(constraints.Min.X-outerFixed, 0)
	minEditorWidth = min(minEditorWidth, maxEditorWidth)
	editor = insetInputGroupEditor(editor, leftPaddingDp, rightPaddingDp, verticalPaddingDp)
	minEditorHeight := 0
	maxEditorHeight := constraints.Max.Y
	if g.multiline {
		lineHeight := max(gtx.Sp(tokens.LineHeight), 0)
		contentHeight := inputGroupTextAreaContentHeight(lineHeight, g.textArea.resolvedRows(), constraints.Max.Y)
		minEditorHeight = max(gtx.Dp(tokens.TextAreaMinHeight), contentHeight+verticalPadding*2)
		minEditorHeight = min(minEditorHeight, constraints.Max.Y)
		maxEditorHeight = minEditorHeight
	}
	editorChild := recordInputGroupLayout(gtx, editor, layout.Constraints{
		Min: image.Pt(minEditorWidth+leftPadding+rightPadding, minEditorHeight),
		Max: image.Pt(maxEditorWidth+leftPadding+rightPadding, maxEditorHeight),
	})

	prefixHeight := inputGroupSlotHeight(prefix.size.Y, verticalPadding, g.prefix != nil, g.multiline)
	suffixHeight := inputGroupSlotHeight(suffix.size.Y, verticalPadding, g.suffix != nil, g.multiline)
	size := image.Pt(
		prefixWidth+editorChild.size.X+suffixWidth,
		max(minHeight, max(prefixHeight, max(editorChild.size.Y, suffixHeight))),
	)
	size = constraints.Constrain(size)
	radius := min(max(gtx.Dp(tokens.Radius), 1), min(size.X, size.Y)/2)
	ringWidth := state.RingWidth(gtx, style.RingWidth, frame.ActiveTheme(ctx).Motion)

	opacity := paint.PushOpacity(gtx.Ops, style.Opacity)
	drawInputGroupFrame(gtx, frame.ActiveTheme(ctx), image.Rectangle{Max: size}, radius, style, ringWidth)
	x := 0
	if g.prefix != nil {
		addRecordedInputGroupChild(gtx, prefix, image.Pt(prefixLeft, inputGroupChildY(size.Y, prefix.size.Y, verticalPadding, g.multiline)))
		drawInputGroupDivider(gtx, prefixWidth-divider, size.Y, divider, style)
		x += prefixWidth
	}
	addRecordedInputGroupChild(gtx, editorChild, image.Pt(x, inputGroupChildY(size.Y, editorChild.size.Y, 0, false)))
	x += editorChild.size.X
	if g.suffix != nil {
		drawInputGroupDivider(gtx, x, size.Y, divider, style)
		addRecordedInputGroupChild(gtx, suffix, image.Pt(x+divider+suffixLeft, inputGroupChildY(size.Y, suffix.size.Y, verticalPadding, g.multiline)))
	}
	opacity.Pop()
	addInputGroupPointer(gtx, &state.State, size, !enabled)
	return layout.Dimensions{Size: size}
}

func insetInputGroupEditor(editor layout.Widget, left, right, vertical unit.Dp) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: vertical, Right: right, Bottom: vertical, Left: left}.Layout(gtx, editor)
	}
}

func inputGroupTextAreaContentHeight(lineHeight, rows, maxHeight int) int {
	if lineHeight <= 0 || rows <= 0 || maxHeight <= 0 {
		return 0
	}
	if rows > maxHeight/lineHeight {
		return maxHeight
	}
	return lineHeight * rows
}

func inputGroupSlotHeight(childHeight, top int, present, multiline bool) int {
	if !present {
		return 0
	}
	if multiline {
		return childHeight + top
	}
	return childHeight
}

func inputGroupChildY(height, childHeight, top int, alignTop bool) int {
	if alignTop {
		return min(max(top, 0), max(height-childHeight, 0))
	}
	return max((height-childHeight)/2, 0)
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
