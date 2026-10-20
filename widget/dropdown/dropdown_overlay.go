package dropdown

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/menu"
)

func (d Widget) registerRootOverlay(ctx *core.Context, state *dropdownState, open bool, progress float32, disabled bool, dropdownVisual theme.DropdownVisual, menuOverride *theme.MenuOverride) {
	anchor := state.triggerRect
	if state.hasContextAnchor {
		anchor = state.contextAnchor
	}
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       state.key,
		Layer:     core.OverlayLayerPopup,
		Anchor:    anchor,
		HasAnchor: true,
		Disabled:  disabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			overlayOpen, dismissed := d.handleOverlayEvents(ctx, gtx, state, open, interactive)
			if dismissed {
				core.DismissActiveOverlay(ctx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}
			return d.layoutRootOverlay(ctx, gtx, state, anchor, overlayOpen, progress, interactive && gtx.Enabled(), dropdownVisual, menuOverride)
		},
	})
}

func (d Widget) handleOverlayEvents(ctx *core.Context, gtx layout.Context, state *dropdownState, open, interactive bool) (bool, bool) {
	for state.dialog.Clicked(gtx) {
	}
	if state.dialog.TakePressed() {
		core.PreserveFocus(ctx)
	}
	for state.arrow.Clicked(gtx) {
	}
	if state.arrow.TakePressed() {
		// The arrow is part of the popup surface. Keep the current focus while
		// the pointer crosses the gap between the trigger and the menu.
		core.PreserveFocus(ctx)
	}
	dismissed := false
	for index := range state.dismiss {
		for state.dismiss[index].Clicked(gtx) {
			dismissed = true
		}
		if state.dismiss[index].TakePressed() {
			// Menus dismiss on the outside press. Finish the exit transition in
			// the same frame so a held pointer cannot leave a faded panel visible.
			dismissed = true
			core.PreserveFocus(ctx)
		}
	}
	if dismissed && open {
		state.skipRestore = true
		open = state.requestOpenFrom(ctx, d, false, OpenChangeOutside)
		state.transition.Set(0, 0, gtx.Now)
	}
	if !interactive || !open {
		return open, dismissed
	}
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if ok && event.State == key.Press {
			core.RequestFocusVisible(ctx, &state.trigger, true)
			state.skipRestore = true
			open = state.requestOpenFrom(ctx, d, false, OpenChangeKeyboard)
		}
	}
	return open, dismissed
}

func (d Widget) layoutRootOverlay(ctx *core.Context, gtx layout.Context, state *dropdownState, anchor image.Rectangle, open bool, progress float32, interactive bool, dropdownVisual theme.DropdownVisual, menuOverride *theme.MenuOverride) layout.Dimensions {
	bounds := gtx.Constraints.Max
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Max: bounds}
	menuWidget := d.menu
	if d.matchTriggerWidth {
		menuWidget = menu.WithMinimumWidthPx(menuWidget, anchor.Dx())
	}
	var runtime menu.Runtime
	runtime = menuWidget.RuntimeWithOverride(ctx, state.key, "menu", func(focusVisible bool) {
		runtime.CloseSubmenus()
		core.RequestFocusVisible(ctx, &state.trigger, focusVisible)
		state.skipRestore = true
		state.requestOpenFrom(ctx, d, false, OpenChangeMenu)
	}, menuOverride)
	visual := runtime.Visual(ctx)
	if !open {
		runtime.CloseSubmenus()
	}
	state.menuHovered = state.dialog.Hovered() || runtime.HoveredWithSubmenus(ctx) || (d.arrow && state.arrow.Hovered())

	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return runtime.LayoutWithVisual(ctx, panelGtx, open && (interactive || runtime.HasActiveSubmenu()), visual)
	})
	panelCall := macro.Stop()
	result := core.ResolvePosition(core.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           d.panelGapPx(gtx, dropdownVisual),
		Placement:        d.placement.Placement(),
		Flip:             d.flipEnabled(),
		AvoidOverflow:    d.overflowAvoidanceEnabled(),
	})
	placement := result.Placement.PopoverPlacement()
	arrowSize := 0
	arrowAnchor := float32(0)
	arrowRect := image.Rectangle{}
	if d.arrow {
		arrowSize = gtx.Dp(dropdownVisual.ArrowSize)
		panelRadius := min(max(gtx.Dp(visual.Metrics.Radius), 0), min(panelDims.Size.X, panelDims.Size.Y)/2)
		arrowAnchor = core.ArrowAnchor(anchor, result.Position, panelDims.Size, placement, panelRadius, arrowSize)
		arrowRect = core.ArrowRect(panelDims.Size, placement, arrowAnchor, arrowSize)
	}
	tokens := visual.Metrics
	baseScale := tokens.EnterScale
	if !open {
		baseScale = tokens.ExitScale
	}
	scale := baseScale + (1-baseScale)*progress
	origin := core.PanelTransformOriginAt(anchor, result.Position, panelDims.Size, result.Placement)
	slide := core.SlideOffset(gtx.Dp(tokens.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := core.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	animatedArrow := core.AffineRectBounds(arrowRect, panelTransform)
	d.layoutDismissAndBlocker(gtx, state, bounds, animatedPanel, anchor, animatedArrow)

	offset := op.Offset(panelOffset).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	panelCall.Add(gtx.Ops)
	if d.arrow {
		surface, border := runtime.PanelColorsWithVisual(ctx, visual)
		core.DrawArrow(gtx, placement, panelDims.Size, arrowAnchor, arrowSize, gtx.Dp(tokens.BorderWidth), surface, border)
		layoutArrowBlocker(gtx, state, arrowRect)
	}
	opacity.Pop()
	transform.Pop()
	offset.Pop()

	if open && interactive {
		core.AfterOverlays(ctx, func() {
			if !core.OverlayBecameTopmost(ctx, core.OverlayLayerPopup, state.key) {
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

func (d Widget) panelGapPx(gtx layout.Context, dropdownVisual theme.DropdownVisual) int {
	gap := 0
	if d.hasOffset {
		gap = gtx.Dp(d.offset)
	} else {
		gap = gtx.Dp(dropdownVisual.PanelGap)
	}
	if d.arrow {
		gap += gtx.Dp(dropdownVisual.ArrowSize * 2 / 3)
	}
	return gap
}

func (d Widget) layoutDismissAndBlocker(gtx layout.Context, state *dropdownState, viewport image.Point, blocker image.Rectangle, excluded ...image.Rectangle) {
	areas := core.DismissRectsExcluding(image.Rectangle{Max: viewport}, excluded...)
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

func layoutArrowBlocker(gtx layout.Context, state *dropdownState, rect image.Rectangle) {
	if rect.Empty() {
		return
	}
	arrowGtx := gtx
	arrowGtx.Constraints = layout.Exact(rect.Size())
	offset := op.Offset(rect.Min).Push(gtx.Ops)
	state.arrow.Layout(arrowGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: rect.Size()}
	})
	offset.Pop()
}
