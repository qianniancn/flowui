package flowui

import (
	"image"
	"image/color"
	"math"
	"strings"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
)

func (s SelectWidget) layout(ctx *Context, gtx layout.Context, state *selectState, open bool) layout.Dimensions {
	style := selectStyleFor(ctx.Theme, s.variant, state.trigger.Hovered(), false, s.disabled, s.invalid)
	gap := gtx.Dp(ctx.Theme.Components.Select.ContentGap)
	var labelDims, triggerDims layout.Dimensions
	children := make([]layout.FlexChild, 0, 3)
	if s.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			labelDims = s.layoutLabel(ctx, gtx, style)
			return labelDims
		}))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		triggerDims = s.layoutTrigger(ctx, gtx, state, open)
		return triggerDims
	}))
	if message, messageColor := s.supportMessage(style); message != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text(message).
				Size(float32(ctx.Theme.Components.Select.DescriptionSize)).
				Color(messageColor).
				Layout(ctx, gtx)
		}))
	}
	dims := layout.Flex{
		Axis: layout.Vertical,
		Gap:  gap,
	}.Layout(gtx, children...)
	triggerY := 0
	if s.label != "" {
		triggerY = labelDims.Size.Y + gap
	}
	state.triggerRect = image.Rectangle{
		Min: image.Pt(0, triggerY),
		Max: image.Pt(triggerDims.Size.X, triggerY+triggerDims.Size.Y),
	}
	return dims
}

func (s SelectWidget) layoutLabel(ctx *Context, gtx layout.Context, style selectStyle) layout.Dimensions {
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text(s.label).
				Size(float32(ctx.Theme.Components.Select.LabelTextSize)).
				Weight(font.Medium).
				Color(style.label).
				Layout(ctx, gtx)
		}),
	}
	if s.required {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text("*").
				Size(float32(ctx.Theme.Components.Select.LabelTextSize)).
				Color(style.error).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Baseline,
		Gap:       gtx.Dp(ctx.Theme.Components.Select.RequiredMarkOffset),
	}.Layout(gtx, children...)
}

func (s SelectWidget) supportMessage(style selectStyle) (string, color.NRGBA) {
	if s.invalid {
		if s.errorMessage == "" {
			return "", style.error
		}
		return s.errorMessage, style.error
	}
	return s.description, style.description
}

func (s SelectWidget) layoutTrigger(ctx *Context, gtx layout.Context, state *selectState, open bool) layout.Dimensions {
	animGtx := gtx
	if s.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	theme := ctx.Theme.Components.Select
	height := min(gtx.Dp(theme.Height), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)

	return state.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		value, selected := s.displayValue()
		focusVisible := state.focus.focusVisible(gtx.Focused(&state.trigger), state.trigger.History())
		style := selectStyleFor(ctx.Theme, s.variant, state.trigger.Hovered(), focusVisible, s.disabled, s.invalid)
		style.field.bg = state.field.background(animGtx, style.field.bg)
		style.field.border = state.field.borderColor(animGtx, style.field.border)

		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(s.semanticLabel(value)).Add(gtx.Ops)
		if s.description != "" || (s.invalid && s.errorMessage != "") {
			description := s.description
			if s.invalid && s.errorMessage != "" {
				description = s.errorMessage
			}
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)

		left := gtx.Dp(theme.TriggerPaddingX)
		right := gtx.Dp(theme.IndicatorWidth)
		vertical := gtx.Dp(theme.TriggerPaddingY)
		maxX := max(gtx.Constraints.Max.X-left-right, 0)
		minX := min(max(gtx.Constraints.Min.X-left-right, 0), maxX)
		valueGtx := gtx
		valueGtx.Constraints = layout.Constraints{Min: image.Pt(minX, 0), Max: image.Pt(maxX, gtx.Constraints.Max.Y)}

		macro := op.Record(gtx.Ops)
		valueDims := layoutSelectValue(ctx, valueGtx, value, selected, style)
		valueCall := macro.Stop()

		size := image.Pt(valueDims.Size.X+left+right, max(valueDims.Size.Y+vertical*2, height))
		size = gtx.Constraints.Constrain(size)
		rect := image.Rectangle{Max: size}
		radius := min(max(gtx.Dp(theme.Radius), 1), min(size.X, size.Y)/2)
		drawSelectTrigger(gtx, rect, radius, style)

		valueOffset := op.Offset(image.Pt(left, max((size.Y-valueDims.Size.Y)/2, 0))).Push(gtx.Ops)
		valueCall.Add(gtx.Ops)
		valueOffset.Pop()

		indicatorSize := image.Pt(min(right, size.X), size.Y)
		indicatorOffset := op.Offset(image.Pt(max(size.X-indicatorSize.X, 0), 0)).Push(gtx.Ops)
		s.layoutIndicator(ctx, animGtx, indicatorSize, state.iconProgress(animGtx, open), style.field.placeholder)
		indicatorOffset.Pop()
		return layout.Dimensions{Size: size}
	})
}

func (s SelectWidget) layoutIndicator(ctx *Context, gtx layout.Context, size image.Point, progress float32, color color.NRGBA) layout.Dimensions {
	if s.indicator == nil {
		drawSelectIndicator(gtx, ctx.Theme, size, progress, color)
		return layout.Dimensions{Size: size}
	}
	gtx.Constraints = layout.Exact(size)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	transform := op.Affine(f32.AffineId().Rotate(center, progress*float32(math.Pi))).Push(gtx.Ops)
	indicator := s.indicator
	if text, ok := indicator.(TextWidget); ok && !text.hasColor {
		indicator = text.Color(color)
	}
	dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		iconSize := min(gtx.Dp(ctx.Theme.Components.Select.IndicatorSize), min(size.X, size.Y))
		gtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
		return indicator.Layout(ctx, gtx)
	})
	transform.Pop()
	return dims
}

func layoutSelectValue(ctx *Context, gtx layout.Context, value string, selected bool, style selectStyle) layout.Dimensions {
	label := material.Label(ctx.Theme.Material, ctx.Theme.Components.Select.TextSize, value)
	label.Color = style.field.placeholder
	if selected {
		label.Color = style.field.fg
	}
	return label.Layout(gtx)
}

func (s SelectWidget) semanticLabel(value string) string {
	parts := make([]string, 0, 2)
	if s.label != "" {
		parts = append(parts, s.label)
	}
	if value != "" {
		parts = append(parts, value)
	}
	return strings.Join(parts, ", ")
}

func (s SelectWidget) layoutPopover(ctx *Context, gtx layout.Context, state *selectState, triggerRect image.Rectangle, open bool, progress float32) {
	trigger := triggerRect.Size()
	bounds := gtx.Constraints.Max
	if trigger.X <= 0 || trigger.Y <= 0 || bounds.X <= 0 || bounds.Y <= 0 {
		return
	}
	ctx.deferOverlay(gtx, func(gtx layout.Context) layout.Dimensions {
		return s.layoutOverlay(ctx, gtx, state, triggerRect, bounds, open, progress)
	})
}

func (s SelectWidget) layoutOverlay(ctx *Context, gtx layout.Context, state *selectState, triggerRect image.Rectangle, bounds image.Point, open bool, progress float32) layout.Dimensions {
	theme := ctx.Theme.Components.Select
	trigger := triggerRect.Size()
	width := min(trigger.X, bounds.X)
	maxHeight := min(gtx.Dp(theme.PanelMaxHeight), bounds.Y)
	if width <= 0 || maxHeight <= 0 {
		return layout.Dimensions{Size: bounds}
	}
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Min: image.Pt(width, 0), Max: image.Pt(width, maxHeight)}
	macro := op.Record(gtx.Ops)
	panelDims := s.layoutPanel(ctx, panelGtx, state, open)
	panelCall := macro.Stop()

	gap := gtx.Dp(theme.PanelGap)
	result := overlayResolvePosition(overlayPositionConfig{
		Trigger:          trigger,
		TriggerOrigin:    triggerRect.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           gap,
		Placement:        popoverOverlayPlacement(s.placement),
		Flip:             s.flipEnabled(),
		AvoidOverflow:    s.overflowAvoidanceEnabled(),
	})
	s.layoutDismissAreas(ctx, gtx, state, triggerRect.Union(result.Rect))

	origin := overlayPanelTransformOriginAt(triggerRect, result.Position, panelDims.Size, result.Placement)
	scale := theme.AnimationScale + (1-theme.AnimationScale)*progress
	slide := overlaySlideOffset(gtx.Dp(theme.AnimationDistance), progress, result.Placement)
	offset := op.Offset(result.Position.Add(slide)).Push(gtx.Ops)
	transform := op.Affine(f32.AffineId().Scale(origin, f32.Pt(scale, scale))).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	s.layoutDialogBlocker(gtx, state, panelDims.Size)
	panelCall.Add(gtx.Ops)
	opacity.Pop()
	transform.Pop()
	offset.Pop()
	return layout.Dimensions{Size: bounds}
}

func (s SelectWidget) layoutPanel(ctx *Context, gtx layout.Context, state *selectState, open bool) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	var dims layout.Dimensions
	func() {
		restore := ctx.pushForeground(ctx.Theme.Palette.overlayForegroundColor())
		defer restore()
		list := s.listBox(ctx, state, open)
		dims = list.Layout(ctx, gtx)
	}()
	call := macro.Stop()

	radius := min(max(gtx.Dp(ctx.Theme.Components.Select.PanelRadius), 1), min(dims.Size.X, dims.Size.Y)/2)
	rect := image.Rectangle{Max: dims.Size}
	drawSelectPanel(gtx, ctx.Theme, rect, radius)
	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	if open {
		s.focusPendingOption(ctx, state)
	}
	return dims
}

func (s SelectWidget) listBox(ctx *Context, state *selectState, open bool) ListBoxWidget {
	key := s.listBoxKey()
	var list ListBoxWidget
	if s.selectionMode == SelectSelectionMultiple {
		if len(s.sections) > 0 {
			list = ListBoxMultipleSections(key, s.selectedKeys, s.sections)
		} else {
			list = ListBoxMultiple(key, s.selectedKeys, s.items)
		}
		list = list.OnSelectionChange(s.onSelectionChange)
	} else {
		if len(s.sections) > 0 {
			list = ListBoxSections(key, s.selectedKey, s.sections)
		} else {
			list = ListBox(key, s.selectedKey, s.items)
		}
		list = list.
			OnChange(s.onChange).
			OnAction(func(string) {
				state.focusIntent = selectFocusNone
				state.requestOpen(ctx, s, false)
				ctx.requestFocus(&state.trigger)
			})
	}
	return list.
		EmptyText(s.emptyText).
		DisabledKeys(s.disabledKeys).
		Disabled(!open || s.disabled).
		withPadding(ctx.Theme.Components.Select.PanelPadding)
}

func (s SelectWidget) listBoxKey() string {
	return s.key + ":options"
}

func (s SelectWidget) focusPendingOption(ctx *Context, state *selectState) {
	if state.focusIntent == selectFocusNone {
		return
	}
	items := s.allItems()
	index, ok := s.focusOptionIndex(items, state.focusIntent)
	focusVisible := state.focusVisibleIntent
	state.focusIntent = selectFocusNone
	state.focusVisibleIntent = false
	if !ok {
		return
	}
	listState := ctx.listBoxes[ctx.fullKey(s.listBoxKey())]
	if listState == nil || listState.items[items[index].Key] == nil {
		return
	}
	itemState := listState.items[items[index].Key]
	itemState.prepareFocus(focusVisible)
	ctx.requestFocus(&itemState.clickable)
}

func (s SelectWidget) focusOptionIndex(items []ListBoxItem, intent selectFocusIntent) (int, bool) {
	switch intent {
	case selectFocusFirst:
		return listBoxFirstEnabled(items, s.disabledKeys)
	case selectFocusLast:
		return listBoxLastEnabled(items, s.disabledKeys)
	default:
		selected := s.selectedKeys
		if s.selectionMode == SelectSelectionSingle {
			selected = []string{s.selectedKey}
		}
		for _, key := range selected {
			if index := listBoxIndexByKey(items, key); index >= 0 && !listBoxItemDisabled(items[index], s.disabledKeys) {
				return index, true
			}
		}
		return listBoxFirstEnabled(items, s.disabledKeys)
	}
}

func (s SelectWidget) layoutDismissAreas(ctx *Context, gtx layout.Context, state *selectState, excluded image.Rectangle) {
	viewport := ctx.overlayViewport(gtx.Constraints.Max)
	if viewport.X <= 0 || viewport.Y <= 0 {
		return
	}
	bounds := image.Rect(-viewport.X, -viewport.Y, viewport.X, viewport.Y)
	areas := overlayDismissRects(bounds, excluded)
	for i, area := range areas {
		if area.Empty() {
			continue
		}
		areaGtx := gtx
		areaGtx.Constraints = layout.Exact(area.Size())
		offset := op.Offset(area.Min).Push(gtx.Ops)
		pass := pointer.PassOp{}.Push(gtx.Ops)
		state.dismiss[i].Layout(areaGtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: area.Size()}
		})
		pass.Pop()
		offset.Pop()
	}
}

func (s SelectWidget) layoutDialogBlocker(gtx layout.Context, state *selectState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}
