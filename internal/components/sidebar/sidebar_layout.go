package sidebar

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/animation"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/components/tooltip"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func (w Widget) layout(ctx *frame.Context, gtx layout.Context, state *sidebarState, entries []entry) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Sidebar
	expandedWidth := tokens.Width
	if w.width > 0 {
		expandedWidth = w.width
	}
	collapsedWidth := tokens.CollapsedWidth
	if w.collapsedWidth > 0 {
		collapsedWidth = w.collapsedWidth
	}
	target := float32(1)
	easing := animation.EaseCubicOut
	if w.collapsed {
		target = 0
		easing = animation.EaseCubicInOut
	}
	expansion := animation.Tween("expansion", target).
		Easing(easing).
		Value(ctx, gtx)
	expandedWidthPx := min(gtx.Dp(expandedWidth), gtx.Constraints.Max.X)
	width := int(animation.LerpFloat(float32(gtx.Dp(collapsedWidth)), float32(expandedWidthPx), expansion) + .5)
	size := gtx.Constraints.Constrain(image.Pt(width, gtx.Constraints.Max.Y))
	expandedContentWidth := max(expandedWidthPx-gtx.Dp(tokens.Padding)*2, 0)
	rootGtx := gtx
	rootGtx.Constraints = layout.Exact(size)
	if w.alt != "" {
		semantic.DescriptionOp(w.alt).Add(rootGtx.Ops)
	}

	macro := op.Record(rootGtx.Ops)
	var content layout.Dimensions
	func() {
		restore := frame.PushColors(ctx, activeTheme.Palette.SurfaceForeground, activeTheme.Palette.Surface)
		defer restore()
		content = layoutui.LayoutTrackedInset(ctx, rootGtx, layout.UniformInset(tokens.Padding), func(gtx layout.Context) layout.Dimensions {
			children := make([]frame.Widget, 0, 3)
			if w.header != nil {
				children = append(children, sidebarTransitionContent(w.header, w.targetContentOpacity(expansion), w.expandedLayoutWidth(expandedContentWidth)))
			}
			children = append(children, layoutui.Expanded(frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
				return w.layoutEntries(ctx, gtx, state, entries, expansion)
			})))
			if w.footer != nil {
				children = append(children, sidebarTransitionContent(w.footer, w.targetContentOpacity(expansion), w.expandedLayoutWidth(expandedContentWidth)))
			}
			return layoutui.LayoutTrackedFlex(ctx, gtx, layout.Vertical, tokens.ContentGap, layout.Start, children...)
		})
	}()
	call := macro.Stop()

	drawSidebarRoot(rootGtx, activeTheme, size)
	area := clip.Rect{Max: size}.Push(rootGtx.Ops)
	call.Add(rootGtx.Ops)
	area.Pop()
	return layout.Dimensions{Size: size, Baseline: content.Baseline}
}

func (w Widget) layoutEntries(ctx *frame.Context, gtx layout.Context, state *sidebarState, entries []entry, expansion float32) layout.Dimensions {
	if len(entries) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return text.New(w.emptyText).
				Size(float32(frame.ActiveTheme(ctx).Components.Sidebar.ItemTextSize)).
				Color(frame.ActiveTheme(ctx).Palette.MutedForeground).
				Layout(ctx, gtx)
		})
	}
	tokens := frame.ActiveTheme(ctx).Components.Sidebar
	state.list.Axis = layout.Vertical
	state.list.Gap = gtx.Dp(tokens.ItemGap)
	state.list.Alignment = layout.Start
	state.list.ScrollToEnd = false
	state.list.ScrollAnyAxis = false
	return layoutui.LayoutTrackedScrollbar(ctx, gtx, &state.list, &state.bar, len(entries), w.disabled, true, func(gtx layout.Context, index int) layout.Dimensions {
		entry := entries[index]
		if entry.section {
			return w.layoutSection(ctx, gtx, entry.title, expansion)
		}
		return w.layoutItem(ctx, gtx, state, state.item(entry.item.Key), entry.item, expansion)
	})
}

func (w Widget) layoutSection(ctx *frame.Context, gtx layout.Context, title string, expansion float32) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Sidebar
	height := min(gtx.Dp(tokens.SectionHeight), gtx.Constraints.Max.Y)
	sectionGtx := gtx
	sectionGtx.Constraints.Min.Y = height
	sectionGtx.Constraints.Max.Y = height
	fade := paint.PushOpacity(gtx.Ops, w.targetContentOpacity(expansion))
	defer fade.Pop()
	if w.collapsed {
		return layout.Center.Layout(sectionGtx, func(gtx layout.Context) layout.Dimensions {
			width := max(gtx.Constraints.Max.X-gtx.Dp(tokens.SectionSeparatorInset)*2, 0)
			height := max(gtx.Dp(tokens.BorderWidth), 1)
			rect := image.Rect(0, 0, width, height)
			paint.FillShape(gtx.Ops, frame.ActiveTheme(ctx).Palette.Border, clip.Rect(rect).Op())
			return layout.Dimensions{Size: rect.Size()}
		})
	}
	return layout.W.Layout(sectionGtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(frame.ActiveTheme(ctx).Material, tokens.SectionTextSize, title)
		label.Color = frame.ActiveTheme(ctx).Palette.MutedForeground
		label.Font.Weight = font.Medium
		label.MaxLines = 1
		return label.Layout(gtx)
	})
}

func (w Widget) targetContentOpacity(expansion float32) float32 {
	if w.collapsed {
		return 1 - expansion
	}
	return expansion
}

func (w Widget) expandedLayoutWidth(width int) int {
	if w.collapsed {
		return 0
	}
	return width
}

func sidebarTransitionContent(child frame.Widget, opacity float32, width int) frame.Widget {
	return frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		if width > 0 {
			gtx.Constraints.Min.X = width
			gtx.Constraints.Max.X = width
		}
		content := recordSidebarContent(ctx, gtx, child)
		drawSidebarContent(gtx, content, image.Point{}, opacity)
		return content.dims
	})
}

func (w Widget) layoutItem(ctx *frame.Context, gtx layout.Context, sidebarState *sidebarState, itemState *sidebarItemState, item Item, expansion float32) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Sidebar
	disabled := w.itemDisabled(item)
	if gtx.Focused(&itemState.clickable) {
		sidebarState.focusedKey = item.Key
	}
	if !disabled {
		for itemState.clickable.Clicked(gtx) {
			w.activate(item.Key)
		}
	}
	itemHeight := tokens.ItemHeight
	if w.itemHeight > 0 {
		itemHeight = w.itemHeight
	}
	height := min(max(gtx.Dp(itemHeight), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	size := image.Pt(gtx.Constraints.Max.X, height)
	itemGtx := gtx
	itemGtx.Constraints = layout.Exact(size)
	presses := stateutil.ActivePresses(itemState.clickable.History())
	if !disabled {
		frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}
	selected := item.Key == w.selectedKey
	style := sidebarItemStyleFor(activeTheme, selected, itemState.clickable.Hovered() && !disabled, itemState.clickable.Pressed() && !disabled, disabled)
	focusVisible := frame.FocusVisible(ctx, &itemState.clickable, itemGtx.Focused(&itemState.clickable))
	focus := itemState.focus.Opacity(gtx, focusVisible && !disabled, frame.ActiveTheme(ctx).Motion)

	if !w.collapsed || expansion > 0 || item.Label == "" {
		return w.layoutItemTrigger(ctx, itemGtx, size, itemState, item, style, selected, disabled, focus, expansion)
	}
	trigger := frame.WidgetFunc(func(ctx *frame.Context, triggerGtx layout.Context) layout.Dimensions {
		return w.layoutItemTrigger(ctx, triggerGtx, size, itemState, item, style, selected, disabled, focus, expansion)
	})
	return tooltip.Tooltip("sidebar-item-label:"+item.Key, trigger, text.New(item.Label)).
		Placement(overlay.PopoverRight).
		Delay(0).
		Layout(ctx, itemGtx)
}

func (w Widget) layoutItemTrigger(ctx *frame.Context, gtx layout.Context, size image.Point, itemState *sidebarItemState, item Item, style sidebarItemStyle, selected, disabled bool, focus, expansion float32) layout.Dimensions {
	gtx.Constraints = layout.Exact(size)
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return w.layoutItemVisual(ctx, gtx, item, style, selected, !disabled, focus, expansion)
	})
}

func (w Widget) layoutItemVisual(ctx *frame.Context, gtx layout.Context, item Item, style sidebarItemStyle, selected, enabled bool, focus, expansion float32) layout.Dimensions {
	semantic.Button.Add(gtx.Ops)
	semantic.LabelOp(item.Label).Add(gtx.Ops)
	semantic.SelectedOp(selected).Add(gtx.Ops)
	semantic.EnabledOp(enabled).Add(gtx.Ops)

	size := gtx.Constraints.Min
	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	var contentPlacement frame.OverlayPlacement
	func() {
		restore := frame.PushColors(ctx, style.foreground, style.background)
		defer restore()
		contentGtx := gtx
		contentGtx.Constraints.Min = image.Pt(size.X, 0)
		contentGtx.Constraints.Max = size
		contentDims, contentPlacement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return w.layoutItemContent(ctx, contentGtx, item, style.foreground, expansion)
		})
	}()
	content := macro.Stop()
	contentOffset := image.Pt(0, max((size.Y-contentDims.Size.Y)/2, 0))
	contentPlacement.PlaceOffset(contentOffset)

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	drawSidebarItem(gtx, frame.ActiveTheme(ctx), size, style, focus)
	offset := op.Offset(contentOffset).Push(gtx.Ops)
	content.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	return layout.Dimensions{Size: size}
}

func (w Widget) layoutItemContent(ctx *frame.Context, gtx layout.Context, item Item, foreground color.NRGBA, expansion float32) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Sidebar
	width := gtx.Constraints.Max.X
	padding := gtx.Dp(tokens.ItemPaddingX)
	gap := gtx.Dp(tokens.ItemContentGap)
	collapsedWidth := tokens.CollapsedWidth
	if w.collapsedWidth > 0 {
		collapsedWidth = w.collapsedWidth
	}
	collapsedContentWidth := min(max(gtx.Dp(collapsedWidth)-gtx.Dp(tokens.Padding)*2, 0), width)
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}
	childGtx.Constraints.Max.X = max(width-padding*2, 0)

	var leading, initial sidebarContent
	if item.Leading != nil {
		leading = recordSidebarContent(ctx, childGtx, item.Leading)
	} else if expansion < 1 {
		initial = recordSidebarContent(ctx, childGtx, sidebarItemLabel(item.Label, foreground, tokens.ItemTextSize, true))
	}

	var trailing sidebarContent
	if expansion > 0 && item.Trailing != nil {
		trailing = recordSidebarContent(ctx, childGtx, item.Trailing)
	}
	trailingX := width - padding - trailing.dims.Size.X

	leadingWidth := leading.dims.Size.X
	labelX := padding
	leadingX := 0
	if leadingWidth > 0 {
		collapsedX := (collapsedContentWidth - leadingWidth) / 2
		leadingX = int(animation.LerpFloat(float32(collapsedX), float32(padding), expansion) + .5)
		labelX = leadingX + leadingWidth + gap
	}

	var label sidebarContent
	if expansion > 0 {
		labelGtx := childGtx
		labelRight := width - padding
		if trailing.dims.Size.X > 0 {
			labelRight = trailingX - gap
		}
		labelGtx.Constraints.Max.X = max(labelRight-labelX, 0)
		label = recordSidebarContent(ctx, labelGtx, sidebarItemLabel(item.Label, foreground, tokens.ItemTextSize, false))
	}

	height := max(leading.dims.Size.Y, initial.dims.Size.Y, label.dims.Size.Y, trailing.dims.Size.Y)
	centerY := func(content sidebarContent) int { return max((height-content.dims.Size.Y)/2, 0) }
	if leadingWidth > 0 {
		drawSidebarContent(gtx, leading, image.Pt(leadingX, centerY(leading)), 1)
	} else {
		initialX := max((collapsedContentWidth-initial.dims.Size.X)/2, 0)
		drawSidebarContent(gtx, initial, image.Pt(initialX, centerY(initial)), 1-expansion)
	}
	drawSidebarContent(gtx, label, image.Pt(labelX, centerY(label)), expansion)
	drawSidebarContent(gtx, trailing, image.Pt(trailingX, centerY(trailing)), expansion)
	return layout.Dimensions{Size: image.Pt(width, height)}
}

type sidebarContent struct {
	call      op.CallOp
	dims      layout.Dimensions
	placement frame.OverlayPlacement
}

func recordSidebarContent(ctx *frame.Context, gtx layout.Context, child frame.Widget) sidebarContent {
	if child == nil {
		return sidebarContent{}
	}
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackWidgetPlacement(ctx, gtx, child)
	return sidebarContent{call: macro.Stop(), dims: dims, placement: placement}
}

func drawSidebarContent(gtx layout.Context, content sidebarContent, position image.Point, opacity float32) {
	content.placement.PlaceOffset(position)
	content.placement.SetOpacity(opacity)
	if opacity <= 0 {
		return
	}
	offset := op.Offset(position).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	content.call.Add(gtx.Ops)
	fade.Pop()
	offset.Pop()
}

func sidebarItemLabel(label string, foreground color.NRGBA, size unit.Sp, initial bool) frame.Widget {
	if initial {
		label = sidebarInitial(label)
	}
	return frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		value := material.Label(frame.ActiveTheme(ctx).Material, size, label)
		value.Color = foreground
		value.Font.Weight = font.Medium
		value.MaxLines = 1
		return value.Layout(gtx)
	})
}

func sidebarInitial(label string) string {
	for _, value := range label {
		return string(value)
	}
	return ""
}

type sidebarItemStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	focus      color.NRGBA
	opacity    float32
}

func sidebarItemStyleFor(activeTheme *theme.Theme, selected, hovered, pressed, disabled bool) sidebarItemStyle {
	style := sidebarItemStyle{foreground: activeTheme.Palette.Foreground, focus: activeTheme.Palette.Focus, opacity: 1}
	if hovered {
		style.background = activeTheme.Palette.SurfaceTertiary
	}
	if pressed {
		style.background = activeTheme.Palette.SurfacePressed
	}
	if selected {
		style.background = activeTheme.Palette.AccentSoft
		style.foreground = activeTheme.Palette.AccentSoftForeground
		if hovered || pressed {
			style.background = activeTheme.Palette.AccentSoftHover
		}
	}
	if disabled {
		style.opacity = activeTheme.DisabledOpacityValue()
		style.focus = color.NRGBA{}
	}
	return style
}

func drawSidebarRoot(gtx layout.Context, activeTheme *theme.Theme, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	paint.FillShape(gtx.Ops, activeTheme.Palette.Surface, clip.Rect{Max: size}.Op())
	width := max(gtx.Dp(activeTheme.Components.Sidebar.BorderWidth), 1)
	border := image.Rect(max(size.X-width, 0), 0, size.X, size.Y)
	paint.FillShape(gtx.Ops, activeTheme.Palette.Border, clip.Rect(border).Op())
}

func drawSidebarItem(gtx layout.Context, activeTheme *theme.Theme, size image.Point, style sidebarItemStyle, focus float32) {
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(activeTheme.Components.Sidebar.ItemRadius), 0), min(size.X, size.Y)/2)
	if style.background.A != 0 {
		paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(activeTheme.Components.Sidebar.FocusRingWidth), 1)
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	focusColor := style.focus
	focusColor.A = byte(float32(focusColor.A)*focus + .5)
	stroke := clip.Stroke{Path: clip.UniformRRect(focusRect, max(radius-inset, 0)).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, focusColor)
	stroke.Pop()
}
