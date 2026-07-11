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
	"github.com/qianniancn/FlowUI/internal/components/closebutton"
	"github.com/qianniancn/FlowUI/internal/components/spinner"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/render"
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
	providerState.updateTimers(gtx, p.paused || providerState.paused(gtx), func(entry *toastEntryState) {
		entry.requestClose(p.onClose)
	})

	tokens := frame.ActiveTheme(ctx).Components.Toast
	maxVisible := p.resolvedMaxVisible(ctx)
	width := min(gtx.Dp(p.resolvedWidth(ctx)), max(viewport.X-2*gtx.Dp(tokens.Inset), 0))
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
		targetIndex := entry.stackTo
		if !exiting {
			targetIndex = float32(presentIndex)
			presentIndex++
		} else if !entry.stackReady {
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
		toastGtx.Constraints.Max = image.Pt(width, max(viewport.Y-2*gtx.Dp(tokens.Inset), 0))
		interactive := targetIndex == 0 && !exiting && gtx.Enabled()
		if !interactive && frontHeight > 0 {
			height := min(frontHeight, toastGtx.Constraints.Max.Y)
			toastGtx.Constraints.Min.Y = height
			toastGtx.Constraints.Max.Y = height
		}
		restoreItem := frame.PushKey(ctx, entry.item.key)
		macro := op.Record(gtx.Ops)
		dims := p.layoutToast(ctx, toastGtx, entry, interactive, mobile)
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

	p.paintRecords(ctx, gtx, viewport, records)
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

func (p ToastProviderWidget) paintRecords(ctx *frame.Context, gtx layout.Context, viewport image.Point, records []toastRecord) {
	if len(records) == 0 {
		return
	}
	tokens := frame.ActiveTheme(ctx).Components.Toast
	inset := gtx.Dp(tokens.Inset)
	gap := gtx.Dp(p.resolvedGap(ctx))
	scaleFactor := p.resolvedScaleFactor(ctx)
	paintOrder := append([]toastRecord(nil), records...)
	sort.SliceStable(paintOrder, func(i, j int) bool {
		if paintOrder[i].stack != paintOrder[j].stack {
			return paintOrder[i].stack > paintOrder[j].stack
		}
		return !paintOrder[i].exiting && paintOrder[j].exiting
	})

	for _, record := range paintOrder {
		scale := max(1-record.stack*scaleFactor, 0.5)
		x := toastRegionX(viewport.X, record.dims.Size.X, inset, p.placement)
		y := inset + int(record.stack*float32(gap)+0.5)
		enterOffset := -int(float32(record.dims.Size.Y) * (1 - record.progress))
		if p.bottomPlacement() {
			y = viewport.Y - inset - record.dims.Size.Y - int(record.stack*float32(gap)+0.5)
			enterOffset = -enterOffset
		}
		y += enterOffset

		offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
		transform := render.Scale(record.dims.Size, scale).Push(gtx.Ops)
		opacity := paint.PushOpacity(gtx.Ops, record.progress)
		record.call.Add(gtx.Ops)
		opacity.Pop()
		transform.Pop()
		offset.Pop()
	}
}

func (p ToastProviderWidget) layoutToast(ctx *frame.Context, gtx layout.Context, entry *toastEntryState, interactive, mobile bool) layout.Dimensions {
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
	drawToastSurface(gtx, frame.ActiveTheme(ctx), rect, radius, style.surface)

	if !interactive {
		clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
		contentCall.Add(gtx.Ops)
		clipStack.Pop()
		entry.rootFocus.Visible(false, nil)
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
	focusVisible := entry.rootFocus.Visible(gtx.Focused(&entry.root), nil)
	focusOpacity := entry.rootFocus.Opacity(gtx, focusVisible)
	drawToastFocus(gtx, rect, radius, style.focus, frame.ActiveTheme(ctx).Components.Toast.FocusRingWidth, focusOpacity)
	p.layoutToastClose(ctx, gtx, entry, size, style)
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
		drawToastIndicator(gtx, image.Pt(size, size), style.surface, style.indicator, item.variant)
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

func (p ToastProviderWidget) layoutToastClose(ctx *frame.Context, gtx layout.Context, entry *toastEntryState, toastSize image.Point, style toastStyle) {
	tokens := frame.ActiveTheme(ctx).Components.Toast
	show := entry.hovered || entry.close.Hovered() || gtx.Focused(&entry.root) || gtx.Focused(&entry.close)
	if !show {
		entry.closeFocus.Visible(false, entry.close.History())
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
		closebutton.DrawIcon(gtx, image.Pt(iconSize, iconSize), style.description)
		iconOffset.Pop()
		return layout.Dimensions{Size: image.Pt(size, size)}
	})
	focusVisible := entry.closeFocus.Visible(gtx.Focused(&entry.close), entry.close.History())
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
