package listbox

import (
	"slices"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Item struct {
	Key         string
	Label       string
	Description string
	Leading     core.Widget
	Trailing    core.Widget
	Indicator   func(selected bool) core.Widget
	Disabled    bool
	Variant     ItemVariant
}

type Section struct {
	Title string
	Items []Item
}

type ItemVariant int

const (
	ItemDefault ItemVariant = iota
	ItemDanger
)

type SelectionMode int

const (
	SelectionSingle SelectionMode = iota
	SelectionMultiple
	SelectionNone
)

type Widget struct {
	override            theme.ListBoxOverride
	scrollToKey         string
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
	selectedKeySet      core.StringSet
	items               []Item
	sections            []Section
	dataVersion         uint64
	hasDataVersion      bool
	emptyText           string
	hasEmptyText        bool
	onChange            func(string)
	onSelectionChange   func([]string)
	onAction            func(string)
	selectionMode       SelectionMode
	disabledKeys        []string
	disabledKeySet      core.StringSet
	disabled            bool
	fullWidth           bool
	allowEmpty          bool
	hideIndicator       bool
	maxHeight           int
	padding             unit.Dp
	hasPadding          bool

	partsOnly   bool
	hasOverride bool
}

func (l Widget) withPadding(padding unit.Dp) Widget {
	l.padding = max(padding, 0)
	l.hasPadding = true
	return l
}

func WithPadding(widget Widget, padding unit.Dp) Widget {
	return widget.withPadding(padding)
}

// WithDerivedIdentity binds an embedded ListBox to an already-resolved owner
// identity without exposing a user-key-shaped internal key.
func WithDerivedIdentity(widget Widget, owner, role string) Widget {
	widget.derivedOwner = owner
	widget.derivedRole = role
	return widget
}

// New creates a list box. It starts uncontrolled with no selection; SelectedKey
// sets the controlled value and DefaultSelectedKey the initial one.
func New(items []Item) Widget {
	return Widget{
		key:           core.AutoKey(),
		items:         items,
		selectionMode: SelectionSingle,
	}
}

// Key sets an explicit identity for the list state.
func (l Widget) Key(key string) Widget {
	l.key = key
	return l
}

func Sections(key, selectedKey string, sections []Section) Widget {
	return Widget{
		key:            key,
		selectedKey:    selectedKey,
		hasSelectedKey: true,
		sections:       sections,
		selectionMode:  SelectionSingle,
	}
}

func MultipleSections(key string, selectedKeys []string, sections []Section) Widget {
	return Widget{
		key:             key,
		selectedKeys:    selectedKeys,
		hasSelectedKeys: true,
		sections:        sections,
		selectionMode:   SelectionMultiple,
	}
}

func (l Widget) EmptyText(text string) Widget {
	l.emptyText = text
	l.hasEmptyText = true
	return l
}

// SelectedKey sets controlled mode with an external single selection value.
func (l Widget) SelectedKey(key string) Widget {
	l.selectedKey = key
	l.hasSelectedKey = true
	return l
}

// DefaultSelectedKey sets uncontrolled mode with an initial single selection value.
func (l Widget) DefaultSelectedKey(key string) Widget {
	l.defaultSelectedKey = key
	l.hasDefaultKey = true
	l.hasSelectedKey = false
	return l
}

// SelectedKeys sets controlled mode with external multiple selection values.
// SelectedKeys sets the controlled selection and switches to multiple selection.
func (l Widget) SelectedKeys(keys []string) Widget {
	l.selectedKeys = keys
	l.hasSelectedKeys = true
	l.selectionMode = SelectionMultiple
	return l
}

// DefaultSelectedKeys sets uncontrolled mode with initial multiple selection values.
func (l Widget) DefaultSelectedKeys(keys []string) Widget {
	l.defaultSelectedKeys = keys
	l.hasDefaultKeys = true
	l.hasSelectedKeys = false
	return l
}

// DataVersion enables validation and flattened-data reuse. Increase version
// whenever the item data or section structure changes.
func (l Widget) DataVersion(version uint64) Widget {
	l.dataVersion = version
	l.hasDataVersion = true
	return l
}

func (l Widget) OnChange(fn func(string)) Widget {
	l.onChange = fn
	return l
}

func (l Widget) OnSelectionChange(fn func([]string)) Widget {
	l.onSelectionChange = fn
	return l
}

func (l Widget) OnAction(fn func(string)) Widget {
	l.onAction = fn
	return l
}

func (l Widget) SelectionMode(mode SelectionMode) Widget {
	l.selectionMode = mode
	return l
}

func (l Widget) DisabledKeys(keys []string) Widget {
	l.disabledKeys = keys
	return l
}

func (l Widget) Disabled(disabled bool) Widget {
	l.disabled = disabled
	return l
}

func (l Widget) FullWidth() Widget {
	l.fullWidth = true
	return l
}

func (l Widget) AllowEmptySelection() Widget {
	l.allowEmpty = true
	return l
}

func (l Widget) HideIndicator() Widget {
	l.hideIndicator = true
	return l
}

func (l Widget) MaxHeight(dp int) Widget {
	l.maxHeight = dp
	return l
}

// ScrollToKey reveals an item when its key or index changes, without moving
// keyboard focus. An empty key leaves scrolling under user control.
func (l Widget) ScrollToKey(key string) Widget {
	l.scrollToKey = key
	return l
}

func (l Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	var override *theme.ListBoxOverride
	if l.hasOverride {
		override = &l.override
	}
	return l.layoutWithOverride(ctx, gtx, override, l.partsOnly)
}

func (l Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, instanceOverride *theme.ListBoxOverride, partsOnly bool) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	rootKey := core.FullKey(ctx, l.key)
	if l.derivedOwner != "" {
		rootKey = core.DerivedKey(ctx, l.derivedOwner, l.derivedRole)
	}
	state := l.stateFor(ctx)
	state.beginFrame()
	defer state.endFrame()

	// Bind disclosure and get current values
	if l.selectionMode == SelectionMultiple {
		state.bindMultiple(l)
		l.selectedKeys = state.currentMultipleValue(l)
	} else if l.selectionMode == SelectionSingle {
		state.bindSingle(l)
		l.selectedKey = state.currentSingleValue(l)
	}

	l.selectedKeySet = state.selectedKeys.Resolve(l.selectedKeys)
	l.disabledKeySet = state.disabledKeys.Resolve(l.disabledKeys)
	entries, items := state.resolveEntries(l)

	if !l.disabled {
		result := state.updateKeys(gtx, items, l, state.keyboardActiveKey(l))
		if result.focusKey != "" {
			core.RequestFocus(ctx, &state.item(result.focusKey).Clickable)
		}
		if result.actionKey != "" {
			l.activate(state, result.actionKey)
		}
	}
	if l.disabled {
		gtx = gtx.Disabled()
	}
	hovered, pressed, focused := false, false, false
	for itemKey, itemState := range state.items {
		if _, ok := state.itemKeys[itemKey]; !ok {
			continue
		}
		hovered = hovered || itemState.Clickable.Hovered()
		pressed = pressed || itemState.Clickable.Pressed()
		focused = focused || gtx.Focused(&itemState.Clickable)
	}
	selected := l.selectedKey != "" || len(l.selectedKeys) > 0
	request := theme.ListBoxRequest{}
	if instanceOverride != nil {
		request.Override = instanceOverride
	}
	visual := state.listBoxVisualForRequest(ctx, request)
	instanceStyle := visual.Override
	state.resolvedPartCache.prepare(ctx, core.ActiveTheme(ctx), instanceStyle)
	rootStyle := instanceStyle
	if partsOnly {
		rootStyle = style.Style{}
	}
	return layoutui.Styled(ctx, gtx, rootKey, style.State{
		Hovered:  hovered,
		Pressed:  pressed,
		Focused:  focused,
		Disabled: l.disabled || !gtx.Enabled(),
		Selected: selected,
	}, rootStyle, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return l.layout(ctx, gtx, rootKey, state, visual, entries, len(items) > 0, instanceStyle)
	}))
}

func (l Widget) activate(state *listBoxState, key string) {
	if l.onAction != nil {
		l.onAction(key)
	}
	switch l.selectionMode {
	case SelectionNone:
		return
	case SelectionMultiple:
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

func (l Widget) isSelected(key string) bool {
	switch l.selectionMode {
	case SelectionNone:
		return false
	case SelectionMultiple:
		return core.StringSetContains(l.selectedKeys, l.selectedKeySet, key)
	default:
		return key == l.selectedKey
	}
}

func (l Widget) allItems() []Item {
	_, items := l.entriesAndItems()
	return items
}

func (l Widget) entriesAndItems() ([]listBoxEntry, []Item) {
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
	items := make([]Item, 0, count)
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

func (l Widget) itemDisabled(item Item) bool {
	return l.disabled || item.Disabled || core.StringSetContains(l.disabledKeys, l.disabledKeySet, item.Key)
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

// Style merges instance declarations into the list.
func (l Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	l.override.Declaration = style.Join(l.override.Declaration, declaration)
	l.hasOverride = true
	return l
}

// StyleParts adapts a compound owner's declaration for embedded items.
func (l Widget) StyleParts(declaration style.Style) Widget {
	l.override.Declaration = declaration
	l.partsOnly = true
	l.hasOverride = !declaration.IsEmpty()
	return l
}
