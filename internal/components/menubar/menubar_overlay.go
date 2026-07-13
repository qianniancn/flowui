package menubar

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/components/menu"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

func (m Widget) registerOverlay(ctx *frame.Context, state *menubarState, item Item, bar, trigger image.Rectangle, open bool, progress float32) {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       state.key,
		Layer:     frame.OverlayLayerPopup,
		Anchor:    bar,
		HasAnchor: true,
		Disabled:  m.disabled,
		Layout: func(gtx layout.Context, resolvedBar image.Rectangle, interactive bool) layout.Dimensions {
			resolvedTrigger := menubarResolvedChildRect(bar, resolvedBar, trigger)
			overlayOpen := m.handleOverlayEvents(ctx, gtx, state, item.key, open, interactive)
			return m.layoutOverlay(ctx, gtx, state, item, resolvedBar, resolvedTrigger, overlayOpen, progress, interactive && gtx.Enabled())
		},
	})
}

func (m Widget) handleOverlayEvents(ctx *frame.Context, gtx layout.Context, state *menubarState, itemKey string, open, interactive bool) bool {
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
		state.requestOpen(ctx, m, "")
		state.focusPanelKey = ""
		open = false
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
			state.requestOpen(ctx, m, "")
			state.focusPanelKey = ""
			state.focusTrigger(ctx, itemKey, true)
			open = false
		}
	}
	return open
}

func (m Widget) layoutOverlay(ctx *frame.Context, gtx layout.Context, state *menubarState, item Item, bar, trigger image.Rectangle, open bool, progress float32, interactive bool) layout.Dimensions {
	bounds := gtx.Constraints.Max
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Max: bounds}
	content := item.menu
	if m.disabled || item.disabled {
		content = content.Disabled(true)
	}

	var runtime menu.Runtime
	runtime = content.Runtime(ctx, state.key, "menu:"+item.key, func(focusVisible bool) {
		runtime.CloseSubmenus()
		state.requestOpen(ctx, m, "")
		state.focusPanelKey = ""
		state.focusTrigger(ctx, item.key, focusVisible)
	})
	if m.orientation == Horizontal {
		runtime = runtime.RootNavigation(
			func() {
				runtime.CloseSubmenus()
				m.switchFromMenu(ctx, gtx, state, item.key, -1)
			},
			func() {
				runtime.CloseSubmenus()
				m.switchFromMenu(ctx, gtx, state, item.key, 1)
			},
		)
	}
	if !open {
		runtime.CloseSubmenus()
	}

	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return runtime.Layout(ctx, panelGtx, open && (interactive || runtime.HasActiveSubmenu()))
	})
	panelCall := macro.Stop()

	placement := overlay.Placement{Side: overlay.SideBottom, Align: overlay.AlignStart}
	if m.orientation == Vertical {
		placement.Side = overlay.SideRight
	}
	result := overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          trigger.Size(),
		TriggerOrigin:    trigger.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           gtx.Dp(frame.ActiveTheme(ctx).Components.Menubar.PanelGap),
		Placement:        placement,
		Flip:             true,
		AvoidOverflow:    true,
	})
	tokens := frame.ActiveTheme(ctx).Components.Menu
	baseScale := tokens.EnterScale
	if !open {
		baseScale = tokens.ExitScale
	}
	scale := baseScale + (1-baseScale)*progress
	origin := overlay.PanelTransformOriginAt(trigger, result.Position, panelDims.Size, result.Placement)
	slide := overlay.SlideOffset(gtx.Dp(tokens.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := overlay.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	m.layoutDismissAreas(gtx, state, bounds, bar, animatedPanel)

	offset := op.Offset(panelOffset).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	m.layoutDialogBlocker(gtx, state, panelDims.Size)
	panelCall.Add(gtx.Ops)
	opacity.Pop()
	transform.Pop()
	offset.Pop()

	if open && interactive && state.focusPanelKey == item.key {
		frame.AfterOverlays(ctx, func() {
			if !frame.OverlayTopmost(ctx, frame.OverlayLayerPopup, state.key) {
				return
			}
			focused := false
			if state.focusLast {
				focused = runtime.FocusLast(ctx, state.focusVisible)
			} else {
				focused = runtime.FocusFirst(ctx, state.focusVisible)
			}
			if focused {
				state.focusPanelKey = ""
				state.focusLast = false
			}
		})
	}
	return layout.Dimensions{Size: bounds}
}

func (m Widget) switchFromMenu(ctx *frame.Context, gtx layout.Context, state *menubarState, currentKey string, delta int) {
	current := m.indexOf(currentKey)
	next := m.moveIndex(current, delta)
	if next < 0 || next >= len(m.items) || next == current {
		return
	}
	key := m.items[next].key
	if state.requestOpen(ctx, m, key) != key {
		return
	}
	state.panelKey = key
	state.focusPanelKey = key
	state.focusLast = false
	state.focusVisible = true
	gtx.Execute(op.InvalidateCmd{})
}

func menubarResolvedChildRect(localParent, resolvedParent, localChild image.Rectangle) image.Rectangle {
	if localParent.Empty() || resolvedParent.Empty() || localChild.Empty() {
		return image.Rectangle{}
	}
	mapX := func(value int) int {
		offset := value - localParent.Min.X
		return resolvedParent.Min.X + int(float64(offset)*float64(resolvedParent.Dx())/float64(localParent.Dx())+0.5)
	}
	mapY := func(value int) int {
		offset := value - localParent.Min.Y
		return resolvedParent.Min.Y + int(float64(offset)*float64(resolvedParent.Dy())/float64(localParent.Dy())+0.5)
	}
	return image.Rect(mapX(localChild.Min.X), mapY(localChild.Min.Y), mapX(localChild.Max.X), mapY(localChild.Max.Y))
}

func (m Widget) layoutDismissAreas(gtx layout.Context, state *menubarState, viewport image.Point, excluded ...image.Rectangle) {
	areas := overlay.DismissRectsExcluding(image.Rectangle{Max: viewport}, excluded...)
	for index, area := range areas {
		if index >= len(state.dismiss) || area.Empty() {
			break
		}
		areaGtx := gtx
		areaGtx.Constraints = layout.Exact(area.Size())
		offset := op.Offset(area.Min).Push(gtx.Ops)
		if m.modal {
			state.dismiss[index].Layout(areaGtx, func(layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: area.Size()}
			})
		} else {
			pass := pointer.PassOp{}.Push(gtx.Ops)
			state.dismiss[index].Layout(areaGtx, func(layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: area.Size()}
			})
			pass.Pop()
		}
		offset.Pop()
	}
}

func (m Widget) layoutDialogBlocker(gtx layout.Context, state *menubarState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}
