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

// Widget presents hierarchical, expandable data with controlled selection and expansion.
type Widget struct {
	key              string
	selectedKey      string
	items            []Item
	expandedKeys     []string
	disabledKeys     []string
	emptyText        string
	onChange         func(string)
	onExpandedChange func([]string)
	onAction         func(string)
	variant          Variant
	selectionMode    SelectionMode
	disabled         bool
	allowEmpty       bool
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

// Variant sets the container treatment.
func (t Widget) Variant(variant Variant) Widget {
	t.variant = variant
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
	state.checkItems(t.items)
	visible := flattenVisibleItems(t.items, treeKeySet(t.expandedKeys))

	if !t.disabled {
		for _, entry := range visible {
			itemState := state.item(entry.item.Key)
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
