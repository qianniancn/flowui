package radiogroup

import (
	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

func (r RadioGroupWidget) layoutItem(ctx *frame.Context, gtx layout.Context, group *radioGroupState, item RadioItem) layout.Dimensions {
	itemState := group.item(item.Key)
	disabled := r.disabled || item.Disabled
	invalid := r.invalid || item.Invalid
	selected := item.Key == r.selectedKey
	animGtx := gtx

	presses := state.ActivePresses(itemState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for itemState.clickable.Clicked(gtx) {
			if !selected && r.onChange != nil {
				r.onChange(item.Key)
			}
		}
		frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.RadioButton.Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		focusVisible := frame.FocusVisible(ctx, &itemState.clickable, gtx.Focused(&itemState.clickable))
		style := radioStyleFor(frame.ActiveTheme(ctx), r.variant, itemState.clickable.Hovered(), itemState.clickable.Pressed(), disabled, invalid)
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
			return drawRadio(gtx, frame.ActiveTheme(ctx), style, scale)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if item.Description == "" {
				return text.New(item.Label).
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
					return text.New(item.Label).
						Size(float32(theme.TextSize)).
						Weight(font.Medium).
						Color(style.fg).
						Layout(ctx, gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return text.New(item.Description).
						Size(float32(theme.DescriptionSize)).
						Color(style.description).
						Layout(ctx, gtx)
				}),
			)
		}),
	)
}
