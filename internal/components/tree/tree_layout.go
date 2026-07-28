package tree

import (
	"fmt"
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
	"github.com/qianniancn/flowui/internal/animation"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/components/spinner"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
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
		label := material.Label(frame.ActiveMaterial(ctx), tokens.ItemTextSize, t.emptyText)
		label.Color = frame.ActiveTheme(ctx).Palette.MutedForeground
		label.MaxLines = 1
		return label.Layout(gtx)
	})
}

func (t Widget) layoutItem(ctx *frame.Context, gtx layout.Context, treeStateValue *treeState, entry flatItem, index int) layout.Dimensions {
	itemState := treeStateValue.item(entry.item.Key)
	itemState.index = index
	disabled := t.itemDisabled(entry.item)
	selected := t.itemSelected(entry.item.Key)
	expanded := treeContainsKey(t.expandedKeys, entry.item.Key)
	animGtx := gtx
	if disabled {
		gtx = gtx.Disabled()
	}

	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	height := min(max(treeItemHeight(gtx, tokens, entry.item), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	width := gtx.Constraints.Max.X
	size := image.Pt(width, height)
	rowGtx := gtx
	rowGtx.Constraints = layout.Exact(size)

	focusVisible := frame.FocusVisible(ctx, &itemState.clickable, rowGtx.Focused(&itemState.clickable))
	focus := itemState.focus.Opacity(animGtx, focusVisible && !disabled, frame.ActiveTheme(ctx).Motion)
	hovered := (itemState.clickable.Hovered() || itemState.toggle.Hovered()) && !disabled
	style := treeItemStyleFor(frame.ActiveTheme(ctx), selected, hovered, disabled)
	if treeContainsKey(treeStateValue.dragSources, entry.item.Key) {
		style.opacity *= 0.6
	}
	expansionTarget := float32(0)
	if expanded {
		expansionTarget = 1
	}
	expansion := itemState.expansion.Value(animGtx, expansionTarget, treeExpandDuration, animation.EaseSmoothstep, frame.ActiveTheme(ctx).Motion)

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
		t.layoutItemContent(ctx, rowGtx, treeStateValue, itemState, entry, style, selected, expanded, expansion, disabled)
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
	interactiveRow := func(gtx layout.Context) layout.Dimensions {
		if t.onDrop == nil || disabled {
			return row(gtx)
		}
		dims := row(gtx)
		pass := pointer.PassOp{}.Push(gtx.Ops)
		var preview layout.Widget
		if treeStateValue.dragSource == entry.item.Key {
			label := entry.item.Label
			if count := len(treeStateValue.dragSources); count > 1 {
				label = fmt.Sprintf("%s +%d", label, count-1)
			}
			offsetPx := gtx.Dp(tokens.DragPreviewOffset)
			offset := itemState.dragPress.Round().Add(image.Pt(offsetPx, offsetPx))
			preview = t.dragPreview(ctx, label, offset)
		}
		itemState.drag.Layout(gtx, func(layout.Context) layout.Dimensions { return dims }, preview)
		pass.Pop()
		registerTreeDragAreas(gtx, itemState, dims.Size, treeItemAcceptsChildren(entry.item))
		return dims
	}
	if !t.hasContextMenu || disabled || treeStateValue.renameKey == entry.item.Key {
		return interactiveRow(rowGtx)
	}
	trigger := frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return interactiveRow(gtx)
	})
	owner := frame.FullKey(ctx, t.key)
	key := frame.DerivedKey(ctx, owner, "context-menu:"+entry.item.Key)
	return menu.ContextMenu(key, trigger, t.contextMenu).
		FocusTarget(&itemState.clickable).
		OnOpenChange(func(open bool) {
			if open && t.onContextMenu != nil {
				t.onContextMenu(entry.item.Key)
			}
		}).
		Layout(ctx, rowGtx)
}

func (t Widget) dragPreview(ctx *frame.Context, label string, offset image.Point) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		stack := op.Offset(offset).Push(gtx.Ops)
		defer stack.Pop()
		return t.layoutDragPreview(ctx, gtx, label)
	}
}

func (t Widget) layoutDragPreview(ctx *frame.Context, gtx layout.Context, label string) layout.Dimensions {
	opacity := paint.PushOpacity(gtx.Ops, 0.5)
	defer opacity.Pop()
	activeTheme := frame.ActiveTheme(ctx)
	tokens := treeTokensFor(activeTheme, t.size)
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(tokens.DragPreviewMaxWidth))
	gtx.Constraints.Max.Y = max(gtx.Constraints.Max.Y, gtx.Dp(tokens.RowHeight+unit.Dp(8)))

	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top: tokens.DragPreviewPaddingY, Right: tokens.DragPreviewPaddingX,
		Bottom: tokens.DragPreviewPaddingY, Left: tokens.DragPreviewPaddingX,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		text := material.Label(theme.MaterialOf(activeTheme), tokens.ItemTextSize, label)
		text.Color = activeTheme.Palette.OverlayForegroundColor()
		text.Font.Weight = font.Medium
		text.MaxLines = 1
		text.Truncator = "..."
		return text.Layout(gtx)
	})
	content := macro.Stop()

	surface := activeTheme.Palette.OverlayColor()
	radius := min(max(gtx.Dp(tokens.DragPreviewRadius), 0), min(dims.Size.X, dims.Size.Y)/2)
	render.DrawSurface(
		gtx,
		image.Rectangle{Max: dims.Size},
		radius,
		surface,
		render.ThemeShadow(activeTheme.Shadows.Overlay, activeTheme.Palette.OverlayShadowColor(), 1),
	)
	content.Add(gtx.Ops)
	return dims
}

func treeItemHeight(gtx layout.Context, tokens theme.TreeTheme, item Item) int {
	height := tokens.RowHeight
	if treeItemDescription(item) != "" {
		height = tokens.DescriptionRowHeight
	}
	return gtx.Dp(height)
}

func registerTreeDragAreas(gtx layout.Context, state *treeItemState, size image.Point, acceptsChildren bool) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	edge := max(size.Y/4, 1)
	breaks := [4]int{0, edge, size.Y - edge, size.Y}
	if !acceptsChildren {
		breaks = [4]int{0, size.Y / 2, size.Y / 2, size.Y}
	}
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

func (t Widget) layoutItemContent(ctx *frame.Context, gtx layout.Context, treeStateValue *treeState, itemState *treeItemState, entry flatItem, style treeItemStyle, selected, expanded bool, expansion float32, disabled bool) layout.Dimensions {
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
				content := func(gtx layout.Context) layout.Dimensions {
					semantic.LabelOp(entry.item.Label).Add(gtx.Ops)
					if description := treeSemanticDescription(ctx, entry.item, expanded); description != "" {
						semantic.DescriptionOp(description).Add(gtx.Ops)
					}
					semantic.SelectedOp(selected).Add(gtx.Ops)
					semantic.EnabledOp(!disabled).Add(gtx.Ops)
					return t.layoutMainContent(ctx, gtx, treeStateValue, entry.item, style, expanded)
				}
				if treeStateValue.renameKey == entry.item.Key {
					return content(gtx)
				}
				return itemState.clickable.Layout(gtx, content)
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
	if !treeItemExpandable(item) {
		return layout.Dimensions{Size: size}
	}
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.toggle.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if item.ChildrenState == ChildrenLoading {
			return spinner.Spinner().Color(spinner.SpinnerCurrent).Size(spinner.SpinnerSmall).Label("Loading children").Layout(ctx, gtx)
		}
		if item.ChildrenState == ChildrenError {
			drawTreeRetryIcon(gtx, tokens, size, frame.ActiveTheme(ctx).Palette.Danger)
			return layout.Dimensions{Size: size}
		}
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

func (t Widget) layoutMainContent(ctx *frame.Context, gtx layout.Context, treeStateValue *treeState, item Item, style treeItemStyle, expanded bool) layout.Dimensions {
	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	children := make([]layout.FlexChild, 0, 3)
	leading := item.Leading
	if expanded && item.ExpandedLeading != nil {
		leading = item.ExpandedLeading
	}
	if leading != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return leading.Layout(ctx, gtx)
		}))
	}
	children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return t.layoutItemText(ctx, gtx, treeStateValue, item, style)
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

func (t Widget) layoutItemText(ctx *frame.Context, gtx layout.Context, treeStateValue *treeState, item Item, style treeItemStyle) layout.Dimensions {
	tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
	if treeStateValue.renameKey == item.Key {
		editor := material.Editor(frame.ActiveMaterial(ctx), &treeStateValue.renameEditor, "")
		editor.TextSize = tokens.ItemTextSize
		editor.Color = style.foreground
		editor.SelectionColor = frame.ActiveTheme(ctx).Palette.Selection
		editor.Font.Weight = font.Medium
		return editor.Layout(gtx)
	}
	descriptionText := treeItemDescription(item)
	labelWidget := func(gtx layout.Context) layout.Dimensions {
		label := material.Label(frame.ActiveMaterial(ctx), tokens.ItemTextSize, item.Label)
		label.Color = style.foreground
		label.Font.Weight = font.Medium
		label.MaxLines = 1
		label.Truncator = "..."
		return label.Layout(gtx)
	}
	if descriptionText == "" {
		return labelWidget(gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(labelWidget),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			description := material.Label(frame.ActiveMaterial(ctx), tokens.ItemDescriptionSize, descriptionText)
			description.Color = style.description
			if item.ChildrenState == ChildrenError {
				description.Color = frame.ActiveTheme(ctx).Palette.Danger
			}
			description.MaxLines = 1
			description.Truncator = "..."
			return description.Layout(gtx)
		}),
	)
}

func treeSemanticDescription(ctx *frame.Context, item Item, expanded bool) string {
	parts := make([]string, 0, 2)
	if description := treeItemDescription(item); description != "" {
		parts = append(parts, description)
	}
	if treeItemExpandable(item) {
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
	loadingState := ""
	switch item.ChildrenState {
	case ChildrenUnloaded:
		loadingState = "Children not loaded"
	case ChildrenLoading:
		loadingState = "Loading children"
	case ChildrenError:
		loadingState = "Load failed; activate the expand control to retry"
	}
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		switch item.ChildrenState {
		case ChildrenUnloaded:
			loadingState = "子项尚未加载"
		case ChildrenLoading:
			loadingState = "正在加载子项"
		case ChildrenError:
			loadingState = "子项加载失败，可激活展开控件重试"
		}
	}
	if loadingState != "" {
		parts = append(parts, loadingState)
	}
	return strings.Join(parts, ". ")
}

func treeItemDescription(item Item) string {
	if item.ChildrenState == ChildrenError && item.LoadError != "" {
		return item.LoadError
	}
	return item.Description
}
