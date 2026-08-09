package sidebar

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

// Item describes one destination in a Sidebar.
type Item struct {
	Key              string
	Label            string
	Leading          frame.Widget
	Trailing         frame.Widget
	Switcher         frame.Widget
	ExpandedSwitcher frame.Widget
	Children         []Item
	Disabled         bool
}

// Section groups related Sidebar destinations.
type Section struct {
	Title string
	Items []Item
}

// ExpandAction controls how a nested navigation group opens while the sidebar
// is expanded. Collapsed groups always open their flyout on hover.
type ExpandAction uint8

const (
	ExpandActionClick ExpandAction = iota
	ExpandActionHover
)

// Widget presents controlled primary application navigation.
type Widget struct {
	key             string
	selectedKey     string
	items           []Item
	sections        []Section
	dataVersion     uint64
	hasDataVersion  bool
	openKeys        []string
	expandAction    ExpandAction
	header          frame.Widget
	footer          frame.Widget
	emptyText       string
	alt             string
	onChange        func(string)
	onAction        func(string)
	onOpenChange    func([]string)
	disabledKeys    []string
	disabledKeySet  stateutil.StringSet
	disabled        bool
	collapsed       bool
	width           unit.Dp
	collapsedWidth  unit.Dp
	padding         unit.Dp
	hasPadding      bool
	itemGap         unit.Dp
	hasItemGap      bool
	itemHeight      unit.Dp
	itemPaddingX    unit.Dp
	hasItemPadding  bool
	itemRadius      unit.Dp
	hasItemRadius   bool
	inlineIndent    unit.Dp
	hasInlineIndent bool
	customStyle     flowstyle.Style
}

// New creates a controlled Sidebar.
func New(key, selectedKey string, items []Item) Widget {
	return Widget{key: key, selectedKey: selectedKey, items: append([]Item(nil), items...), emptyText: "No destinations"}
}

// NewSections creates a controlled Sidebar with grouped destinations.
func NewSections(key, selectedKey string, sections []Section) Widget {
	return Widget{key: key, selectedKey: selectedKey, sections: cloneSections(sections), emptyText: "No destinations"}
}

func (w Widget) Header(header frame.Widget) Widget {
	w.header = header
	return w
}

func (w Widget) Footer(footer frame.Widget) Widget {
	w.footer = footer
	return w
}

// DataVersion enables item validation and flattened-data reuse. Increase
// version whenever the item data or section structure changes.
func (w Widget) DataVersion(version uint64) Widget {
	w.dataVersion = version
	w.hasDataVersion = true
	return w
}

// OpenKeys sets the controlled set of expanded navigation groups.
func (w Widget) OpenKeys(keys []string) Widget {
	w.openKeys = append([]string(nil), keys...)
	return w
}

// OnOpenChange reports a requested change to the expanded navigation groups.
func (w Widget) OnOpenChange(fn func([]string)) Widget {
	w.onOpenChange = fn
	return w
}

// ExpandAction controls whether nested groups open on click or hover while the
// sidebar is expanded. Collapsed groups always open their flyout on hover.
func (w Widget) ExpandAction(action ExpandAction) Widget {
	if action != ExpandActionClick && action != ExpandActionHover {
		panic("flowui: invalid sidebar expand action")
	}
	w.expandAction = action
	return w
}

func (w Widget) Collapsed(collapsed bool) Widget {
	w.collapsed = collapsed
	return w
}

func (w Widget) Width(dp int) Widget {
	if dp <= 0 {
		panic("flowui: sidebar width must be positive")
	}
	w.width = unit.Dp(dp)
	return w
}

func (w Widget) CollapsedWidth(dp int) Widget {
	if dp <= 0 {
		panic("flowui: sidebar collapsed width must be positive")
	}
	w.collapsedWidth = unit.Dp(dp)
	return w
}

// Padding overrides the outer inset around sidebar content.
func (w Widget) Padding(dp int) Widget {
	if dp < 0 {
		panic("flowui: sidebar padding cannot be negative")
	}
	w.padding = unit.Dp(dp)
	w.hasPadding = true
	return w
}

// ItemGap overrides the vertical gap between navigation items.
func (w Widget) ItemGap(dp int) Widget {
	if dp < 0 {
		panic("flowui: sidebar item gap cannot be negative")
	}
	w.itemGap = unit.Dp(dp)
	w.hasItemGap = true
	return w
}

func (w Widget) ItemHeight(dp int) Widget {
	if dp <= 0 {
		panic("flowui: sidebar item height must be positive")
	}
	w.itemHeight = unit.Dp(dp)
	return w
}

// ItemPaddingX overrides horizontal padding inside each navigation item.
func (w Widget) ItemPaddingX(dp int) Widget {
	if dp < 0 {
		panic("flowui: sidebar item padding cannot be negative")
	}
	w.itemPaddingX = unit.Dp(dp)
	w.hasItemPadding = true
	return w
}

// ItemRadius overrides the corner radius of item hover and selected fills.
func (w Widget) ItemRadius(dp int) Widget {
	if dp < 0 {
		panic("flowui: sidebar item radius cannot be negative")
	}
	w.itemRadius = unit.Dp(dp)
	w.hasItemRadius = true
	return w
}

// InlineIndent overrides the horizontal indentation added for each nesting level.
func (w Widget) InlineIndent(dp int) Widget {
	if dp < 0 {
		panic("flowui: sidebar inline indent cannot be negative")
	}
	w.inlineIndent = unit.Dp(dp)
	w.hasInlineIndent = true
	return w
}

func (w Widget) Alt(description string) Widget {
	w.alt = description
	return w
}

func (w Widget) EmptyText(text string) Widget {
	w.emptyText = text
	return w
}

func (w Widget) DisabledKeys(keys []string) Widget {
	w.disabledKeys = append([]string(nil), keys...)
	return w
}

func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}

func (w Widget) OnChange(fn func(string)) Widget {
	w.onChange = fn
	return w
}

func (w Widget) OnAction(fn func(string)) Widget {
	w.onAction = fn
	return w
}

func (w Widget) Style(value flowstyle.Style) Widget {
	w.customStyle = value
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	rootKey := frame.FullKey(ctx, w.key)
	state := sidebarStateFor(ctx, w.key)
	w.disabledKeySet = state.disabledKeys.Resolve(w.disabledKeys)
	entries, items := state.resolveEntries(w)
	state.beginFrame()
	defer state.endFrame()
	if !w.disabled {
		result := state.updateKeys(gtx, w, items)
		if result.focusKey != "" {
			itemState := state.item(result.focusKey)
			frame.RequestFocus(ctx, &itemState.clickable)
			state.ensureVisible(sidebarEntryIndex(entries, result.focusKey))
		}
		if result.actionKey != "" {
			w.activateEntry(entries, result.actionKey)
		}
	}

	restoreKey := frame.PushKey(ctx, w.key)
	defer restoreKey()
	hovered, pressed, focused, focusVisible := false, false, false, false
	for _, item := range items {
		itemState := state.item(item.Key)
		hovered = hovered || itemState.clickable.Hovered()
		pressed = pressed || itemState.clickable.Pressed()
		itemFocused := gtx.Focused(&itemState.clickable)
		focused = focused || itemFocused
		focusVisible = focusVisible || frame.FocusVisible(ctx, &itemState.clickable, itemFocused)
	}
	return layoutui.LayoutStyled(ctx, gtx, rootKey, flowstyle.StyleState{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     w.disabled || !gtx.Enabled(),
		Selected:     w.selectedKey != "",
		Expanded:     !w.collapsed,
	}, w.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return w.layout(ctx, gtx, state, entries)
	}))
}

func (w Widget) activate(key string) {
	if w.onAction != nil {
		w.onAction(key)
	}
	if key != w.selectedKey && w.onChange != nil {
		w.onChange(key)
	}
}

func (w Widget) activateEntry(entries []entry, key string) {
	for _, entry := range entries {
		if entry.section || entry.item.Key != key {
			continue
		}
		if len(entry.item.Children) > 0 && !w.collapsed {
			w.requestOpen(key, !entry.expanded)
			return
		}
		w.activate(key)
		return
	}
}

func (w Widget) requestOpen(key string, open bool) {
	next := toggleOpenKeys(w.openKeys, key, open)
	if w.onOpenChange != nil {
		w.onOpenChange(next)
	}
}

func (w Widget) itemDisabled(item Item) bool {
	return w.disabled || item.Disabled || stateutil.StringSetContains(w.disabledKeys, w.disabledKeySet, item.Key)
}

type entry struct {
	section   bool
	title     string
	item      Item
	depth     int
	parentKey string
	expanded  bool
}

func (w Widget) entriesAndItems() ([]entry, []Item) {
	open := make(map[string]struct{}, len(w.openKeys))
	for _, key := range w.openKeys {
		open[key] = struct{}{}
	}
	return w.visibleEntries(open)
}

func (w Widget) visibleEntries(open map[string]struct{}) ([]entry, []Item) {
	entries := make([]entry, 0)
	items := make([]Item, 0)
	var appendItems func(items []Item, depth int, parentKey string)
	appendItems = func(children []Item, depth int, parentKey string) {
		for _, item := range children {
			_, expanded := open[item.Key]
			entries = append(entries, entry{
				item:      item,
				depth:     depth,
				parentKey: parentKey,
				expanded:  expanded,
			})
			items = append(items, item)
			if !w.collapsed && expanded && len(item.Children) > 0 {
				appendItems(item.Children, depth+1, item.Key)
			}
		}
	}

	if len(w.sections) == 0 {
		appendItems(w.items, 0, "")
		return entries, items
	}
	for _, section := range w.sections {
		if len(section.Items) == 0 {
			continue
		}
		if w.collapsed {
			if len(entries) > 0 {
				entries = append(entries, entry{section: true})
			}
		} else if section.Title != "" {
			entries = append(entries, entry{section: true, title: section.Title})
		}
		appendItems(section.Items, 0, "")
	}
	return entries, items
}

func cloneSections(sections []Section) []Section {
	result := make([]Section, len(sections))
	for index, section := range sections {
		result[index] = section
		result[index].Items = cloneItems(section.Items)
	}
	return result
}

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	result := make([]Item, len(items))
	copy(result, items)
	for index := range result {
		result[index].Children = cloneItems(items[index].Children)
	}
	return result
}

func toggleOpenKeys(keys []string, key string, open bool) []string {
	next := make([]string, 0, len(keys)+1)
	found := false
	for _, value := range keys {
		if value == key {
			found = true
			if open {
				next = append(next, value)
			}
			continue
		}
		next = append(next, value)
	}
	if open && !found {
		next = append(next, key)
	}
	return next
}

func sidebarEntryIndex(entries []entry, key string) int {
	for index, entry := range entries {
		if !entry.section && entry.item.Key == key {
			return index
		}
	}
	return -1
}

func validateSidebarItems(items []Item) {
	seen := make(map[string]struct{}, len(items))
	var validate func([]Item)
	validate = func(items []Item) {
		for _, item := range items {
			if item.Key == "" {
				panic("flowui: empty sidebar item key")
			}
			if _, ok := seen[item.Key]; ok {
				panic(fmt.Sprintf("flowui: duplicate sidebar item key %q", item.Key))
			}
			seen[item.Key] = struct{}{}
			validate(item.Children)
		}
	}
	validate(items)
}

func validateSidebarWidget(w Widget) {
	if len(w.sections) == 0 {
		validateSidebarItems(w.items)
		return
	}
	items := make([]Item, 0)
	for _, section := range w.sections {
		items = append(items, section.Items...)
	}
	validateSidebarItems(items)
}
