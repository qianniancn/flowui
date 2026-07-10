package listbox

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

func (l ListBoxWidget) layout(ctx *frame.Context, gtx layout.Context, state *listBoxState, entries []listBoxEntry, hasItems bool) layout.Dimensions {
	l.applyConstraints(ctx, &gtx)

	macro := op.Record(gtx.Ops)
	dims := l.layoutContent(ctx, gtx, state, entries, hasItems)
	call := macro.Stop()

	dims.Size = gtx.Constraints.Constrain(dims.Size)
	clipStack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (l ListBoxWidget) applyConstraints(ctx *frame.Context, gtx *layout.Context) {
	if l.fullWidth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	maxHeight := frame.ActiveTheme(ctx).Components.ListBox.MaxHeight
	if l.maxHeight > 0 {
		maxHeight = unit.Dp(l.maxHeight)
	}
	if maxHeight <= 0 {
		return
	}
	gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(maxHeight))
	gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
}

func (l ListBoxWidget) layoutContent(ctx *frame.Context, gtx layout.Context, state *listBoxState, entries []listBoxEntry, hasItems bool) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ListBox
	padding := theme.Padding
	if l.hasPadding {
		padding = l.padding
	}
	return layout.UniformInset(padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if !hasItems {
			return l.layoutEmpty(ctx, gtx)
		}
		state.list.Axis = layout.Vertical
		state.list.Gap = gtx.Dp(theme.Gap)
		state.list.Alignment = layout.Start
		state.list.ScrollToEnd = false
		state.list.ScrollAnyAxis = false
		return state.list.Layout(gtx, len(entries), func(gtx layout.Context, index int) layout.Dimensions {
			entry := entries[index]
			if entry.header {
				return l.layoutSectionHeader(ctx, gtx, entry.title)
			}
			return l.layoutItem(ctx, gtx, state, entry.item)
		})
	})
}

type listBoxEntry struct {
	header bool
	title  string
	item   ListBoxItem
}

func (l ListBoxWidget) layoutSectionHeader(ctx *frame.Context, gtx layout.Context, title string) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ListBox
	return layout.Inset{
		Top:    theme.SectionHeaderPaddingY,
		Right:  theme.SectionHeaderPaddingX,
		Bottom: theme.SectionHeaderPaddingY,
		Left:   theme.SectionHeaderPaddingX,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return text.New(title).
			Size(float32(theme.SectionHeaderTextSize)).
			Weight(font.Medium).
			Color(frame.ActiveTheme(ctx).Palette.MutedForeground).
			Layout(ctx, gtx)
	})
}

func (l ListBoxWidget) layoutEmpty(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ListBox
	height := min(gtx.Dp(theme.ItemMinHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return text.New(l.emptyText).
			Size(float32(theme.ItemTextSize)).
			Color(frame.ActiveTheme(ctx).Palette.MutedForeground).
			Layout(ctx, gtx)
	})
}

func (l ListBoxWidget) layoutItem(ctx *frame.Context, gtx layout.Context, listState *listBoxState, item ListBoxItem) layout.Dimensions {
	itemState := listState.item(item.Key)
	disabled := l.itemDisabled(item)
	selected := l.isSelected(item.Key)
	animGtx := gtx
	presses := state.ActivePresses(itemState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for itemState.clickable.Clicked(gtx) {
			l.activate(item.Key)
			frame.RequestFocus(ctx, &itemState.clickable)
		}
		frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		if item.Description != "" {
			semantic.DescriptionOp(item.Description).Add(gtx.Ops)
		}
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)

		theme := frame.ActiveTheme(ctx).Components.ListBox
		minHeight := min(gtx.Dp(theme.ItemMinHeight), gtx.Constraints.Max.Y)

		focusVisible := itemState.focusVisible(gtx.Focused(&itemState.clickable), itemState.clickable.History())
		style := listBoxItemStyleFor(frame.ActiveTheme(ctx), item.Variant, itemState.clickable.Hovered(), itemState.clickable.Pressed(), disabled)
		style.bg = itemState.background(animGtx, style.bg)
		style.selected = itemState.selection(animGtx, selected)
		style.focus = itemState.focusOpacity(animGtx, focusVisible && !disabled)
		scale := listBoxItemScale(animGtx, itemState.clickable.History(), frame.ActiveTheme(ctx), disabled)

		macro := op.Record(gtx.Ops)
		contentGtx := gtx
		contentGtx.Constraints.Min.Y = 0
		contentDims := l.layoutItemContent(ctx, contentGtx, item, style, selected)
		call := macro.Stop()
		size, contentOffset := listBoxItemFrame(gtx.Constraints, minHeight, contentDims.Size)
		dims := contentDims
		dims.Size = size
		dims.Baseline += max(size.Y-contentOffset.Y-contentDims.Size.Y, 0)

		stack := listBoxItemTransform(dims.Size, scale).Push(gtx.Ops)
		drawListBoxItem(gtx, frame.ActiveTheme(ctx), dims.Size, style)
		offset := op.Offset(contentOffset).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
		stack.Pop()
		return dims
	})
}

func listBoxItemFrame(constraints layout.Constraints, minHeight int, content image.Point) (size, offset image.Point) {
	size = constraints.Constrain(image.Pt(content.X, max(minHeight, content.Y)))
	offset.Y = max((size.Y-content.Y)/2, 0)
	return size, offset
}

func (l ListBoxWidget) layoutItemContent(ctx *frame.Context, gtx layout.Context, item ListBoxItem, style listBoxItemStyle, selected bool) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ListBox
	return layout.Inset{
		Top:    theme.ItemPaddingY,
		Right:  theme.ItemPaddingX,
		Bottom: theme.ItemPaddingY,
		Left:   theme.ItemPaddingX,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, 4)
		if item.Leading != nil {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return item.Leading.Layout(ctx, gtx)
			}))
		}
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return l.layoutItemText(ctx, gtx, item, style)
		}))
		if item.Trailing != nil {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return item.Trailing.Layout(ctx, gtx)
			}))
		}
		if l.showIndicator(item) {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return l.layoutIndicator(ctx, gtx, item, style, selected)
			}))
		}
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Gap:       gtx.Dp(theme.ItemContentGap),
		}.Layout(gtx, children...)
	})
}

func (l ListBoxWidget) showIndicator(item ListBoxItem) bool {
	if l.hideIndicator {
		return false
	}
	return item.Indicator != nil || l.selectionMode != ListBoxSelectionNone
}

func (l ListBoxWidget) layoutIndicator(ctx *frame.Context, gtx layout.Context, item ListBoxItem, style listBoxItemStyle, selected bool) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ListBox
	size := image.Pt(gtx.Dp(theme.ItemIndicatorSize), gtx.Dp(theme.ItemIndicatorSize))
	if item.Indicator == nil {
		drawListBoxIndicator(gtx, frame.ActiveTheme(ctx), size, style)
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

func (l ListBoxWidget) layoutItemText(ctx *frame.Context, gtx layout.Context, item ListBoxItem, style listBoxItemStyle) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ListBox
	if item.Description == "" {
		return text.New(item.Label).
			Size(float32(theme.ItemTextSize)).
			Weight(font.Medium).
			Color(style.fg).
			Layout(ctx, gtx)
	}
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(item.Label).
				Size(float32(theme.ItemTextSize)).
				Weight(font.Medium).
				Color(style.fg).
				Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(item.Description).
				Size(float32(theme.ItemDescriptionSize)).
				Color(style.description).
				Layout(ctx, gtx)
		}),
	)
}
