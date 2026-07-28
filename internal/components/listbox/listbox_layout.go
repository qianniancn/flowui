package listbox

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/optionrow"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

func (l ListBoxWidget) layout(ctx *frame.Context, gtx layout.Context, styleKey string, state *listBoxState, entries []listBoxEntry, hasItems bool) layout.Dimensions {
	l.applyConstraints(ctx, &gtx)

	macro := op.Record(gtx.Ops)
	dims := l.layoutContent(ctx, gtx, styleKey, state, entries, hasItems)
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

func (l ListBoxWidget) layoutContent(ctx *frame.Context, gtx layout.Context, styleKey string, state *listBoxState, entries []listBoxEntry, hasItems bool) layout.Dimensions {
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
		return layoutui.LayoutTrackedScrollbar(ctx, gtx, &state.list, &state.bar, len(entries), l.disabled, false, func(gtx layout.Context, index int) layout.Dimensions {
			entry := entries[index]
			if entry.header {
				return l.layoutSectionHeader(ctx, gtx, entry.title)
			}
			return l.layoutItem(ctx, gtx, styleKey, state, entry.item)
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

func (l ListBoxWidget) layoutItem(ctx *frame.Context, gtx layout.Context, styleKey string, listState *listBoxState, item ListBoxItem) layout.Dimensions {
	itemState := listState.item(item.Key)
	if gtx.Focused(&itemState.Clickable) {
		listState.focusedKey = item.Key
	}
	disabled := l.itemDisabled(item)
	selected := l.isSelected(item.Key)
	animGtx := gtx
	presses := state.SnapshotPresses(itemState.Clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for itemState.Clickable.Clicked(gtx) {
			l.activate(listState, item.Key)
			frame.RequestFocusVisible(ctx, &itemState.Clickable, presses.ClickFocusVisible(itemState.Clickable.History()))
		}
		frame.FocusOnPress(ctx, &itemState.Clickable, itemState.Clickable.History(), presses.Active())
	}

	focused := animGtx.Focused(&itemState.Clickable)
	focusVisible := frame.FocusVisible(ctx, &itemState.Clickable, focused)
	itemState.FocusOpacity(animGtx, focusVisible && !disabled, frame.ActiveTheme(ctx).Motion)
	styleState := flowstyle.StyleState{
		Hovered:      !disabled && itemState.Clickable.Hovered(),
		Pressed:      !disabled && itemState.Clickable.Pressed(),
		Focused:      focused,
		FocusVisible: !disabled && focusVisible,
		Disabled:     disabled,
		Selected:     selected,
	}
	part := styleruntime.ResolvePart(
		ctx,
		animGtx,
		frame.DerivedKey(ctx, styleKey, "item:"+item.Key),
		flowstyle.PartItem,
		styleState,
		listBoxItemDefaultDeclaration(frame.ActiveTheme(ctx)),
		listBoxItemVariantDeclaration(frame.ActiveTheme(ctx), item.Variant),
		flowstyle.Style{},
		flowstyle.Join(l.customStyle, l.partsStyle),
	)
	visual := listBoxItemStyleFor(frame.ActiveTheme(ctx), item.Variant, disabled)
	visual.selected = itemState.Selection(animGtx, selected, listBoxItemSelectDuration, frame.ActiveTheme(ctx).Motion)
	if part.Text != nil {
		if col, ok := styleruntime.Color(part.Text.Color); ok {
			visual.fg = col
			visual.description = col
			visual.indicator = col
		}
	}

	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		if item.Description != "" {
			semantic.DescriptionOp(item.Description).Add(gtx.Ops)
		}
		semantic.SelectedOp(selected).Add(gtx.Ops)
		semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)

		return l.layoutItemContent(ctx, gtx, item, visual, selected)
	})
	return layoutui.LayoutInteractiveResolved(ctx, gtx, part, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return itemState.Clickable.Layout(gtx, visual)
	})
}

func (l ListBoxWidget) layoutItemContent(ctx *frame.Context, gtx layout.Context, item ListBoxItem, style listBoxItemStyle, selected bool) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ListBox
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
	return optionrow.LayoutText(
		ctx, gtx, item.Label, item.Description,
		float32(theme.ItemTextSize), float32(theme.ItemDescriptionSize),
		style.fg, style.description,
	)
}
