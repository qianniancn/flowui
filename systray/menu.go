package systray

import (
	"sync"
)

// MenuItemType represents the type of menu item
type MenuItemType int

const (
	MenuItemText MenuItemType = iota
	MenuItemSeparator
	MenuItemCheckbox
	MenuItemRadio
	MenuItemSubmenu
)

// Menu represents a native menu
type Menu struct {
	items []*MenuItem
	impl  menuImpl
	mu    sync.Mutex
}

// menuImpl is the platform-specific menu interface
type menuImpl interface {
	update()
	show()
	destroy()
}

// NewMenu creates a new menu
func NewMenu() *Menu {
	return &Menu{
		items: make([]*MenuItem, 0),
	}
}

// Add adds a text menu item
func (m *Menu) Add(label string) *MenuItem {
	item := newMenuItem(label, MenuItemText)
	m.mu.Lock()
	m.items = append(m.items, item)
	m.mu.Unlock()
	return item
}

// AddSeparator adds a separator
func (m *Menu) AddSeparator() *MenuItem {
	item := newMenuItem("", MenuItemSeparator)
	m.mu.Lock()
	m.items = append(m.items, item)
	m.mu.Unlock()
	return item
}

// AddCheckbox adds a checkbox menu item
func (m *Menu) AddCheckbox(label string, checked bool) *MenuItem {
	item := newMenuItem(label, MenuItemCheckbox)
	item.checked = checked
	m.mu.Lock()
	m.items = append(m.items, item)
	m.mu.Unlock()
	return item
}

// AddRadio adds a radio menu item
func (m *Menu) AddRadio(label string, checked bool) *MenuItem {
	item := newMenuItem(label, MenuItemRadio)
	item.checked = checked
	m.mu.Lock()
	m.items = append(m.items, item)
	m.mu.Unlock()
	return item
}

// AddSubmenu adds a submenu
func (m *Menu) AddSubmenu(label string) *Menu {
	item := newMenuItem(label, MenuItemSubmenu)
	item.submenu = NewMenu()
	m.mu.Lock()
	m.items = append(m.items, item)
	m.mu.Unlock()
	return item.submenu
}

// Items returns the menu items
func (m *Menu) Items() []*MenuItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*MenuItem(nil), m.items...)
}

// Update updates the menu (platform-specific)
func (m *Menu) Update() {
	m.mu.Lock()
	impl := m.impl
	m.mu.Unlock()

	if impl != nil {
		impl.update()
	}
}

// show displays the menu
func (m *Menu) show() {
	m.mu.Lock()
	impl := m.impl
	m.mu.Unlock()

	if impl != nil {
		impl.show()
	}
}

// Destroy destroys the menu
func (m *Menu) Destroy() {
	m.mu.Lock()
	impl := m.impl
	items := m.items
	m.mu.Unlock()

	if impl != nil {
		impl.destroy()
	}

	// Destroy all items
	for _, item := range items {
		item.destroy()
	}
}

// processRadioGroups groups adjacent radio items
func (m *Menu) processRadioGroups() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var radioGroup []*MenuItem

	closeOutRadioGroups := func() {
		if len(radioGroup) > 0 {
			for _, item := range radioGroup {
				item.mu.Lock()
				item.radioGroupMembers = radioGroup
				item.mu.Unlock()
			}
			radioGroup = []*MenuItem{}
		}
	}

	for _, item := range m.items {
		item.mu.Lock()
		itemType := item.itemType
		submenu := item.submenu
		item.mu.Unlock()

		if itemType != MenuItemRadio {
			closeOutRadioGroups()
		}

		if itemType == MenuItemSubmenu && submenu != nil {
			submenu.processRadioGroups()
			continue
		}

		if itemType == MenuItemRadio {
			radioGroup = append(radioGroup, item)
		}
	}
	closeOutRadioGroups()
}
