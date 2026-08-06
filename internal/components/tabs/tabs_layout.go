package tabs

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui-icons-lucide"
	iconui "github.com/qianniancn/flowui/internal/components/icon"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	textui "github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

func (t TabsWidget) layout(ctx *frame.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool) layout.Dimensions {
	if t.panelsForceRender() {
		t.layoutForcedPanels(ctx, gtx, state, selectedKey)
	}
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
		opacity := state.panelOpacityFor(gtx, selectedKey, t.panelTransition, t.panelAnimationDuration(ctx), frame.ActiveTheme(ctx).Motion)
		fade := paint.PushOpacity(gtx.Ops, opacity)
		defer fade.Pop()
		dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return t.layoutPanel(ctx, gtx, item)
		})
		panelPlacement = placement
		return dims
	}
	theme := frame.ActiveTheme(ctx).Components.Tabs
	gap := max(gtx.Dp(theme.RootGap+theme.PanelGap), 0)
	alignment := layout.Start
	if t.centered && t.orientation == TabsHorizontal {
		alignment = layout.Middle
	}
	if t.orientation == TabsVertical {
		if t.placement == TabsEnd {
			dims := layout.Flex{Axis: layout.Horizontal, Alignment: alignment, Gap: gap}.Layout(gtx,
				layout.Flexed(1, panel),
				layout.Rigid(list),
			)
			panelPlacement.PlaceOffset(image.Pt(0, 0))
			return dims
		}
		dims := layout.Flex{Axis: layout.Horizontal, Alignment: alignment, Gap: gap}.Layout(gtx,
			layout.Rigid(list),
			layout.Flexed(1, panel),
		)
		panelPlacement.PlaceOffset(image.Pt(listDims.Size.X+gap, 0))
		return dims
	}
	if t.placement == TabsBottom {
		dims := layout.Flex{Axis: layout.Vertical, Alignment: alignment, Gap: gap}.Layout(gtx,
			layout.Rigid(panel),
			layout.Rigid(list),
		)
		panelPlacement.PlaceOffset(image.Pt(0, 0))
		return dims
	}
	dims := layout.Flex{Axis: layout.Vertical, Alignment: alignment, Gap: gap}.Layout(gtx,
		layout.Rigid(list),
		layout.Rigid(panel),
	)
	panelPlacement.PlaceOffset(image.Pt(0, listDims.Size.Y+gap))
	return dims
}

// layoutForcedPanels performs a one-time hidden layout for inactive panels.
// The frame helper gives each panel a disabled input source and a private ops
// stream, so this pass can initialize component state without adding visible
// paint, semantics, focus targets, or overlays to the current frame.
func (t TabsWidget) layoutForcedPanels(ctx *frame.Context, gtx layout.Context, state *tabsState, selectedKey string) {
	for _, item := range t.items {
		if item.Key == selectedKey || item.Panel == nil {
			continue
		}
		if _, rendered := state.renderedPanels[item.Key]; rendered {
			continue
		}
		scope := t.panelStateScope(ctx, item.Key)
		frame.LayoutHidden(ctx, gtx, scope, frame.WidgetFunc(func(ctx *frame.Context, hiddenGtx layout.Context) layout.Dimensions {
			return t.layoutPanel(ctx, hiddenGtx, item)
		}))
		state.renderedPanels[item.Key] = struct{}{}
	}
}

func (t TabsWidget) layoutList(ctx *frame.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool) layout.Dimensions {
	tabStrip := func(gtx layout.Context) layout.Dimensions {
		return t.layoutTabStrip(ctx, gtx, state, selectedKey, disabled)
	}
	if t.centered && t.orientation == TabsHorizontal {
		// Keep the strip's intrinsic width while centering it in the available
		// bar area. Leading/trailing slots are handled by the surrounding flex,
		// so this centers only within their remaining space.
		centeredStrip := tabStrip
		tabStrip = func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, centeredStrip)
		}
	}
	if t.leading == nil && t.trailing == nil && t.onAdd == nil {
		return tabStrip(gtx)
	}

	theme := frame.ActiveTheme(ctx).Components.Tabs
	children := make([]layout.FlexChild, 0, 4)
	if t.leading != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutSlot(ctx, gtx, "leading", t.leading)
		}))
	}
	stripChild := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return tabStrip(gtx)
	})
	if t.usesIntrinsicTabWidths() {
		stripChild = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return tabStrip(gtx)
		})
	}
	children = append(children, stripChild)
	if t.trailing != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutSlot(ctx, gtx, "trailing", t.trailing)
		}))
	}
	if t.onAdd != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutAddButton(ctx, gtx, state, disabled)
		}))
	}
	axis := layout.Horizontal
	if t.orientation == TabsVertical {
		axis = layout.Vertical
	}
	return layout.Flex{Axis: axis, Alignment: layout.Middle, Gap: gtx.Dp(theme.ExtraContentGap)}.Layout(gtx, children...)
}

func (t TabsWidget) layoutSlot(ctx *frame.Context, gtx layout.Context, role string, content frame.Widget) layout.Dimensions {
	if content == nil {
		return layout.Dimensions{}
	}
	return tabsSlotWidget{hostKey: t.key, role: role, content: content}.Layout(ctx, gtx)
}

func (t TabsWidget) slotWidget(role string, content frame.Widget) frame.Widget {
	if content == nil {
		return nil
	}
	return tabsSlotWidget{hostKey: t.key, role: role, content: content}
}

type tabsSlotWidget struct {
	hostKey string
	role    string
	content frame.Widget
}

func (w tabsSlotWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	restoreHost := frame.PushKey(ctx, w.hostKey)
	defer restoreHost()
	restoreRole := frame.PushKey(ctx, w.role)
	defer restoreRole()
	return w.content.Layout(ctx, gtx)
}

func (w tabsSlotWidget) Measure(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	restoreHost := frame.PushKey(ctx, w.hostKey)
	defer restoreHost()
	restoreRole := frame.PushKey(ctx, w.role)
	defer restoreRole()
	return frame.MeasureWidget(ctx, gtx, w.content)
}

func (t TabsWidget) layoutTabStripDefault(ctx *frame.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Tabs
	variant := t.variant
	padding := gtx.Dp(theme.ListPadding)
	if variant == TabsSecondary {
		padding = 0
	}
	sizeStyle := tabsSizeStyleFor(frame.ActiveTheme(ctx), t.size)
	tabHeight := max(gtx.Dp(sizeStyle.height), 1)
	widths := t.tabWidths(ctx, gtx, padding, sizeStyle)
	tabGap := gtx.Dp(theme.TabGap)
	size := t.listSize(gtx, widths, tabHeight, padding, tabGap)
	listStyle := tabsListStyleFor(frame.ActiveTheme(ctx), variant)
	drawTabsList(gtx, frame.ActiveTheme(ctx), size, t.orientation, variant, listStyle)

	innerSize := image.Pt(max(size.X-padding*2, 0), max(size.Y-padding*2, 0))
	listGtx := gtx
	listGtx.Constraints = layout.Exact(innerSize)
	if disabled {
		listGtx = listGtx.Disabled()
	}
	if state.noteListLayout(innerSize, t.orientation, widths, len(t.items)) {
		state.selectionPending = true
	}
	state.list.Axis = layout.Horizontal
	state.list.Gap = tabGap
	if t.orientation == TabsVertical {
		state.list.Axis = layout.Vertical
	}
	stack := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	clipStack := clip.Rect(image.Rectangle{Max: innerSize}).Push(gtx.Ops)
	contentMacro := op.Record(gtx.Ops)
	semantic.DescriptionOp("Tab list").Add(gtx.Ops)
	listSemanticKey := frame.DerivedKey(ctx, frame.FullKey(ctx, t.key), "tab-list")
	frame.RegisterSemantic(ctx, frame.SemanticNode{
		Key:   listSemanticKey,
		Role:  frame.SemanticTabList,
		Label: "Tab list",
	})
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
	if variant == TabsSecondary {
		drawTabsScrollShadows(gtx, frame.ActiveTheme(ctx), size, t.orientation, previous, next, shadowBackground)
	} else {
		radiusToken := theme.ListRadius
		radius := min(max(gtx.Dp(radiusToken), 1), min(size.X, size.Y)/2)
		shadowClip := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
		drawTabsScrollShadows(gtx, frame.ActiveTheme(ctx), size, t.orientation, previous, next, shadowBackground)
		shadowClip.Pop()
	}
	if variant == TabsSecondary {
		drawTabsScrollShadowBorders(gtx, frame.ActiveTheme(ctx), size, t.orientation, previous, next, listStyle.border)
	}
	t.layoutScrollButtons(ctx, gtx, state, size, disabled)
	return layout.Dimensions{Size: size}
}

func (t TabsWidget) drawSelectionIndicator(ctx *frame.Context, gtx layout.Context, state *tabsState, position layout.Position, widths []int, tabHeight, tabGap int, selectedKey string, tabsDisabled bool) {
	if t.indicatorVisibleSet && !t.indicatorVisible {
		state.indicator.set = false
		return
	}
	index := tabsIndexByKey(t.items, selectedKey)
	if index < 0 {
		state.indicator.set = false
		return
	}
	target := t.tabRect(position, widths, index, tabHeight, tabGap)
	target = t.indicatorTarget(ctx, gtx, target)
	rect := state.indicator.transitionWithDuration(gtx, selectedKey, t.orientation, target, t.indicatorAnimationDuration(ctx), frame.ActiveTheme(ctx).Motion)
	disabled := tabsDisabled || t.items[index].Disabled
	style := tabsItemStyleFor(frame.ActiveTheme(ctx), t.variant, t.color, false, disabled)
	drawTabIndicator(gtx, frame.ActiveTheme(ctx), rect, t.orientation, t.variant, style.indicator)
}

func (t TabsWidget) indicatorTarget(ctx *frame.Context, gtx layout.Context, target image.Rectangle) image.Rectangle {
	tokens := frame.ActiveTheme(ctx).Components.Tabs
	width := t.indicatorWidth
	if width <= 0 {
		width = tokens.IndicatorWidth
	}
	// Align against the measured tab node. Its padding is part of the tab's
	// interactive and visual slot, so start/center/end use the same bounds as
	// the tab item rather than a separately reconstructed content rectangle.
	mainLength := target.Dx()
	if t.orientation == TabsVertical {
		mainLength = target.Dy()
	}
	inset := 0
	minLength := 0
	variant := t.variant
	if variant == TabsSecondary {
		inset = gtx.Dp(tokens.IndicatorInset)
		minLength = gtx.Dp(tokens.IndicatorMinWidth)
	}
	mainLength = min(max(mainLength-inset*2, minLength), mainLength)
	if width > 0 {
		mainLength = min(mainLength, gtx.Dp(width))
	}
	mainLength = max(min(mainLength, max(target.Dx(), target.Dy())), 0)
	align := t.indicatorAlign
	if !t.indicatorAlignSet && variant == TabsSecondary {
		align = TabsIndicatorCenter
	}
	if t.orientation == TabsVertical {
		offset := 0
		switch align {
		case TabsIndicatorCenter:
			offset = (target.Dy() - mainLength) / 2
		case TabsIndicatorEnd:
			offset = target.Dy() - mainLength
		}
		target.Min.Y += offset
		target.Max.Y = target.Min.Y + mainLength
		if variant == TabsSecondary {
			lineWidth := min(max(gtx.Dp(tokens.IndicatorLineWidth), 1), max(target.Dx(), 1))
			if t.placement == TabsEnd {
				target.Min.X = target.Max.X - lineWidth
				target.Max.X = target.Min.X + lineWidth
			}
		}
		return target
	}
	offset := 0
	switch align {
	case TabsIndicatorCenter:
		offset = (target.Dx() - mainLength) / 2
	case TabsIndicatorEnd:
		offset = target.Dx() - mainLength
	}
	target.Min.X += offset
	target.Max.X = target.Min.X + mainLength
	if variant == TabsSecondary && t.placement == TabsBottom {
		lineWidth := min(max(gtx.Dp(tokens.IndicatorLineWidth), 1), max(target.Dy(), 1))
		target.Max.Y = target.Min.Y + lineWidth
	}
	return target
}

func (t TabsWidget) tabRect(position layout.Position, widths []int, index, tabHeight, tabGap int) image.Rectangle {
	gap := tabGap
	itemMainSize := func(item int) int {
		return widths[item]
	}
	if t.orientation == TabsVertical {
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
	tabGap := gtx.Dp(theme.TabGap)
	total := 0
	for index, item := range t.items {
		width := t.measureTabWidth(ctx, gtx, item, sizeStyle)
		if t.orientation == TabsVertical {
			width = max(width, gtx.Dp(theme.TabMinWidth))
		}
		widths[index] = width
		total += width
	}
	if t.orientation != TabsVertical && !t.usesIntrinsicTabWidths() {
		gapTotal := max(len(widths)-1, 0) * tabGap
		extra := available - total - gapTotal
		if extra <= 0 {
			return widths
		}
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

func (t TabsWidget) measureTabWidth(ctx *frame.Context, gtx layout.Context, item TabItem, sizeStyle tabsSizeStyle) int {
	measure := gtx
	measure.Constraints = layout.Constraints{Max: image.Pt(1e6, max(gtx.Constraints.Max.Y, 1))}
	leftPadding := gtx.Dp(sizeStyle.paddingX)
	rightPadding := gtx.Dp(t.tabRightPadding(ctx, sizeStyle, item))
	width := 0
	if item.Content != nil {
		width = t.measureTabContent(ctx, measure, item.Key, item.Content).Size.X
	} else {
		label := textui.New(item.Label).Size(float32(sizeStyle.textSize)).Weight(sizeStyle.weight)
		dims := textui.Measure(ctx, measure, label)
		width = dims.Size.X
	}
	theme := frame.ActiveTheme(ctx).Components.Tabs
	if len(item.Icon) > 0 {
		width += gtx.Dp(theme.IconSize) + gtx.Dp(theme.IconGap)
	}
	for _, leading := range []frame.Widget{item.Leading, item.Trailing} {
		if leading == nil {
			continue
		}
		child := t.measureTabContent(ctx, measure, item.Key, leading)
		width += child.Size.X + gtx.Dp(theme.IconGap)
	}
	if item.Closable && t.onClose != nil {
		closeSize, closeGap := t.closeButtonMetrics(ctx)
		width += gtx.Dp(closeSize) + gtx.Dp(closeGap)
	}
	return width + leftPadding + rightPadding
}

// tabRightPadding keeps a closable tab's close slot inside the normal right
// padding whenever there is enough room. This avoids adding a full padding
// allowance and then adding the hidden close slot on top of it.
func (t TabsWidget) tabRightPadding(ctx *frame.Context, sizeStyle tabsSizeStyle, item TabItem) unit.Dp {
	right := sizeStyle.paddingX
	if item.Closable && t.onClose != nil {
		closeSize, closeGap := t.closeButtonMetrics(ctx)
		slot := closeSize + closeGap
		right = max(right-slot, 0)
	}
	return right
}

func (t TabsWidget) measureTabContent(ctx *frame.Context, gtx layout.Context, itemKey string, content frame.Widget) layout.Dimensions {
	restoreHost := frame.PushKey(ctx, t.key)
	defer restoreHost()
	restoreItem := frame.PushKey(ctx, itemKey)
	defer restoreItem()
	if measurable, ok := content.(frame.Measurable); ok {
		return measurable.Measure(ctx, gtx)
	}
	return frame.MeasureWidget(ctx, gtx, content)
}

func (t TabsWidget) layoutTabContent(ctx *frame.Context, gtx layout.Context, itemKey string, content frame.Widget) layout.Dimensions {
	if content == nil {
		return layout.Dimensions{}
	}
	restoreHost := frame.PushKey(ctx, t.key)
	defer restoreHost()
	restoreItem := frame.PushKey(ctx, itemKey)
	defer restoreItem()
	return content.Layout(ctx, gtx)
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
	if t.usesIntrinsicTabWidths() {
		width = padding * 2
		for _, itemWidth := range widths {
			width += itemWidth
		}
		width += max(len(widths)-1, 0) * tabGap
	}
	return gtx.Constraints.Constrain(image.Pt(width, height))
}

func (t TabsWidget) layoutTab(ctx *frame.Context, gtx layout.Context, componentState *tabsState, sizeStyle tabsSizeStyle, index int, selectedKey string, tabsDisabled bool) layout.Dimensions {
	item := t.items[index]
	itemState := componentState.item(item.Key)
	disabled := tabsDisabled || item.Disabled
	selected := item.Key == selectedKey
	if componentState.editingKey != item.Key {
		itemState.editReady = false
		itemState.editFocused = false
	}
	editing := t.itemIsEditing(componentState, item)
	animGtx := gtx
	presses := state.ActivePresses(itemState.clickable.History())
	closeClicked := false
	if !disabled && !editing && item.Editable {
		for {
			e, ok := gtx.Event(key.Filter{Focus: &itemState.clickable, Name: key.NameF2})
			if !ok {
				break
			}
			if event, ok := e.(key.Event); ok && event.State == key.Press {
				componentState.editingKey = item.Key
				itemState.editReady = false
				itemState.editFocused = false
				editing = true
				frame.RequestFocus(ctx, &itemState.editor)
			}
		}
	}
	if disabled {
		gtx = gtx.Disabled()
	} else if !editing {
		if item.Closable && t.onClose != nil && t.closeVisible(itemState) {
			for itemState.close.Clicked(gtx) {
				closeClicked = true
				frame.PreserveFocus(ctx)
				t.onClose(item.Key)
				if selected {
					if fallback, ok := tabsCloseFallback(t.items, index); ok {
						selectedKey = componentState.requestSelectedKey(t, t.items[fallback].Key)
						if t.hasSelectedKey {
							// Controlled bindings return the old value until the
							// caller feeds the request back. Keep the current frame
							// usable while preserving that contract.
							selectedKey = t.items[fallback].Key
						}
						componentState.selectionPending = true
					} else {
						selectedKey = componentState.requestSelectedKey(t, "")
						if t.hasSelectedKey {
							selectedKey = ""
						}
						componentState.selectionPending = false
					}
				}
			}
		}
		for itemState.clickable.Clicked(gtx) {
			if !closeClicked && !selected {
				// Route pointer activation through the same disclosure binding as
				// keyboard activation so uncontrolled Tabs update their own state.
				componentState.requestSelectedKey(t, item.Key)
			}
		}
		frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Min
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(t.itemAccessibleLabel(item)).Add(gtx.Ops)
		description := item.Description
		if description == "" {
			description = "Tab"
			if selected {
				description = "Selected tab"
			}
		}
		semantic.DescriptionOp(description).Add(gtx.Ops)
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		frame.RegisterSemantic(ctx, frame.SemanticNode{
			Key:         frame.DerivedKey(ctx, frame.FullKey(ctx, t.key), "tab:"+item.Key),
			Role:        frame.SemanticTab,
			Label:       t.itemAccessibleLabel(item),
			Description: description,
			Controls:    frame.DerivedKey(ctx, frame.FullKey(ctx, t.key), "panel:"+item.Key),
			Owner:       frame.DerivedKey(ctx, frame.FullKey(ctx, t.key), "tab-list"),
			Selected:    selected,
			Disabled:    disabled,
			PosInSet:    index + 1,
			SetSize:     len(t.items),
		})

		focusVisible := frame.FocusVisible(ctx, &itemState.clickable, gtx.Focused(&itemState.clickable))
		motion := frame.ActiveTheme(ctx).Motion
		focus := itemState.interaction.Opacity(animGtx, focusVisible && !disabled, motion)
		progress := itemState.selectionProgressWithDuration(animGtx, selected, t.selectionDuration(ctx), motion)
		variant := t.variant
		style := tabsItemStyleFor(frame.ActiveTheme(ctx), variant, t.color, itemState.clickable.Hovered() && !selected, disabled)
		if t.separators && variant != TabsSecondary && index > 0 && !selected && t.items[index-1].Key != selectedKey {
			drawTabSeparator(gtx, frame.ActiveTheme(ctx), size, t.orientation, style.separator)
		}
		foreground := render.LerpColor(style.foreground, style.selectedForeground, progress)
		styleState := flowstyle.StyleState{
			Hovered: itemState.clickable.Hovered(), Pressed: itemState.clickable.Pressed(),
			Focused: gtx.Focused(&itemState.clickable), FocusVisible: focusVisible,
			Disabled: disabled, Selected: selected,
		}
		resolved := t.resolveItemStyle(ctx, animGtx, item.Key, styleState, foreground, sizeStyle)
		if editing {
			t.prepareTabEditor(ctx, gtx, componentState, item, itemState)
		}
		return layoutui.LayoutResolved(ctx, gtx, resolved, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
			textSize := sizeStyle.textSize
			weight := sizeStyle.weight
			color := foreground
			if resolved.Text != nil {
				if resolved.Text.FontSize != nil {
					textSize = *resolved.Text.FontSize
				}
				if resolved.Text.FontWeight != nil {
					weight = font.Weight(*resolved.Text.FontWeight)
				}
				if value, ok := styleruntime.Color(resolved.Text.Color); ok {
					color = value
				}
			}
			children := make([]layout.FlexChild, 0, 4)
			if len(item.Icon) > 0 {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iconui.New(item.Icon).Size(float32(frame.ActiveTheme(ctx).Components.Tabs.IconSize)).Layout(ctx, gtx)
				}))
			}
			if item.Leading != nil {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.layoutTabContent(ctx, gtx, item.Key, item.Leading)
				}))
			}
			labelContent := func(gtx layout.Context) layout.Dimensions {
				if editing {
					semantic.Editor.Add(gtx.Ops)
					semantic.LabelOp(t.itemAccessibleLabel(item)).Add(gtx.Ops)
					return material.Editor(frame.ActiveMaterial(ctx), &itemState.editor, "").Layout(gtx)
				}
				if item.Content != nil {
					return t.layoutTabContent(ctx, gtx, item.Key, item.Content)
				}
				label := material.Label(frame.ActiveMaterial(ctx), textSize, item.Label)
				label.Font.Weight = weight
				label.Color = color
				return layout.Center.Layout(gtx, label.Layout)
			}
			if t.usesIntrinsicTabWidths() {
				children = append(children, layout.Rigid(labelContent))
			} else {
				children = append(children, layout.Flexed(1, labelContent))
			}
			if item.Trailing != nil {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.layoutTabContent(ctx, gtx, item.Key, item.Trailing)
				}))
			}
			theme := frame.ActiveTheme(ctx).Components.Tabs
			contentLayout := func(contentGtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: contentGtx.Dp(theme.IconGap)}.Layout(contentGtx, children...)
			}
			layoutContent := contentLayout
			if item.Closable && t.onClose != nil {
				closeSize, closeGap := t.closeButtonMetrics(ctx)
				closeSizePx := gtx.Dp(closeSize)
				closeGapPx := gtx.Dp(closeGap)
				closeChild := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !t.closeVisible(itemState) {
						size := max(gtx.Dp(closeSize), 1)
						return layout.Dimensions{Size: image.Pt(size, size)}
					}
					return t.layoutCloseButton(ctx, gtx, item, itemState, disabled)
				})
				layoutContent = func(contentGtx layout.Context) layout.Dimensions {
					content := contentGtx
					// The outer Flex already reserves closeGapPx before this
					// rigid child, so only the close button's own width remains
					// to be removed from the content constraint.
					content.Constraints.Max.X = max(content.Constraints.Max.X-closeSizePx, 0)
					content.Constraints.Min.X = min(content.Constraints.Min.X, content.Constraints.Max.X)
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: content.Dp(theme.IconGap)}.Layout(content, children...)
				}
				layout.Inset{
					Left:  sizeStyle.paddingX,
					Right: t.tabRightPadding(ctx, sizeStyle, item),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: closeGapPx}.Layout(gtx,
						layout.Rigid(layoutContent),
						closeChild,
					)
				})
			} else {
				layout.Inset{
					Left:  sizeStyle.paddingX,
					Right: t.tabRightPadding(ctx, sizeStyle, item),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layoutContent(gtx)
				})
			}
			drawTabFocus(gtx, frame.ActiveTheme(ctx), size, t.variant, focus, style.focus)
			return layout.Dimensions{Size: size}
		}))
	})
}

func (t TabsWidget) closeVisible(itemState *tabsItemState) bool {
	if !t.closeOnHover {
		return true
	}
	return itemState.clickable.Hovered() || itemState.close.Hovered() || itemState.close.Pressed()
}

func (t TabsWidget) itemIsEditing(state *tabsState, item TabItem) bool {
	return t.onEdit != nil && item.Editable && state.editingKey == item.Key
}

func (t TabsWidget) prepareTabEditor(ctx *frame.Context, gtx layout.Context, componentState *tabsState, item TabItem, itemState *tabsItemState) {
	editKey := frame.DerivedKey(ctx, frame.FullKey(ctx, t.key), "edit:"+item.Key)
	editKey = frame.ClaimKey(ctx, state.KindInput, editKey)
	frame.RegisterFieldFocus(ctx, editKey, &itemState.editor, gtx.Enabled())
	itemState.editor.SingleLine = true
	itemState.editor.Submit = true
	itemState.editor.ReadOnly = false
	if !itemState.editReady {
		itemState.editor.SetText(item.Label)
		itemState.editReady = true
		frame.RequestFocus(ctx, &itemState.editor)
	}
	for {
		event, ok := itemState.editor.Update(gtx)
		if !ok {
			break
		}
		if _, ok := event.(widget.SubmitEvent); ok {
			t.finishTabEdit(componentState, item, itemState, true)
			return
		}
	}
	for {
		event, ok := gtx.Event(key.Filter{Focus: &itemState.editor, Name: key.NameEscape})
		if !ok {
			break
		}
		if pressed, ok := event.(key.Event); ok && pressed.State == key.Press {
			itemState.editor.SetText(item.Label)
			t.finishTabEdit(componentState, item, itemState, false)
			return
		}
	}
	focused := gtx.Focused(&itemState.editor)
	if focused {
		itemState.editFocused = true
	} else if itemState.editFocused {
		t.finishTabEdit(componentState, item, itemState, true)
	}
}

func (t TabsWidget) finishTabEdit(componentState *tabsState, item TabItem, itemState *tabsItemState, commit bool) {
	value := strings.TrimSpace(itemState.editor.Text())
	componentState.editingKey = ""
	itemState.editReady = false
	itemState.editFocused = false
	if commit && value != "" && t.onEdit != nil {
		t.onEdit(item.Key, value)
	}
}

func (t TabsWidget) layoutCloseButton(ctx *frame.Context, gtx layout.Context, item TabItem, itemState *tabsItemState, disabled bool) layout.Dimensions {
	closeSize, _ := t.closeButtonMetrics(ctx)
	size := max(gtx.Dp(closeSize), 1)
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	buttonGtx := gtx
	if disabled {
		buttonGtx = buttonGtx.Disabled()
	}
	return itemState.close.Layout(buttonGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp("Close " + t.itemAccessibleLabel(item)).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		col := frame.ActiveTheme(ctx).Palette.MutedForeground
		if itemState.close.Hovered() {
			col = frame.ActiveTheme(ctx).Palette.Foreground
		}
		if disabled {
			col = frame.ActiveTheme(ctx).DisabledColor(col)
		}
		drawTabClose(gtx, frame.ActiveTheme(ctx), image.Pt(size, size), col)
		return layout.Dimensions{Size: image.Pt(size, size)}
	})
}

func (t TabsWidget) itemAccessibleLabel(item TabItem) string {
	if item.AccessibleLabel != "" {
		return item.AccessibleLabel
	}
	if item.Label != "" {
		return item.Label
	}
	return item.Key
}

func (t TabsWidget) resolveItemStyle(ctx *frame.Context, gtx layout.Context, itemKey string, state flowstyle.StyleState, foreground color.NRGBA, size tabsSizeStyle) flowstyle.ResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	itemDefaults := flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: foreground}).
		FontSize(size.textSize).
		FontWeight(int(size.weight)).
		Opacity(1).
		Transition(flowstyle.PropOpacity, tabsColorDuration).
		When(flowstyle.All(flowstyle.Hovered, flowstyle.Not(flowstyle.Selected), flowstyle.Not(flowstyle.Disabled)),
			flowstyle.Style{}.Opacity(0.7))
	defaults := flowstyle.Style{}.
		Part(flowstyle.PartItem, itemDefaults.
			When(flowstyle.Disabled,
				flowstyle.Style{}.Opacity(activeTheme.DisabledOpacityValue())))

	return styleruntime.ResolvePart(
		ctx, gtx, frame.DerivedKey(ctx, frame.FullKey(ctx, t.key), "item:"+itemKey),
		flowstyle.PartItem, state, defaults, flowstyle.Style{}, flowstyle.Style{}, t.customStyle,
	)
}

func (t TabsWidget) layoutPanel(ctx *frame.Context, gtx layout.Context, item TabItem) layout.Dimensions {
	semanticRoot := frame.FullKey(ctx, t.key)
	scope := t.panelStateScope(ctx, item.Key)
	restoreHost := frame.PushKey(ctx, t.key)
	defer restoreHost()
	restoreItem := frame.PushKey(ctx, item.Key)
	defer restoreItem()
	if t.panelsKeepAlive() {
		restoreRetention := frame.PushStateRetention(ctx, scope)
		defer restoreRetention()
	}
	semantic.LabelOp(t.itemAccessibleLabel(item)).Add(gtx.Ops)
	semantic.DescriptionOp("Tab panel: " + t.itemAccessibleLabel(item)).Add(gtx.Ops)
	semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
	frame.RegisterSemantic(ctx, frame.SemanticNode{
		Key:         frame.DerivedKey(ctx, semanticRoot, "panel:"+item.Key),
		Role:        frame.SemanticTabPanel,
		Label:       t.itemAccessibleLabel(item),
		Description: "Tab panel: " + t.itemAccessibleLabel(item),
		Owner:       frame.DerivedKey(ctx, semanticRoot, "tab-list"),
		Hidden:      false,
		Selected:    true,
	})
	if t.orientation != TabsVertical {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	defaults := flowstyle.Style{}.
		Part(flowstyle.PartPanel, flowstyle.Style{}.Padding(frame.ActiveTheme(ctx).Components.Tabs.PanelPadding))

	resolved := styleruntime.ResolvePart(
		ctx, gtx, semanticRoot, flowstyle.PartPanel,
		flowstyle.StyleState{Selected: true, Disabled: !gtx.Enabled()},
		defaults, flowstyle.Style{}, flowstyle.Style{}, t.customStyle,
	)
	return layoutui.LayoutResolved(ctx, gtx, resolved, item.Panel)
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

func (t TabsWidget) layoutAddButton(ctx *frame.Context, gtx layout.Context, state *tabsState, disabled bool) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Tabs
	sizeStyle := tabsSizeStyleFor(frame.ActiveTheme(ctx), t.size)
	size := max(gtx.Dp(sizeStyle.height)+gtx.Dp(theme.ListPadding)*2, 1)
	if t.orientation == TabsVertical {
		size = max(gtx.Dp(theme.TabHeight), 1)
	}
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	buttonGtx := gtx
	if disabled {
		buttonGtx = buttonGtx.Disabled()
	}
	if !disabled {
		for state.add.Clicked(buttonGtx) {
			if t.onAdd != nil {
				t.onAdd()
			}
			frame.PreserveFocus(ctx)
		}
	}
	return state.add.Layout(buttonGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp("Add tab").Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
		col := frame.ActiveTheme(ctx).Palette.Foreground
		if state.add.Hovered() {
			col.A = byte(float32(col.A)*0.7 + 0.5)
		}
		if disabled {
			col = frame.ActiveTheme(ctx).DisabledColor(col)
		}
		iconSize := min(gtx.Dp(theme.IconSize), size)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
		iconOffset := op.Offset(image.Pt((size-iconSize)/2, (size-iconSize)/2)).Push(gtx.Ops)
		iconui.New(lucide.Plus).Size(float32(theme.IconSize)).Color(col).Layout(ctx, iconGtx)
		iconOffset.Pop()
		return layout.Dimensions{Size: image.Pt(size, size)}
	})
}
