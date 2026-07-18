package menubar

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type menubarTriggerPlacement struct {
	dims      layout.Dimensions
	placement frame.OverlayPlacement
}

func (m Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if !gtx.Enabled() {
		m.disabled = true
	}
	state := menubarStateFor(ctx, m.key)
	state.bind(m)
	state.beginFrame(m.items)
	defer state.endFrame()

	openKey := state.current(m)
	if !m.hasOpenKey && state.openKey != "" && openKey == "" {
		openKey = state.requestOpen(ctx, m, "")
	}
	state.observeOpen(ctx, openKey)
	state.updateInteractions(ctx, gtx, m)
	openKey = state.current(m)
	state.observeOpen(ctx, openKey)
	progress := state.progress(gtx, openKey != "", frame.ActiveTheme(ctx).Motion)

	macro := op.Record(gtx.Ops)
	dims, triggerRects := m.layoutTriggers(ctx, gtx, state, openKey)
	call := macro.Stop()
	root := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	semantic.EnabledOp(!m.disabled).Add(gtx.Ops)
	label := m.alt
	if label == "" {
		label = "Menubar"
	}
	semantic.DescriptionOp(label).Add(gtx.Ops)
	call.Add(gtx.Ops)
	root.Pop()

	if progress > 0 && state.panelKey != "" {
		if item, ok := m.itemByKey(state.panelKey); ok {
			if anchor, ok := triggerRects[state.panelKey]; ok && !anchor.Empty() {
				m.registerOverlay(ctx, state, item, image.Rectangle{Max: dims.Size}, anchor, openKey == state.panelKey, progress)
			}
		}
	} else if progress <= 0 {
		state.panelKey = ""
	}
	return dims
}

func (m Widget) layoutTriggers(ctx *frame.Context, gtx layout.Context, state *menubarState, openKey string) (layout.Dimensions, map[string]image.Rectangle) {
	axis := layout.Horizontal
	if m.orientation == Vertical {
		axis = layout.Vertical
	}
	placements := make([]menubarTriggerPlacement, len(m.items))
	children := make([]layout.FlexChild, 0, len(m.items))
	for index, item := range m.items {
		index := index
		item := item
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return m.layoutTrigger(ctx, gtx, state, item, openKey == item.key)
			})
			placements[index] = menubarTriggerPlacement{dims: dims, placement: placement}
			return dims
		}))
	}
	gap := max(gtx.Dp(m.themeTokens(frame.ActiveTheme(ctx)).Gap), 0)
	dims := layout.Flex{Axis: axis, Alignment: layout.Middle, Gap: gap}.Layout(gtx, children...)

	maxCross := menubarCrossMinimum(gtx.Constraints, axis)
	for _, child := range placements {
		maxCross = max(maxCross, axis.Convert(child.dims.Size).Y)
	}
	rects := make(map[string]image.Rectangle, len(m.items))
	main := 0
	for index, child := range placements {
		size := axis.Convert(child.dims.Size)
		cross := max((maxCross-size.Y)/2, 0)
		position := axis.Convert(image.Pt(main, cross))
		child.placement.PlaceOffset(position)
		rects[m.items[index].key] = image.Rectangle{Min: position, Max: position.Add(child.dims.Size)}
		main += size.X
		if index < len(placements)-1 {
			main += gap
		}
	}
	return dims, rects
}

func menubarCrossMinimum(constraints layout.Constraints, axis layout.Axis) int {
	if axis == layout.Horizontal {
		return constraints.Min.Y
	}
	return constraints.Min.X
}

func (m Widget) layoutTrigger(ctx *frame.Context, gtx layout.Context, state *menubarState, item Item, open bool) layout.Dimensions {
	triggerState := state.trigger(item.key)
	disabled := m.itemDisabled(item) || !gtx.Enabled()
	eventGtx := gtx
	if disabled {
		eventGtx = eventGtx.Disabled()
	}
	animGtx := gtx
	tokens := m.themeTokens(frame.ActiveTheme(ctx))
	height := min(gtx.Dp(tokens.TriggerHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)

	return triggerState.clickable.Layout(eventGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(item.label).Add(gtx.Ops)
		semantic.SelectedOp(open).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)

		style := menubarTriggerStyleFor(
			frame.ActiveTheme(ctx),
			open,
			triggerState.clickable.Hovered(),
			triggerState.clickable.Pressed(),
			disabled,
		)
		focusVisible := frame.FocusVisible(ctx, &triggerState.clickable, gtx.Focused(&triggerState.clickable))
		focus := triggerState.focus.Opacity(animGtx, focusVisible && !disabled, frame.ActiveTheme(ctx).Motion)
		padding := max(gtx.Dp(tokens.TriggerPaddingX), 0)
		contentGtx := gtx
		contentGtx.Constraints.Min = image.Point{}
		contentGtx.Constraints.Max.X = max(contentGtx.Constraints.Max.X-padding*2, 0)
		macro := op.Record(gtx.Ops)
		background := style.background
		if background.A == 0 {
			background = ctx.BackgroundColor()
		}
		restore := frame.PushColors(ctx, style.foreground, background)
		var contentDims layout.Dimensions
		if item.trigger != nil {
			contentDims = item.trigger.Layout(ctx, contentGtx)
		} else {
			contentDims = text.New(item.label).
				Size(float32(tokens.TriggerTextSize)).
				Weight(font.Normal).
				Color(style.foreground).
				Layout(ctx, contentGtx)
		}
		restore()
		contentCall := macro.Stop()
		size := gtx.Constraints.Constrain(image.Pt(contentDims.Size.X+padding*2, max(height, contentDims.Size.Y)))
		opacity := paint.PushOpacity(gtx.Ops, style.opacity)
		drawMenubarTrigger(gtx, tokens, size, style, focus)
		position := image.Pt(max((size.X-contentDims.Size.X)/2, 0), max((size.Y-contentDims.Size.Y)/2, 0))
		offset := op.Offset(position).Push(gtx.Ops)
		contentCall.Add(gtx.Ops)
		offset.Pop()
		opacity.Pop()
		return layout.Dimensions{Size: size, Baseline: contentDims.Baseline + position.Y}
	})
}
