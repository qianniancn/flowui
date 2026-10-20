package menu

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (m Widget) layoutWithVisual(ctx *core.Context, gtx layout.Context, menuState *menuState, interactive bool, visual theme.MenuVisual) layout.Dimensions {
	m.applyConstraintsWithVisual(ctx, &gtx, menuState, visual)
	children := make([]core.Widget, 0, 3)
	if m.beforeContent != nil {
		children = append(children, m.beforeContent)
	}
	children = append(children, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return m.layoutContentWithVisual(ctx, gtx, menuState, interactive, visual)
	}))
	if m.afterContent != nil {
		children = append(children, m.afterContent)
	}
	return layoutui.TrackedFlex(ctx, gtx, layout.Vertical, 0, layout.Start, children...)
}

func (m Widget) applyConstraintsWithVisual(ctx *core.Context, gtx *layout.Context, menuState *menuState, visual theme.MenuVisual) {
	metrics := visual.Metrics
	constraintMinWidth := max(gtx.Constraints.Min.X, 0)
	maxWidth := gtx.Constraints.Max.X
	if metrics.MaxWidthFraction > 0 && !m.autoWidth {
		viewport := core.OverlayViewport(ctx, gtx.Constraints.Max)
		maxWidth = min(maxWidth, int(float32(viewport.X)*min(max(metrics.MaxWidthFraction, 0), 1)))
	}
	widthPx := 0
	switch {
	case m.width > 0:
		widthPx = gtx.Dp(m.width)
	case m.autoWidth:
		widthPx = m.preferredWidthPxCachedWithVisual(ctx, *gtx, menuState, visual)
	default:
		widthPx = gtx.Dp(metrics.Width)
	}
	if m.minWidthPx > 0 {
		widthPx = max(widthPx, m.minWidthPx)
	}
	if m.hasMinWidth {
		widthPx = max(widthPx, gtx.Dp(m.minWidth))
	}
	if m.hasMaxWidth {
		maxWidth = min(maxWidth, gtx.Dp(m.maxWidth))
	}
	// A component cannot honor a narrower local maximum than the minimum
	// required by its parent. Preserve Gio's invariant instead of returning
	// dimensions outside the incoming constraint range.
	maxWidth = max(maxWidth, constraintMinWidth)
	widthPx = max(widthPx, constraintMinWidth)
	widthPx = min(max(widthPx, 0), max(maxWidth, 0))
	gtx.Constraints.Min.X = widthPx
	gtx.Constraints.Max.X = widthPx
	if metrics.MaxHeight > 0 {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(metrics.MaxHeight))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}
}

func (m Widget) layoutContentWithVisual(ctx *core.Context, gtx layout.Context, menuState *menuState, interactive bool, visual theme.MenuVisual) layout.Dimensions {
	menuState.beginFrame()
	defer menuState.endFrame()
	menuState.submenuActive = false
	entries := menuState.resolveEntries(m)
	actionable := menuState.actionableEntries(entries)
	if interactive && !m.disabled {
		result := menuState.updateKeys(gtx, m, actionable, m.nested)
		if result.focusKey != "" {
			if entry, ok := entryByKey(actionable, result.focusKey); ok {
				menuState.focus(ctx, entry, true)
				menuState.reveal(entries, result.focusKey)
			}
		}
		if result.actionKey != "" {
			if entry, ok := entryByKey(actionable, result.actionKey); ok {
				if itemHasSubmenu(entry.item) {
					menuState.openSubmenu = entry.item.Key
					menuState.submenuFocusVisible = true
				} else if m.activate(entry) && m.onRequestClose != nil {
					m.onRequestClose(true)
				}
			}
		}
		if result.openKey != "" {
			menuState.openSubmenu = result.openKey
			menuState.submenuFocusVisible = true
		}
		if result.close {
			m.closeToParent(ctx)
		}
	}

	metrics := visual.Metrics
	padding := gtx.Dp(metrics.Padding)
	innerGtx := gtx
	innerGtx.Constraints.Min = image.Pt(max(gtx.Constraints.Min.X-padding*2, 0), 0)
	innerGtx.Constraints.Max = image.Pt(max(gtx.Constraints.Max.X-padding*2, 0), max(gtx.Constraints.Max.Y-padding*2, 0))
	if len(entries) == 0 {
		return layout.UniformInset(metrics.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return m.layoutEmptyWithVisual(ctx, gtx, visual)
		})
	}

	menuState.list.Axis = layout.Vertical
	menuState.list.Alignment = layout.Start
	menuState.list.Gap = 0
	menuState.list.ScrollToEnd = false
	menuState.list.ScrollAnyAxis = false
	entrySizes := make(map[int]image.Point, menuState.list.Position.Count+2)
	entryTopGaps := make(map[int]int, menuState.list.Position.Count+2)
	contentOffset := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	content := layoutui.TrackedScrollbarWithVisualOutset(ctx, innerGtx, &menuState.list, &menuState.bar, &menuState.visualOutset, len(entries), m.disabled || !interactive, false, func(gtx layout.Context, index int) layout.Dimensions {
		entry := entries[index]
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		topGap := m.entryTopGapWithVisual(ctx, gtx, entries, index, visual)
		entryTopGaps[index] = topGap
		if topGap > 0 {
			gtx.Constraints.Min.Y = max(gtx.Constraints.Min.Y-topGap, 0)
			gtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y-topGap, 0)
		}
		offset := op.Offset(image.Pt(0, topGap)).Push(gtx.Ops)
		var dims layout.Dimensions
		switch {
		case entry.separator:
			dims = m.layoutSeparatorWithVisual(ctx, gtx, visual)
		case entry.sectionTitle != "":
			dims = m.layoutSectionTitleWithVisual(ctx, gtx, entry.sectionTitle, visual)
		default:
			dims = m.layoutItemWithVisual(ctx, gtx, menuState, entry, interactive, visual)
		}
		offset.Pop()
		if topGap > 0 {
			dims.Size.Y += topGap
		}
		entrySizes[index] = dims.Size
		return dims
	})
	contentOffset.Pop()
	y := padding - menuState.list.Position.Offset
	last := min(menuState.list.Position.First+menuState.list.Position.Count, len(entries))
	for index := menuState.list.Position.First; index < last; index++ {
		entry := entries[index]
		size := entrySizes[index]
		topGap := entryTopGaps[index]
		if !entry.separator && entry.sectionTitle == "" && size.Y > topGap {
			menuState.anchors[entry.item.Key] = image.Rect(padding, y+topGap, padding+size.X, y+size.Y)
		}
		y += size.Y
	}
	content.Size = content.Size.Add(image.Pt(padding*2, padding*2))
	content.Size = gtx.Constraints.Constrain(content.Size)
	m.registerSubmenus(ctx, gtx, menuState, interactive, visual)
	return content
}

func (m Widget) entryTopGapWithVisual(ctx *core.Context, gtx layout.Context, entries []entry, index int, visual theme.MenuVisual) int {
	if !m.entryNeedsGap(entries, index) {
		return 0
	}
	return max(gtx.Dp(visual.Metrics.ItemGap), 0)
}

func (m Widget) entryNeedsGap(entries []entry, index int) bool {
	if index <= 0 || index >= len(entries) {
		return false
	}
	previous := entries[index-1]
	current := entries[index]
	if previous.sectionIndex >= 0 && previous.sectionIndex == current.sectionIndex {
		return false
	}
	return true
}

func (m Widget) layoutEmptyWithVisual(ctx *core.Context, gtx layout.Context, visual theme.MenuVisual) layout.Dimensions {
	metrics := visual.Metrics
	height := min(gtx.Dp(metrics.ItemMinHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if visual.Text == nil {
			panic(theme.ErrIncompleteVisualProvider)
		}
		return visual.Text(gtx, theme.MenuTextRequest{Text: m.resolvedEmptyText(ctx), Color: visual.Colors.Description, Role: theme.MenuTextEmpty})
	})
}

func (m Widget) layoutSectionTitleWithVisual(ctx *core.Context, gtx layout.Context, title string, visual theme.MenuVisual) layout.Dimensions {
	metrics := visual.Metrics
	return layout.Inset{Top: metrics.SectionPaddingTop, Right: metrics.SectionPaddingX, Bottom: metrics.SectionPaddingBottom, Left: metrics.SectionPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if visual.Text == nil {
			panic(theme.ErrIncompleteVisualProvider)
		}
		return visual.Text(gtx, theme.MenuTextRequest{Text: title, Color: visual.Colors.Description, Role: theme.MenuTextSection})
	})
}

func (m Widget) layoutSeparatorWithVisual(ctx *core.Context, gtx layout.Context, visual theme.MenuVisual) layout.Dimensions {
	metrics := visual.Metrics
	marginX := gtx.Dp(metrics.SeparatorMarginX)
	marginY := gtx.Dp(metrics.SeparatorMarginY)
	width := max(gtx.Constraints.Min.X-marginX*2, 0)
	height := max(gtx.Dp(metrics.SeparatorWidth), 1)
	size := gtx.Constraints.Constrain(image.Pt(width+marginX*2, height+marginY*2))
	offset := op.Offset(image.Pt(marginX, marginY)).Push(gtx.Ops)
	visual.Separator(gtx, image.Pt(max(size.X-marginX*2, 0), min(height, size.Y)), visual.Colors.Separator)
	offset.Pop()
	return layout.Dimensions{Size: size}
}

func (m Widget) layoutItemWithVisual(ctx *core.Context, gtx layout.Context, menuState *menuState, entry entry, interactive bool, visual theme.MenuVisual) layout.Dimensions {
	item := entry.item
	itemState := menuState.item(item.Key)
	if gtx.Focused(&itemState.clickable) {
		menuState.focusedKey = item.Key
	}
	disabled := m.itemDisabled(item)
	animGtx := gtx
	eventGtx := gtx
	if disabled || !interactive {
		eventGtx = eventGtx.Disabled()
	} else {
		presses := core.ActivePresses(itemState.clickable.History())
		for itemState.clickable.Clicked(eventGtx) {
			core.RequestFocusVisible(ctx, &itemState.clickable, false)
			if itemHasSubmenu(item) {
				menuState.openSubmenu = item.Key
				menuState.submenuFocusVisible = false
			} else if m.activate(entry) && m.onRequestClose != nil {
				m.onRequestClose(false)
			}
		}
		history := itemState.clickable.History()
		if core.ActivePresses(history) > presses {
			core.RequestFocusVisible(ctx, &itemState.clickable, false)
		}
		core.FocusOnPress(ctx, &itemState.clickable, history, presses)
	}

	focused := animGtx.Focused(&itemState.clickable)
	focusVisible := menuItemFocusVisible(ctx, itemState, focused)
	selected := m.selected(entry)
	// Input ownership does not change appearance: a submenu that does not own
	// input yet, or a menu laid out inside a disabled surface, still looks
	// enabled. Only the item and menu flags disable it visually.
	styleDisabled := disabled
	styleState := style.State{
		Hovered:      !styleDisabled && itemState.clickable.Hovered(),
		Pressed:      !styleDisabled && itemState.clickable.Pressed(),
		Focused:      focused,
		FocusVisible: !styleDisabled && focusVisible,
		Disabled:     styleDisabled,
		Selected:     selected,
		Checked:      selected && (item.Kind == ItemCheckbox || item.Kind == ItemRadio),
		Open:         itemHasSubmenu(item) && menuState.openSubmenu == item.Key,
	}
	itemVisual := visual.Item(theme.MenuItemRequest{
		Variant:  uint8(item.Variant),
		Disabled: styleDisabled,
	})
	part := computed.ResolvePart(
		ctx,
		animGtx,
		core.DerivedKey(ctx, menuState.key, "item:"+item.Key),
		style.PartItem,
		styleState,
		itemVisual.Defaults,
		itemVisual.Variant,
		style.Style{},
		visual.Override,
	)
	if itemVisual.Colors == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	itemColors := itemVisual.Colors(part)
	itemState.focus.Opacity(animGtx, focusVisible && !styleDisabled, visual.Motion)

	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		class := semantic.Button
		mode, _, _ := m.selection(entry)
		if mode == SelectionMultiple {
			class = semantic.CheckBox
		} else if mode == SelectionSingle {
			class = semantic.RadioButton
		}
		class.Add(gtx.Ops)
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		if item.Description != "" {
			semantic.DescriptionOp(item.Description).Add(gtx.Ops)
		}
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		return layout.Stack{Alignment: layout.W}.Layout(gtx,
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return m.layoutItemContentWithVisual(ctx, gtx, entry, part, itemColors, visual)
			}),
		)
	})
	dims := layoutui.InteractiveResolved(ctx, eventGtx, part, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return itemState.clickable.Layout(gtx, visual)
	})
	if interactive && !disabled {
		m.updateSubmenuHover(gtx, menuState, item, itemState.clickable.Hovered(), visual.Metrics.SubmenuOpenDelay)
	}
	return dims
}

func menuItemFocusVisible(ctx *core.Context, itemState *menuItemState, focused bool) bool {
	return core.FocusVisible(ctx, &itemState.clickable, focused)
}

func (m Widget) layoutItemContentWithVisual(ctx *core.Context, gtx layout.Context, entry entry, part style.ResolvedStyle, itemColors theme.MenuItemColors, menuVisual theme.MenuVisual) layout.Dimensions {
	metrics := menuVisual.Metrics
	item := entry.item
	children := make([]layout.FlexChild, 0, 11)
	gap := func(dp unit.Dp) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Spacer{Width: dp}.Layout(gtx)
		})
	}
	if m.itemHasIndicator(entry) {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutIndicatorWithVisual(ctx, gtx, entry, itemColors, menuVisual)
		}), gap(metrics.IndicatorContentGap))
	}
	if item.Leading != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutLeadingWithVisual(ctx, gtx, item, itemColors, menuVisual)
		}), gap(metrics.ItemContentGap))
	}
	children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return m.layoutItemTextWithVisual(ctx, gtx, item, part, itemColors, menuVisual)
	}))
	if item.Shortcut != "" {
		children = append(children, gap(metrics.ItemContentGap), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutShortcutWithVisual(ctx, gtx, item.Shortcut, itemColors, menuVisual)
		}))
	}
	if item.Trailing != nil {
		children = append(children, gap(metrics.ItemContentGap), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			restore := core.PushColors(ctx, itemColors.Shortcut, ctx.BackgroundColor())
			defer restore()
			return item.Trailing.Layout(ctx, gtx)
		}))
	}
	if itemHasSubmenu(item) {
		children = append(children, gap(metrics.ItemContentGap), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if item.SubmenuIndicator != nil {
				restore := core.PushColors(ctx, itemColors.Shortcut, ctx.BackgroundColor())
				defer restore()
				return item.SubmenuIndicator.Layout(ctx, gtx)
			}
			return menuVisual.SubmenuIndicator(gtx, image.Pt(gtx.Dp(metrics.SubmenuIndicatorSize), gtx.Dp(metrics.SubmenuIndicatorSize)), itemColors.Shortcut)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (m Widget) layoutItemTextWithVisual(ctx *core.Context, gtx layout.Context, item Item, part style.ResolvedStyle, colors theme.MenuItemColors, menuVisual theme.MenuVisual) layout.Dimensions {
	if menuVisual.Text == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	if item.Description == "" {
		return menuVisual.Text(gtx, theme.MenuTextRequest{Text: item.Label, Color: colors.Foreground, Role: theme.MenuTextItem, Style: part})
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return menuVisual.Text(gtx, theme.MenuTextRequest{Text: item.Label, Color: colors.Foreground, Role: theme.MenuTextItem, Style: part})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return menuVisual.Text(gtx, theme.MenuTextRequest{Text: item.Description, Color: colors.Description, Role: theme.MenuTextDescription, Style: part})
		}),
	)
}

func (m Widget) layoutLeadingWithVisual(ctx *core.Context, gtx layout.Context, item Item, colors theme.MenuItemColors, menuVisual theme.MenuVisual) layout.Dimensions {
	foreground := colors.Shortcut
	if item.Variant == ItemDanger {
		foreground = colors.Foreground
	}
	restore := core.PushColors(ctx, foreground, ctx.BackgroundColor())
	defer restore()
	if item.Description == "" {
		return item.Leading.Layout(ctx, gtx)
	}
	metrics := menuVisual.Metrics
	height := min(gtx.Dp(metrics.DescriptionLeadingHeight), gtx.Constraints.Max.Y)
	top := min(gtx.Dp(metrics.DescriptionLeadingInsetTop), height)
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}
	childGtx.Constraints.Max.Y = max(height-top, 0)
	macro := op.Record(gtx.Ops)
	childDims := item.Leading.Layout(ctx, childGtx)
	call := macro.Stop()
	offset := op.Offset(image.Pt(0, top)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(childDims.Size.X, height))}
}

func (m Widget) layoutShortcutWithVisual(ctx *core.Context, gtx layout.Context, shortcut string, colors theme.MenuItemColors, menuVisual theme.MenuVisual) layout.Dimensions {
	metrics := menuVisual.Metrics
	height := min(gtx.Dp(metrics.ShortcutHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = height
	return layout.Inset{Left: metrics.ShortcutPaddingX, Right: metrics.ShortcutPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if menuVisual.Text == nil {
				panic(theme.ErrIncompleteVisualProvider)
			}
			return menuVisual.Text(gtx, theme.MenuTextRequest{Text: shortcut, Color: colors.Shortcut, Role: theme.MenuTextShortcut})
		})
	})
}

func (m Widget) layoutIndicatorWithVisual(ctx *core.Context, gtx layout.Context, entry entry, colors theme.MenuItemColors, visual theme.MenuVisual) layout.Dimensions {
	metrics := visual.Metrics
	restore := core.PushColors(ctx, colors.Indicator, ctx.BackgroundColor())
	defer restore()
	sizePx := gtx.Dp(metrics.IndicatorSize)
	size := image.Pt(sizePx, sizePx)
	gtx.Constraints = layout.Exact(size)
	selected := m.selected(entry)
	if entry.item.Indicator != nil {
		indicator := entry.item.Indicator(selected)
		if indicator == nil {
			return layout.Dimensions{Size: size}
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return indicator.Layout(ctx, gtx)
		})
	}
	indicatorType := entry.item.IndicatorType
	if entry.item.Kind == ItemCheckbox {
		indicatorType = IndicatorCheckmark
	} else if entry.item.Kind == ItemRadio {
		indicatorType = IndicatorDot
	}
	if !selected || indicatorType == IndicatorNone {
		return layout.Dimensions{Size: size}
	}
	offset := op.Offset(image.Pt(0, gtx.Dp(metrics.IndicatorOffsetY))).Push(gtx.Ops)
	defer offset.Pop()
	visual.Indicator(gtx, theme.MenuIndicatorRequest{
		Type:          uint8(indicatorType),
		Size:          size,
		DotSize:       gtx.Dp(metrics.RadioDotSize),
		CheckmarkSize: gtx.Dp(metrics.CheckmarkSize),
		OffsetY:       gtx.Dp(metrics.IndicatorOffsetY),
		Color:         colors.Indicator,
	})
	return layout.Dimensions{Size: size}
}

func (m Widget) itemHasIndicator(entry entry) bool {
	return entry.item.Kind == ItemCheckbox || entry.item.Kind == ItemRadio || entry.item.Indicator != nil || entry.item.IndicatorType != IndicatorNone
}
