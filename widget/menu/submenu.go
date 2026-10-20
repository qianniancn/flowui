package menu

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

func (m Widget) updateSubmenuHover(gtx layout.Context, state *menuState, item Item, hovered bool, delay time.Duration) {
	if !hovered {
		return
	}
	if !itemHasSubmenu(item) {
		state.hoverSubmenu = ""
		state.openSubmenu = ""
		return
	}
	if state.openSubmenu == item.Key {
		return
	}
	if state.hoverSubmenu != item.Key {
		state.hoverSubmenu = item.Key
		state.hoverAt = gtx.Now
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(delay)})
		return
	}
	if gtx.Now.Sub(state.hoverAt) < delay {
		gtx.Execute(op.InvalidateCmd{At: state.hoverAt.Add(delay)})
		return
	}
	state.openSubmenu = item.Key
	state.submenuFocusVisible = false
}

func (m Widget) registerSubmenus(ctx *core.Context, gtx layout.Context, state *menuState, interactive bool, visual theme.MenuVisual) {
	for _, entry := range m.actionableEntries() {
		item := entry.item
		if !itemHasSubmenu(item) {
			continue
		}
		anchor, hasAnchor := state.anchors[item.Key]
		if !hasAnchor || anchor.Empty() {
			continue
		}
		child := m.submenu(state, item)
		childState := child.stateFor(ctx)
		open := state.openSubmenu == item.Key
		if open && !childState.submenuWasOpen {
			childState.focusPending = true
			childState.requestedFocusVisible = state.submenuFocusVisible
		}
		childState.submenuWasOpen = open
		progress := childState.submenuProgress(gtx, open, visual.Metrics.EnterDuration, visual.Metrics.ExitDuration, visual.Motion)
		if progress <= 0 {
			continue
		}
		state.submenuActive = true
		child.registerSubmenuOverlay(ctx, gtx, childState, anchor, open, progress, visual)
	}
}

func (s *menuState) submenuProgress(gtx layout.Context, open bool, enter, exit time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := exit
	if open {
		target = 1
		duration = enter
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (m Widget) registerSubmenuOverlay(ctx *core.Context, gtx layout.Context, state *menuState, anchor image.Rectangle, open bool, progress float32, visual theme.MenuVisual) {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       state.key,
		Layer:     core.OverlayLayerPopup,
		Anchor:    anchor,
		HasAnchor: true,
		Disabled:  !gtx.Enabled(),
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			overlayOpen := m.handleSubmenuEvents(ctx, gtx, state, open, interactive)
			return m.layoutSubmenuOverlay(ctx, gtx, state, anchor, overlayOpen, progress, interactive && gtx.Enabled(), visual)
		},
	})
}

func (m Widget) handleSubmenuEvents(ctx *core.Context, gtx layout.Context, state *menuState, open, interactive bool) bool {
	for state.dialog.Clicked(gtx) {
	}
	if state.dialog.TakePressed() {
		core.PreserveFocus(ctx)
	}
	dismissed := false
	for index := range state.dismiss {
		for state.dismiss[index].Clicked(gtx) {
			dismissed = true
		}
	}
	if dismissed && open {
		m.dismissToParent()
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
			m.closeToParent(ctx)
			open = false
		}
	}
	return open
}

func (m Widget) layoutSubmenuOverlay(ctx *core.Context, gtx layout.Context, state *menuState, anchor image.Rectangle, open bool, progress float32, interactive bool, visual theme.MenuVisual) layout.Dimensions {
	bounds := gtx.Constraints.Max
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Max: bounds}
	if !open {
		state.openSubmenu = ""
	}
	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return m.layoutRootWithVisual(ctx, panelGtx, state, open && (interactive || state.openSubmenu != "" || state.submenuActive), visual)
	})
	panelCall := macro.Stop()

	metrics := visual.Metrics
	result := core.ResolvePosition(core.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           gtx.Dp(metrics.SubmenuGap),
		Placement:        core.Placement{Side: core.SideRight, Align: core.AlignStart},
		Flip:             true,
		AvoidOverflow:    true,
	})
	origin := core.PanelTransformOriginAt(anchor, result.Position, panelDims.Size, result.Placement)
	baseScale := metrics.EnterScale
	if !open {
		baseScale = metrics.ExitScale
	}
	scale := baseScale + (1-baseScale)*progress
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(result.Position.X), float32(result.Position.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := core.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	layoutSubmenuDismissAreas(gtx, state, bounds, animatedPanel)

	offset := op.Offset(result.Position).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	layoutSubmenuDialogBlocker(gtx, state, panelDims.Size)
	panelCall.Add(gtx.Ops)
	opacity.Pop()
	transform.Pop()
	offset.Pop()

	if open && interactive && state.focusPending {
		core.AfterOverlays(ctx, func() {
			if core.OverlayTopmost(ctx, core.OverlayLayerPopup, state.key) {
				state.focusFirstEntry(ctx, m, state.requestedFocusVisible)
				state.focusPending = false
			}
		})
	}
	return layout.Dimensions{Size: bounds}
}

func layoutSubmenuDismissAreas(gtx layout.Context, state *menuState, viewport image.Point, excluded ...image.Rectangle) {
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
}

func layoutSubmenuDialogBlocker(gtx layout.Context, state *menuState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}
