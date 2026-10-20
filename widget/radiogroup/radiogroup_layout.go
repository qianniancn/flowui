package radiogroup

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

func (r Widget) layoutItem(ctx *core.Context, gtx layout.Context, group *radioGroupState, item RadioItem, selectedKey string, visual theme.RadioGroupVisual) layout.Dimensions {
	itemState := group.item(item.Key)
	disabled := r.disabled || item.Disabled
	invalid := r.invalid || item.Invalid
	selected := item.Key == selectedKey
	animGtx := gtx

	presses := core.ActivePresses(itemState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for itemState.clickable.Clicked(gtx) {
			if !selected {
				group.requestSelectedKey(r, item.Key)
			}
		}
		core.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		focused := gtx.Focused(&itemState.clickable)
		focusVisible := core.FocusVisible(ctx, &itemState.clickable, focused)
		styleState := style.State{
			Hovered: itemState.clickable.Hovered(), Pressed: itemState.clickable.Pressed(),
			Focused: focused, FocusVisible: focusVisible, Disabled: disabled,
			Selected: selected, Checked: selected, Invalid: invalid,
		}
		key := core.DerivedKey(ctx, core.FullKey(ctx, r.key), item.Key)
		resolved := r.resolveItemStyle(ctx, animGtx, key, styleState, visual)
		resolved.selected = itemState.selectionFor(animGtx, selected, visual.SelectionDuration, visual.Motion)
		resolved.focus = itemState.focusOpacity(animGtx, focusVisible && !disabled, visual.FocusDuration, visual.Motion)
		scale := radioPressScale(animGtx, itemState.clickable.History(), visual, disabled)
		dims := r.layoutItemContent(ctx, gtx, item, resolved, scale, visual)
		semanticClip := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.RadioButton.Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		if item.Description != "" {
			semantic.DescriptionOp(item.Description).Add(gtx.Ops)
		}
		semanticClip.Pop()
		return dims
	})
}

func (r Widget) layoutItemContent(ctx *core.Context, gtx layout.Context, item RadioItem, style radioStyle, scale float32, visual theme.RadioGroupVisual) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(visual.Metrics.ContentGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return drawRadio(ctx, gtx, visual, style, scale)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if item.Description == "" {
				return layoutui.Resolved(ctx, gtx, style.label, text.New(item.Label))
			}
			return layout.Flex{
				Axis: layout.Vertical,
				Gap:  gtx.Dp(visual.Metrics.DescriptionGap),
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutui.Resolved(ctx, gtx, style.label, text.New(item.Label))
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutui.Resolved(ctx, gtx, style.description, text.New(item.Description))
				}),
			)
		}),
	)
}
