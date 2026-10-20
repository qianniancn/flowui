package sidebar

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/dropdown"

	"github.com/qianniancn/flowui/widget/tooltip"
)

func (w Widget) layout(ctx *core.Context, gtx layout.Context, state *sidebarState, entries []entry, visual theme.SidebarVisual) layout.Dimensions {
	tokens := visual.Metrics
	padding := tokens.Padding
	expandedWidth := tokens.Width
	collapsedWidth := tokens.CollapsedWidth
	target := float32(1)
	easing := animation.EasingFor(visual.ExpansionEasing)
	if w.collapsed {
		target = 0
		easing = animation.EasingFor(visual.CollapseEasing)
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
		restore := core.PushColors(ctx, visual.Foreground, visual.Background)
		defer restore()
		content = layoutui.TrackedInset(ctx, rootGtx, layout.UniformInset(padding), func(gtx layout.Context) layout.Dimensions {
			children := make([]core.Widget, 0, 3)
			if w.header != nil {
				children = append(children, sidebarTransitionContent(w.header, w.targetContentOpacity(expansion), w.expandedLayoutWidth(expandedContentWidth)))
			}
			children = append(children, layoutui.Expanded(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
				return w.layoutEntries(ctx, gtx, state, entries, expansion, visual)
			})))
			if w.footer != nil {
				children = append(children, sidebarTransitionContent(w.footer, w.targetContentOpacity(expansion), w.expandedLayoutWidth(expandedContentWidth)))
			}
			return layoutui.TrackedFlex(ctx, gtx, layout.Vertical, tokens.ContentGap, layout.Start, children...)
		})
	}()
	call := macro.Stop()

	if visual.Root != nil {
		visual.Root(rootGtx, size)
	}
	area := clip.Rect{Max: size}.Push(rootGtx.Ops)
	call.Add(rootGtx.Ops)
	area.Pop()
	return layout.Dimensions{Size: size, Baseline: content.Baseline}
}

func (w Widget) layoutEntries(ctx *core.Context, gtx layout.Context, state *sidebarState, entries []entry, expansion float32, visual theme.SidebarVisual) layout.Dimensions {
	if len(entries) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if visual.Text == nil {
				return layout.Dimensions{}
			}
			return visual.Text(gtx, theme.SidebarTextRequest{Text: w.resolvedEmptyText(ctx), Size: visual.Metrics.ItemTextSize, Color: visual.MutedForeground})
		})
	}
	tokens := visual.Metrics
	state.list.Axis = layout.Vertical
	state.list.Gap = gtx.Dp(tokens.ItemGap)
	state.list.Alignment = layout.Start
	state.list.ScrollToEnd = false
	state.list.ScrollAnyAxis = false
	return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &state.list, &state.bar, &state.visualOutset, len(entries), w.disabled, true, func(gtx layout.Context, index int) layout.Dimensions {
		entry := &entries[index]
		if entry.section {
			return w.layoutSection(ctx, gtx, entry.title, expansion, visual)
		}
		return w.layoutItem(ctx, gtx, state, state.item(entry.item.Key), entry, expansion, visual)
	})
}

func (w Widget) layoutSection(ctx *core.Context, gtx layout.Context, title string, expansion float32, visual theme.SidebarVisual) layout.Dimensions {
	tokens := visual.Metrics
	height := min(gtx.Dp(tokens.SectionHeight), gtx.Constraints.Max.Y)
	sectionGtx := gtx
	sectionGtx.Constraints.Min.Y = height
	sectionGtx.Constraints.Max.Y = height
	fade := paint.PushOpacity(gtx.Ops, w.targetContentOpacity(expansion))
	defer fade.Pop()
	if w.collapsed {
		if visual.Section == nil {
			return layout.Dimensions{}
		}
		return visual.Section(sectionGtx, theme.SidebarSectionRequest{Collapsed: true})
	}
	if visual.Section == nil {
		return layout.Dimensions{}
	}
	return visual.Section(sectionGtx, theme.SidebarSectionRequest{Title: title})
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

func sidebarTransitionContent(child core.Widget, opacity float32, width int) core.Widget {
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if width > 0 {
			gtx.Constraints.Min.X = width
			gtx.Constraints.Max.X = width
		}
		content := recordSidebarContent(ctx, gtx, child)
		drawSidebarContent(gtx, content, image.Point{}, opacity)
		return content.dims
	})
}

func (w Widget) layoutItem(ctx *core.Context, gtx layout.Context, sidebarState *sidebarState, itemState *sidebarItemState, entry *entry, expansion float32, visual theme.SidebarVisual) layout.Dimensions {
	item := entry.item
	tokens := visual.Metrics
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
	height := min(max(gtx.Dp(itemHeight), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	size := image.Pt(gtx.Constraints.Max.X, height)
	itemGtx := gtx
	itemGtx.Constraints = layout.Exact(size)
	presses := core.ActivePresses(itemState.clickable.History())
	if !disabled {
		core.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}
	selected := item.Key == w.selectedKey
	hovered := itemState.clickable.Hovered() && !disabled
	itemVisual := theme.SidebarItemVisual{}
	focusVisible := core.FocusVisible(ctx, &itemState.clickable, itemGtx.Focused(&itemState.clickable))
	if visual.Item != nil {
		itemVisual = sidebarState.sidebarItemVisualFor(ctx, visual, theme.SidebarItemRequest{
			Selected: selected, Hovered: hovered,
			Pressed: itemState.clickable.Pressed() && !disabled,
			Focused: focusVisible, Disabled: disabled,
		})
	}
	if w.collapsed || w.expandAction != ExpandActionHover || len(item.Children) == 0 || !hovered || entry.expanded {
		itemState.hoverOpenRequested = false
	} else if !itemState.hoverOpenRequested {
		itemState.hoverOpenRequested = true
		w.requestOpen(item.Key, true)
	}
	focus := itemState.focus.Opacity(gtx, focusVisible && !disabled, visual.Motion)
	if w.collapsed && len(item.Children) > 0 {
		return w.layoutCollapsedSubmenu(ctx, itemGtx, entry, itemVisual, selected, disabled, focus, visual)
	}

	if !w.collapsed || expansion > 0 || item.Label == "" {
		return w.layoutItemTrigger(ctx, itemGtx, size, itemState, entry, itemVisual, selected, disabled, focus, expansion, visual)
	}
	trigger := core.WidgetFunc(func(ctx *core.Context, triggerGtx layout.Context) layout.Dimensions {
		return w.layoutItemTrigger(ctx, triggerGtx, size, itemState, entry, itemVisual, selected, disabled, focus, expansion, visual)
	})
	return tooltip.New("sidebar-item-label:"+item.Key, trigger, item.Label).
		Placement(core.PopoverRight).
		Delay(0).
		Layout(ctx, itemGtx)
}

func (w Widget) layoutItemTrigger(ctx *core.Context, gtx layout.Context, size image.Point, itemState *sidebarItemState, entry *entry, itemVisual theme.SidebarItemVisual, selected, disabled bool, focus, expansion float32, visual theme.SidebarVisual) layout.Dimensions {
	gtx.Constraints = layout.Exact(size)
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return w.layoutItemSurface(ctx, gtx, entry, itemVisual, selected, !disabled, focus, expansion, visual)
	})
}

func (w Widget) layoutItemSurface(ctx *core.Context, gtx layout.Context, entry *entry, itemVisual theme.SidebarItemVisual, selected, enabled bool, focus, expansion float32, visual theme.SidebarVisual) layout.Dimensions {
	item := entry.item
	semantic.Button.Add(gtx.Ops)
	semantic.LabelOp(item.Label).Add(gtx.Ops)
	semantic.SelectedOp(selected).Add(gtx.Ops)
	semantic.EnabledOp(enabled).Add(gtx.Ops)
	if len(item.Children) > 0 {
		semantic.DescriptionOp(sidebarGroupDescription(ctx, entry.expanded)).Add(gtx.Ops)
	}

	size := gtx.Constraints.Min
	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	var contentPlacement core.OverlayPlacement
	func() {
		restore := core.PushColors(ctx, itemVisual.Foreground, itemVisual.Background)
		defer restore()
		contentGtx := gtx
		contentGtx.Constraints.Min = image.Pt(size.X, 0)
		contentGtx.Constraints.Max = size
		contentDims, contentPlacement = core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return w.layoutItemContent(ctx, contentGtx, entry, itemVisual.Foreground, expansion, visual)
		})
	}()
	content := macro.Stop()
	contentOffset := image.Pt(0, max((size.Y-contentDims.Size.Y)/2, 0))
	contentPlacement.PlaceOffset(contentOffset)

	opacity := paint.PushOpacity(gtx.Ops, itemVisual.Opacity)
	if visual.RenderItem != nil {
		visual.RenderItem(gtx, size, itemVisual, focus)
	}
	offset := op.Offset(contentOffset).Push(gtx.Ops)
	content.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	return layout.Dimensions{Size: size}
}

func sidebarGroupDescription(ctx *core.Context, expanded bool) string {
	if expanded {
		return ctx.Text(i18n.SidebarGroupCollapse)
	}
	return ctx.Text(i18n.SidebarGroupExpand)
}

func (w Widget) layoutItemContent(ctx *core.Context, gtx layout.Context, entry *entry, foreground color.NRGBA, expansion float32, visual theme.SidebarVisual) layout.Dimensions {
	item := entry.item
	tokens := visual.Metrics
	width := gtx.Constraints.Max.X
	itemPadding := tokens.ItemPaddingX
	padding := gtx.Dp(itemPadding)
	gap := gtx.Dp(tokens.ItemContentGap)
	outerPadding := tokens.Padding
	collapsedWidth := tokens.CollapsedWidth
	collapsedContentWidth := min(max(gtx.Dp(collapsedWidth)-gtx.Dp(outerPadding)*2, 0), width)
	indent := 0
	if !w.collapsed {
		indent = gtx.Dp(tokens.InlineIndent) * entry.depth
	}
	contentWidth := max(width-indent, 0)
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}
	childGtx.Constraints.Max.X = max(contentWidth-padding*2, 0)

	var switcher sidebarContent
	if !w.collapsed && len(item.Children) > 0 {
		switcher = recordSidebarContent(ctx, childGtx, sidebarItemSwitcher(item, entry.expanded, foreground, tokens.SwitcherSize, visual.Switcher))
	}

	var leading, initial sidebarContent
	if item.Leading != nil {
		leading = recordSidebarContent(ctx, childGtx, item.Leading)
	} else if expansion < 1 {
		initial = recordSidebarLabel(ctx, childGtx, item.Label, foreground, tokens.ItemTextSize, true, visual.Text)
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
		label = recordSidebarLabel(ctx, labelGtx, item.Label, foreground, tokens.ItemTextSize, false, visual.Text)
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

func (w Widget) layoutCollapsedSubmenu(ctx *core.Context, gtx layout.Context, entry *entry, itemVisual theme.SidebarItemVisual, selected, disabled bool, focus float32, visual theme.SidebarVisual) layout.Dimensions {
	trigger := core.WidgetFunc(func(ctx *core.Context, triggerGtx layout.Context) layout.Dimensions {
		return w.layoutItemSurface(ctx, triggerGtx, entry, itemVisual, selected, !disabled, focus, 0, visual)
	})
	popup := dropdown.New("sidebar-submenu:"+entry.item.Key, trigger, w.sidebarDropdownItems(entry.item.Children)).
		Placement(core.PopoverRightStart).
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

func sidebarItemSwitcher(item Item, expanded bool, foreground color.NRGBA, size unit.Dp, render theme.SidebarSwitcherRenderer) core.Widget {
	value := item.Switcher
	if expanded && item.ExpandedSwitcher != nil {
		value = item.ExpandedSwitcher
	}
	if value != nil {
		return value
	}
	if render == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	request := theme.SidebarSwitcherRequest{Expanded: expanded, Foreground: foreground, Size: size}
	return core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return render(gtx, request)
	})
}

type sidebarContent struct {
	call      op.CallOp
	dims      layout.Dimensions
	placement core.OverlayPlacement
}

func recordSidebarContent(ctx *core.Context, gtx layout.Context, child core.Widget) sidebarContent {
	if child == nil {
		return sidebarContent{}
	}
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackWidgetPlacement(ctx, gtx, child)
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

// recordSidebarLabel records plain text directly. Labels cannot register
// overlays, so a placement transform would only add per-item work.
func recordSidebarLabel(ctx *core.Context, gtx layout.Context, label string, foreground color.NRGBA, size unit.Sp, initial bool, render theme.SidebarTextRenderer) sidebarContent {
	if render == nil {
		return sidebarContent{}
	}
	macro := op.Record(gtx.Ops)
	dims := render(gtx, theme.SidebarTextRequest{Text: label, Color: foreground, Size: size, Initial: initial})
	return sidebarContent{call: macro.Stop(), dims: dims}
}
