package tree

import (
	"fmt"
	"image"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/checkbox"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/spinner"
)

func (t Widget) layout(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, visible []flatItem, visual theme.TreeVisual) layout.Dimensions {
	metrics := visual.Metrics
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	maxHeight := metrics.MaxHeight
	if t.maxHeight > 0 {
		maxHeight = unit.Dp(t.maxHeight)
	}
	if maxHeight > 0 {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(maxHeight))
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	func() {
		restore := core.PushColors(ctx, visual.Colors.RootForeground, visual.Colors.RootBackground)
		defer restore()
		contentDims = layout.UniformInset(metrics.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if len(visible) == 0 {
				return t.layoutEmpty(ctx, gtx, visual)
			}
			treeStateValue.list.Axis = layout.Vertical
			treeStateValue.list.Gap = gtx.Dp(metrics.Gap)
			treeStateValue.list.Alignment = layout.Start
			treeStateValue.list.ScrollAnyAxis = false
			return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &treeStateValue.list, &treeStateValue.bar, &treeStateValue.visualOutset, len(visible), t.disabled, false, func(gtx layout.Context, index int) layout.Dimensions {
				return t.layoutItem(ctx, gtx, treeStateValue, visible[index], index, visual)
			})
		})
	}()
	content := macro.Stop()

	size := gtx.Constraints.Constrain(contentDims.Size)
	radius := treeRootRadius(gtx, metrics, t.variant, size)
	visual.Root(gtx, size, metrics.SurfaceRadius, visual.Colors.RootBackground, t.variant == VariantSurface)
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

func (t Widget) layoutEmpty(ctx *core.Context, gtx layout.Context, visual theme.TreeVisual) layout.Dimensions {
	metrics := visual.Metrics
	height := min(gtx.Dp(metrics.RowHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		value := t.emptyText
		if !t.hasEmptyText {
			value = ctx.Text(i18n.CollectionNoItems)
		}
		return visual.Text(gtx, theme.TreeTextRequest{Text: value, Size: metrics.ItemTextSize, Color: visual.Colors.Description})
	})
}

func (t Widget) layoutItem(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, entry flatItem, index int, visual theme.TreeVisual) layout.Dimensions {
	itemState := treeStateValue.item(entry.item.Key)
	itemState.index = index
	disabled := t.itemDisabled(entry.item)
	selected := t.itemSelected(entry.item.Key)
	expanded := treeContainsKey(t.expandedKeys, entry.item.Key)
	animGtx := gtx
	if disabled {
		gtx = gtx.Disabled()
	}

	metrics := visual.Metrics
	height := min(max(treeItemHeight(gtx, metrics, entry.item), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	width := gtx.Constraints.Max.X
	size := image.Pt(width, height)
	rowGtx := gtx
	rowGtx.Constraints = layout.Exact(size)

	focusVisible := core.FocusVisible(ctx, &itemState.clickable, rowGtx.Focused(&itemState.clickable))
	focus := itemState.focus.Opacity(animGtx, focusVisible && !disabled, visual.Motion)
	hovered := (itemState.clickable.Hovered() || itemState.toggle.Hovered() || itemState.check.Hovered()) && !disabled
	itemVisual := visual.Item(theme.TreeItemRequest{Selected: selected, Hovered: hovered, Disabled: disabled})
	if treeContainsKey(treeStateValue.dragSources, entry.item.Key) {
		itemVisual.Opacity *= 0.6
	}
	expansionTarget := float32(0)
	if expanded {
		expansionTarget = 1
	}
	expansion := itemState.expansion.Value(animGtx, expansionTarget, treeExpandDuration, animation.EaseSmoothstep, visual.Motion)

	macro := op.Record(gtx.Ops)
	func() {
		restore := core.PushColors(ctx, itemVisual.Foreground, itemVisual.Background)
		defer restore()
		if t.guides {
			drawTreeGuides(rowGtx, entry, metrics, expanded, t.guideConnectors, t.guideStyle, visual)
		}
		t.layoutItemContent(ctx, rowGtx, treeStateValue, itemState, entry, itemVisual, selected, expanded, expansion, disabled, visual)
	}()
	content := macro.Stop()

	row := func(gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, itemVisual.Opacity)
		visual.Row(gtx, size, itemVisual, focus, metrics.RowRadius)
		content.Add(gtx.Ops)
		opacity.Pop()
		if treeStateValue.dropTarget.drawKey == entry.item.Key {
			visual.DropIndicator(gtx, size, treeDropPosition(treeStateValue.dropTarget.position), treeStateValue.dropTarget.depth, metrics, visual.Colors.DropIndicator)
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
			offsetPx := gtx.Dp(metrics.DragPreviewOffset)
			offset := itemState.drag.Press().Round().Add(image.Pt(offsetPx, offsetPx))
			preview = t.dragPreview(ctx, label, offset, visual)
		}
		itemState.drag.Layout(gtx, func(layout.Context) layout.Dimensions { return dims }, preview)
		pass.Pop()
		registerTreeDragAreas(gtx, itemState, dims.Size, treeItemAcceptsChildren(entry.item), !entry.item.DragDisabled, !entry.item.DropDisabled)
		return dims
	}
	if !t.hasContextMenu || disabled || treeStateValue.renameKey == entry.item.Key {
		return interactiveRow(rowGtx)
	}
	trigger := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return interactiveRow(gtx)
	})
	owner := core.FullKey(ctx, t.key)
	key := core.DerivedKey(ctx, owner, "context-menu:"+entry.item.Key)
	return menu.ContextMenu(key, trigger, t.contextMenu).
		FocusTarget(&itemState.clickable).
		OnOpenChange(func(open bool) {
			if open && t.onContextMenu != nil {
				t.onContextMenu(entry.item.Key)
			}
		}).
		Layout(ctx, rowGtx)
}

func (t Widget) dragPreview(ctx *core.Context, label string, offset image.Point, visual theme.TreeVisual) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		stack := op.Offset(offset).Push(gtx.Ops)
		defer stack.Pop()
		return t.layoutDragPreview(gtx, label, visual)
	}
}

func (t Widget) layoutDragPreview(gtx layout.Context, label string, visual theme.TreeVisual) layout.Dimensions {
	return visual.DragPreview(gtx, theme.TreeDragPreviewRequest{Label: label}, visual.Metrics)
}

func treeItemHeight(gtx layout.Context, metrics theme.TreeMetrics, item Item) int {
	height := metrics.RowHeight
	if treeItemDescription(item) != "" {
		height = metrics.DescriptionRowHeight
	}
	return gtx.Dp(height)
}

func registerTreeDragAreas(gtx layout.Context, state *treeItemState, size image.Point, acceptsChildren, draggable, droppable bool) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	edge := max(size.Y/4, 1)
	breaks := [4]int{0, edge, size.Y - edge, size.Y}
	if !acceptsChildren {
		breaks = [4]int{0, size.Y / 2, size.Y / 2, size.Y}
	}
	if droppable {
		for index := range state.dropTargets {
			if breaks[index] >= breaks[index+1] {
				continue
			}
			state.dropTargets[index].Register(gtx, image.Rect(0, breaks[index], size.X, breaks[index+1]))
		}
	}
	if draggable {
		state.drag.RegisterSource(gtx, image.Rectangle{Max: size})
	}
}

func (t Widget) layoutItemContent(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, itemState *treeItemState, entry flatItem, itemVisual theme.TreeItemVisual, selected, expanded bool, expansion float32, disabled bool, visual theme.TreeVisual) layout.Dimensions {
	metrics := visual.Metrics
	left := metrics.RowPaddingX + unit.Dp(entry.depth)*metrics.Indent
	return layout.Inset{
		Top:    metrics.RowPaddingY,
		Right:  metrics.RowPaddingX,
		Bottom: metrics.RowPaddingY,
		Left:   left,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, 3)
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutToggle(ctx, gtx, itemState, entry.item, itemVisual, expansion, disabled, visual)
		}))
		if t.checkable {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return t.layoutCheckbox(ctx, gtx, treeStateValue, itemState, entry.item)
			}))
		}
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			content := func(gtx layout.Context) layout.Dimensions {
				semantic.LabelOp(entry.item.Label).Add(gtx.Ops)
				if description := treeSemanticDescription(ctx, entry.item, expanded); description != "" {
					semantic.DescriptionOp(description).Add(gtx.Ops)
				}
				semantic.SelectedOp(selected).Add(gtx.Ops)
				semantic.EnabledOp(!disabled).Add(gtx.Ops)
				return t.layoutMainContent(ctx, gtx, treeStateValue, entry.item, itemVisual, selected, expanded, disabled, visual)
			}
			if treeStateValue.renameKey == entry.item.Key {
				return content(gtx)
			}
			return itemState.clickable.Layout(gtx, content)
		}))
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Gap:       gtx.Dp(metrics.ContentGap),
		}.Layout(gtx, children...)
	})
}

func (t Widget) layoutCheckbox(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, itemState *treeItemState, item Item) layout.Dimensions {
	if !t.checkable {
		return layout.Dimensions{}
	}
	disabled := t.itemCheckboxDisabled(item)
	checked := t.itemChecked(item.Key)
	halfChecked := t.itemHalfChecked(item.Key)
	content := func(gtx layout.Context) layout.Dimensions {
		semantic.CheckBox.Add(gtx.Ops)
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		semantic.SelectedOp(checked).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		if halfChecked {
			description := ctx.Text(i18n.TreePartiallySelected)
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		selection := float32(0)
		if checked || halfChecked {
			selection = 1
		}
		return checkbox.DrawControl(ctx, gtx, checkbox.ControlOptions{
			Variant:       checkbox.Primary,
			Selection:     selection,
			Indeterminate: halfChecked,
			Hovered:       itemState.check.Hovered(),
			Pressed:       itemState.check.Pressed(),
			Disabled:      disabled,
			VisualCache:   &treeStateValue.checkboxVisual,
		})
	}
	if disabled {
		gtx = gtx.Disabled()
	}
	return itemState.check.Layout(gtx, content)
}

func (t Widget) layoutToggle(ctx *core.Context, gtx layout.Context, itemState *treeItemState, item Item, itemVisual theme.TreeItemVisual, expansion float32, disabled bool, visual theme.TreeVisual) layout.Dimensions {
	metrics := visual.Metrics
	slot := min(gtx.Dp(metrics.ChevronSlotSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
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
			return spinner.New().Color(spinner.Current).Size(spinner.Small).Label(ctx.Text(i18n.TreeLoadingChildren)).Layout(ctx, gtx)
		}
		if item.ChildrenState == ChildrenError {
			visual.Retry(gtx, size, metrics, visual.Colors.Retry)
			return layout.Dimensions{Size: size}
		}
		switcher := item.Switcher
		if expansion >= .5 && item.ExpandedSwitcher != nil {
			switcher = item.ExpandedSwitcher
		}
		if switcher != nil {
			restore := core.PushColors(ctx, itemVisual.Chevron, visual.Colors.ToggleBackground)
			defer restore()
			return switcher.Layout(ctx, gtx)
		}
		visual.Toggle(
			gtx,
			size,
			metrics,
			expansion,
			t.guides && t.guideConnectors,
			itemVisual.Chevron,
			visual.Colors.ToggleBackground,
		)
		return layout.Dimensions{Size: size}
	})
}

func (t Widget) layoutMainContent(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, item Item, itemVisual theme.TreeItemVisual, selected, expanded, disabled bool, visual theme.TreeVisual) layout.Dimensions {
	if item.Content != nil {
		return item.Content.Layout(ctx, gtx)
	}
	metrics := visual.Metrics
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
		return t.layoutItemText(ctx, gtx, treeStateValue, item, itemVisual, selected, disabled, visual)
	}))
	if item.Trailing != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Trailing.Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(metrics.ContentGap),
	}.Layout(gtx, children...)
}

func (t Widget) layoutItemText(ctx *core.Context, gtx layout.Context, treeStateValue *treeState, item Item, itemVisual theme.TreeItemVisual, selected, disabled bool, visual theme.TreeVisual) layout.Dimensions {
	metrics := visual.Metrics
	if treeStateValue.renameKey == item.Key {
		editorStyle := visual.EditorStyle(theme.TreeItemRequest{Selected: selected, Disabled: disabled})
		editor := material.Editor(core.ActiveMaterial(ctx), &treeStateValue.renameEditor, "")
		editor.TextSize = editorStyle.TextSize
		editor.Color = editorStyle.Foreground
		editor.SelectionColor = editorStyle.Selection
		editor.Font.Weight = editorStyle.FontWeight
		return editor.Layout(gtx)
	}
	descriptionText := treeItemDescription(item)
	labelWidget := func(gtx layout.Context) layout.Dimensions {
		return visual.Text(gtx, theme.TreeTextRequest{Text: item.Label, Size: metrics.ItemTextSize, Color: itemVisual.Foreground, Weight: metrics.ItemTextWeight})
	}
	if descriptionText == "" {
		return labelWidget(gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(labelWidget),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if item.ChildrenState == ChildrenError {
				return visual.Text(gtx, theme.TreeTextRequest{Text: descriptionText, Size: metrics.ItemDescriptionSize, Color: visual.Colors.Retry})
			}
			return visual.Text(gtx, theme.TreeTextRequest{Text: descriptionText, Size: metrics.ItemDescriptionSize, Color: itemVisual.Description})
		}),
	)
}

func treeSemanticDescription(ctx *core.Context, item Item, expanded bool) string {
	parts := make([]string, 0, 2)
	if description := treeItemDescription(item); description != "" {
		parts = append(parts, description)
	}
	if treeItemExpandable(item) {
		state := ctx.Text(i18n.TreeCollapsed)
		if expanded {
			state = ctx.Text(i18n.TreeExpanded)
		}
		parts = append(parts, state)
	}
	loadingState := ""
	switch item.ChildrenState {
	case ChildrenUnloaded:
		loadingState = ctx.Text(i18n.TreeChildrenNotLoaded)
	case ChildrenLoading:
		loadingState = ctx.Text(i18n.TreeLoadingChildren)
	case ChildrenError:
		loadingState = ctx.Text(i18n.TreeLoadChildrenFailed)
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
