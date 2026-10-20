package toast

import (
	"image"
	"sort"

	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type toastRecord struct {
	entry    *toastEntryState
	call     op.CallOp
	dims     layout.Dimensions
	progress float32
	index    int
	stack    float32
	exiting  bool
}

func (p Widget) layoutOverlay(ctx *core.Context, gtx layout.Context, providerState *toastProviderState, baseVisual theme.ToastVisual, override *theme.ToastOverride) layout.Dimensions {
	viewport := gtx.Constraints.Max
	if viewport.X <= 0 || viewport.Y <= 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints = layout.Exact(viewport)

	p.handleEvents(gtx, providerState)
	providerState.updateRegionEvents(gtx)
	expanded := providerState.regionHovered || providerState.paused(gtx)
	motion := baseVisual.Motion
	expansion := providerState.expansionProgress(gtx, expanded, baseVisual.Metrics.AnimationDuration, motion)
	providerState.updateTimers(gtx, p.paused || expanded || expansion > 0, func(entry *toastEntryState) {
		entry.requestClose(p.onClose)
	})

	maxVisible := p.resolvedMaxVisible(baseVisual)
	inset := gtx.Dp(baseVisual.Metrics.Inset)
	offset := gtx.Dp(p.resolvedOffset(baseVisual))
	maxWidth := min(gtx.Dp(p.resolvedWidth(baseVisual)), max(viewport.X-2*inset, 0))
	if maxWidth <= 0 {
		return layout.Dimensions{Size: viewport}
	}
	mobile := viewport.X <= gtx.Dp(768)
	records := make([]toastRecord, 0, maxVisible+1)
	presentIndex := 0
	frontHeight := 0
	var variantVisuals [5]theme.ToastVisual
	var variantVisualReady [5]bool
	restoreProvider := core.PushKey(ctx, p.key)
	defer restoreProvider()
	for _, key := range providerState.order {
		entry := providerState.entry(key)
		if entry == nil {
			continue
		}
		exiting := !entry.present || entry.closeRequested
		targetIndex := entry.stack.Target()
		if !exiting {
			targetIndex = float32(presentIndex)
			presentIndex++
		} else if !entry.stack.Ready() {
			targetIndex = float32(presentIndex)
		}
		stackPosition := entry.stackPosition(gtx, targetIndex, baseVisual.Metrics.AnimationDuration, motion)
		progress := entry.progress(gtx, baseVisual.Metrics.AnimationDuration, motion)
		if progress <= 0 && (!entry.present || entry.closeRequested) {
			continue
		}
		if targetIndex >= float32(maxVisible) && stackPosition >= float32(maxVisible) {
			continue
		}

		toastGtx := gtx
		toastGtx.Constraints.Min = image.Pt(0, 0)
		toastGtx.Constraints.Max = image.Pt(maxWidth, max(viewport.Y-inset-offset, 0))
		if !entry.item.compact {
			toastGtx.Constraints.Min.X = maxWidth
		}
		expandedLayout := expanded || expansion > 0
		interactive := !exiting && gtx.Enabled() && (targetIndex == 0 || expanded)
		if !expandedLayout && targetIndex != 0 && frontHeight > 0 {
			height := min(frontHeight, toastGtx.Constraints.Max.Y)
			toastGtx.Constraints.Min.Y = height
			toastGtx.Constraints.Max.Y = height
		}
		restoreItem := core.PushKey(ctx, entry.item.key)
		visual := baseVisual
		if entry.item.variant != Default {
			variantIndex := int(entry.item.variant)
			if variantIndex >= 0 && variantIndex < len(variantVisuals) {
				if !variantVisualReady[variantIndex] {
					variantVisuals[variantIndex] = p.visualWithCache(ctx, entry.item.variant, override)
					variantVisualReady[variantIndex] = true
				}
				visual = variantVisuals[variantIndex]
			} else {
				visual = p.visualWithCache(ctx, entry.item.variant, override)
			}
		}
		state := style.State{
			Hovered:      entry.hovered,
			Focused:      toastGtx.Focused(&entry.action) || toastGtx.Focused(&entry.close),
			FocusVisible: core.FocusVisible(ctx, &entry.root, toastGtx.Focused(&entry.root)),
			Disabled:     !interactive,
			Expanded:     expanded,
			Loading:      entry.item.loading,
		}
		resolved := computed.Resolve(ctx, toastGtx, core.FullKey(ctx, ""), state,
			visual.Defaults, style.Style{}, style.Style{}, visual.Override)
		macro := op.Record(gtx.Ops)
		dims := p.layoutToast(ctx, toastGtx, entry, interactive, mobile, expandedLayout || providerState.touchMode, visual, resolved)
		call := macro.Stop()
		restoreItem()
		if targetIndex == 0 && !exiting {
			frontHeight = dims.Size.Y
		}
		records = append(records, toastRecord{
			entry:    entry,
			call:     call,
			dims:     dims,
			progress: progress,
			index:    int(targetIndex),
			stack:    stackPosition,
			exiting:  exiting,
		})
	}

	bounds := p.paintRecords(ctx, gtx, viewport, records, expansion, baseVisual)
	providerState.addRegionInput(gtx, bounds)
	return layout.Dimensions{Size: viewport}
}

func (p Widget) handleEvents(gtx layout.Context, providerState *toastProviderState) {
	for _, keyValue := range providerState.order {
		entry := providerState.entry(keyValue)
		if entry == nil {
			continue
		}
		entry.updateRootEvents(gtx)
		for {
			_, ok := gtx.Event(key.FocusFilter{Target: &entry.root})
			if !ok {
				break
			}
		}
		for entry.close.Clicked(gtx) {
			entry.requestClose(p.onClose)
		}
		for entry.action.Clicked(gtx) {
			if p.onAction != nil {
				p.onAction(entry.item.key)
			}
		}
		for {
			event, ok := gtx.Event(key.Filter{Focus: &entry.root, Name: key.NameEscape})
			if !ok {
				break
			}
			if event, ok := event.(key.Event); ok && event.State == key.Press {
				entry.requestClose(p.onClose)
			}
		}
	}
}

func (p Widget) paintRecords(ctx *core.Context, gtx layout.Context, viewport image.Point, records []toastRecord, expansion float32, baseVisual theme.ToastVisual) image.Rectangle {
	if len(records) == 0 {
		return image.Rectangle{}
	}
	motion := baseVisual.Motion
	inset := gtx.Dp(baseVisual.Metrics.Inset)
	offset := gtx.Dp(p.resolvedOffset(baseVisual))
	gap := gtx.Dp(p.resolvedGap(baseVisual))
	scaleFactor := p.resolvedScaleFactor(baseVisual)
	frontHeight := toastFrontHeight(records)
	expandedOffsets := toastExpandedOffsets(records, gap)
	paintOrder := append([]toastRecord(nil), records...)
	sort.SliceStable(paintOrder, func(i, j int) bool {
		if paintOrder[i].stack != paintOrder[j].stack {
			return paintOrder[i].stack > paintOrder[j].stack
		}
		return !paintOrder[i].exiting && paintOrder[j].exiting
	})

	var bounds image.Rectangle
	for _, record := range paintOrder {
		height := record.dims.Size.Y
		if frontHeight > 0 {
			height = int(animation.LerpFloat(float32(frontHeight), float32(height), expansion) + 0.5)
		}
		visibleSize := image.Pt(record.dims.Size.X, height)
		collapsedScale := max(1-record.stack*scaleFactor, 0.5)
		scale := animation.LerpFloat(collapsedScale, 1, expansion)
		x := toastRegionX(viewport.X, record.dims.Size.X, inset, p.placement)
		collapsedOffset := record.stack * float32(gap)
		expandedTarget := toastExpandedOffset(expandedOffsets, float32(record.index))
		expandedOffset := record.entry.expandedPosition(gtx, expandedTarget, baseVisual.Metrics.AnimationDuration, motion)
		stackOffset := int(animation.LerpFloat(collapsedOffset, expandedOffset, expansion) + 0.5)
		y := offset + stackOffset
		enterOffset := -int(float32(height) * (1 - record.progress))
		if p.bottomPlacement() {
			y = viewport.Y - offset - height - stackOffset
			enterOffset = -enterOffset
		}
		y += enterOffset
		rect := image.Rectangle{Min: image.Pt(x, y), Max: image.Pt(x, y).Add(visibleSize)}
		if bounds.Empty() {
			bounds = rect
		} else {
			bounds = bounds.Union(rect)
		}

		offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
		transform := draw.Scale(visibleSize, scale).Push(gtx.Ops)
		opacity := paint.PushOpacity(gtx.Ops, record.progress)
		record.call.Add(gtx.Ops)
		opacity.Pop()
		transform.Pop()
		offset.Pop()
	}
	return bounds
}

func toastFrontHeight(records []toastRecord) int {
	for _, record := range records {
		if record.index == 0 && !record.exiting {
			return record.dims.Size.Y
		}
	}
	return 0
}

func toastExpandedOffsets(records []toastRecord, gap int) []float32 {
	maxIndex := 0
	for _, record := range records {
		maxIndex = max(maxIndex, record.index)
	}
	heights := make([]int, maxIndex+1)
	for _, record := range records {
		if record.index < 0 || record.index >= len(heights) {
			continue
		}
		if heights[record.index] == 0 || !record.exiting {
			heights[record.index] = record.dims.Size.Y
		}
	}
	offsets := make([]float32, len(heights))
	for index := 1; index < len(offsets); index++ {
		offsets[index] = offsets[index-1] + float32(heights[index-1]+gap)
	}
	return offsets
}

func toastExpandedOffset(offsets []float32, stack float32) float32 {
	if len(offsets) == 0 || stack <= 0 {
		return 0
	}
	last := len(offsets) - 1
	if stack >= float32(last) {
		return offsets[last]
	}
	lower := int(stack)
	progress := stack - float32(lower)
	return animation.LerpFloat(offsets[lower], offsets[lower+1], progress)
}

func (p Widget) layoutToast(ctx *core.Context, gtx layout.Context, entry *toastEntryState, interactive, mobile, expanded bool, visual theme.ToastVisual, resolved style.ResolvedStyle) layout.Dimensions {
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if !interactive {
			gtx = gtx.Disabled()
		}
		if entry.item.title != "" {
			semantic.LabelOp(entry.item.title).Add(gtx.Ops)
		}
		if entry.item.description != "" {
			semantic.DescriptionOp(entry.item.description).Add(gtx.Ops)
		}
		return p.layoutToastContent(ctx, gtx, entry, interactive, mobile, visual)
	})
	macro := op.Record(gtx.Ops)
	dims := layoutui.Resolved(ctx, gtx, resolved, content)
	call := macro.Stop()
	call.Add(gtx.Ops)
	if !interactive {
		entry.hovered = false
		return dims
	}
	entry.addRootInput(gtx, dims.Size)
	if !entry.item.compact {
		p.layoutToastClose(ctx, gtx, entry, dims.Size, mobile || expanded, visual)
	}
	return dims
}

func (p Widget) layoutToastContent(ctx *core.Context, gtx layout.Context, entry *toastEntryState, interactive, mobile bool, visual theme.ToastVisual) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	if entry.item.showIndicator() {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutToastIndicator(ctx, gtx, entry.item, visual)
		}))
	}
	if entry.item.compact {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutToastText(ctx, gtx, entry, false, visual)
		}))
	} else {
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return p.layoutToastText(ctx, gtx, entry, interactive && mobile, visual)
		}))
		if interactive && !mobile && entry.item.actionLabel != "" {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutToastAction(ctx, gtx, entry, visual)
			}))
		}
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(visual.Metrics.ContentGap),
	}.Layout(gtx, children...)
}

func (p Widget) layoutToastText(ctx *core.Context, gtx layout.Context, entry *toastEntryState, mobileAction bool, visual theme.ToastVisual) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	if entry.item.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return visual.Text(gtx, theme.ToastTextRenderRequest{Text: entry.item.title, Role: theme.ToastTextTitle})
		}))
	}
	if entry.item.description != "" && !entry.item.compact {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return visual.Text(gtx, theme.ToastTextRenderRequest{Text: entry.item.description, Role: theme.ToastTextDescription})
		}))
	}
	if mobileAction && entry.item.actionLabel != "" && !entry.item.compact {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: visual.Metrics.ActionInsetTop}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutToastAction(ctx, gtx, entry, visual)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (p Widget) layoutToastIndicator(ctx *core.Context, gtx layout.Context, item Item, visual theme.ToastVisual) layout.Dimensions {
	if visual.Indicator == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	var custom func(layout.Context) layout.Dimensions
	if item.hasIndicator && item.indicator != nil {
		custom = func(gtx layout.Context) layout.Dimensions {
			if !gtx.Enabled() {
				gtx = gtx.Disabled()
			}
			return item.indicator.Layout(ctx, gtx)
		}
	}
	return visual.Indicator(gtx, theme.ToastIndicatorRenderRequest{
		Size: visual.Metrics.IndicatorSize, Loading: item.loading, Custom: custom,
	})
}

func (p Widget) layoutToastAction(ctx *core.Context, gtx layout.Context, entry *toastEntryState, visual theme.ToastVisual) layout.Dimensions {
	if visual.Action == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	gtx.Constraints.Min.X = 0
	return visual.Action(gtx, theme.ToastActionRenderRequest{
		Label: entry.item.actionLabel, Variant: theme.ButtonVariant(entry.item.actionVariant),
		Clickable: &entry.action, State: style.State{Hovered: entry.action.Hovered(), Pressed: entry.action.Pressed(), Focused: gtx.Focused(&entry.action)},
	})
}

func (p Widget) layoutToastClose(ctx *core.Context, gtx layout.Context, entry *toastEntryState, toastSize image.Point, alwaysVisible bool, visual theme.ToastVisual) {
	show := alwaysVisible || entry.hovered || entry.close.Hovered() || gtx.Focused(&entry.root) || gtx.Focused(&entry.close)
	if !show {
		return
	}
	size := gtx.Dp(visual.Metrics.CloseButtonSize)
	inset := gtx.Dp(visual.Metrics.CloseInset)
	position := image.Pt(toastSize.X-size-inset, inset)
	closeGtx := gtx
	closeGtx.Constraints = layout.Exact(image.Pt(size, size))
	offset := op.Offset(position).Push(gtx.Ops)
	if visual.Close == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	visual.Close(closeGtx, theme.ToastCloseRenderRequest{
		Label: ctx.Text("Close"), Clickable: &entry.close, Size: visual.Metrics.CloseButtonSize,
		State: style.State{Hovered: entry.close.Hovered(), Pressed: entry.close.Pressed(), Focused: gtx.Focused(&entry.close)},
	})
	offset.Pop()
}

func (p Widget) resolvedOffset(visual theme.ToastVisual) unit.Dp {
	if p.hasOffset {
		return p.offset
	}
	return visual.Metrics.Inset
}

func (p Widget) resolvedGap(visual theme.ToastVisual) unit.Dp {
	if p.hasGap {
		return p.gap
	}
	return visual.Metrics.Gap
}

func (p Widget) resolvedMaxVisible(visual theme.ToastVisual) int {
	if p.hasMaxVisible {
		return p.maxVisible
	}
	return max(visual.Metrics.MaxVisible, 1)
}

func (p Widget) resolvedScaleFactor(visual theme.ToastVisual) float32 {
	if p.hasScale {
		return p.scaleFactor
	}
	return min(max(visual.Metrics.ScaleFactor, 0), 1)
}

func (p Widget) resolvedWidth(visual theme.ToastVisual) unit.Dp {
	if p.hasWidth {
		return p.width
	}
	return visual.Metrics.Width
}

func (p Widget) bottomPlacement() bool {
	return p.placement == Bottom || p.placement == BottomStart || p.placement == BottomEnd
}

func toastRegionX(viewportWidth, toastWidth, inset int, placement Placement) int {
	switch placement {
	case BottomStart, TopStart:
		return inset
	case BottomEnd, TopEnd:
		return max(viewportWidth-inset-toastWidth, 0)
	default:
		return max((viewportWidth-toastWidth)/2, 0)
	}
}
