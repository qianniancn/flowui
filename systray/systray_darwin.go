//go:build darwin && !ios

package systray

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/qianniancn/FlowUI/internal/sys/darwin"
)

type macosSystemTray struct {
	parent *SystemTray
	mu     sync.Mutex

	nsStatusItem   unsafe.Pointer
	nsImage        unsafe.Pointer
	nsMenu         unsafe.Pointer
	nativeMenu     *macosMenu
	isTemplateIcon bool
	destroyed      bool
}

// Icon positions for macOS
const (
	NSImageLeading  = 4
	NSImageLeft     = 2
	NSImageOnly     = 1
	NSImageOverlaps = 5
)

// Button types for click events
const (
	leftButtonDown  = 1
	rightButtonDown = 3
)

// Global map to track system trays by ID
var (
	macosSystemTrayMap   = make(map[uint]*macosSystemTray)
	macosSystemTrayMapMu sync.RWMutex
)

//export systrayClickCallback
func systrayClickCallback(id int64, buttonID int) {
	macosSystemTrayMapMu.RLock()
	tray := macosSystemTrayMap[uint(id)]
	macosSystemTrayMapMu.RUnlock()
	if tray == nil {
		return
	}
	tray.processClick(buttonID)
}

//export systrayPreClickCallback
func systrayPreClickCallback(id int64, buttonID int) int {
	macosSystemTrayMapMu.RLock()
	tray := macosSystemTrayMap[uint(id)]
	macosSystemTrayMapMu.RUnlock()
	if tray == nil {
		return 0
	}
	tray.mu.Lock()
	hasMenu := tray.nsMenu != nil
	tray.mu.Unlock()
	if !hasMenu {
		return 0
	}
	tray.parent.mu.Lock()
	clickHandler := tray.parent.clickHandler
	rightClickHandler := tray.parent.rightClickHandler
	tray.parent.mu.Unlock()

	switch buttonID {
	case leftButtonDown:
		if clickHandler == nil {
			return 1 // Show menu via native tracking
		}
	case rightButtonDown:
		if rightClickHandler == nil {
			return 1 // Show menu via native tracking
		}
	}
	return 0 // Let custom handler fire
}

func newSystemTrayImpl(parent *SystemTray) systemTrayImpl {
	impl := &macosSystemTray{
		parent: parent,
	}
	macosSystemTrayMapMu.Lock()
	macosSystemTrayMap[parent.id] = impl
	macosSystemTrayMapMu.Unlock()
	return impl
}

func (m *macosSystemTray) setLabel(label string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nsStatusItem == nil {
		return
	}
	darwin.SystemTraySetLabel(m.nsStatusItem, label)
}

func (m *macosSystemTray) setTooltip(tooltip string) {
	// Tooltips not supported on macOS status bar
}

func (m *macosSystemTray) setIcon(icon []byte) {
	if err := m.setIconNative(icon); err != nil {
		m.parent.reportError(err)
	}
}

func (m *macosSystemTray) setIconNative(icon []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nsStatusItem == nil || m.destroyed {
		return nil
	}
	nsImage := darwin.ImageFromBytes(icon)
	if nsImage == nil {
		return fmt.Errorf("systray: decode macOS icon")
	}
	m.nsImage = nsImage
	darwin.SystemTraySetIcon(m.nsStatusItem, nsImage, NSImageLeading, m.isTemplateIcon)
	return nil
}

func (m *macosSystemTray) setMenu(menu *Menu) {
	if err := m.setMenuNative(menu); err != nil {
		m.parent.reportError(err)
	}
}

func (m *macosSystemTray) setMenuNative(menu *Menu) error {
	var macMenu *macosMenu
	if menu != nil {
		menu.processRadioGroups()
		macMenu = newMacosMenu(menu, m)
		if macMenu == nil || macMenu.nativeHandle() == nil {
			return fmt.Errorf("systray: create macOS menu")
		}
	}

	m.mu.Lock()
	if m.nsStatusItem == nil || m.destroyed {
		m.mu.Unlock()
		if macMenu != nil {
			macMenu.destroyNative()
		}
		return nil
	}
	oldMenu := m.nativeMenu
	m.nativeMenu = macMenu
	m.nsMenu = nil
	if macMenu != nil {
		m.nsMenu = macMenu.nativeHandle()
	}
	darwin.SystemTraySetCachedMenu(m.nsStatusItem, m.nsMenu)
	m.mu.Unlock()

	if oldMenu != nil {
		oldMenu.detach()
		oldMenu.destroyNative()
	}
	if macMenu != nil {
		menu.mu.Lock()
		menu.impl = macMenu
		menu.mu.Unlock()
	}
	return nil
}

func (m *macosSystemTray) run() {
	statusItem := darwin.SystemTrayNew(int64(m.parent.id))
	if statusItem == nil {
		m.destroy()
		m.parent.fail(fmt.Errorf("systray: create macOS status item"))
		return
	}
	m.mu.Lock()
	if m.destroyed {
		m.mu.Unlock()
		darwin.SystemTrayDestroy(statusItem)
		return
	}
	m.nsStatusItem = statusItem
	m.mu.Unlock()

	// Set label if provided
	m.parent.mu.Lock()
	label := m.parent.label
	icon := m.parent.icon
	menu := m.parent.menu
	m.parent.mu.Unlock()

	if label != "" {
		m.setLabel(label)
	}

	// Set icon if provided
	if len(icon) > 0 {
		if err := m.setIconNative(icon); err != nil {
			m.destroy()
			m.parent.fail(err)
			return
		}
	}

	// Set menu if provided
	if menu != nil {
		if err := m.setMenuNative(menu); err != nil {
			m.destroy()
			m.parent.fail(err)
			return
		}
	}

	m.parent.markReady()

	// macOS runs on the main event loop, no blocking needed
}

func (m *macosSystemTray) destroy() {
	m.mu.Lock()
	m.destroyed = true
	statusItem := m.nsStatusItem
	nativeMenu := m.nativeMenu
	m.nsStatusItem = nil
	m.nsMenu = nil
	m.nativeMenu = nil
	m.mu.Unlock()
	if statusItem != nil {
		darwin.SystemTraySetCachedMenu(statusItem, nil)
		darwin.SystemTrayDestroy(statusItem)
	}
	if nativeMenu != nil {
		nativeMenu.detach()
		nativeMenu.destroyNative()
	}

	macosSystemTrayMapMu.Lock()
	delete(macosSystemTrayMap, m.parent.id)
	macosSystemTrayMapMu.Unlock()
}

func (m *macosSystemTray) removeMenu(menu *macosMenu) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nativeMenu != menu {
		return
	}
	if m.nsStatusItem != nil && !m.destroyed {
		darwin.SystemTraySetCachedMenu(m.nsStatusItem, nil)
	}
	m.nativeMenu = nil
	m.nsMenu = nil
}

func (m *macosSystemTray) show() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nsStatusItem == nil {
		return
	}
	darwin.SystemTrayShow(m.nsStatusItem)
}

func (m *macosSystemTray) hide() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nsStatusItem == nil {
		return
	}
	darwin.SystemTrayHide(m.nsStatusItem)
}

func (m *macosSystemTray) processClick(buttonID int) {
	m.parent.mu.Lock()
	clickHandler := m.parent.clickHandler
	rightClickHandler := m.parent.rightClickHandler
	menu := m.parent.menu
	m.parent.mu.Unlock()
	m.mu.Lock()
	statusItem := m.nsStatusItem
	nsMenu := m.nsMenu
	m.mu.Unlock()

	switch buttonID {
	case leftButtonDown:
		if clickHandler != nil {
			go clickHandler()
			return
		}
		// If no custom handler, show menu
		if menu != nil && statusItem != nil && nsMenu != nil {
			darwin.ShowMenu(statusItem, nsMenu)
		}

	case rightButtonDown:
		if rightClickHandler != nil {
			go rightClickHandler()
			return
		}
		// If no custom handler, show menu
		if menu != nil && statusItem != nil && nsMenu != nil {
			darwin.ShowMenu(statusItem, nsMenu)
		}
	}
}

// SetTemplateIcon sets a template icon (adapts to light/dark mode)
func (t *SystemTray) SetTemplateIcon(icon []byte) *SystemTray {
	owned := append([]byte(nil), icon...)
	t.mu.Lock()
	t.icon = owned
	impl := t.impl
	t.mu.Unlock()

	if macImpl, ok := impl.(*macosSystemTray); ok {
		macImpl.mu.Lock()
		macImpl.isTemplateIcon = true
		var iconErr error
		if macImpl.nsStatusItem != nil {
			nsImage := darwin.ImageFromBytes(owned)
			if nsImage != nil {
				macImpl.nsImage = nsImage
				darwin.SystemTraySetIcon(macImpl.nsStatusItem, nsImage, NSImageLeading, true)
			} else {
				iconErr = fmt.Errorf("systray: decode macOS template icon")
			}
		}
		macImpl.mu.Unlock()
		if iconErr != nil {
			t.reportError(iconErr)
		}
	}
	return t
}
