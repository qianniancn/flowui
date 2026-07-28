package listbox

import (
	"slices"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
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
	key                 string
	derivedOwner        string
	derivedRole         string
	selectedKey         string
	hasSelectedKey      bool
	defaultSelectedKey  string
	hasDefaultKey       bool
	selectedKeys        []string
	hasSelectedKeys     bool
	defaultSelectedKeys []string
	hasDefaultKeys      bool
	selectedKeySet      stateutil.StringSet
	items               []ListBoxItem
	sections            []ListBoxSection
	dataVersion         uint64
	hasDataVersion      bool
	emptyText           string
	onChange            func(string)
	onSelectionChange   func([]string)
	onAction            func(string)
	selectionMode       ListBoxSelectionMode
	disabledKeys        []string
	disabledKeySet      stateutil.StringSet
	disabled            bool
	fullWidth           bool
	allowEmpty          bool
	hideIndicator       bool
	maxHeight           int
	padding             unit.Dp
	hasPadding          bool
	customStyle         flowstyle.Style
	partsStyle          flowstyle.Style
}

func (l ListBoxWidget) withPadding(padding unit.Dp) ListBoxWidget {
	l.padding = max(padding, 0)
	l.hasPadding = true
	return l
}

func WithPadding(widget ListBoxWidget, padding unit.Dp) ListBoxWidget {
	return widget.withPadding(padding)
}

// WithPartsStyle supplies compound-owner parts without applying the owner's
// root declaration to the embedded ListBox root.
func WithPartsStyle(widget ListBoxWidget, value flowstyle.Style) ListBoxWidget {
	widget.partsStyle = value
	return widget
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
		key:            key,
		selectedKey:    selectedKey,
		hasSelectedKey: true,
		items:          items,
		emptyText:      "No items",
		selectionMode:  ListBoxSelectionSingle,
	}
}

func ListBoxMultiple(key string, selectedKeys []string, items []ListBoxItem) ListBoxWidget {
	return ListBoxWidget{
		key:             key,
		selectedKeys:    selectedKeys,
		hasSelectedKeys: true,
		items:           items,
		emptyText:       "No items",
		selectionMode:   ListBoxSelectionMultiple,
	}
}

func ListBoxSections(key, selectedKey string, sections []ListBoxSection) ListBoxWidget {
	return ListBoxWidget{
		key:            key,
		selectedKey:    selectedKey,
		hasSelectedKey: true,
		sections:       sections,
		emptyText:      "No items",
		selectionMode:  ListBoxSelectionSingle,
	}
}

func ListBoxMultipleSections(key string, selectedKeys []string, sections []ListBoxSection) ListBoxWidget {
	return ListBoxWidget{
		key:             key,
		selectedKeys:    selectedKeys,
		hasSelectedKeys: true,
		sections:        sections,
		emptyText:       "No items",
		selectionMode:   ListBoxSelectionMultiple,
	}
}

func (l ListBoxWidget) EmptyText(text string) ListBoxWidget {
	l.emptyText = text
	return l
}

// SelectedKey sets controlled mode with an external single selection value.
func (l ListBoxWidget) SelectedKey(key string) ListBoxWidget {
	l.selectedKey = key
	l.hasSelectedKey = true
	return l
}

// DefaultSelectedKey sets uncontrolled mode with an initial single selection value.
func (l ListBoxWidget) DefaultSelectedKey(key string) ListBoxWidget {
	l.defaultSelectedKey = key
	l.hasDefaultKey = true
	l.hasSelectedKey = false
	return l
}

// SelectedKeys sets controlled mode with external multiple selection values.
func (l ListBoxWidget) SelectedKeys(keys []string) ListBoxWidget {
	l.selectedKeys = keys
	l.hasSelectedKeys = true
	return l
}

// DefaultSelectedKeys sets uncontrolled mode with initial multiple selection values.
func (l ListBoxWidget) DefaultSelectedKeys(keys []string) ListBoxWidget {
	l.defaultSelectedKeys = keys
	l.hasDefaultKeys = true
	l.hasSelectedKeys = false
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

func (l ListBoxWidget) Style(value flowstyle.Style) ListBoxWidget {
	l.customStyle = value
	return l
}

func (l ListBoxWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	rootKey := frame.FullKey(ctx, l.key)
	if l.derivedOwner != "" {
		rootKey = frame.DerivedKey(ctx, l.derivedOwner, l.derivedRole)
	}
	state := l.stateFor(ctx)
	state.beginFrame()
	defer state.endFrame()

	// Bind disclosure and get current values
	if l.selectionMode == ListBoxSelectionMultiple {
		state.bindMultiple(l)
		l.selectedKeys = state.currentMultipleValue(l)
	} else if l.selectionMode == ListBoxSelectionSingle {
		state.bindSingle(l)
		l.selectedKey = state.currentSingleValue(l)
	}

	l.selectedKeySet = state.selectedKeys.Resolve(l.selectedKeys)
	l.disabledKeySet = state.disabledKeys.Resolve(l.disabledKeys)
	entries, items := state.resolveEntries(l)

	if !l.disabled {
		result := state.updateKeys(gtx, items, l, state.keyboardActiveKey(l))
		if result.focusKey != "" {
			frame.RequestFocus(ctx, &state.item(result.focusKey).Clickable)
		}
		if result.actionKey != "" {
			l.activate(state, result.actionKey)
		}
	}
	if l.disabled {
		gtx = gtx.Disabled()
	}
	hovered, pressed, focused := false, false, false
	for _, item := range items {
		itemState := state.item(item.Key)
		hovered = hovered || itemState.Clickable.Hovered()
		pressed = pressed || itemState.Clickable.Pressed()
		focused = focused || gtx.Focused(&itemState.Clickable)
	}
	selected := l.selectedKey != "" || len(l.selectedKeys) > 0
	return layoutui.LayoutStyled(ctx, gtx, rootKey, flowstyle.StyleState{
		Hovered:  hovered,
		Pressed:  pressed,
		Focused:  focused,
		Disabled: l.disabled || !gtx.Enabled(),
		Selected: selected,
	}, l.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return l.layout(ctx, gtx, rootKey, state, entries, len(items) > 0)
	}))
}

func (l ListBoxWidget) activate(state *listBoxState, key string) {
	if l.onAction != nil {
		l.onAction(key)
	}
	switch l.selectionMode {
	case ListBoxSelectionNone:
		return
	case ListBoxSelectionMultiple:
		nextKeys := listBoxToggleSelectedKeys(l.selectedKeys, key)
		if !listBoxSameKeys(nextKeys, l.selectedKeys) {
			state.requestMultipleValue(l, nextKeys)
		}
		return
	default:
		nextKey := key
		if l.allowEmpty && key == l.selectedKey {
			nextKey = ""
		}
		if nextKey != l.selectedKey {
			state.requestSingleValue(l, nextKey)
		}
	}
}

func (l ListBoxWidget) isSelected(key string) bool {
	switch l.selectionMode {
	case ListBoxSelectionNone:
		return false
	case ListBoxSelectionMultiple:
		return stateutil.StringSetContains(l.selectedKeys, l.selectedKeySet, key)
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

func (l ListBoxWidget) itemDisabled(item ListBoxItem) bool {
	return l.disabled || item.Disabled || stateutil.StringSetContains(l.disabledKeys, l.disabledKeySet, item.Key)
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
	return slices.Contains(keys, key)
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
