package selects

import (
	"image"
	"image/color"
	"math"
	"strings"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/components/label"
	"github.com/qianniancn/FlowUI/internal/components/listbox"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

func (s SelectWidget) layout(ctx *frame.Context, gtx layout.Context, state *selectState, open bool) layout.Dimensions {
	style := selectStyleFor(frame.ActiveTheme(ctx), s.variant, state.trigger.Hovered(), false, s.disabled, s.invalid)
	gap := gtx.Dp(frame.ActiveTheme(ctx).Components.Select.ContentGap)
	var labelDims, triggerDims layout.Dimensions
	children := make([]layout.FlexChild, 0, 3)
	if s.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			labelDims = s.layoutLabel(ctx, gtx)
			return labelDims
		}))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		triggerDims = s.layoutTrigger(ctx, gtx, state, open)
		return triggerDims
	}))
	if message, isError := s.supportMessage(); message != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !isError {
				return description.Description(message).
					For(s.key).
					Disabled(s.disabled).
					Layout(ctx, gtx)
			}
			return text.New(message).
				Size(float32(frame.ActiveTheme(ctx).Components.Description.TextSize)).
				Color(style.error).
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

func (s SelectWidget) layoutLabel(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return label.Label(s.label).
		For(s.key).
		Required(s.required).
		Disabled(s.disabled).
		Invalid(s.invalid).
		Layout(ctx, gtx)
}

func (s SelectWidget) supportMessage() (string, bool) {
	if s.invalid {
		if s.errorMessage == "" {
			return "", false
		}
		return s.errorMessage, true
	}
	return s.description, false
}

func (s SelectWidget) layoutTrigger(ctx *frame.Context, gtx layout.Context, state *selectState, open bool) layout.Dimensions {
	animGtx := gtx
	if s.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	theme := frame.ActiveTheme(ctx).Components.Select
	height := min(gtx.Dp(theme.Height), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)

	return state.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		value, selected := s.displayValue()
		focusVisible := frame.FocusVisible(ctx, &state.trigger, gtx.Focused(&state.trigger))
		style := selectStyleFor(frame.ActiveTheme(ctx), s.variant, state.trigger.Hovered(), focusVisible, s.disabled, s.invalid)
		style.field.Background = state.field.Background(animGtx, style.field.Background)
		style.field.Border = state.field.BorderColor(animGtx, style.field.Border)

		semantic.Button.Add(gtx.Ops)
		label := s.label
		if label == "" {
			label = frame.FieldLabel(ctx, state.key)
		}
		semantic.LabelOp(selectSemanticLabel(label, value)).Add(gtx.Ops)
		description := s.description
		if description == "" {
			description = frame.FieldDescription(ctx, state.key)
		}
		if s.invalid {
			description = s.errorMessage
		}
		if description != "" {
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
		s.layoutIndicator(ctx, animGtx, indicatorSize, state.iconProgress(animGtx, open), style.field.Placeholder)
		indicatorOffset.Pop()
		return layout.Dimensions{Size: size}
	})
}

func (s SelectWidget) layoutIndicator(ctx *frame.Context, gtx layout.Context, size image.Point, progress float32, color color.NRGBA) layout.Dimensions {
	if s.indicator == nil {
		drawSelectIndicator(gtx, frame.ActiveTheme(ctx), size, progress, color)
		return layout.Dimensions{Size: size}
	}
	gtx.Constraints = layout.Exact(size)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	transform := op.Affine(f32.AffineId().Rotate(center, progress*float32(math.Pi))).Push(gtx.Ops)
	indicator := s.indicator
	if text, ok := indicator.(text.Widget); ok {
		indicator = text.DefaultColor(color)
	}
	dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		iconSize := min(gtx.Dp(frame.ActiveTheme(ctx).Components.Select.IndicatorSize), min(size.X, size.Y))
		gtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
		return indicator.Layout(ctx, gtx)
	})
	transform.Pop()
	return dims
}

func layoutSelectValue(ctx *frame.Context, gtx layout.Context, value string, selected bool, style selectStyle) layout.Dimensions {
	label := material.Label(frame.ActiveTheme(ctx).Material, frame.ActiveTheme(ctx).Components.Select.TextSize, value)
	label.Color = style.field.Placeholder
	if selected {
		label.Color = style.field.Foreground
	}
	return label.Layout(gtx)
}

func selectSemanticLabel(label, value string) string {
	parts := make([]string, 0, 2)
	if label != "" {
		parts = append(parts, label)
	}
	if value != "" {
		parts = append(parts, value)
	}
	return strings.Join(parts, ", ")
}

func (s SelectWidget) layoutPopover(ctx *frame.Context, state *selectState, triggerRect image.Rectangle, open bool, progress float32, disabled bool) {
	if triggerRect.Empty() {
		return
	}
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       s.resolvedKey(ctx, state),
		Layer:     frame.OverlayLayerPopup,
		Anchor:    triggerRect,
		HasAnchor: true,
		Disabled:  disabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			overlayOpen := open
			if interactive && gtx.Enabled() {
				overlayOpen = state.handleOverlayEvents(ctx, gtx, s, overlayOpen)
			}
			dims := s.layoutOverlay(ctx, gtx, state, anchor, overlayOpen, progress, interactive && gtx.Enabled())
			if overlayOpen && interactive && gtx.Enabled() {
				frame.AfterOverlays(ctx, func() {
					if frame.OverlayTopmost(ctx, frame.OverlayLayerPopup, state.key) {
						s.focusPendingOption(ctx, state)
					}
				})
			}
			return dims
		},
	})
}

func (s SelectWidget) layoutOverlay(ctx *frame.Context, gtx layout.Context, state *selectState, triggerRect image.Rectangle, open bool, progress float32, interactive bool) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Select
	bounds := gtx.Constraints.Max
	trigger := triggerRect.Size()
	width := min(trigger.X, bounds.X)
	maxHeight := min(gtx.Dp(theme.PanelMaxHeight), bounds.Y)
	if width <= 0 || maxHeight <= 0 {
		return layout.Dimensions{Size: bounds}
	}
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Min: image.Pt(width, 0), Max: image.Pt(width, maxHeight)}
	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return s.layoutPanel(ctx, panelGtx, state, open, interactive)
	})
	panelCall := macro.Stop()

	gap := gtx.Dp(theme.PanelGap)
	result := overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          trigger,
		TriggerOrigin:    triggerRect.Min,
		HasTriggerOrigin: true,
		Panel:            panelDims.Size,
		Bounds:           bounds,
		Offset:           gap,
		Placement:        s.placement.Placement(),
		Flip:             s.flipEnabled(),
		AvoidOverflow:    s.overflowAvoidanceEnabled(),
	})
	origin := overlay.PanelTransformOriginAt(triggerRect, result.Position, panelDims.Size, result.Placement)
	scale := theme.AnimationScale + (1-theme.AnimationScale)*progress
	slide := overlay.SlideOffset(gtx.Dp(theme.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := overlay.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
	s.layoutDismissAreas(gtx, state, bounds, triggerRect, animatedPanel)
	offset := op.Offset(panelOffset).Push(gtx.Ops)
	transform := op.Affine(panelScale).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	s.layoutDialogBlocker(gtx, state, panelDims.Size)
	panelCall.Add(gtx.Ops)
	opacity.Pop()
	transform.Pop()
	offset.Pop()
	return layout.Dimensions{Size: bounds}
}

func (s SelectWidget) layoutPanel(ctx *frame.Context, gtx layout.Context, state *selectState, open, interactive bool) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	var dims layout.Dimensions
	func() {
		restore := frame.PushColors(ctx, frame.ActiveTheme(ctx).Palette.OverlayForegroundColor(), frame.ActiveTheme(ctx).Palette.OverlayColor())
		defer restore()
		list := s.listBox(ctx, state, open && interactive)
		dims = list.Layout(ctx, gtx)
	}()
	call := macro.Stop()

	radius := min(max(gtx.Dp(frame.ActiveTheme(ctx).Components.Select.PanelRadius), 1), min(dims.Size.X, dims.Size.Y)/2)
	rect := image.Rectangle{Max: dims.Size}
	drawSelectPanel(gtx, frame.ActiveTheme(ctx), rect, radius)
	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (s SelectWidget) listBox(ctx *frame.Context, state *selectState, open bool) listbox.ListBoxWidget {
	var list listbox.ListBoxWidget
	if s.selectionMode == SelectSelectionMultiple {
		if len(s.sections) > 0 {
			list = listbox.ListBoxMultipleSections(s.key, s.selectedKeys, s.sections)
		} else {
			list = listbox.ListBoxMultiple(s.key, s.selectedKeys, s.items)
		}
		list = list.OnSelectionChange(s.onSelectionChange)
	} else {
		if len(s.sections) > 0 {
			list = listbox.ListBoxSections(s.key, s.selectedKey, s.sections)
		} else {
			list = listbox.ListBox(s.key, s.selectedKey, s.items)
		}
		list = list.
			OnChange(s.onChange).
			OnAction(func(string) {
				state.focusIntent = selectFocusNone
				state.requestOpen(ctx, s, false)
				frame.RequestFocus(ctx, &state.trigger)
			})
	}
	list = list.
		EmptyText(s.emptyText).
		DisabledKeys(s.disabledKeys).
		Disabled(!open || s.disabled)
	list = listbox.WithDerivedIdentity(list, s.resolvedKey(ctx, state), "options")
	return listbox.WithPadding(list, frame.ActiveTheme(ctx).Components.Select.PanelPadding)
}

func (s SelectWidget) resolvedKey(ctx *frame.Context, state *selectState) string {
	if state != nil && state.key != "" {
		return state.key
	}
	return frame.FullKey(ctx, s.key)
}

func (s SelectWidget) focusPendingOption(ctx *frame.Context, state *selectState) {
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
	listbox.FocusDerivedItem(ctx, s.resolvedKey(ctx, state), "options", items[index].Key, focusVisible)
}

func (s SelectWidget) focusOptionIndex(items []SelectItem, intent selectFocusIntent) (int, bool) {
	switch intent {
	case selectFocusFirst:
		return listbox.FirstEnabled(items, s.disabledKeys)
	case selectFocusLast:
		return listbox.LastEnabled(items, s.disabledKeys)
	default:
		selected := s.selectedKeys
		if s.selectionMode == SelectSelectionSingle {
			selected = []string{s.selectedKey}
		}
		for _, key := range selected {
			if index := listbox.IndexByKey(items, key); index >= 0 && !listbox.ItemDisabled(items[index], s.disabledKeys) {
				return index, true
			}
		}
		return listbox.FirstEnabled(items, s.disabledKeys)
	}
}

func (s SelectWidget) layoutDismissAreas(gtx layout.Context, state *selectState, viewport image.Point, excluded ...image.Rectangle) {
	if viewport.X <= 0 || viewport.Y <= 0 {
		return
	}
	bounds := image.Rectangle{Max: viewport}
	areas := overlay.DismissRectsExcluding(bounds, excluded...)
	for i, area := range areas {
		if i >= len(state.dismiss) {
			break
		}
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
