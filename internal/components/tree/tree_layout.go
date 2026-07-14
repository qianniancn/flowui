package tree

import (
	"image"
	"strings"

	"gioui.org/font"
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
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

func (t Widget) layout(ctx *frame.Context, gtx layout.Context, treeStateValue *treeState, visible []flatItem) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Tree
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
				return t.layoutItem(ctx, gtx, treeStateValue, visible[index])
			})
		})
	}()
	content := macro.Stop()

	size := gtx.Constraints.Constrain(contentDims.Size)
	radius := treeRootRadius(gtx, frame.ActiveTheme(ctx), t.variant, size)
	drawTreeRoot(gtx, frame.ActiveTheme(ctx), size, radius, rootStyle)
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
	tokens := frame.ActiveTheme(ctx).Components.Tree
	height := min(gtx.Dp(tokens.RowHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(frame.ActiveTheme(ctx).Material, tokens.ItemTextSize, t.emptyText)
		label.Color = frame.ActiveTheme(ctx).Palette.MutedForeground
		label.MaxLines = 1
		return label.Layout(gtx)
	})
}

func (t Widget) layoutItem(ctx *frame.Context, gtx layout.Context, treeStateValue *treeState, entry flatItem) layout.Dimensions {
	itemState := treeStateValue.item(entry.item.Key)
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

	tokens := frame.ActiveTheme(ctx).Components.Tree
	rowHeight := tokens.RowHeight
	if entry.item.Description != "" {
		rowHeight = tokens.DescriptionRowHeight
	}
	height := min(max(gtx.Dp(rowHeight), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	width := gtx.Constraints.Max.X
	size := image.Pt(width, height)
	rowGtx := gtx
	rowGtx.Constraints = layout.Exact(size)

	focusVisible := itemState.focus.Visible(rowGtx.Focused(&itemState.clickable), itemState.clickable.History())
	focus := itemState.focus.Opacity(animGtx, focusVisible && !disabled)
	style := treeItemStyleFor(frame.ActiveTheme(ctx), selected, itemState.clickable.Hovered() && !disabled, itemState.clickable.Pressed() && !disabled, disabled)
	style.background = itemState.background.update(animGtx, style.background)
	expansionTarget := float32(0)
	if expanded {
		expansionTarget = 1
	}
	expansion := itemState.expansion.update(animGtx, expansionTarget, treeExpandDuration)
	scaleTarget := float32(1)
	if itemState.clickable.Pressed() && !disabled {
		scaleTarget = tokens.PressedScale
		if scaleTarget <= 0 || scaleTarget > 1 {
			scaleTarget = 0.98
		}
	}
	scale := itemState.scale.update(animGtx, scaleTarget, treeScaleDuration)

	macro := op.Record(gtx.Ops)
	func() {
		background := style.background
		if background.A == 0 {
			background = ctx.BackgroundColor()
		}
		restore := frame.PushColors(ctx, style.foreground, background)
		defer restore()
		t.layoutItemContent(ctx, rowGtx, itemState, entry, style, selected, expanded, expansion, disabled)
	}()
	content := macro.Stop()

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	transform := render.Scale(size, scale).Push(gtx.Ops)
	drawTreeRow(gtx, frame.ActiveTheme(ctx), size, style, focus)
	content.Add(gtx.Ops)
	transform.Pop()
	opacity.Pop()
	return layout.Dimensions{Size: size}
}

func (t Widget) layoutItemContent(ctx *frame.Context, gtx layout.Context, itemState *treeItemState, entry flatItem, style treeItemStyle, selected, expanded bool, expansion float32, disabled bool) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Tree
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
	tokens := frame.ActiveTheme(ctx).Components.Tree
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
		drawTreeChevron(gtx, frame.ActiveTheme(ctx), size, expansion, style.chevron)
		return layout.Dimensions{Size: size}
	})
}

func (t Widget) layoutMainContent(ctx *frame.Context, gtx layout.Context, item Item, style treeItemStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Tree
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
	tokens := frame.ActiveTheme(ctx).Components.Tree
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
