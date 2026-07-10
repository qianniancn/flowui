package tabs

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

func (t TabsWidget) layout(ctx *frame.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool) layout.Dimensions {
	var listDims layout.Dimensions
	list := func(gtx layout.Context) layout.Dimensions {
		listDims = t.layoutList(ctx, gtx, state, selectedKey, disabled)
		return listDims
	}
	item, hasPanel := t.selectedItem(selectedKey)
	hasPanel = hasPanel && item.Panel != nil
	if !hasPanel {
		return list(gtx)
	}
	var panelPlacement frame.OverlayPlacement
	panel := func(gtx layout.Context) layout.Dimensions {
		dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return t.layoutPanel(ctx, gtx, item.Panel)
		})
		panelPlacement = placement
		return dims
	}
	theme := frame.ActiveTheme(ctx).Components.Tabs
	gap := max(gtx.Dp(theme.RootGap+theme.PanelGap), 0)
	if t.orientation == TabsVertical {
		dims := layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Start,
			Gap:       gap,
		}.Layout(gtx,
			layout.Rigid(list),
			layout.Flexed(1, panel),
		)
		panelPlacement.PlaceOffset(image.Pt(listDims.Size.X+gap, 0))
		return dims
	}
	dims := layout.Flex{
		Axis: layout.Vertical,
		Gap:  gap,
	}.Layout(gtx,
		layout.Rigid(list),
		layout.Rigid(panel),
	)
	panelPlacement.PlaceOffset(image.Pt(0, listDims.Size.Y+gap))
	return dims
}

func (t TabsWidget) layoutList(ctx *frame.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Tabs
	padding := gtx.Dp(theme.ListPadding)
	if t.variant == TabsSecondary {
		padding = 0
	}
	sizeStyle := tabsSizeStyleFor(frame.ActiveTheme(ctx), t.size)
	tabHeight := max(gtx.Dp(sizeStyle.height), 1)
	widths := t.tabWidths(ctx, gtx, padding, sizeStyle)
	tabGap := gtx.Dp(theme.TabGap)
	size := t.listSize(gtx, widths, tabHeight, padding, tabGap)
	listStyle := tabsListStyleFor(frame.ActiveTheme(ctx), t.variant)
	drawTabsList(gtx, frame.ActiveTheme(ctx), size, t.orientation, t.variant, listStyle)

	innerSize := image.Pt(max(size.X-padding*2, 0), max(size.Y-padding*2, 0))
	listGtx := gtx
	listGtx.Constraints = layout.Exact(innerSize)
	if disabled {
		listGtx = listGtx.Disabled()
	}
	state.list.Axis = layout.Horizontal
	state.list.Gap = 0
	if t.orientation == TabsVertical {
		state.list.Axis = layout.Vertical
		state.list.Gap = tabGap
	}
	stack := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	clipStack := clip.Rect(image.Rectangle{Max: innerSize}).Push(gtx.Ops)
	contentMacro := op.Record(gtx.Ops)
	state.list.Layout(listGtx, len(t.items), func(gtx layout.Context, index int) layout.Dimensions {
		itemSize := image.Pt(widths[index], tabHeight)
		gtx.Constraints = layout.Exact(itemSize)
		return t.layoutTab(ctx, gtx, state, sizeStyle, index, selectedKey, disabled)
	})
	content := contentMacro.Stop()
	position := state.list.Position
	if state.ensureSelectionVisible(t.items, selectedKey) {
		ctx.Invalidate()
	}
	t.drawSelectionIndicator(ctx, gtx, state, position, widths, tabHeight, tabGap, selectedKey, disabled)
	content.Add(gtx.Ops)
	clipStack.Pop()
	stack.Pop()

	previous := state.canScrollPrevious()
	next := state.canScrollNext(len(t.items))
	shadowBackground := listStyle.background
	if shadowBackground.A == 0 {
		shadowBackground = ctx.BackgroundColor()
	}
	if t.variant == TabsSecondary {
		drawTabsScrollShadows(gtx, frame.ActiveTheme(ctx), size, t.orientation, previous, next, shadowBackground)
	} else {
		radius := min(max(gtx.Dp(theme.ListRadius), 1), min(size.X, size.Y)/2)
		shadowClip := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
		drawTabsScrollShadows(gtx, frame.ActiveTheme(ctx), size, t.orientation, previous, next, shadowBackground)
		shadowClip.Pop()
	}
	if t.variant == TabsSecondary {
		drawTabsScrollShadowBorders(gtx, frame.ActiveTheme(ctx), size, t.orientation, previous, next, listStyle.border)
	}
	t.layoutScrollButtons(ctx, gtx, state, size, disabled)
	return layout.Dimensions{Size: size}
}

func (t TabsWidget) drawSelectionIndicator(ctx *frame.Context, gtx layout.Context, state *tabsState, position layout.Position, widths []int, tabHeight, tabGap int, selectedKey string, tabsDisabled bool) {
	index := tabsIndexByKey(t.items, selectedKey)
	if index < 0 {
		state.indicator.set = false
		return
	}
	target := t.tabRect(position, widths, index, tabHeight, tabGap)
	rect := state.indicator.transition(gtx, selectedKey, t.orientation, target)
	disabled := tabsDisabled || t.items[index].Disabled
	style := tabsItemStyleFor(frame.ActiveTheme(ctx), t.variant, t.color, false, disabled)
	drawTabIndicator(gtx, frame.ActiveTheme(ctx), rect, t.orientation, t.variant, style.indicator)
}

func (t TabsWidget) tabRect(position layout.Position, widths []int, index, tabHeight, tabGap int) image.Rectangle {
	gap := 0
	itemMainSize := func(item int) int {
		return widths[item]
	}
	if t.orientation == TabsVertical {
		gap = tabGap
		itemMainSize = func(int) int {
			return tabHeight
		}
	}
	start := -position.Offset
	if index >= position.First {
		for item := position.First; item < index; item++ {
			start += itemMainSize(item) + gap
		}
	} else {
		for item := position.First - 1; item >= index; item-- {
			start -= itemMainSize(item) + gap
		}
	}
	if t.orientation == TabsVertical {
		return image.Rect(0, start, widths[index], start+tabHeight)
	}
	return image.Rect(start, 0, start+widths[index], tabHeight)
}

func (t TabsWidget) tabWidths(ctx *frame.Context, gtx layout.Context, padding int, sizeStyle tabsSizeStyle) []int {
	widths := make([]int, len(t.items))
	if len(widths) == 0 {
		return widths
	}
	theme := frame.ActiveTheme(ctx).Components.Tabs
	available := max(gtx.Constraints.Max.X-padding*2, 0)
	total := 0
	for index, item := range t.items {
		width := t.measureTabWidth(ctx, gtx, item.Label, sizeStyle)
		if t.orientation == TabsVertical {
			width = max(width, gtx.Dp(theme.TabMinWidth))
		}
		widths[index] = width
		total += width
	}
	if t.orientation != TabsVertical && !t.fit && total < available {
		extra := available - total
		for index := range widths {
			share := extra / (len(widths) - index)
			widths[index] += share
			extra -= share
		}
	}
	if t.orientation == TabsVertical {
		maxWidth := 0
		for _, width := range widths {
			maxWidth = max(maxWidth, width)
		}
		maxWidth = min(maxWidth, available)
		for index := range widths {
			widths[index] = maxWidth
		}
	}
	return widths
}

func (t TabsWidget) measureTabWidth(ctx *frame.Context, gtx layout.Context, labelText string, sizeStyle tabsSizeStyle) int {
	var ops op.Ops
	measure := gtx
	measure.Ops = &ops
	measure.Constraints = layout.Constraints{Max: image.Pt(1e6, max(gtx.Constraints.Max.Y, 1))}
	label := material.Label(frame.ActiveTheme(ctx).Material, sizeStyle.textSize, labelText)
	label.Font.Weight = sizeStyle.weight
	dims := label.Layout(measure)
	return dims.Size.X + gtx.Dp(sizeStyle.paddingX)*2
}

func (t TabsWidget) listSize(gtx layout.Context, widths []int, tabHeight, padding, tabGap int) image.Point {
	if t.orientation == TabsVertical {
		width := 0
		for _, itemWidth := range widths {
			width = max(width, itemWidth)
		}
		height := tabHeight*len(widths) + max(len(widths)-1, 0)*tabGap + padding*2
		return gtx.Constraints.Constrain(image.Pt(width+padding*2, height))
	}
	height := tabHeight + padding*2
	width := gtx.Constraints.Max.X
	if t.fit {
		width = padding * 2
		for _, itemWidth := range widths {
			width += itemWidth
		}
	}
	return gtx.Constraints.Constrain(image.Pt(width, height))
}

func (t TabsWidget) layoutTab(ctx *frame.Context, gtx layout.Context, componentState *tabsState, sizeStyle tabsSizeStyle, index int, selectedKey string, tabsDisabled bool) layout.Dimensions {
	item := t.items[index]
	itemState := componentState.item(item.Key)
	disabled := tabsDisabled || item.Disabled
	selected := item.Key == selectedKey
	animGtx := gtx
	presses := state.ActivePresses(itemState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for itemState.clickable.Clicked(gtx) {
			if !selected && t.onChange != nil {
				t.onChange(item.Key)
			}
		}
		frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Min
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)

		focusVisible := itemState.interaction.Visible(gtx.Focused(&itemState.clickable), itemState.clickable.History())
		focus := itemState.interaction.Opacity(animGtx, focusVisible && !disabled)
		progress := itemState.selectionProgress(animGtx, selected)
		style := tabsItemStyleFor(frame.ActiveTheme(ctx), t.variant, t.color, itemState.clickable.Hovered() && !selected, disabled)
		if t.separators && t.variant != TabsSecondary && index > 0 && !selected && t.items[index-1].Key != selectedKey {
			drawTabSeparator(gtx, frame.ActiveTheme(ctx), size, t.orientation, style.separator)
		}
		foreground := render.LerpColor(style.foreground, style.selectedForeground, progress)
		label := material.Label(frame.ActiveTheme(ctx).Material, sizeStyle.textSize, item.Label)
		label.Font.Weight = sizeStyle.weight
		label.Color = foreground
		layout.Center.Layout(gtx, label.Layout)
		drawTabFocus(gtx, frame.ActiveTheme(ctx), size, t.variant, focus, style.focus)
		return layout.Dimensions{Size: size}
	})
}

func (t TabsWidget) layoutPanel(ctx *frame.Context, gtx layout.Context, panel frame.Widget) layout.Dimensions {
	if t.orientation != TabsVertical {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	padding := frame.ActiveTheme(ctx).Components.Tabs.PanelPadding
	var placement frame.OverlayPlacement
	dims := layout.UniformInset(padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var dims layout.Dimensions
		dims, placement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return panel.Layout(ctx, gtx)
		})
		return dims
	})
	offset := gtx.Dp(padding)
	position := image.Pt(offset, offset)
	if gtx.Constraints.Max.X-offset*2 < 0 {
		position.X = 0
	}
	if gtx.Constraints.Max.Y-offset*2 < 0 {
		position.Y = 0
	}
	placement.PlaceOffset(position)
	return dims
}

func (t TabsWidget) layoutScrollButtons(ctx *frame.Context, gtx layout.Context, state *tabsState, size image.Point, disabled bool) {
	previous := state.canScrollPrevious()
	next := state.canScrollNext(len(t.items))
	if previous {
		t.layoutScrollButton(ctx, gtx, state, &state.previous, size, -1, disabled)
	}
	if next {
		t.layoutScrollButton(ctx, gtx, state, &state.next, size, 1, disabled)
	}
}

func (t TabsWidget) layoutScrollButton(ctx *frame.Context, gtx layout.Context, state *tabsState, clickable *widget.Clickable, listSize image.Point, direction int, disabled bool) {
	theme := frame.ActiveTheme(ctx).Components.Tabs
	buttonSize := min(gtx.Dp(theme.ScrollButtonSize), min(listSize.X, listSize.Y))
	inset := gtx.Dp(theme.ScrollButtonInset)
	position := image.Pt(inset, max((listSize.Y-buttonSize)/2, 0))
	if direction > 0 {
		position.X = max(listSize.X-buttonSize-inset, 0)
	}
	if t.orientation == TabsVertical {
		position = image.Pt(max((listSize.X-buttonSize)/2, 0), inset)
		if direction > 0 {
			position.Y = max(listSize.Y-buttonSize-inset, 0)
		}
	}
	buttonGtx := gtx
	buttonGtx.Constraints = layout.Exact(image.Pt(buttonSize, buttonSize))
	if disabled {
		buttonGtx = buttonGtx.Disabled()
	} else {
		for clickable.Clicked(buttonGtx) {
			state.scrollPage(direction, len(t.items))
			ctx.Invalidate()
		}
	}
	stack := op.Offset(position).Push(gtx.Ops)
	clickable.Layout(buttonGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(t.scrollButtonLabel(direction)).Add(gtx.Ops)
		col := frame.ActiveTheme(ctx).Palette.Foreground
		if clickable.Hovered() {
			col.A = byte(float32(col.A)*0.7 + 0.5)
		}
		if disabled {
			col = frame.ActiveTheme(ctx).DisabledColor(col)
		}
		drawTabsChevron(gtx, frame.ActiveTheme(ctx), image.Pt(buttonSize, buttonSize), t.orientation, direction, col)
		return layout.Dimensions{Size: image.Pt(buttonSize, buttonSize)}
	})
	stack.Pop()
}

func (t TabsWidget) scrollButtonLabel(direction int) string {
	if t.orientation == TabsVertical {
		if direction < 0 {
			return "Scroll tabs up"
		}
		return "Scroll tabs down"
	}
	if direction < 0 {
		return "Scroll tabs left"
	}
	return "Scroll tabs right"
}
