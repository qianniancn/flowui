package interaction

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

// Click is the shared input state for button-like interactions.
type Click struct {
	Key        string
	Clickable  *widget.Clickable
	Enabled    bool
	StyleState flowstyle.StyleState
}

// BeginClick consumes queued clicks and derives the common interaction state.
func BeginClick(
	ctx *frame.Context,
	gtx layout.Context,
	key string,
	clickable *widget.Clickable,
	enabled bool,
	handleEvents bool,
	onClick func(),
) Click {
	if clickable == nil {
		key, clickable = frame.ClickableWithKey(ctx, key)
	} else {
		key = frame.FullKey(ctx, key)
	}
	enabled = enabled && gtx.Enabled()
	frame.RegisterFocusGroupItem(ctx, clickable, enabled)
	presses := state.ActivePresses(clickable.History())
	if handleEvents {
		for clickable.Clicked(gtx) {
			if enabled && onClick != nil {
				onClick()
			}
		}
		if enabled {
			frame.FocusOnPress(ctx, clickable, clickable.History(), presses)
		}
	}
	focused := gtx.Focused(clickable)
	return Click{
		Key:       key,
		Clickable: clickable,
		Enabled:   enabled,
		StyleState: flowstyle.StyleState{
			Hovered:      enabled && clickable.Hovered(),
			Pressed:      enabled && clickable.Pressed(),
			Focused:      focused,
			FocusVisible: frame.FocusVisible(ctx, clickable, focused),
			Disabled:     !enabled,
		},
	}
}

// Layout adds the hit target and button semantics around visual.
func (click Click) Layout(gtx layout.Context, visual layout.Widget, label string) layout.Dimensions {
	if click.Clickable == nil {
		panic("flowui: nil click interaction")
	}
	semanticVisual := func(gtx layout.Context) layout.Dimensions {
		macro := op.Record(gtx.Ops)
		dims := visual(gtx)
		call := macro.Stop()
		bounds := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(click.Enabled).Add(gtx.Ops)
		if label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		call.Add(gtx.Ops)
		bounds.Pop()
		return dims
	}
	if !click.Enabled {
		return semanticVisual(gtx.Disabled())
	}
	return click.Clickable.Layout(gtx, semanticVisual)
}
