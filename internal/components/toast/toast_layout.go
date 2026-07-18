package toast

import (
	"image"
	"sort"

	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/components/spinner"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/flowui-icons-lucide"
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

func (p ToastProviderWidget) layoutOverlay(ctx *frame.Context, gtx layout.Context, providerState *toastProviderState) layout.Dimensions {
	viewport := gtx.Constraints.Max
	if viewport.X <= 0 || viewport.Y <= 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints = layout.Exact(viewport)

	p.handleEvents(gtx, providerState)
	providerState.updateRegionEvents(gtx)
	expanded := providerState.regionHovered || providerState.paused(gtx)
	expansion := providerState.expansionProgress(gtx, expanded, frame.ActiveTheme(ctx).Components.Toast.AnimationDuration)
	providerState.updateTimers(gtx, p.paused || expanded || expansion > 0, func(entry *toastEntryState) {
		entry.requestClose(p.onClose)
	})

	tokens := frame.ActiveTheme(ctx).Components.Toast
	maxVisible := p.resolvedMaxVisible(ctx)
	inset := gtx.Dp(tokens.Inset)
	offset := gtx.Dp(p.resolvedOffset(ctx))
	width := min(gtx.Dp(p.resolvedWidth(ctx)), max(viewport.X-2*inset, 0))
	if width <= 0 {
		return layout.Dimensions{Size: viewport}
	}
	mobile := viewport.X <= gtx.Dp(768)
	records := make([]toastRecord, 0, maxVisible+1)
	presentIndex := 0
	frontHeight := 0
	restoreProvider := frame.PushKey(ctx, p.key)
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
		stackPosition := entry.stackPosition(gtx, targetIndex, tokens.AnimationDuration)
		progress := entry.progress(gtx, tokens.AnimationDuration)
		if progress <= 0 && (!entry.present || entry.closeRequested) {
			continue
		}
		if targetIndex >= float32(maxVisible) && stackPosition >= float32(maxVisible) {
			continue
		}

		toastGtx := gtx
		toastGtx.Constraints.Min = image.Pt(width, 0)
		toastGtx.Constraints.Max = image.Pt(width, max(viewport.Y-inset-offset, 0))
		expandedLayout := expanded || expansion > 0
		interactive := !exiting && gtx.Enabled() && (targetIndex == 0 || expanded)
		if !expandedLayout && targetIndex != 0 && frontHeight > 0 {
			height := min(frontHeight, toastGtx.Constraints.Max.Y)
			toastGtx.Constraints.Min.Y = height
			toastGtx.Constraints.Max.Y = height
		}
		restoreItem := frame.PushKey(ctx, entry.item.key)
		macro := op.Record(gtx.Ops)
		dims := p.layoutToast(ctx, toastGtx, entry, interactive, mobile, expandedLayout || providerState.touchMode)
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

	bounds := p.paintRecords(ctx, gtx, viewport, records, expansion)
	providerState.addRegionInput(gtx, bounds)
	return layout.Dimensions{Size: viewport}
}

func (p ToastProviderWidget) handleEvents(gtx layout.Context, providerState *toastProviderState) {
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

func (p ToastProviderWidget) paintRecords(ctx *frame.Context, gtx layout.Context, viewport image.Point, records []toastRecord, expansion float32) image.Rectangle {
	if len(records) == 0 {
		return image.Rectangle{}
	}
	tokens := frame.ActiveTheme(ctx).Components.Toast
	inset := gtx.Dp(tokens.Inset)
	offset := gtx.Dp(p.resolvedOffset(ctx))
	gap := gtx.Dp(p.resolvedGap(ctx))
	scaleFactor := p.resolvedScaleFactor(ctx)
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
			height = int(render.Lerp(float32(frontHeight), float32(height), expansion) + 0.5)
		}
		visibleSize := image.Pt(record.dims.Size.X, height)
		collapsedScale := max(1-record.stack*scaleFactor, 0.5)
		scale := render.Lerp(collapsedScale, 1, expansion)
		x := toastRegionX(viewport.X, record.dims.Size.X, inset, p.placement)
		collapsedOffset := record.stack * float32(gap)
		expandedOffset := toastExpandedOffset(expandedOffsets, record.stack)
		stackOffset := int(render.Lerp(collapsedOffset, expandedOffset, expansion) + 0.5)
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
		transform := render.Scale(visibleSize, scale).Push(gtx.Ops)
		opacity := paint.PushOpacity(gtx.Ops, record.progress)
		style := toastStyleFor(frame.ActiveTheme(ctx), record.entry.item.variant)
		visibleRect := image.Rectangle{Max: visibleSize}
		drawToastSurface(gtx, frame.ActiveTheme(ctx), visibleRect, toastRadius(gtx, tokens.Radius, visibleSize), style.surface)
		if record.index > 0 && expansion < 1 {
			visible := clip.Rect(visibleRect).Push(gtx.Ops)
			record.call.Add(gtx.Ops)
			visible.Pop()
		} else {
			record.call.Add(gtx.Ops)
		}
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
	return render.Lerp(offsets[lower], offsets[lower+1], progress)
}

func (p ToastProviderWidget) layoutToast(ctx *frame.Context, gtx layout.Context, entry *toastEntryState, interactive, mobile, expanded bool) layout.Dimensions {
	style := toastStyleFor(frame.ActiveTheme(ctx), entry.item.variant)
	contentGtx := gtx
	if !interactive {
		contentGtx = contentGtx.Disabled()
	}
	macro := op.Record(gtx.Ops)
	contentDims := p.layoutToastContent(ctx, contentGtx, entry, interactive, mobile, style)
	contentCall := macro.Stop()
	size := gtx.Constraints.Constrain(contentDims.Size)
	rect := image.Rectangle{Max: size}
	radius := toastRadius(gtx, frame.ActiveTheme(ctx).Components.Toast.Radius, size)

	if !interactive {
		clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
		contentCall.Add(gtx.Ops)
		clipStack.Pop()
		entry.hovered = false
		return layout.Dimensions{Size: size}
	}

	semanticClip := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	semantic.LabelOp(entry.item.title).Add(gtx.Ops)
	if entry.item.description != "" {
		semantic.DescriptionOp(entry.item.description).Add(gtx.Ops)
	}
	contentCall.Add(gtx.Ops)
	entry.addRootInput(gtx, size)
	semanticClip.Pop()
	focusVisible := frame.FocusVisible(ctx, &entry.root, gtx.Focused(&entry.root))
	focusOpacity := entry.rootFocus.Opacity(gtx, focusVisible)
	drawToastFocus(gtx, rect, radius, style.focus, frame.ActiveTheme(ctx).Components.Toast.FocusRingWidth, focusOpacity)
	p.layoutToastClose(ctx, gtx, entry, size, style, mobile || expanded)
	return layout.Dimensions{Size: size}
}

func (p ToastProviderWidget) layoutToastContent(ctx *frame.Context, gtx layout.Context, entry *toastEntryState, interactive, mobile bool, style toastStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Toast
	return layout.Inset{
		Top: tokens.PaddingY, Bottom: tokens.PaddingY,
		Left: tokens.PaddingX, Right: tokens.PaddingX,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, 3)
		if entry.item.showIndicator() {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutToastIndicator(ctx, gtx, entry.item, style)
			}))
		}
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return p.layoutToastText(ctx, gtx, entry, interactive && mobile, style)
		}))
		if interactive && !mobile && entry.item.actionLabel != "" {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutToastAction(ctx, gtx, entry)
			}))
		}
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Start,
			Gap:       gtx.Dp(tokens.ContentGap),
		}.Layout(gtx, children...)
	})
}

func (p ToastProviderWidget) layoutToastText(ctx *frame.Context, gtx layout.Context, entry *toastEntryState, mobileAction bool, style toastStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Toast
	children := make([]layout.FlexChild, 0, 3)
	if entry.item.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(entry.item.title).
				Size(float32(tokens.TitleSize)).
				Weight(font.Medium).
				Color(style.title).
				Layout(ctx, gtx)
		}))
	}
	if entry.item.description != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(entry.item.description).
				Size(float32(tokens.DescriptionSize)).
				Color(style.description).
				Layout(ctx, gtx)
		}))
	}
	if mobileAction && entry.item.actionLabel != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutToastAction(ctx, gtx, entry)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (p ToastProviderWidget) layoutToastIndicator(ctx *frame.Context, gtx layout.Context, item ToastItem, style toastStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Toast
	box := gtx.Dp(tokens.IndicatorSize + 2*tokens.IndicatorPadding)
	gtx.Constraints = layout.Exact(image.Pt(box, box))
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Dp(tokens.IndicatorSize)
		gtx.Constraints = layout.Exact(image.Pt(size, size))
		if item.loading {
			restore := frame.PushColors(ctx, style.indicator, style.surface)
			defer restore()
			return spinner.Spinner().Color(spinner.SpinnerCurrent).Size(spinner.SpinnerSmall).Layout(ctx, gtx)
		}
		if item.hasIndicator && item.indicator != nil {
			restore := frame.PushColors(ctx, style.indicator, style.surface)
			defer restore()
			item.indicator.Layout(ctx, gtx)
			return layout.Dimensions{Size: image.Pt(size, size)}
		}
		drawToastIndicator(gtx, image.Pt(size, size), style.indicator, item.variant)
		return layout.Dimensions{Size: image.Pt(size, size)}
	})
}

func (p ToastProviderWidget) layoutToastAction(ctx *frame.Context, gtx layout.Context, entry *toastEntryState) layout.Dimensions {
	gtx.Constraints.Min.X = 0
	onClick := func() {
		if p.onAction != nil {
			p.onAction(entry.item.key)
		}
	}
	action := button.Button("action", text.New(entry.item.actionLabel)).
		Variant(entry.item.actionVariant).
		Size(button.ButtonSmall).
		OnClick(onClick)
	return button.LayoutWithClickable(action, ctx, gtx, &entry.action)
}

func (p ToastProviderWidget) layoutToastClose(ctx *frame.Context, gtx layout.Context, entry *toastEntryState, toastSize image.Point, style toastStyle, alwaysVisible bool) {
	tokens := frame.ActiveTheme(ctx).Components.Toast
	show := alwaysVisible || entry.hovered || entry.close.Hovered() || gtx.Focused(&entry.root) || gtx.Focused(&entry.close)
	if !show {
		return
	}
	size := gtx.Dp(tokens.CloseSize)
	inset := gtx.Dp(tokens.CloseInset)
	position := image.Pt(toastSize.X-size-inset, inset)
	closeGtx := gtx
	closeGtx.Constraints = layout.Exact(image.Pt(size, size))
	offset := op.Offset(position).Push(gtx.Ops)
	dims := entry.close.Layout(closeGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(toastCloseLabel(ctx)).Add(gtx.Ops)
		drawToastCloseButton(gtx, frame.ActiveTheme(ctx), image.Pt(size, size), style, entry.close.Hovered())
		iconSize := gtx.Dp(tokens.CloseIconSize)
		iconOffset := op.Offset(image.Pt((size-iconSize)/2, (size-iconSize)/2)).Push(gtx.Ops)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
		icon.Layout(lucide.X, iconGtx, style.description)
		iconOffset.Pop()
		return layout.Dimensions{Size: image.Pt(size, size)}
	})
	focusVisible := frame.FocusVisible(ctx, &entry.close, gtx.Focused(&entry.close))
	focusOpacity := entry.closeFocus.Opacity(gtx, focusVisible)
	drawToastFocus(gtx, image.Rectangle{Max: dims.Size}, size/2, style.focus, tokens.FocusRingWidth, focusOpacity)
	offset.Pop()
}

func (p ToastProviderWidget) resolvedGap(ctx *frame.Context) unit.Dp {
	if p.hasGap {
		return p.gap
	}
	return frame.ActiveTheme(ctx).Components.Toast.Gap
}

func (p ToastProviderWidget) resolvedOffset(ctx *frame.Context) unit.Dp {
	if p.hasOffset {
		return p.offset
	}
	return frame.ActiveTheme(ctx).Components.Toast.Inset
}

func (p ToastProviderWidget) resolvedMaxVisible(ctx *frame.Context) int {
	if p.hasMaxVisible {
		return p.maxVisible
	}
	return max(frame.ActiveTheme(ctx).Components.Toast.MaxVisible, 1)
}

func (p ToastProviderWidget) resolvedScaleFactor(ctx *frame.Context) float32 {
	if p.hasScale {
		return p.scaleFactor
	}
	return min(max(frame.ActiveTheme(ctx).Components.Toast.ScaleFactor, 0), 1)
}

func (p ToastProviderWidget) resolvedWidth(ctx *frame.Context) unit.Dp {
	if p.hasWidth {
		return p.width
	}
	return frame.ActiveTheme(ctx).Components.Toast.Width
}

func (p ToastProviderWidget) bottomPlacement() bool {
	return p.placement == ToastBottom || p.placement == ToastBottomStart || p.placement == ToastBottomEnd
}

func toastRegionX(viewportWidth, toastWidth, inset int, placement ToastPlacement) int {
	switch placement {
	case ToastBottomStart, ToastTopStart:
		return inset
	case ToastBottomEnd, ToastTopEnd:
		return max(viewportWidth-inset-toastWidth, 0)
	default:
		return max((viewportWidth-toastWidth)/2, 0)
	}
}

func toastRadius(gtx layout.Context, radius unit.Dp, size image.Point) int {
	return min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
}

func toastCloseLabel(ctx *frame.Context) string {
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "关闭"
	}
	return "Close"
}
