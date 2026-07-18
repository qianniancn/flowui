package dropdown

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

func (d Widget) layoutRoot(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := dropdownStateFor(ctx, d.key)
	state.bind(d)
	open := state.isOpen(d)
	if d.disabled || !gtx.Enabled() {
		open = state.requestOpen(ctx, d, false)
	}
	if !open || frame.OverlayInteractive(ctx, frame.OverlayLayerPopup, state.key) {
		open = d.handleTrigger(ctx, gtx, state, open)
	}
	eventGtx := gtx
	if d.disabled {
		eventGtx = eventGtx.Disabled()
	}
	dims := d.layoutTrigger(ctx, eventGtx, state)
	state.triggerRect = image.Rectangle{Max: dims.Size}

	if open && !state.wasOpen {
		frame.ActivateExclusive(ctx, dropdownExclusive, state.key)
	} else if !open {
		frame.ReleaseExclusive(ctx, dropdownExclusive, state.key)
	}
	restoreFocus := !open && state.wasOpen && !d.disabled && gtx.Enabled() && !state.skipRestore
	state.observeOpen(open)
	if restoreFocus {
		frame.AfterOverlays(ctx, func() {
			if !frame.HasTopOverlay(ctx) {
				frame.RequestFocusVisible(ctx, &state.trigger, state.focusVisible)
			}
		})
	}

	progress := state.progress(gtx, open && !d.disabled, frame.ActiveTheme(ctx).Motion)
	if progress > 0 && !state.triggerRect.Empty() {
		d.registerRootOverlay(ctx, state, open, progress, !gtx.Enabled())
	}
	return dims
}

func (d Widget) handleTrigger(ctx *frame.Context, gtx layout.Context, state *dropdownState, open bool) bool {
	focusVisible := frame.FocusVisible(ctx, &state.trigger, gtx.Focused(&state.trigger))
	for state.trigger.Clicked(gtx) {
		if d.triggerMode != TriggerPress {
			continue
		}
		state.focusFirst = !open
		state.focusLast = false
		state.focusVisible = focusVisible
		open = state.requestOpen(ctx, d, !open)
		frame.RequestFocusVisible(ctx, &state.trigger, focusVisible)
	}
	open = d.handleTriggerKeys(ctx, gtx, state, open)
	if d.triggerMode == TriggerLongPress {
		open = d.handleLongPress(ctx, gtx, state, open)
	}
	return open
}

func (d Widget) handleTriggerKeys(ctx *frame.Context, gtx layout.Context, state *dropdownState, open bool) bool {
	filters := []event.Filter{
		key.Filter{Focus: &state.trigger, Name: key.NameDownArrow},
		key.Filter{Focus: &state.trigger, Name: key.NameUpArrow},
		key.Filter{Focus: &state.trigger, Name: key.NameEnter},
		key.Filter{Focus: &state.trigger, Name: key.NameReturn},
		key.Filter{Focus: &state.trigger, Name: key.NameSpace},
		key.Filter{Focus: &state.trigger, Name: key.NameEscape},
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return open
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		if event.Name == key.NameEscape {
			if open {
				state.focusFirst = false
				state.focusLast = false
				open = state.requestOpen(ctx, d, false)
			}
			continue
		}
		if open {
			continue
		}
		state.focusVisible = true
		state.focusFirst = event.Name != key.NameUpArrow
		state.focusLast = event.Name == key.NameUpArrow
		open = state.requestOpen(ctx, d, true)
	}
}

func (d Widget) handleLongPress(ctx *frame.Context, gtx layout.Context, state *dropdownState, open bool) bool {
	for {
		e, ok := gtx.Event(pointer.Filter{Target: &state.longPressTag, Kinds: pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Cancel})
		if !ok {
			break
		}
		event, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		switch event.Kind {
		case pointer.Press:
			if event.Buttons == 0 || event.Buttons.Contain(pointer.ButtonPrimary) {
				state.touchTracking = true
				state.longPressMoved = false
				state.pointerID = event.PointerID
				state.pointerStart = event.Position
				state.pointerAt = gtx.Now
			}
		case pointer.Move, pointer.Drag:
			if state.touchTracking && event.PointerID == state.pointerID {
				threshold := float32(gtx.Dp(10))
				if movedBeyond(state.pointerStart, event.Position, threshold) {
					state.touchTracking = false
					state.longPressMoved = true
				}
			}
		case pointer.Release:
			if event.PointerID == state.pointerID {
				state.touchTracking = false
			}
		}
	}
	history := state.trigger.History()
	if len(history) > 0 {
		press := history[len(history)-1]
		if press.Cancelled || !press.End.IsZero() {
			state.touchTracking = false
			state.longPressMoved = false
		} else if !state.longPressMoved {
			state.touchTracking = true
			state.pointerStart = f32.Pt(float32(press.Position.X), float32(press.Position.Y))
			state.pointerAt = press.Start
		}
	}
	if !state.touchTracking {
		return open
	}
	deadline := state.pointerAt.Add(dropdownLongPress)
	if gtx.Now.Before(deadline) {
		gtx.Execute(op.InvalidateCmd{At: deadline})
		return open
	}
	gtx.Execute(pointer.GrabCmd{Tag: &state.longPressTag, ID: state.pointerID})
	state.touchTracking = false
	state.focusFirst = true
	state.focusLast = false
	state.focusVisible = false
	frame.RequestFocusVisible(ctx, &state.trigger, false)
	return state.requestOpen(ctx, d, true)
}

func movedBeyond(start, current f32.Point, threshold float32) bool {
	dx := current.X - start.X
	if dx < 0 {
		dx = -dx
	}
	dy := current.Y - start.Y
	if dy < 0 {
		dy = -dy
	}
	return dx > threshold || dy > threshold
}

func (d Widget) layoutTrigger(ctx *frame.Context, gtx layout.Context, state *dropdownState) layout.Dimensions {
	if trigger, ok := d.trigger.(button.ButtonWidget); ok {
		return d.layoutButtonTrigger(ctx, gtx, state, trigger)
	}
	if trigger, ok := d.trigger.(*button.ButtonWidget); ok && trigger != nil {
		return d.layoutButtonTrigger(ctx, gtx, state, *trigger)
	}
	animGtx := gtx
	presses := stateutil.ActivePresses(state.trigger.History())
	if !d.disabled {
		frame.FocusOnPress(ctx, &state.trigger, state.trigger.History(), presses)
	}
	macro := op.Record(gtx.Ops)
	dims := state.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		macro := op.Record(gtx.Ops)
		var childDims layout.Dimensions
		if d.trigger != nil {
			pass := pointer.PassOp{}.Push(gtx.Ops)
			childDims = d.trigger.Layout(ctx, gtx)
			pass.Pop()
		}
		call := macro.Stop()
		size := gtx.Constraints.Constrain(childDims.Size)
		focusVisible := frame.FocusVisible(ctx, &state.trigger, gtx.Focused(&state.trigger))
		focus := state.triggerFocus.Opacity(animGtx, focusVisible && !d.disabled, frame.ActiveTheme(ctx).Motion)
		scale := dropdownTriggerScale(animGtx, state.trigger.History(), frame.ActiveTheme(ctx), d.disabled)
		transform := render.Scale(size, scale).Push(gtx.Ops)
		call.Add(gtx.Ops)
		drawDropdownTriggerFocus(gtx, frame.ActiveTheme(ctx), size, focus)
		transform.Pop()
		return layout.Dimensions{Size: size, Baseline: childDims.Baseline}
	})
	call := macro.Stop()
	if d.triggerMode == TriggerLongPress && dims.Size.X > 0 && dims.Size.Y > 0 {
		area := clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops)
		pass := pointer.PassOp{}.Push(gtx.Ops)
		event.Op(gtx.Ops, &state.longPressTag)
		pass.Pop()
		area.Pop()
	}
	pass := pointer.PassOp{}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	pass.Pop()
	return dims
}

func (d Widget) layoutButtonTrigger(ctx *frame.Context, gtx layout.Context, state *dropdownState, trigger button.ButtonWidget) layout.Dimensions {
	presses := stateutil.ActivePresses(state.trigger.History())
	if !d.disabled {
		frame.FocusOnPress(ctx, &state.trigger, state.trigger.History(), presses)
	}
	macro := op.Record(gtx.Ops)
	dims := button.LayoutWithClickableNoEvents(trigger, ctx, gtx, &state.trigger)
	call := macro.Stop()
	if d.triggerMode == TriggerLongPress && dims.Size.X > 0 && dims.Size.Y > 0 {
		area := clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops)
		pass := pointer.PassOp{}.Push(gtx.Ops)
		event.Op(gtx.Ops, &state.longPressTag)
		pass.Pop()
		area.Pop()
	}
	call.Add(gtx.Ops)
	return dims
}
