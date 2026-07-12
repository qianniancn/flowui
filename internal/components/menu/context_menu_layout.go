package menu

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

const (
	contextMenuEnterDuration = 150 * time.Millisecond
	contextMenuExitDuration  = 100 * time.Millisecond
)

func (c ContextMenuWidget) layoutTrigger(ctx *frame.Context, gtx layout.Context, state *contextMenuState, open *bool) layout.Dimensions {
	if !c.disabled && gtx.Enabled() {
		c.updateTrigger(ctx, gtx, state, open)
	} else {
		state.trigger.touchTracking = false
	}
	macro := op.Record(gtx.Ops)
	var dims layout.Dimensions
	if c.trigger != nil {
		dims = c.trigger.Layout(ctx, gtx)
	}
	call := macro.Stop()
	state.triggerSize = dims.Size

	area := clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &state.trigger)
	pass.Pop()
	call.Add(gtx.Ops)
	area.Pop()
	return dims
}

func (c ContextMenuWidget) updateTrigger(ctx *frame.Context, gtx layout.Context, state *contextMenuState, open *bool) {
	for {
		e, ok := gtx.Event(contextMenuTriggerFilters(&state.trigger)...)
		if !ok {
			break
		}
		switch event := e.(type) {
		case key.Event:
			if event.Name == key.NameF10 && event.State == key.Press && event.Modifiers&key.ModShift != 0 {
				point := f32.Pt(float32(state.triggerSize.X)/2, float32(state.triggerSize.Y)/2)
				state.anchor = contextMenuPointRect(point, 1)
				state.hasAnchor = true
				state.focusVisible = true
				*open = state.requestOpen(ctx, c, true)
			}
		case pointer.Event:
			c.updateTriggerPointer(ctx, gtx, state, open, event)
		}
	}
	c.updateLongPress(ctx, gtx, state, open)
}

func (c ContextMenuWidget) updateTriggerPointer(ctx *frame.Context, gtx layout.Context, state *contextMenuState, open *bool, event pointer.Event) {
	if event.Source == pointer.Mouse {
		if event.Kind == pointer.Press && event.Buttons.Contain(pointer.ButtonSecondary) {
			state.anchor = contextMenuPointRect(event.Position, 1)
			state.hasAnchor = true
			state.trigger.touchTracking = false
			state.focusVisible = false
			frame.RequestFocusVisible(ctx, &state.trigger, false)
			*open = state.requestOpen(ctx, c, true)
		}
		return
	}
	if c.longPressDisabled || event.Source != pointer.Touch {
		return
	}
	switch event.Kind {
	case pointer.Press:
		state.trigger.touchTracking = true
		state.trigger.touchID = event.PointerID
		state.trigger.touchStart = event.Position
		state.trigger.touchAt = gtx.Now
	case pointer.Move, pointer.Drag:
		if state.trigger.touchTracking && event.PointerID == state.trigger.touchID {
			threshold := float32(gtx.Dp(10))
			if contextMenuMovedBeyond(state.trigger.touchStart, event.Position, threshold) {
				state.trigger.touchTracking = false
			}
		}
	case pointer.Release, pointer.Cancel:
		if event.PointerID == state.trigger.touchID {
			state.trigger.touchTracking = false
		}
	}
}

func (c ContextMenuWidget) updateLongPress(ctx *frame.Context, gtx layout.Context, state *contextMenuState, open *bool) {
	if !state.trigger.touchTracking {
		return
	}
	deadline := state.trigger.touchAt.Add(contextMenuLongPressDelay)
	if gtx.Now.Before(deadline) {
		gtx.Execute(op.InvalidateCmd{At: deadline})
		return
	}
	gtx.Execute(pointer.GrabCmd{Tag: &state.trigger, ID: state.trigger.touchID})
	state.anchor = contextMenuPointRect(state.trigger.touchStart, gtx.Dp(10))
	state.hasAnchor = true
	state.trigger.touchTracking = false
	state.focusVisible = false
	frame.RequestFocusVisible(ctx, &state.trigger, false)
	*open = state.requestOpen(ctx, c, true)
}

func (c ContextMenuWidget) registerOverlay(ctx *frame.Context, state *contextMenuState, open bool, progress float32, disabled bool) {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       state.key,
		Layer:     frame.OverlayLayerPopup,
		Anchor:    state.anchor,
		HasAnchor: true,
		Disabled:  disabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			overlayOpen := c.handleOverlayEvents(ctx, gtx, state, open, interactive)
			return c.layoutOverlay(ctx, gtx, state, anchor, overlayOpen, progress, interactive && gtx.Enabled())
		},
	})
}

func (c ContextMenuWidget) handleOverlayEvents(ctx *frame.Context, gtx layout.Context, state *contextMenuState, open, interactive bool) bool {
	for state.dialog.Clicked(gtx) {
	}
	if state.dialog.TakePressed() {
		frame.PreserveFocus(ctx)
	}
	dismissed := false
	for index := range state.dismiss {
		for state.dismiss[index].Clicked(gtx) {
			dismissed = true
		}
	}
	if dismissed && open {
		state.skipRestore = true
		open = state.requestOpen(ctx, c, false)
	}
	if !interactive || !open {
		return open
	}
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if ok && event.State == key.Press {
			frame.RequestFocusVisible(ctx, &state.trigger, true)
			open = state.requestOpen(ctx, c, false)
		}
	}
	return open
}

func (c ContextMenuWidget) layoutOverlay(ctx *frame.Context, gtx layout.Context, state *contextMenuState, anchor image.Rectangle, open bool, progress float32, interactive bool) layout.Dimensions {
	bounds := gtx.Constraints.Max
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Max: bounds}
	rootMenu := c.menu.withDerivedIdentity(state.key, "menu")
	menuState := rootMenu.stateFor(ctx)
	rootMenu = rootMenu.withClose(func(focusVisible bool) {
		menuState.openSubmenu = ""
		frame.RequestFocusVisible(ctx, &state.trigger, focusVisible)
		state.skipRestore = true
		state.requestOpen(ctx, c, false)
	})
	if !open {
		menuState.openSubmenu = ""
	}
	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return rootMenu.layout(ctx, panelGtx, menuState, open && (interactive || menuState.openSubmenu != "" || menuState.submenuActive))
	})
	panelCall := macro.Stop()

	tokens := frame.ActiveTheme(ctx).Components.Menu
	result := overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           gtx.Dp(tokens.ContextMenuOffset),
		Placement:        overlay.Placement{Side: overlay.SideBottom, Align: overlay.AlignStart},
		Flip:             true,
		AvoidOverflow:    true,
	})
	origin := overlay.PanelTransformOriginAt(anchor, result.Position, panelDims.Size, result.Placement)
	baseScale := tokens.EnterScale
	if !open {
		baseScale = tokens.ExitScale
	}
	scale := baseScale + (1-baseScale)*progress
	slide := overlay.SlideOffset(gtx.Dp(tokens.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := overlay.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	c.layoutDismissAreas(gtx, state, bounds, animatedPanel)

	offset := op.Offset(panelOffset).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	c.layoutDialogBlocker(gtx, state, panelDims.Size)
	panelCall.Add(gtx.Ops)
	opacity.Pop()
	transform.Pop()
	offset.Pop()

	if open && interactive {
		frame.AfterOverlays(ctx, func() {
			if frame.OverlayBecameTopmost(ctx, frame.OverlayLayerPopup, state.key) {
				menuState.focusFirstEntry(ctx, rootMenu, state.focusVisible)
			}
		})
	}
	return layout.Dimensions{Size: bounds}
}

func (c ContextMenuWidget) layoutDismissAreas(gtx layout.Context, state *contextMenuState, viewport image.Point, excluded ...image.Rectangle) {
	if viewport.X <= 0 || viewport.Y <= 0 {
		return
	}
	areas := overlay.DismissRectsExcluding(image.Rectangle{Max: viewport}, excluded...)
	for index, area := range areas {
		if index >= len(state.dismiss) || area.Empty() {
			break
		}
		areaGtx := gtx
		areaGtx.Constraints = layout.Exact(area.Size())
		offset := op.Offset(area.Min).Push(gtx.Ops)
		pass := pointer.PassOp{}.Push(gtx.Ops)
		state.dismiss[index].Layout(areaGtx, func(layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: area.Size()}
		})
		pass.Pop()
		offset.Pop()
	}
}

func (c ContextMenuWidget) layoutDialogBlocker(gtx layout.Context, state *contextMenuState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}
