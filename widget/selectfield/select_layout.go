package selectfield

import (
	"image"
	"math"
	"strings"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/listbox"
)

func (s Widget) layout(ctx *core.Context, gtx layout.Context, state *selectState, open bool, overrides ...style.Style) layout.Dimensions {
	instanceStyle := firstSelectStyle(overrides)
	bundle := selectStyleDeclarationsCached(core.ActiveTheme(ctx), s.variant, &state.styleCache)
	return s.layoutWithBundle(ctx, gtx, state, open, instanceStyle, bundle)
}

func (s Widget) layoutWithBundle(
	ctx *core.Context,
	gtx layout.Context,
	state *selectState,
	open bool,
	instanceStyle style.Style,
	bundle selectDeclarationBundle,
) layout.Dimensions {
	resolved := s.resolveTriggerStyle(ctx, gtx, state, open, instanceStyle, bundle)
	return s.layoutResolved(ctx, gtx, state, open, resolved)
}

func (s Widget) layoutResolved(ctx *core.Context, gtx layout.Context, state *selectState, open bool, resolved selectResolvedStyle) layout.Dimensions {
	gap := gtx.Dp(resolved.visual.Metrics.ContentGap)
	var labelDims, triggerDims layout.Dimensions
	children := make([]layout.FlexChild, 0, 3)
	if s.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			labelDims = s.layoutLabel(ctx, gtx)
			return labelDims
		}))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		triggerDims = s.layoutTrigger(ctx, gtx, state, open, resolved)
		return triggerDims
	}))
	if message, _ := s.supportMessage(); message != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return resolved.visual.Text(gtx, theme.SelectTextRequest{
				Text: message, Style: resolved.description,
			})
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

func (s Widget) layoutLabel(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return label.New(s.label).
		For(s.key).
		Required(s.required).
		Disabled(s.disabled).
		Invalid(s.invalid).
		Layout(ctx, gtx)
}

func (s Widget) supportMessage() (string, bool) {
	if s.invalid {
		if s.errorMessage == "" {
			return "", false
		}
		return s.errorMessage, true
	}
	return s.description, false
}

func (s Widget) layoutTrigger(ctx *core.Context, gtx layout.Context, state *selectState, open bool, resolved selectResolvedStyle) layout.Dimensions {
	animGtx := gtx
	key := s.resolvedKey(ctx, state)
	if s.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	content := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		value, selected := s.displayValueCached(state)
		if !selected && !s.hasPlaceholder {
			value = s.resolvedPlaceholder(ctx)
		}

		semantic.Button.Add(gtx.Ops)
		label := s.label
		if label == "" {
			label = core.FieldLabel(ctx, key)
		}
		semantic.LabelOp(selectSemanticLabel(label, value)).Add(gtx.Ops)
		description := s.description
		if description == "" {
			description = core.FieldDescription(ctx, key)
		}
		if s.invalid {
			description = s.errorMessage
		}
		if description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)

		indicatorWidth := gtx.Dp(resolved.visual.IndicatorWidth(resolved.indicator))
		maxX := max(gtx.Constraints.Max.X-indicatorWidth, 0)
		minX := min(max(gtx.Constraints.Min.X-indicatorWidth, 0), maxX)
		valueGtx := gtx
		valueGtx.Constraints = layout.Constraints{Min: image.Pt(minX, 0), Max: image.Pt(maxX, gtx.Constraints.Max.Y)}

		macro := op.Record(gtx.Ops)
		valueDims := layoutSelectValue(ctx, valueGtx, value, selected, resolved)
		valueCall := macro.Stop()

		size := image.Pt(valueDims.Size.X+indicatorWidth, valueDims.Size.Y)
		size = gtx.Constraints.Constrain(size)
		valueOffset := op.Offset(image.Pt(0, max((size.Y-valueDims.Size.Y)/2, 0))).Push(gtx.Ops)
		valueCall.Add(gtx.Ops)
		valueOffset.Pop()

		indicatorSize := image.Pt(min(indicatorWidth, size.X), size.Y)
		indicatorOffset := op.Offset(image.Pt(max(size.X-indicatorSize.X, 0), 0)).Push(gtx.Ops)
		s.layoutIndicator(ctx, animGtx, indicatorSize, state.iconProgress(animGtx, open, resolved.visual.Metrics, resolved.visual.Motion), resolved)
		indicatorOffset.Pop()
		return layout.Dimensions{Size: size}
	})
	return layoutui.InteractiveResolved(ctx, gtx, resolved.content, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		macro := op.Record(gtx.Ops)
		dims := visual(gtx)
		call := macro.Stop()
		state.trigger.Layout(gtx, func(layout.Context) layout.Dimensions { return dims })
		call.Add(gtx.Ops)
		return dims
	})
}

func (s Widget) layoutIndicator(ctx *core.Context, gtx layout.Context, size image.Point, progress float32, resolved selectResolvedStyle) layout.Dimensions {
	if s.indicator == nil {
		if resolved.visual.Indicator == nil {
			return layout.Dimensions{Size: size}
		}
		return resolved.visual.Indicator(gtx, size, progress, resolved.indicator)
	}
	gtx.Constraints = layout.Exact(size)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	transform := op.Affine(f32.AffineId().Rotate(center, progress*float32(math.Pi))).Push(gtx.Ops)
	indicator := s.indicator
	dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		iconSize := minPoint(resolved.visual.IconSize(gtx, resolved.icon), size)
		gtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
		return indicator.Layout(ctx, gtx)
	})
	transform.Pop()
	return dims
}

func layoutSelectValue(ctx *core.Context, gtx layout.Context, value string, selected bool, resolved selectResolvedStyle) layout.Dimensions {
	part := resolved.placeholder
	if selected {
		part = resolved.content
	}
	return resolved.visual.Text(gtx, theme.SelectTextRequest{Text: value, Style: part})
}

func minPoint(value, limit image.Point) int {
	return min(min(value.X, value.Y), min(limit.X, limit.Y))
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

func (s Widget) layoutPopoverWithBundle(ctx *core.Context, state *selectState, triggerRect image.Rectangle, open bool, progress float32, disabled bool, bundle selectDeclarationBundle) {
	if triggerRect.Empty() {
		return
	}
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       s.resolvedKey(ctx, state),
		Layer:     core.OverlayLayerPopup,
		Anchor:    triggerRect,
		HasAnchor: true,
		Disabled:  disabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			overlayOpen := open
			if interactive && gtx.Enabled() {
				overlayOpen = state.handleOverlayEvents(ctx, gtx, s, overlayOpen)
			}
			dims := s.layoutOverlayWithBundle(ctx, gtx, state, anchor, overlayOpen, progress, interactive && gtx.Enabled(), bundle)
			if overlayOpen && interactive && gtx.Enabled() {
				core.AfterOverlays(ctx, func() {
					if core.OverlayTopmost(ctx, core.OverlayLayerPopup, state.key) {
						s.focusPendingOption(ctx, state)
					}
				})
			}
			return dims
		},
	})
}

func (s Widget) layoutOverlay(ctx *core.Context, gtx layout.Context, state *selectState, triggerRect image.Rectangle, open bool, progress float32, interactive bool, overrides ...style.Style) layout.Dimensions {
	bundle := selectStyleDeclarationsCached(core.ActiveTheme(ctx), s.variant, &state.styleCache)
	return s.layoutOverlayWithBundle(ctx, gtx, state, triggerRect, open, progress, interactive, bundle)
}

func (s Widget) layoutOverlayWithBundle(ctx *core.Context, gtx layout.Context, state *selectState, triggerRect image.Rectangle, open bool, progress float32, interactive bool, bundle selectDeclarationBundle) layout.Dimensions {
	visual := bundle.visual
	metrics := visual.Metrics
	bounds := gtx.Constraints.Max
	trigger := triggerRect.Size()
	width := min(trigger.X, bounds.X)
	maxHeight := min(gtx.Dp(metrics.PanelMaxHeight), bounds.Y)
	if width <= 0 || maxHeight <= 0 {
		return layout.Dimensions{Size: bounds}
	}
	panelGtx := gtx
	panelGtx.Constraints = layout.Constraints{Min: image.Pt(width, 0), Max: image.Pt(width, maxHeight)}
	macro := op.Record(gtx.Ops)
	panelDims, panelPlacement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return s.layoutPanelWithBundle(ctx, panelGtx, state, open, interactive, bundle.visual.Override, bundle)
	})
	panelCall := macro.Stop()

	gap := gtx.Dp(metrics.PanelGap)
	result := core.ResolvePosition(core.PositionConfig{
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
	origin := core.PanelTransformOriginAt(triggerRect, result.Position, panelDims.Size, result.Placement)
	scale := metrics.AnimationScale + (1-metrics.AnimationScale)*progress
	slide := core.SlideOffset(gtx.Dp(metrics.AnimationDistance), progress, result.Placement)
	panelOffset := result.Position.Add(slide)
	panelScale := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))).Mul(panelScale)
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	animatedPanel := core.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform)
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

func (s Widget) layoutPanelWithOverride(ctx *core.Context, gtx layout.Context, state *selectState, open, interactive bool, overrides ...*theme.SelectOverride) layout.Dimensions {
	activeTheme := core.ActiveTheme(ctx)
	bundle := selectStyleDeclarationsCached(activeTheme, s.variant, &state.styleCache)
	if override := firstSelectOverride(overrides); override != nil {
		bundle = selectVisualBundle(activeTheme, s.variant, override)
	}
	return s.layoutPanelWithBundle(ctx, gtx, state, open, interactive, bundle.visual.Override, bundle)
}

func (s Widget) layoutPanelWithBundle(ctx *core.Context, gtx layout.Context, state *selectState, open, interactive bool, instanceStyle style.Style, declarations selectDeclarationBundle) layout.Dimensions {
	defaults := declarations.visual.Defaults
	part := computed.ResolvePart(
		ctx,
		gtx,
		s.resolvedKey(ctx, state),
		theme.SelectPartOverlay,
		style.State{Disabled: !interactive, Open: open},
		defaults,
		style.Style{},
		style.Style{},
		instanceStyle,
	)
	return layoutui.Resolved(ctx, gtx, part, s.listBoxWithStyle(ctx, state, open && interactive, instanceStyle))
}

func (s Widget) listBox(ctx *core.Context, state *selectState, open bool) core.Widget {
	return s.listBoxWithStyle(ctx, state, open, style.Style{})
}

func (s Widget) listBoxWithStyle(ctx *core.Context, state *selectState, open bool, instanceStyle style.Style) core.Widget {
	var list listbox.Widget
	if s.selectionMode == SelectSelectionMultiple {
		if len(s.sections) > 0 {
			list = listbox.MultipleSections(s.key, s.selectedKeys, s.sections)
		} else {
			list = listbox.New(s.items).Key(s.key).SelectedKeys(s.selectedKeys)
		}
		list = list.OnSelectionChange(func(keys []string) {
			state.requestSelectedKeys(s, keys)
		})
	} else {
		if len(s.sections) > 0 {
			list = listbox.Sections(s.key, s.selectedKey, s.sections)
		} else {
			list = listbox.New(s.items).Key(s.key).SelectedKey(s.selectedKey)
		}
		list = list.
			OnChange(func(key string) {
				state.requestSelectedKey(s, key)
			}).
			OnAction(func(string) {
				state.focusIntent = selectFocusNone
				state.requestOpen(ctx, s, false)
				core.RequestFocus(ctx, &state.trigger)
			})
	}
	if s.hasEmptyText {
		list = list.EmptyText(s.emptyText)
	}
	list = list.
		DisabledKeys(s.disabledKeys).
		Disabled(!open || s.disabled)
	if s.hasDataVersion {
		list = list.DataVersion(s.dataVersion)
	}
	list = listbox.WithDerivedIdentity(list, s.resolvedKey(ctx, state), "options")
	return list.StyleParts(instanceStyle)
}

func firstSelectStyle(overrides []style.Style) style.Style {
	if len(overrides) == 0 {
		return style.Style{}
	}
	return overrides[0]
}

func firstSelectOverride(overrides []*theme.SelectOverride) *theme.SelectOverride {
	if len(overrides) == 0 {
		return nil
	}
	return overrides[0]
}

func (s Widget) resolvedKey(ctx *core.Context, state *selectState) string {
	if state != nil && state.key != "" {
		return state.key
	}
	return core.FullKey(ctx, s.key)
}

func (s Widget) focusPendingOption(ctx *core.Context, state *selectState) {
	if state.focusIntent == selectFocusNone {
		return
	}
	items := state.itemsFor(s)
	index, ok := s.focusOptionIndex(items, state.focusIntent)
	focusVisible := state.focusVisibleIntent
	state.focusIntent = selectFocusNone
	state.focusVisibleIntent = false
	if !ok {
		return
	}
	listbox.FocusDerivedItem(ctx, s.resolvedKey(ctx, state), "options", items[index].Key, focusVisible)
}

func (s Widget) focusOptionIndex(items []SelectItem, intent selectFocusIntent) (int, bool) {
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

func (s Widget) layoutDismissAreas(gtx layout.Context, state *selectState, viewport image.Point, excluded ...image.Rectangle) {
	if viewport.X <= 0 || viewport.Y <= 0 {
		return
	}
	bounds := image.Rectangle{Max: viewport}
	areas := core.DismissRectsExcluding(bounds, excluded...)
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

func (s Widget) layoutDialogBlocker(gtx layout.Context, state *selectState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}
