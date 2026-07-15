package togglebutton

import (
	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

func (b ToggleButtonWidget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	componentState := toggleButtonStateFor(ctx, b.key)
	clickable := &componentState.clickable
	animGtx := gtx
	enabled := gtx.Enabled() && !b.disabled
	frame.RegisterFocusGroupItem(ctx, clickable, enabled, componentState.focus.Prepare)
	presses := state.ActivePresses(clickable.History())
	for clickable.Clicked(gtx) {
		if enabled && b.onChange != nil {
			b.onChange(!b.selected)
		}
	}
	if enabled {
		frame.FocusOnPress(ctx, clickable, clickable.History(), presses)
	}

	sizeStyle := toggleButtonSizeStyleFor(frame.ActiveTheme(ctx), b.size)
	gtx = constrainToggleButton(gtx, sizeStyle.height, b.iconOnly)
	focusVisible := componentState.focus.Visible(gtx.Focused(clickable), clickable.History())
	style := toggleButtonStyleFor(
		frame.ActiveTheme(ctx),
		ctx.ForegroundColor(),
		b.variant,
		b.selected,
		clickable.Hovered() && enabled,
		clickable.Pressed() && enabled,
		!enabled,
	)
	style.background = componentState.animateBackground(animGtx, style.background)
	style.focus = componentState.focus.Opacity(animGtx, focusVisible && enabled)
	targetScale := float32(1)
	if clickable.Pressed() && enabled {
		targetScale = resolvedToggleButtonPressedScale(sizeStyle.pressedScale, b.size)
	}
	scale := componentState.animateScale(animGtx, targetScale)

	if !enabled {
		disabledGtx := gtx.Disabled()
		macro := op.Record(gtx.Ops)
		dims := b.layoutVisual(ctx, disabledGtx, style, sizeStyle, scale, false)
		call := macro.Stop()
		stack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		call.Add(gtx.Ops)
		stack.Pop()
		return dims
	}

	dims := clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return b.layoutVisual(ctx, gtx, style, sizeStyle, scale, true)
	})
	focusScale := render.Scale(dims.Size, scale).Push(gtx.Ops)
	drawToggleButtonFocus(gtx, dims.Size, style)
	focusScale.Pop()
	return dims
}

func (b ToggleButtonWidget) layoutVisual(ctx *frame.Context, gtx layout.Context, style toggleButtonStyle, sizeStyle toggleButtonSizeStyle, scale float32, enabled bool) layout.Dimensions {
	semantic.Button.Add(gtx.Ops)
	semantic.SelectedOp(b.selected).Add(gtx.Ops)
	semantic.EnabledOp(enabled).Add(gtx.Ops)
	if b.label != "" {
		semantic.LabelOp(b.label).Add(gtx.Ops)
	}

	macro := op.Record(gtx.Ops)
	dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return sizeStyle.inset(b.iconOnly).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return b.layoutChild(ctx, gtx, style, sizeStyle)
		})
	})
	call := macro.Stop()

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	transform := render.Scale(dims.Size, scale).Push(gtx.Ops)
	drawToggleButtonSurface(gtx, dims.Size, style)
	call.Add(gtx.Ops)
	transform.Pop()
	opacity.Pop()
	return dims
}

func (b ToggleButtonWidget) layoutChild(ctx *frame.Context, gtx layout.Context, style toggleButtonStyle, sizeStyle toggleButtonSizeStyle) layout.Dimensions {
	if b.child == nil {
		return layout.Dimensions{}
	}
	child := b.child
	if value, ok := child.(text.Widget); ok {
		child = value.
			DefaultColor(style.foreground).
			DefaultSize(sizeStyle.textSize).
			DefaultWeight(font.Medium)
	}
	restore := frame.PushColors(ctx, style.foreground, style.background)
	defer restore()
	return child.Layout(ctx, gtx)
}

func constrainToggleButton(gtx layout.Context, heightDp unit.Dp, iconOnly bool) layout.Context {
	height := gtx.Dp(heightDp)
	height = min(max(height, gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height
	if iconOnly {
		width := min(height, gtx.Constraints.Max.X)
		width = max(width, gtx.Constraints.Min.X)
		gtx.Constraints.Min.X = width
		gtx.Constraints.Max.X = width
	}
	return gtx
}
