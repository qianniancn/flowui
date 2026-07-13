package menubar

import (
	"strings"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/menu"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

// Item binds a top-level trigger to one Menu.
type Item struct {
	key      string
	label    string
	trigger  frame.Widget
	menu     menu.Widget
	disabled bool
}

// NewMenu creates a top-level item backed by a flat Menu.
func NewMenu(key, label string, items []menu.Item) Item {
	return Item{
		key:   key,
		label: label,
		menu:  menu.Menu(key+":menu", items),
	}
}

// NewMenuSections creates a top-level item backed by a sectioned Menu.
func NewMenuSections(key, label string, sections []menu.Section) Item {
	return Item{
		key:   key,
		label: label,
		menu:  menu.MenuSections(key+":menu", sections),
	}
}

// NewMenuContent creates a top-level item from an already configured Menu.
func NewMenuContent(key, label string, content menu.Widget) Item {
	return Item{key: key, label: label, menu: content}
}

// Trigger replaces the default text trigger. Label remains its accessible name.
func (i Item) Trigger(trigger frame.Widget) Item {
	i.trigger = trigger
	return i
}

func (i Item) Disabled(disabled bool) Item {
	i.disabled = disabled
	i.menu = i.menu.Disabled(disabled)
	return i
}

func (i Item) OnAction(fn func(string)) Item {
	i.menu = i.menu.OnAction(fn)
	return i
}

func (i Item) OnChange(fn func(string)) Item {
	i.menu = i.menu.OnChange(fn)
	return i
}

func (i Item) OnSelectionChange(fn func([]string)) Item {
	i.menu = i.menu.OnSelectionChange(fn)
	return i
}

func (i Item) OnCheckedChange(fn func(string, bool)) Item {
	i.menu = i.menu.OnCheckedChange(fn)
	return i
}

func (i Item) OnRadioChange(fn func(string, string)) Item {
	i.menu = i.menu.OnRadioChange(fn)
	return i
}

func (i Item) CloseOnSelect(close bool) Item {
	i.menu = i.menu.CloseOnSelect(close)
	return i
}

func (i Item) Width(dp int) Item {
	i.menu = i.menu.Width(dp)
	return i
}

type Widget struct {
	key               string
	items             []Item
	orientation       Orientation
	loopFocus         bool
	modal             bool
	disabled          bool
	alt               string
	openKey           string
	hasOpenKey        bool
	defaultOpenKey    string
	hasDefaultOpenKey bool
	onOpenChange      func(string)
}

func New(key string, items []Item) Widget {
	return Widget{key: key, items: items, loopFocus: true, modal: true}
}

func (m Widget) Orientation(orientation Orientation) Widget {
	m.orientation = orientation
	return m
}

func (m Widget) LoopFocus(loop bool) Widget {
	m.loopFocus = loop
	return m
}

func (m Widget) Modal(modal bool) Widget {
	m.modal = modal
	return m
}

func (m Widget) Disabled(disabled bool) Widget {
	m.disabled = disabled
	return m
}

func (m Widget) Alt(alt string) Widget {
	m.alt = alt
	return m
}

// OpenKey makes the currently open menu controlled. An empty key closes it.
func (m Widget) OpenKey(key string) Widget {
	m.openKey = key
	m.hasOpenKey = true
	return m
}

func (m Widget) DefaultOpenKey(key string) Widget {
	m.defaultOpenKey = key
	m.hasDefaultOpenKey = true
	return m
}

func (m Widget) OnOpenChange(fn func(string)) Widget {
	m.onOpenChange = fn
	return m
}

func (m Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return m.layout(ctx, gtx)
}

func (m Widget) itemDisabled(item Item) bool {
	return m.disabled || item.disabled
}

func (m Widget) itemEnabled(key string) bool {
	if key == "" {
		return false
	}
	for _, item := range m.items {
		if item.key == key {
			return !m.itemDisabled(item)
		}
	}
	return false
}

func (m Widget) itemByKey(key string) (Item, bool) {
	for _, item := range m.items {
		if item.key == key {
			return item, true
		}
	}
	return Item{}, false
}

func (m Widget) indexOf(key string) int {
	for index, item := range m.items {
		if item.key == key {
			return index
		}
	}
	return -1
}

func (m Widget) firstEnabled() int {
	for index, item := range m.items {
		if !m.itemDisabled(item) {
			return index
		}
	}
	return -1
}

func (m Widget) lastEnabled() int {
	for index := len(m.items) - 1; index >= 0; index-- {
		if !m.itemDisabled(m.items[index]) {
			return index
		}
	}
	return -1
}

func (m Widget) moveIndex(current, delta int) int {
	if len(m.items) == 0 || delta == 0 {
		return current
	}
	if current < 0 {
		if delta > 0 {
			return m.firstEnabled()
		}
		return m.lastEnabled()
	}
	for step := 1; step <= len(m.items); step++ {
		index := current + delta*step
		if m.loopFocus {
			index %= len(m.items)
			if index < 0 {
				index += len(m.items)
			}
		} else if index < 0 || index >= len(m.items) {
			return current
		}
		if !m.itemDisabled(m.items[index]) {
			return index
		}
	}
	return current
}

func (m Widget) typeaheadIndex(current int, query string) int {
	if len(m.items) == 0 || query == "" {
		return -1
	}
	query = strings.ToLower(query)
	for step := 1; step <= len(m.items); step++ {
		index := (current + step + len(m.items)) % len(m.items)
		if m.itemDisabled(m.items[index]) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(m.items[index].label), query) {
			return index
		}
	}
	return -1
}
