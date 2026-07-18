package button

import (
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

type ButtonWidget struct {
	key         string
	child       frame.Widget
	label       string
	onClick     func()
	variant     ButtonVariant
	variantSet  bool
	size        ButtonSize
	sizeSet     bool
	disabled    bool
	disabledSet bool
	loading     bool
	fullWidth   bool
	iconOnly    bool
	group       buttonGroupItemStyle
	prepared    buttonPreparedContent
	preparedSet bool
}

type ButtonVariant int

const (
	ButtonPrimary ButtonVariant = iota
	ButtonSecondary
	ButtonTertiary
	ButtonGhost
	ButtonOutline
	ButtonDanger
	ButtonDangerSoft
)

type ButtonSize int

const (
	ButtonMedium ButtonSize = iota
	ButtonSmall
	ButtonLarge
)

const (
	buttonPressInDuration  = 90 * time.Millisecond
	buttonPressOutDuration = 160 * time.Millisecond
	buttonColorDuration    = 100 * time.Millisecond
	buttonFocusDuration    = 100 * time.Millisecond
	buttonSpinnerPeriod    = 900 * time.Millisecond
)

func Button(key string, child frame.Widget) ButtonWidget {
	return ButtonWidget{key: key, child: child}
}

func (b ButtonWidget) OnClick(fn func()) ButtonWidget {
	b.onClick = fn
	return b
}

func (b ButtonWidget) Label(value string) ButtonWidget {
	b.label = value
	return b
}

func (b ButtonWidget) Disabled(disabled bool) ButtonWidget {
	b.disabled = disabled
	b.disabledSet = true
	return b
}

func (b ButtonWidget) Loading(loading bool) ButtonWidget {
	b.loading = loading
	return b
}

func (b ButtonWidget) Variant(variant ButtonVariant) ButtonWidget {
	b.variant = variant
	b.variantSet = true
	return b
}

func (b ButtonWidget) Size(size ButtonSize) ButtonWidget {
	b.size = size
	b.sizeSet = true
	return b
}

func (b ButtonWidget) FullWidth() ButtonWidget {
	b.fullWidth = true
	return b
}

func (b ButtonWidget) IconOnly() ButtonWidget {
	b.iconOnly = true
	return b
}

func (b ButtonWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, nil, true)
}

// LayoutWithClickable renders a button with caller-owned interaction state.
// It is intended for internal component composition.
func LayoutWithClickable(b ButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, true)
}

// LayoutWithClickableNoEvents renders a button with caller-owned event handling.
func LayoutWithClickableNoEvents(b ButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable) layout.Dimensions {
	return layoutWithClickable(b, ctx, gtx, clickable, false)
}

func layoutWithClickable(b ButtonWidget, ctx *frame.Context, gtx layout.Context, clickable *widget.Clickable, handleEvents bool) layout.Dimensions {
	var key string
	if clickable == nil {
		key, clickable = frame.ClickableWithKey(ctx, b.key)
	} else {
		key = frame.ClaimKey(ctx, state.KindClickable, b.key)
	}
	buttonState := buttonStateFor(ctx, key)
	frame.RegisterFocusGroupItem(ctx, clickable, gtx.Enabled() && !b.disabled && !b.loading)
	animGtx := gtx
	presses := state.ActivePresses(clickable.History())
	if b.disabled || b.loading {
		gtx = gtx.Disabled()
	} else if handleEvents {
		for clickable.Clicked(gtx) {
			if b.onClick != nil {
				b.onClick()
			}
		}
		frame.FocusOnPress(ctx, clickable, clickable.History(), presses)
	}

	sizeStyle := buttonSizeStyle(frame.ActiveTheme(ctx), b.size, b.iconOnly)

	if b.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	height := min(gtx.Dp(sizeStyle.height), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	if b.iconOnly && !b.fullWidth {
		width := min(height, gtx.Constraints.Max.X)
		gtx.Constraints.Max.X = width
		gtx.Constraints.Min.X = width
	}

	dims := clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		focused := gtx.Focused(clickable)
		focusVisible := frame.FocusVisible(ctx, clickable, focused)
		style := b.style(frame.ActiveTheme(ctx), clickable)
		if b.loading && !b.iconOnly && !b.fullWidth {
			style.inset = buttonLoadingInset(gtx, frame.ActiveTheme(ctx), b.size, style.inset)
		}
		motion := frame.ActiveTheme(ctx).Motion
		style.bg = buttonState.background(animGtx, style.bg, motion)
		style.focus = buttonState.focusOpacity(animGtx, focusVisible && !b.disabled, motion)
		child := b.styleChild(style)
		if b.loading {
			animGtx.Execute(op.InvalidateCmd{})
		}

		macro := op.Record(gtx.Ops)
		dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if b.preparedSet {
				b.prepared.call.Add(gtx.Ops)
				return b.prepared.dims
			}
			return style.inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				restore := frame.PushColors(ctx, style.fg, style.bg)
				defer restore()
				return b.layoutContent(ctx, gtx, style, child)
			})
		})
		call := macro.Stop()
		semanticClip := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		if b.label != "" {
			semantic.LabelOp(b.label).Add(gtx.Ops)
		}

		scale := float32(1)
		if !b.group.grouped {
			scale = buttonAnimationScale(gtx, clickable.History(), frame.ActiveTheme(ctx), b.size, b.disabled)
		}
		stack := render.Scale(dims.Size, scale).Push(gtx.Ops)
		drawButton(gtx, dims.Size, style)
		call.Add(gtx.Ops)
		stack.Pop()
		semanticClip.Pop()
		return dims
	})
	return dims
}
