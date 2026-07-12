package menu

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
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

type Item struct {
	Key         string
	Label       string
	Description string
	Shortcut    string
	Leading     frame.Widget
	Trailing    frame.Widget
	Disabled    bool
	Variant     ItemVariant
	Kind        ItemKind
	Checked     bool
	RadioGroup  string
	Value       string
	Children    []Item
	Sections    []Section
	KeepOpen    bool
}

type Section struct {
	Title string
	Items []Item
}

type Widget struct {
	key             string
	derivedOwner    string
	derivedRole     string
	items           []Item
	sections        []Section
	emptyText       string
	onAction        func(string)
	onCheckedChange func(string, bool)
	onRadioChange   func(string, string)
	onRequestClose  func()
	disabled        bool
	width           unit.Dp
	nested          bool
	parentState     *menuState
	parentItemKey   string
}

func Menu(key string, items []Item) Widget {
	return Widget{key: key, items: items, emptyText: "No actions"}
}

func MenuSections(key string, sections []Section) Widget {
	return Widget{key: key, sections: sections, emptyText: "No actions"}
}

func MenuSeparator() Item {
	return Item{Kind: ItemSeparator}
}

func MenuGroupLabel(label string) Item {
	return Item{Kind: ItemGroupLabel, Label: label}
}

func (m Widget) Sections(sections []Section) Widget {
	m.sections = sections
	return m
}

func (m Widget) EmptyText(text string) Widget {
	m.emptyText = text
	return m
}

func (m Widget) OnAction(fn func(string)) Widget {
	m.onAction = fn
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

func (m Widget) Disabled(disabled bool) Widget {
	m.disabled = disabled
	return m
}

func (m Widget) Width(dp int) Widget {
	m.width = unit.Dp(max(dp, 0))
	return m
}

func (m Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := m.stateFor(ctx)
	return m.layout(ctx, gtx, state, !m.disabled)
}

func (m Widget) withDerivedIdentity(owner, role string) Widget {
	m.derivedOwner = owner
	m.derivedRole = role
	return m
}

func (m Widget) withClose(fn func()) Widget {
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
		child = MenuSections(item.Key, item.Sections)
	} else {
		child = Menu(item.Key, item.Children)
	}
	child.onAction = m.onAction
	child.onCheckedChange = m.onCheckedChange
	child.onRadioChange = m.onRadioChange
	child.onRequestClose = m.onRequestClose
	child.disabled = m.disabled
	child.width = m.width
	return child.
		withDerivedIdentity(state.key, "submenu:"+item.Key).
		withParent(state, item.Key)
}

func (m Widget) closeToParent(ctx *frame.Context) {
	if m.parentState == nil {
		return
	}
	m.parentState.openSubmenu = ""
	if item := m.parentState.items[m.parentItemKey]; item != nil {
		frame.RequestFocus(ctx, &item.clickable)
	}
}

func (m Widget) entries() []entry {
	if len(m.sections) == 0 {
		entries := make([]entry, 0, len(m.items))
		for _, item := range m.items {
			entries = append(entries, entryForItem(item))
		}
		return entries
	}
	count := 0
	for _, section := range m.sections {
		count += len(section.Items) + 2
	}
	entries := make([]entry, 0, count)
	for index, section := range m.sections {
		if index > 0 {
			entries = append(entries, entry{kind: ItemSeparator})
		}
		if section.Title != "" {
			entries = append(entries, entry{kind: ItemGroupLabel, label: section.Title})
		}
		for _, item := range section.Items {
			entries = append(entries, entryForItem(item))
		}
	}
	return entries
}

func entryForItem(item Item) entry {
	return entry{kind: item.Kind, label: item.Label, item: item}
}

type entry struct {
	kind  ItemKind
	label string
	item  Item
}

func (m Widget) actionableItems() []Item {
	entries := m.entries()
	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		if entry.kind != ItemSeparator && entry.kind != ItemGroupLabel {
			items = append(items, entry.item)
		}
	}
	return items
}

func (m Widget) activate(item Item) {
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
	case ItemSubmenu:
		return
	default:
		if m.onAction != nil {
			m.onAction(item.Key)
		}
	}
	if !item.KeepOpen && m.onRequestClose != nil {
		m.onRequestClose()
	}
}

func (m Widget) focusFirst(ctx *frame.Context, state *menuState, visible bool) bool {
	items := m.actionableItems()
	index := menuFirstEnabled(items)
	if index < 0 {
		return false
	}
	item := state.item(items[index].Key)
	item.focus.Prepare(visible)
	frame.RequestFocus(ctx, &item.clickable)
	return true
}
