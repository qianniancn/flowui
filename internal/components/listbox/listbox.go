package listbox

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type ListBoxItem struct {
	Key         string
	Label       string
	Description string
	Leading     frame.Widget
	Trailing    frame.Widget
	Indicator   func(selected bool) frame.Widget
	Disabled    bool
	Variant     ListBoxItemVariant
}

type ListBoxSection struct {
	Title string
	Items []ListBoxItem
}

type ListBoxItemVariant int

const (
	ListBoxItemDefault ListBoxItemVariant = iota
	ListBoxItemDanger
)

type ListBoxSelectionMode int

const (
	ListBoxSelectionSingle ListBoxSelectionMode = iota
	ListBoxSelectionMultiple
	ListBoxSelectionNone
)

type ListBoxWidget struct {
	key               string
	derivedOwner      string
	derivedRole       string
	selectedKey       string
	selectedKeys      []string
	items             []ListBoxItem
	sections          []ListBoxSection
	dataVersion       uint64
	hasDataVersion    bool
	emptyText         string
	onChange          func(string)
	onSelectionChange func([]string)
	onAction          func(string)
	selectionMode     ListBoxSelectionMode
	disabledKeys      []string
	disabled          bool
	fullWidth         bool
	allowEmpty        bool
	hideIndicator     bool
	maxHeight         int
	padding           unit.Dp
	hasPadding        bool
}

func (l ListBoxWidget) withPadding(padding unit.Dp) ListBoxWidget {
	l.padding = max(padding, 0)
	l.hasPadding = true
	return l
}

func WithPadding(widget ListBoxWidget, padding unit.Dp) ListBoxWidget {
	return widget.withPadding(padding)
}

// WithDerivedIdentity binds an embedded ListBox to an already-resolved owner
// identity without exposing a user-key-shaped internal key.
func WithDerivedIdentity(widget ListBoxWidget, owner, role string) ListBoxWidget {
	widget.derivedOwner = owner
	widget.derivedRole = role
	return widget
}

const (
	listBoxItemColorDuration    = 100 * time.Millisecond
	listBoxItemSelectDuration   = 200 * time.Millisecond
	listBoxItemFocusDuration    = 100 * time.Millisecond
	listBoxItemPressInDuration  = 80 * time.Millisecond
	listBoxItemPressOutDuration = 140 * time.Millisecond
)

func ListBox(key, selectedKey string, items []ListBoxItem) ListBoxWidget {
	return ListBoxWidget{
		key:           key,
		selectedKey:   selectedKey,
		items:         items,
		emptyText:     "No items",
		selectionMode: ListBoxSelectionSingle,
	}
}

func ListBoxMultiple(key string, selectedKeys []string, items []ListBoxItem) ListBoxWidget {
	return ListBoxWidget{
		key:           key,
		selectedKeys:  selectedKeys,
		items:         items,
		emptyText:     "No items",
		selectionMode: ListBoxSelectionMultiple,
	}
}

func ListBoxSections(key, selectedKey string, sections []ListBoxSection) ListBoxWidget {
	return ListBoxWidget{
		key:           key,
		selectedKey:   selectedKey,
		sections:      sections,
		emptyText:     "No items",
		selectionMode: ListBoxSelectionSingle,
	}
}

func ListBoxMultipleSections(key string, selectedKeys []string, sections []ListBoxSection) ListBoxWidget {
	return ListBoxWidget{
		key:           key,
		selectedKeys:  selectedKeys,
		sections:      sections,
		emptyText:     "No items",
		selectionMode: ListBoxSelectionMultiple,
	}
}

func (l ListBoxWidget) EmptyText(text string) ListBoxWidget {
	l.emptyText = text
	return l
}

// DataVersion enables validation and flattened-data reuse. Increase version
// whenever the item data or section structure changes.
func (l ListBoxWidget) DataVersion(version uint64) ListBoxWidget {
	l.dataVersion = version
	l.hasDataVersion = true
	return l
}

func (l ListBoxWidget) OnChange(fn func(string)) ListBoxWidget {
	l.onChange = fn
	return l
}

func (l ListBoxWidget) OnSelectionChange(fn func([]string)) ListBoxWidget {
	l.onSelectionChange = fn
	return l
}

func (l ListBoxWidget) OnAction(fn func(string)) ListBoxWidget {
	l.onAction = fn
	return l
}

func (l ListBoxWidget) SelectionMode(mode ListBoxSelectionMode) ListBoxWidget {
	l.selectionMode = mode
	return l
}

func (l ListBoxWidget) DisabledKeys(keys []string) ListBoxWidget {
	l.disabledKeys = keys
	return l
}

func (l ListBoxWidget) Disabled(disabled bool) ListBoxWidget {
	l.disabled = disabled
	return l
}

func (l ListBoxWidget) FullWidth() ListBoxWidget {
	l.fullWidth = true
	return l
}

func (l ListBoxWidget) AllowEmptySelection() ListBoxWidget {
	l.allowEmpty = true
	return l
}

func (l ListBoxWidget) HideIndicator() ListBoxWidget {
	l.hideIndicator = true
	return l
}

func (l ListBoxWidget) MaxHeight(dp int) ListBoxWidget {
	l.maxHeight = dp
	return l
}

func (l ListBoxWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := l.stateFor(ctx)
	state.beginFrame()
	defer state.endFrame()
	entries, items := state.resolveEntries(l)

	if !l.disabled {
		result := state.updateKeys(gtx, items, l.disabledKeys, state.keyboardActiveKey(l))
		if result.focusKey != "" {
			frame.RequestFocus(ctx, &state.item(result.focusKey).Clickable)
		}
		if result.actionKey != "" {
			l.activate(result.actionKey)
		}
	}
	if l.disabled {
		gtx = gtx.Disabled()
	}
	return l.layout(ctx, gtx, state, entries, len(items) > 0)
}

func (l ListBoxWidget) activate(key string) {
	if l.onAction != nil {
		l.onAction(key)
	}
	switch l.selectionMode {
	case ListBoxSelectionNone:
		return
	case ListBoxSelectionMultiple:
		nextKeys := listBoxToggleSelectedKeys(l.selectedKeys, key)
		if !listBoxSameKeys(nextKeys, l.selectedKeys) && l.onSelectionChange != nil {
			l.onSelectionChange(nextKeys)
		}
		return
	default:
		nextKey := key
		if l.allowEmpty && key == l.selectedKey {
			nextKey = ""
		}
		if nextKey != l.selectedKey && l.onChange != nil {
			l.onChange(nextKey)
		}
	}
}

func (l ListBoxWidget) isSelected(key string) bool {
	switch l.selectionMode {
	case ListBoxSelectionNone:
		return false
	case ListBoxSelectionMultiple:
		return listBoxContainsKey(l.selectedKeys, key)
	default:
		return key == l.selectedKey
	}
}

func (l ListBoxWidget) allItems() []ListBoxItem {
	_, items := l.entriesAndItems()
	return items
}

func (l ListBoxWidget) entriesAndItems() ([]listBoxEntry, []ListBoxItem) {
	if len(l.sections) == 0 {
		entries := make([]listBoxEntry, 0, len(l.items))
		for _, item := range l.items {
			entries = append(entries, listBoxEntry{item: item})
		}
		return entries, l.items
	}
	count := 0
	for _, section := range l.sections {
		count += len(section.Items)
	}
	entries := make([]listBoxEntry, 0, count+len(l.sections))
	items := make([]ListBoxItem, 0, count)
	for _, section := range l.sections {
		if len(section.Items) == 0 {
			continue
		}
		if section.Title != "" {
			entries = append(entries, listBoxEntry{
				header: true,
				title:  section.Title,
			})
		}
		items = append(items, section.Items...)
		for _, item := range section.Items {
			entries = append(entries, listBoxEntry{item: item})
		}
	}
	return entries, items
}

func (l ListBoxWidget) hasItems() bool {
	if len(l.sections) == 0 {
		return len(l.items) > 0
	}
	for _, section := range l.sections {
		if len(section.Items) > 0 {
			return true
		}
	}
	return false
}

func (l ListBoxWidget) itemDisabled(item ListBoxItem) bool {
	return l.disabled || item.Disabled || listBoxContainsKey(l.disabledKeys, item.Key)
}

func listBoxToggleSelectedKeys(selectedKeys []string, key string) []string {
	nextKeys := make([]string, 0, len(selectedKeys)+1)
	removed := false
	for _, selectedKey := range selectedKeys {
		if selectedKey == key {
			removed = true
			continue
		}
		if listBoxContainsKey(nextKeys, selectedKey) {
			continue
		}
		nextKeys = append(nextKeys, selectedKey)
	}
	if removed {
		return nextKeys
	}
	nextKeys = append(nextKeys, key)
	return nextKeys
}

func listBoxContainsKey(keys []string, key string) bool {
	for _, current := range keys {
		if current == key {
			return true
		}
	}
	return false
}

func listBoxSameKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
