package tree

import (
	"fmt"
	"image"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/menu"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// Item describes one node in a Tree.
type Item struct {
	Key             string
	Label           string
	Description     string
	Leading         frame.Widget
	ExpandedLeading frame.Widget
	Trailing        frame.Widget
	Children        []Item
	ChildrenState   ChildrenState
	LoadError       string
	AcceptsChildren bool
	Renamable       bool
	Disabled        bool
}

// ChildrenState describes controlled asynchronous child loading.
type ChildrenState uint8

const (
	ChildrenLoaded ChildrenState = iota
	ChildrenUnloaded
	ChildrenLoading
	ChildrenError
)

// Variant selects the Tree container treatment.
type Variant uint8

const (
	VariantDefault Variant = iota
	VariantSurface
)

// SelectionMode controls whether rows can be selected.
type SelectionMode uint8

const (
	SelectionSingle SelectionMode = iota
	SelectionMultiple
	SelectionNone
)

// Size selects the Tree density preset.
type Size uint8

const (
	SizeMedium Size = iota
	SizeSmall
)

// GuideStyle selects the indentation guide stroke pattern.
type GuideStyle uint8

const (
	GuideSolid GuideStyle = iota
	GuideDashed
)

// DropPosition describes where a dragged item was dropped relative to its target.
type DropPosition uint8

const (
	DropBefore DropPosition = iota
	DropInside
	DropAfter
)

// DropEvent describes a requested move within one Tree.
type DropEvent struct {
	// SourceKey is the primary dragged item and is retained for single-item handlers.
	SourceKey string
	// SourceKeys contains all dragged items in tree order.
	SourceKeys []string
	TargetKey  string
	Position   DropPosition
}

// Widget presents hierarchical, expandable data with controlled selection and expansion.
type Widget struct {
	theme             func(*theme.Theme)
	key               string
	selectedKey       string
	selectedKeys      []string
	selectedKeySet    stateutil.StringSet
	items             []Item
	dataVersion       uint64
	hasDataVersion    bool
	expandedKeys      []string
	disabledKeys      []string
	disabledKeySet    stateutil.StringSet
	emptyText         string
	onChange          func(string)
	onSelectionChange func([]string)
	onExpandedChange  func([]string)
	onAction          func(string)
	onDrop            func(DropEvent)
	canDrop           func(DropEvent) bool
	onLoadChildren    func(string)
	onRename          func(string, string)
	renameRequestKey  string
	renameRequest     uint64
	contextMenu       menu.Widget
	hasContextMenu    bool
	onContextMenu     func(string)
	variant           Variant
	size              Size
	selectionMode     SelectionMode
	disabled          bool
	allowEmpty        bool
	guides            bool
	guideConnectors   bool
	guideStyle        GuideStyle
	expandOnRowClick  bool
	maxHeight         int
}

// New creates a controlled Tree.
func New(key, selectedKey string, items []Item) Widget {
	return Widget{
		key:         key,
		selectedKey: selectedKey,
		items:       items,
		emptyText:   "No items",
	}
}

// DataVersion enables validation and flattened-data reuse. Increase version
// whenever the item hierarchy or item content changes.
func (t Widget) DataVersion(version uint64) Widget {
	t.dataVersion = version
	t.hasDataVersion = true
	return t
}

// ExpandedKeys sets the controlled set of expanded branch keys.
func (t Widget) ExpandedKeys(keys []string) Widget {
	t.expandedKeys = keys
	return t
}

// OnChange registers a callback for selection changes.
func (t Widget) OnChange(fn func(string)) Widget {
	t.onChange = fn
	return t
}

// SelectedKeys sets the controlled selection for SelectionMultiple.
func (t Widget) SelectedKeys(keys []string) Widget {
	t.selectedKeys = keys
	return t
}

// OnSelectionChange registers a callback for multiple-selection changes.
func (t Widget) OnSelectionChange(fn func([]string)) Widget {
	t.onSelectionChange = fn
	return t
}

// OnExpandedChange registers a callback for expansion changes.
func (t Widget) OnExpandedChange(fn func([]string)) Widget {
	t.onExpandedChange = fn
	return t
}

// OnAction registers a callback for row activation, independent of selection mode.
func (t Widget) OnAction(fn func(string)) Widget {
	t.onAction = fn
	return t
}

// OnDrop enables item dragging and reports a requested move.
func (t Widget) OnDrop(fn func(DropEvent)) Widget {
	t.onDrop = fn
	return t
}

// CanDrop validates a proposed drop for both indicator and delivery. It may be
// called every frame while dragging and must not mutate application state.
func (t Widget) CanDrop(fn func(DropEvent) bool) Widget {
	t.canDrop = fn
	return t
}

// OnLoadChildren handles initial loading and retries for asynchronous branches.
func (t Widget) OnLoadChildren(fn func(string)) Widget {
	t.onLoadChildren = fn
	return t
}

// OnRename handles committed inline label edits for Renamable items.
func (t Widget) OnRename(fn func(string, string)) Widget {
	t.onRename = fn
	return t
}

// RequestRename starts inline editing once when revision changes.
func (t Widget) RequestRename(key string, revision uint64) Widget {
	t.renameRequestKey = key
	t.renameRequest = revision
	return t
}

// ContextMenu sets the menu opened from an enabled item row.
func (t Widget) ContextMenu(content menu.Widget) Widget {
	t.contextMenu = content
	t.hasContextMenu = true
	return t
}

// OnContextMenu reports the item whose context menu is opening.
func (t Widget) OnContextMenu(fn func(string)) Widget {
	t.onContextMenu = fn
	return t
}

// Variant sets the container treatment.
func (t Widget) Variant(variant Variant) Widget {
	t.variant = variant
	return t
}

// Size sets the Tree density preset.
func (t Widget) Size(size Size) Widget {
	t.size = size
	return t
}

// SelectionMode controls whether activation changes the selected key.
func (t Widget) SelectionMode(mode SelectionMode) Widget {
	t.selectionMode = mode
	return t
}

// DisabledKeys disables individual nodes by key.
func (t Widget) DisabledKeys(keys []string) Widget {
	t.disabledKeys = keys
	return t
}

// Disabled disables all Tree interaction.
func (t Widget) Disabled(disabled bool) Widget {
	t.disabled = disabled
	return t
}

// AllowEmptySelection allows activating the selected row to clear selection.
func (t Widget) AllowEmptySelection() Widget {
	t.allowEmpty = true
	return t
}

// Guides controls whether expanded branches draw indentation guides.
func (t Widget) Guides(enabled bool) Widget {
	t.guides = enabled
	return t
}

// GuideConnectors controls whether indentation guides connect horizontally to child content.
func (t Widget) GuideConnectors(enabled bool) Widget {
	t.guideConnectors = enabled
	return t
}

// GuideStyle sets the indentation guide stroke pattern.
func (t Widget) GuideStyle(style GuideStyle) Widget {
	t.guideStyle = style
	return t
}

// ExpandOnRowClick controls whether clicking a branch row also toggles it.
func (t Widget) ExpandOnRowClick(enabled bool) Widget {
	t.expandOnRowClick = enabled
	return t
}

// EmptyText sets the message displayed when the Tree has no nodes.
func (t Widget) EmptyText(value string) Widget {
	t.emptyText = value
	return t
}

// MaxHeight overrides the theme maximum height in dp.
func (t Widget) MaxHeight(dp int) Widget {
	t.maxHeight = max(dp, 0)
	return t
}

func (t Widget) Theme(fn func(*theme.Theme)) Widget {
	t.theme = fn
	return t
}

func (t Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, t.theme); restore != nil {
		defer restore()
	}
	state := treeStateFor(ctx, t.key)
	state.beginFrame()
	defer state.endFrame()
	t.selectedKeySet = state.selectedKeys.Resolve(t.selectedKeys)
	t.disabledKeySet = state.disabledKeys.Resolve(t.disabledKeys)
	visible := state.resolveVisible(t)
	if state.renameKey != "" && treeVisibleIndex(visible, state.renameKey) < 0 {
		state.finishRename(t, false)
	}
	state.applyRenameRequest(ctx, t, visible)
	state.updateRename(gtx, t)
	state.dragSource = ""
	state.dragSources = state.dragSources[:0]
	state.dropTarget = treeDropTarget{}
	var dragSelection []string
	if t.onDrop != nil && t.selectionMode == SelectionMultiple {
		dragSelection = t.selectedDragKeys(visible)
	}

	if !t.disabled {
		dragIndex := -1
		for itemKey, itemState := range state.items {
			index := state.visibleIndex(visible, itemKey, itemState)
			if index < 0 {
				continue
			}
			entry := visible[index]
			disabled := t.itemDisabled(entry.item)
			presses := stateutil.ActivePresses(itemState.clickable.History())
			for itemState.toggle.Clicked(gtx) {
				if !disabled && treeItemExpandable(entry.item) {
					t.requestItemToggle(entry.item)
				}
			}
			for {
				click, ok := itemState.clickable.Update(gtx)
				if !ok {
					break
				}
				if disabled {
					continue
				}
				t.activateWithModifiers(state, visible, entry.item.Key, click.Modifiers)
				selectionModifier := click.Modifiers.Contain(key.ModShift) || click.Modifiers.Contain(key.ModCtrl) || click.Modifiers.Contain(key.ModCommand)
				if t.expandOnRowClick && !selectionModifier && treeItemExpandable(entry.item) {
					t.requestItemToggle(entry.item)
				}
			}
			if !disabled {
				frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
			}
			if t.onDrop != nil && !disabled {
				t.updateDropEvents(gtx, state, itemState, visible, entry)
				sources := []string{entry.item.Key}
				if t.selectionMode == SelectionMultiple && treeContainsKey(t.selectedKeys, entry.item.Key) && len(dragSelection) > 0 {
					sources = dragSelection
				}
				if itemState.updateDrag(gtx, state.dragMIME, sources) && dragIndex < 0 {
					state.dragSource = entry.item.Key
					state.dragSources = append(state.dragSources[:0], sources...)
					dragIndex = index
				}
			}
		}
		if dragIndex >= 0 {
			tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
			heights := make([]int, len(visible))
			for index, entry := range visible {
				heights[index] = treeItemHeight(gtx, tokens, entry.item)
			}
			sourceState := state.item(state.dragSource)
			pointerY := sourceState.dragPress.Y + sourceState.drag.Pos().Y
			target := treeDropTargetAt(visible, dragIndex, pointerY, heights, gtx.Dp(tokens.Gap))
			viewportHeight := gtx.Constraints.Max.Y - 2*gtx.Dp(tokens.Padding)
			maxHeight := tokens.MaxHeight
			if t.maxHeight > 0 {
				maxHeight = unit.Dp(t.maxHeight)
			}
			if maxHeight > 0 {
				viewportHeight = min(viewportHeight, gtx.Dp(maxHeight)-2*gtx.Dp(tokens.Padding))
			}
			pointerViewportY := treeDragViewportY(heights, gtx.Dp(tokens.Gap), state.list.Position, dragIndex, pointerY)
			state.updateDragScroll(gtx, pointerViewportY, max(viewportHeight, 0), gtx.Dp(unit.Dp(24)), len(visible))
			if t.dropAllowed(visible, state.dragSources, target.key, target.position) {
				state.dropTarget = treeDropIndicatorTarget(visible, target)
			}
			targetIndex := treeVisibleIndex(visible, state.dropTarget.key)
			hoverExpandable := targetIndex >= 0 && state.dropTarget.position == DropInside && treeItemExpandable(visible[targetIndex].item) && !treeContainsKey(t.expandedKeys, state.dropTarget.key)
			if state.updateDragHover(gtx, state.dropTarget.key, hoverExpandable) {
				t.requestItemToggle(visible[targetIndex].item)
			}
		} else {
			state.resetDragAssist()
		}
		result := state.updateKeys(gtx, t, visible)
		if result.focusKey != "" {
			if result.rangeKey == "" && t.selectionMode == SelectionMultiple {
				state.selectionAnchor = result.focusKey
			}
			frame.RequestFocus(ctx, &state.item(result.focusKey).clickable)
			state.ensureVisible(treeVisibleIndex(visible, result.focusKey))
		}
		if result.loadKey != "" {
			t.requestLoad(result.loadKey)
		}
		if result.toggleKey != "" {
			t.requestToggle(result.toggleKey)
		}
		if result.renameKey != "" {
			index := treeVisibleIndex(visible, result.renameKey)
			if index >= 0 {
				state.beginRename(visible[index].item)
				frame.RequestFocus(ctx, &state.renameEditor)
			}
		}
		if result.rangeKey != "" {
			t.activateWithModifiers(state, visible, result.rangeKey, key.ModShift)
		}
		if result.actionKey != "" {
			t.activateWithModifiers(state, visible, result.actionKey, result.actionModifiers)
		}
	}
	dims := t.layout(ctx, gtx, state, visible)
	if state.dragSource != "" {
		dragIndex := treeVisibleIndex(visible, state.dragSource)
		if dragIndex >= 0 && !treeListPositionContains(state.list.Position, dragIndex) {
			tokens := treeTokensFor(frame.ActiveTheme(ctx), t.size)
			heights := make([]int, len(visible))
			for index, entry := range visible {
				heights[index] = treeItemHeight(gtx, tokens, entry.item)
			}
			sourceState := state.item(state.dragSource)
			sourceTop := treeDragViewportY(heights, gtx.Dp(tokens.Gap), state.list.Position, dragIndex, 0)
			press := sourceState.dragPress.Round()
			previewOffset := image.Pt(
				gtx.Dp(tokens.Padding)+press.X+gtx.Dp(unit.Dp(12)),
				gtx.Dp(tokens.Padding)+int(sourceTop)+press.Y+gtx.Dp(unit.Dp(12)),
			)
			label := visible[dragIndex].item.Label
			if count := len(state.dragSources); count > 1 {
				label = fmt.Sprintf("%s +%d", label, count-1)
			}
			offscreenGtx := gtx
			offscreenGtx.Constraints.Min = image.Point{}
			sourceState.drag.Layout(
				offscreenGtx,
				func(layout.Context) layout.Dimensions { return layout.Dimensions{Size: image.Pt(1, 1)} },
				t.dragPreview(ctx, label, previewOffset),
			)
		}
	}
	return dims
}

func (t Widget) activateWithModifiers(state *treeState, visible []flatItem, itemKey string, modifiers key.Modifiers) {
	rangeSelection := t.selectionMode == SelectionMultiple && modifiers.Contain(key.ModShift) && state != nil
	if t.onAction != nil && !rangeSelection {
		t.onAction(itemKey)
	}
	switch t.selectionMode {
	case SelectionSingle:
		next := itemKey
		if t.allowEmpty && itemKey == t.selectedKey {
			next = ""
		}
		if next != t.selectedKey && t.onChange != nil {
			t.onChange(next)
		}
	case SelectionMultiple:
		var next []string
		switch {
		case rangeSelection:
			next = t.rangeKeys(visible, state.selectionAnchor, itemKey)
		case modifiers.Contain(key.ModCtrl) || modifiers.Contain(key.ModCommand):
			state.selectionAnchor = itemKey
			next = toggleTreeKey(t.selectedKeys, itemKey)
		default:
			state.selectionAnchor = itemKey
			if t.allowEmpty && len(t.selectedKeys) == 1 && t.selectedKeys[0] == itemKey {
				next = nil
			} else {
				next = []string{itemKey}
			}
		}
		if !treeKeysEqual(next, t.selectedKeys) && t.onSelectionChange != nil {
			t.onSelectionChange(next)
		}
	}
}

func (t Widget) requestToggle(key string) {
	if t.onExpandedChange != nil {
		t.onExpandedChange(toggleTreeKey(t.expandedKeys, key))
	}
}

func (t Widget) requestItemToggle(item Item) {
	expanded := treeContainsKey(t.expandedKeys, item.Key)
	if item.ChildrenState == ChildrenError {
		t.requestLoad(item.Key)
		if !expanded {
			t.requestToggle(item.Key)
		}
		return
	}
	if item.ChildrenState == ChildrenUnloaded && !expanded {
		t.requestLoad(item.Key)
	}
	t.requestToggle(item.Key)
}

func (t Widget) requestLoad(key string) {
	if t.onLoadChildren != nil {
		t.onLoadChildren(key)
	}
}

func (t Widget) itemDisabled(item Item) bool {
	return t.disabled || item.Disabled || stateutil.StringSetContains(t.disabledKeys, t.disabledKeySet, item.Key)
}

func (t Widget) itemSelected(key string) bool {
	if t.selectionMode == SelectionNone {
		return false
	}
	if t.selectionMode == SelectionMultiple {
		return stateutil.StringSetContains(t.selectedKeys, t.selectedKeySet, key)
	}
	return key == t.selectedKey
}

func (t Widget) keyboardActiveIndex(visible []flatItem) int {
	if t.selectionMode == SelectionMultiple {
		for index := len(t.selectedKeys) - 1; index >= 0; index-- {
			visibleIndex := treeVisibleIndex(visible, t.selectedKeys[index])
			if visibleIndex >= 0 && !t.itemDisabled(visible[visibleIndex].item) {
				return visibleIndex
			}
		}
		return -1
	}
	index := treeVisibleIndex(visible, t.selectedKey)
	if index >= 0 && !t.itemDisabled(visible[index].item) {
		return index
	}
	return -1
}

func (t Widget) rangeKeys(visible []flatItem, anchor, target string) []string {
	anchorIndex := treeVisibleIndex(visible, anchor)
	targetIndex := treeVisibleIndex(visible, target)
	if anchorIndex < 0 {
		for _, selected := range t.selectedKeys {
			if index := treeVisibleIndex(visible, selected); index >= 0 {
				anchorIndex = index
				break
			}
		}
	}
	if targetIndex < 0 {
		return append([]string(nil), t.selectedKeys...)
	}
	if anchorIndex < 0 {
		anchorIndex = targetIndex
	}
	start, end := min(anchorIndex, targetIndex), max(anchorIndex, targetIndex)
	next := make([]string, 0, end-start+1)
	for _, selected := range t.selectedKeys {
		if index := treeVisibleIndex(visible, selected); index >= 0 && t.itemDisabled(visible[index].item) {
			next = append(next, selected)
		}
	}
	for index := start; index <= end; index++ {
		if !t.itemDisabled(visible[index].item) && !treeContainsKey(next, visible[index].item.Key) {
			next = append(next, visible[index].item.Key)
		}
	}
	return next
}

func (t Widget) selectedDragKeys(visible []flatItem) []string {
	if t.selectionMode != SelectionMultiple || len(t.selectedKeys) == 0 {
		return nil
	}
	selected := treeKeySet(t.selectedKeys)
	result := make([]string, 0, len(selected))
	for _, entry := range visible {
		if _, ok := selected[entry.item.Key]; !ok || t.itemDisabled(entry.item) {
			continue
		}
		nested := false
		for parent := entry.parentKey; parent != ""; {
			if _, ok := selected[parent]; ok {
				nested = true
				break
			}
			index := treeVisibleIndex(visible, parent)
			if index < 0 {
				break
			}
			parent = visible[index].parentKey
		}
		if !nested {
			result = append(result, entry.item.Key)
		}
	}
	return result
}
