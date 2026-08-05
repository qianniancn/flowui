package menu

import (
	"slices"

	"gioui.org/layout"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type ItemKind uint8

const (
	ItemAction ItemKind = iota
	ItemCheckbox
	ItemRadio
	ItemSubmenu
	ItemSeparator
	ItemGroupLabel
)

type ItemVariant uint8

const (
	ItemDefault ItemVariant = iota
	ItemDanger
)

type SelectionMode uint8

const (
	SelectionNone SelectionMode = iota
	SelectionSingle
	SelectionMultiple
)

type IndicatorType uint8

const (
	IndicatorNone IndicatorType = iota
	IndicatorCheckmark
	IndicatorDot
)

// ActionEvent describes an activated menu item and its path from the root
// menu. Path includes the activated item's key.
type ActionEvent struct {
	Item Item
	Key  string
	Path []string
}

type Item struct {
	Key              string
	Label            string
	Description      string
	Shortcut         string
	Leading          frame.Widget
	Trailing         frame.Widget
	Indicator        func(bool) frame.Widget
	SubmenuIndicator frame.Widget
	Disabled         bool
	Variant          ItemVariant
	Kind             ItemKind
	IndicatorType    IndicatorType
	Checked          bool
	RadioGroup       string
	Value            string
	Children         []Item
	Sections         []Section
	KeepOpen         bool
	OnAction         func()
}

type Section struct {
	Title             string
	Items             []Item
	SeparatorBefore   bool
	SelectionMode     SelectionMode
	SelectedKeys      []string
	OnSelectionChange func([]string)
}

type Widget struct {
	key                  string
	derivedOwner         string
	derivedRole          string
	items                []Item
	sections             []Section
	dataVersion          uint64
	hasDataVersion       bool
	autoSeparateSections bool
	beforeContent        frame.Widget
	afterContent         frame.Widget
	emptyText            string
	selectedKey          string
	selectedKeys         []string
	disabledKeys         []string
	selectionMode        SelectionMode
	onAction             func(string)
	onActionEvent        func(ActionEvent)
	onChange             func(string)
	onSelectionChange    func([]string)
	onCheckedChange      func(string, bool)
	onRadioChange        func(string, string)
	onRequestClose       func(bool)
	onRootPrevious       func()
	onRootNext           func()
	closeOnSelect        bool
	hasCloseOnSelect     bool
	disabled             bool
	compact              bool
	width                unit.Dp
	minWidthPx           int
	autoWidth            bool
	minWidth             unit.Dp
	hasMinWidth          bool
	maxWidth             unit.Dp
	hasMaxWidth          bool
	nested               bool
	actionPath           []string
	parentState          *menuState
	parentItemKey        string
	customStyle          flowstyle.Style
}

func Menu(key string, items []Item) Widget {
	return Widget{key: key, items: items, emptyText: "No actions"}
}

func MenuSections(key string, sections []Section) Widget {
	return Widget{key: key, sections: sections, autoSeparateSections: true, emptyText: "No actions"}
}

func MenuSeparator() Item {
	return Item{Kind: ItemSeparator}
}

func MenuGroupLabel(label string) Item {
	return Item{Kind: ItemGroupLabel, Label: label}
}

func (m Widget) Sections(sections []Section) Widget {
	m.sections = sections
	m.autoSeparateSections = true
	return m
}

func (m Widget) AutoSeparateSections(enabled bool) Widget {
	m.autoSeparateSections = enabled
	return m
}

func (m Widget) BeforeContent(content frame.Widget) Widget {
	m.beforeContent = content
	return m
}

func (m Widget) AfterContent(content frame.Widget) Widget {
	m.afterContent = content
	return m
}

func (m Widget) EmptyText(text string) Widget {
	m.emptyText = text
	return m
}

// DataVersion enables item validation and reuse of flattened data and
// AutoWidth measurements. Increase the version whenever menu content,
// section structure, or width-affecting child content changes.
func (m Widget) DataVersion(version uint64) Widget {
	m.dataVersion = version
	m.hasDataVersion = true
	return m
}

func (m Widget) SelectionMode(mode SelectionMode) Widget {
	m.selectionMode = mode
	return m
}

func (m Widget) SelectedKey(key string) Widget {
	m.selectedKey = key
	return m
}

func (m Widget) SelectedKeys(keys []string) Widget {
	m.selectedKeys = keys
	return m
}

func (m Widget) DisabledKeys(keys []string) Widget {
	m.disabledKeys = keys
	return m
}

func (m Widget) OnAction(fn func(string)) Widget {
	m.onAction = fn
	return m
}

// OnActionEvent reports the complete activated item and its submenu path.
func (m Widget) OnActionEvent(fn func(ActionEvent)) Widget {
	m.onActionEvent = fn
	return m
}

func (m Widget) OnChange(fn func(string)) Widget {
	m.onChange = fn
	return m
}

func (m Widget) OnSelectionChange(fn func([]string)) Widget {
	m.onSelectionChange = fn
	return m
}

func (m Widget) OnCheckedChange(fn func(string, bool)) Widget {
	m.onCheckedChange = fn
	return m
}

func (m Widget) OnRadioChange(fn func(string, string)) Widget {
	m.onRadioChange = fn
	return m
}

func (m Widget) CloseOnSelect(close bool) Widget {
	m.closeOnSelect = close
	m.hasCloseOnSelect = true
	return m
}

func (m Widget) Disabled(disabled bool) Widget {
	m.disabled = disabled
	return m
}

// Compact uses desktop menu density without changing the default HeroUI menu.
func (m Widget) Compact(compact bool) Widget {
	m.compact = compact
	return m
}

func (m Widget) Width(dp int) Widget {
	m.width = unit.Dp(max(dp, 0))
	return m
}

// WithMinimumWidthPx applies an owner-provided physical minimum width without
// adding that implementation detail to the public MenuWidget API.
func WithMinimumWidthPx(m Widget, px int) Widget {
	m.minWidthPx = max(px, 0)
	return m
}

// AutoWidth sizes the menu to its content, subject to MinWidth and MaxWidth.
func (m Widget) AutoWidth() Widget {
	m.autoWidth = true
	return m
}

// MinWidth sets the minimum menu width when AutoWidth is enabled.
func (m Widget) MinWidth(dp int) Widget {
	m.minWidth = unit.Dp(max(dp, 0))
	m.hasMinWidth = true
	return m
}

// MaxWidth limits the menu width when AutoWidth is enabled.
func (m Widget) MaxWidth(dp int) Widget {
	m.maxWidth = unit.Dp(max(dp, 0))
	m.hasMaxWidth = true
	return m
}

func (m Widget) Style(value flowstyle.Style) Widget {
	m.customStyle = value
	return m
}

func (m Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := m.stateFor(ctx)
	return m.layoutRoot(ctx, gtx, state, !m.disabled)
}

func (m Widget) layoutRoot(ctx *frame.Context, gtx layout.Context, state *menuState, interactive bool) layout.Dimensions {
	hovered, pressed, focused := false, false, false
	for _, item := range state.items {
		hovered = hovered || item.clickable.Hovered()
		pressed = pressed || item.clickable.Pressed()
		focused = focused || gtx.Focused(&item.clickable)
	}
	root := flowstyle.Join(menuRootDeclaration(frame.ActiveTheme(ctx), m.themeTokens(ctx)), m.customStyle)
	return layoutui.LayoutStyled(ctx, gtx, state.key, flowstyle.StyleState{
		Hovered:  hovered,
		Pressed:  pressed,
		Focused:  focused,
		Disabled: m.disabled || !gtx.Enabled(),
		Selected: m.selectedKey != "" || len(m.selectedKeys) > 0,
		Open:     state.openSubmenu != "",
	}, root, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return m.layout(ctx, gtx, state, interactive)
	}))
}

func (m Widget) withDerivedIdentity(owner, role string) Widget {
	m.derivedOwner = owner
	m.derivedRole = role
	return m
}

func (m Widget) withClose(fn func(bool)) Widget {
	m.onRequestClose = fn
	return m
}

func (m Widget) withParent(state *menuState, itemKey string) Widget {
	m.nested = true
	m.parentState = state
	m.parentItemKey = itemKey
	return m
}

func (m Widget) submenu(state *menuState, item Item) Widget {
	var child Widget
	if len(item.Sections) > 0 {
		child = MenuSections(item.Key, item.Sections).AutoSeparateSections(false)
	} else {
		child = Menu(item.Key, item.Children)
	}
	child.beforeContent = nil
	child.afterContent = nil
	child.selectedKey = m.selectedKey
	child.selectedKeys = m.selectedKeys
	child.disabledKeys = m.disabledKeys
	child.selectionMode = m.selectionMode
	child.onAction = m.onAction
	child.onActionEvent = m.onActionEvent
	child.onChange = m.onChange
	child.onSelectionChange = m.onSelectionChange
	child.onCheckedChange = m.onCheckedChange
	child.onRadioChange = m.onRadioChange
	child.onRequestClose = m.onRequestClose
	child.onRootPrevious = m.onRootPrevious
	child.onRootNext = m.onRootNext
	child.closeOnSelect = m.closeOnSelect
	child.hasCloseOnSelect = m.hasCloseOnSelect
	child.disabled = m.disabled
	child.compact = m.compact
	child.dataVersion = m.dataVersion
	child.hasDataVersion = m.hasDataVersion
	child.autoSeparateSections = m.autoSeparateSections
	child.width = m.width
	child.minWidthPx = m.minWidthPx
	child.autoWidth = m.autoWidth
	child.minWidth = m.minWidth
	child.hasMinWidth = m.hasMinWidth
	child.maxWidth = m.maxWidth
	child.hasMaxWidth = m.hasMaxWidth
	child.customStyle = m.customStyle
	child.actionPath = append(append([]string(nil), m.actionPath...), item.Key)
	return child.
		withDerivedIdentity(state.key, "submenu:"+item.Key).
		withParent(state, item.Key)
}

func (m Widget) themeTokens(ctx *frame.Context) theme.MenuTheme {
	tokens := frame.ActiveTheme(ctx).Components.Menu
	if !m.compact {
		return tokens
	}
	tokens.Padding = min(tokens.Padding, 4)
	tokens.Radius = min(tokens.Radius, 8)
	tokens.ItemGap = min(tokens.ItemGap, 1)
	tokens.ItemMinHeight = min(tokens.ItemMinHeight, 28)
	tokens.ItemRadius = min(tokens.ItemRadius, 4)
	tokens.ItemPaddingX = min(tokens.ItemPaddingX, 8)
	tokens.ItemPaddingY = min(tokens.ItemPaddingY, 3)
	tokens.ItemContentGap = min(tokens.ItemContentGap, 8)
	tokens.ItemTextSize = min(tokens.ItemTextSize, 13)
	tokens.ItemDescriptionSize = min(tokens.ItemDescriptionSize, 11)
	tokens.ShortcutTextSize = min(tokens.ShortcutTextSize, 12)
	tokens.ShortcutHeight = min(tokens.ShortcutHeight, 20)
	tokens.ShortcutPaddingX = min(tokens.ShortcutPaddingX, 4)
	tokens.IndicatorSize = min(tokens.IndicatorSize, 14)
	tokens.CheckmarkSize = min(tokens.CheckmarkSize, 9)
	tokens.RadioDotSize = min(tokens.RadioDotSize, 6)
	tokens.SubmenuIndicatorSize = min(tokens.SubmenuIndicatorSize, 12)
	tokens.SectionTextSize = min(tokens.SectionTextSize, 11)
	tokens.SectionPaddingX = min(tokens.SectionPaddingX, 6)
	tokens.SectionPaddingTop = min(tokens.SectionPaddingTop, 4)
	tokens.SectionPaddingBottom = min(tokens.SectionPaddingBottom, 3)
	return tokens
}

func (m Widget) closeToParent(ctx *frame.Context) {
	m.dismissToParent()
	if m.parentState == nil {
		return
	}
	if item := m.parentState.items[m.parentItemKey]; item != nil {
		frame.RequestFocusVisible(ctx, &item.clickable, true)
	}
}

func (m Widget) dismissToParent() {
	if m.parentState == nil {
		return
	}
	m.parentState.openSubmenu = ""
}

type entry struct {
	item         Item
	sectionIndex int
	separator    bool
	sectionTitle string
}

func (m Widget) entries() []entry {
	if len(m.sections) == 0 {
		entries := make([]entry, 0, len(m.items))
		for _, item := range m.items {
			entries = append(entries, entryForItem(item, -1))
		}
		return entries
	}
	count := 0
	for _, section := range m.sections {
		count += len(section.Items) + 2
	}
	entries := make([]entry, 0, count)
	for index, section := range m.sections {
		if index > 0 && (m.autoSeparateSections || section.SeparatorBefore) {
			entries = append(entries, entry{separator: true, sectionIndex: -1})
		}
		if section.Title != "" {
			entries = append(entries, entry{sectionIndex: index, sectionTitle: section.Title})
		}
		for _, item := range section.Items {
			entries = append(entries, entryForItem(item, index))
		}
	}
	return entries
}

func entryForItem(item Item, sectionIndex int) entry {
	sectionTitle := ""
	if item.Kind == ItemGroupLabel {
		sectionTitle = item.Label
	}
	return entry{
		item:         item,
		sectionIndex: sectionIndex,
		separator:    item.Kind == ItemSeparator,
		sectionTitle: sectionTitle,
	}
}

func (m Widget) actionableEntries() []entry {
	entries := m.entries()
	result := make([]entry, 0, len(entries))
	for _, entry := range entries {
		if !entry.separator && entry.sectionTitle == "" {
			result = append(result, entry)
		}
	}
	return result
}

func (m Widget) itemDisabled(item Item) bool {
	return m.disabled || item.Disabled || slices.Contains(m.disabledKeys, item.Key)
}

func (m Widget) selection(entry entry) (SelectionMode, []string, func([]string)) {
	if entry.item.Kind == ItemCheckbox {
		keys := []string(nil)
		if entry.item.Checked {
			keys = []string{entry.item.Key}
		}
		return SelectionMultiple, keys, nil
	}
	if entry.item.Kind == ItemRadio {
		keys := []string(nil)
		if entry.item.Checked {
			keys = []string{entry.item.Key}
		}
		return SelectionSingle, keys, nil
	}
	if entry.sectionIndex >= 0 && entry.sectionIndex < len(m.sections) {
		section := m.sections[entry.sectionIndex]
		if section.SelectionMode != SelectionNone {
			return section.SelectionMode, section.SelectedKeys, section.OnSelectionChange
		}
	}
	keys := m.selectedKeys
	if m.selectedKey != "" && len(keys) == 0 {
		keys = []string{m.selectedKey}
	}
	return m.selectionMode, keys, m.onSelectionChange
}

func (m Widget) selected(entry entry) bool {
	mode, keys, _ := m.selection(entry)
	return mode != SelectionNone && slices.Contains(keys, entry.item.Key)
}

func (m Widget) activate(entry entry) bool {
	item := entry.item
	if item.OnAction != nil {
		item.OnAction()
	}
	if m.onActionEvent != nil {
		path := append(append([]string(nil), m.actionPath...), item.Key)
		m.onActionEvent(ActionEvent{Item: item, Key: item.Key, Path: path})
	}
	switch item.Kind {
	case ItemCheckbox:
		if m.onCheckedChange != nil {
			m.onCheckedChange(item.Key, !item.Checked)
		}
	case ItemRadio:
		if m.onRadioChange != nil {
			value := item.Value
			if value == "" {
				value = item.Key
			}
			m.onRadioChange(item.RadioGroup, value)
		}
	default:
		if m.onAction != nil {
			m.onAction(item.Key)
		}
		mode, selected, onSelectionChange := m.selection(entry)
		if mode != SelectionNone {
			next := append([]string(nil), selected...)
			if mode == SelectionSingle {
				next = []string{item.Key}
				if !m.sectionOverridesSelection(entry.sectionIndex) && m.onChange != nil {
					m.onChange(item.Key)
				}
			} else if index := slices.Index(next, item.Key); index >= 0 {
				next = append(next[:index], next[index+1:]...)
			} else {
				next = append(next, item.Key)
			}
			if onSelectionChange != nil {
				onSelectionChange(next)
			}
		}
	}
	close := true
	if m.hasCloseOnSelect {
		close = m.closeOnSelect
	}
	return close && !item.KeepOpen
}

func (m Widget) sectionOverridesSelection(index int) bool {
	return index >= 0 && index < len(m.sections) && m.sections[index].SelectionMode != SelectionNone
}

func itemHasSubmenu(item Item) bool {
	return item.Kind == ItemSubmenu || len(item.Children) > 0 || len(item.Sections) > 0
}
