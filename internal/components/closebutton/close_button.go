package closebutton

import (
	"image"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const (
	closeButtonColorDuration = 100 * time.Millisecond
	closeButtonScaleDuration = 250 * time.Millisecond
)

type CloseButtonWidget struct {
	key      string
	onClick  func()
	disabled bool
	icon     frame.Widget
	label    string
}

func CloseButton(key string) CloseButtonWidget {
	return CloseButtonWidget{key: key}
}

func (b CloseButtonWidget) OnClick(fn func()) CloseButtonWidget {
	b.onClick = fn
	return b
}

func (b CloseButtonWidget) Disabled(disabled bool) CloseButtonWidget {
	b.disabled = disabled
	return b
}

func (b CloseButtonWidget) Icon(icon frame.Widget) CloseButtonWidget {
	b.icon = icon
	return b
}

func (b CloseButtonWidget) Label(label string) CloseButtonWidget {
	b.label = label
	return b
}

func (b CloseButtonWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key, clickable := frame.ClickableWithKey(ctx, b.key)
	return layoutWithClickable(b, ctx, gtx, clickable, closeButtonStateFor(ctx, key), true)
}

// LayoutWithClickableNoEvents renders a close button with caller-owned state and events.
func LayoutWithClickableNoEvents(b CloseButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable, buttonState *State) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, buttonState, false)
}

func layoutWithClickable(b CloseButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable, buttonState *State, handleEvents bool) layout.Dimensions {
	if clickable == nil {
		panic("flowui: nil close button clickable")
	}
	if buttonState == nil {
		buttonState = new(State)
	}
	animGtx := gtx
	presses := state.ActivePresses(clickable.History())
	enabled := gtx.Enabled() && !b.disabled
	if handleEvents {
		for clickable.Clicked(gtx) {
			if enabled && b.onClick != nil {
				b.onClick()
			}
		}
		if enabled {
			frame.FocusOnPress(ctx, clickable, clickable.History(), presses)
		}
	}

	size := closeButtonSize(gtx, frame.ActiveTheme(ctx).Components.CloseButton.Size)
	gtx.Constraints = layout.Exact(size)
	focused := gtx.Focused(clickable)
	focusVisible := frame.FocusVisible(ctx, clickable, focused)
	style := closeButtonStyleFor(frame.ActiveTheme(ctx), clickable.Hovered(), !enabled)
	motion := frame.ActiveTheme(ctx).Motion
	style.background = buttonState.background(animGtx, style.background, motion)
	style.focusOpacity = buttonState.focus.Opacity(animGtx, focusVisible && enabled, motion)
	targetScale := float32(1)
	if clickable.Pressed() && enabled {
		targetScale = closeButtonPressedScale(style.pressedScale)
	}
	scale := buttonState.scale(animGtx, targetScale, motion)

	if !enabled {
		semanticClip := clip.Rect{Max: size}.Push(gtx.Ops)
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(b.semanticLabel(ctx)).Add(gtx.Ops)
		semantic.EnabledOp(false).Add(gtx.Ops)
		b.drawVisual(ctx, gtx, size, style, scale, true)
		semanticClip.Pop()
		return layout.Dimensions{Size: size}
	}

	dims := clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(b.semanticLabel(ctx)).Add(gtx.Ops)
		semantic.EnabledOp(true).Add(gtx.Ops)
		b.drawVisual(ctx, gtx, size, style, scale, false)
		return layout.Dimensions{Size: size}
	})
	stack := render.Scale(size, scale).Push(gtx.Ops)
	drawCloseButtonFocus(gtx, image.Rectangle{Max: size}, closeButtonRadius(gtx, size, style.radius), style)
	stack.Pop()
	return dims
}

func (b CloseButtonWidget) drawVisual(ctx *frame.Context, gtx layout.Context, size image.Point, style closeButtonStyle, scale float32, disabled bool) {
	macro := op.Record(gtx.Ops)
	drawCloseButton(gtx, size, style)
	b.layoutIcon(ctx, gtx, size, style, disabled)
	call := macro.Stop()
	stack := render.Scale(size, scale).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()
}

func closeButtonSize(gtx layout.Context, preferred unit.Dp) image.Point {
	diameter := gtx.Dp(preferred)
	diameter = min(diameter, min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	diameter = max(diameter, 0)
	return image.Pt(diameter, diameter)
}

func (b CloseButtonWidget) semanticLabel(ctx *frame.Context) string {
	if b.label != "" {
		return b.label
	}
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "关闭"
	}
	return "Close"
}
