package tree

import (
	"image"
	"strings"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func (t Widget) layout(ctx *frame.Context, gtx layout.Context, treeStateValue *treeState, visible []flatItem) layout.Dimensions {
	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	maxHeight := tokens.MaxHeight
	if t.maxHeight > 0 {
		maxHeight = unit.Dp(t.maxHeight)
	}
	if maxHeight > 0 {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(maxHeight))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}

	rootStyle := treeRootStyleFor(frame.ActiveTheme(ctx), t.variant)
	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	func() {
		background := ctx.BackgroundColor()
		if rootStyle.background.A != 0 {
			background = rootStyle.background
		}
		restore := frame.PushColors(ctx, rootStyle.foreground, background)
		defer restore()
		contentDims = layout.UniformInset(tokens.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if len(visible) == 0 {
				return t.layoutEmpty(ctx, gtx)
			}
			treeStateValue.list.Axis = layout.Vertical
			treeStateValue.list.Gap = gtx.Dp(tokens.Gap)
			treeStateValue.list.Alignment = layout.Start
			treeStateValue.list.ScrollAnyAxis = false
			return layoutui.LayoutTrackedScrollbar(ctx, gtx, &treeStateValue.list, &treeStateValue.bar, len(visible), t.disabled, false, func(gtx layout.Context, index int) layout.Dimensions {
				return t.layoutItem(ctx, gtx, treeStateValue, visible[index], index)
			})
		})
	}()
	content := macro.Stop()

	size := gtx.Constraints.Constrain(contentDims.Size)
	radius := treeRootRadius(gtx, tokens, t.variant, size)
	drawTreeRoot(gtx, frame.ActiveTheme(ctx), tokens, size, radius, rootStyle)
	if t.variant == VariantSurface {
		root := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
		content.Add(gtx.Ops)
		root.Pop()
	} else {
		root := clip.Rect{Max: size}.Push(gtx.Ops)
		content.Add(gtx.Ops)
		root.Pop()
	}
	return layout.Dimensions{Size: size, Baseline: contentDims.Baseline}
}

func (t Widget) layoutEmpty(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	height := min(gtx.Dp(tokens.RowHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(frame.ActiveTheme(ctx).Material, tokens.ItemTextSize, t.emptyText)
		label.Color = frame.ActiveTheme(ctx).Palette.MutedForeground
		label.MaxLines = 1
		return label.Layout(gtx)
	})
}

func (t Widget) layoutItem(ctx *frame.Context, gtx layout.Context, treeStateValue *treeState, entry flatItem, index int) layout.Dimensions {
	itemState := treeStateValue.item(entry.item.Key)
	itemState.index = index
	disabled := t.itemDisabled(entry.item)
	selected := t.selectionMode != SelectionNone && entry.item.Key == t.selectedKey
	expanded := treeContainsKey(t.expandedKeys, entry.item.Key)
	animGtx := gtx
	presses := state.ActivePresses(itemState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}

	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	height := min(max(treeItemHeight(gtx, tokens, entry.item), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	width := gtx.Constraints.Max.X
	size := image.Pt(width, height)
	rowGtx := gtx
	rowGtx.Constraints = layout.Exact(size)

	focusVisible := itemState.focus.Visible(rowGtx.Focused(&itemState.clickable), itemState.clickable.History())
	focus := itemState.focus.Opacity(animGtx, focusVisible && !disabled)
	hovered := (itemState.clickable.Hovered() || itemState.toggle.Hovered()) && !disabled
	style := treeItemStyleFor(frame.ActiveTheme(ctx), selected, hovered, disabled)
	if treeStateValue.dragSource == entry.item.Key {
		style.opacity *= 0.6
	}
	expansionTarget := float32(0)
	if expanded {
		expansionTarget = 1
	}
	expansion := itemState.expansion.update(animGtx, expansionTarget, treeExpandDuration)

	macro := op.Record(gtx.Ops)
	func() {
		background := style.background
		if background.A == 0 {
			background = ctx.BackgroundColor()
		}
		restore := frame.PushColors(ctx, style.foreground, background)
		defer restore()
		if t.guides {
			drawTreeGuides(rowGtx, entry, tokens, expanded, t.guideConnectors, t.guideStyle, frame.ActiveTheme(ctx).Palette.MutedForeground)
		}
		t.layoutItemContent(ctx, rowGtx, itemState, entry, style, selected, expanded, expansion, disabled)
	}()
	content := macro.Stop()

	row := func(gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, style.opacity)
		drawTreeRow(gtx, tokens, size, style, focus)
		content.Add(gtx.Ops)
		opacity.Pop()
		if treeStateValue.dropTarget.drawKey == entry.item.Key {
			drawTreeDropIndicator(gtx, size, treeStateValue.dropTarget.position, treeStateValue.dropTarget.depth, tokens, frame.ActiveTheme(ctx).Palette.Accent)
		}
		return layout.Dimensions{Size: size}
	}
	if t.onDrop == nil || disabled {
		return row(rowGtx)
	}
	dims := row(rowGtx)
	pass := pointer.PassOp{}.Push(rowGtx.Ops)
	itemState.drag.Layout(rowGtx, func(layout.Context) layout.Dimensions { return dims }, nil)
	pass.Pop()
	registerTreeDragAreas(rowGtx, itemState, dims.Size)
	return dims
}

func treeItemHeight(gtx layout.Context, tokens theme.TreeTheme, item Item) int {
	height := tokens.RowHeight
	if item.Description != "" {
		height = tokens.DescriptionRowHeight
	}
	return gtx.Dp(height)
}

func registerTreeDragAreas(gtx layout.Context, state *treeItemState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	edge := max(size.Y/4, 1)
	breaks := [4]int{0, edge, size.Y - edge, size.Y}
	pass := pointer.PassOp{}.Push(gtx.Ops)
	for index := range state.dropTags {
		if breaks[index] >= breaks[index+1] {
			continue
		}
		area := clip.Rect(image.Rect(0, breaks[index], size.X, breaks[index+1])).Push(gtx.Ops)
		event.Op(gtx.Ops, &state.dropTags[index])
		area.Pop()
	}
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	event.Op(gtx.Ops, &state.dragTag)
	area.Pop()
	pass.Pop()
}

func (t Widget) layoutItemContent(ctx *frame.Context, gtx layout.Context, itemState *treeItemState, entry flatItem, style treeItemStyle, selected, expanded bool, expansion float32, disabled bool) layout.Dimensions {
	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	left := tokens.RowPaddingX + unit.Dp(entry.depth)*tokens.Indent
	return layout.Inset{
		Top:    tokens.RowPaddingY,
		Right:  tokens.RowPaddingX,
		Bottom: tokens.RowPaddingY,
		Left:   left,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := []layout.FlexChild{
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return t.layoutToggle(ctx, gtx, itemState, entry.item, style, expansion, disabled)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					semantic.LabelOp(entry.item.Label).Add(gtx.Ops)
					if description := treeSemanticDescription(ctx, entry.item, expanded); description != "" {
						semantic.DescriptionOp(description).Add(gtx.Ops)
					}
					semantic.SelectedOp(selected).Add(gtx.Ops)
					semantic.EnabledOp(!disabled).Add(gtx.Ops)
					return t.layoutMainContent(ctx, gtx, entry.item, style)
				})
			}),
		}
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Gap:       gtx.Dp(tokens.ContentGap),
		}.Layout(gtx, children...)
	})
}

func (t Widget) layoutToggle(ctx *frame.Context, gtx layout.Context, itemState *treeItemState, item Item, style treeItemStyle, expansion float32, disabled bool) layout.Dimensions {
	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	slot := min(gtx.Dp(tokens.ChevronSlotSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	size := image.Pt(max(slot, 0), max(slot, 0))
	gtx.Constraints = layout.Exact(size)
	if len(item.Children) == 0 {
		return layout.Dimensions{Size: size}
	}
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.toggle.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		drawTreeToggleIcon(
			gtx,
			tokens,
			size,
			expansion,
			t.guides && t.guideConnectors,
			style.chevron,
			frame.ActiveTheme(ctx).Palette.SurfaceTertiary,
		)
		return layout.Dimensions{Size: size}
	})
}

func (t Widget) layoutMainContent(ctx *frame.Context, gtx layout.Context, item Item, style treeItemStyle) layout.Dimensions {
	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	children := make([]layout.FlexChild, 0, 3)
	if item.Leading != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Leading.Layout(ctx, gtx)
		}))
	}
	children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return t.layoutItemText(ctx, gtx, item, style)
	}))
	if item.Trailing != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Trailing.Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(tokens.ContentGap),
	}.Layout(gtx, children...)
}

func (t Widget) layoutItemText(ctx *frame.Context, gtx layout.Context, item Item, style treeItemStyle) layout.Dimensions {
	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	labelWidget := func(gtx layout.Context) layout.Dimensions {
		label := material.Label(frame.ActiveTheme(ctx).Material, tokens.ItemTextSize, item.Label)
		label.Color = style.foreground
		label.Font.Weight = font.Medium
		label.MaxLines = 1
		label.Truncator = "..."
		return label.Layout(gtx)
	}
	if item.Description == "" {
		return labelWidget(gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(labelWidget),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			description := material.Label(frame.ActiveTheme(ctx).Material, tokens.ItemDescriptionSize, item.Description)
			description.Color = style.description
			description.MaxLines = 1
			description.Truncator = "..."
			return description.Layout(gtx)
		}),
	)
}

func treeSemanticDescription(ctx *frame.Context, item Item, expanded bool) string {
	parts := make([]string, 0, 2)
	if item.Description != "" {
		parts = append(parts, item.Description)
	}
	if len(item.Children) > 0 {
		state := "Collapsed"
		if expanded {
			state = "Expanded"
		}
		if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
			state = "已折叠"
			if expanded {
				state = "已展开"
			}
		}
		parts = append(parts, state)
	}
	return strings.Join(parts, ". ")
}
