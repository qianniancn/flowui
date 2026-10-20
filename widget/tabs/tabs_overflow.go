package tabs

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
	dropdownui "github.com/qianniancn/flowui/widget/dropdown"
	menuui "github.com/qianniancn/flowui/widget/menu"
	textui "github.com/qianniancn/flowui/widget/text"
)

type tabsOverflowLayout struct {
	visible  []TabItem
	hidden   []TabItem
	moreSize image.Point
	listSize image.Point
	gap      int
}

// layoutTabStrip dispatches the opt-in overflow menu path while preserving the
// existing scroll-list implementation for the default mode.
func (t Widget) layoutTabStrip(ctx *core.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool, visual theme.TabsVisual) layout.Dimensions {
	if t.overflowMode != OverflowMenu && t.overflowMode != OverflowAuto {
		return t.layoutTabStripDefault(ctx, gtx, state, selectedKey, disabled, visual)
	}
	overflow := t.measureOverflow(ctx, gtx, selectedKey, visual)
	if len(overflow.hidden) == 0 {
		return t.layoutTabStripDefault(ctx, gtx, state, selectedKey, disabled, visual)
	}
	return t.layoutOverflowTabStrip(ctx, gtx, state, selectedKey, disabled, overflow, visual)
}

func (t Widget) measureOverflow(ctx *core.Context, gtx layout.Context, selectedKey string, visuals ...theme.TabsVisual) tabsOverflowLayout {
	var visual theme.TabsVisual
	if len(visuals) > 0 {
		visual = visuals[0]
	} else {
		visual = t.visual(ctx)
	}
	metrics := visual.Metrics
	padding := gtx.Dp(metrics.ListPadding)
	if t.variant == Secondary {
		padding = 0
	}
	sizeStyle := tabsSizeStyleForVisual(metrics, t.size)
	tabHeight := max(gtx.Dp(sizeStyle.height), 1)
	measureTabs := t
	// Force intrinsic widths for the visibility calculation. The normal list
	// may distribute spare width when Fit and Centered are not requested.
	measureTabs.fit = true
	measureTabs.centered = true
	widths := measureTabs.tabWidths(ctx, gtx, padding, sizeStyle, visual)
	if len(widths) == 0 {
		return tabsOverflowLayout{}
	}
	moreWidth := t.moreMainSize(ctx, gtx, sizeStyle, tabHeight, padding, visual)
	moreMain := moreWidth
	moreCross := tabHeight + padding*2
	available := gtx.Constraints.Max.X
	if t.orientation == Vertical {
		available = gtx.Constraints.Max.Y
		moreMain = tabHeight + padding*2
		moreCross = max(moreWidth, widths[0]+padding*2, gtx.Constraints.Min.X)
		for _, width := range widths {
			moreCross = max(moreCross, width+padding*2)
		}
	}
	available = max(available, 0)
	if moreMain > available {
		moreMain = available
	}
	visibleCount := len(t.items)
	tabGap := gtx.Dp(metrics.TabGap)
	for visibleCount > 0 {
		main := padding * 2
		if t.orientation == Vertical {
			main += visibleCount * tabHeight
			main += max(visibleCount-1, 0) * tabGap
		} else {
			for index := 0; index < visibleCount; index++ {
				main += widths[index]
			}
			main += max(visibleCount-1, 0) * tabGap
		}
		if main+tabGap+moreMain <= available {
			break
		}
		visibleCount--
	}
	if visibleCount == len(t.items) {
		return tabsOverflowLayout{}
	}
	visibleCount = max(visibleCount, 0)
	listItems, hiddenItems := t.overflowItems(visibleCount, selectedKey)
	gap := tabGap
	if moreMain >= available || available-moreMain <= gap {
		gap = 0
	}
	moreSize := image.Pt(moreMain, moreCross)
	listMain := max(available-moreMain-gap, 0)
	listSize := image.Pt(gtx.Constraints.Max.X, tabHeight+padding*2)
	if t.orientation == Vertical {
		moreSize = image.Pt(moreCross, moreMain)
		listSize = image.Pt(moreCross, listMain)
	} else {
		listSize.X = listMain
	}
	// These are siblings inside the overflow flex row/column. Applying the
	// parent's minimum constraint to each one would make both children consume
	// the full available axis (for example, two 640px children in a 640px
	// strip). Clamp only to the parent's maximum; the computed allocation above
	// already accounts for the More trigger and inter-item gap.
	moreSize = clampOverflowSize(gtx, moreSize)
	listSize = clampOverflowSize(gtx, listSize)
	return tabsOverflowLayout{
		visible:  listItems,
		hidden:   hiddenItems,
		moreSize: moreSize,
		listSize: listSize,
		gap:      gap,
	}
}

func clampOverflowSize(gtx layout.Context, size image.Point) image.Point {
	size.X = max(size.X, 0)
	size.Y = max(size.Y, 0)
	if maxWidth := gtx.Constraints.Max.X; maxWidth >= 0 {
		size.X = min(size.X, maxWidth)
	}
	if maxHeight := gtx.Constraints.Max.Y; maxHeight >= 0 {
		size.Y = min(size.Y, maxHeight)
	}
	return size
}

// overflowItems keeps the active tab in the visible strip whenever there is
// room for at least one tab. Without this promotion, selecting an item from
// More would leave the active indicator inside the menu and make the current
// tab impossible to discover from the tab bar itself.
func (t Widget) overflowItems(visibleCount int, selectedKey string) ([]TabItem, []TabItem) {
	visibleCount = min(max(visibleCount, 0), len(t.items))
	if visibleCount == 0 {
		return nil, append([]TabItem(nil), t.items...)
	}
	visible := append([]TabItem(nil), t.items[:visibleCount]...)
	selectedIndex := tabsIndexByKey(t.items, selectedKey)
	if selectedIndex >= visibleCount {
		visible[len(visible)-1] = t.items[selectedIndex]
	}
	visibleKeys := make(map[string]struct{}, len(visible))
	for _, item := range visible {
		visibleKeys[item.Key] = struct{}{}
	}
	hidden := make([]TabItem, 0, len(t.items)-len(visible))
	for _, item := range t.items {
		if _, ok := visibleKeys[item.Key]; ok {
			continue
		}
		hidden = append(hidden, item)
	}
	return visible, hidden
}

func (t Widget) moreMainSize(ctx *core.Context, gtx layout.Context, sizeStyle tabsSizeStyle, tabHeight, padding int, visual theme.TabsVisual) int {
	if t.moreContent != nil {
		measure := gtx
		measure.Constraints = layout.Constraints{Max: image.Pt(1e6, max(gtx.Constraints.Max.Y, 1))}
		dims := core.MeasureWidget(ctx, measure, t.slotWidget("overflow-trigger", t.moreContent))
		return max(tabHeight+padding*2, dims.Size.X+gtx.Dp(sizeStyle.paddingX)*2)
	}
	label := t.moreText(ctx)
	measure := gtx
	measure.Constraints = layout.Constraints{Max: image.Pt(1e6, max(gtx.Constraints.Max.Y, 1))}
	dims := textui.Measure(ctx, measure, textui.WithResolvedText(textui.New(label), t.itemTextStyle(ctx, visual, visual.Colors.ScrollForeground)))
	return max(
		tabHeight+padding*2,
		dims.Size.X+gtx.Dp(sizeStyle.paddingX)*2+gtx.Dp(visual.Metrics.IconSize)+gtx.Dp(visual.Metrics.IconGap),
	)
}

func (t Widget) layoutOverflowTabStrip(ctx *core.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool, overflow tabsOverflowLayout, visual theme.TabsVisual) layout.Dimensions {
	gap := overflow.gap
	listWidget := t
	listWidget.items = overflow.visible
	listWidget.overflowMode = OverflowScroll
	listGtx := gtx
	listGtx.Constraints = layout.Exact(overflow.listSize)
	moreGtx := gtx
	moreGtx.Constraints = layout.Exact(overflow.moreSize)
	more := t.layoutOverflowMenu(ctx, moreGtx, state, selectedKey, disabled, overflow.hidden, visual)
	children := make([]layout.FlexChild, 0, 2)
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return listWidget.layoutTabStripDefault(ctx, listGtx, state, selectedKey, disabled, visual)
	}))
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return more.Layout(ctx, moreGtx)
	}))
	axis := layout.Horizontal
	if t.orientation == Vertical {
		axis = layout.Vertical
	}
	return layout.Flex{Axis: axis, Alignment: layout.Start, Gap: gap}.Layout(gtx, children...)
}

func (t Widget) layoutOverflowMenu(ctx *core.Context, gtx layout.Context, state *tabsState, selectedKey string, disabled bool, hidden []TabItem, visual theme.TabsVisual) core.Widget {
	items := make([]menuui.Item, 0, len(hidden))
	for _, item := range hidden {
		items = append(items, menuui.Item{
			Key:         item.Key,
			Label:       t.itemAccessibleLabel(item),
			Description: item.Description,
			Disabled:    item.Disabled,
		})
	}
	moreKey := core.DerivedKey(ctx, core.FullKey(ctx, t.key), "overflow-menu")
	label := t.moreText(ctx)
	defaultContent := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		color := visual.Colors.ScrollForeground
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(visual.Metrics.IconGap)}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if visual.Overflow == nil {
					panic(theme.ErrIncompleteVisualProvider)
				}
				size := gtx.Constraints.Constrain(image.Pt(gtx.Dp(visual.Metrics.IconSize), gtx.Dp(visual.Metrics.IconSize)))
				visual.Overflow(gtx, theme.TabsOverflowRenderRequest{
					Size: size, Color: color,
				})
				return layout.Dimensions{Size: size}
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return textui.WithResolvedText(textui.New(label), t.itemTextStyle(ctx, visual, color)).Layout(ctx, gtx)
			}),
		)
	})
	var triggerContent core.Widget = defaultContent
	if t.moreContent != nil {
		triggerContent = t.slotWidget("overflow-trigger", t.moreContent)
	}
	trigger := button.New("").Key(moreKey + ":trigger").Content(triggerContent).Label(label).Variant(button.Ghost)
	menu := dropdownui.New(moreKey, trigger, items).
		AutoWidth().
		SelectionMode(menuui.SelectionSingle).
		SelectedKey(selectedKey).
		OnActionEvent(func(event menuui.ActionEvent) {
			if tabsIndexByKey(t.items, event.Key) < 0 {
				return
			}
			selected := state.requestSelectedKey(t, event.Key)
			if t.hasSelectedKey {
				selected = event.Key
			}
			state.selectedKey = selected
			state.selectionPending = true
			core.RequestFocus(ctx, &state.item(event.Key).clickable)
		})
	if disabled {
		menu = menu.Disabled(true)
	}
	return menu
}

func (t Widget) moreText(ctx *core.Context) string {
	if t.moreLabel != "" {
		return t.moreLabel
	}
	return ctx.Text(i18n.TabsMore)
}
