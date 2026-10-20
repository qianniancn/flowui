package input

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type recordedInputGroupChild struct {
	call op.CallOp
	size image.Point
}

func (g GroupWidget) layoutFrame(ctx *core.Context, gtx layout.Context, state *inputGroupState, style inputGroupResolvedStyle, enabled bool, editor layout.Widget, actionVisual theme.InputActionVisual) layout.Dimensions {
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return g.layoutContent(ctx, gtx, state, style, editor, actionVisual)
	})
	return layoutui.InteractiveResolved(ctx, gtx, style.root, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		dims := visual(gtx)
		state.AddPointer(gtx, dims.Size, !enabled)
		return dims
	})
}

func (g GroupWidget) layoutContent(ctx *core.Context, gtx layout.Context, state *inputGroupState, style inputGroupResolvedStyle, editor layout.Widget, actionVisual theme.InputActionVisual) layout.Dimensions {
	actionMetrics := actionVisual.Metrics
	constraints := gtx.Constraints
	minHeight := constraints.Min.Y
	paddingLeftDp, paddingRightDp := resolvedHorizontalPadding(style.content)
	verticalPaddingDp := unit.Dp(0)
	if g.multiline {
		verticalPaddingDp = resolvedVerticalPadding(style.content)
	}
	verticalPadding := gtx.Dp(verticalPaddingDp)
	prefixLeftDp, prefixRightDp := paddingLeftDp, paddingRightDp
	if g.prefixAction {
		prefixLeftDp = actionMetrics.SlotPadding
		prefixRightDp = actionMetrics.SlotPadding
	}
	if left, right, ok := resolvedExplicitHorizontalPadding(style.prefix); ok {
		prefixLeftDp, prefixRightDp = left, right
	}
	prefixLeft, prefixRight := gtx.Dp(prefixLeftDp), gtx.Dp(prefixRightDp)
	suffixLeftDp, suffixRightDp := paddingLeftDp, paddingRightDp
	if g.suffixAction {
		suffixLeftDp = actionMetrics.SlotPadding
		suffixRightDp = actionMetrics.SlotPadding
	}
	if left, right, ok := resolvedExplicitHorizontalPadding(style.suffix); ok {
		suffixLeftDp, suffixRightDp = left, right
	}
	suffixLeft, suffixRight := gtx.Dp(suffixLeftDp), gtx.Dp(suffixRightDp)
	dividerDp := unit.Dp(0)
	if style.divider.Box != nil && style.divider.Box.Width != nil {
		dividerDp = *style.divider.Box.Width
	}
	divider := max(gtx.Dp(dividerDp), 0)

	prefix := recordInputGroupWidget(ctx, gtx, g.prefix, style.prefix, actionVisual, g.prefixAction)
	suffix := recordInputGroupWidget(ctx, gtx, g.suffix, style.suffix, actionVisual, g.suffixAction)
	prefixWidth := inputGroupSlotWidth(prefix, prefixLeft, prefixRight, divider, g.prefix != nil)
	suffixWidth := inputGroupSlotWidth(suffix, suffixLeft, suffixRight, divider, g.suffix != nil)
	leftPaddingDp := paddingLeftDp
	if g.prefix != nil {
		leftPaddingDp = 0
	}
	rightPaddingDp := paddingRightDp
	if g.suffix != nil {
		rightPaddingDp = 0
	}
	leftPadding, rightPadding := gtx.Dp(leftPaddingDp), gtx.Dp(rightPaddingDp)

	outerFixed := prefixWidth + suffixWidth + leftPadding + rightPadding
	maxEditorWidth := max(constraints.Max.X-outerFixed, 0)
	minEditorWidth := max(constraints.Min.X-outerFixed, 0)
	minEditorWidth = min(minEditorWidth, maxEditorWidth)
	editor = insetInputGroupEditor(editor, leftPaddingDp, rightPaddingDp, verticalPaddingDp)
	editor = inputGroupEditorWithCursor(editor, style.content)
	minEditorHeight := 0
	maxEditorHeight := constraints.Max.Y
	if g.multiline {
		lineHeight := max(gtx.Sp(style.editor.LineHeight), 0)
		contentHeight := inputGroupTextAreaContentHeight(lineHeight, g.textArea.resolvedRows(), constraints.Max.Y)
		minEditorHeight = max(gtx.Dp(resolvedMinHeight(style.content)), contentHeight+verticalPadding*2)
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
	x := 0
	var prefixActionBounds, suffixActionBounds image.Rectangle
	if g.prefix != nil {
		prefixOffset := image.Pt(prefixLeft, inputGroupChildY(size.Y, prefix.size.Y, verticalPadding, g.multiline))
		addRecordedInputGroupChild(gtx, prefix, prefixOffset)
		if g.prefixAction {
			prefixActionBounds = image.Rectangle{Min: prefixOffset, Max: prefixOffset.Add(prefix.size)}
		}
		layoutInputGroupDivider(ctx, gtx, prefixWidth-divider, size.Y, divider, style.divider)
		x += prefixWidth
	}
	addRecordedInputGroupChild(gtx, editorChild, image.Pt(x, inputGroupChildY(size.Y, editorChild.size.Y, 0, false)))
	x += editorChild.size.X
	if g.suffix != nil {
		layoutInputGroupDivider(ctx, gtx, x, size.Y, divider, style.divider)
		suffixOffset := image.Pt(x+divider+suffixLeft, inputGroupChildY(size.Y, suffix.size.Y, verticalPadding, g.multiline))
		addRecordedInputGroupChild(gtx, suffix, suffixOffset)
		if g.suffixAction {
			suffixActionBounds = image.Rectangle{Min: suffixOffset, Max: suffixOffset.Add(suffix.size)}
		}
	}
	state.setActionBounds(prefixActionBounds, suffixActionBounds)
	return layout.Dimensions{Size: size}
}

func resolvedHorizontalPadding(resolved style.ResolvedStyle) (unit.Dp, unit.Dp) {
	if resolved.Box == nil || resolved.Box.Padding == nil {
		return 0, 0
	}
	return resolved.Box.Padding.Left, resolved.Box.Padding.Right
}

func resolvedExplicitHorizontalPadding(resolved style.ResolvedStyle) (unit.Dp, unit.Dp, bool) {
	if resolved.Box == nil || resolved.Box.Padding == nil {
		return 0, 0, false
	}
	padding := resolved.Box.Padding
	return padding.Left, padding.Right, true
}

func resolvedVerticalPadding(resolved style.ResolvedStyle) unit.Dp {
	if resolved.Box == nil || resolved.Box.Padding == nil {
		return 0
	}
	return resolved.Box.Padding.Top
}

func resolvedMinHeight(resolved style.ResolvedStyle) unit.Dp {
	if resolved.Box == nil || resolved.Box.MinHeight == nil {
		return 0
	}
	return *resolved.Box.MinHeight
}

func inputGroupEditorWithCursor(editor layout.Widget, resolved style.ResolvedStyle) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		macro := op.Record(gtx.Ops)
		dims := editor(gtx)
		call := macro.Stop()
		call.Add(gtx.Ops)
		if resolved.Box != nil && resolved.Box.Cursor != nil {
			area := clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops)
			resolved.Box.Cursor.Add(gtx.Ops)
			area.Pop()
		}
		return dims
	}
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

func recordInputGroupWidget(ctx *core.Context, gtx layout.Context, widget core.Widget, resolved style.ResolvedStyle, actionVisual theme.InputActionVisual, actionSlot bool) recordedInputGroupChild {
	if widget == nil {
		return recordedInputGroupChild{}
	}
	child := widget
	if actionSlot {
		if action, ok := widget.(GroupActionWidget); ok {
			child = core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				if action.override == nil {
					return action.layoutWithVisual(ctx, gtx, actionVisual)
				}
				// The action owns an instance declaration, so it resolves a
				// bundle that carries it instead of the group's preview.
				visual := inputGroupActionVisualCached(ctx, action.key, theme.InputActionRequest{Override: action.override})
				return action.layoutWithVisual(ctx, gtx, visual)
			})
		}
	}
	return recordInputGroupLayout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Point{}
		return layoutui.Resolved(ctx, gtx, resolved, child)
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
