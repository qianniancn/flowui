package radiogroup

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

func (r RadioGroupWidget) layoutItem(ctx *frame.Context, gtx layout.Context, group *radioGroupState, item RadioItem, selectedKey string) layout.Dimensions {
	itemState := group.item(item.Key)
	disabled := r.disabled || item.Disabled
	invalid := r.invalid || item.Invalid
	selected := item.Key == selectedKey
	animGtx := gtx

	presses := state.ActivePresses(itemState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for itemState.clickable.Clicked(gtx) {
			if !selected {
				group.requestSelectedKey(r, item.Key)
			}
		}
		frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.RadioButton.Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		focused := gtx.Focused(&itemState.clickable)
		focusVisible := frame.FocusVisible(ctx, &itemState.clickable, focused)
		styleState := flowstyle.StyleState{
			Hovered: itemState.clickable.Hovered(), Pressed: itemState.clickable.Pressed(),
			Focused: focused, FocusVisible: focusVisible, Disabled: disabled,
			Selected: selected, Checked: selected, Invalid: invalid,
		}
		key := frame.DerivedKey(ctx, frame.FullKey(ctx, r.key), item.Key)
		style := r.resolveItemStyle(ctx, animGtx, key, styleState)
		motion := frame.ActiveTheme(ctx).Motion
		style.selected = itemState.selection(animGtx, selected, motion)
		style.focus = itemState.focusOpacity(animGtx, focusVisible && !disabled, motion)
		scale := radioPressScale(animGtx, itemState.clickable.History(), frame.ActiveTheme(ctx), disabled)
		return r.layoutItemContent(ctx, gtx, item, style, scale)
	})
}

func (r RadioGroupWidget) layoutItemContent(ctx *frame.Context, gtx layout.Context, item RadioItem, style radioStyle, scale float32) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.RadioGroup
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(theme.ContentGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return drawRadio(ctx, gtx, frame.ActiveTheme(ctx), style, scale)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if item.Description == "" {
				return layoutui.LayoutResolved(ctx, gtx, style.label, text.New(item.Label))
			}
			return layout.Flex{
				Axis: layout.Vertical,
				Gap:  gtx.Dp(theme.DescriptionGap),
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutui.LayoutResolved(ctx, gtx, style.label, text.New(item.Label))
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutui.LayoutResolved(ctx, gtx, style.description, text.New(item.Description))
				}),
			)
		}),
	)
}
