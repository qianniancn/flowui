package dropdown

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
)

func (d Widget) layoutRoot(ctx *core.Context, gtx layout.Context, override *theme.DropdownOverride, menuOverride *theme.MenuOverride) layout.Dimensions {
	state := d.preparedState
	visual := theme.DropdownVisual{}
	if state != nil && d.preparedVisual != nil && override == nil {
		visual = *d.preparedVisual
	} else {
		state = dropdownStateFor(ctx, d.key)
		visual = d.visual(ctx, override, &state.visualCache)
	}
	state.bind(d)
	open := state.isOpen(d)
	if open && state.openSource != OpenChangeContextMenu {
		state.hasContextAnchor = false
	}
	if d.disabled || !gtx.Enabled() {
		open = state.requestOpen(ctx, d, false)
	}
	if !open || core.OverlayInteractive(ctx, core.OverlayLayerPopup, state.key) {
		open = d.handleTrigger(ctx, gtx, state, open, visual)
	}
	eventGtx := gtx
	if d.disabled {
		eventGtx = eventGtx.Disabled()
	}
	focused := eventGtx.Focused(&state.trigger)
	focusVisible := core.FocusVisible(ctx, &state.trigger, focused)
	dims := layoutui.Styled(ctx, eventGtx, state.key, style.State{
		Hovered:      state.trigger.Hovered(),
		Pressed:      state.trigger.Pressed(),
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     d.disabled || !gtx.Enabled(),
		Open:         open,
	}, visual.Override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutTrigger(ctx, gtx, state, visual)
	}))
	state.triggerRect = image.Rectangle{Max: dims.Size}

	if open && !state.wasOpen {
		core.ActivateExclusive(ctx, dropdownExclusive, state.key)
	} else if !open {
		core.ReleaseExclusive(ctx, dropdownExclusive, state.key)
	}
	restoreFocus := !open && state.wasOpen && !d.disabled && gtx.Enabled() && !state.skipRestore
	state.observeOpen(open)
	if restoreFocus {
		core.AfterOverlays(ctx, func() {
			if !core.HasTopOverlay(ctx) {
				core.RequestFocusVisible(ctx, &state.trigger, state.focusVisible)
			}
		})
	}

	progress := state.progressFor(gtx, open && !d.disabled, visual.EnterDuration, visual.ExitDuration, visual.Motion)
	if progress > 0 && !state.triggerRect.Empty() {
		d.registerRootOverlay(ctx, state, open, progress, !gtx.Enabled(), visual, menuOverride)
	}
	return dims
}

func (d Widget) handleTrigger(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool, visual theme.DropdownVisual) bool {
	focusVisible := core.FocusVisible(ctx, &state.trigger, gtx.Focused(&state.trigger))
	for state.trigger.Clicked(gtx) {
		if d.triggerMode != TriggerPress {
			continue
		}
		state.focusFirst = !open
		state.focusLast = false
		state.focusVisible = focusVisible
		state.hasContextAnchor = false
		open = state.requestOpenFrom(ctx, d, !open, OpenChangeTrigger)
		core.RequestFocusVisible(ctx, &state.trigger, focusVisible)
	}
	open = d.handleTriggerKeys(ctx, gtx, state, open)
	if d.triggerMode == TriggerLongPress {
		open = d.handleLongPress(ctx, gtx, state, open, visual.LongPressDuration)
	} else if d.triggerMode == TriggerHover {
		open = d.handleHover(ctx, gtx, state, open, visual.HoverOpenDuration, visual.HoverCloseDuration)
	} else if d.triggerMode == TriggerContextMenu {
		open = d.handleContextMenu(ctx, gtx, state, open)
	}
	return open
}

func (d Widget) handleHover(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool, defaultOpen, defaultClose time.Duration) bool {
	if d.disabled || !gtx.Enabled() {
		state.hoverOpenAt = time.Time{}
		state.hoverCloseAt = time.Time{}
		return false
	}

	triggerHovered := state.trigger.Hovered()
	menuHovered := state.menuHovered
	if open && (triggerHovered || menuHovered) {
		state.hoverOpenAt = time.Time{}
		state.hoverCloseAt = time.Time{}
		return open
	}
	if triggerHovered || menuHovered {
		state.hoverCloseAt = time.Time{}
	} else if open {
		if state.hoverCloseAt.IsZero() {
			closeDelay := defaultClose
			if d.hasHoverCloseDelay {
				closeDelay = d.hoverCloseDelay
			}
			state.hoverCloseAt = gtx.Now.Add(closeDelay)
		}
		if gtx.Now.Before(state.hoverCloseAt) {
			gtx.Execute(op.InvalidateCmd{At: state.hoverCloseAt})
			return open
		}
		state.hoverCloseAt = time.Time{}
		state.hoverOpenAt = time.Time{}
		return state.requestOpenFrom(ctx, d, false, OpenChangeOutside)
	}

	if !triggerHovered {
		state.hoverOpenAt = time.Time{}
		return open
	}
	if state.hoverOpenAt.IsZero() {
		openDelay := defaultOpen
		if d.hasHoverOpenDelay {
			openDelay = d.hoverOpenDelay
		}
		state.hoverOpenAt = gtx.Now.Add(openDelay)
	}
	if gtx.Now.Before(state.hoverOpenAt) {
		gtx.Execute(op.InvalidateCmd{At: state.hoverOpenAt})
		return open
	}
	state.hoverOpenAt = time.Time{}
	state.focusFirst = true
	state.focusLast = false
	state.focusVisible = false
	return state.requestOpenFrom(ctx, d, true, OpenChangeTrigger)
}

func (d Widget) handleTriggerKeys(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool) bool {
	filters := []event.Filter{
		key.Filter{Focus: &state.trigger, Name: key.NameDownArrow},
		key.Filter{Focus: &state.trigger, Name: key.NameUpArrow},
		key.Filter{Focus: &state.trigger, Name: key.NameEnter},
		key.Filter{Focus: &state.trigger, Name: key.NameReturn},
		key.Filter{Focus: &state.trigger, Name: key.NameSpace},
		key.Filter{Focus: &state.trigger, Name: key.NameEscape},
	}
	if d.triggerMode == TriggerContextMenu {
		filters = append(filters, key.Filter{Focus: &state.trigger, Name: key.NameF10, Required: key.ModShift})
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
				open = state.requestOpenFrom(ctx, d, false, OpenChangeKeyboard)
			}
			continue
		}
		if event.Name == key.NameF10 && d.triggerMode == TriggerContextMenu {
			state.contextAnchor = image.Rect(state.triggerRect.Dx()/2, state.triggerRect.Dy()/2, state.triggerRect.Dx()/2+1, state.triggerRect.Dy()/2+1)
			state.hasContextAnchor = true
			state.focusFirst = true
			state.focusLast = false
			state.focusVisible = true
			open = state.requestOpenFrom(ctx, d, true, OpenChangeContextMenu)
			continue
		}
		if open {
			continue
		}
		state.focusVisible = true
		state.focusFirst = event.Name != key.NameUpArrow
		state.focusLast = event.Name == key.NameUpArrow
		state.hasContextAnchor = false
		open = state.requestOpenFrom(ctx, d, true, OpenChangeKeyboard)
	}
}

func (d Widget) handleLongPress(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool, defaultDuration time.Duration) bool {
	for {
		eventValue, ok := core.NextPointerEvent(gtx, &state.longPressTag, pointer.Press|pointer.Release|pointer.Move|pointer.Drag|pointer.Cancel)
		if !ok {
			break
		}
		switch eventValue.Kind {
		case pointer.Press:
			if eventValue.Buttons == 0 || core.IsPrimaryPointerPress(eventValue) {
				state.touchTracking = true
				state.longPressMoved = false
				state.pointerID = eventValue.PointerID
				state.pointerStart = eventValue.Position
				state.pointerAt = gtx.Now
			}
		case pointer.Move, pointer.Drag:
			if state.touchTracking && eventValue.PointerID == state.pointerID {
				threshold := float32(gtx.Dp(10))
				if movedBeyond(state.pointerStart, eventValue.Position, threshold) {
					state.touchTracking = false
					state.longPressMoved = true
				}
			}
		case pointer.Release:
			if eventValue.PointerID == state.pointerID {
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
	longPressDelay := defaultDuration
	if d.hasLongPressDelay {
		longPressDelay = d.longPressDelay
	}
	deadline := state.pointerAt.Add(longPressDelay)
	if gtx.Now.Before(deadline) {
		gtx.Execute(op.InvalidateCmd{At: deadline})
		return open
	}
	gtx.Execute(pointer.GrabCmd{Tag: &state.longPressTag, ID: state.pointerID})
	state.touchTracking = false
	state.focusFirst = true
	state.focusLast = false
	state.focusVisible = false
	core.RequestFocusVisible(ctx, &state.trigger, false)
	return state.requestOpenFrom(ctx, d, true, OpenChangeTrigger)
}

func (d Widget) handleContextMenu(ctx *core.Context, gtx layout.Context, state *dropdownState, open bool) bool {
	for {
		eventValue, ok := core.NextPointerEvent(gtx, &state.contextTag, pointer.Press)
		if !ok {
			break
		}
		if eventValue.Source != pointer.Mouse || !eventValue.Buttons.Contain(pointer.ButtonSecondary) {
			continue
		}
		x := int(eventValue.Position.X + 0.5)
		y := int(eventValue.Position.Y + 0.5)
		state.contextAnchor = image.Rect(x, y, x+1, y+1)
		state.hasContextAnchor = true
		state.focusFirst = true
		state.focusLast = false
		state.focusVisible = false
		core.RequestFocusVisible(ctx, &state.trigger, false)
		open = state.requestOpenFrom(ctx, d, true, OpenChangeContextMenu)
	}
	return open
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

func (d Widget) layoutTrigger(ctx *core.Context, gtx layout.Context, state *dropdownState, visual theme.DropdownVisual) layout.Dimensions {
	if trigger, ok := d.trigger.(button.Widget); ok {
		return d.layoutButtonTrigger(ctx, gtx, state, trigger, visual)
	}
	if trigger, ok := d.trigger.(*button.Widget); ok && trigger != nil {
		return d.layoutButtonTrigger(ctx, gtx, state, *trigger, visual)
	}
	animGtx := gtx
	presses := core.ActivePresses(state.trigger.History())
	if !d.disabled {
		core.FocusOnPress(ctx, &state.trigger, state.trigger.History(), presses)
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
		focusVisible := core.FocusVisible(ctx, &state.trigger, gtx.Focused(&state.trigger))
		focus := state.triggerFocus.Opacity(animGtx, focusVisible && !d.disabled, visual.Motion)
		scale := dropdownTriggerScale(animGtx, state.trigger.History(), visual, d.disabled)
		transform := draw.Scale(size, scale).Push(gtx.Ops)
		call.Add(gtx.Ops)
		visual.TriggerFocus(gtx, theme.DropdownTriggerFocusRequest{Size: size, Opacity: focus})
		transform.Pop()
		return layout.Dimensions{Size: size, Baseline: childDims.Baseline}
	})
	call := macro.Stop()
	d.registerTriggerPointerEvents(gtx, state, dims.Size)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	pass.Pop()
	return dims
}

func (d Widget) layoutButtonTrigger(ctx *core.Context, gtx layout.Context, state *dropdownState, trigger button.Widget, visual theme.DropdownVisual) layout.Dimensions {
	presses := core.ActivePresses(state.trigger.History())
	if !d.disabled {
		core.FocusOnPress(ctx, &state.trigger, state.trigger.History(), presses)
	}
	macro := op.Record(gtx.Ops)
	dims := button.LayoutWithClickableNoEvents(trigger.Style(visual.Override), ctx, gtx, &state.trigger)
	call := macro.Stop()
	d.registerTriggerPointerEvents(gtx, state, dims.Size)
	call.Add(gtx.Ops)
	return dims
}

func (d Widget) registerTriggerPointerEvents(gtx layout.Context, state *dropdownState, size image.Point) {
	if (d.triggerMode != TriggerLongPress && d.triggerMode != TriggerContextMenu) || size.X <= 0 || size.Y <= 0 {
		return
	}
	area := clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	if d.triggerMode == TriggerLongPress {
		event.Op(gtx.Ops, &state.longPressTag)
	} else {
		event.Op(gtx.Ops, &state.contextTag)
	}
	pass.Pop()
	area.Pop()
}
