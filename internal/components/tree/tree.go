package tree

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// Item describes one node in a Tree.
type Item struct {
	Key         string
	Label       string
	Description string
	Leading     frame.Widget
	Trailing    frame.Widget
	Children    []Item
	Disabled    bool
}

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
	SourceKey string
	TargetKey string
	Position  DropPosition
}

// Widget presents hierarchical, expandable data with controlled selection and expansion.
type Widget struct {
	key              string
	selectedKey      string
	items            []Item
	dataVersion      uint64
	hasDataVersion   bool
	expandedKeys     []string
	disabledKeys     []string
	emptyText        string
	onChange         func(string)
	onExpandedChange func([]string)
	onAction         func(string)
	onDrop           func(DropEvent)
	variant          Variant
	size             Size
	selectionMode    SelectionMode
	disabled         bool
	allowEmpty       bool
	guides           bool
	guideConnectors  bool
	guideStyle       GuideStyle
	maxHeight        int
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

func (t Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := treeStateFor(ctx, t.key)
	state.beginFrame()
	defer state.endFrame()
	visible := state.resolveVisible(t)
	state.dragSource = ""
	state.dropTarget = treeDropTarget{}

	if !t.disabled {
		dragIndex := -1
		for key, itemState := range state.items {
			index := state.visibleIndex(visible, key, itemState)
			if index < 0 {
				continue
			}
			entry := visible[index]
			disabled := t.itemDisabled(entry.item)
			for itemState.toggle.Clicked(gtx) {
				if !disabled && len(entry.item.Children) > 0 {
					t.requestToggle(entry.item.Key)
				}
			}
			for itemState.clickable.Clicked(gtx) {
				if !disabled {
					t.activate(entry.item.Key)
				}
			}
			if t.onDrop != nil && !disabled {
				t.updateDropEvents(gtx, state, visible, entry)
				if itemState.updateDrag(gtx, state.dragMIME, entry.item.Key) && dragIndex < 0 {
					state.dragSource = entry.item.Key
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
			if treeDropAllowed(t, visible, state.dragSource, target.key) {
				state.dropTarget = treeDropIndicatorTarget(visible, target)
			}
		}
		result := state.updateKeys(gtx, t, visible)
		if result.focusKey != "" {
			frame.RequestFocus(ctx, &state.item(result.focusKey).clickable)
			state.ensureVisible(treeVisibleIndex(visible, result.focusKey))
		}
		if result.toggleKey != "" {
			t.requestToggle(result.toggleKey)
		}
		if result.actionKey != "" {
			t.activate(result.actionKey)
		}
	}
	return t.layout(ctx, gtx, state, visible)
}

func (t Widget) activate(key string) {
	if t.onAction != nil {
		t.onAction(key)
	}
	if t.selectionMode == SelectionNone {
		return
	}
	next := key
	if t.allowEmpty && key == t.selectedKey {
		next = ""
	}
	if next != t.selectedKey && t.onChange != nil {
		t.onChange(next)
	}
}

func (t Widget) requestToggle(key string) {
	if t.onExpandedChange != nil {
		t.onExpandedChange(toggleTreeKey(t.expandedKeys, key))
	}
}

func (t Widget) itemDisabled(item Item) bool {
	return t.disabled || item.Disabled || treeContainsKey(t.disabledKeys, item.Key)
}
