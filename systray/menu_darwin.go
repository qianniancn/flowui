//go:build darwin && !ios

package systray

import (
	"sync"
	"unsafe"

	"github.com/qianniancn/FlowUI/internal/sys/darwin"
)

type macosMenu struct {
	parent    *Menu
	owner     *macosSystemTray
	mu        sync.Mutex
	nsMenu    unsafe.Pointer
	destroyed bool
}

func newMacosMenu(parent *Menu, owner *macosSystemTray) *macosMenu {
	menu := &macosMenu{
		parent: parent,
		owner:  owner,
	}
	menu.update()
	return menu
}

func (m *macosMenu) update() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.destroyed {
		return
	}
	if m.nsMenu == nil {
		m.nsMenu = darwin.CreateNSMenu(m.parent.Label())
	} else {
		darwin.ClearMenu(m.nsMenu)
	}
	m.buildMenu(m.nsMenu, m.parent)
}

func (m *macosMenu) buildMenu(parentNSMenu unsafe.Pointer, menu *Menu) {
	items := menu.Items()
	for _, item := range items {
		item.mu.Lock()
		oldImpl, _ := item.impl.(*macosMenuItem)
		itemType := item.itemType
		label := item.label
		hidden := item.hidden
		submenu := item.submenu
		item.mu.Unlock()
		if oldImpl != nil {
			oldImpl.destroy()
		}

		switch itemType {
		case MenuItemSeparator:
			darwin.AddMenuSeparator(parentNSMenu)

		case MenuItemSubmenu:
			if submenu != nil {
				nsSubmenu := darwin.CreateNSMenu(label)
				m.buildMenu(nsSubmenu, submenu)

				menuItem := newMacosMenuItem(item, m)
				item.mu.Lock()
				item.impl = menuItem
				item.mu.Unlock()

				darwin.AddMenuItem(parentNSMenu, menuItem.nsMenuItem)
				darwin.SetMenuItemSubmenu(menuItem.nsMenuItem, nsSubmenu)
				darwin.ReleaseNSMenu(nsSubmenu)
			}

		default:
			menuItem := newMacosMenuItem(item, m)
			item.mu.Lock()
			item.impl = menuItem
			item.mu.Unlock()

			if hidden {
				menuItem.setHidden(true)
			}

			darwin.AddMenuItem(parentNSMenu, menuItem.nsMenuItem)
		}
	}
}

func (m *macosMenu) show() {
	// Menu is shown via darwin.ShowMenu from systray implementation
}

func (m *macosMenu) destroy() {
	if m.owner != nil {
		m.owner.removeMenu(m)
	}
	m.detach()
	m.destroyNative()
}

func (m *macosMenu) nativeHandle() unsafe.Pointer {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nsMenu
}

func (m *macosMenu) destroyNative() {
	m.mu.Lock()
	nsMenu := m.nsMenu
	m.nsMenu = nil
	m.destroyed = true
	m.mu.Unlock()
	m.destroyItemImplementations(m.parent)
	if nsMenu != nil {
		darwin.ReleaseNSMenu(nsMenu)
	}
}

func (m *macosMenu) destroyItemImplementations(menu *Menu) {
	if menu == nil {
		return
	}
	for _, item := range menu.Items() {
		item.mu.Lock()
		impl, _ := item.impl.(*macosMenuItem)
		submenu := item.submenu
		item.mu.Unlock()
		if impl != nil && impl.menu == m {
			impl.destroy()
		}
		m.destroyItemImplementations(submenu)
	}
}

func (m *macosMenu) detach() {
	if m.parent == nil {
		return
	}
	m.parent.mu.Lock()
	if m.parent.impl == m {
		m.parent.impl = nil
	}
	m.parent.mu.Unlock()
}

// Label returns the menu label
func (m *Menu) Label() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Menu doesn't have a label field in our implementation
	// Could add one if needed
	return ""
}

type macosMenuItem struct {
	parent     *MenuItem
	menu       *macosMenu
	mu         sync.Mutex
	nsMenuItem unsafe.Pointer
}

//export menuItemCallback
func menuItemCallback(id int64) {
	item := GetMenuItemByID(uint(id))
	if item != nil {
		go item.Click()
	}
}

func newMacosMenuItem(parent *MenuItem, menu *macosMenu) *macosMenuItem {
	parent.mu.Lock()
	label := parent.label
	id := parent.id
	checked := parent.checked
	disabled := parent.disabled
	parent.mu.Unlock()

	nsMenuItem := darwin.CreateNSMenuItem(label, int64(id))

	item := &macosMenuItem{
		parent:     parent,
		menu:       menu,
		nsMenuItem: nsMenuItem,
	}

	// Set initial state
	if checked {
		darwin.SetMenuItemChecked(nsMenuItem, true)
	}
	if disabled {
		darwin.SetMenuItemEnabled(nsMenuItem, false)
	}

	return item
}

func (m *macosMenuItem) setLabel(label string) {
	m.menu.update()
}

func (m *macosMenuItem) setTooltip(tooltip string) {
	// macOS menu items don't support tooltips
}

func (m *macosMenuItem) setDisabled(disabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nsMenuItem == nil {
		return
	}
	darwin.SetMenuItemEnabled(m.nsMenuItem, !disabled)
}

func (m *macosMenuItem) setChecked(checked bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nsMenuItem == nil {
		return
	}
	darwin.SetMenuItemChecked(m.nsMenuItem, checked)
}

func (m *macosMenuItem) setHidden(hidden bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nsMenuItem == nil {
		return
	}
	darwin.SetMenuItemHidden(m.nsMenuItem, hidden)
}

func (m *macosMenuItem) destroy() {
	m.parent.mu.Lock()
	if m.parent.impl == m {
		m.parent.impl = nil
	}
	m.parent.mu.Unlock()

	m.mu.Lock()
	nsMenuItem := m.nsMenuItem
	m.nsMenuItem = nil
	m.mu.Unlock()
	if nsMenuItem != nil {
		darwin.ReleaseNSMenuItem(nsMenuItem)
	}
}
