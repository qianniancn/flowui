// Package core is the shared interaction kernel: input to State and intent callbacks.
package core

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/style"
)

// Click is the shared input state for button-like interactions.
type Click struct {
	Key       string
	Clickable *widget.Clickable
	Enabled   bool
	State     style.State
}

// Begin consumes queued clicks and derives the common interaction state.
// key must be non-empty for retained interaction (enforced in debug later).
func Begin(
	ctx *Context,
	gtx layout.Context,
	key string,
	clickable *widget.Clickable,
	enabled bool,
	handleEvents bool,
	onClick func(),
) Click {
	if clickable == nil {
		key, clickable = ClickableWithKey(ctx, key)
	} else {
		key = FullKey(ctx, key)
	}
	enabled = enabled && gtx.Enabled()
	RegisterFocusGroupItem(ctx, clickable, enabled)
	presses := ActivePresses(clickable.History())
	if handleEvents {
		for clickable.Clicked(gtx) {
			if enabled && onClick != nil {
				onClick()
			}
		}
		if enabled {
			FocusOnPress(ctx, clickable, clickable.History(), presses)
		}
	}
	focused := gtx.Focused(clickable)
	return Click{
		Key:       key,
		Clickable: clickable,
		Enabled:   enabled,
		State: style.State{
			Hovered:      enabled && clickable.Hovered(),
			Pressed:      enabled && clickable.Pressed(),
			Focused:      focused,
			FocusVisible: FocusVisible(ctx, clickable, focused),
			Disabled:     !enabled,
		},
	}
}

// Layout adds the hit target and button semantics around visual.
func (click Click) Layout(gtx layout.Context, visual layout.Widget, label string) layout.Dimensions {
	return click.layout(gtx, visual, label, true)
}

// LayoutVisualOverflow lays out a clickable visual without clipping paint to
// the hit target. Semantic operations and the pointer region remain bounded
// by the measured dimensions. Button styles use this for decorations such as
// the click wave, which intentionally expands beyond the control.
func (click Click) LayoutVisualOverflow(gtx layout.Context, visual layout.Widget, label string) layout.Dimensions {
	return click.layout(gtx, visual, label, false)
}

func (click Click) layout(gtx layout.Context, visual layout.Widget, label string, clipVisual bool) layout.Dimensions {
	if click.Clickable == nil {
		panic("flowui: nil click interaction")
	}
	var overflowCall op.CallOp
	semanticVisual := func(gtx layout.Context) layout.Dimensions {
		macro := op.Record(gtx.Ops)
		dims := visual(gtx)
		overflowCall = macro.Stop()
		clipStack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(click.Enabled).Add(gtx.Ops)
		if label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		if clipVisual {
			overflowCall.Add(gtx.Ops)
		}
		clipStack.Pop()
		return dims
	}
	var dims layout.Dimensions
	if !click.Enabled {
		dims = semanticVisual(gtx.Disabled())
	} else {
		dims = click.Clickable.Layout(gtx, semanticVisual)
	}
	if !clipVisual {
		overflowCall.Add(gtx.Ops)
	}
	return dims
}
