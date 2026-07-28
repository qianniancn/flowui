package dropdown

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
)

func (d Widget) registerRootOverlay(ctx *frame.Context, state *dropdownState, open bool, progress float32, disabled bool) {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       state.key,
		Layer:     frame.OverlayLayerPopup,
		Anchor:    state.triggerRect,
		HasAnchor: true,
		Disabled:  disabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			overlayOpen := d.handleOverlayEvents(ctx, gtx, state, open, interactive)
			return d.layoutRootOverlay(ctx, gtx, state, anchor, overlayOpen, progress, interactive && gtx.Enabled())
		},
	})
}

func (d Widget) handleOverlayEvents(ctx *frame.Context, gtx layout.Context, state *dropdownState, open, interactive bool) bool {
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
		dismissed = state.dismiss[index].TakePressed() || dismissed
	}
	if dismissed && open {
		state.skipRestore = true
		open = state.requestOpen(ctx, d, false)
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
			state.skipRestore = true
			open = state.requestOpen(ctx, d, false)
		}
	}
	return open
}

func (d Widget) layoutRootOverlay(ctx *frame.Context, gtx layout.Context, state *dropdownState, anchor image.Rectangle, open bool, progress float32, interactive bool) layout.Dimensions {
	bounds := gtx.Constraints.Max
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Max: bounds}
	var runtime menu.Runtime
	runtime = d.menu.Runtime(ctx, state.key, "menu", func(focusVisible bool) {
		runtime.CloseSubmenus()
		frame.RequestFocusVisible(ctx, &state.trigger, focusVisible)
		state.skipRestore = true
		state.requestOpen(ctx, d, false)
	})
	if !open {
		runtime.CloseSubmenus()
	}

	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return runtime.Layout(ctx, panelGtx, open && (interactive || runtime.HasActiveSubmenu()))
	})
	panelCall := macro.Stop()
	result := overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           d.panelGapPx(ctx, gtx),
		Placement:        d.placement.Placement(),
		Flip:             d.flipEnabled(),
		AvoidOverflow:    d.overflowAvoidanceEnabled(),
	})
	tokens := frame.ActiveTheme(ctx).Components.Menu
	baseScale := tokens.EnterScale
	if !open {
		baseScale = tokens.ExitScale
	}
	scale := baseScale + (1-baseScale)*progress
	origin := overlay.PanelTransformOriginAt(anchor, result.Position, panelDims.Size, result.Placement)
	slide := overlay.SlideOffset(gtx.Dp(tokens.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := overlay.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	d.layoutDismissAndBlocker(gtx, state, bounds, animatedPanel, anchor)

	offset := op.Offset(panelOffset).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	panelCall.Add(gtx.Ops)
	opacity.Pop()
	transform.Pop()
	offset.Pop()

	if open && interactive {
		frame.AfterOverlays(ctx, func() {
			if !frame.OverlayBecameTopmost(ctx, frame.OverlayLayerPopup, state.key) {
				return
			}
			if state.focusLast {
				runtime.FocusLast(ctx, state.focusVisible)
			} else if state.focusFirst {
				runtime.FocusFirst(ctx, state.focusVisible)
			}
			state.focusFirst = false
			state.focusLast = false
		})
	}
	return layout.Dimensions{Size: bounds}
}

func (d Widget) panelGapPx(ctx *frame.Context, gtx layout.Context) int {
	if d.hasOffset {
		return gtx.Dp(d.offset)
	}
	return gtx.Dp(frame.ActiveTheme(ctx).Components.Dropdown.PanelGap)
}

func (d Widget) layoutDismissAndBlocker(gtx layout.Context, state *dropdownState, viewport image.Point, blocker image.Rectangle, excluded ...image.Rectangle) {
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
	if blocker.Empty() {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(blocker.Size())
	offset := op.Offset(blocker.Min).Push(gtx.Ops)
	state.dialog.Layout(blockerGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: blocker.Size()}
	})
	offset.Pop()
}
