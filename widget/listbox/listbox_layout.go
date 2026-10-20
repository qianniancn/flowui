package listbox

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (l Widget) layout(ctx *core.Context, gtx layout.Context, styleKey string, state *listBoxState, visual theme.ListBoxVisual, entries []listBoxEntry, hasItems bool, instanceStyle style.Style) layout.Dimensions {
	l.applyConstraints(&gtx, visual.Metrics)

	macro := op.Record(gtx.Ops)
	dims := l.layoutContent(ctx, gtx, styleKey, state, visual, entries, hasItems, instanceStyle)
	call := macro.Stop()

	dims.Size = gtx.Constraints.Constrain(dims.Size)
	clipStack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (l Widget) applyConstraints(gtx *layout.Context, metrics theme.ListBoxMetrics) {
	if l.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	maxHeight := metrics.MaxHeight
	if l.maxHeight > 0 {
		maxHeight = unit.Dp(l.maxHeight)
	}
	if maxHeight <= 0 {
		return
	}
	gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(maxHeight))
	gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
}

func (l Widget) layoutContent(ctx *core.Context, gtx layout.Context, styleKey string, state *listBoxState, visual theme.ListBoxVisual, entries []listBoxEntry, hasItems bool, instanceStyle style.Style) layout.Dimensions {
	metrics := visual.Metrics
	active := core.ActiveTheme(ctx)
	padding := metrics.Padding
	if l.hasPadding {
		padding = l.padding
	}
	// The item Part override is identical for every visible row. Build it once
	// per layout instead of merging the same immutable declarations per item.
	return layout.UniformInset(padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if !hasItems {
			state.scrollToKey = ""
			return l.layoutEmpty(ctx, gtx, visual)
		}
		state.list.Axis = layout.Vertical
		state.list.Gap = gtx.Dp(metrics.Gap)
		state.list.Alignment = layout.Start
		state.list.ScrollToEnd = false
		state.list.ScrollAnyAxis = false
		if l.scrollToKey == "" {
			state.scrollToKey = ""
		} else {
			for index, entry := range entries {
				if entry.header || entry.item.Key != l.scrollToKey {
					continue
				}
				if l.scrollToKey != state.scrollToKey || index != state.scrollToIndex {
					position := state.list.Position
					if index <= position.First || index >= position.First+position.Count-1 {
						state.list.ScrollTo(index)
					}
					state.scrollToKey, state.scrollToIndex = l.scrollToKey, index
				}
				break
			}
		}
		return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &state.list, &state.bar, &state.visualOutset, len(entries), l.disabled, false, func(gtx layout.Context, index int) layout.Dimensions {
			entry := entries[index]
			if entry.header {
				return l.layoutSectionHeader(ctx, gtx, entry.title, visual)
			}
			return l.layoutItem(ctx, gtx, styleKey, state, active, visual, instanceStyle, entry.item)
		})
	})
}

type listBoxEntry struct {
	header bool
	title  string
	item   Item
}

func (l Widget) layoutSectionHeader(ctx *core.Context, gtx layout.Context, title string, visual theme.ListBoxVisual) layout.Dimensions {
	metrics := visual.Metrics
	return layout.Inset{
		Top:    metrics.SectionHeaderPaddingY,
		Right:  metrics.SectionHeaderPaddingX,
		Bottom: metrics.SectionHeaderPaddingY,
		Left:   metrics.SectionHeaderPaddingX,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return visual.Text(gtx, theme.ListBoxTextRequest{
			Text:  title,
			Color: visual.Colors.MutedForeground,
			Role:  theme.ListBoxTextSectionHeader,
		})
	})
}

func (l Widget) layoutEmpty(ctx *core.Context, gtx layout.Context, visual theme.ListBoxVisual) layout.Dimensions {
	metrics := visual.Metrics
	height := min(gtx.Dp(metrics.ItemMinHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		value := l.emptyText
		if !l.hasEmptyText {
			value = ctx.Text(i18n.CollectionNoItems)
		}
		return visual.Text(gtx, theme.ListBoxTextRequest{
			Text:  value,
			Color: visual.Colors.MutedForeground,
			Role:  theme.ListBoxTextEmpty,
		})
	})
}

func (l Widget) layoutItem(ctx *core.Context, gtx layout.Context, styleKey string, listState *listBoxState, active *theme.Theme, visual theme.ListBoxVisual, itemStyle style.Style, item Item) layout.Dimensions {
	itemState := listState.item(item.Key)
	if gtx.Focused(&itemState.Clickable) {
		listState.focusedKey = item.Key
	}
	disabled := l.itemDisabled(item)
	selected := l.isSelected(item.Key)
	animGtx := gtx
	presses := core.SnapshotPresses(itemState.Clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for itemState.Clickable.Clicked(gtx) {
			l.activate(listState, item.Key)
			core.RequestFocusVisible(ctx, &itemState.Clickable, presses.ClickFocusVisible(itemState.Clickable.History()))
		}
		core.FocusOnPress(ctx, &itemState.Clickable, itemState.Clickable.History(), presses.Active())
	}

	focused := animGtx.Focused(&itemState.Clickable)
	focusVisible := core.FocusVisible(ctx, &itemState.Clickable, focused)
	itemState.FocusOpacity(animGtx, focusVisible && !disabled, visual.Motion)
	styleState := style.State{
		Hovered:      !disabled && itemState.Clickable.Hovered(),
		Pressed:      !disabled && itemState.Clickable.Pressed(),
		Focused:      focused,
		FocusVisible: !disabled && focusVisible,
		Disabled:     disabled,
		Selected:     selected,
	}
	if itemState.styleKey == "" {
		itemState.styleKey = core.DerivedKey(ctx, styleKey, "item:"+item.Key)
		partKey := core.DerivedKey(ctx, itemState.styleKey, "style-part:"+string(style.PartItem))
		itemState.transitionKey = core.DerivedKey(ctx, partKey, "style-transition")
	}
	if itemState.transitionKey == "" {
		partKey := core.DerivedKey(ctx, itemState.styleKey, "style-part:"+string(style.PartItem))
		itemState.transitionKey = core.DerivedKey(ctx, partKey, "style-transition")
	}
	itemRequest := theme.ListBoxItemRequest{
		Variant: uint8(item.Variant),
		State:   styleState,
	}
	itemVisual := listState.itemVisualFor(active, visual, itemRequest)
	part := listState.resolvedItemPart(
		ctx,
		animGtx,
		active,
		itemState.transitionKey,
		itemRequest,
		itemVisual,
		itemStyle,
	)
	if itemVisual.Colors == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	itemColors := itemVisual.Colors(part)
	selectionProgress := itemState.Selection(animGtx, selected, visual.Metrics.SelectionDuration, visual.Motion)

	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		if item.Description != "" {
			semantic.DescriptionOp(item.Description).Add(gtx.Ops)
		}
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)

		return l.layoutItemContent(ctx, gtx, item, itemColors, selectionProgress, selected, visual)
	})
	return layoutui.InteractiveResolved(ctx, gtx, part, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return itemState.Clickable.Layout(gtx, visual)
	})
}

func (l Widget) layoutItemContent(ctx *core.Context, gtx layout.Context, item Item, itemColors theme.ListBoxItemColors, selectionProgress float32, selected bool, visual theme.ListBoxVisual) layout.Dimensions {
	metrics := visual.Metrics
	var children [4]layout.FlexChild
	count := 0
	if item.Leading != nil {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Leading.Layout(ctx, gtx)
		})
		count++
	}
	children[count] = layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return l.layoutItemText(gtx, item, itemColors, visual)
	})
	count++
	if item.Trailing != nil {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Trailing.Layout(ctx, gtx)
		})
		count++
	}
	if l.showIndicator(item) {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return l.layoutIndicator(ctx, gtx, item, itemColors, selectionProgress, selected, visual)
		})
		count++
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(metrics.ItemContentGap),
	}.Layout(gtx, children[:count]...)
}

func (l Widget) showIndicator(item Item) bool {
	if l.hideIndicator {
		return false
	}
	return item.Indicator != nil || l.selectionMode != SelectionNone
}

func (l Widget) layoutIndicator(ctx *core.Context, gtx layout.Context, item Item, itemColors theme.ListBoxItemColors, selectionProgress float32, selected bool, visual theme.ListBoxVisual) layout.Dimensions {
	size := image.Pt(gtx.Dp(visual.Metrics.ItemIndicatorSize), gtx.Dp(visual.Metrics.ItemIndicatorSize))
	if item.Indicator == nil {
		visual.Indicator(gtx, size, selectionProgress, itemColors.Indicator)
		return layout.Dimensions{Size: size}
	}
	indicator := item.Indicator(selected)
	if indicator == nil {
		return layout.Dimensions{Size: size}
	}
	childGtx := gtx
	childGtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(childGtx, func(gtx layout.Context) layout.Dimensions {
		return indicator.Layout(ctx, gtx)
	})
}

func (l Widget) layoutItemText(gtx layout.Context, item Item, itemColors theme.ListBoxItemColors, visual theme.ListBoxVisual) layout.Dimensions {
	if item.Description == "" {
		return visual.Text(gtx, theme.ListBoxTextRequest{
			Text:  item.Label,
			Color: itemColors.Foreground,
			Role:  theme.ListBoxTextItem,
		})
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return visual.Text(gtx, theme.ListBoxTextRequest{
				Text:  item.Label,
				Color: itemColors.Foreground,
				Role:  theme.ListBoxTextItem,
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return visual.Text(gtx, theme.ListBoxTextRequest{
				Text:  item.Description,
				Color: itemColors.Description,
				Role:  theme.ListBoxTextDescription,
			})
		}),
	)
}
