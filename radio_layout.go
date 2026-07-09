package flowui

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
)

func layoutRadioItems(gtx layout.Context, horizontal bool, columnGap, rowGap int, children []layout.Widget) layout.Dimensions {
	if horizontal {
		return layoutRadioWrap(gtx, columnGap, rowGap, children)
	}
	flexChildren := make([]layout.FlexChild, 0, len(children))
	for _, child := range children {
		flexChildren = append(flexChildren, layout.Rigid(child))
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  rowGap,
	}.Layout(gtx, flexChildren...)
}

func layoutRadioWrap(gtx layout.Context, columnGap, rowGap int, children []layout.Widget) layout.Dimensions {
	rows := make([]radioWrapRow, 0)
	maxWidth := gtx.Constraints.Max.X
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}

	var x, y, rowHeight, width int
	var row radioWrapRow
	for _, child := range children {
		macro := op.Record(gtx.Ops)
		dims := child(childGtx)
		call := macro.Stop()

		if x > 0 && x+columnGap+dims.Size.X > maxWidth {
			rows = append(rows, row)
			y += rowHeight + rowGap
			x = 0
			rowHeight = 0
			row = radioWrapRow{}
		}
		if x > 0 {
			x += columnGap
		}
		row.children = append(row.children, radioWrapChild{
			call: call,
			pos:  image.Pt(x, y),
		})
		x += dims.Size.X
		width = max(width, x)
		rowHeight = max(rowHeight, dims.Size.Y)
	}
	if len(row.children) > 0 {
		rows = append(rows, row)
	}

	size := gtx.Constraints.Constrain(image.Pt(width, y+rowHeight))
	for _, row := range rows {
		for _, child := range row.children {
			trans := op.Offset(child.pos).Push(gtx.Ops)
			child.call.Add(gtx.Ops)
			trans.Pop()
		}
	}
	return layout.Dimensions{Size: size}
}

type radioWrapRow struct {
	children []radioWrapChild
}

type radioWrapChild struct {
	call op.CallOp
	pos  image.Point
}

func (r RadioGroupWidget) layoutItem(ctx *Context, gtx layout.Context, group *radioGroupState, item RadioItem) layout.Dimensions {
	state := group.item(item.Key)
	disabled := r.disabled || item.Disabled
	invalid := r.invalid || item.Invalid
	selected := item.Key == r.selectedKey
	animGtx := gtx

	presses := activePresses(state.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for state.clickable.Clicked(gtx) {
			if !selected && r.onChange != nil {
				r.onChange(item.Key)
			}
		}
		ctx.focusOnPress(&state.clickable, state.clickable.History(), presses)
	}

	return state.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.RadioButton.Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		focusVisible := state.focusVisible(gtx.Focused(&state.clickable), state.clickable.History())
		style := radioStyleFor(ctx.Theme, r.variant, state.clickable.Hovered(), state.clickable.Pressed(), disabled, invalid)
		style.selected = state.selection(animGtx, selected)
		style.focus = state.focusOpacity(animGtx, focusVisible && !disabled)
		scale := radioPressScale(animGtx, state.clickable.History(), ctx.Theme, disabled)
		return r.layoutItemContent(ctx, gtx, item, style, scale)
	})
}

func (r RadioGroupWidget) layoutItemContent(ctx *Context, gtx layout.Context, item RadioItem, style radioStyle, scale float32) layout.Dimensions {
	theme := ctx.Theme.Components.RadioGroup
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(theme.ContentGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return drawRadio(gtx, ctx.Theme, style, scale)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if item.Description == "" {
				return Text(item.Label).
					Size(float32(theme.TextSize)).
					Weight(font.Medium).
					Color(style.fg).
					Layout(ctx, gtx)
			}
			return layout.Flex{
				Axis: layout.Vertical,
				Gap:  gtx.Dp(theme.DescriptionGap),
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return Text(item.Label).
						Size(float32(theme.TextSize)).
						Weight(font.Medium).
						Color(style.fg).
						Layout(ctx, gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return Text(item.Description).
						Size(float32(theme.DescriptionSize)).
						Color(style.description).
						Layout(ctx, gtx)
				}),
			)
		}),
	)
}
