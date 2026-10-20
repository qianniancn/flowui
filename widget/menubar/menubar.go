package menubar

import (
	"github.com/qianniancn/flowui/style"
	"unicode"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/navigation"
)

type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

// Item binds a top-level trigger to one Menu.
type Item struct {
	key       string
	label     string
	trigger   core.Widget
	menu      menu.Widget
	disabled  bool
	accessKey rune // lowercase letter for Alt+key activation; 0 disables
}

// NewMenu creates a top-level item backed by a flat Menu.
func NewMenu(key, label string, items []menu.Item) Item {
	return Item{
		key:   key,
		label: label,
		menu:  menu.New(key+":menu", items),
	}
}

// NewMenuSections creates a top-level item backed by a sectioned Menu.
func NewMenuSections(key, label string, sections []menu.Section) Item {
	return Item{
		key:   key,
		label: label,
		menu:  menu.Sections(key+":menu", sections),
	}
}

// NewMenuContent creates a top-level item from an already configured Menu.
func NewMenuContent(key, label string, content menu.Widget) Item {
	return Item{key: key, label: label, menu: content}
}

// Trigger replaces the default text trigger. Label remains its accessible name.
func (i Item) Trigger(trigger core.Widget) Item {
	i.trigger = trigger
	return i
}

func (i Item) Disabled(disabled bool) Item {
	i.disabled = disabled
	i.menu = i.menu.Disabled(disabled)
	return i
}

// AccessKey sets the Alt+letter accelerator that opens this top-level menu.
// Pass 0 to clear. Matching is case-insensitive.
func (i Item) AccessKey(key rune) Item {
	if key == 0 {
		i.accessKey = 0
		return i
	}
	i.accessKey = unicode.ToLower(key)
	return i
}

// OnActionEvent reports the complete activated item and its submenu path.
func (i Item) OnActionEvent(fn func(menu.ActionEvent)) Item {
	i.menu = i.menu.OnActionEvent(fn)
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
	override          *theme.MenubarOverride
	key               string
	items             []Item
	orientation       Orientation
	loopFocus         bool
	modal             bool
	compact           bool
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

// Compact uses desktop title-bar density for triggers and popup menus.
func (m Widget) Compact(compact bool) Widget {
	m.compact = compact
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

func (m Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return m.layoutWithOverride(ctx, gtx, m.override)
}

func (m Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.MenubarOverride) layout.Dimensions {
	return m.layout(ctx, gtx, override)
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

func (m Widget) navList() navigation.List {
	return navigation.List{
		Count:    len(m.items),
		Disabled: func(i int) bool { return m.itemDisabled(m.items[i]) },
		Label:    func(i int) string { return m.items[i].label },
	}
}

func (m Widget) firstEnabled() int {
	index, _ := navigation.First(m.navList())
	return index
}

func (m Widget) lastEnabled() int {
	index, _ := navigation.Last(m.navList())
	return index
}

func (m Widget) moveIndex(current, delta int) int {
	if len(m.items) == 0 || delta == 0 {
		return current
	}
	// Menubar wrap is configurable via LoopFocus.
	next, ok := navigation.Move(m.navList(), current, delta, m.loopFocus)
	if !ok && current >= 0 {
		return current
	}
	return next
}

func (m Widget) typeaheadIndex(current int, query string) int {
	if index, ok := navigation.Match(m.navList(), current, query); ok {
		return index
	}
	return -1
}

// Style merges instance declarations into the widget.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.MenubarOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
