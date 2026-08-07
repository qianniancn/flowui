package menu

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/internal/components/icon"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

func (m Widget) layout(ctx *frame.Context, gtx layout.Context, menuState *menuState, interactive bool) layout.Dimensions {
	m.applyConstraints(ctx, &gtx, menuState)
	children := make([]frame.Widget, 0, 3)
	if m.beforeContent != nil {
		children = append(children, m.beforeContent)
	}
	children = append(children, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return m.layoutContent(ctx, gtx, menuState, interactive)
	}))
	if m.afterContent != nil {
		children = append(children, m.afterContent)
	}
	return layoutui.LayoutTrackedFlex(ctx, gtx, layout.Vertical, 0, layout.Start, children...)
}

func (m Widget) applyConstraints(ctx *frame.Context, gtx *layout.Context, menuState *menuState) {
	tokens := m.themeTokens(ctx)
	constraintMinWidth := max(gtx.Constraints.Min.X, 0)
	maxWidth := gtx.Constraints.Max.X
	if tokens.MaxWidthFraction > 0 {
		viewport := frame.OverlayViewport(ctx, gtx.Constraints.Max)
		maxWidth = min(maxWidth, int(float32(viewport.X)*min(max(tokens.MaxWidthFraction, 0), 1)))
	}
	widthPx := 0
	switch {
	case m.width > 0:
		widthPx = gtx.Dp(m.width)
	case m.autoWidth:
		widthPx = m.preferredWidthPxCached(ctx, *gtx, menuState)
	default:
		widthPx = gtx.Dp(tokens.Width)
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
	if tokens.MaxHeight > 0 {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(tokens.MaxHeight))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}
}

func (m Widget) layoutContent(ctx *frame.Context, gtx layout.Context, menuState *menuState, interactive bool) layout.Dimensions {
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

	tokens := m.themeTokens(ctx)
	padding := gtx.Dp(tokens.Padding)
	innerGtx := gtx
	innerGtx.Constraints.Min = image.Pt(max(gtx.Constraints.Min.X-padding*2, 0), 0)
	innerGtx.Constraints.Max = image.Pt(max(gtx.Constraints.Max.X-padding*2, 0), max(gtx.Constraints.Max.Y-padding*2, 0))
	if len(entries) == 0 {
		return layout.UniformInset(tokens.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return m.layoutEmpty(ctx, gtx)
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
	content := layoutui.LayoutTrackedScrollbarWithVisualOutset(ctx, innerGtx, &menuState.list, &menuState.bar, &menuState.visualOutset, len(entries), m.disabled || !interactive, false, func(gtx layout.Context, index int) layout.Dimensions {
		entry := entries[index]
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		topGap := m.entryTopGap(ctx, gtx, entries, index)
		entryTopGaps[index] = topGap
		if topGap > 0 {
			gtx.Constraints.Min.Y = max(gtx.Constraints.Min.Y-topGap, 0)
			gtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y-topGap, 0)
		}
		offset := op.Offset(image.Pt(0, topGap)).Push(gtx.Ops)
		var dims layout.Dimensions
		switch {
		case entry.separator:
			dims = m.layoutSeparator(ctx, gtx)
		case entry.sectionTitle != "":
			dims = m.layoutSectionTitle(ctx, gtx, entry.sectionTitle)
		default:
			dims = m.layoutItem(ctx, gtx, menuState, entry, interactive)
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
	m.registerSubmenus(ctx, gtx, menuState, interactive)
	return content
}

func (m Widget) entryTopGap(ctx *frame.Context, gtx layout.Context, entries []entry, index int) int {
	if !m.entryNeedsGap(entries, index) {
		return 0
	}
	return max(gtx.Dp(m.themeTokens(ctx).ItemGap), 0)
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

func (m Widget) layoutEmpty(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := m.themeTokens(ctx)
	height := min(gtx.Dp(tokens.ItemMinHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return text.New(m.emptyText).Size(float32(tokens.ItemTextSize)).Color(menuMutedColor(frame.ActiveTheme(ctx))).Layout(ctx, gtx)
	})
}

func (m Widget) layoutSectionTitle(ctx *frame.Context, gtx layout.Context, title string) layout.Dimensions {
	tokens := m.themeTokens(ctx)
	return layout.Inset{Top: tokens.SectionPaddingTop, Right: tokens.SectionPaddingX, Bottom: tokens.SectionPaddingBottom, Left: tokens.SectionPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return text.New(title).Size(float32(tokens.SectionTextSize)).Weight(font.Medium).Color(menuMutedColor(frame.ActiveTheme(ctx))).Layout(ctx, gtx)
	})
}

func (m Widget) layoutSeparator(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := m.themeTokens(ctx)
	marginX := gtx.Dp(tokens.SeparatorMarginX)
	marginY := gtx.Dp(tokens.SeparatorMarginY)
	width := max(gtx.Constraints.Min.X-marginX*2, 0)
	height := max(gtx.Dp(tokens.SeparatorWidth), 1)
	size := gtx.Constraints.Constrain(image.Pt(width+marginX*2, height+marginY*2))
	offset := op.Offset(image.Pt(marginX, marginY)).Push(gtx.Ops)
	drawMenuSeparator(gtx, image.Pt(max(size.X-marginX*2, 0), min(height, size.Y)), frame.ActiveTheme(ctx).Palette.SeparatorColor())
	offset.Pop()
	return layout.Dimensions{Size: size}
}

func (m Widget) layoutItem(ctx *frame.Context, gtx layout.Context, menuState *menuState, entry entry, interactive bool) layout.Dimensions {
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
		presses := stateutil.ActivePresses(itemState.clickable.History())
		for itemState.clickable.Clicked(eventGtx) {
			frame.RequestFocusVisible(ctx, &itemState.clickable, false)
			if itemHasSubmenu(item) {
				menuState.openSubmenu = item.Key
				menuState.submenuFocusVisible = false
			} else if m.activate(entry) && m.onRequestClose != nil {
				m.onRequestClose(false)
			}
		}
		history := itemState.clickable.History()
		if stateutil.ActivePresses(history) > presses {
			frame.RequestFocusVisible(ctx, &itemState.clickable, false)
		}
		frame.FocusOnPress(ctx, &itemState.clickable, history, presses)
	}

	focused := animGtx.Focused(&itemState.clickable)
	focusVisible := menuItemFocusVisible(ctx, itemState, focused)
	selected := m.selected(entry)
	styleDisabled := disabled || !interactive
	styleState := flowstyle.StyleState{
		Hovered:      !styleDisabled && itemState.clickable.Hovered(),
		Pressed:      !styleDisabled && itemState.clickable.Pressed(),
		Focused:      focused,
		FocusVisible: !styleDisabled && focusVisible,
		Disabled:     styleDisabled,
		Selected:     selected,
		Checked:      selected && (item.Kind == ItemCheckbox || item.Kind == ItemRadio),
		Open:         itemHasSubmenu(item) && menuState.openSubmenu == item.Key,
	}
	part := styleruntime.ResolvePart(
		ctx,
		animGtx,
		frame.DerivedKey(ctx, menuState.key, "item:"+item.Key),
		flowstyle.PartItem,
		styleState,
		menuItemDefaultDeclaration(frame.ActiveTheme(ctx), m.themeTokens(ctx)),
		menuItemVariantDeclaration(frame.ActiveTheme(ctx), item.Variant),
		flowstyle.Style{},
		m.customStyle,
	)
	style := menuItemStyle(frame.ActiveTheme(ctx), item.Variant)
	if part.Text != nil {
		if col, ok := styleruntime.Color(part.Text.Color); ok {
			style.foreground = col
			style.description = col
			style.shortcut = col
		}
		if part.Text.FontWeight != nil {
			style.fontWeight = font.Weight(*part.Text.FontWeight)
		}
	}
	itemState.focus.Opacity(animGtx, focusVisible && !styleDisabled, frame.ActiveTheme(ctx).Motion)

	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
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
				return m.layoutItemContent(ctx, gtx, entry, style)
			}),
		)
	})
	dims := layoutui.LayoutInteractiveResolved(ctx, eventGtx, part, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return itemState.clickable.Layout(gtx, visual)
	})
	if interactive && !disabled {
		m.updateSubmenuHover(gtx, menuState, item, itemState.clickable.Hovered())
	}
	return dims
}

func menuItemFocusVisible(ctx *frame.Context, itemState *menuItemState, focused bool) bool {
	return frame.FocusVisible(ctx, &itemState.clickable, focused)
}

func (m Widget) layoutItemContent(ctx *frame.Context, gtx layout.Context, entry entry, style itemStyle) layout.Dimensions {
	tokens := m.themeTokens(ctx)
	item := entry.item
	children := make([]layout.FlexChild, 0, 11)
	gap := func(dp unit.Dp) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Spacer{Width: dp}.Layout(gtx)
		})
	}
	if m.itemHasIndicator(entry) {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutIndicator(ctx, gtx, entry, style)
		}), gap(tokens.IndicatorContentGap))
	}
	if item.Leading != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutLeading(ctx, gtx, item, style)
		}), gap(tokens.ItemContentGap))
	}
	children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return m.layoutItemText(ctx, gtx, item, style)
	}))
	if item.Shortcut != "" {
		children = append(children, gap(tokens.ItemContentGap), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutShortcut(ctx, gtx, item.Shortcut, style)
		}))
	}
	if item.Trailing != nil {
		children = append(children, gap(tokens.ItemContentGap), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			restore := frame.PushColors(ctx, style.shortcut, ctx.BackgroundColor())
			defer restore()
			return item.Trailing.Layout(ctx, gtx)
		}))
	}
	if itemHasSubmenu(item) {
		children = append(children, gap(tokens.ItemContentGap), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if item.SubmenuIndicator != nil {
				restore := frame.PushColors(ctx, style.shortcut, ctx.BackgroundColor())
				defer restore()
				return item.SubmenuIndicator.Layout(ctx, gtx)
			}
			return icon.New(lucide.ChevronRight).Size(float32(tokens.SubmenuIndicatorSize)).Color(style.shortcut).Layout(ctx, gtx)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (m Widget) layoutItemText(ctx *frame.Context, gtx layout.Context, item Item, style itemStyle) layout.Dimensions {
	tokens := m.themeTokens(ctx)
	if item.Description == "" {
		return text.New(item.Label).Size(float32(tokens.ItemTextSize)).Weight(style.fontWeight).Color(style.foreground).Layout(ctx, gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(item.Label).Size(float32(tokens.ItemTextSize)).Weight(style.fontWeight).Color(style.foreground).Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(item.Description).Size(float32(tokens.ItemDescriptionSize)).Color(style.description).Layout(ctx, gtx)
		}),
	)
}

func (m Widget) layoutLeading(ctx *frame.Context, gtx layout.Context, item Item, style itemStyle) layout.Dimensions {
	foreground := style.shortcut
	if item.Variant == ItemDanger {
		foreground = style.foreground
	}
	restore := frame.PushColors(ctx, foreground, ctx.BackgroundColor())
	defer restore()
	if item.Description == "" {
		return item.Leading.Layout(ctx, gtx)
	}
	tokens := m.themeTokens(ctx)
	height := min(gtx.Dp(tokens.DescriptionLeadingHeight), gtx.Constraints.Max.Y)
	top := min(gtx.Dp(tokens.DescriptionLeadingInsetTop), height)
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

func (m Widget) layoutShortcut(ctx *frame.Context, gtx layout.Context, shortcut string, style itemStyle) layout.Dimensions {
	tokens := m.themeTokens(ctx)
	height := min(gtx.Dp(tokens.ShortcutHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = height
	return layout.Inset{Left: tokens.ShortcutPaddingX, Right: tokens.ShortcutPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return text.New(shortcut).Size(float32(tokens.ShortcutTextSize)).Weight(font.Medium).Color(style.shortcut).Layout(ctx, gtx)
		})
	})
}

func (m Widget) layoutIndicator(ctx *frame.Context, gtx layout.Context, entry entry, style itemStyle) layout.Dimensions {
	tokens := m.themeTokens(ctx)
	restore := frame.PushColors(ctx, style.indicator, ctx.BackgroundColor())
	defer restore()
	sizePx := gtx.Dp(tokens.IndicatorSize)
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
	offset := op.Offset(image.Pt(0, gtx.Dp(tokens.IndicatorOffsetY))).Push(gtx.Ops)
	defer offset.Pop()
	if indicatorType == IndicatorDot {
		drawMenuDot(gtx, size, gtx.Dp(tokens.RadioDotSize), style.indicator)
		return layout.Dimensions{Size: size}
	}
	iconSize := tokens.CheckmarkSize
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints = layout.Exact(image.Pt(gtx.Dp(iconSize), gtx.Dp(iconSize)))
		return icon.New(lucide.Check).Size(float32(iconSize)).Color(style.indicator).Layout(ctx, gtx)
	})
}

func (m Widget) itemHasIndicator(entry entry) bool {
	return entry.item.Kind == ItemCheckbox || entry.item.Kind == ItemRadio || entry.item.Indicator != nil || entry.item.IndicatorType != IndicatorNone
}
