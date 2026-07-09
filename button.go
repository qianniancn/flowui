package flowui

import (
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
)

type ButtonWidget struct {
	key       string
	child     Widget
	onClick   func()
	variant   ButtonVariant
	size      ButtonSize
	disabled  bool
	loading   bool
	fullWidth bool
	iconOnly  bool
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

func Button(key string, child Widget) ButtonWidget {
	return ButtonWidget{key: key, child: child}
}

func (b ButtonWidget) OnClick(fn func()) ButtonWidget {
	b.onClick = fn
	return b
}

func (b ButtonWidget) Disabled(disabled bool) ButtonWidget {
	b.disabled = disabled
	return b
}

func (b ButtonWidget) Loading(loading bool) ButtonWidget {
	b.loading = loading
	return b
}

func (b ButtonWidget) Variant(variant ButtonVariant) ButtonWidget {
	b.variant = variant
	return b
}

func (b ButtonWidget) Size(size ButtonSize) ButtonWidget {
	b.size = size
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

func (b ButtonWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	key, clickable := ctx.clickableWithKey(b.key)
	state := ctx.buttonState(key)
	animGtx := gtx
	presses := activePresses(clickable.History())
	if b.disabled || b.loading {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			if b.onClick != nil {
				b.onClick()
			}
		}
		ctx.focusOnPress(clickable, clickable.History(), presses)
	}

	sizeStyle := buttonSizeStyle(ctx.Theme, b.size, b.iconOnly)

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

	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		focused := gtx.Focused(clickable)
		focusVisible := state.focusVisible(focused, clickable.History())
		style := b.style(ctx.Theme, clickable)
		style.bg = state.background(animGtx, style.bg)
		style.focus = state.focusOpacity(animGtx, focusVisible && !b.disabled)
		child := b.styleChild(style)
		if b.loading {
			animGtx.Execute(op.InvalidateCmd{})
		}

		semantic.Button.Add(gtx.Ops)
		macro := op.Record(gtx.Ops)
		dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return style.inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return b.layoutContent(ctx, gtx, style, child)
			})
		})
		call := macro.Stop()

		scale := buttonAnimationScale(gtx, clickable.History(), ctx.Theme, b.size, b.disabled)
		stack := buttonScale(dims.Size, scale).Push(gtx.Ops)
		drawButton(gtx, dims.Size, style)
		call.Add(gtx.Ops)
		stack.Pop()
		return dims
	})
}
