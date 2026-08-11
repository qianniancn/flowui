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
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/dropdown"
	"github.com/qianniancn/flowui/internal/components/icon"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/components/tooltip"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/render"
	stateutil "github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

func (w Widget) layout(ctx *frame.Context, gtx layout.Context, state *sidebarState, entries []entry) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Sidebar
	padding := tokens.Padding
	if w.hasPadding {
		padding = w.padding
	}
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
	expandedContentWidth := max(expandedWidthPx-gtx.Dp(padding)*2, 0)
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
		content = layoutui.LayoutTrackedInset(ctx, rootGtx, layout.UniformInset(padding), func(gtx layout.Context) layout.Dimensions {
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
	itemGap := tokens.ItemGap
	if w.hasItemGap {
		itemGap = w.itemGap
	}
	state.list.Gap = gtx.Dp(itemGap)
	state.list.Alignment = layout.Start
	state.list.ScrollToEnd = false
	state.list.ScrollAnyAxis = false
	return layoutui.LayoutTrackedScrollbarWithVisualOutset(ctx, gtx, &state.list, &state.bar, &state.visualOutset, len(entries), w.disabled, true, func(gtx layout.Context, index int) layout.Dimensions {
		entry := entries[index]
		if entry.section {
			return w.layoutSection(ctx, gtx, entry.title, expansion)
		}
		return w.layoutItem(ctx, gtx, state, state.item(entry.item.Key), entry, expansion)
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
		label := material.Label(frame.ActiveMaterial(ctx), tokens.SectionTextSize, title)
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

func (w Widget) layoutItem(ctx *frame.Context, gtx layout.Context, sidebarState *sidebarState, itemState *sidebarItemState, entry entry, expansion float32) layout.Dimensions {
	item := entry.item
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Sidebar
	disabled := w.itemDisabled(item)
	if gtx.Focused(&itemState.clickable) {
		sidebarState.focusedKey = item.Key
	}
	if !disabled {
		for itemState.clickable.Clicked(gtx) {
			if len(item.Children) > 0 && !w.collapsed {
				w.requestOpen(item.Key, !entry.expanded)
			} else {
				w.activate(item.Key)
			}
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
	hovered := itemState.clickable.Hovered() && !disabled
	style := sidebarItemStyleFor(activeTheme, selected, hovered, itemState.clickable.Pressed() && !disabled, disabled)
	if w.collapsed || w.expandAction != ExpandActionHover || len(item.Children) == 0 || !hovered || entry.expanded {
		itemState.hoverOpenRequested = false
	} else if !itemState.hoverOpenRequested {
		itemState.hoverOpenRequested = true
		w.requestOpen(item.Key, true)
	}
	focusVisible := frame.FocusVisible(ctx, &itemState.clickable, itemGtx.Focused(&itemState.clickable))
	focus := itemState.focus.Opacity(gtx, focusVisible && !disabled, frame.ActiveTheme(ctx).Motion)
	if w.collapsed && len(item.Children) > 0 {
		return w.layoutCollapsedSubmenu(ctx, itemGtx, entry, style, selected, disabled, focus)
	}

	if !w.collapsed || expansion > 0 || item.Label == "" {
		return w.layoutItemTrigger(ctx, itemGtx, size, itemState, entry, style, selected, disabled, focus, expansion)
	}
	trigger := frame.WidgetFunc(func(ctx *frame.Context, triggerGtx layout.Context) layout.Dimensions {
		return w.layoutItemTrigger(ctx, triggerGtx, size, itemState, entry, style, selected, disabled, focus, expansion)
	})
	return tooltip.Tooltip("sidebar-item-label:"+item.Key, trigger, text.New(item.Label)).
		Placement(overlay.PopoverRight).
		Delay(0).
		Layout(ctx, itemGtx)
}

func (w Widget) layoutItemTrigger(ctx *frame.Context, gtx layout.Context, size image.Point, itemState *sidebarItemState, entry entry, style sidebarItemStyle, selected, disabled bool, focus, expansion float32) layout.Dimensions {
	gtx.Constraints = layout.Exact(size)
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return w.layoutItemSurface(ctx, gtx, entry, style, selected, !disabled, focus, expansion)
	})
}

func (w Widget) layoutItemSurface(ctx *frame.Context, gtx layout.Context, entry entry, style sidebarItemStyle, selected, enabled bool, focus, expansion float32) layout.Dimensions {
	item := entry.item
	semantic.Button.Add(gtx.Ops)
	semantic.LabelOp(item.Label).Add(gtx.Ops)
	semantic.SelectedOp(selected).Add(gtx.Ops)
	semantic.EnabledOp(enabled).Add(gtx.Ops)
	if len(item.Children) > 0 {
		description := "Navigation group. Activate to expand."
		if entry.expanded {
			description = "Navigation group. Activate to collapse."
		}
		semantic.DescriptionOp(description).Add(gtx.Ops)
	}

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
			return w.layoutItemContent(ctx, contentGtx, entry, style.foreground, expansion)
		})
	}()
	content := macro.Stop()
	contentOffset := image.Pt(0, max((size.Y-contentDims.Size.Y)/2, 0))
	contentPlacement.PlaceOffset(contentOffset)

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	itemRadius := frame.ActiveTheme(ctx).Components.Sidebar.ItemRadius
	if w.hasItemRadius {
		itemRadius = w.itemRadius
	}
	drawSidebarItem(gtx, frame.ActiveTheme(ctx), size, style, focus, itemRadius)
	offset := op.Offset(contentOffset).Push(gtx.Ops)
	content.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	return layout.Dimensions{Size: size}
}

func (w Widget) layoutItemContent(ctx *frame.Context, gtx layout.Context, entry entry, foreground color.NRGBA, expansion float32) layout.Dimensions {
	item := entry.item
	tokens := frame.ActiveTheme(ctx).Components.Sidebar
	width := gtx.Constraints.Max.X
	itemPadding := tokens.ItemPaddingX
	if w.hasItemPadding {
		itemPadding = w.itemPaddingX
	}
	padding := gtx.Dp(itemPadding)
	gap := gtx.Dp(tokens.ItemContentGap)
	outerPadding := tokens.Padding
	if w.hasPadding {
		outerPadding = w.padding
	}
	collapsedWidth := tokens.CollapsedWidth
	if w.collapsedWidth > 0 {
		collapsedWidth = w.collapsedWidth
	}
	collapsedContentWidth := min(max(gtx.Dp(collapsedWidth)-gtx.Dp(outerPadding)*2, 0), width)
	indent := 0
	if !w.collapsed && w.hasInlineIndent {
		indent = gtx.Dp(w.inlineIndent) * entry.depth
	} else if !w.collapsed {
		indent = gtx.Dp(tokens.InlineIndent) * entry.depth
	}
	contentWidth := max(width-indent, 0)
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}
	childGtx.Constraints.Max.X = max(contentWidth-padding*2, 0)

	var switcher sidebarContent
	if !w.collapsed && len(item.Children) > 0 {
		switcher = recordSidebarContent(ctx, childGtx, sidebarItemSwitcher(item, entry.expanded, foreground, tokens.SwitcherSize))
	}

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
	switcherX := width - padding - switcher.dims.Size.X
	trailingX := switcherX - trailing.dims.Size.X
	if trailing.dims.Size.X > 0 && switcher.dims.Size.X > 0 {
		trailingX -= gap
	}

	leadingWidth := leading.dims.Size.X
	labelX := indent + padding
	leadingX := 0
	if leadingWidth > 0 {
		collapsedX := (collapsedContentWidth - leadingWidth) / 2
		leadingX = int(animation.LerpFloat(float32(collapsedX), float32(labelX), expansion) + .5)
		labelX = leadingX + leadingWidth + gap
	}

	var label sidebarContent
	if expansion > 0 {
		labelGtx := childGtx
		labelRight := switcherX
		if trailing.dims.Size.X > 0 {
			labelRight = trailingX - gap
		} else if switcher.dims.Size.X > 0 {
			labelRight -= gap
		}
		labelGtx.Constraints.Max.X = max(labelRight-labelX, 0)
		label = recordSidebarContent(ctx, labelGtx, sidebarItemLabel(item.Label, foreground, tokens.ItemTextSize, false))
	}

	height := max(leading.dims.Size.Y, initial.dims.Size.Y, label.dims.Size.Y, trailing.dims.Size.Y, switcher.dims.Size.Y)
	centerY := func(content sidebarContent) int { return max((height-content.dims.Size.Y)/2, 0) }
	if leadingWidth > 0 {
		drawSidebarContent(gtx, leading, image.Pt(leadingX, centerY(leading)), 1)
	} else {
		initialX := max((collapsedContentWidth-initial.dims.Size.X)/2, 0)
		drawSidebarContent(gtx, initial, image.Pt(initialX, centerY(initial)), 1-expansion)
	}
	drawSidebarContent(gtx, label, image.Pt(labelX, centerY(label)), expansion)
	drawSidebarContent(gtx, trailing, image.Pt(trailingX, centerY(trailing)), expansion)
	if switcher.dims.Size.X > 0 {
		drawSidebarContent(gtx, switcher, image.Pt(switcherX, centerY(switcher)), 1)
	}
	return layout.Dimensions{Size: image.Pt(width, height)}
}

func (w Widget) layoutCollapsedSubmenu(ctx *frame.Context, gtx layout.Context, entry entry, style sidebarItemStyle, selected, disabled bool, focus float32) layout.Dimensions {
	trigger := frame.WidgetFunc(func(ctx *frame.Context, triggerGtx layout.Context) layout.Dimensions {
		return w.layoutItemSurface(ctx, triggerGtx, entry, style, selected, !disabled, focus, 0)
	})
	popup := dropdown.New("sidebar-submenu:"+entry.item.Key, trigger, w.sidebarDropdownItems(entry.item.Children)).
		Placement(overlay.PopoverRightStart).
		// A collapsed navigation item opens its flyout on hover. ExpandAction
		// only controls inline groups while the sidebar is expanded.
		TriggerMode(dropdown.TriggerHover).
		AutoWidth().
		SelectionMode(dropdown.SelectionSingle).
		SelectedKey(w.selectedKey).
		Disabled(disabled || w.disabled).
		OnActionEvent(func(event dropdown.ActionEvent) { w.activate(event.Key) })
	if w.hasDataVersion {
		popup = popup.DataVersion(w.dataVersion)
	}
	return popup.Layout(ctx, gtx)
}

func (w Widget) sidebarDropdownItems(items []Item) []dropdown.Item {
	result := make([]dropdown.Item, 0, len(items))
	for _, item := range items {
		result = append(result, dropdown.Item{
			Key:      item.Key,
			Label:    item.Label,
			Leading:  item.Leading,
			Trailing: item.Trailing,
			Disabled: w.itemDisabled(item),
			Children: w.sidebarDropdownItems(item.Children),
		})
	}
	return result
}

func sidebarItemSwitcher(item Item, expanded bool, foreground color.NRGBA, size unit.Dp) frame.Widget {
	value := item.Switcher
	if expanded && item.ExpandedSwitcher != nil {
		value = item.ExpandedSwitcher
	}
	if value != nil {
		return value
	}
	data := lucide.ChevronRight
	if expanded {
		data = lucide.ChevronDown
	}
	return icon.New(data).Size(float32(size)).Color(foreground)
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
		value := material.Label(frame.ActiveMaterial(ctx), size, label)
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

func drawSidebarItem(gtx layout.Context, activeTheme *theme.Theme, size image.Point, style sidebarItemStyle, focus float32, itemRadius unit.Dp) {
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(itemRadius), 0), min(size.X, size.Y)/2)
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
	render.DrawRoundedInsetStroke(gtx, rect, max(radius-inset, 0), width, inset, focusColor)
}
